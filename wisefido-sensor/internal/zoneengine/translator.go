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

	// BedStatus + BedStatusTs：直接从 engine transition 时刻取。
	// BedStandby(bayesian 中性带) 优先 → 8 待机；anchor 用 SinceTs(进待机带时刻)。
	if e.NewState.BedStandby {
		out.BedStatus = card.BedStatusStandby
		out.BedStatusTs = e.NewState.SinceTs
	} else if e.NewState.IsPresent() {
		out.BedStatus = card.BedStatusInBed
		out.BedStatusTs = e.NewState.LastEnterTs
	} else {
		out.BedStatus = card.BedStatusNotInBed
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

// TranslateRoomState 把 ZoneEvent 翻成 RoomState 的 sensor-owner 字段（不读外部 prev）。
//
// **不在此处填**：AloneSinceTs / RiskLevel / RiskLevelTs — 这三项需要 prev cache 才能正确
// 计算（state-change-anchored 锚点 + 真实 aloneSec 时长），由 stream_publisher.applyAloneAndRisk
// 在 cache 入口前统一写入（单源真相，CLAUDE.md 规则 #1.3）。
//
// roomType 入参保留以兼容 caller，但本函数内不再使用（applyAloneAndRisk 用 publisher
// roomKindByCIDR 查）。
//
// ts 选取（engine 已保证 state-change-anchored）：
//   - LastEnterTs: engine.NewState.LastEnterTs（仅 0→N+ transition 时刻）
//   - LastExitTs:  engine.NewState.LastExitTs（仅 N→0 transition 时刻）
//   - 其它 ts = NewState.UpdatedAt
func TranslateRoomState(e ZoneEvent, roomType int) *card.RoomState {
	_ = roomType
	nowMs := e.NewState.UpdatedAt
	return &card.RoomState{
		RoomIdentifier: card.RoomIdentifier{RoomID: e.ZoneID},
		TotalPeople:    e.NewState.Count,
		TotalPeopleTs:  nowMs,
		LastEnterTs:    e.NewState.LastEnterTs,
		LastExitTs:     e.NewState.LastExitTs,
	}
}
