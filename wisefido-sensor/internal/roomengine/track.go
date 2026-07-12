package roomengine

import "math"

// TimedPoint 带时间戳的画布坐标（cm 整数）
type TimedPoint struct {
	X, Y, Z int // canvas（旋转后）；raw firmware 不存——发布 floor 坐标时由 canvas 经 CanvasToRadar 单源逆算
	TMs     int64
}

// TrackState 单个 track 的完整生命档案。
// 运动粒子：Kalman + 即时判定 + 状态机转移。
// 不累计空间属性（那些归 Cell）。
type TrackState struct {
	// ---- 身份 ----
	TrackID    int
	DeviceAddr string
	RoomID     string
	// LogicID 出生时锚定的稳定逻辑身份（firmware track_id 不稳：复用/跳变/分裂）。
	// 格式 uidlast4 + track_id + mmssms（出生时刻分秒毫秒）。无-enter 新 track 经最近邻
	// 关联继承上一条的 LogicID（见 assignLogicID）。下游 ghost/verdict 按 LogicID 聚合。
	LogicID string

	// ---- 出生档案 ----
	BirthPos TimedPoint

	// ---- Kalman（内部 float，不持久化；track 死即销毁）----
	Kalman *KalmanFilter2D

	// ---- 历史窗口（滚动 HistoryLen=30 帧；判定只用近 MotionWindowSec 秒）----
	History    []TimedPoint
	FrameCount int

	// ---- 核心姿态状态机（Retract / Fall 检测用）----
	PrevCore     CorePose
	LieEnteredAt int64 // 进入 Lie 的时间（ms；0 = 未在 Lie 态）
	LieEnteredX  int
	LieEnteredY  int

	// ---- 静止状态机（still-box 单源：cell engine 久静量消费 box，见 updateContinuousIndicators/scoreMovement）----
	LongStillReported bool // 防 LongStill 重复上报
	StillFallReported bool // 防 still-fall 重复上报（bathroom + pose=Stand + 15/18min）

	// ---- 最后观测 ----
	LastPose int
	// Prev2Pose 上上帧 pose（N-2）。teleport 闸1 看跳前 2 帧 pose 白名单，O(1) 免 History 扫
	// （History 不存 pose）。每帧观测左移记入；新生轨默认 0(Unknown)，闸1 保守不删。
	Prev2Pose int
	LastZ     int

	// ---- 瞬移干扰待删窗（teleport_interference.go）----
	// SuspectInterferenceSinceMs：闸1 过、闸2 未到的待删窗起点 ms（0=非嫌疑）。置位期间
	// SnapshotTrackStatuses 零其 StillBoxSec（排出 floor 累积），等闸2/超时再 PURGE。
	SuspectInterferenceSinceMs int64
	PreJumpX, PreJumpY         int // 跳变前真人原区域（闸2 查"是否被另一活轨接住"）
	DriftX, DriftY             int // 漂移落点（PURGE 后压制重建匹配 + reflector candidate）
	// FloorArtifactSinceMs：immature-coast 反射伪迹的 floor 抑制 latch（0=未标）。一条「真实寿命极短 +
	// 转 coast + 无倒地前兆 + 有活轨接住共存」的孤儿，coast 起标置位，SnapshotTrackStatuses 零其
	// StillBoxSec 不喂 floor；不删轨（25min lostStillCarry 自然 GC）；再被观测且移走→解标。
	FloorArtifactSinceMs int64
	// InterferBornSinceMs：interfer 出生伪迹的 floor 抑制 latch（0=未标）。诞生在 interfer cell + 无进门 +
	// 无近邻可继承 + 出生时孤立（≥120cm）的孤迹（帘扇杂波），出生即置位，SnapshotTrackStatuses 零其
	// StillBoxSec 不喂 floor；走出出生点/现摔倒前兆→解标（churn 原地重生的真人摔倒得以恢复 floor 网）。
	// 软压非 purge：出生无姿态历史无法即时定性，留轨待撤销（见 stampInterferBornIfIsolated）。
	InterferBornSinceMs int64
	// LastFwAreaID firmware 直发的 area_id（present 帧更新；miss tick 不更新 = 冻结末值）。
	// 床判定 N：命中某床 areaId → 该床占用。255=声明区外/0=无区。
	LastFwAreaID int
	// LastCellArea 最近一次 SnapshotTrackStatuses 在画布坐标查到的 cell 真实区域类型（几何事实，单源）。
	// -1 = 无效（越界 / 无格）。本字段永远是当前落格真值，
	// 不随 still-box 锁定/移动期抹成 Unknown（Unknown(0) 是有效区型不是缺省哨兵）。
	// 发布层 / evidence / veto 复用此值，禁用 raw 坐标重新 CellAt（帧错 + 双写 drift）。
	LastCellArea int
	// firmware 直发的最后一帧 raw 雷达本地坐标。仅用于 alarm/event publish 时作 parent track 原样上抛；
	// 内部算法（grid/cell/fall）一律读 Kalman 后的画布坐标，不读这里。
	LastRawH, LastRawV, LastRawZ int
	// LastUpdateMs：每次 processFrameAt（含 miss tick）都会刷新到 nowMs，反映"engine
	// 最近一次为该 track 计算/预测过的时间"。不能用作"最后真正看到 track 的时间"。
	LastUpdateMs int64
	// LastObservedMs：最近一次真实收到帧（PushPoint）的时间。miss tick 不更新。
	// 用途：track 真实失踪点 ms。PR-9 前曾用于 number_people=0 ExitRoom 兜底比对（已删）。
	// 字段保留供 PR-10 BathroomLostFall + PR-11 bedroom lost_fall 重写时作"失锁时刻"基准。
	LastObservedMs int64

	// ---- cell 穿越追踪（Walk 学习用）----
	// 初始为 -1 表示尚未定位；新 cell 进入且 core==Move 时调 grid.MarkTraverse 计数 ++。
	LastCellCol int
	LastCellRow int

	// ---- StillBox（静止无移动）检测（box 判据，2026-05-03 由 byte-equal 改为 box）----
	// StillBoxRunStart：当前 still box run 起点 ms（0 = 未达 still 状态）。
	// 判据：失锁前 30s 滚动窗口内位移 box per-axis (max-min) <= StillBoxCm(50) → 视为 still。
	// 抗 X/Y 抖动 / 抗摔倒抽搐（box 容差替代 byte-equal 严格相等）。
	// 用于 lost-fall pending 计算 credit（半计入等待）+ PR-C 流式 cancel 守卫。
	StillBoxRunStart int64
	// still-box 单源派生（updateContinuousIndicators 同步算，cell engine 久静量读，替代旧 StillSince/StillX/StillY）：
	StillBoxStartX, StillBoxStartY, StillBoxStartZ int // box run 起点 canvas（=History[0] 回填点）→ MarkDwell/cell + floor fall 坐标(经 CanvasToRadar 逆算 raw,Z 透传)
	// box 起点锁定的 chair 区久坐兜底 → floor 连续 tFloor 单源（仅 chair 区）。box-break 清。
	StillBoxInChair     bool
	StillBoxChairMu     float64 // 本椅 14 日久坐均值 AV（秒）；0=冷启
	StillBoxChairSigma  float64 // 本椅 14 日久坐标准差（秒）
	StillBoxChairMaxSit float64 // 本椅 false_alarm 反馈棘轮（秒）
	StillBoxChairCap    float64 // 本椅 floor 分级封顶（秒，§12）：人 pin=90min / 影子学习=30min；0=回退 90min
	StillBoxBreakDurMs             int64 // 本帧 still-box 刚 break 的 dwell 时长（0=未 break）→ 移动块 MarkDwell 消费

	// ---- AreaSit 4 通道学习累加器（sit_learning.go；StillBox run = episode，box-break = walk-away emit）----
	sitFwContigStartMs int64   // 当前连续 pose=Sitting 段起点（断则 0；用于 3s 软门 + fwMax）
	sitFwMaxMs         int64   // 本 episode 内最长连续 Sitting 时长（ms）
	sitZBest           float64 // 本 episode 内 max sitZMembership（仅 z>0 帧更新；缺失中性）

	// DBNConfidence：DBN per-track 置信度（realness PReal 0-100，S0.c-4 第三腿 A·R13/R15.3-2，
	// 经 SetTrackConfidence 每帧写回）。<0 = DBN 未设 → payloadFromTrack TrackConfidence 取 0。
	// 仍是门控/occupancy/forensic(MaxGhost) 的每帧真值，可随几何升降。
	DBNConfidence int
	// EmitConf：发到 FE 的显示置信度（单向 max-hold）。反证法下 realness 判决单向（RealProven 只增），
	// 显示亦应单向：一旦 proven 就棘轮不回落（除非 split=新身份，自然随新 LogicID 重置）。churn 继承 parent。
	// 与 DBNConfidence 解耦——后者供门控随区漂移升降，本值只管"是否真"的稳定显示。<0=未设→取 0。
	EmitConf int
	// lastEmitReal：上次发布时 EmitConf 是否在 real 档(≥80)。用于 real↔ghost 跃迁时绕过 20s 节流即发。
	lastEmitReal bool

	// ---- real-by-provenance latch（Phase 1.5，doc/area_selfcheck_and_ghost_raw_spec.md）----
	// RealProven：出生/present 帧任一时刻 provenance 满足（门第=EnterRoom 或 raw 末点区型=AreaEnter；
	//   床=raw 末点区型∈{Bed,MonBed} ∧ sleepad 占用）→ 单调置真。挂 lid：出生 nearestAliveTrack 继承
	//   parent.RealProven，带得过 churn（治 cd2b 重生睡者被全判 ghost）。消费端一律读 realLatchActive(area)
	//   （= RealProven ∧ 当场查 QueryAreaType(rawLastPos)∈rest 区，Option B confine），非裸读本字段。split 重分配硬清。
	RealProven bool
	// MaxGhost：生涯峰值 census p_mirror（forensic/lost_fall delete §10.3-b）。p_mirror=1-PReal（realness 2 态互补）。
	MaxGhost float64
	// ghostSustainRun：当前连续 p_mirror≥0.8 的 tick 数（断则清零）。
	ghostSustainRun int
	// MaxGhostSustained：曾达 p_mirror≥0.8 持续 ≥3 tick（条件2④ 定稿：sustained 非峰值）。
	MaxGhostSustained bool
	// MaxRoomNp：born→now 期间本房观测到的最大 np。≥2 = 曾有共存伙伴（某真人的反射佐证）。
	MaxRoomNp int

	// ---- split-ghost 时机（split_ghost.go；vote→spliter→group→step3 软压）----
	// SplitObservingSinceMs：split_group 成员标记起点 ms（0=非 group 轨）。出生时 step1 vote 命中 spliter
	// 且 spliter 静止门通过 + offspring 无 EnterRoom 时，由 stampSplitGroupOnBirth 给组内各成员打。
	SplitObservingSinceMs int64
	// SplitSpliterX/Y：本 group 的 spliter（静止干扰源）锚点（前一tick位）。成员到此 ≤glue 即赖着不走=ghost。
	SplitSpliterX, SplitSpliterY int
	// SplitLastLatchMs：step3 ~2s 节流上次时刻（见 splitLatchIntervalMs）。
	SplitLastLatchMs int64
	// SplitEverWalkedOut：离锚点净距曾 ≥ splitWalkOutCm(200) = 真人 walk → 永久 real，不被后续 still 翻案。
	SplitEverWalkedOut bool
	// splitGlueRun：GLUED(离锚点≤50) 连续 tick 计数（split 本地 3-tick 坐实棘轮；HOLD/WALKOUT 断则清零）。
	splitGlueRun int
	// SplitGhostSinceMs：赖在 spliter 锚点的 ghost 软压 latch（>0 → SnapshotTrackStatuses 零其 StillBoxSec
	// 不喂 floor）。**单调**：坐实(3-tick GLUED)置位，只由 walkout/露倒地相解除；穿过 HOLD 不清。不删轨（避免 churn）。
	SplitGhostSinceMs int64
	// SplitSpliterLID：本 group 的 spliter（占用 mantle 持有者/静止干扰源）身份 lid，出生时由 stampSplitGroupOnBirth 记下。
	// succession(§6.12) 靠它反向定位"我的锚是谁"——spliter 恰不带 SplitObservingSinceMs，无此链则其 silent-lost 无人可继。
	SplitSpliterLID string
	// SplitProvisionalReal：succession 暂持 real mantle（HSRP standby 顶上，逐帧可撤销、非 latch）。spliter silent-lost
	// 且 living 候选 ≥2 时，PReal 最高者置真占 nr（census ForceReal ← RealProven||此位）；跨硬门槛/living 剩 1 → 转 RealProven 清此位。
	SplitProvisionalReal bool
}

