package belief

import "testing"

// track_test.go — DBN P1 Track 层 A_T 结构化路由验证（合成）。
// 证明本会话三类 lost_track gate 在 T 层是转移/发射的自然结果，非外部 if：
//   - ghost 闪灭 → None（method-2：Ghost→Lost≈0）
//   - 门区消失 → JustLeft（门区 exit 推断）
//   - 真人开阔地板消失 → Lost（候选倒地，喂 Room 层 Fallen 前置）

func present(g float64, geom Geom, ts int64) TObservation {
	return TObservation{Kind: TObsPresent, Ghostness: g, Geom: geom, Conf: 0.9, Ts: ts, Fresh: true}
}
func absent(geom Geom, ts int64) TObservation {
	return TObservation{Kind: TObsAbsent, Geom: geom, Conf: 0.9, Ts: ts, Fresh: true}
}

// 喂 n 帧 present（每帧 +1s），再喂 m 帧 absent（每帧 +1s），返回末态。
func runTrack(t *testing.T, presentN int, g float64, presentGeom, absentGeom Geom, absentM int) TVector {
	tb := NewTrackBelief()
	ts := int64(1000)
	for i := 0; i < presentN; i++ {
		tb.Step(ts, []TObservation{present(g, presentGeom, ts)})
		ts += 1000
	}
	for i := 0; i < absentM; i++ {
		tb.Step(ts, []TObservation{absent(absentGeom, ts)})
		ts += 1000
	}
	return tb.Vector()
}

func TestTrackPresentRealVsGhost(t *testing.T) {
	real := runTrack(t, 10, 0.0, GeomOpenFloor, GeomOpenFloor, 0)
	if s, _ := real.Max(); s != TReal {
		t.Errorf("纯真人 present 应收敛 TReal，得 %s (%v)", s, real)
	}
	ghost := runTrack(t, 10, 1.0, GeomOpenFloor, GeomOpenFloor, 0)
	if s, _ := ghost.Max(); s != TGhost {
		t.Errorf("纯 ghost present 应收敛 TGhost，得 %s (%v)", s, ghost)
	}
}

func TestTrackRealLostOpenFloor(t *testing.T) {
	// 真人在开阔地板 present 后突然消失 → TLost 主导（候选倒地）。
	v := runTrack(t, 10, 0.0, GeomOpenFloor, GeomOpenFloor, 8)
	if s, p := v.Max(); s != TLost || p < 0.5 {
		t.Errorf("真人开阔地板消失应 TLost 主导，得 %s p=%.3f (%v)", s, p, v)
	}
}

func TestTrackPeerLiveSuppressesLost(t *testing.T) {
	// 真人开阔地板 present 后消失：若同房别台此刻仍见真人（TObsPeerLive）→ 本 track 是重影/重复，
	// TLost 被压住（对照 TestTrackRealLostOpenFloor 无 peer 时 TLost 主导）。
	tb := NewTrackBelief()
	ts := int64(1000)
	for i := 0; i < 10; i++ {
		tb.Step(ts, []TObservation{present(0.0, GeomOpenFloor, ts)})
		ts += 1000
	}
	for i := 0; i < 8; i++ {
		tb.Step(ts, []TObservation{
			absent(GeomOpenFloor, ts),
			{Kind: TObsPeerLive, Conf: 0.9, Ts: ts, Fresh: true},
		})
		ts += 1000
	}
	v := tb.Vector()
	if v.P(TLost) > 0.15 {
		t.Errorf("同房别台见真人 → 本 track 应被压住不 Lost，得 P(Lost)=%.3f (%v)", v.P(TLost), v)
	}
}

func TestTrackGhostVanishNotLost(t *testing.T) {
	// 结构核心：ghost present 后消失 → None，绝不 Lost（method-2 在 A_T 内）。
	v := runTrack(t, 10, 1.0, GeomOpenFloor, GeomOpenFloor, 8)
	if v.P(TLost) > 0.15 {
		t.Errorf("ghost 闪灭不得变 Lost，得 P(Lost)=%.3f (%v)", v.P(TLost), v)
	}
	if s, _ := v.Max(); s != TNone {
		t.Errorf("ghost 闪灭应收敛 None，得 %s (%v)", s, v)
	}
}

func TestTrackDoorExitNotLost(t *testing.T) {
	// 真人在门区消失 → JustLeft/None，不 Lost（门区 exit 推断在 absent 发射 geom 条件内）。
	v := runTrack(t, 10, 0.0, GeomInEnter, GeomInEnter, 8)
	if v.P(TLost) > 0.15 {
		t.Errorf("门区消失不得变 Lost，得 P(Lost)=%.3f (%v)", v.P(TLost), v)
	}
	if s, _ := v.Max(); s != TNone && s != TJustLeft {
		t.Errorf("门区消失应 None/JustLeft，得 %s (%v)", s, v)
	}
}

func TestTrackExitEventCancelsLost(t *testing.T) {
	// 真人开阔地板 present → 先喂 ExitRoom 事件 → 再消失：firmware 明确离场 → 不 Lost。
	tb := NewTrackBelief()
	ts := int64(1000)
	for i := 0; i < 10; i++ {
		tb.Step(ts, []TObservation{present(0.0, GeomOpenFloor, ts)})
		ts += 1000
	}
	tb.Step(ts, []TObservation{{Kind: TObsExit, Conf: 0.9, Ts: ts, Fresh: true}})
	ts += 1000
	for i := 0; i < 8; i++ {
		tb.Step(ts, []TObservation{absent(GeomOpenFloor, ts)})
		ts += 1000
	}
	v := tb.Vector()
	if v.P(TLost) > 0.2 {
		t.Errorf("ExitRoom 后消失不得 Lost，得 P(Lost)=%.3f (%v)", v.P(TLost), v)
	}
}

func TestTrackBedVanishNotLost(t *testing.T) {
	// 床区消失（躺下被遮挡）→ 不 Lost。
	v := runTrack(t, 10, 0.0, GeomInBed, GeomInBed, 8)
	if v.P(TLost) > 0.15 {
		t.Errorf("床区消失不得 Lost，得 P(Lost)=%.3f (%v)", v.P(TLost), v)
	}
}
