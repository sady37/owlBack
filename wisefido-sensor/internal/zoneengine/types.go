// Package zoneengine 是统一空间占用状态引擎。
//
// 设计原则：
//   - 一个 engine 一套核心逻辑；ZoneType 只是 parameter，不是分支。
//   - 内部模型用统一 ZoneState；外部 Redis 投影（BedState/RoomState）
//     由 RedisAdapter 翻译输出（zone-type-specific extras 由 satellite supervisor 维护）。
//     bathroom 不再有独立 state，翻译为 RoomState 带 Kind=bathroom。
//   - 所有阈值/权重外置 yaml（zone_rules.yaml），支持 hot reload + spatial_config 覆盖。
//   - 负反馈源 MVP 只有 2 个：self_contradiction（自相矛盾秒级 rollback）+
//     subset_invariant（bed⊆room 跨 zone reconcile）。cell_history / operator_feedback /
//     behavioral_baseline 三类是后期。
//
// 与 owl-common/card.CardStatus 关系：
//   - card.BedState/RoomState 是给前端的 JSON contract。
//   - zoneengine.ZoneState 是引擎内部模型，唯一真相源。
//   - RedisAdapter 写 card.* 时做翻译；翻译丢失的字段（SleepStage / StaySec 等）
//     由独立的 satellite supervisor 在收到 ZoneEvent 后写入。
package zoneengine

// ZoneType 区域类型，决定 yaml 配置族 + Redis 投影字段族 + 不变量规则。
type ZoneType int

const (
	ZoneTypeBed ZoneType = iota
	ZoneTypeRoom
	ZoneTypeBathroom
)

func (z ZoneType) String() string {
	switch z {
	case ZoneTypeBed:
		return "bed"
	case ZoneTypeRoom:
		return "room"
	case ZoneTypeBathroom:
		return "bathroom"
	default:
		return "unknown"
	}
}

// ZoneStatus 三态状态机（老人友好设计）。
//
//	Vacant   ─── enter ≥ +threshold ──► Occupied
//	                                      │
//	                                      │ leave ≤ -threshold
//	                                      ▼
//	   ◀──── timer expires ──── Leaving
//	                                      │
//	                                      │ enter ≥ +threshold (老人坐回)
//	                                      ▼
//	                                  Occupied
//
// Leaving 是 Occupied 与 Vacant 之间的"软离开"中间态。老人坐起来 / 站起来 / 试探下床
// 这些行为短时间内 sleepad 压感消失，但人仍在床区或附近。Leaving 给出 N 秒确认窗，
// 窗内回弹则取消离开。subset_invariant 视 Leaving 仍属"在区域内"（IsPresent=true）。
type ZoneStatus int

const (
	StatusVacant   ZoneStatus = iota // 无人
	StatusOccupied                   // 已确认在
	StatusLeaving                    // 正在离开（计时器跑中，未确认）
)

func (s ZoneStatus) String() string {
	switch s {
	case StatusVacant:
		return "vacant"
	case StatusOccupied:
		return "occupied"
	case StatusLeaving:
		return "leaving"
	default:
		return "unknown"
	}
}

// Transition 翻转沿语义（对应 ZoneEvent.Transition 字段）。
const (
	TransitionOccupied    = "occupied"     // Vacant → Occupied
	TransitionVacant      = "vacant"       // Leaving → Vacant (timer 超时)
	TransitionLeaving     = "leaving"      // Occupied → Leaving (老人坐起 / 离床事件触发)
	TransitionReturned    = "returned"     // Leaving → Occupied (老人坐回)
	TransitionCountChange = "count_change" // NumberPeople 直写
)

// ZoneState 单 zone 的瞬时状态（engine 内部统一模型）。
//
// 各 ZoneType 用同一结构。zone-type-specific 扩展字段（Bed.SleepStage /
// Bathroom.StayFSMPhase / Room.AreaPeople 等）由消费 ZoneEvent 的 satellite
// supervisor 独立维护，不入 engine —— engine 只关心"是否被占用 + 多少人 +
// 何时翻转"，把派生属性留给观察者模式。
type ZoneState struct {
	ZoneType    ZoneType   `json:"zone_type"`
	ZoneID      string     `json:"zone_id"` // CIDR text，e.g. "fd00:0:3:111:3:101::/96" (bed) / /88 (room)
	Status      ZoneStatus `json:"status"`           // Vacant / Occupied / Leaving 三态
	Occupied    bool       `json:"occupied"`         // 兼容字段：== (Status != Vacant)，下游消费者保留无感
	Count       int        `json:"count"`            // Room: total_people；Bed: 0/1
	Score       int        `json:"score"`            // 累积 confidence，正向 occupied / 负向 vacant
	SinceTs     int64      `json:"since_ts"`         // 进入当前状态的时间戳 ms
	LastEnterTs int64      `json:"last_enter_ts,omitempty"`
	LastExitTs  int64      `json:"last_exit_ts,omitempty"`
	LeavingSince int64     `json:"leaving_since,omitempty"` // Status==Leaving 时记录进入时间，timer 用
	LastSource  string     `json:"last_source,omitempty"`   // 最近翻转的决定信号源
	UpdatedAt   int64      `json:"updated_at"`

	// AloneContinuousTs — engine 内部独居锚点（Count==1 起点 ms）；Count 变到 1 时 set，
	// Count!=1 时 0。translator/publisher 读此 anchor 派生 RoomState.AloneContinuousMin。
	// 不入 RoomState（FE 不需 ts）；engine 内部状态。
	AloneContinuousTs int64 `json:"alone_continuous_ts,omitempty"`

	// Bayesian debug 字段（仅 bed zone + useBedBayesian=true 时填）。
	// 不入 FE bedstate；仅供 zoneengine 运维观测 / log / 测试断言。
	BayesianLogOdds float64 `json:"bayesian_log_odds,omitempty"`
	BayesianProb    float64 `json:"bayesian_prob,omitempty"`
	BayesianGamma   float64 `json:"bayesian_gamma,omitempty"`
}

