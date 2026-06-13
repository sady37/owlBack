package belief

import "owl-common/observation"

// rawLikelihood 给一条观测算 P(o|s) 的未归一化 likelihood 向量（中性=1.0）。
// 所有 LR 数值集中在 calibration.go（P2.6/R7,带来源,可单点调灵敏度、可审）。
// 对照表见 signal_map §2 发射表。
func rawLikelihood(o Observation) Vector {
	switch o.Kind {
	case ObsPose:
		// provenance 加权：AreaConf<1 时把 area 条件似然向中性(area-unknown) blend。
		// = 暂定 rest-zone 的跌倒抑制按信任打折(软先验替硬 exempt 闸)；FE 画 bed(AreaConf=1)全抑制不变。
		ctx := AreaCtx{AreaType: o.AreaType, NearDoor: o.NearDoor, BedReleased: o.BedReleased}
		gc := o.AreaConf
		if gc <= 0 || gc >= 1 {
			return poseLikelihood(int(o.Value), ctx)
		}
		neutral := AreaCtx{AreaType: areaUnknown}
		return lerpVec(poseLikelihood(int(o.Value), neutral), poseLikelihood(int(o.Value), ctx), gc)
	case ObsZBand:
		// P2.3 z 三档 posture(A类round-z3 / feedback 原则#2)。z 只喂 posture,**绝不写 SFallen**
		// (R5:z 不确认不否决 fall)。与 pose=standing→SStandWalk 同型(P2.2 认可的 posture 通道,非 fall 压制)。
		// z 仅高值可信:>80 直立;30–80 坐;<30 假低(坐/躺也常报低)→ 无信息 lk(nil)。
		switch zcm := o.Value; {
		case zcm > zUprightCm:
			return lk(map[State]float64{SOpenFloor: lrZStand})
		case zcm >= zSitMinCm:
			return lk(map[State]float64{SSit: lrZSit})
		default:
			return lk(nil)
		}
	case ObsDwellStill:
		// dwell 生存函数 ramp（P4.1：−ln S_vol(d|zone)+1）单源走 survival.go fallLRFromDwell（per-zone 尾 +
		// 开阔地 cell tolerance 拉长尾 + P4.3 夜间短尾）。toilet 15min/open 8min×tol；bed/enter/unknown 不报。
		fallLR := fallLRFromDwell(o.Value, o.ToleranceFactor, o.RoomType, o.AreaType, o.Night)
		if fallLR <= 1.0 {
			return lk(nil) // 中性（不在尾表 / dwell≤0）
		}
		return lk(map[State]float64{SFallen: fallLR})
	case ObsVitalPresent:
		// 有生命体征 → 必有真人，压 Empty。
		if o.Value >= 0.5 {
			return lk(map[State]float64{SEmpty: lrVitalEmpty})
		}
		return lk(map[State]float64{SEmpty: lrNoVitalEmpty}) // 无 vital 是弱信号（mmWave 常漏）
	case ObsBedOccupied:
		// 嵌套 bed 贝叶斯输出 P(InBed)。标定对齐 bed scorer LR 表：高占用强抬床态、压地板/站立。
		p := clamp01(o.Value)
		return lk(map[State]float64{
			SBed:       1 + gainBedLying*p,
			SFallen:    1 - dampBedFallen*p,
			SOpenFloor: 1 - dampBedStandWalk*p,
			SSit:       1 - dampBedSit*p,
			SEmpty:     1 - dampBedEmpty*p,
		})
	case ObsEnterExit:
		if o.Value > 0 { // EnterRoom：有人进来
			return lk(map[State]float64{SEmpty: lrEnterEmpty, SOpenFloor: lrEnterStandWalk, SLeft: lrEnterLeft})
		}
		if o.Value < 0 { // ExitRoom：离开，绝非倒地。强 Left（lrExitLeft）是 Blind* 唯一离场出口（A 无逃生阀）；同时压盲区（离场了就不在盲）。
			return lk(map[State]float64{
				SLeft: lrExitLeft, SEmpty: lrExitEmpty, SFallen: lrExitFallen, SOpenFloor: lrExitStandWalk,
				SBlindRest: lrExitBlind, SBlindOpen: lrExitBlind,
			})
		}
		return lk(nil)
	case ObsNumberPeople:
		if o.Value < 0.5 { // number_people=0：弱倾向空/离，**不反驳已倒地**（铁律：np=0 是 corroboration 非
			// substitution——金属垃圾桶/镜子→ghost、淋浴水气→衰减都会假报 np=0；真倒地证据须仍能竞争）。
			// 强 Empty 拉力交 ObsEnterExit<0(SLeft:8) + ObsReachableExit。详 doc/belief_p2_absence_emission §3.3 (#14)。
			return lk(map[State]float64{
				SEmpty: lrNp0Empty, SLeft: lrNp0Left,
				SSit: lrNp0Sit, SOpenFloor: lrNp0StandWalk, SFallen: lrNp0Fallen,
			})
		}
		return lk(map[State]float64{SEmpty: lrNpOccEmpty}) // 有人
	case ObsNeighbor:
		// §5.5.2 弱耦合：邻居 room P(占用) 高 → 本房更可能空/已离，更不可能倒地。
		n := clamp01(o.Value)
		return lk(map[State]float64{
			SEmpty:     1 + gainNbrEmpty*n,
			SLeft:      1 + gainNbrLeft*n,
			SFallen:    1 - dampNbrFallen*n,
			SBed:       1 - dampNbrBedLying*n,
			SOpenFloor: 1 - dampNbrStandWalk*n,
		})
	case ObsNoDetect:
		// P(no-detect|s)：realness 门控后无 real track 解析的状态条件似然（非时长斜坡）。
		// **可见区（Bed/Sit/Open/Bath）"该被 track 解析到却没有" → 压低**；无 track 的合理解释
		// （盲区占用 BlindRest/BlindOpen、倒地、空、已离）保留/抬。Rest vs Open 由 A 从消失前态决定，**发射不分**。
		// **固定强度、不含时长项**——"消失越久 Fallen 越高"由重复观测 × Fallen 近吸收涌现 + Decider 窗，非手调 gain。
		// 方向仲裁交其它观测：reachableExit/ExitRoom→Left、np=0→Empty/Left。
		// P6.1a Fallen 因子门控 1+gain·P(real)·(1−k·P(door-exit))：ghost 消失(RealnessP→0)或门区可达走出
		// (DoorExitP→1)→ →1.0 中性(不裸 absence 抬 fall)；真人非门区消失(1,0)→ 1.6。连续边缘化非硬𝟙。
		ri := clamp01(o.RealnessP)
		dx := clamp01(o.DoorExitP)
		return lk(map[State]float64{
			SBed: lrNoDetVisible, SSit: lrNoDetVisible, SOpenFloor: lrNoDetVisible, SBath: lrNoDetVisible,
			SBlindRest: lrNoDetBlind, SBlindOpen: lrNoDetBlind,
			SFallen: 1 + noDetGainFallen*ri*(1-noDetDoorSuppressK*dx), // door-exit 留 floor(审查㉓:不全否决,门口真摔仍浮出)
			// SEmpty/SLeft 默认 1.0 中性，留给 np=0/exit/neighbor 仲裁
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
	}
	return lk(nil)
}

// poseLikelihood radar pose × 区域上下文条件似然。区域→占用态：area=Bed→SBed / Sit→SSit / Toilet,Shower→SBath /
// Active,Deny→SOpenFloor。关键二义性：pose=Lying 在床区(area=Bed 且床态未释放) → SBed；床态已释放(bed_state 离床)
// 或开阔地/卫浴 → 倒地候选(SFallen)。门优先：NearDoor∨area==Enter 视为门区，压过 area 占用态。
func poseLikelihood(pose int, c AreaCtx) Vector {
	inEnter := c.NearDoor || c.AreaType == areaEnter
	inBed := !inEnter && c.AreaType == areaBed
	inBath := !inEnter && (c.AreaType == areaShower || c.AreaType == areaToilet)
	switch pose {
	case observation.PoseWalking, observation.PoseRunning:
		// P2.2(R5):pose 对 fall 只正向 —— 不压 SFallen。SOpenFloor/SBed 是占用区分非 fall 压制。
		m := map[State]float64{SOpenFloor: lrPoseWalkStandWalk, SBed: dampPoseWalkBed}
		if inEnter {
			m[SLeft] = lrPoseWalkEnterLeft // 门口走动 → 可能正离场
		}
		return lk(m)
	case observation.PoseSuspectedFall:
		return lk(map[State]float64{SFallen: lrPoseSuspFallen, SOpenFloor: dampPoseSuspStand})
	case observation.PoseSitting:
		return lk(map[State]float64{SSit: lrPoseSitSit, SOpenFloor: dampPoseSitStand})
	case observation.PoseStanding:
		// P2.2(R5):stand 对 fall 中性,不否决跌倒。
		return lk(map[State]float64{SOpenFloor: lrPoseStand, SSit: lrPoseStandSit})
	case observation.PoseFallen:
		m := map[State]float64{SFallen: lrPoseFallenBase, SOpenFloor: dampPoseFallenStand, SSit: dampPoseFallenSit}
		switch {
		case inBed:
			m[SFallen] = lrPoseFallenInBed // 床上 pose=fallen 多是躺姿误读，降权
			m[SBed] = lrPoseFallenBedLy
		case inBath || (!inEnter && isOpenFloor(c.AreaType)):
			m[SFallen] = lrPoseFallenOpen // 卫浴/开阔地板确认倒地，升权
		}
		return lk(m)
	case observation.PoseLying:
		switch {
		case inBed && !c.BedReleased:
			// 床上躺=睡，走 SBed。不报跌倒走 cell rest-zone/human-bed 豁免，不靠 pose 压 fall。
			return lk(map[State]float64{SBed: lrPoseLyingBedLying})
		case (!inEnter && isOpenFloor(c.AreaType)) || (inBed && c.BedReleased):
			// 开阔地躺=倒地候选；床区躺但 bed_state 离床(BedReleased)=床边真摔(cd2b 出口,bed_state 权威>几何)。
			return lk(map[State]float64{SFallen: lrPoseLyingOpenFall, SBed: dampPoseLyingOpenBed, SOpenFloor: dampPoseLyingOpenSW})
		case inBath:
			// 卫浴躺=倒地候选(高风险区)。
			return lk(map[State]float64{SFallen: lrPoseLyingBathFall, SBath: dampPoseLyingBath})
		default:
			return lk(map[State]float64{SBed: lrPoseLyingDefBed, SFallen: lrPoseLyingDefFall})
		}
	case observation.PoseSuspectedSitGround, observation.PoseSitGround:
		return lk(map[State]float64{SFallen: lrPoseSitGroundFall, SSit: lrPoseSitGroundSit})
	case observation.PoseBedSitUp, observation.PoseSuspectedBedSitUp, observation.PoseConfirmedBedSitUp:
		return lk(map[State]float64{SBed: lrPoseBedSitUpRest})
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
