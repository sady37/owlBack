package roomengine

// cell_learning.go
//
// 硬阈值"区域升格"规则——独立于 UpdateBelief 的概率似然路径。
// 设计哲学（用户确认）：
//   - layout 提供精确尺寸（床、马桶、淋浴），位置可能有误差 → 用容差判断"在床附近"
//   - 椅子/沙发 layout 经常没标 → 必须靠 ActiveTime[Sit] 学
//   - 走道 layout 不标 → 必须靠 ActiveTime[Move] + TraverseCount 学
//   - 床外 Lie 是跌倒嫌疑 → 单独累计 LieAnomalyCount 给 fall 检测消费
//
// 与 UpdateBelief 的分工：
//   - 本文件（硬规则）：达到阈值即写 Belief[].Type = SourceLearned，确定性结论
//   - UpdateBelief（软规则）：基于似然 + Confidence 的概率演化，用于细粒度风险评分
//   两者并存，硬规则优先（Source==Human 不覆写；Source==Learned 可被刷成更强证据）

// LearnParams cell_learning 硬阈值规则的运行时参数。
// 由 wisefido-ai/internal/config 从 yaml 加载，engine 在 scanBeliefAll 时传入。
//
// 时间档关系（决定能不能跨天累积）：
//   - ActiveType[Move] 半衰期 3 d，单次穿过 2s 易满足
//   - ActiveType[Sit]  半衰期 24 h，连续 15s 即升格
//   - ActiveType[Lie]  半衰期 7 d，床外 5s 即记 LieAnomaly
//   - TraverseCount    半衰期 7 d（长档），跨天累积稀疏事件
//
// 升格策略：
//   - Walk: 跨天累积 5 次穿越 + 单次 ≥ 2s → AreaActive
//   - Sit:  连续 15s 坐姿 → AreaSit（优先于 Walk）
//   - Lie:  床外 5s 躺 → 累计 LieAnomaly 给 fall 检测消费
type LearnParams struct {
	WalkActiveX10     int // Walk 单帧门槛（0.1 秒为单位，默认 20 = 2s）
	WalkTraverse      int // Walk 跨天累计（默认 2 次）
	SitActiveX10      int // Sit 升格阈值（默认 150 = 15s）
	LieAnomalyX10     int // Lie 异常阈值（默认 50 = 5s）
	BedToleranceCm    int // layout 床矩形容差（cm，默认 30）
	ToiletToleranceCm int // layout 马桶矩形容差（cm，默认 20）
	ShowerToleranceCm int // layout 淋浴矩形容差（cm，默认 20）
	ConfFloor         int // promoteCell 起跑置信度（默认 60）
	ConfFull          int // promoteCell 上限置信度（默认 95）

	// MoveSpeedCms：Kalman 速度阈值（cm/s），TrackManager 用作"在动"兜底判定。
	// 雷达 pose 把慢走老人误报 Standing → ActiveType[Move] 永远不涨；
	// 此阈值让"pose 不是 Walking 但速度 > 阈值"也能升为 Move pose。
	// 默认 20 cm/s（≈2 cells/s，慢走门槛；坐立 jitter 一般 < 15）
	MoveSpeedCms int

	// NearTraverse 双档：邻居被穿越次数 + 本 cell 直接触碰 = 双因素确认。
	//   - NearTraverseWalk（默认 5）：邻居走过 ≥ 5 次 + 本 cell 至少有 1 次直接 Move 穿过 → 确认 Walk
	//     （"一次经过就确认"：直接证据强；只有邻居证据无直接证据时不升 Walk，留给 Deny 判断）
	//   - NearTraverseDeny（默认 20）：邻居走过 ≥ 20 次 + 本 cell 完全无直接触碰 → 反转 Deny
	//     （"20 次都没经过 = 系统性绕开 = 障碍"）
	// 顺序：Walk 优先（有直接证据时立即确认），再 Deny（没直接证据时高阈值反转）。
	//
	// PR-15.3 — Walk vs Auto-Deny 扩散速度差异：
	//   Walk 扩散快：阈值低（5）+ 直接 Move 触碰即学，几小时-1 天内升格
	//   Auto-Deny 扩散慢：阈值高（20）+ 5-cell 软共识 + 持续 15 天时间门控
	//   原因：Auto-Deny 后果更严重（拒绝 fall 检测），需要更严格的证据
	NearTraverseWalk int
	NearTraverseDeny int

	// PR-15.3 Auto-Deny 时间门控（与 cell.AutoDenyQualifiedSinceMs 配合）。
	//
	// 双阈值滞后设计（防抖 + 抗噪声）：
	//   - 5-cell 软共识用 TraverseCount < AutoDenyTraverseTolerate（默认 3）
	//     容忍背景 ghost/jitter ≈ 0.3/天 噪声，类 PR-13 sit 学习的 90% 容忍
	//   - reset 用 TraverseCount >= AutoDenyTraverseReset（默认 5）
	//     高于此说明真有人走过，重置时间计时
	//   - hysteresis 区间 [3, 5)：维持当前状态，不前进不重置（避免阈值附近抖动）
	//
	// 时间门控（默认 10 天，PR-15.5 实测调整）：cell 持续满足条件 ≥ 10 天才升 AreaDeny。
	// 物理含义：10 天人都不踩到的 cell，几乎可肯定是家具/障碍。
	//   - 老 walk 已搬走的 cell：10 天内重新走过 → reset 重学
	//   - 真 furniture 核心：10 天持续无穿越 → 升格
	//
	// 实测对比 (Kitchen-min layout 5 cycles 16.25 天等效数据)：
	//   15d/3/5：第 5 cycle (~16d) 触发，最终 200 cells
	//   10d/3/5（当前默认）：第 4 cycle (~13d) 触发，最终 200 cells（加速 5 天）
	//   10d/2/4（更严）：触发时机相同，最终 cell 数相同
	// → 时间门控是主导因素；tolerance/reset 在 Kitchen-min 双峰分布影响小，保留 3/5 抗噪
	//
	// playback 限制：3 天数据 < 10 天阈值 → 单 cycle playback 中此规则不生效（预期）
	AutoDenyMinPersistDays   int // 时间门控最少持续天数（默认 10）
	AutoDenyTraverseTolerate int // 5-cell 软共识容忍 TraverseCount 上限（< 此值算"未走过"，默认 3）
	AutoDenyTraverseReset    int // 重置时间计时的 TraverseCount 阈值（>= 此值视为"真走过"，默认 5）
}

