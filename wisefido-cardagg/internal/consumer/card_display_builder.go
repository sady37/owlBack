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

// BuildCardDisplay 从 CardStatus + 卡静态属性派生 CardDisplay（自身视角，不含 /80 unit picker）。
// 调用方：SensorStateProjector / UnitPicker / AlarmRouter 收到事件后调用。
//
// hasBedDevice / isBathroom 都从 CardMeta 派生（静态属性，DB 静态查 + config:card:stream 失效）：
//   - hasBedDevice = CardMeta.HasBed()         物理上有 sleepad 床设备
//   - isBathroom   = CardMeta.IsBathroom()     room_type=1（rooms.room_type LPM /88）
func BuildCardDisplay(s *card.CardStatus, hasBedDevice bool, isBathroom bool) *card.CardDisplay {
	if s == nil {
		return nil
	}
	now := time.Now().UnixMilli()
	d := &card.CardDisplay{UpdatedAt: now}

	bedHas := s.BedState != nil && s.BedState.MaxTs() > 0
	roomHas := s.RoomState != nil && s.RoomState.MaxTs() > 0

	// ===== Section2.left =====
	// section2_left_mode 是辅助 hint；raw 字段族（BedStatus / RoomIconKind / RoomPersonCount /
	// RoomRiskLevel / StayAlarmEnabled）必须永远反映底层 state，不受 mode 门控——
	// FE 用 CardPriority 单 switch 派生 icon（spec §3.2），其余字段用作 icon 变体细化。
	d.Section2LeftMode = pickSection2Left(s, bedHas, roomHas)
	if bedHas {
		bs := s.BedState.BedStatus
		d.BedStatus = &bs // *int 跟 observation/track.go 同惯例；nil=未知/不适用
		d.SleepStage = s.BedState.SleepStage
	} else if hasBedDevice {
		// 无 fresh BedState 时默认 NotInBed：在床需要证据（Sleepace InBed event / radar vital signs），
		// 缺省视为不在床。FE 按 bed_status=1 渲 outofbed icon，避免 lying-bed-black 假状态。
		bs := card.BedStatusNotInBed
		d.BedStatus = &bs
	}
	if roomHas {
		d.RoomPersonCount = s.RoomState.TotalPeople
		if isBathroom {
			d.RoomIconKind = card.RoomIconKindBathroom
		} else {
			d.RoomIconKind = card.RoomIconKindRoom
		}
		d.RoomRiskLevel = s.RoomState.RiskLevel
		d.StayAlarmEnabled = true
	}
	d.CardPriority = pickCardPriority(d, hasBedDevice)

	// ===== Section3.up.left ActiveState =====
	// active_anchor_ms = TargetState.LastActiveTs（用户 walk/stand **真实活动** anchor）。
	// **不要** 用 bed/room state UpdatedAt 凑数（那是状态写入时刻，跟活动无关）。
	// 没有 Target 或 LastActiveTs=0 → active_anchor=0，FE 显 "—" 不显假活跃。
	if s.Target != nil && s.Target.LastActiveTs > 0 {
		d.ActiveAnchorMs = s.Target.LastActiveTs
		if now-d.ActiveAnchorMs < 60_000 {
			d.ActiveState = card.ActiveStateNow
		} else {
			d.ActiveState = card.ActiveStateInactive
		}
	}

	// ===== Section3.up.right SceneState =====
	d.SceneState, d.SceneAnchorMs = pickScene(s, bedHas, roomHas, isBathroom)

	// ===== Section3.down.left VisitorState + BedAnchorMs fallback =====
	d.VisitorState, d.VisitorAnchorMs = pickVisitor(s)
	if bedHas {
		d.BedAnchorMs = s.BedState.BedStatusTs
	}

	// ===== Section3.down.right VitalTrendLevel (W2 WeakBio 横条) =====
	d.VitalTrendLevel = pickVitalTrendLevel(s)

	return d
}

