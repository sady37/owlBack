// split_ghost.go — 单人贴静止干扰源时 firmware tid 交换/分裂的收治（track-lifecycle 轴，fire-neutral）。
//
// 物理：室内有强干扰源（金属/反射，**基本不动**）。真人走近时 firmware 把 tid 标签在「真人」与
// 「干扰源幻影」之间交换/分裂——**tid 会交换**，不能按 tid 大小/出生先后认 real，只能纯行为证伪。
//
// 三段式（对象=split group，不 per-tid 各算）：
//   step1 找 spliter：np↑ 出新 offspring（无 EnterRoom）时，t1 各轨投最近的 t0（前一tick）轨；得票最多者=spliter。
//   spliter 静止门（仅 spliter 受限，因干扰源不动）：spliter 在 t-1,t-2 逐轴位移 ≤ splitSpliterStillCm。
//   step2 建 split_group：投了 spliter 的 t1 轨入组，锚点=spliter 前一tick位。offspring 不设距离约束（真人可乱走）。
//   step3 每帧维护、~2s 节流：按**离锚点净距**三档——≥200 走开=real（EverWalkedOut 永久锁）／≤50 赖锚点
//         连续 3 tick 坐实=ghost → 单调软压 SplitGhostSinceMs（SnapshotTrackStatuses 零其 StillBoxSec，
//         不喂 floor，不删轨=无 churn）／50~200 HOLD 存疑不动。
//
// 安全性：软压**可撤销**（离开锚点/走开 → 清）；露倒地相（fall-precursor pose）立即解压并锁 real（FN-safe，
// 真人贴干扰源摔倒绝不被压）。purge 一律不用（EvictTrack→firmware 重报→每帧重建 lid churn）。
// FN 兜底仍在 firmware fire / tFloor 两条独立线（真人 silent fall 走开后正常攒 floor）。

package roomengine

import (
	"math"

	"go.uber.org/zap"
)

const (
	splitLatchIntervalMs = 2_000 // step3 节流：每轨每 ~2s 算一次（离锚点净距是保持量，不漏）
	splitSpliterStillCm  = 10    // spliter t-1,t-2 逐轴位移 ≤此 = 静止干扰源（雷达噪声 ±5-10cm 地板）
	splitGluedCm         = 50    // 离锚点净距 ≤此 = GLUED（对齐 stillbox 50×50 半框，吸收旧 glue 10 + 噪声）
	splitWalkOutCm       = 200   // 离锚点净距 ≥此 = WALKOUT 真人 walk → EverWalkedOut 永久 real
	splitConvictTicks    = 3     // GLUED 连续 ≥此 tick（~6s）坐实 → SplitGhostSinceMs 单调置位（抗噪）
	splitMinVotes        = 2     // spliter 至少 2 票（自身续 + offspring）才算 split（一源裂多轨）
	splitLiveMoveCm      = 150   // succession 2-tick 净走动 >此 = 瞬时 liveness（反射钉死原处 2 帧走不出）。基准 1frame/s；帧率变则重标
)

func splitAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// glueState 离锚点净距三档（§4）：GLUED 赖锚点=删除/定罪候选 / HOLD 存疑（压制信宽、删除信严，两消费者反向处置）/ WALKOUT 走开=real。
type glueState int

const (
	glueGLUED glueState = iota
	glueHOLD
	glueWALKOUT
)

// kinematicGlued 当前位（Kalman 末次滤波位）离锚点净距三档，纯几何瞬时判定（无时间/latch 态）。
// 锚点每路语义不同（§3.2）：split=spliter 前一tick位（SplitSpliterX/Y）、interfer/门距=出生点。复用 Stage1 的 splitGluedCm/splitWalkOutCm。
func kinematicGlued(ts *TrackState, anchorX, anchorY int) glueState {
	xF, yF := ts.Kalman.Position()
	d := distInt(int(math.Round(xF)), int(math.Round(yF)), anchorX, anchorY)
	switch {
	case d >= splitWalkOutCm:
		return glueWALKOUT
	case d <= splitGluedCm:
		return glueGLUED
	default:
		return glueHOLD
	}
}

