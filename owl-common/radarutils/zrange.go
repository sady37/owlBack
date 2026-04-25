package radarutils

// Z 方向可探测范围（简化版）。
//
// 本期简化：
//   MaxZAt = installHeight (Center.Z)   // 目标不可能高于雷达本身
//   MinZAt = 0                          // 地面
//
// 二期精细化：用 ElevationFOV + installHeight 算锥形，按 (x,y) 点的水平距离 r 反算：
//   Ceiling: MaxZ = h - r / tan(efov/2)
//   Wall:    MaxZ = h + d * tan(efov/2), MinZ = max(0, h - d * tan(efov/2))
// 见 git history 的精细版本。

// MaxZAt 返回在画布 (x, y) 处雷达可探测目标的最大 Z（cm）。
// 简化：恒等于 installHeight。未安装参数时返回 0。
func (m RadarMount) MaxZAt(x, y int) int {
	if m.Center.Z <= 0 {
		return 0
	}
	return m.Center.Z
}

// MinZAt 返回在画布 (x, y) 处雷达可探测目标的最小 Z（cm）。
// 简化：恒 0（地面）。
func (m RadarMount) MinZAt(x, y int) int {
	return 0
}
