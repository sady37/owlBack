package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CardMappingService 管理 deviceID:cardID 的映射
// 订阅 config:card:stream 中的 config.card.created/updated/deleted 事件
// 维护以下映射：
// 1. Redis: device:card:mapping -> {deviceID: cardID, ...}（便于所有 qinglan 实例访问）
// 2. 本地内存缓存: 快速访问
type CardMappingService struct {
	redisClient *redis.Client
	logger      *zap.Logger

	// 本地缓存：key 为 tenantID:deviceID，value 为 cardID
	// 格式：tenant1:device1 -> cardID1
	localCache *sync.Map

	// 缓存锁，用于防止并发更新冲突
	cacheLock *sync.RWMutex
}

// DeviceItemForMessage 卡片中的设备信息
type DeviceItemForMessage struct {
	DeviceID   string      `json:"device_id"`
	DeviceUID  string      `json:"device_uid"`
	DeviceCode string      `json:"device_code,omitempty"`
	DeviceName string      `json:"device_name,omitempty"`
	DeviceType interface{} `json:"device_type,omitempty"`
}

// CardChangeData 卡片变更消息中的数据
type CardChangeData struct {
	TenantID  string                 `json:"tenant_id"`
	CardID    string                 `json:"card_id"`
	UnitID    string                 `json:"unit_id"`
	BranchID  string                 `json:"branch_id"`
	Timestamp int64                  `json:"timestamp_ms"`
	Devices   []DeviceItemForMessage `json:"devices,omitempty"`
}

// NewCardMappingService 创建卡片映射服务
func NewCardMappingService(redisClient *redis.Client, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		redisClient: redisClient,
		logger:      logger,
		localCache:  &sync.Map{},
		cacheLock:   &sync.RWMutex{},
	}
}

// GetCardIDByDeviceID 通过 deviceID 获取 cardID（优先查本地缓存，然后查 Redis）
func (s *CardMappingService) GetCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if tenantID == "" || deviceID == "" {
		return "", fmt.Errorf("tenant_id and device_id are required")
	}

	key := s.getLocalCacheKey(tenantID, deviceID)

	// 1. 优先查本地缓存
	if cardID, ok := s.localCache.Load(key); ok {
		s.logger.Debug("Found card ID in local cache",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("card_id", cardID.(string)))
		return cardID.(string), nil
	}

	// 2. 查 Redis 缓存
	redisKey := s.getRedisKey(tenantID)
	mapping, err := s.redisClient.HGetAll(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to query Redis for device:card mapping",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err))
		// 不返回错误，继续尝试本地缓存
	}

	if len(mapping) > 0 {
		// 将 Redis 中的映射加载到本地缓存
		for devID, cardID := range mapping {
			localKey := s.getLocalCacheKey(tenantID, devID)
			s.localCache.Store(localKey, cardID)
		}

		if cardID, ok := mapping[deviceID]; ok {
			s.logger.Debug("Found card ID in Redis cache",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("card_id", cardID))
			return cardID, nil
		}
	}

	// 3. 本地缓存和 Redis 都没有找到
	return "", fmt.Errorf("card ID not found for device %s", deviceID)
}

// HandleCardCreated 处理卡片创建事件
func (s *CardMappingService) HandleCardCreated(ctx context.Context, data map[string]interface{}) error {
	cardData := &CardChangeData{}
	if err := s.parseCardChangeData(data, cardData); err != nil {
		return fmt.Errorf("failed to parse card created data: %w", err)
	}

	s.logger.Info("Handling card created event",
		zap.String("tenant_id", cardData.TenantID),
		zap.String("card_id", cardData.CardID),
		zap.Int("device_count", len(cardData.Devices)))

	return s.updateDeviceCardMapping(ctx, cardData.TenantID, cardData.CardID, cardData.Devices, "created")
}

// HandleCardUpdated 处理卡片更新事件
func (s *CardMappingService) HandleCardUpdated(ctx context.Context, data map[string]interface{}) error {
	cardData := &CardChangeData{}
	if err := s.parseCardChangeData(data, cardData); err != nil {
		return fmt.Errorf("failed to parse card updated data: %w", err)
	}

	s.logger.Info("Handling card updated event",
		zap.String("tenant_id", cardData.TenantID),
		zap.String("card_id", cardData.CardID),
		zap.Int("device_count", len(cardData.Devices)))

	// 更新前，先清除旧的映射（如果设备有变化）
	// 这里先清除所有旧映射，然后重新建立
	s.clearOldCardMapping(ctx, cardData.TenantID, cardData.CardID)

	return s.updateDeviceCardMapping(ctx, cardData.TenantID, cardData.CardID, cardData.Devices, "updated")
}

