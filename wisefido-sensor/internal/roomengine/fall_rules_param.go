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

// silentFallParam Silent Fall（sleepad+radar 融合：床上方遮挡 / 离床后矛盾）参数
//
// 触发链（用户设计 2026-04-27）：
//  1. sleepad InBed 时同房间 radar 也 InBed → 床区获得证据（多人为多 sleepad 累计）
//  2. 持续在床 ≥ MinInBedSec（默认 5min）才作为 precondition（过滤短暂坐床）
//  3. sleepad LeftBed 触发后开始等待
//  4. 等待时长按 HR/RR 与人数：
//     - LeftBed 时 sleepad 有 HR/RR + 单人 → WaitVitalSec=60s（信心高）
//     - 否则 → WaitNoVitalSec=120s（保守）
//  5. 等待期满时 radar 仍在 Bed 邻域 → silent fall（人离床但 radar 还看到 bed → 跌到地上）
type silentFallParam struct {
	MinInBedSec     int // 在床持续时间下限，作为 precondition
	WaitNoVitalSec  int // sleepad LeftBed 时无 HR/RR：等待 120s
	WaitVitalSec    int // sleepad LeftBed 时有 HR/RR + 单人：等待 60s
	BedNeighborhood int // 视为「仍在 Bed 邻域」的距离 cm（与 BedsideMarginCm 同语义，独立可调）
}

// stillFallParam Still Fall（一直可见 + stand 静止）参数
type stillFallParam struct {
	// 基线时长（risk-time，按 cell areaType 分）
	ToiletShowerSec int // AreaToilet/AreaShower 基线 15min
	DenyZoneSec     int // AreaDeny 基线 5min
	DefaultSec      int // 其它 (Enter/Active/Unknown) 基线 8min

	// 非风险时段放宽因子（risk-time × 此因子 = non-risk-time 时长）
	// 默认 1.2 → 15min × 1.2 = 18min
	NonRiskTimeFactor float64

	// 静止判定：|Δpos| < StaticPosCm 视为静止；超过即重置计时
	StaticPosCm int
}

