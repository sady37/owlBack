package roomengine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// rawDumpRow d523 类 raw monitor_stream dump 行：{ts(ISO),device_addr,stream_type,payload}。
type rawDumpRow struct {
	Ts         string                   `json:"ts"`
	StreamType string                   `json:"stream_type"`
	Payload    []map[string]interface{} `json:"payload"`
}

// loadRawDump 读 monitor_stream_*.json（DB row dump），按 ts 升序。
func loadRawDump(t *testing.T, dir string) []rawDumpRow {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(casesDir, dir))
	if err != nil {
		t.Skipf("dir 缺 %s: %v", dir, err)
	}
	var file string
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), "monitor_stream") && strings.HasSuffix(en.Name(), ".json") {
			file = en.Name()
			break
		}
	}
	if file == "" {
		t.Skipf("%s 无 monitor_stream_*.json", dir)
	}
	raw, err := os.ReadFile(filepath.Join(casesDir, dir, file))
	if err != nil {
		t.Fatalf("读 dump %s: %v", file, err)
	}
	var rows []rawDumpRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("解析 dump %s: %v", file, err)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Ts < rows[j].Ts })
	return rows
}

// parseDumpTs 解析 dump ts（ISO，可能带小数/时区）→ UTC epoch ms。
func parseDumpTs(s string) (int64, bool) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999", "2006-01-02T15:04:05"} {
		if tt, err := time.Parse(layout, s); err == nil {
			return tt.UTC().UnixMilli(), true
		}
	}
	return 0, false
}

// belief_veto_harness_test.go — 否决精度 harness（委员会上线唯一 gate：否决精度 ≥95%）。
// R0 只读：firmware-fire 真案上跑 DBN belief shadow，采集**正否决证据**算 would-veto。
//   ★委员会实质1：would-veto **须正证据**（ghost verdict / frozen / bed-occupancy），**不是 P<阈**——
//     「DBN 对 fall 没信心(中 P)」≠「有 ghost 正证据」；中 P 无证据 → **默认放行**（防误否真摔破 95% gate）。
//   正确否决 = would-veto ∧ false-alarm；错误否决 = would-veto ∧ real-fall（破 gate）。
//   否决精度 = 正确否决 / 全部否决 ≥ 95%（需 500+ 样本；本骨架样本<5，闭环非定量，不声称数字）。
// 不碰生产开火（R0）。**实质2 待补**：延时窗 5min（窗内无证据→放行 / 早退）+ 双轴（覆盖+精度）+ 窗长参数化。

type vetoCase struct {
	dir     string
	format  string // v1=带topic_type window / v2=不带(category推断) / raw=monitor dump / txt=test_record.txt
	truth   string // real-fall(firmware对,DBN不该否决) / false-alarm(firmware错,DBN该否决)
	note    string
	dayUTC  string // txt 格式:窗口日期(UTC,如"2026-06-05")
	sleepad string // txt 格式:sleepad addr(:978 路由)
}

// 首批：能直接 load 的两案（其余格式 d523 raw / d5f7 txt 下周期补统一 loader）。
var vetoCases = []vetoCase{
	{"unit201-handoff-0609-bathroom-333B", "v2", "real-fall", "#9 firmware 真摔(pose=5)→该不否决", "", ""},
	{"cd2b-fall-0607-0127", "v1", "false-alarm", "#7 firmware 在床误报→该否决", "", ""},
	{"hunzi-cabb-lost-0601-2247-FP", "v2", "false-alarm", "#5 lost_track 误报→该否决", "", ""},
	// ★委员会 gate-critical:#2 = 床边跌倒身靠床→sleepad 检 HR/RR = **床占用为真的真摔**。验 bed-veto 是否误否。
	{"bedtest-0605-2-bedside-fall-fw-detect", "txt", "real-fall", "#2 床边真摔(pose5)+床占用→bed-veto 不得误否", "2026-06-05", "fd00:0:3:111:3:101:2470:978"},
	// ghost/frozen 真案（覆盖唯一来源，委员会：bed/lost 不能覆盖）——cabb frozen-sit ghost FP，v1 格式+真 layout：
	{"cabb-ghost-frozen-sit-0415", "v1", "false-alarm", "frozen-sit ghost 误报→该否决(ghost/frozen 证据)", "", ""},
	{"cabb-ghost-frozen-sit-2117", "v1", "false-alarm", "frozen-sit ghost 误报→该否决(ghost/frozen 证据)", "", ""},
	// #13 d523-ghost：raw loader 已就绪，但 d523 的 room_visual_layout.canvas **无 Radar object**
	// （只 Wall/Enter/Furniture）→ ParseLayoutConfig 建不出 mount → pipeline 跑不了。待补带雷达 mount 的
	// d523 layout（不得捏造）后启用。raw 格式 loader 本身已验（hunzi v2 + 此 raw 路径）。
}

