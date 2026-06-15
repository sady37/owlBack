package belief

import (
	"math"
	"testing"
)

// filter_fusion_test.go — W3.1 四轴融合验收（DBN-wire-roadmap §6，C 复审 gate①）：
//   WF1 零回归 oracle：中性 (rho=0, pFallReal=1) → Step 逐 tick 等价 Predict(·,0)+Correct（融合包装透明）
//   WF2 neighbor 整流：rho>0 → Blind from-行 →Fallen 整流入 →Left（P(Fallen) 降，转移先验非 gate）
//   WF3 realness 调制：pFallReal<1 → SFallen 发射被 ×P(real) 压低（ghost 摔喂不动）；=1 不压（真人摔不抑制）

const wfeps = 1e-12

// 构造一个非平凡 logPhi（偏向若干态），让 alpha 真演化、等价测试有意义。
func nontrivialPhi(js *JointSpace) JointVector {
	phi := js.NewJointVector()
	phi[js.idx(SOpenFloor, 0)] = math.Log(3)
	phi[js.idx(SFallen, 0)] = math.Log(2)
	return phi
}

// WF1：中性融合 = 恒等。Step(...,0,1) 必须逐 tick 等价 Predict(online,0)+Correct(psi,phi)。
func TestFusionNeutralZeroRegression(t *testing.T) {
	fA := NewFilter(DefaultModel(), 1)
	fB := NewFilter(DefaultModel(), 1)
	js := fA.space
	online := BedOnline{false}
	for k := 1; k <= 6; k++ {
		now := int64(k) * 1000
		phi := nontrivialPhi(js)
		fA.Step(now, online, nil, phi, 0, 1) // 融合中性
		// B：Step 不含融合时的底层操作
		if fB.lastTs > 0 && now > fB.lastTs {
			fB.Predict(online, 0)
		}
		fB.Correct(nil, phi)
		if now > fB.lastTs {
			fB.lastTs = now
		}
		for i := range fA.alpha {
			if math.Abs(fA.alpha[i]-fB.alpha[i]) > wfeps {
				t.Fatalf("tick %d 态 %d：中性融合 α=%.15g ≠ 基线 %.15g（融合包装不透明=wire 接错）", k, i, fA.alpha[i], fB.alpha[i])
			}
		}
	}
	t.Logf("WF1 ✓ 零回归 oracle：6 tick 中性融合逐态等价基线（realness+neighbor 中性 = S/B-only）")
}

// WF2：neighbor rho>0 → Blind 行 F 整流入 L。
func TestFusionNeighborRectifiesBlind(t *testing.T) {
	mk := func() *Filter {
		f := NewFilter(DefaultModel(), 1)
		js := f.space
		for i := range f.alpha {
			f.alpha[i] = math.Inf(-1)
		}
		f.alpha[js.idx(SBlindOpen, 0)] = 0 // 全质量压 Blind-Open（lost-track）
		f.alpha.LogNormalize()
		f.lastTs = 1000
		return f
	}
	f0 := mk()
	f0.Predict(BedOnline{false}, 0) // 无 hand-off
	fR := mk()
	fR.Predict(BedOnline{false}, 0.8) // fresh hand-off
	p0 := f0.space.PFallen(f0.alpha)
	pR := fR.space.PFallen(fR.alpha)
	pL0 := f0.space.MarginalS(f0.alpha)[SLeft]
	pLR := fR.space.MarginalS(fR.alpha)[SLeft]
	if !(pR < p0) {
		t.Errorf("rho>0 应整流 F→L：P(Fallen) rho0=%.4f 应 > rho0.8=%.4f", p0, pR)
	}
	if !(pLR > pL0) {
		t.Errorf("rho>0 应整流 F→L：P(Left) rho0=%.4f 应 < rho0.8=%.4f", pL0, pLR)
	}
	t.Logf("WF2 ✓ neighbor 整流：Blind 出发 P(Fallen) %.3f→%.3f、P(Left) %.3f→%.3f（人挪去邻房非真摔）", p0, pR, pL0, pLR)
}

// WF3：realness pFallReal<1 压 SFallen 发射；=1 不压。
func TestFusionRealnessDampsGhostFall(t *testing.T) {
	// 均匀起点（prior 的 SFallen=0 会让两边都为 0，观测不到压制）。
	mk := func() *Filter {
		f := NewFilter(DefaultModel(), 1)
		for i := range f.alpha {
			f.alpha[i] = 0
		}
		f.alpha.LogNormalize()
		return f
	}
	js := mk().space
	phi := js.NewJointVector()
	phi[js.idx(SFallen, 0)] = math.Log(10) // 强 fall 证据（lying+still）

	fReal := mk()
	fReal.Correct(nil, fReal.foldRealness(phi, 1.0)) // 真人：不压
	fGhost := mk()
	fGhost.Correct(nil, fGhost.foldRealness(phi, 0.1)) // ghost：×0.1

	pReal := fReal.space.PFallen(fReal.alpha)
	pGhost := fGhost.space.PFallen(fGhost.alpha)
	if !(pGhost < pReal) {
		t.Errorf("ghost 摔应喂不动 SFallen：P(Fallen) real(1.0)=%.4f 应 > ghost(0.1)=%.4f", pReal, pGhost)
	}
	// 共生律侧：pFallReal=1（真人在场，含 track 消失）→ 与无 realness 折叠等价（不抑制真摔）。
	fBase := mk()
	fBase.Correct(nil, phi)
	if math.Abs(fReal.space.PFallen(fReal.alpha)-fBase.space.PFallen(fBase.alpha)) > wfeps {
		t.Errorf("pFallReal=1 应不抑制（真人摔不被当 ghost）：%.6f vs 基线 %.6f", pReal, fBase.space.PFallen(fBase.alpha))
	}
	t.Logf("WF3 ✓ realness 调制：P(Fallen) 真人=%.4f > ghost=%.4f；真人=1 等价基线（不抑制真摔）", pReal, pGhost)
}
