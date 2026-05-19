package roomengine

import (
	"testing"
	"time"

	"owl-common/observation"
	"owl-common/radarutils"
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

	// PR-14 入场门控：radar InBed 同房 ±15s 内到达
	tm.RecordRadarEvent(RadarTrackEvent{DeviceUID: "radar-1", TMs: t0 + 1_000, EventName: "InBed", TrackID: 0, Status: "instant"})

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
	// PR-14 入场门控：radar InBed 同房 ±15s 内到达
	tm.RecordRadarEvent(RadarTrackEvent{DeviceUID: "radar-1", TMs: t0 + 1_000, EventName: "InBed", TrackID: 0, Status: "instant"})

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
	// PR-14 入场门控：radar InBed 同房 ±15s 内到达
	tm.RecordRadarEvent(RadarTrackEvent{DeviceUID: "radar-1", TMs: t0 + 1_000, EventName: "InBed", TrackID: 0, Status: "instant"})

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
// PR-5.4: 双 track 运动对称性（仅 GhostPenalty ∈ [70, 80) 启用）
// ============================================================================

// TestMotionSymmetry_HitsAt30Degrees 紧邻 + 同向位移 → 命中
func TestMotionSymmetry_HitsAt30Degrees(t *testing.T) {
	tm, _ := newTestTM()

	// partner: verdict=Real，位置 (100, 100)，2s 内位移 (20, 0) — 沿 X 走 20cm
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		BirthPos: TimedPoint{X: 80, Y: 100, TMs: 0},
		History: []TimedPoint{
			{X: 80, Y: 100, TMs: 0},
			{X: 100, Y: 100, TMs: 2000},
		},
		LastUpdateMs: 2000,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(100, 100)

	// 候选 ghost: 紧邻 (105, 105), 同向位移 (20, 5) — 与 partner 夹角 ≈ 14°
	ts := &TrackState{
		TrackID:      0,
		GhostPenalty: 70,
		BirthPos:     TimedPoint{X: 85, Y: 100, TMs: 0},
		History: []TimedPoint{
			{X: 85, Y: 100, TMs: 0},
			{X: 105, Y: 105, TMs: 2000},
		},
		LastUpdateMs: 2000,
	}
	ts.Kalman = NewKalmanFilter2D(105, 105)
	tm.tracks[0] = ts

	if !tm.checkMotionSymmetry(ts, 2000) {
		t.Errorf("expected motion symmetry hit (parallel movement, close)")
	}
}

// TestMotionSymmetry_NotHitWhenFar 距离 > 100cm → 不命中
func TestMotionSymmetry_NotHitWhenFar(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		History:      []TimedPoint{{X: 0, Y: 0, TMs: 0}, {X: 20, Y: 0, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(20, 0)

	ts := &TrackState{
		TrackID: 0, GhostPenalty: 70,
		History:      []TimedPoint{{X: 200, Y: 200, TMs: 0}, {X: 220, Y: 200, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	ts.Kalman = NewKalmanFilter2D(220, 200)
	tm.tracks[0] = ts

	if tm.checkMotionSymmetry(ts, 2000) {
		t.Errorf("far apart (>100cm) should not trigger; got hit")
	}
}

// TestMotionSymmetry_NotHitWhenOpposite 反向运动 → 不命中
func TestMotionSymmetry_NotHitWhenOpposite(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		History:      []TimedPoint{{X: 100, Y: 100, TMs: 0}, {X: 120, Y: 100, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(120, 100)

	// 反方向：partner +X 走，ghost 候选 -X 走
	ts := &TrackState{
		TrackID: 0, GhostPenalty: 70,
		History:      []TimedPoint{{X: 130, Y: 105, TMs: 0}, {X: 110, Y: 105, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	ts.Kalman = NewKalmanFilter2D(110, 105)
	tm.tracks[0] = ts

	if tm.checkMotionSymmetry(ts, 2000) {
		t.Errorf("opposite direction should not trigger; got hit")
	}
}

// TestMotionSymmetry_NotHitWhenStatic 双方都静止 → 不命中（位移 < 10cm 噪声过滤）
func TestMotionSymmetry_NotHitWhenStatic(t *testing.T) {
	tm, _ := newTestTM()
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		History:      []TimedPoint{{X: 100, Y: 100, TMs: 0}, {X: 102, Y: 101, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(102, 101)

	ts := &TrackState{
		TrackID: 0, GhostPenalty: 70,
		History:      []TimedPoint{{X: 110, Y: 105, TMs: 0}, {X: 113, Y: 106, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	ts.Kalman = NewKalmanFilter2D(113, 106)
	tm.tracks[0] = ts

	if tm.checkMotionSymmetry(ts, 2000) {
		t.Errorf("both static (<10cm displacement) should not trigger; got hit")
	}
}

// TestMotionSymmetry_NotHitWithoutPartner 没有 verdict=Real 的另一 track → 不命中
func TestMotionSymmetry_NotHitWithoutPartner(t *testing.T) {
	tm, _ := newTestTM()
	ts := &TrackState{
		TrackID: 0, GhostPenalty: 70,
		History:      []TimedPoint{{X: 100, Y: 100, TMs: 0}, {X: 120, Y: 100, TMs: 2000}},
		LastUpdateMs: 2000,
	}
	ts.Kalman = NewKalmanFilter2D(120, 100)
	tm.tracks[0] = ts

	if tm.checkMotionSymmetry(ts, 2000) {
		t.Errorf("no Real partner should not trigger; got hit")
	}
}

// ============================================================================
// PR-7.2: stand-static 自学习 → AreaSit
// ============================================================================

// TestAreaSitAutoLearn_FromActive 12min stand-static in non-Sit cell → AreaSit auto-learned
func TestAreaSitAutoLearn_FromActive(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom") // 非 bathroom；不会触发 still-fall
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)

	// 站立静止 13min（>12min 阈值）
	_ = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 13*60)

	cell := tm.grid.CellAt(100, 100)
	if cell == nil {
		t.Fatalf("cell at (100,100) is nil")
	}
	if cell.Belief[0].Type != AreaSit {
		t.Errorf("expected AreaSit auto-learned, got %v", cell.Belief[0].Type)
	}
	if cell.Belief[0].Source != SourceHuman {
		t.Errorf("expected SourceHuman after lock, got %v", cell.Belief[0].Source)
	}
}

// TestAreaSitAutoLearn_NotInBathroom bathroom-room 不学（让 still-fall 处理）
func TestAreaSitAutoLearn_NotInBathroom(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("Bathroom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 60, observation.PoseStanding, lastMs)}, lastMs)
	_ = runStillStandFor(tm, tid, 100, 100, lastMs+1000, 13*60)

	cell := tm.grid.CellAt(100, 100)
	if cell.Belief[0].Type == AreaSit {
		t.Errorf("bathroom room should NOT auto-learn AreaSit (still-fall takes over); got AreaSit")
	}
}

// TestAreaSitAutoLearn_NotBeforeThreshold 11min < 12min 阈值 → 不学
// 用 z=0 关闭 A 路径（z-jump），单独考察 B 路径（累积时长）阈值生效。
func TestAreaSitAutoLearn_NotBeforeThreshold(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// z=0 全程 → A 路径短路（要求 prev.Z>0）；只测 B 路径
	tms := lastMs + 1000
	for i := 0; i < 11*60; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 0, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}

	cell := tm.grid.CellAt(100, 100)
	if cell.Belief[0].Type == AreaSit {
		t.Errorf("11min < 12min threshold; should NOT auto-learn")
	}
}

// TestAreaSitAutoLearn_PoseSitTriggersRegionStatic
// PR-13 后 pose=Sit 静止 ≥12min 也触发 region static 自学习（不再仅 pose=Stand）。
// 这是 PR-13 的核心改进：pose=Sit/Stand/Lie 都纳入区域静止判定。
func TestAreaSitAutoLearn_PoseSitTriggersRegionStatic(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// 全程 pose=Sit 静止 13min（>12min 默认阈值）
	tms := lastMs + 1000
	for i := 0; i < 13*60; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 0, observation.PoseSitting, tms)}, tms)
		tms += 1000
	}

	cell := tm.grid.CellAt(100, 100)
	if cell.Belief[0].Type != AreaSit || cell.Belief[0].Source != SourceHuman {
		t.Errorf("PR-13: pose=Sit 13min static should auto-learn AreaSit, got type=%v source=%v",
			cell.Belief[0].Type, cell.Belief[0].Source)
	}
}

// ============================================================================
// PR-13: region static 双 cell 自学习 + 90% 容忍 + Z 突变路径
// ============================================================================

// TestRegionStatic_BothCellsMarked B 路径触发后 prev cell + cur cell 都标 AreaSit
func TestRegionStatic_BothCellsMarked(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// 在 (100,100) 与 (108,108) 之间漂移（|dx|≤15）→ region static，跨 cell
	tms := lastMs + 1000
	for i := 0; i < 13*60; i++ {
		x, y := 100, 100
		if i%2 == 0 {
			x, y = 108, 108
		}
		tm.processFrameAt([]TrackFrame{frameAt(tid, x, y, 0, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}

	cellA := tm.grid.CellAt(100, 100)
	cellB := tm.grid.CellAt(108, 108)
	if cellA.Belief[0].Type != AreaSit {
		t.Errorf("cell (100,100) should be AreaSit, got %v", cellA.Belief[0].Type)
	}
	if cellB.Belief[0].Type != AreaSit {
		t.Errorf("cell (108,108) should be AreaSit (PR-13 双 cell 加分), got %v", cellB.Belief[0].Type)
	}
}

// TestRegionStatic_90PercentTolerance 静止 ≥90% 容忍单帧噪声
func TestRegionStatic_90PercentTolerance(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// 13min × 60 = 780 帧；每 20 帧插一次 dx=20 跳变（5% 打断 < 10% 容忍）
	tms := lastMs + 1000
	for i := 0; i < 13*60; i++ {
		x, y := 100, 100
		if i%20 == 0 {
			x = 125 // 单帧大跳：dx=25 (打断)
		}
		tm.processFrameAt([]TrackFrame{frameAt(tid, x, y, 0, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}

	cell := tm.grid.CellAt(100, 100)
	if cell.Belief[0].Type != AreaSit {
		t.Errorf("PR-13 90%% tolerance: 5%% noise frames should not block AreaSit learning, got %v",
			cell.Belief[0].Type)
	}
}

// TestRegionStatic_ResetOnLargeMove dx>50 立即 reset region
func TestRegionStatic_ResetOnLargeMove(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetRoomName("LivingRoom")
	loc, _ := time.LoadLocation("UTC")
	tm.SetTimezone(loc)

	startTms := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC).UnixMilli()
	const tid = 0
	lastMs := runFramesUntilReal(tm, tid, 95, 100, startTms, 1)
	// 6min static 后大跨步（dx=100），region reset；继续 6min 在新位置 → 都不到 12min
	tms := lastMs + 1000
	for i := 0; i < 6*60; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, 100, 100, 0, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}
	for i := 0; i < 6*60; i++ {
		tm.processFrameAt([]TrackFrame{frameAt(tid, 200, 100, 0, observation.PoseStanding, tms)}, tms)
		tms += 1000
	}

	cell := tm.grid.CellAt(100, 100)
	if cell.Belief[0].Type == AreaSit && cell.Belief[0].Source == SourceHuman {
		t.Errorf("两段 6min 各自累积，都不到 12min，不该锁 AreaSit")
	}
}

// ============================================================================
// Frozen 检测（box 判据）：失锁前 30s 滚动窗口内位移 box ≤ StillBoxCm 时 FrozenRunStart 被填上
// ============================================================================

func TestFrozenFrameDetection(t *testing.T) {
	tm, _ := newTestTM()
	const tid = 7
	const x, y, z = 100, 100, 50
	const pose = 4

	// 第 1 帧出生
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: 1000},
	}, 1000)
	ts := tm.tracks[tid]
	if ts == nil {
		t.Fatal("track not created")
	}
	if ts.FrozenRunStart != 0 {
		t.Errorf("after 1 frame FrozenRunStart should be 0 (need >=2 history), got %d", ts.FrozenRunStart)
	}

	// 第 2 帧位置完全相同：进入 still box → FrozenRunStart 设为 History 最早帧（=1000）
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: 2000},
	}, 2000)
	if ts.FrozenRunStart != 1000 {
		t.Errorf("after 2 still frames FrozenRunStart want 1000 (oldest history), got %d", ts.FrozenRunStart)
	}

	// 喂 23 帧抖动在 box 内（±10cm 抖动，仍 ≤ 30cm box）—— byte-equal 旧判据这里就清零，box 判据应保持
	for i := 0; i < 23; i++ {
		tms := int64(3000 + i*1000)
		dx := 0
		if i%2 == 0 {
			dx = 10
		} else {
			dx = -10
		}
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: x + dx, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: tms},
		}, tms)
	}
	if ts.FrozenRunStart == 0 {
		t.Errorf("box judge: still 在 ±10cm 抖动应保持 FrozenRunStart > 0")
	}

	// 一帧大幅跳出 box → 重置（位移 > 30cm）
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: x + 100, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: 26000},
	}, 26000)
	if ts.FrozenRunStart != 0 {
		t.Errorf("after big jump FrozenRunStart should reset, got %d", ts.FrozenRunStart)
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
	// 物理含义：这位置之前有人坐过（沙发未标 layout）；测试 lost fall 走 frozen credit 路径
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
		t.Fatalf("FrozenRunStart should be set after 30 still box frames, got 0")
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

// TestLostFall_SkippedAfterBedsideFall 验证 dedup：track 已报过 bedside_fall
// 后，firmware 失锁不应再进 pendingLostFalls 池（防止同一物理事件双报）。
//
// 场景模拟：
//
//	t=0      track 出生在合法 lost-fall 位置（远离 entry），升 Real
//	t=N      标记 BedsideFallReported = true（模拟 R4 床边晕倒已 fire）
//	t=N+12s  firmware 停止上报 → engine 应判 lost，但因 BedsideFallReported
//	         跳过 pending 入池
func TestLostFall_SkippedAfterBedsideFall(t *testing.T) {
	tm, g := newTestTM()
	// 角落 cell 设为 AreaEnter（让消失点 (-90,320) 距 entry 足够远）
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

	const tid = 0
	startTms := int64(1_000_000)

	// 1) track 升 Real
	tms := runFramesUntilReal(tm, tid, 100, 100, startTms, 30)

	// 2) 移到合法 lost-fall 位置（远离 entry）
	tm.processFrameAt([]TrackFrame{
		{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
	}, tms)
	tms += 1000

	// 3) 标记 BedsideFallReported（模拟 R4 已 fire）
	ts := tm.tracks[tid]
	if ts == nil {
		t.Fatal("track lost before flag set")
	}
	ts.BedsideFallReported = true

	// 4) firmware 停止上报，等 MissCount > MaxMissCount
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}
	if _, exists := tm.tracks[tid]; exists {
		t.Fatal("track should be deleted after MissCount > MaxMissCount")
	}

	// 5) 关键断言：pendingLostFalls 应为空（被 dedup 跳过）
	if got := len(tm.pendingLostFalls); got != 0 {
		t.Errorf("pendingLostFalls should be empty after BedsideFallReported, got %d", got)
	}
}

// ============================================================================
// PR-C: NumberPeople=0 ExitRoom 兜底
// ============================================================================

// settleAtPos 喂 N 帧同位置同 pose，让 Kalman 速度收敛到 ~0；
// 不这样做，runFramesUntilReal 留下的高速度会让 PredictOnly 把消失点甩飞到屋外，CellAt=nil 导致 checkLostFall 直接 false。
func settleAtPos(tm *TrackManager, tid, x, y, z int, pose int, startTms int64, frames int) int64 {
	tms := startTms
	for i := 0; i < frames; i++ {
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: x, Y: y, Z: z, Pose: pose, TrackConfidence: 60, TMs: tms},
		}, tms)
		tms += 1000
	}
	return tms
}

// TestLostFall_NumberPeopleZeroSkipsPendingCreation：
// 实测 D523 firmware 在 FOV 边角离场不发 ExitRoom，但发 number_people=0（早 track=88 36-44ms）。
// 设计：track 终止前若 ±5s 内有 number_people=0 → 跳过 pending 入池，避免误报 lost_fall。
func TestLostFall_NumberPeopleZeroSkipsPendingCreation(t *testing.T) {
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

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)
	// 喂 8 帧让 Kalman 在消失点稳定（避免后续 PredictOnly 把位置甩出屋外）
	tms = settleAtPos(tm, tid, -90, 320, 0, 4, tms, 8)

	// 模拟实测时序：firmware 即将失锁前 number_people=0 到达（无 ExitRoom）
	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID:    "dev1",
		EventName:    "NumberPeople",
		TMs:          tms,
		NumberPeople: 0,
	})
	if tm.lastNumberPeopleZeroMs != tms {
		t.Errorf("lastNumberPeopleZeroMs want %d got %d", tms, tm.lastNumberPeopleZeroMs)
	}

	// 再跑 12 帧空 tick 让 MissCount > MaxMissCount=10 → track 失锁判定
	// MissCount 累到 11 时（约第 4 个 miss）触发 lost-fall 入池逻辑。此时 nowMs - lastNumberPeopleZeroMs ≈ 4s，在 5s 窗口内 → 应跳过 pending 创建
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}

	if got := len(tm.pendingLostFalls); got != 0 {
		t.Errorf("pendingLostFalls should be empty (skipped by number_people=0), got %d", got)
	}
	if tm.lostFallPendingCancelled != 1 {
		t.Errorf("lostFallPendingCancelled want 1 (skip counted), got %d", tm.lostFallPendingCancelled)
	}
	if tm.lostFallPendingCreated != 0 {
		t.Errorf("lostFallPendingCreated want 0, got %d", tm.lostFallPendingCreated)
	}
}

