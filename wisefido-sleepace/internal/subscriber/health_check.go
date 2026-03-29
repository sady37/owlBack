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

// Run 仅 10 分钟全量探测兜底。
func (h *HealthCheck) Run(ctx context.Context) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	tick := time.NewTicker(10 * time.Minute)
	defer tick.Stop()
	h.Logger.Info("sleepace health_check started", zap.Duration("full_probe", 10*time.Minute))
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			h.scanAll(ctx)
		}
	}
}

// ProbeAfterCardChange 收到 config.card：优先 device_id；否则 tenant+unit 或 card_id 枚举设备键（再按 device_id/device_uid 解析）；无 device_id 时可将同一 UUID 当作 card_id 再枚举。
func (h *HealthCheck) ProbeAfterCardChange(ctx context.Context, tenantID, unitID, cardID, deviceID string) {
	if h.CardMapping == nil || h.SleepaceAPI == nil || h.Publisher == nil {
		return
	}
	h.Logger.Info("health_check probe after card change",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.String("card_id", cardID),
		zap.String("device_id", deviceID))

	if deviceID != "" {
		if b, ok := h.CardMapping.ResolveBaseline(ctx, deviceID); ok && h.isSleepadType(b.DeviceType) && b.DeviceCode != "" {
			h.probeOne(ctx, &b)
			return
		}
		dids, uids := h.CardMapping.DeviceKeysInCard(ctx, deviceID)
		if len(dids)+len(uids) > 0 {
			h.probeKeys(ctx, dids, uids)
			return
		}
		h.Logger.Debug("health_check: device_id not resolved as device or card", zap.String("device_id", deviceID))
		return
	}
	if tenantID != "" && unitID != "" {
		dids, uids := h.CardMapping.DeviceKeysInTenantUnit(ctx, tenantID, unitID)
		h.probeKeys(ctx, dids, uids)
		return
	}
	if cardID != "" {
		dids, uids := h.CardMapping.DeviceKeysInCard(ctx, cardID)
		h.probeKeys(ctx, dids, uids)
		return
	}
	h.scanAll(ctx)
}

func (h *HealthCheck) probeKeys(ctx context.Context, dids, uids []string) {
	seen := make(map[string]struct{})
	for _, k := range append(dids, uids...) {
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		if b, ok := h.CardMapping.ResolveBaseline(ctx, k); ok && h.isSleepadType(b.DeviceType) && b.DeviceCode != "" {
			h.probeOne(ctx, &b)
		}
	}
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
	if b.DeviceID != "" {
		nowMs := time.Now().UnixMilli()
		tsMs := nowMs
		var eventName string
		var item observation.EventItem
		if online {
			eventName = alarm.AlarmTypeOfflineRecover
			item = observation.EventItem{
				DataCategory: eventName,
				EventName:    eventName,
				EventSince:   tsMs,
				EventStatus:  "end",
				EventValue:   0,
				TrackID:      observation.TrackDevice,
			}
		} else {
			eventName = alarm.AlarmTypeOffline
			item = observation.EventItem{
				DataCategory: eventName,
				EventName:    eventName,
				EventSince:   tsMs,
				EventStatus:  "start",
				EventValue:   1,
				TrackID:      observation.TrackDevice,
			}
		}
		alarmData, _ := observation.EventItemToDataMap(&item)
		if alarmData == nil {
			alarmData = make(map[string]any)
		}
		alarmData[observation.FieldEventName] = eventName
		alarmData[rediscommon.DataCategoryKey] = eventName
		if online {
			alarmData[observation.FieldOffline] = 0
		} else {
			alarmData[observation.FieldOffline] = 1
		}
		msg := rediscommon.NewIoTStreamMessageWithData(b.TenantID, b.EffectiveCardID(), b.DeviceUID, b.DeviceID, deviceType, nowMs, "alarm", eventName, alarmData)
		if err := h.Publisher.PublishAlarm(ctx, msg); err != nil {
			h.Logger.Warn("connectionStatus publish alarm failed", zap.String("device_uid", b.DeviceUID), zap.Bool("online", online), zap.Error(err))
		}
	}
}
