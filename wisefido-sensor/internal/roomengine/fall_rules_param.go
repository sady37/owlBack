package roomengine

// fall_rules_param.go
//
// 跌倒检测三类规则（still / silent / lost）+ cell history integral 的统一调参常量。
//
// 设计原则（用户对齐 2026-04-27）：
//   1. 集中：所有跌倒相关阈值 / 时长 / 因子放在 FallRulesParam，禁止散落 magic number
//   2. 编译期：Go const 风格的不可变结构（var + 大写命名 + 约定不修改）
//   3. risk-time / non-risk-time 命名（不是 day/night）：
//        - risk-time = 高风险时段（默认夜间 23:30-07:30）→ 严（短阈值，快报警）
//        - non-risk-time = 非风险时段（默认白天）→ 宽（长阈值，少误报）
//      逻辑：高风险时段跌倒更易发、人手少 → 更快响应
//   4. 修改方式：改本文件 → 重编 → 部署。属于「设计参数」不是「部署配置」。

// stillFallParam Still Fall（一直可见 + stand 静止）参数
type stillFallParam struct {
	// 基线时长（risk-time，按 cell areaType 分）
	ToiletShowerSec int // AreaToilet/AreaShower 基线 15min
	DenyZoneSec     int // AreaDeny 基线 5min
	DefaultSec      int // 其它 (Enter/Active/Unknown) 基线 8min

	// 非风险时段放宽因子（risk-time × 此因子 = non-risk-time 时长）
	// 默认 1.2 → 15min × 1.2 = 18min
	NonRiskTimeFactor float64
}

// lostFallParam Lost Fall（突然消失 + 屋内任意处）参数
// lostFallParam:gate-list lost-fall pending 退役后只剩 DBN shadow / track 生命周期 / layout / track-birth
// 仍消费的 5 个常量(其余 8 个 wait/speed/fallback 字段随 lostFallWaitMs+pending 创建删除一并退役)。
type lostFallParam struct {
	// 离最近出口最小距离（cm）—— ≤此值视为「贴近门口」。layout_parser 隐式入口推导消费。
	ExitDistMinCm int

	// StillBox（静止无移动）检测：失锁前 30s 滚动窗口 per-axis box(max-min) <= 此值视为 still box。
	// track 生命周期 still-box 检测消费（updateContinuousIndicators）。
	StillBoxCm int

	// 走动前置：DBN shadow lost-track 路径只认"走动中突然消失"——消失前 still-box ≥ 此值=静止态,不进 lost-fall。
	MovingPreconditionMs int

	// Birth verdict 多流时序兜底：track 帧与 EnterRoom 分两条流,birth 时仅打初步分留 Pending,此 ms 后 deadline 重算。
	BirthFinalGraceMs int

	// 距离闸（cm，平面/俯视距）：丢轨点距雷达 > 此值 → 贴地弱回波盲区,DBN shadow no-detect 完全中性化（lostFarFromRadar）。
	// 0 = 关闸。依据 doc/AI_fall_detect.md §3.7。
	DistanceGateCm int
}

// cellHistoryParam Cell History Integral（自适应阈值）参数
type cellHistoryParam struct {
	// 触发阈值：FakeAlarmCount + ToleratedStillCount 达到此值时 tolerance_factor = MaxToleranceFactor
	FakeAlarmThreshold      int // 人工标 fake 计数门槛
	ToleratedStillThreshold int // 自然观察长时静止计数门槛

	// 阈值放宽上限：cell.EffectiveStillTimeoutSec ≤ base × MaxToleranceFactor
	MaxToleranceFactor float64

	// 时间衰减半衰期（天）—— 历史包袱不无限累计
	DecayHalfLifeDays int
}

// fallRulesParam 顶层（不导出，避免外部直接构造；用 FallRulesParam 单例）
// neighborParam 跨房 hand-off 抑制（仅二义 lost-fall）参数。铁律（非全空间监控）：只有本房丢轨与邻房
// 出现两事件**极近 + 有向**才能排除本房跌倒；stale/durable「上次在哪」不可用（人可能穿盲区真摔）。
type neighborParam struct {
	// HandoffWindowMs 邻房 hand-off 事件相对本房丢轨（st.lastSeenMs）的最大**滞后**（人先走后到）。
	// 固定绝对阈，**不随房间距离/邻接伸缩**（穿行越久=穿过盲区越多=摔机会越大，放宽=最该警惕时松手）。
	HandoffWindowMs int64
	// JitterMs 传感器上报时序抖动的反向余量（允许邻房事件略早于本房丢轨此值内，非真"先到后走"）。
	JitterMs int64
}

type fallRulesParam struct {
	Still       stillFallParam
	Lost        lostFallParam
	Neighbor    neighborParam
	CellHistory cellHistoryParam
}

// InsideEnterLearnThreshold sensor_v2 决定 20：inside_enter 自学习升格门槛。
// 累计 "track 失锁 + 3s 内同 cell 重生" 事件 ≥ 此值 → InsideEnterLearned=true。
// 默认 5，prod 校准。
const InsideEnterLearnThreshold = 5

// FallRulesParam 全局单例。所有跌倒规则统一引用此变量。
//
// 约定：只读。修改值须改本文件并重编。运行时不修改字段。
var FallRulesParam = fallRulesParam{
	Still: stillFallParam{
		ToiletShowerSec:   15 * 60, // 15 min
		DenyZoneSec:       5 * 60,  // 5 min
		DefaultSec:        8 * 60,  // 8 min
		NonRiskTimeFactor: 1.2,     // non-risk × 1.2 → 18 min
	},
	Lost: lostFallParam{
		ExitDistMinCm:        30,     // 30cm（贴近门口）；layout 隐式入口推导
		StillBoxCm:           50,     // 30s 滚动窗 per-axis box <= 50cm 视为 still（实测倒地质心抖 50×40cm）
		MovingPreconditionMs: 60_000, // 消失前 still-box ≥60s = 静止态 → 不进 lost-fall
		BirthFinalGraceMs:    2000,   // birth 终判延迟 2s
		DistanceGateCm:       500,    // 丢轨点距雷达 >500cm（d_fall）→ 贴地弱回波盲区中性化
	},
	Neighbor: neighborParam{
		HandoffWindowMs: 60_000, // 默认 60s（30-60 适宜；容下"先走后到"先后差）；超窗→中间可能盲区真摔→不排除
		JitterMs:        5_000,  // 5s：仅留传感器时序抖动反向余量
	},
	CellHistory: cellHistoryParam{
		FakeAlarmThreshold:      3,
		ToleratedStillThreshold: 5,
		MaxToleranceFactor:      2.0,
		DecayHalfLifeDays:       30,
	},
}
