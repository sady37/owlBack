package zonealarm

import (
	"path/filepath"
	"testing"

	"owl-common/alarm"
)

func TestLoadFromFile_DefaultYaml(t *testing.T) {
	// 烟雾测试：从仓库的 config/zone_alarm.yaml 加载并校验规则数 + 字段解析。
	path := filepath.Join("..", "..", "config", "zone_alarm.yaml")
	rules, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile %s: %v", path, err)
	}
	if len(rules) != 3 {
		t.Fatalf("want 3 rules in default yaml, got %d", len(rules))
	}
	wants := []string{alarm.Stay, alarm.LeftBed, alarm.NightAbsence}
	for i, want := range wants {
		if rules[i].AlarmType != want {
			t.Errorf("rule[%d] AlarmType: want %s, got %s", i, want, rules[i].AlarmType)
		}
	}

	// 抽查 Stay：anchor=bathroom AloneContinuousMin, KeepCounting Self=alone, 双档阈值
	stay := findRule(rules, alarm.Stay)
	if stay == nil {
		t.Fatal("Stay rule not found")
	}
	if stay.Anchor != EntityBathroom || stay.AnchorField != AnchorAloneContinuousMin {
		t.Errorf("Stay anchor: got (%v, %s)", stay.Anchor, stay.AnchorField)
	}
	if len(stay.KeepCounting) != 1 || !stay.KeepCounting[0].UsesSelf || stay.KeepCounting[0].State != StateAlone {
		t.Errorf("Stay KeepCounting unexpected: %+v", stay.KeepCounting)
	}
	if stay.ThresholdDaySec != 2700 || stay.ThresholdNightSec != 1800 {
		t.Errorf("Stay thresholds: got day=%d night=%d", stay.ThresholdDaySec, stay.ThresholdNightSec)
	}

	// 抽查 LeftBed：Self=vacant + peer room=vacant
	leftBed := findRule(rules, alarm.LeftBed)
	if leftBed == nil {
		t.Fatal("LeftBed rule not found")
	}
	if leftBed.Anchor != EntityBed || leftBed.AnchorField != AnchorBedStatusTs {
		t.Errorf("LeftBed anchor: got (%v, %s)", leftBed.Anchor, leftBed.AnchorField)
	}
	if len(leftBed.KeepCounting) != 2 {
		t.Errorf("LeftBed should have 2 conditions, got %d", len(leftBed.KeepCounting))
	}

	// 抽查 NightAbsence：NightOnly 21-7 + peer bed=vacant
	night := findRule(rules, alarm.NightAbsence)
	if night == nil {
		t.Fatal("NightAbsence rule not found")
	}
	if night.NightOnly == nil || night.NightOnly.StartH != 21 || night.NightOnly.EndH != 7 {
		t.Errorf("NightAbsence NightOnly: %+v", night.NightOnly)
	}
}

func TestDefaultRules_Resolved(t *testing.T) {
	rules := DefaultRules()
	if len(rules) != 3 {
		t.Fatalf("want 3 default rules, got %d", len(rules))
	}
	for i, r := range rules {
		if r.AlarmType == "" {
			t.Errorf("rule[%d] AlarmType empty", i)
		}
		switch r.Anchor {
		case EntityBed, EntityRoom, EntityBathroom:
		default:
			t.Errorf("rule[%d] Anchor not resolved: %v", i, r.Anchor)
		}
		if r.AnchorField == "" {
			t.Errorf("rule[%d] AnchorField empty", i)
		}
		if r.ThresholdDaySec <= 0 {
			t.Errorf("rule[%d] ThresholdDaySec must > 0", i)
		}
		for j, c := range r.KeepCounting {
			switch c.Entity {
			case EntityBed, EntityRoom, EntityBathroom:
			default:
				t.Errorf("rule[%d].KeepCounting[%d] Entity not resolved: %v", i, j, c.Entity)
			}
		}
	}
}

func TestParseEntityKind_Errors(t *testing.T) {
	cases := []string{"", "kitchen", "BED2"}
	for _, c := range cases {
		if _, err := parseEntityKind(c); err == nil {
			t.Errorf("parseEntityKind(%q) should fail", c)
		}
	}
}

func TestParseEntityState_Errors(t *testing.T) {
	cases := []string{"", "active", "VACANT2"}
	for _, c := range cases {
		if _, err := parseEntityState(c); err == nil {
			t.Errorf("parseEntityState(%q) should fail", c)
		}
	}
}

func findRule(rules []Rule, alarmType string) *Rule {
	for i := range rules {
		if rules[i].AlarmType == alarmType {
			return &rules[i]
		}
	}
	return nil
}
