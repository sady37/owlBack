package decode

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/radar"
	//"owl-common/utils"
)

// RadarDecoder 解码 Radar 数据（与订阅无关，负责把所有收到的 MQTT 做统一转换）
// topicType: "monitor"|"stat"|"event"|"alarm"|"prop"|"func"
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
	case "prop":
		return decodeRadarProp(data)
	case "func":
		return decodeRadarFunc(data)
	default:
		return []map[string]interface{}{}, nil
	}
}

// normalizeRequestID 归一化 request_id 字段（requestId/requestID/request_id → 统一 request_id，去重）
func normalizeRequestID(data map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		out[k] = v
	}

	if id, ok := data["requestId"].(string); ok && id != "" {
		out["request_id"] = id
	} else if id, ok := data["requestID"].(string); ok && id != "" {
		out["request_id"] = id
	} else if id, ok := data["request_id"].(string); ok && id != "" {
		out["request_id"] = id
	}
	delete(out, "requestId")
	delete(out, "requestID")

	return out
}

// decodeRadarProp 解码属性响应（/prop/.../post）：归一化 request_id，data 做 key/单位规范化后返回
func decodeRadarProp(data map[string]interface{}) (interface{}, error) {
	resp := &radar.PropResponse{Data: make(map[string]interface{})}
	if id, ok := data["requestId"].(string); ok && id != "" {
		resp.RequestID = id
	} else if id, ok := data["requestID"].(string); ok && id != "" {
		resp.RequestID = id
	} else if id, ok := data["request_id"].(string); ok && id != "" {
		resp.RequestID = id
	}
	if dataMap, ok := data["data"].(map[string]interface{}); ok {
		resp.Data = normalizePropData(dataMap)
	}
	return resp, nil
}

// normalizePropData 将设备 prop data 转为规范 key 与单位（与 radar.Key* 一致，单位 cm）。
// 设备 key → 规范 key：radar_install_height(dm)→install_height(cm), radar_install_style→install_model,
// radar_func_ctrl→work_model, rectangle(dm)→rectangle(cm), ssid_password→wifi_ssid+wifi_password；其余透传。
func normalizePropData(data map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(data)+2)
	for k, v := range data {
		if v == nil {
			continue
		}
		switch k {
		case "radar_install_height":
			if n, ok := propNum(v); ok {
				out[radar.KeyInstallHeight] = n * 10 // dm → cm
			}
		case "radar_install_style":
			if n, ok := propNum(v); ok {
				out[radar.KeyInstallModel] = n // 0/1/2
			}
		case "radar_func_ctrl":
			if n, ok := propNum(v); ok {
				out[radar.KeyWorkModel] = n // 3track/7fall/11sleep/15func
			}
		case "rectangle":
			if s, ok := v.(string); ok && s != "" {
				if cm := rectangleDMToCM(s); cm != "" {
					out[radar.KeyRectangle] = cm
				} else {
					out[radar.KeyRectangle] = s
				}
			}
		case "declare_area":
			if s, ok := v.(string); ok && s != "" {
				if cm := declareAreaDMToCM(s); cm != "" {
					out[radar.KeyDeclareArea] = cm
				} else {
					out[radar.KeyDeclareArea] = v
				}
			} else {
				out[radar.KeyDeclareArea] = v
			}
		case "fall_param", "heart_breath_param":
			out[k] = v // base64 透传，下游用 DecodeFallParam/DecodeHeartBreathParam
		case "accelera", "wifi_rssi", "ip_port", "run_Horizontal", "voice_end_tip":
			out[k] = v
		case "ssid_password":
			if s, ok := v.(string); ok && s != "" {
				if idx := strings.Index(s, ":"); idx >= 0 {
					out[radar.KeyWifiSsid] = strings.TrimSpace(s[:idx])
					out[radar.KeyWifiPassword] = strings.TrimSpace(s[idx+1:])
				} else {
					out[radar.KeyWifiSsid] = s
				}
			}
		default:
			out[k] = v
		}
	}
	return out
}

