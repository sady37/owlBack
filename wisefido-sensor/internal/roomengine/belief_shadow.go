package roomengine

import (
	"owl-common/alarm"
	"owl-common/observation"
	"wisefido-sensor/internal/roomengine/belief"

	"go.uber.org/zap"
)

// belief_shadow.go — §9 第 3b 步：belief 信念引擎旁路 shadow（**只 log 不 fire**）。
// 在 publishTrackStatuses（per-room 串行，无锁）每帧喂 belief 观测、跑 forward、Decider 确认，
// 把 belief 的 Fall 判定 log 出来与 gate-list alarm 离线对账。**绝不触发任何 alarm**。
// 灰度开关 beliefShadowEnabled 可秒关；整段 recover 保护，任何 panic 不影响引擎。
//
// lost-fall = 走动中突然消失（消失前 still-box < MovingPreconditionMs(60s)），ramp + Decider 确认窗。
// production engine 已做 ghost 检测（Verdict），shadow 沿用 Verdict 过滤；event(ExitRoom/InBed) 喂入
// 经 beliefShadowEvent（v1 仅 track-frame + 门区 geom 抑制；跨房 exit 事件喂入留 v1.1）。

const beliefShadowEnabled = true

const beliefShadowLostTTLMs = 60_000 // track 超此无帧 = 丢失（对齐 trackLostAnchorMs）

// P6.1b-D(审查㉛ Opt-1)小卫生间 provisional/分级窗:
//   设备富(unit 有其它设备)→ 30min cancel 窗(覆盖立项 np=0 +335s),窗到未佐证升级。
//   设备贫(浴室独苗,无跨设备 cancel 可能)→ 短窗早决断压制(省真摔无谓延迟,resource-scaled v3)。
const (
	beliefProvisionalRichWindowMs = 30 * 60 * 1000 // 设备富:30min cancel 窗
	beliefProvisionalPoorWindowMs = 120 * 1000     // 设备贫:2min 早决断(无 cancel 可能)
)

type beliefShadowTrack struct {
	lastSeenMs       int64
	stillBoxAgeMs    int64 // 最后一帧时的 still-box 时长（消失前是否走动的依据）
	geom             belief.Geom
	lastX, lastY     int     // 最后一帧位置 → 算丢失点离门距离 d（P2 reachable-exit）
	approachSpeedCmS float64 // A：丢失前朝门定向逼近速度（在 track 仍活时算好 stash，丢失后 ts 可能已销毁）
	lostAnchor       int64
	// P6.1b-D provisional 状态机(小卫生间分支,跨 tick):
	provisionalSince    int64 // 进 provisional 的 ms(0=未进);小卫生间 lost 即置 + log provisional-now
	provisionalResolved bool  // cancel/escalate/suppressed 已决断 + log(防重复 log)
}

// beliefShadowTLayer DBN P1 Track 层（per-track T_t）。与 Room 层 tracks 刻意分离：
// Room 层对 ghost 用 delete+method-2 gate；Track 层**保留 ghost track**让 A_T 结构化处理
// （Ghost→None，不通 Lost），正是 P1 要对照的"gate vs A_T 一致性"。
type beliefShadowTLayer struct {
	tb           *belief.TrackBelief
	lastSeen     int64
	geom         belief.Geom
	device       string // 源雷达 device_addr（同房对等雷达占用对账排除自身用）
	loggedLo     bool   // 已 log 过本次 Lost 峰（防重复）
	lastX, lastY int    // P3.1:上帧位置,算本帧空间跳跃 Δ/dt(独立 shadow realness 探测器①)
	lastPosTs    int64  // 上帧位置时刻
	lastPose     int    // P3.2:上帧 pose/z,算 pose/z 锁死帧数(冻结伪迹 B 佐证③)
	lastZ        int
	poseZLock    int     // pose&z 连续恒定帧数
	realLO       float64 // P3.3:realness log-odds(带遗忘 γ;>0 real <0 ghost),摔前走路 realness 带进倒地窗
}

