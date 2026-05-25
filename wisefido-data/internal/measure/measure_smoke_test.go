package measure

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSmokeWithRealTracks(t *testing.T) {
	xy := [][2]int{
		{0, 0}, {0, 0}, {0, 0},
		{-280, 470}, {-250, 460}, {-230, 470}, {-160, 430}, {-60, 420},
		{-50, 370}, {-90, 300}, {-160, 280}, {-270, 280}, {-290, 340},
		{-280, 420}, {-280, 470}, {-290, 490}, {-310, 480}, {-310, 480}, {-310, 480}, {-310, 480},
	}
	seen := map[[2]int]int{}
	for _, p := range xy {
		col, row := p[0]/10, p[1]/10
		seen[[2]int{col, row}]++
	}
	cells := make([]Cell, 0)
	for k, v := range seen {
		cells = append(cells, Cell{Col: k[0], Row: k[1], HitCount: v})
	}
	fmt.Printf("input: %d unique cells from 20 track points\n\n", len(cells))

	poly, ok := FitPolygon(cells, 1)
	b, _ := json.MarshalIndent(poly, "", "  ")
	fmt.Printf("FitPolygon ok=%v vertices=%d\n%s\n\n", ok, len(poly.Vertices), b)

	rect, ok := FitRectangle(cells, 1)
	b, _ = json.MarshalIndent(rect, "", "  ")
	fmt.Printf("FitRectangle ok=%v\n%s\n\n", ok, b)

	fov := Polygon{Vertices: []Point{
		{-500, -100}, {500, -100}, {500, 600}, {-500, 600},
	}}
	inscribed, ok := MaxInscribedRect(poly, fov)
	b, _ = json.MarshalIndent(inscribed, "", "  ")
	fmt.Printf("MaxInscribedRect ok=%v (firmware self-measured rect was 310×490 cm)\n%s\n", ok, b)
}
