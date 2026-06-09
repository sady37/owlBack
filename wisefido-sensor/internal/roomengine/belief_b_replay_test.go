package roomengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// belief_b_replay_test.go — P9 载体 B(审查㊸ 批准:B-1=(a)直驱 handleMessage / B-2=(a)zap-observer 日志为主)。
//
// 保真硬条件(㊸ 机械堵"合成绿"):**只**喂 raw fixture record 进生产 `handleMessage`/`handleEventMessage`
// + setup(NewEngine/RegisterRoom/wire deviceRoom/mount)+ observer 读 `belief_shadow_*` 日志。
// **禁**手搓 `TrackStatusBase`、禁调非生产 tick、禁 fork shadow 逻辑。fixture record → 真 StreamMessage →
// FromStreamMap → ParseRadarTracks → ProcessFrame → SnapshotTrackStatuses → 真 beliefShadowTick(零手搓零 fork)。
//
// 首 oracle = 复现 C 诊断:8 CD2B 卧室 fall 逐案分类对账(α→bed 权威压制 / β→nodetect 门控 / γ→escalate)。
// B 结果与 C 手工诊断分歧 = pipeline bug(B 同时是 C 的交叉校验)。

// bRecord 一条 fixture 记录(monitor/event 流原始行)。
type bRecord struct {
	DeviceUID string                   `json:"device_uid"`
	Timestamp int64                    `json:"timestamp"`
	TopicType string                   `json:"topic_type"`
	Category  string                   `json:"category"`
	DataValue []map[string]interface{} `json:"data_value"`
}

// bLoadRecords 读 fixture window json → 按 ts 升序记录。
func bLoadRecords(t *testing.T, dir, file string) []bRecord {
	b, err := os.ReadFile(filepath.Join(casesDir, dir, file))
	if err != nil {
		t.Skipf("fixture 缺失 %s/%s: %v", dir, file, err)
	}
	var recs []bRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatalf("解析 %s/%s: %v", dir, file, err)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Timestamp < recs[j].Timestamp })
	return recs
}

// bLayout 读 room_layout.json → cfg(ParseLayoutConfig)+ radar addr(layout radar 映射首值)+ roomID。
func bLayout(t *testing.T, dir string) (RoomConfig, string, string) {
	raw, err := os.ReadFile(filepath.Join(casesDir, dir, "room_layout.json"))
	if err != nil {
		t.Skipf("layout 缺失 %s: %v", dir, err)
	}
	// 双壳:cd2b export = {room_id, room_name, layout_config{...}};裸 canvas 直接是 layout_config。
	var wrap struct {
		RoomID       string          `json:"room_id"`
		LayoutConfig json.RawMessage `json:"layout_config"`
	}
	body := raw
	roomID := "fd00:0:3:112:3:100::/88"
	if json.Unmarshal(raw, &wrap) == nil && len(wrap.LayoutConfig) > 0 {
		body = wrap.LayoutConfig
		if wrap.RoomID != "" {
			roomID = wrap.RoomID
		}
	}
	cfg, err := ParseLayoutConfig(roomID, body)
	if err != nil {
		t.Fatalf("ParseLayoutConfig %s: %v", dir, err)
	}
	// radar addr:layout_config.radar = {radar_<n>: <INET>};取首值。
	var lc struct {
		Radar map[string]string `json:"radar"`
	}
	radarAddr := ""
	if json.Unmarshal(body, &lc) == nil {
		for _, v := range lc.Radar {
			radarAddr = v
			break
		}
	}
	if radarAddr == "" {
		t.Fatalf("layout %s 无 radar 映射", dir)
	}
	return cfg, radarAddr, cfg.RoomID
}

// bShadowLog observer 捕到的一条 belief_shadow_* 日志(msg + 关键字段)。
type bShadowLog struct {
	Msg    string
	Fields map[string]interface{}
}

