package repository

import (
	"encoding/json"
	"fmt"
)

// extractTrackingFields 提取轨迹字段
// 返回：tracking_id (88 → NULL), radar_pos_x, radar_pos_y, radar_pos_z
func extractTrackingFields(data map[string]interface{}) (trackingID *int, posX, posY, posZ *int) {
	// target_id → tracking_id (88 → NULL)
	if targetID, ok := data["target_id"]; ok {
		if tid, err := parseInt(targetID); err == nil {
			// 88 表示无人，映射为 NULL
			if tid != 88 {
				trackingID = &tid
			}
		}
	}

	// position_x → radar_pos_x (已转换：厘米)
	if posXVal, ok := data["position_x"]; ok {
		if x, err := parseInt(posXVal); err == nil {
			posX = &x
		}
	}

	// position_y → radar_pos_y (已转换：厘米)
	if posYVal, ok := data["position_y"]; ok {
		if y, err := parseInt(posYVal); err == nil {
			posY = &y
		}
	}

	// position_z → radar_pos_z (已转换：厘米)
	if posZVal, ok := data["position_z"]; ok {
		if z, err := parseInt(posZVal); err == nil {
			posZ = &z
		}
	}

	return trackingID, posX, posY, posZ
}

// extractPostureFields 提取姿态字段
// 返回：posture_snomed_code, posture_display
func extractPostureFields(data map[string]interface{}) (snomedCode, display *string) {
	if code, ok := data["pose_snomed_code"].(string); ok && code != "" {
		snomedCode = &code
	}
	if disp, ok := data["pose_display_en"].(string); ok && disp != "" {
		display = &disp
	}
	return snomedCode, display
}

// extractEventFields 提取事件字段
// 返回：event_type, event_snomed_code, event_display, area_id
func extractEventFields(data map[string]interface{}) (eventType, snomedCode, display *string, areaID *int) {
	if et, ok := data["event_type"].(string); ok && et != "" {
		eventType = &et
	}
	if code, ok := data["event_snomed_code"].(string); ok && code != "" {
		snomedCode = &code
	}
	if disp, ok := data["event_display_en"].(string); ok && disp != "" {
		display = &disp
	}
	if areaIDVal, ok := data["area_id"]; ok {
		if aid, err := parseInt(areaIDVal); err == nil {
			areaID = &aid
		}
	}
	return eventType, snomedCode, display, areaID
}

// extractVitalSignsFields 提取生命体征字段
// 返回：heart_rate_code, heart_rate_display, heart_rate, respiratory_rate_code, respiratory_rate_display, respiratory_rate
func extractVitalSignsFields(data map[string]interface{}) (hrCode, hrDisplay *string, hr *int, rrCode, rrDisplay *string, rr *int) {
	// 心率固定值
	hrCodeStr := "364075005"
	hrDisplayStr := "Heart rate"
	hrCode = &hrCodeStr
	hrDisplay = &hrDisplayStr

	// heart_rate 字段（已转换：bpm）
	if hrVal, ok := data["heart_rate"]; ok {
		if h, err := parseInt(hrVal); err == nil {
			hr = &h
		}
	}

	// 呼吸频率固定值
	rrCodeStr := "86290005"
	rrDisplayStr := "Respiratory rate"
	rrCode = &rrCodeStr
	rrDisplay = &rrDisplayStr

	// respiratory_rate 或 breath_rate 字段（已转换：次/分钟）
	if rrVal, ok := data["respiratory_rate"]; ok {
		if r, err := parseInt(rrVal); err == nil {
			rr = &r
		}
	} else if breathRate, ok := data["breath_rate"]; ok {
		// Sleepace 使用 breath_rate
		if r, err := parseInt(breathRate); err == nil {
			rr = &r
		}
	}

	return hrCode, hrDisplay, hr, rrCode, rrDisplay, rr
}

// extractSleepStateFields 提取睡眠状态字段
// 返回：sleep_state_snomed_code, sleep_state_display
func extractSleepStateFields(data map[string]interface{}) (snomedCode, display *string) {
	// 优先使用 sleep_state_snomed_code 和 sleep_state_display_en（Radar）
	if code, ok := data["sleep_state_snomed_code"].(string); ok && code != "" {
		snomedCode = &code
	} else if code, ok := data["sleep_status_snomed_code"].(string); ok && code != "" {
		// monitor.bh 使用 sleep_status_snomed_code
		snomedCode = &code
	}

	if disp, ok := data["sleep_state_display_en"].(string); ok && disp != "" {
		display = &disp
	} else if disp, ok := data["sleep_status_display_en"].(string); ok && disp != "" {
		// monitor.bh 使用 sleep_status_display_en
		display = &disp
	} else if disp, ok := data["sleepStage_display_en"].(string); ok && disp != "" {
		// Sleepace 使用 sleepStage_display_en
		display = &disp
	} else if code, ok := data["sleepStage_snomed_code"].(string); ok && code != "" {
		// Sleepace 使用 sleepStage_snomed_code
		snomedCode = &code
	}

	return snomedCode, display
}

