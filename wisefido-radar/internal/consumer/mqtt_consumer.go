package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"wisefido-radar/internal/alarm"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/repository"
	"wisefido-radar/pkg/mqtt"

	mqttcommon "owl-common/mqtt"
	rediscommon "owl-common/redis"
	"wisefido-radar/internal/encode"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// MQTTConsumer MQTT消息消费者
// 根据协议文档 3.3.3 节，订阅设备发布的 6 个主题
type MQTTConsumer struct {
	deviceCache         *sync.Map // 设备缓存（key: uid, value: *DeviceWithLocation）
	config              *config.Config
	mqttClient          *mqttcommon.Client
	redisClient         *redis.Client
	deviceRepo          *repository.DeviceRepository
	alarmHandler        *alarm.DeviceAlarmHandler
	logger              *zap.Logger
	subscriptionManager interface { // 避免循环依赖，使用接口
		AutoSubscribe(ctx context.Context, uid string) error
	}
}

// NewMQTTConsumer 创建MQTT消费者
func NewMQTTConsumer(
	cfg *config.Config,
	mqttClient *mqttcommon.Client,
	redisClient *redis.Client,
	deviceRepo *repository.DeviceRepository,
	logger *zap.Logger,
) *MQTTConsumer {
	// 创建设备报警处理器
	alarmHandler := alarm.NewDeviceAlarmHandler(deviceRepo, logger)

	return &MQTTConsumer{
		deviceCache:  &sync.Map{},
		config:       cfg,
		mqttClient:   mqttClient,
		redisClient:  redisClient,
		deviceRepo:   deviceRepo,
		alarmHandler: alarmHandler,
		logger:       logger,
	}
}

// SetSubscriptionManager 设置订阅管理器（避免循环依赖）
func (c *MQTTConsumer) SetSubscriptionManager(manager interface {
	AutoSubscribe(ctx context.Context, uid string) error
}) {
	c.subscriptionManager = manager
}

// Start 启动消费者
// 根据协议文档 3.3.3 节，订阅设备发布的 6 个主题：
// 1. /prefix/prop/productId/UID/post - 属性响应
// 2. /prefix/monitor/productId/UID/post - 实时数据
// 3. /prefix/func/productId/UID/post - 功能响应
// 4. /prefix/stat/productId/UID/post - 统计数据
// 5. /prefix/event/productId/UID/post - 事件/日志
// 6. /prefix/alarm/productId/UID/post - 告警
func (c *MQTTConsumer) Start(ctx context.Context) error {
	cfg := c.config.Radar.DeviceMQTT

	// 构建订阅主题（使用通配符 + 匹配所有设备）
	// 注意：如果 prefix 为空，格式为 /{type}/{productId}/+/post
	// 如果 prefix 不为空，格式为 /{prefix}/{type}/{productId}/+/post
	topics := []string{
		mqtt.BuildPropertyPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildMonitorPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildFunctionPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildStatPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildEventPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildAlarmPostTopic(cfg.Prefix, cfg.ProductID, "+"),
	}

	// 订阅所有主题
	for _, topic := range topics {
		if err := c.mqttClient.Subscribe(topic, 1, c.handleMessage); err != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
		c.logger.Info("Subscribed to MQTT topic",
			zap.String("topic", topic),
		)
	}

	// 同时保留对旧格式主题的兼容（radar/+/data）
	if c.config.Radar.Topics.Data != "" {
		if err := c.mqttClient.Subscribe(c.config.Radar.Topics.Data, 1, c.handleLegacyMessage); err != nil {
			c.logger.Warn("Failed to subscribe to legacy data topic",
				zap.String("topic", c.config.Radar.Topics.Data),
				zap.Error(err),
			)
		} else {
			c.logger.Info("Subscribed to legacy MQTT topic",
				zap.String("topic", c.config.Radar.Topics.Data),
			)
		}
	}

	c.logger.Info("MQTT consumer started",
		zap.Int("topics_count", len(topics)),
	)

	// 等待上下文取消
	<-ctx.Done()
	return nil
}

// ClearDeviceCache 清除指定设备的缓存
func (c *MQTTConsumer) ClearDeviceCache(tenantID, deviceID string) {
	if tenantID == "" || deviceID == "" {
		return
	}

	// 需要根据 deviceID 找到对应的 UID
	var uid string
	var found bool

	c.deviceCache.Range(func(key, value interface{}) bool {
		deviceWithLocation := value.(*DeviceWithLocation)
		if deviceWithLocation.Device.DeviceID == deviceID && deviceWithLocation.Device.TenantID == tenantID {
			uid = key.(string)
			found = true
			return false // 停止遍历
		}
		return true // 继续遍历
	})

	if !found {
		c.logger.Warn("Device not found in cache, cannot clear",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)
		return
	}

	// 清除缓存
	c.deviceCache.Delete(uid)
	c.logger.Info("Cleared device cache",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("uid", uid),
	)
}

