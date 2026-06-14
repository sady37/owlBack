package belief

import (
	"math"
	"testing"
)

// coupling_test.go — §3/§4/§E 不变量验收：
//   C1 κ 无 max（几何高冷启被失配事件压下，max 做不到）
//   C2 κ 互活门控（非 live → 不更新）
//   C3 Fallen 行物理常量（κ=1 → ε_art；κ=0 → 退火中性 1）
//   C4 §E mixture FN-safe（双床 a≈0.5、一床陈旧 occ、人摔 → Ψ(F) 不被压死，product 会）
//   C5 远离床 g^xy→0 → a_∅→1 → Ψ 中性（logPsi≈0）

const ceps = 1e-9

// C1：κ 可降。高几何冷启 + 持续失配 → κ 被 EMA 拉下（max(几何,时间) 钉死 1 做不到）。
func TestKappaNoMax(t *testing.T) {
	c := NewCoupling([]BedGeom{{Covers: 1, Onbed: 1, Overlap: 1}}) // 冷启 κ=1
	if c.Kappa(0) != 1.0 {
		t.Fatalf("冷启 κ=%.4f 应=1（covers·onbed）", c.Kappa(0))
	}
	live := []bool{true}
	miss := []bool{false}
	for i := 0; i < 20; i++ {
		c.UpdateKappa(miss, live)
	}
	if c.Kappa(0) > 0.05 {
		t.Errorf("持续失配 20 步后 κ=%.4f 应趋 0（无 max，可降）", c.Kappa(0))
	}
}

// C2：互活门控。非 live → κ 不动（对方离线=无信息，非"不同床"）。
func TestKappaLivenessGate(t *testing.T) {
	c := NewCoupling([]BedGeom{{Covers: 0.5, Onbed: 1, Overlap: 0.5}})
	k0 := c.Kappa(0)
	for i := 0; i < 10; i++ {
		c.UpdateKappa([]bool{true}, []bool{false}) // matched 但非 live
	}
	if math.Abs(c.Kappa(0)-k0) > ceps {
		t.Errorf("非 live 时 κ 从 %.4f 动到 %.4f，应不更新（互活门控）", k0, c.Kappa(0))
	}
}

// C3：Fallen 行物理常量。ψ̃(F,occ)@κ=1=ε_art；@κ=0=1（退火中性）。
func TestPsiFallenRow(t *testing.T) {
	c := NewCoupling([]BedGeom{{Covers: 1, Onbed: 1, Overlap: 1}})
	psiTilde := func(kappa float64) float64 {
		return kappa*c.psiPhys(SFallen, BOcc, 0) + (1 - kappa)
	}
	if got := psiTilde(1.0); math.Abs(got-c.p.epsArt) > ceps {
		t.Errorf("ψ̃(F,occ)@κ=1=%.4f 应=ε_art=%.4f（Fallen 行物理常量）", got, c.p.epsArt)
	}
	if got := psiTilde(0.0); math.Abs(got-1.0) > ceps {
		t.Errorf("ψ̃(F,occ)@κ=0=%.4f 应=1（κ→0 退火中性）", got)
	}
}

// C4：§E mixture FN-safe。双床各 κ=1、a≈0.5，bed0 陈旧 occ、bed1 vac，人摔（SFallen）。
// mixture Ψ(F)≈a1+a∅ 保 F 竞争；product 会被 bed0 occ 的 ε_art 乘死=漏报。
func TestPsiMixtureFNSafe(t *testing.T) {
	c := NewCoupling([]BedGeom{
		{Covers: 1, Onbed: 1, Overlap: 1},
		{Covers: 1, Onbed: 1, Overlap: 1},
	})
	js := NewJointSpace(2)
	logPsi := c.LogPsi(js, []float64{1.0, 1.0}) // g^xy 等权 → a0≈a1

	bmask := 1 << 0                  // bed0 occ, bed1 vac
	idxF := js.idx(SFallen, bmask)   // 人摔 + bed0 陈旧 occ
	psiF := math.Exp(logPsi[idxF])

	// mixture：Ψ(F)=a0·ε_art + a1·1 + a∅ ≈ 0.45·0.01+0.45·1+0.09 ≈ 0.55 → F 存活
	if psiF < 0.4 {
		t.Errorf("mixture Ψ(F, bed0-occ)=%.4f 应 >0.4（FN-safe，F 不被压死）", psiF)
	}
	// product 对照（手算）：ψ̃0(F,occ)·ψ̃1(F,vac)=ε_art·1=ε_art → 会漏报
	prod := c.p.epsArt * 1.0
	if prod > 0.1 {
		t.Fatalf("product 手算=%.4f（基准应≈ε_art 小）", prod)
	}
	t.Logf("§E ✓ mixture Ψ(F)=%.4f 存活 vs product=%.4f 压死（差 %.0f×）", psiF, prod, psiF/prod)
}

// C5：远离床 g^xy→0 → a_∅→1 → Ψ 中性（所有 (S,bmask) ≈1，logPsi≈0）。
func TestPsiFarFromBedNeutral(t *testing.T) {
	c := NewCoupling([]BedGeom{{Covers: 1, Onbed: 1, Overlap: 1}})
	js := NewJointSpace(1)
	logPsi := c.LogPsi(js, []float64{0.0}) // g^xy=0：雷达看不见任何床
	for i := 0; i < js.size; i++ {
		if math.Abs(logPsi[i]) > 0.05 { // exp≈1 ±5%
			s, b := js.decode(i)
			t.Errorf("远离床 logPsi(S=%v,bmask=%d)=%.4f 应≈0（Ψ 中性）", s, b, logPsi[i])
		}
	}
}
