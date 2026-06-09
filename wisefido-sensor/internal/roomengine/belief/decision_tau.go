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

// DecideTau 读出当前 P(Fallen) 落哪个工作点 + 命中的阈值（confirm 优先）。
func DecideTau(pFallen float64) (TauDecision, float64) {
	if pFallen > TauConfirm.Tau() {
		return TauConfirmLevel, TauConfirm.Tau()
	}
	if pFallen > TauSuspect.Tau() {
		return TauSuspectLevel, TauSuspect.Tau()
	}
	return TauNone, TauSuspect.Tau()
}
