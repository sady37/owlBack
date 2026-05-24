// card_display_builder_test.go — W2 VitalTrendLevel 派生 T1 单元测试。
//
// 覆盖：4 档阈值 + 边界 + Target=nil 兜底。

package consumer

import (
	"testing"

	"owl-common/card"
)

func TestPickVitalTrendLevel(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  int
	}{
		{"0", 0, card.VitalTrendLevelNone},
		{"29 boundary low", 29, card.VitalTrendLevelNone},
		{"30 boundary gray", 30, card.VitalTrendLevelGray},
		{"45 mid gray", 45, card.VitalTrendLevelGray},
		{"59 boundary gray", 59, card.VitalTrendLevelGray},
		{"60 boundary yellow", 60, card.VitalTrendLevelYellow},
		{"70 mid yellow", 70, card.VitalTrendLevelYellow},
		{"79 boundary yellow", 79, card.VitalTrendLevelYellow},
		{"80 boundary red", 80, card.VitalTrendLevelRed},
		{"90 mid red", 90, card.VitalTrendLevelRed},
		{"100 cap red", 100, card.VitalTrendLevelRed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &card.CardStatus{Target: &card.TargetState{WeakBiometricSignal: tc.score}}
			if got := pickVitalTrendLevel(s); got != tc.want {
				t.Errorf("score=%d: got %d want %d", tc.score, got, tc.want)
			}
		})
	}
}

func TestPickVitalTrendLevel_NilTarget(t *testing.T) {
	s := &card.CardStatus{Target: nil}
	if got := pickVitalTrendLevel(s); got != card.VitalTrendLevelNone {
		t.Errorf("nil Target: got %d want None", got)
	}
}

// BuildCardDisplay 端到端：VitalTrendLevel 进 display。
func TestBuildCardDisplay_VitalTrendLevelFromTarget(t *testing.T) {
	s := &card.CardStatus{
		CardID: "fd00:0:3:111:3:101::/96",
		Target: &card.TargetState{WeakBiometricSignal: 65, UpdatedAt: 1_700_000_000_000},
	}
	d := BuildCardDisplay(s, nil)
	if d == nil {
		t.Fatal("display nil")
	}
	if d.VitalTrendLevel != card.VitalTrendLevelYellow {
		t.Errorf("expected Yellow (65), got %d", d.VitalTrendLevel)
	}
}

func TestBuildCardDisplay_VitalTrendLevel_NoTargetNoTrend(t *testing.T) {
	s := &card.CardStatus{CardID: "fd00:0:3:111:3:101::/96"}
	d := BuildCardDisplay(s, nil)
	if d.VitalTrendLevel != card.VitalTrendLevelNone {
		t.Errorf("no Target should be None, got %d", d.VitalTrendLevel)
	}
}
