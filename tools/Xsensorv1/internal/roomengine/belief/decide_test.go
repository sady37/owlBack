package belief

import (
	"math"
	"testing"
)

// decide_test.go — §8 期望损失裁决 + A 阶段3 三立场验收：
//   DEC1 fire 条件 + T_hold 持续（瞬时满足须持续 ≥ T_hold；瞬时断开复位）
//   DEC2 风险分层（同 P^F：独处+夜→发；多人白天→不发）
//   DEC3 代价翻转非 argmax（P^F=0.4<AtBed 0.45，独处高 C_FN 仍发）
//   DEC4 不可判由同一框架处理（无独立分支）：Λ→1 不可判，高风险独处仍 fire；Λ 不 gate
//   DEC5 ComputeLambda 诊断（informative≫1 / 全暗→1）

const deps = 1e-9

// DEC1：fire 条件 + T_hold 持续。
func TestDecideSustain(t *testing.T) {
	d := NewDecider()
	rc := RiskContext{PeopleCount: 1} // 独处白天，cFN=base=2
	// P^F=0.5：margin=0.5*2-0.5*1=0.5>0 瞬时满足
	if dec := d.Step(1000, 0.5, 5.0, rc); !dec.InstFire || dec.Fire {
		t.Fatalf("t=1000 应 InstFire 但未到 T_hold 不 Fire；得 inst=%v fire=%v", dec.InstFire, dec.Fire)
	}
	if dec := d.Step(50_000, 0.5, 5.0, rc); dec.Fire {
		t.Errorf("t=50s（49s<90s）不应 Fire")
	}
	if dec := d.Step(91_000, 0.5, 5.0, rc); !dec.Fire {
		t.Errorf("t=91s（≥90s 持续）应 Fire")
	}
	// 瞬时断开复位：新 decider
	d2 := NewDecider()
	d2.Step(1000, 0.5, 5, rc)            // 武装
	d2.Step(2000, 0.0, 5, rc)            // margin<0 → 复位
	dec := d2.Step(2000+91_000, 0.0, 5, rc)
	if dec.Fire {
		t.Errorf("瞬时断开后 P^F=0 不应 Fire（持续已复位）")
	}
}

// DEC2：风险分层——同 P^F=0.3，独处+夜 fire；多人白天不 fire。
func TestDecideRiskStratified(t *testing.T) {
	pF := 0.3
	aloneNight := RiskContext{AloneContinuousMin: 30, Night: true, PeopleCount: 1}
	if dec := NewDecider().Step(1000, pF, 5, aloneNight); !dec.InstFire {
		t.Errorf("独处+夜 P^F=0.3 应瞬时 fire（cFN=%.1f margin=%.2f）", dec.CFN, dec.Margin)
	}
	multiDay := RiskContext{PeopleCount: 3} // 白天多人
	if dec := NewDecider().Step(1000, pF, 5, multiDay); dec.InstFire {
		t.Errorf("多人白天 P^F=0.3 不应 fire（有人代发现，cFN=%.2f margin=%.2f）", dec.CFN, dec.Margin)
	}
}

// DEC3：代价翻转非 argmax。P^F=0.4（< 假想 AtBed 0.45，argmax 会选 AtBed），
// 独处高 C_FN → 仍 fire；同 P^F 多人 → 不 fire。证明裁决靠代价不对称非 P^F 量级。
func TestDecideCostFlipNotArgmax(t *testing.T) {
	pF := 0.4 // argmax 下输给 AtBed 0.45
	alone := RiskContext{AloneContinuousMin: 30, Night: true, PeopleCount: 1}
	if dec := NewDecider().Step(1000, pF, 1.0, alone); !dec.InstFire {
		t.Errorf("独处 P^F=0.4 应 fire（代价翻转，非 argmax）margin=%.2f", dec.Margin)
	}
	multi := RiskContext{PeopleCount: 4}
	if dec := NewDecider().Step(1000, pF, 1.0, multi); dec.InstFire {
		t.Errorf("多人 P^F=0.4 不应 fire（同 P^F，代价不翻转）margin=%.2f", dec.Margin)
	}
}

// DEC4：不可判由同一框架处理（A 立场①：无独立分支，Λ 不 gate）。
// Λ→1（全暗不可判）+ 高风险独处低 P^F=0.1 → 期望损失仍翻转 fire；Identifiable=false 不阻断。
func TestDecideUnidentifiableNoSpecialBranch(t *testing.T) {
	lambda := 1.0 // 全暗不可判
	alone := RiskContext{AloneContinuousMin: 30, Night: true, Disabled: true, PeopleCount: 1}
	dec := NewDecider().Step(1000, 0.1, lambda, alone)
	if dec.Identifiable {
		t.Fatalf("Λ=1 应判不可判（Identifiable=false）")
	}
	if !dec.InstFire {
		t.Errorf("不可判但高风险独处 P^F=0.1 应仍 fire（同一期望损失框架，无 special-case）cFN=%.1f margin=%.2f", dec.CFN, dec.Margin)
	}
}

// DEC5：ComputeLambda 诊断。
func TestComputeLambda(t *testing.T) {
	js := NewJointSpace(1)
	// 全暗：logPsi/logPhi 全 0 → Λ=1
	if l := ComputeLambda(js, js.NewJointVector(), js.NewJointVector()); math.Abs(l-1.0) > deps {
		t.Errorf("全暗 Λ=%.4f 应=1", l)
	}
	// informative：logPhi 抬 F 压 AtBed → Λ≫1
	logPhi := js.NewJointVector()
	for b := 0; b < js.bmaskN; b++ {
		logPhi[js.idx(SFallen, b)] = math.Log(4) // 抬 F
		logPhi[js.idx(SBed, b)] = math.Log(0.25) // 压 AtBed
	}
	if l := ComputeLambda(js, js.NewJointVector(), logPhi); l < 10 {
		t.Errorf("informative Λ=%.2f 应≫1（F 抬 AtBed 压）", l)
	}
}
