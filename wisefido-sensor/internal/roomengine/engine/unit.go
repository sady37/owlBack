package engine

import (
	"sort"
	"strconv"

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
	roomID  string
	tMs     int64
	slow    bool // 现身=sleepad InBed（走+躺慢接力）→ 用 sleepad 慢窗（W_sleepad 150s）
	claimed bool // 守恒配对已被某离开候选认领（消费出池）→ 不再供别的 lost 借
}

// nrSample 一房某时刻的 n_r 采样（守恒 quota 基线：查"最早离开锚−ε 时刻"邻居真人数）。
type nrSample struct {
	tMs int64
	nr  int
}

// depKind 离开候选类型（配对优先级：ExitFirst 先于 LostFirst）。
type depKind int

const (
	depExitFirst depKind = iota // 门口离开(ExitRoom 结账)
	depLostFirst                // 冻结/丢轨(LostFirst 锚)
)

// depRec 一条本房离开候选（PresentCount>0→0 前累积，空房触发时与 gains 守恒配对）。
type depRec struct {
	kind     depKind
	anchorMs int64
	logicID  string
}

// handoffEpsilonMs 配对方向性负容差：arrival 可早于离开锚至多此值（门槛过门两雷达 ±1~2 tick 同步噪声）。
const handoffEpsilonMs = 2000

