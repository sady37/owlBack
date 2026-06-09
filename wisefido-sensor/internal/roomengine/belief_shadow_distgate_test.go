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
