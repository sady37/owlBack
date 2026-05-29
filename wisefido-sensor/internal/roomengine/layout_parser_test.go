package roomengine

import (
	"os"
	"path/filepath"
	"testing"

	"owl-common/radarutils"
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

// TestDeriveImplicitEntersFromPolygon — 单元测试隐式 Enter 几何
func TestDeriveImplicitEntersFromPolygon(t *testing.T) {
	// 4-vertex 闭合矩形多边形（典型 wall polygon），周长 4 条边 → 4 个 Enter strip
	poly := []radarutils.Point{
		{X: 0, Y: 0},     // 左上
		{X: 400, Y: 0},   // 右上
		{X: 400, Y: 300}, // 右下
		{X: 0, Y: 300},   // 左下
	}
	enters := deriveImplicitEntersFromPolygon(poly, 30)
	if len(enters) != 4 {
		t.Fatalf("expected 4 strips, got %d", len(enters))
	}
	// 顶边 strip: y ∈ [-30, 30], x ∈ [-30, 430]
	if enters[0].X1 != -30 || enters[0].Y1 != -30 || enters[0].X2 != 430 || enters[0].Y2 != 30 {
		t.Errorf("top strip wrong: %+v", enters[0])
	}
	// 右边 strip: x ∈ [370, 430], y ∈ [-30, 330]
	if enters[1].X1 != 370 || enters[1].Y1 != -30 || enters[1].X2 != 430 || enters[1].Y2 != 330 {
		t.Errorf("right strip wrong: %+v", enters[1])
	}
}

// TestParseLayoutConfig_ImplicitEnter — 无 Enter 时自动从 WallPolygon 生成
func TestParseLayoutConfig_ImplicitEnter(t *testing.T) {
	// 极简 layout：1 个 Radar + 1 个 Wall rect，无 Enter
	raw := []byte(`{
		"objects": [
			{
				"typeName": "Radar",
				"geometry": {"type": "point", "data": {"x": 200, "y": 150, "z": 280}},
				"angle": 0,
				"device": {"iot": {"radar": {
					"installModel": "ceiling",
					"boundary": {"leftH": 300, "rightH": 300, "frontV": 200, "rearV": 200}
				}}}
			},
			{
				"typeName": "Wall",
				"geometry": {"type": "rectangle", "data": {"vertices": [
					{"x": 0, "y": 0}, {"x": 400, "y": 0},
					{"x": 0, "y": 300}, {"x": 400, "y": 300}
				]}}
			}
		]
	}`)
	cfg, err := ParseLayoutConfig("test-implicit", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cfg.Enters) == 0 {
		t.Fatal("expected implicit Enters when none drawn, got 0")
	}
	if len(cfg.Enters) != len(cfg.EnterHeights) || len(cfg.Enters) != len(cfg.EnterTargets) {
		t.Errorf("Enter arrays length mismatch: rects=%d heights=%d targets=%d",
			len(cfg.Enters), len(cfg.EnterHeights), len(cfg.EnterTargets))
	}
	// 每个 target 应该是 "" (inside_enter 默认)
	for i, tg := range cfg.EnterTargets {
		if tg != "" {
			t.Errorf("implicit Enter[%d].target expected \"\", got %q", i, tg)
		}
	}
	t.Logf("derived %d implicit Enter strips from WallPolygon", len(cfg.Enters))
}

// TestParseLayoutConfig_ExplicitEnter_NoImplicit — 用户已画 Enter 时不应叠加隐式
func TestParseLayoutConfig_ExplicitEnter_NoImplicit(t *testing.T) {
	raw := []byte(`{
		"objects": [
			{
				"typeName": "Radar",
				"geometry": {"type": "point", "data": {"x": 200, "y": 150, "z": 280}},
				"angle": 0,
				"device": {"iot": {"radar": {
					"installModel": "ceiling",
					"boundary": {"leftH": 300, "rightH": 300, "frontV": 200, "rearV": 200}
				}}}
			},
			{
				"typeName": "Wall",
				"geometry": {"type": "rectangle", "data": {"vertices": [
					{"x": 0, "y": 0}, {"x": 400, "y": 0},
					{"x": 0, "y": 300}, {"x": 400, "y": 300}
				]}}
			},
			{
				"typeName": "Enter",
				"geometry": {"type": "rectangle", "data": {"vertices": [
					{"x": 180, "y": 0}, {"x": 220, "y": 0},
					{"x": 180, "y": 30}, {"x": 220, "y": 30}
				]}}
			}
		]
	}`)
	cfg, err := ParseLayoutConfig("test-explicit", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// 只有 1 个用户画的 Enter，不应该额外叠加 implicit 4 个
	if len(cfg.Enters) != 1 {
		t.Errorf("expected 1 explicit Enter (no implicit overlay), got %d", len(cfg.Enters))
	}
}

