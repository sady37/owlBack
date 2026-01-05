package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	rediscommon "owl-common/redis"
	"strings"
	"time"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// EventConsumer 事件消费者（订阅 Redis Streams）
type EventConsumer struct {
	config      *config.Config
	redisClient *redis.Client
	cache       *CacheManager
	cardRepo    *repository.CardRepository
	logger      *zap.Logger
	tenantID    string
}

// NewEventConsumer 创建事件消费者
func NewEventConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	cache *CacheManager,
	cardRepo *repository.CardRepository,
	logger *zap.Logger,
	tenantID string,
) *EventConsumer {
	return &EventConsumer{
		config:      cfg,
		redisClient: redisClient,
		cache:       cache,
		cardRepo:    cardRepo,
		logger:      logger,
		tenantID:    tenantID,
	}
}

// Start 启动事件消费者（订阅 Redis Streams）
func (c *EventConsumer) Start(ctx context.Context, evaluator Evaluator) error {
	stream := "iot:data:stream"
	consumerGroup := "wisefido-alarm-events"
	consumerName := "alarm-consumer-1"

	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, consumerGroup); err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	c.logger.Info("Event consumer started",
		zap.String("stream", stream),
		zap.String("consumer_group", consumerGroup),
		zap.String("consumer_name", consumerName),
		zap.String("tenant_id", c.tenantID),
	)

	// 持续消费消息
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Event consumer stopped")
			return nil
		default:
			// 从 Stream 读取消息
			messages, err := rediscommon.ReadFromStream(
				ctx,
				c.redisClient,
				stream,
				consumerGroup,
				consumerName,
				10, // batch size
			)
			if err != nil {
				c.logger.Error("Failed to read from stream",
					zap.Error(err),
				)
				time.Sleep(time.Second)
				continue
			}

			// 处理消息
			for _, msg := range messages {
				if err := c.processMessage(ctx, msg, evaluator); err != nil {
					c.logger.Error("Failed to process message",
						zap.String("stream_id", msg.ID),
						zap.Error(err),
					)
					// 继续处理下一条消息，不中断
				}
			}
		}
	}
}

// processMessage 处理单条消息
func (c *EventConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage, evaluator Evaluator) error {
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
	var iotData models.IoTDataMessage
	if err := json.Unmarshal([]byte(dataStr), &iotData); err != nil {
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}

	// 过滤：处理相关事件
	if iotData.EventType == nil {
		return nil // 跳过非事件消息
	}

	eventType := *iotData.EventType
	c.logger.Debug("Received event",
		zap.String("event_type", eventType),
		zap.String("device_id", iotData.DeviceID),
		zap.String("tenant_id", iotData.TenantID),
		zap.String("stream_id", msg.ID),
	)

	// 根据事件类型分发处理
	switch eventType {
	case "BED_LEFT":
		return c.handleBED_LEFT_Event(ctx, iotData, evaluator)
	case "ENTER_ROOM":
		return c.handleENTER_ROOM_Event(ctx, iotData, evaluator)
	case "LEFT_ROOM":
		return c.handleLEFT_ROOM_Event(ctx, iotData, evaluator)
	case "PERSON_COUNT_CHANGED":
		return c.handlePERSON_COUNT_CHANGED_Event(ctx, iotData, evaluator)
	default:
		return nil // 跳过其他事件
	}
}

