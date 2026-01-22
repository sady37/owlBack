package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/decode"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
)

// DeviceLastSeenUpdater 设备最后收到消息时间更新器接口
type DeviceLastSeenUpdater interface {
	UpdateLastSeen(deviceUID string)
}

// MQTTConsumer MQTT消费者
type MQTTConsumer struct {
	config              *config.Config
	mqttClient          *mqtt.Client
	redisClient         *redis.Client
	deviceRepo          repository.DeviceRepository
	streamPublisher     *StreamPublisher
	subscriptionManager DeviceLastSeenUpdater // 设备最后收到消息时间更新器接口
	subscribedTopics    map[string]struct{}   // 保存已订阅的设备主题（key: topic, value: struct{}）
	mu                  sync.RWMutex
}

// GetMessageHandler 获取消息处理器（用于传递给subscriptionManager）
func (c *MQTTConsumer) GetMessageHandler() func(topic string, payload []byte) error {
	return c.handleMessage
}

// SetSubscriptionManager 设置订阅管理器（用于UpdateLastSeen）
func (c *MQTTConsumer) SetSubscriptionManager(manager DeviceLastSeenUpdater) {
	c.subscriptionManager = manager
}

// NewMQTTConsumer 创建MQTT消费者
func NewMQTTConsumer(
	cfg *config.Config,
	mqttClient *mqtt.Client,
	redisClient *redis.Client,
	deviceRepo repository.DeviceRepository,
	streamPublisher *StreamPublisher,
	subscriptionManager DeviceLastSeenUpdater,
) (*MQTTConsumer, error) {
	return &MQTTConsumer{
		config:              cfg,
		mqttClient:          mqttClient,
		redisClient:         redisClient,
		deviceRepo:          deviceRepo,
		streamPublisher:     streamPublisher,
		subscriptionManager: subscriptionManager,
		subscribedTopics:    make(map[string]struct{}),
	}, nil
}

// Start 启动消费者
// 改为不订阅任何主题，等待设备认证成功后动态订阅
func (c *MQTTConsumer) Start(ctx context.Context) error {
	log.Println("Starting MQTT consumer (no wildcard subscriptions)...")

	// 启动连接监控goroutine，检测重连后重新订阅已认证设备
	go c.monitorConnection(ctx)

	log.Println("MQTT consumer started (waiting for device authentication to subscribe)")

	return nil
}

// monitorConnection 监控MQTT连接状态，重连后重新订阅
func (c *MQTTConsumer) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var wasConnected bool
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			isConnected := c.mqttClient.IsConnected()

			// 检测从断开到重连的状态变化
			if !wasConnected && isConnected {
				log.Println("MQTT client reconnected, resubscribing to topics...")
				c.resubscribeTopics()
			}

			wasConnected = isConnected
		}
	}
}

// resubscribeTopics 重新订阅所有已订阅的主题（用于重连后）
func (c *MQTTConsumer) resubscribeTopics() {
	c.mu.RLock()
	topics := make([]string, 0, len(c.subscribedTopics))
	for topic := range c.subscribedTopics {
		topics = append(topics, topic)
	}
	c.mu.RUnlock()

	for _, topic := range topics {
		if err := c.mqttClient.Subscribe(topic, c.handleMessage); err != nil {
			log.Printf("Failed to resubscribe to topic %s: %v", topic, err)
		} else {
			log.Printf("✅ Resubscribed to device topic: %s", topic)
		}
	}
	log.Printf("Resubscribed to %d device topics", len(topics))
}

