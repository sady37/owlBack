package roomengine

import (
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"owl-common/observation"
	"owl-common/roomutil"
)

// TrackOutput Room Engine 对外输出的单条 track 评估结果
type TrackOutput struct {
	TrackID  int
	DeviceID string
	RoomID   string
	Verdict  TrackVerdict
	Score    int // [0,100]
	Risk     int // [0,100]
	Anomaly  Anomaly
	X, Y, Z  int // 当前估计位置（画布坐标 cm）
	VX, VY   int // cm/s
	StillSec int
	Source   string // "radar_direct" / "engine_silent"
}

// TrackFrame 一帧输入（已由 engine 层做完 RadarToCanvas 转换）
type TrackFrame struct {
	TrackID  int
	DeviceID string
	X, Y, Z  int // 画布坐标 cm
	Pose     int
	AreaType int // 雷达给的 area_id（保留兼容字段，engine 不信其判定，用 cell.AreaType）
	TMs      int64

	// 用于 frozen-frame 检测的辅助字段（每个值是 firmware 给的原始信号）
	// firmware 失锁后这些字段会保持 byte-equal 5+ 分钟，是判 frozen 的强证据
	TrackConfidence int // 0-100
	RemainingTime   int
}

// TrackManager 管理一个房间内所有 track 的生命周期
type TrackManager struct {
	mu      sync.Mutex
	roomID  string
	grid    *RoomGrid
	tracks  map[int]*TrackState
	outputs map[int]*TrackOutput

	// pendingSilentFalls：消失后挂起 60 秒等待复现的 track 池
	// key = 原 trackID（仅做唯一标识；匹配靠位置距离）
	pendingSilentFalls map[int]*PendingSilentFall

	// pendingLostFalls：lost-fall 规则的挂起池。
	// 时长按消失点 cell areaType（5min walkway / 60min bed / ...），含 frozen credit；
	// 取消条件：新 track 出生 / ExitRoom 事件 / room.NumberPeople ≥ 2。
	pendingLostFalls map[int]*PendingLostFall

	// bathroomRealCount：当前帧 ProcessFrame 起点时，所在 cell 为 AreaToilet/AreaShower 的 Real track 数。
	// 用途：≥2 时视为护工陪同，scoreMovement 跳过 long-still 15min 超时报警。
	// 由 ProcessFrame 入口刷新，scoreMovement 读取（同步在锁内，无竞态）。
	bathroomRealCount int

	// sleepadStates：同房间 sleepad 设备的最新观测（device_uid → obs）
	// 由 ProcessSleepadObservation 写入；silent fall 触发前查 sleepadInBed 做 short-circuit
	sleepadStates map[string]*SleepadObservation

	// bedPersonCount：每张床（device_uid → 人数）当前在床人数计数。
	// 由 ProcessSleepadBedEvent 增减：InBed +1，LeftBed -1（floor 0）。
	// 用途：bed-fall 触发前判断"仅 1 人"避免家属/护工陪同时误报。
	bedPersonCount map[string]int

	// 累计计数（仅供 playback / 监控读取，不改变行为）
	pendingCreatedCount   int // 进入 pending 的总数（消失且通过几何检查）
	pendingCancelledCount int // 被新 track 出生取消的数量
	silentFallReportCount int // 60s 超时真报 silent fall 的数量

	// Lost fall 统计
	lostFallPendingCreated   int
	lostFallPendingCancelled int // 含 birth-recovery / ExitRoom / NumberPeople 三类取消
	lostFallReported         int

	sleepadInBedCount int

	// moveSpeedCms：Kalman 速度阈值（cm/s）。> 此速度的帧即使 pose 不是 Walking 也算 Move。
	// 设计动机：雷达对老人慢走常报 Standing → ActiveType[Move]/TraverseCount 永远不涨。
	// 默认 20（≈2 cells/s）；由 engine.Configure / playback.Run 从 yaml 注入。
	moveSpeedCms int

	// lastLeftBedAt：任意来源（sleepad event / sleepad bed_status 转换 / 未来 radar event）
	// 最近一次 LeftBed 事件的时间戳。R4（AnomalyBedsideFall）开窗判定用。
	// 0 = 从未有 LeftBed 事件（或上次 InBed 后又被 InBed 抹掉）。
	lastLeftBedAt int64

	// bedsideFallCfg：R4（床边晕倒）参数；由 SetBedsideFallConfig 注入。
	// 全 0 = 用默认（180s / 100cm / 900s）。
	bedsideFallCfg BedsideFallConfig

	// recentRadarAlarms / recentRadarEvents：来自 iot:alarm:stream / iot:event:stream 的 radar 来源记录。
	// 仅"落账"，当前无消费方；未来段 7 (radar fall verify) 会读取做 narrative。
	// 保留窗口 = recentBufferMs（默认 5 min），由 RecordXxx 顺手 evict。
	recentRadarAlarms map[int64]*RadarFallAlarm  // key = TMs
	recentRadarEvents map[int64]*RadarTrackEvent // key = TMs
	recentBufferMs    int64                      // 默认 5 min

	// logger：用于 ai.log 输出 ghost / fall 结构化事件。
	// 默认 zap.NewNop()，engine.Run 会调 SetLogger 注入真 logger。
	logger *zap.Logger

	// timezone：本房间 unit 的 IANA 时区（如 America/Denver），由 engine.RegisterRoom 注入。
	// IsNightTime 调用时传入；nil 时退化为 UTC（错位风险，bootstrap 必须设置）。
	timezone *time.Location

	// roomName：rooms.room_name，由 engine.RegisterRoom 注入。
	// Still fall 触发时与 cell.Belief[0].Type 取并集判 bathroom 语义（见 owl-common/roomutil.ClassifyRoomType）。
	roomName string
}

// BedsideFallConfig R4 床边晕倒参数。
// 物理含义：风险时段（IsNightTime）内 LeftBed 之后 WindowSec 秒内，
// 若 track 距 AreaBed cell ≤ BedsideMarginCm 且静止 > StillTimeoutSec → AnomalyBedsideFall。
type BedsideFallConfig struct {
	WindowSec       int // LeftBed 后多少秒内是开窗期
	BedsideMarginCm int // 距 AreaBed 此值内视为"床边"
	StillTimeoutSec int // 静止此秒数触发
}

// 默认值（与 config.go 中 setRoomEngineDefaults 一致；零值兜底用）
var defaultBedsideFallCfg = BedsideFallConfig{
	WindowSec:       180,
	BedsideMarginCm: 100,
	StillTimeoutSec: 900,
}

// NewTrackManager 创建 track 管理器
func NewTrackManager(roomID string, grid *RoomGrid) *TrackManager {
	return &TrackManager{
		roomID:             roomID,
		grid:               grid,
		tracks:             make(map[int]*TrackState),
		outputs:            make(map[int]*TrackOutput),
		pendingSilentFalls: make(map[int]*PendingSilentFall),
		pendingLostFalls:   make(map[int]*PendingLostFall),
		sleepadStates:      make(map[string]*SleepadObservation),
		bedPersonCount:     make(map[string]int),
		moveSpeedCms:       20, // 默认值（与 DefaultLearnParams.MoveSpeedCms 一致）
		bedsideFallCfg:     defaultBedsideFallCfg,
		recentRadarAlarms:  make(map[int64]*RadarFallAlarm),
		recentRadarEvents:  make(map[int64]*RadarTrackEvent),
		recentBufferMs:     5 * 60 * 1000, // 5 min
		logger:             zap.NewNop(),
	}
}

// SetLogger 注入 zap logger（engine.Run 启动时调用）。
// nil 输入会被替换为 NopLogger（防止后续 nil deref）。
func (tm *TrackManager) SetLogger(l *zap.Logger) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if l == nil {
		tm.logger = zap.NewNop()
	} else {
		tm.logger = l
	}
}

// SetTimezone 注入本房间所在 unit 的 IANA 时区。
// 由 engine.RegisterRoom 调用；nil 表示未配置（IsNightTime 会退化为 UTC）。
func (tm *TrackManager) SetTimezone(loc *time.Location) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.timezone = loc
}

// SetRoomName 注入 rooms.room_name；用于 still fall 触发时按房间名判 bathroom 语义。
// 由 engine.RegisterRoom 调用。
func (tm *TrackManager) SetRoomName(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.roomName = name
}

