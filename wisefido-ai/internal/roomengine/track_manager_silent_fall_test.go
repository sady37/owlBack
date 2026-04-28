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
// 新版 Silent Fall — sleepad LeftBed + radar 仍在 Bed 邻域（PR-2c）
// ============================================================================

// setupBedAt 在 grid 上标一个 AreaBed 矩形（SourceHuman），让 IsNearPriorType 命中。
func setupBedAt(g *RoomGrid, x1, y1, x2, y2 int) {
	for r := 0; r < g.Height; r++ {
		for c := 0; c < g.Width; c++ {
			cx, cy := g.ToCanvas(c, r)
			if cx >= x1 && cx <= x2 && cy >= y1 && cy <= y2 {
				cell := &g.Cells[r*g.Width+c]
				cell.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
				cell.AreaType = AreaBed
			}
		}
	}
}

// TestSilentFallLeftBed_FiresWithoutHRRR：在床 ≥5min，无 HR/RR，LeftBed 后 120s radar 仍在床
func TestSilentFallLeftBed_FiresWithoutHRRR(t *testing.T) {
	tm, g := newTestTM()
	setupBedAt(g, 80, 80, 140, 140) // bed 区域

	const dev = "sleepad-1"
	const t0 = int64(1_000_000)

	// 1) sleepad InBed
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: true, TMs: t0, Status: "instant"})

	// 2) 喂一个 track 到 bed 区附近，且推进时间 ≥5min
	tms := t0
	for i := 0; i < 15; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)
		tms += 1000
	}
	if ts, ok := tm.tracks[0]; ok {
		ts.Verdict = VerdictReal
		ts.Score = ScoreConfirmTh
	}

	// 3) 推进 5min（保 InBed precondition 满足）
	tms = t0 + int64(FallRulesParam.Silent.MinInBedSec+10)*1000
	tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)

	// 4) sleepad LeftBed
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: false, TMs: tms, Status: "instant"})

	// 5) 120s 内仍喂 track（radar 还在 bed 区）
	for i := 0; i < 130; i++ {
		tms += 1000
		tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)
	}

	if got := tm.silentFallLeftbedReported; got != 1 {
		t.Errorf("expected 1 silent_fall_leftbed reported, got %d", got)
	}
}

// TestSilentFallLeftBed_FiresWithHRRR60s：单人 + HR/RR → 60s 阈值
func TestSilentFallLeftBed_FiresWithHRRR60s(t *testing.T) {
	tm, g := newTestTM()
	setupBedAt(g, 80, 80, 140, 140)

	const dev = "sleepad-1"
	const t0 = int64(1_000_000)

	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: true, TMs: t0, Status: "instant"})
	// 提供 HR/RR 观测
	tm.ProcessSleepadObservation(SleepadObservation{
		DeviceUID: dev, TMs: t0 + 1000, InBed: true, HeartRate: 70, RespiratoryRate: 16,
	})

	tms := t0 + int64(FallRulesParam.Silent.MinInBedSec+10)*1000
	for i := 0; i < 5; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)
		tms += 1000
	}
	if ts, ok := tm.tracks[0]; ok {
		ts.Verdict = VerdictReal
		ts.Score = ScoreConfirmTh
	}

	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: false, TMs: tms, Status: "instant"})

	// 60s 内仍在床区
	for i := 0; i < 70; i++ {
		tms += 1000
		tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)
	}
	if got := tm.silentFallLeftbedReported; got != 1 {
		t.Errorf("expected 1 silent_fall_leftbed (60s vital path), got %d", got)
	}
}

