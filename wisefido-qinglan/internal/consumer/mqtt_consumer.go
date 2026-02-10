package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/radar"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/decode"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
)

const (
	// DeviceTypeRadar 设备类型常量（wisefido-qinglan网关的所有设备都是Radar）
	DeviceTypeRadar = "Radar"
)

// CardIDProvider 定义获取设备卡片映射信息的接口
type CardIDProvider interface {
	GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*repository.CardDeviceInfo, error)
}

// DeviceLastSeenUpdater 设备最后收到消息时间更新器接口
type DeviceLastSeenUpdater interface {
	UpdateLastSeen(deviceUID string)
	UpdateLastSeenByType(deviceUID, topicType string)
	// PublishOnlineForConnectedDevices 对已在列表且在线的设备发布上线通知（由 mqtt 启动 1 分钟后检查队列后调用）
	PublishOnlineForConnectedDevices(ctx context.Context, deviceUIDs []string)
}

// MQTTConsumer MQTT消费者
type MQTTConsumer struct {
	config              *config.Config
	mqttClient          *mqtt.Client
	redisClient         *redis.Client
	deviceRepo          repository.DeviceRepository
	cardMappingService  CardIDProvider // CardIDProvider用于获取deviceUID对应的cardID
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
	cardMappingService CardIDProvider,
	streamPublisher *StreamPublisher,
	subscriptionManager DeviceLastSeenUpdater,
) (*MQTTConsumer, error) {
	return &MQTTConsumer{
		config:              cfg,
		mqttClient:          mqttClient,
		redisClient:         redisClient,
		deviceRepo:          deviceRepo,
		cardMappingService:  cardMappingService,
		streamPublisher:     streamPublisher,
		subscriptionManager: subscriptionManager,
		subscribedTopics:    make(map[string]struct{}),
	}, nil
}

// Start 启动消费者
// 启动时主动订阅所有符合条件的设备（allow_access=TRUE 且 business_access='approved'）
func (c *MQTTConsumer) Start(ctx context.Context) error {
	log.Println("Starting MQTT consumer...")

	// 启动连接监控goroutine，检测重连后重新订阅已认证设备
	go c.monitorConnection(ctx)

	// 启动时主动订阅所有符合条件的设备
	go c.subscribeAllAccessibleDevices(ctx)

	// 启动 1 分钟后检查 MQTT 队列（subscribedTopics），检查所有 func 主题对应的设备，对接入的发布上线通知
	go c.publishOnlineForConnectedAfterStartup(ctx)

	log.Println("MQTT consumer started")

	return nil
}

// publishOnlineForConnectedAfterStartup 启动 1 分钟后，从已订阅主题（MQTT 队列）中取所有 func 主题对应的设备 UID，通知订阅管理器对已接入的发布上线
func (c *MQTTConsumer) publishOnlineForConnectedAfterStartup(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(1 * time.Minute):
	}
	c.mu.RLock()
	topics := make([]string, 0, len(c.subscribedTopics))
	for topic := range c.subscribedTopics {
		topics = append(topics, topic)
	}
	c.mu.RUnlock()
	uidSet := make(map[string]struct{})
	for _, topic := range topics {
		if c.extractTopicType(topic) != "func" {
			continue
		}
		uid, err := c.extractUIDFromTopic(topic)
		if err != nil {
			continue
		}
		uidSet[uid] = struct{}{}
	}
	uids := make([]string, 0, len(uidSet))
	for uid := range uidSet {
		uids = append(uids, uid)
	}
	if len(uids) == 0 {
		return
	}
	if c.subscriptionManager != nil {
		c.subscriptionManager.PublishOnlineForConnectedDevices(ctx, uids)
	}
}

