package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	rediscommon "owl-common/redis"
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	EventType string `json:"event_type"` // "alarm_device_updated" 或 "alarm_cloud_updated"
	TenantID  string `json:"tenant_id"`
	DeviceID  string `json:"device_id,omitempty"` // 仅当 event_type 为 "alarm_device_updated" 时存在
	Timestamp int64  `json:"timestamp"`
}

// ConfigConsumer 配置变更消费者
// 订阅 config:change:stream，处理配置变更事件
type ConfigConsumer struct {
	config          *config.Config
	redisClient     *redis.Client
	alarmDeviceRepo *repository.AlarmDeviceRepository
	logger          *zap.Logger
}

// NewConfigConsumer 创建配置消费者
func NewConfigConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	alarmDeviceRepo *repository.AlarmDeviceRepository,
	logger *zap.Logger,
) *ConfigConsumer {
	return &ConfigConsumer{
		config:          cfg,
		redisClient:     redisClient,
		alarmDeviceRepo: alarmDeviceRepo,
		logger:          logger,
	}
}

// Start 启动配置消费者
func (c *ConfigConsumer) Start(ctx context.Context) error {
	stream := "config:change:stream"
	consumerGroup := "wisefido-card-aggregator:config"
	consumerName := "config-consumer-1"

	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, consumerGroup); err != nil {
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	c.logger.Info("Config consumer started",
		zap.String("consumer_group", consumerGroup),
		zap.String("consumer_name", consumerName),
		zap.String("stream", stream),
	)

	// 启动消费循环
	backoffDuration := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumeStream(ctx, stream, consumerGroup, consumerName); err != nil {
				c.logger.Error("Failed to consume config stream",
					zap.Error(err),
					zap.Duration("backoff", backoffDuration),
				)

				// 指数退避
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoffDuration):
					backoffDuration *= 2
					if backoffDuration > maxBackoff {
						backoffDuration = maxBackoff
					}
				}
			} else {
				backoffDuration = time.Second
			}
		}
	}
}

// consumeStream 消费配置变更流
func (c *ConfigConsumer) consumeStream(ctx context.Context, stream, consumerGroup, consumerName string) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		stream,
		consumerGroup,
		consumerName,
		10, // 批量读取 10 条消息
	)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}

	// 处理消息
	for _, msg := range messages {
		if err := c.processMessage(ctx, msg); err != nil {
			c.logger.Error("Failed to process config change message",
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}

	return nil
}

// processMessage 处理单条配置变更消息
func (c *ConfigConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage) error {
	// 解析消息数据
	var dataStr string
	if val, ok := msg.Values["data"]; ok {
		if str, ok := val.(string); ok {
			dataStr = str
		} else {
			return fmt.Errorf("invalid data format in message")
		}
	} else {
		return fmt.Errorf("missing data field in message")
	}

	// 解析 JSON
	var event ConfigChangeEvent
	if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
		c.logger.Error("Failed to parse config change event",
			zap.String("stream_id", msg.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal config change event: %w", err)
	}

	c.logger.Debug("Processing config change event",
		zap.String("event_type", event.EventType),
		zap.String("tenant_id", event.TenantID),
		zap.String("device_id", event.DeviceID),
	)

	// 根据事件类型处理
	switch event.EventType {
	case "alarm_device_updated":
		return c.handleAlarmDeviceUpdated(ctx, event)
	case "alarm_cloud_updated":
		return c.handleAlarmCloudUpdated(ctx, event)
	default:
		c.logger.Warn("Unknown config change event type",
			zap.String("event_type", event.EventType),
		)
		return nil // 未知事件类型，不返回错误
	}
}

// handleAlarmDeviceUpdated 处理设备报警配置更新事件
func (c *ConfigConsumer) handleAlarmDeviceUpdated(ctx context.Context, event ConfigChangeEvent) error {
	if event.DeviceID == "" {
		c.logger.Warn("alarm_device_updated event missing device_id",
			zap.String("tenant_id", event.TenantID),
		)
		return nil
	}

	c.logger.Info("Alarm device config updated",
		zap.String("tenant_id", event.TenantID),
		zap.String("device_id", event.DeviceID),
	)

	// 注意：wisefido-card-aggregator 使用懒加载策略
	// 当配置变更时，不需要立即更新缓存
	// 下次查询时会自动从数据库重新加载最新配置
	// 这样可以减少内存使用，并且配置变更频率低，1秒延迟可接受

	// 如果需要立即清除缓存，可以在这里实现
	// 例如：清除 alarmDeviceRepo 的内存缓存（如果存在）

	return nil
}

// handleAlarmCloudUpdated 处理云端报警配置更新事件
func (c *ConfigConsumer) handleAlarmCloudUpdated(ctx context.Context, event ConfigChangeEvent) error {
	c.logger.Info("Alarm cloud config updated",
		zap.String("tenant_id", event.TenantID),
	)

	// 注意：wisefido-card-aggregator 使用懒加载策略
	// 当配置变更时，不需要立即更新缓存
	// 下次查询时会自动从数据库重新加载最新配置

	return nil
}
