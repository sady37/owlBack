package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/repository"
	"wisefido-radar/pkg/mqtt"

	"owl-common/encode"
	mqttcommon "owl-common/mqtt"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// MQTTConsumer MQTT消息消费者
// 根据协议文档 3.3.3 节，订阅设备发布的 6 个主题
type MQTTConsumer struct {
	config      *config.Config
	mqttClient  *mqttcommon.Client
	redisClient *redis.Client
	deviceRepo  *repository.DeviceRepository
	logger      *zap.Logger
	subscriptionManager interface { // 避免循环依赖，使用接口
		AutoSubscribe(ctx context.Context, uid string) error
	}
}

// NewMQTTConsumer 创建MQTT消费者
func NewMQTTConsumer(
	cfg *config.Config,
	mqttClient *mqttcommon.Client,
	redisClient *redis.Client,
	deviceRepo *repository.DeviceRepository,
	logger *zap.Logger,
) *MQTTConsumer {
	return &MQTTConsumer{
		config:      cfg,
		mqttClient:  mqttClient,
		redisClient: redisClient,
		deviceRepo:  deviceRepo,
		logger:      logger,
	}
}

// SetSubscriptionManager 设置订阅管理器（避免循环依赖）
func (c *MQTTConsumer) SetSubscriptionManager(manager interface {
	AutoSubscribe(ctx context.Context, uid string) error
}) {
	c.subscriptionManager = manager
}

// Start 启动消费者
// 根据协议文档 3.3.3 节，订阅设备发布的 6 个主题：
// 1. /prefix/prop/productId/UID/post - 属性响应
// 2. /prefix/monitor/productId/UID/post - 实时数据
// 3. /prefix/func/productId/UID/post - 功能响应
// 4. /prefix/stat/productId/UID/post - 统计数据
// 5. /prefix/event/productId/UID/post - 事件/日志
// 6. /prefix/alarm/productId/UID/post - 告警
func (c *MQTTConsumer) Start(ctx context.Context) error {
	cfg := c.config.Radar.DeviceMQTT

	// 构建订阅主题（使用通配符 + 匹配所有设备）
	// 注意：如果 prefix 为空，格式为 /{type}/{productId}/+/post
	// 如果 prefix 不为空，格式为 /{prefix}/{type}/{productId}/+/post
	topics := []string{
		mqtt.BuildPropertyPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildMonitorPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildFunctionPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildStatPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildEventPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildAlarmPostTopic(cfg.Prefix, cfg.ProductID, "+"),
	}

	// 订阅所有主题
	for _, topic := range topics {
		if err := c.mqttClient.Subscribe(topic, 1, c.handleMessage); err != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
		c.logger.Info("Subscribed to MQTT topic",
			zap.String("topic", topic),
		)
	}

	// 同时保留对旧格式主题的兼容（radar/+/data）
	if c.config.Radar.Topics.Data != "" {
		if err := c.mqttClient.Subscribe(c.config.Radar.Topics.Data, 1, c.handleLegacyMessage); err != nil {
			c.logger.Warn("Failed to subscribe to legacy data topic",
				zap.String("topic", c.config.Radar.Topics.Data),
				zap.Error(err),
			)
		} else {
			c.logger.Info("Subscribed to legacy MQTT topic",
				zap.String("topic", c.config.Radar.Topics.Data),
			)
		}
	}

	c.logger.Info("MQTT consumer started",
		zap.Int("topics_count", len(topics)),
	)

	// 等待上下文取消
	<-ctx.Done()
	return nil
}

// Stop 停止消费者
func (c *MQTTConsumer) Stop(ctx context.Context) error {
	cfg := c.config.Radar.DeviceMQTT

	// 取消订阅所有主题
	topics := []string{
		mqtt.BuildPropertyPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildMonitorPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildFunctionPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildStatPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildEventPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildAlarmPostTopic(cfg.Prefix, cfg.ProductID, "+"),
	}

	if c.config.Radar.Topics.Data != "" {
		topics = append(topics, c.config.Radar.Topics.Data)
	}

	if err := c.mqttClient.Unsubscribe(topics...); err != nil {
		c.logger.Error("Failed to unsubscribe", zap.Error(err))
	}

	c.logger.Info("MQTT consumer stopped")
	return nil
}

