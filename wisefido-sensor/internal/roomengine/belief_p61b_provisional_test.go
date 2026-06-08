package roomengine

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"owl-common/card"
	"wisefido-sensor/internal/roomengine/belief"
)

// belief_p61b_provisional_test.go — P6.1b-D 阶段2(审查㉛ Opt-1)provisional 分级状态机 e2e。
// 验:小卫生间 lost → provisional-now;设备富窗到未佐证→escalate;recapture→cancel;设备贫→suppressed。

const (
	p61bRoom     = "fd00:0:3:222:b::/88"
	p61bSuite    = "fd00:0:3:222::/80"
	p61bBathDev  = "fd00:0:3:222:b::1"
	p61bOtherDev = "fd00:0:3:222:r::1"
)

// mkP61bEngine 建带小卫生间 gate 的 Engine + seed 一个丢失的浴室 track。
// rich=true → suite 有别台设备(设备富);false → 浴室独苗(设备贫)。
func mkP61bEngine(t *testing.T, nowMs int64, rich bool, census *SuiteCensusManager) (*Engine, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	deviceRoom := map[string]string{p61bBathDev: p61bRoom}
	if rich {
		deviceRoom[p61bOtherDev] = "fd00:0:3:222:r::/88" // 同 suite 别房别台
	}
	e := &Engine{
		logger:        zap.New(core),
		rooms:         map[string]*TrackManager{p61bRoom: {roomID: p61bRoom, tracks: map[int]*TrackState{}, bedCount: 1}},
		grids:         map[string]*RoomGrid{},
		roomSuiteID:   map[string]string{p61bRoom: p61bSuite, "fd00:0:3:222:r::/88": p61bSuite},
		roomType:      map[string]int{p61bRoom: card.RoomTypeBathroom},
		smallBathroom: map[string]bool{p61bRoom: true}, // gate ON(阶段1)
		deviceRoom:    deviceRoom,
		beliefShadows: map[string]*beliefShadow{},
		suiteCensus:   census,
	}
	sh := e.beliefShadowFor(p61bRoom)
	sh.tracks[7] = &beliefShadowTrack{
		lastSeenMs:    nowMs - 70_000, // > 60s TTL = lost
		stillBoxAgeMs: 0,              // moving→lost
		geom:          belief.GeomInToilet,
		lastX:         50, lastY: 50,
		lostAnchor: nowMs - 70_000,
	}
	sh.tlayer[7] = &beliefShadowTLayer{tb: belief.NewTrackBelief(), device: p61bBathDev, realLO: 2.0} // real(走失者)
	return e, logs
}

func hasMsg(logs *observer.ObservedLogs, msg string) bool { return logs.FilterMessage(msg).Len() > 0 }

// ① 设备富 + 无佐证 → provisional-now,窗(30min)到 → escalate(真摔,延迟但不漏)。
func TestP61bRichEscalate(t *testing.T) {
	nowMs := int64(10_000_000)
	e, logs := mkP61bEngine(t, nowMs, true, NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil))
	e.beliefShadowTick(p61bRoom, nil, nowMs)
	if !hasMsg(logs, "belief_shadow_lostfall_provisional") {
		t.Fatalf("首 lost tick 应 log provisional-now")
	}
	if hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("窗未到不该 escalate")
	}
	e.beliefShadowTick(p61bRoom, nil, nowMs+beliefProvisionalRichWindowMs+2000) // 窗到
	if !hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("设备富窗到未佐证应 escalate(全 sev 真摔)")
	}
	if hasMsg(logs, "belief_shadow_lostfall_cancel") || hasMsg(logs, "belief_shadow_lostfall_suppressed") {
		t.Fatalf("设备富无佐证不该 cancel/suppress")
	}
}

// ② recapture(sole resident 回床)→ cancel(离场,不 escalate)。
func TestP61bRecaptureCancel(t *testing.T) {
	nowMs := int64(10_000_000)
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	census.GetOrCreate(p61bSuite).Persons["r"] = &SuitePerson{PersonID: "r", Role: SuitePersonResident, SleepadAnchored: true, AnchorRoomType: card.RoomTypeDefault}
	e, logs := mkP61bEngine(t, nowMs, true, census)
	e.beliefShadowTick(p61bRoom, nil, nowMs)
	if !hasMsg(logs, "belief_shadow_lostfall_provisional") {
		t.Fatalf("应先 provisional")
	}
	if !hasMsg(logs, "belief_shadow_lostfall_cancel") {
		t.Fatalf("sole resident 回床 recapture 应 cancel(离场)")
	}
	e.beliefShadowTick(p61bRoom, nil, nowMs+beliefProvisionalRichWindowMs+2000)
	if hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("已 cancel 不该再 escalate")
	}
}

// ③ 设备贫(浴室独苗)→ provisional,短窗(2min)到 → suppressed(压制+LOG,不静默)。
func TestP61bPoorSuppress(t *testing.T) {
	nowMs := int64(10_000_000)
	e, logs := mkP61bEngine(t, nowMs, false, NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil))
	e.beliefShadowTick(p61bRoom, nil, nowMs)
	if !hasMsg(logs, "belief_shadow_lostfall_provisional") {
		t.Fatalf("应先 provisional")
	}
	e.beliefShadowTick(p61bRoom, nil, nowMs+beliefProvisionalPoorWindowMs+2000) // 短窗到
	if !hasMsg(logs, "belief_shadow_lostfall_suppressed") {
		t.Fatalf("设备贫短窗到应 suppressed(压制+LOG)")
	}
	if hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("设备贫不该 escalate(走压制路径)")
	}
}
