package belief

import (
	"testing"

	"owl-common/observation"
)

// TestReasonForMapping — P7.3 主导 ObsKind → reason 类（纯映射）。
func TestReasonForMapping(t *testing.T) {
	cases := map[ObsKind]FallReason{
		ObsNoDetect:    ReasonMoving,
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
	return Observation{Kind: ObsPose, Value: float64(observation.PoseFallen), Conf: 1, Fresh: true, AreaType: areaActive, AreaConf: 1}
}
func dwellStillObs() Observation {
	return Observation{Kind: ObsDwellStill, Value: 1200, Conf: 1, Fresh: true, AreaType: areaToilet}
}
func noDetectObs() Observation {
	return Observation{Kind: ObsNoDetect, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 0.9, DoorExitP: 0.0}
}

// TestNoDetectRealnessNeutralizes — P2 距离闸用的全压杠杆:RealnessP=0 → no-detect fall lift **完全中性化**
// (factor=1)。距离闸把"远距弱回波"映射成 effRealnessP=0 → 远距 no-detect 不抬 fall。
// （★对照:DoorExitP=1 不全中性化——有意留 floor「门口真摔仍浮出」,故距离闸不走 door-exit 通道。）
func TestNoDetectRealnessNeutralizes(t *testing.T) {
	lift := rawLikelihood(Observation{Kind: ObsNoDetect, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 0.9, DoorExitP: 0.0})[SFallen]
	if lift <= 1.0 {
		t.Fatalf("RealnessP=0.9 应抬 fall(LR>1),得 %.3f", lift)
	}
	neutral := rawLikelihood(Observation{Kind: ObsNoDetect, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 0.0, DoorExitP: 0.0})[SFallen]
	if neutral > 1.0+1e-9 {
		t.Fatalf("RealnessP=0(距离闸全压)应中性化 fall lift(factor=1),得 %.3f", neutral)
	}
	// door-exit 有 floor:DoorExitP=1 仍 >1（不全否决），证距离闸不可复用此通道。
	doorFloor := rawLikelihood(Observation{Kind: ObsNoDetect, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 0.9, DoorExitP: 1.0})[SFallen]
	if doorFloor <= 1.0+1e-9 {
		t.Fatalf("door-exit 应留 floor(LR>1,门口真摔仍浮出),得 %.3f", doorFloor)
	}
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
		{"no_detect→moving", noDetectObs(), ReasonMoving},
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
	suppressive := Observation{Kind: ObsNeighbor, Value: 1, Conf: 1, Fresh: true, AreaType: areaUnknown} // LR<1 压 fall
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
