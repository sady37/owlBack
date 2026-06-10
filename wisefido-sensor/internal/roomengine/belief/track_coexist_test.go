package belief

import "testing"

// TestCoExistGhostCoupling — P1①+②(委员会 b356843 钉「①ρ-prior+②对称发射必须同建」)：
// ① ρ-prior=「ghost 可能性」(共存才可能,孤立=0);② 对称发射=「谁是 ghost」判别器。
// 单测必覆盖三案(委员会:不只测孤立,必测 2 真人防 ① over-flag)：
//
//	a) 孤立(ρ=0)→ P(Ghost)≈0(long-lie 真受害者结构安全)
//	b) 2 真人(ρ 高但无对称→Ghostness 低)→ 都 Real(② 对称判别防 ① 单独把 2 真人都误判 ghost)
//	c) 1 真+1 ghost(ρ 高+对称→Ghostness 高)→ ghost 被识别
func TestCoExistGhostCoupling(t *testing.T) {
	run := func(rho, ghostness float64) (pGhost, pReal float64) {
		tb := NewTrackBelief()
		ts := int64(1000)
		for i := 0; i < 40; i++ {
			tb.StepCoupled(ts, []TObservation{{
				Kind: TObsPresent, Ghostness: ghostness, Geom: GeomOpenFloor, Conf: 0.9, Ts: ts, Fresh: true,
			}}, rho)
			ts += 1000
		}
		v := tb.Vector()
		return v.P(TGhost), v.P(TReal)
	}
	loneG, _ := run(0.0, 0.95)     // a) 孤立:高 Ghostness 也救不回
	bystG, bystR := run(0.9, 0.15) // b) 2 真人:ρ 高(共存)但无对称(Ghostness 低)
	ghostG, _ := run(0.9, 0.9)     // c) 1 真+1 ghost:ρ 高 + 对称
	t.Logf("a)孤立 P(Ghost)=%.4f | b)2真人(无对称) P(Ghost)=%.4f P(Real)=%.4f | c)1真1ghost(对称) P(Ghost)=%.4f", loneG, bystG, bystR, ghostG)
	if loneG > 0.05 {
		t.Errorf("★a 孤立 P(Ghost)=%.4f 应≈0——无共存 Real 不可能 ghost(long-lie 安全)", loneG)
	}
	if bystG > 0.2 || bystR < 0.5 {
		t.Errorf("★b 2真人(ρ高无对称)P(Ghost)=%.4f P(Real)=%.4f 应 Ghost 低/Real 高——② 对称判别防 ① 把 2 真人都误判 ghost(委员会钉)", bystG, bystR)
	}
	if ghostG < 0.5 {
		t.Errorf("c 1真+1ghost(ρ高+对称)P(Ghost)=%.4f 应高——ghost 被识别", ghostG)
	}
}

// TestRiskStratifiedTau — P1④(委员会 d48e0da):一切看风险 = C_FN ∝ 无人救,从代价涌现非硬规则。
// 独处=最高风险=基线 fire-leaning;有他人在场=可救=降险 → C_FN↓ → τ*↑ → 容抑制(降险但不归零 τ*<1)。
func TestRiskStratifiedTau(t *testing.T) {
	alone := TauConfirm.withCtx(TauContext{OthersPresent: false}).Tau()
	present := TauConfirm.withCtx(TauContext{OthersPresent: true}).Tau()
	t.Logf("独处 τ*=%.3f / 有人在场 τ*=%.3f", alone, present)
	if present <= alone {
		t.Errorf("★有人在场 τ*=%.3f 应 > 独处 %.3f——降险→C_FN↓→τ*↑→容抑制(风险分层从代价涌现)", present, alone)
	}
	if present >= 1.0 {
		t.Errorf("有人在场 τ*=%.3f 应 <1——委员会细化2:降险但不归零(非无条件抑制)", present)
	}
}
