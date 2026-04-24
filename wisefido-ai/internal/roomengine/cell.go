package roomengine

import "math"

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
	AreaSit                     // 3 允许坐姿（沙发/椅子/马桶）
	AreaActive                  // 4 活动区（走廊、开放空间）
	AreaDeny                    // 5 禁止 track（墙/家具/Metal，track 出现 = Fake）
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

// CorePose 核心姿态（从雷达 13 种 pose 还原到 4 种）
type CorePose uint8

const (
	CorePoseUnknown CorePose = 0
	CorePoseMove    CorePose = 1
	CorePoseStand   CorePose = 2
	CorePoseSit     CorePose = 3
	CorePoseLie     CorePose = 4
)

// Pose 常量（与 owl-common/observation/fields.go EnumPose 对齐）
const (
	PoseInit            = 0
	PoseWalking         = 1
	PoseSuspectedFall   = 2
	PoseSitting         = 3
	PoseStanding        = 4
	PoseFall            = 5
	PoseLying           = 6
	PoseSuspectedFloor  = 7
	PoseSittingOnGround = 8
	PoseBedSitUp        = 9
	PoseSuspectedBedUp  = 10
	PoseConfirmedBedUp  = 11
	PoseRunning         = 12
	NumPoses            = 13
)

// RadarPoseToCore 把雷达 13 种 pose 还原到 4 种核心姿态。
// 核心哲学：不信雷达的 Suspected/Confirmed 分类和 area 派生事件——
// 疑似与确认归为同一核心姿态，由我们自己做时间过滤和事件检测。
func RadarPoseToCore(pose int) CorePose {
	switch pose {
	case PoseWalking, PoseRunning:
		return CorePoseMove
	case PoseStanding:
		return CorePoseStand
	case PoseSitting, PoseSuspectedFloor, PoseSittingOnGround,
		PoseBedSitUp, PoseSuspectedBedUp, PoseConfirmedBedUp:
		return CorePoseSit
	case PoseLying, PoseSuspectedFall, PoseFall:
		return CorePoseLie
	}
	return CorePoseUnknown
}

// ActiveType 索引常量
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
	ActiveType [4]uint8 // [Move, Stand, Sit, Lie] 0-255 累计分（可视化作 RGBA 通道）
	AreaType   AreaType // 推断属性，与 Belief[0].Type 镜像（query 加速）

	// ---- 访问可信度（track 层按本帧 quality 分桶喂）----
	RealDecay  int
	GhostDecay int

	// ---- 流动与停留（track 瞬时 → cell 累计）----
	FlowX, FlowY   int // 平均速度 EMA（cm/s）
	DwellEMA       int // 单次停留 EMA（秒）
	LongStillCount int // 静止 > 15min 次数

	// ---- 事件累计（track 层触发）----
	FallEventCount int // prevCore=Stand/Move → Lie + Z 骤降
	LieRetract     int // Stand/Move → Lie → Stand/Move 短暂回撤（沙发签名）

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

// 衰减常量（秒）。Decay 每小时调一次。
const (
	HalfLifeShort    = 24 * 3600     // 24h：普通证据（访问/姿态/流动/停留）
	HalfLifeLong     = 7 * 24 * 3600 // 7d：事件类（Fall/Retract/Sleepad/Door/LongStill）
	eventHalfLifeMul = 7             // 长档 = 短档 × 7
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

// IsRestZone 是否为合理静止区（Bed 或 Sit）。默认读 Belief[0]。
func (c *Cell) IsRestZone() bool {
	t := c.Belief[0].Type
	return t == AreaBed || t == AreaSit
}

// StillTimeoutSec 静止超时阈值（秒；0 = 不限）。默认读 Belief[0]。
func (c *Cell) StillTimeoutSec() int {
	switch c.Belief[0].Type {
	case AreaBed, AreaSit:
		return 0 // 床/沙发/椅子不限
	case AreaDeny:
		return 5 * 60 // Deny 区静止 5 min 即可疑
	}
	return 8 * 60 // Enter/Active/Unknown: 8 min
}

// ========================================================================
// Decay（时间窗口衰减）
// ========================================================================

// Decay 按半衰期衰减累计字段。普通证据用 halfLifeSec；事件用 × eventHalfLifeMul。
// 物理烤入（InRoom/InFOV/EdgeDist/MaxZ/MinZ）和 Belief 不衰减。
func (c *Cell) Decay(dtSec, halfLifeSec float64) {
	if dtSec <= 0 || halfLifeSec <= 0 {
		return
	}
	f := math.Pow(0.5, dtSec/halfLifeSec)
	fEv := math.Pow(0.5, dtSec/(halfLifeSec*eventHalfLifeMul))

	// 短档（普通证据）
	c.RealDecay = scaleInt(c.RealDecay, f)
	c.GhostDecay = scaleInt(c.GhostDecay, f)
	c.FlowX = scaleInt(c.FlowX, f)
	c.FlowY = scaleInt(c.FlowY, f)
	c.DwellEMA = scaleInt(c.DwellEMA, f)
	for i := range c.ActiveType {
		c.ActiveType[i] = uint8(scaleInt(int(c.ActiveType[i]), f))
	}

	// 长档（事件类）
	c.LongStillCount = scaleInt(c.LongStillCount, fEv)
	c.FallEventCount = scaleInt(c.FallEventCount, fEv)
	c.LieRetract = scaleInt(c.LieRetract, fEv)
	c.SleepadInBedCount = scaleInt(c.SleepadInBedCount, fEv)
	c.SleepadLeftBedCount = scaleInt(c.SleepadLeftBedCount, fEv)
	c.DoorEventCount = scaleInt(c.DoorEventCount, fEv)
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
	activeSum := int(c.ActiveType[0]) + int(c.ActiveType[1]) +
		int(c.ActiveType[2]) + int(c.ActiveType[3])
	if total < 3 && eventsSum < 1 && activeSum < 10 {
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

// satAddUint8 饱和加法（上限 255），用于 ActiveType 累加。
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
