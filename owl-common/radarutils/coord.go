package radarutils

import "math"

// RadarToCanvas 雷达本地坐标 (h, v, z) → 画布（房间）坐标 (x, y, z)。
// 与前端 radarUtils.ts `toCanvasCoordinate` 公式一致：
//
//	localX = -h                                // H 左正 → X 右正，取反（镜像）
//	localY = v                                 // V 下正 → Y 下正，保持
//	rotDeg = Rotation + internalRotationDeg    // corn 时 +(-45°)
//	rad    = -rotDeg * π/180                   // 负号：因镜像导致旋转方向反转
//	rotX   = localX·cos(rad) − localY·sin(rad)
//	rotY   = localX·sin(rad) + localY·cos(rad)
//	canvasX = Center.X + rotX
//	canvasY = Center.Y + rotY
//
// Z 直接透传 rp.Z（不做 fallback；贴地 track Z≈0 是 Silent Fall 关键信号）。
//
// 内部用 float64 做三角运算，最后 math.Round 取整返回。
func RadarToCanvas(rp RadarPoint, m RadarMount) Point {
	localX := float64(-rp.H)
	localY := float64(rp.V)

	rad := float64(-m.effectiveRotationDeg()) * math.Pi / 180
	c, s := math.Cos(rad), math.Sin(rad)
	rotX := localX*c - localY*s
	rotY := localX*s + localY*c

	return Point{
		X: m.Center.X + roundInt(rotX),
		Y: m.Center.Y + roundInt(rotY),
		Z: rp.Z,
	}
}

// CanvasToRadar 画布（房间）坐标 (x, y) → 雷达本地坐标 (h, v)。
// RadarToCanvas 的逆运算，与前端 `toRadarCoordinate` 对齐。
// 返回 Z 默认 0（画布 → 雷达本地时 Z 不参与 2D 转换）。
func CanvasToRadar(p Point, m RadarMount) RadarPoint {
	localX := float64(p.X - m.Center.X)
	localY := float64(p.Y - m.Center.Y)

	rad := float64(-m.effectiveRotationDeg()) * math.Pi / 180
	c, s := math.Cos(rad), math.Sin(rad)
	// 旋转矩阵 R(rad) 的转置（= R(-rad)）
	unrotX := localX*c + localY*s
	unrotY := -localX*s + localY*c

	return RadarPoint{
		H: roundInt(-unrotX), // X 右正 → H 左正，取反
		V: roundInt(unrotY),  // Y 下正 → V 下正，保持
		Z: 0,
	}
}

// roundInt 四舍五入取整（内部 helper，接口边界统一用此转换）。
func roundInt(x float64) int {
	return int(math.Round(x))
}