func propNum(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		return n, err == nil
	default:
		return 0, false
	}
}

// declareAreaDMToCM 解析设备 declare_area 字符串（分米），转为 cm 字符串。多区域用逗号分隔，格式 {area-id,area-type,x1,y1;x2,y2;x3,y3;x4,y4}[,{...}]
func declareAreaDMToCM(s string) string {
	s = strings.TrimSpace(s)
	for len(s) > 0 && (s[len(s)-1] == ',' || s[len(s)-1] == ';') {
		s = s[:len(s)-1]
	}
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "},{")
	var out []string
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if i > 0 && p[0] != '{' {
			p = "{" + p
		}
		if i < len(parts)-1 && len(p) > 0 && p[len(p)-1] != '}' {
			p = p + "}"
		}
		if !strings.HasPrefix(p, "{") {
			p = "{" + p
		}
		if !strings.HasSuffix(p, "}") {
			p = p + "}"
		}
		d, ok := radar.ParseDeclareAreaFromDM(p)
		if !ok {
			return ""
		}
		out = append(out, d.String())
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, ",")
}

// rectangleDMToCM 解析设备 rectangle 字符串（分米），转为 cm 字符串。格式 {x1,y1;x2,y2;x3,y3;x4,y4} 或含空格。
func rectangleDMToCM(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return ""
	}
	if s[0] == '{' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '}' {
		s = s[:len(s)-1]
	}
	points := strings.Split(s, ";")
	if len(points) != 4 {
		return ""
	}
	var nums []int
	for _, p := range points {
		parts := strings.Split(strings.TrimSpace(p), ",")
		if len(parts) != 2 {
			return ""
		}
		x, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		y, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return ""
		}
		nums = append(nums, x*10, y*10)
	}
	return fmt.Sprintf("{%d,%d;%d,%d;%d,%d;%d,%d}",
		nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7])
}

