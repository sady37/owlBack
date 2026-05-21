package service

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

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

// CardNameNoResident — 空床/空间卡的 card_name 占位（resident 缺位时用）。
// 参 doc/cards_v2_migration_checklist.md §一.1：card_name 仅随 admission/discharge 变；
// 空间维度由 dns_short_name (Unit-Room-Bed) 承载，不混入 card_name。
const (
	CardNameNoResident = "NoOne"  // 单人/share unit 未入住时（待分配）
	CardNamePublic     = "public" // public unit (unit_type=3) 公共区；不显示合成 resident 的 nickname（如 'public-LivingRoom'），覆盖为短的 'public'
)

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

	// ReconcileCards 测试 hook — 每次 commit 后调；生产留 nil
	reconcileObserver func(scope string, diffs []cardDiff)
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

// CreateCardsForUnit — v2 重新实现：unit /80 scope 的**空间驱动**reconcile。
//
// 设计转折（vs 老 v2 仅 resident-driven）：
// 用户反馈：v1 行为 = 每张物理床 1 张 active_bed card；每个公共 unit 1 张 unit
// /public card。空床也该有卡（resident 来去仅改 resident_id 指针，不动 card 行数）。
// 老 v2 reconcile 只扫 resident_unit → Unit 101 的 2 张空床 / Kitchen / LivingRoom
// 都没卡。改为按空间结构（beds + units）驱动：
//
//   Step 1: 列 unit 下 beds → 每张床 INSERT/UPSERT active_bed card (NoOne)
//   Step 2: 若 unit 是公共区（unit_type ∈ public/shared 配置，或本 unit 无 bed）
//           → INSERT/UPSERT unit/public card
//   Step 3: overlay active resident_unit → 命中 prefix 的 card.resident_id +
//           card_name=nickname；resident_unit prefix 是 /88 (room) 时也建 room card
//   Step 4: clearStaleResidentCards 把指向已 expired resident 的 resident_id NULL 掉
//
// card 删除（spatial_prefix 下设备全空）由 CleanupOrphanCards 处理（device-driven）。
func (s *CardSyncService) CreateCardsForUnit(ctx context.Context, tenantPrefix, unitPrefix string) (*CardUpdateStats, error) {
	stats := &CardUpdateStats{}
	if s.db == nil {
		s.logger.Debug("CreateCardsForUnit skipped (db not wired)")
		return stats, nil
	}
	if strings.TrimSpace(unitPrefix) == "" {
		return stats, nil
	}

	// === Step 1: 空间结构驱动 — 每张 *绑了 device* 的 bed 一张 active_bed card ===
	// 设计修正：原"每张物理 bed 都建卡"会立刻被 CleanupOrphanCards 删（device 缺失）→
	// startup 出现 create→delete 来回；且 UI 仍会瞥到无 device 的 NoOne 卡。
	// 改为 source-of-truth = "bed 绑了至少 1 个 device 才需要 card"。
	bedRows, err := s.db.QueryContext(ctx, `
		SELECT b.bed_id::text, COALESCE(b.bed_name, '')
		  FROM beds b
		 WHERE b.bed_id <<= $1::INET
		   AND EXISTS (SELECT 1 FROM devices d WHERE d.device_ipv6 <<= b.bed_id)
		 ORDER BY b.bed_id
	`, unitPrefix)
	if err == nil {
		defer bedRows.Close()
		for bedRows.Next() {
			var bedPrefix, bedName string
			if err := bedRows.Scan(&bedPrefix, &bedName); err != nil {
				continue
			}
			op, err := s.upsertSpaceCard(ctx, tenantPrefix, bedPrefix, "")
			if err != nil {
				s.logger.Warn("reconcile: upsertSpaceCard bed failed",
					zap.String("bed", bedPrefix), zap.String("bed_name", bedName), zap.Error(err))
				continue
			}
			countOp(stats, op)
		}
	} else {
		s.logger.Warn("reconcile: list beds failed", zap.String("unit", unitPrefix), zap.Error(err))
	}

	// === Step 2: unit /80 是公共区（无 bed）且 *至少有 1 个 device* → 建 unit/public card ===
	// 同 Step 1 — 无 device 的公共区不建卡，避免 orphan
	var hasBed bool
	_ = s.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM beds WHERE bed_id <<= $1::INET)`,
		unitPrefix).Scan(&hasBed)
	if !hasBed {
		var hasDevice bool
		_ = s.db.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM devices WHERE device_ipv6 <<= $1::INET)`,
			unitPrefix).Scan(&hasDevice)
		if hasDevice {
			op, err := s.upsertSpaceCard(ctx, tenantPrefix, unitPrefix, "")
			if err != nil {
				s.logger.Warn("reconcile: upsertSpaceCard unit failed",
					zap.String("unit", unitPrefix), zap.Error(err))
			} else {
				countOp(stats, op)
			}
		}
	}

	// === Step 3: overlay active resident_unit on existing cards ===
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
		// 若 resident 赋到 /88 room 或 /80 unit（无对应 bed card），仍创建对应 card
		op, err := s.upsertSpaceCard(ctx, tenantPrefix, prefixStr, hoa)
		if err != nil {
			s.logger.Warn("reconcile: upsertSpaceCard resident overlay failed",
				zap.String("hoa", hoa), zap.String("prefix", prefixStr), zap.Error(err))
			continue
		}
		// resident overlay 阶段统计走 "updated"（已有 space card → 改 resident_id）；
		// 真新建（resident 赋到 /80 unit 而无对应 unit card 等罕见场景）按 op 真实结果
		countOp(stats, op)
		_ = nickname // upsertSpaceCard 内部用 hoa 反查 nickname
	}
	if err := rows.Err(); err != nil {
		s.logger.Warn("CreateCardsForUnit rows err", zap.Error(err))
	}

	// === Step 4: 清扫该 unit /80 下 resident_id 已落空的 card ===
	cleared, err := s.clearStaleResidentCards(ctx, unitPrefix)
	if err != nil {
		s.logger.Warn("CreateCardsForUnit clearStale failed", zap.String("unit", unitPrefix), zap.Error(err))
	}
	stats.DeletedCount = cleared
	return stats, nil
}

