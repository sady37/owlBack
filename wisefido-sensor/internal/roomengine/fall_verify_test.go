// fall_verify_test.go — A "WeakBio≥80 force real" 短路 T1 单元测试。
//
// 覆盖：
//   - WeakBioSource nil → 不短路（走原评分）
//   - WeakBio score=0 → 不短路
//   - WeakBio score=79 → 不短路（不到阈值）
//   - WeakBio score=80 → force real (verdict="real", score=100)
//   - WeakBio score=100 → force real
//   - WeakBio≥80 + ts.Verdict=Ghost（重 penalty 应判 ghost）→ 仍 force real
//   - WeakBio≥80 仍记 breakdown["weak_bio_force_real"] 审计

package roomengine

import (
	"testing"

	"owl-common/observation"
)

// fakeWeakBioSource 测试 mock，按 spatial prefix 返预设 score。
type fakeWeakBioSource struct {
	scores map[string]int
}

func (f *fakeWeakBioSource) WeakBioScore(sp string) int {
	if f == nil || f.scores == nil {
		return 0
	}
	return f.scores[sp]
}

func newTestVerifyTM(weakBioScore int) *TrackManager {
	tm, _ := newTestTM()
	if weakBioScore > 0 {
		tm.SetWeakBioSource(&fakeWeakBioSource{scores: map[string]int{tm.roomID: weakBioScore}})
	}
	// 给 verifier 一个 minimal track 防 early-return "no_engine_track"
	tm.tracks[7] = &TrackState{
		TrackID:    7,
		Verdict:    VerdictReal,
		Score:      ScoreConfirmTh,
		FrameCount: 20,
		Kalman:     NewKalmanFilter2D(50, 50),
	}
	return tm
}

func TestVerify_NoWeakBioSource_OriginalScoring(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[7] = &TrackState{
		TrackID:    7,
		Verdict:    VerdictReal,
		Score:      ScoreConfirmTh,
		FrameCount: 20,
		Kalman:     NewKalmanFilter2D(50, 50),
	}
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if _, hasFlag := r.Breakdown["weak_bio_force_real"]; hasFlag {
		t.Errorf("no WeakBioSource: should NOT set weak_bio_force_real, got breakdown=%+v", r.Breakdown)
	}
}

func TestVerify_WeakBioZero_NoShortCircuit(t *testing.T) {
	tm := newTestVerifyTM(0) // source set but score=0 (no entry)
	// 上面 helper score=0 时不 set source，强制设 source 但 lookup 返 0
	tm.SetWeakBioSource(&fakeWeakBioSource{scores: map[string]int{}})
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if _, hasFlag := r.Breakdown["weak_bio_force_real"]; hasFlag {
		t.Error("WeakBio score=0 should NOT short-circuit")
	}
}

func TestVerify_WeakBioBelowThreshold_NoShortCircuit(t *testing.T) {
	tm := newTestVerifyTM(79) // 阈值 80，79 不到
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if _, hasFlag := r.Breakdown["weak_bio_force_real"]; hasFlag {
		t.Errorf("WeakBio=79 < 80 should NOT short-circuit; got breakdown=%+v", r.Breakdown)
	}
}

func TestVerify_WeakBioAtThreshold_ForceReal(t *testing.T) {
	tm := newTestVerifyTM(80)
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if r.Verdict != "real" {
		t.Errorf("WeakBio=80: verdict should be 'real', got %q", r.Verdict)
	}
	if r.Score != 100 {
		t.Errorf("WeakBio≥80 force: score should be 100, got %d", r.Score)
	}
	if r.Reason != "weak_bio_force_real" {
		t.Errorf("reason should be 'weak_bio_force_real', got %q", r.Reason)
	}
	if r.Breakdown["weak_bio_force_real"] != 80 {
		t.Errorf("breakdown should record WeakBio=80; got %+v", r.Breakdown)
	}
}

func TestVerify_WeakBioFull_ForceReal(t *testing.T) {
	tm := newTestVerifyTM(100)
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if r.Verdict != "real" || r.Score != 100 {
		t.Errorf("WeakBio=100 force real: got verdict=%q score=%d", r.Verdict, r.Score)
	}
}

// 即使 track 看上去像 ghost (heavy GhostPenalty)，WeakBio≥80 仍强制 real
// （"宁可误报不可漏报" — 老人体征弱时不允许 fall 被 ghost-suppressed）
func TestVerify_WeakBioOverridesGhostPenalty(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[7] = &TrackState{
		TrackID:      7,
		Verdict:      VerdictGhost,
		Score:        10,
		FrameCount:   20,
		GhostPenalty: 80, // 重 ghost 信号，原 score 会 ≤ 30 判 ghost
		Kalman:       NewKalmanFilter2D(50, 50),
	}
	tm.SetWeakBioSource(&fakeWeakBioSource{scores: map[string]int{tm.roomID: 90}})
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if r.Verdict != "real" {
		t.Errorf("WeakBio=90 should override ghost penalty 80; got verdict=%q score=%d breakdown=%+v",
			r.Verdict, r.Score, r.Breakdown)
	}
	// breakdown 应该同时含 ghost_penalty (扣分历史) + weak_bio_force_real (短路标记)
	if _, hasGhost := r.Breakdown["ghost_penalty"]; !hasGhost {
		t.Error("breakdown should still record ghost_penalty for audit")
	}
	if r.Breakdown["weak_bio_force_real"] != 90 {
		t.Errorf("breakdown should record WeakBio=90; got %+v", r.Breakdown)
	}
}