// decodeRadarFunc 解码功能响应（/func/.../post）：归一化 request_id，返回结构体
func decodeRadarFunc(data map[string]interface{}) (interface{}, error) {
	resp := &radar.FuncResponse{}
	if id, ok := data["requestId"].(string); ok && id != "" {
		resp.RequestID = id
	} else if id, ok := data["requestID"].(string); ok && id != "" {
		resp.RequestID = id
	} else if id, ok := data["request_id"].(string); ok && id != "" {
		resp.RequestID = id
	}
	return resp, nil
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
			for _, trackData := range trackDataList {
				trackObj := map[string]interface{}{
					"dataCategory":   "track",
					"track_id":       trackData.TrackID,
					"position_x":     trackData.PositionX,
					"position_y":     trackData.PositionY,
					"position_z":     trackData.PositionZ,
					"remaining_time": trackData.RemainingTime,
					"area_id":        trackData.AreaID,
				}
				trackObj["pose"] = trackData.Pose
				trackObj["event"] = trackData.Event
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
				"dataCategory":     "vital",
				"vital_flag":       vitalData.VitalFlag,
				"respiratory_rate": vitalData.RespiratoryRate,
				"heart_rate":       vitalData.HeartRate,
				// "raw_original":     bhBase64,
			}
			vitalObj["sleep_status"] = vitalData.SleepStatus
			vitalObj["stability"] = vitalData.Stability
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

// toFloat64 从 map 取值转为 float64，支持 float64/int/int32/int64/string；失败返回 (0,false)
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// decodeRadarStat 解码统计数据（单位：长度 m/cm，时长 秒/分/小时；walk_distance 保持 m，各 duration 整型秒）
// 协议 3.6：data.data.track / data.data.sleep 为 base64；无 base64 时兼容顶层 version/people_count 等。
/*
[
  {
    "dataCategory": "track",
    "version": 1,
    "people_count": 1,
    "walk_distance": 12.5,
    "walk_duration": 300,
    "lie_duration": 600,
    "stand_duration": 100,
    "multi_person_duration": 0
  },
  {
    "dataCategory": "sleep",
    "sleep_flag": 1,
    "respiratory_rate": 16,
    "heart_rate": 72,
    "avg_respiratory_rate": 15,
    "avg_heart_rate": 70,
    "breath_state": 0,
    "heart_state": 0,
    "vital_signs_state": 0,
    "sleep_state": 2,
    "stat_numeric_codes": ["sleep_breath_01", "sleep_vitals_11"]   //hr,rr,vital
  }
]
*/
func decodeRadarStat(data map[string]interface{}) (interface{}, error) {
	var dataValue []map[string]interface{}

	// 1. 处理活动性统计（协议字段名 data.track，内容为 people_count/walk_distance 等，与 monitor 轨迹不同）：优先从 base64 解码
	var trackBase64 string
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if track, ok := dataField["track"].(string); ok && track != "" {
			trackBase64 = track
		}
	} else if track, ok := data["track"].(string); ok && track != "" {
		trackBase64 = track
	}
	// stat 协议里该 base64 块叫 "track"，但内容为老人活动性统计，与 monitor 的轨迹(track)语义不同，故 dataCategory 用 "activity"
	if trackBase64 != "" {
		trackData, err := decodeStatTrackFromBase64(trackBase64)
		if err == nil {
			trackObj := map[string]interface{}{
				"dataCategory":          "activity",
				"version":               trackData.Version,
				"people_count":          trackData.PeopleCount,
				"walk_distance":         trackData.WalkDistance,
				"walk_duration":         trackData.WalkDuration,
				"lie_duration":          trackData.LieDuration,
				"stand_duration":        trackData.StandDuration,
				"multi_person_duration": trackData.MultiPersonDuration,
				// "raw_original":          trackBase64,
			}
			dataValue = append(dataValue, trackObj)
		}
	}
	// 无 base64 track 时兼容顶层字段
	if len(dataValue) == 0 && data["version"] != nil {
		trackObj := map[string]interface{}{"dataCategory": "activity", "version": data["version"]}
		if peopleCount, ok := data["people_count"]; ok {
			trackObj["people_count"] = peopleCount
		}
		if v := data["walk_distance"]; v != nil {
			if f, ok := toFloat64(v); ok {
				trackObj["walk_distance"] = f
			}
		}
		if v := data["walk_duration"]; v != nil {
			if f, ok := toFloat64(v); ok {
				trackObj["walk_duration"] = int(math.Round(f))
			}
		}
		if v := data["lie_duration"]; v != nil {
			if f, ok := toFloat64(v); ok {
				trackObj["lie_duration"] = int(math.Round(f))
			}
		}
		if v := data["stand_duration"]; v != nil {
			if f, ok := toFloat64(v); ok {
				trackObj["stand_duration"] = int(math.Round(f))
			}
		}
		if v := data["multi_person_duration"]; v != nil {
			if f, ok := toFloat64(v); ok {
				trackObj["multi_person_duration"] = int(math.Round(f))
			}
		}
		if len(trackObj) > 2 {
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
				"dataCategory":         "sleep",
				"sleep_flag":           sleepData.SleepFlag,
				"respiratory_rate":     sleepData.RespiratoryRate,
				"heart_rate":           sleepData.HeartRate,
				"avg_respiratory_rate": sleepData.AvgRespiratoryRate,
				"avg_heart_rate":       sleepData.AvgHeartRate,
				// "raw_original":         sleepBase64,
			}
			sleepObj["breath_state"] = sleepData.BreathStatus
			sleepObj["heart_state"] = sleepData.HeartStatus
			sleepObj["vital_signs_state"] = sleepData.VitalSignsStatus
			sleepObj["sleep_state"] = sleepData.SleepStatus

			// byte13 的 4 组 2-bit 状态（protocol.go StatSleepData 已按 bit 拆分）
			breathState := sleepData.BreathStatus & 0x03         // bit[1:0] 呼吸: 00=正常 01=过低 10=过高 11=暂停
			heartState := sleepData.HeartStatus & 0x03           // bit[3:2] 心率: 00=正常 01=过低 10=过高 11=未定义
			vitalSignsState := sleepData.VitalSignsStatus & 0x03 // bit[5:4] 生命体征: 00=正常 01=未定义 10=未定义 11=弱
			sleepState := sleepData.SleepStatus & 0x03           // bit[7:6] 睡眠: 00=未定义 01=浅睡 10=深睡 11=清醒

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
			if sleepState != 0 { // 01=浅睡 10=深睡 11=清醒
				sleepObj["stat_numeric_code_sleep"] = fmt.Sprintf("sleep_sleep_%02d", sleepState)
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
// 协议参考：docs/Radar_MQTT_v3.0.md 3.5 事件告警上报
// type=1 进出事件 data 为数组（多 track）；type=2 姿态变化 data 为数组（多 track）
// type=3 人数变化 data 为对象；type=5/7/8 设备状态 data 为对象
// 返回 []interface{}，每个元素为一条 track/item 的 map[string]interface{}
/*
items 是 []interface{}，每个元素是一个 map[string]interface{}。
举例，type=1 有 2 条 track, type=2 有 1 条 track, type=5 有 1 条 status:
[
  {"event_name": "InBed", "event_type": 1, "track_id": 0, "event": 3, "area_type": 2},
  {"event_name": "LeftBed", "event_type": 1, "track_id": 1, "event": 4, "area_type": 2},
  {"event_name": "Fall", "event_type": 2, "track_id": 2, "pose": 5},
  {"event_name": "Offline", "event_type": 5, "status_type": "...", "status_value": "1"}
]
// event_name 全集（与 alarm.Enter2OutMap / PoseMap / NumberPeopleCategory / buildStatus 一致）：
//   - event_type=1 进出: EnterRoom, ExitRoom, InBed, LeftBed, EnterSensingArea, ExitSensingArea, EnterMonitor, ExitMonitor
//   - event_type=2 姿态: Initialization, Walking, Fall, Sitting, Standing, Lying, SittingOnGround, BedSitUp
//   - event_type=3 人数: number-people
//   - event_type=5: OfflineAlarm
//   - event_type=7: SignalPoor
//   - event_type=8: AngleException
*/
func decodeRadarEvent(data map[string]interface{}) ([]interface{}, error) {
	eventType := toInt(data["type"])
	log.Printf("[DECODE_EVENT_DEBUG] raw data: type=%v, data=%v", eventType, data["data"])

	var items []interface{}

	dataField, hasData := data["data"]
	if !hasData {
		dataField = data
	}

	switch eventType {
	case 1:
		if arr, ok := dataField.([]interface{}); ok {
			for _, elem := range arr {
				if m, ok := elem.(map[string]interface{}); ok {
					items = append(items, buildEnter2OutTrack(m))
				}
			}
		}
	case 2:
		if arr, ok := dataField.([]interface{}); ok {
			for _, elem := range arr {
				if m, ok := elem.(map[string]interface{}); ok {
					items = append(items, buildPoseTrack(m))
				}
			}
		}
	case 3:
		if m, ok := dataField.(map[string]interface{}); ok {
			items = append(items, buildPeopleNumber(m))
		}
	case 5, 7, 8:
		if m, ok := dataField.(map[string]interface{}); ok {
			items = append(items, buildStatus(eventType, m))
		}
	}

	return items, nil
}

// decodeRadarAlarm 解码告警数据（与 event 相同）
func decodeRadarAlarm(data map[string]interface{}) ([]interface{}, error) {
	return decodeRadarEvent(data)
}

// buildEnter2OutTrack type=1 进出事件
func buildEnter2OutTrack(m map[string]interface{}) map[string]interface{} {
	event := toInt(m["event"])
	areaType := toInt(m["area_type"])
	dataCategory, _ := alarm.LookupEnter2Out(event, areaType)
	return map[string]interface{}{
		"event_name": dataCategory,
		"event_type": 1,
		"track_id":   toInt(m["track-id"]),
		"event":      event,
		"area_type":  areaType,
	}
}

// buildPoseTrack type=2 姿态事件
func buildPoseTrack(m map[string]interface{}) map[string]interface{} {
	pose := toInt(m["pose"])
	dataCategory, _ := alarm.LookupPose(pose)
	return map[string]interface{}{
		"event_name": dataCategory,
		"event_type": 2,
		"track_id":   toInt(m["track-id"]),
		"pose":       pose,
	}
}

// buildPeopleNumber type=3 人数统计；event_name 用 alarm 定义，取值字段用 observation.FieldNumberPeople
func buildPeopleNumber(m map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"event_name":                  alarm.NumberPeople,
		"event_type":                  3,
		observation.FieldNumberPeople: toInt(m["number_people"]),
	}
}

// buildStatus type=5/7/8 设备状态事件
func buildStatus(eventType int, m map[string]interface{}) map[string]interface{} {
	var dataCategory, statusType, statusValue string
	switch eventType {
	case 5:
		dataCategory = alarm.AlarmTypeOffline
		statusType = observation.FieldOffline
		statusValue = "0"
		if toInt(m["isOnline"]) != 0 {
			statusValue = "1"
		}
	case 7:
		dataCategory = alarm.SignalPoor
		statusType = observation.FieldSignalPoor
		statusValue = "1"
		if toInt(m["recovery"]) != 0 {
			statusValue = "0"
		}
	case 8:
		dataCategory = alarm.AngleException
		statusType = observation.FieldAngleAbnormal
		statusValue = "1"
		if toInt(m["recovery"]) != 0 {
			statusValue = "0"
		}
	}

	return map[string]interface{}{
		"event_name":   dataCategory,
		"event_type":   eventType,
		"status_type":  statusType,
		"status_value": statusValue,
	}
}

// toInt 从 interface{} 提取 int 值（兼容 int / float64 / string），无法解析返回 0
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		var n int
		fmt.Sscanf(val, "%d", &n)
		return n
	}
	return 0
}

