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
	{"cd2b-fall-0607-twindow", "v2", "false-alarm", "#7 firmware 在床误报→该否决(T_fire 居中重导[−2m,+15m]支撑切窗)", "", ""},
	{"hunzi-cabb-lost-0601-twindow", "v2", "false-alarm", "#5 lost_track 误报(T_fire 重导;lost-fall 安全螺丝**不可**否决=诚实分母成员,不否=安全正确)", "", ""},
	// ★recovery-FP（委员会 value 段关键样本）：firmware 误火 Fall → 人 W 内起身/移动自证活着 → DBN 经 recovery 证据否。
	// 这是「5min 砍 90%」value 命题唯一能测的类（ghost=W 无关/lost=不可否）。Walking 到达 +8.7min → W=5 漏/W=10/15 命中。
	{"recovery-fp-5934-0609-walking", "v2", "false-alarm", "recovery-FP:误火→+8.7min Walking 自证→该 W 内否(value W-依赖段)", "", ""},
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

	// 切窗双轴（委员会 85e41a1 钉档）：各**安全**否决证据的**最早到达 ts**（ms,0=未现）。窗内覆盖按
	// 「到达 ts ≤ T_fire+W」算。摔前 ghost→ts<T_fire=W 无关;摔后 recovery→ts>T_fire=W 相关。
	ghostTsMs    int64 // 最早安全 ghost 证据(realness<.5/Artifact/TGhost/VerdictGhost)到达
	bedVetoTsMs  int64 // 最早 bedConf≥0.9 确凿在床到达
	recoveryTsMs int64 // 最早正向恢复(recapture/handoff)到达
}

func leTsMs(le observer.LoggedEntry) int64 {
	if v, ok := le.ContextMap()["ts_ms"].(int64); ok {
		return v
	}
	return 0
}

func minNonZero(a, b int64) int64 { // 取最早到达（0=未现，不参与 min）
	if a == 0 {
		return b
	}
	if b == 0 || a < b {
		return a
	}
	return b
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
				ev.ghostTsMs = minNonZero(ev.ghostTsMs, leTsMs(le))
			}
		case "belief_shadow_track_lost":
			if s, _ := le.ContextMap()["argmax_tstate"].(string); s == "Ghost" {
				ev.ghost = true // lost track → Ghost verdict = 正否决证据
				ev.argmax = "TGhost"
				ev.ghostTsMs = minNonZero(ev.ghostTsMs, leTsMs(le))
			}
		case "belief_shadow_nodetect_gated":
			// P6.1a:realnessP<0.5 = ghost 消失(镜面/反射/冻结伪迹被 realness 识别)= 正否决证据。
			// door-exit(p6_1a_door_exit>0.5)是离场≈recovery,归 recovery 不归 ghost。
			if ri, ok := le.ContextMap()["p6_1a_Ri"].(float64); ok && ri < 0.5 {
				ev.ghost = true
				ev.argmax = "realness-ghost"
				ev.ghostTsMs = minNonZero(ev.ghostTsMs, leTsMs(le))
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
				ev.ghostTsMs = minNonZero(ev.ghostTsMs, leTsMs(le))
			}
		case "belief_shadow_bed_occupied_suppress":
			ev.bed = true
			if c, ok := le.ContextMap()["p5_bed_conf"].(float64); ok && c > ev.bedConf {
				ev.bedConf = c // 捕最高床占用 conf,判「确凿在床」vs「靠床边」
			}
			if c, ok := le.ContextMap()["p5_bed_conf"].(float64); ok && c >= 0.9 {
				ev.bedVetoTsMs = minNonZero(ev.bedVetoTsMs, leTsMs(le)) // 确凿在床(conf≥.9)的最早到达
			}
		case "belief_shadow_exit_recapture", "belief_shadow_lostfall_cancel", "belief_shadow_neighbor_handoff":
			ev.recovery = true // 正向恢复（sole-resident recapture/回床/邻房 hand-off）= 正否决证据
			ev.recoveryTsMs = minNonZero(ev.recoveryTsMs, leTsMs(le))
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
		// ★对齐校验前置(no silent caps):false-alarm 若 firmware **未火**(t_fire_ms=null)→ 根本无告警可否决,
		// **不是覆盖候选**,排除出分母(否则虚抬覆盖)。cabb-frozen ×2(CABB 05-04 早于 alarm 管道 05-21)即此类。
		if a, ok := loadTFire(t, vc.dir); ok && vc.truth == "false-alarm" && a.TFireMs == nil {
			t.Logf("[排除] %s firmware 未火→非否决候选,移出覆盖分母(%s)", vc.dir, a.Note)
			continue
		}
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