// histPos 取 History 倒数第 (ticksAgo+1) 个点（ticksAgo=0 当前帧 / 1 前一tick / 2 前二tick）。
func histPos(ts *TrackState, ticksAgo int) (x, y int, ok bool) {
	idx := len(ts.History) - 1 - ticksAgo
	if idx < 0 {
		return 0, 0, false
	}
	p := ts.History[idx]
	return p.X, p.Y, true
}

// stampSplitGroupOnBirth 出生钩子（调用方已判：offspring 无 EnterRoom + 无继承）。step1 vote → spliter
// 静止门 → step2 标 group。offspring 尚未入 tm.tracks（调用方随后加），故 range 不含它，直接在 ts 上打标。
func (tm *TrackManager) stampSplitGroupOnBirth(offspring *TrackState, f TrackFrame, nowMs int64) {
	// t0：同设备已存在轨在「前一tick」的坐标
	type t0c struct {
		ts     *TrackState
		px, py int
	}
	var t0 []t0c
	for _, o := range tm.tracks {
		if o.DeviceAddr != f.DeviceAddr || o.Kalman == nil {
			continue
		}
		if px, py, ok := histPos(o, 1); ok {
			t0 = append(t0, t0c{o, px, py})
		}
	}
	if len(t0) == 0 {
		return
	}
	// t1：offspring（当前帧位）+ 已存在轨（当前 Kalman 位）
	type t1c struct {
		ts   *TrackState
		x, y int
	}
	t1 := []t1c{{offspring, f.X, f.Y}}
	for _, o := range tm.tracks {
		if o.DeviceAddr != f.DeviceAddr || o.Kalman == nil {
			continue
		}
		ox, oy := o.Kalman.Position()
		t1 = append(t1, t1c{o, int(ox), int(oy)})
	}
	// 投票：每条 t1 投给最近 t0；并列(相等)不投
	votes := map[*TrackState]int{}
	pick := map[*TrackState]*TrackState{}
	for _, a := range t1 {
		bestD, tie := 1<<30, false
		var best *TrackState
		for i := range t0 {
			d := distInt(a.x, a.y, t0[i].px, t0[i].py)
			if d < bestD {
				bestD, best, tie = d, t0[i].ts, false
			} else if d == bestD {
				tie = true
			}
		}
		if best != nil && !tie {
			votes[best]++
			pick[a.ts] = best
		}
	}
	// spliter = 得票最多
	var spliter *TrackState
	top := 0
	for ts, v := range votes {
		if v > top {
			top, spliter = v, ts
		}
	}
	if spliter == nil || top < splitMinVotes {
		return
	}
	// spliter 静止门（仅 spliter）：t-1,t-2 逐轴 ≤ splitSpliterStillCm（动了=非干扰源,交 teleport/static）
	x1, y1, ok1 := histPos(spliter, 1)
	x2, y2, ok2 := histPos(spliter, 2)
	if !ok1 || !ok2 || splitAbs(x1-x2) > splitSpliterStillCm || splitAbs(y1-y2) > splitSpliterStillCm {
		return
	}
	// step2 标 group：投了 spliter 的 t1 轨，锚点=spliter 前一tick位
	n := 0
	for ts, sp := range pick {
		if sp != spliter {
			continue
		}
		ts.SplitObservingSinceMs = nowMs
		ts.SplitSpliterX, ts.SplitSpliterY = x1, y1
		ts.SplitSpliterLID = spliter.LogicID // succession 反向链：组员记住占用 mantle 持有者身份
		n++
	}
	tm.logger.Info("split_group_formed",
		zap.String("device_uid", f.DeviceAddr),
		zap.Int("spliter_track_id", spliter.TrackID),
		zap.Int("anchor_x", x1), zap.Int("anchor_y", y1),
		zap.Int("group_size", n),
		zap.Int("offspring_track_id", offspring.TrackID),
	)
}

