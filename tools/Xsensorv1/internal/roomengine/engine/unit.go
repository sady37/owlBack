package engine

import (
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// unit.go — 多房 unit 编排器（§57 步4，纯时间窗 neighbor，接线非攻坚——§A 数学已在 belief/neighbor.go）。
// 驱动 N 个 Room；跨设备「丢↔现」守恒（规则②）+ 时间窗（band-pass wDir，规则③）→ ρ_xroom → 喂该房 Room.Tick。
//   §A.2 GateBlindRow 据 ρ 把该房 Blind→Fallen 整流入 Left（人挪去兄弟房 vs 本房真摔，联合推断）。
// **不碰房间相邻/空间路由**（§57 作废分支）：任一兄弟设备时间比对，不看哪房挨哪房。
// 单房无兄弟 → 无 sibling gain → ρ≡0（cd2b 零回归，与单房 Room 逐帧等价）。
//
// hand-off 权威信号 = 守恒（丢的人在兄弟房重现），非裸「兄弟房有人」：源房 LostReal@t_lost ↔
//   兄弟房 GainedReal@t_gain，Δ=t_gain−t_lost 经 band-pass wDir、宿房去 ghost 后验 GainedReal 作权。

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
	}
}

// Tick 一个设备帧（roomID 设备所属房）到来：算该房 ρ_xroom（兄弟房守恒+时间窗）→ Room.Tick → 更新账本。
func (u *Unit) Tick(roomID string, fi adapter.FrameInput) Frame {
	rho := u.rhoFor(roomID, fi.NowMs)
	u.lastRho[roomID] = rho
	fr := u.rooms[roomID].Tick(fi, rho)

	if fr.GainedReal > 0 {
		u.gains = append(u.gains, gainEvent{roomID, fi.NowMs, fr.GainedReal})
		delete(u.lostAt, roomID) // 本房又现真人 = 丢的人回来了/新人到 → 清本房待解析（非穿房）
	}
	if fr.LostReal {
		u.lostAt[roomID] = fi.NowMs
	}
	u.pruneGains(fi.NowMs)
	return fr
}

// LastRho 末次喂房 roomID 的 ρ_xroom（forensic / 测试）。
func (u *Unit) LastRho(roomID string) float64 { return u.lastRho[roomID] }

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
