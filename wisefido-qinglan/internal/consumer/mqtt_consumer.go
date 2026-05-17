package consumer

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
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
	"go.uber.org/zap"
)

const (
	// DeviceTypeRadar 设备类型常量（wisefido-qinglan网关的所有设备都是Radar）
	DeviceTypeRadar = "Radar"
)

// resolveIotPolicy 租户级：access=TRUE 才向 iot:* 发 event/alarm/stat；monitor 流另需 monitoring_enabled。
// tenantID 来自 devices.tenant_id，供写入 iot 流，避免仅 card 解析时 tenant 为空导致 wisefido-iot 缺 tenant_id。
//
// Access 是 platform_admin 审批位（devices.access bool）；v1 的 allow_access/business_access 双字段
// 系同源同语义，已合并为单一 Access。
func (c *MQTTConsumer) resolveIotPolicy(ctx context.Context, deviceUID string) (canIoT, canMonitor bool, tenantID string) {
	if c.cardMappingService != nil {
		if b, ok := c.cardMappingService.BaselineFor(deviceUID); ok {
			tenantID = strings.TrimSpace(b.TenantID)
			if tenantID == "" {
				return false, false, ""
			}
			if !b.Access {
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
	if !dev.Access {
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
	db                  *sql.DB // database for OTA status updates
	logger              *zap.Logger
}

// SetDB sets the database connection for OTA operations
func (c *MQTTConsumer) SetDB(db *sql.DB) {
	c.db = db
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
	logger *zap.Logger,
) (*MQTTConsumer, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MQTTConsumer{
		config:              cfg,
		mqttClient:          mqttClient,
		redisClient:         redisClient,
		deviceRepo:          deviceRepo,
		cardMappingService:  cardMappingService,
		streamPublisher:     streamPublisher,
		subscriptionManager: subscriptionManager,
		subscribedTopics:    make(map[string]struct{}),
		logger:              logger,
	}, nil
}

// Start 启动消费者
// 启动时主动订阅所有符合条件的设备（allow_access=TRUE 且 business_access='approved'）
func (c *MQTTConsumer) Start(ctx context.Context) error {
	go c.monitorConnection(ctx)
	go c.subscribeAllAccessibleDevices(ctx)
	c.logger.Info("mqtt consumer started")
	return nil
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
		c.logger.Warn("mqtt not connected, skip initial subscription", zap.Duration("waited", maxWaitTime))
		return
	}

	deviceUIDs, err := c.deviceRepo.GetAllAccessibleDevices(ctx)
	if err != nil {
		c.logger.Error("get accessible devices", zap.Error(err))
		return
	}
	if len(deviceUIDs) == 0 {
		c.logger.Info("no accessible devices, skip subscription")
		return
	}
	c.logger.Info("allowed devices", zap.Int("count", len(deviceUIDs)), zap.Strings("uids", deviceUIDs))

	successCount := 0
	for _, deviceUID := range deviceUIDs {
		if err := c.SubscribeDeviceTopics(deviceUID); err != nil {
			c.logger.Error("subscribe device topics", zap.String("device_uid", deviceUID), zap.Error(err))
		} else {
			successCount++
		}
	}
	c.logger.Info("subscribed", zap.Int("success", successCount), zap.Int("total", len(deviceUIDs)))
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

			if !wasConnected && isConnected {
				c.logger.Info("mqtt reconnected, resubscribing")
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
			c.logger.Warn("resubscribe topic", zap.String("topic", topic), zap.Error(err))
		}
	}
	c.logger.Info("resubscribed", zap.Int("count", len(topics)))
}

// SubscribeDeviceTopics 订阅指定设备的 6 个主题（不打印每条订阅，由调用方打印允许设备列表）
func (c *MQTTConsumer) SubscribeDeviceTopics(deviceUID string) error {
	if !c.mqttClient.IsConnected() {
		c.logger.Warn("mqtt not connected, cannot subscribe", zap.String("device_uid", deviceUID))
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
			c.logger.Warn("subscribe topic", zap.String("topic", topic), zap.Error(err))
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
			c.logger.Warn("unsubscribe topic", zap.String("topic", topic), zap.Error(err))
			continue
		}
		delete(c.subscribedTopics, topic)
		c.logger.Debug("unsubscribed topic", zap.String("topic", topic))
	}
	c.logger.Info("unsubscribed device", zap.String("device_uid", deviceUID), zap.Int("count", len(topics)))
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
	c.logger.Info("mqtt consumer stopping")
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
			c.logger.Warn("device not in cache, db lookup failed", zap.String("uid", uid), zap.Error(err))
			domain.AllowAccessCache.Store(uid, false)
			return false
		}
		if !ds.Access {
			c.logger.Warn("device blocked by dfm: access=FALSE", zap.String("uid", uid))
			domain.AllowAccessCache.Store(uid, false)
			return false
		}
		domain.AllowAccessCache.Store(uid, true)
		return true
	}
	if allowedBool, ok := cached.(bool); !ok || !allowedBool {
		c.logger.Debug("device blocked by cache", zap.String("uid", uid))
		return false
	}
	return true
}

