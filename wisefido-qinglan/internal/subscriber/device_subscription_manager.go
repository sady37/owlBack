package subscriber

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/publisher"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DeviceSubscription 设备订阅信息
type DeviceSubscription struct {
	DeviceUID                   string
	DeviceAddr                    string
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
	TenantID                    string
	DeviceCode                  string
	DeviceType                  string
	onlineStatusStopChan        chan struct{} // 用于停止周期性 online 状态发布的定时器
	mu                          sync.RWMutex
}

// MessageHandler MQTT消息处理器接口
type MessageHandler func(topic string, payload []byte) error

// RadarService 雷达服务接口，用于读取设备属性
type RadarService interface {
	GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error)
}

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
	subscriptionsByID        map[string]*DeviceSubscription // device_addr -> subscription，与 subscriptionsByUID 指向同一个对象
	unsubscribedDueToTimeout map[string]struct{}            // 90s 超时后强制取消订阅，需重认证
	mu                       sync.RWMutex
	checkInterval            time.Duration // 检查间隔（30秒）
	// offlineTimeout removed — 离线检测交由 cardagg
	monitorMaxAge            time.Duration // monitor订阅最大时长（1小时）
	defaultContent           int           // 默认订阅内容：0-同时订阅，1-轨迹，2-呼吸心率
	defaultDuration          int           // 默认订阅时长（秒），默认 3600
	prevHealth               map[string]DeviceHealthStatus // health_check 前一次状态，用于检测 1→0 恢复
	stopChan                 chan struct{}
	wg                       sync.WaitGroup
	radarService             RadarService // 雷达服务，用于健康检查时读取设备属性
}

