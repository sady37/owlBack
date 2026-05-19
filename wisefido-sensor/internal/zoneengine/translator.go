// translator.go — ZoneEvent → card.BedState/RoomState 纯函数翻译。
//
// 设计原则：sensor 只输出自己负责的 engine 字段（"sensor 不读 card:state"，
// 不传 prev 进来；跨进程合并/保留非 sensor owner 字段是 cardagg 端职责）。
//
// Sensor owner 字段（2026-05-19 S4+S5b 扩展）：
//   · BedState (via bed.state category): UpdatedAt / BedStatus / BedEvent / StartTime /
//                                         DurationSec / TrackNumber (S5b) / BedConfidence (S5b)
//   · BedState (via bed.sleepstage category, separate publish path):
//                                         SleepStage / SleepConfidence (S4)
//   · RoomState: UpdatedAt / RoomType / TotalPeople / LastEnterTime / LastExitTime / StaySec / RiskLevel
//
// 非 sensor owner 字段（zero value 输出，cardagg projector 字段级 merge 时不覆盖）：
//   · RoomState: LastExitToOutside / StandingContinuousMin
//
// 关键：BedState.StartTime + DurationSec 跨 Vacant transition 仍可正确计算，
// 因为 zoneengine ZoneState.LastEnterTs 在 Vacant transition 内保留（不重置）。

package zoneengine

import "owl-common/card"

// TranslateBedState 仅填 sensor owner 字段（不读外部 prev）。
func TranslateBedState(e ZoneEvent) *card.BedState {
	out := &card.BedState{
		UpdatedAt: e.NewState.UpdatedAt,
		StartTime: e.NewState.LastEnterTs, // engine 跨 transition 保留
	}

	// BedStatus：0=在床, 1=离床
	if e.NewState.IsPresent() {
		out.BedStatus = 0
	} else {
		out.BedStatus = 1
	}

	// BedEvent：0=InBed / 1=LeftBed / 8=None
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.BedEvent = 0
	case TransitionVacant:
		out.BedEvent = 1
	case TransitionLeaving, TransitionCountChange:
		out.BedEvent = 8
	}

	// DurationSec — 用 engine LastEnterTs 算（不依赖 prev）
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.DurationSec = 0
	case TransitionVacant:
		if e.NewState.LastEnterTs > 0 && e.NewState.LastExitTs > e.NewState.LastEnterTs {
			out.DurationSec = int((e.NewState.LastExitTs - e.NewState.LastEnterTs) / 1000)
		}
	case TransitionLeaving:
		if e.NewState.LastEnterTs > 0 && e.NewState.UpdatedAt >= e.NewState.LastEnterTs {
			out.DurationSec = int((e.NewState.UpdatedAt - e.NewState.LastEnterTs) / 1000)
		}
	}

	// S5b: TrackNumber + BedConfidence
	//   TrackNumber  = NewState.Count（bed FSM 0/1；多人共床 v2 不建模，cap 0..2 是 v1 历史）
	//   BedConfidence = source-based 数据可信度（v1 真意 [[cardagg_v1_to_v2_migration_audit]]）：
	//                    sleepace → 90 / radar → 60 / 其它/未知 → 0
	out.TrackNumber = e.NewState.Count
	out.BedConfidence = bedConfidenceForSource(e.NewState.LastSource)

	return out
}

// bedConfidenceForSource 信号源 → 数据可信度（与 v1 PublishBedStateFromEvent 同口径）。
func bedConfidenceForSource(source string) int {
	switch source {
	case "sleepace":
		return 90
	case "radar":
		return 60
	}
	return 0
}

// TranslateRoomState 仅填 sensor owner 字段（不读外部 prev）。
// roomType 由 caller 显式指定（ZoneTypeRoom→Default / ZoneTypeBathroom→Bathroom），
// 每次都写入；不存在"未传时保留旧值"语义——sensor 知道自己处理的是什么 zone。
//
// StaySec 累积语义（zone event 触发时即时算）：
//   - 占用期间（TotalPeople > 0 且 LastEnterTime > 0）：(UpdatedAt - LastEnterTime) / 1000
//   - 空房：0
//   - 进房瞬间（Occupied/Returned/0→N）：LastEnterTime 重置成新 enter 时戳，StaySec 同帧算下来=0
//
// RiskLevel 用本次输出的 RoomState 自评估（不读 prev StandingContinuousMin —— 该字段
// 已挪到 TargetState，risk_evaluator 未来改从 cardagg 合并 target 读；当前 sensor 这里
// StandingContinuousMin=0，bathroom standing risk 路径暂不触发，由 cardagg 接 TargetMerger
// 后回流注入 / 或 risk_evaluator 重构）。
func TranslateRoomState(e ZoneEvent, roomType int) *card.RoomState {
	out := &card.RoomState{
		RoomType:    roomType,
		UpdatedAt:   e.NewState.UpdatedAt,
		TotalPeople: e.NewState.Count,
	}
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.LastEnterTime = e.NewState.LastEnterTs
	case TransitionVacant:
		out.LastExitTime = e.NewState.LastExitTs
	case TransitionCountChange:
		if e.PrevState.Count == 0 && e.NewState.Count > 0 {
			out.LastEnterTime = e.NewState.UpdatedAt
		} else if e.PrevState.Count > 0 && e.NewState.Count == 0 {
			out.LastExitTime = e.NewState.UpdatedAt
		} else {
			// 1→N / N→M 同房间 enter/exit 不发生；engine LastEnterTs 仍是当前段开始时
			out.LastEnterTime = e.NewState.LastEnterTs
		}
	default:
		out.LastEnterTime = e.NewState.LastEnterTs
		out.LastExitTime = e.NewState.LastExitTs
	}

	// StaySec：占用即时计算（risk_evaluator 同帧读到一致值）
	if out.TotalPeople > 0 && out.LastEnterTime > 0 && out.UpdatedAt >= out.LastEnterTime {
		out.StaySec = int((out.UpdatedAt - out.LastEnterTime) / 1000)
	} else {
		out.StaySec = 0
	}

	// standingMin=0：StandingContinuousMin 字段已挪到 TargetState (per-device)，sensor 翻转
	// 点没有 card 视图；行为退化为永远不触发 standing risk（同挪动前无 writer 现实）。
	out.RiskLevel = EvaluateRoomRiskLevel(out, 0, e.NewState.UpdatedAt, nil)
	return out
}