// RefreshDeviceCache 原子更新设备缓存
func (c *MQTTConsumer) RefreshDeviceCache(tenantID, deviceID string) {
	if tenantID == "" || deviceID == "" {
		return
	}

	// 异步原子更新
	go c.refreshDeviceCacheAsync(tenantID, deviceID)
}

// refreshDeviceCacheAsync 异步原子更新设备缓存
func (c *MQTTConsumer) refreshDeviceCacheAsync(tenantID, deviceID string) {
	// 1. 根据 deviceID 找到对应的 UID
	var uid string

	c.deviceCache.Range(func(key, value interface{}) bool {
		deviceWithLocation := value.(*DeviceWithLocation)
		if deviceWithLocation.Device.DeviceID == deviceID && deviceWithLocation.Device.TenantID == tenantID {
			uid = key.(string)
			return false // 停止遍历
		}
		return true // 继续遍历
	})

	if uid == "" {
		// 设备不在缓存中，不需要刷新
		c.logger.Debug("Device not in cache, skipping refresh",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)
		return
	}

	// 2. 直接调用现有的 getOrCreateDeviceWithLocation 函数
	// 它会自动查询最新信息并更新缓存
	_, _, err := c.getOrCreateDeviceWithLocation(uid, "config:refresh")
	if err != nil {
		c.logger.Error("Failed to refresh device cache",
			zap.String("device_id", deviceID),
			zap.String("uid", uid),
			zap.Error(err),
		)
		return
	}

	c.logger.Info("Refreshed device cache",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("uid", uid),
	)
}

// ClearDeviceCacheByUID 根据 UID 清除设备缓存
func (c *MQTTConsumer) ClearDeviceCacheByUID(uid string) {
	if uid == "" {
		return
	}

	c.deviceCache.Delete(uid)
	c.logger.Info("Cleared device cache by UID",
		zap.String("uid", uid),
	)
}

// ClearCacheForUnit 清除指定 unit 下所有设备的缓存
func (c *MQTTConsumer) ClearCacheForUnit(tenantID, unitID string) {
	if tenantID == "" || unitID == "" {
		return
	}

	// 遍历缓存，清除属于该 unit 的设备
	devicesCleared := 0
	c.deviceCache.Range(func(key, value interface{}) bool {
		deviceWithLocation := value.(*DeviceWithLocation)
		if deviceWithLocation.Device.TenantID == tenantID &&
			deviceWithLocation.LocationInfo != nil &&
			deviceWithLocation.LocationInfo.UnitID != nil &&
			*deviceWithLocation.LocationInfo.UnitID == unitID {
			uid := key.(string)
			c.deviceCache.Delete(uid)
			devicesCleared++
		}
		return true // 继续遍历
	})

	c.logger.Info("Cleared cache for unit",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Int("devices_cleared", devicesCleared),
	)
}

// ClearCacheForBranch 清除指定 branch 下所有设备的缓存
func (c *MQTTConsumer) ClearCacheForBranch(tenantID, branchID string) {
	if tenantID == "" || branchID == "" {
		return
	}

	// 遍历缓存，清除属于该 branch 的设备
	devicesCleared := 0
	c.deviceCache.Range(func(key, value interface{}) bool {
		deviceWithLocation := value.(*DeviceWithLocation)
		if deviceWithLocation.Device.TenantID == tenantID &&
			deviceWithLocation.LocationInfo != nil &&
			deviceWithLocation.LocationInfo.BranchID != nil &&
			*deviceWithLocation.LocationInfo.BranchID == branchID {
			uid := key.(string)
			c.deviceCache.Delete(uid)
			devicesCleared++
		}
		return true // 继续遍历
	})

	c.logger.Info("Cleared cache for branch",
		zap.String("tenant_id", tenantID),
		zap.String("branch_id", branchID),
		zap.Int("devices_cleared", devicesCleared),
	)
}

// ClearCacheForBuilding 清除指定 building 下所有设备的缓存
func (c *MQTTConsumer) ClearCacheForBuilding(tenantID, buildingID string) {
	if tenantID == "" || buildingID == "" {
		return
	}

	// 遍历缓存，清除属于该 building 的设备
	devicesCleared := 0
	c.deviceCache.Range(func(key, value interface{}) bool {
		deviceWithLocation := value.(*DeviceWithLocation)
		if deviceWithLocation.Device.TenantID == tenantID &&
			deviceWithLocation.LocationInfo != nil &&
			deviceWithLocation.LocationInfo.BuildingID != nil &&
			*deviceWithLocation.LocationInfo.BuildingID == buildingID {
			uid := key.(string)
			c.deviceCache.Delete(uid)
			devicesCleared++
		}
		return true // 继续遍历
	})

	c.logger.Info("Cleared cache for building",
		zap.String("tenant_id", tenantID),
		zap.String("building_id", buildingID),
		zap.Int("devices_cleared", devicesCleared),
	)
}

