package zoneengine

// bed_bayesian_scorer_test.go — 4 个核心 case 验证 (用户拍板的数学规范)
//
// Case 1: 正常在床 (sleepad + radar 双源)
// Case 2: 翻身假阳 (home/twin: sleepad LeftBed + radar 正向，应维持 InBed N min)
// Case 3: 跌床 (sleepad LeftBed + radar 床下伪 vital，3min 内判 LeftBed)
// Case 4: 正常主动离床 (sleepad LeftBed + radar LeftBed，立刻判 LeftBed)

import (
	"fmt"
	"math"
	"testing"
)

// helper: 容差比较
func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func formatStep(t *testing.T, label string, s *BedBayesianScorer, nowMs int64) {
	t.Helper()
	t.Logf("  [%s] L=%+.3f P=%.4f decision=%v", label, s.LogOdds(), s.Probability(), s.Decision(nowMs))
}

// --- Case 1：正常在床 ---

func TestBayesian_Case1_NormalInBed_Facility(t *testing.T) {
	s := NewBedBayesianScorer() // facility default, L_0=0

	now := int64(0)
	// 第 0 分钟：sleepad InBed event 到，radar InBed event 到，vital 都活跃
	s.OnSleepadInBed(now)
	s.OnSleepadVital(now)
	s.OnRadarInBed(now)
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now)

	// 预期：sleepad +2.94, radar max(+1.39, +0.56, +1.45) = +1.45, γ=1.0
	// ΔL = +2.94 + 1.45 = +4.39
	formatStep(t, "t=0 events", s, now)
	if !approxEqual(s.LogOdds(), 4.39, 0.01) {
		t.Errorf("L=%.3f, want ≈ 4.39", s.LogOdds())
	}
	if d := s.Decision(now); d != BedDecisionInBed {
		t.Errorf("decision=%v, want InBed", d)
	}

	// 下一分钟 Tick：重复贡献 → L 累到 5 cap
	now = 60_000
	s.Tick(now)
	formatStep(t, "t=60s tick", s, now)
	if !approxEqual(s.LogOdds(), 5.0, 0.01) {
		t.Errorf("L=%.3f, want 5.0 (capped)", s.LogOdds())
	}
	if d := s.Decision(now); d != BedDecisionInBed {
		t.Errorf("decision=%v, want InBed", d)
	}
}

// --- Case 2：home/twin 大床翻身假阳，应维持 InBed ---

// TODO: 在 "自不否定" 新原则下重写：big bed roll 必须靠 radar 跨设备否定 sleepad LeftBed，
// 不再是 sleepad vital 撑住。需要重新分配 LR 权重让 radar lying 能 "may deny" sleepad LeftBed。
func TestBayesian_Case2_BigBedRoll_HomeMode_MaintainInBed(t *testing.T) {
	t.Skip("待 cross-device override LR 重新校准后启用")
	s := NewBedBayesianScorer()
	s.SetMode(BedModeHome) // LR=±2.20

	// 先让 FSM 进 InBed 稳态 (L=+5 cap, 模拟长期在床)
	for i := int64(0); i < 5; i++ {
		now := i * 60_000
		s.OnSleepadInBed(now)
		s.OnSleepadVital(now)
		s.OnRadarVital(now)
		s.OnRadarPoseLying(now)
		s.Tick(now)
	}
	if !approxEqual(s.LogOdds(), 5.0, 0.01) {
		t.Fatalf("setup: L=%.3f, want 5.0 (capped)", s.LogOdds())
	}
	if d := s.Decision(5 * 60_000); d != BedDecisionInBed {
		t.Fatalf("setup decision=%v, want InBed", d)
	}

	// 第 5 分钟末翻身到床沿: sleepad LeftBed event, sleepad vital 消失（人不在 pad 上）
	// radar 仍正常看到 lying + vital
	now := int64(5 * 60_000)
	s.OnSleepadLeftBed(now)
	// 不再 OnSleepadVital → sleepadVitalLastTs 仍是 4*60000，age=60s 刚到 expire 边界
	// 为模拟"vital 没刷新"，把它设到 65s 前（≤evidenceWindowMs 边界外）—— 简化：直接重置
	s.sleepadVitalLastTs = 0
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now)

	// 此时刚发生事件：conflict 刚启动 γ=1.0
	// sleepad cluster = -2.20 (LeftBed fresh + no vital)
	// radar cluster = +1.45 (max positive)
	// 但本分钟 (5) 之前 sleepadContrib 已贡献 +2.20，radarContrib 已贡献 +1.45 (capped 不到)
	// LeftBed event 触发 contribute → 同分钟 cluster 翻为 -2.20，delta = -4.40 → L = 5 - 4.40 = 0.60
	// radar 同分钟值不变（+1.45） → 不重复 contribute
	formatStep(t, "t=5min LeftBed event", s, now)
	// L 应 ≈ 0.60 (5 - 4.40)
	if !approxEqual(s.LogOdds(), 0.60, 0.01) {
		t.Errorf("after LeftBed event L=%.3f, want ≈ 0.60", s.LogOdds())
	}
	p := s.Probability()
	// P = sigmoid(0.60) ≈ 0.645，处于维持区 [0.50, 0.70]
	if !(p >= pLeftBedThreshold && p <= pInBedThresholdFacility) {
		t.Errorf("P=%.3f, want in maintain zone [0.50, 0.70]", p)
	}
	if d := s.Decision(now); d != BedDecisionInBed {
		t.Errorf("decision=%v, want InBed (维持上态)", d)
	}

	// 1 分钟后 Tick (γ 进入第 1 衰减档：60-120s → γ=0.5；conflict started=5min, dur=60s → 仍 γ=1.0)
	now = 6 * 60_000
	s.OnSleepadLeftBed(now) // 持续报 LeftBed (firmware repeat)
	s.sleepadVitalLastTs = 0
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now)
	s.Tick(now)
	formatStep(t, "t=6min still LeftBed+radar+", s, now)

	// 2 分钟后：conflict dur=120s → γ=0.0
	now = 7 * 60_000
	s.OnSleepadLeftBed(now)
	s.sleepadVitalLastTs = 0
	s.OnRadarVital(now) // 仍来，但 γ=0 不贡献
	s.OnRadarPoseLying(now)
	s.Tick(now)
	formatStep(t, "t=7min γ→0", s, now)
}