// Track 生命周期常量
const (
	HistoryLen      = 30   // 滚动窗口帧数
	MotionWindowSec = 5    // 运动学判定窗口（近 5 秒）
	StillThreshCm   = 15   // 帧间位移 < 此值视为静止
	LieRetractMs    = 8000 // 进入 Lie 后 < 此时长回到 Stand/Move → Retract
	// 经验值：真跌倒 8 秒内爬起概率 < 5%；雷达固件的 fallSec 典型 10-30 秒，
	// 8 秒远小于其最小值，确保只捕获"雷达尚未升级为 Fall 就回撤"的误报。
)

// BedSession 单 sleepad 设备的"在床会话"。silent-fall（依赖 radar 印证/HR-RR latch 那套）随 gate-list 退役后,
// 只保留 BedOccupancyState→cardagg bed_state 的存活三元组:身份 + InBed/LeftBed 时刻。
type BedSession struct {
	DeviceUID    string // sleepad device_uid
	InBedSinceMs int64  // 首次 InBed 到达的时间戳；0 = 未在床
	LeftBedAtMs  int64  // 0 = 仍在床；>0 = 已离床
}

// NewTrackState 新 track 出生
func NewTrackState(trackID int, deviceAddr, roomID string, x, y, z int, tMs int64) *TrackState {
	birth := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	return &TrackState{
		TrackID:        trackID,
		DeviceAddr:     deviceAddr,
		RoomID:         roomID,
		BirthPos:       birth,
		Kalman:         NewKalmanFilter2D(float64(x), float64(y)),
		History:        []TimedPoint{birth},
		FrameCount:     1,
		DBNConfidence:  -1, // <0 = DBN 未设，payloadFromTrack TrackConfidence 取 0
		EmitConf:       -1, // <0 = 未设 → payloadFromTrack 取 0；proven 后棘轮 max-hold
		LastZ:          z,
		LastUpdateMs:   tMs,
		LastObservedMs: tMs,
		LastCellCol:    -1,
		LastCellRow:    -1,
	}
}

