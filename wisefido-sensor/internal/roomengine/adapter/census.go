package adapter

import (
	"math"
	"sort"

	"go.uber.org/zap"
	"wisefido-sensor/internal/roomengine/belief"
)

// census.go — 房内多 track 人数 N_r（§G 主职：排 ghost，防 1 真人+1 影子被当 2 人 → C_FN 误折扣漏报）。
//   身份 = 最小作功距离 logicID（§G六）：新 track 发新号 / 平时跟随 / 相遇按最近距离保号。
//   realness = belief.RealnessTrack 2 态前向滤波(Real/Mirror，转移矩阵+软发射，device-room-zone.md §9)；
//     本层只译软证据喂之：出生地距门 D(§9.3①)、自主位移、co-existence ρ(速度同步度)、墙外反射裕度(桶二几何)、配对存在。
//   ghost 仅 track==2 由 Coexist 耦合涌现(孤轨 Coexist=0 → mirror 发射中性 → 永 Real)，非本层硬判。
//   reflectSep(桶二 §69)：radar→ghost 连线穿 wall 取距 ghost 最近交点；≥ReflSepCm 才算反射(确定性几何阈)。

// TrackCensusParams 形态参数（form-anchor，标定留 oracle [[fall_data_is_artificial_test]]）。
type TrackCensusParams struct {
	MoveCm         int   // 偏离出生位 > 此 = Displaced（走动/位移）
	AssocCm        int   // 帧间关联最大合理位移（新坐标离最近现有 track > 此 = 新 logicID 发号）
	ReflSepCm      int   // 桶二（§69）：交点(最近 ghost)→ghost ≥此才算反射（确定性几何阈：雷达精度/墙太靠内不算）
	WallScaleCm    int   // §9.2 墙外反射裕度归一尺度（sep/此 → WallMargin[0,1] 软发射）
	MirrorWindowMs int64 // 镜像/sync 判定窗：出生后 ≤此算几何+sync 喂 realness，过则冻结判定（省 reflectSep 算力）
}

// DefaultTrackCensusParams 形态默认（非权威值，留 oracle；ReflSepCm 是确定性几何阈非 oracle 软参）。
func DefaultTrackCensusParams() TrackCensusParams {
	return TrackCensusParams{MoveCm: 50, AssocCm: 250, ReflSepCm: 30, WallScaleCm: 100, MirrorWindowMs: 5000}
}

// TrackObs 一帧一条 raw track（§57 步2 全量：纯雷达量）——既数 N_r 又驱动 per-track 滤波。
// 桶二（§69）：IsReflection 由 census 本层几何算（出生位 + 墙 + 雷达坐标），不再外部标注透传。
type TrackObs struct {
	RadarTrack             // Online/Pose/X,Y,Z/HR,RR/StillSec（X,Y 经嵌入提升供 census 速度/realness 几何）
	LogicID        string  // tm 出生锚定的唯一逻辑身份（nearestAliveTrack 最小作功距离继承+EnterRoom 门）；census 不再按位置重造 int logicID，直接引用此
	SplitConvicted bool    // roomengine split 运动学坐实(SplitGhostSinceMs>0)→ 喂 realness mEv 降 PReal（孤迹 latch 也开）
	ForceReal      bool    // Phase 1.5：roomengine real-by-provenance latch 生效（provenance 证真 ∧ 仍在 rest 区）→ census 强制 real 保 nr≥1，跳几何/sync 判定
}

