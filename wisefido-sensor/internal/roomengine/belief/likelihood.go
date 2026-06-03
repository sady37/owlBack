package belief

import "owl-common/observation"

// rawLikelihood 给一条观测算 P(o|s) 的未归一化 likelihood 向量（中性=1.0）。
// 标定来源：fall_rules_three_classes（pose/运动学阈值转似然）、bed_bayesian_scorer LR 表
// （ObsBedOccupied）、firmware_fall_qualification（ObsFirmwareFall）。
// 每条 case 的具体数值是"白盒条目"——可单点调灵敏度，对照表见 belief_gate_to_matrix.md。
func rawLikelihood(o Observation) Vector {
	switch o.Kind {
	case ObsPose:
		return poseLikelihood(int(o.Value), o.Geom)
	case ObsKinematics:
		// 跌倒运动学签名 [0,1]：z 骤降 + 随后静止。高 → S5；中等运动 → S4。
		f := clamp01(o.Value)
		return lk(map[State]float64{
			SFallen:    1 + 7*f,
			SStandWalk: 1 + 1.5*(1-f),
			SBedLying:  1 - 0.5*f,
			SEmpty:     1 - 0.5*f,
		})
	case ObsVitalPresent:
		// 有生命体征 → 必有真人，压 Empty/Artifact。
		if o.Value >= 0.5 {
			return lk(map[State]float64{SEmpty: 0.2, SArtifact: 0.3})
		}
		return lk(map[State]float64{SEmpty: 1.2}) // 无 vital 是弱信号（mmWave 常漏）
	case ObsBedOccupied:
		// 嵌套 bed 贝叶斯输出 P(InBed)。标定对齐 bed scorer LR 表：高占用强抬床态、压地板/站立。
		p := clamp01(o.Value)
		return lk(map[State]float64{
			SBedLying:    1 + 5*p,
			SBedRestless: 1 + 3*p,
			SFallen:      1 - 0.7*p,
			SStandWalk:   1 - 0.6*p,
			SSit:         1 - 0.5*p,
			SEmpty:       1 - 0.5*p,
		})
	case ObsSleepStage:
		// stage 0/8=过渡/醒 → Restless；1/2/3=睡 → Lying。粗调，conf 偏低。
		st := int(o.Value)
		if st == 0 || st == 8 {
			return lk(map[State]float64{SBedRestless: 2})
		}
		return lk(map[State]float64{SBedLying: 2})
	case ObsFirmwareFall:
		// firmware 30~90s pose2→5 自带升级，最强 Fall 证据（Option C 不重判）。
		if o.Value >= 0.5 {
			return lk(map[State]float64{SFallen: 10, SStandWalk: 0.3, SBedLying: 0.3})
		}
		return lk(nil)
	case ObsEnterExit:
		if o.Value > 0 { // EnterRoom：有人进来
			return lk(map[State]float64{SEmpty: 0.2, SStandWalk: 2, SLeft: 0.3})
		}
		if o.Value < 0 { // ExitRoom：离开，绝非倒地
			return lk(map[State]float64{SLeft: 8, SEmpty: 2, SFallen: 0.2, SStandWalk: 0.5})
		}
		return lk(nil)
	case ObsNumberPeople:
		if o.Value < 0.5 { // number_people=0：弱倾向空/离，**不反驳已倒地**（铁律：np=0 是 corroboration 非
			// substitution——金属垃圾桶/镜子→ghost、淋浴水气→衰减都会假报 np=0；真倒地证据须仍能竞争）。
			// 强 Empty 拉力交 ObsEnterExit<0(SLeft:8) + ObsReachableExit。详 doc/belief_p2_absence_emission §3.3 (#14)。
			return lk(map[State]float64{
				SEmpty: 1.5, SLeft: 2,
				SSit: 0.5, SStandWalk: 0.5, SFallen: 1.0,
			})
		}
		return lk(map[State]float64{SEmpty: 0.3}) // 有人
	case ObsStandDuration:
		// bathroom 内久站静止 → still-fall 嫌疑（弱）。v2 由 HSMM 状态时长替代。
		if o.Geom == GeomInToilet && o.Value > 0 {
			return lk(map[State]float64{SFallen: 1 + 0.4*clampCap(o.Value, 8)})
		}
		return lk(nil)
	case ObsTrackPresent:
		// Value = ghost-ness [0,1]（GhostPenalty/Verdict 合成）。高 → Artifact。
		g := clamp01(o.Value)
		return lk(map[State]float64{
			SArtifact:  1 + 6*g,
			SStandWalk: 1 - 0.5*g,
			SFallen:    1 - 0.5*g,
		})
	case ObsNeighbor:
		// §5.5.2 弱耦合：邻居 room P(占用) 高 → 本房更可能空/已离，更不可能倒地。
		n := clamp01(o.Value)
		return lk(map[State]float64{
			SEmpty:     1 + 3*n,
			SLeft:      1 + 2*n,
			SFallen:    1 - 0.7*n,
			SBedLying:  1 - 0.4*n,
			SStandWalk: 1 - 0.4*n,
		})
	case ObsLostWhileMoving:
		// 走动中突然消失 → 候选倒地（走动→跌倒+遮挡→丢 track）。前置（消失前60s在走动）由
		// engine/adapter 把关；此处只管似然：抬 P(Fallen)、压走动/空/离场。
		// neighbor/exit/返回 证据通过各自似然压低 Fallen 对冲——无对冲 → fire；有对冲（走出/返回）→ 不 fire。
		// Value = 丢失后时长归一 [0,1]（ramp，非帧率依赖）。
		a := clamp01(o.Value)
		return lk(map[State]float64{
			SFallen:    1 + lostMovingFallGain*a,
			SStandWalk: 1 - 0.3*a,
			SEmpty:     1 - 0.3*a,
			SLeft:      1 - 0.3*a,
		})
	case ObsReachableExit:
		// 可达退场证据 e=f_dist·f_reach。高（近门 + 单帧可达）→ 人很可能从门口走出（Left），不是倒地 →
		// 压 Fallen。e≈0（远离门/不可达，如真跌倒在开阔地板）→ 各因子→1 = identity，不干预。
		// 替 30cm 硬门闸的悬崖（#2/#8 软化）。详 doc/belief_p2_absence_emission §3.1-3.2。
		e := clamp01(o.Value)
		return lk(map[State]float64{
			SLeft:   1 + 6*e,
			SEmpty:  1 + 2*e,
			SFallen: 1 - 0.9*e,
		})
	case ObsTimeContext:
		return lk(nil) // 调 prior / θ_fire，不在 diag 更新（见 Model.prior）
	}
	return lk(nil)
}

