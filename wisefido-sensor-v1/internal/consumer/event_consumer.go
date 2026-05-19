package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"owl-common/alarm"
	rediscommon "owl-common/redis"
	"strings"
	"time"
	"wisefido-sensor-v1/internal/config"
	"wisefido-sensor-v1/internal/models"
	"wisefido-sensor-v1/internal/repository"

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
	// 订阅统一 streams（event 和 alarm）
	streams := []string{
		c.config.Alarm.IoTStream.Event,
		c.config.Alarm.IoTStream.Alarm,
	}

	consumerGroup := c.config.Alarm.IoTStream.ConsumerGroup
	consumerName := c.config.Alarm.IoTStream.ConsumerName

	// 创建消费者组
	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, consumerGroup); err != nil {
			c.logger.Warn("Failed to create consumer group for stream, will retry",
				zap.String("stream", stream),
				zap.Error(err),
			)
			// 继续处理其他 streams，不中断
		}
	}

	c.logger.Info("Event consumer started",
		zap.Strings("streams", streams),
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
			// 消费 event 和 alarm streams
			eventErr := c.consumeStream(ctx, c.config.Alarm.IoTStream.Event, consumerGroup, consumerName, evaluator)
			alarmErr := c.consumeStream(ctx, c.config.Alarm.IoTStream.Alarm, consumerGroup, consumerName, evaluator)

			if eventErr != nil && alarmErr != nil {
				c.logger.Error("Failed to consume all streams", zap.Error(eventErr))
				time.Sleep(time.Second)
			} else {
				if eventErr != nil {
					c.logger.Error("Failed to consume event stream", zap.Error(eventErr))
				}
				if alarmErr != nil {
					c.logger.Error("Failed to consume alarm stream", zap.Error(alarmErr))
				}
			}
		}
	}
}

// consumeStream 消费单个 Stream
func (c *EventConsumer) consumeStream(ctx context.Context, stream, consumerGroup, consumerName string, evaluator Evaluator) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		stream,
		consumerGroup,
		consumerName,
		c.config.Alarm.IoTStream.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("failed to read from stream %s: %w", stream, err)
	}

	// 处理消息
	for _, msg := range messages {
		if err := c.processMessage(ctx, msg, evaluator); err != nil {
			c.logger.Error("Failed to process message",
				zap.String("stream", stream),
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}

	return nil
}

// processMessage 处理单条消息
func (c *EventConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage, evaluator Evaluator) error {
	// 解析消息数据（直接从设备 streams 读取）
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

	// 解析 JSON - 设备 streams 的格式包含 device_id, tenant_id, device_type, topic_type, timestamp, data_value 等
	var streamData map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &streamData); err != nil {
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}

	// 提取必要字段
	deviceID, _ := streamData["device_id"].(string)
	tenantID, _ := streamData["tenant_id"].(string)
	topicType, _ := streamData["topic_type"].(string)
	
	if deviceID == "" || tenantID == "" {
		return nil // 跳过无效消息
	}

	// 从 dataValue 中提取事件类型（键名规范：dataValue）
	var eventType string
	if topicType == "event" || topicType == "alarm" {
		payload := streamData[rediscommon.DataValueKey]
		if dataValue, ok := payload.(map[string]interface{}); ok {
			if category, ok := dataValue["dataCategory"].(string); ok && category != "" {
				eventType = category
			} else if category, ok := dataValue["category"].(string); ok {
				eventType = category
			} else if evt, ok := dataValue["event"].(string); ok {
				eventType = evt
			}
		} else if dataValueArray, ok := payload.([]interface{}); ok && len(dataValueArray) > 0 {
			if firstItem, ok := dataValueArray[0].(map[string]interface{}); ok {
				if category, ok := firstItem["dataCategory"].(string); ok && category != "" {
					eventType = category
				} else if category, ok := firstItem["category"].(string); ok {
					eventType = category
				} else if evt, ok := firstItem["event"].(string); ok {
					eventType = evt
				}
			}
		}
	}

	// 过滤：只处理相关事件
	if eventType == "" {
		return nil // 跳过非事件消息
	}

	// 映射事件类型到 wisefido-sensor 期望的格式
	// 例如：从 "Enter room" 映射到 "ENTER_ROOM"
	mappedEventType := c.mapEventType(eventType)
	if mappedEventType == "" {
		return nil // 跳过不支持的事件类型
	}

	c.logger.Debug("Received event",
		zap.String("event_type", mappedEventType),
		zap.String("original_event", eventType),
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.String("stream_id", msg.ID),
	)

	// 构建 IoTDataMessage 格式（用于兼容现有处理函数）
	iotData := models.IoTDataMessage{
		DeviceID:   deviceID,
		TenantID:   tenantID,
		DeviceType: streamData["device_type"].(string),
		Timestamp:  getTimestamp(streamData["timestamp"]),
		EventType:  &mappedEventType,
	}

	// 根据事件类型分发处理
	switch mappedEventType {
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

// mapEventType 映射事件类型到 wisefido-sensor 期望的格式
func (c *EventConsumer) mapEventType(eventType string) string {
	// 事件类型映射表
	eventTypeMap := map[string]string{
		// 离床事件
		"Left bed":   "BED_LEFT",
		alarm.LeftBed: "BED_LEFT",
		"left_bed":   "BED_LEFT",
		// 进入房间事件
		"Enter room":   "ENTER_ROOM",
		alarm.EnterRoom: "ENTER_ROOM",
		"enter_room":   "ENTER_ROOM",
		// 离开房间事件
		"Leave room": "LEFT_ROOM",
		"LeaveRoom":  "LEFT_ROOM",
		"leave_room": "LEFT_ROOM",
		// 人数变化事件
		"number-people":     "PERSON_COUNT_CHANGED",
		"NumberPeople":      "PERSON_COUNT_CHANGED",
		alarm.NumberPeople:  "PERSON_COUNT_CHANGED",
	}

	if mapped, ok := eventTypeMap[eventType]; ok {
		return mapped
	}
	
	// 尝试不区分大小写匹配
	for k, v := range eventTypeMap {
		if strings.EqualFold(k, eventType) {
			return v
		}
	}
	
	return ""
}

// getTimestamp 从 interface{} 提取时间戳
func getTimestamp(ts interface{}) int64 {
	switch v := ts.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	default:
		return time.Now().Unix()
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
	if card.CardType != "ActiveBedCard" {
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
