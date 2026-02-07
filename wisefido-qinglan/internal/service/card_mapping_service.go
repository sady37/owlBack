package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonRedis "owl-common/redis"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CardMappingService 管理 deviceUID → 完整映射信息的查询服务
// 订阅 config:card:stream 中的 config.card 事件
// 维护以下 Redis 结构：
//  1. owl:cache:branch:{branchID} → Hash
//     Field: deviceUID (MQTT 标准字段)
//     Value: CardDeviceInfo 的 JSON 序列化
//  2. owl:lookup:device:{deviceUID} → String
//     Value: branchID（用于第一步快速查询）
//
// 查询流程（两步）：
// Step 1: GET owl:lookup:device:{deviceUID} → 获取 branchID
// Step 2: HGET owl:cache:branch:{branchID} {deviceUID} → 获取完整的 CardDeviceInfo
type CardMappingService struct {
	redisClient *redis.Client
	cardRepo    repository.CardRepository
	logger      *zap.Logger
}

// NewCardMappingService 创建卡片映射服务
func NewCardMappingService(redisClient *redis.Client, cardRepo repository.CardRepository, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		redisClient: redisClient,
		cardRepo:    cardRepo,
		logger:      logger,
	}
}

// GetCardIDByDeviceUID 通过 deviceUID 获取完整映射信息
// 两步查询（仅需 deviceUID，无需 tenantID）：
// Step 1: GET owl:lookup:device:{deviceUID} → 获取 branchID
// Step 2: HGET owl:cache:branch:{branchID} {deviceUID} → 获取完整的 CardDeviceInfo
// 如果缓存缺失，返回错误（需要在消费层触发 ReloadBranchCache 回填）
func (s *CardMappingService) GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*repository.CardDeviceInfo, error) {
	if deviceUID == "" {
		return nil, fmt.Errorf("device_uid is required")
	}

	// Step 1: 查询 branchID
	lookupKey := fmt.Sprintf("owl:lookup:device:%s", deviceUID)
	branchID, err := s.redisClient.Get(ctx, lookupKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to query Redis lookup key",
			zap.String("device_uid", deviceUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query Redis lookup: %w", err)
	}

	if branchID == "" {
		return nil, fmt.Errorf("device_uid not found in cache: %s", deviceUID)
	}

	// Step 2: 从分支缓存获取完整信息
	cacheKey := fmt.Sprintf("owl:cache:branch:%s", branchID)
	jsonValue, err := s.redisClient.HGet(ctx, cacheKey, deviceUID).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to query Redis branch cache",
			zap.String("branch_id", branchID),
			zap.String("device_uid", deviceUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query Redis cache: %w", err)
	}

	if jsonValue == "" {
		return nil, fmt.Errorf("device_uid not found in branch cache: %s", deviceUID)
	}

	// 反序列化 JSON
	var cdi repository.CardDeviceInfo
	if err := json.Unmarshal([]byte(jsonValue), &cdi); err != nil {
		s.logger.Warn("Failed to unmarshal CardDeviceInfo",
			zap.String("device_uid", deviceUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal CardDeviceInfo: %w", err)
	}

	s.logger.Debug("Found CardDeviceInfo in Redis",
		zap.String("device_uid", deviceUID),
		zap.String("device_id", cdi.DeviceID),
		zap.String("card_id", cdi.CardID),
		zap.String("branch_id", cdi.BranchID),
		zap.String("tenant_id", cdi.TenantID))

	return &cdi, nil
}

// HandleCardChanged 处理卡片配置变更事件（统一入口）
// 支持创建、更新、删除三种操作
// 策略：失效该租户分支的缓存并重建（卡片更新频率低，避免脏数据）
func (s *CardMappingService) HandleCardChanged(ctx context.Context, data map[string]interface{}) error {
	cardData := &commonRedis.CardChangeData{}
	if err := s.parseCardChangeData(data, cardData); err != nil {
		return fmt.Errorf("failed to parse card change data: %w", err)
	}

	s.logger.Info("Handling card change event",
		zap.String("tenant_id", cardData.TenantID),
		zap.String("card_id", cardData.CardID),
		zap.String("branch_id", cardData.BranchID))

	// 重建该租户分支的缓存
	return s.ReloadBranchCache(ctx, cardData.TenantID, cardData.BranchID)
}

// ReloadBranchCache 失效该租户分支缓存并重建（卡片更新频率低，避免脏数据）
// 清理策略：
// 1. 查询 Redis 中该分支的旧设备列表（用于清理孤立的全局索引）
// 2. Repository 查询 DB：该租户该分支的所有设备-卡片映射
// 3. Service Pipeline：
//   - 删除旧的分支 Hash 缓存
//   - 删除旧设备的全局查找索引（避免孤立指针）
//   - 写入新数据到 cache 和 lookup
func (s *CardMappingService) ReloadBranchCache(ctx context.Context, tenantID string, branchID string) error {
	if tenantID == "" || branchID == "" {
		return fmt.Errorf("tenant_id and branch_id are required")
	}

	if s.cardRepo == nil {
		return fmt.Errorf("card repository not set")
	}

	s.logger.Info("Reloading branch cache",
		zap.String("tenant_id", tenantID),
		zap.String("branch_id", branchID))

	branchKey := fmt.Sprintf("owl:cache:branch:%s", branchID)

	// 1. 获取旧的设备 UID 列表（用于清理全局索引）
	oldDeviceUIDs, err := s.redisClient.HKeys(ctx, branchKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to get old device UIDs for cleanup",
			zap.String("branch_id", branchID),
			zap.Error(err))
		// 不中断流程，继续重建缓存
	}

	// 2. Repository 层：查询该租户该分支的所有设备-卡片映射
	mappingRecords, err := s.cardRepo.GetDeviceCardMappingsByBranch(ctx, tenantID, branchID)
	if err != nil {
		s.logger.Error("Failed to query cards and devices from repository",
			zap.String("tenant_id", tenantID),
			zap.String("branch_id", branchID),
			zap.Error(err))
		return fmt.Errorf("failed to query card-device mapping by branch: %w", err)
	}

	s.logger.Debug("Queried devices from DB",
		zap.String("tenant_id", tenantID),
		zap.String("branch_id", branchID),
		zap.Int("device_count", len(mappingRecords)))

	// 3. Service 层：准备 Redis Pipeline
	pipe := s.redisClient.Pipeline()

	// 3.1 清除旧的分支 Hash 记录
	pipe.Del(ctx, branchKey)

	// 3.2 清理旧设备的全局查找索引（避免孤立指针）
	// 只清理从该分支移除的设备，保留新设备的索引
	newDeviceUIDs := make(map[string]bool)
	for _, device := range mappingRecords {
		newDeviceUIDs[device.DeviceUID] = true
	}

	for _, oldUID := range oldDeviceUIDs {
		if !newDeviceUIDs[oldUID] {
			// 这个设备已从分支删除，清理它的全局索引
			lookupKey := fmt.Sprintf("owl:lookup:device:%s", oldUID)
			pipe.Del(ctx, lookupKey)
			s.logger.Debug("Cleaning up orphaned lookup index",
				zap.String("device_uid", oldUID),
				zap.String("branch_id", branchID))
		}
	}

	// 3.3 批量写入新数据
	for _, device := range mappingRecords {
		// 序列化为 JSON
		jsonData, err := json.Marshal(device)
		if err != nil {
			s.logger.Warn("Failed to marshal CardDeviceInfo",
				zap.String("device_uid", device.DeviceUID),
				zap.Error(err))
			continue
		}

		// Pipeline 操作：写入分支缓存
		// Key: owl:cache:branch:{branch_id}, Field: {device_uid}, Value: JSON
		pipe.HSet(ctx, branchKey, device.DeviceUID, jsonData)

		// Pipeline 操作：写入/更新全局查找索引
		// Key: owl:lookup:device:{device_uid}, Value: {branch_id}
		// 设置 TTL 为 24 小时，防止长期孤立
		lookupKey := fmt.Sprintf("owl:lookup:device:%s", device.DeviceUID)
		pipe.Set(ctx, lookupKey, branchID, 24*time.Hour)
	}

	// 4. 执行 Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("Failed to execute Redis pipeline",
			zap.String("tenant_id", tenantID),
			zap.String("branch_id", branchID),
			zap.Error(err))
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	s.logger.Info("Successfully updated cache for branch",
		zap.String("tenant_id", tenantID),
		zap.String("branch_id", branchID),
		zap.Int("device_count", len(mappingRecords)),
		zap.Int("orphaned_lookups_cleaned", len(oldDeviceUIDs)-len(mappingRecords)))

	return nil
}

// parseCardChangeData 解析卡片变更数据
func (s *CardMappingService) parseCardChangeData(data map[string]interface{}, cardData *commonRedis.CardChangeData) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal card change data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, cardData); err != nil {
		return fmt.Errorf("failed to unmarshal card change data: %w", err)
	}

	if cardData.TenantID == "" || cardData.CardID == "" {
		return fmt.Errorf("tenant_id and card_id are required in card change data")
	}

	return nil
}
