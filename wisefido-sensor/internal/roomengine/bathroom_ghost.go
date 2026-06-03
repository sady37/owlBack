// bathroom_ghost.go — sensor_v2 §4.A.3 BathroomGhostAdjudicator 真实实现（PR-6）
//
// 4 条规则（决定 14 共生律 + 决定 12 单入口约束的物理表达）：
//
//   规则 1：出生地判据
//     新出生 track 出生 cell 距 cell.EnterTarget=="bathroom" 入口点 > 30cm → Ghost
//     物理依据：决定 12 保证入口是唯一物理通道；从入口外位置出生 = 必非真人
//
//   规则 2：占用一致性
//     bathroom 内活 track 数 > census.BathroomCount（来自 PR-5 入口流量）→ 多余 track 是 Ghost
//     选择策略：远离 entry 的优先 ghost（"从入口轨迹连续"启发的简化版）
//
//   规则 3：共生律自校正（决定 14）
//     track count > 0 → BathroomCount 期望 ≥ ceil(track_count / max_ghost_per_person=3)
//     若 BathroomCount < 期望 → 调 AdjustBathroomCount 修正；不翻 ghost verdict（已由 §4.A.5 disambiguate）
//
//   规则 5：Split-ghost 出生邻近性（决定 14 瞬时表现）
//     新出生 track 与已存在非 ghost track 距离 < 60cm → Ghost
//     物理胎记：split ghost 是雷达同时收直射+反射回波，出生瞬间紧贴真人
//
// 规则 4 fallback：入口在 radar 盲区
//   bathroom 内所有 EnterTarget=="bathroom" cells 都不在 InFOV 内
//   → 规则 1（用 entry 距离） + 规则 2/3（用 BathroomCount）都不可信，仅启用规则 5
//   → fire warning log + 部署 escalate 提示
//
// 实现策略：
//   - per-room 维护 prevTracks set 识别新出生（与 PR-5 BathroomGate.members 同构，独立维护避免耦合）
//   - 状态字段单 goroutine RW（Adjudicate 由 publishTrackStatuses 单 goroutine 调用）
//   - VerdictDelta 按 trackID 聚合：多规则命中同 track → 一条 delta，reasons 拼接
//   - 与 Anchored 守卫交互：Anchored 不被翻 Ghost（PR-4 applyVerdictDeltas 强制），
//     但 PenaltyDelta 仍累加，PR-X 可借此持续观察 anchored resident 是否"持续疑似"

package roomengine

import (
	"math"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// 决定 14 衍生参数：1 个真人最多产生几个 ghost（镜面/金属/玻璃）。
// 默认 3 = 镜面 + 金属 + 玻璃各一；实测若多于 3 个建议复审 §4.A.5 镜像学习参数。
const bathroomMaxGhostsPerPerson = 3

// 规则 1 出生地距离阈值（cm）：newborn track 距最近 entry > 此值 → 判 ghost。
// 70cm（2026-06-03 由 30 放大）：放宽"贴门合法出生"区，避免把"门口几步内才被雷达捕获的真人"
// 误判 ghost（漏报真跌倒，宁可误报不可漏报）；镜面/金属 ghost 通常远在 >100cm，不受影响。
const bathroomBirthEntryThresholdCm = 70

// 规则 5 split-ghost 邻近距离阈值（cm）。默认 60cm；可调 50-80cm 范围（§4.A.3 注）。
const bathroomSplitGhostDistCm = 60

// bathroomEntryPoint bathroom 入口 cell 的画布坐标 + FOV 标记（规则 4 fallback 判定输入）。
type bathroomEntryPoint struct {
	cx, cy int
	inFOV  bool
}

// BathroomGhostAdjudicator 实现 §4.A.3 的 4 条规则 + 规则 4 fallback。
// 仅在 room.kind=="bathroom" 的 publishTrackStatuses 调用路径上生效（Engine.pickAdjudicator 分流）。
//
// 状态：
//   prevTracks[roomID] = 上一帧 trackID 集合，用来识别"新出生"（cur - prev）
//
// 不持有 v1 累积器 / 不写 TrackState / 不发 redis — 严格遵守 PR-4 4 条铁律。
type BathroomGhostAdjudicator struct {
	census      *SuiteCensusManager
	gridLookup  func(roomID string) *RoomGrid
	suiteLookup func(roomID string) string
	logger      *zap.Logger

	prevTracks map[string]map[int]struct{}

	// layoutWarned roomID → 已 warn 过 layout error（多片不连续 entry）。
	// 用途：决定 15 约定 bathroom layout 上 entry "恰好 1 个连续 cells 区域"；
	// 实际遇到多片说明 layout 编辑器校验绕过 — warn 一次留 audit trace，不阻塞判定。
	layoutWarned map[string]bool
}

// NewBathroomGhostAdjudicator 构造。
// census / gridLookup / suiteLookup 任一为 nil 时 Adjudicate 退化 noop（warn log 一次）。
// logger nil 时降级 zap.NewNop（单测构造）。
func NewBathroomGhostAdjudicator(
	census *SuiteCensusManager,
	gridLookup func(roomID string) *RoomGrid,
	suiteLookup func(roomID string) string,
	logger *zap.Logger,
) *BathroomGhostAdjudicator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BathroomGhostAdjudicator{
		census:       census,
		gridLookup:   gridLookup,
		suiteLookup:  suiteLookup,
		logger:       logger,
		prevTracks:   make(map[string]map[int]struct{}),
		layoutWarned: make(map[string]bool),
	}
}

