package radarutils

import "math"

// BoundaryVertices 返回用户设置 Boundary 在画布坐标下的顶点序列（cm，整数）。
// 仅作 UI 参考显示用。真实信号可达域见 signal.go（按物理 FOV 计算）。
//
// 与前端 `getRadarBoundaryVertices` 对齐，按安装模式给出 4 个顶点：
//
//   - Ceiling 矩形： {-rightH,-rearV}/{leftH,-rearV}/{-rightH,frontV}/{leftH,frontV}
//   - Wall 楔形   ： rearV=0 → {-rightH,0}/{leftH,0}/{-rightH,frontV}/{leftH,frontV}
//   - Corn 菱形   ： 雷达在一个顶点，使用 leftH/rightH 作为菱形两臂长
//
// Ceiling/Wall 通过 RadarToCanvas 转换；Corn 菱形不走 RadarToCanvas（独立公式）。
func BoundaryVertices(m RadarMount) []Point {
	b := m.Boundary
	var radarV []RadarPoint

	switch m.InstallModel {
	case InstallCeiling:
		radarV = []RadarPoint{
			{H: -b.RightH, V: -b.RearV},
			{H: b.LeftH, V: -b.RearV},
			{H: -b.RightH, V: b.FrontV},
			{H: b.LeftH, V: b.FrontV},
		}
	case InstallWall:
		// 贴墙：rearV=0，雷达紧贴墙面
		radarV = []RadarPoint{
			{H: -b.RightH, V: 0},
			{H: b.LeftH, V: 0},
			{H: -b.RightH, V: b.FrontV},
			{H: b.LeftH, V: b.FrontV},
		}
	case InstallCorn:
		return cornVertices(m)
	default:
		return nil
	}

	out := make([]Point, len(radarV))
	for i, rp := range radarV {
		out[i] = RadarToCanvas(rp, m)
	}
	return out
}

// cornVertices Corn（墙角）菱形的 4 个画布顶点。
// 参照前端实现：R 在 +X 侧、L 在 -X 侧，雷达为顶点。
// 旋转用 R(-α)：Canvas Y↓ 导致视觉逆时针需负角。
//
// 顶点索引语义（与前端 getRadarBoundaryVertices/RadarCanvas 一致）：
//   [0] = R 侧（+X 臂）
//   [1] = 顶点（雷达本身）
//   [2] = 底顶点
//   [3] = L 侧（-X 臂）
// 这个顺序**不是**周长走法 —— 渲染/包含判定时调用方需重排为 [1]→[0]→[2]→[3]。
func cornVertices(m RadarMount) []Point {
	L := float64(m.Boundary.LeftH)
	R := float64(m.Boundary.RightH)
	const sqrt2 = 1.4142135623730951

	rad := float64(m.Rotation) * math.Pi / 180
	c := math.Cos(rad)
	s := math.Sin(rad)

	// 菱形 4 顶点（相对雷达中心的 dx/dy；angle=0 时朝下）
	type d struct{ dx, dy float64 }
	diamond := []d{
		{R / sqrt2, R / sqrt2},             // [0] R 侧 (+X)
		{0, 0},                             // [1] 顶点（雷达位置）
		{(R - L) / sqrt2, (R + L) / sqrt2}, // [2] 底顶点
		{-L / sqrt2, L / sqrt2},            // [3] L 侧 (-X)
	}

	out := make([]Point, len(diamond))
	for i, q := range diamond {
		out[i] = Point{
			X: m.Center.X + roundInt(q.dx*c+q.dy*s),
			Y: m.Center.Y + roundInt(-q.dx*s+q.dy*c),
			Z: m.Center.Z,
		}
	}
	return out
}

// ContainsCanvas 判断画布坐标 (x, y) 是否在 Boundary 顶点构成的多边形内。
// 仅 UI 参考判定（判信号真实可达域用 signal.go 的 SignalReachable）。
//
// 重排逻辑与 room_svg.go::drawFOVOutline 保持一致：BoundaryVertices 返回的顺序不是
// 周长走法，PolygonContains 的射线算法需要闭合简单多边形，否则结果错误。
func ContainsCanvas(m RadarMount, x, y int) bool {
	poly := BoundaryVertices(m)
	if len(poly) < 3 {
		return false
	}
	switch m.InstallModel {
	case InstallCeiling, InstallWall:
		// [右上, 左上, 右下, 左下] → [右上, 左上, 左下, 右下]
		poly = []Point{poly[0], poly[1], poly[3], poly[2]}
	case InstallCorn:
		// [R侧, 雷达顶, 底, L侧] → [雷达顶, R侧, 底, L侧]（菱形周长走一圈）
		poly = []Point{poly[1], poly[0], poly[2], poly[3]}
	}
	return PolygonContains(poly, x, y)
}
