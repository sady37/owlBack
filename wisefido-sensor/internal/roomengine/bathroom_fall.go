// bathroom_fall.go — sensor_v2 §6.A BathroomFallRules（PR-10）
//
// 4 类 bathroom 内 fall 触发，主语 = SuitePerson（决定 8 / 14），不是任意 track：
//
//   10a BathroomStillFall（§6.A.1）
//     cell.AreaType ∈ {Toilet, Shower} + pose=Stand + 静止 ≥ 10/12 min（风险/非风险）
//
//   10b BathroomBedsideFall（§6.A.2）
//     bathroom 内任意位置（非 Toilet/Shower）静止 ≥ 8 min，
//     90s grace（BathroomCount 0→positive 后）+ 决定 18：真多人不 fire
//
//   10c BathroomLostFall 最强档（§6.A.3）
//     BathroomCount ≥ 1 且 bathroom 内活 track 数 == 0 持续 ≥ 30s
//     → 真人完全失锁（连 ghost 反射源都没有 = 极端异常）
//
//   10d BathroomLostFall 次强档（§6.A.3）
//     存在 SuitePerson w/ AnchorRoomType=bathroom + bathroom 内仍有 track（不论 verdict）
//     + 任一 track 静止 ≥ 7 min
//     → SuitePerson 静默但 ghost 间接证据还在
//
// Public bathroom 适配（§6.A 顶部 + 决定 24）：
//   - 无 SuitePerson → 改用"bathroom 内 disambiguate 真人 track"作触发主体
//   - census.IsPublicBathroom=true 时 SuitePerson 查询 fallback "any verdict ∈ {Real, Anchored, Pending} track"
//   - Fall 归因仅 location，PersonID 留空
//
// 系统健康检查（§6.A.0 简化版）：
//   - SuiteCensus 存在（PR-2 LoadFromRedis 后非空 = census 在 1h 内被更新过 proxy）
//   - bases 非空 OR 上一帧 bases 非空（radar 还在汇报）
//   - 不在维护窗口（PR-10 阶段不接入维护窗口检测，留 PR-X）

package roomengine

import (
	"context"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"

	"go.uber.org/zap"
)

// 阈值常量（决定 17 + 19 + risk-time）
const (
	bathroomStillFallRiskSec   = 10 * 60 // §6.A.1 风险时段（夜间 23:30-07:30）10 min
	bathroomStillFallNormalSec = 12 * 60 // §6.A.1 非风险时段 12 min
	bathroomBedsideGraceSec    = 90      // §6.A.2 BathroomCount 0→positive 后 90s grace
	bathroomBedsideStaticSec   = 8 * 60  // §6.A.2 静止 8 min
	bathroomLostStrongEmptySec = 30      // §6.A.3 最强档 bathroom 内 0 track 持续 30s
	bathroomLostWeakSilentSec  = 7 * 60  // §6.A.3 次强档 track 静止 7 min
)

// BathroomFallRules 实现 §6.A 4 类 fall 规则。
//
// 状态语义：
//   - state per-roomID（每 bathroom room 独立状态机）
//   - 单 goroutine 读写（engine.publishTrackStatuses 串行调用），无锁
//   - 重启全清（与 BathroomGate 同 in-memory 哲学）
type BathroomFallRules struct {
	census      *SuiteCensusManager
	gridLookup  func(roomID string) *RoomGrid
	suiteLookup func(roomID string) string
	aiPublisher AIPublisher
	timezone    *time.Location // 风险时段判定；nil 时退化 UTC
	logger      *zap.Logger

	state map[string]*bathroomFallState
}

