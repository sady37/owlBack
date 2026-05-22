package service

import (
	"context"
	"database/sql"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"owl-common/ddns"

	"go.uber.org/zap"
)

// CardSyncService — cards 表唯一写入入口。
//
// 范式：事件驱动 per-tenant ReconcileCards（card_reconcile.go）。
// 触发：
//   - resident lifecycle (residents repo onResidentUnitChange hook)
//   - device bind/unbind (devices repo onDeviceChange hook)
//   - bed/room/unit CRUD (unit_service.syncCardsForUnit)
//   - 启动期每 tenant 一次（main.go 启动 reconcile job）

// CardNameNoResident / CardNamePublic — card_name 占位语义。
//   card_name 仅随 admission / discharge / nickname 变；
//   空间维度由 card_dns (Unit-Room-Bed 短码) 承载，不混入 card_name。
const (
	CardNameNoResident = "NoOne"  // 单人/share unit 未入住时（待分配）
	CardNamePublic     = "public" // public unit (unit_type=3) 公共区
)

type CardSyncService struct {
	cardRepo  *repository.PostgresCardRepository
	publisher *publisher.ConfigPublisher
	realtime  *CardRealtimeService
	logger    *zap.Logger

	db        *sql.DB
	ddns      *ddns.Client
	owlDomain string

	// ReconcileCards 测试 hook — 每次 commit 后调；生产留 nil。
	reconcileObserver func(scope string, diffs []cardDiff)
}

func NewCardSyncService(
	cardRepo *repository.PostgresCardRepository,
	publisher *publisher.ConfigPublisher,
	realtime *CardRealtimeService,
	logger *zap.Logger,
) *CardSyncService {
	return &CardSyncService{
		cardRepo:  cardRepo,
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
// 仅 main.go 启动期注入一次。
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

// CleanupOrphanCards 删除其物理锚点已不存在的孤儿卡（spatial_prefix 下无 device）。
// 删除前 emit cardChange deleted CloudEvent，供 cardagg 清 Redis card:state:*。
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

// ClearAllCards 全删卡片记录（推送 deleted CloudEvent，供 cardagg 清缓存）。灾难恢复入口。
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

// globalCardSync — 包级单例，供同 package 内 hook caller (unit_service / device_service) 调用。
var globalCardSync *CardSyncService

// InitGlobalCardSync 在 main.go CardSyncService 创建后调用一次。
func InitGlobalCardSync(cs *CardSyncService) {
	globalCardSync = cs
}
