package roomengine

// track_manager_io.go — track_manager 生产发布层焊回（S0.c）：Xsensor replay 道裁掉 AI publish/verdict 输出腿，
// 生产必需（cardagg ghost 覆盖源 + 固件 Fall 即时转发）。逐字复原自旧 ws track_manager.go（git HEAD）。

import (
	"context"

	"owl-common/alarm"
	"owl-common/observation"
)

type AIPayload struct {
	DeviceAddr string // 源 sensor UUID（FK to devices.device_addr）
	RoomID     string

	// AI 写的观测（与上游 firmware/engine 同一 schema）。
	// AI 只填它要表达的字段，其它留零值 = "AI 对此无意见"。
	Track observation.Track

	// Source 数据生产者节点身份（WHO）。当前 TrackManager 全部派生自护工角色，
	// engine.publishAIMessage 默认填 cfg.AIPublish.Source（如 "AI.Caregiver01"）。
	// 本字段保留作未来多角色 override 钩子（例：同一 TrackManager 内派生健康风险
	// verdict 时可显式设为 "AI.Doctor01"）。空 = 用 engine 默认。
	Source string

	// Reason AI 派生决策的理由路径（WHY）。填本地 Reason* 路径常量；空 = 非 AI 派生。
	Reason string

	// Evidence 证据 KV map（score / penalty / context 等审计字段，下游不解析）。
	Evidence map[string]interface{}

	// Event 决策事件类型（审计用，写 sensor_decision_log.event）：verdict_change /
	// lostfall_pending / lostfall_cancel / lostfall_fire / lostfall_suppress。空 = 不是决策审计事件。
	Event string

	// EventStatus 事件生命周期阶段（"start" / "end" / "instant"）。空 = "instant"（默认）。
	// 用于 firmware 撤销链路：qinglan 收到 Initialization (last_pose=2/7) → forward end →
	// sensor 透传 → cardagg AlarmRouter 按 Registry[Fall/SittingOnGround].EndPolicy=AutoResolve 关 alarm。
	EventStatus string

	// IncidentMs 实际发生时刻 ms（推断类 fall 才有值）。
	// firmware Fall = 0（incident == alerted == nowMs）；silent/lost/still fall = last_active/still_box_start/leftbed/empty_since 等真实 incident 时刻。
	// 写到 fields["incident_ts_ms"]，cardagg AlarmRouter 当 alarm_events.triggered_at；nowMs 当 alerted_at。
	// 0 = 不写字段，下游回退 triggered_at = alerted_at = nowMs。
	IncidentMs int64
}

// CategoryTrackVerdict 是 track verdict 的 category 路由键（事件 TYPE，不是 verdict label）。
const CategoryTrackVerdict = "track_verdict"

// CategorySensorDecision 是 lost-fall 决策审计事件的 category（与 verdict 同流 ai:track:verdict:stream，
// iot 落 sensor_decision_log；cardagg override 只认 track_verdict，按 category 跳过本类）。
const CategorySensorDecision = "sensor_decision"

// Reason 常量本地定义——目前 wisefido-sensor 是唯一 producer，下游 cardagg /
// wisefido-data 透传字符串不解析。未来若出现第二个 AI producer（如健康风险
// 模块），再上移到 owl-common 共享。
//
// Source（"AI.Caregiver01" 等）由 engine 在 publishAIMessage 默认填入
// （取自 cfg.AIPublish.Source），TrackManager 不直接设置——避免节点身份
// 散落到业务代码。
const (
	// Reason: ghost 判定（track_verdict category）
	ReasonGhostPostReal = "ghost_post_real" // Real → Ghost 翻转（已确认 real 后 penalty 累积超阈）
	ReasonGhostPenalty  = "ghost_penalty"   // 主路径：累积 penalty ≥ 阈值（含 motion_symmetry / no_enter_pair）
	ReasonGhostLowScore = "ghost_low_score" // probation 期满 score 低于阈值

	// Reason: fall 判定（alarm category，4 个 fall 派生子类型）
	ReasonLostTrack            = "lost_track"             // track 异常消失（可能是真摔倒后失锁）
	ReasonStillInBathroom      = "still_in_bathroom"      // 浴室长时间静止 still-fall
	ReasonBedsideSilent        = "bedside_silent"         // LeftBed 后床边静止过久 R4
	ReasonSleepadRadarConflict = "sleepad_radar_conflict" // sleepad LeftBed + radar 仍在床

	// sensor_v2 PR-10 BathroomFallRules（§6.A）— SuitePerson 主语 fall：
	ReasonBathroomStill              = "bathroom_still"                        // §6.A.1 cell Toilet/Shower + Stand + 10/12min
	ReasonBathroomLongStatic         = "bathroom_long_static"                  // §6.A.2 90s grace + 任意位置 8min（非 Toilet/Shower）
	ReasonSuitePersonCompletelyLost  = "suite_person_completely_lost_no_ghost" // §6.A.3 最强档 bathroom 内活 track==0 ≥ 30s
	ReasonSuitePersonSilentWithGhost = "suite_person_silent_with_ghost_proxy"  // §6.A.3 次强档 SuitePerson static ≥ 7min

	// sensor_v2 PR-11 BedroomFallRules（§6.B）— SuitePerson 主语 bedroom fall：
	ReasonBedroomBedsideStatic = "bedroom_bedside_static" // §6.B.3 BedState→Vacant + 床边 ≤100cm 静止 ≥15min
	ReasonBedroomPersonSilent  = "bedroom_person_silent"  // §6.B.2 SuitePerson AnchorRoomType=bedroom + LastActiveMs > threshold
)

