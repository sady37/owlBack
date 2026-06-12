package belief

import (
	"testing"

	"owl-common/observation"
)

// R5-calibration-lock(审查53 裁 B-纠正版)——机械化锁 SFallen 似然符号,按各源 R5 角色分类,
// 防 calibration.go 常量被误调成破 R5(用非-pose/z 之外的源压 fall,或 pose/z 反向压 fall)。
//
// **不是"全 lr*Fallen 中性锁"**(那会误红合法压制源 → 破 DBN filter)。正确不变量按源角色:
//
//	┌── pose/z(R5 铁律核心) ── ObsPose*(全 pose×全 geom)/ObsZBand → SFallen-LR ≥ 1(只正向,永不 <1)
//	├── 中性                ── ObsNumberPeople                      → SFallen-LR = 1.0
//	├── 抬升                ── ObsDwellStill/ObsNoDetect → SFallen-LR ≥ 1
//	└── 可靠压制源(豁免,设计许可 <1) ── 见 permittedFallSuppressors:各源 source 依据 + 本测试
//	                            正向断言"满证据时确 <1"(证其仍是 DBN 合法压制通道,被误中性化也红)
//
// R5 字面(feedback 原则#1):"用 pose/z 抑制/否决跌倒 = 违规(制造 FN)"。核心锁 = pose/z 的 SFallen 永不 <1。

const r5Eps = 1e-9

// allAreas — 枚举全部 area_type,确保 pose/z 锁覆盖每种位置条件(Bed 降权后仍须 ≥1)。
var allAreas = []int{areaUnknown, areaBed, areaEnter, areaActive, areaToilet}

// TestR5LockPoseZNeverSuppressFall — pose/z 是 R5 铁律核心:任何 pose 值 × 任何 geom,SFallen-LR 永不 <1
// (lrPoseFallenInBed=1.5 降权仍 >1;无 SFallen 项的 pose=1.0 中性;ZBand 绝不写 SFallen)。
// 红 = 谁把某 pose/geom 或 z 档的 SFallen 调到 <1 = 用 pose/z 压 fall = R5 违规。
func TestR5LockPoseZNeverSuppressFall(t *testing.T) {
	// 枚举固件全 pose(0..12,含 Unknown/未来越界值 13..15 防御),× 全 geom。
	for pose := 0; pose <= 15; pose++ {
		for _, g := range allAreas {
			v := rawLikelihood(Observation{Kind: ObsPose, Value: float64(pose), Conf: 1, Fresh: true, AreaType: g})
			if v[SFallen] < 1.0-r5Eps {
				t.Fatalf("R5 核心违规:ObsPose pose=%d area=%s 的 SFallen-LR=%.4f <1 = 用 pose 压 fall(制造 FN)",
					pose, AreaTypeLabel(g), v[SFallen])
			}
		}
	}
	// ZBand 三档(低/坐/直立)× 全 geom:绝不写 SFallen → 恒 1.0(≥1)。
	for _, z := range []float64{10, 50, 100} {
		for _, g := range allAreas {
			v := rawLikelihood(Observation{Kind: ObsZBand, Value: z, Conf: 1, Fresh: true, AreaType: g})
			if v[SFallen] < 1.0-r5Eps {
				t.Fatalf("R5 核心违规:ObsZBand z=%.0f area=%s 的 SFallen-LR=%.4f <1 = 用 z 压 fall", z, AreaTypeLabel(g), v[SFallen])
			}
		}
	}
}

// TestR5LockNumberPeopleNeutral — ObsNumberPeople 对 fall 中性(np=0 corroboration 非 substitution;np>0 只压 Empty)。
// 红 = lrNp0Fallen 被调离 1.0 → np 进 SFallen 决策(㉝:np=0 压 fall 漏真摔 / np>0 抬 fall 无依据)。
func TestR5LockNumberPeopleNeutral(t *testing.T) {
	for _, np := range []float64{0, 1, 2, 5} {
		v := rawLikelihood(Observation{Kind: ObsNumberPeople, Value: np, Conf: 0.8, Fresh: true, AreaType: areaUnknown})
		if d := v[SFallen] - 1.0; d < -r5Eps || d > r5Eps {
			t.Fatalf("R5 违规:ObsNumberPeople np=%.0f 的 SFallen-LR=%.4f ≠1.0(中性) → np 入 fall 决策", np, v[SFallen])
		}
	}
}

// TestR5LockLiftSourcesNeverSuppressFall — 抬升类(dwell/no-detect)对 fall 只抬不压:SFallen-LR ≥ 1。
// 红 = 谁把抬升源的 SFallen 调 <1(把"久驻/消失"算成 fall 的反证,语义颠倒)。
func TestR5LockLiftSourcesNeverSuppressFall(t *testing.T) {
	cases := []Observation{
		{Kind: ObsDwellStill, Value: 600, Conf: 1, Fresh: true, AreaType: areaToilet},
		{Kind: ObsDwellStill, Value: 600, Conf: 1, Fresh: true, AreaType: areaActive, ToleranceFactor: 1.0},
		{Kind: ObsDwellStill, Value: 600, Conf: 1, Fresh: true, AreaType: areaBed}, // rest → lk(nil) → 1.0
		{Kind: ObsNoDetect, Value: 0, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 1, DoorExitP: 0},
		{Kind: ObsNoDetect, Value: 0, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 1, DoorExitP: 1}, // door-exit 仍留 floor ≥1
		{Kind: ObsNoDetect, Value: 0, Conf: 1, Fresh: true, AreaType: areaActive, RealnessP: 0, DoorExitP: 0}, // ghost 消失 → 中性 1.0
	}
	for _, o := range cases {
		v := rawLikelihood(o)
		if v[SFallen] < 1.0-r5Eps {
			t.Fatalf("R5 违规:抬升源 %s(area=%s,real=%.0f,dx=%.0f)的 SFallen-LR=%.4f <1(只抬不压)",
				o.Kind, AreaTypeLabel(o.AreaType), o.RealnessP, o.DoorExitP, v[SFallen])
		}
	}
}

