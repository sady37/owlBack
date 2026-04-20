package roomengine

import (
	"sync"
	"time"
)

// TrackOutput Room Engine 对外输出的单条 track 评估结果。
type TrackOutput struct {
	TrackID   int
	DeviceID  string
	RoomID    string
	Verdict   TrackVerdict // real / ghost / pending
	Score     float64      // 可信度 [0,100]
	Risk      float64      // 综合风险分（含时间/在场因子）
	Anomaly   Anomaly      // 异常类型
	X, Y, Z   float64      // 当前估计位置（房间坐标 cm）
	VX, VY    float64      // 当前估计速度 cm/s
	StillSec  float64      // 已静止时长（秒）
}

// TrackFrame 一帧输入（从 iot:monitor:stream 解析后）。
type TrackFrame struct {
	TrackID  int
	DeviceID string
	X, Y, Z  float64 // 房间坐标 cm
	Pose     int
	AreaType int
	TMs      int64   // 毫秒时间戳
}

// TrackManager 管理一个房间内所有 track 的生命周期和评分。
type TrackManager struct {
	mu      sync.Mutex
	roomID  string
	grid    *RoomGrid
	tracks  map[int]*TrackState // trackID → state
	outputs map[int]*TrackOutput // 最新输出缓存

	// Sleepad 上床事件计数（由外部注入）
	sleepadInBedCount int
}

// NewTrackManager 创建 track 管理器。
func NewTrackManager(roomID string, grid *RoomGrid) *TrackManager {
	return &TrackManager{
		roomID:  roomID,
		grid:    grid,
		tracks:  make(map[int]*TrackState),
		outputs: make(map[int]*TrackOutput),
	}
}

// SetSleepadInBedCount 外部更新 Sleepad 在床人数。
func (tm *TrackManager) SetSleepadInBedCount(count int) {
	tm.mu.Lock()
	tm.sleepadInBedCount = count
	tm.mu.Unlock()
}

// ProcessFrame 处理一帧所有 track 数据。frames 为本帧所有 track 观测。
// 返回本帧所有 track 的评估结果。
func (tm *TrackManager) ProcessFrame(frames []TrackFrame) []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	nowMs := time.Now().UnixMilli()
	if len(frames) > 0 {
		nowMs = frames[0].TMs
	}
	activeIDs := make(map[int]bool)

	// ---- 1. 处理每个观测到的 track ----
	for _, f := range frames {
		activeIDs[f.TrackID] = true
		ts, exists := tm.tracks[f.TrackID]

		if !exists {
			// 新 track 出生
			ts = NewTrackState(f.TrackID, f.DeviceID, tm.roomID, f.X, f.Y, f.Z, f.TMs)
			ts.BirthScore = tm.birthScore(f.X, f.Y, f.TMs)
			ts.Score = ts.BirthScore
			tm.tracks[f.TrackID] = ts
		} else {
			// 已有 track：Kalman predict + update
			dt := float64(f.TMs-ts.LastUpdateMs) / 1000.0
			if dt <= 0 {
				dt = 1
			}
			ts.Kalman.Predict(dt)
			residual := ts.Kalman.Update(f.X, f.Y)
			ts.PushPoint(f.X, f.Y, f.Z, f.TMs)

			// 残差打分
			tm.scoreResidual(ts, residual)

			// 运动/静止检测
			tm.scoreMovement(ts, f.X, f.Y, f.Pose, nowMs)

			// 跌倒检测（z 突降）
			tm.detectFall(ts, f.Z, f.Pose)

			// pose vs 运动学矛盾检测
			tm.detectPoseMismatch(ts, f.Pose, f.X, f.Y)

			ts.LastPose = f.Pose
			ts.LastZ = f.Z
		}
	}

	// ---- 2. 处理消失的 track ----
	for id, ts := range tm.tracks {
		if activeIDs[id] {
			continue
		}
		// 本帧未观测到 → Kalman 空转
		dt := float64(nowMs-ts.LastUpdateMs) / 1000.0
		if dt <= 0 {
			dt = 1
		}
		ts.Kalman.PredictOnly(dt)
		ts.LastUpdateMs = nowMs

		// 连续丢失 > 10 帧 → 判定消失
		if ts.Kalman.MissCount > 10 {
			// 检查消失位置是否在出口
			px, py := ts.Kalman.Position()
			if !tm.grid.IsEdge(px, py, 50) && ts.Verdict == VerdictReal {
				// 非出口处消失 → 异常
				ts.CurrentAnomaly = AnomalyPathBreak
			}
			// 更新网格
			if ts.Verdict == VerdictGhost {
				tm.grid.MarkGhostBirth(ts.BirthPos.X, ts.BirthPos.Y, nowMs)
			}
			delete(tm.tracks, id)
			continue
		}
	}

	// ---- 3. 试用期判定 ----
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
			// 分数在中间 → 继续观察，不急着判
		}
	}

	// ---- 4. 构建输出 ----
	results := make([]TrackOutput, 0, len(tm.tracks))
	for _, ts := range tm.tracks {
		px, py := ts.Kalman.Position()
		vx, vy := ts.Kalman.Velocity()
		stillSec := 0.0
		if ts.StillSince > 0 {
			stillSec = float64(nowMs-ts.StillSince) / 1000.0
		}

		risk := tm.computeRisk(ts, stillSec, nowMs)

		out := TrackOutput{
			TrackID:  ts.TrackID,
			DeviceID: ts.DeviceID,
			RoomID:   ts.RoomID,
			Verdict:  ts.Verdict,
			Score:    ts.Score,
			Risk:     risk,
			Anomaly:  ts.CurrentAnomaly,
			X:        px,
			Y:        py,
			Z:        ts.LastZ,
			VX:       vx,
			VY:       vy,
			StillSec: stillSec,
		}
		results = append(results, out)
		tm.outputs[ts.TrackID] = &out
	}

	// ---- 5. 更新网格（confirmed track 走过的路径）----
	for _, ts := range tm.tracks {
		if ts.Verdict == VerdictReal && len(ts.History) > 0 {
			last := ts.History[len(ts.History)-1]
			vx, vy := ts.Kalman.Velocity()
			tm.grid.MarkRealVisit(last.X, last.Y, vx, vy, nowMs)
		}
	}

	return results
}