// TestSilentFallLeftBed_CancelOnRadarLeavesBed：120s 内 radar 也离开了 bed 区 → 取消
func TestSilentFallLeftBed_CancelOnRadarLeavesBed(t *testing.T) {
	tm, g := newTestTM()
	setupBedAt(g, 80, 80, 140, 140)

	const dev = "sleepad-1"
	const t0 = int64(1_000_000)

	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: true, TMs: t0, Status: "instant"})

	tms := t0 + int64(FallRulesParam.Silent.MinInBedSec+10)*1000
	tm.processFrameAt([]TrackFrame{frameAt(0, 100, 100, 50, observation.PoseLying, tms)}, tms)
	if ts, ok := tm.tracks[0]; ok {
		ts.Verdict = VerdictReal
		ts.Score = ScoreConfirmTh
	}
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: false, TMs: tms, Status: "instant"})

	// 60s 后 radar 移走（远离 bed > 100cm）
	for i := 0; i < 60; i++ {
		tms += 1000
		x := 100 + i*5 // 远离 bed
		tm.processFrameAt([]TrackFrame{frameAt(0, x, 300, 50, observation.PoseStanding, tms)}, tms)
	}
	// 再推到 130s（>120s 等待窗）
	for i := 0; i < 75; i++ {
		tms += 1000
		tm.processFrameAt([]TrackFrame{frameAt(0, 400, 300, 50, observation.PoseStanding, tms)}, tms)
	}
	if got := tm.silentFallLeftbedReported; got != 0 {
		t.Errorf("radar moved away → should NOT fire, got reported=%d", got)
	}
	if got := tm.silentFallLeftbedCancelled; got != 1 {
		t.Errorf("expected 1 cancelled, got %d", got)
	}
}

// TestSilentFallLeftBed_NoFireBelowMinInBed：在床不足 5min 不进入等待状态
func TestSilentFallLeftBed_NoFireBelowMinInBed(t *testing.T) {
	tm, g := newTestTM()
	setupBedAt(g, 80, 80, 140, 140)

	const dev = "sleepad-1"
	const t0 = int64(1_000_000)
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: true, TMs: t0, Status: "instant"})
	tms := t0 + 60*1000 // 仅 60s 在床（< 5min）
	tm.ProcessSleepadBedEvent(SleepadBedEvent{DeviceUID: dev, IsInBed: false, TMs: tms, Status: "instant"})
	for i := 0; i < 200; i++ {
		tms += 1000
		tm.processFrameAt(nil, tms)
	}
	if got := tm.silentFallLeftbedReported; got != 0 {
		t.Errorf("InBed <5min should not arm; got reported=%d", got)
	}
	if got := tm.silentFallLeftbedCancelled; got != 0 {
		t.Errorf("should not even cancel (never armed); got cancelled=%d", got)
	}
}

// ============================================================================
// Still Fall — bathroom + Stand 静止 ≥ 15/18min（PR-3）
// ============================================================================

// runStillStandFor 在 (x,y) 静止站立 N 秒（每秒一帧 pose=Stand）。
// pose 不变、x/y 不变让 frozen 检测尽量不干扰；调用方先把 track 提到 Real。
func runStillStandFor(tm *TrackManager, tid, x, y int, startTms int64, secs int) int64 {
	tms := startTms
	for i := 0; i < secs; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, x, y, 60, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}
	return tms
}

// TestStillFall_FiresOnAreaToiletCell：cell 类型直接是 toilet，stand 静止 ≥ 15min（risk-time）→ 报警
func TestStillFall_FiresOnAreaToiletCell(t *testing.T) {
	tm, g := newTestTM()
	// 把 (100,100) cell 标成 AreaToilet
	c := g.CellAt(100, 100)
	c.Belief[0] = BeliefState{Type: AreaToilet, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaToilet
	// 强制 risk-time（用 UTC，让默认 23:30-07:30 命中 0:00）
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	// 用一个夜间 ts 起步
	startTms := time.Date(2026, 4, 27, 0, 30, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	// runFramesUntilReal 移动到 (100, 100) 附近，Verdict=Real
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1) // dx=1，结束附近 (108,100)
	// 强制把 track 拉到 (100,100)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	// 静止 16min（>15min 风险时段阈值）
	lastMs = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 16*60+5)

	if got := tm.stillFallReportCount; got != 1 {
		t.Errorf("expected 1 still_fall reported (toilet cell + stand 16min night), got %d", got)
	}
}

// TestStillFall_FiresByRoomNameOnly：cell 未学到（AreaUnknown），但 roomName="Bathroom" → 报警
func TestStillFall_FiresByRoomNameOnly(t *testing.T) {
	tm, g := newTestTM()
	tm.SetRoomName("Bathroom") // cell ∪ room 并集中的 room name 路径
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	_ = g
	startTms := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC).UnixMilli() // non-risk → 18min
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	// 19min 静止，跨过 18min non-risk 阈值
	lastMs = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 19*60)
	_ = lastMs

	if got := tm.stillFallReportCount; got != 1 {
		t.Errorf("expected 1 still_fall reported (room=Bathroom + stand 19min day), got %d", got)
	}
}

