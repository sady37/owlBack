package belief

import "math"

// emission.go — §5 联合发射 Φ → log 域 JointVector（喂 filter.Correct 的 logPhi）。
//   Φ_t = Π_j ℓ_sj(o^sj|B^j)^{w_sj}  ·  Π_c ℓ_c(o^c|S)^{w_c}
//          └── 接触→B^j（只依赖 bmask 第 j 位）──┘  └── 雷达 pose/dwell/hrrr→S（只依赖 S）──┘
// 两轴在 log 域可分：logPhi[(S,bmask)] = contactLogB(bmask) + radarLogS(S)。
// 权重作指数（log 域 = w·log ℓ）：w_sj=onbed, w_pose=covers, w_hrrr=covers·1[nearBed]。
//
// 形态锚（铁律 [[fall_data_is_artificial_test]]：单 case 不刻精确参数，只定可判侧形态）：
// L_in/L_left≫1（接触强）、pose lying 二义(AtBed=F 同 boost)、HR/RR 非对称 + §D absent 门控。
// 标定见 feedback-p6C。
//
// 注（§32 拆补丁）：δ floor-strip 已删——cd2b 不靠雷达 XY 精确空间，靠 LeftBed→B vac 经 §4 Ψ 相容
// 涌现 SFallen（C 独立测试 0.995 验证）。floor-strip 是补丁，框架 Ψ 让 cd2b 零补丁涌现。

type emissionParams struct {
	lIn        float64 // ℓ_sj(InBed|occ)=L_in≫1
	lLeft      float64 // ℓ_sj(LeftBed|vac)=L_left≫1（sleepad 接触 LeftBed：果断清，残留~0.02）
	lLeftRadar float64 // radar 几何 LeftBed：弱清（<50% 可信，残留~0.10），lLeftRadar<lLeft
	lPose float64 // ℓ_pose(lying|AtBed)=ℓ_pose(lying|F)>1（二义，刻意）
	lHR   float64 // L_hr：present|AtBed 倍数（absent|AtBed=1/L_hr）
	lArea float64 // area_type 正向压制倍数（bed/sit/toilet→抬对应静止态）。area 误学(如假 Sit)的真摔
	//                  由 FloorGuard 非累加总时长兜底（still 不进 emission，避免前向滤波累积），故 lArea 仅作 redirect 偏置
	// 直立证据（移自 floor stillDiscount，单一归 emission）：分两类施加，取最强(min) 压。
	//   **压 SFallen（<1）⟺ 与"倒地静止"物理互斥**：z≥zStandCm（身高硬测，质心高≠贴地）∨ walk（运动⊥倒地静止：
	//     摔者不动→不会被标 walking→压不到真摔；故 walk 压不要 z 门，纯标签可压）。standing/sit 是静态直立，与
	//     "静态倒地（刚摔未及标 Lying）"共享静止→可混→**只抬不压**（z 未知不赌 fall，z<30 中性红线延续）。
	//   **抬（>1）治 Empty/安置**：在场直立（walk/z≥80/standing）→ 抬 SOpenFloor（否则压腾出的质量归一漏进
	//     默认 Empty=present 真人却判空房 + Empty→Fallen 种子 0 拖慢真摔）；sit → 抬 SSit。
	//   walk∧z≥80：压取 min 一次、抬 SOpenFloor 一次（不叠，防过抬偏 SOpenFloor）。⚠️walk 边界：地板挣扎摔者
	//     （低 z+乱动）可能误标 walk 被压→FN，靠 floor 时长兜底+转 Lying 接住（留 oracle 真挣扎数据验）。
	supWalk  float64 // walk → SFallen ×supWalk（压，运动⊥倒地）
	supStand float64 // z≥zStandCm → SFallen ×supStand（压，身高硬测）
	redOpen  float64 // 在场直立 → SOpenFloor ×redOpen（抬，治 Empty/安置）
	redSit   float64 // sit → SSit ×redSit（抬，弱排斥，不压 SFallen）
}

// zStandCm 站立身高阈：z≥此 → 与"躺地"互斥（摔倒质心 z 必低）。高 bar，单帧噪声 z<80 不误压真摔。
const zStandCm = 80

// roomengine.AreaType 枚举值（belief 不 import roomengine，本地常量对齐）。
const (
	areaSit    = 3
	areaActive = 4
	areaEnter  = 1
	areaDeny   = 5
	areaShower = 6
	areaToilet = 7
)

func defaultEmissionParams() emissionParams {
	return emissionParams{
		lIn: 20, lLeft: 60, lLeftRadar: 8, lPose: 3, lHR: 5,
		lArea:   2,
		supWalk: 0.5, supStand: 0.5,
		redOpen: 2, redSit: 2,
	}
}

// roomBathroom = card.RoomTypeBathroom（belief 不 import card，本地常量对齐）。
const roomBathroom = 1

