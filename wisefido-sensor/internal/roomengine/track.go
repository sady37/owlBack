package roomengine

import "math"

// TrackVerdict track 判定结果
//
// sensor_v2 §10.1.1：Verdict 是 Layer 1 → Layer 2 契约字段。
// VerdictAnchored 与 VerdictReal 平级（"真人"语义），但 LongSurvival/StartupGrace
// 锚定后不可翻 Ghost；下游可借此区分"可信真人"与"普通真人"。
// 数值非连续是为兼容存量（Ghost=2 已在 v1 全量使用，Anchored 追加在 3 位）；
// 流序列化用 VerdictName 走字符串契约，避免下游耦合 int。
type TrackVerdict uint8

const (
	VerdictPending  TrackVerdict = 0
	VerdictReal     TrackVerdict = 1
	VerdictGhost    TrackVerdict = 2
	VerdictAnchored TrackVerdict = 3
)

// VerdictName 流序列化字符串契约（不暴露 int 给下游）。
func VerdictName(v TrackVerdict) string {
	switch v {
	case VerdictReal:
		return "real"
	case VerdictGhost:
		return "ghost"
	case VerdictAnchored:
		return "anchored"
	}
	return "pending"
}

// Anomaly track 异常类型
type Anomaly uint8

const (
	AnomalyNone         Anomaly = 0
	AnomalyFall         Anomaly = 1 // 跌倒（z 骤降 / Silent Fall）
	AnomalyStillTooLong Anomaly = 2 // 静止超时
	AnomalyPathBreak    Anomaly = 3 // 轨迹断裂（非出口消失）
	AnomalyPoseMismatch Anomaly = 4 // pose 与运动学矛盾
	AnomalyBedFall      Anomaly = 5 // 床下跌倒（雷达坐标仍在床 + sleepad 离床 + 仅 1 人）
	AnomalyBedsideFall  Anomaly = 6 // 床边晕倒（夜间 LeftBed 后人未走开，床边静止超时；R4）
)

// TimedPoint 带时间戳的画布坐标（cm 整数）
type TimedPoint struct {
	X, Y, Z int
	TMs     int64
}

