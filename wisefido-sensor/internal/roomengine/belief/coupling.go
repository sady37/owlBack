package belief

// coupling.go — §3 同床耦合 κ（EMA 无 max 互活门控）+ §4 软归属 a_j + 片内相容势 Ψ。
//   κ：几何冷启 + 事件驱动时间修，可升可降（不用 max，否则几何一旦重叠钉死 1）。
//   a_j：软床归属，κ·g^xy 归一（远离任何床 g^xy→0 → a_j→0 → Ψ 中性）。
//   Ψ：同帧 S×B 两轴相容势（不是转移，耦合只在此施一次，§6）。Fallen 行物理常量不被 κ 覆盖。
//   Ψ 取 mixture（§E，FN-safe）：Ψ=Σ_j a_j ψ̃_j + a_∅，非 product（product 任一床 occ 即把 F 压死=漏报）。

// couplingParams §3/§4 形态参数。开发阶段锚方向/符号，标定见 feedback-p6C（γ 遗忘率 / c_∅ / ε_art §C3）。
type couplingParams struct {
	gammaEvt  float64 // 强化腿:同向共跳 EMA（强,稀疏触发,15s 窗;建立绑同人）
	gammaHold float64 // 维持腿:持续共态 per-frame EMA（弱,抗衰减;熟睡托）
	cEmpty    float64 // a_∅ 的 c_∅（无床归属质量；远离床时 Ψ→中性的载体）
	epsArt    float64 // ψ_phys(F,occ) 被子残留极小（§E/§C3：用在床段 pose=Lying 占比反推，非凭空 1e-3）
	lLeftSup       float64 // sleepad 接触 LeftBed 直接压 SBed（onbed 权重,不门控 g^xy；果断清,残留~0.02；oracle）
	lLeftSupRadar  float64 // radar 几何 LeftBed 压 SBed（<50% 可信,弱清,残留~0.10），<lLeftSup
	lLeftOpen      float64 // ③ LeftBed→SOpenFloor 抬升（near-bed,按 a[j]；radar 接质量主力靠 emission redOpen）
	lLeftOpenRadar float64 // radar LeftBed→SOpenFloor 抬升，<lLeftOpen
}

func defaultCouplingParams() couplingParams {
	return couplingParams{gammaEvt: 0.2, gammaHold: 0.01, cEmpty: 0.2, epsArt: 1e-2,
		lLeftSup: 75, lLeftSupRadar: 13, lLeftOpen: 40, lLeftOpenRadar: 6}
}

// Coupling 持有每床 κ 状态（跨帧 EMA）+ 几何。numBeds 显式（B1）。
type Coupling struct {
	p     couplingParams
	geom  []BedGeom // 长 numBeds
	kappa []float64 // κ_{r,s_j}，长 numBeds；init=几何冷启
}

// NewCoupling 建耦合状态，κ 初值=各床几何冷启 covers·onbed。
func NewCoupling(geom []BedGeom) *Coupling {
	k := make([]float64, len(geom))
	for j := range geom {
		k[j] = geom[j].kappaColdStart()
	}
	return &Coupling{p: defaultCouplingParams(), geom: geom, kappa: k}
}

// Kappa 第 j 床当前耦合置信（诊断/probe 用）。
func (c *Coupling) Kappa(j int) float64 { return c.kappa[j] }

// emaKappa §3 互活门控 EMA（无 max，可升可降）。仅 live[j]=true 才更新；否则该床无信息、κ 不动
// （带遗忘 Beta–Bernoulli，几何作先验伪计数）。两腿共用，差在 γ。
func (c *Coupling) emaKappa(target, live []bool, gamma float64) {
	for j := range c.kappa {
		if j < len(live) && live[j] {
			m := 0.0
			if j < len(target) && target[j] {
				m = 1.0
			}
			c.kappa[j] = (1-gamma)*c.kappa[j] + gamma*m // 无 max：可升可降
		}
	}
}

// UpdateKappa 强化腿：同向共跳 matched（15s 窗解析，强 γ_evt；时间共现绑同人，破空间平局）。
func (c *Coupling) UpdateKappa(matched, live []bool) { c.emaKappa(matched, live, c.p.gammaEvt) }

// MaintainKappa 维持腿：持续共态 agree（per-frame，弱 γ_hold；熟睡持续在床托住 κ 抗衰减）。
func (c *Coupling) MaintainKappa(agree, live []bool) { c.emaKappa(agree, live, c.p.gammaHold) }

