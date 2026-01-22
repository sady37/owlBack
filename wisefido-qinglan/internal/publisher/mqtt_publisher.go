package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/mqtt"
)

// MQTTPublisher MQTT发布者
type MQTTPublisher struct {
	config     *config.Config
	mqttClient *mqtt.Client
}

// NewMQTTPublisher 创建MQTT发布者
func NewMQTTPublisher(cfg *config.Config, mqttClient *mqtt.Client) *MQTTPublisher {
	return &MQTTPublisher{
		config:     cfg,
		mqttClient: mqttClient,
	}
}

// GetDeviceProperties 读取设备属性
func (p *MQTTPublisher) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	log.Printf("Getting device properties for %s: %v", uid, keys)
	
	// 构建命令
	command := map[string]interface{}{
		"cmd":       "read",
		"requestId": generateRequestID(),
		"data": map[string]interface{}{
			"key": keys,
		},
	}
	
	// 构建主题
	topic := p.mqttClient.BuildCommandTopic("prop", uid)
	
	// 发送命令
	if err := p.publish(topic, command); err != nil {
		return nil, fmt.Errorf("failed to publish property read command: %w", err)
	}
	
	log.Printf("Property read command sent to %s", uid)
	
	// 这里应该实现等待设备响应的逻辑
	// 从Redis读取响应数据
	return map[string]interface{}{
		"status": "command_sent",
		"uid":    uid,
		"request_id": command["requestId"],
	}, nil
}

// SetDeviceProperties 设置设备属性
func (p *MQTTPublisher) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	log.Printf("Setting device properties for %s: %v", uid, properties)
	
	// 构建命令
	command := map[string]interface{}{
		"cmd":       "update",
		"requestId": generateRequestID(),
		"data":      properties,
	}
	
	// 构建主题
	topic := p.mqttClient.BuildCommandTopic("prop", uid)
	
	// 发送命令
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish property update command: %w", err)
	}
	
	log.Printf("Property update command sent to %s", uid)
	return nil
}

// SubscribeRealtimeData 订阅实时数据
func (p *MQTTPublisher) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	log.Printf("Subscribing realtime data for %s: content=%v, duration=%d", uid, content, duration)
	
	// 验证参数
	if duration < 1 || duration > 3600 {
		return fmt.Errorf("invalid duration: %d (must be between 1 and 3600)", duration)
	}
	
	// 构建命令
	// 注意：根据厂家已知问题，content 字段必须使用字符串而不是数字
	// 文档中"0"被描述成数值，但固件要求使用字符串 "0"
	contentStr := fmt.Sprintf("%v", content)
	
	command := map[string]interface{}{
		"cmd": "subscription",
		"data": map[string]interface{}{
			"content":  contentStr, // 使用字符串格式，兼容当前固件
			"duration": duration,
		},
	}
	
	// 构建主题
	topic := p.mqttClient.BuildCommandTopic("monitor", uid)
	
	// 发送命令
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish monitor subscription command: %w", err)
	}
	
	log.Printf("Monitor subscription command sent to %s", uid)
	return nil
}

// CallDeviceFunction 调用设备功能
func (p *MQTTPublisher) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	log.Printf("Calling device function for %s: dev=%d", uid, dev)
	
	// 验证参数
	validDevValues := map[int]bool{
		0:   true, // 重启雷达和主控
		1:   true, // 只重启雷达
		2:   true, // 只重启主控
		100: true, // 清除设备数据
		101: true, // 清除雷达数据
		102: true, // 清除主控数据
	}
	if !validDevValues[dev] {
		return fmt.Errorf("invalid dev value: %d", dev)
	}
	
	// 构建命令
	command := map[string]interface{}{
		"cmd":       "control",
		"requestId": generateRequestID(),
		"data": map[string]interface{}{
			"dev": dev,
		},
	}
	
	// 构建主题
	topic := p.mqttClient.BuildCommandTopic("func", uid)
	
	// 发送命令
	if err := p.publish(topic, command); err != nil {
		return fmt.Errorf("failed to publish function control command: %w", err)
	}
	
	log.Printf("Function control command sent to %s", uid)
	return nil
}

// publish 发布消息到MQTT
func (p *MQTTPublisher) publish(topic string, data interface{}) error {
	// 序列化数据
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	// 发布消息
	if err := p.mqttClient.Publish(topic, 1, false, payload); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}
	
	log.Printf("Published to topic %s: %s", topic, string(payload))
	return nil
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}