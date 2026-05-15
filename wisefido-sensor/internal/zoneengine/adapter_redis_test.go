package zoneengine

import (
	"testing"

	"owl-common/card"
)

func newRedisAdapterForTest() *RedisAdapter {
	// nil reader/writer 是 ok 的 —— 测试只调 deriveXxx 不走 IO
	return &RedisAdapter{}
}

func TestRedisAdapter_BedOccupiedSetsFields(t *testing.T) {
	a := newRedisAdapterForTest()
	now := int64(1_700_000_000_000)
	prev := &card.BedState{
		SleepStage:      card.SleepStageDeep,
		SleepConfidence: 90,
		TrackNumber:     1,
	}
	e := ZoneEvent{
		Transition: TransitionOccupied,
		NewState: ZoneState{
			Status:      StatusOccupied,
			Occupied:    true,
			LastEnterTs: now,
			UpdatedAt:   now,
		},
	}
	out := a.deriveBedState(e, prev)
	if out.BedStatus != 0 {
		t.Errorf("want bed_status=0, got %d", out.BedStatus)
	}
	if out.BedEvent != 0 {
		t.Errorf("want bed_event=0 (InBed), got %d", out.BedEvent)
	}
	if out.StartTime != now {
		t.Errorf("want start_time=%d, got %d", now, out.StartTime)
	}
	if out.DurationSec != 0 {
		t.Errorf("want duration_sec=0 on enter, got %d", out.DurationSec)
	}
	// 关键：保留 prev 字段
	if out.SleepStage != card.SleepStageDeep {
		t.Errorf("SleepStage not preserved: %d", out.SleepStage)
	}
	if out.SleepConfidence != 90 {
		t.Errorf("SleepConfidence not preserved: %d", out.SleepConfidence)
	}
	if out.TrackNumber != 1 {
		t.Errorf("TrackNumber not preserved: %d", out.TrackNumber)
	}
}

func TestRedisAdapter_BedVacantSetsLeftBedAndDuration(t *testing.T) {
	a := newRedisAdapterForTest()
	start := int64(1_700_000_000_000)
	exit := start + 30_000 // 30s 在床
	prev := &card.BedState{
		StartTime: start,
		BedStatus: 0,
	}
	e := ZoneEvent{
		Transition: TransitionVacant,
		NewState: ZoneState{
			Status:     StatusVacant,
			LastExitTs: exit,
			UpdatedAt:  exit,
		},
	}
	out := a.deriveBedState(e, prev)
	if out.BedStatus != 1 {
		t.Errorf("want bed_status=1, got %d", out.BedStatus)
	}
	if out.BedEvent != 1 {
		t.Errorf("want bed_event=1 (LeftBed), got %d", out.BedEvent)
	}
	if out.DurationSec != 30 {
		t.Errorf("want duration_sec=30, got %d", out.DurationSec)
	}
}

func TestRedisAdapter_BedLeavingMidStateMarksNoneEvent(t *testing.T) {
	a := newRedisAdapterForTest()
	start := int64(1_700_000_000_000)
	mid := start + 5_000
	prev := &card.BedState{StartTime: start, BedStatus: 0, BedEvent: 0}
	e := ZoneEvent{
		Transition: TransitionLeaving,
		NewState: ZoneState{
			Status:    StatusLeaving,
			Occupied:  true, // Leaving still IsPresent
			UpdatedAt: mid,
		},
	}
	out := a.deriveBedState(e, prev)
	if out.BedStatus != 0 {
		t.Errorf("want bed_status=0 (still IsPresent during Leaving), got %d", out.BedStatus)
	}
	if out.BedEvent != 8 {
		t.Errorf("want bed_event=8 (None) during Leaving, got %d", out.BedEvent)
	}
	if out.DurationSec != 5 {
		t.Errorf("want duration_sec=5 during Leaving, got %d", out.DurationSec)
	}
}

func TestRedisAdapter_BedReturnedResetsStartAndDuration(t *testing.T) {
	a := newRedisAdapterForTest()
	originalStart := int64(1_700_000_000_000)
	returnedAt := originalStart + 6_000
	prev := &card.BedState{
		StartTime:   originalStart,
		BedStatus:   0,
		BedEvent:    8,
		DurationSec: 6,
	}
	e := ZoneEvent{
		Transition: TransitionReturned,
		NewState: ZoneState{
			Status:      StatusOccupied,
			Occupied:    true,
			LastEnterTs: returnedAt,
			UpdatedAt:   returnedAt,
		},
	}
	out := a.deriveBedState(e, prev)
	// Returned 视同新会话开始（用户老人坐回 = 新一段在床区间）
	if out.StartTime != returnedAt {
		t.Errorf("want StartTime reset on Returned, got %d", out.StartTime)
	}
	if out.DurationSec != 0 {
		t.Errorf("want DurationSec reset on Returned, got %d", out.DurationSec)
	}
	if out.BedEvent != 0 {
		t.Errorf("want bed_event=0 (InBed) on Returned, got %d", out.BedEvent)
	}
}

