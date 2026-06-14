package belief

// filter.go — 联合 forward 滤波（DBN-Zone-Room §7）。
//   Predict: ᾱ(S,{B^j}) = Σ T_S(S|S')·∏_j T_B(B^j|B'^j)·α_{t-1}
//   Correct: α ∝ Ψ·Φ·ᾱ，Σ=1
// §6 纪律：转移纯因子化，T_S 不被 B' 调制（耦合只在 Ψ，防双施）。
// 阶段 1 骨架：Ψ/Φ 中性占位（nil）→ Correct 恒等。先验数值正确性（T1–T5），阶段 2 接 emission/coupling。

// Filter 联合滤波器。numBeds 随 space 显式持有（B1）。
type Filter struct {
	space  *JointSpace
	model  *Model
	bedP   bedAxisParams
	alpha  JointVector
	lastTs int64
}

// NewFilter 建滤波器；numBeds 经 NewJointSpace 做 P-5 超界断言。
// 初始 α = Prior(S) ⊗ [全 vac]（空房默认，两轴独立外积）。
func NewFilter(model *Model, numBeds int) *Filter {
	space := NewJointSpace(numBeds)
	f := &Filter{
		space: space,
		model: model,
		bedP:  defaultBedAxisParams(),
		alpha: space.NewJointVector(),
	}
	f.initAlpha()
	return f
}

func (f *Filter) initAlpha() {
	js := f.space
	const bmaskAllVac = 0
	for s := 0; s < numStates; s++ {
		f.alpha[js.idx(State(s), bmaskAllVac)] = f.model.Prior[s]
	}
	f.alpha.normalize()
}

func (f *Filter) Alpha() JointVector { return f.alpha }
func (f *Filter) Space() *JointSpace { return f.space }
func (f *Filter) NumBeds() int       { return f.space.numBeds }

// Predict 因子化时间转移。online[j]=第 j 床 ρ_t（决定 T_B 用 K^obs/K^unobs）；越界/缺省视为离线。
// T_S 对所有 bmask 同一阵（不被 B' 调制，§6）；T_B 对所有 S 同一核（纯因子化）。
func (f *Filter) Predict(online []bool) {
	js := f.space
	nb := js.numBeds

	kernels := make([]bedKernel, nb)
	for j := 0; j < nb; j++ {
		on := false
		if j < len(online) {
			on = online[j]
		}
		kernels[j] = f.bedP.tBKernel(on)
	}

	next := js.NewJointVector()
	for iFrom := 0; iFrom < js.size; iFrom++ {
		p := f.alpha[iFrom]
		if p == 0 {
			continue
		}
		sFrom, bmaskFrom := js.decode(iFrom)
		for sTo := 0; sTo < numStates; sTo++ {
			aS := f.model.A[sFrom][sTo]
			if aS == 0 {
				continue
			}
			for bmaskTo := 0; bmaskTo < js.bmaskN; bmaskTo++ {
				bw := 1.0
				for j := 0; j < nb; j++ {
					bw *= kernels[j][bedOf(bmaskFrom, j)][bedOf(bmaskTo, j)]
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

// Correct 观测更新 α ∝ Ψ·Φ·ᾱ。阶段 1：psi/phi 传 nil（中性）→ 恒等（只归一化）。
func (f *Filter) Correct(psi, phi JointVector) {
	for i := 0; i < f.space.size; i++ {
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

// Step 一帧：Predict（按 online）后 Correct。阶段 1 传 nil/nil = 中性。
func (f *Filter) Step(nowMs int64, online []bool, psi, phi JointVector) {
	if f.lastTs > 0 && nowMs > f.lastTs {
		f.Predict(online)
	}
	f.Correct(psi, phi)
	if nowMs > f.lastTs {
		f.lastTs = nowMs
	}
}
