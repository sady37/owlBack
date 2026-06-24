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

// 值与厂家 declare_area 概念对齐（仅概念，下发走 layout typeName 不读本值）。语义=该区"正常姿态/占用"。
const (
	AreaUnknown    AreaType = iota // 0 未知
	AreaDeny                       // 1 家具不可走（无人占用，留 90min 背板，非豁免）
	AreaBed                        // 2 接触真床（sleepad/vital → 豁免）
	AreaReflector                  // 3 镜/金属静态反射（无真人 → realness 豁免）
	AreaEnter                      // 4 门/进入区
	AreaMonitorBed                 // 5 监护床（同 Bed 行为 + 呼吸心率监测）
	AreaInterfer                   // 6 帘/扇/植物运动干扰（noisy，人可能在 → 90min）
	AreaSit                        // 7 坐区（椅/马桶）
	AreaLying                      // 8 非床躺区（沙发，无 sleepad → 90min）
	AreaActive                     // 9 活动区（走道/淋浴/开阔）
	NumAreaTypes
)

// Source 类型信息来源
type Source uint8

const (
	SourceUnset    Source = 0
	SourceHuman    Source = 1 // FE 刻意画(layout SetPrior)/ interactive-confirmed lying —— 神圣，AI 不覆盖、verified 真摔不擦
	SourceLearned  Source = 2 // 纯自动学(mirror/static-reflector/AutoDeny) —— 可被 Feedback 覆盖、verified 真摔擦
	SourceGeometry Source = 3
	SourceFeedback Source = 4 // alarm feedback / 观测刷新确认的 rest-zone(sit 等) —— 信任 > Learned，但 verified 真摔擦(非 FE 画，可作废)
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

	// ---- sensor_v2 决定 15: EnterTarget 入口类型标注 ----
	// 仅当 cell.AreaType==AreaEnter 时有意义。取值：
	//   ""          → inside_enter（默认）
	//   "outside"   → 通向 unit 外（必须人工标）
	//   "bathroom"  → 通向 bathroom（必须人工标；v2 单 bathroom 无需 id）
	// 由 layout 编辑器写入；自动学习仅产 inside_enter（决定 20，决不写 outside/bathroom）
	EnterTarget string

	// ---- sensor_v2 决定 20: inside_enter 自学习 ----
	// InsideEnterEvidenceN: "track 失锁 + 3s 内同位置 ≤30cm 内重生" 累计次数（仅单 device 内可观测）
	// InsideEnterLearned:   累计 ≥ N (默认 5) 即升格 inside_enter，写 ai.log 给运维审核
	InsideEnterEvidenceN int
	InsideEnterLearned   bool

	// ---- Cell history integral（自适应阈值反馈，参数见本文件 cellFakeAlarmThreshold 等）----
	// FakeAlarmCount: 在该 cell 触发的 fall 报警被人工标 false_alarm 累计次数（"Other" 兜底）
	// ToleratedStillCount: track 在该 cell 长时间 stand-static 但最终自己离开（无 fall 报警）累计次数
	// BlindSpotRecoveryCount: 人在 lost-fall pending 期间从该 cell 重新出现（盲区返回端点）次数
	// 全部衰减由 Decay() 用 EventSec 半衰期处理
	FakeAlarmCount         int
	ToleratedStillCount    int
	BlindSpotRecoveryCount int

	// ---- PR-15.3 Auto-Deny 时间门控 ----
	// AutoDenyQualifiedSinceMs: 该 cell 首次满足 5-cell consensus 软共识（TraverseCount<3 + RealDecay<5）
	// 的时间戳。持续满足 ≥ 15 天才升格 AreaDeny。
	//
	// 设计原理（双阈值滞后 + 类 PR-13 90% 容忍）：
	//   - 5-cell consensus 软共识用 TraverseCount<3：容忍背景 ghost/jitter 噪声（~0.3/天）
	//     避免严格 ==0 导致单次误触即重置
	//   - reset 用 TraverseCount>=5：超出此值视为"真有人走过"，重置 15 天计时
	//   - 中间区 [3,5)：保持当前状态（既不前进也不重置），形成 hysteresis 防抖
	//   - 时间门控 15 天：filter 短期偶发，要求长期持续不被走过才确认 island/furniture
	//
	// 重置（=0）：TraverseCount>=5 OR RealDecay>=5（活动证据）OR Belief 升为非 Unknown/Active
	AutoDenyQualifiedSinceMs int64

	// ---- PR-6 结构化人类反馈（False Alarm Reason / Observed Conditions 多选 checkbox）----
	// GhostCount: false_alarm 反馈勾"NoPerson / Electric / AC Interference" → 该 cell 是 ghost 多发点
	//             birth filter 因子 6 / fall verifier 都参考此值
	// RestZoneConfirmed: false_alarm 勾 "Sit on Chair / Sofa / Wheelchair" → 该 cell 是真人坐姿/卧姿区
	//                    强力学习成 AreaSit / AreaBed；wheelchair 反馈乘 0.3 防 mobile 污染
	// RealFallCount: verified 勾 "Fall / Sitting on the Ground" → 该 cell 是真 fall 高发点
	//                未来可用于反向调整 still-fall 灵敏度
	GhostCount        int
	RestZoneConfirmed int
	RealFallCount     int

	// LearnBlocked: sticky 否决（[[feedback_lying_learning_and_layout_authority]] §1）——
	// verified 真摔时人勾"永不在此自动学跌倒抑制"。置位后该 cell 的自动/反馈抑制学习
	// （MarkRestZoneByFeedback / MarkMirrorBounce / AutoDeny / UpdateBelief 翻抑制类）永久跳过，
	// 跨重启保留（persist）。SourceHuman FE 画不受影响（本就神圣、绕过这些路径）。
	LearnBlocked bool

	// MirrorBounceCount: L1 mirror pair 检测命中后在 bounce point 周围 2×2 微块累加。
	// ≥ MirrorPromoteThreshold (3) 即升 Belief[0]=AreaDeny+SourceLearned（人工 SourceHuman 不覆盖）。
	// LastMirrorMs: 最近一次累加时戳，Decay 用。
	MirrorBounceCount int
	LastMirrorMs      int64

	// StaticReflectorCount: 静止金属反射体自学习（运动镜像之外的另一类）——一条 track 长期钉死在
	// 此 cell（金属把手/淋浴架直射回波）+ 同时另一真人 track 在别处游走（corroboration）时累加。
	// 跨多次独立 episode ≥ StaticReflectorPromoteThreshold 升 AreaDeny（Phase B；Phase A 仅累计+log）。
	// 这类点人工标不了（幻影坐标依赖雷达角度/材质/多径，非肉眼可预知），只能自动学。
	StaticReflectorCount  int
	LastStaticReflectorMs int64

	// SitScore: AreaSit 4 通道 log-odds 累积（sit_learning.go；每条自走长坐 episode 加 sitEpisodeLLR）。
	// ≥ sitPromoteTau 升 AreaSit(SourceLearned)；HL DecayParams.SitScoreSec(默认 4d,config 可调) 指数衰减（隔离 episode 自然褪）。
	SitScore float64

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
// 由 owlBack/tools/Xsensorv1/internal/config 从 yaml 加载并传给 engine，engine 在 decayLoop / playback 中调用 DecayAll(dtSec, p)。
type DecayParams struct {
	ImmediateSec float64 // Real/Ghost/Flow/DwellEMA/ActiveType[Stand]
	WalkSec      float64 // ActiveType[Move]
	SitSec       float64 // ActiveType[Sit]
	LieSec       float64 // ActiveType[Lie]
	EventSec     float64 // Traverse/Fall/Retract/Sleepad/Door/LongStill/LieAnomaly
	SitScoreSec  float64 // AreaSit 4 通道 log-odds SitScore（默认 4d；孤立 episode 自然褪，需近日 2-3 次累过 τ）

	// PR-11: Belief[0].Confidence 按 AreaType 分档半衰期（秒）。索引 = AreaType 值。
	// 设计：T_half = days_to_10 / log2(10) ≈ days_to_10 / 3.32
	// 物理稳定性 → 衰退快慢：床/卫浴稳定，沙发/走道相对动态。
	// 衰到 < 10 触发降级 AreaUnknown + Source=Unset（cell.Decay 内）。
	// 0 = 该类型不衰减（保留给特殊用途）。
	BeliefHalfLifeByType [NumAreaTypes]float64
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
		SitScoreSec:  4 * 24 * 3600, // SitScore HL 4d

		// PR-11: 按 AreaType 分档 Belief 衰退（半衰期，秒）
		//   AreaSit (沙发/椅子)：    7 天衰到 10  HL=2.1d  — 可移动家具
		//   AreaBed (床/lying)：     9 天衰到 10  HL=2.71d — 防周末移床方向；与 sleep 4h+ 刷新机制配合
		//   AreaToilet/Shower：      60 天衰到 10 HL=18d   — 固定家具/管道连接，几乎不变
		//   AreaActive (walk path)：  14 天衰到 10 HL=4.2d  — 走道路径稳定，但易被新家具污染
		//   AreaEnter (门洞)：        30 天衰到 10 HL=9d    — 几何固定但 layout 可能改
		//   AreaDeny (墙/家具/镜子)： 30 天衰到 10 HL=9d    — 同上
		//   AreaUnknown：             0（不衰退；本来就是 0）
		BeliefHalfLifeByType: [NumAreaTypes]float64{
			AreaUnknown:    0,
			AreaEnter:      9 * 24 * 3600,    // 30天
			AreaBed:        2.71 * 24 * 3600, // 9天
			AreaMonitorBed: 2.71 * 24 * 3600, // 9天（同 Bed）
			AreaSit:        2.1 * 24 * 3600,  // 7天
			AreaLying:      2.1 * 24 * 3600,  // 7天（沙发可移动，同 Sit）
			AreaActive:     4.2 * 24 * 3600,  // 14天
			AreaDeny:       9 * 24 * 3600,    // 30天（家具固定）
			AreaReflector:  9 * 24 * 3600,    // 30天（镜/金属固定）
			AreaInterfer:   4.2 * 24 * 3600,  // 14天（帘/扇可移动）
		},
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

// IsRestZone 是否为「可长时间停留」的休息区。仅 Bed / Sit。
//
// 语义（用户 2026-04-29 对齐）：
//   - Bed / Sit (沙发/椅子)：人可长时间坐/躺；still-fall / silent-fall 都应排除
//   - Toilet / Shower：**人不应久留**（>15-20min 异常）；属于 still-fall 触发场景，不在此集合
//
// 旧版本曾包含 AreaToilet，是 bug：toilet 上消失/久留应触发告警，不应被
// "rest zone exempt" 跳过。
func (c *Cell) IsRestZone() bool {
	t := c.Belief[0].Type
	return t == AreaBed || t == AreaMonitorBed || t == AreaSit || t == AreaLying
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

// still-fall-area 分类阈（cell-learning 用）：按 cell areaType 分的"久静多久算 fall-relevant"。
// **这不是 fall 决策阈**——DBN silent-fall 阈权威 = belief/floor.go（契约 6A 其十五）。本组只供
// cell-learning 判"哪些区不学 AreaSit（toilet/shower/deny）"(track_manager stillFallTimeoutSec) +
// LongStill grid 标记(EffectiveStillTimeoutSec)。
const (
	stillAreaToiletShowerSec = 15 * 60 // AreaToilet/AreaShower
	stillAreaDenySec         = 5 * 60  // AreaDeny
	stillAreaDefaultSec      = 8 * 60  // 其它（Enter/Active/Unknown）
	stillAreaNonRiskFactor   = 1.2     // 非风险时段放宽因子
)

// cell history 自适应容忍：反复假报/长静自然离开 → 放宽 EffectiveStillTimeoutSec，上限 MaxTolerance。
const (
	cellFakeAlarmThreshold      = 3
	cellToleratedStillThreshold = 5
	cellMaxToleranceFactor      = 2.0
)

// StillTimeoutSec 静止超时阈值（秒；0 = 不限）。默认读 Belief[0]。risk-time=基线（严格）；
// non-risk-time=基线 × stillAreaNonRiskFactor（宽松）。时段见 math_util.go::IsNightTime。
func (c *Cell) StillTimeoutSec(isRiskTime bool) int {
	base := c.stillTimeoutBase()
	if base == 0 {
		return 0
	}
	if !isRiskTime {
		return int(float64(base) * stillAreaNonRiskFactor)
	}
	return base
}

// stillTimeoutBase risk-time 基线值（严格档）。
// 床/沙发：不限（休息合理）；马桶/淋浴：15min；Deny: 5min；其它：8min。
func (c *Cell) stillTimeoutBase() int {
	switch c.Belief[0].Type {
	case AreaBed, AreaMonitorBed, AreaSit, AreaLying:
		return 0
	case AreaDeny, AreaInterfer:
		return stillAreaDenySec
	}
	return stillAreaDefaultSec
}

// EffectiveStillTimeoutSec 综合 cell history 的最终静止超时阈值（秒）。
//
// = StillTimeoutSec(isRiskTime) × toleranceFactor()
//
// toleranceFactor 由该 cell 累计的 FakeAlarmCount + ToleratedStillCount 决定：
//   - 全新装机（无历史）→ factor = 1.0（基线，最严）
//   - 反复假报后 → 自动放宽，最多至 MaxToleranceFactor（默认 2.0）
//
// 设计目的：让系统自动消化「这个雷达在这个位置就是会假报」的客观存在，不靠人工调阈值。
func (c *Cell) EffectiveStillTimeoutSec(isRiskTime bool) int {
	base := c.StillTimeoutSec(isRiskTime)
	if base == 0 {
		return 0
	}
	return int(float64(base) * c.toleranceFactor())
}

// toleranceFactor 自适应放宽因子 ∈ [1.0, MaxToleranceFactor]。
//
// 累计证据 ratio = (FakeAlarmCount + ToleratedStillCount) / (FakeAlarmThreshold + ToleratedStillThreshold)
// 线性饱和：ratio=0 → factor=1.0；ratio>=1 → factor=MaxToleranceFactor。
func (c *Cell) toleranceFactor() float64 {
	thresholdSum := cellFakeAlarmThreshold + cellToleratedStillThreshold
	score := float64(c.FakeAlarmCount + c.ToleratedStillCount)
	ratio := score / float64(thresholdSum)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return 1.0 + (cellMaxToleranceFactor-1.0)*ratio
}

// IncrFakeAlarm 人工标记 false_alarm 反馈到该 cell（在告警触发位置调用）。
// 累计后衰减由 Decay() 用 EventSec 半衰期处理。
func (c *Cell) IncrFakeAlarm() {
	c.FakeAlarmCount++
}

// IncrToleratedStill 该 cell 上长时间 stand-static 未报警 + track 自然离开 反馈（容忍证据）。
// 典型场景：水池/灶台/化妆台/熨衣板等"人本来就站着不动办事"的位置。
func (c *Cell) IncrToleratedStill() {
	c.ToleratedStillCount++
}

// IncrBlindSpotRecovery 人从 lost-fall pending 期间在该 cell 重新出现（盲区返回端点）。
// 该 cell 学到自己是「人能不能突然出现」的语义；未来 track 在此 cell 出生时，
// birth-score 加成（避免重复误判 ghost）。
func (c *Cell) IncrBlindSpotRecovery() {
	c.BlindSpotRecoveryCount++
}

// IncrGhostCount 人工标记 "NoPerson / Electric / AC Interference"（false_alarm 反馈第 1 项）。
// 强 ghost 证据；累计后影响 birth filter 因子 6 (cell.GhostRatio) 与 fall verifier。
func (c *Cell) IncrGhostCount() {
	c.GhostCount++
}

// IncrRestZoneConfirmed 人工标记 "Sit on Chair / Sofa / Wheelchair"（false_alarm 反馈第 2-4 项）。
// 三类统一 +1，wheelchair mobile 污染依靠 Decay() 自然过滤
// （wheelchair 各位置累积低 → 衰减期内不到学习阈值 → 不会误学成 AreaSit）。
func (c *Cell) IncrRestZoneConfirmed() {
	c.RestZoneConfirmed++
}

// IncrRealFallCount 人工标记 "Fall / Sitting on the Ground"（verified 反馈核心 ground truth）。
// 累计该 cell 真 fall 次数；未来用于反向调整 still-fall 灵敏度（真 fall 高发区收紧阈值）。
func (c *Cell) IncrRealFallCount() {
	c.RealFallCount++
}

// MarkRestZoneByFeedback PR-7.1 / PR-9.2 / PR-11：依据反馈/持续观测即时锁定 cell.AreaType。
// target ∈ {AreaSit, AreaBed, AreaToilet, AreaShower}：
//
//	AreaSit     — Chair / Wheelchair（橙色坐姿）
//	AreaBed     — Sofa / Lounge chair（蓝色躺姿）+ lying ≥4h on bed 持续观测刷新
//	AreaToilet  — sit ≥5min on toilet 持续观测刷新
//	AreaShower  — 待加（暂同 toilet）
//
// Confidence=95, Source=SourceFeedback（非 FE 画，verified 真摔可擦；FE 画的走 SetPrior+SourceHuman 神圣）。
//
// 保护规则（不覆盖；防止降级）：
//   - AreaDeny / AreaEnter  layout 锁定（墙/门洞）
//   - AreaBed 不被 AreaSit 降级（lying 信息更精确）
//   - AreaToilet/Shower 不被 AreaBed/Sit 降级（卫浴更具体）
//   - 已是 SourceHuman（FE 画）不被 feedback 覆盖
//   - 已是 SourceFeedback + 同 target  幂等不 reset
func (c *Cell) MarkRestZoneByFeedback(target AreaType) bool {
	cur := c.Belief[0].Type
	// sticky 否决：永不在此自动/反馈学抑制区
	if c.LearnBlocked {
		return false
	}
	// 不覆盖 layout 物理锁定（墙/门）
	if cur == AreaDeny || cur == AreaEnter {
		return false
	}
	// AreaBed/MonitorBed 不被 AreaSit 降级（lying 信息更精确）
	if (cur == AreaBed || cur == AreaMonitorBed) && target == AreaSit {
		return false
	}
	// FE 画的（SourceHuman）神圣，feedback 不覆盖
	if c.Belief[0].Source == SourceHuman {
		return false
	}
	// 幂等：已是 SourceFeedback + 同 target
	if cur == target && c.Belief[0].Source == SourceFeedback {
		return false
	}
	c.Belief[0] = BeliefState{Type: target, Confidence: 95, Source: SourceFeedback}
	c.AreaType = target
	return true
}

// ClearNonHumanLearnedZone verified 真摔时擦除该 cell 的**非 FE 画**(SourceFeedback/Learned/Geometry)
// rest/deny 分类 → 回 Unknown（相当于没学过）。真摔证明该处的抑制/deny 分类是错的（有真人在此倒地）。
// FE 画的（SourceHuman）不擦——由调用方 IncrRealFallCount 记录，反复真摔触发人工复审（[[feedback_lying_learning_and_layout_authority]] §3.1）。
func (c *Cell) ClearNonHumanLearnedZone() bool {
	b := c.Belief[0]
	if b.Source == SourceHuman || b.Source == SourceUnset {
		return false
	}
	switch b.Type {
	case AreaBed, AreaMonitorBed, AreaSit, AreaLying, AreaDeny, AreaInterfer:
	default:
		return false
	}
	c.Belief[0] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
	c.Belief[1] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
	c.Belief[2] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
	c.AreaType = AreaUnknown
	return true
}

// MarkLearnBlocked sticky 否决（§1）：置 LearnBlocked，永久阻止该 cell 自动/反馈学习跌倒抑制区。
// 由 feedback verified 分支在人勾 sticky veto 时调（配合 ClearNonHumanLearnedZone 先擦后封）。
// 返回是否新置位（已置位返回 false，便于日志区分）。
func (c *Cell) MarkLearnBlocked() bool {
	if c.LearnBlocked {
		return false
	}
	c.LearnBlocked = true
	return true
}

// isSuppressiveArea 该 AreaType 是否会抑制/拒绝跌倒报警（rest/卫浴/deny）。sticky 否决用。
func isSuppressiveArea(t AreaType) bool {
	switch t {
	case AreaBed, AreaMonitorBed, AreaSit, AreaLying, AreaDeny, AreaInterfer, AreaReflector:
		return true
	}
	return false
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

	// Cell history integral：用 EventSec 半衰期（与设计 30 天一致时配置 EventSec=30d）。
	// 当前先用 EventSec，后续若需精细化再加独立 HistorySec 衰减档。
	c.FakeAlarmCount = scaleInt(c.FakeAlarmCount, fEv)
	c.ToleratedStillCount = scaleInt(c.ToleratedStillCount, fEv)
	c.BlindSpotRecoveryCount = scaleInt(c.BlindSpotRecoveryCount, fEv)

	// PR-6 结构化反馈：同 EventSec 半衰期
	c.GhostCount = scaleInt(c.GhostCount, fEv)
	c.RestZoneConfirmed = scaleInt(c.RestZoneConfirmed, fEv)
	c.RealFallCount = scaleInt(c.RealFallCount, fEv)

	// L1 mirror bounce 同 EventSec 半衰期；持续累不到 promote 阈值的 cell 自然衰退归零，
	// 避免一次性误报永久污染 grid。
	c.MirrorBounceCount = scaleInt(c.MirrorBounceCount, fEv)
	// 静止反射体同半衰期：偶发一次静止伪迹自然衰退，需跨多次独立 episode 才累到 promote。
	c.StaticReflectorCount = scaleInt(c.StaticReflectorCount, fEv)

	// SitScore（log-odds）按 SitScoreSec（默认 4d，config 可调）指数衰减：孤立 episode 自然褪，需 2–3 次近日自走长坐才够 τ。
	c.SitScore *= factor(dtSec, p.SitScoreSec)

	// PR-11: Belief[0].Confidence 按 AreaType 分档衰减（半衰期 BeliefHalfLifeByType[type]）
	// 衰减后 Confidence < 10 → 降级 AreaUnknown + Source=Unset（不再贡献 IsRestZone 等判定）。
	// 仅衰减 [0]（最强信念）；[1][2] 是历史快照，保持原样。
	// 半衰期 0 时该类型不衰减（如 AreaUnknown 已是 0）。
	if c.Belief[0].Confidence > 0 {
		hl := p.BeliefHalfLifeByType[c.Belief[0].Type]
		if hl > 0 {
			newConf := scaleInt(c.Belief[0].Confidence, factor(dtSec, hl))
			if newConf < 10 {
				c.Belief[0] = BeliefState{Type: AreaUnknown, Confidence: 0, Source: SourceUnset}
				c.AreaType = AreaUnknown
			} else {
				c.Belief[0].Confidence = newConf
			}
		}
	}
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
		// sticky 否决：vetoed cell 不自动翻成抑制类（防 verified 真摔擦回 Unknown 后又自学回 bed/deny 重新抑制）
		if c.LearnBlocked && isSuppressiveArea(bestType) {
			break
		}
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
