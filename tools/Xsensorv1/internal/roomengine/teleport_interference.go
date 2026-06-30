// teleport_interference.go — 单 tick 瞬移干扰轨的 PURGE（track-lifecycle 轴，fire-neutral）。
//
// 物理：固件偶发把一条轨瞬间漂到镜面多径假位置（velocity 远超人体）。引擎若把 Kalman 拖到漂移点继续跟，
// 人走开后这条轨冻死在假位置 → 转 lost 进 blind-faller hold → still_sec 累积 → FloorGuard 兜底误火。
// 本机制只处理最无歧义的一类：**单 tick 瞬移 ≥ teleportJumpCmS（200cm/s）**。sub-200 缓漂不在此范围。
//
// 删的是 artifact（本就不是人），走 track-lifecycle 轴；**不接 realness→fire veto**（铁律
// [[realness_never_vetoes_fall]]）。fire 因 track 集修正而变是删非人轨的副作用，绝不压真人摔。
//
// 检测挂在摄入循环 Kalman 更新**之前**（看原始跳变），O(1)/轨：一次距离 + 一次除法。
// 过两道安全闸再决定 PURGE：
//   闸1（FN-safety 硬底线）：跳变帧 + 跳前 2 帧无摔倒前兆，才允许删。
//   闸2（确认）：真人已被另一条 present 活轨接住（换新 tid 重跟），删这条零损失。
// 闸1 过、闸2 未到 = 待删窗：打 SuspectInterference 标记，从 floor/still-box 累积排除（SnapshotTrackStatuses
// 置 StillBoxSec=0），等闸2 或超时（teleportSuspectTimeoutMs）再 PURGE。
// PURGE 后固件持续重报该 tid（~30s），靠 interferenceSuppress 集压制重建直到它真消失（NoTarget）或移出漂移点。

package roomengine

import (
	"owl-common/observation"

	"go.uber.org/zap"
)

const (
	teleportJumpCmS          = 200   // 跳变速度阈（cm/s）：≥ 此视为瞬移嫌疑（Kalman innovation 尖峰的简化代理）
	teleportMinJumpCm        = 150   // 跳变绝对位移下限（cm）：物理 teleport 是大跳，排 117cm 量级缓漂 + 抖动（防小 dt 把小位移算成高速）
	teleportMinDtSec         = 0.3   // 观测间隔下限（s）：< 此 = 突发/重号/异常帧，velocity 不可靠 → 不判（radar 实测 0.6~1.2s）
	teleportZDropCm          = 40    // 闸1：相邻帧 |Δz| > 此 = 质心骤降（塌陷前兆）→ 不删
	teleportSuspectTimeoutMs = 5_000 // 待删窗超时（ms）：闸2 仍不来且闸1 仍成立 → 也 PURGE（真人必已重生 tid）
	interferenceCoverCm      = 100   // 闸2：另一条活轨距真人原区域 ≤ 此 = 接住了
	interferenceSuppressCm   = 50    // 压制匹配半径：固件重报落点距漂移点 ≤ 此 = 仍滞留 → 压制重建
	interferenceSuppressTTL  = 5_000 // 压制集 GC：tid 距上次在漂移点重报超此 ms = 已消失 → 回收条目
	immatureLifeMs           = 5_000  // immature 阈：出生到末次真实观测 < 此 = 没确立为真人（1-tick→0）
	floorArtifactDeferMs     = 30_000 // 「延后 30s 再看」：coast 须持续 ≥ 此才 latch（真人瞬时断帧会更早重现→不误抑制；floor 20min 预算下零代价）
)

