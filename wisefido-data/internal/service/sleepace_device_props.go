package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"owl-common/alarm"

	"wisefido-data/internal/domain"
)

// ConvertHardwareResponseToAlarmItems converts sleepad cloud API camelCase response
// to []alarm.AlarmItem. Only populates IsEnabled (from flags) and AlarmParams (from thresholds).
// AlarmLevel is NOT set here — the hardware only knows on/off, not the level.
// Caller must merge AlarmLevel from DB.
func ConvertHardwareResponseToAlarmItems(hw map[string]interface{}) []alarm.AlarmItem {
	defaults := alarm.GetDefaultAlarmItems("sleepad")
	if defaults == nil {
		return nil
	}

	reverseParamKeys := make(map[string]map[string]string)
	for at, pm := range alarmTypeParamMapping {
		rev := make(map[string]string)
		for generic, flat := range pm {
			rev[flat] = generic
		}
		reverseParamKeys[at] = rev
	}

	camelToFlat := map[string]string{
		"leftBedStartHour": "left_bed_start_hour", "leftBedStartMinute": "left_bed_start_minute",
		"leftBedEndHour": "left_bed_end_hour", "leftBedEndMinute": "left_bed_end_minute",
		"leftBedDuration": "left_bed_duration", "leftBedFlag": "left_bed_alarm_level",
		"minHeartRate": "min_heart_rate", "maxHeartRate": "max_heart_rate",
		"heartRateSlowDuration": "heart_rate_slow_duration", "heartRateFastDuration": "heart_rate_fast_duration",
		"heartRateSlowFlag": "heart_rate_slow_alarm_level", "heartRateFastFlag": "heart_rate_fast_alarm_level",
		"minBreathRate": "min_breath_rate", "maxBreathRate": "max_breath_rate",
		"breathRateSlowDuration": "breath_rate_slow_duration", "breathRateFastDuration": "breath_rate_fast_duration",
		"breathRateSlowFlag": "breath_rate_slow_alarm_level", "breathRateFastFlag": "breath_rate_fast_alarm_level",
		"breathPauseDuration": "breath_pause_duration", "breathPauseFlag": "breath_pause_alarm_level",
		"bodyMoveDuration": "body_move_duration", "bodyMoveFlag": "body_move_alarm_level",
		"nobodyMoveDuration": "nobody_move_duration", "nobodyMoveFlag": "nobody_move_alarm_level",
		"noTurnOverDuration": "no_turn_over_duration", "noTurnOverFlag": "no_turn_over_alarm_level",
		"situpFlag":     "situp_alarm_level",
		"onbedDuration": "onbed_duration", "onbedFlag": "onbed_alarm_level",
	}

	flat := make(map[string]interface{})
	for camel, snake := range camelToFlat {
		v, ok := hw[camel]
		if !ok {
			continue
		}
		if isFlag(snake) {
			flat[snake] = toFloat(v) == 1
		} else {
			flat[snake] = toInt(v)
		}
	}

	result := make([]alarm.AlarmItem, 0, len(defaults))
	for _, item := range defaults {
		out := item
		// AlarmLevel is intentionally left as default — must be merged from DB

		if levelKeys, ok := alarmTypeToSleepaceLevelKeys[item.AlarmType]; ok {
			anyEnabled := false
			for _, lk := range levelKeys {
				if enabled, exists := flat[lk]; exists {
					if b, ok := enabled.(bool); ok && b {
						anyEnabled = true
					}
				}
			}
			if anyEnabled {
				on := alarm.IsEnabledOn
				out.IsEnabled = &on
			} else {
				off := alarm.IsEnabledOff
				out.IsEnabled = &off
			}
		}

		if rev, ok := reverseParamKeys[item.AlarmType]; ok {
			params := make(map[string]interface{})
			for flatKey, genericKey := range rev {
				if v, exists := flat[flatKey]; exists {
					params[genericKey] = v
				}
			}
			if len(params) > 0 {
				out.AlarmParams = params
			}
		} else {
			// 无 hw 映射（如 SleepadSetting）——hw 完全不返回此类型的 params。
			// 置 nil 避免 defaults 里的默认值（如 timezone="", report_upload_time=8）
			// 在 MergeHardwareIntoBaseline 里被当成 HW 值覆盖 baseline 的 per-device 覆盖。
			out.AlarmParams = nil
		}

		result = append(result, out)
	}
	return result
}

