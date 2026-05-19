// bathroom_ghost_test.go — sensor_v2 PR-6 BathroomGhostAdjudicator 单测
//
// 4 条规则 + Rule 4 fallback 验收：
//   - Rule 1: 远离 EnterTarget 出生 → ghost
//   - Rule 2: track count > BathroomCount → 远的 ghost
//   - Rule 3: BathroomCount undercount → 修正 count，不翻 verdict
//   - Rule 5: 紧贴已存在 track (< 60cm) → ghost
//   - Rule 4 fallback: entry 全在盲区 → 仅启用 rule 5
//
// 设计：测试用 200×200cm 的 grid，CellSize=10，entry cell 放在 (col=0,row=10) ≈ (-100, 100) 位置；
// 真人 track 放在 (X=-95, Y=105) 紧贴 entry；远离 entry 的 ghost 放在 (X=50, Y=150) 中心。

package roomengine

import (
	"testing"

	"go.uber.org/zap"
)

const (
	tgRoom  = "fd00:0:3:222:80::/128"
	tgSuite = "fd00:0:3:222::/80"
	tgRes   = "fd00:0:3:222:ff00:01:1:1/128"
)

// makeBathroomGrid 建一个 200×200 cm grid，entry cell 在 col=0,row=10（cell 中心 ≈ (-95, 105)）。
// inFOVEntries=true 时把 entry cell InFOV=true，否则放在盲区（用于 Rule 4 测试）。
func makeBathroomGrid(t *testing.T, inFOVEntries bool) *RoomGrid {
	t.Helper()
	g := NewRoomGrid(200, 200, 10)
	// 标记入口 cell — col=0 row=10
	if c := &g.Cells[10*g.Width+0]; true {
		c.EnterTarget = "bathroom"
		c.InFOV = inFOVEntries
	}
	// 房间内其它 cells 全开 InFOV，避免被其他 Rule 误判
	for i := range g.Cells {
		if g.Cells[i].EnterTarget == "bathroom" {
			continue
		}
		g.Cells[i].InFOV = true
		g.Cells[i].InRoom = true
	}
	return g
}

// makeBathroomAdj 标准 setup：grid + census + adjudicator
func makeBathroomAdj(t *testing.T, grid *RoomGrid, nowMs int64) (*BathroomGhostAdjudicator, *SuiteCensusManager) {
	t.Helper()
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	// 升 resident（gate 跨 0 边界 anchor flip 会用；ghost rules 也依赖 census）
	if _, ok := m.UpdatePersonFromTrack(tgSuite, tgRes, 1, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	a := NewBathroomGhostAdjudicator(
		m,
		func(_ string) *RoomGrid { return grid },
		func(_ string) string { return tgSuite },
		zap.NewNop(),
	)
	return a, m
}

// 入口 cell 中心画布坐标（基于 NewRoomGrid 200×200 + col=0 row=10）。
// OriginX = -100, OriginY = 0；ToCanvas(0,10) = (-100+5, 0+105) = (-95, 105)。
const (
	entryX = -95
	entryY = 105
)

// 规则 1：远离 entry 新出生 → ghost
func TestBathroomGhost_Rule1_BirthFarFromEntry(t *testing.T) {
	nowMs := int64(1_000_000)
	g := makeBathroomGrid(t, true)
	a, _ := makeBathroomAdj(t, g, nowMs)

	// 新 track 出生在 (50, 150)，距 entry ≈ √(145² + 45²) ≈ 152cm > 30cm threshold
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: 50, Y: 150, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)
	if len(deltas) != 1 || deltas[0].TrackID != 7 {
		t.Fatalf("expected 1 ghost delta on track 7, got %+v", deltas)
	}
	if deltas[0].NewVerdict == nil || *deltas[0].NewVerdict != VerdictGhost {
		t.Errorf("verdict want Ghost, got %v", deltas[0].NewVerdict)
	}
	if deltas[0].PenaltyDelta != 100 {
		t.Errorf("penalty want 100, got %d", deltas[0].PenaltyDelta)
	}
}

