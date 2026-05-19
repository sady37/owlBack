package roomengine

import (
	"os"
	"path/filepath"
	"testing"
)

// renderLayoutAsPNG 公共辅助：解析 layout → 构建 grid → 渲染 PNG
func renderLayoutAsPNG(t *testing.T, layoutPath, outPath string) {
	t.Helper()
	raw, err := os.ReadFile(layoutPath)
	if err != nil {
		t.Skipf("skip: %v", err)
		return
	}

	cfg, err := ParseLayoutConfig("room", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, 10)
	if cfg.OriginX != 0 || cfg.OriginY != 0 {
		grid.OriginX = cfg.OriginX
		grid.OriginY = cfg.OriginY
	}
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}
	grid.StampRadar(cfg.Radar)
	grid.StampEnters(cfg.Enters, cfg.EnterTargets)

	for _, r := range cfg.Enters {
		grid.SetPrior(r, AreaEnter, 99, SourceHuman)
	}
	for _, r := range cfg.Beds {
		grid.SetPrior(r, AreaBed, 99, SourceHuman)
	}
	for _, r := range cfg.Toilets {
		grid.SetPrior(r, AreaToilet, 99, SourceHuman)
	}
	for _, r := range cfg.Showers {
		grid.SetPrior(r, AreaShower, 99, SourceHuman)
	}
	for _, r := range cfg.Chairs {
		grid.SetPrior(r, AreaSit, 80, SourceHuman)
	}
	for _, r := range cfg.Furnitures {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}
	for _, r := range cfg.Interferes {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}

	if err := RenderGridPNG(grid, cfg.Radar, cfg.Enters, cfg.WallPolygon, outPath); err != nil {
		t.Fatalf("render: %v", err)
	}

	// 统计几何状态分布
	var inRoomInFOV, inRoomNoFOV, noRoomInFOV, noRoomNoFOV int
	for i := range grid.Cells {
		c := &grid.Cells[i]
		switch {
		case c.InRoom && c.InFOV:
			inRoomInFOV++
		case c.InRoom && !c.InFOV:
			inRoomNoFOV++
		case !c.InRoom && c.InFOV:
			noRoomInFOV++
		default:
			noRoomNoFOV++
		}
	}

	t.Logf("PNG: %s  (grid %d×%d = %dpx×%dpx)",
		outPath, grid.Width, grid.Height, grid.Width*10, grid.Height*10)
	t.Logf("Cells: InRoom×InFOV=%d  InRoom×!InFOV=%d  !InRoom×InFOV=%d  !InRoom×!InFOV=%d",
		inRoomInFOV, inRoomNoFOV, noRoomInFOV, noRoomNoFOV)
	t.Logf("Radar: (%d,%d,%d) install=%d",
		cfg.Radar.Center.X, cfg.Radar.Center.Y, cfg.Radar.Center.Z, cfg.Radar.InstallModel)
}

func TestRenderGrid_D523(t *testing.T) {
	renderLayoutAsPNG(t,
		filepath.Join("..", "..", "..", "doc", "layout-D523-demo.json"),
		filepath.Join("..", "..", "..", "doc", "grid-D523-demo.png"),
	)
}

func TestRenderGrid_D5F7(t *testing.T) {
	renderLayoutAsPNG(t,
		filepath.Join("..", "..", "..", "doc", "layout-D5F7-bathroom.json"),
		filepath.Join("..", "..", "..", "doc", "grid-D5F7-bathroom.png"),
	)
}
