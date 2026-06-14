package roomengine

import (
	"context"
	"math"
	"os"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"owl-common/radarutils"
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

// dbnMode cutover 三档（用户拍板,可逆,运维 .env 翻 DBN_MODE）。两条正交轴=否决 firmware × DBN 自发:
//
//	0 = DBN 静默：不自发 + 不否决；firmware 所有 fire 直接通过（不进 stash/resolve,禁 DBN 处理；学习期安全档,firmware 仍兜底真摔）
//	1 = 不否决 firmware fire + DBN 自发（firmware 地板不可挡,union；最不漏 firmware）
//	2 = 否决 firmware fire + DBN 自发（全开,DBN 权威）
//
// 否决 firmware 的依据 = 复用 DBN 自发的 co-existence ghost + 风险分层 τ*（仅档 2）。
// 未设/非法 → 1（保 firmware 地板不可挡,最保守不漏 firmware）。秒翻档回滚。
var dbnMode = parseDBNMode(os.Getenv("DBN_MODE"))

func parseDBNMode(s string) int {
	switch s {
	case "0":
		return 0
	case "2":
		return 2
	default:
		return 1
	}
}

// ── 工单5 cold-start per-unit 升档闸（委员会 2026-06-13 §6 裁决，Phase A 纯内存）──
//
// 全局 dbnMode（DBN_MODE env）是 ceiling，每 unit 另有成熟度 cap，effectiveMode=min(两者)：
// 自发=effectiveMode≥1，否决 firmware=effectiveMode==2。unit 刚上线 cell grid 空白
// （tolerance 全 1.0 无学得语义）→ dbnMode=2 会在常坐区密集 dwell-silent FP，故 cap 启动=1
// （可自发但不否决：firmware 兜底真摔，cd2b 这类 firmware 漏判靠 DBN 自发补，FP>>FN 安全），
// grid 学满后才升 2。成熟判据=纯时钟（§6.1 裁 C，覆盖率门首版不做）:
// now−firstTrack ≥ max(T_cold, T_floor)。T_cold 默认 7d 可 env 覆盖；T_floor=24h 硬下限。
// 默认值用户拍 7d（覆盖完整周循环,委员会 §6.1 原定 72h 太短改回提案本意 1 周）。
const coldFloorMs int64 = 24 * 60 * 60 * 1000 // §6.2 硬下限,不可 env 覆盖

var coldGraduateMs = parseColdHours(os.Getenv("DBN_COLD_HOURS")) // 默认 7d

func parseColdHours(s string) int64 {
	if s != "" {
		if h, err := time.ParseDuration(s + "h"); err == nil && h > 0 {
			return h.Milliseconds()
		}
	}
	return 7 * 24 * 60 * 60 * 1000
}

// markUnitTrack 在 unit 首次见 track 时打桩 firstTrackMs（单调,只落一次不覆盖）。
func (e *Engine) markUnitTrack(suiteID string, nowMs int64) {
	if suiteID == "" {
		return
	}
	e.coldStartMu.Lock()
	if e.unitFirstTrackMs[suiteID] == 0 {
		e.unitFirstTrackMs[suiteID] = nowMs
	}
	e.coldStartMu.Unlock()
}

// unitCap 本 unit 当前成熟度档：未满 cold 期=1（不否决 firmware）；满=2（可否决）。
// firstTrack=0（从没见过 track）→ 全新 unit → 1。
func (e *Engine) unitCap(suiteID string, nowMs int64) int {
	e.coldStartMu.Lock()
	first := e.unitFirstTrackMs[suiteID]
	e.coldStartMu.Unlock()
	graduate := coldGraduateMs
	if coldFloorMs > graduate {
		graduate = coldFloorMs
	}
	if first == 0 || nowMs-first < graduate {
		return 1
	}
	return 2
}

// dbnSelfFireEnabledFor / dbnVetoFirmwareEnabledFor 是 dbnSelfFireEnabled/dbnVetoFirmwareEnabled
// 的 per-unit 版：effectiveMode = min(全局 dbnMode, unitCap)。dbnMode=0 时 min 仍 0（运维全局静默不变）。
func (e *Engine) dbnSelfFireEnabledFor(suiteID string, nowMs int64) bool {
	return min(dbnMode, e.unitCap(suiteID, nowMs)) >= 1
}

func (e *Engine) dbnVetoFirmwareEnabledFor(roomID string, nowMs int64) bool {
	return min(dbnMode, e.unitCap(e.SuiteIDForRoom(roomID), nowMs)) == 2
}

// dbnVetoRecoveryEnabled P2 recovery-veto 独立子开关（委员会 6dafdc6,默认 OFF,可单独 live toggle）。
// recovery-veto 漏报-safe by construction(同人+正向 up+track-loss 螺丝+self-rescue≥15s+默认放行)。
// OFF=只 log belief_dbn_recovery_evidence(R0);ON=窗内纯误火 recovery 时抑制/auto-resolve(留 cutover)。
var dbnVetoRecoveryEnabled = os.Getenv("DBN_VETO_RECOVERY") == "1"

// recoveryGenuineFallenMs P2 安全判别:firmware 火后持续倒地超此=**自救型真摔**(倒了又起)≠纯误火 →
// 禁 recovery(自救真摔不可静默抹,呼应 silent_leftbed)。15s≈真倒地 vs firmware 误火即起的分界(实证 #9 73s/5934 ~0)。
const recoveryGenuineFallenMs = 15_000

// recoveryUprightSustainMs P2:摔者持续直立须达此才认正向恢复(防 1 帧 tracking 伪迹把真受害者误判 walk)。
const recoveryUprightSustainMs = 3_000

// movingFallRecentMs P3 moving_fall:摔前此窗内有 MoveActive = 移动中倒地(dbn_moving),否则开阔地静躺(dbn_pose_lying)。
const movingFallRecentMs = 5_000

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
	lastAreaType     int   // 消失前 last cell.area_type：toilet 守恒判定 + 诊断日志
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
	lastAreaType int    // 消失前 last cell.area_type：TObsAbsent 分流去向
	device       string // 源雷达 device_addr（同房对等雷达占用对账排除自身用）
	logicID      string // 出生锚定的稳定逻辑身份（G-1 id-swap 守恒：失锁后查 logic_id 是否仍活在别 track）
	loggedLo     bool   // 已 log 过本次 Lost 峰（防重复）
	lastX, lastY int    // P3.1:上帧位置,算本帧空间跳跃 Δ/dt(独立 shadow realness 探测器①)
	lastPosTs    int64  // 上帧位置时刻
	lastPose     int    // P3.2:上帧 pose/z,算 pose/z 锁死帧数(冻结伪迹 B 佐证③)
	lastZ        int
	poseZLock    int     // pose&z 连续恒定帧数
	realLO       float64 // P3.3:realness log-odds(带遗忘 γ;>0 real <0 ghost),摔前走路 realness 带进倒地窗

	// P2 recovery-veto:per-track upright/fallen 连续段 + 首现(同人绑定)。
	uprightSince int64 // 连续直立(Stand/Move)起始 ms(0=非直立);firmware 火后同人持续直立=recovery
	fallenSince  int64 // 连续倒地(Lie/Fallen/SitGround)起始 ms;火后持续倒地≥阈=自救真摔禁 recovery
	firstSeenMs  int64 // 首现 ms(护工后进=新 track firstSeen>T_fire→排除,同人绑定)
	lastMoveMs   int64 // P3 moving_fall:最后一次 MoveActive 时刻(摔前在动 → 移动中倒地=moving 非开阔静躺)
}

