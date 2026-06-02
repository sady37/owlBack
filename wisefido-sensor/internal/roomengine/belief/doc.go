// Package belief 是房间级人状态信念估计器：用显式概率信念 b∈Δ(S) 替代 gate-list 规则引擎。
//
// 数学同源 zoneengine/bed_bayesian_scorer.go（床二态 log-odds），本包升维到房间多态（§3 S 九态）。
// 每帧两步 forward：① b←A·b 时间转移；② b←normalize(diag(P(o|s))·b) 观测更新。
// 缺证据（Conf=0 / !Fresh）→ 该观测 likelihood=I 不更新，治本 9h person_silent 整类 bug。
//
// 设计文档：owlBack/doc/room_belief_state_machine.md（总）/ belief_input_normalization.md（输入）/
// belief_gate_to_matrix.md（≈100 gate→矩阵条目对照）。v1 scope=单实体 + §5.5.2 弱耦合，shadow-only。
package belief
