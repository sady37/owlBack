package belief

import (
	"math"
	"testing"
)

// TestFallLRFromDwellEquivalence — P4.1 单源化等价复现:toilet/open 日间 = 旧 inline 1+(d/scale)²（封顶）。
func TestFallLRFromDwellEquivalence(t *testing.T) {
	check := func(zone Geom, scale, d float64) {
		want := 1 + (d/scale)*(d/scale)
		if want > dwellFallCap {
			want = dwellFallCap
		}
		got := fallLRFromDwell(d, 1.0, zone, false)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("zone=%v d=%.0f：survival=%.4f 应 = 旧 inline %.4f", zone, d, got, want)
		}
	}
	check(GeomInToilet, dwellScaleToiletSec, 300)
	check(GeomInToilet, dwellScaleToiletSec, 1800) // 过封顶
	check(GeomOpenFloor, dwellScaleOpenSec, 240)
	check(GeomOpenFloor, dwellScaleOpenSec, 600)
}

// TestFallLRFromDwellZoneTable — bed/enter/unknown 不在尾表 → 1.0 中性(不报);dwell≤0 → 1.0。
func TestFallLRFromDwellZoneTable(t *testing.T) {
	for _, z := range []Geom{GeomInBed, GeomInEnter, GeomUnknown} {
		if lr := fallLRFromDwell(600, 1.0, z, false); lr != 1.0 {
			t.Fatalf("zone=%v 应不报(LR=1.0),得 %.4f", z, lr)
		}
	}
	if lr := fallLRFromDwell(0, 1.0, GeomOpenFloor, false); lr != 1.0 {
		t.Fatalf("dwell=0 应中性,得 %.4f", lr)
	}
}

// TestFallLRFromDwellNightShortensTail — P4.3 夜间短尾:同 dwell 夜间 LR > 日间(更敏感);单调上升;封顶。
func TestFallLRFromDwellNightShortensTail(t *testing.T) {
	const d = 300.0
	day := fallLRFromDwell(d, 1.0, GeomOpenFloor, false)
	night := fallLRFromDwell(d, 1.0, GeomOpenFloor, true)
	if !(night > day) {
		t.Fatalf("夜间短尾应更敏感:night LR(%.4f) > day(%.4f)", night, day)
	}
	// 夜间等价于 scale×mult：night LR == 1+(d/(scale×mult))²（未封顶时）。
	wantNight := 1 + math.Pow(d/(dwellScaleOpenSec*dwellNightTailMult), dwellShape)
	if wantNight > dwellFallCap {
		wantNight = dwellFallCap
	}
	if math.Abs(night-wantNight) > 1e-9 {
		t.Fatalf("night LR=%.4f 应 = 1+(d/(scale×%.2f))^%.0f=%.4f", night, dwellNightTailMult, dwellShape, wantNight)
	}
	// 单调:dwell 越久 LR 越高（到封顶）。
	if fallLRFromDwell(120, 1.0, GeomOpenFloor, false) >= fallLRFromDwell(240, 1.0, GeomOpenFloor, false) {
		t.Fatalf("dwell 单调上升错")
	}
}

// TestFallLRFromDwellToleranceLengthensTail — 开阔地 cell tolerance 拉长尾:tol>1 → 同 dwell LR 更低(久站真人不报)。
func TestFallLRFromDwellToleranceLengthensTail(t *testing.T) {
	const d = 300.0
	noTol := fallLRFromDwell(d, 1.0, GeomOpenFloor, false)
	hiTol := fallLRFromDwell(d, 2.0, GeomOpenFloor, false)
	if !(hiTol < noTol) {
		t.Fatalf("高 tolerance 应拉长尾→LR 更低:hiTol(%.4f) < noTol(%.4f)", hiTol, noTol)
	}
	// tol<1 视为 1（不放宽）。
	if fallLRFromDwell(d, 0.5, GeomOpenFloor, false) != noTol {
		t.Fatalf("tol<1 应当 1.0,不该收紧尾")
	}
}