type beliefShadow struct {
	b            *belief.Belief
	decider      belief.Decider
	tracks       map[int]*beliefShadowTrack
	fired        bool                        // 已 log 过本次 confirm（防 confirm 持续期重复 log）
	loggedBedSuppress bool // 已 log 过本次 bed_occupied_suppress（防重复）
	tlayer       map[int]*beliefShadowTLayer // DBN P1 Track 层
	deviceSpeed  map[string]*deviceSpeedStat // P2.1：per-device 学习走速封顶（device→room 稳定，跨 track 累积）
	lastLostArea int                         // #3：最近一次丢失点 cell.area_type（fall log 辨床/桶区误报用）

	// P2 recovery-veto:firmware Fall 时刻(T_fire)= post-fire 倒地时长/recovery 窗的参考(engine hook 喂)。
	firmwareFallTs      int64 // firmware Fall 事件 ts(0=未火);recovery 须 ts>此
	recoveryGenuineFall bool  // firmware 火后摔者持续倒地≥阈=自救真摔→禁 recovery(不静默抹)
	recoveryEmitted     bool  // 已 emit 本次 recovery(防重复)

	// pendingFwFalls 档 2:firmware fall 异步到达时**不立即裁决**(那会用 cache-time 的旧 co-existence,
	// partner 消失后误否孤立真摔=漏报)。暂存,到下一 beliefShadowTick 用**当帧 bases** 现算 co-existence+ghost
	// 裁决(与 self-fire 同源同 tick,结构上消除"事件 vs tick"竞态)。孤立(当帧<2 track)→ 必发(铁律)。
	pendingFwFalls []pendingFwFall
}

