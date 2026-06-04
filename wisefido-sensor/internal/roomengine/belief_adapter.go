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
	beliefLostTTLMs = 4_000 // track 超此无帧 = 丢失
	// P2 absence 发射的可达退场（reachable-exit）参数（doc/belief_p2_absence_emission §3）。
	beliefExitDistScaleCm  = 80   // f_dist = exp(-d/scale) 软距离尺度（≈1 步；替 30cm 硬 cutoff）
	beliefReportIntervalMs = 1000 // radar 1Hz 上报间隔 = reachability 的 Δt
	// A（走速来源）：丢失前 5s 窗口的**实测**定向逼近，而非 firmware 分钟聚合（时间错位 +
	// 循环：firmware 若有确信态就轮不到 lost_track）。窗口取 track History（30 帧滚动，丢失后不清）。
	beliefReachWindowMs = 5_000 // [loss-5s, last_frame] 逼近测速窗（对齐 trigger_at ±5s）
	// B（安全偏置）：逼近速度慢速封顶 = 老人保守（越弱越易摔且难恢复 → 宁少抑制少漏报）。
	// A 实测本设备运动已是"本设备值"最强形态；A 测不出 → v=0 → 不抑制（最安全，非假设 slow 走出去）。
	beliefReachSpeedCapCmS = 60 // P2.1 学习未足时的全局兜底封顶
	// P2.1：per-device 学习封顶。雷达自测本住户走速（非 PHI/非 age），个性化"可达天花板"——
	// 独居老人步态慢 → 封顶自然低 → 绝不给"他本可走出去"虚高信用 → 不漏报。
	beliefSpeedSampleWinMs   = 2_000 // 采样窗：DisplacementWithinMs(2s) 当一次走速观测
	beliefSpeedMoveMinCmS    = 15    // 低于此 = 站立/抖动，不采样（只学走动段）
	beliefSpeedEwmaAlpha     = 0.05  // EWMA 平滑（抗单帧量化噪声）
	beliefSpeedMinSamples    = 30    // 学习样本不足 → 用全局兜底（短 fixture 自然走兜底，不扰 oracle）
	beliefSpeedCapFactor     = 1.5   // 封顶 = 1.5×EWMA 走速（mean→ceiling 余量）
	beliefSpeedCapFloorCmS   = 30    // 封顶下限（防学得过低误压真退场）
	beliefSpeedCapCeilCmS    = 150   // 封顶上限（生理：挡噪声伪造超人速度）
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

// geomTrustFromSource cell Source → geom 信任权重（provenance）。只对抑制跌倒的 rest geom 有意义。
// FE 画(SourceHuman)=1.0 全信(保现有 full 抑制)；feedback=0.6；纯自学=0.4 —— 越暂定抑制越弱。
func geomTrustFromSource(s Source) float64 {
	switch s {
	case SourceFeedback:
		return 0.6
	case SourceLearned, SourceGeometry:
		return 0.4
	default: // SourceHuman / SourceUnset
		return 1.0
	}
}

// geomConfFromGrid 该 (x,y) geom 的 provenance 信任 [0,1]，喂 belief Observation.GeomConf。
// 仅 rest-suppress geom(InBed/InToilet)按 cell Source 打折；Enter/开阔地板/Unknown 全信(1.0)。
func geomConfFromGrid(g *RoomGrid, x, y int) float64 {
	if g == nil {
		return 1.0
	}
	if g.NearestEntryDist(x, y) <= beliefEnterMarginCm {
		return 1.0 // Enter 是几何，全信
	}
	c := g.CellAt(x, y)
	if c == nil || len(c.Belief) == 0 {
		return 1.0
	}
	switch c.Belief[0].Type {
	case AreaBed, AreaToilet, AreaShower: // 抑制跌倒的 rest geom → 按 provenance 给信任
		return geomTrustFromSource(c.Belief[0].Source)
	default:
		return 1.0
	}
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
	gc := geomConfFromGrid(grid, x, y) // geom provenance 信任（FE 画=1 / feedback=.6 / 自学=.4）

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
		{Source: t.LogicID, Kind: belief.ObsPose, Value: float64(t.Pose), Conf: poseConf, Ts: nowMs, Fresh: motionFresh, Geom: g, GeomConf: gc},
		{Source: t.LogicID, Kind: belief.ObsKinematics, Value: f, Conf: 0.7, Ts: nowMs, Fresh: motionFresh, Geom: g},
		{Source: t.LogicID, Kind: belief.ObsVitalPresent, Value: vitalVal, Conf: 0.6, Ts: nowMs, Fresh: motionFresh, Geom: g},
		{Source: t.LogicID, Kind: belief.ObsTrackPresent, Value: ghost, Conf: 0.8, Ts: nowMs, Fresh: tsFresh, Geom: g},
	}

	// 注：no-detect **不在逐帧 adapter 里发**。它是「走动中突然消失」的**消失事件**，由 engine/replay
	// 在 track 丢失时每 tick 调 noDetectObs 发出（前置=消失前 60s 在走动 / still-box<60s，见 replay 扫掠逻辑）。
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

// noDetectObs track 丢失（走动前置已满足）→ P(no-detect|s) 发射。每 tick 固定强度（无时长项，时长→P3）；
// exit/返回/reachableExit/np=0 经各自似然仲裁方向。geom 留作未来 geom 条件 no-detect（床/桶遮挡）扩展位。
func noDetectObs(g belief.Geom, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsNoDetect, Conf: 0.8, Ts: nowMs, Fresh: true, Geom: g}
}

