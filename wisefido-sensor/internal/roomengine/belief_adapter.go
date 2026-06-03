package roomengine

import (
	"math"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"wisefido-sensor/internal/roomengine/belief"

	"go.uber.org/zap"
)

// belief_adapter.go — §9 第 2 步：把各源 struct 翻成统一 belief.Observation（log-only，未接决策）。
// 引擎只见 Observation 不见原始 struct（belief_input_normalization.md §5）。
// 设计文档：owlBack/doc/belief_gate_to_matrix.md。

const (
	beliefRadarTTLMs      = 4_000   // radar 1Hz，超 4s 当 stale（normalization §4）
	beliefSleepadTTLMs    = 35_000  // sleepad vital 窗（对齐 bed scorer vitalWindowMs）
	beliefStillBoxStaleMs = 120_000 // 长冻 radar track → 命门 stale（治本 John.Y 9h）
	beliefEnterMarginCm   = 30      // 离门 ≤30cm 判 Enter 区（对齐 ExitDistMinCm）
	beliefFallDropRefCm   = 60      // z 骤降 60cm ≈ 满跌倒运动学签名
	beliefEventWindowMs   = 5_000   // room transition ts 在此窗内才当一次 Enter/Exit 观测
	// lost-fall（走动中突然消失）：消失前 still-box < MovingPreconditionMs(60s) = 走动中（对齐 gate-list 止血）。
	// 走速只用于 ghost-filter（>ImpossibleSpeedCm 的 track-swap 假跳先剔除），不做"走速判走动"——
	// radar 位置量化重（走动也常报 median=0），per-frame 走速不可靠，still-box(spread) 才 robust。
	beliefLostTTLMs         = 4_000 // track 超此无帧 = 丢失
	beliefLostWaitWindowSec = 30    // 丢失后 ObsLostWhileMoving ramp 到满龄的时长
	// P2 absence 发射的可达退场（reachable-exit）参数（doc/belief_p2_absence_emission §3）。
	beliefExitDistScaleCm  = 80   // f_dist = exp(-d/scale) 软距离尺度（≈1 步；替 30cm 硬 cutoff）
	beliefReportIntervalMs = 1000 // radar 1Hz 上报间隔 = reachability 的 Δt
	// mobility-tier 步速先验默认：行动不便 0.6 m/s = 最慢档 = 保守（少抑制、少漏真跌倒）。
	// 观测走速（固件 walk_distance/walk_duration）接入后覆盖此先验；per-resident tier 经 PHI 边界派生注入（§4）。
	beliefPriorWalkSpeedCmS = 60
)

// geomFromArea cell AreaType → belief.Geom。
func geomFromArea(a AreaType) belief.Geom {
	switch a {
	case AreaBed:
		return belief.GeomInBed
	case AreaEnter:
		return belief.GeomInEnter
	case AreaToilet, AreaShower:
		return belief.GeomInToilet
	case AreaSit, AreaActive, AreaDeny:
		return belief.GeomOpenFloor
	default:
		return belief.GeomUnknown
	}
}

// geomFromGrid device 坐标 ∩ layout → belief.Geom。近门优先判 Enter（离场二义性命门）。
func geomFromGrid(g *RoomGrid, x, y int) belief.Geom {
	if g == nil {
		return belief.GeomUnknown
	}
	if g.NearestEntryDist(x, y) <= beliefEnterMarginCm {
		return belief.GeomInEnter
	}
	if c := g.CellAt(x, y); c != nil && len(c.Belief) > 0 {
		return geomFromArea(c.Belief[0].Type)
	}
	return belief.GeomUnknown
}