// pendingFwFall 档 2 暂存的 firmware fall:待下一 tick 用 fresh bases 裁决放行/否决。
type pendingFwFall struct {
	a         RadarFallAlarm
	arrivedMs int64
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
	o, ok := radarEventToObs(eventName, nowMs)
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
	unitSuiteID := e.roomSuiteID[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return
	}
	// 工单5 cold-start:本 unit 见到 track → 打桩 firstTrack（升档计时起点）。
	if len(bases) > 0 {
		e.markUnitTrack(unitSuiteID, nowMs)
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)

	night := IsNightTime(nowMs, tm.timezone) // P4.3 风险时段:夜间 dwell 生存尾缩短(久静更可疑)

	// bedside FN 止血(委员会裁 a22b6c1,bed_state 驱动):床态取一次,供 pose-geom 翻转 + 下方 bedAdapter 复用。
	// bedReleased = 有床数据且离床 → 床区里躺着的 track 不是睡觉而是床边真摔(等价 silent_leftbed 判据)。
	// 无床数据(无 sleepad=无解,不强报防睡觉 FP)/ 占用 → 维持在床躺。
	bedSt := tm.BedOccupancyState(nowMs)
	bedReleased := bedSt.BedConfidence > 0 && bedSt.BedStatus != 0

	var obs []belief.Observation
	cur := make(map[int]struct{}, len(bases))
	curTL := make(map[int]struct{}, len(bases))
	// ★P1-final:循环前快照各 track 上帧位置(tlayer.lastX 循环中会被更新→避 ordering 坑),算 motion 对称的运动向量。
	prevPos := make(map[int][2]int, len(sh.tlayer))
	for tid, tl := range sh.tlayer {
		if tl.lastPosTs > 0 {
			prevPos[tid] = [2]int{tl.lastX, tl.lastY}
		}
	}
	interferes := tm.GetInterferes() // P1-final 增量2:反射面快照(线程安全),算 mirror 静态对称
	coExist := len(bases) >= 2                    // 共存 partner(ghost=真人反射必有共存 Real;孤立=real)
	ghostTracks := make(map[int]bool, len(bases)) // per-track co-existence ghost(本 tick;firmware-veto 裁决用)
	for i := range bases {
		b := &bases[i]

		// DBN P1 Track 层：每帧（含 ghost）喂 present 发射。
		// P3.1(委员会审查⑦裁定):ghostness 由**独立 shadow realness**(XY 三探测器)算,
		// **不复用生产 GhostPenalty/Verdict**(那套漏 cd2b 冻结 ghost)→ shadow R_i 才能抓生产漏的 ghost。
		// 与下方 Room 层 ghost-delete 刻意分离：Track 层让 A_T 把 ghost 路由 →None。
		curTL[b.TrackID] = struct{}{}
		tl := sh.tlayer[b.TrackID]
		if tl == nil {
			tl = &beliefShadowTLayer{tb: belief.NewTrackBelief(), firstSeenMs: nowMs} // P2:首现=同人绑定锚
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
		if b.MoveActive {
			tl.lastMoveMs = nowMs // P3 moving_fall:记最后在动时刻(摔前在动→移动中倒地)
		}
		// P2 recovery-veto:维护 upright(Stand/Move)/fallen(Lie/SitGround)连续段(条件③持续非瞬时)。
		switch RadarPoseToCore(b.Pose) {
		case CorePoseStand, CorePoseMove:
			if tl.uprightSince == 0 {
				tl.uprightSince = nowMs
			}
			tl.fallenSince = 0
		case CorePoseLie:
			if tl.fallenSince == 0 {
				tl.fallenSince = nowMs
			}
			tl.uprightSince = 0
		default: // 坐/未知:两者都清(不累计)
			tl.uprightSince, tl.fallenSince = 0, 0
		}
		// ★P1-final 完成(发射换源):Ghostness 全由 **DBN 自有 co-existence 对称**(motion 紧贴同向多径 + mirror
		// 反射面镜像位静态 cd2b 冻结类)驱动,**不再读 gate-list b.Verdict**——删 gate-list 真前置(ghost 检测面)达成。
		// 单 track frozenGhost/jumpGhost 不再驱动 Ghost 发射(realLO 仅留日志);配 P1① 耦合孤立 ρ=0→P(Ghost)=0。
		ghostness := 0.15
		symGhost := dbnMotionSymmetryGhost(b, bases, prevPos) || dbnMirrorSymmetryGhost(b, bases, interferes)
		if symGhost {
			ghostness = 0.9
		}
		ghostTracks[b.TrackID] = coExist && symGhost // firmware-veto 缓存:复用 self-fire 的 co-existence ghost 判据
		// ★P1①(co-existence 耦合):进 Ghost 的转移 ×ρ,ρ=房内其它 track 的 P(Real) 峰。孤立 track ρ=0 →
		// P(Ghost)→0(即使 Ghostness 高也救不回)→ long-lie 真受害者结构性安全。ghost=反射必有共存 Real partner。
		rho := dbnCoExistRho(sh, b.TrackID)
		tl.tb.StepCoupled(nowMs, []belief.TObservation{{
			Kind: belief.TObsPresent, Ghostness: ghostness,
			Conf: 0.9, Ts: nowMs, Fresh: true,
		}}, rho)
		tl.lastSeen = nowMs
		tl.lastAreaType = int(b.CellAreaType)
		tl.device = b.DeviceAddr
		if ts := tm.tracks[b.TrackID]; ts != nil {
			tl.logicID = ts.LogicID // G-1：趁活时 stash logic_id，失锁 sweep 查身份守恒
		}
		tl.loggedLo = false // 重新检出 → 允许后续再次 Lost 时重新 log

		// R0 结构化否决证据(委员会 276e852):realness 基础的 ghost/frozen track 每 tick forensic emit。
		// gate-list 删后 VerdictGhost 停产→b.Verdict==VerdictGhost 条件永假,已清 #1.2。
		if jumpGhost || frozenGhost {
			reason := "shadow_realness_jump"
			if frozenGhost {
				reason = "shadow_realness_frozen"
			}
			e.logger.Info("ghost_veto",
				zap.String("reason", reason),
				zap.String("room_id", roomID),
				zap.Int("track_id", b.TrackID),
				zap.Int64("ts_ms", nowMs),
				zap.Bool("ev_ghost", true),
				zap.Bool("ev_frozen", frozenGhost),
				zap.Bool("would_veto", true),
				zap.Float64("track_ghostness", tlGhostness),
			)
		}

		// gate-list 删后 VerdictGhost 永不设→Room 层 ghost track delete 条件永假,已清 #1.2。
		// (原逻辑:real→ghost→消失镜面反射误触发 lost-while-moving→delete 防此,现 DBN 自有 ghost 检测处理)
		cur[b.TrackID] = struct{}{}
		x, y, z := b.X, b.Y, b.Z
		tr := observation.Track{BedStatus: observation.BedStatusUnchanged, Pose: b.Pose, PositionX: &x, PositionY: &y, PositionZ: &z}
		ts := tm.tracks[b.TrackID]
		if ts != nil {
			// ★P1③(真伪前置 = 耦合加权,非顺序 gate):本 track 的 fall 证据(dwell/pose)Conf ×P(T=Real)。
			// ghost(pReal 低)→ 证据 Conf 低 → likelihood 趋中性 → **喂不动 SFallen**(不会被 dwell 喂成假 still-fall)。
			// real lone faller pReal≈1(孤立 ρ=0 不可能 ghost)→ 无折扣 → 真摔照常 ramp。
			pReal := tl.tb.Vector().P(belief.TReal)
			for _, o := range radarFrameAdapter(tr, ts, grid, nowMs, night) {
				o.Conf *= pReal
				if o.Kind == belief.ObsDwellStill {
					o.RoomType = roomType
					o.AreaType = int(b.CellAreaType)
				}
				// ★bedside FN 止血(D3):bed_state 离床 → 标 BedReleased,poseLikelihood 据此把床区(area=Bed)躺
				// 翻成倒地候选(取代旧 geom 翻转;门优先在 poseLikelihood 内复刻)。床/摔判别由 bed_state 定,非几何。
				if bedReleased {
					o.BedReleased = true
					if o.Kind == belief.ObsPose && int(o.Value) == observation.PoseLying &&
						o.AreaType == int(AreaBed) && !o.NearDoor {
						e.logger.Info("belief_dbn_bedside_unbed", zap.String("room_id", roomID),
							zap.Int("track_id", b.TrackID), zap.Int64("ts_ms", nowMs),
							zap.Int("bed_status", bedSt.BedStatus), zap.Int("bed_conf", bedSt.BedConfidence))
					}
				}
				obs = append(obs, o)
			}
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
				zap.String("last_area", belief.AreaTypeLabel(st.lastAreaType)),
			)
		}
		st.lastSeenMs = nowMs
		st.lostAnchor = 0
		st.lastAreaType = int(b.CellAreaType)
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
		// P5 定稿(silent vs moving):消失前曾静止(有 still-box)= silent —— lost 算作原位 still-box 持续。
		// dwell 连续(含 loss 前 still 时长)喂 survival ramp,**不进 moving 路径**(neighbor/noDetect/Blind)。
		// 治本译反:原此处 `continue 归 Still-fall 域` 把"倒地→静止→被丢"两头不沾(lost 排除 + 无帧发不出 dwell)=silent-miss。
		// recapture 框内续/框外断由 present 路径的 still-box 维护(track_manager)处理:返回框内→StillBoxRunStart 续→
		// stillBoxAgeMs 大;返回框外→track_manager 重置→stillBoxAgeMs 归零→自然不再 silent。
		if st.stillBoxAgeMs >= int64(FallRulesParam.Lost.MovingPreconditionMs) {
			dwellArea := st.lastAreaType
			if bedReleased && dwellArea == int(AreaBed) {
				dwellArea = int(AreaActive) // bed_state 离床 → 床区静止=床边真摔,按开阔地 dwell(否则 bed zone 不 ramp)
			}
			dwellTol := 1.0
			if grid != nil && isOpenFloorArea(AreaType(dwellArea)) {
				dwellTol = grid.ToleranceFactorAt(st.lastX, st.lastY)
			}
			contDwellMs := st.stillBoxAgeMs + (nowMs - st.lastSeenMs) // 连续时钟:loss 前 still + loss 后流逝
			obs = append(obs, belief.Observation{
				Kind: belief.ObsDwellStill, Value: float64(contDwellMs) / 1000, Conf: 0.7, Ts: nowMs, Fresh: true,
				ToleranceFactor: dwellTol, RoomType: roomType, AreaType: dwellArea, Night: night, RadarDistCm: st.lastRawDistCm,
			})
			e.logger.Info("belief_shadow_silent_dwell", // silent:倒地后静止被丢,dwell 连续 ramp(替译反的 silent-miss)
				zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
				zap.Int64("p5_cont_dwell_ms", contDwellMs),
				zap.Int64("p5_still_before_loss_ms", st.stillBoxAgeMs),
				zap.String("dwell_area", belief.AreaTypeLabel(dwellArea)),
				zap.Bool("p5_bed_released", bedReleased))
			continue
		}
		// 以下 = moving:消失前在走动(无 still-box)。neighbor hand-off / smallBath / noDetect→Blind / reachable-exit。
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
					zap.String("last_area", belief.AreaTypeLabel(st.lastAreaType)),
					zap.Int64("still_box_age_ms", st.stillBoxAgeMs))
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
				obs = append(obs, noDetectObs(realnessP, 0, nowMs)) // dx=0 干净 ramp(门距退化,disambiguation 交 cancel 窗)
				if elapsed >= beliefProvisionalRichWindowMs && !st.provisionalResolved {
					e.logger.Info("belief_shadow_lostfall_escalate", // 窗到未佐证 → 全 sev 真摔(延迟但不漏)
						zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
						zap.Int64("p6_1b_provisional_ms", elapsed))
					st.provisionalResolved = true
				}
			} else { // 设备贫(浴室独苗):无跨设备 cancel 可能 → 短窗早决断
				if elapsed < beliefProvisionalPoorWindowMs {
					obs = append(obs, noDetectObs(realnessP, 0, nowMs)) // 短窗内仍 ramp(provisional)
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
		if (st.lastAreaType == int(AreaToilet) || st.lastAreaType == int(AreaShower)) && e.suiteCensus != nil {
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
					zap.String("last_area", belief.AreaTypeLabel(st.lastAreaType)),
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
		obs = append(obs, noDetectObs(effRealnessP, doorExitP, nowMs))
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
				zap.String("last_area", belief.AreaTypeLabel(st.lastAreaType)),
			)
		}
		sh.lastLostArea = st.lastAreaType // #3：记最近丢失点 area，供 fall log 辨别床/桶区
		if grid != nil {
			obs = append(obs, reachableExitObs(exitDist, st.approachSpeedCmS, nowMs))
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
			Kind: belief.TObsAbsent, AreaType: tl.lastAreaType, Conf: 0.9, Ts: nowMs, Fresh: true,
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
				zap.String("last_area", belief.AreaTypeLabel(tl.lastAreaType)),
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
	bedObs := bedAdapter(bedSt, nowMs) // 复用 tick 顶部已取的床态(同 nowMs,免重复 BedOccupancyState 锁)
	obs = append(obs, bedObs...)

	suppressActive := false
	for _, o := range bedObs {
		if o.Kind == belief.ObsBedOccupied && o.Fresh && o.Value >= 0.5 {
			suppressActive = true
			if !sh.loggedBedSuppress {
				sh.loggedBedSuppress = true
				e.logger.Info("belief_shadow_bed_occupied_suppress", // 占用概率压 radar fall(无 radar-on-bed 要求)
					zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
					zap.Float64("p5_bed_occupancy", o.Value), zap.Float64("p5_bed_conf", o.Conf))
			}
		}
	}
	if !suppressActive {
		sh.loggedBedSuppress = false
	}

	// 死源#2(审查51 NP-1):wire firmware number_people 单 latch → ObsNumberPeople(弱 corroboration)。
	// likelihood 铁律:np=0 弱压 Empty/Left(corroborate 非 substitution,镜面 ghost/水气衰减都假报 0);np≥1 压 Empty。
	// R5:np 不进 SFallen(只调 Empty/Left/Occupied 间分布,不抬不压 fall);TTL 内才喂(stale 当缺失)。
	if npCount, npFresh := tm.CurrentNumberPeople(nowMs); npFresh {
		obs = append(obs, belief.Observation{
			Kind: belief.ObsNumberPeople, Value: float64(npCount), Conf: 0.8,
			Ts: nowMs, Fresh: true,
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
	// ★P1④:human-presence 入决策代价。房内 ≥2 present track = 有他人可救 = 降险 → C_FN↓ → τ*↑ → 容抑制;
	// 独处(≤1)=基线最高 fire-leaning。「一切看风险」从代价涌现(非硬规则)。
	tauCtx := belief.TauContext{
		Bathroom:      roomType == card.RoomTypeBathroom,
		Night:         night,
		OthersPresent: len(bases) >= 2,
	}
	tauCtxDec, tauCtxHit := belief.DecideTauCtx(pFallen, tauCtx)
	// ★档 2 firmware-veto 裁决(option A):用**当帧** bases 现算的 co-existence/ghost/τ* 裁决暂存的 firmware fall,
	// 不用 cache-time 旧值(消除"事件 vs tick"竞态)。孤立(coExist=false)→ 必发(铁律:孤立 track 永不判 ghost)。
	if len(sh.pendingFwFalls) > 0 {
		e.resolvePendingFirmwareFalls(sh, tm, roomID, coExist, ghostTracks, pFallen, tauCtxHit, nowMs)
	}
	e.logger.Debug("belief_shadow_trace",
		zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
		zap.Float64("p_fallen", pFallen),
		zap.String("argmax_state", argTraceS.String()), zap.Float64("argmax_p", argTraceP),
		zap.String("last_lost_area", belief.AreaTypeLabel(sh.lastLostArea)),
		zap.String("p7_1_tau_decision", tauDec.String()), zap.Float64("p7_1_tau", tauHit),
		zap.Bool("p7_2_bathroom", tauCtx.Bathroom), zap.Bool("p7_2_night", tauCtx.Night),
		zap.String("p7_2_tau_decision", tauCtxDec.String()), zap.Float64("p7_2_tau", tauCtxHit))
	// ── Tsensor debug:每 tick 无条件 dump belief 全向量 + reason 到 Tsensor.log(分析原因用)──
	{
		belMap := make(map[string]float64, 9)
		for s := belief.SEmpty; s <= belief.SLeft; s++ {
			belMap[s.String()] = v.P(s)
		}
		obsKinds := make([]string, 0, len(obs))
		for _, o := range obs {
			obsKinds = append(obsKinds, o.Kind.String())
		}
		frReason, frDom, frLR := belief.FallReasonFor(obs)
		e.logger.Info("tsensor_belief",
			zap.String("room_id", roomID), zap.Int64("ts_ms", nowMs),
			zap.String("ts_human", time.UnixMilli(nowMs).Format("15:04:05.000")),
			zap.Float64("p_fallen", pFallen),
			zap.String("argmax", argTraceS.String()), zap.Float64("argmax_p", argTraceP),
			zap.Any("belief", belMap),
			zap.Strings("obs", obsKinds),
			zap.String("fall_reason", frReason.String()),
			zap.String("dominant_obs", frDom.String()), zap.Float64("dominant_lr", frLR),
			zap.Float64("tau", tauCtxHit), zap.Int("track_count", len(bases)),
		)
	}
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
			zap.String("last_lost_area", belief.AreaTypeLabel(sh.lastLostArea)), // #3：InBed/InToilet=床/桶区误确认嫌疑
			zap.String("p7_1_tau_decision", tauDec.String()),       // P7.1：base confirm（τ*=0.55，等价历史 θ_fire）
			zap.Float64("p7_1_tau_confirm", belief.TauConfirm.Tau()),
			zap.Bool("p7_2_bathroom", tauCtx.Bathroom), zap.Bool("p7_2_night", tauCtx.Night), // P7.2：context
			zap.String("p7_2_tau_decision", tauCtxDec.String()), zap.Float64("p7_2_tau", tauCtxHit),
			zap.String("p7_3_reason", p7Reason.String()), // P7.3：主导节点成因
			zap.String("p7_3_dominant_obs", p7Dom.String()), zap.Float64("p7_3_dominant_lr", p7LR),
			zap.Bool("p7_4_human_bed_veto", p7HumanBedVeto), // P7.4：Conf≥99 人工床 → 生产前置短路（R0 只读出）
			zap.Int("p7_4_min_conf", humanBedExemptMinConfidence),
		)
		// ★cutover 三档:档 1/2(dbnSelfFireEnabled)→ DBN 自发推断 fall。保守 ghost veto + 风险分层 τ* + 默认放行。
		// 档 0 = DBN 不自发(只当 firmware 过滤器,见 stashPendingFirmwareFall/resolvePendingFirmwareFalls)。
		// 工单5:per-unit effectiveMode（cold unit cap=1 仍允许自发，cd2b 这类 firmware 漏判靠它补）。
		if e.dbnSelfFireEnabledFor(unitSuiteID, nowMs) {
			fb, ok := dbnFallerBase(bases, sh)
			// ★★ghost 判据 = co-existence（用户铁律,安全包线）：ghost 是真人的反射 → 必有共存的真人 partner
			// → 画面里 **≥2 track**;**孤立 1 track = 没有被反射的源 = 不可能是 ghost = 一定是真人 → 永远发,绝不否**
			// (护住 long-lie 躺地不动的真受害者——他越不动越像 frozen,但孤立就证明他是真人)。**绝不**用 realLO/
			// present-realness 单 track 判 ghost(那把躺地真受害者误判=灾难)。只 **≥2 track 且 faller 是 VerdictGhost**
			// (gate-list 镜像/运动对称,触发即需另一 Real partner)才否。endgame:DBN 自己的 frozenGhost/realLO 是
			// 单 track 病根,待重做成 co-existence(找 Real partner),不再依赖 gate-list VerdictGhost。
			// ★P1-final 完成:DBN 自有 co-existence 对称(motion/mirror),不再读 b.Verdict。coExist 用循环前算的房级值。
			dbnGhost := coExist && (dbnMotionSymmetryGhost(&fb, bases, prevPos) || dbnMirrorSymmetryGhost(&fb, bases, interferes))
			switch {
			case ok && dbnGhost:
				e.logger.Info("ghost_veto", zap.String("reason", "dbn_coexist"), zap.String("room_id", roomID), // co-existence ghost 抑制
					zap.Float64("p_fallen", pFallen), zap.String("p7_3_reason", p7Reason.String()), zap.Int("track_count", len(bases)))
			case ok && pFallen < tauCtxHit:
				// ★P1④ 风险分层(从代价涌现):有他人在场 → C_FN↓ → τ*↑(tauCtxHit)。p_fallen<τ* 的 marginal fall 抑制。
				// 独处 τ* 基线低 → p_fallen≥τ*(已 confirmed)→ 不触此 = **必发**。委员会细化2「降险不归零」:τ*>0 恒。
				e.logger.Info("ghost_veto", zap.String("reason", "risk"), zap.String("room_id", roomID),
					zap.Float64("p_fallen", pFallen), zap.Float64("tau_ctx", tauCtxHit), zap.Bool("others_present", tauCtx.OthersPresent))
			case ok:
				// P3 moving_fall tag:pose_lying 主导 + 摔者摔前在动(lastMove 近)= 移动中突变倒地 → ReasonMoving;
				// 否则开阔地静躺=ReasonPoseLying。Room 层按 motion context **赋值** reason(诚实分层),tag 统一 "dbn_"+
				// reason.String()(词表单源 fall_reason.go,消 #1.1/#1.3 inline 字面量)。DBN 输出全 4-tag。
				var lastMove int64
				if tlf := sh.tlayer[fb.TrackID]; tlf != nil {
					lastMove = tlf.lastMoveMs
				}
				reasonTag := "dbn_" + dbnMovingReason(p7Reason, lastMove, nowMs).String()
				incidentMs := nowMs
				if p7Dom == belief.ObsDwellStill {
					for _, o := range obs {
						if o.Kind == belief.ObsDwellStill && o.Value > 0 {
							incidentMs = nowMs - int64(o.Value*1000)
							break
						}
					}
				}
				e.PublishAIAlarm(context.Background(), AIPayload{
					DeviceAddr: fb.DeviceAddr, RoomID: roomID,
					Track: observation.Track{
						BedStatus: observation.BedStatusUnchanged, TrackID: fb.TrackID, Pose: fb.Pose,
						PositionX: intPtr(fb.RawH), PositionY: intPtr(fb.RawV), PositionZ: intPtr(fb.RawZ),
					},
					Reason:     reasonTag,
					Evidence:   map[string]interface{}{"p_fallen": pFallen, "dominant_obs": p7Dom.String(), "bathroom": tauCtx.Bathroom},
					IncidentMs: incidentMs,
				}, alarm.Fall, nowMs)
				e.logger.Info("belief_dbn_fire", zap.String("room_id", roomID), // R0 在此结束:DBN 真发告警
					zap.String("reason", reasonTag), zap.Float64("p_fallen", pFallen), zap.Int("track_id", fb.TrackID))
			}
		}
	}
	sh.fired = confirmed

	// ★P2 recovery-veto 检测(委员会 approve,DBN_VETO_RECOVERY 子开关):firmware 火后,**摔者本人**
	// (firstSeen≤T_fire,护工后进=新 track 排除)持续直立(≥3s,正向 up 非 track-presence)= 纯误火恢复。
	// 但火后持续倒地≥15s=**自救型真摔**→禁(不静默抹,呼应 silent_leftbed)。track-lost(无正向 up)→无 recovery=安全。
	if sh.firmwareFallTs != 0 && !sh.recoveryEmitted {
		for _, tl := range sh.tlayer { // 自救真摔判别:火后(T_fire 后)持续倒地≥阈
			if tl.firstSeenMs > sh.firmwareFallTs || tl.fallenSince == 0 {
				continue
			}
			postStart := tl.fallenSince
			if sh.firmwareFallTs > postStart {
				postStart = sh.firmwareFallTs
			}
			if nowMs-postStart >= recoveryGenuineFallenMs {
				sh.recoveryGenuineFall = true
			}
		}
		if !sh.recoveryGenuineFall {
			for tid, tl := range sh.tlayer {
				if tl.firstSeenMs > sh.firmwareFallTs || tl.uprightSince <= sh.firmwareFallTs {
					continue // 护工后进排除;直立须火后起(摔后起身)
				}
				if nowMs-tl.uprightSince < recoveryUprightSustainMs {
					continue // 持续非瞬时
				}
				e.logger.Info("belief_dbn_recovery_evidence", // R0 log;ON 时窗内抑制/auto-resolve(留 cutover)
					zap.String("room_id", roomID), zap.Int("track_id", tid), zap.Int64("ts_ms", nowMs),
					zap.String("logic_id", tl.logicID), zap.Int64("post_fire_ms", nowMs-sh.firmwareFallTs),
					zap.Bool("would_veto", dbnVetoRecoveryEnabled))
				sh.recoveryEmitted = true
				break
			}
		}
	}
}

// recordBeliefShadowFirmwareFall P2:engine 收 firmware Fall(category=Fall)时喂 T_fire 进 shadow(recovery 参考)。
func (e *Engine) recordBeliefShadowFirmwareFall(roomID string, tMs int64) {
	if !beliefShadowEnabled || tMs <= 0 {
		return
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)
	if tMs > sh.firmwareFallTs {
		sh.firmwareFallTs = tMs
		sh.recoveryGenuineFall = false
		sh.recoveryEmitted = false
	}
}

// dbnFwPendingTimeoutMs 暂存的 firmware fall 最久等待:超此仍没被 tick 裁决(雷达彻底静默,无帧无 88 心跳→无
// beliefShadowTick)→ firmwarePendingDrainLoop **clock-based** force-forward(fail-safe 偏发不偏漏,不靠新事件触发)。
// 雷达活着时每条 radar message 都跑 beliefShadowTick,正常 ≤1 帧即裁决,远不到此窗。
const dbnFwPendingTimeoutMs = 5_000

// dbnFwPendingDrainInterval firmwarePendingDrainLoop 扫描间隔(<timeout,使滞留 fall 在 ~timeout+interval 内必送出)。
const dbnFwPendingDrainInterval = 2 * time.Second

// stashPendingFirmwareFall 档 2:firmware fall 到达**不立即裁决**,仅暂存——待下一 beliefShadowTick 用 fresh bases
// 现算 co-existence 裁决(resolvePendingFirmwareFalls),或雷达静默时由 firmwarePendingDrainLoop 超时 force-forward。
// 档 1 不走此路(firmware 地板立即发)。
func (e *Engine) stashPendingFirmwareFall(roomID string, a RadarFallAlarm) {
	if !beliefShadowEnabled {
		return
	}
	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	sh := e.beliefShadowFor(roomID)
	// arrivedMs 用**服务器墙钟**(非 a.TMs firmware event_since):drain 的超时按 time.Now() 比,
	// 必须同钟,否则 firmware 时钟偏移会让 age 算错(慢→立即误触发 / 快→永不触发)。
	sh.pendingFwFalls = append(sh.pendingFwFalls, pendingFwFall{a: a, arrivedMs: time.Now().UnixMilli()})
}

// firmwarePendingDrainLoop clock-based fail-safe:雷达静默(无帧→无 beliefShadowTick→pending 永不裁决)时,
// 周期 force-forward 超时滞留的 firmware fall。闭合"暂存后雷达彻底失声→真摔静默丢"的漏报缺口(委员会终审 blocker)。
func (e *Engine) firmwarePendingDrainLoop(ctx context.Context) {
	if !beliefShadowEnabled {
		return
	}
	ticker := time.NewTicker(dbnFwPendingDrainInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer func() { // 安全网:drain 内"不可能"的 nil-panic 响亮记录,不静默、不崩 goroutine(rule 1.4)
					if r := recover(); r != nil {
						e.logger.Error("firmware_pending_drain_panic", zap.Any("recover", r))
					}
				}()
				e.drainStalePendingFirmwareFalls(time.Now().UnixMilli())
			}()
		}
	}
}

