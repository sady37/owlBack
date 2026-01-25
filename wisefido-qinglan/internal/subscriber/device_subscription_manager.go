package subscriber

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/publisher"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DeviceSubscription 设备订阅信息
type DeviceSubscription struct {
	DeviceUID      string
	DeviceID       string
	PropTopic      string
	MonitorTopic   string
	FuncTopic      string
	StatTopic      string
	EventTopic     string
	AlarmTopic     string
	MonitorSubTime time.Time // monitor订阅时间
	LastSeen       time.Time // 最后收到stat的时间
	Status         string    // online/offline/unsubscribed
	mu             sync.RWMutex
}

// MessageHandler MQTT消息处理器接口
type MessageHandler func(topic string, payload []byte) error

// DeviceSubscriptionManager 设备订阅管理器
type DeviceSubscriptionManager struct {
	config                   *config.Config
	mqttClient               *mqtt.Client
	mqttPublisher            *publisher.MQTTPublisher
	mqttConsumer             *consumer.MQTTConsumer    // MQTT消费者，用于订阅/取消订阅设备主题
	streamPublisher          *consumer.StreamPublisher // Redis Stream发布器，用于发布设备在线状态
	db                       *sql.DB
	redisClient              *redis.Client // Redis客户端，用于共享设备在线状态
	logger                   *zap.Logger
	messageHandler           MessageHandler                 // MQTT消息处理器（来自MQTT consumer）
	subscriptions            map[string]*DeviceSubscription // device_uid -> subscription
	unsubscribedDueToTimeout map[string]struct{}            // 90s 超时后强制取消订阅，需重认证
	mu                       sync.RWMutex
	checkInterval            time.Duration // 检查间隔（30秒）
	offlineTimeout           time.Duration // 离线超时（90秒）
	monitorMaxAge            time.Duration // monitor订阅最大时长（1小时）
	defaultContent           int           // 默认订阅内容：0-同时订阅，1-轨迹，2-呼吸心率
	defaultDuration          int           // 默认订阅时长（秒），默认 3600
	stopChan                 chan struct{}
	wg                       sync.WaitGroup
}

// NewDeviceSubscriptionManager 创建设备订阅管理器
func NewDeviceSubscriptionManager(
	cfg *config.Config,
	mqttClient *mqtt.Client,
	db *sql.DB,
	logger *zap.Logger,
	messageHandler MessageHandler, // MQTT消息处理器
) *DeviceSubscriptionManager {
	mqttPublisher := publisher.NewMQTTPublisher(cfg, mqttClient)
	return &DeviceSubscriptionManager{
		config:                   cfg,
		mqttClient:               mqttClient,
		mqttPublisher:            mqttPublisher,
		mqttConsumer:             nil, // 稍后通过SetMQTTConsumer设置
		db:                       db,
		redisClient:              nil, // 稍后通过SetRedisClient设置
		logger:                   logger,
		subscriptions:            make(map[string]*DeviceSubscription),
		unsubscribedDueToTimeout: make(map[string]struct{}),
		checkInterval:            90 * time.Second, // 每90秒检查一次
		offlineTimeout:           90 * time.Second,
		monitorMaxAge:            1 * time.Hour,
		defaultContent:           0,    // 0-同时订阅（轨迹和呼吸心率）
		defaultDuration:          3600, // 默认1小时
		stopChan:                 make(chan struct{}),
	}
}

// SetRedisClient 设置Redis客户端（用于共享设备在线状态）
func (m *DeviceSubscriptionManager) SetRedisClient(redisClient *redis.Client) {
	m.redisClient = redisClient
}

// SetStreamPublisher 设置Redis Stream发布器（用于发布设备在线状态到config stream）
func (m *DeviceSubscriptionManager) SetStreamPublisher(streamPublisher *consumer.StreamPublisher) {
	m.streamPublisher = streamPublisher
}

// SetMQTTConsumer 设置MQTT消费者（用于订阅/取消订阅设备主题）
func (m *DeviceSubscriptionManager) SetMQTTConsumer(mqttConsumer *consumer.MQTTConsumer) {
	m.mqttConsumer = mqttConsumer
}

// Start 启动订阅管理器
func (m *DeviceSubscriptionManager) Start(ctx context.Context) error {
	log.Println("Starting device subscription manager...")

	// 注意：不自动恢复已认证设备的MQTT订阅
	// 原因：
	// 1. 设备侧有自己的重连逻辑：如果 MQTT 连接失败，设备会重试；如果 10 分钟内无法连接，设备会重启
	// 2. 设备会重新进行 HTTPS 认证，认证成功后服务端会订阅设备的 MQTT 主题
	// 3. 服务重启时无法判断设备是什么时候断线的，如果设备已经断线很久，恢复订阅没有意义
	// 4. 如果设备刚断线，设备会自己重连并重新认证，认证成功后服务端再订阅
	// 因此，依赖设备重新认证来触发订阅，而不是在服务重启时自动恢复

	// 启动心跳监测goroutine
	m.wg.Add(1)
	go m.heartbeatMonitor(ctx)

	// 启动订阅续期goroutine
	m.wg.Add(1)
	go m.subscriptionRenewal(ctx)

	log.Println("Device subscription manager started")
	return nil
}