// SubscribeDeviceTopics 订阅指定设备的6个主题
func (c *MQTTConsumer) SubscribeDeviceTopics(deviceUID string) error {
	log.Printf("🔍 SubscribeDeviceTopics called for device %s", deviceUID)

	// 检查 MQTT 连接状态
	if !c.mqttClient.IsConnected() {
		log.Printf("❌ MQTT client is not connected, cannot subscribe to device topics for %s", deviceUID)
		return fmt.Errorf("MQTT client is not connected, cannot subscribe to device topics for %s", deviceUID)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := c.config.MQTT.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	// 构建设备的6个主题
	topics := []string{
		buildDeviceTopic(prefix, productID, "prop", deviceUID),
		buildDeviceTopic(prefix, productID, "monitor", deviceUID),
		buildDeviceTopic(prefix, productID, "func", deviceUID),
		buildDeviceTopic(prefix, productID, "stat", deviceUID),
		buildDeviceTopic(prefix, productID, "event", deviceUID),
		buildDeviceTopic(prefix, productID, "alarm", deviceUID),
	}

	log.Printf("📋 Attempting to subscribe to %d topics for device %s", len(topics), deviceUID)

	subscribedCount := 0
	skippedCount := 0

	// 订阅所有主题
	for _, topic := range topics {
		if _, alreadySubscribed := c.subscribedTopics[topic]; alreadySubscribed {
			log.Printf("⏭️  Topic %s already subscribed, skipping", topic)
			skippedCount++
			continue
		}

		if err := c.mqttClient.Subscribe(topic, c.handleMessage); err != nil {
			log.Printf("❌ Failed to subscribe to topic %s: %v", topic, err)
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}

		c.subscribedTopics[topic] = struct{}{}
		subscribedCount++
		log.Printf("✅ Subscribed to device topic: %s", topic)
	}

	log.Printf("📊 Subscribed to %d topics (skipped %d already subscribed) for device %s", subscribedCount, skippedCount, deviceUID)
	return nil
}

// UnsubscribeDeviceTopics 取消订阅指定设备的6个主题
func (c *MQTTConsumer) UnsubscribeDeviceTopics(deviceUID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := c.config.MQTT.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	// 构建设备的6个主题
	topics := []string{
		buildDeviceTopic(prefix, productID, "prop", deviceUID),
		buildDeviceTopic(prefix, productID, "monitor", deviceUID),
		buildDeviceTopic(prefix, productID, "func", deviceUID),
		buildDeviceTopic(prefix, productID, "stat", deviceUID),
		buildDeviceTopic(prefix, productID, "event", deviceUID),
		buildDeviceTopic(prefix, productID, "alarm", deviceUID),
	}

	// 取消订阅所有主题
	for _, topic := range topics {
		if _, subscribed := c.subscribedTopics[topic]; !subscribed {
			continue
		}

		if err := c.mqttClient.Unsubscribe(topic); err != nil {
			log.Printf("Failed to unsubscribe from topic %s: %v", topic, err)
			// 继续尝试取消订阅其他主题
			continue
		}

		delete(c.subscribedTopics, topic)
		log.Printf("✅ Unsubscribed from device topic: %s", topic)
	}

	log.Printf("Unsubscribed from %d topics for device %s", len(topics), deviceUID)
	return nil
}

// buildDeviceTopic 构建设备特定主题
func buildDeviceTopic(prefix, productID, topicType, deviceUID string) string {
	if prefix == "" {
		return fmt.Sprintf("/%s/%s/%s/post", topicType, productID, deviceUID)
	}
	return fmt.Sprintf("/%s/%s/%s/%s/post", prefix, topicType, productID, deviceUID)
}

// buildWildcardTopic 构建通配符主题（保留，可能用于其他用途）
func buildWildcardTopic(prefix, productID, topicType string) string {
	if prefix == "" {
		return fmt.Sprintf("/%s/%s/+/post", topicType, productID)
	}
	return fmt.Sprintf("/%s/%s/%s/+/post", prefix, topicType, productID)
}

// Stop 停止消费者
func (c *MQTTConsumer) Stop(ctx context.Context) error {
	log.Println("Stopping MQTT consumer...")
	// 这里应该实现取消订阅的逻辑
	return nil
}

// buildTopics 构建订阅主题列表
func (c *MQTTConsumer) buildTopics() []string {
	cfg := c.config.MQTT.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	topics := []string{}

	// 构建4个主题（使用通配符+匹配所有设备）
	// prop/func 暂不订阅
	types := []string{"monitor", "stat", "event", "alarm"}
	for _, topicType := range types {
		var topic string
		if prefix == "" {
			topic = fmt.Sprintf("/%s/%s/+/post", topicType, productID)
		} else {
			topic = fmt.Sprintf("/%s/%s/%s/+/post", prefix, topicType, productID)
		}
		topics = append(topics, topic)
	}

	return topics
}

// handleMessage 处理MQTT消息
// 注意：现在只订阅已认证设备的主题，未认证设备无法发送消息到服务端
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	log.Printf("Received MQTT message on topic: %s", topic)

	// 解析主题，提取设备UID
	uid, err := c.extractUIDFromTopic(topic)
	if err != nil {
		log.Printf("Failed to extract UID from topic %s: %v", topic, err)
		return nil // 不返回错误，继续处理其他消息
	}

	// 检查设备是否已认证（allow_access=TRUE）
	// 这是安全机制：只有已认证的设备才能处理消息，防止未认证设备攻击
	ctx := context.Background()
	deviceStoreInfo, _, err := c.deviceRepo.GetDeviceStoreInfoAndLocation(ctx, uid)
	if err != nil {
		log.Printf("Device %s not found in device_store, skipping message (device not authenticated)", uid)
		return nil
	}
	if deviceStoreInfo == nil || !deviceStoreInfo.AllowAccess {
		log.Printf("Device %s not authenticated (allow_access=FALSE), skipping message", uid)
		return nil
	}

	// 更新设备最后收到消息的时间
	// 参考wisefido-radar的实现：UpdateLastSeen会检测首次连接并自动发送monitor订阅命令
	if c.subscriptionManager != nil {
		c.subscriptionManager.UpdateLastSeen(uid)
	}

	// 解析消息体
	var message map[string]interface{}
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("Failed to parse MQTT message: %v", err)
		return nil
	}

	// 根据主题类型处理消息
	topicType := c.extractTopicType(topic)
	switch topicType {
	case "prop":
		return c.handlePropertyMessage(uid, message)
	case "monitor":
		return c.handleMonitorMessage(uid, message)
	case "func":
		return c.handleFunctionMessage(uid, message)
	case "stat":
		return c.handleStatMessage(uid, message)
	case "event":
		return c.handleEventMessage(uid, message)
	case "alarm":
		return c.handleAlarmMessage(uid, message)
	default:
		log.Printf("Unknown topic type: %s", topicType)
		return nil
	}
}

