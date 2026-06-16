package adapter

import (
	"math"
	"sort"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// census.go — 房内多 track 人数 N_r（§G 主职：排 ghost，防 1 真人+1 影子被当 2 人 → C_FN 误折扣漏报）。
//   身份 = 最小作功距离 logicID（§G六）：新 track 发新号 / 平时跟随 / 相遇按最近距离保号。
//   realness = belief.RealnessTrack 2 态前向滤波(Real/Mirror，转移矩阵+软发射，device-room-zone.md §9)；
//     本层只译软证据喂之：出生地距门 D(§9.3①)、自主位移、co-existence ρ(速度同步度)、墙外反射裕度(桶二几何)、配对存在。
//   ghost 仅 track==2 由 Coexist 耦合涌现(孤轨 Coexist=0 → mirror 发射中性 → 永 Real)，非本层硬判。
//   reflectSep(桶二 §69)：radar→ghost 连线穿 wall 取距 ghost 最近交点；≥ReflSepCm 才算反射(确定性几何阈)。

// TrackCensusParams 形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type TrackCensusParams struct {
	MoveCm      int // 偏离出生位 > 此 = Displaced（走动/位移）
	AssocCm     int // 帧间关联最大合理位移（新坐标离最近现有 track > 此 = 新 logicID 发号）
	ReflSepCm   int // 桶二（§69）：交点(最近 ghost)→ghost ≥此才算反射（确定性几何阈：雷达精度/墙太靠内不算）
	WallScaleCm int // §9.2 墙外反射裕度归一尺度（sep/此 → WallMargin[0,1] 软发射）
}

// DefaultTrackCensusParams 形态默认（非权威值，留 oracle；ReflSepCm 是确定性几何阈非 oracle 软参）。
func DefaultTrackCensusParams() TrackCensusParams {
	return TrackCensusParams{MoveCm: 50, AssocCm: 250, ReflSepCm: 30, WallScaleCm: 100}
}

// TrackObs 一帧一条 raw track（§57 步2 全量：纯雷达量）——既数 N_r 又驱动 per-track 滤波。
// 桶二（§69）：IsReflection 由 census 本层几何算（出生位 + 墙 + 雷达坐标），不再外部标注透传。
type TrackObs struct {
	RadarTrack // Online/Pose/X,Y,Z/HR,RR/StillSec（X,Y 经嵌入提升供 census 关联/速度）
}

type censusTrack struct {
	rt             *belief.RealnessTrack
	birthX, birthY int
	birthMs        int64    // 出生时戳（§9.1 LaterBorn：成对取后到者破同步对称）
	doorD          float64  // §9.3① 出生地→最近门距离 cm（出生时算一次；<0=无 enter 区）
	isRefl         bool     // 桶二：本 track 末帧反射几何判定（forensic 观测）
	x, y           int      // 上一 tick 坐标
	lastTick       int64    // 最近匹配 tick
	lastMs         int64    // 最近匹配时戳（算 cm/s 速度，帧间隔非恒 1s）
	lastObs        TrackObs // 最近一帧全量 obs（喂 engine per-track 滤波发射；消失态持住末帧）
}

// TrackCensus 房内多 track 集：最小作功 logicID + 每 track realness + N_r。
type TrackCensus struct {
	p      TrackCensusParams
	nextID int
	tracks map[int]*censusTrack
	tick   int64
}

// NewTrackCensus 建空 census。
func NewTrackCensus(p TrackCensusParams) *TrackCensus {
	return &TrackCensus{p: p, tracks: map[int]*censusTrack{}}
}

