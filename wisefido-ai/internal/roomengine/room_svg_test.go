package roomengine

import (
	"os"
	"path/filepath"
	"testing"
)

func renderRoomLayoutSVG(t *testing.T, roomID, layoutPath, outPath string) {
	t.Helper()
	raw, err := os.ReadFile(layoutPath)
	if err != nil {
		t.Skipf("skip: %v", err)
		return
	}
	cfg, err := ParseLayoutConfig(roomID, raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 构建 grid + Stamp + 注入人标 prior（与 RegisterRoom 行为一致）
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, 10)
	if cfg.OriginX != 0 || cfg.OriginY != 0 {
		grid.OriginX = cfg.OriginX
		grid.OriginY = cfg.OriginY
	}
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}
	grid.StampRadar(cfg.Radar)
	grid.StampEnters(cfg.Enters)

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

	if err := RenderRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, roomID, outPath,
		RoomSVGOptions{
			ShowFOV:  true,
			Sleepads: cfg.Sleepads,
		}); err != nil {
		t.Fatalf("render: %v", err)
	}
	t.Logf("SVG: %s  (grid %d×%d, origin=(%d,%d))",
		outPath, grid.Width, grid.Height, grid.OriginX, grid.OriginY)
}

func TestRenderRoomSVG_D523(t *testing.T) {
	renderRoomLayoutSVG(t, "D523-demo",
		filepath.Join("..", "..", "..", "doc", "layout-D523-demo.json"),
		filepath.Join("..", "..", "..", "doc", "room-D523-demo.svg"),
	)
}

func TestRenderRoomSVG_D5F7(t *testing.T) {
	renderRoomLayoutSVG(t, "D5F7-bathroom",
		filepath.Join("..", "..", "..", "doc", "layout-D5F7-bathroom.json"),
		filepath.Join("..", "..", "..", "doc", "room-D5F7-bathroom.svg"),
	)
}

func TestRenderRoomSVG_09E7(t *testing.T) {
	renderRoomLayoutSVG(t, "09E7-room101",
		filepath.Join("..", "..", "..", "doc", "layout-09E7-room101.json"),
		filepath.Join("..", "..", "..", "doc", "room-09E7-room101.svg"),
	)
}
