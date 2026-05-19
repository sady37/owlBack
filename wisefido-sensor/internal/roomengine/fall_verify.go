// fall_verify.go — radar firmware Fall alarm 后验评分（PR-5）。
//
// 设计目标（用户 2026-04-28 校正）：
//
//	→ 去 ghost：固件 Fall 报警源 track 是 engine 判定的 ghost / birth 不通过 / 位移
//	            与 birth 推算速度爆表时，标记为 ghost-induced false alarm。
//	→ 找到漏报：由 engine 独立的 silent/lost/still fall 路径承担（PR-2b/2c/3 已实现）。
//
// 不使用「firmware 触发 Fall 后会冻结输出」这类伪信号 ——
// frozen 帧、alarm 时刻 pose=Sit/Lie 都是固件 fall 状态的副产物，不是假报标志。
//
// 评分（baseline 50，clamp [0,100]）：
//
//	强 ghost 信号（择一最大减分）：
//	  ts.Verdict == Ghost                                 -40
//	  ts.BirthReason ∈ {no_enter_pair, far_from_enter,    -25
//	                     unknown_area_multitrack}
//	  ts.MaxImpliedSpeedFromBirth > SuspectSpeedCm         -20
//	  fresh_birth(<3s) && BirthScore < 50                  -20
//
//	辅助：
//	  cell.FakeAlarmCount > 0   : -min(15, fc*3)  (历史误报点 — 自学习反馈)
//	  ghost coexist within 10s  : -15
//	  sleepad InBed (30s)       : -10  (人在床压传感器侧确认睡 / 卧床)
//
//	加分（真 fall 上下文）：
//	  最近 60s 有 EnterRoom     : +10  (人活跃在房间)
//	  最近 0.2-3s 进入 Lie 态   : +25  (典型跌倒后躺下转换)
//
// 阈值：
//
//	score < 30  → ghost   (强建议抑制；engine 倾向认为是 ghost-induced false fall)
//	30 ≤ s ≤ 70 → suspect (engine 不否决；保留固件原决策)
//	score > 70  → real    (engine 同意，可作为强证据)

package roomengine

import (
	"math"

	"go.uber.org/zap"
)

// FallVerifyResult 单次 firmware Fall verify 的评分结果与决策。
type FallVerifyResult struct {
	Score     int            // [0,100]
	Verdict   string         // "ghost" / "suspect" / "real"
	Reason    string         // 主导信号 key（绝对值最大的减/加分项）
	Breakdown map[string]int // 各信号贡献分（debug log 用）
}

// FallVerify 阈值（暴露给测试 / cardagg 联动）
const (
	FallVerifyGhostThreshold = 30 // < 此分 → 视为 ghost-induced false fall
	FallVerifyRealThreshold  = 70 // > 此分 → 视为高信心真 fall

	// WeakBioForceRealThreshold WeakBio score ≥ 此值 → fall verifier 短路提级 real。
	// 设计：老人体征长期弱（WeakBio≥80=Red 横条），任何 fall 报警都按真处理（"宁可
	// 误报不可漏报" + WeakBio 是风险放大系数；详 [[target_state_weak_bio_signal_design]]）。
	WeakBioForceRealThreshold = 80
)

// WeakBioSource fall verifier 查询 spatial 实体当前 WeakBio score；service.TargetStateAggregator 隐式满足。
// nil 注入或 score=0 时 verifier 走原 ghost/suspect/real 三档评分（无短路）。
type WeakBioSource interface {
	WeakBioScore(spatialPrefix string) int
}

