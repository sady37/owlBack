package subscriber

import (
	"context"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	"wisefido-sleepace/internal/consumer"
	"wisefido-sleepace/internal/service"

	"go.uber.org/zap"
)

// healthCheckInterval = polling 周期。
//
// 选 80s 的原因：
//   - cardagg 90s TTL 之内必须能补一次"在线"信号，否则离床期 mode=1 sleepad 会被误判 offline
//   - 80s × 14 台 ≈ 每 6 秒一次厂家 API 调用，无压力
//   - 留 10s 冗余给 API 抖动
const healthCheckInterval = 80 * time.Second

// probeMinInterval = on-demand probe 的最小间隔。
//
// 启动时 config:card:stream 会一次性放出全部 backlog（曾观测 28 条）。
// 旧实现里 deviceID="" 的 cardChange 会 fallback 到 scanAll，叠加 14 台 Sleepad
// 形成同一 UID 1s 内被探 17 次的风暴，对厂家 API 是 ~390 RPS 抖动。
//
// per-UID 30s 去重：同一台设备在 30s 内只接受一次 probe，超过即跳过。
// 80s ticker 不受此限（它本就是 80s 间隔），所以 mode=1 离床期续命不受影响。
const probeMinInterval = 30 * time.Second

// HealthCheck Sleepace 连接态探测（与 qinglan subscriber/health_check 命名对齐）：config.card 触发 + 周期兜底。
// 仅依赖 CardMappingService：用 device_id / device_uid 解析 baseline（含 device_code、三权限），不再直接持有 CardDB。
//
// 去重：lastPublished 缓存上次发布的在线状态；周期 probe 仅在状态翻转时才 publish 到 iot:alarm:stream，
// 避免离线设备每 80s 灌一条 Offline 告警（cardagg.InsertAlarmAndUpdateCard 无去重）。
type HealthCheck struct {
	CardMapping   *service.CardMappingService
	SleepaceAPI   *service.SleepaceAPI
	Publisher     *consumer.StreamPublisher
	StatusTracker *service.DeviceStatusTracker
	CardDB        *card.CardDB // 可选：device_store.firmware_version 同步用
	Logger        *zap.Logger
	DeviceTypes   []string

	mu              sync.Mutex
	lastPublished   map[string]bool      // device_uid → 上次 publish 时的 online 状态（true=online）
	lastPublishedOK map[string]bool      // 是否曾经成功 publish 过；首次必发
	lastProbedAt    map[string]time.Time // device_uid → 上次实际进入 probeOne 的时间，用于 on-demand 去重
	lastScanAllAt   time.Time            // 上次 scanAll 进入时间，用于 cardChange fallback 去重
}

func NewHealthCheck(
	cardMapping *service.CardMappingService,
	sleepaceAPI *service.SleepaceAPI,
	publisher *consumer.StreamPublisher,
	statusTracker *service.DeviceStatusTracker,
	logger *zap.Logger,
) *HealthCheck {
	return &HealthCheck{
		CardMapping:     cardMapping,
		SleepaceAPI:     sleepaceAPI,
		Publisher:       publisher,
		StatusTracker:   statusTracker,
		Logger:          logger,
		lastPublished:   make(map[string]bool),
		lastPublishedOK: make(map[string]bool),
		lastProbedAt:    make(map[string]time.Time),
	}
}

func (h *HealthCheck) deviceTypes() []string {
	if len(h.DeviceTypes) > 0 {
		return h.DeviceTypes
	}
	return []string{"Sleepad", "SleepPad", "sleepad"}
}

func (h *HealthCheck) isSleepadType(t string) bool {
	t = strings.TrimSpace(t)
	if t == "" {
		return false
	}
	for _, x := range h.deviceTypes() {
		if t == x {
			return true
		}
	}
	return false
}

// Run 周期全量探测兜底；启动后先做一次 scanAll，避免新添加的 firmware_version sync 要等到第一个 tick。
//
// 周期 = healthCheckInterval（80s），需小于 cardagg PruneStaleDevices TTL（90s），
// 这样 mode=1 离床期 sleepad 不会被 cardagg 误判 offline——本 worker 在 TTL 命中前补一次 online 信号。
func (h *HealthCheck) Run(ctx context.Context) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	tick := time.NewTicker(healthCheckInterval)
	defer tick.Stop()
	h.Logger.Info("sleepace health_check started", zap.Duration("interval", healthCheckInterval))
	// 启动 + ticker 走 scanAllNow，不受 cardChange 路径的去重影响——保证 80s 续命节奏
	h.scanAllNow(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			h.scanAllNow(ctx)
		}
	}
}