// DefaultLearnParams 与 config.yaml::roomengine.learn 默认值一致。
// playback / 测试场景调用，prod 路径由 engine.Configure(...) 注入。
func DefaultLearnParams() LearnParams {
	return LearnParams{
		WalkActiveX10:            20,
		WalkTraverse:             2, // 配合 speed 兜底：原 5 太高，散点单次穿越凑不齐；降到 2 让"走过 2 次"即升格
		SitActiveX10:             150,
		LieAnomalyX10:            50,
		BedToleranceCm:           30,
		ToiletToleranceCm:        20,
		ShowerToleranceCm:        20,
		ConfFloor:                60,
		ConfFull:                 95,
		MoveSpeedCms:             20,
		NearTraverseWalk:         5,
		NearTraverseDeny:         20,
		AutoDenyMinPersistDays:   10, // PR-15.5: 10 天持续门控（实测 vs 15d 加速 5 天，结果一致）
		AutoDenyTraverseTolerate: 3,  // 软共识：TraverseCount<3 算未走过（容忍背景噪声 ~0.3/天）
		AutoDenyTraverseReset:    5,  // 重置：TraverseCount>=5 视为真走过 → 重置计时
	}
}

// Deprecated: 兼容旧代码。新代码用 LearnParams。
const (
	LearnWalkActiveX10 = 20
	LearnWalkTraverse  = 5
	LearnSitActiveX10  = 150
	LearnLieAnomalyX10 = 50
	BedToleranceCm     = 30
	ToiletToleranceCm  = 20
	ShowerToleranceCm  = 20
)