// TestLostFall_NumberPeopleZeroCancelsExisting：
// number_people=0 在 pending 已入池后到达 → 取消已存在的 pending。
// 防御性路径，正常流程不应触发（number_people=0 普遍领先 track 终止），但 firmware
// 时序若反转（pending 先创建，number_people=0 几秒后到）兜底兼容。
func TestLostFall_NumberPeopleZeroCancelsExisting(t *testing.T) {
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

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)
	tms = settleAtPos(tm, tid, -90, 320, 0, 4, tms, 8)
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}
	if len(tm.pendingLostFalls) != 1 {
		t.Fatalf("setup should produce 1 pending lost fall (after settle), got %d", len(tm.pendingLostFalls))
	}

	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID:    "dev1",
		EventName:    "NumberPeople",
		TMs:          tms,
		NumberPeople: 0,
	})
	if got := len(tm.pendingLostFalls); got != 0 {
		t.Errorf("pendingLostFalls should be empty after number_people=0, got %d", got)
	}
	if tm.lostFallPendingCancelled != 1 {
		t.Errorf("lostFallPendingCancelled want 1, got %d", tm.lostFallPendingCancelled)
	}
}

// TestLostFall_NumberPeopleZeroOutsideWindow：
// number_people=0 距离 track 真实失踪 >60s（窗口外）→ 不跳过，pending 正常创建。
// 防止过度抑制：上一个人很久之前离场过，又有新人进屋失踪时 lost_fall 仍能触发。
func TestLostFall_NumberPeopleZeroOutsideWindow(t *testing.T) {
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

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)

	// 一个很久之前的 number_people=0（远早于后续 track 失踪）
	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID:    "dev1",
		EventName:    "NumberPeople",
		TMs:          tms,
		NumberPeople: 0,
	})
	earlyZeroMs := tms

	// settle 到 lost-fall 位置喂 70 帧 — LastObservedMs 推到 number_people=0 之后 70s
	// （超过 NumberPeopleZeroFallbackMs=60s 窗口）
	tms = settleAtPos(tm, tid, -90, 320, 0, 4, tms, 70)

	// firmware 停报：number_people=0 已远早于 last frame（>60s）→ 应正常入池
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}

	if got := len(tm.pendingLostFalls); got != 1 {
		t.Errorf("pendingLostFalls want 1 (number_people=0 too old, %dms before last frame), got %d",
			tms-earlyZeroMs, got)
	}
}