// ProbeAfterCardChange 收到 config.card：优先按 affected UIDs / device_id 解析并探测单设备，
// 仅当无法定位到具体 Sleepad 时才 scanAll。scanAll/probeOne 内置去重（probeMinInterval）防止风暴。
//
// affectedUIDs 来自 config payload 的 affected_device_uids；非空时只探这些 Sleepad，
// 避免一条 cardChange 触发对所有 14 台设备的全量重扫（启动时 backlog 28 条 → 同一 UID 1s 17 次的来源）。
func (h *HealthCheck) ProbeAfterCardChange(ctx context.Context, tenantID, unitID, cardID, deviceID string, affectedUIDs []string) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	h.Logger.Info("health_check probe after card change",
		zap.String("card_id", cardID),
		zap.String("device_id", deviceID),
		zap.Int("affected_uids", len(affectedUIDs)))

	// 优先按 affected_device_uids：只解析为 Sleepad 的探测，非 Sleepad 直接跳过
	if len(affectedUIDs) > 0 {
		probed := 0
		for _, uid := range affectedUIDs {
			if uid == "" {
				continue
			}
			b, ok := h.CardMapping.ResolveBaseline(ctx, uid)
			if !ok || !h.isSleepadType(b.DeviceType) || b.DeviceCode == "" {
				continue
			}
			h.probeOne(ctx, &b)
			probed++
		}
		if probed > 0 {
			return
		}
		// affected 全是非 Sleepad，不需要 fallback 到 scanAll
		return
	}

	if deviceID != "" {
		if b, ok := h.CardMapping.ResolveBaseline(ctx, deviceID); ok && h.isSleepadType(b.DeviceType) && b.DeviceCode != "" {
			h.probeOne(ctx, &b)
			return
		}
	}
	// 没有具体设备目标（reset 或解析失败）→ scanAll，但内置 probeMinInterval 去重防风暴
	h.scanAll(ctx)
}

// scanAll 全量探测当前所有 Sleepad。
//
// 对外暴露 force 参数：80s 周期 ticker 调用 scanAllNow（不去重；保证续命节奏），
// cardChange fallback 调用 scanAll（去重；防止 backlog 风暴）。
func (h *HealthCheck) scanAll(ctx context.Context) {
	h.mu.Lock()
	if !h.lastScanAllAt.IsZero() && time.Since(h.lastScanAllAt) < probeMinInterval {
		skippedAt := h.lastScanAllAt
		h.mu.Unlock()
		h.Logger.Debug("health_check scanAll debounced",
			zap.Duration("since_last", time.Since(skippedAt)))
		return
	}
	h.lastScanAllAt = time.Now()
	h.mu.Unlock()
	h.scanAllNow(ctx)
}

func (h *HealthCheck) scanAllNow(ctx context.Context) {
	h.mu.Lock()
	h.lastScanAllAt = time.Now()
	h.mu.Unlock()
	list, err := h.CardMapping.ListSleepadBaselinesForHealth(ctx, h.deviceTypes())
	if err != nil {
		h.Logger.Warn("health_check list baselines", zap.Error(err))
		return
	}
	for i := range list {
		b := &list[i]
		if b.DeviceCode == "" {
			continue
		}
		h.probeOne(ctx, b)
	}
}

