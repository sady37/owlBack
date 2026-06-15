package adapter

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// realness.go — lean-extract 出生档案原语（DBN-Zone-Room §G）：跨帧维护单房单 track 出生位，产
// belief.RealnessObs 的 Displaced（偏离出生位 = 走动/跌倒位移 → 自主走动锁 Real）。
// co-existence 字段（CoexistRho/IsReflection，mirror）由 MM 邻居在 W3.4b 填，本层留零值中性。
// §G 删 Static：still-box / AgeLongStatic 原语随 Static 类删——孤立纯金属 firmware 30s 已过滤，
//   本层不分静止金属；且静止判 ghost = 病根（分不开金属 vs 摔倒静止的人）。

// RealnessParams 出生档案形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type RealnessParams struct {
	MoveCm int // 偏离出生位 > 此 = Displaced（走动/跌倒位移，雷达 XY 噪声地板之上）
}

// DefaultRealnessParams 形态默认（非权威值，留 oracle）。
func DefaultRealnessParams() RealnessParams { return RealnessParams{MoveCm: 50} }

// TrackArchive 单 track 出生位（跨帧）。单房单 track 一次出生（首 Online 帧），**绝不**因雷达断流/稀疏重生
//
//	——track 消失正是「摔倒静止降功率」最危险态，出生位须跨消失存活（realness.go 核心物理）。
type TrackArchive struct {
	p      RealnessParams
	born   bool
	birthX int
	birthY int
}

// NewTrackArchive 建空档案（出生待首帧 Observe 填）。
func NewTrackArchive(p RealnessParams) *TrackArchive { return &TrackArchive{p: p} }

// Observe 一帧 → belief.RealnessObs。仅 Online 帧调用（track absent = 不分类，realness 后验由 caller
// 持住）。返回 birth=true 仅首 Online 帧 → caller 据此起 RealnessTrack。
func (a *TrackArchive) Observe(t RadarTrack) (belief.RealnessObs, bool) {
	birth := false
	if !a.born {
		a.born = true
		a.birthX, a.birthY = t.X, t.Y
		birth = true
	}
	disp := math.Hypot(float64(t.X-a.birthX), float64(t.Y-a.birthY))
	// CoexistRho/IsReflection：MM 拓扑，W3.4b 填（本轮零值中性 → 单房 PReal≡1 不压）。
	return belief.RealnessObs{Displaced: disp > float64(a.p.MoveCm)}, birth
}
