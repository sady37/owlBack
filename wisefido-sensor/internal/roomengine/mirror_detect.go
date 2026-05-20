// mirror_detect.go — L1 镜面对称 ghost 配对检测（无 layout 镜面坐标先验）。
//
// 物理：人 A 在镜前 → 雷达多径 → "镜后" 出现 ghost B = reflect(A, mirror_line)。
// 算法仅靠 5 帧配对样本 + 雷达位置三个数据，识别 (A, B) 是否镜像对、并判 B 为 ghost：
//
//   I1 (B-A) 方向恒定           ：镜面线方向固定，|B-A| 始终垂直于它 → v_t 角度方差小
//   I2 中点 M_t 共线 且 ⊥ v       ：M_t = (A_t+B_t)/2 全部落在镜面线上
//   I3 同步等速                  ：|A_t-A_{t-1}| ≈ |B_t-B_{t-1}|（独立两人不可能）
//
// Tiebreaker（谁是 ghost）：物理上 ghost 反射路径更长 → 距 radar 远者 = ghost。
//
// 自学习：命中时每帧的 bounce 点（线段 R'-A_t 与 mirror_line 交点）调 grid.MarkMirrorBounce
//   累 2×2 微块 MBC；同一 cell 多次独立配对累 ≥ MirrorPromoteThreshold (3) 后升 AreaDeny + SourceLearned。
//   不写回 cfg.Interferes；layout 是人工真相由 FE 接受用户操作后通过 wisefido-data API 改。

package roomengine

import (
	"math"
	"sort"

	"owl-common/radarutils"

	"go.uber.org/zap"
)

// mirror 阈值常量（实测可调）
const (
	mirrorMinSamples       = 5    // 满 5 帧才评估
	mirrorVAngleStdRad     = 0.26 // ≈ 15°：v_t 角度标准差上限（rad）
	mirrorMidpointRMSECm   = 20.0 // 中点到拟合直线 RMSE 上限（cm）
	mirrorOrthogonalCos    = 0.20 // |cos(midline 方向, mean v)| 上限（cos 5°≈0.087；放宽到 ~78° 容忍噪声）
	mirrorSpeedRatioMax    = 0.40 // |s_A - s_B| / max(s_A, s_B) 上限（独立行人不会这么同步）
	mirrorMinTotalMoveCm   = 30   // 5 帧累计位移下限：静止场景跳过判定
	mirrorGhostPenaltyAdd  = 50   // 命中 ghost 累加的 GhostPenalty
)

// mirrorPairKey 两 track 配对（小 ID 在前，固定 key 顺序）
type mirrorPairKey struct {
	A, B int
}

func newMirrorPairKey(a, b int) mirrorPairKey {
	if a > b {
		a, b = b, a
	}
	return mirrorPairKey{a, b}
}

// mirrorPairSample 单帧两 track 配对快照
type mirrorPairSample struct {
	Ax, Ay int
	Bx, By int
	Ts     int64
}

// mirrorPairBuffer per-pair 滑窗
type mirrorPairBuffer struct {
	samples []mirrorPairSample
	// 命中后冷却：避免每帧重复 paint bounce + 加 penalty
	lastConfirmedMs int64
}

// pushSample 滑窗追加，超过 mirrorMinSamples 丢最旧。
func (b *mirrorPairBuffer) pushSample(s mirrorPairSample) {
	b.samples = append(b.samples, s)
	if len(b.samples) > mirrorMinSamples {
		b.samples = b.samples[len(b.samples)-mirrorMinSamples:]
	}
}

// mirrorEvaluation evaluateMirrorPair 输出
type mirrorEvaluation struct {
	IsMirror      bool
	GhostTrackID  int      // tiebreaker 选出的 ghost 端（A 或 B 的 trackID）
	RealTrackID   int      // 配对中的 real 端
	BouncePoints  []radarutils.Point
	MidlinePoint  radarutils.Point // 镜面线上锚点（中点重心）
	MidlineDirX   float64          // 镜面线方向单位向量
	MidlineDirY   float64
}