// FindUnitForRoomOrBed 根据 room_id 或 bed_id 查找对应的 unit_id
// 注意：这个方法需要查询数据库，应该谨慎使用
func (c *MQTTConsumer) FindUnitForRoomOrBed(tenantID, addressID, addressType string) (string, error) {
	if tenantID == "" || addressID == "" || addressType == "" {
		return "", fmt.Errorf("missing required parameters")
	}

	// 这里需要调用 deviceRepo 的方法来查询
	// 暂时返回空，实际实现需要查询数据库
	c.logger.Warn("FindUnitForRoomOrBed not implemented yet",
		zap.String("tenant_id", tenantID),
		zap.String("address_id", addressID),
		zap.String("address_type", addressType),
	)

	return "", nil
}

// Stop 停止消费者
func (c *MQTTConsumer) Stop(ctx context.Context) error {
	cfg := c.config.Radar.DeviceMQTT

	// 取消订阅所有主题
	topics := []string{
		mqtt.BuildPropertyPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildMonitorPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildFunctionPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildStatPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildEventPostTopic(cfg.Prefix, cfg.ProductID, "+"),
		mqtt.BuildAlarmPostTopic(cfg.Prefix, cfg.ProductID, "+"),
	}

	if c.config.Radar.Topics.Data != "" {
		topics = append(topics, c.config.Radar.Topics.Data)
	}

	if err := c.mqttClient.Unsubscribe(topics...); err != nil {
		c.logger.Error("Failed to unsubscribe", zap.Error(err))
	}

	c.logger.Info("MQTT consumer stopped")
	return nil
}