// TestLostFall_FrozenOverridesNumberPeopleZeroSkip：
// frozen 状态下 track 失锁时即使有近期 number_people=0，也不应跳过 pending。
// frozen ↔ number_people=0 互斥；若两者同现说明是 firmware 残影结束（CD2B 类盲区返回），
// 应进 pending 池等 birth-recovery 取消，不能误抑制。
func TestLostFall_FrozenOverridesNumberPeopleZeroSkip(t *testing.T) {
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

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)

	// 在 (-90, 320) 喂 30 帧字面完全相同 → 触发 FrozenRunStart > 0
	for i := 0; i < 30; i++ {
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
		}, tms)
		tms += 1000
	}
	if ts := tm.tracks[tid]; ts == nil || ts.FrozenRunStart == 0 {
		t.Fatalf("track should be in frozen state after 30 identical frames, got %v", ts)
	}

	// number_people=0 在 last frame 1s 后到达（窗口内）
	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID:    "dev1",
		EventName:    "NumberPeople",
		TMs:          tms,
		NumberPeople: 0,
	})

	// firmware 停报，等 MaxMissCount 触发判失锁
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}

	// 关键断言：尽管 number_people=0 在窗口内，frozen 状态强制走 pending（不抑制）
	if got := len(tm.pendingLostFalls); got != 1 {
		t.Errorf("pendingLostFalls want 1 (frozen overrides number_people=0 skip), got %d", got)
	}
	for _, p := range tm.pendingLostFalls {
		if p.FrozenStartMs == 0 {
			t.Errorf("pending should carry FrozenStartMs > 0, got %d", p.FrozenStartMs)
		}
	}
}

