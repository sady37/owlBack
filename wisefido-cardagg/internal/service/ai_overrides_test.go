// ai_overrides_test.go — S5a AIOverrideCache T1 单元测试（同 package，纯逻辑无 IO）。
//
// 覆盖：mode 校验 + Set/Apply 双模式 + ClearDevice + GC TTL + Stats。

package service

import (
	"testing"

	"go.uber.org/zap"
)

func newTestCache(mode string) *AIOverrideCache {
	return NewAIOverrideCache(mode, 60, zap.NewNop())
}

func TestIsValidAIOverrideMode(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"sandbox", true},
		{"release", true},
		{"", false},
		{"bogus", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := IsValidAIOverrideMode(tc.in); got != tc.want {
				t.Errorf("IsValidAIOverrideMode(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestNewAIOverrideCache_InvalidModeFallsBackToSandbox(t *testing.T) {
	c := NewAIOverrideCache("bogus", 60, zap.NewNop())
	if c.Mode() != AIOverrideModeSandbox {
		t.Errorf("invalid mode should fallback sandbox, got %s", c.Mode())
	}
}

func TestSetAndApply_SandboxNoMutate(t *testing.T) {
	c := newTestCache("sandbox")
	c.Set("dev-a", 3, AIVerdict{Confidence: 20, Source: "AI.X", Reason: "ghost_penalty", UpdatedMs: 1_700_000_000_000})

	fields := map[string]interface{}{"track_confidence": float64(80)}
	c.Apply("dev-a", 3, fields)

	// sandbox 模式：fields 不动
	if fields["track_confidence"] != float64(80) {
		t.Errorf("sandbox should NOT mutate fields, got %v", fields["track_confidence"])
	}
	if _, ok := fields["ai_source"]; ok {
		t.Error("sandbox should NOT add ai_source")
	}
	if c.Stats().ApplyHits != 1 {
		t.Errorf("apply hit count = %d, want 1", c.Stats().ApplyHits)
	}
}

func TestSetAndApply_ReleaseOverwrites(t *testing.T) {
	c := newTestCache("release")
	c.Set("dev-a", 3, AIVerdict{Confidence: 20, Source: "AI.X", UpdatedMs: 1_700_000_000_000})

	fields := map[string]interface{}{"track_confidence": float64(80)}
	c.Apply("dev-a", 3, fields)

	if fields["track_confidence"] != 20 {
		t.Errorf("release should overwrite to 20, got %v", fields["track_confidence"])
	}
	if fields["ai_source"] != "AI.X" {
		t.Errorf("release should write ai_source, got %v", fields["ai_source"])
	}
}

func TestApply_MissDoesNotMutate(t *testing.T) {
	c := newTestCache("release")
	fields := map[string]interface{}{"track_confidence": float64(80)}
	c.Apply("dev-unknown", 3, fields)
	if fields["track_confidence"] != float64(80) {
		t.Errorf("miss should not mutate; got %v", fields["track_confidence"])
	}
	if c.Stats().ApplyMisses != 1 {
		t.Errorf("miss count = %d, want 1", c.Stats().ApplyMisses)
	}
}

func TestApply_DifferentTrackIDIsMiss(t *testing.T) {
	c := newTestCache("release")
	c.Set("dev-a", 3, AIVerdict{Confidence: 20})
	fields := map[string]interface{}{"track_confidence": float64(80)}
	c.Apply("dev-a", 5, fields) // 不同 tid
	if fields["track_confidence"] != float64(80) {
		t.Errorf("different tid should miss; got %v", fields["track_confidence"])
	}
}

func TestClearDevice_RemovesAllTracks(t *testing.T) {
	c := newTestCache("sandbox")
	c.Set("dev-a", 1, AIVerdict{Confidence: 30})
	c.Set("dev-a", 2, AIVerdict{Confidence: 40})
	c.Set("dev-b", 1, AIVerdict{Confidence: 50})

	c.ClearDevice("dev-a")
	stats := c.Stats()
	if stats.Devices != 1 || stats.Tracks != 1 {
		t.Errorf("after ClearDevice(dev-a): devs=%d tracks=%d want 1/1", stats.Devices, stats.Tracks)
	}
}

func TestGC_DropsExpired(t *testing.T) {
	c := NewAIOverrideCache("sandbox", 60, zap.NewNop()) // 60s TTL
	now := int64(1_700_000_000_000)
	c.Set("dev-a", 1, AIVerdict{Confidence: 20, UpdatedMs: now - 120_000}) // 2min 前 = 过期
	c.Set("dev-a", 2, AIVerdict{Confidence: 30, UpdatedMs: now - 30_000})  // 30s 前 = fresh
	c.Set("dev-b", 1, AIVerdict{Confidence: 40, UpdatedMs: now - 90_000})  // 90s 前 = 过期

	removed := c.GC(now)
	if removed != 2 {
		t.Errorf("GC removed = %d, want 2", removed)
	}
	stats := c.Stats()
	if stats.Tracks != 1 {
		t.Errorf("after GC: tracks=%d want 1 (only dev-a tid=2 fresh)", stats.Tracks)
	}
}

func TestSetMode_RuntimeSwitch(t *testing.T) {
	c := newTestCache("sandbox")
	c.SetMode("release")
	if c.Mode() != AIOverrideModeRelease {
		t.Errorf("SetMode release: got %s", c.Mode())
	}
	c.SetMode("invalid") // 不变
	if c.Mode() != AIOverrideModeRelease {
		t.Errorf("invalid SetMode should noop; got %s", c.Mode())
	}
}

func TestReadIntFieldAI(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		key  string
		want int
	}{
		{"float64", map[string]interface{}{"k": 3.7}, "k", 3},
		{"int", map[string]interface{}{"k": 5}, "k", 5},
		{"int64", map[string]interface{}{"k": int64(7)}, "k", 7},
		{"missing", map[string]interface{}{}, "k", 0},
		{"unsupported", map[string]interface{}{"k": "5"}, "k", 0},
		{"nil-map", nil, "k", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := readIntFieldAI(tc.m, tc.key); got != tc.want {
				t.Errorf("readIntFieldAI(%v) = %d, want %d", tc.m, got, tc.want)
			}
		})
	}
}
