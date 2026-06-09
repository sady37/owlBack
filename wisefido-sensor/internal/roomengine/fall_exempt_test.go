package roomengine

import (
	"testing"

	"owl-common/radarutils"
)

// humanBedGrid 造一个测试网格：bed 区(100..200)人工标定 Conf=99；learn 区(300..400)SourceHuman 但 Conf=95。
func humanBedGrid(t *testing.T) *RoomGrid {
	t.Helper()
	g := NewRoomGrid(800, 400, 0)
	g.SetPrior(radarutils.Rect{X1: 100, Y1: 100, X2: 200, Y2: 200}, AreaBed, 99, SourceHuman) // 真人工床
	g.SetPrior(radarutils.Rect{X1: 300, Y1: 100, X2: 400, Y2: 200}, AreaBed, 95, SourceHuman) // 自学习 lock
	return g
}

// TestIsHumanBedConfidenceCut — P7.4 核心判别:Conf≥99 人工床=豁免;Conf=95 自学习不豁免;非床/nil 不豁免。
func TestIsHumanBedConfidenceCut(t *testing.T) {
	g := humanBedGrid(t)
	if !isHumanBedAt(g, 150, 150) {
		t.Fatalf("Conf=99 人工床应豁免(isHumanBedAt=true)")
	}
	if isHumanBedAt(g, 350, 150) {
		t.Fatalf("Conf=95 自学习 lock 不该当人工床豁免(应 false)——命名 quirk 不可误判")
	}
	if isHumanBedAt(g, 600, 350) {
		t.Fatalf("非床区不该豁免")
	}
	if isHumanBedAt(nil, 150, 150) {
		t.Fatalf("nil grid 应默认不豁免")
	}
}

// TestHumanBedVetoAt — P7.4 veto:任一 fall 位置在人工床 → veto;全不在 → 不 veto。
func TestHumanBedVetoAt(t *testing.T) {
	g := humanBedGrid(t)
	if !humanBedVetoAt(g, [][2]int{{600, 350}, {150, 150}}) {
		t.Fatalf("任一位置(150,150)在人工床应 veto")
	}
	if humanBedVetoAt(g, [][2]int{{600, 350}, {350, 150}}) {
		t.Fatalf("全不在人工床(95 自学习不算)不该 veto")
	}
	if humanBedVetoAt(g, nil) {
		t.Fatalf("无位置不该 veto")
	}
}
