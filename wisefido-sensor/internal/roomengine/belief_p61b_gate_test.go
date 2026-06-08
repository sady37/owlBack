package roomengine

import (
	"testing"

	"owl-common/card"
	"owl-common/radarutils"
)

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
