// static_reflector.go — 静止金属反射体自学习（运动镜像 mirror_detect.go 之外的另一类）。
//
// 物理：金属支架（把手/淋浴架/毛巾杆）给雷达持续直射回波 → 钉死在固定坐标的假 track
// （CABB x=-95 把手 / 101 浴室同型）。这类点**人工标不了**——哪块金属产生幻影、幻影落在哪，
// 取决于雷达角度/材质/多径，非肉眼可预知 → 只能让 cell engine 从雷达自身行为自动学。
//
// 与 mirror_detect 分工：mirror 抓"跟真人同步运动"的反射对；本检测抓"静止直射"伪迹。
//
// 金属位三签名（用户 2026-06-04 定，全满足才记一次 candidate；Phase A 仅累计+log，不改 verdict）：
//   ① 出生即在此 ——BirthPos ≈ 当前位置（金属一出生就在金属点；**人是从门口出生走进来的**，
//                  BirthPos 在门口不在墙边 → 这条比 z≈0 更干净地排除"贴墙洗漱的真人"）
//   ② 从不移走   ——出生后一直没离开出生点，且存活够久（人迟早走开，金属永远不走）
//   ③ 近墙       ——金属支架螺墙上（有高度，故**不**卡 z≈0，否则漏墙上金属）
// 游走真人共存只作 log 信息**不 gate**（否则漏掉无人时纯金属幻影）。
// 跨独立 episode 累 ≥ StaticReflectorPromoteThreshold → 升 AreaDeny（Phase B）；
// demote 仍由 cell_learning hasRealActivity 兜底（真人真走过 → 退回 Unknown，误学自纠）。

package roomengine

import (
	"math"

	"owl-common/radarutils"

	"go.uber.org/zap"
)

const (
	staticReflectorConfineCm      = 50      // ① 当前位置到 BirthPos ≤ 此 = 没离开出生点（含 radar 抖动）
	staticReflectorMinLifeMs      = 300_000 // ② 出生后存活 ≥ 此仍未离开 = 久未移走（人会更早走开）
	staticReflectorWallMarginCm   = 40      // ③ 近墙：到 wall 多边形边 ≤ 此
	staticReflectorRoamMinCm      = 40      // 仅 log：另一 Real track 近 30s 位移 ≥ 此 = 游走真人
	staticReflectorMarkIntervalMs = 60_000  // 同一 track 去抖：每 60s 最多记一次（=独立 episode 计数）
)

// scanStaticReflectors 每帧调用（processFrameAt 内，调用方持锁）。Phase A：仅累计 + log，不改 verdict。
func (tm *TrackManager) scanStaticReflectors(nowMs int64) {
	if len(tm.wallPolygon) < 3 || tm.grid == nil {
		return
	}

	// 游走真人（仅作 log 信息，不 gate——无人时纯金属幻影也要能学）
	roamer := -1 // firmware track_id（仅 log；-1=无游走真人）
	for _, ts := range tm.tracks {
		if ts.DBNConfidence >= 50 && ts.DisplacementWithinMs(30_000, nowMs) >= staticReflectorRoamMinCm {
			roamer = ts.TrackID
			break
		}
	}

	for key, ts := range tm.tracks {
		pxF, pyF := ts.Kalman.Position()
		px, py := int(math.Round(pxF)), int(math.Round(pyF))
		// ① 出生即在此 + ② 从不移走：当前位置仍贴 BirthPos
		if distInt(ts.BirthPos.X, ts.BirthPos.Y, px, py) > staticReflectorConfineCm {
			continue // 已离开出生点 = 走动过 = 人
		}
		// ② 久未移走：出生后存活够久仍没走（人会更早走开）
		if nowMs-ts.BirthPos.TMs < staticReflectorMinLifeMs {
			continue
		}
		// ③ 近墙
		wallDist := distToWallEdgeCm(tm.wallPolygon, px, py)
		if wallDist > staticReflectorWallMarginCm {
			continue
		}
		if last, ok := tm.staticReflectorLastMark[key]; ok && nowMs-last < staticReflectorMarkIntervalMs {
			continue // 去抖
		}
		tm.staticReflectorLastMark[key] = nowMs
		tm.grid.MarkStaticReflector(px, py, nowMs)

		count := 0
		if c := tm.grid.CellAt(px, py); c != nil {
			count = c.StaticReflectorCount
		}
		tm.logger.Info("static_reflector_candidate", // Phase A：log-only，不改 verdict
			zap.Int("track_id", ts.TrackID),
			zap.Int("x", px), zap.Int("y", py),
			zap.Int("birth_x", ts.BirthPos.X), zap.Int("birth_y", ts.BirthPos.Y),
			zap.Int("wall_dist_cm", wallDist),
			zap.Int("age_sec", ts.AgeSec()),
			zap.Int("roamer_track_id", roamer), // -1 = 无游走真人（纯金属幻影）
			zap.Int("static_reflector_count", count),
			zap.Int("promote_threshold", StaticReflectorPromoteThreshold),
		)
	}
	tm.gcStaticReflectorMarks()
}

// gcStaticReflectorMarks 清掉已死 track 的去抖时戳。
func (tm *TrackManager) gcStaticReflectorMarks() {
	if len(tm.staticReflectorLastMark) == 0 {
		return
	}
	for tid := range tm.staticReflectorLastMark {
		if _, ok := tm.tracks[tid]; !ok {
			delete(tm.staticReflectorLastMark, tid)
		}
	}
}

// distToWallEdgeCm 点到 wall 多边形最近边的距离（cm）。多边形闭合（末点接首点）。
func distToWallEdgeCm(poly []radarutils.Point, x, y int) int {
	if len(poly) < 2 {
		return math.MaxInt32
	}
	best := math.MaxFloat64
	for i := 0; i < len(poly); i++ {
		a := poly[i]
		b := poly[(i+1)%len(poly)]
		d := pointToSegmentDistF(float64(x), float64(y), float64(a.X), float64(a.Y), float64(b.X), float64(b.Y))
		if d < best {
			best = d
		}
	}
	return int(math.Round(best))
}

// pointToSegmentDistF 点 (px,py) 到线段 (ax,ay)-(bx,by) 的欧氏距离。
func pointToSegmentDistF(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	if dx == 0 && dy == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}
