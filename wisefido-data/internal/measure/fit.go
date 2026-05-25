package measure

import "sort"

// FitRectangle 轴对齐外接矩形（AABB）。bed/long_sofa intent 用；雷达坐标系下永远 axis-aligned。
func FitRectangle(cells []Cell, minHits int) (Rectangle, bool) {
	filtered := filterByHits(cells, minHits)
	if len(filtered) < 3 {
		return Rectangle{}, false
	}
	b := cellsBBox(filtered)
	return Rectangle{
		X:      b.MinX - CellSize/2,
		Y:      b.MinY - CellSize/2,
		Width:  b.MaxX - b.MinX + CellSize,
		Height: b.MaxY - b.MinY + CellSize,
	}, true
}

// FitPolygon 凸包（Andrew's monotone chain）。wall intent 用：best-effort，L 形房间需人工调整。
func FitPolygon(cells []Cell, minHits int) (Polygon, bool) {
	filtered := filterByHits(cells, minHits)
	if len(filtered) < 3 {
		return Polygon{}, false
	}
	pts := make([]Point, len(filtered))
	for i, c := range filtered {
		pts[i] = c.Center()
	}
	sort.Slice(pts, func(i, j int) bool {
		if pts[i].X != pts[j].X {
			return pts[i].X < pts[j].X
		}
		return pts[i].Y < pts[j].Y
	})
	n := len(pts)
	hull := make([]Point, 0, 2*n)
	for _, p := range pts {
		for len(hull) >= 2 && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	lower := len(hull) + 1
	for i := n - 2; i >= 0; i-- {
		p := pts[i]
		for len(hull) >= lower && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	if len(hull) > 0 {
		hull = hull[:len(hull)-1]
	}
	if len(hull) < 3 {
		return Polygon{}, false
	}
	return Polygon{Vertices: hull}, true
}

// FitEntryLines 红色 cell 连通簇 → 线段集合。enter intent 用。
// 簇内顶点取 bbox 对角作为线段端点（粗糙够用，雷达 ±10cm 精度匹配）。
func FitEntryLines(cells []Cell, minHits int) []LineSegment {
	filtered := filterByHits(cells, minHits)
	if len(filtered) == 0 {
		return nil
	}
	clusters := connectedClusters(filtered)
	out := make([]LineSegment, 0, len(clusters))
	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}
		b := cellsBBox(cluster)
		w := b.MaxX - b.MinX
		h := b.MaxY - b.MinY
		if w >= h {
			out = append(out,
				LineSegment{Start: Point{X: b.MinX, Y: (b.MinY + b.MaxY) / 2},
					End: Point{X: b.MaxX, Y: (b.MinY + b.MaxY) / 2}})
		} else {
			out = append(out,
				LineSegment{Start: Point{X: (b.MinX + b.MaxX) / 2, Y: b.MinY},
					End: Point{X: (b.MinX + b.MaxX) / 2, Y: b.MaxY}})
		}
	}
	return out
}

func filterByHits(cells []Cell, minHits int) []Cell {
	if minHits <= 1 {
		return cells
	}
	out := make([]Cell, 0, len(cells))
	for _, c := range cells {
		if c.HitCount >= minHits {
			out = append(out, c)
		}
	}
	return out
}

func cross(o, a, b Point) int {
	return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}

// connectedClusters 8-邻接连通分量；cell 距离 ≤ √2 算邻居。
func connectedClusters(cells []Cell) [][]Cell {
	idx := make(map[[2]int]int, len(cells))
	for i, c := range cells {
		idx[[2]int{c.Col, c.Row}] = i
	}
	visited := make([]bool, len(cells))
	var out [][]Cell
	for start := range cells {
		if visited[start] {
			continue
		}
		queue := []int{start}
		visited[start] = true
		var cluster []Cell
		for len(queue) > 0 {
			i := queue[0]
			queue = queue[1:]
			cluster = append(cluster, cells[i])
			c := cells[i]
			for dc := -1; dc <= 1; dc++ {
				for dr := -1; dr <= 1; dr++ {
					if dc == 0 && dr == 0 {
						continue
					}
					if j, ok := idx[[2]int{c.Col + dc, c.Row + dr}]; ok && !visited[j] {
						visited[j] = true
						queue = append(queue, j)
					}
				}
			}
		}
		out = append(out, cluster)
	}
	return out
}