// still-box 判据：失锁前 30s 滚动窗 per-axis box(max-min) ≤此 cm 视为 still（updateContinuousIndicators 消费）。
// 实测倒地质心抖 50×40cm → 50 才正确判 still 不误判"动"。
const stillBoxCm = 50

// stillPathCm：30s 窗内逐帧累计路程 >此 cm → 判"50cm 盒内踱步"非静止，清零 still-box。
// 补 box(max-min) 盲区：盒内反复踱步包围盒小但累计走好几米；用运动量直接量出，不靠 pose。
const stillPathCm = 200

// stillStepFloorCm：累计路程只计单段位移 ≥此 cm 的帧间步；< 此 = firmware 抖动(±5-10cm,
// 段间距≈1.77σ≤18cm)不计。源头滤抖动 → 静止者路程恒 0 不误破盒;真挪步(≥30cm)逐段穿过。
const stillStepFloorCm = 30

// PushPoint 追加一帧观测到历史窗口
func (ts *TrackState) PushPoint(x, y, z int, tMs int64) {
	pt := TimedPoint{X: x, Y: y, Z: z, TMs: tMs}
	ts.History = append(ts.History, pt)
	if len(ts.History) > HistoryLen {
		ts.History = ts.History[len(ts.History)-HistoryLen:]
	}
	ts.FrameCount++
	ts.LastUpdateMs = tMs
	ts.LastObservedMs = tMs
}