// handleMessage 处理 MQTT 消息（仅 .../post）
// 注意：现在只订阅已认证设备的主题，未认证设备无法发送消息到服务端
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
	uid, err := c.extractUIDFromTopic(topic)
	if err != nil {
		c.logger.Warn("extract uid from topic", zap.String("topic", topic), zap.Error(err))
		return nil
	}
	topicType := c.extractTopicType(topic)

	// Info 给人看：用 rxSummary 翻成人话，不带 raw payload（payload 想看走 Debug，或查 event_log/monitor_stream）。
	//   event_type=1  → EnterRoom/ExitRoom/InBed/LeftBed
	//   event_type=2  → Fall/SittingOnGround（含 Suspected*）+ pose 名称
	//   prop topic    → 设备属性查询/设置应答（key=value）
	// 其它（monitor/stat 周期流 + number_people + 设备状态 + func 应答）→ Debug 带 raw payload
	if rxIsAlarmClass(topicType, payload) {
		c.logger.Info("mqtt rx",
			zap.String("uid", uid),
			zap.String("type", topicType),
			zap.String("info", rxSummary(topicType, payload)),
		)
	} else {
		c.logger.Debug("mqtt rx",
			zap.String("uid", uid),
			zap.String("type", topicType),
			zap.String("topic", topic),
			zap.ByteString("payload", payload),
		)
	}

	if c.cardMappingService != nil {
		if b, ok := c.cardMappingService.BaselineFor(uid); ok {
			if !b.Access {
				c.logger.Debug("blocked by baseline cache", zap.String("uid", uid))
				return nil
			}
		} else if !c.allowAccessFromCacheOrDB(uid) {
			return nil
		}
	} else if !c.allowAccessFromCacheOrDB(uid) {
		return nil
	}

	var message map[string]interface{}
	if err := json.Unmarshal(payload, &message); err != nil {
		c.logger.Warn("parse mqtt payload", zap.Error(err))
		return nil
	}

	if c.subscriptionManager != nil {
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
		c.logger.Debug("unknown topic type", zap.String("type", topicType))
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

// rxSummary 把原始 MQTT payload 翻译成人话，供 Info-level mqtt rx log 的 summary 字段。
// 失败返回空串（payload 已在 ByteString 字段保留，不丢信息）。
func rxSummary(topicType string, payload []byte) string {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return ""
	}
	switch topicType {
	case "prop":
		// {"data":{"<key>":"<value>"}, "cmd":"read|update", "code":200, "requestId":"..."}
		cmd, _ := msg["cmd"].(string)
		code := asInt(msg["code"])
		parts := []string{}
		if cmd != "" {
			parts = append(parts, fmt.Sprintf("cmd=%s", cmd))
		}
		if code != 0 {
			parts = append(parts, fmt.Sprintf("code=%d", code))
		}
		if data, ok := msg["data"].(map[string]interface{}); ok {
			for k, v := range data {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
		}
		return strings.Join(parts, " ")
	case "event", "alarm":
		eventType := asInt(msg["type"])
		switch eventType {
		case 1:
			// {"data":[{"track_id":N,"event":N,"area_type":N,"area_id":N}]}
			arr, _ := msg["data"].([]interface{})
			if len(arr) == 0 {
				return fmt.Sprintf("event_type=1")
			}
			m, _ := arr[0].(map[string]interface{})
			ev := asInt(m["event"])
			areaType := asInt(m["area_type"])
			cat, _ := alarm.LookupEnter2Out(ev, areaType)
			if cat == "" {
				cat = "?"
			}
			return fmt.Sprintf("%s track=%v area_id=%v area_type=%d", cat, m["track_id"], m["area_id"], areaType)
		case 2:
			// {"data":[{"track_id":N,"pose":N,"last_pose":N}]}
			arr, _ := msg["data"].([]interface{})
			if len(arr) == 0 {
				return "event_type=2"
			}
			m, _ := arr[0].(map[string]interface{})
			pose := asInt(m["pose"])
			lastPose := asInt(m["last_pose"])
			cat, _ := alarm.LookupPose(pose)
			if cat == "" {
				cat = "?"
			}
			return fmt.Sprintf("%s pose=%d(%s) last_pose=%d(%s) track=%v",
				cat, pose, poseDisplay(pose), lastPose, poseDisplay(lastPose), m["track_id"])
		case 3:
			data, _ := msg["data"].(map[string]interface{})
			return fmt.Sprintf("number_people=%v", data["number_people"])
		case 5, 7, 8:
			cat := alarm.DeviceEventMap[eventType]
			return fmt.Sprintf("device_status=%s", cat)
		}
		return fmt.Sprintf("event_type=%d", eventType)
	}
	return ""
}

// poseDisplay pose 数值 → 友好名（observation.EnumPose）
func poseDisplay(p int) string {
	if p >= 0 && p < len(observation.EnumPose) {
		return observation.EnumPose[p]
	}
	return "unknown"
}

// rxIsAlarmClass 只 Info-log 关注路径，其它走 Debug：
//   prop topic                     → Info（设备属性查询/设置应答）
//   event/alarm + event_type=1     → Info（Enter/ExitRoom + InBed/LeftBed）
//   event/alarm + event_type=2     → Info（Fall/SittingOnGround）
// 其它：monitor/stat 周期流、number_people (type=3)、device-status (type=5/7/8)、func 命令应答 → Debug
// 注：device healthcheck fail/recover 跃迁不在此处 — 那是 outbound（health_check.go publishHealthIfChanged 加 Info）
//
// 用 string-scan 而非 json.Unmarshal，避免每条 MQTT 入站都付完整解码代价。
func rxIsAlarmClass(topicType string, payload []byte) bool {
	if topicType == "prop" {
		return true
	}
	if topicType != "event" && topicType != "alarm" {
		return false
	}
	// 协议 firmware 顶层 "type": N（见 Radar_MQTT_v3.0.md §3.8）
	idx := bytes.Index(payload, []byte(`"type"`))
	if idx < 0 {
		return false
	}
	// 跳过 `"type"` + 可空格 + `:` + 可空格 + 数字
	tail := payload[idx+len(`"type"`):]
	for len(tail) > 0 && (tail[0] == ' ' || tail[0] == '\t' || tail[0] == ':') {
		tail = tail[1:]
	}
	if len(tail) == 0 {
		return false
	}
	switch tail[0] {
	case '1', '2':
		return true
	}
	return false
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
	// 属性读其实是高频（30s × 设备数 → 数百行/小时），轮询的设备原始回包仅 verbose 时打。
	// SetDeviceProperties 是低频（手动配置触发），始终打。
	if cmd != "read" {
		c.logger.Debug("device prop raw",
			zap.String("op", logLabel),
			zap.String("device", uid),
			zap.String("request_id", requestIDRaw),
			zap.Any("msg", message),
		)
	}

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
			c.logger.Warn("prop response store redis",
				zap.String("op", logLabel),
				zap.String("request_id", requestID),
				zap.Error(err),
			)
		} else if cmd != "read" {
			c.logger.Debug("prop response stored",
				zap.String("op", logLabel),
				zap.String("device", uid),
				zap.String("request_id", requestID),
			)
		}
	} else {
		c.logger.Warn("prop response missing requestId",
			zap.String("op", logLabel),
			zap.String("device", uid),
		)
	}

	return nil
}