// TestLostFall_FrozenPendingNotCancelledByNumberPeopleZero：
// frozen 状态下入了池的 pending，即使后来收到 number_people=0 也不应取消。
// 同 frozen ↔ number_people=0 互斥语义；保留 pending 等 recovery / 正常超时。
func TestLostFall_FrozenPendingNotCancelledByNumberPeopleZero(t *testing.T) {
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

	const tid = 0
	tms := runFramesUntilReal(tm, tid, 100, 100, 1_000_000, 30)
	// 30 帧 frozen
	for i := 0; i < 30; i++ {
		tm.processFrameAt([]TrackFrame{
			{TrackID: tid, DeviceID: "dev1", X: -90, Y: 320, Z: 0, Pose: 4, TrackConfidence: 60, TMs: tms},
		}, tms)
		tms += 1000
	}
	// firmware 停报 → pending 创建（带 FrozenStartMs>0）
	for i := 0; i < 12; i++ {
		tm.processFrameAt(nil, tms)
		tms += 1000
	}
	if len(tm.pendingLostFalls) != 1 {
		t.Fatalf("expected 1 pending (frozen → not skipped), got %d", len(tm.pendingLostFalls))
	}
	var pending *PendingLostFall
	for _, p := range tm.pendingLostFalls {
		pending = p
	}
	if pending.FrozenStartMs == 0 {
		t.Fatalf("pending must carry FrozenStartMs>0 to test frozen guard")
	}

	// number_people=0 到达 → 不应取消（frozen guard）
	tm.RecordRadarEvent(RadarTrackEvent{
		DeviceUID:    "dev1",
		EventName:    "NumberPeople",
		TMs:          tms,
		NumberPeople: 0,
	})
	if got := len(tm.pendingLostFalls); got != 1 {
		t.Errorf("frozen pending should be kept (not cancelled by number_people=0), got %d remaining", got)
	}
}

