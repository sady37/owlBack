package wiring

import (
	"net/netip"
	"testing"
	"time"

	"owl-common/spatial"

	"go.uber.org/zap"
)

type stubGeo struct{ reach int }

func (s stubGeo) RadarBedReachCount(string) int { return s.reach }

// 你的场景：room 内 bed1、bed2；sleepad 绑 bed1；radar 房绑，几何只够到 1 张床（m=1，n=2 → 候选 0.5）。
func TestMatrixCache_BuildUnit_CandidateCovers(t *testing.T) {
	unit := netip.MustParsePrefix("fd00:0:0:0:1::/80")
	room := netip.MustParsePrefix("fd00:0:0:0:1:300::/88")
	bed1 := netip.MustParsePrefix("fd00:0:0:0:1:364::/96")
	bed2 := netip.MustParsePrefix("fd00:0:0:0:1:399::/96")
	radarAddr := netip.MustParseAddr("fd00:0:0:0:1:300:0:a1") // 在 room /88 内，不在任何 bed /96 内
	slpdAddr := netip.MustParseAddr("fd00:0:0:0:1:364:0:b2")  // 在 bed1 /96 内

	sc := &SpatialCache{
		units:           make(map[netip.Prefix]*UnitData),
		refreshInterval: time.Minute,
		logger:          zap.NewNop(),
	}
	sc.units[unit] = &UnitData{
		UnitID: unit,
		Rooms:  map[netip.Prefix]*RoomMeta{room: {Prefix: room}},
		Beds:   map[netip.Prefix]*BedMeta{bed1: {Prefix: bed1}, bed2: {Prefix: bed2}},
		Devices: map[netip.Addr]*DeviceMeta{
			radarAddr: {Addr: radarAddr, DeviceType: "Radar"},
			slpdAddr:  {Addr: slpdAddr, DeviceType: "Sleepad"},
		},
		loadedAt: time.Now(),
	}

	mc := NewMatrixCache(sc, stubGeo{reach: 1}, zap.NewNop())
	mx := mc.MatrixFor(unit)
	if mx == nil {
		t.Fatal("MatrixFor 返回 nil")
	}

	radarPfx := netip.PrefixFrom(radarAddr, 128)
	slpdPfx := netip.PrefixFrom(slpdAddr, 128)

	// covers 候选：m=1,n=2 → 0.5（两床都 0.5，因几何无法贴 A/B 标签）
	if c := mx.Conf(radarPfx, bed1); c != 0.5 {
		t.Errorf("covers(radar,bed1) 应 0.5 候选，got %v", c)
	}
	if c := mx.Conf(radarPfx, bed2); c != 0.5 {
		t.Errorf("covers(radar,bed2) 应 0.5 候选，got %v", c)
	}
	// onbed 确定
	if c := mx.Conf(slpdPfx, bed1); c != spatial.ConfCertain {
		t.Errorf("onbed(sleepad,bed1) 应 1，got %v", c)
	}
	// 派生同床：min(covers 0.5, onbed 1)=0.5，候选不确定性传到 DBN 先验
	if c := mx.SameBedConf(radarPfx, slpdPfx); c != 0.5 {
		t.Errorf("samebed(radar,sleepad) 应 0.5（候选传播），got %v", c)
	}
	// FuseGroup(bed1)=radar@0.5 + sleepad@1
	fg := mx.FuseGroup(bed1)
	if len(fg) != 2 {
		t.Fatalf("FuseGroup(bed1) 应 2 源，got %v", fg)
	}
}

