package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/repository"
	
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
	// 订阅雷达数据主题
	if err := c.mqttClient.Subscribe(c.config.Radar.Topics.Data, 1, c.handleMessage); err != nil {
		return fmt.Errorf("failed to subscribe to data topic: %w", err)
	}
	
	c.logger.Info("MQTT consumer started",
		zap.String("topic", c.config.Radar.Topics.Data),
	)
	
	// 等待上下文取消
	<-ctx.Done()
	return nil
}

// Stop 停止消费者
func (c *MQTTConsumer) Stop(ctx context.Context) error {
	// 取消订阅
	if err := c.mqttClient.Unsubscribe(c.config.Radar.Topics.Data); err != nil {
		c.logger.Error("Failed to unsubscribe", zap.Error(err))
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
	
	// 1. 从主题中提取设备标识符
	// 主题格式: radar/{device_id}/data
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid topic format: %s", topic)
	}
	deviceIdentifier := parts[1] // 可能是 serial_number 或 uid
	
	// 2. 解析消息
	var mqttData map[string]interface{}
	if err := json.Unmarshal(payload, &mqttData); err != nil {
		c.logger.Error("Failed to unmarshal MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	// 3. 查询设备信息
	device, err := c.deviceRepo.GetDeviceBySerialNumber(deviceIdentifier)
	if err != nil {
		// 尝试使用 UID 查询
		device, err = c.deviceRepo.GetDeviceByUID(deviceIdentifier)
		if err != nil {
			c.logger.Warn("Device not found",
				zap.String("identifier", deviceIdentifier),
				zap.Error(err),
			)
			return fmt.Errorf("device not found: %s", deviceIdentifier)
		}
	}
	
	// 4. 构建标准化数据
	standardizedData := map[string]interface{}{
		"device_id":      device.DeviceID,
		"tenant_id":      device.TenantID,
		"serial_number":  device.SerialNumber,
		"uid":            device.UID,
		"device_type":    "Radar",
		"raw_data":       mqttData,
		"timestamp":      time.Now().Unix(),
		"topic":          topic,
	}
	
	// 5. 发布到 Redis Streams
	streamName := "radar:data:stream"
	streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, standardizedData)
	if err != nil {
		c.logger.Error("Failed to publish to Redis Streams",
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish to stream: %w", err)
	}
	
	c.logger.Info("Published radar data to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	
	return nil
}

