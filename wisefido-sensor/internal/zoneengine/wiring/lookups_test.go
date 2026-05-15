package wiring

import (
	"testing"
)

// 没接 DB 的 fallback：BedSizeLookup 应默认 small（与 zoneengine.BedSizeBucket 兜底一致）。
func TestBedSizeLookup_NilDBDefaultsSmall(t *testing.T) {
	l := NewBedSizeLookup(nil, nil)
	got := l.BedSizeBucket("fd00:0:3:111:3:101::/96")
	if got != "small" {
		t.Errorf("nil DB should default 'small', got %q", got)
	}
	// 第二次命中缓存
	got2 := l.BedSizeBucket("fd00:0:3:111:3:101::/96")
	if got2 != "small" {
		t.Errorf("cache miss after first query? got %q", got2)
	}
}

func TestBedSizeLookup_EmptyZoneID(t *testing.T) {
	l := NewBedSizeLookup(nil, nil)
	if got := l.BedSizeBucket(""); got != "small" {
		t.Errorf("empty zone id should default small, got %q", got)
	}
}

func TestBedSizeLookup_InvalidateAll(t *testing.T) {
	l := NewBedSizeLookup(nil, nil)
	l.BedSizeBucket("fd00:0:3:111:3:101::/96") // populate cache
	l.InvalidateAll()
	if len(l.cache) != 0 {
		t.Errorf("cache not cleared, size=%d", len(l.cache))
	}
}

// nil DB → BathroomLookup 默认 false。
func TestBathroomLookup_NilDBDefaultsFalse(t *testing.T) {
	l := NewBathroomLookup(nil, nil)
	if got := l.IsBathroom("fd00:0:3:111:3:100::/88"); got {
		t.Errorf("nil DB should default false, got true")
	}
}

func TestBathroomLookup_EmptyZoneID(t *testing.T) {
	l := NewBathroomLookup(nil, nil)
	if l.IsBathroom("") {
		t.Errorf("empty zone id should default false")
	}
}

func TestBathroomLookup_InvalidateAll(t *testing.T) {
	l := NewBathroomLookup(nil, nil)
	l.IsBathroom("fd00:0:3:111:3:100::/88") // populate cache
	l.InvalidateAll()
	if len(l.cache) != 0 {
		t.Errorf("cache not cleared, size=%d", len(l.cache))
	}
}