// IsPresent zone 是否被占用（Vacant 之外都算）。
// subset_invariant 和 frontend 主要逻辑用这个；Leaving 算"还有人在"。
func (s ZoneState) IsPresent() bool { return s.Status != StatusVacant }

// SignalEvidence 引擎输入：one signal contributes confidence delta to one zone.
//
// Source / Kind 是审计字段，决定优先级（latch）和负反馈降权对象。Delta 由
// Scorer 按 yaml 规则计算后填入；如果 Source 是 NumberPeople 这类直写信号，
// Delta 为 0、Count 字段携带新值。
//
// sensor 内部只用 (ZoneType, ZoneID) 物理寻址；ZoneID = CIDR 文本，下游想判
// "同 bed/room/unit" 用 netip.Prefix.Contains() 做子网包含运算，不依赖 card 概念。
type SignalEvidence struct {
	ZoneType    ZoneType               `json:"zone_type"`
	ZoneID      string                 `json:"zone_id"`
	Source      string                 `json:"source"` // "sleepace" / "radar" / "polygon" / "hr_rr" / "number_people" ...
	Kind        string                 `json:"kind"`   // "enter" / "leave" / "sustain" / "count_change"
	Delta       int                    `json:"delta"`  // 签名 confidence delta；+ 推 occupied / - 推 vacant
	Count       int                    `json:"count,omitempty"` // 显式 count override（NumberPeople 走这条）
	LatchSec    int                    `json:"latch_sec,omitempty"`
	Ts          int64                  `json:"ts"`
	TriggerData map[string]interface{} `json:"trigger,omitempty"` // 供 trace / replay
}

// ZoneEvent 引擎输出：zone 状态翻转沿（下游派生消费 + RedisAdapter 投影）。
//
// 注意 Transition="count_change" 时 Occupied 可能不变（同房间 1→2 人不算新 enter）。
// Trace 保留触发本次翻转的最近 N 条 SignalEvidence，供 audit + replay。
type ZoneEvent struct {
	ZoneType   ZoneType         `json:"zone_type"`
	ZoneID     string           `json:"zone_id"`
	Transition string           `json:"transition"` // "occupied" / "vacant" / "count_change"
	NewState   ZoneState        `json:"new"`
	PrevState  ZoneState        `json:"prev"`
	Trace      []SignalEvidence `json:"trace,omitempty"`
	Ts         int64            `json:"ts"`
}

// FeedbackEvent 负反馈输入（self_contradiction + subset_invariant 触发）。
//
// MVP 两类：
//
//	self_contradiction:  zone 翻转后短窗内信号自打脸 → rollback prev state +
//	                     Affected 信号源临时降权 Penalty (持续 DurationMin)
//	subset_invariant:    bed⊆room 关系违反 → reconcile（详 validator.go）
//
// 后期 cell_history / operator_feedback / behavioral_baseline 走相同结构。
type FeedbackEvent struct {
	ZoneType    ZoneType `json:"zone_type"`
	ZoneID      string   `json:"zone_id"`
	Reason      string   `json:"reason"` // "self_contradiction" / "subset_invariant"
	Affected    []string `json:"affected,omitempty"`     // 受影响的信号源
	Penalty     int      `json:"penalty,omitempty"`      // confidence 降权
	DurationMin int      `json:"duration_min,omitempty"` // 降权持续时长
	Ts          int64    `json:"ts"`
}

// StateKey engine 内部状态表的主键。物理寻址：(zone_type, zone_id) 唯一对应一个 ZoneState。
// 不引入 card 概念 —— sensor 只处理物理层；同一 /96 床的多源 sleepad+radar 喂同一 FSM，
// Scorer 合并 evidence 自然多源融合。card 是 cardagg/FE 域，下游按 device_addr LPM 投影。
type StateKey struct {
	ZoneType ZoneType
	ZoneID   string
}

// ZoneEventListener 翻转事件回调（下游派生消费 + Redis 投影 adapter 走这条）。
type ZoneEventListener interface {
	OnZoneEvent(e ZoneEvent)
}

// FeedbackListener 负反馈事件回调（监控 / Redis 存储 / 后期 ML 训练样本走这条）。
type FeedbackListener interface {
	OnFeedback(e FeedbackEvent)
}
