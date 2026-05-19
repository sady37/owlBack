// translator_test.go — TranslateRoomState / TranslateBedState 纯函数表驱动测试。
//
// T1 契约测试第一版（card_display_projector handoff Task 2.1 配套）：
//   - StaySec 累积公式（占用即时算 / 空房归零 / 进房瞬间=0）
//   - LastEnterTime / LastExitTime / TotalPeople / RoomType 派生
//   - BedState BedStatus / BedEvent / DurationSec 派生
//
// 不测：RiskLevel（由 EvaluateRoomRiskLevel 单独覆盖）；prev 保留字段（StandingContinuousMin 等）
// 在 Task 2.2 producer 落地后再测。

package zoneengine

import (
	"testing"

	"owl-common/card"
)

func TestTranslateRoomState_StaySec(t *testing.T) {
	tests := []struct {
		name           string
		prev           *card.RoomState
		event          ZoneEvent
		wantTotal      int
		wantStaySec    int
		wantLastEnter  int64
		wantLastExit   int64
	}{
		{
			name:  "occupied — 进房 LastEnterTs 重置，StaySec 同帧=0",
			prev:  nil,
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState: ZoneState{
					Count:       1,
					LastEnterTs: 1_700_000_000_000,
					UpdatedAt:   1_700_000_000_000,
				},
			},
			wantTotal:     1,
			wantStaySec:   0,
			wantLastEnter: 1_700_000_000_000,
		},
		{
			name:  "occupied 30s — StaySec = (now - enter)/1000",
			prev:  &card.RoomState{TotalPeople: 1, LastEnterTime: 1_700_000_000_000},
			event: ZoneEvent{
				Transition: TransitionOccupied, // 同 zone 内累加（occupied transition 重置 LastEnterTime）
				NewState: ZoneState{
					Count:       1,
					LastEnterTs: 1_700_000_000_000, // 保持原 enter ts
					UpdatedAt:   1_700_000_030_000, // +30s
				},
			},
			wantTotal:     1,
			wantStaySec:   30,
			wantLastEnter: 1_700_000_000_000,
		},
		{
			name:  "count_change 0→2 — LastEnterTime 设为 UpdatedAt，StaySec=0",
			prev:  &card.RoomState{TotalPeople: 0, LastExitTime: 1_700_000_000_000},
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 0},
				NewState: ZoneState{
					Count:     2,
					UpdatedAt: 1_700_000_010_000,
				},
			},
			wantTotal:     2,
			wantStaySec:   0,
			wantLastEnter: 1_700_000_010_000,
		},
		{
			name: "count_change 1→3 持续 60s — 不重置 LastEnterTime，StaySec 累计",
			prev: &card.RoomState{
				TotalPeople:   1,
				LastEnterTime: 1_700_000_000_000,
			},
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 1},
				NewState: ZoneState{
					Count:     3,
					UpdatedAt: 1_700_000_060_000,
				},
			},
			wantTotal:     3,
			wantStaySec:   60,
			wantLastEnter: 1_700_000_000_000,
		},
		{
			name: "vacant — TotalPeople=0，StaySec 归零",
			prev: &card.RoomState{
				TotalPeople:   1,
				LastEnterTime: 1_700_000_000_000,
				StaySec:       300, // prev 残留
			},
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState: ZoneState{
					Count:      0,
					LastExitTs: 1_700_000_120_000,
					UpdatedAt:  1_700_000_120_000,
				},
			},
			wantTotal:    0,
			wantStaySec:  0,
			wantLastExit: 1_700_000_120_000,
		},
		{
			name: "count_change 2→0 — 退到空房，StaySec 归零，LastExitTime 设",
			prev: &card.RoomState{
				TotalPeople:   2,
				LastEnterTime: 1_700_000_000_000,
				StaySec:       180,
			},
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 2},
				NewState: ZoneState{
					Count:     0,
					UpdatedAt: 1_700_000_200_000,
				},
			},
			wantTotal:    0,
			wantStaySec:  0,
			wantLastExit: 1_700_000_200_000,
		},
		{
			name: "钟漂保护：UpdatedAt < LastEnterTime → StaySec=0 而非负数",
			prev: &card.RoomState{
				TotalPeople:   1,
				LastEnterTime: 1_700_000_100_000,
			},
			event: ZoneEvent{
				Transition: TransitionLeaving,
				NewState: ZoneState{
					Count:     1,
					UpdatedAt: 1_700_000_050_000, // 早于 enter
				},
			},
			wantTotal:   1,
			wantStaySec: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateRoomState(tc.event, tc.prev, card.RoomTypeDefault)
			if out.TotalPeople != tc.wantTotal {
				t.Errorf("TotalPeople = %d, want %d", out.TotalPeople, tc.wantTotal)
			}
			if out.StaySec != tc.wantStaySec {
				t.Errorf("StaySec = %d, want %d", out.StaySec, tc.wantStaySec)
			}
			if tc.wantLastEnter > 0 && out.LastEnterTime != tc.wantLastEnter {
				t.Errorf("LastEnterTime = %d, want %d", out.LastEnterTime, tc.wantLastEnter)
			}
			if tc.wantLastExit > 0 && out.LastExitTime != tc.wantLastExit {
				t.Errorf("LastExitTime = %d, want %d", out.LastExitTime, tc.wantLastExit)
			}
		})
	}
}

