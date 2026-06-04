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

type beliefShadowTrack struct {
	lastSeenMs       int64
	stillBoxAgeMs    int64 // 最后一帧时的 still-box 时长（消失前是否走动的依据）
	geom             belief.Geom
	lastX, lastY     int     // 最后一帧位置 → 算丢失点离门距离 d（P2 reachable-exit）
	approachSpeedCmS float64 // A：丢失前朝门定向逼近速度（在 track 仍活时算好 stash，丢失后 ts 可能已销毁）
	lostAnchor       int64
}

// beliefShadowTLayer DBN P1 Track 层（per-track T_t）。与 Room 层 tracks 刻意分离：
// Room 层对 ghost 用 delete+method-2 gate；Track 层**保留 ghost track**让 A_T 结构化处理
// （Ghost→None，不通 Lost），正是 P1 要对照的"gate vs A_T 一致性"。
type beliefShadowTLayer struct {
	tb       *belief.TrackBelief
	lastSeen int64
	geom     belief.Geom
	device   string // 源雷达 device_addr（同房对等雷达占用对账排除自身用）
	loggedLo bool   // 已 log 过本次 Lost 峰（防重复）
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

		// DBN P1 Track 层：每帧（含 ghost）喂 present 发射。ghostness 由 GhostPenalty。
		// 与下方 Room 层 ghost-delete 刻意分离：Track 层让 A_T 把 ghost 路由 →None。
		curTL[b.TrackID] = struct{}{}
		tl := sh.tlayer[b.TrackID]
		if tl == nil {
			tl = &beliefShadowTLayer{tb: belief.NewTrackBelief()}
			sh.tlayer[b.TrackID] = tl
		}
		tlGhostness := float64(b.GhostPenalty) / 100.0
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
		tr := observation.Track{Pose: b.Pose, PositionX: &x, PositionY: &y, PositionZ: &z}
		ts := tm.tracks[b.TrackID]
		if ts != nil {
			obs = append(obs, radarFrameAdapter(tr, ts, grid, nowMs)...)
		}
		st := sh.tracks[b.TrackID]
		if st == nil {
			st = &beliefShadowTrack{}
			sh.tracks[b.TrackID] = st
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
		// P2：门区不再硬 continue；改软门——no-detect(状态条件抬 Fallen) 与 reachable-exit(近门+定向逼近→压 Fallen) 同 tick 对冲。
		obs = append(obs, noDetectObs(st.geom, nowMs))
		sh.lastLostGeom = st.geom // #3：记最近丢失点 geom，供 fall log 辨别床/桶区
		if grid != nil {
			obs = append(obs, reachableExitObs(grid.NearestEntryDist(st.lastX, st.lastY), st.approachSpeedCmS, st.geom, nowMs))
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