// stillMuSigma 正常停留 (μ,σ) 秒 = cell area 与 room **保守合并**（取 μ 更大 = 更晚报 = 低 FP，§H）。
//
//	解决「bathroom 房未画 toilet → cell 落 unknown → 用激进 default 过早误报」：bathroom 房至少按 bathsec 兜。
//	bed 区(床, areaType=2)的 cell 走 default(unknown)——床边跌倒靠接触轴(sleepad InBed 压/LeftBed 放行)区分，不改 areaType。
func stillMuSigma(areaType, roomType int) (mu, sigma float64) {
	cMu, cSig := cellMuSigma(areaType)
	rMu, rSig := roomMuSigma(roomType)
	if rMu > cMu { // 取保守（长 μ）那组
		return rMu, rSig
	}
	return cMu, cSig
}

func cellMuSigma(areaType int) (mu, sigma float64) {
	switch areaType {
	case areaSit:
		return MuSitSec, SigmaSitSec
	case areaToilet, areaShower:
		return MuBathSec, SigmaBathSec
	default: // unknown/active/enter/bed → default(unknown)
		return MuDefaultSec, SigmaDefaultSec
	}
}

func roomMuSigma(roomType int) (mu, sigma float64) {
	if roomType == roomBathroom {
		return MuBathSec, SigmaBathSec // bathroom 房保守下限：即使 cell 未画 toilet
	}
	return MuDefaultSec, SigmaDefaultSec
}

// Emission 无状态发射器（参数 + 每床 onbed 权重）。geom 长 numBeds。
type Emission struct {
	p    emissionParams
	geom []BedGeom
}

func NewEmission(geom []BedGeom) *Emission {
	return &Emission{p: defaultEmissionParams(), geom: geom}
}

// LogPhi 一帧发射 → log 域 JointVector。
func (e *Emission) LogPhi(js *JointSpace, o Observation) JointVector {
	radarS := e.radarLogS(o)     // [numStates]：雷达轴对 S 的 log 似然
	contactB := e.contactLogB(o) // [numBeds][numBedStates]：接触轴对各 B^j 的 log 似然
	// sleepad 接触 vital(InBed+HR/RR fresh)→ 活体在垫,抬 SBed。独立于 radar online/NearBed(接触权威；
	// cd2b radar 不返 vital,这是唯一 vital 印证腿)。lHR 复用(与 radar HRRR 同力度)。
	if o.SleepadVitalPresent {
		radarS[SBed] += math.Log(e.p.lHR)
	}
	out := js.NewJointVector()
	for i := 0; i < js.size; i++ {
		s, bmask := js.decode(i)
		v := radarS[s]
		for j := 0; j < js.numBeds; j++ {
			v += contactB[j][bedOf(bmask, j)]
		}
		out[i] = v
	}
	return out
}

// contactLogB 接触轴 → B^j。w_sj=onbed(s_j,j)。NoReport/离线 → ℓ≡1（log 0）。
func (e *Emission) contactLogB(o Observation) [][numBedStates]float64 {
	nb := len(e.geom)
	out := make([][numBedStates]float64, nb)
	for j := 0; j < nb; j++ {
		w := e.geom[j].Onbed
		var rd BedReading
		if j < len(o.Sleepad) {
			rd = o.Sleepad[j]
		}
		switch rd {
		case BedInBed: // ℓ(InBed|occ)=L_in, ℓ(InBed|vac)=1
			out[j][BOcc] = w * math.Log(e.p.lIn)
		case BedLeftBed: // sleepad 接触 LeftBed：果断清 BVac（残留~0.02）
			out[j][BVac] = w * math.Log(e.p.lLeft)
		case BedLeftBedRadar: // radar 几何 LeftBed：弱清（<50% 可信，残留~0.10）
			out[j][BVac] = w * math.Log(e.p.lLeftRadar)
			// BedNoReport: 两态皆 0 = 中性
		}
	}
	return out
}

