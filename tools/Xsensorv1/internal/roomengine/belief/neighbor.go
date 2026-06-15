package belief

import "math"

// neighbor.go — 跨房 neighbor 隐轴（§A ρ_xroom 有向门控；方程 DBN-Zone-Room §A.1–§A.3，C §16 审过）。
//
// 本房 real-present track 消失（lost-track，进 Blind*）后，查同 unit 兄弟房有无 **fresh 有向 hand-off**
// （人先走本房、后到兄弟房）→ ρ_xroom 驱动本房 T_S 把 Blind 行的 →Fallen 整流入 →Left（人挪去邻房 vs
// 本房真摔，belief 联合推断）。**ρ=0（无 hand-off）→ 行不变 → 保 lost-fall（安全默认）**：stale/无事件/
// 多住户归因弱 → 不抑制（铁律 [[partial_monitoring_fall_suppression_law]]）。
//
// 有向性（区别 ghost 对称，A 校正点）：源房**先**丢轨、宿房**后**占用——w^dir 对 sign(Δ) **不对称**
// （反向仅容 jitter 抖动，非真 hand-off）。κ(§3)/ghost ρ(§10) 是对称共现，唯 neighbor 带时序方向。
//
// 与 §10 接口：q_r' 吃兄弟房**去 ghost** 占用后验（ghost 轴喂 neighbor，兄弟房 ghost 不算落点）；
// neighbor 是房间 T_S 转移耦合（不加本房 J 隐维，状态空间不爆）；η(rc) 与 §8 C_FN 同 census。

// NeighborParams §A 形态/标定锚（feedback-p6C §9 / 现役 belief_neighbor.go：W=60s/J=5s；曲线留 oracle）。
type NeighborParams struct {
	HandoffWindowMs int64   // W：先走后到最大滞后（固定绝对，不随房距伸缩）
	JitterMs        int64   // J：反向余量（时钟抖动容差）
	TauHMs          float64 // τ_h：新鲜度指数衰减时标
	TauJMs          float64 // τ_j：jitter 衰减（≪τ_h，仅容噪声）
	Beta            float64 // η sole-resident 连续衰减率
}

func DefaultNeighborParams() NeighborParams {
	return NeighborParams{HandoffWindowMs: 60_000, JitterMs: 5_000, TauHMs: 30_000, TauJMs: 2_000, Beta: 0.7}
}

// SiblingHandoff 一个兄弟房的 hand-off 证据（adapter 由跨房 belief 读出 + census 译入）。
type SiblingHandoff struct {
	ArrivalDeltaMs int64   // Δ = t_arrive(兄弟房) − t_lost(本房)；>0 = 先走后到 = 有向命中
	CAttr          float64 // 源型可信度（sleepad InBed 0.9 / room-enter 0.8 / radar-only 0.2）
	PRealPresent   float64 // 兄弟房**去 ghost** 占用后验（§10 接口：用 P(real-present)·(1−P(ghost))）
}

// wDir §A.1 有向新鲜度核（对 sign(Δ) 不对称）。
func wDir(p NeighborParams, deltaMs int64) float64 {
	d := float64(deltaMs)
	switch {
	case deltaMs >= 0 && deltaMs <= p.HandoffWindowMs:
		return math.Exp(-d / p.TauHMs) // 先走后到，新鲜度衰减
	case deltaMs < 0 && -deltaMs <= p.JitterMs:
		return math.Exp(d / p.TauJMs) // 抖动反向余量（d<0）
	default:
		return 0 // 窗外 / 真反向 → 非 hand-off（stale 证不了此刻在哪，不压 fall）
	}
}

// RhoXroom §A.1：ρ_xroom = η(rc) · max_r'[ w^dir(Δ_r') · q_r' ]，q_r' = c_attr·P_realPresent。
// N-4 单合并：至多取最强命中（sole-resident「人只在一处」前提）。
func RhoXroom(p NeighborParams, residentCount int, sibs []SiblingHandoff) float64 {
	if residentCount < 1 {
		residentCount = 1
	}
	eta := math.Exp(-p.Beta * float64(residentCount-1)) // 单住户 η=1；多住户弱但不归零
	best := 0.0
	for _, s := range sibs {
		if v := wDir(p, s.ArrivalDeltaMs) * s.CAttr * s.PRealPresent; v > best {
			best = v
		}
	}
	rho := eta * best
	if rho > 1 {
		rho = 1
	}
	return rho
}

// GateBlindRow §A.2：把 Blind 行的 →Fallen 倾向按 ρ_xroom 整流入 →Left（仅 lost-track/Blind 行用）。
//   T̃_S(F|S') = (1−ρ)·T0(F|S')，T̃_S(L|S') = T0(L|S') + ρ·T0(F|S')；其余项不变，F↔L 转移行和守恒。
// ρ=0 → 行不变（Blind 照常 ramp Fallen = lost-fall 安全默认）；ρ→1 → Fallen 倾向整流入 Left。
func GateBlindRow(row [numStates]float64, rho float64) [numStates]float64 {
	moved := rho * row[SFallen]
	row[SFallen] -= moved
	row[SLeft] += moved
	return row
}
