package belief

import (
	"math"
	"testing"
)

const tol = 1e-9

func sumJoint(v JointVector) float64 {
	s := 0.0
	for _, p := range v {
		s += p
	}
	return s
}

// T1 归一化守恒：任意 numBeds∈{0..3}，初始 + 连续 50 步 Predict 后 |Σα−1|<1e-9。
func TestT1NormalizationConserved(t *testing.T) {
	m := DefaultModel()
	for nb := 0; nb <= maxBeds; nb++ {
		f := NewFilter(m, nb)
		online := make([]bool, nb)
		for j := range online {
			online[j] = true
		}
		if d := math.Abs(sumJoint(f.Alpha()) - 1); d >= tol {
			t.Fatalf("numBeds=%d 初始 Σα 偏差 %.2e", nb, d)
		}
		for step := 1; step <= 50; step++ {
			f.Predict(online)
			if d := math.Abs(sumJoint(f.Alpha()) - 1); d >= tol {
				t.Fatalf("numBeds=%d 第 %d 步 Σα 偏差 %.2e ≥ %.0e", nb, step, d, tol)
			}
		}
	}
}

// T2 单床退化基数 + numBeds 显式持有（B1）+ numBeds=0 退化回单实体 Prior。
func TestT2DegenerateCardinality(t *testing.T) {
	m := DefaultModel()
	want := map[int]int{0: 9, 1: 18, 2: 36, 3: 72}
	for nb := 0; nb <= maxBeds; nb++ {
		f := NewFilter(m, nb)
		if got := f.Space().Size(); got != want[nb] {
			t.Fatalf("numBeds=%d size=%d 期望 %d", nb, got, want[nb])
		}
		if f.NumBeds() != nb {
			t.Fatalf("NumBeds()=%d 期望 %d（B1 显式持有）", f.NumBeds(), nb)
		}
	}
	// numBeds=0：MarginalS 初始逐分量 == model.Prior。
	f := NewFilter(m, 0)
	ms := f.Space().MarginalS(f.Alpha())
	for s := 0; s < numStates; s++ {
		if d := math.Abs(ms[s] - m.Prior[s]); d >= tol {
			t.Fatalf("numBeds=0 退化 MarginalS[%v]=%.6f 期望 Prior %.6f", State(s), ms[s], m.Prior[s])
		}
	}
}

// T3 maxBeds 超界硬断言（P-5）。
func TestT3MaxBedsPanic(t *testing.T) {
	for _, nb := range []int{-1, 4, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("numBeds=%d 应 panic（P-5）", nb)
				}
			}()
			NewJointSpace(nb)
		}()
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("numBeds=maxBeds=%d 不应 panic，got %v", maxBeds, r)
			}
		}()
		NewJointSpace(maxBeds)
	}()
}

// T4 D-2 核心：ε≪λ 复现 30s staleness（cd2b 治本不变量）。
func TestT4StalenessUnderEpsilonLLLambda(t *testing.T) {
	m := DefaultModel()

	// 离线组：初始 occ 主导 (SBed,bmask=1)，sleepad 离线，看 P(occ) 跌破 0.5 的步数。
	f := NewFilter(m, 1)
	js := f.Space()
	a := js.NewJointVector()
	a[js.idx(SBed, 1)] = 1.0 // 人在床睡 + 床 occ
	a.normalize()
	f.alpha = a
	if p0 := js.MarginalB(f.Alpha(), 0); p0 < 0.99 {
		t.Fatalf("初始 P(occ)=%.4f 应 ≥0.99", p0)
	}
	offline := []bool{false}
	crossStep := -1
	for step := 1; step <= 120; step++ {
		f.Predict(offline)
		if js.MarginalB(f.Alpha(), 0) < 0.5 {
			crossStep = step
			break
		}
	}
	if crossStep <= 0 || crossStep > 30 {
		t.Fatalf("离线 crossStep=%d 应 ∈(0,30]（陈旧 occ 在 30s staleness 窗内蒸发）", crossStep)
	}

	// 在线对照：同初始，sleepad 在线，30 步后 P(occ) 仍 ≥0.5（K^obs 高自持，证蒸发归因离线核非数值漂移）。
	g := NewFilter(m, 1)
	b := js.NewJointVector()
	b[js.idx(SBed, 1)] = 1.0
	b.normalize()
	g.alpha = b
	online := []bool{true}
	for step := 0; step < 30; step++ {
		g.Predict(online)
	}
	if p := g.Space().MarginalB(g.Alpha(), 0); p < 0.5 {
		t.Fatalf("在线 30 步后 P(occ)=%.4f 应 ≥0.5（ε≪λ：在线核不蒸发）", p)
	}
}

// T5 §C vac 吸收态：空房离线不被弛豫成伪占用（箭头记法方向守门）。
func TestT5VacAbsorbing(t *testing.T) {
	m := DefaultModel()
	f := NewFilter(m, 1)
	js := f.Space()
	a := js.NewJointVector()
	a[js.idx(SEmpty, 0)] = 1.0 // 空房 + 床 vac
	a.normalize()
	f.alpha = a
	offline := []bool{false}
	for step := 0; step < 120; step++ {
		f.Predict(offline)
	}
	if p := js.MarginalB(f.Alpha(), 0); p > 0.05 {
		t.Fatalf("空房离线 120 步后 P(occ)=%.4f 应 ≤0.05（vac 吸收，§C 方向）", p)
	}
}
