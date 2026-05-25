// Package measure 计算 radar 测量会话产出的几何拟合（多边形 / 矩形 / 入口线段 / 检测矩形）。
// 全部算法无状态：输入 cell 集合或多边形顶点，输出几何形状。坐标单位 cm。
package measure

const (
	CellSize       = 10
	WallSatHits    = 5
	BedSatHits     = 5
	SofaSatHits    = 5
	EntrySatHits   = 10
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Cell struct {
	Col      int `json:"col"`
	Row      int `json:"row"`
	HitCount int `json:"hit_count"`
}

func (c Cell) Center() Point {
	return Point{X: c.Col*CellSize + CellSize/2, Y: c.Row*CellSize + CellSize/2}
}

type Rectangle struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Width    int `json:"width"`
	Height   int `json:"height"`
	Rotation int `json:"rotation"`
}

type Polygon struct {
	Vertices []Point `json:"vertices"`
}

type LineSegment struct {
	Start Point `json:"start"`
	End   Point `json:"end"`
}

type BBox struct {
	MinX, MinY, MaxX, MaxY int
}

func cellsBBox(cells []Cell) BBox {
	b := BBox{MinX: 1 << 30, MinY: 1 << 30, MaxX: -(1 << 30), MaxY: -(1 << 30)}
	for _, c := range cells {
		p := c.Center()
		if p.X < b.MinX {
			b.MinX = p.X
		}
		if p.X > b.MaxX {
			b.MaxX = p.X
		}
		if p.Y < b.MinY {
			b.MinY = p.Y
		}
		if p.Y > b.MaxY {
			b.MaxY = p.Y
		}
	}
	return b
}

func polygonBBox(p Polygon) BBox {
	b := BBox{MinX: 1 << 30, MinY: 1 << 30, MaxX: -(1 << 30), MaxY: -(1 << 30)}
	for _, v := range p.Vertices {
		if v.X < b.MinX {
			b.MinX = v.X
		}
		if v.X > b.MaxX {
			b.MaxX = v.X
		}
		if v.Y < b.MinY {
			b.MinY = v.Y
		}
		if v.Y > b.MaxY {
			b.MaxY = v.Y
		}
	}
	return b
}

// pointInPolygon ray casting；polygon 顶点按序但可顺/逆时针，开环或闭环皆可。
func pointInPolygon(p Point, poly Polygon) bool {
	n := len(poly.Vertices)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, yj := poly.Vertices[i].Y, poly.Vertices[j].Y
		xi, xj := poly.Vertices[i].X, poly.Vertices[j].X
		if ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}