// IsBathroomByRoomName 用 owl-common/roomutil.ClassifyRoomType 判定本房间是否 bathroom。
// 与 cell.Belief[0].Type ∈ {AreaToilet, AreaShower} 取并集驱动 still fall。
func (tm *TrackManager) IsBathroomByRoomName() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return roomutil.IsBathroom(tm.roomName)
}

// ProcessSleepadBedEvent 接收 sleepad InBed/LeftBed 事件，更新人数计数。
// 由 Engine 路由 iot:event:stream 中 device_type=Sleepad 时调用。
func (tm *TrackManager) ProcessSleepadBedEvent(evt SleepadBedEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if evt.IsInBed {
		tm.bedPersonCount[evt.DeviceUID]++
	} else {
		c := tm.bedPersonCount[evt.DeviceUID] - 1
		if c < 0 {
			c = 0
		}
		tm.bedPersonCount[evt.DeviceUID] = c
		// LeftBed 事件 → R4 开窗。任一来源（sleepad event / 状态机转换）触发即更新。
		if evt.TMs > tm.lastLeftBedAt {
			tm.lastLeftBedAt = evt.TMs
		}
	}
}

// totalBedPeople 同房间所有 sleepad 床的总人数（多床房间累加）
func (tm *TrackManager) totalBedPeople() int {
	n := 0
	for _, c := range tm.bedPersonCount {
		n += c
	}
	return n
}

// ProcessSleepadObservation 接收 sleepad 一帧观测，按设备 UID 保留最新状态。
// 由 Engine.handleMessage 路由 device_type=Sleepad 时调用；
// silent fall 报警前会查询此状态做 short-circuit（"sleepad 确认在床有 vital 即不报"）。
func (tm *TrackManager) ProcessSleepadObservation(obs SleepadObservation) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cur, ok := tm.sleepadStates[obs.DeviceUID]
	if !ok || obs.TMs > cur.TMs {
		// 状态转换 InBed=true → InBed=false 视为 LeftBed（与 ProcessSleepadBedEvent 等价）
		// 单独的 event 流不一定到（部分固件只发 monitor），所以这里也要打开 R4 窗口。
		if ok && cur.InBed && !obs.InBed && obs.TMs > tm.lastLeftBedAt {
			tm.lastLeftBedAt = obs.TMs
		}
		copyObs := obs
		tm.sleepadStates[obs.DeviceUID] = &copyObs
	}
}

// sleepadInBed 检查同房间任一 sleepad 在 30s 内报告 InBed（不要求 HR/RR）。
// 设计动机：用户明确"确认在床 = 雷达坐标在床 + sleepad InBed，不要求 HR/RR"
//   原因：人坐床上 sleepad HR/RR 信号可能弱；只要 InBed 就是床压传感器侧的存在性证据。
// 30s 阈值：sleepad 数据偶有延迟，太老（>30s）就不可信。
func (tm *TrackManager) sleepadInBed(nowMs int64) bool {
	const maxStaleMs = 30_000
	for _, s := range tm.sleepadStates {
		if nowMs-s.TMs > maxStaleMs {
			continue
		}
		if s.InBed {
			return true
		}
	}
	return false
}

// sleepadOffBed 检查同房间任一 sleepad 在 30s 内**有数据**且**显示离床**。
// 与 sleepadInBed 不同：sleepadInBed 看到 InBed=true 即返回；
// sleepadOffBed 要求至少有一条新鲜数据，且没有任何在床信号。
// 用于 bed-fall（雷达坐标在床 + sleepad 离床）检测。
func (tm *TrackManager) sleepadOffBed(nowMs int64) bool {
	const maxStaleMs = 30_000
	hasFresh := false
	for _, s := range tm.sleepadStates {
		if nowMs-s.TMs > maxStaleMs {
			continue
		}
		hasFresh = true
		if s.InBed {
			return false // 任一 sleepad 在床即否决"全离床"
		}
	}
	return hasFresh
}

// SetMoveSpeedCms 注入"在动"速度阈值。<=0 保留默认。
func (tm *TrackManager) SetMoveSpeedCms(v int) {
	if v <= 0 {
		return
	}
	tm.mu.Lock()
	tm.moveSpeedCms = v
	tm.mu.Unlock()
}

// SetBedsideFallConfig 注入 R4 床边晕倒参数；任一字段 0 保留默认值。
func (tm *TrackManager) SetBedsideFallConfig(c BedsideFallConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if c.WindowSec > 0 {
		tm.bedsideFallCfg.WindowSec = c.WindowSec
	}
	if c.BedsideMarginCm > 0 {
		tm.bedsideFallCfg.BedsideMarginCm = c.BedsideMarginCm
	}
	if c.StillTimeoutSec > 0 {
		tm.bedsideFallCfg.StillTimeoutSec = c.StillTimeoutSec
	}
}

// RecordRadarAlarm 落账 radar 来源的 alarm（当前阶段仅 Fall）。
// 仅写入 recentRadarAlarms + 顺手 evict 老条目，不做任何 verify / 抑制。
// 调用方（engine.handleAlarmMessage）应当紧跟 tm.Tick(alarm.TMs) 触发段 4-6 立即跑一次。
func (tm *TrackManager) RecordRadarAlarm(a RadarFallAlarm) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := a
	tm.recentRadarAlarms[a.TMs] = &cp
	tm.evictOldRadarAlarms(a.TMs)
}

// RecordRadarEvent 落账 radar 来源的事件（EnterRoom/ExitRoom/InBed/LeftBed）。
// 仅落账。InBed/LeftBed 也"复用 sleepad 通道"更新 lastLeftBedAt（R4 开窗 + 床压计数）：
//   - LeftBed → 更新 tm.lastLeftBedAt（任意来源都开 R4 窗口）
//
// 注：radar 的 InBed/LeftBed 不更新 bedPersonCount——bedPersonCount 是 sleepad 床压传感器累计的"在床人数"，
// radar event 是空间检测，两者语义不同；混用会导致 bed-fall 段 5 的"仅 1 人"判定错乱。
func (tm *TrackManager) RecordRadarEvent(e RadarTrackEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := e
	tm.recentRadarEvents[e.TMs] = &cp
	tm.evictOldRadarEvents(e.TMs)
	if e.EventName == "LeftBed" && e.TMs > tm.lastLeftBedAt {
		tm.lastLeftBedAt = e.TMs
	}
	// ExitRoom 事件 → 取消所有挂起的 lost-fall（人正常走出房间，不再悬念）
	// 注：silent fall 不取消（其语义是床上方遮挡，与 ExitRoom 无关）
	if e.EventName == "ExitRoom" && len(tm.pendingLostFalls) > 0 {
		for pid, p := range tm.pendingLostFalls {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_exit_room",
				zap.String("device_uid", p.DeviceID),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("pending_age_ms", e.TMs-p.DisappearMs),
				zap.Int64("exit_room_ms", e.TMs),
			)
			delete(tm.pendingLostFalls, pid)
		}
	}
}

// evictOldRadarAlarms / evictOldRadarEvents：删除超出 recentBufferMs 的旧记录。
// 调用方持锁。
func (tm *TrackManager) evictOldRadarAlarms(nowMs int64) {
	cutoff := nowMs - tm.recentBufferMs
	for k := range tm.recentRadarAlarms {
		if k < cutoff {
			delete(tm.recentRadarAlarms, k)
		}
	}
}

func (tm *TrackManager) evictOldRadarEvents(nowMs int64) {
	cutoff := nowMs - tm.recentBufferMs
	for k := range tm.recentRadarEvents {
		if k < cutoff {
			delete(tm.recentRadarEvents, k)
		}
	}
}

// Tick 不带新 frame 的扫描入口，用 ts 作为 nowMs 推进段 2-6。
// 用途：alarm/event 到达时立即触发，让 silent fall pending 检查 / bed-fall 矛盾检测 / R4
// 都用 alarm 的时间戳重新评估（比等下一帧 monitor 更及时，最差能省 0-1s）。
//
// 返回当前所有 track 的输出快照（同 ProcessFrame 语义）。
func (tm *TrackManager) Tick(ts int64) []TrackOutput {
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	return tm.processFrameAt(nil, ts)
}

// SetSleepadInBedCount 外部更新 Sleepad 在床人数
func (tm *TrackManager) SetSleepadInBedCount(count int) {
	tm.mu.Lock()
	tm.sleepadInBedCount = count
	tm.mu.Unlock()
}

