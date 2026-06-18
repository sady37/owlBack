package engine

import (
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// unit.go — 多房 unit 编排器（§57 步4）。两层职责：
//  1. ρ_xroom（neighbor handoff 守恒+时间窗）→ 喂该房 Room.Tick，belief 塑形（GateBlindRow Blind→Left）。
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
	pReal  float64
}

// Unit 一个居住单元的多房编排器。
type Unit struct {
	rooms         map[string]*Room
	np            belief.NeighborParams
	residentCount int                // η sole-resident 衰减（单住户=1）
	cAttr         float64            // 源型可信度：雷达新现=room-enter 事件（§A 默认 0.8，form-anchor 留 oracle）
	gains         []gainEvent        // 近期跨房新现真人（窗内 pruned）
	lostAt        map[string]int64   // 每房最近丢真人时戳（待 hand-off 解析；再现/解析后清）
	lastRho       map[string]float64 // 末次喂各房的 ρ（forensic / 测试可观测）
	roomPresent map[string]bool  // roomID → 末 tick 是否在场（PresentCount>0；UnitState forensic）
	roomTickMs  map[string]int64 // roomID → 末 tick 时戳（新鲜度）
}

// NewUnit 建多房编排器。residentCount = unit 住户数（η 用）；pub = unit 公共度（规则④ 定找人窗 W）。
func NewUnit(rooms map[string]*Room, residentCount int, pub belief.UnitPublicness) *Unit {
	if residentCount < 1 {
		residentCount = 1
	}
	np := belief.DefaultNeighborParams()
	np.HandoffWindowMs = belief.HandoffWindowFor(pub) // 规则④：public 45s / share 60s / private 90s
	return &Unit{
		rooms:         rooms,
		np:            np,
		residentCount: residentCount,
		cAttr:         0.8, // room-enter（雷达新现 track = 有人进房）
		lostAt:        map[string]int64{},
		lastRho:       map[string]float64{},
		roomPresent:   map[string]bool{},
		roomTickMs:    map[string]int64{},
	}
}

// Tick 一个设备帧（roomID 设备所属房）到来：算 ρ_xroom → Room.Tick → 更新 handoff 账本 + UD timer 兜底。
func (u *Unit) Tick(roomID string, fi adapter.FrameInput) Frame {
	rho := u.rhoFor(roomID, fi.NowMs)
	u.lastRho[roomID] = rho
	fr := u.rooms[roomID].Tick(fi, rho)
	nowMs := fi.NowMs

	// 本房在场：PresentCount>0 = 有新鲜观测 track；frozen 残留 Present=false 不计（走 present 路径前已计）。
	u.roomPresent[roomID] = fr.PresentCount > 0
	u.roomTickMs[roomID] = nowMs

	// ρ_xroom 账本（sibling-handoff，belief 塑形层）。
	if fr.GainedReal > 0 {
		u.gains = append(u.gains, gainEvent{roomID, nowMs, fr.GainedReal})
		delete(u.lostAt, roomID) // 本房又现真人 = 丢的人回来了/新人到 → 清本房待解析
	}
	if fr.LostReal {
		u.lostAt[roomID] = nowMs
	}

	// §I 合体：floor（stillbox 计时器，engine.go per-track，StillSec≥tFloor→保底发 SFallen）= 唯一时长兜底；
	//   旧 D/DU 决断窗退役被吸收（floor 用实际 StillSec，比 lostMs 倒计时统一——present 久静 + lost 续算同源）。
	//
	// (b) **belief 抢发照常、只砍兜底**：band=lost/report（belief ≥0.85 抢发）不动；只对 floor 兜底腿做资源克制：
	//   - **单房 unit**（len(rooms)==1）：资源少、无邻房印证、FP 风险剧增 → **不兜底**（belief 抢发仍报久躺强证据，§I/#1）；
	//   - **hand-off**（ρ>0，本人现身隔壁，W 窗内）：人没在本房摔 → 不兜底。
	//   其余取消已在 engine.go floor 块内：StillSec==0 → StillSec<tFloor 自然不 fire；exitL≥flip → floor 条件已挡。
	if fr.Decision.Band == "floor" && (len(u.rooms) == 1 || rho > 0) {
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

// LastRho 末次喂房 roomID 的 ρ_xroom（forensic / 测试）。
func (u *Unit) LastRho(roomID string) float64 { return u.lastRho[roomID] }

// UnitState forensic（unit 是否有在场 track / 是否有 neighbor）。
func (u *Unit) UnitState(nowMs int64) (unitHasTrack, hasNeighbor bool) {
	return u.unitHasTrack(nowMs), len(u.rooms) > 1
}

// rhoFor 房 roomID 的 ρ_xroom：若它有待解析的丢失，找**兄弟房**窗内新现真人 → SiblingHandoff → RhoXroom。
//
//	同房新现不算 hand-off（人没穿房）；窗外/反向由 wDir 归零（§A.1）。
func (u *Unit) rhoFor(roomID string, nowMs int64) float64 {
	lostMs, ok := u.lostAt[roomID]
	if !ok {
		return 0 // 本房没丢人 → 无 lost-fall 二义 → 不抑制
	}
	var sibs []belief.SiblingHandoff
	for _, g := range u.gains {
		if g.roomID == roomID {
			continue // 兄弟房才算（守恒=丢的人去了别的房）
		}
		sibs = append(sibs, belief.SiblingHandoff{
			ArrivalDeltaMs: g.tMs - lostMs, // >0 = 先丢后现（有向，§A 区别 ghost 对称）
			CAttr:          u.cAttr,
			GainedReal:     g.pReal,
		})
	}
	return belief.RhoXroom(u.np, u.residentCount, sibs)
}

// pruneGains 丢弃超 hand-off 窗的旧 gain（stale 证不了此刻在哪，[[partial_monitoring_fall_suppression_law]]）。
func (u *Unit) pruneGains(nowMs int64) {
	kept := u.gains[:0]
	for _, g := range u.gains {
		if nowMs-g.tMs <= u.np.HandoffWindowMs {
			kept = append(kept, g)
		}
	}
	u.gains = kept
}
