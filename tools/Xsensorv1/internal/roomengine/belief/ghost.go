package belief

// ghost.go — realness 隐轴（§10 换轴 ghost；§34 基数锁：track=2 单场景）。
// 废 Track 层 Conf×P(Real) 硬外挂，realness 内化为隐变量，**P(ghost) 纯从 co-existence ρ 涌现**。
//
// 结构（C §34 钉死：pairwise ρ，不做 (S×T)² 全联合）：成对 realness (T^A,T^B)，T∈{Real,Ghost}，
// 4 态 RR/RG/GR/GG。S^(i)（姿态/区域）仍在各自 S-filter；realness 是与 S 正交的 pairwise 轴。
// 与 §3 κ 同源：co-existence 是「两 track 看是不是同一物理实体（一真一镜）」的成对耦合。
//
// 病根规避（[[fall_detection_risk_stratified_design]]）：ghost **绝不**由单 track 静止/realness 判——
// 只由 co-existence ρ（两 track 共动+镜面）涌现。track=1 无对照→不处理；ρ=0（独立）→ 两 track 皆 Real
// （§10「孤立 ρ=0 → P(Ghost)=0」结构安全：久躺单人是真人不是 ghost）。

// realness pair 索引：idx = TA*2 + TB，TA/TB ∈ {0=Real, 1=Ghost}。
const (
	pRR = 0 // 两真（独立两人）
	pRG = 1 // A 真 B 镜
	pGR = 2 // A 镜 B 真
	pGG = 3 // 两镜（无真源，近不可能）
)

const (
	ghostEpsGG = 1e-3 // ψ(G,G)：两 track 都是镜=没有真源，近 0
	ghostLeak  = 0.3  // 每帧向均匀遗忘（防 lock-in + co-existence 消失可恢复 Real）
)

// GhostPair 成对 realness 后验（4 态，Σ=1）。
type GhostPair struct {
	p [4]float64
}

// NewGhostPair 初始两真主导（先验：默认真人，未见 co-existence 不预疑 ghost）。
func NewGhostPair() *GhostPair {
	return &GhostPair{p: [4]float64{0.94, 0.02, 0.02, 0.02}}
}

// Update 一帧 co-existence 证据更新（§10 对称共现）。
//   rho   ∈[0,1]：co-existence 强度（两 track 共动 + 镜面几何相关；mirror_detect/static_reflector 几何输入）。
//   biasA ∈[0,1]：镜面几何指向「A 是反射」的置信（破 RG/GR 对称）。0.5=不知哪个是镜。
// co-exist 高 → 恰一个 ghost（RG/GR），压 RR（非两真）/GG（非两镜）；ρ=0 → 仅 RR 存活（皆真）。
func (g *GhostPair) Update(rho, biasA float64) {
	// 向均匀遗忘（leaky，使 co-existence 消失后能恢复 Real）。
	for i := range g.p {
		g.p[i] = (1-ghostLeak)*g.p[i] + ghostLeak*0.25
	}
	// co-existence 成对势（与 κ 同源，对称；mirror 几何破对称）。
	psi := [4]float64{
		pRR: 1 - rho,           // 共现 ⇒ 非两真
		pRG: rho * (1 - biasA), // B 是镜
		pGR: rho * biasA,       // A 是镜
		pGG: ghostEpsGG,        // 非两镜
	}
	sum := 0.0
	for i := range g.p {
		g.p[i] *= psi[i]
		sum += g.p[i]
	}
	if sum <= 0 { // 证据全灭（理论不可达）→ 退两真先验
		g.p = [4]float64{1, 0, 0, 0}
		return
	}
	for i := range g.p {
		g.p[i] /= sum
	}
}

// PGhostA track A 为 ghost 的后验 = P(GR)+P(GG)。
func (g *GhostPair) PGhostA() float64 { return g.p[pGR] + g.p[pGG] }

// PGhostB track B 为 ghost 的后验 = P(RG)+P(GG)。
func (g *GhostPair) PGhostB() float64 { return g.p[pRG] + g.p[pGG] }

// PExactlyOneGhost 恰一个 ghost 的后验 = P(RG)+P(GR)（co-existence 成立的标志）。
func (g *GhostPair) PExactlyOneGhost() float64 { return g.p[pRG] + g.p[pGR] }

// PReal track 的 realness 真度（喂 decide：fall 发射×P(Real)，ghost 的「摔」喂不动 SFallen，§10/dbn_cutover ③）。
func (g *GhostPair) PRealA() float64 { return 1 - g.PGhostA() }
func (g *GhostPair) PRealB() float64 { return 1 - g.PGhostB() }