// drainStalePendingFirmwareFalls force-forward 所有超时(雷达静默未被 tick 裁决)的 pending firmware fall。
// 先在 e.mu 下快照 rooms(避免持 beliefShadowMu 再取 e.rooms 造成 e.mu↔beliefShadowMu 反向锁),
// 再在 beliefShadowMu 下转发(锁序 beliefShadowMu→tm.mu,与 beliefShadowTick 一致)。
func (e *Engine) drainStalePendingFirmwareFalls(nowMs int64) {
	e.mu.RLock()
	rooms := make(map[string]*TrackManager, len(e.rooms))
	for id, tm := range e.rooms {
		rooms[id] = tm
	}
	e.mu.RUnlock()

	e.beliefShadowMu.Lock()
	defer e.beliefShadowMu.Unlock()
	for roomID, sh := range e.beliefShadows {
		if len(sh.pendingFwFalls) == 0 {
			continue
		}
		// tm 必非 nil:房间从不注销(全仓无 delete(e.rooms)),beliefShadow 只为已注册房懒建。
		// 与 resolvePendingFirmwareFalls 一致信任非 nil;真 nil=wiring bug → forwardFirmwareFall 内 panic 暴露
		// (上面 drain loop 的 recover 兜成响亮日志,不静默丢 rule 1.4)。
		tm := rooms[roomID]
		kept := sh.pendingFwFalls[:0]
		for _, p := range sh.pendingFwFalls {
			if nowMs-p.arrivedMs <= dbnFwPendingTimeoutMs {
				kept = append(kept, p)
				continue
			}
			e.forwardFirmwareFall(sh, tm, p.a)
			e.logger.Info("belief_dbn_firmware_forward", zap.String("room_id", roomID),
				zap.Int("track_id", p.a.TrackID), zap.String("reason", "radar_silent_drain"),
				zap.Int64("pending_age_ms", nowMs-p.arrivedMs), zap.Int64("ts_ms", nowMs))
		}
		sh.pendingFwFalls = kept
	}
}

