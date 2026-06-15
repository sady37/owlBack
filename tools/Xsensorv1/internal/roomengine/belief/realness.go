package belief

// realness.go — realness 隐轴（§10 换轴 ghost；§34 track=2；§35 三缺口补全）。
//
// 统一原理（用户领域知识，2026-06-16）：**真人 = 入口出生 + 会动；非人违反其一。**
//   - 动态镜像 ghost：与真人同步共动（mirror 几何）。
//   - 静止金属反射体：出生即在此 + 久未移走 + 近墙（static_reflector 三签名）。
// 三类 realness T^(i) ∈ {Real, Mirror, Static}，废 Track 层 Conf×P(Real) 硬外挂，realness 内化为隐轴。
//
// 核心物理（用户纠正，关乎漏报）：雷达功率随人体活动自适应——静止的东西（ghost 弱反射 / 金属 /
// **摔倒静止的人**）都因降功率被滤掉、track 消失。「track 消失」双关：ghost/金属消失=无事；
// 真人摔倒消失=最危险。区分靠**消失前积累的 real 信心**。故 realness 与 fall 必须联合、不能 gate。
//
// 两层证据累积 real 信心（连续 P(real)，非硬阈——A 其六，硬阈=请回 gate + 标定陷阱铁律）：
//   快（出生几帧）：出生地(BirthScore 入口近→Real) + 运动(走动→Real / 共动→Mirror) + 区域(AreaDeny→Static)。
//   慢（兜底）：survival——跨过静止降功率期仍在 → Real（ghost/金属活不过静止期，真人活得过）。
//
// 闸门（连续，结构性自带）：Static 签名要求「困在 BirthPos 久未移走」；真人一旦走开（movedFromBirth）
//   就**永不能**被后续静止重判 Static/Mirror → 已确认 real 的静止消失 = 摔倒，不被当 ghost 否定
//   （cd2b 病另一面防护）。P(real) 连续喂 decide：fall×P(real)（ghost 的「摔」喂不动 SFallen）。
//
// 共生律（决定14）：动态镜像必有真人作反射源（金属反射体不需）。故任一 track 的 P(Mirror)>0 →
//   抬「房内有真人」后验 PRoomHasReal——丢真人 track 但有镜像 = 真人仍在 → 维持 fall 后验（不抑制）。

// RealClass realness 三类。
type RealClass int

const (
	RCReal   RealClass = iota // 真人（入口出生 + 会动 / 跨静止期存活）
	RCMirror                  // 动态镜像 ghost（与真人共动，必有真人源）
	RCStatic                  // 静止金属反射体（困 BirthPos + 久未移走 + 近墙，无真人源）
	numRealClass
)

// realness 发射形态锚（铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle）。
const (
	rcMoveGain      = 4.0 // 走动 → Real
	rcSurvivalGain  = 3.0 // 跨静止期存活 → Real（慢证据）
	rcMirrorGain    = 8.0 // 共动 + 镜面指向本 track → Mirror
	rcStaticGain    = 8.0 // 困 BirthPos + 久 + 近墙（且从未走开）→ Static
	rcConfirmedReal = 10.0 // 已确认 real（自主走动闩）后每帧强锁 Real
	rcLeak          = 0.2 // 未确认时向均匀遗忘（防 lock-in，可恢复）；已确认 real 不 leak（信心不被静止侵蚀）
	rcAutonomousRho = 0.5 // 共动强度 > 此 + IsReflection = 同步反射（非自主），不算"自主走动"
)

// RealnessObs 一帧 realness 证据（adapter 由 track 出生档案 + cell 标注译入；此处只认 schema）。
type RealnessObs struct {
	Displaced          bool    // 本帧相对 BirthPos 有位移（走动）→ Real + 置 movedFromBirth 闩
	ConfinedNearWall   bool    // 困 BirthPos 附近 + 近墙（static 签名①③）
	AgeLongStatic      bool    // 存活够久仍困住（static 签名②，≥5min）
	CoexistRho         float64 // 与配对 track 共动强度 [0,1]（mirror）
	IsReflection       bool    // 镜面几何指向本 track 是反射（破对称）
	CrossedStillPeriod bool    // 跨过静止降功率期仍在（real 慢证据：ghost/金属活不过）
}