// radarFrameAdapter 一帧 radar track → []Observation（pose / 运动学 / vital / track-present）。
// 命门：长冻 track 的 pose/kinematics/vital 置 Fresh=false（Conf=0 不更新），即便 LastObservedMs 仍新
// （firmware 冻结期持续 1Hz 推同帧）。ghost-ness 不受冻结影响（verdict 仍有效）。
func radarFrameAdapter(t observation.Track, ts *TrackState, grid *RoomGrid, nowMs int64) []belief.Observation {
	x, y, z := 0, 0, 0
	if t.PositionX != nil {
		x = *t.PositionX
	}
	if t.PositionY != nil {
		y = *t.PositionY
	}
	if t.PositionZ != nil {
		z = *t.PositionZ
	}
	g := geomFromGrid(grid, x, y)

	tsFresh := nowMs-ts.LastObservedMs <= beliefRadarTTLMs
	stillBox := ts.StillBoxRunStart > 0 && nowMs-ts.StillBoxRunStart >= beliefStillBoxStaleMs
	motionFresh := tsFresh && !stillBox

	poseConf := 0.7
	if t.PoseConfidence > 0 {
		poseConf = float64(t.PoseConfidence) / 100
	}

	// 跌倒运动学签名 f：z 相对上帧骤降 → [0,1]。v2 再叠位移/隐含速度。
	f := 0.0
	if dz := ts.LastZ - z; dz > 0 {
		f = clampUnit(float64(dz) / beliefFallDropRefCm)
	}

	vitalVal := 0.0
	if t.HeartRate > 0 || t.RespiratoryRate > 0 {
		vitalVal = 1
	}

	// ghost-ness：GhostPenalty(≥80→1) 与 verdict 合成。Ghost verdict 直接拉满。
	ghost := clampUnit(float64(ts.GhostPenalty) / 80)
	if ts.Verdict == VerdictGhost {
		ghost = 1
	}

	out := []belief.Observation{
		{Source: t.LogicID, Kind: belief.ObsPose, Value: float64(t.Pose), Conf: poseConf, Ts: nowMs, Fresh: motionFresh, Geom: g},
		{Source: t.LogicID, Kind: belief.ObsKinematics, Value: f, Conf: 0.7, Ts: nowMs, Fresh: motionFresh, Geom: g},
		{Source: t.LogicID, Kind: belief.ObsVitalPresent, Value: vitalVal, Conf: 0.6, Ts: nowMs, Fresh: motionFresh, Geom: g},
		{Source: t.LogicID, Kind: belief.ObsTrackPresent, Value: ghost, Conf: 0.8, Ts: nowMs, Fresh: tsFresh, Geom: g},
	}

	// 注：lost-fall **不在逐帧 adapter 里发**。它是「走动中突然消失」的**消失事件**，由 engine/replay
	// 在 track 丢失时调 lostWhileMovingToObs 发出（前置=消失前 60s 在走动 / still-box<60s，见 replay 扫掠逻辑）。
	// 不再用"静止时长 ramp"(判反)、不再用 pose 门控(radar pose 不可靠,2026-06-01 用户实证)。
	return out
}

// isGhostJump 帧间速 > ImpossibleSpeedCm(200cm/s) = ghost/track-swap 假跳（2米+ 瞬跳），
// 喂 still-box/belief 前应剔除（2026-06-01 用户实证：坐姿信号丢失/ghost swap 才出 2m+）。
func isGhostJump(dCm float64, dtMs int64) bool {
	if dtMs <= 0 {
		return false
	}
	return dCm*1000/float64(dtMs) > float64(FallRulesParam.Lost.ImpossibleSpeedCm)
}

// lostWhileMovingToObs track 丢失（走动前置已满足）→ 候选倒地观测。ageFrac=丢失后时长归一 [0,1]。
// engine/replay 在 track 消失后每 tick 调（ramp 抬 P(Fallen)），exit/返回经各自似然对冲取消。
func lostWhileMovingToObs(ageFrac float64, g belief.Geom, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsLostWhileMoving, Value: clampUnit(ageFrac), Conf: 0.8, Ts: nowMs, Fresh: true, Geom: g}
}

// reachableExitObs 丢失点的"可达退场"证据 e = f_dist(d) · f_reach(v,d) ∈ [0,1]，与 lostWhileMovingToObs
// 同 tick 发出对冲：e 高（近门 + 单帧可达）→ 偏 Left 压 Fallen；e≈0（远离门/不可达）→ identity 不干预真跌倒。
// 替 30cm 硬门闸的悬崖（CABB 73cm 落悬崖外被误报；doc/belief_p2_absence_emission §3）。
// speedCmS = 步速（观测走速‖mobility-tier 先验，缺则用 beliefPriorWalkSpeedCmS 兜底）。
func reachableExitObs(distCm int, speedCmS float64, g belief.Geom, nowMs int64) belief.Observation {
	d := float64(distCm)
	if d < 1 {
		d = 1
	}
	if speedCmS <= 0 {
		speedCmS = beliefPriorWalkSpeedCmS
	}
	fDist := math.Exp(-d / beliefExitDistScaleCm)
	fReach := clampUnit(speedCmS * float64(beliefReportIntervalMs) / 1000 / d)
	return belief.Observation{Kind: belief.ObsReachableExit, Value: fDist * fReach, Conf: 0.8, Ts: nowMs, Fresh: true, Geom: g}
}

// radarEventToObs 离散 radar 事件 → Observation（EnterRoom/ExitRoom/Fall）。
func radarEventToObs(eventName string, nowMs int64, g belief.Geom) (belief.Observation, bool) {
	switch eventName {
	case alarm.EnterRoom:
		return belief.Observation{Kind: belief.ObsEnterExit, Value: +1, Conf: 0.9, Ts: nowMs, Fresh: true, Geom: g}, true
	case alarm.ExitRoom:
		return belief.Observation{Kind: belief.ObsEnterExit, Value: -1, Conf: 0.9, Ts: nowMs, Fresh: true, Geom: g}, true
	case alarm.Fall:
		return belief.Observation{Kind: belief.ObsFirmwareFall, Value: 1, Conf: 0.9, Ts: nowMs, Fresh: true, Geom: g}, true
	}
	return belief.Observation{}, false
}

