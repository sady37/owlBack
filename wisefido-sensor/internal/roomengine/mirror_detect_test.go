// mirror_detect_test.go — L1 镜面对称 ghost 配对检测 Tier1。
//
// 覆盖：
//   - reflectPoint 几何（轴水平 / 竖直 / 斜 45°）
//   - segmentIntersectLine 交点
//   - evaluateMirrorPair 三不变量 + radar 距离 tiebreaker（正例 + 反例：独立行人 / 静态）
//   - MarkMirrorBounce 2×2 累加 + 越界跳过 + ≥3 晋升 + SourceHuman 不被覆盖
//   - persist round-trip v8 MBC/LMM

package roomengine

import (
	"math"
	"testing"

	"owl-common/radarutils"
)

// -- 几何 --

func approxPoint(t *testing.T, got, want radarutils.Point, tolCm int, msg string) {
	t.Helper()
	if absInt(got.X-want.X) > tolCm || absInt(got.Y-want.Y) > tolCm {
		t.Errorf("%s: got (%d,%d), want (%d,%d) ±%d", msg, got.X, got.Y, want.X, want.Y, tolCm)
	}
}
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestReflectPoint_HorizontalAxis(t *testing.T) {
	// 镜面线水平：anchor=(0,10), dir=(1,0)；点 (5, 20) 镜像应为 (5, 0)
	got := reflectPoint(radarutils.Point{X: 5, Y: 20}, radarutils.Point{X: 0, Y: 10}, 1, 0)
	approxPoint(t, got, radarutils.Point{X: 5, Y: 0}, 1, "horizontal mirror")
}

func TestReflectPoint_VerticalAxis(t *testing.T) {
	// 镜面线竖直：anchor=(10,0), dir=(0,1)；点 (20, 5) 镜像应为 (0, 5)
	got := reflectPoint(radarutils.Point{X: 20, Y: 5}, radarutils.Point{X: 10, Y: 0}, 0, 1)
	approxPoint(t, got, radarutils.Point{X: 0, Y: 5}, 1, "vertical mirror")
}

func TestReflectPoint_Diagonal45(t *testing.T) {
	// 镜面线 y=x（斜 45°）：dir=(√2/2, √2/2)；点 (10, 0) 镜像应为 (0, 10)
	d := math.Sqrt(2) / 2
	got := reflectPoint(radarutils.Point{X: 10, Y: 0}, radarutils.Point{X: 0, Y: 0}, d, d)
	approxPoint(t, got, radarutils.Point{X: 0, Y: 10}, 1, "diagonal 45 mirror")
}

func TestSegmentIntersectLine_Crosses(t *testing.T) {
	// 线段 (0,0)→(10,0) 与水平线 y=0 经过 (5,0) → 端点在线上，取中间随便一个交点
	// 用线段 (5, -5)→(5, 5)，水平线 y=0 dir=(1,0) anchor=(0,0) → 交点 (5, 0)
	got, ok := segmentIntersectLine(
		radarutils.Point{X: 5, Y: -5}, radarutils.Point{X: 5, Y: 5},
		radarutils.Point{X: 0, Y: 0}, 1, 0,
	)
	if !ok {
		t.Fatal("should intersect")
	}
	approxPoint(t, got, radarutils.Point{X: 5, Y: 0}, 1, "segment×line")
}

func TestSegmentIntersectLine_SameSide(t *testing.T) {
	// 线段 (1,1)→(2,3) 都在 y=0 上方 → 不相交
	_, ok := segmentIntersectLine(
		radarutils.Point{X: 1, Y: 1}, radarutils.Point{X: 2, Y: 3},
		radarutils.Point{X: 0, Y: 0}, 1, 0,
	)
	if ok {
		t.Error("same side must not intersect")
	}
}

// -- 三不变量配对 --

