package roomengine

import (
	"owl-common/radarutils"
)

// RoomGrid 房间的 2D 网格（"底片"），每个格子 CellSize×CellSize cm。
//
// 坐标系：画布坐标（顶部中心为原点，X 左负右正，Y 下正）。
// Grid[0][0] 左上角对应的画布坐标为 (OriginX, OriginY)。
// 默认 OriginX = -RoomW/2, OriginY = 0（与前端 layout_config 对齐）。
//
// 构建流程（在 engine.RegisterRoom 里依次调用）：
//  1. NewRoomGrid(roomW, roomH, cellSize)        // 空网格
//  2. grid.StampRoomPolygon(wallPoly)            // 刷 InRoom
//  3. grid.StampRadar(mount)                      // 刷 InFOV / EdgeDist / MaxZ / MinZ
//  4. grid.StampEnters(enterRects)                // 记 Enters 列表 + 覆写 InRoom
//  5. grid.SetPrior(rect, AreaType, conf, src)    // 人标先验注入
//
// 运行时：
//   CellAt / NearestEntryDist / IsEdge           // 查询
//   MarkOccupancy / MarkPoseTime / ...           // 每帧累积（历史流）
//   DecayAll                                     // 定时衰减
type RoomGrid struct {
	CellSize int
	Width    int
	Height   int
	RoomW    int
	RoomH    int
	OriginX  int
	OriginY  int

	Cells  []Cell
	Enters []radarutils.Rect
}

// NewRoomGrid 创建空网格。默认画布坐标系（顶部中心原点）。
func NewRoomGrid(roomW, roomH, cellSize int) *RoomGrid {
	if cellSize <= 0 {
		cellSize = radarutils.CellSize
	}
	w := roomW/cellSize + 1
	h := roomH/cellSize + 1
	return &RoomGrid{
		CellSize: cellSize,
		Width:    w,
		Height:   h,
		RoomW:    roomW,
		RoomH:    roomH,
		OriginX:  -roomW / 2,
		OriginY:  0,
		Cells:    make([]Cell, w*h),
	}
}

// ToIndex 画布坐标 → 格子索引
func (g *RoomGrid) ToIndex(x, y int) (col, row int) {
	col = (x - g.OriginX) / g.CellSize
	row = (y - g.OriginY) / g.CellSize
	return
}

// CellAt 画布坐标 (x, y) → 格子指针。越界返回 nil。
func (g *RoomGrid) CellAt(x, y int) *Cell {
	col, row := g.ToIndex(x, y)
	if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
		return nil
	}
	return &g.Cells[row*g.Width+col]
}

// ToCanvas 格子索引 → 格子中心画布坐标
func (g *RoomGrid) ToCanvas(col, row int) (x, y int) {
	x = g.OriginX + col*g.CellSize + g.CellSize/2
	y = g.OriginY + row*g.CellSize + g.CellSize/2
	return
}

// IsEdge 画布坐标是否在房间边界附近
func (g *RoomGrid) IsEdge(x, y, margin int) bool {
	return x-g.OriginX < margin ||
		y-g.OriginY < margin ||
		g.OriginX+g.RoomW-x < margin ||
		g.OriginY+g.RoomH-y < margin
}

// ========================================================================
// Rasterize 接口（构建时一次性调用）
// ========================================================================

// StampRoomPolygon 把 Wall 多边形刷到 cells 的 InRoom 字段。
func (g *RoomGrid) StampRoomPolygon(poly []radarutils.Point) {
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cx, cy := g.ToCanvas(col, row)
			if radarutils.PolygonContains(poly, cx, cy) {
				g.Cells[row*g.Width+col].InRoom = true
			}
		}
	}
}

// StampRadar 把雷达物理 FOV 刷到 cells 的 InFOV / EdgeDist / MaxZ / MinZ。
func (g *RoomGrid) StampRadar(mount radarutils.RadarMount) {
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cx, cy := g.ToCanvas(col, row)
			c := &g.Cells[row*g.Width+col]
			c.InFOV = mount.SignalReachable(cx, cy, 0)
			if c.InFOV {
				c.EdgeDist = mount.SignalEdgeDist(cx, cy, 0)
			}
			c.MaxZ = mount.MaxZAt(cx, cy)
			c.MinZ = mount.MinZAt(cx, cy)
		}
	}
}

// StampEnters 记录 Enter 矩形集合，覆写矩形内 cell 的 InRoom=true（门洞可穿越）
func (g *RoomGrid) StampEnters(rects []radarutils.Rect) {
	g.Enters = rects
	for _, rect := range rects {
		rect = rect.Norm()
		c1, r1 := g.ToIndex(rect.X1, rect.Y1)
		c2, r2 := g.ToIndex(rect.X2, rect.Y2)
		for row := r1; row <= r2; row++ {
			if row < 0 || row >= g.Height {
				continue
			}
			for col := c1; col <= c2; col++ {
				if col < 0 || col >= g.Width {
					continue
				}
				g.Cells[row*g.Width+col].InRoom = true
			}
		}
	}
}

