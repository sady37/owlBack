package zonealarm

import (
	"context"
	"sync"
	"testing"
	"time"

	"owl-common/alarm"
	"wisefido-sensor-v1/internal/zoneengine"
)

// captureFirer 收集 Arm/Cancel/Fire 事件供断言。
type captureFirer struct {
	mu      sync.Mutex
	arms    []Pending
	cancels []PendingKey
	fires   []Pending
}

func (c *captureFirer) Arm(_ context.Context, p Pending) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.arms = append(c.arms, p)
	return nil
}
func (c *captureFirer) Cancel(_ context.Context, k PendingKey, _ zoneengine.ZoneEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancels = append(c.cancels, k)
	return nil
}
func (c *captureFirer) Fire(_ context.Context, p Pending) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fires = append(c.fires, p)
	return nil
}
func (c *captureFirer) snapArms() []Pending {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Pending, len(c.arms))
	copy(out, c.arms)
	return out
}
func (c *captureFirer) snapCancels() []PendingKey {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]PendingKey, len(c.cancels))
	copy(out, c.cancels)
	return out
}
func (c *captureFirer) snapFires() []Pending {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Pending, len(c.fires))
	copy(out, c.fires)
	return out
}

// helper：构造 ZoneEvent，省略字段填默认。
func mkEv(zt zoneengine.ZoneType, status zoneengine.ZoneStatus, transition string, cardID, zoneID string, count int, ts int64) zoneengine.ZoneEvent {
	return zoneengine.ZoneEvent{
		CardID:     cardID,
		ZoneType:   zt,
		ZoneID:     zoneID,
		Transition: transition,
		NewState: zoneengine.ZoneState{
			Status: status,
			Count:  count,
		},
		Ts: ts,
	}
}

// 1. Stay — bathroom occupied → arm; bathroom→vacant 之前 fire
func TestSupervisor_StayArmAndFire(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	// arm: bathroom occupied
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now))
	if len(cap.snapArms()) != 1 {
		t.Fatalf("want 1 arm, got %v", cap.snapArms())
	}
	if cap.snapArms()[0].Key.AlarmType != alarm.Stay {
		t.Errorf("want alarm.Stay, got %s", cap.snapArms()[0].Key.AlarmType)
	}

	// 推 Tick 至 due 之后 → fire
	s.Tick(now + 600_000 + 1)
	if len(cap.snapFires()) != 1 {
		t.Fatalf("want 1 fire after due, got %v", cap.snapFires())
	}
}

// 2. Stay cancel — bathroom→vacant 后 fire 不应再发生
func TestSupervisor_StayCancelByBathroomVacant(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now))
	// cancel before due
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/88", 0, now+100))

	if len(cap.snapCancels()) != 1 {
		t.Fatalf("want 1 cancel, got %v", cap.snapCancels())
	}
	s.Tick(now + 600_000 + 1)
	if len(cap.snapFires()) != 0 {
		t.Fatalf("should not fire after cancel, got %v", cap.snapFires())
	}
	if s.PendingCount() != 0 {
		t.Errorf("pending should be empty, got %d", s.PendingCount())
	}
}

// 3. Stay cancel by InBed (peer-zone)
func TestSupervisor_StayCancelByBedOccupied(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now))
	// InBed (peer-zone) cancels Stay
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/96", 1, now+1000))

	if len(cap.snapCancels()) == 0 {
		t.Fatalf("Stay should be canceled by InBed")
	}
}

// 4. LeftBed — bed→vacant arm；rest window 默认全天，应 arm
func TestSupervisor_LeftBedArmFire(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now))
	if !hasArm(cap.snapArms(), alarm.LeftBed) {
		t.Fatalf("LeftBed should arm on bed→vacant, got %v", cap.snapArms())
	}
	s.Tick(now + 1800_000 + 1)
	if !hasFire(cap.snapFires(), alarm.LeftBed) {
		t.Fatalf("LeftBed should fire after 30min, got %v", cap.snapFires())
	}
}

// 5. LeftBed cancel by EnterRoom (人回房算回床区域，用户拍板)
func TestSupervisor_LeftBedCancelByEnterRoom(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now))
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now+1000))

	if !hasCancel(cap.snapCancels(), alarm.LeftBed) {
		t.Fatalf("LeftBed should cancel by EnterRoom, got %v", cap.snapCancels())
	}
}

