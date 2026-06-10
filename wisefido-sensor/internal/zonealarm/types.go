// Package zonealarm — state-anchored zone-derived alarm 引擎。
//
// 不订阅 ZoneEvent 流；Supervisor.Tick 周期从 sensor StreamPublisher 拿 RoomState/BedState
// 快照，按规则比阈值 → fire（fire-once-until-anchor-invalidates）。
//
// 跟 zoneengine 的关系：
//   - zonealarm 是 sensor 派生 state 的纯快照消费者，**不修改** zone state
//   - 不实现 ZoneEventListener 接口（旧 OnZoneEvent 路径已废）
//   - fire 唯一出口 = AlarmFirer.Fire（wiring 包装 sensor consumer.AlarmBackChannel）
//
// 三条 rule 同 shape（statee-anchored）：
//   1. Stay         — bathroom alone 锚 AloneContinuousMin，day 45min / night 30min；
//                     KeepCounting=[bathroom.Alone] → multi/vacant 即清 fired
//   2. LeftBed      — bed 锚 BedStatusTs（NotInBed 时），30min；
//                     KeepCounting=[bed.Vacant, room.Vacant] → 床或房任一占用即清
//   3. NightAbsence — room 锚 LastExitTs，30min + 21-7 时段；
//                     KeepCounting=[room.Vacant, bed.Vacant] → 房或床任一占用即清
package zonealarm

import (
	"context"
	"fmt"
	"time"
)

// EntityKind 规则锚定 / KeepCounting 条件指向哪类 entity。
//
//	EntityBed       /96 床 — sensor StreamPublisher.SnapshotBedStates
//	EntityRoom      /88 普通 room — SnapshotRoomStates Kind=RoomTypeDefault
//	EntityBathroom  /88 卫生间 — SnapshotRoomStates Kind=RoomTypeBathroom
type EntityKind int

const (
	EntityBed EntityKind = iota
	EntityRoom
	EntityBathroom
)

func (k EntityKind) String() string {
	switch k {
	case EntityBed:
		return "bed"
	case EntityRoom:
		return "room"
	case EntityBathroom:
		return "bathroom"
	default:
		return fmt.Sprintf("EntityKind(%d)", int(k))
	}
}

// EntityState 规则 KeepCounting 条件描述 entity 必须处于的状态。
//
//	StateVacant     room.TotalPeople==0 / bed.BedStatus==NotInBed
//	StateOccupied   inverse of Vacant
//	StateAlone      room.TotalPeople==1（仅 room/bathroom 适用）
type EntityState int

const (
	StateVacant EntityState = iota
	StateOccupied
	StateAlone
)

func (s EntityState) String() string {
	switch s {
	case StateVacant:
		return "vacant"
	case StateOccupied:
		return "occupied"
	case StateAlone:
		return "alone"
	default:
		return fmt.Sprintf("EntityState(%d)", int(s))
	}
}

// AnchorField 规则计时起点。两种语义：
//
//	AnchorAloneContinuousMin  RoomState.AloneContinuousMin (Stay) — MIN 整数，sensor 算好 push
//	AnchorLastExitTs          RoomState.LastExitTs (NightAbsence) — TS，evaluator 现算 elapsed
//	AnchorBedStatusTs         BedState.BedStatusTs（NotInBed 起点）— TS，evaluator 现算 elapsed
//
// evaluator 端 readElapsedSec 统一返 elapsed 秒数（MIN 类乘 60），threshold 仍以 sec 比。
const (
	AnchorAloneContinuousMin = "AloneContinuousMin"
	AnchorLastExitTs         = "LastExitTs"
	AnchorBedStatusTs        = "BedStatusTs"
)

// Condition KeepCounting 单条条件（全 true 才视 timer 有效）。
type Condition struct {
	EntityStr string `yaml:"entity"` // "bed" / "room" / "bathroom"；空 = Self（同 rule.Anchor）
	StateStr  string `yaml:"state"`  // "vacant" / "occupied" / "alone"

	// 解析后字段（LoadFromFile / DefaultRules resolveRule 填）
	Entity    EntityKind  `yaml:"-"`
	State     EntityState `yaml:"-"`
	UsesSelf  bool        `yaml:"-"` // EntityStr 空时 true（运行时用 rule.Anchor）
}

