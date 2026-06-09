package belief

import (
	"testing"

	"owl-common/observation"
)

// TestReasonForMapping — P7.3 主导 ObsKind → reason 类（纯映射）。
func TestReasonForMapping(t *testing.T) {
	cases := map[ObsKind]FallReason{
		ObsNoDetect:    ReasonLost,
		ObsDwellStill:  ReasonSilent,
		ObsPose:        ReasonPoseLying,
		ObsNeighbor:    ReasonUnknown, // 压制源不是 fall 成因
		ObsBedOccupied: ReasonUnknown,
		ObsKind(-1):    ReasonUnknown, // 无主导
	}
	for k, want := range cases {
		if got := reasonFor(k); got != want {
			t.Fatalf("reasonFor(%v)=%v,期望 %v", k, got, want)
		}
	}
}

// 构造各 fall-ward obs（rawLikelihood[SFallen]>1）。
func poseFallenObs() Observation {
	return Observation{Kind: ObsPose, Value: float64(observation.PoseFallen), Conf: 1, Fresh: true, Geom: GeomOpenFloor, GeomConf: 1}
}
func dwellStillObs() Observation {
	return Observation{Kind: ObsDwellStill, Value: 1200, Conf: 1, Fresh: true, Geom: GeomInToilet}
}
func noDetectObs() Observation {
	return Observation{Kind: ObsNoDetect, Conf: 1, Fresh: true, Geom: GeomOpenFloor, RealnessP: 0.9, DoorExitP: 0.0}
}

// TestFallReasonForEachPath — 各主导 obs → 对应 reason（端到端，经 rawLikelihood 真算）。
func TestFallReasonForEachPath(t *testing.T) {
	cases := []struct {
		name string
		obs  Observation
		want FallReason
	}{
		{"pose_fallen→pose_lying", poseFallenObs(), ReasonPoseLying},
		{"dwell_still→silent", dwellStillObs(), ReasonSilent},
		{"no_detect→lost", noDetectObs(), ReasonLost},
	}
	for _, c := range cases {
		// 先确认确实 fall-ward（LR>1），否则测的是 Unknown 退化。
		if lr := rawLikelihood(c.obs)[SFallen]; lr <= 1.0 {
			t.Fatalf("%s：obs 应 fall-ward(LR>1)，得 LR=%.3f（构造无效）", c.name, lr)
		}
		got, _, _ := FallReasonFor([]Observation{c.obs})
		if got != c.want {
			t.Fatalf("%s：FallReasonFor=%v,期望 %v", c.name, got, c.want)
		}
	}
}

// TestDominantFallObsPicksMaxIgnoresSuppressive — 多 obs 取 SFallen-LR 最高者;压制源(LR<1)与中性不入选。
func TestDominantFallObsPicksMaxIgnoresSuppressive(t *testing.T) {
	suppressive := Observation{Kind: ObsNeighbor, Value: 1, Conf: 1, Fresh: true, Geom: GeomUnknown} // LR<1 压 fall
	neutral := Observation{Kind: ObsZBand, Value: 100, Conf: 1, Fresh: true}                         // z stand,不写 SFallen
	stale := poseFallenObs()
	stale.Fresh = false // stale → effConf 0 → 不计

	dom, lr := DominantFallObs([]Observation{suppressive, neutral, stale, poseFallenObs()})
	if dom != ObsPose {
		t.Fatalf("dominant 应为 ObsPose(唯一 fall-ward 且最高),得 %v", dom)
	}
	if lr <= 1.0 {
		t.Fatalf("dominant LR 应 >1,得 %.3f", lr)
	}

	// 全压制/中性/stale → 无 fall-ward 主导 → Unknown。
	r, _, _ := FallReasonFor([]Observation{suppressive, neutral, stale})
	if r != ReasonUnknown {
		t.Fatalf("无 fall-ward obs 应 ReasonUnknown,得 %v", r)
	}
}
