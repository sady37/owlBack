package belief

import "math"

// filter.go — 联合 forward 滤波（DBN-Zone-Room §7，log 域）。
//   Predict: ᾱ_t = log Σ_{S',{B'^j}} exp( log T_S(S|S') + Σ_j log T_B(B^j|B'^j) + α_{t-1} )
//   Correct: α_t ∝ log Ψ_t + log Φ_t + ᾱ_t     （log 域元素加，再 LogNormalize）
//
// §6：转移纯因子化，T_S 不被 B' 调制（耦合只在 Ψ，防双施）。
// 阶段 1（骨架）：Ψ/Φ 中性占位（log=0），VerifySum() 验 Σα=1。

// Filter 联合滤波器。numBeds 进字段（B1 闭合：床数随 Filter 显式持有）。
type Filter struct {
	space   *JointSpace
	model   *Model        // S 轴转移 T_S（复用 state.go/model.go）
	bedP    bedAxisParams // B 轴 T_B 参数
	logA    [numStates][numStates]float64 // 预存 log T_S
	alpha   JointVector   // 当前联合后验（log 域）
	lastTs  int64
}

// NewFilter 建滤波器。numBeds 经 NewJointSpace 做 P-5 超界断言。
// 初始分布：S 轴用 model.Prior，B 轴均匀。
func NewFilter(model *Model, numBeds int) *Filter {
	space := NewJointSpace(numBeds) // P-5：numBeds 超界在此 panic
	bedP := defaultBedAxisParams()
	f := &Filter{
		space:  space,
		model:  model,
		bedP:   bedP,
		alpha:  space.initFromPrior(model.Prior),
		lastTs: 0,
	}
	// 预存 log T_S
	for i := 0; i < numStates; i++ {
		for j := 0; j < numStates; j++ {
			f.logA[i][j] = logP(model.A[i][j])
		}
	}
	return f
}

// Alpha 当前联合后验（log 域）。
func (f *Filter) Alpha() JointVector { return f.alpha }

// Space 联合空间元信息。
func (f *Filter) Space() *JointSpace { return f.space }

// NumBeds 本房床数（B1：显式持有）。
func (f *Filter) NumBeds() int { return f.space.numBeds }

// bedOnline 第 j 床 sleepad 在线标志（ρ_t）。长 = numBeds。
type bedOnline []bool

// Predict 联合时间转移（因子化 T_S ⊗ T_B，log 域）。
// online[j] = 第 j 床 ρ_t（决定 T_B 用 K^obs/K^unobs）。
func (f *Filter) Predict(online bedOnline) {
	js := f.space
	nb := js.numBeds

	// 预算每床的 log T_B 核（按各自 ρ）。
	logK := make([]logKernel, nb)
	for j := 0; j < nb; j++ {
		on := false
		if j < len(online) {
			on = online[j]
		}
		logK[j] = makeLogKernel(f.bedP.tBKernel(on))
	}

	next := js.NewJointVector()
	// ᾱ(sTo, bTo) = LogSumExp_{sFrom, bFrom} ( logA[sFrom][sTo] + Σ_j logK_j + α )
	for sTo := 0; sTo < numStates; sTo++ {
		for bTo := 0; bTo < js.bmaskN; bTo++ {
			acc := math.Inf(-1)
			for sFrom := 0; sFrom < numStates; sFrom++ {
				la := f.logA[sFrom][sTo]
				if math.IsInf(la, -1) {
					continue
				}
				for bFrom := 0; bFrom < js.bmaskN; bFrom++ {
					if f.alpha[js.idx(State(sFrom), bFrom)] == math.Inf(-1) {
						continue
					}
					// Σ_j log T_B^j(B^j_to | B^j_from)
					logB := 0.0
					zero := false
					for j := 0; j < nb; j++ {
						from := bedOf(bFrom, j)
						to := bedOf(bTo, j)
						lb := logK[j][from][to]
						if math.IsInf(lb, -1) {
							zero = true
							break
						}
						logB += lb
					}
					if zero {
						continue
					}
					v := la + logB + f.alpha[js.idx(State(sFrom), bFrom)]
					if !math.IsInf(v, -1) {
						acc = logAdd(acc, v)
					}
				}
			}
			next[js.idx(State(sTo), bTo)] = acc
		}
	}
	next.LogNormalize()
	f.alpha = next
}

// Correct 观测更新 α ∝ Ψ · Φ · ᾱ（log 域元素加）。
// 阶段 1（骨架）：psi/phi 传 nil → 中性恒等（只 LogNormalize）。
func (f *Filter) Correct(logPsi, logPhi JointVector) {
	js := f.space
	for i := 0; i < js.size; i++ {
		if logPsi != nil {
			f.alpha[i] += logPsi[i]
		}
		if logPhi != nil {
			f.alpha[i] += logPhi[i]
		}
	}
	f.alpha.LogNormalize()
}

// Step 一帧推进。dtMs≤0（同帧重入）跳过 Predict。
func (f *Filter) Step(nowMs int64, online bedOnline, logPsi, logPhi JointVector) {
	if f.lastTs > 0 && nowMs > f.lastTs {
		f.Predict(online)
	}
	f.Correct(logPsi, logPhi)
	if nowMs > f.lastTs {
		f.lastTs = nowMs
	}
}

// VerifySum Σ exp(α) — 应 = 1.0（数值容差内）。阶段 1 验收 T1 用。
func (f *Filter) VerifySum() float64 {
	return math.Exp(LogSumExp(f.alpha))
}