// vetoEvidence 一案跑完 DBN 后采集的**正否决证据**（委员会实质1：否决须正证据，非 P<阈）。
// ★安全螺丝（be229fd 四）：恢复只认**正向**证据；track 消失/lostfall_escalate **绝不**算否决（可能昏迷重伤盲区）。
type vetoEvidence struct {
	peak     float64 // peak P(Fallen)（仅诊断，不作判据）
	ghost    bool    // belief_shadow_track_lost argmax=Ghost（镜面/反射伪迹，definitively 非摔）
	bed      bool    // belief_shadow_bed_occupied_suppress 触发（**单独不够**，须配 bedConf 确凿在床）
	bedConf  float64 // 床占用最高 conf（确凿在床判据：sleepad 0.9/human-bed 0.99 ≫ radar 0.6「靠床边」）
	recovery bool    // 正向恢复：recapture(回床/离场)/neighbor-handoff(人证在邻房)——非 track 消失
	escalate bool    // belief_shadow_lostfall_escalate（窗到未佐证=真摔，**否决逻辑不得覆盖**，仅记防误否）
	argmax   string  // 诊断:ghost 证据来源（Artifact/TGhost）
	frames   int

	maxGhostness float64 // belief_shadow_veto_evidence 的 track_ghostness 峰值（诊断：实证不可作否决，真摔也→0.99）
	verdictGhost bool    // veto_reason==production_verdict_ghost（信号级多径/RCS，唯一安全 ghost 否决源）
}

// vetoGhostX = 否决证据置信阈（委员会 pin：x = 否决证据置信阈，非 P(Fallen)）。
// track_ghostness 须 ≥ 此值才认 ghost 正否决——真摔的瞬时 jump/frozen 二值会 fire，但积分 realLO 压低 ghostness。
const vetoGhostX = 0.90

// txtEventRe 解析 test_record.txt 的 EVENT 行：time | :dev | EVENT <name>。（track 行复用 txtTrackRe，:9e7 专用）
var txtEventRe = regexp.MustCompile(`^(\d\d:\d\d:\d\d)\s*\|\s*:(\w+)\s*\|\s*EVENT\s+(\w+)`)

// feedTxtCase 解析 #1/#2 类 test_record.txt：radar :9e7 track（txtTrackRe）+ radar/sleepad InBed/LeftBed/Enter/Exit
// 事件，路由 :9e7→radar / :978→sleepad，喂生产入口。dayUTC=窗口日期(UTC)。
func feedTxtCase(t *testing.T, dir, dayUTC, radarAddr, sleepadAddr string,
	feedDev func(addr, devType, topic, cat string, dv []map[string]interface{}, ts int64)) int {
	t.Helper()
	f, err := os.Open(filepath.Join(casesDir, dir, "test_record.txt"))
	if err != nil {
		t.Skipf("test_record.txt 缺 %s: %v", dir, err)
	}
	defer f.Close()
	atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }
	tsOf := func(hms string) (int64, bool) {
		tt, err := time.Parse("2006-01-02 15:04:05", dayUTC+" "+hms)
		if err != nil {
			return 0, false
		}
		return tt.UTC().UnixMilli(), true
	}
	frames := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if m := txtTrackRe.FindStringSubmatch(line); m != nil { // :9e7 track
			ts, ok := tsOf(m[1])
			if !ok {
				continue
			}
			feedDev(radarAddr, "radar", "monitor", "track", []map[string]interface{}{{
				"pose": atoi(m[2]), "position_x": atoi(m[3]), "position_y": atoi(m[4]), "position_z": atoi(m[5]),
				"track_id": atoi(m[6]), "area_id": atoi(m[7]), "track_confidence": atoi(m[8]),
			}}, ts)
			frames++
		} else if m := txtEventRe.FindStringSubmatch(line); m != nil { // EVENT <name>
			ts, ok := tsOf(m[1])
			if !ok {
				continue
			}
			addr, dt := radarAddr, "radar"
			if m[2] == "978" {
				addr, dt = sleepadAddr, "sleepad"
			}
			feedDev(addr, dt, "event", m[3], []map[string]interface{}{{"event_status": "start", "heart_rate": -1, "respiratory_rate": -1}}, ts)
			frames++
		}
	}
	return frames
}

