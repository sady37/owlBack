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

	// P3.1 shadow realness 三探测器阈(委员会审查⑦:独立于生产 Verdict;室内-老人 shadow 占位,P9.6 待 oracle)。
	// **生产闸 SuspectSpeedCm/ImpossibleSpeedCm 一律不动**(R0)。超阈→近确定 ghost(P(瞬移|real elderly)≈0)。
	beliefIndoorSpeedCeilCmS = 120.0 // ①空间跳跃(本帧 Δ/dt)+ ③隐含速度(全程平均 dist/age) 室内天花板 cm/s
	beliefShadowKalmanResid  = 100.0 // ②Kalman 峰值残差(innovation 距离 cm):预测偏离>1m=急变 ghost(正常 ~20cm)

	// P3.2 冻结伪迹复合签名门控阈(委员会 review③:判 ghost = A∧(B≥2);shadow 占位,P9.6 待 oracle)。
	// A=近必要前置(跳变出生 ∨ cell=AreaDeny);B=③pose/z锁死 ④钉死小区 ⑤距门远 中 ≥2。真人远角久站缺 A→不判。
	beliefFrozenPoseZLockTicks = 5     // ③pose&z 连续 N 帧恒定(伪迹无生理变化;判别力弱,仅作 B 佐证)
	beliefFrozenSpreadCm       = 50    // ④30s box 散布 ≤ 此 = 钉死小区(真人微抖也小,故仅佐证非独证)
	beliefFrozenWindowMs       = 30000 // ④散布窗
	beliefFrozenDoorFarCm      = 100   // ⑤距最近门 > 此 = 距门远(排除门口正常驻留)

	// P3.3 记忆 L_R filter(承审查⑨:二值检测→连续 P(real) log-odds 带遗忘 γ;shadow 占位,P9.6 待 oracle)。
	// 摔倒瞬间 v≈0 无当下证据 → realness 靠摔前走路累积的 L_R 经 γ 带进倒地窗(cabb-0605)。
	// **B′ 标定(审查⑬/⑪ γ 风险)**:γ 改**时间基 per-sec**(原 0.9/帧≈6.6s 半衰过快,cabb-0605 躺52s 必塌)。
	// 0.99/s ≈ bed_scorer leak 0.55/min 量纲(半衰~69s);52s 躺保留 0.99^52≈59% → realness 存活。终值 P9 用 cabb-0605 标。
	beliefRealnessDecayPerSec = 0.99 // realness log-odds 每秒遗忘率(帧率无关;cabb-0605 52s 窗存活)
	beliefRealnessLOCap       = 5.0  // log-odds 截断 [-5,5]
	lnLRMoveReal              = 0.69 // ln2:每走动帧 +realness(MoveActive,XY 派生,R5-safe)
	lnLRJumpGhost             = 2.94 // ln19:P3.1 跳变/急变 近确定 ghost(强 −,每帧施加不靠衰减)
	lnLRFrozenGhost           = 2.94 // ln19:P3.2 冻结伪迹 近确定 ghost(强 −,每帧施加)

	// P3.4 recapture 软恢复:曾丢失(lost-fall ramping)的 track 返回 ≥ 此 = self-rescue candidate
	// (跌后自救可能,不硬 cancel 抹真摔)。shadow 占位(cd2b 5.85min 返回合格),P9.6 待 oracle。
	beliefSelfRescueMinGapMs = 60_000
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
	beliefSpeedSampleWinMs = 2_000 // 采样窗：DisplacementWithinMs(2s) 当一次走速观测
	beliefSpeedMoveMinCmS  = 15    // 低于此 = 站立/抖动，不采样（只学走动段）
	beliefSpeedEwmaAlpha   = 0.05  // EWMA 平滑（抗单帧量化噪声）
	beliefSpeedMinSamples  = 30    // 学习样本不足 → 用全局兜底（短 fixture 自然走兜底，不扰 oracle）
	beliefSpeedCapFactor   = 1.5   // 封顶 = 1.5×EWMA 走速（mean→ceiling 余量）
	beliefSpeedCapFloorCmS = 30    // 封顶下限（防学得过低误压真退场）
	beliefSpeedCapCeilCmS  = 150   // 封顶上限（生理：挡噪声伪造超人速度）

	beliefNumberPeopleTTLMs = 70_000 // 审查51:number_people event 新鲜窗(firmware 分钟级 push,>1min stale 不喂 ObsNumberPeople)
)

