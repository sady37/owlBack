package belief

import (
	"math"
	"testing"
)

// TestTauStarEquivalence — P7.1 等价复现:confirm 工作点 τ*=0.55 = 历史 θ_fire(单源取代魔数),
// thFire 须由 TauConfirm.Tau() 派生(改 τ* 即改决策,无第二份魔数)。
func TestTauStarEquivalence(t *testing.T) {
	if d := TauConfirm.Tau() - 0.55; math.Abs(d) > 1e-9 {
		t.Fatalf("confirm τ* 期望 0.55(等价历史 θ_fire),得 %.4f", TauConfirm.Tau())
	}
	if d := TauSuspect.Tau() - 0.30; math.Abs(d) > 1e-9 {
		t.Fatalf("suspect τ* 期望 0.30,得 %.4f", TauSuspect.Tau())
	}
	if d := thFire - TauConfirm.Tau(); math.Abs(d) > 1e-9 {
		t.Fatalf("thFire 须派生自 TauConfirm.Tau()(单源),得 thFire=%.4f vs τ*=%.4f", thFire, TauConfirm.Tau())
	}
}

// TestTauStarFormula — τ* = C_FP/(C_FP+C_FN):C_FN≫C_FP → τ* 小(更敏感);C_FP 大 → τ* 大(更保守)。
func TestTauStarFormula(t *testing.T) {
	hiFN := CostRatio{CFP: 1, CFN: 9}  // 漏报代价高 → 低阈早报
	hiFP := CostRatio{CFP: 9, CFN: 1}  // 误报代价高 → 高阈保守
	if hiFN.Tau() >= 0.5 || hiFP.Tau() <= 0.5 {
		t.Fatalf("τ* 方向错:C_FN≫C_FP 应 <0.5(得 %.3f)、C_FP≫C_FN 应 >0.5(得 %.3f)", hiFN.Tau(), hiFP.Tau())
	}
	if TauSuspect.Tau() >= TauConfirm.Tau() {
		t.Fatalf("suspect τ*(%.2f) 应 < confirm τ*(%.2f)——预警比确认更敏感", TauSuspect.Tau(), TauConfirm.Tau())
	}
}

// TestDecideTauThreeBands — 三档读出边界:≤suspect→none / (suspect,confirm]→suspect / >confirm→confirm。
func TestDecideTauThreeBands(t *testing.T) {
	cases := []struct {
		p    float64
		want TauDecision
	}{
		{0.10, TauNone},         // 远低
		{0.30, TauNone},         // == suspect 边界(严格 >)
		{0.31, TauSuspectLevel}, // 刚过 suspect
		{0.50, TauSuspectLevel}, // suspect 区
		{0.55, TauSuspectLevel}, // == confirm 边界(严格 >)
		{0.56, TauConfirmLevel}, // 刚过 confirm
		{0.90, TauConfirmLevel}, // 高
	}
	for _, c := range cases {
		got, _ := DecideTau(c.p)
		if got != c.want {
			t.Fatalf("DecideTau(%.2f)=%v,期望 %v", c.p, got, c.want)
		}
	}
}

// TestDeciderUnchangedByTauRefactor — P7.1 不改 Decider 行为:thFire 仍 0.55,确认窗 confirmMs 仍生效。
func TestDeciderUnchangedByTauRefactor(t *testing.T) {
	var d Decider
	hi := Vector{}
	hi[SFallen] = 0.6              // > thFire
	t0 := int64(1_700_000_000_000) // 真 epoch ms(避开 fireSince==0 的 t=0 歧义)
	// 维持未达 confirmMs → 不确认。
	if got := d.Update(hi, t0); got != DecisionNone {
		t.Fatalf("起报应 None(未达确认窗),得 %v", got)
	}
	// 维持满 confirmMs → 确认 Fall。
	if got := d.Update(hi, t0+confirmMs); got != DecisionFall {
		t.Fatalf("维持满 confirmMs 应 DecisionFall,得 %v", got)
	}
}
