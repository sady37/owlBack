package roomengine

import (
	"testing"

	"owl-common/observation"
)

// TestGhost_SingleTrackNotJudgedGhost — ghost≥2 闸（用户第6点）。
// ghost = 真 track 的镜像/反射，无第二条 track 作母体就不该判 ghost。
// 单 track 即便 GhostPenalty≥阈值（出生远离门 / 不可能速度累积）也不判 ghost——
// 否则会把"从盲区返回的真人 / 倒地真摔"误判 ghost 压制掉。
//
// 现状（无闸）：单 track 高 penalty → Verdict=Ghost → 本用例 FAIL。
// 加闸后：单 track → 不判 ghost → GREEN。
func TestGhost_SingleTrackNotJudgedGhost(t *testing.T) {
	tm, _ := newTestTM()
	tms := int64(1_000_000)

	// 出生 track 0
	tm.processFrameAt([]TrackFrame{frameAt(0, 300, 300, 50, observation.PoseWalking, tms)}, tms)
	ts := tm.tracks[0]
	if ts == nil {
		t.Fatal("track 0 未创建")
	}
	// 强制高 ghost penalty + Real + 关掉豁免
	ts.Verdict = VerdictReal
	ts.GhostPenalty = 90
	ts.LongSurvivalAnchored = false
	ts.StartupGrace = false

	// 不喂帧触发 verdict loop（段1 不重置；单 track）
	tms += 1000
	tm.processFrameAt(nil, tms)

	if tm.tracks[0] != nil && tm.tracks[0].Verdict == VerdictGhost {
		t.Errorf("单 track + GhostPenalty=90 不该判 ghost（ghost 需 ≥2 母体），got Verdict=Ghost")
	}
}