// TestParseRadarTrackEvents_NumberPeople：解析 number_people category 入 RadarTrackEvent
func TestParseRadarTrackEvents_NumberPeople(t *testing.T) {
	dv := []interface{}{
		map[string]interface{}{
			"track_id":      float64(10),
			"event_since":   float64(1777703748870),
			"event_status":  "start",
			"number_people": float64(0),
		},
	}

	evts := ParseRadarTrackEvents(dv, "E598A2ACD523", "number_people", 0)
	if len(evts) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evts))
	}
	e := evts[0]
	if e.EventName != "NumberPeople" {
		t.Errorf("EventName want 'NumberPeople' got %q", e.EventName)
	}
	if e.NumberPeople != 0 {
		t.Errorf("NumberPeople want 0 got %d", e.NumberPeople)
	}
	if e.TMs != 1777703748870 {
		t.Errorf("TMs want 1777703748870 got %d", e.TMs)
	}

	// number_people=2 也应被解析（虽然 RecordRadarEvent 不动作）
	dv2 := []interface{}{
		map[string]interface{}{
			"track_id":      float64(10),
			"event_since":   float64(1777703704233),
			"event_status":  "start",
			"number_people": float64(2),
		},
	}
	evts2 := ParseRadarTrackEvents(dv2, "E598A2ACD523", "NumberPeople", 0)
	if len(evts2) != 1 || evts2[0].NumberPeople != 2 {
		t.Errorf("expected NumberPeople=2 parsed, got %+v", evts2)
	}
}