// LearnCellAreas 扫全图，按硬阈值升格 cell.Belief[].Type。
//
// 规则优先级（自上而下）：
//  1. Source==SourceHuman 的 cell 不覆写（人标神圣）
//  2. Sit 阈值（停留语义最强）
//  3. Walk 直升（直接走过 ≥ WalkTraverse 次）
//  4. Walk 邻接确认（邻居走过 ≥ NearTraverseWalk 次 + 本 cell 至少 1 次直接 Move 穿过）
//     —— "一次经过就确认"：直接证据强，立即确认 Walk
//  5. Auto-Deny 反转（PR-15.4 双触发路径）
//     A. NearTraverseCount ≥ 20（walk 边缘 furniture）
//     B. BFS depth ≥ 2 或不可达（island 核心 / 不可达角落）
//     A 或 B 任一 + 5-cell 软共识 + 持续 15 天
//
// 写入 3 组 Belief 都标 SourceLearned。
// 同 type 升格用 max(cur, mapped) 累积——多日观测可把 Confidence 推向 ConfFull（默认 95）。
//
// nowMs 用于 PR-15.3 时间门控；调用方传当前模拟时间（playback）或 wall-clock（prod）。
func (g *RoomGrid) LearnCellAreas(p LearnParams, nowMs int64) {
	// PR-15.4: 一次性计算全图 BFS 距离场（从 walk/enter/sit/bed 出发），
	// 用于 Auto-Deny 路径 B 抓 island 核心。
	dist := g.computeWalkDistance()
	for i := range g.Cells {
		c := &g.Cells[i]
		if !c.InRoom || !c.InFOV {
			continue
		}
		// 人标 cell 不动
		if c.Belief[0].Source == SourceHuman {
			continue
		}

		// PR-15: AreaSit 升格已禁用此简单规则。
		// 旧规则 ActiveType[Sit] >= 15s 触发 → 雷达近场常误判 pose=Sit 累积；
		// 实测在 D523 bookroom 雷达下方 (~1.5m 近场) 学出假的 Sit 区。
		// 现在 AreaSit 仅由两条更严格路径产生：
		//   - PR-13 RegionStatic A 路径：region static ≥2min + |dz| ≥10cm（双方 z>0）
		//   - PR-13 RegionStatic B 路径：region static ≥8-12min + 90% 静止帧比
		//   - PR-7.2 stand-static：pose=Stand 静止 ≥8-12min + 非 still-fall 作用域
		// 三条都直接调 MarkRestZoneByFeedback，无需走此 LearnCellAreas 路径。

		// Walk 直升：直接走过 ≥ WalkTraverse 次（默认 2）
		t := c.Belief[0].Type
		if int(c.ActiveType[ActiveIdxMove]) >= p.WalkActiveX10 &&
			int(c.TraverseCount) >= p.WalkTraverse {
			conf := mapToConf(int(c.TraverseCount), p.WalkTraverse, p.WalkTraverse*5, p.ConfFloor, p.ConfFull)
			if t == AreaUnknown || t == AreaActive {
				promoteCell(c, AreaActive, conf)
			}
			continue
		}

		// Walk 邻接确认：邻居走过 ≥ NearTraverseWalk + 本 cell 有直接 Move 触碰（TraverseCount > 0）
		// "一次经过就确认"：哪怕只走过 1 次（不够 WalkTraverse=2 直升），有邻居证据也算 Walk
		if int(c.NearTraverseCount) >= p.NearTraverseWalk && c.TraverseCount > 0 {
			if t == AreaUnknown {
				conf := mapToConf(int(c.NearTraverseCount), p.NearTraverseWalk, p.NearTraverseWalk*4, p.ConfFloor, p.ConfFull-10)
				promoteCell(c, AreaActive, conf)
			}
			continue
		}

		// Auto-Deny 反转 (PR-15.4 simplified) — 单一 BFS 距离场 + 软共识 + 15 天门控
		//
		// 设计原理（差异化扩散模型）：
		//   Walk 扩散快：阈值低（5 邻居） + 直接 Move 触碰即学，几小时-1 天升格
		//   Auto-Deny 扩散慢：BFS 距离 ≥2 + 5-cell 软共识 + 持续 15 天才升格
		//   原因：Auto-Deny 后果严重（拒绝 fall 检测），需更严格证据
		//
		// 触发条件 (AND)：
		//   - BFS depth ≥ 2 (dist[i] >= 2 或 dist[i] == -1 不可达)
		//     物理含义：cell 距任何"活动" cell ≥ 20cm 或几何隔绝
		//     Walk + Enter + Sit/Bed/Toilet/Shower 是 BFS source；
		//     仅传播过 AreaUnknown（停留区不传播；不污染 island）
		//   - 5-cell 软共识：center + 4 邻居 TraverseCount < 3 + RealDecay < 5
		//     容忍背景 ghost/jitter ~0.3/天 噪声（类 PR-13 sit 学习的 90% 容忍）
		//   - cell 当前是 AreaUnknown（Walk/Sit/Bed 不动）
		//   - 15 天时间门控（cell.AutoDenyQualifiedSinceMs）：持续满足才升 Deny
		//
		// reset：TraverseCount >= 5 或 RealDecay >= 5（真活动证据）→ 重置计时
		// hysteresis 区 TraverseCount [3, 5)：维持现状（既不前进也不重置）
		//
		// 自纠错：cell 已是 AreaDeny + 被走过（TraverseCount >= 5）→ 回退 Unknown
		// 让 Walk 规则重新评估（应对家具搬走/重布局）
		//
		// 历史教训：
		//   PR-15.0/15.1 严格 ==0：jitter 单次触发即重置
		//   PR-15.2 BFS 无时间门控：早期误标
		//   PR-15.3 仅 NearTraverseCount：抓不到 island 核心（互斥几何）
		//   PR-15.4 = BFS 距离场 + 5-cell 软共识 + 15 天门控 + 自纠错

		// 自纠错：Deny cell 被走过 → 回退 Unknown
		resetThresh := p.AutoDenyTraverseReset
		if resetThresh < 1 {
			resetThresh = 5
		}
		hasRealActivity := int(c.TraverseCount) >= resetThresh ||
			c.RealDecay >= realDecayDenyTolerance
		if t == AreaDeny && c.Belief[0].Source == SourceLearned && hasRealActivity {
			c.Belief[0] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
			c.Belief[1] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
			c.Belief[2] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
			c.AreaType = AreaUnknown
			c.AutoDenyQualifiedSinceMs = 0
			continue
		}

		qualifies := (dist[i] >= 2 || dist[i] == -1) &&
			t == AreaUnknown &&
			fiveCellSoftConsensus(g, i, p.AutoDenyTraverseTolerate)

		if hasRealActivity {
			c.AutoDenyQualifiedSinceMs = 0 // 重置计时
		} else if qualifies {
			if c.AutoDenyQualifiedSinceMs == 0 {
				c.AutoDenyQualifiedSinceMs = nowMs // 首次满足，启动 15 天计时
			} else {
				persistDays := p.AutoDenyMinPersistDays
				if persistDays < 1 {
					persistDays = 15
				}
				if nowMs-c.AutoDenyQualifiedSinceMs >= int64(persistDays)*24*3600*1000 {
					// 持续 ≥15 天 → 升格 AreaDeny
					// confidence 用 BFS 距离映射（深度越大越自信，cap ConfFull）
					depthFactor := dist[i]
					if depthFactor < 0 {
						depthFactor = 5 // 不可达视为强证据
					}
					conf := mapToConf(depthFactor, 2, 5, p.ConfFloor, p.ConfFull)
					promoteCell(c, AreaDeny, conf)
					c.AutoDenyQualifiedSinceMs = 0
					continue
				}
			}
		}
		// 否则（!qualifies && !resetCondition）：保持 hysteresis 区，计时不前进也不重置
	}
}

