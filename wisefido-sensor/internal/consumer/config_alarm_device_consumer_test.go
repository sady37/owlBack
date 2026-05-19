// config_alarm_device_consumer_test.go — Tier1 单元测试：CloudEvents envelope 解析 + Invalidate 调用。

package consumer

import (
	"encoding/json"
	"sync"
	"testing"

	"go.uber.org/zap"
)

// fakeInvalidator 记录 Invalidate 调用，供测试断言。
type fakeInvalidator struct {
	mu      sync.Mutex
	called  []string
}

func (f *fakeInvalidator) Invalidate(deviceAddr string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called = append(f.called, deviceAddr)
}

func (f *fakeInvalidator) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.called))
	copy(out, f.called)
	return out
}

func makeRaw(t *testing.T, deviceAddr, settingType string) map[string]interface{} {
	t.Helper()
	inner, err := json.Marshal(map[string]interface{}{
		"device_addr":  deviceAddr,
		"setting_type": settingType,
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := json.Marshal(map[string]interface{}{
		"type": "config.alarmDevice",
		"data": json.RawMessage(inner),
	})
	if err != nil {
		t.Fatal(err)
	}
	return map[string]interface{}{"data": string(envelope)}
}

func newTestConfigConsumer(inv AlarmEnablementInvalidator) *AlarmDeviceConfigConsumer {
	return &AlarmDeviceConfigConsumer{invalidator: inv, logger: zap.NewNop()}
}

func TestAlarmDeviceConfigConsumer_HappyPath_InvalidatesByDeviceAddr(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	addr := "fd00:0:3:111:3:101::abcd"
	c.handleRaw(makeRaw(t, addr, "alarm"))
	got := inv.Calls()
	if len(got) != 1 || got[0] != addr {
		t.Errorf("expected exactly 1 Invalidate(%q), got %+v", addr, got)
	}
}

func TestAlarmDeviceConfigConsumer_EmptyDeviceAddr_NoOp(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	c.handleRaw(makeRaw(t, "", "alarm"))
	if got := inv.Calls(); len(got) != 0 {
		t.Errorf("empty device_addr must not invalidate, got %+v", got)
	}
}

func TestAlarmDeviceConfigConsumer_MissingDataField_NoOp(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	c.handleRaw(map[string]interface{}{}) // no "data" key
	if got := inv.Calls(); len(got) != 0 {
		t.Errorf("missing data field must not invalidate, got %+v", got)
	}
}

func TestAlarmDeviceConfigConsumer_BadEnvelopeJSON_NoOp(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	c.handleRaw(map[string]interface{}{"data": "not-json{{{"})
	if got := inv.Calls(); len(got) != 0 {
		t.Errorf("bad envelope must not invalidate, got %+v", got)
	}
}

func TestAlarmDeviceConfigConsumer_BadInnerJSON_NoOp(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	envelope, _ := json.Marshal(map[string]interface{}{
		"type": "config.alarmDevice",
		"data": json.RawMessage(`not-json`),
	})
	c.handleRaw(map[string]interface{}{"data": string(envelope)})
	if got := inv.Calls(); len(got) != 0 {
		t.Errorf("bad inner data must not invalidate, got %+v", got)
	}
}

func TestAlarmDeviceConfigConsumer_MultipleSettingTypes_InvalidatesEach(t *testing.T) {
	inv := &fakeInvalidator{}
	c := newTestConfigConsumer(inv)
	addrA := "fd00:0:3:111:3:101::a"
	addrB := "fd00:0:3:111:3:101::b"
	c.handleRaw(makeRaw(t, addrA, "alarm"))
	c.handleRaw(makeRaw(t, addrB, "monitor"))
	c.handleRaw(makeRaw(t, addrA, "sleep"))
	got := inv.Calls()
	if len(got) != 3 {
		t.Fatalf("want 3 invalidate calls, got %d: %+v", len(got), got)
	}
	if got[0] != addrA || got[1] != addrB || got[2] != addrA {
		t.Errorf("invalidate order mismatch, got %+v", got)
	}
}
