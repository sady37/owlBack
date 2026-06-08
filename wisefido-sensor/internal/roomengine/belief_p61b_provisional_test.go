package roomengine

import (
	"os"
	"path/filepath"
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

// ④ 审查㉝ 关键对抗:np=0 在场(firmware 报屋内空)但无 recapture → 仍 escalate(np 永不 cancel)。
// 证 np=0 是 lost-fall 共有条件非判别器(摔/离共有);用它 cancel 会 cancel 真摔=漏报。
func TestP61bNp0DoesNotCancel(t *testing.T) {
	nowMs := int64(10_000_000)
	e, logs := mkP61bEngine(t, nowMs, true, NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil))
	e.rooms[p61bRoom].lastNumberPeopleZeroMs = nowMs - 1000 // firmware np=0(屋内空断言)在场
	e.beliefShadowTick(p61bRoom, nil, nowMs)
	e.beliefShadowTick(p61bRoom, nil, nowMs+beliefProvisionalRichWindowMs+2000)
	if hasMsg(logs, "belief_shadow_lostfall_cancel") {
		t.Fatalf("np=0 在场不该 cancel(㉝:absence≠leave,cancel 真摔=漏报)")
	}
	if !hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("np=0 在场无 recapture 仍应 escalate(不漏报)")
	}
}

// ⑤ 审查㉝/㉚ 对抗:护工(visitor)在场 + sole resident 未回床(仍在浴室)→ 不 cancel 仍 escalate。
// 证 cancel 绑走失者本人 anchor(per-identity);visitor 移动非走失者重现 → 不触发 recapture。
func TestP61bVisitorDoesNotCancel(t *testing.T) {
	nowMs := int64(10_000_000)
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	c := census.GetOrCreate(p61bSuite)
	c.Persons["r"] = &SuitePerson{PersonID: "r", Role: SuitePersonResident, SleepadAnchored: false, AnchorRoomType: card.RoomTypeBathroom} // 仍在浴室,未回床=未 recapture
	c.Persons["v"] = &SuitePerson{PersonID: "v", Role: SuitePersonVisitor, AnchorRoomType: card.RoomTypeDefault}                            // 护工在别区
	e, logs := mkP61bEngine(t, nowMs, true, census)
	e.beliefShadowTick(p61bRoom, nil, nowMs)
	e.beliefShadowTick(p61bRoom, nil, nowMs+beliefProvisionalRichWindowMs+2000)
	if hasMsg(logs, "belief_shadow_lostfall_cancel") {
		t.Fatalf("护工在场+resident未回床不该 cancel(cancel 绑走失者 anchor,visitor 非重现)")
	}
	if !hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("护工在场无 recapture 仍应 escalate(不漏报)")
	}
}

// TestP61bCABBRealLayoutEngagesDPath — 审查㉟ 放行gate第1+3项:真 CABB 立项 layout 帧过真 beliefShadowTick D-path。
// 用真 CABB grid(boundary 派生)+ 真 layout 算 smallBathroom + lost track 在浴室内部(geom=OpenFloor,真 CABB
// 无 toilet 对象)→ 断言 D-branch engage(provisional→escalate)。证 D-path 在 founding 案真 engage 非 silent-miss,
// 且广义 geom 条件(非仅 InToilet/InEnter)接住 CABB OpenFloor。
func TestP61bCABBRealLayoutEngagesDPath(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../../../doc/cases", "hunzi-cabb-lost-0601-2247-FP/room_layout.json"))
	if err != nil {
		t.Skipf("CABB fixture 缺失: %v", err)
	}
	cfg, err := ParseLayoutConfig("fd00:0:3:411::/128", b)
	if err != nil {
		t.Fatalf("ParseLayoutConfig: %v", err)
	}
	cfg.RoomType = card.RoomTypeBathroom
	grid, _ := buildGridFromLayout(t, "hunzi-cabb-lost-0601-2247-FP/room_layout.json")
	const room, suite, dev = "fd00:0:3:411::/88", "fd00:0:3:411::/80", "fd00:0:3:411:1:200:10d5:cabb"
	lostX, lostY := -50, 250 // CABB 浴室内部(radar 处),真 grid geom
	g := geomFromGrid(grid, lostX, lostY)
	if g == belief.GeomInToilet || g == belief.GeomInEnter {
		t.Logf("注:CABB 内部 geom=%v(若是门区则旧窄条件也接;本测专证非门区也 engage)", g)
	}
	nowMs := int64(10_000_000)
	core, logs := observer.New(zapcore.InfoLevel)
	e := &Engine{
		logger:        zap.New(core),
		rooms:         map[string]*TrackManager{room: {roomID: room, tracks: map[int]*TrackState{}, bedCount: 1}},
		grids:         map[string]*RoomGrid{room: grid},
		roomSuiteID:   map[string]string{room: suite, "fd00:0:3:411:r::/88": suite}, // 别台房同 suite
		roomType:      map[string]int{room: card.RoomTypeBathroom},
		smallBathroom: map[string]bool{room: isSmallBathroomCfg(cfg, cfg.RoomW, cfg.RoomH)}, // 真 layout 算
		deviceRoom:    map[string]string{dev: room, "fd00:0:3:411:r::1": "fd00:0:3:411:r::/88"}, // 设备富(别台)
		beliefShadows: map[string]*beliefShadow{},
		suiteCensus:   NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil),
	}
	if !e.smallBathroom[room] {
		t.Fatalf("真 CABB layout 应判小卫生间(gate fire);否则 D-path 不 engage=silent-miss")
	}
	sh := e.beliefShadowFor(room)
	sh.tracks[7] = &beliefShadowTrack{lastSeenMs: nowMs - 70_000, stillBoxAgeMs: 0, geom: g, lastX: lostX, lastY: lostY, lostAnchor: nowMs - 70_000}
	sh.tlayer[7] = &beliefShadowTLayer{tb: belief.NewTrackBelief(), device: dev, realLO: 2.0}
	e.beliefShadowTick(room, nil, nowMs)
	if !hasMsg(logs, "belief_shadow_lostfall_provisional") {
		t.Fatalf("真 CABB lost track(geom=%v)应进 D-branch → provisional(广义 geom 条件接住非门区)", g)
	}
	e.beliefShadowTick(room, nil, nowMs+beliefProvisionalRichWindowMs+2000)
	if !hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("真 CABB 无 recapture → 窗到应 escalate(预期 lean-surface,founding 案 engage 证实)")
	}
}

