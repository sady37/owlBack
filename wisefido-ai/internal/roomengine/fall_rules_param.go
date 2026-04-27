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

	// 离最近出口最小距离（cm）—— 小于此值认为可能正常走出
	ExitDistMinCm int

	// 「空间跳跃」加权因子：表现过空间跳跃的 track 等待时间 × 此因子
	// 0.5 = 跳过的等待时间砍半（更敏感）
	SpatialJumpFactor float64

	// Frozen 检测：连续 N 帧 (track_id, x, y, z, pose, tc, rt) 字面相同 → 判 frozen
	// 实测 firmware 失锁后字面 byte-equal（无 noise），25 阈值给极少数抖动留余量
	FrozenSameThreshold int

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
		RestZoneWaitSec:       60 * 60, // 60 min
		DenyZoneWaitSec:       5 * 60,
		WalkwayWaitSec:        5 * 60,
		ExitDistMinCm:         100, // 1 m
		SpatialJumpFactor:     0.5,
		FrozenSameThreshold:   25,    // 连续 25 帧字面相同
		ImpossibleSpeedCm:     200,   // 硬 ghost：老人最快 100-150cm/s
		SuspectSpeedCm:        100,   // 软 ghost：需 EnterRoom 反证
		BirthFinalGraceMs:     2000,  // birth 终判延迟 2s
		EffectiveWaitFloorSec: 60,    // 兜底最少等 60s
	},
	CellHistory: cellHistoryParam{
		FakeAlarmThreshold:      3,
		ToleratedStillThreshold: 5,
		MaxToleranceFactor:      2.0,
		DecayHalfLifeDays:       30,
	},
}
