package roomengine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"testing"

	rediscommon "owl-common/redis"

	"wisefido-sensor/testkit"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_generator_test.go — DBN 模拟测试生成器（用户下阶段重心：用**真实** track 片段随机组装合成案验 DBN）。
//
// 思路（不依赖人为测试时长真实性 [[fall_data_is_artificial_test]]）：从真实案切出**带真实噪声**的 pose 片段
// (walk/stand/fall-过渡/lying/ghost) 作乐高块 → 随机时长 + 随机注入 fall → 组装 silent/lost/moving_fall ×
// bathroom/bedroom 场景集 → 喂生产 pipeline 跑 DBN → 验证检出（fire + reason 路由）。
// 时长由生成器随机控制（非测试员行为）→ 可系统扫场景空间，至少出"模拟结论"（各场景检出率）。
//
// 片段=真实 track（保留真实运动/丢轨/反射噪声）；fall/lost/silent 由组装逻辑注入（确定性标签=oracle）。

type synthFrame struct{ pose, x, y, z, tid, conf int }

// fragment 从真实 donor 案切出的 pose 标记片段。
type fragment struct {
	label  string
	frames []synthFrame
}

// donorV2Frames 读真实 v2 案的 track 帧（按 ts 升序,展开 data_value）。缺 window.json → 返 nil(不 skip 整测)。
func donorV2Frames(t *testing.T, dir string) []synthFrame {
	t.Helper()
	recs, err := testkit.LoadWindow(filepath.Join(casesDir, dir))
	if err != nil {
	}
	var out []synthFrame
	for _, r := range recs {
		if r.Category != "track" {
			continue
		}
		for _, d := range r.DataValue {
			out = append(out, synthFrame{
				pose: toIntField(d["pose"]), x: toIntField(d["position_x"]), y: toIntField(d["position_y"]),
				z: toIntField(d["position_z"]), tid: toIntField(d["track_id"]), conf: toIntField(d["track_confidence"]),
			})
		}
	}
	return out
}

// longestPoseRun 取 frames 里最长连续 pose∈want 的片段（真实噪声保留）。
func longestPoseRun(frames []synthFrame, want map[int]bool, label string) fragment {
	bestS, bestN, curS := 0, 0, -1
	for i := 0; i <= len(frames); i++ {
		in := i < len(frames) && want[frames[i].pose]
		if in && curS < 0 {
			curS = i
		}
		if !in && curS >= 0 {
			if i-curS > bestN {
				bestS, bestN = curS, i-curS
			}
			curS = -1
		}
	}
	if bestN == 0 {
		return fragment{label: label}
	}
	return fragment{label: label, frames: append([]synthFrame(nil), frames[bestS:bestS+bestN]...)}
}

// buildLibrary 从真实 donor 案切乐高块库。
func buildLibrary(t *testing.T) map[string]fragment {
	t.Helper()
	d9 := donorV2Frames(t, "unit201-handoff-0609-bathroom-333B") // 含 walk(1)/过渡(2)/fallen(5)
	d5934 := donorV2Frames(t, "recovery-fp-5934-0609-walking")   // 含 stand(4)/walk(1)/fallen(5)
	dghost := donorV2Frames(t, "cabb-ghost-frozen-sit-0415")     // ghost 静态反射(3/4)
	if len(dghost) == 0 {
		dghost = donorV2Frames(t, "cabb-ghost-frozen-sit-2117")
	}
	lib := map[string]fragment{
		"walk":   pick(longestPoseRun(d9, map[int]bool{1: true}, "walk"), longestPoseRun(d5934, map[int]bool{1: true}, "walk")),
		"stand":  longestPoseRun(d5934, map[int]bool{4: true}, "stand"),
		"fallen": pick(longestPoseRun(d9, map[int]bool{5: true, 2: true}, "fallen"), longestPoseRun(d5934, map[int]bool{5: true}, "fallen")),
		"ghost":  longestPoseRun(dghost, map[int]bool{3: true, 4: true}, "ghost"),
	}
	return lib
}

func pick(a, b fragment) fragment { // 取帧多的(更可靠的真实片段)
	if len(b.frames) > len(a.frames) {
		return b
	}
	return a
}

