package roomengine

import (
	"math"

	"owl-common/observation"
)

// ========================================================================
// 枚举定义
// ========================================================================

// AreaType 区域类型（按 engine 视角：允许什么行为，不是"这是什么家具"）
// 替代原 CellType 10 种 → 5 种功能分类。
type AreaType uint8

const (
	AreaUnknown AreaType = iota // 0 未知
	AreaEnter                   // 1 进入区（门）
	AreaBed                     // 2 可躺区
	AreaSit                     // 3 允许坐姿（沙发/椅子）
	AreaActive                  // 4 活动区（走廊、开放空间）
	AreaDeny                    // 5 禁止 track（墙/家具/Metal，track 出现 = Fake）
	AreaShower                  // 6 淋浴区（高风险：潮湿+跌倒）
	AreaToilet                  // 7 马桶（高风险：起身晕眩/坐姿突倒）
	NumAreaTypes
)

// Source 类型信息来源
type Source uint8

const (
	SourceUnset    Source = 0
	SourceHuman    Source = 1
	SourceLearned  Source = 2
	SourceGeometry Source = 3
)

// CorePose 核心姿态（从雷达 13 种 pose 还原到 4 种）。
// 常量值与 owl-common/observation/fields.go 里对应 Pose* 常量保持一致，
// 便于调试时直接对比源 pose 和还原后的核心姿态。
type CorePose uint8

const (
	CorePoseUnknown CorePose = CorePose(observation.PoseUnknown)  // 0
	CorePoseMove    CorePose = CorePose(observation.PoseWalking)  // 1
	CorePoseSit     CorePose = CorePose(observation.PoseSitting)  // 3（蹲坐，也是沙发坐姿的代表）
	CorePoseStand   CorePose = CorePose(observation.PoseStanding) // 4
	CorePoseLie     CorePose = CorePose(observation.PoseLying)    // 6
)

// RadarPoseToCore 把雷达 13 种 pose 还原到 4 种核心姿态。
// 核心哲学：不信雷达的 Suspected/Confirmed 分类和 area 派生事件——
// 疑似与确认归为同一核心姿态，由我们自己做时间过滤和事件检测。
// Pose 常量一律引用 owl-common/observation（唯一权威源）。
func RadarPoseToCore(pose int) CorePose {
	switch pose {
	case observation.PoseWalking, observation.PoseRunning:
		return CorePoseMove
	case observation.PoseStanding:
		return CorePoseStand
	case observation.PoseSitting, observation.PoseSuspectedSitGround, observation.PoseSitGround,
		observation.PoseBedSitUp, observation.PoseSuspectedBedSitUp, observation.PoseConfirmedBedSitUp:
		return CorePoseSit
	case observation.PoseLying, observation.PoseSuspectedFall, observation.PoseFallen:
		return CorePoseLie
	}
	return CorePoseUnknown
}

// ActiveType 索引常量（独立于 CorePose 的值，因为 CorePose 值不连续 0/1/3/4/6）
const (
	ActiveIdxMove  = 0
	ActiveIdxStand = 1
	ActiveIdxSit   = 2
	ActiveIdxLie   = 3
)

// CorePoseToActiveIdx CorePose → ActiveType 数组索引（-1 = 不累计）
func CorePoseToActiveIdx(core CorePose) int {
	switch core {
	case CorePoseMove:
		return ActiveIdxMove
	case CorePoseStand:
		return ActiveIdxStand
	case CorePoseSit:
		return ActiveIdxSit
	case CorePoseLie:
		return ActiveIdxLie
	}
	return -1
}

// ========================================================================
// Cell 结构体
// ========================================================================

// BeliefState 单组参数下对此 cell 的信念
type BeliefState struct {
	Type       AreaType // 当前判定类型
	Confidence int      // [0,100]
	RiskScore  int      // [0,100] 独立跌倒风险
	Source     Source
}