func TestTranslateRoomState_RoomType(t *testing.T) {
	prev := &card.RoomState{RoomType: card.RoomTypeBathroom}

	// 已是 bathroom，新事件传 Default → 保留 bathroom（roomType==Default 时不覆盖）
	out := TranslateRoomState(
		ZoneEvent{NewState: ZoneState{Count: 1, UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000}, Transition: TransitionOccupied},
		prev,
		card.RoomTypeDefault,
	)
	if out.RoomType != card.RoomTypeBathroom {
		t.Errorf("RoomType reset, got %d want %d (bathroom)", out.RoomType, card.RoomTypeBathroom)
	}

	// 显式传 Bathroom → 设
	out2 := TranslateRoomState(
		ZoneEvent{NewState: ZoneState{Count: 1, UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000}, Transition: TransitionOccupied},
		nil,
		card.RoomTypeBathroom,
	)
	if out2.RoomType != card.RoomTypeBathroom {
		t.Errorf("RoomType = %d, want %d", out2.RoomType, card.RoomTypeBathroom)
	}
}

func TestTranslateBedState_BedEvent(t *testing.T) {
	tests := []struct {
		name          string
		prev          *card.BedState
		event         ZoneEvent
		wantBedStatus int
		wantBedEvent  int
		wantDuration  int
	}{
		{
			name: "occupied — InBed (bed_status=0, bed_event=0)",
			prev: nil,
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState:   ZoneState{Count: 1, Status: StatusOccupied, LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000},
			},
			wantBedStatus: 0,
			wantBedEvent:  0,
			wantDuration:  0,
		},
		{
			name: "vacant — LeftBed (bed_status=1, bed_event=1)，DurationSec 算到 LastExitTs",
			prev: &card.BedState{StartTime: 1_700_000_000_000},
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState:   ZoneState{Count: 0, Status: StatusVacant, LastExitTs: 1_700_000_120_000, UpdatedAt: 1_700_000_120_000},
			},
			wantBedStatus: 1,
			wantBedEvent:  1,
			wantDuration:  120,
		},
		{
			name: "leaving — bed_status=0（人还在），bed_event=8（None），duration 算到 UpdatedAt",
			prev: &card.BedState{StartTime: 1_700_000_000_000},
			event: ZoneEvent{
				Transition: TransitionLeaving,
				NewState:   ZoneState{Count: 1, Status: StatusLeaving, UpdatedAt: 1_700_000_045_000},
			},
			wantBedStatus: 0,
			wantBedEvent:  8,
			wantDuration:  45,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateBedState(tc.event, tc.prev)
			if out.BedStatus != tc.wantBedStatus {
				t.Errorf("BedStatus = %d, want %d", out.BedStatus, tc.wantBedStatus)
			}
			if out.BedEvent != tc.wantBedEvent {
				t.Errorf("BedEvent = %d, want %d", out.BedEvent, tc.wantBedEvent)
			}
			if out.DurationSec != tc.wantDuration {
				t.Errorf("DurationSec = %d, want %d", out.DurationSec, tc.wantDuration)
			}
		})
	}
}
