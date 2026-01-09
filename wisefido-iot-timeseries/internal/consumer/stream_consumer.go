package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-iot-timeseries/internal/config"
	"wisefido-iot-timeseries/internal/publisher"
	"wisefido-iot-timeseries/internal/repository"

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
	// 创建消费者组
	streams := []string{
		c.config.Streams.Monitor,
		c.config.Streams.Stat,
		c.config.Streams.Event,
		c.config.Streams.Alarm,
	}

	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.ConsumerGroup); err != nil {
			return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
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
			// 并行消费多个 streams
			monitorErr := c.consumeStream(ctx, c.config.Streams.Monitor)
			statErr := c.consumeStream(ctx, c.config.Streams.Stat)
			eventErr := c.consumeStream(ctx, c.config.Streams.Event)
			alarmErr := c.consumeStream(ctx, c.config.Streams.Alarm)

			// 如果所有流都出错，才进行退避
			if monitorErr != nil && statErr != nil && eventErr != nil && alarmErr != nil {
				c.logger.Error("Failed to consume all streams",
					zap.NamedError("monitor_error", monitorErr),
					zap.NamedError("stat_error", statErr),
					zap.NamedError("event_error", eventErr),
					zap.NamedError("alarm_error", alarmErr),
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
	// 解析数据（从 Redis Stream 的 data 字段读取 JSON）
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		return fmt.Errorf("invalid data format: missing 'data' field")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// 确定 data_type（与 repository 中的逻辑一致）
	dataType := "observation"
	if dt, ok := data["data_type"].(string); ok && dt != "" {
		dataType = dt
	} else if topicType, ok := data["topic_type"].(string); ok {
		// 根据 topic_type 判断
		if topicType == "alarm" {
			dataType = "alarm"
		}
	} else if dataKey, ok := data["data_key"].(string); ok {
		// Sleepace 数据根据 data_key 判断
		if dataKey == "alarmNotify" {
			dataType = "alarm"
		}
	}

	// 提取 category（如果存在）
	var category string
	if cat, ok := data["category"].(string); ok && cat != "" {
		category = cat
	}

	// 写入 PostgreSQL（位置信息已在 Insert 方法中获取并插入，无需后续 UPDATE）
	id, err := c.iotRepo.Insert(data)
	if err != nil {
		return fmt.Errorf("failed to insert to iot_timeseries: %w", err)
	}
	// 注意：unit_id 和 room_id 已在 Insert 方法中直接插入，无需后续 UpdateLocation 操作
	// 优化：从 3 次数据库操作（INSERT + SELECT + UPDATE）减少到 2 次（SELECT + INSERT）
	
	// 提取 deviceID 用于日志（如果需要）
	deviceID, _ := data["device_id"].(string)

	// 发布到下游 stream（可选，触发下游服务）
	if c.config.Streams.Output != "" {
		outputData := map[string]interface{}{
			"iot_timeseries_id": id,
			"device_id":         deviceID,
			"tenant_id":         data["tenant_id"],
			"device_type":       data["device_type"],
			"timestamp":         data["timestamp"],
			"data_type":         dataType,
		}

		// 添加 category（如果存在）
		if category != "" {
			outputData["category"] = category
		}

		// 添加 event_type（如果存在）
		if eventType, ok := data["event_type"].(string); ok && eventType != "" {
			outputData["event_type"] = eventType
		}

		if _, err := c.publisher.Publish(ctx, c.config.Streams.Output, outputData); err != nil {
			c.logger.Warn("Failed to publish to output stream",
				zap.String("stream", c.config.Streams.Output),
				zap.Error(err),
			)
		}
	}

	return nil
}
