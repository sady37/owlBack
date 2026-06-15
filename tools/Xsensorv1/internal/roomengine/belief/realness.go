package belief

import "math"

// realness.go — realness 隐轴（DBN-Zone-Room §G 重定，2026-06-14）：两类 {Real, Mirror}，Static 溶解。
//
// 框架（§G）：留到本层的持续 track 要么真人、要么有真人源的动态镜像——孤立纯金属由 firmware 在底下
//   30s 过滤（无目标 ID=88），不传到本层；故本层不分「静止金属反射体」（Static 删）。
//
// 主职 = N_r 排除 ghost（非压 ghost 自己的摔）：ghost 被当第 2 人 → N_r 虚增 → decide 折扣 C_FN →
//   独处真人风险被误降 → 漏报。N_r 计数排除留 W3.4 多 track。
//
// 两轴 FN-safe 非对称默认（「一切看风险」）：fall 轴 pFallReal 默认中性 ≡1，**仅** co-existence
//   （mirror 共现）正证据才 <1。「无 ghost 证据」≠「是 ghost」——绝不 leak 向均匀凭空造 ghost 质量
//   （§G 删 leak 治本，治 §50 误抑制；连带删旧 survival/Static 凑 Real 的反 leak 补偿）。
//
// 闸门：自主走动（Displaced 且非同步反射）→ movedFromBirth 锁 Real，此后共现也不重判 Mirror
//   （已确认真人摔倒静止/消失不被当 ghost 否定，cd2b 病另一面防护）。

// realness 形态锚（铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle）。
const (
	rcMirrorGain    = 8.0 // 共现镜面单帧累积强度（mScore += rcMirrorGain·CoexistRho）
	rcAutonomousRho = 0.5 // 共动强度 > 此 + IsReflection = 同步反射（非自主），不算「自主走动」
)

// RealnessObs 一帧 realness 证据（adapter 由出生档案 + MM 拓扑译入；此处只认 schema）。
type RealnessObs struct {
	Displaced    bool    // 本帧相对 BirthPos 有位移（走动/跌倒位移）→ 自主走动锁 Real
	CoexistRho   float64 // 与配对 track 共动强度 [0,1]（mirror；MM 邻居在 W3.4b 填，单房=0）
	IsReflection bool    // 镜面几何指向本 track 是反射（破对称；MM 在 W3.4b 填，单房=false）
}

// RealnessTrack 单 track realness 后验（Real vs Mirror）+ 走动闩。
//
//	mScore = 共现镜面证据累积（≥0，起 0）；P(Mirror)=1-e^{-mScore}，pFallReal=P(Real)=e^{-mScore}。
//	起 mScore=0 → P(Real)=1：无 co-existence 证据 → 中性不压（§G「无 ghost 证据不造 ghost 质量」）。
type RealnessTrack struct {
	mScore         float64
	movedFromBirth bool // 单调闩：曾自主走开出生点 = 确认真人，此后不重判 Mirror
}

// NewRealnessTrack 起中性真人（P(Real)=1）。无出生先验分叉——默认即 Real（§G：无 ghost 证据不压）。
func NewRealnessTrack() *RealnessTrack { return &RealnessTrack{} }

// Update 一帧证据更新。
func (r *RealnessTrack) Update(o RealnessObs) {
	// 自主走动 = Displaced 且非同步反射（镜像也 Displaced，但那是 synced 非自主）→ 确认真人。
	if o.Displaced && !(o.IsReflection && o.CoexistRho > rcAutonomousRho) {
		r.movedFromBirth = true
	}
	if r.movedFromBirth {
		r.mScore = 0 // 已确认真人：锁 Real，co-existence 不能再判 Mirror
		return
	}
	if o.CoexistRho > 0 && o.IsReflection {
		r.mScore += rcMirrorGain * o.CoexistRho // 共现镜面 → Mirror 累积（唯一拉低 P(real) 的正证据）
	}
}

// PReal 真人后验 = pFallReal（喂 SFallen 发射调制；无 co-existence → 1 不压）。
func (r *RealnessTrack) PReal() float64 { return math.Exp(-r.mScore) }

// PMirror 动态镜像后验。
func (r *RealnessTrack) PMirror() float64 { return 1 - math.Exp(-r.mScore) }

// PGhost 非真人后验（Static 溶解 → ghost 仅 mirror）。
func (r *RealnessTrack) PGhost() float64 { return r.PMirror() }

// PRoomHasReal 共生律（§G）：任一 track（Real 或 Mirror）皆蕴含一个真人源（金属已被 firmware 过滤）
//
//	→ 有任一 track 即房内有真人。用途：真人 track 丢失但镜像存活 → 仍判房内有真人 → fall 不被「无 track」抑制。
func PRoomHasReal(tracks []*RealnessTrack) float64 {
	if len(tracks) == 0 {
		return 0
	}
	return 1
}
