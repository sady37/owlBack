package roomengine

import (
	"testing"
)

// TestClearNonHumanLearnedZone verified 真摔擦非 FE 画的 rest/deny，保 FE 画 bed。
func TestClearNonHumanLearnedZone(t *testing.T) {
	// SourceFeedback sit → 擦
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 95, Source: SourceFeedback}
	c.AreaType = AreaSit
	if !c.ClearNonHumanLearnedZone() || c.Belief[0].Type != AreaUnknown || c.AreaType != AreaUnknown {
		t.Errorf("SourceFeedback AreaSit 应被擦回 Unknown, got %+v / area=%v", c.Belief[0], c.AreaType)
	}

	// SourceLearned deny(金属)→ 擦（真摔证明该处不是 deny）
	c = &Cell{}
	c.Belief[0] = BeliefState{Type: AreaDeny, Confidence: 70, Source: SourceLearned}
	c.AreaType = AreaDeny
	if !c.ClearNonHumanLearnedZone() || c.Belief[0].Type != AreaUnknown {
		t.Errorf("SourceLearned AreaDeny 应被擦回 Unknown, got %+v", c.Belief[0])
	}

	// SourceHuman FE 画 bed → 不擦（神圣，仅记录+人工复审）
	c = &Cell{}
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaBed
	if c.ClearNonHumanLearnedZone() || c.Belief[0].Type != AreaBed || c.Belief[0].Source != SourceHuman {
		t.Errorf("SourceHuman AreaBed 不该被擦, got %+v", c.Belief[0])
	}
}

// TestMarkRestZoneByFeedback PR-7.1：人类反馈即时锁 AreaType
func TestMarkRestZoneByFeedback_LocksAreaSit(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaUnknown, Confidence: 50, Source: SourceLearned}
	c.AreaType = AreaUnknown

	if !c.MarkRestZoneByFeedback(AreaSit) {
		t.Errorf("expected MarkRestZoneByFeedback to lock; got false")
	}
	if c.AreaType != AreaSit {
		t.Errorf("AreaType expected AreaSit, got %v", c.AreaType)
	}
	if c.Belief[0].Source != SourceFeedback || c.Belief[0].Confidence != 95 {
		t.Errorf("Belief[0] should be SourceFeedback/95, got %+v", c.Belief[0])
	}
}

func TestMarkRestZoneByFeedback_PreservesLayoutBed(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	c.AreaType = AreaBed

	if c.MarkRestZoneByFeedback(AreaSit) {
		t.Errorf("should NOT overwrite layout AreaBed; got true")
	}
	if c.AreaType != AreaBed {
		t.Errorf("AreaType should stay AreaBed, got %v", c.AreaType)
	}
}

func TestMarkRestZoneByFeedback_PreservesLayoutDenyAndEnter(t *testing.T) {
	for _, kind := range []AreaType{AreaDeny, AreaEnter, AreaToilet, AreaShower} {
		c := &Cell{}
		c.Belief[0] = BeliefState{Type: kind, Confidence: 99, Source: SourceHuman}
		c.AreaType = kind
		if c.MarkRestZoneByFeedback(AreaSit) {
			t.Errorf("should NOT overwrite layout %v; got true", kind)
		}
	}
}

func TestMarkRestZoneByFeedback_NoDoubleSet(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 95, Source: SourceHuman}
	c.AreaType = AreaSit
	if c.MarkRestZoneByFeedback(AreaSit) {
		t.Errorf("should be no-op when already SourceHuman+AreaSit; got true")
	}
}

// PR-11: BeliefHalfLifeByType 分档衰减 + 降级
// AreaSit: 半衰期 2.1d → 2.1 天衰到 50 ✓
func TestBeliefDecay_AreaSit_HalfLife(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaSit

	p := DefaultDecayParams()
	c.Decay(2.1*24*3600, p)
	if c.Belief[0].Confidence < 45 || c.Belief[0].Confidence > 55 {
		t.Errorf("AreaSit after 2.1d expect ~50, got %d", c.Belief[0].Confidence)
	}
}

