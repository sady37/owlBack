package zoneengine

import (
	"testing"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
)

func newSleepaceTestAdapter(t *testing.T) (*SleepaceAdapter, *captureListener) {
	t.Helper()
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	cap := &captureListener{}
	engine.AddListener(cap)
	a := NewSleepaceAdapter(nil, engine, nil)
	return a, cap
}

func TestSleepaceAdapter_InBedFlipsBedOccupied(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", alarm.InBed, "card-1", now, nil)
	a.handleMsg(msg)

	events := cap.Events()
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
	if bedEv.NewState.LastSource != "sleepace" {
		t.Fatalf("want LastSource=sleepace, got %s", bedEv.NewState.LastSource)
	}
}

func TestSleepaceAdapter_LeftBedTriggersLeaving(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	addr := "fd00:0:3:111:3:101:a2ac:d523"
	a.handleMsg(mkMsg(addr, "Sleepad", alarm.InBed, "card-1", now, nil))
	// sleepace enter latch=10s（与 radar 同），LeftBed 在 11s 后才被 accept
	a.handleMsg(mkMsg(addr, "Sleepad", alarm.LeftBed, "card-1", now+11_000, nil))

	var bedTrans []string
	for _, e := range cap.Events() {
		if e.ZoneType == ZoneTypeBed {
			bedTrans = append(bedTrans, e.Transition)
		}
	}
	if len(bedTrans) < 2 || bedTrans[0] != TransitionOccupied || bedTrans[1] != TransitionLeaving {
		t.Fatalf("want [occupied leaving], got %v", bedTrans)
	}
}

func TestSleepaceAdapter_BedSitUpNotConsumed(t *testing.T) {
	// BedSitUp 是行为分类 alarm（坐起），不应驱动 zone state；它由 cardagg / 其他业务
	// 路径独立处理，presence 通道只走 InBed/LeftBed 双流（sleepace alarm + radar event）。
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	addr := "fd00:0:3:111:3:101:a2ac:d523"
	a.handleMsg(mkMsg(addr, "Sleepad", alarm.InBed, "card-1", now, nil))
	a.handleMsg(mkMsg(addr, "Sleepad", alarm.BedSitUp, "card-1", now+11_000, nil))

	var bedTrans []string
	for _, e := range cap.Events() {
		if e.ZoneType == ZoneTypeBed {
			bedTrans = append(bedTrans, e.Transition)
		}
	}
	// 只有 InBed → Occupied 一条；BedSitUp 不应触发任何翻转
	if len(bedTrans) != 1 || bedTrans[0] != TransitionOccupied {
		t.Fatalf("want only [occupied] (BedSitUp ignored), got %v", bedTrans)
	}
}

func TestSleepaceAdapter_RadarSkipped(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Radar", alarm.InBed, "card-1", now, nil)
	a.handleMsg(msg)
	if len(cap.Events()) != 0 {
		t.Fatalf("Radar alarm should be skipped by sleepace adapter, got %v", cap.Events())
	}
}

func TestSleepaceAdapter_SleeppadVariantAccepted(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	// 注意拼写变体 "Sleeppad"（双 p）也是合法 sleepace device_type
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleeppad", alarm.InBed, "card-1", now, nil)
	a.handleMsg(msg)
	if len(cap.Events()) == 0 {
		t.Fatalf("Sleeppad variant should be accepted")
	}
}

func TestSleepaceAdapter_EventStatusEndSkipped(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	// event_status="end" → alarm 解除，与 zone 翻转无关，跳过
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", alarm.InBed, "card-1", now, map[string]interface{}{
		observation.FieldEventStatus: "end",
	})
	a.handleMsg(msg)
	if len(cap.Events()) != 0 {
		t.Fatalf("event_status=end should be skipped, got %v", cap.Events())
	}
}

func TestSleepaceAdapter_UnboundDeviceSkipped(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", alarm.InBed, "", now, nil)
	a.handleMsg(msg)
	if len(cap.Events()) != 0 {
		t.Fatalf("unbound device should be skipped, got %v", cap.Events())
	}
}

func TestSleepaceAdapter_StaleMessageDropped(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	stale := time.Now().UnixMilli() - 60_000 // 60s 前 > 30s 阈值
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", alarm.InBed, "card-1", stale, nil)
	a.handleMsg(msg)
	if len(cap.Events()) != 0 {
		t.Fatalf("stale (>30s) message should be dropped, got %v", cap.Events())
	}
}

func TestSleepaceAdapter_OtherAlarmsIgnored(t *testing.T) {
	a, cap := newSleepaceTestAdapter(t)
	now := time.Now().UnixMilli()
	// sleepace 其它报警（如 NoBodyMove）不应该影响 zone state
	msg := mkMsg("fd00:0:3:111:3:101:a2ac:d523", "Sleepad", "NoBodyMove", "card-1", now, nil)
	a.handleMsg(msg)
	if len(cap.Events()) != 0 {
		t.Fatalf("non-bed alarm should be ignored, got %v", cap.Events())
	}
}