// makeMirrorSamples 模拟人 A 沿某轨迹移动，B 是 A 关于 y=ymir 水平镜面的镜像
func makeMirrorSamples(ymir int, aPath []radarutils.Point) []mirrorPairSample {
	out := make([]mirrorPairSample, len(aPath))
	for i, a := range aPath {
		b := radarutils.Point{X: a.X, Y: 2*ymir - a.Y}
		out[i] = mirrorPairSample{Ax: a.X, Ay: a.Y, Bx: b.X, By: b.Y, Ts: int64(i * 100)}
	}
	return out
}

func TestEvaluateMirrorPair_HappyPath_HorizontalMirror(t *testing.T) {
	// 镜面 y=100；人 A 沿 (50,150)→(60,160)→(70,170)→(80,180)→(90,190) 走（远离镜面）
	// B 应在 y < 50 区间（镜像），距 radar(0,0) 比 A 远 → ghost = B
	apath := []radarutils.Point{
		{X: 50, Y: 150}, {X: 60, Y: 160}, {X: 70, Y: 170}, {X: 80, Y: 180}, {X: 90, Y: 190},
	}
	samples := makeMirrorSamples(100, apath)
	res := evaluateMirrorPair(samples, 1, 2, 0, 0) // radar at (0, 0)
	if !res.IsMirror {
		t.Fatalf("expected mirror pair detected, got %+v", res)
	}
	// A 在 y>100, B 在 y<100：A 距 radar < B 距 radar (A: ~50-190; B: ~50→-(150-100)*2+... wait)
	// 实际 B = (50, 50), (60, 40), (70, 30), (80, 20), (90, 10) — 全部距 radar (0,0) 比 A 小！
	// 反过来：A 远 → A = ghost
	// 重新看：A 距 radar 大概 √(50²+150²) ~ 158... B = √(50²+50²) ~ 71
	// So B is closer to radar → A is the ghost.
	if res.GhostTrackID != 1 {
		t.Errorf("ghost should be A (id=1, farther from radar) but got id=%d", res.GhostTrackID)
	}
	if res.RealTrackID != 2 {
		t.Errorf("real should be B (id=2, closer)")
	}
	if len(res.BouncePoints) == 0 {
		t.Error("should produce bounce points")
	}
}

func TestEvaluateMirrorPair_TooFewSamples(t *testing.T) {
	apath := []radarutils.Point{{X: 50, Y: 150}, {X: 60, Y: 160}}
	samples := makeMirrorSamples(100, apath)
	res := evaluateMirrorPair(samples, 1, 2, 0, 0)
	if res.IsMirror {
		t.Error("4 samples should not detect (need 5)")
	}
}

func TestEvaluateMirrorPair_StaticScene_Skipped(t *testing.T) {
	// 5 帧两 track 都没动 → 累计位移 < 30cm → 跳过
	stat := []radarutils.Point{
		{X: 50, Y: 150}, {X: 50, Y: 150}, {X: 50, Y: 150}, {X: 50, Y: 150}, {X: 50, Y: 150},
	}
	samples := makeMirrorSamples(100, stat)
	res := evaluateMirrorPair(samples, 1, 2, 0, 0)
	if res.IsMirror {
		t.Error("static scene should be skipped")
	}
}

func TestEvaluateMirrorPair_IndependentWalkers_NotMirror(t *testing.T) {
	// 两个独立行人，运动方向不同 + 速度不同步
	apath := []radarutils.Point{
		{X: 50, Y: 150}, {X: 60, Y: 160}, {X: 70, Y: 170}, {X: 80, Y: 180}, {X: 90, Y: 190},
	}
	// B 不是 A 的镜像，而是另一条独立轨迹
	bpath := []radarutils.Point{
		{X: 200, Y: 50}, {X: 180, Y: 60}, {X: 150, Y: 70}, {X: 130, Y: 80}, {X: 100, Y: 90},
	}
	samples := make([]mirrorPairSample, len(apath))
	for i := range apath {
		samples[i] = mirrorPairSample{
			Ax: apath[i].X, Ay: apath[i].Y,
			Bx: bpath[i].X, By: bpath[i].Y,
			Ts: int64(i * 100),
		}
	}
	res := evaluateMirrorPair(samples, 1, 2, 0, 0)
	if res.IsMirror {
		t.Errorf("independent walkers should NOT trigger; got %+v", res)
	}
}

