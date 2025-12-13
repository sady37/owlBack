package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	rediscommon "owl-common/redis"
)

// EventConsumer 事件消费者
type EventConsumer struct {
	redisClient *redis.Client
	cardCreator *aggregator.CardCreator
	cardRepo    *repository.CardRepository
	logger      *zap.Logger
	stream      string
	groupName   string
	consumerName string
	batchSize   int64
}

// CardEvent 卡片事件
type CardEvent struct {
	EventType string                 `json:"event_type"`
	TenantID  string                 `json:"tenant_id"`
	UnitID    string                 `json:"unit_id"`
	BedID     string                 `json:"bed_id,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	ResidentID string                `json:"resident_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewEventConsumer 创建事件消费者
func NewEventConsumer(
	redisClient *redis.Client,
	cardCreator *aggregator.CardCreator,
	cardRepo *repository.CardRepository,
	logger *zap.Logger,
	stream string,
	groupName string,
	consumerName string,
	batchSize int64,
) *EventConsumer {
	return &EventConsumer{
		redisClient:  redisClient,
		cardCreator:  cardCreator,
		cardRepo:     cardRepo,
		logger:       logger,
		stream:       stream,
		groupName:    groupName,
		consumerName: consumerName,
		batchSize:    batchSize,
	}
}

// Start 启动事件消费者
func (c *EventConsumer) Start(ctx context.Context) error {
	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, c.stream, c.groupName); err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	c.logger.Info("Event consumer started",
		zap.String("stream", c.stream),
		zap.String("consumer_group", c.groupName),
		zap.String("consumer_name", c.consumerName),
	)

	// 消费事件（带指数退避）
	backoffDuration := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumeEvents(ctx); err != nil {
				c.logger.Error("Failed to consume events",
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
				// 成功时重置退避时间
				backoffDuration = time.Second
			}
		}
	}
}

// consumeEvents 消费事件
func (c *EventConsumer) consumeEvents(ctx context.Context) error {
	// 从 Redis Streams 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		c.stream,
		c.groupName,
		c.consumerName,
		c.batchSize,
	)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}

	// 处理消息
	for _, msg := range messages {
		if err := c.processEvent(ctx, msg); err != nil {
			c.logger.Error("Failed to process event",
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		} else {
			// 处理成功后确认消息
			if err := c.ackMessage(ctx, msg.ID); err != nil {
				c.logger.Warn("Failed to ack message",
					zap.String("message_id", msg.ID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// processEvent 处理单个事件
func (c *EventConsumer) processEvent(ctx context.Context, msg rediscommon.StreamMessage) error {
	// 解析事件
	event, err := c.parseEvent(msg)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	c.logger.Info("Processing card event",
		zap.String("event_type", event.EventType),
		zap.String("tenant_id", event.TenantID),
		zap.String("unit_id", event.UnitID),
	)

	// 根据事件类型触发卡片重新计算
	switch event.EventType {
	case "device.bound", "device.unbound", "device.monitoring_changed":
		// 设备绑定/解绑/监护状态变化
		if event.UnitID != "" {
			return c.cardCreator.CreateCardsForUnit(event.TenantID, event.UnitID)
		} else if event.BedID != "" {
			// 如果只有 bed_id，需要查询 unit_id
			unitID, err := c.getUnitIDByBedID(event.TenantID, event.BedID)
			if err != nil {
				return fmt.Errorf("failed to get unit_id by bed_id: %w", err)
			}
			return c.cardCreator.CreateCardsForUnit(event.TenantID, unitID)
		}

	case "resident.bound", "resident.unbound", "resident.status_changed":
		// 住户绑定/解绑/状态变化
		if event.UnitID != "" {
			return c.cardCreator.CreateCardsForUnit(event.TenantID, event.UnitID)
		} else if event.BedID != "" {
			// 如果只有 bed_id，需要查询 unit_id
			unitID, err := c.getUnitIDByBedID(event.TenantID, event.BedID)
			if err != nil {
				return fmt.Errorf("failed to get unit_id by bed_id: %w", err)
			}
			return c.cardCreator.CreateCardsForUnit(event.TenantID, unitID)
		}

	case "bed.status_changed", "bed.device_count_changed":
		// 床位状态变化（ActiveBed ↔ NonActiveBed）
		if event.BedID != "" {
			unitID, err := c.getUnitIDByBedID(event.TenantID, event.BedID)
			if err != nil {
				return fmt.Errorf("failed to get unit_id by bed_id: %w", err)
			}
			return c.cardCreator.CreateCardsForUnit(event.TenantID, unitID)
		}

	case "unit.info_changed":
		// 单元信息变化（地址、名称等）
		if event.UnitID != "" {
			return c.cardCreator.CreateCardsForUnit(event.TenantID, event.UnitID)
		}

	default:
		c.logger.Warn("Unknown event type",
			zap.String("event_type", event.EventType),
		)
		return nil
	}

	return nil
}

// parseEvent 解析事件消息
func (c *EventConsumer) parseEvent(msg rediscommon.StreamMessage) (*CardEvent, error) {
	// 尝试从 data 字段解析 JSON
	if dataStr, ok := msg.Values["data"].(string); ok {
		var event CardEvent
		if err := json.Unmarshal([]byte(dataStr), &event); err == nil {
			return &event, nil
		}
	}

	// 如果 data 字段不存在，直接从 Values 解析
	event := &CardEvent{}
	
	if eventType, ok := msg.Values["event_type"].(string); ok {
		event.EventType = eventType
	}
	if tenantID, ok := msg.Values["tenant_id"].(string); ok {
		event.TenantID = tenantID
	}
	if unitID, ok := msg.Values["unit_id"].(string); ok {
		event.UnitID = unitID
	}
	if bedID, ok := msg.Values["bed_id"].(string); ok {
		event.BedID = bedID
	}
	if deviceID, ok := msg.Values["device_id"].(string); ok {
		event.DeviceID = deviceID
	}
	if residentID, ok := msg.Values["resident_id"].(string); ok {
		event.ResidentID = residentID
	}

	if event.EventType == "" || event.TenantID == "" {
		return nil, fmt.Errorf("invalid event: missing event_type or tenant_id")
	}

	return event, nil
}

// getUnitIDByBedID 根据 bed_id 获取 unit_id
func (c *EventConsumer) getUnitIDByBedID(tenantID, bedID string) (string, error) {
	return c.cardRepo.GetUnitIDByBedID(tenantID, bedID)
}

// ackMessage 确认消息
func (c *EventConsumer) ackMessage(ctx context.Context, messageID string) error {
	return c.redisClient.XAck(ctx, c.stream, c.groupName, messageID).Err()
}