// lostFallParam Lost Fall（突然消失 + 屋内任意处）参数
type lostFallParam struct {
	// 等待时长按消失点 cell areaType 分
	RestZoneWaitSec int // AreaBed / AreaSit 60min（睡觉/坐着丢 track 常见）
	DenyZoneWaitSec int // AreaDeny 5min
	WalkwayWaitSec  int // 其它（Active / Unknown / Enter）5min
	// AreaToilet / AreaShower 不在此表—与 still fall 同时长（运行时取）

	// 离最近出口最小距离（cm）—— ≤此值视为「贴近门口」可能正常走出，>此值未配
	// ExitRoom 事件全部进 lost-fall pending 池。
	// 2026-04-30 收紧：原 100cm（1 米/约 1 秒走完）→ 30cm（必须真贴在门口）。
	// 收紧动机：elder care 宁可误报不可漏报；门口 30cm-1m 区间的真摔倒原先被
	// 当作"可能正常出门"漏掉，30cm 以内才视为正常通过门框的合理误差。
	ExitDistMinCm int

	// 「空间跳跃」加权因子：表现过空间跳跃的 track 等待时间 × 此因子
	// 0.5 = 跳过的等待时间砍半（更敏感）
	SpatialJumpFactor float64

	// StillBox（静止无移动）检测（box 判据，2026-05-03 由 byte-equal 改为 box）：
	//   StillBoxCm：失锁前 30s 滚动窗口内的 (max-min) 位移 <= 此值视为 still box
	//   原 byte-equal "连续 25 帧字面相同"判据被 firmware X/Y 抖动 + 摔倒抽搐打破，
	//   改 box 容差更鲁棒。判据见 updateContinuousIndicators。
	StillBoxCm int

	// 走动前置（2026-06-01）：lost-fall 只认"走动中突然消失"（走动→跌倒+遮挡→丢 track）。
	// 消失前若已 settle 进长静止（still-box 持续 ≥ 此值）= 站洗手台/坐马桶/卧床不动 = 正常静止态，
	// 归 Still-fall 域（长阈值），**不进 lost-fall**。判据用 box-based still-box（抗坐姿 jitter，
	// 路径长/移动时长会被 jitter 骗）。治本：CABB/MoM 多条冻结站/坐 lost_track FP（无真跌倒）。
	MovingPreconditionMs int

	// 速度二档判定（cm/s）：
	// > ImpossibleSpeedCm 硬 ghost（无 EnterRoom 也判定）：老人最快 100-150cm/s
	// > SuspectSpeedCm + 无 EnterRoom 软 ghost（probation）：成人快走可达
	// < SuspectSpeedCm 默认真人：老人 60-80cm/s 步速正常活动
	ImpossibleSpeedCm int
	SuspectSpeedCm    int

	// Birth verdict 多流时序兜底：track 帧与 EnterRoom 事件分两条流，
	// 出生瞬时 EnterRoom 可能还在路上 → birth 时仅打初步分留 Pending；
	// 此 ms 后到达 deadline 重算（window 扩展到 [T-3s, deadline]）
	BirthFinalGraceMs int

	// Lost-fall pending 等待时间下限（兜底）
	EffectiveWaitFloorSec int

	// NumberPeople=0 ExitRoom 兜底窗口（ms）：
	// 部分 firmware（如 D523）在 FOV 边角离场不发 ExitRoom，但会发 number_people=0。
	// track 终止入 pendingLostFall 池前，若 number_people=0 在 ts.LastObservedMs ±此值
	// 内到达 → 视作 ExitRoom 兜底，跳过 pending。**前提：track 未进入 frozen 状态**
	// （frozen 与 number_people=0 在 firmware 状态机里互斥；frozen 状态下 firmware
	// 仍认为屋内有人，绝不会发 number_people=0。若两者同现说明 firmware 异常或
	// 是 firmware-frozen 残影结束后的新状态——后者属于 CD2B 类盲区返回 case，
	// 应正常入 lost_fall pending 池等待 recovery 取消，不能被 number_people=0 抑制）。
	// 默认 60000ms：与用户设计语义对齐——"t1+60s 前收到 number_people=0 视为正常离房"。
	NumberPeopleZeroFallbackMs int64

	// 距离闸（cm，平面/俯视距）：丢轨点距雷达 > 此值则抑制 lost_fall。
	// 跌倒检测半径 d_fall（贴地弱回波，~500cm）远小于人员检测半径（~700cm）——
	// 超出 d_fall 处即便真摔，firmware 也读不出地面态、发不出 pose=5，lost_track 退化为
	// 结构性盲猜（治 D523 边缘 5.8m 丢轨误报）。平面距 = √(RawH²+RawV²)（雷达在本地系原点）。
	// 依据见 doc/AI_fall_detect.md §3.7。0 = 关闸。
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
type fallRulesParam struct {
	Silent      silentFallParam
	Still       stillFallParam
	Lost        lostFallParam
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
	Silent: silentFallParam{
		MinInBedSec:     5 * 60, // 在床 ≥5min 才作 precondition（过滤短坐）
		WaitNoVitalSec:  120,    // 无 HR/RR：保守 120s
		WaitVitalSec:    60,     // 有 HR/RR + 单人：60s
		BedNeighborhood: 100,    // ≤1m 视为 Bed 邻域
	},
	Still: stillFallParam{
		ToiletShowerSec:   15 * 60, // 15 min
		DenyZoneSec:       5 * 60,  // 5 min
		DefaultSec:        8 * 60,  // 8 min
		NonRiskTimeFactor: 1.2,     // non-risk × 1.2 → 18 min
		StaticPosCm:       20,
	},
	Lost: lostFallParam{
		RestZoneWaitSec:            60 * 60, // 60 min
		DenyZoneWaitSec:            5 * 60,
		WalkwayWaitSec:             5 * 60,
		ExitDistMinCm:              30, // 30cm（贴近门口）；2026-04-30 从 100cm 收紧
		SpatialJumpFactor:          0.5,
		StillBoxCm:                 50,    // 30s 滚动窗 per-axis box(dx,dy 各)<= 50cm 视为 still（50×50 方框）
		//                                  2026-06-06 30→50 + 判据由对角线改 per-axis(见 BoxRangeWithinMs)：
		//                                  倒地质心抖动实测 50×40cm，旧 30/对角线把真摔躺着误判成"动"，still 累计不起来。
		MovingPreconditionMs:       60_000, // 消失前 still-box ≥60s = 静止态 → 不进 lost-fall（走 Still-fall）
		ImpossibleSpeedCm:          200,   // 硬 ghost：老人最快 100-150cm/s
		SuspectSpeedCm:             100,   // 软 ghost：需 EnterRoom 反证
		BirthFinalGraceMs:          2000,  // birth 终判延迟 2s
		EffectiveWaitFloorSec:      60,    // 兜底最少等 60s
		NumberPeopleZeroFallbackMs: 60000, // ExitRoom 缺失时 number_people=0 兜底窗口 60s（仅 non-frozen 时生效）
		DistanceGateCm:             500,   // 丢轨点距雷达 >500cm（d_fall）抑制 lost_fall（贴地弱回波，见 §3.7）
	},
	CellHistory: cellHistoryParam{
		FakeAlarmThreshold:      3,
		ToleratedStillThreshold: 5,
		MaxToleranceFactor:      2.0,
		DecayHalfLifeDays:       30,
	},
}
