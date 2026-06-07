package roomengine

import (
	"testing"

	"owl-common/alarm"
	"owl-common/observation"
	"wisefido-sensor/internal/roomengine/belief"
)

// belief_adapter_test.go — §9 第 2 步验收：用 CABB/John.Y/真跌倒 的代表性 struct 流验证
// adapter 产出的 Observation 序列符合预期（尤其 John.Y 长时间静止 → ObsPose Conf 转 0 的命门）。

func ptr(v int) *int { return &v }

func findObs(obs []belief.Observation, k belief.ObsKind) (belief.Observation, bool) {
	for _, o := range obs {
		if o.Kind == k {
			return o, true
		}
	}
	return belief.Observation{}, false
}

// 命门：长时间静止(StillBox) track（StillBoxRunStart 超 120s）即便 LastObservedMs 仍新，pose/kinematics/vital
// 必须 Fresh=false（不更新）；ghost-ness 不受影响仍 Fresh。
func TestAdapterStillBoxTrackKillsMotionObs(t *testing.T) {
	now := int64(10_000_000)
	ts := &TrackState{
		LastObservedMs: now - 1_000,            // 1s 前刚收帧（静止期持续推同帧）
		StillBoxRunStart: now - beliefStillBoxStaleMs - 5_000, // 静止已超阈值
		LastZ:          30,
		Verdict:        VerdictReal,
	}
	tr := observation.Track{LogicID: "L1", Pose: observation.PoseStanding, PositionX: ptr(-390), PositionY: ptr(30), PositionZ: ptr(30)}
	obs := radarFrameAdapter(tr, ts, nil, now)

	pose, _ := findObs(obs, belief.ObsPose)
	if pose.Fresh {
		t.Fatalf("StillBox track ObsPose 应 Fresh=false（命门），得 Fresh=true")
	}
	tp, _ := findObs(obs, belief.ObsTrackPresent)
	if !tp.Fresh {
		t.Fatalf("ObsTrackPresent(ghost) 不该被静止影响，应 Fresh=true")
	}
}

// 非静止 track：pose 正常透传，Geom 由 grid 算。
func TestAdapterFreshTrackPassesPose(t *testing.T) {
	now := int64(10_000_000)
	ts := &TrackState{LastObservedMs: now - 500, StillBoxRunStart: 0, LastZ: 160, Verdict: VerdictReal}
	tr := observation.Track{LogicID: "L1", Pose: observation.PoseWalking, PositionX: ptr(100), PositionY: ptr(100), PositionZ: ptr(160), PoseConfidence: 80}
	obs := radarFrameAdapter(tr, ts, nil, now)
	pose, ok := findObs(obs, belief.ObsPose)
	if !ok || !pose.Fresh || pose.Value != float64(observation.PoseWalking) || pose.Conf != 0.8 {
		t.Fatalf("fresh walking pose 透传错: %+v", pose)
	}
}

// AreaType → Geom 映射（纯函数，不需 grid）。
func TestGeomFromArea(t *testing.T) {
	cases := map[AreaType]belief.Geom{
		AreaBed:    belief.GeomInBed,
		AreaEnter:  belief.GeomInEnter,
		AreaToilet: belief.GeomInToilet,
		AreaShower: belief.GeomInToilet,
		AreaActive: belief.GeomOpenFloor,
		AreaSit:    belief.GeomOpenFloor,
	}
	for a, want := range cases {
		if got := geomFromArea(a); got != want {
			t.Fatalf("geomFromArea(%v)=%v want %v", a, got, want)
		}
	}
}

// 真跌倒：adapter → belief 端到端。开阔地板行走 → z 骤降 pose=fallen + firmware Fall 事件。
func TestAdapterGenuineFallFires(t *testing.T) {
	be := belief.New(belief.DefaultModel())
	now := int64(1_000)
	be.Step(now, []belief.Observation{{Kind: belief.ObsEnterExit, Value: 1, Conf: 0.9, Ts: now, Fresh: true, Geom: belief.GeomInEnter}})

	z := 165
	for i := 0; i < 5; i++ {
		now += 1000
		ts := &TrackState{LastObservedMs: now, LastZ: z, Verdict: VerdictReal}
		tr := observation.Track{LogicID: "L1", Pose: observation.PoseWalking, PositionX: ptr(50), PositionY: ptr(50), PositionZ: ptr(z), PoseConfidence: 80}
		be.Step(now, radarFrameAdapter(tr, ts, nil, now))
	}
	// 跌倒帧：pose=fallen + firmware fall。P2.1 删 ObsKinematics 后,真摔靠正向 pose 持续累积
	// (≥2 帧 pose-fallen ~2s);真实摔倒本就持续躺地,非单帧 z↓ 冲击(延迟 +~1s,委员会确认中)。
	now += 1000
	ts := &TrackState{LastObservedMs: now, LastZ: z, Verdict: VerdictReal}
	tr := observation.Track{LogicID: "L1", Pose: observation.PoseFallen, PositionX: ptr(50), PositionY: ptr(50), PositionZ: ptr(20), PoseConfidence: 80}
	frame := radarFrameAdapter(tr, ts, nil, now)
	fallEvt, _ := radarEventToObs(alarm.Fall, now, belief.GeomOpenFloor)
	be.Step(now, append(frame, fallEvt))

	now += 1000
	ts2 := &TrackState{LastObservedMs: now, LastZ: 20, Verdict: VerdictReal}
	tr2 := observation.Track{LogicID: "L1", Pose: observation.PoseFallen, PositionX: ptr(50), PositionY: ptr(50), PositionZ: ptr(20), PoseConfidence: 80}
	be.Step(now, radarFrameAdapter(tr2, ts2, nil, now))

	if d := be.Decide(); d != belief.DecisionFall {
		t.Fatalf("adapter 真跌倒未 fire: decision=%v b=%v", d, be.Vector())
	}
}

