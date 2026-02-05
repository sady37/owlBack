package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	commonredis "owl-common/redis"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber 配置变更订阅器
// 订阅以下 config:* 流：
// 1. config:device_status:stream - 设备在线状态变更（config.device.online/offline/unsubscribed）
// 2. config:card:stream - 卡片变更事件（config.card.created/updated/deleted）
type ConfigSubscriber struct {
	redisClient    *redis.Client
	config         *config.Config
	logger         *zap.Logger
	deviceRepo     repository.DeviceRepository
	deviceCache    *sync.Map          // 设备缓存（key: device_uid, value: *domain.DeviceWithLocation）
	cardMappingSvc CardMappingService // 卡片映射服务（可选，处理卡片变更）
	consumerGroup  string             // Consumer Group 名称
	consumerName   string             // Consumer 名称（用于标识当前实例）
	stopChan       chan struct{}
	// 去重机制：只处理每个 device_uid 的最新消息
	lastProcessed sync.Map // key: "device_uid", value: message.ID (string)
}

// CardMappingService 卡片映射服务接口
type CardMappingService interface {
	HandleCardCreated(ctx context.Context, data map[string]interface{}) error
	HandleCardUpdated(ctx context.Context, data map[string]interface{}) error
	HandleCardDeleted(ctx context.Context, data map[string]interface{}) error
}

// NewConfigSubscriber 创建配置变更订阅器
func NewConfigSubscriber(
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
	deviceRepo repository.DeviceRepository,
	deviceCache *sync.Map,
	cardMappingSvc CardMappingService, // 卡片映射服务（可选，处理卡片变更）
) *ConfigSubscriber {
	return &ConfigSubscriber{
		redisClient:    redisClient,
		config:         cfg,
		logger:         logger,
		deviceRepo:     deviceRepo,
		deviceCache:    deviceCache,
		cardMappingSvc: cardMappingSvc,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动配置变更订阅器
func (s *ConfigSubscriber) Start(ctx context.Context) error {
	// 订阅 config:device.alarm.setting:stream
	// 该流由 wisefido-data 的 PublishDeviceAlarmSettingMessage 发出
	// 卡片消息 (config.card.*) 当前 wisefido-data 还未发布，暂不订阅
	streams := []string{
		commonredis.StreamConfigDeviceAlarmSetting.Name, // config:device.alarm.setting:stream
	}

	// 创建 Consumer Group（如果不存在）
	for _, stream := range streams {
		if err := s.createConsumerGroupIfNotExists(ctx, stream); err != nil {
			s.logger.Warn("Failed to create consumer group, will use XRead instead",
				zap.String("stream", stream),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Starting config change subscriber",
		zap.Strings("streams", streams),
		zap.String("consumer_group", s.consumerGroup),
		zap.String("consumer_name", s.consumerName),
	)

	// 启动消费协程
	go s.consumeConfigChanges(ctx, streams)

	return nil
}

// createConsumerGroupIfNotExists 创建 Consumer Group（如果不存在）
func (s *ConfigSubscriber) createConsumerGroupIfNotExists(ctx context.Context, stream string) error {
	// 尝试创建 Consumer Group，从 stream 开始位置（"0"）开始消费
	err := s.redisClient.XGroupCreate(ctx, stream, s.consumerGroup, "0").Err()
	if err != nil {
		// 如果错误是 "BUSYGROUP"，说明 Consumer Group 已存在，这是正常的
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil
		}
		return err
	}
	return nil
}

// Stop 停止配置变更订阅器
func (s *ConfigSubscriber) Stop() {
	close(s.stopChan)
	s.logger.Info("Config change subscriber stopped")
}

// consumeConfigChanges 消费配置变更事件
func (s *ConfigSubscriber) consumeConfigChanges(
	ctx context.Context,
	streams []string,
) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Config change subscriber context done")
			return
		case <-s.stopChan:
			s.logger.Info("Config change subscriber stopped by stop channel")
			return
		default:
			// 使用 Consumer Group 读取消息
			streamsWithIDs := make([]string, 0, len(streams)*2)
			for _, stream := range streams {
				streamsWithIDs = append(streamsWithIDs, stream, ">") // ">" 表示只读取未确认的消息
			}

			result, err := s.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    s.consumerGroup,
				Consumer: s.consumerName,
				Streams:  streamsWithIDs,
				Count:    10,
				Block:    5 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// 超时，继续等待
					continue
				}
				s.logger.Error("Failed to read config change streams",
					zap.Error(err),
				)
				time.Sleep(1 * time.Second)
				continue
			}

			// 处理读取到的消息
			for _, stream := range result {
				for _, message := range stream.Messages {
					// 处理消息
					if err := s.handleConfigChangeMessage(ctx, stream.Stream, message); err != nil {
						s.logger.Error("Failed to handle config change message",
							zap.String("stream", stream.Stream),
							zap.String("message_id", message.ID),
							zap.Error(err),
						)
						// 处理失败时不确认消息，让 Consumer Group 重新分配
						continue
					}

					// 处理成功后确认消息
					if err := s.redisClient.XAck(ctx, stream.Stream, s.consumerGroup, message.ID).Err(); err != nil {
						s.logger.Error("Failed to acknowledge message",
							zap.String("stream", stream.Stream),
							zap.String("message_id", message.ID),
							zap.Error(err),
						)
					}
				}
			}
		}
	}
}

