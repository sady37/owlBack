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

// 升格阈值（单位：ActiveType 的 0.1 秒，TraverseCount 次数，距离 cm）
//
// 时间档关系（决定能不能跨天累积）：
//   - ActiveType[Move/Sit/Lie] 半衰期 15 min（短档），只看"最近一段"
//   - TraverseCount 半衰期 7 d（长档），跨天累积稀疏事件
//
// 因此 Walk 学习的策略：
//   - LearnWalkTraverse 主导（每天 5 次穿越就升格）
//   - LearnWalkActiveX10 只做"门槛"——证明本帧确实是 Move 不是站立
//   - 阈值降到 20（2 秒），单次穿过即可命中
const (
	// Walk: 单次穿越 ≥ 2 秒 + 累计穿越 ≥ 5 次 → AreaActive
	LearnWalkActiveX10 = 20 // 2 sec × 10（门槛，单次穿过可达）
	LearnWalkTraverse  = 5  // 长档累计：跨天 5 次

	// Sit: 15 秒连续坐姿（短档窗口内一次性达标；3×3 内核抑制 hovering 散点）
	LearnSitActiveX10 = 150 // 15 sec × 10

	// Lie 异常：5 秒 Lie 在床外 → ++LieAnomalyCount（fall 检测消费）
	LearnLieAnomalyX10 = 50 // 5 sec × 10

	// 容差（cm）：layout 位置误差补偿
	BedToleranceCm    = 30
	ToiletToleranceCm = 20
	ShowerToleranceCm = 20
)

// LearnCellAreas 扫全图，按硬阈值升格 cell.Belief[].Type。
//
// 规则优先级：
//  1. Source==SourceHuman 的 cell 不覆写（人标神圣）
//  2. Sit 阈值优先 Walk（一个 cell 既被穿越又被坐过，按"人停留更久"判定 Sit）
//  3. Walk 升格只在 cell 当前是 Unknown/Active 时才生效（避免覆写已学到的 Sit）
//
// 写入 3 组 Belief 都标 SourceLearned，置信度按"超阈值多少"映射到 [60, 95]。
func (g *RoomGrid) LearnCellAreas() {
	for i := range g.Cells {
		c := &g.Cells[i]
		if !c.InRoom || !c.InFOV {
			continue
		}
		// 人标 cell 不动
		if c.Belief[0].Source == SourceHuman {
			continue
		}

		// Sit 优先（停留 > 穿越的语义更强）
		if c.ActiveType[ActiveIdxSit] >= LearnSitActiveX10 {
			conf := mapToConf(int(c.ActiveType[ActiveIdxSit]), LearnSitActiveX10, LearnSitActiveX10*4)
			promoteCell(c, AreaSit, conf)
			continue
		}

		// Walk 升格
		if c.ActiveType[ActiveIdxMove] >= LearnWalkActiveX10 &&
			c.TraverseCount >= LearnWalkTraverse {
			conf := mapToConf(int(c.TraverseCount), LearnWalkTraverse, LearnWalkTraverse*5)
			// 只覆写 Unknown / Active（不动 Bed / Toilet / Shower / Sit / Deny）
			t := c.Belief[0].Type
			if t == AreaUnknown || t == AreaActive {
				promoteCell(c, AreaActive, conf)
			}
		}
	}
}

// LearnLyingAnomalies 扫全图，把"床外 Lie"事件累计到 LieAnomalyCount。
// 床/沙发/卫生间内的 Lie 不算异常（卫生间已是高风险，由 StillTimeoutSec 把控）。
//
// 沙发签名：cell 自身 LieRetract > 0 → 已学到沙发，不算异常
//
// 此函数每次扫描会"增量"++ LieAnomalyCount（对达到阈值的 cell 触发一次），
// 跨周期复发会累积（Decay 长档 7 天，缓慢淡出）。fall_check 消费此字段做跌倒嫌疑判定。
func (g *RoomGrid) LearnLyingAnomalies() int {
	hits := 0
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			c := &g.Cells[row*g.Width+col]
			if c.ActiveType[ActiveIdxLie] < LearnLieAnomalyX10 {
				continue
			}
			// 沙发签名：本 cell 学到回撤 → 不算异常
			if c.LieRetract > 0 {
				continue
			}
			x, y := g.ToCanvas(col, row)
			// 在床 / 卫生间 prior 容差内 → 不算异常
			if g.IsNearPriorType(x, y, AreaBed, BedToleranceCm) {
				continue
			}
			if g.IsNearPriorType(x, y, AreaToilet, ToiletToleranceCm) {
				continue
			}
			if g.IsNearPriorType(x, y, AreaShower, ShowerToleranceCm) {
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
func promoteCell(c *Cell, t AreaType, conf int) {
	for bi := 0; bi < 3; bi++ {
		c.Belief[bi].Type = t
		c.Belief[bi].Confidence = conf
		c.Belief[bi].Source = SourceLearned
	}
	c.AreaType = t
}

// mapToConf 把"实测值"映射到 [60, 95] 置信度区间。
// minVal = 阈值（刚过线 → 60），fullVal = 4 倍阈值（饱和 → 95）。
func mapToConf(val, minVal, fullVal int) int {
	if val <= minVal {
		return 60
	}
	if val >= fullVal {
		return 95
	}
	span := fullVal - minVal
	return 60 + (val-minVal)*35/span
}
