package engine

import (
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

// §84-A lost ramp：blind 的 fire = 时间驱动把 P^F(SFallen) ramp 到真阈 0.85（不绕阈、不 latch）。
//
//	lostFireThresh = lost 专用高阈（present 摔走 firmware/dbn_mode，与此分开）。
//	lostRampTargetLogOdds = logit(0.85)−logit(0.5) = ln(0.85/0.15) ≈ 1.7346（二义 0.5 起爬到 0.85 的 log-odds 跨度）。
//	delta/tick = target / (reset 窗 tick 数) × RiskTime 系数（dWindowMs 定 reset 窗，~1Hz）；从 at-loss 值起 ramp：
//	摔(高起点)早破 / 二义(0.5)reset 到点破 / 站着·离床(低起点·sleepad 压)reset 内到不了 → FP 被证据自然挡。form-anchor 留 oracle。
const (
	lostFireThresh        = 0.85
	lostRampTargetLogOdds = 1.7346
)

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
	LostReal   bool    // 本帧一条真人 track 消失（上帧在场真人、本帧不在）= hand-off 源候选
	GainedReal float64 // 本帧新现真人 track 的去 ghost 后验（0=无）= hand-off 宿候选（守恒重现）
	// forensic（全切片观测，不参与裁决）：
	PresentCount int             // 本帧在场 track 数（§61 共存源/消费门控判据）
	Tracks       []TrackForensic // 每 track 内部量（realness/ghost/消费门控/per-track 裁决）
}

// TrackForensic 单 track 的 DBN 内部量（forensic 暴露，X 光全切片用，不参与裁决）。
type TrackForensic struct {
	LogicID      int
	Present      bool
	PReal        float64 // 真人后验（realness 轴；ghost→低）
	PMirror      float64 // 镜像后验
	IsReflection bool    // 桶二镜面几何判定
	PFallen      float64 // per-track P^F
	Fire         bool    // per-track 裁决（持续≥T_hold）
	Band         string  // per-track 档（report/no/tie/indeterminate）
	X, Y         int     // forensic：末帧 canvas 坐标
	Sep          float64 // forensic：reflectSep cm（墙外反射裕度）
	WallMargin   float64 // forensic：mEv 墙外项
	Rho          float64 // forensic：CoexistRho 同步移动强度
	LaterBorn    bool    // forensic：成对后到
}

// Room 单房多 track 引擎：每 logicID 一份 belief 滤波 + 裁决器；census 管身份/realness/人数。
type Room struct {
	model         *belief.Model
	js            *belief.JointSpace // 共享读出/发射空间（各 track 滤波同维，idx 一致）
	cp            *belief.Coupling
	em            *belief.Emission
	census        *adapter.TrackCensus
	filters       map[int]*belief.Filter     // 每 logicID 一份 S/B 联合滤波（§A.3② 隐维复制）
	deciders      map[int]*belief.Decider    // 每 logicID 一份持续计时裁决
	floorGuards   map[int]*belief.FloorGuard // 每 logicID 一份 FN-safe 兜底（总时长 floor，契约其十五）
	nb            int
	p             adapter.Params
	lastMarg      belief.Vector // 末帧房间代表 track 的 S 边缘（MarginalS 读出）
	realStreak    map[int]int   // 每 logicID 连续在场真人帧数（步4 hand-off：confirmed=streak≥K，抗噪声 churn）
	prevConfirmed map[int]bool  // 上 tick 已确认真人(streak≥K)的 logicID 集（算 lost）
	// F1 独居连续计时（跨 tick，与 realStreak 同层）：真人占用==1 连续起点 ms（0=当前非独居）。
	//   占用判据 = PReal≥0.5 ∧ S∉{Empty,Left}（含 blind 续存的 faller，filter 后的 MarginalS），
	//   **非** census.Nr() present-only——否则独居者摔进 blind 时占用掉 0 误清零计时（lost-fall FN）。
	aloneStreakStartMs int64
	notSoloFrames      int // 连续占用≠1 帧数（重置抗抖动柱②：≥arrivalConfirmFrames 才清 alone-streak）
	// §82 D 窗（neighbor lost-fall 兜底耐心窗）：blind track 起算时戳（per logicID）+ 窗长（bootstrap 注入
	//   = thresholdNonRest+2min；0=未注入→D 关闭=旧行为）。只压"这条 blind 自己的 lost-fall fire"。
	blindSinceMs map[int]int64
	dWindowMs    int64 // §84-A：lost ramp 的 reset 窗（定 ramp 速率）+ abort-2 用；bootstrap 注入 thresholdNonRest+2min
}