// 规则 1：紧贴 entry 新出生 → 不 ghost
func TestBathroomGhost_Rule1_BirthNearEntry_NoGhost(t *testing.T) {
	nowMs := int64(2_000_000)
	g := makeBathroomGrid(t, true)
	a, _ := makeBathroomAdj(t, g, nowMs)

	// 新 track 出生在 (-90, 100)，距 entry ≈ √(5² + 5²) ≈ 7cm < 30cm threshold
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)
	if len(deltas) != 0 {
		t.Fatalf("near-entry birth should not be ghost, got %+v", deltas)
	}
}

// 规则 5：新出生紧贴已存在 track → ghost（即使靠近 entry 也 ghost）
func TestBathroomGhost_Rule5_SplitGhostAdjacent(t *testing.T) {
	nowMs := int64(3_000_000)
	g := makeBathroomGrid(t, true)
	a, _ := makeBathroomAdj(t, g, nowMs)

	// 帧 1：建立 track 7 = real
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
	}, nowMs)

	// 帧 2：track 7 仍在，新出生 track 8 紧贴 track 7（距离 ≈ 14cm < 60cm）
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: -80, Y: 110, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)

	found8Ghost := false
	for _, d := range deltas {
		if d.TrackID == 8 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			found8Ghost = true
		}
	}
	if !found8Ghost {
		t.Errorf("track 8 should be ghost by Rule 5 (split-ghost adjacent), got deltas %+v", deltas)
	}
}

// 规则 5：紧贴 ghost track（不是 real）→ 不触发（ghost 不能作反射源）
// 隔离要求：BathroomCount 设足够大避免 Rule 2 干扰
func TestBathroomGhost_Rule5_NotTriggeredByGhostSource(t *testing.T) {
	nowMs := int64(4_000_000)
	g := makeBathroomGrid(t, true)
	a, m := makeBathroomAdj(t, g, nowMs)

	// 提前把 BathroomCount 顶到 2（避免 Rule 2 occupancy excess 干扰本测）
	m.AdjustBathroomCount(tgSuite, 2, nowMs)

	// 帧 1：track 7 = ghost
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictGhost},
	}, nowMs)

	// 帧 2：track 8 新出生 紧贴 track 7（但 7 是 ghost，不能作反射源）
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictGhost},
		{TrackID: 8, RoomID: tgRoom, X: -80, Y: 110, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)

	// track 8 (-80,110) 距 entry ~22cm < 30cm → Rule 1 不触发
	// Rule 5 应跳过（reflection source 是 ghost）
	// BathroomCount=2 == live count=2 → Rule 2 不触发
	// 综合：track 8 应该完全无 delta
	for _, d := range deltas {
		if d.TrackID == 8 {
			t.Errorf("track 8 should not be ghost (source is ghost, near entry, count OK), got %+v", d)
		}
	}
}

// 规则 2：track count > BathroomCount → 远离 entry 的多余 ghost
func TestBathroomGhost_Rule2_ExcessOverBathroomCount(t *testing.T) {
	nowMs := int64(5_000_000)
	g := makeBathroomGrid(t, true)
	a, m := makeBathroomAdj(t, g, nowMs)

	// 模拟 PR-5 BathroomGate 已记 1 人入 bathroom
	m.AdjustBathroomCount(tgSuite, 1, nowMs)

	// 帧 1：建立 2 个 real track（gate count 滞后于实际）
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: 50, Y: 150, Verdict: VerdictReal},
	}, nowMs)

	// 帧 2：同 2 个 track（不是新出生 → Rule 1/5 不动），但 live count=2 > BathroomCount=1
	// 注：Rule 3 期望 ≥ ceil(2/3) = 1，已满足，不修正 count
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: 50, Y: 150, Verdict: VerdictReal},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)

	// 远离 entry 的 track 8（dist ≈ 152cm）应被 ghost；track 7（dist ≈ 7cm）应保留
	found8Ghost := false
	for _, d := range deltas {
		if d.TrackID == 8 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			found8Ghost = true
		}
		if d.TrackID == 7 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			t.Errorf("track 7 (near entry) should not be ghost by Rule 2, got %+v", d)
		}
	}
	if !found8Ghost {
		t.Errorf("track 8 (far from entry) should be ghost by Rule 2, deltas=%+v", deltas)
	}
}