// AIPublisher PR-8 解耦：TrackManager 不直接持有 redis client，由 engine 实现接口注入。
type AIPublisher interface {
	PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64)
	PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64)
	// DeviceUIDHex 反查 device addr → hex MAC（log 双格式人眼可读）；缺失返回空字符串
	DeviceUIDHex(deviceAddr string) string
}

func (tm *TrackManager) SetAIPublisher(p AIPublisher) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.aiPublisher = p
}

func (tm *TrackManager) roomLedgerEmpty() bool {
	return tm.lastExitMs > tm.lastEnterMs
}

func (tm *TrackManager) RoomLedgerEmpty() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.roomLedgerEmpty()
}

func (tm *TrackManager) payloadFromTrack(ts *TrackState) AIPayload {
	conf := ts.DBNConfidence // S0.c-4 第三腿：DBN realness PReal 优先；<0 回退旧 ghost-penalty 派生
	if conf < 0 {
		conf = 100 - ts.GhostPenalty
	}
	if conf < 0 {
		conf = 0
	}
	if conf > 100 {
		conf = 100
	}
	return AIPayload{
		DeviceAddr: ts.DeviceAddr,
		RoomID:     ts.RoomID,
		Track: observation.Track{
			BedStatus:       observation.BedStatusUnchanged,
			TrackID:         ts.TrackID,
			LogicID:         ts.LogicID,
			PositionX:       intPtr(ts.LastRawH),
			PositionY:       intPtr(ts.LastRawV),
			PositionZ:       intPtr(ts.LastRawZ),
			Pose:            ts.LastPose,
			TrackConfidence: conf,
		},
	}
}

func (tm *TrackManager) emitAIEvent(p AIPayload, category string, nowMs int64) {
	if tm.aiPublisher == nil {
		return
	}
	tm.aiPublisher.PublishAIEvent(context.Background(), p, category, nowMs)
}

func (tm *TrackManager) emitAIAlarm(p AIPayload, category string, nowMs int64) {
	if tm.aiPublisher == nil {
		return
	}
	tm.aiPublisher.PublishAIAlarm(context.Background(), p, category, nowMs)
}

func (tm *TrackManager) emitGhostVerdict(ts *TrackState, reason, context string, nowMs int64) {
	p := tm.payloadFromTrack(ts)
	p.Reason = reason
	p.Evidence = map[string]interface{}{
		"score":         ts.Score,
		"birth_score":   ts.BirthScore,
		"ghost_penalty": ts.GhostPenalty,
		"frame_count":   ts.FrameCount,
	}
	if context != "" {
		p.Evidence["context"] = context
	}
	p.Event = "verdict_change"
	p.Evidence["verdict"] = int(ts.Verdict)
	tm.emitAIEvent(p, CategoryTrackVerdict, nowMs)
}