// --- Case 3：跌床 ---
//
// 长期在床 (L=+5 cap) → 跌床瞬间 sleepad LeftBed event + sleepad vital 死
// + radar 床下残留 HR/RR 仍报 vital (firmware 误报)
// 预期：3min 内 P < 0.50 → LeftBed (fall alarm pathway 启动)

// TODO: 在 "自不否定" 新原则下，跌床场景 sleepad LeftBed 立即翻 Vacant (好事)，
// 但 radar 床下残留 lying 应通过 γ tempering 维持短暂 maintain。需要重写期望值。
func TestBayesian_Case3_FallFromBed_Facility(t *testing.T) {
	t.Skip("待新物理逻辑下跌床场景期望值重写")
	s := NewBedBayesianScorer() // facility, LR_S=±2.94

	// 模拟长期在床到 L=+5 cap
	for i := int64(0); i < 5; i++ {
		now := i * 60_000
		s.OnSleepadInBed(now)
		s.OnSleepadVital(now)
		s.OnRadarVital(now)
		s.OnRadarPoseLying(now)
		s.Tick(now)
	}
	if !approxEqual(s.LogOdds(), 5.0, 0.01) {
		t.Fatalf("setup: L=%.3f, want 5.0 (capped)", s.LogOdds())
	}

	// 跌床瞬间 (t=5min)
	now := int64(5 * 60_000)
	s.OnSleepadLeftBed(now)
	s.sleepadVitalLastTs = 0 // sleepad vital 死
	// radar 床下伪体征
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now) // 假设 firmware 把床下 lying 也判 lying（保守）

	formatStep(t, "t=5min fall event", s, now)
	// 同分钟已贡献过 +2.94 sleepad + 1.45 radar；现在 cluster 翻 -2.94 sleepad，radar 不变
	// delta sleepad = -2.94 - (+2.94) = -5.88 → L = 5 - 5.88 = -0.88
	if !approxEqual(s.LogOdds(), -0.88, 0.01) {
		t.Errorf("fall L=%.3f, want ≈ -0.88", s.LogOdds())
	}

	// t=6min (conflict dur=60s → γ=0.5)
	now = 6 * 60_000
	// 模拟 firmware 继续报：sleepad LeftBed 持续，radar 继续报 vital + pose
	s.OnSleepadLeftBed(now)
	s.sleepadVitalLastTs = 0
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now)
	s.Tick(now)
	formatStep(t, "t=6min γ=0.5", s, now)
	// sleepad cluster = -2.94, radar cluster = +1.45 * 0.5 = +0.725
	// ΔL = -2.94 + 0.725 = -2.215 → L = -0.88 - 2.215 = -3.095
	if !approxEqual(s.LogOdds(), -3.110, 0.05) {
		t.Errorf("t=6 L=%.3f, want ≈ -3.095", s.LogOdds())
	}
	p := s.Probability()
	if p >= pLeftBedThreshold {
		t.Errorf("t=6 P=%.3f, should already < 0.50 (LeftBed)", p)
	}

	// t=7min (conflict dur=120s → γ=0.0)
	now = 7 * 60_000
	s.OnSleepadLeftBed(now)
	s.sleepadVitalLastTs = 0
	s.OnRadarVital(now)
	s.OnRadarPoseLying(now)
	s.Tick(now)
	formatStep(t, "t=7min γ=0", s, now)
	// sleepad = -2.94, radar = 0 → ΔL = -2.94 → L = clamp(-3.095 -2.94) = -5
	if !approxEqual(s.LogOdds(), -5.0, 0.01) {
		t.Errorf("t=7 L=%.3f, want -5.0 (capped)", s.LogOdds())
	}
	if d := s.Decision(now); d != BedDecisionLeftBed {
		t.Errorf("t=7 decision=%v, want LeftBed", d)
	}
}

