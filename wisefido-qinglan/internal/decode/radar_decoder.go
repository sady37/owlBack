package decode

import (
	"encoding/base64"
	"fmt"
)

// RadarDecoder 解码 Radar 数据（基于 broker.go 优化）
// data: 包含原始数据字段（从 MQTT 报文解析）
// topicType: 主题类型 ("monitor", "stat", "event", "alarm")
// 返回: data_value（对象或数组），只包含解码后的数据值，不包含元数据
func RadarDecoder(data map[string]interface{}, topicType string) (interface{}, error) {
	switch topicType {
	case "monitor":
		return decodeRadarMonitor(data)
	case "stat":
		return decodeRadarStat(data)
	case "event":
		return decodeRadarEvent(data)
	case "alarm":
		return decodeRadarAlarm(data)
	default:
		return []map[string]interface{}{}, nil
	}
}

// MonitorTrackData track 解码数据
type MonitorTrackData struct {
	TargetID      int
	PositionX     int // 厘米
	PositionY     int // 厘米
	PositionZ     int
	RemainingTime int
	Pose          int
	Event         int
	AreaID        int
}

// VitalData vital 解码数据
type VitalData struct {
	VitalFlag       int
	RespiratoryRate int
	HeartRate       int
	SleepStatus     int // bit 7:6
	Stability       int // bit 1:0
}

// SleepData sleep 解码数据
type SleepData struct {
	SleepFlag          int
	RespiratoryRate    int
	HeartRate          int
	AvgRespiratoryRate int
	AvgHeartRate       int
	BreathStatus       int // bit 1:0
	HeartStatus        int // bit 3:2
	VitalSignsStatus   int // bit 5:4
	SleepStatus        int // bit 7:6
}

// decodeRadarMonitor 解码实时数据
func decodeRadarMonitor(data map[string]interface{}) (interface{}, error) {
	var dataValue []map[string]interface{}

	// 1. 处理 track 数据
	var trackBase64 string
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if track, ok := dataField["track"].(string); ok && track != "" {
			trackBase64 = track
		}
	} else if track, ok := data["track"].(string); ok && track != "" {
		trackBase64 = track
	}

	if trackBase64 != "" {
		trackDataList, err := decodeMonitorTrackFromBase64(trackBase64)
		if err == nil {
			for i, trackData := range trackDataList {
				trackObj := map[string]interface{}{
					"category":       "track",
					"target_id":      trackData.TargetID,
					"position_x":     trackData.PositionX,
					"position_y":     trackData.PositionY,
					"position_z":     trackData.PositionZ,
					"remaining_time": trackData.RemainingTime,
					"area_id":        trackData.AreaID,
				}
				// 应用 SNOMED 映射（标准化，避免被单个厂家绑死）
				applyRadarSNOMedMapping(trackObj, "pose", "monitor.track.pose", trackData.Pose)
				applyRadarSNOMedMapping(trackObj, "event", "monitor.track.event", trackData.Event)
				if i == len(trackDataList)-1 {
					trackObj["raw_original"] = trackBase64
				}
				dataValue = append(dataValue, trackObj)
			}
		}
	}

	// 2. 处理 vital 数据
	var bhBase64 string
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if bh, ok := dataField["bh"].(string); ok && bh != "" {
			bhBase64 = bh
		}
	} else if bh, ok := data["bh"].(string); ok && bh != "" {
		bhBase64 = bh
	}

	if bhBase64 != "" {
		vitalData, err := decodeMonitorVitalFromBase64(bhBase64)
		if err == nil {
			vitalObj := map[string]interface{}{
				"category":         "vital",
				"vital_flag":       vitalData.VitalFlag,
				"respiratory_rate": vitalData.RespiratoryRate,
				"heart_rate":       vitalData.HeartRate,
				"raw_original":     bhBase64,
			}
			// 应用 SNOMED 映射（标准化，避免被单个厂家绑死）
			applyRadarSNOMedMapping(vitalObj, "sleep_status", "monitor.vital.sleep_status", vitalData.SleepStatus)
			applyRadarSNOMedMapping(vitalObj, "stability", "monitor.vital.stability", vitalData.Stability)
			dataValue = append(dataValue, vitalObj)
		}
	}

	if len(dataValue) == 0 {
		return map[string]interface{}{}, nil
	} else if len(dataValue) == 1 {
		return dataValue[0], nil
	}
	return dataValue, nil
}

