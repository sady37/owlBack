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
	"net/netip"
	"strconv"
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
	data := map[string]interface{}{observation.FieldSleepStage: float64(sleepStage)}
	dvBytes, _ := json.Marshal([]interface{}{data})
	tsStr := strconv.FormatInt(ts, 10)
	return map[string]interface{}{
		"producer":               producer,
		"subject_entity":         subject,
		"sequence_number":        "1",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            deviceType,
		"timestamp":              tsStr,
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

// ---------------------------------------------------------------------------
// C: OOB drop + device_failure emit
// ---------------------------------------------------------------------------

type fakeBedChecker struct {
	occupied map[string]bool
}

func (f *fakeBedChecker) IsBedOccupied(sp string) bool {
	if f == nil || f.occupied == nil {
		return false
	}
	return f.occupied[sp]
}

type fakeFailEmitter struct {
	calls []fakeFailCall
}

type fakeFailCall struct {
	deviceAddr    netip.Addr
	subjectEntity string
	eventName     string
	level         string
	tsMs          int64
	triggerData   map[string]interface{}
}

func (f *fakeFailEmitter) PublishAlarmFire(ctx context.Context, deviceAddr netip.Addr, subjectEntity, eventName, level string, tsMs int64, triggerData map[string]interface{}) (string, error) {
	f.calls = append(f.calls, fakeFailCall{deviceAddr, subjectEntity, eventName, level, tsMs, triggerData})
	return "1-0", nil
}

// 编译期验证 interface 兼容
var (
	_ BedOccupancyChecker  = (*fakeBedChecker)(nil)
	_ DeviceFailureEmitter = (*fakeFailEmitter)(nil)
)

func TestHandleRaw_OOB_DropsAndEmitsDeviceFailure(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	bedChecker := &fakeBedChecker{occupied: map[string]bool{}} // 全部 Vacant
	failEmit := &fakeFailEmitter{}
	c.SetBedChecker(bedChecker)
	c.SetDeviceFailureEmitter(failEmit)

	raw := buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
		"sleepad", 1700000000000, 2)
	c.handleRaw(context.Background(), raw)

	if len(sink.calls) != 0 {
		t.Errorf("OOB should drop publish: got %d sink calls", len(sink.calls))
	}
	if len(failEmit.calls) != 1 {
		t.Fatalf("OOB should emit device_failure once: got %d", len(failEmit.calls))
	}
	got := failEmit.calls[0]
	if got.eventName != alarm.AlarmTypeDeviceFailure {
		t.Errorf("eventName = %q, want DeviceFailure", got.eventName)
	}
	if got.level != alarm.AlarmLevelErr {
		t.Errorf("level = %q, want Err", got.level)
	}
	if got.triggerData["reason"] != "sleepstage_out_of_bed" {
		t.Errorf("triggerData.reason = %v, want sleepstage_out_of_bed", got.triggerData["reason"])
	}
}

func TestHandleRaw_OOB_DedupWithin5min(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.SetBedChecker(&fakeBedChecker{occupied: map[string]bool{}}) // 全 Vacant
	failEmit := &fakeFailEmitter{}
	c.SetDeviceFailureEmitter(failEmit)

	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))
	// 同 spatial 30s 后再触发 OOB：dedup 拦截
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000030000, 4))

	if len(failEmit.calls) != 1 {
		t.Errorf("OOB within 5min: should dedup to 1 alarm, got %d", len(failEmit.calls))
	}
}

func TestHandleRaw_OOB_AdmitAfter5min(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.SetBedChecker(&fakeBedChecker{occupied: map[string]bool{}})
	failEmit := &fakeFailEmitter{}
	c.SetDeviceFailureEmitter(failEmit)

	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))
	// 5min+ 后再触发：放行
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000+5*60*1000, 4))

	if len(failEmit.calls) != 2 {
		t.Errorf("OOB after 5min dedup window: should admit second alarm, got %d", len(failEmit.calls))
	}
}

func TestHandleRaw_BedOccupied_NormalLadder(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.SetBedChecker(&fakeBedChecker{occupied: map[string]bool{
		"fd00:0:3:111:3:101::/96": true,
	}})
	c.SetDeviceFailureEmitter(&fakeFailEmitter{})

	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))

	if len(sink.calls) != 1 {
		t.Errorf("bed occupied: should publish normally, got %d", len(sink.calls))
	}
}

func TestHandleRaw_BedCheckerNil_BackwardCompat(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink) // 不 SetBedChecker

	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))

	if len(sink.calls) != 1 {
		t.Errorf("no bedChecker: should fall back to normal ladder, got %d publishes", len(sink.calls))
	}
}

func TestHandleRaw_OOB_NoEmitterStillDrops(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.SetBedChecker(&fakeBedChecker{occupied: map[string]bool{}})
	// 不 SetDeviceFailureEmitter

	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 2))

	if len(sink.calls) != 0 {
		t.Errorf("OOB still drops even without emitter: got %d publishes", len(sink.calls))
	}
}

// ---------------------------------------------------------------------------
// D: SleepStage clear on bed transition
// ---------------------------------------------------------------------------

func TestOnBedVacant_ClearsStateAndEmitsZero(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)

	// 先填 state
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"sleepad", 1700000000000, 4))
	if len(sink.calls) != 1 {
		t.Fatalf("setup: expected 1 publish, got %d", len(sink.calls))
	}

	// 触发 vacant
	c.OnBedVacant(context.Background(), "fd00:0:3:111:3:101::/96", 1700000060000)

	if len(sink.calls) != 2 {
		t.Fatalf("OnBedVacant should emit clear publish, got %d total", len(sink.calls))
	}
	got := sink.calls[1]
	if got.sleepStage != 0 || got.confidence != 0 {
		t.Errorf("clear publish should be (0,0), got (%d,%d)", got.sleepStage, got.confidence)
	}

	// state 应已清掉：下一个 lower-confidence radar event 也应能 admit
	c.handleRaw(context.Background(),
		buildSleepRawFields("fd00:0:3:111:3:101::1", "fd00:0:3:111:3:101::/96",
			"Radar", 1700000120000, 2))
	if len(sink.calls) != 3 {
		t.Errorf("after clear, radar=60 should admit (state cleared); got %d total", len(sink.calls))
	}
}

func TestOnBedVacant_NoPriorState_NoOp(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)

	c.OnBedVacant(context.Background(), "fd00:0:3:111:3:101::/96", 1700000000000)

	if len(sink.calls) != 0 {
		t.Errorf("OnBedVacant with no prior state: should be no-op, got %d publishes", len(sink.calls))
	}
}

func TestOnBedVacant_EmptyPrefix_NoOp(t *testing.T) {
	sink := &fakeSleepSink{}
	c := newTestSleepConsumer(sink)
	c.OnBedVacant(context.Background(), "", 1700000000000)
	if len(sink.calls) != 0 {
		t.Errorf("empty prefix: should be no-op")
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
