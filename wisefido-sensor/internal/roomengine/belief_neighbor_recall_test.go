package roomengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"owl-common/card"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_neighbor_recall_test.go — recall(真摔召回)真数据载体:把 nbpRun 的合成整单元升级成
// **真 unit201 fixture**(CD2B 卧室 + 333B 浴室,doc/cases/unit201-handoff-0609-*)。
// neighbor_verification_spec §2 唯一仍 blocked 项 = recall;数据 2026-06-09 导出后即可跑。
//
// 保真硬条件同 bReplay/bReplayUnit:只喂 raw fixture record 进生产 handleMessage/handleEventMessage,
// 多房注册同 suite + per-device 路由 + census seed + 按 ts 全局合并。
//
// **export 格式差异(必处理)**:unit201 window.json 的 topic_type=null,event kind 落在 category
// (EnterRoom/ExitRoom/Fall/number_people/Walking),monitor 子流 category=track/heart/activity。
// 故事件/监控分类按 category 判(非 topic_type)。
// **layout 失真(export bug)**:两房 room_layout.json 的 room_id=null 且 radar 映射都误指 cd2b →
// 不能复用 bLayout 的 radar/roomID;此处显式给 distinct roomID + RoomType + 路由 addr。

const (
	rclSuite       = "fd00:0:3:112:3:100::/80"
	rclBedroomRoom = "fd00:0:3:112:3:100:b::/88"
	rclBathRoom    = "fd00:0:3:112:3:100:a::/88"
	rclBedroomAddr = "fd00:0:3:112:3:100:32a1:cd2b" // CD2B 真 addr(:3:100:)
	rclBathAddr    = "fd00:0:3:112:3:200:59b8:333b" // 333B 真 addr(:3:200:;export layout radar 误标 cd2b,真 addr 据 DB alarm_events)
)

// rclRoom 一个真房:fixture 目录 + 注册参数 + 路由 addr。
type rclRoom struct {
	dir      string
	roomID   string
	roomType int
	addr     string
	label    string
}

// rclLoadCfg 读真 room_layout.json → cfg(几何真),覆盖 roomID/RoomType/SuiteID(export 失真字段不可信)。
func rclLoadCfg(t *testing.T, r rclRoom, suite string) RoomConfig {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(casesDir, r.dir, "room_layout.json"))
	if err != nil {
		t.Skipf("layout 缺失 %s: %v", r.dir, err)
	}
	var wrap struct {
		LayoutConfig json.RawMessage `json:"layout_config"`
	}
	body := raw
	if json.Unmarshal(raw, &wrap) == nil && len(wrap.LayoutConfig) > 0 {
		body = wrap.LayoutConfig
	}
	cfg, err := ParseLayoutConfig(r.roomID, body)
	if err != nil {
		t.Fatalf("ParseLayoutConfig %s: %v", r.dir, err)
	}
	cfg.RoomID = r.roomID
	cfg.RoomType = r.roomType
	cfg.SuiteID = suite
	return cfg
}

// rclMergedRecord 合并视图:一条 record + 它所属房的路由 addr。
type rclMergedRecord struct {
	bRecord
	addr string
}

var rclMonitorCats = map[string]bool{"track": true, "heart": true, "activity": true}

// rclReplay 跑真 unit201 整单元 recall:注册两房同 suite,按 ts 全局合并喂生产路径,返回 shadow 日志。
func rclReplay(t *testing.T, rooms []rclRoom, census *SuiteCensusManager) ([]bShadowLog, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(zapcore.DebugLevel)
	e := NewEngine(nil, zap.New(core))
	e.suiteCensus = census

	var merged []rclMergedRecord
	for _, r := range rooms {
		cfg := rclLoadCfg(t, r, rclSuite)
		e.RegisterRoom(cfg)
		e.deviceRoom[r.addr] = r.roomID
		e.deviceMounts[r.addr] = cfg.Radar
		for _, rec := range bLoadRecords(t, r.dir, "window.json") {
			merged = append(merged, rclMergedRecord{bRecord: rec, addr: r.addr})
		}
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].Timestamp < merged[j].Timestamp })

	mk := func(addr, topic, cat string, dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON, _ := json.Marshal(dv)
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": addr, "device_type": "radar", "topic_type": topic, "category": cat,
			"timestamp": strconv.FormatInt(ts, 10), "dataValue": string(dvJSON),
		}}
	}

	for _, rec := range merged {
		if rclMonitorCats[rec.Category] {
			e.handleMessage(nil, mk(rec.addr, "monitor", rec.Category, rec.DataValue, rec.Timestamp))
		} else {
			e.handleEventMessage(mk(rec.addr, "event", rec.Category, rec.DataValue, rec.Timestamp))
		}
	}

	var out []bShadowLog
	for _, entry := range logs.All() {
		out = append(out, bShadowLog{Msg: entry.Message, Fields: entry.ContextMap()})
	}
	return out, logs
}

