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

func isLoopbackMQTTBroker(host string) bool {
	h := strings.TrimSpace(strings.ToLower(host))
	h = strings.TrimPrefix(h, "tcp://")
	if i := strings.Index(h, ":"); i >= 0 {
		h = h[:i]
	}
	switch h {
	case "", "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

// effectiveMQTTConnect 本服务订阅/发布所用的 broker 与账号。
// 若顶层 mqtt.broker 为回环而 radar_device_mqtt.server 为外网地址，则设备与青兰不在同一 broker，会永远收不到上行；此时改用雷达配置建连（与主题里的 product_id/prefix 一致）。
func effectiveMQTTConnect(cfg *config.MQTTConfig) (host string, port int, username, password string) {
	host = strings.TrimSpace(cfg.Broker)
	port = cfg.Port
	username = cfg.Username
	password = cfg.Password
	rs := strings.TrimSpace(cfg.RadarDeviceMQTT.Server)
	if rs == "" {
		return host, port, username, password
	}
	if isLoopbackMQTTBroker(host) && !isLoopbackMQTTBroker(rs) {
		host = rs
		if cfg.RadarDeviceMQTT.Port > 0 {
			port = cfg.RadarDeviceMQTT.Port
		}
		if username == "" && password == "" {
			username = cfg.RadarDeviceMQTT.Account
			password = cfg.RadarDeviceMQTT.Password
		}
	}
	return host, port, username, password
}

// EffectiveBrokerDialString 与 NewClient 实际拨号地址一致，供启动日志使用。
func EffectiveBrokerDialString(cfg *config.MQTTConfig) string {
	broker, port, _, _ := effectiveMQTTConnect(cfg)
	if strings.HasPrefix(broker, "tcp://") || strings.HasPrefix(broker, "ssl://") || strings.HasPrefix(broker, "mqtts://") {
		return broker
	}
	return fmt.Sprintf("tcp://%s:%d", broker, port)
}

// NewClient 创建MQTT客户端
func NewClient(cfg *config.MQTTConfig) (*Client, error) {
	var brokerURL string

	broker, port, user, pass := effectiveMQTTConnect(cfg)

	// 检查Broker URL是否已经包含了协议前缀
	if strings.HasPrefix(broker, "tcp://") || strings.HasPrefix(broker, "ssl://") || strings.HasPrefix(broker, "mqtts://") {
		brokerURL = broker
	} else {
		brokerURL = fmt.Sprintf("tcp://%s:%d", broker, port)
	}

	// 转换配置格式
	mqttCfg := &commonconfig.MQTTConfig{
		Broker:   brokerURL,
		ClientID: cfg.ClientID,
		Username: user,
		Password: pass,
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
		return handler(topic, payload)
	}
}