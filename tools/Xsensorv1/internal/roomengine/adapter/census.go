package adapter

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// census.go — 房内多 track 人数 N_r（§G 主职：排 ghost，防 1 真人+1 影子被当 2 人 → C_FN 误折扣漏报）。
//   身份 = 最小作功距离 logicID（§G六）：新 track 发新号 / 平时跟随 / 相遇按最近距离保号。
//   ghost(mirror) 仅 track 数 == 2 处理：1=永发（不进 ghost）；2=跑 mirror；3+=不处理（人多风险低）。
//   co-existence ρ = 两 track 速度同步度（本层算，镜像与源同时同速运动）；
//   IsReflection（哪条是反射、破对称）= 反射几何 = MM 拓扑，W3.4b 填（本层不算几何）。

// TrackCensusParams 形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type TrackCensusParams struct {
	MoveCm  int // 偏离出生位 > 此 = Displaced（走动/位移）
	AssocCm int // 帧间关联最大合理位移（新坐标离最近现有 track > 此 = 新 track 发号）
}

// DefaultTrackCensusParams 形态默认（非权威值，留 oracle）。
func DefaultTrackCensusParams() TrackCensusParams { return TrackCensusParams{MoveCm: 50, AssocCm: 100} }

// TrackObs 一帧一条 raw track（坐标 + MM 的反射标注）。
type TrackObs struct {
	X, Y         int
	IsReflection bool // MM 拓扑：本 track 是另一条的反射（W3.4b 填；本层不算几何）
}

type censusTrack struct {
	rt             *belief.RealnessTrack
	birthX, birthY int
	x, y           int   // 上一 tick 坐标
	lastTick       int64 // 最近匹配 tick
}

// TrackCensus 房内多 track 集：最小作功 logicID + 每 track realness + N_r。
type TrackCensus struct {
	p      TrackCensusParams
	nextID int
	tracks map[int]*censusTrack
	tick   int64
}

// NewTrackCensus 建空 census。
func NewTrackCensus(p TrackCensusParams) *TrackCensus { return &TrackCensus{p: p, tracks: map[int]*censusTrack{}} }

// Update 一帧推进：① 最小作功距离关联 raw→logicID；② track 数==2 算 ρ；③ 各 RealnessTrack.Update。
func (c *TrackCensus) Update(obs []TrackObs) {
	c.tick++
	ids := c.associate(obs)

	// 帧间位移大小（速度）—— co-existence ρ 用。
	speeds := make([]float64, len(obs))
	for i, o := range obs {
		t := c.tracks[ids[i]]
		speeds[i] = math.Hypot(float64(o.X-t.x), float64(o.Y-t.y))
	}
	rho := 0.0
	if len(obs) == 2 { // §G六：ghost(mirror) 仅 track 数==2 处理
		rho = speedSync(speeds[0], speeds[1])
	}

	for i, o := range obs {
		t := c.tracks[ids[i]]
		ro := belief.RealnessObs{
			Displaced: math.Hypot(float64(o.X-t.birthX), float64(o.Y-t.birthY)) > float64(c.p.MoveCm),
		}
		if len(obs) == 2 { // 1=永发不喂 / 3+=不处理不喂
			ro.CoexistRho = rho
			ro.IsReflection = o.IsReflection
		}
		t.rt.Update(ro)
		t.x, t.y, t.lastTick = o.X, o.Y, c.tick
	}
}

// associate 最小作功距离关联（贪心最近邻；≤2 track 足够，多 track 交叉的最优指派后续再说）。
func (c *TrackCensus) associate(obs []TrackObs) []int {
	ids := make([]int, len(obs))
	used := map[int]bool{}
	for i, o := range obs {
		best, bestD := -1, math.MaxFloat64
		for id, t := range c.tracks {
			if used[id] {
				continue
			}
			if d := math.Hypot(float64(o.X-t.x), float64(o.Y-t.y)); d < bestD {
				best, bestD = id, d
			}
		}
		if best >= 0 && bestD <= float64(c.p.AssocCm) {
			ids[i], used[best] = best, true
			continue
		}
		c.nextID++
		c.tracks[c.nextID] = &censusTrack{rt: belief.NewRealnessTrack(), birthX: o.X, birthY: o.Y, x: o.X, y: o.Y}
		ids[i], used[c.nextID] = c.nextID, true
	}
	return ids
}

// Nr 房内真人数 = 本 tick 在场且非 mirror 的 track 数（PReal ≥ 0.5）。
//
//	（Blind 续存的 lost-Fallen track 计入留 S^(i) 整合后；本片只数在场，不数已消失。）
func (c *TrackCensus) Nr() int {
	n := 0
	for _, t := range c.tracks {
		if t.lastTick == c.tick && t.rt.PReal() >= 0.5 {
			n++
		}
	}
	return n
}

// speedSync 两 track 速度同步度 ∈[0,1]：同时同速运动 → 1（镜像与源同步）；速度差大或都不动 → 0。
func speedSync(a, b float64) float64 {
	m := math.Max(a, b)
	if m == 0 {
		return 0 // 都不动：无共动证据（静止绝不判 mirror）
	}
	return 1 - math.Abs(a-b)/m
}
