package belief

// fall_reason.go — P7.3 reason 路由（读出可解释性）。fall 成因 reason 由"哪个 fall-ward 观测节点
// 后验主导（把 SFallen 似然比拉最高）"决定，取代 fall_unified 散落 reason 函数。shadow-first(R0)：只读出到 log。

// FallReason fall 后验主导成因类（映射 gate-list 散落 reason 的语义类，impl_plan §6 P7.3）。
type FallReason int

const (
	ReasonUnknown   FallReason = iota // 无 fall-ward 主导（所有 fresh obs 对 SFallen LR≤1）
	ReasonLost                        // NoDetect 主导：走动中消失=lost-fall（≈gate-list lost_track）
	ReasonSilent                      // DwellStill 主导：dwell 久静=silent/still fall（≈bathroom_still/bedroom_person_silent）
	ReasonPoseLying                   // Pose 主导：lying/fallen posture 抬 fall（开阔地倒姿）
)

var fallReasonLabel = [...]string{"unknown", "lost", "silent", "pose_lying"}

func (r FallReason) String() string {
	if int(r) < 0 || int(r) >= len(fallReasonLabel) {
		return "unknown"
	}
	return fallReasonLabel[r]
}

// reasonFor 主导 ObsKind → reason 类。
func reasonFor(k ObsKind) FallReason {
	switch k {
	case ObsNoDetect:
		return ReasonLost
	case ObsDwellStill:
		return ReasonSilent
	case ObsPose:
		return ReasonPoseLying
	default:
		return ReasonUnknown
	}
}

// DominantFallObs 返回把 SFallen 拉最高的 fresh obs 及其似然比（仅 LR>1 算 fall-ward；无则 kind=-1,lr=1）。
// LR 用 rawLikelihood[SFallen]（R5 校准约定:1.0=中性 / >1 拉向 Fallen / <1 压制）。
func DominantFallObs(obs []Observation) (ObsKind, float64) {
	dom := ObsKind(-1)
	best := 1.0
	for _, o := range obs {
		if o.effConf() <= 0 {
			continue
		}
		if lr := rawLikelihood(o)[SFallen]; lr > best {
			best, dom = lr, o.Kind
		}
	}
	return dom, best
}

// FallReasonFor obs 序列 → 主导 reason + 主导 obs kind + LR（P7.3 读出）。
func FallReasonFor(obs []Observation) (FallReason, ObsKind, float64) {
	dom, lr := DominantFallObs(obs)
	return reasonFor(dom), dom, lr
}
