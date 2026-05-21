// translator.go — ZoneEvent → card.BedState/RoomState 纯函数翻译。
//
// **2026-05-20 重构后语义（per-field ts）**：
//   sensor 输出"当下观测":值 + ts = engine NewState.UpdatedAt（sensor 知道的最近变更时刻）。
//   state-change-anchored 规则由 cardagg projector merge 端实施（值同 → ts 保 prev）。
//
// sensor owner 字段：
//   · BedState (via bed.state):  BedStatus / TrackNumber / BedConfidence / BedEvent
//   · BedState (via bed.sleepstage): SleepStage / SleepConfidence (S4 SleepStageConsumer)
//   · RoomState: RoomType / TotalPeople / LastEnterTs / LastExitTs / RiskLevel / AloneSinceTs
//
// 非 sensor owner 字段：
//   · RoomState: LastExitToOutside / (StandingContinuousMin 已挪 TargetState)
//
// BedStatus*Ts / LastEnterTs / LastExitTs 这几个 ts 由 engine 显式语义来源（applyTransitionToState
// 在翻 Occupied 时更新 LastEnterTs，翻 Vacant 时更新 LastExitTs）—— sensor 直接 forward。

package zoneengine

import "owl-common/card"

// TranslateBedState 仅填 sensor owner 字段（不读外部 prev）。
// ts 选取约定（state-change-anchored 由 cardagg merge 端把关）：
//   - BedStatusTs: 当前 BedStatus 的真实起始时刻
//       InBed (BedStatus=0)  → NewState.LastEnterTs
//       LeftBed (BedStatus=1) → NewState.LastExitTs
//   - 其它字段 ts = NewState.UpdatedAt（cardagg merge 比 prev 决定是否实采用）
//   - BedEvent 是事件不是状态，BedEventTs 永远 = NewState.UpdatedAt
func TranslateBedState(e ZoneEvent) *card.BedState {
	out := &card.BedState{}
	nowMs := e.NewState.UpdatedAt

	// BedStatus + BedStatusTs：直接从 engine transition 时刻取
	if e.NewState.IsPresent() {
		out.BedStatus = 0
		out.BedStatusTs = e.NewState.LastEnterTs
	} else {
		out.BedStatus = 1
		out.BedStatusTs = e.NewState.LastExitTs
	}

	// TrackNumber + Ts （bed FSM Count 0/1）
	out.TrackNumber = e.NewState.Count
	out.TrackNumberTs = nowMs

	// BedConfidence + Ts （source-based: sleepace=90 / radar=60）
	out.BedConfidence = bedConfidenceForSource(e.NewState.LastSource)
	out.BedConfidenceTs = nowMs

	// BedEvent：0=InBed / 1=LeftBed / 8=None；事件类，每次 transition 都刷新 ts
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.BedEvent = 0
	case TransitionVacant:
		out.BedEvent = 1
	case TransitionLeaving, TransitionCountChange:
		out.BedEvent = 8
	}
	out.BedEventTs = nowMs

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
// roomType 是 caller 现场 hint（仅供 RiskLevel 计算）—— 不再写入 RoomState；
// room_type 是静态属性，由 cardagg 通过 CardMeta (rooms.room_type LPM) 派生，sensor 不发。
//
// ts 选取（state-change-anchored 由 cardagg merge 端把关）：
//   - LastEnterTs: engine.NewState.LastEnterTs（仅 0→N+ transition 时刻；engine 已正确维护）
//   - LastExitTs:  engine.NewState.LastExitTs（仅 N→0 transition 时刻）
//   - AloneSinceTs: 当前 Count==1 时输出 NewState.UpdatedAt（cardagg merge 判 0→1 / N→1 才用）
//   - 其它 ts = NewState.UpdatedAt
//
// RiskLevel: sensor 自评（StandingContinuousMin 已挪 TargetState，此处 0；bathroom standing
// risk 待 cardagg merge target 后重构）。
func TranslateRoomState(e ZoneEvent, roomType int) *card.RoomState {
	nowMs := e.NewState.UpdatedAt
	out := &card.RoomState{
		TotalPeople:   e.NewState.Count,
		TotalPeopleTs: nowMs,
		LastEnterTs:   e.NewState.LastEnterTs,
		LastExitTs:    e.NewState.LastExitTs,
	}

	// AloneSinceTs：当前 Count==1 时输出"刚才"为可能锚点；
	// cardagg merge 端按 prev.TotalPeople 决定是 0→1 / N→1 重锚还是延续 prev。
	if e.NewState.Count == 1 {
		out.AloneSinceTs = nowMs
	}

	// RiskLevel + Ts：sensor 自评（同 zone-event 帧内一致）— 用 roomType 参数
	out.RiskLevel = EvaluateRoomRiskLevel(out, roomType, 0, nowMs, nil)
	out.RiskLevelTs = nowMs

	return out
}
