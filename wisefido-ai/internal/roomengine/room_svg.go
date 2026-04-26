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
	TrackPaths  []TrackPath        // 历史轨迹叠加（per-track polyline）
	TitleSuffix string             // 标题追加（playback 用，例如 " | 2026-04-25 12:34:56"）
}

// TrackPath 一条 track 的历史路径（按时间排序的画布坐标点序列）。
// 用于在 SVG 上叠加最近 N 分钟轨迹，方便观察"history track 是否覆盖学习区域"。
type TrackPath struct {
	TrackID int                // 用于颜色稳定性（不同 track 不同色）
	Points  []radarutils.Point // 按时间顺序
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

	// <defs>：AI 学到的 Deny cell 用斜线 pattern，与人标 Deny 区分
	sb.WriteString(`<defs>` +
		`<pattern id="aiDeny" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">` +
		`<rect width="6" height="6" fill="#5a5a5a"/>` +
		`<line x1="0" y1="0" x2="0" y2="6" stroke="#9a9a9a" stroke-width="2"/>` +
		`</pattern>` +
		`</defs>`)

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

	// 4.6) 历史轨迹（半透明 polyline，每条 track 一色）
	for _, tp := range opt.TrackPaths {
		drawTrackPath(&sb, tp)
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

// cellSVGFill 按 Belief[0].Type + Source 返回填充色。
// AreaDeny 三种来源用三种深浅 / 纹理：
//   - !InRoom（几何）       → #2a2a2a 最深（边界外）
//   - SourceHuman Deny      → #4a4a4a 实色中灰（layout 标的家具）
//   - SourceLearned Deny    → url(#aiDeny) 斜线 pattern（AI 推断的家具/障碍）
func cellSVGFill(c *Cell) string {
	// 几何状态优先：!InRoom 一律最深 Deny
	if !c.InRoom {
		return "#2a2a2a"
	}
	// 雷达盲区（InRoom 但 !InFOV）— 学不到，保守归 Unknown
	if !c.InFOV {
		return "#c8c8c8"
	}
	// InRoom × InFOV：按 Belief[0].Type 上色
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
		// Source 区分：AI 学的用斜线，人标用实色
		if c.Belief[0].Source == SourceLearned {
			return "url(#aiDeny)"
		}
		return "#4a4a4a" // 人标 / Geometry / 默认
	}
	// AreaUnknown
	return "#c8c8c8"
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

	// 图例（横向排开；Deny 三色分开：墙外 / 人标家具 / AI 学习）
	type item struct {
		fill, label, textColor string
	}
	legend := []item{
		{"#c8c8c8", "Unknown", "#222"},
		{"#ffffff", "Walk", "#222"},
		{"#ff9933", "Sit", "#222"},
		{"#4488dd", "Lying", "#fff"},
		{"#44cc66", "Enter", "#222"},
		{"#2a2a2a", "Out", "#fff"},        // !InRoom
		{"#4a4a4a", "Deny(layout)", "#fff"}, // SourceHuman
		{"url(#aiDeny)", "Deny(AI)", "#fff"}, // SourceLearned 斜线
	}
	const itemW = 78
	totalW := len(legend) * itemW
	x0 := -totalW / 2
	y0 := canvasH + 30
	for i, it := range legend {
		x := x0 + i*itemW
		fmt.Fprintf(sb,
			`<rect x="%d" y="%d" width="68" height="14" fill="%s" stroke="#888" stroke-width="0.5"/>`+
				`<text x="%d" y="%d" text-anchor="middle" fill="%s" font-size="10" font-weight="600">%s</text>`,
			x, y0, it.fill,
			x+34, y0+10, it.textColor, it.label)
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
// trackPathColors 8 种半透明色循环（按 trackID 取模）
var trackPathColors = []string{
	"rgba(255, 70, 70, 0.55)",  // 红
	"rgba(70, 130, 255, 0.55)", // 蓝
	"rgba(170, 70, 200, 0.55)", // 紫
	"rgba(70, 180, 170, 0.55)", // 青
	"rgba(255, 140, 50, 0.55)", // 橙
	"rgba(80, 180, 80, 0.55)",  // 绿
	"rgba(200, 140, 60, 0.55)", // 棕
	"rgba(120, 120, 120, 0.55)",// 灰
}

func drawTrackPath(sb *strings.Builder, tp TrackPath) {
	if len(tp.Points) < 2 {
		return
	}
	color := trackPathColors[tp.TrackID%len(trackPathColors)]
	// SVG polyline points="x1,y1 x2,y2 ..."
	sb.WriteString(`<polyline fill="none" stroke="`)
	sb.WriteString(color)
	sb.WriteString(`" stroke-width="1.6" stroke-linejoin="round" stroke-linecap="round" points="`)
	for i, p := range tp.Points {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(sb, "%d,%d", p.X, p.Y)
	}
	sb.WriteString(`"/>`)
	// 起点小圆点（淡）+ 终点空心圆（明显）
	start := tp.Points[0]
	end := tp.Points[len(tp.Points)-1]
	fmt.Fprintf(sb,
		`<circle cx="%d" cy="%d" r="1.6" fill="%s"/>`+
			`<circle cx="%d" cy="%d" r="2.6" fill="none" stroke="%s" stroke-width="1.4"/>`,
		start.X, start.Y, color,
		end.X, end.Y, color)
}

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
