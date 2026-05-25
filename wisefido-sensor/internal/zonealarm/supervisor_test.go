package zonealarm

import (
	"context"
	"sync"
	"testing"
	"time"

	"owl-common/alarm"
	"owl-common/card"
)

// captureFirer 收集 Fire 事件供断言。
type captureFirer struct {
	mu    sync.Mutex
	fires []FireEvent
}

func (c *captureFirer) Fire(_ context.Context, e FireEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fires = append(c.fires, e)
	return nil
}

func (c *captureFirer) snapFires() []FireEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]FireEvent, len(c.fires))
	copy(out, c.fires)
	return out
}

func hasFireFor(events []FireEvent, alarmType, zoneID string) bool {
	for _, e := range events {
		if e.Key.AlarmType == alarmType && e.Key.ZoneID == zoneID {
			return true
		}
	}
	return false
}

// captureProvider 测试用 StateProvider stub。
type captureProvider struct {
	rooms map[string]RoomEntry
	beds  map[string]*card.BedState
}

func (c *captureProvider) SnapshotRooms() map[string]RoomEntry             { return c.rooms }
func (c *captureProvider) SnapshotBeds() map[string]*card.BedState         { return c.beds }

// nopRefresher 测试用：不真 publish。
type nopRefresher struct{}

func (nopRefresher) RefreshRoomRiskAndMaybePublish(_ context.Context, _ string, _ int64) bool {
	return false
}

const (
	tBathroom = "fd00:0:3:111:3:300::/88"
	tRoom     = "fd00:0:3:111:3:100::/88"
	tBed      = "fd00:0:3:111:3:100:101::/96"
)

// helper: 跑一次 Tick 并返回 firer 收集的 fires。
func tickOnce(rules []Rule, rooms map[string]RoomEntry, beds map[string]*card.BedState, nowMs int64) (*Supervisor, *captureFirer) {
	cap := &captureFirer{}
	s := NewSupervisor(rules, cap, nil)
	s.SetProvider(&captureProvider{rooms: rooms, beds: beds})
	s.SetRefresher(nopRefresher{})
	s.Tick(nowMs)
	return s, cap
}

// 1. Stay — bathroom AloneSinceTs 超 45min（day）→ fire
func TestStay_AloneOverDayThreshold_Fires(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli() // 14:00 UTC = white-day
	rs := &card.RoomState{
		TotalPeople:  1,
		AloneContinuousMin: 46, // 46min 前
	}
	rooms := map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}
	_, cap := tickOnce(rules, rooms, nil, now)
	if !hasFireFor(cap.snapFires(), alarm.Stay, tBathroom) {
		t.Fatalf("Stay should fire at 46min alone day, got %+v", cap.snapFires())
	}
}

// 2. Stay — bathroom 多人（TotalPeople==2）即使 alone-time 超阈也不 fire（caregiver 协助）
func TestStay_MultiPeople_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()
	rs := &card.RoomState{
		TotalPeople:        2, // 多人
		AloneContinuousMin: 0, // sensor 已 gate
	}
	rooms := map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}
	_, cap := tickOnce(rules, rooms, nil, now)
	if hasFireFor(cap.snapFires(), alarm.Stay, tBathroom) {
		t.Errorf("Stay should NOT fire when multi-people, got %+v", cap.snapFires())
	}
}

// 3. Stay — day vs night 阈值切换：23:00 (night) 31min alone → fire（night 30min）
//          14:00 (day) 31min alone → 不 fire（day 45min）
func TestStay_NightThresholdLowerThanDay(t *testing.T) {
	rules := DefaultRules()

	nightHour := time.Date(2026, 5, 24, 23, 0, 0, 0, time.UTC).UnixMilli()
	dayHour := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	rsNight := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 31}
	roomsNight := map[string]RoomEntry{tBathroom: {State: rsNight, Kind: card.RoomTypeBathroom}}
	_, capNight := tickOnce(rules, roomsNight, nil, nightHour)
	if !hasFireFor(capNight.snapFires(), alarm.Stay, tBathroom) {
		t.Errorf("Stay should fire 31min alone at night (threshold 30min)")
	}

	rsDay := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 31}
	roomsDay := map[string]RoomEntry{tBathroom: {State: rsDay, Kind: card.RoomTypeBathroom}}
	_, capDay := tickOnce(rules, roomsDay, nil, dayHour)
	if hasFireFor(capDay.snapFires(), alarm.Stay, tBathroom) {
		t.Errorf("Stay should NOT fire 31min alone at day (threshold 45min)")
	}
}