// extractUIDFromTopic 从主题中提取设备UID
func (c *MQTTConsumer) extractUIDFromTopic(topic string) (string, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid topic format: %s", topic)
	}

	// 主题格式: /{prefix}/{type}/{productID}/{uid}/post
	// 或者: /{type}/{productID}/{uid}/post
	uidIndex := len(parts) - 2 // UID在倒数第二个位置
	return parts[uidIndex], nil
}

// extractTopicType 从主题中提取主题类型
func (c *MQTTConsumer) extractTopicType(topic string) string {
	parts := strings.Split(topic, "/")

	// 主题格式: /{prefix}/{type}/{productID}/{uid}/post
	// 或者: /{type}/{productID}/{uid}/post
	if len(parts) >= 4 {
		// 如果有前缀，类型在索引2，否则在索引1
		typeIndex := 1
		if c.config.MQTT.RadarDeviceMQTT.Prefix != "" {
			typeIndex = 2
		}
		if typeIndex < len(parts) {
			return parts[typeIndex]
		}
	}

	return "unknown"
}

// getDeviceFromCache 从 Auth 缓存中获取设备信息（包含位置信息）
// 如果缓存中没有，则查询数据库（降级处理）
// 使用 GetDeviceStoreInfoAndLocation 一次性获取设备和位置信息，与 Auth 流程保持一致
func (c *MQTTConsumer) getDeviceFromCache(ctx context.Context, uid string) (*domain.Device, *domain.DeviceLocationInfo) {
	// 1. 尝试从 Auth 缓存获取（使用 domain.DeviceCache）
	if cached, ok := domain.DeviceCache.Load(uid); ok {
		if deviceWithLocation, ok := cached.(*domain.DeviceWithLocation); ok {
			// 如果缓存中有设备信息，直接返回
			if deviceWithLocation.Device != nil {
				return deviceWithLocation.Device, deviceWithLocation.LocationInfo
			}
			// 如果只有位置信息，尝试从位置信息构建设备对象
			if deviceWithLocation.LocationInfo != nil {
				device := buildDeviceFromLocationInfo(deviceWithLocation.LocationInfo)
				if device != nil {
					// 更新缓存
					deviceWithLocation.Device = device
					domain.DeviceCache.Store(uid, deviceWithLocation)
					return device, deviceWithLocation.LocationInfo
				}
				// 即使构建设备失败，也返回位置信息
				return nil, deviceWithLocation.LocationInfo
			}
		}
	}

	// 2. 缓存中没有，使用 GetDeviceStoreInfoAndLocation 一次性查询（与 Auth 流程一致）
	// 这个方法一次性获取 device_store 和 devices 表的信息，包括位置信息
	// 注意：device_id 始终来自 device_store，即使设备还没有在 devices 表中创建记录
	deviceStoreInfo, locationInfo, err := c.deviceRepo.GetDeviceStoreInfoAndLocation(ctx, uid)
	if err != nil {
		// 输出错误日志，便于诊断问题
		log.Printf("Failed to get device info for %s from database: %v", uid, err)
		return nil, nil
	}

	// 3. 从位置信息构建设备对象（DeviceLocationInfo 包含了 domain.Device 的所有字段）
	// 注意：即使 devices 表中没有记录，locationInfo 也会包含 device_store 的 device_id
	var device *domain.Device
	if locationInfo != nil {
		device = buildDeviceFromLocationInfo(locationInfo)
		if device == nil {
			log.Printf("Failed to build device from locationInfo for %s (DeviceID: %s)", uid, locationInfo.DeviceID)
		}
	} else if deviceStoreInfo != nil {
		// 如果 locationInfo 为 nil（不应该发生，因为我们已经修复了 GetDeviceStoreInfoAndLocation），
		// 至少使用 device_store 信息构建设备对象
		log.Printf("locationInfo is nil but deviceStoreInfo exists for %s, building from deviceStoreInfo", uid)
		device = &domain.Device{
			DeviceID:   deviceStoreInfo.DeviceID,
			DeviceUID:  deviceStoreInfo.DeviceUID,
			TenantID:   deviceStoreInfo.TenantID,
			DeviceType: sql.NullString{String: deviceStoreInfo.DeviceType, Valid: true},
		}
		locationInfo = &domain.DeviceLocationInfo{
			DeviceID:   deviceStoreInfo.DeviceID,
			DeviceUID:  deviceStoreInfo.DeviceUID,
			TenantID:   deviceStoreInfo.TenantID,
			DeviceType: sql.NullString{String: deviceStoreInfo.DeviceType, Valid: true},
		}
	} else {
		log.Printf("Both locationInfo and deviceStoreInfo are nil for %s", uid)
	}

	// 4. 更新缓存
	if locationInfo != nil {
		deviceWithLocation := &domain.DeviceWithLocation{
			Device:       device,
			LocationInfo: locationInfo,
		}
		domain.DeviceCache.Store(uid, deviceWithLocation)
	}

	return device, locationInfo
}

