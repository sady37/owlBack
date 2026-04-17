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
	"owl-common/card"
	"owl-common/observation"
	"owl-common/radar"
	rediscommon "owl-common/redis"
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

func businessAccessApproved(s string) bool {
	switch s {
	case "approved", "enable":
		return true
	default:
		return false
	}
}

// resolveIotPolicy 租户级：business_access 为 approved|enable 才向 iot:* 发 event/alarm/stat；monitor 流另需 monitoring_enabled。
// tenantID 来自 devices.tenant_id，供写入 iot 流，避免仅 card 解析时 tenant 为空导致 wisefido-iot 缺 tenant_id。
func (c *MQTTConsumer) resolveIotPolicy(ctx context.Context, deviceUID string) (canIoT, canMonitor bool, tenantID string) {
	if c.cardMappingService != nil {
		if b, ok := c.cardMappingService.BaselineFor(deviceUID); ok {
			tenantID = strings.TrimSpace(b.TenantID)
			if tenantID == "" {
				return false, false, ""
			}
			if !businessAccessApproved(b.BusinessAccess) {
				return false, false, tenantID
			}
			return true, b.MonitoringEnabled, tenantID
		}
	}
	dev, err := c.deviceRepo.GetDeviceByUID(ctx, deviceUID)
	if err != nil || dev == nil {
		return false, false, ""
	}
	tenantID = strings.TrimSpace(dev.TenantID)
	if !businessAccessApproved(dev.BusinessAccess) {
		return false, false, tenantID
	}
	return true, dev.MonitoringEnabled, tenantID
}

// CardIDProvider 定义获取设备卡片映射信息的接口
type CardIDProvider interface {
	GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*card.DeviceBaseline, error)
	BaselineFor(deviceUID string) (card.DeviceBaseline, bool)
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

// publishOnlineForConnectedAfterStartup 启动 1 分钟后，从已订阅主题（MQTT 队列）中取所有 func 主题对应的设备 UID，按统一 IoTHead 发 topic=alarm、category=OfflineRecover 到 iot:alarm:stream，由 cardagg 更新 cardstatus.device_status
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
	if len(uidSet) == 0 {
		return
	}
	ts := time.Now().UnixMilli()
	for uid := range uidSet {
		canIoT, _, policyTid := c.resolveIotPolicy(ctx, uid)
		if !canIoT {
			continue
		}
		tid, _, _, cid, did, _, _, ok := c.resolveDeviceIdentity(ctx, uid)
		if !ok {
			continue
		}
		if policyTid != "" {
			tid = policyTid
		}
		// card_id comes from resolveDeviceIdentity via CardMappingService
		item := observation.EventItem{
			DataCategory: alarm.AlarmTypeOfflineRecover,
			EventName:    alarm.AlarmTypeOfflineRecover,
			EventSince:   ts,
			EventStatus:  "end",
			TrackID:      observation.TrackDevice,
		}
		data, _ := observation.EventItemToDataMap(&item)
		if data == nil {
			data = make(map[string]interface{})
		}
		data[observation.FieldOffline] = 0
		msg := rediscommon.NewSingleItemMessage(tid, cid, uid, did, DeviceTypeRadar, ts, "alarm", alarm.AlarmTypeOfflineRecover, data)
		_ = c.streamPublisher.PublishAlarm(ctx, msg)
	}
	// 已由上方 iot:alarm:stream OfflineRecover 由 cardagg 更新 device_status，不再走 event 流
	// if c.subscriptionManager != nil {
	// 	uids := make([]string, 0, len(uidSet))
	// 	for uid := range uidSet {
	// 		uids = append(uids, uid)
	// 	}
	// 	c.subscriptionManager.PublishOnlineForConnectedDevices(ctx, uids)
	// }
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
func (c *MQTTConsumer) allowAccessFromCacheOrDB(uid string) bool {
	cached, ok := domain.AllowAccessCache.Load(uid)
	if !ok {
		ds, err := c.deviceRepo.GetDeviceStoreInfo(context.Background(), uid)
		if err != nil {
			log.Printf("Device %s not in cache, rejecting message (device not authenticated): db lookup failed: %v", uid, err)
			domain.AllowAccessCache.Store(uid, false)
			return false
		}
		if !ds.AllowAccess {
			log.Printf("Device %s blocked by device_store: allow_access=FALSE", uid)
			domain.AllowAccessCache.Store(uid, false)
			return false
		}
		domain.AllowAccessCache.Store(uid, true)
		return true
	}
	if allowedBool, ok := cached.(bool); !ok || !allowedBool {
		log.Printf("Device %s blocked by cache: allow_access=FALSE", uid)
		return false
	}
	return true
}

