package adapter

import (
	"math"
	"sort"

	"go.uber.org/zap"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// census.go — 房内多 track 人数 N_r（§G 主职：排 ghost，防 1 真人+1 影子被当 2 人 → C_FN 误折扣漏报）。
//   身份 = 最小作功距离 logicID（§G六）：新 track 发新号 / 平时跟随 / 相遇按最近距离保号。
//   realness = belief.RealnessTrack 2 态前向滤波(Real/Mirror，转移矩阵+软发射，device-room-zone.md §9)；
//     本层只译软证据喂之：出生地距门 D(§9.3①)、自主位移、co-existence ρ(速度同步度)、墙外反射裕度(桶二几何)、配对存在。
//   反证法(ghost_disproof_birth_certificate_spec)：realness 塌成 latch，判决 ⊥ 显示——
//     判决轴 RealProven(bool，出生证 latch，N_r/floor 门读)；显示轴 PReal(连续，按离 born 净距 R 分档，纯 FE)。

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
	rMax           float64  // 离 born 最远净距 cm（sticky max，纯显示分档：ghostDisplayPReal→PReal）
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
func (c *TrackCensus) Update(nowMs int64, obs []TrackObs, entrances []Rect) {
	c.tick++
	// 身份直接引用 tm 出生 LogicID（§G六 关联在 tm nearestAliveTrack 已做,census 不重造）；
	//   首次见的 LogicID 初始化 censusTrack（出生窗 realness 锚:doorD/birthXY/birthMs）。
	ids := make([]string, len(obs))
	for i, o := range obs {
		ids[i] = o.LogicID
		if _, ok := c.tracks[ids[i]]; !ok {
			dd := birthDoorDist(o.X, o.Y, entrances)
			c.tracks[ids[i]] = &censusTrack{rt: belief.NewRealnessTrack(), birthX: o.X, birthY: o.Y, birthMs: nowMs, doorD: dd, x: o.X, y: o.Y, lastMs: nowMs}
		}
	}

	// disp = 单帧位移幅值（cm），喂跨设备同人融合（frame-invariant）。
	disp := make([]float64, len(obs))
	for i, o := range obs {
		t := c.tracks[ids[i]]
		disp[i] = math.Hypot(float64(o.X-t.x), float64(o.Y-t.y))
	}
	// 反证法（ghost_disproof_birth_certificate_spec）：realness 塌成 latch 驱动。判决 ⊥ 显示：
	//   判决轴 = o.ForceReal（roomengine RealProven，裸 latch）→ ForceReal / 无证 → ForceGhost，**gate fire/N_r**。
	//   显示轴 = 无证轨按离 born 最远净距 R 分档喂 SetDisplay（纯 FE confidence，gate 任何决策皆无）。
	// 删旧 mirror 几何轴（reflectSep/WallMargin/sync/rcRealBase/rcWAuto）+ §G六 独处 force-real（那让空房 phantom 误火）。
	for i, o := range obs {
		t := c.tracks[ids[i]]
		if r := math.Hypot(float64(o.X-t.birthX), float64(o.Y-t.birthY)); r > t.rMax {
			t.rMax = r // 离 born 最远净距（sticky，仅供显示分档）
		}
		if o.ForceReal {
			t.rt.ForceReal() // 有出生证 → real（判决 latch + 显示满格）
		} else {
			t.rt.ForceGhost()                          // 无证 → ghost 判决（压过漂移 + 独处 force-real）
			t.rt.SetDisplay(ghostDisplayPReal(t.rMax)) // 显示按 R 分档（gate 无）
		}
		t.isRefl, t.fSep, t.fWallMargin, t.fRho, t.fLater = false, 0, 0, 0, false // 几何轴退役，forensic 恒零
		if !t.birthLogged {
			t.birthLogged = true
			verdict := "ghost"
			if o.ForceReal {
				verdict = "real"
			}
			zap.L().Info("birth_verdict", zap.String("logic_id", ids[i]), zap.String("verdict", verdict),
				zap.Int("birth_x", t.birthX), zap.Int("birth_y", t.birthY), zap.Float64("door_d", t.doorD))
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
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.RealProven() {
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
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.RealProven() {
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
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.RealProven() {
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
		if t.lastTick == c.tick && t.lastObs.Online && t.rt.RealProven() && !inBed[id] {
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
	RealProven   bool     // 判决轴：出生证 latch（engine floor/N_r 占用门读它，替代旧 PReal≥0.5）
	PReal        float64  // 显示轴：连续 confidence（纯 FE ×100；gate 任何决策皆无——判决读 RealProven）
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
		out = append(out, TrackState{LogicID: id, Obs: t.lastObs, RealProven: t.rt.RealProven(), PReal: t.rt.PReal(), Present: t.lastTick == c.tick && t.lastObs.Online, PMirror: t.rt.PMirror(), IsReflection: t.isRefl,
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

// ghostDisplayPReal 无证轨显示 confidence（spec 判决 ⊥ 显示：纯 FE ×100，gate 任何决策皆无）。
// 按离 born 最远净距 R 分档：≤50 GLUED=0 / 50-100 ramp 线性→0.5 / 100-200 HOLD=0.5 / ≥200 WALKOUT=1（同刻 RealProven→true）。
func ghostDisplayPReal(R float64) float64 {
	switch {
	case R <= 50:
		return 0
	case R < 100:
		return (R - 50) / 50 * 0.5
	case R < 200:
		return 0.5
	default:
		return 1
	}
}
