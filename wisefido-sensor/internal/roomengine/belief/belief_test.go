package belief

import (
	"testing"

	"owl-common/observation"
)

// belief_test.go — P5 全空间区域占用重写后的骨架 oracle（工单3 前半段）。
// 断言走**机制方向性**（区域路由 / blind ramp / 离场反证），不赌首版 A/L 标定的绝对 fire 阈值——
// 那（已知 TP/FP fixture 过确认窗）归工单3 后半段 + 工单4 cd2b oracle。

func ob(ts int64, kind ObsKind, val, conf float64, area int) Observation {
	return Observation{Kind: kind, Value: val, Conf: conf, Ts: ts, Fresh: true, AreaType: area}
}

func poseObs(ts int64, pose, area int, released bool) Observation {
	o := ob(ts, ObsPose, float64(pose), 0.8, area)
	o.BedReleased = released
	return o
}

// 状态空间形状：9 态，Fallen 仍 index 5（Decider 引用稳定），label 不变。
func TestStateSpaceShape(t *testing.T) {
	if numStates != 9 {
		t.Fatalf("numStates 应 9，得 %d", numStates)
	}
	if SFallen != 5 {
		t.Fatalf("SFallen 应 index 5（Decider 引用稳定），得 %d", SFallen)
	}
	if stateLabel[SFallen] != "Floor-Fallen" {
		t.Fatalf("SFallen label 应 Floor-Fallen，得 %q", stateLabel[SFallen])
	}
}

// 床上躺（bed_state 未离床）= 睡，主导 SBed，不报 Fallen。
func TestLyingInBedStaysBed(t *testing.T) {
	be := New(DefaultModel())
	now := int64(1_000)
	for i := 0; i < 30; i++ {
		now += 1000
		be.Step(now, []Observation{poseObs(now, observation.PoseLying, areaBed, false)})
	}
	if s, _ := be.Vector().Max(); s != SBed {
		t.Fatalf("床上躺未离床应主导 SBed，得 %s", s)
	}
	if p := be.Vector().P(SFallen); p > 0.2 {
		t.Fatalf("床上躺 P(Fallen) 应低，得 %.3f", p)
	}
}

// cd2b 出口（belief 层）：床区躺 + bed_state 离床（BedReleased）→ 倒地候选，P(Fallen) 抬过 P(Bed)。
func TestBedReleasedRoutesFallen(t *testing.T) {
	be := New(DefaultModel())
	now := int64(1_000)
	for i := 0; i < 30; i++ {
		now += 1000
		be.Step(now, []Observation{poseObs(now, observation.PoseLying, areaBed, true)})
	}
	if f, b := be.Vector().P(SFallen), be.Vector().P(SBed); f <= b {
		t.Fatalf("床区离床躺应 P(Fallen)>P(Bed)，得 Fallen=%.3f Bed=%.3f", f, b)
	}
}

// lost-fall 核心机制：占用（走动）后丢轨 → 重复 ObsNoDetect → 信念经 Blind 漂向 Fallen（占用+不可见+未离场）。
func TestLostFallRampsFromOccupied(t *testing.T) {
	be := New(DefaultModel())
	now := int64(1_000)
	now += 1000
	be.Step(now, []Observation{{Kind: ObsEnterExit, Value: +1, Conf: 0.9, Ts: now, Fresh: true}})
	for i := 0; i < 10; i++ { // 开阔地走动，track 解析到
		now += 1000
		be.Step(now, []Observation{poseObs(now, observation.PoseWalking, areaActive, false)})
	}
	pEarly := be.Vector().P(SFallen)
	for i := 0; i < 90; i++ { // track 丢失：每 tick ObsNoDetect（realnessP=1 真人 / doorExit=0 非门区）
		now += 1000
		be.Step(now, []Observation{{Kind: ObsNoDetect, Conf: 0.8, Ts: now, Fresh: true, RealnessP: 1.0, DoorExitP: 0.0}})
	}
	pLate := be.Vector().P(SFallen)
	if pLate <= pEarly {
		t.Fatalf("丢轨久缺 P(Fallen) 应 ramp 上升，得 early=%.3f late=%.3f", pEarly, pLate)
	}
	switch s, _ := be.Vector().Max(); s {
	case SFallen, SBlindOpen, SBlindRest: // 占用但不可见/倒地 —— 非可见区、非 Empty
	default:
		t.Fatalf("占用后丢轨久缺，argmax 应在 Fallen/Blind*，得 %s", s)
	}
}

// 离场反证（方向）：丢轨 ramp 中收到 ExitRoom → 信念背离 Fallen、转向 Left（lrExitLeft=8 是 Blind* 唯一离场
// 出口，A 无逃生阀；exit 同时压盲）。不验"一帧翻盘"——near-absorbing Fallen 不应被单个 ExitRoom 瞬间抹掉
// （安全：可能第二人离场/误事件），cancel 靠持续证据，故只验方向。
func TestExitRoomCancelsLostFall(t *testing.T) {
	be := New(DefaultModel())
	now := int64(1_000)
	now += 1000
	be.Step(now, []Observation{{Kind: ObsEnterExit, Value: +1, Conf: 0.9, Ts: now, Fresh: true}})
	for i := 0; i < 10; i++ {
		now += 1000
		be.Step(now, []Observation{poseObs(now, observation.PoseWalking, areaActive, false)})
	}
	for i := 0; i < 20; i++ {
		now += 1000
		be.Step(now, []Observation{{Kind: ObsNoDetect, Conf: 0.8, Ts: now, Fresh: true, RealnessP: 1.0}})
	}
	fallenBefore, leftBefore := be.Vector().P(SFallen), be.Vector().P(SLeft)
	now += 1000
	be.Step(now, []Observation{{Kind: ObsEnterExit, Value: -1, Conf: 0.9, Ts: now, Fresh: true}})
	fallenAfter, leftAfter := be.Vector().P(SFallen), be.Vector().P(SLeft)
	if fallenAfter >= fallenBefore {
		t.Fatalf("ExitRoom 应压 Fallen，得 before=%.3f after=%.3f", fallenBefore, fallenAfter)
	}
	if leftAfter <= leftBefore {
		t.Fatalf("ExitRoom 应抬 Left，得 before=%.3f after=%.3f", leftBefore, leftAfter)
	}
}
