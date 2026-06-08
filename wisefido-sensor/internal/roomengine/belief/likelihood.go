package belief

import "owl-common/observation"

// rawLikelihood 给一条观测算 P(o|s) 的未归一化 likelihood 向量（中性=1.0）。
// 所有 LR 数值集中在 calibration.go（P2.6/R7,带来源,可单点调灵敏度、可审）。
// 对照表见 signal_map §2 发射表。
func rawLikelihood(o Observation) Vector {
	switch o.Kind {
	case ObsPose:
		// provenance 加权：GeomConf<1 时把 geom 条件似然向 geom-中性(Unknown) blend。
		// = 暂定 rest-zone 的跌倒抑制按信任打折(软先验替硬 exempt 闸)；FE 画 bed(GeomConf=1)全抑制不变。
		gc := o.GeomConf
		if gc <= 0 || gc >= 1 {
			return poseLikelihood(int(o.Value), o.Geom)
		}
		return lerpVec(poseLikelihood(int(o.Value), GeomUnknown), poseLikelihood(int(o.Value), o.Geom), gc)
	case ObsZBand:
		// P2.3 z 三档 posture(A类round-z3 / feedback 原则#2)。z 只喂 posture,**绝不写 SFallen**
		// (R5:z 不确认不否决 fall)。与 pose=standing→SStandWalk 同型(P2.2 认可的 posture 通道,非 fall 压制)。
		// z 仅高值可信:>80 直立;30–80 坐;<30 假低(坐/躺也常报低)→ 无信息 lk(nil)。
		switch zcm := o.Value; {
		case zcm > zUprightCm:
			return lk(map[State]float64{SStandWalk: lrZStand})
		case zcm >= zSitMinCm:
			return lk(map[State]float64{SSit: lrZSit})
		default:
			return lk(nil)
		}
	case ObsDwellStill:
		// dwell 生存函数 ramp S_vol(d|zone)=exp(−(d/scale)²) → fallLR=1+(d/scale)²(封顶),**平滑取代硬悬崖**。
		// geom-条件 scale(本 switch 一处,Value 恒 raw still 秒):
		//   • toilet/shower(P4.1 裁决⑮ B):Z_cell-无关,scale=dwellScaleToiletSec。
		//   • 开阔地(P4.4 裁决⑱ B2):**前置 Z_cell tolerance gate**,scale=dwellScaleOpenSec×tol —— 被容忍久站
		//     的 cell 尾拉长→久站真人不报;tol 走 cell 几何/历史(R3 只读),**Z 只正向 R5**,不用 z 反向压。
		//   • rest/bed/enter/unknown:久驻正常,不报(unknown 无 cell→无 tolerance 依据,保守不报防裸 ramp FP)。
		if o.Value <= 0 {
			return lk(nil)
		}
		var scale float64
		switch o.Geom {
		case GeomInToilet:
			scale = dwellScaleToiletSec
		case GeomOpenFloor:
			tol := o.ToleranceFactor
			if tol < 1.0 {
				tol = 1.0 // 默认/未设=1.0(无容忍证据→不放宽)
			}
			scale = dwellScaleOpenSec * tol
		default:
			return lk(nil) // bed/enter/unknown → rest 或无依据,不报
		}
		r := o.Value / scale
		fallLR := 1 + r*r
		if fallLR > dwellFallCap {
			fallLR = dwellFallCap
		}
		return lk(map[State]float64{SFallen: fallLR})
	case ObsVitalPresent:
		// 有生命体征 → 必有真人，压 Empty/Artifact。
		if o.Value >= 0.5 {
			return lk(map[State]float64{SEmpty: lrVitalEmpty, SArtifact: lrVitalArtifact})
		}
		return lk(map[State]float64{SEmpty: lrNoVitalEmpty}) // 无 vital 是弱信号（mmWave 常漏）
	case ObsBedOccupied:
		// 嵌套 bed 贝叶斯输出 P(InBed)。标定对齐 bed scorer LR 表：高占用强抬床态、压地板/站立。
		p := clamp01(o.Value)
		return lk(map[State]float64{
			SBedLying:    1 + gainBedLying*p,
			SBedRestless: 1 + gainBedRestless*p,
			SFallen:      1 - dampBedFallen*p,
			SStandWalk:   1 - dampBedStandWalk*p,
			SSit:         1 - dampBedSit*p,
			SEmpty:       1 - dampBedEmpty*p,
		})
	case ObsSleepStage:
		// stage 0/8=过渡/醒 → Restless；1/2/3=睡 → Lying。粗调，conf 偏低。
		st := int(o.Value)
		if st == 0 || st == 8 {
			return lk(map[State]float64{SBedRestless: lrSleepRestless})
		}
		return lk(map[State]float64{SBedLying: lrSleepLying})
	case ObsFirmwareFall:
		// P2.4(§10#2/R5):firmware pose=5 是 pose 派生(不可信),原 ×10 极权与"pose 不可信"矛盾。
		// shadow 期降到保守 SFallen:2(委员会#5 ≤×2),且**不再强压** SStandWalk/SBedLying ——
		// 降权信号不该用 off-diag 压制竞争真态(误 firmware-fall 时让 stand/bed 证据能竞争)。
		// 终值待 P9 用 firmware_fall 真机 TP/FP 率标定;现网 Device_ALARM 直发不动(R1,本改仅 shadow LR)。
		if o.Value >= 0.5 {
			return lk(map[State]float64{SFallen: lrFirmwareFallen})
		}
		return lk(nil)
	case ObsEnterExit:
		if o.Value > 0 { // EnterRoom：有人进来
			return lk(map[State]float64{SEmpty: lrEnterEmpty, SStandWalk: lrEnterStandWalk, SLeft: lrEnterLeft})
		}
		if o.Value < 0 { // ExitRoom：离开，绝非倒地
			return lk(map[State]float64{SLeft: lrExitLeft, SEmpty: lrExitEmpty, SFallen: lrExitFallen, SStandWalk: lrExitStandWalk})
		}
		return lk(nil)
	case ObsNumberPeople:
		if o.Value < 0.5 { // number_people=0：弱倾向空/离，**不反驳已倒地**（铁律：np=0 是 corroboration 非
			// substitution——金属垃圾桶/镜子→ghost、淋浴水气→衰减都会假报 np=0；真倒地证据须仍能竞争）。
			// 强 Empty 拉力交 ObsEnterExit<0(SLeft:8) + ObsReachableExit。详 doc/belief_p2_absence_emission §3.3 (#14)。
			return lk(map[State]float64{
				SEmpty: lrNp0Empty, SLeft: lrNp0Left,
				SSit: lrNp0Sit, SStandWalk: lrNp0StandWalk, SFallen: lrNp0Fallen,
			})
		}
		return lk(map[State]float64{SEmpty: lrNpOccEmpty}) // 有人
	case ObsStandDuration:
		// bathroom 内久站静止 → still-fall 嫌疑（弱）。v2 由 HSMM 状态时长替代。
		if o.Geom == GeomInToilet && o.Value > 0 {
			return lk(map[State]float64{SFallen: standFallBase + gainStandFall*clampCap(o.Value, standCapMin)})
		}
		return lk(nil)
	case ObsTrackPresent:
		// Value = ghost-ness [0,1]（GhostPenalty/Verdict 合成）。高 → Artifact。
		g := clamp01(o.Value)
		return lk(map[State]float64{
			SArtifact:  1 + gainGhostArtifact*g,
			SStandWalk: 1 - dampGhostStandWalk*g,
			SFallen:    1 - dampGhostFallen*g,
		})
	case ObsNeighbor:
		// §5.5.2 弱耦合：邻居 room P(占用) 高 → 本房更可能空/已离，更不可能倒地。
		n := clamp01(o.Value)
		return lk(map[State]float64{
			SEmpty:     1 + gainNbrEmpty*n,
			SLeft:      1 + gainNbrLeft*n,
			SFallen:    1 - dampNbrFallen*n,
			SBedLying:  1 - dampNbrBedLying*n,
			SStandWalk: 1 - dampNbrStandWalk*n,
		})
	case ObsNoDetect:
		// P(no-detect|s)：本 tick "看了没测到" 的状态条件似然（非时长斜坡）。
		// 机理：**可检测态（站/走/坐）凭空消失是坏解释 → 压低**；可合理消失态（倒地贴地遮挡、空房、
		// 已离、ghost 闪灭、床上被遮）保留/略升。走动者突然消失又不在 Empty/Left → Fallen 来解释它。
		// **关键：固定强度、不含时长项**——时长（要维持多久才确认）归 Decider/P3 HSMM，非本发射（P2/P3 切割）。
		// 方向仲裁交其它观测：reachableExit→Left、np=0→Empty/Left、ExitRoom→Left、verdict→Artifact；
		// 且 A 禁 StandWalk→Empty 直达（Empty 必经 Left）→ 无退场证据时 Fallen 胜出，有则被压回。
		// 每 tick 施加，"消失越久 Fallen 越高" 由重复观测 × Fallen 近吸收涌现，非手调 gain。
		// P6.1a(阻塞项#1):Fallen 因子门控——不裸 absence 抬 fall。1+gain·P(real)·(1−P(door-exit))：
		// ghost 消失(RealnessP→0)或门区可达走出(DoorExitP→1)→ →1.0 中性(退场由 SLeft/SEmpty 仲裁);
		// 真人非门区消失(1,0)→ 1.6(=旧上限)。连续边缘化非硬𝟙(realness 判错平滑退化)。
		ri := clamp01(o.RealnessP)
		dx := clamp01(o.DoorExitP)
		return lk(map[State]float64{
			SStandWalk:   lrNoDetStandWalk,
			SSit:         lrNoDetSit,
			SBedRestless: lrNoDetBedRestless,
			SBedLying:    lrNoDetBedLying,
			SFallen:      1 + noDetGainFallen*ri*(1-noDetDoorSuppressK*dx), // door-exit 留 floor(审查㉓:不全否决,门口真摔仍浮出)
			// SEmpty/SLeft/STransition/SArtifact 默认 1.0 中性，留给其它观测/转移仲裁
		})
	case ObsReachableExit:
		// 可达退场证据 e=f_dist·f_reach。高（近门 + 单帧可达）→ 人很可能从门口走出（Left），不是倒地 →
		// 压 Fallen。e≈0（远离门/不可达，如真跌倒在开阔地板）→ 各因子→1 = identity，不干预。
		// 替 30cm 硬门闸的悬崖（#2/#8 软化）。详 doc/belief_p2_absence_emission §3.1-3.2。
		e := clamp01(o.Value)
		return lk(map[State]float64{
			SLeft:   1 + gainReachLeft*e,
			SEmpty:  1 + gainReachEmpty*e,
			SFallen: 1 - dampReachFallen*e,
		})
	case ObsTimeContext:
		return lk(nil) // 调 prior / θ_fire，不在 diag 更新（见 Model.prior）
	}
	return lk(nil)
}