func (h *HealthCheck) probeOne(ctx context.Context, b *card.DeviceBaseline) {
	deviceType := b.DeviceType
	if deviceType == "" {
		deviceType = "Sleepad"
	}
	// per-UID 去重：on-demand 路径（cardChange / probeAfterCardChange）+ scanAll 共用此闸门，
	// 同一 UID 30s 内只允许进入一次 vendor API。80s ticker 自然 > 30s，不被卡住。
	h.mu.Lock()
	if last, ok := h.lastProbedAt[b.DeviceUID]; ok && time.Since(last) < probeMinInterval {
		h.mu.Unlock()
		h.Logger.Debug("health_check probe debounced",
			zap.String("device_uid", b.DeviceUID),
			zap.Duration("since_last", time.Since(last)))
		return
	}
	h.lastProbedAt[b.DeviceUID] = time.Now()
	h.mu.Unlock()

	status, err := h.SleepaceAPI.GetConnectionStatus(b.DeviceCode)
	if err != nil {
		h.Logger.Warn("health_check GetConnectionStatus failed",
			zap.String("device_uid", b.DeviceUID),
			zap.String("device_code", b.DeviceCode),
			zap.Error(err))
		return
	}
	online := status != 0
	if h.StatusTracker != nil {
		h.StatusTracker.UpdateConnection(b.DeviceUID, online)
	}
	h.Logger.Info("health_check connectionStatus",
		zap.String("device_uid", b.DeviceUID),
		zap.String("device_code", b.DeviceCode),
		zap.Bool("online", online))
	// 在线时顺手 sync device_store.firmware_version；MQTT connectionStatus 只在 TCP 状态变化时推一次，
	// 已在线设备不会重新触发 sync，所以这里 probe 时也补一刀。函数自带 current==reported 短路，幂等。
	if online && h.CardDB != nil {
		go h.syncDeviceStoreVersion(context.Background(), b.DeviceCode)
	}
	if b.DeviceID == "" {
		return
	}

	// 非对称 dedup：
	//   - online (OfflineRecover) 永不 dedup —— cardagg 端 KeepAlive 续命依赖周期性收到 OfflineRecover；
	//     OfflineRecover 在 cardagg 端是幂等的（HandleRecoveryWithTypes 对无 active alarm 是 noop，
	//     SetDeviceOnline 写 Redis 也幂等），重复发无副作用。
	//   - offline (Offline) 仅在状态翻转时 publish —— InsertAlarmAndUpdateCard 无内部去重，
	//     重复发会灌入 alarm_events 表 + 重复 push 通知。首次离线、online→offline 翻转时才发。
	if !online {
		h.mu.Lock()
		prev, hadPrev := h.lastPublished[b.DeviceUID], h.lastPublishedOK[b.DeviceUID]
		if hadPrev && prev == false {
			h.mu.Unlock()
			return
		}
		h.lastPublished[b.DeviceUID] = false
		h.lastPublishedOK[b.DeviceUID] = true
		h.mu.Unlock()
	} else {
		// online 总是发，仅记录状态供下次 offline 判翻转用。
		h.mu.Lock()
		h.lastPublished[b.DeviceUID] = true
		h.lastPublishedOK[b.DeviceUID] = true
		h.mu.Unlock()
	}

	nowMs := time.Now().UnixMilli()
	tsMs := nowMs
	var eventName string
	var item observation.EventItem
	if online {
		eventName = alarm.AlarmTypeOfflineRecover
		item = observation.EventItem{
			EventSince:  tsMs,
			EventStatus: "end",
			EventValue:  0,
			TrackID:     observation.TrackDevice,
		}
	} else {
		eventName = alarm.AlarmTypeOffline
		item = observation.EventItem{
			EventSince:  tsMs,
			EventStatus: "start",
			EventValue:  1,
			TrackID:     observation.TrackDevice,
		}
	}
	alarmData, _ := observation.EventItemToDataMap(&item)
	if alarmData == nil {
		alarmData = make(map[string]any)
	}
	if online {
		alarmData[observation.FieldOffline] = 0
	} else {
		alarmData[observation.FieldOffline] = 1
	}
	msg := rediscommon.NewIoTStreamMessageWithData(b.TenantID, b.CardID, b.DeviceUID, b.DeviceID, deviceType, nowMs, "alarm", eventName, alarmData)
	if err := h.Publisher.PublishAlarm(ctx, msg); err != nil {
		h.Logger.Warn("connectionStatus publish alarm failed", zap.String("device_uid", b.DeviceUID), zap.Bool("online", online), zap.Error(err))
		// publish 失败时回滚 lastPublishedOK，下次再试；offline 路径会重发；online 本来就每次发
		h.mu.Lock()
		h.lastPublishedOK[b.DeviceUID] = false
		h.mu.Unlock()
	}
}

// syncDeviceStoreVersion 跟 mqtt_consumer 同名函数同语义：拉厂家 deviceInfo → 写 device_store.firmware_version。
// 与 mqtt_consumer 版本独立实现避免循环依赖（subscriber → consumer → subscriber）。
func (h *HealthCheck) syncDeviceStoreVersion(ctx context.Context, deviceCode string) {
	if deviceCode == "" || h.CardDB == nil || h.SleepaceAPI == nil {
		return
	}
	info, err := h.SleepaceAPI.GetDeviceInfoByDeviceId(deviceCode)
	if err != nil {
		h.Logger.Debug("health_check version sync: get device info", zap.String("device_code", deviceCode), zap.Error(err))
		return
	}
	if info == nil || info.Version == "" {
		return
	}
	if err := h.CardDB.UpdateDeviceStoreReportedVersion(ctx, deviceCode, info.Version); err != nil {
		h.Logger.Warn("health_check version sync: update device_store", zap.String("device_code", deviceCode), zap.String("version", info.Version), zap.Error(err))
		return
	}
	h.Logger.Info("device_store version synced (health_check)", zap.String("device_code", deviceCode), zap.String("version", info.Version))
}