// handleMessage 处理MQTT消息（协议文档格式）
// 根据协议文档 3.3.3 节，处理设备发布的 6 种主题
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	c.logger.Debug("Received MQTT message",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)),
	)

	// 1. 解析主题，提取 prefix, type, productId, UID
	topicInfo, err := mqtt.ParseTopic(topic)
	if err != nil {
		c.logger.Error("Failed to parse topic",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to parse topic: %w", err)
	}

	// 2. 从主题中提取设备 UID
	uid := topicInfo.UID
	if uid == "+" {
		// 通配符主题，无法处理
		return fmt.Errorf("wildcard topic cannot be processed: %s", topic)
	}

	// 3. 解析消息
	var mqttData map[string]interface{}
	if err := json.Unmarshal(payload, &mqttData); err != nil {
		c.logger.Error("Failed to unmarshal MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 4. 获取设备信息（带位置信息）
	deviceWithLocation, isFirstConnection, err := c.getOrCreateDeviceWithLocation(uid, topic)
	if err != nil {
		return err
	}

	device := deviceWithLocation.Device
	locationInfo := deviceWithLocation.LocationInfo

	// 设备首次连接，自动订阅实时数据
	if isFirstConnection && c.config.Radar.Subscription.AutoSubscribe && c.subscriptionManager != nil {
		go func() {
			if err := c.subscriptionManager.AutoSubscribe(context.Background(), uid); err != nil {
				c.logger.Warn("Failed to auto-subscribe on device first connection",
					zap.String("uid", uid),
					zap.Error(err),
				)
			} else {
				c.logger.Info("Auto-subscribed on device first connection",
					zap.String("uid", uid),
				)
			}
		}()
	}

	// 5. 处理命令响应（如果是属性响应或功能响应，需要存储到 Redis 供 CommandService 使用）
	// 注意：prop 和 func 类型的消息是命令响应，属于同步的 request-response 模式
	// 这些响应不应该发布到 Streams，只需要存储到 Redis 供 CommandService 读取即可
	if topicInfo.Type == mqtt.TopicTypeProp {
		// 属性响应：检查是否有 requestId，如果有则存储到 Redis
		if requestID, ok := mqttData["requestId"].(string); ok && requestID != "" {
			responseKey := fmt.Sprintf("radar:response:%s", requestID)
			responseJSON, _ := json.Marshal(mqttData)
			if err := c.redisClient.Set(context.Background(), responseKey, responseJSON, 30*time.Second).Err(); err != nil {
				c.logger.Error("Failed to store property response in Redis",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Stored property response in Redis",
					zap.String("request_id", requestID),
					zap.String("uid", uid),
				)
			}
		}
		// 命令响应处理完毕，直接返回，不发布到 Streams
		return nil
	} else if topicInfo.Type == mqtt.TopicTypeFunc {
		// 功能响应：检查是否有 requestId，如果有则存储到 Redis
		if requestID, ok := mqttData["requestId"].(string); ok && requestID != "" {
			responseKey := fmt.Sprintf("radar:response:func:%s", requestID)
			responseJSON, _ := json.Marshal(mqttData)
			if err := c.redisClient.Set(context.Background(), responseKey, responseJSON, 30*time.Second).Err(); err != nil {
				c.logger.Error("Failed to store function response in Redis",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Stored function response in Redis",
					zap.String("request_id", requestID),
					zap.String("uid", uid),
				)
			}
		}
		// 命令响应处理完毕，直接返回，不发布到 Streams
		return nil
	}

	// 6. 处理数据上报消息（monitor, stat, event, alarm）
	// 这些消息需要编码后发布到 Redis Streams
	// 构建基础数据（直接展开原始数据，不保存在 raw_data 中）
	data := make(map[string]interface{})

	// 添加元数据（用于RadarDecoder，但不会出现在最终输出中）
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["device_uid"] = device.DeviceUID
	data["device_type"] = "Radar"
	data["topic_type"] = string(topicInfo.Type)
	data["timestamp"] = time.Now().Unix()
	data["topic"] = topic

	// 直接展开原始数据到顶层
	for k, v := range mqttData {
		data[k] = v
	}

	// 7. 调用 decode 公共函数进行解码转换
	dataValue, err := encode.RadarDecoder(data, string(topicInfo.Type))
	if err != nil {
		c.logger.Error("Failed to decode radar data",
			zap.String("device_id", device.DeviceID),
			zap.String("uid", uid),
			zap.String("topic_type", string(topicInfo.Type)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to decode radar data: %w", err)
	}

	// 8. 构建完整的输出对象（包含元数据 + data_value）
	topicTypeStr := string(topicInfo.Type)
	// 统一使用 "stat"，不再转换为 "statistics"

	// 9. 对于 event 和 stat 类型，检查是否应该发布为报警
	// 注意：monitor 类型不检查报警，直接发布到 monitor stream
	finalTopicType := topicTypeStr
	if topicInfo.Type == mqtt.TopicTypeEvent || topicInfo.Type == mqtt.TopicTypeStat {
		shouldPublishAsAlarm, possibleAlarmTypes, err := c.alarmHandler.ShouldPublishAsAlarm(
			context.Background(),
			device.TenantID,
			device.DeviceUID,
			topicTypeStr,
			dataValue,
		)
		if err != nil {
			c.logger.Warn("Failed to check alarm enablement, publishing as normal event/stat",
				zap.String("device_id", device.DeviceID),
				zap.String("topic_type", topicTypeStr),
				zap.Error(err),
			)
			// 查询失败，发布为普通 event/stat（避免误报）
		} else if shouldPublishAsAlarm {
			// 应该发布为报警，将 topic_type 改为 "alarm"
			finalTopicType = "alarm"
			c.logger.Debug("Publishing as alarm",
				zap.String("device_id", device.DeviceID),
				zap.Strings("possible_alarm_types", possibleAlarmTypes),
			)
		}
	}

	// 10. 处理多条数据分开发送
	// 根据标准文档 3.3：如果同时收到多条 track 或 vital 或数据，应该分开发送
	// 对于monitor和stat类型，必须将每个数据项单独发送
	streamName := c.getOutputStreamNameForTopicTypeStr(finalTopicType)

	// 将dataValue统一转换为数组格式，确保每条数据单独发送
	var itemsToSend []map[string]interface{}

	switch v := dataValue.(type) {
	case []interface{}:
		// dataValue 是数组，转换为map数组
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemsToSend = append(itemsToSend, itemMap)
			}
		}

	case []map[string]interface{}:
		// dataValue 是 map 数组，直接使用
		itemsToSend = v

	case map[string]interface{}:
		// dataValue 是单个对象，转换为数组
		itemsToSend = []map[string]interface{}{v}

	default:
		// 其他类型，尝试转换为map
		if itemMap, ok := dataValue.(map[string]interface{}); ok {
			itemsToSend = []map[string]interface{}{itemMap}
		} else {
			c.logger.Warn("Unknown dataValue type, cannot split into items",
				zap.String("uid", uid),
				zap.String("topic_type", finalTopicType),
			)
			return fmt.Errorf("unknown dataValue type: %T", dataValue)
		}
	}

	// 分开发送每个数据项
	for i, itemMap := range itemsToSend {
		// 提取当前对象的 category
		itemCategory := ""
		if cat, ok := itemMap["category"].(string); ok {
			itemCategory = cat
		}

		// 构建单个对象的 encodedData
		// buildEncodedData 内部会调用 cleanDataValue 清理 itemMap 中的不应该出现的字段
		encodedData := c.buildEncodedData(device, locationInfo, finalTopicType, itemCategory, itemMap)

		// 发布到 Redis Streams（使用该 stream 的配置）
		maxLen, retentionSeconds := c.config.GetStreamConfig(streamName)
		streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData, maxLen, retentionSeconds)
		if err != nil {
			c.logger.Error("Failed to publish item to Redis Streams",
				zap.String("stream", streamName),
				zap.Int("index", i),
				zap.Error(err),
			)
			continue
		}

		c.logger.Debug("Published radar data item to Redis Streams",
			zap.String("device_id", device.DeviceID),
			zap.String("uid", uid),
			zap.String("final_topic_type", finalTopicType),
			zap.String("category", itemCategory),
			zap.Int("index", i),
			zap.String("stream", streamName),
			zap.String("stream_id", streamID),
		)
	}

	c.logger.Info("Published radar data items to Redis Streams",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.String("original_topic_type", string(topicInfo.Type)),
		zap.String("final_topic_type", finalTopicType),
		zap.String("stream", streamName),
		zap.Int("items_count", len(itemsToSend)),
	)

	return nil
}

// handleLegacyMessage 处理旧格式的MQTT消息（兼容性）
// 主题格式: radar/{device_id}/data
func (c *MQTTConsumer) handleLegacyMessage(topic string, payload []byte) error {
	c.logger.Debug("Received legacy MQTT message",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)),
	)

	// 1. 从主题中提取设备标识符
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid legacy topic format: %s", topic)
	}
	deviceIdentifier := parts[1] // 可能是 serial_number 或 uid

	// 2. 解析消息
	var mqttData map[string]interface{}
	if err := json.Unmarshal(payload, &mqttData); err != nil {
		c.logger.Error("Failed to unmarshal legacy MQTT message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 3. 查询设备信息（如果不存在，尝试从 device_store 自动创建）
	device, err := c.deviceRepo.GetDeviceBySerialNumber(deviceIdentifier)
	if err != nil {
		// 尝试使用 UID 查询
		device, err = c.deviceRepo.GetDeviceByUID(deviceIdentifier)
		if err != nil {
			// 设备不存在，尝试从 device_store 自动创建
			device, err = c.deviceRepo.GetOrCreateDeviceFromStore(context.Background(), deviceIdentifier, topic)
			if err != nil {
				c.logger.Warn("Device not found and cannot be created from device_store",
					zap.String("identifier", deviceIdentifier),
					zap.String("mqtt_topic", topic),
					zap.Error(err),
				)
				return fmt.Errorf("device not found: %s", deviceIdentifier)
			}
			// 设备已从 device_store 自动创建
			c.logger.Info("Device auto-created from device_store on MQTT connection",
				zap.String("device_id", device.DeviceID),
				zap.String("identifier", deviceIdentifier),
				zap.String("mqtt_topic", topic),
			)
		}
	}

	// 4. 构建基础数据（直接展开原始数据）
	data := make(map[string]interface{})

	// 添加元数据（用于RadarDecoder，但不会出现在最终输出中）
	data["device_id"] = device.DeviceID
	data["tenant_id"] = device.TenantID
	data["device_uid"] = device.DeviceUID
	data["device_type"] = "Radar"
	data["topic_type"] = "legacy"
	data["timestamp"] = time.Now().Unix()
	data["topic"] = topic

	// 直接展开原始数据到顶层
	for k, v := range mqttData {
		data[k] = v
	}

	// 5. 调用 decode 公共函数进行解码转换
	dataValue, err := encode.RadarDecoder(data, "legacy")
	if err != nil {
		c.logger.Error("Failed to decode legacy radar data",
			zap.String("device_id", device.DeviceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to decode legacy radar data: %w", err)
	}

	// 6. 处理多条数据分开发送（legacy 消息也支持）
	streamName := "iot:data:stream"

	// 判断 dataValue 的类型
	switch v := dataValue.(type) {
	case []interface{}:
		// dataValue 是数组，需要分开发送
		for i, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// 提取当前对象的 category
				itemCategory := ""
				if cat, ok := itemMap["category"].(string); ok {
					itemCategory = cat
				}

				// 构建单个对象的 encodedData
				encodedData := map[string]interface{}{
					"device_uid":  getStringOrNull(device.DeviceUID),
					"device_type": "Radar",
					"tenant_id":   getStringOrNull(device.TenantID),
					"timestamp":   time.Now().Unix(),
					"topic_type":  "legacy",
					"category":    itemCategory,
					"data_value":  itemMap,
				}

				// 添加设备绑定信息字段（如果存在）
				if device.BoundBedID != nil && *device.BoundBedID != "" {
					encodedData["bed_id"] = *device.BoundBedID
				}
				if device.BoundRoomID != nil && *device.BoundRoomID != "" {
					encodedData["room_id"] = *device.BoundRoomID
				}

				// 发布到 Redis Streams
				maxLen, retentionSeconds := c.config.GetStreamConfig(streamName)
				streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData, maxLen, retentionSeconds)
				if err != nil {
					c.logger.Error("Failed to publish legacy array item to Redis Streams",
						zap.String("stream", streamName),
						zap.Int("index", i),
						zap.Error(err),
					)
					continue
				}

				c.logger.Debug("Published legacy radar data array item to Redis Streams",
					zap.String("device_id", device.DeviceID),
					zap.String("category", itemCategory),
					zap.Int("index", i),
					zap.String("stream", streamName),
					zap.String("stream_id", streamID),
				)
			}
		}

		c.logger.Info("Published legacy radar data array to Redis Streams",
			zap.String("device_id", device.DeviceID),
			zap.String("stream", streamName),
			zap.Int("items_count", len(v)),
		)

	case []map[string]interface{}:
		// dataValue 是 map 数组，需要分开发送
		for i, itemMap := range v {
			// 提取当前对象的 category
			itemCategory := ""
			if cat, ok := itemMap["category"].(string); ok {
				itemCategory = cat
			}

			// 构建单个对象的 encodedData
			encodedData := map[string]interface{}{
				"device_uid":  getStringOrNull(device.DeviceUID),
				"device_type": "Radar",
				"tenant_id":   getStringOrNull(device.TenantID),
				"timestamp":   time.Now().Unix(),
				"topic_type":  "legacy",
				"category":    itemCategory,
				"data_value":  itemMap,
			}

			// 添加设备绑定信息字段（如果存在）
			if device.BoundBedID != nil && *device.BoundBedID != "" {
				encodedData["bed_id"] = *device.BoundBedID
			}
			if device.BoundRoomID != nil && *device.BoundRoomID != "" {
				encodedData["room_id"] = *device.BoundRoomID
			}

			// 发布到 Redis Streams（使用配置的 MaxLen 限制 Stream 长度）
			maxLen, retentionSeconds := c.config.GetStreamConfig(streamName)
			streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData, maxLen, retentionSeconds)
			if err != nil {
				c.logger.Error("Failed to publish legacy map array item to Redis Streams",
					zap.String("stream", streamName),
					zap.Int("index", i),
					zap.Error(err),
				)
				continue
			}

			c.logger.Debug("Published legacy radar data map array item to Redis Streams",
				zap.String("device_id", device.DeviceID),
				zap.String("category", itemCategory),
				zap.Int("index", i),
				zap.String("stream", streamName),
				zap.String("stream_id", streamID),
			)
		}

		c.logger.Info("Published legacy radar data map array to Redis Streams",
			zap.String("device_id", device.DeviceID),
			zap.String("stream", streamName),
			zap.Int("items_count", len(v)),
		)

	default:
		// dataValue 是单个对象，直接发送
		// 提取 category 字段到顶层
		topLevelCategory := extractCategory(dataValue)

		encodedData := map[string]interface{}{
			"device_uid":  getStringOrNull(device.DeviceUID),
			"device_type": "Radar",
			"tenant_id":   getStringOrNull(device.TenantID),
			"timestamp":   time.Now().Unix(),
			"topic_type":  "legacy",
			"category":    topLevelCategory,
			"data_value":  dataValue,
		}

		// 添加设备绑定信息字段（如果存在）
		if device.BoundBedID != nil && *device.BoundBedID != "" {
			encodedData["bed_id"] = *device.BoundBedID
		}
		if device.BoundRoomID != nil && *device.BoundRoomID != "" {
			encodedData["room_id"] = *device.BoundRoomID
		}

		// 发布到 Redis Streams（使用该 stream 的配置）
		maxLen, retentionSeconds := c.config.GetStreamConfig(streamName)
		streamID, err := rediscommon.PublishJSONToStream(context.Background(), c.redisClient, streamName, encodedData, maxLen, retentionSeconds)
		if err != nil {
			c.logger.Error("Failed to publish legacy data to Redis Streams",
				zap.String("stream", streamName),
				zap.Error(err),
			)
			return fmt.Errorf("failed to publish to stream: %w", err)
		}

		c.logger.Info("Published legacy radar data to Redis Streams",
			zap.String("device_id", device.DeviceID),
			zap.String("stream", streamName),
			zap.String("stream_id", streamID),
		)
	}

	return nil
}