// evaluateMirrorPair 给定 5 帧配对样本 + 雷达位置，跑三不变量 + tiebreaker。
// IDs 是两 track id（与样本顺序一致：第一个传 A trackID，第二个传 B trackID）。
func evaluateMirrorPair(samples []mirrorPairSample, idA, idB int, radarX, radarY int) mirrorEvaluation {
	out := mirrorEvaluation{}
	if len(samples) < mirrorMinSamples {
		return out
	}

	// 累计位移门：5 帧两 track 位移都过小 → 静态场景跳过
	if !hasMinMove(samples, mirrorMinTotalMoveCm) {
		return out
	}

	// === I1 v_t 方向恒定 ===
	angles := make([]float64, len(samples))
	for i, s := range samples {
		vx := float64(s.Bx - s.Ax)
		vy := float64(s.By - s.Ay)
		angles[i] = math.Atan2(vy, vx)
	}
	if circStd(angles) > mirrorVAngleStdRad {
		return out
	}

	// === I2 中点共线 + 方向 ⊥ v ===
	mids := make([]radarutils.Point, len(samples))
	for i, s := range samples {
		mids[i] = radarutils.Point{X: (s.Ax + s.Bx) / 2, Y: (s.Ay + s.By) / 2}
	}
	dirX, dirY, rmse := fitLineDir(mids)
	if rmse > mirrorMidpointRMSECm {
		return out
	}
	// |cos(midline 方向, mean v)| 应≈0（正交）
	meanVX := math.Cos(circMean(angles))
	meanVY := math.Sin(circMean(angles))
	if math.Abs(dirX*meanVX+dirY*meanVY) > mirrorOrthogonalCos {
		return out
	}

	// === I3 同步等速 ===
	for i := 1; i < len(samples); i++ {
		dAx := float64(samples[i].Ax - samples[i-1].Ax)
		dAy := float64(samples[i].Ay - samples[i-1].Ay)
		dBx := float64(samples[i].Bx - samples[i-1].Bx)
		dBy := float64(samples[i].By - samples[i-1].By)
		sA := math.Hypot(dAx, dAy)
		sB := math.Hypot(dBx, dBy)
		max := math.Max(sA, sB)
		if max < 1 { // 这一帧两 track 都没动，跳过比例检查
			continue
		}
		if math.Abs(sA-sB)/max > mirrorSpeedRatioMax {
			return out
		}
	}

	// === Tiebreaker：距 radar 远者 = ghost ===
	meanDistA, meanDistB := 0.0, 0.0
	for _, s := range samples {
		meanDistA += math.Hypot(float64(s.Ax-radarX), float64(s.Ay-radarY))
		meanDistB += math.Hypot(float64(s.Bx-radarX), float64(s.By-radarY))
	}
	meanDistA /= float64(len(samples))
	meanDistB /= float64(len(samples))

	out.IsMirror = true
	if meanDistA > meanDistB {
		out.GhostTrackID, out.RealTrackID = idA, idB
	} else {
		out.GhostTrackID, out.RealTrackID = idB, idA
	}

	// 镜面线锚点 = 中点重心；方向 = fitLineDir 输出
	cx, cy := 0, 0
	for _, m := range mids {
		cx += m.X
		cy += m.Y
	}
	cx /= len(mids)
	cy /= len(mids)
	out.MidlinePoint = radarutils.Point{X: cx, Y: cy}
	out.MidlineDirX = dirX
	out.MidlineDirY = dirY

	// === Bounce points：R' = reflect(R, mirror_line)；M_t = 线段 R'-A_t ∩ mirror_line ===
	r := radarutils.Point{X: radarX, Y: radarY}
	rPrime := reflectPoint(r, out.MidlinePoint, dirX, dirY)
	for _, s := range samples {
		// real 端取 RealTrackID 对应的位置
		var ax, ay int
		if out.RealTrackID == idA {
			ax, ay = s.Ax, s.Ay
		} else {
			ax, ay = s.Bx, s.By
		}
		bp, ok := segmentIntersectLine(rPrime, radarutils.Point{X: ax, Y: ay}, out.MidlinePoint, dirX, dirY)
		if ok {
			out.BouncePoints = append(out.BouncePoints, bp)
		}
	}
	return out
}

// hasMinMove 检查任一 track 5 帧累计位移 ≥ minCm。
func hasMinMove(samples []mirrorPairSample, minCm int) bool {
	var sumA, sumB float64
	for i := 1; i < len(samples); i++ {
		sumA += math.Hypot(float64(samples[i].Ax-samples[i-1].Ax), float64(samples[i].Ay-samples[i-1].Ay))
		sumB += math.Hypot(float64(samples[i].Bx-samples[i-1].Bx), float64(samples[i].By-samples[i-1].By))
	}
	minF := float64(minCm)
	return sumA >= minF || sumB >= minF
}