// TrackState 单个 track 的完整生命档案。
// 运动粒子：Kalman + 即时判定 + 状态机转移。
// 不累计空间属性（那些归 Cell）。
type TrackState struct {
	// ---- 身份 ----
	TrackID  int
	DeviceID string
	RoomID   string

	// ---- 出生档案 ----
	BirthPos    TimedPoint
	BirthScore  int    // 5 因子初始分 [0,100]
	BirthReason string // birth filter 短路时的原因（"far_from_enter" / "no_enter_pair" / "unknown_area_multitrack" 等）

	// LoggedGhost：verdict 第一次升 Ghost 时已写过 ai.log，避免后续重复 log
	LoggedGhost bool

	// ---- Kalman（内部 float，不持久化；track 死即销毁）----
	Kalman *KalmanFilter2D

	// ---- 历史窗口（滚动 HistoryLen=30 帧；判定只用近 MotionWindowSec 秒）----
	History    []TimedPoint
	FrameCount int

	// ---- Kalman 打分 ----
	Score         int // [0,100]
	Verdict       TrackVerdict
	ConfirmedAtMs int64

	// ---- Z 抖动判断（本 track 内部，不累计到 cell）----
	ZNoiseCount int // Z 突变次数（单 track 内统计）

	// ---- 运动学矛盾（本 track 内部）----
	PoseMismatchCount int

	// ---- 核心姿态状态机（Retract / Fall 检测用）----
	PrevCore     CorePose
	LieEnteredAt int64 // 进入 Lie 的时间（ms；0 = 未在 Lie 态）
	LieEnteredX  int
	LieEnteredY  int

	// ---- 静止状态机 ----
	StillSince        int64
	StillX, StillY    int
	LongStillReported bool // 防 LongStill 重复上报
	StillFallReported bool // 防 still-fall 重复上报（bathroom + pose=Stand + 15/18min）
	// BedsideFallReported R4 床边晕倒已报过 → lost_fall pending 入池跳过该 track，
	// 防双报（track 仍活时报 bedside_fall，后续 firmware 丢失 track 又触发 lost_fall）
	BedsideFallReported bool

	// ---- PR-7.2 stand-static 自学习 → AreaSit 强化 ----
	// StandStaticSince：pose=Stand 且静止的起点 ms（0 = 不在 stand-static 状态）。
	// AreaSitAutoLearned：track 生命周期内已触发过 AreaSit 自学习（防重复）。
	StandStaticSince   int64
	AreaSitAutoLearned bool

	// ---- PR-11 持续观测刷新（防 Belief 衰退后 layout 标记被吃掉）----
	// LyingOnBedSinceMs：pose=Lie on AreaBed cell 持续起点；累计 ≥4h → MarkRestZoneByFeedback(AreaBed) refresh
	// SitOnToiletSinceMs：pose=Sit on AreaToilet cell 持续起点；累计 ≥5min → MarkRestZoneByFeedback(AreaToilet) refresh
	// AreaBedRefreshed / AreaToiletRefreshed：per-track 一次性 flag（防同 cell 反复触发）
	LyingOnBedSinceMs   int64
	SitOnToiletSinceMs  int64
	AreaBedRefreshed    bool
	AreaToiletRefreshed bool

	// ---- PR-13 region static 自学习 → AreaSit（双 cell 加分 + 90% 容忍累积）----
	// RegionStaticStartedMs：进入 region static（|dx|≤15 AND |dy|≤15）的起点 ms；0 = 不在 region。
	// RegionStaticFrames / RegionTotalFrames：用于计算 ratio = static / total，要求 ≥0.90。
	// RegionStartZ：region 起点的 z 值（A 路径检测 |dz|≥10 用）。
	// AreaSitLearnedRegion：per-track 一次性 flag。
	RegionStaticStartedMs int64
	RegionStaticFrames    int
	RegionTotalFrames     int
	RegionStartZ          int
	AreaSitLearnedRegion  bool

	// ---- 异常 ----
	CurrentAnomaly Anomaly

	// ---- 最后观测 ----
	LastPose     int
	LastZ        int
	// LastUpdateMs：每次 processFrameAt（含 miss tick）都会刷新到 nowMs，反映"engine
	// 最近一次为该 track 计算/预测过的时间"。不能用作"最后真正看到 track 的时间"。
	LastUpdateMs int64
	// LastObservedMs：最近一次真实收到帧（PushPoint）的时间。miss tick 不更新。
	// 用途：track 真实失踪点 ms。PR-9 前曾用于 number_people=0 ExitRoom 兜底比对（已删）。
	// 字段保留供 PR-10 BathroomLostFall + PR-11 bedroom lost_fall 重写时作"失锁时刻"基准。
	LastObservedMs int64

	// ---- cell 穿越追踪（Walk 学习用）----
	// 初始为 -1 表示尚未定位；新 cell 进入且 core==Move 时调 grid.MarkTraverse 计数 ++。
	LastCellCol int
	LastCellRow int

	// ---- Frozen-frame 检测（box 判据，2026-05-03 由 byte-equal 改为 box）----
	// FrozenRunStart：当前 still box run 起点 ms（0 = 未达 still 状态）。
	// 判据：失锁前 30s 滚动窗口内位移 box (max-min) <= StillBoxCm(30) → 视为 still。
	// 抗 X/Y 抖动 / 抗摔倒抽搐（box 容差替代 byte-equal 严格相等）。
	// 用于 lost-fall pending 计算 credit（半计入等待）+ PR-C 流式 cancel 守卫。
	FrozenRunStart int64

	// ---- Birth-coherence Kalman 域（每帧 O(1) 维护）----
	// MaxKalmanResidual：track 生命周期内的峰值残差（Mahalanobis-like）。
	// MaxImpliedSpeedFromBirth：max over life of dist(current, birth) / age（cm/s）。
	// 用途：firmware 复用 track_id 拼接两个不相关反射时（如 D5F7 case），
	// 出生分通过 + 即时 Kalman 残差小，但累计位移除以累计时间隐含速度爆表 → 强 ghost 信号。
	MaxKalmanResidual        float64
	MaxImpliedSpeedFromBirth int

	// ---- Birth verdict 多流时序兜底（方案 A）----
	// BirthFinalDeadlineMs：若 verdict 仍 Pending 且 nowMs >= 此值 → 重算 birth score
	// （把 EnterRoom 反查窗口扩展到 [Birth-3s, deadline]），给 event-stream 缓冲 2s。
	// 0 = 已终判过，不再重算。
	BirthFinalDeadlineMs int64

	// ---- PR-5.x Ghost penalty 累积器（与 BirthScore 互补）----
	// GhostPenalty：track 生命周期累积的 ghost 扣分（>= 0），≥ 80 判 Ghost。
	// 出生时由 birthScore 因子 1/2/4/6 累积；持续期由因子 3/5 增量。
	GhostPenalty int

	// LifetimeFactorsApplied：bitmask，防止 lifetime 因子（如 30s 静止）重复扣。
	// bit 0 = factor 3 (30s static) applied
	LifetimeFactorsApplied uint32

	// LongSurvivalAnchored：track 存活 ≥ 5min 后兜底锚定 Real（不再翻 Ghost）。
	LongSurvivalAnchored bool

	// StartupGrace：service 启动 5min grace 期内 first-seen 的 track，默认 Real。
	StartupGrace bool
}