// isRestArea rest-provenance 区：静止真人合法待着的区（床/坐/躺/门）。开阔活动区/deny/reflector/interfer
// 非 rest → 静止可疑，latch 不豁免 ghost 判定（Option B confine）。
func isRestArea(a AreaType) bool {
	switch a {
	case AreaBed, AreaMonitorBed, AreaSit, AreaLying, AreaEnter:
		return true
	}
	return false
}

// realLatchActive real-by-provenance latch 当前是否生效（Option B）：已 provenance 证真 ∧ 当前 raw 自查区（area
// = QueryAreaType(rawLastPos)，调用方当场查）仍在 rest 区。true → 消费端强制 real / 豁免 ghost。飘出 rest 区
// （真摔在开阔地板 / 冒名 phantom 走开）→ false → ghost 检测器与 floor 照常，绝不压真摔。
func (ts *TrackState) realLatchActive(area AreaType) bool {
	return ts.RealProven && isRestArea(area)
}

// HasHistory 是否有足够帧数做 Kalman
func (ts *TrackState) HasHistory() bool {
	return ts.FrameCount >= 2
}

// DisplacementWithinMs 计算 History 中过去 windowMs 内的位移 box 对角线（cm）。
//
// 用于 StillBox（静止无移动）检测的 box 判据：仅看 (max-min) 范围而非每帧累计，对 firmware
// X/Y 抖动 + 摔倒抽搐鲁棒——只要轨迹被框在 box 内，box 对角线即上限。
//
// nowMs 通常是当前帧时间戳；窗口 = [nowMs - windowMs, nowMs]。
// History 最多 HistoryLen=30 帧 = 30s 滚动窗口；超出此长度无法回看。
func (ts *TrackState) DisplacementWithinMs(windowMs int64, nowMs int64) int {
	cutoff := nowMs - windowMs
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32
	count := 0
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		count++
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if count < 2 {
		return 0
	}
	dx := maxX - minX
	dy := maxY - minY
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
}