// updateSplitGhost step3（每观测帧调，~2s 节流）：group 成员按「离锚点净距」三档定 real/ghost。
// 基准=离锚点(spliter 前一tick位)净距，**非离出生点**——续轨旧 BirthPos 是 split 前走过的污染源
// （把 split 前的位移误当 split 后走出 → 秒判 real）；且绕锚点画圈的慢漂 ghost 骗得过离出生点净距。
//   - ≥200 WALKOUT → 真人 walk → SplitEverWalkedOut 永久 real，解压。
//   - ≤50  GLUED   → 喂 split 本地 3-tick 棘轮；连续坐实 → SplitGhostSinceMs 单调置位（软压）。
//   - 50~200 HOLD  → 存疑：不定罪（断棘轮）、不清既有定罪（穿过 HOLD）。
// 压制 SplitGhostSinceMs 单调，只由 walkout / 倒地相两个永久口解除——无逐帧 glue-flap（漂进 HOLD 松压
// 会让慢漂 ghost 复活）。露倒地相 → 解压 + 锁 real（FN-safe，真人贴干扰源摔绝不被压）。
// 注：不复用 census 的 ghostSustainRun（:481-483，其被 SetTrackPReal 按 p_mirror 逐帧清零，孤迹 ghost
// ρ 低撑不住；且置 MaxGhostSustained 会改 lost_fall 删除判据=Stage 2/3），故 split 走本地棘轮。
func (tm *TrackManager) updateSplitGhost(ts *TrackState, f TrackFrame, nowMs int64) {
	if ts.SplitObservingSinceMs == 0 {
		return
	}
	if ts.SplitEverWalkedOut {
		return
	}
	if teleportFallPrecursorPose(f.Pose) {
		ts.SplitGhostSinceMs = 0
		ts.SplitEverWalkedOut = true
		return
	}
	if nowMs-ts.SplitLastLatchMs < splitLatchIntervalMs {
		return
	}
	ts.SplitLastLatchMs = nowMs

	d := distInt(f.X, f.Y, ts.SplitSpliterX, ts.SplitSpliterY)
	switch {
	case d >= splitWalkOutCm:
		ts.SplitEverWalkedOut = true
		ts.SplitGhostSinceMs = 0
		ts.splitGlueRun = 0
		tm.logger.Info("split_real_latched",
			zap.String("device_uid", f.DeviceAddr),
			zap.Int("track_id", ts.TrackID),
			zap.String("logic_id", ts.LogicID),
			zap.Int("anchor_dist_cm", d),
		)
	case d <= splitGluedCm:
		ts.splitGlueRun++
		if ts.splitGlueRun >= splitConvictTicks && ts.SplitGhostSinceMs == 0 {
			ts.SplitGhostSinceMs = nowMs
			tm.logger.Info("split_ghost_convicted",
				zap.String("device_uid", f.DeviceAddr),
				zap.Int("track_id", ts.TrackID),
				zap.String("logic_id", ts.LogicID),
				zap.Int("anchor_dist_cm", d),
			)
		}
	default:
		ts.splitGlueRun = 0 // HOLD：断棘轮（需重新 3 连坐实）；既有定罪 SplitGhostSinceMs 不清
	}
}

// splitSuccessionHardGate succession 的两道瞬时硬门槛（另两道 InBed/ExitRoom 走门第/床第，已由 evalProvenance
// 并入 RealProven，故此处只判 walkout 与 2-tick liveness）。跨任一 → 候选确认 real、latch 终止重选。
func (tm *TrackManager) splitSuccessionHardGate(ts *TrackState) bool {
	if ts.SplitEverWalkedOut { // R≥200 离锚点净走出
		return true
	}
	x2, y2, ok2 := histPos(ts, 2)
	x0, y0, ok0 := histPos(ts, 0)
	return ok2 && ok0 && distInt(x2, y2, x0, y0) > splitLiveMoveCm // 2-tick 净走动（两 tick 历史齐备才判，不齐=不可判维持暂持）
}

