package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigStreamConsumer Config Stream 消费者（消费 config:device_status:stream）
type ConfigStreamConsumer struct {
	redisClient   *redis.Client
	logger        *zap.Logger
	consumerGroup string
	consumerName  string
	batchSize     int
}

// NewConfigStreamConsumer 创建 Config Stream 消费者
func NewConfigStreamConsumer(
	redisClient *redis.Client,
	logger *zap.Logger,
	consumerGroup, consumerName string,
	batchSize int,
) *ConfigStreamConsumer {
	return &ConfigStreamConsumer{
		redisClient:   redisClient,
		logger:        logger,
		consumerGroup: consumerGroup,
		consumerName:  consumerName,
		batchSize:     batchSize,
	}
}

// Start 启动消费者
func (c *ConfigStreamConsumer) Start(ctx context.Context) error {
	streamName := rediscommon.StreamConfigDeviceStatus.Name

	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, streamName, c.consumerGroup); err != nil {
		c.logger.Warn("Failed to create consumer group for config stream, will retry",
			zap.String("stream", streamName),
			zap.Error(err),
		)
	}

	c.logger.Info("Config Stream consumer started",
		zap.String("stream", streamName),
		zap.String("consumer_group", c.consumerGroup),
		zap.String("consumer_name", c.consumerName),
	)

	// 启动消费循环
	backoffDuration := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := c.consumeStream(ctx, streamName)
			if err != nil {
				c.logger.Error("Failed to consume config stream",
					zap.String("stream", streamName),
					zap.Error(err),
					zap.Duration("backoff", backoffDuration),
				)

				// 指数退避
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoffDuration):
					backoffDuration *= 2
					if backoffDuration > maxBackoff {
						backoffDuration = maxBackoff
					}
				}
			} else {
				backoffDuration = time.Second
			}
		}
	}
}

// consumeStream 消费单个 Stream
func (c *ConfigStreamConsumer) consumeStream(ctx context.Context, stream string) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		stream,
		c.consumerGroup,
		c.consumerName,
		int64(c.batchSize),
	)
	if err != nil {
		// redis.Nil 表示没有新消息，这是正常的，不应该返回错误
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to read from stream: %w", err)
	}

	// 处理消息
	if len(messages) > 0 {
		c.logger.Info("Received config stream messages",
			zap.Int("count", len(messages)),
			zap.String("stream", stream),
		)
	}
	for _, msg := range messages {
		if err := c.processMessage(ctx, msg); err != nil {
			c.logger.Error("Failed to process config message",
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		} else {
			c.logger.Debug("Successfully processed config message",
				zap.String("stream_id", msg.ID),
			)
		}
	}

	return nil
}

// processMessage 处理单条消息
func (c *ConfigStreamConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage) error {
	// 解析数据：支持两种格式
	// 1. 包装格式（wisefido-qinglan）：type 在 msg.Values["type"]，data 在 msg.Values["data"]（JSON 字符串，含 device_uid/status 等）
	// 2. 展开格式：字段直接展开在 msg.Values 中
	var streamData map[string]interface{}
	var eventType string
	wrapperFormat := false

	if len(msg.Values) == 0 {
		return fmt.Errorf("invalid data format: empty message values")
	}

	if dataStr, ok := msg.Values["data"].(string); ok {
		// 包装格式：type 在外层 msg.Values["type"]，data 解析为 streamData
		wrapperFormat = true
		if typeStr, ok := msg.Values["type"].(string); ok {
			eventType = typeStr
		}
		if err := json.Unmarshal([]byte(dataStr), &streamData); err != nil {
			c.logger.Error("Failed to parse message data from 'data' field",
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to unmarshal data from 'data' field: %w", err)
		}
	} else {
		// 展开格式：streamData = msg.Values，eventType 在 streamData["type"]
		streamData = make(map[string]interface{})
		for k, v := range msg.Values {
			if strVal, ok := v.(string); ok {
				if k == "data_value" {
					var jsonVal map[string]interface{}
					if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
						streamData[k] = jsonVal
					} else {
						streamData[k] = strVal
					}
				} else if k == "timestamp" {
					if ts, err := strconv.ParseInt(strVal, 10, 64); err == nil {
						streamData[k] = ts
					} else {
						streamData[k] = strVal
					}
				} else {
					streamData[k] = strVal
				}
			} else {
				streamData[k] = v
			}
		}
		if et, _ := streamData["type"].(string); et != "" {
			eventType = et
		}
	}

	// 只处理设备在线状态消息（type 为 config.device.online/offline/unsubscribed）
	if eventType != "config.device.online" && eventType != "config.device.offline" && eventType != "config.device.unsubscribed" {
		c.logger.Debug("Skipping non-device-online-status message",
			zap.String("stream_id", msg.ID),
			zap.String("event_type", eventType),
		)
		return nil
	}

	// 从事件类型提取 category：config.device.{status}
	var category string
	parts := strings.Split(eventType, ".")
	if len(parts) >= 3 && parts[0] == "config" && parts[1] == "device" {
		category = parts[2]
	}
	if category == "" {
		if s, _ := streamData["status"].(string); s != "" {
			category = s
		}
	}

	// 提取 device_uid：包装格式下 streamData 即 data payload，含 device_uid；展开格式可能嵌套在 data 中
	var deviceUID string
	if wrapperFormat {
		deviceUID, _ = streamData["device_uid"].(string)
	}
	if deviceUID == "" {
		if dataStr, ok := streamData["data"].(string); ok {
			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
				deviceUID, _ = dataMap["device_uid"].(string)
			}
		} else if dataMap, ok := streamData["data"].(map[string]interface{}); ok {
			deviceUID, _ = dataMap["device_uid"].(string)
		}
	}

	if deviceUID == "" {
		c.logger.Warn("Missing device_uid in config message data",
			zap.String("stream_id", msg.ID),
			zap.String("event_type", eventType),
			zap.Any("stream_data", streamData),
		)
		return nil // 跳过无效消息
	}

	// 将设备在线状态存储到 Redis Hash（供 ListDevices 读取）
	key := "device:online_status:" + deviceUID
	err := c.redisClient.HSet(ctx, key, "status", category).Err()
	if err != nil {
		c.logger.Warn("Failed to store device online status in Redis",
			zap.String("device_uid", deviceUID),
			zap.String("status", category),
			zap.Error(err),
		)
	} else {
		// 设置过期时间（5分钟），如果设备长时间无消息，状态会自动过期
		c.redisClient.Expire(ctx, key, 300*time.Second)
		c.logger.Info("Stored device online status in Redis",
			zap.String("device_uid", deviceUID),
			zap.String("status", category),
			zap.String("redis_key", key),
		)
	}

	return nil
}

// getMapKeys 获取 map 的所有键（用于日志）
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
