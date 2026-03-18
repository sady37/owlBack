package service

import (
	"context"
	"strings"

	"owl-common/alarm"
)

// CardHasRoomRadar 该卡是否绑有房间雷达（非卫生间雷达）。无房间雷达时 RoomState.TotalPeople 置 -1 表示不维护。
func CardHasRoomRadar(ctx context.Context, meta *CardMeta, tenantID string, enablement *AlarmEnablementCache) bool {
	if meta == nil || len(meta.Devices) == 0 {
		return false
	}
	for deviceID := range meta.Devices {
		dm := meta.Devices[deviceID]
		if dm == nil {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if !strings.Contains(t, "radar") {
			continue
		}
		if !IsBathroomDevice(ctx, meta, deviceID, tenantID, enablement) {
			return true
		}
	}
	return false
}

// IsBathroomDevice 判断该卡下该设备是否按卫生间维护 RoomState（创建卡/事件时用）。
// a) RoomName 含 WC/Toilet/Bathroom → ClassifyRoomType == "bathroom"
// b) 使能表开启 Stay(EventStay) → 视为卫生间（一房一 Radar 才开 Stay）
func IsBathroomDevice(ctx context.Context, meta *CardMeta, deviceID, tenantID string, enablement *AlarmEnablementCache) bool {
	if meta == nil || deviceID == "" {
		return false
	}
	dm := meta.Devices[deviceID]
	if dm != nil && strings.TrimSpace(dm.BoundRoomName) != "" && ClassifyRoomType(dm.BoundRoomName) == "bathroom" {
		return true
	}
	if enablement != nil && strings.TrimSpace(tenantID) != "" {
		if _, ok := enablement.IsEnabled(ctx, tenantID, deviceID, alarm.Stay); ok {
			return true
		}
	}
	return false
}
