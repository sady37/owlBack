package adapter

import (
	"math"
	"sort"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// census.go — 房内多 track 人数 N_r（§G 主职：排 ghost，防 1 真人+1 影子被当 2 人 → C_FN 误折扣漏报）。
//   身份 = 最小作功距离 logicID（§G六）：新 track 发新号 / 平时跟随 / 相遇按最近距离保号。
//   两类 ghost：mirror（co-existence ρ + IsReflection）**仅 track 数==2**（§G六，1=永发/3+=不处理）；
//     运动伪迹（超速异常量，§G七桶一）**单 track 每帧**——数量×时间累积、非硬阈，无真人源（§54 独立分量）。
//   co-existence ρ = 两 track 速度同步度（本层算）；IsReflection（哪条是反射）= MM 拓扑 W3.4b 填。

// TrackCensusParams 形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type TrackCensusParams struct {
	MoveCm       int     // 偏离出生位 > 此 = Displaced（走动/位移）
	AssocCm      int     // 帧间关联最大合理位移（新坐标离最近现有 track > 此 = 新 track 发号）
	SpeedCeilCm  int     // 室内合理速度上限（cm/帧≈cm/s）；超出 = 运动伪迹异常量（§G七桶一）
	ArtifactGain float64 // 超速幅度 → 伪迹 aScore 累积增益（数量×时间，非硬阈）
}

// DefaultTrackCensusParams 形态默认（非权威值，留 oracle）。
func DefaultTrackCensusParams() TrackCensusParams {
	return TrackCensusParams{MoveCm: 50, AssocCm: 250, SpeedCeilCm: 100, ArtifactGain: 0.002}
}

// TrackObs 一帧一条 raw track（§57 步2 全量：雷达量 + MM 反射标注）——既数 N_r 又驱动 per-track 滤波。
type TrackObs struct {
	RadarTrack        // Online/Pose/X,Y,Z/HR,RR/StillSec（X,Y 经嵌入提升供 census 关联/速度）
	IsReflection bool // MM 拓扑：本 track 是另一条的反射（W3.4b 填；本层不算几何）
}

type censusTrack struct {
	rt             *belief.RealnessTrack
	birthX, birthY int
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

// Update 一帧推进：① 最小作功距离关联 raw→logicID；② track 数==2 算 ρ；③ 各 RealnessTrack.Update。
// nowMs = 本帧时戳：速度按 cm/s（位移/dt）算——帧间隔非恒 1s（replay 真实时戳有多秒间隔），按帧位移
//
//	会把跨时隙的正常走动误判成超速伪迹（cd2b 回归根因）；归一到 cm/s 才是「室内合理速度上限」语义。
func (c *TrackCensus) Update(nowMs int64, obs []TrackObs) {
	c.tick++
	ids := c.associate(obs, nowMs)

	// 帧间速度（cm/s）—— 伪迹判别 + co-existence ρ 用。出生/同帧（dt≤0）= 无速度基准 → 0。
	speeds := make([]float64, len(obs))
	for i, o := range obs {
		t := c.tracks[ids[i]]
		if dt := float64(nowMs-t.lastMs) / 1000.0; dt > 0 {
			speeds[i] = math.Hypot(float64(o.X-t.x), float64(o.Y-t.y)) / dt
		}
	}
	rho := 0.0
	if len(obs) == 2 { // §G六：ghost(mirror) 仅 track 数==2 处理
		rho = speedSync(speeds[0], speeds[1])
	}

	for i, o := range obs {
		t := c.tracks[ids[i]]
		ro := belief.RealnessObs{
			Displaced:       math.Hypot(float64(o.X-t.birthX), float64(o.Y-t.birthY)) > float64(c.p.MoveCm),
			ArtifactQuantum: c.artifactQuantum(speeds[i]), // 单 track 即算（§G七桶一）→ 喂 N_r 排假人头。
			// 注意：artifact **不在 census 层压孤轨的摔**——「无共存源→不压 pFallReal」由 engine 消费门控管
			//   （§61：计算门控加 track==2=错修砍 N_r 功能；消费门控=对）。census 只产 PReal/N_r，不决定压不压摔。
		}
		if len(obs) == 2 { // mirror 仅 track==2（成对 co-existence ρ + 反射几何）；伪迹 artifact 不受此门控
			ro.CoexistRho = rho
			ro.IsReflection = o.IsReflection
		}
		t.rt.Update(ro)
		t.x, t.y, t.lastTick, t.lastMs, t.lastObs = o.X, o.Y, c.tick, nowMs, o
	}
}

// associate 最小作功距离关联（贪心最近邻；≤2 track 足够，多 track 交叉的最优指派后续再说）。
func (c *TrackCensus) associate(obs []TrackObs, nowMs int64) []int {
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
		c.tracks[c.nextID] = &censusTrack{rt: belief.NewRealnessTrack(), birthX: o.X, birthY: o.Y, x: o.X, y: o.Y, lastMs: nowMs}
		ids[i], used[c.nextID] = c.nextID, true
	}
	return ids
}

// Nr 房内真人数 = 本 tick 在场且非 ghost（mirror+伪迹）的 track 数（PReal ≥ 0.5）。
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
	LogicID int
	Obs     TrackObs // 在场=本帧 / 消失=末帧（喂 per-track 滤波发射；消失态走 blind 续存仅 Predict）
	PReal   float64  // 真人后验（ghost→低）；engine OR 聚合只算 PReal≥0.5 的真人 track
	Present bool     // 本 tick 匹配到 raw（false=消失，engine 据此走 blind 续存 Predict-only）
}

// Tracks engine 读 census 已知的全部 track（在场 + 消失未 drop），按 LogicID 升序（确定性）。
// census 是身份/realness 单源（§59③）；engine 据此每 LogicID 跑一份滤波，不重做关联。
func (c *TrackCensus) Tracks() []TrackState {
	out := make([]TrackState, 0, len(c.tracks))
	for id, t := range c.tracks {
		out = append(out, TrackState{LogicID: id, Obs: t.lastObs, PReal: t.rt.PReal(), Present: t.lastTick == c.tick})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LogicID < out[j].LogicID })
	return out
}

// Drop 移除一条 track（engine 在其 S^(i) 被吸收到 Left/Empty 后调用；非 TTL，状态驱动）。
func (c *TrackCensus) Drop(logicID int) { delete(c.tracks, logicID) }

// artifactQuantum 本帧运动伪迹异常量（§G七桶一）：速度超室内合理上限的幅度 × 增益（数量×时间累积、
//
//	非硬阈——单帧不判，持续超速才积成 ghost）。新生帧 speed=0 → 0；teleport >AssocCm 已成新 track 不在此。
func (c *TrackCensus) artifactQuantum(speed float64) float64 {
	if speed <= float64(c.p.SpeedCeilCm) {
		return 0
	}
	return (speed - float64(c.p.SpeedCeilCm)) * c.p.ArtifactGain
}

// speedSync 两 track 速度同步度 ∈[0,1]：同时同速运动 → 1（镜像与源同步）；速度差大或都不动 → 0。
func speedSync(a, b float64) float64 {
	m := math.Max(a, b)
	if m == 0 {
		return 0 // 都不动：无共动证据（静止绝不判 mirror）
	}
	return 1 - math.Abs(a-b)/m
}
