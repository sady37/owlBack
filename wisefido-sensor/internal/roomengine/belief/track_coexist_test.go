package belief

import "testing"

// TestCoExistGhostCoupling — P1①(委员会 d48e0da 数学规格)：co-existence 耦合验证。
// 同样高 Ghostness 发射,ρ 不同：孤立 ρ=0 → P(Ghost)→0(进 Ghost 转移被乘 0,即使发射强也救不回)；
// 共存 ρ=0.9(房内有 Real partner)→ P(Ghost) 抬起。证明 long-lie 真受害者(孤立)结构性安全,
// 无需任何 "1 track 不否" 规则——ghost=反射必有共存 Real partner,长在转移矩阵里。
func TestCoExistGhostCoupling(t *testing.T) {
	run := func(rho float64) float64 {
		tb := NewTrackBelief()
		ts := int64(1000)
		for i := 0; i < 40; i++ {
			tb.StepCoupled(ts, []TObservation{{
				Kind: TObsPresent, Ghostness: 0.95, Geom: GeomOpenFloor, Conf: 0.9, Ts: ts, Fresh: true,
			}}, rho)
			ts += 1000
		}
		return tb.Vector().P(TGhost)
	}
	lone := run(0.0)    // 孤立:无共存 Real
	coexist := run(0.9) // 共存:房内 Real partner P(Real)=0.9
	t.Logf("孤立 ρ=0: P(Ghost)=%.4f / 共存 ρ=0.9: P(Ghost)=%.4f", lone, coexist)
	if lone > 0.05 {
		t.Errorf("★孤立 track P(Ghost)=%.4f 应≈0——无共存 Real 不可能 ghost(long-lie 真受害者越不动越像 frozen 但孤立证明是真人)", lone)
	}
	if coexist <= lone {
		t.Errorf("共存 ρ=0.9 P(Ghost)=%.4f 应 > 孤立 %.4f——有 Real partner + 高 Ghostness 才允许 ghost", coexist, lone)
	}
}