// handleFunctionMessage 处理功能响应消息（/func/.../post）
// 实时交互通过 HTTP API 对外提供：客户端调 POST /api/v1/radar/devices/{uid}/function 时，
// RadarService 发 MQTT 到 /func/.../get 并轮询 Redis；设备在 /func/.../post 回包后，此处提取
// requestId 并存 Redis，RadarService 方能取到响应并返回给 HTTP 调用方。
func (c *MQTTConsumer) handleFunctionMessage(uid string, message map[string]interface{}) error {
	if cmd, ok := message["cmd"].(string); ok && cmd == "ota_return" {
		c.handleOTAReturn(uid, message)
		return nil
	}

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
			c.logger.Warn("function response store redis", zap.String("request_id", requestID), zap.Error(err))
		} else {
			c.logger.Debug("function response stored", zap.String("request_id", requestID), zap.String("device", uid))
		}
	} else {
		c.logger.Debug("function status notification (no requestId)", zap.String("device", uid))
	}
	return nil
}

// handleOTAReturn handles OTA result messages from devices
// 协议字段见 Radar_MQTT_v3.0.md §3.8.4.2：data.{progress,errMsg,uid}
//   progress: 10=雷达下载完成 25=雷达升级完成 56=主控下载完成 100=全部完成 -1=失败
func (c *MQTTConsumer) handleOTAReturn(uid string, message map[string]interface{}) {
	data, _ := message["data"].(map[string]interface{})
	if data == nil {
		c.logger.Warn("ota_return missing data", zap.String("uid", uid))
		return
	}

	progress := 0
	progressSet := false
	if v, ok := data["progress"].(float64); ok {
		progress = int(v)
		progressSet = true
	}
	errMsg := ""
	if v, ok := data["errMsg"].(string); ok {
		errMsg = v
	}

	c.logger.Info("ota return",
		zap.String("uid", uid),
		zap.Int("progress", progress),
		zap.String("err_msg", errMsg),
		zap.Bool("progress_set", progressSet),
	)

	if !progressSet {
		if errMsg == "" {
			c.logger.Warn("ota_return missing progress + errMsg, skip", zap.String("uid", uid))
			return
		}
		progress = -1
	}

	// 映射：progress → ota_status
	//   -1 → failed；100 → success；其它 → pushing（中间态）
	var status string
	switch {
	case progress == -1:
		status = "failed"
	case progress == 100:
		status = "success"
	default:
		status = "pushing"
	}

	if c.db != nil {
		// v2: device_ota PK = device_ipv6；UPDATE 走 dfm → devices.device_ipv6 反查
		_, err := c.db.ExecContext(context.Background(), `
			UPDATE device_ota
			SET status = $1, progress = $2, error = $3, updated_at = NOW()
			WHERE device_ipv6 = (
				SELECT d.device_ipv6
				FROM devices d
				JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
				WHERE dfm.device_uid = $4
			)
		`, status, progress, errMsg, uid)
		if err != nil {
			c.logger.Warn("ota_return device_ota update failed", zap.String("uid", uid), zap.Error(err))
		} else {
			c.logger.Info("ota_return device_ota updated",
				zap.String("uid", uid),
				zap.String("status", status),
				zap.Int("progress", progress),
			)
		}
	}
}

