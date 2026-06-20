package belief

// observation.go — 联合模型 emission 输入（§5）。原始物理观测直进 Φ，**不经派生中间量**
// （不吃 zoneengine 算好的 Probability()/ObsBedOccupied，硬 O_b 内化为隐变量 B^j）。
// 按 attachment 分轴：接触(sleepad InBed/LeftBed)→B / 雷达(pose,dwell,hrrr,δ位置)→S。

// BedReading 单床 sleepad 本 tick 读数。NoReport=离线/无报 → ℓ≡1 中性（§5：离线不冻结、不永久否决）。
type BedReading int

const (
	BedNoReport    BedReading = iota // 离线/无报：零信息
	BedInBed                         // InBed 翻转（接触/事件）
	BedLeftBed                       // sleepad 接触 LeftBed：果断清床占用（接触确定）
	BedLeftBedRadar                  // radar 几何 LeftBed：弱清（radar 在床稳定性<50%，可能是 track 漂移误报离床）
)

// Observation 一帧的原始观测。Sleepad 长 = numBeds（逐床接触读数）。
type Observation struct {
	// 接触轴（每床 sleepad）→ B^j。长 numBeds。
	Sleepad []BedReading

	// 雷达轴 → S。RadarOnline=false → 雷达全轴 ℓ≡1（离线=中性）。
	RadarOnline  bool
	PoseLying    bool    // pose=Lying（二义：AtBed ∨ Fallen，刻意不分）→ boost SBed+SFallen
	PoseWalking  bool    // pose∈{Walking,Running}（逐帧）→ 压 SFallen（运动⊥倒地静止）+ 抬 SOpenFloor
	PoseStanding bool    // pose=Standing（软站，z 未知不硬压）→ 只抬 SOpenFloor（治 Empty，不压 SFallen）
	PoseSit      bool    // pose=Sitting（仅椅/沙发，刻意不含坐地/床坐起=摔二义）→ 抬 SSit（弱排斥，不压 SFallen）
	Z            int     // 雷达本帧高度(canvas cm)：z≥80=站立身高，与"躺地"互斥 → 压 SFallen + 抬 SOpenFloor
	StillSec     float64 // still-box raw 总时长（连续静止秒）：喂 FloorGuard 纯计时器（不再含直立折扣）。emission 不消费
	NearBed      bool    // HR/RR 空间邻域门控（§5 用 nearBed 非 enterBed，门控 Online）
	// NearBedMask 逐床几何邻近（XY 在该床 NearBedMargin 内）。不门控 Online——present 用当前 XY，
	// lost 用冻结坐标。FloorGuard 用它把"接触 InBed 豁免"收窄到**本 track 所在床**，避免同房他人
	// 在床误豁免地板上的摔者。长 numBeds，与 Sleepad 同索引。
	NearBedMask []bool
	// AreaType track 当前 cell.Belief[0].Type（CellAreaType 每帧读活的，经 seam）。FN-safe **正向压制**：
	// sit/toilet 区 → 抬对应静止态压 Fallen（redirect）。权重有上限（守门1：低到 still 久静能翻，不锁死）。
	// Sit=3/Active=4/Deny=5/Shower=6/Toilet=7（同 roomengine.AreaType）。床(Bed)这一支已移出 → RadarBedHitMask 驱动。
	AreaType int

	// RadarBedHitMask 逐床 N：firmware area_id 命中该床声明的 areaId（= 雷达判"人在该床上"）。长 numBeds。
	// 替代旧 cell-area 床判定：cell 几何随 canvas drift 误判，firmware area_id 是雷达地面真值。
	// emission 床 boost = lArea × 该床 covers(M) × N（命中才抬 SBed）。离线 = 全 false。
	RadarBedHitMask []bool

	// RoomType 房型(card.RoomType: 1=Bathroom)：still 高斯 CDF 的 (μ,σ) 与 cell area **保守合并**(取 max μ)——
	// bathroom 房即使 cell 未画 toilet(落 unknown) 也至少用 bathsec(18min)，避免激进 default(12min) 过早误报。
	RoomType int

	// HR/RR（§5 非对称 + §D 门控）：
	HRRRObserved      bool // 本 tick 有 vital 通道读数；false=无通道 → 中性
	HRRRPresent       bool // HR>0 ∨ RR>0
	VitalSourceOnline bool // 房内**独立在线** vital 源(sleepad)；§D：radar 自身 absent 不得否决 AtBed
	// 注（§32）：原 FloorStripXY（δ floor-strip）已删——cd2b 靠 LeftBed→B vac 经 Ψ 相容涌现，非雷达 XY 几何。
}
