package roomengine

import (
	"math"
	"sync"
	"time"

	"owl-common/observation"
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
}

// TrackManager 管理一个房间内所有 track 的生命周期
type TrackManager struct {
	mu      sync.Mutex
	roomID  string
	grid    *RoomGrid
	tracks  map[int]*TrackState
	outputs map[int]*TrackOutput

	sleepadInBedCount int
}

// NewTrackManager 创建 track 管理器
func NewTrackManager(roomID string, grid *RoomGrid) *TrackManager {
	return &TrackManager{
		roomID:  roomID,
		grid:    grid,
		tracks:  make(map[int]*TrackState),
		outputs: make(map[int]*TrackOutput),
	}
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
	tm.mu.Lock()
	defer tm.mu.Unlock()

	nowMs := time.Now().UnixMilli()
	if len(frames) > 0 {
		nowMs = frames[0].TMs
	}
	activeIDs := make(map[int]bool)

	// ========== 段 1: 观测到的 track ==========
	for _, f := range frames {
		activeIDs[f.TrackID] = true
		ts, exists := tm.tracks[f.TrackID]

		var quality, vx, vy, dtSec int

		if !exists {
			// 新 track 出生
			ts = NewTrackState(f.TrackID, f.DeviceID, tm.roomID, f.X, f.Y, f.Z, f.TMs)
			ts.BirthScore = tm.birthScore(f.X, f.Y)
			ts.Score = ts.BirthScore
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
			// 消失判定
			if ts.Verdict == VerdictReal && tm.checkSilentFall(ts) {
				ts.CurrentAnomaly = AnomalyFall
				ts.SilentFall = true
				if n := len(ts.History); n > 0 {
					last := ts.History[n-1]
					tm.grid.MarkFallEvent(last.X, last.Y, nowMs)
				}
				tm.writeSilentFallOutput(ts, nowMs)
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
		if ts.FrameCount >= ProbationFrames {
			if ts.Score >= ScoreConfirmTh {
				ts.Verdict = VerdictReal
				ts.ConfirmedAtMs = nowMs
			} else if ts.Score < ScoreGhostTh {
				ts.Verdict = VerdictGhost
			}
		}
	}

	// ========== 段 4: 构建输出 ==========
	results := make([]TrackOutput, 0, len(tm.tracks))
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

// writeSilentFallOutput 消失前为 Silent Fall 写一条输出
func (tm *TrackManager) writeSilentFallOutput(ts *TrackState, nowMs int64) {
	pxF, pyF := ts.Kalman.Position()
	out := &TrackOutput{
		TrackID:  ts.TrackID,
		DeviceID: ts.DeviceID,
		RoomID:   ts.RoomID,
		Verdict:  ts.Verdict,
		Score:    ts.Score,
		Risk:     tm.computeRisk(ts, 0, nowMs),
		Anomaly:  AnomalyFall,
		X:        int(math.Round(pxF)),
		Y:        int(math.Round(pyF)),
		Z:        ts.LastZ,
		Source:   "engine_silent",
	}
	tm.outputs[ts.TrackID] = out
}

// ========================================================================
// 出生打分（5 因子）
// ========================================================================

func (tm *TrackManager) birthScore(x, y int) int {
	score := 50

	// 因子 1: d_enter
	dEntry := tm.grid.NearestEntryDist(x, y)
	switch {
	case dEntry < 50:
		score += 20
	case dEntry < 150:
		score += 10
	case dEntry > 300:
		score -= 25
	default:
		score -= 10
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
		}
	}

	// 因子 3: 房间已有 track 数
	nExisting := len(tm.tracks)
	if nExisting == 0 {
		score += 10
	} else if nExisting >= 2 {
		score -= 10
	}

	return clampInt(score, 0, 100)
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
		// 静止超时
		if cell != nil {
			timeout := cell.StillTimeoutSec()
			if timeout > 0 {
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
	} else {
		// 在动：刚从静止恢复，记 Dwell EMA
		if ts.StillSince > 0 {
			dwellSec := int((nowMs - ts.StillSince) / 1000)
			if dwellSec > 0 {
				tm.grid.MarkDwell(ts.StillX, ts.StillY, dwellSec, nowMs)
			}
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

func (tm *TrackManager) checkSilentFall(ts *TrackState) bool {
	if ts.AgeSec() < 10 {
		return false
	}
	n := len(ts.History)
	if n < 3 {
		return false
	}

	// 消失前 3 帧 min(Z) < 20
	minZ := ts.History[n-1].Z
	for i := n - 3; i < n; i++ {
		if i < 0 {
			continue
		}
		if ts.History[i].Z < minZ {
			minZ = ts.History[i].Z
		}
	}
	if minZ >= 20 {
		return false
	}

	// 消失 cell EdgeDist > 30 且非 Enter 区
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

	// 消失前最后两帧位移 > 80
	d := distInt(ts.History[n-1].X, ts.History[n-1].Y, ts.History[n-2].X, ts.History[n-2].Y)
	if d <= 80 {
		return false
	}
	return true
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
