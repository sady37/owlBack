package service

import (
	"fmt"
	"strings"
)

func metaLookup(meta *CardMeta, deviceID string) *DeviceMeta {
	if meta == nil || meta.Devices == nil {
		return nil
	}
	return meta.Devices[deviceID]
}

func IsDeviceHealthy(dm *DeviceMeta) bool {
	if dm == nil || !dm.DeviceAddr.IsValid() {
		return false
	}
	for _, k := range []string{"detached", "angle_abnormal", "signal_poor"} {
		if dm.RuntimeStatus != nil && dm.RuntimeStatus[k] == "1" {
			return false
		}
	}
	return true
}

func isSleepad(deviceType string) bool {
	t := strings.ToLower(deviceType)
	return strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad")
}

func toInt(v any) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		var i int
		fmt.Sscanf(x, "%d", &i)
		return i
	default:
		return 0
	}
}

func toInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return int64(toInt(v))
	}
}

func boolFromAny(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	default:
		return false
	}
}
