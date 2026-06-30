package belief

// observation.go — 联合模型 emission 输入（§5）。原始物理观测直进 Φ，**不经派生中间量**
// （不吃 zoneengine 算好的 Probability()/ObsBedOccupied，硬 O_b 内化为隐变量 B^j）。
// 按 attachment 分轴：接触(sleepad InBed/LeftBed)→B / 雷达(pose,dwell,hrrr,δ位置)→S。

// BedReading 单床 sleepad 本 tick 读数。NoReport=离线/无报 → ℓ≡1 中性（§5：离线不冻结、不永久否决）。
type BedReading int

const (
	BedNoReport     BedReading = iota // 离线/无报：零信息
	BedInBed                          // InBed 翻转（接触/事件）
	BedLeftBed                        // sleepad 接触 LeftBed：果断清床占用（接触确定）
	BedLeftBedRadar                   // radar 几何 LeftBed：弱清（radar 在床稳定性<50%，可能是 track 漂移误报离床）
)

// Observation 一帧的原始观测。Sleepad 长 = numBeds（逐床接触读数）。
type Observation struct {
	// 接触轴（每床 sleepad）→ B^j。长 numBeds。
	Sleepad []BedReading

	// 雷达轴 → S。RadarOnline=false → 雷达全轴 ℓ≡1（离线=中性）。
	RadarOnline  bool
	PoseLying    bool // pose=Lying（二义：AtBed ∨ Fallen，刻意不分）→ boost SBed+SFallen
	PoseWalking  bool // pose∈{Walking,Running}（逐帧）→ 压 SFallen（运动⊥倒地静止）+ 抬 SOpenFloor
	PoseStanding bool // pose=Standing（软站，z 未知不硬压）→ 只抬 SOpenFloor（治 Empty，不压 SFallen）
	PoseSit      bool // pose=Sitting（仅椅/沙发，刻意不含坐地/床坐起=摔二义）→ 抬 SSit（弱排斥，不压 SFallen）
	// 固件已分类的跌倒（per-frame，非二义）→ 强 boost SFallen，经 confirm 窗/N_r/floor 照常裁决（不绕 belief）。
	PoseFallen      bool    // pose=5 确认跌倒（固件最高置信跌倒证据）→ 强抬 SFallen，不抬 SBed
	PoseSuspectFall bool    // pose=2 疑似跌倒（弱）→ 弱抬 SFallen，证据不足靠累积/floor 兜底
	Z               int     // 雷达本帧高度(canvas cm)：z≥80=站立身高，与"躺地"互斥 → 压 SFallen + 抬 SOpenFloor
	StillSec        float64 // still-box raw 总时长（连续静止秒）：喂 FloorGuard 纯计时器（不再含直立折扣）。emission 不消费
	NearBed         bool    // HR/RR 空间邻域门控（§5 用 nearBed 非 enterBed，门控 Online）
	// NearBedMask 逐床几何邻近（XY 在该床 NearBedMargin 内）。不门控 Online——present 用当前 XY，
	// lost 用冻结坐标。FloorGuard 用它把"接触 InBed 豁免"收窄到**本 track 所在床**，避免同房他人
	// 在床误豁免地板上的摔者。长 numBeds，与 Sleepad 同索引。
	NearBedMask []bool
	// AreaType track 当前 cell.Belief[0].Type（CellAreaType 每帧读活的，经 seam）。FN-safe **正向压制**：
	// sit/lying 区 → 抬对应静止态压 Fallen（redirect）。权重有上限（守门1：低到 still 久静能翻，不锁死）。
	// 值同 roomengine.AreaType：deny=1/bed=2/reflector=3/enter=4/monitorbed=5/interfer=6/sit=7/lying=8/active=9。
	// 床(bed/monitorbed)与 reflector 在 floor.Step 豁免；bed 占用走 RadarBedHitMask(firmware N) 驱动。
	AreaType int

	// Chair 区久坐兜底（hydrate 自 28，replay 期不学）：floor 只对 chair 区(InChair)算连续阈
	//   tFloor=clamp(max(1.5·ChairMu+10min, ChairMaxSit),≤90min)；冷启(ChairMu≤0)回退 90min。pin 几何 gate，免疫 Belief 翻转。
	InChair     bool
	ChairMu     float64 // 本椅 14 日久坐均值 AV（秒）；0=冷启
	ChairMaxSit float64 // 本椅 false_alarm 反馈棘轮（秒，只读 hydrate）

	// RadarBedHitMask 逐床 N：firmware area_id 命中该床声明的 areaId（= 雷达判"人在该床上"）。长 numBeds。
	// 替代旧 cell-area 床判定：cell 几何随 canvas drift 误判，firmware area_id 是雷达地面真值。
	// emission 床 boost = lArea × 该床 covers(M) × N（命中才抬 SBed）。离线 = 全 false。
	RadarBedHitMask []bool

	// RoomType 房型(card.RoomType: 1=Bathroom)：still 高斯 CDF 的 (μ,σ) 与 cell area **保守合并**(取 max μ)——
	// bathroom 房即使 cell 未画 toilet(落 unknown) 也至少用 bathsec(18min)，避免激进 default(12min) 过早误报。
	RoomType int

	// IsRiskTime 风险时段(夜间 IsNightTime)：只缩短 floor 时长兜底阈(tFloorFor 的 μ+k·σ 中 k 变小)，
	// **不进 C_FN/报警阈**——risktime 纯时间轴(管"等多久兜底",不管"报不报";后者由证据+可救援性定)。
	IsRiskTime bool

	// HR/RR（§5 非对称 + §D 门控）：
	HRRRObserved      bool // 本 tick 有 vital 通道读数；false=无通道 → 中性
	HRRRPresent       bool // HR>0 ∨ RR>0
	VitalSourceOnline bool // 房内**独立在线** vital 源(sleepad)；§D：radar 自身 absent 不得否决 AtBed
	// SleepadVitalPresent sleepad 接触 vital(InBed + HR/RR fresh within TTL)：活体在垫 → 抬 SBed。
	// 接触权威，**不经 radar NearBed/online 门控**(radar 直立位置漂不可靠；sleepad 垫压即在床)。
	// cd2b radar 不返 vital → 这条是 SBed 的唯一 vital 印证腿(radar HRRR 那条在 cd2b 恒空)。
	SleepadVitalPresent bool
	// 注（§32）：原 FloorStripXY（δ floor-strip）已删——cd2b 靠 LeftBed→B vac 经 Ψ 相容涌现，非雷达 XY 几何。
}
