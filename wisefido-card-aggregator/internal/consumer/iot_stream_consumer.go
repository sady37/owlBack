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

// IoTStreamConsumer IoT Stream 消费者
//
// 消费 iot:monitor/stat/event/alarm 等，写 vital-focus:card:{id}:realtime 与 :alarms。
// 注意：wisefido-data 不订阅 iot stream，仅订阅 config:device_status:stream；本 iot_stream_consumer 归属 wisefido-card-aggregator。
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
	// 订阅所有设备的 streams（报警流统一为 iot:alarm:stream）
	streams := []string{
		// Radar 设备 streams
		c.config.Aggregator.IoTStream.RadarMonitor,
		c.config.Aggregator.IoTStream.RadarStat,
		c.config.Aggregator.IoTStream.RadarEvent,
		// Sleepace 设备 streams
		c.config.Aggregator.IoTStream.SleepaceMonitor,
		c.config.Aggregator.IoTStream.SleepaceEvent,
		// 报警流：Radar / Sleepad 统一写 iot:alarm:stream；消息含 device_type，可区分报警来源与传感器
		c.config.Aggregator.IoTStream.Alarm,
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
			alarmErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.Alarm)
			sleepaceMonitorErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.SleepaceMonitor)
			sleepaceEventErr := c.consumeStream(ctx, c.config.Aggregator.IoTStream.SleepaceEvent)

			// 收集所有错误
			errors := []error{
				radarMonitorErr, radarStatErr, radarEventErr, alarmErr,
				sleepaceMonitorErr, sleepaceEventErr,
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
				if alarmErr != nil {
					c.logger.Error("Failed to consume iot alarm stream", zap.Error(alarmErr))
				}
				if sleepaceMonitorErr != nil {
					c.logger.Error("Failed to consume sleepace monitor stream", zap.Error(sleepaceMonitorErr))
				}
				if sleepaceEventErr != nil {
					c.logger.Error("Failed to consume sleepace event stream", zap.Error(sleepaceEventErr))
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
					// 尝试解析 JSON 字符串（如 data_value 可能是 JSON 对象或数组）
					if k == "data_value" {
						// 先尝试解析为数组
						var jsonArray []interface{}
						if err := json.Unmarshal([]byte(strVal), &jsonArray); err == nil {
							streamData[k] = jsonArray
						} else {
							// 如果不是数组，尝试解析为对象
							var jsonVal map[string]interface{}
							if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
								streamData[k] = jsonVal
							} else {
								// 如果不是 JSON，保持原字符串
								streamData[k] = strVal
							}
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
	
	// 2. 从stream消息中提取设备数据并更新缓存
	// 注意：现在MQTT网关直接发送到 iot:*:stream，不再经过DB，直接从事件中提取数据
	deviceData := c.extractDeviceDataFromStream(streamData, deviceID, tenantID, deviceType, topicType)
	if deviceData != nil {
		// 更新设备数据缓存（用于融合）
		if err := c.updateDeviceDataCache(ctx, cardInfo.CardID, deviceID, deviceData); err != nil {
			c.logger.Warn("Failed to update device data cache",
				zap.String("card_id", cardInfo.CardID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
		}
	}
	
	// 3. 从缓存融合卡片的所有设备数据（不再从DB读取）
	realtimeData, err := c.fuseCardDataFromCache(ctx, cardInfo.TenantID, cardInfo.CardID, cardInfo.CardType)
	if err != nil {
		c.logger.Error("Failed to fuse card data from cache",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		// 如果融合失败，继续处理报警，但不更新实时数据缓存
		realtimeData = nil
	}
	
	// 3. 检测设备直接报警（从 topic_type 和 data_value 中提取）
	// 如果 topic_type 是 "alarm"，或者 data_value 中包含报警信息，创建报警事件
	if topicType == "alarm" && c.alarmHandler != nil {
		// event_type 来自 data_value["category"], 为最早源头，保持不变
		var eventType string
		if dataValue, ok := streamData["data_value"].(map[string]interface{}); ok {
			if category, ok := dataValue["category"].(string); ok {
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
				var heartRate, respiratoryRate *int
				if realtimeData.Sleepad != nil && realtimeData.Sleepad.Heart != nil {
					heartRate = realtimeData.Sleepad.Heart
				} else if realtimeData.Radar != nil && realtimeData.Radar.Heart != nil {
					heartRate = realtimeData.Radar.Heart
				}
				if realtimeData.Sleepad != nil && realtimeData.Sleepad.Breath != nil {
					respiratoryRate = realtimeData.Sleepad.Breath
				} else if realtimeData.Radar != nil && realtimeData.Radar.Breath != nil {
					respiratoryRate = realtimeData.Radar.Breath
				}
				triggerData = &models.TriggerData{
					EventType:       eventType,
					Source:          deviceType,
					HeartRate:       heartRate,
					RespiratoryRate: respiratoryRate,
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
	// 注意：完整的 VitalFocusCard 由 AggregatorService 的定时聚合（间隔 CARD_AGGREGATION_INTERVAL，默认 1 秒）生成并写入 vital-focus:card:{card_id}:full
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

// extractDeviceDataFromStream 从stream消息中提取设备数据
func (c *IoTStreamConsumer) extractDeviceDataFromStream(
	streamData map[string]interface{},
	deviceID, tenantID, deviceType, topicType string,
) *models.IoTTimeSeries {
	// 提取时间戳
	var timestamp time.Time
	if ts, ok := streamData["timestamp"].(int64); ok {
		timestamp = time.Unix(ts, 0)
	} else if tsStr, ok := streamData["timestamp"].(string); ok {
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			timestamp = time.Unix(ts, 0)
		} else {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	data := &models.IoTTimeSeries{
		ID:         fmt.Sprintf("%s_%d", deviceID, timestamp.Unix()),
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		Timestamp:  timestamp,
	}

	// 从 data_value 中提取数据
	// 注意：data_value 在 Redis Stream 中是 JSON 字符串，需要先解析
	var dataValue map[string]interface{}
	
	// 先尝试从 streamData 中获取 data_value（可能是字符串或已解析的对象）
	if dataValueStr, ok := streamData["data_value"].(string); ok {
		// data_value 是 JSON 字符串，需要解析
		if err := json.Unmarshal([]byte(dataValueStr), &dataValue); err != nil {
			// 解析失败，尝试解析为数组
			var dataValueArray []interface{}
			if err2 := json.Unmarshal([]byte(dataValueStr), &dataValueArray); err2 == nil && len(dataValueArray) > 0 {
				if firstItem, ok := dataValueArray[0].(map[string]interface{}); ok {
					dataValue = firstItem
				}
			}
		}
	} else if dataValueMap, ok := streamData["data_value"].(map[string]interface{}); ok {
		// data_value 已经是 map（已解析）
		dataValue = dataValueMap
	} else if dataValueArray, ok := streamData["data_value"].([]interface{}); ok && len(dataValueArray) > 0 {
		// data_value 是数组（已解析）
		if firstItem, ok := dataValueArray[0].(map[string]interface{}); ok {
			dataValue = firstItem
		}
	}

	if dataValue != nil {
		// 根据 category 区分数据类型
		category, _ := dataValue["category"].(string)
		
		// Monitor track 数据（category="track"）
		if category == "track" {
			// 日志：打印接收到的 track 数据（原始 dataValue）
			c.logger.Info("[CARD_AGGREGATOR] received track data",
				zap.String("device_id", deviceID),
				zap.Any("data_value", dataValue),
				zap.String("position_x_type", fmt.Sprintf("%T", dataValue["position_x"])),
				zap.String("position_y_type", fmt.Sprintf("%T", dataValue["position_y"])),
				zap.String("position_z_type", fmt.Sprintf("%T", dataValue["position_z"])))
			
			// 提取姿态数据（pose 是 SNOMED 映射后的字符串，如 "lying", "sitting", "standing"）
			if pose, ok := dataValue["pose"].(string); ok {
				data.PostureSNOMEDCode = &pose
				data.PostureDisplay = &pose // SNOMED 映射后的值直接作为 display
			}
			// 提取 tracking_id（使用 target_id）
			if targetID, ok := dataValue["target_id"].(float64); ok {
				targetIDStr := fmt.Sprintf("%.0f", targetID)
				data.TrackingID = &targetIDStr
			} else if targetID, ok := dataValue["target_id"].(int); ok {
				// 兼容 int 类型
				targetIDStr := fmt.Sprintf("%d", targetID)
				data.TrackingID = &targetIDStr
			}
			// 提取位置数据（兼容 float64 和 int 类型）
			var posXInt, posYInt, posZInt int
			var hasPosX, hasPosY, hasPosZ bool
			if posX, ok := dataValue["position_x"].(float64); ok {
				posXInt = int(posX)
				hasPosX = true
			} else if posX, ok := dataValue["position_x"].(int); ok {
				posXInt = posX
				hasPosX = true
			} else if posX, ok := dataValue["position_x"].(int64); ok {
				posXInt = int(posX)
				hasPosX = true
			}
			if hasPosX {
				data.PositionX = &posXInt
			}
			
			if posY, ok := dataValue["position_y"].(float64); ok {
				posYInt = int(posY)
				hasPosY = true
			} else if posY, ok := dataValue["position_y"].(int); ok {
				posYInt = posY
				hasPosY = true
			} else if posY, ok := dataValue["position_y"].(int64); ok {
				posYInt = int(posY)
				hasPosY = true
			}
			if hasPosY {
				data.PositionY = &posYInt
			}
			
			if posZ, ok := dataValue["position_z"].(float64); ok {
				posZInt = int(posZ)
				hasPosZ = true
			} else if posZ, ok := dataValue["position_z"].(int); ok {
				posZInt = posZ
				hasPosZ = true
			} else if posZ, ok := dataValue["position_z"].(int64); ok {
				posZInt = int(posZ)
				hasPosZ = true
			}
			if hasPosZ {
				data.PositionZ = &posZInt
			}
			
			// 日志：打印提取后的位置数据
			c.logger.Info("[CARD_AGGREGATOR] extracted track positions",
				zap.String("device_id", deviceID),
				zap.Bool("has_pos_x", hasPosX),
				zap.Bool("has_pos_y", hasPosY),
				zap.Bool("has_pos_z", hasPosZ),
				zap.Any("position_x", data.PositionX),
				zap.Any("position_y", data.PositionY),
				zap.Any("position_z", data.PositionZ))
			
			if areaID, ok := dataValue["area_id"].(float64); ok {
				areaIDInt := int(areaID)
				data.AreaID = &areaIDInt
			} else if areaID, ok := dataValue["area_id"].(int); ok {
				areaIDInt := areaID
				data.AreaID = &areaIDInt
			}
		}
		
		// Monitor vital 数据（category="vital"）
		if category == "vital" {
			// 提取生命体征
			if hr, ok := dataValue["heart_rate"].(float64); ok {
				hrInt := int(hr)
				data.HeartRate = &hrInt
			}
			if rr, ok := dataValue["respiratory_rate"].(float64); ok {
				rrInt := int(rr)
				data.RespiratoryRate = &rrInt
			}
			// 提取睡眠状态（sleep_status 是 SNOMED 映射后的字符串）
			if sleepStatus, ok := dataValue["sleep_status"].(string); ok {
				data.SleepStateSNOMEDCode = &sleepStatus
			}
		}
		
		// Stat sleep 数据（category="sleep"）
		if category == "sleep" {
			// 提取生命体征
			if hr, ok := dataValue["heart_rate"].(float64); ok {
				hrInt := int(hr)
				data.HeartRate = &hrInt
			}
			if rr, ok := dataValue["respiratory_rate"].(float64); ok {
				rrInt := int(rr)
				data.RespiratoryRate = &rrInt
			}
			// 提取睡眠状态（sleep_state 是 SNOMED 映射后的字符串）
			if sleepState, ok := dataValue["sleep_state"].(string); ok {
				data.SleepStateSNOMEDCode = &sleepState
			}
		}
		
		// Event/Alarm 数据（category="enter2out" 等）
		if category == "enter2out" {
			// 提取床状态（event 字段，SNOMED 映射后的值，如 "left_bed", "enter_bed"）
			if event, ok := dataValue["event"].(string); ok {
				// 根据 event 值判断床状态
				if event == "left_bed" || event == "LEFT_BED" {
					bedStatus := "LEFT_BED"
					data.BedStatusSNOMEDCode = &bedStatus
				} else if event == "enter_bed" || event == "ENTER_BED" {
					bedStatus := "ENTER_BED"
					data.BedStatusSNOMEDCode = &bedStatus
				}
			}
		}
	}

	return data
}

// updateDeviceDataCache 更新设备数据缓存（Redis）
// key: vital-focus:card:{card_id}:device:{device_id}:data
//
// 按字段 merge：本次消息未带的字段保留旧值，直到 key 的 TTL 超时（6s = 2s×3）。
// 例如 track 只带 pose/tracking_id，不碰 HR/RR/sleep，则沿用该设备 cache 中旧 HR/RR/sleep；
// 只有在新消息明确带某字段时才更新，实现「无新值则保持旧值，直到超时」。
func (c *IoTStreamConsumer) updateDeviceDataCache(
	ctx context.Context,
	cardID, deviceID string,
	deviceData *models.IoTTimeSeries,
) error {
	key := fmt.Sprintf("vital-focus:card:%s:device:%s:data", cardID, deviceID)

	merged := *deviceData
	if val, err := c.redisClient.Get(ctx, key).Result(); err == nil {
		var old models.IoTTimeSeries
		if json.Unmarshal([]byte(val), &old) == nil {
			if merged.HeartRate == nil {
				merged.HeartRate = old.HeartRate
			}
			if merged.RespiratoryRate == nil {
				merged.RespiratoryRate = old.RespiratoryRate
			}
			if merged.HeartRateCode == nil {
				merged.HeartRateCode = old.HeartRateCode
			}
			if merged.HeartRateDisplay == nil {
				merged.HeartRateDisplay = old.HeartRateDisplay
			}
			if merged.RespiratoryRateCode == nil {
				merged.RespiratoryRateCode = old.RespiratoryRateCode
			}
			if merged.RespiratoryRateDisplay == nil {
				merged.RespiratoryRateDisplay = old.RespiratoryRateDisplay
			}
			if merged.SleepStateSNOMEDCode == nil {
				merged.SleepStateSNOMEDCode = old.SleepStateSNOMEDCode
			}
			if merged.SleepStateDisplay == nil {
				merged.SleepStateDisplay = old.SleepStateDisplay
			}
			if merged.BedStatusSNOMEDCode == nil {
				merged.BedStatusSNOMEDCode = old.BedStatusSNOMEDCode
			}
			if merged.BedStatusDisplay == nil {
				merged.BedStatusDisplay = old.BedStatusDisplay
			}
			if merged.PostureSNOMEDCode == nil {
				merged.PostureSNOMEDCode = old.PostureSNOMEDCode
			}
			if merged.PostureDisplay == nil {
				merged.PostureDisplay = old.PostureDisplay
			}
			if merged.TrackingID == nil {
				merged.TrackingID = old.TrackingID
			}
			if merged.PositionX == nil {
				merged.PositionX = old.PositionX
			}
			if merged.PositionY == nil {
				merged.PositionY = old.PositionY
			}
			if merged.PositionZ == nil {
				merged.PositionZ = old.PositionZ
			}
			if merged.AreaID == nil {
				merged.AreaID = old.AreaID
			}
		}
	}

	dataJSON, err := json.Marshal(&merged)
	if err != nil {
		return fmt.Errorf("failed to marshal device data: %w", err)
	}

	// 日志：打印写入 Redis 缓存的数据
	c.logger.Info("[CARD_AGGREGATOR] writing device cache",
		zap.String("card_id", cardID),
		zap.String("device_id", deviceID),
		zap.String("redis_key", key),
		zap.Any("position_x", merged.PositionX),
		zap.Any("position_y", merged.PositionY),
		zap.Any("position_z", merged.PositionZ),
		zap.String("json", string(dataJSON)))

	// 设备数据 TTL：2s×3=6s，与 HR/RR 2s 周期对齐，容错约 3 次漏报，离床后较快过期旧体征。
	if err := c.redisClient.Set(ctx, key, dataJSON, 6*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set device data cache: %w", err)
	}

	return nil
}

// fuseCardDataFromCache 从缓存融合卡片的所有设备数据（不再从DB读取）
func (c *IoTStreamConsumer) fuseCardDataFromCache(
	ctx context.Context,
	tenantID, cardID, cardType string,
) (*models.RealtimeData, error) {
	// 1. 获取卡片关联的所有设备
	devices, err := c.cardRepo.GetCardDevices(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card devices: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices found for card: %s", cardID)
	}

	// 2. 从缓存中读取所有设备的最新数据
	var sleepaceData []*models.IoTTimeSeries
	var radarData []*models.IoTTimeSeries
	var maxTimestamp time.Time

	for _, device := range devices {
		deviceType := device.DeviceType
		if deviceType != "Radar" && deviceType != "Sleepace" && deviceType != "SleepPad" && deviceType != "Sleepad" {
			continue
		}

		// 从缓存读取设备数据
		key := fmt.Sprintf("vital-focus:card:%s:device:%s:data", cardID, device.DeviceID)
		val, err := c.redisClient.Get(ctx, key).Result()
		if err != nil {
			// 缓存未命中，跳过该设备
			continue
		}

		var deviceData models.IoTTimeSeries
		if err := json.Unmarshal([]byte(val), &deviceData); err != nil {
			c.logger.Warn("Failed to unmarshal device data from cache",
				zap.String("device_id", device.DeviceID),
				zap.Error(err),
			)
			continue
		}

		// 更新最大时间戳
		if deviceData.Timestamp.After(maxTimestamp) {
			maxTimestamp = deviceData.Timestamp
		}

		// 分类数据
		if deviceType == "Sleepace" || deviceType == "SleepPad" || deviceType == "Sleepad" {
			sleepaceData = append(sleepaceData, &deviceData)
		} else if deviceType == "Radar" {
			radarData = append(radarData, &deviceData)
		}
	}

	// 3. 融合数据
	resultTimestamp := time.Now().Unix()
	if !maxTimestamp.IsZero() {
		resultTimestamp = maxTimestamp.Unix()
	}

	result := &models.RealtimeData{
		Timestamp: resultTimestamp,
		Postures:  []models.Posture{},
	}

	// 判断是否需要融合
	needFusion := len(sleepaceData) > 0 && len(radarData) > 0

	if needFusion {
		// 融合 HR/RR（优先 Sleepace）
		c.fuseVitalSignsFromCache(sleepaceData, radarData, result)
		c.fuseBedAndSleepStatusFromCache(sleepaceData, radarData, result)
	} else {
		// 直接使用设备数据
		if len(sleepaceData) > 0 {
			c.useSingleDeviceDataFromCache(sleepaceData[0], result)
		}
		if len(radarData) > 0 {
			c.useRadarDeviceDataFromCache(radarData[0], result)
		}
	}

	// 处理姿态数据（直接使用 Radar 数据）
	c.useRadarPosturesFromCache(radarData, result)

	return result, nil
}

// fuseVitalSignsFromCache 按源写入 HR/RR 到 radar / sleepad（不再融合）
func (c *IoTStreamConsumer) fuseVitalSignsFromCache(
	sleepaceData []*models.IoTTimeSeries,
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	if len(radarData) > 0 {
		v := &models.VitalSource{}
		for _, d := range radarData {
			if d.HeartRate != nil {
				v.Heart = d.HeartRate
				break
			}
		}
		for _, d := range radarData {
			if d.RespiratoryRate != nil {
				v.Breath = d.RespiratoryRate
				break
			}
		}
		result.Radar = v
	}
	if len(sleepaceData) > 0 {
		v := &models.VitalSource{}
		for _, d := range sleepaceData {
			if d.HeartRate != nil {
				v.Heart = d.HeartRate
				break
			}
		}
		for _, d := range sleepaceData {
			if d.RespiratoryRate != nil {
				v.Breath = d.RespiratoryRate
				break
			}
		}
		result.Sleepad = v
	}
}

// fuseBedAndSleepStatusFromCache 按源写入床状态、睡眠状态到 radar / sleepad
func (c *IoTStreamConsumer) fuseBedAndSleepStatusFromCache(
	sleepaceData []*models.IoTTimeSeries,
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	if result.Radar != nil && len(radarData) > 0 {
		for _, d := range radarData {
			if d.BedStatusSNOMEDCode != nil {
				result.Radar.BedStatus = d.BedStatusSNOMEDCode
				break
			}
		}
		for _, d := range radarData {
			if d.SleepStateDisplay != nil && *d.SleepStateDisplay != "" {
				result.Radar.SleepStatus = d.SleepStateDisplay
				break
			}
			if d.SleepStateSNOMEDCode != nil {
				result.Radar.SleepStatus = d.SleepStateSNOMEDCode
				break
			}
		}
	}
	if result.Sleepad != nil && len(sleepaceData) > 0 {
		for _, d := range sleepaceData {
			if d.BedStatusSNOMEDCode != nil {
				result.Sleepad.BedStatus = d.BedStatusSNOMEDCode
				break
			}
		}
		for _, d := range sleepaceData {
			if d.SleepStateDisplay != nil && *d.SleepStateDisplay != "" {
				result.Sleepad.SleepStatus = d.SleepStateDisplay
				break
			}
			if d.SleepStateSNOMEDCode != nil {
				result.Sleepad.SleepStatus = d.SleepStateSNOMEDCode
				break
			}
		}
	}
}

// useSingleDeviceDataFromCache 使用单个 Sleepace 设备的数据，按源写入 sleepad
func (c *IoTStreamConsumer) useSingleDeviceDataFromCache(
	data *models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	sleepStatus := data.SleepStateDisplay
	if sleepStatus == nil || *sleepStatus == "" {
		sleepStatus = data.SleepStateSNOMEDCode
	}
	result.Sleepad = &models.VitalSource{
		Heart:       data.HeartRate,
		Breath:      data.RespiratoryRate,
		BedStatus:   data.BedStatusSNOMEDCode,
		SleepStatus: sleepStatus,
	}
}

// useRadarDeviceDataFromCache 使用单个 Radar 设备的数据，按源写入 radar
func (c *IoTStreamConsumer) useRadarDeviceDataFromCache(
	data *models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	sleepStatus := data.SleepStateDisplay
	if sleepStatus == nil || *sleepStatus == "" {
		sleepStatus = data.SleepStateSNOMEDCode
	}
	result.Radar = &models.VitalSource{
		Heart:       data.HeartRate,
		Breath:      data.RespiratoryRate,
		BedStatus:   data.BedStatusSNOMEDCode,
		SleepStatus: sleepStatus,
	}
}

// useRadarPosturesFromCache 使用 Radar 设备的姿态数据
func (c *IoTStreamConsumer) useRadarPosturesFromCache(
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	trackingMap := make(map[string]struct {
		posture   *models.Posture
		timestamp time.Time
	})

	for _, data := range radarData {
		if data.TrackingID != nil && (data.PostureDisplay != nil || data.PostureSNOMEDCode != nil) {
			trackingID := *data.TrackingID
			postureCode := ""
			if data.PostureDisplay != nil && *data.PostureDisplay != "" {
				postureCode = *data.PostureDisplay
			} else if data.PostureSNOMEDCode != nil {
				postureCode = *data.PostureSNOMEDCode
			}
			postureDisplay := postureCode
			if data.PostureDisplay != nil && *data.PostureDisplay != "" {
				postureDisplay = *data.PostureDisplay
			}
			posture := &models.Posture{
				TrackingID:     trackingID,
				PostureCode:    postureCode,
				PostureDisplay: postureDisplay,
			}

			if data.PositionX != nil {
				posture.PositionX = data.PositionX
			}
			if data.PositionY != nil {
				posture.PositionY = data.PositionY
			}
			if data.PositionZ != nil {
				posture.PositionZ = data.PositionZ
			}
			if data.AreaID != nil {
				posture.AreaID = data.AreaID
			}

			// 如果该 tracking_id 已存在，比较时间戳，使用更新的数据
			if existing, ok := trackingMap[trackingID]; ok {
				if data.Timestamp.After(existing.timestamp) {
					trackingMap[trackingID] = struct {
						posture   *models.Posture
						timestamp time.Time
					}{
						posture:   posture,
						timestamp: data.Timestamp,
					}
				}
			} else {
				trackingMap[trackingID] = struct {
					posture   *models.Posture
					timestamp time.Time
				}{
					posture:   posture,
					timestamp: data.Timestamp,
				}
			}
		}
	}

	// 转换为列表
	for _, entry := range trackingMap {
		result.Postures = append(result.Postures, *entry.posture)
	}

	result.PersonCount = len(result.Postures)
}

