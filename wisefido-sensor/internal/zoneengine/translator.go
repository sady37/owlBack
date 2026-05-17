// translator.go — ZoneEvent → card.BedState/RoomState/BathRoomState 纯函数翻译。
//
// Engine-owned 字段（types.go 注释明确）：
//   · BedState:      BedStatus / BedEvent / UpdatedAt / StartTime / DurationSec
//   · RoomState:     TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti
//   · BathRoomState: TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti
//
// 其余字段（TrackNumber / SleepStage / AreaPeople / StayFSMPhase / RoomName 等）由 prev
// 保留 —— stream_publisher 调时传入 prev。

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

// TranslateRoomState engine 字段 + prev 保留 AreaPeople/StandingContinuousMin/HasRisk。
func TranslateRoomState(e ZoneEvent, prev *card.RoomState) *card.RoomState {
	out := &card.RoomState{}
	if prev != nil {
		*out = *prev
	}
	out.UpdatedAt = e.NewState.UpdatedAt
	out.TotalPeople = e.NewState.Count
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.LastEnterTime = e.NewState.LastEnterTs
	case TransitionVacant:
		out.LastExitTime = e.NewState.LastExitTs
	}
	out.HasMulti = out.TotalPeople > 1
	return out
}

// TranslateBathRoomState engine 字段 + prev 保留 StayFSMPhase/Stay*/DeviceID/RoomName 等。
func TranslateBathRoomState(e ZoneEvent, prev *card.BathRoomState) *card.BathRoomState {
	out := &card.BathRoomState{}
	if prev != nil {
		*out = *prev
	}
	out.UpdatedAt = e.NewState.UpdatedAt
	out.TotalPeople = e.NewState.Count
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.LastEnterTime = e.NewState.LastEnterTs
	case TransitionVacant:
		out.LastExitTime = e.NewState.LastExitTs
	}
	out.HasMulti = out.TotalPeople > 1
	return out
}
