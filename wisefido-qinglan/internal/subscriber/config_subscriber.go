package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber 配置变更订阅器
type ConfigSubscriber struct {
	redisClient    *redis.Client
	config         *config.Config
	logger         *zap.Logger
	deviceRepo     repository.DeviceRepository
	cardMappingSvc *service.CardMappingService
	deviceCache    interface{} // 设备缓存（sync.Map），用于缓存设备认证信息
	consumerGroup  string
	consumerName   string
	mu             sync.RWMutex
}

// NewConfigSubscriber 创建配置变更订阅器
func NewConfigSubscriber(
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
	deviceRepo repository.DeviceRepository,
	deviceCache interface{}, // 设备缓存（sync.Map），用于缓存设备认证信息
	cardMappingSvc *service.CardMappingService,
) *ConfigSubscriber {
	return &ConfigSubscriber{
		redisClient:    redisClient,
		config:         cfg,
		logger:         logger,
		deviceRepo:     deviceRepo,
		deviceCache:    deviceCache,
		cardMappingSvc: cardMappingSvc,
		consumerGroup:  "qinglan-config-consumer",
		consumerName:   "qinglan-config-consumer-1",
	}
}

// Start 启动配置变更订阅器
func (s *ConfigSubscriber) Start(ctx context.Context) error {
	// 订阅配置变更流（wisefido-data 发送）：
	// 1. config:alarmDevice:stream - 告警设备配置（BuildAlarmDeviceMessage）
	// 2. config:card:stream - 卡片配置（BuildCardChangeMessage）
	streams := []string{
		rediscommon.StreamConfigAlarmDevice.Name, // config:alarmDevice:stream
		rediscommon.StreamConfigCard.Name,        // config:card:stream
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

	// 注意：消费协程现在由 main.go 的 subscribeConfigStream 函数启动
	// 这里仅做消费者组的初始化

	return nil
}

// HandleConfigChangeMessage 处理配置变更消息（支持 CloudEvents 格式和旧格式）
func (s *ConfigSubscriber) HandleConfigChangeMessage(ctx context.Context, stream string, message redis.XMessage) error {
	s.logger.Debug("Received config change message",
		zap.String("stream", stream),
		zap.String("message_id", message.ID),
	)

	// 解析消息数据（支持展开格式和包装格式）
	var configMsg rediscommon.ConfigChangeMessage

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
	switch configMsg.Type {
	case "config.alarmDevice":
		// 告警设备配置变更 (BuildAlarmDeviceMessage)
		s.handleAlarmDeviceChange(configMsg, message.ID)
	case "config.card":
		// 卡片配置变更 (BuildCardChangeMessage)
		s.handleCardChange(configMsg, message.ID)
	// 旧事件类型已注释（原: "config.device.alarm.setting.updated"）
	// case "config.device.alarm.setting.updated":
	//     s.handleDeviceAlarmSettingChange(configMsg, message.ID)
	default:
		// 其他类型的事件暂不处理，跳过
		s.logger.Debug("Skipping config change event (not subscribed)",
			zap.String("event_type", configMsg.Type),
		)
	}

	return nil
}

// Stop 停止配置变更订阅器
func (s *ConfigSubscriber) Stop() error {
	// TODO: 实现优雅关闭逻辑
	s.logger.Info("Config change subscriber stopped")
	return nil
}

// handleAlarmDeviceChange 处理告警设备配置变更 (BuildAlarmDeviceMessage)
// 清除对应设备的报警使能配置缓存，使下次查询重新加载最新配置
func (s *ConfigSubscriber) handleAlarmDeviceChange(configMsg rediscommon.ConfigChangeMessage, messageID string) {
	tenantID, _ := configMsg.Data["tenant_id"].(string)
	deviceUID, _ := configMsg.Data["device_uid"].(string)
	settingType, _ := configMsg.Data["setting_type"].(string)

	if tenantID == "" || deviceUID == "" {
		s.logger.Warn("Alarm device change message missing tenant_id or device_uid",
			zap.String("message_id", messageID),
			zap.String("tenant_id", tenantID),
			zap.String("device_uid", deviceUID))
		return
	}

	// 检查消息时间戳，防止过时消息覆盖新配置
	// 允许最大时间差 5 秒（包括网络延迟和时钟偏差）
	var messageTimestampMs int64
	if ts, ok := configMsg.Data["timestamp_ms"].(float64); ok {
		messageTimestampMs = int64(ts)
	} else if ts, ok := configMsg.Data["timestamp_ms"].(int64); ok {
		messageTimestampMs = ts
	}

	if messageTimestampMs > 0 {
		currentTimeMs := time.Now().UnixMilli()
		timeDiffMs := currentTimeMs - messageTimestampMs
		const maxAllowedDelayMs = 5000 // 5 秒

		if timeDiffMs > maxAllowedDelayMs {
			s.logger.Warn("Discarding stale alarm device config message (too old)",
				zap.String("message_id", messageID),
				zap.String("tenant_id", tenantID),
				zap.String("device_uid", deviceUID),
				zap.Int64("time_diff_ms", timeDiffMs),
				zap.Int64("max_allowed_ms", maxAllowedDelayMs))
			return
		}

		if timeDiffMs < -5000 { // 消息时间戳在未来
			s.logger.Warn("Discarding alarm device config message with future timestamp",
				zap.String("message_id", messageID),
				zap.String("tenant_id", tenantID),
				zap.String("device_uid", deviceUID),
				zap.Int64("time_diff_ms", timeDiffMs))
			return
		}
	}

	s.logger.Info("Handling alarm device config change",
		zap.String("message_id", messageID),
		zap.String("tenant_id", tenantID),
		zap.String("device_uid", deviceUID),
		zap.String("setting_type", settingType))

	// 清除对应设备的报警使能配置缓存
	// 这样下次 mqtt_consumer 查询该设备的使能配置时，会从 PG 重新加载最新数据
	s.deviceRepo.ClearAlarmEnablementCache(tenantID, deviceUID)

	s.logger.Info("Alarm device cache cleared",
		zap.String("message_id", messageID),
		zap.String("tenant_id", tenantID),
		zap.String("device_uid", deviceUID))
}

// handleCardChange 处理卡片配置变更 (BuildCardChangeMessage)
// 当 Type 为 config.card 时：正常处理卡片变更（调用 CardMappingService）
// 当 Type 为 config.card.device_store 时：这是 device_store 变化信号（委托给 handleDeviceStoreChange）
func (s *ConfigSubscriber) handleCardChange(configMsg rediscommon.ConfigChangeMessage, messageID string) {
	ctx := context.Background()

	// 检查消息时间戳，防止过时消息覆盖新配置
	// 允许最大时间差 5 秒（包括网络延迟和时钟偏差）
	var messageTimestampMs int64
	if ts, ok := configMsg.Data["timestamp_ms"].(float64); ok {
		messageTimestampMs = int64(ts)
	} else if ts, ok := configMsg.Data["timestamp_ms"].(int64); ok {
		messageTimestampMs = ts
	}

	if messageTimestampMs > 0 {
		currentTimeMs := time.Now().UnixMilli()
		timeDiffMs := currentTimeMs - messageTimestampMs
		const maxAllowedDelayMs = 5000 // 5 秒

		if timeDiffMs > maxAllowedDelayMs {
			s.logger.Warn("Discarding stale card config message (too old)",
				zap.String("message_id", messageID),
				zap.String("type", configMsg.Type),
				zap.Int64("time_diff_ms", timeDiffMs),
				zap.Int64("max_allowed_ms", maxAllowedDelayMs))
			return
		}

		if timeDiffMs < -5000 { // 消息时间戳在未来
			s.logger.Warn("Discarding card config message with future timestamp",
				zap.String("message_id", messageID),
				zap.String("type", configMsg.Type),
				zap.Int64("time_diff_ms", timeDiffMs))
			return
		}
	}

	s.logger.Info("Handling card config change",
		zap.String("message_id", messageID),
		zap.String("type", configMsg.Type),
	)

	// 消息数据已经是 map[string]interface{}
	cardData := configMsg.Data

	// 提取关键字段
	tenantID, _ := cardData["tenant_id"].(string)

	if tenantID == "" {
		s.logger.Warn("Card change message missing tenant_id",
			zap.String("message_id", messageID))
		return
	}

	// 根据消息类型区分处理
	switch configMsg.Type {
	case rediscommon.ConfigCardDeviceStoreChanged:
		// device_store 变化信号
		s.handleDeviceStoreChange(ctx, cardData, messageID)

	case rediscommon.ConfigCardChanged:
		// 正常的卡片变更
		s.handleNormalCardChange(ctx, cardData, messageID)

	default:
		s.logger.Warn("Unknown card change message type",
			zap.String("type", configMsg.Type),
			zap.String("message_id", messageID))
	}
}

// handleNormalCardChange 处理正常的卡片配置变更 (ConfigCardChanged)
// 调用 CardMappingService 按 branch_id 重新计算卡片-设备映射
func (s *ConfigSubscriber) handleNormalCardChange(ctx context.Context, cardData map[string]interface{}, messageID string) {
	branchID, _ := cardData["branch_id"].(string)
	cardID, _ := cardData["card_id"].(string)
	tenantID, _ := cardData["tenant_id"].(string)

	if branchID == "" {
		s.logger.Warn("Card change message missing branch_id",
			zap.String("message_id", messageID),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID))
		return
	}

	s.logger.Info("Card change details",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("branch_id", branchID))

	// 调用 CardMappingService 处理卡片变更
	// cardMappingSvc 已在初始化时注入
	if err := s.cardMappingSvc.HandleCardChanged(ctx, cardData); err != nil {
		s.logger.Error("Failed to handle card change",
			zap.String("message_id", messageID),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.Error(err))
		return
	}

	s.logger.Info("Card change processed successfully",
		zap.String("message_id", messageID),
		zap.String("tenant_id", tenantID),
		zap.String("branch_id", branchID))
}