// handleMessage 处理 MQTT 消息（仅 .../post）
// 注意：现在只订阅已认证设备的主题，未认证设备无法发送消息到服务端
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	if isQinglanVerboseLog() {
		log.Printf("[MQTT_RX] topic=%s payload=%s", topic, string(payload))
	}

	// 解析主题，提取设备UID
	uid, err := c.extractUIDFromTopic(topic)
	if err != nil {
		log.Printf("Failed to extract UID from topic %s: %v", topic, err)
		return nil // 不返回错误，继续处理其他消息
	}

	// 优先 DeviceBaseline 缓存（web auth / cardChange 后 RefreshBaseline）；未命中再走 AllowAccessCache / DB
	if c.cardMappingService != nil {
		if b, ok := c.cardMappingService.BaselineFor(uid); ok {
			if !b.AllowAccess {
				log.Printf("Device %s blocked by baseline cache: allow_access=FALSE", uid)
				return nil
			}
		} else if !c.allowAccessFromCacheOrDB(uid) {
			return nil
		}
	} else if !c.allowAccessFromCacheOrDB(uid) {
		return nil
	}

	// 解析消息体
	var message map[string]interface{}
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("Failed to parse MQTT message: %v", err)
		return nil
	}

	// 根据主题类型处理消息
	topicType := c.extractTopicType(topic)

	// if topicType == "event" || topicType == "alarm" || topicType == "stat" {
	// 	log.Printf("[MQTT_RAW] topic=%s uid=%s payload=%s", topic, uid, string(payload))
	// }

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
	case "event", "alarm":
		return c.handleEventMessage(uid, message)
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

// DEPRECATED: publishDecodedData — 已被 observation.Message 模式替代，monitor/stat/event 均已迁移。
// 保留备查，待确认无遗漏后删除。
/*
func (c *MQTTConsumer) publishDecodedData(
	ctx context.Context, cardID, tenantID, deviceID, topicType, category string,
	dataValue interface{}, originalMessage map[string]interface{},
) error { ... }
*/

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
	// 属性读/写为低频，始终打设备原始回包（不依赖 QINGLAN_VERBOSE_LOG）
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

// handleFunctionMessage 处理功能响应消息（/func/.../post）
// 实时交互通过 HTTP API 对外提供：客户端调 POST /api/v1/radar/devices/{uid}/function 时，
// RadarService 发 MQTT 到 /func/.../get 并轮询 Redis；设备在 /func/.../post 回包后，此处提取
// requestId 并存 Redis，RadarService 方能取到响应并返回给 HTTP 调用方。
func (c *MQTTConsumer) handleFunctionMessage(uid string, message map[string]interface{}) error {
	log.Printf("Handling function message for device %s", uid)
	var requestID string
	if id, ok := message["requestId"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := message["request_id"].(string); ok && id != "" {
		requestID = id
	} else if id, ok := message["requestID"].(string); ok && id != "" {
		requestID = id
	}
	if requestID != "" {
		ctx := context.Background()
		if err := c.streamPublisher.StoreCommandResponse(ctx, requestID, message); err != nil {
			log.Printf("Failed to store function response for requestId %s: %v", requestID, err)
		} else {
			log.Printf("Stored function response for requestId %s from device %s", requestID, uid)
		}
	} else {
		log.Printf("Function message from device %s has no requestId, treating as function status notification", uid)
	}
	return nil
}

// resolveDeviceIdentity 统一解析 device_uid → (tid, bid, unitID, cid, did, bedID, roomID)。monitor/stat/event 共用。
// ok=false 表示无 device_store 或 DeviceID 为空，调用方应直接 return。
func (c *MQTTConsumer) resolveDeviceIdentity(ctx context.Context, uid string) (tid, bid, unitID, cid, did, bedID, roomID string, ok bool) {
	ds, err := c.deviceRepo.GetDeviceStoreInfo(ctx, uid)
	if err != nil || ds == nil || ds.DeviceID == "" {
		return "", "", "", "", "", "", "", false
	}
	tid = strings.TrimSpace(ds.TenantID)
	did = strings.TrimSpace(ds.DeviceID)
	if tid == "" {
		return "", "", "", "", "", "", "", false
	}
	if info, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid); err == nil && info != nil {
		bid, unitID, cid = info.BranchID, info.UnitID, info.CardID
		bedID, roomID = info.BedID, info.RoomID
	}
	return tid, bid, unitID, cid, did, bedID, roomID, true
}