// buildDeviceFromLocationInfo 从 DeviceLocationInfo 构建 domain.Device
// DeviceLocationInfo 已经包含了 domain.Device 的所有字段
func buildDeviceFromLocationInfo(locationInfo *domain.DeviceLocationInfo) *domain.Device {
	if locationInfo == nil {
		log.Printf("buildDeviceFromLocationInfo: locationInfo is nil")
		return nil
	}
	if locationInfo.DeviceID == "" {
		log.Printf("buildDeviceFromLocationInfo: DeviceID is empty for device_uid: %s", locationInfo.DeviceUID)
		return nil
	}

	return &domain.Device{
		DeviceID:          locationInfo.DeviceID,
		DeviceUID:         locationInfo.DeviceUID,
		TenantID:          locationInfo.TenantID,
		DeviceName:        locationInfo.DeviceName,
		BoundRoomID:       locationInfo.BoundRoomID,
		BoundBedID:        locationInfo.BoundBedID,
		Status:            locationInfo.Status,
		BusinessAccess:    locationInfo.BusinessAccess,
		MonitoringEnabled: locationInfo.MonitoringEnabled,
		DeviceType:        locationInfo.DeviceType,
		DeviceModel:       locationInfo.DeviceModel,
		IMEI:              locationInfo.IMEI,
		CommMode:          locationInfo.CommMode,
		MCUModel:          locationInfo.MCUModel,
		FirmwareVersion:   locationInfo.FirmwareVersion,
	}
}