// Track 生命周期常量
const (
	HistoryLen       = 30   // 滚动窗口帧数
	MotionWindowSec  = 5    // 运动学判定窗口（近 5 秒）
	ProbationFrames  = 5    // 试用期帧数
	ScoreConfirmTh   = 50   // Score ≥ 此值 → Real
	ScoreGhostTh     = 20   // Score < 此值 → Ghost
	StillThreshCm    = 15   // 帧间位移 < 此值视为静止
	MaxMissCount     = 10   // 连续丢失 > 此值 → 消失（约 10 秒）
	LieRetractMs     = 8000 // 进入 Lie 后 < 此时长回到 Stand/Move → Retract
	// 经验值：真跌倒 8 秒内爬起概率 < 5%；雷达固件的 fallSec 典型 10-30 秒，
	// 8 秒远小于其最小值，确保只捕获"雷达尚未升级为 Fall 就回撤"的误报。
)

// BedSession 单 sleepad 设备的"在床会话"状态机。
//
// 用途：实现 silent fall（基于 sleepad LeftBed 与 radar 仍在 bed 邻域的矛盾）。
//
// 入场门控（PR-14）：
//   只有当同房间 sleepad InBed 与 radar InBed 在 ±15s 内都到齐才会"成立"——
//   RadarInBedConfirmedMs > 0 是 silent fall 触发先决条件；
//   防止单源（仅 sleepad / 仅 radar）误报。
//
// 生命周期：
//   sleepad InBed 首次到达             → InBedSinceMs 记录起点；HasHRRR 重置
//   radar InBed 同房间 ±15s 内到达     → RadarInBedConfirmedMs 标记
//   sleepad observation 含 HR/RR > 0   → HasHRRR = true
//   sleepad LeftBed 到达                → 若已满 MinInBedSec，进入「等待矛盾」状态
//   每 tick 检查超时                     → 若 radar 仍在 Bed 邻域 → 报 silent fall
//   重新 InBed                          → 重置 session
//
// LeftBedHadHRRR / LeftBedMaxPeople 在 LeftBed 时刻 latch，不受后续观测影响。
// PR-9: MaxPeople 中间状态删除（依赖 bedPersonCount 已删）；LeftBedMaxPeople latch 字段保留
// 但 PR-9 阶段无 source 值固定 0，PR-11 silent_fall 重写时改由 SuiteCensus 派生。
type BedSession struct {
	DeviceUID             string // sleepad device_uid
	InBedSinceMs          int64  // 首次 InBed 到达的时间戳；0 = 未在床
	RadarInBedConfirmedMs int64  // PR-14：sleepad+radar InBed ±15s 一致确认时刻；0 = 未双源确认
	HasHRRR               bool   // in-bed 期间是否观测到 HR/RR > 0
	LeftBedAtMs           int64  // 0 = 仍在床；>0 = 等待矛盾窗口
	LeftBedHadHRRR        bool   // LeftBed 时刻的 HasHRRR latch
	LeftBedMaxPeople      int    // LeftBed 时刻的"多人床"latch；PR-9 阶段无 source 固定 0，PR-11 重写
	SilentFallAlerted     bool   // 防重复触发
}