// subscribeAllAccessibleDevices 订阅所有可访问的设备
func (c *MQTTConsumer) subscribeAllAccessibleDevices(ctx context.Context) {
	// 等待MQTT连接建立
	maxWaitTime := 30 * time.Second
	waitInterval := 500 * time.Millisecond
	var waited time.Duration

	for !c.mqttClient.IsConnected() && waited < maxWaitTime {
		time.Sleep(waitInterval)
		waited += waitInterval
	}

	if !c.mqttClient.IsConnected() {
		log.Printf("⚠️ MQTT client not connected after %v, skipping initial device subscription", maxWaitTime)
		return
	}

	// 查询所有可访问的设备
	deviceUIDs, err := c.deviceRepo.GetAllAccessibleDevices(ctx)
	if err != nil {
		log.Printf("❌ Failed to get accessible devices: %v", err)
		return
	}

	if len(deviceUIDs) == 0 {
		log.Println("No accessible devices found, skipping subscription")
		return
	}

	// 打印被允许的设备列表（不打印每条订阅信息）；启动即监听所有允许设备，1 分钟内有 MQTT 则置 online，超时则 offline 并取消订阅
	log.Printf("Allowed devices (%d): %v", len(deviceUIDs), deviceUIDs)

	successCount := 0
	for _, deviceUID := range deviceUIDs {
		if err := c.SubscribeDeviceTopics(deviceUID); err != nil {
			log.Printf("❌ Failed to subscribe to device %s: %v", deviceUID, err)
		} else {
			successCount++
		}
	}

	log.Printf("Subscribed to topics for %d/%d allowed devices", successCount, len(deviceUIDs))
}

// monitorConnection 监控MQTT连接状态，重连后重新订阅
func (c *MQTTConsumer) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 初始化为当前连接状态，避免启动后第一次 tick 误判为“重连”而重复 resubscribe
	wasConnected := c.mqttClient.IsConnected()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			isConnected := c.mqttClient.IsConnected()

			// 仅当从断开变为已连接时才 resubscribe（真实重连）
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
		}
	}
	log.Printf("Resubscribed to %d device topics (reconnected)", len(topics))
}