type censusTrack struct {
	rt             *belief.RealnessTrack
	birthX, birthY int
	birthMs        int64    // 出生时戳（§9.1 LaterBorn：成对取后到者破同步对称）
	doorD          float64  // §9.3① 出生地→最近门距离 cm（出生时算一次；<0=无 enter 区）
	isRefl         bool     // 桶二：本 track 末帧反射几何判定（forensic 观测）
	fSep           float64  // forensic：末帧 reflectSep（墙外反射裕度 cm，0=墙内/非反射）
	fWallMargin    float64  // forensic：末帧 WallMargin（sep/WallScale，喂 mEv 墙外项）
	fRho           float64  // forensic：末帧 CoexistRho（同步移动强度）
	fLater         bool     // forensic：末帧 LaterBorn（成对后到=ghost 候选）
	birthLogged    bool     // 出生窗冻结那帧已写过 birth_verdict log（一次性 forensic，避免逐帧刷）
	x, y           int      // 上一 tick 坐标
	lastTick       int64    // 最近匹配 tick
	lastMs         int64    // 最近匹配时戳（算 cm/s 速度，帧间隔非恒 1s）
	lastObs        TrackObs // 最近一帧全量 obs（喂 engine per-track 滤波发射；消失态持住末帧）
	pathAccum      float64  // 跨设备融合：窗口内累计移动路程(指数衰减)；多雷达帧交错→比窗口路程不比单帧位移
	pathMs         int64    // pathAccum 末次更新时戳
}

// recentPath 把累计路程衰减到 nowMs（读时统一时基，因各雷达帧交错更新时刻不同）。
func (t *censusTrack) recentPath(nowMs int64) float64 {
	if t.pathMs == 0 {
		return 0
	}
	return t.pathAccum * math.Exp(-float64(nowMs-t.pathMs)/fusePathTauMs)
}

// 跨设备同人融合（同 Room 不同 radar 看同一人 → N_r 只数 1；契约 doc/fusion-absorption-contract.md）。
// **唯一影响 N_r**：不动 track 集/belief/realness/fire。同房=同一 census(MM room⊃device 前缀派生)；
// 同人=5s 窗内每 tick 移动幅值同步(frame-invariant，免坐标对齐)。同设备(DevKey 同)是镜像由 realness 管，不在此。
const (
	fusePathTauMs    = 2500 // 累计路程指数衰减时间常数 ms（≈5s 窗的"近期移动"度量）
	fuseMoveThreshCm = 40   // 窗口路程 ≥ 此才算"有意义移动"（都静止=不喂证据，保持已锁定判定）
	fuseEMA          = 0.3  // 窗口路程同步 EMA 权
	fuseMinMoves     = 5    // 累计有效移动帧 ≥ 此才允许判同人（多帧抗单帧巧合）
	fuseAgreeThresh  = 0.6  // 同步 EMA ≥ 此 → 锁定同人
	fuseStaleMs      = 3000 // 链双方久未同帧 → GC（人离场=释放；一台退化不解锁，仍同一人）
	fuseExitHystTick = 3    // 多人证据持续 ≥此 tick 才退出折叠（B §7.2 解合并防抖；滤 realness 单帧 ghost 闪）
)

// fuseLink 跨设备一对 track 的同人证据（运动幅值同步 EMA + 锁定态）。
type fuseLink struct {
	a, b   string  // 配对 logicID（a<b 固定序）
	agree  float64 // 运动幅值同步 EMA [0,1]
	moves  int     // 累计有效移动帧（证据量）
	same   bool    // 锁定：判同人 → Nr 折叠
	lastMs int64   // 末次更新（GC 陈旧）
}

// TrackCensus 房内多 track 集：最小作功 logicID + 每 track realness + N_r。
type TrackCensus struct {
	p               TrackCensusParams
	tracks          map[string]*censusTrack // 键=tm 出生 LogicID（唯一身份；census 不再自发 int 号）
	fuseLinks       map[string]*fuseLink    // 跨设备同人链（键=ordered logicID 对）；仅折叠 N_r
	multiPersonStrk int                     // 连续多人(任一雷达≥2真人)tick 数；≥fuseExitHystTick 才退折叠(解合并防抖)
	multiPersonLogd bool                    // no-silent-caps：多人退出已 LOG(转换沿一次)
	tick            int64
}

// NewTrackCensus 建空 census。
func NewTrackCensus(p TrackCensusParams) *TrackCensus {
	return &TrackCensus{p: p, tracks: map[string]*censusTrack{}, fuseLinks: map[string]*fuseLink{}}
}

// fusePairKey 两 logicID 的固定序键。
func fusePairKey(a, b string) string {
	if b < a {
		a, b = b, a
	}
	return a + "\x00" + b
}

// DevKey logicID → uid_last4（makeLogicID 前缀，per-device 键：床耦合 + mirror 同设备配对单源）。
func DevKey(logicID string) string {
	if len(logicID) >= 4 {
		return logicID[:4]
	}
	return logicID
}

