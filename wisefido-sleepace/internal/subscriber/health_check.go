package subscriber

import (
	"context"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	"wisefido-sleepace/internal/consumer"
	"wisefido-sleepace/internal/service"

	"go.uber.org/zap"
)

// HealthCheck Sleepace 连接态探测（与 qinglan subscriber/health_check 命名对齐）：config.card 触发 + 10 分钟兜底。
// 仅依赖 CardMappingService：用 device_id / device_uid 解析 baseline（含 device_code、三权限），不再直接持有 CardDB。
type HealthCheck struct {
	CardMapping   *service.CardMappingService
	SleepaceAPI   *service.SleepaceAPI
	Publisher     *consumer.StreamPublisher
	StatusTracker *service.DeviceStatusTracker
	CardDB        *card.CardDB // 可选：device_store.firmware_version 同步用
	Logger        *zap.Logger
	DeviceTypes   []string
}

func NewHealthCheck(
	cardMapping *service.CardMappingService,
	sleepaceAPI *service.SleepaceAPI,
	publisher *consumer.StreamPublisher,
	statusTracker *service.DeviceStatusTracker,
	logger *zap.Logger,
) *HealthCheck {
	return &HealthCheck{
		CardMapping:   cardMapping,
		SleepaceAPI:   sleepaceAPI,
		Publisher:     publisher,
		StatusTracker: statusTracker,
		Logger:        logger,
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

// Run 仅 10 分钟全量探测兜底；启动后先做一次 scanAll，避免新添加的 firmware_version sync 要等到第一个 10min tick。
func (h *HealthCheck) Run(ctx context.Context) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	tick := time.NewTicker(10 * time.Minute)
	defer tick.Stop()
	h.Logger.Info("sleepace health_check started", zap.Duration("full_probe", 10*time.Minute))
	h.scanAll(ctx) // startup probe
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			h.scanAll(ctx)
		}
	}
}

// ProbeAfterCardChange 收到 config.card：优先按 device_id 解析并探测单设备，否则 scanAll 兜底。
func (h *HealthCheck) ProbeAfterCardChange(ctx context.Context, tenantID, unitID, cardID, deviceID string) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	h.Logger.Info("health_check probe after card change",
		zap.String("card_id", cardID),
		zap.String("device_id", deviceID))

	if deviceID != "" {
		if b, ok := h.CardMapping.ResolveBaseline(ctx, deviceID); ok && h.isSleepadType(b.DeviceType) && b.DeviceCode != "" {
			h.probeOne(ctx, &b)
			return
		}
	}
	h.scanAll(ctx)
}

func (h *HealthCheck) scanAll(ctx context.Context) {
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
	if b.DeviceID != "" {
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
		}
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