type beliefShadow struct {
	b           *belief.Belief
	decider     belief.Decider
	tracks      map[int]*beliefShadowTrack
	fired       bool                         // 已 log 过本次 confirm（防 confirm 持续期重复 log）
	tlayer       map[int]*beliefShadowTLayer // DBN P1 Track 层
	deviceSpeed  map[string]*deviceSpeedStat // P2.1：per-device 学习走速封顶（device→room 稳定，跨 track 累积）
	lastLostGeom belief.Geom                 // #3：最近一次丢失点 geom（fall log 辨床/桶区误报用）
}

// tLostLogThresh Track 层 P(TLost) 超此即 log（对照 gate-list lost_track 报警）。
const tLostLogThresh = 0.5

func (e *Engine) beliefShadowFor(roomID string) *beliefShadow {
	sh := e.beliefShadows[roomID]
	if sh == nil {
		sh = &beliefShadow{
			b:           belief.New(belief.DefaultModel()),
			tracks:      map[int]*beliefShadowTrack{},
			tlayer:      map[int]*beliefShadowTLayer{},
			deviceSpeed: map[string]*deviceSpeedStat{},
		}
		e.beliefShadows[roomID] = sh
	}
	return sh
}

// beliefShadowEvent 把离散事件喂进房间 shadow（EnterRoom/ExitRoom）——cancellation 主源。
// handleMessage 处理这些事件时调；log-only，不影响任何路径。
func (e *Engine) beliefShadowEvent(roomID, eventName string, nowMs int64) {
	if !beliefShadowEnabled {
		return
	}
	defer func() { _ = recover() }()
	o, ok := radarEventToObs(eventName, nowMs, belief.GeomInEnter)
	if !ok {
		return
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)
	sh.b.Step(nowMs, []belief.Observation{o})

	// DBN P1 Track 层：firmware ExitRoom → 各 track 喂 TObsExit（→JustLeft，压 Lost）。
	if eventName == alarm.ExitRoom {
		for _, tl := range sh.tlayer {
			tl.tb.Step(nowMs, []belief.TObservation{{Kind: belief.TObsExit, Conf: 0.9, Ts: nowMs, Fresh: true}})
		}
	}
}