// forwardFirmwareFall 放行一条 firmware fall:转发 cardagg + 喂 T_fire(recovery 参考),二者**原子耦合**——
// 只在真转发后更新 T_fire(否则 recovery-veto 会基于一个从没发出去的 fall 误判)。调用方持 beliefShadowMu。
// tm 必非 nil(调用点保证;rule 1.4:不写静默 nil 兜底——若真 nil 则 panic 暴露 wiring bug,beliefShadowTick recover 兜住)。
// 直接改 sh.firmwareFallTs(已持锁,不走 recordBeliefShadowFirmwareFall 以免重入 beliefShadowMu 死锁)。
func (e *Engine) forwardFirmwareFall(sh *beliefShadow, tm *TrackManager, a RadarFallAlarm) {
	tm.RecordRadarAlarm(a)
	if a.Category == alarm.Fall && a.TMs > sh.firmwareFallTs {
		sh.firmwareFallTs = a.TMs
		sh.recoveryGenuineFall = false
		sh.recoveryEmitted = false
	}
}

// resolvePendingFirmwareFalls 在 beliefShadowTick 末用**当帧** co-existence/ghost/τ* 裁决暂存的 firmware fall。
// 调用方持 beliefShadowMu(且 ghostTracks/coExist/pFallen/tauCtxHit 均为本帧现算值)。
// 孤立(coExist=false)→ 必发(铁律:孤立 track 永不判 ghost);共存成立 → 命中 ghost(faller 与 Real partner
// 运动/镜像对称)或 risk(room P(Fallen)<context τ*,DBN 不佐证)→ 否决,否则放行。
func (e *Engine) resolvePendingFirmwareFalls(sh *beliefShadow, tm *TrackManager, roomID string,
	coExist bool, ghostTracks map[int]bool, pFallen, tauCtxHit float64, nowMs int64) {
	for _, p := range sh.pendingFwFalls {
		switch {
		case coExist && ghostTracks[p.a.TrackID]:
			e.logger.Info("belief_dbn_veto_firmware", zap.String("room_id", roomID),
				zap.Int("track_id", p.a.TrackID), zap.String("veto_reason", "ghost"),
				zap.Int("track_count", len(ghostTracks)), zap.Int64("ts_ms", nowMs))
		case coExist && pFallen < tauCtxHit:
			e.logger.Info("belief_dbn_veto_firmware", zap.String("room_id", roomID),
				zap.Int("track_id", p.a.TrackID), zap.String("veto_reason", "risk"),
				zap.Float64("p_fallen", pFallen), zap.Float64("tau_ctx", tauCtxHit), zap.Int64("ts_ms", nowMs))
		default:
			e.forwardFirmwareFall(sh, tm, p.a)
			e.logger.Info("belief_dbn_firmware_forward", zap.String("room_id", roomID),
				zap.Int("track_id", p.a.TrackID), zap.Bool("co_exist", coExist), zap.Int64("ts_ms", nowMs))
		}
	}
	sh.pendingFwFalls = nil
}