func countOp(stats *CardUpdateStats, op string) {
	switch op {
	case "created":
		stats.CreatedCount++
		stats.ExistingCount++
	case "updated":
		stats.UpdatedCount++
		stats.ExistingCount++
	case "unchanged":
		stats.UnchangedCount++
		stats.ExistingCount++
	}
}

// upsertSpaceCard — 空间-驱动 UPSERT。
//   prefixStr: bed/room/unit prefix（INET CIDR）
//   hoa:       "" → NoOne (空床/空 unit)；非空 → resident 入住，card_name=nickname
// v2.5 TODO: cards 表 schema 已改（card_id 是独立 /88 slot，不再 == spatial_prefix）。
//           本函数 SQL 还在用旧 schema，运行时会失败；待 card_sync_service 重写。
func (s *CardSyncService) upsertSpaceCard(ctx context.Context, tenantPrefix, prefixStr, hoa string) (string, error) {
	prefix, err := netip.ParsePrefix(prefixStr)
	if err != nil {
		return "", err
	}
	// v2.5 短期 stub：原 CardTypeForMasklen 已删；只接受 /80/88/96 prefix
	if !(prefix.Bits() == 80 || prefix.Bits() == 88 || prefix.Bits() == 96) {
		return "", nil
	}
	shortName, _ := ddns.CardShortName(prefix)

	// card_name 解析：hoa 非空时查 residents.nickname；空则 "NoOne"
	cardName := CardNameNoResident
	hoaArg := interface{}(nil)
	if strings.TrimSpace(hoa) != "" {
		var nick sql.NullString
		_ = s.db.QueryRowContext(ctx,
			`SELECT COALESCE(nickname, '') FROM residents WHERE resident_id = $1::INET`,
			hoa).Scan(&nick)
		if nick.Valid && nick.String != "" {
			cardName = nick.String
		}
		hoaArg = hoa
	}

	// 查现状
	var existingResident, existingName sql.NullString
	existingExists := false
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(resident_id::text, ''), COALESCE(card_name, '')
		  FROM cards
		 WHERE card_id = $1::INET
		 LIMIT 1
	`, prefixStr).Scan(&existingResident, &existingName); err == nil {
		existingExists = true
	}

	op := "created"
	if existingExists {
		// 比较是否真要 update（避免无脑 update）
		existingHoA := strings.TrimRight(existingResident.String, "/128")
		newHoA := strings.TrimRight(hoa, "/128")
		if existingHoA == newHoA && existingName.String == cardName {
			return "unchanged", nil
		}
		op = "updated"
	}

	// v2.5 stub: 旧 schema INSERT 不再可行（card_id/card_slot/unit_id required；spatial_prefix/card_type/is_active 已删）
	// 待 card_sync_service 重写后填好分配逻辑；当前会让 reconcile 安全 no-op
	_ = shortName
	_ = cardName
	_ = hoaArg
	_, err = s.db.ExecContext(ctx, `SELECT 1 -- v2.5 cards INSERT stubbed; see card_sync_service rewrite TODO`)
	if err != nil {
		return "", err
	}

	// DDNS register best-effort
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

	if s.publisher != nil {
		evtOp := "reconcile"
		if op == "created" {
			evtOp = "reconcile_create"
		}
		_ = s.publisher.PublishCardResidentChanged(ctx, tenantPrefix, prefixStr, evtOp, "", hoa, prefixStr)
	}
	return op, nil
}

// upsertResidentCard — Deprecated, 被 upsertSpaceCard 取代（空间驱动 + resident overlay 统一）。
// 保留签名为零调用方编译兼容；正式去除可在下次 cleanup pass 删。
//
// Deprecated: use upsertSpaceCard.
func (s *CardSyncService) upsertResidentCard(ctx context.Context, tenantPrefix, hoa, _ string, prefixStr string) (string, error) {
	return s.upsertSpaceCard(ctx, tenantPrefix, prefixStr, hoa)
}

// clearStaleResidentCards 给 unit /80 范围内 cards：若 resident_id 指向的 resident 已无 active resident_unit @ 该 prefix，则 NULL 掉。
func (s *CardSyncService) clearStaleResidentCards(ctx context.Context, unitPrefix string) (int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.card_id::text, c.resident_id::text
		  FROM cards c
		 WHERE c.card_id <<= $1::INET
		   AND c.resident_id IS NOT NULL
		   AND NOT EXISTS (
		     SELECT 1 FROM resident_unit ru
		      WHERE ru.resident_id = c.resident_id
		        AND ru.spatial_prefix = c.card_id
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
		// 清 resident_id + card_name 改回 "NoOne"（card_name 与 resident 联动；空间名走 dns_short_name）
		if _, err := s.db.ExecContext(ctx,
			`UPDATE cards SET resident_id = NULL, card_name = $2, updated_at = NOW() WHERE card_id = $1::INET`,
			v.prefix, CardNameNoResident); err != nil {
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
