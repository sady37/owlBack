package engine

import (
	"wisefido-sensor/internal/roomengine/adapter"
	"wisefido-sensor/internal/roomengine/belief"
)

// unit.go — 多房 unit 编排器（§57 步4）。两层职责：
//  1. hand-off 注入对数似然（neighbor 守恒+矩形时间窗，§7.7 v2）→ 喂该房 Room.Tick → engine 层 SLeft 注入。
//  2. **UD timer 兜底**（lost-fall deadline）：lost 后若 unit 全空、且 belief 未自然到 0.85，到点补 SFall fire。
//     bathroom（room_type）→ D（抢救紧迫，短 20min）；其它 → UD（unit 级最长 90min=AreaSit 久坐容忍）。
//     **任何 track 现身 unit（任一房 present）→ 取消所有 timer**（有人在场=低风险，资源克制，一招统一
//     本人起身/他人到场/兄弟房现人）。**无 neighbor（单房 unit）→ 不设 timer**（克制）。
//     frozen 残留（track present，PresentCount>0）走 Room present 路径 FloorGuard，不进 UD；
//     完全丢轨（无 track）才进 UD timer。数值留 oracle（form-anchor [[fall_data_is_artificial_test]]）。
//
// 单房无兄弟 → 无 sibling gain → ρ≡0 + 无 neighbor 不设 UD（cd2b 单房零回归，与单房 Room 逐帧等价）。

const (
	presenceFreshMs = 30 * 1000 // 房在场新鲜度：超此未 tick 的房视作 stale→不计在场（FN-safe，不被 stale 误取消）
)

// gainEvent 一条跨房「新现真人」事件（窗内留账，守恒比对用）。
type gainEvent struct {
	roomID string
	tMs    int64
	slow   bool // 现身=sleepad InBed（走+躺慢接力）→ 用 sleepad 慢窗（W_sleepad 150s）
}

// Unit 一个居住单元的多房编排器。
type Unit struct {
	rooms         map[string]*Room
	np            belief.NeighborParams
	residentCount int                // 多住户 hard-gate（==1 才注入 hand-off；承重，§7.7 v2）
	gains         []gainEvent        // 近期跨房新现真人（窗内 pruned）
	lostAt        map[string]int64   // 每房最近丢真人的 **last-observed 时戳**（非 coast LostReal 帧；centering）
	lastHandoffL  map[string]float64 // 末次喂各房的 hand-off 注入对数似然（forensic / 测试可观测）
	roomPresent   map[string]bool    // roomID → 末 tick 是否在场（PresentCount>0；UnitState forensic）
	roomTickMs    map[string]int64   // roomID → 末 tick 时戳（新鲜度）
}

// NewUnit 建多房编排器。residentCount = unit 住户数（η 用）。
//
//	radar hand-off 核（τpeak=15s/W=90s）不随公共度缩放（DefaultNeighborParams 固定）：elder-care 套房无陌生人、
//	多住户已 hard-gate、单住户无巧合 → 宽窗零 FN 更优，公共度收紧无意义（详 neighbor.go 注释）。
func NewUnit(rooms map[string]*Room, residentCount int) *Unit {
	if residentCount < 1 {
		residentCount = 1
	}
	np := belief.DefaultNeighborParams()
	return &Unit{
		rooms:         rooms,
		np:            np,
		residentCount: residentCount,
		lostAt:        map[string]int64{},
		lastHandoffL:  map[string]float64{},
		roomPresent:   map[string]bool{},
		roomTickMs:    map[string]int64{},
	}
}

