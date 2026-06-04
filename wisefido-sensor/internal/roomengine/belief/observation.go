package belief

// ObsKind belief 引擎认的全部观测种类（owlBack/doc/belief_input_normalization.md §2）。
// 派生信号（WeakBiometricSignal/RiskLevel/AloneContinuousMin）禁入——会双重计数
// （铁律见 feedback_no_dynamic_threshold_modulation）。引擎只吃原始物理观测。
type ObsKind int

const (
	ObsPose            ObsKind = iota // radar Track.Pose 枚举
	ObsKinematics                     // Δz↓ + 位移 + 隐含速度 合成 [0,1]（跌倒运动学签名）
	ObsVitalPresent                   // HR>0 ∨ RR>0
	ObsBedOccupied                    // bed 贝叶斯 P(InBed)，嵌套 belief 子证据
	ObsSleepStage                     // sleepad 睡眠分期（S1 躺细分）
	ObsFirmwareFall                   // firmware pose2→5 升级确认
	ObsEnterExit                      // EnterRoom=+1 / ExitRoom=-1
	ObsNumberPeople                   // 房间人数 0-8
	ObsStandDuration                  // 连续站立分钟（bathroom still-fall 时长门，v2 由 HSMM 替代）
	ObsTrackPresent                   // track Verdict/GhostPenalty 合成 ghost-ness [0,1]
	ObsNeighbor                       // §5.5.2 弱耦合：邻居 room P(占用) [0,1]
	ObsTimeContext                    // 夜/昼 + 房型，调 prior 非硬观测
	ObsNoDetect                       // P(no-detect|s)：本 tick 看了没测到（前置=消失前60s在走动）。状态条件似然=可检测态压低/可合理消失态保留；时长→P3，非斜坡
	ObsReachableExit                  // 丢失点可达退场证据 e=f_dist·f_reach [0,1]（P2 软门：近门+单帧可达→偏 Left 压 Fallen）
)

// Geom 位置语义标签——由 device 坐标 ∩ layout polygon 现算（belief_input_normalization.md §1）。
// belief 隐状态 S 不含坐标，位置只作 P(o|s) 的条件。
type Geom int

const (
	GeomUnknown   Geom = iota
	GeomInBed          // 在 bed polygon 内
	GeomInEnter        // 在 Enter 区内（门口）
	GeomOpenFloor      // 开阔地板（远离 bed/enter）
	GeomInToilet       // 马桶/淋浴区
)

var geomLabel = [...]string{"Unknown", "InBed", "InEnter", "OpenFloor", "InToilet"}

func (g Geom) String() string {
	if int(g) < 0 || int(g) >= len(geomLabel) {
		return "Unknown"
	}
	return geomLabel[g]
}

// Observation 统一七元组 schema（belief_input_normalization.md §1）。
// 所有源（radar/sleepad/room/bed/neighbor）翻成同格式，引擎只认 schema 不认设备类型。
type Observation struct {
	Source string  // 谁产出：device /128 / room /88 / bed /96
	Kind   ObsKind // 决定查 P(o|s) 哪一行
	Value  float64 // 归一化数值（连续）或枚举序号（离散）
	Conf   float64 // 可信度 [0,1]，0=缺失/stale → likelihood=I 不更新
	Ts     int64   // 观测时刻 unix ms
	Fresh  bool    // 在 TTL 内；stale → 当缺失（Conf 视为 0）
	Geom   Geom    // 位置语义（grid 交集算）
	// GeomConf geom 可靠度 [0,1]——provenance 信任权重（cell Source：FE 画=1 / feedback=0.6 / 自学=0.4）。
	// 0 或未设=1.0（全信 geom，向后兼容）。<1 时把 geom 条件似然向 geom-中性(Unknown) blend——
	// 软先验替硬闸：暂定 rest-zone(InBed/InToilet) 抑制跌倒更弱，强跌倒证据仍能盖过，不漏真摔。
	GeomConf float64
}

// effConf 有效置信度：stale 观测当缺失。
func (o Observation) effConf() float64 {
	if !o.Fresh {
		return 0
	}
	return o.Conf
}