// ========================================================================
// ProcessFrame：每帧双维度喂（即时流 + 历史流）
// ========================================================================

func (tm *TrackManager) ProcessFrame(frames []TrackFrame) []TrackOutput {
	nowMs := time.Now().UnixMilli()
	if len(frames) > 0 {
		nowMs = frames[0].TMs
	}
	return tm.processFrameAt(frames, nowMs)
}

// processFrameAt 内部入口，nowMs 显式传入，**测试可直接控制时间推进**。
// 无 frames 也可调用（推进 pending 超时检查 + 段 5 bed-fall + 段 6 输出）。
func (tm *TrackManager) processFrameAt(frames []TrackFrame, nowMs int64) []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	activeIDs := make(map[int]bool)

	// 入口先盘点 bathroom 内 real track 数（caregiver 例外用）
	tm.bathroomRealCount = 0
	for _, t := range tm.tracks {
		if t.Verdict != VerdictReal {
			continue
		}
		pxF, pyF := t.Kalman.Position()
		c := tm.grid.CellAt(int(math.Round(pxF)), int(math.Round(pyF)))
		if c == nil {
			continue
		}
		if c.Belief[0].Type == AreaToilet || c.Belief[0].Type == AreaShower {
			tm.bathroomRealCount++
		}
	}

	// ========== 段 1: 观测到的 track ==========
	for _, f := range frames {
		activeIDs[f.TrackID] = true
		ts, exists := tm.tracks[f.TrackID]

		var quality, vx, vy, dtSec int

		if !exists {
			// 新 track 出生 — 取消挂起的 silent fall 候选（occlusion 复现保护）
			// 匹配条件：100cm 内 + 新 track 姿态非 Lie（如果出生姿就是 Lie，
			// 可能是被遮挡后真倒下，保留挂起继续等 60s 超时）
			tm.cancelPendingByBirth(f.X, f.Y, RadarPoseToCore(f.Pose))

			// 是否有 pending lost-fall 在等候 → 人从盲区返回；取消 + 学习盲区出口
			recoveredFromLost := tm.cancelPendingLostFallByBirth(f.X, f.Y, f.TMs)

			ts = NewTrackState(f.TrackID, f.DeviceID, tm.roomID, f.X, f.Y, f.Z, f.TMs)
			b := tm.birthScore(f.X, f.Y, f.TMs)
			ts.BirthScore = b.score
			ts.BirthReason = b.reason
			ts.Score = ts.BirthScore
			// 盲区返回路径：直接给 Real verdict，绕过 ghost 检查
			// （这是漏报场景下人从盲区返回的预期行为：track 看似凭空出现但实为真人）
			if recoveredFromLost {
				ts.Verdict = VerdictReal
				ts.Score = ScoreConfirmTh
				ts.BirthScore = ScoreConfirmTh
				ts.BirthReason = "recovered_from_lost_fall"
				ts.ConfirmedAtMs = f.TMs
			}
			// 初始化 frozen 检测签名：把出生帧记为第一个签名（FrozenSameCount=1）
			ts.LastFrameSig = frameSigOf(f)
			ts.FrozenSameCount = 1
			tm.tracks[f.TrackID] = ts
			ts.PrevCore = RadarPoseToCore(f.Pose)
			quality = ts.Score
			dtSec = 1
		} else {
			// 已有 track
			dt := float64(f.TMs-ts.LastUpdateMs) / 1000.0
			if dt <= 0 {
				dt = 1
			}
			dtSec = int(math.Round(dt))
			if dtSec < 1 {
				dtSec = 1
			}

			ts.Kalman.Predict(dt)
			residualF := ts.Kalman.Update(float64(f.X), float64(f.Y))
			ts.PushPoint(f.X, f.Y, f.Z, f.TMs)
			residual := int(math.Round(residualF))

			// 连续指标（frozen + Kalman birth-coherence），在 Kalman update 之后维护
			tm.updateContinuousIndicators(ts, f, nowMs, residualF)

			// 维度 A: 即时流
			tm.scoreResidual(ts, residual)
			tm.scoreMovement(ts, f.X, f.Y, nowMs)
			tm.detectZNoise(ts, f.Z)
			tm.detectPoseMismatch(ts, f.Pose)
			tm.updateLieStateMachine(ts, f.Pose, f.X, f.Y, f.Z, nowMs)

			quality = ts.Score
			vxF, vyF := ts.Kalman.Velocity()
			vx = int(math.Round(vxF))
			vy = int(math.Round(vyF))

			ts.LastPose = f.Pose
			ts.LastZ = f.Z
		}

		// 维度 B: 历史流（每帧无条件）
		tm.grid.MarkOccupancy(f.X, f.Y, quality, vx, vy, nowMs)
		core := RadarPoseToCore(f.Pose)
		// Speed 兜底：雷达对慢走老人常报 Standing → ActiveType[Move] 永远不涨。
		// 仅对 Stand / Unknown 做升格（这两个本来就是"静态站立或不确定"）；
		// Sit / Lie 不动 —— 坐着上半身晃动 / 床上翻身也可能让 Kalman 速度 > 阈值，强升 Move 会污染 Sit/Bed 学习。
		// 阈值由 SetMoveSpeedCms 注入，默认 20 cm/s。
		if core == CorePoseStand || core == CorePoseUnknown {
			speed := math.Hypot(float64(vx), float64(vy))
			if speed > float64(tm.moveSpeedCms) {
				core = CorePoseMove
			}
		}
		tm.grid.MarkPoseTime(f.X, f.Y, core, dtSec, nowMs)

		// Walk 区学习：core==Move 且进入新 cell 时 ++ TraverseCount
		curCol, curRow := tm.grid.ToIndex(f.X, f.Y)
		if core == CorePoseMove && (curCol != ts.LastCellCol || curRow != ts.LastCellRow) {
			tm.grid.MarkTraverse(f.X, f.Y, nowMs)
		}
		ts.LastCellCol = curCol
		ts.LastCellRow = curRow
	}

	// ========== 段 2: 未观测到的 track ==========
	for id, ts := range tm.tracks {
		if activeIDs[id] {
			continue
		}
		dt := float64(nowMs-ts.LastUpdateMs) / 1000.0
		if dt <= 0 {
			dt = 1
		}
		ts.Kalman.PredictOnly(dt)
		ts.LastUpdateMs = nowMs

		if ts.Kalman.MissCount > MaxMissCount {
			// 消失判定：silent fall 优先（60s + sleepad 兜底，bedroom 专用）
			// 不满足 silent → 检 lost fall（按 cell areaType 分时长，全屋通用，verdict 未定也算）
			if ts.Verdict == VerdictReal && tm.checkSilentFall(ts) {
				// 不立即报；挂起 60 秒等复现窗口（segment 1 取消 / segment 5 超时报）
				pxF, pyF := ts.Kalman.Position()
				tm.pendingSilentFalls[id] = &PendingSilentFall{
					OriginalTrackID: id,
					DeviceID:        ts.DeviceID,
					RoomID:          ts.RoomID,
					LastX:           int(math.Round(pxF)),
					LastY:           int(math.Round(pyF)),
					LastZ:           ts.LastZ,
					LastScore:       ts.Score,
					LastVerdict:     ts.Verdict,
					DisappearMs:     nowMs,
				}
				tm.pendingCreatedCount++
			} else if (ts.Verdict == VerdictReal || ts.Verdict == VerdictPending) && tm.checkLostFall(ts) {
				pxF, pyF := ts.Kalman.Position()
				px := int(math.Round(pxF))
				py := int(math.Round(pyF))
				cell := tm.grid.CellAt(px, py)
				cellArea := AreaUnknown
				if cell != nil {
					cellArea = cell.Belief[0].Type
				}
				tm.pendingLostFalls[id] = &PendingLostFall{
					OriginalTrackID: id,
					DeviceID:        ts.DeviceID,
					RoomID:          ts.RoomID,
					LastX:           px,
					LastY:           py,
					LastZ:           ts.LastZ,
					LastScore:       ts.Score,
					LastVerdict:     ts.Verdict,
					LastCellArea:    cellArea,
					DisappearMs:     nowMs,
					FrozenStartMs:   ts.FrozenRunStart,
					SpatialJump:     ts.MaxImpliedSpeedFromBirth > FallRulesParam.Lost.SuspectSpeedCm,
				}
				tm.lostFallPendingCreated++
			} else if ts.Verdict == VerdictReal {
				pxF, pyF := ts.Kalman.Position()
				px := int(math.Round(pxF))
				py := int(math.Round(pyF))
				if !tm.grid.IsEdge(px, py, 50) {
					ts.CurrentAnomaly = AnomalyPathBreak
				}
			}
			delete(tm.tracks, id)
			continue
		}
	}

	// ========== 段 3: 试用期判定 ==========
	for _, ts := range tm.tracks {
		if ts.Verdict != VerdictPending {
			continue
		}
		// Birth grace recompute（方案 A）：deadline 到了再扫扩展窗，给 EnterRoom event-stream 缓冲
		tm.tryGraceUpgrade(ts, nowMs)

		if ts.FrameCount >= ProbationFrames {
			if ts.Score >= ScoreConfirmTh {
				ts.Verdict = VerdictReal
				ts.ConfirmedAtMs = nowMs
			} else if ts.Score < ScoreGhostTh {
				ts.Verdict = VerdictGhost
				if !ts.LoggedGhost {
					reason := ts.BirthReason
					if reason == "" {
						reason = "low_score"
					}
					pxF, pyF := ts.Kalman.Position()
					tm.logger.Info("track_verdict_ghost",
						zap.String("device_uid", ts.DeviceID),
						zap.Int("track_id", ts.TrackID),
						zap.String("verdict", "ghost"),
						zap.Int("score", ts.Score),
						zap.Int("birth_score", ts.BirthScore),
						zap.String("reason", reason),
						zap.Int("x", int(math.Round(pxF))),
						zap.Int("y", int(math.Round(pyF))),
						zap.Int64("ts_ms", nowMs),
					)
					ts.LoggedGhost = true
				}
			}
		}
	}

	// ========== 段 4: 扫挂起的 silent fall，超时即报 ==========
	// 60 秒内没被 segment 1 取消（人没复现）→ 真报 silent fall
	results := make([]TrackOutput, 0, len(tm.tracks)+len(tm.pendingSilentFalls))
	for pid, p := range tm.pendingSilentFalls {
		if nowMs-p.DisappearMs < PendingSilentFallMs {
			continue
		}
		// 多传感器融合 short-circuit：同房间 sleepad 30s 内确认 InBed（不要求 HR/RR）
		// → 床压传感器证明人还在床上，雷达消失只是失锁；不报 silent fall
		// （即使消失位置 cell 不是 AreaBed —— 人可能在床边 cell，雷达坐标轻微偏离床区
		// 但 sleepad 床压是物理硬证据，sleepad 在床即在床）
		if tm.sleepadInBed(nowMs) {
			tm.pendingCancelledCount++
			delete(tm.pendingSilentFalls, pid)
			continue
		}
		// 超时 → MarkFallEvent + 写一条输出
		tm.silentFallReportCount++
		tm.grid.MarkFallEvent(p.LastX, p.LastY, nowMs)
		tm.logger.Info("real_fall",
			zap.String("device_uid", p.DeviceID),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("kind", "engine_silent_fall"),
			zap.Int("score", p.LastScore),
			zap.Int("risk", 100),
			zap.String("reason", "track_disappeared_in_non_rest_zone_60s"),
			zap.Int("x", p.LastX), zap.Int("y", p.LastY), zap.Int("z", p.LastZ),
			zap.Int64("ts_ms", nowMs),
		)
		out := TrackOutput{
			TrackID:  p.OriginalTrackID,
			DeviceID: p.DeviceID,
			RoomID:   p.RoomID,
			Verdict:  p.LastVerdict,
			Score:    p.LastScore,
			Risk:     100, // silent fall 最高风险
			Anomaly:  AnomalyFall,
			X:        p.LastX,
			Y:        p.LastY,
			Z:        p.LastZ,
			Source:   "engine_silent",
		}
		results = append(results, out)
		tm.outputs[p.OriginalTrackID] = &out
		delete(tm.pendingSilentFalls, pid)
	}

	// ========== 段 4b: 扫挂起的 lost fall，按 cell-area-typed wait + frozen credit 超时即报 ==========
	// 取消条件已在他处处理（cancelPendingByBirth / handleExitRoom / numberPeopleCancel）。
	// 此处仅扫超时触发：等待时间到达且未被取消 → 报 lost fall。
	isRiskTime := IsNightTime(nowMs, tm.timezone)
	for pid, p := range tm.pendingLostFalls {
		waitMs := tm.lostFallWaitMs(p, isRiskTime)
		if nowMs-p.DisappearMs < waitMs {
			continue
		}
		// 实时再检多人入屋（与 birth/event 触发的取消互补；保 segment 4b 兜底）
		if tm.realTrackCount() >= 2 {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_multiple_real",
				zap.String("device_uid", p.DeviceID),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("nowMs", nowMs),
			)
			delete(tm.pendingLostFalls, pid)
			continue
		}
		// 超时 → MarkFallEvent + 写输出（kind=engine_lost_fall）
		tm.lostFallReported++
		tm.grid.MarkFallEvent(p.LastX, p.LastY, nowMs)
		tm.logger.Info("real_fall",
			zap.String("device_uid", p.DeviceID),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("kind", "engine_lost_fall"),
			zap.Int("score", p.LastScore),
			zap.Int("risk", 100),
			zap.String("reason", "track_lost_no_exit_room_no_recovery"),
			zap.Int("cell_area_type", int(p.LastCellArea)),
			zap.Int64("frozen_start_ms", p.FrozenStartMs),
			zap.Bool("spatial_jump", p.SpatialJump),
			zap.Int64("wait_ms", waitMs),
			zap.Int("x", p.LastX), zap.Int("y", p.LastY), zap.Int("z", p.LastZ),
			zap.Int64("ts_ms", nowMs),
		)
		out := TrackOutput{
			TrackID:  p.OriginalTrackID,
			DeviceID: p.DeviceID,
			RoomID:   p.RoomID,
			Verdict:  p.LastVerdict,
			Score:    p.LastScore,
			Risk:     100,
			Anomaly:  AnomalyFall,
			X:        p.LastX,
			Y:        p.LastY,
			Z:        p.LastZ,
			Source:   "engine_lost",
		}
		results = append(results, out)
		tm.outputs[p.OriginalTrackID] = &out
		delete(tm.pendingLostFalls, pid)
	}

	// ========== 段 5: Bed-Fall 物理矛盾检测 ==========
	// 物理意义：人从床上跌落到地面/床边，但雷达因坐标精度仍认为人在床
	// 触发条件：
	//   1. 雷达仍 track 着，坐标在 AreaBed cell
	//   2. sleepad 30s 内有数据且全部显示离床
	//   3. 房间总人数 == 1（避免家属/护工陪同时误报）
	// 房间总人数 = max(realTrackCount, totalBedPeople)
	//   - radar 多 track 即多人（雷达本来就追多人）
	//   - sleepad 床上有 ≥2 人即多人（连续 InBed 事件累计）
	if tm.sleepadOffBed(nowMs) {
		realCount := 0
		var soleReal *TrackState
		for _, ts := range tm.tracks {
			if ts.Verdict == VerdictReal {
				realCount++
				soleReal = ts
			}
		}
		bedPeople := tm.totalBedPeople()
		totalPeople := realCount
		if bedPeople > totalPeople {
			totalPeople = bedPeople
		}
		if totalPeople == 1 && realCount == 1 && soleReal != nil {
			pxF, pyF := soleReal.Kalman.Position()
			px, py := int(math.Round(pxF)), int(math.Round(pyF))
			cell := tm.grid.CellAt(px, py)
			if cell != nil && cell.Belief[0].Type == AreaBed {
				// 矛盾确认：雷达说在床 + sleepad 说离床 + 房间仅 1 人 → bed-fall
				if soleReal.CurrentAnomaly != AnomalyBedFall {
					soleReal.CurrentAnomaly = AnomalyBedFall
					tm.grid.MarkFallEvent(px, py, nowMs)
					tm.logger.Info("real_fall",
						zap.String("device_uid", soleReal.DeviceID),
						zap.Int("track_id", soleReal.TrackID),
						zap.String("kind", "engine_bed_fall"),
						zap.Int("score", soleReal.Score),
						zap.Int("risk", 100),
						zap.String("reason", "radar_in_bed_cell_but_sleepad_off_bed_solo"),
						zap.Int("x", px), zap.Int("y", py),
						zap.Int64("ts_ms", nowMs),
					)
				}
			}
		}
	}

	// ========== 段 6: 构建输出 ==========
	for _, ts := range tm.tracks {
		pxF, pyF := ts.Kalman.Position()
		vxF, vyF := ts.Kalman.Velocity()
		px := int(math.Round(pxF))
		py := int(math.Round(pyF))
		vx := int(math.Round(vxF))
		vy := int(math.Round(vyF))

		stillSec := 0
		if ts.StillSince > 0 {
			stillSec = int((nowMs - ts.StillSince) / 1000)
		}

		source := "radar_direct"
		if ts.SilentFall {
			source = "engine_silent"
		}

		out := TrackOutput{
			TrackID:  ts.TrackID,
			DeviceID: ts.DeviceID,
			RoomID:   ts.RoomID,
			Verdict:  ts.Verdict,
			Score:    ts.Score,
			Risk:     tm.computeRisk(ts, stillSec, nowMs),
			Anomaly:  ts.CurrentAnomaly,
			X:        px,
			Y:        py,
			Z:        ts.LastZ,
			VX:       vx,
			VY:       vy,
			StillSec: stillSec,
			Source:   source,
		}
		results = append(results, out)
		tm.outputs[ts.TrackID] = &out
	}

	return results
}

