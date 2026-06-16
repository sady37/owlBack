// bathroom_gate.go — sensor_v2 §4.A.2 BathroomGate 入口流量子模块（PR-5）
//
// 职责：
//   - 监测某 bathroom 房间内 track 的成员状态翻转（新出现 / 长时间消失）
//   - 跨 0 边界时联动 SuiteCensus 翻转 sole-resident.AnchorRoomType（PR-5c）
//   - 不识别 person 身份（bathroom 不跑 UpdatePersonFromTrack — PR-3 决策）；
//     anchor flip 由 TryFlipSoleResidentRoomType 单 resident 兜底
//
// 设计选择：
//   1. 每 bathroom 房间一个 gate（key = roomID）。多 bathroom suite 各自维护成员表，
//      census.BathroomCount 是聚合 — AdjustBathroomCount delta 求和即正确。
//      trackID 是 radar-local，per-gate 维护避免跨 radar trackID 冲突。
//
//   2. 不区分"crossing entry cells"vs"appears anywhere in bathroom"——
//      bathroom 是单 room（radar 不会跨 room 切换），track 只能从 entry cells 入场（物理）。
//      cell.EnterTarget=="bathroom" 信号仅作为出生 cell 的 audit 字段，
//      帮助 PR-6 BathroomGhost 规则 1（出生地判据 / 远离 entry 出生 → ghost）。
//
//   3. 60s timeout exit — track 失锁 60s 后视为离开，与 trackLastSeen 失锁清理同阈值。
//      短于此（如 30s）firmware coast 期间会误触发；长于此 (如 5min) BathroomCount 准确性恶化。
//
// 不在 PR-5 范围（PR-X 留作 TODO）：
//   - 跨 room 兜底 -1（"bedroom 入口附近新生 track" 推断 bathroom 漏报出门）
//   - 多 bathroom suite 内 person 跨 gate 跟随（v3）
//   - 兜底 -1 cross-room signal correlation

package roomengine

import (
	"owl-common/card"

	"go.uber.org/zap"
)

// gateMemberTimeoutMs track 失锁阈值。超过此时长未在 bathroom 帧中再出现 → 视为离开。
// 与 Engine.trackLostAnchorMs 同 60s 选值（保 PR-3 trackLastSeen 清理与 gate 出场判定一致节奏）。
const gateMemberTimeoutMs int64 = 60_000

// BathroomGate 单 bathroom 房间的入口流量计数器。
//
// 状态：
//   members[trackID] = lastSeenMs   该 track 上次在 bathroom 帧中出现的时刻
//
// 并发：Process 由 Engine.publishTrackStatuses 单 goroutine 调用，无锁需要。
type BathroomGate struct {
	roomID  string
	suiteID string
	census  *SuiteCensusManager
	logger  *zap.Logger

	members map[int]int64
}

// NewBathroomGate 构造单个 bathroom 房间的 gate。
// census 必须非 nil（suite-bathroom 模式前提）；suiteID 必须非空。
// logger nil 时降级 zap.NewNop（单测构造）。
func NewBathroomGate(roomID, suiteID string, census *SuiteCensusManager, logger *zap.Logger) *BathroomGate {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BathroomGate{
		roomID:  roomID,
		suiteID: suiteID,
		census:  census,
		logger:  logger,
		members: make(map[int]int64),
	}
}