// John.Y 9h：adapter 长时间静止 → ObsPose Conf 0（命门）+ sleepad InBed 弱耦合 → 永不误报。
func TestAdapterJohnY9hNoFalseFire(t *testing.T) {
	be := belief.New(belief.DefaultModel())
	now := int64(1_000)
	be.Step(now, []belief.Observation{{Kind: belief.ObsEnterExit, Value: 1, Conf: 0.9, Ts: now, Fresh: true, Geom: belief.GeomInEnter}})
	// 走向床区边界
	for i := 0; i < 4; i++ {
		now += 1000
		ts := &TrackState{LastObservedMs: now, LastZ: 160, Verdict: VerdictReal}
		tr := observation.Track{LogicID: "L1", Pose: observation.PoseWalking, PositionX: ptr(-380), PositionY: ptr(30), PositionZ: ptr(160), PoseConfidence: 80}
		be.Step(now, radarFrameAdapter(tr, ts, nil, now))
	}
	stillStart := now
	// 长时间静止 9h 量级：D523 持续推同帧（LastObservedMs 一直新）但 StillBoxRunStart 不动 → adapter 判 stale。
	// 新设计：lost-fall 由"走动中消失"事件触发，非 frozen 帧。此处 D523 持续推帧（未消失）→ 不发 lost-fall；
	// 但 sleepad 在 09E7 床区报 InBed → ObsNeighbor 对冲压低 P(Fallen)。这是与 cabb-fall-A
	// （同样 open-floor 静止但**无** neighbor）的唯一区别：有对冲不 fire，无对冲 fire。
	for i := 0; i < 60; i++ {
		now += 540_000
		ts := &TrackState{LastObservedMs: now - 1_000, StillBoxRunStart: stillStart, LastZ: 30, Verdict: VerdictReal}
		tr := observation.Track{LogicID: "L1", Pose: observation.PoseStanding, PositionX: ptr(-390), PositionY: ptr(30), PositionZ: ptr(30)}
		frame := radarFrameAdapter(tr, ts, nil, now)
		neighbor := neighborToObs(0.9, 0.85, now) // sleepad 在 09E7 床区报 InBed
		be.Step(now, append(frame, neighbor))
	}
	if d := be.Decide(); d == belief.DecisionFall {
		t.Fatalf("John.Y 9h 经 adapter 仍误报（neighbor 对冲未压住 lost-still）: b=%v", be.Vector())
	}
	if be.Vector().P(belief.SFallen) > thFireProbe {
		t.Fatalf("John.Y P(Fallen)=%.3f 应低于 θ_fire", be.Vector().P(belief.SFallen))
	}
}

// thFireProbe 镜像 belief.thFire（包内未导出），仅测试断言用。
const thFireProbe = 0.55

// 线上真机 FP（2026-06-01 MoM/Bathroom/4D8710F41797,alarm 092279f3）：lost_track,
// context=track_lost_no_exit_room_no_recovery,**still_box_start_ms=0**(track 干净消失,无静止 box),
// cell_area_type=0。判别关键：人走出 FOV/雷达丢目标 = track 消失(无 still-box)≠ 倒地后躺住
// 新设计：lost-fall 走动前置(消失前60s在走动)。MoM 走动后消失，但有 ExitRoom → 取消。无走动消失=Still-fall 域。此处
// lost-still + 帧停→stale→A 漂离 Fallen → 不 fire（gate-list 却凭"丢 track+5min"误报）。
func TestMoMLostTrackVanishNoFire(t *testing.T) {
	be := belief.New(belief.DefaultModel())
	now := int64(1_000)
	be.Step(now, []belief.Observation{{Kind: belief.ObsEnterExit, Value: 1, Conf: 0.9, Ts: now, Fresh: true, Geom: belief.GeomInEnter}})
	// 浴室内行走几帧
	for i := 0; i < 5; i++ {
		now += 1000
		ts := &TrackState{LastObservedMs: now, StillBoxRunStart: 0, LastZ: 160, Verdict: VerdictReal}
		tr := observation.Track{LogicID: "L", Pose: observation.PoseWalking, PositionX: ptr(20), PositionY: ptr(130), PositionZ: ptr(160), PoseConfidence: 80}
		be.Step(now, radarFrameAdapter(tr, ts, nil, now))
	}
	// track 消失：此后无帧（adapter 不被调用），仅时间推进 5min（=gate-list 误报窗）
	for i := 0; i < 30; i++ {
		now += 10_000
		be.Step(now, nil)
	}
	if d := be.Decide(); d == belief.DecisionFall {
		t.Fatalf("MoM lost_track(vanish,无 still-box) 经 belief 仍误报: b=%v", be.Vector())
	}
	if be.Vector().P(belief.SFallen) > 0.05 {
		t.Fatalf("MoM P(Fallen)=%.3f 应极低（无 still-box 不该升 Fallen）", be.Vector().P(belief.SFallen))
	}
}