// radarLogS 雷达轴 → S 的 log 似然（[numStates]）。RadarOnline=false → 全 0 中性。
func (e *Emission) radarLogS(o Observation) [numStates]float64 {
	var logS [numStates]float64
	// still-box 时长不进 emission：per-tick 注入会被前向滤波累积（同一久静证据按帧重复计、清零不释放）。
	// 静止→fall 改由 FloorGuard 非累加消费当前总时长（present-only，engine.go OR verdict）。
	if !o.RadarOnline {
		return logS // 离线=中性
	}
	w := e.geom0Covers() // w_pose=w_δ=covers(r,·)

	// pose lying（二义）：boost AtBed 与 F（不分）。
	if o.PoseLying {
		addLogLk(&logS, Vector{SBed: e.p.lPose, SFallen: e.p.lPose}, w, SBed, SFallen)
	}

	// 床（N×M）：firmware area_id 命中床 areaId（N=1）才抬 SBed，权重 = 该床 covers（M，§F per-bed 不取 max）。
	//   N 是雷达地面真值（不随 canvas drift），替代旧 cell area=Bed 判定。多床命中各用自己床的 covers。
	for j, hit := range o.RadarBedHitMask {
		if hit && j < len(e.geom) {
			addLogLk(&logS, Vector{SBed: e.p.lArea}, e.geom[j].Covers, SBed)
		}
	}

	// area_type 正向压制（每帧读活的 cell）：FN-safe 默认偏 Fallen，由位置正向证据 redirect 到对应静止态。
	//   sit→SSit / toilet·shower→SBath+SSit / active·enter→SOpenFloor；deny·unknown 中性。床走上方 firmware N。
	//   area 误学(如假 Sit)的真摔由 FloorGuard 总时长兜底（不靠 emission still 路径翻 area）。
	switch o.AreaType {
	case areaSit:
		addLogLk(&logS, Vector{SSit: e.p.lArea}, w, SSit)
	case areaToilet, areaShower:
		addLogLk(&logS, Vector{SBath: e.p.lArea, SSit: e.p.lArea}, w, SBath, SSit)
	case areaActive, areaEnter:
		addLogLk(&logS, Vector{SOpenFloor: e.p.lArea}, w, SOpenFloor)
	}

	// 直立证据（移自 floor stillDiscount，单一归 emission）：压 SFallen（仅与"倒地静止"物理互斥者）+ 抬直立态
	//   （治 Empty/安置）。详见 emissionParams 注释。pose=Lying 真倒地 → 下列全不命中 → 不压（lPose ×3 boost 照常）。
	d := 1.0
	if o.PoseWalking { // 运动⊥倒地静止：摔者不动不会被标 walk，纯标签可压（不要 z 门）
		d = math.Min(d, e.p.supWalk)
	}
	if o.Z >= zStandCm { // 身高硬测：质心 z≥80 与"贴地"互斥
		d = math.Min(d, e.p.supStand)
	}
	if d < 1.0 {
		addLogLk(&logS, Vector{SFallen: d}, w, SFallen) // 压（取最强 min）
	}
	// 抬：在场直立 → SOpenFloor（present-upright 兜底，否则归一漏进 Empty=判空房 + Empty→Fallen 种子 0 拖慢真摔）。
	//   软站(Standing,z<80)只抬不压（z 未知不赌）；walk∧z≥80 也只抬一次（不叠）。
	if o.PoseWalking || o.Z >= zStandCm || o.PoseStanding {
		addLogLk(&logS, Vector{SOpenFloor: e.p.redOpen}, w, SOpenFloor)
	}
	// 坐=弱排斥：只抬 SSit（不压 SFallen——坐着可能下刻滑倒，软 redirect 保边界报路）。
	if o.PoseSit {
		addLogLk(&logS, Vector{SSit: e.p.redSit}, w, SSit)
	}

	// HR/RR（非对称 + §D 门控）：w_hrrr=covers·1[nearBed]。
	if o.NearBed && o.HRRRObserved {
		if o.HRRRPresent {
			// present|AtBed=L_hr；present|¬AtBed=1/L_hr（vital 近床 → 在床非摔/非空）。
			pv := uniformVec(1.0 / e.p.lHR)
			pv[SBed] = e.p.lHR
			addLogLkAll(&logS, pv, w)
		} else if o.VitalSourceOnline {
			// §D：absent 否决 AtBed 须 gate 在独立在线 vital 源下；radar 自身 absent=零信息不否决。
			// absent|AtBed=1/L_hr；absent|¬AtBed=1（不指定替代，F/Sit/O 交 pose/dwell）。
			av := uniformVec(1.0)
			av[SBed] = 1.0 / e.p.lHR
			addLogLkAll(&logS, av, w)
		}
	}

	return logS
}

// geom0Covers 雷达对本房床的覆盖权重 w=covers。多床取 max_j covers（C2 取定，DBN-Zone-Room §F）：
// 保 Φ 的 S/B 分轴清洁（雷达轴只挂 S 不按床分解）；代价=多床 coverage 高估，留 per-state covers 作触发式退路。
// 远边缘/看不见 → 小 → 雷达轴自然弱化。numBeds=0（无床房）仍允许雷达轴（covers 取 1 全权）。
func (e *Emission) geom0Covers() float64 {
	if len(e.geom) == 0 {
		return 1.0
	}
	mx := 0.0
	for _, g := range e.geom {
		if g.Covers > mx {
			mx = g.Covers
		}
	}
	return mx
}

// addLogLk 把似然向量 comp（仅 keys 给定态非中性，其余视为 1）按权重 w 加进 logS。
func addLogLk(logS *[numStates]float64, comp Vector, w float64, keys ...State) {
	for _, s := range keys {
		logS[s] += w * math.Log(comp[s])
	}
}

// addLogLkAll 把完整似然向量 comp（每态都有意义）按权重 w 加进 logS。
func addLogLkAll(logS *[numStates]float64, comp Vector, w float64) {
	for s := 0; s < numStates; s++ {
		logS[s] += w * math.Log(comp[s])
	}
}

// uniformVec 全态置 v 的向量。
func uniformVec(v float64) Vector {
	var out Vector
	for i := range out {
		out[i] = v
	}
	return out
}
