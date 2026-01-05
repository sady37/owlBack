package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"wisefido-data/internal/models"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// VitalFocusService Vital Focus 卡片服务接口
type VitalFocusService interface {
	// ListCards 获取卡片列表（根据角色权限过滤）
	// 从 Redis 读取卡片数据（已包含实时数据和报警数据），应用权限过滤
	ListCards(ctx context.Context, req ListCardsRequest) (*ListCardsResponse, error)

	// GetCard 获取单个卡片（包含实时数据和报警数据）
	// 从 Redis 读取卡片数据，验证权限
	GetCard(ctx context.Context, req GetCardRequest) (*models.VitalFocusCard, error)
}

// vitalFocusService Vital Focus 卡片服务实现
type vitalFocusService struct {
	kv            store.KV
	usersRepo     repository.UsersRepository
	residentsRepo repository.ResidentsRepository
	unitsRepo     repository.UnitsRepository
	db            *sql.DB // 用于复杂 SQL 查询（resident_caregivers, location_tag）
	logger        *zap.Logger
}

// NewVitalFocusService 创建 Vital Focus 服务
func NewVitalFocusService(
	kv store.KV,
	usersRepo repository.UsersRepository,
	residentsRepo repository.ResidentsRepository,
	unitsRepo repository.UnitsRepository,
	db *sql.DB,
	logger *zap.Logger,
) VitalFocusService {
	return &vitalFocusService{
		kv:            kv,
		usersRepo:     usersRepo,
		residentsRepo: residentsRepo,
		unitsRepo:     unitsRepo,
		db:            db,
		logger:        logger,
	}
}

// ListCardsRequest 获取卡片列表请求
type ListCardsRequest struct {
	TenantID string

	// 用户信息（从 HTTP Header 获取）
	UserID   string // X-User-Id
	UserRole string // X-User-Role（可选，如果提供可减少 DB 查询）
	UserType string // "staff" | "resident"（从登录类型推断）

	// 分页参数（可选）
	Page     int // 默认 1
	PageSize int // 默认 10
}

// ListCardsResponse 获取卡片列表响应
type ListCardsResponse struct {
	Items      []models.VitalFocusCard
	Pagination models.BackendPagination
}

// GetCardRequest 获取单个卡片请求
type GetCardRequest struct {
	TenantID string
	CardID   string

	// 权限参数
	UserID   string
	UserRole string // 可选
	UserType string // "staff" | "resident"
}

// 错误定义
var (
	ErrCardNotFound      = errors.New("card not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidUserType   = errors.New("invalid user type")
	ErrRedisNotAvailable = errors.New("redis not available")
)

// ListCards 获取卡片列表（根据角色权限过滤）
func (s *vitalFocusService) ListCards(ctx context.Context, req ListCardsRequest) (*ListCardsResponse, error) {
	// 1. 从 Redis 扫描所有 full cache 键
	keys, err := s.kv.ScanKeys(ctx, "vital-focus:card:*:full")
	if err != nil {
		// 联调友好：当 Redis 不可用/没有跑 aggregator 时，不要让前端报错；返回空列表即可
		s.logger.Warn("ScanKeys failed, returning empty cards list", zap.Error(err))
		return &ListCardsResponse{
			Items: []models.VitalFocusCard{},
			Pagination: models.BackendPagination{
				Size:      req.PageSize,
				Page:      req.Page,
				Count:     0,
				Sort:      "",
				Direction: 0,
			},
		}, nil
	}

	// 2. 读取并解析所有卡片数据
	all := make([]models.VitalFocusCard, 0, len(keys))
	for _, key := range keys {
		raw, err := s.kv.Get(ctx, key)
		if err != nil {
			continue
		}

		card, ok := decodeAndNormalizeFullCard(raw)
		if !ok {
			s.logger.Debug("Failed to decode and normalize card", zap.String("key", key))
			continue
		}

		// 3. 按 tenant_id 过滤
		if req.TenantID != "" && card.TenantID != req.TenantID {
			continue
		}

		all = append(all, card)
	}

	// 4. 权限过滤
	filtered := all
	if req.UserID != "" {
		var err error
		if req.UserType == "resident" {
			filtered, err = s.filterCardsForResident(ctx, req.UserID, req.TenantID, all)
			if err != nil {
				return nil, fmt.Errorf("failed to filter cards for resident: %w", err)
			}
		} else if req.UserType == "staff" || req.UserType == "" {
			// 默认当作 staff
			filtered, err = s.filterCardsForStaff(ctx, req.UserID, req.TenantID, all)
			if err != nil {
				return nil, fmt.Errorf("failed to filter cards for staff: %w", err)
			}
		} else {
			return nil, fmt.Errorf("%w: %s", ErrInvalidUserType, req.UserType)
		}
	}

	// 5. 排序（按 card_id）
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CardID < filtered[j].CardID
	})

	// 6. 分页
	total := len(filtered)
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	var items []models.VitalFocusCard
	if start < end {
		items = filtered[start:end]
	} else {
		items = []models.VitalFocusCard{}
	}

	return &ListCardsResponse{
		Items: items,
		Pagination: models.BackendPagination{
			Size:      pageSize,
			Page:      page,
			Count:     total,
			Sort:      "",
			Direction: 0,
		},
	}, nil
}