// gate1NoPrecursorFrozen 闸1（immature-coast 路径）：看已冻结轨的末真实 pose 非倒地前兆 + Z 平稳。
// 与 teleport 路径的区别：immature 轨可能只有 1 帧（Prev2Pose=0 无 N-2）→ N-2 仅在确有观测(>0)时才校验，
// 不因"没有 N-2"而拒（否则 1-tick 反射永远 latch 不上，正是 lid389）。
func gate1NoPrecursorFrozen(ts *TrackState) bool {
	if teleportFallPrecursorPose(ts.LastPose) || !teleportNonPrecursorPose(ts.LastPose) {
		return false // 末真实帧必须明确非倒地（pose∈{1,3,4}）
	}
	if ts.Prev2Pose != observation.PoseUnknown && teleportFallPrecursorPose(ts.Prev2Pose) {
		return false // 有 N-2 且是倒地前兆 → 不抑制
	}
	if n := len(ts.History); n >= 2 && abs(ts.History[n-1].Z-ts.History[n-2].Z) > teleportZDropCm {
		return false // 末两帧质心骤降 = 塌陷前兆
	}
	return true
}

// hasOtherPresentRealTrack immature-coast 闸2（共存）：除 exceptKey 外是否还有 present、PReal≥0.5 的活轨
// （不要求邻近——远墙反射的真人通常不在它身边）。单轨独处=false → 永不抑制（铁律：独处全发）。
func (tm *TrackManager) hasOtherPresentRealTrack(exceptKey trackKey, nowMs int64) bool {
	for k, ts2 := range tm.tracks {
		if k == exceptKey || ts2.PReal < 0.5 {
			continue
		}
		if nowMs-ts2.LastObservedMs < presenceCoastMs {
			return true
		}
	}
	return false
}

// maybeLatchFloorArtifact coast 续存循环调用（持锁）：一条转 coast 的孤儿若「真实寿命 < immatureLifeMs +
// 闸1 无倒地前兆 + 有活轨共存」→ latch FloorArtifactSinceMs，其 StillBoxSec 从此不喂 floor。一次性置位。
func (tm *TrackManager) maybeLatchFloorArtifact(ts *TrackState, key trackKey, nowMs int64) {
	if ts.FloorArtifactSinceMs > 0 {
		return // 已标
	}
	if nowMs-ts.LastObservedMs < floorArtifactDeferMs {
		return // 延后 30s：coast 不够久 → 可能真人瞬时断帧，等它确实没回来再说
	}
	if ts.LastObservedMs-ts.BirthPos.TMs >= immatureLifeMs {
		return // 成熟轨（观测够久）→ 可能真 faller，不抑制
	}
	if !gate1NoPrecursorFrozen(ts) {
		return // 露过倒地相 → 绝不抑制（FN-safe）
	}
	if !tm.hasOtherPresentRealTrack(key, nowMs) {
		return // 独处 → 不抑制（这条可能就是唯一摔倒的人）
	}
	ts.FloorArtifactSinceMs = nowMs
	tm.logger.Info("floor_artifact_latched",
		zap.String("device_uid", key.dev), zap.Int("track_id", key.tid),
		zap.String("logic_id", ts.LogicID),
		zap.Int("birth_x", ts.BirthPos.X), zap.Int("birth_y", ts.BirthPos.Y),
		zap.Int64("real_life_ms", ts.LastObservedMs-ts.BirthPos.TMs),
		zap.Int("last_pose", ts.LastPose))
}

// interferenceSuppressRec 已判 artifact 的 (device,tid) 压制重建记录。
type interferenceSuppressRec struct {
	driftX, driftY int
	lastSeenMs     int64 // 最近一次在漂移点重报；GC 基准（NoTarget 后 TTL 内回收）
}

// teleportNonPrecursorPose 闸1 白名单：跳变前 2 帧 pose 须全在此（行走/蹲坐/站立）。
// Lying(6)/Unknown(0) 不在内 → 自然挡住（含未攒够姿态历史的新生轨，FN-safe）。
func teleportNonPrecursorPose(pose int) bool {
	return pose == observation.PoseWalking || pose == observation.PoseSitting || pose == observation.PoseStanding
}

// teleportFallPrecursorPose 跳变帧 pose 命中任一摔倒前兆 → 闸1 不通过（可能真摔者被干扰带跳，绝不删）。
func teleportFallPrecursorPose(pose int) bool {
	return pose == observation.PoseSuspectedFall || pose == observation.PoseFallen ||
		pose == observation.PoseSuspectedSitGround || pose == observation.PoseSitGround
}