// NewRoom 建单房引擎。geom = 床几何（adapter.BedGeoms 从 layout 派生）；nb = 床数；
// dWindowMs = §82 neighbor D 窗长（bootstrap 注入 thresholdNonRest+2min；0=D 关闭）。
func NewRoom(geom []belief.BedGeom, nb int, dWindowMs int64) *Room {
	return &Room{
		model:         belief.DefaultModel(),
		js:            belief.NewJointSpace(nb),
		cp:            belief.NewCoupling(geom),
		em:            belief.NewEmission(geom),
		census:        adapter.NewTrackCensus(adapter.DefaultTrackCensusParams()),
		filters:       map[int]*belief.Filter{},
		deciders:      map[int]*belief.Decider{},
		floorGuards:   map[int]*belief.FloorGuard{},
		nb:            nb,
		p:             adapter.DefaultParams(),
		realStreak:    map[int]int{},
		prevConfirmed: map[int]bool{},
		blindSinceMs:  map[int]int64{},
		dWindowMs:     dWindowMs,
	}
}

type trackResult struct {
	d        belief.Decision
	pF       float64
	lam      float64
	eligible bool // 参与房间 OR 的资格：有共存源时须真人(PReal≥0.5 排 ghost)；无共存源=孤轨→永发(§61)
	f        *belief.Filter
}

