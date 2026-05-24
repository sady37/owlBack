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
// standingMin 在 zoneengine 翻转点没有 card 视图传 0 — standing 分支永远不触发；
// cardagg 端 TargetMerger 合并后可派 standing-driven risk（cardagg 二次评估覆盖 RiskLevel）。
func EvaluateRoomRiskLevel(rs *card.RoomState, roomType int, standingMin int, nowMs int64, loc *time.Location) int {
	if rs == nil || rs.TotalPeople <= 0 {
		return card.RiskNormal
	}
	isRiskTime := card.IsRiskTime(nowMs, loc)
	multi := rs.TotalPeople >= 2

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
		// 独居时长（state-change-anchored）：仅在 AloneSinceTs > 0 时计算
		aloneSec := 0
		if rs.AloneSinceTs > 0 && nowMs >= rs.AloneSinceTs {
			aloneSec = int((nowMs - rs.AloneSinceTs) / 1000)
		}
		aloneTh := card.BathroomAlone
		var attnSec, riskSec int
		if isRiskTime {
			attnSec, riskSec = aloneTh.RiskTimeAttentionMin*60, aloneTh.RiskTimeRiskMin*60
		} else {
			attnSec, riskSec = aloneTh.DayAttentionMin*60, aloneTh.DayRiskMin*60
		}
		if aloneSec >= riskSec {
			return card.RiskRisk
		}
		if aloneSec >= attnSec {
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
