// ghost_adjudicator.go — sensor_v2 Layer 1 ghost adjudication 接口契约
//
// sensor_v2 §4.A / §4.B / §10.3.1 / PR-4。
//
// 设计原则（4 条铁律）：
//
//  1. **adjudicator 是后置纯函数**
//     TrackManager.ProcessFrame 跑完 v1 累积器 → engine SnapshotTrackStatuses 取 base 投影
//     → adjudicator.Adjudicate(bases, nowMs) 输出 []VerdictDelta → engine 应用到 TrackStatus
//     副本（不改 TrackState 内部累积器）。
//     adjudicator 不能持有 TrackManager 引用 / 不能写 PG / 不能发 redis stream，
//     边界外的副作用走 FeedbackEvent 反馈通道（§3.3，PR-X）。
//
//  2. **Verdict 单源真相在 TrackStatus 副本**
//     engine.applyVerdictDeltas 是唯一的 verdict 写点（PR-3 v1 累积器赋的 base.Verdict + 本函数 override）。
//     CLAUDE.md #1.3 — 不允许 adjudicator 自己绕开 applyVerdictDeltas 改 status。
//
//  3. **Anchored 不可翻 Ghost**（决定 21 / §10.1.1 / Q4 守卫）
//     applyVerdictDeltas 强制此不变量：任何 adjudicator 返回 NewVerdict=Ghost
//     但当前 base.Verdict=Anchored 时，verdict 字段保持 Anchored，penalty 仍可累加供审计观察。
//     未来若需 unanchor，必须走显式重置路径（重启 / FeedbackEvent 清 LongSurvival/StartupGrace），
//     不能在单帧 delta 里悄悄做。
//
//  4. **分支选择由 room.kind binary 决定**（决定 16 / §10.3.1）
//     room_kind == "bathroom" → BathroomGhostAdjudicator (§4.A)
//     其他（含默认 "") → GeneralGhostAdjudicator (§4.B)
//     选择发生在 engine.publishTrackStatuses，每帧重判（不缓存房间 → adjudicator 映射），
//     运行时 RoomConfig 变更 (PR-X hot reload) 自动跟随。

package roomengine

import "go.uber.org/zap"

// VerdictDelta adjudicator 对单 track 的判定建议。
//
// 语义：
//   - NewVerdict != nil → 建议把 verdict 改成 *NewVerdict（受 Anchored 守卫约束）
//   - PenaltyDelta != 0 → 在当前 status.GhostPenalty 上累加（结果 clamp 到 [0, 100]）
//   - Reason → 写 ai.log 留审计；adjudicator 必填，至少一句"为什么"
//
// 不在本结构里出现的字段（明确不属 adjudicator 职责）：
//   - Anchor 状态变更：走 SuiteCensus.UpdatePersonFromTrack / ClearAnchorOnLostTrack
//   - 床/区状态翻转：走 zoneengine 既有路径
//   - alarm 触发：走 fall verifier，consume TrackStatus 后自己决策
type VerdictDelta struct {
	TrackID      int
	NewVerdict   *TrackVerdict
	PenaltyDelta int
	Reason       string
}

// GhostAdjudicator Layer 1 ghost 判定器接口。
// PR-4 引入；PR-6 BathroomGhostAdjudicator 填充 §4.A 4 条规则；
// PR-7 GeneralGhostAdjudicator 接入 §4.B 统计累积式信号。
//
// 输入 bases 为本帧所有 live track 的 Layer 1 投影（含 v1 累积器赋的 verdict + LongSurvival/StartupGrace 升 Anchored）；
// 输出 deltas 长度可 < len(bases)，未提及的 trackID 保留 base 状态。
type GhostAdjudicator interface {
	Adjudicate(bases []TrackStatusBase, nowMs int64) []VerdictDelta
}

// NoopGhostAdjudicator 永远返回空 delta 的默认实现。
// 用途：PR-4 默认 wire；保证"未注入真实 adjudicator 时行为 == v1"。
type NoopGhostAdjudicator struct{}

// Adjudicate 返回 nil（等价空 slice，applyVerdictDeltas 跳过）。
func (NoopGhostAdjudicator) Adjudicate(_ []TrackStatusBase, _ int64) []VerdictDelta {
	return nil
}

// BathroomGhostAdjudicator §4.A bathroom ghost 判定器。
// 完整实现在 bathroom_ghost.go（PR-6 落地）。本处仅保留前向声明意图：
// 详 sensor_v2 §4.A.3 4 条规则（出生地 / 占用一致性 / 共生律 / split-ghost）+ §4.A.3 规则 4 fallback。

// GeneralGhostAdjudicator §4.B 默认 (bedroom + 其他非 bathroom room) 判定器骨架。
//
// PR-4 范围：Noop 实现；v1 ghost 判定逻辑仍在 TrackManager 内部累积器（保持 v1 行为）。
// PR-7 范围：接入 §4.B 4 条统计累积式信号（Sleepad 跨源排他 / Anchor 互斥 / Traverse 差异 / Motion correlation）。
//
// 设计意图：v2 §4.B 是"非 bathroom 的兜底 + 长时间累积观察"路径，不能像 v1 那样在 TrackManager
// 内部硬编码 — 必须可被 NoopAdjudicator 替代，方便 dev / playback / fall-score-replay 跑纯 v1 baseline。
type GeneralGhostAdjudicator struct {
	logger *zap.Logger
}

// NewGeneralGhostAdjudicator 构造。logger nil 时降级 zap.NewNop。
func NewGeneralGhostAdjudicator(logger *zap.Logger) *GeneralGhostAdjudicator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &GeneralGhostAdjudicator{logger: logger}
}

// Adjudicate PR-4 阶段 Noop（行为 == v1）；PR-7 填 §4.B 信号。
func (a *GeneralGhostAdjudicator) Adjudicate(_ []TrackStatusBase, _ int64) []VerdictDelta {
	return nil
}