// 6. NightAbsence — TimeWindow 21-7 才 arm（白天不 arm）
func TestSupervisor_NightAbsenceWindowArmingOnly(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)

	// 设固定时间 = 中午 12:00 local time
	noon := time.Date(2026, 5, 14, 12, 0, 0, 0, time.Local).UnixMilli()
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/88", 0, noon))
	// 应 arm: LeftBed 不应（bed 没翻），但 Room→Vacant 中午不 arm Night
	for _, p := range cap.snapArms() {
		if p.Key.AlarmType == alarm.NightAbsence {
			t.Fatalf("NightAbsence should NOT arm at noon (outside 21-7 window)")
		}
	}

	// 凌晨 03:00 应 arm
	earlyMorning := time.Date(2026, 5, 14, 3, 0, 0, 0, time.Local).UnixMilli()
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-2", "fd00::/88", 0, earlyMorning))
	if !hasArmFor(cap.snapArms(), alarm.NightAbsence, "card-2") {
		t.Fatalf("NightAbsence should arm at 03:00 for card-2, got %v", cap.snapArms())
	}
}

// 7. NightAbsence cancel by NumberPeople>0 (count_gt_zero)
func TestSupervisor_NightAbsenceCancelByNumberPeople(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	earlyMorning := time.Date(2026, 5, 14, 3, 0, 0, 0, time.Local).UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/88", 0, earlyMorning))
	if !hasArmFor(cap.snapArms(), alarm.NightAbsence, "card-1") {
		t.Fatalf("precondition: NightAbsence should arm, got %v", cap.snapArms())
	}

	// NumberPeople=2 (count_change with count > 0) → cancel
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusVacant,
		zoneengine.TransitionCountChange, "card-1", "fd00::/88", 2, earlyMorning+5000))

	if !hasCancel(cap.snapCancels(), alarm.NightAbsence) {
		t.Fatalf("NightAbsence should cancel by NumberPeople>0, got cancels=%v", cap.snapCancels())
	}
}

// 8. BedNightAbsence — bed→vacant 凌晨 arm；InBed cancel
func TestSupervisor_BedNightAbsenceArmCancel(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	earlyMorning := time.Date(2026, 5, 14, 3, 0, 0, 0, time.Local).UnixMilli()

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, earlyMorning))
	if !hasArmFor(cap.snapArms(), alarm.BedNightAbsence, "card-1") {
		t.Fatalf("BedNightAbsence should arm 03:00, got %v", cap.snapArms())
	}
	// 也应同时 arm LeftBed（同一 ZoneEvent 命中两规则）
	if !hasArmFor(cap.snapArms(), alarm.LeftBed, "card-1") {
		t.Fatalf("LeftBed should also arm same event, got %v", cap.snapArms())
	}

	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/96", 1, earlyMorning+1000))
	// 两个 alarm 都应 cancel
	if !hasCancel(cap.snapCancels(), alarm.BedNightAbsence) {
		t.Fatalf("BedNightAbsence not canceled by InBed, got %v", cap.snapCancels())
	}
	if !hasCancel(cap.snapCancels(), alarm.LeftBed) {
		t.Fatalf("LeftBed not canceled by InBed, got %v", cap.snapCancels())
	}
}

// 9. TimeWindow.Active 跨午夜
func TestTimeWindow_Active(t *testing.T) {
	w := &TimeWindow{StartH: 21, StartM: 0, EndH: 7, EndM: 0}
	cases := []struct {
		h, m int
		want bool
	}{
		{12, 0, false}, // noon
		{20, 59, false}, // 一分钟前
		{21, 0, true},  // 启动
		{23, 30, true},
		{3, 0, true},
		{6, 59, true},
		{7, 0, false}, // 截止
		{8, 0, false},
	}
	for _, c := range cases {
		ts := time.Date(2026, 5, 14, c.h, c.m, 0, 0, time.Local)
		got := w.Active(ts)
		if got != c.want {
			t.Errorf("TimeWindow.Active(%02d:%02d): want %v, got %v", c.h, c.m, c.want, got)
		}
	}
}

// 10. nil window = 全天 active
func TestTimeWindow_NilAllDay(t *testing.T) {
	var w *TimeWindow
	if !w.Active(time.Now()) {
		t.Errorf("nil window should be all-day active")
	}
}

// 11. count_change 不参与 arm（仅 cancel via CountGtZero）
func TestSupervisor_CountChangeNotArm(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	earlyMorning := time.Date(2026, 5, 14, 3, 0, 0, 0, time.Local).UnixMilli()

	// count_change 即使 status=Vacant 也不应触发 NightAbsence arm
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeRoom, zoneengine.StatusVacant,
		zoneengine.TransitionCountChange, "card-1", "fd00::/88", 0, earlyMorning))
	for _, p := range cap.snapArms() {
		if p.Key.AlarmType == alarm.NightAbsence {
			t.Errorf("NightAbsence should NOT arm on count_change, got %v", p)
		}
	}
}

