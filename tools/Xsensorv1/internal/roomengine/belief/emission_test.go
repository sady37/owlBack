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

// E4：δ floor-strip → Fallen 抬、AtBed 压。
func TestEmissionDeltaFloorStrip(t *testing.T) {
	em := NewEmission(bed1Geom())
	js := NewJointSpace(1)
	logPhi := em.LogPhi(js, Observation{RadarOnline: true, FloorStripXY: true, Sleepad: []BedReading{BedNoReport}})
	dF := logPhi[js.idx(SFallen, 0)]
	dBed := logPhi[js.idx(SBed, 0)]
	if dF <= 0 || dBed >= 0 || dF <= dBed {
		t.Errorf("δ floor-strip：logPhi(F)=%.4f 应>0、logPhi(AtBed)=%.4f 应<0、且 F>AtBed", dF, dBed)
	}
}

// E5：cd2b 离线态 —— sleepad 离线 + HR/RR absent 无独立源 + pose lying + floor-strip XY。
// 经一帧 Correct，SFallen 后验应超 SBed（δ≫0 几何救回，不靠 sleepad/HRRR）。
func TestEmissionCd2bOfflineFallen(t *testing.T) {
	model := DefaultModel()
	f := NewFilter(model, 1)
	js := f.Space()
	geom := bed1Geom()
	em := NewEmission(geom)
	cp := NewCoupling(geom) // κ=1 全物理表 → ψ_phys(F,occ)=ε_art 拉 B→vac

	// 初始：床占用主导（陈旧 occ），人原在床。
	for i := range f.alpha {
		f.alpha[i] = math.Inf(-1)
	}
	f.alpha[js.idx(SBed, 1)] = 0
	f.alpha.LogNormalize()

	o := Observation{
		RadarOnline: true, PoseLying: true, StillSec: 120,
		NearBed: true, HRRRObserved: true, HRRRPresent: false, VitalSourceOnline: false, // radar absent 零信息
		FloorStripXY: true,                      // δ：床沿地条（cd2b 主解）
		Sleepad:      []BedReading{BedNoReport}, // sleepad 离线
	}
	gxy := []float64{1.0} // 雷达能定位到床
	// 多帧推进（离线 B 弛豫 + Ψ ε_art + δ 持续把 S 拉向 Fallen）。
	for step := 0; step < 90; step++ {
		f.Step(int64(step+1)*1000, bedOnline{false}, cp.LogPsi(js, gxy), em.LogPhi(js, o))
	}
	ms := js.MarginalS(f.alpha)
	if ms[SFallen] <= ms[SBed] {
		t.Errorf("cd2b 离线态：P(Fallen)=%.4f 应 > P(AtBed)=%.4f（δ 几何救回）", ms[SFallen], ms[SBed])
	}
	t.Logf("cd2b 离线态 ✓ P(Fallen)=%.4f > P(AtBed)=%.4f（不靠 sleepad/HRRR，δ≫0 几何）", ms[SFallen], ms[SBed])
}
