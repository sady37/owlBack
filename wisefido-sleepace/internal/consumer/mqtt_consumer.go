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
		zap.String("stream", c.config.Sleepace.Stream),
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
	// 1. 查询设备信息
	device, err := c.deviceRepo.GetDeviceByCode(msg.DeviceId)
	if err != nil {
		c.logger.Warn("Device not found",
			zap.String("device_code", msg.DeviceId),
			zap.Error(err),
		)
		return fmt.Errorf("device not found: %s", msg.DeviceId)
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
	
	// 构建标准化数据（与 wisefido-radar 格式一致）
	standardizedData := map[string]interface{}{
		"device_id":      device.DeviceID,
		"tenant_id":      device.TenantID,
		"serial_number":  device.SerialNumber,
		"uid":            device.UID,
		"device_type":    "Sleepace", // 或 "SleepPad"
		"raw_data": map[string]interface{}{
			"breath":        realtimeData.Breath,
			"heart":         realtimeData.Heart,
			"turnOver":      realtimeData.TurnOver,
			"bodyMove":      realtimeData.BodyMove,
			"sitUp":         realtimeData.SitUp,
			"initStatus":    realtimeData.InitStatus,
			"bedStatus":     realtimeData.BedStatus,
			"signalQuality": realtimeData.SignalQuality,
			"leftRight":     realtimeData.LeftRight,
		},
		"timestamp": msg.TimeStamp,
		"topic":     "sleepace/realtime",
	}
	
	// 发布到 Redis Streams
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, c.config.Sleepace.Stream, standardizedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published sleepace realtime data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", c.config.Sleepace.Stream),
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
	
	// 构建标准化数据
	standardizedData := map[string]interface{}{
		"device_id":      device.DeviceID,
		"tenant_id":      device.TenantID,
		"serial_number":  device.SerialNumber,
		"uid":            device.UID,
		"device_type":    "Sleepace",
		"raw_data": map[string]interface{}{
			"sleepStage": sleepStageData.SleepStage,
			"leftRight":  sleepStageData.LeftRight,
		},
		"timestamp": msg.TimeStamp,
		"topic":     "sleepace/sleepStage",
	}
	
	// 发布到 Redis Streams
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, c.config.Sleepace.Stream, standardizedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published sleepace sleep stage data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", c.config.Sleepace.Stream),
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
	
	standardizedData := map[string]interface{}{
		"device_id":      device.DeviceID,
		"tenant_id":      device.TenantID,
		"serial_number":  device.SerialNumber,
		"uid":            device.UID,
		"device_type":    "Sleepace",
		"raw_data": map[string]interface{}{
			"connectionStatus": connData.ConnectionStatus,
		},
		"timestamp": msg.TimeStamp,
		"topic":     "sleepace/connectionStatus",
	}
	
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, c.config.Sleepace.Stream, standardizedData)
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Debug("Published sleepace connection status to Redis Streams",
		zap.String("device_id", device.DeviceID),
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
	
	standardizedData := map[string]interface{}{
		"device_id":      device.DeviceID,
		"tenant_id":      device.TenantID,
		"serial_number":  device.SerialNumber,
		"uid":            device.UID,
		"device_type":    "Sleepace",
		"raw_data": map[string]interface{}{
			"alarmId":       alarmData.Id,
			"alarmType":     alarmData.Type,
			"alarmStatus":   alarmData.Status,
			"userId":        alarmData.UserId,
			"relieveReason": alarmData.RelieveReason,
			"relieveTime":   alarmData.RelieveTime,
		},
		"timestamp": msg.TimeStamp,
		"topic":     "sleepace/alarmNotify",
	}
	
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, c.config.Sleepace.Stream, standardizedData)
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

