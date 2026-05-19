// translator_test.go — TranslateRoomState / TranslateBedState 纯函数表驱动测试。
//
// Sensor 不读 prev（"sensor 不读 card:state"原则），signature 不再带 prev 参数。
// 测试只覆盖 sensor owner 字段：
//   - StaySec 累积公式（占用即时算 / 空房归零 / 进房瞬间=0 / 钟漂保护）
//   - LastEnterTime / LastExitTime / TotalPeople / RoomType 派生
//   - BedState BedStatus / BedEvent / StartTime / DurationSec 派生
//
// 不测：RiskLevel（由 EvaluateRoomRiskLevel 单独覆盖）；非 sensor owner 字段
// (SleepStage / TrackNumber 等)由 cardagg projector 字段级 merge 测试覆盖。

package zoneengine

import (
	"testing"

	"owl-common/card"
)

func TestTranslateRoomState_StaySec(t *testing.T) {
	tests := []struct {
		name          string
		event         ZoneEvent
		wantTotal     int
		wantStaySec   int
		wantLastEnter int64
		wantLastExit  int64
	}{
		{
			name: "occupied — 进房 LastEnterTs 重置，StaySec 同帧=0",
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
			name: "occupied 30s — StaySec = (now - enter)/1000",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState: ZoneState{
					Count:       1,
					LastEnterTs: 1_700_000_000_000,
					UpdatedAt:   1_700_000_030_000, // +30s
				},
			},
			wantTotal:     1,
			wantStaySec:   30,
			wantLastEnter: 1_700_000_000_000,
		},
		{
			name: "count_change 0→2 — LastEnterTime 设为 UpdatedAt，StaySec=0",
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
			name: "count_change 1→3 持续 60s — 沿用 engine LastEnterTs，StaySec 累计",
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 1},
				NewState: ZoneState{
					Count:       3,
					LastEnterTs: 1_700_000_000_000,
					UpdatedAt:   1_700_000_060_000,
				},
			},
			wantTotal:     3,
			wantStaySec:   60,
			wantLastEnter: 1_700_000_000_000,
		},
		{
			name: "vacant — TotalPeople=0，StaySec 归零",
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
			event: ZoneEvent{
				Transition: TransitionLeaving,
				NewState: ZoneState{
					Count:       1,
					LastEnterTs: 1_700_000_100_000,
					UpdatedAt:   1_700_000_050_000, // 早于 enter
				},
			},
			wantTotal:   1,
			wantStaySec: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateRoomState(tc.event, card.RoomTypeDefault)
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
	out := TranslateRoomState(
		ZoneEvent{
			Transition: TransitionOccupied,
			NewState:   ZoneState{Count: 1, UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000},
		},
		card.RoomTypeBathroom,
	)
	if out.RoomType != card.RoomTypeBathroom {
		t.Errorf("RoomType = %d, want %d (bathroom)", out.RoomType, card.RoomTypeBathroom)
	}

	out2 := TranslateRoomState(
		ZoneEvent{
			Transition: TransitionOccupied,
			NewState:   ZoneState{Count: 1, UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000},
		},
		card.RoomTypeDefault,
	)
	if out2.RoomType != card.RoomTypeDefault {
		t.Errorf("RoomType = %d, want %d (default)", out2.RoomType, card.RoomTypeDefault)
	}
}

