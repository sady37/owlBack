package consumer

import (
	"context"
	"fmt"
	"time"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// CacheConsumer 缓存消费者（轮询 Redis 实时数据缓存）
type CacheConsumer struct {
	config   *config.Config
	cache    *CacheManager
	cardRepo *repository.CardRepository
	logger   *zap.Logger
	tenantID string // 租户ID（从配置或环境变量获取）
}

// NewCacheConsumer 创建缓存消费者
func NewCacheConsumer(
	cfg *config.Config,
	cache *CacheManager,
	cardRepo *repository.CardRepository,
	logger *zap.Logger,
	tenantID string,
) *CacheConsumer {
	return &CacheConsumer{
		config:   cfg,
		cache:    cache,
		cardRepo: cardRepo,
		logger:   logger,
		tenantID: tenantID,
	}
}

// Start 启动消费者（轮询模式）
func (c *CacheConsumer) Start(ctx context.Context, evaluator Evaluator) error {
	c.logger.Info("Cache consumer started",
		zap.String("tenant_id", c.tenantID),
		zap.Int("poll_interval", c.config.Alarm.PollInterval),
	)

	ticker := time.NewTicker(time.Duration(c.config.Alarm.PollInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	if err := c.evaluateAllCards(ctx, evaluator); err != nil {
		c.logger.Error("Failed to evaluate cards on startup",
			zap.Error(err),
		)
	}

	// 定期轮询
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Cache consumer stopped")
			return nil
		case <-ticker.C:
			if err := c.evaluateAllCards(ctx, evaluator); err != nil {
				c.logger.Error("Failed to evaluate cards",
					zap.Error(err),
				)
				// 继续执行，不中断
			}
		}
	}
}

// evaluateAllCards 评估所有卡片
func (c *CacheConsumer) evaluateAllCards(ctx context.Context, evaluator Evaluator) error {
	// 1. 从 PostgreSQL 获取所有卡片
	cards, err := c.cardRepo.GetAllCards(c.tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all cards: %w", err)
	}

	c.logger.Debug("Evaluating cards",
		zap.Int("card_count", len(cards)),
	)

	// 2. 批量评估（按配置的批量大小）
	batchSize := c.config.Alarm.Evaluation.BatchSize
	for i := 0; i < len(cards); i += batchSize {
		end := i + batchSize
		if end > len(cards) {
			end = len(cards)
		}

		batch := cards[i:end]
		if err := c.evaluateBatch(ctx, batch, evaluator); err != nil {
			c.logger.Error("Failed to evaluate batch",
				zap.Int("batch_start", i),
				zap.Int("batch_end", end),
				zap.Error(err),
			)
			// 继续处理下一批，不中断
		}
	}

	return nil
}

// evaluateBatch 批量评估卡片
func (c *CacheConsumer) evaluateBatch(ctx context.Context, cards []repository.CardInfo, evaluator Evaluator) error {
	for _, card := range cards {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读取实时数据
		realtimeData, err := c.cache.GetRealtimeData(card.CardID)
		if err != nil {
			// 如果实时数据不存在，跳过（可能是卡片还没有数据）
			c.logger.Debug("Realtime data not found for card",
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			continue
		}

		// 评估报警
		alarms, err := evaluator.Evaluate(c.tenantID, card, realtimeData)
		if err != nil {
			c.logger.Error("Failed to evaluate card",
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			continue
		}

		// 更新报警缓存（只更新活跃的报警）
		if len(alarms) > 0 {
			// 过滤出活跃的报警（alarm_status = 'active'）
			activeAlarms := make([]models.AlarmEvent, 0)
			for _, alarm := range alarms {
				if alarm.AlarmStatus == "active" {
					activeAlarms = append(activeAlarms, alarm)
				}
			}

			if len(activeAlarms) > 0 {
				if err := c.cache.UpdateAlarmCache(card.CardID, activeAlarms); err != nil {
					c.logger.Error("Failed to update alarm cache",
						zap.String("card_id", card.CardID),
						zap.Error(err),
					)
				}
			}
		}
	}

	return nil
}

// Evaluator 报警评估器接口
type Evaluator interface {
	// Evaluate 评估卡片数据，返回报警事件列表
	Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error)
}