// AreaSit 7.5 天衰到 < 10 → 降级（7d 边界精度，给 0.5 margin）
func TestBeliefDecay_AreaSit_DowngradeAt7d(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaSit

	p := DefaultDecayParams()
	c.Decay(7.5*24*3600, p)
	if c.Belief[0].Type != AreaUnknown {
		t.Errorf("AreaSit 7.5d expect downgrade to AreaUnknown, got %v conf=%d",
			c.Belief[0].Type, c.Belief[0].Confidence)
	}
}

// AreaBed: 9.5 天衰到 < 10 → 降级
func TestBeliefDecay_AreaBed_DowngradeAt9d(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaBed

	p := DefaultDecayParams()
	c.Decay(9.5*24*3600, p)
	if c.Belief[0].Type != AreaUnknown {
		t.Errorf("AreaBed 9.5d expect downgrade, got %v conf=%d",
			c.Belief[0].Type, c.Belief[0].Confidence)
	}
}

// AreaToilet: 60 天衰到 10（远长于 AreaSit/AreaBed）
func TestBeliefDecay_AreaToilet_StillStrongAt9d(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaToilet, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaToilet

	p := DefaultDecayParams()
	c.Decay(9*24*3600, p)
	// 9 天对 toilet 来说还很强（60d 才衰完）
	if c.Belief[0].Type != AreaToilet {
		t.Errorf("AreaToilet 9d should remain AreaToilet, got %v", c.Belief[0].Type)
	}
	if c.Belief[0].Confidence < 50 {
		t.Errorf("AreaToilet 9d expect confidence still > 50, got %d", c.Belief[0].Confidence)
	}
}

// PR-10: boostNeighborSameType 4 邻居 ×1.3
// 用 200×200 grid + 中心 (10,50) 避免边缘越界
func TestBoostNeighborSameType_4Neighbors(t *testing.T) {
	g := NewRoomGrid(200, 200, 10)
	// 中心 (10, 50) + 4 邻居 (0,50)/(20,50)/(10,40)/(10,60) — 都在 grid 内
	for _, p := range [][2]int{{10, 50}, {0, 50}, {20, 50}, {10, 40}, {10, 60}} {
		c := g.CellAt(p[0], p[1])
		if c == nil {
			t.Fatalf("cell at %v nil", p)
		}
		c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 50, Source: SourceHuman}
		c.AreaType = AreaSit
	}
	// 对角邻居 (0, 40) 应不被增强
	corner := g.CellAt(0, 40)
	corner.Belief[0] = BeliefState{Type: AreaSit, Confidence: 50, Source: SourceHuman}

	g.boostNeighborSameType(10, 50, AreaSit, 12345)

	for _, p := range [][2]int{{0, 50}, {20, 50}, {10, 40}, {10, 60}} {
		c := g.CellAt(p[0], p[1])
		if c.Belief[0].Confidence != 65 {
			t.Errorf("neighbor %v expect 65, got %d", p, c.Belief[0].Confidence)
		}
	}
	if corner.Belief[0].Confidence != 50 {
		t.Errorf("diagonal neighbor should stay 50, got %d", corner.Belief[0].Confidence)
	}
	center := g.CellAt(10, 50)
	if center.Belief[0].Confidence != 50 {
		t.Errorf("center cell self should stay 50, got %d", center.Belief[0].Confidence)
	}
}

func TestBoostNeighborSameType_NoTypeMismatch(t *testing.T) {
	g := NewRoomGrid(200, 200, 10)
	c := g.CellAt(10, 50)
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 50, Source: SourceHuman}

	bedNeighbor := g.CellAt(0, 50)
	bedNeighbor.Belief[0] = BeliefState{Type: AreaBed, Confidence: 50, Source: SourceHuman}

	g.boostNeighborSameType(10, 50, AreaSit, 12345)

	if bedNeighbor.Belief[0].Confidence != 50 {
		t.Errorf("Bed neighbor should not be boosted by Sit target, got %d", bedNeighbor.Belief[0].Confidence)
	}
}