// Unit 一个居住单元的多房编排器。
type Unit struct {
	rooms         map[string]*Room
	np            belief.NeighborParams
	residentCount int                   // 多住户 hard-gate（==1 才注入 hand-off；承重，§7.7 v2）
	gains         []gainEvent           // 近期跨房新现真人 gain event（churn 会多记同人；配对时按每房净到达 quota 卡上限）
	roomNr        map[string]int        // 每房当前 n_r（守恒 quota 现值）
	matchDbg      map[string]string     // forensic：末次空房配对的 quota/purged 摘要（xray 观测）
	nrHist        map[string][]nrSample // 每房 n_r 历史（查"最早离开锚−ε 时刻"的基线；窗内 pruned）
	departures    map[string][]depRec   // 每房离开候选台账（ExitFirst+LostFirst，PresentCount>0→0 前累积，空房触发配对后清）
	lostAt        map[string]int64      // 每房最近丢真人的 loss 锚（冻结→LostFirst still-box 起点，未冻结→last-observed；非 coast LostReal 帧；centering）
	lastHandoffL  map[string]float64    // 末次喂各房的 hand-off 注入对数似然（forensic / 测试可观测）
	roomPresent   map[string]bool       // roomID → 末 tick 是否在场（PresentCount>0；UnitState forensic）
	roomTickMs    map[string]int64      // roomID → 末 tick 时戳（新鲜度）
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
		roomNr:        map[string]int{},
		matchDbg:      map[string]string{},
		nrHist:        map[string][]nrSample{},
		departures:    map[string][]depRec{},
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
	wasPresent := u.roomPresent[roomID]
	u.roomPresent[roomID] = fr.PresentCount > 0
	u.roomTickMs[roomID] = nowMs

	// 每房当前 n_r + 历史采样（守恒 quota：churn 不改 n_r，是"净到达"的抗 churn 权威）。
	u.roomNr[roomID] = fr.Decision.PeopleCount
	u.nrHist[roomID] = append(u.nrHist[roomID], nrSample{tMs: nowMs, nr: fr.Decision.PeopleCount})

	// gain event 记账（churn 会为同人多 lid 多记；配对时按每房净到达 quota 卡上限，见 matchAndPurge）。
	if fr.GainedReal > 0 {
		u.gains = append(u.gains, gainEvent{roomID: roomID, tMs: nowMs, slow: fr.GainedFromSleepad})
		delete(u.lostAt, roomID) // 本房又现真人 = 丢的人回来了/新人到 → 清本房待解析(旧 handoffLFor 路径)
	}
	if fr.LostReal {
		u.lostAt[roomID] = fr.LostAtMs // loss 锚(冻结→LostFirst / 未冻结→last-observed)：Δ centering，gain−lostAt 回正区间
	}

	// 守恒配对台账：本房离开候选（ExitFirst/LostFirst）累积，空房触发时与 gains 一对一认领消费。
	for _, e := range fr.ExitFirsts {
		u.departures[roomID] = append(u.departures[roomID], depRec{kind: depExitFirst, anchorMs: e.ExitFirstMs, logicID: e.LogicID})
	}
	for _, ld := range fr.LostDepartures {
		u.departures[roomID] = append(u.departures[roomID], depRec{kind: depLostFirst, anchorMs: ld.LostFirstMs, logicID: ld.LogicID})
	}
	// 查询门 = 本 Room PresentCount 跳 0（雷达空房信号）：此刻对累积的离开候选与 sibling gains 做守恒配对；
	//   matched LostFirst 残迹 → 硬 purge（drop + 回传 tm evict）；未 matched → 不 purge（保 floor 兜底，FN-safe）。
	if wasPresent && fr.PresentCount == 0 {
		for _, id := range u.matchAndPurge(roomID) {
			u.rooms[roomID].dropTrack(id)
			fr.DroppedLogicIDs = append(fr.DroppedLogicIDs, id)
		}
		u.departures[roomID] = nil // 本轮空房已配对，清账（下次进人重新累积）
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

// matchAndPurge 本房空房那刻的守恒配对：sibling gains 一对一认领本房离开候选，认领即消费(claimed 出池)。
//
//	方向性 + ε 容差：只配 arrival ≥ 锚−ε 且 ≤ W(先离后现,守因果;ε 吸收门槛过门两雷达同步噪声)。
//	优先级：ExitFirst 先于 LostFirst，同类 |Δ| 小先(挑最贴的过门配对)。
//	返回 matched 的 LostFirst logicID（有守恒宿主的冻结残迹 → 调用方 purge）；ExitFirst 已走 ExitRoom 自离场，仅消费其 arrival。
//	多住户 hard-gate(residentCount!=1)保守不配对（承重 FN，同 HandoffLogOdds）。
//
// nrAt 房 rid 在 ≤tMs 最近一次采样的 n_r（quota 基线）。tMs 早于全部历史→取最早采样(保守，起点即已在场者)；
//
//	无历史→回退当前 n_r（→quota 0，不信用无据到达，FN-safe）。
func (u *Unit) nrAt(rid string, tMs int64) int {
	h := u.nrHist[rid]
	if len(h) == 0 {
		return u.roomNr[rid]
	}
	base := h[0].nr
	for _, s := range h {
		if s.tMs > tMs {
			break
		}
		base = s.nr
	}
	return base
}

func (u *Unit) matchAndPurge(roomID string) []string {
	if u.residentCount != 1 {
		return nil
	}
	deps := u.departures[roomID]
	if len(deps) == 0 {
		return nil
	}
	// 每房净到达 quota = 邻居 n_r(现在) − n_r(最早离开锚−ε 时刻)：抗 churn 上限，churn 多记的 gain 只算 quota 个真到达。
	earliest := deps[0].anchorMs
	for _, d := range deps {
		if d.anchorMs < earliest {
			earliest = d.anchorMs
		}
	}
	quota := map[string]int{}
	for rid, nrNow := range u.roomNr {
		if rid == roomID {
			continue
		}
		if q := nrNow - u.nrAt(rid, earliest-handoffEpsilonMs); q > 0 {
			quota[rid] = q
		}
	}
	dbg := "quota{"
	for rid, q := range quota {
		dbg += rid[len(rid)-4:] + ":" + strconv.Itoa(q) + " "
	}
	dbg += "} deps:" + strconv.Itoa(len(deps))
	type cand struct {
		gi, di int
		absD   int64
		isExit bool
	}
	var cands []cand
	for gi := range u.gains {
		g := &u.gains[gi]
		if g.claimed || g.roomID == roomID {
			continue // 已被认领 / 同房新现(人没穿房)不算宿
		}
		w := u.np.HandoffWindowMs
		if g.slow {
			w = u.np.SleepadWindowMs
		}
		for di := range deps {
			delta := g.tMs - deps[di].anchorMs
			if delta < -handoffEpsilonMs || delta > w {
				continue // 方向性+ε：arrival 早于锚超 ε / 超窗上界 → 非本次接力
			}
			ad := delta
			if ad < 0 {
				ad = -ad
			}
			cands = append(cands, cand{gi, di, ad, deps[di].kind == depExitFirst})
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].isExit != cands[j].isExit {
			return cands[i].isExit // ExitFirst 先
		}
		return cands[i].absD < cands[j].absD // 再 |Δ| 小
	})
	claimedDep := make([]bool, len(deps))
	roomClaims := map[string]int{}
	var purged []string
	for _, c := range cands {
		rid := u.gains[c.gi].roomID
		if u.gains[c.gi].claimed || claimedDep[c.di] || roomClaims[rid] >= quota[rid] {
			continue // gain 已认领 / dep 已配 / 该邻居房净到达名额已用光(churn 多记的重复 gain 挡在此)
		}
		u.gains[c.gi].claimed = true
		claimedDep[c.di] = true
		roomClaims[rid]++
		if deps[c.di].kind == depLostFirst {
			purged = append(purged, deps[c.di].logicID)
		}
	}
	u.matchDbg[roomID] = dbg + " purged:" + strconv.Itoa(len(purged))
	return purged
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

// MatchDbg 末次空房守恒配对摘要（quota/deps/purged）；forensic。
func (u *Unit) MatchDbg(roomID string) string { return u.matchDbg[roomID] }

// UnitState forensic（unit 是否有在场 track / 是否有 neighbor）。
func (u *Unit) UnitState(nowMs int64) (unitHasTrack, hasNeighbor bool) {
	return u.unitHasTrack(nowMs), len(u.rooms) > 1
}

// handoffLFor 房 roomID 的 hand-off SLeft 注入对数似然：若它有待解析丢失，找**兄弟房**窗内新现真人 → 矩形核。
//
//	同房新现不算 hand-off（人没穿房）；Δ=gain−loss 锚(冻结→LostFirst / 未冻结→last-observed，centering，先离后现)；窗外/多住户由 HandoffLogOdds 归零。
func (u *Unit) handoffLFor(roomID string, nowMs int64) float64 {
	lostMs, ok := u.lostAt[roomID]
	if !ok {
		return 0 // 本房没丢人 → 无 lost-fall 二义 → 不注入
	}
	var sibs []belief.SiblingHandoff
	for _, g := range u.gains {
		if g.roomID == roomID || g.claimed {
			continue // 兄弟房才算（守恒=丢的人去了别的房）；已被守恒配对认领的到达不再供旧注入路径借
		}
		sibs = append(sibs, belief.SiblingHandoff{
			ArrivalDeltaMs: g.tMs - lostMs, // >0 = 先离后现（lostMs=loss 锚:冻结→LostFirst / 未冻结→last-observed）
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
	// n_r 历史同窗剪（保留一条 keepMs 前的锚点供 nrAt 基线插值）。
	for rid, h := range u.nrHist {
		cut := 0
		for cut < len(h)-1 && nowMs-h[cut+1].tMs > keepMs {
			cut++
		}
		u.nrHist[rid] = h[cut:]
	}
}
