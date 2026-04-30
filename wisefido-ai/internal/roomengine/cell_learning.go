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
	NearTraverseWalk int
	NearTraverseDeny int
}

// DefaultLearnParams 与 config.yaml::roomengine.learn 默认值一致。
// playback / 测试场景调用，prod 路径由 engine.Configure(...) 注入。
func DefaultLearnParams() LearnParams {
	return LearnParams{
		WalkActiveX10:     20,
		WalkTraverse:      2, // 配合 speed 兜底：原 5 太高，散点单次穿越凑不齐；降到 2 让"走过 2 次"即升格
		SitActiveX10:      150,
		LieAnomalyX10:     50,
		BedToleranceCm:    30,
		ToiletToleranceCm: 20,
		ShowerToleranceCm: 20,
		ConfFloor:         60,
		ConfFull:          95,
		MoveSpeedCms:      20,
		NearTraverseWalk:  5,
		NearTraverseDeny:  20,
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
//  5. Auto-Deny 反转（邻居走过 ≥ NearTraverseDeny 次 + 本 cell 完全无直接触碰）
//     —— "20 次都没经过" = 系统性绕开 = 障碍
//
// 4 与 5 是同一信号 NearTraverseCount 的双因素解释：
//   - 有直接 Move 触碰（TraverseCount > 0）→ 低阈值即确认 Walk
//   - 完全无直接触碰（RealDecay == 0 且 TraverseCount == 0）→ 高阈值反转 Deny
//
// 写入 3 组 Belief 都标 SourceLearned。
// 同 type 升格用 max(cur, mapped) 累积——多日观测可把 Confidence 推向 ConfFull（默认 95）。
func (g *RoomGrid) LearnCellAreas(p LearnParams) {
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

		// Auto-Deny 反转：邻居走过 ≥ NearTraverseDeny + 5-cell consensus（中心 + 4 邻居全无直接触碰）
		// PR-15：原规则只看中心 cell TraverseCount==0，10cm 网格上 jitter / 跳格频繁让中心未被穿过 → 误判。
		// 改用中心+上下左右 5 cell 的"完全绕开"共识：任一 cell 有 TraverseCount>0 或 RealDecay>0 即否决。
		// 物理含义：人若实际能走到这一带（哪怕中心 cell 没被精确踩中），5 cell 至少有一个会被穿过；
		// 此规则覆盖"furniture 边缘紧贴 walk"的 case（counter / sofa 边）。
		//
		// PR-15.2 实验：尝试用 BFS 距离场补 island 核心检测——失败回退。
		// 失败原因：island 核心 cell 的 NearTraverseCount=0（其邻居也是 island 内 cell），
		// BFS 加门控到不了核心；BFS 不加门控会误覆盖 walk 衰减的 cell + 无观测角落。
		// 结论：island 核心仍需 layout 显式标 Furniture（人标 SourceHuman 不会被算法覆写）。
		if int(c.NearTraverseCount) >= p.NearTraverseDeny &&
			(t == AreaUnknown || t == AreaActive) &&
			fiveCellAllUnreached(g, i) {
			conf := mapToConf(int(c.NearTraverseCount), p.NearTraverseDeny, p.NearTraverseDeny*3, p.ConfFloor, p.ConfFull)
			promoteCell(c, AreaDeny, conf)
			continue
		}
	}
}

// fiveCellAllUnreached 检查 cell + 4 邻居（N/S/E/W）是否全部"基本未被走"。
//
// PR-15 Auto-Deny 5-cell consensus：减少 jitter / 跳格导致的"中心 cell 永不被走"误判。
// PR-15.1 RealDecay < realDecayDenyTolerance：容忍偶发轻触碰。
//
// RealDecay 半衰期 15min，一次触碰约 60min 内衰到 < 1；阈值 5 容忍"≤ 30min 内一次轻触"。
// TraverseCount 长档 7 天；要求严格 == 0（Move 状态穿越是强证据）。
// 越界邻居视为 unreached（不否决；房间边角不因邻居越界无法学 Deny）。
const realDecayDenyTolerance = 5

func fiveCellAllUnreached(g *RoomGrid, idx int) bool {
	if g.Cells[idx].RealDecay >= realDecayDenyTolerance || g.Cells[idx].TraverseCount != 0 {
		return false
	}
	row := idx / g.Width
	col := idx % g.Width
	deltas := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range deltas {
		r, cc := row+d[0], col+d[1]
		if r < 0 || r >= g.Height || cc < 0 || cc >= g.Width {
			continue // 越界跳过
		}
		nb := &g.Cells[r*g.Width+cc]
		if nb.RealDecay >= realDecayDenyTolerance || nb.TraverseCount != 0 {
			return false
		}
	}
	return true
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