// dbnMirrorSymmetryGhost P1-final 增量2:DBN **自有** mirror 静态对称 ghost(绕开 gate-list b.Verdict)。复刻
// tm.checkMirrorSymmetry:self 关于某反射面 m 的镜像位 (rx,ry),若有共存 VerdictReal partner 距镜像位 ≤50cm →
// self 是该真人的镜面反射 → ghost。catch 静态冻结 ghost(cd2b 类,motion 对称漏的)。复用 reflectAcrossMirror(lock-free)。
func dbnMirrorSymmetryGhost(self *TrackStatusBase, bases []TrackStatusBase, interferes []radarutils.Rect) bool {
	const mirrorDistCm = 50
	for _, m := range interferes {
		rx, ry := reflectAcrossMirror(self.X, self.Y, m)
		for i := range bases {
			p := &bases[i]
			if p.TrackID == self.TrackID || p.Verdict != VerdictReal {
				continue
			}
			if distInt(p.X, p.Y, rx, ry) <= mirrorDistCm {
				return true // self 在某 Real partner 的镜像位 → self 是 mirror ghost
			}
		}
	}
	return false
}

// dbnMovingReason P3:pose_lying 主导时,摔者摔前 movingFallRecentMs 内在动 → ReasonMoving(移动中突变倒地);
// 否则保持(开阔地静躺=ReasonPoseLying / lost / silent 不变)。Room 层按 motion context 赋值,词表单源(fall_reason.go)。
func dbnMovingReason(base belief.FallReason, lastMoveMs, nowMs int64) belief.FallReason {
	if base == belief.ReasonPoseLying && lastMoveMs > 0 && nowMs-lastMoveMs < movingFallRecentMs {
		return belief.ReasonMoving
	}
	return base
}