// MergeAlarmLevelFromDB takes hardware-sourced items (IsEnabled + AlarmParams)
// and fills in AlarmLevel from DB items. Hardware IsEnabled/AlarmParams take priority.
func MergeAlarmLevelFromDB(hwItems, dbItems []alarm.AlarmItem) []alarm.AlarmItem {
	dbMap := make(map[string]alarm.AlarmItem, len(dbItems))
	for _, item := range dbItems {
		dbMap[item.AlarmType] = item
	}

	result := make([]alarm.AlarmItem, 0, len(hwItems))
	for _, hw := range hwItems {
		if db, ok := dbMap[hw.AlarmType]; ok && db.AlarmLevel != nil {
			hw.AlarmLevel = db.AlarmLevel
		}
		result = append(result, hw)
	}
	return result
}

// MergeHardwareIntoBaseline overlays hardware IsEnabled and AlarmParams onto baseline (DB or defaults).
// AlarmLevel and other fields stay from baseline. Use for display: show DB values first, overlay Sleepace switch/threshold only.
func MergeHardwareIntoBaseline(baseline, hwItems []alarm.AlarmItem) []alarm.AlarmItem {
	hwMap := make(map[string]alarm.AlarmItem, len(hwItems))
	for _, item := range hwItems {
		if item.AlarmType != "" {
			hwMap[item.AlarmType] = item
		}
	}
	result := make([]alarm.AlarmItem, 0, len(baseline))
	for _, b := range baseline {
		out := b
		if hw, ok := hwMap[b.AlarmType]; ok {
			if hw.IsEnabled != nil {
				out.IsEnabled = hw.IsEnabled
			}
			// Per-key merge: 硬件字段覆盖 baseline 同名字段，但保留 baseline 里硬件不返回的字段
			// （如 SleepadSetting.timezone / report_upload_time —— 厂家 getalarmnotifyconfig 不返回这些，
			// 若整体替换会导致 DB 里的 per-device 覆盖值在 GET 响应里丢失，UI 显示不出来）
			if len(hw.AlarmParams) > 0 {
				merged := make(map[string]interface{}, len(out.AlarmParams)+len(hw.AlarmParams))
				for k, v := range out.AlarmParams {
					merged[k] = v
				}
				for k, v := range hw.AlarmParams {
					merged[k] = v
				}
				out.AlarmParams = merged
			}
		}
		result = append(result, out)
	}
	return result
}