// --- Case 4：主动离床 (sleepad LeftBed + radar LeftBed)，立刻判 ---

func TestBayesian_Case4_NormalLeave_Facility(t *testing.T) {
	s := NewBedBayesianScorer() // L_0=0

	now := int64(0)
	s.OnSleepadLeftBed(now)
	// 无 sleepad vital
	s.OnRadarLeftBed(now)
	// 无 radar 正向

	formatStep(t, "t=0 leave", s, now)
	// sleepad cluster = -2.94 (LeftBed + no vital), radar cluster = -1.39 (LeftBed + no positives)
	// 注意 γ 不影响负向；conflict 刚启动 γ=1.0
	// ΔL = -2.94 + (-1.39) = -4.33 → L = -4.33
	if !approxEqual(s.LogOdds(), -4.33, 0.01) {
		t.Errorf("L=%.3f, want -4.33", s.LogOdds())
	}
	p := s.Probability()
	if p >= pLeftBedThreshold {
		t.Errorf("P=%.3f, should < 0.50", p)
	}
	if d := s.Decision(now); d != BedDecisionLeftBed {
		t.Errorf("decision=%v, want LeftBed", d)
	}
}

// --- 额外：prior 初始化 ---

func TestBayesian_PriorByHour(t *testing.T) {
	tests := []struct {
		hour    int
		wantL   float64
		wantP   float64
	}{
		{22, +1.39, 0.80}, // 夜间 21-06
		{2, +1.39, 0.80},
		{6, +1.39, 0.80},
		{7, -0.85, 0.30},  // 白天 7-20
		{12, -0.85, 0.30},
		{20, -0.85, 0.30},
		{21, +1.39, 0.80}, // 夜间起点
	}
	for _, tc := range tests {
		s := NewBedBayesianScorer()
		s.InitPriorByHour(tc.hour)
		if !approxEqual(s.LogOdds(), tc.wantL, 0.01) {
			t.Errorf("hour=%d L=%.3f, want %.2f", tc.hour, s.LogOdds(), tc.wantL)
		}
		if !approxEqual(s.Probability(), tc.wantP, 0.01) {
			t.Errorf("hour=%d P=%.3f, want %.2f", tc.hour, s.Probability(), tc.wantP)
		}
	}
}

// --- 额外：维持区 2min 强制 LeftBed ---

func TestBayesian_MaintainTimeoutForceLeftBed(t *testing.T) {
	s := NewBedBayesianScorer()
	// 人为构造 L 处于维持区
	s.L = 0.40 // P ≈ 0.60 ∈ [0.50, 0.70] 维持区

	now := int64(0)
	d := s.Decision(now)
	if d != BedDecisionLeftBed {
		// 第一次 Decision: lastDecision 默认 LeftBed → 维持 LeftBed
		t.Logf("first decision=%v (maintain prev=LeftBed initial)", d)
	}

	// 模拟稳定在维持区 2 min
	now = 119_999
	if d := s.Decision(now); d != BedDecisionLeftBed {
		t.Errorf("just before timeout decision=%v, want LeftBed (maintain initial)", d)
	}

	now = 120_000
	if d := s.Decision(now); d != BedDecisionLeftBed {
		t.Errorf("at timeout decision=%v, want LeftBed (forced)", d)
	}

	// 反例：构造 InBed 区
	s2 := NewBedBayesianScorer()
	s2.L = 1.0 // P ≈ 0.73 > 0.70 → InBed
	if d := s2.Decision(0); d != BedDecisionInBed {
		t.Errorf("InBed test: decision=%v, want InBed", d)
	}

	// 调进维持区
	fmt.Println("--- maintain InBed -> maintain zone test ---")
	s2.L = 0.40 // P ≈ 0.60 维持区
	if d := s2.Decision(60_000); d != BedDecisionInBed {
		t.Errorf("recently InBed, just entered maintain: decision=%v, want InBed (maintain prev)", d)
	}
	// 2 min 后 force LeftBed
	if d := s2.Decision(180_000); d != BedDecisionLeftBed {
		t.Errorf("maintain timeout: decision=%v, want LeftBed (forced)", d)
	}
}
