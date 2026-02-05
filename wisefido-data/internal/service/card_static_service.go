package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	commoncard "owl-common/card"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

const staticKeyPrefix = "vital-focus:card:"
const staticKeySuffix = ":static"

// AllowedCardIDsProvider 提供用户有权查看的卡片 ID 列表
type AllowedCardIDsProvider interface {
	GetAllowedCardIDs(ctx context.Context, tenantID, userID, userType, userRole string) ([]string, error)
	// GetAllowedCardIDsByBranches 按 branch 分组返回，用于 Redis 缓存
	GetAllowedCardIDsByBranches(ctx context.Context, tenantID, userID, userType, userRole string) (map[string][]string, error)
}

// ListVitalFocusCardInfoRequest 后端专用：静态卡片列表，不对前端开放，不作安全检查
// 仅限本机内部调用（handler 需根据 IsRequestFromLocalhost 限制入口）
type ListVitalFocusCardInfoRequest struct {
	TenantID string // 必填；"*" 表示查询所有租户，不按 tenant 筛选
	BranchID string // 可选
	UnitID   string // 可选
	CardID   string // 可选，限制为特定卡片ID
}

// UserCardsCacheProvider 可选：用户卡片缓存，nil 时不使用缓存
type UserCardsCacheProvider interface {
	Get(ctx context.Context, tenantID, userID string, branchIDs []string) ([]string, error)
	Set(ctx context.Context, tenantID, userID string, byBranch map[string][]string) error
	GetBranchIDsFromIndex(ctx context.Context, tenantID, userID string) ([]string, error)
}

// CardStaticService 静态卡片服务
type CardStaticService struct {
	kv             store.KV
	unitsRepo      repository.UnitsRepository
	perm           AllowedCardIDsProvider         // 权限校验
	userCardsCache UserCardsCacheProvider         // 可选，nil 时直接计算
	residentsRepo  repository.ResidentsRepository // 用于查询 resident_caregivers 等表
	logger         *zap.Logger
}

// NewCardStaticService 创建静态卡片服务；userCardsCache 可选，nil 时不使用缓存
func NewCardStaticService(
	kv store.KV,
	unitsRepo repository.UnitsRepository,
	perm AllowedCardIDsProvider,
	userCardsCache UserCardsCacheProvider,
	residentsRepo repository.ResidentsRepository,
	logger *zap.Logger,
) *CardStaticService {
	return &CardStaticService{kv: kv, unitsRepo: unitsRepo, perm: perm, userCardsCache: userCardsCache, residentsRepo: residentsRepo, logger: logger}
}

// ListVitalFocusCardInfo 后端专用：从 cache 获取静态卡片列表，仅按 tenant/branch/unit 过滤，不做用户权限校验
func (s *CardStaticService) ListVitalFocusCardInfo(ctx context.Context, req ListVitalFocusCardInfoRequest) ([]commoncard.VitalFocusCardInfo, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	return s.listCardsWithFilter(ctx, req, nil)
}