func toIntField(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	}
	return 0
}

// analyzeRecoveryV2 — recovery-veto value 段 R0 验证（harness-side 选 B,不碰 production 安全语义；待委员会裁
// A/B 后再决定 production wiring）。从 v2 window 帧 pose 时间线直接算判别力——验「摔后倒地时长」能否分开
// **自救型真摔(#9 倒73s,不可否)** vs **误报(5934 +2s 即站,可否)**：
//
//	① 同人 = track 在 T_fire 前已现(护工后进=新 track firstSeen>T_fire→排除,条件②)。
//	② 摔后(T_fire 后)持续倒地 ≥genuineFallenMs(15s) = 自救型真摔 → genuineFall=true → 不算恢复(不否)。
//	③ 否则首个持续直立(≥3s,条件③)= 恢复,到达 ts 入 W-曲线。
func analyzeRecoveryV2(t *testing.T, vc vetoCase, tFireMs int64) (recoveryTsMs, postFireFallenMs int64, genuineFall bool) {
	if vc.format != "v2" {
		return 0, 0, false // recovery 候选都是 v2;bedtest txt 真摔不需
	}
	const sustainUprightMs, genuineFallenMs = 3000, 15000
	const lostGapMs = 60000                                                                // track 超此无帧=丢失(对齐 beliefShadowLostTTLMs)
	isUpright := func(p int) bool { return p == 1 || p == 4 }                              // Walking/Standing
	isFallen := func(p int) bool { return p == 2 || p == 5 || p == 6 || p == 7 || p == 8 } // 倒地/卧/坐地
	type tk struct {
		uprightSince, fallenSince, firstSeen, lastTs int64
		lostGap                                      bool // 摔后曾丢轨>TTL(委员会螺丝:track-lost≠up,盲区受害者倒地显短是假象)
	}
	tracks := map[int]*tk{}
	for _, r := range legoLoadWindow(t, vc.dir) {
		if r.Category != "track" {
			continue
		}
		for _, d := range r.DataValue {
			pose, tid, ts := toIntField(d["pose"]), toIntField(d["track_id"]), r.Timestamp
			s := tracks[tid]
			if s == nil {
				s = &tk{firstSeen: ts, lastTs: ts}
				tracks[tid] = s
			}
			// ★track-lost 安全螺丝:摔后帧间隔 >TTL = 曾丢轨 → 倒地段"显短"是丢轨假象非真站起 → 永久禁本 track 恢复
			//   (盲区躺地受害者雷达看不见,绝不当"短倒地=误火"否掉=long-lie 同类灾难)。
			if s.lastTs > tFireMs && ts-s.lastTs > lostGapMs {
				s.lostGap = true
			}
			s.lastTs = ts
			switch {
			case isUpright(pose):
				if s.uprightSince == 0 {
					s.uprightSince = ts
				}
				s.fallenSince = 0
			case isFallen(pose):
				if s.fallenSince == 0 {
					s.fallenSince = ts
				}
				s.uprightSince = 0
				postStart := s.fallenSince
				if tFireMs > postStart {
					postStart = tFireMs
				}
				if ts > tFireMs && ts-postStart > postFireFallenMs {
					postFireFallenMs = ts - postStart
				}
			default:
				s.uprightSince, s.fallenSince = 0, 0
			}
			// 恢复=同人(firstSeen≤T_fire)+摔后(ts>T_fire)+持续**正向直立**≥sustain+**无丢轨 gap**(正向 up 证据非 track-presence)。
			if s.firstSeen <= tFireMs && !s.lostGap && ts > tFireMs && s.uprightSince > tFireMs &&
				ts-s.uprightSince >= sustainUprightMs && recoveryTsMs == 0 {
				recoveryTsMs = ts
			}
		}
	}
	genuineFall = postFireFallenMs >= genuineFallenMs
	if genuineFall {
		recoveryTsMs = 0 // 自救型真摔:不算恢复(留告警,不静默抹,呼应 silent_leftbed)
	}
	return recoveryTsMs, postFireFallenMs, genuineFall
}

