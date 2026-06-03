package roomengine

import "testing"

// 同房多雷达占用对账（A 简版）：本台 lost-fall pending 成熟时，另一台雷达近期仍见真人 →
// 人还在房（被别台看着，本台只是丢/幻影）→ 抑制。治 D523 无床雷达持幻影 + 09E7 同房看真人那类 FP。

func newMaturedPending(dev string, now int64) *PendingLostFall {
	return &PendingLostFall{
		OriginalTrackID: 0,
		DeviceAddr:      dev,
		RoomID:          "test-room",
		LastX:           100,
		LastY:           100,
		LastCellArea:    AreaActive,      // → 默认 walkway wait（几分钟）
		DisappearMs:     now - 60*60_000, // 1h 前消失，远超任何 wait → 必达 fire 闸
	}
}

func TestLostFall_SuppressedByOtherRadarInRoom(t *testing.T) {
	tm, _ := newTestTM()
	var now int64 = 1_700_000_000_000
	tm.pendingLostFalls[0] = newMaturedPending("devA", now)
	tm.lastRealTrackByDevice["devB"] = now - 2_000 // 同房另一台 2s 前见真人

	tickAt(tm, now)

	if len(tm.pendingLostFalls) != 0 {
		t.Fatalf("pending 应被同房另一雷达取消，仍剩 %d", len(tm.pendingLostFalls))
	}
	if tm.lostFallPendingCancelled == 0 {
		t.Error("应记 cancelled（by other radar in room）")
	}
	if tm.lostFallReported != 0 {
		t.Errorf("不应 fire，lostFallReported=%d", tm.lostFallReported)
	}
}

func TestLostFall_FiresWhenNoOtherRadar(t *testing.T) {
	tm, _ := newTestTM()
	var now int64 = 1_700_000_000_000
	tm.pendingLostFalls[0] = newMaturedPending("devA", now)
	tm.lastRealTrackByDevice["devA"] = now - 2_000 // 只有本台自己（被 exclude）→ 不抑制

	tickAt(tm, now)

	if tm.lostFallReported == 0 {
		t.Error("同房无他台真人 → 应正常 fire")
	}
}

func TestOtherDeviceRealTrackRecent(t *testing.T) {
	tm, _ := newTestTM()
	var now int64 = 1_700_000_000_000
	tm.lastRealTrackByDevice["devB"] = now - 2_000

	if !tm.otherDeviceRealTrackRecent("devA", now) {
		t.Error("devB 2s 前活，对 devA 应为 true")
	}
	if tm.otherDeviceRealTrackRecent("devB", now) {
		t.Error("只有 devB 自己（被 exclude）应为 false")
	}
	if tm.otherDeviceRealTrackRecent("devA", now+otherDeviceRealTTLMs+1_000) {
		t.Error("超 TTL 应为 false")
	}
}