// SetPrior 对 rect 内所有 cell 的三组 Belief 统一刷 (AreaType, conf, src)
// 用于 Layout 人标 Bed/Toilet/Enter 等的初始先验注入。
func (g *RoomGrid) SetPrior(rect radarutils.Rect, t AreaType, conf int, src Source) {
	rect = rect.Norm()
	c1, r1 := g.ToIndex(rect.X1, rect.Y1)
	c2, r2 := g.ToIndex(rect.X2, rect.Y2)
	for row := r1; row <= r2; row++ {
		if row < 0 || row >= g.Height {
			continue
		}
		for col := c1; col <= c2; col++ {
			if col < 0 || col >= g.Width {
				continue
			}
			c := &g.Cells[row*g.Width+col]
			for bi := 0; bi < 3; bi++ {
				c.Belief[bi].Type = t
				c.Belief[bi].Confidence = conf
				c.Belief[bi].Source = src
			}
			c.AreaType = t // mirror
		}
	}
}

// NearestEntryDist 到最近 Enter 矩形的距离（cm）
func (g *RoomGrid) NearestEntryDist(x, y int) int {
	if len(g.Enters) == 0 {
		return 9999
	}
	best := 9999
	for _, r := range g.Enters {
		d := r.DistTo(x, y)
		if d < best {
			best = d
		}
	}
	return best
}

// DecayAll 对所有 cell 做时间衰减
func (g *RoomGrid) DecayAll(dtSec, halfLifeSec float64) {
	for i := range g.Cells {
		g.Cells[i].Decay(dtSec, halfLifeSec)
	}
}

// ========================================================================
// Mark 接口（每帧 track 层调用，历史流累积）
// ========================================================================

// qOccupancyThreshold 本帧质量分阈值（Kalman 残差等打分后）
const qOccupancyThreshold = 50

// MarkOccupancy 按本帧质量 q 分桶：q 高 → RealDecay + FlowEMA，q 低 → GhostDecay。
// 历史流不过滤 Verdict——每帧都累积，时间/衰减自淘汰。
func (g *RoomGrid) MarkOccupancy(x, y, quality, vx, vy int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	if quality >= qOccupancyThreshold {
		c.RealDecay++
		// FlowEMA（α=0.1 int 定点）
		c.FlowX = (9*c.FlowX + vx) / 10
		c.FlowY = (9*c.FlowY + vy) / 10
	} else {
		c.GhostDecay++
	}
	c.LastUpdateMs = nowMs
}

// MarkPoseTime 按本帧核心姿态累加 ActiveType 对应槽（饱和到 255）。
// 每帧都调（无过滤）。dtSec 为本帧时长（秒，通常 1）。
func (g *RoomGrid) MarkPoseTime(x, y int, core CorePose, dtSec int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil || dtSec <= 0 {
		return
	}
	idx := CorePoseToActiveIdx(core)
	if idx < 0 {
		return
	}
	c.ActiveType[idx] = satAddUint8(c.ActiveType[idx], dtSec)
	c.LastUpdateMs = nowMs
}

// MarkFallEvent 自推"跌倒事件"触发时调（Stand/Move → Lie + Z 骤降）
func (g *RoomGrid) MarkFallEvent(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.FallEventCount++
	c.LastUpdateMs = nowMs
}

// MarkLieRetract 短暂 Lie 回撤（Stand/Move → Lie → Stand/Move < 3 秒）
func (g *RoomGrid) MarkLieRetract(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.LieRetract++
	c.LastUpdateMs = nowMs
}

// MarkDwell 静止期结束时调，把本次 dwell 秒灌入 DwellEMA
func (g *RoomGrid) MarkDwell(x, y, dwellSec int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.DwellEMA = (4*c.DwellEMA + dwellSec) / 5
	c.LastUpdateMs = nowMs
}

// MarkLongStill 静止超 15min 阈值事件（一次 ++）
func (g *RoomGrid) MarkLongStill(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.LongStillCount++
	c.LastUpdateMs = nowMs
}

// MarkSleepadInBed Sleepad 压床事件（HR/RR/InBed）
func (g *RoomGrid) MarkSleepadInBed(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.SleepadInBedCount++
	c.LastUpdateMs = nowMs
}

// MarkSleepadLeftBed Sleepad 离床事件
func (g *RoomGrid) MarkSleepadLeftBed(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.SleepadLeftBedCount++
	c.LastUpdateMs = nowMs
}

// MarkDoorEvent Enter/Exit 门事件
func (g *RoomGrid) MarkDoorEvent(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.DoorEventCount++
	c.LastUpdateMs = nowMs
}
