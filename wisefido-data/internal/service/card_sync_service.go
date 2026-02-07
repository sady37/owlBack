package service

import (
	"context"
	"encoding/json"
	"fmt"

	"wisefido-data/internal/card"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

// VitalFocusStaticCacheWriter 写 VitalFocusCard 静态缓存（可选）
type VitalFocusStaticCacheWriter interface {
	WriteCardStatic(ctx context.Context, tenantID, cardID string) error
	DeleteCardStatic(ctx context.Context, cardID string) error
}

// UserCardsCacheInvalidator 用户卡片缓存失效（可选）
type UserCardsCacheInvalidator interface {
	InvalidateByTenantBranch(ctx context.Context, tenantID, branchID string) error
}

// CardSyncService 卡片同步服务：只依赖 DB card repo + publisher；写 DB 后发 config，并写 VitalFocusCard 静态缓存
type CardSyncService struct {
	cardRepo       *repository.PostgresCardRepository
	creator        *card.CardCreator
	publisher      *publisher.ConfigPublisher
	vitalCache     VitalFocusStaticCacheWriter // 可选：发 config 后写静态缓存供前端快速返回
	userCardsCache UserCardsCacheInvalidator   // 可选：card 变更时失效 user:cards 缓存
	logger         *zap.Logger
}

// NewCardSyncService 创建卡片同步服务（仅 DB card repo + publisher）
func NewCardSyncService(
	cardRepo *repository.PostgresCardRepository,
	publisher *publisher.ConfigPublisher,
	logger *zap.Logger,
) *CardSyncService {
	creator := card.NewCardCreator(cardRepo, logger)
	return &CardSyncService{
		cardRepo:  cardRepo,
		creator:   creator,
		publisher: publisher,
		logger:    logger,
	}
}

// SetVitalFocusStaticCache 设置 VitalFocusCard 静态缓存写入器（可选）
func (s *CardSyncService) SetVitalFocusStaticCache(c VitalFocusStaticCacheWriter) {
	s.vitalCache = c
}

// SetUserCardsCache 设置用户卡片缓存失效器（可选），card 变更时调用 InvalidateByTenantBranch
func (s *CardSyncService) SetUserCardsCache(c UserCardsCacheInvalidator) {
	s.userCardsCache = c
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

// CreateAllCards 为所有单元创建/更新卡片（启动时调用，参考 card-manage CreateAllCards）
func (s *CardSyncService) CreateAllCards(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		s.logger.Warn("tenant_id is empty, skipping card sync")
		return fmt.Errorf("tenant_id is required")
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

	// Step 1: 写静态缓存（DB 已写完，直接写缓存）
	if s.vitalCache != nil {
		if err := s.vitalCache.WriteCardStatic(ctx, a.TenantID, a.CardID); err != nil {
			s.logger.Warn("Failed to write card static cache",
				zap.String("card_id", a.CardID),
				zap.Error(err),
			)
			// 不中断流程，缓存写入失败不阻止
		}
	}

	// Step 2: 发送消息到 config:card:stream（缓存已准备好）
	// 消费者收到消息后根据 branch_id 全量查询卡片，自动同步 create/update/delete
	if err := s.publisher.PublishCardChangeMessage(ctx, a.TenantID, a.CardID, a.UnitID, branchID); err != nil {
		return err
	}

	// Step 3: 失效动态缓存（按 tenant+branch）
	if s.userCardsCache != nil && a.UnitID != "" {
		invalidateBranchID := branchID
		if invalidateBranchID == "" {
			invalidateBranchID = "_"
		}
		if err := s.userCardsCache.InvalidateByTenantBranch(ctx, a.TenantID, invalidateBranchID); err == nil {
			s.logger.Debug("invalidate user cards cache", zap.String("tenant_id", a.TenantID), zap.String("branch_id", invalidateBranchID))
		} else {
			s.logger.Warn("invalidate user cards cache failed", zap.String("tenant_id", a.TenantID), zap.String("branch_id", invalidateBranchID), zap.Error(err))
		}
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