// Update 一帧推进：① 最小作功距离关联 raw→logicID；② track 数==2 算 co-existence ρ；
// ③ 各 track 译软证据喂 RealnessTrack 前向滤波（§9）。速度按 cm/s（位移/dt，帧间隔非恒 1s）。
func (c *TrackCensus) Update(nowMs int64, obs []TrackObs, radar Point, walls, entrances []Rect) {
	c.tick++
	ids := c.associate(obs, nowMs, entrances)

	// 帧间速度（cm/s）→ co-existence ρ。出生/同帧（dt≤0）= 无速度基准 → 0。
	speeds := make([]float64, len(obs))
	for i, o := range obs {
		t := c.tracks[ids[i]]
		if dt := float64(nowMs-t.lastMs) / 1000.0; dt > 0 {
			speeds[i] = math.Hypot(float64(o.X-t.x), float64(o.Y-t.y)) / dt
		}
	}
	rho, coexist := 0.0, 0.0
	laterID := -1
	if len(obs) == 2 { // §G六：mirror co-existence 仅 track 数==2（1=永发 / 3+=不处理 → coexist 0）
		rho = speedSync(speeds[0], speeds[1])
		coexist = 1
		a, b := c.tracks[ids[0]], c.tracks[ids[1]] // §9.1 后到者破同步对称（同时生→无后到→都不归同步）
		if a.birthMs > b.birthMs {
			laterID = ids[0]
		} else if b.birthMs > a.birthMs {
			laterID = ids[1]
		}
	}

	for i, o := range obs {
		t := c.tracks[ids[i]]
		sep := reflectSep(o.X, o.Y, radar, walls, c.p.ReflSepCm) // 桶二：墙外反射裕度 cm（0=墙内/非反射）
		t.isRefl = sep > 0
		wallMargin := 0.0
		if sep > 0 {
			if wallMargin = sep / float64(c.p.WallScaleCm); wallMargin > 1 {
				wallMargin = 1
			}
		}
		t.rt.Update(belief.RealnessObs{
			BirthDoorD: t.doorD,
			Displaced:  math.Hypot(float64(o.X-t.birthX), float64(o.Y-t.birthY)) > float64(c.p.MoveCm),
			CoexistRho: rho,
			LaterBorn:  ids[i] == laterID,
			WallMargin: wallMargin,
			Coexist:    coexist,
			DtMs:       nowMs - t.lastMs,
		})
		t.x, t.y, t.lastTick, t.lastMs, t.lastObs = o.X, o.Y, c.tick, nowMs, o
	}
}

// associate 最小作功距离关联（贪心最近邻；≤2 track 足够，多 track 交叉的最优指派后续再说）。
func (c *TrackCensus) associate(obs []TrackObs, nowMs int64, entrances []Rect) []int {
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
		c.tracks[c.nextID] = &censusTrack{rt: belief.NewRealnessTrack(), birthX: o.X, birthY: o.Y, birthMs: nowMs, doorD: birthDoorDist(o.X, o.Y, entrances), x: o.X, y: o.Y, lastMs: nowMs}
		ids[i], used[c.nextID] = c.nextID, true
	}
	return ids
}

// Nr 房内真人数 = 本 tick 在场且 PReal≥0.5 的 track 数（realness 后验软阈；排 Mirror ghost）。
//
//	Blind 续存（摔倒静止降功率被滤的 lost-Fallen track 仍算 1 人，§G六②）**不靠 census TTL**——
//	架构师 2026-06-15 裁定「随 S^(i) 涌现，本刀不加 TTL」：N_r=Σ1[S^(i)∉{E,L}]（DBN-Zone-Room line57）
//	本就定义在 S^(i) 上，blind 续存计数由 step2 多 track 进 belief 主管线后经 S^(i) 转移自持涌现
//	（belief 自持非 staleness 补丁，呼应 §32 删 TTL）。单 faller 续存已由 belief 自持、且其计不计入
//	N_r 不改自身上报（≤1 人不折扣 / ≥55% 必报）。故本片只数在场，不在 census 层补续存窗。
func (c *TrackCensus) Nr() int {
	n := 0
	for _, t := range c.tracks {
		if t.lastTick == c.tick && t.rt.PReal() >= 0.5 {
			n++
		}
	}
	return n
}

// TrackState 一条 track 对 engine 的读出（§57 步2：身份 + 末帧 obs + realness + 是否在场）。
type TrackState struct {
	LogicID      int
	Obs          TrackObs // 在场=本帧 / 消失=末帧（喂 per-track 滤波发射；消失态走 blind 续存仅 Predict）
	PReal        float64  // 真人后验（ghost→低）；engine OR 聚合只算 PReal≥0.5 的真人 track
	Present      bool     // 本 tick 匹配到 raw（false=消失，engine 据此走 blind 续存 Predict-only）
	PMirror      float64  // 镜像后验（有真人源的 ghost）；forensic 观测
	IsReflection bool     // 桶二镜面几何判定（radar→ghost 连线穿墙取最近交点≥阈）；forensic 观测
}

// Tracks engine 读 census 已知的全部 track（在场 + 消失未 drop），按 LogicID 升序（确定性）。
// census 是身份/realness 单源（§59③）；engine 据此每 LogicID 跑一份滤波，不重做关联。
func (c *TrackCensus) Tracks() []TrackState {
	out := make([]TrackState, 0, len(c.tracks))
	for id, t := range c.tracks {
		out = append(out, TrackState{LogicID: id, Obs: t.lastObs, PReal: t.rt.PReal(), Present: t.lastTick == c.tick, PMirror: t.rt.PMirror(), IsReflection: t.isRefl})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LogicID < out[j].LogicID })
	return out
}

