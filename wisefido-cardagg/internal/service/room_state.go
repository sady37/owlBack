package service

import (
	"context"
	"sort"
	"strings"

	"owl-common/alarm"
)

/*
卫生间（BathRoomState）/ Stay 检测链——说明摘要

1）什么房间会走这条链？
   - 绑定房间的名称经 ClassifyRoomType 判为 bathroom（wc/bathroom/restroom/toilet 等）→ 该设备始终按卫生间层处理。
   - 否则：仅当告警使能里对该设备开启了 Stay，且该设备解析到的 room 在租户 beds 表中「无床」（BoundRoomHasBed=false）时，才允许用 Stay 走卫生间链；有床的 room 不走，避免与床态、房间人数语义冲突。

2）「风险」相关状态在哪体现？
   - BathRoomState / RoomState 的 HasMulti、HasRisk、TotalPeople、进出时间、站立累计、StaySec 等由 state_service 里 Publish* 与 UpdateStateFromActivity 按事件合并；Stay 告警 pending 由 event_handler 在进出窗口内结合人数触发。

3）怎么检测与绑定设备？
   - device_meta 加载 cards.devices 后查 rooms、beds 得到 BoundRoomName、BoundRoomHasBed。
   - IsBathroomDevice 决定事件进卫生间路径还是房间路径；Init 用 PickBathroomDeviceIDForInit 按「bathroom+Stay > Stay（无床）> 仅 bathroom」且 device_id 有序选定唯一 BathRoomState 绑定设备；Redis 已绑定时仅匹配该 device 的写才生效（见 UpdateStateFromActivity）。
*/

// SortedDeviceIDs 稳定排序 device_id，避免 map 遍历随机。
func SortedDeviceIDs(meta *CardMeta) []string {
	if meta == nil || len(meta.Devices) == 0 {
		return nil
	}
	ids := make([]string, 0, len(meta.Devices))
	for id := range meta.Devices {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

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

func deviceStayEnabled(ctx context.Context, tenantID, deviceID string, enablement *AlarmEnablementCache) bool {
	if enablement == nil || strings.TrimSpace(tenantID) == "" || deviceID == "" {
		return false
	}
	_, ok := enablement.IsEnabled(ctx, tenantID, deviceID, alarm.Stay)
	return ok
}

// PickBathroomDeviceIDForInit 一卡一条 BathRoomState：bathroom+Stay > Stay（绑定 room 无床）> 仅 bathroom；同档内 SortedDeviceIDs 序。
func PickBathroomDeviceIDForInit(ctx context.Context, meta *CardMeta, tenantID string, enablement *AlarmEnablementCache) string {
	if meta == nil {
		return ""
	}
	ids := SortedDeviceIDs(meta)
	for _, deviceID := range ids {
		dm := meta.Devices[deviceID]
		if dm == nil || ClassifyRoomType(dm.BoundRoomName) != "bathroom" {
			continue
		}
		if deviceStayEnabled(ctx, tenantID, deviceID, enablement) {
			return deviceID
		}
	}
	for _, deviceID := range ids {
		dm := meta.Devices[deviceID]
		if dm == nil {
			continue
		}
		if deviceStayEnabled(ctx, tenantID, deviceID, enablement) && !dm.BoundRoomHasBed {
			return deviceID
		}
	}
	for _, deviceID := range ids {
		dm := meta.Devices[deviceID]
		if dm == nil {
			continue
		}
		if ClassifyRoomType(dm.BoundRoomName) == "bathroom" {
			return deviceID
		}
	}
	return ""
}

// IsBathroomDevice 是否走卫生间事件/Stay 检测路径：bathroom 名恒真；否则须 Stay 且绑定 room 无床。
func IsBathroomDevice(ctx context.Context, meta *CardMeta, deviceID, tenantID string, enablement *AlarmEnablementCache) bool {
	if meta == nil || deviceID == "" {
		return false
	}
	dm := meta.Devices[deviceID]
	if dm == nil {
		return false
	}
	if ClassifyRoomType(dm.BoundRoomName) == "bathroom" {
		return true
	}
	if dm.BoundRoomHasBed {
		return false
	}
	return deviceStayEnabled(ctx, tenantID, deviceID, enablement)
}

// BathroomRoomFieldsFromDevice 写入 BathRoomState.room_id / room_name（与 device_meta 解析一致）。
func BathroomRoomFieldsFromDevice(dm *DeviceMeta) (roomID, roomName string) {
	if dm == nil {
		return "", ""
	}
	return dm.EffectiveRoomID, dm.BoundRoomName
}
