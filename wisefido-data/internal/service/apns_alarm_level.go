package service

import (
	"strings"

	"owl-common/observation"
)

// AlarmLevelStringToPushIndex 将 alarm_events.alarm_level 字符串映射为 APNs payload 用的 0..4
func AlarmLevelStringToPushIndex(level string) int {
	switch strings.ToUpper(observation.NormalizeEventLevel(strings.TrimSpace(level))) {
	case "EMERG":
		return 0
	case "ALERT":
		return 1
	case "CRITICAL":
		return 2
	case "ERROR":
		return 3
	case "WARNING", "WARN":
		return 4
	default:
		return 4
	}
}
