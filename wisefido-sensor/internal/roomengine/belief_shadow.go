package roomengine

import (
	"context"
	"os"

	"owl-common/alarm"
	"owl-common/card"
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

// dbnFireEnabled cutover 可逆开关（委员会 6c376e4 安全包络④）。env DBN_FIRE=1 → DBN 真发推断 fall（union
// firmware∨DBN）+ gate-list 推断 fall 短路；默认 OFF=现状（shadow log-only,gate-list 发）。秒翻回 gate-list。
// 部署此 commit 行为零变化(默认 OFF),翻开关才进 production——第一个碰 alarm 路径的提交,R0 在开关 ON 时结束。
var dbnFireEnabled = os.Getenv("DBN_FIRE") == "1"

// P6.1b-D(审查㉛ Opt-1)小卫生间 provisional/分级窗:
//
//	设备富(unit 有其它设备)→ 30min cancel 窗(覆盖立项 np=0 +335s),窗到未佐证升级。
//	设备贫(浴室独苗,无跨设备 cancel 可能)→ 短窗早决断压制(省真摔无谓延迟,resource-scaled v3)。
const (
	beliefProvisionalRichWindowMs = 30 * 60 * 1000 // 设备富:30min cancel 窗
	beliefProvisionalPoorWindowMs = 120 * 1000     // 设备贫:2min 早决断(无 cancel 可能)
)

type beliefShadowTrack struct {
	lastSeenMs       int64
	stillBoxAgeMs    int64 // 最后一帧时的 still-box 时长（消失前是否走动的依据）
	geom             belief.Geom
	lastX, lastY     int     // 最后一帧位置 → 算丢失点离门距离 d（P2 reachable-exit）
	lastRawDistCm    int     // P2 距离闸：最后一帧距雷达（raw 本地系原点）平面距 → 超 d_fall=贴地弱回波区
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
	logicID      string // 出生锚定的稳定逻辑身份（G-1 id-swap 守恒：失锁后查 logic_id 是否仍活在别 track）
	loggedLo     bool   // 已 log 过本次 Lost 峰（防重复）
	lastX, lastY int    // P3.1:上帧位置,算本帧空间跳跃 Δ/dt(独立 shadow realness 探测器①)
	lastPosTs    int64  // 上帧位置时刻
	lastPose     int    // P3.2:上帧 pose/z,算 pose/z 锁死帧数(冻结伪迹 B 佐证③)
	lastZ        int
	poseZLock    int     // pose&z 连续恒定帧数
	realLO       float64 // P3.3:realness log-odds(带遗忘 γ;>0 real <0 ghost),摔前走路 realness 带进倒地窗
}

type beliefShadow struct {
	b            *belief.Belief
	decider      belief.Decider
	tracks       map[int]*beliefShadowTrack
	fired        bool                        // 已 log 过本次 confirm（防 confirm 持续期重复 log）
	tlayer       map[int]*beliefShadowTLayer // DBN P1 Track 层
	deviceSpeed  map[string]*deviceSpeedStat // P2.1：per-device 学习走速封顶（device→room 稳定，跨 track 累积）
	lastLostGeom belief.Geom                 // #3：最近一次丢失点 geom（fall log 辨床/桶区误报用）
}

// tLostLogThresh Track 层 P(TLost) 超此即 log（对照 gate-list lost_track 报警）。
const tLostLogThresh = 0.5

// lostFarFromRadar P2 距离闸：丢轨点距雷达（raw 本地系原点）平面距 > d_fall（DistanceGateCm，≈500cm）
// → 贴地/躺姿弱回波区，跌倒有效检测半径远小于人员检测半径；该处"消失"是结构性盲猜非 fall 证据。
// 复用 gate 同一常量（R7 单源）；gate=0 关闭。
func lostFarFromRadar(rawDistCm int) bool {
	g := FallRulesParam.Lost.DistanceGateCm
	return g > 0 && rawDistCm > g
}

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
	if o.Fresh && o.Conf > 0 { // source-fidelity 审计:event 路径喂的 obs(EnterExit)也计入 populated
		e.logger.Debug("belief_shadow_obs", zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
			zap.Strings("populated_kinds", []string{o.Kind.String()}))
	}
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
	roomType := e.roomType[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)

	night := IsNightTime(nowMs, tm.timezone) // P4.3 风险时段:夜间 dwell 生存尾缩短(久静更可疑)

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
		var tlGhostness float64 // 单 track frozen ghostness:**仅留诊断日志**,P1② 后不再驱动 belief Ghost 发射
		tl.realLO, tlGhostness = realnessStep(tl.realLO, dtSec, b.MoveActive, jumpGhost, frozenGhost)
		tl.lastX, tl.lastY, tl.lastPosTs = b.X, b.Y, nowMs
		tl.lastPose, tl.lastZ = b.Pose, b.Z
		tlGeom := geomFromArea(b.CellAreaType)
		// ★P1②(发射换源):Ghostness 主源 = **co-existence 对称**(b.Verdict=gate-list 镜像/运动对称结果,
		// 正常处理时算好,lock-free)。**单 track frozenGhost/jumpGhost 不再驱动 Ghost 发射**(realLO 仅留日志)——
		// 治"真人静止 bystander 被单 track frozen 误判 ghost"(委员会细化1:≥2 真人)。endgame 换 DBN 自己的对称计算。
		// 标定 0.9/0.15 临时(待委员会:symmetry hit→Ghostness 走 calibration.go LR?)。
		ghostness := 0.15
		if b.Verdict == VerdictGhost {
			ghostness = 0.9
		}
		// ★P1①(co-existence 耦合):进 Ghost 的转移 ×ρ,ρ=房内其它 track 的 P(Real) 峰。孤立 track ρ=0 →
		// P(Ghost)→0(即使 Ghostness 高也救不回)→ long-lie 真受害者结构性安全。ghost=反射必有共存 Real partner。
		rho := dbnCoExistRho(sh, b.TrackID)
		tl.tb.StepCoupled(nowMs, []belief.TObservation{{
			Kind: belief.TObsPresent, Ghostness: ghostness, Geom: tlGeom,
			Conf: 0.9, Ts: nowMs, Fresh: true,
		}}, rho)
		tl.lastSeen = nowMs
		tl.geom = tlGeom
		tl.device = b.DeviceAddr
		if ts := tm.tracks[b.TrackID]; ts != nil {
			tl.logicID = ts.LogicID // G-1：趁活时 stash logic_id，失锁 sweep 查身份守恒
		}
		tl.loggedLo = false // 重新检出 → 允许后续再次 Lost 时重新 log

		// R0 结构化否决证据(委员会 276e852 升关键路径):ghost/frozen verdict 此前只喂 belief、不 log 成
		// 可消费形态(VerdictGhost 静默 delete+continue / frozenGhost 仅进 realLO)→ harness 测不到覆盖。
		// 仅多 log 不动作(下方 delete/continue 不变,shadow=生产同代码)。schema=harness vetoEvidence。
		if jumpGhost || frozenGhost || b.Verdict == VerdictGhost {
			reason := "shadow_realness_jump"
			if frozenGhost {
				reason = "shadow_realness_frozen"
			}
			if b.Verdict == VerdictGhost {
				reason = "production_verdict_ghost"
			}
			e.logger.Info("belief_shadow_veto_evidence",
				zap.String("room_id", roomID),
				zap.Int("track_id", b.TrackID),
				zap.Int64("ts_ms", nowMs),
				zap.Bool("ev_ghost", true),
				zap.Bool("ev_frozen", frozenGhost),
				zap.Bool("would_veto", true),
				zap.String("veto_reason", reason),
				zap.Float64("track_ghostness", tlGhostness),
			)
		}

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
			obs = append(obs, radarFrameAdapter(tr, ts, grid, nowMs, night)...)
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
				zap.Int64("p3_4_recapture_ms", nowMs-st.lastSeenMs), // 丢失多久后返回(cd2b≈5.85min)
				zap.Bool("p3_4_would_cancel", true),                 // production 会硬 cancel pending lost-fall
				zap.Bool("p3_4_self_rescue_candidate", true),        // → shadow 标低 severity 非抹掉
				zap.String("last_geom", st.geom.String()),
			)
		}
		st.lastSeenMs = nowMs
		st.lostAnchor = 0
		st.geom = geomFromArea(b.CellAreaType)
		st.lastX, st.lastY = b.X, b.Y
		if ts != nil { // P2 距离闸：stash 距雷达 raw 平面距（丢失后 ts 可能销毁，趁活时算）
			st.lastRawDistCm = distInt(ts.LastRawH, ts.LastRawV, 0, 0)
		}
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

		// 死源#3(Neighbor,shadow-first R0)：仅压**二义 lost-fall**——本房丢轨(摔/走二义)后 hand-off 窗内,
		// 邻房出现新鲜有向事件=人挪去邻房 → 喂 ObsNeighbor 压 phantom fall(直击 CABB/John.Y lost_track 主类)。
		// 铁律(非全空间监控)：stale/durable 不算,仅极近两事件可排除;N-3 sole-resident 门 + N-6 单合并。
		// corroboration≠substitution：邻房真出现正证据(occ),非裸 absence。已确证 fall 不经此路径(R5 许可压制)。
		if nb := e.neighborHandoff(roomID, st.lastSeenMs, nowMs); nb.hit {
			obs = append(obs, neighborToObs(1, nb.conf, nowMs))
			e.logger.Info("belief_shadow_neighbor_handoff", // 仅 log，无 alarm(R0)
				zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
				zap.Int64("n3_lost_seen_ms", st.lastSeenMs),
				zap.Float64("n3_neighbor_conf", nb.conf),
				zap.String("n3_neighbor_src", nb.src))
		} else if nb.staleOcc {
			// no-silent-caps(审查62)：邻房有 durable 占用但相关度 stale(留驻 gap)→ plan-A 不压(铁律:证不了此刻在哪),
			// 但不静默吞缺口 → LOG gap 量化 plan-B(静态人证兜底)价值。镜像 belief_shadow_recapture_skip_multiresident 数据闸。
			e.logger.Info("belief_shadow_neighbor_stale_corr",
				zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
				zap.Int64("n3_lost_seen_ms", st.lastSeenMs),
				zap.Int64("n3_stale_gap_ms", nb.staleGap),
				zap.String("n3_stale_src", nb.staleSrc),
				zap.Bool("n3_neighbor_occ", true))
		}

		// P6.1b-D(审查㉛ Opt-1)小卫生间 lost → provisional/分级状态机(替标准 reachableExit 抑制)。
		// 门距退化(处处近门)→ 不靠 door-distance;Fallen 经 NoDetect 真 ramp(dx=0),离场判别交 cancel 窗。
		// cancel 须**既 attribution-safe 又 leave-discriminating**(扩展不变量 审查㉝):正向离场证据,非缺证。
		// **cancel = recapture ONLY**(走失者本人 anchor 正向重现,per-identity,过两条)。
		// **np=0 永不 cancel**(审查㉝:realness-empty 看不到已丢失摔倒者→≈np=0;np=0 是 lost-fall 定义性条件,
		// 摔/离共有,非判别器)→ np=0 仅 aux LOG。可靠离场-cancel 升级 = Opt-3 边界穿越(非 np=0)。
		// 默认升级硬约束:歧义→escalate。设备富 30min cancel 窗;设备贫(独苗)短窗早决断→压制+LOG(v3 resource-scaled)。
		// **小卫生间整间都是门距退化区**(审查⑳"处处近门")→ D-path 对该房**任一 geom** 的 lost track 生效
		// (smallBath gate 已限定是小卫生间);**不限 InToilet/InEnter**——真 CABB layout 常无 toilet 对象,
		// 内部 lost track geom=OpenFloor,若限门区则 D-path 不 engage=silent-miss 治不了 CABB(审查㉟ 暴露)。
		if smallBath {
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
			np0Recent := tm.LastNumberPeopleZeroMs() > 0 && nowMs-tm.LastNumberPeopleZeroMs() <= beliefProvisionalRichWindowMs // 审查51:读统一 latch 派生视图(count==0),仅 aux LOG 不 gate cancel
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
		// P2 距离闸内化（gate→DBN）：丢轨点距雷达 > d_fall → 贴地/躺姿弱回波区，firmware 物理读不出地面态，
		// "消失" 零 fall 信息（结构性盲猜，非 fall 证据）→ no-detect **完全中性化**（effRealnessP=0 → factor→1）。
		// ★不走 door-exit 通道：door-exit 有意留 floor（noDetDoorSuppressK<1，「门口真摔仍浮出」），是"可能离开也
		// 可能摔"；距离闸是传感器盲区"零信息"，须全压 → 走 realnessP=0。复用 gate DistanceGateCm 常量（R7）。
		// G-2 空房账内化（gate→DBN，委员会自纠：原 G-2① 误判"已覆盖"）：roomLedgerEmpty（最近 ExitRoom 晚于
		// EnterRoom = 房已空，铁律只信 ExitRoom 不信 np=0）→ 此刻失锁 track 是"人走后残影"（冻住的镜面反射）→
		// no-detect 不抬 fall（治 D5F7 残影 ExitRoom 后才消失漏 cancel）。durable 状态读单源 roomLedgerEmpty(#1.3)，
		// 全压（同距离闸：ExitRoom 是确实离开的空间证据，残影零 fall 信息）。
		effRealnessP := realnessP
		distSuppress, roomEmptySuppress := 0.0, 0.0
		if lostFarFromRadar(st.lastRawDistCm) {
			distSuppress = 1.0
		}
		if tm.RoomLedgerEmpty() {
			roomEmptySuppress = 1.0
		}
		if distSuppress > 0 || roomEmptySuppress > 0 {
			effRealnessP = 0 // 距离盲区 ∨ 房已空 → no-detect 完全中性化（factor→1）
		}
		// P2：门区不再硬 continue；改软门——no-detect(R_i+door-exit 门控抬 Fallen) 与 reachable-exit(压 Fallen) 同 tick 对冲。
		obs = append(obs, noDetectObs(st.geom, effRealnessP, doorExitP, nowMs))
		if realnessP < 0.5 || doorExitP > 0.5 || distSuppress > 0 || roomEmptySuppress > 0 { // 门控生效→ 弱抬/不抬
			e.logger.Info("belief_shadow_nodetect_gated", // observability:量 no-detect 被 R_i/door-exit/dist/room-empty 压的频率
				zap.String("room_id", roomID),
				zap.Int("track_id", tid),
				zap.Int64("ts_ms", nowMs),
				zap.Float64("p6_1a_Ri", realnessP),                                                          // P(real track);低=ghost 消失,不抬（原始,非 dist-eff）
				zap.Float64("p6_1a_door_exit", doorExitP),                                                   // P(door-exit);高=门区可达走出,不抬
				zap.Float64("p2_dist_suppress", distSuppress), zap.Int("p2_lost_dist_cm", st.lastRawDistCm), // P2 距离闸:1=远距全压
				zap.Float64("p2_room_empty_suppress", roomEmptySuppress), // G-2 空房账:1=房空残影全压
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
		// G-1 id-swap 守恒（gate→DBN 内化 lost_fall_skipped_id_swap）：本 track 失锁但同 logic_id 仍活在
		// 另一条 live track = firmware 换 ID，无人真消失 → 喂 TObsLogicAlive 压 TLost（与生产 gate 硬 skip 同向）。
		if tm.HasOtherLiveTrackWithLogicID(tl.logicID, tid) {
			tobs = append(tobs, belief.TObservation{Kind: belief.TObsLogicAlive, Conf: 0.9, Ts: nowMs, Fresh: true})
		}
		// G-2 空房账 durable（Track 层一致：TObsExit 原仅 event 瞬时会衰减；房账空是 durable 状态 → 失锁期每 tick
		// 按 roomLedgerEmpty 重申 TObsExit 压 TLost，与 Room 层 no-detect 全压同向）。读单源 roomLedgerEmpty(#1.3)。
		if tm.RoomLedgerEmpty() {
			tobs = append(tobs, belief.TObservation{Kind: belief.TObsExit, Conf: 0.9, Ts: nowMs, Fresh: true})
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
				zap.Int("exit_dist_cm", exitDist),        // P2 诊断：丢失点离最近门距离
				zap.Float64("approach_v_cms", approachV), // A：实测朝门定向逼近速度（0=未逼近/测不出）
				zap.Float64("reach_cap_cms", reachCap),   // P2.1：本设备学习封顶（学习未足=全局 60）
				zap.Float64("reach_e", reachE),           // 可达退场分 e=f_dist·f_reach（0=不抑制）
			)
		}
		tl.loggedLo = pLost >= tLostLogThresh
	}

	// P5(审查㊾ 重裁)：wire 既有 bed 贝叶斯 Markov(bedAdapter→ObsBedOccupied)进 shadow → 占用概率压 SFallen
	// **无 radar-on-bed 要求**(治 0127:radar 误读 on-bed 人为 Floor-Fallen 时,可靠床占用仍压)。R5-clean
	// (压走接触式占用概率非 pose/z);漏报-safe(any-source-OR LeftBed → BedOccupancyState 占用降 → 释放 → Fall 浮出)。
	// = 一举两得:修 P5(退 radar-on-bed leg)+ wire 死源#1(BedOccupied)。retire bedLeakState/bedAuthorityObs(#1.2)。
	bedObs := bedAdapter(tm.BedOccupancyState(nowMs), nowMs)
	obs = append(obs, bedObs...)
	for _, o := range bedObs {
		if o.Kind == belief.ObsBedOccupied && o.Fresh && o.Value >= 0.5 {
			e.logger.Info("belief_shadow_bed_occupied_suppress", // 占用概率压 radar fall(无 radar-on-bed 要求)
				zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
				zap.Float64("p5_bed_occupancy", o.Value), zap.Float64("p5_bed_conf", o.Conf))
		}
	}

	// 死源#2(审查51 NP-1):wire firmware number_people 单 latch → ObsNumberPeople(弱 corroboration)。
	// likelihood 铁律:np=0 弱压 Empty/Left(corroborate 非 substitution,镜面 ghost/水气衰减都假报 0);np≥1 压 Empty。
	// R5:np 不进 SFallen(只调 Empty/Left/Occupied 间分布,不抬不压 fall);TTL 内才喂(stale 当缺失)。
	if npCount, npFresh := tm.CurrentNumberPeople(nowMs); npFresh {
		obs = append(obs, belief.Observation{
			Kind: belief.ObsNumberPeople, Value: float64(npCount), Conf: 0.8,
			Ts: nowMs, Fresh: true, Geom: belief.GeomUnknown,
		})
	}

	// source-fidelity 审计(审查㊺ 系统性逐 obs 对账):emit 本 tick 真路径**实际 fresh**(populate 了)的 obs 源
	// (Debug,生产默认不开;B 聚合每案 populated/dead obs 集,挖"obs 源≠生产实际源/死管"类 gap)。
	if len(obs) > 0 {
		seenKind := map[belief.ObsKind]bool{}
		var kinds []string
		for _, o := range obs {
			if o.Fresh && o.Conf > 0 && !seenKind[o.Kind] {
				seenKind[o.Kind] = true
				kinds = append(kinds, o.Kind.String())
			}
		}
		e.logger.Debug("belief_shadow_obs", zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
			zap.Strings("populated_kinds", kinds))
	}

	sh.b.Step(nowMs, obs)
	v := sh.b.Vector()
	// 每 tick belief 轨迹(Debug,生产默认不开;B harness/逐案查 observer 捕之取 P(Fallen) 峰值/末态)。
	// 委员会㊸ B-2:数值进生产 observability 非 test 钩子;为人在环逐案查提供"关键 P 值清 log"。
	argTraceS, argTraceP := v.Max()
	pFallen := v.P(belief.SFallen)
	// P7.1 base 代价比 τ* 读出 + P7.2 context-dependent τ*（zone+risk-time → C_FN 加权 → τ* 随风险降）。
	// shadow-first R0：只 log 不接 alarm；suspect/confirm 两工作点 vs gate-list 对账（P9 oracle 用）。
	tauDec, tauHit := belief.DecideTau(pFallen)
	tauCtx := belief.TauContext{Bathroom: roomType == card.RoomTypeBathroom, Night: night}
	tauCtxDec, tauCtxHit := belief.DecideTauCtx(pFallen, tauCtx)
	e.logger.Debug("belief_shadow_trace",
		zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
		zap.Float64("p_fallen", pFallen),
		zap.String("argmax_state", argTraceS.String()), zap.Float64("argmax_p", argTraceP),
		zap.String("last_lost_geom", sh.lastLostGeom.String()),
		zap.String("p7_1_tau_decision", tauDec.String()), zap.Float64("p7_1_tau", tauHit),
		zap.Bool("p7_2_bathroom", tauCtx.Bathroom), zap.Bool("p7_2_night", tauCtx.Night),
		zap.String("p7_2_tau_decision", tauCtxDec.String()), zap.Float64("p7_2_tau", tauCtxHit))
	confirmed := sh.decider.Update(v, nowMs) == belief.DecisionFall
	if confirmed && !sh.fired {
		argS, argP := v.Max()
		// P7.3 reason 路由：由把 SFallen 拉最高的 fall-ward obs 节点决定成因（lost/silent/pose_lying），
		// 取代散落 reason 函数；shadow-first R0 只读出（vs gate-list reason 对账 P9）。
		p7Reason, p7Dom, p7LR := belief.FallReasonFor(obs)
		// P7.4 human-bed 豁免 veto（decision 层前置短路）：fall 位置（present track ∨ lost 锚点）在 Conf≥99
		// 人工床 = 躺床 normal use 非跌倒 → 生产前置短路不进 τ* 判决。shadow-first R0：只读出 would-veto，不改发射。
		var fallPos [][2]int
		for i := range bases {
			fallPos = append(fallPos, [2]int{bases[i].X, bases[i].Y})
		}
		for _, st := range sh.tracks {
			if nowMs-st.lastSeenMs > beliefShadowLostTTLMs {
				fallPos = append(fallPos, [2]int{st.lastX, st.lastY})
			}
		}
		p7HumanBedVeto := humanBedVetoAt(grid, fallPos)
		e.logger.Info("belief_shadow_fall", // 仅 log，无 alarm
			zap.String("room_id", roomID),
			zap.Int64("ts_ms", nowMs),
			zap.Float64("p_fallen", pFallen),
			zap.String("argmax_state", argS.String()),
			zap.Float64("argmax_p", argP),
			zap.String("last_lost_geom", sh.lastLostGeom.String()), // #3：InBed/InToilet=床/桶区误确认嫌疑
			zap.String("p7_1_tau_decision", tauDec.String()),       // P7.1：base confirm（τ*=0.55，等价历史 θ_fire）
			zap.Float64("p7_1_tau_confirm", belief.TauConfirm.Tau()),
			zap.Bool("p7_2_bathroom", tauCtx.Bathroom), zap.Bool("p7_2_night", tauCtx.Night), // P7.2：context
			zap.String("p7_2_tau_decision", tauCtxDec.String()), zap.Float64("p7_2_tau", tauCtxHit),
			zap.String("p7_3_reason", p7Reason.String()), // P7.3：主导节点成因
			zap.String("p7_3_dominant_obs", p7Dom.String()), zap.Float64("p7_3_dominant_lr", p7LR),
			zap.Bool("p7_4_human_bed_veto", p7HumanBedVeto), // P7.4：Conf≥99 人工床 → 生产前置短路（R0 只读出）
			zap.Int("p7_4_min_conf", humanBedExemptMinConfidence),
		)
		// ★cutover（委员会 6c376e4 envelope）:开关 ON → DBN 真发推断 fall（union firmware∨DBN）。
		// 保守 ghost veto（envelope③）+ 默认放行;bed/recovery veto 暂不开。R0 在开关 ON 时结束(计划内)。
		if dbnFireEnabled {
			fb, ok := dbnFallerBase(bases, sh)
			// ★★ghost 判据 = co-existence（用户铁律,安全包线）：ghost 是真人的反射 → 必有共存的真人 partner
			// → 画面里 **≥2 track**;**孤立 1 track = 没有被反射的源 = 不可能是 ghost = 一定是真人 → 永远发,绝不否**
			// (护住 long-lie 躺地不动的真受害者——他越不动越像 frozen,但孤立就证明他是真人)。**绝不**用 realLO/
			// present-realness 单 track 判 ghost(那把躺地真受害者误判=灾难)。只 **≥2 track 且 faller 是 VerdictGhost**
			// (gate-list 镜像/运动对称,触发即需另一 Real partner)才否。endgame:DBN 自己的 frozenGhost/realLO 是
			// 单 track 病根,待重做成 co-existence(找 Real partner),不再依赖 gate-list VerdictGhost。
			coExist := len(bases) >= 2 // 有共存的 partner(否则孤立 = real)
			if ok && coExist && fb.Verdict == VerdictGhost {
				e.logger.Info("belief_dbn_veto_ghost", zap.String("room_id", roomID), // co-existence ghost 抑制
					zap.Float64("p_fallen", pFallen), zap.String("p7_3_reason", p7Reason.String()), zap.Int("track_count", len(bases)))
			} else if ok {
				e.PublishAIAlarm(context.Background(), AIPayload{
					DeviceAddr: fb.DeviceAddr, RoomID: roomID,
					Track: observation.Track{
						BedStatus: observation.BedStatusUnchanged, TrackID: fb.TrackID, Pose: fb.Pose,
						PositionX: intPtr(fb.RawH), PositionY: intPtr(fb.RawV), PositionZ: intPtr(fb.RawZ),
					},
					Reason:     "dbn_" + p7Reason.String(), // 分类 tag:dbn_lost/dbn_silent/dbn_pose_lying
					Evidence:   map[string]interface{}{"p_fallen": pFallen, "dominant_obs": p7Dom.String(), "bathroom": tauCtx.Bathroom},
					IncidentMs: nowMs,
				}, alarm.Fall, nowMs)
				e.logger.Info("belief_dbn_fire", zap.String("room_id", roomID), // R0 在此结束:DBN 真发告警
					zap.String("reason", "dbn_"+p7Reason.String()), zap.Float64("p_fallen", pFallen), zap.Int("track_id", fb.TrackID))
			}
		}
	}
	sh.fired = confirmed
}