// decodeRadarStat 解码统计数据
func decodeRadarStat(data map[string]interface{}) (interface{}, error) {
	var dataValue []map[string]interface{}

	// 1. 处理 track 统计数据
	if version, ok := data["version"]; ok {
		trackObj := map[string]interface{}{
			"category": "track",
			"version":  version,
		}
		if peopleCount, ok := data["people_count"]; ok {
			trackObj["people_count"] = peopleCount
		}
		if walkDistance, ok := data["walk_distance"]; ok {
			trackObj["walk_distance"] = walkDistance
		}
		if walkDuration, ok := data["walk_duration"]; ok {
			trackObj["walk_duration"] = walkDuration
		}
		if lieDuration, ok := data["lie_duration"]; ok {
			trackObj["lie_duration"] = lieDuration
		}
		if standDuration, ok := data["stand_duration"]; ok {
			trackObj["stand_duration"] = standDuration
		}
		if multiPersonDuration, ok := data["multi_person_duration"]; ok {
			trackObj["multi_person_duration"] = multiPersonDuration
		}
		if len(trackObj) > 1 {
			dataValue = append(dataValue, trackObj)
		}
	}

	// 2. 处理 sleep 统计数据
	var sleepBase64 string
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if sleep, ok := dataField["sleep"].(string); ok && sleep != "" {
			sleepBase64 = sleep
		}
	} else if sleep, ok := data["sleep"].(string); ok && sleep != "" {
		sleepBase64 = sleep
	}

	if sleepBase64 != "" {
		sleepData, err := decodeStatSleepFromBase64(sleepBase64)
		if err == nil {
			sleepObj := map[string]interface{}{
				"category":             "sleep",
				"sleep_flag":           sleepData.SleepFlag,
				"respiratory_rate":     sleepData.RespiratoryRate,
				"heart_rate":           sleepData.HeartRate,
				"avg_respiratory_rate": sleepData.AvgRespiratoryRate,
				"avg_heart_rate":       sleepData.AvgHeartRate,
				"raw_original":         sleepBase64,
			}
			// 应用 SNOMED 映射（标准化，避免被单个厂家绑死）
			applyRadarSNOMedMapping(sleepObj, "breath_state", "statistics.sleep.breath_state", sleepData.BreathStatus)
			applyRadarSNOMedMapping(sleepObj, "heart_state", "statistics.sleep.heart_state", sleepData.HeartStatus)
			applyRadarSNOMedMapping(sleepObj, "vital_signs_state", "statistics.sleep.vital_signs_state", sleepData.VitalSignsStatus)
			applyRadarSNOMedMapping(sleepObj, "sleep_state", "statistics.sleep.sleep_state", sleepData.SleepStatus)

			// 分别提取4组状态（bit 1:0, bit 3:2, bit 5:4, bit 7:6）
			// 用于生成数字组合，便于在 device_monitor 中查找配置
			breathState := sleepData.BreathStatus & 0x03         // bit 1:0: 呼吸状态 (00=正常, 01=过低, 10=过高, 11=暂停)
			heartState := sleepData.HeartStatus & 0x03           // bit 3:2: 心率状态 (00=正常, 01=过低, 10=过高, 11=未定义)
			vitalSignsState := sleepData.VitalSignsStatus & 0x03 // bit 5:4: 生命体征 (00=正常, 01=未定义, 02=未定义, 11=弱)
			// sleepState := sleepData.SleepStatus & 0x03     // bit 7:6: 睡眠状态 (00=未定义, 01=浅睡, 10=深睡, 11=清醒) - 通常不是报警

			// 生成数字组合字段（用于查找 device_monitor 配置）
			// 格式：sleep_{type}_{state}，例如 "sleep_breath_01", "sleep_heart_10", "sleep_vitals_11"
			var numericCodes []string
			if breathState != 0 { // 00=正常，非0表示异常
				sleepObj["stat_numeric_code_breath"] = fmt.Sprintf("sleep_breath_%02d", breathState)
				numericCodes = append(numericCodes, fmt.Sprintf("sleep_breath_%02d", breathState))
			}
			if heartState != 0 { // 00=正常，非0表示异常
				sleepObj["stat_numeric_code_heart"] = fmt.Sprintf("sleep_heart_%02d", heartState)
				numericCodes = append(numericCodes, fmt.Sprintf("sleep_heart_%02d", heartState))
			}
			if vitalSignsState == 3 { // 11=生命体征弱
				sleepObj["stat_numeric_code_vitals"] = "sleep_vitals_11"
				numericCodes = append(numericCodes, "sleep_vitals_11")
			}
			// 保留所有数字组合（可能同时有多个异常）
			if len(numericCodes) > 0 {
				sleepObj["stat_numeric_codes"] = numericCodes
			}

			dataValue = append(dataValue, sleepObj)
		}
	}

	if len(dataValue) == 0 {
		return map[string]interface{}{}, nil
	} else if len(dataValue) == 1 {
		return dataValue[0], nil
	}
	return dataValue, nil
}

