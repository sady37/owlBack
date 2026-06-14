package roomengine

import "testing"

// TestLostFarFromRadar — P2 距离闸边界:距雷达 > DistanceGateCm(500)→ 远距弱回波区(suppress);≤ 不 suppress。
func TestLostFarFromRadar(t *testing.T) {
	g := FallRulesParam.Lost.DistanceGateCm
	if g <= 0 {
		t.Skipf("DistanceGateCm 未启用(%d)", g)
	}
	if lostFarFromRadar(g) {
		t.Fatalf("距=gate(%d) 不该判远(严格 >)", g)
	}
	if !lostFarFromRadar(g + 1) {
		t.Fatalf("距=gate+1(%d) 应判远", g+1)
	}
	if lostFarFromRadar(0) {
		t.Fatalf("距=0(近/未 stash)不该判远")
	}
	if lostFarFromRadar(g - 100) {
		t.Fatalf("距 gate 内不该判远")
	}
}

// TestRoomLedgerEmptyWrapper — G-2 空房账自锁 wrapper:lastExit>lastEnter=空(残影抑制) / 反之非空 / 默认非空。
func TestRoomLedgerEmptyWrapper(t *testing.T) {
	tm := &TrackManager{}
	if tm.RoomLedgerEmpty() {
		t.Fatalf("默认(无事件)应非空——铁律保守不抑制")
	}
	tm.lastEnterMs, tm.lastExitMs = 1000, 2000 // ExitRoom 晚于 EnterRoom → 房空
	if !tm.RoomLedgerEmpty() {
		t.Fatalf("lastExit>lastEnter 应判空(残影抑制)")
	}
	tm.lastEnterMs, tm.lastExitMs = 3000, 2000 // 又 EnterRoom → 非空
	if tm.RoomLedgerEmpty() {
		t.Fatalf("lastEnter>lastExit 应判非空(人又进来)")
	}
}
