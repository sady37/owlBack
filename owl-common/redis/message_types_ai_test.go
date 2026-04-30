package redis

import "testing"

func TestSplitDeviceType(t *testing.T) {
	cases := []struct {
		in       string
		wantBase string
		wantAI   string
	}{
		{"Radar", "Radar", ""},
		{"Sleepad", "Sleepad", ""},
		{"Radar.AI01", "Radar", "AI01"},
		{"Sleepad.AI02", "Sleepad", "AI02"},
		{"Radar.AI", "Radar", "AI"},
		{"", "", ""},
		{"Radar.foo", "Radar.foo", ""}, // 非 .AI 后缀不动
	}
	for _, c := range cases {
		base, ai := SplitDeviceType(c.in)
		if base != c.wantBase || ai != c.wantAI {
			t.Errorf("SplitDeviceType(%q) = (%q, %q), want (%q, %q)",
				c.in, base, ai, c.wantBase, c.wantAI)
		}
		if BaseDeviceType(c.in) != c.wantBase {
			t.Errorf("BaseDeviceType(%q) = %q, want %q", c.in, BaseDeviceType(c.in), c.wantBase)
		}
		if IsAIDerived(c.in) != (c.wantAI != "") {
			t.Errorf("IsAIDerived(%q) = %v, want %v", c.in, IsAIDerived(c.in), c.wantAI != "")
		}
	}
}