// forwardFirmwareFall 固件 Fall 即时转发腿（S0.c 焊回）：DBN 把固件 Fall 作 observation，
// 但生产仍需即时转 iot:alarm:stream（"宁可误报不可漏报"的固件 ground floor，DBN_MODE=0 也发）。
func (tm *TrackManager) forwardFirmwareFall(a RadarFallAlarm) {
	if tm.aiPublisher == nil {
		return // playback / 测试场景
	}
	emitStatus := a.Status
	if emitStatus == "" {
		emitStatus = "start"
	}
	payload := AIPayload{
		DeviceAddr:  a.DeviceUID,
		RoomID:      tm.roomID,
		EventStatus: emitStatus,
		Track: observation.Track{
			BedStatus: observation.BedStatusUnchanged,
			TrackID:   a.TrackID,
			Pose:      a.Pose,
		},
		Reason: "firmware_radar_fall",
		Evidence: map[string]interface{}{
			"context":         "qinglan_publisher_event_stream_passthrough",
			"firmware_status": a.Status,
		},
	}
	emitCat := a.Category
	if emitCat == "" {
		emitCat = alarm.Fall
	}
	tm.aiPublisher.PublishAIAlarm(context.Background(), payload, emitCat, a.TMs)
}

// trackByLogicID 按 LogicID 反查在管 TrackState（fire/drop 发布腿用）。调用方持 tm.mu。
func (tm *TrackManager) trackByLogicID(lid string) *TrackState {
	for _, ts := range tm.tracks {
		if ts.LogicID == lid {
			return ts
		}
	}
	return nil
}

// PublishDBNFall DBN fire 发布腿（S0.c-4，A·R12.3）：fired logicID → 构 payload → emitAIAlarm(alarm.Fall)。
// DBN_MODE 门控在 engine 内 dbnSelfFireEnabledFor（本方法只发，调用方已判门控）。
func (tm *TrackManager) PublishDBNFall(lid string, nowMs int64) {
	tm.mu.Lock()
	ts := tm.trackByLogicID(lid)
	var p AIPayload
	if ts != nil {
		p = tm.payloadFromTrack(ts)
	}
	tm.mu.Unlock()
	if ts == nil {
		return
	}
	p.Reason = "dbn_fall"
	tm.emitAIAlarm(p, alarm.Fall, nowMs)
}

// EmitDBNGhostVerdict DBN dropped 输出腿（S0.c-4，守卫① ghost 覆盖源不丢）：
// dropped(ghost/realness 抑制) logicID → emitGhostVerdict → ai:track:verdict:stream（cardagg ghost 覆盖）。
func (tm *TrackManager) EmitDBNGhostVerdict(lid string, nowMs int64) {
	tm.mu.Lock()
	ts := tm.trackByLogicID(lid)
	tm.mu.Unlock()
	if ts == nil {
		return
	}
	tm.emitGhostVerdict(ts, "dbn_dropped", "dbn_state_drop", nowMs)
}

// SetTrackConfidence 写回 DBN per-track 置信度（第三腿 A·R13/R15.3-2，不门控）：
// logicID → ts.TrackConfidence（0-100），供 payloadFromTrack 下发 cardagg 实时 DBN 置信度。
func (tm *TrackManager) SetTrackConfidence(lid string, conf int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if ts := tm.trackByLogicID(lid); ts != nil {
		ts.DBNConfidence = conf
	}
}

// confEmitThrottleMs: per-track DBN confidence 发布节流（20s，在 cardagg ai_override 60s TTL 内刷新，低频不 flood）。
const confEmitThrottleMs = 20_000

// EmitTrackConfidence 周期发 present track 的 DBN 置信度（track_verdict 事件，带 track_confidence）。
// 治本「confidence 第三腿只写回不发布」gap（S0.c-4 SetTrackConfidence 只写 ts.DBNConfidence）：
//
//	旧 publishTrackStatuses 删后，present ghost（低 PReal、未 drop）的低 conf 不发 → cardagg override 缺刷新 →
//	FE 渲不出 50% 透明。本方法对每个 present track 周期发 DBN conf → cardagg ai_override(release) 每帧盖 realtime card → FE 透明度。
//
// 节流 per-(device,track_id) 20s；不受 DBN_MODE 门控（informational 非 alarm，A·R15.3-1）。
func (tm *TrackManager) EmitTrackConfidence(lid string, nowMs int64) {
	tm.mu.Lock()
	ts := tm.trackByLogicID(lid)
	if ts == nil {
		tm.mu.Unlock()
		return
	}
	key := trackKey{ts.DeviceAddr, ts.TrackID}
	if nowMs-tm.confEmitMs[key] < confEmitThrottleMs {
		tm.mu.Unlock()
		return
	}
	tm.confEmitMs[key] = nowMs
	p := tm.payloadFromTrack(ts)
	tm.mu.Unlock()
	p.Reason = "dbn_confidence"
	p.Event = "confidence"
	tm.emitAIEvent(p, CategoryTrackVerdict, nowMs)
}