// isOpenFloorArea cell area∈{Sit,Active,Deny} = 开阔地板（非 rest-zone 非门）。
func isOpenFloorArea(a AreaType) bool {
	return a == AreaSit || a == AreaActive || a == AreaDeny
}

// areaConfFromSource cell Source → area provenance 信任权重。只对抑制跌倒的 rest-zone 有意义。
// FE 画(SourceHuman)=1.0 全信(保现有 full 抑制)；feedback=0.6；纯自学=0.4 —— 越暂定抑制越弱。
func areaConfFromSource(s Source) float64 {
	switch s {
	case SourceFeedback:
		return 0.6
	case SourceLearned, SourceGeometry:
		return 0.4
	default: // SourceHuman / SourceUnset
		return 1.0
	}
}

// areaConfFromGrid 该 (x,y) area 的 provenance 信任 [0,1]，喂 belief Observation.AreaConf。
// 仅 rest-suppress area(Bed/Toilet/Shower)按 cell Source 打折；Enter/开阔地板/Unknown 全信(1.0)。
func areaConfFromGrid(g *RoomGrid, x, y int) float64 {
	// nil-safety 收敛进 CellPrior accessor;g==nil → AreaTypeAt ok=false → 落 default 全信 1.0。
	if g.NearestEntryDistCm(x, y) <= beliefEnterMarginCm {
		return 1.0 // Enter 是几何，全信
	}
	a, ok := g.AreaTypeAt(x, y)
	if !ok {
		return 1.0
	}
	switch a {
	case AreaBed, AreaToilet, AreaShower: // 抑制跌倒的 rest-zone → 按 provenance 给信任
		s, _ := g.SourceAt(x, y)
		return areaConfFromSource(s)
	default:
		return 1.0
	}
}