// beliefShadowTick 每帧 per-room shadow：喂 track 观测 + 消失触发 lost-fall + Decider 确认 → log。
// **log-only，绝不 fire alarm**。recover 保护 + 灰度开关。
func (e *Engine) beliefShadowTick(roomID string, bases []TrackStatusBase, nowMs int64) {
	if !beliefShadowEnabled || e.logger == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			e.logger.Warn("belief_shadow_panic", zap.String("room_id", roomID), zap.Any("recover", r))
		}
	}()

	e.mu.RLock()
	tm := e.rooms[roomID]
	grid := e.grids[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)

	var obs []belief.Observation
	cur := make(map[int]struct{}, len(bases))
	curTL := make(map[int]struct{}, len(bases))
	for i := range bases {
		b := &bases[i]

		// DBN P1 Track 层：每帧（含 ghost）喂 present 发射。
		// P3.1(委员会审查⑦裁定):ghostness 由**独立 shadow realness**(XY 三探测器)算,
		// **不复用生产 GhostPenalty/Verdict**(那套漏 cd2b 冻结 ghost)→ shadow R_i 才能抓生产漏的 ghost。
		// 与下方 Room 层 ghost-delete 刻意分离：Track 层让 A_T 把 ghost 路由 →None。
		curTL[b.TrackID] = struct{}{}
		tl := sh.tlayer[b.TrackID]
		if tl == nil {
			tl = &beliefShadowTLayer{tb: belief.NewTrackBelief()}
			sh.tlayer[b.TrackID] = tl
		}
		// 本帧空间跳跃 Δ/dt(探测器①);缺前帧 → -1 跳过①,靠②③(读 ts raw 动学量)。
		frameJumpCmS := -1.0
		if tl.lastPosTs > 0 && nowMs > tl.lastPosTs {
			frameJumpCmS = float64(distInt(b.X, b.Y, tl.lastX, tl.lastY)) * 1000 / float64(nowMs-tl.lastPosTs)
		}
		// P3.2 pose/z 锁死帧数(冻结伪迹 B 佐证③):pose&z 连续恒定累加,变则清零。
		if tl.lastPosTs > 0 && tl.lastPose == b.Pose && tl.lastZ == b.Z {
			tl.poseZLock++
		} else {
			tl.poseZLock = 0
		}
		tsRaw := tm.tracks[b.TrackID]
		// 独立 shadow realness 二值检测:P3.1 跳变/急变 / P3.2 冻结伪迹复合门控(补静止反射;真人久站缺 A→不判)。
		jumpGhost := shadowTrackGhostness(tsRaw, frameJumpCmS) == 1
		frozenGhost := !jumpGhost && shadowFrozenArtifact(tsRaw, tl.poseZLock, grid, b.X, b.Y, b.CellAreaType, nowMs)
		// P3.3 记忆 L_R:二值检测 + 走动 real 证据 → 连续 P(ghost) 带遗忘 γ(摔前 realness 带进倒地窗)。
		dtSec := 0.0
		if tl.lastPosTs > 0 && nowMs > tl.lastPosTs {
			dtSec = float64(nowMs-tl.lastPosTs) / 1000
		}
		var tlGhostness float64
		tl.realLO, tlGhostness = realnessStep(tl.realLO, dtSec, b.MoveActive, jumpGhost, frozenGhost)
		tl.lastX, tl.lastY, tl.lastPosTs = b.X, b.Y, nowMs
		tl.lastPose, tl.lastZ = b.Pose, b.Z
		tlGeom := geomFromArea(b.CellAreaType)
		tl.tb.Step(nowMs, []belief.TObservation{{
			Kind: belief.TObsPresent, Ghostness: tlGhostness, Geom: tlGeom,
			Conf: 0.9, Ts: nowMs, Fresh: true,
		}})
		tl.lastSeen = nowMs
		tl.geom = tlGeom
		tl.device = b.DeviceAddr
		tl.loggedLo = false // 重新检出 → 允许后续再次 Lost 时重新 log

		if b.Verdict == VerdictGhost {
			// 沿用 production ghost 检测；且把已注册 track 移出追踪：
			// real→ghost→消失 的镜面反射不得触发 lost-while-moving（ghost 闪灭 ≠ 倒地）。
			// 与 gate 侧 method-2 + belief replay guard 同构。
			delete(sh.tracks, b.TrackID)
			continue
		}
		cur[b.TrackID] = struct{}{}
		x, y, z := b.X, b.Y, b.Z
		tr := observation.Track{BedStatus: observation.BedStatusUnchanged, Pose: b.Pose, PositionX: &x, PositionY: &y, PositionZ: &z}
		ts := tm.tracks[b.TrackID]
		if ts != nil {
			obs = append(obs, radarFrameAdapter(tr, ts, grid, nowMs)...)
		}
		st := sh.tracks[b.TrackID]
		if st == nil {
			st = &beliefShadowTrack{}
			sh.tracks[b.TrackID] = st
		} else if isSelfRescueRecapture(st.lostAnchor, st.lastSeenMs, nowMs) {
			// P3.4 recapture 软恢复:曾丢失(lost-fall ramping)的 track 返回。production 硬 cancel pending
			// lost-fall(R0 不动);shadow 不硬 cancel,记 self-rescue candidate —— 跌后自救可能,真发应留**低
			// severity**(只 log 不 fire,R1)。否则把"摔了又自己爬回"当没事抹掉。
			e.logger.Info("belief_shadow_recapture", // 仅 log,无 alarm
				zap.String("room_id", roomID),
				zap.Int("track_id", b.TrackID),
				zap.Int64("ts_ms", nowMs),
				zap.Int64("p3_4_recapture_ms", nowMs-st.lastSeenMs),   // 丢失多久后返回(cd2b≈5.85min)
				zap.Bool("p3_4_would_cancel", true),                   // production 会硬 cancel pending lost-fall
				zap.Bool("p3_4_self_rescue_candidate", true),          // → shadow 标低 severity 非抹掉
				zap.String("last_geom", st.geom.String()),
			)
		}
		st.lastSeenMs = nowMs
		st.lostAnchor = 0
		st.geom = geomFromArea(b.CellAreaType)
		st.lastX, st.lastY = b.X, b.Y
		// P2.1：track 仍活时学本设备走速 → 个性化封顶。A：算好朝门定向逼近（丢失后 ts 可能已销毁）。
		ds := sh.deviceSpeed[b.DeviceAddr]
		if ds == nil {
			ds = &deviceSpeedStat{}
			sh.deviceSpeed[b.DeviceAddr] = ds
		}
		before := ds.samples
		ds.observe(sampleWalkSpeed(ts, nowMs))
		if before < beliefSpeedMinSamples && ds.samples >= beliefSpeedMinSamples {
			e.logger.Info("belief_shadow_speed_learned", // P2.1：本设备走速学满 → 个性化封顶生效
				zap.String("room_id", roomID),
				zap.String("device_addr", b.DeviceAddr),
				zap.Float64("walk_ewma_cms", ds.ewmaCmS),
				zap.Float64("reach_cap_cms", ds.capCmS()),
			)
		}
		st.approachSpeedCmS = approachSpeedTowardExit(ts, grid, beliefReachWindowMs, nowMs, ds.capCmS())
		if ts != nil && ts.StillBoxRunStart > 0 {
			st.stillBoxAgeMs = nowMs - ts.StillBoxRunStart
		} else {
			st.stillBoxAgeMs = 0
		}
	}

	// P6.1b-D(审查㉛ Opt-1)小卫生间 provisional 分级 的 room 级前置(每 tick 一次)。
	smallBath := e.IsSmallBathroom(roomID)
	suiteID := ""
	if smallBath {
		suiteID = e.SuiteIDForRoom(roomID)
	}

	// 丢失扫掠：track 超 TTL 无帧 + 消失前走动（still-box<60s）+ 非门区 → lost-fall ramp。
	for tid, st := range sh.tracks {
		if _, alive := cur[tid]; alive || nowMs-st.lastSeenMs <= beliefShadowLostTTLMs {
			continue
		}
		if st.stillBoxAgeMs >= int64(FallRulesParam.Lost.MovingPreconditionMs) {
			continue // 静止态消失 → Still-fall 域
		}
		if st.lostAnchor == 0 {
			st.lostAnchor = st.lastSeenMs
		}

		// P6.1b-D(审查㉛ Opt-1)小卫生间 lost → provisional/分级状态机(替标准 reachableExit 抑制)。
		// 门距退化(处处近门)→ 不靠 door-distance;Fallen 经 NoDetect 真 ramp(dx=0),离场判别交 cancel 窗。
		// cancel 须**既 attribution-safe 又 leave-discriminating**(扩展不变量 审查㉝):正向离场证据,非缺证。
		// **cancel = recapture ONLY**(走失者本人 anchor 正向重现,per-identity,过两条)。
		// **np=0 永不 cancel**(审查㉝:realness-empty 看不到已丢失摔倒者→≈np=0;np=0 是 lost-fall 定义性条件,
		// 摔/离共有,非判别器)→ np=0 仅 aux LOG。可靠离场-cancel 升级 = Opt-3 边界穿越(非 np=0)。
		// 默认升级硬约束:歧义→escalate。设备富 30min cancel 窗;设备贫(独苗)短窗早决断→压制+LOG(v3 resource-scaled)。
		if smallBath && (st.geom == belief.GeomInToilet || st.geom == belief.GeomInEnter) {
			tl := sh.tlayer[tid]
			realnessP := 1.0
			lostDevice := ""
			if tl != nil {
				realnessP = realnessPFromLO(tl.realLO)
				lostDevice = tl.device
			}
			if st.provisionalSince == 0 {
				st.provisionalSince = nowMs
				e.logger.Info("belief_shadow_lostfall_provisional", // provisional-now 低 sev(真摔即时有声,不静默 5.5min)
					zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
					zap.String("last_geom", st.geom.String()))
			}
			// cancel 佐证 = recapture ONLY(审查㉝:正向重现,过 attribution-safe + leave-discriminating 两条)
			recaptured := false
			if e.suiteCensus != nil {
				if rc, recap := e.suiteCensus.SoleResidentRecaptureState(suiteID); rc == 1 && recap {
					recaptured = true
				}
			}
			np0Recent := tm.lastNumberPeopleZeroMs > 0 && nowMs-tm.lastNumberPeopleZeroMs <= beliefProvisionalRichWindowMs // 仅 aux LOG,不 gate cancel
			if recaptured {
				if !st.provisionalResolved {
					e.logger.Info("belief_shadow_lostfall_cancel", // recapture 正向重现 → 软降(P3.4),不 ramp Fallen
						zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
						zap.Bool("p6_1b_recapture", true),
						zap.Bool("p6_1b_np0_aux", np0Recent), // np=0 仅 aux observability(㉝:absence≠leave,不 cancel)
						zap.Int64("p6_1b_provisional_ms", nowMs-st.provisionalSince))
					st.provisionalResolved = true
				}
				continue // 离场确认(recapture)→ 不喂 lost-fall 发射
			}
			rich := e.SuiteHasOtherDevice(suiteID, lostDevice) // 设备密度=机构资源代理(v3)
			elapsed := nowMs - st.provisionalSince
			if rich {
				obs = append(obs, noDetectObs(st.geom, realnessP, 0, nowMs)) // dx=0 干净 ramp(门距退化,disambiguation 交 cancel 窗)
				if elapsed >= beliefProvisionalRichWindowMs && !st.provisionalResolved {
					e.logger.Info("belief_shadow_lostfall_escalate", // 窗到未佐证 → 全 sev 真摔(延迟但不漏)
						zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
						zap.Int64("p6_1b_provisional_ms", elapsed))
					st.provisionalResolved = true
				}
			} else { // 设备贫(浴室独苗):无跨设备 cancel 可能 → 短窗早决断
				if elapsed < beliefProvisionalPoorWindowMs {
					obs = append(obs, noDetectObs(st.geom, realnessP, 0, nowMs)) // 短窗内仍 ramp(provisional)
				} else if !st.provisionalResolved {
					e.logger.Info("belief_shadow_lostfall_suppressed", // 设备贫→压制不 page,但 LOG 疑似摔(no-silent-caps,机构可回看)
						zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
						zap.Int64("p6_1b_provisional_ms", elapsed), zap.Bool("p6_1b_resource_poor", true))
					st.provisionalResolved = true
				}
			}
			continue // D-path 已处理,跳过标准 P6.5①/P6.1a 发射
		}

		// P6.5①(审查㉑批准,选项A)跨设备 track 守恒:bathroom track 丢失时,若 sole resident 已重现别处
		// (SleepadAnchored 回床铁证 ∨ 跨 BathroomGate 返回 bedroom)= 人移到别处 = exit 非 fall →
		// 不喂 lost-fall 发射(shadow 抑制,只 log 不 fire,R1;自洽 P3.4 超窗重现=自救)。
		// per-identity 绑 sole resident(census 只读 accessor 锁内读,visitor 不计);**多 resident → gate OFF**
		// (census 决定19 跳过 anchor-flip → 无人带 Bathroom anchor → 无法 per-identity)= 漏报-safe 保留告警 + skip LOG(数据闸)。
		if st.geom == belief.GeomInToilet && e.suiteCensus != nil {
			residentCount, recaptured := e.suiteCensus.SoleResidentRecaptureState(e.SuiteIDForRoom(roomID))
			switch {
			case residentCount == 1 && recaptured:
				e.logger.Info("belief_shadow_exit_recapture", // 仅 log,无 alarm(R1)
					zap.String("room_id", roomID),
					zap.Int("track_id", tid),
					zap.Int64("ts_ms", nowMs),
					zap.Int64("p6_5_lost_anchor_ms", st.lostAnchor),
					zap.Int64("p6_5_gap_ms", nowMs-st.lastSeenMs),
					zap.Bool("p6_5_exit_confirmed", true),          // track 守恒:sole resident 重现别处 = exit 非 fall
					zap.Bool("p6_5_would_suppress_lostfall", true), // shadow 会抑制此 lost-fall(production 不动,R0)
					zap.String("last_geom", st.geom.String()),
				)
				continue // exit 确认 → 不喂 lost-fall 发射(shadow 抑制)
			case residentCount > 1:
				e.logger.Info("belief_shadow_recapture_skip_multiresident", // 数据闸(no silent caps):量多resident浴室-lost FP频率
					zap.String("room_id", roomID),
					zap.Int("track_id", tid),
					zap.Int64("ts_ms", nowMs),
					zap.Int("p6_5_resident_count", residentCount),
					zap.Bool("p6_5_recapture_skipped", true), // 多resident anchor-flip被census决定19跳过→gate OFF→保留告警(零跨身份漏报)
				)
				// fall through → 正常 lost-fall 发射(保留告警)
			}
		}
		// P6.1a(阻塞项#1):no-detect 抬 Fallen 须门控(R_i + door-exit),不裸 absence 抬 fall(治 dropout-FP)。
		// R_i=P(real)=σ(realLO)(丢失时用消失前最后已算 realness:消失时是不是真人);缺 track 层 → 默认 1.0
		// (保守不漏报)。door-exit=reachableExitScore(近门+定向可达走出 → 压抬;门区丢轨不抬,治 D5F7/D523)。
		realnessP := 1.0
		if tl := sh.tlayer[tid]; tl != nil {
			realnessP = realnessPFromLO(tl.realLO) // σ(LO)=P(real)
		}
		doorExitP := 0.0
		exitDist := -1
		if grid != nil {
			exitDist = grid.NearestEntryDist(st.lastX, st.lastY)
			doorExitP = reachableExitScore(exitDist, st.approachSpeedCmS)
		}
		// P2：门区不再硬 continue；改软门——no-detect(R_i+door-exit 门控抬 Fallen) 与 reachable-exit(压 Fallen) 同 tick 对冲。
		obs = append(obs, noDetectObs(st.geom, realnessP, doorExitP, nowMs))
		if realnessP < 0.5 || doorExitP > 0.5 { // P6.1a 门控生效(ghost消失/门区可达走出)→ 弱抬/不抬 Fallen
			e.logger.Info("belief_shadow_nodetect_gated", // observability:量 no-detect 被 R_i/door-exit 压的频率
				zap.String("room_id", roomID),
				zap.Int("track_id", tid),
				zap.Int64("ts_ms", nowMs),
				zap.Float64("p6_1a_Ri", realnessP),       // P(real);低=ghost 消失,不抬
				zap.Float64("p6_1a_door_exit", doorExitP), // P(door-exit);高=门区可达走出,不抬
				zap.Bool("p6_1a_nodetect_gated", true),
				zap.String("last_geom", st.geom.String()),
			)
		}
		sh.lastLostGeom = st.geom // #3：记最近丢失点 geom，供 fall log 辨别床/桶区
		if grid != nil {
			obs = append(obs, reachableExitObs(exitDist, st.approachSpeedCmS, st.geom, nowMs))
		}
	}

	// DBN P1 Track 层 absent 扫掠 + log：超 TTL 无帧喂 absent（不 skip ghost——A_T 自处理）。
	// P(TLost) 越阈即 log，与 gate-list lost_track 报警离线对照。**只 log 不 fire**。
	for tid, tl := range sh.tlayer {
		if _, alive := curTL[tid]; alive {
			continue
		}
		if nowMs-tl.lastSeen <= beliefShadowLostTTLMs {
			continue
		}
		tobs := []belief.TObservation{{
			Kind: belief.TObsAbsent, Geom: tl.geom, Conf: 0.9, Ts: nowMs, Fresh: true,
		}}
		// 同房多雷达占用对账（仅单床房）：本台丢失，但同房别台此刻仍见真人 → 本 track 更可能是
		// 别台真人的重影/重复 → 喂 TObsPeerLive 压 TLost（与生产 gate-list 守卫同构；单床闸在此把关）。
		if tm.bedCount == 1 && tm.otherDeviceRealTrackRecent(tl.device, nowMs) {
			tobs = append(tobs, belief.TObservation{Kind: belief.TObsPeerLive, Conf: 0.9, Ts: nowMs, Fresh: true})
		}
		// C3 共享算子：reachable-exit 同源喂 Track 层（近门 + 定向逼近 → 偏 JustLeft 压 Lost）。
		// ghost track 在 Room 层被 delete（sh.tracks 无），A_T 自走 Ghost→None，无需此观测。
		exitDist, approachV, reachE := -1, 0.0, 0.0
		reachCap := float64(beliefReachSpeedCapCmS)
		if ds := sh.deviceSpeed[tl.device]; ds != nil {
			reachCap = ds.capCmS()
		}
		if st := sh.tracks[tid]; st != nil && grid != nil {
			exitDist = grid.NearestEntryDist(st.lastX, st.lastY)
			approachV = st.approachSpeedCmS
			reachE = reachableExitScore(exitDist, approachV)
			if reachE > 0 {
				tobs = append(tobs, belief.TObservation{Kind: belief.TObsReachableExit, Value: reachE, Conf: 0.8, Ts: nowMs, Fresh: true})
			}
		}
		tl.tb.Step(nowMs, tobs)
		v := tl.tb.Vector()
		pLost := v.P(belief.TLost)
		if pLost >= tLostLogThresh && !tl.loggedLo {
			argT, argP := v.Max()
			e.logger.Info("belief_shadow_track_lost", // 仅 log，无 alarm
				zap.String("room_id", roomID),
				zap.String("device_addr", tl.device),
				zap.Int("track_id", tid),
				zap.Int64("ts_ms", nowMs),
				zap.Float64("p_tlost", pLost),
				zap.String("argmax_tstate", argT.String()),
				zap.Float64("argmax_p", argP),
				zap.String("last_geom", tl.geom.String()),
				zap.Int("exit_dist_cm", exitDist),       // P2 诊断：丢失点离最近门距离
				zap.Float64("approach_v_cms", approachV), // A：实测朝门定向逼近速度（0=未逼近/测不出）
				zap.Float64("reach_cap_cms", reachCap),   // P2.1：本设备学习封顶（学习未足=全局 60）
				zap.Float64("reach_e", reachE),           // 可达退场分 e=f_dist·f_reach（0=不抑制）
			)
		}
		tl.loggedLo = pLost >= tLostLogThresh
	}

	sh.b.Step(nowMs, obs)
	confirmed := sh.decider.Update(sh.b.Vector(), nowMs) == belief.DecisionFall
	if confirmed && !sh.fired {
		v := sh.b.Vector()
		argS, argP := v.Max()
		e.logger.Info("belief_shadow_fall", // 仅 log，无 alarm
			zap.String("room_id", roomID),
			zap.Int64("ts_ms", nowMs),
			zap.Float64("p_fallen", v.P(belief.SFallen)),
			zap.String("argmax_state", argS.String()),
			zap.Float64("argmax_p", argP),
			zap.String("last_lost_geom", sh.lastLostGeom.String()), // #3：InBed/InToilet=床/桶区误确认嫌疑
		)
	}
	sh.fired = confirmed
}
