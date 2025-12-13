package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/models"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
)

// CacheManager Redis 缓存管理器（用于报警服务）
type CacheManager struct {
	config      *config.Config
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(
	cfg *config.Config,
	redisClient *redis.Client,
	logger *zap.Logger,
) *CacheManager {
	return &CacheManager{
		config:      cfg,
		redisClient: redisClient,
		logger:      logger,
	}
}

// GetRealtimeData 从 Redis 读取实时数据
func (c *CacheManager) GetRealtimeData(cardID string) (*models.RealtimeData, error) {
	// 构建缓存键
	key := fmt.Sprintf("%s%s%s", 
		c.config.Alarm.Cache.RealtimeKeyPrefix, 
		cardID, 
		c.config.Alarm.Cache.RealtimeSuffix,
	)

	// 从 Redis 读取
	ctx := context.Background()
	val, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("realtime data not found for card: %s", cardID)
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	// 反序列化
	var data models.RealtimeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal realtime data: %w", err)
	}

	return &data, nil
}

// UpdateAlarmCache 更新报警缓存
func (c *CacheManager) UpdateAlarmCache(cardID string, alarms []models.AlarmEvent) error {
	// 构建缓存键
	key := fmt.Sprintf("%s%s%s",
		c.config.Alarm.Cache.AlarmKeyPrefix,
		cardID,
		c.config.Alarm.Cache.AlarmSuffix,
	)

	// 序列化数据
	jsonData, err := json.Marshal(alarms)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm data: %w", err)
	}

	// 写入 Redis（设置 TTL）
	ctx := context.Background()
	err = c.redisClient.Set(
		ctx,
		key,
		jsonData,
		time.Duration(c.config.Alarm.Cache.AlarmTTL)*time.Second,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to set alarm cache: %w", err)
	}

	c.logger.Debug("Updated alarm cache",
		zap.String("card_id", cardID),
		zap.String("key", key),
		zap.Int("alarm_count", len(alarms)),
	)

	return nil
}

// GetAllCardIDs 获取所有卡片的 ID（通过扫描 Redis 键）
// 注意：这个方法效率较低，建议后续优化为从 PostgreSQL 查询
func (c *CacheManager) GetAllCardIDs(ctx context.Context) ([]string, error) {
	// 构建匹配模式
	pattern := fmt.Sprintf("%s*%s",
		c.config.Alarm.Cache.RealtimeKeyPrefix,
		c.config.Alarm.Cache.RealtimeSuffix,
	)

	// 扫描所有匹配的键
	var cardIDs []string
	iter := c.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// 提取 card_id（去掉前缀和后缀）
		cardID := key[len(c.config.Alarm.Cache.RealtimeKeyPrefix):]
		cardID = cardID[:len(cardID)-len(c.config.Alarm.Cache.RealtimeSuffix)]
		cardIDs = append(cardIDs, cardID)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	return cardIDs, nil
}

