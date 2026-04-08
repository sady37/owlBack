package service

import (
	"strconv"
	"strings"

	"owl-common/alarm"
	"owl-common/observation"
)

// AlarmLevelStringToPushIndex 将 alarm_events.alarm_level 映射为 APNs payload 用的 0..4。
// 与 owl-common/alarm.AlarmLevelInt* 一致：0=EMERG, 1=ALERT, 2=CRITICAL, 3=ERROR, 4=WARNING；
// 库中亦存数字字符串 '0'..'4'（见 card/alarm_db 统计 SQL）。
func AlarmLevelStringToPushIndex(level string) int {
	s := strings.TrimSpace(level)
	if n, err := strconv.Atoi(s); err == nil {
		if n >= alarm.AlarmLevelIntEmerg && n <= alarm.AlarmLevelIntWarning {
			return n
		}
	}
	switch strings.ToUpper(alarm.NormalizeAlarmLevel(observation.NormalizeEventLevel(s))) {
	case alarm.AlarmLevelEmerg:
		return alarm.AlarmLevelIntEmerg
	case alarm.AlarmLevelAlert:
		return alarm.AlarmLevelIntAlert
	case alarm.AlarmLevelCrit:
		return alarm.AlarmLevelIntCrit
	case alarm.AlarmLevelErr:
		return alarm.AlarmLevelIntErr
	case alarm.AlarmLevelWarn, "WARN":
		return alarm.AlarmLevelIntWarning
	default:
		return alarm.AlarmLevelIntWarning
	}
}
