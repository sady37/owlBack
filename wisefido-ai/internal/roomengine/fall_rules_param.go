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
	Still       stillFallParam
	Lost        lostFallParam
	CellHistory cellHistoryParam
}

// FallRulesParam 全局单例。所有跌倒规则统一引用此变量。
//
// 约定：只读。修改值须改本文件并重编。运行时不修改字段。
var FallRulesParam = fallRulesParam{
	Still: stillFallParam{
		ToiletShowerSec:   15 * 60, // 15 min
		DenyZoneSec:       5 * 60,  // 5 min
		DefaultSec:        8 * 60,  // 8 min
		NonRiskTimeFactor: 1.2,     // non-risk × 1.2 → 18 min
		StaticPosCm:       20,
	},
	Lost: lostFallParam{
		RestZoneWaitSec:   60 * 60, // 60 min
		DenyZoneWaitSec:   5 * 60,
		WalkwayWaitSec:    5 * 60,
		ExitDistMinCm:     100, // 1 m
		SpatialJumpFactor: 0.5,
	},
	CellHistory: cellHistoryParam{
		FakeAlarmThreshold:      3,
		ToleratedStillThreshold: 5,
		MaxToleranceFactor:      2.0,
		DecayHalfLifeDays:       30,
	},
}
