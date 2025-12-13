package mqtt

import (
	"fmt"
	"owl-common/config"
	
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(topic string, payload []byte) error

// Client MQTT客户端封装
type Client struct {
	client mqtt.Client
	config *config.MQTTConfig
}

// NewClient 创建MQTT客户端
func NewClient(cfg *config.MQTTConfig, logger interface{}) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	
	client := mqtt.NewClient(opts)
	
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	
	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// Subscribe 订阅主题
func (c *Client) Subscribe(topic string, qos byte, handler MessageHandler) error {
	if token := c.client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		if err := handler(msg.Topic(), msg.Payload()); err != nil {
			// 记录错误，但不中断处理
			fmt.Printf("Error handling MQTT message: %v\n", err)
		}
	}); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	
	return nil
}

// Publish 发布消息
func (c *Client) Publish(topic string, qos byte, retained bool, payload []byte) error {
	token := c.client.Publish(topic, qos, retained, payload)
	token.Wait()
	
	if token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, token.Error())
	}
	
	return nil
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(topics ...string) error {
	token := c.client.Unsubscribe(topics...)
	token.Wait()
	
	if token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe: %w", token.Error())
	}
	
	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.client.Disconnect(250) // 250ms等待时间
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

