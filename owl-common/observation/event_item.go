package observation

import "encoding/json"

// EventItem 是 iot:event:stream / iot:alarm:stream dataValue 中单条事件/报警 payload。
//
// **统一归一化模型（跨硬件）**：不按厂家/设备类型分结构。所有硬件（雷达 / sleepad / 未来 C…）
// 把各自原始报文归一进**同一套观测字段词汇**（与 observation.Field* 常量一致）；每种事件只填它
// 真正携带的字段，其余留 nil。加新硬件 = 映进已有字段，**零新结构**。
//
// **指针语义（业务字段一律 *T + omitempty）**：
//   - nil    = 该事件/该硬件不携带此字段 → 省略，绝不伪造（"没有别插"）。
//   - 非 nil = 携带 → 发出，**含值 0/false**（根除 int+omitempty 把合法 0 吃掉的 bug）。
//
// 判别两级：EventType = 固件粗类型（1/2/3/9/16…，item 自描述用）；envelope.Category = 细粒度语义名
// （EnterRoom/Fall/activity/SleepStage…，类型权威，下游路由依据）。两者不同粒度，非冗余。
// EventStatus 与 alarm Registry[type].EndPolicy 配对：start/instant/pending 触发；end 由 cardagg 路由。
type EventItem struct {
	// ── 生命周期信封（服务端添加；所有事件共享）──
	EventID     string `json:"event_id,omitempty"`     // 设备侧 session id
	EventType   int    `json:"event_type,omitempty"`   // 固件事件类型判别器：1进出 2姿态 3人数 5离线 7信号差 8倾角 9活动统计 16睡眠统计
	EventSince  int64  `json:"event_since"`            // 必填，发生时间戳 ms
	EventStatus string `json:"event_status"`           // 必填：start / end / instant / pending
	EventEnd    int64  `json:"event_end,omitempty"`    // 解除时间戳 ms（仅 end）
	EventReason string `json:"event_reason,omitempty"` // 解除原因
	EventLevel  string `json:"event_level,omitempty"`  // 严重等级（默认由 Registry.DefaultLevel 决定）

	// ── 归一化观测词汇（*T；填=携带，nil=未携带；A/B/C 各填子集）──

	// 目标 / 进出 / 区域
	TrackID  *int `json:"track_id,omitempty"`  // 人员/占位轨迹号（雷达 track / sleepad leftRight）；0 合法
	Event    *int `json:"event,omitempty"`     // 进出码：1进房 2离房 3进区 4离区 5进监护 6退监护
	AreaType *int `json:"area_type,omitempty"` // 区域类型：2普通床 5监护床 6感应区
	AreaID   *int `json:"area_id,omitempty"`   // 区域号（携带方填）
	Pose     *int `json:"pose,omitempty"`      // 姿态枚举（EnumPose）：5确认跌倒 8确认坐地 11床上坐起 …

	// 空间级（非 per-track）
	NumberPeople *int `json:"number_people,omitempty"` // 房内人数

	// 床 / 睡眠统计（sleep stat，type=16：归一 sleep_stage + 分钟级均值；瞬时 HR/RR 不在此，属 monitor 流）
	BedStatus          *int `json:"bed_status,omitempty"`           // 0在床 1离床
	SleepStage         *int `json:"sleep_stage,omitempty"`          // 1清醒 2浅睡 4深睡 8未知（归一自 raw 0-3）
	AvgHeartRate       *int `json:"avg_heart_rate,omitempty"`       // 分钟级均值 bpm
	AvgRespiratoryRate *int `json:"avg_respiratory_rate,omitempty"` // 分钟级均值 rpm

	// 活动统计（activity stat，type=9：分钟级聚合）
	WalkDistance        *int `json:"walk_distance,omitempty"`         // 行走距离 m
	WalkDuration        *int `json:"walk_duration,omitempty"`         // 行走时长 s
	LieDuration         *int `json:"lie_duration,omitempty"`          // 躺卧时长 s
	StandDuration       *int `json:"stand_duration,omitempty"`        // 站立时长 s
	MultiPersonDuration *int `json:"multi_person_duration,omitempty"` // 多人存在时长 s

	// 生命体征（vital 告警 alarm 流：装触发实测值；瞬时实时值在 monitor 流。设备状态 offline/signal/angle
	// 不设字段——由 envelope.Category + event_status 表达，无数值可装）
	HeartRate           *int `json:"heart_rate,omitempty"`            // bpm（0=测得0bpm=告警态，照发；nil=未携带）
	RespiratoryRate     *int `json:"respiratory_rate,omitempty"`      // rpm（同上）
	WeakBiometricSignal *int `json:"weak_biometric_signal,omitempty"` // 信号弱风险度 0-100
}

// NewEventItem 构造生命周期信封；业务字段默认 nil，由 publisher 按事件类型填指针。
//
//	item := observation.NewEventItem(ts, "start")
//	item.TrackID  = observation.IntPtr(trackID)   // 携带才填
//	item.Event    = observation.IntPtr(event)
//	item.AreaType = observation.IntPtr(areaType)
func NewEventItem(eventSince int64, eventStatus string) EventItem {
	return EventItem{
		EventSince:  eventSince,
		EventStatus: eventStatus,
	}
}

// IntPtr / BoolPtr：业务字段赋值辅助（避免散落的 v := x; &v 临时变量）。
func IntPtr(v int) *int    { return &v }
func BoolPtr(v bool) *bool { return &v }

// EventItemToDataMap 将 EventItem 序列化为 map[string]interface{}（键与 json 标签一致）。
// nil 指针字段经 omitempty 自动省略；非 nil（含 *0/*false）保留。
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
