package engine

import (
	"owl-common/card"

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
	udBathroomMs    = 20 * 60 * 1000 // bathroom D：20min（抢救紧迫，卫浴信号易丢+高危+私密无人看=唯一安全网）
	udDefaultMs     = 90 * 60 * 1000 // 其它 UD：90min（unit 级最长，lost 不知人在哪区用最宽容；= AreaSit 久坐容忍）
	presenceFreshMs = 30 * 1000      // 房在场新鲜度：超此未 tick 的房视作 stale→不计在场（FN-safe，不被 stale 误取消）
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
	// UD timer 兜底
	roomType    map[string]int   // roomID → card.RoomType（1=Bathroom）→ per-room deadline
	udDeadline  map[string]int64 // roomID → 活动 UD timer 到点 ms（0/缺=无）
	roomPresent map[string]bool  // roomID → 末 tick 是否在场（PresentCount>0）
	roomTickMs  map[string]int64 // roomID → 末 tick 时戳（新鲜度）
	udMul       float64          // timer 时长乘子（验证旋钮 XSENSOR_UD_MUL，默认 1；调小可短 case 验机制）
}

// NewUnit 建多房编排器。residentCount = unit 住户数（η 用）；pub = unit 公共度（规则④ 定找人窗 W）；
// roomType = 各房 card.RoomType（UD timer 判 bathroom）；udMul = timer 时长乘子（验证旋钮，≤0 取 1）。
func NewUnit(rooms map[string]*Room, residentCount int, pub belief.UnitPublicness, roomType map[string]int, udMul float64) *Unit {
	if residentCount < 1 {
		residentCount = 1
	}
	if udMul <= 0 {
		udMul = 1
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
		roomType:      roomType,
		udDeadline:    map[string]int64{},
		roomPresent:   map[string]bool{},
		roomTickMs:    map[string]int64{},
		udMul:         udMul,
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

	// D/UD timer 兜底（per lost-room；各房在各自 tick 处理自己的 timer）。**取消条件按房型分**（架构师拍）：
	//   - bathroom 走 D：私密关门，邻房有人也看不见里面 → **只认本人交代** = recovery（雷达重抓到=本房 present，
	//     正常如厕会动→不停重抓→到不了 20min）OR handoff（本人现身隔壁，rho>0，neighbor W=45-90s）。**不认别房有人**。
	//   - 其它走 UD：开放空间 → unit 内**任何 track** 都取消（recovery/handoff/backup 都算）。
	// 无 neighbor（单房 unit）→ 不设 timer（克制）。
	if len(u.rooms) > 1 {
		isBath := u.roomType[roomID] == card.RoomTypeBathroom
		var cancelled bool
		if isBath {
			cancelled = u.roomPresent[roomID] || rho > 0 // D：recovery 或 handoff
		} else {
			cancelled = u.unitHasTrack(nowMs) // UD：任何 track
		}
		switch {
		case cancelled:
			delete(u.udDeadline, roomID)
		default:
			if fr.LostReal && u.udDeadline[roomID] == 0 {
				u.udDeadline[roomID] = nowMs + u.udLenFor(roomID)
			}
			if dl := u.udDeadline[roomID]; dl > 0 && nowMs >= dl && !fr.Decision.Fire {
				fr.Decision.Fire = true
				if isBath {
					fr.Decision.Band = "d_lost"
				} else {
					fr.Decision.Band = "ud_lost"
				}
				delete(u.udDeadline, roomID)
			}
		}
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

// udLenFor 本房 UD timer 时长：bathroom→D(20min)，其它→UD(90min)；×udMul（验证旋钮）。
func (u *Unit) udLenFor(roomID string) int64 {
	base := int64(udDefaultMs)
	if u.roomType[roomID] == card.RoomTypeBathroom {
		base = udBathroomMs
	}
	return int64(float64(base) * u.udMul)
}

// LastRho 末次喂房 roomID 的 ρ_xroom（forensic / 测试）。
func (u *Unit) LastRho(roomID string) float64 { return u.lastRho[roomID] }

// UDState UD timer forensic（roomID 的 timer 到点 ms[0=无活动 timer] / unit 是否有在场 track / 是否有 neighbor）。
func (u *Unit) UDState(roomID string, nowMs int64) (deadlineMs int64, unitHasTrack, hasNeighbor bool) {
	return u.udDeadline[roomID], u.unitHasTrack(nowMs), len(u.rooms) > 1
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