// GetCard 获取单个卡片（包含实时数据和报警数据）
func (s *vitalFocusService) GetCard(ctx context.Context, req GetCardRequest) (*models.VitalFocusCard, error) {
	// 1. 从 Redis 读取卡片
	key := fmt.Sprintf("vital-focus:card:%s:full", req.CardID)
	raw, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, ErrCardNotFound
	}

	card, ok := decodeAndNormalizeFullCard(raw)
	if !ok {
		return nil, ErrCardNotFound
	}

	// 2. 验证 tenant_id 匹配
	if req.TenantID != "" && card.TenantID != req.TenantID {
		return nil, ErrCardNotFound
	}

	// 3. 权限验证
	if req.UserID != "" {
		hasPermission, err := s.hasCardPermission(ctx, req.UserID, req.UserRole, req.UserType, req.TenantID, card)
		if err != nil {
			return nil, fmt.Errorf("failed to check permission: %w", err)
		}
		if !hasPermission {
			return nil, ErrPermissionDenied
		}
	}

	return &card, nil
}

// filterCardsForStaff 过滤 Staff 用户可见的卡片
func (s *vitalFocusService) filterCardsForStaff(
	ctx context.Context,
	userID, tenantID string,
	cards []models.VitalFocusCard,
) ([]models.VitalFocusCard, error) {
	if s.db == nil {
		// 如果没有 DB，返回所有卡片（向后兼容）
		return cards, nil
	}

	// 1. 查询用户信息（role, alarm_scope, tags）
	user, err := s.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Admin 角色：返回所有 tenant 下的卡片
	if user.Role == "Admin" {
		return cards, nil
	}

	// 3. ALL scope：返回所有 tenant 下的卡片
	if user.AlarmScope.Valid && user.AlarmScope.String == "ALL" {
		return cards, nil
	}

	// 4. LOCATION scope：根据 location_tag 匹配 users.tags
	// 注意：文档中提到 locations.location_tag，但实际的 units 表中没有 location_tag 字段
	// 根据卡片地址计算规则，卡片地址是 branch_name + "-" + building_name + "-" + unit_name
	// 可能 location_tag 实际上是指 branch_name，或者需要从其他表查询
	// 暂时跳过 LOCATION 权限过滤，返回空列表（待后续确认 location_tag 的实际含义）
	if user.AlarmScope.Valid && user.AlarmScope.String == "LOCATION" {
		s.logger.Warn("LOCATION scope permission not yet implemented",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
		)
		// TODO: 实现 LOCATION 权限过滤（需要确认 location_tag 的实际含义和存储位置）
		// 暂时返回空列表，避免返回所有卡片
		return []models.VitalFocusCard{}, nil
	}

	// 5. ASSIGNED_ONLY scope：根据 resident_caregivers 表过滤
	if user.AlarmScope.Valid && user.AlarmScope.String == "ASSIGNED_ONLY" {
		// 查询用户负责的 resident_id 列表
		// 注意：resident_caregivers 表中没有 is_active 字段，去掉该条件
		var assignedResidentIDs []string
		rows, err := s.db.QueryContext(ctx,
			`SELECT resident_id FROM resident_caregivers 
			 WHERE tenant_id = $1 AND caregiver_id = $2`,
			tenantID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query resident_caregivers: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var residentID string
			if err := rows.Scan(&residentID); err != nil {
				continue
			}
			assignedResidentIDs = append(assignedResidentIDs, residentID)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("failed to scan resident_caregivers: %w", err)
		}

		// 如果没有分配的住户，返回空列表
		if len(assignedResidentIDs) == 0 {
			return []models.VitalFocusCard{}, nil
		}

		// 查询 Location 卡片：获取分配给用户的住户的 location_id 列表
		var assignedLocationIDs []string
		if len(assignedResidentIDs) > 0 {
			query := `
				SELECT DISTINCT r.unit_id 
				FROM residents r
				WHERE r.tenant_id = $1 
				  AND r.resident_id = ANY($2)
				  AND r.unit_id IS NOT NULL
			`
			rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(assignedResidentIDs))
			if err != nil {
				return nil, fmt.Errorf("failed to query resident locations: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var locationID string
				if err := rows.Scan(&locationID); err != nil {
					continue
				}
				assignedLocationIDs = append(assignedLocationIDs, locationID)
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("failed to scan resident locations: %w", err)
			}
		}

		// 转换为 map 以便快速查找
		assignedResidentIDMap := make(map[string]bool)
		for _, id := range assignedResidentIDs {
			assignedResidentIDMap[id] = true
		}
		assignedLocationIDMap := make(map[string]bool)
		for _, id := range assignedLocationIDs {
			assignedLocationIDMap[id] = true
		}

		// 过滤卡片
		filtered := make([]models.VitalFocusCard, 0)
		for _, card := range cards {
			// ActiveBed 卡片：primary_resident_id 在分配的列表中
			if card.CardType == "ActiveBed" && card.PrimaryResidentID != "" {
				if assignedResidentIDMap[card.PrimaryResidentID] {
					filtered = append(filtered, card)
					continue
				}
			}

			// Location 卡片：location_id 在分配的列表中
			if card.CardType == "Location" && card.LocationID != "" {
				if assignedLocationIDMap[card.LocationID] {
					filtered = append(filtered, card)
					continue
				}
			}
		}
		return filtered, nil
	}

	// 默认：返回空列表（未知的 alarm_scope）
	return []models.VitalFocusCard{}, nil
}

