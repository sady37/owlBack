// sensor_state_projector_test.go — mergeBedStateSensorOwner / mergeBedStateSleepStage / mergeRoomState
// state-change-anchored 单元测试（2026-05-20 per-field ts 重构后）。
//
// 核心 invariant：
//   - 值不变 → ts 不变（保 prev）
//   - 值变  → 用 incoming 值 + ts
//   - incoming.ts=0 (placeholder) → 全保 prev
//   - BedEvent 是事件类，每次 incoming.BedEventTs > 0 都更新（不走 state-change-anchored）
//   - AloneContinuousMin / StandingContinuousMin 是 incoming-wins (sensor 已 gate + 算)

package consumer

import (
	"testing"

	"owl-common/card"
)

func TestMergeBedField_ValueUnchangedKeepsPrevTs(t *testing.T) {
	v, ts := mergeBedField(1, 100, 1, 200)
	if v != 1 || ts != 100 {
		t.Errorf("value unchanged should keep prev ts: got (%d,%d), want (1,100)", v, ts)
	}
}

func TestMergeBedField_ValueChangedUsesIncoming(t *testing.T) {
	v, ts := mergeBedField(0, 100, 1, 200)
	if v != 1 || ts != 200 {
		t.Errorf("value changed should use incoming: got (%d,%d), want (1,200)", v, ts)
	}
}

func TestMergeBedField_IncomingNoTsKeepsPrev(t *testing.T) {
	// incoming.ts=0 → placeholder，保 prev
	v, ts := mergeBedField(0, 100, 1, 0)
	if v != 0 || ts != 100 {
		t.Errorf("incoming ts=0 should keep prev: got (%d,%d), want (0,100)", v, ts)
	}
}

func TestMergeBedStateSensorOwner_StateChangeAnchored(t *testing.T) {
	prev := &card.BedState{
		BedStatus: 0, BedStatusTs: 100,
		TrackNumber: 1, TrackNumberTs: 100,
		BedConfidence: 90, BedConfidenceTs: 100,
		SleepStage: 2, SleepStageTs: 50, // sleepstage 字段 prev 保留
	}
	// incoming：BedStatus 翻转 0→1，TrackNumber 不变（仍1），BedEvent fired
	incoming := &card.BedState{
		BedStatus: 1, BedStatusTs: 200,
		TrackNumber: 1, TrackNumberTs: 200, // 值不变 ts 不更新
		BedConfidence: 90, BedConfidenceTs: 200, // 值不变 ts 不更新
		BedEvent: 1, BedEventTs: 200, // 事件类，更新
	}
	out := mergeBedStateSensorOwner(prev, incoming)

	if out.BedStatus != 1 || out.BedStatusTs != 200 {
		t.Errorf("BedStatus changed: got (%d,%d), want (1,200)", out.BedStatus, out.BedStatusTs)
	}
	if out.TrackNumber != 1 || out.TrackNumberTs != 100 {
		t.Errorf("TrackNumber unchanged: got (%d,%d), want (1,100)", out.TrackNumber, out.TrackNumberTs)
	}
	if out.BedConfidence != 90 || out.BedConfidenceTs != 100 {
		t.Errorf("BedConfidence unchanged: got (%d,%d), want (90,100)", out.BedConfidence, out.BedConfidenceTs)
	}
	if out.BedEvent != 1 || out.BedEventTs != 200 {
		t.Errorf("BedEvent event-class always update: got (%d,%d), want (1,200)", out.BedEvent, out.BedEventTs)
	}
	// sleepstage 字段保留 prev
	if out.SleepStage != 2 || out.SleepStageTs != 50 {
		t.Errorf("SleepStage path should be untouched: got (%d,%d), want (2,50)", out.SleepStage, out.SleepStageTs)
	}
}

func TestMergeBedStateSleepStage_OnlySleepFields(t *testing.T) {
	prev := &card.BedState{
		BedStatus: 0, BedStatusTs: 100,
		SleepStage: 1, SleepStageTs: 100,
		SleepConfidence: 60, SleepConfidenceTs: 100,
	}
	incoming := &card.BedState{
		SleepStage: 4, SleepStageTs: 200,
		SleepConfidence: 90, SleepConfidenceTs: 200,
	}
	out := mergeBedStateSleepStage(prev, incoming)

	if out.SleepStage != 4 || out.SleepStageTs != 200 {
		t.Errorf("SleepStage updated: got (%d,%d), want (4,200)", out.SleepStage, out.SleepStageTs)
	}
	if out.SleepConfidence != 90 || out.SleepConfidenceTs != 200 {
		t.Errorf("SleepConfidence updated: got (%d,%d), want (90,200)", out.SleepConfidence, out.SleepConfidenceTs)
	}
	// BedStatus 保留 prev
	if out.BedStatus != 0 || out.BedStatusTs != 100 {
		t.Errorf("BedStatus untouched: got (%d,%d), want (0,100)", out.BedStatus, out.BedStatusTs)
	}
}

// AloneContinuousMin/StandingContinuousMin 是 sensor 60s tick 算好的整数 MIN，
// mergeRoomState 直接 incoming 覆 prev（不做 cross-field 派生 — sensor 端已 gate）。
func TestMergeRoomState_AloneContinuousMin_IncomingOverwrites(t *testing.T) {
	prev := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 5}
	incoming := &card.RoomState{
		TotalPeople: 1, TotalPeopleTs: 200,
		AloneContinuousMin: 6, // sensor 下一 tick 算到 6min
	}
	out := mergeRoomState(prev, incoming)
	if out.AloneContinuousMin != 6 {
		t.Errorf("incoming MIN should overwrite: got %d, want 6", out.AloneContinuousMin)
	}
}

func TestMergeRoomState_AloneContinuousMin_ClearOnMultiPeople(t *testing.T) {
	prev := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 10}
	incoming := &card.RoomState{
		TotalPeople: 2, TotalPeopleTs: 200,
		AloneContinuousMin: 0, // sensor 已 gate（多人）
	}
	out := mergeRoomState(prev, incoming)
	if out.AloneContinuousMin != 0 {
		t.Errorf("multi-people should clear MIN: got %d", out.AloneContinuousMin)
	}
	if out.TotalPeople != 2 {
		t.Errorf("TotalPeople should = 2, got %d", out.TotalPeople)
	}
}

func TestMergeRoomState_StandingContinuousMin_IncomingOverwrites(t *testing.T) {
	prev := &card.RoomState{TotalPeople: 1, StandingContinuousMin: 3}
	incoming := &card.RoomState{
		TotalPeople: 1, TotalPeopleTs: 200,
		StandingContinuousMin: 5,
	}
	out := mergeRoomState(prev, incoming)
	if out.StandingContinuousMin != 5 {
		t.Errorf("incoming Standing MIN should overwrite: got %d, want 5", out.StandingContinuousMin)
	}
}

func TestMergeRoomState_LastExitTs_OnVacant(t *testing.T) {
	prev := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 10, LastEnterTs: 100}
	incoming := &card.RoomState{
		TotalPeople: 0, TotalPeopleTs: 500,
		LastExitTs:         500,
		AloneContinuousMin: 0,
	}
	out := mergeRoomState(prev, incoming)
	if out.LastExitTs != 500 {
		t.Errorf("1→0: LastExitTs should = 500, got %d", out.LastExitTs)
	}
	if out.AloneContinuousMin != 0 {
		t.Errorf("1→0: AloneContinuousMin should clear, got %d", out.AloneContinuousMin)
	}
}
