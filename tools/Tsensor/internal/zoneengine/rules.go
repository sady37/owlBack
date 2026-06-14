package zoneengine

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Rules zoneengine 全部可调参数；从 zone_rules.yaml 加载，spatial_config 长前缀覆盖。
//
// 设计约定：
//   - 数值都是 placeholder；最终由用户 calibrate（参考 Bed 段的相对比例外推到 Room/Bathroom）。
//   - 同 zone_type 的所有 zone 实例共享一份 rules；后期可加 spatial_config tenant 覆盖。
//   - hot reload：监听 config:zone_rules:stream，调用 Engine.ReloadRules。
type Rules struct {
	Bed      ZoneRules     `yaml:"zone_bed"`
	Room     ZoneRules     `yaml:"zone_room"`
	Bathroom ZoneRules     `yaml:"zone_bathroom"`
	Feedback FeedbackRules `yaml:"negative_feedback"`
}

// ZoneRules 单 ZoneType 的评分 + 状态机参数。
type ZoneRules struct {
	// Enter 推 score 上行的事件；map key 是 SignalEvidence.Source（"sleepace" / "radar" / "polygon" 等）。
	Enter map[string]EvidenceStrength `yaml:"enter_evidence,omitempty"`

	// Sustain 维持 occupied 不单独翻转的持续证据（HR/RR 等）；
	// recent_enter_bonus 在 EnterEvent 后 window_sec 内 +bonus。
	Sustain SustainRules `yaml:"sustain_evidence,omitempty"`

	// Leave 推 score 下行的事件。
	// 当 SizeDependent=true 时，按 size_kind 走 SmallBed / LargeBed 两档（仅 Bed zone 用到）。
	Leave LeaveRules `yaml:"leave_evidence,omitempty"`

	StateMachine StateMachineRules `yaml:"state_machine"`
}

// EvidenceStrength 单源信号的强度参数。
type EvidenceStrength struct {
	Strength int `yaml:"strength"`            // confidence 绝对值；方向由 enter/leave 类别决定
	LatchSec int `yaml:"latch_sec,omitempty"` // 施加 latch 后窗内不被低强度覆盖
}

// SustainRules 持续证据规则。
type SustainRules struct {
	HRRRPresentAny struct {
		Strength int      `yaml:"strength"` // HR/RR 任意源任一在 → 这个 score
		Sources  []string `yaml:"sources"`  // 允许的源（"sleepad" / "radar"）
	} `yaml:"hr_rr_present_any,omitempty"`

	RecentEnterBonus struct {
		WindowSec int `yaml:"window_sec"` // EnterEvent 后 window 内
		Bonus     int `yaml:"bonus"`      // 额外 +bonus
	} `yaml:"recent_enter_bonus,omitempty"`
}

// LeaveRules 离开证据规则。
type LeaveRules struct {
	SizeDependent bool                        `yaml:"bed_size_dependent,omitempty"` // 仅 Bed zone：按床型分档
	Default       map[string]EvidenceStrength `yaml:"default,omitempty"`            // 非 Bed zone 或 SizeDependent=false 时用
	SmallBed      LeaveBedSizeBucket          `yaml:"small_bed,omitempty"`
	LargeBed      LeaveBedSizeBucket          `yaml:"large_bed,omitempty"`
}

// LeaveBedSizeBucket Bed zone size 相关的 leave 阈值（小床 sleepad 权威，大床降到 radar 同档）。
type LeaveBedSizeBucket struct {
	BedKinds []string                    `yaml:"bed_kinds"`
	Sources  map[string]EvidenceStrength `yaml:"sources"`
}