// ListCards 获取当前用户有权查看的卡片简化列表
// branchIDs 为空时返回所有可访问的卡片，不为空时只返回指定 branch 下的卡片
func (s *CardStaticService) ListCards(ctx context.Context, tenantID, userID, userRole string, branchIDs []string) ([]commoncard.CardIndexItem, error) {
	if tenantID == "" || userID == "" {
		return nil, fmt.Errorf("tenant_id, user_id are required")
	}
	if s.perm == nil {
		return nil, fmt.Errorf("permission provider not configured")
	}

	// 获取用户有权访问的所有 card IDs
	allowedIDs, err := s.perm.GetAllowedCardIDs(ctx, tenantID, userID, "staff", userRole)
	if err != nil {
		return nil, fmt.Errorf("get allowed card IDs: %w", err)
	}

	if len(allowedIDs) == 0 {
		return []commoncard.CardIndexItem{}, nil
	}

	allowedSet := make(map[string]bool, len(allowedIDs))
	for _, id := range allowedIDs {
		allowedSet[id] = true
	}

	// 构建 branch filter
	var branchFilter map[string]bool
	if len(branchIDs) > 0 {
		branchFilter = make(map[string]bool, len(branchIDs))
		for _, bid := range branchIDs {
			branchFilter[bid] = true
		}
	}

	// 从 Redis 扫描所有卡片
	keys, err := s.kv.ScanKeys(ctx, staticKeyPrefix+"*"+staticKeySuffix)
	if err != nil {
		return nil, fmt.Errorf("scan static cache keys: %w", err)
	}

	items := make([]commoncard.CardIndexItem, 0, len(keys))
	for _, key := range keys {
		cardID := extractCardIDFromStaticKey(key)
		if cardID == "" || !allowedSet[cardID] {
			continue
		}

		raw, err := s.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				continue
			}
			s.logger.Warn("get static cache failed", zap.String("key", key), zap.Error(err))
			continue
		}

		var card commoncard.VitalFocusCardInfo
		if err := json.Unmarshal([]byte(raw), &card); err != nil {
			s.logger.Debug("unmarshal static card failed", zap.String("key", key), zap.Error(err))
			continue
		}

		// 按 tenant 过滤
		if card.TenantID != tenantID {
			continue
		}

		// 按 branch 过滤
		if branchFilter != nil && !branchFilter[card.BranchID] {
			continue
		}

		// 转换为 CardIndexItem
		deviceIDs := extractDeviceIDs(card.Devices)
		item := commoncard.CardIndexItem{
			CardID:            card.CardID,
			CardName:          card.CardName,
			CardAddress:       card.CardAddress,
			BranchID:          card.BranchID,
			IconAlarmLevel:    derefInt(card.IconAlarmLevel),
			PopAlarmEmerge:    derefInt(card.PopAlarmEmerge),
			DeviceIDs:         deviceIDs,
			PrimaryResidentID: card.PrimaryResidentID,
		}
		items = append(items, item)
	}

	return items, nil
}

// GetCardInfo 获取单个卡片的完整详情（用户需有权限）
func (s *CardStaticService) GetCardInfo(ctx context.Context, tenantID, userID, userRole, cardID string) (*commoncard.VitalFocusCardInfo, error) {
	if tenantID == "" || userID == "" || cardID == "" {
		return nil, fmt.Errorf("tenant_id, user_id, card_id are required")
	}
	if s.perm == nil {
		return nil, fmt.Errorf("permission provider not configured")
	}

	// 检查用户权限
	allowedIDs, err := s.perm.GetAllowedCardIDs(ctx, tenantID, userID, "staff", userRole)
	if err != nil {
		return nil, fmt.Errorf("get allowed card IDs: %w", err)
	}

	allowed := false
	for _, id := range allowedIDs {
		if id == cardID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("user does not have permission to access this card")
	}

	// 从 Redis 读取卡片详情
	key := staticKeyPrefix + cardID + staticKeySuffix
	raw, err := s.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, store.ErrMiss) {
			return nil, fmt.Errorf("card not found")
		}
		return nil, fmt.Errorf("get card from cache: %w", err)
	}

	var card commoncard.VitalFocusCardInfo
	if err := json.Unmarshal([]byte(raw), &card); err != nil {
		return nil, fmt.Errorf("unmarshal card: %w", err)
	}

	// 再次验证 tenant
	if card.TenantID != tenantID {
		return nil, fmt.Errorf("card tenant mismatch")
	}

	return &card, nil
}

// getAllowedCardIDsWithCache 优先读缓存，miss 时重算并写入
func (s *CardStaticService) getAllowedCardIDsWithCache(ctx context.Context, tenantID, userID, userType, userRole string) ([]string, error) {
	if s.userCardsCache == nil {
		return s.perm.GetAllowedCardIDs(ctx, tenantID, userID, userType, userRole)
	}
	branchIDs, err := s.userCardsCache.GetBranchIDsFromIndex(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if len(branchIDs) == 0 {
		// index miss：重算并写入
		byBranch, err := s.perm.GetAllowedCardIDsByBranches(ctx, tenantID, userID, userType, userRole)
		if err != nil {
			return nil, err
		}
		if err := s.userCardsCache.Set(ctx, tenantID, userID, byBranch); err != nil {
			s.logger.Warn("set user cards cache failed", zap.String("user_id", userID), zap.Error(err))
		}
		return s.perm.GetAllowedCardIDs(ctx, tenantID, userID, userType, userRole)
	}
	ids, err := s.userCardsCache.Get(ctx, tenantID, userID, branchIDs)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		// 任一 branch miss：重算并写入
		byBranch, err := s.perm.GetAllowedCardIDsByBranches(ctx, tenantID, userID, userType, userRole)
		if err != nil {
			return nil, err
		}
		if err := s.userCardsCache.Set(ctx, tenantID, userID, byBranch); err != nil {
			s.logger.Warn("set user cards cache failed", zap.String("user_id", userID), zap.Error(err))
		}
		return s.perm.GetAllowedCardIDs(ctx, tenantID, userID, userType, userRole)
	}
	return ids, nil
}