// handleMessage 处理MQTT消息（协议文档格式）
// 根据协议文档 3.3.3 节，处理设备发布的 6 种主题
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	c.logger.Debug("Received MQTT message",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)),
	)

	// 1. 解析主题，提取 prefix, type, productId, UID
	topicInfo, err := mqtt.ParseTopic(topic)
	if err != nil {
		c.logger.Error("Failed to parse topic",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to parse topic: %w", err)
	}

	// 2. 从主题中提取设备 UID
	uid := topicInfo.UID
	if uid == "+" {
		// 通配符主题，无法处理
		return fmt.Errorf("wildcard topic cannot be processed: %s", topic)
	}

	// 3. 解析消息
	var mqttData map[string]interface{}
	if err := json.Unmarshal(payload, &mqttData); err != nil {
		c.logger.Error("Failed to unmarshal MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 4. 查询设备信息（如果不存在，尝试从 device_store 自动创建）
	device, err := c.deviceRepo.GetDeviceByUID(uid)
	isFirstConnection := false
	if err != nil {
		// 设备不存在，尝试从 device_store 自动创建
		device, err = c.deviceRepo.GetOrCreateDeviceFromStore(context.Background(), uid, topic)
		if err != nil {
			c.logger.Warn("Device not found and cannot be created from device_store",
				zap.String("uid", uid),
				zap.String("mqtt_topic", topic),
				zap.Error(err),
			)
			return fmt.Errorf("device not found: %s", uid)
		}
		// 设备已从 device_store 自动创建
		c.logger.Info("Device auto-created from device_store on MQTT connection",
			zap.String("device_id", device.DeviceID),
			zap.String("uid", uid),
			zap.String("mqtt_topic", topic),
		)
		isFirstConnection = true
	} else {
		// 设备已存在，检查是否是首次收到消息（通过检查订阅状态）
		// 如果设备存在但没有订阅记录，也视为首次连接
		subscriptionKey := fmt.Sprintf("radar:subscription:%s", uid)
		exists, err := c.redisClient.Exists(context.Background(), subscriptionKey).Result()
		if err == nil && exists == 0 {
			isFirstConnection = true
		}
	}
	
	// 设备首次连接，自动订阅实时数据
	if isFirstConnection && c.config.Radar.Subscription.AutoSubscribe && c.subscriptionManager != nil {
		go func() {
			if err := c.subscriptionManager.AutoSubscribe(context.Background(), uid); err != nil {
				c.logger.Warn("Failed to auto-subscribe on device first connection",
					zap.String("uid", uid),
					zap.Error(err),
				)
			} else {
				c.logger.Info("Auto-subscribed on device first connection",
					zap.String("uid", uid),
				)
			}
		}()
	}

	// 5. 处理命令响应（如果是属性响应或功能响应，需要存储到 Redis 供 CommandService 使用）
	// 注意：prop 和 func 类型的消息是命令响应，属于同步的 request-response 模式
	// 这些响应不应该发布到 Streams，只需要存储到 Redis 供 CommandService 读取即可
	if topicInfo.Type == mqtt.TopicTypeProp {
		// 属性响应：检查是否有 requestId，如果有则存储到 Redis
		if requestID, ok := mqttData["requestId"].(string); ok && requestID != "" {
			responseKey := fmt.Sprintf("radar:response:%s", requestID)
			responseJSON, _ := json.Marshal(mqttData)
			if err := c.redisClient.Set(context.Background(), responseKey, responseJSON, 30*time.Second).Err(); err != nil {
				c.logger.Error("Failed to store property response in Redis",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Stored property response in Redis",
					zap.String("request_id", requestID),
					zap.String("uid", uid),
				)
			}
		}
		// 命令响应处理完毕，直接返回，不发布到 Streams
		return nil
	} else if topicInfo.Type == mqtt.TopicTypeFunc {
		// 功能响应：检查是否有 requestId，如果有则存储到 Redis
		if requestID, ok := mqttData["requestId"].(string); ok && requestID != "" {
			responseKey := fmt.Sprintf("radar:response:func:%s", requestID)
			responseJSON, _ := json.Marshal(mqttData)
			if err := c.redisClient.Set(context.Background(), responseKey, responseJSON, 30*time.Second).Err(); err != nil {
				c.logger.Error("Failed to store function response in Redis",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Stored function response in Redis",
					zap.String("request_id", requestID),
					zap.String("uid", uid),
				)
			}
		}
		// 命令响应处理完毕，直接返回，不发布到 Streams
		return nil
	}

	// 6. 处理数据上报消息（monitor, stat, event, alarm）
	// 这些消息需要编码后发布到 Redis Streams
	// 构建基础数据（直接展开原始数据，不保存在 raw_data 中）
	data := make(map[string]interface{})

	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Radar"
	data["topic_type"] = string(topicInfo.Type)
	data["timestamp"] = time.Now().Unix()
	data["topic"] = topic

	// 直接展开原始数据到顶层
	for k, v := range mqttData {
		data[k] = v
	}

	// 7. 调用 encode 公共函数进行编码转换
	encodedData, err := encode.RadarEncode(data, string(topicInfo.Type))
	if err != nil {
		c.logger.Error("Failed to encode radar data",
			zap.String("device_id", device.DeviceID),
			zap.String("uid", uid),
			zap.String("topic_type", string(topicInfo.Type)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to encode radar data: %w", err)
	}

	// 8. 根据 topic_type 确定输出 stream
	streamName := c.getOutputStreamName(topicInfo.Type)

	// 9. 发布到 Redis Streams
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		c.logger.Error("Failed to publish to Redis Streams",
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	c.logger.Info("Published radar data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.String("topic_type", string(topicInfo.Type)),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// handleLegacyMessage 处理旧格式的MQTT消息（兼容性）
// 主题格式: radar/{device_id}/data
func (c *MQTTConsumer) handleLegacyMessage(topic string, payload []byte) error {
	c.logger.Debug("Received legacy MQTT message",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)),
	)

	// 1. 从主题中提取设备标识符
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid legacy topic format: %s", topic)
	}
	deviceIdentifier := parts[1] // 可能是 serial_number 或 uid

	// 2. 解析消息
	var mqttData map[string]interface{}
	if err := json.Unmarshal(payload, &mqttData); err != nil {
		c.logger.Error("Failed to unmarshal legacy MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 3. 查询设备信息（如果不存在，尝试从 device_store 自动创建）
	device, err := c.deviceRepo.GetDeviceBySerialNumber(deviceIdentifier)
	if err != nil {
		// 尝试使用 UID 查询
		device, err = c.deviceRepo.GetDeviceByUID(deviceIdentifier)
		if err != nil {
			// 设备不存在，尝试从 device_store 自动创建
			device, err = c.deviceRepo.GetOrCreateDeviceFromStore(context.Background(), deviceIdentifier, topic)
			if err != nil {
				c.logger.Warn("Device not found and cannot be created from device_store",
					zap.String("identifier", deviceIdentifier),
					zap.String("mqtt_topic", topic),
					zap.Error(err),
				)
				return fmt.Errorf("device not found: %s", deviceIdentifier)
			}
			// 设备已从 device_store 自动创建
			c.logger.Info("Device auto-created from device_store on MQTT connection",
				zap.String("device_id", device.DeviceID),
				zap.String("identifier", deviceIdentifier),
				zap.String("mqtt_topic", topic),
			)
		}
	}

	// 4. 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})

	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Radar"
	data["topic_type"] = "legacy"
	data["timestamp"] = time.Now().Unix()
	data["topic"] = topic

	// 直接展开原始数据到顶层
	for k, v := range mqttData {
		data[k] = v
	}

	// 5. 调用 encode 公共函数进行编码转换
	encodedData, err := encode.RadarEncode(data, "legacy")
	if err != nil {
		c.logger.Error("Failed to encode legacy radar data",
			zap.String("device_id", device.DeviceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to encode legacy radar data: %w", err)
	}

	// 6. 发布到 Redis Streams（legacy 消息发布到默认 stream）
	streamName := "iot:data:stream"
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		c.logger.Error("Failed to publish to Redis Streams",
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	c.logger.Info("Published legacy radar data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// getOutputStreamName 根据主题类型获取输出 Redis Stream 名称（统一使用 iot: 前缀）
func (c *MQTTConsumer) getOutputStreamName(topicType mqtt.TopicType) string {
	switch topicType {
	case mqtt.TopicTypeMonitor:
		return "iot:monitor:stream" // 实时数据
	case mqtt.TopicTypeStat:
		return "iot:stat:stream" // 统计数据
	case mqtt.TopicTypeEvent:
		return "iot:event:stream" // 事件/日志
	case mqtt.TopicTypeAlarm:
		return "iot:alarm:stream" // 告警
	default:
		return "iot:data:stream" // 默认
	}
}

// getStreamNameForTopicType 根据主题类型获取 Redis Stream 名称（保留用于兼容性，但不再使用）
// 注意：属性响应和功能响应不发布到 Streams，而是存储到 Redis 用于 request-response
func (c *MQTTConsumer) getStreamNameForTopicType(topicType mqtt.TopicType) string {
	switch topicType {
	case mqtt.TopicTypeProp:
		return "radar:prop:stream" // 属性响应（不发布到 Streams）
	case mqtt.TopicTypeMonitor:
		return "iot:monitor:stream" // 实时数据
	case mqtt.TopicTypeFunc:
		return "radar:func:stream" // 功能响应（不发布到 Streams）
	case mqtt.TopicTypeStat:
		return "iot:stat:stream" // 统计数据
	case mqtt.TopicTypeEvent:
		return "iot:event:stream" // 事件/日志
	case mqtt.TopicTypeAlarm:
		return "iot:alarm:stream" // 告警
	default:
		return "iot:data:stream" // 默认
	}
}
