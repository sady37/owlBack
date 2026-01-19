package encode

import (
	"encoding/base64"
	"fmt"
)

// RadarDecoder 解码 Radar 数据
// data: 包含原始数据字段（从 MQTT 报文解析）
// topicType: 主题类型 ("monitor", "stat", "event", "alarm")
// 返回: data_value（对象或数组），只包含解码后的数据值，不包含元数据
func RadarDecoder(data map[string]interface{}, topicType string) (interface{}, error) {
	// 根据 topic_type 进行不同的解码转换
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
		// 默认：返回空数组
		return []map[string]interface{}{}, nil
	}
}

// decodeRadarMonitor 解码实时数据
// 返回: data_value（对象或数组），包含 category: "track" 和 category: "vital" 对象
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

	var trackDataList []MonitorTrackData
	if trackBase64 != "" {
		var err error
		trackDataList, err = decodeMonitorTrackFromBase64(trackBase64)
		if err != nil {
			// 解码失败，记录错误但继续处理
		}
	}

	// 2. 构建 track category 对象
	if len(trackDataList) > 0 {
		// 多人情况：每个 track 数据一个对象
		for i, trackData := range trackDataList {
			// 按照厂家文档字节顺序构建对象：target_id, position_x, position_y, position_z, remaining_time, pose, event, area_id
			trackObj := map[string]interface{}{
				"category":       "track",
				"target_id":      trackData.TargetID,
				"position_x":     trackData.PositionX,
				"position_y":     trackData.PositionY,
				"position_z":     trackData.PositionZ,
				"remaining_time": trackData.RemainingTime,
			}

			// SNOMED 映射（按照字节顺序：pose在event之前）
			applyRadarSNOMedMapping(trackObj, "pose", "monitor.track.pose", trackData.Pose)
			applyRadarSNOMedMapping(trackObj, "event", "monitor.track.event", trackData.Event)
			trackObj["area_id"] = trackData.AreaID

			// 只在最后一个 track 对象中添加 raw_original（如果有）
			if i == len(trackDataList)-1 && trackBase64 != "" {
				trackObj["raw_original"] = trackBase64
			}

			dataValue = append(dataValue, trackObj)
		}
	} else {
		// 兼容性处理：如果字段已经存在（已解码），直接使用
		hasTrackData := false
		trackObj := make(map[string]interface{})
		trackObj["category"] = "track"

		// 位置坐标
		if posX, ok := data["position_x"]; ok {
			if x, err := parseInt(posX); err == nil {
				if conv, err := GetFieldConversion("monitor.track.position_x"); err == nil && conv.UnitConversion != nil {
					trackObj["position_x"] = x * 10 // dm -> cm
				} else {
					trackObj["position_x"] = x
				}
				hasTrackData = true
			} else {
				trackObj["position_x"] = posX
				hasTrackData = true
			}
		}
		if posY, ok := data["position_y"]; ok {
			if y, err := parseInt(posY); err == nil {
				if conv, err := GetFieldConversion("monitor.track.position_y"); err == nil && conv.UnitConversion != nil {
					trackObj["position_y"] = y * 10 // dm -> cm
				} else {
					trackObj["position_y"] = y
				}
				hasTrackData = true
			} else {
				trackObj["position_y"] = posY
				hasTrackData = true
			}
		}
		if posZ, ok := data["position_z"]; ok {
			trackObj["position_z"] = posZ
			hasTrackData = true
		}
		if targetID, ok := data["target_id"]; ok {
			trackObj["target_id"] = targetID
			hasTrackData = true
		}
		if areaID, ok := data["area_id"]; ok {
			trackObj["area_id"] = areaID
			hasTrackData = true
		}
		if remainingTime, ok := data["remaining_time"]; ok {
			trackObj["remaining_time"] = remainingTime
			hasTrackData = true
		}

		// 姿态和事件（SNOMED 映射）
		if pose, ok := data["pose"]; ok {
			applyRadarSNOMedMapping(trackObj, "pose", "monitor.track.pose", pose)
			hasTrackData = true
		}
		if event, ok := data["event"]; ok {
			applyRadarSNOMedMapping(trackObj, "event", "monitor.track.event", event)
			hasTrackData = true
		}

		if hasTrackData {
			dataValue = append(dataValue, trackObj)
		}
	}

	// 3. 处理 vital 数据（呼吸心率）
	var bhBase64 string
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if bh, ok := dataField["bh"].(string); ok && bh != "" {
			bhBase64 = bh
		}
	} else if bh, ok := data["bh"].(string); ok && bh != "" {
		bhBase64 = bh
	}

	hasVitalData := false
	vitalObj := make(map[string]interface{})
	vitalObj["category"] = "vital"

	if bhBase64 != "" {
		// 解码 base64
		// 按照标准文档顺序：category, vital_flag, respiratory_rate, heart_rate, sleep_status, stability, raw_original
		bhBytes, err := base64.StdEncoding.DecodeString(bhBase64)
		if err == nil && len(bhBytes) >= 14 {
			// 提取 vital_flag (字节 0)
			vitalObj["vital_flag"] = 0
			hasVitalData = true

			// 提取呼吸率和心率（字节 1 和 2）
			if len(bhBytes) > 2 {
				vitalObj["respiratory_rate"] = int(bhBytes[1])
				vitalObj["heart_rate"] = int(bhBytes[2])
				hasVitalData = true
			}

			// 提取 sleep_status (字节 13, bit 7:6)
			if len(bhBytes) > 13 {
				byte13 := int(bhBytes[13])
				if sleepStatusBits, err := ParseBitField(byte13, "7:6"); err == nil {
					applyRadarSNOMedMapping(vitalObj, "sleep_status", "monitor.vital.sleep_status", sleepStatusBits)
					hasVitalData = true
				}
				// 提取 stability (字节 14, bit 1:0)
				if len(bhBytes) > 14 {
					byte14 := int(bhBytes[14])
					if stabilityBits, err := ParseBitField(byte14, "1:0"); err == nil {
						applyRadarSNOMedMapping(vitalObj, "stability", "monitor.vital.stability", stabilityBits)
						hasVitalData = true
					}
				}
			}

			// raw_original 放在最后
			vitalObj["raw_original"] = bhBase64
		}
	} else {
		// 兼容性处理：如果字段已经存在（已解码），直接使用
		if vitalFlag, ok := data["vital_flag"]; ok {
			vitalObj["vital_flag"] = vitalFlag
			hasVitalData = true
		}
		if respRate, ok := data["respiratory_rate"]; ok {
			vitalObj["respiratory_rate"] = respRate
			hasVitalData = true
		}
		if heartRate, ok := data["heart_rate"]; ok {
			vitalObj["heart_rate"] = heartRate
			hasVitalData = true
		}
		if sleepStatus, ok := data["sleep_status"]; ok {
			applyRadarSNOMedMapping(vitalObj, "sleep_status", "monitor.vital.sleep_status", sleepStatus)
			hasVitalData = true
		}
		if stability, ok := data["stability"]; ok {
			applyRadarSNOMedMapping(vitalObj, "stability", "monitor.vital.stability", stability)
			hasVitalData = true
		}
	}

	if hasVitalData {
		dataValue = append(dataValue, vitalObj)
	}

	// 4. 返回 data_value
	if len(dataValue) == 0 {
		// 如果没有数据，返回空对象
		return map[string]interface{}{}, nil
	} else if len(dataValue) == 1 {
		// 单个对象：直接返回对象
		return dataValue[0], nil
	} else {
		// 多个对象：返回数组
		return dataValue, nil
	}
}

