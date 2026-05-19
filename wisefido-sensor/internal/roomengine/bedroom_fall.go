// bedroom_fall.go — sensor_v2 §6.B BedroomFallRules（PR-11 11b + 11c）
//
// 2 类 bedroom 内 fall 触发，主语 = SuitePerson（决定 8）：
//
//   11b BedroomBedsideFall（§6.B.3）
//     BedState transition Occupied/Leaving → Vacant 后（= BedSession.LeftBedAtMs latched）
//     + Real track 在床边 ≤100cm 静止 ≥15min（夜间）
//     + RoomState.Count == 1（多人 unit 时放宽到 ≤2 但降级 SuspectedFall — PR-12 实现）
//
//   11c BedroomLostFall（§6.B.2）
//     存在 SuitePerson p where p.AnchorRoomType == "bedroom"
//     + p.LastActiveMs > cell-area-typed 时长（PR-11 baseline: 固定 10min, PR-X 接 cell-typed）
//     + 期间无 p 跨房间转移事件
//     不再用"任意 track lost"作为触发——决定 14 ghost 滞留不能压制告警
//
// 不在 PR-11 范围（11a silent_fall PR-9 后已对齐 §6.B.1，仍在 TrackManager.scanSilentFallLeftBed
// 内运行；§6.B.4 still_fall 决定 16 后 bedroom 路径不实现）。
//
// outRoom 事件不可靠（§6.A.0' 决定 — radar 盲区作 enter cells）：
//   - 11c 取消条件只接受正向证据（看到 person 活动 = 取消），不接受 outRoom event
//   - p.LastActiveMs 由 UpdatePersonFromTrack 在 moveActive=true 时更新，反映真实观测
//
// 阈值常量见底部；PR-12 引入 risk_factor 矩阵后这些值由配置决定。

package roomengine

import (
	"context"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"

	"go.uber.org/zap"
)

// PR-11 阈值常量（PR-12 引入 risk_factor 矩阵后可调）
const (
	bedroomBedsideMarginCm       = 100     // §6.B.3 床边距离阈值
	bedroomBedsideNightStaticSec = 15 * 60 // §6.B.3 夜间静止时长（PR-12 引入 risk_factor 后 day=22.5min）

	// §6.B.2 BedroomLostFall cell-typed 静默时长（PR-X 补强）。
	// 落到老人在不同 cell 上的合理"无动作"时长：床上 2h、沙发 30min、走道 5min、未知 10min。
	// 仅人工标定（SourceHuman）的 AreaBed（床/long-sofa 躺区）直接 exempt — 老人在床上
	// 久躺是 normal use case；坐姿/马桶/淋浴突然不动是真跌倒，不豁免。
	bedroomLostFallSilentSecBed     = 2 * 60 * 60 // AreaBed: 2h（学习/几何来源）
	bedroomLostFallSilentSecSit     = 30 * 60     // AreaSit: 30min
	bedroomLostFallSilentSecActive  = 5 * 60      // AreaActive: 5min（走道伫立）
	bedroomLostFallSilentSecDefault = 10 * 60     // 默认 / AreaUnknown: 10min
)

// BedroomFallRules 实现 §6.B 11b + 11c。
// 与 BathroomFallRules 同构（per-roomID 状态机，单 goroutine 调用），但触发条件不同。
type BedroomFallRules struct {
	census      *SuiteCensusManager
	gridLookup  func(roomID string) *RoomGrid
	suiteLookup func(roomID string) string
	aiPublisher AIPublisher
	timezone    *time.Location
	logger      *zap.Logger

	state map[string]*bedroomFallState
}

// bedroomFallState 单 bedroom room 的 fall 状态机内部状态。
type bedroomFallState struct {
	// 11b dedup: per-LeftBed-session（LeftBedAtMs 作 key，下次 LeftBed 时新 ts 允许再 fire）
	BedsideFiredForLeftBedAt int64
	// 11c dedup: personID → 已 fire；person 离开 bedroom (AnchorRoomType 变) 时 reset
	LostFallFired map[string]bool
	// LastBases sticky（同 PR-10 健康检查模式）
	LastBases    []TrackStatusBase
	EverObserved bool
}

