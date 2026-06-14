package belief

import (
	"math"
	"testing"
)

// joint_test.go — 阶段 1 骨架验收（与 δ/neighbor 零依赖）：
//   T1 Σα=1（归一化守恒，每步后）
//   T2 单床退化：numBeds=1 → 18 态（9·2）；numBeds=0 → 9 态（回 S 轴单链）
//   T3 P-5 maxBeds 超界 panic
//   T4 D-2：ε≪λ 复现 30s staleness（陈旧 occ 离线后在 ~30s 内蒸发到 vac 主导）

const eps = 1e-9

func sumAlpha(v JointVector) float64 {
	s := 0.0
	for _, p := range v {
		s += p
	}
	return s
}

// T1：归一化守恒——初始 + 多步 Predict 后 Σα 恒 =1。
func TestJointNormalization(t *testing.T) {
	for _, nb := range []int{0, 1, 2, 3} {
		f := NewFilter(DefaultModel(), nb)
		if got := sumAlpha(f.Alpha()); math.Abs(got-1) > eps {
			t.Fatalf("nb=%d 初始 Σα=%.12f ≠ 1", nb, got)
		}
		online := make(bedOnline, nb)
		for j := range online {
			online[j] = true
		}
		for step := 0; step < 50; step++ {
			f.Predict(online)
			if got := sumAlpha(f.Alpha()); math.Abs(got-1) > eps {
				t.Fatalf("nb=%d step=%d Σα=%.12f ≠ 1", nb, step, got)
			}
		}
	}
}

// T2：单床退化基数。numBeds=1 → 18；=0 → 9；=2 → 36；=3 → 72。
func TestJointDegeneracy(t *testing.T) {
	cases := map[int]int{0: 9, 1: 18, 2: 36, 3: 72}
	for nb, want := range cases {
		f := NewFilter(DefaultModel(), nb)
		if got := f.Space().Size(); got != want {
			t.Errorf("numBeds=%d: size=%d, want %d", nb, got, want)
		}
		if f.NumBeds() != nb {
			t.Errorf("numBeds 字段=%d, want %d (B1 显式持有)", f.NumBeds(), nb)
		}
	}
	// numBeds=0：联合空间 == S 轴单链，MarginalS 应等于原 Belief 行为。
	f0 := NewFilter(DefaultModel(), 0)
	ms := f0.Space().MarginalS(f0.Alpha())
	if math.Abs(ms[SEmpty]-DefaultModel().Prior[SEmpty]) > eps {
		t.Errorf("nb=0 MarginalS(Empty)=%.6f ≠ Prior(Empty)=%.6f", ms[SEmpty], DefaultModel().Prior[SEmpty])
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
	// 边界 maxBeds 本身合法（不 panic）。
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("numBeds=maxBeds=%d 不应 panic", maxBeds)
			}
		}()
		NewJointSpace(maxBeds)
	}()
}

// T4（D-2 核心验收）：ε≪λ 复现 30s staleness。
// 场景：单床，初始 occ 主导（人在床）；t=0 起 sleepad 离线（ρ=0），无任何观测。
// 期望：陈旧 occ 经 K^unobs_λ 单向泄漏，在 ~30s（1Hz → 30 步）内 P(B=occ) 衰减到 <0.5（vac 主导），
//       复现原 30s staleness 行为——这是“ε≪λ 替代 staleness/TTL”待验不变量的显式验证（feedback-p6C D-2）。
func TestStaleness30sReproduction(t *testing.T) {
	f := NewFilter(DefaultModel(), 1)
	js := f.Space()

	// 构造初始 occ 主导：把质量集中到 (任意 S, B=occ)。用 SBed×occ 作“人在床睡”。
	for i := range f.alpha {
		f.alpha[i] = 0
	}
	f.alpha[js.idx(SBed, 1)] = 1.0 // bmask=1 → 床 occ
	f.alpha.normalize()

	if p0 := js.MarginalB(f.alpha, 0); p0 < 0.99 {
		t.Fatalf("初始 P(occ)=%.4f 应 ≈1", p0)
	}

	// sleepad 离线：ρ=false。逐步 Predict（1Hz），记录 P(occ) 跌破 0.5 的步数。
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
		t.Fatalf("D-2 失败：离线 120s 内陈旧 occ 从未跌破 0.5（λ 太小，ε≪λ 不成立 → staleness 不复现）")
	}
	// 复现 30s staleness：跌破点应在 30s 量级（半衰期 ≈ ln2/λ）。默认 λ=0.05 → 半衰期≈14s，
	// occ 主导跌破 0.5 约 1 个半衰期 ≈ 14步，应 < 30 步（在原 staleness 窗内蒸发）。
	if crossStep > 30 {
		t.Errorf("D-2：陈旧 occ 跌破 0.5 用了 %d 步 > 30s 窗——λ 偏小，未在 staleness 窗内蒸发", crossStep)
	}
	t.Logf("D-2 ✓ 离线后陈旧 occ 在 %d 步(≈%ds)蒸发到 vac 主导（30s staleness 复现；半衰期≈%.0fs）",
		crossStep, crossStep, math.Ln2/0.05)

	// 对照：若在线（ρ=true），occ 应保持主导（高自持，不蒸发）——证明蒸发是离线核 K^unobs 所致。
	f2 := NewFilter(DefaultModel(), 1)
	for i := range f2.alpha {
		f2.alpha[i] = 0
	}
	f2.alpha[js.idx(SBed, 1)] = 1.0
	f2.alpha.normalize()
	online := bedOnline{true}
	for step := 1; step <= 30; step++ {
		f2.Predict(online)
	}
	if pOcc := js.MarginalB(f2.alpha, 0); pOcc < 0.5 {
		t.Errorf("对照失败：在线 30s 后 P(occ)=%.4f < 0.5——在线核不该让 occ 蒸发（ε 太大）", pOcc)
	}
}

// T5：vac 吸收态验证（§C 单向泄漏的核心）——空房离线，vac 不被弛豫成伪占用。
func TestVacAbsorbing(t *testing.T) {
	f := NewFilter(DefaultModel(), 1)
	js := f.Space()
	for i := range f.alpha {
		f.alpha[i] = 0
	}
	f.alpha[js.idx(SEmpty, 0)] = 1.0 // 空房 + 床 vac
	f.alpha.normalize()

	offline := bedOnline{false}
	for step := 1; step <= 120; step++ {
		f.Predict(offline)
	}
	// vac 吸收：离线 120s 后 P(occ) 仍应 ≈0（vac→occ=0），不被弛豫成 0.5 伪占用。
	if pOcc := js.MarginalB(f.alpha, 0); pOcc > 0.05 {
		t.Errorf("vac 吸收失败：空房离线 120s 后 P(occ)=%.4f 应 ≈0（§C vac→occ=0 单向泄漏）", pOcc)
	}
}