// RealnessTrack 单 track realness 后验 + 走动闩。
type RealnessTrack struct {
	p              [numRealClass]float64 // {Real, Mirror, Static}，Σ=1
	movedFromBirth bool                  // 单调闩：曾走开出生点 = 真人 + 闸门（此后静止不重判 Static）
}

// NewRealnessTrack 出生先验（出生地快证据，one-shot）：
//   bornNearEntrance → 偏 Real；inAreaDeny（金属区出生）→ 偏 Static；都不 → 弱不确定（待后续证据）。
func NewRealnessTrack(bornNearEntrance, inAreaDeny bool) *RealnessTrack {
	r := &RealnessTrack{}
	switch {
	case bornNearEntrance:
		r.p = [numRealClass]float64{0.8, 0.1, 0.1} // 入口出生 → 偏真人
	case inAreaDeny:
		r.p = [numRealClass]float64{0.2, 0.1, 0.7} // 金属区凭空出生 → 偏静止反射
	default:
		r.p = [numRealClass]float64{0.5, 0.25, 0.25} // 远离入口非金属 → 不确定，待运动/survival 定
	}
	return r
}

// Update 一帧证据更新（连续后验，非硬阈）。
func (r *RealnessTrack) Update(o RealnessObs) {
	// 自主走动 = Displaced 且非同步反射（镜像共动也 Displaced，但那是 synced 非自主）→ 真人闩。
	autonomous := o.Displaced && !(o.IsReflection && o.CoexistRho > rcAutonomousRho)
	if autonomous {
		r.movedFromBirth = true // 单调闩：一旦自主走动 = 确认真人
	}

	psi := [numRealClass]float64{1, 1, 1}
	leak := rcLeak
	if r.movedFromBirth {
		// 已确认 real（闸门）：强锁 Real + 不 leak。后续静止/消失 = 摔倒（走 fall 轴），**绝不**被
		// 静止重判 Mirror/Static（cd2b 病另一面防护：真人摔倒静止消失不当 ghost 漏报）。
		psi[RCReal] = rcConfirmedReal
		leak = 0
	} else {
		if o.Displaced {
			psi[RCReal] *= rcMoveGain // 走动 → Real
		}
		if o.CrossedStillPeriod {
			psi[RCReal] *= rcSurvivalGain // 慢证据：跨静止期存活 → Real
		}
		if o.CoexistRho > 0 && o.IsReflection {
			psi[RCMirror] *= 1 + rcMirrorGain*o.CoexistRho // 共动镜面 → Mirror
		}
		if o.ConfinedNearWall && o.AgeLongStatic { // 未确认才可判 Static（!movedFromBirth 在此分支已成立）
			psi[RCStatic] *= rcStaticGain
		}
	}

	sum := 0.0
	for i := range r.p {
		r.p[i] = (1-leak)*r.p[i] + leak/float64(numRealClass)
		r.p[i] *= psi[i]
		sum += r.p[i]
	}
	if sum <= 0 {
		r.p = [numRealClass]float64{1, 0, 0}
		return
	}
	for i := range r.p {
		r.p[i] /= sum
	}
}

// PReal 真人后验（连续，喂 decide：fall×P(real)；消失时 P(real) 高→Blind→Fallen ramp）。
func (r *RealnessTrack) PReal() float64 { return r.p[RCReal] }

// PMirror / PStatic 两类 ghost 后验。
func (r *RealnessTrack) PMirror() float64 { return r.p[RCMirror] }
func (r *RealnessTrack) PStatic() float64 { return r.p[RCStatic] }

// PGhost 非真人后验（镜像 + 静止反射）。
func (r *RealnessTrack) PGhost() float64 { return r.p[RCMirror] + r.p[RCStatic] }

// PRoomHasReal 共生律（决定14）：动态镜像必有真人源（金属不需）。
//   房内有真人后验 = 1 - Π_i(1 - [P(real_i)+P(mirror_i)])。
// 用途：真人 track 丢失但有镜像存活 → PRoomHasReal 仍高 → fall 后验不被「无 track」抑制。
func PRoomHasReal(tracks []*RealnessTrack) float64 {
	noReal := 1.0
	for _, t := range tracks {
		impliesReal := t.p[RCReal] + t.p[RCMirror] // static(金属)不蕴含真人
		if impliesReal > 1 {
			impliesReal = 1
		}
		noReal *= 1 - impliesReal
	}
	return 1 - noReal
}
