// ai_verdict_handler_test.go — S5a AIVerdictHandler T1 单元测试。
//
// 覆盖 handleRaw 主路径：合法 → Set / invalid device / track_id ≤ 0 / nil data / 解析失败。

package consumer

import (
	"encoding/json"
	"net/netip"
	"testing"

	rediscommon "owl-common/redis"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// canonicalAddr 把 IPv6 字符串转为 netip canonical 形式（与 FromStreamMap 内部一致）。
func canonicalAddr(s string) string {
	a, _ := netip.ParseAddr(s)
	return a.String()
}

func newTestVerdictHandler() (*AIVerdictHandler, *service.AIOverrideCache) {
	cache := service.NewAIOverrideCache("release", 60, zap.NewNop())
	h := NewAIVerdictHandler(nil, cache, zap.NewNop())
	return h, cache
}

func buildVerdictRawFields(producer, deviceAddr string, ts int64, trackID, confidence int, source, reason string) map[string]interface{} {
	data := map[string]interface{}{
		"track_id":         float64(trackID),
		"track_confidence": float64(confidence),
		"source":           source,
		"reason":           reason,
	}
	dvBytes, _ := json.Marshal([]interface{}{data})
	return map[string]interface{}{
		"producer":               producer,
		"subject_entity":         "",
		"sequence_number":        "1",
		"device_addr":            deviceAddr,
		"device_type":            "Radar",
		"timestamp":              "1700000000000",
		"topic_type":             "event",
		"category":               "track_verdict",
		rediscommon.DataValueKey: string(dvBytes),
	}
}

func TestAIVerdictHandle_SetsCache(t *testing.T) {
	h, cache := newTestVerdictHandler()
	addr := "fd00:0:3:111:3:101::1"
	raw := buildVerdictRawFields("AI.Caregiver01", addr,
		1_700_000_000_000, 3, 20, "AI.Caregiver01", "ghost_penalty")
	h.handleRaw(raw)

	if cache.Stats().SetCount != 1 {
		t.Fatalf("Set should be called once, got %d", cache.Stats().SetCount)
	}

	// release 模式下 Apply 应命中（用 canonical IPv6 lookup，与 handleRaw 内部一致）
	fields := map[string]interface{}{"track_confidence": float64(80)}
	cache.Apply(canonicalAddr(addr), 3, fields)
	if fields["track_confidence"] != 20 {
		t.Errorf("after handle: confidence should be overridden 20, got %v", fields["track_confidence"])
	}
	if fields["ai_source"] != "AI.Caregiver01" {
		t.Errorf("ai_source not set: %v", fields["ai_source"])
	}
}

func TestAIVerdictHandle_InvalidDeviceAddrDropped(t *testing.T) {
	h, cache := newTestVerdictHandler()
	raw := buildVerdictRawFields("AI.X", "invalid-addr", 1_700_000_000_000, 3, 20, "AI.X", "ghost")
	h.handleRaw(raw)
	if cache.Stats().SetCount != 0 {
		t.Errorf("invalid device_addr should drop, got %d sets", cache.Stats().SetCount)
	}
}

func TestAIVerdictHandle_TrackIDZeroDropped(t *testing.T) {
	h, cache := newTestVerdictHandler()
	raw := buildVerdictRawFields("AI.X", "fd00:0:3:111:3:101::1", 1_700_000_000_000,
		0, 20, "AI.X", "ghost")
	h.handleRaw(raw)
	if cache.Stats().SetCount != 0 {
		t.Errorf("tid=0 should drop, got %d sets", cache.Stats().SetCount)
	}
}

func TestAIVerdictHandle_TrackIDNegativeDropped(t *testing.T) {
	h, cache := newTestVerdictHandler()
	raw := buildVerdictRawFields("AI.X", "fd00:0:3:111:3:101::1", 1_700_000_000_000,
		-1, 20, "AI.X", "ghost")
	h.handleRaw(raw)
	if cache.Stats().SetCount != 0 {
		t.Errorf("negative tid should drop, got %d sets", cache.Stats().SetCount)
	}
}

func TestAIVerdictHandle_NilDataDropped(t *testing.T) {
	h, cache := newTestVerdictHandler()
	raw := map[string]interface{}{
		"producer":               "AI.X",
		"subject_entity":         "",
		"sequence_number":        "1",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            "Radar",
		"timestamp":              "1700000000000",
		"topic_type":             "event",
		"category":               "track_verdict",
		rediscommon.DataValueKey: "[null]", // dataValue=[null] → FirstDataValue returns nil
	}
	h.handleRaw(raw)
	if cache.Stats().SetCount != 0 {
		t.Errorf("nil data should drop, got %d sets", cache.Stats().SetCount)
	}
}

func TestAIVerdictHandle_HighConfidenceVerdictApplies(t *testing.T) {
	h, cache := newTestVerdictHandler()
	addr := "fd00:0:3:111:3:101::1"
	raw := buildVerdictRawFields("AI.X", addr, 1_700_000_000_000,
		3, 85, "AI.X", "high_confidence_real")
	h.handleRaw(raw)
	fields := map[string]interface{}{"track_confidence": float64(30)}
	cache.Apply(canonicalAddr(addr), 3, fields)
	if fields["track_confidence"] != 85 {
		t.Errorf("AI verdict can raise confidence too (not just lower); got %v", fields["track_confidence"])
	}
}
