package roomengine

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"owl-common/card"
	"wisefido-sensor/internal/roomengine/belief"
)

// belief_p65_test.go — P6.5①(审查㉑批准)跨设备 track 守恒 per-identity recapture。
// 两层验:(1) SoleResidentRecaptureState 判别核心(①②③④);(2) beliefShadowTick lost-sweep 端到端
// (recapture→suppress log / 多resident→skip-LOG,委员会③-log observability)。

const (
	p65Suite = "fd00:0:3:111::/80"
	p65Room  = "fd00:0:3:111:b::/88"
)

func p65Resident(id string, sleepad bool, room int) *SuitePerson {
	return &SuitePerson{PersonID: id, Role: SuitePersonResident, SleepadAnchored: sleepad, AnchorRoomType: room}
}

// TestP6SoleResidentRecaptureState — per-identity 判别核心(委员会 Q1/Q2 + ④ visitor 语义)。
func TestP6SoleResidentRecaptureState(t *testing.T) {
	mk := func(setup func(c *SuiteBedroomCensus)) *SuiteCensusManager {
		m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
		setup(m.GetOrCreate(p65Suite))
		return m
	}

	// ① 单 resident 回床(SleepadAnchored)→ recapture(exit 确认,lost-fall 抑制)。
	m1 := mk(func(c *SuiteBedroomCensus) { c.Persons["r"] = p65Resident("r", true, card.RoomTypeDefault) })
	if n, rec := m1.SoleResidentRecaptureState(p65Suite); n != 1 || !rec {
		t.Fatalf("① 单resident回床应 recapture:count=%d rec=%v", n, rec)
	}

	// ② 单 resident 仍在浴室(AnchorBathroom,未回床)→ 不 recapture(lost 浮出)。
	m2 := mk(func(c *SuiteBedroomCensus) { c.Persons["r"] = p65Resident("r", false, card.RoomTypeBathroom) })
	if n, rec := m2.SoleResidentRecaptureState(p65Suite); n != 1 || rec {
		t.Fatalf("② 单resident仍在浴室应不 recapture:count=%d rec=%v", n, rec)
	}

	// ③ 多 resident:A 浴室真摔 + 室友 B 在 bedroom recent → gate OFF(count>1 不 recapture,A 不被抑制)。
	m3 := mk(func(c *SuiteBedroomCensus) {
		c.Persons["A"] = p65Resident("A", false, card.RoomTypeBathroom)
		c.Persons["B"] = p65Resident("B", false, card.RoomTypeDefault)
	})
	if n, rec := m3.SoleResidentRecaptureState(p65Suite); n != 2 || rec {
		t.Fatalf("③ 多resident应 gate OFF(不抑制A):count=%d rec=%v", n, rec)
	}

	// ④ 单 resident(浴室未回)+ visitor 护工(bedroom):visitor 不计 → count==1,gate active,此例不 recapture。
	m4 := mk(func(c *SuiteBedroomCensus) {
		c.Persons["r"] = p65Resident("r", false, card.RoomTypeBathroom)
		c.Persons["v"] = &SuitePerson{PersonID: "v", Role: SuitePersonVisitor, AnchorRoomType: card.RoomTypeDefault}
	})
	if n, rec := m4.SoleResidentRecaptureState(p65Suite); n != 1 || rec {
		t.Fatalf("④ 单resident+visitor:visitor 不计 count 应==1 且此例不 recapture:count=%d rec=%v", n, rec)
	}

	// ④b 反证:同 ④ 但 resident 回床 → gate 未被 visitor 关,recapture 仍工作(锁 residentCount≠personCount 语义)。
	m4b := mk(func(c *SuiteBedroomCensus) {
		c.Persons["r"] = p65Resident("r", true, card.RoomTypeDefault)
		c.Persons["v"] = &SuitePerson{PersonID: "v", Role: SuitePersonVisitor, AnchorRoomType: card.RoomTypeDefault}
	})
	if n, rec := m4b.SoleResidentRecaptureState(p65Suite); n != 1 || !rec {
		t.Fatalf("④b 单resident回床+visitor:gate 应仍 active recapture 工作:count=%d rec=%v", n, rec)
	}
}