// attachment §4 软床归属：a_j=κ_j g^xy_j / (Σ_j' κ_j' g^xy_j' + c_∅)，a_∅=c_∅/(·)。
// gxy[j]=雷达对床 j 的几何 XY 可分性（能分→尖峰 / 只知在某床→床间均匀 / 看不见→0）。长 numBeds。
func (c *Coupling) attachment(gxy []float64) (a []float64, aEmpty float64) {
	nb := len(c.kappa)
	a = make([]float64, nb)
	denom := c.p.cEmpty
	for j := 0; j < nb; j++ {
		g := 1.0
		if j < len(gxy) {
			g = gxy[j]
		}
		a[j] = c.kappa[j] * g
		denom += a[j]
	}
	for j := range a {
		a[j] /= denom
	}
	return a, c.p.cEmpty / denom
}

// psiPhys §4 物理相容表 ψ_phys(S, B^j)（固定）：
//   AtBed: occ=1（人在床→垫占）, vac=1-o_j（床内头/脚端错位）
//   F:     occ=ε_art（摔倒却垫占只能是被子残留）, vac=1（摔倒→离垫→垫空，F 通道全开）
//   else:  1, 1（开阔地跌倒不牵涉床垫，靠 g^xy 门控非显式势惩罚）
// Fallen 行物理常量不被 κ 覆盖（覆盖在 psiTilde 的退火，非这里）。
func (c *Coupling) psiPhys(s State, b BedState, j int) float64 {
	switch s {
	case SBed: // §1 的 AtBed
		if b == BOcc {
			return 1.0
		}
		return 1.0 - c.geom[j].Overlap
	case SFallen:
		if b == BOcc {
			return c.p.epsArt
		}
		return 1.0
	default:
		return 1.0
	}
}

// LogPsi §4 相容势 Ψ → log 域 JointVector（喂 filter.Correct 的 logPsi）。
//   ψ̃_j(S,B^j)=κ_j ψ_phys(S,B^j)+(1-κ_j)   （κ→0 退火中性=诚实无知；κ→1 全物理表）
//   Ψ_t(S,bmask)=Σ_j a_j ψ̃_j(S, bit_j(bmask)) + a_∅   （§E mixture，非 product，FN-safe）
func (c *Coupling) LogPsi(js *JointSpace, gxy []float64) JointVector {
	a, aEmpty := c.attachment(gxy)
	out := js.NewJointVector()
	for i := 0; i < js.size; i++ {
		s, bmask := js.decode(i)
		psi := aEmpty
		for j := 0; j < js.numBeds; j++ {
			b := bedOf(bmask, j)
			psiTilde := c.kappa[j]*c.psiPhys(s, b, j) + (1 - c.kappa[j])
			psi += a[j] * psiTilde
		}
		out[i] = logP(psi)
	}
	return out
}

// LeftBedOpenLogS ③ LeftBed 清空：直接压 SBed（床空→人非 at-bed）+ 抬 SOpenFloor（present）/SBlindRest（lost）。
//
// 压 SBed 用 κ_j（同床耦合,跨帧 track-床绑定）权重，**不用 a_j**：a_j=κ·g^xy 含瞬时位置门控，而 LeftBed 语义=
// 人正离床位置→g^xy→0→a_j→0 会废掉清空（gate 遗留门控，本次绕除）。κ_j 跨帧记住"此 track 绑哪张床"，既不被
// 瞬时离床咬，又防多住户串（邻床 LeftBed 对本床 track κ≈0→不压）。质量去向 FN-safe：present∧walk 时 emission
// supWalk 压 Fallen + redOpen 抬 OpenFloor → 去 OpenFloor 不误火；摔进盲区无 radar → 质量去 Fallen=床边摔浮出。
// 抬 target 仍按 a_j（人在床边离床才抬此处；走远靠 emission redOpen 接质量）。
// sleepad 接触=果断清（lLeftSup,残留~0.02）/ radar 几何=弱清（lLeftSupRadar,<50% 可信,残留~0.10）。
func (c *Coupling) LeftBedOpenLogS(o Observation, gxy []float64, present bool) [numStates]float64 {
	var out [numStates]float64
	a, _ := c.attachment(gxy)
	target := SOpenFloor
	if !present {
		target = SBlindRest
	}
	for j := range c.kappa {
		if j >= len(o.Sleepad) {
			continue
		}
		var lSup, lOpen float64
		switch o.Sleepad[j] {
		case BedLeftBed: // sleepad 接触：果断清
			lSup, lOpen = c.p.lLeftSup, c.p.lLeftOpen
		case BedLeftBedRadar: // radar 几何：弱清（<50% 可信）
			lSup, lOpen = c.p.lLeftSupRadar, c.p.lLeftOpenRadar
		default:
			continue
		}
		out[SBed] -= c.kappa[j] * logP(lSup) // 压 at-bed：κ 权重（跨帧绑定，防多床串，不被瞬时离床 g^xy 咬）
		out[target] += a[j] * logP(lOpen)    // 抬 OpenFloor/BlindRest：人在床边才抬（走远靠 emission redOpen 接）
	}
	return out
}
