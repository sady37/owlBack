package belief

import (
	"fmt"
	"math"
)

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

// BedOnline 第 j 床 sleepad 在线标志（ρ_t）。长 = numBeds。
type BedOnline []bool

// Predict 联合时间转移（因子化 T_S ⊗ T_B，log 域）。
// online[j] = 第 j 床 ρ_t（决定 T_B 用 K^obs/K^unobs）。
// rhoXroom = neighbor ρ_xroom（§A.2，W3.1）：>0 仅在 lost-track 激活，把 Blind from-行的 →Fallen 整流入
//   →Left（人挪去邻房非本房真摔）= **转移先验**，非 gate。rhoXroom≤0 → 用静态 logA（逐 tick 等价，零回归 oracle）。
// B1 契约：len(online) 须 == numBeds（床在线标志逐床显式）；不符 = 上游 wiring 错 → panic（规则 1.4）。
func (f *Filter) Predict(online BedOnline, rhoXroom float64) {
	js := f.space
	nb := js.numBeds
	if len(online) != nb {
		panic(fmt.Sprintf("belief: Predict online 长度=%d ≠ numBeds=%d（B1 契约）", len(online), nb))
	}

	logTB := f.buildLogTBCol(online) // bmaskN×bmaskN 因子化 log T_B 表（提出 S 循环外）

	// neighbor 整流：仅 Blind from-行（SBlindRest/SBlindOpen）按 ρ 改向 F→L，其余行不变。
	logA := &f.logA
	var gated [numStates][numStates]float64
	if rhoXroom > 0 {
		gated = f.logA
		for _, sFrom := range [...]State{SBlindRest, SBlindOpen} {
			g := GateBlindRow(f.model.A[sFrom], rhoXroom) // prob 行 F→L 整流（行和守恒）
			for sTo := 0; sTo < numStates; sTo++ {
				gated[sFrom][sTo] = logP(g[sTo])
			}
		}
		logA = &gated
	}

	next := js.NewJointVector()
	// ᾱ(sTo, bTo) = LogSumExp_{sFrom, bFrom} ( logA[sFrom][sTo] + logTB[bFrom][bTo] + α )
	for sTo := 0; sTo < numStates; sTo++ {
		for bTo := 0; bTo < js.bmaskN; bTo++ {
			acc := math.Inf(-1)
			for sFrom := 0; sFrom < numStates; sFrom++ {
				la := logA[sFrom][sTo]
				if math.IsInf(la, -1) {
					continue
				}
				for bFrom := 0; bFrom < js.bmaskN; bFrom++ {
					lb := logTB[bFrom][bTo]
					if math.IsInf(lb, -1) {
						continue
					}
					ap := f.alpha[js.idx(State(sFrom), bFrom)]
					if math.IsInf(ap, -1) {
						continue
					}
					acc = logAdd(acc, la+lb+ap)
				}
			}
			next[js.idx(State(sTo), bTo)] = acc
		}
	}
	next.LogNormalize()
	f.alpha = next
}

// buildLogTBCol 预算 bmaskN×bmaskN 的因子化 log T_B 表：
// logTB[bFrom][bTo] = Σ_j log T_B^j(bit_j(bTo) | bit_j(bFrom))，按各床 online[j] 选 K^obs/K^unobs。
// 仅依赖 (bFrom,bTo)、不依赖 S → 提出 S 循环外（T_B 因子化是 Predict 第一步，B 评审建议）。
// -inf（如离线 vac→occ=0）经 float 加法自然传播。
func (f *Filter) buildLogTBCol(online BedOnline) [][]float64 {
	js := f.space
	nb := js.numBeds
	logK := make([]logKernel, nb)
	for j := 0; j < nb; j++ {
		logK[j] = makeLogKernel(f.bedP.tBKernel(online[j]))
	}
	logTB := make([][]float64, js.bmaskN)
	for bFrom := 0; bFrom < js.bmaskN; bFrom++ {
		logTB[bFrom] = make([]float64, js.bmaskN)
		for bTo := 0; bTo < js.bmaskN; bTo++ {
			s := 0.0
			for j := 0; j < nb; j++ {
				s += logK[j][bedOf(bFrom, j)][bedOf(bTo, j)]
			}
			logTB[bFrom][bTo] = s
		}
	}
	return logTB
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

// foldRealness realness 折进 logPhi（C §42：SFallen 发射 ×P(real)，走同一 Correct 路径 = 真内化，
//   非软 gate）。pFallReal = 本房 fall 证据来自真人的后验 ∈[0,1]；≥1 中性（假设真人 → 零回归 oracle）；
//   →0 = ghost 的「摔」喂不动 SFallen（真人摔倒 P(real) 高则不被抑制 = 共生律 [[bed_fusion_authority_model]]）。
func (f *Filter) foldRealness(logPhi JointVector, pFallReal float64) JointVector {
	if pFallReal >= 1.0 {
		return logPhi // 中性：原样返回（零回归 oracle 的一半）
	}
	js := f.space
	out := js.NewJointVector()
	if logPhi != nil {
		copy(out, logPhi)
	}
	lr := logP(pFallReal)
	for b := 0; b < js.bmaskN; b++ {
		out[js.idx(SFallen, b)] += lr
	}
	return out
}

// Step 一帧推进。dtMs≤0（同帧重入）跳过 Predict。
// rhoXroom（neighbor，进 Predict）/ pFallReal（realness，折 logPhi）；中性值 (0, 1) → 逐 tick 等价 S/B-only。
func (f *Filter) Step(nowMs int64, online BedOnline, logPsi, logPhi JointVector, rhoXroom, pFallReal float64) {
	if f.lastTs > 0 && nowMs > f.lastTs {
		f.Predict(online, rhoXroom)
	}
	f.Correct(logPsi, f.foldRealness(logPhi, pFallReal))
	if nowMs > f.lastTs {
		f.lastTs = nowMs
	}
}

// VerifySum Σ exp(α) — 应 = 1.0（数值容差内）。阶段 1 验收 T1 用。
func (f *Filter) VerifySum() float64 {
	return math.Exp(LogSumExp(f.alpha))
}