func isFlag(snake string) bool {
	l := len(snake)
	return l > 12 && snake[l-12:] == "_alarm_level"
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

// UpdateSleepaceSettingsToHardware pushes alarm settings via wisefido-sleepace transparent proxy.
func UpdateSleepaceSettingsToHardware(ctx context.Context, gateway *SleepaceGatewayClient, device *domain.Device, deviceID string, settings map[string]interface{}) error {
	if gateway == nil {
		return fmt.Errorf("sleepace gateway is nil")
	}
	if !device.DeviceCode.Valid || device.DeviceCode.String == "" {
		return fmt.Errorf("device has no device_code")
	}
	deviceCode := device.DeviceCode.String
	hardwareSettings := ConvertFlatSettingsToSleepaceFormat(deviceCode, device.DeviceUID, settings)
	return gateway.UpdateAlarmConfig(ctx, hardwareSettings)
}

// ConvertFlatSettingsToSleepaceFormat converts flat snake_case settings to sleepad hardware API format.
// userId = device_factory_meta.device_uid (logMAC)；deviceId = device_factory_meta.device_code。
func ConvertFlatSettingsToSleepaceFormat(deviceCode, deviceUID string, settings map[string]interface{}) map[string]interface{} {
	hardwareSettings := make(map[string]interface{})

	hardwareSettings["userId"] = deviceUID
	hardwareSettings["deviceId"] = deviceCode

	// 离床时间配置
	if leftBedStartHour, ok := settings["left_bed_start_hour"]; ok {
		if hour, ok := leftBedStartHour.(float64); ok {
			hardwareSettings["leftBedStartHour"] = int(hour)
		} else if hour, ok := leftBedStartHour.(int); ok {
			hardwareSettings["leftBedStartHour"] = hour
		}
	}
	if leftBedStartMinute, ok := settings["left_bed_start_minute"]; ok {
		if minute, ok := leftBedStartMinute.(float64); ok {
			hardwareSettings["leftBedStartMinute"] = int(minute)
		} else if minute, ok := leftBedStartMinute.(int); ok {
			hardwareSettings["leftBedStartMinute"] = minute
		}
	}
	if leftBedEndHour, ok := settings["left_bed_end_hour"]; ok {
		if hour, ok := leftBedEndHour.(float64); ok {
			hardwareSettings["leftBedEndHour"] = int(hour)
		} else if hour, ok := leftBedEndHour.(int); ok {
			hardwareSettings["leftBedEndHour"] = hour
		}
	}
	if leftBedEndMinute, ok := settings["left_bed_end_minute"]; ok {
		if minute, ok := leftBedEndMinute.(float64); ok {
			hardwareSettings["leftBedEndMinute"] = int(minute)
		} else if minute, ok := leftBedEndMinute.(int); ok {
			hardwareSettings["leftBedEndMinute"] = minute
		}
	}
	if leftBedDuration, ok := settings["left_bed_duration"]; ok {
		if duration, ok := leftBedDuration.(float64); ok {
			hardwareSettings["leftBedDuration"] = int(duration)
		} else if duration, ok := leftBedDuration.(int); ok {
			hardwareSettings["leftBedDuration"] = duration
		}
	}
	// leftBedFlag: 1 = enabled, 0 = disabled
	if leftBedLevel, ok := settings["left_bed_alarm_level"].(string); ok {
		if leftBedLevel != "disabled" && leftBedLevel != "" {
			hardwareSettings["leftBedFlag"] = 1
		} else {
			hardwareSettings["leftBedFlag"] = 0
		}
	}

	// 心率配置
	if minHeartRate, ok := settings["min_heart_rate"]; ok {
		if rate, ok := minHeartRate.(float64); ok {
			hardwareSettings["minHeartRate"] = int(rate)
		} else if rate, ok := minHeartRate.(int); ok {
			hardwareSettings["minHeartRate"] = rate
		}
	}
	if maxHeartRate, ok := settings["max_heart_rate"]; ok {
		if rate, ok := maxHeartRate.(float64); ok {
			hardwareSettings["maxHeartRate"] = int(rate)
		} else if rate, ok := maxHeartRate.(int); ok {
			hardwareSettings["maxHeartRate"] = rate
		}
	}
	if heartRateSlowDuration, ok := settings["heart_rate_slow_duration"]; ok {
		if duration, ok := heartRateSlowDuration.(float64); ok {
			hardwareSettings["heartRateSlowDuration"] = int(duration)
		} else if duration, ok := heartRateSlowDuration.(int); ok {
			hardwareSettings["heartRateSlowDuration"] = duration
		}
	}
	if heartRateFastDuration, ok := settings["heart_rate_fast_duration"]; ok {
		if duration, ok := heartRateFastDuration.(float64); ok {
			hardwareSettings["heartRateFastDuration"] = int(duration)
		} else if duration, ok := heartRateFastDuration.(int); ok {
			hardwareSettings["heartRateFastDuration"] = duration
		}
	}
	if heartRateSlowLevel, ok := settings["heart_rate_slow_alarm_level"].(string); ok {
		if heartRateSlowLevel != "disabled" && heartRateSlowLevel != "" {
			hardwareSettings["heartRateSlowFlag"] = 1
		} else {
			hardwareSettings["heartRateSlowFlag"] = 0
		}
	}
	if heartRateFastLevel, ok := settings["heart_rate_fast_alarm_level"].(string); ok {
		if heartRateFastLevel != "disabled" && heartRateFastLevel != "" {
			hardwareSettings["heartRateFastFlag"] = 1
		} else {
			hardwareSettings["heartRateFastFlag"] = 0
		}
	}

	// 呼吸率配置：范围 9–30 来自 Sleepace API 报错 "minBreathRate should be between 9 and 30 (status: 9)"，对接文档未写具体区间
	const breathRateLo, breathRateHi = 9, 30
	if minBreathRate, ok := settings["min_breath_rate"]; ok {
		var v int
		hasVal := false
		if rate, ok := minBreathRate.(float64); ok {
			v, hasVal = int(rate), true
		} else if rate, ok := minBreathRate.(int); ok {
			v, hasVal = rate, true
		}
		if hasVal {
			if v < breathRateLo {
				v = breathRateLo
			} else if v > breathRateHi {
				v = breathRateHi
			}
			hardwareSettings["minBreathRate"] = v
		}
	}
	if maxBreathRate, ok := settings["max_breath_rate"]; ok {
		var v int
		hasVal := false
		if rate, ok := maxBreathRate.(float64); ok {
			v, hasVal = int(rate), true
		} else if rate, ok := maxBreathRate.(int); ok {
			v, hasVal = rate, true
		}
		if hasVal {
			if v < breathRateLo {
				v = breathRateLo
			} else if v > breathRateHi {
				v = breathRateHi
			}
			hardwareSettings["maxBreathRate"] = v
		}
	}
	if breathRateSlowDuration, ok := settings["breath_rate_slow_duration"]; ok {
		if duration, ok := breathRateSlowDuration.(float64); ok {
			hardwareSettings["breathRateSlowDuration"] = int(duration)
		} else if duration, ok := breathRateSlowDuration.(int); ok {
			hardwareSettings["breathRateSlowDuration"] = duration
		}
	}
	if breathRateFastDuration, ok := settings["breath_rate_fast_duration"]; ok {
		if duration, ok := breathRateFastDuration.(float64); ok {
			hardwareSettings["breathRateFastDuration"] = int(duration)
		} else if duration, ok := breathRateFastDuration.(int); ok {
			hardwareSettings["breathRateFastDuration"] = duration
		}
	}
	if breathRateSlowLevel, ok := settings["breath_rate_slow_alarm_level"].(string); ok {
		if breathRateSlowLevel != "disabled" && breathRateSlowLevel != "" {
			hardwareSettings["breathRateSlowFlag"] = 1
		} else {
			hardwareSettings["breathRateSlowFlag"] = 0
		}
	}
	if breathRateFastLevel, ok := settings["breath_rate_fast_alarm_level"].(string); ok {
		if breathRateFastLevel != "disabled" && breathRateFastLevel != "" {
			hardwareSettings["breathRateFastFlag"] = 1
		} else {
			hardwareSettings["breathRateFastFlag"] = 0
		}
	}

	// 呼吸暂停配置
	if breathPauseDuration, ok := settings["breath_pause_duration"]; ok {
		if duration, ok := breathPauseDuration.(float64); ok {
			hardwareSettings["breathPauseDuration"] = int(duration)
		} else if duration, ok := breathPauseDuration.(int); ok {
			hardwareSettings["breathPauseDuration"] = duration
		}
	}
	if breathPauseLevel, ok := settings["breath_pause_alarm_level"].(string); ok {
		if breathPauseLevel != "disabled" && breathPauseLevel != "" {
			hardwareSettings["breathPauseFlag"] = 1
		} else {
			hardwareSettings["breathPauseFlag"] = 0
		}
	}

	// 身体移动配置
	if bodyMoveDuration, ok := settings["body_move_duration"]; ok {
		if duration, ok := bodyMoveDuration.(float64); ok {
			hardwareSettings["bodyMoveDuration"] = int(duration)
		} else if duration, ok := bodyMoveDuration.(int); ok {
			hardwareSettings["bodyMoveDuration"] = duration
		}
	}
	if bodyMoveLevel, ok := settings["body_move_alarm_level"].(string); ok {
		if bodyMoveLevel != "disabled" && bodyMoveLevel != "" {
			hardwareSettings["bodyMoveFlag"] = 1
		} else {
			hardwareSettings["bodyMoveFlag"] = 0
		}
	}

	// 无身体移动配置
	if nobodyMoveDuration, ok := settings["nobody_move_duration"]; ok {
		if duration, ok := nobodyMoveDuration.(float64); ok {
			hardwareSettings["nobodyMoveDuration"] = int(duration)
		} else if duration, ok := nobodyMoveDuration.(int); ok {
			hardwareSettings["nobodyMoveDuration"] = duration
		}
	}
	if nobodyMoveLevel, ok := settings["nobody_move_alarm_level"].(string); ok {
		if nobodyMoveLevel != "disabled" && nobodyMoveLevel != "" {
			hardwareSettings["nobodyMoveFlag"] = 1
		} else {
			hardwareSettings["nobodyMoveFlag"] = 0
		}
	}

	// 无翻身配置
	if noTurnOverDuration, ok := settings["no_turn_over_duration"]; ok {
		if duration, ok := noTurnOverDuration.(float64); ok {
			hardwareSettings["noTurnOverDuration"] = int(duration)
		} else if duration, ok := noTurnOverDuration.(int); ok {
			hardwareSettings["noTurnOverDuration"] = duration
		}
	}
	if noTurnOverLevel, ok := settings["no_turn_over_alarm_level"].(string); ok {
		if noTurnOverLevel != "disabled" && noTurnOverLevel != "" {
			hardwareSettings["noTurnOverFlag"] = 1
		} else {
			hardwareSettings["noTurnOverFlag"] = 0
		}
	}

	// 坐起配置
	if situpLevel, ok := settings["situp_alarm_level"].(string); ok {
		if situpLevel != "disabled" && situpLevel != "" {
			hardwareSettings["situpFlag"] = 1
		} else {
			hardwareSettings["situpFlag"] = 0
		}
	}

	// 在床配置
	if onbedDuration, ok := settings["onbed_duration"]; ok {
		if duration, ok := onbedDuration.(float64); ok {
			hardwareSettings["onbedDuration"] = int(duration)
		} else if duration, ok := onbedDuration.(int); ok {
			hardwareSettings["onbedDuration"] = duration
		}
	}
	if onbedLevel, ok := settings["onbed_alarm_level"].(string); ok {
		if onbedLevel != "disabled" && onbedLevel != "" {
			hardwareSettings["onbedFlag"] = 1
		} else {
			hardwareSettings["onbedFlag"] = 0
		}
	}

	// 传感器跌落配置：前端发送 fallFlag (boolean)
	if fallFlag, ok := settings["fallFlag"].(bool); ok {
		if fallFlag {
			hardwareSettings["fallFlag"] = 1
		} else {
			hardwareSettings["fallFlag"] = 0
		}
	}

	return hardwareSettings
}

// alarmTypeToSleepaceLevelKeys AlarmType → 对应的 Sleepace flat level key(s)
// 仅包含 Sleepace 设备支持的报警项；NightAbsence/SensorDetached/ResetTime/NapTime 等不在此处，不写入设备
// 单个 AlarmType 可能对应多个 level key（如 HeartRateAlert → slow + fast）
var alarmTypeToSleepaceLevelKeys = map[string][]string{
	alarm.LeftBed:              {"left_bed_alarm_level"},
	alarm.HeartRateAlert:       {"heart_rate_slow_alarm_level", "heart_rate_fast_alarm_level"},
	alarm.RespRateAlert:        {"breath_rate_slow_alarm_level", "breath_rate_fast_alarm_level"},
	alarm.ApneaHypopnea:        {"breath_pause_alarm_level"},
	alarm.AbnormalBodyMovement: {"body_move_alarm_level"},
	alarm.NoBodyMove:           {"nobody_move_alarm_level"},
	alarm.NoTurnOver:           {"no_turn_over_alarm_level"},
	alarm.BedSitUp:             {"situp_alarm_level"},
	alarm.InBed:                {"onbed_alarm_level"},
}

// alarmTypeParamMapping AlarmType → (generic param key → sleepace flat key)
var alarmTypeParamMapping = map[string]map[string]string{
	alarm.HeartRateAlert: {
		"min":               "min_heart_rate",
		"max":               "max_heart_rate",
		"slow_duration_sec": "heart_rate_slow_duration",
		"fast_duration_sec": "heart_rate_fast_duration",
	},
	alarm.RespRateAlert: {
		"min":               "min_breath_rate",
		"max":               "max_breath_rate",
		"slow_duration_sec": "breath_rate_slow_duration",
		"fast_duration_sec": "breath_rate_fast_duration",
	},
	alarm.LeftBed: {
		"duration_sec": "left_bed_duration",
	},
	alarm.ApneaHypopnea: {
		"duration_sec": "breath_pause_duration",
	},
	alarm.AbnormalBodyMovement: {
		"duration_min": "body_move_duration",
	},
	alarm.NoBodyMove: {
		"duration_min": "nobody_move_duration",
	},
	alarm.NoTurnOver: {
		"duration_min": "no_turn_over_duration",
	},
	alarm.InBed: {
		"duration_min": "onbed_duration",
	},
}

// ConvertAlarmItemsToSleepaceConfig converts AlarmItem[] to sleepad cloud API format.
// userId = device_factory_meta.device_uid (logMAC)，deviceId = device_factory_meta.device_code。
// resetTime 为租户作息（alarm_cloud.metadata），非 nil 且存在 LeftBed 时写入 left_bed_start/end 下发设备。
func ConvertAlarmItemsToSleepaceConfig(deviceCode, deviceUID string, alarmItems []alarm.AlarmItem, resetTime *alarm.ResetTimeParams) map[string]interface{} {
	flat := make(map[string]interface{})

	for _, item := range alarmItems {
		levelVal := "disabled"
		if item.IsEnabled != nil && *item.IsEnabled == alarm.IsEnabledOn && item.AlarmLevel != nil && *item.AlarmLevel != "" {
			levelVal = *item.AlarmLevel
		}

		if levelKeys, ok := alarmTypeToSleepaceLevelKeys[item.AlarmType]; ok {
			for _, k := range levelKeys {
				flat[k] = levelVal
			}
		}

		if paramMap, ok := alarmTypeParamMapping[item.AlarmType]; ok {
			for genericKey, flatKey := range paramMap {
				if v, exists := item.AlarmParams[genericKey]; exists {
					flat[flatKey] = v
				}
			}
		}
	}

	if resetTime != nil && resetTime.InBedTime != "" && resetTime.OutBedTime != "" {
		if inBed, err := time.Parse("15:04", resetTime.InBedTime); err == nil {
			flat["left_bed_start_hour"] = inBed.Hour()
			flat["left_bed_start_minute"] = inBed.Minute()
		}
		if outBed, err := time.Parse("15:04", resetTime.OutBedTime); err == nil {
			flat["left_bed_end_hour"] = outBed.Hour()
			flat["left_bed_end_minute"] = outBed.Minute()
		}
	}

	return ConvertFlatSettingsToSleepaceFormat(deviceCode, deviceUID, flat)
}