// runDBNVeto 喂一案真数据走真 pipeline → 采集 peak P + 正否决证据（R0 只读）。
func runDBNVeto(t *testing.T, vc vetoCase) vetoEvidence {
	t.Helper()
	cfg, radarAddr, roomID := bLayout(t, vc.dir)
	core, logs := observer.New(zapcore.DebugLevel)
	e := NewEngine(nil, zap.New(core))
	e.RegisterRoom(cfg)
	e.deviceRoom[radarAddr] = roomID
	e.deviceMounts[radarAddr] = cfg.Radar

	mk := func(topic, cat string, dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON, _ := json.Marshal(dv)
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": radarAddr, "device_type": "radar", "topic_type": topic, "category": cat,
			"timestamp": strconv.FormatInt(ts, 10), "dataValue": string(dvJSON),
		}}
	}
	feed := func(topic, cat string, dv []map[string]interface{}, ts int64) {
		if topic == "event" {
			e.handleEventMessage(mk(topic, cat, dv, ts))
		} else {
			e.handleMessage(nil, mk(topic, cat, dv, ts))
		}
	}
	// feedDev：txt 用，按 addr/devType 路由（radar :9e7 / sleepad :978）。
	feedDev := func(addr, devType, topic, cat string, dv []map[string]interface{}, ts int64) {
		dvJSON, _ := json.Marshal(dv)
		msg := rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": addr, "device_type": devType, "topic_type": topic, "category": cat,
			"timestamp": strconv.FormatInt(ts, 10), "dataValue": string(dvJSON),
		}}
		if topic == "event" {
			e.handleEventMessage(msg)
		} else {
			e.handleMessage(nil, msg)
		}
	}

	frames := 0
	switch vc.format {
	case "v2": // {category,device_uid,timestamp,data_value} 无 topic_type → 按 category 推
		for _, r := range legoLoadWindow(t, vc.dir) {
			topic := "monitor"
			if legoEventCategory(r.Category) {
				topic = "event"
			}
			feed(topic, r.Category, r.DataValue, r.Timestamp)
			frames++
		}
	case "v1": // bRecord {device_uid,timestamp,topic_type,category,data_value}
		file := bFindWindowFile(t, vc.dir)
		for _, r := range bLoadRecords(t, vc.dir, file) {
			topic := r.TopicType
			if topic == "" {
				topic = "monitor"
				if legoEventCategory(r.Category) {
					topic = "event"
				}
			}
			feed(topic, r.Category, r.DataValue, r.Timestamp)
			frames++
		}
	case "raw": // monitor_stream dump {ts,stream_type,payload}（monitor-only）
		for _, r := range loadRawDump(t, vc.dir) {
			ts, ok := parseDumpTs(r.Ts)
			if !ok {
				continue
			}
			cat := strings.TrimPrefix(r.StreamType, "radar.") // radar.track→track / radar.heart→heart
			feed("monitor", cat, r.Payload, ts)
			frames++
		}
	case "txt": // test_record.txt（#1/#2）：radar track + radar/sleepad 事件，:978 sleepad 路由
		if vc.sleepad != "" {
			e.deviceRoom[vc.sleepad] = roomID
		}
		frames = feedTxtCase(t, vc.dir, vc.dayUTC, radarAddr, vc.sleepad, feedDev)
	}
	ev := vetoEvidence{frames: frames}
	if os.Getenv("VDIAG") != "" { // 只读诊断:dump 本案所有 belief_shadow_* 消息分布 + ghost 相关字段
		mc := map[string]int{}
		for _, le := range logs.All() {
			if strings.HasPrefix(le.Message, "belief_shadow_") {
				mc[le.Message]++
			}
		}
		t.Logf("VDIAG[%s] msgs=%v", vc.dir, mc)
	}
	for _, le := range logs.All() {
		switch le.Message {
		case "belief_shadow_trace":
			if v, ok := le.ContextMap()["p_fallen"].(float64); ok && v > ev.peak {
				ev.peak = v
			}
			if s, _ := le.ContextMap()["argmax_state"].(string); s == "Artifact" {
				ev.ghost = true // Room 层 argmax=Artifact（present 冻结/反射伪迹）= ghost 正否决证据
				ev.argmax = s
			}
		case "belief_shadow_track_lost":
			if s, _ := le.ContextMap()["argmax_tstate"].(string); s == "Ghost" {
				ev.ghost = true // lost track → Ghost verdict = 正否决证据
				ev.argmax = "TGhost"
			}
		case "belief_shadow_nodetect_gated":
			// P6.1a:realnessP<0.5 = ghost 消失(镜面/反射/冻结伪迹被 realness 识别)= 正否决证据。
			// door-exit(p6_1a_door_exit>0.5)是离场≈recovery,归 recovery 不归 ghost。
			if ri, ok := le.ContextMap()["p6_1a_Ri"].(float64); ok && ri < 0.5 {
				ev.ghost = true
				ev.argmax = "realness-ghost"
			}
		case "belief_shadow_veto_evidence":
			// R0 结构化否决证据(belief_shadow.go 生产 emit)。★实证铁律(本 harness 揭出):stillness-based
			// frozenGhost/jumpGhost/track_ghostness **不可作否决证据**——真摔躺地不动 = frozen 伪迹同貌,
			// 检测器在真摔上 fire(gn→0.99)、在真 sit-ghost 上反不 fire(反相关)。唯一安全 ghost 否决 =
			// **信号级 production VerdictGhost**(多径/RCS,非静止)。只认 veto_reason==production_verdict_ghost。
			if gn, ok := le.ContextMap()["track_ghostness"].(float64); ok && gn > ev.maxGhostness {
				ev.maxGhostness = gn
			}
			if r, _ := le.ContextMap()["veto_reason"].(string); r == "production_verdict_ghost" {
				ev.verdictGhost = true
			}
		case "belief_shadow_bed_occupied_suppress":
			ev.bed = true
			if c, ok := le.ContextMap()["p5_bed_conf"].(float64); ok && c > ev.bedConf {
				ev.bedConf = c // 捕最高床占用 conf,判「确凿在床」vs「靠床边」
			}
		case "belief_shadow_exit_recapture", "belief_shadow_lostfall_cancel", "belief_shadow_neighbor_handoff":
			ev.recovery = true // 正向恢复（sole-resident recapture/回床/邻房 hand-off）= 正否决证据
		case "belief_shadow_lostfall_escalate":
			ev.escalate = true // 窗到未佐证=真摔；**否决逻辑不得覆盖**（track 消失≠恢复，可能昏迷重伤）
		}
	}
	return ev
}

