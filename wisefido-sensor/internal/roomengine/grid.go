package roomengine

import (
	"owl-common/radarutils"
)

// RoomGrid 房间的 2D 网格（"底片"），每个格子 CellSize×CellSize cm。
//
// 坐标系：画布坐标（顶部中心为原点，X 左负右正，Y 下正）。
// Grid[0][0] 左上角对应的画布坐标为 (OriginX, OriginY)。
// 默认 OriginX = -RoomW/2, OriginY = 0（与前端 layout_config 对齐）。
//
// 构建流程（在 engine.RegisterRoom 里依次调用）：
//  1. NewRoomGrid(roomW, roomH, cellSize)        // 空网格
//  2. grid.StampRoomPolygon(wallPoly)            // 刷 InRoom
//  3. grid.StampRadar(mount)                      // 刷 InFOV / EdgeDist / MaxZ / MinZ
//  4. grid.StampEnters(enterRects)                // 记 Enters 列表 + 覆写 InRoom
//  5. grid.SetPrior(rect, AreaType, conf, src)    // 人标先验注入
//
// 运行时：
//
//	CellAt / NearestEntryDist / IsEdge           // 查询
//	MarkOccupancy / MarkPoseTime / ...           // 每帧累积（历史流）
//	DecayAll                                     // 定时衰减
type RoomGrid struct {
	CellSize int
	Width    int
	Height   int
	RoomW    int
	RoomH    int
	OriginX  int
	OriginY  int

	Cells  []Cell
	Enters []radarutils.Rect

	// AreaZones layout 声明区（源自 DB canvas 的解析缓存）。RegisterRoom 时随 grid 一次性建、发布后只读；
	// layout 变更走 reload 重建整根 grid（与 Cells/Enters 同生命周期），故无需锁。人标区型只活在此、
	// 不烙 belief、不进 28 snapshot，QueryAreaType 用时查——UpdateBelief 侵蚀碰不到它。
	AreaZones []AreaZone
}

// NewRoomGrid 创建空网格。默认画布坐标系（顶部中心原点）。
func NewRoomGrid(roomW, roomH, cellSize int) *RoomGrid {
	if cellSize <= 0 {
		cellSize = radarutils.CellSize
	}
	w := roomW/cellSize + 1
	h := roomH/cellSize + 1
	return &RoomGrid{
		CellSize: cellSize,
		Width:    w,
		Height:   h,
		RoomW:    roomW,
		RoomH:    roomH,
		OriginX:  -roomW / 2,
		OriginY:  0,
		Cells:    make([]Cell, w*h),
	}
}

// ToIndex 画布坐标 → 格子索引
func (g *RoomGrid) ToIndex(x, y int) (col, row int) {
	col = (x - g.OriginX) / g.CellSize
	row = (y - g.OriginY) / g.CellSize
	return
}

// CellAt 画布坐标 (x, y) → 格子指针。越界返回 nil。
func (g *RoomGrid) CellAt(x, y int) *Cell {
	col, row := g.ToIndex(x, y)
	if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
		return nil
	}
	return &g.Cells[row*g.Width+col]
}

// QueryAreaType 坐标此刻的有效区型（B 组决策唯一读点）：优先 layout 声明区（AreaZones，覆盖点内
// 按 areaPriority 取最高，等价旧 SetPrior 重叠合并），无声明区落 learned Belief[0]。
// 人标只在 AreaZones、learned 只在 Belief——两源分离，learn 侵蚀不影响声明区。
func (g *RoomGrid) QueryAreaType(x, y int) AreaType {
	best := AreaUnknown
	found := false
	for _, z := range g.AreaZones {
		if !z.Rect.Contains(x, y) {
			continue
		}
		if !found || areaPriority(z.AreaType) > areaPriority(best) {
			best = z.AreaType
			found = true
		}
	}
	if found {
		return best
	}
	if c := g.CellAt(x, y); c != nil {
		return c.Belief[0].Type
	}
	return AreaUnknown
}

