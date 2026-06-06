package roomengine

// gate_replay_test.go — 用真实录制 case 驱动 GATE 路径（TrackManager：processFrameAt +
// NoTargetTick + ProcessSleepadBedEvent + RecordRadarEvent），断言止血侧规则引擎的 fall 行为。
// 与 belief_replay_test（喂 belief/DBN）互补：那条验信念层，这条验 gate 驱逐 → silent/lost vanish。
//
// 复用 belief_replay_test.go 的 loadFixture / buildGridFromLayout / mi。

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"owl-common/alarm"
	"owl-common/radarutils"
)

// isSleepadUID sleepad device_uid 以 "BM" 开头（厂家 logMAC）；radar 是 hex。
func isSleepadUID(uid string) bool {
	return strings.HasPrefix(strings.ToUpper(uid), "BM")
}

// captureAIPub 捕获 TrackManager emit 的 sensor_decision / alarm（回放里抓 lost-fall 抑制 reason）。
type captureAIPub struct {
	decisions []string // "event:reason"
	alarms    []string // alarm event
}

func (c *captureAIPub) PublishAIEvent(_ context.Context, p AIPayload, _ string, nowMs int64) {
	c.decisions = append(c.decisions, fmt.Sprintf("%s:%s@%d", p.Event, p.Reason, nowMs))
}
func (c *captureAIPub) PublishAIAlarm(_ context.Context, p AIPayload, _ string, _ int64) {
	c.alarms = append(c.alarms, p.Event+":"+p.Reason)
}
func (c *captureAIPub) DeviceUIDHex(string) string { return "" }

// gateReplay 把 fixture record 按 engine.go handleMessage 的路由喂进 TrackManager。
//   track 帧(tid 0-8) → processFrameAt；tid=88 心跳 → NoTargetTick（1+2 驱逐修复路径）
//   InBed/LeftBed：sleepad → ProcessSleepadBedEvent；radar → RecordRadarEvent
//   EnterRoom/ExitRoom/number_people → RecordRadarEvent
func gateReplay(t *testing.T, recs []fxRecord, grid *RoomGrid, mount radarutils.RadarMount, radarAddr string, bedCount int, pub AIPublisher) *TrackManager {
	t.Helper()
	tm := NewTrackManager("test-room", grid)
	tm.bedCount = bedCount
	if pub != nil {
		tm.SetAIPublisher(pub) // 注入捕获器抓 emitDecision 的 reason
	}

	for _, r := range recs {
		now := r.Timestamp
		switch r.Category {
		case "track":
			var frames []TrackFrame
			heartbeat := false
			for _, dv := range r.DataValue {
				id, _ := mi(dv, "track_id")
				if id == 88 { // firmware 无目标心跳
					heartbeat = true
					continue
				}
				if id < 0 || id > 8 {
					continue
				}
				rawX, _ := mi(dv, "position_x")
				rawY, _ := mi(dv, "position_y")
				rawZ, _ := mi(dv, "position_z")
				if rawX == 0 && rawY == 0 && rawZ == 0 {
					continue
				}
				canvas := radarutils.RadarToCanvas(radarutils.RadarPoint{H: rawX, V: rawY, Z: rawZ}, mount)
				pose, _ := mi(dv, "pose")
				frames = append(frames, TrackFrame{
					TrackID: id, DeviceAddr: radarAddr,
					X: canvas.X, Y: canvas.Y, Z: canvas.Z, Pose: pose, TMs: now,
				})
			}
			switch {
			case len(frames) > 0:
				tm.processFrameAt(frames, now)
				// replay 不跑 ghost adjudicator → 强制升 Real（与 belief_replay 同口径）
				for _, f := range frames {
					if ts, ok := tm.tracks[f.TrackID]; ok {
						ts.Verdict = VerdictReal
						ts.Score = ScoreConfirmTh
					}
				}
			case heartbeat:
				tm.NoTargetTick(now) // 88 心跳 → 无目标 tick（触发 88-加速驱逐）
			default:
				tm.processFrameAt(nil, now)
			}

		case alarm.InBed, alarm.LeftBed:
			if isSleepadUID(r.DeviceUID) {
				tm.ProcessSleepadBedEvent(SleepadBedEvent{
					DeviceUID: r.DeviceUID, IsInBed: r.Category == alarm.InBed, TMs: now, Status: "instant",
				})
			} else {
				tm.RecordRadarEvent(RadarTrackEvent{
					DeviceUID: r.DeviceUID, EventName: r.Category, TMs: now, Status: "instant",
				})
			}

		case alarm.EnterRoom, alarm.ExitRoom:
			tm.RecordRadarEvent(RadarTrackEvent{
				DeviceUID: r.DeviceUID, EventName: r.Category, TMs: now, Status: "instant",
			})

		case "number_people":
			np := 0
			for _, dv := range r.DataValue {
				if n, ok := mi(dv, "number_people"); ok {
					np = n
				}
			}
			tm.RecordRadarEvent(RadarTrackEvent{
				DeviceUID: r.DeviceUID, EventName: EventNameNumberPeople, NumberPeople: np, TMs: now, Status: "instant",
			})
		}
	}
	return tm
}