// handleBED_LEFT_Event 处理 BED_LEFT 事件
func (c *EventConsumer) handleBED_LEFT_Event(
	ctx context.Context,
	iotData models.IoTDataMessage,
	evaluator Evaluator,
) error {
	// 1. 通过 device_id 查询 card_id
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		// 设备可能未绑定到卡片，忽略
		c.logger.Debug("Card not found for device",
			zap.String("device_id", iotData.DeviceID),
			zap.Error(err),
		)
		return nil
	}

	// 2. 前置条件检查
	if !c.checkPrerequisites(*cardInfo) {
		return nil // 前置条件不满足，退出
	}

	// 3. 从 Redis 缓存读取实时数据（不查DB）
	realtimeData, err := c.cache.GetRealtimeData(cardInfo.CardID)
	if err != nil {
		// 实时数据不存在，忽略
		c.logger.Debug("Realtime data not found for card",
			zap.String("card_id", cardInfo.CardID),
			zap.Error(err),
		)
		return nil
	}

	// 4. 处理床上跌落检测
	alarms, err := evaluator.Evaluate(iotData.TenantID, *cardInfo, realtimeData)
	if err != nil {
		return fmt.Errorf("failed to evaluate bed fall detection: %w", err)
	}

	// 5. 更新报警缓存（只更新活跃的报警）
	if len(alarms) > 0 {
		activeAlarms := make([]models.AlarmEvent, 0)
		for _, alarm := range alarms {
			if alarm.AlarmStatus == "active" {
				activeAlarms = append(activeAlarms, alarm)
			}
		}

		if len(activeAlarms) > 0 {
			if err := c.cache.UpdateAlarmCache(cardInfo.CardID, activeAlarms); err != nil {
				c.logger.Error("Failed to update alarm cache",
					zap.String("card_id", cardInfo.CardID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// checkPrerequisites 检查前置条件
// 返回 true 表示满足前置条件，可以继续处理
func (c *EventConsumer) checkPrerequisites(card repository.CardInfo) bool {
	// 1. 检查必须是 ActiveBed 卡片
	if card.CardType != "ActiveBed" {
		return false
	}

	// 2. 获取卡片绑定的设备列表
	devices, err := c.cardRepo.GetCardDevices(card.CardID)
	if err != nil {
		c.logger.Debug("Failed to get card devices",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
		return false
	}

	// 3. 检查是否有 Radar 设备
	hasRadar := false
	for _, device := range devices {
		if device.DeviceType == "Radar" {
			hasRadar = true
			break
		}
	}

	// 4. 如果没有 Radar，退出
	return hasRadar
}

// handleENTER_ROOM_Event 处理 ENTER_ROOM 事件（触发 Event 3 监测）
func (c *EventConsumer) handleENTER_ROOM_Event(
	ctx context.Context,
	iotData models.IoTDataMessage,
	evaluator Evaluator,
) error {
	// 1. 通过 device_id 查询 card_id
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		// 设备可能未绑定到卡片，忽略
		c.logger.Debug("Card not found for device",
			zap.String("device_id", iotData.DeviceID),
			zap.Error(err),
		)
		return nil
	}

	// 2. 检查是否是 bathroom
	if !c.checkBathroomForEvent3(*cardInfo) {
		return nil // 不是 bathroom，退出
	}

	// 3. 从 Redis 缓存读取实时数据
	realtimeData, err := c.cache.GetRealtimeData(cardInfo.CardID)
	if err != nil {
		// 实时数据不存在，忽略
		c.logger.Debug("Realtime data not found for card",
			zap.String("card_id", cardInfo.CardID),
			zap.Error(err),
		)
		return nil
	}

	// 4. 检查 track_id 数量（必须 == 1）
	if len(realtimeData.Postures) != 1 {
		c.logger.Debug("Not exactly 1 track_id, skipping Event 3",
			zap.String("card_id", cardInfo.CardID),
			zap.Int("track_count", len(realtimeData.Postures)),
		)
		return nil
	}

	// 5. 处理 Event 3 监测（进入事件触发）
	alarms, err := evaluator.Evaluate(iotData.TenantID, *cardInfo, realtimeData)
	if err != nil {
		return fmt.Errorf("failed to evaluate Event 3: %w", err)
	}

	// 6. 更新报警缓存
	if len(alarms) > 0 {
		activeAlarms := make([]models.AlarmEvent, 0)
		for _, alarm := range alarms {
			if alarm.AlarmStatus == "active" {
				activeAlarms = append(activeAlarms, alarm)
			}
		}

		if len(activeAlarms) > 0 {
			if err := c.cache.UpdateAlarmCache(cardInfo.CardID, activeAlarms); err != nil {
				c.logger.Error("Failed to update alarm cache",
					zap.String("card_id", cardInfo.CardID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// handleLEFT_ROOM_Event 处理 LEFT_ROOM 事件（停止 Event 3 监测）
func (c *EventConsumer) handleLEFT_ROOM_Event(
	ctx context.Context,
	iotData models.IoTDataMessage,
	evaluator Evaluator,
) error {
	// 1. 通过 device_id 查询 card_id
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		return nil // 设备可能未绑定到卡片，忽略
	}

	// 2. 检查是否是 bathroom
	if !c.checkBathroomForEvent3(*cardInfo) {
		return nil // 不是 bathroom，退出
	}

	// 3. 清除 Event 3 状态（离开 bathroom，停止监测）
	// 注意：这里需要清除所有可能的 track_id 状态
	// 由于不知道具体的 track_id，需要从实时数据中获取
	realtimeData, err := c.cache.GetRealtimeData(cardInfo.CardID)
	if err != nil {
		return nil // 实时数据不存在，忽略
	}

	// 清除所有 track_id 的状态
	// 通过调用 Evaluate 方法，Event 3 会检测到不在 bathroom 并清除状态
	// 如果实时数据中已经没有人在 bathroom，直接清除所有可能的状态
	for _, posture := range realtimeData.Postures {
		c.logger.Info("Left bathroom, clearing Event 3 state",
			zap.String("card_id", cardInfo.CardID),
			zap.String("track_id", posture.TrackingID),
		)
		// 调用 Evaluate，Event 3 会检测到不在 bathroom 并清除状态
		_, _ = evaluator.Evaluate(iotData.TenantID, *cardInfo, realtimeData)
	}

	return nil
}

// handlePERSON_COUNT_CHANGED_Event 处理 PERSON_COUNT_CHANGED 事件（检测多人进入）
func (c *EventConsumer) handlePERSON_COUNT_CHANGED_Event(
	ctx context.Context,
	iotData models.IoTDataMessage,
	evaluator Evaluator,
) error {
	// 1. 通过 device_id 查询 card_id
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		return nil // 设备可能未绑定到卡片，忽略
	}

	// 2. 检查是否是 bathroom
	if !c.checkBathroomForEvent3(*cardInfo) {
		return nil // 不是 bathroom，退出
	}

	// 3. 从 Redis 缓存读取实时数据
	realtimeData, err := c.cache.GetRealtimeData(cardInfo.CardID)
	if err != nil {
		return nil // 实时数据不存在，忽略
	}

	// 4. 如果人数 != 1，清除 Event 3 状态（多人进入，退出监测）
	if realtimeData.PersonCount != 1 || len(realtimeData.Postures) != 1 {
		c.logger.Info("Person count changed, clearing Event 3 state",
			zap.String("card_id", cardInfo.CardID),
			zap.Int("person_count", realtimeData.PersonCount),
			zap.Int("track_count", len(realtimeData.Postures)),
		)
		// 调用 Evaluate，Event 3 会检测到人数 != 1 并清除状态
		_, _ = evaluator.Evaluate(iotData.TenantID, *cardInfo, realtimeData)
	}

	return nil
}

// checkBathroomForEvent3 检查是否是 bathroom（用于 Event 3）
func (c *EventConsumer) checkBathroomForEvent3(card repository.CardInfo) bool {
	// 从卡片绑定的设备中获取 room_name
	devices, err := c.cardRepo.GetCardDevices(card.CardID)
	if err != nil {
		c.logger.Debug("Failed to get card devices",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
		return false
	}

	// 检查设备绑定的房间名称
	for _, device := range devices {
		if device.RoomName != nil {
			roomNameLower := strings.ToLower(*device.RoomName)
			if strings.Contains(roomNameLower, "bathroom") ||
				strings.Contains(roomNameLower, "restroom") ||
				strings.Contains(roomNameLower, "bath") ||
				strings.Contains(roomNameLower, "rest") ||
				strings.Contains(roomNameLower, "toilet") {
				return true
			}
		}
	}

	return false
}