// radarFrameAdapter 一帧 radar track → []Observation（pose / 运动学 / vital / track-present）。
// 命门：长冻 track 的 pose/kinematics/vital 置 Fresh=false（Conf=0 不更新），即便 LastObservedMs 仍新
// （firmware 冻结期持续 1Hz 推同帧）。ghost-ness 不受冻结影响（verdict 仍有效）。
func radarFrameAdapter(t observation.Track, ts *TrackState, grid *RoomGrid, nowMs int64, night bool) []belief.Observation {
	x, y, z := 0, 0, 0
	if t.PositionX != nil {
		x = *t.PositionX
	}
	if t.PositionY != nil {
		y = *t.PositionY
	}
	if t.PositionZ != nil {
		z = *t.PositionZ // P2.3:z 喂 ObsZBand posture(高度档),非 fall(R5)
	}
	// 权威源直填 ObsPose（取代 geom intermediate）：门距几何 + cell.area_type。门优先在 poseLikelihood 内复刻。
	ac := areaConfFromGrid(grid, x, y) // area provenance 信任（FE 画=1 / feedback=.6 / 自学=.4）
	nearDoor := grid.NearestEntryDistCm(x, y) <= beliefEnterMarginCm
	var area AreaType
	if a, ok := grid.AreaTypeAt(x, y); ok {
		area = a
	}
	areaTypeAt := int(area)

	tsFresh := nowMs-ts.LastObservedMs <= beliefRadarTTLMs
	stillBox := ts.StillBoxRunStart > 0 && nowMs-ts.StillBoxRunStart >= beliefStillBoxStaleMs
	motionFresh := tsFresh && !stillBox

	poseConf := 0.7
	if t.PoseConfidence > 0 {
		poseConf = float64(t.PoseConfidence) / 100
	}

	// P2.1(§10#3a):删 kinematics Δz —— z 骤降是环境噪声,不当 fall 正向证据(R5 pose/z 对 fall 只正向,
	// z↓ 连正向都不算)。fall 压制只走 realness/spatial/sleepad/recapture/human-bed,不走 z。
	vitalVal := 0.0
	if t.HeartRate > 0 || t.RespiratoryRate > 0 {
		vitalVal = 1
	}

	// P5:ghost/反射归 Track 层(TrackBelief co-existence + Conf×P(Real) 退火),Room 层不再喂 ObsTrackPresent。

	// P4.1 dwell 生存函数:still 时长(StillBoxRunStart 起)→ ObsDwellStill ramp。Fresh=tsFresh(track 在即可;
	// dwell-fall 恰是"久静"语义,不受 still-box motion-stale 命门压制——与 pose/kinematics 不同)。
	dwellSec := 0.0
	if ts.StillBoxRunStart > 0 && nowMs > ts.StillBoxRunStart {
		dwellSec = float64(nowMs-ts.StillBoxRunStart) / 1000
	}
	// P4.4(裁决⑱ B2):开阔地 dwell 生存尾按 cell tolerance 拉长(被容忍久站→尾长→不报)。
	// 只读 CellPrior(R3),不触学习;非开阔地不消费此值(toilet Z_cell-无关 / rest 不报)。
	dwellTol := 1.0
	if !nearDoor && isOpenFloorArea(area) {
		dwellTol = grid.ToleranceFactorAt(x, y)
	}
	// 雷达远边缘 still/pose/z 不可信 → dwell 尾 ×dwellEdgeMult。用 live raw 距离（本地系原点）：
	// 真摔经久静到 fire 时人已 settle，这个 live 坐标即稳定的最终位置（用户点 2）。
	radarDistCm := distInt(ts.LastRawH, ts.LastRawV, 0, 0)

	out := []belief.Observation{
		{Source: t.LogicID, Kind: belief.ObsPose, Value: float64(t.Pose), Conf: poseConf, Ts: nowMs, Fresh: motionFresh, AreaConf: ac, AreaType: areaTypeAt, NearDoor: nearDoor},
		{Source: t.LogicID, Kind: belief.ObsZBand, Value: float64(z), Conf: 0.7, Ts: nowMs, Fresh: motionFresh},                                         // P2.3 z 高度档 posture(不进 fall,R5)
		{Source: t.LogicID, Kind: belief.ObsDwellStill, Value: dwellSec, Conf: 0.7, Ts: nowMs, Fresh: tsFresh, ToleranceFactor: dwellTol, Night: night, RadarDistCm: radarDistCm}, // P4.1+P4.4 dwell 生存 ramp(开阔地带 tol gate)+P4.3 夜尾+雷达边缘×1.5
		{Source: t.LogicID, Kind: belief.ObsVitalPresent, Value: vitalVal, Conf: 0.6, Ts: nowMs, Fresh: motionFresh},
	}

	// 注：no-detect **不在逐帧 adapter 里发**。它是「走动中突然消失」的**消失事件**，由 engine/replay
	// 在 track 丢失时每 tick 调 noDetectObs 发出（前置=消失前 60s 在走动 / still-box<60s，见 replay 扫掠逻辑）。
	// 不再用"静止时长 ramp"(判反)、不再用 pose 门控(radar pose 不可靠,2026-06-01 用户实证)。
	return out
}

// shadowTrackGhostness — P3.1 独立 shadow realness ghost-ness ∈{0,1}(委员会审查⑦裁定:
// **不复用生产 Verdict/GhostPenalty** —— 生产 Verdict 正是漏 cd2b 冻结 ghost 的那套,复用=继承漏报)。
// 三探测器从 XY raw 现算/读(facts 非 verdict):① 本帧空间跳跃 Δ/dt ② Kalman 峰值残差(预测偏离)
// ③ 隐含速度(全程平均 dist(now,birth)/age,拼接 birth-incoherence)。任一超室内-老人 shadow 阈 →
// 近确定 ghost(g=1):P(瞬移|real elderly)≈0。catch 生产 Verdict 漏判的冻结/拼接 ghost。
// **只读 raw 动学量,不碰生产判定/生产闸**(R0)。frameJumpCmS<0 表本帧无法算跳变(缺前帧)→ 跳过①。
func shadowTrackGhostness(ts *TrackState, frameJumpCmS float64) float64 {
	if ts == nil {
		return 0
	}
	if frameJumpCmS > beliefIndoorSpeedCeilCmS { // ① 空间跳跃(逐步 raw teleport)
		return 1
	}
	if float64(ts.MaxImpliedSpeedFromBirth) > beliefIndoorSpeedCeilCmS { // ③ 隐含速度(全程平均,birth-incoherence)
		return 1
	}
	if ts.MaxKalmanResidual > beliefShadowKalmanResid { // ② Kalman 残差(逐步 model-relative 急变)
		return 1
	}
	return 0
}