// publishDecodedData 发布解码后的数据到 Redis Stream
// 一条 base64 消息解码后是一个对象，不是数组
// - monitor/stat: 如果返回数组（多个数据项），则分开发送每个数据项
// - event/alarm: 一条消息只包含一个事件，直接发送单个对象（不是数组）
func (c *MQTTConsumer) publishDecodedData(
	ctx context.Context,
	device *domain.Device,
	locationInfo *domain.DeviceLocationInfo,
	topicType string,
	category string,
	dataValue interface{},
	originalMessage map[string]interface{},
) error {
	streamName := c.streamPublisher.GetOutputStreamName(topicType)

	// event 和 alarm 类型：一条消息只包含一个事件，直接发送单个对象
	if topicType == "event" || topicType == "alarm" {
		// 将 dataValue 转换为单个对象
		var eventObj map[string]interface{}
		switch v := dataValue.(type) {
		case map[string]interface{}:
			// dataValue 是单个对象（正常情况）
			eventObj = v
		case []map[string]interface{}:
			// 兼容处理：如果是数组，只取第一个（实际不应该出现）
			if len(v) > 0 {
				eventObj = v[0]
			}
		case []interface{}:
			// 兼容处理：如果是数组，只取第一个（实际不应该出现）
			if len(v) > 0 {
				if itemMap, ok := v[0].(map[string]interface{}); ok {
					eventObj = itemMap
				}
			}
		default:
			// 其他类型，降级处理：使用原始消息
			log.Printf("Unexpected dataValue type for %s message: %T, using original message", topicType, dataValue)
			encodedData := c.streamPublisher.BuildEncodedData(device, locationInfo, topicType, category, originalMessage)
			streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
			if err != nil {
				log.Printf("Failed to publish %s data to stream: %v", topicType, err)
				return err
			}
			// 输出 stream 发布日志（stat, event, auth）
			if topicType == "stat" || topicType == "event" || topicType == "auth" {
				log.Printf("Published %s data to stream %s (stream_id: %s) for device %s", topicType, streamName, streamID, device.DeviceUID)
			}
			return nil
		}

		// 如果没有事件对象，跳过
		if len(eventObj) == 0 {
			log.Printf("No event data to publish for %s message", topicType)
			return nil
		}

		// 提取事件自己的 category（enter2out, pose, number-people 等）
		// 如果没有 category，使用传入的默认 category
		eventCategory := category
		if cat, ok := eventObj["category"].(string); ok && cat != "" {
			eventCategory = cat
		}

		// 直接发送单个对象
		encodedData := c.streamPublisher.BuildEncodedData(device, locationInfo, topicType, eventCategory, eventObj)
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data to stream: %v", topicType, err)
			return err
		}

		// 输出 stream 发布日志（stat, event, auth）
		if topicType == "stat" || topicType == "event" || topicType == "auth" {
			log.Printf("Published %s data to stream %s (stream_id: %s) for device %s", topicType, streamName, streamID, device.DeviceUID)
		} else {
			log.Printf("Published %s data (single event) for device %s to stream %s", topicType, device.DeviceUID, streamName)
		}
		return nil
	}

	// monitor 和 stat 类型：分开发送每个数据项
	var itemsToSend []map[string]interface{}

	switch v := dataValue.(type) {
	case []interface{}:
		// dataValue 是数组，转换为 map 数组
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemsToSend = append(itemsToSend, itemMap)
			}
		}
	case []map[string]interface{}:
		// dataValue 是 map 数组，直接使用
		itemsToSend = v
	case map[string]interface{}:
		// dataValue 是单个对象（一条 base64 消息解码后是一个对象）
		itemsToSend = []map[string]interface{}{v}
	default:
		// 其他类型，降级处理：使用原始消息
		log.Printf("Unexpected dataValue type for %s message: %T, using original message", topicType, dataValue)
		encodedData := c.streamPublisher.BuildEncodedData(device, locationInfo, topicType, category, originalMessage)
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data to stream: %v", topicType, err)
			return err
		}
		// 输出 stream 发布日志（stat, event, auth）
		if topicType == "stat" || topicType == "event" || topicType == "auth" {
			log.Printf("Published %s data to stream %s (stream_id: %s) for device %s", topicType, streamName, streamID, device.DeviceUID)
		}
		return nil
	}

	// 如果没有数据项，跳过
	if len(itemsToSend) == 0 {
		log.Printf("No data items to publish for %s message", topicType)
		return nil
	}

	// 分开发送每个数据项（例如 monitor 可能包含 track 和 vital）
	for i, itemMap := range itemsToSend {
		// 提取数据项自己的 category（track, vital, sleep 等）
		// 如果没有 category，使用传入的默认 category
		itemCategory := category
		if cat, ok := itemMap["category"].(string); ok && cat != "" {
			itemCategory = cat
		}

		// 构建单个对象的 encodedData
		// 注意：itemMap 中的 category 字段会被保留在 data_value 中
		encodedData := c.streamPublisher.BuildEncodedData(device, locationInfo, topicType, itemCategory, itemMap)

		// 发布到 Redis Stream
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data item %d to stream: %v", topicType, i, err)
			continue
		}

		// 输出 stream 发布日志（stat, event, auth）
		if topicType == "stat" || topicType == "event" || topicType == "auth" {
			log.Printf("Published %s data item %d to stream %s (stream_id: %s) for device %s", topicType, i+1, streamName, streamID, device.DeviceUID)
		}
	}

	// 输出 stream 发布日志（stat, event, auth）
	if topicType == "stat" || topicType == "event" || topicType == "auth" {
		log.Printf("Published %d %s data items to stream %s for device %s", len(itemsToSend), topicType, streamName, device.DeviceUID)
	} else {
		log.Printf("Published %d %s data items for device %s to stream %s", len(itemsToSend), topicType, device.DeviceUID, streamName)
	}
	return nil
}