// VerifyRadarFall 对 firmware Fall 进行后验评分。调用方持锁。
//
// nowMs：当前评估时间（一般等于 alarm.TMs）。
func (tm *TrackManager) verifyRadarFall(a RadarFallAlarm, nowMs int64) FallVerifyResult {
	const (
		scoreBaseline   = 50
		recentEnterMs   = 60_000
		ghostCoexistMs  = 10_000
	)

	score := scoreBaseline
	breakdown := map[string]int{"baseline": scoreBaseline}

	ts := tm.tracks[a.TrackID]
	if ts == nil {
		// engine 没有匹配的 track（firmware 复用 track_id / engine 已 evict）
		score -= 25
		breakdown["no_engine_track"] = -25
		return finalizeFallVerifyResult(score, breakdown)
	}

	// === 反 ghost 强保护（PR-5.3）===
	if ts.LongSurvivalAnchored {
		breakdown["anti_ghost_long_survival"] = 50 // 锚定 Real 后强保护
		score = 100
		return finalizeFallVerifyResult(score, breakdown)
	}
	if ts.StartupGrace {
		breakdown["anti_ghost_startup_grace"] = 50
		score = 100
		return finalizeFallVerifyResult(score, breakdown)
	}

	// === Ghost 信号 ===
	// 主信号：累积 GhostPenalty（PR-5.x 新机制）
	if ts.GhostPenalty > 0 {
		// penalty 直接转化为 fall verifier 的扣分（按比例缩放）
		// GhostPenalty 80 → -50 (强 ghost)；20 → -12.5；clamp [-50, 0]
		p := -int(math.Round(float64(ts.GhostPenalty) * 50.0 / 80.0))
		if p < -50 {
			p = -50
		}
		score += p
		breakdown["ghost_penalty"] = p
	}

	// 辅助信号：原有 ts.Verdict（兜底，用于 GhostPenalty 未累积但 verdict 已定的边缘 case）
	if ts.Verdict == VerdictGhost && ts.GhostPenalty < 30 {
		score -= 20
		breakdown["verdict_ghost_fallback"] = -20
	}

	// === 辅助减分 ===

	// cell.FakeAlarmCount：自学习的历史误报点（PR-4 false_alarm 反馈链产物）
	pxF, pyF := ts.Kalman.Position()
	px, py := int(math.Round(pxF)), int(math.Round(pyF))
	if cell := tm.grid.CellAt(px, py); cell != nil {
		if fc := cell.FakeAlarmCount; fc > 0 {
			d := -fc * 3
			if d < -15 {
				d = -15
			}
			score += d
			breakdown["cell_fake_history"] = d
		}
	}

	// Ghost 共存：当前还有 ghost track 活着 → 雷达此刻不稳定（多次反射 / 旁边有干扰源）
	for tid, other := range tm.tracks {
		if tid == a.TrackID {
			continue
		}
		if other.Verdict == VerdictGhost && nowMs-other.LastUpdateMs < ghostCoexistMs {
			score -= 15
			breakdown["ghost_coexist"] = -15
			break
		}
	}

	// Sleepad InBed：床压侧确认人在床 → 站立 fall 不合理
	if tm.sleepadInBed(nowMs) {
		score -= 10
		breakdown["sleepad_in_bed"] = -10
	}

	// === 加分项 ===

	// 进入 Lie 态在 fall 前 0.2-3s 内 → 典型跌倒后躺下转换
	if ts.LieEnteredAt > 0 {
		elapsed := nowMs - ts.LieEnteredAt
		if elapsed >= 200 && elapsed <= 3000 {
			score += 25
			breakdown["recent_lie_transition"] = 25
		}
	}

	// 最近 60s 有 EnterRoom → 人活跃在房间内
	if tm.hasRecentEnterRoomBetween(nowMs-recentEnterMs, nowMs) {
		score += 10
		breakdown["recent_enter_room"] = 10
	}

	// === WeakBio force real 短路（A 风险放大消费者）===
	// 老人长期体征弱（aggregator WeakBio 30min 滑窗 score ≥ 80 = Red 横条）→ fall
	// 无论原 score 多少都强制 "real"；breakdown 留审计标记。
	//
	// v2 简化 spatial 查询：tm.roomID（/88 room CIDR）作 key，aggregator 不在 /88 时
	// 返 0 自然回到原评分路径（无副作用）。v3 logicID 真融合可优化为 fall device 推
	// 到的具体 /96 bed prefix 查询。
	if tm.weakBioSource != nil {
		if wb := tm.weakBioSource.WeakBioScore(tm.roomID); wb >= WeakBioForceRealThreshold {
			breakdown["weak_bio_force_real"] = wb
			return FallVerifyResult{
				Score:     100,
				Verdict:   "real",
				Reason:    "weak_bio_force_real",
				Breakdown: breakdown,
			}
		}
	}

	return finalizeFallVerifyResult(score, breakdown)
}

func finalizeFallVerifyResult(score int, breakdown map[string]int) FallVerifyResult {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	verdict := "suspect"
	if score < FallVerifyGhostThreshold {
		verdict = "ghost"
	} else if score >= FallVerifyRealThreshold {
		verdict = "real"
	}
	maxAbs := 0
	maxKey := ""
	for k, v := range breakdown {
		if k == "baseline" {
			continue
		}
		av := v
		if av < 0 {
			av = -av
		}
		if av > maxAbs {
			maxAbs = av
			maxKey = k
		}
	}
	return FallVerifyResult{
		Score:     score,
		Verdict:   verdict,
		Reason:    maxKey,
		Breakdown: breakdown,
	}
}

// logFallVerify 把 verifier 结果写 ai.log（structured，cardagg 可订阅）。
func (tm *TrackManager) logFallVerify(a RadarFallAlarm, r FallVerifyResult) {
	fields := []zap.Field{
		zap.String("device_uid", a.DeviceUID),
		zap.Int("track_id", a.TrackID),
		zap.Int("firmware_pose", a.Pose),
		zap.Int("score", r.Score),
		zap.String("verdict", r.Verdict),
		zap.String("dominant_signal", r.Reason),
		zap.Int64("ts_ms", a.TMs),
	}
	for k, v := range r.Breakdown {
		fields = append(fields, zap.Int("bd_"+k, v))
	}
	tm.logger.Info("radar_fall_verify", fields...)
}
