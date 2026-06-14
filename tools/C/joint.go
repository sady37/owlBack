package belief

import "fmt"

// joint.go — 联合隐状态 J=(S,{B^j})（DBN-Zone-Room §1, §7）。
// S 轴（9 态全空间，state.go）× B 轴（每床 vac/occ，bed_axis.go）的笛卡尔积。
// 基数 = numStates · 2^numBeds。索引：jointIdx = s·(2^numBeds) + bmask，bmask 第 j 位 = B^j。
//
// P-5（状态爆炸 + κ 的 O(|B|^2)）：numBeds 硬 bound 为 maxBeds，超界 panic（不静默膨胀）。
// 养老院房 ≤3–4 床；maxBeds=3 → 9·8=72 态，轻量全表（DBN-Zone-Room §7）。

const maxBeds = 3 // P-5 硬上界。|B|>maxBeds → panic。

// JointVector 联合后验分布 α_t(S,{B^j})，Σ=1。长度 = numStates·2^numBeds（运行期定）。
type JointVector []float64

// JointSpace 联合状态空间元信息。numBeds 进字段（B1 闭合：床数显式持有，不靠隐式推断）。
type JointSpace struct {
	numBeds int // 本房床数 ∈ [0, maxBeds]
	bmaskN  int // 2^numBeds，B 轴配置数（缓存，避免反复移位）
	size    int // numStates · bmaskN，联合空间基数
}

// NewJointSpace 建联合空间。numBeds 超界（<0 或 >maxBeds）→ panic（P-5：拒绝超界，不静默）。
func NewJointSpace(numBeds int) *JointSpace {
	if numBeds < 0 || numBeds > maxBeds {
		panic(fmt.Sprintf("belief: numBeds=%d 超界 [0,%d]（P-5 状态爆炸硬 bound）", numBeds, maxBeds))
	}
	bmaskN := 1 << numBeds // 2^numBeds
	return &JointSpace{
		numBeds: numBeds,
		bmaskN:  bmaskN,
		size:    numStates * bmaskN,
	}
}

// NumBeds / Size 访问器。
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

// bedOf 取 bmask 第 j 位 = B^j（0=vac,1=occ）。
func bedOf(bmask, j int) BedState {
	return BedState((bmask >> j) & 1)
}

// NewJointVector 零向量（长度 = size）。
func (js *JointSpace) NewJointVector() JointVector {
	return make(JointVector, js.size)
}

// normalize 归一化为分布。Σ≤0（证据全湮灭，理论不可达）→ 退均匀分布。
func (v JointVector) normalize() {
	sum := 0.0
	for _, p := range v {
		sum += p
	}
	if sum <= 0 {
		u := 1.0 / float64(len(v))
		for i := range v {
			v[i] = u
		}
		return
	}
	for i := range v {
		v[i] /= sum
	}
}

// MarginalS 边缘出 S 轴分布 P(S) = Σ_{bmask} α(S,bmask)。读出/与 §8 裁决衔接用。
func (js *JointSpace) MarginalS(v JointVector) Vector {
	var out Vector
	for i, p := range v {
		s, _ := js.decode(i)
		out[s] += p
	}
	return out
}

// MarginalB 边缘出第 j 床 P(B^j=occ) = Σ_{S, bmask: bit_j=1} α。
func (js *JointSpace) MarginalB(v JointVector, j int) float64 {
	occ := 0.0
	for i, p := range v {
		_, bmask := js.decode(i)
		if bedOf(bmask, j) == BOcc {
			occ += p
		}
	}
	return occ
}

// PFallen 读出 P^F_t = Σ_{bmask} α(SFallen, bmask)（DBN-Zone-Room §8 裁决输入）。
func (js *JointSpace) PFallen(v JointVector) float64 {
	pf := 0.0
	for bmask := 0; bmask < js.bmaskN; bmask++ {
		pf += v[js.idx(SFallen, bmask)]
	}
	return pf
}