// GetOutputs 返回当前所有 track 的最新输出
func (tm *TrackManager) GetOutputs() []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	out := make([]TrackOutput, 0, len(tm.outputs))
	for _, o := range tm.outputs {
		out = append(out, *o)
	}
	return out
}

// cancelPendingByBirth 新 track 出生时尝试取消挂起的 silent fall 候选。
// 取消条件：100cm 内 + 新 track 姿态非 Lie（occlusion 复现 = 人没事）。
// 出生姿态就是 Lie 时不取消，因为可能是被遮挡后真倒下了，仍需 60s 超时机制把关。
func (tm *TrackManager) cancelPendingByBirth(x, y int, birthCore CorePose) {
	if birthCore == CorePoseLie {
		return
	}
	for pid, p := range tm.pendingSilentFalls {
		if distInt(x, y, p.LastX, p.LastY) < PendingMatchDistCm {
			delete(tm.pendingSilentFalls, pid)
			tm.pendingCancelledCount++
		}
	}
}

// hasPendingLostFallInRoom 当前房间是否有挂起的 lost-fall（与 RoomID 等价于 tm.roomID）。
// 用于 birth verdict bypass：人从盲区返回时新 track 不该被判 ghost。
// 调用方持锁。
func (tm *TrackManager) hasPendingLostFallInRoom() bool {
	return len(tm.pendingLostFalls) > 0
}

