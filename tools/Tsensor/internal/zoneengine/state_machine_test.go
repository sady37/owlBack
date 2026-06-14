package zoneengine

import "testing"

func newBedSM() *StateMachine {
	r := DefaultRules()
	return NewStateMachine(&r.Bed.StateMachine)
}

func TestSM_InitiallyVacant(t *testing.T) {
	sm := newBedSM()
	if sm.Status() != StatusVacant {
		t.Errorf("state machine should start StatusVacant, got %v", sm.Status())
	}
	if sm.IsPresent() {
		t.Errorf("initially IsPresent should be false")
	}
}

func TestSM_FlipsOnEnterThreshold(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	r := sm.Evaluate(60, now)
	if !r.Flipped {
		t.Errorf("score=60 should flip to occupied")
	}
	if r.NewStatus != StatusOccupied {
		t.Errorf("after flip, NewStatus = %v, want StatusOccupied", r.NewStatus)
	}
	if r.Transition != TransitionOccupied {
		t.Errorf("transition = %s, want %s", r.Transition, TransitionOccupied)
	}
}

func TestSM_DoesNotFlipInBand(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	r := sm.Evaluate(30, now)
	if r.Flipped {
		t.Errorf("score=30 should NOT flip")
	}
	if r.NewStatus != StatusVacant {
		t.Errorf("should remain StatusVacant")
	}
}

// Occupied + score≤exit_threshold → Leaving（非 Vacant！）
func TestSM_OccupiedExitThresholdEntersLeaving(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now) // → occupied
	// 等过 hysteresis_sec=3
	r := sm.Evaluate(-60, now+5_000)
	if !r.Flipped {
		t.Errorf("score=-60 after hysteresis should flip from occupied")
	}
	if r.NewStatus != StatusLeaving {
		t.Errorf("expected StatusLeaving (老人友好软离开), got %v", r.NewStatus)
	}
	if r.Transition != TransitionLeaving {
		t.Errorf("transition = %s, want %s", r.Transition, TransitionLeaving)
	}
}

// Leaving + score 回升 → Returned
func TestSM_LeavingReturnsOnRecoveredScore(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)        // occupied
	sm.Evaluate(-60, now+5_000) // leaving
	// score 回升到 +60 ≥ enter_threshold
	r := sm.Evaluate(60, now+6_000)
	if !r.Flipped {
		t.Errorf("score recovered should flip out of Leaving")
	}
	if r.NewStatus != StatusOccupied {
		t.Errorf("expected StatusOccupied (returned), got %v", r.NewStatus)
	}
	if r.Transition != TransitionReturned {
		t.Errorf("transition = %s, want %s (老人坐回床)", r.Transition, TransitionReturned)
	}
}

// Leaving + timer 超时 → Vacant
func TestSM_LeavingTimeoutFlipsVacant(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)        // occupied
	sm.Evaluate(-60, now+5_000) // leaving (默认 leaving_window_sec=8)
	// score 仍然负，但还在窗口内
	r := sm.Evaluate(-30, now+10_000)
	if r.Flipped {
		t.Errorf("within leaving window should not flip yet")
	}
	if r.NewStatus != StatusLeaving {
		t.Errorf("should remain Leaving")
	}
	// 超过 leaving_window_sec=8s（自 leaving 开始算 → now+5000+8000+1ms=now+13001）
	r = sm.Evaluate(-30, now+5_000+8_001)
	if !r.Flipped {
		t.Errorf("after leaving_window_sec, should flip Vacant")
	}
	if r.NewStatus != StatusVacant {
		t.Errorf("expected StatusVacant, got %v", r.NewStatus)
	}
	if r.Transition != TransitionVacant {
		t.Errorf("transition = %s, want %s", r.Transition, TransitionVacant)
	}
}

// 反向滞回：Occupied → Leaving 后立刻又 score 回升不能立刻 Returned（hysteresis 限制不适用 Returned，
// 因为 Returned 不算反向 ——验证它仍然能立即响应老人坐回）
func TestSM_ReturnedNotBlockedByHysteresis(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)        // occupied
	sm.Evaluate(-60, now+5_000) // leaving @ t+5s
	// 1s 后立即 score 回升 → 应允许 Returned
	r := sm.Evaluate(60, now+6_000)
	if !r.Flipped {
		t.Errorf("Returned should NOT be blocked by hysteresis (老人坐回必须立即响应)")
	}
	if r.NewStatus != StatusOccupied {
		t.Errorf("expected occupied after returned, got %v", r.NewStatus)
	}
}

func TestSM_HysteresisBlocksImmediateOccupiedToLeaving(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now) // occupied @ t=0
	// 1s 后即翻 leave → 应被 hysteresis 拒
	r := sm.Evaluate(-80, now+1_000)
	if r.Flipped {
		t.Errorf("Occupied→Leaving within hysteresis_sec should be blocked")
	}
	if r.NewStatus != StatusOccupied {
		t.Errorf("should remain occupied")
	}
}

func TestSM_RollbackForcesState(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)
	sm.Rollback(StatusVacant, now+500)
	if sm.Status() != StatusVacant {
		t.Errorf("after Rollback(StatusVacant), status should be Vacant")
	}
}

// TestSM_LeavingWindowSecFallbackPreventsStuckLeaving 防御：yaml 漏配 leaving_window_sec=0
// 不应导致 Leaving 永远不超时（review fix）。
func TestSM_LeavingWindowSecFallbackPreventsStuckLeaving(t *testing.T) {
	// 模拟 yaml 漏配场景：rules 里 LeavingWindowSec=0
	rules := &StateMachineRules{
		EnterThreshold:   50,
		ExitThreshold:    -50,
		HysteresisSec:    3,
		LeavingWindowSec: 0, // ← 漏配
	}
	sm := NewStateMachine(rules)
	// NewStateMachine 应已 patch 兜底
	if rules.LeavingWindowSec != DefaultLeavingWindowSecFallback {
		t.Errorf("NewStateMachine should patch LeavingWindowSec to %d, got %d",
			DefaultLeavingWindowSecFallback, rules.LeavingWindowSec)
	}

	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)        // → Occupied
	sm.Evaluate(-60, now+5_000) // → Leaving

	// 超过兜底 LeavingWindowSec=8s → 应翻 Vacant
	r := sm.Evaluate(-30, now+5_000+9_000)
	if !r.Flipped || r.NewStatus != StatusVacant {
		t.Errorf("Leaving should timeout to Vacant with fallback window; got Flipped=%v Status=%v",
			r.Flipped, r.NewStatus)
	}
}

func TestSM_ForceSetSkipsHysteresis(t *testing.T) {
	sm := newBedSM()
	now := int64(1_000_000_000_000)
	sm.Evaluate(80, now)
	sm.ForceSet(StatusVacant, now+500)
	if sm.Status() != StatusVacant {
		t.Errorf("ForceSet should bypass hysteresis")
	}
}