// bathroomFallState 单 bathroom room 的 fall 状态机内部状态。
type bathroomFallState struct {
	// 10c 最强档：bathroom 内 0 track 起点 ms（0 = 当前非空）
	EmptySinceMs int64
	// 10c dedup：当前 empty 期是否已 fire 过；新 track 进入或 EmptySinceMs reset 时一并重置
	LostStrongFired bool
	// 10b dedup：per-bathroom-session（不是 per-track）。
	// 值 = 已 fire 的那一段 occupied session 起点 BathroomOccupiedSinceMs；0 = 当前 session 未 fire 过。
	// session 重启（occupied → vacant → occupied）时新起点 != 上次 fire 起点 → 允许再 fire。
	BedsideFiredAtSessionMs int64
	// 10a dedup：trackID → 已 fire（track 离场后释放）
	StillAlerted map[int]bool
	// 10d dedup：personID → 已 fire；occupied session 结束时一并 reset
	LostWeakAlerted map[string]bool
	// 上一次"非空"观测的 bases 快照（10c 最强档 fire 时作"最后已知位置"凭据）。
	// 关键：空 bases 帧不覆盖 LastBases，保持 sticky 直到下次非空观测。
	LastBases []TrackStatusBase
	// 10b grace 起点：BathroomCount 从 0 翻 positive 的时刻（per room 单值）
	BathroomOccupiedSinceMs int64
	// 是否曾观测到 bases（系统健康检查 proxy；之前 LastBases 长度判定误把 reset 当未观测）
	EverObserved bool
}

// NewBathroomFallRules 构造。aiPublisher nil 时所有 fire 路径退化为 log-only（dev/playback 模式）。
//
// **bootstrap-wiring 依赖审计**：本规则只覆盖 Critical fall（4 类）。Warning-level 兜底
// （bathroom Stay > 10min）由 zonealarm.Supervisor 提供，独立 wire 在 zoneengine 子系统。
// v2 main.go bootstrap 必须同时调用 wiring.NewSubsystem（zoneengine + zonealarm）和
// SetBathroomFallRules（PR-10）—— 缺任一层都会留漏报盲点：
//   - 缺 zoneengine：bathroom 内 Critical 漏报（如 still_fall 阈值未到）时无 Warning 兜底
//   - 缺 PR-10：bathroom 内 Critical 全部失效，仅靠 Warning 一档（10min 才触发，太晚）
// 启动后通过 "bathroom_fall_rules_initialized" log 行可验证 PR-10 层已 wire。
func NewBathroomFallRules(
	census *SuiteCensusManager,
	gridLookup func(roomID string) *RoomGrid,
	suiteLookup func(roomID string) string,
	aiPublisher AIPublisher,
	logger *zap.Logger,
) *BathroomFallRules {
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("bathroom_fall_rules_initialized",
		zap.String("ruleset", "decision_14+8+18"),
		zap.Int("still_risk_sec", bathroomStillFallRiskSec),
		zap.Int("still_normal_sec", bathroomStillFallNormalSec),
		zap.Int("bedside_grace_sec", bathroomBedsideGraceSec),
		zap.Int("bedside_static_sec", bathroomBedsideStaticSec),
		zap.Int("lost_strong_empty_sec", bathroomLostStrongEmptySec),
		zap.Int("lost_weak_silent_sec", bathroomLostWeakSilentSec),
		zap.String("warning_fallback_dependency", "zonealarm.Supervisor Stay rule (10min, must be wired separately)"),
	)
	return &BathroomFallRules{
		census:      census,
		gridLookup:  gridLookup,
		suiteLookup: suiteLookup,
		aiPublisher: aiPublisher,
		logger:      logger,
		state:       make(map[string]*bathroomFallState),
	}
}

// SetTimezone 注入 IANA 时区（用于风险时段判定）。nil = UTC（错位风险，bootstrap 必设）。
func (r *BathroomFallRules) SetTimezone(loc *time.Location) {
	r.timezone = loc
}

