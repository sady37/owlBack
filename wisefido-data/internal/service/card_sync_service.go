package service

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"owl-common/card"
	"owl-common/ddns"
	"owl-common/spatial"

	"go.uber.org/zap"
)

// CardSyncService — v2 stub。
//
// v1 范式：按 unit 批量 reconcile cards（CreateCardsForUnit + SyncDeviceCards + RecalcAllCardsAlarmState）。
// v2 范式：事件驱动 per-prefix（resident lifecycle / device bind 触发，单卡 INSERT/UPDATE/DELETE）。
//
// 本文件保留 API 兼容 caller（main.go 启动 reconcile job + 各 service 触发 SyncUnitCards）。
// 实际写入逻辑暂为 no-op；只保留 emit 路径以便 reset notification 仍可发。
//
// 真正的 v2 实现见 [doc/cards_v2_migration_checklist.md § Phase F]。
type CardSyncService struct {
	cardRepo  *repository.PostgresCardRepository
	creator   *CardCreateService
	publisher *publisher.ConfigPublisher
	realtime  *CardRealtimeService
	logger    *zap.Logger

	// Phase F+ — 启动 reconcile / per-tenant DDNS push 用
	db         *sql.DB
	ddns       *ddns.Client
	owlDomain  string
}

// NewCardSyncService 创建卡片同步服务
func NewCardSyncService(
	cardRepo *repository.PostgresCardRepository,
	publisher *publisher.ConfigPublisher,
	realtime *CardRealtimeService,
	logger *zap.Logger,
) *CardSyncService {
	creator := NewCardCreateService(cardRepo, logger)
	return &CardSyncService{
		cardRepo:  cardRepo,
		creator:   creator,
		publisher: publisher,
		realtime:  realtime,
		logger:    logger,
	}
}

// PublishConfigCardReset 发送 configCard reset 通知，委托给 ConfigPublisher。
func (s *CardSyncService) PublishConfigCardReset(ctx context.Context) error {
	return s.publisher.PublishConfigCardReset(ctx)
}

// SetReconcileDeps 装配启动 reconcile 所需依赖（db / DDNS / owlDomain）。
// 仅 main.go 启动期注入一次；后续 ReconcileResidentCards 调用 will use 这些。
func (s *CardSyncService) SetReconcileDeps(db *sql.DB, ddnsClient *ddns.Client, owlDomain string) {
	s.db = db
	s.ddns = ddnsClient
	if strings.TrimSpace(owlDomain) != "" {
		s.owlDomain = owlDomain
		if !strings.HasSuffix(s.owlDomain, ".") {
			s.owlDomain += "."
		}
	} else {
		s.owlDomain = "owl."
	}
}

