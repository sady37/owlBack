// track_manager_clear_device_test.go — TrackManager.ClearDevice T1 单元测试。
//
// 验证 "device offline = 内存重启" 原则：清 bedSessions + sleepadStates；不动 tracks。

package roomengine

import (
	"testing"
)

func TestClearDevice_RemovesBedSessionAndSleepadState(t *testing.T) {
	tm, _ := newTestTM()
	addr := "fd00:0:3:111:3:101::abcd"

	tm.bedSessions[addr] = &BedSession{InBedSinceMs: 1_700_000_000_000}
	tm.sleepadStates[addr] = &SleepadObservation{InBed: true}

	tm.ClearDevice(addr)

	if _, ok := tm.bedSessions[addr]; ok {
		t.Error("ClearDevice: bedSessions entry should be cleared")
	}
	if _, ok := tm.sleepadStates[addr]; ok {
		t.Error("ClearDevice: sleepadStates entry should be cleared")
	}
}

func TestClearDevice_OtherDevicesUntouched(t *testing.T) {
	tm, _ := newTestTM()
	addrA := "fd00:0:3:111:3:101::a"
	addrB := "fd00:0:3:111:3:101::b"

	tm.bedSessions[addrA] = &BedSession{InBedSinceMs: 1000}
	tm.bedSessions[addrB] = &BedSession{InBedSinceMs: 2000}
	tm.sleepadStates[addrA] = &SleepadObservation{InBed: true}
	tm.sleepadStates[addrB] = &SleepadObservation{InBed: false}

	tm.ClearDevice(addrA)

	if _, ok := tm.bedSessions[addrA]; ok {
		t.Error("addrA bedSessions should be cleared")
	}
	if _, ok := tm.bedSessions[addrB]; !ok {
		t.Error("addrB bedSessions should be untouched")
	}
	if _, ok := tm.sleepadStates[addrA]; ok {
		t.Error("addrA sleepadStates should be cleared")
	}
	if _, ok := tm.sleepadStates[addrB]; !ok {
		t.Error("addrB sleepadStates should be untouched")
	}
}

func TestClearDevice_TracksNotTouched(t *testing.T) {
	tm, _ := newTestTM()
	addr := "fd00:0:3:111:3:101::abcd"

	// 加 track（按 trackID 索引，非 deviceAddr）
	tm.tracks[7] = &TrackState{TrackID: 7, DeviceAddr: addr, Kalman: NewKalmanFilter2D(50, 50)}
	tm.bedSessions[addr] = &BedSession{InBedSinceMs: 1000}

	tm.ClearDevice(addr)

	// bedSession 清
	if _, ok := tm.bedSessions[addr]; ok {
		t.Error("bedSession should be cleared")
	}
	// tracks 不清（自然 evict 走 timeout）
	if _, ok := tm.tracks[7]; !ok {
		t.Error("tracks should NOT be touched (per design: trackID 索引非 deviceAddr，自然 evict 走 timeout)")
	}
}

func TestClearDevice_NilReceiverNoCrash(t *testing.T) {
	var tm *TrackManager
	tm.ClearDevice("fd00:0:3:111:3:101::abcd") // 不应 panic
}

func TestClearDevice_EmptyAddrNoOp(t *testing.T) {
	tm, _ := newTestTM()
	tm.bedSessions["fd00:0:3:111:3:101::a"] = &BedSession{InBedSinceMs: 1000}

	tm.ClearDevice("")

	if _, ok := tm.bedSessions["fd00:0:3:111:3:101::a"]; !ok {
		t.Error("empty addr: should not affect existing entries")
	}
}

// Engine.OnDeviceUnfit — 反查 deviceRoom 后转发 tm.ClearDevice
func TestEngineOnDeviceUnfit_DispatchesToCorrectRoom(t *testing.T) {
	tm, _ := newTestTM()
	addr := "fd00:0:3:111:3:101::abcd"
	tm.bedSessions[addr] = &BedSession{InBedSinceMs: 1000}

	e := newTestEngineForPublicBathroom(t)
	e.rooms[tm.roomID] = tm
	e.deviceRoom[addr] = tm.roomID

	e.OnDeviceUnfit(addr)

	if _, ok := tm.bedSessions[addr]; ok {
		t.Error("Engine.OnDeviceUnfit should dispatch to tm.ClearDevice")
	}
}

func TestEngineOnDeviceUnfit_UnroutedDeviceNoOp(t *testing.T) {
	e := newTestEngineForPublicBathroom(t)
	// 不应 panic（未路由 device 在 deviceRoom map 里返 ""，rooms[""]==nil）
	e.OnDeviceUnfit("fd00:0:3:111:3:101::abcd")
}
