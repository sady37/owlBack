package roomengine

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"owl-common/alarm"
	"owl-common/card"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_neighbor_pipeline_test.go — bReplayUnit 整单元(multi-room 同 suite)全-pipeline 证明:
// raw stream record → 生产 handleMessage/handleEventMessage → ParseRadarTracks/ProcessFrame →
// SnapshotTrackStatuses → 真 beliefShadowTick → 真 lost-sweep → ObsNeighbor fire(非直驱 tick 单测)。
// 借真 cabb layout 几何(FOV/mount),喂**合成 in-FOV 帧**(时序可控→丢轨时刻确定)。
// 覆盖 spec V1-V4(doc/neighbor_verification_spec.md):V1 fire / V2 stale_corr / V3 N-6 单合并 / V4 gate-OFF。
// recall(真摔召回)合成证不了,仍 blocked on unit201 真数据(spec §2)。

const (
	nbpSuite  = "fd00:0:9:201::/80"
	nbpRoomB  = "fd00:0:9:201:b::/88"
	nbpRadarA = "fd00:0:9:201:a::1"
	nbpRadarB = "fd00:0:9:201:b::1"
	nbpT0     = int64(1_780_000_000_000)
)

type nbpEvt struct {
	name string
	ts   int64
}

// nbpRun 跑一遍整单元全-pipeline:本房 A radar 走动后丢轨(lastSeen=t0+4s,t0+66s decoy→超 60s TTL→真
// lost-sweep)+ 兄弟房 B 的 caller 指定事件(hand-off / durable / bed)。census 决定 N-3 门。返回 belief_shadow_* 日志。
func nbpRun(t *testing.T, census *SuiteCensusManager, sibEvents []nbpEvt) []bShadowLog {
	t.Helper()
	dir := "hunzi-cabb-lost-0601-2247-FP" // 仅借真 layout 几何
	cfgA, _, roomA := bLayout(t, dir)
	cfgB, _, _ := bLayout(t, dir)
	cfgB.RoomID = nbpRoomB
	cfgA.SuiteID = nbpSuite
	cfgB.SuiteID = nbpSuite

	core, logs := observer.New(zapcore.DebugLevel)
	e := NewEngine(nil, zap.New(core))
	e.RegisterRoom(cfgA)
	e.RegisterRoom(cfgB)
	e.deviceRoom[nbpRadarA] = roomA
	e.deviceRoom[nbpRadarB] = nbpRoomB
	e.deviceMounts[nbpRadarA] = cfgA.Radar
	e.deviceMounts[nbpRadarB] = cfgB.Radar
	e.suiteCensus = census

	mk := func(addr, topic, cat string, dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON, _ := json.Marshal(dv)
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": addr, "device_type": "radar", "topic_type": topic, "category": cat,
			"timestamp": strconv.FormatInt(ts, 10), "dataValue": string(dvJSON),
		}}
	}
	track := func(addr string, tid, x, y int, ts int64) {
		e.handleMessage(nil, mk(addr, "monitor", "track", []map[string]interface{}{
			{"track_id": tid, "position_x": x, "position_y": y, "position_z": 0, "pose": 4, "area_id": 255, "track_confidence": 80},
		}, ts))
	}
	event := func(addr, name string, ts int64) {
		e.handleEventMessage(mk(addr, "event", name, []map[string]interface{}{
			{"event_since": ts, "event_status": "start", "heart_rate": -1, "respiratory_rate": -1},
		}, ts))
	}

	t0 := nbpT0
	// 本房 A：5 帧移动 track #1（in-FOV，过 birth grace），lastSeen=t0+4s。
	track(nbpRadarA, 1, 0, 120, t0)
	track(nbpRadarA, 1, 20, 140, t0+1000)
	track(nbpRadarA, 1, 50, 170, t0+2000)
	track(nbpRadarA, 1, 80, 200, t0+3000)
	track(nbpRadarA, 1, 110, 230, t0+4000)
	// 兄弟房 B：caller 指定事件（在 decoy 前喂 → lastEnterMs/bed 在 lost-sweep 时已就绪）。
	for _, ev := range sibEvents {
		event(nbpRadarB, ev.name, ev.ts)
	}
	// 本房 A：t0+66s decoy track(id=2,≤8)→ track#1 absent 超 TTL(62s>60s)→ 真 lost-sweep → neighbor feed。
	track(nbpRadarA, 2, 10, 100, t0+66000)

	var out []bShadowLog
	for _, entry := range logs.All() {
		if strings.HasPrefix(entry.Message, "belief_shadow_") {
			out = append(out, bShadowLog{Msg: entry.Message, Fields: entry.ContextMap()})
		}
	}
	return out
}

func nbpCount(logs []bShadowLog, msg string) int {
	n := 0
	for _, l := range logs {
		if l.Msg == msg {
			n++
		}
	}
	return n
}

