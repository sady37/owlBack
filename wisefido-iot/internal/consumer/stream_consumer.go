package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"wisefido-iot/internal/config"
	"wisefido-iot/internal/publisher"
	"wisefido-iot/internal/repository"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	rediscommon "owl-common/redis"
)

// StreamConsumer Redis Streams 消费者
type StreamConsumer struct {
	config    *config.Config
	redisClient *redis.Client
	iotRepo   *repository.IoTTimeSeriesRepository
	publisher *publisher.StreamPublisher
	logger    *zap.Logger
}

// NewStreamConsumer 创建 Streams 消费者
func NewStreamConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	iotRepo *repository.IoTTimeSeriesRepository,
	publisher *publisher.StreamPublisher,
	logger *zap.Logger,
) *StreamConsumer {
	return &StreamConsumer{
		config:      cfg,
		redisClient: redisClient,
		iotRepo:     iotRepo,
		publisher:   publisher,
		logger:      logger,
	}
}

// Start 启动消费者
func (c *StreamConsumer) Start(ctx context.Context) error {
	// 创建消费者组 - 订阅统一 IoT streams（所有设备类型）
	streams := []string{
		c.config.Streams.Monitor, // iot:monitor:stream - 所有设备的实时数据
		c.config.Streams.Stat,    // iot:stat:stream    - 所有设备的统计数据
		c.config.Streams.Event,   // iot:event:stream   - 所有设备的事件数据
		c.config.Streams.Alarm,   // iot:alarm:stream   - 所有设备的告警数据
		c.config.Streams.Auth,    // iot:auth:stream    - 所有设备的认证数据
	}

	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.ConsumerGroup); err != nil {
			c.logger.Warn("Failed to create consumer group for stream, will retry",
				zap.String("stream", stream),
				zap.Error(err),
			)
			// 继续处理其他 streams，不中断
		}
	}

	c.logger.Info("Stream consumer started",
		zap.String("consumer_group", c.config.ConsumerGroup),
		zap.String("consumer_name", c.config.ConsumerName),
		zap.Strings("streams", streams),
	)

	// 启动消费循环
	backoffDuration := time.Second // 初始退避时间
	maxBackoff := 30 * time.Second // 最大退避时间

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 并行消费统一 IoT streams
			monitorErr := c.consumeStream(ctx, c.config.Streams.Monitor)
			statErr := c.consumeStream(ctx, c.config.Streams.Stat)
			eventErr := c.consumeStream(ctx, c.config.Streams.Event)
			alarmErr := c.consumeStream(ctx, c.config.Streams.Alarm)
			authErr := c.consumeStream(ctx, c.config.Streams.Auth)

			// 收集所有错误
			errors := []error{
				monitorErr, statErr, eventErr, alarmErr, authErr,
			}
			
			// 如果所有流都出错，才进行退避
			allFailed := true
			for _, err := range errors {
				if err == nil {
					allFailed = false
					break
				}
			}

			if allFailed {
				c.logger.Error("Failed to consume all streams",
					zap.Duration("backoff", backoffDuration),
				)

				// 指数退避：等待后重试
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoffDuration):
					// 指数退避，但不超过最大值
					backoffDuration *= 2
					if backoffDuration > maxBackoff {
						backoffDuration = maxBackoff
					}
				}
			} else {
				// 至少一个流成功时重置退避时间
				backoffDuration = time.Second

				// 记录单个流的错误（但不中断）
				if monitorErr != nil {
					c.logger.Error("Failed to consume monitor stream", zap.Error(monitorErr))
				}
				if statErr != nil {
					c.logger.Error("Failed to consume stat stream", zap.Error(statErr))
				}
				if eventErr != nil {
					c.logger.Error("Failed to consume event stream", zap.Error(eventErr))
				}
				if alarmErr != nil {
					c.logger.Error("Failed to consume alarm stream", zap.Error(alarmErr))
				}
				if authErr != nil {
					c.logger.Error("Failed to consume auth stream", zap.Error(authErr))
				}
			}
		}
	}
}