// Process 由 Engine.publishTrackStatuses 在 bathroom room 帧后调用。
// bases 已含 PR-3 cell area type / EnterTarget 投影。
//
// 算法：
//   1) 扫 bases：新 trackID（非 members）→ onEntry；已知 trackID → 更新 lastSeenMs
//   2) 扫 members：lastSeenMs < nowMs - gateMemberTimeoutMs 且未在 bases → onExit
//
// 不变量：
//   - census.BathroomCount = Σ over all gates in suite (len(gate.members))
//     （单 bathroom suite 退化为单 gate；多 bathroom 自然聚合）
//   - PublicBathroom census 不会被本函数触达（AdjustBathroomCount/TryFlip 都会返回 noop）
func (g *BathroomGate) Process(bases []TrackStatusBase, nowMs int64) {
	if g.census == nil || g.suiteID == "" {
		return
	}
	// Step 1: detect entries + refresh lastSeenMs。
	// `b := &bases[i]` 形式取元素地址（不是 range 循环变量地址）—
	// Go <1.22 range 变量复用导致 `&b` 拿到最后一次迭代的指针；用索引取址永远安全，
	// 也防止未来 onEntry 改成异步 / 缓存路径时踩坑。
	seen := make(map[int]struct{}, len(bases))
	for i := range bases {
		b := &bases[i]
		seen[b.TrackID] = struct{}{}
		if _, wasMember := g.members[b.TrackID]; !wasMember {
			g.onEntry(b, nowMs)
		}
		g.members[b.TrackID] = nowMs
	}
	// Step 2: timeout-based exits
	for tid, lastMs := range g.members {
		if _, alive := seen[tid]; alive {
			continue
		}
		if nowMs-lastMs < gateMemberTimeoutMs {
			continue
		}
		g.onExit(tid, lastMs, nowMs)
		delete(g.members, tid)
	}
}

// onEntry 处理 track 首次出现在 bathroom 帧：count++ + 跨 0 边界 anchor flip。
// cell.EnterTarget 信号作为 audit 字段写 log，PR-6 BathroomGhost 规则 1 用同源数据判定。
func (g *BathroomGate) onEntry(b *TrackStatusBase, nowMs int64) {
	oldCount, newCount := g.census.AdjustBathroomCount(g.suiteID, +1, nowMs)
	if oldCount < 0 {
		// suite 不存在 / public mode → noop，但仍要 audit
		g.logger.Warn("bathroom_gate_entry_noop",
			zap.String("suite_id", g.suiteID),
			zap.String("room_id", g.roomID),
			zap.Int("track_id", b.TrackID),
			zap.String("reason", "suite_missing_or_public"),
		)
		return
	}
	flipped := false
	if oldCount == 0 && newCount > 0 {
		flipped = g.census.TryFlipSoleResidentRoomType(g.suiteID, card.RoomTypeBathroom, nowMs)
	}
	g.logger.Info("bathroom_gate_entry",
		zap.String("suite_id", g.suiteID),
		zap.String("room_id", g.roomID),
		zap.Int("track_id", b.TrackID),
		zap.String("birth_cell_area", areaTypeWireName(b.CellAreaType)),
		zap.String("birth_enter_target", b.EnterTarget),
		zap.Int("old_count", oldCount),
		zap.Int("new_count", newCount),
		zap.Bool("anchor_flipped", flipped),
	)
}

// onExit 处理 track 超时离开：count-- + 跨 0 边界 anchor flip back。
func (g *BathroomGate) onExit(trackID int, lastSeenMs, nowMs int64) {
	oldCount, newCount := g.census.AdjustBathroomCount(g.suiteID, -1, nowMs)
	if oldCount < 0 {
		g.logger.Warn("bathroom_gate_exit_noop",
			zap.String("suite_id", g.suiteID),
			zap.String("room_id", g.roomID),
			zap.Int("track_id", trackID),
			zap.String("reason", "suite_missing_or_public"),
		)
		return
	}
	flipped := false
	if oldCount > 0 && newCount == 0 {
		flipped = g.census.TryFlipSoleResidentRoomType(g.suiteID, card.RoomTypeDefault, nowMs)
	}
	g.logger.Info("bathroom_gate_exit",
		zap.String("suite_id", g.suiteID),
		zap.String("room_id", g.roomID),
		zap.Int("track_id", trackID),
		zap.Int64("last_seen_ms", lastSeenMs),
		zap.Int64("idle_ms", nowMs-lastSeenMs),
		zap.Int("old_count", oldCount),
		zap.Int("new_count", newCount),
		zap.Bool("anchor_flipped", flipped),
	)
}

// MembersCount 当前 gate 跟踪的 track 数（unit-test / audit 用）。
func (g *BathroomGate) MembersCount() int {
	return len(g.members)
}

func areaTypeWireName(t AreaType) string {
	switch t {
	case AreaEnter:
		return "enter"
	case AreaBed:
		return "bed"
	case AreaSit:
		return "sit"
	case AreaActive:
		return "active"
	case AreaDeny:
		return "deny"
	case AreaShower:
		return "shower"
	case AreaToilet:
		return "toilet"
	}
	return "unknown"
}