// Tick 一帧推进。rhoXroom（neighbor，单房=0）。每条 track 各跑一份滤波，房间 OR 聚合真人 fall。
func (r *Room) Tick(fi adapter.FrameInput, rhoXroom float64) Frame {
	r.census.Update(fi.NowMs, fi.Tracks, fi.RadarPos, fi.Walls, fi.Entrances)
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

	curReal := map[int]float64{} // 本 tick 在场真人(PReal≥0.5) logicID→PReal（算 lost/gained）
	var dropIDs []int
	realOccupancy := 0 // F1 本 tick 真人占用数（PReal≥0.5 ∧ S∉{E,L}）→ 末更 alone-streak
	for _, ts := range tracks {
		if ts.Present && ts.PReal >= 0.5 {
			curReal[ts.LogicID] = ts.PReal
		}
		f := r.filters[ts.LogicID]
		dec := r.deciders[ts.LogicID]
		fg := r.floorGuards[ts.LogicID]
		if f == nil {
			f = belief.NewFilter(r.model, r.nb) // 出生：新 logicID 起一份滤波（隐维复制）
			dec = belief.NewDecider()
			fg = belief.NewFloorGuard()
			r.filters[ts.LogicID], r.deciders[ts.LogicID], r.floorGuards[ts.LogicID] = f, dec, fg
		}

		var logPsi, logPhi belief.JointVector
		var obs belief.Observation
		if ts.Present {
			obs = adapter.BuildObservation(ts.Obs.RadarTrack, fi.Sleepads, fi.Beds, r.p)
			logPsi = r.cp.LogPsi(r.js, adapter.Gxy(ts.Obs.RadarTrack, fi.Beds, r.p))
			logPhi = r.em.LogPhi(r.js, obs)
		} else {
			// §84-A 消失态：雷达轴中性（RadarOnline=false），**接触轴(sleepad)仍应用**——在床 InBed→SBed
			//   保护睡眠者不被 ramp 误推；LeftBed→B vac→Ψ 放行 SFallen。再叠 lost ramp（无离房趋势时）把
			//   SFallen 往 0.85 推（从 at-loss 值起：站着低起点 / 在床 SBed 压 → reset 内到不了 0.85 → 不报）。
			obs = adapter.BuildObservation(adapter.RadarTrack{Online: false}, fi.Sleepads, fi.Beds, r.p)
			logPhi = r.em.LogPhi(r.js, obs)
			// lost ramp **封顶**：只在 blind-elapsed < reset 窗(dWindowMs)内累加 δ → 总量≈target(0.5→0.85 跨度)，
			//   reset 到点即停（不无界 overshoot）。fire ⟺ at-loss P^F≥0.5：二义(0.5)reset 到点到 0.85；
			//   低起点(站立/sleepad 压)reset 内到不了；过 reset 停 ramp（Predict 缓降，不再推高）。
			started := r.blindSinceMs[ts.LogicID]
			if started == 0 {
				started = fi.NowMs
			}
			if !ts.Obs.ExitTrend && fi.NowMs-started < r.dWindowMs {
				ramp := belief.LostRampPhi(r.js, r.lostRampDelta(rc))
				for i := range logPhi {
					logPhi[i] += ramp[i]
				}
			}
		}
		// 消失态：雷达中性 + 接触轴 + lost ramp（非离房趋势）；present 全发射。Predict 自持，无 TTL。
		f.Step(fi.NowMs, online, logPsi, logPhi, rhoXroom)

		mS := f.Space().MarginalS(f.Alpha())
		// F1 真人占用：PReal≥0.5（排 ghost）∧ S∉{Empty,Left}（含 blind 续存的 faller，未 absorbed 离场）。
		//   摔进 blind 时 S=Fallen∉{E,L} → 仍计占用 → 独居计时不被摔倒清零（present-only 的 Nr() 会误清）。
		if ts.PReal >= 0.5 && mS[belief.SEmpty]+mS[belief.SLeft] < absorbedThresh {
			realOccupancy++
		}

		pF := f.Space().PFallen(f.Alpha())
		lam := belief.ComputeLambda(f.Space(), neutralIfNil(logPsi, r.js), neutralIfNil(logPhi, r.js))
		d := dec.Step(fi.NowMs, pF, lam, rc)
		// FN-safe 兜底 floor（契约其十五）：present 时总时长 ≥ T_floor(按 area) 且无正向休息证据 → 强制 fire。
		//   接住 emission 被 area误学/z假阳/接触假阳误压的真摔。消失态不走 floor（走 blind 续存）。
		//   zUp/contactInBed 抽取已收进 FloorGuard.Step(obs)——engine 只 OR verdict 不碰 obs。
		if ts.Present && fg.Step(obs) {
			d.Fire = true
			if d.Band == "" || d.Band == "no" || d.Band == "indeterminate" {
				d.Band = "floor"
			}
		}
		// §84-A lost fire：blind 的 fire = ramp 把 P^F(SFallen) 真推到 0.85（不绕阈、不 latch；柱A）。
		//   present 摔走 firmware/dbn_mode（与此分开，不进此分支=结构免疫）。三 cancel 经"压低 P^F 使到不了 0.85"实现：
		//   离房趋势→不 ramp+下方 drop；handoff→GateBlindRow→SLeft→P^F 掉；在床→sleepad InBed→SBed 压→到不了；
		//   恢复→转 Present 出本分支。dWindowMs≤0→无 ramp→P^F 不升→不报（旧行为/零回归）。blindSinceMs 留 abort-2。
		if ts.Present {
			delete(r.blindSinceMs, ts.LogicID)
		} else {
			if r.blindSinceMs[ts.LogicID] == 0 {
				r.blindSinceMs[ts.LogicID] = fi.NowMs
			}
			if pF >= lostFireThresh && !ts.Obs.ExitTrend {
				d.Fire, d.Band = true, "lost"
			} else {
				d.Fire = false
			}
		}
		// 资格恒真：realness 绝不按 PR 把 fall 排出 room OR。任一 track（含 blind 续存）的摔都进房间 OR——
		//   ghost fall 可能是真人摔的镜像，宁报不漏。realness 的影响只在 N_r（→ C_FN 折扣，帮 fire）。
		eligible := true
		results = append(results, trackResult{d: d, pF: pF, lam: lam, eligible: eligible, f: f})
		forensic = append(forensic, TrackForensic{LogicID: ts.LogicID, Present: ts.Present, PReal: ts.PReal,
			PMirror: ts.PMirror, IsReflection: ts.IsReflection, PFallen: pF, Fire: d.Fire, Band: d.Band,
			X: ts.Obs.X, Y: ts.Obs.Y, Sep: ts.Sep, WallMargin: ts.WallMargin, Rho: ts.Rho, LaterBorn: ts.LaterBorn})

		// drop（状态驱动）：消失 track 吸收到 {Left,Empty}（handoff 经 GateBlindRow 整流）或离房趋势
		//   （§84 步3：走向出口）→ 离场确认 → drop（非 TTL，cancel 非 fire）。
		if !ts.Present && (mS[belief.SLeft]+mS[belief.SEmpty] >= absorbedThresh || ts.Obs.ExitTrend) {
			dropIDs = append(dropIDs, ts.LogicID)
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
	newStreak := make(map[int]int, len(curReal))
	for id := range curReal {
		newStreak[id] = r.realStreak[id] + 1 // 连续在场真人 → 累计；断了 → 不在 newStreak（归零）
	}
	var gained float64 // 本 tick 刚跨过确认阈（streak==K）的真人 = 确认新现（守恒重现落点）
	for id, pr := range curReal {
		if newStreak[id] == arrivalConfirmFrames && pr > gained {
			gained = pr
		}
	}
	lost := false // 上 tick 已确认真人、本 tick 不在场真人 = 确认离场（hand-off 源）
	for id := range r.prevConfirmed {
		if _, ok := curReal[id]; !ok {
			lost = true
		}
	}
	r.realStreak = newStreak
	r.prevConfirmed = map[int]bool{}
	for id, s := range newStreak {
		if s >= arrivalConfirmFrames {
			r.prevConfirmed[id] = true
		}
	}

	// abort-2（§82,乙）：本帧新现 confirmed-real 活人(gained>0)→放弃同房 **D 窗内** blind faller 待报。
	//   依据：新活人 = 本人起身(无需报) 或 他人到场(自然能发现 fall，无需报) = 风险分层降级(署名接受 FN 残余)。
	//   只撤 D 窗内未释放的(past-D 已 fire 的不撤)；gained 必来自 present track，不会误伤 faller 自身。
	if gained > 0 {
		for i := range results {
			f := forensic[i]
			if !f.Present && r.blindSinceMs[f.LogicID] != 0 &&
				r.dWindowMs > 0 && fi.NowMs-r.blindSinceMs[f.LogicID] < r.dWindowMs {
				results[i].d.Fire = false
				r.dropTrack(f.LogicID)
			}
		}
	}

	fr := r.aggregate(results, nr, fi.NowMs)
	fr.LostReal, fr.GainedReal = lost, gained
	fr.PresentCount, fr.Tracks = presentCount, forensic
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
		Probe:    belief.Snapshot(rep.f.Space(), rep.f, r.cp, dec, rep.lam, nowMs),
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

// lostRampDelta §84-A：每 blind tick 给 SFallen 加的 log-odds 增量 = 目标跨度 / reset 窗 tick 数 × RiskTime 系数。
//
//	dWindowMs≤0 → 0（D 关闭=不 ramp，零回归）；~1Hz 假设 tick 数 ≈ dWindowMs/1000。数值留 oracle。
func (r *Room) lostRampDelta(rc belief.RiskContext) float64 {
	if r.dWindowMs <= 0 {
		return 0
	}
	nTicks := float64(r.dWindowMs) / 1000.0
	if nTicks < 1 {
		nTicks = 1
	}
	return lostRampTargetLogOdds / nTicks * riskRampMul(rc)
}

// riskRampMul §84-A RiskTime：高风险时段证据权重放大 → ramp 更陡更快到 0.85（夜间/独居/失能；form-anchor 留 oracle）。
func riskRampMul(rc belief.RiskContext) float64 {
	m := 1.0
	if rc.Night {
		m *= 1.2
	}
	if rc.AloneContinuousMin > 0 {
		m *= 1.2
	}
	if rc.Disabled {
		m *= 1.2
	}
	return m
}

// aloneMinAsOf 当前独居连续分钟（真人占用==1 连续时长，跨 tick streak；非独居=0）。
// 用上 tick 末的 streak 状态（占用 streak 在 Tick 末更）→ 喂本 tick rc，1 帧滞后。
func (r *Room) aloneMinAsOf(nowMs int64) float64 {
	if r.aloneStreakStartMs == 0 {
		return 0
	}
	return float64(nowMs-r.aloneStreakStartMs) / 60000.0
}

// dropTrack 移除一条 track 的全部 per-logicID 状态（filter/decider/floor/D 计时 + census）。状态驱动，非 TTL。
func (r *Room) dropTrack(id int) {
	delete(r.filters, id)
	delete(r.deciders, id)
	delete(r.floorGuards, id)
	delete(r.blindSinceMs, id)
	r.census.Drop(id)
}
