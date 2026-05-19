// card_display_builder.go — 从 CardStatus 派生 CardDisplay（纯函数，无 IO）。
//
// 设计：cardagg 是 card:state.display 的唯一 writer（owl-common rule #1.3）。
// 本 builder 处理 /88 /96 自身视角；/80 unit picker 见 unit_picker.go。
//
// 详 owlBack/owl-common/card/card_types.go（CardDisplay struct + enum 定义）+
// owlBack/doc/card_display.md。

package consumer

import (
	"time"

	"owl-common/card"
)

// BuildCardDisplay 从 CardStatus 派生 CardDisplay（自身视角，不含 /80 unit picker）。
// 调用方：SensorStateProjector 收到 /88 room 或 /96 bed 事件后立即调用。
//
// Section1.DownLeft = ""（/88 /96 卡无 room_label 自显需；/80 unit 卡 picker 时另填）
// Section1.DownRight = ""（alarmEvent 由 AlarmRouter 路径填）
// Section2.left mode picker：
//   - no bed_state + no room_state → mode=None
//   - room_state.risk_level > 0 → RoomStatus（Risk 优先）
//   - bed_state.bed_status == 0 → SleepStage
//   - otherwise → recency pick by updated_at
func BuildCardDisplay(s *card.CardStatus) *card.CardDisplay {
	if s == nil {
		return nil
	}
	now := time.Now().UnixMilli()
	d := &card.CardDisplay{UpdatedAt: now}

	bedHas := s.BedState != nil && s.BedState.UpdatedAt > 0
	roomHas := s.RoomState != nil && s.RoomState.UpdatedAt > 0

	// ===== Section2.left =====
	d.Section2LeftMode = pickSection2Left(s, bedHas, roomHas)
	switch d.Section2LeftMode {
	case card.Section2LeftModeSleepStage:
		d.SleepStage = s.BedState.SleepStage
	case card.Section2LeftModeRoomStatus:
		d.RoomPersonCount = s.RoomState.TotalPeople
		if s.RoomState.RoomType == card.RoomTypeBathroom {
			d.RoomIconKind = card.RoomIconKindBathroom
		} else {
			d.RoomIconKind = card.RoomIconKindRoom
		}
		d.RoomRiskLevel = s.RoomState.RiskLevel
		// RoomName 由调用方根据 spatial_prefix 反查（builder 无 DB access）；
		// 本 builder 留空，由 SensorStateProjector 注入。
	}

	// ===== Section3.up.left ActiveState =====
	// /88 /96 卡的 active_anchor_ms = max(bed_state.updated_at, room_state.updated_at)。
	// /80 unit picker 时另算（取 active child 的 anchor）。
	d.ActiveAnchorMs = maxUpdatedAt(s)
	if d.ActiveAnchorMs > 0 && now-d.ActiveAnchorMs < 60_000 {
		d.ActiveState = card.ActiveStateNow
	} else {
		d.ActiveState = card.ActiveStateInactive
	}

	// ===== Section3.up.right SceneState =====
	d.SceneState, d.SceneAnchorMs = pickScene(s, bedHas, roomHas)

	// ===== Section3.down VisitorState =====
	d.VisitorState, d.VisitorAnchorMs = pickVisitor(s)

	return d
}

// pickSection2Left 选 Section2.left 显示模式。Risk 优先 > 最新优先。
func pickSection2Left(s *card.CardStatus, bedHas, roomHas bool) int {
	if !bedHas && !roomHas {
		return card.Section2LeftModeNone
	}
	// RoomState Risk > 0 → RoomStatus（即使床上有人）
	if roomHas && s.RoomState.RiskLevel > 0 && s.RoomState.TotalPeople > 0 {
		return card.Section2LeftModeRoomStatus
	}
	// 在床（bedStatus=0）且 room 无 risk → SleepStage
	if bedHas && s.BedState.BedStatus == 0 {
		return card.Section2LeftModeSleepStage
	}
	// 房间有人 → RoomStatus
	if roomHas && s.RoomState.TotalPeople > 0 {
		return card.Section2LeftModeRoomStatus
	}
	// 都无人 → 按 recency 选
	if bedHas && roomHas {
		if s.BedState.UpdatedAt >= s.RoomState.UpdatedAt {
			return card.Section2LeftModeSleepStage
		}
		return card.Section2LeftModeRoomStatus
	}
	if bedHas {
		return card.Section2LeftModeSleepStage
	}
	return card.Section2LeftModeRoomStatus
}

// pickScene 选 Section3.up.right SceneState + anchor。
//
//	bathroom 有人 → InBath, anchor = last_enter_time
//	room 有人 + bed 在床 → InBed, anchor = bed start_time
//	room 有人 + bed 离床 → InRoom, anchor = last_enter_time
//	bed 离床 (无 room state) → OOB, anchor = bed start_time
//	room 无人 + last_exit_to_outside → OOU, anchor = last_exit_time
//	room 无人 → OOR, anchor = last_exit_time
//	无信号 → OOR, anchor = 0
func pickScene(s *card.CardStatus, bedHas, roomHas bool) (int, int64) {
	if roomHas && s.RoomState.TotalPeople > 0 {
		if s.RoomState.RoomType == card.RoomTypeBathroom {
			return card.SceneStateInBath, s.RoomState.LastEnterTime
		}
		if bedHas && s.BedState.BedStatus == 0 {
			return card.SceneStateInBed, s.BedState.StartTime
		}
		return card.SceneStateInRoom, s.RoomState.LastEnterTime
	}
	if bedHas && s.BedState.BedStatus == 1 {
		return card.SceneStateOOB, s.BedState.StartTime
	}
	if roomHas {
		return card.SceneStateOOR, s.RoomState.LastExitTime
	}
	return card.SceneStateOOR, 0
}

// pickVisitor 从 TargetState 派生 visitor 显示。
func pickVisitor(s *card.CardStatus) (int, int64) {
	if s.Target == nil {
		return card.VisitorStateNone, 0
	}
	if s.Target.VisitorStartTs > 0 {
		// 2h 内 → "Visitor now"；之外仍 today
		nowMs := time.Now().UnixMilli()
		if nowMs-s.Target.VisitorStartTs < 2*3600_000 {
			return card.VisitorStateNow, s.Target.VisitorStartTs
		}
		return card.VisitorStateToday, s.Target.VisitorStartTs
	}
	if s.Target.HasVisitorToday {
		return card.VisitorStateToday, 0
	}
	return card.VisitorStateNone, 0
}

func maxUpdatedAt(s *card.CardStatus) int64 {
	var m int64
	if s.BedState != nil && s.BedState.UpdatedAt > m {
		m = s.BedState.UpdatedAt
	}
	if s.RoomState != nil && s.RoomState.UpdatedAt > m {
		m = s.RoomState.UpdatedAt
	}
	if s.AlarmState != nil && s.AlarmState.UpdatedAt > m {
		m = s.AlarmState.UpdatedAt
	}
	if s.Target != nil && s.Target.UpdatedAt > m {
		m = s.Target.UpdatedAt
	}
	return m
}
