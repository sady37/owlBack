package engine

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// engine.go — 单房 roomengine 主循环（build order ③ + §57 步2 多 track 进 belief 主管线）。
// §A.3② 隐维复制：每条 track 各跑一份独立 9·2^|B| 滤波（并排，不做笛卡尔积、不进 J 基数）；
//   census 是身份/realness/人数单源（§59③：census 出 N_r，Filter 出人态，不双算）。
// 四隐轴：S/B ← adapter 译 per-track 雷达 + 房共享床证据；realness ← census 每 track PReal（折 SFallen 发射）；
//   neighbor ← rhoXroom（W3.4 跨房 census 读出；单房=0）。
// 房间决策（§60 架构师拍）：OR over 真人 track（PReal≥0.5）——任一真人 Fallen 过阈 → 房间 fire，报到房间。
// blind 续存（§60 反-TTL）：track 消失 → 其 Filter 仅 Predict 自持（S^(i) 留 Fallen）→ 告警连续；
//   消失 track 的 S^(i) 被吸收到 {Left,Empty} → 人确认离场 → drop（状态驱动，非 TTL）。

// absorbedThresh 消失 track 的 S 边缘 P(Left)+P(Empty) ≥ 此 = 确认离场 → drop（form-anchor，留 oracle）。
const absorbedThresh = 0.9

// lostFireThresh = lost 专用高阈：blind 时 belief 自然演化到此 → fire（present 摔走 firmware/dbn_mode，与此分开）。
//
//	二义 lost（pF 自然到不了 0.85）不在此发，交 Unit UD timer 兜底。form-anchor 留 oracle。
const lostFireThresh = 0.85

// exitFlipLogOdds 离房翻转阈 = logit(0.85)：注入的 SLeft 对数似然 ≥ 此 → 确认离房 → Unit timer cancel。
//
//	与 track_manager 离房公式 V 的 0.85 翻转点同源（ExitRoom 硬证据 log-odds=8 ≫ 此，单发即翻）。
var exitFlipLogOdds = math.Log(0.85 / 0.15)

// arrivalConfirmFrames hand-off 信号「确认真人」所需连续在场真人帧数（§64 噪声防线，form-anchor）。
//
//	gained/lost 不用瞬时 PReal≥0.5——cd2b 式噪声尖峰 >AssocCm 瞬拆**新 logicID**（PReal=1 未及去 ghost），
//	瞬时判会把它当合法 GainedReal → 假 hand-off 整流真摔成 Left → 漏报。要求连续 K 帧真人=确认重现/离场，
//	噪声 churn（每帧新 id、1 帧即逝）/瞬时尖峰永不达 K → 不造假 lost/gain。
const arrivalConfirmFrames = 3

// Frame 单帧引擎输出（房间代表 probe + 房间 OR 聚合裁决 + 跨房 hand-off 信号）。
type Frame struct {
	Probe    belief.FrameProbe
	Decision belief.Decision
	// 步4 跨设备 hand-off 信号（Unit 编排器消费，§A 守恒+时间窗）：
	LostReal          bool    // 本帧一条真人 track 消失（上帧在场真人、本帧不在）= hand-off 源候选
	LostAtMs          int64   // 消失轨的 loss 锚：已冻结→LostFirst(still-box 起点)，未冻结→last-observed（非本帧 coast 判定时刻）→ Unit Δ centering（gain−LostAtMs）
	LostExited        bool    // 消失 track 里有人本人 ExitRoom 过门（按 track_id 反查）= Unit D/UD timer cancel（人走了无人会摔）
	GainedReal        float64 // 本帧新现真人 track 的去 ghost 后验（0=无）= hand-off 宿候选（守恒重现）
	GainedFromSleepad bool    // 该 gain 来自 sleepad-only 房合成 track（走+躺慢接力）→ rho 用 sleepad 慢核
	// forensic（全切片观测，不参与裁决）：
	PresentCount int                  // 本帧在场 track 数（§61 共存源/消费门控判据）
	Tracks       []TrackForensic      // 每 track 内部量（realness/ghost/消费门控/per-track 裁决）
	FuseSync     []adapter.FuseStatus // forensic：跨雷达同人运动同步对（agree/same/moves，可见双雷达同步检查）
	// FiredLogicIDs 本帧 fall fire 的 LogicID：回传 track_manager 复位 still-box（跨层 StillSec 在 tm），
	//   与本帧已就地复位的 belief 配套——fall fire = 该 track 推断 episode 结束 → 从 0 热机重判。
	FiredLogicIDs []string
	// DroppedLogicIDs 本帧 belief 状态驱动 drop（确认离场/空：!Present ∧ SLeft+SEmpty≥absorbedThresh）的 LogicID：
	//   回传 track_manager 立即 evict，停止 12s coast 期对已离场 track 的 re-feed（否则 census 每帧重发
	//   新 logicID = churn）。与 FiredLogicIDs 互斥（fire⇒SFallen 高⇒不进 drop）。闪失重捕的 coast 不受影响
	//   （无离场证据 belief 不 drop）。身份用 LogicID（多雷达同房 firmware track_id 撞号已根治）。
	DroppedLogicIDs []string
	// ExitFirsts 本帧发 ExitRoom(HardExited)的离开候选：ExitFirstMs=通往该次 ExitRoom 的连续 Enter 区驻留起点
	//   （Unit 守恒配对用，优先认领 neighbor arrival，consume 后 lost 才吃剩下的）。
	ExitFirsts []ExitDeparture
	// LostDepartures 本帧确认真人消失的离开候选(LostFirst 锚 + logicID)：Unit 守恒配对宿主，matched → purge。
	LostDepartures []LostDeparture
}

