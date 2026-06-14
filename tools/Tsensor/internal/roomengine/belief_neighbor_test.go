package roomengine

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"owl-common/card"
)

// belief_neighbor_test.go — 死源#3 Neighbor hand-off：synthetic 2-room（同 unit 本房+邻房）。
// 铁律（非全空间监控）：仅本房丢轨（二义 lost-fall）+ 邻房**极近有向** hand-off 才压；
// stale/durable、超窗、多resident、邻房空 一律不压（漏报-safe）。

const (
	nbSuite = "fd00:0:9:201::/80"
	nbHome  = "fd00:0:9:201:a::/88" // 本房（丢轨）
	nbSib   = "fd00:0:9:201:b::/88" // 邻房
)

func TestNeighborHandoffSuppress(t *testing.T) {
	const nowMs = int64(10_000_000)
	const lostSeenMs = nowMs - 70_000 // 70s 前丢轨（>60s TTL）；hand-off 窗=[lostSeen-5s, lostSeen+60s]

	soleCensus := func() *SuiteCensusManager {
		m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
		m.GetOrCreate(nbSuite).Persons["r"] = p65Resident("r", false, card.RoomTypeBathroom)
		return m
	}
	multiCensus := func() *SuiteCensusManager {
		m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
		c := m.GetOrCreate(nbSuite)
		c.Persons["A"] = p65Resident("A", false, card.RoomTypeBathroom)
		c.Persons["B"] = p65Resident("B", false, card.RoomTypeDefault)
		return m
	}

	mkEngine := func(census *SuiteCensusManager, sib *TrackManager) (*Engine, *observer.ObservedLogs) {
		core, logs := observer.New(zapcore.InfoLevel)
		e := &Engine{
			logger: zap.New(core),
			rooms: map[string]*TrackManager{
				nbHome: {roomID: nbHome, tracks: map[int]*TrackState{}, bedCount: 1},
				nbSib:  sib,
			},
			grids:         map[string]*RoomGrid{},
			roomSuiteID:   map[string]string{nbHome: nbSuite, nbSib: nbSuite},
			beliefShadows: map[string]*beliefShadow{},
			suiteCensus:   census,
		}
		sh := e.beliefShadowFor(nbHome)
		sh.tracks[7] = &beliefShadowTrack{
			lastSeenMs:    lostSeenMs,
			stillBoxAgeMs: 0, // moving→lost（进 lost-fall 域）
			lastX:         50, lastY: 50,
			lostAnchor: lostSeenMs,
		}
		return e, logs
	}
	fired := func(logs *observer.ObservedLogs) bool {
		return logs.FilterMessage("belief_shadow_neighbor_handoff").Len() > 0
	}
	staleLogged := func(logs *observer.ObservedLogs) bool { // no-silent-caps 留驻 gap LOG(审查62)
		return logs.FilterMessage("belief_shadow_neighbor_stale_corr").Len() > 0
	}
	sib := func() *TrackManager {
		return &TrackManager{roomID: nbSib, tracks: map[int]*TrackState{}, bedCount: 1}
	}

	// ① room hand-off 在窗内 + sole → 压。
	s1 := sib()
	s1.lastEnterMs = lostSeenMs + 5_000
	e1, l1 := mkEngine(soleCensus(), s1)
	e1.beliefShadowTick(nbHome, nil, nowMs)
	if !fired(l1) {
		t.Fatalf("① 窗内 room hand-off + sole 应压")
	}

	// ② bed hand-off（接触式）在窗内 + sole → 压。
	s2 := sib()
	s2.lastRadarInBedMs = lostSeenMs + 3_000
	e2, l2 := mkEngine(soleCensus(), s2)
	e2.beliefShadowTick(nbHome, nil, nowMs)
	if !fired(l2) {
		t.Fatalf("② 窗内 bed hand-off 应压")
	}

	// ③ stale 留驻：邻房 enter 远早于本房丢轨（durable「上次在哪」）→ 不压（铁律）+ 须 stale_corr LOG（no-silent-caps）。
	s3 := sib()
	s3.lastEnterMs = lostSeenMs - 230_000
	e3, l3 := mkEngine(soleCensus(), s3)
	e3.beliefShadowTick(nbHome, nil, nowMs)
	if fired(l3) {
		t.Fatalf("③ stale/durable 占用不该压（非全空间监控：人可能穿盲区真摔）")
	}
	if !staleLogged(l3) {
		t.Fatalf("③ 留驻 gap 须 LOG belief_shadow_neighbor_stale_corr（no-silent-caps，审查62）")
	}

	// ④ 超窗：邻房 enter 晚于本房丢轨 >60s → 不压（中间可能盲区摔）+ 占用账在 → 同属 stale_corr。
	s4 := sib()
	s4.lastEnterMs = lostSeenMs + 65_000
	e4, l4 := mkEngine(soleCensus(), s4)
	e4.beliefShadowTick(nbHome, nil, nowMs)
	if fired(l4) {
		t.Fatalf("④ 超窗 hand-off 不该压")
	}
	if !staleLogged(l4) {
		t.Fatalf("④ 超窗但占用账在 → 须 stale_corr LOG")
	}

	// ⑤ multi-resident：窗内 hand-off 但 N-3 gate-OFF → 不压 + 不记 stale（归因不安全，非本特性 gap）。
	s5 := sib()
	s5.lastEnterMs = lostSeenMs + 5_000
	e5, l5 := mkEngine(multiCensus(), s5)
	e5.beliefShadowTick(nbHome, nil, nowMs)
	if fired(l5) {
		t.Fatalf("⑤ 多resident 应 gate-OFF 不压（漏报-safe）")
	}
	if staleLogged(l5) {
		t.Fatalf("⑤ 多resident gate-OFF 不该记 stale_corr（非本特性 deferred gap）")
	}

	// ⑥ 邻房空（enter 后又 exit）→ 无占用 → 不压 + 不记 stale（corroboration≠substitution）。
	s6 := sib()
	s6.lastEnterMs = lostSeenMs + 5_000
	s6.lastExitMs = lostSeenMs + 10_000
	e6, l6 := mkEngine(soleCensus(), s6)
	e6.beliefShadowTick(nbHome, nil, nowMs)
	if fired(l6) {
		t.Fatalf("⑥ 邻房已 exit（空）不该压")
	}
	if staleLogged(l6) {
		t.Fatalf("⑥ 邻房空（无占用账）不该记 stale_corr")
	}
}