func TestRedisAdapter_RoomOccupiedSetsLastEnter(t *testing.T) {
	a := newRedisAdapterForTest()
	now := int64(1_700_000_000_000)
	prev := &card.RoomState{
		AreaPeople:            map[string]int{"sofa": 1},
		StandingContinuousMin: 5,
	}
	e := ZoneEvent{
		Transition: TransitionOccupied,
		NewState: ZoneState{
			Status:      StatusOccupied,
			Count:       1,
			LastEnterTs: now,
			UpdatedAt:   now,
		},
	}
	out := a.deriveRoomState(e, prev)
	if out.TotalPeople != 1 {
		t.Errorf("want total_people=1, got %d", out.TotalPeople)
	}
	if out.LastEnterTime != now {
		t.Errorf("want last_enter_time=%d, got %d", now, out.LastEnterTime)
	}
	if out.HasMulti {
		t.Errorf("want has_multi=false for count=1")
	}
	// 保留 satellite 字段
	if out.AreaPeople["sofa"] != 1 {
		t.Errorf("AreaPeople not preserved")
	}
	if out.StandingContinuousMin != 5 {
		t.Errorf("StandingContinuousMin not preserved")
	}
}

func TestRedisAdapter_RoomCountChangeSetsHasMulti(t *testing.T) {
	a := newRedisAdapterForTest()
	now := int64(1_700_000_000_000)
	e := ZoneEvent{
		Transition: TransitionCountChange,
		NewState: ZoneState{
			Count:     3,
			UpdatedAt: now,
		},
	}
	out := a.deriveRoomState(e, nil)
	if out.TotalPeople != 3 {
		t.Errorf("want total_people=3, got %d", out.TotalPeople)
	}
	if !out.HasMulti {
		t.Errorf("want has_multi=true for count=3")
	}
	// count_change 不应改 last_enter / last_exit
	if out.LastEnterTime != 0 || out.LastExitTime != 0 {
		t.Errorf("count_change should not touch enter/exit times: %+v", out)
	}
}

func TestRedisAdapter_BathroomPreservesStayFSM(t *testing.T) {
	a := newRedisAdapterForTest()
	now := int64(1_700_000_000_000)
	prev := &card.BathRoomState{
		DeviceID:          "dev-uuid",
		RoomName:          "Bathroom 1",
		StayFSMPhase:      "Armed",
		StayArmEnterAt:    now - 60_000,
		StayResolveExitAt: 0,
		StaySec:           60,
	}
	e := ZoneEvent{
		Transition: TransitionVacant,
		NewState: ZoneState{
			Status:     StatusVacant,
			Count:      0,
			LastExitTs: now,
			UpdatedAt:  now,
		},
	}
	out := a.deriveBathroomState(e, prev)
	if out.TotalPeople != 0 {
		t.Errorf("want total_people=0, got %d", out.TotalPeople)
	}
	if out.LastExitTime != now {
		t.Errorf("want last_exit_time=%d, got %d", now, out.LastExitTime)
	}
	// stay_fsm satellite 字段必须保留 —— engine 不动
	if out.StayFSMPhase != "Armed" {
		t.Errorf("StayFSMPhase not preserved: %q", out.StayFSMPhase)
	}
	if out.StayArmEnterAt != now-60_000 {
		t.Errorf("StayArmEnterAt not preserved: %d", out.StayArmEnterAt)
	}
	if out.StaySec != 60 {
		t.Errorf("StaySec not preserved: %d", out.StaySec)
	}
	if out.DeviceID != "dev-uuid" {
		t.Errorf("DeviceID not preserved")
	}
	if out.RoomName != "Bathroom 1" {
		t.Errorf("RoomName not preserved")
	}
}

func TestRedisAdapter_translateRoutesByZoneType(t *testing.T) {
	a := newRedisAdapterForTest()
	cur := &card.CardStatus{}

	t.Run("bed", func(t *testing.T) {
		out := a.translate(ZoneEvent{ZoneType: ZoneTypeBed, Transition: TransitionOccupied}, cur)
		if out.bedState == nil || out.roomState != nil || out.bathRoomState != nil {
			t.Errorf("bed event should only set bed: %+v", out)
		}
	})
	t.Run("room", func(t *testing.T) {
		out := a.translate(ZoneEvent{ZoneType: ZoneTypeRoom, Transition: TransitionOccupied}, cur)
		if out.roomState == nil || out.bedState != nil || out.bathRoomState != nil {
			t.Errorf("room event should only set room: %+v", out)
		}
	})
	t.Run("bathroom", func(t *testing.T) {
		out := a.translate(ZoneEvent{ZoneType: ZoneTypeBathroom, Transition: TransitionOccupied}, cur)
		if out.bathRoomState == nil || out.bedState != nil || out.roomState != nil {
			t.Errorf("bathroom event should only set bathroom: %+v", out)
		}
	})
}