func nbpHandoffSrc(logs []bShadowLog) string {
	s := ""
	for _, l := range logs {
		if l.Msg == "belief_shadow_neighbor_handoff" {
			s, _ = l.Fields["n3_neighbor_src"].(string)
		}
	}
	return s
}

func nbpSoleCensus() *SuiteCensusManager {
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	m.GetOrCreate(nbpSuite).Persons["r"] = &SuitePerson{PersonID: "r", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeBathroom}
	return m
}

func nbpMultiCensus() *SuiteCensusManager {
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	c := m.GetOrCreate(nbpSuite)
	c.Persons["A"] = &SuitePerson{PersonID: "A", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeBathroom}
	c.Persons["B"] = &SuitePerson{PersonID: "B", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeDefault}
	return m
}

// TestNeighborFullPipelineFires — V1 sole-resident:全-pipeline 应 fire ObsNeighbor（本房丢轨 + 兄弟房窗内 hand-off）。
func TestNeighborFullPipelineFires(t *testing.T) {
	logs := nbpRun(t, nbpSoleCensus(), []nbpEvt{{alarm.EnterRoom, nbpT0 + 5000}})
	if n := nbpCount(logs, "belief_shadow_neighbor_handoff"); n != 1 {
		t.Fatalf("V1：sole-resident 全-pipeline 应 fire ObsNeighbor 恰 1 次，得 %d", n)
	}
}

// TestNeighborFullPipelineMultiResidentGateOff — V4 multi-resident:N-3 gate-OFF 不 fire 不记 stale（漏报-safe）。
func TestNeighborFullPipelineMultiResidentGateOff(t *testing.T) {
	logs := nbpRun(t, nbpMultiCensus(), []nbpEvt{{alarm.EnterRoom, nbpT0 + 5000}})
	if n := nbpCount(logs, "belief_shadow_neighbor_handoff"); n != 0 {
		t.Fatalf("V4：multi-resident 全-pipeline 应 gate-OFF 不 fire(漏报-safe),却 fire %d 次", n)
	}
	if n := nbpCount(logs, "belief_shadow_neighbor_stale_corr"); n != 0 {
		t.Fatalf("V4：multi-resident gate-OFF 不该记 stale_corr(非本特性 deferred gap),却记 %d", n)
	}
}

// TestNeighborFullPipelineStaleCorr — V2 stale_corr no-silent-caps：兄弟房 durable 占用（enter 远早于丢轨，留驻）
// 但相关度 stale → 不压（铁律：证不了此刻在哪）+ LOG belief_shadow_neighbor_stale_corr（不静默吞缺口）。
func TestNeighborFullPipelineStaleCorr(t *testing.T) {
	logs := nbpRun(t, nbpSoleCensus(), []nbpEvt{{alarm.EnterRoom, nbpT0 - 200000}}) // 丢轨前 200s = durable 留驻
	if n := nbpCount(logs, "belief_shadow_neighbor_handoff"); n != 0 {
		t.Fatalf("V2：durable 留驻不该压(不该 fire handoff)，却 fire %d 次", n)
	}
	if n := nbpCount(logs, "belief_shadow_neighbor_stale_corr"); n != 1 {
		t.Fatalf("V2：留驻 gap 须 LOG belief_shadow_neighbor_stale_corr(no-silent-caps)，得 %d", n)
	}
}

// TestNeighborFullPipelineSingleMerge — V3 N-6 单合并。两段：
// (a) bed-only 窗内 → handoff fire 且 src=bed（证 bed 占用经真路径注册，非退化成 room）；
// (b) room+bed 同窗双占用 → handoff 恰 1 次（OR-合并单条 occ，非 room/bed 各喂一条 → 无双似然过压）；conf=max→src=room。
func TestNeighborFullPipelineSingleMerge(t *testing.T) {
	bedOnly := nbpRun(t, nbpSoleCensus(), []nbpEvt{{alarm.InBed, nbpT0 + 8000}})
	if n := nbpCount(bedOnly, "belief_shadow_neighbor_handoff"); n != 1 {
		t.Fatalf("V3(a)：bed-only 窗内应 fire handoff，得 %d", n)
	}
	if src := nbpHandoffSrc(bedOnly); src != "bed" {
		t.Fatalf("V3(a)：bed-only 的 n3_neighbor_src 期望 bed(证 bed 占用经真路径注册)，得 %q", src)
	}

	both := nbpRun(t, nbpSoleCensus(), []nbpEvt{{alarm.InBed, nbpT0 + 8000}, {alarm.EnterRoom, nbpT0 + 9000}})
	if n := nbpCount(both, "belief_shadow_neighbor_handoff"); n != 1 {
		t.Fatalf("V3(b)：room+bed 双占用应 OR-合并为单次 handoff(N-6 单合并非双似然)，得 %d 次", n)
	}
	if src := nbpHandoffSrc(both); src != "room" {
		t.Fatalf("V3(b)：conf=max 应取 room(0.8)>bed(0.3)，src 期望 room，得 %q", src)
	}
}