// HandleCardDeleted 处理卡片删除事件
func (s *CardMappingService) HandleCardDeleted(ctx context.Context, data map[string]interface{}) error {
	cardData := &CardChangeData{}
	if err := s.parseCardChangeData(data, cardData); err != nil {
		return fmt.Errorf("failed to parse card deleted data: %w", err)
	}

	s.logger.Info("Handling card deleted event",
		zap.String("tenant_id", cardData.TenantID),
		zap.String("card_id", cardData.CardID))

	return s.clearOldCardMapping(ctx, cardData.TenantID, cardData.CardID)
}

// updateDeviceCardMapping 更新 deviceID:cardID 映射
func (s *CardMappingService) updateDeviceCardMapping(
	ctx context.Context,
	tenantID, cardID string,
	devices []DeviceItemForMessage,
	operation string,
) error {
	if len(devices) == 0 {
		s.logger.Debug("No devices to map for card",
			zap.String("card_id", cardID),
			zap.String("operation", operation))
		return nil
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	// 构建本次更新的 deviceID 列表
	updatedDeviceIDs := make([]string, 0, len(devices))

	// 更新 Redis 映射
	redisKey := s.getRedisKey(tenantID)
	for _, device := range devices {
		if device.DeviceID == "" {
			continue
		}
		updatedDeviceIDs = append(updatedDeviceIDs, device.DeviceID)

		// 更新本地缓存
		localKey := s.getLocalCacheKey(tenantID, device.DeviceID)
		s.localCache.Store(localKey, cardID)

		// 更新 Redis（使用 HSet）
		if err := s.redisClient.HSet(ctx, redisKey, device.DeviceID, cardID).Err(); err != nil {
			s.logger.Error("Failed to update Redis device:card mapping",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", device.DeviceID),
				zap.String("card_id", cardID),
				zap.Error(err))
			// 继续处理其他设备
		}
	}

	// 设置 Redis key 的过期时间（24 小时）
	if err := s.redisClient.Expire(ctx, redisKey, 24*time.Hour).Err(); err != nil {
		s.logger.Warn("Failed to set expiration for Redis device:card mapping",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
	}

	s.logger.Info("Updated device:card mappings",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.Strings("device_ids", updatedDeviceIDs),
		zap.String("operation", operation))

	return nil
}

// clearOldCardMapping 清除卡片的旧映射
// 通过查询 Redis 找到该卡片关联的所有 deviceID，然后删除它们的映射
func (s *CardMappingService) clearOldCardMapping(ctx context.Context, tenantID, cardID string) error {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	redisKey := s.getRedisKey(tenantID)

	// 从 Redis 中查询所有映射
	mapping, err := s.redisClient.HGetAll(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to query Redis for device:card mapping when clearing",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.Error(err))
		// 即使查询失败，也继续清除本地缓存
	}

	// 找到属于该卡片的所有 deviceID，并删除
	deletedDeviceIDs := make([]string, 0)
	for deviceID, cID := range mapping {
		if cID == cardID {
			// 从本地缓存删除
			localKey := s.getLocalCacheKey(tenantID, deviceID)
			s.localCache.Delete(localKey)

			// 从 Redis 删除
			if err := s.redisClient.HDel(ctx, redisKey, deviceID).Err(); err != nil {
				s.logger.Warn("Failed to delete Redis device:card mapping",
					zap.String("tenant_id", tenantID),
					zap.String("device_id", deviceID),
					zap.String("card_id", cardID),
					zap.Error(err))
			}
			deletedDeviceIDs = append(deletedDeviceIDs, deviceID)
		}
	}

	if len(deletedDeviceIDs) > 0 {
		s.logger.Info("Cleared old device:card mappings",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.Strings("deleted_device_ids", deletedDeviceIDs))
	}

	return nil
}

// parseCardChangeData 解析卡片变更数据
func (s *CardMappingService) parseCardChangeData(data map[string]interface{}, cardData *CardChangeData) error {
	// 使用 JSON 转换确保类型正确
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

// getLocalCacheKey 获取本地缓存键
func (s *CardMappingService) getLocalCacheKey(tenantID, deviceID string) string {
	return tenantID + ":" + deviceID
}

// getRedisKey 获取 Redis 键
func (s *CardMappingService) getRedisKey(tenantID string) string {
	return "device:card:mapping:" + tenantID
}