// ToCanvas 格子索引 → 格子中心画布坐标
func (g *RoomGrid) ToCanvas(col, row int) (x, y int) {
	x = g.OriginX + col*g.CellSize + g.CellSize/2
	y = g.OriginY + row*g.CellSize + g.CellSize/2
	return
}

// IsEdge 画布坐标是否在房间边界附近
func (g *RoomGrid) IsEdge(x, y, margin int) bool {
	return x-g.OriginX < margin ||
		y-g.OriginY < margin ||
		g.OriginX+g.RoomW-x < margin ||
		g.OriginY+g.RoomH-y < margin
}

// ========================================================================
// Rasterize 接口（构建时一次性调用）
// ========================================================================

// StampRoomPolygon 把 Wall 多边形刷到 cells 的 InRoom 字段。
func (g *RoomGrid) StampRoomPolygon(poly []radarutils.Point) {
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cx, cy := g.ToCanvas(col, row)
			if radarutils.PolygonContains(poly, cx, cy) {
				g.Cells[row*g.Width+col].InRoom = true
			}
		}
	}
}

// StampRadar 把雷达物理 FOV 刷到 cells 的 InFOV / EdgeDist / MaxZ / MinZ。
func (g *RoomGrid) StampRadar(mount radarutils.RadarMount) {
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cx, cy := g.ToCanvas(col, row)
			c := &g.Cells[row*g.Width+col]
			c.InFOV = mount.SignalReachable(cx, cy, 0)
			if c.InFOV {
				c.EdgeDist = mount.SignalEdgeDist(cx, cy, 0)
			}
			c.MaxZ = mount.MaxZAt(cx, cy)
			c.MinZ = mount.MinZAt(cx, cy)
		}
	}
}

// StampEnters 记录 Enter 矩形集合，覆写矩形内 cell 的 InRoom=true（门洞可穿越）。
// sensor_v2 决定 15: 同时把 enterTargets[i] 对应的 EnterTarget 字符串盖到 cell 上。
// enterTargets 长度可不等于 rects（缺失或更短时按 "" 填充——视作 inside_enter 默认）。
func (g *RoomGrid) StampEnters(rects []radarutils.Rect, enterTargets []string) {
	g.Enters = rects
	for i, rect := range rects {
		target := ""
		if i < len(enterTargets) {
			target = enterTargets[i]
		}
		rect = rect.Norm()
		c1, r1 := g.ToIndex(rect.X1, rect.Y1)
		c2, r2 := g.ToIndex(rect.X2, rect.Y2)
		for row := r1; row <= r2; row++ {
			if row < 0 || row >= g.Height {
				continue
			}
			for col := c1; col <= c2; col++ {
				if col < 0 || col >= g.Width {
					continue
				}
				c := &g.Cells[row*g.Width+col]
				c.InRoom = true
				c.EnterTarget = target
			}
		}
	}
}

// areaPriority 区域重叠危险等级（高→低）：豁免类(Reflector/Interfer) > Bed > Lying > Sit > Deny > Active > Enter > Unknown。
// masking/无真人区(reflector/interfer)压过床/躺/坐，避免反射杂波被当真人/床。多对象叠一 cell 时的主排序键。
func areaPriority(t AreaType) int {
	switch t {
	case AreaReflector, AreaInterfer:
		return 7
	case AreaBed, AreaMonitorBed:
		return 6
	case AreaLying:
		return 5
	case AreaSit:
		return 4
	case AreaDeny:
		return 3
	case AreaActive:
		return 2
	case AreaEnter:
		return 1
	default: // AreaUnknown
		return 0
	}
}