// ============================================================================
// PR-D: 镜面对称 ghost 检测
// ============================================================================

// TestReflectAcrossMirror_HorizontalMirror：水平镜（W>=H）反射 Y 分量。
// 镜面 Y=(Y1+Y2)/2，X 不变，Y 关于镜面镜像。
func TestReflectAcrossMirror_HorizontalMirror(t *testing.T) {
	// 长 200cm、宽 10cm 的水平镜，中线 Y=100
	m := radarutils.Rect{X1: 0, Y1: 95, X2: 200, Y2: 105}
	rx, ry := reflectAcrossMirror(50, 30, m)
	if rx != 50 || ry != 170 {
		t.Errorf("horizontal mirror at Y=100: (50,30) -> want (50,170) got (%d,%d)", rx, ry)
	}
	// 镜上方点应反射到下方等距处
	rx2, ry2 := reflectAcrossMirror(150, 200, m)
	if rx2 != 150 || ry2 != 0 {
		t.Errorf("horizontal mirror at Y=100: (150,200) -> want (150,0) got (%d,%d)", rx2, ry2)
	}
}

// TestReflectAcrossMirror_VerticalMirror：垂直镜（H>W）反射 X 分量。
func TestReflectAcrossMirror_VerticalMirror(t *testing.T) {
	// 长 200cm、宽 10cm 的垂直镜，中线 X=100
	m := radarutils.Rect{X1: 95, Y1: 0, X2: 105, Y2: 200}
	rx, ry := reflectAcrossMirror(30, 50, m)
	if rx != 170 || ry != 50 {
		t.Errorf("vertical mirror at X=100: (30,50) -> want (170,50) got (%d,%d)", rx, ry)
	}
}