// ExitDeparture 一条门口离开候选（ExitRoom 结账）。
type ExitDeparture struct {
	LogicID     string
	ExitFirstMs int64 // 连续 Enter 区驻留首帧（0=退时不在 Enter 区，回退 ExitRoomMs）
	ExitRoomMs  int64 // ExitRoom(HardExited)确认时刻
}

// LostDeparture 一条冻结/丢轨离开候选（本帧确认真人消失）。
type LostDeparture struct {
	LogicID     string
	LostFirstMs int64 // 冻结首帧(still-box 起点)否则 last-observed
}

// TrackForensic 单 track 的 DBN 内部量（forensic 暴露，X 光全切片用，不参与裁决）。
type TrackForensic struct {
	LogicID       string
	Present       bool
	RealProven    bool      // 判决轴：出生证 latch（SetTrackPReal 二值回灌 tm 决策门读它，替代旧 PReal≥0.5）
	PReal         float64   // 显示轴：连续 confidence（纯 FE ×100；gate 任何决策皆无）
	PMirror       float64   // 镜像后验（forensic/FE）
	IsReflection  bool      // 桶二镜面几何判定
	PFallen       float64   // per-track P^F
	Fire          bool      // per-track 裁决（持续≥T_hold）
	Band          string    // per-track 档（report/no/tie/indeterminate）
	X, Y          int       // forensic：末帧 canvas 坐标
	Sep           float64   // forensic：reflectSep cm（墙外反射裕度）
	WallMargin    float64   // forensic：mEv 墙外项
	Rho           float64   // forensic：CoexistRho 同步移动强度
	LaterBorn     bool      // forensic：成对后到
	DoorD         float64   // forensic：出生地→最近门 cm（设出生先验 bR₀）
	RepeatR       float64   // forensic：重复摔残余 R（§J；前科强度，>0=有近期摔史）
	SelfRecovered bool      // forensic：本帧 fall episode 刚结束（起身自救）= 记录线 self_recovered
	StillSec      float64   // forensic：still-box 总时长秒（FloorGuard 纯计时器输入；调 floor/still 看这个）
	SMarg         []float64 // forensic：本 track 自己的 S 9 态边缘后验（看 D523 自身 SBed/SFallen，不只房级代表）
	Covers        []float64 // forensic：本 track 所属设备 per-bed covers(MM)（D523=0 验证跟床解耦）
}

// Room 单房多 track 引擎：每 logicID 一份 belief 滤波 + 裁决器；census 管身份/realness/人数。
type Room struct {
	model *belief.Model
	js    *belief.JointSpace // 共享读出/发射空间（各 track 滤波同维，idx 一致）
	// per-(设备×床) 床耦合：每雷达设备(uid_last4)一份 Coupling/Emission/κ-ledger，懒建（仿 filters）。
	//   covers 从 MM 来（设备画了该床=1，否则 0）→ D523 没画床→covers=0→κ 冷启 0→sleepad 床读数不灌进 D523。
	//   defGeom = 房级回退几何(covers=1，与改造前逐 tick 等价)：单雷达房/sleepad-only 合成 track/未注册设备走它。
	defGeom       []belief.BedGeom
	geoms         map[string][]belief.BedGeom // uid4 → 该设备 per-bed 几何（bootstrap 注入；无=用 defGeom）
	cps           map[string]*belief.Coupling // uid4 → 该设备耦合（懒建）
	ems           map[string]*belief.Emission // uid4 → 该设备发射（懒建）
	ledgers       map[string][]bedKappaLedger // uid4 → 该设备 ① 强化腿 15s 窗台账（懒建）
	census        *adapter.TrackCensus
	filters       map[string]*belief.Filter       // 每 logicID(tm 出生 string) 一份 S/B 联合滤波（§A.3② 隐维复制）
	deciders      map[string]*belief.Decider      // 每 logicID 一份持续计时裁决
	floorGuards   map[string]*belief.FloorGuard   // 每 logicID 一份 FN-safe 兜底（总时长 floor，契约其十五）
	escalators    map[string]*RepeatFallEscalator // 每 logicID 一份重复摔残余器（§J；第一次归 firmware，只提前重复摔）
	nb            int
	p             adapter.Params
	lastMarg      belief.Vector    // 末帧房间代表 track 的 S 边缘（MarginalS 读出）
	realStreak    map[string]int   // 每 logicID 连续在场真人帧数（步4 hand-off：confirmed=streak≥K，抗噪声 churn）
	prevConfirmed map[string]bool  // 上 tick 已确认真人(streak≥K)的 logicID 集（算 lost）
	lastSeenMs    map[string]int64 // 每 logicID 末次 fresh-present 时戳（census Present=本 tick 匹配）→ lost 的 last-observed centering
	lostAnchorMs  map[string]int64 // 每 logicID 末次快照的 StillBoxRunStart（>0=已冻结的 LostFirst 冻结首帧）→ lost 锚前提到此，避免冻结 coast 把 last-observed 拖后致 hand-off Δ 误判反向
	enterSinceMs  map[string]int64 // 每 logicID 当前连续处于 Enter 区(fi.Entrances 点测)的首帧 ms（离开 Enter 区清 0）→ ExitRoom 结账时=ExitFirst 锚
	fallLatched   map[string]bool  // 每 logicID 已证摔 latch（SFallen 到过 0.85 或见 pose=fall）→ hand-off 注入永免疫（§7.7 v2 ③）
	// F1 独居连续计时（跨 tick，与 realStreak 同层）：真人占用==1 连续起点 ms（0=当前非独居）。
	//   占用判据 = PReal≥0.5 ∧ S∉{Empty,Left}（含 blind 续存的 faller，filter 后的 MarginalS），
	//   **非** census.Nr() present-only——否则独居者摔进 blind 时占用掉 0 误清零计时（lost-fall FN）。
	aloneStreakStartMs int64
	notSoloFrames      int // 连续占用≠1 帧数（重置抗抖动柱②：≥arrivalConfirmFrames 才清 alone-streak）
}

