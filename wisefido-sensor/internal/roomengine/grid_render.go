package roomengine

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"owl-common/radarutils"
)

// RenderGridPNG 把 grid 渲染成 PNG 图，每 cell 10×10 像素，按 cell 状态着色。
//
// 颜色编码：
//
//	InRoom && InFOV  → 按 Belief[0].Type 染色（白/黄/橙/粉/浅蓝/绿/浅灰）
//	!InRoom && InFOV → 深红（墙外 ghost 区）
//	InRoom && !InFOV → 深灰（房内雷达盲区）
//	!InRoom && !InFOV→ 黑（完全不可达）
//
// 额外叠加：
//   - 红色十字 = Radar 位置
//   - 蓝色虚线 = Enter 矩形边界
//   - 白色虚线 = Wall 多边形
func RenderGridPNG(grid *RoomGrid, mount radarutils.RadarMount, enters []radarutils.Rect,
	wallPoly []radarutils.Point, path string) error {
	// 点状渲染：每 cell 占 20px 格位，但实际色块只画中央 12×12，外围 4px 留白
	// 这样每 cell 视觉上是独立"小方块"，Unknown 大片保持背景色。
	const gridSpan = 20 // 每 cell 格位
	const dotSize = 12  // cell 中央色块尺寸
	const dotOffset = (gridSpan - dotSize) / 2 // 4
	w := grid.Width * gridSpan
	h := grid.Height * gridSpan
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// 背景：浅色
	bg := color.RGBA{248, 248, 250, 255}
	fillRect(img, 0, 0, w, h, bg)

	// 1) 每 cell 只在中央画小方块；**Unknown 的 cell 不画**（保持背景色，视觉上就是"空"）
	for row := 0; row < grid.Height; row++ {
		for col := 0; col < grid.Width; col++ {
			c := &grid.Cells[row*grid.Width+col]
			clr, draw := cellDotColor(c)
			if !draw {
				continue // Unknown / AreaActive 不画点，保持背景
			}
			fillRect(img,
				col*gridSpan+dotOffset, row*gridSpan+dotOffset,
				dotSize, dotSize,
				clr)
		}
	}

	// 每 10 cell（= 100cm）画一条浅网格线作为标尺
	majorLine := color.RGBA{210, 210, 220, 255}
	for row := 0; row <= grid.Height; row += 10 {
		y := row * gridSpan
		if y >= h {
			y = h - 1
		}
		fillRect(img, 0, y, w, 1, majorLine)
	}
	for col := 0; col <= grid.Width; col += 10 {
		x := col * gridSpan
		if x >= w {
			x = w - 1
		}
		fillRect(img, x, 0, 1, h, majorLine)
	}
	// 后续 Wall/Enter/Radar 用同样格位换算
	const cellPx = gridSpan

	// 2) Wall 多边形（深色粗线覆盖在 cell 之上）
	if len(wallPoly) >= 2 {
		for i := 0; i < len(wallPoly); i++ {
			j := (i + 1) % len(wallPoly)
			c1, r1 := grid.ToIndex(wallPoly[i].X, wallPoly[i].Y)
			c2, r2 := grid.ToIndex(wallPoly[j].X, wallPoly[j].Y)
			drawThickLine(img,
				c1*cellPx+cellPx/2, r1*cellPx+cellPx/2,
				c2*cellPx+cellPx/2, r2*cellPx+cellPx/2,
				color.RGBA{20, 20, 20, 255}, 2)
		}
	}

	// 3) Enter 矩形蓝色 2px 边框
	for _, r := range enters {
		r = r.Norm()
		c1, r1 := grid.ToIndex(r.X1, r.Y1)
		c2, r2 := grid.ToIndex(r.X2, r.Y2)
		drawThickRectOutline(img,
			c1*cellPx, r1*cellPx,
			(c2-c1+1)*cellPx, (r2-r1+1)*cellPx,
			color.RGBA{30, 100, 220, 255}, 2)
	}

	// 4) Radar 位置（红色十字 + 圆圈，放大）
	radarCol, radarRow := grid.ToIndex(mount.Center.X, mount.Center.Y)
	rcx := radarCol*cellPx + cellPx/2
	rcy := radarRow*cellPx + cellPx/2
	drawCross(img, rcx, rcy, 12, color.RGBA{255, 0, 0, 255})
	drawCircleOutline(img, rcx, rcy, 16, color.RGBA{255, 0, 0, 255})
	drawCircleOutline(img, rcx, rcy, 17, color.RGBA{255, 0, 0, 255})

	// 保存
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	return png.Encode(f, img)
}