// Adjudicate 见文件头 4 条规则注释。bases 为空时返回 nil（无 track 可判）。
func (a *BathroomGhostAdjudicator) Adjudicate(bases []TrackStatusBase, nowMs int64) []VerdictDelta {
	if len(bases) == 0 {
		return nil
	}
	roomID := bases[0].RoomID
	var grid *RoomGrid
	if a.gridLookup != nil {
		grid = a.gridLookup(roomID)
	}
	var suiteID string
	if a.suiteLookup != nil {
		suiteID = a.suiteLookup(roomID)
	}
	if grid == nil {
		// 无 grid 无法判定（bathroom 必须 RegisterRoom 走过 layout）；
		// 不返 delta，仅 warn — 后续帧若 grid 出现再恢复判定
		a.logger.Warn("bathroom_ghost_no_grid", zap.String("room_id", roomID))
		return nil
	}

	// Step 1：识别新出生（cur - prev）
	cur := make(map[int]struct{}, len(bases))
	for i := range bases {
		cur[bases[i].TrackID] = struct{}{}
	}
	prev := a.prevTracks[roomID]
	newBorn := make(map[int]bool, len(bases))
	for tid := range cur {
		if _, was := prev[tid]; !was {
			newBorn[tid] = true
		}
	}
	a.prevTracks[roomID] = cur

	// Step 2：扫 bathroom entry cells + 判 Rule 4 fallback + layout error audit
	entryPoints := findBathroomEntryPoints(grid)
	entryInFOV := false
	for _, ep := range entryPoints {
		if ep.inFOV {
			entryInFOV = true
			break
		}
	}
	if !entryInFOV && len(entryPoints) > 0 {
		a.logger.Warn("bathroom_entry_in_blind_spot",
			zap.String("room_id", roomID),
			zap.Int("entry_cells", len(entryPoints)),
		)
	}
	// 决定 15 约定 entry 是单一连续区域；多片 = layout 编辑器校验绕过，warn 一次留 audit。
	// 不阻塞判定（distanceToNearestEntry 仍 work，但提示运维 layout 有问题）。
	if !a.layoutWarned[roomID] && len(entryPoints) >= 2 && hasDisconnectedEntries(entryPoints, grid.CellSize) {
		a.logger.Warn("bathroom_entry_layout_multi_component",
			zap.String("room_id", roomID),
			zap.Int("entry_cells", len(entryPoints)),
			zap.String("hint", "decision_15 expects single contiguous entry region"),
		)
		a.layoutWarned[roomID] = true
	}

	// Step 3：跑 4 条规则，按 trackID 聚合 delta
	pending := make(map[int]*pendingGhost)
	markGhost := func(trackID int, reason string) {
		d, ok := pending[trackID]
		if !ok {
			d = &pendingGhost{}
			pending[trackID] = d
		}
		ghostV := VerdictGhost
		d.newVerdict = &ghostV
		d.penalty = 100
		d.reasons = append(d.reasons, reason)
	}

	// 规则 1：出生地判据（仅 entry 在 FOV 才可信）
	if entryInFOV {
		for i := range bases {
			b := &bases[i]
			if !newBorn[b.TrackID] {
				continue
			}
			distCm := distanceToNearestEntry(b.X, b.Y, entryPoints)
			if distCm > bathroomBirthEntryThresholdCm {
				markGhost(b.TrackID, "rule1_birth_far_from_entry")
				a.logger.Info("bathroom_ghost_rule1",
					zap.String("room_id", roomID),
					zap.Int("track_id", b.TrackID),
					zap.Int("birth_dist_cm", distCm),
					zap.Int("threshold_cm", bathroomBirthEntryThresholdCm),
				)
			}
		}
	}

	// 规则 5：split-ghost 出生邻近性（不依赖 entry，独立有效）
	for i := range bases {
		b := &bases[i]
		if !newBorn[b.TrackID] {
			continue
		}
		for j := range bases {
			if i == j {
				continue
			}
			other := &bases[j]
			if newBorn[other.TrackID] {
				continue
			}
			if other.Verdict == VerdictGhost {
				continue
			}
			d := euclideanDistCm(b.X, b.Y, other.X, other.Y)
			if d < bathroomSplitGhostDistCm {
				markGhost(b.TrackID, "rule5_split_ghost_near_real")
				a.logger.Info("bathroom_ghost_rule5",
					zap.String("room_id", roomID),
					zap.Int("track_id", b.TrackID),
					zap.Int("source_track_id", other.TrackID),
					zap.Int("dist_cm", d),
				)
				break
			}
		}
	}

	// 规则 3：共生律自校正（仅 entry 在 FOV 才可信；不翻 verdict 只修 count）
	//
	// 已知 limitation：trackCount 包含已被 Rule 1/5 标 ghost 的 track（Rule 3 不区分 verdict）。
	// 极端场景（≥50% track 已是 ghost）下 expectedMin 偏低，Rule 2 在 count 修正后可能多杀 real。
	// 决定 14 共生律下 fall 不因 ghost 抑制——告警触发主体只要"看到的真人 track ≥ 1"即可，
	// 多杀的 real 不影响 fall 检测。接受此 limitation 换实现简洁；
	// 如未来 fall 主体改用精确 real count，再用 trackCount - alreadyGhostCount 计算 expectedMin。
	if entryInFOV && suiteID != "" && a.census != nil {
		trackCount := len(bases)
		if trackCount > 0 {
			expectedMin := (trackCount + bathroomMaxGhostsPerPerson - 1) / bathroomMaxGhostsPerPerson
			count := a.census.GetBathroomCount(suiteID)
			if count >= 0 && count < expectedMin {
				delta := expectedMin - count
				a.census.AdjustBathroomCount(suiteID, delta, nowMs)
				a.logger.Warn("gate_undercount_corrected",
					zap.String("room_id", roomID),
					zap.String("suite_id", suiteID),
					zap.Int("track_count", trackCount),
					zap.Int("expected_min", expectedMin),
					zap.Int("count_before", count),
					zap.Int("count_after", count+delta),
				)
			}
		}
	}

	// 规则 2：占用一致性（仅 entry 在 FOV 才可信；放在规则 3 之后让 count 已修正）
	if entryInFOV && suiteID != "" && a.census != nil {
		count := a.census.GetBathroomCount(suiteID)
		liveCount := len(bases)
		if count >= 0 && liveCount > count {
			// 找未被规则 1/5 标 ghost 且 verdict 非 ghost 的候选；按距 entry 远近降序，远的优先 ghost
			type tidDist struct {
				tid    int
				distCm int
			}
			candidates := make([]tidDist, 0, liveCount)
			for i := range bases {
				b := &bases[i]
				if _, already := pending[b.TrackID]; already {
					continue
				}
				if b.Verdict == VerdictGhost {
					continue
				}
				candidates = append(candidates, tidDist{b.TrackID, distanceToNearestEntry(b.X, b.Y, entryPoints)})
			}
			// 距 entry 远的优先 ghost；distCm 相同时按 TrackID 升序作 tiebreaker，
			// 让多 track 等距时 verdict 选择确定（避免 sort.Slice 非稳定排序导致
			// verdict 帧间跳动 — 下游 fall verifier / dev playback 都需要稳定信号）。
			sort.Slice(candidates, func(i, j int) bool {
				if candidates[i].distCm != candidates[j].distCm {
					return candidates[i].distCm > candidates[j].distCm
				}
				return candidates[i].tid < candidates[j].tid
			})
			excess := liveCount - count
			for i := 0; i < excess && i < len(candidates); i++ {
				markGhost(candidates[i].tid, "rule2_excess_over_bathroom_count")
				a.logger.Info("bathroom_ghost_rule2",
					zap.String("room_id", roomID),
					zap.Int("track_id", candidates[i].tid),
					zap.Int("dist_cm", candidates[i].distCm),
					zap.Int("live_count", liveCount),
					zap.Int("bathroom_count", count),
				)
			}
		}
	}

	// Step 4：collect deltas
	if len(pending) == 0 {
		return nil
	}
	result := make([]VerdictDelta, 0, len(pending))
	for tid, d := range pending {
		result = append(result, VerdictDelta{
			TrackID:      tid,
			NewVerdict:   d.newVerdict,
			PenaltyDelta: d.penalty,
			Reason:       strings.Join(d.reasons, ","),
		})
	}

	// 汇总 log：prod 调参用——监控"bathroom X 每分钟产生 N 个 ghost verdict"频率，
	// 比逐 rule log 更易聚合。仅在有输出时打（无 ghost 不打 = 静默正常态）。
	a.logger.Info("bathroom_ghost_adjudication_summary",
		zap.String("room_id", roomID),
		zap.Int("track_count", len(bases)),
		zap.Int("ghost_count", len(result)),
		zap.Bool("entry_in_fov", entryInFOV),
		zap.Int("entry_cells", len(entryPoints)),
		zap.Int64("ts_ms", nowMs),
	)
	return result
}