// publishRadarMonitorHeartbeat monitoring 关闭时发单条 track_id=11（设备级），category=heart，供 cardagg MonitorBuffer 推导在线。
func (c *MQTTConsumer) publishRadarMonitorHeartbeat(ctx context.Context, tid, cid, uid, did string, ts int64) error {
	t := observation.Track{TrackID: observation.TrackDevice, TrackConfidence: 60}
	data := t.ToFieldMap()
	msg := rediscommon.NewSingleItemMessage(tid, cid, uid, did, DeviceTypeRadar, ts, "monitor", observation.CategoryHeart, data)
	return c.streamPublisher.PublishMonitor(ctx, msg)
}

// handleMonitorMessage 处理实时数据消息 (target 模式)
// 解码后经 TargetMergeVital 合并/拆分，每条发到 iot:monitor:stream，category 均为 track（与 field/type 一致）。
// 流使用 device_uid；无 device_store 记录不发。租户级需 business_access；monitoring_enabled 关闭时仅发 track_id=11 心跳。
func (c *MQTTConsumer) handleMonitorMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	tid, bid, unitID, cid, did, bedID, roomID, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, policyTid := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
	}
	if policyTid != "" {
		tid = policyTid
	}
	// card_id comes from resolveDeviceIdentity via CardMappingService
	dataValue, err := decode.RadarDecoder(message, "monitor")
	if err != nil {
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
		return nil
	}

	if len(items) > 0 && canMonitor && isQinglanVerboseLog() {
		for _, it := range items {
			if dc, _ := it["dataCategory"].(string); dc == "track" {
				log.Printf("[MONITOR_TRACK] device_uid=%s track_id=%v position_x=%v position_y=%v position_z=%v remaining_time=%v area_id=%v pose=%v event=%v",
					uid, it["track_id"], it["position_x"], it["position_y"], it["position_z"],
					it["remaining_time"], it["area_id"], it["pose"], it["event"])
			}
		}
	}

	if len(items) == 0 {
		return nil
	}

	ts := time.Now().UnixMilli()
	if canMonitor {
		msgs := TargetMergeVital(items, ts, tid, bid, unitID, cid, uid, did, bedID, roomID)
		var lastErr error
		for _, msg := range msgs {
			if err := c.streamPublisher.PublishMonitor(ctx, msg); err != nil {
				log.Printf("Failed to publish monitor for device %s: %v", uid, err)
				lastErr = err
			}
		}
		return lastErr
	}
	return c.publishRadarMonitorHeartbeat(ctx, tid, cid, uid, did, ts)
}