// mapAreaType 映射 area_type 数值到可读字符串
// 2=Bed 4=Door 5=MonitorBed 6=SensingArea
func mapAreaType(areaType int) string {
	switch areaType {
	case 2:
		return "Bed"
	case 4:
		return "Door"
	case 5:
		return "MonitorBed"
	case 6:
		return "SensingArea"
	default:
		return fmt.Sprintf("%d", areaType)
	}
}

// decodeMonitorTrackFromBase64 解码 monitor track，与 radar.MonitorTrackData 布局一致
func decodeMonitorTrackFromBase64(base64Track string) ([]radar.MonitorTrackData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 track: %w", err)
	}
	if len(data)%16 != 0 {
		return nil, fmt.Errorf("invalid track length: %d (must be multiple of 16)", len(data))
	}
	personCount := len(data) / 16
	results := make([]radar.MonitorTrackData, personCount)
	for i := 0; i < personCount; i++ {
		offset := i * 16
		seg := data[offset : offset+16]
		event := int(seg[14])
		if event > 4 {
			event = 0
		}
		results[i] = radar.MonitorTrackData{
			TrackID:       int(seg[0]),
			PositionX:     int(int8(seg[1])) * 10, //dm to cm
			PositionY:     int(int8(seg[2])) * 10, //dm to cm
			PositionZ:     int(seg[3]),            //cm
			RemainingTime: int(seg[12]),
			Pose:          int(seg[13]),
			Event:         event,
			AreaID:        int(seg[15]),
		}
		copy(results[i].Reserved[:], seg[4:12])
	}
	return results, nil
}