// Tick 一个设备帧（roomID 设备所属房）到来：算 ρ_xroom → Room.Tick → 更新 handoff 账本 + UD timer 兜底。
func (u *Unit) Tick(roomID string, fi adapter.FrameInput) Frame {
	handoffL := u.handoffLFor(roomID, fi.NowMs)
	u.lastHandoffL[roomID] = handoffL
	fr := u.rooms[roomID].Tick(fi, handoffL)
	nowMs := fi.NowMs

	// 本房在场：PresentCount>0 = 有新鲜观测 track；frozen 残留 Present=false 不计（走 present 路径前已计）。
	u.roomPresent[roomID] = fr.PresentCount > 0
	u.roomTickMs[roomID] = nowMs

	// ρ_xroom 账本（sibling-handoff，belief 塑形层）。
	if fr.GainedReal > 0 {
		u.gains = append(u.gains, gainEvent{roomID, nowMs, fr.GainedFromSleepad})
		delete(u.lostAt, roomID) // 本房又现真人 = 丢的人回来了/新人到 → 清本房待解析
	}
	if fr.LostReal {
		u.lostAt[roomID] = fr.LostAtMs // **last-observed 时戳**（非 coast LostReal 帧）：Δ centering，gain−lostAt 回正区间
	}

	// floor（stillbox 计时器，engine.go per-track，StillSec≥tFloor→保底发 SFallen）= 唯一时长兜底。
	//   单房 unit（len(rooms)==1）：无邻房印证、FP 风险剧增 → **不兜底**（belief ≥0.85 抢发仍报久躺强证据，§I/#1）。
	//   hand-off 抑制不再在此（旧 rho>0 砍兜底=瞬时压制，会绕过 latch）→ 改 engine 层 SLeft 注入（latch-first）：
	//     注入抬 exitL≥flip → engine floor 闸（exitL<flip）已挡 + durable purge 撤轨；已证摔 latch 不注入 → 照兜底。
	if fr.Decision.Band == "floor" && len(u.rooms) == 1 {
		fr.Decision.Fire = false
		fr.Decision.Band = "no"
	}

	u.pruneGains(nowMs)
	return fr
}

// unitHasTrack unit 内任一房末 tick 在场（且新鲜）。frozen 残留不计（Present=false）。
func (u *Unit) unitHasTrack(nowMs int64) bool {
	for id, present := range u.roomPresent {
		if present && nowMs-u.roomTickMs[id] < presenceFreshMs {
			return true
		}
	}
	return false
}

// LastHandoffL 末次喂房 roomID 的 hand-off 注入对数似然（forensic / 测试）。
func (u *Unit) LastHandoffL(roomID string) float64 { return u.lastHandoffL[roomID] }

// PendingLostMs 本房待 hand-off 解析的丢轨时戳（0=无待解析 lost；>0=已注册 lost，handoffLFor 在跑找接力）。forensic。
func (u *Unit) PendingLostMs(roomID string) int64 { return u.lostAt[roomID] }

// SiblingGainCount unit 内当前窗口存活的跨房新现真人 gain 数（hand-off 宿候选；0=无人在别房现身）。forensic。
func (u *Unit) SiblingGainCount() int { return len(u.gains) }

// UnitState forensic（unit 是否有在场 track / 是否有 neighbor）。
func (u *Unit) UnitState(nowMs int64) (unitHasTrack, hasNeighbor bool) {
	return u.unitHasTrack(nowMs), len(u.rooms) > 1
}

// handoffLFor 房 roomID 的 hand-off SLeft 注入对数似然：若它有待解析丢失，找**兄弟房**窗内新现真人 → 矩形核。
//
//	同房新现不算 hand-off（人没穿房）；Δ=gain−last_observed（centering，先离后现）；窗外/多住户由 HandoffLogOdds 归零。
func (u *Unit) handoffLFor(roomID string, nowMs int64) float64 {
	lostMs, ok := u.lostAt[roomID]
	if !ok {
		return 0 // 本房没丢人 → 无 lost-fall 二义 → 不注入
	}
	var sibs []belief.SiblingHandoff
	for _, g := range u.gains {
		if g.roomID == roomID {
			continue // 兄弟房才算（守恒=丢的人去了别的房）
		}
		sibs = append(sibs, belief.SiblingHandoff{
			ArrivalDeltaMs: g.tMs - lostMs, // >0 = 先离后现（lostMs=last-observed centering）
			Slow:           g.slow,         // sleepad InBed 现身 → 慢窗（走+躺，用 W_sleepad）
		})
	}
	return belief.HandoffLogOdds(u.np, u.residentCount, sibs)
}

// pruneGains 丢弃超 hand-off 窗的旧 gain（stale 证不了此刻在哪，[[partial_monitoring_fall_suppression_law]]）。
//
//	保留窗取 radar/sleepad 两核较大者（sleepad 慢核窗 150s > radar W）：sleepad gain 须存活到慢核窗满，
//	否则被 radar 窗（90s）先剪掉 → 慢接力还没解析就丢账。per-gain 是否真在窗内由 HandoffLogOdds 矩形窗各自判。
func (u *Unit) pruneGains(nowMs int64) {
	keepMs := u.np.HandoffWindowMs
	if u.np.SleepadWindowMs > keepMs {
		keepMs = u.np.SleepadWindowMs
	}
	kept := u.gains[:0]
	for _, g := range u.gains {
		if nowMs-g.tMs <= keepMs {
			kept = append(kept, g)
		}
	}
	u.gains = kept
}
