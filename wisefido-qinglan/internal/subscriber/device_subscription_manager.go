package subscriber

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/publisher"
	"wisefido-qinglan/internal/repository"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DeviceSubscription 设备订阅信息
type DeviceSubscription struct {
	DeviceUID                   string
	DeviceID                    string
	PropTopic                   string
	MonitorTopic                string
	FuncTopic                   string
	StatTopic                   string
	EventTopic                  string
	AlarmTopic                  string
	MonitorSubTime              time.Time // monitor订阅时间
	LastSeen                    time.Time // 最后收到消息的时间（取 monitor/stat/event/alarm 中最新的）
	LastMonitorTime             time.Time // 最后收到 monitor 消息的时间
	LastStatTime                time.Time // 最后收到 stat 消息的时间
	LastEventTime               time.Time // 最后收到 event 消息的时间
	LastAlarmTime               time.Time // 最后收到 alarm 消息的时间
	LastOnlineStatusPublishTime time.Time // 上次向 config stream 发布 online 的时间，用于节流
	Status                      string    // online/offline/unsubscribed
	// Location info (addressInfo) - stored during subscription initialization
	TenantID             string
	DeviceCode           string
	DeviceType           string
	BranchID             *string
	BuildingID           *string
	UnitID               *string
	RoomID               *string
	BedID                *string
	onlineStatusStopChan chan struct{} // 用于停止周期性 online 状态发布的定时器
	mu                   sync.RWMutex
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
	deviceRepo               repository.DeviceRepository // 设备仓库，用于查询设备位置信息
	redisClient              *redis.Client               // Redis客户端，用于共享设备在线状态
	logger                   *zap.Logger
	messageHandler           MessageHandler                 // MQTT消息处理器（来自MQTT consumer）
	subscriptionsByUID       map[string]*DeviceSubscription // device_uid -> subscription
	subscriptionsByID        map[string]*DeviceSubscription // device_id -> subscription，与 subscriptionsByUID 指向同一个对象
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
		deviceRepo:               repository.NewPostgresDeviceRepository(db), // 创建设备仓库实例
		redisClient:              nil,                                        // 稍后通过SetRedisClient设置
		logger:                   logger,
		subscriptionsByUID:       make(map[string]*DeviceSubscription),
		subscriptionsByID:        make(map[string]*DeviceSubscription),
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

	// 启动 1 分钟后检查 MQTT 队列、所有 func 并发布上线：由 mqtt_consumer.publishOnlineForConnectedAfterStartup 统一处理

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
			sub := &DeviceSubscription{
				DeviceUID:       uid,
				DeviceID:        id,
				PropTopic:       m.mqttClient.BuildTopic("prop", uid),
				MonitorTopic:    m.mqttClient.BuildTopic("monitor", uid),
				FuncTopic:       m.mqttClient.BuildTopic("func", uid),
				StatTopic:       m.mqttClient.BuildTopic("stat", uid),
				EventTopic:      m.mqttClient.BuildTopic("event", uid),
				AlarmTopic:      m.mqttClient.BuildTopic("alarm", uid),
				MonitorSubTime:  now,
				LastSeen:        time.Time{}, // 初始化为零值，等待设备发送第一条消息
				LastMonitorTime: time.Time{}, // 初始化为零值
				LastStatTime:    time.Time{}, // 初始化为零值
				LastEventTime:   time.Time{}, // 初始化为零值
				LastAlarmTime:   time.Time{}, // 初始化为零值
				Status:          "online",    // 初始状态为online
			}
			// 同时存储到两个 map
			m.subscriptionsByUID[uid] = sub
			if id != "" {
				m.subscriptionsByID[id] = sub
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
	if sub, exists := m.subscriptionsByUID[deviceUID]; exists {
		// 检查订阅是否即将过期（在过期前10分钟续订）
		sub.mu.RLock()
		monitorSubTime := sub.MonitorSubTime
		sub.mu.RUnlock()

		// 如果订阅还在有效期内（距离过期还有10分钟以上），不需要重新订阅
		if time.Since(monitorSubTime) < m.monitorMaxAge-10*time.Minute {
			// 更新订阅信息
			sub.mu.Lock()
			oldDeviceID := sub.DeviceID
			sub.DeviceID = deviceID
			sub.LastSeen = time.Now()
			sub.mu.Unlock()
			// 更新 subscriptionsByID（如果 deviceID 变化）
			if deviceID != "" && deviceID != oldDeviceID {
				m.mu.Lock()
				m.subscriptionsByID[deviceID] = sub
				if oldDeviceID != "" {
					delete(m.subscriptionsByID, oldDeviceID)
				}
				m.mu.Unlock()
			}
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
	if m.mqttConsumer != nil {
		// 检查 MQTT 连接状态，如果未连接则等待重试
		maxRetries := 5
		retryDelay := 500 * time.Millisecond
		var subscribeErr error
		for i := 0; i < maxRetries; i++ {
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

	// 查询设备 location info (addressInfo) 用于后续 alarm/event 消息
	// 使用 GetDeviceLocationInfo 方法，通过 device_uid 查询
	locationInfo, err := m.deviceRepo.GetDeviceLocationInfo(ctx, deviceUID)
	if err != nil {
		m.logger.Warn("Failed to get device location info during subscription",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		// 查询失败时返回错误，不创建订阅记录
		return fmt.Errorf("failed to get device location info: %w", err)
	}

	// 从 locationInfo 中提取字段
	tenantID := locationInfo.TenantID
	deviceCode := "" // GetDeviceLocationInfo 不返回 device_code，需要从 device_store 查询
	deviceType := ""
	if locationInfo.DeviceType.Valid {
		deviceType = locationInfo.DeviceType.String
	}
	var branchID, buildingID, unitID, roomID, bedID *string
	if locationInfo.BranchID.Valid {
		branchID = &locationInfo.BranchID.String
	}
	if locationInfo.BuildingID.Valid {
		buildingID = &locationInfo.BuildingID.String
	}
	if locationInfo.UnitID.Valid {
		unitID = &locationInfo.UnitID.String
	}
	if locationInfo.RoomID.Valid {
		roomID = &locationInfo.RoomID.String
	}
	if locationInfo.BedID.Valid {
		bedID = &locationInfo.BedID.String
	}

	// 查询 device_code（GetDeviceLocationInfo 不返回 device_code）
	deviceCodeQuery := `SELECT device_code FROM device_store WHERE device_uid = $1 AND allow_access = TRUE AND device_type = 'Radar' LIMIT 1`
	err = m.db.QueryRowContext(ctx, deviceCodeQuery, deviceUID).Scan(&deviceCode)
	if err != nil {
		m.logger.Warn("Failed to get device_code, using empty string",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		deviceCode = ""
	}

	// 创建或更新订阅记录
	// 注意：LastSeen初始化为零值，表示还没有收到设备消息
	now := time.Now()
	sub := &DeviceSubscription{
		DeviceUID:       deviceUID,
		DeviceID:        deviceID,
		PropTopic:       propTopic,
		MonitorTopic:    monitorTopic,
		FuncTopic:       funcTopic,
		StatTopic:       statTopic,
		EventTopic:      eventTopic,
		AlarmTopic:      alarmTopic,
		MonitorSubTime:  now,
		LastSeen:        time.Time{}, // 初始化为零值，表示还没有收到设备消息
		LastMonitorTime: time.Time{}, // 初始化为零值
		LastStatTime:    time.Time{}, // 初始化为零值
		LastEventTime:   time.Time{}, // 初始化为零值
		LastAlarmTime:   time.Time{}, // 初始化为零值
		Status:          "online",    // 初始状态为online
		// Location info (addressInfo)
		TenantID:   tenantID,
		DeviceCode: deviceCode,
		DeviceType: deviceType,
		BranchID:   branchID,
		BuildingID: buildingID,
		UnitID:     unitID,
		RoomID:     roomID,
		BedID:      bedID,
	}
	// 同时存储到两个 map
	m.subscriptionsByUID[deviceUID] = sub
	if deviceID != "" {
		m.subscriptionsByID[deviceID] = sub
	}

	// auth 成功后，立即发布 online event 到 alarm stream 并加入 online 列表
	go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "online")

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
	sub, exists := m.subscriptionsByUID[deviceUID]
	if exists {
		// 已存在订阅记录，仅更新信息
		sub.mu.Lock()
		oldDeviceID := sub.DeviceID
		sub.DeviceID = deviceID
		sub.LastSeen = time.Now()
		sub.mu.Unlock()
		// 更新 subscriptionsByID（如果 deviceID 变化）
		if deviceID != "" && deviceID != oldDeviceID {
			m.mu.Lock()
			m.subscriptionsByID[deviceID] = sub
			if oldDeviceID != "" {
				delete(m.subscriptionsByID, oldDeviceID)
			}
			m.mu.Unlock()
		}
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

	// 查询设备 location info (addressInfo) 用于后续 alarm/event 消息
	// 使用 GetDeviceLocationInfo 方法，通过 device_uid 查询
	locationInfo, err := m.deviceRepo.GetDeviceLocationInfo(ctx, deviceUID)
	if err != nil {
		m.logger.Warn("Failed to get device location info during periodic subscription",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		// 查询失败时返回错误，不创建订阅记录
		return fmt.Errorf("failed to get device location info: %w", err)
	}

	// 从 locationInfo 中提取字段
	tenantID := locationInfo.TenantID
	deviceCode := "" // GetDeviceLocationInfo 不返回 device_code，需要从 device_store 查询
	deviceType := ""
	if locationInfo.DeviceType.Valid {
		deviceType = locationInfo.DeviceType.String
	}
	var branchID, buildingID, unitID, roomID, bedID *string
	if locationInfo.BranchID.Valid {
		branchID = &locationInfo.BranchID.String
	}
	if locationInfo.BuildingID.Valid {
		buildingID = &locationInfo.BuildingID.String
	}
	if locationInfo.UnitID.Valid {
		unitID = &locationInfo.UnitID.String
	}
	if locationInfo.RoomID.Valid {
		roomID = &locationInfo.RoomID.String
	}
	if locationInfo.BedID.Valid {
		bedID = &locationInfo.BedID.String
	}

	// 查询 device_code（GetDeviceLocationInfo 不返回 device_code）
	deviceCodeQuery := `SELECT device_code FROM device_store WHERE device_uid = $1 AND allow_access = TRUE AND device_type = 'Radar' LIMIT 1`
	err = m.db.QueryRowContext(ctx, deviceCodeQuery, deviceUID).Scan(&deviceCode)
	if err != nil {
		m.logger.Warn("Failed to get device_code, using empty string",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		deviceCode = ""
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
		DeviceUID:       deviceUID,
		DeviceID:        deviceID,
		PropTopic:       propTopic,
		MonitorTopic:    monitorTopic,
		FuncTopic:       funcTopic,
		StatTopic:       statTopic,
		EventTopic:      eventTopic,
		AlarmTopic:      alarmTopic,
		MonitorSubTime:  now,
		LastSeen:        time.Time{}, // 初始化为零值，表示还没有收到设备消息
		LastMonitorTime: time.Time{}, // 初始化为零值
		LastStatTime:    time.Time{}, // 初始化为零值
		LastEventTime:   time.Time{}, // 初始化为零值
		LastAlarmTime:   time.Time{}, // 初始化为零值
		Status:          "online",    // 初始状态为online
		TenantID:        tenantID,
		DeviceCode:      deviceCode,
		DeviceType:      deviceType,
	}

	// 设置 location info（如果存在）
	if branchID != nil {
		sub.BranchID = branchID
	}
	if buildingID != nil {
		sub.BuildingID = buildingID
	}
	if unitID != nil {
		sub.UnitID = unitID
	}
	if roomID != nil {
		sub.RoomID = roomID
	}
	if bedID != nil {
		sub.BedID = bedID
	}
	m.subscriptionsByUID[deviceUID] = sub

	// auth 成功后，立即发布 online event 到 alarm stream 并加入 online 列表
	go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "online")

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
// 根据消息类型（monitor/stat/event/alarm）更新对应的时间戳
// 如果设备还没有发送monitor订阅命令，在收到第一条消息时自动发送
// 注意：此方法已废弃，应使用 UpdateLastSeenByType
func (m *DeviceSubscriptionManager) UpdateLastSeen(deviceUID string) {
	// 为了向后兼容，默认使用 "stat" 类型
	m.UpdateLastSeenByType(deviceUID, "stat")
}

// UpdateLastSeenByType 根据消息类型更新设备最后收到消息的时间
// topicType: "monitor", "stat", "event", "alarm"
func (m *DeviceSubscriptionManager) UpdateLastSeenByType(deviceUID, topicType string) {
	m.mu.Lock()
	sub, exists := m.subscriptionsByUID[deviceUID]
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
	// 根据消息类型更新对应的时间戳
	switch topicType {
	case "monitor":
		sub.LastMonitorTime = now
	case "stat":
		sub.LastStatTime = now
	case "event":
		sub.LastEventTime = now
	case "alarm":
		sub.LastAlarmTime = now
	}

	// 取4个时间戳中的最新值作为 LastSeen
	oldLastSeen := sub.LastSeen
	times := []time.Time{
		sub.LastMonitorTime,
		sub.LastStatTime,
		sub.LastEventTime,
		sub.LastAlarmTime,
	}
	newLastSeen := oldLastSeen
	for _, t := range times {
		if !t.IsZero() && t.After(newLastSeen) {
			newLastSeen = t
		}
	}
	sub.LastSeen = newLastSeen

	// 状态恢复逻辑
	switch oldStatus {
	case "offline":
		// offline状态收到任何MQTT消息，如果now - oldLastSeen不超过180秒，恢复为online
		timeSinceLastSeen := now.Sub(oldLastSeen)
		if timeSinceLastSeen <= 180*time.Second {
			sub.Status = "online"
			sub.mu.Unlock()

			log.Printf("✅ Device %s recovered from offline to online (received %s message)", deviceUID, topicType)
			m.logger.Info("Device recovered from offline to online",
				zap.String("device_uid", deviceUID),
				zap.String("message_type", topicType),
			)
			m.triggerStatusChange(deviceUID, "offline", "online")

			// 发送 online 事件到 event stream（按照 Reside_stream_stand.md 格式）
			sub.mu.RLock()
			deviceID := sub.DeviceID
			sub.mu.RUnlock()
			go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "online")
		} else {
			// 超过180秒，应该已经被取消订阅了，这里不应该发生
			sub.mu.Unlock()
		}

	case "unsubscribed":
		// 已取消订阅的设备收到任何MQTT消息，需要重新认证
		sub.mu.Unlock()
		log.Printf("⚠️ Unsubscribed device %s received MQTT message (type: %s), requires re-authentication", deviceUID, topicType)
		m.logger.Warn("Unsubscribed device received MQTT message, requires re-authentication",
			zap.String("device_uid", deviceUID),
			zap.String("message_type", topicType),
		)

	default:
		// online 状态：不再在这里触发周期性发布，由独立的定时器 goroutine 处理
		sub.mu.Unlock()
	}

	m.mu.Unlock()

	// 如果这是设备首次收到消息（LastSeen为零值），发送monitor订阅命令并发布 online event
	if oldLastSeen.IsZero() {
		log.Printf("Device %s first MQTT message received (type: %s), sending monitor subscription command and publishing online event", deviceUID, topicType)

		// 发布 online event 到 alarm stream
		m.mu.RLock()
		sub, exists := m.subscriptionsByUID[deviceUID]
		m.mu.RUnlock()
		if exists {
			sub.mu.RLock()
			deviceID := sub.DeviceID
			sub.mu.RUnlock()
			if deviceID != "" {
				go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "online")
			}
		}

		// 发送 monitor 订阅命令
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

	// 如果设备不在订阅列表中，创建订阅记录（包含 location info）
	m.mu.Lock()
	_, exists := m.subscriptionsByUID[deviceUID]
	if !exists {
		// 查询设备 location info (addressInfo) 用于后续 alarm/event 消息
		locationInfo, err := m.deviceRepo.GetDeviceLocationInfo(ctx, deviceUID)
		if err != nil {
			m.logger.Warn("Failed to get device location info on first message, creating subscription without location info",
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
			// 即使查询失败，也创建订阅记录（但没有 location info）
			now := time.Now()
			sub := &DeviceSubscription{
				DeviceUID:       deviceUID,
				DeviceID:        deviceID,
				PropTopic:       m.mqttClient.BuildTopic("prop", deviceUID),
				MonitorTopic:    m.mqttClient.BuildTopic("monitor", deviceUID),
				FuncTopic:       m.mqttClient.BuildTopic("func", deviceUID),
				StatTopic:       m.mqttClient.BuildTopic("stat", deviceUID),
				EventTopic:      m.mqttClient.BuildTopic("event", deviceUID),
				AlarmTopic:      m.mqttClient.BuildTopic("alarm", deviceUID),
				MonitorSubTime:  now,
				LastSeen:        time.Time{}, // 初始化为零值
				LastMonitorTime: time.Time{}, // 初始化为零值
				LastStatTime:    time.Time{}, // 初始化为零值
				LastEventTime:   time.Time{}, // 初始化为零值
				LastAlarmTime:   time.Time{}, // 初始化为零值
				Status:          "online",    // 初始状态为online
			}
			// 同时存储到两个 map
			m.subscriptionsByUID[deviceUID] = sub
			if deviceID != "" {
				m.subscriptionsByID[deviceID] = sub
			}
			m.mu.Unlock()
			return
		}

		// 从 locationInfo 中提取字段
		tenantID := locationInfo.TenantID
		deviceCode := "" // GetDeviceLocationInfo 不返回 device_code，需要从 device_store 查询
		deviceType := ""
		if locationInfo.DeviceType.Valid {
			deviceType = locationInfo.DeviceType.String
		}
		var branchID, buildingID, unitID, roomID, bedID *string
		if locationInfo.BranchID.Valid {
			branchID = &locationInfo.BranchID.String
		}
		if locationInfo.BuildingID.Valid {
			buildingID = &locationInfo.BuildingID.String
		}
		if locationInfo.UnitID.Valid {
			unitID = &locationInfo.UnitID.String
		}
		if locationInfo.RoomID.Valid {
			roomID = &locationInfo.RoomID.String
		}
		if locationInfo.BedID.Valid {
			bedID = &locationInfo.BedID.String
		}

		// 查询 device_code（GetDeviceLocationInfo 不返回 device_code）
		deviceCodeQuery := `SELECT device_code FROM device_store WHERE device_uid = $1 AND allow_access = TRUE AND device_type = 'Radar' LIMIT 1`
		err = m.db.QueryRowContext(ctx, deviceCodeQuery, deviceUID).Scan(&deviceCode)
		if err != nil {
			m.logger.Warn("Failed to get device_code on first message, using empty string",
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
			deviceCode = ""
		}

		// 创建订阅记录（包含 location info）
		now := time.Now()
		sub := &DeviceSubscription{
			DeviceUID:       deviceUID,
			DeviceID:        deviceID,
			PropTopic:       m.mqttClient.BuildTopic("prop", deviceUID),
			MonitorTopic:    m.mqttClient.BuildTopic("monitor", deviceUID),
			FuncTopic:       m.mqttClient.BuildTopic("func", deviceUID),
			StatTopic:       m.mqttClient.BuildTopic("stat", deviceUID),
			EventTopic:      m.mqttClient.BuildTopic("event", deviceUID),
			AlarmTopic:      m.mqttClient.BuildTopic("alarm", deviceUID),
			MonitorSubTime:  now,
			LastSeen:        time.Time{}, // 初始化为零值
			LastMonitorTime: time.Time{}, // 初始化为零值
			LastStatTime:    time.Time{}, // 初始化为零值
			LastEventTime:   time.Time{}, // 初始化为零值
			LastAlarmTime:   time.Time{}, // 初始化为零值
			Status:          "online",    // 初始状态为online
			// Location info (addressInfo)
			TenantID:   tenantID,
			DeviceCode: deviceCode,
			DeviceType: deviceType,
			BranchID:   branchID,
			BuildingID: buildingID,
			UnitID:     unitID,
			RoomID:     roomID,
			BedID:      bedID,
		}
		// 同时存储到两个 map
		m.subscriptionsByUID[deviceUID] = sub
		if deviceID != "" {
			m.subscriptionsByID[deviceID] = sub
		}
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
// 检查 monitor/stat/event/alarm 最新消息，取4者最新值更新 LastSeen
// 当 monitor/stat/event/alarm 最新消息均超过90秒，标记为 offline 并发送 alarm 事件
func (m *DeviceSubscriptionManager) checkDeviceHeartbeat(ctx context.Context) {
	m.mu.RLock()
	subs := make([]*DeviceSubscription, 0, len(m.subscriptionsByUID))
	for _, sub := range m.subscriptionsByUID {
		subs = append(subs, sub)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, sub := range subs {
		sub.mu.Lock()
		deviceUID := sub.DeviceUID
		deviceID := sub.DeviceID
		status := sub.Status

		// 获取4种消息类型的最新时间戳
		lastMonitorTime := sub.LastMonitorTime
		lastStatTime := sub.LastStatTime
		lastEventTime := sub.LastEventTime
		lastAlarmTime := sub.LastAlarmTime

		// 取4个时间戳中的最新值作为 LastSeen
		times := []time.Time{
			lastMonitorTime,
			lastStatTime,
			lastEventTime,
			lastAlarmTime,
		}
		newLastSeen := time.Time{}
		for _, t := range times {
			if !t.IsZero() && t.After(newLastSeen) {
				newLastSeen = t
			}
		}
		// 更新 LastSeen
		if !newLastSeen.IsZero() {
			sub.LastSeen = newLastSeen
		}
		lastSeen := sub.LastSeen
		sub.mu.Unlock()

		// 跳过还未收到过任何消息的设备
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
			delete(m.subscriptionsByUID, deviceUID)
			if deviceID != "" {
				delete(m.subscriptionsByID, deviceID)
			}
			m.mu.Unlock()
			continue
		}

		// 检查4种消息类型是否都超过90秒
		timeSinceMonitor := time.Duration(0)
		timeSinceStat := time.Duration(0)
		timeSinceEvent := time.Duration(0)
		timeSinceAlarm := time.Duration(0)

		if !lastMonitorTime.IsZero() {
			timeSinceMonitor = now.Sub(lastMonitorTime)
		} else {
			timeSinceMonitor = 999 * time.Hour // 未收到过消息，视为超时
		}
		if !lastStatTime.IsZero() {
			timeSinceStat = now.Sub(lastStatTime)
		} else {
			timeSinceStat = 999 * time.Hour
		}
		if !lastEventTime.IsZero() {
			timeSinceEvent = now.Sub(lastEventTime)
		} else {
			timeSinceEvent = 999 * time.Hour
		}
		if !lastAlarmTime.IsZero() {
			timeSinceAlarm = now.Sub(lastAlarmTime)
		} else {
			timeSinceAlarm = 999 * time.Hour
		}

		// 检查是否所有消息类型都超过90秒
		allTimeout := timeSinceMonitor > 90*time.Second &&
			timeSinceStat > 90*time.Second &&
			timeSinceEvent > 90*time.Second &&
			timeSinceAlarm > 90*time.Second

		timeSinceLastSeen := now.Sub(lastSeen)

		switch status {
		case "online":
			// 阶段1：所有消息类型（monitor/stat/event/alarm）都超过90秒 → 标记offline
			if allTimeout {
				m.markDeviceOffline(deviceUID, now)
			} else {
				// online 状态：检查是否需要周期性发布 online（每2分钟一次）
				sub.mu.Lock()
				lastPub := sub.LastOnlineStatusPublishTime
				sub.mu.Unlock()
				if now.Sub(lastPub) >= 2*time.Minute {
					sub.mu.Lock()
					sub.LastOnlineStatusPublishTime = now
					sub.mu.Unlock()
					m.publishDeviceOnlineStatusToConfigStream(ctx, deviceUID, "online")
				}
			}

		case "offline":
			// 阶段2：180秒无任何消息 → 取消订阅
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
// 当 monitor/stat/event/alarm 最新消息均超过90秒时调用
func (m *DeviceSubscriptionManager) markDeviceOffline(deviceUID string, _ time.Time) {
	m.mu.Lock()
	sub, exists := m.subscriptionsByUID[deviceUID]
	m.mu.Unlock()

	if !exists {
		return
	}

	sub.mu.Lock()
	if sub.Status == "online" {
		sub.Status = "offline"
		deviceID := sub.DeviceID
		sub.mu.Unlock()

		log.Printf("⚠️ Device %s marked as offline (all message types timeout > 90s)", deviceUID)
		m.logger.Info("Device marked as offline (all message types timeout > 90s)",
			zap.String("device_uid", deviceUID),
		)

		// 触发offline事件，发布到 config stream
		m.triggerStatusChange(deviceUID, "online", "offline")

		// 发送 offline 事件到 alarm stream（offline 是告警）
		go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "offline")
	} else {
		sub.mu.Unlock()
	}
}

// PublishOnlineForConnectedDevices 对 deviceUIDs 中已在列表且状态为 online 的设备发布上线通知（由 mqtt_consumer 启动 1 分钟后检查队列后调用）
func (m *DeviceSubscriptionManager) PublishOnlineForConnectedDevices(ctx context.Context, deviceUIDs []string) {
	published := 0
	for _, deviceUID := range deviceUIDs {
		m.mu.RLock()
		sub, ok := m.subscriptionsByUID[deviceUID]
		m.mu.RUnlock()
		if !ok {
			continue
		}
		sub.mu.RLock()
		status, deviceID := sub.Status, sub.DeviceID
		sub.mu.RUnlock()
		if status != "online" {
			continue
		}
		go m.publishDeviceOnlineStatusEvent(context.Background(), deviceUID, deviceID, "online")
		published++
	}
	m.logger.Info("Published online notification for connected devices after startup",
		zap.Int("published", published),
		zap.Int("requested", len(deviceUIDs)),
	)
}

// publishDeviceOnlineStatusEvent 发送设备在线状态事件
// 无论 online/offline，统一发送到 alarm stream
// 格式：topic_type="alarm", category="isOnline" 或 "OfflineAlarm", data_value=[{"category":"isOnline|OfflineAlarm","device_status":"online|offline","device_uid":"xxx"}]
// deviceStatus: "online" 或 "offline"（字符串，按照 Reside_stream_stand.md 规范）
// 注意：只使用内存中的 addressInfo，如果没有则不查询数据库，直接返回
func (m *DeviceSubscriptionManager) publishDeviceOnlineStatusEvent(ctx context.Context, deviceUID, deviceID, deviceStatus string) {
	if m.streamPublisher == nil {
		m.logger.Warn("StreamPublisher not set, skipping device online status event publish",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
		)
		return
	}

	// 从内存中的 DeviceSubscription 获取 location info (addressInfo)
	m.mu.RLock()
	sub, exists := m.subscriptionsByUID[deviceUID]
	m.mu.RUnlock()

	if !exists {
		m.logger.Warn("Device subscription not found in memory, skipping device online status event",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
		)
		return
	}

	// 检查是否有 addressInfo（至少要有 TenantID）
	sub.mu.RLock()
	tenantID := sub.TenantID
	deviceType := sub.DeviceType
	branchID := sub.BranchID
	buildingID := sub.BuildingID
	unitID := sub.UnitID
	roomID := sub.RoomID
	bedID := sub.BedID
	sub.mu.RUnlock()

	// 如果没有 addressInfo（TenantID 为空），不查询数据库，直接返回
	if tenantID == "" {
		m.logger.Warn("Device addressInfo not found in memory, skipping device online status event",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
		)
		return
	}

	// 如果 tenant_id 为空字符串，使用默认值
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000000"
	}

	// 构建 data_value（按照 Reside_stream_stand.md 格式）
	// device_status 使用字符串 "online" 或 "offline"（对应原始值 0=online, 1=offline）
	// offline 时直接使用 "OfflineAlarm" 作为 category，online 时使用 "isOnline"
	var category string
	if deviceStatus == "offline" {
		category = "OfflineAlarm"
	} else {
		category = "isOnline"
	}
	dataValueArray := []map[string]interface{}{
		{
			"category":      category,
			"device_status": deviceStatus, // "online" 或 "offline"（字符串）
			"device_uid":    deviceUID,
		},
	}

	// 序列化 data_value 为 JSON 字符串
	dataValueJSON, err := json.Marshal(dataValueArray)
	if err != nil {
		m.logger.Warn("Failed to marshal data_value for device online status event",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
			zap.Error(err),
		)
		return
	}

	// 构建位置信息（从内存中获取）
	var locInfo *rediscommon.LocationInfo
	if branchID != nil || buildingID != nil || unitID != nil || roomID != nil || bedID != nil {
		locInfo = &rediscommon.LocationInfo{}
		if branchID != nil && *branchID != "" {
			locInfo.BranchID = branchID
		}
		if buildingID != nil && *buildingID != "" {
			locInfo.BuildingID = buildingID
		}
		if unitID != nil && *unitID != "" {
			locInfo.UnitID = unitID
		}
		if roomID != nil && *roomID != "" {
			locInfo.RoomID = roomID
		}
		if bedID != nil && *bedID != "" {
			locInfo.BedID = bedID
		}
	}

	// 统一发送到 alarm stream（无论 online/offline）
	// 原因：同一件事（设备状态变化），且 Alarm 事件较少，易于后端过滤
	// 方便订阅方处理，只需订阅一个 stream
	var topicType string
	var streamName string
	topicType = "alarm"
	streamName = rediscommon.StreamAlarm.Name

	// 使用 BuildIoTStreamMessage 构建事件消息
	msg := rediscommon.BuildIoTStreamMessage(
		deviceID,
		deviceUID,
		deviceType,
		tenantID,
		time.Now().Unix(),
		topicType, // "event" 或 "alarm"
		category,  // "OfflineAlarm" 或 "isOnline"
		map[string]interface{}{"data_value": dataValueArray}, // dataValue (作为 map，包含 data_value 数组)
		locInfo,
	)

	// 转换为 map（按照 iotStreamMessageToMap 的逻辑）
	eventData := make(map[string]interface{})
	eventData["device_id"] = msg.DeviceID
	eventData["device_uid"] = msg.DeviceUID
	eventData["device_type"] = msg.DeviceType
	eventData["tenant_id"] = msg.TenantID
	eventData["timestamp"] = fmt.Sprintf("%d", msg.Timestamp)
	eventData["topic_type"] = msg.TopicType
	eventData["category"] = msg.Category
	eventData["data_value"] = string(dataValueJSON) // 序列化为 JSON 字符串

	// 位置信息字段（可选）
	if locInfo != nil {
		if locInfo.BranchID != nil && *locInfo.BranchID != "" {
			eventData["branch_id"] = *locInfo.BranchID
		}
		if locInfo.BuildingID != nil && *locInfo.BuildingID != "" {
			eventData["building_id"] = *locInfo.BuildingID
		}
		if locInfo.UnitID != nil && *locInfo.UnitID != "" {
			eventData["unit_id"] = *locInfo.UnitID
		}
		if locInfo.RoomID != nil && *locInfo.RoomID != "" {
			eventData["room_id"] = *locInfo.RoomID
		}
		if locInfo.BedID != nil && *locInfo.BedID != "" {
			eventData["bed_id"] = *locInfo.BedID
		}
	}

	// 发布到对应的 stream
	_, err = m.streamPublisher.PublishToStream(ctx, streamName, eventData)
	if err != nil {
		m.logger.Warn("Failed to publish device online status event",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
			zap.String("stream", streamName),
			zap.Error(err),
		)
	} else {
		m.logger.Info("Published device online status event",
			zap.String("device_uid", deviceUID),
			zap.String("device_status", deviceStatus),
			zap.String("stream", streamName),
			zap.String("topic_type", topicType),
		)
		log.Printf("✅ Published device %s status event (%s) to %s stream", deviceUID, deviceStatus, streamName)
	}
}

// unsubscribeDevice 第二阶段：取消订阅设备
func (m *DeviceSubscriptionManager) unsubscribeDevice(deviceUID string) {
	m.mu.Lock()
	sub, exists := m.subscriptionsByUID[deviceUID]
	m.mu.Unlock()

	if !exists {
		return
	}

	sub.mu.Lock()
	deviceID := sub.DeviceID // 先获取 deviceID，用于删除索引
	if sub.Status == "offline" {
		sub.Status = "unsubscribed"
		sub.mu.Unlock()

		log.Printf("🚫 Device %s unsubscribed (180s no messages)", deviceUID)
		m.logger.Info("Device unsubscribed (180s no messages)",
			zap.String("device_uid", deviceUID),
		)

		// 1. 发送 unsubscribed 事件到 config stream
		m.triggerStatusChange(deviceUID, "offline", "unsubscribed")

		// 2. 取消MQTT订阅
		if m.mqttConsumer != nil {
			go func() {
				m.mqttConsumer.UnsubscribeDeviceTopics(deviceUID)
				log.Printf("✅ Unsubscribed MQTT topics for device %s", deviceUID)
				m.logger.Info("Unsubscribed device MQTT topics",
					zap.String("device_uid", deviceUID),
				)
			}()
		}

		// 3. 从订阅列表中移除并标记需要重认证（注销内存中的Device）
		m.mu.Lock()
		delete(m.subscriptionsByUID, deviceUID)
		if deviceID != "" {
			delete(m.subscriptionsByID, deviceID)
		}
		m.unsubscribedDueToTimeout[deviceUID] = struct{}{}
		m.mu.Unlock()

		// 4. 触发取消订阅事件（用于其他模块通知）
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
	subs := make([]*DeviceSubscription, 0, len(m.subscriptionsByUID))
	for _, sub := range m.subscriptionsByUID {
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
// 注意：状态由 checkDeviceHeartbeat 定期检查并更新（每90秒），这里直接返回内存中的状态
// 自动判断传入的是 device_uid 还是 device_id：
//   - 如果是 UUID 格式（36个字符，4个连字符），认为是 device_id，从 subscriptionsByID 查找
//   - 否则认为是 device_uid，从 subscriptionsByUID 查找
//
// 不查数据库，完全基于内存
func (m *DeviceSubscriptionManager) GetDeviceOnlineStatus(identifier string) string {
	var sub *DeviceSubscription
	var exists bool
	var deviceUID string

	m.mu.RLock()
	// 自动判断是 device_id (UUID格式) 还是 device_uid (短字符串格式)
	// UUID格式: 8-4-4-4-12 (例如: 791fc634-69de-4987-b7eb-803c17e545a5)
	// device_uid格式: 通常12个字符，如 E598A2ACD523
	if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
		// 是 device_id (UUID格式)，从 subscriptionsByID 查找
		sub, exists = m.subscriptionsByID[identifier]
		if exists {
			deviceUID = sub.DeviceUID
		}
	} else {
		// 是 device_uid (短字符串格式)，从 subscriptionsByUID 查找
		sub, exists = m.subscriptionsByUID[identifier]
		if exists {
			deviceUID = identifier
		}
	}
	m.mu.RUnlock()

	if !exists {
		// 设备不在订阅列表中，可能是未认证或已取消订阅
		// 检查是否因超时被取消订阅
		if deviceUID != "" {
			m.mu.RLock()
			_, unsubscribed := m.unsubscribedDueToTimeout[deviceUID]
			m.mu.RUnlock()
			if unsubscribed {
				return "unsubscribed"
			}
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
