package roomengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_recall_realdata_test.go — Tier-1 recall 闭环（真数据走真 pipeline）：
// 喂 export_case_v2 的 window.json（真摔）→ 生产 handleMessage/handleEventMessage → 真 beliefShadowTick，
// 断言 DBN belief shadow 抬起 P(Fallen)（真摔不被压成 Vacant）。委员会建序「第一个 recall 闭环」。
//
// loader 把 v2 window.json（{category,device_uid,timestamp,data_value}，无 topic_type）按 category
// 推断 topic_type（event 名→event / track·heart→monitor），喂生产 StreamMessage（同归一化，含 IPv6）。

// legoEventCategory v2 window 里属"事件流"的 category（其余 track/heart 走 monitor）。
func legoEventCategory(cat string) bool {
	switch cat {
	case "Fall", "EnterRoom", "ExitRoom", "InBed", "LeftBed", "number_people",
		"activity", "Walking", "Sitting", "Standing":
		return true
	}
	return false
}

type legoV2Record struct {
	Category  string                   `json:"category"`
	DeviceUID string                   `json:"device_uid"`
	Timestamp int64                    `json:"timestamp"`
	DataValue []map[string]interface{} `json:"data_value"`
}

// legoLoadWindow 读 v2 window.json，按 ts 升序。
func legoLoadWindow(t *testing.T, dir string) []legoV2Record {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(casesDir, dir, "window.json"))
	if err != nil {
		t.Skipf("window 缺 %s: %v", dir, err)
	}
	var recs []legoV2Record
	if err := json.Unmarshal(raw, &recs); err != nil {
		t.Fatalf("解析 window %s: %v", dir, err)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Timestamp < recs[j].Timestamp })
	return recs
}

func TestRecallRealFall_201Handoff333B(t *testing.T) {
	dir := "unit201-handoff-0609-bathroom-333B" // 201 bathroom 0609 真摔(firmware Fall pose=5)
	cfg, radarAddr, roomID := bLayout(t, dir)
	recs := legoLoadWindow(t, dir)

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
		if legoEventCategory(r.Category) {
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