// Update 一帧推进：① 最小作功距离关联 raw→logicID；② 同设备 2 track 算 co-existence ρ；
// ③ 各 track 译软证据喂 RealnessTrack 前向滤波（§9）。速度按 cm/s（位移/dt，帧间隔非恒 1s）。
func (c *TrackCensus) Update(nowMs int64, obs []TrackObs, radar Point, walls, entrances []Rect) {
	c.tick++
	// 身份直接引用 tm 出生 LogicID（§G六 关联在 tm nearestAliveTrack 已做,census 不重造）；
	//   首次见的 LogicID 初始化 censusTrack（出生窗 realness 锚:doorD/birthXY/birthMs）。
	ids := make([]string, len(obs))
	for i, o := range obs {
		ids[i] = o.LogicID
		if _, ok := c.tracks[ids[i]]; !ok {
			dd := birthDoorDist(o.X, o.Y, entrances)
			c.tracks[ids[i]] = &censusTrack{rt: belief.NewRealnessTrack(dd), birthX: o.X, birthY: o.Y, birthMs: nowMs, doorD: dd, x: o.X, y: o.Y, lastMs: nowMs}
		}
	}

	// 帧间速度（cm/s）→ co-existence ρ。出生/同帧（dt≤0）= 无速度基准 → 0。
	// disp = 单帧位移幅值（cm），喂跨设备同人融合（frame-invariant）。
	speeds := make([]float64, len(obs))
	disp := make([]float64, len(obs))
	for i, o := range obs {
		t := c.tracks[ids[i]]
		disp[i] = math.Hypot(float64(o.X-t.x), float64(o.Y-t.y))
		if dt := float64(nowMs-t.lastMs) / 1000.0; dt > 0 {
			speeds[i] = disp[i] / dt
		}
	}
	// mirror co-existence 仅在【同一设备】恰好 2 track 之间（§G六 + 多雷达订正）：反射 ghost 是单台雷达
	// FOV 内多径，**跨雷达两条 track 是对同一人的冗余探测，绝不互为镜像**（09e7 case：09E7 摔轨被 D523
	// 当镜像源压成 ghost）。按 devKey(uid_last4) 分组，仅 size==2 的设备组算 ρ/coexist/laterBorn；
	// 1=永发 / 3+=不处理 → coexist 0。
	coexistOf := make([]float64, len(obs))
	rhoOf := make([]float64, len(obs))
	laterOf := make([]bool, len(obs))
	byDev := map[string][]int{}
	for i := range ids {
		byDev[DevKey(ids[i])] = append(byDev[DevKey(ids[i])], i)
	}
	for _, grp := range byDev {
		if len(grp) != 2 {
			continue
		}
		p, q := grp[0], grp[1]
		rho := speedSync(speeds[p], speeds[q])
		coexistOf[p], coexistOf[q] = 1, 1
		rhoOf[p], rhoOf[q] = rho, rho
		a, b := c.tracks[ids[p]], c.tracks[ids[q]] // §9.1 后到者破同步对称（同时生→无后到→都不归同步）
		if a.birthMs > b.birthMs {
			laterOf[p] = true
		} else if b.birthMs > a.birthMs {
			laterOf[q] = true
		}
	}

	for i, o := range obs {
		t := c.tracks[ids[i]]
		// Phase 1.5：roomengine real-by-provenance latch 生效 → 强制 real（保 nr≥1），跳过几何/sync 判定 +
		// 出生窗定性。confine（是否仍在 rest 区）已在 roomengine 侧判完，此处只消费结果。
		if o.ForceReal {
			t.rt.ForceReal()
			t.isRefl, t.fSep, t.fWallMargin, t.fRho, t.fLater = false, 0, 0, 0, false
			c.updateTrackTail(t, o, nowMs, disp[i])
			continue
		}
		// split 坐实(运动学 3-tick 赖锚点)是持续判决 → 即便出生窗后也须续喂 realness 排干 bR（否则冻结在窗末 PReal=1，
		// nr 顶 2）；其率不依赖几何/Coexist(见 belief.Update)，故跳过 reflectSep。窗内轨照旧走几何+sync 判定。
		inWindow := nowMs-t.birthMs <= c.p.MirrorWindowMs
		if inWindow || o.SplitConvicted {
			// reflectSep（最贵的穿墙求交）仅出生窗内 coexist>0=track==2 才算（成本：ghost 仅 track==2；
			// 孤轨 mEv=Coexist×(…)=0,跳过结果中性）；窗后的 split-convicted 靠 rcWSplit 率，不需几何。
			sep := 0.0
			wallMargin := 0.0
			if inWindow && coexistOf[i] > 0 {
				sep = reflectSep(o.X, o.Y, radar, walls, c.p.ReflSepCm) // 桶二：墙外反射裕度 cm（0=墙内/非反射）
				if sep > 0 {
					if wallMargin = sep / float64(c.p.WallScaleCm); wallMargin > 1 {
						wallMargin = 1
					}
				}
			}
			t.isRefl = sep > 0
			later := laterOf[i]
			t.rt.Update(belief.RealnessObs{
				Displaced:      math.Hypot(float64(o.X-t.birthX), float64(o.Y-t.birthY)) > float64(c.p.MoveCm),
				CoexistRho:     rhoOf[i],
				LaterBorn:      later,
				WallMargin:     wallMargin,
				Coexist:        coexistOf[i],
				SplitConvicted: o.SplitConvicted,
				DtMs:           nowMs - t.lastMs,
			})
			t.fSep, t.fWallMargin, t.fRho, t.fLater = sep, wallMargin, rhoOf[i], later // forensic
		} else {
			// 窗后：判定冻结，**不算 reflectSep**（省算力，最贵的穿墙求交）。仅守孤轨永 Real override：
			//   配对源离开 → Coexist==0 → 立即纠回 Real，防冻结成 Mirror 的轨变孤轨后 N_r 算 0。
			if !t.birthLogged {
				// 出生窗冻结那帧固化一次判定（ForceReal/reset 之前 = 5s 窗累积的几何+sync 裁决）。
				t.birthLogged = true
				pr := t.rt.PReal()
				verdict := "real"
				if pr < 0.5 {
					verdict = "ghost"
				}
				zap.L().Info("birth_verdict",
					zap.String("logic_id", ids[i]),
					zap.String("verdict", verdict),
					zap.Float64("p_real", pr), zap.Float64("p_mirror", t.rt.PMirror()),
					zap.Bool("is_refl", t.isRefl), zap.Float64("sep", t.fSep),
					zap.Float64("wall_margin", t.fWallMargin), zap.Float64("rho", t.fRho),
					zap.Bool("later_born", t.fLater),
					zap.Int("birth_x", t.birthX), zap.Int("birth_y", t.birthY),
					zap.Float64("door_d", t.doorD))
			}
			if coexistOf[i] == 0 {
				t.rt.ForceReal()
			}
			t.isRefl, t.fSep, t.fWallMargin, t.fRho = false, 0, 0, 0 // forensic 标冻结（不再喂几何）
		}
		c.updateTrackTail(t, o, nowMs, disp[i])
	}

	c.updateFuse(nowMs)
	c.noteMultiPerson()
}