// 4. Stay — fire-once-until-vacant：fire 后不再 re-fire；TotalPeople 离开 1 即清 fired
func TestStay_FireOnceUntilNotAlone(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	cap := &captureFirer{}
	s := NewSupervisor(rules, cap, nil)
	prov := &captureProvider{}
	s.SetProvider(prov)
	s.SetRefresher(nopRefresher{})

	rs := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 46}
	prov.rooms = map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}
	s.Tick(now)
	if len(cap.snapFires()) != 1 {
		t.Fatalf("first tick should fire once, got %d", len(cap.snapFires()))
	}
	// 再 tick：fired 仍 set → 不重 fire
	s.Tick(now + 1000)
	if len(cap.snapFires()) != 1 {
		t.Errorf("second tick should not re-fire, got %d", len(cap.snapFires()))
	}
	// 多人进入 → KeepCounting 假 → 清 fired
	rs.TotalPeople = 2
	rs.AloneContinuousMin = 0
	s.Tick(now + 2000)
	if len(cap.snapFires()) != 1 {
		t.Errorf("multi-people should not fire, got %d", len(cap.snapFires()))
	}
	if s.evaluator.FiredCount() != 0 {
		t.Errorf("fired gate should clear after multi-people, got %d", s.evaluator.FiredCount())
	}
	// 重新独居 + 46min → 重新 fire
	rs.TotalPeople = 1
	rs.AloneContinuousMin = 46
	s.Tick(now + 3000)
	if len(cap.snapFires()) != 2 {
		t.Errorf("after alone-reset should re-fire, got %d", len(cap.snapFires()))
	}
}

// 5. LeftBed — bed Vacant + room Vacant 超 30min → fire
func TestLeftBed_VacantOver30Min_Fires(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	bs := &card.BedState{
		BedStatus:   1, // NotInBed
		BedStatusTs: now - (31 * 60 * 1000),
	}
	rs := &card.RoomState{TotalPeople: 0} // 房间也空
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, now)
	if !hasFireFor(cap.snapFires(), alarm.LeftBed, tBed) {
		t.Fatalf("LeftBed should fire at 31min vacant, got %+v", cap.snapFires())
	}
}

// 6. LeftBed — bed Vacant 但 room Occupied（人回房）→ 不 fire
func TestLeftBed_PeerRoomOccupied_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	bs := &card.BedState{BedStatus: 1, BedStatusTs: now - (31 * 60 * 1000)}
	rs := &card.RoomState{TotalPeople: 1} // 人在房
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, now)
	if hasFireFor(cap.snapFires(), alarm.LeftBed, tBed) {
		t.Errorf("LeftBed should NOT fire when peer room occupied, got %+v", cap.snapFires())
	}
}

// 7. LeftBed — bed Occupied → KeepCounting Self=vacant 失败 → 不 fire
func TestLeftBed_BedOccupied_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	bs := &card.BedState{BedStatus: 0, BedStatusTs: now - (60 * 60 * 1000)} // InBed
	rs := &card.RoomState{TotalPeople: 0}
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, now)
	if hasFireFor(cap.snapFires(), alarm.LeftBed, tBed) {
		t.Errorf("LeftBed should NOT fire when bed occupied")
	}
}

// 8. NightAbsence — 夜间窗内 + room Vacant + bed Vacant 超 30min → fire
// Supervisor.loc 默认 nil → Evaluator 用 UTC；本测试统一用 UTC 时间避免 server local TZ 干扰。
func TestNightAbsence_NightWindowVacantOver30Min_Fires(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 3, 0, 0, 0, time.UTC).UnixMilli() // 03:00 UTC ∈ 21-07

	rs := &card.RoomState{TotalPeople: 0, LastExitTs: now - (31 * 60 * 1000)}
	bs := &card.BedState{BedStatus: 1, BedStatusTs: now - (60 * 60 * 1000)}
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, now)
	if !hasFireFor(cap.snapFires(), alarm.NightAbsence, tRoom) {
		t.Fatalf("NightAbsence should fire 31min vacant in night window, got %+v", cap.snapFires())
	}
}

// 9. NightAbsence — 白天 12:00 即使超阈也不 fire（NightOnly window）
func TestNightAbsence_NoonNotInWindow_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	noon := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC).UnixMilli()

	rs := &card.RoomState{TotalPeople: 0, LastExitTs: noon - (31 * 60 * 1000)}
	bs := &card.BedState{BedStatus: 1, BedStatusTs: noon - (60 * 60 * 1000)}
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, noon)
	if hasFireFor(cap.snapFires(), alarm.NightAbsence, tRoom) {
		t.Errorf("NightAbsence should NOT fire outside night window, got %+v", cap.snapFires())
	}
}

// 10. NightAbsence — bed Occupied (peer-zone bed has resident sleeping) → 不 fire
func TestNightAbsence_PeerBedOccupied_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 3, 0, 0, 0, time.UTC).UnixMilli()

	rs := &card.RoomState{TotalPeople: 0, LastExitTs: now - (31 * 60 * 1000)}
	bs := &card.BedState{BedStatus: 0, BedStatusTs: now - (60 * 60 * 1000)} // bed Occupied
	rooms := map[string]RoomEntry{tRoom: {State: rs, Kind: card.RoomTypeDefault}}
	beds := map[string]*card.BedState{tBed: bs}
	_, cap := tickOnce(rules, rooms, beds, now)
	if hasFireFor(cap.snapFires(), alarm.NightAbsence, tRoom) {
		t.Errorf("NightAbsence should NOT fire when peer bed occupied, got %+v", cap.snapFires())
	}
}

