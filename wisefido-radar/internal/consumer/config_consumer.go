package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	rediscommon "owl-common/redis"
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	EventType string `json:"event_type"` // "alarm_device_updated" 或 "alarm_cloud_updated"
	TenantID  string `json:"tenant_id"`
	DeviceUID string `json:"device_uid,omitempty"` // 仅当 event_type 为 "alarm_device_updated" 时存在
	Timestamp int64  `json:"timestamp"`
}

// ConfigConsumer 配置变更消费者
// 订阅 config:change:stream，处理配置变更事件，清除报警使能缓存
type ConfigConsumer struct {
	config      *config.Config
	redisClient *redis.Client
	deviceRepo  *repository.DeviceRepository
	logger      *zap.Logger
}

// NewConfigConsumer 创建配置消费者
func NewConfigConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	deviceRepo *repository.DeviceRepository,
	logger *zap.Logger,
) *ConfigConsumer {
	return &ConfigConsumer{
		config:      cfg,
		redisClient: redisClient,
		deviceRepo:  deviceRepo,
		logger:      logger,
	}
}

// Start 启动配置消费者
func (c *ConfigConsumer) Start(ctx context.Context) error {
	stream := "config:change:stream"
	consumerGroup := "wisefido-radar:config"
	consumerName := "config-consumer-1"

	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, consumerGroup); err != nil {
		c.logger.Warn("Failed to create consumer group for config stream, will retry",
			zap.String("stream", stream),
			zap.Error(err),
		)
		// 继续执行，不中断服务
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
		zap.String("device_uid", event.DeviceUID),
	)

	// 根据事件类型处理
	switch event.EventType {
	case "alarm_device_updated":
		return c.handleAlarmDeviceUpdated(ctx, event)
	case "alarm_cloud_updated":
		// alarm_cloud_updated 影响所有设备，但 wisefido-radar 只关心设备级别的配置
		// 这里可以选择清除所有缓存，或者不做处理（因为设备配置会单独发送 alarm_device_updated）
		c.logger.Debug("Received alarm_cloud_updated event, no action needed for radar service",
			zap.String("tenant_id", event.TenantID),
		)
		return nil
	default:
		c.logger.Warn("Unknown config change event type",
			zap.String("event_type", event.EventType),
		)
		return nil // 未知事件类型，不返回错误
	}
}

// handleAlarmDeviceUpdated 处理设备报警配置更新事件
func (c *ConfigConsumer) handleAlarmDeviceUpdated(ctx context.Context, event ConfigChangeEvent) error {
	if event.DeviceUID == "" {
		c.logger.Warn("alarm_device_updated event missing device_uid",
			zap.String("tenant_id", event.TenantID),
		)
		return nil
	}

	c.logger.Info("Alarm device config updated, clearing cache",
		zap.String("tenant_id", event.TenantID),
		zap.String("device_uid", event.DeviceUID),
	)

	// 清除该设备的报警使能缓存
	// 下次查询时会重新从数据库加载最新配置
	c.deviceRepo.ClearAlarmEnablementCache(event.TenantID, event.DeviceUID)

	return nil
}
