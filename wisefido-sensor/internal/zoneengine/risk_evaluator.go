// risk_evaluator.go — sensor zoneengine 评估 RoomState 的 RiskLevel。
//
// 阈值族（owl-common/card/risk_thresholds.go）：
//   · Bathroom: Day 8/15 standing + alone 30/45 day; RiskTime 5/8 standing + alone 20/30 night；multi 清零
//   · Default:  Day 8/15 / RiskTime 8/15 standing；multi 降一档
//   · Kitchen:  Day 12/18 / RiskTime 8/15 standing；multi 降一档
// RiskTime 默认 21:00-08:00（owl-common IsRiskTime）。
//
// timezone 暂用 UTC（loc=nil 退化）；per-zone timezone 注入留后续 wiring。
//
// 注：OOB 不构成独立风险（单源证据不够，需 cross-room）；OOB 作为输入证据喂到
// roomengine 的 BedsideFall / BedroomLostFall / Still fall 链路。

package zoneengine

import (
	"time"

	"owl-common/card"
)

// EvaluateRoomRiskLevel 按 room_type + risktime + multi-people 算 RoomState.RiskLevel。
//
// public 单元的 room 不应进入本函数 — 由 caller 按 unit_type==Public 直接跳过。
//
// 输入全部来自 RoomState 本身（都是 MIN 整数；无 ts 时间运算）：
//   - rs.AloneContinuousMin     独居持续分钟（bathroom alone 通道）
//   - rs.StandingContinuousMin  连续站立分钟（standing 通道）
func EvaluateRoomRiskLevel(rs *card.RoomState, roomType int, nowMs int64, loc *time.Location) int {
	if rs == nil || rs.TotalPeople <= 0 {
		return card.RiskNormal
	}
	isRiskTime := card.IsRiskTime(nowMs, loc)
	multi := rs.TotalPeople >= 2
	standingMin := rs.StandingContinuousMin
	aloneMin := rs.AloneContinuousMin
	standingTh := card.LookupStandingThresholds(roomType, isRiskTime)

	if roomType == card.RoomTypeBathroom {
		if multi {
			return card.RiskNormal
		}
		if standingMin >= standingTh.RiskMin {
			return card.RiskRisk
		}
		if standingMin >= standingTh.AttentionMin {
			return card.RiskAttention
		}
		aloneTh := card.BathroomAlone
		var attnMin, riskMin int
		if isRiskTime {
			attnMin, riskMin = aloneTh.RiskTimeAttentionMin, aloneTh.RiskTimeRiskMin
		} else {
			attnMin, riskMin = aloneTh.DayAttentionMin, aloneTh.DayRiskMin
		}
		if aloneMin >= riskMin {
			return card.RiskRisk
		}
		if aloneMin >= attnMin {
			return card.RiskAttention
		}
		return card.RiskNormal
	}

	// Default / Kitchen — standing-only
	if standingMin >= standingTh.RiskMin {
		if multi {
			return card.RiskAttention
		}
		return card.RiskRisk
	}
	if standingMin >= standingTh.AttentionMin {
		if multi {
			return card.RiskNormal
		}
		return card.RiskAttention
	}
	return card.RiskNormal
}
