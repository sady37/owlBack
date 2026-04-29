package roomengine

import (
	"testing"
)

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
	if c.Belief[0].Source != SourceHuman || c.Belief[0].Confidence != 95 {
		t.Errorf("Belief[0] should be SourceHuman/95, got %+v", c.Belief[0])
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

// PR-10: BeliefSec 衰减 + 降级
func TestBeliefDecay_HalfLife2Days(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaSit

	p := DefaultDecayParams()
	// 2 天后应衰到 ~50
	c.Decay(2*24*3600, p)
	if c.Belief[0].Confidence < 45 || c.Belief[0].Confidence > 55 {
		t.Errorf("after 2d expect ~50, got %d", c.Belief[0].Confidence)
	}
	if c.Belief[0].Type != AreaSit {
		t.Errorf("Type should remain AreaSit at 50 confidence, got %v", c.Belief[0].Type)
	}
}

func TestBeliefDecay_DowngradeBelow10(t *testing.T) {
	c := &Cell{}
	c.Belief[0] = BeliefState{Type: AreaSit, Confidence: 100, Source: SourceHuman}
	c.AreaType = AreaSit

	p := DefaultDecayParams()
	// 7 天后应衰到 ~10 → 触发降级 → AreaUnknown + Source=Unset
	c.Decay(7*24*3600, p)
	if c.Belief[0].Type != AreaUnknown {
		t.Errorf("after 7d expect AreaUnknown downgrade, got %v conf=%d",
			c.Belief[0].Type, c.Belief[0].Confidence)
	}
	if c.Belief[0].Source != SourceUnset {
		t.Errorf("Source should be Unset after downgrade, got %v", c.Belief[0].Source)
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
