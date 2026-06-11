package roomengine

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"testing"

	rediscommon "owl-common/redis"
	"wisefido-sensor/testkit"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_recall_realdata_test.go — Tier-1 recall 闭环（真数据走真 pipeline）：
// 喂 export_case_v2 的 window.json（真摔）→ 生产 handleMessage/handleEventMessage → 真 beliefShadowTick，
// 断言 DBN belief shadow 抬起 P(Fallen)（真摔不被压成 Vacant）。委员会建序「第一个 recall 闭环」。
//
// loader 已迁至 testkit 包（LoadWindow/EventCategory/LegoV2Record），本文件 import 使用。

// mustLoadWindow 桥接 testkit.LoadWindow → 旧 legoLoadWindow(t, dir) 调用方。
// 只做 path join + error→t.Fatal，不重写逻辑。
func mustLoadWindow(t *testing.T, dir string) []testkit.LegoV2Record {
	t.Helper()
	recs, err := testkit.LoadWindow(filepath.Join(casesDir, dir))
	if err != nil {
		t.Fatalf("LoadWindow %s: %v", dir, err)
	}
	return recs
}

func TestRecallRealFall_201Handoff333B(t *testing.T) {
	// Step 2: 走 manifest.ResolveCase 加载（替代硬编码路径），断言强化。
	manifestPath := filepath.Join(casesDir, "legos", "manifest.json")
	m, err := testkit.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	var case3 *testkit.ManifestCase
	for i := range m.Cases {
		if m.Cases[i].ID == "case-3" {
			case3 = &m.Cases[i]
			break
		}
	}
	if case3 == nil {
		t.Fatal("manifest 中找不到 case-3")
	}
	recs, _, err := testkit.ResolveCase(*case3, casesDir)
	if err != nil {
		t.Fatalf("ResolveCase case-3: %v", err)
	}

	dir := case3.SourceFixture
	cfg, radarAddr, roomID := bLayout(t, dir)

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
	for _, r := range recs {
		if testkit.EventCategory(r.Category) {
			e.handleEventMessage(mk("event", r.Category, r.DataValue, r.Timestamp))
		} else { // track / heart → monitor
			e.handleMessage(nil, mk("monitor", r.Category, r.DataValue, r.Timestamp))
		}
	}

	// 断言强化(委员会 Step 2):P(Fallen)≥0.3 + fire + reason==pose_lying + veto_ghost=0。
	fired := logs.FilterMessage("belief_shadow_fall").Len()
	vetoed := logs.FilterMessage("belief_dbn_veto_ghost").Len()
	var peak float64
	var reason string
	for _, le := range logs.All() {
		if le.Message == "belief_shadow_trace" {
			if v, ok := le.ContextMap()["p_fallen"].(float64); ok && v > peak {
				peak = v
			}
		}
		if le.Message == "belief_shadow_fall" {
			if r, ok := le.ContextMap()["p7_3_reason"].(string); ok {
				reason = r
			}
		}
	}
	t.Logf("case-3: fire=%d veto_ghost=%d peak=%.3f reason=%s", fired, vetoed, peak, reason)
	if fired == 0 {
		t.Fatalf("真摔 case-3 DBN 未 fire → recall 漏")
	}
	if peak < 0.3 {
		t.Fatalf("真摔 case-3 P(Fallen)=%.3f < 0.3 → recall 漏", peak)
	}
	if reason != "pose_lying" {
		t.Errorf("真摔 case-3 p7_3_reason=%s, want pose_lying", reason)
	}
	if vetoed > 0 {
		t.Errorf("真摔 case-3 veto_ghost=%d, 真摔不应被 ghost veto", vetoed)
	}
}
