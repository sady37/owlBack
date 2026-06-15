package belief

import (
	"math"
	"testing"
)

// emission_test.go — §5/§D 不变量验收：
//   E1 离线=中性（雷达 off + sleepad NoReport → logPhi 全 0）
//   E2 接触 InBed → B occ 抬（contact 轴只挂 B）
//   E3 §D HR/RR absent 门控（无独立源 → 不否决 AtBed；有 sleepad 在线 → 否决）
//   E4 δ floor-strip → Fallen 抬、AtBed 压（cd2b 主解，δ≫0 离线也可判）
//   E5 cd2b 离线态（sleepad off + HR/RR absent 无源 + pose lying + floor-strip）→ SFallen 浮出

const eeps = 1e-9

func bed1Geom() []BedGeom { return []BedGeom{{Covers: 1, Onbed: 1, Overlap: 1}} }

// E1：离线=中性。
func TestEmissionOfflineNeutral(t *testing.T) {
	em := NewEmission(bed1Geom())
	js := NewJointSpace(1)
	logPhi := em.LogPhi(js, Observation{RadarOnline: false, Sleepad: []BedReading{BedNoReport}})
	for i := 0; i < js.size; i++ {
		if math.Abs(logPhi[i]) > eeps {
			s, b := js.decode(i)
			t.Errorf("离线 logPhi(S=%v,bmask=%d)=%.6f 应=0（中性）", s, b, logPhi[i])
		}
	}
}

// E2：接触 InBed → B occ 抬。同 S 下 occ - vac = onbed·log(L_in)。
func TestEmissionContactInBed(t *testing.T) {
	em := NewEmission(bed1Geom())
	js := NewJointSpace(1)
	logPhi := em.LogPhi(js, Observation{RadarOnline: false, Sleepad: []BedReading{BedInBed}})
	idxOcc := js.idx(SBed, 1)
	idxVac := js.idx(SBed, 0)
	want := math.Log(em.p.lIn) // onbed=1
	if got := logPhi[idxOcc] - logPhi[idxVac]; math.Abs(got-want) > eeps {
		t.Errorf("InBed: occ-vac=%.4f 应=log(L_in)=%.4f", got, want)
	}
}

// E3：§D HR/RR absent 门控。
func TestEmissionHRRRAbsentGate(t *testing.T) {
	em := NewEmission(bed1Geom())
	js := NewJointSpace(1)
	base := Observation{
		RadarOnline: true, NearBed: true, HRRRObserved: true, HRRRPresent: false,
		Sleepad: []BedReading{BedNoReport},
	}
	// 无独立在线 vital 源 → radar 自身 absent 不否决 AtBed（零信息）。
	noSrc := em.LogPhi(js, base)
	if math.Abs(noSrc[js.idx(SBed, 0)]) > eeps {
		t.Errorf("无独立源 absent 不应否决 AtBed，logPhi(SBed)=%.4f 应=0", noSrc[js.idx(SBed, 0)])
	}
	// 有 sleepad 在线 → absent 否决 AtBed（§D gate 满足）。
	withSrc := base
	withSrc.VitalSourceOnline = true
	got := em.LogPhi(js, withSrc)
	want := math.Log(1.0 / em.p.lHR) // covers=1
	if math.Abs(got[js.idx(SBed, 0)]-want) > eeps {
		t.Errorf("有独立源 absent 应否决 AtBed，logPhi(SBed)=%.4f 应=log(1/L_hr)=%.4f", got[js.idx(SBed, 0)], want)
	}
}

// E4/E5（§32 框架涌现，替原 δ floor-strip + 离线补丁测试）—— cd2b 零补丁涌现（AC-拆1）。
// 框架纯路径：sleepad 在线 InBed→LeftBed，雷达 pose 躺+静止，**零 floor-strip、零离线**。
// 验 LeftBed→B vac→§4 Ψ 压 SBed 留 SFallen→pose 躺静止→(SFallen,vac) 自涌现 ≥55%。
func TestFrameworkEmergenceCd2b(t *testing.T) {
	geom := bed1Geom() // Overlap=1 → Ψ(SBed,vac)=1-o_j=0
	f := NewFilter(DefaultModel(), 1)
	js := f.Space()
	cp := NewCoupling(geom)
	em := NewEmission(geom)
	gxy := []float64{1.0}

	// 阶段A 在床睡：sleepad InBed + 躺 + 静止 → B occ、S→Bed。
	inBed := Observation{Sleepad: []BedReading{BedInBed}, RadarOnline: true, PoseLying: true, StillSec: 60, NearBed: true}
	for s := 1; s <= 30; s++ {
		f.Step(int64(s)*1000, BedOnline{true}, cp.LogPsi(js, gxy), em.LogPhi(js, inBed), 0, 1)
	}
	if pb := js.MarginalB(f.Alpha(), 0); pb < 0.99 {
		t.Fatalf("阶段A P(B occ)=%.4f 应≈1（在床）", pb)
	}

	// 阶段B 离床摔：sleepad LeftBed + 雷达仍躺 + 静止 + 近床。**零 floor-strip**。
	left := Observation{Sleepad: []BedReading{BedLeftBed}, RadarOnline: true, PoseLying: true, StillSec: 120, NearBed: true}
	for s := 31; s <= 120; s++ {
		f.Step(int64(s)*1000, BedOnline{true}, cp.LogPsi(js, gxy), em.LogPhi(js, left), 0, 1)
	}
	pF := js.PFallen(f.Alpha())
	if pF < 0.55 {
		t.Errorf("框架涌现失败：cd2b 零补丁 P(SFallen)=%.4f 应≥0.55（LeftBed→B vac→Ψ 涌现）", pF)
	}
	t.Logf("AC-拆1 ✓ 框架零补丁涌现：P(SFallen)=%.4f P(SBed)=%.4f P(B occ)=%.4f（无 floor-strip/无离线）",
		pF, js.MarginalS(f.Alpha())[SBed], js.MarginalB(f.Alpha(), 0))
}