// CreateCardsForUnit — v2 重新实现：unit /80 scope 的 resident → card reconcile。
//
// 与 v1 区别（参 doc/cards_v2_migration_checklist.md §一）：
//   - v1 按 unit JSONB residents/devices 数组重建 → v2 不再 maintain JSONB
//   - v1 在 cards 表 INSERT/DELETE 完整生命周期 → v2 cards 与 resident 解耦，
//     这里只 reconcile **resident_id 指针**（INSERT new active_bed/room/unit
//     card 对应 active resident_unit；NULL out 落空 resident_id）
//   - card 删除（spatial_prefix 下设备全空）由 CleanupOrphanCards 处理（device-driven）
//
// 启动 reconcile 走 main.go for tenant→for unit 循环现状（无破坏），但实际工作
// 在第一轮 (unit 重复扫描 resident_unit) 即完成；后续 unit 调用基本空跑。
// 简单清晰 > 性能（启动 reconcile 不在 hot path）。
func (s *CardSyncService) CreateCardsForUnit(ctx context.Context, tenantPrefix, unitPrefix string) (*CardUpdateStats, error) {
	stats := &CardUpdateStats{}
	if s.db == nil {
		s.logger.Debug("CreateCardsForUnit skipped (db not wired)")
		return stats, nil
	}
	if strings.TrimSpace(unitPrefix) == "" {
		return stats, nil
	}

	// 1) 扫 active resident_unit (resident_id, spatial_prefix) 在该 unit /80 下
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.resident_id::text, COALESCE(r.nickname, ''), ru.spatial_prefix::text
		  FROM residents r
		  JOIN resident_unit ru ON ru.resident_id = r.resident_id AND ru.valid_to IS NULL
		 WHERE r.status = 'active'
		   AND ru.spatial_prefix <<= $1::INET
	`, unitPrefix)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var hoa, nickname, prefixStr string
		if err := rows.Scan(&hoa, &nickname, &prefixStr); err != nil {
			s.logger.Warn("CreateCardsForUnit scan failed", zap.Error(err))
			continue
		}
		op, err := s.upsertResidentCard(ctx, tenantPrefix, hoa, nickname, prefixStr)
		if err != nil {
			s.logger.Warn("reconcile: upsertResidentCard failed",
				zap.String("hoa", hoa), zap.String("prefix", prefixStr), zap.Error(err))
			continue
		}
		switch op {
		case "created":
			stats.CreatedCount++
		case "updated":
			stats.UpdatedCount++
		case "unchanged":
			stats.UnchangedCount++
		}
		stats.ExistingCount++
	}
	if err := rows.Err(); err != nil {
		s.logger.Warn("CreateCardsForUnit rows err", zap.Error(err))
	}

	// 2) 清扫该 unit /80 下 resident_id 已落空的 card（resident 出院 / 转走但 card 行还指向旧 HoA）
	cleared, err := s.clearStaleResidentCards(ctx, unitPrefix)
	if err != nil {
		s.logger.Warn("CreateCardsForUnit clearStale failed", zap.String("unit", unitPrefix), zap.Error(err))
	}
	stats.DeletedCount = cleared // 复用现有字段：表示 resident 指针被清的数量（非物理删 card）
	return stats, nil
}

// upsertResidentCard 调 cards 表 UPSERT + 尝试 DDNS register；返回 op ∈ {"created","updated","unchanged"}。
func (s *CardSyncService) upsertResidentCard(ctx context.Context, tenantPrefix, hoa, nickname, prefixStr string) (string, error) {
	prefix, err := netip.ParsePrefix(prefixStr)
	if err != nil {
		return "", err
	}
	cardType := card.CardTypeForMasklen(prefix.Bits())
	if cardType == "" {
		return "", nil // 不支持的 masklen，跳过
	}
	shortName, _ := ddns.CardShortName(prefix)
	cardName := nickname
	if cardName == "" {
		cardName = shortName
	}

	// 查现状（现 resident_id / card_name）— card_id ≡ spatial_prefix，无独立 UUID 列
	var existingResident, existingName sql.NullString
	existingExists := false
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(resident_id::text, ''), COALESCE(card_name, '')
		  FROM cards
		 WHERE spatial_prefix = $1::INET
		 LIMIT 1
	`, prefixStr).Scan(&existingResident, &existingName); err == nil {
		existingExists = true
	}

	op := "created"
	if existingExists {
		if strings.TrimRight(existingResident.String, "/128") == strings.TrimRight(hoa, "/128") &&
			existingName.String == cardName {
			return "unchanged", nil
		}
		op = "updated"
	}

	// PK = spatial_prefix → UPSERT 用 ON CONFLICT (spatial_prefix)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO cards (spatial_prefix, card_type, card_name, dns_short_name, resident_id, is_active, enabled_at)
		VALUES ($1::inet, $2, $3, $4, $5::inet, TRUE, NOW())
		ON CONFLICT (spatial_prefix) DO UPDATE SET
			card_name      = COALESCE(EXCLUDED.card_name, cards.card_name),
			dns_short_name = COALESCE(EXCLUDED.dns_short_name, cards.dns_short_name),
			resident_id    = EXCLUDED.resident_id,
			is_active      = TRUE,
			enabled_at     = COALESCE(cards.enabled_at, NOW()),
			updated_at     = NOW()
	`, prefixStr, cardType, cardName, shortName, hoa)
	cardID := prefixStr // card_id ≡ spatial_prefix
	if err != nil {
		return "", err
	}

	// DDNS register best-effort（启动 reconcile 时全量推一遍，failed = log warn 继续）
	if s.ddns != nil && shortName != "" {
		tenantSlot, _, _, _, _, _, _, derr := spatial.SlotsOf(prefix.Addr())
		if derr == nil {
			zone := ddns.ZoneForTenant(tenantSlot, s.owlDomain)
			if err := s.ddns.RegisterCardName(ctx, prefix, shortName, zone); err != nil {
				s.logger.Warn("reconcile: DDNS RegisterCardName failed",
					zap.String("prefix", prefixStr), zap.String("zone", zone), zap.Error(err))
			}
		}
	}

	// publish CloudEvent（reconcile 期不区分 admission/transfer；用统一 op="reconcile"）
	if s.publisher != nil {
		_ = s.publisher.PublishCardResidentChanged(ctx, tenantPrefix, cardID, "reconcile", "", hoa, prefixStr)
	}
	return op, nil
}

// clearStaleResidentCards 给 unit /80 范围内 cards：若 resident_id 指向的 resident 已无 active resident_unit @ 该 prefix，则 NULL 掉。
// card_id ≡ spatial_prefix（PK），所以扫到 (prefix, oldHoA) 直接 UPDATE WHERE spatial_prefix=$prefix。
func (s *CardSyncService) clearStaleResidentCards(ctx context.Context, unitPrefix string) (int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.spatial_prefix::text, c.resident_id::text
		  FROM cards c
		 WHERE c.spatial_prefix <<= $1::INET
		   AND c.resident_id IS NOT NULL
		   AND NOT EXISTS (
		     SELECT 1 FROM resident_unit ru
		      WHERE ru.resident_id = c.resident_id
		        AND ru.spatial_prefix = c.spatial_prefix
		        AND ru.valid_to IS NULL
		   )
	`, unitPrefix)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type stale struct{ prefix, oldHoA string }
	var list []stale
	for rows.Next() {
		var v stale
		if err := rows.Scan(&v.prefix, &v.oldHoA); err == nil {
			list = append(list, v)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	cleared := 0
	for _, v := range list {
		if _, err := s.db.ExecContext(ctx,
			`UPDATE cards SET resident_id = NULL, updated_at = NOW() WHERE spatial_prefix = $1::INET`, v.prefix); err != nil {
			s.logger.Warn("clearStale: UPDATE failed", zap.String("prefix", v.prefix), zap.Error(err))
			continue
		}
		// tenant prefix derive from unit prefix
		tenantPrefix := ""
		_ = s.db.QueryRowContext(ctx, `SELECT set_masklen($1::inet, 48)::text`, v.prefix).Scan(&tenantPrefix)
		if s.publisher != nil {
			_ = s.publisher.PublishCardResidentChanged(ctx, tenantPrefix, v.prefix, "reconcile_discharge", v.oldHoA, "", v.prefix)
		}
		cleared++
	}
	return cleared, nil
}

// SyncDeviceCards v2 no-op。
// DeviceCard 概念在 v2 改为 card_type='device' /128，业务层不再批量创建。
func (s *CardSyncService) SyncDeviceCards(ctx context.Context, tenantID string) (int, error) {
	return 0, nil
}

// EnsureDeviceCard v2 no-op（同上）。
func (s *CardSyncService) EnsureDeviceCard(ctx context.Context, tenantID string, device domain.Device) {
}

// CleanupDeviceCard v2 no-op。
func (s *CardSyncService) CleanupDeviceCard(ctx context.Context, tenantID, deviceID string) {
}

// CreateAllCards v2 no-op；启动 reconcile job 不再批量重建。
func (s *CardSyncService) CreateAllCards(ctx context.Context, tenantID string) error {
	s.logger.Debug("CardSyncService.CreateAllCards skipped (v2 event-driven; pending Phase F)",
		zap.String("tenant_id", tenantID))
	return nil
}

// CleanupOrphanCards 删除其物理锚点已不存在的孤儿卡（spatial_prefix 下无 device）。
// v2: repo.ListOrphanCards 已基于 LPM。删除前 emit cardChange deleted。
func (s *CardSyncService) CleanupOrphanCards(ctx context.Context) (int, error) {
	affected, err := s.cardRepo.ListOrphanCards(ctx)
	if err != nil {
		return 0, err
	}
	cleaned := 0
	for _, a := range affected {
		if err := s.emitCardChange(ctx, a); err != nil {
			s.logger.Warn("emit orphan card deleted failed", zap.String("card_id", a.CardID), zap.Error(err))
		}
		if err := s.cardRepo.DeleteCard(a.TenantID, a.CardID); err != nil {
			s.logger.Warn("delete orphan card failed", zap.String("card_id", a.CardID), zap.Error(err))
			continue
		}
		cleaned++
	}
	if cleaned > 0 {
		s.logger.Info("orphan cards cleaned on startup", zap.Int("count", cleaned))
	}
	return cleaned, nil
}

// ClearAllCards 清理全局所有卡片记录（推送 deleted CloudEvent，供 cardagg 清缓存）。
// Deprecated: v2 仅供灾难恢复使用。
func (s *CardSyncService) ClearAllCards(ctx context.Context) error {
	affected, err := s.cardRepo.ListAllCardsForClear(ctx)
	if err != nil {
		return err
	}
	for _, a := range affected {
		if err := s.emitCardChange(ctx, a); err != nil {
			s.logger.Warn("emit card deleted before clear failed", zap.String("card_id", a.CardID), zap.Error(err))
		}
	}
	return s.cardRepo.ClearAllCards()
}

// RecalcAllCardsAlarmState v2 no-op（cards 表无 alarm counter 列；实时聚合）。
func (s *CardSyncService) RecalcAllCardsAlarmState(ctx context.Context, db interface{}) (ok, fail int, err error) {
	return 0, 0, nil
}

// emitCardChange 发送 config.card.* CloudEvent。
func (s *CardSyncService) emitCardChange(ctx context.Context, a domain.CardSyncAffected) error {
	branchID := ""
	if a.CardID != "" {
		bid, err := s.cardRepo.GetBranchIDByCard(ctx, a.TenantID, a.CardID)
		if err != nil {
			s.logger.Warn("get branch_id for emitCardChange failed",
				zap.String("card_id", a.CardID), zap.Error(err))
		} else {
			branchID = bid
		}
	}

	uids := a.AffectedDeviceUIDs
	if len(uids) == 0 && a.CardID != "" {
		uids = s.cardRepo.GetDeviceUIDsForCard(a.TenantID, a.CardID)
	}
	extra := map[string]interface{}{"op": a.Op}
	if len(uids) > 0 {
		extra["affected_device_uids"] = uids
	}
	if err := s.publisher.PublishCardChangeMessageWithExtra(ctx, a.TenantID, a.CardID, a.UnitID, branchID, extra); err != nil {
		return err
	}
	if s.realtime != nil && branchID != "" {
		s.realtime.UpdateByBranch(ctx, a.TenantID, branchID)
	}
	return nil
}

// ========== 全局 CardSync 入口 ==========

var globalCardSync *CardSyncService

// InitGlobalCardSync 在 main.go 中 CardSyncService 创建后调用一次。
func InitGlobalCardSync(cs *CardSyncService) {
	globalCardSync = cs
}

// SyncUnitCards v2 no-op；保留全局入口避免 caller 改动。
func SyncUnitCards(ctx context.Context, tenantID, unitID string) {
	if globalCardSync == nil {
		return
	}
	// v2: 事件驱动；批量同步已 retire
	_ = tenantID
	_ = unitID
}

// EnsureDeviceCardGlobal v2 no-op。
func EnsureDeviceCardGlobal(ctx context.Context, tenantID string, device domain.Device) {
	if globalCardSync != nil {
		globalCardSync.EnsureDeviceCard(ctx, tenantID, device)
	}
}

// CleanupDeviceCardGlobal v2 no-op。
func CleanupDeviceCardGlobal(ctx context.Context, tenantID, deviceID string) {
	if globalCardSync != nil {
		globalCardSync.CleanupDeviceCard(ctx, tenantID, deviceID)
	}
}