// shadowFrozenArtifact — P3.2 冻结伪迹复合签名门控(委员会 review③):判 ghost = A∧(B≥2)。
// 补 P3.1 跳变检测漏的**静止反射体**模式。**门控形非裸佐证**:真人远角久站(B③④⑤可全中但缺 A)→不判,
// 防补漏报反引 FP。realness 用 XY/几何/cell(只读)+ pose/z **方差/锁死**(非 pose/z 值喂 fall,R5)。
//
//	A 近必要(二者居一):跳变出生(隐含速度超室内天花板,track-swap)∨ cell=AreaDeny(常驻反射已学,§9 只读)。
//	B 佐证(≥2):③ pose/z 锁死(连续帧恒定,生理无变)④ 钉死小区(30s box 散布小)⑤ 距门远(非门口驻留)。
func shadowFrozenArtifact(ts *TrackState, poseZLock int, grid *RoomGrid, x, y int, areaType AreaType, nowMs int64) bool {
	if ts == nil {
		return false
	}
	jumpBirth := ts.MaxImpliedSpeedFromBirth > int(beliefIndoorSpeedCeilCmS)
	cellDeny := areaType == AreaDeny
	if !jumpBirth && !cellDeny { // A 不成立(真人远角久站:无跳变出生 + cell 非 Deny)→ 不判 ghost
		return false
	}
	b := 0
	if poseZLock >= beliefFrozenPoseZLockTicks { // ③ pose/z 锁死
		b++
	}
	if ts.BoxRangeWithinMs(beliefFrozenWindowMs, nowMs) <= beliefFrozenSpreadCm { // ④ 钉死小区
		b++
	}
	if grid != nil && grid.NearestEntryDistCm(x, y) > beliefFrozenDoorFarCm { // ⑤ 距门远
		b++
	}
	return b >= 2
}

// realnessStep — P3.3 记忆 L_R filter(承审查⑨:把 P3.1/P3.2 二值 ghost 检测 + 走动 real 证据
// 积分成**连续 P(real) log-odds 带遗忘 γ**,供边缘化 P(fall)=P(fall|real)·P(real))。
// 摔倒瞬间 v≈0 无当下证据,realness 靠摔前走路累积的 L_R 经 γ 带进倒地窗(cabb-0605:走动→倒地不塌)。
// L>0 偏 real,L<0 偏 ghost。返回 (新 L_R, ghostness=P(ghost)=1−σ(L_R))。
// (WeakBio/出生地 real 证据 v2 接:base 暂无 vital;走动 MoveActive 是 XY 派生 real 证据,R5-safe。)
// dtSec=本帧距上帧秒数(帧率无关遗忘);首帧/dtSec≤0 → 不衰(prevLO 一般为 0)。
func realnessStep(prevLO, dtSec float64, moving, jumpGhost, frozenGhost bool) (float64, float64) {
	d := 0.0
	if moving { // 走动 = 人(累积 realness,摔前)
		d += lnLRMoveReal
	}
	if jumpGhost { // P3.1 跳变/急变 → 近确定 ghost
		d -= lnLRJumpGhost
	}
	if frozenGhost { // P3.2 冻结伪迹 → 近确定 ghost
		d -= lnLRFrozenGhost
	}
	decay := 1.0
	if dtSec > 0 {
		decay = math.Pow(beliefRealnessDecayPerSec, dtSec) // 时间基遗忘(半衰~69s,cabb-0605 52s 窗存活)
	}
	lo := decay*prevLO + d
	if lo > beliefRealnessLOCap {
		lo = beliefRealnessLOCap
	} else if lo < -beliefRealnessLOCap {
		lo = -beliefRealnessLOCap
	}
	return lo, 1.0 / (1.0 + math.Exp(lo)) // P(ghost)=1−σ(LO)
}

// realnessPFromLO P6.1a:shadow realness log-odds → P(R_i=real)=σ(LO)。与 realnessStep 的 ghostness=1−σ(LO) 对偶。
func realnessPFromLO(lo float64) float64 { return 1.0 / (1.0 + math.Exp(-lo)) }

