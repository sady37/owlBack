package belief

import (
	"math"
)

// joint.go — 联合隐状态 J=(S,{B^j})（DBN-Zone-Room §1, §7）。
// S 轴（9 态全空间，state.go）× B 轴（每床 vac/occ，bed_axis.go）的笛卡尔积。
// 基数 = numStates · 2^numBeds。索引：idx = s·(2^numBeds) + bmask，bmask 第 j 位 = B^j。
//
// P-5（状态爆炸硬 bound）：养老院房 ≤4 床 + 折进 Beds 轴的 LongSofa，MaxBeds=6 → 9·64=576 态，仍轻量。
// §7 数值方案：全程 log 域 + log-sum-exp 归一化（ε_art 极小 / L_in 大倍数 → 线性域不安全）。

// MaxBeds P-5 硬上界。超界不再 panic：NewJointSpace clamp 到本值（房降级跑，绝不拖垮全进程）；
// 上游 bootstrap 另发 warn 日志带 roomID。提高本值是指数代价（态数 = 9·2^MaxBeds），勿轻易上调。
const MaxBeds = 6

// JointSpace 联合状态空间元信息。numBeds 进字段（B1 闭合：床数显式持有，不靠隐式推断）。
type JointSpace struct {
	numBeds int // 本房床数 ∈ [0, MaxBeds]
	bmaskN  int // 2^numBeds，B 轴配置数
	size    int // numStates · bmaskN，联合空间基数
}

// NewJointSpace 建联合空间。numBeds 超界自动 clamp 到 [0, MaxBeds]（P-5：降级不 panic，单点安全网）。
func NewJointSpace(numBeds int) *JointSpace {
	if numBeds < 0 {
		numBeds = 0
	}
	if numBeds > MaxBeds {
		numBeds = MaxBeds
	}
	bmaskN := 1 << numBeds
	return &JointSpace{
		numBeds: numBeds,
		bmaskN:  bmaskN,
		size:    numStates * bmaskN,
	}
}

func (js *JointSpace) NumBeds() int { return js.numBeds }
func (js *JointSpace) Size() int    { return js.size }

// idx 联合索引：(S, bmask) → 扁平 index。bmask ∈ [0, 2^numBeds)。
func (js *JointSpace) idx(s State, bmask int) int {
	return int(s)*js.bmaskN + bmask
}

// decode 扁平 index → (S, bmask)。
func (js *JointSpace) decode(i int) (State, int) {
	return State(i / js.bmaskN), i % js.bmaskN
}

// bedOf 取 bmask 第 j 位 = B^j（0=vac, 1=occ）。
func bedOf(bmask, j int) BedState {
	return BedState((bmask >> j) & 1)
}

// JointVector 联合后验在 log 域：α[i] = log P(J_i)。Σ exp(α) = 1。
// §7：全程 log 域 + 归一化用 log-sum-exp（数值稳定）。
type JointVector []float64

// NewJointVector 零向量（长度 = size）。
func (js *JointSpace) NewJointVector() JointVector {
	return make(JointVector, js.size)
}

// LogSumExp log(Σ exp(x_i))，数值稳定。
func LogSumExp(x []float64) float64 {
	if len(x) == 0 {
		return math.Inf(-1)
	}
	max := x[0]
	for _, v := range x[1:] {
		if v > max {
			max = v
		}
	}
	if math.IsInf(max, -1) {
		return math.Inf(-1)
	}
	sum := 0.0
	for _, v := range x {
		if v-max > -50 { // 跳过可忽略项，防 subnormal
			sum += math.Exp(v - max)
		}
	}
	return max + math.Log(sum)
}

// LogNormalize 原地 log 归一化：v[i] -= LogSumExp(v)，使得 Σ exp(v) = 1。
func (v JointVector) LogNormalize() {
	lse := LogSumExp(v)
	for i := range v {
		v[i] -= lse
	}
}

// logUniform 均匀 log 先验向量（长度 = size）。
func (js *JointSpace) logUniform() JointVector {
	v := js.NewJointVector()
	lu := -math.Log(float64(js.size))
	for i := range v {
		v[i] = lu
	}
	return v
}

// initFromPrior 初始 α = log Prior(S) + log(1/bmaskN)，B 轴均匀。
func (js *JointSpace) initFromPrior(prior Vector) JointVector {
	v := js.NewJointVector()
	logNB := -math.Log(float64(js.bmaskN))
	for s := 0; s < numStates; s++ {
		lp := logP(prior[s]) + logNB
		base := int(s) * js.bmaskN
		for b := 0; b < js.bmaskN; b++ {
			v[base+b] = lp
		}
	}
	v.LogNormalize()
	return v
}

// AddLogToS 给 S==s 的所有联合 cell 的 log 值加 delta（注入 S 轴对数似然，如离房证据 → SLeft）。
// 等价于发射 logPhi 对 S=s 整行加偏置（B 轴不变）：log 域 += → exp 域 ×，即似然比注入。
func (js *JointSpace) AddLogToS(v JointVector, s State, delta float64) {
	base := int(s) * js.bmaskN
	for b := 0; b < js.bmaskN; b++ {
		v[base+b] += delta
	}
}

// MarginalS 边缘出 S 轴分布 P(S) = Σ_bmask exp(α(S,bmask))。
func (js *JointSpace) MarginalS(v JointVector) Vector {
	var out Vector
	for s := 0; s < numStates; s++ {
		base := int(s) * js.bmaskN
		out[s] = math.Exp(LogSumExp(v[base : base+js.bmaskN]))
	}
	return out
}

// MarginalB 边缘出第 j 床 P(B^j=occ) = Σ_{S, bmask: bit_j=1} exp(α)。
func (js *JointSpace) MarginalB(v JointVector, j int) float64 {
	var sumOcc []float64
	for i, lpv := range v {
		_, bmask := js.decode(i)
		if bedOf(bmask, j) == BOcc {
			sumOcc = append(sumOcc, lpv)
		}
	}
	if len(sumOcc) == 0 {
		return 0
	}
	return math.Exp(LogSumExp(sumOcc))
}

// PFallen 读出 P^F_t = Σ_bmask exp(α(SFallen, bmask))（DBN-Zone-Room §8 裁决输入）。
func (js *JointSpace) PFallen(v JointVector) float64 {
	base := int(SFallen) * js.bmaskN
	return math.Exp(LogSumExp(v[base : base+js.bmaskN]))
}

// logP 取 log(p)，p≤0 返回 -inf。
func logP(p float64) float64 {
	if p <= 0 {
		return math.Inf(-1)
	}
	return math.Log(p)
}

// logAdd log(exp(a) + exp(b))，数值稳定。
func logAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	if a > b {
		return a + math.Log1p(math.Exp(b-a))
	}
	return b + math.Log1p(math.Exp(a-b))
}