// 规则 3：BathroomCount < expected_min → 修正 count，verdict 不变
func TestBathroomGhost_Rule3_UndercountCorrection(t *testing.T) {
	nowMs := int64(6_000_000)
	g := makeBathroomGrid(t, true)
	a, m := makeBathroomAdj(t, g, nowMs)

	// gate 完全漏报（BathroomCount=0），但 bathroom 内有 4 track
	// expected_min = ceil(4/3) = 2
	// → 修正 count 从 0 到 2

	// 4 个 real track 都紧贴 entry（避免触发 Rule 1）
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: -85, Y: 105, Verdict: VerdictReal},
		{TrackID: 9, RoomID: tgRoom, X: -90, Y: 110, Verdict: VerdictReal},
		{TrackID: 10, RoomID: tgRoom, X: -85, Y: 100, Verdict: VerdictReal},
	}, nowMs)

	// 2nd frame: track count=4, BathroomCount=0 → Rule 3 修正 count = 2
	// 注：Rule 5 split-ghost 会触发（4 个 track 互相距离 < 60cm），先有新出生 ghost。
	// 但 Rule 3 修正 count 不依赖 verdict — 修正发生在 Rule 5 之后，但读取 count 在 Rule 3 内部
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: -85, Y: 105, Verdict: VerdictReal},
		{TrackID: 9, RoomID: tgRoom, X: -90, Y: 110, Verdict: VerdictReal},
		{TrackID: 10, RoomID: tgRoom, X: -85, Y: 100, Verdict: VerdictReal},
	}, nowMs+1000)

	if got := m.GetBathroomCount(tgSuite); got < 2 {
		t.Errorf("Rule 3 should correct BathroomCount to >= ceil(4/3)=2, got %d", got)
	}
}

// 规则 4 fallback：entry 全在盲区 → 跳过 Rule 1/2/3，仅 Rule 5 生效
func TestBathroomGhost_Rule4_EntryInBlindSpot(t *testing.T) {
	nowMs := int64(7_000_000)
	g := makeBathroomGrid(t, false) // entry InFOV=false
	a, m := makeBathroomAdj(t, g, nowMs)
	_ = m

	// 帧 1：建立 track 7 = real
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
	}, nowMs)

	// 帧 2：track 8 远离 entry 出生 — Rule 1 在 fallback 模式下应跳过（entry 不可信）
	//        track 9 紧贴 track 7 — Rule 5 仍生效
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tgRoom, X: 50, Y: 150, Verdict: VerdictPending},
		{TrackID: 9, RoomID: tgRoom, X: -80, Y: 110, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)

	// track 8: 应 NOT 被 Rule 1 ghost（fallback 关闭）
	for _, d := range deltas {
		if d.TrackID == 8 {
			t.Errorf("Rule 1 should be disabled in fallback mode, track 8 got %+v", d)
		}
	}
	// track 9: 应被 Rule 5 ghost（仍生效）
	found9Ghost := false
	for _, d := range deltas {
		if d.TrackID == 9 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			found9Ghost = true
		}
	}
	if !found9Ghost {
		t.Errorf("Rule 5 should remain active in fallback, track 9 not ghosted; deltas=%+v", deltas)
	}
}

// 空 bases 不 panic 不返 delta
func TestBathroomGhost_EmptyBasesNoOp(t *testing.T) {
	nowMs := int64(8_000_000)
	g := makeBathroomGrid(t, true)
	a, _ := makeBathroomAdj(t, g, nowMs)

	deltas := a.Adjudicate(nil, nowMs)
	if len(deltas) != 0 {
		t.Errorf("empty bases should return no deltas, got %+v", deltas)
	}
}