// handleConfigChangeMessage 处理配置变更消息（支持 CloudEvents 格式和旧格式）
func (s *ConfigSubscriber) handleConfigChangeMessage(ctx context.Context, stream string, message redis.XMessage) error {
	s.logger.Debug("Received config change message",
		zap.String("stream", stream),
		zap.String("message_id", message.ID),
	)

	// 解析消息数据（支持展开格式和包装格式）
	var configMsg commonredis.ConfigChangeMessage

	// 优先尝试包装格式（由 PublishJSONToStream 发送）
	if len(message.Values) > 0 {
		// 检查是否有 "data" 字段（包装格式，由 PublishJSONToStream 发送）
		if dataStr, ok := message.Values["data"].(string); ok {
			// 包装格式：从 "data" 字段解析 JSON（JSON 字符串）
			if err := json.Unmarshal([]byte(dataStr), &configMsg); err != nil {
				return fmt.Errorf("failed to parse config change message from 'data' field: %w", err)
			}
		} else {
			// 展开格式：直接使用 message.Values（字段直接展开，需要转换类型）
			configMsg = parseConfigMessageFromValues(message.Values)
		}
	} else {
		return fmt.Errorf("invalid message format: empty message values")
	}

	// 根据事件类型处理配置变更
	// 当前只订阅 config.device.alarm.setting.updated
	switch configMsg.Type {
	case "config.device.alarm.setting.updated":
		// 设备告警配置更新
		s.handleDeviceAlarmSettingChange(ctx, configMsg, message.ID)
	default:
		// 其他类型的事件暂不处理，跳过
		s.logger.Debug("Skipping config change event (not subscribed)",
			zap.String("event_type", configMsg.Type),
		)
	}

	return nil
}

// parseConfigMessageFromValues 从展开的 message.Values 解析 ConfigChangeMessage（CloudEvents 标准格式）
func parseConfigMessageFromValues(values map[string]interface{}) commonredis.ConfigChangeMessage {
	msg := commonredis.ConfigChangeMessage{}

	// CloudEvents 标准字段
	if specversion, ok := values["specversion"].(string); ok {
		msg.SpecVersion = specversion
	}
	if id, ok := values["id"].(string); ok {
		msg.ID = id
	}
	if source, ok := values["source"].(string); ok {
		msg.Source = source
	}
	if eventType, ok := values["type"].(string); ok {
		msg.Type = eventType
	}
	if timeStr, ok := values["time"].(string); ok {
		msg.Time = timeStr
	}

	// 解析 data 字段（可能是 JSON 字符串或 map）
	if dataStr, ok := values["data"].(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			msg.Data = dataMap
		}
	} else if dataMap, ok := values["data"].(map[string]interface{}); ok {
		msg.Data = dataMap
	}

	return msg
}

