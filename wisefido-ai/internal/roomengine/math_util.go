package roomengine

import "math"

// dist 二维欧几里得距离。
func dist(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// clamp 将 v 限制在 [lo, hi] 范围内。
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampInt 将 v 限制在 [lo, hi] 范围内（int 版本）。
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// distInt 画布坐标（cm 整数）下两点欧氏距离，四舍五入取整。
func distInt(x1, y1, x2, y2 int) int {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return int(math.Round(math.Sqrt(dx*dx + dy*dy)))
}
