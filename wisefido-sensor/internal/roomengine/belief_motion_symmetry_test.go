package roomengine

import "testing"

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