// updateTrackTail 每 track 收尾：窗口路程 EMA 累计（fuse 同步证据）+ 位置/tick/末观测快照（下 tick 位移基线）。
func (c *TrackCensus) updateTrackTail(t *censusTrack, o TrackObs, nowMs int64, dispI float64) {
	if t.pathMs > 0 {
		t.pathAccum *= math.Exp(-float64(nowMs-t.pathMs) / fusePathTauMs)
	}
	t.pathAccum += dispI
	t.pathMs = nowMs
	t.x, t.y, t.lastTick, t.lastMs, t.lastObs = o.X, o.Y, c.tick, nowMs, o
}

// noteMultiPerson 维护"多人证据"连续 tick 数（解合并 hysteresis）+ no-silent-caps LOG。
// 单雷达见 ≥2 真人 = 本融合机制不覆盖的真多人；持续 fuseExitHystTick 才认（滤 realness 单帧 ghost 闪），
// 并在转换沿 LOG 一次（B §7.2：机制不覆盖不静默）。
func (c *TrackCensus) noteMultiPerson() {
	dev := map[string]int{}
	multi := false
	for id, t := range c.tracks {
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.PReal() >= 0.5 {
			dev[DevKey(id)]++
			if dev[DevKey(id)] >= 2 {
				multi = true
			}
		}
	}
	if multi {
		c.multiPersonStrk++
		if c.multiPersonStrk == fuseExitHystTick && !c.multiPersonLogd {
			zap.L().Info("fusion_exit_multiperson",
				zap.Int64("tick", c.tick),
				zap.Int("hysteresis_ticks", fuseExitHystTick))
			c.multiPersonLogd = true
		}
	} else {
		c.multiPersonStrk = 0
		c.multiPersonLogd = false
	}
}