// resolveDeviceIdentity 统一解析 device_uid → 设备身份。monitor/stat/event 共用。
// ok=false 表示无 device_store 或 DeviceAddr 无效，调用方应直接 return。
//
// device_ipv6 单程票后 addr 是路由层主键；其余 tid/bid/unitID/cid/did/bedID/roomID 多数已可从 addr 派生，
// 保留作 cardID（cid）与 SubjectEntity 路由 + 旧业务逻辑兼容。
func (c *MQTTConsumer) resolveDeviceIdentity(ctx context.Context, uid string) (tid, bid, unitID, cid, did, bedID, roomID string, addr netip.Addr, ok bool) {
	ds, err := c.deviceRepo.GetDeviceStoreInfo(ctx, uid)
	if err != nil || ds == nil || ds.DeviceID == "" {
		return "", "", "", "", "", "", "", netip.Addr{}, false
	}
	tid = strings.TrimSpace(ds.TenantID)
	did = strings.TrimSpace(ds.DeviceID)
	addr = ds.DeviceAddr
	if tid == "" || !addr.IsValid() {
		return "", "", "", "", "", "", "", netip.Addr{}, false
	}
	if info, err := c.cardMappingService.GetCardIDByDeviceUID(ctx, uid); err == nil && info != nil {
		bid, unitID, cid = info.BranchID, info.UnitID, info.CardID
		bedID, roomID = info.BedID, info.RoomID
	}
	return tid, bid, unitID, cid, did, bedID, roomID, addr, true
}