// TargetMergeVital 根据 decode 结果合并 vital：仅当本条 MQTT 内「在 Bed 区域」的 track 恰有 1 个时合并到该 track，否则 vital 单独成条 track_id=9。
// 入参 deviceUID 为流用设备标识；deviceID 写入包头；bedID/roomID 用于 vital 单独成条时的 area_id。返回每条 category 均为 track。
func TargetMergeVital(
	items []map[string]interface{},
	ts int64,
	tid, bid, unitID, cid, deviceUID, deviceID, bedID, roomID string,
) []*rediscommon.IoTStreamMessage {
	var tracks []map[string]interface{}
	var vitals []map[string]interface{}
	for _, m := range items {
		cat, _ := m["category"].(string)
		if cat == "" {
			cat, _ = m["dataCategory"].(string) // decode 输出为 dataCategory
		}
		switch cat {
		case "track":
			tracks = append(tracks, m)
		case "vital":
			vitals = append(vitals, m)
		}
	}

	areaID := bedID
	if areaID == "" {
		areaID = roomID
	}

	var oneTrack, oneVital map[string]interface{}
	if len(tracks) == 1 && len(vitals) == 1 {
		oneTrack, oneVital = tracks[0], vitals[0]
	}

	var out []*rediscommon.IoTStreamMessage
	if oneTrack != nil && oneVital != nil {
		data := radarTrackToData(oneTrack)
		mergeRadarVital(data, oneVital)
		out = append(out, rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
	} else {
		for _, tr := range tracks {
			data := radarTrackToData(tr)
			if len(tracks) > 1 {
				data[observation.FieldTrackCount] = len(tracks)
			}
			out = append(out, rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
		}
		for _, v := range vitals {
			data := map[string]any{
				observation.FieldTrackID: observation.TrackUnknownPerson,
			}
			if areaID != "" {
				data[observation.FieldAreaID] = areaID
			}
			mergeRadarVital(data, v)
			out = append(out, rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
		}
	}
	return out
}

// radarTrackToData converts a decoded radar track item to observation standard field map.
func radarTrackToData(tr map[string]interface{}) map[string]any {
	t := observation.Track{}
	if v, ok := tr[observation.FieldTrackID]; ok {
		t.TrackID = asInt(v)
	}
	if v, ok := tr["position_x"]; ok {
		px := asInt(v)
		t.PositionX = &px
	}
	if v, ok := tr["position_y"]; ok {
		py := asInt(v)
		t.PositionY = &py
	}
	if v, ok := tr["position_z"]; ok {
		pz := asInt(v)
		t.PositionZ = &pz
	}
	if v, ok := tr["pose"]; ok {
		t.Pose = asInt(v)
	}
	if v, ok := tr["event"]; ok {
		t.Event = asInt(v)
	}
	data := t.ToFieldMap()
	if v, ok := tr["remaining_time"]; ok {
		data[observation.FieldRemainingTime] = asInt(v)
	}
	if v, ok := tr["area_id"]; ok {
		data[observation.FieldAreaID] = asInt(v)
	}
	// 雷达无 signal_quality，轨迹置信度固定 60
	data[observation.FieldTrackConfidence] = 60
	return data
}

// mergeRadarVital merges decoded radar vital fields into the observation data map.
func mergeRadarVital(data map[string]any, vital map[string]interface{}) {
	if hr, ok := vital["heart_rate"]; ok {
		if v := asInt(hr); v > 0 {
			data[observation.FieldHeartRate] = v
		}
	}
	if rr, ok := vital["respiratory_rate"]; ok {
		if v := asInt(rr); v > 0 {
			data[observation.FieldRespiratoryRate] = v
		}
	}
	if st, ok := vital["stability"]; ok {
		switch asInt(st) {
		case 3: // 11 无干扰
			data[observation.FieldVitalConfidence] = 60 //vitalConfidence 60*N/3
		case 2: // 10 较小动作
			data[observation.FieldVitalConfidence] = 40 //2*20=40
		case 1: // 01 较大动作
			data[observation.FieldVitalConfidence] = 20
		}
	}
	if ss, ok := vital["sleep_status"]; ok {
		data[observation.FieldSleepStage] = radarSleepStageToSleepad(asInt(ss))
	}
	if _, set := data[observation.FieldVitalConfidence]; !set {
		data[observation.FieldVitalConfidence] = 60
	}
}

// radarSleepStageToSleepad 将雷达 sleep_status(0=unknown,1=lightSleep,2=deepSleep,3=awake) 转为 observation 统一值：1=清醒, 2=浅睡, 4=深睡, 8=未知。
func radarSleepStageToSleepad(radar int) int {
	switch radar {
	case 3:
		return 1 // awake
	case 1:
		return 2 // light
	case 2:
		return 4 // deep
	case 0:
		return 8 // unknown
	default:
		return 8
	}
}

func asInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int64:
		return int(val)
	}
	return 0
}

// handleStatMessage 处理统计数据消息（activity×1 + sleep×1）。
//   - stat track (activity) → iot:event:stream, category=activity（状态汇总：people_count, walk_distance 等）
//   - stat sleep            → iot:alarm/event（见 publishStatSleep）
func (c *MQTTConsumer) handleStatMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	tid, bid, unitID, cid, did, _, _, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, policyTid := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
	}
	if policyTid != "" {
		tid = policyTid
	}
	// card_id comes from resolveDeviceIdentity via CardMappingService
	dataValue, err := decode.RadarDecoder(message, "stat")
	if err != nil {
		log.Printf("[STAT_HANDLER] decode failed for device %s: %v", uid, err)
		return nil
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
		return nil
	}

	ts := time.Now().UnixMilli()
	var lastErr error

	for _, m := range items {
		cat, _ := m["dataCategory"].(string)
		if cat == "" {
			cat, _ = m["category"].(string)
		}
		switch cat {
		case observation.FieldActivity:
			if err := c.publishStatActivity(ctx, tid, bid, unitID, cid, uid, did, ts, m); err != nil {
				lastErr = err
			}
		case observation.CategorySleep:
			if err := c.publishStatSleep(ctx, tid, bid, unitID, cid, uid, did, ts, m); err != nil {
				lastErr = err
			}
		}
	}
	if canIoT && !canMonitor && len(items) > 0 {
		if err := c.publishRadarMonitorHeartbeat(ctx, tid, cid, uid, did, ts); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// publishStatActivity 发布老人活动性状态汇总到 iot:event:stream，payload 符合 EventItem 格式，activity 字段平铺。
func (c *MQTTConsumer) publishStatActivity(ctx context.Context, tid, bid, unitID, cid, deviceUID, deviceID string, ts int64, m map[string]interface{}) error {
	item := observation.EventItem{
		DataCategory: observation.FieldActivity,
		EventName:    observation.FieldActivity,
		EventSince:   ts,
		EventStatus:  "instant",
		TrackID:      observation.TrackUnknownPerson,
	}
	data, err := observation.EventItemToDataMap(&item)
	if err != nil {
		return err
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	if v, ok := m["people_count"]; ok {
		data[observation.FieldTrackCount] = asInt(v)
	}
	if v, ok := m["walk_distance"]; ok {
		data[observation.FieldWalkDistance] = asInt(v)
	}
	if v, ok := m["walk_duration"]; ok {
		data[observation.FieldWalkDuration] = asInt(v)
	}
	if v, ok := m["lie_duration"]; ok {
		data[observation.FieldLieDuration] = asInt(v)
	}
	if v, ok := m["stand_duration"]; ok {
		data[observation.FieldStandDuration] = asInt(v)
	}
	if v, ok := m["multi_person_duration"]; ok {
		data[observation.FieldMultiPersonDuration] = asInt(v)
	}
	msg := rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "event", observation.FieldActivity, data)
	return c.streamPublisher.PublishEvent(ctx, msg)
}

// publishStatSleep 雷达 state.sleep 异常：按 monitor 阀值已明确判定，发 iot:alarm:stream；三者都正常发 event sleep。
// breath: 00=正常 01→RespRateAlertHigh 10→RespRateAlertLow 11→ApneaHypopnea
// heart: 00=正常 01→HeartRateAlertLow 10→HeartRateAlertHigh 11=未定义不告警
// vital_signs: 00=正常 11→WeakBiometricSignal
func (c *MQTTConsumer) publishStatSleep(ctx context.Context, tid, bid, unitID, cid, deviceUID, deviceID string, ts int64, m map[string]interface{}) error {
	breathState := asInt(m["breath_state"]) & 0x03
	heartState := asInt(m["heart_state"]) & 0x03
	vitalSignsState := asInt(m["vital_signs_state"]) & 0x03

	publishAlarm := func(category string, eventValue int64, eventReason string, rawDecode map[string]interface{}) error {
		payloadJSON, _ := json.Marshal(rawDecode)
		item := observation.EventItem{
			DataCategory: category,
			EventName:    category,
			EventSince:   ts,
			EventStatus:  "instant",
			EventValue:   eventValue,
			EventReason:  eventReason,
			EventPayload: string(payloadJSON),
			TrackID:      observation.TrackUnknownPerson,
		}
		alarmData, _ := observation.EventItemToDataMap(&item)
		if alarmData == nil {
			alarmData = make(map[string]interface{})
		}
		alarmMsg := rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "alarm", category, alarmData)
		return c.streamPublisher.PublishAlarm(ctx, alarmMsg)
	}

	if breathState != 0 {
		ev := int64(asInt(m["respiratory_rate"]))
		if ev == 0 {
			ev = int64(asInt(m["avg_respiratory_rate"]))
		}
		avgBreathe := asInt(m["avg_respiratory_rate"])
		reason := fmt.Sprintf("avgBreathe=%d", avgBreathe)
		switch breathState {
		case 1:
			_ = publishAlarm(alarm.RespRateAlertHigh, ev, reason, m)
		case 2:
			_ = publishAlarm(alarm.RespRateAlertLow, ev, reason, m)
		case 3:
			_ = publishAlarm(alarm.ApneaHypopnea, ev, reason, m)
		}
	}
	if heartState == 1 {
		hr := int64(asInt(m["heart_rate"]))
		if hr == 0 {
			hr = int64(asInt(m["avg_heart_rate"]))
		}
		avgHeart := asInt(m["avg_heart_rate"])
		_ = publishAlarm(alarm.HeartRateAlertLow, hr, fmt.Sprintf("avgHeart=%d", avgHeart), m)
	} else if heartState == 2 {
		hr := int64(asInt(m["heart_rate"]))
		if hr == 0 {
			hr = int64(asInt(m["avg_heart_rate"]))
		}
		avgHeart := asInt(m["avg_heart_rate"])
		_ = publishAlarm(alarm.HeartRateAlertHigh, hr, fmt.Sprintf("avgHeart=%d", avgHeart), m)
	}
	if vitalSignsState == 3 {
		_ = publishAlarm(alarm.WeakBiometricSignal, int64(vitalSignsState)*20, alarm.WeakBiometricSignal, m)
	}

	if breathState != 0 || heartState != 0 || vitalSignsState == 3 {
		return nil
	}
	data := make(map[string]interface{}, len(m)+4)
	for k, v := range m {
		data[k] = v
	}
	data["event_name"] = alarm.SleepStage
	data["event_since"] = ts
	data["event_status"] = "instant"
	data["track_id"] = observation.TrackUnknownPerson
	cat := alarm.SleepStage // 与 Sleepad sleepStage 统一为 sleep-stage
	msg := rediscommon.NewSingleItemMessage(tid, cid, deviceUID, deviceID, DeviceTypeRadar, ts, "event", cat, data)
	return c.streamPublisher.PublishEvent(ctx, msg)
}

