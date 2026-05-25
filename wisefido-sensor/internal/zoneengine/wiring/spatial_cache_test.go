package wiring

import (
	"net/netip"
	"testing"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// classifyRoomType — DB raw smallint 0/1/2，name 关键词作 0=Default 时兜底。
func TestClassifyRoomType(t *testing.T) {
	cases := []struct {
		roomType int
		roomName string
		want     int
	}{
		{card.RoomTypeBathroom, "", card.RoomTypeBathroom},   // 1 直判
		{card.RoomTypeKitchen, "", card.RoomTypeKitchen},     // 2 直判
		{card.RoomTypeDefault, "Master Bathroom", card.RoomTypeBathroom}, // 0 + name 兜底
		{card.RoomTypeDefault, "Toilet 2F", card.RoomTypeBathroom},
		{card.RoomTypeDefault, "Living Room", card.RoomTypeDefault},
		{card.RoomTypeDefault, "", card.RoomTypeDefault},
	}
	for _, c := range cases {
		if got := classifyRoomType(c.roomType, c.roomName); got != c.want {
			t.Errorf("classifyRoomType(%d, %q): want %d, got %d",
				c.roomType, c.roomName, c.want, got)
		}
	}
}

// preferredDeviceTypeForAlarm (pure func) — alarm 类型 → device 类型映射。
func TestPreferredDeviceTypeForAlarm(t *testing.T) {
	cases := []struct {
		alarmType string
		want      string
	}{
		{alarm.LeftBed, "Radar"},
		{alarm.InBed, "Radar"},
		{alarm.Fall, "Radar"},
		{alarm.SuspectedFall, "Radar"},
		{alarm.SittingOnGround, "Radar"},
		{alarm.SuspectedSittingOnGround, "Radar"},
		{alarm.Stay, ""},
		{alarm.NightAbsence, ""},
		{"unknown", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := preferredDeviceTypeForAlarm(c.alarmType); got != c.want {
			t.Errorf("preferredDeviceTypeForAlarm(%q): want %q, got %q", c.alarmType, c.want, got)
		}
	}
}

// maskToUnit (pure) — /88/96/128 截到 /80。
func TestMaskToUnit(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"fd00:0:3:111:3:300::/88", "fd00:0:3:111:3::/80"},
		{"fd00:0:3:111:3:101::/96", "fd00:0:3:111:3::/80"},
		{"fd00:0:3:111:3:300:abcd:ef01/128", "fd00:0:3:111:3::/80"},
	}
	for _, c := range cases {
		p, err := netip.ParsePrefix(c.in)
		if err != nil {
			t.Fatalf("parse %s: %v", c.in, err)
		}
		got := maskToUnit(p).String()
		if got != c.want {
			t.Errorf("maskToUnit(%s): want %s, got %s", c.in, c.want, got)
		}
	}
}

// SpatialCache 公共方法 — nil DB 退化语义验证。
func TestSpatialCache_NilDB_DefaultBehavior(t *testing.T) {
	c := NewSpatialCache(nil, zap.NewNop())
	// LoadAll on nil DB = no-op
	if err := c.LoadAll(nil); err != nil {
		t.Errorf("nil DB LoadAll should succeed: %v", err)
	}
	// IsBathroom: 未配置 → false
	if c.IsBathroom("fd00:0:3:111:3:300::/88") {
		t.Errorf("nil DB: IsBathroom should be false")
	}
	// BedSizeBucket: 未配置 → "small"
	if c.BedSizeBucket("fd00:0:3:111:3:101::/96") != "small" {
		t.Errorf("nil DB: BedSizeBucket should default 'small'")
	}
	// BedMode: 未配置 → Facility
	if c.BedMode("fd00:0:3:111:3:101::/96") != zoneengine.BedModeFacility {
		t.Errorf("nil DB: BedMode should default Facility")
	}
	// FindPrimaryDevice: 不归因 alarm 类型 → 0 addr
	if c.FindPrimaryDevice("fd00:0:3:111:3:101::/96", alarm.Stay).IsValid() {
		t.Errorf("FindPrimaryDevice for Stay should return zero")
	}
}

// SpatialCache 公共方法 — 注入 per-unit cache 后 Lookup 行为。
func TestSpatialCache_PerUnitCache(t *testing.T) {
	c := NewSpatialCache(nil, zap.NewNop())
	unitID := mustPrefix(t, "fd00:0:3:111:3::/80")
	bathPrefix := mustPrefix(t, "fd00:0:3:111:3:300::/88")
	bedPrefix := mustPrefix(t, "fd00:0:3:111:3:101::/96")
	devAddr := mustPrefix(t, "fd00:0:3:111:3:100:abcd:ef01/128").Addr()

	c.units[unitID] = &UnitData{
		UnitID:       unitID,
		UnitProperty: UnitPropertyHome,
		Rooms:        map[netip.Prefix]*RoomMeta{bathPrefix: {Prefix: bathPrefix, RoomType: card.RoomTypeBathroom}},
		Beds:         map[netip.Prefix]*BedMeta{bedPrefix: {Prefix: bedPrefix, SizeKindText: "queen"}},
		Devices:      map[netip.Addr]*DeviceMeta{devAddr: {Addr: devAddr, DeviceType: "Radar", Access: true, Monitoring: true}},
		loadedAt:     time.Now(),
	}

	if !c.IsBathroom("fd00:0:3:111:3:300::/88") {
		t.Error("IsBathroom should be true for bathroom-typed entry")
	}
	if c.IsBathroom("fd00:0:3:111:3:301::/88") {
		t.Error("IsBathroom should be false for unknown prefix")
	}
	if c.BedSizeBucket("fd00:0:3:111:3:101::/96") != "large" {
		t.Errorf("BedSizeBucket should derive 'large' from raw 'queen', got %s",
			c.BedSizeBucket("fd00:0:3:111:3:101::/96"))
	}
	if c.BedMode("fd00:0:3:111:3:101::/96") != zoneengine.BedModeHome {
		t.Error("BedMode should follow unit_property=Home")
	}
	addr := c.FindPrimaryDevice("fd00:0:3:111:3:100::/88", alarm.Fall)
	if !addr.IsValid() {
		t.Error("FindPrimaryDevice should hit the radar in /88 range")
	}
}

// SpatialCache InvalidateAll 不 panic — DB nil 时 LoadAll 走 no-op。
func TestSpatialCache_InvalidateAll(t *testing.T) {
	c := NewSpatialCache(nil, zap.NewNop())
	c.InvalidateAll() // 不应 panic
}

func mustPrefix(t *testing.T, s string) netip.Prefix {
	t.Helper()
	p, err := netip.ParsePrefix(s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return p
}
