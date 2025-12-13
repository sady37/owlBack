package consumer

import (
	"encoding/json"
	"fmt"
	"time"
	"wisefido-sensor-fusion/internal/config"
	"wisefido-sensor-fusion/internal/models"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
)

// CacheManager Redis 缓存管理器
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

// UpdateRealtimeData 更新实时数据缓存
func (c *CacheManager) UpdateRealtimeData(cardID string, data *models.RealtimeData) error {
	// 构建缓存键
	key := fmt.Sprintf("%s%s:realtime", c.config.Fusion.Cache.RealtimeKeyPrefix, cardID)
	
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal realtime data: %w", err)
	}
	
	// 写入 Redis（设置 TTL）
	err = c.redisClient.Set(
		c.redisClient.Context(),
		key,
		jsonData,
		time.Duration(c.config.Fusion.Cache.RealtimeTTL)*time.Second,
	).Err()
	
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	
	c.logger.Debug("Updated realtime cache",
		zap.String("card_id", cardID),
		zap.String("key", key),
	)
	
	return nil
}