// fiveCellSoftConsensus 检查 cell + 4 邻居（N/S/E/W）是否全部"基本未被走"——软共识。
//
// PR-15.3：从严格 TraverseCount==0 改为 TraverseCount<traverseTolerate（默认 3），
// 容忍背景 ghost/jitter 噪声（实测 ~0.3/天）。类 PR-13 sit 学习的 90% 容忍思路。
//
// 检查项：
//   - RealDecay < realDecayDenyTolerance（5）：30 min 内无 Real 触碰（HL 15min）
//   - TraverseCount < traverseTolerate（3）：未累积"显著"Move 穿越（HL 7d，3 次≈背景上限）
//
// 越界邻居视为 unreached（不否决；房间边角不因邻居越界无法学 Deny）。
const realDecayDenyTolerance = 5

func fiveCellSoftConsensus(g *RoomGrid, idx int, traverseTolerate int) bool {
	if traverseTolerate < 1 {
		traverseTolerate = 3
	}
	c := &g.Cells[idx]
	if c.RealDecay >= realDecayDenyTolerance || int(c.TraverseCount) >= traverseTolerate {
		return false
	}
	row := idx / g.Width
	col := idx % g.Width
	deltas := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range deltas {
		r, cc := row+d[0], col+d[1]
		if r < 0 || r >= g.Height || cc < 0 || cc >= g.Width {
			continue
		}
		nb := &g.Cells[r*g.Width+cc]
		if nb.RealDecay >= realDecayDenyTolerance || int(nb.TraverseCount) >= traverseTolerate {
			return false
		}
	}
	return true
}

