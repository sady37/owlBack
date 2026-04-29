package roomengine

import "testing"

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
