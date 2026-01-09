package encode

import (
	"encoding/base64"
)

// RadarEncode 编码 Radar 数据
// data: 包含 device_id, tenant_id, device_type, topic_type, timestamp 和原始数据字段
// topicType: 主题类型 ("monitor", "stat", "event", "alarm")
// 返回: 编码后的数据（单位转换、格式统一、SNOMED 映射）
func RadarEncode(data map[string]interface{}, topicType string) (map[string]interface{}, error) {
	encoded := make(map[string]interface{})

	// 1. 保留元数据
	encoded["device_id"] = data["device_id"]
	encoded["tenant_id"] = data["tenant_id"]
	encoded["device_type"] = data["device_type"]
	encoded["topic_type"] = topicType
	encoded["timestamp"] = data["timestamp"]

	// 2. 根据 topic_type 进行不同的编码转换
	switch topicType {
	case "monitor":
		return encodeRadarMonitor(data, encoded)
	case "stat":
		return encodeRadarStat(data, encoded)
	case "event":
		return encodeRadarEvent(data, encoded)
	case "alarm":
		return encodeRadarAlarm(data, encoded)
	default:
		// 默认：直接复制其他字段
		copyOtherFields(data, encoded, []string{"device_id", "tenant_id", "device_type", "topic_type", "timestamp"})
		return encoded, nil
	}
}