// StateMachineRules 翻转阈值 + 滞回参数 + Leaving 中间态时长。
type StateMachineRules struct {
	EnterThreshold int `yaml:"enter_threshold"` // score > threshold 且 prev=vacant → 翻转 occupied
	ExitThreshold  int `yaml:"exit_threshold"`  // score < -threshold 且 prev=occupied → 翻转 leaving (非 vacant)
	HysteresisSec  int `yaml:"hysteresis_sec"`  // 翻转后冷却时间，期间不响应反向信号

	// LeavingWindowSec 离床软确认窗口（老人友好设计）。
	//   exit_threshold 触发后 zone 进入 Leaving 中间态，本窗口内 score 回弹 ≥ enter_threshold
	//   即 Returned 取消离开（"老人坐回床"）；窗口超时无回弹才真正 Vacant。
	//
	// 敏感度建议：
	//   高（即时）  : 0~2s  — 老人下床快或不需要软确认的场景
	//   中（推荐）  : 5~8s  — 老人坐起再躺 / 试探离床
	//   低（保守）  : 15s+ — 老人极慢、误报代价高
	LeavingWindowSec int `yaml:"leaving_window_sec"`
}

// FeedbackRules 负反馈规则（MVP 仅 self_contradiction + subset_invariant）。
type FeedbackRules struct {
	SelfContradiction SelfContradictionRule `yaml:"self_contradiction"`
	SubsetInvariant   SubsetInvariantRule   `yaml:"subset_invariant"`
}

// SelfContradictionRule 例：bed=occupied 后 15s 内 HR=0 RR=0 持续 → rollback。
type SelfContradictionRule struct {
	WindowSec        int `yaml:"window_sec"`         // 翻转后多久内检测
	RequireSustainAt int `yaml:"require_sustain_at"` // 要求 score 维持在 ≥ 此值，否则视为矛盾
	PenaltyWeight    int `yaml:"penalty_weight"`     // 受影响信号源临时降权
	PenaltyDurMin    int `yaml:"penalty_duration_min"`
}

// SubsetInvariantRule bed⊆room 违反 → 强制 reconcile。
//
// 两个修复入口：
//  1. 即时修复 (Engine.maybeReconcileSubset)：bed flip 时同步抬升 room。
//  2. 周期巡检 (Engine.repairSubsetInvariant，由 Tick 驱动)：
//     双向修复 — bed.IsPresent + room.Vacant 时
//     · bed UpdatedAt > stale_bed_threshold_sec → 强制 bed=Vacant（信 room）
//     · 否则 → 抬升 room=Occupied（信 bed）
//     这两个修复入口配套防止"room 误判 Vacant 但 bed 长期无更新"两难。
type SubsetInvariantRule struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"` // "lift_parent" / "drop_child"

	// RepairIntervalSec 周期巡检间隔（秒），0 表示禁用周期修复，只靠 maybeReconcileSubset 即时入口。
	RepairIntervalSec int `yaml:"repair_interval_sec"`

	// StaleBedThresholdSec bed.UpdatedAt 超过此值视为 stale。
	// 默认 300s (5min) — 远长于任何正常事件间隔，仅捕获"sleepace 设备掉线 + bed 卡 Occupied"等场景。
	StaleBedThresholdSec int `yaml:"stale_bed_threshold_sec"`
}

// RulesStore 运行时 rules 容器，支持 hot reload。
type RulesStore struct {
	mu    sync.RWMutex
	rules *Rules
}

func NewRulesStore(initial *Rules) *RulesStore {
	if initial == nil {
		initial = DefaultRules()
	}
	return &RulesStore{rules: initial}
}

func (s *RulesStore) Get() *Rules {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules
}

func (s *RulesStore) Replace(r *Rules) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = r
}

// LoadRulesFromFile 从 zone_rules.yaml 加载。
func LoadRulesFromFile(path string) (*Rules, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	r := &Rules{}
	if err := yaml.Unmarshal(raw, r); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	return r, nil
}

