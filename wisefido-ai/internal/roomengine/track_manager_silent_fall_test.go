package roomengine

import (
	"testing"
	"time"

	"owl-common/observation"
)

// ============================================================================
// Test setup helpers
// ============================================================================

// newTestGrid 50×50 cell（500×500cm）grid，全 InRoom + InFOV，远离边缘
func newTestGrid() *RoomGrid {
	g := NewRoomGrid(500, 500, 10)
	for i := range g.Cells {
		g.Cells[i].InRoom = true
		g.Cells[i].InFOV = true
		g.Cells[i].EdgeDist = 100
	}
	return g
}

func newTestTM() (*TrackManager, *RoomGrid) {
	g := newTestGrid()
	return NewTrackManager("test-room", g), g
}

func frameAt(tid int, x, y, z int, pose int, tms int64) TrackFrame {
	return TrackFrame{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TMs: tms}
}

// runFramesUntilReal 喂 ≥12 帧（保证 AgeSec≥10 满足 checkSilentFall）+ 强制升 Verdict=Real。
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

// ============================================================================
// Silent Fall 基础流程
// ============================================================================

func TestSilentFall_ReportsAfter60s(t *testing.T) {
	tm, _ := newTestTM()
	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 5)

	// 推进让 track 消失（MissCount > MaxMissCount）
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	if tm.SilentFallStatsSnapshot().PendingCreated != 1 {
		t.Fatalf("expected pending=1, got %+v", tm.SilentFallStatsSnapshot())
	}

	// 推进 60+s 超时报警
	lastMs += 65_000
	out := tickAt(tm, lastMs)
	if r := tm.SilentFallStatsSnapshot().Reported; r != 1 {
		t.Errorf("expected reported=1 after timeout, got %d", r)
	}
	found := false
	for _, o := range out {
		if o.Anomaly == AnomalyFall && o.Source == "engine_silent" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected AnomalyFall/engine_silent in outputs, got %v", out)
	}
}

func TestSilentFall_CancelByBirth(t *testing.T) {
	tm, _ := newTestTM()
	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 5)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	if tm.SilentFallStatsSnapshot().PendingCreated != 1 {
		t.Fatalf("setup pending != 1")
	}

	// 30s 后新 track 在 50cm 内出生（pose=Standing 非 Lie）
	lastMs += 30_000
	tm.processFrameAt([]TrackFrame{frameAt(99, 130, 100, 50, observation.PoseStanding, lastMs)}, lastMs)
	if c := tm.SilentFallStatsSnapshot().PendingCancelled; c != 1 {
		t.Errorf("expected cancelled=1, got %d", c)
	}
	// 60s 后不应该报
	lastMs += 60_000
	tickAt(tm, lastMs)
	if r := tm.SilentFallStatsSnapshot().Reported; r != 0 {
		t.Errorf("after cancel expected reported=0, got %d", r)
	}
}

func TestSilentFall_NotCancelByLieBirth(t *testing.T) {
	tm, _ := newTestTM()
	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 5)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	lastMs += 30_000
	tm.processFrameAt([]TrackFrame{frameAt(99, 130, 100, 0, observation.PoseLying, lastMs)}, lastMs)
	if c := tm.SilentFallStatsSnapshot().PendingCancelled; c != 0 {
		t.Errorf("Lie birth should NOT cancel, got cancelled=%d", c)
	}
}

func TestSilentFall_FilteredByAreaBed(t *testing.T) {
	tm, g := newTestTM()
	c := g.CellAt(100, 100)
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaBed

	startMs := int64(1_000_000)
	// dx=0 让 track 静止在 (100,100) AreaBed cell 上，消失时 last pos 也在此
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 0)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	if c := tm.SilentFallStatsSnapshot().PendingCreated; c != 0 {
		t.Errorf("AreaBed cell should be filtered, got pending=%d", c)
	}
}

func TestSilentFall_FilteredByLikelyRest(t *testing.T) {
	tm, g := newTestTM()
	g.CellAt(100, 100).ActiveType[ActiveIdxSit] = 50

	startMs := int64(1_000_000)
	// dx=0 同上理由
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 0)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	if c := tm.SilentFallStatsSnapshot().PendingCreated; c != 0 {
		t.Errorf("ActiveType[Sit]>=30 should filter, got pending=%d", c)
	}
}

// ============================================================================
// Sleepad 融合：silent fall short-circuit
// ============================================================================

