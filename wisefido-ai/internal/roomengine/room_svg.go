package roomengine

import (
	"fmt"
	"os"
	"strings"

	"owl-common/radarutils"
)

// RoomSVGOptions 渲染时的覆盖层选项
type RoomSVGOptions struct {
	ShowFOV     bool               // 画雷达 FOV 蓝色虚线框
	Sleepads    []radarutils.Point // Sleepad 位置（黄色图标）
	LiveTracks  []radarutils.Point // 实时 track 位置（小人图标）
	TrackLabels []string           // 对应 LiveTracks 的标签 (P0/P1...)
	TitleSuffix string             // 标题追加（playback 用，例如 " | 2026-04-25 12:34:56"）
}

// RenderRoomSVG 把 grid 渲染成 SVG 文件。内部委托 BuildRoomSVG 拼字符串再写文件。
func RenderRoomSVG(grid *RoomGrid, mount radarutils.RadarMount, wallPoly []radarutils.Point,
	enters []radarutils.Rect, roomID, path string, opts ...RoomSVGOptions) error {
	svg := BuildRoomSVG(grid, mount, wallPoly, enters, roomID, opts...)
	return os.WriteFile(path, []byte(svg), 0644)
}

// BuildRoomSVG 把 grid 渲染成 SVG 字符串（playback 内联到 HTML 用）。
//
// 6 色语义（用户确定）：
//
//	Unknown : gray   #c8c8c8   未学习/未标注 cell（绝大多数初始状态）
//	Walk    : white  #ffffff   可活动区（AreaActive）
//	Sit     : orange #ff9933   坐姿区（沙发/椅子/马桶 都归 Sit）
//	Lying   : blue   #4488dd   躺姿区（床）
//	Enter   : green  #44cc66   门/入口
//	Deny    : 灰黑   #444444   禁区（墙外/家具/反射区）
func BuildRoomSVG(grid *RoomGrid, mount radarutils.RadarMount, wallPoly []radarutils.Point,
	enters []radarutils.Rect, roomID string, opts ...RoomSVGOptions) string {

	var opt RoomSVGOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var sb strings.Builder

	const (
		canvasW      = 600
		canvasH      = 600
		marginX      = 50
		marginTop    = 40
		marginBottom = 60 // 留底部图例
	)
	viewMinX := -canvasW/2 - marginX
	viewMinY := -marginTop
	viewW := canvasW + 2*marginX
	viewH := canvasH + marginTop + marginBottom

	fmt.Fprintf(&sb,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="%d %d %d %d" `+
			`style="font-family:system-ui,monospace;background:#f4f4f4">`,
		viewMinX, viewMinY, viewW, viewH)

	// 画布外缘（淡灰）
	fmt.Fprintf(&sb,
		`<rect x="%d" y="%d" width="%d" height="%d" fill="#fafafa" stroke="#ccc" stroke-width="1"/>`,
		-canvasW/2, 0, canvasW, canvasH)

	// 1) 渲染所有 cell（6 色铺满）
	cellSize := grid.CellSize
	for row := 0; row < grid.Height; row++ {
		for col := 0; col < grid.Width; col++ {
			c := &grid.Cells[row*grid.Width+col]
			x := grid.OriginX + col*cellSize
			y := grid.OriginY + row*cellSize
			fill := cellSVGFill(c)
			fmt.Fprintf(&sb,
				`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
				x, y, cellSize, cellSize, fill)
		}
	}

	// 2) 100 cm 网格 + 坐标数字
	drawAxes(&sb, canvasW, canvasH)

	// 3) Wall 多边形（黑色粗描边，覆盖 cell 之上）
	if len(wallPoly) >= 3 {
		fmt.Fprintf(&sb, `<polygon points="`)
		for _, p := range wallPoly {
			fmt.Fprintf(&sb, "%d,%d ", p.X, p.Y)
		}
		fmt.Fprintf(&sb, `" fill="none" stroke="#222" stroke-width="3"/>`)
	}

	// 4) Enter 矩形蓝色边框（标注门洞的强调）
	for _, r := range enters {
		r = r.Norm()
		fmt.Fprintf(&sb,
			`<rect x="%d" y="%d" width="%d" height="%d" `+
				`fill="none" stroke="#1a6018" stroke-width="2" stroke-dasharray="4 2"/>`,
			r.X1, r.Y1, r.X2-r.X1, r.Y2-r.Y1)
	}

	// 4.5) FOV 边界蓝色虚线（前端常见）
	if opt.ShowFOV {
		drawFOVOutline(&sb, mount)
	}

	// 4.6) Sleepad 图标（蓝色小圆点 + 标签）
	for i, sp := range opt.Sleepads {
		drawSleepad(&sb, sp, fmt.Sprintf("Sleepad_%d", i+1))
	}

	// 4.7) 实时 track（绿色小人）
	for i, tp := range opt.LiveTracks {
		label := ""
		if i < len(opt.TrackLabels) {
			label = opt.TrackLabels[i]
		}
		drawLiveTrack(&sb, tp, label)
	}

	// 5) Radar（红三角 + 文字）
	drawRadar(&sb, mount)

	// 6) 标题 + 图例
	titleSuffix := opt.TitleSuffix
	drawTitleAndLegend(&sb, roomID, grid, mount, canvasW, canvasH, titleSuffix)

	sb.WriteString(`</svg>`)

	return sb.String()
}