// handlePropertyMessage 处理属性响应消息
func (c *MQTTConsumer) handlePropertyMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling property message for device %s", uid)

	// 这里可以添加属性消息的处理逻辑
	// 例如：更新设备属性缓存，触发相关事件等

	return nil
}

// handleMonitorMessage 处理实时数据消息
func (c *MQTTConsumer) handleMonitorMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling monitor message for device %s", uid)

	// 从 Auth 缓存获取设备信息（包含位置信息）
	ctx := context.Background()
	device, locationInfo := c.getDeviceFromCache(ctx, uid)
	if device == nil {
		// 如果 device 为 nil，但 locationInfo 可能不为 nil（包含 device_store 的 device_id）
		// 尝试使用 locationInfo 构建设备对象
		if locationInfo != nil && locationInfo.DeviceID != "" {
			log.Printf("Device object is nil but locationInfo exists, building device from locationInfo for %s", uid)
			device = buildDeviceFromLocationInfo(locationInfo)
		}

		// 如果仍然为 nil，说明设备不在 device_store 中
		if device == nil {
			log.Printf("Device %s not found in device_store, skipping monitor message", uid)
			return nil
		}
	}

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "monitor")
	if err != nil {
		log.Printf("Failed to decode monitor data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 发布解码后的数据
	// 注意：publishDecodedData 会使用数据项自己的 category（track, vital）作为顶层 category
	return c.publishDecodedData(ctx, device, locationInfo, "monitor", "", dataValue, message)
}

// handleFunctionMessage 处理功能响应消息
func (c *MQTTConsumer) handleFunctionMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling function message for device %s", uid)

	// 这里可以添加功能响应消息的处理逻辑
	// 例如：更新命令执行状态，触发回调等

	return nil
}