// Drop 移除一条 track（engine 在其 S^(i) 被吸收到 Left/Empty 后调用；非 TTL，状态驱动）。
func (c *TrackCensus) Drop(logicID int) { delete(c.tracks, logicID) }

// birthDoorDist 出生地→最近 enter 区(门)距离 cm；无 enter 区 → -1（§9.3① 跳过近门似然，
//
//	enter 区常画不准 → 用距离软评分不用硬门，截断重放/无门房自然落 -1）。
func birthDoorDist(x, y int, entrances []Rect) float64 {
	if len(entrances) == 0 {
		return -1
	}
	best := math.MaxFloat64
	for _, e := range entrances {
		if d := distCm(x, y, e); d < best {
			best = d
		}
	}
	return best
}

// speedSync 两 track 速度同步度 ∈[0,1]：同时同速运动 → 1（镜像与源同步）；速度差大或都不动 → 0。
func speedSync(a, b float64) float64 {
	m := math.Max(a, b)
	if m == 0 {
		return 0 // 都不动：无共动证据（静止绝不判 mirror）
	}
	return 1 - math.Abs(a-b)/m
}

// reflectSep 桶二镜像几何（§69 五步，架构师定法）→ 反射裕度 cm（0=非反射，喂 WallMargin 软发射）：
//
//	① ghost 在任一 wall 矩形内（墙内 = 非反射）→ 0；
//	② radar→ghost 连线与各 wall 边求交，取**距 ghost 最近的交点 = mirror 点**（多墙取全局最近，§69 步4）；
//	③ 连线不穿任何墙 → 0；④ mirror 点→ghost <sepCm（雷达精度/墙太靠内）→ 0，FN-safe 偏 0。
//
// FN-safe（§69 柱B）：宁可漏判镜像不可误判真人为镜像。真人在房内（墙内/不穿墙）几何天然 0；
// 只有 ghost 真在墙外且连线穿墙 ≥阈才返裕度。无墙配置 → 恒 0（零回归，§69 柱E）。
func reflectSep(gx, gy int, radar Point, walls []Rect, sepCm int) float64 {
	for _, w := range walls {
		if insideRect(gx, gy, w) {
			return 0 // ① ghost 在墙内 → 非反射
		}
	}
	gxf, gyf := float64(gx), float64(gy)
	best, found := math.MaxFloat64, false
	for _, w := range walls {
		for _, p := range segRectIntersections(float64(radar.X), float64(radar.Y), gxf, gyf, w) {
			if d := math.Hypot(p.x-gxf, p.y-gyf); d < best { // ② 距 ghost 最近交点（含多墙全局最近）
				best, found = d, true
			}
		}
	}
	if !found || best < float64(sepCm) {
		return 0 // ③ 不穿墙 / ④ 交点→ghost <阈 → 非反射
	}
	return best // 反射裕度（mirror 点→ghost 距，cm）
}

// insideRect 点在矩形内（含边界）。
func insideRect(x, y int, r Rect) bool {
	return x >= r.X1 && x <= r.X2 && y >= r.Y1 && y <= r.Y2
}

type ptf struct{ x, y float64 }

// segRectIntersections 线段 AB 与矩形 4 条边的全部交点（穿墙可得 2 点：进/出边）。
func segRectIntersections(ax, ay, bx, by float64, r Rect) []ptf {
	x1, y1, x2, y2 := float64(r.X1), float64(r.Y1), float64(r.X2), float64(r.Y2)
	edges := [4][4]float64{
		{x1, y1, x2, y1}, // top
		{x1, y2, x2, y2}, // bottom
		{x1, y1, x1, y2}, // left
		{x2, y1, x2, y2}, // right
	}
	var out []ptf
	for _, e := range edges {
		if p, ok := segSeg(ax, ay, bx, by, e[0], e[1], e[2], e[3]); ok {
			out = append(out, p)
		}
	}
	return out
}

// segSeg 线段 AB 与线段 CD 的交点（平行/不相交 → ok=false）。
func segSeg(ax, ay, bx, by, cx, cy, dx, dy float64) (ptf, bool) {
	r1x, r1y := bx-ax, by-ay
	r2x, r2y := dx-cx, dy-cy
	denom := r1x*r2y - r1y*r2x
	if denom == 0 {
		return ptf{}, false // 平行/共线
	}
	t := ((cx-ax)*r2y - (cy-ay)*r2x) / denom
	u := ((cx-ax)*r1y - (cy-ay)*r1x) / denom
	if t < 0 || t > 1 || u < 0 || u > 1 {
		return ptf{}, false
	}
	return ptf{ax + t*r1x, ay + t*r1y}, true
}