// FuseStatus 跨雷达同人融合一对 track 的运动同步证据（forensic：同步检查可见化，不参与裁决）。
type FuseStatus struct {
	A, B  string  // 配对 logicID（跨设备）
	Agree float64 // 运动幅值同步 EMA [0,1]（≥fuseAgreeThresh 锁同人）
	Moves int     // 累计有效移动帧（证据量，≥fuseMinMoves 才允许判）
	Same  bool    // 已锁定判同人 → Nr 折叠
}

// FuseForensic 当前所有跨雷达同人候选对的运动同步状态（每 tick xray 可见：两雷达 track 同步检查到多少）。
func (c *TrackCensus) FuseForensic() []FuseStatus {
	out := make([]FuseStatus, 0, len(c.fuseLinks))
	for _, l := range c.fuseLinks {
		out = append(out, FuseStatus{A: l.a, B: l.b, Agree: l.agree, Moves: l.moves, Same: l.same})
	}
	return out
}

// fuseCandidate 触发/退出门：**恰 2 台 radar，每台仅 1 个在场真人 track**——唯一有风险场景
// （屋内仅1人却被2雷达各报1轨数成2人）。返回该对 (a,b) 与 ok。
//
//	退出（ok=false）：在场真人 ≠2 / 同设备（镜像，realness 管）/ 任一设备 ≥2 真人（确有多人）。
//	≥2人无独处风险→过数无害，不折叠也不算同步（省算）。按真人(PReal≥0.5)计数：1真人+1ghost 仍 1 人。
func (c *TrackCensus) fuseCandidate() (a, b string, ok bool) {
	var ids []string
	for id, t := range c.tracks {
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.PReal() >= 0.5 {
			ids = append(ids, id)
			if len(ids) > 2 {
				return "", "", false // >2 真人 → 确有多人，退出
			}
		}
	}
	if len(ids) != 2 || DevKey(ids[0]) == DevKey(ids[1]) {
		return "", "", false
	}
	return ids[0], ids[1], true
}

// updateFuse 触发配置下累计那对 track 的**窗口路程**同步证据，锁定同人（仅供 Nr 折叠）。
//
//	多雷达帧交错：任一帧只有一台 fresh、另一台 coasted(单帧位移=0)，故比**5s 窗累计路程**(recentPath)
//	而非单帧位移——同一人被两台看到，各自帧里积出的路程相近。
//	都静止(max 路程<thresh)不喂证据 → 保持已锁定(人躺床两台都不动)。
//	**锁定后不解锁**(一台退化/摔后冻轨 ≠ 两个人)——释放只靠 GC(退出触发配置 staleMs)。under-count FN-safe。
func (c *TrackCensus) updateFuse(nowMs int64) {
	a, b, ok := c.fuseCandidate()
	if ok {
		key := fusePairKey(a, b)
		link := c.fuseLinks[key]
		if link == nil {
			link = &fuseLink{a: a, b: b}
			if b < a {
				link.a, link.b = b, a
			}
			c.fuseLinks[key] = link
		}
		mvA, mvB := c.tracks[a].recentPath(nowMs), c.tracks[b].recentPath(nowMs)
		if mx := math.Max(mvA, mvB); mx >= fuseMoveThreshCm { // 有意义移动才喂同步证据
			ag := 1 - math.Abs(mvA-mvB)/mx
			link.agree = (1-fuseEMA)*link.agree + fuseEMA*ag
			link.moves++
			if link.moves >= fuseMinMoves && link.agree >= fuseAgreeThresh {
				link.same = true
			}
		}
		link.lastMs = nowMs
	}
	for k, l := range c.fuseLinks {
		if nowMs-l.lastMs > fuseStaleMs {
			delete(c.fuseLinks, k)
		}
	}
}

