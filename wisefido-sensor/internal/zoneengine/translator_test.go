// translator_test.go — TranslateRoomState / TranslateBedState 纯函数表驱动测试。
//
// **2026-05-20 per-field ts 重构后**：sensor 不再计算 StaySec / DurationSec
// （消费者用 now - <field>_ts 现算）。本测试仅覆盖 sensor owner 字段 + 关键 ts 选取：
//   - BedStatusTs: InBed→LastEnterTs / LeftBed→LastExitTs（state-change-anchored 起点）
//   - LastEnterTs / LastExitTs: engine NewState 直 forward
//   - AloneSinceTs: Count==1 时输出 UpdatedAt
//   - BedEvent / BedEventTs / TrackNumber / BedConfidence 与旧测试相同语义

package zoneengine

import (
	"testing"

	"owl-common/card"
)

func TestTranslateRoomState_Anchors(t *testing.T) {
	const t1 = int64(1_700_000_000_000)
	const t2 = int64(1_700_000_120_000)

	tests := []struct {
		name             string
		event            ZoneEvent
		wantTotal        int
		wantLastEnter    int64
		wantLastExit     int64
		wantAloneSinceTs int64
	}{
		{
			name: "0→1 进房 — LastEnterTs 设；AloneSinceTs = UpdatedAt (count==1)",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState:   ZoneState{Count: 1, LastEnterTs: t1, UpdatedAt: t1},
			},
			wantTotal:        1,
			wantLastEnter:    t1,
			wantAloneSinceTs: t1,
		},
		{
			name: "count_change 1→2 — count!=1，AloneSinceTs=0（独居结束）",
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 1},
				NewState:   ZoneState{Count: 2, LastEnterTs: t1, UpdatedAt: t2},
			},
			wantTotal:        2,
			wantLastEnter:    t1,
			wantAloneSinceTs: 0,
		},
		{
			name: "Vacant — LastExitTs 设；AloneSinceTs=0",
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState:   ZoneState{Count: 0, LastExitTs: t2, UpdatedAt: t2},
			},
			wantTotal:        0,
			wantLastExit:     t2,
			wantAloneSinceTs: 0,
		},
		{
			name: "count_change N→1 — count==1，AloneSinceTs 设（独居重启）",
			event: ZoneEvent{
				Transition: TransitionCountChange,
				PrevState:  ZoneState{Count: 3},
				NewState:   ZoneState{Count: 1, LastEnterTs: t1, UpdatedAt: t2},
			},
			wantTotal:        1,
			wantLastEnter:    t1,
			wantAloneSinceTs: t2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := TranslateRoomState(tc.event, card.RoomTypeDefault)
			if out.TotalPeople != tc.wantTotal {
				t.Errorf("TotalPeople = %d, want %d", out.TotalPeople, tc.wantTotal)
			}
			if tc.wantLastEnter > 0 && out.LastEnterTs != tc.wantLastEnter {
				t.Errorf("LastEnterTs = %d, want %d", out.LastEnterTs, tc.wantLastEnter)
			}
			if tc.wantLastExit > 0 && out.LastExitTs != tc.wantLastExit {
				t.Errorf("LastExitTs = %d, want %d", out.LastExitTs, tc.wantLastExit)
			}
			if out.AloneSinceTs != tc.wantAloneSinceTs {
				t.Errorf("AloneSinceTs = %d, want %d", out.AloneSinceTs, tc.wantAloneSinceTs)
			}
		})
	}
}

// RoomType 已挪静态属性（不在 RoomState）；TranslateRoomState 的 roomType 参数仅供 RiskLevel 计算。
// bathroom 单人非 night 阈值未到 → 仍 RiskNormal；该 test 验证 roomType 走到 risk 分支即可。
func TestTranslateRoomState_BathroomRoomType(t *testing.T) {
	out := TranslateRoomState(
		ZoneEvent{
			Transition: TransitionOccupied,
			NewState:   ZoneState{Count: 1, UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000},
		},
		card.RoomTypeBathroom,
	)
	if out.TotalPeople != 1 {
		t.Errorf("TotalPeople = %d, want 1", out.TotalPeople)
	}
	if out.RiskLevelTs == 0 {
		t.Errorf("RiskLevelTs should be set")
	}
}

func TestTranslateBedState_StatusAndEvent(t *testing.T) {
	const tEnter = int64(1_700_000_000_000)
	const tExit = int64(1_700_000_120_000)

	tests := []struct {
		name            string
		event           ZoneEvent
		wantBedStatus   int
		wantBedEvent    int
		wantBedStatusTs int64
	}{
		{
			name: "Occupied — bed_status=0 / bed_event=0；BedStatusTs = LastEnterTs",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState:   ZoneState{Count: 1, Status: StatusOccupied, LastEnterTs: tEnter, UpdatedAt: tEnter},
			},
			wantBedStatus:   0,
			wantBedEvent:    0,
			wantBedStatusTs: tEnter,
		},
		{
			name: "Vacant — bed_status=1 / bed_event=1；BedStatusTs = LastExitTs",
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState:   ZoneState{Count: 0, Status: StatusVacant, LastEnterTs: tEnter, LastExitTs: tExit, UpdatedAt: tExit},
			},
			wantBedStatus:   1,
			wantBedEvent:    1,
			wantBedStatusTs: tExit,
		},
		{
			name: "Leaving — bed_status=0（仍 present）/ bed_event=8；BedStatusTs = LastEnterTs",
			event: ZoneEvent{
				Transition: TransitionLeaving,
				NewState:   ZoneState{Count: 1, Status: StatusLeaving, LastEnterTs: tEnter, UpdatedAt: 1_700_000_045_000},
			},
			wantBedStatus:   0,
			wantBedEvent:    8,
			wantBedStatusTs: tEnter,
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
			if out.BedStatusTs != tc.wantBedStatusTs {
				t.Errorf("BedStatusTs = %d, want %d", out.BedStatusTs, tc.wantBedStatusTs)
			}
		})
	}
}

// TrackNumber + BedConfidence 测试（来源 forwarded）
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
				NewState:   ZoneState{Count: 1, Status: StatusOccupied, LastSource: "sleepace", UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000},
			},
			wantTrackNumber:   1,
			wantBedConfidence: 90,
		},
		{
			name: "occupied by radar: TrackNumber=1 / BedConfidence=60",
			event: ZoneEvent{
				Transition: TransitionOccupied,
				NewState:   ZoneState{Count: 1, Status: StatusOccupied, LastSource: "radar", UpdatedAt: 1_700_000_000_000, LastEnterTs: 1_700_000_000_000},
			},
			wantTrackNumber:   1,
			wantBedConfidence: 60,
		},
		{
			name: "vacant: TrackNumber=0 / BedConfidence 仍取 source",
			event: ZoneEvent{
				Transition: TransitionVacant,
				NewState:   ZoneState{Count: 0, Status: StatusVacant, LastSource: "sleepace", LastEnterTs: 1_700_000_000_000, LastExitTs: 1_700_000_120_000, UpdatedAt: 1_700_000_120_000},
			},
			wantTrackNumber:   0,
			wantBedConfidence: 90,
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
