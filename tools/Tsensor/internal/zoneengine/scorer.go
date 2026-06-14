package zoneengine

import (
	"strings"
	"sync"
)

// Scorer 单 zone 评分器（线程安全）。
//
// 一个 (cardID, zoneType, zoneID) 三元组对应一个 Scorer 实例；engine 主表负责实例化。
//
// 模型：
//
//	pos = max(enter.strength_decayed, sustain.strength_decayed) + recent_enter_bonus
//	neg = leave.strength_decayed
//	score = pos - neg                     // ∈ [-100, +100+bonus]
//
// 多源同方向取 max（非 sum）—— sleepace + radar 同时报 InBed 不会累计成 170 假强。
// 反向 latch：高优先级 evidence 在 latch 窗内拒绝反向信号（"厂家权威"语义）。
// 衰减：线性，120s 走完一档 strength（DecayWindowMs 可调）。
type Scorer struct {
	rules     *ZoneRules
	bedBucket string // "small" / "large" / "" (非 bed zone)

	mu sync.Mutex

	// 三类 evidence 各自的最新一条（同方向只保留最强 + 最近）
	enterStrength   int
	enterTs         int64
	enterSource     string
	enterLatchUntil int64

	leaveStrength   int
	leaveTs         int64
	leaveSource     string
	leaveLatchUntil int64

	sustainStrength int
	sustainTs       int64

	// 用于 recent_enter_bonus 判定（任意 enter 事件最新时间，与 enterTs 同步刷新）
	lastEnterEventTs int64
}

// DecayWindowMs 各方向 strength 线性衰减完成时间。
// 用 120s：sleepace InBed 单次后，120s 无补充信号就完全失效 —— 跟实际睡眠 / 离床
// 时间尺度匹配。HR/RR sustain 自然会持续补强，不会跑到衰减完。
const DecayWindowMs = 120_000

// SustainStaleMs sustain evidence 超过这个时间没刷新即视为停。
//
// 60s = sleepad 4G 设备最大上报周期 30s + 一次丢包余量。WiFi 设备 ~2-5s 一帧远小于此。
// 2026-05-24 从 10s 改 60s：旧值小于设备上报间隔，导致 sustain 必然 stale；配合 self_contradiction
// window 15s 必触发 rollback Vacant，造成 bed FSM 5-7s 周期翻转（见 doc/cases/yang-r-flip-0524）。
const SustainStaleMs = 60_000

// NewScorer 创建 scorer。bedBucket 仅在 Bed zone + LeaveRules.SizeDependent=true 时使用，
// 其他 zone 传 "" 即可。
func NewScorer(rules *ZoneRules, bedBucket string) *Scorer {
	return &Scorer{rules: rules, bedBucket: bedBucket}
}

// UpdateRules 替换 scorer 内部使用的规则指针（P1-2 hot reload）。
//
// **不重置任何运行时状态**（enterStrength / leaveStrength / sustainStrength / latch 等）
// —— 仅切换规则表。bedBucket 来自 BedSizeLookup 跟 yaml 无关，不动。
//
// 调用方需保证 rules 对应同一 ZoneType（caller 责任）。
func (s *Scorer) UpdateRules(rules *ZoneRules) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = rules
}

// Score 计算当前 score（线程安全）。
func (s *Scorer) Score(nowMs int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.compute(nowMs)
}

// Snapshot 提供 score + 最近触发的 source 信息（trace 用）。
type ScoreSnapshot struct {
	Score        int
	EnterSource  string
	LeaveSource  string
	SustainAlive bool
}

// lastSustainTsForDebug 供 debug log 用，读 sustainTs 加锁。
func (s *Scorer) lastSustainTsForDebug() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sustainTs
}

// LastEvidenceTs 最近一次"被接受的 evidence"时间戳（max(enterTs, leaveTs, sustainTs)）。
// 供 stale 检测用：repair 路径不能用 ZoneState.UpdatedAt（Tick 每秒刷它），
// 必须基于真实 evidence 时间。被拒绝的 evidence 不影响此值。
func (s *Scorer) LastEvidenceTs() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	max := s.enterTs
	if s.leaveTs > max {
		max = s.leaveTs
	}
	if s.sustainTs > max {
		max = s.sustainTs
	}
	return max
}

func (s *Scorer) Snapshot(nowMs int64) ScoreSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ScoreSnapshot{
		Score:        s.compute(nowMs),
		EnterSource:  s.enterSource,
		LeaveSource:  s.leaveSource,
		SustainAlive: nowMs-s.sustainTs < SustainStaleMs,
	}
}