// TestStillFall_NotFireOutsideBathroom：cell AreaUnknown + roomName="Bedroom" → 不报
func TestStillFall_NotFireOutsideBathroom(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("Bedroom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	_ = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 25*60) // 25min 静止 stand

	if got := tm.stillFallReportCount; got != 0 {
		t.Errorf("not bathroom should not fire still_fall, got %d", got)
	}
}

// TestStillFall_NotFireWhenPoseSit：cell toilet + 但 pose=Sit（坐马桶）→ 不报
func TestStillFall_NotFireWhenPoseSit(t *testing.T) {
	tm, g := newTestTM()
	c := g.CellAt(100, 100)
	c.Belief[0] = BeliefState{Type: AreaToilet, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaToilet
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 27, 0, 30, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// 用 pose=Sit 静止 20min
	tms := lastMs + 1000
	for i := 0; i < 20*60; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 0, observation.PoseSitting, tms)}, tms)
		tms += 1000
	}

	if got := tm.stillFallReportCount; got != 0 {
		t.Errorf("pose=Sit should not fire still_fall (sitting on toilet OK), got %d", got)
	}
}

// TestStillFall_FiresWithStayAlarmEnabledOnly：roomName 不是 bathroom，
// cell 也不是 toilet/shower，但 Stay alarm 启用 → 视为 bathroom，触发 still fall。
func TestStillFall_FiresWithStayAlarmEnabledOnly(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("Bedroom") // 房间名不命中
	tm.SetStayAlarmEnabled(true)
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC).UnixMilli() // non-risk → 18min
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	_ = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 19*60)

	if got := tm.stillFallReportCount; got != 1 {
		t.Errorf("Stay-alarm-enabled bedroom should fire still_fall on stand 19min, got %d", got)
	}
}

// TestStillFall_NoFireWhenStayAlarmDisabled：Bedroom + Stay alarm 关闭 → 不报
// （等价于 case_lostfall 的真实配置）
func TestStillFall_NoFireWhenStayAlarmDisabled(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("Bedroom")
	tm.SetStayAlarmEnabled(false)
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	_ = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 25*60) // 25min stand

	if got := tm.stillFallReportCount; got != 0 {
		t.Errorf("Bedroom + Stay disabled should NOT fire, got %d", got)
	}
}

// ============================================================================
// IsNightTime
// ============================================================================

// ============================================================================
// Frozen-frame 检测：连续 25 帧字面 byte-equal 时 FrozenRunStart 被填上
// ============================================================================

func TestFrozenFrameDetection(t *testing.T) {
	tm, _ := newTestTM()
	const tid = 7
	const x, y, z = 100, 100, 50
	const pose = 4 // standing

	// 第一帧出生
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: 1000},
	}, 1000)
	ts := tm.tracks[tid]
	if ts == nil {
		t.Fatal("track not created")
	}
	if ts.FrozenSameCount != 1 {
		t.Errorf("after 1 frame want FrozenSameCount=1 got %d", ts.FrozenSameCount)
	}
	if ts.FrozenRunStart != 0 {
		t.Errorf("after 1 frame FrozenRunStart should be 0, got %d", ts.FrozenRunStart)
	}

	// 喂 24 帧完全相同（共 25 帧）
	for i := 1; i < 25; i++ {
		tms := int64(1000 + i*1000)
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: tms},
		}, tms)
	}
	if ts.FrozenSameCount != 25 {
		t.Errorf("after 25 same frames want FrozenSameCount=25 got %d", ts.FrozenSameCount)
	}
	// FrozenRunStart 应在 nowMs(=25000) - 24*1000 = 1000 (出生帧)
	expectedStart := int64(25000 - 24*1000)
	if ts.FrozenRunStart != expectedStart {
		t.Errorf("FrozenRunStart want %d got %d", expectedStart, ts.FrozenRunStart)
	}

	// 一帧坐标变化 → 重置
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: x + 50, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: 26000},
	}, 26000)
	if ts.FrozenSameCount != 1 {
		t.Errorf("after pos change want FrozenSameCount=1 got %d", ts.FrozenSameCount)
	}
	if ts.FrozenRunStart != 0 {
		t.Errorf("after pos change FrozenRunStart should reset to 0, got %d", ts.FrozenRunStart)
	}
}