// SubscribeDeviceTopics 订阅指定设备的 6 个主题（不打印每条订阅，由调用方打印允许设备列表）
func (c *MQTTConsumer) SubscribeDeviceTopics(deviceUID string) error {
	if !c.mqttClient.IsConnected() {
		log.Printf("❌ MQTT client not connected, cannot subscribe for %s", deviceUID)
		return fmt.Errorf("MQTT client is not connected, cannot subscribe to device topics for %s", deviceUID)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := c.config.MQTT.RadarDeviceMQTT
	prefix := cfg.Prefix
	productID := cfg.ProductID

	topics := []string{
		buildDeviceTopic(prefix, productID, "prop", deviceUID),
		buildDeviceTopic(prefix, productID, "monitor", deviceUID),
		buildDeviceTopic(prefix, productID, "func", deviceUID),
		buildDeviceTopic(prefix, productID, "stat", deviceUID),
		buildDeviceTopic(prefix, productID, "event", deviceUID),
		buildDeviceTopic(prefix, productID, "alarm", deviceUID),
	}

	for _, topic := range topics {
		if _, alreadySubscribed := c.subscribedTopics[topic]; alreadySubscribed {
			continue
		}
		if err := c.mqttClient.Subscribe(topic, c.handleMessage); err != nil {
			log.Printf("❌ Failed to subscribe to topic %s: %v", topic, err)
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
		c.subscribedTopics[topic] = struct{}{}
	}
	return nil
}

// IsDeviceTopicsSubscribed 检查设备的6个主题是否都已订阅
func (c *MQTTConsumer) IsDeviceTopicsSubscribed(deviceUID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

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

	// 检查所有主题是否都已订阅
	for _, topic := range topics {
		if _, subscribed := c.subscribedTopics[topic]; !subscribed {
			return false
		}
	}

	return true
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

// Radar MQTT 主题约定（与本 consumer 的关系）：
//
//   - /prefix/.../post：设备上报，本 consumer 订阅并处理。handleMessage 按 type(prop/monitor/func/stat/event/alarm) 分发。
//   - /prefix/prop/.../get、/prefix/func/.../get：我们下发命令的主题，本 consumer 不订阅、不接收。
//     谁在发：internal/service/radar_service.go — GetDeviceProperties/SetDeviceProperties 发布到 prop/get，
//     CallDeviceFunction 发布到 func/get；设备收到后回复到 .../post，由本 consumer 的 handlePropertyMessage、handleFunctionMessage 处理并存 Redis。
//
// handleMessage 处理 MQTT 消息（仅 .../post）
// 注意：现在只订阅已认证设备的主题，未认证设备无法发送消息到服务端
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	// log.Printf("Received MQTT message on topic: %s", topic) // 已关闭，减少刷屏

	// 解析主题，提取设备UID
	uid, err := c.extractUIDFromTopic(topic)
	if err != nil {
		log.Printf("Failed to extract UID from topic %s: %v", topic, err)
		return nil // 不返回错误，继续处理其他消息
	}

	// 快速安全滤清：仅从缓存检查 allow_access（mqtt中只查cache）
	// 缓存由 auth_service 维护，Device的认证结果写入缓存
	cached, ok := domain.AllowAccessCache.Load(uid)
	if !ok {
		// 缓存未命中：回退到数据库查询 device_store（避免因为短期缓存缺失导致拒绝）
		ds, err := c.deviceRepo.GetDeviceStoreInfoAndLocation(context.Background(), uid)
		if err != nil {
			log.Printf("Device %s not in cache, rejecting message (device not authenticated): db lookup failed: %v", uid, err)
			// 为避免频繁DB命中，短期内缓存为 false
			domain.AllowAccessCache.Store(uid, false)
			return nil
		}
		if !ds.AllowAccess {
			log.Printf("Device %s blocked by device_store: allow_access=FALSE", uid)
			domain.AllowAccessCache.Store(uid, false)
			return nil
		}
		// 设备在 device_store 中允许，写入缓存并继续处理
		domain.AllowAccessCache.Store(uid, true)
	} else {
		// 检查缓存值
		if allowedBool, ok := cached.(bool); !ok || !allowedBool {
			// 类型转换失败 或 缓存中明确记录该设备不被授权
			log.Printf("Device %s blocked by cache: allow_access=FALSE", uid)
			return nil
		}
	}
	// 缓存中存在且为 true，设备已认证，继续处理

	// 解析消息体
	var message map[string]interface{}
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("Failed to parse MQTT message: %v", err)
		return nil
	}

	// 根据主题类型处理消息
	topicType := c.extractTopicType(topic)

	// 更新设备最后收到消息的时间（根据消息类型）
	// 参考wisefido-radar的实现：UpdateLastSeenByType会检测首次连接并自动发送monitor订阅命令
	if c.subscriptionManager != nil {
		// 只更新 monitor/stat/event/alarm 类型的消息时间戳
		if topicType == "monitor" || topicType == "stat" || topicType == "event" || topicType == "alarm" {
			c.subscriptionManager.UpdateLastSeenByType(uid, topicType)
		}
	}
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

// publishDecodedData 发布解码后的数据到 Redis Stream
// 一条 base64 消息解码后是一个对象，不是数组
// - monitor/stat: 如果返回数组（多个数据项），则分开发送每个数据项
// - event/alarm: 一条消息只包含一个事件，直接发送单个对象（不是数组）
// 注意：cardID为空时仍继续处理（无接收端），但输出警告日志
func (c *MQTTConsumer) publishDecodedData(
	ctx context.Context,
	cardID string,
	tenantID string,
	deviceID string,
	topicType string,
	category string,
	dataValue interface{},
	originalMessage map[string]interface{},
) error {
	// cardID为空时输出警告日志，但仍继续处理
	if cardID == "" {
		log.Printf("⚠️ Device has no cardID, but still publishing %s message. deviceID=%s, tenantID=%s", topicType, deviceID, tenantID)
	}

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
			encodedData := c.streamPublisher.BuildEncodedData(cardID, tenantID, deviceID, topicType, category, []interface{}{originalMessage})
			streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
			if err != nil {
				log.Printf("Failed to publish %s data to stream: %v", topicType, err)
				return err
			}
			// 输出 stream 发布日志（auth, alarm, event）
			if topicType == "auth" || topicType == "alarm" || topicType == "event" {
				log.Printf("Published %s data to stream %s (stream_id: %s) for cardID %s", topicType, streamName, streamID, cardID)
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
		encodedData := c.streamPublisher.BuildEncodedData(cardID, tenantID, deviceID, topicType, eventCategory, []interface{}{eventObj})
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data to stream: %v", topicType, err)
			return err
		}

		// 输出 stream 发布日志（auth, alarm, event）
		if topicType == "auth" || topicType == "alarm" || topicType == "event" {
			log.Printf("Published %s data to stream %s (stream_id: %s) for cardID %s", topicType, streamName, streamID, cardID)
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
		encodedData := c.streamPublisher.BuildEncodedData(cardID, tenantID, deviceID, topicType, category, []interface{}{originalMessage})
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data to stream: %v", topicType, err)
			return err
		}
		// 输出 stream 发布日志（auth, alarm, event）
		if topicType == "auth" || topicType == "alarm" || topicType == "event" || topicType == "monitor" {
			log.Printf("Published %s data to stream %s (stream_id: %s) for cardID %s", topicType, streamName, streamID, cardID)
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

		// 构建单个对象的 encodedData（data_value 为单元素数组）
		encodedData := c.streamPublisher.BuildEncodedData(cardID, tenantID, deviceID, topicType, itemCategory, []interface{}{itemMap})

		// 发布到 Redis Stream
		streamID, err := c.streamPublisher.PublishToStream(ctx, streamName, encodedData)
		if err != nil {
			log.Printf("Failed to publish %s data item %d to stream: %v", topicType, i, err)
			continue
		}

		// 输出 stream 发布日志（auth, alarm, event）
		if topicType == "auth" || topicType == "alarm" || topicType == "event" {
			log.Printf("Published %s data item %d to stream %s (stream_id: %s) for deviceID %s", topicType, i+1, streamName, streamID, deviceID)
		}
	}

	// 输出 stream 发布日志（auth, alarm, event）
	if topicType == "auth" || topicType == "alarm" || topicType == "event" {
		log.Printf("Published %d %s data items to stream %s for deviceID %s", len(itemsToSend), topicType, streamName, deviceID)
	}
	return nil
}

// handlePropertyMessage 处理属性响应消息（/prop/.../post）：读/写回包共用，根据 cmd 区分日志
func (c *MQTTConsumer) handlePropertyMessage(uid string, message map[string]interface{}) error {
	requestIDRaw := ""
	if id, ok := message["requestId"].(string); ok && id != "" {
		requestIDRaw = id
	} else if id, ok := message["request_id"].(string); ok && id != "" {
		requestIDRaw = id
	} else if id, ok := message["requestID"].(string); ok && id != "" {
		requestIDRaw = id
	}
	cmd, _ := message["cmd"].(string)
	logLabel := "SetDeviceProperties"
	if cmd == "read" {
		logLabel = "GetDeviceProperties"
	}
	log.Printf("%s receive MQTT (device raw): device=%s, requestId=%s, msg=%+v", logLabel, uid, requestIDRaw, message)

	decoded, err := decode.RadarDecoder(message, "prop")
	if err != nil || decoded == nil {
		decoded = message
	}
	var payload map[string]interface{}
	if resp, ok := decoded.(*radar.PropResponse); ok {
		payload = map[string]interface{}{"request_id": resp.RequestID, "data": resp.Data}
	} else if m, ok := decoded.(map[string]interface{}); ok {
		payload = m
	}
	if payload == nil {
		payload = message
	}
	// 透传设备 code/msg，否则 RadarService.waitForResponse 从 Redis 取出的 response 无 code，会被当成 0 误报失败
	if _, has := payload["code"]; !has && message["code"] != nil {
		payload["code"] = message["code"]
	}
	if _, has := payload["msg"]; !has && message["msg"] != nil {
		payload["msg"] = message["msg"]
	}

	requestID := ""
	if id, ok := payload["request_id"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := payload["requestId"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := payload["requestID"].(string); ok && id != "" {
		requestID = id
	}

	if requestID != "" {
		ctx := context.Background()
		if err := c.streamPublisher.StoreCommandResponse(ctx, requestID, payload); err != nil {
			log.Printf("❌ %s store Redis: requestId=%s: %v", logLabel, requestID, err)
		} else {
			log.Printf("✅ %s send Redis: device=%s, requestId=%s, payload=%+v", logLabel, uid, requestID, payload)
		}
	} else {
		log.Printf("⚠️ %s: no requestId, device=%s", logLabel, uid)
	}

	return nil
}

// getMapKeys 获取map的所有键（用于调试）
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// handleMonitorMessage 处理实时数据消息
// 方案 B：一条 MQTT decode 一次，发一条 stream。data_value 为数组，一条一组 [{category,...},...]；
// sleep 归于 vital；顶层 category 为 trackN.vitalN（仅一种时为 trackN 或 vitalN），N 为该类条数。
func (c *MQTTConsumer) handleMonitorMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()

	// 使用 CardMappingService 获取 cardID 和其他设备映射信息
	cardDeviceInfo, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid)
	if err != nil {
		log.Printf("Failed to get cardID for device %s: %v", uid, err)
		// cardID 获取失败，不能处理该设备消息
		return nil
	}

	// cardDeviceInfo 不能为 nil，否则无法处理
	if cardDeviceInfo == nil {
		log.Printf("Device %s not found in card mapping, skipping monitor message", uid)
		return nil
	}

	var cardID, tenantID, deviceID string
	cardID = cardDeviceInfo.CardID
	tenantID = cardDeviceInfo.TenantID
	deviceID = cardDeviceInfo.DeviceID

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "monitor")
	if err != nil {
		// log.Printf("Failed to decode monitor data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	var items []map[string]interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, m)
			}
		}
	case []map[string]interface{}:
		items = v
	case map[string]interface{}:
		items = []map[string]interface{}{v}
	default:
		return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "monitor", "", []interface{}{message}, message)
	}

	if len(items) == 0 {
		return nil
	}

	nTrack, nVital := 0, 0
	for _, m := range items {
		cat, _ := m["category"].(string)
		if cat == "sleep" {
			cat = "vital"
			m["category"] = "vital"
		}
		switch cat {
		case "track":
			nTrack++
		case "vital":
			nVital++
		}
	}

	var topCategory string
	if nTrack > 0 && nVital > 0 {
		topCategory = fmt.Sprintf("track%d.vital%d", nTrack, nVital)
	} else if nTrack > 0 {
		topCategory = fmt.Sprintf("track%d", nTrack)
	} else if nVital > 0 {
		topCategory = fmt.Sprintf("vital%d", nVital)
	}

	// data_value：数组，每项自带 category
	dvSlice := make([]interface{}, len(items))
	for i, m := range items {
		dvSlice[i] = m
	}
	return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "monitor", topCategory, dvSlice, message)
}

