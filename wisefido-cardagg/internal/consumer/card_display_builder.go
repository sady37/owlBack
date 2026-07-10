// card_display_builder.go — 派生 CardDisplay：cardagg 完成全部业务派生，
// FE 单 switch 渲 icon + 直显 number/text/risk，不需要业务判断。
//
// Section1.down.left = 空间 scope 标签（"GuestRoom" / "Whole Unit" 等）
// Section2.left = (icon, number, text, risk) 四元组，FE dumb renderer
// Section3 = scene/visitor/vital，FE 本地计时（anchor 同步推）

package consumer

import (
	"fmt"
	"net/netip"
	"time"

	"owl-common/card"
	"wisefido-cardagg/internal/service"
)

// BuildCardDisplay 从 CardStatus + SpatialCache 原始数据派生 CardDisplay。
func BuildCardDisplay(s *card.CardStatus, cache *service.SpatialCache) *card.CardDisplay {
	if s == nil {
		return nil
	}
	now := time.Now().UnixMilli()
	d := &card.CardDisplay{UpdatedAt: now}

	bedHas := s.BedState != nil && s.BedState.MaxTs() > 0
	roomHas := s.RoomState != nil && s.RoomState.MaxTs() > 0

	// 卡 prefix + cache 反查
	var cardPfx netip.Prefix
	if pfx, err := netip.ParsePrefix(s.CardID); err == nil {
		cardPfx = pfx
	}
	hasBedDevice := cache != nil && cardPfx.IsValid() && cache.HasBed(cardPfx)
	isBathroom := false
	if cache != nil && roomHas && s.RoomState.RoomID != "" {
		if roomPfx, err := netip.ParsePrefix(s.RoomState.RoomID); err == nil {
			if rm := cache.LookupRoom(roomPfx); rm != nil && rm.RoomType == card.RoomTypeBathroom {
				isBathroom = true
			}
		}
	}

	// Source 决定 bed 还是 room latest
	bedTs := int64(0)
	roomTs := int64(0)
	if bedHas {
		bedTs = s.BedState.MaxTs()
	}
	if roomHas {
		roomTs = s.RoomState.MaxTs()
	}
	bedSource := bedTs > 0 && bedTs >= roomTs

	// ===== Section1.down.left — 空间 scope 标签 =====
	d.Section1DownLeft = pickSection1DownLeft(s, cache, cardPfx, bedSource, hasBedDevice)

	// ===== Section2.left =====
	d.Section2LeftIcon = pickSection2LeftIcon(s, bedHas, roomHas, hasBedDevice, isBathroom, bedSource)
	// BedInUse 时 up/down 都不显（隐私 + icon 已表达；Up 房间名/人数、Down 站立/独处时长都不在床场景用）
	if d.Section2LeftIcon != card.Section2IconBedInUse {
		d.Section2LeftUp = pickSection2LeftUp(s, cache, d.Section2LeftIcon)
		d.Section2LeftDown = pickSection2LeftDown(s, cache)
	}
	if roomHas {
		d.Section2LeftRisk = s.RoomState.RiskLevel
	}

	// ===== Section3.up.left ActiveState =====
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

	// ===== Section3.down.left VisitorState + BedAnchorMs =====
	d.VisitorState, d.VisitorAnchorMs = pickVisitor(s)
	if bedHas {
		d.BedAnchorMs = s.BedState.BedStatusTs
	}

	// ===== Section3.down.right VitalTrendLevel =====
	d.VitalTrendLevel = pickVitalTrendLevel(s)

	return d
}

// pickSection1DownLeft 空间 scope 标签。
// bed source → "" (床场景隐私 + icon 已含 BedInUse 语义，不重复标 BedA)。
// room source → cache.entries[RoomID].Name。
// 全空 → 卡 mask fallback：/80 "Whole Unit" / /88 room name / /96 bed name。
func pickSection1DownLeft(s *card.CardStatus, cache *service.SpatialCache, cardPfx netip.Prefix, bedSource, hasBedDevice bool) string {
	if cache == nil {
		return ""
	}
	// bed source：床场景不显空间名（隐私 + icon 已表达）
	if bedSource {
		return ""
	}
	// room source：取 RoomState.RoomID 对应 entry
	if !bedSource && s.RoomState != nil && s.RoomState.RoomID != "" {
		if rp, err := netip.ParsePrefix(s.RoomState.RoomID); err == nil {
			if ent := cache.LookupEntry(rp); ent != nil {
				return ent.Name
			}
		}
	}
	// fallback by mask
	if !cardPfx.IsValid() {
		return ""
	}
	switch cardPfx.Bits() {
	case 80:
		return "Whole Unit"
	case 88:
		if ent := cache.LookupEntry(cardPfx); ent != nil {
			return ent.Name
		}
	case 96:
		if ent := cache.LookupEntry(cardPfx); ent != nil {
			return ent.Name
		}
	}
	return ""
}

// pickSection2LeftIcon — FE 单 switch 决定 base icon；(Source, RoomType, Risk) 三维度合成。
// 床 sub-variant 由 FE 自家从 BedState/monitor 派生，不在此处细分。
//
// icon 数值序即优先级：床是 floor（有床设备/床源恒保 BedInUse 不退 Empty），
// room/bathroom 按 (risk_tier, 卫生间>房间) 取大盖过床。bedSource 不抢先压掉 room——
// 卫生间占用(≥BathroomNormal 3)必然盖过 BedInUse(2)，risk 档再逐级压过 Normal。
func pickSection2LeftIcon(s *card.CardStatus, bedHas, roomHas, hasBedDevice, isBathroom, bedSource bool) int {
	icon := card.Section2IconEmpty
	if hasBedDevice || bedSource {
		icon = card.Section2IconBedInUse
	}
	if roomHas {
		if ri := roomIconFor(isBathroom, s.RoomState.RiskLevel, s.RoomState.TotalPeople); ri > icon {
			icon = ri
		}
	}
	return icon
}