func TestTranslateBedState_BedEvent(t *testing.T) {
	tests := []struct {
		name           string
		event          ZoneEvent
		wantBedStatus  int
		wantBedEvent   int
		wantStartTime  int64
		wantDuration   int
	}{
		{
			name: "occupied — InBed (bed_status=0, bed_event=0)，StartTime = LastEnterTs",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState:   ZoneState{Count: 1, Status: StatusOccupied, LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000},
			},
			wantBedStatus: 0,
			wantBedEvent:  0,
			wantStartTime: 1_700_000_000_000,
			wantDuration:  0,
		},
		{
			name: "vacant — LeftBed (bed_status=1, bed_event=1)，DurationSec = exit - enter",
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState:   ZoneState{Count: 0, Status: StatusVacant, LastEnterTs: 1_700_000_000_000, LastExitTs: 1_700_000_120_000, UpdatedAt: 1_700_000_120_000},
			},
			wantBedStatus: 1,
			wantBedEvent:  1,
			wantStartTime: 1_700_000_000_000,
			wantDuration:  120,
		},
		{
			name: "leaving — bed_status=0（人还在），bed_event=8（None），duration 算到 UpdatedAt",
			event: ZoneEvent{
				Transition: TransitionLeaving,
				NewState:   ZoneState{Count: 1, Status: StatusLeaving, LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_045_000},
			},
			wantBedStatus: 0,
			wantBedEvent:  8,
			wantStartTime: 1_700_000_000_000,
			wantDuration:  45,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateBedState(tc.event)
			if out.BedStatus != tc.wantBedStatus {
				t.Errorf("BedStatus = %d, want %d", out.BedStatus, tc.wantBedStatus)
			}
			if out.BedEvent != tc.wantBedEvent {
				t.Errorf("BedEvent = %d, want %d", out.BedEvent, tc.wantBedEvent)
			}
			if out.StartTime != tc.wantStartTime {
				t.Errorf("StartTime = %d, want %d", out.StartTime, tc.wantStartTime)
			}
			if out.DurationSec != tc.wantDuration {
				t.Errorf("DurationSec = %d, want %d", out.DurationSec, tc.wantDuration)
			}
		})
	}
}

// S5b TrackNumber + BedConfidence —— TranslateBedState 端到端取值
func TestTranslateBedState_TrackNumberAndBedConfidence(t *testing.T) {
	tests := []struct {
		name              string
		event             ZoneEvent
		wantTrackNumber   int
		wantBedConfidence int
	}{
		{
			name: "occupied by sleepace: TrackNumber=1 / BedConfidence=90",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState: ZoneState{
					Count: 1, Status: StatusOccupied,
					LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000,
					LastSource: "sleepace",
				},
			},
			wantTrackNumber:   1,
			wantBedConfidence: 90,
		},
		{
			name: "occupied by radar: TrackNumber=1 / BedConfidence=60",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState: ZoneState{
					Count: 1, Status: StatusOccupied,
					LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000,
					LastSource: "radar",
				},
			},
			wantTrackNumber:   1,
			wantBedConfidence: 60,
		},
		{
			name: "vacant: TrackNumber=0 / BedConfidence 仍由 LastSource 决定（不清零）",
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState: ZoneState{
					Count: 0, Status: StatusVacant,
					LastEnterTs: 1_700_000_000_000, LastExitTs: 1_700_000_120_000,
					UpdatedAt:  1_700_000_120_000,
					LastSource: "sleepace",
				},
			},
			wantTrackNumber:   0,
			wantBedConfidence: 90,
		},
		{
			name: "unknown source: BedConfidence=0 (兜底)",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState: ZoneState{
					Count: 1, Status: StatusOccupied,
					LastEnterTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000,
					LastSource: "polygon",
				},
			},
			wantTrackNumber:   1,
			wantBedConfidence: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateBedState(tc.event)
			if out.TrackNumber != tc.wantTrackNumber {
				t.Errorf("TrackNumber = %d, want %d", out.TrackNumber, tc.wantTrackNumber)
			}
			if out.BedConfidence != tc.wantBedConfidence {
				t.Errorf("BedConfidence = %d, want %d", out.BedConfidence, tc.wantBedConfidence)
			}
		})
	}
}

func TestBedConfidenceForSource(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"sleepace", 90},
		{"radar", 60},
		{"polygon", 0},
		{"", 0},
		{"unknown", 0},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := bedConfidenceForSource(tc.in); got != tc.want {
				t.Errorf("bedConfidenceForSource(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