// handleDeviceStoreChange 处理 device_store 变化信号（Type 为 config.card.device_store）
// 当 device_store 发生变化时，查询最新的设备信息，刷新缓存
// 无需区分具体变化类型，统一处理方式：查库 device_store → 根据 device_uid 和 allow_access 更新缓存
func (s *ConfigSubscriber) handleDeviceStoreChange(ctx context.Context, data map[string]interface{}, messageID string) {
	deviceID, _ := data["device_id"].(string)

	if deviceID == "" {
		s.logger.Warn("Device store change message missing device_id",
			zap.String("message_id", messageID))
		return
	}

	s.logger.Info("Handling device store change",
		zap.String("message_id", messageID),
		zap.String("device_id", deviceID))

	// 查询 device_store 表获取最新的 device_uid 和 allow_access
	deviceStoreInfo, err := s.deviceRepo.GetDeviceStoreByDeviceID(ctx, deviceID)
	if err != nil {
		s.logger.Warn("Failed to get device store info",
			zap.String("device_id", deviceID),
			zap.Error(err))
		return
	}

	s.logger.Info("Retrieved device store info, refreshing cache",
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceStoreInfo.DeviceUID),
		zap.Bool("allow_access", deviceStoreInfo.AllowAccess))

	// 根据最新的 allow_access 更新缓存
	// DeviceCache: key=device_uid, value=*DeviceWithLocation
	// AllowAccessCache: key=device_uid, value=bool
	if s.deviceCache != nil {
		if deviceStoreCache, ok := s.deviceCache.(*sync.Map); ok {
			if deviceStoreInfo.AllowAccess {
				// 允许访问：保留缓存中的设备信息
				s.logger.Debug("Device allowed, keeping cache",
					zap.String("device_uid", deviceStoreInfo.DeviceUID),
					zap.String("device_id", deviceID))
			} else {
				// 禁止访问：从缓存中删除设备
				deviceStoreCache.Delete(deviceStoreInfo.DeviceUID)
				s.logger.Info("Device denied access, removed from cache",
					zap.String("device_uid", deviceStoreInfo.DeviceUID),
					zap.String("device_id", deviceID))
			}
		}
	}

	s.logger.Info("Device store change processed",
		zap.String("message_id", messageID),
		zap.String("device_id", deviceID))
}

// 旧方法已注释（原: handleDeviceAlarmSettingChange）
// func (s *ConfigSubscriber) handleDeviceAlarmSettingChange(configMsg rediscommon.ConfigChangeMessage, messageID string) {
//     // 旧业务逻辑已移除
// }

// createConsumerGroupIfNotExists 创建消费者组（如果不存在）
func (s *ConfigSubscriber) createConsumerGroupIfNotExists(ctx context.Context, stream string) error {
	// TODO: 实现消费者组创建逻辑
	return nil
}

// consumeConfigChanges 消费配置变更消息
// TODO: 实现消息消费循环
func (s *ConfigSubscriber) consumeConfigChanges(ctx context.Context, streams []string) {
	// TODO: 实现消费逻辑
}

// parseConfigMessageFromValues 从 Redis 消息值中解析配置消息
func parseConfigMessageFromValues(values map[string]interface{}) rediscommon.ConfigChangeMessage {
	// TODO: 实现解析逻辑
	return rediscommon.ConfigChangeMessage{}
}
