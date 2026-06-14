package roomengine

import (
	"testing"

	"owl-common/card"
	"owl-common/observation"
	"owl-common/radarutils"
	"wisefido-sensor/internal/roomengine/belief"
)

// P5(审查㊾ 重裁)bed 占用概率压 radar fall。retire bedLeakState/bedAuthorityObs(审查㊴ P5c 的 radar-on-bed
// leg 是祸首:0127 实证 radar 误读 on-bed 人为 Floor-Fallen 时门控开→P5 不 engage)。改 wire 既有 bed 贝叶斯
// Markov(bedAdapter→ObsBedOccupied,占用概率压 SFallen,**无 radar-on-bed 要求**;R5-clean 接触占用非 pose/z;
// any-source-OR LeftBed → BedOccupancyState 占用降 → 释放 → Fall 浮出,漏报-safe)。

func p5ob(ts int64, kind belief.ObsKind, val, conf float64, area int) belief.Observation {
	return belief.Observation{Kind: kind, Value: val, Conf: conf, Ts: ts, Fresh: true, AreaType: area}
}

// TestP5BedOccupancySuppressesRadarFloor — 治 0127:sleepad InBed 占用概率压 radar 误读 Floor-Fallen,
// **无 radar-on-bed 要求**(radar geom=OpenFloor 仍压)。验有床 P(Fallen) ≪ 无床基线。
func TestP5BedOccupancySuppressesRadarFloor(t *testing.T) {
	build := func(withBed bool) *belief.Belief {
		be := belief.New(belief.DefaultModel())
		tt := int64(1000)
		for i := 0; i < 8; i++ {
			tt += 1000
			o := []belief.Observation{p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, int(AreaActive))} // radar 把床上人误读 Floor-Fallen
			if withBed {
				o = append(o, bedAdapter(card.BedState{BedStatus: 0, BedConfidence: 90, BedStatusTs: tt}, tt)...) // InBed 占用概率
			}
			be.Step(tt, o)
		}
		return be
	}
	sup := build(true).Vector().P(belief.SFallen)
	base := build(false).Vector().P(belief.SFallen)
	if sup >= base {
		t.Fatalf("bed 占用概率应压 radar Floor-Fallen(无 radar-on-bed):有床 P(Fallen)=%.4f 应 < 无床 %.4f", sup, base)
	}
}

// TestP5LeftBedReleases — any-source-OR LeftBed → 占用降(bedVal=0)→ 不压 → Fall 浮出(R5 不漏滚下床真摔)。
func TestP5LeftBedReleases(t *testing.T) {
	be := belief.New(belief.DefaultModel())
	tt := int64(1000)
	fired := false
	for i := 0; i < 8 && !fired; i++ {
		tt += 1000
		o := []belief.Observation{p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, int(AreaActive))}
		o = append(o, bedAdapter(card.BedState{BedStatus: 1, BedConfidence: 90, BedStatusTs: tt}, tt)...) // LeftBed → bedVal=0 不压
		be.Step(tt, o)
		fired = be.Decide() == belief.DecisionFall
	}
	if !fired {
		t.Fatalf("LeftBed 后(滚下床真摔)绝不可被压 → 必确认 Fall(R5 红线);P(Fallen)=%.4f", be.Vector().P(belief.SFallen))
	}
}

// TestP5BedSuppressIgnoresZ — R5 专项断言:占用压制由占用概率驱动,与 z **无关**(占用概率非 pose/z)。
func TestP5BedSuppressIgnoresZ(t *testing.T) {
	run := func(z float64) float64 {
		be := belief.New(belief.DefaultModel())
		tt := int64(1000)
		for i := 0; i < 8; i++ {
			tt += 1000
			o := []belief.Observation{
				p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, int(AreaActive)),
				p5ob(tt, belief.ObsZBand, z, 0.7, int(AreaActive)),
			}
			o = append(o, bedAdapter(card.BedState{BedStatus: 0, BedConfidence: 90, BedStatusTs: tt}, tt)...)
			be.Step(tt, o)
		}
		return be.Vector().P(belief.SFallen)
	}
	zlow := run(10)   // z低(P5a 会当真摔放行);占用压制不看 z → 仍压
	zhigh := run(200) // z高
	if zlow > 0.5 || zhigh > 0.5 {
		t.Fatalf("R5:占用压制应 z 无关使两者 P(Fallen) 均低;zlow=%.4f zhigh=%.4f", zlow, zhigh)
	}
}

// TestP5BedOccupancyStateAnySourceOR — BedOccupancyState(审查㊾)any-source-OR LeftBed:任一源 LeftBed
// (sleepad ∨ radar)晚于最近 InBed → NotInBed(BedStatus=1)→ 释放压制(解 sleepad 滞后漏真摔)。
func TestP5BedOccupancyStateAnySourceOR(t *testing.T) {
	tm := NewTrackManager("r", NewRoomGrid(300, 300, radarutils.CellSize))
	// sleepad InBed@1000,无 LeftBed → 占用(BedStatus=0)
	tm.bedSessions["s1"] = &BedSession{DeviceUID: "s1", InBedSinceMs: 1000}
	if bs := tm.BedOccupancyState(2000); bs.BedStatus != 0 || bs.BedConfidence != bedConfSleepad {
		t.Fatalf("sleepad InBed → 占用(BedStatus=0,conf=%d),得 status=%d conf=%d", bedConfSleepad, bs.BedStatus, bs.BedConfidence)
	}
	// radar LeftBed@1500(晚于 InBed,sleepad 没报)→ any-source-OR veto → NotInBed
	tm.lastRadarLeftBedMs = 1500
	if bs := tm.BedOccupancyState(2000); bs.BedStatus != 1 {
		t.Fatalf("radar LeftBed(any-source-OR)晚于 InBed → NotInBed(BedStatus=1),得 %d", bs.BedStatus)
	}
	// 再 InBed@2000(任一源新 InBed 晚于 LeftBed)→ 回占用
	tm.lastRadarInBedMs = 2000
	if bs := tm.BedOccupancyState(3000); bs.BedStatus != 0 {
		t.Fatalf("最新 InBed 晚于 LeftBed → 回占用(BedStatus=0),得 %d", bs.BedStatus)
	}
}
