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

// TestTauContextBaseEquivalence — P7.2 base context(日间非浴室)= P7.1 工作点等价复现。
func TestTauContextBaseEquivalence(t *testing.T) {
	base := TauContext{}
	if d := TauConfirmFor(base).Tau() - TauConfirm.Tau(); math.Abs(d) > 1e-9 {
		t.Fatalf("base confirm τ* 应 = P7.1 TauConfirm(%.4f),得 %.4f", TauConfirm.Tau(), TauConfirmFor(base).Tau())
	}
	if d := TauSuspectFor(base).Tau() - TauSuspect.Tau(); math.Abs(d) > 1e-9 {
		t.Fatalf("base suspect τ* 应 = P7.1 TauSuspect(%.4f),得 %.4f", TauSuspect.Tau(), TauSuspectFor(base).Tau())
	}
}

// TestTauContextRiskLowersTau — P7.2 方向:浴室/夜间 C_FN↑ → τ*↓ → 更敏感;叠乘最高风险 τ* 最低。
func TestTauContextRiskLowersTau(t *testing.T) {
	base := TauConfirmFor(TauContext{}).Tau()
	bath := TauConfirmFor(TauContext{Bathroom: true}).Tau()
	night := TauConfirmFor(TauContext{Night: true}).Tau()
	both := TauConfirmFor(TauContext{Bathroom: true, Night: true}).Tau()
	if !(both < bath && both < night && bath < base && night < base) {
		t.Fatalf("τ* 风险单调错:base=%.3f bath=%.3f night=%.3f both=%.3f(期望 both<bath/night<base)", base, bath, night, both)
	}
	// 叠乘:both 的 C_FN 乘子 = bathroom×night = 1.5×1.5=2.25。
	wantBoth := TauConfirm.CFP / (TauConfirm.CFP + TauConfirm.CFN*bathroomCFNMult*nightCFNMult)
	if d := both - wantBoth; math.Abs(d) > 1e-9 {
		t.Fatalf("both τ* 应 = CFP/(CFP+CFN×2.25)=%.4f,得 %.4f", wantBoth, both)
	}
}

// TestDecideTauCtxSensitivity — 同一 P(Fallen) 在高风险 context 下达 confirm,base 下仅 suspect(更敏感)。
func TestDecideTauCtxSensitivity(t *testing.T) {
	p := 0.45 // base: suspect(0.30<0.45≤0.55);浴室+夜:confirm(0.45>0.352)
	if dec, _ := DecideTauCtx(p, TauContext{}); dec != TauSuspectLevel {
		t.Fatalf("base context P=%.2f 期望 suspect,得 %v", p, dec)
	}
	if dec, _ := DecideTauCtx(p, TauContext{Bathroom: true, Night: true}); dec != TauConfirmLevel {
		t.Fatalf("浴室+夜 context P=%.2f 期望 confirm(τ* 降至更敏感),得 %v", p, dec)
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