// poseEvictable 离场残迹可删白名单：排倒地前兆(fall/sitGnd) + walking + sitting（真摔者/行走/真坐者绝不按残迹删）。
func poseEvictable(pose int) bool {
	return !teleportFallPrecursorPose(pose) && pose != observation.PoseWalking && pose != observation.PoseSitting
}

// teleportGate1NoPrecursor 闸1：跳变帧 + 跳前 2 帧无摔倒前兆，全满足才允许删。
// 调用方持锁；读 ts.LastPose(N-1)/Prev2Pose(N-2)/History tail Z(N-1,N-2)，均 O(1)。
func teleportGate1NoPrecursor(ts *TrackState, f TrackFrame) bool {
	if teleportFallPrecursorPose(f.Pose) {
		return false
	}
	if !teleportNonPrecursorPose(ts.LastPose) || !teleportNonPrecursorPose(ts.Prev2Pose) {
		return false
	}
	n := len(ts.History)
	if n < 2 {
		return false // z 历史不足 N-2 → 保守不删
	}
	zPrev := ts.History[n-1].Z
	zPrev2 := ts.History[n-2].Z
	if abs(f.Z-zPrev) > teleportZDropCm || abs(zPrev-zPrev2) > teleportZDropCm {
		return false // 跳前到跳变帧之间有质心骤降 = 塌陷前兆
	}
	return true
}

// interferenceGate2Covered 闸2：除 excludeKey 外是否存在 present(在场)、PReal≥0.5 的活轨覆盖 (x,y)（真人原区域）。
// 有 = 真人换新 tid 被重新跟踪，删 artifact 零损失。事件驱动 O(T)，仅瞬移命中时算。调用方持锁。
func (tm *TrackManager) interferenceGate2Covered(excludeKey trackKey, x, y int, nowMs int64) bool {
	for k, ts2 := range tm.tracks {
		if k == excludeKey || ts2.PReal < 0.5 {
			continue
		}
		if nowMs-ts2.LastObservedMs >= presenceCoastMs {
			continue // 非 present
		}
		pxF, pyF := ts2.Kalman.Position()
		if distInt(int(pxF), int(pyF), x, y) <= interferenceCoverCm {
			return true
		}
	}
	return false
}

// interferenceSuppressed 压制检查（段 1 帧循环最前调）：该 (device,tid) 是否仍滞留 flagged 漂移点。
// 是 → 压制重建（调用方丢弃该帧不创建/更新 track）；移出漂移点 = 真人复用该 tid → 解压制。调用方持锁。
func (tm *TrackManager) interferenceSuppressed(key trackKey, f TrackFrame, nowMs int64) bool {
	rec, ok := tm.interferenceSuppress[key]
	if !ok {
		return false
	}
	if distInt(f.X, f.Y, rec.driftX, rec.driftY) <= interferenceSuppressCm {
		rec.lastSeenMs = nowMs
		return true
	}
	delete(tm.interferenceSuppress, key)
	return false
}