// dbnMotionSymmetryGhost P1-final:DBN **自有** motion 对称 ghost(lock-free,绕开 gate-list b.Verdict)。
// self 与某共存 **VerdictReal + 在动** track 紧贴(<100cm)+ 同向运动(cos>0.866) = 多径反射跟随真人 → ghost。
// 复刻 tm.checkMotionSymmetry 但用 bases(安全快照)+prevPos(循环前上帧位置),不碰 tm 锁。需 self 也在动(静止
// 走 mirror 路,本函数只 motion;静态 mirror 对称留下增量需 interference 几何)。
func dbnMotionSymmetryGhost(self *TrackStatusBase, bases []TrackStatusBase, prevPos map[int][2]int) bool {
	const (
		coexistDistMaxCm = 100
		minDispSqCm      = 100 // 单 track 最小位移²(<10cm=静止 noise)
		cosThreshold     = 0.866
	)
	if !self.MoveActive {
		return false
	}
	pS, okS := prevPos[self.TrackID]
	if !okS {
		return false
	}
	dxS, dyS := self.X-pS[0], self.Y-pS[1]
	if dxS*dxS+dyS*dyS < minDispSqCm {
		return false
	}
	for i := range bases {
		p := &bases[i]
		if p.TrackID == self.TrackID || p.Verdict != VerdictReal || !p.MoveActive {
			continue
		}
		if distInt(self.X, self.Y, p.X, p.Y) > coexistDistMaxCm {
			continue
		}
		pp, okP := prevPos[p.TrackID]
		if !okP {
			continue
		}
		dxP, dyP := p.X-pp[0], p.Y-pp[1]
		magSq := float64(dxP*dxP + dyP*dyP)
		if magSq < float64(minDispSqCm) {
			continue
		}
		dot := float64(dxS*dxP + dyS*dyP)
		mags := math.Sqrt(float64(dxS*dxS+dyS*dyS)) * math.Sqrt(magSq)
		if mags > 0 && dot/mags > cosThreshold {
			return true // 紧贴 + 同向 = motion 对称 ghost(DBN 自算,不依赖 gate-list)
		}
	}
	return false
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
