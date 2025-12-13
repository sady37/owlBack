package consumer

import (
	"context"
	"fmt"
	"time"
	"wisefido-data-transformer/internal/config"
	"wisefido-data-transformer/internal/models"
	"wisefido-data-transformer/internal/repository"
	"wisefido-data-transformer/internal/transformer"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	rediscommon "owl-common/redis"
)

// StreamMessage 适配 rediscommon.StreamMessage
type StreamMessage struct {
	ID     string
	Stream string
	Values map[string]interface{}
}

// StreamConsumer Redis Streams 消费者
type StreamConsumer struct {
	config              *config.Config
	redisClient         *redis.Client
	snomedRepo          *repository.SNOMEDRepository
	iotRepo             *repository.IoTTimeSeriesRepository
	radarTransformer    *transformer.RadarTransformer
	sleepaceTransformer *transformer.SleepaceTransformer
	logger              *zap.Logger
}

// NewStreamConsumer 创建 Streams 消费者
func NewStreamConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	snomedRepo *repository.SNOMEDRepository,
	iotRepo *repository.IoTTimeSeriesRepository,
	radarTransformer *transformer.RadarTransformer,
	sleepaceTransformer *transformer.SleepaceTransformer,
	logger *zap.Logger,
) *StreamConsumer {
	return &StreamConsumer{
		config:              cfg,
		redisClient:         redisClient,
		snomedRepo:          snomedRepo,
		iotRepo:             iotRepo,
		radarTransformer:    radarTransformer,
		sleepaceTransformer: sleepaceTransformer,
		logger:              logger,
	}
}

// Start 启动消费者
func (c *StreamConsumer) Start(ctx context.Context) error {
	// 创建消费者组
	streams := []string{
		c.config.Transformer.Streams.Radar,
		c.config.Transformer.Streams.Sleepace,
	}
	
	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.Transformer.ConsumerGroup); err != nil {
			return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
		}
	}
	
	c.logger.Info("Stream consumer started",
		zap.String("consumer_group", c.config.Transformer.ConsumerGroup),
		zap.String("consumer_name", c.config.Transformer.ConsumerName),
	)
	
	// 启动消费循环
	backoffDuration := time.Second // 初始退避时间
	maxBackoff := 30 * time.Second // 最大退避时间
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			radarErr := c.consumeStream(ctx, c.config.Transformer.Streams.Radar)
			sleepaceErr := c.consumeStream(ctx, c.config.Transformer.Streams.Sleepace)
			
			// 如果两个流都出错，才进行退避
			if radarErr != nil && sleepaceErr != nil {
				c.logger.Error("Failed to consume streams", 
					zap.Error(radarErr),
					zap.NamedError("sleepace_error", sleepaceErr),
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
				if radarErr != nil {
					c.logger.Error("Failed to consume radar stream", zap.Error(radarErr))
				}
				if sleepaceErr != nil {
					c.logger.Error("Failed to consume sleepace stream", zap.Error(sleepaceErr))
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
		c.config.Transformer.ConsumerGroup,
		c.config.Transformer.ConsumerName,
		c.config.Transformer.BatchSize,
	)
	
	if err != nil {
		return fmt.Errorf("failed to read from stream %s: %w", streamName, err)
	}
	
	// 处理消息
	for _, msg := range messages {
		streamMsg := &StreamMessage{
			ID:     msg.ID,
			Stream: msg.Stream,
			Values: msg.Values,
		}
		if err := c.processMessage(ctx, streamMsg); err != nil {
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
func (c *StreamConsumer) processMessage(ctx context.Context, streamMsg *StreamMessage) error {
	// 解析原始设备数据
	rawData, err := models.ParseRawDeviceData(streamMsg.ID, streamMsg.Stream, streamMsg.Values)
	if err != nil {
		return fmt.Errorf("failed to parse raw device data: %w", err)
	}
	
	// 根据设备类型选择转换器
	var stdData *models.StandardizedData
	switch rawData.DeviceType {
	case "Radar":
		stdData, err = c.radarTransformer.Transform(rawData)
		if err != nil {
			return fmt.Errorf("failed to transform radar data: %w", err)
		}
	case "SleepPad", "Sleepace":
		stdData, err = c.sleepaceTransformer.Transform(rawData)
		if err != nil {
			return fmt.Errorf("failed to transform sleepace data: %w", err)
		}
	default:
		return fmt.Errorf("unknown device type: %s", rawData.DeviceType)
	}
	
	// 写入 PostgreSQL
	id, err := c.iotRepo.Insert(stdData)
	if err != nil {
		return fmt.Errorf("failed to insert to iot_timeseries: %w", err)
	}
	
	// 更新位置信息（unit_id, room_id）
	unitID, roomID, err := c.iotRepo.GetDeviceLocation(stdData.DeviceID)
	if err != nil {
		c.logger.Warn("Failed to get device location", zap.Error(err))
	} else {
		if err := c.iotRepo.UpdateLocation(id, unitID, roomID); err != nil {
			c.logger.Warn("Failed to update location", zap.Error(err))
		}
	}
	
	// 发布到输出 Stream（触发下游服务）
	// 注意：device_type 从 rawData.DeviceType 获取，已在 ParseRawDeviceData 中解析
	outputData := map[string]interface{}{
		"iot_timeseries_id": id,
		"device_id":         stdData.DeviceID,
		"tenant_id":         stdData.TenantID,
		"device_type":       rawData.DeviceType, // 添加 device_type 字段
		"timestamp":         stdData.Timestamp.Unix(),
		"data_type":         stdData.DataType,
		"category":          stdData.Category,
	}
	
	if _, err := rediscommon.PublishJSONToStream(ctx, c.redisClient, c.config.Transformer.Streams.Output, outputData); err != nil {
		c.logger.Warn("Failed to publish to output stream", zap.Error(err))
	}
	
	c.logger.Info("Processed and transformed data",
		zap.String("device_id", stdData.DeviceID),
		zap.Int64("iot_timeseries_id", id),
		zap.String("data_type", stdData.DataType),
		zap.String("category", stdData.Category),
	)
	
	return nil
}

