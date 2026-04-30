package service

import (
	"testing"
	"time"
)

func TestAIOverrideCache_DefaultsToSandbox(t *testing.T) {
	c := NewAIOverrideCache("invalid", 60, nil)
	if c.Mode() != AIOverrideModeSandbox {
		t.Errorf("invalid mode should default to sandbox, got %q", c.Mode())
	}
	c2 := NewAIOverrideCache("", 60, nil)
	if c2.Mode() != AIOverrideModeSandbox {
		t.Errorf("empty mode should default to sandbox, got %q", c2.Mode())
	}
}

func TestAIOverrideCache_SandboxDoesNotMutate(t *testing.T) {
	c := NewAIOverrideCache("sandbox", 60, nil)
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})

	fields := map[string]interface{}{
		"track_id":         float64(0),
		"track_confidence": float64(60),
	}
	c.Apply("dev1", 0, fields)

	if got := readIntField(fields, "track_confidence"); got != 60 {
		t.Errorf("sandbox should NOT mutate track_confidence, got %d (want 60)", got)
	}
	if _, has := fields["ai_source"]; has {
		t.Errorf("sandbox should NOT add ai_source, got %v", fields["ai_source"])
	}
}

func TestAIOverrideCache_ReleaseOverwritesConfidence(t *testing.T) {
	c := NewAIOverrideCache("release", 60, nil)
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})

	fields := map[string]interface{}{
		"track_id":         float64(0),
		"track_confidence": float64(60),
	}
	c.Apply("dev1", 0, fields)

	if got := readIntField(fields, "track_confidence"); got != 20 {
		t.Errorf("release should overwrite to 20, got %d", got)
	}
	if got, _ := fields["ai_source"].(string); got != "ai_ghost_penalty" {
		t.Errorf("release should set ai_source, got %q", got)
	}
}

func TestAIOverrideCache_ClearDevice(t *testing.T) {
	c := NewAIOverrideCache("release", 60, nil)
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})
	c.Set("dev1", 1, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})
	c.Set("dev2", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})

	c.ClearDevice("dev1")

	stats := c.Stats()
	if stats.Devices != 1 || stats.Tracks != 1 {
		t.Errorf("after ClearDevice(dev1): Devices=%d Tracks=%d, want 1/1", stats.Devices, stats.Tracks)
	}

	// dev2 应保留
	fields := map[string]interface{}{"track_confidence": float64(60)}
	c.Apply("dev2", 0, fields)
	if got := readIntField(fields, "track_confidence"); got != 20 {
		t.Errorf("dev2 verdict should still apply, got %d", got)
	}

	// dev1 应被清
	fields2 := map[string]interface{}{"track_confidence": float64(60)}
	c.Apply("dev1", 0, fields2)
	if got := readIntField(fields2, "track_confidence"); got != 60 {
		t.Errorf("dev1 verdict should be gone, got %d", got)
	}
}

func TestAIOverrideCache_GC(t *testing.T) {
	c := NewAIOverrideCache("release", 1, nil) // TTL 1s
	now := time.Now().UnixMilli()
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty", UpdatedMs: now - 5000})
	c.Set("dev1", 1, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty", UpdatedMs: now})
	c.Set("dev2", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty", UpdatedMs: now - 5000})

	removed := c.GC(now)
	if removed != 2 {
		t.Errorf("GC should remove 2 stale, got %d", removed)
	}
	stats := c.Stats()
	if stats.Devices != 1 || stats.Tracks != 1 {
		t.Errorf("after GC: Devices=%d Tracks=%d, want 1/1", stats.Devices, stats.Tracks)
	}
}

func TestAIOverrideCache_TrackIDIsolation(t *testing.T) {
	// 验证：dev1.track_id=0 不影响 dev2.track_id=0（device_uid 维度隔离）
	c := NewAIOverrideCache("release", 60, nil)
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})

	fields := map[string]interface{}{"track_confidence": float64(60)}
	c.Apply("dev2", 0, fields)
	if got := readIntField(fields, "track_confidence"); got != 60 {
		t.Errorf("dev2.track 0 should not see dev1's verdict, got %d", got)
	}
}

func TestAIOverrideCache_SetModeSwitches(t *testing.T) {
	c := NewAIOverrideCache("sandbox", 60, nil)
	c.Set("dev1", 0, AIVerdict{Confidence: 20, Source: "ai_ghost_penalty"})

	// sandbox: no mutation
	fields := map[string]interface{}{"track_confidence": float64(60)}
	c.Apply("dev1", 0, fields)
	if got := readIntField(fields, "track_confidence"); got != 60 {
		t.Errorf("sandbox: confidence should be unchanged, got %d", got)
	}

	// switch to release
	c.SetMode("release")
	fields2 := map[string]interface{}{"track_confidence": float64(60)}
	c.Apply("dev1", 0, fields2)
	if got := readIntField(fields2, "track_confidence"); got != 20 {
		t.Errorf("release after switch: confidence should be 20, got %d", got)
	}
}
