package roomengine

import (
	"encoding/json"
	"os"
	"path/filepath"
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
// R0 只读算 would-veto：firmware-fire 真案上跑 DBN belief shadow → peak P(Fallen)；
//   would-veto = P < vetoThreshold（DBN 判「非真摔」→ 否决 firmware fire）。
//   正确否决 = would-veto ∧ ground-truth=false-alarm；错误否决 = would-veto ∧ ground-truth=real-fall。
//   否决精度 = 正确否决 / 全部否决 ≥ 95%（需 500+ 样本，本骨架先走通闭环）。
// 不碰生产开火（R0）；#9 P=0.998「正确不否决」是 gate 要的正数据点（委员会收回「易方向无意义」）。

const vetoThreshold = 0.5 // DBN P(Fallen) < 此 → would-veto（firmware fire 被 DBN 判为误报）

type vetoCase struct {
	dir    string
	format string // v1=带topic_type window / v2=不带(category推断)
	truth  string // real-fall(firmware对,DBN不该否决) / false-alarm(firmware错,DBN该否决)
	note   string
}

// 首批：能直接 load 的两案（其余格式 d523 raw / d5f7 txt 下周期补统一 loader）。
var vetoCases = []vetoCase{
	{"unit201-handoff-0609-bathroom-333B", "v2", "real-fall", "#9 firmware 真摔(pose=5)→该不否决"},
	{"cd2b-fall-0607-0127", "v1", "false-alarm", "#7 firmware 在床误报→该否决"},
	{"hunzi-cabb-lost-0601-2247-FP", "v2", "false-alarm", "#5 lost_track 误报→该否决"},
	// #13 d523-ghost：raw loader 已就绪，但 d523 的 room_visual_layout.canvas **无 Radar object**
	// （只 Wall/Enter/Furniture）→ ParseLayoutConfig 建不出 mount → pipeline 跑不了。待补带雷达 mount 的
	// d523 layout（不得捏造）后启用。raw 格式 loader 本身已验（hunzi v2 + 此 raw 路径）。
}

// runDBNPeakFallen 喂一案真数据走真 pipeline → 返回 belief shadow 的 peak P(Fallen)（R0 只读）。
func runDBNPeakFallen(t *testing.T, vc vetoCase) (peak float64, frames int, skipped bool) {
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
	}
	for _, le := range logs.All() {
		if le.Message == "belief_shadow_trace" {
			if v, ok := le.ContextMap()["p_fallen"].(float64); ok && v > peak {
				peak = v
			}
		}
	}
	return peak, frames, false
}

func TestVetoPrecisionHarness(t *testing.T) {
	var vetos, correctVetos, wrongVetos int
	for _, vc := range vetoCases {
		peak, frames, _ := runDBNPeakFallen(t, vc)
		wouldVeto := peak < vetoThreshold
		correct := ""
		if wouldVeto {
			vetos++
			if vc.truth == "false-alarm" {
				correctVetos++
				correct = "正确否决✓"
			} else {
				wrongVetos++
				correct = "★错误否决(否掉真摔)✗"
			}
		} else {
			if vc.truth == "real-fall" {
				correct = "正确不否决✓(同意firmware真摔)"
			} else {
				correct = "未否决(漏否决false-alarm,非gate但记)"
			}
		}
		t.Logf("[%s] %s  帧=%d peak P=%.3f  would-veto=%v  %s", vc.truth, vc.dir, frames, peak, wouldVeto, correct)
	}
	prec := 1.0
	if vetos > 0 {
		prec = float64(correctVetos) / float64(vetos)
	}
	t.Logf("=== 否决精度 = 正确否决 %d / 全部否决 %d = %.1f%%（gate≥95%%；样本 %d 远<500,骨架闭环非定量）",
		correctVetos, vetos, prec*100, len(vetoCases))
	if wrongVetos > 0 {
		t.Logf("⚠ 有 %d 例错误否决(否掉真摔)——上线前须查", wrongVetos)
	}
}