// cancelPendingLostFallByBirth 新 track 出生时尝试取消挂起的 lost-fall。
//
// 与 silent fall cancel 区别：lost-fall 不限位置 + 不限 birth pose（人从盲区任何角度返回都算）。
// 命中时：① 取消 pending ② cell.IncrBlindSpotRecovery（学习盲区出口） ③ 累计 cancelled 统计。
//
// 返回是否命中（调用方据此决定是否给新 track verdict 直接 Real / 加分）。
func (tm *TrackManager) cancelPendingLostFallByBirth(birthX, birthY int, nowMs int64) bool {
	if len(tm.pendingLostFalls) == 0 {
		return false
	}
	// 单房间一般只有 1 个 pending；遍历全部，全部取消（人回来 = 所有 lost 候选都解除）
	hit := false
	for pid, p := range tm.pendingLostFalls {
		hit = true
		tm.lostFallPendingCancelled++
		tm.logger.Info("lost_fall_cancelled_by_recovery",
			zap.String("device_uid", p.DeviceID),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("room_id", p.RoomID),
			zap.Int64("pending_age_ms", nowMs-p.DisappearMs),
			zap.Int("recovered_at_x", birthX), zap.Int("recovered_at_y", birthY),
		)
		delete(tm.pendingLostFalls, pid)
	}
	if hit {
		// 学习盲区出口：人在新 track 出生位置 cell 累计一个 BlindSpotRecovery
		tm.grid.MarkBlindSpotRecovery(birthX, birthY, nowMs)
	}
	return hit
}

// SilentFallStats 给 playback / 监控用的统计快照
type SilentFallStats struct {
	PendingCreated   int // 进入挂起的总数
	PendingCancelled int // 60s 内被取消的数量（occlusion 复现）
	Reported         int // 真报的 silent fall 数量
	Outstanding      int // 当前仍在挂起池中
}

// SilentFallStats 返回累计统计（无锁——调用方在 playback 单线程场景；prod 路径如需可加锁）
func (tm *TrackManager) SilentFallStatsSnapshot() SilentFallStats {
	return SilentFallStats{
		PendingCreated:   tm.pendingCreatedCount,
		PendingCancelled: tm.pendingCancelledCount,
		Reported:         tm.silentFallReportCount,
		Outstanding:      len(tm.pendingSilentFalls),
	}
}

// LostFallStats lost-fall 统计快照（含三类 cancel 来源汇总）
type LostFallStats struct {
	PendingCreated   int // 进入挂起的总数
	PendingCancelled int // 取消累计（含 birth-recovery / ExitRoom / 多人入屋）
	Reported         int // 超时真报的 lost fall 数量
	Outstanding      int // 当前仍在挂起池中
}

// LostFallStatsSnapshot 返回 lost-fall 累计统计
func (tm *TrackManager) LostFallStatsSnapshot() LostFallStats {
	return LostFallStats{
		PendingCreated:   tm.lostFallPendingCreated,
		PendingCancelled: tm.lostFallPendingCancelled,
		Reported:         tm.lostFallReported,
		Outstanding:      len(tm.pendingLostFalls),
	}
}

// ========================================================================
// 出生打分（基于 elder care 步速物理常识）
// ========================================================================
//
// 物理判据（用户 2026-04-26 对齐）：
//   - 老人步速上限 1.0 m/s，青壮年 1.5 m/s
//   - 出生位置距 Enter > 150cm = 1 秒走不到 = 物理不可能新人 → 直接 Ghost
//   - 出生位置距 Enter ≤ 150cm 但没有近 3s 内的 EnterRoom 配对事件 → 也是 Ghost
//     （radar firmware 的 enter2out 事件没触发 = 没经过门 = 不是真人入场）
//   - 出生在镜子/家具 (AreaDeny) cell → 物理不可能 → Ghost
//
// 双因素：地理（dEntry） + 时间（EnterRoom 配对）
//
// 注：本函数返回初始 score。score < ScoreGhostTh(20) → 5 帧后判 Ghost；score >= ScoreConfirmTh(50) → 真人。
// 没采用"birth 时直接 setVerdict=Ghost"是为了保留试用期内 EnterRoom 事件迟到时的纠正机会。

const (
	enterPairWindowMs   = 3_000 // EnterRoom 与 birth 的最大时间差（ms）
	birthMaxRealisticCm = 150   // 距 Enter 此值内才可能是 1 秒走入的真人（青壮年 1.5m/s 上限）
	birthEnterPairBonus = 20    // 有 EnterRoom 配对加分（不让分数超过 100 边界）
)

// birthScoreResult birthScore 计算结果（score + 短路原因，调用方写入 ts.BirthReason 用于 ai.log）
type birthScoreResult struct {
	score  int
	reason string // 空 = 正常打分；非空 = ghost 短路原因（"far_from_enter" / "no_enter_pair"）
}