// NewBedroomFallRules 构造。aiPublisher nil 时 fire 路径退化为 log-only。
func NewBedroomFallRules(
	census *SuiteCensusManager,
	gridLookup func(roomID string) *RoomGrid,
	suiteLookup func(roomID string) string,
	aiPublisher AIPublisher,
	logger *zap.Logger,
) *BedroomFallRules {
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("bedroom_fall_rules_initialized",
		zap.String("ruleset", "decision_8+14"),
		zap.Int("bedside_margin_cm", bedroomBedsideMarginCm),
		zap.Int("bedside_night_static_sec", bedroomBedsideNightStaticSec),
		zap.Int("lost_fall_silent_sec_bed", bedroomLostFallSilentSecBed),
		zap.Int("lost_fall_silent_sec_sit", bedroomLostFallSilentSecSit),
		zap.Int("lost_fall_silent_sec_active", bedroomLostFallSilentSecActive),
		zap.Int("lost_fall_silent_sec_default", bedroomLostFallSilentSecDefault),
		zap.String("silent_fall_path", "TrackManager.scanSilentFallLeftBed (PR-9 dependency cleanup)"),
	)
	return &BedroomFallRules{
		census:      census,
		gridLookup:  gridLookup,
		suiteLookup: suiteLookup,
		aiPublisher: aiPublisher,
		logger:      logger,
		state:       make(map[string]*bedroomFallState),
	}
}

// SetTimezone 注入 IANA 时区（用于夜间风险时段判定）。
func (r *BedroomFallRules) SetTimezone(loc *time.Location) {
	r.timezone = loc
}

// Evaluate 由 engine.publishTrackStatuses 在非-bathroom room 帧后调用。
// 调用方保证 room.kind != "bathroom"（bathroom 走 BathroomFallRules）。
//
// 参数 beds = TrackManager.SnapshotBedSessions（PR-11 新增）的快照；
// 11b BedsideFall 依赖 BedSession.LeftBedAtMs 作触发起点。
func (r *BedroomFallRules) Evaluate(roomID string, bases []TrackStatusBase, beds []BedSessionLatch, nowMs int64) {
	if roomID == "" {
		return
	}
	state := r.getOrCreateState(roomID)

	suiteID := ""
	if r.suiteLookup != nil {
		suiteID = r.suiteLookup(roomID)
	}
	var c *SuiteBedroomCensus
	if r.census != nil && suiteID != "" {
		c = r.census.Get(suiteID)
	}

	// 健康检查（同 PR-10 模式）
	if !r.isSystemHealthy(c, bases, state) {
		if len(bases) > 0 {
			state.LastBases = cloneBases(bases)
		}
		return
	}

	// 11b BedroomBedsideFall
	r.evaluateBedsideFall(bases, beds, roomID, suiteID, c, state, nowMs)

	// 11c BedroomLostFall
	r.evaluateLostFall(bases, beds, roomID, suiteID, c, state, nowMs)

	// LastBases sticky
	if len(bases) > 0 {
		state.LastBases = cloneBases(bases)
	}
}

func (r *BedroomFallRules) getOrCreateState(roomID string) *bedroomFallState {
	s, ok := r.state[roomID]
	if !ok {
		s = &bedroomFallState{
			LostFallFired: make(map[string]bool),
		}
		r.state[roomID] = s
	}
	return s
}

// isSystemHealthy 简化版 §6.A.0：census 存在 + radar 曾激活过。
// public bathroom 不会走到此函数（bathroom 分支自有 BathroomFallRules）；
// public-mode-flagged census 在 bedroom rule 里也不合理（不应当存在），直接 noop。
func (r *BedroomFallRules) isSystemHealthy(c *SuiteBedroomCensus, bases []TrackStatusBase, state *bedroomFallState) bool {
	if c == nil || c.IsPublicBathroom {
		return false
	}
	if len(bases) > 0 {
		state.EverObserved = true
	}
	return state.EverObserved
}

