package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/alarm"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/fusion"
	"wisefido-card-aggregator/internal/models"
	"wisefido-card-aggregator/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	rediscommon "owl-common/redis"
)

// IoTStreamConsumer IoT Stream 消费者（消费 iot:data:stream）
type IoTStreamConsumer struct {
	config          *config.Config
	redisClient     *redis.Client
	cardRepo        *repository.CardRepository
	iotRepo         *repository.IoTTimeSeriesRepository
	fusion          *fusion.SensorFusion
	cacheManager    CacheManagerInterface // 用于更新实时数据缓存
	alarmEventsRepo *repository.AlarmEventsRepository
	alarmDeviceRepo *repository.AlarmDeviceRepository
	alarmHandler    *alarm.AlarmHandler
	logger          *zap.Logger
}

// CacheManagerInterface 缓存管理器接口（避免循环依赖）
type CacheManagerInterface interface {
	UpdateRealtimeDataCache(ctx context.Context, cardID string, realtimeData *models.RealtimeData) error
}

// NewIoTStreamConsumer 创建 IoT Stream 消费者
func NewIoTStreamConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	cardRepo *repository.CardRepository,
	iotRepo *repository.IoTTimeSeriesRepository,
	fusion *fusion.SensorFusion,
	cacheManager CacheManagerInterface,
	alarmEventsRepo *repository.AlarmEventsRepository,
	alarmDeviceRepo *repository.AlarmDeviceRepository,
	alarmHandler *alarm.AlarmHandler,
	logger *zap.Logger,
) *IoTStreamConsumer {
	return &IoTStreamConsumer{
		config:          cfg,
		redisClient:     redisClient,
		cardRepo:        cardRepo,
		iotRepo:         iotRepo,
		fusion:          fusion,
		cacheManager:    cacheManager,
		alarmEventsRepo: alarmEventsRepo,
		alarmDeviceRepo: alarmDeviceRepo,
		alarmHandler:    alarmHandler,
		logger:          logger,
	}
}

// Start 启动消费者
func (c *IoTStreamConsumer) Start(ctx context.Context) error {
	stream := c.config.Aggregator.IoTStream.Stream
	
	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.Aggregator.IoTStream.ConsumerGroup); err != nil {
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}
	
	c.logger.Info("IoT Stream consumer started",
		zap.String("consumer_group", c.config.Aggregator.IoTStream.ConsumerGroup),
		zap.String("consumer_name", c.config.Aggregator.IoTStream.ConsumerName),
		zap.String("stream", stream),
	)
	
	// 启动消费循环
	backoffDuration := time.Second
	maxBackoff := 30 * time.Second
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumeStream(ctx, stream); err != nil {
				c.logger.Error("Failed to consume stream",
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
func (c *IoTStreamConsumer) consumeStream(ctx context.Context, stream string) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		stream,
		c.config.Aggregator.IoTStream.ConsumerGroup,
		c.config.Aggregator.IoTStream.ConsumerName,
		c.config.Aggregator.IoTStream.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}
	
	// 处理消息
	for _, msg := range messages {
		if err := c.processMessage(ctx, msg); err != nil {
			c.logger.Error("Failed to process message",
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}
	
	return nil
}

// processMessage 处理单条消息
func (c *IoTStreamConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage) error {
	// 解析消息数据
	var dataStr string
	if val, ok := msg.Values["data"]; ok {
		if str, ok := val.(string); ok {
			dataStr = str
		} else {
			return fmt.Errorf("invalid data format in message")
		}
	} else {
		return fmt.Errorf("missing data field in message")
	}
	
	// 解析 JSON
	var iotData models.IoTDataMessage
	if err := json.Unmarshal([]byte(dataStr), &iotData); err != nil {
		c.logger.Error("Failed to parse message data",
			zap.String("stream_id", msg.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}
	
	c.logger.Debug("Processing IoT data",
		zap.String("device_id", iotData.DeviceID),
		zap.String("device_type", iotData.DeviceType),
		zap.String("tenant_id", iotData.TenantID),
	)
	
	// 1. 根据 device_id 和 tenant_id 查询关联的卡片
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		c.logger.Warn("Card not found for device",
			zap.String("device_id", iotData.DeviceID),
			zap.String("tenant_id", iotData.TenantID),
			zap.Error(err),
		)
		return nil // 设备可能未绑定到卡片，忽略
	}
	
	// 2. 融合卡片的所有设备数据
	realtimeData, err := c.fusion.FuseCardData(cardInfo.TenantID, cardInfo.CardID, cardInfo.CardType)
	if err != nil {
		c.logger.Error("Failed to fuse card data",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", iotData.DeviceID),
			zap.String("tenant_id", iotData.TenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to fuse card data: %w", err)
	}
	
	// 3. 检测设备直接报警
	if iotData.EventType != nil && c.alarmHandler != nil {
		var triggerData *models.TriggerData
		if realtimeData != nil {
			triggerData = &models.TriggerData{
				EventType:       *iotData.EventType,
				Source:          iotData.DeviceType,
				HeartRate:       realtimeData.Heart,
				RespiratoryRate: realtimeData.Breath,
			}
		} else {
			triggerData = &models.TriggerData{
				EventType: *iotData.EventType,
				Source:    iotData.DeviceType,
			}
		}
		
		// 创建报警事件（如果失败，只记录警告，不影响数据融合流程）
		iotTimeSeriesID := &iotData.IoTTimeSeriesID
		if err := c.alarmHandler.CreateDeviceAlarm(
			ctx,
			iotData.TenantID,
			iotData.DeviceID,
			*iotData.EventType,
			iotTimeSeriesID,
			triggerData,
		); err != nil {
			c.logger.Warn("Failed to create device alarm",
				zap.String("device_id", iotData.DeviceID),
				zap.String("event_type", *iotData.EventType),
				zap.Error(err),
			)
			// 不返回错误，继续处理
		}
	}
	
	// 4. 更新 Redis 缓存（融合后的实时数据）
	// 写入 vital-focus:card:{card_id}:realtime
	// 注意：完整的 VitalFocusCard 由 AggregatorService 的定时聚合（2秒一次）生成并写入 vital-focus:card:{card_id}:full
	if c.cacheManager != nil {
		if err := c.cacheManager.UpdateRealtimeDataCache(ctx, cardInfo.CardID, realtimeData); err != nil {
			c.logger.Error("Failed to update realtime data cache",
				zap.String("card_id", cardInfo.CardID),
				zap.String("device_id", iotData.DeviceID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to update realtime data cache: %w", err)
		}
	}
	
	c.logger.Debug("Successfully processed IoT data",
		zap.String("card_id", cardInfo.CardID),
		zap.String("device_id", iotData.DeviceID),
	)
	
	return nil
}

