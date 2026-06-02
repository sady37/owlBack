package belief

import "math"

// State 隐状态 S（owlBack/doc/room_belief_state_machine.md §3）。
// S0 Empty + S7 Left 是第一类公民——9h person_silent 的直接缺失态。
// 位置语义不进 S，由 Observation.Geom 作 P(o|s) 条件（§3 设计要点）。
type State int

const (
	SEmpty       State = iota // 房内无人
	SBedLying                 // 在床躺（含睡眠）
	SBedRestless              // 床上翻动/坐起
	SSit                      // 坐（椅/沙发/马桶）
	SStandWalk                // 站立或行走
	SFallen                   // 倒地 ← 唯一 fire Fall 的目标态
	STransition               // 过渡中（吸收瞬时不确定）
	SLeft                     // 从门离开（→ 收敛回 Empty）
	SArtifact                 // ghost / 反射 / 设备伪迹
)

const numStates = 9

var stateLabel = [numStates]string{
	"Empty", "Bed-Lying", "Bed-Restless", "Sit", "Stand-Walk",
	"Floor-Fallen", "Transition", "Left-via-Door", "Artifact",
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
