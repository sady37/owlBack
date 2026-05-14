package service

import (
	"strings"
)

// Affinity — 设备间亲和性评分（保留供 AI 层使用）。权重常量见 weights.go。

// ComputeAffinity returns the affinity score between two devices
// based on spatial proximity and functional relationship.
func ComputeAffinity(a, b *DeviceMeta) int {
	if a == nil || b == nil {
		return 0
	}

	spatial := 0
	aBed, bBed := a.BedPref(), b.BedPref()
	aRoom, bRoom := a.RoomPref(), b.RoomPref()
	aUnit, bUnit := a.UnitPref(), b.UnitPref()
	if aBed != "" && bBed != "" && aBed == bBed {
		spatial = SpatialSameBed
	} else if aRoom != "" && bRoom != "" && aRoom == bRoom {
		spatial = SpatialSameRoom
	} else if aUnit != "" && bUnit != "" && aUnit == bUnit {
		spatial = SpatialSameUnit
	}

	functional := 0
	aType := strings.ToLower(a.DeviceType)
	bType := strings.ToLower(b.DeviceType)
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
