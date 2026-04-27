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

// TrackState 单个 track 的完整生命档案。
// 运动粒子：Kalman + 即时判定 + 状态机转移。
// 不累计空间属性（那些归 Cell）。
type TrackState struct {
	// ---- 身份 ----
	TrackID  int
	DeviceID string
	RoomID   string

	// ---- 出生档案 ----
	BirthPos   TimedPoint
	BirthScore int // 5 因子初始分 [0,100]

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

// NewTrackState 新 track 出生
func NewTrackState(trackID int, deviceID, roomID string, x, y, z int, tMs int64) *TrackState {
	birth := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	return &TrackState{
		TrackID:      trackID,
		DeviceID:     deviceID,
		RoomID:       roomID,
		BirthPos:     birth,
		Kalman:       NewKalmanFilter2D(float64(x), float64(y)),
		History:      []TimedPoint{birth},
		FrameCount:   1,
		Score:        50,
		Verdict:      VerdictPending,
		LastZ:        z,
		LastUpdateMs: tMs,
		LastCellCol:  -1,
		LastCellRow:  -1,
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