// Cell 网格中的一个像素（10cm × 10cm）。
//
// 分层原则（用户确认）：
//   - Cell = 空间属性容器（长期累计、区域推断、事件累计）
//   - Track = 运动粒子（Kalman、即时判定、状态机转移）
//   - 数据流：Track 每帧检测 → 触发事件 Mark 到 Cell → Cell 累计涌现空间属性
//
// 全 int/uint8/bool（系统 10cm 量化，小数无物理意义）。
type Cell struct {
	// ---- 几何（静态，NewRoomGrid + StampRoomPolygon 后固定）----
	InRoom bool
	InFOV  bool

	// ---- 物理烤入（构建时 rasterize，运行时只读）----
	EdgeDist int // 到信号域边缘的 XY 距离（cm）
	MaxZ     int // 该位置雷达可探测的目标最大 Z（cm）
	MinZ     int // Z 下限（cm）

	// ---- 核心空间属性 ----
	// ActiveType[i] 单位 = 0.1 秒（×10 定点），用 uint16 容纳到 6553s（≈1.8h）。
	// 定点是为了让 3×3 内核里"邻居 80% 权重"用整数表达（中心 +10×dt，邻居 +8×dt）。
	ActiveType        [4]uint16 // [Move, Stand, Sit, Lie] 累计 0.1 秒
	TraverseCount     uint16    // Move 状态下穿越本 cell 的次数（Walk 升格用）
	NearTraverseCount uint16    // 邻居被 Move 穿越的累计次数（auto-Deny 推断用 —— 兵家必争之地的"绕开"证据）
	AreaType          AreaType  // 推断属性，与 Belief[0].Type 镜像（query 加速）

	// ---- 访问可信度（track 层按本帧 quality 分桶喂）----
	RealDecay  int
	GhostDecay int

	// ---- 流动与停留（track 瞬时 → cell 累计）----
	FlowX, FlowY   int // 平均速度 EMA（cm/s）
	DwellEMA       int // 单次停留 EMA（秒）
	LongStillCount int // 静止 > 15min 次数

	// ---- 事件累计（track 层触发）----
	FallEventCount  int // prevCore=Stand/Move → Lie + Z 骤降
	LieRetract      int // Stand/Move → Lie → Stand/Move 短暂回撤（沙发签名）
	LieAnomalyCount int // 床外/沙发外的 Lie 累计（cell_learning 写入；fall 检测消费）

	// ---- 外部传感器事件 ----
	SleepadInBedCount   int
	SleepadLeftBedCount int
	DoorEventCount      int

	// ---- 信念（3 组并行参数，独立演化）----
	Belief [3]BeliefState

	LastUpdateMs int64
}

// ========================================================================
// 时间窗口（半衰期）
// ========================================================================

// DecayParams 按字段语义分档的半衰期（秒）。Decay() 用此结构决定每个 counter 的衰减速度。
//
// 设计原则：每个字段的半衰期反映该字段的"语义时间尺度"——
//   - 床/沙发/躺姿：人不挪 → 长（默认 7 d）
//   - 椅子/坐姿：偶尔搬 → 中（默认 24 h）
//   - 走道/Move：路径稳定但易污染 → 中长（默认 3 d）
//   - 即时态（Real/Ghost/Flow/Stand）：反映"最近一段在干嘛" → 短（默认 15 min）
//   - 事件类（Fall/Sleepad/Door/Traverse）：稀疏 → 长（默认 7 d 跨天累计）
//
// 由 wisefido-ai/internal/config 从 yaml 加载并传给 engine，engine 在 decayLoop / playback 中调用 DecayAll(dtSec, p)。
type DecayParams struct {
	ImmediateSec float64 // Real/Ghost/Flow/DwellEMA/ActiveType[Stand]
	WalkSec      float64 // ActiveType[Move]
	SitSec       float64 // ActiveType[Sit]
	LieSec       float64 // ActiveType[Lie]
	EventSec     float64 // Traverse/Fall/Retract/Sleepad/Door/LongStill/LieAnomaly
}

// DefaultDecayParams 与 config.yaml::roomengine.decay 默认值一致。
// playback / 测试场景调用，prod 路径由 engine.Configure(...) 注入。
func DefaultDecayParams() DecayParams {
	return DecayParams{
		ImmediateSec: 15 * 60,
		WalkSec:      3 * 24 * 3600,
		SitSec:       24 * 3600,
		LieSec:       7 * 24 * 3600,
		EventSec:     7 * 24 * 3600,
	}
}