// VetoRect 对 rect 内每个 cell 擦掉非 Human 抑制/deny（ClearNonHumanLearnedZone）；sticky 时额外
// MarkLearnBlocked 永久封自动学习。返回被清/被封 cell 数。handle "Clear/Never auto-suppress" 按 fire
// 点足迹（40×40）整片否决，而非单 10cm cell——火点 cm 级抖动 + 真实危险点足迹 > 单格。
func (g *RoomGrid) VetoRect(rect radarutils.Rect, sticky bool) (cleared, blocked int) {
	rect = rect.Norm()
	c1, r1 := g.ToIndex(rect.X1, rect.Y1)
	c2, r2 := g.ToIndex(rect.X2, rect.Y2)
	for row := r1; row <= r2; row++ {
		if row < 0 || row >= g.Height {
			continue
		}
		for col := c1; col <= c2; col++ {
			if col < 0 || col >= g.Width {
				continue
			}
			c := &g.Cells[row*g.Width+col]
			if c.ClearNonHumanLearnedZone() {
				cleared++
			}
			if sticky && c.MarkLearnBlocked() {
				blocked++
			}
		}
	}
	return cleared, blocked
}

// NearestEntryDist 到最近 Enter 矩形的距离（cm）
func (g *RoomGrid) NearestEntryDist(x, y int) int {
	if len(g.Enters) == 0 {
		return 9999
	}
	best := 9999
	for _, r := range g.Enters {
		d := r.DistTo(x, y)
		if d < best {
			best = d
		}
	}
	return best
}

// DecayAll 对所有 cell 做时间衰减（按字段语义分档，p 一次性传入）
func (g *RoomGrid) DecayAll(dtSec float64, p DecayParams) {
	for i := range g.Cells {
		g.Cells[i].Decay(dtSec, p)
	}
}

// ========================================================================
// Mark 接口（每帧 track 层调用，历史流累积）
// ========================================================================

// qOccupancyThreshold 本帧质量分阈值（Kalman 残差等打分后）
const qOccupancyThreshold = 50

// MarkOccupancy 按本帧质量 q 分桶：q 高 → RealDecay + FlowEMA，q 低 → GhostDecay。
// 历史流不过滤 Verdict——每帧都累积，时间/衰减自淘汰。
func (g *RoomGrid) MarkOccupancy(x, y, quality, vx, vy int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	if quality >= qOccupancyThreshold {
		c.RealDecay++
		// FlowEMA（α=0.1 int 定点）
		c.FlowX = (9*c.FlowX + vx) / 10
		c.FlowY = (9*c.FlowY + vy) / 10
	} else {
		c.GhostDecay++
	}
	c.LastUpdateMs = nowMs
}

// MarkPoseTime 按本帧核心姿态累加 ActiveType（×10 定点 = 0.1 秒/单位）。
// 用 3×3 内核：中心 cell 加 dtSec×10，8 邻居各加 dtSec×8（80% 权重）。
// 设计原因：补偿雷达位置 jitter——人坐稳时中心持续命中，散点 hovering 时分散到 9 cell 都过不了升格阈值。
// 每帧都调（无过滤）。dtSec 为本帧时长（秒，通常 1）。
func (g *RoomGrid) MarkPoseTime(x, y int, core CorePose, dtSec int, nowMs int64) {
	if dtSec <= 0 {
		return
	}
	idx := CorePoseToActiveIdx(core)
	if idx < 0 {
		return
	}
	col, row := g.ToIndex(x, y)
	if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
		return
	}

	const centerW = 10 // ×10 定点：本帧中心格子 = dtSec × 1.0
	const neighW = 8   // 邻居 = dtSec × 0.8

	for dr := -1; dr <= 1; dr++ {
		rr := row + dr
		if rr < 0 || rr >= g.Height {
			continue
		}
		for dc := -1; dc <= 1; dc++ {
			cc := col + dc
			if cc < 0 || cc >= g.Width {
				continue
			}
			c := &g.Cells[rr*g.Width+cc]
			w := neighW
			if dr == 0 && dc == 0 {
				w = centerW
			}
			c.ActiveType[idx] = satAddUint16(c.ActiveType[idx], dtSec*w)
			c.LastUpdateMs = nowMs
		}
	}
}