// PR-6.1 #4: hasDisconnectedEntries 连通性判定
func TestHasDisconnectedEntries(t *testing.T) {
	// 单 cell — 不算多片
	one := []bathroomEntryPoint{{cx: 0, cy: 0, inFOV: true}}
	if hasDisconnectedEntries(one, 10) {
		t.Errorf("single entry cell should not be flagged")
	}

	// 2 个相邻 cell（距 10cm）— 连续
	adjacent := []bathroomEntryPoint{{cx: 0, cy: 0}, {cx: 10, cy: 0}}
	if hasDisconnectedEntries(adjacent, 10) {
		t.Errorf("two adjacent cells (10cm apart) should be connected")
	}

	// 2 个对角邻接 cell（距 ~14cm）— 连续（八方向邻接）
	diag := []bathroomEntryPoint{{cx: 0, cy: 0}, {cx: 10, cy: 10}}
	if hasDisconnectedEntries(diag, 10) {
		t.Errorf("two diagonal cells should be connected via 八方向")
	}

	// 2 片：每片 1 cell，距 100cm — 多片
	split := []bathroomEntryPoint{{cx: 0, cy: 0}, {cx: 100, cy: 0}}
	if !hasDisconnectedEntries(split, 10) {
		t.Errorf("two cells 100cm apart should be flagged as disconnected")
	}

	// 3 cells 一字排开（0, 10, 20）— 连续
	line := []bathroomEntryPoint{{cx: 0, cy: 0}, {cx: 10, cy: 0}, {cx: 20, cy: 0}}
	if hasDisconnectedEntries(line, 10) {
		t.Errorf("three cells in a row should be connected")
	}

	// 2 片：每片 2 cells，片间距 100cm — 多片
	twoBlobs := []bathroomEntryPoint{
		{cx: 0, cy: 0}, {cx: 10, cy: 0},
		{cx: 100, cy: 0}, {cx: 110, cy: 0},
	}
	if !hasDisconnectedEntries(twoBlobs, 10) {
		t.Errorf("two 2-cell blobs 100cm apart should be flagged")
	}
}

// 无 grid 时安全降级
func TestBathroomGhost_NoGridNoOp(t *testing.T) {
	nowMs := int64(9_000_000)
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	a := NewBathroomGhostAdjudicator(
		m,
		func(_ string) *RoomGrid { return nil }, // 故意返 nil
		func(_ string) string { return tgSuite },
		zap.NewNop(),
	)
	deltas := a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: 50, Y: 150},
	}, nowMs)
	if len(deltas) != 0 {
		t.Errorf("no-grid should return no deltas, got %+v", deltas)
	}
}

// Anchored guard 集成：Rule 1 命中 Anchored track 时，applyVerdictDeltas 拒绝降级
// （PR-4 Q4 守卫的 end-to-end 验证）
func TestBathroomGhost_AnchoredTrack_VerdictKept(t *testing.T) {
	nowMs := int64(10_000_000)
	g := makeBathroomGrid(t, true)
	a, _ := makeBathroomAdj(t, g, nowMs)

	// 新出生 + Anchored verdict（实际场景：LongSurvival 锚定）
	// 远离 entry → Rule 1 试图判 ghost
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tgRoom, X: 50, Y: 150, Verdict: VerdictAnchored},
	}
	deltas := a.Adjudicate(bases, nowMs+1000)
	if len(deltas) != 1 {
		t.Fatalf("expect 1 delta from adjudicator (verdict change suggested), got %d", len(deltas))
	}

	// 模拟 engine.applyVerdictDeltas：Anchored guard 应该拒绝降级
	e := &Engine{logger: zap.NewNop()}
	statuses := []*TrackStatus{
		{TrackID: 7, Verdict: VerdictAnchored, GhostPenalty: 0},
	}
	e.applyVerdictDeltas(statuses, deltas)

	if statuses[0].Verdict != VerdictAnchored {
		t.Errorf("Anchored must survive Rule 1 ghost suggestion, got %v", statuses[0].Verdict)
	}
	// penalty 仍累加供审计
	if statuses[0].GhostPenalty != 100 {
		t.Errorf("penalty should still accumulate to 100, got %d", statuses[0].GhostPenalty)
	}
}