// Deprecated: 兼容旧 playback / 测试。新代码用 DecayParams。
const (
	HalfLifeShort = 15 * 60       // 15 min
	HalfLifeLong  = 7 * 24 * 3600 // 7 d
)

// ========================================================================
// Helper 方法
// ========================================================================

// GhostRatio ghost 概率 [0,1]（float 返回给似然计算用）
func (c *Cell) GhostRatio() float64 {
	total := c.RealDecay + c.GhostDecay
	if total < 1 {
		return 0
	}
	return float64(c.GhostDecay) / float64(total)
}

// TypeOf 第 g 组信念下的 AreaType
func (c *Cell) TypeOf(g int) AreaType { return c.Belief[g].Type }

// IsEntry 是否为合法入口（Enter 区）。默认读 Belief[0]。
func (c *Cell) IsEntry() bool {
	return c.Belief[0].Type == AreaEnter
}

// IsRestZone 是否为合理静止区（Bed / Sit / Toilet）。默认读 Belief[0]。
func (c *Cell) IsRestZone() bool {
	t := c.Belief[0].Type
	return t == AreaBed || t == AreaSit || t == AreaToilet
}

// IsLikelyRestZone 比 IsRestZone 更宽松——除已升格的休息区外，
// cell 累计有 ≥ 3s 的 Sit/Lie 观测也算可疑休息区（家具未标 + 学习未到阈值的过渡情况）。
// 用于 silent fall 判断：拒报"曾经有人坐过/躺过"位置的静止丢失（防误报）。
//
// 30 = 3 秒 × 10 定点；远低于 SitActiveX10=150 (15s) 的升格门槛，
// 但已是足够强的"这里有人停留过"证据。
func (c *Cell) IsLikelyRestZone() bool {
	if c.IsRestZone() {
		return true
	}
	return c.ActiveType[ActiveIdxSit] >= 30 || c.ActiveType[ActiveIdxLie] >= 30
}

// StillTimeoutSec 静止超时阈值（秒；0 = 不限）。默认读 Belief[0]。
//
// isDay = true 时所有非零阈值 ×1.2（白天家属/护工活动多，放宽 20% 减误报）。
// isDay = false 时用基线值（夜间敏感）。基线值即此函数返回值（夜间标准）。
//
// 时段判定见 math_util.go::IsNightTime（23:30-07:30 = 夜）。
func (c *Cell) StillTimeoutSec(isDay bool) int {
	base := c.stillTimeoutBaseNight()
	if base == 0 {
		return 0
	}
	if isDay {
		// 白天 ×1.2，向上取整到秒
		return base + base/5
	}
	return base
}

// stillTimeoutBaseNight 夜间基线值（不带白天放宽因子）。
// 床/沙发：不限（休息合理）；马桶/淋浴：15min；Deny: 5min；其它：8min。
func (c *Cell) stillTimeoutBaseNight() int {
	switch c.Belief[0].Type {
	case AreaBed, AreaSit:
		return 0
	case AreaToilet:
		return 15 * 60 // 马桶 15 min（起身晕眩风险）
	case AreaShower:
		return 15 * 60 // 淋浴 15 min（潮湿跌倒风险）
	case AreaDeny:
		return 5 * 60 // Deny 区静止 5 min 即可疑
	}
	return 8 * 60 // Enter/Active/Unknown: 8 min
}

// ========================================================================
// Decay（时间窗口衰减）
// ========================================================================