// filterCardsForResident 过滤 Resident 用户可见的卡片
func (s *vitalFocusService) filterCardsForResident(
	ctx context.Context,
	residentID, tenantID string,
	cards []models.VitalFocusCard,
) ([]models.VitalFocusCard, error) {
	// 1. 查询住户信息（bed_id, unit_id）
	resident, err := s.residentsRepo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	filtered := make([]models.VitalFocusCard, 0)

	for _, card := range cards {
		// ActiveBed 卡片：bed_id == resident.bed_id 且 primary_resident_id == resident_id
		if card.CardType == "ActiveBed" {
			if card.BedID != "" && card.BedID == resident.BedID {
				if card.PrimaryResidentID == residentID {
					filtered = append(filtered, card)
					continue
				}
			}
		}

		// Location 卡片：location_id == resident.unit_id 且住户在 card.residents 中
		if card.CardType == "Location" {
			if card.LocationID != "" && card.LocationID == resident.UnitID {
				// 检查住户是否在 card.residents 中
				for _, r := range card.Residents {
					if r.ResidentID == residentID {
						filtered = append(filtered, card)
						break
					}
				}
			}
		}
	}

	return filtered, nil
}

// hasCardPermission 检查用户是否有权限查看指定卡片
func (s *vitalFocusService) hasCardPermission(
	ctx context.Context,
	userID, userRole, userType, tenantID string,
	card models.VitalFocusCard,
) (bool, error) {
	// 获取所有可见的卡片列表
	var allCards []models.VitalFocusCard
	var err error

	if userType == "resident" {
		allCards, err = s.filterCardsForResident(ctx, userID, tenantID, []models.VitalFocusCard{card})
		if err != nil {
			return false, err
		}
	} else if userType == "staff" || userType == "" {
		allCards, err = s.filterCardsForStaff(ctx, userID, tenantID, []models.VitalFocusCard{card})
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("%w: %s", ErrInvalidUserType, userType)
	}

	// 检查卡片是否在过滤后的列表中
	for _, c := range allCards {
		if c.CardID == card.CardID {
			return true, nil
		}
	}

	return false, nil
}