// decodeRadarEvent 解码事件数据
// 返回: data_value（单个对象），一条消息只包含一个事件，不是数组
// 支持系统1格式：cmd="event", type=1/2/3/5/7/8/9 (int), data="" (事件内容)
// 注意：一条 MQTT 消息只包含一个事件，没有多数组的情况
func decodeRadarEvent(data map[string]interface{}) (interface{}, error) {
	// 系统1格式：事件类型：1-进出事件，2-姿态变化事件，3-人数变化事件，5-设备在线状态，7-信号差，8-倾角异常，9-其他告警
	// eventType 从顶层 data["type"] 获取
	eventType, _ := data["type"].(int)

	// 处理 data 字段（一条消息只有一个事件）
	var eventObj map[string]interface{}
	if dataField, ok := data["data"]; ok {
		// data 可能是单个对象
		if dataMap, ok := dataField.(map[string]interface{}); ok {
			eventObj = buildEventObjectFromType(eventType, dataMap)
		} else if dataArray, ok := dataField.([]interface{}); ok && len(dataArray) > 0 {
			// 兼容处理：如果 data 是数组，只取第一个（实际不应该出现）
			if eventMap, ok := dataArray[0].(map[string]interface{}); ok {
				eventObj = buildEventObjectFromType(eventType, eventMap)
			}
		}
	} else {
		// 如果没有 data 字段，直接使用顶层数据
		eventObj = buildEventObjectFromType(eventType, data)
	}

	// 返回单个对象（不是数组）
	if len(eventObj) == 0 {
		return map[string]interface{}{}, nil
	}
	return eventObj, nil
}

// decodeRadarAlarm 解码告警数据（与 event 相同）
func decodeRadarAlarm(data map[string]interface{}) (interface{}, error) {
	return decodeRadarEvent(data)
}

