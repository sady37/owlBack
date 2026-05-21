// risk_evaluator.go — sensor zoneengine 评估 RoomState 的 RiskLevel。
//
// 设计要点：
//  · room_type-specific 阈值族（Bathroom vs 其它）—— bathroom 高风险，stay+standing 双轴严苛
//  · risk-time vs non-risk-time —— 夜间阈值更严（活动稀少时长静止 = 风险信号强）
//  · multi-people 修正 —— 有人陪伴时风险降级（bathroom 直接清零，其它降一档）
//
// 注：OOB 不构成独立风险（单源证据不够，需 cross-room）；OOB 作为输入证据喂到
// roomengine 的 BedsideFall / BedroomLostFall / Still fall 链路（详 sensor_fusion.go +
// track_manager.go），不在本 evaluator 范围。
//
// timezone 暂用 UTC（IsNightTime nil loc 退化）；per-zone timezone 注入留后续 wiring。
// risk-time 配置 23:30-07:30 默认；engine 后续可注入 SetRiskTimeConfig 覆盖。

package zoneengine

import (
	"time"

	"owl-common/card"
)

// RiskTimeConfig 风险时段（夜间）阈值。0 值视为未设置，回退默认。
type RiskTimeConfig struct {
	NightStartH int
	NightStartM int
	NightEndH   int
	NightEndM   int
}

var nightCfg = RiskTimeConfig{
	NightStartH: 23, NightStartM: 30,
	NightEndH: 7, NightEndM: 30,
}

// SetRiskTimeConfig engine 启动注入。0 值跳过。
func SetRiskTimeConfig(c RiskTimeConfig) {
	if c.NightStartH == 0 && c.NightStartM == 0 && c.NightEndH == 0 && c.NightEndM == 0 {
		return
	}
	nightCfg = c
}

// isNightTime 用 ms 时间戳判断是否风险时段。
// loc=nil 退化 UTC（后续 per-zone timezone 注入时由调用方传 IANA Location）。
func isNightTime(nowMs int64, loc *time.Location) bool {
	if loc == nil {
		loc = time.UTC
	}
	t := time.UnixMilli(nowMs).In(loc)
	h, m := t.Hour(), t.Minute()
	if h > nightCfg.NightStartH || (h == nightCfg.NightStartH && m >= nightCfg.NightStartM) {
		return true
	}
	if h < nightCfg.NightEndH || (h == nightCfg.NightEndH && m < nightCfg.NightEndM) {
		return true
	}
	return false
}

// EvaluateRoomRiskLevel 按 room_type + night + multi-people 算 RoomState.RiskLevel。
//
// 阈值族：
//   · Bathroom: stay 30/45min attention, 45/60min risk（night/day）；standing 5min attention 8min risk；multi 清零
//   · Default(其它，含 Kitchen): 仅看 standing（5min attention / 8min risk）；multi 降一档
//
// 注：public 单元的 room 不应进入本函数 —— 应由 caller 按 unit_type==Public 直接跳过。
//
// standingMin 是该 room 内 standing 累计分钟数；2026-05-18 后从 RoomState 字段挪到 TargetState
// （per-device），sensor 在 zoneengine 翻转点没有 card 视图，传 0 即可——standing risk 分支
// 退化为永远不触发（与挪动前 RoomState.StandingContinuousMin 无 writer 的真实行为一致）。
// cardagg 端 TargetMerger 合并后 max(devices) 可派生卡视图 standing，但 sensor 本评估在卡视图
// 之前——若未来需要 standing-driven risk，应由 cardagg 二次评估覆盖 RiskLevel。
func EvaluateRoomRiskLevel(rs *card.RoomState, roomType int, standingMin int, nowMs int64, loc *time.Location) int {
	if rs == nil || rs.TotalPeople <= 0 {
		return card.RiskNormal
	}
	isNight := isNightTime(nowMs, loc)
	multi := rs.TotalPeople >= 2

	if roomType == card.RoomTypeBathroom {
		if multi {
			return card.RiskNormal
		}
		if standingMin >= 8 {
			return card.RiskRisk
		}
		if standingMin >= 5 {
			return card.RiskAttention
		}
		// 独居时长（state-change-anchored）：仅在 AloneSinceTs > 0 时计算，避免 0 anchor 算出 huge 值
		aloneSec := 0
		if rs.AloneSinceTs > 0 && nowMs >= rs.AloneSinceTs {
			aloneSec = int((nowMs - rs.AloneSinceTs) / 1000)
		}
		if isNight {
			if aloneSec >= 30*60 {
				return card.RiskRisk
			}
			if aloneSec >= 20*60 {
				return card.RiskAttention
			}
		} else {
			if aloneSec >= 45*60 {
				return card.RiskRisk
			}
			if aloneSec >= 30*60 {
				return card.RiskAttention
			}
		}
		return card.RiskNormal
	}

	// Default / Kitchen / 其它 — standing-only
	if standingMin >= 8 {
		if multi {
			return card.RiskAttention
		}
		return card.RiskRisk
	}
	if standingMin >= 5 {
		if multi {
			return card.RiskNormal
		}
		return card.RiskAttention
	}
	return card.RiskNormal
}