// hasSoleResidentInBedroom 返回 bedroom 内的 sole resident（AnchorRoomType=bedroom + Role=resident）；
// 0 个或多个（多 resident decision 19 留 PR-X）返回 nil。
// PR-11 baseline 单 resident 假设。
func (r *BedroomFallRules) hasSoleResidentInBedroom(c *SuiteBedroomCensus) *SuitePerson {
	if c == nil {
		return nil
	}
	var sole *SuitePerson
	for _, p := range c.Persons {
		if p.Role != SuitePersonResident {
			continue
		}
		if p.AnchorRoomType != card.RoomTypeDefault {
			continue
		}
		if sole != nil {
			return nil // 多 resident 场景 — 决定 19 留 PR-X
		}
		sole = p
	}
	return sole
}

// firstNonGhostNearBed 找第一个 cell 邻近 AreaBed ≤ 100cm + 非 ghost + 静止 ≥ 15min 的 track。
// 11b BedsideFall 触发 anchor 选择。
func (r *BedroomFallRules) firstNonGhostNearBed(bases []TrackStatusBase, grid *RoomGrid) *TrackStatusBase {
	if grid == nil {
		return nil
	}
	for i := range bases {
		b := &bases[i]
		if b.Verdict == VerdictGhost {
			continue
		}
		if b.StillSec < bedroomBedsideNightStaticSec {
			continue
		}
		if !grid.IsNearPriorType(b.X, b.Y, AreaBed, bedroomBedsideMarginCm) {
			continue
		}
		return b
	}
	return nil
}

// evaluateBedsideFall §6.B.3 BedroomBedsideFall。
// 依赖 BedSession.LeftBedAtMs（PR-9 仍保留 latch），不依赖 v1 lastLeftBedAt（PR-9 已删）。
//
// 触发：
//   1. 任一 BedSession 有 LeftBedAtMs > 0（最近一次离床事件）
//   2. 夜间风险时段（IsNightTime）
//   3. 存在 track 在床边 ≤ 100cm + 非 ghost + 静止 ≥ 15min
//
// dedup per-LeftBed-session：BedsideFiredForLeftBedAt 等于当前 LeftBedAtMs 时不再 fire；
// 下次 LeftBed 事件 latches 新 ts → 允许再 fire（不同 session 互不影响）。
func (r *BedroomFallRules) evaluateBedsideFall(
	bases []TrackStatusBase, beds []BedSessionLatch, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bedroomFallState, nowMs int64,
) {
	if !IsNightTime(nowMs, r.timezone) {
		return
	}
	// 取最近一次 LeftBedAtMs（多 BedSession 取最大）
	var latestLeftBed int64
	for i := range beds {
		if beds[i].LeftBedAtMs > latestLeftBed {
			latestLeftBed = beds[i].LeftBedAtMs
		}
	}
	if latestLeftBed == 0 {
		return // 从未观测到 LeftBed
	}
	if state.BedsideFiredForLeftBedAt == latestLeftBed {
		return // 本 session 已 fire 过
	}
	var grid *RoomGrid
	if r.gridLookup != nil {
		grid = r.gridLookup(roomID)
	}
	anchor := r.firstNonGhostNearBed(bases, grid)
	if anchor == nil {
		return
	}
	// 人工标定躺区豁免：track 实际就在 AreaBed cell 上 = 人在床上不动（sleepad LeftBed 矛盾归
	// silent_fall 路径处理；此处不应再 fire bedside）。详 fall_exempt.go。
	if isHumanBedAt(grid, anchor.X, anchor.Y) {
		r.logger.Debug("bedroom_bedside_fall_exempt_human_bed",
			zap.String("room_id", roomID),
			zap.Int("track_id", anchor.TrackID),
		)
		return
	}
	r.fireFall(anchor, roomID, suiteID, ReasonBedroomBedsideStatic, map[string]interface{}{
		"context":         "bedside_static_after_leftbed",
		"still_sec":       anchor.StillSec,
		"timeout_sec":     bedroomBedsideNightStaticSec,
		"margin_cm":       bedroomBedsideMarginCm,
		"leftbed_at_ms":   latestLeftBed,
	}, nowMs)
	state.BedsideFiredForLeftBedAt = latestLeftBed
}