func (tm *TrackManager) birthScore(x, y int, tMs int64) birthScoreResult {
	score := 50
	reason := ""

	// 因子 1: d_enter — elder care 物理上限（青壮年 1.5 m/s, 老人 1.0 m/s）
	dEntry := tm.grid.NearestEntryDist(x, y)
	switch {
	case dEntry < 50:
		score += 20
	case dEntry <= birthMaxRealisticCm:
		if !tm.hasRecentEnterRoom(tMs) {
			return birthScoreResult{0, "no_enter_pair"} // 50-150cm 但无 EnterRoom 配对
		}
		score += birthEnterPairBonus
	default:
		return birthScoreResult{0, "far_from_enter"} // > 150cm 物理不可能
	}

	// 因子 2: 出生 cell 的 GhostRatio + AreaType
	cell := tm.grid.CellAt(x, y)
	if cell != nil {
		ghostR := cell.GhostRatio()
		if ghostR > 0.7 {
			score -= 25
		} else if ghostR > 0.4 {
			score -= 10
		}
		if cell.IsEntry() {
			score += 15
		}
		if cell.Belief[0].Type == AreaDeny {
			score -= 30 // 出生在 Deny 区直接重扣
			reason = "born_in_deny"
		}
	}

	// 因子 3: 房间已有 track 数
	nExisting := len(tm.tracks)
	if nExisting == 0 {
		score += 10
	} else if nExisting >= 2 {
		score -= 10
	}

	// 因子 4: 在已有 track 的同时，出生在"从未学到任何活动语义"的 cell（AreaUnknown）
	// AreaType 已经过 LearnCellAreas 推断 + Decay 衰减，反映"近期活动 + 远期遗忘"。
	// AreaUnknown 在 prod 长期运行后只剩"真没有人去过的角落"——多 track 共存时凭空出生几乎一定 ghost。
	// 注：playback 短窗测试 grid 全 Unknown 时此规则会过度触发，prod 用 Persister hydrate 解决。
	if cell != nil && cell.Belief[0].Type == AreaUnknown && nExisting >= 1 {
		score -= 30
		if reason == "" {
			reason = "born_in_unknown_area_with_other_track"
		}
	}

	return birthScoreResult{clampInt(score, 0, 100), reason}
}

// hasRecentEnterRoom 检查 [tMs - enterPairWindowMs, tMs] 窗口内有无 EnterRoom 事件。
// 调用方持锁（segment 1 已持锁）。
//
// 注：仅 look-back（不 look-forward）—— 出生瞬时如 event-stream 还没到 → 判错 false ghost。
// 修正路径：BirthFinalDeadlineMs 给 grace 缓冲（FallRulesParam.Lost.BirthFinalGraceMs，默认 2s），
// 到点用 hasRecentEnterRoomBetween([T-3s, deadline]) 重检（见 tryGraceUpgrade）。
func (tm *TrackManager) hasRecentEnterRoom(tMs int64) bool {
	return tm.hasRecentEnterRoomBetween(tMs-enterPairWindowMs, tMs)
}

// hasRecentEnterRoomBetween 在显式时间窗 [fromMs, toMs] 内查 EnterRoom 事件。
func (tm *TrackManager) hasRecentEnterRoomBetween(fromMs, toMs int64) bool {
	for k, e := range tm.recentRadarEvents {
		if k < fromMs || k > toMs {
			continue
		}
		if e.EventName == "EnterRoom" {
			return true
		}
	}
	return false
}

// tryGraceUpgrade Birth verdict 多流时序兜底（方案 A）。
//
// 出生时 EnterRoom event-stream 可能落后于 monitor stream 1-3s，birth score 短路为 0。
// BirthFinalDeadlineMs 给 grace 缓冲；到点用扩展窗 [birth-3s, deadline] 再扫，
// 如有 EnterRoom 命中 + 出生位置物理可达（≤150cm to entry）→ 抬 birth_score。
//
// 一次性 recompute（不论结果如何，之后清空 deadline 不再重算）。
// 调用位置：segment 3 promote 之前；调用方持锁。
func (tm *TrackManager) tryGraceUpgrade(ts *TrackState, nowMs int64) {
	if ts.BirthFinalDeadlineMs == 0 {
		return // 已 finalize
	}
	if nowMs < ts.BirthFinalDeadlineMs {
		return // 还在 grace 期内
	}
	defer func() { ts.BirthFinalDeadlineMs = 0 }() // 不论结果，之后不再重算

	// 仅当 birth 是因「找不到 EnterRoom」而被短路时才补救
	if ts.BirthReason != "no_enter_pair" && ts.BirthReason != "far_from_enter" {
		tm.logger.Debug("birth_grace_skip_reason",
			zap.Int("track_id", ts.TrackID),
			zap.String("birth_reason", ts.BirthReason))
		return
	}
	// 距离判断：>150cm 物理仍不可能
	dEntry := tm.grid.NearestEntryDist(ts.BirthPos.X, ts.BirthPos.Y)
	if dEntry > birthMaxRealisticCm {
		tm.logger.Info("birth_grace_skip_far_from_enter",
			zap.Int("track_id", ts.TrackID),
			zap.Int("d_entry", dEntry),
			zap.Int("birth_x", ts.BirthPos.X), zap.Int("birth_y", ts.BirthPos.Y))
		return
	}
	// 扩展窗 [birth-3s, deadline] 重扫
	if !tm.hasRecentEnterRoomBetween(ts.BirthPos.TMs-enterPairWindowMs, ts.BirthFinalDeadlineMs) {
		tm.logger.Info("birth_grace_skip_no_enter_event_in_window",
			zap.Int("track_id", ts.TrackID),
			zap.Int64("birth_ms", ts.BirthPos.TMs),
			zap.Int64("deadline", ts.BirthFinalDeadlineMs),
			zap.Int("event_count", len(tm.recentRadarEvents)))
		return
	}
	// 命中：抬 birth score
	newScore := 50 + birthEnterPairBonus // base + entry bonus
	if dEntry < 50 {
		newScore = 50 + 20 // 与 birthScore 因子 1 dEntry<50 加分对齐
	}
	if newScore > ts.BirthScore {
		delta := newScore - ts.BirthScore
		ts.BirthScore = newScore
		ts.BirthReason = "grace_enter_pair_recovered"
		ts.Score = clampInt(ts.Score+delta, 0, 100)
		tm.logger.Info("birth_grace_upgraded",
			zap.String("device_uid", ts.DeviceID),
			zap.Int("track_id", ts.TrackID),
			zap.Int("new_birth_score", ts.BirthScore),
			zap.Int("d_entry_cm", dEntry),
			zap.Int64("birth_ms", ts.BirthPos.TMs),
			zap.Int64("nowMs", nowMs),
		)
	}
}

// ========================================================================
// 即时流打分（不累计到 cell，只在 track 内部）
// ========================================================================

// scoreResidual Kalman 残差打分
func (tm *TrackManager) scoreResidual(ts *TrackState, residual int) {
	switch {
	case residual < 30:
		ts.AdjustScore(2)
	case residual < 80:
		// 中等偏差
	case residual < 200:
		ts.AdjustScore(-5)
	default:
		ts.AdjustScore(-20) // 空间跳跃
	}
}

