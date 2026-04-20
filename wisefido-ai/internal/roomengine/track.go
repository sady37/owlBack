package roomengine

// TrackVerdict track 判定结果。
type TrackVerdict uint8

const (
	VerdictPending   TrackVerdict = 0 // 试用期中，尚未判定
	VerdictReal      TrackVerdict = 1 // 确认为真实 track
	VerdictGhost     TrackVerdict = 2 // 判定为 ghost
)

// Anomaly track 异常类型。
type Anomaly uint8

const (
	AnomalyNone         Anomaly = 0
	AnomalyFall         Anomaly = 1 // 跌倒（z 突降 + 速度归零）
	AnomalyStillTooLong Anomaly = 2 // 静止超时（空地>8min / 卫生间>15min）
	AnomalyPathBreak    Anomaly = 3 // 轨迹断裂（非出口处突然消失）
	AnomalyPoseMismatch Anomaly = 4 // pose 与运动学矛盾（金属干扰）
)

// TimedPoint 带时间戳的坐标（房间坐标 cm，时间 ms）。
type TimedPoint struct {
	X, Y float64
	Z    float64 // 高度 cm（用于跌倒检测）
	TMs  int64   // 毫秒时间戳
}

// TrackState 单个 track 的完整生命档案。
type TrackState struct {
	// ---- 身份 ----
	TrackID   int    // 雷达原始 track_id
	DeviceID  string // 来源设备
	RoomID    string // 所在房间

	// ---- 出生 ----
	BirthPos  TimedPoint // 首次出现的位置和时间
	BirthScore float64   // 5 因子初始评分 [0,100]

	// ---- Kalman 跟踪 ----
	Kalman    *KalmanFilter2D // 预测/更新/残差

	// ---- 历史轨迹 ----
	History   []TimedPoint // 最近 HistoryLen 帧的坐标（滚动窗口）
	FrameCount int         // 已累计帧数

	// ---- 评分 ----
	Score     float64      // 当前综合可信度 [0,100]
	Verdict   TrackVerdict // 判定结果
	ConfirmedAtMs int64    // 被确认 real 的时间（ms），0=未确认

	// ---- 静止检测 ----
	StillSince  int64   // 开始静止的时间（ms），0=在动
	StillX, StillY float64 // 静止起始坐标

	// ---- 异常 ----
	CurrentAnomaly Anomaly
	LastPose       int     // 上一帧的 pose（用于 mismatch 检测）
	LastZ          float64 // 上一帧的 z（用于跌倒检测）

	// ---- 最后更新 ----
	LastUpdateMs int64
}

const (
	HistoryLen      = 30  // 保留最近 30 帧（30 秒 @1Hz）
	ProbationFrames = 5   // 试用期 5 帧（5 秒）
	ScoreConfirmTh  = 50  // 连续 5 帧 score > 50 → confirmed
	ScoreGhostTh    = 20  // 连续 5 帧 score < 20 → ghost
	StillThreshCm   = 15  // 位移 < 15cm/帧 视为静止
)

// NewTrackState 新 track 出生。
func NewTrackState(trackID int, deviceID, roomID string, x, y, z float64, tMs int64) *TrackState {
	birth := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	return &TrackState{
		TrackID:      trackID,
		DeviceID:     deviceID,
		RoomID:       roomID,
		BirthPos:     birth,
		Kalman:       NewKalmanFilter2D(x, y),
		History:      []TimedPoint{birth},
		FrameCount:   1,
		Score:        50, // 中立起始分
		Verdict:      VerdictPending,
		LastZ:        z,
		LastUpdateMs: tMs,
	}
}

// PushPoint 追加一帧观测，更新历史窗口。
func (ts *TrackState) PushPoint(x, y, z float64, tMs int64) {
	pt := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	ts.History = append(ts.History, pt)
	if len(ts.History) > HistoryLen {
		ts.History = ts.History[len(ts.History)-HistoryLen:]
	}
	ts.FrameCount++
	ts.LastUpdateMs = tMs
}

// HasHistory 是否有足够帧数做 Kalman（至少 2 帧才有速度）。
func (ts *TrackState) HasHistory() bool {
	return ts.FrameCount >= 2
}

// TotalDisplacement 历史窗口内的总位移（cm）。
func (ts *TrackState) TotalDisplacement() float64 {
	if len(ts.History) < 2 {
		return 0
	}
	var sum float64
	for i := 1; i < len(ts.History); i++ {
		sum += dist(ts.History[i-1].X, ts.History[i-1].Y, ts.History[i].X, ts.History[i].Y)
	}
	return sum
}

// AgeSec 存活时长（秒）。
func (ts *TrackState) AgeSec() float64 {
	if len(ts.History) < 2 {
		return 0
	}
	return float64(ts.History[len(ts.History)-1].TMs-ts.BirthPos.TMs) / 1000.0
}

// AdjustScore 调整分数，限制在 [0,100]。
func (ts *TrackState) AdjustScore(delta float64) {
	ts.Score = clamp(ts.Score+delta, 0, 100)
}