func TestEvaluateMirrorPair_GhostFarther_TiebreakerCorrect(t *testing.T) {
	// 镜面 x=100（竖直）；人 A 在 x>100 区，radar 在 x=0 → A 比 B 距 radar 远 → A = ghost
	apath := []radarutils.Point{
		{X: 150, Y: 50}, {X: 160, Y: 60}, {X: 170, Y: 70}, {X: 180, Y: 80}, {X: 190, Y: 90},
	}
	samples := make([]mirrorPairSample, len(apath))
	for i, a := range apath {
		b := radarutils.Point{X: 2*100 - a.X, Y: a.Y} // 镜面 x=100 反射
		samples[i] = mirrorPairSample{Ax: a.X, Ay: a.Y, Bx: b.X, By: b.Y, Ts: int64(i * 100)}
	}
	res := evaluateMirrorPair(samples, 7, 8, 0, 0)
	if !res.IsMirror {
		t.Fatalf("vertical mirror should be detected, got %+v", res)
	}
	if res.GhostTrackID != 7 {
		t.Errorf("A (x>100, farther from radar at x=0) should be ghost id=7; got id=%d", res.GhostTrackID)
	}
}

// -- MarkMirrorBounce + 晋升 --

func makeMirrorGrid(t *testing.T) *RoomGrid {
	t.Helper()
	g := NewRoomGrid(500, 500, 10) // 500×500 房间，cellSize 10cm
	for i := range g.Cells {
		g.Cells[i].InRoom = true
		g.Cells[i].InFOV = true
	}
	return g
}

func TestMarkMirrorBounce_SingleCellMarked(t *testing.T) {
	g := makeMirrorGrid(t)
	// (X=100, Y=100) → col=(100-(-250))/10=35, row=10；只命中中心 cell
	g.MarkMirrorBounce(100, 100, 1_000_000)
	col, row := 35, 10
	c := &g.Cells[row*g.Width+col]
	if c.MirrorBounceCount != 1 {
		t.Errorf("center cell MBC: want 1, got %d", c.MirrorBounceCount)
	}
	// 邻居 cells 不应被涂（单 cell 涂法）
	for _, idx := range [][2]int{{34, 10}, {36, 10}, {35, 9}, {35, 11}} {
		nc := &g.Cells[idx[1]*g.Width+idx[0]]
		if nc.MirrorBounceCount != 0 {
			t.Errorf("neighbor (col=%d,row=%d) must not be painted; got MBC=%d", idx[0], idx[1], nc.MirrorBounceCount)
		}
	}
}

func TestMarkMirrorBounce_PromoteAfter3Hits(t *testing.T) {
	g := makeMirrorGrid(t)
	// 同一 cell 命中 3 次（模拟 3 次独立配对 bounce 落在同一 cell）→ 晋升 AreaDeny+SourceLearned
	for i := 0; i < 3; i++ {
		g.MarkMirrorBounce(100, 100, int64(1_000_000+i*1000))
	}
	col, row := 35, 10
	c := &g.Cells[row*g.Width+col]
	if c.MirrorBounceCount != 3 {
		t.Fatalf("MBC should be 3, got %d", c.MirrorBounceCount)
	}
	if c.Belief[0].Type != AreaDeny || c.Belief[0].Source != SourceLearned {
		t.Errorf("expected promoted to AreaDeny+SourceLearned, got type=%d source=%d",
			c.Belief[0].Type, c.Belief[0].Source)
	}
}

// 2 次命中不晋升（边界：阈值 = 3 才升）
func TestMarkMirrorBounce_NoPromoteBelowThreshold(t *testing.T) {
	g := makeMirrorGrid(t)
	for i := 0; i < 2; i++ {
		g.MarkMirrorBounce(100, 100, int64(1_000_000+i*1000))
	}
	c := &g.Cells[10*g.Width+35]
	if c.MirrorBounceCount != 2 {
		t.Errorf("MBC want 2, got %d", c.MirrorBounceCount)
	}
	if c.Belief[0].Source == SourceLearned {
		t.Error("must not promote below threshold")
	}
}