// permittedFallSuppressor — 「许可压制清单」(审查53):**非 pose/z 的可靠证据**,按 DBN filter 设计本就该压 fall。
// 锁住"它们确实仍是合法压制通道"——满证据时 SFallen 必 <1;谁把某条误中性化(破 DBN)同样红。
// source 依据逐条标注(R5 原则#5"fall 压制只走 reliable:realness/rest-zone/sleepad/recapture/human-bed")。
type permittedFallSuppressor struct {
	name   string
	source string // 为何许可压 fall(R5 #5 reliable 通道依据)
	obs    Observation
}

var permittedFallSuppressors = []permittedFallSuppressor{
	{"BedOccupied", "接触式床占用概率(sleepad/human-bed,P7.4 豁免;非 pose/z)", Observation{Kind: ObsBedOccupied, Value: 1, Conf: 1, Fresh: true, AreaType: areaBed}},
	{"TrackPresent-Ghost", "realness/ghost-ness 运动学(R5 #4:镜面/反射伪迹非真倒地)", Observation{Kind: ObsTrackPresent, Value: 1, Conf: 1, Fresh: true, AreaType: areaActive}},
	{"Neighbor", "§5.5.2 弱耦合:邻房占用→本房更可能空/离(非 pose/z)", Observation{Kind: ObsNeighbor, Value: 1, Conf: 1, Fresh: true, AreaType: areaUnknown}},
	{"ReachableExit", "近门单帧可达→从门走出(Left),替 30cm 硬门闸(非 pose/z)", Observation{Kind: ObsReachableExit, Value: 1, Conf: 1, Fresh: true, AreaType: areaEnter}},
	{"EnterExit-Exit", "ExitRoom 事件正向退场(R5 #3 event present;事件非 pose/z)", Observation{Kind: ObsEnterExit, Value: -1, Conf: 1, Fresh: true, AreaType: areaEnter}},
}

// TestR5LockPermittedSuppressionRegistry — 「许可压制清单」机械化:每条满证据时 SFallen 必 <1
// (证其仍是 DBN 合法压制通道)。红 = 某可靠压制源被误中性化(damp→0 / lrExitFallen→1)→ 破 DBN filter
// (可靠退场/床占用/ghost 不再压 fall → FP 回潮)。与上三测试合成 R5 总闸:压制只许走本清单,pose/z/np/抬升永不压。
func TestR5LockPermittedSuppressionRegistry(t *testing.T) {
	for _, p := range permittedFallSuppressors {
		v := rawLikelihood(p.obs)
		if v[SFallen] >= 1.0-r5Eps {
			t.Fatalf("许可压制源 %s 应压 fall(SFallen-LR<1)却得 %.4f → 被误中性化,破 DBN filter(依据:%s)",
				p.name, v[SFallen], p.source)
		}
	}
}

// TestR5LockNeighborCorroborationNotSubstitution — ObsNeighbor 只在邻房**真占用**(Value>0)时压 fall;
// Value=0(邻房无信号)→ SFallen-LR=1.0 不压。锁"corroboration≠substitution"(wire 前置#4,同 np=0/NoDetect realness):
// 邻房**缺信号 ≠ 人在别处**,裸 absence 不得压本房 fall(否则邻房静默时压掉本房真摔=漏报)。
// (gate 层"多resident→不发 ObsNeighbor"的漏报方向锁在 roomengine/belief_neighbor_test.go ⑤——
//  census 在 roomengine 包,belief 包看不到,故 gate 锁落 roomengine,似然锁落此。)
func TestR5LockNeighborCorroborationNotSubstitution(t *testing.T) {
	v0 := rawLikelihood(Observation{Kind: ObsNeighbor, Value: 0, Conf: 1, Fresh: true, AreaType: areaUnknown})
	if d := v0[SFallen] - 1.0; d < -r5Eps || d > r5Eps {
		t.Fatalf("R5 违规:ObsNeighbor occ=0 应不压 fall(SFallen-LR=1.0)却得 %.4f → 裸 absence 压 fall=漏报", v0[SFallen])
	}
	v1 := rawLikelihood(Observation{Kind: ObsNeighbor, Value: 1, Conf: 1, Fresh: true, AreaType: areaUnknown})
	if v1[SFallen] >= 1.0-r5Eps {
		t.Fatalf("R5 违规:ObsNeighbor occ=1 应压 fall(SFallen-LR<1)却得 %.4f → 真占用证据被误中性化", v1[SFallen])
	}
}

// 文档化:ObsVitalPresent 对 SFallen 不写(恒 1.0 中性)——既非抬升亦非压制,
// 走自己态(Empty/Artifact/Bed*/prior),不入 fall 通道。无需单测(上三类已覆盖"非清单源不得 <1"由 pose/z/lift/np 守;
// 这三源本就不碰 SFallen,留此注记备 source-fidelity 审计对账)。
var _ = observation.PoseFallen // 锚定 observation 依赖(pose 枚举随固件演进时编译期提示)
