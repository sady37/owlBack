package roomengine

import (
	"encoding/json"
	"strconv"
	"testing"

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

const (
	nbpSuite  = "fd00:0:9:201::/80"
	nbpRoomB  = "fd00:0:9:201:b::/88"
	nbpRadarA = "fd00:0:9:201:a::1"
	nbpRadarB = "fd00:0:9:201:b::1"
)

// nbpRun 跑一遍整单元全-pipeline:本房 A radar 丢轨 + 兄弟房 B EnterRoom hand-off,
// census 由 caller 决定(sole→应 fire / multi→gate-OFF 不 fire)。返回 handoff log 次数。
func nbpRun(t *testing.T, census *SuiteCensusManager) int {
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

	t0 := int64(1_780_000_000_000)
	// 本房 A：5 帧移动 track #1（in-FOV，过 birth grace），lastSeen=t0+4s。
	track(nbpRadarA, 1, 0, 120, t0)
	track(nbpRadarA, 1, 20, 140, t0+1000)
	track(nbpRadarA, 1, 50, 170, t0+2000)
	track(nbpRadarA, 1, 80, 200, t0+3000)
	track(nbpRadarA, 1, 110, 230, t0+4000)
	// 兄弟房 B：EnterRoom @ t0+5s（在 A 丢轨窗 [lastSeen−5s, +60s] 内）。
	event(nbpRadarB, "EnterRoom", t0+5000)
	// 本房 A：t0+66s decoy track(id=2,≤8)→ track#1 absent 超 TTL(62s>60s)→ 真 lost-sweep → neighbor feed。
	track(nbpRadarA, 2, 10, 100, t0+66000)

	return logs.FilterMessage("belief_shadow_neighbor_handoff").Len()
}

func nbpSoleCensus() *SuiteCensusManager {
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	m.GetOrCreate(nbpSuite).Persons["r"] = &SuitePerson{PersonID: "r", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeBathroom}
	return m
}

// TestNeighborFullPipelineFires — sole-resident:全-pipeline 应 fire ObsNeighbor（本房丢轨 + 兄弟房窗内 hand-off）。
func TestNeighborFullPipelineFires(t *testing.T) {
	if n := nbpRun(t, nbpSoleCensus()); n == 0 {
		t.Fatalf("sole-resident 全-pipeline 应 fire ObsNeighbor(belief_shadow_neighbor_handoff),得 0")
	}
}

// TestNeighborFullPipelineMultiResidentGateOff — multi-resident:全-pipeline N-3 gate-OFF 不 fire（漏报-safe）。
func TestNeighborFullPipelineMultiResidentGateOff(t *testing.T) {
	cm := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	c := cm.GetOrCreate(nbpSuite)
	c.Persons["A"] = &SuitePerson{PersonID: "A", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeBathroom}
	c.Persons["B"] = &SuitePerson{PersonID: "B", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeDefault}
	if n := nbpRun(t, cm); n != 0 {
		t.Fatalf("multi-resident 全-pipeline 应 gate-OFF 不 fire(漏报-safe),却 fire %d 次", n)
	}
}
