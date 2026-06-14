package belief

import (
	"math"
	"testing"
)

// decide_test.go — §26 55% 三分判据验收（资源稀缺前提，用户裁定）：
//   DEC1 ≥55% 报 + T_hold 持续（瞬时断开复位）
//   DEC2 45-55% 可判两可：C_FN 风险偏好打破平衡（独处报 / 多人不报）
//   DEC3 高度不可判（Λ 无判别力）默认不报 —— 即使独处（§26 推翻旧「不可判兜底 fire」）
//   DEC4 ≤45% 不报 —— 即使独处（C_FN 不再救低 P^F）
//   DEC5 ≥55% 报 与风险无关（证据自足，多人也报）
//   DEC6 ComputeLambda（informative≫1 / 全暗→1）

const deps = 1e-9

var ident = 5.0   // Λ 可判（> lambdaInformative=3）
var unident = 1.0 // Λ 高度不可判

// DEC1：≥55% 报 + 持续。
func TestDecideReportSustain(t *testing.T) {
	d := NewDecider()
	rc := RiskContext{PeopleCount: 1}
	if dec := d.Step(1000, 0.6, ident, rc); !dec.InstFire || dec.Fire || dec.Band != "report" {
		t.Fatalf("P^F=0.6 应 InstFire/report 未到 T_hold 不 Fire；得 inst=%v fire=%v band=%s", dec.InstFire, dec.Fire, dec.Band)
	}
	if dec := d.Step(91_000, 0.6, ident, rc); !dec.Fire {
		t.Errorf("证据自足档持续 90s 应 Fire")
	}
	// 瞬时断开复位
	d2 := NewDecider()
	d2.Step(1000, 0.6, ident, rc)
	d2.Step(2000, 0.2, ident, rc) // 跌破 → 复位
	if dec := d2.Step(2000+91_000, 0.2, ident, rc); dec.Fire {
		t.Errorf("跌破后不应 Fire（持续复位）")
	}
}

// DEC2：45-55% 可判两可 → C_FN 打破平衡。独处报、多人不报。
func TestDecideTieWindow(t *testing.T) {
	if dec := NewDecider().Step(1000, 0.50, ident, RiskContext{AloneContinuousMin: 30, Night: true, PeopleCount: 1}); !dec.InstFire || dec.Band != "tie" {
		t.Errorf("两可+独处+夜 应 tie 报（cFN=%.1f band=%s）", dec.CFN, dec.Band)
	}
	if dec := NewDecider().Step(1000, 0.50, ident, RiskContext{PeopleCount: 3}); dec.InstFire {
		t.Errorf("两可+多人 不应报（有人代发现，band=%s cFN=%.2f）", dec.Band, dec.CFN)
	}
}

// DEC3（§26 核心反转）：高度不可判 → 默认不报，即使独处。
func TestDecideIndeterminateNoReport(t *testing.T) {
	dec := NewDecider().Step(1000, 0.50, unident, RiskContext{AloneContinuousMin: 30, Night: true, Disabled: true, PeopleCount: 1})
	if dec.Identifiable {
		t.Fatalf("Λ=%.1f 应判高度不可判", unident)
	}
	if dec.InstFire || dec.Band != "indeterminate" {
		t.Errorf("高度不可判应默认不报（即使独处高风险）band=%s inst=%v —— §26 守告警可信度", dec.Band, dec.InstFire)
	}
}

// DEC4：≤45% 不报，即使独处（C_FN 不救低 P^F，§26 收窄作用域）。
func TestDecideLowNoReport(t *testing.T) {
	dec := NewDecider().Step(1000, 0.30, ident, RiskContext{AloneContinuousMin: 30, Night: true, Disabled: true, PeopleCount: 1})
	if dec.InstFire || dec.Band != "no" {
		t.Errorf("P^F=0.30≤45%% 应不报（即使独处，C_FN 作用域不含此档）band=%s", dec.Band)
	}
}

// DEC5：≥55% 报 与风险无关（证据自足）。
func TestDecideReportSelfSufficient(t *testing.T) {
	dec := NewDecider().Step(1000, 0.70, ident, RiskContext{PeopleCount: 5}) // 多人
	if !dec.InstFire || dec.Band != "report" {
		t.Errorf("P^F=0.70≥55%% 多人也应报（证据自足，C_FN 不需要）band=%s", dec.Band)
	}
}

// DEC6：ComputeLambda 诊断。
func TestComputeLambda(t *testing.T) {
	js := NewJointSpace(1)
	if l := ComputeLambda(js, js.NewJointVector(), js.NewJointVector()); math.Abs(l-1.0) > deps {
		t.Errorf("全暗 Λ=%.4f 应=1", l)
	}
	logPhi := js.NewJointVector()
	for b := 0; b < js.bmaskN; b++ {
		logPhi[js.idx(SFallen, b)] = math.Log(4)
		logPhi[js.idx(SBed, b)] = math.Log(0.25)
	}
	if l := ComputeLambda(js, js.NewJointVector(), logPhi); l < 10 {
		t.Errorf("informative Λ=%.2f 应≫1", l)
	}
}