// runSplitSuccession split 选错的回退（§6.12）：占用 mantle 持有者=spliter **silent lost**（非 ExitRoom 走）时，
// 在 living split-group 成员里按 PReal 逐帧重选 real（HSRP 热备顶上）。收敛=跨硬门槛/living 剩 1 → RealProven latch。
// FN-safe「只加不减」：succession-active 组内全体候选解 SplitGhost 软压 → floor 网覆盖每条（坐着的真人摔倒不漏，
// 不依赖任何人 walk-out）。每帧调（processFrameAt 段2 之后，持 tm.mu）。
func (tm *TrackManager) runSplitSuccession(nowMs int64) {
	// lid → 该身份最近一次真实观测 ms（判 spliter/候选是否 living：coast 窗内=在场）
	lastObsByLID := map[string]int64{}
	for _, ts := range tm.tracks {
		if ts.LastObservedMs > lastObsByLID[ts.LogicID] {
			lastObsByLID[ts.LogicID] = ts.LastObservedMs
		}
	}
	// living 组员（自身未过 coast）按 spliter 分组
	membersBySpliter := map[string][]*TrackState{}
	for _, ts := range tm.tracks {
		if ts.SplitSpliterLID == "" || nowMs-ts.LastObservedMs >= presenceCoastMs {
			continue
		}
		membersBySpliter[ts.SplitSpliterLID] = append(membersBySpliter[ts.SplitSpliterLID], ts)
	}
	for spliterLID, members := range membersBySpliter {
		if last, ok := lastObsByLID[spliterLID]; ok && nowMs-last < presenceCoastMs {
			continue // spliter 仍在场（coast 未过窗）→ mantle 未失，不触发
		}
		if _, exited := tm.hardExitedLIDs[spliterLID]; exited {
			continue // S1：spliter 经 ExitRoom 离房 = 真人走了，不重选（残迹交 tFloor）
		}
		alreadyReal := false // 组内已有 latch real（walkout/门第/床）→ mantle 已定
		for _, m := range members {
			if m.RealProven {
				alreadyReal = true
			}
		}
		if alreadyReal {
			for _, m := range members {
				m.SplitProvisionalReal = false
			}
			continue
		}
		// S3：spliter silent-lost + living 候选。只加不减 → 全体解软压保 floor 网
		for _, m := range members {
			m.SplitGhostSinceMs = 0
			m.splitGlueRun = 0
		}
		latched := false
		for _, m := range members {
			if tm.splitSuccessionHardGate(m) {
				tm.promoteSuccession(m, members, nowMs, "hard_gate")
				latched = true
				break
			}
		}
		if latched {
			continue
		}
		if len(members) == 1 { // 排除法：living 剩 1 → 末者即 real
			tm.promoteSuccession(members[0], members, nowMs, "sole_living")
			continue
		}
		best := members[0] // 暂持：PReal 最高者顶 mantle（逐帧可撤销，非 latch）
		for _, m := range members[1:] {
			if m.PReal > best.PReal {
				best = m
			}
		}
		for _, m := range members {
			m.SplitProvisionalReal = (m == best)
		}
	}
}

// promoteSuccession 收敛出口：winner latch RealProven（census 唯一真化通道）+ 解自身软压；清全组暂持位。
func (tm *TrackManager) promoteSuccession(winner *TrackState, group []*TrackState, nowMs int64, reason string) {
	winner.RealProven = true
	winner.SplitGhostSinceMs = 0
	for _, m := range group {
		m.SplitProvisionalReal = false
	}
	tm.logger.Info("split_succession_promote",
		zap.String("device_uid", winner.DeviceAddr),
		zap.Int("track_id", winner.TrackID),
		zap.String("logic_id", winner.LogicID),
		zap.String("spliter_lid", winner.SplitSpliterLID),
		zap.String("reason", reason),
		zap.Float64("p_real", winner.PReal),
	)
}
