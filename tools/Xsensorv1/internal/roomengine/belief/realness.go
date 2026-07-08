package belief

// realness.go — 反证法（ghost_disproof_birth_certificate_spec）两轴，判决 ⊥ 显示：
//   判决轴 RealProven(bool，出生证 latch)：N_r 计数 / F1 floor 占用门读它；无证 = ghost，硬 latch。
//   显示轴 PReal(连续 0-1)：纯 FE ×100 confidence，**gate 任何决策皆无**；census 按离 born 净距 R 分档塞入。
// 唯一真化通道 = census 由 roomengine RealProven latch 走 ForceReal；到不了 = 无证 = ForceGhost（默认 ghost，须挣出生证）。

// RealnessTrack 反证法 realness：判决 proven + 显示 pReal 双轴。默认 ghost 起步（proven=false）。
type RealnessTrack struct {
	proven bool    // 判决轴：出生证 latch（N_r/floor 门读）
	pReal  float64 // 显示轴：连续 confidence（纯 FE ×100，gate 无）
}

// realDisplay 证真轨显示 confidence（=80%）：FE 门槛之上→显示（区别于未证 ghost 的门第分档 10~50，见 census.ghostDisplayPReal）。
const realDisplay = 0.80

// NewRealnessTrack 起 realness，默认 ghost（须挣出生证；门第先验由 census ForceReal/ForceGhost 立即覆盖，故起点纯零）。
func NewRealnessTrack() *RealnessTrack { return &RealnessTrack{} }

// ForceReal 有出生证 → real：判决 latch 置真 + 显示 80%（FE 显示阈之上）。
func (r *RealnessTrack) ForceReal() { r.proven, r.pReal = true, realDisplay }

// ForceGhost 反证法：无出生证 → ghost 判决（proven=false）。显示 pReal 由 census 按门第 door_d 分档另设（SetDisplay），不在此动。
func (r *RealnessTrack) ForceGhost() { r.proven = false }

// SetDisplay 设无证轨显示 confidence（census 按出生门第 door_d 分档 10~50；纯 FE，gate 任何决策皆无）。
func (r *RealnessTrack) SetDisplay(pReal float64) { r.pReal = pReal }

// RealProven 判决轴：出生证 latch（N_r 计数 / F1 floor 占用门读它）。
func (r *RealnessTrack) RealProven() bool { return r.proven }

// PReal 显示轴：连续 confidence（→ FE ×100；gate 任何决策皆无）。
func (r *RealnessTrack) PReal() float64 { return r.pReal }

// PMirror 镜像后验（有真人源 ghost；forensic + FE 观测）。
func (r *RealnessTrack) PMirror() float64 { return 1 - r.pReal }