// Apply 接收一条 SignalEvidence，更新内部累加器；返回 (newScore, accepted)。
// accepted=false 表示被反向 latch 拒绝（信号 trace 仍可记录但不影响 score）。
func (s *Scorer) Apply(ev SignalEvidence, nowMs int64) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accepted := true
	switch strings.ToLower(ev.Kind) {
	case "enter":
		if nowMs < s.leaveLatchUntil {
			accepted = false
			break
		}
		strength := s.lookupEnterStrength(ev.Source)
		latch := s.lookupEnterLatch(ev.Source)
		if strength <= 0 {
			accepted = false // 未配置此源 → 忽略
			break
		}
		// 同方向 max：新 evidence 比当前未衰减 strength 强 OR 当前已衰减 → 替换
		eff := decayStrength(s.enterStrength, nowMs-s.enterTs)
		if strength >= eff {
			s.enterStrength = strength
			s.enterTs = ev.Ts
			s.enterSource = ev.Source
		}
		s.lastEnterEventTs = ev.Ts
		// 反方向清零：enter event 翻转语义意味着"离床证据失效"，避免残留旧 leaveStrength 抑制翻转
		s.leaveStrength = 0
		s.leaveTs = 0
		s.leaveSource = ""
		if latch > 0 {
			until := ev.Ts + int64(latch)*1000
			if until > s.enterLatchUntil {
				s.enterLatchUntil = until
			}
		}

	case "leave":
		if nowMs < s.enterLatchUntil {
			accepted = false
			break
		}
		strength := s.lookupLeaveStrength(ev.Source)
		latch := s.lookupLeaveLatch(ev.Source)
		if strength <= 0 {
			accepted = false
			break
		}
		eff := decayStrength(s.leaveStrength, nowMs-s.leaveTs)
		if strength >= eff {
			s.leaveStrength = strength
			s.leaveTs = ev.Ts
			s.leaveSource = ev.Source
		}
		// 反方向清零：leave event 意味着"在床证据失效"。
		// 物理证据（sustain HR/RR）仍然独立，未必清掉 —— 那是另一回事，sleepace 撒谎时 HR/RR 还会让 score 升回去。
		s.enterStrength = 0
		s.enterTs = 0
		s.enterSource = ""
		if latch > 0 {
			until := ev.Ts + int64(latch)*1000
			if until > s.leaveLatchUntil {
				s.leaveLatchUntil = until
			}
		}

	case "sustain":
		// HR/RR present 等持续证据：strength 取 yaml 配置（任意源同一档）。
		strength := s.rules.Sustain.HRRRPresentAny.Strength
		if strength <= 0 {
			break
		}
		s.sustainStrength = strength
		s.sustainTs = ev.Ts

	case "count_change":
		// NumberPeople 直写不入 score（zone state.count 由 engine state 维护，不通过 score 决策）。
		accepted = true

	default:
		accepted = false
	}

	return s.compute(nowMs), accepted
}

// compute 计算瞬时 score（不加锁，caller 必须已锁）。
func (s *Scorer) compute(nowMs int64) int {
	enterEff := decayStrength(s.enterStrength, nowMs-s.enterTs)
	leaveEff := decayStrength(s.leaveStrength, nowMs-s.leaveTs)

	sustainEff := 0
	if nowMs-s.sustainTs < SustainStaleMs {
		sustainEff = s.sustainStrength
	}

	pos := enterEff
	if sustainEff > pos {
		pos = sustainEff
	}

	// recent_enter_bonus：仅在 sustain 活跃（HR/RR 在）时加成。
	// 用户原意"在床状态 80 + window 内有上床事件 +15"——bonus 是 sustain mode 的附加值，
	// 不是 enter event 自己的加成（否则单 enter event 立即 105 不合理）。
	if sustainEff > 0 {
		if rb := s.rules.Sustain.RecentEnterBonus; rb.WindowSec > 0 && rb.Bonus > 0 {
			if nowMs-s.lastEnterEventTs < int64(rb.WindowSec)*1000 {
				pos += rb.Bonus
			}
		}
	}

	return pos - leaveEff
}

// lookupEnterStrength 找该 source 在 ZoneRules.Enter 的 strength。
func (s *Scorer) lookupEnterStrength(source string) int {
	if s.rules == nil {
		return 0
	}
	if e, ok := s.rules.Enter[source]; ok {
		return e.Strength
	}
	return 0
}

func (s *Scorer) lookupEnterLatch(source string) int {
	if s.rules == nil {
		return 0
	}
	if e, ok := s.rules.Enter[source]; ok {
		return e.LatchSec
	}
	return 0
}

// lookupLeaveStrength leaveRules 可能是 SizeDependent 也可能是 Default。
func (s *Scorer) lookupLeaveStrength(source string) int {
	if s.rules == nil {
		return 0
	}
	leave := s.rules.Leave
	if leave.SizeDependent {
		bucket := leave.SmallBed
		if s.bedBucket == "large" {
			bucket = leave.LargeBed
		}
		if e, ok := bucket.Sources[source]; ok {
			return e.Strength
		}
		return 0
	}
	if e, ok := leave.Default[source]; ok {
		return e.Strength
	}
	return 0
}

func (s *Scorer) lookupLeaveLatch(source string) int {
	if s.rules == nil {
		return 0
	}
	leave := s.rules.Leave
	if leave.SizeDependent {
		bucket := leave.SmallBed
		if s.bedBucket == "large" {
			bucket = leave.LargeBed
		}
		if e, ok := bucket.Sources[source]; ok {
			return e.LatchSec
		}
		return 0
	}
	if e, ok := leave.Default[source]; ok {
		return e.LatchSec
	}
	return 0
}

// decayStrength 线性衰减：t=0 时 strength，t=DecayWindowMs 时 0。
func decayStrength(strength int, ageMs int64) int {
	if strength <= 0 {
		return 0
	}
	if ageMs <= 0 {
		return strength
	}
	if ageMs >= DecayWindowMs {
		return 0
	}
	return int(int64(strength) * (DecayWindowMs - ageMs) / DecayWindowMs)
}