// Decay 按字段语义分档衰减累计字段。
// 物理烤入（InRoom/InFOV/EdgeDist/MaxZ/MinZ）和 Belief 不衰减。
//
// 关键改动：ActiveType[Move/Sit/Lie] 不再共用 ImmediateSec，分别对应 WalkSec/SitSec/LieSec。
// 这样"床上躺姿"7天才衰一半、"椅子上坐姿"24h、"走道 Move"3d——符合 area 语义稳定性。
func (c *Cell) Decay(dtSec float64, p DecayParams) {
	if dtSec <= 0 {
		return
	}
	fImm := factor(dtSec, p.ImmediateSec)
	fWalk := factor(dtSec, p.WalkSec)
	fSit := factor(dtSec, p.SitSec)
	fLie := factor(dtSec, p.LieSec)
	fEv := factor(dtSec, p.EventSec)

	// 即时档 — 反映"最近这一段什么姿态/质量"
	c.RealDecay = scaleInt(c.RealDecay, fImm)
	c.GhostDecay = scaleInt(c.GhostDecay, fImm)
	c.FlowX = scaleInt(c.FlowX, fImm)
	c.FlowY = scaleInt(c.FlowY, fImm)
	c.DwellEMA = scaleInt(c.DwellEMA, fImm)
	c.ActiveType[ActiveIdxStand] = uint16(scaleInt(int(c.ActiveType[ActiveIdxStand]), fImm))

	// 分档 ActiveType
	c.ActiveType[ActiveIdxMove] = uint16(scaleInt(int(c.ActiveType[ActiveIdxMove]), fWalk))
	c.ActiveType[ActiveIdxSit] = uint16(scaleInt(int(c.ActiveType[ActiveIdxSit]), fSit))
	c.ActiveType[ActiveIdxLie] = uint16(scaleInt(int(c.ActiveType[ActiveIdxLie]), fLie))

	// 长档（事件类 / 稀疏次数）— 跨天积累
	c.TraverseCount = uint16(scaleInt(int(c.TraverseCount), fEv))
	c.NearTraverseCount = uint16(scaleInt(int(c.NearTraverseCount), fEv))
	c.LongStillCount = scaleInt(c.LongStillCount, fEv)
	c.FallEventCount = scaleInt(c.FallEventCount, fEv)
	c.LieRetract = scaleInt(c.LieRetract, fEv)
	c.LieAnomalyCount = scaleInt(c.LieAnomalyCount, fEv)
	c.SleepadInBedCount = scaleInt(c.SleepadInBedCount, fEv)
	c.SleepadLeftBedCount = scaleInt(c.SleepadLeftBedCount, fEv)
	c.DoorEventCount = scaleInt(c.DoorEventCount, fEv)
}

// factor 计算半衰期衰减因子；半衰期非正时不衰减（返回 1）
func factor(dtSec, halfLifeSec float64) float64 {
	if halfLifeSec <= 0 {
		return 1
	}
	return math.Pow(0.5, dtSec/halfLifeSec)
}

// ========================================================================
// UpdateBelief（信念更新 + 似然计算）
// ========================================================================

// ParamSet 单组信念更新参数（运行时配置，非持久化）
type ParamSet struct {
	Alpha  float64 // 一致证据增长系数
	Beta   float64 // 矛盾证据削弱系数
	FlipTh int     // Confidence 低于此值允许翻转
	Name   string
}

// UpdateBelief 用第 g 组参数更新第 g 组 Belief
func (c *Cell) UpdateBelief(g int, p ParamSet) {
	// 观测量门禁
	total := c.RealDecay + c.GhostDecay
	eventsSum := c.FallEventCount + c.LieRetract +
		c.SleepadInBedCount + c.SleepadLeftBedCount + c.DoorEventCount
	// activeSum 单位 = 0.1 秒；< 100 即 < 10 秒，证据不够更新 Belief
	activeSum := int(c.ActiveType[0]) + int(c.ActiveType[1]) +
		int(c.ActiveType[2]) + int(c.ActiveType[3])
	if total < 3 && eventsSum < 1 && activeSum < 100 {
		return
	}

	likelihoods := c.computeLikelihoods()

	curType := c.Belief[g].Type
	curL := likelihoods[curType]

	bestType := curType
	bestL := curL
	for t := AreaType(0); t < NumAreaTypes; t++ {
		if likelihoods[t] > bestL {
			bestType = t
			bestL = likelihoods[t]
		}
	}

	switch {
	case bestType == curType:
		delta := p.Alpha * float64(100-c.Belief[g].Confidence) * curL
		c.Belief[g].Confidence += int(math.Round(delta))
		if c.Belief[g].Confidence > 100 {
			c.Belief[g].Confidence = 100
		}

	case c.Belief[g].Confidence < p.FlipTh:
		c.Belief[g].Type = bestType
		c.Belief[g].Confidence = int(math.Round(bestL * 50))
		c.Belief[g].Source = SourceLearned

	default:
		delta := p.Beta * (bestL - curL) * 100
		c.Belief[g].Confidence -= int(math.Round(delta))
		if c.Belief[g].Confidence < 0 {
			c.Belief[g].Confidence = 0
		}
	}

	// 同步 AreaType mirror（当主组翻转时）
	if g == 0 {
		c.AreaType = c.Belief[0].Type
	}

	c.updateRisk(g)
}

