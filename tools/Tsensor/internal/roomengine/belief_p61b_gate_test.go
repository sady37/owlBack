package roomengine

import (
	"os"
	"path/filepath"
	"testing"

	"owl-common/card"
	"owl-common/radarutils"
)

// TestP61bCABBGateEngages — 审查㉟ 放行 gate:真 CABB 立项案(doc/cases/hunzi-cabb-lost-0601-2247-FP)
// **无显式 WallPolygon/RoomW/H** → ParseLayoutConfig 由 radar **boundary**(非 FOV/signalRadius)派生 WallPolygon。
// 验:派生 WallPolygon bbox 最小边 ≤200 → 小卫生间 gate **fire** → D-path 在立项案 engage(非 silent-miss)。
// 委员会 silent-miss 担忧(若 bbox 退 radar FOV>200 则不 engage)被证伪:boundaryPolygonForStamp 用 boundary。
func TestP61bCABBGateEngages(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../../../doc/cases", "hunzi-cabb-lost-0601-2247-FP/room_layout.json"))
	if err != nil {
		t.Skipf("CABB 立项 fixture 缺失: %v", err)
	}
	cfg, err := ParseLayoutConfig("fd00:0:3:411::/128", b)
	if err != nil {
		t.Fatalf("ParseLayoutConfig: %v", err)
	}
	if len(cfg.WallPolygon) < 3 {
		t.Fatalf("无 wall 应由 radar boundary 派生 WallPolygon,得 %d 点", len(cfg.WallPolygon))
	}
	cfg.RoomType = card.RoomTypeBathroom // 真 CABB 是浴室(生产 RoomType,非 layout JSON)
	if !isSmallBathroomCfg(cfg, cfg.RoomW, cfg.RoomH) {
		t.Fatalf("真 CABB 立项案 gate 应 fire(boundary 派生 bbox 最小边≤200);否则 D-path 不 engage=silent-miss 治不了 CABB")
	}
}

// TestP6SmallBathroomGate — P6.1b-D(审查㉛ Opt-1)阶段1:小卫生间 gate 判据。
func TestP6SmallBathroomGate(t *testing.T) {
	poly := func(w, h int) []radarutils.Point {
		return []radarutils.Point{{X: 0, Y: 0}, {X: w, Y: 0}, {X: w, Y: h}, {X: 0, Y: h}}
	}
	cases := []struct {
		name        string
		roomType    int
		wall        []radarutils.Point
		rawW, rawH  int
		wantSmall   bool
	}{
		{"小卫生间-wall最小边180", card.RoomTypeBathroom, poly(180, 300), 600, 600, true},
		{"大卫生间-wall最小边260", card.RoomTypeBathroom, poly(260, 400), 600, 600, false},
		{"边界200恰好", card.RoomTypeBathroom, poly(200, 500), 600, 600, true},
		{"非卫生间(卧室)小", card.RoomTypeDefault, poly(180, 300), 600, 600, false},
		{"小卫生间-无wall退rawW180", card.RoomTypeBathroom, nil, 180, 400, true},
		{"大卫生间-无wall退rawW260", card.RoomTypeBathroom, nil, 260, 400, false},
	}
	for _, c := range cases {
		got := isSmallBathroomCfg(RoomConfig{RoomType: c.roomType, WallPolygon: c.wall}, c.rawW, c.rawH)
		if got != c.wantSmall {
			t.Errorf("%s: isSmallBathroom=%v 期望 %v", c.name, got, c.wantSmall)
		}
	}
}