// 11. TimeWindow.Active 跨午夜
func TestTimeWindow_Active(t *testing.T) {
	w := &TimeWindow{StartH: 21, StartM: 0, EndH: 7, EndM: 0}
	cases := []struct {
		h, m int
		want bool
	}{
		{12, 0, false},
		{20, 59, false},
		{21, 0, true},
		{23, 30, true},
		{3, 0, true},
		{6, 59, true},
		{7, 0, false},
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

// 12. nil window = 全天 active
func TestTimeWindow_NilAllDay(t *testing.T) {
	var w *TimeWindow
	if !w.Active(time.Now()) {
		t.Errorf("nil window should be all-day active")
	}
}

// 13. ReloadRules 替换规则集 — fired 不动
func TestSupervisor_ReloadRules(t *testing.T) {
	s := NewSupervisor(DefaultRules(), nil, nil)
	if len(s.rules) != 3 {
		t.Fatalf("default rules should be 3, got %d", len(s.rules))
	}
	s.ReloadRules([]Rule{})
	if len(s.rules) != 0 {
		t.Errorf("after reload should be 0")
	}
}

// 14. Stay — anchor=0 表示无独居锚点 → 不 fire（即使阈值满足）
func TestStay_AnchorZero_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()
	rs := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 0} // 异常状态：count=1 但 MIN=0
	rooms := map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}
	_, cap := tickOnce(rules, rooms, nil, now)
	if len(cap.snapFires()) > 0 {
		t.Errorf("Stay should not fire with anchor=0, got %+v", cap.snapFires())
	}
}

// 15a. per-device 阈值覆盖：spatial_config 配 60min Stay → 46min alone 不 fire；61min fire
type fakeResolver struct {
	overrideSec int
	enabled     bool
}

func (f fakeResolver) Resolve(_ context.Context, _, _ string, fallbackSec int) (int, bool) {
	if !f.enabled {
		return fallbackSec, false
	}
	if f.overrideSec > 0 {
		return f.overrideSec, true
	}
	return fallbackSec, true
}

type fakeLookup struct{ addr string }

func (f fakeLookup) FindPrimaryDevice(_, _ string) string { return f.addr }

func TestStay_PerDeviceThresholdOverride(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	rs := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 46}
	rooms := map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}

	cap := &captureFirer{}
	s := NewSupervisor(rules, cap, nil)
	s.SetProvider(&captureProvider{rooms: rooms})
	s.SetRefresher(nopRefresher{})
	s.SetDeviceLookup(fakeLookup{addr: "fd00::1"})
	s.SetThresholdResolver(fakeResolver{overrideSec: 60 * 60, enabled: true}) // 老人 X 配 60min
	s.Tick(now)
	if len(cap.snapFires()) != 0 {
		t.Errorf("46min < 60min override → should NOT fire, got %+v", cap.snapFires())
	}

	// 61min → 超 60min override 才 fire
	rs.AloneContinuousMin = 61
	s.Tick(now + 1000)
	if !hasFireFor(cap.snapFires(), alarm.Stay, tBathroom) {
		t.Errorf("61min > 60min override → should fire, got %+v", cap.snapFires())
	}
}

// 15b. per-device disabled → 不 fire 且 fired 清零
func TestStay_PerDeviceDisabled_DoesNotFire(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()

	rs := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 60}
	rooms := map[string]RoomEntry{tBathroom: {State: rs, Kind: card.RoomTypeBathroom}}

	cap := &captureFirer{}
	s := NewSupervisor(rules, cap, nil)
	s.SetProvider(&captureProvider{rooms: rooms})
	s.SetRefresher(nopRefresher{})
	s.SetDeviceLookup(fakeLookup{addr: "fd00::1"})
	s.SetThresholdResolver(fakeResolver{enabled: false}) // device-level disabled
	s.Tick(now)
	if len(cap.snapFires()) != 0 {
		t.Errorf("disabled → should not fire, got %+v", cap.snapFires())
	}
}

// 16. Stay — Room Kind=Default 不被 bathroom rule 评估
func TestStay_OnlyBathroomKind(t *testing.T) {
	rules := DefaultRules()
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).UnixMilli()
	rs := &card.RoomState{TotalPeople: 1, AloneContinuousMin: 46}
	rooms := map[string]RoomEntry{
		tRoom: {State: rs, Kind: card.RoomTypeDefault}, // 不是 bathroom
	}
	_, cap := tickOnce(rules, rooms, nil, now)
	if hasFireFor(cap.snapFires(), alarm.Stay, tRoom) {
		t.Errorf("Stay should not fire on non-bathroom room, got %+v", cap.snapFires())
	}
}
