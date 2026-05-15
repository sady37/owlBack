package zonealarm

import (
	"path/filepath"
	"testing"

	"owl-common/alarm"
	"wisefido-sensor/internal/zoneengine"
)

func TestLoadFromFile_DefaultYaml(t *testing.T) {
	// 烟雾测试：从仓库的 config/zone_alarm.yaml 加载并校验规则数 + 字段解析。
	path := filepath.Join("..", "..", "config", "zone_alarm.yaml")
	rules, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile %s: %v", path, err)
	}
	if len(rules) != 4 {
		t.Fatalf("want 4 rules in default yaml, got %d", len(rules))
	}
	wants := []string{alarm.Stay, alarm.LeftBed, alarm.NightAbsence, alarm.BedNightAbsence}
	for i, want := range wants {
		if rules[i].AlarmType != want {
			t.Errorf("rule[%d] AlarmType: want %s, got %s", i, want, rules[i].AlarmType)
		}
	}
	// 抽查 NightAbsence 的 count_gt_zero 解析
	var night Rule
	for _, r := range rules {
		if r.AlarmType == alarm.NightAbsence {
			night = r
			break
		}
	}
	hasCountGt := false
	for _, c := range night.Cancels {
		if c.CountGtZero {
			hasCountGt = true
		}
	}
	if !hasCountGt {
		t.Errorf("NightAbsence should have count_gt_zero cancel trigger")
	}
	if night.TimeWindow == nil || night.TimeWindow.StartH != 21 {
		t.Errorf("NightAbsence TimeWindow not parsed: %+v", night.TimeWindow)
	}
}

func TestDefaultRules_Resolved(t *testing.T) {
	rules := DefaultRules()
	for i, r := range rules {
		if r.AlarmType == "" {
			t.Errorf("rule[%d] AlarmType empty", i)
		}
		// arm_zone/status 字符串应已 resolved 为 enum
		switch r.ArmZone {
		case zoneengine.ZoneTypeBed, zoneengine.ZoneTypeRoom, zoneengine.ZoneTypeBathroom:
			// ok
		default:
			t.Errorf("rule[%d] ArmZone not resolved: %v", i, r.ArmZone)
		}
		switch r.ArmStatus {
		case zoneengine.StatusOccupied, zoneengine.StatusVacant, zoneengine.StatusLeaving:
		default:
			t.Errorf("rule[%d] ArmStatus not resolved: %v", i, r.ArmStatus)
		}
	}
}

func TestParseZoneType_Errors(t *testing.T) {
	cases := []string{"", "kitchen", "BED2"}
	for _, c := range cases {
		if _, err := parseZoneType(c); err == nil {
			t.Errorf("parseZoneType(%q) should fail", c)
		}
	}
}

func TestParseZoneStatus_Errors(t *testing.T) {
	cases := []string{"", "active", "VACANT2"}
	for _, c := range cases {
		if _, err := parseZoneStatus(c); err == nil {
			t.Errorf("parseZoneStatus(%q) should fail", c)
		}
	}
}
