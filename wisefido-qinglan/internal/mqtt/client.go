package mqtt

import (
	"fmt"
	"strings"

	commonconfig "owl-common/config"
	mqttcommon "owl-common/mqtt"
	"wisefido-qinglan/internal/config"
)

// Client MQTT客户端封装
type Client struct {
	client *mqttcommon.Client
	config *config.MQTTConfig
}

// NewClient 创建MQTT客户端
func NewClient(cfg *config.MQTTConfig) (*Client, error) {
	var brokerURL string
	
	// 检查Broker URL是否已经包含了协议前缀
	broker := cfg.Broker
	if strings.HasPrefix(broker, "tcp://") || strings.HasPrefix(broker, "ssl://") || strings.HasPrefix(broker, "mqtts://") {
		// 如果已经有协议前缀，则直接使用
		brokerURL = broker
	} else {
		// 否则添加协议前缀
		brokerURL = fmt.Sprintf("tcp://%s:%d", broker, cfg.Port)
	}

	// 转换配置格式
	mqttCfg := &commonconfig.MQTTConfig{
		Broker:   brokerURL,
		ClientID: cfg.ClientID,
		Username: cfg.Username,
		Password: cfg.Password,
		QoS:      1,
	}

	// 创建客户端
	client, err := mqttcommon.NewClient(mqttCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MQTT client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// Subscribe 订阅主题
func (c *Client) Subscribe(topic string, handler mqttcommon.MessageHandler) error {
	return c.client.Subscribe(topic, 1, handler)
}

// Publish 发布消息
func (c *Client) Publish(topic string, qos byte, retained bool, payload []byte) error {
	return c.client.Publish(topic, qos, retained, payload)
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(topics ...string) error {
	return c.client.Unsubscribe(topics...)
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.client.Disconnect()
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// BuildTopic 构建主题
func (c *Client) BuildTopic(topicType, uid string) string {
	cfg := c.config.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	if prefix == "" {
		return fmt.Sprintf("/%s/%s/%s/post", topicType, productID, uid)
	}
	return fmt.Sprintf("/%s/%s/%s/%s/post", prefix, topicType, productID, uid)
}

// BuildCommandTopic 构建命令主题
func (c *Client) BuildCommandTopic(topicType, uid string) string {
	cfg := c.config.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	if prefix == "" {
		return fmt.Sprintf("/%s/%s/%s/get", topicType, productID, uid)
	}
	return fmt.Sprintf("/%s/%s/%s/%s/get", prefix, topicType, productID, uid)
}

// MessageHandler 消息处理函数
type MessageHandler func(topic string, payload []byte) error

// DefaultMessageHandler 默认消息处理器
func DefaultMessageHandler(handler MessageHandler) mqttcommon.MessageHandler {
	return func(topic string, payload []byte) error {
		// log.Printf("Received MQTT message on topic: %s", topic) // 已关闭，减少刷屏
		return handler(topic, payload)
	}
}