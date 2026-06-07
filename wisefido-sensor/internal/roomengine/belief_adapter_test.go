package roomengine

import (
	"testing"

	"owl-common/alarm"
	"owl-common/card"
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
	// 跌倒帧 + 持续躺地。断言**不变量**:持续 pose-fallen 在窗内必确认 —— 不锁帧数。
	// 删 Δz(P2.1)/ firmware 降权(P2.4)后单帧冲击退场,genuine-fall 由正向 pose 多帧累积;
	// latency 是 shadow-moot 量(belief Decide 不接 alarm;production fall 走 firmware Device_ALARM 独立路径)。
	now += 1000
	ts := &TrackState{LastObservedMs: now, LastZ: 20, Verdict: VerdictReal}
	tr := observation.Track{LogicID: "L1", Pose: observation.PoseFallen, PositionX: ptr(50), PositionY: ptr(50), PositionZ: ptr(20), PoseConfidence: 80}
	frame := radarFrameAdapter(tr, ts, nil, now)
	fallEvt, _ := radarEventToObs(alarm.Fall, now, belief.GeomOpenFloor)
	be.Step(now, append(frame, fallEvt))

	fired := be.Decide() == belief.DecisionFall
	for i := 0; i < 6 && !fired; i++ {
		now += 1000
		tsN := &TrackState{LastObservedMs: now, LastZ: 20, Verdict: VerdictReal}
		trN := observation.Track{LogicID: "L1", Pose: observation.PoseFallen, PositionX: ptr(50), PositionY: ptr(50), PositionZ: ptr(20), PoseConfidence: 80}
		be.Step(now, radarFrameAdapter(trN, tsN, nil, now))
		fired = be.Decide() == belief.DecisionFall
	}
	if !fired {
		t.Fatalf("adapter 真跌倒持续躺地窗内未确认: b=%v", be.Vector())
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

// P2.5 absence≠负向(原则#3):无近期 enter/exit 事件 → 不发 ObsEnterExit。
// 锁不变量:不得用"曾经离开/无 exit"持续喂反向证据(信号丢失≠状态);present 事件才发(正向)。
func TestAbsenceNotNegativeEnterExit(t *testing.T) {
	now := int64(10_000_000)
	// (a) 陈旧 exit(超事件窗)→ 不该再发(absence≠负向)。
	stale := card.RoomState{LastExitTs: now - beliefEventWindowMs - 1_000}
	if _, ok := findObs(roomAdapter(stale, 0, now), belief.ObsEnterExit); ok {
		t.Fatalf("陈旧 exit(超事件窗)不该再发 ObsEnterExit(absence≠负向)")
	}
	// (b) 无任何事件 → 不发。
	if _, ok := findObs(roomAdapter(card.RoomState{}, 0, now), belief.ObsEnterExit); ok {
		t.Fatalf("无 enter/exit 事件不该发 ObsEnterExit")
	}
	// (c) 新鲜 exit(事件窗内)→ 应发(present=正向证据,对照组)。
	fresh := card.RoomState{LastExitTs: now - 1_000}
	if _, ok := findObs(roomAdapter(fresh, 0, now), belief.ObsEnterExit); !ok {
		t.Fatalf("新鲜 exit(事件窗内)应发 ObsEnterExit(present 正向)")
	}
}

// P3.1 硬验收(委员会审查⑦):独立 shadow realness 判出**生产 Verdict 漏掉**的冻结 ghost(cd2b)。
// cd2b:跌床→雷达丢真人→track 跳椅子冻住(ghost),但生产 Verdict=Real 漏判。
// shadow 三探测器从 XY raw 算,**不看 Verdict** → 隐含速度超室内天花板 → 判 ghost(g=1)。
func TestShadowRealnessCatchesFrozenGhost(t *testing.T) {
	// cd2b 冻结 ghost:生产判 Real(漏),但隐含速度 150cm/s 超室内天花板 120。
	cd2b := &TrackState{Verdict: VerdictReal, MaxImpliedSpeedFromBirth: 150}
	if g := shadowTrackGhostness(cd2b, -1); g != 1 {
		t.Fatalf("cd2b 冻结 ghost(隐含速度150>120)shadow 应判 ghost(g=1),得 %v —— 生产 Verdict=Real 漏判正是 P3.1 要抓的", g)
	}
	// Kalman 残差急变(预测偏离 130cm>100)也判 ghost,即便 Verdict=Real。
	resid := &TrackState{Verdict: VerdictReal, MaxKalmanResidual: 130}
	if g := shadowTrackGhostness(resid, -1); g != 1 {
		t.Fatalf("Kalman 残差130>100 应判 ghost,得 %v", g)
	}
	// 本帧空间跳跃(200cm/s>120)判 ghost。
	if g := shadowTrackGhostness(&TrackState{Verdict: VerdictReal}, 200); g != 1 {
		t.Fatalf("本帧跳跃200>120 应判 ghost,得 %v", g)
	}
	// 真人正常 track:隐含速度40、残差20、帧速60,全在室内带内 → 不判 ghost(g=0)。
	real := &TrackState{Verdict: VerdictReal, MaxImpliedSpeedFromBirth: 40, MaxKalmanResidual: 20}
	if g := shadowTrackGhostness(real, 60); g != 0 {
		t.Fatalf("真人正常动学 不该判 ghost,得 %v", g)
	}
	// **独立性证明**:即便生产 Verdict=Ghost,shadow 仍只看 XY(此处全带内)→ g=0,不复用 Verdict。
	if g := shadowTrackGhostness(&TrackState{Verdict: VerdictGhost, MaxImpliedSpeedFromBirth: 40, MaxKalmanResidual: 20}, 60); g != 0 {
		t.Fatalf("shadow realness 必须独立于生产 Verdict(不 passthrough),得 %v", g)
	}
}