// handleCardChange 处理卡片变更事件（created/updated/deleted）
func (s *ConfigSubscriber) handleCardChange(
	ctx context.Context,
	operation string,
	msg commonredis.ConfigChangeMessage,
	messageID string,
) {
	if s.cardMappingSvc == nil {
		s.logger.Debug("Card mapping service not configured, skipping card change event",
			zap.String("operation", operation),
			zap.String("event_type", msg.Type))
		return
	}

	// 从 data 字段中提取卡片信息
	if msg.Data == nil {
		s.logger.Warn("Invalid card change event, missing data field",
			zap.String("operation", operation),
			zap.String("event_type", msg.Type))
		return
	}

	var tenantID, cardID, unitID, branchID string
	var timestampMs float64

	if tid, ok := msg.Data["tenant_id"].(string); ok {
		tenantID = tid
	}
	if cid, ok := msg.Data["card_id"].(string); ok {
		cardID = cid
	}
	if uid, ok := msg.Data["unit_id"].(string); ok {
		unitID = uid
	}
	if bid, ok := msg.Data["branch_id"].(string); ok {
		branchID = bid
	}
	if ts, ok := msg.Data["timestamp_ms"].(float64); ok {
		timestampMs = ts
	}

	if tenantID == "" || cardID == "" {
		s.logger.Warn("Invalid card change event, missing required fields",
			zap.String("operation", operation),
			zap.String("event_type", msg.Type),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID))
		return
	}

	// 去重：只处理同一个卡片的最新消息
	key := "card:" + operation + ":" + cardID
	if lastID, exists := s.lastProcessed.Load(key); exists {
		if lastID.(string) >= messageID {
			// 已处理过更新的消息，跳过
			s.logger.Debug("Skipping duplicate card change event",
				zap.String("operation", operation),
				zap.String("card_id", cardID),
				zap.String("last_message_id", lastID.(string)),
				zap.String("current_message_id", messageID))
			return
		}
	}
	s.lastProcessed.Store(key, messageID)

	s.logger.Info("Received card change event",
		zap.String("operation", operation),
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("unit_id", unitID),
		zap.String("branch_id", branchID),
		zap.Float64("timestamp_ms", timestampMs),
		zap.String("message_id", messageID),
		zap.String("event_id", msg.ID),
		zap.String("event_type", msg.Type))

	// 调用卡片映射服务处理
	var err error
	switch operation {
	case "created":
		err = s.cardMappingSvc.HandleCardCreated(ctx, msg.Data)
	case "updated":
		err = s.cardMappingSvc.HandleCardUpdated(ctx, msg.Data)
	case "deleted":
		err = s.cardMappingSvc.HandleCardDeleted(ctx, msg.Data)
	}

	if err != nil {
		s.logger.Error("Failed to handle card change event",
			zap.String("operation", operation),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.Error(err))
	} else {
		s.logger.Info("Successfully handled card change event",
			zap.String("operation", operation),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID))
	}
}

// clearDeviceCache 清除指定设备的缓存
func (s *ConfigSubscriber) clearDeviceCache(deviceUID string) {
	if s.deviceCache != nil && deviceUID != "" {
		s.deviceCache.Delete(deviceUID)
		s.logger.Debug("Cleared device cache",
			zap.String("device_uid", deviceUID),
		)
	}
}