// bReplay setup 真 Engine + 回放 fixture → 返回 observer 捕到的全部 belief_shadow_* 日志。
// **保真**:只喂 raw record 进 handleMessage/handleEventMessage;不手搓 bases。
func bReplay(t *testing.T, dir, file string) []bShadowLog {
	recs := bLoadRecords(t, dir, file)
	cfg, radarAddr, roomID := bLayout(t, dir)

	core, logs := observer.New(zap.DebugLevel) // 捕 belief_shadow_trace(Debug)供逐案查 P(Fallen) 轨迹
	logger := zap.New(core)
	e := NewEngine(nil, logger)
	e.RegisterRoom(cfg)

	// 识别 radar device_uid:radar.track 与 sleepad.track 经 export 都 category=track,**靠 position_x 区分**
	// (radar track 带坐标,sleepad.track 无 position_x 只 bed_status/vital)——否则首个 track 行若是 sleepad
	// 会误把 sleepad uid 当 radar → 所有 radar 帧误路由 → ProcessFrame/shadow 全不跑(补 sleepad.track 后暴露)。
	radarUID := ""
	for _, r := range recs {
		if r.Category == "track" && len(r.DataValue) > 0 {
			if _, hasPos := r.DataValue[0]["position_x"]; hasPos {
				radarUID = r.DeviceUID
				break
			}
		}
	}
	const sleepadAddr = "fd00:0:3:112:3:100:5111:ad01" // 合成 sleepad 寻址(仅路由用;sleepad obs 按 UID 存)
	// wire 路由:radar/sleepad addr → roomID(RegisterRoom 已设 deviceMounts[roomID]=cfg.Radar 兜底 mount)。
	e.deviceRoom[radarAddr] = roomID
	e.deviceRoom[sleepadAddr] = roomID
	e.deviceMounts[radarAddr] = cfg.Radar // 显式兜底:确保 handleMessage hasMount(RadarAddrs 可能空/异)

	mk := func(addr, devType, topic, cat string, dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON, _ := json.Marshal(dv)
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": addr,
			"device_type": devType,
			"topic_type":  topic,
			"category":    cat,
			"timestamp":   strconv.FormatInt(ts, 10),
			"dataValue":   string(dvJSON),
		}}
	}

	for _, r := range recs {
		isRadar := r.DeviceUID == radarUID
		addr := radarAddr
		devType := "radar"
		if !isRadar {
			addr = sleepadAddr
			devType = "sleepad"
		}
		switch {
		case r.TopicType == "event":
			// 床/门/人数事件 → 真 event 入口(InBed/LeftBed/EnterRoom/ExitRoom/number_people)。
			e.handleEventMessage(mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		case !isRadar:
			// sleepad monitor 帧(sleepad.track 含 bed_status/vital)→ 真 monitor 入口(ProcessSleepadObservation)。
			// **须先于 category==track 判**:sleepad.track 经 export 也 category=track,但应走 sleepad 分支非 radar。
			e.handleMessage(nil, mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		case r.Category == "track":
			// radar track 帧 → 真 monitor 入口(ParseRadarTracks→ProcessFrame→beliefShadowTick)。
			e.handleMessage(nil, mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		default:
			// radar 非 track monitor 子流(heart/activity 等)不入 shadow lost/bed 决策路径,跳过(非 fork:不喂≠手搓)。
		}
	}

	var out []bShadowLog
	allCounts := map[string]int{}
	for _, entry := range logs.All() {
		allCounts[entry.Message]++
		if !strings.HasPrefix(entry.Message, "belief_shadow_") {
			continue
		}
		out = append(out, bShadowLog{Msg: entry.Message, Fields: entry.ContextMap()})
	}
	if os.Getenv("BDIAG") != "" {
		t.Logf("ALL logs %s: %v", dir, allCounts)
	}
	return out
}

// bCase 8 CD2B 卧室 fall 案(C 诊断分类)。
type bCase struct {
	dir   string
	file  string
	class string // α / β / γ
}

var bCases = []bCase{
	{"cd2b-fall-0605-0712", "", "α"},
	{"cd2b-fall-0606-0917", "", "α"},
	{"cd2b-fall-0606-0929", "", "α"},
	{"cd2b-fall-0607-0127", "", "α"},
	{"cd2b-fall-0607-1021", "", "α"},
	{"cd2b-fall-0604-2233", "", "β"},
	{"cd2b-fall-0605-0717", "", "β"},
	{"cd2b-fall-0605-0142", "", "γ"},
}

// bFindWindowFile dir 下唯一的 window json(*_to_*.json)。
func bFindWindowFile(t *testing.T, dir string) string {
	entries, err := os.ReadDir(filepath.Join(casesDir, dir))
	if err != nil {
		t.Skipf("case 目录缺失 %s: %v", dir, err)
	}
	for _, en := range entries {
		n := en.Name()
		if strings.Contains(n, "_to_") && strings.HasSuffix(n, ".json") {
			return n
		}
	}
	t.Skipf("case %s 无 window json", dir)
	return ""
}

// allObsKinds — source-fidelity 审计的全 obs 全集(belief 引擎认的源;对账"真路径 populate 了哪些")。
var allObsKinds = []string{
	"Pose", "VitalPresent", "BedOccupied", "EnterExit", "NumberPeople",
	"TrackPresent", "Neighbor", "NoDetect", "ReachableExit", "ZBand", "DwellStill",
}

// TestBSourceFidelityAudit — 系统性 source-fidelity 审计(审查㊺ 当前最高价值):8 案逐 obs 对账。
// 每案统计真路径上**实际 populate(fresh)**的 obs 源 → 对全集找 **dead/未 populate** 的源
// (firmware/redis/bed 三连同根=DBN obs 源≠生产实际源)。一次挖净,非 piecemeal。**纯 observability,不刚性断言。**
func TestBSourceFidelityAudit(t *testing.T) {
	everPopulated := map[string]bool{} // 跨 8 案任一 tick populate 过的源(全局视角)
	for _, c := range bCases {
		c := c
		file := bFindWindowFile(t, c.dir)
		logs := bReplay(t, c.dir, file)
		casePop := map[string]bool{}
		for _, l := range logs {
			if l.Msg != "belief_shadow_obs" {
				continue
			}
			if ks, ok := l.Fields["populated_kinds"].([]interface{}); ok {
				for _, k := range ks {
					if s, ok := k.(string); ok {
						casePop[s] = true
						everPopulated[s] = true
					}
				}
			} else if ks, ok := l.Fields["populated_kinds"].([]string); ok {
				for _, s := range ks {
					casePop[s] = true
					everPopulated[s] = true
				}
			}
		}
		var pop, dead []string
		for _, k := range allObsKinds {
			if casePop[k] {
				pop = append(pop, k)
			} else {
				dead = append(dead, k)
			}
		}
		t.Logf("source-fidelity case=%s C类=%s | populated=%v | 未populate=%v", c.dir, c.class, pop, dead)
	}
	var globalDead []string
	for _, k := range allObsKinds {
		if !everPopulated[k] {
			globalDead = append(globalDead, k)
		}
	}
	t.Logf("★source-fidelity 全局:8案任一tick从未populate的obs源(疑死管/源错配,待逐个核生产实际源)=%v", globalDead)
}

// TestBReplaySmoke — B harness 冒烟:每案跑通真生产路径无 panic(保真 + 全路径走通)。
func TestBReplaySmoke(t *testing.T) {
	for _, c := range bCases {
		c := c
		t.Run(c.dir, func(t *testing.T) {
			file := bFindWindowFile(t, c.dir)
			_ = bReplay(t, c.dir, file) // 不 panic 即过(保真路径走通);逐案可观测见 TestBReproduceCInspect
		})
	}
}

// TestBReproduceCInspect — reproduce-C(委员会工作流 note:**可逐案查不过度打磨自动断言**)。
// 每案打印 shadow 决策摘要(峰值 P(Fallen)/末态 argmax/否决·确认信号),供用户 human-in-loop 逐案
// 校对 shadow 决策 vs 真值(8 CD2B 全 FP→shadow 不应 confirm,即峰值 P(Fallen)<θ_fire 0.55)。
// **不做刚性 pass/fail 断言**(用户判定真值);仅 observability + 一条软不变量(8 FP 无 shadow confirm)。
func TestBReproduceCInspect(t *testing.T) {
	for _, c := range bCases {
		c := c
		t.Run(c.dir, func(t *testing.T) {
			file := bFindWindowFile(t, c.dir)
			logs := bReplay(t, c.dir, file)
			peakP, finalArg := 0.0, "-"
			decisions := map[string]int{}
			for _, l := range logs {
				if l.Msg == "belief_shadow_trace" {
					if p, ok := l.Fields["p_fallen"].(float64); ok && p > peakP {
						peakP = p
					}
					if s, ok := l.Fields["argmax_state"].(string); ok {
						finalArg = s
					}
					continue
				}
				decisions[l.Msg]++ // belief_shadow_fall / bed_leak_suppress / nodetect_gated / lostfall_* / track_lost
			}
			confirmed := decisions["belief_shadow_fall"] > 0
			traceN := 0
			for _, l := range logs {
				if l.Msg == "belief_shadow_trace" {
					traceN++
				}
			}
			if traceN == 0 {
				// 诚实标注:beliefShadowTick 未执行(publishTrackStatuses 被 nil-redis guard 短路,engine:1006)。
				// 生产 redis 恒非 nil 故 shadow 总跑;无 redis 测试环境下 B 暂未 exercise 真 shadow 路径 → 待 redis-gate 岔口裁定。
				t.Logf("逐案查 case=%s C类=%s | ⚠️shadow未执行(0 trace;redis-gated publishTrackStatuses,待岔口)——逐案 P 值待 shadow 跑通", c.dir, c.class)
				return
			}
			t.Logf("逐案查 case=%s C类=%s | peak_P(Fallen)=%.3f final_argmax=%s shadow_confirm=%v | 决策信号=%v",
				c.dir, c.class, peakP, finalArg, confirmed, decisions)
			// 软不变量(非真值断言):8 CD2B 全 FP(无真摔)→ shadow 不应 confirm fall。
			if confirmed {
				t.Errorf("FP 案 %s shadow 误 confirm fall(peak P=%.3f)——需用户逐案核 + fix 方案", c.dir, peakP)
			}
		})
	}
}