// evaluateLostFall §6.B.2 BedroomLostFall。
//
// 触发：sole resident in bedroom + LastActiveMs > cell-typed 静默阈值。
//
// **Cell-typed 阈值 (PR-X)**：anchor cell 的 Belief[0].Type 决定 idle 阈值——
// AreaBed 2h / AreaSit 30min / AreaActive 5min / 未知 10min。
// 仅 SourceHuman + AreaBed（人工标定的躺区，床/long-sofa）直接 exempt — 老人在床上久躺
// 是 normal use case，不报。坐姿/马桶/淋浴突然静止反而是真跌倒高发场景，不豁免。
//
// **Sleepad 主源 gating (PR-11.1)**：当任一 BedSession 处于 active in-bed 状态
// （InBedSinceMs > 0 AND LeftBedAtMs == 0），意味着 sleepad 床压传感器当前判定有人在床 ——
// 此时 LastActiveMs 不更新是预期行为（睡眠静止），fall 必须抑制。即使 cell-typed AreaBed
// exempt 已覆盖大多数床场景，sleepad gating 仍作第二道保险（覆盖 cell 未被标床但人在床上的情况）。
//
// 取消（dedup reset）：person.AnchorRoomType 变（走到 bathroom）→ LostFallFired flag 重置；
//   下次 person 回 bedroom 静默时允许再 fire。
func (r *BedroomFallRules) evaluateLostFall(
	bases []TrackStatusBase, beds []BedSessionLatch, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bedroomFallState, nowMs int64,
) {
	person := r.hasSoleResidentInBedroom(c)
	// 释放已离开 bedroom 的 person 的 dedup flag
	for pid := range state.LostFallFired {
		if person != nil && person.PersonID == pid {
			continue
		}
		delete(state.LostFallFired, pid)
	}
	if person == nil {
		return
	}
	if state.LostFallFired[person.PersonID] {
		return
	}
	if person.LastActiveMs == 0 {
		return // 数据完整性问题，等下一次 active 更新
	}

	// 取 anchor：bedroom 内任一 non-ghost track 当作位置载体；无则用 LastBases 头
	var anchor *TrackStatusBase
	for i := range bases {
		if bases[i].Verdict != VerdictGhost {
			anchor = &bases[i]
			break
		}
	}
	if anchor == nil && len(state.LastBases) > 0 {
		anchor = &state.LastBases[0]
	}

	// Cell 查询：决定 exempt + cell-typed 阈值。anchor 缺失时退化为 default 阈值。
	silentSec := bedroomLostFallSilentSecDefault
	var (
		cellAreaType AreaType = AreaUnknown
		cellSource   Source   = SourceUnset
	)
	if anchor != nil && r.gridLookup != nil {
		if grid := r.gridLookup(roomID); grid != nil {
			if cell := grid.CellAt(anchor.X, anchor.Y); cell != nil {
				bel := cell.Belief[0]
				cellAreaType, cellSource = bel.Type, bel.Source
				// 人工标定的躺区 — 总闸豁免（详 fall_exempt.go）
				if isHumanBedAt(grid, anchor.X, anchor.Y) {
					r.logger.Debug("bedroom_lost_fall_exempt_human_bed",
						zap.String("room_id", roomID),
						zap.String("person_id", person.PersonID),
					)
					return
				}
				silentSec = lostFallSilentSecForArea(bel.Type)
			}
		}
	}

	idleMs := nowMs - person.LastActiveMs
	if idleMs < int64(silentSec)*1000 {
		return
	}

	// Sleepad 主源 gating：当前任一 BedSession active in-bed → 抑制（决定 14 共生律的睡眠场景表达）
	if hasActiveBedSession(beds) {
		r.logger.Debug("bedroom_lost_fall_suppressed_by_sleepad",
			zap.String("room_id", roomID),
			zap.String("person_id", person.PersonID),
			zap.Int64("idle_ms", idleMs),
			zap.String("reason", "active_bed_session"),
		)
		return
	}

	if anchor == nil {
		// 兜底：无 track 几何信息 — 仅 log，下游 cardagg 反查 device 拿位置
		r.logger.Warn("bedroom_lost_fall_no_anchor",
			zap.String("room_id", roomID),
			zap.String("person_id", person.PersonID),
			zap.Int64("idle_ms", idleMs),
		)
		return
	}
	r.fireFall(anchor, roomID, suiteID, ReasonBedroomPersonSilent, map[string]interface{}{
		"context":        "suite_person_silent_in_bedroom",
		"person_id":      person.PersonID,
		"person_role":    person.Role,
		"idle_ms":        idleMs,
		"timeout_sec":    silentSec,
		"cell_area_type": int(cellAreaType),
		"cell_source":    int(cellSource),
		"last_active_ms": person.LastActiveMs,
	}, nowMs)
	state.LostFallFired[person.PersonID] = true
}