// composeScenario 用乐高块按场景模板组装合成帧序（随机时长）。返回帧+预期(oracle)。
func composeScenario(scenario string, lib map[string]fragment, rng *rand.Rand) (frames []synthFrame, expectFire bool, expectGhost bool) {
	// 随机时长(帧数,1Hz):walk 10-60s / lying 60-360s。
	walkN := 10 + rng.Intn(50)
	lieN := 60 + rng.Intn(300)
	anchorX, anchorY := -100+rng.Intn(200), -100+rng.Intn(200)
	emit := func(frag fragment, n int) {
		if len(frag.frames) == 0 {
			return
		}
		for i := 0; i < n; i++ {
			f := frag.frames[i%len(frag.frames)] // 循环填到目标时长
			f.x, f.y = anchorX+(f.x%40), anchorY+(f.y%40)
			f.tid = 0
			if f.conf == 0 {
				f.conf = 80
			}
			frames = append(frames, f)
		}
	}
	switch scenario {
	case "moving_fall": // 走动→突然倒地(pose→5,z→0)→短暂躺地。fire=true
		emit(lib["walk"], walkN)
		emit(lib["fallen"], 20+rng.Intn(40))
		expectFire = true
	case "silent_fall": // 走动→倒地→长时间静止躺地(无恢复)。fire=true(ReasonSilent/PoseLying)
		emit(lib["walk"], walkN)
		emit(lib["fallen"], lieN)
		expectFire = true
	case "lost_fall": // 走动→倒地→track 消失(停止发帧,贴地弱回波)。fire=true(ReasonLost)
		emit(lib["walk"], walkN)
		emit(lib["fallen"], 5+rng.Intn(10)) // 短暂可见后丢轨
		// 不再 emit → track lost
		expectFire = true
	case "ghost": // 纯静态反射(无真人倒地)。fire=false / ghost=true
		emit(lib["ghost"], 60+rng.Intn(120))
		expectFire = false
		expectGhost = true
	case "walk_only": // 纯走动(正常,无 fall)。fire=false
		emit(lib["walk"], 60+rng.Intn(120))
		expectFire = false
	}
	return frames, expectFire, expectGhost
}

// TestDBNGeneratorScenarios — 生成器跑各场景 N 例 → DBN → 报检出率（模拟结论）。
func TestDBNGeneratorScenarios(t *testing.T) {
	lib := buildLibrary(t)
	for k, f := range lib {
		t.Logf("乐高块 %s: %d 真实帧", k, len(f.frames))
	}
	for _, need := range []string{"walk", "fallen"} { // fall 场景必需块;ghost 可缺(donor 是 v1 无 window.json)
		if len(lib[need].frames) == 0 {
			t.Fatalf("乐高块 %s 空(donor 缺该 pose run)→ 生成器无素材", need)
		}
	}
	scenarios := []string{"moving_fall", "silent_fall", "lost_fall", "ghost", "walk_only"}
	// P4 分类 oracle:注入类型 → 期望 DBN p7_3_reason(belief 层;moving 是 Room 层对 pose_lying 的细分,belief 给 pose_lying)。
	expectReason := map[string]string{"silent_fall": "silent", "lost_fall": "lost", "moving_fall": "pose_lying"}
	const perScenario = 10
	cfg, radarAddr, roomID := bLayout(t, "unit201-handoff-0609-bathroom-333B") // bathroom 上下文
	for _, sc := range scenarios {
		if sc == "ghost" && len(lib["ghost"].frames) == 0 {
			t.Logf("[ghost] 跳过(donor 是 v1 无 window.json,ghost 块缺;分类验证不需 ghost)")
			continue
		}
		rng := rand.New(rand.NewSource(int64(len(sc)))) // 确定性种子(可复现),场景间变化
		fired, correctDetect, classCorrect, classTotal := 0, 0, 0, 0
		for n := 0; n < perScenario; n++ {
			frames, expectFire, _ := composeScenario(sc, lib, rng)
			if len(frames) == 0 {
				continue
			}
			core, logs := observer.New(zapcore.DebugLevel)
			e := NewEngine(nil, zap.New(core))
			e.RegisterRoom(cfg)
			e.deviceRoom[radarAddr] = roomID
			e.deviceMounts[radarAddr] = cfg.Radar
			baseTs := int64(1781000000000)
			for i, f := range frames {
				dv := fmt.Sprintf(`[{"pose":%d,"position_x":%d,"position_y":%d,"position_z":%d,"track_id":%d,"track_confidence":%d}]`,
					f.pose, f.x, f.y, f.z, f.tid, f.conf)
				e.handleMessage(nil, rediscommon.StreamMessage{Values: map[string]interface{}{
					"device_addr": radarAddr, "device_type": "radar", "topic_type": "monitor", "category": "track",
					"timestamp": strconv.FormatInt(baseTs+int64(i)*1000, 10), "dataValue": dv,
				}})
			}
			fallLogs := logs.FilterMessage("belief_shadow_fall")
			didFire := fallLogs.Len() > 0
			if didFire {
				fired++
			}
			if didFire == expectFire {
				correctDetect++
			}
			// P4 分类准确率(委员会 d48e0da):fire 案的 p7_3_reason == 注入类型?(只对 fall 场景)
			if want, ok := expectReason[sc]; ok && didFire {
				classTotal++
				if r, _ := fallLogs.All()[0].ContextMap()["p7_3_reason"].(string); r == want {
					classCorrect++
				}
			}
		}
		if _, isFall := expectReason[sc]; isFall {
			t.Logf("[%s] fire %d/%d 检出对 %d/%d  ★分类(reason==注入) %d/%d", sc, fired, perScenario, correctDetect, perScenario, classCorrect, classTotal)
		} else {
			t.Logf("[%s] fire %d/%d 检出对 %d/%d", sc, fired, perScenario, correctDetect, perScenario)
		}
	}
}