// roomIsBathroom 解析 roomID prefix 反查 cache room type == Bathroom。
// DisplayRebuilder 选"代表房" 与此处判 isBathroom 共用，单源不 drift。
func roomIsBathroom(cache *service.SpatialCache, roomID string) bool {
	if cache == nil || roomID == "" {
		return false
	}
	roomPfx, err := netip.ParsePrefix(roomID)
	if err != nil {
		return false
	}
	rm := cache.LookupRoom(roomPfx)
	return rm != nil && rm.RoomType == card.RoomTypeBathroom
}

// roomIconFor 单房 → section2 icon 值（值序即优先级；count==0 → Empty）。
// DisplayRebuilder 遍历选"代表房" 与 display 出图标共用此函数，优先级定义单源。
func roomIconFor(isBathroom bool, risk, count int) int {
	if count <= 0 {
		return card.Section2IconEmpty
	}
	switch {
	case risk >= card.RiskRisk:
		if isBathroom {
			return card.Section2IconBathroomRisk
		}
		return card.Section2IconRoomRisk
	case risk == card.RiskAttention:
		if isBathroom {
			return card.Section2IconBathroomAttention
		}
		return card.Section2IconRoomAttention
	default:
		if isBathroom {
			return card.Section2IconBathroomNormal
		}
		return card.Section2IconRoomNormal
	}
}

// pickSection2LeftDown icon 底部时长 metric — 跟 RiskLevel 来源同源：
//   - standing 超阈值 → "Stand Xm"
//   - bathroom 独居超阈（sensor 已写 RiskLevel）→ "Alone Xh Ym"
//   - 都未触发 → ""
//
// **单源真相收敛**：sensor 60s tick 算好 MIN + RiskLevel 直接 push；cardagg 端 display 只
// switch RiskLevel 选模板 + 拼数值，不做任何阈值判定也不读 ts（FE 不依赖 client 时钟）。
func pickSection2LeftDown(s *card.CardStatus, cache *service.SpatialCache) string {
	if s.RoomState == nil || s.RoomState.TotalPeople == 0 {
		return ""
	}
	risk := s.RoomState.RiskLevel
	if risk != card.RiskAttention && risk != card.RiskRisk {
		return ""
	}
	if s.RoomState.StandingContinuousMin > 0 {
		return fmt.Sprintf("Stand %dm", s.RoomState.StandingContinuousMin)
	}
	if s.RoomState.AloneContinuousMin > 0 {
		return formatAloneText(s.RoomState.AloneContinuousMin)
	}
	return ""
}

// formatAloneText "Alone Xh Ym" / "Alone Xm" 紧凑显示。
func formatAloneText(min int) string {
	if min < 60 {
		return fmt.Sprintf("Alone %dm", min)
	}
	h := min / 60
	m := min % 60
	if m == 0 {
		return fmt.Sprintf("Alone %dh", h)
	}
	return fmt.Sprintf("Alone %dh %dm", h, m)
}

// pickSection2LeftUp icon 顶部小标签 — 用 icon 单 switch 决定：
//   - icon=0 Empty / icon=2 BedInUse → "" (空 / 床场景隐私)
//   - icon ∈ {1,3,4,5,6,7} (room/bathroom 各档) → "RoomName" 或 "RoomName:XP"
//     · TotalPeople > 1 → "RoomName:XP"
//     · TotalPeople == 1 → "RoomName"
//
// RoomName 从 cache.LookupEntry(RoomState.RoomID).Name 拿。
func pickSection2LeftUp(s *card.CardStatus, cache *service.SpatialCache, icon int) string {
	if icon == card.Section2IconEmpty || icon == card.Section2IconBedInUse {
		return ""
	}
	if s.RoomState == nil || s.RoomState.RoomID == "" || cache == nil {
		return ""
	}
	rp, err := netip.ParsePrefix(s.RoomState.RoomID)
	if err != nil {
		return ""
	}
	ent := cache.LookupEntry(rp)
	if ent == nil || ent.Name == "" {
		return ""
	}
	if s.RoomState.TotalPeople > 1 {
		return fmt.Sprintf("%s:%dP", ent.Name, s.RoomState.TotalPeople)
	}
	return ent.Name
}

// pickVitalTrendLevel WeakBio score → 4 档配色。
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

// pickScene Section3.up.right SceneState + anchor。
func pickScene(s *card.CardStatus, bedHas, roomHas, isBathroom bool) (int, int64) {
	if roomHas && s.RoomState.TotalPeople > 0 {
		if isBathroom {
			return card.SceneStateInBath, s.RoomState.LastEnterTs
		}
		if bedHas && s.BedState.BedStatus == card.BedStatusInBed {
			return card.SceneStateInBed, s.BedState.BedStatusTs
		}
		return card.SceneStateInRoom, s.RoomState.LastEnterTs
	}
	if bedHas && s.BedState.BedStatus == card.BedStatusInBed {
		return card.SceneStateInBed, s.BedState.BedStatusTs
	}
	if bedHas && s.BedState.BedStatus == card.BedStatusNotInBed {
		return card.SceneStateOOB, s.BedState.BedStatusTs
	}
	if roomHas {
		return card.SceneStateOOR, s.RoomState.LastExitTs
	}
	return card.SceneStateOOR, 0
}

// pickVisitor Section3.down.left VisitorState + anchor。
func pickVisitor(s *card.CardStatus) (int, int64) {
	if s.Target == nil {
		return card.VisitorStateNone, 0
	}
	if s.Target.VisitorStartTs > 0 {
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