// handleEventMessage 处理事件/告警消息（event + alarm topic 统一入口），payload 符合 EventItem 格式。
// decoder 输出 event_name（如 InBed、Fall）；consumer 用其值作 category 注入 dataCategory，用于业务路由。
// 按 event_type 分配 track_id：type=1/2 人、type=3 TrackSpace、type=5/7/8 TrackDevice。流使用 device_uid。
func (c *MQTTConsumer) handleEventMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	tid, _, _, cid, did, _, _, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, policyTid := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
	}
	if policyTid != "" {
		tid = policyTid
	}
	// card_id comes from resolveDeviceIdentity via CardMappingService
	dataValue, err := decode.RadarDecoder(message, "event")
	if err != nil {
		log.Printf("[EVENT_DECODE_ERROR] device=%s: %v", uid, err)
		return nil
	}

	var items []map[string]interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, m)
			}
		}
	case map[string]interface{}:
		items = []map[string]interface{}{v}
	default:
		return nil
	}

	ts := time.Now().UnixMilli()
	var lastErr error

	for _, m := range items {
		eventName, _ := m["event_name"].(string)
		if eventName == "" {
			eventName, _ = m["data_category"].(string)
		}
		if eventName == "" {
			continue
		}
		eventType := asInt(m["event_type"])

		item := observation.EventItem{
			DataCategory: eventName,
			EventName:    eventName,
			EventSince:   ts,
			EventStatus:  "start",
		}
		switch eventType {
		case 1:
			item.TrackID = asInt(m["track_id"])
		case 2:
			item.TrackID = asInt(m["track_id"])
		case 3:
			item.TrackID = observation.TrackSpace
		case 5, 7, 8:
			item.TrackID = observation.TrackDevice
		default:
			continue
		}
		data, err := observation.EventItemToDataMap(&item)
		if err != nil {
			continue
		}
		if data == nil {
			data = make(map[string]interface{})
		}
		switch eventType {
		case 1:
			// area_type 写入 data，供下游区分区域（床区/感应区等）

		case 2:
			// Fall/SittingOnGround 走 alarm cardagg event_handler 按使能落库
			alarmCat := ""
			switch eventName {
			case alarm.Fall, alarm.SuspectedFall:
				alarmCat = alarm.Fall
			case alarm.SittingOnGround, alarm.SuspectedSittingOnGround:
				alarmCat = alarm.SittingOnGround
			}
			if alarmCat != "" {
				payloadJSON, _ := json.Marshal(m)
				alarmItem := observation.EventItem{
					DataCategory: alarmCat,
					EventName:    alarmCat,
					EventSince:   ts,
					EventStatus:  "start",
					EventPayload: string(payloadJSON),
					TrackID:      asInt(m["track_id"]),
				}
				alarmData, _ := observation.EventItemToDataMap(&alarmItem)
				if alarmData == nil {
					alarmData = make(map[string]interface{})
				}
				evMsg := rediscommon.NewSingleItemMessage(tid, cid, uid, did, DeviceTypeRadar, ts, "alarm", alarmCat, alarmData)
				if err := c.streamPublisher.PublishAlarm(ctx, evMsg); err != nil {
					log.Printf("[EVENT_HANDLER] event publish failed device=%s cat=%s: %v", uid, alarmCat, err)
					lastErr = err
				}
				continue
			}
			if v, ok := m["pose"]; ok {
				data[observation.FieldPose] = asInt(v)
			}
		case 3:
			if v, ok := m[observation.FieldNumberPeople]; ok {
				data[observation.FieldNumberPeople] = asInt(v)
			}
		case 5, 7, 8:
			if v, ok := m["status_type"]; ok {
				if s, ok := v.(string); ok {
					data[s] = asInt(m["status_value"])
				}
			}
		}

		msg := rediscommon.NewSingleItemMessage(tid, cid, uid, did, DeviceTypeRadar, ts, "event", eventName, data)
		if err := c.streamPublisher.PublishEvent(ctx, msg); err != nil {
			log.Printf("[EVENT_HANDLER] publish failed device=%s cat=%s: %v", uid, eventName, err)
			lastErr = err
		}
	}
	if canIoT && !canMonitor && len(items) > 0 {
		if err := c.publishRadarMonitorHeartbeat(ctx, tid, cid, uid, did, ts); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DEPRECATED: publishSingleTrack — 已被 observation.Message 模式替代。
// func (c *MQTTConsumer) publishSingleTrack(...) error { ... }

// DEPRECATED: eventDataCategory — 已被 handleEventMessage 内联替代。
// func eventDataCategory(track interface{}) string { ... }