// PendingLostFall 已消失但等待 cell-area-typed 时长复现窗口的 track（lost-fall 规则）。
//
// 等待时长按消失点 cell areaType（5min~1h）+ frozen credit；
// 触发：track 消失 + checkLostFall 通过（离门 >1m，非 Enter 区，年龄足够）
// 取消：① 新 track 出生（含 BlindSpotRecovery 反馈）② ExitRoom 事件 ③ room.NumberPeople ≥ 2
type PendingLostFall struct {
	OriginalTrackID int
	DeviceID        string
	RoomID          string
	LastX, LastY    int
	LastZ           int
	LastScore       int
	LastVerdict     TrackVerdict
	LastCellArea    AreaType // 消失点 cell.Belief[0].Type，决定 wait 时长
	DisappearMs     int64
	FrozenStartMs   int64 // 0 = 无 still box run；>0 = run 起点（半计入 wait credit + PR-C 守卫）
	SpatialJump     bool  // MaxImpliedSpeedFromBirth > SuspectSpeedCm（更敏感）
}

// NewTrackState 新 track 出生
func NewTrackState(trackID int, deviceID, roomID string, x, y, z int, tMs int64) *TrackState {
	birth := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	return &TrackState{
		TrackID:              trackID,
		DeviceID:             deviceID,
		RoomID:               roomID,
		BirthPos:             birth,
		Kalman:               NewKalmanFilter2D(float64(x), float64(y)),
		History:              []TimedPoint{birth},
		FrameCount:           1,
		Score:                50,
		Verdict:              VerdictPending,
		LastZ:                z,
		LastUpdateMs:         tMs,
		LastObservedMs:       tMs,
		LastCellCol:          -1,
		LastCellRow:          -1,
		BirthFinalDeadlineMs: tMs + int64(FallRulesParam.Lost.BirthFinalGraceMs),
	}
}

// PushPoint 追加一帧观测到历史窗口
func (ts *TrackState) PushPoint(x, y, z int, tMs int64) {
	pt := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	ts.History = append(ts.History, pt)
	if len(ts.History) > HistoryLen {
		ts.History = ts.History[len(ts.History)-HistoryLen:]
	}
	ts.FrameCount++
	ts.LastUpdateMs = tMs
	ts.LastObservedMs = tMs
}

// HasHistory 是否有足够帧数做 Kalman
func (ts *TrackState) HasHistory() bool {
	return ts.FrameCount >= 2
}

// DisplacementWithinMs 计算 History 中过去 windowMs 内的位移 box 对角线（cm）。
//
// 用于 frozen 检测的 box 判据：仅看 (max-min) 范围而非每帧累计，对 firmware
// X/Y 抖动 + 摔倒抽搐鲁棒——只要轨迹被框在 box 内，box 对角线即上限。
//
// nowMs 通常是当前帧时间戳；窗口 = [nowMs - windowMs, nowMs]。
// History 最多 HistoryLen=30 帧 = 30s 滚动窗口；超出此长度无法回看。
func (ts *TrackState) DisplacementWithinMs(windowMs int64, nowMs int64) int {
	cutoff := nowMs - windowMs
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32
	count := 0
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		count++
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if count < 2 {
		return 0
	}
	dx := maxX - minX
	dy := maxY - minY
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
}

// TotalDisplacement 历史窗口内的总位移（cm）
func (ts *TrackState) TotalDisplacement() int {
	if len(ts.History) < 2 {
		return 0
	}
	sum := 0
	for i := 1; i < len(ts.History); i++ {
		sum += distInt(ts.History[i-1].X, ts.History[i-1].Y, ts.History[i].X, ts.History[i].Y)
	}
	return sum
}

// AgeSec 存活时长（秒）
func (ts *TrackState) AgeSec() int {
	if len(ts.History) < 2 {
		return 0
	}
	lastMs := ts.History[len(ts.History)-1].TMs
	return int((lastMs - ts.BirthPos.TMs) / 1000)
}

// AdjustScore 调整分数，限制在 [0, 100]
func (ts *TrackState) AdjustScore(delta int) {
	ts.Score = clampInt(ts.Score+delta, 0, 100)
}
