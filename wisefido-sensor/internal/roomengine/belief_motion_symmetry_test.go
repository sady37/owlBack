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

// TestDBNMotionSymmetry_SideBySideTwoReal — 整改单(委员会 713daed):**同向两真人并排走** = motion 对称单独
// **分不开「self 是反射」vs「self 是第二个真人」**(继承 tm.checkMotionSymmetry 同限)。**显式文档化**:此模式
// **判 ghost**(false-positive),**「一切看风险」下接受**——共存(≥2 track)=有人在场=可救=误判走动真人为 ghost
// 的摔=低代价(有证人),且 P1④ OthersPresent→τ*↑亦已为共存场景升抑制阈。**非默认对,是 risk-accepted**;
// cos/dist 阈对真并排走数据判别力=post-cutover 标定;增量2 mirror 对称(几何位置)才能真分开反射 vs 第二真人。
func TestDBNMotionSymmetry_SideBySideTwoReal(t *testing.T) {
	a := TrackStatusBase{TrackID: 1, X: 30, Y: 0, MoveActive: true}                        // 真人 A:(0,0)→(30,0)
	b := TrackStatusBase{TrackID: 2, X: 30, Y: 60, MoveActive: true, Verdict: VerdictReal} // 真人 B 并排:(0,60)→(30,60),距 60cm 同向
	prev := map[int][2]int{1: {0, 0}, 2: {0, 60}}
	if dbnMotionSymmetryGhost(&a, []TrackStatusBase{a, b}, prev) {
		t.Logf("同向两真人并排 → 判 ghost(motion 对称限,risk-accepted:共存可救+P1④τ*↑;增量2 mirror 待真分开反射vs第二真人)")
	} else {
		t.Logf("注:同向两真人未判 ghost(阈/增量2 已分开?)→ 更新文档")
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