// publishRadarMonitorHeartbeat monitoring 关闭时发单条 track_id=11（设备级），category=heart，供 cardagg MonitorBuffer 推导在线。
//
// 手写 map（不复用 observation.Track.ToFieldMap）：那条路径强制写零值 bed_status，
// 雷达心跳与床无关，泄漏出去会污染下游协议。
//
// device_ipv6 单程票：addr 是路由层主键；cid 是 SubjectEntity（unbound 时为空，cardagg LPM 反查兜底）。
func (c *MQTTConsumer) publishRadarMonitorHeartbeat(ctx context.Context, addr netip.Addr, cid string, ts int64) error {
	data := map[string]any{
		observation.FieldTrackID:         observation.TrackDevice,
		observation.FieldTrackConfidence: 80,
	}
	msg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "monitor", observation.CategoryHeart, data)
	return c.streamPublisher.PublishMonitor(ctx, msg)
}

// handleMonitorMessage 处理实时数据消息 (target 模式)
// 解码后经 TargetMergeVital 合并/拆分，每条发到 iot:monitor:stream，category 均为 track（与 field/type 一致）。
// 流使用 device_uid；无 device_store 记录不发。租户级需 business_access；monitoring_enabled 关闭时仅发 track_id=11 心跳。
func (c *MQTTConsumer) handleMonitorMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	_, _, _, cid, _, bedID, roomID, addr, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, _ := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
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

	if len(items) == 0 {
		return nil
	}

	ts := time.Now().UnixMilli()
	if canMonitor {
		msgs := TargetMergeVital(items, ts, addr, cid, bedID, roomID)
		var lastErr error
		for _, msg := range msgs {
			if err := c.streamPublisher.PublishMonitor(ctx, msg); err != nil {
				c.logger.Warn("publish monitor", zap.String("device", uid), zap.Error(err))
				lastErr = err
			}
		}
		return lastErr
	}
	return c.publishRadarMonitorHeartbeat(ctx, addr, cid, ts)
}

