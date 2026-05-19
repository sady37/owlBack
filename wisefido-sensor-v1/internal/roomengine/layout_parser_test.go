package roomengine

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseLayoutConfig_D523 用真实 D523 客厅 layout 验证 parser
func TestParseLayoutConfig_D523(t *testing.T) {
	path := filepath.Join("..", "..", "..", "doc", "layout-D523-demo.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("skip: %v", err)
		return
	}

	cfg, err := ParseLayoutConfig("room-d523", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	t.Logf("RoomID=%s  Size=%dx%d cm", cfg.RoomID, cfg.RoomW, cfg.RoomH)
	t.Logf("WallPolygon (bbox 闭合): %+v", cfg.WallPolygon)
	t.Logf("Enters(%d), Beds(%d), Toilets(%d), Chairs(%d), Furnitures(%d), Interferes(%d)",
		len(cfg.Enters), len(cfg.Beds), len(cfg.Toilets),
		len(cfg.Chairs), len(cfg.Furnitures), len(cfg.Interferes))
	t.Logf("Radar: center=%+v rotation=%d install=%d boundary=%+v hfov=%d vfov=%d",
		cfg.Radar.Center, cfg.Radar.Rotation, cfg.Radar.InstallModel,
		cfg.Radar.Boundary, cfg.Radar.HFOV, cfg.Radar.VFOV)

	// 基本断言
	if cfg.Radar.Center.X == 0 && cfg.Radar.Center.Y == 0 {
		t.Error("radar center not parsed")
	}
	if len(cfg.WallPolygon) < 4 {
		t.Errorf("wall polygon expected >=4 vertices, got %d", len(cfg.WallPolygon))
	}
	if len(cfg.Enters) == 0 && len(cfg.Furnitures) == 0 {
		t.Error("expected at least some furniture or entries")
	}
}

// TestParseLayoutConfig_D5F7 用 bathroom layout 验证 parser
func TestParseLayoutConfig_D5F7(t *testing.T) {
	path := filepath.Join("..", "..", "..", "doc", "layout-D5F7-bathroom.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("skip: %v", err)
		return
	}

	cfg, err := ParseLayoutConfig("room-d5f7", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	t.Logf("RoomID=%s  Size=%dx%d cm", cfg.RoomID, cfg.RoomW, cfg.RoomH)
	t.Logf("WallPolygon: %+v", cfg.WallPolygon)
	t.Logf("Enters(%d), Beds(%d), Toilets(%d), Chairs(%d), Furnitures(%d), Interferes(%d)",
		len(cfg.Enters), len(cfg.Beds), len(cfg.Toilets),
		len(cfg.Chairs), len(cfg.Furnitures), len(cfg.Interferes))
	t.Logf("Radar: center=%+v rotation=%d install=%d boundary=%+v",
		cfg.Radar.Center, cfg.Radar.Rotation, cfg.Radar.InstallModel, cfg.Radar.Boundary)

	// D5F7 应该 ceiling 模式
	if cfg.Radar.InstallModel != 0 /* ceiling */ {
		t.Errorf("expected ceiling (0), got %d", cfg.Radar.InstallModel)
	}
	// 应该有 Interfere（镜子）
	if len(cfg.Interferes) == 0 {
		t.Error("expected at least one Interfere (mirror)")
	}
}

