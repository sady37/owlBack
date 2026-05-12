package service

import (
	"context"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

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

// CreateCardsForUnit v2 no-op；真正的 INSERT 由 resident lifecycle 路径完成。
func (s *CardSyncService) CreateCardsForUnit(ctx context.Context, tenantID, unitID string) (*CardUpdateStats, error) {
	s.logger.Debug("CardSyncService.CreateCardsForUnit skipped (v2 event-driven; pending Phase F)",
		zap.String("tenant_id", tenantID), zap.String("unit_id", unitID))
	return &CardUpdateStats{}, nil
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
