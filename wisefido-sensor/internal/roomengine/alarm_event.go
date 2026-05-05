package roomengine

// alarm_event.go
//
// 解析 iot:alarm:stream + iot:event:stream 中的 radar 来源消息。
//
// 设计哲学（与用户对齐）：
//  - 当前阶段只做"触发器 + 落账"，不做 verify / 否决
//  - alarm/event 到达 → engine 立即调 tm.Tick(ts) → 复用现有段 4/5/6 用 alarm 时间戳跑一次
//  - TrackManager 把记录存进 recentRadarAlarms / recentRadarEvents，未来段 7 (fall verify) 消费
//
// 不做：PendingRadarFall / 否决 / 5 条验证规则。这些留给独立 future 任务。

import (
	"encoding/json"
)

// RadarFallAlarm radar 固件直接发的 Fall 报警（来自 iot:alarm:stream）。
//
// 数据样例（从 iot_timeseries 实测）：
//
//	{"event_name": "Fall", "event_since": 1777180187386,
//	 "dataCategory": "Fall", "event_status": "start",
//	 "event_payload": "{\"event_type\":2,\"pose\":2,\"track_id\":0}"}
//
// 注意：event_payload 是嵌套 JSON 字符串，需要二次解析才能拿到 track_id / pose。
type RadarFallAlarm struct {
	DeviceUID string
	TMs       int64

	TrackID int    // event_payload.track_id
	Pose    int    // event_payload.pose（雷达 pose 编码）
	EventType int  // event_payload.event_type（固件 fall 子类型）
	Status    string // "start" / "end"
}

// RadarTrackEvent radar 固件发的 EnterRoom/ExitRoom/InBed/LeftBed/NumberPeople（来自 iot:event:stream）。
//
// 数据样例：
//
//	{"track_id": 11, "event_name": "EnterRoom", "event_status": "instant",
//	 "event_since": 1777179605010, "track_count": 1}
//
//	{"track_id": 10, "event_name": "number_people", "event_status": "start",
//	 "event_since": 1777703748870, "number_people": 0}
//
// 用途：硬证据"人数变化"或"床压状态"，未来用于交叉验证 sleepad / 校准 segment 1 realCount 推断。
// NumberPeople=0 也作为 ExitRoom 兜底（部分 firmware 在 FOV 边缘离场不发 ExitRoom）。
type RadarTrackEvent struct {
	DeviceUID    string
	TMs          int64
	EventName    string // "EnterRoom" / "ExitRoom" / "InBed" / "LeftBed" / "NumberPeople"
	TrackID      int
	Status       string // "instant" / "start" / "end"
	NumberPeople int    // 仅 EventName=="NumberPeople" 时有效
}

// ParseRadarFallAlarm 解析 alarm:stream 一条消息的 data_value 为 RadarFallAlarm 列表。
// 仅当 envelopeCat="Fall" 才返回（envelope.Category 是事件类型唯一权威）。
func ParseRadarFallAlarm(dv interface{}, deviceUID, envelopeCat string, fallbackTs int64) []RadarFallAlarm {
	if envelopeCat != "Fall" {
		return nil
	}
	arr := jsonArrayOfObjects(dv)
	if len(arr) == 0 {
		return nil
	}
	out := make([]RadarFallAlarm, 0, len(arr))
	for _, m := range arr {
		st, _ := m["event_status"].(string)
		ts := int64FromAny(m["event_since"])
		if ts == 0 {
			ts = fallbackTs
		}
		alarm := RadarFallAlarm{
			DeviceUID: deviceUID,
			TMs:       ts,
			Status:    st,
		}
		// event_payload 是嵌套 JSON 字符串，二次解析
		if pl, ok := m["event_payload"].(string); ok && pl != "" {
			var p map[string]interface{}
			if err := json.Unmarshal([]byte(pl), &p); err == nil {
				alarm.TrackID = jsonInt(p["track_id"])
				alarm.Pose = jsonInt(p["pose"])
				alarm.EventType = jsonInt(p["event_type"])
			}
		}
		out = append(out, alarm)
	}
	return out
}

// ParseRadarTrackEvents 解析 event:stream 一条消息的 data_value 为 RadarTrackEvent 列表。
// 仅当 envelopeCat ∈ {EnterRoom, ExitRoom, InBed, LeftBed} 才返回（envelope.Category 是事件类型唯一权威）。
//
// 实测（2026-04-27 case_lostfall）：
//   - EnterRoom / ExitRoom / InBed 只发 status="start"
//   - LeftBed 双发 ("instant" + "start")
//
// 这里接受 "start" 与 "instant"。LeftBed 双发时同 TMs，RecordRadarEvent 用 TMs 做 map key
// 会自然覆盖（无重复计数风险）。"end" 状态（如 SignalPoorRecover end）不在白名单内会被 envelopeCat 过滤。
func ParseRadarTrackEvents(dv interface{}, deviceUID, envelopeCat string, fallbackTs int64) []RadarTrackEvent {
	// envelopeCat 是 iot:event:stream envelope 字段，与 firmware 内部 event_name 不完全同名。
	// number_people（小写下划线）是 envelope 形式，统一规整成 EventName="NumberPeople" 便于内部判断。
	canonical := envelopeCat
	switch envelopeCat {
	case "EnterRoom", "ExitRoom", "InBed", "LeftBed":
	case "number_people", "NumberPeople":
		canonical = "NumberPeople"
	default:
		return nil
	}
	arr := jsonArrayOfObjects(dv)
	if len(arr) == 0 {
		return nil
	}
	out := make([]RadarTrackEvent, 0, len(arr))
	for _, m := range arr {
		st, _ := m["event_status"].(string)
		if st != "start" && st != "instant" {
			continue
		}
		ts := int64FromAny(m["event_since"])
		if ts == 0 {
			ts = fallbackTs
		}
		evt := RadarTrackEvent{
			DeviceUID: deviceUID,
			TMs:       ts,
			EventName: canonical,
			TrackID:   jsonInt(m["track_id"]),
			Status:    st,
		}
		if canonical == "NumberPeople" {
			evt.NumberPeople = jsonInt(m["number_people"])
		}
		out = append(out, evt)
	}
	return out
}

// jsonArrayOfObjects 把 dv（可能是 []interface{} / []map / string JSON）统一转为 []map[string]interface{}
func jsonArrayOfObjects(dv interface{}) []map[string]interface{} {
	if dv == nil {
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
	return arr
}
