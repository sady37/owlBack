package belief

import (
	"math"
	"testing"
)

// multibed_test.go — 阶段4 多床验收（补 C §10 自认单床盲区：cd2b fixture |B|=1 测不到 |B|≥2 的
// mixture/routing/covers-max）。
//   MB1 §E mixture |B|=3 FN-safe（C 单床 fixture 触不到的多床塌缩）
//   MB2 跌倒/占用路由到正确床（attachment a_j 按 g^xy 归属）
//   MB3 covers max C2（多床异覆盖 → 雷达轴权重取 max_j）
//   MB4 probe 快照完整性（Σα=1、边缘一致）

const mbeps = 1e-9

// MB1：§E mixture |B|=3 FN-safe。bed0 陈旧 occ、bed1/2 vac、人摔 → Ψ(F) 存活，product 会压死。
func TestMixtureThreeBeds(t *testing.T) {
	geom := []BedGeom{
		{Covers: 1, Onbed: 1, Overlap: 1},
		{Covers: 1, Onbed: 1, Overlap: 1},
		{Covers: 1, Onbed: 1, Overlap: 1},
	}
	c := NewCoupling(geom)
	js := NewJointSpace(3)
	logPsi := c.LogPsi(js, []float64{1, 1, 1})

	bmask := 1 << 0 // bed0 occ，bed1/2 vac
	psiF := math.Exp(logPsi[js.idx(SFallen, bmask)])
	if psiF < 0.5 {
		t.Errorf("|B|=3 mixture Ψ(F,bed0-occ)=%.4f 应 >0.5（FN-safe，C 单床盲区现可验）", psiF)
	}
	prod := c.p.epsArt * 1 * 1 // product 手算会压死
	t.Logf("MB1 ✓ |B|=3 mixture Ψ(F)=%.4f 存活 vs product=%.4f（差 %.0f×）", psiF, prod, psiF/prod)
}

// MB2：占用路由到正确床。bed0 sleepad InBed + 雷达定位 bed0（g^xy=[1,0]），人在床 →
// P(B0=occ) 升、P(B1=occ) 不升；S* = SBed。验联合滤波多床归属正确。
func TestRoutingTwoBeds(t *testing.T) {
	geom := []BedGeom{
		{Covers: 1, Onbed: 1, Overlap: 1},
		{Covers: 1, Onbed: 1, Overlap: 1},
	}
	f := NewFilter(DefaultModel(), 2)
	js := f.Space()
	cp := NewCoupling(geom)
	em := NewEmission(geom)

	o := Observation{
		RadarOnline: true, PoseLying: true, NearBed: true,
		Sleepad: []BedReading{BedInBed, BedNoReport}, // 仅 bed0 InBed
	}
	gxy := []float64{1.0, 0.0} // 雷达只定位到 bed0
	online := bedOnline{true, false}
	for step := 0; step < 30; step++ {
		f.Step(int64(step+1)*1000, online, cp.LogPsi(js, gxy), em.LogPhi(js, o))
	}
	b0 := js.MarginalB(f.Alpha(), 0)
	b1 := js.MarginalB(f.Alpha(), 1)
	if b0 < 0.9 {
		t.Errorf("bed0 InBed 后 P(B0=occ)=%.4f 应高", b0)
	}
	if b1 > b0 {
		t.Errorf("证据只指 bed0，P(B1=occ)=%.4f 不应超 P(B0=occ)=%.4f（路由错床）", b1, b0)
	}
	if s, _ := js.MarginalS(f.Alpha()).Max(); s != SBed {
		t.Errorf("人在床 S*=%v 应=SBed", s)
	}
	t.Logf("MB2 ✓ 路由正确：P(B0)=%.3f >> P(B1)=%.3f，S*=SBed", b0, b1)
}

// MB3：covers max C2。多床异覆盖 [0.3, 1.0] → geom0Covers 取 max=1.0（保 Φ 分轴清洁，ground truth §F）。
func TestCoversMaxC2(t *testing.T) {
	em := NewEmission([]BedGeom{{Covers: 0.3, Onbed: 1}, {Covers: 1.0, Onbed: 1}})
	if w := em.geom0Covers(); math.Abs(w-1.0) > mbeps {
		t.Errorf("多床异覆盖 geom0Covers=%.4f 应=max=1.0（C2 裁定）", w)
	}
	// 无床房仍允许雷达轴全权
	if w := NewEmission(nil).geom0Covers(); math.Abs(w-1.0) > mbeps {
		t.Errorf("无床房 geom0Covers=%.4f 应=1.0", w)
	}
}

// MB4：probe 快照完整性。
func TestProbeSnapshot(t *testing.T) {
	geom := []BedGeom{{Covers: 1, Onbed: 1, Overlap: 1}, {Covers: 1, Onbed: 1, Overlap: 1}}
	f := NewFilter(DefaultModel(), 2)
	js := f.Space()
	cp := NewCoupling(geom)
	dec := Decision{Fire: false, Margin: -0.2}
	p := Snapshot(js, f, cp, dec, 2.5, 12345)

	if p.Ts != 12345 || len(p.MarginalB) != 2 || len(p.Kappa) != 2 || len(p.Alpha) != js.size {
		t.Fatalf("快照维度错：Ts=%d |B|=%d |κ|=%d |α|=%d(want %d)", p.Ts, len(p.MarginalB), len(p.Kappa), len(p.Alpha), js.size)
	}
	sum := 0.0
	for _, a := range p.Alpha {
		sum += a
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Errorf("快照 Σα=%.12f 应=1", sum)
	}
	sumS := 0.0
	for _, ps := range p.MarginalS {
		sumS += ps
	}
	if math.Abs(sumS-1.0) > 1e-9 {
		t.Errorf("快照 ΣP(S)=%.12f 应=1", sumS)
	}
	if p.Line() == "" {
		t.Error("Line() 不应空")
	}
}
