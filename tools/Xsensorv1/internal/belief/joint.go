package belief

import "fmt"

// joint.go — 联合隐状态 J=(S,{B^j}) 的空间与索引（DBN-Zone-Room §1,§7）。
// S 轴 9 态 × B 轴每床 2 态的笛卡尔积，基数 numStates·2^numBeds。
// 扁平索引 i = s·2^numBeds + bmask，bmask 第 j 位 = B^j（0=vac,1=occ）。
// P-5：numBeds 硬 bound maxBeds，超界 panic（不静默膨胀）。

const maxBeds = 3 // P-5 上界（养老院房 ≤3–4 床）；maxBeds=3 → 9·8=72 态全表。

// JointVector 联合后验 α_t(S,{B^j})，Σ=1，长度 numStates·2^numBeds。
type JointVector []float64

// JointSpace 联合空间元信息。numBeds 显式持有（B1：不靠隐式推断）。
type JointSpace struct {
	numBeds int
	bmaskN  int // 2^numBeds
	size    int // numStates·bmaskN
}

// NewJointSpace 建空间；numBeds 超界（<0 或 >maxBeds）→ panic（P-5）。
func NewJointSpace(numBeds int) *JointSpace {
	if numBeds < 0 || numBeds > maxBeds {
		panic(fmt.Sprintf("belief: numBeds=%d 超界 [0,%d]（P-5 状态爆炸硬 bound）", numBeds, maxBeds))
	}
	bmaskN := 1 << numBeds
	return &JointSpace{numBeds: numBeds, bmaskN: bmaskN, size: numStates * bmaskN}
}

func (js *JointSpace) NumBeds() int { return js.numBeds }
func (js *JointSpace) Size() int    { return js.size }

func (js *JointSpace) idx(s State, bmask int) int { return int(s)*js.bmaskN + bmask }
func (js *JointSpace) decode(i int) (State, int)  { return State(i / js.bmaskN), i % js.bmaskN }

// bedOf 取 bmask 第 j 位 = B^j。
func bedOf(bmask, j int) BedState { return BedState((bmask >> j) & 1) }

// NewJointVector 零向量（长度 = size）。
func (js *JointSpace) NewJointVector() JointVector { return make(JointVector, js.size) }

// normalize 归一化；Σ≤0（理论不可达）退均匀。
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

// MarginalS 边缘 P(S)=Σ_bmask α(S,bmask)；与 §8 裁决衔接。
func (js *JointSpace) MarginalS(v JointVector) Vector {
	var out Vector
	for i, p := range v {
		s, _ := js.decode(i)
		out[s] += p
	}
	return out
}

// MarginalB 边缘第 j 床 P(B^j=occ)。
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

// PFallen 读出 P^F=Σ_bmask α(SFallen,bmask)（§8 裁决输入）。
func (js *JointSpace) PFallen(v JointVector) float64 {
	pf := 0.0
	for bmask := 0; bmask < js.bmaskN; bmask++ {
		pf += v[js.idx(SFallen, bmask)]
	}
	return pf
}
