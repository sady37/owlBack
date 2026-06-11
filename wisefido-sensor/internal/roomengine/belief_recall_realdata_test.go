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
	dir := "unit201-handoff-0609-bathroom-333B" // 201 bathroom 0609 真摔(firmware Fall pose=5)
	cfg, radarAddr, roomID := bLayout(t, dir)
	recs, err := testkit.LoadWindow(filepath.Join(casesDir, dir))
	if err != nil {
		t.Fatalf("LoadWindow %s: %v", dir, err)
	}

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

	// 断言：DBN 在真摔上抬 P(Fallen)（recall：真摔不被压）。
	fired := logs.FilterMessage("belief_shadow_fall").Len()
	var peak float64
	for _, le := range logs.All() {
		if le.Message == "belief_shadow_trace" {
			if v, ok := le.ContextMap()["p_fallen"].(float64); ok && v > peak {
				peak = v
			}
		}
	}
	t.Logf("真摔 #9-333B：belief_shadow_fall=%d  peak P(Fallen)=%.3f", fired, peak)
	if fired == 0 && peak < 0.3 {
		t.Fatalf("真摔(201 0609 333B firmware Fall)DBN 既未 fire 也未抬 P(Fallen)(peak=%.3f) → recall 漏", peak)
	}
}