// DefaultRules provisional values。来源标注在 doc/zone_rules_eval.md。
// 用户最后 calibrate；这里只保证骨架能跑、相对关系合理。
func DefaultRules() *Rules {
	return &Rules{
		Bed: ZoneRules{
			Enter: map[string]EvidenceStrength{
				"sleepace": {Strength: 90, LatchSec: 10},
				"radar":    {Strength: 80, LatchSec: 10},
			},
			Sustain: SustainRules{
				HRRRPresentAny: struct {
					Strength int      `yaml:"strength"`
					Sources  []string `yaml:"sources"`
				}{Strength: 80, Sources: []string{"sleepad", "radar"}},
				RecentEnterBonus: struct {
					WindowSec int `yaml:"window_sec"`
					Bonus     int `yaml:"bonus"`
				}{WindowSec: 60, Bonus: 15},
			},
			Leave: LeaveRules{
				SizeDependent: true,
				SmallBed: LeaveBedSizeBucket{
					BedKinds: []string{"standard", "hospital", "twin"},
					Sources: map[string]EvidenceStrength{
						// LatchSec 仅防 vendor 自抖动，不应阻塞老人正常"坐起→坐回"行为，
						// 故 2s 足够（小于 leaving_window_sec=8s，让 Returned 路径生效）。
						"sleepace": {Strength: 80, LatchSec: 2},
						"radar":    {Strength: 70, LatchSec: 2},
					},
				},
				LargeBed: LeaveBedSizeBucket{
					BedKinds: []string{"full", "queen", "king", "california_king"},
					Sources: map[string]EvidenceStrength{
						"sleepace": {Strength: 70, LatchSec: 2}, // 大床覆盖不全，降到 radar 同档
						"radar":    {Strength: 70, LatchSec: 2},
					},
				},
			},
			StateMachine: StateMachineRules{
				EnterThreshold:   50,
				ExitThreshold:    -50,
				HysteresisSec:    3,
				LeavingWindowSec: 8, // 中敏感度上限，老人友好
			},
		},
		// Room/Bathroom: provisional baseline（按 Bed 段相对比例外推；用户 calibrate）。
		//
		// 注意 NumberPeople 信号统一以 Kind="count_change" 走旁路（只改 Count 字段不翻
		// occupied），故此处 Leave.Default 不配 number_people_0：count=0 不等于"人离开"
		// （可能只是静止 / 检测失败 / 雷达盲区）。如未来要支持 count=0 触发离开，应在
		// Tick 中加"持续 N 秒 count=0 且无其他 occupied 证据"的复合规则，不能简单 leave。
		Room: ZoneRules{
			Enter: map[string]EvidenceStrength{
				"radar":   {Strength: 80, LatchSec: 2},
				"polygon": {Strength: 70, LatchSec: 2}, // 自实现多边形上线后用
			},
			Leave: LeaveRules{
				Default: map[string]EvidenceStrength{
					"radar":   {Strength: 80, LatchSec: 2},
					"polygon": {Strength: 70, LatchSec: 2},
				},
			},
			StateMachine: StateMachineRules{
				EnterThreshold:   50,
				ExitThreshold:    -50,
				HysteresisSec:    2,
				LeavingWindowSec: 5, // room 边界相对清晰，短一些
			},
		},
		Bathroom: ZoneRules{
			Enter: map[string]EvidenceStrength{
				"radar":   {Strength: 80, LatchSec: 2},
				"polygon": {Strength: 70, LatchSec: 2},
			},
			Leave: LeaveRules{
				Default: map[string]EvidenceStrength{
					"radar":   {Strength: 80, LatchSec: 2},
					"polygon": {Strength: 70, LatchSec: 2},
				},
			},
			StateMachine: StateMachineRules{
				EnterThreshold:   50,
				ExitThreshold:    -50,
				HysteresisSec:    2,
				LeavingWindowSec: 5,
			},
		},
		Feedback: FeedbackRules{
			SelfContradiction: SelfContradictionRule{
				WindowSec:        15,
				RequireSustainAt: 30,
				PenaltyWeight:    5,
				PenaltyDurMin:    60,
			},
			SubsetInvariant: SubsetInvariantRule{
				Enabled:              true,
				Mode:                 "lift_parent", // bed=occupied 但 room=vacant → 抬升 room
				RepairIntervalSec:    10,            // 每 10s 周期巡检
				StaleBedThresholdSec: 300,           // 5min 无更新视为 stale
			},
		},
	}
}

// BedSizeBucket 判断床型属于 small 还是 large bucket（standard/unknown 走 small）。
func BedSizeBucket(kind string) string {
	switch kind {
	case "full", "queen", "king", "california_king":
		return "large"
	default:
		return "small"
	}
}
