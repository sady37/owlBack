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
type CacheManager struct {
	config      *config.Config
	kv          KVStore
	logger      *zap.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(
	cfg *config.Config,
	kv KVStore,
	logger *zap.Logger,
) *CacheManager {
	return &CacheManager{
		config:      cfg,
		kv:          kv,
		logger:      logger,
	}
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