// encodeRadarMonitor 编码实时数据
func encodeRadarMonitor(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// duration 字段（track 字节 12）：剩余时间，单位为秒，取值 0-60，无需转换
	if duration, ok := data["duration"]; ok {
		encoded["duration"] = duration
	}

	// 位置坐标单位转换：使用转换表
	if posX, ok := data["position_x"]; ok {
		if x, err := parseInt(posX); err == nil {
			if conv, err := GetFieldConversion("monitor.track.position_x"); err == nil && conv.UnitConversion != nil {
				encoded["position_x"] = x * 10 // dm -> cm
			} else {
				encoded["position_x"] = x
			}
		} else {
			encoded["position_x"] = posX
		}
	}
	if posY, ok := data["position_y"]; ok {
		if y, err := parseInt(posY); err == nil {
			if conv, err := GetFieldConversion("monitor.track.position_y"); err == nil && conv.UnitConversion != nil {
				encoded["position_y"] = y * 10 // dm -> cm
			} else {
				encoded["position_y"] = y
			}
		} else {
			encoded["position_y"] = posY
		}
	}
	if posZ, ok := data["position_z"]; ok {
		// z 坐标已经是厘米，无需转换
		encoded["position_z"] = posZ
	}

	// 姿态值转换：使用转换表获取 SNOMED 映射
	if pose, ok := data["pose"]; ok {
		applySNOMedMapping(encoded, "pose", "monitor.track.pose", pose)
	}

	// 事件值转换：使用转换表获取 SNOMED 映射
	if event, ok := data["event"]; ok {
		applySNOMedMapping(encoded, "event", "monitor.track.event", event)
	}

	// monitor.bh 字段：处理 base64 编码的呼吸心率数据
	// 如果 bh 是 base64 字符串，需要解码并提取位字段
	if bh, ok := data["bh"].(string); ok && bh != "" {
		// 解码 base64
		bhBytes, err := base64.StdEncoding.DecodeString(bh)
		if err == nil && len(bhBytes) >= 14 {
			// 提取 sleep_status (字节 13, bit 7:6)
			if len(bhBytes) > 13 {
				byte13 := int(bhBytes[13])
				if sleepStatusBits, err := ParseBitField(byte13, "7:6"); err == nil {
					applySNOMedMapping(encoded, "sleep_status", "monitor.bh.sleep_status", sleepStatusBits)
				}
				// 提取 stability (字节 14, bit 1:0)
				if len(bhBytes) > 14 {
					byte14 := int(bhBytes[14])
					if stabilityBits, err := ParseBitField(byte14, "1:0"); err == nil {
						applySNOMedMapping(encoded, "stability", "monitor.bh.stability", stabilityBits)
					}
				}
			}
			// 提取 breath_rate (字节 1) 和 heart_rate (字节 2)
			if len(bhBytes) > 2 {
				encoded["breath_rate"] = int(bhBytes[1])
				encoded["heart_rate"] = int(bhBytes[2])
			}
		}
	}

	// 如果 sleep_status 和 stability 已经作为独立字段存在（已解码），直接进行 SNOMED 映射
	if sleepStatus, ok := data["sleep_status"]; ok {
		if _, exists := encoded["sleep_status"]; !exists {
			applySNOMedMapping(encoded, "sleep_status", "monitor.bh.sleep_status", sleepStatus)
		}
	}
	if stability, ok := data["stability"]; ok {
		if _, exists := encoded["stability"]; !exists {
			applySNOMedMapping(encoded, "stability", "monitor.bh.stability", stability)
		}
	}

	// 复制其他字段（包括 bh 的原始 base64 字符串，如果存在）
	excludeFields := []string{"device_id", "tenant_id", "device_type", "topic_type", "timestamp", "duration", "position_x", "position_y", "position_z", "pose", "event", "sleep_status", "stability", "breath_rate", "heart_rate"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeRadarStat 编码统计数据
func encodeRadarStat(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// track 字段中的时长字段：单位为秒，无需转换
	if walkDuration, ok := data["walk_duration"]; ok {
		encoded["walk_duration"] = walkDuration
	}
	if lieDuration, ok := data["lie_duration"]; ok {
		encoded["lie_duration"] = lieDuration
	}
	if standDuration, ok := data["stand_duration"]; ok {
		encoded["stand_duration"] = standDuration
	}
	if multiPersonDuration, ok := data["multi_person_duration"]; ok {
		encoded["multi_person_duration"] = multiPersonDuration
	}

	// walk_distance 单位转换：米 → 厘米
	if walkDistance, ok := data["walk_distance"]; ok {
		if dist, err := parseInt(walkDistance); err == nil {
			if conv, err := GetFieldConversion("stat.track.walk_distance"); err == nil && conv.UnitConversion != nil {
				encoded["walk_distance"] = dist * 100 // m -> cm
			} else {
				encoded["walk_distance"] = dist
			}
		} else {
			encoded["walk_distance"] = walkDistance
		}
	}

	// stat.sleep 字段：处理 base64 编码的睡眠统计数据
	// 如果 sleep 是 base64 字符串，需要解码并提取位字段
	if sleep, ok := data["sleep"].(string); ok && sleep != "" {
		// 解码 base64
		sleepBytes, err := base64.StdEncoding.DecodeString(sleep)
		if err == nil && len(sleepBytes) >= 14 {
			// 提取 hr_breath_event (字节 13) 的各个位字段
			if len(sleepBytes) > 13 {
				byte13 := int(sleepBytes[13])

				// breath_state (bit 1:0) - 呼吸状态
				if breathStateBits, err := ParseBitField(byte13, "1:0"); err == nil {
					applySNOMedMapping(encoded, "breath_state", "stat.sleep.hr_breath_event.breath_state", breathStateBits)
				}
				// heart_state (bit 3:2) - 心率状态
				if heartStateBits, err := ParseBitField(byte13, "3:2"); err == nil {
					applySNOMedMapping(encoded, "heart_state", "stat.sleep.hr_breath_event.heart_state", heartStateBits)
				}
				// vital_signs_state (bit 5:4) - 生命体征情况
				if vitalSignsStateBits, err := ParseBitField(byte13, "5:4"); err == nil {
					applySNOMedMapping(encoded, "vital_signs_state", "stat.sleep.hr_breath_event.vital_signs_state", vitalSignsStateBits)
				}
				// sleep_state (bit 7:6) - 睡眠状态
				if sleepStateBits, err := ParseBitField(byte13, "7:6"); err == nil {
					applySNOMedMapping(encoded, "sleep_state", "stat.sleep.hr_breath_event.sleep_state", sleepStateBits)
				}
			}
			// 提取其他字段
			if len(sleepBytes) > 6 {
				encoded["realtime_breath_rate"] = int(sleepBytes[1])
				encoded["realtime_heart_rate"] = int(sleepBytes[2])
				encoded["avg_breath_rate"] = int(sleepBytes[5])
				encoded["avg_heart_rate"] = int(sleepBytes[6])
			}
		}
	}

	// 如果位字段已经作为独立字段存在（已解码），直接进行 SNOMED 映射
	if breathState, ok := data["breath_state"]; ok {
		if _, exists := encoded["breath_state"]; !exists {
			applySNOMedMapping(encoded, "breath_state", "stat.sleep.hr_breath_event.breath_state", breathState)
		}
	}
	if heartState, ok := data["heart_state"]; ok {
		if _, exists := encoded["heart_state"]; !exists {
			applySNOMedMapping(encoded, "heart_state", "stat.sleep.hr_breath_event.heart_state", heartState)
		}
	}
	if vitalSignsState, ok := data["vital_signs_state"]; ok {
		if _, exists := encoded["vital_signs_state"]; !exists {
			applySNOMedMapping(encoded, "vital_signs_state", "stat.sleep.hr_breath_event.vital_signs_state", vitalSignsState)
		}
	}
	if sleepState, ok := data["sleep_state"]; ok {
		if _, exists := encoded["sleep_state"]; !exists {
			applySNOMedMapping(encoded, "sleep_state", "stat.sleep.hr_breath_event.sleep_state", sleepState)
		}
	}

	// 复制其他字段（包括 sleep 和 track 的原始 base64 字符串，如果存在）
	excludeFields := []string{"device_id", "tenant_id", "device_type", "topic_type", "timestamp", "walk_duration", "lie_duration", "stand_duration", "multi_person_duration", "walk_distance", "breath_state", "heart_state", "vital_signs_state", "sleep_state", "realtime_breath_rate", "realtime_heart_rate", "avg_breath_rate", "avg_heart_rate"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeRadarEvent 编码事件数据
func encodeRadarEvent(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// 事件类型：1-进出事件，2-姿态变化事件，3-人数变化事件
	eventType, _ := data["type"].(int)

	// 处理 data 字段（可能是数组）
	if dataField, ok := data["data"]; ok {
		// data 可能是数组，包含多个事件
		if dataArray, ok := dataField.([]interface{}); ok {
			// 处理数组中的每个事件
			var processedEvents []map[string]interface{}
			for _, eventItem := range dataArray {
				if eventMap, ok := eventItem.(map[string]interface{}); ok {
					processedEvent := make(map[string]interface{})

					// 根据事件类型处理
					if eventType == 1 {
						// 进出事件：event 字段需要 SNOMED 映射
						if event, ok := eventMap["event"]; ok {
							applySNOMedMapping(processedEvent, "event", "event.enter_leave.event", event)
							processedEvent["event"] = event
						}
						if areaType, ok := eventMap["area_type"]; ok {
							processedEvent["area_type"] = areaType
						}
						if trackID, ok := eventMap["track-id"]; ok {
							processedEvent["track-id"] = trackID
						}
					} else if eventType == 2 {
						// 姿态变化事件：pose 字段需要 SNOMED 映射
						if pose, ok := eventMap["pose"]; ok {
							applySNOMedMapping(processedEvent, "pose", "event.pose_change.pose", pose)
							processedEvent["pose"] = pose
						}
						if trackID, ok := eventMap["track-id"]; ok {
							processedEvent["track-id"] = trackID
						}
					}

					// 保留其他字段
					for k, v := range eventMap {
						if _, exists := processedEvent[k]; !exists {
							processedEvent[k] = v
						}
					}

					processedEvents = append(processedEvents, processedEvent)
				} else {
					// 如果不是 map，直接保留
					processedEvents = append(processedEvents, map[string]interface{}{"data": eventItem})
				}
			}
			if len(processedEvents) > 0 {
				encoded["data"] = processedEvents
			}
		} else if dataMap, ok := dataField.(map[string]interface{}); ok {
			// data 可能是单个对象（人数变化事件 type=3）
			if eventType == 3 {
				if numberPeople, ok := dataMap["number-people"]; ok {
					encoded["number-people"] = numberPeople
				}
				if numberPeople, ok := dataMap["number_people"]; ok {
					encoded["number_people"] = numberPeople
				}
			}
			encoded["data"] = dataMap
		} else {
			// 其他类型，直接保留
			encoded["data"] = dataField
		}
	}

	// 兼容处理：如果 event 和 pose 字段直接在顶层
	if eventType == 1 {
		// 进出事件：event 字段需要 SNOMED 映射
		if event, ok := data["event"]; ok {
			if _, exists := encoded["event"]; !exists {
				applySNOMedMapping(encoded, "event", "event.enter_leave.event", event)
			}
		}
		// area_type 不需要 SNOMED 映射（标识符）
		if areaType, ok := data["area_type"]; ok {
			if _, exists := encoded["area_type"]; !exists {
				encoded["area_type"] = areaType
			}
		}
	} else if eventType == 2 {
		// 姿态变化事件：pose 字段需要 SNOMED 映射
		if pose, ok := data["pose"]; ok {
			if _, exists := encoded["pose"]; !exists {
				applySNOMedMapping(encoded, "pose", "event.pose_change.pose", pose)
			}
		}
	} else if eventType == 3 {
		// 人数变化事件：number-people 不需要 SNOMED 映射（数值）
		if numberPeople, ok := data["number-people"]; ok {
			if _, exists := encoded["number-people"]; !exists {
				encoded["number-people"] = numberPeople
			}
		}
		if numberPeople, ok := data["number_people"]; ok {
			if _, exists := encoded["number_people"]; !exists {
				encoded["number_people"] = numberPeople
			}
		}
	}

	// 保留 type 字段
	if eventTypeVal, ok := data["type"]; ok {
		encoded["type"] = eventTypeVal
	}

	// 保留 track-id 字段（标识符，不需要 SNOMED 映射）
	if trackID, ok := data["track-id"]; ok {
		if _, exists := encoded["track-id"]; !exists {
			encoded["track-id"] = trackID
		}
	}
	if trackID, ok := data["track_id"]; ok {
		if _, exists := encoded["track_id"]; !exists {
			encoded["track_id"] = trackID
		}
	}

	// 保留 cmd 字段
	if cmd, ok := data["cmd"]; ok {
		encoded["cmd"] = cmd
	}

	// 复制其他字段
	excludeFields := []string{"device_id", "tenant_id", "device_type", "topic_type", "timestamp", "type", "event", "pose", "area_type", "track-id", "track_id", "number-people", "number_people", "data", "cmd"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeRadarAlarm 编码告警数据
func encodeRadarAlarm(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// 告警数据格式与事件数据类似
	// 告警类型：1-进出事件，2-姿态变化事件，3-人数变化事件（通常告警是姿态相关）
	alarmType, _ := data["type"].(int)

	// 处理 data 字段（可能是数组或对象）
	if dataField, ok := data["data"]; ok {
		if dataArray, ok := dataField.([]interface{}); ok {
			// data 是数组，包含多个告警事件
			var processedAlarms []map[string]interface{}
			for _, alarmItem := range dataArray {
				if alarmMap, ok := alarmItem.(map[string]interface{}); ok {
					processedAlarm := make(map[string]interface{})

					// 根据告警类型处理
					if alarmType == 1 {
						// 进出事件告警：event 字段需要 SNOMED 映射
						if event, ok := alarmMap["event"]; ok {
							applySNOMedMapping(processedAlarm, "event", "event.enter_leave.event", event)
							processedAlarm["event"] = event
						}
					} else if alarmType == 2 {
						// 姿态变化告警：pose 字段需要 SNOMED 映射
						if pose, ok := alarmMap["pose"]; ok {
							applySNOMedMapping(processedAlarm, "pose", "event.pose_change.pose", pose)
							processedAlarm["pose"] = pose
						}
					}

					// 保留其他字段
					for k, v := range alarmMap {
						if _, exists := processedAlarm[k]; !exists {
							processedAlarm[k] = v
						}
					}

					processedAlarms = append(processedAlarms, processedAlarm)
				} else {
					processedAlarms = append(processedAlarms, map[string]interface{}{"data": alarmItem})
				}
			}
			if len(processedAlarms) > 0 {
				encoded["data"] = processedAlarms
			}
		} else if dataMap, ok := dataField.(map[string]interface{}); ok {
			// data 是单个对象
			if alarmType == 3 {
				if numberPeople, ok := dataMap["number-people"]; ok {
					encoded["number-people"] = numberPeople
				}
				if numberPeople, ok := dataMap["number_people"]; ok {
					encoded["number_people"] = numberPeople
				}
			}
			encoded["data"] = dataMap
		} else {
			encoded["data"] = dataField
		}
	}

	// 兼容处理：如果 event 和 pose 字段直接在顶层
	if alarmType == 1 {
		// 进出事件告警：event 字段需要 SNOMED 映射
		if event, ok := data["event"]; ok {
			if _, exists := encoded["event"]; !exists {
				applySNOMedMapping(encoded, "event", "event.enter_leave.event", event)
			}
		}
	} else if alarmType == 2 {
		// 姿态变化告警：pose 字段需要 SNOMED 映射
		if pose, ok := data["pose"]; ok {
			if _, exists := encoded["pose"]; !exists {
				applySNOMedMapping(encoded, "pose", "event.pose_change.pose", pose)
			}
		}
	} else if alarmType == 3 {
		// 人数变化告警：number-people 不需要 SNOMED 映射
		if numberPeople, ok := data["number-people"]; ok {
			if _, exists := encoded["number-people"]; !exists {
				encoded["number-people"] = numberPeople
			}
		}
		if numberPeople, ok := data["number_people"]; ok {
			if _, exists := encoded["number_people"]; !exists {
				encoded["number_people"] = numberPeople
			}
		}
	}

	// 保留 type 字段
	if alarmTypeVal, ok := data["type"]; ok {
		encoded["type"] = alarmTypeVal
	}

	// 保留 track-id 字段（标识符，不需要 SNOMED 映射）
	if trackID, ok := data["track-id"]; ok {
		if _, exists := encoded["track-id"]; !exists {
			encoded["track-id"] = trackID
		}
	}
	if trackID, ok := data["track_id"]; ok {
		if _, exists := encoded["track_id"]; !exists {
			encoded["track_id"] = trackID
		}
	}

	// 保留 cmd 字段
	if cmd, ok := data["cmd"]; ok {
		encoded["cmd"] = cmd
	}

	// 复制其他字段
	excludeFields := []string{"device_id", "tenant_id", "device_type", "topic_type", "timestamp", "type", "event", "pose", "track-id", "track_id", "number-people", "number_people", "data", "cmd"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// applySNOMedMapping 已在 snomed_mapping.go 中定义（使用 applyRadarSNOMedMapping）
