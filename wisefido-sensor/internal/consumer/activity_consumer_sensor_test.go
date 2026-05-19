// activity_consumer_sensor_test.go — S2 (FU4) ActivityConsumer T1 单元测试。
//
// table-driven 覆盖：
//   - intFromActivityField    — 4 类型 + nil
//   - handleRaw               — activity 字段透传 / 非 activity drop / 防 loop / 空 subject drop

package consumer

import (
	"encoding/json"
	"testing"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

func newTestActivityConsumer(sink ActivitySink) *ActivityConsumer {
	return NewActivityConsumer(nil, sink, zap.NewNop())
}

type fakeActivitySink struct {
	calls []fakeActivityCall
}

type fakeActivityCall struct {
	spatialPrefix, deviceAddr                                              string
	tsMs                                                                   int64
	walkDistanceMeters, walkDurationSec, standDurationSec, multiPersonSec  int
}

func (f *fakeActivitySink) PushEventFields(spatialPrefix, deviceAddr string, tsMs int64,
	walkDistanceMeters, walkDurationSec, standDurationSec, multiPersonDurationSec int) {
	f.calls = append(f.calls, fakeActivityCall{
		spatialPrefix:      spatialPrefix,
		deviceAddr:         deviceAddr,
		tsMs:               tsMs,
		walkDistanceMeters: walkDistanceMeters,
		walkDurationSec:    walkDurationSec,
		standDurationSec:   standDurationSec,
		multiPersonSec:     multiPersonDurationSec,
	})
}

func TestIntFromActivityField(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want int
	}{
		{"int", 7, 7},
		{"int64", int64(13), 13},
		{"float64", 2.7, 2},
		{"unsupported", "5", 0},
		{"nil", nil, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := intFromActivityField(tc.in); got != tc.want {
				t.Errorf("intFromActivityField(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func buildActivityRawFields(producer, subject, category string, data map[string]interface{}) map[string]interface{} {
	dvBytes, _ := json.Marshal([]interface{}{data})
	return map[string]interface{}{
		"producer":               producer,
		"subject_entity":         subject,
		"sequence_number":        "1",
		"device_addr":            "fd00:0:3:111:3:101::1",
		"device_type":            "Radar",
		"timestamp":              "1700000060000",
		"topic_type":             "event",
		"category":               category,
		rediscommon.DataValueKey: string(dvBytes),
	}
}

func TestActivityHandleRaw_PushesAllFields(t *testing.T) {
	sink := &fakeActivitySink{}
	c := newTestActivityConsumer(sink)
	raw := buildActivityRawFields(
		"fd00:0:3:111:3:101::1",
		"fd00:0:3:111:3:101::/96",
		alarm.Activity,
		map[string]interface{}{
			observation.FieldWalkDistance:        float64(3),
			observation.FieldWalkDuration:        float64(12),
			observation.FieldStandDuration:       float64(55),
			observation.FieldMultiPersonDuration: float64(0),
		},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 1 {
		t.Fatalf("expected 1 push, got %d", len(sink.calls))
	}
	got := sink.calls[0]
	if got.spatialPrefix != "fd00:0:3:111:3:101::/96" {
		t.Errorf("spatialPrefix=%q", got.spatialPrefix)
	}
	if got.walkDistanceMeters != 3 || got.walkDurationSec != 12 ||
		got.standDurationSec != 55 || got.multiPersonSec != 0 {
		t.Errorf("fields mismatch: walk_d=%d walk_t=%d stand=%d mp=%d",
			got.walkDistanceMeters, got.walkDurationSec, got.standDurationSec, got.multiPersonSec)
	}
}

func TestActivityHandleRaw_SelfProducerSkipped(t *testing.T) {
	sink := &fakeActivitySink{}
	c := newTestActivityConsumer(sink)
	raw := buildActivityRawFields(
		"fd00:0:fff1::1", // sensor slot
		"fd00:0:3:111:3:101::/96",
		alarm.Activity,
		map[string]interface{}{observation.FieldWalkDistance: float64(3)},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("self-producer should be skipped, got %d calls", len(sink.calls))
	}
}

func TestActivityHandleRaw_NonActivityCategoryDropped(t *testing.T) {
	sink := &fakeActivitySink{}
	c := newTestActivityConsumer(sink)
	raw := buildActivityRawFields(
		"fd00:0:3:111:3:101::1",
		"fd00:0:3:111:3:101::/96",
		alarm.Fall, // 非 activity
		map[string]interface{}{},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("non-activity category should be dropped, got %d calls", len(sink.calls))
	}
}

func TestActivityHandleRaw_EmptySubjectDropped(t *testing.T) {
	sink := &fakeActivitySink{}
	c := newTestActivityConsumer(sink)
	raw := buildActivityRawFields(
		"fd00:0:3:111:3:101::1",
		"", // 缺 SubjectEntity
		alarm.Activity,
		map[string]interface{}{observation.FieldWalkDistance: float64(3)},
	)
	c.handleRaw(raw)
	if len(sink.calls) != 0 {
		t.Fatalf("empty subject_entity should be dropped, got %d calls", len(sink.calls))
	}
}