// restoreAuthenticatedDeviceSubscriptions 恢复已认证设备的MQTT订阅
// 服务重启后，从DB查询所有已认证的 Radar 设备（allow_access = TRUE AND device_type = 'Radar'），重新订阅并发送monitor命令
// 注意：wisefido-qinglan 只管理 Radar 设备，不管理 Sleepace 设备
// 简化逻辑：不检查内存中的状态记录，直接从DB重新订阅所有已认证的 Radar 设备
func (m *DeviceSubscriptionManager) restoreAuthenticatedDeviceSubscriptions(ctx context.Context) {
	log.Println("Restoring MQTT subscriptions for authenticated Radar devices from database...")

	// 查询所有已认证的 Radar 设备（allow_access = TRUE AND device_type = 'Radar'）
	// 注意：wisefido-qinglan 只管理 Radar 设备，Sleepace 设备由其他服务管理
	query := `
		SELECT device_uid, device_id
		FROM device_store
		WHERE allow_access = TRUE
		  AND device_type = 'Radar'
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		m.logger.Error("Failed to query authenticated devices",
			zap.Error(err),
		)
		log.Printf("❌ Failed to query authenticated devices: %v", err)
		return
	}
	defer rows.Close()

	restoredCount := 0
	for rows.Next() {
		var deviceUID, deviceID string
		if err := rows.Scan(&deviceUID, &deviceID); err != nil {
			m.logger.Warn("Failed to scan device row",
				zap.Error(err),
			)
			continue
		}

		// 创建订阅记录并订阅设备的6个MQTT主题；收到设备第一笔 MQTT 数据时由 UpdateLastSeen 发送 monitor 命令
		go func(uid, id string) {
			m.mu.Lock()
			now := time.Now()
			m.subscriptions[uid] = &DeviceSubscription{
				DeviceUID:      uid,
				DeviceID:       id,
				PropTopic:      m.mqttClient.BuildTopic("prop", uid),
				MonitorTopic:   m.mqttClient.BuildTopic("monitor", uid),
				FuncTopic:      m.mqttClient.BuildTopic("func", uid),
				StatTopic:      m.mqttClient.BuildTopic("stat", uid),
				EventTopic:     m.mqttClient.BuildTopic("event", uid),
				AlarmTopic:     m.mqttClient.BuildTopic("alarm", uid),
				MonitorSubTime: now,
				LastSeen:       time.Time{}, // 初始化为零值，等待设备发送第一条消息
				Status:         "online",    // 初始状态为online
			}
			m.mu.Unlock()

			// 订阅设备的6个MQTT主题（不使用通配符，仅订阅该设备的主题）
			if m.mqttConsumer != nil {
				if err := m.mqttConsumer.SubscribeDeviceTopics(uid); err != nil {
					m.logger.Warn("Failed to restore device MQTT topics subscription",
						zap.String("device_uid", uid),
						zap.Error(err),
					)
					log.Printf("⚠️ Failed to restore MQTT topics subscription for device %s: %v", uid, err)
				} else {
					m.logger.Info("Restored device MQTT topics subscription",
						zap.String("device_uid", uid),
						zap.String("device_id", id),
					)
					log.Printf("✅ Restored MQTT topics subscription for device %s", uid)
				}
			}

			m.logger.Info("Restored subscription record for device (monitor on first message)",
				zap.String("device_uid", uid),
				zap.String("device_id", id),
			)
		}(deviceUID, deviceID)

		restoredCount++
	}

	if err := rows.Err(); err != nil {
		m.logger.Error("Error iterating authenticated devices",
			zap.Error(err),
		)
		log.Printf("❌ Error iterating authenticated devices: %v", err)
		return
	}

	log.Printf("✅ Restored subscription records for %d authenticated Radar devices (monitor command on first MQTT message)", restoredCount)
	m.logger.Info("Restoring MQTT subscriptions for authenticated Radar devices",
		zap.Int("count", restoredCount),
	)
}

// Stop 停止订阅管理器
func (m *DeviceSubscriptionManager) Stop(ctx context.Context) error {
	log.Println("Stopping device subscription manager...")
	close(m.stopChan)
	m.wg.Wait()
	log.Println("Device subscription manager stopped")
	return nil
}

// SubscribeDevice 订阅设备的实时数据（认证成功后调用）
// 通过向设备发送MQTT命令来订阅，而不是订阅MQTT主题
func (m *DeviceSubscriptionManager) SubscribeDevice(ctx context.Context, deviceUID, deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已订阅
	if sub, exists := m.subscriptions[deviceUID]; exists {
		// 检查订阅是否即将过期（在过期前10分钟续订）
		sub.mu.RLock()
		monitorSubTime := sub.MonitorSubTime
		sub.mu.RUnlock()

		// 如果订阅还在有效期内（距离过期还有10分钟以上），不需要重新订阅
		if time.Since(monitorSubTime) < m.monitorMaxAge-10*time.Minute {
			// 更新订阅信息
			sub.mu.Lock()
			sub.DeviceID = deviceID
			sub.LastSeen = time.Now()
			sub.mu.Unlock()
			log.Printf("Device %s already subscribed (valid until %v), updating subscription info", deviceUID, monitorSubTime.Add(m.monitorMaxAge))

			// 注意：不在这里发送monitor订阅命令
			// 参考wisefido-radar的实现：在收到设备的第一条MQTT消息时再发送（见UpdateLastSeen方法）
			// 这样可以确保设备已经连接MQTT并准备好接收命令
			return nil
		}
		// 订阅即将过期，需要续订
		log.Printf("Device %s subscription expiring soon, renewing...", deviceUID)
	}

	// 注意：不在这里发送monitor订阅命令
	// 参考wisefido-radar的实现：在收到设备的第一条MQTT消息时再发送（见UpdateLastSeen方法）
	// 根据协议文档，设备在MQTT连接成功后等待5秒才关闭蓝牙，所以延迟2秒可能不够
	// 改为在收到设备的第一条MQTT消息时再发送，确保设备已经连接MQTT并准备好接收命令

	// 订阅设备的6个MQTT主题（不使用通配符，仅订阅该设备的主题）
	log.Printf("🔍 Checking mqttConsumer for device %s: %v", deviceUID, m.mqttConsumer != nil)
	if m.mqttConsumer != nil {
		log.Printf("🔧 About to call SubscribeDeviceTopics for device %s", deviceUID)
		// 检查 MQTT 连接状态，如果未连接则等待重试
		maxRetries := 5
		retryDelay := 500 * time.Millisecond
		var subscribeErr error
		for i := 0; i < maxRetries; i++ {
			log.Printf("🔄 Calling SubscribeDeviceTopics for device %s (attempt %d/%d)", deviceUID, i+1, maxRetries)
			if err := m.mqttConsumer.SubscribeDeviceTopics(deviceUID); err != nil {
				subscribeErr = err
				if i < maxRetries-1 {
					m.logger.Info("MQTT subscription failed, retrying",
						zap.String("device_uid", deviceUID),
						zap.Int("retry", i+1),
						zap.Int("max_retries", maxRetries),
						zap.Error(err),
					)
					log.Printf("⚠️ MQTT subscription failed for device %s, retrying (%d/%d): %v", deviceUID, i+1, maxRetries, err)
					time.Sleep(retryDelay)
					continue
				}
			} else {
				subscribeErr = nil
				break
			}
		}

		if subscribeErr != nil {
			m.logger.Warn("Failed to subscribe device MQTT topics after retries",
				zap.String("device_uid", deviceUID),
				zap.Int("retries", maxRetries),
				zap.Error(subscribeErr),
			)
			log.Printf("⚠️ Failed to subscribe MQTT topics for device %s after %d retries: %v", deviceUID, maxRetries, subscribeErr)
		} else {
			m.logger.Info("Subscribed device MQTT topics",
				zap.String("device_uid", deviceUID),
			)
			log.Printf("✅ Subscribed MQTT topics for device %s", deviceUID)
		}
	}

	// 构建主题字符串（用于记录，不实际订阅）
	// 根据协议文档 3.3.3，设备发布的主题有6个：prop, monitor, func, stat, event, alarm
	propTopic := m.mqttClient.BuildTopic("prop", deviceUID)
	monitorTopic := m.mqttClient.BuildTopic("monitor", deviceUID)
	funcTopic := m.mqttClient.BuildTopic("func", deviceUID)
	statTopic := m.mqttClient.BuildTopic("stat", deviceUID)
	eventTopic := m.mqttClient.BuildTopic("event", deviceUID)
	alarmTopic := m.mqttClient.BuildTopic("alarm", deviceUID)

	// 创建或更新订阅记录
	// 注意：LastSeen初始化为零值，表示还没有收到设备消息
	now := time.Now()
	sub := &DeviceSubscription{
		DeviceUID:      deviceUID,
		DeviceID:       deviceID,
		PropTopic:      propTopic,
		MonitorTopic:   monitorTopic,
		FuncTopic:      funcTopic,
		StatTopic:      statTopic,
		EventTopic:     eventTopic,
		AlarmTopic:     alarmTopic,
		MonitorSubTime: now,
		LastSeen:       time.Time{}, // 初始化为零值，表示还没有收到设备消息
		Status:         "online",    // 初始状态为online
	}
	m.subscriptions[deviceUID] = sub

	// 不在此处更新 DB 为 online；收到第一笔数据时在 UpdateLastSeen 中更新

	m.logger.Info("Device subscription recorded, sending monitor command",
		zap.String("device_uid", deviceUID),
		zap.String("device_id", deviceID),
		zap.Time("monitor_sub_time", now),
		zap.Int("duration", m.defaultDuration),
		zap.Int("content", m.defaultContent),
	)
	log.Printf("✅ Device subscription recorded for %s (device_id: %s). Sending monitor command.", deviceUID, deviceID)

	// 立即发送 monitor 订阅命令（设备已经收到认证响应，应该很快会连接 MQTT）
	// 使用 goroutine 异步发送，避免阻塞认证流程
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
			m.logger.Warn("Failed to send monitor subscription command after device subscription",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			log.Printf("⚠️ Failed to send monitor subscription command to device %s: %v", deviceUID, err)
		} else {
			m.logger.Info("Sent monitor subscription command after device subscription",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
			)
			log.Printf("✅ Sent monitor subscription command to device %s", deviceUID)
		}
	}()

	return nil
}

// EnablePeriodicSubscription 开启周期性订阅（认证成功后调用）
// 检查是否已订阅MQTT主题：已订阅则仅记录，否则开启订阅
func (m *DeviceSubscriptionManager) EnablePeriodicSubscription(ctx context.Context, deviceUID, deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在订阅管理器中存在记录
	sub, exists := m.subscriptions[deviceUID]
	if exists {
		// 已存在订阅记录，仅更新信息
		sub.mu.Lock()
		sub.DeviceID = deviceID
		sub.LastSeen = time.Now()
		sub.mu.Unlock()
		log.Printf("Device %s already has subscription record, updating subscription info only", deviceUID)
		m.logger.Info("Device subscription record updated",
			zap.String("device_uid", deviceUID),
			zap.String("device_id", deviceID),
		)
		return nil
	}

	// 检查是否已订阅MQTT主题
	isMQTTSubscribed := false
	if m.mqttConsumer != nil {
		isMQTTSubscribed = m.mqttConsumer.IsDeviceTopicsSubscribed(deviceUID)
	}

	// 如果未订阅MQTT主题，则订阅
	if !isMQTTSubscribed {
		if m.mqttConsumer != nil {
			log.Printf("Device %s MQTT topics not subscribed, subscribing now...", deviceUID)
			if err := m.mqttConsumer.SubscribeDeviceTopics(deviceUID); err != nil {
				m.logger.Warn("Failed to subscribe device MQTT topics",
					zap.String("device_uid", deviceUID),
					zap.Error(err),
				)
				log.Printf("⚠️ Failed to subscribe MQTT topics for device %s: %v", deviceUID, err)
				// 继续创建订阅记录，即使订阅失败
			} else {
				m.logger.Info("Subscribed device MQTT topics",
					zap.String("device_uid", deviceUID),
				)
				log.Printf("✅ Subscribed MQTT topics for device %s", deviceUID)
			}
		}
	} else {
		log.Printf("Device %s MQTT topics already subscribed, skipping subscription", deviceUID)
		m.logger.Info("Device MQTT topics already subscribed, skipping",
			zap.String("device_uid", deviceUID),
		)
	}

	// 构建主题字符串（用于记录）
	propTopic := m.mqttClient.BuildTopic("prop", deviceUID)
	monitorTopic := m.mqttClient.BuildTopic("monitor", deviceUID)
	funcTopic := m.mqttClient.BuildTopic("func", deviceUID)
	statTopic := m.mqttClient.BuildTopic("stat", deviceUID)
	eventTopic := m.mqttClient.BuildTopic("event", deviceUID)
	alarmTopic := m.mqttClient.BuildTopic("alarm", deviceUID)

	// 创建订阅记录
	now := time.Now()
	sub = &DeviceSubscription{
		DeviceUID:      deviceUID,
		DeviceID:       deviceID,
		PropTopic:      propTopic,
		MonitorTopic:   monitorTopic,
		FuncTopic:      funcTopic,
		StatTopic:      statTopic,
		EventTopic:     eventTopic,
		AlarmTopic:     alarmTopic,
		MonitorSubTime: now,
		LastSeen:       time.Time{}, // 初始化为零值，表示还没有收到设备消息
		Status:         "online",    // 初始状态为online
	}
	m.subscriptions[deviceUID] = sub

	m.logger.Info("Periodic subscription enabled, sending monitor command",
		zap.String("device_uid", deviceUID),
		zap.String("device_id", deviceID),
		zap.Bool("mqtt_already_subscribed", isMQTTSubscribed),
		zap.Time("monitor_sub_time", now),
		zap.Int("duration", m.defaultDuration),
		zap.Int("content", m.defaultContent),
	)
	log.Printf("✅ Periodic subscription enabled for %s (device_id: %s, MQTT already subscribed: %v). Sending monitor command.", deviceUID, deviceID, isMQTTSubscribed)

	// 发送 monitor 订阅命令（设备已经收到认证响应，应该很快会连接 MQTT）
	// 使用 goroutine 异步发送，避免阻塞认证流程
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
			m.logger.Warn("Failed to send monitor subscription command after enabling periodic subscription",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			log.Printf("⚠️ Failed to send monitor subscription command to device %s: %v", deviceUID, err)
		} else {
			m.logger.Info("Sent monitor subscription command after enabling periodic subscription",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
			)
			log.Printf("✅ Sent monitor subscription command to device %s", deviceUID)
		}
	}()

	return nil
}

// UpdateLastSeen 更新设备最后收到消息的时间（MQTT consumer收到消息时调用）
// 如果设备还没有发送monitor订阅命令，在收到第一条消息时自动发送
func (m *DeviceSubscriptionManager) UpdateLastSeen(deviceUID string) {
	m.mu.Lock()
	sub, exists := m.subscriptions[deviceUID]
	if !exists {
		// 设备可能已被取消订阅，需要重新认证
		m.mu.Unlock()
		// 设备未在订阅列表中，可能是首次收到消息
		// 需要查询device_id并发送monitor订阅命令
		go m.autoSubscribeOnFirstMessage(context.Background(), deviceUID)
		return
	}

	now := time.Now()
	oldStatus := sub.Status

	sub.mu.Lock()
	lastSeen := sub.LastSeen
	sub.LastSeen = now

	// 状态恢复逻辑
	switch oldStatus {
	case "offline":
		// offline状态收到stat，如果now - lastSeen不超过180秒，恢复为online
		// 注意：此时LastSeen已经更新为now，所以需要检查更新前的时间
		timeSinceLastSeen := now.Sub(lastSeen)
		if timeSinceLastSeen <= 180*time.Second {
			sub.Status = "online"
			sub.mu.Unlock()

			log.Printf("✅ 设备 %s 从offline恢复为online", deviceUID)
			m.logger.Info("Device recovered from offline to online",
				zap.String("device_uid", deviceUID),
			)
			m.triggerStatusChange(deviceUID, "offline", "online")
		} else {
			// 超过180秒，应该已经被取消订阅了，这里不应该发生
			sub.mu.Unlock()
		}

	case "unsubscribed":
		// 已取消订阅的设备收到stat，需要重新认证
		sub.mu.Unlock()
		log.Printf("⚠️ 已取消订阅的设备 %s 收到stat，需要重新认证", deviceUID)
		m.logger.Warn("Unsubscribed device received stat, requires re-authentication",
			zap.String("device_uid", deviceUID),
		)

	default:
		// online状态，正常更新
		sub.mu.Unlock()
		// 设备在线，发布到 config stream
		// 注意：只在收到 stat 消息时发布（避免频繁发布）
		// 但 UpdateLastSeen 会被所有消息类型调用，所以需要检查消息类型
		// 这里简化：每次 UpdateLastSeen 都发布（因为已经有去重机制）
		m.publishDeviceOnlineStatusToConfigStream(context.Background(), deviceUID, "online")
	}

	m.mu.Unlock()

	// 如果这是设备首次收到消息（LastSeen为零值），发送monitor订阅命令并更新状态为online
	if lastSeen.IsZero() {
		log.Printf("Device %s first MQTT message received, sending monitor subscription command", deviceUID)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
				m.logger.Warn("Failed to send monitor subscription command on first message",
					zap.String("device_uid", deviceUID),
					zap.Error(err),
				)
			} else {
				m.logger.Info("Sent monitor subscription command on first message",
					zap.String("device_uid", deviceUID),
				)
				log.Printf("✅ Sent monitor subscription command to device %s on first message", deviceUID)
			}
		}()
	}

	// DB status 不存 on/offline；on/offline 仅存内存（LastSeen + 90s 超时）
}

// autoSubscribeOnFirstMessage 在收到设备第一条消息时自动发送monitor订阅命令
func (m *DeviceSubscriptionManager) autoSubscribeOnFirstMessage(ctx context.Context, deviceUID string) {
	// 查询device_id（只查询 Radar 设备）
	query := `SELECT device_id FROM device_store WHERE device_uid = $1 AND allow_access = TRUE AND device_type = 'Radar'`
	var deviceID string
	err := m.db.QueryRowContext(ctx, query, deviceUID).Scan(&deviceID)
	if err != nil {
		m.logger.Warn("Failed to get device_id for auto-subscribe",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return
	}

	// 订阅设备的6个MQTT主题（不使用通配符，仅订阅该设备的主题）
	if m.mqttConsumer != nil {
		if err := m.mqttConsumer.SubscribeDeviceTopics(deviceUID); err != nil {
			m.logger.Warn("Failed to subscribe device MQTT topics on first message",
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
			log.Printf("⚠️ Failed to subscribe MQTT topics for device %s on first message: %v", deviceUID, err)
		} else {
			m.logger.Info("Subscribed device MQTT topics on first message",
				zap.String("device_uid", deviceUID),
			)
			log.Printf("✅ Subscribed MQTT topics for device %s on first message", deviceUID)
		}
	}

	// 如果设备不在订阅列表中，创建订阅记录
	m.mu.Lock()
	_, exists := m.subscriptions[deviceUID]
	if !exists {
		// 创建订阅记录（不实际订阅，因为通配符订阅已经覆盖）
		now := time.Now()
		sub := &DeviceSubscription{
			DeviceUID:      deviceUID,
			DeviceID:       deviceID,
			PropTopic:      m.mqttClient.BuildTopic("prop", deviceUID),
			MonitorTopic:   m.mqttClient.BuildTopic("monitor", deviceUID),
			FuncTopic:      m.mqttClient.BuildTopic("func", deviceUID),
			StatTopic:      m.mqttClient.BuildTopic("stat", deviceUID),
			EventTopic:     m.mqttClient.BuildTopic("event", deviceUID),
			AlarmTopic:     m.mqttClient.BuildTopic("alarm", deviceUID),
			MonitorSubTime: now,
			LastSeen:       time.Time{}, // 初始化为零值
			Status:         "online",    // 初始状态为online
		}
		m.subscriptions[deviceUID] = sub
		m.logger.Info("Created subscription record for device on first message",
			zap.String("device_uid", deviceUID),
			zap.String("device_id", deviceID),
		)
	}
	m.mu.Unlock()

	// 发送monitor订阅命令（设备已经连接MQTT，命令不会丢失）
	if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
		m.logger.Warn("Failed to send monitor subscription command on first message",
			zap.String("device_uid", deviceUID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
	} else {
		m.logger.Info("Auto-subscribed on device first message",
			zap.String("device_uid", deviceUID),
			zap.String("device_id", deviceID),
		)
		log.Printf("✅ Auto-sent monitor subscription command to device %s on first message", deviceUID)
	}
}

// heartbeatMonitor 心跳监测goroutine（每90秒检查一次）
func (m *DeviceSubscriptionManager) heartbeatMonitor(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkDeviceHeartbeat(ctx)
		}
	}
}

// checkDeviceHeartbeat 检查设备心跳（每90秒检查一次）
func (m *DeviceSubscriptionManager) checkDeviceHeartbeat(ctx context.Context) {
	m.mu.RLock()
	subs := make([]*DeviceSubscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subs = append(subs, sub)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, sub := range subs {
		sub.mu.RLock()
		lastSeen := sub.LastSeen
		status := sub.Status
		deviceUID := sub.DeviceUID
		deviceID := sub.DeviceID
		sub.mu.RUnlock()

		// 跳过还未收到过stat的设备
		if lastSeen.IsZero() {
			continue
		}

		// disabled 设备：从 DB 查 status，立即从订阅中移除
		dbStatus, err := m.getDeviceStatus(ctx, deviceID)
		if err != nil {
			m.logger.Debug("Failed to get device status for heartbeat check",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			continue
		}
		if dbStatus == "disabled" {
			m.logger.Info("Device disabled, removing from subscriptions",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
			)
			m.mu.Lock()
			delete(m.subscriptions, deviceUID)
			m.mu.Unlock()
			continue
		}

		timeSinceLastSeen := now.Sub(lastSeen)

		switch status {
		case "online":
			// 阶段1：90秒无stat → 标记offline
			if timeSinceLastSeen > 90*time.Second {
				m.markDeviceOffline(deviceUID, now)
			}

		case "offline":
			// 阶段2：180秒无stat → 取消订阅
			if timeSinceLastSeen > 180*time.Second {
				m.unsubscribeDevice(deviceUID)
			}

		case "unsubscribed":
			// 阶段3：已取消订阅的设备，保持状态
			// 需要重新认证才能恢复
		}
	}
}

// markDeviceOffline 第一阶段：标记设备为offline
func (m *DeviceSubscriptionManager) markDeviceOffline(deviceUID string, _ time.Time) {
	m.mu.Lock()
	sub, exists := m.subscriptions[deviceUID]
	m.mu.Unlock()

	if !exists {
		return
	}

	sub.mu.Lock()
	if sub.Status == "online" {
		sub.Status = "offline"
		sub.mu.Unlock()

		log.Printf("⚠️ 设备 %s 标记为offline（90秒无stat）", deviceUID)
		m.logger.Info("Device marked as offline (90s no stat)",
			zap.String("device_uid", deviceUID),
		)

		// 触发offline事件，但不取消订阅
		m.triggerStatusChange(deviceUID, "online", "offline")
	} else {
		sub.mu.Unlock()
	}
}

// unsubscribeDevice 第二阶段：取消订阅设备
func (m *DeviceSubscriptionManager) unsubscribeDevice(deviceUID string) {
	m.mu.Lock()
	sub, exists := m.subscriptions[deviceUID]
	m.mu.Unlock()

	if !exists {
		return
	}

	sub.mu.Lock()
	if sub.Status == "offline" {
		sub.Status = "unsubscribed"
		sub.mu.Unlock()

		log.Printf("🚫 设备 %s 取消订阅（180秒无stat）", deviceUID)
		m.logger.Info("Device unsubscribed (180s no stat)",
			zap.String("device_uid", deviceUID),
		)

		// 1. 取消MQTT订阅
		if m.mqttConsumer != nil {
			go func() {
				m.mqttConsumer.UnsubscribeDeviceTopics(deviceUID)
				log.Printf("✅ 已取消设备 %s 的MQTT订阅", deviceUID)
				m.logger.Info("Unsubscribed device MQTT topics",
					zap.String("device_uid", deviceUID),
				)
			}()
		}

		// 2. 从订阅列表中移除并标记需要重认证
		m.mu.Lock()
		delete(m.subscriptions, deviceUID)
		m.unsubscribedDueToTimeout[deviceUID] = struct{}{}
		m.mu.Unlock()

		// 3. 触发取消订阅事件
		m.triggerUnsubscribeEvent(deviceUID)
	} else {
		sub.mu.Unlock()
	}
}

// triggerStatusChange 触发状态变更事件
func (m *DeviceSubscriptionManager) triggerStatusChange(deviceUID, oldStatus, newStatus string) {
	m.logger.Info("Device status changed",
		zap.String("device_uid", deviceUID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", newStatus),
	)

	// 发布设备在线状态到 config stream
	m.publishDeviceOnlineStatusToConfigStream(context.Background(), deviceUID, newStatus)
}

// publishDeviceOnlineStatusToConfigStream 发布设备在线状态到 config stream
// 使用统一的 config:device_status:stream 格式，Category 直接使用 online/offline/unsubscribed
func (m *DeviceSubscriptionManager) publishDeviceOnlineStatusToConfigStream(ctx context.Context, deviceUID, onlineStatus string) {
	if m.streamPublisher == nil {
		m.logger.Warn("StreamPublisher not set, skipping device online status publish",
			zap.String("device_uid", deviceUID),
		)
		return
	}

	// 查询设备信息（device_id, tenant_id, device_code, device_type）
	// 注意：只处理 Radar 设备
	var deviceID, tenantID, deviceCode, deviceType string
	query := `SELECT ds.device_id::text, COALESCE(d.tenant_id::text, '00000000-0000-0000-0000-000000000000'), COALESCE(ds.device_code, ''), COALESCE(ds.device_type, 'Radar')
			  FROM device_store ds
			  LEFT JOIN devices d ON ds.device_id = d.device_id
			  WHERE ds.device_uid = $1 AND ds.allow_access = TRUE AND ds.device_type = 'Radar'
			  LIMIT 1`
	err := m.db.QueryRowContext(ctx, query, deviceUID).Scan(&deviceID, &tenantID, &deviceCode, &deviceType)
	if err != nil {
		m.logger.Warn("Failed to get device info for online status publish",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return
	}

	// 如果 tenant_id 为空，使用默认值
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000000"
	}

	// 构建 ConfigChangeMessage
	msg := rediscommon.BuildDeviceOnlineStatusMessage("wisefido-qinglan", tenantID, deviceID, deviceUID, deviceCode, deviceType, onlineStatus)

	// 转换为 map（用于发布到 Redis Stream）
	data := configChangeMessageToMap(msg)

	// 发布到 config:device_status:stream
	streamName := rediscommon.StreamConfigDeviceStatus.Name
	_, err = m.streamPublisher.PublishToStream(ctx, streamName, data)
	if err != nil {
		m.logger.Warn("Failed to publish device online status to config stream",
			zap.String("device_uid", deviceUID),
			zap.String("status", onlineStatus),
			zap.Error(err),
		)
	}
}

// configChangeMessageToMap 将 ConfigChangeMessage 转换为 map（用于发布到 Redis Stream）
// 使用 CloudEvents 标准格式
func configChangeMessageToMap(msg rediscommon.ConfigChangeMessage) map[string]interface{} {
	result := make(map[string]interface{})

	// CloudEvents 标准字段
	result["specversion"] = msg.SpecVersion
	result["id"] = msg.ID
	result["source"] = msg.Source
	result["type"] = msg.Type
	result["time"] = msg.Time

	// data 为 JSON 字符串
	dataJSON, _ := json.Marshal(msg.Data)
	result["data"] = string(dataJSON)

	return result
}

// triggerUnsubscribeEvent 触发取消订阅事件
func (m *DeviceSubscriptionManager) triggerUnsubscribeEvent(deviceUID string) {
	m.logger.Info("Device unsubscribed event",
		zap.String("device_uid", deviceUID),
	)
	// 可以在这里添加其他取消订阅处理逻辑，如通知其他模块
}

// subscriptionRenewal 订阅续期goroutine（定期重新订阅monitor主题）
func (m *DeviceSubscriptionManager) subscriptionRenewal(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次是否需要续期
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.renewSubscriptions(ctx)
		}
	}
}

// renewSubscriptions 续期订阅（monitor订阅最长1小时）
func (m *DeviceSubscriptionManager) renewSubscriptions(ctx context.Context) {
	m.mu.RLock()
	subs := make([]*DeviceSubscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subs = append(subs, sub)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, sub := range subs {
		sub.mu.RLock()
		monitorSubTime := sub.MonitorSubTime
		deviceUID := sub.DeviceUID
		deviceID := sub.DeviceID
		sub.mu.RUnlock()

		// 检查monitor订阅是否超过最大时长（1小时），在过期前10分钟续订
		renewTime := monitorSubTime.Add(m.monitorMaxAge - 10*time.Minute)
		if now.After(renewTime) {
			// 重新发送monitor订阅命令给设备
			if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
				m.logger.Warn("Failed to renew monitor subscription command",
					zap.String("device_uid", deviceUID),
					zap.String("device_id", deviceID),
					zap.Error(err),
				)
			} else {
				sub.mu.Lock()
				sub.MonitorSubTime = now
				sub.mu.Unlock()
				m.logger.Info("Renewed monitor subscription command",
					zap.String("device_uid", deviceUID),
					zap.String("device_id", deviceID),
					zap.Duration("age", now.Sub(monitorSubTime)),
					zap.Int("duration", m.defaultDuration),
				)
				log.Printf("✅ Renewed monitor subscription command for device %s (device_id: %s)", deviceUID, deviceID)
			}
		}
	}
}

// getDeviceStatus 获取设备当前状态（仅用于 disabled 检查，DB status 不存 on/offline）
func (m *DeviceSubscriptionManager) getDeviceStatus(ctx context.Context, deviceID string) (string, error) {
	var status string
	query := `SELECT status FROM devices WHERE device_id = $1`
	err := m.db.QueryRowContext(ctx, query, deviceID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// IsForceUnsubscribed 是否因 90s 超时被强制取消订阅（需重认证）
func (m *DeviceSubscriptionManager) IsForceUnsubscribed(deviceUID string) bool {
	m.mu.RLock()
	_, ok := m.unsubscribedDueToTimeout[deviceUID]
	m.mu.RUnlock()
	return ok
}

// ClearForceUnsubscribed 认证成功后清除强制重认证标记，允许首条消息时再次订阅
func (m *DeviceSubscriptionManager) ClearForceUnsubscribed(deviceUID string) {
	m.mu.Lock()
	delete(m.unsubscribedDueToTimeout, deviceUID)
	m.mu.Unlock()
}

// GetDeviceOnlineStatus 获取设备的在线状态（online/offline/unsubscribed）
// 用于其他服务查询设备的实时在线状态
func (m *DeviceSubscriptionManager) GetDeviceOnlineStatus(deviceUID string) string {
	m.mu.RLock()
	sub, exists := m.subscriptions[deviceUID]
	m.mu.RUnlock()

	if !exists {
		// 设备不在订阅列表中，可能是未认证或已取消订阅
		// 检查是否因超时被取消订阅
		m.mu.RLock()
		_, unsubscribed := m.unsubscribedDueToTimeout[deviceUID]
		m.mu.RUnlock()
		if unsubscribed {
			return "unsubscribed"
		}
		return "offline" // 默认返回 offline
	}

	sub.mu.RLock()
	status := sub.Status
	sub.mu.RUnlock()

	return status
}

// GetDeviceOnlineStatusByDeviceID 通过 device_id 获取设备的在线状态
// 需要先通过 device_id 查询 device_uid
func (m *DeviceSubscriptionManager) GetDeviceOnlineStatusByDeviceID(ctx context.Context, deviceID string) (string, error) {
	// 查询 device_uid
	query := `SELECT device_uid FROM device_store WHERE device_id = $1`
	var deviceUID string
	err := m.db.QueryRowContext(ctx, query, deviceID).Scan(&deviceUID)
	if err != nil {
		return "offline", err
	}

	return m.GetDeviceOnlineStatus(deviceUID), nil
}