// NewRoom 建单房引擎。geom = 床几何（adapter.BedGeoms 从 layout 派生）；nb = 床数。
// lost-fall 兜底（D/UD timer）已上移 Unit 层（unit 级"任何 track 取消"+ per-room_type deadline）。
func NewRoom(geom []belief.BedGeom, nb int) *Room {
	return &Room{
		model:         belief.DefaultModel(),
		js:            belief.NewJointSpace(nb),
		defGeom:       geom,
		geoms:         map[string][]belief.BedGeom{},
		cps:           map[string]*belief.Coupling{},
		ems:           map[string]*belief.Emission{},
		ledgers:       map[string][]bedKappaLedger{},
		census:        adapter.NewTrackCensus(adapter.DefaultTrackCensusParams()),
		filters:       map[string]*belief.Filter{},
		deciders:      map[string]*belief.Decider{},
		floorGuards:   map[string]*belief.FloorGuard{},
		escalators:    map[string]*RepeatFallEscalator{},
		nb:            nb,
		p:             adapter.DefaultParams(),
		realStreak:    map[string]int{},
		prevConfirmed: map[string]bool{},
		lastSeenMs:    map[string]int64{},
		lostAnchorMs:  map[string]int64{},
		enterSinceMs:  map[string]int64{},
		fallLatched:   map[string]bool{},
	}
}

// SetDeviceGeom bootstrap 注入某雷达设备(uid_last4)的 per-bed 几何（covers 从 MM：画了该床=1 否则 0）。
// 未注入的设备/合成 track 走 defGeom（covers=1=改造前行为）。须在任何 Tick 前调用。
func (r *Room) SetDeviceGeom(uid4 string, geom []belief.BedGeom) { r.geoms[uid4] = geom }

// geomOf uid4 对应几何，无注入→defGeom。
func (r *Room) geomOf(uid4 string) []belief.BedGeom {
	if g := r.geoms[uid4]; g != nil {
		return g
	}
	return r.defGeom
}

// byDev 取/懒建该设备的 Coupling+Emission（κ 冷启自 geomOf 的 covers·onbed）。
func (r *Room) byDev(uid4 string) (*belief.Coupling, *belief.Emission) {
	if cp := r.cps[uid4]; cp != nil {
		return cp, r.ems[uid4]
	}
	g := r.geomOf(uid4)
	cp := belief.NewCoupling(g)
	em := belief.NewEmission(g)
	r.cps[uid4] = cp
	r.ems[uid4] = em
	r.ledgers[uid4] = make([]bedKappaLedger, r.nb)
	return cp, em
}

// bedKappaLedger ① 强化腿 per-bed 15s deferred 窗：检测 sleepad/radar 床事件同向共跳。
type bedKappaLedger struct {
	sleepadPrev belief.BedReading
	radarPrev   bool
	pend        bool
	pendMs      int64
	pendSDir    int // +1→InBed / -1→LeftBed
	pendRDir    int // +1 enter / -1 leave
}

const kappaWindowMs = 15000 // ① 强化窗 K（case1/2 的 15s）

// radarInBed ① 的 radar 半：FwAreaID==床。**不 gate Online**（钉① lost 边界）——lost-coasting track 的
// FwAreaID 是 M×N 冻结末值；radar 半冻结、sleepad 半实时把关：人真离床→sleepad LeftBed→agree/matched=F
// →κ 衰减，冻结 N 不会错误维持。evicted track 不在 fi.Tracks，自然不计。
func deviceRadarInBed(uid4 string, fi adapter.FrameInput, j int) bool {
	if j >= len(fi.BedAreaIDs) || fi.BedAreaIDs[j] == 0 {
		return false
	}
	for _, t := range fi.Tracks {
		if adapter.DevKey(t.LogicID) == uid4 && t.FwAreaID == fi.BedAreaIDs[j] {
			return true
		}
	}
	return false
}

// maintainKappaDev ① 维持腿（per-frame 弱 γ）：持续共态 agree=sIn∧rIn → 熟睡持续在床托住 κ 抗衰减。
//
//	per-device：radar 半只看本设备 track；covers≤0 的床（本设备没画）跳过—κ 恒 0 不动（D523 解耦）。
func (r *Room) maintainKappaDev(uid4 string, cp *belief.Coupling, geom []belief.BedGeom, fi adapter.FrameInput) {
	live := make([]bool, r.nb)
	agree := make([]bool, r.nb)
	for j := 0; j < r.nb; j++ {
		if j >= len(geom) || geom[j].Covers <= 0 {
			continue
		}
		sOn := j < len(fi.Sleepads) && fi.Sleepads[j].Present
		sIn := sOn && fi.Sleepads[j].Reading == belief.BedInBed
		rIn := deviceRadarInBed(uid4, fi, j)
		live[j] = sOn && (sIn || rIn) // both-vacant→冻结；sleepad 离线→不动非衰减
		agree[j] = sIn && rIn         // 持续共占→维持抬；真离床 sIn=F→agree=F→弱衰减
	}
	cp.MaintainKappa(agree, live)
}

