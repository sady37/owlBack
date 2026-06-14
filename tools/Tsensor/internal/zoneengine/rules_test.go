package zoneengine

import (
	"testing"
)

// TestDefaultRulesShape 烟测 DefaultRules 字段都填了 + BedSizeBucket 分类正确。
func TestDefaultRulesShape(t *testing.T) {
	r := DefaultRules()

	if r.Bed.Enter["sleepace"].Strength != 90 {
		t.Errorf("bed sleepace enter strength: got %d, want 90", r.Bed.Enter["sleepace"].Strength)
	}
	if r.Bed.Enter["radar"].Strength != 80 {
		t.Errorf("bed radar enter strength: got %d, want 80", r.Bed.Enter["radar"].Strength)
	}
	if r.Bed.Sustain.HRRRPresentAny.Strength != 80 {
		t.Errorf("bed hr_rr sustain strength: got %d, want 80", r.Bed.Sustain.HRRRPresentAny.Strength)
	}
	if r.Bed.Sustain.RecentEnterBonus.WindowSec != 60 {
		t.Errorf("bed recent_enter_bonus window_sec: got %d, want 60", r.Bed.Sustain.RecentEnterBonus.WindowSec)
	}
	if !r.Bed.Leave.SizeDependent {
		t.Errorf("bed leave should be SizeDependent=true")
	}
	if r.Bed.Leave.SmallBed.Sources["sleepace"].Strength != 80 {
		t.Errorf("small_bed sleepace strength: got %d, want 80",
			r.Bed.Leave.SmallBed.Sources["sleepace"].Strength)
	}
	if r.Bed.Leave.LargeBed.Sources["sleepace"].Strength != 70 {
		t.Errorf("large_bed sleepace strength (大床覆盖不全降权): got %d, want 70",
			r.Bed.Leave.LargeBed.Sources["sleepace"].Strength)
	}

	if r.Feedback.SelfContradiction.WindowSec != 15 {
		t.Errorf("self_contradiction window_sec: got %d, want 15", r.Feedback.SelfContradiction.WindowSec)
	}
	if r.Feedback.SubsetInvariant.Mode != "lift_parent" {
		t.Errorf("subset_invariant mode: got %s, want lift_parent", r.Feedback.SubsetInvariant.Mode)
	}
}

// TestBedSizeBucket 床型 → bucket 分类（standard 是 default）。
func TestBedSizeBucket(t *testing.T) {
	cases := []struct {
		kind string
		want string
	}{
		{"", "small"}, // 空字符串 = default = standard 走 small
		{"standard", "small"},
		{"hospital", "small"},
		{"twin", "small"},
		{"unknown_kind", "small"}, // 未识别都走 small
		{"full", "large"},
		{"queen", "large"},
		{"king", "large"},
		{"california_king", "large"},
	}
	for _, c := range cases {
		got := BedSizeBucket(c.kind)
		if got != c.want {
			t.Errorf("BedSizeBucket(%q) = %q, want %q", c.kind, got, c.want)
		}
	}
}

// TestLoadRulesFromYaml 验证 yaml 文件 ↔ Go struct 序列化往返；
// 用 config/zone_rules.yaml 真实文件做 fixture（路径相对 sensor module root）。
func TestLoadRulesFromYaml(t *testing.T) {
	r, err := LoadRulesFromFile("../../config/zone_rules.yaml")
	if err != nil {
		t.Fatalf("LoadRulesFromFile: %v", err)
	}
	if r.Bed.Enter["sleepace"].Strength != 90 {
		t.Errorf("yaml bed sleepace enter strength: got %d, want 90",
			r.Bed.Enter["sleepace"].Strength)
	}
	if r.Bed.Leave.SmallBed.Sources["radar"].Strength != 70 {
		t.Errorf("yaml small_bed radar leave: got %d, want 70",
			r.Bed.Leave.SmallBed.Sources["radar"].Strength)
	}
	if !r.Feedback.SubsetInvariant.Enabled {
		t.Errorf("yaml subset_invariant should be enabled")
	}
}