// GetOutputs 返回当前所有 track 的最新输出（供外部查询）。
func (tm *TrackManager) GetOutputs() []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	out := make([]TrackOutput, 0, len(tm.outputs))
	for _, o := range tm.outputs {
		out = append(out, *o)
	}
	return out
}

// ======================== 内部评分逻辑 ========================

// birthScore 5 因子初始评分。
func (tm *TrackManager) birthScore(x, y float64, tMs int64) float64 {
	score := 50.0

	// 因子 1: d_edge — 离最近入口的距离
	dEntry := tm.grid.NearestEntryDist(x, y)
	if dEntry < 50 { // < 50cm = 在入口附近
		score += 20
	} else if dEntry < 150 {
		score += 10
	} else if dEntry > 300 { // > 3m = 远离入口，凭空出现
		score -= 25
	} else {
		score -= 10
	}

	// 因子 2: grid_history — 出生格子的 ghost 概率
	cell := tm.grid.CellAt(x, y)
	if cell != nil {
		ghostRatio := cell.GhostRatio()
		if ghostRatio > 0.7 {
			score -= 25 // 反射热点
		} else if ghostRatio > 0.4 {
			score -= 10
		}
		// 入口格子加分
		if cell.IsEntry() {
			score += 15
		}
	}

	// 因子 3: n_co — 出生时已有多少 track
	nExisting := len(tm.tracks)
	if nExisting == 0 {
		score += 10 // 孤立出现 → 大概率真
	} else if nExisting >= 2 {
		score -= 10 // 已经有多个 track → 伴生嫌疑
	}

	// 因子 4/5 (Δd 和 Σd/age) 在后续帧中累计，出生时无数据

	return clamp(score, 0, 100)
}

// scoreResidual 根据 Kalman 残差调整分数。
func (tm *TrackManager) scoreResidual(ts *TrackState, residual float64) {
	if residual < 30 { // < 30cm = 运动平滑
		ts.AdjustScore(2)
	} else if residual < 80 { // 中等偏差
		ts.AdjustScore(0) // 不加不减
	} else if residual < 200 { // 较大偏差
		ts.AdjustScore(-5)
	} else { // > 2m = 空间跳跃
		ts.AdjustScore(-20)
	}
}

