package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/models"

	"go.uber.org/zap"
)

// CacheManager Redis 缓存管理器（用于数据聚合）
//
// 输出 key、value 的 JSON 与 TTL 定义见：docs/OUTPUT_FORMAT.md
type CacheManager struct {
	config *config.Config
	kv     KVStore
	logger *zap.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(
	cfg *config.Config,
	kv KVStore,
	logger *zap.Logger,
) *CacheManager {
	return &CacheManager{
		config: cfg,
		kv:     kv,
		logger: logger,
	}
}

// UpdateRealtimeDataCache 更新实时数据缓存（由 IoTStreamConsumer 调用）
func (c *CacheManager) UpdateRealtimeDataCache(ctx context.Context, cardID string, realtimeData *models.RealtimeData) error {
	key := fmt.Sprintf("vital-focus:card:%s:realtime", cardID)

	// 序列化数据
	jsonData, err := json.Marshal(realtimeData)
	if err != nil {
		return fmt.Errorf("failed to marshal realtime data: %w", err)
	}

	// 写入 Redis（设置 TTL 为 10 秒，与 full 缓存一致）
	err = c.kv.Set(ctx, key, string(jsonData), 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("Updated realtime data cache",
		zap.String("card_id", cardID),
		zap.String("key", key),
	)

	return nil
}

// UpdateFullCardCache 更新完整的卡片缓存
func (c *CacheManager) UpdateFullCardCache(ctx context.Context, cardID string, vitalCard *models.VitalFocusCard) error {
	key := fmt.Sprintf("vital-focus:card:%s:full", cardID)

	// 序列化数据
	jsonData, err := json.Marshal(vitalCard)
	if err != nil {
		return fmt.Errorf("failed to marshal vital card: %w", err)
	}

	// 写入 Redis（设置 TTL 为 10 秒）
	err = c.kv.Set(ctx, key, string(jsonData), 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("Updated full card cache",
		zap.String("card_id", cardID),
		zap.String("key", key),
	)

	return nil
}