// TargetMergeVital 根据 decode 结果合并 vital：仅当本条 MQTT 内「在 Bed 区域」的 track 恰有 1 个时合并到该 track，否则 vital 单独成条 track_id=9。
// 入参 deviceUID 为流用设备标识；deviceID 写入包头；bedID/roomID 用于 vital 单独成条时的 area_id。返回每条 category 均为 track。
func TargetMergeVital(
	items []map[string]interface{},
	ts int64,
	addr netip.Addr,
	cid, bedID, roomID string,
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
		out = append(out, rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
	} else {
		for _, tr := range tracks {
			data := radarTrackToData(tr)
			if len(tracks) > 1 {
				data[observation.FieldTrackCount] = len(tracks)
			}
			out = append(out, rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
		}
		for _, v := range vitals {
			data := map[string]any{
				observation.FieldTrackID: observation.TrackUnknownPerson,
			}
			if areaID != "" {
				data[observation.FieldAreaID] = areaID
			}
			mergeRadarVital(data, v)
			out = append(out, rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "monitor", observation.CategoryTrack, data))
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
	// 雷达无 signal_quality，轨迹置信度固定 80（落在 owlfront ≥80 全饱和档；AI 判 ghost 时覆写到 ≤20）
	data[observation.FieldTrackConfidence] = 80
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
//   - !canMonitor: drop stat data, only send heart
//   - canMonitor:  stat activity → iot:event:stream; stat sleep → iot:alarm/event + always send heart
func (c *MQTTConsumer) handleStatMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	_, _, _, cid, _, _, _, addr, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, _ := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
	}
	// card_id comes from resolveDeviceIdentity via CardMappingService
	dataValue, err := decode.RadarDecoder(message, "stat")
	if err != nil {
		c.logger.Warn("stat decode", zap.String("device", uid), zap.Error(err))
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

	if len(items) == 0 {
		return nil
	}

	// !canMonitor: drop stat data, only send heart
	if !canMonitor {
		return c.publishRadarMonitorHeartbeat(ctx, addr, cid, ts)
	}

	// canMonitor: process stat data normally + send heart
	var lastErr error
	for _, m := range items {
		cat, _ := m["dataCategory"].(string)
		if cat == "" {
			cat, _ = m["category"].(string)
		}
		switch cat {
		case observation.FieldActivity:
			if err := c.publishStatActivity(ctx, addr, cid, ts, m); err != nil {
				lastErr = err
			}
		case observation.CategorySleep:
			if err := c.publishStatSleep(ctx, addr, cid, ts, m); err != nil {
				lastErr = err
			}
		}
	}
	if err := c.publishRadarMonitorHeartbeat(ctx, addr, cid, ts); err != nil {
		lastErr = err
	}
	return lastErr
}

// publishStatActivity 发布老人活动性状态汇总到 iot:event:stream，payload 符合 EventItem 格式，activity 字段平铺。
func (c *MQTTConsumer) publishStatActivity(ctx context.Context, addr netip.Addr, cid string, ts int64, m map[string]interface{}) error {
	item := observation.NewEventItem(ts, "instant")
	item.TrackID = observation.TrackUnknownPerson
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
	msg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "event", observation.FieldActivity, data)
	return c.streamPublisher.PublishEvent(ctx, msg)
}

// publishStatSleep 雷达 state.sleep 异常：按 monitor 阀值已明确判定，发 iot:alarm:stream；三者都正常发 event sleep。
// breath: 00=正常 01→RespRateAlertHigh 10→RespRateAlertLow 11→ApneaHypopnea
// heart: 00=正常 01→HeartRateAlertLow 10→HeartRateAlertHigh 11=未定义不告警
// vital_signs: 00=正常 11→WeakBiometricSignal
func (c *MQTTConsumer) publishStatSleep(ctx context.Context, addr netip.Addr, cid string, ts int64, m map[string]interface{}) error {
	breathState := asInt(m["breath_state"]) & 0x03
	heartState := asInt(m["heart_state"]) & 0x03
	vitalSignsState := asInt(m["vital_signs_state"]) & 0x03

	// 业务字段（heart_rate / respiratory_rate / weak_biometric_signal）按 category 平铺到 dataValue 顶层。
	// HR/RR 升级为 EventItem first-class 字段（NewEventItem 默认 -1=未测量）；
	// HR 类告警 publisher 显式填 item.HeartRate；RR 类告警 publisher 显式填 item.RespiratoryRate。
	publishAlarm := func(category string, hr, rr int, weakSignal int64, eventReason string) error {
		item := observation.NewEventItem(ts, "instant")
		item.EventReason = eventReason
		item.TrackID = observation.TrackUnknownPerson
		if hr >= 0 {
			item.HeartRate = hr
		}
		if rr >= 0 {
			item.RespiratoryRate = rr
		}
		alarmData, _ := observation.EventItemToDataMap(&item)
		if alarmData == nil {
			alarmData = make(map[string]interface{})
		}
		if weakSignal > 0 {
			alarmData[observation.FieldWeakBiometricSignal] = weakSignal
		}
		alarmMsg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "alarm", category, alarmData)
		return c.streamPublisher.PublishAlarm(ctx, alarmMsg)
	}

	if breathState != 0 {
		rr := asInt(m["respiratory_rate"])
		if rr == 0 {
			rr = asInt(m["avg_respiratory_rate"])
		}
		avgBreathe := asInt(m["avg_respiratory_rate"])
		reason := fmt.Sprintf("avgBreathe=%d", avgBreathe)
		switch breathState {
		case 1:
			_ = publishAlarm(alarm.RespRateAlertHigh, -1, rr, 0, reason)
		case 2:
			_ = publishAlarm(alarm.RespRateAlertLow, -1, rr, 0, reason)
		case 3:
			_ = publishAlarm(alarm.ApneaHypopnea, -1, rr, 0, reason)
		}
	}
	if heartState == 1 {
		hr := asInt(m["heart_rate"])
		if hr == 0 {
			hr = asInt(m["avg_heart_rate"])
		}
		avgHeart := asInt(m["avg_heart_rate"])
		_ = publishAlarm(alarm.HeartRateAlertLow, hr, -1, 0, fmt.Sprintf("avgHeart=%d", avgHeart))
	} else if heartState == 2 {
		hr := asInt(m["heart_rate"])
		if hr == 0 {
			hr = asInt(m["avg_heart_rate"])
		}
		avgHeart := asInt(m["avg_heart_rate"])
		_ = publishAlarm(alarm.HeartRateAlertHigh, hr, -1, 0, fmt.Sprintf("avgHeart=%d", avgHeart))
	}
	if vitalSignsState == 3 {
		_ = publishAlarm(alarm.WeakBiometricSignal, -1, -1, int64(vitalSignsState)*20, alarm.WeakBiometricSignal)
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
	msg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "event", cat, data)
	return c.streamPublisher.PublishEvent(ctx, msg)
}

