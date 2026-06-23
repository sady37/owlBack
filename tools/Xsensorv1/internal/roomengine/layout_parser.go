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
			ID          string          `json:"id"`
			TypeName    string          `json:"typeName"`
			Geometry    json.RawMessage `json:"geometry"`
			Angle       *float64        `json:"angle,omitempty"`
			Device      json.RawMessage `json:"device,omitempty"`
			Height      *int            `json:"height,omitempty"`      // 物体顶部高度 cm；缺失时按 typeName 默认值
			EnterTarget string          `json:"enter_target,omitempty"` // sensor_v2 决定 15：""/inside_enter / "outside" / "bathroom"
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
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Enters = append(cfg.Enters, *rect)
				cfg.EnterHeights = append(cfg.EnterHeights, objHeight)
				cfg.EnterTargets = append(cfg.EnterTargets, hdr.EnterTarget) // sensor_v2 决定 15
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Bed", "MonitorBed", "LongSofa":
			// LongSofa = 无 sleepad 的床：可长躺排 fall、占用/vital 走 radar，与床同处理。
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Beds = append(cfg.Beds, *rect)
				cfg.BedHeights = append(cfg.BedHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Toilet":
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Toilets = append(cfg.Toilets, *rect)
				cfg.ToiletHeights = append(cfg.ToiletHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Shower":
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Showers = append(cfg.Showers, *rect)
				cfg.ShowerHeights = append(cfg.ShowerHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Chair":
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Chairs = append(cfg.Chairs, *rect)
				cfg.ChairHeights = append(cfg.ChairHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Furniture", "Table", "Other":
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
				cfg.Furnitures = append(cfg.Furnitures, *rect)
				cfg.FurnitureHeights = append(cfg.FurnitureHeights, objHeight)
				allObjectPoints = append(allObjectPoints, rectCorners(*rect)...)
			}

		case "Interfere", "MetalCan", "WheelChair", "GlassTV", "Curtain":
			// 干扰/反射/金属 → AreaDeny
			// 注意：吊灯也走 Interfere，但 height>200 表示空中不阻挡通行（未来 RoomEngine 用）
			if rect := parseRectFromGeometry(hdr.Geometry, hdr.Angle); rect != nil {
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

	// 床区 area_id 不在此产出：换源为固件活体 declare_area（bootstrap 走 wisefido-data HTTP
	// 覆盖 cfg.BedAreaIDs），治 canvas 下发区域 vs 固件几何漂移。

	// 根据 Wall 顶点围出多边形（bbox）
	cfg.WallPolygon = buildWallPolygon(wallPoints, cfg.Enters)
	// 客户未画 Wall（顶点 < 2）时用雷达 Boundary 顶点作为兜底多边形，
	// 让 cell.InRoom 至少等于 InFOV——否则所有 InRoom-依赖逻辑（cell_learning /
	// birth scoring isOutdoorButInFOV / room_svg）退化失效。
	// trade-off：boundary 是设备配置矩形而非物理墙，若 boundary 设得过大会让墙外
	// ghost 漏判 !InRoom——但客户没画 wall 本来就是放弃这层防御，兜底比裸跑好。
	//
	// 重要：BoundaryVertices 返回 [右上,左上,右下,左下] 不是闭合简单多边形（直接连线是
	// 蝴蝶结），StampRoomPolygon 走 PolygonContains 射线算法会判错。Ceiling/Wall 重排成
	// [右上,左上,左下,右下]；Corn 菱形重排成 [雷达顶, R侧, 底, L侧]——与
	// owl-common/radarutils.ContainsCanvas + signal.go::SignalEdgeDist 保持一致。
	if cfg.WallPolygon == nil {
		cfg.WallPolygon = boundaryPolygonForStamp(cfg.Radar)
	}

	// 客户未画 Enter 时沿 WallPolygon 周长生成隐式 Enter 薄带，让 NearestEntryDist
	// 在 wall 边缘范围内回 0，lost-fall 的 "贴在门口正常通过" 检测（≤30cm）能继续
	// 工作。否则 wall-fit-no-enter 的房间所有 track 失锁都会被判为非门口消失 →
	// lost-fall pending 池被无关 track 灌满。
	// trade-off：把整个 wall perimeter 当作 entry，比真画门口宽很多 → over-skip
	// 部分 lost-fall。但客户没画 enter 本来就放弃这层精度，兜底比裸跑好。
	if len(cfg.Enters) == 0 && len(cfg.WallPolygon) >= 3 {
		cfg.Enters = deriveImplicitEntersFromPolygon(cfg.WallPolygon, exitDistMinCm)
		cfg.EnterHeights = make([]int, len(cfg.Enters))
		cfg.EnterTargets = make([]string, len(cfg.Enters)) // "" = inside_enter 默认
		for i := range cfg.Enters {
			cfg.EnterHeights[i] = defaultHeightForType("Enter")
			allObjectPoints = append(allObjectPoints, rectCorners(cfg.Enters[i])...)
		}
	}

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
	case "LongSofa":
		return 40
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
// canvas 存的是 PRE-rotation 轴对齐顶点 + 单独 angle（FE 画图时 ctx.rotate(-angle) 才转），
// 故必须按 angle 旋转顶点再取 AABB；否则旋转矩形（90/270 长宽互换、任意角朝向错）会贴歪先验。
func parseRectFromGeometry(geom json.RawMessage, angle *float64) *radarutils.Rect {
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
		pts = rotatePointsAroundCenter(pts, angle)
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

// rotatePointsAroundCenter 把顶点绕其几何中心旋转 angle 度（画布坐标系，与 FE drawObjects 对齐）。
// nil/0 不转。旋转矩形的 AABB 与旋转方向无关（宽=w|cosθ|+h|sinθ|），故 angle 正负不影响 FromVertices 结果。
func rotatePointsAroundCenter(pts []radarutils.Point, angle *float64) []radarutils.Point {
	if angle == nil || len(pts) == 0 || math.Mod(*angle, 360) == 0 {
		return pts
	}
	var cx, cy float64
	for _, p := range pts {
		cx += float64(p.X)
		cy += float64(p.Y)
	}
	n := float64(len(pts))
	cx /= n
	cy /= n
	rad := *angle * math.Pi / 180
	c, s := math.Cos(rad), math.Sin(rad)
	out := make([]radarutils.Point, len(pts))
	for i, p := range pts {
		dx := float64(p.X) - cx
		dy := float64(p.Y) - cy
		out[i] = radarutils.Point{
			X: int(math.Round(cx + dx*c - dy*s)),
			Y: int(math.Round(cy + dx*s + dy*c)),
		}
	}
	return out
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

// deriveImplicitEntersFromPolygon 沿多边形周长生成隐式 Enter 薄带。
// exitDistMinCm 隐式入口推导厚度（cm）：墙多边形上距最近出口 ≤此值算"贴近门口"。
const exitDistMinCm = 30

// 每条边 → 一个 AABB rect 含端点 + thickness 余量。
// 用于客户未画 Enter 的兜底：墙边 ≤ thickness 的 cell 被算作 entry 区，
// NearestEntryDist 在此范围内回 0，"贴边 30cm" 的检测覆盖整圈周长。
func deriveImplicitEntersFromPolygon(poly []radarutils.Point, thickness int) []radarutils.Rect {
	n := len(poly)
	if n < 3 {
		return nil
	}
	out := make([]radarutils.Rect, 0, n)
	for i := 0; i < n; i++ {
		p1 := poly[i]
		p2 := poly[(i+1)%n]
		minX, maxX := p1.X, p2.X
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		minY, maxY := p1.Y, p2.Y
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		out = append(out, radarutils.Rect{
			X1: minX - thickness,
			Y1: minY - thickness,
			X2: maxX + thickness,
			Y2: maxY + thickness,
		})
	}
	return out
}

// boundaryPolygonForStamp 把 BoundaryVertices 重排成 PolygonContains 可用的闭合简单多边形。
//
// owl-common BoundaryVertices 返回的顶点序不是周长走法（直接连线是蝴蝶结）：
//   - Ceiling/Wall: [右上, 左上, 右下, 左下]   → 重排为 [右上, 左上, 左下, 右下]
//   - Corn:         [R侧, 雷达顶, 底, L侧]    → 重排为 [雷达顶, R侧, 底, L侧]
//
// 与 owl-common/radarutils.ContainsCanvas + signal.go::SignalEdgeDist 同源（它们也内嵌
// 重排）；这里独立一份避免循环依赖（roomengine 不依赖 ContainsCanvas，只要顶点）。
func boundaryPolygonForStamp(m radarutils.RadarMount) []radarutils.Point {
	v := radarutils.BoundaryVertices(m)
	if len(v) < 4 {
		return nil
	}
	switch m.InstallModel {
	case radarutils.InstallCeiling, radarutils.InstallWall:
		return []radarutils.Point{v[0], v[1], v[3], v[2]}
	case radarutils.InstallCorn:
		return []radarutils.Point{v[1], v[0], v[2], v[3]}
	default:
		return nil
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