// cellSVGFill 6 色映射 + 派生几何状态（Deny 优先）
func cellSVGFill(c *Cell) string {
	// 几何状态优先：!InRoom 一律 Deny（包括墙外 ghost 区）
	if !c.InRoom {
		return "#3a3a3a" // Deny 灰黑
	}
	// 雷达盲区（InRoom 但 !InFOV）— 学不到，保守归 Unknown
	if !c.InFOV {
		return "#c8c8c8" // Unknown 灰
	}
	// InRoom × InFOV：按 Belief[0].Type 上 6 色
	switch c.Belief[0].Type {
	case AreaEnter:
		return "#44cc66" // 绿
	case AreaBed:
		return "#4488dd" // 蓝（Lying）
	case AreaSit, AreaToilet:
		return "#ff9933" // 橙（Sit；马桶也归 Sit）
	case AreaActive:
		return "#ffffff" // 白（Walk）
	case AreaShower:
		return "#ffffff" // 白（暂归 Walk，后期可独立色）
	case AreaDeny:
		return "#3a3a3a" // 灰黑
	}
	// AreaUnknown
	return "#c8c8c8" // 灰
}

func drawAxes(sb *strings.Builder, w, h int) {
	fmt.Fprintf(sb, `<g stroke="#bbb" stroke-width="0.5" stroke-dasharray="2 3" fill="none">`)
	for x := -w / 2; x <= w/2; x += 100 {
		fmt.Fprintf(sb, `<line x1="%d" y1="0" x2="%d" y2="%d"/>`, x, x, h)
	}
	for y := 0; y <= h; y += 100 {
		fmt.Fprintf(sb, `<line x1="%d" y1="%d" x2="%d" y2="%d"/>`, -w/2, y, w/2, y)
	}
	sb.WriteString(`</g>`)

	fmt.Fprintf(sb, `<g fill="#666" font-size="10">`)
	for x := -w / 2; x <= w/2; x += 100 {
		fmt.Fprintf(sb, `<text x="%d" y="-6" text-anchor="middle">%d</text>`, x, x)
	}
	for y := 0; y <= h; y += 100 {
		fmt.Fprintf(sb, `<text x="%d" y="%d" text-anchor="end">%d</text>`, -w/2-4, y+4, y)
	}
	sb.WriteString(`</g>`)
}

func drawRadar(sb *strings.Builder, m radarutils.RadarMount) {
	cx, cy := m.Center.X, m.Center.Y
	fmt.Fprintf(sb,
		`<g transform="translate(%d %d) rotate(%d)">`+
			`<circle cx="0" cy="0" r="14" fill="rgba(255,80,80,0.16)"/>`+
			`<polygon points="0,15 -8,-3 8,-3" fill="rgba(220,40,40,0.95)" stroke="#600" stroke-width="0.5"/>`+
			`<circle cx="0" cy="0" r="3" fill="white" stroke="#a00" stroke-width="1.5"/>`+
			`</g>`,
		cx, cy, m.Rotation)
	fmt.Fprintf(sb,
		`<text x="%d" y="%d" fill="#a00" font-size="11" font-weight="bold">Radar (%s, h=%dcm)</text>`,
		cx+18, cy-8, installLabel(m.InstallModel), m.Center.Z)
}