// Rule 一条 state-anchored alarm 规则。
type Rule struct {
	// AlarmType cardagg alarm_handler 接收 event_name；取自 owl-common/alarm 常量。
	AlarmType string `yaml:"alarm_type"`

	// Anchor 锚定 entity 类型 — 决定 Tick 时遍历哪个 snapshot map。
	AnchorStr string     `yaml:"anchor"` // "bed" / "room" / "bathroom"
	Anchor    EntityKind `yaml:"-"`

	// AnchorField 从 anchor entity 上读哪个 ts 字段当计时起点（const AnchorXxx）。
	AnchorField string `yaml:"anchor_field"`

	// KeepCounting 计时有效条件：所有 condition 全 true 才计；任一 false 即清 fired。
	KeepCounting []Condition `yaml:"keep_counting"`

	// ThresholdDaySec 白天阈值（秒，超过即 fire）。必填。
	ThresholdDaySec int `yaml:"threshold_day_sec"`

	// ThresholdNightSec 夜间（card.IsRiskTime 21-08）阈值；0 表示沿用 Day。
	ThresholdNightSec int `yaml:"threshold_night_sec,omitempty"`

	// NightOnly 仅本地时段内允许 fire（NightAbsence 用 21-7）。nil = 全天。
	NightOnly *TimeWindow `yaml:"night_only,omitempty"`
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

// FireKey fire-once gate map 主键 — 同 zone 内同 AlarmType 同时只允许一条 fire。
// 物理寻址：ZoneID = CIDR 文本（/96 床 / /88 房）。
type FireKey struct {
	ZoneID    string
	AlarmType string
}

// FireEvent fire 出口的载荷 — evaluator 触发 fire 时填，传给 AlarmFirer.Fire。
type FireEvent struct {
	Key       FireKey
	AnchorTs  int64 // 锚点 ts ms（计时起点）
	NowMs     int64 // fire 时刻 ms
	Snapshot  any   // 引用快照（*card.RoomState / *card.BedState；wiring 端按需做 device 路由）
}

// AlarmFirer zonealarm fire 出口接口；wiring 层包装 alarm_back_channel 实现。
//
// state-anchored 重构后 Arm/Cancel 概念消失（无 pending FSM），仅保 Fire。
type AlarmFirer interface {
	Fire(ctx context.Context, e FireEvent) error
}

// NopFirer 测试 / 兜底用。
type NopFirer struct{}

func (NopFirer) Fire(context.Context, FireEvent) error { return nil }

// DeviceLookup zone CIDR → device /128 canonical IPv6 字符串。
// wiring 层包装 BedDeviceLookup 实现；nil 时退化 zone CIDR host=0。
type DeviceLookup interface {
	// FindPrimaryDevice 与 AlarmFirer 同语义（per [[track_fusion_and_gate_cardid]]）：
	// LeftBed→Radar / Stay→Bathroom radar / NightAbsence→Room radar。
	// 返空字符串 = 找不到（用 zone CIDR fallback）。
	FindPrimaryDevice(zoneID, alarmType string) string
}

// ThresholdResolver per-device alarm 阈值 + 使能查询接口。
//
// wiring 层包装 service.AlarmEnablementCache + DeviceLookup 实现。
//
// Resolve 返回：
//   - sec       生效的阈值（秒）；spatial_config 有则用之，否则 fallback
//   - enabled   该 alarm 在 device 级是否启用；false → Evaluator 应清 fired 不 fire
//
// 实现端从 AlarmEnablementCache.IsEnabled 拿 EnabledAlarm.AlarmParams 解析
// duration_sec / duration_min*60 任意一个；都没有就回 fallback。
type ThresholdResolver interface {
	Resolve(ctx context.Context, deviceAddr, alarmType string, fallbackSec int) (sec int, enabled bool)
}

// NopResolver 测试 / 未注入时用：永远返 (fallback, true) — 即按 yaml 默认全启用。
type NopResolver struct{}

func (NopResolver) Resolve(_ context.Context, _, _ string, fallbackSec int) (int, bool) {
	return fallbackSec, true
}
