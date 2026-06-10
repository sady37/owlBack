package roomengine

import (
	"testing"

	"wisefido-sensor/internal/roomengine/belief"
)

// TestDBNMotionSymmetry — P1-final:DBN 自有 motion 对称 ghost(绕开 gate-list b.Verdict)。
// 紧贴(<100cm)+同向(cos>0.866) = 多径反射跟随真人→ghost;独立反向(两真人)→否;孤立→否。
func TestDBNMotionSymmetry(t *testing.T) {
	self := TrackStatusBase{TrackID: 1, X: 30, Y: 0, MoveActive: true}                              // (0,0)→(30,0) +x
	partnerSame := TrackStatusBase{TrackID: 2, X: 80, Y: 0, MoveActive: true, Verdict: VerdictReal} // (50,0)→(80,0) +x 同向,距50
	prev := map[int][2]int{1: {0, 0}, 2: {50, 0}}
	if !dbnMotionSymmetryGhost(&self, []TrackStatusBase{self, partnerSame}, prev) {
		t.Errorf("★紧贴+同向 应判 motion 对称 ghost(self 是反射)")
	}
	partnerOpp := TrackStatusBase{TrackID: 2, X: 20, Y: 0, MoveActive: true, Verdict: VerdictReal} // (50,0)→(20,0) -x 反向
	if dbnMotionSymmetryGhost(&self, []TrackStatusBase{self, partnerOpp}, prev) {
		t.Errorf("★反向运动(两真人独立) 不应判 ghost(委员会细化1:2 真人)")
	}
	if dbnMotionSymmetryGhost(&self, []TrackStatusBase{self}, prev) {
		t.Errorf("★孤立无 partner 不应判 ghost")
	}
}

// TestDBNMovingReason — P3 整改单#2(委员会 a48a09b):moving/pose_lying 判别单测(词表单源后)。
func TestDBNMovingReason(t *testing.T) {
	const now = int64(100_000)
	// pose_lying 主导 + 摔前在动(lastMove 3s 前<5s)→ moving
	if r := dbnMovingReason(belief.ReasonPoseLying, now-3_000, now); r != belief.ReasonMoving {
		t.Errorf("摔前在动应 ReasonMoving,得 %v", r)
	}
	// pose_lying + lastMove 越窗(10s 前>5s)→ 保持 pose_lying(开阔地静躺)
	if r := dbnMovingReason(belief.ReasonPoseLying, now-10_000, now); r != belief.ReasonPoseLying {
		t.Errorf("lastMove 越窗应 ReasonPoseLying,得 %v", r)
	}
	// pose_lying + 从未动(lastMove=0)→ pose_lying
	if r := dbnMovingReason(belief.ReasonPoseLying, 0, now); r != belief.ReasonPoseLying {
		t.Errorf("从未动应 ReasonPoseLying,得 %v", r)
	}
	// lost/silent 主导不受影响(只 pose_lying 才细分 moving)
	if r := dbnMovingReason(belief.ReasonLost, now-1_000, now); r != belief.ReasonLost {
		t.Errorf("lost 不应变 moving,得 %v", r)
	}
	if r := dbnMovingReason(belief.ReasonSilent, now-1_000, now); r != belief.ReasonSilent {
		t.Errorf("silent 不应变 moving,得 %v", r)
	}
}