// handleAlarmDeviceChange 处理 alarm_device 配置变更
func (s *ConfigSubscriber) handleAlarmDeviceChange(msg commonredis.ConfigChangeMessage, messageID string) {
	// 从 data 字段中提取设备信息（CloudEvents 格式，由 BuildAlarmDeviceMessage 生成）
	var deviceUID, deviceID, deviceCode, deviceType, tenantID string

	if msg.Data != nil {
		if uid, ok := msg.Data["device_uid"].(string); ok {
			deviceUID = uid
		}
		if id, ok := msg.Data["device_id"].(string); ok {
			deviceID = id
		}
		if code, ok := msg.Data["device_code"].(string); ok {
			deviceCode = code
		}
		if dtype, ok := msg.Data["device_type"].(string); ok {
			deviceType = dtype
		}
		if tid, ok := msg.Data["tenant_id"].(string); ok {
			tenantID = tid
		}
	}

	if deviceUID == "" {
		s.logger.Warn("Invalid alarm_device change event, missing device_uid",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)
		return
	}

	// 去重：只处理同一个 device_uid 的最新消息
	key := "alarm_device:" + deviceUID
	if lastID, exists := s.lastProcessed.Load(key); exists {
		if lastID.(string) >= messageID {
			// 已处理过更新的消息，跳过
			s.logger.Debug("Skipping duplicate alarm_device change event",
				zap.String("device_uid", deviceUID),
				zap.String("last_message_id", lastID.(string)),
				zap.String("current_message_id", messageID),
			)
			return
		}
	}
	s.lastProcessed.Store(key, messageID)

	s.logger.Info("Received alarm_device config change event",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("device_code", deviceCode),
		zap.String("device_type", deviceType),
		zap.String("message_id", messageID),
		zap.String("event_id", msg.ID),
		zap.String("event_type", msg.Type),
	)

	// alarm_device 变更：清除指定设备的缓存
	s.clearDeviceCache(deviceUID)

	// 清除报警使能配置缓存
	s.deviceRepo.ClearAlarmEnablementCache(tenantID, deviceUID)

	// 预加载新的报警使能配置到缓存（异步，避免阻塞）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.deviceRepo.PreloadAlarmEnablement(ctx, tenantID, deviceUID); err != nil {
			s.logger.Warn("Failed to preload alarm enablement after config change",
				zap.String("tenant_id", tenantID),
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
		} else {
			s.logger.Info("Preloaded alarm enablement cache after config change",
				zap.String("tenant_id", tenantID),
				zap.String("device_uid", deviceUID),
			)
		}
	}()

	s.logger.Info("Cleared device cache and alarm enablement cache for alarm_device change",
		zap.String("tenant_id", tenantID),
		zap.String("device_uid", deviceUID),
		zap.String("message_id", messageID),
	)

	// 确认消息
	if err := s.redisClient.XAck(context.Background(), commonredis.StreamConfigDeviceStatus.Name, s.consumerGroup, messageID).Err(); err != nil {
		s.logger.Warn("Failed to acknowledge alarm_device_updated message",
			zap.String("message_id", messageID),
			zap.Error(err),
		)
	}
}

// handleDeviceAlarmSettingChange 处理设备告警配置变更
// 当设备告警配置（如启用/禁用告警等）发生变更时，将变更通知发送给设备
// 消息格式由 BuildDeviceAlarmSettingMessage 生成，包含以下字段：
// - tenant_id, device_id, device_uid, setting_type, timestamp_ms
// - 其他配置数据会合并到 data 中
func (s *ConfigSubscriber) handleDeviceAlarmSettingChange(ctx context.Context, configMsg commonredis.ConfigChangeMessage, messageID string) {
	dataMap := configMsg.Data
	if dataMap == nil {
		s.logger.Warn("Invalid data format in device alarm setting change: data is nil",
			zap.String("message_id", messageID),
		)
		return
	}

	// 解析标准字段
	tenantID, _ := dataMap["tenant_id"].(string)
	deviceID, _ := dataMap["device_id"].(string)
	deviceUID, _ := dataMap["device_uid"].(string)
	settingType, _ := dataMap["setting_type"].(string)
	timestampMs, _ := dataMap["timestamp_ms"].(float64)

	if tenantID == "" || deviceUID == "" {
		s.logger.Warn("Missing required fields in device alarm setting change",
			zap.String("tenant_id", tenantID),
			zap.String("device_uid", deviceUID),
		)
		return
	}

	s.logger.Info("Device alarm setting changed",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("setting_type", settingType),
		zap.Float64("timestamp_ms", timestampMs),
	)

	// TODO: 根据 setting_type 和 dataMap 中的数据决定是否需要发送 MQTT 消息到设备
	// 例如，可能需要发送告警使能/禁用指令到设备
	// 其他配置数据（如 monitor_config）已合并到 dataMap 中，可按需提取

	// 确认消息
	if err := s.redisClient.XAck(context.Background(), commonredis.StreamConfigDeviceAlarmSetting.Name, s.consumerGroup, messageID).Err(); err != nil {
		s.logger.Warn("Failed to acknowledge device alarm setting message",
			zap.String("message_id", messageID),
			zap.Error(err),
		)
	}
}
