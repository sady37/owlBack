package zoneengine

import (
	"sync"
)

// StateMachine 单 zone 三态状态机：Vacant ↔ Occupied ↔ Leaving（老人友好软离开）。
//
// 翻转规则：
//
//	Vacant   + score ≥ enter_threshold → Occupied
//	Occupied + score ≤ exit_threshold  → Leaving (而非 Vacant，启 leaving timer)
//	Leaving  + score ≥ enter_threshold → Occupied (老人坐回，取消离开)
//	Leaving  + leaving_window_sec 超时 → Vacant   (真正确认离开)
//
// 滞回：翻转后 hysteresis_sec 内拒绝反向翻转（防抖）。注意 Leaving→Occupied (Returned) 不算
// 反向翻转，因为方向一致（都是 IsPresent=true）。
//
// Transition="count_change" 不经过 StateMachine（Engine 内单独路径）。
//
// 内部状态：status / lastTransition / leavingSince。ZoneState 投影由 Engine 维护。
type StateMachine struct {
	rules *StateMachineRules

	mu             sync.Mutex
	status         ZoneStatus
	lastTransition int64
	leavingSince   int64 // status==Leaving 时记录进入时间，timer 用
}

// DefaultLeavingWindowSecFallback 安全兜底值（防 yaml 漏配 leaving_window_sec=0 时 Leaving 状态永不超时）。
// 这是工程防御性默认，不是业务推荐值——业务侧请显式在 yaml 配置。
const DefaultLeavingWindowSecFallback = 8

// NewStateMachine 创建 state machine。初始 status=Vacant。
// 若 rules.LeavingWindowSec <= 0（yaml 漏配 or 显式 0），用 DefaultLeavingWindowSecFallback 兜底，
// 避免"Leaving 永远不超时"的静默失效。
func NewStateMachine(rules *StateMachineRules) *StateMachine {
	if rules.LeavingWindowSec <= 0 {
		rules.LeavingWindowSec = DefaultLeavingWindowSecFallback
	}
	return &StateMachine{rules: rules}
}

// Status 返回当前状态。
func (sm *StateMachine) Status() ZoneStatus {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.status
}

// IsPresent 是否被占用（含 Leaving）。
func (sm *StateMachine) IsPresent() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.status != StatusVacant
}

// IsOccupied 兼容旧 API：== IsPresent。
func (sm *StateMachine) IsOccupied() bool { return sm.IsPresent() }

// LeavingSince Status==Leaving 时返回进入时间，否则 0。
func (sm *StateMachine) LeavingSince() int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.leavingSince
}

// TransitionResult Evaluate / EvaluateLeavingTimeout 的返回值。
type TransitionResult struct {
	Flipped       bool       // 是否翻转
	NewStatus     ZoneStatus // 翻转后状态（Flipped=false 时 == 翻转前）
	Transition    string     // "occupied" / "vacant" / "leaving" / "returned"（仅 Flipped=true 时填）
	Reason        string     // 翻转原因 / 被拒绝原因（便于 trace）
}

// Evaluate 用当前 score + 时间评估翻转。线程安全。
//
// 调用时机：Engine 在每条 SignalEvidence 处理后 + 周期 Tick 都调用。
// 周期 Tick 必要因为 score 会因 decay 自然下降，且 Leaving timer 由 Tick 推进。
func (sm *StateMachine) Evaluate(score int, nowMs int64) TransitionResult {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Hysteresis 防抖：仅作用于反向翻转。
	//   Vacant→Occupied 和 Occupied→Leaving 之间需要 hysteresis 防止抖动。
	//   Leaving→Occupied (Returned) 不算反向（仍 IsPresent），不受 hysteresis 约束 ——
	//     反而需要尽快响应"老人坐回"。
	withinHyst := sm.rules.HysteresisSec > 0 && sm.lastTransition > 0 &&
		nowMs-sm.lastTransition < int64(sm.rules.HysteresisSec)*1000

	switch sm.status {
	case StatusVacant:
		if withinHyst {
			return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "hysteresis"}
		}
		if score >= sm.rules.EnterThreshold {
			sm.status = StatusOccupied
			sm.lastTransition = nowMs
			sm.leavingSince = 0
			return TransitionResult{Flipped: true, NewStatus: StatusOccupied,
				Transition: TransitionOccupied, Reason: "enter_threshold_crossed"}
		}
		return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "below_enter_threshold"}

	case StatusOccupied:
		if withinHyst {
			return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "hysteresis"}
		}
		if score <= sm.rules.ExitThreshold {
			sm.status = StatusLeaving
			sm.lastTransition = nowMs
			sm.leavingSince = nowMs
			return TransitionResult{Flipped: true, NewStatus: StatusLeaving,
				Transition: TransitionLeaving, Reason: "exit_threshold_crossed_enter_leaving"}
		}
		return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "above_exit_threshold"}

	case StatusLeaving:
		// Returned (回到 Occupied) 不受 hysteresis 约束 —— 老人坐回要立即响应。
		if score >= sm.rules.EnterThreshold {
			sm.status = StatusOccupied
			sm.lastTransition = nowMs
			sm.leavingSince = 0
			return TransitionResult{Flipped: true, NewStatus: StatusOccupied,
				Transition: TransitionReturned, Reason: "score_recovered_returned"}
		}
		// Timer expired check
		if sm.rules.LeavingWindowSec > 0 &&
			nowMs-sm.leavingSince >= int64(sm.rules.LeavingWindowSec)*1000 {
			sm.status = StatusVacant
			sm.lastTransition = nowMs
			sm.leavingSince = 0
			return TransitionResult{Flipped: true, NewStatus: StatusVacant,
				Transition: TransitionVacant, Reason: "leaving_window_expired"}
		}
		return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "in_leaving_window"}
	}

	return TransitionResult{Flipped: false, NewStatus: sm.status, Reason: "unknown_status"}
}

// Rollback 强制回到指定 status（self_contradiction 触发用；仅 occupied→vacant 方向调用）。
// 调用方自己定语义，本方法不校验。
func (sm *StateMachine) Rollback(toStatus ZoneStatus, nowMs int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.status = toStatus
	sm.lastTransition = nowMs
	sm.leavingSince = 0
}

// ForceSet 强制设置状态（subset_invariant lift_parent 等跨 zone 协调用）。
func (sm *StateMachine) ForceSet(toStatus ZoneStatus, nowMs int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.status = toStatus
	sm.lastTransition = nowMs
	sm.leavingSince = 0
}

// LastTransitionTs 上次翻转时间。
func (sm *StateMachine) LastTransitionTs() int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.lastTransition
}

// UpdateRules 替换状态机规则指针（P1-2 hot reload）。
//
// **不重置 status / lastTransition / leavingSince** —— 当前状态机记忆保留。
// 同样套用 LeavingWindowSec=0 兜底（防 yaml 漏配新规则）。
func (sm *StateMachine) UpdateRules(rules *StateMachineRules) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if rules.LeavingWindowSec <= 0 {
		rules.LeavingWindowSec = DefaultLeavingWindowSecFallback
	}
	sm.rules = rules
}