// TestVetoWindowScan — 切窗双轴（委员会 85e41a1 钉档）：窗=[T_fire−warmup, T_fire+W],喂全窗,覆盖按
// 否决证据**到达 ts ≤ T_fire+W** 算。摔前 ghost(ts<T_fire)=W 无关(任何 W 否得掉);摔后 recovery(ts>T_fire)
// =W 相关(W 内起身才否)。⟹ 覆盖-W 曲线 = 即时 ghost(平) + recovery(随 W 涨)。精度=0(真摔任何 W 不否)。
// 仅纳入窗含 T_fire 的案(对齐校验合格);非火案天然排除(无 T_fire)。
func TestVetoWindowScan(t *testing.T) {
	scan := []int{5, 10, 15}
	type wstat struct{ fpVetoed, totalFP, correctVetos, vetos, wrongVetos int }
	stats := map[int]*wstat{5: {}, 10: {}, 15: {}}
	rel := func(ts, tf int64) string {
		if ts == 0 {
			return "—"
		}
		return fmt.Sprintf("%+.1fmin", float64(ts-tf)/60000)
	}
	for _, vc := range vetoCases {
		a, ok := loadTFire(t, vc.dir)
		if !ok || a.TFireMs == nil {
			continue // 非火/无锚 → 不是切窗候选
		}
		mn, mx, n := caseTimeSpanMs(t, vc)
		if n == 0 || *a.TFireMs < mn || *a.TFireMs > mx {
			t.Logf("[跳过] %s 窗不含 T_fire（需 [T_fire−warmup,+W] 重导）", vc.dir)
			continue
		}
		ev := runDBNVeto(t, vc)
		tf := *a.TFireMs
		// recovery 信号:R0 harness-side(选 B)算自 pose 时间线 + 摔后倒地闸(self-rescue 判别)。
		// production wiring(Fall 事件进 shadow)待委员会裁 A/B;现先验判别力在真数据上分不分得开。
		recTs, postFallenMs, genuineFall := analyzeRecoveryV2(t, vc, tf)
		for _, w := range scan {
			cutoff := tf + int64(w)*60000
			ghostV := ev.ghost && ev.ghostTsMs != 0 && ev.ghostTsMs <= cutoff
			bedV := ev.bedVetoTsMs != 0 && ev.bedVetoTsMs <= cutoff
			recV := recTs != 0 && recTs <= cutoff
			wouldVeto := ghostV || bedV || recV
			s := stats[w]
			if vc.truth == "false-alarm" {
				s.totalFP++
				if wouldVeto {
					s.fpVetoed++
					s.vetos++
					s.correctVetos++
				}
			} else if wouldVeto { // real-fall 被否 = 破精度 gate
				s.vetos++
				s.wrongVetos++
			}
		}
		t.Logf("[%s] %s  ghost=%s  bed≥.9=%s  recovery=%s  摔后倒地=%.0fs%s", vc.truth, vc.dir, rel(ev.ghostTsMs, tf), rel(ev.bedVetoTsMs, tf), rel(recTs, tf), float64(postFallenMs)/1000, map[bool]string{true: " ★自救真摔→不否", false: ""}[genuineFall])
	}
	t.Logf("=== 覆盖-W 退化曲线（摔前 ghost=W 无关平 / 摔后 recovery=随 W 涨；委员会钉档语义）===")
	for _, w := range scan {
		s := stats[w]
		cov, prec := 0.0, 1.0
		if s.totalFP > 0 {
			cov = float64(s.fpVetoed) / float64(s.totalFP)
		}
		if s.vetos > 0 {
			prec = float64(s.correctVetos) / float64(s.vetos)
		}
		t.Logf("W=%2dmin: 覆盖=%d/%d=%.0f%%  精度=%d/%d=%.0f%%  真摔错否=%d", w, s.fpVetoed, s.totalFP, cov*100, s.correctVetos, s.vetos, prec*100, s.wrongVetos)
		if s.wrongVetos > 0 {
			t.Errorf("★W=%dmin 精度破 gate：错否真摔 %d 例——must=0（切窗只减证据,真摔被否=逻辑错）", w, s.wrongVetos)
		}
	}
	t.Logf("⚠ 样本极小（窗-valid 案 <5）→ 曲线是结构验证非定量；待 bedtest 重导 + 良性 FP 扩样")
}