// TestMirrorSymmetry_HitsRealMirrorImage：典型浴室 case。
// 镜面在中线 Y=100，partner real 在 (50, 30)，候选 ghost 在 (50, 170) → 镜像 (50, 30) ≈ partner → 命中
func TestMirrorSymmetry_HitsRealMirrorImage(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetInterferes([]radarutils.Rect{
		{X1: 0, Y1: 95, X2: 200, Y2: 105},
	})

	// partner 真人 @ (50, 30)，verdict=Real
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		History:      []TimedPoint{{X: 50, Y: 30, TMs: 0}},
		LastUpdateMs: 0,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(50, 30)

	// 候选 ghost @ (50, 170) — 关于 Y=100 镜面对称于 partner
	ts := &TrackState{
		TrackID:      0,
		GhostPenalty: 70, // 卡 [70,80) 边缘
		Kalman:       NewKalmanFilter2D(50, 170),
	}
	tm.tracks[0] = ts

	tm.applyLifetimeGhostFactors(ts, 5_000)
	if ts.GhostPenalty < GhostPenaltyThreshold {
		t.Errorf("expected mirror symmetry hit to push penalty ≥ %d, got %d",
			GhostPenaltyThreshold, ts.GhostPenalty)
	}
	if ts.BirthReason != "mirror_image_of_real_track" {
		t.Errorf("expected BirthReason=mirror_image_of_real_track, got %q", ts.BirthReason)
	}
}