func drawTitleAndLegend(sb *strings.Builder, roomID string, grid *RoomGrid,
	m radarutils.RadarMount, canvasW, canvasH int, titleSuffix string) {
	// 标题
	fmt.Fprintf(sb,
		`<text x="0" y="%d" text-anchor="middle" fill="#222" font-size="13" font-weight="bold">`+
			`Room: %s | %d×%d cells | radar @(%d,%d,%d) %s%s</text>`,
		canvasH+18, roomID, grid.Width, grid.Height,
		m.Center.X, m.Center.Y, m.Center.Z, installLabel(m.InstallModel), titleSuffix)

	// 6 色图例（横向排开）
	type item struct {
		fill, label, textColor string
	}
	legend := []item{
		{"#c8c8c8", "Unknown", "#222"},
		{"#ffffff", "Walk", "#222"},
		{"#ff9933", "Sit", "#222"},
		{"#4488dd", "Lying", "#fff"},
		{"#44cc66", "Enter", "#222"},
		{"#3a3a3a", "Deny", "#fff"},
	}
	totalW := len(legend) * 70
	x0 := -totalW / 2
	y0 := canvasH + 30
	for i, it := range legend {
		x := x0 + i*70
		fmt.Fprintf(sb,
			`<rect x="%d" y="%d" width="60" height="14" fill="%s" stroke="#888" stroke-width="0.5"/>`+
				`<text x="%d" y="%d" text-anchor="middle" fill="%s" font-size="10" font-weight="600">%s</text>`,
			x, y0, it.fill,
			x+30, y0+10, it.textColor, it.label)
	}
}

// drawFOVOutline 雷达 boundary 蓝色虚线框（前端常见样式）
func drawFOVOutline(sb *strings.Builder, m radarutils.RadarMount) {
	poly := radarutils.BoundaryVertices(m)
	if len(poly) < 3 {
		return
	}
	if m.InstallModel == radarutils.InstallCeiling || m.InstallModel == radarutils.InstallWall {
		poly = []radarutils.Point{poly[0], poly[1], poly[3], poly[2]}
	}
	fmt.Fprintf(sb, `<polygon points="`)
	for _, p := range poly {
		fmt.Fprintf(sb, "%d,%d ", p.X, p.Y)
	}
	fmt.Fprintf(sb,
		`" fill="none" stroke="#1c8acc" stroke-width="1.5" stroke-dasharray="6 4"/>`)
}

// drawSleepad Sleepad 图标：黄色虚线小框（2 个并排 cell，20×10cm）+ 3 个黄点 + 标签
func drawSleepad(sb *strings.Builder, p radarutils.Point, label string) {
	const yellow = "#d9a800"
	fmt.Fprintf(sb,
		`<g>`+
			`<rect x="%d" y="%d" width="20" height="10" fill="none" stroke="%s" `+
			`stroke-width="1" stroke-dasharray="2 1.5"/>`+
			`<circle cx="%d" cy="%d" r="1.5" fill="%s"/>`+
			`<circle cx="%d" cy="%d" r="1.5" fill="%s"/>`+
			`<circle cx="%d" cy="%d" r="1.5" fill="%s"/>`+
			`<text x="%d" y="%d" fill="%s" font-size="9" text-anchor="middle">%s</text>`+
			`</g>`,
		p.X-10, p.Y-5, yellow,
		p.X-5, p.Y, yellow,
		p.X, p.Y, yellow,
		p.X+5, p.Y, yellow,
		p.X, p.Y+16, yellow, label)
}

// drawLiveTrack 实时 track 小人图标（前端 P0 风格的绿色立人）
func drawLiveTrack(sb *strings.Builder, p radarutils.Point, label string) {
	// 简易立人：圆头 + 矩形身躯
	fmt.Fprintf(sb,
		`<g>`+
			`<circle cx="%d" cy="%d" r="5" fill="#2bb24c" stroke="#0a7028" stroke-width="1"/>`+
			`<rect x="%d" y="%d" width="10" height="14" fill="#2bb24c" stroke="#0a7028" stroke-width="1" rx="2"/>`+
			`<text x="%d" y="%d" fill="#0a7028" font-size="9" font-weight="bold" text-anchor="middle">%s</text>`+
			`</g>`,
		p.X, p.Y-12,
		p.X-5, p.Y-6,
		p.X, p.Y-18, label)
}

func installLabel(m radarutils.InstallModel) string {
	switch m {
	case radarutils.InstallCeiling:
		return "ceiling"
	case radarutils.InstallWall:
		return "wall"
	case radarutils.InstallCorn:
		return "corn"
	}
	return "?"
}
