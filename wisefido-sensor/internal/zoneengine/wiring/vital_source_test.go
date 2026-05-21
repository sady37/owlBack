package wiring

import (
	"testing"
	"time"

	"owl-common/observation"
	"wisefido-sensor/internal/service"
)

func TestMonitorVitalSource_EmitsForFreshHRRR(t *testing.T) {
	buf := service.NewMonitorBuffer()
	src := NewMonitorVitalSource(buf)

	now := time.Now().UnixMilli()
	deviceAddr := "fd00:0:3:111:3:101:a2ac:d523"
	buf.Write("card-1", deviceAddr, "Sleepad", "0", map[string]any{
		observation.FieldHeartRate:       float64(72),
		observation.FieldRespiratoryRate: float64(15),
	}, now)

	var got []struct {
		cardID, bedZoneID string
		ts                int64
	}
	src.ScanActiveBedVitals(now+1000, 30_000, func(cid, bid string, ts int64) {
		got = append(got, struct {
			cardID, bedZoneID string
			ts                int64
		}{cid, bid, ts})
	})
	if len(got) != 1 {
		t.Fatalf("want 1 emit, got %d (%v)", len(got), got)
	}
	if got[0].cardID != "card-1" {
		t.Errorf("cardID: %q", got[0].cardID)
	}
	if got[0].bedZoneID != "fd00:0:3:111:3:101::/96" {
		t.Errorf("bedZoneID: %q", got[0].bedZoneID)
	}
	if got[0].ts != now {
		t.Errorf("ts: %d (want %d)", got[0].ts, now)
	}
}

func TestMonitorVitalSource_SkipStale(t *testing.T) {
	buf := service.NewMonitorBuffer()
	src := NewMonitorVitalSource(buf)

	stale := time.Now().UnixMilli() - 60_000 // 60s 前
	buf.Write("card-1", "fd00:0:3:111:3:101:a2ac:d523", "Sleepad", "0", map[string]any{
		observation.FieldHeartRate:       float64(72),
		observation.FieldRespiratoryRate: float64(15),
	}, stale)

	var got int
	src.ScanActiveBedVitals(time.Now().UnixMilli(), 30_000, func(_, _ string, _ int64) { got++ })
	if got != 0 {
		t.Errorf("stale (>30s) should be skipped, got %d emits", got)
	}
}

func TestMonitorVitalSource_RequiresBothHRandRR(t *testing.T) {
	buf := service.NewMonitorBuffer()
	src := NewMonitorVitalSource(buf)
	now := time.Now().UnixMilli()

	// 仅 HR
	buf.Write("card-hr", "fd00:0:3:111:3:101:a2ac:d523", "Sleepad", "0", map[string]any{
		observation.FieldHeartRate: float64(72),
	}, now)
	// 仅 RR
	buf.Write("card-rr", "fd00:0:3:111:3:102:a2ac:d524", "Sleepad", "0", map[string]any{
		observation.FieldRespiratoryRate: float64(15),
	}, now)
	// HR=0 RR=0
	buf.Write("card-zero", "fd00:0:3:111:3:103:a2ac:d525", "Sleepad", "0", map[string]any{
		observation.FieldHeartRate:       float64(0),
		observation.FieldRespiratoryRate: float64(0),
	}, now)

	var got int
	src.ScanActiveBedVitals(now+1000, 30_000, func(_, _ string, _ int64) { got++ })
	if got != 0 {
		t.Errorf("none should emit (need both HR>0 AND RR>0), got %d", got)
	}
}

func TestMonitorVitalSource_RadarSkipped(t *testing.T) {
	// 设计约束（2026-05-20）：radar HR/RR ≠ in-bed 信号，不发 sustain
	buf := service.NewMonitorBuffer()
	src := NewMonitorVitalSource(buf)
	now := time.Now().UnixMilli()
	buf.Write("card-radar", "fd00:0:3:111:3:101:a2ac:d523", "Radar", "0", map[string]any{
		observation.FieldHeartRate:       float64(72),
		observation.FieldRespiratoryRate: float64(15),
	}, now)
	var got int
	src.ScanActiveBedVitals(now+1000, 30_000, func(_, _ string, _ int64) { got++ })
	if got != 0 {
		t.Errorf("radar HR/RR should NOT emit bed sustain, got %d emits", got)
	}
}

func TestMonitorVitalSource_NilBufferIsNoOp(t *testing.T) {
	src := NewMonitorVitalSource(nil)
	src.ScanActiveBedVitals(time.Now().UnixMilli(), 30_000, func(_, _ string, _ int64) {
		t.Error("nil buffer should not emit")
	})
}

func TestBedPrefixFromDeviceID(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"fd00:0:3:111:3:101:a2ac:d523", "fd00:0:3:111:3:101::/96"},
		{"fd00:0:3:111:3:101:a2ac:d523/128", "fd00:0:3:111:3:101::/96"}, // 带 mask 也接受
		{"", ""},
		{"not-an-ip", ""},
	}
	for _, c := range cases {
		got := bedPrefixFromDeviceID(c.in)
		if got != c.want {
			t.Errorf("bedPrefixFromDeviceID(%q): want %q, got %q", c.in, c.want, got)
		}
	}
}