// scoreMovement 运动/静止评分。
func (tm *TrackManager) scoreMovement(ts *TrackState, x, y float64, pose int, nowMs int64) {
	if len(ts.History) < 2 {
		return
	}
	prev := ts.History[len(ts.History)-2]
	d := dist(prev.X, prev.Y, x, y)

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
			// 非休息区静止 → 扣分（ghost 不动）
			ts.AdjustScore(-3)
		}
		// 静止超时检查
		if cell != nil {
			timeout := cell.StillTimeoutSec()
			if timeout > 0 {
				stillSec := float64(nowMs-ts.StillSince) / 1000.0
				if stillSec > float64(timeout) {
					ts.CurrentAnomaly = AnomalyStillTooLong
				}
			}
		}
		// 更新网格静止计数
		dtSec := float64(nowMs-prev.TMs) / 1000.0
		tm.grid.MarkStill(x, y, dtSec, nowMs)
	} else {
		// 在动
		ts.StillSince = 0
		if ts.CurrentAnomaly == AnomalyStillTooLong {
			ts.CurrentAnomaly = AnomalyNone // 动了，清除静止超时
		}
		// 有位移 → 加分（真人在动）
		speed := ts.Kalman.Speed()
		if speed > 10 && speed < 150 { // 10-150 cm/s = 正常老人步速
			ts.AdjustScore(3)
		} else if speed >= 150 { // > 1.5m/s = 老人不太可能
			ts.AdjustScore(-2)
		}
	}

	// 因子 5: Σd/age — 总位移÷存活时间
	age := ts.AgeSec()
	if age > 5 { // 至少 5 秒后才看这个指标
		avgSpeed := ts.TotalDisplacement() / age
		if avgSpeed < 2 && !isRest { // < 2cm/s 平均速度，不在休息区
			ts.AdjustScore(-2)
		}
	}
}

// detectFall z 突降检测。
func (tm *TrackManager) detectFall(ts *TrackState, z float64, pose int) {
	if ts.LastZ == 0 {
		return
	}
	dz := ts.LastZ - z // 正值 = 高度下降
	speed := ts.Kalman.Speed()

	// z 下降 > 50cm 且速度降到接近 0 → 跌倒
	if dz > 50 && speed < 20 {
		ts.CurrentAnomaly = AnomalyFall
	}
}

// detectPoseMismatch pose 与运动学矛盾检测。
// 例如：pose=walking 但位移=0，或 pose=standing 但 z 在地面。
func (tm *TrackManager) detectPoseMismatch(ts *TrackState, pose int, x, y float64) {
	speed := ts.Kalman.Speed()
	// pose=1(walking) 但速度 ≈ 0
	if pose == 1 && speed < 5 && ts.FrameCount > 3 {
		tm.grid.MarkPoseMismatch(x, y, ts.LastUpdateMs)
		ts.CurrentAnomaly = AnomalyPoseMismatch
		ts.AdjustScore(-3)
	}
	// pose=3(standing) 但 z < 50cm（在地面）
	if pose == 3 && ts.LastZ > 0 && ts.LastZ < 50 {
		cell := tm.grid.CellAt(x, y)
		if cell != nil && cell.PoseMismatch > 3 {
			// 此格子历史上 pose 经常不准 → 以运动学为准
			ts.CurrentAnomaly = AnomalyFall
		}
		tm.grid.MarkPoseMismatch(x, y, ts.LastUpdateMs)
	}
}

// computeRisk 综合风险分 = 异常基础分 × 时间因子 × 在场因子。
func (tm *TrackManager) computeRisk(ts *TrackState, stillSec float64, nowMs int64) float64 {
	if ts.Verdict == VerdictGhost {
		return 0 // ghost 不计风险
	}

	base := 0.0
	switch ts.CurrentAnomaly {
	case AnomalyFall:
		base = 100
	case AnomalyStillTooLong:
		base = 60
	case AnomalyPathBreak:
		base = 80
	case AnomalyPoseMismatch:
		base = 70
	default:
		base = 0
	}

	if base == 0 {
		return 0
	}

	tf := timeFactor(nowMs)
	of := tm.occupancyFactor()

	return clamp(base*tf*of, 0, 100)
}

// timeFactor 时间风险因子。
func timeFactor(nowMs int64) float64 {
	hour := time.UnixMilli(nowMs).Hour()
	switch {
	case hour >= 5 && hour < 6:
		return 2.0 // 生理最低谷
	case hour >= 22 || hour < 5:
		return 1.5 // 夜间
	case hour >= 6 && hour < 8:
		return 1.3 // 晨起
	default:
		return 1.0
	}
}

// occupancyFactor 在场因子。
func (tm *TrackManager) occupancyFactor() float64 {
	realCount := 0
	for _, ts := range tm.tracks {
		if ts.Verdict == VerdictReal {
			realCount++
		}
	}
	bedOccupied := tm.sleepadInBedCount > 0

	if realCount <= 1 {
		return 1.5 // 独自一人
	}
	if realCount == 2 && bedOccupied {
		return 1.2 // 另一人在床
	}
	return 1.0
}
