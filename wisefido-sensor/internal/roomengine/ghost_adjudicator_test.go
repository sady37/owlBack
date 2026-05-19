// ghost_adjudicator_test.go — sensor_v2 PR-4 验收
//
// 覆盖 4 条铁律：
//   1. Noop 默认（PR-4 行为 == v1）
//   2. Bathroom 委托 General（PR-4 行为 == 走默认 General）
//   3. applyVerdictDeltas 应用 NewVerdict / PenaltyDelta + clamp
//   4. **Anchored → Ghost 拒绝**（Q4 守卫；PR-6 BathroomGhost 实测前必须立住）

package roomengine

import (
	"testing"

	"go.uber.org/zap"
)

func TestNoopAdjudicatorReturnsNoDeltas(t *testing.T) {
	a := NoopGhostAdjudicator{}
	bases := []TrackStatusBase{{TrackID: 1, Verdict: VerdictReal}}
	got := a.Adjudicate(bases, 1000)
	if len(got) != 0 {
		t.Fatalf("Noop should return empty deltas, got %d", len(got))
	}
}

// 核心 Q4 守卫测试 — Anchored 不可翻 Ghost。
func TestApplyVerdictDeltas_AnchoredToGhostRejected(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 1, Verdict: VerdictAnchored, GhostPenalty: 20},
	}
	ghost := VerdictGhost
	e.applyVerdictDeltas(statuses, []VerdictDelta{
		{TrackID: 1, NewVerdict: &ghost, PenaltyDelta: 60, Reason: "split_ghost_detected"},
	})

	if statuses[0].Verdict != VerdictAnchored {
		t.Fatalf("Anchored must not flip to Ghost, got %v", statuses[0].Verdict)
	}
	// PenaltyDelta 仍累加（PR-6 BathroomGhost 持续观察）
	if statuses[0].GhostPenalty != 80 {
		t.Fatalf("PenaltyDelta should still accumulate on rejected verdict change, got %d", statuses[0].GhostPenalty)
	}
}

// Anchored → Pending / Real 不在禁单（理论上 PR-X 显式 unanchor 路径），但 PR-4 阶段也确认行为：
// applyVerdictDeltas 不阻止，让上层 (即未来的 FeedbackEvent) 决策。
func TestApplyVerdictDeltas_AnchoredToReal_Allowed(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 1, Verdict: VerdictAnchored},
	}
	real := VerdictReal
	e.applyVerdictDeltas(statuses, []VerdictDelta{
		{TrackID: 1, NewVerdict: &real, Reason: "unanchor_via_feedback"},
	})
	if statuses[0].Verdict != VerdictReal {
		t.Fatalf("Anchored → Real should be allowed (not in Q4 禁单), got %v", statuses[0].Verdict)
	}
}

// Pending → Ghost / Real → Ghost 等正常 verdict 翻转必须 honor。
func TestApplyVerdictDeltas_NormalTransitions(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 1, Verdict: VerdictPending},
		{TrackID: 2, Verdict: VerdictReal},
	}
	ghost := VerdictGhost
	real := VerdictReal
	e.applyVerdictDeltas(statuses, []VerdictDelta{
		{TrackID: 1, NewVerdict: &ghost, Reason: "pending_to_ghost"},
		{TrackID: 2, NewVerdict: &real, Reason: "real_kept"}, // 同状态写入也是合法
	})
	if statuses[0].Verdict != VerdictGhost {
		t.Fatalf("Pending → Ghost should apply, got %v", statuses[0].Verdict)
	}
	if statuses[1].Verdict != VerdictReal {
		t.Fatalf("Real → Real should apply, got %v", statuses[1].Verdict)
	}
}

// PenaltyDelta clamp 边界（[0, 100]）。
func TestApplyVerdictDeltas_PenaltyClamp(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 1, GhostPenalty: 80},
		{TrackID: 2, GhostPenalty: 10},
	}
	e.applyVerdictDeltas(statuses, []VerdictDelta{
		{TrackID: 1, PenaltyDelta: 50},  // 80 + 50 = 130 → clamp 100
		{TrackID: 2, PenaltyDelta: -30}, // 10 - 30 = -20 → clamp 0
	})
	if statuses[0].GhostPenalty != 100 {
		t.Fatalf("clamp upper should be 100, got %d", statuses[0].GhostPenalty)
	}
	if statuses[1].GhostPenalty != 0 {
		t.Fatalf("clamp lower should be 0, got %d", statuses[1].GhostPenalty)
	}
}

// Unknown TrackID 静默丢弃（防 stale delta）。
func TestApplyVerdictDeltas_UnknownTrackIDDropped(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 1, Verdict: VerdictReal},
	}
	ghost := VerdictGhost
	e.applyVerdictDeltas(statuses, []VerdictDelta{
		{TrackID: 99, NewVerdict: &ghost, Reason: "stale"},
	})
	if statuses[0].Verdict != VerdictReal {
		t.Fatalf("delta for unknown TrackID should not affect existing tracks")
	}
}

// pickAdjudicator 分支选择正确性。
func TestPickAdjudicator_RoomKindDispatch(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	general := NoopGhostAdjudicator{}
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	bath := NewBathroomGhostAdjudicator(census, nil, nil, zap.NewNop())
	e.SetGhostAdjudicators(general, bath)

	if got := e.pickAdjudicator("bathroom"); got != bath {
		t.Fatalf("bathroom room should pick bathroom adjudicator, got %T", got)
	}
	if got := e.pickAdjudicator(""); got != general {
		t.Fatalf("empty kind (default bedroom) should pick general, got %T", got)
	}
	if got := e.pickAdjudicator("livingroom"); got != general {
		t.Fatalf("non-bathroom kind should pick general, got %T", got)
	}
}

// pickAdjudicator 未注入时 fallback Noop。
func TestPickAdjudicator_UninitializedFallsBackToNoop(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	got := e.pickAdjudicator("bathroom")
	if _, ok := got.(NoopGhostAdjudicator); !ok {
		t.Fatalf("uninitialized engine should fallback to Noop, got %T", got)
	}
}

// nil bathroom → Noop fallback（PR-6 后约定：未注入真实 BathroomGhostAdjudicator
// 时直接走 Noop，不再 wrap general — 因为 BathroomGhostAdjudicator 不再有 delegate 字段）。
func TestSetGhostAdjudicators_NilBathroomFallsBackToNoop(t *testing.T) {
	e := &Engine{logger: zap.NewNop()}
	general := NoopGhostAdjudicator{}
	e.SetGhostAdjudicators(general, nil)
	if _, ok := e.bathroomGhostAdj.(NoopGhostAdjudicator); !ok {
		t.Fatalf("nil bathroom 应 fallback Noop，got %T", e.bathroomGhostAdj)
	}
}

// --- helpers ---

type stubAdj struct {
	deltas []VerdictDelta
}

func (s *stubAdj) Adjudicate(_ []TrackStatusBase, _ int64) []VerdictDelta {
	return s.deltas
}
