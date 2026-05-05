package roomengine

import (
	"encoding/json"
	"fmt"
	"math"

	"owl-common/radarutils"
)

// ParseLayoutConfig 把 room.layout_config JSONB 解析为 RoomConfig。
//
// 输入 JSON 结构（对齐前端 objects.ts saveCanvas）：
//
//	{
//	  "objects": [BaseObject, ...],
//	  "radar":   {"<objId>": "<deviceId>"},
//	  "sleepad": {...},
//	  "sensor":  {...},
//	  "params":  {...},
//	  "timestamp": "..."
//	}
//
// 每个 BaseObject：
//
//	{
//	  "typeName": "Radar" | "Bed" | "Enter" | "Wall" | "Furniture" | ...,
//	  "geometry": {"type": "point|rectangle|line|...", "data": {...}},
//	  "angle": 0-360,
//	  "device": {"iot": {"radar": {installModel, rotation, boundary, hfov, vfov}}}
//	}
//
// roomID 由调用方传入（从 DB room_id 查表得来）。
func ParseLayoutConfig(roomID string, layoutJSON []byte) (RoomConfig, error) {
	cfg := RoomConfig{RoomID: roomID}

	var layout struct {
		Objects []json.RawMessage `json:"objects"`
	}
	if err := json.Unmarshal(layoutJSON, &layout); err != nil {
		return cfg, fmt.Errorf("unmarshal layout: %w", err)
	}

	var wallPoints []radarutils.Point
	var allObjectPoints []radarutils.Point

	for _, objRaw := range layout.Objects {
		var hdr struct {
			TypeName string          `json:"typeName"`
			Geometry json.RawMessage `json:"geometry"`
			Angle    *float64        `json:"angle,omitempty"`
			Device   json.RawMessage `json:"device,omitempty"`
			Height   *int            `json:"height,omitempty"` // 物体顶部高度 cm；缺失时按 typeName 默认值
		}
		if err := json.Unmarshal(objRaw, &hdr); err != nil {
			continue
		}

		// height：优先 JSON 字段；缺失时取 typeName 默认；都没有则 0
		objHeight := defaultHeightForType(hdr.TypeName)
		if hdr.Height != nil && *hdr.Height >= 0 {
			objHeight = *hdr.Height
		}

		switch hdr.TypeName {
		case "Radar":
			m, err := parseRadarMount(hdr.Geometry, hdr.Angle, hdr.Device)
			if err == nil {
				cfg.Radar = m
			}

		case "Wall":
			// Wall 可能是 line 段或 rectangle —— 把所有涉及的顶点收进来
			// （Wall 自身有高度但当前 RoomEngine 不消费，先不存）
			pts := parseWallPoints(hdr.Geometry)
			wallPoints = append(wallPoints, pts...)
			allObjectPoints = append(allObjectPoints, pts...)

		case "Enter", "Door":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Enters = append(cfg.Enters, *rect)
				cfg.EnterHeights = append(cfg.EnterHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Bed", "MonitorBed":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Beds = append(cfg.Beds, *rect)
				cfg.BedHeights = append(cfg.BedHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Toilet":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Toilets = append(cfg.Toilets, *rect)
				cfg.ToiletHeights = append(cfg.ToiletHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Shower":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Showers = append(cfg.Showers, *rect)
				cfg.ShowerHeights = append(cfg.ShowerHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Chair":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Chairs = append(cfg.Chairs, *rect)
				cfg.ChairHeights = append(cfg.ChairHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Furniture", "Table", "Other":
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Furnitures = append(cfg.Furnitures, *rect)
				cfg.FurnitureHeights = append(cfg.FurnitureHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Interfere", "MetalCan", "WheelChair", "GlassTV", "Curtain":
			// 干扰/反射/金属 → AreaDeny
			// 注意：吊灯也走 Interfere，但 height>200 表示空中不阻挡通行（未来 RoomEngine 用）
			if rect := parseRectFromGeometry(hdr.Geometry); rect != nil {
				cfg.Interferes = append(cfg.Interferes, *rect)
				cfg.InterfereHeights = append(cfg.InterfereHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Sleepad":
			// Sleepad 是 point 几何，记位置供事件路由 + 可视化
			if pt := parsePointFromGeometry(hdr.Geometry); pt != nil {
				cfg.Sleepads = append(cfg.Sleepads, *pt)
				allObjectPoints = append(allObjectPoints, *pt)
			}
		}
	}

	// 根据 Wall 顶点围出多边形（bbox）
	cfg.WallPolygon = buildWallPolygon(wallPoints, cfg.Enters)

	// 推算房间画布范围（覆盖所有物体 + 留 margin）
	cfg.RoomW, cfg.RoomH, cfg.OriginX, cfg.OriginY =
		deriveRoomBounds(cfg.WallPolygon, allObjectPoints)

	return cfg, nil
}

// ------------------------------------------------------------------------
// Radar 解析
// ------------------------------------------------------------------------

// defaultHeightForType 与前端 owlFront/src/utils/radar/types.ts::FURNITURE_CONFIGS[type].defaultHeight 对齐。
// layout JSON 缺 height 字段时（老数据 / 前端没填）的兜底。
func defaultHeightForType(typeName string) int {
	switch typeName {
	case "Bed", "MonitorBed":
		return 60
	case "Interfere":
		return 120 // 默认洗手台/镜子；空中吊灯由前端 Toolbar 改 240+
	case "Enter", "Door":
		return 0 // 门洞地面到顶，不阻挡通行
	case "Wall":
		return 240
	case "Furniture", "Other", "MetalCan":
		return 80
	case "GlassTV":
		return 120
	case "Table":
		return 75
	case "Chair":
		return 90
	case "Curtain":
		return 240
	case "WheelChair":
		return 100
	case "Toilet", "Shower":
		return 0 // 当前未在 typescript 配置中；保守取地面
	}
	return 0
}

func parseRadarMount(geom json.RawMessage, outerAngle *float64, device json.RawMessage) (radarutils.RadarMount, error) {
	var m radarutils.RadarMount

	// 1) 位置（geometry.point）
	var geo struct {
		Type string `json:"type"`
		Data struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"data"`
	}
	if err := json.Unmarshal(geom, &geo); err != nil {
		return m, fmt.Errorf("radar geometry: %w", err)
	}
	if geo.Type != "point" {
		return m, fmt.Errorf("radar geometry must be point, got %s", geo.Type)
	}
	m.Center = radarutils.Point{
		X: int(math.Round(geo.Data.X)),
		Y: int(math.Round(geo.Data.Y)),
		Z: int(math.Round(geo.Data.Z)),
	}

	// 2) 旋转角（优先 object.angle；否则 device.iot.radar.rotation）
	var devWrap struct {
		Iot struct {
			Radar struct {
				InstallModel string `json:"installModel"`
				Rotation     *float64 `json:"rotation"`
				HFOV         *float64 `json:"hfov"`
				VFOV         *float64 `json:"vfov"`
				Boundary     struct {
					LeftH  float64 `json:"leftH"`
					RightH float64 `json:"rightH"`
					FrontV float64 `json:"frontV"`
					RearV  float64 `json:"rearV"`
				} `json:"boundary"`
			} `json:"radar"`
		} `json:"iot"`
	}
	_ = json.Unmarshal(device, &devWrap)
	radar := devWrap.Iot.Radar

	switch {
	case outerAngle != nil:
		m.Rotation = int(math.Round(*outerAngle))
	case radar.Rotation != nil:
		m.Rotation = int(math.Round(*radar.Rotation))
	default:
		m.Rotation = 0
	}

	// 3) InstallModel 字符串 → uint8
	m.InstallModel = parseInstallModel(radar.InstallModel)

	// 4) 物理 FOV
	if radar.HFOV != nil {
		m.HFOV = int(math.Round(*radar.HFOV))
	}
	if radar.VFOV != nil {
		m.VFOV = int(math.Round(*radar.VFOV))
	}

	// 5) Boundary
	m.Boundary = radarutils.Boundary{
		LeftH:  int(math.Round(radar.Boundary.LeftH)),
		RightH: int(math.Round(radar.Boundary.RightH)),
		FrontV: int(math.Round(radar.Boundary.FrontV)),
		RearV:  int(math.Round(radar.Boundary.RearV)),
	}

	return m, nil
}

func parseInstallModel(s string) radarutils.InstallModel {
	switch s {
	case "ceiling":
		return radarutils.InstallCeiling
	case "wall":
		return radarutils.InstallWall
	case "corn":
		return radarutils.InstallCorn
	}
	return radarutils.InstallCeiling // 默认
}

// ------------------------------------------------------------------------
// Rectangle 解析
// ------------------------------------------------------------------------

// parseRectFromGeometry 从 geometry = rectangle 的 data.vertices 构造 Rect。
// 前端 vertices 是 4 点，可能非轴对齐（旋转矩形），这里取 Bounding Box。
func parseRectFromGeometry(geom json.RawMessage) *radarutils.Rect {
	var g struct {
		Type string `json:"type"`
		Data struct {
			Vertices []struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"vertices"`
			// 或者 circle 有 center + radius，但多数家具是 rectangle
			Center struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"center"`
			Radius float64 `json:"radius"`
		} `json:"data"`
	}
	if err := json.Unmarshal(geom, &g); err != nil {
		return nil
	}

	switch g.Type {
	case "rectangle":
		if len(g.Data.Vertices) == 0 {
			return nil
		}
		pts := make([]radarutils.Point, 0, len(g.Data.Vertices))
		for _, v := range g.Data.Vertices {
			pts = append(pts, radarutils.Point{
				X: int(math.Round(v.X)),
				Y: int(math.Round(v.Y)),
			})
		}
		r := radarutils.FromVertices(pts)
		return &r

	case "circle":
		r := int(math.Round(g.Data.Radius))
		return &radarutils.Rect{
			X1: int(math.Round(g.Data.Center.X)) - r,
			Y1: int(math.Round(g.Data.Center.Y)) - r,
			X2: int(math.Round(g.Data.Center.X)) + r,
			Y2: int(math.Round(g.Data.Center.Y)) + r,
		}
	}
	return nil
}

// parsePointFromGeometry 解析 geometry.type=point 的 (x,y,z)
func parsePointFromGeometry(geom json.RawMessage) *radarutils.Point {
	var g struct {
		Type string `json:"type"`
		Data struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"data"`
	}
	if err := json.Unmarshal(geom, &g); err != nil || g.Type != "point" {
		return nil
	}
	return &radarutils.Point{
		X: int(math.Round(g.Data.X)),
		Y: int(math.Round(g.Data.Y)),
		Z: int(math.Round(g.Data.Z)),
	}
}

func rectCorners(r radarutils.Rect) []radarutils.Point {
	return []radarutils.Point{
		{X: r.X1, Y: r.Y1},
		{X: r.X2, Y: r.Y1},
		{X: r.X1, Y: r.Y2},
		{X: r.X2, Y: r.Y2},
	}
}

// ------------------------------------------------------------------------
// Wall 段 → 多边形
// ------------------------------------------------------------------------

// parseWallPoints 把 Wall 的几何（line / rectangle / polygon）解析为顶点集合。
// 统一收进 Point 列表，最后 buildWallPolygon 取 bbox 围矩形。
func parseWallPoints(geom json.RawMessage) []radarutils.Point {
	var g struct {
		Type string `json:"type"`
		Data struct {
			Start    *struct{ X, Y float64 } `json:"start"`
			End      *struct{ X, Y float64 } `json:"end"`
			Vertices []struct{ X, Y float64 } `json:"vertices"`
		} `json:"data"`
	}
	if err := json.Unmarshal(geom, &g); err != nil {
		return nil
	}
	var out []radarutils.Point
	if g.Data.Start != nil {
		out = append(out, radarutils.Point{
			X: int(math.Round(g.Data.Start.X)),
			Y: int(math.Round(g.Data.Start.Y)),
		})
	}
	if g.Data.End != nil {
		out = append(out, radarutils.Point{
			X: int(math.Round(g.Data.End.X)),
			Y: int(math.Round(g.Data.End.Y)),
		})
	}
	for _, v := range g.Data.Vertices {
		out = append(out, radarutils.Point{
			X: int(math.Round(v.X)),
			Y: int(math.Round(v.Y)),
		})
	}
	return out
}

// buildWallPolygon 把收集到的 Wall 顶点 + Enter 矩形角点取 bbox 作为封闭多边形。
//
// 对应**用户判断 1**：即使用户画的 Wall 留了门洞（开放多边形），engine 也用 bbox
// 强制围出封闭房间矩形。Enter 矩形会在 grid.StampEnters 时覆写门洞 cell 的
// InRoom=true 允许穿越。这样 ghost 超出 bbox 立刻判 !InRoom。
//
// 若 Wall 顶点数量 < 2（不够围成 bbox），返回 nil（engine 不判 InRoom，仅靠 InFOV）。
func buildWallPolygon(wallPoints []radarutils.Point, enters []radarutils.Rect) []radarutils.Point {
	if len(wallPoints) < 2 {
		return nil
	}

	// 合并 Wall 顶点 + Enter 矩形角点
	pts := make([]radarutils.Point, 0, len(wallPoints)+len(enters)*4)
	pts = append(pts, wallPoints...)
	for _, r := range enters {
		pts = append(pts, rectCorners(r)...)
	}

	bbox := radarutils.FromVertices(pts)

	// 返回 bbox 4 角（顺时针）形成封闭矩形多边形
	return []radarutils.Point{
		{X: bbox.X1, Y: bbox.Y1}, // 左上
		{X: bbox.X2, Y: bbox.Y1}, // 右上
		{X: bbox.X2, Y: bbox.Y2}, // 右下
		{X: bbox.X1, Y: bbox.Y2}, // 左下
	}
}

// ------------------------------------------------------------------------
// 房间尺寸推算
// ------------------------------------------------------------------------

// deriveRoomBounds 从所有对象的 bounding box 推算 grid 画布范围。
// 返回 (RoomW, RoomH, OriginX, OriginY)：
//   - OriginX/Y 是 grid[0][0] 左上角对应的画布坐标（让 grid 真正覆盖物体）
//   - RoomW/H 是 grid 在画布坐标系下的宽高
//
// 优先用 Wall bbox；为防止物体贴边再额外留 margin（每边 30cm）。
// 最终 clamp 到 [100, MaxRoomWidth/Height]。
func deriveRoomBounds(wallPoly []radarutils.Point, allPoints []radarutils.Point) (int, int, int, int) {
	src := wallPoly
	if len(src) == 0 {
		src = allPoints
	}
	if len(src) == 0 {
		return radarutils.MaxRoomWidth, radarutils.MaxRoomHeight, -radarutils.MaxRoomWidth / 2, 0
	}

	bbox := radarutils.FromVertices(src)
	const margin = 30 // 边缘 margin cm

	originX := bbox.X1 - margin
	originY := bbox.Y1 - margin
	w := bbox.X2 - bbox.X1 + 2*margin
	h := bbox.Y2 - bbox.Y1 + 2*margin

	if w < 100 {
		w = 100
	}
	if h < 100 {
		h = 100
	}
	if w > radarutils.MaxRoomWidth {
		w = radarutils.MaxRoomWidth
	}
	if h > radarutils.MaxRoomHeight {
		h = radarutils.MaxRoomHeight
	}
	return w, h, originX, originY
}