// eventKappaDev ① 强化腿（15s deferred 窗，强 γ）：同向共跳→强抬（时间绑同人）。钉②：只"sleepad 在床
// ∧ radar 持续别床"才强降；"sleepad 跳入 + radar 持续在床(agree)"→不强更新，交维持腿托（不误判矛盾）。
//
//	per-device：本设备 track + 本设备台账；covers≤0 床跳过（D523 不参与）。
func (r *Room) eventKappaDev(uid4 string, cp *belief.Coupling, geom []belief.BedGeom, ledger []bedKappaLedger, fi adapter.FrameInput) {
	matched := make([]bool, r.nb)
	live := make([]bool, r.nb)
	for j := 0; j < r.nb; j++ {
		if j >= len(geom) || geom[j].Covers <= 0 {
			continue
		}
		L := &ledger[j]
		cur, sOn := belief.BedNoReport, false
		if j < len(fi.Sleepads) {
			cur, sOn = fi.Sleepads[j].Reading, fi.Sleepads[j].Present
		}
		sDir := 0
		if cur != L.sleepadPrev {
			if cur == belief.BedInBed {
				sDir = +1
			} else if cur == belief.BedLeftBed {
				sDir = -1
			}
		}
		L.sleepadPrev = cur
		rNow := deviceRadarInBed(uid4, fi, j)
		rDir := 0
		if rNow != L.radarPrev {
			if rNow {
				rDir = +1
			} else {
				rDir = -1
			}
		}
		L.radarPrev = rNow
		if sDir != 0 || rDir != 0 {
			if !L.pend {
				L.pend, L.pendMs = true, fi.NowMs
			}
			if sDir != 0 {
				L.pendSDir = sDir
			}
			if rDir != 0 {
				L.pendRDir = rDir
			}
		}
		if L.pend && fi.NowMs-L.pendMs >= kappaWindowMs {
			coTrans := L.pendSDir != 0 && L.pendRDir != 0 && L.pendSDir == L.pendRDir
			sInNow := sOn && cur == belief.BedInBed
			contradiction := sInNow && !rNow // 钉②：仅 sleepad 在床∧radar 持续别床=真矛盾
			matched[j] = coTrans
			live[j] = coTrans || contradiction // 钉②：sleepad 跳入+radar 持续在床→双 F→no-op 交维持腿
			*L = bedKappaLedger{sleepadPrev: cur, radarPrev: rNow}
		}
	}
	cp.UpdateKappa(matched, live)
}

// updateKappa 每帧驱动所有在场设备的 κ 两腿（维持+强化）。本帧有 track 的设备各自更新（仿 filters 懒建）；
//
//	按 devKey 去重，covers≤0 床在两腿内跳过。
func (r *Room) updateKappa(fi adapter.FrameInput) {
	seen := map[string]bool{}
	for _, t := range fi.Tracks {
		uid4 := adapter.DevKey(t.LogicID)
		if seen[uid4] {
			continue
		}
		seen[uid4] = true
		cp, _ := r.byDev(uid4)
		geom := r.geomOf(uid4)
		r.maintainKappaDev(uid4, cp, geom, fi)
		r.eventKappaDev(uid4, cp, geom, r.ledgers[uid4], fi)
	}
}

// applyLeftBedOpen ③：把 LeftBed→SOpenFloor/SBlindRest 的 S 轴满幅似然注入 logPhi（按 a_j 归属）。
func (r *Room) applyLeftBedOpen(cp *belief.Coupling, logPhi belief.JointVector, obs belief.Observation, gxy []float64, present bool) {
	lb := cp.LeftBedOpenLogS(obs, gxy, present)
	for s := range lb {
		if lb[s] != 0 {
			r.js.AddLogToS(logPhi, belief.State(s), lb[s])
		}
	}
}

type trackResult struct {
	d        belief.Decision
	pF       float64
	lam      float64
	eligible bool // 参与房间 OR 的资格：有共存源时须真人(PReal≥0.5 排 ghost)；无共存源=孤轨→永发(§61)
	f        *belief.Filter
	cp       *belief.Coupling // 该 track 所属设备的耦合（probe κ 快照用代表 track 自己的设备）
}