// sleepadAdapter sleepad 一帧 → []Observation（接触式 vital + 床占用）。
func sleepadAdapter(o SleepadObservation, nowMs int64) []belief.Observation {
	fresh := nowMs-o.TMs <= beliefSleepadTTLMs
	vitalVal := 0.0
	if o.HasVitalSign() {
		vitalVal = 1
	}
	bedVal := 0.0
	if o.InBed {
		bedVal = 1
	}
	return []belief.Observation{
		{Source: o.DeviceAddr, Kind: belief.ObsVitalPresent, Value: vitalVal, Conf: 0.9, Ts: o.TMs, Fresh: fresh, Geom: belief.GeomInBed},
		{Source: o.DeviceAddr, Kind: belief.ObsBedOccupied, Value: bedVal, Conf: 0.9, Ts: o.TMs, Fresh: fresh, Geom: belief.GeomInBed},
	}
}

// bedAdapter bed 贝叶斯 scorer 输出 → ObsBedOccupied（嵌套 belief 子证据）+ ObsSleepStage。
// BedConfidence(0/60/90) 透传为 Conf；0=无数据→不更新。
func bedAdapter(b card.BedState, nowMs int64) []belief.Observation {
	bedVal := 0.0
	if b.BedStatus == 0 { // 0=InBed
		bedVal = 1
	}
	out := []belief.Observation{
		{Source: "bed", Kind: belief.ObsBedOccupied, Value: bedVal, Conf: float64(b.BedConfidence) / 100, Ts: b.BedStatusTs, Fresh: b.BedConfidence > 0, Geom: belief.GeomInBed},
	}
	if b.SleepConfidence > 0 {
		out = append(out, belief.Observation{Source: "bed", Kind: belief.ObsSleepStage, Value: float64(b.SleepStage), Conf: float64(b.SleepConfidence) / 100, Ts: b.SleepStageTs, Fresh: true, Geom: belief.GeomInBed})
	}
	return out
}

// roomAdapter room 聚合状态 → []Observation。
// 铁律：AloneContinuousMin/RiskLevel 等派生信号禁入（feedback_no_dynamic_threshold_modulation）。
func roomAdapter(r card.RoomState, roomType int, nowMs int64) []belief.Observation {
	out := []belief.Observation{
		{Source: "room", Kind: belief.ObsNumberPeople, Value: float64(r.TotalPeople), Conf: 0.8, Ts: r.TotalPeopleTs, Fresh: true, Geom: belief.GeomUnknown},
	}
	standGeom := belief.GeomUnknown
	if roomType == card.RoomTypeBathroom {
		standGeom = belief.GeomInToilet
	}
	out = append(out, belief.Observation{Source: "room", Kind: belief.ObsStandDuration, Value: float64(r.StandingContinuousMin), Conf: 0.7, Ts: nowMs, Fresh: true, Geom: standGeom})

	// room transition（最近 5s 内的 0↔N 翻转当一次 Enter/Exit 观测）
	if r.LastExitTs > 0 && nowMs-r.LastExitTs <= beliefEventWindowMs {
		out = append(out, belief.Observation{Source: "room", Kind: belief.ObsEnterExit, Value: -1, Conf: 0.7, Ts: r.LastExitTs, Fresh: true, Geom: belief.GeomInEnter})
	} else if r.LastEnterTs > 0 && nowMs-r.LastEnterTs <= beliefEventWindowMs {
		out = append(out, belief.Observation{Source: "room", Kind: belief.ObsEnterExit, Value: +1, Conf: 0.7, Ts: r.LastEnterTs, Fresh: true, Geom: belief.GeomInEnter})
	}
	return out
}

// neighborToObs §5.5.2 弱耦合：邻居 room 占用信念 → 本房 ObsNeighbor。
// occ = 邻居 P(占用) [0,1]；conf = 邻居 belief 确定度。9h 治本近似：sleepad InBed 当 occ。
func neighborToObs(occ, conf float64, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsNeighbor, Value: clampUnit(occ), Conf: conf, Ts: nowMs, Fresh: true, Geom: belief.GeomUnknown}
}

// logBeliefObservations shadow 接线前的 log-only 出口：打一行 Observation 序列。
func logBeliefObservations(logger *zap.Logger, deviceAddr, roomID string, obs []belief.Observation) {
	if logger == nil {
		return
	}
	fresh := 0
	for _, o := range obs {
		if o.Fresh && o.Conf > 0 {
			fresh++
		}
	}
	logger.Debug("belief_observations",
		zap.String("device_addr", deviceAddr),
		zap.String("room_id", roomID),
		zap.Int("obs_count", len(obs)),
		zap.Int("effective_count", fresh),
	)
}

func clampUnit(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