// hasDisconnectedEntries 检测 entry cells 是否分多个不连续片（layout error）。
// 算法：BFS 通过"邻居 = cell-center 距离 ≤ cellSize × 1.5"（含八方向邻接）传递性聚类，
// 起点 cell 出发 reachable 数量 < total → 有不可达片 → 多片。
//
// 单 cell entry 与 N 个独立 cell entry 都不算多片（前者天然连续，后者每片各 1 cell —
// 但 layout error 主要 case 是"门洞 2 cells + 另一片 2 cells"形态，N=4 总数下 2 个不可达即触发）。
//
// 复杂度 O(N²)，N 通常 < 10（door 物理大小有限），单帧成本可忽略。
func hasDisconnectedEntries(eps []bathroomEntryPoint, cellSize int) bool {
	if len(eps) <= 1 {
		return false
	}
	threshold := int(float64(cellSize) * 1.5)
	visited := make([]bool, len(eps))
	queue := []int{0}
	visited[0] = true
	reachable := 1
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for i := range eps {
			if visited[i] {
				continue
			}
			if euclideanDistCm(eps[cur].cx, eps[cur].cy, eps[i].cx, eps[i].cy) <= threshold {
				visited[i] = true
				reachable++
				queue = append(queue, i)
			}
		}
	}
	return reachable < len(eps)
}

