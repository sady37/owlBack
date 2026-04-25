package radarutils

import "math"

// 雷达信号可达域判定（简化版）。
//
// 本期简化策略：
//   - 不用物理 FOV 公式（hfov/vfov/afov/efov/installHeight 算扇形）
//   - 直接用前端画的 Boundary 多边形（ceiling 矩形/wall 楔形/corn 菱形）
//   - 判定：SignalReachable = BoundaryVertices.Contains(x, y)
//   - Engine 最终可达区 = InRoom && InFOV（Boundary ∩ 房间多边形）
//
// 对齐前端 radarUtils.ts getRadarBoundaryVertices + isPointInRadarBoundary。
// 物理 FOV 字段（HFOV/VFOV/AzimuthFOV/ElevationFOV）保留在 RadarMount 但不参与判定，
// 作为二期精细化的数据位。

// SignalReachable 画布 (x, y) 是否在雷达 Boundary 多边形内。
// Z 参数保留但当前不用（二期物理 FOV 才需要考虑高度）。
func (m RadarMount) SignalReachable(x, y, z int) bool {
	return ContainsCanvas(m, x, y)
}

// SignalEdgeDist 画布 (x, y) 到 Boundary 多边形最近边的距离（cm）。
// 在 Boundary 外返回 0，越靠近中心值越大。
// 用于 Silent Fall 第 4 要素："消失位置的 EdgeDist > 30 → 非合法离场"。
func (m RadarMount) SignalEdgeDist(x, y, z int) int {
	poly := BoundaryVertices(m)
	if len(poly) < 3 {
		return 0
	}
	// Ceiling/Wall 的顶点序是 [右上, 左上, 右下, 左下]，闭合需重排为 [0,1,3,2]
	if m.InstallModel == InstallCeiling || m.InstallModel == InstallWall {
		poly = []Point{poly[0], poly[1], poly[3], poly[2]}
	}
	if !PolygonContains(poly, x, y) {
		return 0
	}
	return pointToPolygonEdgeDist(poly, x, y)
}

// IsZReachable 画布 (x, y, z) 处目标是否在 Z 可达范围内。
// 简化版：Z ∈ [0, installHeight]（地面到雷达高度之间）。
func (m RadarMount) IsZReachable(x, y, z int) bool {
	h := m.Center.Z
	if h <= 0 {
		return z == 0
	}
	return z >= 0 && z <= h
}

// pointToPolygonEdgeDist 点到多边形所有边的最短距离（cm，整数）。
// 点在内部：返回最近边的距离；点在外部：也返回最近边的距离（不区分符号）。
func pointToPolygonEdgeDist(poly []Point, x, y int) int {
	n := len(poly)
	if n < 2 {
		return 0
	}
	best := math.Inf(1)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		d := pointToSegmentDist(x, y, poly[i].X, poly[i].Y, poly[j].X, poly[j].Y)
		if d < best {
			best = d
		}
	}
	return int(math.Round(best))
}

// pointToSegmentDist 点 (px, py) 到线段 (ax,ay)-(bx,by) 的最短距离（cm，float）。
func pointToSegmentDist(px, py, ax, ay, bx, by int) float64 {
	dx := float64(bx - ax)
	dy := float64(by - ay)
	lenSq := dx*dx + dy*dy
	if lenSq < 1e-9 {
		ddx := float64(px - ax)
		ddy := float64(py - ay)
		return math.Sqrt(ddx*ddx + ddy*ddy)
	}
	t := (float64(px-ax)*dx + float64(py-ay)*dy) / lenSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx := float64(ax) + t*dx
	cy := float64(ay) + t*dy
	return math.Hypot(float64(px)-cx, float64(py)-cy)
}