// circStd 角度数组（rad，含 ±π 跨越）标准差。用 sin/cos 平均的方式避免边界跳跃。
func circStd(angles []float64) float64 {
	mx, my := 0.0, 0.0
	for _, a := range angles {
		mx += math.Cos(a)
		my += math.Sin(a)
	}
	mx /= float64(len(angles))
	my /= float64(len(angles))
	R := math.Hypot(mx, my) // 0..1：1 = 完全一致，0 = 均匀分布
	if R > 1 {
		R = 1
	}
	if R < 1e-9 {
		return math.Pi
	}
	return math.Sqrt(-2 * math.Log(R))
}

// circMean 角度数组循环均值（rad）。
func circMean(angles []float64) float64 {
	mx, my := 0.0, 0.0
	for _, a := range angles {
		mx += math.Cos(a)
		my += math.Sin(a)
	}
	return math.Atan2(my, mx)
}

// fitLineDir 给定 N 个 2D 点，PCA fit 主方向，返回 (dirX, dirY, rmse-to-line)。
// dir 是单位向量。rmse = 点到直线的均方根距离。
func fitLineDir(pts []radarutils.Point) (dirX, dirY, rmse float64) {
	n := float64(len(pts))
	if n < 2 {
		return 1, 0, 0
	}
	var mx, my float64
	for _, p := range pts {
		mx += float64(p.X)
		my += float64(p.Y)
	}
	mx /= n
	my /= n
	var sxx, syy, sxy float64
	for _, p := range pts {
		dx := float64(p.X) - mx
		dy := float64(p.Y) - my
		sxx += dx * dx
		syy += dy * dy
		sxy += dx * dy
	}
	// 主轴角 = 0.5 * atan2(2*sxy, sxx-syy)
	theta := 0.5 * math.Atan2(2*sxy, sxx-syy)
	dirX = math.Cos(theta)
	dirY = math.Sin(theta)
	// rmse = 残差均方根
	var ssq float64
	for _, p := range pts {
		dx := float64(p.X) - mx
		dy := float64(p.Y) - my
		// 法线方向投影 = 与 dir 正交方向上的距离
		nx, ny := -dirY, dirX
		d := dx*nx + dy*ny
		ssq += d * d
	}
	rmse = math.Sqrt(ssq / n)
	return
}

// reflectPoint 点 p 关于经过 anchor 方向 (dirX, dirY) 的直线的镜像。
func reflectPoint(p, anchor radarutils.Point, dirX, dirY float64) radarutils.Point {
	dx := float64(p.X - anchor.X)
	dy := float64(p.Y - anchor.Y)
	// 投影到 dir 上
	t := dx*dirX + dy*dirY
	footX := float64(anchor.X) + t*dirX
	footY := float64(anchor.Y) + t*dirY
	return radarutils.Point{
		X: int(math.Round(2*footX - float64(p.X))),
		Y: int(math.Round(2*footY - float64(p.Y))),
	}
}

// mirrorCandidate 内部使用：scanMirrorGhostPairs 收集的 live track 快照。
// 定义为 package-level 类型让 gcMirrorBuffer 能引用同一个类型。
type mirrorCandidate struct {
	ID   int
	X, Y int
}

