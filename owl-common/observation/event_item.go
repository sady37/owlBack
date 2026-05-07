package observation

import "encoding/json"

// EventItem 表示 iot:event:stream / iot:alarm:stream 的 dataValue 中单条事件/报警 payload。
//
// 协议契约（Plan B 后）：
//
//   - **envelope（IoTStreamMessage）承担 routing**：producer / subject_entity / category / topic_type / ...
//     事件类型唯一权威 = envelope.Category（已废弃 dataCategory / event_name）。
//
//   - **EventItem 字段分两类**：
//
//       (a) 生命周期元数据（所有事件共享）—— EventID / EventSince / EventStatus / EventEnd /
//           EventReason / EventLevel。
//
//       (b) 业务字段（按需 first-class 字段）—— gateway 按事件子类需要填，都 omitempty 可选；
//           常用跨 gateway 字段（HeartRate / RespiratoryRate）即使不测量也总在 wire 出现，
//           用 -1 sentinel 表示"未测量"（0 是合法测量值——心跳停止 / 无呼吸为告警态）。
//
//     当前已定义：
//       TrackID（雷达/sleepace 的 device-local 编号；device-class 事件可填 TrackDevice 占位）
//       Pose（雷达姿态枚举，Fall/SittingOnGround）
//       HeartRate（bpm，-1=未测量；HR/RR 类告警必填）
//       RespiratoryRate（rpm，-1=未测量；HR/RR 类告警必填）
//
//   - **wire 序列化** = EventItem 字段平铺的 map；DB 查询直接 `trigger_data->>'heart_rate'`，
//     筛真实 HR 用 `WHERE heart_rate > 0`。
//
// EventStatus 与 owl-common/alarm Registry[type].EndPolicy 配对：
//
//   - "start" / "instant" / "pending" → 触发新报警
//   - "end" → 由 cardagg AlarmHandler 入口集中按 EndPolicy 路由（auto_resolve / manual_ack / ignore）
//
// **构造方式**：publisher 应当使用 NewEventItem(eventSince, eventStatus) 构造器，
// 它把 HeartRate/RespiratoryRate 默认填 -1（未测量）；之后再按需覆盖（如 HR 告警 publisher
// 设 item.HeartRate = 120）。直接用 struct literal 会让 HR/RR 默认值是 0，被误读为合法测量值。
type EventItem struct {
	// 生命周期元数据
	EventID     string `json:"event_id,omitempty"`     // 设备侧 session id（如 sleepace alarm 151）
	EventSince  int64  `json:"event_since"`            // 必填，发生时间戳 ms
	EventStatus string `json:"event_status"`           // 必填：start / end / instant / pending
	EventEnd    int64  `json:"event_end,omitempty"`    // 解除时间戳 ms（仅 end 时填）
	EventReason string `json:"event_reason,omitempty"` // 解除原因
	EventLevel  string `json:"event_level,omitempty"`  // 严重等级（可选；默认由 alarm Registry.DefaultLevel 决定）

	// 业务字段（first-class，omitempty 按需填）
	TrackID int `json:"track_id,omitempty"` // device-local 编号；雷达 1-N / sleepace LeftRight 0-1 / 占位 TrackDevice 等
	Pose    int `json:"pose,omitempty"`     // 雷达姿态枚举（observation.EnumPose 0-12）；适用 Fall/SittingOnGround

	// 生命体征（无 omitempty——总在 wire 出现；-1=未测量，0=测得 0bpm/rpm 真值=告警态）
	HeartRate       int `json:"heart_rate"`       // bpm
	RespiratoryRate int `json:"respiratory_rate"` // rpm
}

// NewEventItem 构造 EventItem 默认值：HeartRate / RespiratoryRate = -1（未测量）。
// 所有 publisher 应通过本构造器创建 EventItem，避免直接 struct literal 让 HR/RR 默认 0
// 被误读为合法测量值（0 在临床上 = 心跳停止 / 无呼吸 = 告警态）。
//
//	item := observation.NewEventItem(ts, "start")
//	item.TrackID = trackID
//	item.HeartRate = hr  // 仅 HR 告警 publisher 显式填测量值
func NewEventItem(eventSince int64, eventStatus string) EventItem {
	return EventItem{
		EventSince:      eventSince,
		EventStatus:     eventStatus,
		HeartRate:       -1,
		RespiratoryRate: -1,
	}
}

// EventItemToDataMap 将 EventItem 序列化为 map[string]interface{}（键与 json 标签一致）。
// 业务字段直接通过 EventItem 的 first-class 字段（TrackID/Pose/HeartRate/RespiratoryRate）填，
// 平铺序列化后即在 map 顶层。仍有少量额外字段（如 BedStatus / NumberPeople）尚未升级为
// first-class 时，可在 ToDataMap 后写入 map：
//
//	item := observation.NewEventItem(ts, "start")
//	data, _ := observation.EventItemToDataMap(&item)
//	data[observation.FieldNumberPeople] = n
//
// 长期目标：常用业务字段都升级为 EventItem first-class，避免 map-write 形式。
func EventItemToDataMap(item *EventItem) (map[string]interface{}, error) {
	if item == nil {
		return nil, nil
	}
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