// lostMovingFallGain ObsLostWhileMoving 满龄（Value=1）时对 Fallen 的似然增益。
// 标定：open-floor 冻结无对冲 → 越过 θ_fire；有 neighbor(occ .9) 对冲 → 压回 θ_fire 下。
const lostMovingFallGain = 3.0

// poseLikelihood radar pose × Geom 条件似然。
// 关键二义性：pose=Lying 在 InBed → Bed-Lying；在 OpenFloor → 倒地候选。Geom 是命门。
func poseLikelihood(pose int, g Geom) Vector {
	switch pose {
	case observation.PoseWalking, observation.PoseRunning:
		m := map[State]float64{SStandWalk: 6, SFallen: 0.3, SBedLying: 0.3}
		if g == GeomInEnter {
			m[SLeft] = 3 // 门口走动 → 可能正离场
		}
		return lk(m)
	case observation.PoseSuspectedFall:
		return lk(map[State]float64{SFallen: 3, STransition: 2, SStandWalk: 0.8})
	case observation.PoseSitting:
		return lk(map[State]float64{SSit: 6, SStandWalk: 0.8})
	case observation.PoseStanding:
		return lk(map[State]float64{SStandWalk: 6, SSit: 1.5, SFallen: 0.4})
	case observation.PoseFallen:
		m := map[State]float64{SFallen: 8, SStandWalk: 0.3, SSit: 0.3}
		if g == GeomInBed {
			m[SFallen] = 1.5 // 床上 pose=fallen 多是躺姿误读，降权
			m[SBedLying] = 4
		} else if g == GeomOpenFloor {
			m[SFallen] = 10 // 开阔地板确认倒地，升权
		}
		return lk(m)
	case observation.PoseLying:
		switch g {
		case GeomInBed:
			return lk(map[State]float64{SBedLying: 6, SBedRestless: 3, SFallen: 0.3})
		case GeomOpenFloor:
			return lk(map[State]float64{SFallen: 4, SBedLying: 0.5, SStandWalk: 0.4})
		default:
			return lk(map[State]float64{SBedLying: 2, SFallen: 1.5})
		}
	case observation.PoseSuspectedSitGround, observation.PoseSitGround:
		return lk(map[State]float64{SFallen: 3, SSit: 1.5})
	case observation.PoseBedSitUp, observation.PoseSuspectedBedSitUp, observation.PoseConfirmedBedSitUp:
		return lk(map[State]float64{SBedRestless: 4, SBedLying: 1.5})
	}
	return lk(nil)
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