// TestLostFallFrozenThenRecovery 模拟 case_lostfall_cd2b_11351148 的核心时序：
//
//	t=0       人进房，VerdictReal
//	t=20s     track 进入"卧室盲区前"位置 (-90, 320)
//	t=20-50s  连续 30 帧字面相同（frozen）
//	t=50+10s  firmware 放弃 → engine 看不到 → 入 pendingLostFalls
//	t=120s    人重新出现在 (-240, 220) → 取消 pending（盲区返回）
//
// 校验：① pendingLostFalls 被创建 ② 被 birth 取消（不报 lost fall） ③ cell 学到 BlindSpotRecovery
func TestLostFallFrozenThenRecovery(t *testing.T) {
	tm, g := newTestTM()
	// 角落 cell 设为 AreaEnter（离 (-90,320) 足够远，触发 NearestEntryDist > ExitDistMinCm）
	for i := range g.Cells {
		c := &g.Cells[i]
		col := i % g.Width
		row := i / g.Width
		x, y := g.ToCanvas(col, row)
		if x < -200 && y < -200 {
			c.Belief[0].Type = AreaEnter
			c.Belief[0].Confidence = 99
			c.Belief[0].Source = SourceHuman
		}
	}
	// 让消失点 (-90, 320) 的 cell 已累计 Sit 观测 → IsLikelyRestZone=true
	// → checkSilentFall 跳过（视为正常静止丢失）→ 走 lost fall 路径
	// 物理含义：这位置之前有人坐过（沙发未标 layout），所以雷达失锁不算 silent fall
	disappearCell := g.CellAt(-90, 320)
	if disappearCell != nil {
		disappearCell.ActiveType[ActiveIdxSit] = 50 // ≥30 触发 IsLikelyRestZone
	}

	const tid = 0
	startTms := int64(1_000_000)

	// 1) 走 13 帧让 track 升 Real
	tms := runFramesUntilReal(tm, tid, 100, 100, startTms, 30)
	ts := tm.tracks[tid]
	if ts == nil || ts.Verdict != VerdictReal {
		t.Fatalf("track should be Real, got %v", ts)
	}

	// 2) 移到 (-90, 320)（远离 entry，> ExitDistMinCm）
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
	}, tms)
	tms += 1000

	// 3) 连续 30 帧字面完全相同（frozen 模拟 firmware 失锁）
	for i := 0; i < 30; i++ {
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
		}, tms)
		tms += 1000
	}
	if ts.FrozenRunStart == 0 {
		t.Fatalf("FrozenRunStart should be set after 30 frozen frames, got 0")
	}
	if ts.FrozenSameCount < 25 {
		t.Errorf("FrozenSameCount want >=25, got %d", ts.FrozenSameCount)
	}
	frozenStart := ts.FrozenRunStart

	// 4) firmware giveup — 不再喂 frame；engine 走 segment 2 判失锁
	// MaxMissCount = 10，连续 11 个空 tick 后 track 应该消失
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}
	if _, stillExists := tm.tracks[tid]; stillExists {
		t.Fatalf("track should be deleted after MissCount > MaxMissCount")
	}
	if len(tm.pendingLostFalls) != 1 {
		t.Fatalf("expected 1 pendingLostFall, got %d", len(tm.pendingLostFalls))
	}
	var pending *PendingLostFall
	for _, p := range tm.pendingLostFalls {
		pending = p
	}
	if pending.FrozenStartMs != frozenStart {
		t.Errorf("PendingLostFall should carry FrozenStartMs from track, want %d got %d",
			frozenStart, pending.FrozenStartMs)
	}

	// 5) 人重新出现 — 新 track 出生 → 触发 cancel by recovery
	const newTid = 1
	const recoverX, recoverY = -240, 220
	tm.processFrameAt([]TrackFrame{
		{TrackID: newTid, DeviceID: "dev1", X: recoverX, Y: recoverY, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
	}, tms)
	if len(tm.pendingLostFalls) != 0 {
		t.Errorf("pendingLostFalls should be empty after recovery, got %d", len(tm.pendingLostFalls))
	}
	if tm.lostFallPendingCancelled != 1 {
		t.Errorf("lostFallPendingCancelled want 1, got %d", tm.lostFallPendingCancelled)
	}
	if tm.lostFallReported != 0 {
		t.Errorf("lostFallReported want 0 (cancelled in time), got %d", tm.lostFallReported)
	}
	// 新 track 应被升 Real（盲区返回的 verdict bypass）
	newTs := tm.tracks[newTid]
	if newTs == nil {
		t.Fatal("new track not created")
	}
	if newTs.Verdict != VerdictReal {
		t.Errorf("new track should be VerdictReal (recovered_from_lost), got %v", newTs.Verdict)
	}
	if newTs.BirthReason != "recovered_from_lost_fall" {
		t.Errorf("new track BirthReason want 'recovered_from_lost_fall', got %q", newTs.BirthReason)
	}
	// 落点 cell 应有 BlindSpotRecovery 计数
	recoverCell := g.CellAt(recoverX, recoverY)
	if recoverCell == nil || recoverCell.BlindSpotRecoveryCount < 1 {
		t.Errorf("recovery cell should have BlindSpotRecoveryCount>=1, got %v", recoverCell)
	}
}