// 12. fire-once-until-zone-leave — Stay 已 fire 后，bathroom 仍 Occupied 期间不再 arm
func TestSupervisor_StayFireOnceUntilBathroomVacant(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	// arm + fire (10min later)
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now))
	s.Tick(now + 600_000 + 1)
	if len(cap.snapFires()) != 1 {
		t.Fatalf("want 1 fire, got %d", len(cap.snapFires()))
	}

	// bathroom 仍 Occupied — 再来一条 Occupied 事件（zone re-trigger），不应再 arm
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now+700_000))
	if len(cap.snapArms()) != 1 {
		t.Fatalf("should not re-arm while fired, got %d arms", len(cap.snapArms()))
	}

	// bathroom→Leaving（软离开）也不应清 fired
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusLeaving,
		zoneengine.TransitionVacant, "card-1", "fd00::/88", 0, now+800_000))
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now+810_000))
	if len(cap.snapArms()) != 1 {
		t.Fatalf("Leaving→Occupied bounce should NOT re-arm (fired still set), got %d arms",
			len(cap.snapArms()))
	}

	// bathroom→Vacant — 真正离开，清 fired
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/88", 0, now+1_200_000))

	// 重新进入 bathroom → arm OK
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now+1_300_000))
	if len(cap.snapArms()) != 2 {
		t.Fatalf("after Vacant clear should re-arm, got %d arms", len(cap.snapArms()))
	}
}

// 13. fire-once-until-zone-leave — LeftBed (arm=Vacant) 已 fire 后，仅 bed→Occupied (完成确认) 才清 fired
func TestSupervisor_LeftBedFireOnceClearedOnlyByBedOccupied(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	// arm + fire (30min later)
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now))
	s.Tick(now + 1800_000 + 1)
	if !hasFire(cap.snapFires(), alarm.LeftBed) {
		t.Fatalf("LeftBed should fire after 30min")
	}

	// bed→Vacant 再触发不应 arm (fired set)
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now+1900_000))
	armCount := func() int {
		n := 0
		for _, p := range cap.snapArms() {
			if p.Key.AlarmType == alarm.LeftBed {
				n++
			}
		}
		return n
	}
	if armCount() != 1 {
		t.Fatalf("LeftBed should not re-arm while fired, got %d arms", armCount())
	}

	// bed→Leaving (未完成确认) → 不清 fired
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusLeaving,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 1, now+2000_000))
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now+2050_000))
	if armCount() != 1 {
		t.Fatalf("bed→Leaving should NOT clear fired (incomplete), got %d arms", armCount())
	}

	// bed→Occupied (完成确认人回床) → 清 fired
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/96", 1, now+2200_000))

	// 再次离床 → 可以重新 arm
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusVacant,
		zoneengine.TransitionVacant, "card-1", "fd00::/96", 0, now+2300_000))
	if armCount() != 2 {
		t.Fatalf("LeftBed should re-arm after bed→Occupied clears fired, got %d arms", armCount())
	}
}

// 14. fire-once — cancel 不该清 fired（peer-zone cancel 时 arm zone 还在 arm state）
func TestSupervisor_StayCancelByPeerZoneDoesNotClearFired(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	now := time.Now().UnixMilli()

	// arm Stay + fire
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now))
	s.Tick(now + 600_000 + 1)
	if len(cap.snapFires()) != 1 {
		t.Fatalf("expect 1 fire")
	}

	// bed→Occupied (peer-zone cancel trigger)，但 bathroom 还在 Occupied
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBed, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/96", 1, now+700_000))

	// bathroom 再触发 Occupied → 不应该 re-arm（fired 仍标记，因 bathroom 没真离开过）
	s.OnZoneEvent(mkEv(zoneengine.ZoneTypeBathroom, zoneengine.StatusOccupied,
		zoneengine.TransitionOccupied, "card-1", "fd00::/88", 1, now+800_000))
	if len(cap.snapArms()) != 1 {
		t.Fatalf("peer-zone cancel should NOT clear fired; got %d arms (want 1)", len(cap.snapArms()))
	}
}

// 15. ReloadRules 替换规则集
func TestSupervisor_ReloadRules(t *testing.T) {
	cap := &captureFirer{}
	s := NewSupervisor(DefaultRules(), cap, nil)
	if len(s.rules) != 4 {
		t.Fatalf("default rules should be 4, got %d", len(s.rules))
	}
	s.ReloadRules([]Rule{}) // empty
	if len(s.rules) != 0 {
		t.Errorf("after reload should be 0")
	}
}

// helper assertions
func hasArm(arms []Pending, alarmType string) bool {
	for _, p := range arms {
		if p.Key.AlarmType == alarmType {
			return true
		}
	}
	return false
}

func hasArmFor(arms []Pending, alarmType, cardID string) bool {
	for _, p := range arms {
		if p.Key.AlarmType == alarmType && p.Key.CardID == cardID {
			return true
		}
	}
	return false
}

func hasCancel(cancels []PendingKey, alarmType string) bool {
	for _, k := range cancels {
		if k.AlarmType == alarmType {
			return true
		}
	}
	return false
}

func hasFire(fires []Pending, alarmType string) bool {
	for _, p := range fires {
		if p.Key.AlarmType == alarmType {
			return true
		}
	}
	return false
}
