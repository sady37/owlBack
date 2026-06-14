package belief

import (
	"math"
	"testing"
)

// joint_test.go — 阶段 1 骨架验收（与 δ/neighbor 零依赖）：
//   T1 Σα=1（归一化守恒，每步后）
//   T2 单床退化：numBeds=1→18 态；numBeds=0→9 态，MarginalS = model.Prior
//   T3 P-5 maxBeds 超界 panic
//   T4 D-2：ε≪λ 复现 30s staleness（陈旧 occ 离线在 ~30s 内蒸发到 vac 主导）
//   T5 vac 吸收态（§C 单向泄漏核心——空房离线不被弛豫成伪占用）

const eps = 1e-9

func verifySum(t *testing.T, f *Filter, label string) {
	t.Helper()
	if s := f.VerifySum(); math.Abs(s-1.0) > eps {
		t.Fatalf("%s Σα=%.12f ≠ 1", label, s)
	}
}

// T1：归一化守恒——初始 + 多步 Predict 后 Σα 恒 =1。
func TestJointNormalization(t *testing.T) {
	for _, nb := range []int{0, 1, 2, 3} {
		f := NewFilter(DefaultModel(), nb)
		verifySum(t, f, "nb="+itoa(nb)+" init")

		online := make(bedOnline, nb)
		for j := range online {
			online[j] = true
		}
		for step := 0; step < 50; step++ {
			f.Predict(online)
			verifySum(t, f, "nb="+itoa(nb)+" step="+itoa(step))
		}
	}
}

// T2：单床退化基数 + B1 显式持有。
func TestJointDegeneracy(t *testing.T) {
	cases := map[int]int{0: 9, 1: 18, 2: 36, 3: 72}
	for nb, want := range cases {
		f := NewFilter(DefaultModel(), nb)
		if got := f.Space().Size(); got != want {
			t.Errorf("numBeds=%d: size=%d, want %d", nb, got, want)
		}
		if f.NumBeds() != nb {
			t.Errorf("numBeds=%d: NumBeds()=%d (B1 显式持有)", nb, f.NumBeds())
		}
	}
	// numBeds=0：MarginalS 应等于 model.Prior
	f0 := NewFilter(DefaultModel(), 0)
	ms := f0.Space().MarginalS(f0.Alpha())
	for s := 0; s < numStates; s++ {
		if math.Abs(ms[s]-f0.model.Prior[s]) > eps {
			t.Errorf("nb=0 MarginalS[%d]=%.6f ≠ Prior[%d]=%.6f", s, ms[s], s, f0.model.Prior[s])
		}
	}
}

// T3：P-5 超界 panic。
func TestMaxBedsAssertion(t *testing.T) {
	for _, nb := range []int{-1, maxBeds + 1, 99} {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("numBeds=%d 应 panic（P-5 硬 bound），未 panic", nb)
				}
			}()
			NewJointSpace(nb)
		}()
	}
	// 边界 maxBeds 合法
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("numBeds=maxBeds=%d 不应 panic", maxBeds)
			}
		}()
		NewJointSpace(maxBeds)
	}()
}

// T4：D-2 核心验收——ε≪λ 复现 30s staleness。
// 场景：单床，初始 occ 主导；t=0 起 sleepad 离线。陈旧 occ 经 K^unobs_λ 单向泄漏，
// 在 ~30s 内 P(occ) 衰减到 <0.5（vac 主导）。
func TestStaleness30sReproduction(t *testing.T) {
	f := NewFilter(DefaultModel(), 1)
	js := f.Space()

	// 构造初始 occ 主导：(SBed, bmask=1) = 1.0
	for i := range f.alpha {
		f.alpha[i] = math.Inf(-1)
	}
	f.alpha[js.idx(SBed, 1)] = 0 // log(1)
	f.alpha.LogNormalize()

	if p0 := js.MarginalB(f.alpha, 0); p0 < 0.99 {
		t.Fatalf("初始 P(occ)=%.4f 应 ≈1", p0)
	}

	// sleepad 离线。逐步 Predict，记录 P(occ) 跌破 0.5 的步数。
	offline := bedOnline{false}
	crossStep := -1
	for step := 1; step <= 120; step++ {
		f.Predict(offline)
		pOcc := js.MarginalB(f.alpha, 0)
		if crossStep < 0 && pOcc < 0.5 {
			crossStep = step
		}
	}

	if crossStep < 0 {
		t.Fatalf("D-2 失败：离线 120s 内陈旧 occ 从未跌破 0.5（λ 太小）")
	}
	if crossStep > 30 {
		t.Errorf("D-2：陈旧 occ 跌破 0.5 用了 %d 步 > 30s 窗（λ 偏小，staleness 未复现）", crossStep)
	}
	t.Logf("D-2 ✓ 离线后陈旧 occ 在 %d 步(≈%ds)蒸发到 vac 主导（半衰期≈%.0fs）",
		crossStep, crossStep, math.Ln2/0.05)

	// 对照组：在线应保持 occ 主导（蒸发归因离线核，非数值漂移）
	f2 := NewFilter(DefaultModel(), 1)
	for i := range f2.alpha {
		f2.alpha[i] = math.Inf(-1)
	}
	f2.alpha[js.idx(SBed, 1)] = 0
	f2.alpha.LogNormalize()
	online := bedOnline{true}
	for step := 1; step <= 30; step++ {
		f2.Predict(online)
	}
	if pOcc := js.MarginalB(f2.alpha, 0); pOcc < 0.5 {
		t.Errorf("对照失败：在线 30s 后 P(occ)=%.4f < 0.5（在线核不该蒸发 occ）", pOcc)
	}
}

// T5：vac 吸收态验证（§C 单向泄漏核心——空房离线不被弛豫成伪占用）。
func TestVacAbsorbing(t *testing.T) {
	f := NewFilter(DefaultModel(), 1)
	js := f.Space()
	for i := range f.alpha {
		f.alpha[i] = math.Inf(-1)
	}
	f.alpha[js.idx(SEmpty, 0)] = 0 // log(1)：空房 + bed vac
	f.alpha.LogNormalize()

	offline := bedOnline{false}
	for step := 1; step <= 120; step++ {
		f.Predict(offline)
	}
	if pOcc := js.MarginalB(f.alpha, 0); pOcc > 0.05 {
		t.Errorf("vac 吸收失败：空房离线 120s 后 P(occ)=%.4f 应 ≈0（§C vac→occ=0）", pOcc)
	}
}

func itoa(n int) string {
	return string(rune('0'+n)) // works for 0-9 only, sufficient for numBeds range
}