// pickCardPriority 按 spec card_display.md §3.2 优先级派生 CardPriority 标量。
// 高值优先（/80 UnitPicker 取 MAX(child)）。
//
// 输入：已填好 raw 字段族的 display，加 hasBedDevice（卡是否物理上绑了 sleepad 床设备）。
// 注：hasBedDevice 来自 CardMeta（卡静态属性），不是 bed_state 是否在 Redis 写过——
// 空床久了 sensor 不再 emit bed_state，但卡"有床"事实不变，应仍走 BedInUse 分支。
func pickCardPriority(d *card.CardDisplay, hasBedDevice bool) int {
	isBath := d.RoomIconKind == card.RoomIconKindBathroom
	count := d.RoomPersonCount
	risk := d.RoomRiskLevel

	// risk=3（红/危险）
	if count > 0 && risk >= card.RiskRisk {
		if isBath {
			return card.CardPriorityBathroomRisk
		}
		return card.CardPriorityRoomRisk
	}
	// risk=2（黄/attention）
	if count > 0 && risk == card.RiskAttention {
		if isBath {
			return card.CardPriorityBathroomAttention
		}
		return card.CardPriorityRoomAttention
	}
	// bathroom 占用，无 risk
	if isBath && count > 0 {
		return card.CardPriorityBathroomNormal
	}
	// 有床（卡上绑了 sleepad 设备）—— in/out 由 bed_status / sleep_stage 决定 icon 变体
	if hasBedDevice {
		return card.CardPriorityBedInUse
	}
	// 非 bathroom 非 bed 房间有人 + stay 监控启用
	if count > 0 && d.StayAlarmEnabled {
		return card.CardPriorityRoomNormal
	}
	return card.CardPriorityEmpty
}

// pickVitalTrendLevel 把 Target.WeakBiometricSignal score 映射到 4 档配色
// （详 [[target_state_weak_bio_signal_design]] §"阈值表"）：
//
//	0-29  → 0 None (hide)
//	30-59 → 1 Gray (Attention)
//	60-79 → 2 Yellow (Watch)
//	80-100→ 3 Red (Alert)
//
// staleness 已由 target_merger 在 merge 层过滤（offline + 30min UpdatedAt）；
// builder 信任 s.Target.WeakBiometricSignal 是 fresh 值。
func pickVitalTrendLevel(s *card.CardStatus) int {
	if s.Target == nil {
		return card.VitalTrendLevelNone
	}
	score := s.Target.WeakBiometricSignal
	switch {
	case score >= 80:
		return card.VitalTrendLevelRed
	case score >= 60:
		return card.VitalTrendLevelYellow
	case score >= 30:
		return card.VitalTrendLevelGray
	}
	return card.VitalTrendLevelNone
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
	// 都无人 → 按 recency 选（用各 state 的 max ts）
	if bedHas && roomHas {
		if s.BedState.MaxTs() >= s.RoomState.MaxTs() {
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
func pickScene(s *card.CardStatus, bedHas, roomHas, isBathroom bool) (int, int64) {
	if roomHas && s.RoomState.TotalPeople > 0 {
		if isBathroom {
			return card.SceneStateInBath, s.RoomState.LastEnterTs
		}
		if bedHas && s.BedState.BedStatus == 0 {
			return card.SceneStateInBed, s.BedState.BedStatusTs
		}
		return card.SceneStateInRoom, s.RoomState.LastEnterTs
	}
	// bed-only path（典型场景：/80 unit 卡无 /88 子卡，bed_state 直接落 /80；或 unit-level
	// 单 bed 视图）—— 仅看 BedStatus 决定 InBed/OOB
	if bedHas && s.BedState.BedStatus == 0 {
		return card.SceneStateInBed, s.BedState.BedStatusTs
	}
	if bedHas && s.BedState.BedStatus == 1 {
		return card.SceneStateOOB, s.BedState.BedStatusTs
	}
	if roomHas {
		return card.SceneStateOOR, s.RoomState.LastExitTs
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

// maxUpdatedAt 已废弃 — active_anchor_ms 现在直接读 TargetState.LastActiveTs
// （per-state UpdatedAt 不再当 active anchor，避免假活跃问题）。
// 保留函数空壳兼容老 caller；返回 0。