// poseLikelihood radar pose × Geom 条件似然。
// 关键二义性：pose=Lying 在 InBed → Bed-Lying；在 OpenFloor → 倒地候选。Geom 是命门。
func poseLikelihood(pose int, g Geom) Vector {
	switch pose {
	case observation.PoseWalking, observation.PoseRunning:
		// P2.2(R5):pose 对 fall 只正向 —— 删 SFallen 压制(firmware 误把摔后标 walk 不得抹 fall)。
		// 保留 SStandWalk/SBedLying(posture 区分,非 fall 压制)。
		m := map[State]float64{SStandWalk: lrPoseWalkStandWalk, SBedLying: dampPoseWalkBed}
		if g == GeomInEnter {
			m[SLeft] = lrPoseWalkEnterLeft // 门口走动 → 可能正离场
		}
		return lk(m)
	case observation.PoseSuspectedFall:
		return lk(map[State]float64{SFallen: lrPoseSuspFallen, STransition: lrPoseSuspTrans, SStandWalk: dampPoseSuspStand})
	case observation.PoseSitting:
		return lk(map[State]float64{SSit: lrPoseSitSit, SStandWalk: dampPoseSitStand})
	case observation.PoseStanding:
		// P2.2(R5):删 SFallen:0.4 压制 —— stand 对 fall 中性,不否决跌倒。
		return lk(map[State]float64{SStandWalk: lrPoseStand, SSit: lrPoseStandSit})
	case observation.PoseFallen:
		m := map[State]float64{SFallen: lrPoseFallenBase, SStandWalk: dampPoseFallenStand, SSit: dampPoseFallenSit}
		if g == GeomInBed {
			m[SFallen] = lrPoseFallenInBed // 床上 pose=fallen 多是躺姿误读，降权
			m[SBedLying] = lrPoseFallenBedLy
		} else if g == GeomOpenFloor {
			m[SFallen] = lrPoseFallenOpen // 开阔地板确认倒地，升权
		}
		return lk(m)
	case observation.PoseLying:
		switch g {
		case GeomInBed:
			// P2.2(R5):删 SFallen:0.3 压制 —— 床上不报跌倒走 cell rest-zone/human-bed 豁免(P7.4),
			// 不靠 pose 压 fall。SBedLying:6 强抬床态自然主导。
			return lk(map[State]float64{SBedLying: lrPoseLyingBedLying, SBedRestless: lrPoseLyingBedRest})
		case GeomOpenFloor:
			return lk(map[State]float64{SFallen: lrPoseLyingOpenFall, SBedLying: dampPoseLyingOpenBed, SStandWalk: dampPoseLyingOpenSW})
		default:
			return lk(map[State]float64{SBedLying: lrPoseLyingDefBed, SFallen: lrPoseLyingDefFall})
		}
	case observation.PoseSuspectedSitGround, observation.PoseSitGround:
		return lk(map[State]float64{SFallen: lrPoseSitGroundFall, SSit: lrPoseSitGroundSit})
	case observation.PoseBedSitUp, observation.PoseSuspectedBedSitUp, observation.PoseConfirmedBedSitUp:
		return lk(map[State]float64{SBedRestless: lrPoseBedSitUpRest, SBedLying: lrPoseBedSitUpLying})
	}
	return lk(nil)
}

// lerpVec 线性插值 a→b（t∈[0,1]）：out = a + t·(b−a)。geom provenance 加权用。
func lerpVec(a, b Vector, t float64) Vector {
	var out Vector
	for i := range out {
		out[i] = a[i] + t*(b[i]-a[i])
	}
	return out
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func clampCap(x, cap float64) float64 {
	if x > cap {
		return cap
	}
	if x < 0 {
		return 0
	}
	return x
}
