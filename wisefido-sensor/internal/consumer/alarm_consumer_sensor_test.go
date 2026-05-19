// alarm_consumer_sensor_test.go — S1 (FU3) AlarmConsumer T1 单元测试（同 package，纯逻辑无 IO）。
//
// 覆盖 4 个纯函数 + handleRaw 主路径（用 fakeSink 记录调用）：
//   - normalizeWeakBioCategory  — 4 base + 4 High/Low 变体 + 1 不识别
//   - isSelfProducer            — slot 内 / hardcoded name / sensor. 前缀 / 非平台 IPv6 / 非法字符串 / 空
//   - intFromAlarmField         — int / int64 / float64 / 不支持类型 / nil
//   - handleRaw                 — parse / 防 loop / SubjectEntity 缺失 / category 不识别 / WeakBio state×20 / HR

package consumer

import (
	"encoding/json"
	"testing"

	"owl-common/alarm"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

func newTestAlarmConsumer(sink AlarmSink) *AlarmConsumer {
	return NewAlarmConsumer(nil, sink, zap.NewNop())
}

// fakeSink 记录 PushAlarmFields 调用供断言。
type fakeSink struct {
	calls []fakeSinkCall
}

type fakeSinkCall struct {
	spatialPrefix, alarmType, producer string
	tsMs                               int64
	rawValue                           int
}

func (f *fakeSink) PushAlarmFields(spatialPrefix, alarmType, producer string, tsMs int64, rawValue int) {
	f.calls = append(f.calls, fakeSinkCall{spatialPrefix, alarmType, producer, tsMs, rawValue})
}

// -----------------------------------------------------------------------------
// normalizeWeakBioCategory
// -----------------------------------------------------------------------------

func TestNormalizeWeakBioCategory(t *testing.T) {
	tests := []struct {
		input    string
		wantBase string
		wantOK   bool
	}{
		{alarm.WeakBiometricSignal, alarm.WeakBiometricSignal, true},
		{alarm.HeartRateAlert, alarm.HeartRateAlert, true},
		{alarm.HeartRateAlertHigh, alarm.HeartRateAlert, true},
		{alarm.HeartRateAlertLow, alarm.HeartRateAlert, true},
		{alarm.RespRateAlert, alarm.RespRateAlert, true},
		{alarm.RespRateAlertHigh, alarm.RespRateAlert, true},
		{alarm.RespRateAlertLow, alarm.RespRateAlert, true},
		{alarm.ApneaHypopnea, alarm.ApneaHypopnea, true},
		{alarm.Fall, "", false},
		{alarm.InBed, "", false},
		{"", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := normalizeWeakBioCategory(tc.input)
			if got != tc.wantBase || ok != tc.wantOK {
				t.Fatalf("normalizeWeakBioCategory(%q) = (%q, %v), want (%q, %v)",
					tc.input, got, ok, tc.wantBase, tc.wantOK)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// isSelfProducer
// -----------------------------------------------------------------------------

func TestIsSelfProducer(t *testing.T) {
	c := newTestAlarmConsumer(&fakeSink{})

	tests := []struct {
		name     string
		producer string
		want     bool
	}{
		{"empty", "", false},
		{"hardcoded wisefido-sensor", "wisefido-sensor", true},
		{"sensor. prefix caregiver01", "sensor.caregiver01", true},
		{"sensor. prefix any", "sensor.foo", true},
		{"slot fd00:0:fff1::1 sensor node-1", "fd00:0:fff1::1", true},
		{"slot fd00:0:fff1::ffff sensor node-other", "fd00:0:fff1::ffff", true},
		{"non-slot IPv6 device", "fd00:0:3:111:3:101:a2ac:d523", false},
		{"non-slot IPv6 other agent slot", "fd00:0:fff2::1", false}, // 不同 agent slot
		{"non-IPv6 string", "garbage-string", false},
		{"hardcoded must be exact", "sensor", false}, // 没 "." 不匹配前缀
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := c.isSelfProducer(tc.producer)
			if got != tc.want {
				t.Fatalf("isSelfProducer(%q) = %v, want %v", tc.producer, got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// intFromAlarmField
// -----------------------------------------------------------------------------

func TestIntFromAlarmField(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want int
	}{
		{"int", 3, 3},
		{"int64", int64(7), 7},
		{"float64 truncates", 2.9, 2},
		{"float64 zero", 0.0, 0},
		{"unsupported string", "5", 0},
		{"nil", nil, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := intFromAlarmField(tc.in)
			if got != tc.want {
				t.Fatalf("intFromAlarmField(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// handleRaw — 主路径
// -----------------------------------------------------------------------------

// buildRawStreamFields 构造一条 stream-form alarm 消息（与 owl-common/redis.FromStreamMap 互逆）。
// dataValue 走 JSON 走 string 形态（同 sensor publisher path）。
func buildRawStreamFields(producer, subject, category string, tsMs int64, data map[string]interface{}) map[string]interface{} {
	dvBytes, _ := json.Marshal([]interface{}{data})
	return map[string]interface{}{
		"producer":               producer,
		"subject_entity":         subject,
		"sequence_number":        "1",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            "Radar",
		"timestamp":              "1700000000000",
		"topic_type":             "alarm",
		"category":               category,
		rediscommon.DataValueKey: string(dvBytes),
		// envelope timestamp 已在 timestamp 字段；下面 explicit override 给 tsMs
	}
}

func TestHandleRaw_WeakBioState60_PushedWithRaw60(t *testing.T) {
	sink := &fakeSink{}
	c := newTestAlarmConsumer(sink)
	raw := buildRawStreamFields(
		"fd00:0:3:111:3:101::1",                                           // device producer (非 sensor)
		"fd00:0:3:111:3:101::/96",                                         // /96 bed
		alarm.WeakBiometricSignal,
		1700000000000,
		map[string]interface{}{"state": float64(3)}, // firmware state=3 → raw=60
	)
	c.handleRaw(raw)
	if len(sink.calls) != 1 {
		t.Fatalf("expected 1 push, got %d", len(sink.calls))
	}
	got := sink.calls[0]
	if got.alarmType != alarm.WeakBiometricSignal {
		t.Errorf("alarmType=%q want %q", got.alarmType, alarm.WeakBiometricSignal)
	}
	if got.rawValue != 60 {
		t.Errorf("rawValue=%d want 60 (state=3 * 20)", got.rawValue)
	}
	if got.spatialPrefix != "fd00:0:3:111:3:101::/96" {
		t.Errorf("spatialPrefix=%q", got.spatialPrefix)
	}
}

func TestHandleRaw_HighVariantNormalizedToBase(t *testing.T) {
	sink := &fakeSink{}
	c := newTestAlarmConsumer(sink)
	raw := buildRawStreamFields(
		"fd00:0:3:111:3:101::2",
		"fd00:0:3:111:3:101::/96",
		alarm.HeartRateAlertHigh,
		1700000000000,
		map[string]interface{}{},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 1 {
		t.Fatalf("expected 1 push, got %d", len(sink.calls))
	}
	if sink.calls[0].alarmType != alarm.HeartRateAlert {
		t.Errorf("alarmType=%q want %q (normalized base)", sink.calls[0].alarmType, alarm.HeartRateAlert)
	}
	if sink.calls[0].rawValue != 0 {
		t.Errorf("non-WeakBio rawValue should be 0, got %d", sink.calls[0].rawValue)
	}
}

func TestHandleRaw_SelfProducer_Skipped(t *testing.T) {
	sink := &fakeSink{}
	c := newTestAlarmConsumer(sink)
	raw := buildRawStreamFields(
		"fd00:0:fff1::1", // self sensor agent slot
		"fd00:0:3:111:3:101::/96",
		alarm.WeakBiometricSignal,
		1700000000000,
		map[string]interface{}{"state": float64(3)},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("expected 0 push (self loop guard), got %d", len(sink.calls))
	}
}

func TestHandleRaw_NonWeakBioCategory_Skipped(t *testing.T) {
	sink := &fakeSink{}
	c := newTestAlarmConsumer(sink)
	raw := buildRawStreamFields(
		"fd00:0:3:111:3:101::1",
		"fd00:0:3:111:3:101::/96",
		alarm.Fall, // event-stream category，不在 WeakBio 4 类
		1700000000000,
		map[string]interface{}{},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("expected 0 push (non-WeakBio), got %d", len(sink.calls))
	}
}

func TestHandleRaw_EmptySubjectEntity_Skipped(t *testing.T) {
	sink := &fakeSink{}
	c := newTestAlarmConsumer(sink)
	raw := buildRawStreamFields(
		"fd00:0:3:111:3:101::1",
		"", // 缺 SubjectEntity（cardagg 未绑卡 / R-009）
		alarm.WeakBiometricSignal,
		1700000000000,
		map[string]interface{}{"state": float64(3)},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("expected 0 push (empty subject_entity), got %d", len(sink.calls))
	}
}

// ---------------------------------------------------------------------------
// S6 设备类 alarm fan-out → FitnessSink
// ---------------------------------------------------------------------------

type fakeFitnessSink struct {
	marks   []fitnessCall
	clears  []fitnessCall
}

type fitnessCall struct {
	deviceAddr string
	reason     uint8
}

func (f *fakeFitnessSink) MarkUnfit(addr string, reason uint8) {
	f.marks = append(f.marks, fitnessCall{addr, reason})
}

func (f *fakeFitnessSink) ClearReason(addr string, reason uint8) {
	f.clears = append(f.clears, fitnessCall{addr, reason})
}

func TestClassifyDeviceFitnessAlarm(t *testing.T) {
	tests := []struct {
		cat       string
		wantR     uint8
		wantUnfit bool
		wantRec   bool
	}{
		{alarm.AlarmTypeOffline, fitnessReasonOffline, true, false},
		{alarm.AlarmTypeOfflineRecover, fitnessReasonOffline, false, true},
		{alarm.SensorDetached, fitnessReasonSensorDetached, true, false},
		{alarm.SensorDetachedRecover, fitnessReasonSensorDetached, false, true},
		{alarm.AngleException, fitnessReasonAngleException, true, false},
		{alarm.AngleExceptionRecover, fitnessReasonAngleException, false, true},
		{alarm.SignalPoor, fitnessReasonSignalPoor, true, false},
		{alarm.SingalPoorRecover, fitnessReasonSignalPoor, false, true},
		{alarm.WeakBiometricSignal, 0, false, false},
		{alarm.Fall, 0, false, false},
		{"", 0, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.cat, func(t *testing.T) {
			r, u, rec := classifyDeviceFitnessAlarm(tc.cat)
			if r != tc.wantR || u != tc.wantUnfit || rec != tc.wantRec {
				t.Errorf("classify(%q) = (%d, %v, %v), want (%d, %v, %v)",
					tc.cat, r, u, rec, tc.wantR, tc.wantUnfit, tc.wantRec)
			}
		})
	}
}

func TestHandleRaw_DeviceClassMarkUnfit(t *testing.T) {
	tests := []struct {
		category   string
		wantReason uint8
	}{
		{alarm.AlarmTypeOffline, fitnessReasonOffline},
		{alarm.SensorDetached, fitnessReasonSensorDetached},
		{alarm.AngleException, fitnessReasonAngleException},
		{alarm.SignalPoor, fitnessReasonSignalPoor},
	}
	for _, tc := range tests {
		t.Run(tc.category, func(t *testing.T) {
			sink := &fakeSink{}
			fs := &fakeFitnessSink{}
			c := newTestAlarmConsumer(sink)
			c.SetFitnessSink(fs)
			raw := buildRawStreamFields(
				"fd00:0:3:111:3:101::1",
				"fd00:0:3:111:3:101::/96",
				tc.category,
				1700000000000,
				map[string]interface{}{},
			)
			c.handleRaw(raw)
			if len(fs.marks) != 1 || fs.marks[0].reason != tc.wantReason {
				t.Errorf("%s should MarkUnfit reason=%d, got %+v", tc.category, tc.wantReason, fs.marks)
			}
			if len(fs.clears) != 0 {
				t.Errorf("%s should not Clear; got %+v", tc.category, fs.clears)
			}
			if len(sink.calls) != 0 {
				t.Errorf("device class should not enter WeakBio path, got %d sink calls", len(sink.calls))
			}
		})
	}
}

func TestHandleRaw_DeviceClassRecoverClearReason(t *testing.T) {
	tests := []struct {
		category   string
		wantReason uint8
	}{
		{alarm.AlarmTypeOfflineRecover, fitnessReasonOffline},
		{alarm.SensorDetachedRecover, fitnessReasonSensorDetached},
		{alarm.AngleExceptionRecover, fitnessReasonAngleException},
		{alarm.SingalPoorRecover, fitnessReasonSignalPoor},
	}
	for _, tc := range tests {
		t.Run(tc.category, func(t *testing.T) {
			sink := &fakeSink{}
			fs := &fakeFitnessSink{}
			c := newTestAlarmConsumer(sink)
			c.SetFitnessSink(fs)
			raw := buildRawStreamFields(
				"fd00:0:3:111:3:101::1",
				"fd00:0:3:111:3:101::/96",
				tc.category,
				1700000000000,
				map[string]interface{}{},
			)
			c.handleRaw(raw)
			if len(fs.clears) != 1 || fs.clears[0].reason != tc.wantReason {
				t.Errorf("%s should ClearReason=%d, got %+v", tc.category, tc.wantReason, fs.clears)
			}
			if len(fs.marks) != 0 {
				t.Errorf("%s should not Mark; got %+v", tc.category, fs.marks)
			}
		})
	}
}

func TestHandleRaw_DeviceClassSelfProducerSkipped(t *testing.T) {
	fs := &fakeFitnessSink{}
	c := newTestAlarmConsumer(&fakeSink{})
	c.SetFitnessSink(fs)
	raw := buildRawStreamFields(
		"fd00:0:fff1::1", // self sensor agent slot
		"fd00:0:3:111:3:101::/96",
		alarm.SensorDetached,
		1700000000000,
		map[string]interface{}{},
	)
	c.handleRaw(raw)
	if len(fs.marks) != 0 {
		t.Errorf("self producer should skip fitness fan-out; got %+v", fs.marks)
	}
}

func TestHandleRaw_NoFitnessSinkNoOp(t *testing.T) {
	c := newTestAlarmConsumer(&fakeSink{}) // no fitness sink wired
	raw := buildRawStreamFields(
		"fd00:0:3:111:3:101::1",
		"fd00:0:3:111:3:101::/96",
		alarm.AlarmTypeOffline,
		1700000000000,
		map[string]interface{}{},
	)
	// 不应 panic；device class 在没 wire 时直接 return
	c.handleRaw(raw)
}
