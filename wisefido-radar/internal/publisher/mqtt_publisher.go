package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"wisefido-radar/internal/config"
	"wisefido-radar/pkg/mqtt"
	
	"go.uber.org/zap"
	mqttcommon "owl-common/mqtt"
)

// MQTTPublisher MQTT 发布者
// 用于向 Radar 设备发送命令
type MQTTPublisher struct {
	config     *config.Config
	mqttClient *mqttcommon.Client
	logger     *zap.Logger
}

// NewMQTTPublisher 创建 MQTT 发布者
func NewMQTTPublisher(cfg *config.Config, mqttClient *mqttcommon.Client, logger *zap.Logger) *MQTTPublisher {
	return &MQTTPublisher{
		config:     cfg,
		mqttClient: mqttClient,
		logger:     logger,
	}
}

// PublishPropertyReadCommand 发布属性读取命令
// 参考协议文档 3.4 节和 third-api-testing-program/radar-ql-v3/radar-v3-json-massages.md
// 主题：/prefix/prop/productId/UID/get
// 消息：{"cmd":"read","data":{"key":[]},"requestId":"..."}
func (p *MQTTPublisher) PublishPropertyReadCommand(ctx context.Context, uid string, keys []string, requestID string) error {
	cfg := p.config.Radar.DeviceMQTT
	
	// 构建主题
	topic := mqtt.BuildPropertyGetTopic(cfg.Prefix, cfg.ProductID, uid)
	
	// 构建命令消息
	command := map[string]interface{}{
		"cmd":       "read",
		"requestId": requestID,
		"data": map[string]interface{}{
			"key": keys,
		},
	}
	
	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal property read command: %w", err)
	}
	
	// 发布消息
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish property read command: %w", err)
	}
	
	p.logger.Info("Published property read command",
		zap.String("uid", uid),
		zap.String("topic", topic),
		zap.String("request_id", requestID),
		zap.Strings("keys", keys),
	)
	
	return nil
}

// PublishPropertyUpdateCommand 发布属性设置命令
// 参考协议文档 3.4 节
// 主题：/prefix/prop/productId/UID/get
// 消息：{"cmd":"update","data":{...},"requestId":"..."}
func (p *MQTTPublisher) PublishPropertyUpdateCommand(ctx context.Context, uid string, properties map[string]interface{}, requestID string) error {
	cfg := p.config.Radar.DeviceMQTT
	
	// 构建主题
	topic := mqtt.BuildPropertyGetTopic(cfg.Prefix, cfg.ProductID, uid)
	
	// 构建命令消息
	command := map[string]interface{}{
		"cmd":       "update",
		"requestId": requestID,
		"data":      properties,
	}
	
	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal property update command: %w", err)
	}
	
	// 发布消息
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish property update command: %w", err)
	}
	
	p.logger.Info("Published property update command",
		zap.String("uid", uid),
		zap.String("topic", topic),
		zap.String("request_id", requestID),
		zap.Any("properties", properties),
	)
	
	return nil
}

// PublishMonitorSubscriptionCommand 发布实时数据订阅命令
// 参考协议文档 3.7.1 节和 third-api-testing-program/radar-ql-v3/radar-v3-json-massages.md
// 主题：/prefix/monitor/productId/UID/get
// 消息：{"cmd":"subscription","data":{"content":"0","duration":3600}}
func (p *MQTTPublisher) PublishMonitorSubscriptionCommand(ctx context.Context, uid string, content interface{}, duration int) error {
	cfg := p.config.Radar.DeviceMQTT
	
	// 构建主题
	topic := mqtt.BuildMonitorGetTopic(cfg.Prefix, cfg.ProductID, uid)
	
	// 构建命令消息
	// 注意：content 可能是数字或字符串，根据测试程序使用字符串
	command := map[string]interface{}{
		"cmd": "subscription",
		"data": map[string]interface{}{
			"content":  content,  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
			"duration": duration, // 订阅时长（秒），最大 3600
		},
	}
	
	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal monitor subscription command: %w", err)
	}
	
	// 发布消息
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish monitor subscription command: %w", err)
	}
	
	p.logger.Info("Published monitor subscription command",
		zap.String("uid", uid),
		zap.String("topic", topic),
		zap.Any("content", content),
		zap.Int("duration", duration),
	)
	
	return nil
}

// PublishFunctionControlCommand 发布功能调用命令（重启等）
// 参考协议文档 3.8.1 节
// 主题：/prefix/func/productId/UID/get
// 消息：{"cmd":"control","data":{"dev":0},"requestId":"..."}
func (p *MQTTPublisher) PublishFunctionControlCommand(ctx context.Context, uid string, dev int, requestID string) error {
	cfg := p.config.Radar.DeviceMQTT
	
	// 构建主题
	topic := mqtt.BuildFunctionGetTopic(cfg.Prefix, cfg.ProductID, uid)
	
	// 构建命令消息
	command := map[string]interface{}{
		"cmd":       "control",
		"requestId": requestID,
		"data": map[string]interface{}{
			"dev": dev, // 0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
		},
	}
	
	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal function control command: %w", err)
	}
	
	// 发布消息
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish function control command: %w", err)
	}
	
	p.logger.Info("Published function control command",
		zap.String("uid", uid),
		zap.String("topic", topic),
		zap.String("request_id", requestID),
		zap.Int("dev", dev),
	)
	
	return nil
}