// decodeMonitorVitalFromBase64 解码 monitor vital，与 radar.MonitorVitalData 布局一致
func decodeMonitorVitalFromBase64(base64Bh string) (*radar.MonitorVitalData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Bh)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 vital: %w", err)
	}
	if len(data) < 15 {
		return nil, fmt.Errorf("invalid vital length: %d (must be at least 15)", len(data))
	}
	byte13 := int(data[13])
	sleepStatus := (byte13 >> 6) & 0x03
	stability := int(data[14]) & 0x03
	out := &radar.MonitorVitalData{
		VitalFlag:       0,
		RespiratoryRate: int(data[1]),
		HeartRate:       int(data[2]),
		SleepStatus:     sleepStatus,
		Stability:       stability,
	}
	copy(out.Reserved[:], data[3:13])
	if len(data) > 15 {
		out.Reserved15 = data[15]
	}
	return out, nil
}

// decodeStatSleepFromBase64 解码 stat sleep，与 radar.StatSleepData 布局一致
func decodeStatSleepFromBase64(base64Sleep string) (*radar.StatSleepData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Sleep)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 sleep: %w", err)
	}
	if len(data) < 14 {
		return nil, fmt.Errorf("invalid sleep length: %d (must be at least 14)", len(data))
	}
	byte13 := int(data[13])
	out := &radar.StatSleepData{
		SleepFlag:          255,
		RespiratoryRate:    int(data[1]),
		HeartRate:          int(data[2]),
		AvgRespiratoryRate: int(data[5]),
		AvgHeartRate:       int(data[6]),
		BreathStatus:       byte13 & 0x03,
		HeartStatus:        (byte13 >> 2) & 0x03,
		VitalSignsStatus:   (byte13 >> 4) & 0x03,
		SleepStatus:        (byte13 >> 6) & 0x03,
	}
	copy(out.Reserved_3_4[:], data[3:5])
	copy(out.Reserved_7_12[:], data[7:13])
	if len(data) >= 16 {
		copy(out.Reserved_14_15[:], data[14:16])
	}
	return out, nil
}

