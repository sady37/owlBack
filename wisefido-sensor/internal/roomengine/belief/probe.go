package belief

import (
	"fmt"
	"math"
	"strings"
)

// probe.go — §9 逐帧 forensic 快照（cd2b 三态对照 / sdl 复盘）。纯诊断，不参与推断。
// 吐 α 全分量 + 边缘 S/B + P^F + Λ + κ + 裁决，供 Xsensorv1 vs Tsensor 逐帧 diff（baseline bd70194）。

// FrameProbe 一帧完整内部态快照。
type FrameProbe struct {
	Ts        int64     // 观测时刻
	MarginalS Vector    // P(S)（边缘出 S 轴）
	MarginalB []float64 // P(B^j=occ) 每床
	Kappa     []float64 // κ_{r,s_j} 每床（耦合置信）
	PFallen   float64   // P^F_t
	Lambda    float64   // Λ_t（可判性诊断，不参与 fire）
	Decision  Decision  // 裁决快照（含 margin/CFN/Fire）
	Alpha     []float64 // α 全分量（线性域，size=js.size），Σ=1
}

// Snapshot 抓一帧快照。cp 可为 nil（无耦合层时 κ 留空）。
func Snapshot(js *JointSpace, f *Filter, cp *Coupling, dec Decision, lambda float64, ts int64) FrameProbe {
	alpha := f.Alpha()
	lin := make([]float64, js.size)
	for i := range alpha {
		lin[i] = math.Exp(alpha[i])
	}
	mb := make([]float64, js.numBeds)
	for j := 0; j < js.numBeds; j++ {
		mb[j] = js.MarginalB(alpha, j)
	}
	var kappa []float64
	if cp != nil {
		kappa = make([]float64, len(cp.kappa))
		copy(kappa, cp.kappa)
	}
	return FrameProbe{
		Ts:        ts,
		MarginalS: js.MarginalS(alpha),
		MarginalB: mb,
		Kappa:     kappa,
		PFallen:   js.PFallen(alpha),
		Lambda:    lambda,
		Decision:  dec,
		Alpha:     lin,
	}
}

// Line 人类可读单行（cd2b 时间线对照用）。
func (p FrameProbe) Line() string {
	var b strings.Builder
	fmt.Fprintf(&b, "t=%d P^F=%.3f Λ=%.2f%s fire=%v |", p.Ts, p.PFallen, p.Lambda, identTag(p.Decision.Identifiable), p.Decision.Fire)
	top, tp := p.MarginalS.Max()
	fmt.Fprintf(&b, " S*=%v(%.2f)", top, tp)
	for j, pb := range p.MarginalB {
		fmt.Fprintf(&b, " B%d=%.2f", j, pb)
	}
	if len(p.Kappa) > 0 {
		b.WriteString(" κ=")
		for j, k := range p.Kappa {
			if j > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "%.2f", k)
		}
	}
	fmt.Fprintf(&b, " margin=%.3f", p.Decision.Margin)
	return b.String()
}

func identTag(ok bool) string {
	if ok {
		return "✓"
	}
	return "?"
}
