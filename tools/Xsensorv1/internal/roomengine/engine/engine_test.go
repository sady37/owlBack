package engine

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// engine_test.go — W3.2 单房 roomengine 端到端验收（DBN-wire-roadmap §6，C 复审 gate②）：
//   EG1 cd2b 单房零回归：合成"在床 InBed → 在线 LeftBed + 雷达仍躺静止近床"经 Room.Tick →
//        Ψ 相容涌现 SFallen（belief 单元的床边真摔结果在真 engine 主循环复现，不靠 fixture）。

func bedRect(x1, y1, x2, y2 int) adapter.Rect { return adapter.Rect{X1: x1, Y1: y1, X2: x2, Y2: y2} }

func TestEngineCd2bSingleRoomReproduces(t *testing.T) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	geomFI := adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}
	r := NewRoom(adapter.BedGeoms(geomFI), 1)

	mk := func(now int64, reading belief.BedReading, x int) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs:    now,
			Track:    adapter.RadarTrack{Online: true, Pose: 6, X: x, Y: 100, StillSec: 120},
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
		}
	}

	// 阶段A：在床 InBed 充 occ。
	for step := 1; step <= 30; step++ {
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50), 0, 1)
	}
	// 阶段B：在线 LeftBed + 雷达仍躺静止近床（cd2b 床边真摔，§32 二态不离线）。
	var fired bool
	for step := 31; step <= 150; step++ {
		fr := r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130), 0, 1)
		if fr.Decision.Fire {
			fired = true
		}
	}

	pF := r.MarginalS()[belief.SFallen]
	if pF < 0.5 {
		t.Fatalf("cd2b 经 Room.Tick P(SFallen)=%.4f 应涌现（Ψ 相容，belief 单元结果复现）", pF)
	}
	if !fired {
		t.Errorf("cd2b 经 Room.Tick 应 fire（床边真摔）")
	}
	t.Logf("EG1 ✓ cd2b 单房零回归：Room.Tick 端到端 P(SFallen)=%.4f fire=%v（belief 涌现在真 engine 复现）", pF, fired)
}