// decodeStatTrackFromBase64 解码 stat track，与 radar.StatTrackData 布局一致。定长 16 字节，字节 2~3 为 walk_distance Big-Endian。
func decodeStatTrackFromBase64(base64Track string) (*radar.StatTrackData, error) {
	data, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 stat track: %w", err)
	}
	if len(data) < 9 {
		return nil, fmt.Errorf("invalid stat track length: %d (must be at least 9)", len(data))
	}
	walkDistance := int(data[2])<<8 | int(data[3])
	out := &radar.StatTrackData{
		Version:             int(data[0]),
		PeopleCount:         int(data[1]),
		WalkDistance:        walkDistance,
		WalkDuration:        int(data[4]),
		Reserved5:           data[5],
		LieDuration:         int(data[6]),
		StandDuration:       int(data[7]),
		MultiPersonDuration: int(data[8]),
	}
	if len(data) >= 16 {
		copy(out.Reserved_9_15[:], data[9:16])
	}
	return out, nil
}

// decodeFallParamFromBase64 解码 fall_param base64 为 radar.FallParamData
func decodeFallParamFromBase64(fallParamBase64 string) (*radar.FallParamData, error) {
	data, err := base64.StdEncoding.DecodeString(fallParamBase64)
	if err != nil {
		return nil, err
	}
	if len(data) < 6 {
		return nil, nil
	}
	out := &radar.FallParamData{
		FallDuration10s:    int(data[3]),
		SittingEnabled:     data[4]&1 != 0,
		PostureEnabled:     data[4]>>1&1 != 0,
		BedSitUpEnabled:    data[4]>>2&1 != 0,
		SittingDuration10s: int(data[5]),
	}
	copy(out.Reserved_0_2[:], data[0:3])
	if len(data) >= 16 {
		copy(out.Reserved_6_15[:], data[6:16])
	}
	return out, nil
}

