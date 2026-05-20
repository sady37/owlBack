// device_status_tracker_test.go — Tier1：OfflineCallback 注册 + edge 检测。
//
// 覆盖：
//   - RegisterOfflineCallback 多个 callback 全注册
//   - fireOffline 广播到所有注册者，参数透传 deviceAddr
//   - recordOffline edge 检测：prev online → false (edge) / prev offline → false (no edge) /
//     未见过 device → false (no edge)
//   - nil tracker / nil fn 安全

package consumer

import (
	"testing"

	"go.uber.org/zap"
)

func newTestStatusTracker() *DeviceStatusTracker {
	return &DeviceStatusTracker{
		state:  make(map[string]*deviceLiveness),
		logger: zap.NewNop(),
	}
}

func TestRegisterOfflineCallback_Dispatches(t *testing.T) {
	tr := newTestStatusTracker()
	var got1, got2 []string
	tr.RegisterOfflineCallback(func(addr string) { got1 = append(got1, addr) })
	tr.RegisterOfflineCallback(func(addr string) { got2 = append(got2, addr) })

	tr.fireOffline("fd00::1")
	tr.fireOffline("fd00::2")

	if len(got1) != 2 || got1[0] != "fd00::1" || got1[1] != "fd00::2" {
		t.Errorf("callback 1 got %v, want [fd00::1 fd00::2]", got1)
	}
	if len(got2) != 2 || got2[0] != "fd00::1" || got2[1] != "fd00::2" {
		t.Errorf("callback 2 got %v, want [fd00::1 fd00::2]", got2)
	}
}

func TestRegisterOfflineCallback_NilSafe(t *testing.T) {
	var tr *DeviceStatusTracker
	tr.RegisterOfflineCallback(func(string) {}) // 不 panic

	tr2 := newTestStatusTracker()
	tr2.RegisterOfflineCallback(nil) // 不 panic
	tr2.fireOffline("fd00::1")       // 无 callback 也不 panic
}

func TestRecordOffline_EdgeFromOnline(t *testing.T) {
	tr := newTestStatusTracker()
	tr.state["fd00::1"] = &deviceLiveness{online: true, deviceType: "radar"}

	edge := tr.recordOffline("fd00::1", "radar")
	if !edge {
		t.Errorf("online → offline should return edge=true")
	}
	if tr.state["fd00::1"].online {
		t.Errorf("state should be marked offline")
	}
}

func TestRecordOffline_NoEdgeFromOffline(t *testing.T) {
	tr := newTestStatusTracker()
	tr.state["fd00::1"] = &deviceLiveness{online: false, deviceType: "radar"}

	edge := tr.recordOffline("fd00::1", "radar")
	if edge {
		t.Errorf("offline → offline should return edge=false (callback would be duplicate)")
	}
}

func TestRecordOffline_NoEdgeFromUnseen(t *testing.T) {
	tr := newTestStatusTracker()

	edge := tr.recordOffline("fd00::99", "radar")
	if edge {
		t.Errorf("unseen → offline should return edge=false (no prior state to clear)")
	}
	if tr.state["fd00::99"] == nil {
		t.Errorf("recordOffline should create state entry for unseen device")
	}
	if tr.state["fd00::99"].online {
		t.Errorf("new entry should be offline")
	}
}

func TestRecordOffline_BroadcastWiring(t *testing.T) {
	tr := newTestStatusTracker()
	var calls []string
	tr.RegisterOfflineCallback(func(addr string) { calls = append(calls, addr) })

	tr.state["fd00::1"] = &deviceLiveness{online: true, deviceType: "radar"}
	if tr.recordOffline("fd00::1", "radar") {
		tr.fireOffline("fd00::1")
	}
	if tr.recordOffline("fd00::1", "radar") {
		tr.fireOffline("fd00::1") // 二次调用应被 edge gate 挡掉
	}

	if len(calls) != 1 || calls[0] != "fd00::1" {
		t.Errorf("edge-gated dispatch should fire exactly once, got %v", calls)
	}
}