// TestGateReplay_Case2Quilt_FiresAfterFastEvict 用真实 Case2 录制验 1+2 驱逐修复：
// 被子真 track 22:23:58 丢 → 88 心跳 → 6-12s 快速驱逐 → silent 窗（LeftBed+120s）时床边无 track
// → vanish-fire。旧码（仅 MissCount=10，稀疏心跳下 ~5min 才驱逐）会让 silent 窗仍看到陈旧被子
// → human-bed 豁免 → 不报：本用例正是区分 old(0) vs new(≥1) 的真值回归。
func TestGateReplay_Case2Quilt_FiresAfterFastEvict(t *testing.T) {
	dir := "case2-quilt-bedside-fall-0604"
	recs := loadFixture(t, dir+"/2026-06-04_16-14_to_16-31_MDT.json")
	grid, mount := buildGridFromLayout(t, dir+"/room_layout.json")

	const radarAddr = "fd00:0:3:112:3:100:32a1:cd2b"
	tm := gateReplay(t, recs, grid, mount, radarAddr, 1, nil)

	if tm.silentFallLeftbedReported == 0 {
		t.Errorf("Case2 quilt replay: expected silent_leftbed fall (vanish after fast evict), got reported=0 "+
			"(cancelled=%d) — 88-加速驱逐失效 / 陈旧被子仍命中 human-bed 豁免?", tm.silentFallLeftbedCancelled)
	}
}

// TestGateReplay_CD2BLostFall_Observe 实测回放 case_lostfall_cd2b_11351148（真摔漏报 fixture）：
// 人 InBed → LeftBed(疑似跌倒) → radar frozen 5.5min → 11:47 重新出现。合并 radar+sleepad 双源按时间回放，
// 打出当前引擎对 lost-fall / silent_leftbed 的实际处置（created/reported/cancelled），观测当前行为。
func TestGateReplay_CD2BLostFall_Observe(t *testing.T) {
	dir := "case_lostfall_cd2b_11351148"
	recs := loadFixture(t, dir+"/2026-04-27_11-35_to_11-48_MDT.json")
	slp := loadFixture(t, dir+"/sleepad_BM1641_2026-04-27_11-35_to_11-48_MDT.json")
	recs = append(recs, slp...)
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Timestamp < recs[j].Timestamp })

	grid, mount := buildGridFromLayout(t, dir+"/room_layout.json")
	const radarAddr = "fd00:0:3:112:3:100:32a1:cd2b"
	cap := &captureAIPub{}
	tm := gateReplay(t, recs, grid, mount, radarAddr, 1, cap)

	t.Logf("cd2b lostfall replay: lostFall created=%d reported=%d cancelled=%d | silent_leftbed reported=%d cancelled=%d",
		tm.lostFallPendingCreated, tm.lostFallReported, tm.lostFallPendingCancelled,
		tm.silentFallLeftbedReported, tm.silentFallLeftbedCancelled)
	t.Logf("  decisions(event:reason)=%v", cap.decisions)
	t.Logf("  alarms=%v", cap.alarms)
}