// fakeWeakBioSource 隐式满足 roomengine.WeakBioSource interface 编译期校验。
var _ WeakBioSource = (*fakeWeakBioSource)(nil)

// ---- Cell.FallSuppressUntilMs 拦截点（FE "Lying Area" 反馈临时禁报窗）----

// track 落在 cell 上、FallSuppressUntilMs 在未来 → verifier 短路 verdict=ghost
func TestVerify_CellFallSuppressActive_ShortCircuitGhost(t *testing.T) {
	tm, g := newTestTM()
	tm.tracks[7] = &TrackState{
		TrackID:    7,
		Verdict:    VerdictReal,
		Score:      ScoreConfirmTh,
		FrameCount: 20,
		Kalman:     NewKalmanFilter2D(50, 50),
	}
	const nowMs = int64(1_700_000_000_000)
	// 给 (50, 50) cell 写 2H 禁报窗
	cell := g.CellAt(50, 50)
	if cell == nil {
		t.Fatal("setup: cell (50,50) out of grid")
	}
	cell.FallSuppressUntilMs = nowMs + 2*60*60_000

	a := RadarFallAlarm{TrackID: 7, TMs: nowMs}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, nowMs)
	if r.Verdict != "ghost" || r.Reason != "cell_fall_suppress_window" {
		t.Errorf("active suppress window must short-circuit ghost, got verdict=%q reason=%q breakdown=%+v",
			r.Verdict, r.Reason, r.Breakdown)
	}
	if _, ok := r.Breakdown["cell_fall_suppress_until"]; !ok {
		t.Error("breakdown should record remaining suppress duration")
	}
}

// FallSuppressUntilMs 已过期 → 正常评分路径，不短路
func TestVerify_CellFallSuppressExpired_NoShortCircuit(t *testing.T) {
	tm, g := newTestTM()
	tm.tracks[7] = &TrackState{
		TrackID:    7,
		Verdict:    VerdictReal,
		Score:      ScoreConfirmTh,
		FrameCount: 20,
		Kalman:     NewKalmanFilter2D(50, 50),
	}
	const nowMs = int64(1_700_000_000_000)
	cell := g.CellAt(50, 50)
	cell.FallSuppressUntilMs = nowMs - 1000 // 1s 前已过期

	a := RadarFallAlarm{TrackID: 7, TMs: nowMs}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, nowMs)
	if r.Reason == "cell_fall_suppress_window" {
		t.Errorf("expired suppress must NOT short-circuit, got reason=%q", r.Reason)
	}
}

// FallSuppressUntilMs == 0 (默认/unset) → 不短路
func TestVerify_CellFallSuppressUnset_NoShortCircuit(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[7] = &TrackState{
		TrackID:    7,
		Verdict:    VerdictReal,
		Score:      ScoreConfirmTh,
		FrameCount: 20,
		Kalman:     NewKalmanFilter2D(50, 50),
	}
	a := RadarFallAlarm{TrackID: 7, TMs: 1_700_000_000_000}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	r := tm.verifyRadarFall(a, a.TMs)
	if r.Reason == "cell_fall_suppress_window" {
		t.Errorf("unset suppress (default 0) must NOT short-circuit, got reason=%q", r.Reason)
	}
}
func newTestTM() (*TrackManager, *RoomGrid) {
	g := newTestGrid()
	return NewTrackManager("test-room", g), g
}

func frameAt(tid int, x, y, z int, pose int, tms int64) TrackFrame {
	return TrackFrame{TrackID: tid, DeviceAddr: "dev1", X: x, Y: y, Z: z, Pose: pose, TMs: tms}
}

// runFramesUntilReal 喂 ≥12 帧（保证 AgeSec≥10）+ 强制升 Verdict=Real。
// 测试 grid 无 Enter 区，birthScore 默认起步低，自然升级困难；测试关心消失/融合逻辑，
// 所以兜底强制改 Verdict 不影响本意。
func runFramesUntilReal(tm *TrackManager, tid, x, y int, startTms int64, dx int) int64 {
	tms := startTms
	for i := 0; i < 14; i++ { // 14 帧 × 1s = 13s gap，AgeSec=13 ≥ 10
		tm.processFrameAt([]TrackFrame{frameAt(tid, x+i*dx, y, 50, observation.PoseWalking, tms)}, tms)
		tms += 1000
	}
	if ts, ok := tm.tracks[tid]; ok {
		ts.Verdict = VerdictReal
		ts.Score = ScoreConfirmTh
		ts.CurrentAnomaly = AnomalyNone
	}
	return tms
}

func tickAt(tm *TrackManager, nowMs int64) []TrackOutput {
	return tm.processFrameAt(nil, nowMs)
}

// PR-9: v1 Bed-Fall (AnomalyBedFall) 物理矛盾测试已删除（功能 dropped 段 5）。
// PR-9: v1 bedPersonCount counter 测试已删除（字段 dropped）。
// PR-11 silent_fall 重写后会有等价的"床上方矛盾"测试，用 BedSession + SuiteCensus 表达。

// ============================================================================
func newTestGrid() *RoomGrid {
	g := NewRoomGrid(500, 500, 10)
	for i := range g.Cells {
		g.Cells[i].InRoom = true
		g.Cells[i].InFOV = true
		g.Cells[i].EdgeDist = 100
	}
	return g
}

