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
		bedZoneID string
		ts        int64
	}
	var gotSource string
	src.ScanActiveBedVitals(now+1000, 30_000, func(bid, source string, ts int64) {
		gotSource = source
		got = append(got, struct {
			bedZoneID string
			ts        int64
		}{bid, ts})
	})
	if len(got) != 1 {
		t.Fatalf("want 1 emit, got %d (%v)", len(got), got)
	}
	if got[0].bedZoneID != "fd00:0:3:111:3:101::/96" {
		t.Errorf("bedZoneID: %q", got[0].bedZoneID)
	}
	if gotSource != "vital" {
		t.Errorf("sleepad source: %q (want vital)", gotSource)
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

// 2026-05-24 修订：HR>0 OR RR>0 OR body_move>0 OR turn_over>0 OR bed_status==0 任一成立都算压感证据。
// 单 HR / 单 RR / 单 body_move 都应 emit；全 0 不 emit。
func TestMonitorVitalSource_AnyPresenceEvidenceEmits(t *testing.T) {
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
	// 全 0 → 不 emit
	buf.Write("card-zero", "fd00:0:3:111:3:103:a2ac:d525", "Sleepad", "0", map[string]any{
		observation.FieldHeartRate:       float64(0),
		observation.FieldRespiratoryRate: float64(0),
	}, now)

	var got int
	src.ScanActiveBedVitals(now+1000, 30_000, func(_, _ string, _ int64) { got++ })
	if got != 2 {
		t.Errorf("HR-only + RR-only should both emit (2 emit), got %d", got)
	}
}

type stubBedResolver struct{ bed string }

func (r stubBedResolver) ResolveBed(_, _ string) (string, float64) { return r.bed, 1.0 }

// radar HR/RR 由 firmware bed-enter 门控，存在即在床；经 BedResolver 归床，source=radar。
func TestMonitorVitalSource_RadarResolvedToBed(t *testing.T) {
	buf := service.NewMonitorBuffer()
	src := NewMonitorVitalSource(buf)
	src.SetBedResolver(stubBedResolver{bed: "fd00:0:3:111:3:200::/96"})
	now := time.Now().UnixMilli()
	buf.Write("card-radar", "fd00:0:3:111:3:101:a2ac:d523", "Radar", "0", map[string]any{
		observation.FieldHeartRate:       float64(72),
		observation.FieldRespiratoryRate: float64(15),
	}, now)
	var gotBed, gotSource string
	var got int
	src.ScanActiveBedVitals(now+1000, 30_000, func(bid, source string, _ int64) {
		got++
		gotBed, gotSource = bid, source
	})
	if got != 1 {
		t.Fatalf("radar HR/RR (bed-gated) should emit, got %d", got)
	}
	if gotBed != "fd00:0:3:111:3:200::/96" {
		t.Errorf("radar bed via resolver: %q", gotBed)
	}
	if gotSource != "radar" {
		t.Errorf("radar source: %q (want radar)", gotSource)
	}
}

// 多床候选未定 → BedResolver 返回 "" → 不发，让位 sleepad；无 resolver 注入同样静默跳过。
func TestMonitorVitalSource_RadarUnresolvedSkipped(t *testing.T) {
	buf := service.NewMonitorBuffer()
	now := time.Now().UnixMilli()
	buf.Write("card-radar", "fd00:0:3:111:3:101:a2ac:d523", "Radar", "0", map[string]any{
		observation.FieldHeartRate:       float64(72),
		observation.FieldRespiratoryRate: float64(15),
	}, now)
	for _, tc := range []struct {
		name string
		src  *MonitorVitalSource
	}{
		{"nil resolver", NewMonitorVitalSource(buf)},
		{"empty resolve", func() *MonitorVitalSource {
			s := NewMonitorVitalSource(buf)
			s.SetBedResolver(stubBedResolver{bed: ""})
			return s
		}()},
	} {
		var got int
		tc.src.ScanActiveBedVitals(now+1000, 30_000, func(_, _ string, _ int64) { got++ })
		if got != 0 {
			t.Errorf("%s: radar unresolved should not emit, got %d", tc.name, got)
		}
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
		got := bedPrefixFromDeviceAddr(c.in)
		if got != c.want {
			t.Errorf("bedPrefixFromDeviceAddr(%q): want %q, got %q", c.in, c.want, got)
		}
	}
}