// handleEventMessage 处理事件/告警消息（event + alarm topic 统一入口），payload 符合 EventItem 格式。
// !canMonitor: drop event/alarm data, only send heart。
// canMonitor: decoder 输出 event_name（如 InBed、Fall），按 event_type 分配 track_id，流使用 device_uid。
func (c *MQTTConsumer) handleEventMessage(uid string, message map[string]interface{}) error {
	ctx := context.Background()
	_, _, _, cid, _, _, _, addr, ok := c.resolveDeviceIdentity(ctx, uid)
	if !ok {
		return nil
	}
	canIoT, canMonitor, _ := c.resolveIotPolicy(ctx, uid)
	if !canIoT {
		return nil
	}
	// card_id comes from resolveDeviceIdentity via CardMappingService
	dataValue, err := decode.RadarDecoder(message, "event")
	if err != nil {
		c.logger.Warn("event decode", zap.String("device", uid), zap.Error(err))
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

	if len(items) == 0 {
		return nil
	}

	// !canMonitor: drop event/alarm data, only send heart
	if !canMonitor {
		return c.publishRadarMonitorHeartbeat(ctx, addr, cid, ts)
	}

	// canMonitor: process event/alarm data normally
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

		item := observation.NewEventItem(ts, "start")
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
			// 4 个 pose-class category 各自保留（不再合并 SuspectedFall→Fall）：
			//   - SuspectedFall          (pose=2)  firmware 第一帧检测 → 下游 WARNING
			//   - Fall                   (pose=5)  firmware 30~90s qualification 通过 → 下游 ALERT/CRITICAL
			//   - SuspectedSittingOnGround (pose=7) 同理 WARNING
			//   - SittingOnGround        (pose=8)  同理 ALERT
			// 都走 event stream，由 sensor verifier 接管 → alarm_back_channel → cardagg 落库。
			// 见 doc/cardagg_sensor_split.md + memory firmware_fall_qualification。
			alarmCat := ""
			switch eventName {
			case alarm.Fall, alarm.SuspectedFall, alarm.SittingOnGround, alarm.SuspectedSittingOnGround:
				alarmCat = eventName
			}
			if alarmCat != "" {
				// Plan B：业务字段 first-class 平铺（TrackID / Pose 都是 EventItem 字段）。
				eventItem := observation.NewEventItem(ts, "start")
				eventItem.TrackID = asInt(m["track_id"])
				eventItem.Pose = asInt(m["pose"])
				eventData, _ := observation.EventItemToDataMap(&eventItem)
				if eventData == nil {
					eventData = make(map[string]interface{})
				}
				// topic_type=event；category 保留 alarm.Fall / alarm.SittingOnGround 让 sensor 路由识别
				evMsg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "event", alarmCat, eventData)
				if err := c.streamPublisher.PublishEvent(ctx, evMsg); err != nil {
					c.logger.Warn("publish fall event", zap.String("device", uid), zap.String("cat", alarmCat), zap.Error(err))
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

		msg := rediscommon.NewSingleItemMessage(addr, cid, DeviceTypeRadar, ts, "event", eventName, data)
		if err := c.streamPublisher.PublishEvent(ctx, msg); err != nil {
			c.logger.Warn("publish event", zap.String("device", uid), zap.String("cat", eventName), zap.Error(err))
			lastErr = err
		}
	}
	return lastErr
}

// DEPRECATED: publishSingleTrack — 已被 observation.Message 模式替代。
// func (c *MQTTConsumer) publishSingleTrack(...) error { ... }

// DEPRECATED: eventDataCategory — 已被 handleEventMessage 内联替代。
// func eventDataCategory(track interface{}) string { ... }
