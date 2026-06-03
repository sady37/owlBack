package mqtt

import (
	"fmt"
	"owl-common/config"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(topic string, payload []byte) error

// dispatchItem 收包后入队的待处理消息（payload 已拷贝，脱离 paho 缓冲区复用）
type dispatchItem struct {
	topic   string
	payload []byte
	handler MessageHandler
}

// Client MQTT客户端封装
type Client struct {
	client mqtt.Client
	config *config.MQTTConfig
	logger *zap.Logger
	done   chan struct{}
	msgCh  chan dispatchItem
}

// NewClient 创建MQTT客户端
func NewClient(cfg *config.MQTTConfig, logger interface{}) (*Client, error) {
	lg, _ := logger.(*zap.Logger)
	if lg == nil {
		lg = zap.NewNop()
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	opts.SetCleanSession(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	// 断线自动重连 + 重连后重放订阅。关键前提：业务 handler 不能阻塞 paho 收包 goroutine，
	// 否则 paho 的掉线检测与重连机制本身也会被一起卡死（实测掉线后既不报 lost 也不重连）。
	// 收包不阻塞由下面 Subscribe 的"入队即返回"+dispatchWorker 保证。
	opts.SetAutoReconnect(true)
	opts.SetResumeSubs(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		lg.Warn("MQTT connection lost", zap.String("client_id", cfg.ClientID), zap.Error(err))
	})
	opts.SetOnConnectHandler(func(_ mqtt.Client) {
		lg.Info("MQTT connected", zap.String("client_id", cfg.ClientID))
	})

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	c := &Client{
		client: client,
		config: cfg,
		logger: lg,
		done:   make(chan struct{}),
		msgCh:  make(chan dispatchItem, 65536),
	}
	go c.dispatchWorker()
	return c, nil
}

// dispatchWorker 串行执行业务 handler，与 paho 收包 goroutine 解耦。
// paho 默认在收包 goroutine 里同步跑回调，handler 一阻塞(DB/Redis/锁)收包就停，
// 连 broker 的 PINGRESP 都读不到→keepalive 判死→掉线，且 paho 重连机制也随之卡死。
// 这里回调只"拷贝+入队即返回"，收包永不阻塞；单 worker 保证消息处理顺序。
func (c *Client) dispatchWorker() {
	for {
		select {
		case <-c.done:
			return
		case item := <-c.msgCh:
			if err := item.handler(item.topic, item.payload); err != nil {
				c.logger.Warn("error handling MQTT message", zap.String("topic", item.topic), zap.Error(err))
			}
		}
	}
}

// Subscribe 订阅主题。paho 回调只做"拷贝 payload + 入队"即返回，业务 handler 由 dispatchWorker 异步串行执行。
func (c *Client) Subscribe(topic string, qos byte, handler MessageHandler) error {
	if token := c.client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		payload := append([]byte(nil), msg.Payload()...) // paho 回调返回后复用底层缓冲区，必须拷贝
		select {
		case c.msgCh <- dispatchItem{topic: msg.Topic(), payload: payload, handler: handler}:
		case <-c.done:
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
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	c.client.Disconnect(250)
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}
