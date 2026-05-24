// Package zonealarm — zone state 派生 alarm 引擎。
//
// 订阅 zoneengine.Engine 的 ZoneEvent 流，按 yaml 规则集 arm pending alarm；timer
// 到期 fire；中途收到 cancel trigger 即 cancel pending。
//
// 跟 zoneengine 的关系：
//   - zonealarm 是 zoneengine 的下游消费者，**不修改** zone state
//   - 通过 ZoneEventListener 接口反向注册到 Engine，跟 RedisAdapter 平级
//   - 不直接 publish iot:alarm:stream — 通过 AlarmFirer 接口由 wiring 层注入
//     （wiring 包装 sensor consumer.AlarmBackChannel + 解析 device_addr）
//
// 当前规则集（4 条；硬约束在 [[zoneengine_adapters_done]] memory）：
//
//   1. Stay              — bathroom 持续 N min →  fire；cancel: bathroom→vacant /
//                          其它 room.EnterRoom / bed.InBed
//   2. LeftBed           — bed→vacant 持续 N min → fire；cancel: bed.InBed /
//                          room.EnterRoom（人回房算回床区域，用户拍板）
//   3. NightAbsence      — room→vacant 持续 N min（21-7 时段） → fire；cancel:
//                          room.EnterRoom / bed.InBed / room.NumberPeople>0
//                          （bed-scope NightAbsence 已删除：床上无人不等于风险，可用 LeftBed 长时长替代）
package zonealarm

import (
	"context"
	"time"

	"wisefido-sensor/internal/zoneengine"
)

// Rule 一条 alarm 派生规则。同一规则可以匹配多张卡，每张卡独立 pending。
type Rule struct {
	// AlarmType cardagg alarm_handler 接收时识别的 event_name；
	// 取自 owl-common/alarm 常量（Stay / LeftBed / NightAbsence）。
	AlarmType string `yaml:"alarm_type"`

	// Level cardagg PersistAlarmAndPublish 用的 alarm_level（"WARN" / "CRIT" 等）。
	Level string `yaml:"level"`

	// ArmZone / ArmStatus arm 触发条件：当指定 ZoneType 翻转到指定 status 时启 pending。
	ArmZone   zoneengine.ZoneType   `yaml:"-"` // yaml 字符串通过 ArmZoneStr 转
	ArmStatus zoneengine.ZoneStatus `yaml:"-"`

	ArmZoneStr   string `yaml:"arm_zone"`   // "bed" / "room" / "bathroom"
	ArmStatusStr string `yaml:"arm_status"` // "occupied" / "vacant" / "leaving"

	// Duration arm 后多久 fire（秒）。0 = 立即 fire（不走 pending）。
	DurationSec int `yaml:"duration_sec"`

	// Cancels arm 后哪些 ZoneEvent 取消本条 pending（任一命中即 cancel）。
	Cancels []CancelTrigger `yaml:"cancels"`

	// TimeWindow 仅在指定本地时段内允许 arm（cancel 不受窗口限制）。nil = 24h 任何时间。
	TimeWindow *TimeWindow `yaml:"time_window,omitempty"`
}

// CancelTrigger arm 中 pending 的取消条件。
//
//	Zone / Status — 哪种 ZoneEvent 触发 cancel
//	CountGtZero   — true 时仅当 NewState.Count > 0 才命中（Night Out-of-Room 用 NumberPeople 确认）
type CancelTrigger struct {
	ZoneStr     string                `yaml:"zone"`           // "bed" / "room" / "bathroom"
	StatusStr   string                `yaml:"status"`         // "occupied" / "vacant" / "leaving"; 空 = 任意
	CountGtZero bool                  `yaml:"count_gt_zero,omitempty"`
	Zone        zoneengine.ZoneType   `yaml:"-"`
	Status      zoneengine.ZoneStatus `yaml:"-"`
	StatusAny   bool                  `yaml:"-"` // StatusStr 空时为 true
}

// TimeWindow 本地时段（HH:MM）。支持跨午夜（如 21:00-07:00）。
type TimeWindow struct {
	StartH int `yaml:"start_h"`
	StartM int `yaml:"start_m"`
	EndH   int `yaml:"end_h"`
	EndM   int `yaml:"end_m"`
}

// Active 时间 t（local）是否在窗口内。支持跨午夜。
func (w *TimeWindow) Active(t time.Time) bool {
	if w == nil {
		return true
	}
	cur := t.Hour()*60 + t.Minute()
	start := w.StartH*60 + w.StartM
	end := w.EndH*60 + w.EndM
	if start <= end {
		return cur >= start && cur < end
	}
	// 跨午夜
	return cur >= start || cur < end
}

// Pending arm 中的 alarm 实例（per zone + alarmType）。
type Pending struct {
	Key       PendingKey
	ArmedAt   int64       // ms
	DueAt     int64       // ms
	Level     string
	Trigger   zoneengine.ZoneEvent // 触发 arm 的 event；fire 时作 trigger_data
}

// PendingKey 用于 supervisor 内部 map 主键 — 一个 spatial zone 一种 alarmType 同时只能有一个 pending。
// 物理寻址：ZoneID = CIDR 文本（/96 床 / /88 房）；与 card 概念解耦（card 是 cardagg/FE 域）。
type PendingKey struct {
	ZoneID    string
	AlarmType string
}

// AlarmFirer zonealarm 的 fire 出口接口；wiring 层包装 alarm_back_channel 实现。
//
// **Arm / Cancel 是 sensor 内部 lifecycle 钩子**（仅 log / metric / 审计用），不外发 cardagg；
// 真正的 sensor→cardagg 路径只有 Fire（PublishAlarmFire confirmed alarm）。
// PR1 时代的 pending_arm / pending_cancel sentinel 已删 — pending state v2 完全在 Supervisor
// 内部 maintain，外部不可见。
type AlarmFirer interface {
	Arm(ctx context.Context, p Pending) error
	Cancel(ctx context.Context, key PendingKey, reason zoneengine.ZoneEvent) error
	Fire(ctx context.Context, p Pending) error
}

// nopFirer 测试 / 兜底用。
type NopFirer struct{}

func (NopFirer) Arm(context.Context, Pending) error                                 { return nil }
func (NopFirer) Cancel(context.Context, PendingKey, zoneengine.ZoneEvent) error      { return nil }
func (NopFirer) Fire(context.Context, Pending) error                                 { return nil }