func TestVetoPrecisionHarness(t *testing.T) {
	var vetos, correctVetos, wrongVetos int
	var totalFP, fpVetoed, totalReal int // 双轴(be229fd 五):覆盖=fpVetoed/totalFP / 精度=correctVetos/vetos
	for _, vc := range vetoCases {
		ev := runDBNVeto(t, vc)
		// ★委员会实质1:否决须**正证据**；中 P 无证据→**默认放行**。P 仅诊断。
		// ★bed-veto 精度漏洞(委员会 gate-critical):bed 单独**不够**——#2 床边真摔靠床→radar InBed(conf 0.6)
		//   触发 bed_occupied_suppress 会误否真摔。收紧成「**确凿在床**」:bedConf ≥ 0.9(sleepad 接触式/human-bed),
		//   区别于「靠床边」partial(radar-only 0.6)。「靠床」的真摔受害者 conf 低→不否;「确凿睡床」conf 高→否。
		// ★四:track 消失/escalate **不**算否决证据（可能昏迷重伤盲区）→ 不进 wouldVeto。
		const bedConfFloor = 0.9
		bedVeto := ev.bed && ev.bedConf >= bedConfFloor
		// ghost 否决=Room 层积分判定(Artifact/TGhost/realness<.5,无检测真摔不触发) 或 信号级 VerdictGhost；
		// **绝不**用 stillness-based track_ghostness(实证真摔躺地→0.99 误否,gn 仅诊断打印)。
		ghostVeto := ev.ghost || ev.verdictGhost
		wouldVeto := ghostVeto || bedVeto || ev.recovery
		reason := ""
		if ghostVeto {
			src := ev.argmax
			if ev.verdictGhost {
				src = "verdict-signal"
			}
			reason += fmt.Sprintf("ghost(%s,gn=%.2f) ", src, ev.maxGhostness)
		}
		if ev.bed {
			if bedVeto {
				reason += "bed-确凿(conf≥.9) "
			} else {
				reason += "bed-靠床(conf<.9不否) "
			}
		}
		if ev.recovery {
			reason += "recovery "
		}
		if reason == "" {
			reason = "(无正证据→默认放行)"
		}
		if ev.escalate {
			reason += "[escalate:窗到未佐证=真摔,不否]"
		}
		correct := ""
		if vc.truth == "false-alarm" {
			totalFP++
		} else {
			totalReal++
		}
		if wouldVeto {
			vetos++
			if vc.truth == "false-alarm" {
				correctVetos++
				fpVetoed++
				correct = "正确否决✓"
			} else {
				wrongVetos++
				correct = "★错误否决(否掉真摔)✗ 破gate"
			}
		} else if vc.truth == "real-fall" {
			correct = "正确不否决✓(默认放行,真摔保留)"
		} else {
			correct = "未否决(无正证据放行;漏覆盖,非 gate)"
		}
		t.Logf("[%s] %s  帧=%d  peakP=%.3f bedConf=%.2f 正证据=%s  would-veto=%v  %s", vc.truth, vc.dir, ev.frames, ev.peak, ev.bedConf, reason, wouldVeto, correct)

		// ★realness-veto 精度安全闸（委员会 370c594 点名「真摔久躺→断言 realness 不 flag ghost」）：
		// 真摔（尤其 #2 bedtest 627 帧久躺）的 stillness 信号 maxGhostness **可高**（=long-lie 危险真实存在,
		// frozen 检测器同貌触发）——但安全 ghost 源（realness-no-detect / Room argmax / 信号 VerdictGhost）
		// **绝不**得 flag。二者分离 = 安全闸守住：危险可见但不误否。比聚合 wouldVeto 更精准定位失效属性。
		if vc.truth == "real-fall" {
			if ev.ghost || ev.verdictGhost {
				t.Errorf("★long-lie 安全闸破:真摔 %s(%d帧,maxGhostness=%.2f)被**安全** ghost 源误判(argmax=%s,verdictGhost=%v)——realness-veto 误杀久躺真受害者", vc.dir, ev.frames, ev.maxGhostness, ev.argmax, ev.verdictGhost)
			} else if ev.maxGhostness >= vetoGhostX {
				t.Logf("  ↳ long-lie 安全闸✓:%s stillness maxGhostness=%.2f≥x(危险信号真实存在)但安全源未 flag→不误否", vc.dir, ev.maxGhostness)
			}
		}
	}
	cov, prec := 0.0, 1.0
	if totalFP > 0 {
		cov = float64(fpVetoed) / float64(totalFP)
	}
	if vetos > 0 {
		prec = float64(correctVetos) / float64(vetos)
	}
	t.Logf("=== 双轴(be229fd 五) ===")
	t.Logf("覆盖 coverage = 误报否掉 %d/%d = %.0f%%（期望≈90%%，须真否决案够样本）", fpVetoed, totalFP, cov*100)
	t.Logf("精度 precision = 正确否决 %d/全部否决 %d = %.0f%%（gate≥95%%；真摔错否 %d 须=0）", correctVetos, vetos, prec*100, wrongVetos)
	t.Logf("⚠ 样本 FP=%d real=%d 远<500 → 闭环结构非定量，不声称数字（be229fd 五）", totalFP, totalReal)
	// ★延时窗 5min + T_fire 锚点待补：firmware 开火时刻在 alarm_events（非 monitor 窗），需 export 含 alarm
	//   才能精确 [T_fire, T_fire+5min] 窗 + 早退；当前为整案证据（近似），窗长参数化/早退下增量。
	if wrongVetos > 0 {
		t.Fatalf("★精度破 gate：%d 例错误否决真摔——must=0", wrongVetos)
	}
}