// scoreMovement 运动/静止评分 + 累计静止时长 + Dwell 记录
func (tm *TrackManager) scoreMovement(ts *TrackState, x, y int, nowMs int64) {
	if len(ts.History) < 2 {
		return
	}
	prev := ts.History[len(ts.History)-2]
	d := distInt(prev.X, prev.Y, x, y)

	cell := tm.grid.CellAt(x, y)
	isRest := cell != nil && cell.IsRestZone()

	if d < StillThreshCm {
		// 静止
		if ts.StillSince == 0 {
			ts.StillSince = nowMs
			ts.StillX = x
			ts.StillY = y
		}
		if !isRest && ts.Verdict == VerdictPending {
			ts.AdjustScore(-3)
		}
		// 静止超时（综合 cell history 的自适应阈值）
		if cell != nil {
			isRiskTime := IsNightTime(nowMs, tm.timezone)
			timeout := cell.EffectiveStillTimeoutSec(isRiskTime)
			if timeout > 0 {
				// Bathroom caregiver 例外：本 cell 在 toilet/shower + ≥2 real track 同在 bathroom
				// → 第二个 track 大概率是护工陪同，长时间静止合理（如老人坐马桶、护工旁边照看）
				inBathroom := cell.Belief[0].Type == AreaToilet || cell.Belief[0].Type == AreaShower
				skipTimeout := inBathroom && tm.bathroomRealCount >= 2
				if !skipTimeout {
					stillSec := int((nowMs - ts.StillSince) / 1000)
					if stillSec > timeout {
						ts.CurrentAnomaly = AnomalyStillTooLong
						if !ts.LongStillReported {
							tm.grid.MarkLongStill(x, y, nowMs)
							ts.LongStillReported = true
						}
					}
				}
			}
		}

		// R4: 床边晕倒（升级 AnomalyStillTooLong → AnomalyBedsideFall）
		// 触发：风险时段 + 最近 WindowSec 内有 LeftBed + 当前位置距 AreaBed ≤ BedsideMarginCm
		//       + 静止 > StillTimeoutSec
		// 物理意义：人离床去卫生间途中在床边滑倒/晕倒，雷达失锁前最后位置仍在床边。
		// 与 R2(BedFall) 区别：R2 是"仍在床矛盾"，R4 是"离床后到不了远方"。
		if tm.lastLeftBedAt > 0 &&
			nowMs-tm.lastLeftBedAt < int64(tm.bedsideFallCfg.WindowSec)*1000 &&
			IsNightTime(nowMs, tm.timezone) &&
			tm.grid.IsNearPriorType(x, y, AreaBed, tm.bedsideFallCfg.BedsideMarginCm) {
			stillSec := int((nowMs - ts.StillSince) / 1000)
			if stillSec > tm.bedsideFallCfg.StillTimeoutSec && ts.CurrentAnomaly != AnomalyBedsideFall {
				ts.CurrentAnomaly = AnomalyBedsideFall
				tm.grid.MarkFallEvent(x, y, nowMs)
				ts.LongStillReported = true // 复用 flag 防 LongStill 重复 mark
				tm.logger.Info("real_fall",
					zap.String("device_uid", ts.DeviceID),
					zap.Int("track_id", ts.TrackID),
					zap.String("kind", "engine_bedside_fall_R4"),
					zap.Int("score", ts.Score),
					zap.Int("risk", 100),
					zap.String("reason", "night_left_bed_then_bedside_still_15min"),
					zap.Int("x", x), zap.Int("y", y),
					zap.Int("still_sec", stillSec),
					zap.Int64("ts_ms", nowMs),
				)
			}
		}
	} else {
		// 在动：刚从静止恢复，记 Dwell EMA
		if ts.StillSince > 0 {
			dwellSec := int((nowMs - ts.StillSince) / 1000)
			if dwellSec > 0 {
				tm.grid.MarkDwell(ts.StillX, ts.StillY, dwellSec, nowMs)
			}
		}
		// Cell history integral：之前曾被系统判为 long-still（LongStillReported=true）但 track 自己走了
		// → 系统判错（容忍证据），喂给该 cell 自动放宽未来阈值
		if ts.LongStillReported {
			tm.grid.MarkToleratedStill(ts.StillX, ts.StillY, nowMs)
		}
		ts.StillSince = 0
		ts.LongStillReported = false
		if ts.CurrentAnomaly == AnomalyStillTooLong {
			ts.CurrentAnomaly = AnomalyNone
		}
		// 速度合理性
		speed := int(math.Round(ts.Kalman.Speed()))
		if speed > 10 && speed < 150 {
			ts.AdjustScore(3)
		} else if speed >= 150 {
			ts.AdjustScore(-2)
		}
	}

	// 因子 5: 平均速度
	age := ts.AgeSec()
	if age > 5 {
		avgSpeed := ts.TotalDisplacement() / age
		if avgSpeed < 2 && !isRest {
			ts.AdjustScore(-2)
		}
	}
}

// detectZNoise Z 突变检测（本 track 内部统计，不累计到 cell）
func (tm *TrackManager) detectZNoise(ts *TrackState, z int) {
	if ts.LastZ == 0 {
		return
	}
	dz := ts.LastZ - z
	if dz < 0 {
		dz = -dz
	}
	if dz > 50 { // Z 单帧突变 > 50cm
		ts.ZNoiseCount++
	}
}

// detectPoseMismatch pose 与运动学矛盾（track 内部累计）
func (tm *TrackManager) detectPoseMismatch(ts *TrackState, pose int) {
	speed := int(math.Round(ts.Kalman.Speed()))
	// pose=Walking 但速度 ≈ 0
	if pose == observation.PoseWalking && speed < 5 && ts.FrameCount > 3 {
		ts.PoseMismatchCount++
		ts.CurrentAnomaly = AnomalyPoseMismatch
		ts.AdjustScore(-3)
	}
	// pose=Standing 但 Z < 50（在地面）
	if pose == observation.PoseStanding && ts.LastZ > 0 && ts.LastZ < 50 {
		ts.PoseMismatchCount++
	}
}

// ========================================================================
// 核心姿态状态机（替代原 Pose 两阶段状态机）
// 只做 Lie 进出检测：Stand/Move → Lie → Stand/Move < 3 秒 = LieRetract
//                    Stand/Move → Lie + Z 骤降 = FallEvent
// ========================================================================

func (tm *TrackManager) updateLieStateMachine(ts *TrackState, pose, x, y, z int, nowMs int64) {
	curCore := RadarPoseToCore(pose)
	prevCore := ts.PrevCore

	// 进入 Lie 态
	if curCore == CorePoseLie && prevCore != CorePoseLie {
		ts.LieEnteredAt = nowMs
		ts.LieEnteredX = x
		ts.LieEnteredY = y

		// 自推 Fall：前姿态是 Stand/Move 且 Z 骤降
		if (prevCore == CorePoseStand || prevCore == CorePoseMove) &&
			ts.LastZ > 50 && z < 20 {
			tm.grid.MarkFallEvent(x, y, nowMs)
			ts.CurrentAnomaly = AnomalyFall
			tm.logger.Info("real_fall",
				zap.String("device_uid", ts.DeviceID),
				zap.Int("track_id", ts.TrackID),
				zap.String("kind", "engine_z_drop"),
				zap.Int("score", ts.Score),
				zap.Int("risk", 100),
				zap.String("reason", "stand_or_move_to_lie_with_z_drop"),
				zap.Int("x", x), zap.Int("y", y), zap.Int("z", z),
				zap.Int("prev_z", ts.LastZ),
				zap.Int64("ts_ms", nowMs),
			)
		}
	}

	// 退出 Lie 态
	if prevCore == CorePoseLie && curCore != CorePoseLie {
		if ts.LieEnteredAt > 0 {
			lieDuration := nowMs - ts.LieEnteredAt
			// 短暂 Lie 后回 Stand/Move → Retract
			if lieDuration < LieRetractMs &&
				(curCore == CorePoseStand || curCore == CorePoseMove) {
				tm.grid.MarkLieRetract(ts.LieEnteredX, ts.LieEnteredY, nowMs)
			}
			ts.LieEnteredAt = 0
		}
	}

	ts.PrevCore = curCore
}

// ========================================================================
// Silent Fall（5 要素）
// ========================================================================

// checkSilentFall 判断 track 消失是否疑似 silent fall（"挂起候选"——还需 60s 复现窗口）
//
// 设计变更（2026-04 与用户对齐）：
//   - 去掉 minZ<20：36% 设备完全不报 Z；间歇性 Z 报告下"持续 N 帧"会过滤真信号
//   - 去掉位移>80：垂直跌倒水平位移本来就接近 0，要求 >80 排除典型"原地倒下"
//   - 保留几何：消失位置不在边缘 + 不在 Enter 区（出门不算跌）
//   - 新加 prev pose：消失前最后一帧 core != Lie（本来就在 Lie 中消失，多半是入睡/正常静止丢失）
//
// 通过此函数仅"挂起"候选；最终是否报 silent fall 由 segment 5 在 60 秒后决定。
//
// 注：silent fall 完整重构（sleepad+radar 融合 + 120s/60s 双窗）见 PR-2c 待做。
// 此函数保留旧语义不动，与 lost fall 互斥（silent 优先在 segment 2 trigger 顺序）。
func (tm *TrackManager) checkSilentFall(ts *TrackState) bool {
	if ts.AgeSec() < 10 {
		return false
	}
	n := len(ts.History)
	if n < 3 {
		return false
	}

	// 消失 cell 不在边缘 + 不在 Enter 区（出门 / 走到房间边缘消失合法）
	pxF, pyF := ts.Kalman.Position()
	px := int(math.Round(pxF))
	py := int(math.Round(pyF))
	cell := tm.grid.CellAt(px, py)
	if cell == nil {
		return false
	}
	if cell.EdgeDist <= 30 {
		return false
	}
	if cell.Belief[0].Type == AreaEnter {
		return false
	}

	// 消失位置已是休息区（床/沙发/马桶 OR 累计有 Sit/Lie 观测）→ 静止丢失合理
	// 用 IsLikelyRestZone 比 IsRestZone 宽松，覆盖"沙发未标 layout + 学习未达 15s 阈值"过渡期。
	// 雷达对该 cell pose 报错（如沙发被认成 Stand）不影响 —— 历史 ActiveType[Sit] 是确凿证据。
	if cell.IsLikelyRestZone() {
		return false
	}

	// 消失前最后一帧不能本来就 Lie（睡着/卧床期间丢失多半是雷达静止失锁，非新跌倒）
	if ts.PrevCore == CorePoseLie {
		return false
	}

	return true
}