// Tick 一帧推进。rhoXroom（neighbor，单房=0）。每条 track 各跑一份滤波，房间 OR 聚合真人 fall。
func (r *Room) Tick(fi adapter.FrameInput, handoffL float64) Frame {
	r.census.Update(fi.NowMs, fi.Tracks, fi.Entrances)
	r.updateKappa(fi) // ① κ 两腿 per-device（维持 per-frame 弱 γ + 强化 15s 窗 强 γ，同向共跳）
	online := adapter.Online(fi)
	nr := r.census.Nr()
	// 独居连续分钟用**上 tick 末**的 streak 状态（占用 streak 在本 tick 末更新，同 realStreak:208）→ 1 帧滞后，
	//   30min 饱和量可忽略，且与 hand-off streak 时序一致。
	rc := adapter.BuildRiskContext(fi, nr, r.aloneMinAsOf(fi.NowMs))

	var results []trackResult
	var forensic []TrackForensic
	tracks := r.census.Tracks()
	presentCount := 0 // 本帧在场 track 数 = 共存源判据（§61 消费门控，用 raw 在场数非 N_r 防循环）
	for _, ts := range tracks {
		if ts.Present {
			presentCount++
		}
	}

	curReal := map[string]float64{} // 本 tick 在场真人(RealProven) logicID→PReal（算 lost/gained）
	var dropIDs []string            // 本帧状态驱动 drop 的 LogicID（dropTrack + 回传 tm evict，停 re-feed churn）
	var exitFirsts []ExitDeparture  // 本帧发 ExitRoom 的门口离开候选（ExitFirst 结账 → Unit 守恒配对）
	var firedLogicIDs []string      // 本帧 fall fire 的 LogicID（就地复位 belief + 回传 tm 复位 still-box）
	realOccupancy := 0              // F1 本 tick 真人占用数（PReal≥0.5 ∧ S∉{E,L}）→ 末更 alone-streak
	lostExited := false             // 消失 track 里有人本人 ExitRoom 过门（按 track_id 反查）→ Unit timer cancel
	for _, ts := range tracks {
		if ts.Present && ts.RealProven {
			curReal[ts.LogicID] = ts.PReal
		}
		if ts.Present {
			r.lastSeenMs[ts.LogicID] = fi.NowMs                                             // fresh-present（census 本 tick 匹配）→ lost 的 last-observed centering 参考
			r.lostAnchorMs[ts.LogicID] = stillOnsetMs(fi.NowMs, ts.Obs.RadarTrack.StillSec) // 冻结轨=LostFirst 冻结首帧(still-box 起点)；未冻结=0 → lost 回退 last-observed
			if inAnyRect(ts.Obs.X, ts.Obs.Y, fi.Entrances) {
				if r.enterSinceMs[ts.LogicID] == 0 {
					r.enterSinceMs[ts.LogicID] = fi.NowMs // 进 Enter 区首帧 → 连续驻留起点(ExitRoom 结账=ExitFirst)
				}
			} else {
				r.enterSinceMs[ts.LogicID] = 0 // 离开 Enter 区 → 断连续
			}
			if ts.Obs.RadarTrack.Pose == poseFall {
				r.fallLatched[ts.LogicID] = true // 见 pose=fall → 永久免疫 hand-off 注入（§7.7 v2 ③ latch-first）
			}
		}
		f := r.filters[ts.LogicID]
		dec := r.deciders[ts.LogicID]
		fg := r.floorGuards[ts.LogicID]
		esc := r.escalators[ts.LogicID]
		if f == nil {
			f = belief.NewFilter(r.model, r.nb) // 出生：新 logicID 起一份滤波（隐维复制）
			dec = belief.NewDecider()
			fg = belief.NewFloorGuard()
			esc = NewRepeatFallEscalator()
			r.filters[ts.LogicID], r.deciders[ts.LogicID], r.floorGuards[ts.LogicID], r.escalators[ts.LogicID] = f, dec, fg, esc
		}

		cp, em := r.byDev(adapter.DevKey(ts.LogicID)) // per-(设备×床) 耦合+发射：covers 从 MM（D523 covers=0 跟床解耦）

		var logPsi, logPhi belief.JointVector
		var obs belief.Observation
		exitL := 0.0 // 离房对数几率（per-track）：present=0；消失态由 ExitLogOdds 算，floor 撤销(②)复用
		if ts.Present {
			obs = adapter.BuildObservation(ts.Obs.RadarTrack, fi.Sleepads, fi.Beds, fi.BedAreaIDs, r.p, fi.Census.Night)
			gxy := adapter.Gxy(ts.Obs.RadarTrack, fi.Beds, r.p)
			logPsi = cp.LogPsi(r.js, gxy)
			logPhi = em.LogPhi(r.js, obs)
			r.applyLeftBedOpen(cp, logPhi, obs, gxy, true) // ③ 在场 → SOpenFloor
			// present 静止 ghost「已离房」：真人离房后镜面/家具残留 z≡0 轨，被 room-empty 信号 + 门距划 SLeft
			//   → 压 floor 不误火 + 攒 SLeft 丢轨即 absorbed-drop。硬门(Δz/pose)在 track_manager 否决真摔。
			//   🔴 latch-first：已证摔绝不划（与 hand-off 同纪律）；不设 lostExited（非真人过门）。
			if fi.GhostLeftLogOdds != nil && !r.fallLatched[ts.LogicID] {
				if exitL = fi.GhostLeftLogOdds(ts.LogicID, fi.NowMs); exitL > 0 {
					r.js.AddLogToS(logPhi, belief.SLeft, exitL)
				}
			}
		} else {
			// 消失态：雷达轴中性（RadarOnline=false），**接触轴(sleepad)仍应用**——在床 InBed→SBed
			//   保护睡眠者；LeftBed→B vac→Ψ 放行 SFallen。belief **自然演化**（无人工 ramp，已作废）：
			//   loss 时已高 pF（已确认摔）→ 自然到 0.85 即发；二义 lost（pF~0.5）不自爬 → 交 Unit UD timer 兜底。
			// Online=false 雷达 pose/z/area redirect 中性，但带续算 StillSec+AreaType(① 冻结坐标)→
			//   emission still 高斯 CDF 在消失态仍喂(人摔进盲区/爬出后持续静止的异常度)。
			obs = adapter.BuildObservation(adapter.RadarTrack{
				Online: false, StillSec: ts.Obs.RadarTrack.StillSec,
				AreaType: ts.Obs.RadarTrack.AreaType, RoomType: ts.Obs.RadarTrack.RoomType,
				InChair: ts.Obs.RadarTrack.InChair, ChairMu: ts.Obs.RadarTrack.ChairMu, ChairSigma: ts.Obs.RadarTrack.ChairSigma, ChairMaxSit: ts.Obs.RadarTrack.ChairMaxSit,
				BathMu: ts.Obs.RadarTrack.BathMu, BathSigma: ts.Obs.RadarTrack.BathSigma, BathMaxSit: ts.Obs.RadarTrack.BathMaxSit,
			}, fi.Sleepads, fi.Beds, fi.BedAreaIDs, r.p, fi.Census.Night)
			logPhi = em.LogPhi(r.js, obs)
			r.applyLeftBedOpen(cp, logPhi, obs, adapter.Gxy(ts.Obs.RadarTrack, fi.Beds, r.p), false) // ③ lost → SBlindRest（gxy 用冻结末位）
			// 离房证据注入 SLeft 对数似然（同一通道两源）：① ExitRoom 硬 + trend+np 软（track_manager）；
			//   ② **hand-off 软 ExitRoom**（§7.7 v2 矩形核，Unit 算 handoffL）——人挪去邻房非本房真摔。
			//   抬 SLeft → 压 pF（不 fire）+ ≥flip 即压 floor 兜底 + 够强 durable purge（absorbed-drop 撤轨）。
			//   🔴 latch-first（③）：已证摔（fallLatched）绝不注入 hand-off → 真摔照常 floor 兜底，不被软清。
			if fi.ExitLogOdds != nil {
				exitL = fi.ExitLogOdds(ts.LogicID, fi.NowMs)
			}
			if handoffL > 0 && !r.fallLatched[ts.LogicID] {
				exitL += handoffL
			}
			if exitL > 0 {
				r.js.AddLogToS(logPhi, belief.SLeft, exitL)
				if exitL >= exitFlipLogOdds {
					lostExited = true // ≥flip → Unit timer cancel（离房/接力确认）
				}
			}
		}
		// 消失态：雷达中性 + 接触轴；present 全发射。Predict 自持，无 TTL（hand-off 已改 SLeft 注入，不进 Predict）。
		f.Step(fi.NowMs, online, logPsi, logPhi)

		mS := f.Space().MarginalS(f.Alpha())
		// F1 真人占用：RealProven（判决轴，排 ghost）∧ S∉{Empty,Left}（含 blind 续存的 faller，未 absorbed 离场）。
		//   摔进 blind 时 S=Fallen∉{E,L} → 仍计占用 → 独居计时不被摔倒清零（present-only 的 Nr() 会误清）。
		if ts.RealProven && mS[belief.SEmpty]+mS[belief.SLeft] < absorbedThresh {
			realOccupancy++
		}

		pF := f.Space().PFallen(f.Alpha())
		if pF >= lostFireThresh {
			r.fallLatched[ts.LogicID] = true // SFallen 到过 confirm → 永久免疫 hand-off 注入（§7.7 v2 ③ latch-first）
		}
		lam := belief.ComputeLambda(f.Space(), neutralIfNil(logPsi, r.js), neutralIfNil(logPhi, r.js))
		d := dec.Step(fi.NowMs, pF, lam, rc)
		// lost fire（自然 belief）：blind 时 belief 自然演化到 0.85 → fire "lost"（loss 时已确认摔的 carry-over）。
		//   二义 lost（pF<0.85）不在此发，交 floor 总时长兜底 / Unit UD timer。三 cancel（离房趋势/handoff→SLeft/在床→SBed）
		//   经压低 P^F 使到不了 0.85 实现；present 摔走 firmware/dbn_mode（不进此分支=结构免疫）。
		if !ts.Present {
			// 离房抑制已由上方 SLeft 注入承担（抬 SLeft→压 pF）：朝门走/过门 → pF 到不了阈 → 不 fire。
			if pF >= lostFireThresh {
				d.Fire, d.Band = true, "lost"
			} else {
				d.Fire = false
			}
		}
		// FN-safe 兜底 floor（契约其十五）：总时长 ≥ T_floor(按 area) 且无正向休息证据 → 强制 fire。
		//   置于 lost fire **之后**：floor 是独立兜底腿（数时长不看 pF），不被消失态 pF<0.85 的 d.Fire=false 压回。
		//   ② 解锁消失态：present 用 obs；lost 用 track 续算的 StillSec(① 冻结坐标累积)构造 floorObs
		//   —— 1Hz 固件下精度做不了 lost 判定，靠时长累积兜住摔进盲区/pose 错报的真摔。
		//   撤销(人走了不兜底)：exitL≥flip(ExitRoom 硬证据 / trend×np) → 不 floor-fire(③复用 SLeft 注入)。
		floorObs := obs
		if !ts.Present {
			floorObs = adapter.BuildObservation(ts.Obs.RadarTrack, fi.Sleepads, fi.Beds, fi.BedAreaIDs, r.p, fi.Census.Night)
		}
		if fg.Step(floorObs) && exitL < exitFlipLogOdds {
			d.Fire = true
			if d.Band == "" || d.Band == "no" || d.Band == "indeterminate" {
				d.Band = "floor"
			}
		}
		// 重复跌倒提前报（§J）：第一次/孤立摔归 firmware（escalator R_prior>0 闸天然不触发）；
		//   只在"刚摔过又摔"时抢在 firmware 前 fire。仅 present 帧（pose 须 live；消失态 pose 陈旧不喂）。
		repeatR := 0.0
		selfRecovered := false
		if ts.Present {
			early, recovered := esc.Step(fi.NowMs, ts.Obs.RadarTrack.Pose, ts.RealProven, ts.Obs.RadarTrack.AreaType)
			repeatR = esc.Residual()
			selfRecovered = recovered
			if early {
				d.Fire = true
				if d.Band == "" || d.Band == "no" || d.Band == "indeterminate" {
					d.Band = "repeat"
				}
			}
		}
		// 无雷达房硬闸：sleepad-only 测不了姿态 → 永不出 Fall（合成 bed-track 的 Lying 是占用载体非真摔）。
		if fi.RadarLess {
			d.Fire = false
		}
		// fall fire 后该 track 推断 episode 结束（fired ⇒ 非 absorbed-drop，互斥）：记下供下方 belief 就地复位
		//   + LogicID 回传 tm 复位 still-box，从 0 热机重判（不动 census 身份/realness、不动 repeat 前科）。
		if d.Fire {
			firedLogicIDs = append(firedLogicIDs, ts.LogicID)
		}
		// 资格恒真：realness 绝不按 PR 把 fall 排出 room OR。任一 track（含 blind 续存）的摔都进房间 OR——
		//   ghost fall 可能是真人摔的镜像，宁报不漏。realness 的影响只在 N_r（→ C_FN 折扣，帮 fire）。
		eligible := true
		results = append(results, trackResult{d: d, pF: pF, lam: lam, eligible: eligible, f: f, cp: cp})
		devGeom := r.geomOf(adapter.DevKey(ts.LogicID))
		covers := make([]float64, len(devGeom))
		for i := range devGeom {
			covers[i] = devGeom[i].Covers
		}
		forensic = append(forensic, TrackForensic{LogicID: ts.LogicID, Present: ts.Present, RealProven: ts.RealProven, PReal: ts.PReal,
			PMirror: ts.PMirror, IsReflection: ts.IsReflection, PFallen: pF, Fire: d.Fire, Band: d.Band,
			X: ts.Obs.X, Y: ts.Obs.Y, Sep: ts.Sep, WallMargin: ts.WallMargin, Rho: ts.Rho, LaterBorn: ts.LaterBorn,
			DoorD: ts.DoorD, RepeatR: repeatR, SelfRecovered: selfRecovered, StillSec: ts.Obs.RadarTrack.StillSec,
			SMarg: append([]float64(nil), mS[:]...), Covers: covers})

		// drop（状态驱动）：消失 track 吸收到 {Left,Empty}（hand-off / 离房证据注入抬 SLeft → durable purge）
		//   → 离场确认 → drop（非 TTL，cancel 非 fire）。离房趋势已折进 SLeft 后验，不再单列 bool 门。
		if !ts.Present && mS[belief.SLeft]+mS[belief.SEmpty] >= absorbedThresh {
			dropIDs = append(dropIDs, ts.LogicID) // dropTrack(census) + 回传 tm evict：停 12s coast re-feed
		} else if fi.HardExited != nil && fi.HardExited(ts.LogicID, fi.NowMs) {
			// 逐帧离房事件(byte-14 event==2)硬 drop：门事件权威，绕过 SLeft 累积（治 ExitRoom 8.0 不灌进 SLeft
			//   的 pre-existing 慢 drop）。真摔固件自己 fire fall，故不读姿态/present。
			dropIDs = append(dropIDs, ts.LogicID)
			ef := r.enterSinceMs[ts.LogicID] // ExitFirst=连续 Enter 区驻留起点；退时不在 Enter 区(0)→回退 ExitRoom 时刻
			if ef == 0 {
				ef = fi.NowMs
			}
			exitFirsts = append(exitFirsts, ExitDeparture{LogicID: ts.LogicID, ExitFirstMs: ef, ExitRoomMs: fi.NowMs})
		}
	}
	for _, id := range dropIDs {
		r.dropTrack(id)
	}

	// F1 alone-streak 末更（同 realStreak 时序）：真人占用==1 续起点。
	//   blind 续存(S∉{E,L})仍计占用 → 摔倒不掉出 1 → 不清零（占用判据已含 blind，见循环内 realOccupancy）。
	//   重置抗抖动（柱②）：占用离开 1 须**持续 ≥arrivalConfirmFrames** 才清零——瞬态 churn（噪声 >AssocCm
	//   瞬拆新 logicID 起始 PReal=1，同 :20 hand-off churn）不重置；只令 alone 偏高 ≤K 帧（更易 fire）=严格 FN-safe。
	if realOccupancy == 1 {
		r.notSoloFrames = 0
		if r.aloneStreakStartMs == 0 {
			r.aloneStreakStartMs = fi.NowMs
		}
	} else {
		r.notSoloFrames++
		if r.notSoloFrames >= arrivalConfirmFrames {
			r.aloneStreakStartMs = 0
		}
	}

	// 步4 hand-off 信号（§64 噪声防线：连续 K 帧真人才算确认，抗噪声 churn 造假 lost/gain）。
	newStreak := make(map[string]int, len(curReal))
	for id := range curReal {
		newStreak[id] = r.realStreak[id] + 1 // 连续在场真人 → 累计；断了 → 不在 newStreak（归零）
	}
	var gained float64     // 本 tick 刚跨过确认阈（streak==K）的真人 = 确认新现（守恒重现落点）
	var gainedSleepad bool // 该 gain 来自 sleepad-only 房合成 track（lid=="sleepad-bed"）= 慢核接力（走+躺,~60s）
	for id, pr := range curReal {
		if newStreak[id] == arrivalConfirmFrames && pr > gained {
			// sleepad gain 不强制 vital（实证 InBed 不必然有 HR/RR：刚躺下没锁/RR firmware 不返，硬 vital-gate 会拦真接力）。
			//   去-ghost 判别 = InBed 转移 + streak K(3帧抗噪) + GainedReal≥0.5（HandoffLogOdds）+ 单住户 hard-gate。
			gained = pr
			gainedSleepad = id == "sleepad-bed" // 与 main.go onRoomFrame 合成 bed-track 的 lid 同源
		}
	}
	lost := false // 上 tick 已确认真人、本 tick 不在场真人 = 确认离场（hand-off 源）
	var lostAtMs int64
	var lostDeps []LostDeparture
	for id := range r.prevConfirmed {
		if _, ok := curReal[id]; !ok {
			lost = true
			anchor := r.lastSeenMs[id]
			if lf := r.lostAnchorMs[id]; lf > 0 && lf < anchor {
				anchor = lf // 冻结轨：锚 LostFirst(still-box 起点)而非被冻结 coast 拖后的 last-observed → hand-off Δ 回正区间(先离后现)
			}
			lostDeps = append(lostDeps, LostDeparture{LogicID: id, LostFirstMs: anchor})
			if anchor > lostAtMs {
				lostAtMs = anchor // 取最近离开者锚（先离后现的 loss 参考，Δ centering）
			}
		}
	}
	r.realStreak = newStreak
	r.prevConfirmed = map[string]bool{}
	for id, s := range newStreak {
		if s >= arrivalConfirmFrames {
			r.prevConfirmed[id] = true
		}
	}

	// abort（本帧新现 confirmed-real 活人 gained>0）由 Unit UD timer 的"任何 track 取消"统管（unit 级，
	//   含同房本人起身/他人到场/兄弟房现人）——Room 不再单独撤同房 blind faller（旧 abort-2 随 ramp 退役）。

	fr := r.aggregate(results, nr, fi.NowMs)
	fr.LostReal, fr.GainedReal, fr.GainedFromSleepad = lost, gained, gainedSleepad
	fr.LostAtMs = lostAtMs
	fr.LostExited = lostExited
	fr.PresentCount, fr.Tracks = presentCount, forensic
	fr.ExitFirsts = exitFirsts
	fr.LostDepartures = lostDeps
	fr.FuseSync = r.census.FuseForensic() // forensic：双雷达运动同步对状态
	// 可救援数（forensic，不门控 fire）：每 track 的 S 峰值是否 SBed（共用 belief.ArgmaxIsBed），
	//   躺床者不计 → census 折叠（A 范围 RescuableCount，同人对两端都不在床才减）。doc/cfn-rescuable-design.md。
	inBed := make(map[string]bool, len(forensic))
	for _, tf := range forensic {
		if belief.ArgmaxIsBed(tf.SMarg) {
			inBed[tf.LogicID] = true
		}
	}
	fr.Decision.RescuableCount = r.census.RescuableCount(inBed)
	// belief 就地复位（置于 aggregate 后：本帧 forensic 仍显发火态，下帧起从 prior 热机）。
	//   只复位 filter(S/B 信念)+decider(发火计时)；census/escalator(repeat 前科)保留。
	for _, id := range firedLogicIDs {
		r.filters[id] = belief.NewFilter(r.model, r.nb)
		r.deciders[id] = belief.NewDecider()
	}
	fr.FiredLogicIDs = firedLogicIDs
	fr.DroppedLogicIDs = dropIDs
	return fr
}