// TestP61bCABBPoorSuppresses — 审查㊱:engage≠治愈。CABB 是 FP 案(真离场,np=0+335s,无recapture)。
// **设备贫**(浴室独苗)→ suppress = CABB FP **治愈**(不 page,LOG)。对照 rich→escalate(FP 仍在需 Opt-3)。
// CABB 实际设备 tier = ①fleet 事实(待用户);此测证"若 CABB 是设备贫则 P6.1b-D 治愈其 FP"。
func TestP61bCABBPoorSuppresses(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../../../doc/cases", "hunzi-cabb-lost-0601-2247-FP/room_layout.json"))
	if err != nil {
		t.Skipf("CABB fixture 缺失: %v", err)
	}
	cfg, err := ParseLayoutConfig("fd00:0:3:411::/128", b)
	if err != nil {
		t.Fatalf("ParseLayoutConfig: %v", err)
	}
	cfg.RoomType = card.RoomTypeBathroom
	grid, _ := buildGridFromLayout(t, "hunzi-cabb-lost-0601-2247-FP/room_layout.json")
	const room, suite, dev = "fd00:0:3:411::/88", "fd00:0:3:411::/80", "fd00:0:3:411:1:200:10d5:cabb"
	nowMs := int64(10_000_000)
	core, logs := observer.New(zapcore.InfoLevel)
	e := &Engine{
		logger:        zap.New(core),
		rooms:         map[string]*TrackManager{room: {roomID: room, tracks: map[int]*TrackState{}, bedCount: 1}},
		grids:         map[string]*RoomGrid{room: grid},
		roomSuiteID:   map[string]string{room: suite},
		roomType:      map[string]int{room: card.RoomTypeBathroom},
		smallBathroom: map[string]bool{room: isSmallBathroomCfg(cfg, cfg.RoomW, cfg.RoomH)},
		deviceRoom:    map[string]string{dev: room}, // 浴室独苗=设备贫(无别台)
		beliefShadows: map[string]*beliefShadow{},
		suiteCensus:   NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil),
	}
	sh := e.beliefShadowFor(room)
	g := geomFromGrid(grid, -50, 250)
	sh.tracks[7] = &beliefShadowTrack{lastSeenMs: nowMs - 70_000, stillBoxAgeMs: 0, geom: g, lastX: -50, lastY: 250, lostAnchor: nowMs - 70_000}
	sh.tlayer[7] = &beliefShadowTLayer{tb: belief.NewTrackBelief(), device: dev, realLO: 2.0}
	e.beliefShadowTick(room, nil, nowMs)
	e.beliefShadowTick(room, nil, nowMs+beliefProvisionalPoorWindowMs+2000)
	if !hasMsg(logs, "belief_shadow_lostfall_suppressed") {
		t.Fatalf("设备贫 CABB 应 suppress(FP 治愈,不 page+LOG)")
	}
	if hasMsg(logs, "belief_shadow_lostfall_escalate") {
		t.Fatalf("设备贫不该 escalate(走压制治愈,非 FP)")
	}
}
