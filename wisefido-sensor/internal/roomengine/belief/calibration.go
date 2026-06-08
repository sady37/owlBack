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
	// P6.1a(阻塞项#1):no-detect 抬 Fallen 不再固定,改门控 1+noDetGainFallen·P(real)·(1−P(door-exit))。
	// 真人非门区消失(P(real)=1,door-exit=0)→ 1+0.6=1.6(=旧上限);ghost 消失/门区可达走出 → →1 中性
	// (不裸 absence 抬 fall,治 dropout-FP)。连续边缘化非硬𝟙(realness 判错平滑退化,§4.3 ghost 融合同构)。
	noDetGainFallen = 0.6
	// P6.1a(审查㉓ door-exit 放行前置):door-exit **不全否决** Fallen,留 floor —— door-distance 是 P6.5 弧
	// 自判"小卫生间不可靠"的弱信号,不该有否决权(门口真摔=朝门逼近+末帧栽倒,窗均速仍高→doorExit≈1,
	// 全抑制则漏报)。k<1 留残余抬升:真离场靠 ExitRoom/np=0 事件压 SLeft 区分(强信号),非门距否决;
	// 门口真摔无 exit 事件 → 残余 Fallen 经 lost 窗累积仍浮出。k 由 oracle 双向判别力定(非预拍)。
	noDetDoorSuppressK = 0.6

	// ── ObsReachableExit(e=f_dist·f_reach;近门可达→Left 压 Fallen;替 30cm 硬门悬崖)──
	gainReachLeft   = 6.0 // 1+6e
	gainReachEmpty  = 2.0 // 1+2e
	dampReachFallen = 0.9 // 1−0.9e

	// ── ObsDwellStill(P4.1 裁决⑮B:dwell 生存函数 ramp;S_vol=exp(−(d/scale)²) → fallLR=1+(d/scale)² 封顶)──
	// **仅 toilet/shower**(Z_cell-无关,fixture 声明可验证)。scale 镜像生产 ToiletShowerSec 作 **shadow 占位**
	// (R0:不碰生产),P9.6 待 oracle。封顶温和(真 dwell-fall 靠 Decider 窗累积)。
	// **开阔地 dwell-fall(P4.4 裁决⑱ B2):scale=dwellScaleOpenSec×ToleranceFactor**——前置 Z_cell tolerance gate
	// (被容忍久站的 cell 尾拉长→久站真人不报;tol 走 cell 几何/历史方差,**Z 只正向 R5**,不用 z 反向压)。
	dwellScaleToiletSec = 900.0 // toilet/shower 15min(对齐生产 ToiletShowerSec)
	dwellScaleOpenSec   = 480.0 // 开阔地 8min(对齐生产 StillTimeoutSec 量级;P4.4,P9.6 待 oracle)
	dwellFallCap        = 2.5   // fallLR 封顶(温和;真 dwell-fall 靠 Decider 窗累积)

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
