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
	// 创建消费者组 - 订阅所有设备的 streams
	streams := []string{
		// Radar 设备 streams
		c.config.Streams.RadarMonitor,
		c.config.Streams.RadarStat,
		c.config.Streams.RadarEvent,
		c.config.Streams.RadarAlarm,
		// Sleepace 设备 streams
		c.config.Streams.SleepaceMonitor,
		c.config.Streams.SleepaceEvent,
		c.config.Streams.SleepaceAlarm,
		// 注意：Sleepace 没有 stat 数据
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
			// 并行消费所有设备的 streams
			radarMonitorErr := c.consumeStream(ctx, c.config.Streams.RadarMonitor)
			radarStatErr := c.consumeStream(ctx, c.config.Streams.RadarStat)
			radarEventErr := c.consumeStream(ctx, c.config.Streams.RadarEvent)
			radarAlarmErr := c.consumeStream(ctx, c.config.Streams.RadarAlarm)
			sleepaceMonitorErr := c.consumeStream(ctx, c.config.Streams.SleepaceMonitor)
			sleepaceEventErr := c.consumeStream(ctx, c.config.Streams.SleepaceEvent)
			sleepaceAlarmErr := c.consumeStream(ctx, c.config.Streams.SleepaceAlarm)

			// 收集所有错误
			errors := []error{
				radarMonitorErr, radarStatErr, radarEventErr, radarAlarmErr,
				sleepaceMonitorErr, sleepaceEventErr, sleepaceAlarmErr,
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
				if radarMonitorErr != nil {
					c.logger.Error("Failed to consume radar monitor stream", zap.Error(radarMonitorErr))
				}
				if radarStatErr != nil {
					c.logger.Error("Failed to consume radar stat stream", zap.Error(radarStatErr))
				}
				if radarEventErr != nil {
					c.logger.Error("Failed to consume radar event stream", zap.Error(radarEventErr))
				}
				if radarAlarmErr != nil {
					c.logger.Error("Failed to consume radar alarm stream", zap.Error(radarAlarmErr))
				}
				if sleepaceMonitorErr != nil {
					c.logger.Error("Failed to consume sleepace monitor stream", zap.Error(sleepaceMonitorErr))
				}
				if sleepaceEventErr != nil {
					c.logger.Error("Failed to consume sleepace event stream", zap.Error(sleepaceEventErr))
				}
				if sleepaceAlarmErr != nil {
					c.logger.Error("Failed to consume sleepace alarm stream", zap.Error(sleepaceAlarmErr))
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

	// 写入 PostgreSQL（窄表结构：id, uid, timestamp, data_type, data_values JSONB）
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