// computeWalkDistance PR-15.4：从所有"活动" cell（Walk/Enter/Sit/Bed/Toilet/Shower）
// 出发做 4-连通 BFS，传播仅通过 AreaUnknown 的 cell（停留区作为 source 而非 propagator
// 防止 BFS 跨 island 内部）。
//
// 返回 dist[i]：cell i 到最近"活动" cell 的最短跳数。
//   dist[i] = 0：cell 是 source（活动区）
//   dist[i] >= 1：cell 是 Unknown，且 BFS 可达
//   dist[i] == -1：cell 不可达（被 Deny 或几何隔绝）→ Auto-Deny 路径 B 也接受
//
// 用途：Auto-Deny 路径 B：dist >= 2 或 -1 视为"远离 walk" → island 核心候选。
// 调用方持锁；scoping in caller LearnCellAreas。
func (g *RoomGrid) computeWalkDistance() []int {
	dist := make([]int, len(g.Cells))
	for i := range dist {
		dist[i] = -1
	}
	queue := make([]int, 0, 256)
	for i := range g.Cells {
		c := &g.Cells[i]
		if !c.InRoom || !c.InFOV {
			continue
		}
		switch c.Belief[0].Type {
		case AreaActive, AreaEnter, AreaSit, AreaBed, AreaToilet, AreaShower:
			dist[i] = 0
			queue = append(queue, i)
		}
	}
	deltas := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for head := 0; head < len(queue); head++ {
		idx := queue[head]
		row := idx / g.Width
		col := idx % g.Width
		for _, d := range deltas {
			r, cc := row+d[0], col+d[1]
			if r < 0 || r >= g.Height || cc < 0 || cc >= g.Width {
				continue
			}
			ni := r*g.Width + cc
			if dist[ni] >= 0 {
				continue
			}
			n := &g.Cells[ni]
			if !n.InRoom || !n.InFOV {
				continue
			}
			// 仅传播过 AreaUnknown：stop at Deny / 已学到的活动区（不污染 island）
			if n.Belief[0].Type != AreaUnknown {
				continue
			}
			dist[ni] = dist[idx] + 1
			queue = append(queue, ni)
		}
	}
	return dist
}

