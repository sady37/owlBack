package belief

// ObsKind belief 引擎认的全部观测种类（owlBack/doc/belief_input_normalization.md §2）。
// 派生信号（WeakBiometricSignal/RiskLevel/AloneContinuousMin）禁入——会双重计数
// （铁律见 feedback_no_dynamic_threshold_modulation）。引擎只吃原始物理观测。
type ObsKind int

const (
	ObsPose          ObsKind = iota // radar Track.Pose 枚举
	ObsVitalPresent                 // HR>0 ∨ RR>0
	ObsBedOccupied                  // bed 贝叶斯 P(InBed)，嵌套 belief 子证据
	ObsEnterExit                    // EnterRoom=+1 / ExitRoom=-1
	ObsNumberPeople                 // 房间人数 0-8
	ObsNeighbor                     // §5.5.2 弱耦合：邻居 room P(占用) [0,1]
	ObsNoDetect                     // P(no-detect|s)：本 tick 看了没测到（前置=消失前60s在走动）。状态条件似然=可检测态压低/可合理消失态保留；时长→P3，非斜坡
	ObsReachableExit                // 丢失点可达退场证据 e=f_dist·f_reach [0,1]（P2 软门：近门+单帧可达→偏 Left 压 Fallen）
	ObsZBand                        // P2.3 z 高度档(Value=z cm)：z>80→stand / 30–80→sit / <30→噪声无信息。只喂 posture,**绝不写 SFallen**(R5)
	ObsDwellStill                   // P4.1 dwell 生存函数(Value=still 秒,zone=room_type×area_type)：-lnS_vol(d|zone) 平滑 ramp 取代"still≥阈即报"硬悬崖
)

var obsKindLabel = [...]string{
	"Pose", "VitalPresent", "BedOccupied", "EnterExit", "NumberPeople",
	"Neighbor", "NoDetect", "ReachableExit", "ZBand", "DwellStill",
}

// String ObsKind 标签（source-fidelity 审计 / observability）。
func (k ObsKind) String() string {
	if int(k) < 0 || int(k) >= len(obsKindLabel) {
		return "Unknown"
	}
	return obsKindLabel[int(k)]
}

// Observation 统一 schema（belief_input_normalization.md §1）。位置语义由 AreaType/NearDoor/BedReleased
// 直接携带权威源（cell.area_type + 门距 + bed_state），不再经 Geom intermediate。belief 隐状态 S 不含坐标，位置只作 P(o|s) 条件。
// 所有源（radar/sleepad/room/bed/neighbor）翻成同格式，引擎只认 schema 不认设备类型。
type Observation struct {
	Source string  // 谁产出：device /128 / room /88 / bed /96
	Kind   ObsKind // 决定查 P(o|s) 哪一行
	Value  float64 // 归一化数值（连续）或枚举序号（离散）
	Conf   float64 // 可信度 [0,1]，0=缺失/stale → likelihood=I 不更新
	Ts     int64   // 观测时刻 unix ms
	Fresh  bool    // 在 TTL 内；stale → 当缺失（Conf 视为 0）
	// AreaConf area provenance 信任 [0,1]——cell Source 权重（FE 画=1 / feedback=0.6 / 自学=0.4）。
	// 0 或未设=1.0（全信，向后兼容）。<1 时把 area 条件似然向中性(area-unknown) blend——
	// 软先验替硬闸：暂定 rest-zone(Bed/Toilet) 抑制跌倒更弱，强跌倒证据仍能盖过，不漏真摔。
	AreaConf float64
	// ToleranceFactor cell 自适应容忍因子 ∈[1.0, MaxToleranceFactor]（P4.4 裁决⑱ B2；adapter 经 CellPrior.ToleranceFactorAt 读入）。
	// 仅 ObsDwellStill 开阔地消费：dwell 生存尾 scale*=tol（被容忍久站的 cell 尾更长→久站真人不报）。
	// 0 或未设=1.0（向后兼容，非 dwell 观测留零值即可）。语义=容忍越高尾越长，Value 恒 raw（不缩放）。
	ToleranceFactor float64
	// Night P4.3 风险时段标志（仅 ObsDwellStill 消费）：夜间 dwell 生存尾缩短→更快 ramp（久静夜间更可疑）。
	// adapter 由 IsNightTime(nowMs, room timezone) 填。0/未设=昼。
	Night bool
	// RealnessP/DoorExitP P6.1a 门控输入（仅 ObsNoDetect 消费；plumbing 同 B2 显式字段，公式留 likelihood）。
	// RealnessP=P(R_i=real)∈[0,1]（adapter 由 shadow realness σ(LO) 填）；DoorExitP=P(door-exit)∈[0,1]
	// （reachableExitScore 填）。ObsNoDetect 抬 Fallen 因子 = 1+gain·RealnessP·(1−DoorExitP):
	// ghost 消失(RealnessP→0)或门区可达走出(DoorExitP→1)→ 因子→1 中性(不裸 absence 抬 fall)。0=未设。
	RealnessP float64
	DoorExitP float64
	// RoomType/AreaType cell 权威值（dwellTailFor roomType×areaType 查尾表 + poseLikelihood 位置语义）。
	// 0=未设/向后兼容(default path)。
	RoomType int
	AreaType int
	// NearDoor 离最近门≤30cm（门距几何，area_type 不编码动态门距；构造侧由 grid.NearestEntryDistCm 填）。
	// BedReleased bed_state 离床（权威>几何）：床区(area=Bed)躺着的 ObsPose 翻睡→倒地候选（取代旧止血 geom 翻转）。
	NearDoor    bool
	BedReleased bool
}

// effConf 有效置信度：stale 观测当缺失。
func (o Observation) effConf() float64 {
	if !o.Fresh {
		return 0
	}
	return o.Conf
}