// getOutputStreamName 根据主题类型获取输出 Redis Stream 名称（使用 iot: 前缀）
func (c *MQTTConsumer) getOutputStreamName(topicType mqtt.TopicType) string {
	switch topicType {
	case mqtt.TopicTypeMonitor:
		return "iot:monitor:stream" // 实时数据
	case mqtt.TopicTypeStat:
		return "iot:stat:stream" // 统计数据
	case mqtt.TopicTypeEvent:
		return "iot:event:stream" // 事件/日志
	case mqtt.TopicTypeAlarm:
		return "iot:alarm:stream" // 告警
	default:
		return "iot:data:stream" // 默认
	}
}

// getOutputStreamNameForTopicTypeStr 根据 topic_type 字符串获取输出 Redis Stream 名称
// 用于处理经过报警判断后的最终 topic_type（可能是 "alarm"）
func (c *MQTTConsumer) getOutputStreamNameForTopicTypeStr(topicTypeStr string) string {
	switch topicTypeStr {
	case "monitor":
		return "iot:monitor:stream"
	case "stat":
		return "iot:stat:stream"
	case "event":
		return "iot:event:stream"
	case "alarm":
		return "iot:alarm:stream"
	default:
		return "iot:data:stream"
	}
}

// getStringOrNull 如果字符串为空，返回 nil，否则返回字符串
func getStringOrNull(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// getStringOrNullPtr 如果指针为 nil 或字符串为空，返回 nil，否则返回字符串
func getStringOrNullPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

// extractCategory 从 data_value 中提取 category 字段到顶层
func extractCategory(dataValue interface{}) string {
	switch v := dataValue.(type) {
	case map[string]interface{}:
		if category, ok := v["category"].(string); ok {
			return category
		}
	case []interface{}:
		// 如果是数组，取第一个元素的category
		if len(v) > 0 {
			if firstItem, ok := v[0].(map[string]interface{}); ok {
				if category, ok := firstItem["category"].(string); ok {
					return category
				}
			}
		}
	case []map[string]interface{}:
		// 如果是map数组，取第一个元素的category
		if len(v) > 0 {
			if category, ok := v[0]["category"].(string); ok {
				return category
			}
		}
	}
	return ""
}

// getStreamNameForTopicType 根据主题类型获取 Redis Stream 名称（保留用于兼容性，但不再使用）
// 注意：属性响应和功能响应不发布到 Streams，而是存储到 Redis 用于 request-response
func (c *MQTTConsumer) getStreamNameForTopicType(topicType mqtt.TopicType) string {
	switch topicType {
	case mqtt.TopicTypeProp:
		return "iot:prop:stream" // 属性响应（不发布到 Streams）
	case mqtt.TopicTypeMonitor:
		return "iot:monitor:stream" // 实时数据
	case mqtt.TopicTypeFunc:
		return "iot:func:stream" // 功能响应（不发布到 Streams）
	case mqtt.TopicTypeStat:
		return "iot:stat:stream" // 统计数据
	case mqtt.TopicTypeEvent:
		return "iot:event:stream" // 事件/日志
	case mqtt.TopicTypeAlarm:
		return "iot:alarm:stream" // 告警
	default:
		return "iot:data:stream" // 默认
	}
}

// DeviceWithLocation 带位置信息的设备结构
type DeviceWithLocation struct {
	Device       *repository.Device
	LocationInfo *repository.DeviceLocationInfo
	LastUpdated  time.Time
}

// getOrCreateDeviceWithLocation 获取或创建设备（带位置信息）
// 只使用device_uid，不需要device_id, serial, uid
// 注意：locationInfo 包含 name 字段和硬件信息字段，但不应传递到 dataValue 中
func (c *MQTTConsumer) getOrCreateDeviceWithLocation(uid, topic string) (*DeviceWithLocation, bool, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.deviceCache.Load(uid); ok {
		deviceWithLocation := cached.(*DeviceWithLocation)
		return deviceWithLocation, false, nil
	}

	// 2. 缓存未命中，查询数据库（只使用device_uid）
	device, err := c.deviceRepo.GetDeviceByUID(uid)
	isFirstConnection := false
	if err != nil {
		// 设备不存在，尝试从 device_store 自动创建（只使用device_uid）
		device, err = c.deviceRepo.GetOrCreateDeviceFromStore(context.Background(), uid, topic)
		if err != nil {
			c.logger.Warn("Device not found and cannot be created from device_store",
				zap.String("uid", uid),
				zap.String("mqtt_topic", topic),
				zap.Error(err),
			)
			return nil, false, fmt.Errorf("device not found: %s", uid)
		}
		// 设备已从 device_store 自动创建
		c.logger.Info("Device auto-created from device_store on MQTT connection",
			zap.String("device_id", device.DeviceID),
			zap.String("uid", uid),
			zap.String("mqtt_topic", topic),
		)
		isFirstConnection = true
	} else {
		// 设备已存在，检查是否是首次收到消息（通过检查订阅状态）
		// 如果设备存在但没有订阅记录，也视为首次连接
		subscriptionKey := fmt.Sprintf("radar:subscription:%s", uid)
		exists, err := c.redisClient.Exists(context.Background(), subscriptionKey).Result()
		if err == nil && exists == 0 {
			isFirstConnection = true
		}
	}

	// 3. 查询设备位置信息（包括address信息）
	// 注意：locationInfo 包含 name 字段和硬件信息字段，但不应传递到 dataValue 中
	locationInfo, err := c.deviceRepo.GetDeviceLocationInfoByIdentifier(context.Background(), uid)
	if err != nil {
		c.logger.Warn("Failed to get device location info",
			zap.String("uid", uid),
			zap.Error(err),
		)
		// 创建空的 LocationInfo，避免空指针
		locationInfo = &repository.DeviceLocationInfo{}
	}

	// 4. 创建带位置信息的设备对象
	deviceWithLocation := &DeviceWithLocation{
		Device:       device,
		LocationInfo: locationInfo,
		LastUpdated:  time.Now(),
	}

	// 5. 存入缓存
	c.deviceCache.Store(uid, deviceWithLocation)

	c.logger.Debug("Cached device with location info",
		zap.String("uid", uid),
		zap.String("device_id", device.DeviceID),
		zap.Bool("has_location", locationInfo != nil),
	)

	return deviceWithLocation, isFirstConnection, nil
}

// buildEncodedData 构建包含完整位置信息的输出数据
// 符合 RADAR_REDIS_STREAM_FORMAT_STANDARD.md 标准格式
// 字段顺序：device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
// dataValue: 来自 RadarDecoder 的返回值，类型为 interface{}（可能是 map[string]interface{} 或 []map[string]interface{}）
// 注意：dataValue 可能包含元数据字段（device_id, comm_mode, device_model等），需要清理
func (c *MQTTConsumer) buildEncodedData(
	device *repository.Device,
	locationInfo *repository.DeviceLocationInfo,
	topicType, category string,
	dataValue interface{},
) map[string]interface{} {
	// 如果category为空，根据topic_type设置默认值
	finalCategory := category
	if finalCategory == "" {
		switch topicType {
		case "monitor":
			finalCategory = "track"
		case "stat":
			finalCategory = "track" // 默认使用track，如果data_value有sleep则会被覆盖
		case "event":
			finalCategory = "other"
		case "alarm":
			finalCategory = "alarm"
		default:
			finalCategory = "unknown"
		}
	}

	// 清理dataValue（itemMap），移除不应该出现的字段
	// 注意：itemMap 可能包含元数据字段（device_id, comm_mode, device_model等），需要清理
	cleanedDataValue := c.cleanDataValue(dataValue)

	// 构建标准格式数据，严格按照字段顺序
	// 字段顺序：device_id → device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
	// 使用有序map确保字段顺序（Go的map在JSON序列化时可能无序，但按顺序添加可以保证大部分情况下有序）
	encodedData := make(map[string]interface{}, 13)

	// 必需字段（按标准顺序添加）
	encodedData["device_id"] = getStringOrNull(device.DeviceID)
	encodedData["device_uid"] = getStringOrNull(device.DeviceUID)

	// device_type：优先使用locationInfo中的，否则使用默认值
	deviceType := "Radar"
	if locationInfo != nil && locationInfo.DeviceType != nil {
		deviceType = *locationInfo.DeviceType
	}
	encodedData["device_type"] = deviceType

	encodedData["tenant_id"] = getStringOrNull(device.TenantID)
	encodedData["timestamp"] = time.Now().Unix()
	encodedData["topic_type"] = topicType
	encodedData["category"] = finalCategory
	encodedData["data_value"] = cleanedDataValue

	// 位置信息字段（按标准顺序：branch_id → building_id → unit_id → room_id → bed_id）
	// 只包含ID字段，不包含name字段和硬件信息字段
	if locationInfo != nil {
		encodedData["branch_id"] = locationInfo.BranchID
		encodedData["building_id"] = locationInfo.BuildingID
		encodedData["unit_id"] = locationInfo.UnitID
		encodedData["room_id"] = locationInfo.RoomID
		encodedData["bed_id"] = locationInfo.BedID
	} else {
		// 如果位置信息为空，使用设备绑定信息
		encodedData["branch_id"] = nil
		encodedData["building_id"] = nil
		encodedData["unit_id"] = nil
		encodedData["room_id"] = getStringOrNullPtr(device.BoundRoomID)
		encodedData["bed_id"] = getStringOrNullPtr(device.BoundBedID)
	}

	return encodedData
}

// cleanDataValue 清理dataValue，移除不应该出现的字段
// 移除：device_id, device_serial, device_uid, tenant_id, timestamp, topic_type, topic等元数据字段
// 移除：comm_mode, device_model, firmware_version, mcu_model等硬件信息字段
func (c *MQTTConsumer) cleanDataValue(dataValue interface{}) interface{} {
	// 需要移除的字段列表
	excludedFields := map[string]bool{
		// 元数据字段
		"device_id":   true,
		"device_uid":  true,
		"tenant_id":   true,
		"timestamp":   true,
		"topic_type":  true,
		"topic":       true,
		"device_type": true,
		// 硬件信息字段
		"comm_mode":        true,
		"device_model":     true,
		"firmware_version": true,
		"mcu_model":        true,
		"imei":             true,
	}

	switch v := dataValue.(type) {
	case map[string]interface{}:
		cleaned := make(map[string]interface{})
		for k, val := range v {
			if !excludedFields[k] {
				cleaned[k] = val
			}
		}
		return cleaned

	case []interface{}:
		cleaned := make([]interface{}, 0, len(v))
		for _, item := range v {
			cleaned = append(cleaned, c.cleanDataValue(item))
		}
		return cleaned

	case []map[string]interface{}:
		cleaned := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if cleanedItem := c.cleanDataValue(item); cleanedItem != nil {
				if cleanedMap, ok := cleanedItem.(map[string]interface{}); ok {
					cleaned = append(cleaned, cleanedMap)
				}
			}
		}
		return cleaned

	default:
		return dataValue
	}
}