// computeLikelihoods 算各 AreaType 的似然 ∈ [0,1]
func (c *Cell) computeLikelihoods() [NumAreaTypes]float64 {
	var out [NumAreaTypes]float64

	// ActiveType 分布
	total := int(c.ActiveType[0]) + int(c.ActiveType[1]) +
		int(c.ActiveType[2]) + int(c.ActiveType[3])
	ratio := func(idx int) float64 {
		if total < 1 {
			return 0
		}
		return float64(c.ActiveType[idx]) / float64(total)
	}
	moveR := ratio(ActiveIdxMove)
	standR := ratio(ActiveIdxStand)
	sitR := ratio(ActiveIdxSit)
	lieR := ratio(ActiveIdxLie)

	flowMag := math.Hypot(float64(c.FlowX), float64(c.FlowY))
	ghostR := c.GhostRatio()

	// AreaUnknown: 常数底
	out[AreaUnknown] = 0.2

	// AreaEnter: 门事件主导
	out[AreaEnter] = clamp01(float64(c.DoorEventCount) / 5)

	// AreaBed: 躺姿主导 + Sleepad 决定性 + 长静止
	bedScore := lieR
	if c.SleepadInBedCount > 0 {
		bedScore += 0.8 // Sleepad 强证据
	}
	bedScore += clamp01(float64(c.LongStillCount)/3) * 0.3
	out[AreaBed] = clamp01(bedScore)

	// AreaSit: 坐姿主导 × 长 Dwell × (沙发签名：LieRetract 高 / FallEvent 低)
	sofaBase := sitR * clamp01(float64(c.DwellEMA)/300)
	sofaBonus := 1 + float64(c.LieRetract)/5
	sofaPenalty := clamp01(1 - float64(c.FallEventCount)/5)
	out[AreaSit] = clamp01(sofaBase * sofaBonus * sofaPenalty)

	// AreaActive: 移动姿态主导 + |Flow| 大 + Dwell 短 + Stand 辅助
	activeScore := moveR +
		clamp01(flowMag/50)*0.3 +
		clamp01(1-float64(c.DwellEMA)/10)*0.2 +
		standR*0.2
	out[AreaActive] = clamp01(activeScore)

	// AreaDeny: 高 GhostRatio + 无 Real + 无姿态累积
	denyScore := ghostR
	if total < 5 && c.RealDecay < 2 {
		denyScore += 0.3
	}
	out[AreaDeny] = clamp01(denyScore)

	return out
}

// updateRisk 按事件累计累计独立的跌倒风险评分
func (c *Cell) updateRisk(g int) {
	r := c.FallEventCount*10 + c.LieRetract*1
	if r > 100 {
		r = 100
	}
	c.Belief[g].RiskScore = r
}

// ========================================================================
// Helper 函数
// ========================================================================

// scaleInt 把 int 乘以浮点系数并四舍五入取整。
func scaleInt(v int, f float64) int {
	if v == 0 {
		return 0
	}
	return int(math.Round(float64(v) * f))
}

// satAddUint8 饱和加法（上限 255）
func satAddUint8(v uint8, add int) uint8 {
	r := int(v) + add
	if r < 0 {
		return 0
	}
	if r > 255 {
		return 255
	}
	return uint8(r)
}

// satAddUint16 饱和加法（上限 65535），用于 ActiveType / TraverseCount。
func satAddUint16(v uint16, add int) uint16 {
	r := int(v) + add
	if r < 0 {
		return 0
	}
	if r > 65535 {
		return 65535
	}
	return uint16(r)
}

// clamp01 限制到 [0, 1]
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
