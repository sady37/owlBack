package service

import (
	"strings"

	"owl-common/redis"
)

// Affinity — 设备间亲和性评分（保留供 AI 层使用）。权重常量见 weights.go。

// ComputeAffinity returns the affinity score between two devices
// based on spatial proximity and functional relationship.
func ComputeAffinity(a, b *DeviceMeta) int {
	if a == nil || b == nil {
		return 0
	}

	spatial := 0
	if a.BoundBedID != "" && b.BoundBedID != "" && a.BoundBedID == b.BoundBedID {
		spatial = SpatialSameBed
	} else if a.BoundRoomID != "" && b.BoundRoomID != "" && a.BoundRoomID == b.BoundRoomID {
		spatial = SpatialSameRoom
	} else if a.UnitID != "" && b.UnitID != "" && a.UnitID == b.UnitID {
		spatial = SpatialSameUnit
	}

	functional := 0
	// 用 BaseDeviceType 解 AI 后缀，避免 "Radar.AI01" vs "Radar" 被误判为异构
	aType := strings.ToLower(redis.BaseDeviceType(a.DeviceType))
	bType := strings.ToLower(redis.BaseDeviceType(b.DeviceType))
	if aType == bType {
		if aType == "radar" {
			functional = FuncFullMatch
		} else {
			functional = FuncSleepMatch
		}
	} else {
		functional = FuncComplement
	}

	return spatial + functional
}
