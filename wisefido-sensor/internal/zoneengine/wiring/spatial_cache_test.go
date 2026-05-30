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

// stubFitness 实现 RadarFitnessChecker — fit 集合白名单。
type stubFitness struct {
	fit map[string]bool
}

func (s stubFitness) IsFit(addr string) bool {
	if s.fit == nil {
		return true
	}
	return s.fit[addr]
}

// HasRadarInRoom — 三档：纯 sleepace / 有 radar 但 unfit / 有 radar 且 fit。
// card f61e8o R1：DB 有 radar 但物理 offline 时,等同无 radar；详 [[subset_invariant_no_radar_gate]]。
func TestSpatialCache_HasRadarInRoom_FitnessGate(t *testing.T) {
	unitID := mustPrefix(t, "fd00:0:3:111:3::/80")
	// /88 100 装了 radarA + radarB；/88 200 仅 sleepace
	radarA := mustPrefix(t, "fd00:0:3:111:3:100:aaaa:1111/128").Addr()
	radarB := mustPrefix(t, "fd00:0:3:111:3:100:bbbb:2222/128").Addr()
	sleepad := mustPrefix(t, "fd00:0:3:111:3:201:cccc:3333/128").Addr()

	makeCache := func(fit map[string]bool) *SpatialCache {
		c := NewSpatialCache(nil, zap.NewNop())
		c.units[unitID] = &UnitData{
			UnitID: unitID,
			Devices: map[netip.Addr]*DeviceMeta{
				radarA:  {Addr: radarA, DeviceType: "Radar", Access: true},
				radarB:  {Addr: radarB, DeviceType: "Radar", Access: true},
				sleepad: {Addr: sleepad, DeviceType: "Sleepad", Access: true},
			},
			loadedAt: time.Now(),
		}
		c.SetFitness(stubFitness{fit: fit})
		return c
	}

	// 全 fit：/88 100 有 radar 视为有；/88 200 纯 sleepace 视为无
	c := makeCache(map[string]bool{radarA.String(): true, radarB.String(): true})
	if !c.HasRadarInRoom("fd00:0:3:111:3:100::/88") {
		t.Error("HasRadarInRoom(/88 100) should be true when radars are fit")
	}
	if c.HasRadarInRoom("fd00:0:3:111:3:200::/88") {
		t.Error("HasRadarInRoom(/88 200) should be false (sleepace only)")
	}

	// radarA offline、radarB 仍 fit → /88 100 视为有 radar（任一 fit 即可）
	c = makeCache(map[string]bool{radarA.String(): false, radarB.String(): true})
	if !c.HasRadarInRoom("fd00:0:3:111:3:100::/88") {
		t.Error("HasRadarInRoom should remain true while radarB still fit")
	}

	// radarA + radarB 全 offline → /88 100 等同无 radar（R1 闭环：drop_no_radar 应 fire）
	c = makeCache(map[string]bool{radarA.String(): false, radarB.String(): false})
	if c.HasRadarInRoom("fd00:0:3:111:3:100::/88") {
		t.Error("HasRadarInRoom should be false when all radars in /88 are unfit")
	}

	// fitness 未注入（启动期 / nil） → 退化 DB-only，全部 fit 默认通过
	c = NewSpatialCache(nil, zap.NewNop())
	c.units[unitID] = &UnitData{
		UnitID: unitID,
		Devices: map[netip.Addr]*DeviceMeta{
			radarA: {Addr: radarA, DeviceType: "Radar", Access: true},
		},
		loadedAt: time.Now(),
	}
	if !c.HasRadarInRoom("fd00:0:3:111:3:100::/88") {
		t.Error("HasRadarInRoom should return true when fitness not set (DB-only fallback)")
	}
}

func mustPrefix(t *testing.T, s string) netip.Prefix {
	t.Helper()
	p, err := netip.ParsePrefix(s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return p
}
