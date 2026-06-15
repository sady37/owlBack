package adapter

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// realness.go — lean-extract 出生档案 + still-box 纯几何/时间原语（C §44 边界：提干净自包含函数，
// 非 copy-paste-modify Tsensor track_manager 2442 行 alarm/card/gate-list tangle）。跨帧维护单房单
// track 的出生位/时 + still-box run，产 belief.RealnessObs 的房内可源字段
// （Displaced / AgeLongStatic / CrossedStillPeriod）。
// Mirror/Static 的拓扑字段（CoexistRho/IsReflection/ConfinedNearWall）由 MM 邻居设备/反射体在 W3.4b
// neighbor 接通时填（[[mm_relationship_matrix]]：邻居 room/device 经 MM 拓扑廉价可知），本层留零值中性。

// RealnessParams 出生档案 + still-box 形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type RealnessParams struct {
	MoveCm        int   // 偏离出生位 > 此 = Displaced（走动/跌倒位移，雷达 XY 噪声地板之上）
	StillBoxCm    int   // still-box 滚动窗内 per-axis 包络 ≤ 此 = 静止 run
	StillWindowMs int64 // still-box 滚动窗物理长度
	StillPeriodMs int64 // 静止 run 持续 ≥ 此 = 跨过一个静止降功率期（survival real 慢证据）
	LongStaticMs  int64 // age ≥ 此且仍困出生点 = AgeLongStatic（static 签名②）
}

// DefaultRealnessParams 形态默认（非权威值，留 oracle）。
func DefaultRealnessParams() RealnessParams {
	return RealnessParams{
		MoveCm:     50,
		StillBoxCm: 50, StillWindowMs: 30_000, StillPeriodMs: 60_000,
		LongStaticMs: 300_000,
	}
}

type histPt struct {
	tMs  int64
	x, y int
}

// TrackArchive 单 track 出生档案 + still-box run（跨帧状态）。Room 持有，逐帧 Observe。
// 单房单 track 一次出生（首 Online 帧），**绝不**因雷达断流/稀疏重生——track 消失正是「摔倒静止降功率」
// 最危险态，realness 信心必须跨消失存活（realness.go 核心物理）；track 再识别（区分返回的人 vs 新人）
// 是 MM/TrackID territory（W3.4+），且安全默认=保 Real 不擦写成抑制。
type TrackArchive struct {
	p          RealnessParams
	born       bool
	birthX     int
	birthY     int
	birthMs    int64
	stillStart int64 // 当前静止 run 起点 ms（0=未在静止 run）
	crossed    bool  // 曾完成一个 ≥StillPeriodMs 的静止 run（survival 慢证据，单调置位）
	hist       []histPt
}

// NewTrackArchive 建空档案（出生待首帧 Observe 填）。
func NewTrackArchive(p RealnessParams) *TrackArchive { return &TrackArchive{p: p} }

// Observe 一帧 → belief.RealnessObs。仅 Online 帧调用（track absent = 不分类，realness 后验由 caller
// 持住）。返回 birth=true 仅首 Online 帧 → caller 据此起 RealnessTrack 出生先验。
func (a *TrackArchive) Observe(t RadarTrack, nowMs int64) (belief.RealnessObs, bool) {
	birth := false
	if !a.born {
		a.born = true
		a.birthX, a.birthY, a.birthMs = t.X, t.Y, nowMs
		birth = true
	}

	dispFromBirth := math.Hypot(float64(t.X-a.birthX), float64(t.Y-a.birthY))
	displaced := dispFromBirth > float64(a.p.MoveCm)
	confinedToBirth := dispFromBirth <= float64(a.p.MoveCm)
	ageLongStatic := nowMs-a.birthMs >= a.p.LongStaticMs && confinedToBirth

	a.updateStillBox(t, nowMs)

	return belief.RealnessObs{
		Displaced:          displaced,
		AgeLongStatic:      ageLongStatic,
		CrossedStillPeriod: a.crossed,
	}, birth
}

// updateStillBox 滚动窗 still-box：窗内 per-axis 包络 ≤ StillBoxCm = 静止 run；run 满 StillPeriodMs
// 即单调置 crossed（survival 慢证据：ghost/金属活不过静止降功率期，真人活得过）。
// per-axis 包络（非对角线）：倒地质心抖 50×40 框对角线 64 会误判"动"，per-axis 算 50 才正确判 still。
func (a *TrackArchive) updateStillBox(t RadarTrack, nowMs int64) {
	a.hist = append(a.hist, histPt{nowMs, t.X, t.Y})
	cut := nowMs - a.p.StillWindowMs
	drop := 0
	for drop < len(a.hist) && a.hist[drop].tMs < cut {
		drop++
	}
	a.hist = a.hist[drop:]

	minX, minY, maxX, maxY := t.X, t.Y, t.X, t.Y
	for _, h := range a.hist {
		minX, maxX = mini(minX, h.x), maxi(maxX, h.x)
		minY, maxY = mini(minY, h.y), maxi(maxY, h.y)
	}
	if maxi(maxX-minX, maxY-minY) <= a.p.StillBoxCm && len(a.hist) >= 2 {
		if a.stillStart == 0 {
			a.stillStart = a.hist[0].tMs
		} else if nowMs-a.stillStart >= a.p.StillPeriodMs {
			a.crossed = true
		}
	} else {
		a.stillStart = 0
	}
}

func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}