// dbnCoExistRho P1①:共存 Real 信念 ρ=房内**其它** track 的 P(TReal) 峰(选 max,委员会待定 vs 软 OR)。
// ghost=真人反射 → 必有共存 Real partner 才可能;孤立(无其它 Real track)ρ=0 → 转移进 Ghost 倾向=0。
func dbnCoExistRho(sh *beliefShadow, selfTid int) float64 {
	rho := 0.0
	for tid, tl := range sh.tlayer {
		if tid == selfTid || tl.tb == nil {
			continue
		}
		if p := tl.tb.Vector().P(belief.TReal); p > rho {
			rho = p
		}
	}
	return rho
}

// dbnFallerBase cutover:取摔者位置作 alarm payload。present 取 z 最低(最倒地);否则 lost 用最近 tlayer 锚点。
func dbnFallerBase(bases []TrackStatusBase, sh *beliefShadow) (TrackStatusBase, bool) {
	if len(bases) > 0 {
		fb := bases[0]
		for _, b := range bases[1:] {
			if b.RawZ < fb.RawZ {
				fb = b
			}
		}
		return fb, true
	}
	var best *beliefShadowTLayer
	var bestTid int
	for tid, tl := range sh.tlayer {
		if best == nil || tl.lastSeen > best.lastSeen {
			best, bestTid = tl, tid
		}
	}
	if best != nil && best.device != "" {
		return TrackStatusBase{DeviceAddr: best.device, TrackID: bestTid, RawH: best.lastX, RawV: best.lastY}, true
	}
	return TrackStatusBase{}, false
}