// Evaluate 由 engine.publishTrackStatuses 在 bathroom room 帧后调用（room.kind=="bathroom" 才入）。
// bases 含当前帧所有 live track 的 Layer 1 投影（PR-3 含 StillSec / CellAreaType / Verdict）；
// 允许空 bases（10c 最强档需要扫上帧 anchor 判 empty-timeout）。
//
// 不变量：
//   - 调用方保证仅 bathroom room 入此函数 + 传入正确 roomID（不靠 bases[0].RoomID 推断）
//   - census / lookups 都允许 nil（noop fallback，方便单测）
//   - 单 goroutine 调用（per-room 串行），状态无锁读写
func (r *BathroomFallRules) Evaluate(roomID string, bases []TrackStatusBase, nowMs int64) {
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

	// §6.A.0 简化系统健康检查
	if !r.isSystemHealthy(c, bases, state, nowMs) {
		state.LastBases = cloneBases(bases)
		return
	}

	// 维护 BathroomOccupiedSinceMs（10b grace 计时基）
	r.updateOccupiedSince(c, state, nowMs)

	// 跑 4 条规则（执行顺序：先 trackID-level 的 still / bedside，再 room-level 的 lost-strong /
	// person-level 的 lost-weak — 同帧多 fire 由各自 dedup 保证不重）
	r.evaluateStillFall(bases, roomID, suiteID, c, state, nowMs)
	r.evaluateBedsideFall(bases, roomID, suiteID, c, state, nowMs)
	r.evaluateLostFallStrong(bases, roomID, suiteID, c, state, nowMs)
	r.evaluateLostFallWeak(bases, roomID, suiteID, c, state, nowMs)

	// 释放已失锁 track 的 dedup flag（下次该 trackID 复用时 fire 不被旧 flag 抑制）
	r.releaseDedupOnLostTracks(bases, state)

	// LastBases sticky：只在"非空"观测时覆盖。空帧保留上一次非空快照，让 10c lost-strong
	// 在 timeout 后仍能取 anchor，不会因 cloneBases(nil) 而失忆。
	if len(bases) > 0 {
		state.LastBases = cloneBases(bases)
	}
}

func (r *BathroomFallRules) getOrCreateState(roomID string) *bathroomFallState {
	s, ok := r.state[roomID]
	if !ok {
		s = &bathroomFallState{
			StillAlerted:    make(map[int]bool),
			LostWeakAlerted: make(map[string]bool),
		}
		r.state[roomID] = s
	}
	return s
}

// isSystemHealthy 简化版 §6.A.0：census 存在 + bathroom 历史曾观测到 track（radar 已激活过）。
// PR-X 接入维护窗口 / 心跳健康度时扩展。
func (r *BathroomFallRules) isSystemHealthy(c *SuiteBedroomCensus, bases []TrackStatusBase, state *bathroomFallState, nowMs int64) bool {
	if c == nil {
		return false
	}
	if len(bases) > 0 {
		state.EverObserved = true
	}
	if !state.EverObserved {
		// radar 从未上线（启动后未见任何 track）；不当 fall
		return false
	}
	return true
}

// updateOccupiedSince 维护 BathroomOccupiedSinceMs：BathroomCount 0→positive 时 reset，positive→0 时清零。
// 10b grace 周期判定用；session 结束时一并 reset 10b/10d dedup（让下次 occupied 重新可 fire）。
func (r *BathroomFallRules) updateOccupiedSince(c *SuiteBedroomCensus, state *bathroomFallState, nowMs int64) {
	if c == nil {
		return
	}
	if c.BathroomCount > 0 && state.BathroomOccupiedSinceMs == 0 {
		state.BathroomOccupiedSinceMs = nowMs
	}
	if c.BathroomCount == 0 && state.BathroomOccupiedSinceMs > 0 {
		state.BathroomOccupiedSinceMs = 0
		state.BedsideFiredAtSessionMs = 0
		state.LostWeakAlerted = make(map[string]bool)
	}
}

// hasSuitePersonInBathroom 触发主语判定：
//   - suite mode：census 内存在 Role==resident（或 visitor）+ AnchorRoomType==bathroom 的 person
//   - public mode：bathroom 内存在 verdict ∈ {Real, Anchored, Pending} 的 track（不识别身份）
func (r *BathroomFallRules) hasSuitePersonInBathroom(c *SuiteBedroomCensus, bases []TrackStatusBase) bool {
	if c == nil {
		return false
	}
	if c.IsPublicBathroom {
		for i := range bases {
			if bases[i].Verdict != VerdictGhost {
				return true
			}
		}
		return false
	}
	for _, p := range c.Persons {
		if p.AnchorRoomType == card.RoomTypeBathroom && p.Role != "" {
			return true
		}
	}
	return false
}

// firstSuitePersonInBathroom 取 bathroom 内第一个 person（按 map 迭代顺序，单 resident 场景唯一）。
// public mode 返回 nil（无 person 概念）。
func (r *BathroomFallRules) firstSuitePersonInBathroom(c *SuiteBedroomCensus) *SuitePerson {
	if c == nil || c.IsPublicBathroom {
		return nil
	}
	for _, p := range c.Persons {
		if p.AnchorRoomType == card.RoomTypeBathroom && p.Role != "" {
			return p
		}
	}
	return nil
}

