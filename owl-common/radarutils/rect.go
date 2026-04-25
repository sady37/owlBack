package radarutils

import "math"

// Contains 判断点是否在矩形内部或边上。
func (r Rect) Contains(x, y int) bool {
	r = r.Norm()
	return x >= r.X1 && x <= r.X2 && y >= r.Y1 && y <= r.Y2
}

// DistTo 点到矩形的最近距离（cm，整数）。矩形内部返回 0。
// 等价于 max(0, x - x2, x1 - x) 与 y 方向分量的欧氏距离。
func (r Rect) DistTo(x, y int) int {
	r = r.Norm()
	dx := maxI(maxI(r.X1-x, 0), x-r.X2)
	dy := maxI(maxI(r.Y1-y, 0), y-r.Y2)
	return roundInt(math.Hypot(float64(dx), float64(dy)))
}

// Center 矩形中心点（画布坐标）。
func (r Rect) Center() Point {
	r = r.Norm()
	return Point{X: (r.X1 + r.X2) / 2, Y: (r.Y1 + r.Y2) / 2}
}

// Width 矩形宽度（cm）。
func (r Rect) Width() int {
	r = r.Norm()
	return r.X2 - r.X1
}

// Height 矩形高度（cm）。
func (r Rect) Height() int {
	r = r.Norm()
	return r.Y2 - r.Y1
}

// FromVertices 从 4 个顶点（任意顺序）构造轴对齐矩形的 Bounding Box。
// 用于把前端 Geometry.rectangle 的 vertices 解析成 Rect。
// 旋转矩形会返回其外接矩形（精度粗化，但对 10cm 量化足够）。
func FromVertices(pts []Point) Rect {
	if len(pts) == 0 {
		return Rect{}
	}
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := pts[0].X, pts[0].Y
	for _, p := range pts[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return Rect{X1: minX, Y1: minY, X2: maxX, Y2: maxY}
}

// PolygonContains 射线法判断点是否在任意多边形内（顶点序列，首尾自动闭合）。
// 用于 Wall 多边形、旋转矩形、扇形近似等非轴对齐区域。
func PolygonContains(poly []Point, x, y int) bool {
	n := len(poly)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := poly[i].X, poly[i].Y
		xj, yj := poly[j].X, poly[j].Y
		// 射线向右（x+∞）穿过边的条件
		if (yi > y) != (yj > y) {
			// 用 float 做除法避免 int 截断错误
			xCross := float64(xj-xi)*float64(y-yi)/float64(yj-yi) + float64(xi)
			if float64(x) < xCross {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// maxI 返回两个 int 中的较大者。
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
