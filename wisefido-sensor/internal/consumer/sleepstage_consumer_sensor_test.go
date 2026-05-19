// sleepstage_consumer_sensor_test.go — S4 (FU1) SleepStageConsumer T1 单元测试。
//
// 覆盖 ladder 行为 + 防 loop + handleRaw 主路径：
//   - confidenceFromDeviceType 4 类型
//   - intFromSleepStageField 5 类型
//   - ladderAdmits: 首入 / 同级覆盖 / 更高覆盖 / 更低拒绝
//   - handleRaw: sleepad 推 / radar 推 / 低 confidence drop / 自家 producer skip /
//     非 SleepStage category drop / 空 subject drop / publish 失败 state rollback

package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

type fakeSleepSink struct {
	calls   []fakeSleepCall
	errOnce error // 首次 publish 返回此 err（非 nil 测 rollback 路径）
	fired   int
}

type fakeSleepCall struct {
	cardID                string
	sleepStage, confidence int
}

func (f *fakeSleepSink) PublishBedSleepStage(ctx context.Context, cardID string, sleepStage, sleepConfidence int) error {
	f.fired++
	if f.errOnce != nil && f.fired == 1 {
		err := f.errOnce
		f.errOnce = nil
		return err
	}
	f.calls = append(f.calls, fakeSleepCall{cardID, sleepStage, sleepConfidence})
	return nil
}

func newTestSleepConsumer(sink SleepStageSink) *SleepStageConsumer {
	return NewSleepStageConsumer(nil, sink, zap.NewNop())
}

func TestConfidenceFromDeviceType(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"sleepad", 90},
		{"SleepPad", 90}, // case-insensitive
		{"Radar", 60},
		{"radar", 60},
		{"unknown", 0},
		{"", 0},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := confidenceFromDeviceType(tc.in); got != tc.want {
				t.Errorf("confidenceFromDeviceType(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestIntFromSleepStageField(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want int
	}{
		{"int", 2, 2},
		{"int64", int64(4), 4},
		{"float64", 2.0, 2},
		{"unsupported", "1", 0},
		{"nil", nil, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := intFromSleepStageField(tc.in); got != tc.want {
				t.Errorf("intFromSleepStageField(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestLadderAdmits_FirstEntry(t *testing.T) {
	c := newTestSleepConsumer(&fakeSleepSink{})
	if !c.ladderAdmits("card-a", 2, 60, 1000) {
		t.Error("first entry should be admitted")
	}
}

func TestLadderAdmits_HigherConfidenceOverrides(t *testing.T) {
	c := newTestSleepConsumer(&fakeSleepSink{})
	c.ladderAdmits("card-a", 2, 60, 1000) // radar
	if !c.ladderAdmits("card-a", 4, 90, 2000) {
		t.Error("higher confidence (sleepad 90 > radar 60) should override")
	}
}

func TestLadderAdmits_SameConfidenceOverrides(t *testing.T) {
	c := newTestSleepConsumer(&fakeSleepSink{})
	c.ladderAdmits("card-a", 2, 60, 1000)
	if !c.ladderAdmits("card-a", 4, 60, 2000) {
		t.Error("equal confidence should still admit (v1 行为：更高或同级覆盖)")
	}
}

func TestLadderAdmits_LowerConfidenceRejected(t *testing.T) {
	c := newTestSleepConsumer(&fakeSleepSink{})
	c.ladderAdmits("card-a", 2, 90, 1000) // sleepad
	if c.ladderAdmits("card-a", 4, 60, 2000) {
		t.Error("lower confidence (radar 60 < sleepad 90) should be rejected")
	}
}

// ---------------------------------------------------------------------------
// handleRaw 主路径
// ---------------------------------------------------------------------------

func buildSleepRawFields(producer, subject, deviceType string, ts int64, sleepStage int) map[string]interface{} {
	_ = ts // 用 timestamp string 不用 int64
	data := map[string]interface{}{observation.FieldSleepStage: float64(sleepStage)}
	dvBytes, _ := json.Marshal([]interface{}{data})
	return map[string]interface{}{
		"producer":               producer,
		"subject_entity":         subject,
		"sequence_number":        "1",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            deviceType,
		"timestamp":              "1700000000000",
		"topic_type":             "event",
		"category":               alarm.SleepStage,
		rediscommon.DataValueKey: string(dvBytes),
	}
}

func TestHandleRaw_SleepadPublishedWith90(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000000000, 2)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(sink.calls))
	}
	if sink.calls[0].confidence != 90 || sink.calls[0].sleepStage != 2 {
		t.Errorf("sleepad: conf=%d stage=%d want 90/2", sink.calls[0].confidence, sink.calls[0].sleepStage)
	}
}

func TestHandleRaw_RadarPublishedWith60(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"Radar", 1700000000000, 4)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 1 || sink.calls[0].confidence != 60 || sink.calls[0].sleepStage != 4 {
		t.Errorf("radar: got %+v want conf=60 stage=4", sink.calls)
	}
}

func TestHandleRaw_RadarAfterSleepad_Rejected(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"Radar", 1700000060000, 4))
	if len(sink.calls) != 1 {
		t.Errorf("radar after sleepad should be ladder-rejected: %d publishes", len(sink.calls))
	}
}

func TestHandleRaw_SelfProducerSkipped(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:fff1::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000000000, 2)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 0 {
		t.Errorf("self-producer must skip, got %d publishes", len(sink.calls))
	}
}

func TestHandleRaw_NonSleepStageCategoryDropped(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	data := map[string]interface{}{observation.FieldSleepStage: float64(2)}
	dvBytes, _ := json.Marshal([]interface{}{data})
	raw := map[string]interface{}{
		"producer":               "fd00:0:3:111:3:101::1",
		"subject_entity":         "fd00:0:3:111:3:101::/96",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            "sleepad",
		"timestamp":              "1700000000000",
		"topic_type":             "event",
		"category":               alarm.Fall, // 非 SleepStage
		rediscommon.DataValueKey: string(dvBytes),
	}
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 0 {
		t.Errorf("non-SleepStage category should drop, got %d", len(sink.calls))
	}
}

func TestHandleRaw_EmptySubjectDropped(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "", "sleepad", 1700000000000, 2)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 0 {
		t.Errorf("empty subject should drop, got %d", len(sink.calls))
	}
}

func TestHandleRaw_SleepStageZeroDropped(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000000000, 0)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 0 {
		t.Errorf("sleepStage=0 should drop, got %d", len(sink.calls))
	}
}

func TestHandleRaw_UnknownDeviceTypeDropped(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"unknown-device", 1700000000000, 2)
	c.handleRaw(context.Background(), raw)
	if len(sink.calls) != 0 {
		t.Errorf("unknown device_type should drop, got %d", len(sink.calls))
	}
}

func TestHandleRaw_PublishFailRollsBackState(t *testing.T) {
	sink := &fakeSleepSink{errOnce: errors.New("redis fail")}
	c := newTestSleepConsumer(sink)
	raw1 := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000000000, 2)
	c.handleRaw(context.Background(), raw1) // publish fail → state rollback

	// 第二轮：同卡同 confidence 应允许重试（state 已回滚）
	raw2 := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000060000, 2)
	c.handleRaw(context.Background(), raw2)
	if len(sink.calls) != 1 {
		t.Errorf("after rollback retry should succeed, got %d publishes", len(sink.calls))
	}
}
