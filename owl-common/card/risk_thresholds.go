// risk_thresholds.go — RoomState.RiskLevel 派生 + Section2LeftDown 显示阈值。
//
// 临床业务参数（不放 config.yaml）：阈值确定后稳定，sensor 端 risk_evaluator
// 跟 cardagg 端 Section2LeftDown 派生共用同一份，避免跨服务漂移。
//
// 修改阈值后重启 owlback 全套。dev 环境调试用短阈值，请直接编辑此文件。

package card

import "time"

// StandingThresholds 站立时长阈值（分钟）。
//
//	standingMin >= RiskMin              → RiskRisk
//	RiskMin > standingMin >= AttentionMin → RiskAttention
//	AttentionMin > standingMin            → 不触发
type StandingThresholds struct {
	AttentionMin int
	RiskMin      int
}

// StandingThresholdsByTime room_type 的 standing 阈值（day / risktime 双档）。
// risktime 默认 21:00-08:00：夜间老人活动模式异常 → 阈值收紧。
type StandingThresholdsByTime struct {
	Day      StandingThresholds
	RiskTime StandingThresholds
}

// StandingByRoomType — 阈值三档 × day/risktime 二档：
//
//	Bathroom: Day 8/15 / RiskTime 5/8         — 夜间 fall risk 高敏感（夜尿/起夜跌倒）
//	Default:  Day 8/15 / RiskTime 8/15        — 平衡，不分 day/night
//	Kitchen:  Day 12/18 / RiskTime 8/15       — 白天容忍做饭长站；夜间做饭异常按 default 严
var StandingByRoomType = map[int]StandingThresholdsByTime{
	RoomTypeBathroom: {
		Day:      StandingThresholds{AttentionMin: 8, RiskMin: 15},
		RiskTime: StandingThresholds{AttentionMin: 5, RiskMin: 8},
	},
	RoomTypeDefault: {
		Day:      StandingThresholds{AttentionMin: 8, RiskMin: 15},
		RiskTime: StandingThresholds{AttentionMin: 8, RiskMin: 15},
	},
	RoomTypeKitchen: {
		Day:      StandingThresholds{AttentionMin: 12, RiskMin: 18},
		RiskTime: StandingThresholds{AttentionMin: 8, RiskMin: 15},
	},
}

// LookupStandingThresholds 按 (room_type, isNight) 取阈值；未注册 type 回退 Default。
func LookupStandingThresholds(roomType int, isRiskTime bool) StandingThresholds {
	th, ok := StandingByRoomType[roomType]
	if !ok {
		th = StandingByRoomType[RoomTypeDefault]
	}
	if isRiskTime {
		return th.RiskTime
	}
	return th.Day
}

// BathroomAloneThresholds bathroom 独居时长阈值（仅 bathroom）— alone time fall/晕厥风险信号。
//
//	RiskTime (21:00-08:00):  Attention 20min / Risk 30min
//	Day:                     Attention 30min / Risk 45min
type BathroomAloneThresholds struct {
	RiskTimeAttentionMin int
	RiskTimeRiskMin      int
	DayAttentionMin      int
	DayRiskMin           int
}

var BathroomAlone = BathroomAloneThresholds{
	RiskTimeAttentionMin: 20,
	RiskTimeRiskMin:      30,
	DayAttentionMin:      30,
	DayRiskMin:           45,
}

// RiskTimeConfig 风险时段（夜间）阈值。默认 21:00-08:00（老人作息）。
type RiskTimeConfig struct {
	StartH int
	StartM int
	EndH   int
	EndM   int
}

var riskTimeCfg = RiskTimeConfig{
	StartH: 21, StartM: 0,
	EndH: 8, EndM: 0,
}

// SetRiskTimeConfig engine 启动注入；0 值跳过。
func SetRiskTimeConfig(c RiskTimeConfig) {
	if c.StartH == 0 && c.StartM == 0 && c.EndH == 0 && c.EndM == 0 {
		return
	}
	riskTimeCfg = c
}

// IsRiskTime 判断 ms 时间戳是否在 risktime 窗口（21:00-08:00 默认）。
// loc=nil 退化 UTC（caller 应传 unit timezone）。
func IsRiskTime(nowMs int64, loc *time.Location) bool {
	if loc == nil {
		loc = time.UTC
	}
	t := time.UnixMilli(nowMs).In(loc)
	h, m := t.Hour(), t.Minute()
	if h > riskTimeCfg.StartH || (h == riskTimeCfg.StartH && m >= riskTimeCfg.StartM) {
		return true
	}
	if h < riskTimeCfg.EndH || (h == riskTimeCfg.EndH && m < riskTimeCfg.EndM) {
		return true
	}
	return false
}
