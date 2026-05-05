package roomengine

// sensor_fusion.go
//
// 多传感器融合的扩展点：roomengine 接收非雷达传感器观测（Sleepad / 未来 Geophone 等），
// 与雷达 track 数据交叉验证，提升异常判定准确度。
//
// 当前已实现：
//   - SleepadObservation：床压传感器（在床/离床 + 呼吸/心率）
//   - 用途：silent fall 触发时，若同房间 sleepad 确认"在床有生命体征"，
//     说明雷达消失只是失锁，不报警
//
// 未来扩展模板：
//   1. 定义 XxxObservation struct（包含 DeviceID + TMs + 关键状态字段）
//   2. TrackManager 加 xxxStates map[deviceID]*XxxObservation
//   3. 加 ProcessXxxObservation(obs) 方法（per-device 保留最新）
//   4. Engine.handleMessage 按 device_type 路由
//   5. 异常判定（silent fall / long still / 等）查询相关传感器状态做 short-circuit

import (
	"encoding/json"
	"strconv"
)

// SleepadObservation 一帧 sleepad 观测（与 radar TrackFrame 平行）
// 字段对应 iot_timeseries.data_value 中 sleepad 的 JSON
//
// 字段语义说明（实测数据 2026-04-25）：
//   - bed_status: **0 = 在床（未脱离），1 = 离床（detached）**
//     注意极性反直觉！通过 monitor 帧 bed_status==0 + heart_rate>0 验证。
//   - event_name=InBed → bed_status=0；event_name=LeftBed → bed_status=1
//   - 故 InBed = (bed_status == 0)
type SleepadObservation struct {
	DeviceID        string
	DeviceUID       string
	TMs             int64
	InBed           bool // bed_status == 0
	HeartRate       int  // 0 = no signal
	RespiratoryRate int  // 0 = no signal
	BodyMove        int  // 0/1
	TurnOver        int  // 0/1
}

// HasVitalSign 任一生命体征非零（可信存活信号）
func (o *SleepadObservation) HasVitalSign() bool {
	return o.HeartRate > 0 || o.RespiratoryRate > 0
}

// ParseSleepadObservations 把 sleepad iot:monitor:stream 的 data_value 解析为多帧观测
// data_value 是 array of object，每个 object 一条记录（同一设备一帧通常只 1 条 track_id=0）
func ParseSleepadObservations(dv interface{}, deviceID, deviceUID string, fallbackTs int64) []SleepadObservation {
	if dv == nil {
		return nil
	}
	// 兼容 data_value 直接是 array 或 string-encoded JSON
	var arr []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, it := range v {
			if m, ok := it.(map[string]interface{}); ok {
				arr = append(arr, m)
			}
		}
	case []map[string]interface{}:
		arr = v
	case string:
		_ = json.Unmarshal([]byte(v), &arr)
	}
	if len(arr) == 0 {
		return nil
	}

	out := make([]SleepadObservation, 0, len(arr))
	for _, m := range arr {
		obs := SleepadObservation{
			DeviceID:  deviceID,
			DeviceUID: deviceUID,
			TMs:       fallbackTs,
		}
		if v, ok := m["bed_status"]; ok {
			// 注意：0 = 在床（未 detached），1 = 离床
			if jsonInt(v) == 0 {
				obs.InBed = true
			}
		}
		if v, ok := m["heart_rate"]; ok {
			obs.HeartRate = jsonInt(v)
		}
		if v, ok := m["respiratory_rate"]; ok {
			obs.RespiratoryRate = jsonInt(v)
		}
		if v, ok := m["body_move"]; ok {
			obs.BodyMove = jsonInt(v)
		}
		if v, ok := m["turn_over"]; ok {
			obs.TurnOver = jsonInt(v)
		}
		out = append(out, obs)
	}
	return out
}

// SleepadBedEvent sleepad 床压事件（来自 iot:event:stream），用于人数计数。
// 实测发现一次状态变化会发 2 条事件（status=instant + status=start），需要用 status 去重。
type SleepadBedEvent struct {
	DeviceUID string
	TMs       int64
	IsInBed   bool // event_name="InBed" → true；"LeftBed" → false
	Status    string // "instant" / "start"（去重用，只取 instant）
}

// ParseSleepadBedEvents 把 sleepad event 流的 data_value 解析为床压事件。
// 仅返回 envelopeCat ∈ {InBed, LeftBed} 且 status=instant 的事件（避免重复计数）。
// envelopeCat 来自 IoTStreamMessage.Category（事件类型唯一权威）。
func ParseSleepadBedEvents(dv interface{}, deviceUID, envelopeCat string, fallbackTs int64) []SleepadBedEvent {
	if dv == nil {
		return nil
	}
	var isInBed bool
	switch envelopeCat {
	case "InBed":
		isInBed = true
	case "LeftBed":
		isInBed = false
	default:
		return nil
	}
	var arr []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, it := range v {
			if m, ok := it.(map[string]interface{}); ok {
				arr = append(arr, m)
			}
		}
	case []map[string]interface{}:
		arr = v
	case string:
		_ = json.Unmarshal([]byte(v), &arr)
	}
	if len(arr) == 0 {
		return nil
	}
	out := make([]SleepadBedEvent, 0, len(arr))
	for _, m := range arr {
		st, _ := m["event_status"].(string)
		if st != "instant" {
			continue // 只用 instant 事件，避免 instant+start 重复
		}
		out = append(out, SleepadBedEvent{DeviceUID: deviceUID, TMs: fallbackTs, IsInBed: isInBed, Status: st})
	}
	return out
}

// jsonInt 把 JSON 数字（float64 / json.Number / int / string）安全转 int
func jsonInt(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}
