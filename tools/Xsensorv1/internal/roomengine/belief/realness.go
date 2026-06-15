package belief

import "math"

// realness.go — realness 隐轴（DBN-Zone-Room §G/§G七）：真人 vs 两类 ghost（mirror 有真人源 / 伪迹无真人源）。
//
// 两类 ghost 独立累积、语义不同（§54 共生律防污染）：
//   - mirror（mScore，co-existence ρ 累积，仅 track==2）：真人的反射，**有真人源**（共生律 → 贡献 PRoomHasReal）。
//   - 伪迹（aScore，运动异常量累积，§G七桶一，单 track 每帧）：固件 track-swap/超速，**无真人源**（不贡献 PRoomHasReal）。
// 两类对下游：pFallReal（压自己的摔）都压；N_r 都排；**PRoomHasReal 只 mirror 贡献真人源**。
//
// 数量×时间、非硬 gate（[[fall_data_is_artificial_test]]）：两 score 皆「每帧异常量随时间累积」，
//   持续越强信心越高、偶发一帧 ≈ 不判（同 §3 κ EMA 范式）。
//
// 闸门：自主走动（plausible 位移、非伪迹速度、非同步反射）→ movedFromBirth 锁 mirror（mScore=0），
//   此后共现不重判 Mirror（已确认真人摔倒静止/消失不当 mirror 否定，cd2b 病另一面）。

// realness 形态锚（铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle）。
const (
	rcMirrorGain    = 8.0 // 共现镜面单帧累积强度（mScore += rcMirrorGain·CoexistRho）
	rcAutonomousRho = 0.5 // 共动强度 > 此 + IsReflection = 同步反射（非自主），不算「自主走动」
)

// RealnessObs 一帧 realness 证据（adapter 由出生位/MM/census 译入；此处只认 schema）。
type RealnessObs struct {
	Displaced       bool    // 本帧相对 BirthPos 有位移（走动/跌倒位移）→ 自主走动锁 Real
	CoexistRho      float64 // 与配对 track 共动强度 [0,1]（mirror；MM 邻居 W3.4b 填，单房=0）
	IsReflection    bool    // 镜面几何指向本 track 是反射（mirror；MM 填，单房=false）
	ArtifactQuantum float64 // 本帧运动伪迹异常量 ≥0（超速幅度；census §G七桶一算，单 track 每帧）
}

// RealnessTrack 单 track realness 后验 + 走动闩。两类 ghost 证据独立累积：
//
//	mScore = mirror（有真人源）；aScore = 伪迹（无真人源，§54 独立分量防共生律污染）。
//	pFallReal = PReal = e^{-(mScore+aScore)}（两类 ghost 都压自己的摔）；无证据 → 1 中性不压。
type RealnessTrack struct {
	mScore         float64
	aScore         float64
	movedFromBirth bool // 单调闩：曾自主走开出生点 = 确认真人，此后不重判 mirror
}

// NewRealnessTrack 起中性真人（PReal=1）。
func NewRealnessTrack() *RealnessTrack { return &RealnessTrack{} }

// Update 一帧证据更新。
func (r *RealnessTrack) Update(o RealnessObs) {
	// 自主走动 = plausible 位移（非伪迹速度）且非同步反射 → 确认真人，锁 mirror。
	if o.Displaced && o.ArtifactQuantum == 0 && !(o.IsReflection && o.CoexistRho > rcAutonomousRho) {
		r.movedFromBirth = true
	}
	if r.movedFromBirth {
		r.mScore = 0 // 已确认真人：锁 mirror（不重判 Mirror）
	} else if o.CoexistRho > 0 && o.IsReflection {
		r.mScore += rcMirrorGain * o.CoexistRho // 共现镜面 → mirror 累积
	}
	r.aScore += o.ArtifactQuantum // 伪迹累积：独立于 mirror/movedFromBirth（运动异常无真人源）
}

// PReal 真人后验 = pFallReal（喂 SFallen 发射调制；无 ghost 证据 → 1 不压）。
func (r *RealnessTrack) PReal() float64 { return math.Exp(-(r.mScore + r.aScore)) }

// PMirror 镜像后验（有真人源 ghost）。
func (r *RealnessTrack) PMirror() float64 { return r.ghostShare(r.mScore) }

// PArtifact 伪迹后验（无真人源 ghost，§G七桶一）。
func (r *RealnessTrack) PArtifact() float64 { return r.ghostShare(r.aScore) }

// ghostShare 非真人质量 (1-PReal) 按两 score 占比分给 mirror/伪迹。
func (r *RealnessTrack) ghostShare(s float64) float64 {
	tot := r.mScore + r.aScore
	if tot == 0 {
		return 0
	}
	return (1 - math.Exp(-tot)) * s / tot
}

// PGhost 非真人后验（mirror + 伪迹）。
func (r *RealnessTrack) PGhost() float64 { return 1 - r.PReal() }

// PRoomHasReal 共生律（§54）：mirror 蕴含真人源、伪迹**不**蕴含 → impliesReal = PReal + PMirror。
//
//	用途：真人 track 丢失但镜像存活 → 仍判房内有真人 → fall 不被「无 track」抑制；纯伪迹 track 不抬此后验。
func PRoomHasReal(tracks []*RealnessTrack) float64 {
	noReal := 1.0
	for _, t := range tracks {
		impliesReal := t.PReal() + t.PMirror() // 伪迹(PArtifact)无真人源，不计入
		if impliesReal > 1 {
			impliesReal = 1
		}
		noReal *= 1 - impliesReal
	}
	return 1 - noReal
}
