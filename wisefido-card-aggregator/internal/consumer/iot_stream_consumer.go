package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	// 订阅所有设备的 streams
	streams := []string{
		// Radar 设备 streams
		c.config.Aggregator.IoTStream.RadarMonitor,
		c.config.Aggregator.IoTStream.RadarStat,
		c.config.Aggregator.IoTStream.RadarEvent,
		c.config.Aggregator.IoTStream.RadarAlarm,
		// Sleepace 设备 streams
		c.config.Aggregator.IoTStream.SleepaceMonitor,
		c.config.Aggregator.IoTStream.SleepaceEvent,
		c.config.Aggregator.IoTStream.SleepaceAlarm,
		// 注意：Sleepace 没有 stat 数据
	}
	
	// 创建消费者组
	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.Aggregator.IoTStream.ConsumerGroup); err != nil {
			c.logger.Warn("Failed to create consumer group for stream, will retry",
				zap.String("stream", stream),
				zap.Error(err),
			)
			// 继续处理其他 streams，不中断
		}
	}
	
	c.logger.Info("IoT Stream consumer started",
		zap.String("consumer_group", c.config.Aggregator.IoTStream.ConsumerGroup),
		zap.String("consumer_name", c.config.Aggregator.IoTStream.ConsumerName),
		zap.Strings("streams", streams),
	)
	
	// 启动消费循环
	backoffDuration := time.Second
	maxBackoff := 30 * time.Second
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 并行消费所有设备的 streams
			radarMonitorErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.RadarMonitor)
			radarStatErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.RadarStat)
			radarEventErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.RadarEvent)
			radarAlarmErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.RadarAlarm)
			sleepaceMonitorErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.SleepaceMonitor)
			sleepaceEventErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.SleepaceEvent)
			sleepaceAlarmErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.SleepaceAlarm)

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
	// 解析数据：支持两种格式
	// 1. 展开格式（推荐）：字段直接展开在 msg.Values 中（device_id, device_uid, timestamp, topic_type, category, data_value, ...）
	// 2. 包装格式（兼容）：字段包装在 msg.Values["data"] 中（JSON 字符串）
	var streamData map[string]interface{}

	// 优先尝试展开格式（直接使用 msg.Values）
	if len(msg.Values) > 0 {
		// 检查是否有 "data" 字段（包装格式）
		if dataStr, ok := msg.Values["data"].(string); ok {
			// 包装格式：从 "data" 字段解析 JSON
			if err := json.Unmarshal([]byte(dataStr), &streamData); err != nil {
				c.logger.Error("Failed to parse message data from 'data' field",
					zap.String("stream_id", msg.ID),
					zap.Error(err),
				)
				return fmt.Errorf("failed to unmarshal data from 'data' field: %w", err)
			}
		} else {
			// 展开格式：直接使用 msg.Values（需要转换类型）
			// Redis Stream 返回的值都是字符串（由 PublishToStream 转换）
			streamData = make(map[string]interface{})
			for k, v := range msg.Values {
				if strVal, ok := v.(string); ok {
					// 尝试解析 JSON 字符串（如 data_value 可能是 JSON 对象）
					if k == "data_value" {
						var jsonVal map[string]interface{}
						if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
							streamData[k] = jsonVal
						} else {
							// 如果不是 JSON，保持原字符串
							streamData[k] = strVal
						}
					} else if k == "timestamp" {
						// timestamp 可能是 int64 字符串，尝试转换
						if ts, err := strconv.ParseInt(strVal, 10, 64); err == nil {
							streamData[k] = ts
						} else {
							streamData[k] = strVal
						}
					} else {
						streamData[k] = strVal
					}
				} else {
					// 非字符串类型直接使用（理论上不应该出现，但兼容处理）
					streamData[k] = v
				}
			}
		}
	} else {
		return fmt.Errorf("invalid data format: empty message values")
	}
	
	// 提取必要字段
	deviceID, _ := streamData["device_id"].(string)
	tenantID, _ := streamData["tenant_id"].(string)
	deviceType, _ := streamData["device_type"].(string)
	topicType, _ := streamData["topic_type"].(string)
	
	if deviceID == "" || tenantID == "" {
		c.logger.Warn("Missing required fields in message",
			zap.String("stream_id", msg.ID),
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
		)
		return nil // 跳过无效消息
	}
	
	c.logger.Debug("Processing IoT data",
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
		zap.String("tenant_id", tenantID),
		zap.String("topic_type", topicType),
	)
	
	// 1. 根据 device_id 和 tenant_id 查询关联的卡片
	cardInfo, err := c.cardRepo.GetCardByDeviceID(tenantID, deviceID)
	if err != nil {
		c.logger.Warn("Card not found for device",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil // 设备可能未绑定到卡片，忽略
	}
	
	// 2. 融合卡片的所有设备数据
	realtimeData, err := c.fusion.FuseCardData(cardInfo.TenantID, cardInfo.CardID, cardInfo.CardType)
	if err != nil {
		c.logger.Error("Failed to fuse card data",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to fuse card data: %w", err)
	}
	
	// 3. 检测设备直接报警（从 topic_type 和 data_value 中提取）
	// 如果 topic_type 是 "alarm"，或者 data_value 中包含报警信息，创建报警事件
	if topicType == "alarm" && c.alarmHandler != nil {
		// 从 data_value 中提取 event_type（需要根据实际数据格式解析）
		var eventType string
		if dataValue, ok := streamData["data_value"].(map[string]interface{}); ok {
			// 尝试从 data_value 中提取报警类型
			if category, ok := dataValue["category"].(string); ok {
				// 根据 category 映射到 event_type
				eventType = category
			}
		} else if dataValueArray, ok := streamData["data_value"].([]interface{}); ok && len(dataValueArray) > 0 {
			// data_value 可能是数组
			if firstItem, ok := dataValueArray[0].(map[string]interface{}); ok {
				if category, ok := firstItem["category"].(string); ok {
					eventType = category
				}
			}
		}
		
		if eventType != "" {
			var triggerData *models.TriggerData
			if realtimeData != nil {
				triggerData = &models.TriggerData{
					EventType:       eventType,
					Source:          deviceType,
					HeartRate:       realtimeData.Heart,
					RespiratoryRate: realtimeData.Breath,
				}
			} else {
				triggerData = &models.TriggerData{
					EventType: eventType,
					Source:    deviceType,
				}
			}
			
			// 创建报警事件（不再需要 iot_timeseries_id）
			if err := c.alarmHandler.CreateDeviceAlarm(
				ctx,
				tenantID,
				deviceID,
				eventType,
				nil, // iot_timeseries_id 不再需要
				triggerData,
			); err != nil {
				c.logger.Warn("Failed to create device alarm",
					zap.String("device_id", deviceID),
					zap.String("event_type", eventType),
					zap.Error(err),
				)
				// 不返回错误，继续处理
			}
		}
	}
	
	// 4. 更新 Redis 缓存（融合后的实时数据）
	// 写入 vital-focus:card:{card_id}:realtime
	// 注意：完整的 VitalFocusCard 由 AggregatorService 的定时聚合（2秒一次）生成并写入 vital-focus:card:{card_id}:full
	if c.cacheManager != nil {
		if err := c.cacheManager.UpdateRealtimeDataCache(ctx, cardInfo.CardID, realtimeData); err != nil {
			c.logger.Error("Failed to update realtime data cache",
				zap.String("card_id", cardInfo.CardID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to update realtime data cache: %w", err)
		}
	}
	
	c.logger.Debug("Successfully processed IoT data",
		zap.String("card_id", cardInfo.CardID),
		zap.String("device_id", deviceID),
	)
	
	return nil
}

