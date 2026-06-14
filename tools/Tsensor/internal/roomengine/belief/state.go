package belief

import "math"

// State 隐状态 S：全空间区域占用 × {直立, 倒地}（P5 重写，feedback_a5 委员会批准）。
// 区域（占用在哪）进 S；姿态由 Fallen 吸收态承载。ghost/伪迹归 Track 层（Conf×P(Real) 退火），S 不设伪迹态。
// Blind* = 否定定义（占用但无 real track 解析 + 未离场）；Rest/Open 由消失前区域经转移阵 A 决定，不靠几何（盲区无几何）。
type State int

const (
	SEmpty     State = iota // 0 房内确认无人
	SBed                    // 1 占用在床（睡/休息）
	SSit                    // 2 占用在休息区（椅/沙发）——久坐正常
	SOpenFloor              // 3 占用在开阔活动区（站/走，摔在这里）
	SBath                   // 4 占用在卫浴（高风险，常盲）
	SFallen                 // 5 倒地 ← 唯一 fire Fall 的目标态，近吸收（index 不变 → Decider 引用稳定）
	SBlindRest              // 6 占用但无 real track 解析 + 未离场，从 rest 区（床/卫浴）进盲
	SBlindOpen              // 7 同上，从开阔区进盲
	SLeft                   // 8 从门离开（→ 收敛回 Empty）
)

const numStates = 9

var stateLabel = [numStates]string{
	"Empty", "Bed", "Sit", "Open-Floor", "Bath",
	"Floor-Fallen", "Blind-Rest", "Blind-Open", "Left-via-Door",
}

func (s State) String() string { return stateLabel[s] }

// Vector 是 Δ(S) 上的概率向量，Σ=1。永远是分布，不是确定状态。
type Vector [numStates]float64

// normalize 归一化为分布。和为 0（理论不可达，证据全湮灭）时退回均匀分布。
func (v *Vector) normalize() {
	sum := 0.0
	for _, p := range v {
		sum += p
	}
	if sum <= 0 {
		for i := range v {
			v[i] = 1.0 / numStates
		}
		return
	}
	for i := range v {
		v[i] /= sum
	}
}

// Max 返回最大分量及其状态——决策层判"信念是否够确定"用。
func (v Vector) Max() (State, float64) {
	best, bestP := SEmpty, v[0]
	for i := 1; i < numStates; i++ {
		if v[i] > bestP {
			best, bestP = State(i), v[i]
		}
	}
	return best, bestP
}

// P 取某态后验概率。
func (v Vector) P(s State) float64 { return v[s] }

// lk 构造 likelihood 向量：默认全 1.0（中性），只覆盖 set 中给定的状态权重。
// weight>1 偏好该态，<1 压制该态，永不取 0（最小 likelihoodFloor，防分量湮灭）。
func lk(set map[State]float64) Vector {
	var v Vector
	for i := range v {
		v[i] = 1.0
	}
	for s, w := range set {
		if w < likelihoodFloor {
			w = likelihoodFloor
		}
		v[s] = w
	}
	return v
}

const likelihoodFloor = 0.1

// temper 用 Conf 对 likelihood 退火：w^Conf。
// Conf=0 → 全 1.0（=I 单位阵，不更新，§2.3 命门）；Conf=1 → 全强度。
// 与 bed scorer 的 γ tempering 同构（gammaLocked）。
func temper(w Vector, conf float64) Vector {
	if conf <= 0 {
		return lk(nil)
	}
	if conf >= 1 {
		return w
	}
	var out Vector
	for i := range w {
		out[i] = math.Pow(w[i], conf)
	}
	return out
}