func rclSoleCensus() *SuiteCensusManager {
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	m.GetOrCreate(rclSuite).Persons["r"] = &SuitePerson{
		PersonID: "r", Role: SuitePersonResident, AnchorRoomType: card.RoomTypeBathroom,
	}
	return m
}

// rclShadowField 拉 belief_shadow_* 日志里某字段(room scoped 由 caller filter)。
func rclSummary(t *testing.T, logs []bShadowLog) {
	t.Helper()
	peakByRoom := map[string]float64{}
	counts := map[string]int{}
	for _, l := range logs {
		counts[l.Msg]++
		if l.Msg == "belief_shadow_trace" {
			room, _ := l.Fields["room_id"].(string)
			if p, ok := l.Fields["p_fallen"].(float64); ok && p > peakByRoom[room] {
				peakByRoom[room] = p
			}
		}
	}
	t.Logf("★unit201 recall shadow 日志计数 = %v", counts)
	for room, p := range peakByRoom {
		t.Logf("  room=%s peak_P(Fallen)=%.3f", room, p)
	}
	for _, l := range logs {
		switch l.Msg {
		case "belief_shadow_fall":
			t.Logf("  [FALL] confirm fall: %v", l.Fields)
		case "belief_dbn_recovery_evidence":
			t.Logf("  [RECOVERY] %v", l.Fields)
		case "belief_shadow_neighbor_handoff", "belief_shadow_neighbor_stale_corr":
			t.Logf("  [NEIGHBOR] %s: %v", l.Msg, l.Fields)
		}
	}
}

// TestNeighborRecallUnit201 — recall 真数据 **ground-truth 已定**(用户 2026-06-10 标注:06-09
// 07:16:11 Denver 浴室 333B = **真摔**,然后人去卧室=真 hand-off)。硬断言 recall 成功。
// 场景:浴室 333B 人 Walking(101s)→**Fall(131s=T_fire,firmware pose5)**→lying→ExitRoom(230s,自救);
// 卧室 CD2B 落 fall 时空(ExitRoom@95s,下次 Enter@226s,在 fall confirm@198s 之后才 hand-off)。
// recall 关键:即便人事后 hand-off 去卧室,真摔已在浴室确认在前 → neighbor **不得**回溯压制真摔。
func TestNeighborRecallUnit201(t *testing.T) {
	rooms := []rclRoom{
		{"unit201-handoff-0609-bedroom-CD2B", rclBedroomRoom, card.RoomTypeDefault, rclBedroomAddr, "bedroom"},
		{"unit201-handoff-0609-bathroom-333B", rclBathRoom, card.RoomTypeBathroom, rclBathAddr, "bathroom"},
	}
	logs, _ := rclReplay(t, rooms, rclSoleCensus())
	rclSummary(t, logs)

	bathConfirm, bedPeak, neighborHandoff, recoveryVeto := 0, 0.0, 0, false
	for _, l := range logs {
		switch l.Msg {
		case "belief_shadow_fall":
			if rid, _ := l.Fields["room_id"].(string); rid == rclBathRoom {
				bathConfirm++
			}
		case "belief_shadow_trace":
			if rid, _ := l.Fields["room_id"].(string); rid == rclBedroomRoom {
				if p, _ := l.Fields["p_fallen"].(float64); p > bedPeak {
					bedPeak = p
				}
			}
		case "belief_shadow_neighbor_handoff":
			neighborHandoff++
		case "belief_dbn_recovery_evidence":
			if v, _ := l.Fields["would_veto"].(bool); v {
				recoveryVeto = true
			}
		}
	}

	// 真摔召回:浴室真摔必须 DBN confirm(spec §2 recall 铁律:漏报=不可接受)。
	if bathConfirm < 1 {
		t.Errorf("recall 漏报:真摔(06-09 07:16 浴室 333B)DBN 未 confirm(belief_shadow_fall=0)")
	}
	// 真摔不得被 neighbor hand-off 回溯压制(人事后去卧室≠没摔)。
	if neighborHandoff > 0 {
		t.Errorf("recall 误压:真摔被 neighbor hand-off 压制 %d 次(人事后 hand-off 卧室不该回溯压真摔)", neighborHandoff)
	}
	// recovery 不得误撤真摔(倒地≥15s 自救=genuine-fall guard 应禁 veto)。
	if recoveryVeto {
		t.Errorf("recall 误撤:真摔被 recovery-veto would_veto=true(自救≥15s 应 genuine-fall guard 禁)")
	}
	// 邻房(卧室)不得误报 fall。
	if bedPeak >= 0.55 {
		t.Errorf("邻房卧室误报:peak P(Fallen)=%.3f ≥ θ_fire 0.55", bedPeak)
	}
}
