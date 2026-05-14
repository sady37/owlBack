package consumer

import (
	"context"
	"encoding/json"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type AlarmDeviceHandler struct {
	enablement *service.AlarmEnablementCache
	alarms     *service.AlarmService
	logger     *zap.Logger
}

func NewAlarmDeviceHandler(enablement *service.AlarmEnablementCache, alarms *service.AlarmService, logger *zap.Logger) *AlarmDeviceHandler {
	return &AlarmDeviceHandler{enablement: enablement, alarms: alarms, logger: logger}
}

type alarmDeviceData struct {
	DeviceAddr  string `json:"device_addr"` // canonical IPv6；cache key (单程票后取代 device_id UUID)
	SettingType string `json:"setting_type"`
}

func (h *AlarmDeviceHandler) Handle(ctx context.Context, msg interface{}) error {
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}

	dataStr, ok := streamMsg["data"].(string)
	if !ok {
		return nil
	}

	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		h.logger.Warn("parse cloud events", zap.Error(err))
		return nil
	}

	var d alarmDeviceData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		h.logger.Warn("parse alarm device data", zap.Error(err))
		return nil
	}

	if d.DeviceAddr == "" {
		return nil
	}

	h.enablement.Invalidate(d.DeviceAddr)

	// 配置变更后，对该 device 的 pending 做一次清理：
	// 凡是新配置里已 disabled 的 alarm_type，旧 pending 一律删除（参考厂家"关闭报警清除统计"逻辑）。
	// device_ipv6 单程票后 tenant 由 addr.Prefix(48) 派生（PurgeDisabledPendingForDevice 内部 Redis key 仍走 addr 字符串）
	if h.alarms != nil {
		// addr.Prefix(48).String() 派生 tenantPref；这里直接传 addr 双 key 兼容（Redis hash field 仅按字符串命中）
		h.alarms.PurgeDisabledPendingForDevice(ctx, "", d.DeviceAddr)
	}

	h.logger.Info("alarm device config invalidated",
		zap.String("device_addr", d.DeviceAddr),
		zap.String("setting", d.SettingType))

	return nil
}