// NewDeviceSubscriptionManager 创建设备订阅管理器
func NewDeviceSubscriptionManager(
	cfg *config.Config,
	mqttClient *mqtt.Client,
	db *sql.DB,
	logger *zap.Logger,
	messageHandler MessageHandler, // MQTT消息处理器
) *DeviceSubscriptionManager {
	mqttPublisher := publisher.NewMQTTPublisher(cfg, mqttClient, logger)
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
		// offlineTimeout removed
		monitorMaxAge:            1 * time.Hour,
		defaultContent:           0,    // 0-同时订阅（轨迹和呼吸心率）
		defaultDuration:          3600, // 默认1小时
		prevHealth:               make(map[string]DeviceHealthStatus),
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

// SetRadarService 设置雷达服务（用于健康检查时读取设备属性）
func (m *DeviceSubscriptionManager) SetRadarService(radarService RadarService) {
	m.radarService = radarService
}

// Start 启动订阅管理器
func (m *DeviceSubscriptionManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go m.heartbeatMonitor(ctx)
	m.wg.Add(1)
	go m.subscriptionRenewal(ctx)
	m.wg.Add(1)
	go m.healthCheckMonitor(ctx)
	// 重启后从 DB 重建所有已认证 radar 的订阅记录；monitor 命令在设备第一条 MQTT 消息到达时下发。
	// 缺这步则旧 1h 实时订阅到期后无人续，radar.track 全静默（event 流不受影响，难察觉）。
	go m.restoreAuthenticatedDeviceSubscriptions(ctx)
	m.logger.Info("device subscription manager started")
	return nil
}

// restoreAuthenticatedDeviceSubscriptions 恢复已认证设备的MQTT订阅
// 服务重启后，从DB查询所有已认证的 Radar 设备（access = TRUE AND device_type = 'Radar'），重新订阅并发送monitor命令
// 注意：wisefido-qinglan 只管理 Radar 设备，不管理 Sleepace 设备
// 简化逻辑：不检查内存中的状态记录，直接从DB重新订阅所有已认证的 Radar 设备
func (m *DeviceSubscriptionManager) restoreAuthenticatedDeviceSubscriptions(ctx context.Context) {
	m.logger.Info("restoring authenticated Radar subscriptions from db")

	// 已认证 Radar = dfm.device_type='Radar' AND devices.access=TRUE
	// identity = device_uid (logMAC)；订阅内部用 device_addr 作为 deviceAddr 占位（保留字段名兼容）。
	query := `
		SELECT dfm.device_uid, COALESCE(host(d.device_addr), '')
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_uid = dfm.device_uid
		WHERE d.access = TRUE
		  AND dfm.device_type = 'Radar'
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		m.logger.Error("query authenticated devices", zap.Error(err))
		return
	}
	defer rows.Close()

	restoredCount := 0
	for rows.Next() {
		var deviceUID, deviceAddr string
		if err := rows.Scan(&deviceUID, &deviceAddr); err != nil {
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
				DeviceAddr:        id,
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
					m.logger.Warn("restore device MQTT topics subscription",
						zap.String("device_uid", uid),
						zap.Error(err),
					)
				} else {
					m.logger.Info("restored device MQTT topics subscription",
						zap.String("device_uid", uid),
						zap.String("device_addr", id),
					)
				}
			}

			// 记录设备恢复状态
			m.logger.Info("Device status changed",
				zap.String("device_uid", uid),
				zap.String("old_status", ""),
				zap.String("new_status", "online"),
			)

			m.logger.Info("Restored subscription record for device (monitor on first message)",
				zap.String("device_uid", uid),
				zap.String("device_addr", id),
			)
		}(deviceUID, deviceAddr)

		restoredCount++
	}

	if err := rows.Err(); err != nil {
		m.logger.Error("iterate authenticated devices", zap.Error(err))
		return
	}
	m.logger.Info("restored subscriptions",
		zap.Int("count", restoredCount),
		zap.String("note", "monitor cmd sent on first mqtt message"),
	)
}

// Stop 停止订阅管理器
func (m *DeviceSubscriptionManager) Stop(ctx context.Context) error {
	m.logger.Info("device subscription manager stopping")
	close(m.stopChan)
	m.wg.Wait()
	m.logger.Info("device subscription manager stopped")
	return nil
}

// SubscribeDevice 订阅设备的实时数据（认证成功后调用）
// 通过向设备发送MQTT命令来订阅，而不是订阅MQTT主题
func (m *DeviceSubscriptionManager) SubscribeDevice(ctx context.Context, deviceUID, deviceAddr string) error {
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
			oldDeviceAddr := sub.DeviceAddr
			sub.DeviceAddr = deviceAddr
			sub.LastSeen = time.Now()
			sub.mu.Unlock()
			// 更新 subscriptionsByID（如果 deviceAddr 变化）。已在外层 m.mu.Lock 内，不可再 Lock（RWMutex 不可重入＝自死锁）。
			if deviceAddr != "" && deviceAddr != oldDeviceAddr {
				m.subscriptionsByID[deviceAddr] = sub
				if oldDeviceAddr != "" {
					delete(m.subscriptionsByID, oldDeviceAddr)
				}
			}
			m.logger.Debug("device already subscribed",
				zap.String("device_uid", deviceUID),
				zap.Time("valid_until", monitorSubTime.Add(m.monitorMaxAge)),
			)
			return nil
		}
		m.logger.Info("device subscription expiring, renewing", zap.String("device_uid", deviceUID))
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
					m.logger.Warn("MQTT subscription failed, retrying",
						zap.String("device_uid", deviceUID),
						zap.Int("retry", i+1),
						zap.Int("max_retries", maxRetries),
						zap.Error(err),
					)
					time.Sleep(retryDelay)
					continue
				}
			} else {
				subscribeErr = nil
				break
			}
		}

		if subscribeErr != nil {
			m.logger.Warn("subscribe device MQTT topics after retries failed",
				zap.String("device_uid", deviceUID),
				zap.Int("retries", maxRetries),
				zap.Error(subscribeErr),
			)
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

	// 查询 device_type 和 tenant_id
	deviceType := ""
	tenantID := ""
	deviceStoreInfo, err := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
	if err != nil {
		m.logger.Warn("Failed to get device store info during subscription",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get device store info: %w", err)
	}
	if deviceStoreInfo != nil {
		deviceType = deviceStoreInfo.DeviceType
		tenantID = deviceStoreInfo.TenantID
	}

	// 创建或更新订阅记录
	// 注意：LastSeen初始化为零值，表示还没有收到设备消息
	now := time.Now()
	sub := &DeviceSubscription{
		DeviceUID:       deviceUID,
		DeviceAddr:        deviceAddr,
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

		DeviceType: deviceType,
	}
	// 同时存储到两个 map
	m.subscriptionsByUID[deviceUID] = sub
	if deviceAddr != "" {
		m.subscriptionsByID[deviceAddr] = sub
	}

	// auth 成功后，立即发布 online event 到 deviceStatus stream
	go m.streamPublisher.PublishDeviceStatus(context.Background(), deviceStoreInfo.DeviceAddr, deviceUID, sub.DeviceType, map[string]int{
		observation.FieldOffline: 0,
	})
	// 上线 → 根据 offline=0 推导 category=DeviceRecover
	go m.publishDeviceAlarm(context.Background(), tenantID, deviceAddr, deviceUID, observation.FieldOffline, 0)
	// 记录订阅状态
	m.logger.Info("Device status changed",
		zap.String("device_uid", deviceUID),
		zap.String("old_status", ""),
		zap.String("new_status", "online"),
	)

	// 不在此处更新 DB 为 online；收到第一笔数据时在 UpdateLastSeen 中更新

	m.logger.Info("Device subscription recorded, sending monitor command",
		zap.String("device_uid", deviceUID),
		zap.String("device_addr", deviceAddr),
		zap.Time("monitor_sub_time", now),
		zap.Int("duration", m.defaultDuration),
		zap.Int("content", m.defaultContent),
	)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
			m.logger.Warn("send monitor subscription command failed",
				zap.String("device_uid", deviceUID),
				zap.String("device_addr", deviceAddr),
				zap.Error(err),
			)
		} else {
			m.logger.Info("sent monitor subscription command",
				zap.String("device_uid", deviceUID),
				zap.String("device_addr", deviceAddr),
			)
		}
	}()

	return nil
}

// EnablePeriodicSubscription 开启周期性订阅（认证成功后调用）
// 检查是否已订阅MQTT主题：已订阅则仅记录，否则开启订阅
func (m *DeviceSubscriptionManager) EnablePeriodicSubscription(ctx context.Context, deviceUID, deviceAddr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在订阅管理器中存在记录
	sub, exists := m.subscriptionsByUID[deviceUID]
	if exists {
		// 已存在订阅记录：更新信息，重置 LastSeen（设备连上 MQTT 发第一条消息时自动触发 monitor 订阅）
		sub.mu.Lock()
		oldDeviceAddr := sub.DeviceAddr
		sub.DeviceAddr = deviceAddr
		sub.Status = "online"
		sub.MonitorSubTime = time.Now()
		sub.LastSeen = time.Time{}
		sub.LastMonitorTime = time.Time{}
		sub.LastStatTime = time.Time{}
		sub.LastEventTime = time.Time{}
		sub.LastAlarmTime = time.Time{}
		sub.mu.Unlock()
		// 更新 subscriptionsByID（如果 deviceAddr 变化）。已在外层 m.mu.Lock 内，不可再 Lock（RWMutex 不可重入＝自死锁）。
		if deviceAddr != "" && deviceAddr != oldDeviceAddr {
			m.subscriptionsByID[deviceAddr] = sub
			if oldDeviceAddr != "" {
				delete(m.subscriptionsByID, oldDeviceAddr)
			}
		}
		m.logger.Info("device re-authenticated, lastseen reset",
			zap.String("device_uid", deviceUID),
			zap.String("device_addr", deviceAddr),
		)

		// 重新发布 online 状态：先查 device_addr
		var addrForReauth netip.Addr
		if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID); dsErr == nil && dsi != nil {
			addrForReauth = dsi.DeviceAddr
		}
		go m.streamPublisher.PublishDeviceStatus(context.Background(), addrForReauth, deviceUID, sub.DeviceType, map[string]int{
			observation.FieldOffline: 0,
		})
		go m.publishDeviceAlarm(context.Background(), sub.TenantID, deviceAddr, deviceUID, observation.FieldOffline, 0)

		// 延迟重试发送 monitor 订阅命令
		go m.sendMonitorWithRetry(deviceUID, deviceAddr)

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
			if err := m.mqttConsumer.SubscribeDeviceTopics(deviceUID); err != nil {
				m.logger.Warn("subscribe device mqtt topics",
					zap.String("device_uid", deviceUID),
					zap.Error(err),
				)
			} else {
				m.logger.Info("subscribed device mqtt topics", zap.String("device_uid", deviceUID))
			}
		}
	} else {
		m.logger.Debug("device mqtt topics already subscribed", zap.String("device_uid", deviceUID))
	}
	m.logger.Info("auth_flow mqtt_topics_ready",
		zap.String("uid", deviceUID),
		zap.Bool("already_subscribed", isMQTTSubscribed),
	)

	// 查询 device_type 和 tenant_id
	deviceType := ""
	tenantID := ""
	deviceStoreInfo, err := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
	if err != nil {
		m.logger.Warn("Failed to get device store info during periodic subscription",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get device store info: %w", err)
	}
	if deviceStoreInfo != nil {
		deviceType = deviceStoreInfo.DeviceType
		tenantID = deviceStoreInfo.TenantID
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
		DeviceAddr:        deviceAddr,
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

		DeviceType: deviceType,
	}
	m.subscriptionsByUID[deviceUID] = sub

	// auth 成功后，立即发布 online event 到 deviceStatus stream
	go m.streamPublisher.PublishDeviceStatus(context.Background(), deviceStoreInfo.DeviceAddr, deviceUID, sub.DeviceType, map[string]int{
		observation.FieldOffline: 0,
	})
	// 上线 → 根据 offline=0 推导 category=DeviceRecover
	go m.publishDeviceAlarm(context.Background(), sub.TenantID, deviceAddr, deviceUID, observation.FieldOffline, 0)
	// 记录周期订阅状态
	m.logger.Info("Device subscription created (periodic)",
		zap.String("device_uid", deviceUID),
		zap.String("status", "online"),
	)

	m.logger.Info("periodic subscription enabled",
		zap.String("device_uid", deviceUID),
		zap.String("device_addr", deviceAddr),
		zap.String("monitor_topic", monitorTopic),
		zap.String("tenant_id", tenantID),
		zap.Bool("mqtt_already_subscribed", isMQTTSubscribed),
		zap.Time("monitor_sub_time", now),
		zap.Int("duration", m.defaultDuration),
		zap.Int("content", m.defaultContent),
	)

	// 延迟重试发送 monitor 订阅命令：设备 auth 后需要时间连接 MQTT，立即发会丢失
	go m.sendMonitorWithRetry(deviceUID, deviceAddr)

	return nil
}

// sendMonitorWithRetry auth 后延迟重试发送 monitor 订阅命令
// 设备需要时间连接 MQTT，立即发 QoS0 会丢失；在 3s、6s、12s 后重试，收到 monitor 数据则停止
func (m *DeviceSubscriptionManager) sendMonitorWithRetry(deviceUID, deviceAddr string) {
	delays := []time.Duration{3 * time.Second, 6 * time.Second, 12 * time.Second}
	for i, delay := range delays {
		time.Sleep(delay)
		m.logger.Debug("monitor retry attempt",
			zap.String("uid", deviceUID),
			zap.Int("attempt", i),
			zap.Duration("delay", delay),
		)

		m.mu.RLock()
		sub, exists := m.subscriptionsByUID[deviceUID]
		m.mu.RUnlock()
		if !exists {
			m.logger.Warn("monitor retry abort: device no longer in subscription map", zap.String("uid", deviceUID))
			return
		}
		sub.mu.RLock()
		hasMonitorData := !sub.LastMonitorTime.IsZero()
		sub.mu.RUnlock()
		if hasMonitorData {
			m.logger.Info("monitor retry stopped: already receiving monitor data", zap.String("uid", deviceUID))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration)
		cancel()
		if err != nil {
			m.logger.Warn("monitor retry failed",
				zap.String("uid", deviceUID),
				zap.Int("attempt", i),
				zap.Error(err),
			)
		} else {
			m.logger.Info("monitor command sent",
				zap.String("uid", deviceUID),
				zap.Int("attempt", i),
				zap.Duration("after", delay),
			)
		}
	}
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
		// 需要查询device_addr并发送monitor订阅命令
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

	switch oldStatus {
	case "unsubscribed":
		sub.mu.Unlock()
		m.logger.Warn("Unsubscribed device received MQTT message, requires re-authentication",
			zap.String("device_uid", deviceUID),
			zap.String("message_type", topicType),
		)
	default:
		sub.mu.Unlock()
	}

	m.mu.Unlock()

	// 如果这是设备首次收到消息（LastSeen为零值），发送monitor订阅命令并发布 online event
	if oldLastSeen.IsZero() {
		// 发布 online event 到 alarm stream 和 config stream
		m.mu.RLock()
		sub, exists := m.subscriptionsByUID[deviceUID]
		m.mu.RUnlock()
		if exists {
			sub.mu.RLock()
			deviceAddr := sub.DeviceAddr
			deviceType := sub.DeviceType
			tenantID := sub.TenantID
			sub.mu.RUnlock()
			if deviceAddr != "" {
				var addrLocal netip.Addr
				if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(context.Background(), deviceUID); dsErr == nil && dsi != nil {
					addrLocal = dsi.DeviceAddr
				}
				_ = tenantID
				go m.streamPublisher.PublishDeviceStatus(context.Background(), addrLocal, deviceUID, deviceType, map[string]int{
					observation.FieldOffline: 0,
				})
				// 与 Sleepace connectionStatus 一致：首条 MQTT 后发 iot:alarm:stream OfflineRecover，cardagg 据此更新 device_status
				go func() {
					ctx := context.Background()
					item := observation.NewEventItem(time.Now().UnixMilli(), "end")
					item.TrackID = observation.TrackDevice
					data, _ := observation.EventItemToDataMap(&item)
					if data == nil {
						data = make(map[string]interface{})
					}
					data[observation.FieldOffline] = 0
					msg := rediscommon.NewSingleItemMessage(addrLocal, "", deviceType, time.Now().UnixMilli(), "alarm", alarm.AlarmTypeOfflineRecover, data)
					_ = m.streamPublisher.PublishAlarm(ctx, msg)
				}()
			}
		}

		// 发送 monitor 订阅命令
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
				m.logger.Warn("Failed to send monitor subscription on first message",
					zap.String("device_uid", deviceUID),
					zap.Error(err),
				)
			}
		}()

		m.logger.Info("Device online, monitor subscribed",
			zap.String("device_uid", deviceUID),
			zap.String("first_msg_type", topicType),
		)
	}

	// DB status 不存 on/offline；on/offline 仅存内存（LastSeen + 90s 超时）
}

// autoSubscribeOnFirstMessage 在收到设备第一条消息时自动发送monitor订阅命令
func (m *DeviceSubscriptionManager) autoSubscribeOnFirstMessage(ctx context.Context, deviceUID string) {
	// Phase 2: device identity = device_uid；access=TRUE 来自 devices；device_type 来自 dfm
	// 仅校验该 device_uid 在 dfm + devices 双表存在且 access=TRUE 且 device_type=Radar
	query := `
		SELECT dfm.device_uid
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_uid = dfm.device_uid
		WHERE dfm.device_uid = $1 AND d.access = TRUE AND dfm.device_type = 'Radar'`
	var deviceAddr string
	err := m.db.QueryRowContext(ctx, query, deviceUID).Scan(&deviceAddr)
	if err != nil {
		m.logger.Warn("Failed to verify device for auto-subscribe",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return
	}

	// 检查 devices.status，disabled 的设备不创建订阅（避免与 checkDeviceHeartbeat 形成创建→删除死循环）
	dbStatus, err := m.getDeviceStatus(ctx, deviceAddr)
	if err == nil && dbStatus == "disabled" {
		m.logger.Debug("Skipping auto-subscribe for disabled device",
			zap.String("device_uid", deviceUID),
			zap.String("device_addr", deviceAddr),
		)
		return
	}

	// 订阅设备的6个MQTT主题（不使用通配符，仅订阅该设备的主题）
	if m.mqttConsumer != nil {
		if err := m.mqttConsumer.SubscribeDeviceTopics(deviceUID); err != nil {
			m.logger.Warn("Failed to subscribe device MQTT topics",
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
			return
		}
	}

	// 如果设备不在订阅列表中，创建订阅记录
	m.mu.Lock()
	_, exists := m.subscriptionsByUID[deviceUID]
	justCreated := false
	var subDeviceType, subTenantID string
	if !exists {
		// 查询 device_type 和 tenant_id
		deviceStoreInfo, err := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
		if err != nil {
			m.logger.Warn("Failed to get device store info on first message, creating subscription without device info",
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
			// 即使查询失败，也创建订阅记录（但没有 device info）
		} else if deviceStoreInfo != nil {
			subDeviceType = deviceStoreInfo.DeviceType
			subTenantID = deviceStoreInfo.TenantID
		}

		now := time.Now()
		sub := &DeviceSubscription{
			DeviceUID:    deviceUID,
			DeviceAddr:     deviceAddr,
			PropTopic:    m.mqttClient.BuildTopic("prop", deviceUID),
			MonitorTopic: m.mqttClient.BuildTopic("monitor", deviceUID),
			FuncTopic:    m.mqttClient.BuildTopic("func", deviceUID),
			StatTopic:    m.mqttClient.BuildTopic("stat", deviceUID),
			EventTopic:   m.mqttClient.BuildTopic("event", deviceUID),
			AlarmTopic:   m.mqttClient.BuildTopic("alarm", deviceUID),
			MonitorSubTime: now,
			// LastSeen 立刻设为 now：本函数被调到=设备已发首条 MQTT，是"已上线"的实证。
			// 旧逻辑 LastSeen=zero 会让 UpdateLastSeenByType 首次匹配不到 sub 早 return，
			// path A 的 OfflineRecover 直到第二条 MQTT 才触发——心跳间隔 5min 的设备就要等 5min 才上线。
			LastSeen:        now,
			LastMonitorTime: time.Time{},
			LastStatTime:    time.Time{},
			LastEventTime:   time.Time{},
			LastAlarmTime:   time.Time{},
			Status:          "online",
			TenantID:        subTenantID,
			DeviceType:      subDeviceType,
		}
		// 同时存储到两个 map
		m.subscriptionsByUID[deviceUID] = sub
		if deviceAddr != "" {
			m.subscriptionsByID[deviceAddr] = sub
		}
		justCreated = true
	}
	m.mu.Unlock()

	// 首次创建 → 立即发 OfflineRecover（与 UpdateLastSeenByType 的 path A 等价，但提前一拍）。
	// 这里的实证：本函数被调用本身证明 mqtt_consumer 收到了一条该设备的 MQTT 包；
	// 不再依赖第二条 MQTT，离线/在线判定立刻反映真实状态。
	if justCreated && deviceAddr != "" && m.streamPublisher != nil {
		var addrJC netip.Addr
		if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID); dsErr == nil && dsi != nil {
			addrJC = dsi.DeviceAddr
		}
		_ = subTenantID
		go m.streamPublisher.PublishDeviceStatus(context.Background(), addrJC, deviceUID, subDeviceType, map[string]int{
			observation.FieldOffline: 0,
		})
		go func() {
			pubCtx := context.Background()
			item := observation.NewEventItem(time.Now().UnixMilli(), "end")
			item.TrackID = observation.TrackDevice
			data, _ := observation.EventItemToDataMap(&item)
			if data == nil {
				data = make(map[string]interface{})
			}
			data[observation.FieldOffline] = 0
			msg := rediscommon.NewSingleItemMessage(addrJC, "", subDeviceType, time.Now().UnixMilli(), "alarm", alarm.AlarmTypeOfflineRecover, data)
			_ = m.streamPublisher.PublishAlarm(pubCtx, msg)
		}()
		// device 上线是常规事件（startup 时一次性触发 60+ 条），降级到 Debug；
		// 真正异常（Offline onset）走 publishHealthIfChanged 仍是 Info（仅 fail 时 log）。
		m.logger.Debug("Device online (first MQTT after subscription create), OfflineRecover published",
			zap.String("device_uid", deviceUID),
			zap.String("device_addr", deviceAddr),
		)
	}

	// monitor 订阅命令由 UpdateLastSeenByType 首条消息逻辑统一发送，此处不再重复
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
		deviceAddr := sub.DeviceAddr
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
		dbStatus, err := m.getDeviceStatus(ctx, deviceAddr)
		if err != nil {
			m.logger.Debug("Failed to get device status for heartbeat check",
				zap.String("device_uid", deviceUID),
				zap.String("device_addr", deviceAddr),
				zap.Error(err),
			)
			continue
		}
		if dbStatus == "disabled" {
			m.logger.Info("Device disabled, removing from subscriptions",
				zap.String("device_uid", deviceUID),
				zap.String("device_addr", deviceAddr),
			)
			m.mu.Lock()
			delete(m.subscriptionsByUID, deviceUID)
			if deviceAddr != "" {
				delete(m.subscriptionsByID, deviceAddr)
			}
			delete(m.prevHealth, deviceUID)
			m.mu.Unlock()
			continue
		}

		timeSinceLastSeen := now.Sub(lastSeen)

		switch status {
		case "online":
			// 180秒无任何消息 → 直接取消 MQTT 订阅（资源清理）
			// 离线判定交由 cardagg 统一处理
			if timeSinceLastSeen > 180*time.Second {
				m.unsubscribeDevice(deviceUID)
			}

		case "unsubscribed":
			// 已取消订阅，需要重新认证才能恢复
		}
	}
}

// DEPRECATED: markDeviceOffline — 离线检测已交由 cardagg 统一处理。
// func (m *DeviceSubscriptionManager) markDeviceOffline(deviceUID string, _ time.Time) { ... }

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
		status, _, deviceType, _ := sub.Status, sub.DeviceAddr, sub.DeviceType, sub.TenantID
		sub.mu.RUnlock()
		if status != "online" {
			continue
		}
		var addrLocal netip.Addr
		if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(context.Background(), deviceUID); dsErr == nil && dsi != nil {
			addrLocal = dsi.DeviceAddr
		}
		go m.streamPublisher.PublishDeviceStatus(context.Background(), addrLocal, deviceUID, deviceType, map[string]int{
			observation.FieldOffline: 0,
		})
		published++
	}
	m.logger.Info("Published online notification for connected devices after startup",
		zap.Int("published", published),
		zap.Int("requested", len(deviceUIDs)),
	)
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
	deviceAddr := sub.DeviceAddr // 先获取 deviceAddr，用于删除索引
	if sub.Status == "offline" {
		sub.Status = "unsubscribed"
		sub.mu.Unlock()

		m.logger.Info("device unsubscribed (180s silence)",
			zap.String("device_uid", deviceUID),
			zap.String("old_status", "offline"),
			zap.String("new_status", "unsubscribed"),
		)

		if m.mqttConsumer != nil {
			go func() {
				m.mqttConsumer.UnsubscribeDeviceTopics(deviceUID)
				m.logger.Info("unsubscribed device mqtt topics", zap.String("device_uid", deviceUID))
			}()
		}

		// 3. 从订阅列表中移除并标记需要重认证（注销内存中的Device）
		m.mu.Lock()
		delete(m.subscriptionsByUID, deviceUID)
		if deviceAddr != "" {
			delete(m.subscriptionsByID, deviceAddr)
		}
		m.unsubscribedDueToTimeout[deviceUID] = struct{}{}
		// 重新订阅时视作首次观测——清掉旧的 transition 缓存，避免上线"假恢复"
		delete(m.prevHealth, deviceUID)
		m.mu.Unlock()
	} else {
		sub.mu.Unlock()
	}
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
		deviceAddr := sub.DeviceAddr
		sub.mu.RUnlock()

		// 检查monitor订阅是否超过最大时长（1小时），在过期前10分钟续订
		renewTime := monitorSubTime.Add(m.monitorMaxAge - 10*time.Minute)
		if now.After(renewTime) {
			// 重新发送monitor订阅命令给设备
			if err := m.mqttPublisher.SubscribeRealtimeData(ctx, deviceUID, m.defaultContent, m.defaultDuration); err != nil {
				m.logger.Warn("Failed to renew monitor subscription command",
					zap.String("device_uid", deviceUID),
					zap.String("device_addr", deviceAddr),
					zap.Error(err),
				)
			} else {
				sub.mu.Lock()
				sub.MonitorSubTime = now
				sub.mu.Unlock()
				m.logger.Info("renewed monitor subscription",
					zap.String("device_uid", deviceUID),
					zap.String("device_addr", deviceAddr),
					zap.Duration("age", now.Sub(monitorSubTime)),
					zap.Int("duration", m.defaultDuration),
				)
			}
		}
	}
}

// getDeviceStatus 获取设备当前状态（仅用于 disabled 检查；on/offline 不存 DB）
//
// Phase 2: devices.access=FALSE → "disabled"（platform 拒绝接入）；TRUE → "active"
// deviceAddr 入参承载 device_uid (logMAC)。
func (m *DeviceSubscriptionManager) getDeviceStatus(ctx context.Context, deviceAddr string) (string, error) {
	var access bool
	query := `SELECT access FROM devices WHERE device_uid = $1`
	err := m.db.QueryRowContext(ctx, query, deviceAddr).Scan(&access)
	if err != nil {
		return "", err
	}
	if !access {
		return "disabled", nil
	}
	return "active", nil
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
// Phase 2: API ID 体系统一为 IPv6；本函数仍兼容 device_uid 直接查询。
//   - 包含 ':' (IPv6 地址) → 从 subscriptionsByID 查找
//   - 否则当作 device_uid（短字符串 logMAC），从 subscriptionsByUID 查找
//
// 不查数据库，完全基于内存
func (m *DeviceSubscriptionManager) GetDeviceOnlineStatus(identifier string) string {
	var sub *DeviceSubscription
	var exists bool
	var deviceUID string

	m.mu.RLock()
	// 自动判断是 device_addr (IPv6) 还是 device_uid (短字符串 logMAC)
	if strings.Contains(identifier, ":") {
		// IPv6 (device_addr)，从 subscriptionsByID 查找
		sub, exists = m.subscriptionsByID[identifier]
		if exists {
			deviceUID = sub.DeviceUID
		}
	} else {
		// device_uid（短字符串），从 subscriptionsByUID 查找
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

// GetDeviceOnlineStatusByDeviceAddr 通过 device_addr (INET) 获取设备的在线状态
// Phase 2: device_id UUID 列退役；入参承载 device_addr，反查 devices.device_uid。
func (m *DeviceSubscriptionManager) GetDeviceOnlineStatusByDeviceAddr(ctx context.Context, deviceAddr string) (string, error) {
	query := `SELECT device_uid FROM devices WHERE device_addr = $1::INET`
	var deviceUID string
	err := m.db.QueryRowContext(ctx, query, deviceAddr).Scan(&deviceUID)
	if err != nil {
		return "offline", err
	}

	return m.GetDeviceOnlineStatus(deviceUID), nil
}

// GetAllDeviceStatuses 批量获取设备在线状态
// tenantID 为空：返回内存中所有设备；否则按 tenant_id 过滤
func (m *DeviceSubscriptionManager) GetAllDeviceStatuses(tenantID string) []domain.DeviceStatusItem {
	m.mu.RLock()
	subs := make([]*DeviceSubscription, 0, len(m.subscriptionsByUID))
	for _, sub := range m.subscriptionsByUID {
		subs = append(subs, sub)
	}
	unsubMap := make(map[string]struct{})
	for u := range m.unsubscribedDueToTimeout {
		unsubMap[u] = struct{}{}
	}
	m.mu.RUnlock()

	result := make([]domain.DeviceStatusItem, 0, len(subs))
	for _, sub := range subs {
		sub.mu.RLock()
		subTenantID := sub.TenantID
		deviceUID := sub.DeviceUID
		deviceAddr := sub.DeviceAddr
		status := sub.Status
		sub.mu.RUnlock()

		if tenantID != "" && subTenantID != tenantID {
			continue
		}
		if _, unsub := unsubMap[deviceUID]; unsub {
			status = "unsubscribed"
		}
		result = append(result, domain.DeviceStatusItem{
			DeviceUID:  deviceUID,
			DeviceAddr: deviceAddr, // Phase 2: deviceAddr 内部承载 device_addr (canonical IPv6 text)
			TenantID:   subTenantID,
			Status:     status,
		})
	}
	return result
}

// SetTCPDeviceOnline marks a TCP device as online in subscription manager memory
// AND publishes online status to Redis cardagg (so wisefido-data API can read it)
func (m *DeviceSubscriptionManager) SetTCPDeviceOnline(deviceUID, deviceAddr string) {
	m.mu.Lock()

	if sub, exists := m.subscriptionsByUID[deviceUID]; exists {
		sub.mu.Lock()
		sub.Status = "online"
		sub.LastSeen = time.Now()
		sub.mu.Unlock()
		m.mu.Unlock()
	} else {
		// Create a minimal subscription entry for TCP device
		sub := &DeviceSubscription{
			DeviceUID: deviceUID,
			DeviceAddr:  deviceAddr,
			Status:    "online",
			LastSeen:  time.Now(),
		}
		m.subscriptionsByUID[deviceUID] = sub
		if deviceAddr != "" {
			m.subscriptionsByID[deviceAddr] = sub
		}
		m.mu.Unlock()
		m.logger.Info("tcp device online", zap.String("uid", deviceUID))
	}

	if deviceAddr != "" && m.streamPublisher != nil {
		var addrTCP netip.Addr
		if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(context.Background(), deviceUID); dsErr == nil && dsi != nil {
			addrTCP = dsi.DeviceAddr
		}
		go m.streamPublisher.PublishDeviceStatus(context.Background(), addrTCP, deviceUID, "", map[string]int{
			observation.FieldOffline: 0,
		})
		m.logger.Info("tcp online published to cardagg",
			zap.String("uid", deviceUID),
			zap.String("device_addr", deviceAddr),
		)
	}
}

// SetTCPDeviceOffline marks a TCP device as offline in subscription manager memory
// AND publishes offline status to Redis cardagg
func (m *DeviceSubscriptionManager) SetTCPDeviceOffline(deviceUID string) {
	var deviceAddr string
	m.mu.Lock()
	if sub, exists := m.subscriptionsByUID[deviceUID]; exists {
		sub.mu.Lock()
		sub.Status = "offline"
		deviceAddr = sub.DeviceAddr
		sub.mu.Unlock()
		m.logger.Info("tcp device offline", zap.String("uid", deviceUID))
	}
	m.mu.Unlock()

	if deviceAddr != "" && m.streamPublisher != nil {
		var addrTCP netip.Addr
		if dsi, dsErr := m.deviceRepo.GetDeviceStoreInfo(context.Background(), deviceUID); dsErr == nil && dsi != nil {
			addrTCP = dsi.DeviceAddr
		}
		go m.streamPublisher.PublishDeviceStatus(context.Background(), addrTCP, deviceUID, "", map[string]int{
			observation.FieldOffline: 1,
		})
		m.logger.Info("tcp offline published to cardagg", zap.String("uid", deviceUID))
	}
}