// buildEventObjectFromType 根据 eventType 构建事件对象
// eventType 从顶层 data["type"] 获取，eventMap 是 data["data"] 中的单个事件对象
func buildEventObjectFromType(eventType int, eventMap map[string]interface{}) map[string]interface{} {
	eventObj := make(map[string]interface{})

	// 根据事件类型处理
	if eventType == 1 {
		// type=1: 进出事件 → category="enter2out"
		eventObj["category"] = "enter2out"
		// 保存原始的 eventType 用于 ConvertMQTTToAlarmTypeRadar
		eventObj["event_type"] = "1"
		if trackID, ok := eventMap["track-id"]; ok {
			eventObj["track_id"] = trackID
		} else if trackID, ok := eventMap["track_id"]; ok {
			eventObj["track_id"] = trackID
		}
		if event, ok := eventMap["event"]; ok {
			// 保存原始的 event 值（用于生成数字组合，查找 device_monitor 配置）
			if eventInt, ok := event.(int); ok {
				eventObj["event_raw"] = fmt.Sprintf("%d", eventInt)
			} else if eventStr, ok := event.(string); ok {
				eventObj["event_raw"] = eventStr
			}
			// 应用 SNOMED 映射（标准化，避免被单个厂家绑死）
			applyRadarSNOMedMapping(eventObj, "event", "event.enter2out.event", event)
		}
		if areaType, ok := eventMap["area_type"]; ok {
			// 保存原始 area_type 值（用于 ConvertMQTTToAlarmTypeRadar）
			if areaTypeInt, ok := areaType.(int); ok {
				eventObj["area_type_raw"] = fmt.Sprintf("%d", areaTypeInt)
			} else if areaTypeStr, ok := areaType.(string); ok {
				eventObj["area_type_raw"] = areaTypeStr
			}
			// area_type 转换为 display_en 值（用于显示）
			eventObj["area_type"] = mapAreaType(areaType)
		}
		// 保留其他字段（排除已处理的字段）
		excludedKeys := map[string]bool{
			"track-id":  true,
			"track_id":  true,
			"event":     true,
			"area_type": true,
		}
		for k, v := range eventMap {
			if _, exists := eventObj[k]; !exists && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 2 {
		// type=2: 姿态变化事件 → category="pose"
		eventObj["category"] = "pose"
		// 保存原始的 eventType 用于 ConvertMQTTToAlarmTypeRadar
		eventObj["event_type"] = "2"
		if trackID, ok := eventMap["track-id"]; ok {
			eventObj["track_id"] = trackID
		} else if trackID, ok := eventMap["track_id"]; ok {
			eventObj["track_id"] = trackID
		}
		if pose, ok := eventMap["pose"]; ok {
			// 保存原始的 pose 值（字符串格式，用于 ConvertMQTTToAlarmTypeRadar）
			if poseInt, ok := pose.(int); ok {
				eventObj["pose_raw"] = fmt.Sprintf("%d", poseInt)
			} else if poseStr, ok := pose.(string); ok {
				eventObj["pose_raw"] = poseStr
			}
			// 应用 SNOMED 映射（标准化，避免被单个厂家绑死）
			// 注意：fieldPath 使用 event.pose_change.pose（与配置表一致）
			applyRadarSNOMedMapping(eventObj, "pose", "event.pose_change.pose", pose)
		}
		// 保留其他字段
		excludedKeys := map[string]bool{
			"track-id": true,
			"track_id": true,
			"pose":     true,
		}
		for k, v := range eventMap {
			if _, exists := eventObj[k]; !exists && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 3 {
		// type=3: 人数变化事件
		eventObj["category"] = "number-people"
		if numberPeople, ok := eventMap["number-people"]; ok {
			eventObj["number_people"] = numberPeople
		} else if numberPeople, ok := eventMap["number_people"]; ok {
			eventObj["number_people"] = numberPeople
		}
		// 保留其他字段
		excludedKeys := map[string]bool{
			"number-people": true,
			"number_people": true,
		}
		for k, v := range eventMap {
			if _, exists := eventObj[k]; !exists && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 5 {
		// type=5: 设备在线状态 → category="isOnline"
		eventObj["category"] = "isOnline"
		// 转换 isOnline 字段：0="online", 1="offline"
		if isOnline, ok := eventMap["isOnline"]; ok {
			if isOnlineInt, ok := isOnline.(int); ok {
				if isOnlineInt == 0 {
					eventObj["device_status"] = "online"
				} else {
					eventObj["device_status"] = "offline"
				}
			} else if isOnlineStr, ok := isOnline.(string); ok {
				if isOnlineStr == "0" {
					eventObj["device_status"] = "online"
				} else {
					eventObj["device_status"] = "offline"
				}
			} else {
				eventObj["isOnline"] = isOnline
			}
		}
		// 复制其他字段
		excludedKeys := map[string]bool{
			"isOnline": true,
		}
		for k, v := range eventMap {
			if k != "category" && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 7 {
		// type=7: 信号差事件 → category="signal_poor"
		eventObj["category"] = "signal_poor"
		// 转换 recovery 字段：0="signal_poor", 1="signal_recovery"
		if recovery, ok := eventMap["recovery"]; ok {
			if recoveryInt, ok := recovery.(int); ok {
				if recoveryInt == 0 {
					eventObj["recovery"] = "signal_poor"
				} else {
					eventObj["recovery"] = "signal_recovery"
				}
			} else if recoveryStr, ok := recovery.(string); ok {
				if recoveryStr == "0" {
					eventObj["recovery"] = "signal_poor"
				} else {
					eventObj["recovery"] = "signal_recovery"
				}
			} else {
				eventObj["recovery"] = recovery
			}
		}
		// 复制其他字段
		excludedKeys := map[string]bool{
			"recovery": true,
		}
		for k, v := range eventMap {
			if k != "category" && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 8 {
		// type=8: 倾角异常事件 → category="angle_abnormal"
		eventObj["category"] = "angle_abnormal"
		// 保存原始的 eventType 用于 ConvertMQTTToAlarmTypeRadar
		eventObj["event_type"] = "8"
		// 转换 recovery 字段：0="angle_abnormal", 1="angle_recovery"
		if recovery, ok := eventMap["recovery"]; ok {
			if recoveryInt, ok := recovery.(int); ok {
				if recoveryInt == 0 {
					eventObj["recovery"] = "angle_abnormal"
				} else {
					eventObj["recovery"] = "angle_recovery"
				}
			} else if recoveryStr, ok := recovery.(string); ok {
				if recoveryStr == "0" {
					eventObj["recovery"] = "angle_abnormal"
				} else {
					eventObj["recovery"] = "angle_recovery"
				}
			} else {
				eventObj["recovery"] = recovery
			}
		}
		// 复制其他字段
		excludedKeys := map[string]bool{
			"recovery": true,
		}
		for k, v := range eventMap {
			if k != "category" && !excludedKeys[k] {
				eventObj[k] = v
			}
		}
	} else if eventType == 9 {
		// type=9: 其他告警 → category="other"
		eventObj["category"] = "other"
		// 直接复制原始字段，不做值转换
		for k, v := range eventMap {
			if k != "category" {
				eventObj[k] = v
			}
		}
	} else {
		// 未知类型，保留原始数据
		for k, v := range eventMap {
			if k != "type" && k != "event_type" {
				eventObj[k] = v
			}
		}
		if len(eventObj) > 0 {
			eventObj["category"] = "unknown"
		}
	}

	if len(eventObj) > 1 { // 至少有 category 和其他字段
		return eventObj
	}
	return make(map[string]interface{})
}

// mapAreaType 映射 area_type
func mapAreaType(areaType interface{}) interface{} {
	if areaTypeInt, ok := areaType.(int); ok {
		switch areaTypeInt {
		case 2:
			return "Bed"
		case 5:
			return "Monitoring bed"
		case 6:
			return "Sensor area"
		default:
			return areaType
		}
	}
	return areaType
}

// decodeMonitorTrackFromBase64 解码 monitor track（基于 broker.go）
func decodeMonitorTrackFromBase64(base64Track string) ([]MonitorTrackData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 track: %w", err)
	}

	if len(data)%16 != 0 {
		return nil, fmt.Errorf("invalid track length: %d (must be multiple of 16)", len(data))
	}

	personCount := len(data) / 16
	results := make([]MonitorTrackData, personCount)

	for i := 0; i < personCount; i++ {
		offset := i * 16
		segment := data[offset : offset+16]

		// 坐标转换：分米 → 厘米（参考 wisefido-radar）
		positionX := int(int8(segment[1])) * 10
		positionY := int(int8(segment[2])) * 10

		// 事件过滤（参考现有实现）
		event := int(segment[14])
		if event > 4 {
			event = 0
		}

		results[i] = MonitorTrackData{
			TargetID:      int(segment[0]),
			PositionX:     positionX,
			PositionY:     positionY,
			PositionZ:     int(segment[3]),
			RemainingTime: int(segment[12]),
			Pose:          int(segment[13]),
			Event:         event,
			AreaID:        int(segment[15]),
		}
	}

	return results, nil
}

// decodeMonitorVitalFromBase64 解码 monitor vital（基于 broker.go）
func decodeMonitorVitalFromBase64(base64Bh string) (*VitalData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Bh)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 vital: %w", err)
	}

	if len(data) < 14 {
		return nil, fmt.Errorf("invalid vital length: %d (must be at least 14)", len(data))
	}

	byte13 := int(data[13])
	sleepStatus := (byte13 >> 6) & 0x03

	var stability int
	if len(data) > 14 {
		stability = int(data[14]) & 0x03
	}

	return &VitalData{
		VitalFlag:       0,
		RespiratoryRate: int(data[1]),
		HeartRate:       int(data[2]),
		SleepStatus:     sleepStatus,
		Stability:       stability,
	}, nil
}

// decodeStatSleepFromBase64 解码 stat sleep（基于 broker.go）
func decodeStatSleepFromBase64(base64Sleep string) (*SleepData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Sleep)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 sleep: %w", err)
	}

	if len(data) < 14 {
		return nil, fmt.Errorf("invalid sleep length: %d (must be at least 14)", len(data))
	}

	byte13 := int(data[13])
	breathStatus := byte13 & 0x03
	heartStatus := (byte13 >> 2) & 0x03
	vitalSignsStatus := (byte13 >> 4) & 0x03
	sleepStatus := (byte13 >> 6) & 0x03

	return &SleepData{
		SleepFlag:          255,
		RespiratoryRate:    int(data[1]),
		HeartRate:          int(data[2]),
		AvgRespiratoryRate: int(data[5]),
		AvgHeartRate:       int(data[6]),
		BreathStatus:       breathStatus,
		HeartStatus:        heartStatus,
		VitalSignsStatus:   vitalSignsStatus,
		SleepStatus:        sleepStatus,
	}, nil
}