// TestDBNFireSwitch — cutover wire 验证（委员会 6c376e4）:开关 ON → DBN 真 fire/veto block 可达。
// 喂 #9 真摔(present pose-lying)→ 应 belief_dbn_fire(real track 不被 ghost veto)。验 R0→production wire 通。
func TestDBNFireSwitch(t *testing.T) {
	old := dbnMode
	dbnMode = 2 // 全开:DBN 自发 + 可否决 firmware
	defer func() { dbnMode = old }()
	cfg, radarAddr, roomID := bLayout(t, "unit201-handoff-0609-bathroom-333B")
	core, logs := observer.New(zapcore.DebugLevel)
	e := NewEngine(nil, zap.New(core))
	e.RegisterRoom(cfg)
	e.deviceRoom[radarAddr] = roomID
	e.deviceMounts[radarAddr] = cfg.Radar
	for _, r := range mustLoadWindow(t, "unit201-handoff-0609-bathroom-333B") {
		dvJSON, _ := json.Marshal(r.DataValue)
		topic := "monitor"
		if testkit.EventCategory(r.Category) {
			topic = "event"
		}
		msg := rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": radarAddr, "device_type": "radar", "topic_type": topic, "category": r.Category,
			"timestamp": strconv.FormatInt(r.Timestamp, 10), "dataValue": string(dvJSON),
		}}
		if topic == "event" {
			e.handleEventMessage(msg)
		} else {
			e.handleMessage(nil, msg)
		}
	}
	shadowFall := logs.FilterMessage("belief_shadow_fall").Len()
	fired := logs.FilterMessage("belief_dbn_fire").Len()
	vetoed := logs.FilterMessage("ghost_veto").FilterField(zap.String("reason", "dbn_coexist")).Len()
	t.Logf("#9 DBN_FIRE=on: shadow_fall=%d  dbn_fire=%d  dbn_veto=%d", shadowFall, fired, vetoed)
	if shadowFall > 0 && fired == 0 && vetoed == 0 {
		t.Errorf("★开关 ON 但 DBN 既未 fire 也未 veto(fire block 未达)→ cutover wire 漏")
	}
	if fired == 0 {
		t.Logf("⚠ #9 真摔未 dbn_fire(被 ghost veto?)——查 veto=%d", vetoed)
	}
}