// consumeStream 消费单个 Stream
func (c *StreamConsumer) consumeStream(ctx context.Context, streamName string) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		streamName,
		c.config.ConsumerGroup,
		c.config.ConsumerName,
		c.config.BatchSize,
	)

	if err != nil {
		return fmt.Errorf("failed to read from stream %s: %w", streamName, err)
	}

	// 处理消息
	for _, msg := range messages {
		if err := c.processMessage(ctx, streamName, msg); err != nil {
			c.logger.Error("Failed to process message",
				zap.String("stream", streamName),
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}

	return nil
}

// processMessage 处理单条消息
func (c *StreamConsumer) processMessage(ctx context.Context, streamName string, msg rediscommon.StreamMessage) error {
	// 解析数据：支持两种格式
	// 1. 展开格式（推荐）：字段直接展开在 msg.Values 中（device_id, device_uid, timestamp, topic_type, category, data_value, ...）
	// 2. 包装格式（兼容）：字段包装在 msg.Values["data"] 中（JSON 字符串）
	var data map[string]interface{}

	// 优先尝试展开格式（直接使用 msg.Values）
	if len(msg.Values) > 0 {
		// 检查是否有 "data" 字段（包装格式）
		if dataStr, ok := msg.Values["data"].(string); ok {
			// 包装格式：从 "data" 字段解析 JSON
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				return fmt.Errorf("failed to unmarshal data from 'data' field: %w", err)
			}
		} else {
			// 展开格式：直接使用 msg.Values（需要转换类型）
			// Redis Stream 返回的值都是字符串（由 PublishToStream 转换）
			data = make(map[string]interface{})
			for k, v := range msg.Values {
				if strVal, ok := v.(string); ok {
					// 尝试解析 JSON 字符串（如 data_value 可能是 JSON 对象）
					if k == "data_value" {
						var jsonVal map[string]interface{}
						if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
							data[k] = jsonVal
						} else {
							// 如果不是 JSON，保持原字符串
							data[k] = strVal
						}
					} else if k == "timestamp" {
						// timestamp 可能是 int64 字符串，尝试转换
						if ts, err := strconv.ParseInt(strVal, 10, 64); err == nil {
							data[k] = ts
						} else {
							data[k] = strVal
						}
					} else {
						data[k] = strVal
					}
				} else {
					// 非字符串类型直接使用（理论上不应该出现，但兼容处理）
					data[k] = v
				}
			}
		}
	} else {
		return fmt.Errorf("invalid data format: empty message values")
	}

	// 写入 PostgreSQL（窄表结构：id, device_id, device_uid, timestamp, topic_type, category, data_values JSONB, ...）
	// data_type 在 Insert 方法中根据 topic_type 自动映射（monitor, statistics, event, alarm）
	// 所有 encode 后的数据存储在 data_values JSONB 中
	// 注意：data 的字段顺序为：device_id → device_type → tenant_id → timestamp → topic_type → data_value → 位置信息
	// category 字段保留在 data_value 内部，不提取到顶层，避免冗余
	id, err := c.iotRepo.Insert(data)
	if err != nil {
		// 如果遇到 db 与 stream 不一致，记录并跳过（根据用户要求）
		deviceID, _ := data["device_id"].(string)
		c.logger.Warn("Failed to insert to iot_timeseries, skipping message",
			zap.String("stream", streamName),
			zap.String("device_id", deviceID),
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
		// 返回 nil 以跳过此消息，继续处理下一条
		return nil
	}
	
	// 提取 deviceID 用于日志
	deviceID, _ := data["device_id"].(string)
	c.logger.Debug("Successfully inserted to iot_timeseries",
		zap.String("stream", streamName),
		zap.String("device_id", deviceID),
		zap.Int64("iot_timeseries_id", id),
	)

	// 不再发布到 iot:data:stream（已移除）

	return nil
}