// stillFallTimeoutSec 风险时段 vs 非风险时段阈值切换（§6.A.1）。
func (r *BathroomFallRules) stillFallTimeoutSec(nowMs int64) int {
	if IsNightTime(nowMs, r.timezone) {
		return bathroomStillFallRiskSec
	}
	return bathroomStillFallNormalSec
}

// evaluateStillFall §6.A.1 BathroomStillFall。
// trackID 维度 dedup；同 track 长时间静止只 fire 一次直到位置/状态变。
func (r *BathroomFallRules) evaluateStillFall(
	bases []TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bathroomFallState, nowMs int64,
) {
	if !r.hasSuitePersonInBathroom(c, bases) {
		return
	}
	timeoutSec := r.stillFallTimeoutSec(nowMs)
	for i := range bases {
		b := &bases[i]
		if state.StillAlerted[b.TrackID] {
			continue
		}
		if b.CellAreaType != AreaToilet && b.CellAreaType != AreaShower {
			continue
		}
		if b.Pose != observation.PoseStanding {
			continue
		}
		if b.StillSec < timeoutSec {
			continue
		}
		r.fireFall(b, roomID, suiteID, c, ReasonBathroomStill, map[string]interface{}{
			"context":     "bathroom_still_in_toilet_shower",
			"still_sec":   b.StillSec,
			"timeout_sec": timeoutSec,
			"risk_time":   IsNightTime(nowMs, r.timezone),
			"cell_area":   areaTypeWireName(b.CellAreaType),
		}, nowMs-int64(b.StillSec)*1000, nowMs)
		state.StillAlerted[b.TrackID] = true
	}
}

// evaluateBedsideFall §6.A.2 BathroomBedsideFall — 90s grace + 任意位置（非 Toilet/Shower）静止 8min。
// 决定 18：真多人（BathroomCount ≥ 2 且非"1 真人+1 ghost"）不 fire。
//
// dedup 语义（per-bathroom-session）：每次 occupied session 内最多 fire 1 次；
// 占用主语翻 vacant 后 session 重启允许再 fire（updateOccupiedSince 内 reset 已处理）。
// 不像 10a per-track dedup：bedside fall 是"bathroom 状态层告警"不是"track 状态告警"。
func (r *BathroomFallRules) evaluateBedsideFall(
	bases []TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bathroomFallState, nowMs int64,
) {
	if c == nil {
		return
	}
	if !r.hasSuitePersonInBathroom(c, bases) {
		return
	}
	if state.BathroomOccupiedSinceMs == 0 {
		return
	}
	if (nowMs - state.BathroomOccupiedSinceMs) < int64(bathroomBedsideGraceSec)*1000 {
		return
	}
	// per-session dedup
	if state.BedsideFiredAtSessionMs == state.BathroomOccupiedSinceMs {
		return
	}
	if r.suppressedByMultiResident(c, bases) {
		return
	}
	// 找第一个 non-ghost track 满足 cell≠Toilet/Shower + StillSec≥8min 作为告警 anchor。
	// ghost 不作触发主体（决定 14 ghost 是真人间接证据但不是"静止主体"）。
	var anchor *TrackStatusBase
	for i := range bases {
		b := &bases[i]
		if b.Verdict == VerdictGhost {
			continue
		}
		if b.CellAreaType == AreaToilet || b.CellAreaType == AreaShower {
			continue
		}
		if b.StillSec < bathroomBedsideStaticSec {
			continue
		}
		anchor = b
		break
	}
	if anchor == nil {
		return
	}
	r.fireFall(anchor, roomID, suiteID, c, ReasonBathroomLongStatic, map[string]interface{}{
		"context":        "bathroom_long_static_outside_toilet_shower",
		"still_sec":      anchor.StillSec,
		"timeout_sec":    bathroomBedsideStaticSec,
		"grace_sec":      bathroomBedsideGraceSec,
		"occupied_since": state.BathroomOccupiedSinceMs,
		"cell_area":      areaTypeWireName(anchor.CellAreaType),
	}, nowMs-int64(anchor.StillSec)*1000, nowMs)
	state.BedsideFiredAtSessionMs = state.BathroomOccupiedSinceMs
}

