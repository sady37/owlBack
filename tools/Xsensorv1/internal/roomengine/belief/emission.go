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
	lIn      float64 // ℓ_sj(InBed|occ)=L_in≫1
	lLeft    float64 // ℓ_sj(LeftBed|vac)=L_left≫1
	lPose    float64 // ℓ_pose(lying|AtBed)=ℓ_pose(lying|F)>1（二义，刻意）
	lHR      float64 // L_hr：present|AtBed 倍数（absent|AtBed=1/L_hr）
	lArea    float64 // area_type 正向压制倍数（bed/sit/toilet→抬对应静止态）。area 误学(如假 Sit)的真摔
	//                  由 still 高斯 CDF 兜底（异常阈 μ+1.5σ 翻 area redirect），故 lArea 仅作 redirect 偏置
	lStill   float64 // still 高斯 CDF → SFallen 的 logit 权重（异常度强度，待标定 oracle）
}

// still 高斯 CDF 的 CDF 钳位（防 odds→0/∞ 累积爆，§H）。
const (
	stillCDFMin = 0.02
	stillCDFMax = 0.98
)

// roomengine.AreaType 枚举值（belief 不 import roomengine，本地常量对齐）。
const (
	areaBed    = 2
	areaSit    = 3
	areaActive = 4
	areaEnter  = 1
	areaDeny   = 5
	areaShower = 6
	areaToilet = 7
)

func defaultEmissionParams() emissionParams {
	return emissionParams{
		lIn: 20, lLeft: 20, lPose: 3, lHR: 5,
		lArea: 2, lStill: 0.1,
	}
}

// normalCDF 标准正态 CDF Φ(x)=½(1+erf(x/√2))。
func normalCDF(x float64) float64 { return 0.5 * (1 + math.Erf(x/math.Sqrt2)) }

// roomBathroom = card.RoomTypeBathroom（belief 不 import card，本地常量对齐）。
const roomBathroom = 1

// stillMuSigma 正常停留 (μ,σ) 秒 = cell area 与 room **保守合并**（取 μ 更大 = 更晚报 = 低 FP，§H）。
//   解决「bathroom 房未画 toilet → cell 落 unknown → 用激进 default 过早误报」：bathroom 房至少按 bathsec 兜。
//   bed 区(areaBed)的 cell 走 default(unknown)——床边跌倒靠接触轴(sleepad InBed 压/LeftBed 放行)区分，不改 areaType。
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
		case BedLeftBed: // ℓ(LeftBed|vac)=L_left, ℓ(LeftBed|occ)=1
			out[j][BVac] = w * math.Log(e.p.lLeft)
			// BedNoReport: 两态皆 0 = 中性
		}
	}
	return out
}

// radarLogS 雷达轴 → S 的 log 似然（[numStates]）。RadarOnline=false → 全 0 中性。
func (e *Emission) radarLogS(o Observation) [numStates]float64 {
	var logS [numStates]float64
	// still-box 时长 → SFallen 异常度（per-area 高斯 CDF，§H；取代旧"不进 emission"+FloorGuard 独立兜底）。
	//   **独立于 RadarOnline**：lost 续算 StillSec(① 冻结坐标)在消失态仍是有效"持续静止"证据，须照常喂。
	//   ℓ=odds(CDF)，CDF=Φ((StillSec−μ)/σ)：正常区<0.5 压 SFallen、=μ 中性、异常阈(μ+1.5σ)=0.93 推。
	//   per-area (μ,σ) 解耦（伪迹区短静止 CDF<0.5 自压防 d523 FP）；bed 区走 default(unknown,接触轴区分)；钳位防累积爆。
	//   权重 lStill 独立(不乘 covers)——静止异常度与床覆盖无关。
	if o.StillSec > 0 {
		mu, sigma := stillMuSigma(o.AreaType, o.RoomType)
		cdf := normalCDF((o.StillSec - mu) / sigma)
		if cdf < stillCDFMin {
			cdf = stillCDFMin
		} else if cdf > stillCDFMax {
			cdf = stillCDFMax
		}
		addLogLk(&logS, Vector{SFallen: cdf / (1 - cdf)}, e.p.lStill, SFallen)
	}

	if !o.RadarOnline {
		return logS // 雷达 pose/z/area redirect 中性；still CDF 已加（消失态续算仍喂）
	}
	w := e.geom0Covers() // w_pose=w_δ=covers(r,·)

	// pose lying（二义）：boost AtBed 与 F（不分）。
	if o.PoseLying {
		addLogLk(&logS, Vector{SBed: e.p.lPose, SFallen: e.p.lPose}, w, SBed, SFallen)
	}

	// area_type 正向压制（每帧读活的 cell）：FN-safe 默认偏 Fallen，由位置正向证据 redirect 到对应静止态。
	//   bed→SBed / sit→SSit / toilet·shower→SBath+SSit / active·enter→SOpenFloor；deny·unknown 中性。
	//   area 误学(如假 Sit)的真摔由 FloorGuard 总时长兜底（不靠 emission still 路径翻 area）。
	switch o.AreaType {
	case areaBed:
		addLogLk(&logS, Vector{SBed: e.p.lArea}, w, SBed)
	case areaSit:
		addLogLk(&logS, Vector{SSit: e.p.lArea}, w, SSit)
	case areaToilet, areaShower:
		addLogLk(&logS, Vector{SBath: e.p.lArea, SSit: e.p.lArea}, w, SBath, SSit)
	case areaActive, areaEnter:
		addLogLk(&logS, Vector{SOpenFloor: e.p.lArea}, w, SOpenFloor)
	}

	// z 直立证据已并入 still-box 折扣（track_manager.stillDiscount：z 坐×0.9 / z 站×0.5 削有效 still 时长），
	// 不再走 emission redirect——同一 z 既压 emission 又削 still = 双压站立瘫倒 → 过压漏报（C 裁定撤 ZBand）。

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