func TestSleepadFusion_InBedShortCircuit(t *testing.T) {
	tm, _ := newTestTM()
	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 5)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	// 60s 超时；sleepad 30s 内 InBed
	lastMs += 65_000
	tm.ProcessSleepadObservation(SleepadObservation{
		DeviceUID: "sleepad-1", TMs: lastMs - 5_000, InBed: true,
	})
	tickAt(tm, lastMs)
	stats := tm.SilentFallStatsSnapshot()
	if stats.Reported != 0 {
		t.Errorf("sleepad InBed should short-circuit, reported=%d", stats.Reported)
	}
	if stats.PendingCancelled != 1 {
		t.Errorf("expected cancelled=1, got %d", stats.PendingCancelled)
	}
}

func TestSleepadFusion_StaleNotShortCircuit(t *testing.T) {
	tm, _ := newTestTM()
	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 5)
	for i := 0; i < MaxMissCount+2; i++ {
		lastMs += 1000
		tickAt(tm, lastMs)
	}
	// sleepad 数据是 60s 之前的（最终 nowMs 时已 >30s 老）
	tm.ProcessSleepadObservation(SleepadObservation{
		DeviceUID: "sleepad-1", TMs: lastMs, InBed: true,
	})
	lastMs += 65_000
	tickAt(tm, lastMs)
	if r := tm.SilentFallStatsSnapshot().Reported; r != 1 {
		t.Errorf("stale sleepad should NOT short-circuit, got reported=%d", r)
	}
}

// ============================================================================
// Bed-Fall 物理矛盾
// ============================================================================

func TestBedFall_TriggersOnContradiction(t *testing.T) {
	tm, g := newTestTM()
	c := g.CellAt(100, 100)
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaBed

	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 0) // 静止在床上

	tm.ProcessSleepadObservation(SleepadObservation{
		DeviceUID: "sleepad-1", TMs: lastMs, InBed: false,
	})
	lastMs += 100
	out := tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseStanding, lastMs)}, lastMs)

	found := false
	for _, o := range out {
		if o.Anomaly == AnomalyBedFall {
			found = true
		}
	}
	if !found {
		t.Errorf("expected AnomalyBedFall, got %+v", out)
	}
}

func TestBedFall_NotTriggerWith2Persons(t *testing.T) {
	tm, g := newTestTM()
	c := g.CellAt(100, 100)
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaBed

	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: true, Status: "instant"})
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: true, Status: "instant"})

	startMs := int64(1_000_000)
	lastMs := runFramesUntilReal(tm, 0, 100, 100, startMs, 0)
	tm.ProcessSleepadObservation(SleepadObservation{
		DeviceUID: "s1", TMs: lastMs, InBed: false,
	})
	lastMs += 100
	out := tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseStanding, lastMs)}, lastMs)

	for _, o := range out {
		if o.Anomaly == AnomalyBedFall {
			t.Errorf("2 people should NOT trigger BedFall, got %+v", o)
		}
	}
}

// ============================================================================
// bedPersonCount
// ============================================================================

func TestBedPersonCount_Counting(t *testing.T) {
	tm, _ := newTestTM()

	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: true, Status: "instant"})
	if got := tm.totalBedPeople(); got != 1 {
		t.Fatalf("after InBed expected 1, got %d", got)
	}
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: true, Status: "instant"})
	if got := tm.totalBedPeople(); got != 2 {
		t.Fatalf("after 2nd InBed expected 2, got %d", got)
	}
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: false, Status: "instant"})
	if got := tm.totalBedPeople(); got != 1 {
		t.Fatalf("after LeftBed expected 1, got %d", got)
	}
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: false, Status: "instant"})
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: "s1", IsInBed: false, Status: "instant"})
	if got := tm.totalBedPeople(); got != 0 {
		t.Errorf("expected floor at 0, got %d", got)
	}
}

// ============================================================================
// IsNightTime
// ============================================================================

func TestIsNightTime(t *testing.T) {
	cases := []struct {
		hour, min int
		want      bool
		desc      string
	}{
		{23, 30, true, "23:30 边界 = 夜"},
		{23, 31, true, "23:31 = 夜"},
		{0, 0, true, "00:00 = 夜"},
		{7, 29, true, "07:29 = 夜"},
		{7, 30, false, "07:30 边界 = 白天"},
		{8, 0, false, "08:00 = 白天"},
		{12, 0, false, "12:00 = 白天"},
		{22, 0, false, "22:00 = 白天"},
		{23, 29, false, "23:29 = 白天"},
	}
	for _, c := range cases {
		ts := time.Date(2026, 4, 25, c.hour, c.min, 0, 0, time.Local).UnixMilli()
		if got := IsNightTime(ts, time.Local); got != c.want {
			t.Errorf("%s: IsNightTime(%02d:%02d)=%v, want %v",
				c.desc, c.hour, c.min, got, c.want)
		}
	}
}