// TestLostFallExitRoomCancel ExitRoom 事件取消 pending lost-fall
func TestLostFallExitRoomCancel(t *testing.T) {
	tm, g := newTestTM()
	for i := range g.Cells {
		c := &g.Cells[i]
		col := i % g.Width
		row := i / g.Width
		x, y := g.ToCanvas(col, row)
		if x < -200 && y < -200 {
			c.Belief[0].Type = AreaEnter
			c.Belief[0].Confidence = 99
			c.Belief[0].Source = SourceHuman
		}
	}
	if disappearCell := g.CellAt(-90, 320); disappearCell != nil {
		disappearCell.ActiveType[ActiveIdxSit] = 50
	}

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TMs: tms},
	}, tms)
	tms += 1000
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}
	if len(tm.pendingLostFalls) == 0 {
		t.Skip("setup didn't produce pending lost fall")
	}

	// ExitRoom event 到达 → 取消
	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID: "dev1",
		EventName: "ExitRoom",
		TMs:       tms,
	})
	if len(tm.pendingLostFalls) != 0 {
		t.Errorf("pendingLostFalls should be empty after ExitRoom, got %d", len(tm.pendingLostFalls))
	}
	if tm.lostFallPendingCancelled != 1 {
		t.Errorf("lostFallPendingCancelled want 1, got %d", tm.lostFallPendingCancelled)
	}
}

// TestImpliedSpeedFromBirth：出生在 (100,100)，10 秒后跳到 (300,100) 距离 200cm → 隐含 20cm/s
// 然后下一帧到 (500,100) age=11 dist=400 → 隐含 36cm/s（取 max）
func TestImpliedSpeedFromBirth(t *testing.T) {
	tm, _ := newTestTM()
	const tid = 8

	// 出生：t=0, (100,100)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 50, observation.PoseStanding, 0)}, 0)
	ts := tm.tracks[tid]
	if ts == nil {
		t.Fatal("track not created")
	}
	// 10s 后到 (300, 100)，dist=200，age=10s，implied=20 cm/s
	tm.processFrameAt([]TrackFrame{frameAt(tid, 300, 100, 50, observation.PoseStanding, 10000)}, 10000)
	if ts.MaxImpliedSpeedFromBirth < 19 || ts.MaxImpliedSpeedFromBirth > 21 {
		t.Errorf("after 10s + 200cm displacement want implied~20, got %d", ts.MaxImpliedSpeedFromBirth)
	}
	// 11s 总，到 (500, 100)，dist=400，implied=36
	tm.processFrameAt([]TrackFrame{frameAt(tid, 500, 100, 50, observation.PoseStanding, 11000)}, 11000)
	if ts.MaxImpliedSpeedFromBirth < 35 || ts.MaxImpliedSpeedFromBirth > 38 {
		t.Errorf("after 11s + 400cm want implied~36, got %d", ts.MaxImpliedSpeedFromBirth)
	}
	// 回到 (100, 100) implied 应不下降（取 max）
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 50, observation.PoseStanding, 12000)}, 12000)
	if ts.MaxImpliedSpeedFromBirth < 35 {
		t.Errorf("MaxImpliedSpeedFromBirth should be max-tracking, got %d", ts.MaxImpliedSpeedFromBirth)
	}
}

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