// MarkTraverse Move 状态下"穿越本 cell"的事件（track_manager 在进入新 cell 且 core==Move 时调）。
//
// 双重作用：
//  1. 本 cell ++TraverseCount → Walk 升格证据（人真走过这里）
//  2. 8 邻居 ++NearTraverseCount → 邻接证据（auto-Deny 推断用：邻居反复被走但本 cell 0 触碰 = 人系统性绕开 = 障碍）
//
// 邻居传播不区分正交/对角（实际行人走位也不沿严格栅格）。后续如需精细加权，可改 8 邻居各自系数。
func (g *RoomGrid) MarkTraverse(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.TraverseCount = satAddUint16(c.TraverseCount, 1)
	c.LastUpdateMs = nowMs

	// 邻居传播
	col, row := g.ToIndex(x, y)
	for dr := -1; dr <= 1; dr++ {
		rr := row + dr
		if rr < 0 || rr >= g.Height {
			continue
		}
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue // 中心 cell 已经 ++ 过了
			}
			cc := col + dc
			if cc < 0 || cc >= g.Width {
				continue
			}
			n := &g.Cells[rr*g.Width+cc]
			n.NearTraverseCount = satAddUint16(n.NearTraverseCount, 1)
		}
	}
}

// IsNearPriorType 检查 (x,y) 是否在某条 AreaType==t 的声明区（AreaZone）容差 marginCm 内。
// 用于 cell_learning：判断"床外 Lie"时给 Bed 声明区 ±30cm 容差，避免 layout 位置误差误报。
// 直接查声明区（人标唯一源），不再读 belief——人标不烙 belief。
func (g *RoomGrid) IsNearPriorType(x, y int, t AreaType, marginCm int) bool {
	for _, z := range g.AreaZones {
		if z.AreaType == t && z.Rect.DistTo(x, y) <= marginCm {
			return true
		}
	}
	return false
}

// MarkFallEvent 自推"跌倒事件"触发时调（Stand/Move → Lie + Z 骤降）
func (g *RoomGrid) MarkFallEvent(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.FallEventCount++
	c.LastUpdateMs = nowMs
}

// MarkLieRetract 短暂 Lie 回撤（Stand/Move → Lie → Stand/Move < 3 秒）
func (g *RoomGrid) MarkLieRetract(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.LieRetract++
	c.LastUpdateMs = nowMs
}

// MarkDwell 静止期结束时调，把本次 dwell 秒灌入 DwellEMA
func (g *RoomGrid) MarkDwell(x, y, dwellSec int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.DwellEMA = (4*c.DwellEMA + dwellSec) / 5
	c.LastUpdateMs = nowMs
}

// MarkDwellRegion 把本次 dwell 摊到 still 区 [minX,maxX]×[minY,maxY] 覆盖的每个 10×10 cell。
// still-box per-axis ≤50cm 跨多格，单格 credit 会让"学的格 ≠ DBN 查的 live 格"→ 学了压不住。
func (g *RoomGrid) MarkDwellRegion(minX, minY, maxX, maxY, dwellSec int, nowMs int64) {
	c0, r0 := g.ToIndex(minX, minY)
	c1, r1 := g.ToIndex(maxX, maxY)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
				continue
			}
			c := &g.Cells[row*g.Width+col]
			c.DwellEMA = (4*c.DwellEMA + dwellSec) / 5
			c.LastUpdateMs = nowMs
		}
	}
}

// MarkToleratedStillRegion 同 MarkDwellRegion，按覆盖格累计 tolerated-still。
func (g *RoomGrid) MarkToleratedStillRegion(minX, minY, maxX, maxY int, nowMs int64) {
	c0, r0 := g.ToIndex(minX, minY)
	c1, r1 := g.ToIndex(maxX, maxY)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if col < 0 || col >= g.Width || row < 0 || row >= g.Height {
				continue
			}
			c := &g.Cells[row*g.Width+col]
			c.IncrToleratedStill()
			c.LastUpdateMs = nowMs
		}
	}
}

// MarkLongStill 静止超 15min 阈值事件（一次 ++）
func (g *RoomGrid) MarkLongStill(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.LongStillCount++
	c.LastUpdateMs = nowMs
}

// MirrorPromoteThreshold L1 mirror cell 晋升 AreaDeny+SourceLearned 的 MBC 阈值。
// "≥3 次独立配对命中"——单 cell 涂法下每次配对单 bounce 点 +1，需 3 次独立运动累积。
const MirrorPromoteThreshold = 3

