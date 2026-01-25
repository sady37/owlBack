package service

import (
	"context"
	"encoding/json"
	"fmt"

	"owl-common/alarm"

	"wisefido-data/internal/domain"
)

// GetSleepaceSettingsFromHardware 从硬件读取 Sleepace 设备配置（参考 v1.0 实现）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::GetDeviceMonitorSettings
func GetSleepaceSettingsFromHardware(ctx context.Context, client *SleepaceClient, device *domain.Device, deviceID string) ([]alarm.AlarmItem, error) {
	if client == nil {
		return nil, fmt.Errorf("sleepace client is nil")
	}

	// 调用 Sleepace API 从硬件读取配置
	request := SleepaceRequest{
		Token: client.token,
		Data: map[string]any{
			"userId": deviceID,
		},
	}

	var response SleepaceResponse
	resp, err := client.httpClient.R().
		SetBody(request).
		SetResult(&response).
		Post("/sleepace/getalarmnotifyconfig")

	if err != nil {
		return nil, fmt.Errorf("failed to call Sleepace API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Sleepace API returned status code: %d", resp.StatusCode())
	}

	if response.Status != 0 {
		return nil, fmt.Errorf("Sleepace API error: %s (status: %d)", response.Msg, response.Status)
	}

	// 解析硬件返回的配置
	// 参考 v1.0: models.SleepaceMonitorSettings 结构
	var hardwareSettings map[string]interface{}
	if err := json.Unmarshal(response.Data, &hardwareSettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hardware settings: %w", err)
	}

	// 将硬件返回的配置转换为 flat 结构（与前端期望的格式一致）
	settings := make(map[string]interface{})

	// 离床时间配置
	if leftBedStartHour, ok := hardwareSettings["leftBedStartHour"].(float64); ok {
		settings["left_bed_start_hour"] = int(leftBedStartHour)
	}
	if leftBedStartMinute, ok := hardwareSettings["leftBedStartMinute"].(float64); ok {
		settings["left_bed_start_minute"] = int(leftBedStartMinute)
	}
	if leftBedEndHour, ok := hardwareSettings["leftBedEndHour"].(float64); ok {
		settings["left_bed_end_hour"] = int(leftBedEndHour)
	}
	if leftBedEndMinute, ok := hardwareSettings["leftBedEndMinute"].(float64); ok {
		settings["left_bed_end_minute"] = int(leftBedEndMinute)
	}
	if leftBedDuration, ok := hardwareSettings["leftBedDuration"].(float64); ok {
		settings["left_bed_duration"] = int(leftBedDuration)
	}
	if leftBedFlag, ok := hardwareSettings["leftBedFlag"].(float64); ok {
		if leftBedFlag == 1 {
			settings["left_bed_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["left_bed_alarm_level"] = "disabled"
		}
	}

	// 心率配置
	if minHeartRate, ok := hardwareSettings["minHeartRate"].(float64); ok {
		settings["min_heart_rate"] = int(minHeartRate)
	}
	if maxHeartRate, ok := hardwareSettings["maxHeartRate"].(float64); ok {
		settings["max_heart_rate"] = int(maxHeartRate)
	}
	if heartRateSlowDuration, ok := hardwareSettings["heartRateSlowDuration"].(float64); ok {
		settings["heart_rate_slow_duration"] = int(heartRateSlowDuration)
	}
	if heartRateFastDuration, ok := hardwareSettings["heartRateFastDuration"].(float64); ok {
		settings["heart_rate_fast_duration"] = int(heartRateFastDuration)
	}
	if heartRateSlowFlag, ok := hardwareSettings["heartRateSlowFlag"].(float64); ok {
		if heartRateSlowFlag == 1 {
			settings["heart_rate_slow_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["heart_rate_slow_alarm_level"] = "disabled"
		}
	}
	if heartRateFastFlag, ok := hardwareSettings["heartRateFastFlag"].(float64); ok {
		if heartRateFastFlag == 1 {
			settings["heart_rate_fast_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["heart_rate_fast_alarm_level"] = "disabled"
		}
	}

	// 呼吸率配置
	if minBreathRate, ok := hardwareSettings["minBreathRate"].(float64); ok {
		settings["min_breath_rate"] = int(minBreathRate)
	}
	if maxBreathRate, ok := hardwareSettings["maxBreathRate"].(float64); ok {
		settings["max_breath_rate"] = int(maxBreathRate)
	}
	if breathRateSlowDuration, ok := hardwareSettings["breathRateSlowDuration"].(float64); ok {
		settings["breath_rate_slow_duration"] = int(breathRateSlowDuration)
	}
	if breathRateFastDuration, ok := hardwareSettings["breathRateFastDuration"].(float64); ok {
		settings["breath_rate_fast_duration"] = int(breathRateFastDuration)
	}
	if breathRateSlowFlag, ok := hardwareSettings["breathRateSlowFlag"].(float64); ok {
		if breathRateSlowFlag == 1 {
			settings["breath_rate_slow_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_rate_slow_alarm_level"] = "disabled"
		}
	}
	if breathRateFastFlag, ok := hardwareSettings["breathRateFastFlag"].(float64); ok {
		if breathRateFastFlag == 1 {
			settings["breath_rate_fast_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_rate_fast_alarm_level"] = "disabled"
		}
	}

	// 呼吸暂停配置
	if breathPauseDuration, ok := hardwareSettings["breathPauseDuration"].(float64); ok {
		settings["breath_pause_duration"] = int(breathPauseDuration)
	}
	if breathPauseFlag, ok := hardwareSettings["breathPauseFlag"].(float64); ok {
		if breathPauseFlag == 1 {
			settings["breath_pause_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_pause_alarm_level"] = "disabled"
		}
	}

	// 身体移动配置
	if bodyMoveDuration, ok := hardwareSettings["bodyMoveDuration"].(float64); ok {
		settings["body_move_duration"] = int(bodyMoveDuration)
	}
	if bodyMoveFlag, ok := hardwareSettings["bodyMoveFlag"].(float64); ok {
		if bodyMoveFlag == 1 {
			settings["body_move_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["body_move_alarm_level"] = "disabled"
		}
	}

	// 无身体移动配置
	if nobodyMoveDuration, ok := hardwareSettings["nobodyMoveDuration"].(float64); ok {
		settings["nobody_move_duration"] = int(nobodyMoveDuration)
	}
	if nobodyMoveFlag, ok := hardwareSettings["nobodyMoveFlag"].(float64); ok {
		if nobodyMoveFlag == 1 {
			settings["nobody_move_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["nobody_move_alarm_level"] = "disabled"
		}
	}

	// 无翻身配置
	if noTurnOverDuration, ok := hardwareSettings["noTurnOverDuration"].(float64); ok {
		settings["no_turn_over_duration"] = int(noTurnOverDuration)
	}
	if noTurnOverFlag, ok := hardwareSettings["noTurnOverFlag"].(float64); ok {
		if noTurnOverFlag == 1 {
			settings["no_turn_over_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["no_turn_over_alarm_level"] = "disabled"
		}
	}

	// 坐起配置
	if situpFlag, ok := hardwareSettings["situpFlag"].(float64); ok {
		if situpFlag == 1 {
			settings["situp_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["situp_alarm_level"] = "disabled"
		}
	}

	// 在床配置
	if onbedDuration, ok := hardwareSettings["onbedDuration"].(float64); ok {
		settings["onbed_duration"] = int(onbedDuration)
	}
	if onbedFlag, ok := hardwareSettings["onbedFlag"].(float64); ok {
		if onbedFlag == 1 {
			settings["onbed_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["onbed_alarm_level"] = "disabled"
		}
	}

	// 传感器跌落配置：返回 fallFlag (boolean)
	if fallFlag, ok := hardwareSettings["fallFlag"].(float64); ok {
		settings["fallFlag"] = fallFlag == 1
	} else {
		settings["fallFlag"] = false
	}

	// 转换为 AlarmItem 数组（简化处理，返回空数组）
	// 注意：硬件读取功能暂时不使用，如果需要可以后续实现转换逻辑
	alarmItems := make([]alarm.AlarmItem, 0)

	return alarmItems, nil
}

// UpdateSleepaceSettingsToHardware 将配置同步到 Sleepace 硬件（参考 v1.0 实现）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::SetDeviceMonitorSettings
func UpdateSleepaceSettingsToHardware(ctx context.Context, client *SleepaceClient, device *domain.Device, deviceID string, settings map[string]interface{}) error {
	if client == nil {
		return fmt.Errorf("sleepace client is nil")
	}

	// 获取设备代码（deviceCode），使用 device_uid
	if device.DeviceUID == "" {
		return fmt.Errorf("device has no device_uid")
	}
	deviceCode := device.DeviceUID

	// 将 flat settings 转换为 SleepaceMonitorSettings 格式（用于硬件 API）
	hardwareSettings := ConvertFlatSettingsToSleepaceFormat(deviceID, deviceCode, settings)

	// 调用 Sleepace API 同步到硬件
	request := struct {
		Token *SleepaceToken         `json:"token"`
		Data  map[string]interface{} `json:"data"`
	}{
		Token: client.token,
		Data:  hardwareSettings,
	}

	var response SleepaceResponse
	resp, err := client.httpClient.R().
		SetBody(request).
		SetResult(&response).
		Post("/sleepace/updatealarmnotifyconfig")

	if err != nil {
		return fmt.Errorf("failed to call Sleepace API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Sleepace API returned status code: %d", resp.StatusCode())
	}

	if response.Status != 0 {
		return fmt.Errorf("Sleepace API error: %s (status: %d)", response.Msg, response.Status)
	}

	return nil
}

// ConvertFlatSettingsToSleepaceFormat 将 flat settings 转换为 SleepaceMonitorSettings 格式（用于硬件 API）
// 参考：wisefido-backend/wisefido-sleepace/models/settings.go::SleepaceMonitorSettings
func ConvertFlatSettingsToSleepaceFormat(deviceID, deviceCode string, settings map[string]interface{}) map[string]interface{} {
	hardwareSettings := make(map[string]interface{})

	// 基本字段
	hardwareSettings["userId"] = deviceID
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

	// 呼吸率配置
	if minBreathRate, ok := settings["min_breath_rate"]; ok {
		if rate, ok := minBreathRate.(float64); ok {
			hardwareSettings["minBreathRate"] = int(rate)
		} else if rate, ok := minBreathRate.(int); ok {
			hardwareSettings["minBreathRate"] = rate
		}
	}
	if maxBreathRate, ok := settings["max_breath_rate"]; ok {
		if rate, ok := maxBreathRate.(float64); ok {
			hardwareSettings["maxBreathRate"] = int(rate)
		} else if rate, ok := maxBreathRate.(int); ok {
			hardwareSettings["maxBreathRate"] = rate
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
