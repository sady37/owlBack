package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"owl-common/card"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

// CardSyncService 卡片同步服务：写 DB 后发 config + 通知 realtime
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


// CreateCardsForUnit 为指定单元创建/更新卡片（写 DB），同步后发送 config.card.*，并写 VitalFocusCard 静态缓存
func (s *CardSyncService) CreateCardsForUnit(ctx context.Context, tenantID, unitID string) (*card.CardUpdateStats, error) {
	s.logger.Info("Creating cards for unit",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
	)

	s.cardRepo.ClearRecorded()
	stats, err := s.creator.CreateCardsForUnit(tenantID, unitID)
	if err != nil {
		s.logger.Error("Failed to create cards for unit",
			zap.String("tenant_id", tenantID),
			zap.String("unit_id", unitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create cards for unit: %w", err)
	}

	s.logger.Info("Cards synced for unit",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Int("existing_count", stats.ExistingCount),
		zap.Int("created_count", stats.CreatedCount),
		zap.Int("updated_count", stats.UpdatedCount),
		zap.Int("deleted_count", stats.DeletedCount),
		zap.Int("unchanged_count", stats.UnchangedCount),
	)

	affected := s.cardRepo.GetRecordedAndClear()
	for _, a := range affected {
		if err := s.emitCardChange(ctx, a); err != nil {
			s.logger.Warn("Failed to emit config.card event",
				zap.String("op", a.Op),
				zap.String("card_id", a.CardID),
				zap.Error(err),
			)
		}
	}

	return stats, nil
}

// RecalcAllCardsAlarmState 启动时按 alarm_events 重算并写回所有 cards（unhandled_alarm_*、pop_alarm_*），与 alarm_events 一致
func (s *CardSyncService) RecalcAllCardsAlarmState(ctx context.Context, db *sql.DB) (ok, fail int, err error) {
	rows, err := db.QueryContext(ctx, `SELECT card_id, tenant_id FROM cards`)
	if err != nil {
		return 0, 0, fmt.Errorf("query cards: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, tid string
		if err := rows.Scan(&cid, &tid); err != nil {
			fail++
			continue
		}
		if _, err := card.RecalcCardAlarmState(ctx, db, cid, tid); err != nil {
			s.logger.Debug("RecalcCardAlarmState", zap.String("card_id", cid), zap.Error(err))
			fail++
		} else {
			ok++
		}
	}
	return ok, fail, rows.Err()
}

// CreateAllCards 为所有单元创建/更新卡片（启动时调用，参考 card-manage CreateAllCards）
func (s *CardSyncService) CreateAllCards(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		s.logger.Warn("tenant_id is empty, skipping card sync")
		return fmt.Errorf("tenant_id is required")
	}

	// 清理所有现有卡片记录（全局清理），避免漏删除
	if err := s.ClearAllCards(ctx); err != nil {
		s.logger.Error("Failed to clear all cards before sync",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to clear all cards: %w", err)
	}

	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all units: %w", err)
	}

	s.logger.Info("Syncing cards for all units",
		zap.String("tenant_id", tenantID),
		zap.Int("unit_count", len(unitIDs)),
	)

	successCount := 0
	errorCount := 0
	for _, unitID := range unitIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := s.CreateCardsForUnit(ctx, tenantID, unitID)
			if err != nil {
				s.logger.Error("Failed to create cards for unit",
					zap.String("tenant_id", tenantID),
					zap.String("unit_id", unitID),
					zap.Error(err),
				)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	s.logger.Info("Completed card sync for all units",
		zap.String("tenant_id", tenantID),
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
		zap.Int("total_units", len(unitIDs)),
	)
	return nil
}

// ClearAllCards 清理全局所有卡片记录（删除前对每张卡推送 config:card:stream deleted，供 cardagg 清缓存）
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

func (s *CardSyncService) emitCardChange(ctx context.Context, a domain.CardSyncAffected) error {
	// 获取 branch_id（从 cards 表直接查询，性能最佳）
	branchID := ""
	if a.CardID != "" {
		bid, err := s.cardRepo.GetBranchIDByCard(ctx, a.TenantID, a.CardID)
		if err != nil {
			s.logger.Warn("get branch_id for emitCardChange failed", zap.String("card_id", a.CardID), zap.Error(err))
		} else {
			branchID = bid
		}
	}

	uids := a.AffectedDeviceUIDs
	if len(uids) == 0 && a.TenantID != "" && a.CardID != "" {
		uids = s.cardRepo.GetDeviceUIDsForCard(a.TenantID, a.CardID)
	}
	extra := map[string]interface{}{"op": a.Op}
	if len(uids) > 0 {
		extra["affected_device_uids"] = uids
	}
	// Step 1: 发送消息到 config:card:stream（携带 op + 可选 affected_device_uids）
	if err := s.publisher.PublishCardChangeMessageWithExtra(ctx, a.TenantID, a.CardID, a.UnitID, branchID, extra); err != nil {
		return err
	}

	// Step 2: 通知 realtime 更新受影响用户的 CardList
	if s.realtime != nil && branchID != "" {
		s.realtime.UpdateByBranch(ctx, a.TenantID, branchID)
	}

	return nil
}

// parseDeviceItemsFromJSON 从 cards.devices JSON 中提取完整 device 信息
// 转换为消息中需要的 DeviceItemForMessage 格式
func parseDeviceItemsFromJSON(devicesJSON []byte) []rediscommon.DeviceItemForMessage {
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil
	}
	items := make([]rediscommon.DeviceItemForMessage, 0, len(devices))
	for _, d := range devices {
		item := rediscommon.DeviceItemForMessage{}
		if id, ok := d["device_id"].(string); ok && id != "" {
			item.DeviceID = id
		}
		if uid, ok := d["device_uid"].(string); ok {
			item.DeviceUID = uid
		}
		if code, ok := d["device_code"].(string); ok {
			item.DeviceCode = code
		}
		if name, ok := d["device_name"].(string); ok {
			item.DeviceName = name
		}
		if dt, ok := d["device_type"]; ok {
			item.DeviceType = dt
		}
		// 只有当至少有 device_id 才加入
		if item.DeviceID != "" {
			items = append(items, item)
		}
	}
	return items
}

// parseDeviceIDsFromJSON 从 cards.devices JSON 中提取 device_id 列表（与 postgres_card.UpdateCardAlarmCounts 一致）
func parseDeviceIDsFromJSON(devicesJSON []byte) []string {
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil
	}
	ids := make([]string, 0, len(devices))
	for _, d := range devices {
		if id, ok := d["device_id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