// handleStatMessage 处理统计数据消息
func (c *MQTTConsumer) handleStatMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling stat message for device %s", uid)

	// 从 Auth 缓存获取设备信息（包含位置信息）
	ctx := context.Background()
	device, locationInfo := c.getDeviceFromCache(ctx, uid)
	if device == nil {
		// 如果 device 为 nil，但 locationInfo 可能不为 nil（包含 device_store 的 device_id）
		// 尝试使用 locationInfo 构建设备对象
		if locationInfo != nil && locationInfo.DeviceID != "" {
			log.Printf("Device object is nil but locationInfo exists, building device from locationInfo for %s (DeviceID: %s)", uid, locationInfo.DeviceID)
			device = buildDeviceFromLocationInfo(locationInfo)
		}

		// 如果仍然为 nil，说明设备不在 device_store 中
		// 安全机制：未认证的设备不能接收消息，防止攻击
		if device == nil {
			log.Printf("Device %s not found in device_store, skipping stat message (device not authenticated). Check if device exists in device_store table.", uid)
			return nil
		}
	}

	log.Printf("Device %s found: DeviceID=%s, TenantID=%s, DeviceType=%s", uid, device.DeviceID, device.TenantID, device.DeviceType.String)

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "stat")
	if err != nil {
		log.Printf("Failed to decode stat data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 获取报警使能配置（如果设备已认证且有tenant_id）
	var enablementItems []alarm.AlarmEnablementItem
	if device.TenantID != "" && device.DeviceUID != "" {
		var err error
		enablementItems, err = c.deviceRepo.GetAlarmEnablement(ctx, device.TenantID, device.DeviceUID)
		if err != nil {
			log.Printf("Failed to get alarm enablement for device %s: %v", uid, err)
			// 如果获取失败，继续发布到 stat stream（降级处理）
			return c.publishDecodedData(ctx, device, locationInfo, "stat", "", dataValue, message)
		}
	} else {
		// 设备未认证，没有tenant_id，直接发布到stat stream
		log.Printf("Device %s not authenticated yet (no tenant_id), publishing stat message directly", uid)
		return c.publishDecodedData(ctx, device, locationInfo, "stat", "", dataValue, message)
	}

	// stat 消息可能包含多个数据项（track, sleep 等），需要逐个检查
	// 如果任一数据项应该转换为 alarm，则整个消息转换为 alarm
	var itemsToCheck []map[string]interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemsToCheck = append(itemsToCheck, itemMap)
			}
		}
	case []map[string]interface{}:
		itemsToCheck = v
	case map[string]interface{}:
		itemsToCheck = []map[string]interface{}{v}
	default:
		// 其他类型，直接发布到 stat stream
		return c.publishDecodedData(ctx, device, locationInfo, "stat", "", dataValue, message)
	}

	// 收集所有 numeric codes，然后使用 GetAlarmEnabledMapByNumericCodes 检查
	var allNumericCodes []string
	for _, itemMap := range itemsToCheck {
		// 优先使用 decoder 已生成的 stat_numeric_codes
		// 注意：decoder生成的是 []string，但在 map[string]interface{} 中可能是 []string 或 []interface{}
		if numericCodes, ok := itemMap["stat_numeric_codes"].([]string); ok {
			// 直接是 []string 类型（decoder直接设置）
			allNumericCodes = append(allNumericCodes, numericCodes...)
		} else if numericCodes, ok := itemMap["stat_numeric_codes"].([]interface{}); ok {
			// []interface{} 类型（JSON序列化/反序列化后）
			for _, code := range numericCodes {
				if codeStr, ok := code.(string); ok {
					allNumericCodes = append(allNumericCodes, codeStr)
				}
			}
		} else {
			// 如果没有 stat_numeric_codes，使用 ExtractNumericCodesFromStat 提取
			codes := alarm.ExtractNumericCodesFromStat(itemMap)
			allNumericCodes = append(allNumericCodes, codes...)
		}
	}

	// 使用 GetAlarmEnabledMapByNumericCodes 直接获取使能表
	shouldConvertToAlarm := false
	if len(allNumericCodes) > 0 && len(enablementItems) > 0 {
		enabledMap := alarm.GetAlarmEnabledMapByNumericCodes(allNumericCodes, enablementItems)
		// 如果任一数字组合对应的报警启用，则转换为 alarm
		for _, enabled := range enabledMap {
			if enabled == 1 {
				shouldConvertToAlarm = true
				break
			}
		}
	}

	if shouldConvertToAlarm {
		// 转换为 alarm，发布到 alarm stream
		log.Printf("Stat message for device %s should be converted to alarm", uid)
		return c.publishDecodedData(ctx, device, locationInfo, "alarm", "", dataValue, message)
	}

	// 保持原 topic，发布到 stat stream
	// 注意：publishDecodedData 会使用数据项自己的 category（track, sleep）作为顶层 category
	return c.publishDecodedData(ctx, device, locationInfo, "stat", "", dataValue, message)
}

