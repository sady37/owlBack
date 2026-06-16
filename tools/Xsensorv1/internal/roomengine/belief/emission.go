package belief

import "math"

// emission.go — §5 联合发射 Φ → log 域 JointVector（喂 filter.Correct 的 logPhi）。
//   Φ_t = Π_j ℓ_sj(o^sj|B^j)^{w_sj}  ·  Π_c ℓ_c(o^c|S)^{w_c}
//          └── 接触→B^j（只依赖 bmask 第 j 位）──┘  └── 雷达 pose/dwell/hrrr→S（只依赖 S）──┘
// 两轴在 log 域可分：logPhi[(S,bmask)] = contactLogB(bmask) + radarLogS(S)。
// 权重作指数（log 域 = w·log ℓ）：w_sj=onbed, w_pose=w_dwell=covers, w_hrrr=covers·1[nearBed]。
//
// 形态锚（铁律 [[fall_data_is_artificial_test]]：单 case 不刻精确参数，只定可判侧形态）：
// L_in/L_left≫1（接触强）、pose lying 二义(AtBed=F 同 boost)、HR/RR 非对称 + §D absent 门控、
// dwell still≥τ 只分静止占用 vs 活动。标定见 feedback-p6C。
//
// 注（§32 拆补丁）：δ floor-strip 已删——cd2b 不靠雷达 XY 精确空间，靠 LeftBed→B vac 经 §4 Ψ 相容
// 涌现 SFallen（C 独立测试 0.995 验证）。floor-strip 是补丁，框架 Ψ 让 cd2b 零补丁涌现。

type emissionParams struct {
	lIn      float64 // ℓ_sj(InBed|occ)=L_in≫1
	lLeft    float64 // ℓ_sj(LeftBed|vac)=L_left≫1
	lPose    float64 // ℓ_pose(lying|AtBed)=ℓ_pose(lying|F)>1（二义，刻意）
	lHR      float64 // L_hr：present|AtBed 倍数（absent|AtBed=1/L_hr）
	dwellHi  float64 // D>1：still≥τ|F = still≥τ|AtBed
	dwellLo  float64 // <1：still≥τ|O/Sit（活动态久静受罚）
	stillTau float64 // dwell 阈 τ（秒）；cell 容忍/夜间对 τ 的调制属 decide/adapter 层（feedback-p6C §6/§7）
	lZ       float64 // ObsZBand 正向抬直立态(Sit/OpenFloor)的倍数；须 ≳ dwellHi/dwellLo 抵消久坐 dwell 误判
}

func defaultEmissionParams() emissionParams {
	return emissionParams{
		lIn: 20, lLeft: 20, lPose: 3, lHR: 5,
		dwellHi: 3, dwellLo: 0.5, stillTau: 60, lZ: 8,
	}
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
	radarS := e.radarLogS(o)        // [numStates]：雷达轴对 S 的 log 似然
	contactB := e.contactLogB(o)    // [numBeds][numBedStates]：接触轴对各 B^j 的 log 似然
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
	if !o.RadarOnline {
		return logS // 离线=中性
	}
	w := e.geom0Covers() // w_pose=w_dwell=w_δ=covers(r,·)

	// pose lying（二义）：boost AtBed 与 F（不分）。
	if o.PoseLying {
		addLogLk(&logS, Vector{SBed: e.p.lPose, SFallen: e.p.lPose}, w, SBed, SFallen)
	}

	// dwell still≥τ：静止占用(F/AtBed) D>1；活动态(O/Sit) <1。dwell 不分 F/AtBed。
	if o.StillSec >= e.p.stillTau {
		addLogLk(&logS, Vector{
			SBed: e.p.dwellHi, SFallen: e.p.dwellHi,
			SOpenFloor: e.p.dwellLo, SSit: e.p.dwellLo,
		}, w, SBed, SFallen, SOpenFloor, SSit)
	}

	// z 高度档(ObsZBand)：**单向正向证据，绝不负向**（device-room-zone.md）。坐高(30-60)→抬 Sit、
	//   站高(>60)→抬 OpenFloor，抵消 dwell 对久坐马桶的误判(dwell 罚 Sit·抬 Fallen)；
	//   贴地/低→ZNone 中性（z=0 不是 fall 证据、不否决任何东西，fall 仍走 dwell）。时间积分=前向滤波逐帧累积。
	switch o.ZBand {
	case ZSit:
		addLogLk(&logS, Vector{SSit: e.p.lZ}, w, SSit)
	case ZStand:
		addLogLk(&logS, Vector{SOpenFloor: e.p.lZ}, w, SOpenFloor)
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