// handleFunctionMessage 处理功能响应消息（/func/.../post）
// 实时交互通过 HTTP API 对外提供：客户端调 POST /api/v1/radar/devices/{uid}/function 时，
// RadarService 发 MQTT 到 /func/.../get 并轮询 Redis；设备在 /func/.../post 回包后，此处提取
// requestId 并存 Redis，RadarService 方能取到响应并返回给 HTTP 调用方。
func (c *MQTTConsumer) handleFunctionMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling function message for device %s", uid)

	// 提取 requestId（支持多种字段名）
	var requestID string
	if id, ok := message["requestId"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := message["request_id"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := message["requestID"].(string); ok && id != "" {
		requestID = id
	}

	// 有 requestId 即命令响应：存 Redis 供 HTTP API（RadarService.waitForResponse）获取后返回客户端
	if requestID != "" {
		ctx := context.Background()
		if err := c.streamPublisher.StoreCommandResponse(ctx, requestID, message); err != nil {
			log.Printf("Failed to store function response for requestId %s: %v", requestID, err)
			// 不返回错误，继续处理
		} else {
			log.Printf("Stored function response for requestId %s from device %s", requestID, uid)
		}
	} else {
		// 没有 requestId，可能是设备主动上报的功能状态
		log.Printf("Function message from device %s has no requestId, treating as function status notification", uid)
		// 这里可以添加功能状态更新处理逻辑
	}

	return nil
}

// handleStatMessage 处理统计数据消息
func (c *MQTTConsumer) handleStatMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()

	// 使用 CardMappingService 获取 cardID 和其他设备映射信息
	cardDeviceInfo, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid)
	if err != nil {
		log.Printf("Failed to get cardID for device %s: %v", uid, err)
		// cardID 获取失败，不能处理该设备消息
		return nil
	}

	// cardDeviceInfo 不能为 nil，否则无法处理
	if cardDeviceInfo == nil {
		log.Printf("Device %s not found in card mapping, skipping stat message (device not authenticated)", uid)
		return nil
	}

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "stat")
	if err != nil {
		log.Printf("Failed to decode stat data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 获取报警使能配置（如果设备已认证且有tenant_id）
	var enablementItems []alarm.AlarmEnablementItem
	if cardDeviceInfo.TenantID != "" && cardDeviceInfo.DeviceUID != "" {
		var err error
		enablementItems, err = c.deviceRepo.GetAlarmEnablement(ctx, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceUID)
		if err != nil {
			log.Printf("Failed to get alarm enablement for device %s: %v", uid, err)
			// 如果获取失败，继续发布到 stat stream（降级处理）
			return c.publishDecodedData(ctx, cardDeviceInfo.CardID, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceID, "stat", "", dataValue, message)
		}
	} else {
		// 设备未认证，没有tenant_id，直接发布到stat stream
		log.Printf("Device %s not authenticated yet (no tenant_id), publishing stat message directly", uid)
		return c.publishDecodedData(ctx, cardDeviceInfo.CardID, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceID, "stat", "", dataValue, message)
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
		return c.publishDecodedData(ctx, cardDeviceInfo.CardID, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceID, "stat", "", dataValue, message)
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
	var alarmLevel, alarmType string
	if len(allNumericCodes) > 0 && len(enablementItems) > 0 {
		enabledMap := alarm.GetAlarmEnabledMapByNumericCodes(allNumericCodes, enablementItems)
		// 如果任一数字组合对应的报警启用，则转换为 alarm
		for numCode, enabled := range enabledMap {
			if enabled == 1 {
				shouldConvertToAlarm = true
				// 从 enablementItems 中查找对应的 AlarmType 和 AlarmLevel
				for _, item := range enablementItems {
					if item.IsEnabled == 1 {
						// 简单匹配：使用第一个启用的报警项的 level 和 type
						alarmLevel = item.AlarmLevel
						alarmType = item.AlarmType
						_ = numCode // 使用 numCode 避免 unused 警告
						break
					}
				}
				break
			}
		}
	}

	if shouldConvertToAlarm {
		// 转换为 alarm，发布到 alarm stream
		alarmCategory := fmt.Sprintf("%s.%s", alarmLevel, alarmType)
		log.Printf("Stat message for device %s should be converted to alarm (category=%s)", uid, alarmCategory)
		return c.publishDecodedData(ctx, cardDeviceInfo.CardID, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceID, "alarm", alarmCategory, dataValue, message)
	}

	// 常规 stat 流，直接发布
	return c.publishDecodedData(ctx, cardDeviceInfo.CardID, cardDeviceInfo.TenantID, cardDeviceInfo.DeviceID, "stat", "", dataValue, message)
}

// handleEventMessage 处理事件消息
func (c *MQTTConsumer) handleEventMessage(uid string, message map[string]interface{}) error {

	ctx := context.Background()

	// 使用 CardMappingService 获取 cardID 和其他设备映射信息
	cardDeviceInfo, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid)
	if err != nil {
		log.Printf("Failed to get cardID for device %s: %v", uid, err)
		// cardID 获取失败，不能处理该设备消息
		return nil
	}

	// cardDeviceInfo 不能为 nil，否则无法处理
	if cardDeviceInfo == nil {
		log.Printf("Device %s not found in card mapping, skipping event message (device not authenticated)", uid)
		return nil
	}

	var cardID, tenantID, deviceID string
	cardID = cardDeviceInfo.CardID
	tenantID = cardDeviceInfo.TenantID
	deviceID = cardDeviceInfo.DeviceID

	// 调试：记录原始消息结构
	//log.Printf("[EVENT_RAW_DEBUG] device=%s raw message keys: %v", uid, getMapKeys(message))
	log.Printf("[EVENT_RAW_DEBUG] device=%s raw message: %+v", uid, message)

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "event")
	if err != nil {
		log.Printf("[EVENT_DECODE_ERROR] Failed to decode event data for device %s: %v", uid, err)
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

	// 调试：记录解码后的 eventObj 结构
	if len(eventObj) > 0 {
		//log.Printf("[EVENT_DEBUG] decoded eventObj keys: %v", getMapKeys(eventObj))
		log.Printf("[EVENT_DEBUG] decoded eventObj: event_type=%v, pose=%v, pose_raw=%v, category=%v",
			eventObj["event_type"], eventObj["pose"], eventObj["pose_raw"], eventObj["category"])
	}

	// 获取报警使能配置
	enablementItems, err := c.deviceRepo.GetAlarmEnablement(ctx, tenantID, uid)
	if err != nil {
		log.Printf("Failed to get alarm enablement for device %s: %v", uid, err)
		// 如果获取失败，继续发布到 event stream（降级处理）
		return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "event", "event", dataValue, message)
	}

	// 判断是否应该转换为 alarm
	// 直接使用 ExtractNumericCodesFromEvent 和 GetAlarmEnabledMapByNumericCodes
	shouldConvertToAlarm := false
	var numericCodes []string
	var enabledMap map[string]int
	var alarmLevel, alarmType string
	if len(eventObj) > 0 && len(enablementItems) > 0 {
		numericCodes = alarm.ExtractNumericCodesFromEvent(eventObj)
		if len(numericCodes) > 0 {
			enabledMap = alarm.GetAlarmEnabledMapByNumericCodes(numericCodes, enablementItems)
			// 如果任一数字组合对应的报警启用，则转换为 alarm
			for numCode, enabled := range enabledMap {
				if enabled == 1 {
					shouldConvertToAlarm = true
					// 从 enablementItems 中查找对应的 AlarmType 和 AlarmLevel
					for _, item := range enablementItems {
						if item.IsEnabled == 1 {
							// 简单匹配：使用第一个启用的报警项的 level 和 type
							alarmLevel = item.AlarmLevel
							alarmType = item.AlarmType
							_ = numCode // 使用 numCode 避免 unused 警告
							break
						}
					}
					break
				}
			}
		}
	}

	// 记录详细的 event 处理日志
	eventType, _ := eventObj["event_type"].(string)
	pose, _ := eventObj["pose"]
	areaType, _ := eventObj["area_type"]
	log.Printf("[EVENT_HANDLER] device=%s event_type=%s pose=%v area_type=%v numeric_codes=%v enabled_map=%v should_convert=%v enablement_count=%d",
		uid, eventType, pose, areaType, numericCodes, enabledMap, shouldConvertToAlarm, len(enablementItems))

	if shouldConvertToAlarm {
		// 转换为 alarm，发布到 alarm stream
		// category 格式："AlarmLevel.AlarmType"（例如 "EMERG.Fall"）
		alarmCategory := fmt.Sprintf("%s.%s", alarmLevel, alarmType)
		log.Printf("[EVENT_TO_ALARM] device=%s event_type=%s pose=%v area_type=%v converted to alarm stream (category=%s)", uid, eventType, pose, areaType, alarmCategory)
		return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "alarm", alarmCategory, dataValue, message)
	}

	// 保持原 topic，发布到 event stream
	log.Printf("[EVENT_STREAM] device=%s event_type=%s pose=%v area_type=%v published to event stream (not converted to alarm)", uid, eventType, pose, areaType)
	return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "event", "event", dataValue, message)
}