// pendingGhost 多规则聚合用的中间态。
type pendingGhost struct {
	newVerdict *TrackVerdict
	penalty    int
	reasons    []string
}

// findBathroomEntryPoints 扫 grid 找出所有 EnterTarget=="bathroom" cells。
// 返回每个 cell 的中心画布坐标 + InFOV 标记（规则 4 fallback 判定用）。
func findBathroomEntryPoints(grid *RoomGrid) []bathroomEntryPoint {
	var out []bathroomEntryPoint
	for row := 0; row < grid.Height; row++ {
		for col := 0; col < grid.Width; col++ {
			c := &grid.Cells[row*grid.Width+col]
			if c.EnterTarget != "bathroom" {
				continue
			}
			cx, cy := grid.ToCanvas(col, row)
			out = append(out, bathroomEntryPoint{cx: cx, cy: cy, inFOV: c.InFOV})
		}
	}
	return out
}

// distanceToNearestEntry 给定画布坐标 (x, y) 到最近 entry 点的欧氏距离（cm）。
// 无 entry 时返回 MaxInt32（caller 据此判定"无入口信号"）。
func distanceToNearestEntry(x, y int, eps []bathroomEntryPoint) int {
	if len(eps) == 0 {
		return math.MaxInt32
	}
	minD := math.MaxInt32
	for _, ep := range eps {
		d := euclideanDistCm(x, y, ep.cx, ep.cy)
		if d < minD {
			minD = d
		}
	}
	return minD
}

// euclideanDistCm 整数欧氏距离（cm）。
func euclideanDistCm(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
}
