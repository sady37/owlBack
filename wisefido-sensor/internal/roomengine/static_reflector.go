// static_reflector.go — 静止金属反射体自学习（运动镜像 mirror_detect.go 之外的另一类）。
//
// 物理：金属把手/淋浴架/毛巾杆等固定金属，给雷达持续的直射回波 → 一个钉死在固定坐标的假 track
// （CABB x=-95 把手 / 101 浴室同型）。这类点**人工标不了**——哪块金属产生幻影、幻影落在哪，
// 取决于雷达角度/材质/多径，非肉眼可预知 → 只能让 cell engine 从雷达自身行为自动学。
//
// 与 mirror_detect 的分工：mirror 抓**跟着真人同步运动**的反射对；本检测抓**静止直射**伪迹。
//
// 学习判据（全部满足才记一次 candidate，Phase A 仅累计 + log，不改 verdict）：
//   1. 近墙          ——金属支架螺墙上；真人不会长期贴墙里，真摔在开阔地板（强安全判据）
//   2. 长期静止       ——StillBox 起点 ≥ staticReflectorMinStillMs（金属永远不动）
//   3. 有游走真人共存  ——另一条 Real track 在别处移动 = 真人在别处，本静止 track 才更可能是幻影
//                       （独居者自己贴墙静止时无 corroborator → 不学，防误学洗漱位）
// **不**用 z≈0 判据：金属支架(把手/淋浴架)多装在墙上有高度，反射点 z 可能 >0，卡 z≈0 会漏。
// 防"贴墙洗漱真人"误学靠 corroborator + 跨多次独立 episode 复现 + demote 兜底。
// 跨独立 episode 累 ≥ StaticReflectorPromoteThreshold → 升 AreaDeny（Phase B）；
// demote 仍由 cell_learning hasRealActivity 兜底（真人真走过 → 自动退回 Unknown）。

package roomengine

import (
	"math"

	"owl-common/radarutils"

	"go.uber.org/zap"
)

const (
	staticReflectorMinStillMs     = 120_000 // 长期静止门槛（2min）
	staticReflectorWallMarginCm   = 40      // 近墙：到 wall 多边形边 ≤ 此
	staticReflectorRoamMinCm      = 40      // corroborator：另一 Real track 近 30s 位移 ≥ 此 = 游走真人
	staticReflectorMarkIntervalMs = 60_000  // 同一 track 去抖：每 60s 最多记一次（=独立 episode 计数）
)

// scanStaticReflectors 每帧调用（processFrameAt 内，调用方持锁）。Phase A：仅累计 + log，不改 verdict。
func (tm *TrackManager) scanStaticReflectors(nowMs int64) {
	if len(tm.wallPolygon) < 3 || tm.grid == nil {
		return // 无墙多边形无法判近墙
	}

	// corroborator：是否存在一条游走的真人 track
	roamer := -1
	for tid, ts := range tm.tracks {
		if ts.Verdict != VerdictReal {
			continue
		}
		if ts.DisplacementWithinMs(30_000, nowMs) >= staticReflectorRoamMinCm {
			roamer = tid
			break
		}
	}
	if roamer < 0 {
		return // 无游走真人 → 不学（避免把独居者贴墙静止位误学成金属）
	}

	for tid, ts := range tm.tracks {
		if tid == roamer {
			continue
		}
		if ts.StillBoxRunStart == 0 || nowMs-ts.StillBoxRunStart < staticReflectorMinStillMs {
			continue // 非长期静止
		}
		pxF, pyF := ts.Kalman.Position()
		px, py := int(math.Round(pxF)), int(math.Round(pyF))
		wallDist := distToWallEdgeCm(tm.wallPolygon, px, py)
		if wallDist > staticReflectorWallMarginCm {
			continue // 不近墙
		}
		if last, ok := tm.staticReflectorLastMark[tid]; ok && nowMs-last < staticReflectorMarkIntervalMs {
			continue // 去抖
		}
		tm.staticReflectorLastMark[tid] = nowMs
		tm.grid.MarkStaticReflector(px, py, nowMs)

		count := 0
		if c := tm.grid.CellAt(px, py); c != nil {
			count = c.StaticReflectorCount
		}
		tm.logger.Info("static_reflector_candidate", // Phase A：log-only，不改 verdict
			zap.Int("track_id", tid),
			zap.Int("x", px), zap.Int("y", py),
			zap.Int("wall_dist_cm", wallDist),
			zap.Int64("still_sec", (nowMs-ts.StillBoxRunStart)/1000),
			zap.Int("roamer_track_id", roamer),
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
