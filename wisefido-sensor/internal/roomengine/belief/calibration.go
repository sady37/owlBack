package belief

// calibration.go — P2.6:发射层 L(o|s) 标定集中化(R7)。
// likelihood.go 的所有 LR 数值集中于此,每条带**来源**,可单点调灵敏度、可审。
// 对照 signal_map §2 发射表 + belief_dbn_impl_plan §1 P2 / §8 P9.6 占位标定清单。
//
// 命名:lr* = 似然乘子(>1 抬该态 / <1 压该态 / =1 中性);gain* = 随证据强度 p∈[0,1] 线性的斜率;
//       damp* = 1−k·p 形压制的斜率 k;z*Cm = 高度阈值 cm;stand* = 久站时长项。
// **shadow 期占位值**(待 P9 oracle 定终值)见 §8 P9.6;勿在 shadow 期拍脑袋抬权。
const (
	// ── ObsZBand(P2.3 z 三档 posture;A类round-z3 / 原则#2;只喂 posture,绝不写 SFallen / R5)──
	zUprightCm = 80.0 // >80 直立(高值可信)
	zSitMinCm  = 30.0 // 30–80 坐;<30 假低→无信息 lk(nil)
	lrZStand   = 2.0  // z>80 → +SStandWalk(P9.6:应 ≥ lrPoseStand,现倒挂待 oracle)
	lrZSit     = 2.0  // 30–80 → +SSit

	// ── ObsVitalPresent(有生命体征→必真人,压 Empty/Artifact)──
	lrVitalEmpty    = 0.2
	lrVitalArtifact = 0.3
	lrNoVitalEmpty  = 1.2 // 无 vital 是弱信号(mmWave 常漏)

	// ── ObsBedOccupied(嵌套 bed 贝叶斯 P(InBed)=p;对齐 bed_bayesian_scorer LR 表)──
	gainBedLying     = 5.0 // 1+5p
	gainBedRestless  = 3.0 // 1+3p
	dampBedFallen    = 0.7 // 1−0.7p
	dampBedStandWalk = 0.6 // 1−0.6p
	dampBedSit       = 0.5 // 1−0.5p
	dampBedEmpty     = 0.5 // 1−0.5p

	// ── ObsSleepStage(粗调,conf 偏低)──
	lrSleepRestless = 2.0 // stage 0/8 过渡/醒
	lrSleepLying    = 2.0 // stage 1/2/3 睡

	// ── ObsFirmwareFall(P2.4 / §10#2:pose 派生不可信,降权;终值待 P9 真机 TP/FP)──
	lrFirmwareFallen = 2.0 // 原 ×10,委员会#5 ≤×2

	// ── ObsEnterExit(event present 正向;absence≠负向 / 原则#3)──
	lrEnterEmpty     = 0.2 // EnterRoom 压 Empty
	lrEnterStandWalk = 2.0
	lrEnterLeft      = 0.3
	lrExitLeft       = 8.0 // ExitRoom → Left(强;committee review③ KEEP)
	lrExitEmpty      = 2.0
	lrExitFallen     = 0.2 // 事件正向退场压 Fallen(非 pose,R5 不管事件)
	lrExitStandWalk  = 0.5

	// ── ObsNumberPeople(np=0 是 corroboration 非 substitution;不反驳已倒地)──
	lrNp0Empty     = 1.5
	lrNp0Left      = 2.0
	lrNp0Sit       = 0.5
	lrNp0StandWalk = 0.5
	lrNp0Fallen    = 1.0 // 中性:真倒地证据须仍能竞争
	lrNpOccEmpty   = 0.3 // np>0 压 Empty

	// ── ObsStandDuration(bathroom 久站→still-fall 弱嫌疑;v2 由 HSMM 替代)──
	standFallBase = 1.0 // 1 + gainStandFall·min(d,standCapMin)
	gainStandFall = 0.4
	standCapMin   = 8.0

	// ── ObsTrackPresent(ghost-ness g∈[0,1]→Artifact)──
	gainGhostArtifact  = 6.0 // 1+6g
	dampGhostStandWalk = 0.5 // 1−0.5g
	dampGhostFallen    = 0.5

	// ── ObsNeighbor(§5.5.2 弱耦合:邻居占用高→本房空/离,不太可能倒地)──
	gainNbrEmpty     = 3.0
	gainNbrLeft      = 2.0
	dampNbrFallen    = 0.7
	dampNbrBedLying  = 0.4
	dampNbrStandWalk = 0.4

	// ── ObsNoDetect(看了没测到的状态条件似然;固定强度不含时长,时长归 Decider/P3)──
	lrNoDetStandWalk   = 0.3 // 可检测态凭空消失=坏解释→压
	lrNoDetSit         = 0.4
	lrNoDetBedRestless = 0.6
	lrNoDetBedLying    = 0.8
	lrNoDetFallen      = 1.6 // 略升:贴地遮挡是走动者消失的典型物理因

	// ── ObsReachableExit(e=f_dist·f_reach;近门可达→Left 压 Fallen;替 30cm 硬门悬崖)──
	gainReachLeft   = 6.0 // 1+6e
	gainReachEmpty  = 2.0 // 1+2e
	dampReachFallen = 0.9 // 1−0.9e

	// ── poseLikelihood(P2.2:对 fall 只正向;SStandWalk/SSit/SBedLying 是 posture 区分非 fall 压制)──
	lrPoseWalkStandWalk  = 6.0
	dampPoseWalkBed      = 0.3
	lrPoseWalkEnterLeft  = 3.0 // 门口走动→可能离场
	lrPoseSuspFallen     = 3.0
	lrPoseSuspTrans      = 2.0
	dampPoseSuspStand    = 0.8
	lrPoseSitSit         = 6.0
	dampPoseSitStand     = 0.8
	lrPoseStand          = 6.0 // pose=standing → StandWalk(P9.6:z>80 权应 ≥ 此)
	lrPoseStandSit       = 1.5
	lrPoseFallenBase     = 8.0  // geom 未知
	lrPoseFallenInBed    = 1.5  // 床上 fallen 多躺姿误读,降权(仍 >1 正向)
	lrPoseFallenOpen     = 10.0 // 开阔地板确认倒地,升权
	dampPoseFallenStand  = 0.3
	dampPoseFallenSit    = 0.3
	lrPoseFallenBedLy    = 4.0
	lrPoseLyingBedLying  = 6.0 // lying@InBed
	lrPoseLyingBedRest   = 3.0
	lrPoseLyingOpenFall  = 4.0 // lying@OpenFloor 倒地候选
	dampPoseLyingOpenBed = 0.5
	dampPoseLyingOpenSW  = 0.4
	lrPoseLyingDefBed    = 2.0 // lying@unknown
	lrPoseLyingDefFall   = 1.5
	lrPoseSitGroundFall  = 3.0
	lrPoseSitGroundSit   = 1.5
	lrPoseBedSitUpRest   = 4.0
	lrPoseBedSitUpLying  = 1.5
)