// MarkMirrorBounce L1 mirror pair 反射点累加：单 cell（10×10cm）+=1。
// 雷达 XY 精度本来就 ±5-10cm 与 cell 同量级，不再 2×2 涂抹（避免单次配对 5 bounce
// 点共 20 次 cell++ 假性达阈值）。命中达 MirrorPromoteThreshold 即 promote
// Belief[0]=AreaDeny+SourceLearned（已是 SourceHuman 的不动）。
func (g *RoomGrid) MarkMirrorBounce(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.MirrorBounceCount++
	c.LastMirrorMs = nowMs
	c.LastUpdateMs = nowMs
	if c.MirrorBounceCount >= MirrorPromoteThreshold && !c.LearnBlocked &&
		c.Belief[0].Source != SourceFeedback {
		c.Belief[0] = BeliefState{Type: AreaDeny, Confidence: 70, Source: SourceLearned}
		c.AreaType = AreaDeny
	}
}

// StaticReflectorPromoteThreshold 静止反射体 cell 晋升 AreaReflector+SourceLearned 的独立 episode 阈值。
const StaticReflectorPromoteThreshold = 3

// MarkStaticReflector 静止金属反射体累加 + 晋升（static_reflector.go 检测命中调）。
// 达 StaticReflectorPromoteThreshold 即 promote Belief[0]=AreaReflector+SourceLearned（已是 SourceHuman/
// SourceFeedback 的不动）。升 AreaReflector 而非 AreaDeny：金属幻影常**孤迹**（无人在房，roamer=-1）→
// realness co-existence 抓不到（需 ≥2 轨），唯一能压住 floor FP 的是 areaReflector 的 floor 整片豁免
// （floor.go：AreaDeny 走 default tFloor=12min 等于不压）。demote 由 cell_learning hasRealActivity 兜底
// （真人真走过 → 退回 Unknown，自纠误学）。
func (g *RoomGrid) MarkStaticReflector(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.StaticReflectorCount++
	c.LastStaticReflectorMs = nowMs
	if c.StaticReflectorCount >= StaticReflectorPromoteThreshold && !c.LearnBlocked &&
		c.Belief[0].Source != SourceFeedback {
		c.Belief[0] = BeliefState{Type: AreaReflector, Confidence: 70, Source: SourceLearned}
		c.AreaType = AreaReflector
	}
	c.LastUpdateMs = nowMs
}

// MarkSleepadInBed Sleepad 压床事件（HR/RR/InBed）
func (g *RoomGrid) MarkSleepadInBed(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.SleepadInBedCount++
	c.LastUpdateMs = nowMs
}

// MarkSleepadLeftBed Sleepad 离床事件
func (g *RoomGrid) MarkSleepadLeftBed(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.SleepadLeftBedCount++
	c.LastUpdateMs = nowMs
}

// MarkDoorEvent Enter/Exit 门事件（决定 20：v2 复活此 caller — runEventLoop EnterRoom/ExitRoom 触发）
// 注：与 MarkInsideEnterCandidate 是两个独立信号源：
//   - MarkDoorEvent：firmware Enter/Exit 事件落到 layout 标注的物理门 cell（priorTypeScore 用）
//   - MarkInsideEnterCandidate：track 失锁/重生 推断 L 型 room 内部盲区入口（决定 20 inside_enter 自学习）
func (g *RoomGrid) MarkDoorEvent(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.DoorEventCount++
	c.LastUpdateMs = nowMs
}