// buildMetadata 构建 metadata JSONB
// 保存统计数据和扩展信息
func buildMetadata(data map[string]interface{}) ([]byte, error) {
	metadata := make(map[string]interface{})

	// 统计数据（stat 类型）
	statistics := make(map[string]interface{})

	// 轨迹统计（track）
	if walkDistance, ok := data["walk_distance"]; ok {
		statistics["walk_distance"] = walkDistance
	}
	if walkDuration, ok := data["walk_duration"]; ok {
		statistics["walk_duration"] = walkDuration
	}
	if lieDuration, ok := data["lie_duration"]; ok {
		statistics["lie_duration"] = lieDuration
	}
	if standDuration, ok := data["stand_duration"]; ok {
		statistics["stand_duration"] = standDuration
	}
	if multiPersonDuration, ok := data["multi_person_duration"]; ok {
		statistics["multi_person_duration"] = multiPersonDuration
	}
	if peopleCount, ok := data["people_count"]; ok {
		statistics["people_count"] = peopleCount
	}
	if version, ok := data["version"]; ok {
		statistics["version"] = version
	}

	// 睡眠统计（sleep）
	if realtimeBreathRate, ok := data["realtime_breath_rate"]; ok {
		statistics["realtime_breath_rate"] = realtimeBreathRate
	}
	if realtimeHeartRate, ok := data["realtime_heart_rate"]; ok {
		statistics["realtime_heart_rate"] = realtimeHeartRate
	}
	if avgBreathRate, ok := data["avg_breath_rate"]; ok {
		statistics["avg_breath_rate"] = avgBreathRate
	}
	if avgHeartRate, ok := data["avg_heart_rate"]; ok {
		statistics["avg_heart_rate"] = avgHeartRate
	}
	if breathState, ok := data["breath_state"]; ok {
		statistics["breath_state"] = breathState
	}
	if heartState, ok := data["heart_state"]; ok {
		statistics["heart_state"] = heartState
	}
	if vitalSignsState, ok := data["vital_signs_state"]; ok {
		statistics["vital_signs_state"] = vitalSignsState
	}
	if sleepState, ok := data["sleep_state"]; ok {
		statistics["sleep_state"] = sleepState
	}

	// 如果有统计数据，添加到 metadata
	if len(statistics) > 0 {
		metadata["statistics"] = statistics
	}

	// 调试信息
	debug := make(map[string]interface{})
	if signalQuality, ok := data["signal_quality"]; ok {
		debug["signal_quality"] = signalQuality
	}
	if stability, ok := data["stability"]; ok {
		debug["stability"] = stability
	}
	if len(debug) > 0 {
		metadata["debug"] = debug
	}

	// 原始值（用于日志追溯）
	original := make(map[string]interface{})
	if pose, ok := data["pose"]; ok {
		original["pose"] = pose
	}
	if event, ok := data["event"]; ok {
		original["event"] = event
	}
	if sleepStatus, ok := data["sleep_status"]; ok {
		original["sleep_status"] = sleepStatus
	}
	if len(original) > 0 {
		metadata["original"] = original
	}

	// 其他扩展字段
	if sitUp, ok := data["sitUp"]; ok {
		metadata["sit_up"] = sitUp
	}
	if bedStatus, ok := data["bedStatus"]; ok {
		metadata["bed_status"] = bedStatus
	}
	if turnOver, ok := data["turnOver"]; ok {
		metadata["turn_over"] = turnOver
	}
	if bodyMove, ok := data["bodyMove"]; ok {
		metadata["body_move"] = bodyMove
	}
	if initStatus, ok := data["initStatus"]; ok {
		metadata["init_status"] = initStatus
	}
	if leftRight, ok := data["leftRight"]; ok {
		metadata["left_right"] = leftRight
	}

	// 如果 metadata 为空，返回空 JSON 对象
	if len(metadata) == 0 {
		return json.Marshal(map[string]interface{}{})
	}

	return json.Marshal(metadata)
}

// buildMinimalAuditData 构建最小化审计数据（用于 raw_original）
// 根据 HIPAA/FDA 要求，只保存必要的审计追溯信息
func buildMinimalAuditData(data map[string]interface{}) ([]byte, error) {
	auditData := make(map[string]interface{})

	// 设备类型
	if deviceType, ok := data["device_type"].(string); ok {
		auditData["device_type"] = deviceType
	}

	// 主题类型（Radar 使用 topic_type，Sleepace 使用 data_key）
	if topicType, ok := data["topic_type"].(string); ok {
		auditData["topic_type"] = topicType
	} else if dataKey, ok := data["data_key"].(string); ok {
		// Sleepace 使用 data_key
		auditData["data_key"] = dataKey
	}

	// 数据类型（如果存在）
	if dataType, ok := data["data_type"].(string); ok && dataType != "" {
		auditData["data_type"] = dataType
	} else {
		// 根据 topic_type 或 data_key 推断
		if topicType, ok := data["topic_type"].(string); ok {
			if topicType == "alarm" {
				auditData["data_type"] = "alarm"
			} else {
				auditData["data_type"] = "observation"
			}
		} else if dataKey, ok := data["data_key"].(string); ok {
			// Sleepace 数据根据 data_key 判断
			if dataKey == "alarmNotify" {
				auditData["data_type"] = "alarm"
			} else {
				auditData["data_type"] = "observation"
			}
		} else {
			auditData["data_type"] = "observation"
		}
	}

	// Category（如果存在）
	if category, ok := data["category"].(string); ok && category != "" {
		auditData["category"] = category
	}

	// 时间戳
	if timestamp, ok := data["timestamp"]; ok {
		auditData["timestamp"] = timestamp
	}

	return json.Marshal(auditData)
}

// parseInt 将 interface{} 转换为 int
func parseInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}