// suppressedByMultiResident §6.A.2 决定 18：BathroomCount ≥ 2 且非"1 真人 + 1 ghost"配对时抑制 fall。
// 共生律 disambiguate（PR-8 镜像学习）实施前的简化版：
//   - count == 1 → 不抑制
//   - count == 2 + bases 内 ≥ 1 ghost → 不抑制（1 真人 + 1 ghost 配对）
//   - count >= 2 + 0 ghost → 抑制（真多人陪同）
//   - count >= 3 → 一律抑制
func (r *BathroomFallRules) suppressedByMultiResident(c *SuiteBedroomCensus, bases []TrackStatusBase) bool {
	if c == nil || c.BathroomCount <= 1 {
		return false
	}
	if c.BathroomCount >= 3 {
		return true
	}
	// count == 2: 看 bases 是否含 ghost
	ghosts := 0
	for i := range bases {
		if bases[i].Verdict == VerdictGhost {
			ghosts++
		}
	}
	return ghosts == 0
}

// evaluateLostFallStrong §6.A.3 最强档 — bathroom 内活 track 数 == 0 持续 ≥ 30s + BathroomCount ≥ 1。
func (r *BathroomFallRules) evaluateLostFallStrong(
	bases []TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bathroomFallState, nowMs int64,
) {
	// 维护 EmptySinceMs
	if len(bases) > 0 {
		state.EmptySinceMs = 0
		state.LostStrongFired = false
		return
	}
	if state.EmptySinceMs == 0 {
		state.EmptySinceMs = nowMs
		return
	}
	if state.LostStrongFired {
		return
	}
	if (nowMs - state.EmptySinceMs) < int64(bathroomLostStrongEmptySec)*1000 {
		return
	}
	// 必须有人在 bathroom（入口流量证据）
	// suite mode：BathroomCount ≥ 1；public mode 同样按 BathroomCount（其语义在 public 模式下来自内部 disambiguate）
	if c == nil || c.BathroomCount < 1 {
		return
	}
	// 用"最后一次观测的 track"位置作 evidence
	evidence := map[string]interface{}{
		"context":          "bathroom_empty_after_count_positive",
		"empty_since_ms":   state.EmptySinceMs,
		"empty_duration_s": (nowMs - state.EmptySinceMs) / 1000,
		"bathroom_count":   c.BathroomCount,
	}
	var anchor *TrackStatusBase
	if len(state.LastBases) > 0 {
		anchor = &state.LastBases[0]
	}
	r.fireFallNoTrack(anchor, roomID, suiteID, c, ReasonSuitePersonCompletelyLost, evidence, state.EmptySinceMs, nowMs)
	state.LostStrongFired = true
}

// evaluateLostFallWeak §6.A.3 次强档 — SuitePerson 在 bathroom + bathroom 内仍有 track + 任一 track static ≥ 7min。
func (r *BathroomFallRules) evaluateLostFallWeak(
	bases []TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, state *bathroomFallState, nowMs int64,
) {
	if c == nil {
		return
	}
	person := r.firstSuitePersonInBathroom(c)
	// public mode person==nil 但 hasSuitePersonInBathroom 仍可能 true（按"任意非 ghost track"判）；
	// 弱档同样适用 public — 但 dedup key 没法 per-person，public 模式用 roomID-anon 作 key
	personID := "public_anon"
	if person != nil {
		personID = person.PersonID
	} else if c.IsPublicBathroom {
		if !r.hasSuitePersonInBathroom(c, bases) {
			return
		}
	} else {
		return
	}

	if state.LostWeakAlerted[personID] {
		return
	}
	// 找任一 track static ≥ 7min（保守不挑 verdict — 决定 14 ghost 不抑制 fall）
	var staticTrack *TrackStatusBase
	for i := range bases {
		if bases[i].StillSec >= bathroomLostWeakSilentSec {
			staticTrack = &bases[i]
			break
		}
	}
	if staticTrack == nil {
		return
	}
	r.fireFall(staticTrack, roomID, suiteID, c, ReasonSuitePersonSilentWithGhost, map[string]interface{}{
		"context":     "suite_person_static_with_ghost_proxy",
		"person_id":   personID,
		"still_sec":   staticTrack.StillSec,
		"timeout_sec": bathroomLostWeakSilentSec,
	}, nowMs-int64(staticTrack.StillSec)*1000, nowMs)
	state.LostWeakAlerted[personID] = true
}