// MarkInsideEnterCandidate inside_enter 自学习信号（sensor_v2 决定 20）。
//
// 调用位点：runEventLoop / processFrameAt 中检测到
//
//	"track 在 cell_x 失锁 + Δt < 3s 内 cell_x 附近 ≤30cm 有新 track 出生"
//
// 此时调用 g.MarkInsideEnterCandidate(cell_x.X, cell_x.Y, nowMs)
//
// 累计 ≥ InsideEnterLearnThreshold（默认 5）次后升格 InsideEnterLearned=true，
// 写 ai.log 给运维审核（不自动写 layout，决定 15）。
//
// 物理含义：L 型 room 雷达盲区 / 雷达 FOV 边界但 room 内部，人能自由进出但雷达看不到。
// 学到后：该 cell 上方的 track 出生/消失不再触发 ghost penalty / lost_fall pending。
//
// 与 outside_enter / bathroom_enter 学习的区别（决定 20）：
//
//	v2 单 device-layout 坐标系内**只能学 inside_enter**；跨 device 没有公共坐标系，
//	outside / bathroom 必须人工标（layout 编辑器 EnterTarget dropdown）。
func (g *RoomGrid) MarkInsideEnterCandidate(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.InsideEnterEvidenceN++
	if c.InsideEnterEvidenceN >= InsideEnterLearnThreshold && !c.InsideEnterLearned {
		c.InsideEnterLearned = true
		// 注：实际升格 cell.AreaType=AreaEnter 不在此处做，由 cell_learning 周期 promoteCell 消费 InsideEnterLearned flag。
		// 这里仅写状态 + 计数。
	}
	c.LastUpdateMs = nowMs
}

// MarkFakeAlarm 人工标记 false_alarm 反馈到该 cell（cell history integral）。
// 来源：alarm_events.handle_type='false_alarm' 触发，定位回该 cell 后调用。
func (g *RoomGrid) MarkFakeAlarm(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrFakeAlarm()
	c.LastUpdateMs = nowMs
}

// MarkToleratedStill 容忍静止反馈：track 在该 cell 长时间 stand-static 后自然离开（无真跌倒）。
// 调用位点：track 从静止恢复移动 + 之前已 LongStillReported（系统曾认为可疑）。
func (g *RoomGrid) MarkToleratedStill(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrToleratedStill()
	c.LastUpdateMs = nowMs
}

// MarkBlindSpotRecovery 盲区返回反馈：lost-fall pending 期内人在此 cell 重新出现。
// 该 cell 是已知盲区出口，未来 birth-score 可加成。
func (g *RoomGrid) MarkBlindSpotRecovery(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrBlindSpotRecovery()
	c.LastUpdateMs = nowMs
}

// MarkGhostFeedback PR-6：人工标 ☑ NoPerson/Electric/AC Interference → cell.GhostCount++
func (g *RoomGrid) MarkGhostFeedback(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrGhostCount()
	c.LastUpdateMs = nowMs
}

// MarkRestZoneFeedback PR-6+7.1+9.2+10：人工标 ☑ Sit on Chair/Sofa/Wheelchair。
// target 区分语义：
//   - AreaSit （Chair / Wheelchair）— 坐姿区，橙色
//   - AreaBed （Sofa / Lounge chair）— 躺姿区（沙发可坐可躺，归 lying），蓝色
//
// 流程：
//  1. 累计 RestZoneConfirmed
//  2. 即时锁定 target Belief
//  3. PR-10：4 邻居中**已是同类**的 cell confidence ×1.3（cap 100）增强；
//     设计意图：sofa/床占多 cell，相邻同类反馈 = 强空间一致性证据，
//     抵抗 BeliefSec=2d 半衰期的衰退。
func (g *RoomGrid) MarkRestZoneFeedback(x, y int, target AreaType, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrRestZoneConfirmed()
	c.MarkRestZoneByFeedback(target)
	c.LastUpdateMs = nowMs
	// PR-10: 邻居增强
	g.boostNeighborSameType(x, y, target, nowMs)
}

// boostNeighborSameType PR-10：4 邻居（上下左右）同 target 类型 cell confidence ×1.3。
// cap 100；不动新 cell 自己；不动不同类邻居。设计平衡：
//   - 一次邻居增强 ×1.3 (+30%) ≈ 1 天衰退（-29%，BeliefSec=2d 算）的反向
//   - 多次确认下能稳态维持；单次反馈不会立即推到 100 锁死
func (g *RoomGrid) boostNeighborSameType(x, y int, target AreaType, nowMs int64) {
	col, row := g.ToIndex(x, y)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nc, nr := col+d[0], row+d[1]
		if nr < 0 || nr >= g.Height || nc < 0 || nc >= g.Width {
			continue
		}
		n := &g.Cells[nr*g.Width+nc]
		if n.Belief[0].Type != target {
			continue
		}
		newConf := n.Belief[0].Confidence * 13 / 10 // ×1.3
		if newConf > 100 {
			newConf = 100
		}
		if newConf > n.Belief[0].Confidence {
			n.Belief[0].Confidence = newConf
			n.LastUpdateMs = nowMs
		}
	}
}