// handleTeleportObservedFrame 摄入循环（已有 track 分支）Kalman 更新前调，O(1)/轨。
// 返回 true = 该 artifact 已 PURGE，调用方应 continue 跳过本帧后续。
func (tm *TrackManager) handleTeleportObservedFrame(ts *TrackState, key trackKey, f TrackFrame, nowMs int64) bool {
	// 已在待删窗：re-check 闸2 / 超时 → PURGE
	if ts.SuspectInterferenceSinceMs > 0 {
		if tm.interferenceGate2Covered(key, ts.PreJumpX, ts.PreJumpY, nowMs) {
			tm.purgeInterferenceArtifact(ts, key, f.X, f.Y, nowMs, "gate2_covered")
			return true
		}
		if nowMs-ts.SuspectInterferenceSinceMs >= teleportSuspectTimeoutMs {
			tm.purgeInterferenceArtifact(ts, key, f.X, f.Y, nowMs, "window_timeout")
			return true
		}
		ts.DriftX, ts.DriftY = f.X, f.Y // 固件漂移点微抖 → 更新供后续 suppress 落点
		return false                    // 待删窗内保活；SnapshotTrackStatuses 已零其 StillBoxSec
	}

	// 未 flagged：raw 跳变检测（用速度抗变 tick 间隔，实测 0.6~1.2s）
	n := len(ts.History)
	if n == 0 {
		return false
	}
	prev := ts.History[n-1] // 末次 raw 位置（coast 期为冻结同位，位置仍有效）
	// dt 用 LastObservedMs（真实观测的数据时钟）而非 History[last].TMs：lost-coast 期 History 被填
	// nowMs(空 tick=墙钟)，replay 加速下与数据钟错位 → dt 失真 → 小位移被算成高速误判（实测 B17F 10cm 误触）。
	dtObs := float64(f.TMs-ts.LastObservedMs) / 1000.0
	if dtObs < teleportMinDtSec {
		return false // dt 过小（突发/重号/异常帧）：velocity 不可靠，不冒险删
	}
	jumpCm := distInt(prev.X, prev.Y, f.X, f.Y)
	// 须**既大位移**（物理 teleport 下限，排 117cm 量级缓漂 + 抖动）**又超速**（>人体上限）。
	if jumpCm < teleportMinJumpCm || float64(jumpCm)/dtObs < teleportJumpCmS {
		return false
	}
	if !teleportGate1NoPrecursor(ts, f) {
		return false // 闸1：有前兆 → 可能真摔者被带跳，绝不删
	}

	// 闸1 过 → 进待删窗（记真人原区域 + 漂移点）
	ts.SuspectInterferenceSinceMs = nowMs
	ts.PreJumpX, ts.PreJumpY = prev.X, prev.Y
	ts.DriftX, ts.DriftY = f.X, f.Y
	if tm.interferenceGate2Covered(key, prev.X, prev.Y, nowMs) {
		tm.purgeInterferenceArtifact(ts, key, f.X, f.Y, nowMs, "gate2_covered") // 闸2 即时满足 → 立即 PURGE
		return true
	}
	tm.logger.Info("teleport_interference_suspect",
		zap.String("device_uid", key.dev), zap.Int("track_id", key.tid),
		zap.Int("drift_x", f.X), zap.Int("drift_y", f.Y),
		zap.Int("pre_jump_x", prev.X), zap.Int("pre_jump_y", prev.Y),
		zap.String("logic_id", ts.LogicID))
	return false
}

// purgeInterferenceArtifact PURGE（非 lost）：整条移除 + 漂移点记一次 reflector candidate + 登记压制集。
// 内联 delete（调用方已持 tm.mu，不能走会重锁的 EvictTrack）。
func (tm *TrackManager) purgeInterferenceArtifact(ts *TrackState, key trackKey, driftX, driftY int, nowMs int64, reason string) {
	if tm.grid != nil {
		// 漂移点排查：只 +1 candidate，需 ≥3 episode + 近墙 + 300s 才升格，误标一次自然衰减无害。
		tm.grid.MarkStaticReflector(driftX, driftY, nowMs)
	}
	tm.interferenceSuppress[key] = &interferenceSuppressRec{
		driftX: driftX, driftY: driftY, lastSeenMs: nowMs,
	}
	delete(tm.tracks, key)
	delete(tm.outputs, key)
	tm.logger.Info("teleport_interference_purge",
		zap.String("reason", reason),
		zap.String("device_uid", key.dev), zap.Int("track_id", key.tid),
		zap.Int("drift_x", driftX), zap.Int("drift_y", driftY),
		zap.Int("pre_jump_x", ts.PreJumpX), zap.Int("pre_jump_y", ts.PreJumpY),
		zap.String("logic_id", ts.LogicID))
}

// gcInterferenceSuppress 回收已消失 tid 的压制条目（NoTarget 后 TTL 内）。每帧 O(suppress-size)，常态 0–2。
func (tm *TrackManager) gcInterferenceSuppress(nowMs int64) {
	for k, rec := range tm.interferenceSuppress {
		if nowMs-rec.lastSeenMs > interferenceSuppressTTL {
			delete(tm.interferenceSuppress, k)
		}
	}
}