func idsToSet(ids []string) map[string]bool {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// listCardsWithFilter 内部方法：按 tenant/branch/unit 过滤，并按 allowedCardIDs 过滤（nil 表示不过滤）
func (s *CardStaticService) listCardsWithFilter(ctx context.Context, req ListVitalFocusCardInfoRequest, allowedCardIDs map[string]bool) ([]commoncard.VitalFocusCardInfo, error) {
	var unitIDsInBranch map[string]bool
	// Branch 过滤需具体 tenant；TenantID="*" 时跳过（units 表按 tenant 隔离）
	if req.BranchID != "" && req.TenantID != "*" && s.unitsRepo != nil {
		units, _, err := s.unitsRepo.ListUnits(ctx, req.TenantID, repository.UnitFilters{BranchID: req.BranchID}, 1, 5000)
		if err != nil {
			return nil, fmt.Errorf("list units by branch: %w", err)
		}
		unitIDsInBranch = make(map[string]bool, len(units))
		for _, u := range units {
			if u != nil && u.UnitID != "" {
				unitIDsInBranch[u.UnitID] = true
			}
		}
	}
	keys, err := s.kv.ScanKeys(ctx, staticKeyPrefix+"*"+staticKeySuffix)
	if err != nil {
		return nil, fmt.Errorf("scan static cache keys: %w", err)
	}
	out := make([]commoncard.VitalFocusCardInfo, 0, len(keys))
	for _, key := range keys {
		cardID := extractCardIDFromStaticKey(key)
		if cardID == "" {
			continue
		}

		// 按 CardID 过滤
		if req.CardID != "" && cardID != req.CardID {
			continue
		}

		raw, err := s.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				continue
			}
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return nil, err
			}
			s.logger.Warn("get static cache failed", zap.String("key", key), zap.Error(err))
			continue
		}
		var card commoncard.VitalFocusCardInfo
		if err := json.Unmarshal([]byte(raw), &card); err != nil {
			s.logger.Debug("unmarshal static card failed", zap.String("key", key), zap.Error(err))
			continue
		}
		if req.TenantID != "*" && card.TenantID != req.TenantID {
			continue
		}
		if req.UnitID != "" && card.UnitID != req.UnitID {
			continue
		}
		if len(unitIDsInBranch) > 0 && !unitIDsInBranch[card.UnitID] {
			continue
		}
		if allowedCardIDs != nil && !allowedCardIDs[card.CardID] {
			continue
		}
		out = append(out, card)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CardID < out[j].CardID })
	return out, nil
}

func extractCardIDFromStaticKey(key string) string {
	if !strings.HasPrefix(key, staticKeyPrefix) || !strings.HasSuffix(key, staticKeySuffix) {
		return ""
	}
	return key[len(staticKeyPrefix) : len(key)-len(staticKeySuffix)]
}

// extractDeviceIDs 从 DeviceInfo 数组提取 device_id 列表
func extractDeviceIDs(devices []commoncard.DeviceInfo) []string {
	if len(devices) == 0 {
		return nil
	}
	ids := make([]string, 0, len(devices))
	for _, d := range devices {
		if d.DeviceID != "" {
			ids = append(ids, d.DeviceID)
		}
	}
	return ids
}

// derefInt 将 *int 转换为 int，nil 时返回 0
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// normPage 规范化分页参数（保留以支持其他地方的使用）
func normPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 30
	}
	if pageSize > 30 {
		pageSize = 30
	}
	return page, pageSize
}
