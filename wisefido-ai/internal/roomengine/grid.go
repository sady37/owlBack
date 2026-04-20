package roomengine

// RoomGrid 房间的 2D 网格（"底片"），每个格子 CellSize×CellSize cm。
// 坐标系：房间坐标（cm），左下角为原点。
type RoomGrid struct {
	CellSize float64 // 格子边长，cm（默认 10）
	Width    int     // 列数（X 方向）
	Height   int     // 行数（Y 方向）
	RoomW    float64 // 房间宽度 cm
	RoomH    float64 // 房间深度 cm
	Cells    []Cell  // row-major: Cells[row*Width + col]

	// 入口点列表（门坐标、FOV 边缘采样点），用于 5 因子中的 d_edge 计算。
	Entries []Point
}

// Point 二维坐标（cm）。
type Point struct {
	X, Y float64
}

// NewRoomGrid 创建空白网格。roomW/roomH 单位 cm，cellSize 单位 cm。
func NewRoomGrid(roomW, roomH, cellSize float64) *RoomGrid {
	if cellSize <= 0 {
		cellSize = 10
	}
	w := int(roomW/cellSize) + 1
	h := int(roomH/cellSize) + 1
	cells := make([]Cell, w*h)
	// 默认所有格子为空地、可行走
	for i := range cells {
		cells[i] = Cell{Type: CellOpen, Walkable: true}
	}
	return &RoomGrid{
		CellSize: cellSize,
		Width:    w,
		Height:   h,
		RoomW:    roomW,
		RoomH:    roomH,
		Cells:    cells,
	}
}

// CellAt 返回房间坐标 (x,y) cm 对应的格子指针。越界返回 nil。
func (g *RoomGrid) CellAt(x, y float64) *Cell {
	col, row := g.ToIndex(x, y)
	if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
		return nil
	}
	return &g.Cells[row*g.Width+col]
}

// ToIndex 房间坐标 → 格子行列索引。
func (g *RoomGrid) ToIndex(x, y float64) (col, row int) {
	col = int(x / g.CellSize)
	row = int(y / g.CellSize)
	return
}

// ToRoom 格子行列索引 → 格子中心的房间坐标。
func (g *RoomGrid) ToRoom(col, row int) (x, y float64) {
	x = (float64(col) + 0.5) * g.CellSize
	y = (float64(row) + 0.5) * g.CellSize
	return
}

// NearestEntryDist 返回 (x,y) 到最近入口点的距离（cm）。无入口返回 9999。
func (g *RoomGrid) NearestEntryDist(x, y float64) float64 {
	best := 9999.0
	for _, e := range g.Entries {
		d := dist(x, y, e.X, e.Y)
		if d < best {
			best = d
		}
	}
	return best
}

// IsEdge 判断 (x,y) 是否在 FOV/房间边缘附近（距边界 < margin cm）。
func (g *RoomGrid) IsEdge(x, y, margin float64) bool {
	return x < margin || y < margin || x > g.RoomW-margin || y > g.RoomH-margin
}

// DecayAll 对所有格子做时间衰减。dtSec 为距上次衰减的秒数。
func (g *RoomGrid) DecayAll(dtSec, halfLifeSec float64) {
	for i := range g.Cells {
		g.Cells[i].Decay(dtSec, halfLifeSec)
	}
}

// SetCellType 批量设置区域类型（启动时从 layout 加载）。
// rect 为房间坐标的矩形区域 [x1,y1, x2,y2] cm。
func (g *RoomGrid) SetCellType(x1, y1, x2, y2 float64, ct CellType, walkable bool) {
	c1, r1 := g.ToIndex(x1, y1)
	c2, r2 := g.ToIndex(x2, y2)
	if c1 > c2 {
		c1, c2 = c2, c1
	}
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if r >= 0 && r < g.Height && c >= 0 && c < g.Width {
				cell := &g.Cells[r*g.Width+c]
				cell.Type = ct
				cell.Walkable = walkable
			}
		}
	}
}

// MarkRealVisit 真实 track 经过格子，更新 RealDecay 和 Flow。
func (g *RoomGrid) MarkRealVisit(x, y, vx, vy float64, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.RealDecay += 1.0
	// 指数移动平均更新 flow 方向（alpha=0.1）
	const alpha = 0.1
	c.FlowX = c.FlowX*(1-alpha) + vx*alpha
	c.FlowY = c.FlowY*(1-alpha) + vy*alpha
	c.LastUpdateMs = nowMs
}

// MarkGhostBirth ghost track 出生在此格子。
func (g *RoomGrid) MarkGhostBirth(x, y float64, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.GhostDecay += 1.0
	c.LastUpdateMs = nowMs
}

// MarkStill track 在此格子静止 dtSec 秒。
func (g *RoomGrid) MarkStill(x, y float64, dtSec float64, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.StillCount += dtSec
	c.LastUpdateMs = nowMs
}

// MarkPoseMismatch pose 与运动学矛盾，在此格子累加。
func (g *RoomGrid) MarkPoseMismatch(x, y float64, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.PoseMismatch += 1.0
	c.LastUpdateMs = nowMs
}