// ========================================================================
// Lost Fall（cell-area-typed wait + ExitRoom + NumberPeople 兜底）
// ========================================================================

// checkLostFall 判断 track 消失是否符合 lost-fall 触发条件。
//
// 与 checkSilentFall 区别：
//   - silent 仅 Real verdict + age≥10s + 非 likely-rest-zone（bedroom occluded-bed 专用）
//   - lost  Real/Pending verdict + age≥5s + 离任一出口 > ExitDistMinCm（全屋通用）
//
// silent 已触发的 case 不再走 lost（segment 2 中已 else if 分支互斥）。
func (tm *TrackManager) checkLostFall(ts *TrackState) bool {
	if ts.AgeSec() < 5 {
		return false
	}
	pxF, pyF := ts.Kalman.Position()
	px := int(math.Round(pxF))
	py := int(math.Round(pyF))
	cell := tm.grid.CellAt(px, py)
	if cell == nil {
		return false
	}
	// 出门正常 — 在 Enter 区消失合法
	if cell.Belief[0].Type == AreaEnter {
		return false
	}
	// 离最近门 ≤ ExitDistMinCm 视为可能正常走出（≈1 秒可达）
	if tm.grid.NearestEntryDist(px, py) <= FallRulesParam.Lost.ExitDistMinCm {
		return false
	}
	return true
}

// lostFallWaitMs 计算 lost-fall pending 的等待时长（毫秒）。
//
// 基线按消失点 cell areaType：
//   - AreaBed / AreaSit：60min（睡觉/坐着雷达丢 track 常见）
//   - AreaToilet / AreaShower：与 still fall 同（risk-time / non-risk-time）
//   - AreaDeny / 其它：5min
//
// Frozen credit：firmware 失锁前已 frozen 过 → 半计入等待
// SpatialJump factor：track 表现过空间跳跃 → 等待时间 ×0.5（更敏感）
// 兜底：min EffectiveWaitFloorSec
func (tm *TrackManager) lostFallWaitMs(p *PendingLostFall, isRiskTime bool) int64 {
	pl := FallRulesParam.Lost
	var base int
	switch p.LastCellArea {
	case AreaBed, AreaSit:
		base = pl.RestZoneWaitSec
	case AreaToilet, AreaShower:
		// 与 still fall 同时长
		if isRiskTime {
			base = FallRulesParam.Still.ToiletShowerSec
		} else {
			base = int(float64(FallRulesParam.Still.ToiletShowerSec) * FallRulesParam.Still.NonRiskTimeFactor)
		}
	case AreaDeny:
		base = pl.DenyZoneWaitSec
	default:
		base = pl.WalkwayWaitSec
	}

	// Frozen credit：half of frozen duration counted toward wait
	if p.FrozenStartMs > 0 {
		frozenDurMs := p.DisappearMs - p.FrozenStartMs
		if frozenDurMs > 0 {
			base -= int(frozenDurMs / 2 / 1000)
		}
	}

	// Spatial jump factor
	if p.SpatialJump {
		base = int(float64(base) * pl.SpatialJumpFactor)
	}

	if base < pl.EffectiveWaitFloorSec {
		base = pl.EffectiveWaitFloorSec
	}
	return int64(base) * 1000
}

// ========================================================================
// 综合风险分
// ========================================================================

func (tm *TrackManager) computeRisk(ts *TrackState, stillSec int, nowMs int64) int {
	if ts.Verdict == VerdictGhost {
		return 0
	}
	base := 0
	switch ts.CurrentAnomaly {
	case AnomalyFall:
		base = 100
	case AnomalyBedFall:
		base = 100 // 双源矛盾确认的 bed-fall，最高风险
	case AnomalyBedsideFall:
		base = 100 // R4 床边晕倒（夜间 + LeftBed 后床边静止超时），最高风险
	case AnomalyStillTooLong:
		base = 60
	case AnomalyPathBreak:
		base = 80
	case AnomalyPoseMismatch:
		base = 70
	}
	if base == 0 {
		return 0
	}
	tf := timeFactor(nowMs)
	of := tm.occupancyFactor()
	risk := float64(base) * tf * of
	return clampInt(int(math.Round(risk)), 0, 100)
}

func timeFactor(nowMs int64) float64 {
	hour := time.UnixMilli(nowMs).Hour()
	switch {
	case hour >= 5 && hour < 6:
		return 2.0
	case hour >= 22 || hour < 5:
		return 1.5
	case hour >= 6 && hour < 8:
		return 1.3
	default:
		return 1.0
	}
}

func (tm *TrackManager) occupancyFactor() float64 {
	realCount := 0
	for _, ts := range tm.tracks {
		if ts.Verdict == VerdictReal {
			realCount++
		}
	}
	bedOccupied := tm.sleepadInBedCount > 0
	if realCount <= 1 {
		return 1.5
	}
	if realCount == 2 && bedOccupied {
		return 1.2
	}
	return 1.0
}

// realTrackCount 当前 VerdictReal track 数（用于 lost-fall NumberPeople≥2 取消判定）。
// 调用方持锁（segment 内部）。
func (tm *TrackManager) realTrackCount() int {
	n := 0
	for _, t := range tm.tracks {
		if t.Verdict == VerdictReal {
			n++
		}
	}
	return n
}

// frameSigOf 把一帧抽成 frameSignature（用于 frozen-frame 检测的字面比对）。
func frameSigOf(f TrackFrame) frameSignature {
	return frameSignature{
		TrackID:         f.TrackID,
		X:               f.X,
		Y:               f.Y,
		Z:               f.Z,
		Pose:            f.Pose,
		TrackConfidence: f.TrackConfidence,
		RemainingTime:   f.RemainingTime,
	}
}

// updateContinuousIndicators 每帧维护 frozen-frame 检测 + Kalman birth-coherence 指标。
//
// 1. Frozen 检测：连续 N 帧 (tid, x, y, z, pose, tc, rt) 字面相同 → FrozenRunStart 记录起点。
//    用于 lost-fall pending 计算 frozen credit（半计入等待）。
// 2. MaxKalmanResidual：track 生命周期峰值残差。
// 3. MaxImpliedSpeedFromBirth：max(dist(current, birth) / age) cm/s；
//    > ImpossibleSpeedCm 判硬 ghost；> SuspectSpeedCm + 无 EnterRoom 判软 ghost。
//
// 调用位置：processFrameAt 已有 track 分支，Kalman.Update 之后。
func (tm *TrackManager) updateContinuousIndicators(ts *TrackState, f TrackFrame, nowMs int64, residualF float64) {
	// ---- Frozen 检测 ----
	sig := frameSigOf(f)
	if sig == ts.LastFrameSig {
		ts.FrozenSameCount++
		if ts.FrozenSameCount >= FallRulesParam.Lost.FrozenSameThreshold && ts.FrozenRunStart == 0 {
			// 估算起点：假设 1Hz 帧率，回填到 (threshold-1) 秒前
			ts.FrozenRunStart = nowMs - int64(FallRulesParam.Lost.FrozenSameThreshold-1)*1000
		}
	} else {
		ts.FrozenSameCount = 1
		ts.FrozenRunStart = 0
		ts.LastFrameSig = sig
	}

	// ---- MaxKalmanResidual ----
	if residualF > ts.MaxKalmanResidual {
		ts.MaxKalmanResidual = residualF
	}

	// ---- MaxImpliedSpeedFromBirth ----
	ageMs := nowMs - ts.BirthPos.TMs
	if ageMs >= 1000 {
		distFromBirth := distInt(f.X, f.Y, ts.BirthPos.X, ts.BirthPos.Y)
		// implied = dist / ageSec = dist * 1000 / ageMs (cm/s)
		implied := int(int64(distFromBirth) * 1000 / ageMs)
		if implied > ts.MaxImpliedSpeedFromBirth {
			ts.MaxImpliedSpeedFromBirth = implied
		}
	}
}
