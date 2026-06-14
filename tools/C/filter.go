package belief

// filter.go — 联合 forward 滤波（DBN-Zone-Room §7）。
//   Predict: ᾱ_t(S,{B^j}) = Σ T_S(S|S')·∏_j T_B(B^j|B'^j)·α_{t-1}
//   Correct: α_t ∝ Ψ_t · Φ_t · ᾱ_t   （Σ=1）
//
// §6 纪律：转移纯因子化，T_S **不被 B' 调制**（耦合只在 Ψ，防双施）。
// 阶段 1（骨架）：Ψ/Φ 中性占位（=1），先验证滤波器数值正确性——
//   Σα=1、单床退化回 18 态、ε≪λ 复现 30s staleness（D-2）。
// 阶段 2 再接 emission.go(Φ) / coupling.go(Ψ)。

// Filter 联合滤波器。numBeds 进字段（B1 闭合：床数随 Filter 显式持有，不靠隐式推断）。
type Filter struct {
	space   *JointSpace
	model   *Model        // S 轴转移 T_S（复用 state.go/model.go，不改）
	bedP    bedAxisParams // B 轴 T_B 参数
	alpha   JointVector   // 当前联合后验 α_t
	lastTs  int64
}

// NewFilter 建滤波器。numBeds 经 NewJointSpace 做 P-5 超界断言。
// 初始分布：S 轴用 model.Prior，B 轴用全 vac（空房默认），两轴独立外积。
func NewFilter(model *Model, numBeds int) *Filter {
	space := NewJointSpace(numBeds) // P-5：numBeds 超界在此 panic
	bedP := defaultBedAxisParams()
	f := &Filter{
		space:  space,
		model:  model,
		bedP:   bedP,
		alpha:  space.NewJointVector(),
		lastTs: 0,
	}
	f.initAlpha()
	return f
}

// initAlpha 初始 α = Prior(S) ⊗ [全 vac]（B 轴初始全空）。
func (f *Filter) initAlpha() {
	js := f.space
	bmaskAllVac := 0 // 所有床 vac
	for s := 0; s < numStates; s++ {
		f.alpha[js.idx(State(s), bmaskAllVac)] = f.model.Prior[s]
	}
	f.alpha.normalize()
}

// Alpha 当前联合后验快照。
func (f *Filter) Alpha() JointVector { return f.alpha }

// Space 联合空间元信息。
func (f *Filter) Space() *JointSpace { return f.space }

// NumBeds 本房床数（B1：显式持有）。
func (f *Filter) NumBeds() int { return f.space.numBeds }

// bedOnline 第 j 床 sleepad 在线标志（ρ_t）。阶段 1 骨架：调用方传入；
// 阶段 2 由 adapter 按 sleepad fresh 填。这里用切片，长度 = numBeds。
type bedOnline []bool

// Predict 联合时间转移（因子化）。online[j] = 第 j 床 ρ_t（决定 T_B 用 K^obs/K^unobs）。
// T_S 对所有 bmask 同一阵（不被 B' 调制，§6）；T_B 对所有 S 同一核（纯因子化）。
func (f *Filter) Predict(online bedOnline) {
	js := f.space
	nb := js.numBeds
	A := f.model.A

	// 预算每床的 2×2 T_B 核（按各自 ρ）。
	kernels := make([]bedKernel, nb)
	for j := 0; j < nb; j++ {
		on := false
		if j < len(online) {
			on = online[j]
		}
		kernels[j] = f.bedP.tBKernel(on)
	}

	next := js.NewJointVector()
	// ᾱ(S,bmask) = Σ_{S',bmask'} A[S'][S] · ∏_j K_j[bit_j(bmask')][bit_j(bmask)] · α(S',bmask')
	for iFrom := 0; iFrom < js.size; iFrom++ {
		p := f.alpha[iFrom]
		if p == 0 {
			continue
		}
		sFrom, bmaskFrom := js.decode(iFrom)
		for sTo := 0; sTo < numStates; sTo++ {
			aS := A[sFrom][sTo]
			if aS == 0 {
				continue
			}
			// B 轴：对每个目标 bmask 累乘各床核。bmaskN 小（≤8），全枚举。
			for bmaskTo := 0; bmaskTo < js.bmaskN; bmaskTo++ {
				bw := 1.0
				for j := 0; j < nb; j++ {
					from := bedOf(bmaskFrom, j)
					to := bedOf(bmaskTo, j)
					bw *= kernels[j][from][to]
					if bw == 0 {
						break
					}
				}
				if bw == 0 {
					continue
				}
				next[js.idx(State(sTo), bmaskTo)] += aS * bw * p
			}
		}
	}
	next.normalize()
	f.alpha = next
}

// Correct 观测更新 α ∝ Ψ·Φ·ᾱ。
// 阶段 1（骨架）：Ψ/Φ 中性（=1）→ Correct 是恒等（只归一化）。占位，阶段 2 接 emission/coupling。
// psi/phi 为每个 joint index 的相容势/发射似然向量（中性=全 1）。
func (f *Filter) Correct(psi, phi JointVector) {
	js := f.space
	for i := 0; i < js.size; i++ {
		w := 1.0
		if psi != nil {
			w *= psi[i]
		}
		if phi != nil {
			w *= phi[i]
		}
		f.alpha[i] *= w
	}
	f.alpha.normalize()
}

// Step 一帧推进：Predict（按 online ρ）后 Correct（阶段 1 传 nil/nil = 中性）。
// dtMs 用于多步推进判定（缺证据长 gap 仍每帧 1 步，与 Tsensor belief 一致）。
func (f *Filter) Step(nowMs int64, online bedOnline, psi, phi JointVector) {
	if f.lastTs > 0 && nowMs > f.lastTs {
		f.Predict(online)
	}
	f.Correct(psi, phi)
	if nowMs > f.lastTs {
		f.lastTs = nowMs
	}
}

// neutralVector 中性占位向量（全 1，长度 = size）。阶段 1 Ψ/Φ 用。
func (f *Filter) neutralVector() JointVector {
	v := make(JointVector, f.space.size)
	for i := range v {
		v[i] = 1.0
	}
	return v
}