// cellDotColor 按 cell 的 Belief[0].Type 给点色。
// 返回 (color, draw) —— draw=false 表示这个 cell 保持背景色（"Unknown 就是 Unknown"）。
//
// 原则：cell 的身份由 Belief.Type 决定，人标或学习过才有色；没就保持空白。
// InRoom/InFOV 是派生几何属性，**不在点色上体现**（那是另一层可视化）。
func cellDotColor(c *Cell) (color.RGBA, bool) {
	switch c.Belief[0].Type {
	case AreaEnter:
		return color.RGBA{80, 200, 110, 255}, true // 绿 门
	case AreaBed:
		return color.RGBA{220, 200, 100, 255}, true // 黄 床
	case AreaSit:
		return color.RGBA{230, 150, 70, 255}, true // 橙 坐
	case AreaShower:
		return color.RGBA{140, 200, 240, 255}, true // 浅蓝 淋浴
	case AreaToilet:
		return color.RGBA{230, 140, 200, 255}, true // 粉 马桶
	case AreaDeny:
		return color.RGBA{150, 150, 155, 255}, true // 深灰 禁区
	case AreaActive:
		// Active 如果将来由学习得到（e.g. 走廊），也画；初始不赋此类则不画
		return color.RGBA{200, 200, 240, 255}, true
	}
	// AreaUnknown → 不画，保持背景空白
	return color.RGBA{}, false
}

// ------------------------------------------------------------------------
// 绘图 helpers
// ------------------------------------------------------------------------

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	bounds := img.Bounds()
	for py := y; py < y+h && py < bounds.Max.Y; py++ {
		for px := x; px < x+w && px < bounds.Max.X; px++ {
			img.SetRGBA(px, py, c)
		}
	}
}

func drawRectOutline(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for px := x; px < x+w; px++ {
		img.SetRGBA(px, y, c)
		img.SetRGBA(px, y+h-1, c)
	}
	for py := y; py < y+h; py++ {
		img.SetRGBA(x, py, c)
		img.SetRGBA(x+w-1, py, c)
	}
}

func drawCross(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	for d := -size; d <= size; d++ {
		img.SetRGBA(cx+d, cy, c)
		img.SetRGBA(cx, cy+d, c)
	}
}

func drawCircleOutline(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	// Bresenham-ish 简化
	for angle := 0; angle < 360; angle++ {
		rad := float64(angle) * 3.141592653589793 / 180
		x := cx + int(float64(r)*cosF(rad))
		y := cy + int(float64(r)*sinF(rad))
		if img.Bounds().At(x, y) != nil {
			img.SetRGBA(x, y, c)
		}
	}
}

func drawThickLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA, thickness int) {
	for ofs := -thickness / 2; ofs <= thickness/2; ofs++ {
		drawLine(img, x1+ofs, y1, x2+ofs, y2, c)
		drawLine(img, x1, y1+ofs, x2, y2+ofs, c)
	}
}

func drawThickRectOutline(img *image.RGBA, x, y, w, h int, c color.RGBA, thickness int) {
	for t := 0; t < thickness; t++ {
		drawRectOutline(img, x+t, y+t, w-2*t, h-2*t, c)
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	x, y := x1, y1
	for {
		if x >= 0 && y >= 0 && x < img.Bounds().Max.X && y < img.Bounds().Max.Y {
			img.SetRGBA(x, y, c)
		}
		if x == x2 && y == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// 简易 sin/cos（avoid math import conflict）
func sinF(rad float64) float64 {
	// Taylor 到 7 阶
	x := rad
	x3 := x * x * x
	x5 := x3 * x * x
	x7 := x5 * x * x
	return x - x3/6 + x5/120 - x7/5040
}
func cosF(rad float64) float64 {
	x := rad
	x2 := x * x
	x4 := x2 * x2
	x6 := x4 * x2
	return 1 - x2/2 + x4/24 - x6/720
}
