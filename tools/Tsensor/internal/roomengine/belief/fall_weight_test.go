package belief

import "testing"

// TestFallEvidenceWeightedByReal — P1③(委员会 d48e0da 规格):Room 层 fall 发射 ×P(Real)。
// fall 证据(ObsDwellStill)Conf 被 faller 的 P(T=Real) 加权 → ghost(低 pReal→低 Conf)喂不动 SFallen,
// real(高 pReal→高 Conf)正常 ramp。真伪前置=耦合加权(非顺序 gate):ghost 不会被 dwell 喂成假 still-fall。
func TestFallEvidenceWeightedByReal(t *testing.T) {
	run := func(conf float64) float64 {
		b := New(DefaultModel())
		ts := int64(1000)
		for i := 0; i < 60; i++ {
			b.Step(ts, []Observation{{
				Kind: ObsDwellStill, Value: float64(60 + i*5), Conf: conf, AreaType: areaActive, Ts: ts, Fresh: true,
			}})
			ts += 1000
		}
		return b.Vector().P(SFallen)
	}
	real := run(0.9)  // real faller:pReal≈0.9 → Conf 高
	ghost := run(0.1) // ghost:pReal≈0.1 → Conf 被折扣
	t.Logf("real(Conf=0.9) P(SFallen)=%.4f / ghost(Conf=0.1) P(SFallen)=%.4f", real, ghost)
	if ghost >= real {
		t.Errorf("★ghost 折扣后 Conf=0.1 P(SFallen)=%.4f 应 < real Conf=0.9 %.4f——P(Real) 加权让 ghost 喂不动 SFallen", ghost, real)
	}
}