// isSelfRescueRecapture — P3.4:曾丢失(lostAnchor>0 = lost-fall ramping)的 track 返回、且丢失 ≥ 阈
// → self-rescue candidate(跌后自救可能)。gate-list lost-fall 退役后,跌后返回的取消/认定全归 DBN
// belief shadow:标 self-rescue 留低 severity(只 log 不 fire,R1)——否则把"摔了又自己爬回来"当没事抹掉
// (对齐记忆 silent_leftbed_fall_recovery_window_gap)。
func isSelfRescueRecapture(lostAnchor, lastSeenMs, nowMs int64) bool {
	return lostAnchor > 0 && nowMs-lastSeenMs >= beliefSelfRescueMinGapMs
}

// noDetectObs track 丢失（走动前置已满足）→ P(no-detect|s) 发射。每 tick 固定强度（无时长项，时长→P3）；
// exit/返回/reachableExit/np=0 经各自似然仲裁方向。geom 留作未来 geom 条件 no-detect（床/桶遮挡）扩展位。
// noDetectObs P6.1a:no-detect 抬 Fallen 须门控。realnessP=P(R_i=real)(shadow realness σ(LO))、
// doorExitP=P(door-exit)(reachableExitScore)——经 Observation 显式字段入 likelihood(1+gain·Ri·(1−doorExit))。
func noDetectObs(realnessP, doorExitP float64, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsNoDetect, Conf: 0.8, Ts: nowMs, Fresh: true, RealnessP: realnessP, DoorExitP: doorExitP}
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
func reachableExitObs(distCm int, approachSpeedCmS float64, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsReachableExit, Value: reachableExitScore(distCm, approachSpeedCmS), Conf: 0.8, Ts: nowMs, Fresh: true}
}

// radarEventToObs 离散 radar 事件 → Observation（EnterRoom/ExitRoom）。
// WF-b(委员会 2026-06-08):firmware Fall **不进 shadow**——shadow 独立白盒由 pose/dwell/nodetect 自判
// (R5-纯,不可信 firmware pose 不污染 belief);firmware Fall 事件仅留生产 gate/RecordRadarAlarm 路径。
func radarEventToObs(eventName string, nowMs int64) (belief.Observation, bool) {
	switch eventName {
	case alarm.EnterRoom:
		return belief.Observation{Kind: belief.ObsEnterExit, Value: +1, Conf: 0.9, Ts: nowMs, Fresh: true}, true
	case alarm.ExitRoom:
		return belief.Observation{Kind: belief.ObsEnterExit, Value: -1, Conf: 0.9, Ts: nowMs, Fresh: true}, true
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
		{Source: o.DeviceAddr, Kind: belief.ObsVitalPresent, Value: vitalVal, Conf: 0.9, Ts: o.TMs, Fresh: fresh},
		{Source: o.DeviceAddr, Kind: belief.ObsBedOccupied, Value: bedVal, Conf: 0.9, Ts: o.TMs, Fresh: fresh},
	}
}

// bedAdapter bed 贝叶斯 scorer 输出 → ObsBedOccupied（嵌套 belief 子证据）。
// BedConfidence(0/60/90) 透传为 Conf；0=无数据→不更新。
func bedAdapter(b card.BedState, nowMs int64) []belief.Observation {
	bedVal := 0.0
	if b.BedStatus == 0 { // 0=InBed
		bedVal = 1
	}
	return []belief.Observation{
		{Source: "bed", Kind: belief.ObsBedOccupied, Value: bedVal, Conf: float64(b.BedConfidence) / 100, Ts: b.BedStatusTs, Fresh: b.BedConfidence > 0},
	}
}

// neighborToObs §5.5.2 弱耦合：邻居 room 占用信念 → 本房 ObsNeighbor。
// occ = 邻居 P(占用) [0,1]；conf = 邻居 belief 确定度。9h 治本近似：sleepad InBed 当 occ。
func neighborToObs(occ, conf float64, nowMs int64) belief.Observation {
	return belief.Observation{Kind: belief.ObsNeighbor, Value: clampUnit(occ), Conf: conf, Ts: nowMs, Fresh: true}
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