func TestMarkMirrorBounce_SourceHumanNotOverridden(t *testing.T) {
	g := makeMirrorGrid(t)
	col, row := 35, 10
	// 该 cell 已是 layout-config 标的 AreaBed (SourceHuman) — 人工真相不被 runtime 改
	g.Cells[row*g.Width+col].Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	g.Cells[row*g.Width+col].AreaType = AreaBed
	// 累 ≥3 次 mirror
	for i := 0; i < 5; i++ {
		g.MarkMirrorBounce(100, 100, int64(1_000_000+i*1000))
	}
	c := &g.Cells[row*g.Width+col]
	if c.Belief[0].Type != AreaBed || c.Belief[0].Source != SourceHuman {
		t.Errorf("SourceHuman cell must NOT be overridden; got type=%d source=%d",
			c.Belief[0].Type, c.Belief[0].Source)
	}
	// MBC 仍累加（用于审计）
	if c.MirrorBounceCount < 3 {
		t.Errorf("MBC should still accumulate, got %d", c.MirrorBounceCount)
	}
}

func TestMarkMirrorBounce_OutOfGridSilentSkip(t *testing.T) {
	g := makeMirrorGrid(t)
	g.MarkMirrorBounce(99999, 99999, 1_000_000) // 远超 grid 范围
	// 不 panic 即可
	// 抽样几个 cell 确认未被误标
	for i := range g.Cells[:10] {
		if g.Cells[i].MirrorBounceCount != 0 {
			t.Errorf("OOB call must not paint any cell; idx=%d MBC=%d", i, g.Cells[i].MirrorBounceCount)
		}
	}
}

// -- persist round-trip v8 --

func TestPersist_MBC_LMM_RoundTrip(t *testing.T) {
	g := makeMirrorGrid(t)
	for i := 0; i < 3; i++ {
		g.MarkMirrorBounce(100, 100, 1_700_000_000_000)
	}
	snap := EncodeSnapshot(g)
	if snap.SchemaVer != 8 {
		t.Errorf("snapshot schema version should be 8, got %d", snap.SchemaVer)
	}
	// 找命中 cell
	col, row := 35, 10
	idx := row*g.Width + col
	var got *CellSnapshot
	for i := range snap.Cells {
		if snap.Cells[i].I == idx {
			got = &snap.Cells[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("cell idx=%d not in snapshot", idx)
	}
	if got.C == nil || got.C.MBC != 3 || got.C.LMM != 1_700_000_000_000 {
		t.Errorf("MBC/LMM round-trip: got C=%+v, want MBC=3 LMM=1.7e12", got.C)
	}

	// Decode 回新 grid
	g2 := makeMirrorGrid(t)
	if err := DecodeSnapshot(snap, g2); err != nil {
		t.Fatal(err)
	}
	c2 := &g2.Cells[idx]
	if c2.MirrorBounceCount != 3 || c2.LastMirrorMs != 1_700_000_000_000 {
		t.Errorf("decode round-trip: want MBC=3 LMM=1.7e12, got MBC=%d LMM=%d",
			c2.MirrorBounceCount, c2.LastMirrorMs)
	}
	// 晋升过的 Belief 也应回灌
	if c2.Belief[0].Type != AreaDeny || c2.Belief[0].Source != SourceLearned {
		t.Errorf("promoted Belief must round-trip; got type=%d source=%d",
			c2.Belief[0].Type, c2.Belief[0].Source)
	}
}

// -- Decay --

func TestCell_MBC_DecayHalvesOverHalflife(t *testing.T) {
	c := &Cell{MirrorBounceCount: 100}
	p := DefaultDecayParams()
	c.Decay(p.EventSec, p) // 一个半衰期
	if c.MirrorBounceCount < 40 || c.MirrorBounceCount > 60 {
		t.Errorf("after 1 half-life, MBC should be ~50, got %d", c.MirrorBounceCount)
	}
}