// releaseDedupOnLostTracks 当 track 不再出现在当前帧 → 释放 10a per-track dedup flag。
// 同 trackID 在未来某帧复用时（firmware track_id 复用）能重新走完整 fire 流程。
// 10b dedup 是 per-session 不是 per-track，不在此释放（updateOccupiedSince 处理）。
func (r *BathroomFallRules) releaseDedupOnLostTracks(bases []TrackStatusBase, state *bathroomFallState) {
	live := make(map[int]struct{}, len(bases))
	for i := range bases {
		live[bases[i].TrackID] = struct{}{}
	}
	for tid := range state.StillAlerted {
		if _, alive := live[tid]; !alive {
			delete(state.StillAlerted, tid)
		}
	}
}

// fireFall 用单 track 信息 fire alarm（10a/10b/10d 路径）。
// incidentMs：实际发生时刻 ms（推断类必传，写 alarm_events.triggered_at）；nowMs = 决策时刻 → alerted_at。
func (r *BathroomFallRules) fireFall(
	b *TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, reason string, evidence map[string]interface{}, incidentMs, nowMs int64,
) {
	r.fireFallCore(b, roomID, suiteID, c, reason, evidence, incidentMs, nowMs)
}

// fireFallNoTrack 用 anchor（上帧 track）信息 fire alarm（10c 路径，当前帧无 track）。
// anchor nil 时仅 log，不实发 alarm（无几何信息），避免 cardagg 端 LPM 反查失败。
func (r *BathroomFallRules) fireFallNoTrack(
	anchor *TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, reason string, evidence map[string]interface{}, incidentMs, nowMs int64,
) {
	if anchor == nil {
		r.logger.Warn("bathroom_fall_no_anchor",
			zap.String("room_id", roomID),
			zap.String("reason", reason),
			zap.Int64("ts_ms", nowMs),
		)
		return
	}
	r.fireFallCore(anchor, roomID, suiteID, c, reason, evidence, incidentMs, nowMs)
}

// fireFallCore 统一 fire 入口。aiPublisher nil 时 log-only（dev / playback 模式安全降级）。
func (r *BathroomFallRules) fireFallCore(
	b *TrackStatusBase, roomID, suiteID string,
	c *SuiteBedroomCensus, reason string, evidence map[string]interface{}, incidentMs, nowMs int64,
) {
	if evidence == nil {
		evidence = map[string]interface{}{}
	}
	evidence["suite_id"] = suiteID
	evidence["room_id"] = roomID
	if c != nil {
		evidence["bathroom_count"] = c.BathroomCount
		evidence["is_public_bathroom"] = c.IsPublicBathroom
	}
	r.logger.Info("bathroom_fall_fired",
		zap.String("room_id", roomID),
		zap.String("suite_id", suiteID),
		zap.String("reason", reason),
		zap.Int("track_id", b.TrackID),
		zap.String("device_addr", b.DeviceAddr),
		zap.Int("canvas_x", b.X), zap.Int("canvas_y", b.Y), zap.Int("canvas_z", b.Z),
		zap.Int("raw_h", b.RawH), zap.Int("raw_v", b.RawV), zap.Int("raw_z", b.RawZ),
		zap.Int64("ts_ms", nowMs),
	)
	if r.aiPublisher == nil {
		return
	}
	payload := AIPayload{
		DeviceAddr: b.DeviceAddr,
		RoomID:     roomID,
		Track: observation.Track{
			TrackID:   b.TrackID,
			PositionX: intPtr(b.RawH),
			PositionY: intPtr(b.RawV),
			PositionZ: intPtr(b.RawZ),
			Pose:      b.Pose,
		},
		Reason:     reason,
		Evidence:   evidence,
		IncidentMs: incidentMs,
	}
	r.aiPublisher.PublishAIAlarm(context.Background(), payload, alarm.Fall, nowMs)
}

// cloneBases 浅拷贝（snapshot for LastBases 持有）；TrackStatusBase 全值类型字段，浅拷即够。
func cloneBases(bases []TrackStatusBase) []TrackStatusBase {
	if len(bases) == 0 {
		return nil
	}
	out := make([]TrackStatusBase, len(bases))
	copy(out, bases)
	return out
}