// Nr 房内真人数 = 本 tick 在场且 PReal≥0.5 的 track 数（realness 后验软阈；排 Mirror ghost）。
//
//	Blind 续存（摔倒静止降功率被滤的 lost-Fallen track 仍算 1 人，§G六②）**不靠 census TTL**——
//	架构师 2026-06-15 裁定「随 S^(i) 涌现，本刀不加 TTL」：N_r=Σ1[S^(i)∉{E,L}]（DBN-Zone-Room line57）
//	本就定义在 S^(i) 上，blind 续存计数由 step2 多 track 进 belief 主管线后经 S^(i) 转移自持涌现
//	（belief 自持非 staleness 补丁，呼应 §32 删 TTL）。单 faller 续存已由 belief 自持、且其计不计入
//	N_r 不改自身上报（≤1 人不折扣 / ≥55% 必报）。故本片只数在场，不在 census 层补续存窗。
func (c *TrackCensus) Nr() int {
	// lastObs.Online=false = 续算/lost track（① 时长外推，非真实观测）→ 不计人数（stillbox 续算与人数解耦）
	present := make(map[string]struct{}, len(c.tracks))
	for id, t := range c.tracks {
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.PReal() >= 0.5 {
			present[id] = struct{}{}
		}
	}
	n := len(present)
	// 折叠 latched 同人对（两端都在场真人）→ 数减 1，不动 track/belief/fire。link.same 只在"恰2雷达各1真人"
	// 配置下由 updateFuse 锁定，故 same 对必是干净的 2雷达1人。多人证据**持续** ≥hysteresis 才退出折叠
	// （滤 realness 单帧 ghost 闪，B §7.2 解合并防抖）。2雷达1人 → 至多一对。
	if c.multiPersonStrk < fuseExitHystTick {
		for _, l := range c.fuseLinks {
			if !l.same {
				continue
			}
			if _, ok := present[l.a]; !ok {
				continue
			}
			if _, ok := present[l.b]; !ok {
				continue
			}
			n--
			break
		}
	}
	return n
}

// RescuableCount 可救援数 = 在场真人(present ∧ PReal≥0.5) 中**不在床**者，应用同人折叠（与 Nr 同口径）。
//
//	inBed[id] 由 engine 按每 track belief S 边缘峰值是否 SBed 判定后传入（census 不依赖 belief 包，
//	只数计数）。躺床/睡着救不了摔倒者 → 不计入"能救人的人"。少算救援者 → C_FN↑ → 更激进 fire =
//	严格 FN-safe。**仅 forensic**（当前 C_FN 不门控 fire，见 belief/decide.go 文件头；doc/cfn-rescuable-design.md）。
func (c *TrackCensus) RescuableCount(inBed map[string]bool) int {
	present := make(map[string]struct{}, len(c.tracks))
	for id, t := range c.tracks {
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.PReal() >= 0.5 && !inBed[id] {
			present[id] = struct{}{}
		}
	}
	n := len(present)
	if c.multiPersonStrk < fuseExitHystTick {
		for _, l := range c.fuseLinks {
			if !l.same {
				continue
			}
			if _, ok := present[l.a]; !ok {
				continue
			}
			if _, ok := present[l.b]; !ok {
				continue
			}
			n--
			break
		}
	}
	return n
}

