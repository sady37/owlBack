package zoneengine

import (
	"testing"
	"time"
)

// stubVitalSource 测试用 VitalSource，固定返回一组 (bedZoneID, ts)。
type stubVitalSource struct {
	readings []struct {
		bedZoneID string
		ts        int64
	}
}

func (s *stubVitalSource) ScanActiveBedVitals(_, _ int64, emit func(string, int64)) {
	for _, r := range s.readings {
		emit(r.bedZoneID, r.ts)
	}
}

func TestVitalAdapter_SustainHoldsBedAcrossDecay(t *testing.T) {
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	cap := &captureListener{}
	engine.AddListener(cap)

	now := time.Now().UnixMilli()
	bedID := "fd00:0:3:111:3:101::/96"

	// 1) 直接发 enter（无 sleepace 来源会被 latch 阻塞，所以构造一条干净的 enter）
	engine.Apply(SignalEvidence{
		ZoneType: ZoneTypeBed, ZoneID: bedID,
		Source: "sleepace", Kind: "enter", Ts: now,
	})

	// 2) 经过 60s（远超 enter strength 衰减一半，但 < DecayWindowMs=120s），靠 sustain 维持
	src := &stubVitalSource{readings: []struct {
		bedZoneID string
		ts        int64
	}{
		{bedZoneID: bedID, ts: now + 60_000},
	}}
	a := NewVitalAdapter(src, engine, nil)
	a.tickOnce(now + 60_000)

	// 3) 推 Tick 一下让 engine 重新计算（sustain 进 score）
	engine.Tick(now + 60_000)

	// 期望：bed 仍然 occupied（不应回落）
	state, ok := engine.GetState(StateKey{ZoneType: ZoneTypeBed, ZoneID: bedID})
	if !ok {
		t.Fatalf("bed state missing")
	}
	if state.Status != StatusOccupied {
		t.Fatalf("want occupied, got %v (score=%d)", state.Status, state.Score)
	}
}

func TestVitalAdapter_NoSourceNoOp(t *testing.T) {
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	a := NewVitalAdapter(nil, engine, nil)
	// 不应 panic
	a.tickOnce(time.Now().UnixMilli())
}

func TestVitalAdapter_SkipEmptyKeys(t *testing.T) {
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	cap := &captureListener{}
	engine.AddListener(cap)
	src := &stubVitalSource{readings: []struct {
		bedZoneID string
		ts        int64
	}{
		{bedZoneID: "", ts: 100}, // empty bedZoneID 应被 skip
	}}
	a := NewVitalAdapter(src, engine, nil)
	a.tickOnce(time.Now().UnixMilli())

	if len(cap.Events()) != 0 {
		t.Fatalf("empty keys should be skipped, got %v", cap.Events())
	}
}

func TestVitalAdapter_WithIntervalAndFreshness(t *testing.T) {
	a := NewVitalAdapter(nil, nil, nil)
	a = a.WithInterval(2 * time.Second).WithFreshness(60_000)
	if a.interval != 2*time.Second {
		t.Errorf("interval not updated: %v", a.interval)
	}
	if a.freshnessMs != 60_000 {
		t.Errorf("freshness not updated: %d", a.freshnessMs)
	}
	// 0 / 负值 不应覆盖
	a.WithInterval(0).WithFreshness(-1)
	if a.interval != 2*time.Second || a.freshnessMs != 60_000 {
		t.Errorf("0/neg should not overwrite: interval=%v freshness=%d", a.interval, a.freshnessMs)
	}
}
