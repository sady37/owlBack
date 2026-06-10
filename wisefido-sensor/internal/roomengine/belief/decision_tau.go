package belief

// decision_tau.go — P7.1 代价比阈 τ*（散落判决阈收敛到单旋钮）。
// τ* = C_FP/(C_FP+C_FN)：Bayes 最优代价敏感阈（§2.4 belief≠报警，报警=belief×非对称代价）。
// C_FN≫C_FP → τ* 小 = 更敏感（漏报代价高，早报）；C_FP 大 → τ* 大 = 更保守（保精度）。
// shadow-first(R0)：本层只供 belief_shadow 读出对账（p7_1_*），不接 alarm。
// 工作点终值待 P9 oracle 真机标定（同 P2.4 deferral）；当前 confirm 取与历史 θ_fire=0.55 等价。

// CostRatio 一个工作点的误报/漏报代价比。
type CostRatio struct{ CFP, CFN float64 }

// Tau 代价比阈 τ* = C_FP/(C_FP+C_FN)。
func (c CostRatio) Tau() float64 { return c.CFP / (c.CFP + c.CFN) }

var (
	// TauConfirm 确认级（保精度高 bar）。τ*=0.55 = 历史 θ_fire 等价复现（单源取代魔数 0.55，见 belief.go thFire）。
	TauConfirm = CostRatio{CFP: 55, CFN: 45}
	// TauSuspect 预警级（漏报代价高 → 低 bar 早报，recall-favoring）。τ*=0.30。
	TauSuspect = CostRatio{CFP: 30, CFN: 70}
)

// TauDecision τ* 读出级（P7.1 30/70 → suspect/confirm 两工作点）。
type TauDecision int

const (
	TauNone         TauDecision = iota // P(Fallen) ≤ suspect
	TauSuspectLevel                    // suspect < P(Fallen) ≤ confirm（预警）
	TauConfirmLevel                    // P(Fallen) > confirm（确认）
)

func (d TauDecision) String() string {
	switch d {
	case TauSuspectLevel:
		return "suspect"
	case TauConfirmLevel:
		return "confirm"
	default:
		return "none"
	}
}

// DecideTau 读出当前 P(Fallen) 落哪个工作点 + 命中的阈值（confirm 优先，base context）。
func DecideTau(pFallen float64) (TauDecision, float64) {
	return DecideTauCtx(pFallen, TauContext{})
}

// ── P7.2 context-dependent τ*（散落 RiskLevel 多档阈 → 单一代价比旋钮随 context） ──
//
// 漏报代价 C_FN 随风险升 → τ* 降 → 更敏感。代价假设（委员会可审，终值待 P9 标定）：
//   - 浴室：滑倒高发 + 关门独处难被发现 → 漏一次代价高。
//   - 夜间：无人在旁 + 可能长时间无人发现 → 漏一次代价高。
//
// 两者叠乘（既浴室又夜间 = 最高风险 → τ* 最低）。base（日间非浴室）= P7.1 工作点（0.55/0.30）等价复现。
const (
	bathroomCFNMult = 1.5
	nightCFNMult    = 1.5
	// ★P1④(风险分层从代价涌现,委员会 d48e0da):「一切看风险」= 漏报代价 C_FN ∝ 无人救。独处=最高风险
	// (摔了没人救=要命)= 基线 C_FN(fire-leaning);**有人在场**=可救=降险 → C_FN×此(<1)→ τ* 升 → 容抑制。
	// 委员会细化2「降险但不归零」:>0(非 0,不无条件抑制)。独处必发/有人激进否**从代价涌现,非硬规则**。
	othersPresentCFNMult = 0.6
)

// TauContext 决策上下文：zone（是否浴室）+ risk-time（是否夜间）+ human-presence（是否有他人在场=可救）。
type TauContext struct {
	Bathroom      bool
	Night         bool
	OthersPresent bool // ★P1④:房内有他人(≥2 track)=可救=降险 → C_FN↓ → τ*↑ → 容抑制;独处=基线 fire-leaning
}

// cFNMult 漏报代价 C_FN 的风险加权（base=1）。
func (c TauContext) cFNMult() float64 {
	m := 1.0
	if c.Bathroom {
		m *= bathroomCFNMult
	}
	if c.Night {
		m *= nightCFNMult
	}
	if c.OthersPresent {
		m *= othersPresentCFNMult // 有人在场降险:C_FN↓→τ*↑→容抑制(独处不乘=基线最高 fire-leaning)
	}
	return m
}

// withCtx 把 base 工作点的 C_FN 按 context 风险加权（C_FP 不动 → τ* 降）。
func (c CostRatio) withCtx(ctx TauContext) CostRatio {
	return CostRatio{CFP: c.CFP, CFN: c.CFN * ctx.cFNMult()}
}

// TauConfirmFor / TauSuspectFor context-adjusted 工作点。
func TauConfirmFor(ctx TauContext) CostRatio { return TauConfirm.withCtx(ctx) }
func TauSuspectFor(ctx TauContext) CostRatio { return TauSuspect.withCtx(ctx) }

// DecideTauCtx context-dependent 三档读出（confirm 优先）。
func DecideTauCtx(pFallen float64, ctx TauContext) (TauDecision, float64) {
	confirm := TauConfirmFor(ctx).Tau()
	suspect := TauSuspectFor(ctx).Tau()
	if pFallen > confirm {
		return TauConfirmLevel, confirm
	}
	if pFallen > suspect {
		return TauSuspectLevel, suspect
	}
	return TauNone, suspect
}
