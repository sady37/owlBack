// translator.go — ZoneEvent → card.BedState/RoomState 纯函数翻译。
//
// Engine-owned 字段：
//   · BedState:  BedStatus / BedEvent / UpdatedAt / StartTime / DurationSec
//   · RoomState: TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti / Kind
//
// 其余字段（TrackNumber / SleepStage / AreaPeople / StaySec 等）由 prev 保留。

package zoneengine

import "owl-common/card"

// TranslateBedState 翻译为完整 BedState（engine 字段从 e 取，其余从 prev 保留）。
func TranslateBedState(e ZoneEvent, prev *card.BedState) *card.BedState {
	out := &card.BedState{}
	if prev != nil {
		*out = *prev
	}
	out.UpdatedAt = e.NewState.UpdatedAt

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

	// StartTime / DurationSec 占用窗口；leaving 期间保持累积
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.StartTime = e.NewState.LastEnterTs
		out.DurationSec = 0
	case TransitionVacant:
		if out.StartTime > 0 && e.NewState.LastExitTs > out.StartTime {
			out.DurationSec = int((e.NewState.LastExitTs - out.StartTime) / 1000)
		}
	case TransitionLeaving:
		if out.StartTime > 0 {
			out.DurationSec = int((e.NewState.UpdatedAt - out.StartTime) / 1000)
		}
	}

	return out
}

// TranslateRoomState engine 字段 + prev 保留 StandingContinuousMin/RiskLevel。
// roomType 决定 risk 阈值族（bathroom 较短 stay 即触发 risk）。
// count_change 直写路径不走 StateMachine，但 0↔N 跨越仍是逻辑 enter/exit，必须更新时间戳，
// 否则 FE OOR 计时永远停在最初 Vacant 时刻（per-card 长期 stale 的根因）。
//
// StaySec 累积语义（zone event 触发时即时算）：
//   - 占用期间（TotalPeople > 0 且 LastEnterTime > 0）：(UpdatedAt - LastEnterTime) / 1000
//   - 空房：0
//   - 进房瞬间（Occupied/Returned/0→N）：LastEnterTime 重置成新 enter 时戳，StaySec 同帧算下来=0
func TranslateRoomState(e ZoneEvent, prev *card.RoomState, roomType int) *card.RoomState {
	out := &card.RoomState{}
	if prev != nil {
		*out = *prev
	}
	if roomType != card.RoomTypeDefault {
		out.RoomType = roomType
	}
	out.UpdatedAt = e.NewState.UpdatedAt
	out.TotalPeople = e.NewState.Count
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
		}
	}

	// StaySec：占用即时计算（risk_evaluator 同帧读到一致值）
	if out.TotalPeople > 0 && out.LastEnterTime > 0 && out.UpdatedAt >= out.LastEnterTime {
		out.StaySec = int((out.UpdatedAt - out.LastEnterTime) / 1000)
	} else {
		out.StaySec = 0
	}

	out.RiskLevel = EvaluateRoomRiskLevel(out, e.NewState.UpdatedAt, nil)
	return out
}