// scanMirrorGhostPairs TrackManager 入口：每帧调用，扫所有"非 ghost & 非 anchored"
// track 两两组合，累配对样本；满 5 帧 + 通过三不变量 + tiebreaker → ghost 端 GhostPenalty +50
// 并把 bounce 点写 grid。调用方持锁（processFrameAt 内调用）。
func (tm *TrackManager) scanMirrorGhostPairs(nowMs int64) {
	// 收集候选 track：Verdict 已 anchored Real（LongSurvivalAnchored / StartupGrace）跳过 —
	// 两个都已确认 Real 不可能互为 mirror；任一未锚定即评估
	var cands []mirrorCandidate
	for tid, ts := range tm.tracks {
		// 已经判 Ghost 不参与配对（节省工作量；它本来就是 ghost）
		if ts.Verdict == VerdictGhost {
			continue
		}
		pxF, pyF := ts.Kalman.Position()
		cands = append(cands, mirrorCandidate{ID: tid, X: int(math.Round(pxF)), Y: int(math.Round(pyF))})
	}
	if len(cands) < 2 {
		// 无配对可能；保险清掉单 track 残留的旧 buffer
		tm.gcMirrorBuffer(cands)
		return
	}
	// 固定 ID 顺序（map 迭代不稳定）
	sort.Slice(cands, func(i, j int) bool { return cands[i].ID < cands[j].ID })

	radarX := tm.radarMount.Center.X
	radarY := tm.radarMount.Center.Y

	// 两两组合
	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			a := cands[i]
			b := cands[j]
			// 跳过两个都已 anchored Real 的 pair
			tsA := tm.tracks[a.ID]
			tsB := tm.tracks[b.ID]
			if (tsA.LongSurvivalAnchored || tsA.StartupGrace) &&
				(tsB.LongSurvivalAnchored || tsB.StartupGrace) {
				continue
			}

			key := newMirrorPairKey(a.ID, b.ID)
			buf, ok := tm.mirrorBuffer[key]
			if !ok {
				buf = &mirrorPairBuffer{}
				tm.mirrorBuffer[key] = buf
			}
			// 冷却中：还是要 push 样本以保持滑窗连续，但不触发评估 + paint + penalty
			buf.pushSample(mirrorPairSample{Ax: a.X, Ay: a.Y, Bx: b.X, By: b.Y, Ts: nowMs})
			if nowMs-buf.lastConfirmedMs < tm.mirrorCooldownMs && buf.lastConfirmedMs > 0 {
				continue
			}

			res := evaluateMirrorPair(buf.samples, a.ID, b.ID, radarX, radarY)
			if !res.IsMirror {
				continue
			}
			// 加 GhostPenalty 到 ghost 端
			ghost := tm.tracks[res.GhostTrackID]
			if ghost != nil {
				ghost.GhostPenalty += mirrorGhostPenaltyAdd
				if ghost.GhostPenalty > 100 {
					ghost.GhostPenalty = 100
				}
				tm.logger.Info("mirror_pair_detected",
					zap.Int("ghost_track_id", res.GhostTrackID),
					zap.Int("real_track_id", res.RealTrackID),
					zap.Int("penalty_added", mirrorGhostPenaltyAdd),
					zap.Int("ghost_penalty_now", ghost.GhostPenalty),
					zap.Int("bounce_points", len(res.BouncePoints)),
				)
			}
			// 把每个 bounce 点 paint 2×2 微块累 MBC
			for _, bp := range res.BouncePoints {
				tm.grid.MarkMirrorBounce(bp.X, bp.Y, nowMs)
			}
			buf.lastConfirmedMs = nowMs
		}
	}

	tm.gcMirrorBuffer(cands)
}

// gcMirrorBuffer 清掉已不在 tracks 集中的 pair buffer（防 dead key 累积）。
func (tm *TrackManager) gcMirrorBuffer(live []mirrorCandidate) {
	if len(tm.mirrorBuffer) == 0 {
		return
	}
	liveSet := make(map[int]struct{}, len(live))
	for _, c := range live {
		liveSet[c.ID] = struct{}{}
	}
	for k := range tm.mirrorBuffer {
		if _, okA := liveSet[k.A]; !okA {
			delete(tm.mirrorBuffer, k)
			continue
		}
		if _, okB := liveSet[k.B]; !okB {
			delete(tm.mirrorBuffer, k)
		}
	}
}

// segmentIntersectLine 线段 (p1, p2) 与经过 anchor 方向 (dirX, dirY) 的直线的交点。
// 若线段两端在直线同侧 → 不交，返回 ok=false。
func segmentIntersectLine(p1, p2, anchor radarutils.Point, dirX, dirY float64) (radarutils.Point, bool) {
	// 直线法向 (nx, ny) = (-dirY, dirX)
	nx, ny := -dirY, dirX
	// signed distance from p1, p2 to line
	d1 := float64(p1.X-anchor.X)*nx + float64(p1.Y-anchor.Y)*ny
	d2 := float64(p2.X-anchor.X)*nx + float64(p2.Y-anchor.Y)*ny
	if (d1 > 0 && d2 > 0) || (d1 < 0 && d2 < 0) {
		return radarutils.Point{}, false
	}
	if math.Abs(d1-d2) < 1e-9 {
		return radarutils.Point{}, false
	}
	t := d1 / (d1 - d2)
	x := float64(p1.X) + t*float64(p2.X-p1.X)
	y := float64(p1.Y) + t*float64(p2.Y-p1.Y)
	return radarutils.Point{X: int(math.Round(x)), Y: int(math.Round(y))}, true
}