// handleEventMessage 处理事件消息
func (c *MQTTConsumer) handleEventMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling event message for device %s", uid)

	// 从 Auth 缓存获取设备信息（包含位置信息）
	ctx := context.Background()
	device, locationInfo := c.getDeviceFromCache(ctx, uid)
	if device == nil {
		// 如果 device 为 nil，但 locationInfo 可能不为 nil（包含 device_store 的 device_id）
		// 尝试使用 locationInfo 构建设备对象
		if locationInfo != nil && locationInfo.DeviceID != "" {
			log.Printf("Device object is nil but locationInfo exists, building device from locationInfo for %s", uid)
			device = buildDeviceFromLocationInfo(locationInfo)
		}

		// 如果仍然为 nil，说明设备不在 device_store 中
		// 安全机制：未认证的设备不能接收消息，防止攻击
		if device == nil {
			log.Printf("Device %s not found in device_store, skipping event message (device not authenticated). Check if device exists in device_store table.", uid)
			return nil
		}
	}

	log.Printf("Device %s found: DeviceID=%s, TenantID=%s, DeviceType=%s", uid, device.DeviceID, device.TenantID, device.DeviceType.String)

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "event")
	if err != nil {
		log.Printf("Failed to decode event data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 检查是否应该转换为 alarm
	// dataValue 应该是单个对象（map[string]interface{}）
	var eventObj map[string]interface{}
	switch v := dataValue.(type) {
	case map[string]interface{}:
		eventObj = v
	case []interface{}:
		if len(v) > 0 {
			if itemMap, ok := v[0].(map[string]interface{}); ok {
				eventObj = itemMap
			}
		}
	default:
		eventObj = make(map[string]interface{})
	}

	// 获取报警使能配置
	enablementItems, err := c.deviceRepo.GetAlarmEnablement(ctx, device.TenantID, device.DeviceUID)
	if err != nil {
		log.Printf("Failed to get alarm enablement for device %s: %v", uid, err)
		// 如果获取失败，继续发布到 event stream（降级处理）
		return c.publishDecodedData(ctx, device, locationInfo, "event", "event", dataValue, message)
	}

	// 判断是否应该转换为 alarm
	// 直接使用 ExtractNumericCodesFromEvent 和 GetAlarmEnabledMapByNumericCodes
	shouldConvertToAlarm := false
	if len(eventObj) > 0 && len(enablementItems) > 0 {
		numericCodes := alarm.ExtractNumericCodesFromEvent(eventObj)
		if len(numericCodes) > 0 {
			enabledMap := alarm.GetAlarmEnabledMapByNumericCodes(numericCodes, enablementItems)
			// 如果任一数字组合对应的报警启用，则转换为 alarm
			for _, enabled := range enabledMap {
				if enabled == 1 {
					shouldConvertToAlarm = true
					break
				}
			}
		}
	}

	if shouldConvertToAlarm {
		// 转换为 alarm，发布到 alarm stream
		log.Printf("Event message for device %s should be converted to alarm", uid)
		return c.publishDecodedData(ctx, device, locationInfo, "alarm", "event", dataValue, message)
	}

	// 保持原 topic，发布到 event stream
	return c.publishDecodedData(ctx, device, locationInfo, "event", "event", dataValue, message)
}

// handleAlarmMessage 处理告警消息
func (c *MQTTConsumer) handleAlarmMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling alarm message for device %s", uid)

	// 从 Auth 缓存获取设备信息（包含位置信息）
	ctx := context.Background()
	device, locationInfo := c.getDeviceFromCache(ctx, uid)
	if device == nil {
		// 如果 device 为 nil，但 locationInfo 可能不为 nil（包含 device_store 的 device_id）
		// 尝试使用 locationInfo 构建设备对象
		if locationInfo != nil && locationInfo.DeviceID != "" {
			log.Printf("Device object is nil but locationInfo exists, building device from locationInfo for %s", uid)
			device = buildDeviceFromLocationInfo(locationInfo)
		}

		// 如果仍然为 nil，说明设备不在 device_store 中
		// 安全机制：未认证的设备不能接收消息，防止攻击
		if device == nil {
			log.Printf("Device %s not found in device_store, skipping alarm message (device not authenticated). Check if device exists in device_store table.", uid)
			return nil
		}
	}

	log.Printf("Device %s found: DeviceID=%s, TenantID=%s, DeviceType=%s", uid, device.DeviceID, device.TenantID, device.DeviceType.String)

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "alarm")
	if err != nil {
		log.Printf("Failed to decode alarm data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 发布解码后的数据
	// 注意：publishDecodedData 会使用数据项自己的 category（enter2out, pose 等）作为顶层 category
	return c.publishDecodedData(ctx, device, locationInfo, "alarm", "", dataValue, message)
}