// LearnLyingAnomalies 扫全图，把"床外 Lie"事件累计到 LieAnomalyCount。
// 床/沙发/卫生间内的 Lie 不算异常（卫生间已是高风险，由 StillTimeoutSec 把控）。
//
// 沙发签名：cell 自身 LieRetract > 0 → 已学到沙发，不算异常
//
// 此函数每次扫描会"增量"++ LieAnomalyCount（对达到阈值的 cell 触发一次），
// 跨周期复发会累积（Decay 长档 7 天，缓慢淡出）。fall_check 消费此字段做跌倒嫌疑判定。
func (g *RoomGrid) LearnLyingAnomalies(p LearnParams) int {
	hits := 0
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			c := &g.Cells[row*g.Width+col]
			if int(c.ActiveType[ActiveIdxLie]) < p.LieAnomalyX10 {
				continue
			}
			// 沙发签名：本 cell 学到回撤 → 不算异常
			if c.LieRetract > 0 {
				continue
			}
			x, y := g.ToCanvas(col, row)
			// 在床 / 卫生间 prior 容差内 → 不算异常
			if g.IsNearPriorType(x, y, AreaBed, p.BedToleranceCm) {
				continue
			}
			if g.IsNearPriorType(x, y, AreaToilet, p.ToiletToleranceCm) {
				continue
			}
			if g.IsNearPriorType(x, y, AreaShower, p.ShowerToleranceCm) {
				continue
			}
			// 已知坐区也排除（沙发坐着躺下属于沙发签名，前面 LieRetract 已挡；
			// 这里多挡一道 layout 标的椅子区域）
			if c.Belief[0].Type == AreaSit {
				continue
			}
			c.LieAnomalyCount++
			hits++
		}
	}
	return hits
}

// promoteCell 把 cell 升格到指定 AreaType（3 组 Belief 同步刷 SourceLearned）。
//
// 累积语义（关键改动）：
//   - 同 type 升格 → Confidence = max(cur, mapped)，多日观测把已升格 cell 推向 ConfFull
//   - 不同 type 升格 → Confidence 直接覆盖为 mapped（语义已变，旧分数无意义）
//
// 这样设计避免了"每次扫描都把 95 重置到 60-95 区间"的问题；
// 真正稳定的区域（每天都被坐到的椅子）经过几轮扫描即可稳坐 ConfFull。
func promoteCell(c *Cell, t AreaType, conf int) {
	for bi := 0; bi < 3; bi++ {
		if c.Belief[bi].Type == t && c.Belief[bi].Source == SourceLearned {
			// 同 type 累积：取 max（不允许新观测把已学到的高分数压低）
			if conf > c.Belief[bi].Confidence {
				c.Belief[bi].Confidence = conf
			}
		} else {
			// 不同 type 或首次升格：覆盖
			c.Belief[bi].Type = t
			c.Belief[bi].Confidence = conf
			c.Belief[bi].Source = SourceLearned
		}
	}
	c.AreaType = t
}

// mapToConf 把"实测值"映射到 [floor, full] 置信度区间。
// minVal = 阈值（刚过线 → floor），fullVal = 多少倍阈值后饱和（→ full）。
func mapToConf(val, minVal, fullVal, floor, full int) int {
	if val <= minVal {
		return floor
	}
	if val >= fullVal {
		return full
	}
	span := fullVal - minVal
	return floor + (val-minVal)*(full-floor)/span
}