// ResolveBed 真消费：单床房解析到床；多床候选(0.5/0.5)留空待 DBN。
func TestMatrixCache_ResolveBed(t *testing.T) {
	// 多床房：radar 几何只够 1 床 → 两床各 0.5 候选 → ResolveBed 留空（退回 radar /96）
	unit := netip.MustParsePrefix("fd00:0:0:0:1::/80")
	room := netip.MustParsePrefix("fd00:0:0:0:1:300::/88")
	bed1 := netip.MustParsePrefix("fd00:0:0:0:1:364::/96")
	bed2 := netip.MustParsePrefix("fd00:0:0:0:1:399::/96")
	radarAddr := netip.MustParseAddr("fd00:0:0:0:1:300:0:a1")

	sc := &SpatialCache{units: make(map[netip.Prefix]*UnitData), refreshInterval: time.Minute, logger: zap.NewNop()}
	sc.units[unit] = &UnitData{
		UnitID:   unit,
		Rooms:    map[netip.Prefix]*RoomMeta{room: {Prefix: room}},
		Beds:     map[netip.Prefix]*BedMeta{bed1: {Prefix: bed1}, bed2: {Prefix: bed2}},
		Devices:  map[netip.Addr]*DeviceMeta{radarAddr: {Addr: radarAddr, DeviceType: "Radar"}},
		loadedAt: time.Now(),
	}
	mc := NewMatrixCache(sc, stubGeo{reach: 1}, zap.NewNop())
	if b, conf := mc.ResolveBed(radarAddr.String(), room.String()); b != "" || conf != 0 {
		t.Errorf("多床候选 0.5/0.5 应留空（待 DBN），got (%q, %v)", b, conf)
	}

	// 单床房：covers=1.0≥阈 → 解析到该床
	u2 := netip.MustParsePrefix("fd00:0:0:0:2::/80")
	r2 := netip.MustParsePrefix("fd00:0:0:0:2:300::/88")
	b2 := netip.MustParsePrefix("fd00:0:0:0:2:364::/96")
	ra2 := netip.MustParseAddr("fd00:0:0:0:2:300:0:a1")
	sc.units[u2] = &UnitData{
		UnitID:   u2,
		Rooms:    map[netip.Prefix]*RoomMeta{r2: {Prefix: r2}},
		Beds:     map[netip.Prefix]*BedMeta{b2: {Prefix: b2}},
		Devices:  map[netip.Addr]*DeviceMeta{ra2: {Addr: ra2, DeviceType: "Radar"}},
		loadedAt: time.Now(),
	}
	if b, conf := mc.ResolveBed(ra2.String(), r2.String()); b != b2.String() || conf != 1.0 {
		t.Errorf("单床房应解析到床 %s 权重 1.0，got (%q, %v)", b2, b, conf)
	}
}

// 单床房：m=1,n=1 → covers 1.0 确定（ResolveSingleBed 的几何泛化）。
func TestMatrixCache_BuildUnit_SingleBedCertain(t *testing.T) {
	unit := netip.MustParsePrefix("fd00:0:0:0:2::/80")
	room := netip.MustParsePrefix("fd00:0:0:0:2:300::/88")
	bed := netip.MustParsePrefix("fd00:0:0:0:2:364::/96")
	radarAddr := netip.MustParseAddr("fd00:0:0:0:2:300:0:a1")

	sc := &SpatialCache{
		units:           make(map[netip.Prefix]*UnitData),
		refreshInterval: time.Minute,
		logger:          zap.NewNop(),
	}
	sc.units[unit] = &UnitData{
		UnitID:   unit,
		Rooms:    map[netip.Prefix]*RoomMeta{room: {Prefix: room}},
		Beds:     map[netip.Prefix]*BedMeta{bed: {Prefix: bed}},
		Devices:  map[netip.Addr]*DeviceMeta{radarAddr: {Addr: radarAddr, DeviceType: "Radar"}},
		loadedAt: time.Now(),
	}

	mc := NewMatrixCache(sc, stubGeo{reach: 1}, zap.NewNop())
	mx := mc.MatrixFor(unit)
	if mx == nil {
		t.Fatal("MatrixFor 返回 nil")
	}
	if c := mx.Conf(netip.PrefixFrom(radarAddr, 128), bed); c != spatial.ConfCertain {
		t.Errorf("单床房 covers 应 1.0 确定，got %v", c)
	}
}
