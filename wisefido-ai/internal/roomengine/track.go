package roomengine

// TrackVerdict track 判定结果
type TrackVerdict uint8

const (
	VerdictPending TrackVerdict = 0
	VerdictReal    TrackVerdict = 1
	VerdictGhost   TrackVerdict = 2
)

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

// frameSignature 用于 frozen-frame 检测：firmware 失锁后 (track_id, x, y, z, pose, tc, rt)
// 会保持 byte-equal 5+ 分钟。直接用 struct equality 比对，简单且无 hash 碰撞风险。
type frameSignature struct {
	TrackID         int
	X, Y, Z         int
	Pose            int
	TrackConfidence int
	RemainingTime   int
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

	// ---- 异常与 Silent Fall ----
	CurrentAnomaly Anomaly
	SilentFall     bool

	// ---- 最后观测 ----
	LastPose     int
	LastZ        int
	LastUpdateMs int64

	// ---- cell 穿越追踪（Walk 学习用）----
	// 初始为 -1 表示尚未定位；新 cell 进入且 core==Move 时调 grid.MarkTraverse 计数 ++。
	LastCellCol int
	LastCellRow int

	// ---- Frozen-frame 检测（lazy 反应式，每帧 O(1) 维护）----
	// LastFrameSig：上一帧关键字段签名；FrozenSameCount：连续相同帧数（含当前帧）；
	// FrozenRunStart：当前 frozen 起点 ms（0 = 未达 frozen 状态）。
	// firmware 失锁特征：连续 25 帧 (tid, x, y, z, pose, tc, rt) 字面相同，
	// 用于 lost-fall pending 时计算 frozen credit（半计入等待）。
	LastFrameSig    frameSignature
	FrozenSameCount int
	FrozenRunStart  int64

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

	// Silent Fall 60 秒挂起：消失后挂起，60 秒内若同位置附近有新 track 出生（且非 Lie 姿）
	// → 视为 occlusion 复现，取消挂起；超时未取消才真报 silent fall。
	PendingSilentFallMs = 60_000
	PendingMatchDistCm  = 100 // 复现匹配距离（人被遮挡再出现，雷达位置可漂 50-100cm）
)

// PendingSilentFall 已消失但等待 60 秒复现窗口的 track。
// 由 segment 2（消失判定）写入，segment 1（新 track 出生）按位置取消，segment 5（扫描）超时报警。
type PendingSilentFall struct {
	OriginalTrackID int
	DeviceID        string
	RoomID          string
	LastX, LastY    int
	LastZ           int
	LastScore       int
	LastVerdict     TrackVerdict
	DisappearMs     int64 // 进入 pending 的时刻（≈MissCount 超阈值的 nowMs）
}

// BedSession 单 sleepad 设备的"在床会话"状态机。
//
// 用途：实现新版 silent fall（基于 sleepad LeftBed 与 radar 仍在 bed 邻域的矛盾）。
//
// 生命周期：
//   sleepad InBed 首次到达    → InBedSinceMs 记录起点；HasHRRR 重置
//   sleepad observation 含 HR/RR > 0 → HasHRRR = true（在 in-bed 期间任意时刻命中）
//   sleepad LeftBed 到达       → 若已满 MinInBedSec，进入「等待矛盾」状态，记 LeftBedAtMs
//   每 tick 检查超时             → 若 radar 仍在 Bed 邻域 → 报 silent fall
//   重新 InBed                  → 重置 session（认为是新一轮上床）
//
// LeftBedHadHRRR / LeftBedMaxPeople 在 LeftBed 时刻 latch，不受后续观测影响。
type BedSession struct {
	DeviceUID         string // sleepad device_uid
	InBedSinceMs      int64  // 首次 InBed 到达的时间戳；0 = 未在床
	MaxPeople         int    // 该 session 期间见过的最大 bedPersonCount[device]
	HasHRRR           bool   // in-bed 期间是否观测到 HR/RR > 0
	LeftBedAtMs       int64  // 0 = 仍在床；>0 = 等待矛盾窗口
	LeftBedHadHRRR    bool   // LeftBed 时刻的 HasHRRR latch
	LeftBedMaxPeople  int    // LeftBed 时刻的 MaxPeople latch
	SilentFallAlerted bool   // 防重复触发
}

// PendingLostFall 已消失但等待 cell-area-typed 时长复现窗口的 track（lost-fall 规则）。
//
// 与 PendingSilentFall 区别：
//   - silent 仅 60s + sleepad InBed 兜底（bedroom 专用）
//   - lost 按消失点 cell areaType 分时长（5min~1h）+ ExitRoom + NumberPeople≥2 兜底（全屋通用）
//
// 触发：track 消失且不满足 silent fall 条件 + checkLostFall 通过（离门 >1m，非 Enter 区，年龄足够）
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
	FrozenStartMs   int64 // 0 = 无 frozen run；>0 = frozen 起点（用于半计入 wait credit）
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
}

// HasHistory 是否有足够帧数做 Kalman
func (ts *TrackState) HasHistory() bool {
	return ts.FrameCount >= 2
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