func TestBoostNeighborSameType_CapAt100(t *testing.T) {
	g := NewRoomGrid(200, 200, 10)
	c := g.CellAt(10, 50)
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 50, Source: SourceHuman}

	n := g.CellAt(0, 50)
	n.Belief[0] = BeliefState{Type: AreaSit, Confidence: 90, Source: SourceHuman}

	g.boostNeighborSameType(10, 50, AreaSit, 12345)
	// 90 * 1.3 = 117 → cap 100
	if n.Belief[0].Confidence != 100 {
		t.Errorf("expect cap 100, got %d", n.Belief[0].Confidence)
	}
}

// 测试 PR-6 notes parser
func TestParseConditions_FalseAlarmReason(t *testing.T) {
	notes := `False Alarm Reason:
☑ Sit on Chair / behind Chair

[2026-04-29 01:27:00] admin (event:abc-1234)`

	pc := parseConditions(notes)
	if !pc.HasFalseAlarmBlock {
		t.Errorf("expected HasFalseAlarmBlock=true")
	}
	if !pc.FAChair {
		t.Errorf("expected FAChair=true")
	}
	if pc.FASofa || pc.FAGhostInterference || pc.FAWheelchair || pc.FAOther {
		t.Errorf("only FAChair should be true; got %+v", pc)
	}
	if !pc.AnyFASelected() {
		t.Errorf("AnyFASelected should be true")
	}
}

func TestParseConditions_FAGhostInterference(t *testing.T) {
	notes := `False Alarm Reason:
☑ NoPerson / Electric / AC Interference
`
	pc := parseConditions(notes)
	if !pc.FAGhostInterference {
		t.Errorf("expected FAGhostInterference=true")
	}
	if pc.FAChair {
		t.Errorf("FAChair should be false")
	}
}

func TestParseConditions_FAMultipleSelected(t *testing.T) {
	notes := `False Alarm Reason:
☑ Sit on Chair / behind Chair
☑ Sit in Wheelchair
`
	pc := parseConditions(notes)
	if !pc.FAChair || !pc.FAWheelchair {
		t.Errorf("expected both Chair + Wheelchair selected: %+v", pc)
	}
}

func TestParseConditions_VerifiedFall(t *testing.T) {
	notes := `Observed Conditions:
☑ Fall / Sitting on the Ground
☑ Visible Bleeding

[2026-04-26 22:15:36] admin (event:xyz)`

	pc := parseConditions(notes)
	if !pc.HasObservedBlock {
		t.Errorf("expected HasObservedBlock=true")
	}
	if !pc.OCFall {
		t.Errorf("expected OCFall=true")
	}
	if !pc.OCBleeding {
		t.Errorf("expected OCBleeding=true")
	}
	if pc.OCAwake || pc.OCVerbal || pc.OCUnresponsive {
		t.Errorf("only Fall + Bleeding expected: %+v", pc)
	}
}

func TestParseConditions_EmptyNotes(t *testing.T) {
	pc := parseConditions("")
	if pc.AnyFASelected() || pc.HasObservedBlock || pc.HasFalseAlarmBlock {
		t.Errorf("empty notes should yield zero ParsedConditions: %+v", pc)
	}
}

func TestParseConditions_OtherOnlyInFalseAlarmBlock(t *testing.T) {
	// Other 关键字仅在 FA block 才算 FAOther（避免 verified block 中 "Other something" 误中）
	notes := `Observed Conditions:
☑ Fall / Sitting on the Ground
Other freeform comment`
	pc := parseConditions(notes)
	if pc.FAOther {
		t.Errorf("FAOther should be false when block is Observed Conditions; got %+v", pc)
	}

	notes2 := `False Alarm Reason:
☑ Other`
	pc2 := parseConditions(notes2)
	if !pc2.FAOther {
		t.Errorf("FAOther should be true in FA block; got %+v", pc2)
	}
}