// tFireArtifact = 每案 committed t_fire.json（firmware 开火时刻，离线从 alarm_events 导，**测时不连库**，
// 守铁律 sensor 不连库 + #1 教训 fixture 须自洽 committed）。t_fire_ms=null 表示 firmware 未火（非否决候选）。
type tFireArtifact struct {
	TFireMs  *int64 `json:"t_fire_ms"`
	TFireUTC string `json:"t_fire_utc"`
	Note     string `json:"note"`
}

func loadTFire(t *testing.T, dir string) (tFireArtifact, bool) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(casesDir, dir, "t_fire.json"))
	if err != nil {
		return tFireArtifact{}, false
	}
	var a tFireArtifact
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatalf("解析 t_fire.json %s: %v", dir, err)
	}
	return a, true
}

// caseTimeSpanMs 复用各格式 loader 提取本案数据时段 [min,max]（epoch ms）。窗须含 T_fire 且窗后 ≥ maxW 才合格。
func caseTimeSpanMs(t *testing.T, vc vetoCase) (int64, int64, int) {
	t.Helper()
	var ts []int64
	switch vc.format {
	case "v2":
		for _, r := range legoLoadWindow(t, vc.dir) {
			ts = append(ts, r.Timestamp)
		}
	case "v1":
		for _, r := range bLoadRecords(t, vc.dir, bFindWindowFile(t, vc.dir)) {
			ts = append(ts, r.Timestamp)
		}
	case "raw":
		for _, r := range loadRawDump(t, vc.dir) {
			if v, ok := parseDumpTs(r.Ts); ok {
				ts = append(ts, v)
			}
		}
	case "txt":
		f, err := os.Open(filepath.Join(casesDir, vc.dir, "test_record.txt"))
		if err != nil {
			return 0, 0, 0
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			m := txtTrackRe.FindStringSubmatch(sc.Text())
			if m == nil {
				continue
			}
			if tt, err := time.Parse("2006-01-02 15:04:05", vc.dayUTC+" "+m[1]); err == nil {
				ts = append(ts, tt.UTC().UnixMilli())
			}
		}
	}
	if len(ts) == 0 {
		return 0, 0, 0
	}
	mn, mx := ts[0], ts[0]
	for _, v := range ts {
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	return mn, mx, len(ts)
}

// TestCaseTFireAlignment — 延时窗前置（委员会授权①）：校验每案窗能否支撑 [T_fire, T_fire+W] 切窗测覆盖。
// 误对齐 = garbage 测量（cd2b 目录名 0607-0127 vs alarm_events Fall 07:29:29 类陷阱）。诊断型：报真实对齐
// 状态，钉死「哪些案现成可切窗 / 哪些需重导 / 哪些根本非否决候选」。强断言唯一已知合格案 #9 不退化。
func TestCaseTFireAlignment(t *testing.T) {
	const baseWMin = 5 // 用户拍基准延时窗;扫 5/10/15。窗后须 ≥ 该 W 才支撑此 W 切窗。
	// warmup 充分性闸（委员会 3 轮要,真做断言不再 doc-only）:误报覆盖案窗左界须含足够摔前历史让 DBN
	// 信念(realLO)预热,否则 ghost 浮现段被砍 → 系统低估覆盖(−2min cd2b 那次=lead 2min→ghost 不否)。
	// 目标 ~10min(realLO 整定),warmupMin=9 容 export 边界抖动,牢卡 −2min。真摔案豁免(精度不需 warmup,
	// 切窗只减证据更安全)。**结构闸**(lead≥warmupMin);belief 级「ghostness 在 [T_fire−ε,T_fire] 已平」是更强
	// 精化(待 recovery-veto 裁后一并,因需跑 DBN 取 ghostness 轨迹)。
	const warmupMin = 9.0
	var warmupViol []string
	scan := []int{5, 10, 15}
	suitBaseW := func(vc vetoCase) (float64, bool) { // 返回窗后 min + 是否含 T_fire
		a, ok := loadTFire(t, vc.dir)
		if !ok || a.TFireMs == nil {
			return 0, false
		}
		mn, mx, n := caseTimeSpanMs(t, vc)
		if n == 0 || *a.TFireMs < mn || *a.TFireMs > mx {
			return 0, false
		}
		return float64(mx-*a.TFireMs) / 60000, true
	}
	good := 0
	for _, vc := range vetoCases {
		a, ok := loadTFire(t, vc.dir)
		if !ok {
			t.Logf("[%s] (无 t_fire.json,跳过对齐校验)", vc.dir)
			continue
		}
		if a.TFireMs == nil {
			t.Logf("[%s] %s = **非否决候选**(firmware 未火)→ 移出覆盖分母。%s", vc.truth, vc.dir, a.Note)
			continue
		}
		mn, mx, n := caseTimeSpanMs(t, vc)
		if n == 0 {
			t.Logf("[%s] %s 数据 0 帧 → 无法对齐", vc.truth, vc.dir)
			continue
		}
		tf := *a.TFireMs
		postMin := float64(mx-tf) / 60000
		var verdict string
		switch {
		case tf < mn || tf > mx:
			verdict = fmt.Sprintf("✗窗不含T_fire(窗尾后%.1fmin)→需重导[T_fire-2min,T_fire+15min]", float64(tf-mx)/60000)
		case postMin < float64(baseWMin):
			verdict = fmt.Sprintf("△窗含T_fire但窗后仅%.1fmin<基准%dmin→需重导补尾", postMin, baseWMin)
		default:
			var ws []string
			for _, w := range scan {
				if postMin >= float64(w) {
					ws = append(ws, fmt.Sprintf("%d", w))
				}
			}
			verdict = fmt.Sprintf("✓合格(窗后%.1fmin,支持 W=%s min)", postMin, strings.Join(ws, "/"))
			good++
			// warmup 闸:仅误报覆盖案需(真摔走精度不需预热)。lead = T_fire − 窗左界。
			if vc.truth == "false-alarm" {
				lead := float64(tf-mn) / 60000
				if lead < warmupMin {
					warmupViol = append(warmupViol, fmt.Sprintf("%s(lead=%.1fmin<%.0f)", vc.dir, lead, warmupMin))
				} else {
					verdict += fmt.Sprintf("  warmup✓(lead=%.1fmin)", lead)
				}
			}
		}
		t.Logf("[%s] %s  T_fire=%s  窗后=%.1fmin  %s", vc.truth, vc.dir, a.TFireUTC, postMin, verdict)
	}
	t.Logf("=== 对齐校验：%d 案现成可切基准 %dmin 窗(其余需重导或非候选) ===", good, baseWMin)
	if len(warmupViol) > 0 {
		t.Errorf("★warmup 充分性闸破:误报覆盖案窗左界不足(ghost 信念预热不够→低估覆盖,如 −2min cd2b 陷阱):%s——须 [T_fire−≥%.0fmin,…] 重导", strings.Join(warmupViol, " "), warmupMin)
	}
	// 强断言：已知合格锚 #9 必须支撑基准 W=5min——对齐器自身的 oracle（否则 caseTimeSpanMs/t_fire 坏）。
	if post, ok := suitBaseW(vetoCases[0]); !ok || post < float64(baseWMin) {
		t.Fatalf("★对齐器自校验破:#9 333B 应支撑基准 %dmin 窗(窗含 T_fire+窗后≥%dmin)却判不合格(ok=%v post=%.1f) → caseTimeSpanMs/t_fire 坏", baseWMin, baseWMin, ok, post)
	}
}