// handleAlarmMessage 处理告警消息
func (c *MQTTConsumer) handleAlarmMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling alarm message for device %s", uid)

	ctx := context.Background()

	// 使用 CardMappingService 获取 cardID 和其他设备映射信息
	cardDeviceInfo, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid)
	if err != nil {
		log.Printf("Failed to get cardID for device %s: %v", uid, err)
		// cardID 获取失败，不能处理该设备消息
		return nil
	}

	// cardDeviceInfo 不能为 nil，否则无法处理
	if cardDeviceInfo == nil {
		log.Printf("Device %s not found in card mapping, skipping alarm message (device not authenticated)", uid)
		return nil
	}

	var cardID, tenantID, deviceID string
	cardID = cardDeviceInfo.CardID
	tenantID = cardDeviceInfo.TenantID
	deviceID = cardDeviceInfo.DeviceID

	// 调用 RadarDecoder 进行协议层面的解码
	dataValue, err := decode.RadarDecoder(message, "alarm")
	if err != nil {
		log.Printf("[ALARM_HANDLER] Failed to decode alarm data for device %s: %v", uid, err)
		// 解码失败时，使用原始消息（降级处理）
		dataValue = message
	}

	// 提取关键字段用于日志
	var eventType, pose, areaType interface{}
	if dataValueMap, ok := dataValue.(map[string]interface{}); ok {
		eventType, _ = dataValueMap["event_type"]
		pose, _ = dataValueMap["pose"]
		areaType, _ = dataValueMap["area_type"]
	}

	// 记录详细的 alarm 处理日志
	log.Printf("[ALARM_HANDLER] device=%s event_type=%v pose=%v area_type=%v published to alarm stream", uid, eventType, pose, areaType)

	// 调试：记录解码后的 dataValue 结构
	if dataValueMap, ok := dataValue.(map[string]interface{}); ok {
		log.Printf("[ALARM_DEBUG] decoded dataValue keys: %v", getMapKeys(dataValueMap))
	}

	// 发布解码后的数据
	// 注意：publishDecodedData 会使用数据项自己的 category（enter2out, pose 等）作为顶层 category
	return c.publishDecodedData(ctx, cardID, tenantID, deviceID, "alarm", "", dataValue, message)
}