// aggregate 房间 OR 聚合（§60）：任一真人 track Fire → 房间 fire；代表 = 真人里(优先 firing)pF 最高者。
func (r *Room) aggregate(results []trackResult, nr int, nowMs int64) Frame {
	anyFire := false
	var rep *trackResult
	for i := range results {
		tr := &results[i]
		if !tr.eligible {
			continue
		}
		if tr.d.Fire {
			anyFire = true
		}
		if rep == nil {
			rep = tr
			continue
		}
		if tr.d.Fire != rep.d.Fire { // 优先 firing 的作代表（告警可见，含 blind 续存的消失 faller）
			if tr.d.Fire {
				rep = tr
			}
			continue
		}
		if tr.pF > rep.pF {
			rep = tr
		}
	}

	if rep == nil { // 无真人 track（空房/全 ghost）→ 空房决策
		r.lastMarg = belief.Vector{}
		return Frame{Decision: belief.Decision{PeopleCount: nr}}
	}
	dec := rep.d
	dec.Fire = anyFire // OR：任一真人摔即房间 fire
	dec.PeopleCount = nr
	r.lastMarg = rep.f.Space().MarginalS(rep.f.Alpha())
	return Frame{
		Probe:    belief.Snapshot(rep.f.Space(), rep.f, rep.cp, dec, rep.lam, nowMs),
		Decision: dec,
	}
}