// deviceSpeedStat per-device 走速学习（P2.1）。雷达自测本住户走动段速度的 EWMA → 个性化封顶。
// 非 PHI：来自 radar 观测，不涉 age/DOB。per-room shadow 持有（device→room 稳定），跨 track 生死累积。
type deviceSpeedStat struct {
	ewmaCmS float64
	samples int
}

// observe 喂一次走速样本（仅走动段，调用方已过滤站立/抖动）。
func (d *deviceSpeedStat) observe(s float64) {
	if s <= 0 {
		return
	}
	if d.samples == 0 {
		d.ewmaCmS = s
	} else {
		d.ewmaCmS += beliefSpeedEwmaAlpha * (s - d.ewmaCmS)
	}
	d.samples++
}

// capCmS 个性化可达封顶。样本不足 → 全局兜底；够则 1.5×EWMA 夹在 [floor, ceil]。
func (d *deviceSpeedStat) capCmS() float64 {
	if d == nil || d.samples < beliefSpeedMinSamples {
		return beliefReachSpeedCapCmS
	}
	c := d.ewmaCmS * beliefSpeedCapFactor
	if c < beliefSpeedCapFloorCmS {
		c = beliefSpeedCapFloorCmS
	}
	if c > beliefSpeedCapCeilCmS {
		c = beliefSpeedCapCeilCmS
	}
	return c
}

// sampleWalkSpeed 从 track 近 beliefSpeedSampleWinMs 的位移箱估一次走速（cm/s）；站立/抖动返回 0。
func sampleWalkSpeed(ts *TrackState, nowMs int64) float64 {
	if ts == nil {
		return 0
	}
	box := ts.DisplacementWithinMs(beliefSpeedSampleWinMs, nowMs)
	v := float64(box) / (float64(beliefSpeedSampleWinMs) / 1000)
	if v < beliefSpeedMoveMinCmS {
		return 0
	}
	return v
}

// approachSpeedTowardExit 估算 track 丢失前朝最近 Enter 的**定向逼近速度**（cm/s，≥0）。
// A：取 History 在 [nowMs-windowMs, nowMs] 的帧，算"窗口内离门最远点 − 末帧"的距离差 = 朝门净逼近量，
// 除以帧跨时。**定向**是命门：背门走 / 原地晃 / 倒地抽搐 → 逼近量 ≤0 → 返回 0（不抑制真跌倒）。
// 用 box 极值（最远→末帧）而非逐帧路径，对 firmware XY 量化抖动 robust（jitter 不灌水）。
// 离门距离用 grid.NearestEntryDist（已含"哪个门最近"+ 多边形几何）；速度按 capCmS（P2.1 个性化）封顶。
func approachSpeedTowardExit(ts *TrackState, grid *RoomGrid, windowMs, nowMs int64, capCmS float64) float64 {
	if ts == nil || grid == nil {
		return 0
	}
	cutoff := nowMs - windowMs
	firstTs, lastTs := int64(-1), int64(-1)
	maxDist, finalDist, count := -1, -1, 0
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		count++
		if firstTs < 0 {
			firstTs = p.TMs
		}
		lastTs = p.TMs
		d := grid.NearestEntryDist(p.X, p.Y)
		if d > maxDist {
			maxDist = d
		}
		finalDist = d // History 时序递增，末次循环即末帧
	}
	closingCm := float64(maxDist - finalDist) // 朝门逼近量
	if count < 2 || closingCm <= 0 {
		return 0 // 帧不足 / 没靠近门（背门、原地）→ 无逼近证据
	}
	spanSec := float64(lastTs-firstTs) / 1000
	if spanSec < 0.5 {
		spanSec = 0.5 // 防短窗放大
	}
	v := closingCm / spanSec
	if capCmS <= 0 {
		capCmS = beliefReachSpeedCapCmS
	}
	if v > capCmS {
		v = capCmS // B/P2.1：个性化慢速封顶
	}
	return v
}

// reachableExitScore 可达退场分 e = f_dist(d) · f_reach(v,d) ∈ [0,1]。
// C3 共享算子：Room 层 ObsReachableExit 与 Track 层 TObsReachableExit 同源调用，杜绝两层离场判别漂移。
// v = approachSpeedTowardExit（定向逼近，已慢速封顶）；v≤0 → f_reach=0 → e=0 → 不干预真跌倒。
// 替 30cm 硬门闸的悬崖（CABB 73cm 落悬崖外被误报；doc/belief_p2_absence_emission §3）。
func reachableExitScore(distCm int, approachSpeedCmS float64) float64 {
	if approachSpeedCmS <= 0 {
		return 0
	}
	d := float64(distCm)
	if d < 1 {
		d = 1
	}
	fDist := math.Exp(-d / beliefExitDistScaleCm)
	fReach := clampUnit(approachSpeedCmS * float64(beliefReportIntervalMs) / 1000 / d)
	return fDist * fReach
}

// reachableExitObs Room 层（S）可达退场观测，与 noDetectObs 同 tick 对冲。
func reachableExitObs(distCm int, approachSpeedCmS float64, g belief.Geom, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsReachableExit, Value: reachableExitScore(distCm, approachSpeedCmS), Conf: 0.8, Ts: nowMs, Fresh: true, Geom: g}
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