// MarkRealFallFeedback PR-6：人工标 verified ☑ Fall (ground truth) → cell.RealFallCount++
func (g *RoomGrid) MarkRealFallFeedback(x, y int, nowMs int64) {
	c := g.CellAt(x, y)
	if c == nil {
		return
	}
	c.IncrRealFallCount()
	c.LastUpdateMs = nowMs
}

// CellStat dump 用：把 cell 关键字段拍平。
type CellStat struct {
	Col               int      `json:"col"`
	Row               int      `json:"row"`
	X                 int      `json:"x"`
	Y                 int      `json:"y"`
	AreaType          AreaType `json:"area_type"`
	Source            Source   `json:"source"`
	Confidence        int      `json:"confidence"`
	MoveSec           float64  `json:"move_sec"`  // ActiveType[Move]/10
	StandSec          float64  `json:"stand_sec"` // ActiveType[Stand]/10
	SitSec            float64  `json:"sit_sec"`   // ActiveType[Sit]/10
	LieSec            float64  `json:"lie_sec"`   // ActiveType[Lie]/10
	TraverseCount     uint16   `json:"traverse_count"`
	NearTraverseCount uint16   `json:"near_traverse_count"`
	RealDecay         int      `json:"real_decay"`
	GhostDecay        int      `json:"ghost_decay"`
	FallEventCount    int      `json:"fall_event_count"`
	LieAnomalyCount   int      `json:"lie_anomaly_count"`
	LongStillCount    int      `json:"long_still_count"`
	LieRetract        int      `json:"lie_retract"`
}

// DumpRect 把矩形（画布坐标 x1..x2, y1..y2）内所有 InRoom cell 平铺为 CellStat 列表。
// 用于 debug：定位"为什么这片 cell 学成 X 类型"。
func (g *RoomGrid) DumpRect(x1, y1, x2, y2 int) []CellStat {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	c1, r1 := g.ToIndex(x1, y1)
	c2, r2 := g.ToIndex(x2, y2)
	if c1 < 0 {
		c1 = 0
	}
	if r1 < 0 {
		r1 = 0
	}
	if c2 >= g.Width {
		c2 = g.Width - 1
	}
	if r2 >= g.Height {
		r2 = g.Height - 1
	}
	out := make([]CellStat, 0, (c2-c1+1)*(r2-r1+1))
	for row := r1; row <= r2; row++ {
		for col := c1; col <= c2; col++ {
			c := &g.Cells[row*g.Width+col]
			if !c.InRoom {
				continue
			}
			x, y := g.ToCanvas(col, row)
			out = append(out, CellStat{
				Col: col, Row: row, X: x, Y: y,
				AreaType:          c.Belief[0].Type,
				Source:            c.Belief[0].Source,
				Confidence:        c.Belief[0].Confidence,
				MoveSec:           float64(c.ActiveType[ActiveIdxMove]) / 10,
				StandSec:          float64(c.ActiveType[ActiveIdxStand]) / 10,
				SitSec:            float64(c.ActiveType[ActiveIdxSit]) / 10,
				LieSec:            float64(c.ActiveType[ActiveIdxLie]) / 10,
				TraverseCount:     c.TraverseCount,
				NearTraverseCount: c.NearTraverseCount,
				RealDecay:         c.RealDecay,
				GhostDecay:        c.GhostDecay,
				FallEventCount:    c.FallEventCount,
				LieAnomalyCount:   c.LieAnomalyCount,
				LongStillCount:    c.LongStillCount,
				LieRetract:        c.LieRetract,
			})
		}
	}
	return out
}