// decodeRadarStat 解码统计数据
// 返回: data_value（对象或数组），包含 category: "track" 和 category: "sleep" 对象
// 注意：解码器只负责数据格式转换，不处理告警判断。告警判断应在 MQTT Consumer 层根据设备告警使能配置进行
func decodeRadarStat(data map[string]interface{}) (interface{}, error) {
	var dataValue []map[string]interface{}

	// 1. 处理 track 统计数据
	// 按照厂家文档字节顺序：version, people_count, walk_distance, walk_duration, lie_duration, stand_duration, multi_person_duration
	hasTrackData := false
	trackObj := make(map[string]interface{})
	trackObj["category"] = "track"

	if version, ok := data["version"]; ok {
		trackObj["version"] = version
		hasTrackData = true
	}
	if peopleCount, ok := data["people_count"]; ok {
		trackObj["people_count"] = peopleCount
		hasTrackData = true
	}
	if walkDistance, ok := data["walk_distance"]; ok {
		if dist, err := parseInt(walkDistance); err == nil {
			// walk_distance 单位：米，不转换为厘米（根据标准文档）
			trackObj["walk_distance"] = dist
			hasTrackData = true
		} else {
			trackObj["walk_distance"] = walkDistance
			hasTrackData = true
		}
	}
	if walkDuration, ok := data["walk_duration"]; ok {
		trackObj["walk_duration"] = walkDuration
		hasTrackData = true
	}
	if lieDuration, ok := data["lie_duration"]; ok {
		trackObj["lie_duration"] = lieDuration
		hasTrackData = true
	}
	if standDuration, ok := data["stand_duration"]; ok {
		trackObj["stand_duration"] = standDuration
		hasTrackData = true
	}
	if multiPersonDuration, ok := data["multi_person_duration"]; ok {
		trackObj["multi_person_duration"] = multiPersonDuration
		hasTrackData = true
	}

	if hasTrackData {
		dataValue = append(dataValue, trackObj)
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

	hasSleepData := false
	sleepObj := make(map[string]interface{})
	sleepObj["category"] = "sleep"

	if sleepBase64 != "" {
		// 解码 base64
		// 按照厂家文档字节顺序：sleep_flag(0), respiratory_rate(1), heart_rate(2), avg_respiratory_rate(5), avg_heart_rate(6), hr_breath_event(13)
		sleepBytes, err := base64.StdEncoding.DecodeString(sleepBase64)
		if err == nil && len(sleepBytes) >= 14 {
			sleepObj["sleep_flag"] = 255
			hasSleepData = true

			// 提取生命体征字段（字节1,2,5,6）
			if len(sleepBytes) > 6 {
				sleepObj["respiratory_rate"] = int(sleepBytes[1])
				sleepObj["heart_rate"] = int(sleepBytes[2])
				sleepObj["avg_respiratory_rate"] = int(sleepBytes[5])
				sleepObj["avg_heart_rate"] = int(sleepBytes[6])
				hasSleepData = true
			}

			// 提取 hr_breath_event (字节 13) 的各个位字段（按照bit顺序：breath_state, heart_state, vital_signs_state, sleep_state）
			if len(sleepBytes) > 13 {
				byte13 := int(sleepBytes[13])

				// breath_state (bit 1:0) - 呼吸状态
				if breathStateBits, err := ParseBitField(byte13, "1:0"); err == nil {
					applyRadarSNOMedMapping(sleepObj, "breath_state", "statistics.sleep.breath_state", breathStateBits)
					hasSleepData = true
				}
				// heart_state (bit 3:2) - 心率状态
				if heartStateBits, err := ParseBitField(byte13, "3:2"); err == nil {
					applyRadarSNOMedMapping(sleepObj, "heart_state", "statistics.sleep.heart_state", heartStateBits)
					hasSleepData = true
				}
				// vital_signs_state (bit 5:4) - 生命体征情况
				if vitalSignsStateBits, err := ParseBitField(byte13, "5:4"); err == nil {
					// 使用临时对象获取映射后的值
					tempObj := make(map[string]interface{})
					applyRadarSNOMedMapping(tempObj, "vital_signs_state", "statistics.sleep.vital_signs_state", vitalSignsStateBits)
					vitalSignsStateMapped := tempObj["vital_signs_state"]
					if vitalSignsStateMapped == nil {
						vitalSignsStateMapped = vitalSignsStateBits // 如果没有映射，使用原始值
					}
					sleepObj["vital_signs_state"] = vitalSignsStateMapped
					hasSleepData = true
				}
				// sleep_state (bit 7:6) - 睡眠状态
				if sleepStateBits, err := ParseBitField(byte13, "7:6"); err == nil {
					applyRadarSNOMedMapping(sleepObj, "sleep_state", "statistics.sleep.sleep_state", sleepStateBits)
					hasSleepData = true
				}
			}

			// raw_original 放在最后
			sleepObj["raw_original"] = sleepBase64
		}
	}

	if hasSleepData {
		dataValue = append(dataValue, sleepObj)
	}

	// 3. 返回 data_value
	if len(dataValue) == 0 {
		// 如果没有数据，返回空对象
		return map[string]interface{}{}, nil
	} else if len(dataValue) == 1 {
		// 单个对象：直接返回对象
		return dataValue[0], nil
	} else {
		// 多个对象：返回数组
		return dataValue, nil
	}
}

// decodeRadarEvent 解码事件数据
// 返回: data_value（数组），包含 category: "enter2out", "pose", "number-people", "isOnline", "signal_poor", "angle_abnormal", "other" 对象
// 支持系统1格式：cmd="event", type=1/2/3/5/7/8/9 (int)
// 注意：解码器只负责数据格式转换，不处理告警判断。告警判断应在 MQTT Consumer 层根据设备告警使能配置进行
func decodeRadarEvent(data map[string]interface{}) (interface{}, error) {
	var dataValue []map[string]interface{}

	// 系统1格式：事件类型：1-进出事件，2-姿态变化事件，3-人数变化事件，5-设备在线状态，7-信号差，8-倾角异常，9-其他告警
	eventType, _ := data["type"].(int)

	// 处理 data 字段（可能是数组）
	if dataField, ok := data["data"]; ok {
		// data 可能是数组，包含多个事件
		if dataArray, ok := dataField.([]interface{}); ok {
			// 处理数组中的每个事件
			for _, eventItem := range dataArray {
				if eventMap, ok := eventItem.(map[string]interface{}); ok {
					eventObj := make(map[string]interface{})

					// 根据事件类型处理
					if eventType == 1 {
						// type=1: 进出事件 → category="enter2out"
						eventObj["category"] = "enter2out"
						if trackID, ok := eventMap["track-id"]; ok {
							eventObj["track_id"] = trackID
						} else if trackID, ok := eventMap["track_id"]; ok {
							eventObj["track_id"] = trackID
						}
						if event, ok := eventMap["event"]; ok {
							applyRadarSNOMedMapping(eventObj, "event", "event.enter2out.event", event)
						}
						if areaType, ok := eventMap["area_type"]; ok {
							// area_type 需要转换为 display_en 值（根据标准文档）
							eventObj["area_type"] = mapAreaType(areaType)
						}
					} else if eventType == 2 {
						// type=2: 姿态变化事件 → category="pose"
						eventObj["category"] = "pose"
						if trackID, ok := eventMap["track-id"]; ok {
							eventObj["track_id"] = trackID
						} else if trackID, ok := eventMap["track_id"]; ok {
							eventObj["track_id"] = trackID
						}
						if pose, ok := eventMap["pose"]; ok {
							// 使用临时对象获取映射后的值
							tempObj := make(map[string]interface{})
							applyRadarSNOMedMapping(tempObj, "pose", "event.pose.pose", pose)
							poseMapped := tempObj["pose"]
							if poseMapped == nil {
								poseMapped = pose // 如果没有映射，使用原始值
							}
							eventObj["pose"] = poseMapped
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
					} else if eventType == 8 {
						// type=8: 倾角异常事件 → category="angle_abnormal"
						eventObj["category"] = "angle_abnormal"
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
					} else if eventType == 9 {
						// type=9: 其他告警 → category="other"
						eventObj["category"] = "other"
					}

					// 保留其他字段（排除已处理的字段）
					excludedKeys := map[string]bool{
						"track-id": true,
						"isOnline": true,
						"recovery": true,
					}
					for k, v := range eventMap {
						if _, exists := eventObj[k]; !exists && !excludedKeys[k] {
							eventObj[k] = v
						}
					}

					if len(eventObj) > 1 { // 至少有 category 和其他字段
						dataValue = append(dataValue, eventObj)
					}
				}
			}
		} else if dataMap, ok := dataField.(map[string]interface{}); ok {
			// data 可能是单个对象
			if eventType == 3 {
				// type=3: 人数变化事件
				eventObj := make(map[string]interface{})
				eventObj["category"] = "number-people"
				if numberPeople, ok := dataMap["number-people"]; ok {
					eventObj["number_people"] = numberPeople
				} else if numberPeople, ok := dataMap["number_people"]; ok {
					eventObj["number_people"] = numberPeople
				}
				dataValue = append(dataValue, eventObj)
			} else if eventType == 5 {
				// type=5: 设备在线状态 → category="isOnline"
				eventObj := make(map[string]interface{})
				eventObj["category"] = "isOnline"
				// 转换 isOnline 字段：0="online", 1="offline"
				if isOnline, ok := dataMap["isOnline"]; ok {
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
				for k, v := range dataMap {
					if k != "category" && k != "isOnline" {
						eventObj[k] = v
					}
				}
				dataValue = append(dataValue, eventObj)
			} else if eventType == 7 {
				// type=7: 信号差事件 → category="signal_poor"
				eventObj := make(map[string]interface{})
				eventObj["category"] = "signal_poor"
				// 转换 recovery 字段：0="signal_poor", 1="signal_recovery"
				if recovery, ok := dataMap["recovery"]; ok {
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
				for k, v := range dataMap {
					if k != "category" && k != "recovery" {
						eventObj[k] = v
					}
				}
				dataValue = append(dataValue, eventObj)
			} else if eventType == 8 {
				// type=8: 倾角异常事件 → category="angle_abnormal"
				eventObj := make(map[string]interface{})
				eventObj["category"] = "angle_abnormal"
				// 转换 recovery 字段：0="angle_abnormal", 1="angle_recovery"
				if recovery, ok := dataMap["recovery"]; ok {
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
				for k, v := range dataMap {
					if k != "category" && k != "recovery" {
						eventObj[k] = v
					}
				}
				dataValue = append(dataValue, eventObj)
			} else if eventType == 9 {
				// type=9: 其他告警 → category="other"
				eventObj := make(map[string]interface{})
				eventObj["category"] = "other"
				// 直接复制原始字段，不做值转换
				for k, v := range dataMap {
					if k != "category" {
						eventObj[k] = v
					}
				}
				dataValue = append(dataValue, eventObj)
			}
		}
	}

	// 返回 data_value
	if len(dataValue) == 0 {
		return []map[string]interface{}{}, nil
	} else {
		return dataValue, nil
	}
}

// mapAreaType 映射 area_type 值到 display_en
// 根据标准文档：2="Bed", 5="Monitoring bed", 6="Sensor area"
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

// decodeRadarAlarm 解码告警数据
// 当前为空函数，因为解码器只负责数据格式转换，不处理告警判断
// 告警判断及发布到 Redis 应在 MQTT Consumer 层根据设备告警使能配置进行
func decodeRadarAlarm(data map[string]interface{}) (interface{}, error) {
	// 空函数，返回空数组
	return []map[string]interface{}{}, nil
}

// applySNOMedMapping 已在 snomed_mapping.go 中定义（使用 applyRadarSNOMedMapping）

// MonitorTrackData 解码后的 monitor track 数据
type MonitorTrackData struct {
	TargetID      int // 字节 0: 目标 ID (0-7 或 88 表示无人)
	PositionX     int // 字节 1: x 坐标（分米，已转换为厘米）
	PositionY     int // 字节 2: y 坐标（分米，已转换为厘米）
	PositionZ     int // 字节 3: z 坐标（厘米）
	RemainingTime int // 字节 12: 剩余时间（秒，0-60）
	Pose          int // 字节 13: 姿态值 (0-11)
	Event         int // 字节 14: 事件值 (0-4)
	AreaID        int // 字节 15: 区域 ID
}

// decodeMonitorTrackFromBase64 解码 monitor track 字段
// base64Track: base64 编码的 track 字符串
// 返回: 解码后的 track 数据数组（每个人一个元素）
func decodeMonitorTrackFromBase64(base64Track string) ([]MonitorTrackData, error) {
	// 1. Base64 解码
	trackBytes, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, err
	}

	// 2. 检查长度是否为 16 的倍数
	if len(trackBytes)%16 != 0 {
		return nil, fmt.Errorf("invalid track length: %d (must be multiple of 16)", len(trackBytes))
	}

	// 3. 按 16 字节分段处理（每个人）
	personCount := len(trackBytes) / 16
	results := make([]MonitorTrackData, personCount)

	for i := 0; i < personCount; i++ {
		offset := i * 16
		personData := trackBytes[offset : offset+16]

		// 字节 0: target_id
		targetID := int(personData[0])

		// 字节 1: position_x (有符号数，分米，转换为厘米)
		positionX := int(int8(personData[1])) // 转换为有符号数
		positionXCm := positionX * 10         // 分米 → 厘米

		// 字节 2: position_y (有符号数，分米，转换为厘米)
		positionY := int(int8(personData[2])) // 转换为有符号数
		positionYCm := positionY * 10         // 分米 → 厘米

		// 字节 3: position_z (厘米)
		positionZ := int(personData[3])

		// 字节 12: remaining_time (秒)
		remainingTime := int(personData[12])

		// 字节 13: pose (姿态值)
		pose := int(personData[13])

		// 字节 14: event (事件值)
		event := int(personData[14])

		// 字节 15: area_id
		areaID := int(personData[15])

		results[i] = MonitorTrackData{
			TargetID:      targetID,
			PositionX:     positionXCm,
			PositionY:     positionYCm,
			PositionZ:     positionZ,
			RemainingTime: remainingTime,
			Pose:          pose,
			Event:         event,
			AreaID:        areaID,
		}
	}

	return results, nil
}