// TestP6ExitRecaptureLostSweep — beliefShadowTick lost-sweep 端到端:bathroom track 丢失 →
// 单resident recapture log+suppress / 多resident skip-LOG(委员会③-log,数据闸)。
func TestP6ExitRecaptureLostSweep(t *testing.T) {
	nowMs := int64(10_000_000)

	mkEngine := func(census *SuiteCensusManager) (*Engine, *observer.ObservedLogs) {
		core, logs := observer.New(zapcore.InfoLevel)
		e := &Engine{
			logger:        zap.New(core),
			rooms:         map[string]*TrackManager{p65Room: {roomID: p65Room, tracks: map[int]*TrackState{}, bedCount: 1}},
			grids:         map[string]*RoomGrid{}, // nil grid:reachableExitObs 自守
			roomSuiteID:   map[string]string{p65Room: p65Suite},
			beliefShadows: map[string]*beliefShadow{},
			suiteCensus:   census,
		}
		// seed 一个丢失的 bathroom track(moving→lost:stillBoxAge=0;geom=InToilet;超 TTL)。
		sh := e.beliefShadowFor(p65Room)
		sh.tracks[7] = &beliefShadowTrack{
			lastSeenMs:    nowMs - 70_000, // > 60s TTL = lost
			stillBoxAgeMs: 0,              // < MovingPrecondition → moving→lost(进 lost-fall 域)
			geom:          belief.GeomInToilet,
			lastX:         50, lastY: 50,
			lostAnchor: nowMs - 70_000,
		}
		return e, logs
	}
	has := func(logs *observer.ObservedLogs, msg string) bool { return logs.FilterMessage(msg).Len() > 0 }

	// ① 单 resident 回床 → exit_recapture log(shadow 抑制)。
	cRec := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	cRec.GetOrCreate(p65Suite).Persons["r"] = p65Resident("r", true, card.RoomTypeDefault)
	e1, logs1 := mkEngine(cRec)
	e1.beliefShadowTick(p65Room, nil, nowMs)
	if !has(logs1, "belief_shadow_exit_recapture") {
		t.Fatalf("① 单resident回床:应 log belief_shadow_exit_recapture")
	}
	if has(logs1, "belief_shadow_panic") {
		t.Fatalf("① panic:%v", logs1.FilterMessage("belief_shadow_panic").All())
	}

	// ② 单 resident 仍在浴室未回 → 不 recapture(lost 浮出,无抑制)。
	cStay := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	cStay.GetOrCreate(p65Suite).Persons["r"] = p65Resident("r", false, card.RoomTypeBathroom)
	e2, logs2 := mkEngine(cStay)
	e2.beliefShadowTick(p65Room, nil, nowMs)
	if has(logs2, "belief_shadow_exit_recapture") {
		t.Fatalf("② 单resident仍在浴室:不该 recapture(真摔须浮出)")
	}

	// ③ 多 resident 对抗:A 浴室丢轨 + B bedroom → A 不被抑制 + skip 须被 LOG(委员会③-log 数据闸)。
	cMulti := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	cm := cMulti.GetOrCreate(p65Suite)
	cm.Persons["A"] = p65Resident("A", false, card.RoomTypeBathroom)
	cm.Persons["B"] = p65Resident("B", false, card.RoomTypeDefault)
	e3, logs3 := mkEngine(cMulti)
	e3.beliefShadowTick(p65Room, nil, nowMs)
	if has(logs3, "belief_shadow_exit_recapture") {
		t.Fatalf("③ 多resident:A 真摔不该被 B 抑制(不该 recapture)")
	}
	if !has(logs3, "belief_shadow_recapture_skip_multiresident") {
		t.Fatalf("③ 多resident:skip 须被 LOG(observability 数据闸,委员会③-log)")
	}
}
