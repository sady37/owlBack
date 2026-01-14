package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/models"
	"wisefido-sleepace/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	rediscommon "owl-common/redis"
	mqttcommon "owl-common/mqtt"
	"owl-common/encode"
)

// MQTTConsumer MQTT消息消费者
type MQTTConsumer struct {
	config     *config.Config
	mqttClient *mqttcommon.Client
	redisClient *redis.Client
	deviceRepo *repository.DeviceRepository
	logger     *zap.Logger
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
		config:     cfg,
		mqttClient: mqttClient,
		redisClient: redisClient,
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

// Start 启动消费者
func (c *MQTTConsumer) Start(ctx context.Context) error {
	// 订阅 Sleepace MQTT 主题（v1.0 格式，Sleepace 厂家提供的主题）
	// 主题格式由 Sleepace 厂家定义，通常在配置中指定
	topic := c.config.Sleepace.Topic // 从配置读取，如 "sleepace-57136"
	if topic == "" {
		return fmt.Errorf("sleepace MQTT topic not configured")
	}
	
	if err := c.mqttClient.Subscribe(topic, 1, c.handleMessage); err != nil {
		return fmt.Errorf("failed to subscribe to sleepace topic: %w", err)
	}
	
	c.logger.Info("MQTT consumer started",
		zap.String("topic", topic),
	)
	
	// 等待上下文取消
	<-ctx.Done()
	return nil
}

// Stop 停止消费者
func (c *MQTTConsumer) Stop(ctx context.Context) error {
	// 取消订阅
	topic := c.config.Sleepace.Topic
	if topic != "" {
		if err := c.mqttClient.Unsubscribe(topic); err != nil {
			c.logger.Error("Failed to unsubscribe", zap.Error(err))
		}
	}
	
	c.logger.Info("MQTT consumer stopped")
	return nil
}

// handleMessage 处理MQTT消息
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	c.logger.Debug("Received MQTT message",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)),
	)
	
	// 1. 解析 Sleepace MQTT 消息（v1.0 格式）
	// Sleepace 消息格式：数组，每个元素是一个 ReceivedMessage
	var messages []models.ReceivedMessage
	if err := json.Unmarshal(payload, &messages); err != nil {
		c.logger.Error("Failed to unmarshal Sleepace MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	// 2. 处理每个消息
	for _, msg := range messages {
		if err := c.processMessage(&msg); err != nil {
			c.logger.Error("Failed to process message",
				zap.String("device_id", msg.DeviceId),
				zap.String("data_key", msg.DataKey),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}
	
	return nil
}

// processMessage 处理单条 Sleepace 消息
func (c *MQTTConsumer) processMessage(msg *models.ReceivedMessage) error {
	// 1. 查询设备信息（如果不存在，尝试从 device_store 自动创建）
	device, err := c.deviceRepo.GetDeviceByCode(msg.DeviceId)
	if err != nil {
		// 设备不存在，尝试从 device_store 自动创建
		device, err = c.deviceRepo.GetOrCreateDeviceFromStore(context.Background(), msg.DeviceId, "sleepace/realtime")
		if err != nil {
			c.logger.Warn("Device not found and cannot be created from device_store",
				zap.String("device_code", msg.DeviceId),
				zap.String("data_key", msg.DataKey),
				zap.Error(err),
			)
			return fmt.Errorf("device not found: %s", msg.DeviceId)
		}
		// 设备已从 device_store 自动创建
		c.logger.Info("Device auto-created from device_store on MQTT connection",
			zap.String("device_id", device.DeviceID),
			zap.String("device_code", msg.DeviceId),
			zap.String("data_key", msg.DataKey),
		)
	}
	
	// 2. 根据 DataKey 处理不同类型的数据
	// 只处理需要发布到 Streams 的数据类型（realtime, sleepStage 等）
	// connectionStatus 和 alarmNotify 可以单独处理或也发布到 Streams
	
	switch msg.DataKey {
	case "realtime":
		return c.handleRealtimeData(msg, device)
	case "sleepStage":
		return c.handleSleepStageData(msg, device)
	case "connectionStatus":
		// 连接状态可以发布到 Streams 或单独处理
		return c.handleConnectionStatus(msg, device)
	case "alarmNotify":
		// 报警通知可以发布到 Streams 或单独处理
		return c.handleAlarmNotify(msg, device)
	default:
		// 其他类型的数据可以忽略或单独处理
		c.logger.Debug("Unhandled data key",
			zap.String("data_key", msg.DataKey),
			zap.String("device_id", msg.DeviceId),
		)
		return nil
	}
}

// handleRealtimeData 处理实时数据
func (c *MQTTConsumer) handleRealtimeData(msg *models.ReceivedMessage, device *repository.Device) error {
	// 解析实时数据
	var realtimeData models.RealtimeData
	if err := json.Unmarshal(msg.Data, &realtimeData); err != nil {
		return fmt.Errorf("failed to unmarshal realtime data: %w", err)
	}
	
	// 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})
	
	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Sleepace"
	data["data_key"] = "realtime"
	data["timestamp"] = msg.TimeStamp
	
	// 直接展开原始数据到顶层
	data["breath"] = realtimeData.Breath
	data["heart"] = realtimeData.Heart
	data["turnOver"] = realtimeData.TurnOver
	data["bodyMove"] = realtimeData.BodyMove
	data["sitUp"] = realtimeData.SitUp
	data["initStatus"] = realtimeData.InitStatus
	data["bedStatus"] = realtimeData.BedStatus
	data["signalQuality"] = realtimeData.SignalQuality
	data["leftRight"] = realtimeData.LeftRight
	
	// 调用 encode 公共函数进行编码转换
	encodedData, err := encode.SleepaceEncode(data, "realtime")
	if err != nil {
		return fmt.Errorf("failed to encode sleepace data: %w", err)
	}
	
	// 调整字段顺序并添加 category 和 topic_type
	encodedData = c.adjustFieldOrder(encodedData, "realtime", "monitor", "monitor")
	
	// 发布到 Redis Streams
	// 数据流分类规则（设备级别 streams）：
	// - realtime/sleepStage → sleepace:monitor:stream (实时数据)
	// - connectionStatus → sleepace:event:stream (事件/日志)
	// - alarmNotify → sleepace:alarm:stream (告警)
	streamName := "sleepace:monitor:stream" // 实时数据
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published sleepace realtime data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	
	return nil
}

// handleSleepStageData 处理睡眠阶段数据
func (c *MQTTConsumer) handleSleepStageData(msg *models.ReceivedMessage, device *repository.Device) error {
	// 解析睡眠阶段数据
	var sleepStageData models.SleepStageData
	if err := json.Unmarshal(msg.Data, &sleepStageData); err != nil {
		return fmt.Errorf("failed to unmarshal sleep stage data: %w", err)
	}
	
	// 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})
	
	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Sleepace"
	data["data_key"] = "sleepStage"
	data["timestamp"] = msg.TimeStamp
	
	// 直接展开原始数据到顶层
	data["sleepStage"] = sleepStageData.SleepStage
	data["leftRight"] = sleepStageData.LeftRight
	
	// 调用 encode 公共函数进行编码转换
	encodedData, err := encode.SleepaceEncode(data, "sleepStage")
	if err != nil {
		return fmt.Errorf("failed to encode sleepace data: %w", err)
	}
	
	// 调整字段顺序并添加 category 和 topic_type
	encodedData = c.adjustFieldOrder(encodedData, "sleepStage", "statistics", "statistics")
	
	// 发布到 Redis Streams
	// 睡眠阶段数据属于实时数据，发布到 monitor stream
	streamName := "sleepace:monitor:stream" // 实时数据
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published sleepace sleep stage data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	
	return nil
}

// handleConnectionStatus 处理连接状态数据
func (c *MQTTConsumer) handleConnectionStatus(msg *models.ReceivedMessage, device *repository.Device) error {
	// 连接状态数据可以发布到 Streams 或单独处理
	// 这里选择发布到 Streams，保持数据流统一
	var connData models.ConnectionStatusData
	if err := json.Unmarshal(msg.Data, &connData); err != nil {
		return fmt.Errorf("failed to unmarshal connection status data: %w", err)
	}
	
	// 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})
	
	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Sleepace"
	data["data_key"] = "connectionStatus"
	data["timestamp"] = msg.TimeStamp
	
	// 直接展开原始数据到顶层
	data["connectionStatus"] = connData.ConnectionStatus
	
	// 调用 encode 公共函数进行编码转换
	encodedData, err := encode.SleepaceEncode(data, "connectionStatus")
	if err != nil {
		return fmt.Errorf("failed to encode sleepace data: %w", err)
	}
	
	// 调整字段顺序并添加 category 和 topic_type
	encodedData = c.adjustFieldOrder(encodedData, "connectionStatus", "event", "event")
	
	// 发布到 Redis Streams
	// 连接状态属于事件数据，发布到 event stream
	streamName := "sleepace:event:stream" // 事件/日志
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Debug("Published sleepace connection status to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	
	return nil
}

// handleAlarmNotify 处理报警通知数据
func (c *MQTTConsumer) handleAlarmNotify(msg *models.ReceivedMessage, device *repository.Device) error {
	// 报警通知数据可以发布到 Streams 或单独处理
	// 这里选择发布到 Streams，保持数据流统一
	var alarmData models.AlarmNotifyData
	if err := json.Unmarshal(msg.Data, &alarmData); err != nil {
		return fmt.Errorf("failed to unmarshal alarm notify data: %w", err)
	}
	
	// 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})
	
	// 添加元数据
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["serial_number"] = device.SerialNumber
	data["uid"] = device.UID
	data["device_type"] = "Sleepace"
	data["data_key"] = "alarmNotify"
	data["timestamp"] = msg.TimeStamp
	
	// 直接展开原始数据到顶层
	data["alarmId"] = alarmData.Id
	data["alarmType"] = alarmData.Type
	data["alarmStatus"] = alarmData.Status
	data["userId"] = alarmData.UserId
	data["relieveReason"] = alarmData.RelieveReason
	data["relieveTime"] = alarmData.RelieveTime
	
	// 调用 encode 公共函数进行编码转换
	encodedData, err := encode.SleepaceEncode(data, "alarmNotify")
	if err != nil {
		return fmt.Errorf("failed to encode sleepace data: %w", err)
	}
	
	// 调整字段顺序并添加 category 和 topic_type
	encodedData = c.adjustFieldOrder(encodedData, "alarmNotify", "alarm", "alarm")
	
	// 发布到 Redis Streams
	// 告警通知属于告警数据，发布到 alarm stream
	streamName := "sleepace:alarm:stream" // 告警
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published sleepace alarm notify to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("alarm_type", alarmData.Type),
		zap.String("stream_id", streamID),
	)
	
	return nil
}

// adjustFieldOrder 调整字段顺序并添加 category 和 topic_type
// dataKey: 原始数据键 ("realtime", "sleepStage", "connectionStatus", "alarmNotify")
// category: 类别 ("monitor", "statistics", "event", "alarm")
// topicType: 主题类型 ("monitor", "statistics", "event", "alarm")
// 返回: 调整后的数据，字段顺序：device_id → device_type → tenant_id → timestamp → topic_type → category → 其他字段
func (c *MQTTConsumer) adjustFieldOrder(encodedData map[string]interface{}, dataKey, category, topicType string) map[string]interface{} {
	// 提取必需字段
	deviceID, _ := encodedData["device_id"]
	deviceType, _ := encodedData["device_type"]
	tenantID, _ := encodedData["tenant_id"]
	timestamp, _ := encodedData["timestamp"]
	
	// 构建新对象，按推荐顺序排列字段
	adjusted := make(map[string]interface{})
	
	// 1. 标识组
	adjusted["device_id"] = deviceID
	adjusted["device_type"] = deviceType
	adjusted["tenant_id"] = tenantID
	
	// 2. 时间组
	adjusted["timestamp"] = timestamp
	
	// 3. 分类组
	adjusted["topic_type"] = topicType
	adjusted["category"] = category
	
	// 4. 数据组和其他字段（保留 data_key 和其他编码后的字段）
	for k, v := range encodedData {
		// 跳过已处理的字段
		if k == "device_id" || k == "device_type" || k == "tenant_id" || k == "timestamp" {
			continue
		}
		adjusted[k] = v
	}
	
	return adjusted
}