// BoxRangeWithinMs 返回过去 windowMs 内 History 位移 box 的 per-axis 边长较大者 max(dx, dy)（cm）。
//
// 与 DisplacementWithinMs（对角线 √(dx²+dy²)）的区别 = 50×50 方框判据:dx,dy 各 ≤ 阈值即静止。
// 对角线会把 50×40 的框算成 64（偏严，把倒地质心抖动误判成"动"）;per-axis 算 50（≤50 仍判 still）。
// still-box 静止判定用此（见 updateContinuousIndicators）;DisplacementWithinMs 仍供
// static_reflector / belief 速度采样等用对角线口径,二者不混。
func (ts *TrackState) BoxRangeWithinMs(windowMs int64, nowMs int64) int {
	cutoff := nowMs - windowMs
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32
	count := 0
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		count++
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if count < 2 {
		return 0
	}
	dx := maxX - minX
	dy := maxY - minY
	if dx > dy {
		return dx
	}
	return dy
}

// PathLengthWithinMs 累计 History 过去 windowMs 内逐帧位移之和（cm）。
//
// 与 BoxRangeWithinMs（只看 max-min 包围盒）互补：包围盒对"50cm 盒内反复踱步"会误判 still
// （盒小但累计走好几米）。路程逐帧累加直接量出运动量，>stillPathCm 即真在动，不靠 firmware pose。
// lost-carry 冻结帧位移恒=0 → 不累加 → 不破真摔者的久静（只接住活人踱步）。
// 单段 <stillStepFloorCm 视为抖动不计：从源头滤掉 ±5-10cm firmware 抖动的逐段累积（否则静止者
// 29 段抖动可累计 >200cm 误破盒 → 漏报真摔者）。
func (ts *TrackState) PathLengthWithinMs(windowMs int64, nowMs int64) int {
	cutoff := nowMs - windowMs
	var sum float64
	prevValid := false
	var px, py int
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		if prevValid {
			dx := float64(p.X - px)
			dy := float64(p.Y - py)
			if seg := math.Sqrt(dx*dx + dy*dy); seg >= stillStepFloorCm {
				sum += seg
			}
		}
		px, py = p.X, p.Y
		prevValid = true
	}
	return int(sum)
}

// StillBoxBoundsWithinMs still 区在窗口内的 bounding box（画布 cm）。ok=false（<2 帧）→ 调用方回退单格。
// still-box per-axis ≤50cm = 多个 10×10 cell；学习按此区摊到覆盖格，与 DBN 查的 live 格对齐。
func (ts *TrackState) StillBoxBoundsWithinMs(windowMs int64, nowMs int64) (minX, minY, maxX, maxY int, ok bool) {
	cutoff := nowMs - windowMs
	minX, minY = math.MaxInt32, math.MaxInt32
	maxX, maxY = math.MinInt32, math.MinInt32
	count := 0
	for _, p := range ts.History {
		if p.TMs < cutoff {
			continue
		}
		count++
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return minX, minY, maxX, maxY, count >= 2
}

// TotalDisplacement 历史窗口内的总位移（cm）
func (ts *TrackState) TotalDisplacement() int {
	if len(ts.History) < 2 {
		return 0
	}
	sum := 0
	for i := 1; i < len(ts.History); i++ {
		sum += distInt(ts.History[i-1].X, ts.History[i-1].Y, ts.History[i].X, ts.History[i].Y)
	}
	return sum
}

// AgeSec 存活时长（秒）
func (ts *TrackState) AgeSec() int {
	if len(ts.History) < 2 {
		return 0
	}
	lastMs := ts.History[len(ts.History)-1].TMs
	return int((lastMs - ts.BirthPos.TMs) / 1000)
}
