package belief

// observation.go — 联合模型 emission 输入（§5）。原始物理观测直进 Φ，**不经派生中间量**
// （不吃 zoneengine 算好的 Probability()/ObsBedOccupied，硬 O_b 内化为隐变量 B^j）。
// 按 attachment 分轴：接触(sleepad InBed/LeftBed)→B / 雷达(pose,dwell,hrrr,δ位置)→S。

// BedReading 单床 sleepad 本 tick 读数。NoReport=离线/无报 → ℓ≡1 中性（§5：离线不冻结、不永久否决）。
type BedReading int

const (
	BedNoReport BedReading = iota // 离线/无报：零信息
	BedInBed                      // sleepad InBed 翻转
	BedLeftBed                    // sleepad LeftBed 翻转
)

// Observation 一帧的原始观测。Sleepad 长 = numBeds（逐床接触读数）。
type Observation struct {
	// 接触轴（每床 sleepad）→ B^j。长 numBeds。
	Sleepad []BedReading

	// 雷达轴 → S。RadarOnline=false → 雷达全轴 ℓ≡1（离线=中性）。
	RadarOnline bool
	PoseLying   bool    // pose=Lying（二义：AtBed ∨ Fallen，刻意不分）
	StillSec    float64 // dwell：连续静止秒（≥τ → 静止占用 D>1）
	NearBed     bool    // HR/RR 空间邻域门控（§5 用 nearBed 非 enterBed）

	// HR/RR（§5 非对称 + §D 门控）：
	HRRRObserved      bool // 本 tick 有 vital 通道读数；false=无通道 → 中性
	HRRRPresent       bool // HR>0 ∨ RR>0
	VitalSourceOnline bool // 房内**独立在线** vital 源(sleepad)；§D：radar 自身 absent 不得否决 AtBed
	// 注（§32）：原 FloorStripXY（δ floor-strip）已删——cd2b 靠 LeftBed→B vac 经 Ψ 相容涌现，非雷达 XY 几何。
}