// DecodeFallParam 从 fall_param base64 解码为 AlarmItem 参数（10秒单位 → 秒），保持原有 map 返回供 DecodeDevicePropsToAlarmItems 使用
func DecodeFallParam(fallParamBase64 string) (map[string]interface{}, error) {
	p, err := decodeFallParamFromBase64(fallParamBase64)
	if err != nil || p == nil {
		return nil, err
	}
	result := make(map[string]interface{})
	result[alarm.Fall+".duration_sec"] = p.FallDuration10s * 10
	if p.SittingEnabled {
		result[alarm.SittingOnGround+".is_enabled"] = 1
		result[alarm.SittingOnGround+".duration_sec"] = p.SittingDuration10s * 10
	} else {
		result[alarm.SittingOnGround+".is_enabled"] = 0
	}
	if p.PostureEnabled {
		result[alarm.PostureDetection+".is_enabled"] = 1
	} else {
		result[alarm.PostureDetection+".is_enabled"] = 0
	}
	if p.BedSitUpEnabled {
		result[alarm.LeftBed+".is_enabled"] = 1
	} else {
		result[alarm.LeftBed+".is_enabled"] = 0
	}
	return result, nil
}

// decodeHeartBreathParamFromBase64 解码 heart_breath_param base64 为 radar.HeartBreathParamData
func decodeHeartBreathParamFromBase64(heartBreathParamBase64 string) (*radar.HeartBreathParamData, error) {
	data, err := base64.StdEncoding.DecodeString(heartBreathParamBase64)
	if err != nil {
		return nil, err
	}
	if len(data) < 7 {
		return nil, nil
	}
	out := &radar.HeartBreathParamData{
		RespRateMax:               int(data[0]),
		HeartRateMax:              int(data[1]),
		RespRateMin:               int(data[2]),
		HeartRateMin:              int(data[3]),
		ContinuousDetect:          int(data[4]),
		WeakBiometricSignalDurMin: int(data[5]),
		WeakBiometricSignalSens:   int(data[6]),
	}
	if len(data) >= 16 {
		copy(out.Reserved_7_15[:], data[7:16])
	}
	return out, nil
}

// DecodeHeartBreathParam 从 heart_breath_param base64 解码为 AlarmItem 参数，保持原有 map 返回
func DecodeHeartBreathParam(heartBreathParamBase64 string) (map[string]interface{}, error) {
	p, err := decodeHeartBreathParamFromBase64(heartBreathParamBase64)
	if err != nil || p == nil {
		return nil, err
	}
	result := make(map[string]interface{})
	result[alarm.RespRateAlert+".max"] = p.RespRateMax
	result[alarm.HeartRateAlert+".max"] = p.HeartRateMax
	result[alarm.RespRateAlert+".min"] = p.RespRateMin
	result[alarm.HeartRateAlert+".min"] = p.HeartRateMin
	result[alarm.WeakBiometricSignal+".duration_min"] = p.WeakBiometricSignalDurMin
	result[alarm.WeakBiometricSignal+".sensitivity"] = p.WeakBiometricSignalSens
	return result, nil
}