// neutralIfNil 消失 track 无发射 → 中性零向量（log 域 0 → ComputeLambda=1 高度不可判；
// 但 pF≥0.55 走 report 档不读 lam，blind 续存告警照常持续）。
func neutralIfNil(v belief.JointVector, js *belief.JointSpace) belief.JointVector {
	if v == nil {
		return js.NewJointVector()
	}
	return v
}

// MarginalS 末帧房间代表 track 的 S 轴边缘后验。
func (r *Room) MarginalS() belief.Vector { return r.lastMarg }

// aloneMinAsOf 当前独居连续分钟（真人占用==1 连续时长，跨 tick streak；非独居=0）。
// 用上 tick 末的 streak 状态（占用 streak 在 Tick 末更）→ 喂本 tick rc，1 帧滞后。
func (r *Room) aloneMinAsOf(nowMs int64) float64 {
	if r.aloneStreakStartMs == 0 {
		return 0
	}
	return float64(nowMs-r.aloneStreakStartMs) / 60000.0
}

// dropTrack 移除一条 track 的全部 per-logicID 状态（filter/decider/floor + census）。状态驱动，非 TTL。
func (r *Room) dropTrack(id string) {
	delete(r.filters, id)
	delete(r.deciders, id)
	delete(r.floorGuards, id)
	delete(r.escalators, id)
	delete(r.lastSeenMs, id)
	delete(r.lostAnchorMs, id)
	delete(r.enterSinceMs, id)
	delete(r.fallLatched, id)
	r.census.Drop(id)
}

// stillOnsetMs 冻结首帧(LostFirst)时戳 = still-box run 起点 = now − 连续静止时长。ss≤0（未冻结）→ 0（lost 锚回退 last-observed）。
func stillOnsetMs(nowMs int64, stillSec float64) int64 {
	if stillSec <= 0 {
		return 0
	}
	return nowMs - int64(stillSec*1000)
}

// inAnyRect 点 (x,y) 是否落在任一矩形内（Enter 区点测；rect 顶点无序，取 min/max）。
func inAnyRect(x, y int, rects []adapter.Rect) bool {
	for _, r := range rects {
		if x >= min(r.X1, r.X2) && x <= max(r.X1, r.X2) && y >= min(r.Y1, r.Y2) && y <= max(r.Y1, r.Y2) {
			return true
		}
	}
	return false
}
