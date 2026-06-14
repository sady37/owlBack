package zoneengine

import (
	"net/netip"
	"testing"
	"time"

	"owl-common/alarm"
	rediscommon "owl-common/redis"
)

// stubBathroom 测试用 BathroomLookup：固定一个 RoomPref → bathroom 命中。
type stubBathroom struct{ bathPref string }

func (s stubBathroom) IsBathroom(roomPref string) bool { return roomPref == s.bathPref }

// newTestAdapter — engine + 默认规则 + listener；redis client 不接（直接调 handleMsg）。
func newTestAdapter(t *testing.T, lookup BathroomLookup) (*RadarAdapter, *Engine, *captureListener) {
	t.Helper()
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	cap := &captureListener{}
	engine.AddListener(cap)
	a := NewRadarAdapter(nil, engine, lookup, nil)
	return a, engine, cap
}

func mkMsg(addr, deviceType, category, cardID string, ts int64, fields map[string]interface{}) *rediscommon.IoTStreamMessage {
	parsed, _ := netip.ParseAddr(addr)
	dataValue := []interface{}{}
	if fields != nil {
		dataValue = []interface{}{fields}
	}
	return &rediscommon.IoTStreamMessage{
		SubjectEntity: cardID,
		DeviceAddr:    parsed,
		DeviceType:    deviceType,
		Timestamp:     ts,
		TopicType:     "event",
		Category:      category,
		DataValue:     dataValue,
	}
}

func TestRadarAdapter_EnterRoomFlipsRoomOccupied(t *testing.T) {
	// device hextet 6 = 0x0101；/88 mask 后 low byte=0 → "fd00:0:3:111:3:100::/88"。
	a, _, cap := newTestAdapter(t, stubBathroom{bathPref: "fd00:0:3:111:3:9ff::/88"})
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.EnterRoom, "card-1", now, map[string]interface{}{
		"track_id":      float64(1),
		"number_people": float64(1),
	})
	a.handleMsg(msg)

	events := cap.Events()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.ZoneType != ZoneTypeRoom {
		t.Fatalf("want ZoneTypeRoom, got %v", ev.ZoneType)
	}
	if ev.Transition != TransitionOccupied {
		t.Fatalf("want occupied, got %s", ev.Transition)
	}
	if ev.ZoneID != "fd00:0:3:111:3:100::/88" {
		t.Fatalf("want /88 prefix, got %s", ev.ZoneID)
	}
}

func TestRadarAdapter_EnterRoomBathroomRoutesToBathroom(t *testing.T) {
	// 与上 test 同 device addr → masked /88 = "fd00:0:3:111:3:100::/88"
	bathPref := "fd00:0:3:111:3:100::/88"
	a, _, cap := newTestAdapter(t, stubBathroom{bathPref: bathPref})
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.EnterRoom, "card-1", now, map[string]interface{}{
		"track_id": float64(1),
	})
	a.handleMsg(msg)

	events := cap.Events()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	if events[0].ZoneType != ZoneTypeBathroom {
		t.Fatalf("want ZoneTypeBathroom, got %v", events[0].ZoneType)
	}
	if events[0].ZoneID != bathPref {
		t.Fatalf("want %s, got %s", bathPref, events[0].ZoneID)
	}
}

func TestRadarAdapter_NumberPeopleIsCountChange(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.NumberPeople, "card-1", now, map[string]interface{}{
		"number_people": float64(3),
	})
	a.handleMsg(msg)

	events := cap.Events()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.Transition != "count_change" {
		t.Fatalf("want count_change, got %s", ev.Transition)
	}
	if ev.NewState.Count != 3 {
		t.Fatalf("want count=3, got %d", ev.NewState.Count)
	}
}

func TestRadarAdapter_InBedFlipsBedOccupied(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.InBed, "card-1", now, map[string]interface{}{
		"track_id": float64(1),
	})
	a.handleMsg(msg)

	// 同步：subset_invariant 抬升 room → bed event + room event 各一条
	events := cap.Events()
	if len(events) < 1 {
		t.Fatalf("want ≥1 event, got %d", len(events))
	}
	var bedEv *ZoneEvent
	for i := range events {
		if events[i].ZoneType == ZoneTypeBed {
			bedEv = &events[i]
			break
		}
	}
	if bedEv == nil {
		t.Fatalf("no bed event in %v", events)
	}
	if bedEv.Transition != TransitionOccupied {
		t.Fatalf("want occupied, got %s", bedEv.Transition)
	}
	if bedEv.ZoneID != "fd00:0:3:111:3:101::/96" {
		t.Fatalf("want /96 prefix, got %s", bedEv.ZoneID)
	}
}

func TestRadarAdapter_LeftBedTriggersLeaving(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	addr := "fd00:0:3:111:3:101:a2ac:d523"
	a.handleMsg(mkMsg(addr, "Radar", alarm.InBed, "card-1", now, nil))
	// 间隔 11s 越过 enter_latch=10s（DefaultRules bed enter latch_sec=10），LeftBed 才能被 scorer accept。
	a.handleMsg(mkMsg(addr, "Radar", alarm.LeftBed, "card-1", now+11_000, nil))

	events := cap.Events()
	// 至少需要 occupied → leaving 两条 bed event
	var bedTrans []string
	for _, e := range events {
		if e.ZoneType == ZoneTypeBed {
			bedTrans = append(bedTrans, e.Transition)
		}
	}
	if len(bedTrans) < 2 {
		t.Fatalf("want ≥2 bed transitions, got %v", bedTrans)
	}
	if bedTrans[0] != TransitionOccupied || bedTrans[1] != TransitionLeaving {
		t.Fatalf("want [occupied leaving], got %v", bedTrans)
	}
}

func TestRadarAdapter_NonRadarSkipped(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", alarm.InBed, "card-1", now, nil)
	a.handleMsg(msg)

	if len(cap.Events()) != 0 {
		t.Fatalf("Sleepad event should be skipped, got %v", cap.Events())
	}
}

func TestRadarAdapter_UnboundDeviceSkipped(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.EnterRoom, "", now, nil)
	a.handleMsg(msg)

	if len(cap.Events()) != 0 {
		t.Fatalf("unbound device should be skipped, got %v", cap.Events())
	}
}

func TestRadarAdapter_StaleMessageDropped(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	stale := time.Now().UnixMilli() - 10_000 // 10s 前
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.EnterRoom, "card-1", stale, nil)
	a.handleMsg(msg)

	if len(cap.Events()) != 0 {
		t.Fatalf("stale (>6s) message should be dropped, got %v", cap.Events())
	}
}

func TestRadarAdapter_PrefixOfMasking(t *testing.T) {
	addr, _ := netip.ParseAddr("fd00:0:3:111:3:101:a2ac:d523")
	cases := []struct {
		bits int
		want string
	}{
		{88, "fd00:0:3:111:3:100::/88"}, // hextet 6 = 0x0101，/88 mask 后 low byte 清零 → 0x0100
		{96, "fd00:0:3:111:3:101::/96"}, // /96 = 6 整 hextet
		{80, "fd00:0:3:111:3::/80"},
	}
	for _, c := range cases {
		got := prefixOf(addr, c.bits)
		if got != c.want {
			t.Errorf("prefixOf(%d): want %s, got %s", c.bits, c.want, got)
		}
	}
}

func TestRadarAdapter_UnknownCategoryIgnored(t *testing.T) {
	a, _, cap := newTestAdapter(t, nil)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", "Walking", "card-1", now, nil)
	a.handleMsg(msg)

	if len(cap.Events()) != 0 {
		t.Fatalf("unknown category should be ignored, got %v", cap.Events())
	}
}
