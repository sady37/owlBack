package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/mqtt"

	"go.uber.org/zap"
)

// MQTTPublisher MQTT发布者
type MQTTPublisher struct {
	config     *config.Config
	mqttClient *mqtt.Client
	logger     *zap.Logger
}

// NewMQTTPublisher 创建MQTT发布者
func NewMQTTPublisher(cfg *config.Config, mqttClient *mqtt.Client, logger *zap.Logger) *MQTTPublisher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MQTTPublisher{
		config:     cfg,
		mqttClient: mqttClient,
		logger:     logger,
	}
}

// GetDeviceProperties 读取设备属性
func (p *MQTTPublisher) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	p.logger.Debug("get device properties", zap.String("uid", uid), zap.Strings("keys", keys))

	command := map[string]interface{}{
		"cmd":       "read",
		"requestId": generateRequestID(),
	}
	if len(keys) > 0 {
		command["data"] = map[string]interface{}{"key": keys}
	}

	topic := p.mqttClient.BuildCommandTopic("prop", uid)
	if err := p.publish(topic, command); err != nil {
		return nil, fmt.Errorf("failed to publish property read command: %w", err)
	}

	p.logger.Debug("property read command sent", zap.String("uid", uid))
	return map[string]interface{}{
		"status":     "command_sent",
		"uid":        uid,
		"request_id": command["requestId"],
	}, nil
}

// SetDeviceProperties 设置设备属性
func (p *MQTTPublisher) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	p.logger.Debug("set device properties", zap.String("uid", uid), zap.Any("props", properties))

	command := map[string]interface{}{
		"cmd":       "update",
		"requestId": generateRequestID(),
		"data":      properties,
	}

	topic := p.mqttClient.BuildCommandTopic("prop", uid)
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish property update command: %w", err)
	}
	p.logger.Debug("property update command sent", zap.String("uid", uid))
	return nil
}

// SubscribeRealtimeData 订阅实时数据
func (p *MQTTPublisher) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	p.logger.Debug("subscribe realtime",
		zap.String("uid", uid),
		zap.Any("content", content),
		zap.Int("duration", duration),
	)

	if duration < 1 || duration > 3600 {
		return fmt.Errorf("invalid duration: %d (must be between 1 and 3600)", duration)
	}

	// content 必须用字符串：固件要求 "0" 而不是数字 0
	contentStr := fmt.Sprintf("%v", content)
	command := map[string]interface{}{
		"cmd": "subscription",
		"data": map[string]interface{}{
			"content":  contentStr,
			"duration": duration,
		},
	}

	topic := p.mqttClient.BuildCommandTopic("monitor", uid)
	if err := p.publish(topic, command); err != nil {
		p.logger.Warn("monitor subscription publish",
			zap.String("uid", uid),
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish monitor subscription command: %w", err)
	}
	return nil
}

// CallDeviceFunction 调用设备功能
func (p *MQTTPublisher) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	p.logger.Info("call device function", zap.String("uid", uid), zap.Int("dev", dev))

	validDevValues := map[int]bool{
		0: true, 1: true, 2: true,
		100: true, 101: true, 102: true,
	}
	if !validDevValues[dev] {
		return fmt.Errorf("invalid dev value: %d", dev)
	}

	command := map[string]interface{}{
		"cmd":       "control",
		"requestId": generateRequestID(),
		"data":      map[string]interface{}{"dev": dev},
	}

	topic := p.mqttClient.BuildCommandTopic("func", uid)
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish function control command: %w", err)
	}
	p.logger.Info("function control command sent", zap.String("uid", uid))
	return nil
}

func (p *MQTTPublisher) publish(topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}
	return nil
}

// PublishOTA publishes an OTA update command to a device via MQTT func topic
func (p *MQTTPublisher) PublishOTA(ctx context.Context, uid string, data map[string]interface{}) error {
	p.logger.Info("publish ota", zap.String("uid", uid), zap.Any("data", data))

	command := map[string]interface{}{
		"cmd":       "ota",
		"requestId": fmt.Sprintf("ota-%s-%d", uid, time.Now().UnixMilli()),
		"data":      data,
	}

	topic := p.mqttClient.BuildCommandTopic("func", uid)
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish OTA command: %w", err)
	}
	p.logger.Info("ota command sent", zap.String("uid", uid), zap.String("topic", topic))
	return nil
}

// PublishReboot publishes a reboot command (dev:0) to a device via MQTT func topic
func (p *MQTTPublisher) PublishReboot(ctx context.Context, uid string) error {
	p.logger.Info("publish reboot", zap.String("uid", uid))

	command := map[string]interface{}{
		"cmd":       "control",
		"requestId": fmt.Sprintf("reboot-%s-%d", uid, time.Now().UnixMilli()),
		"data":      map[string]interface{}{"dev": 0},
	}

	topic := p.mqttClient.BuildCommandTopic("func", uid)
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish reboot command: %w", err)
	}
	p.logger.Info("reboot command sent", zap.String("uid", uid))
	return nil
}

func generateRequestID() string {
	return "sadibaiubd123"
}