// lostFallSilentSecForArea 按 cell AreaType 返回 lost_fall 静默阈值（秒）。
// 学习/几何来源 cell 使用；SourceHuman + 休息区已在 evaluateLostFall exempt 不到此。
func lostFallSilentSecForArea(t AreaType) int {
	switch t {
	case AreaBed:
		return bedroomLostFallSilentSecBed
	case AreaSit:
		return bedroomLostFallSilentSecSit
	case AreaActive:
		return bedroomLostFallSilentSecActive
	}
	return bedroomLostFallSilentSecDefault
}

// hasActiveBedSession 任一 BedSession 处于 active in-bed 状态（PR-11.1 sleepad 主源 gating）。
// 语义：sleepad 收到 InBed 后但未收到 LeftBed → 当前认为人在床上。
// 不检查时戳新鲜度（PR-X 加 staleness guard 防 sleepad 失联误抑制）。
func hasActiveBedSession(beds []BedSessionLatch) bool {
	for i := range beds {
		if beds[i].InBedSinceMs > 0 && beds[i].LeftBedAtMs == 0 {
			return true
		}
	}
	return false
}

// fireFall 统一 fire 入口。aiPublisher nil 时 log-only（dev / playback 模式安全降级）。
func (r *BedroomFallRules) fireFall(
	b *TrackStatusBase, roomID, suiteID, reason string,
	evidence map[string]interface{}, nowMs int64,
) {
	if evidence == nil {
		evidence = map[string]interface{}{}
	}
	evidence["suite_id"] = suiteID
	evidence["room_id"] = roomID
	r.logger.Info("bedroom_fall_fired",
		zap.String("room_id", roomID),
		zap.String("suite_id", suiteID),
		zap.String("reason", reason),
		zap.Int("track_id", b.TrackID),
		zap.String("device_id", b.DeviceID),
		zap.Int("x", b.X), zap.Int("y", b.Y), zap.Int("z", b.Z),
		zap.Int64("ts_ms", nowMs),
	)
	if r.aiPublisher == nil {
		return
	}
	payload := AIPayload{
		DeviceID: b.DeviceID,
		RoomID:   roomID,
		Track: observation.Track{
			TrackID:   b.TrackID,
			PositionX: intPtr(b.X),
			PositionY: intPtr(b.Y),
			PositionZ: intPtr(b.Z),
			Pose:      b.Pose,
		},
		Reason:   reason,
		Evidence: evidence,
	}
	r.aiPublisher.PublishAIAlarm(context.Background(), payload, alarm.Fall, nowMs)
}