// DecodeDevicePropsToAlarmItems 从设备属性更新 AlarmItems（用于设备写入失败时同步实际值）
// 支持工作模式（radar_func_ctrl）、跌倒参数（fall_param）和呼吸心率参数（heart_breath_param）
func DecodeDevicePropsToAlarmItems(items []alarm.AlarmItem, deviceProps map[string]interface{}) ([]alarm.AlarmItem, error) {
	result := make([]alarm.AlarmItem, len(items))
	copy(result, items)

	// 支持规范 key work_model 与设备 key radar_func_ctrl（prop 解码后为 work_model）
	workModeRaw, _ := deviceProps[radar.KeyWorkModel]
	if workModeRaw == nil {
		workModeRaw = deviceProps["radar_func_ctrl"]
	}
	if workModeRaw != nil {
		var mode int
		switch v := workModeRaw.(type) {
		case float64:
			mode = int(v)
		case int:
			mode = v
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				mode = parsed
			} else {
				mode = 0
			}
		default:
			mode = 0
		}
		if mode != 0 {
			for i := range result {
				if result[i].AlarmType == alarm.MonitoringMode {
					if result[i].AlarmParams == nil {
						result[i].AlarmParams = make(map[string]interface{})
					}
					result[i].AlarmParams["mode"] = mode
					break
				}
			}
		}
	}

	fallParam, fallParamExists := deviceProps["fall_param"]
	if fallParamExists && fallParam != nil {
		if fallParamStr, ok := fallParam.(string); ok && fallParamStr != "" {
			decodedParams, err := DecodeFallParam(fallParamStr)
			if err == nil && decodedParams != nil {
				for key, value := range decodedParams {
					parts := splitAlarmItemKey(key)
					if len(parts) == 2 {
						alarmType, paramKey := parts[0], parts[1]
						if paramKey == "is_enabled" {
							updateAlarmItemEnabled(&result, alarmType, value.(int))
						} else {
							updateAlarmItemParam(&result, alarmType, paramKey, value)
						}
					}
				}
			}
		} else {
			disableAlarmItem(&result, alarm.Fall)
			disableAlarmItem(&result, alarm.SittingOnGround)
			disableAlarmItem(&result, alarm.PostureDetection)
			disableAlarmItem(&result, alarm.LeftBed)
		}
	} else {
		disableAlarmItem(&result, alarm.Fall)
		disableAlarmItem(&result, alarm.SittingOnGround)
		disableAlarmItem(&result, alarm.PostureDetection)
		disableAlarmItem(&result, alarm.LeftBed)
	}

	hrbrParam, hrbrParamExists := deviceProps["heart_breath_param"]
	if hrbrParamExists && hrbrParam != nil {
		if hrbrParamStr, ok := hrbrParam.(string); ok && hrbrParamStr != "" {
			decodedParams, err := DecodeHeartBreathParam(hrbrParamStr)
			if err == nil && decodedParams != nil {
				for key, value := range decodedParams {
					parts := splitAlarmItemKey(key)
					if len(parts) == 2 {
						updateAlarmItemParam(&result, parts[0], parts[1], value)
					}
				}
			}
		} else {
			disableAlarmItem(&result, alarm.HeartRateAlert)
			disableAlarmItem(&result, alarm.RespRateAlert)
			disableAlarmItem(&result, alarm.WeakBiometricSignal)
		}
	} else {
		disableAlarmItem(&result, alarm.HeartRateAlert)
		disableAlarmItem(&result, alarm.RespRateAlert)
		disableAlarmItem(&result, alarm.WeakBiometricSignal)
	}

	return result, nil
}

func updateAlarmItemParam(items *[]alarm.AlarmItem, alarmType string, paramKey string, paramValue interface{}) {
	for i := range *items {
		if (*items)[i].AlarmType == alarmType {
			if (*items)[i].AlarmParams == nil {
				(*items)[i].AlarmParams = make(map[string]interface{})
			}
			(*items)[i].AlarmParams[paramKey] = paramValue
			break
		}
	}
}

func updateAlarmItemEnabled(items *[]alarm.AlarmItem, alarmType string, enabled int) {
	for i := range *items {
		if (*items)[i].AlarmType == alarmType {
			(*items)[i].IsEnabled = &enabled
			break
		}
	}
}

func disableAlarmItem(items *[]alarm.AlarmItem, alarmType string) {
	disabled := 0
	for i := range *items {
		if (*items)[i].AlarmType == alarmType {
			(*items)[i].IsEnabled = &disabled
			break
		}
	}
}

func splitAlarmItemKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}
