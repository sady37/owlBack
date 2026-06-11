package belief

import (
	"math"
	"testing"
)

// TestFallLRFromDwellEquivalence — P4.1 单源化等价复现:toilet/open 日间 = 旧 inline 1+(d/scale)²（封顶）。
func TestFallLRFromDwellEquivalence(t *testing.T) {
	// Bathroom Toilet: 20min tail = 1200s
	check := func(roomType, areaType int, scaleSec, d float64) {
		want := 1 + (d/scaleSec)*(d/scaleSec)
		if want > dwellFallCap {
			want = dwellFallCap
		}
		got := fallLRFromDwell(d, 1.0, roomType, areaType, false)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("rt=%d at=%d d=%.0f：survival=%.4f 应 = %.4f", roomType, areaType, d, got, want)
		}
	}
	check(1, 7, 1200, 300)    // Bathroom Toilet 20min: d=300, LR=1+(300/1200)^2=1.0625
	check(1, 7, 1200, 1800)   // 过封顶
	check(0, 4, 60, 30)       // default 60s: d=30, LR=1+(30/60)^2=1.25
	check(0, 4, 60, 90)       // 过封顶
}

// TestFallLRFromDwellZoneTable — bed/enter/unknown 不在尾表 → 1.0 中性(不报);dwell≤0 → 1.0。
func TestFallLRFromDwellZoneTable(t *testing.T) {
	// AreaBed(2) and AreaDeny(5) excluded → LR=1.0
	for _, a := range []int{2, 5} {
		if lr := fallLRFromDwell(600, 1.0, 0, a, false); lr != 1.0 {
			t.Fatalf("area=%d(Bed/Deny) 应不报(LR=1.0),得 %.4f", a, lr)
		}
	}
	// Unknown/Enter → 60s tail → LR > 1.0 (no longer silently excluded)
	if lr := fallLRFromDwell(10, 1.0, 0, 0, false); lr <= 1.0 {
		t.Fatalf("area=Unknown d=10s 应有非零 LR,得 %.4f", lr)
	}
	if lr := fallLRFromDwell(0, 1.0, 0, 0, false); lr != 1.0 {
		t.Fatalf("dwell=0 应中性,得 %.4f", lr)
	}
}

// TestFallLRFromDwellNightShortensTail — P4.3 夜间短尾:同 dwell 夜间 LR > 日间(更敏感);单调上升;封顶。
func TestFallLRFromDwellNightShortensTail(t *testing.T) {
	const d = 50.0
	const areaActive = 4 // AreaActive → default 60s tail
	day := fallLRFromDwell(d, 1.0, 0, areaActive, false)
	night := fallLRFromDwell(d, 1.0, 0, areaActive, true)
	if !(night > day) {
		t.Fatalf("夜间短尾应更敏感:night LR(%.4f) > day(%.4f)", night, day)
	}
	// default tail = 60s. night shortens: want = 1+(d/(60*dwellNightTailMult))²
	const defaultTailSec = 60.0
	wantNight := 1 + math.Pow(d/(defaultTailSec*dwellNightTailMult), dwellShape)
	if wantNight > dwellFallCap {
		wantNight = dwellFallCap
	}
	if math.Abs(night-wantNight) > 1e-9 {
		t.Fatalf("night LR=%.4f 应 = 1+(d/(%.0f×%.2f))^%.0f=%.4f", night, defaultTailSec, dwellNightTailMult, dwellShape, wantNight)
	}
	// 单调:dwell 越久 LR 越高（到封顶）。
	if fallLRFromDwell(10, 1.0, 0, 4, false) >= fallLRFromDwell(30, 1.0, 0, 4, false) {
		t.Fatalf("dwell 单调上升错")
	}
}

// TestFallLRFromDwellToleranceLengthensTail — 开阔地 cell tolerance 拉长尾:tol>1 → 同 dwell LR 更低(久站真人不报)。
func TestFallLRFromDwellToleranceLengthensTail(t *testing.T) {
	const d = 50.0
	noTol := fallLRFromDwell(d, 1.0, 0, 3, false)
	hiTol := fallLRFromDwell(d, 2.0, 0, 3, false)
	if !(hiTol < noTol) {
		t.Fatalf("高 tolerance 应拉长尾→LR 更低:hiTol(%.4f) < noTol(%.4f)", hiTol, noTol)
	}
	// tol<1 视为 1（不放宽）。
	if fallLRFromDwell(d, 0.5, 0, 3, false) != noTol {
		t.Fatalf("tol<1 应当 1.0,不该收紧尾")
	}
}
