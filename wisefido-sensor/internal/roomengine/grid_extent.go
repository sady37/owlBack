package roomengine

// grid_extent.go
//
// Grid 范围优化：
//   原始 layout JSON 的 canvas 尺寸（RoomW/RoomH）经常远大于实际有效区——
//   layout 编辑器为了好看可能给 600×600 的画布，但房间实际只占 400×500，
//   雷达 FOV 又只覆盖其中一小块。
//
//   优化后的 grid 范围 = bbox(WallPolygon ∪ FOV) + N cell 边距
//   - 边距 4 cell（40cm）：留给 EdgeDist 检测和 ghost 区
//   - 各维度 cap 600cm（与 radarutils.MaxRoomWidth 一致）
//
// 收益：
//   - cell 数大幅下降 → snapshot payload 减小、decay/scan 更快
//   - 不影响 InRoom/InFOV 判定（仍按原 polygon/FOV）
//   - 不影响 LayoutHash 稳定性（只要 layout 几何不变，bbox 不变）

import (
	"owl-common/radarutils"
)

// GridExtentMargin grid bbox 外扩 cell 数（每边）
const GridExtentMargin = 4

// OptimizeGridExtent 基于 WallPolygon + FOV 顶点算出最紧凑的 grid 范围。
// 返回 (roomW, roomH, originX, originY)，单位 cm。
//
// fallback：若 wallPoly + FOV 都没有可用几何，返回 cfgW/cfgH（保持原样）。
func OptimizeGridExtent(wallPoly []radarutils.Point, mount radarutils.RadarMount,
	cellSize, cfgW, cfgH int) (roomW, roomH, originX, originY int) {

	// 收集所有"必须在 grid 内"的点：墙顶点 + FOV 顶点
	var pts []radarutils.Point
	pts = append(pts, wallPoly...)
	pts = append(pts, radarutils.BoundaryVertices(mount)...)

	if len(pts) == 0 {
		// 没几何信息可参考——回退到原 cfg
		return cfgW, cfgH, -cfgW / 2, 0
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

	// 加 4 cell 边距
	margin := cellSize * GridExtentMargin
	minX -= margin
	minY -= margin
	maxX += margin
	maxY += margin

	roomW = maxX - minX
	roomH = maxY - minY

	// cap 600cm（与 radarutils.MaxRoomWidth 一致）
	cap := radarutils.MaxRoomWidth
	if roomW > cap {
		// 居中截断（保留中心区域；雷达通常在中部，截掉远端意义更小）
		excess := roomW - cap
		minX += excess / 2
		roomW = cap
	}
	if roomH > cap {
		excess := roomH - cap
		minY += excess / 2
		roomH = cap
	}

	originX = minX
	originY = minY
	return
}

// ApplyOptimizedExtent 把优化结果写回 RoomConfig（in-place）。
// 调用时机：ParseLayoutConfig 之后 / RegisterRoom 之前。
//
// 不修改 cfg.WallPolygon / 各类矩形 / Radar—— 这些都用画布坐标，与 grid origin 解耦。
func ApplyOptimizedExtent(cfg *RoomConfig) {
	w, h, ox, oy := OptimizeGridExtent(cfg.WallPolygon, cfg.Radar, radarutils.CellSize,
		cfg.RoomW, cfg.RoomH)
	cfg.RoomW = w
	cfg.RoomH = h
	cfg.OriginX = ox
	cfg.OriginY = oy
}