// TrackState 一条 track 对 engine 的读出（§57 步2：身份 + 末帧 obs + realness + 是否在场）。
type TrackState struct {
	LogicID      string   // tm 出生唯一身份（与 tm.TrackState.LogicID 同一标识）
	Obs          TrackObs // 在场=本帧 / 消失=末帧（喂 per-track 滤波发射；消失态走 blind 续存仅 Predict）
	PReal        float64  // 真人后验（ghost→低）；engine OR 聚合只算 PReal≥0.5 的真人 track
	Present      bool     // 本 tick 匹配到 raw（false=消失，engine 据此走 blind 续存 Predict-only）
	PMirror      float64  // 镜像后验（有真人源的 ghost）；forensic 观测
	IsReflection bool     // 桶二镜面几何判定（radar→ghost 连线穿墙取最近交点≥阈）；forensic 观测
	Sep          float64  // forensic：reflectSep cm（墙外反射裕度，0=墙内/非反射）
	WallMargin   float64  // forensic：WallMargin（mEv 墙外项）
	Rho          float64  // forensic：CoexistRho（同步移动强度）
	LaterBorn    bool     // forensic：成对后到（ghost 候选，破同步对称）
	DoorD        float64  // forensic：出生地→最近门 cm（设出生先验 bR₀；<0=无 enter 区）
}

// Tracks engine 读 census 已知的全部 track（在场 + 消失未 drop），按 LogicID 升序（确定性）。
// census 是身份/realness 单源（§59③）；engine 据此每 LogicID 跑一份滤波，不重做关联。
func (c *TrackCensus) Tracks() []TrackState {
	out := make([]TrackState, 0, len(c.tracks))
	for id, t := range c.tracks {
		out = append(out, TrackState{LogicID: id, Obs: t.lastObs, PReal: t.rt.PReal(), Present: t.lastTick == c.tick && t.lastObs.Online, PMirror: t.rt.PMirror(), IsReflection: t.isRefl,
			Sep: t.fSep, WallMargin: t.fWallMargin, Rho: t.fRho, LaterBorn: t.fLater, DoorD: t.doorD})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LogicID < out[j].LogicID })
	return out
}

// Drop 移除一条 track（engine 在其 S^(i) 被吸收到 Left/Empty 后调用；非 TTL，状态驱动）。
func (c *TrackCensus) Drop(logicID string) { delete(c.tracks, logicID) }

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

const radarSelfCrossCm = 10 // 交点距 radar <此 = radar 踩墙的自穿(t≈0)，丢弃（真反射体不会贴 radar 这么近）

// reflectSep 桶二镜像几何（§69 五步）→ 反射裕度 cm（0=非反射，喂 WallMargin 软发射）：
//	① ghost 在房间 bbox 内 → 0（真人在房内天然非反射）；
//	② radar→ghost 连线与各 wall 边求距 ghost 最近的交点 = mirror 点，排除 radar 自身处交点；
//	③ 连线不穿任何墙 → 0；④ mirror 点→ghost <sepCm → 0，FN-safe 偏 0。无墙 → 恒 0（§69 柱E）。
//
// ① 用边线并集 bbox 判房内、非逐墙 insideRect：walls 是 wallsFromPolygon 产的零厚度边线 rect，房内点
//   判不进任何单条边线 → insideRect 恒 false 护栏失效，房内真人误判反射。② radar 常装在墙边线上，
//   连线起点(t≈0)假性穿墙取全长当裕度——非真反射，按距 radar 阈丢弃。
func reflectSep(gx, gy int, radar Point, walls []Rect, sepCm int) float64 {
	if len(walls) == 0 {
		return 0
	}
	minX, maxX := walls[0].X1, walls[0].X1
	minY, maxY := walls[0].Y1, walls[0].Y1
	for _, w := range walls {
		minX, maxX = min(minX, w.X1, w.X2), max(maxX, w.X1, w.X2)
		minY, maxY = min(minY, w.Y1, w.Y2), max(maxY, w.Y1, w.Y2)
	}
	if gx >= minX && gx <= maxX && gy >= minY && gy <= maxY {
		return 0 // ① ghost 在房间 bbox 内 → 非反射
	}
	gxf, gyf := float64(gx), float64(gy)
	rxf, ryf := float64(radar.X), float64(radar.Y)
	best, found := math.MaxFloat64, false
	for _, w := range walls {
		for _, p := range segRectIntersections(rxf, ryf, gxf, gyf, w) {
			if math.Hypot(p.x-rxf, p.y-ryf) < radarSelfCrossCm {
				continue // 排除 radar 踩墙的自穿交点（t≈0）
			}
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