// TestMirrorSymmetry_NoHitWhenNotMirrored：partner 真人不在镜面对称位置 → 不命中
func TestMirrorSymmetry_NoHitWhenNotMirrored(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetInterferes([]radarutils.Rect{
		{X1: 0, Y1: 95, X2: 200, Y2: 105},
	})

	// partner 在 (50, 30)；候选 ghost 在 (200, 170) — 关于镜面镜像应为 (200, 30)，离 partner 150cm，不命中
	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		LastUpdateMs: 0,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(50, 30)

	ts := &TrackState{
		TrackID:      0,
		GhostPenalty: 70,
		Kalman:       NewKalmanFilter2D(200, 170),
	}
	tm.tracks[0] = ts

	tm.applyLifetimeGhostFactors(ts, 5_000)
	if ts.GhostPenalty >= GhostPenaltyThreshold {
		t.Errorf("expected no mirror hit (mirror image far from partner), but penalty=%d crossed threshold", ts.GhostPenalty)
	}
}

// TestMirrorSymmetry_GatedByPenaltyThreshold：penalty 没到 70 时即使镜像几何完美也不触发。
// 与 motion_symmetry 一致的 [70,80) 门槛（避免每帧扫所有 track × interferes）。
func TestMirrorSymmetry_GatedByPenaltyThreshold(t *testing.T) {
	tm, _ := newTestTM()
	tm.SetInterferes([]radarutils.Rect{
		{X1: 0, Y1: 95, X2: 200, Y2: 105},
	})

	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		LastUpdateMs: 0,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(50, 30)

	ts := &TrackState{
		TrackID:      0,
		GhostPenalty: 50, // ★ 没到 70 边缘
		Kalman:       NewKalmanFilter2D(50, 170),
	}
	tm.tracks[0] = ts

	tm.applyLifetimeGhostFactors(ts, 5_000)
	if ts.GhostPenalty != 50 {
		t.Errorf("penalty < 70 should skip symmetry checks, got %d (was 50)", ts.GhostPenalty)
	}
}

// TestMirrorSymmetry_NoInterferes：房间无 Interferes → 不调用 mirror 检测，不影响 motion_symmetry。
func TestMirrorSymmetry_NoInterferes(t *testing.T) {
	tm, _ := newTestTM()
	// 不调用 SetInterferes，tm.interferes 为 nil

	tm.tracks[1] = &TrackState{
		TrackID: 1, Verdict: VerdictReal,
		LastUpdateMs: 0,
	}
	tm.tracks[1].Kalman = NewKalmanFilter2D(50, 30)

	ts := &TrackState{
		TrackID:      0,
		GhostPenalty: 70,
		Kalman:       NewKalmanFilter2D(50, 170),
	}
	tm.tracks[0] = ts

	// motion_symmetry 由于 partner 距离 >100cm 不命中（dist((50,170),(50,30))=140cm）
	// 没有 interferes → mirror_symmetry 不计算 → penalty 应保持 70
	tm.applyLifetimeGhostFactors(ts, 5_000)
	if ts.GhostPenalty != 70 {
		t.Errorf("no interferes should leave penalty unchanged at 70, got %d", ts.GhostPenalty)
	}
}
