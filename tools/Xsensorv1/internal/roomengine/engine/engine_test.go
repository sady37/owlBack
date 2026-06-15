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
			Tracks:   []adapter.TrackObs{{X: x, Y: 100}}, // 单住户 → N_r=1
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
		}
	}

	// 阶段A：在床 InBed 充 occ（雷达 X=50 床心=出生位，未位移）。
	for step := 1; step <= 30; step++ {
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50), 0)
	}
	// 阶段B：在线 LeftBed + 雷达仍躺静止近床（cd2b 床边真摔，§32 二态不离线）。
	// X=130 偏离出生位 80cm > MoveCm → Displaced 锁 Real → pFallReal≈1 不抑制 SFallen（RV4 端到端守住）。
	var fired bool
	for step := 31; step <= 150; step++ {
		fr := r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130), 0)
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

// EG2（§56 候选① pillar C）：N_r→PeopleCount 端到端排 ghost。真人 + 运动伪迹 ghost 经 Room.Tick →
//
//	census 排伪迹 → decide.PeopleCount=1（独处），不被影子虚增成 2 人 → 不误折扣 C_FN。
func TestEnginePillarCGhostExcludedToDecide(t *testing.T) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	r := NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)

	var dec belief.Decision
	for step := 1; step <= 12; step++ {
		fi := adapter.FrameInput{
			NowMs: int64(step) * 1000,
			Track: adapter.RadarTrack{Online: true, Pose: 6, X: 60 + step*5, Y: 100, StillSec: 120},
			Tracks: []adapter.TrackObs{
				{X: 60 + step*5, Y: 100},   // 真人正常走（speed≈5cm/帧）
				{X: 60 + step*200, Y: 400}, // 运动伪迹：持续超速 200cm/帧 > SpeedCeil=100 → aScore 累积成 ghost
			},
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: belief.BedLeftBed}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
		dec = r.Tick(fi, 0).Decision
	}
	if dec.PeopleCount != 1 {
		t.Errorf("真人+伪迹 ghost 应 N_r=1 到 decide（排 ghost 防虚增），得 PeopleCount=%d", dec.PeopleCount)
	}
	t.Logf("EG2 ✓ pillar C：census 排伪迹 ghost → decide.PeopleCount=%d（独处真人不被影子降级）", dec.PeopleCount)
}

// EG3（§56 候选① pillar D）：拍法 A 守门——N_r=2 + 一人 ≥55% 真摔 → 必报，C_FN 不参与（≥55% 证据自足）。
//
//	两条真人 track 喂 census（N_r=2），belief 接触轴 cd2b 床边摔涌现 P^F≥0.55 → 即使多人折扣也照报。
func TestEnginePillarDHoldsAtReportBand(t *testing.T) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	r := NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)

	mk := func(now int64, reading belief.BedReading, x int) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs: now,
			Track: adapter.RadarTrack{Online: true, Pose: 6, X: x, Y: 100, StillSec: 120},
			Tracks: []adapter.TrackObs{ // 两条独立真人（异步：甲随 x 动、乙固定不动）→ N_r=2
				{X: x, Y: 100},
				{X: 600, Y: 400},
			},
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
	}
	for step := 1; step <= 30; step++ {
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50), 0)
	}
	var dec belief.Decision
	var fired bool
	for step := 31; step <= 150; step++ {
		dec = r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130), 0).Decision
		if dec.Fire {
			fired = true
		}
	}
	if dec.PeopleCount != 2 {
		t.Fatalf("两独立真人应 N_r=2 到 decide，得 %d", dec.PeopleCount)
	}
	if dec.PFallen < 0.55 { // report 档阈（belief.pFireHi，跨包不可引，锚同值）
		t.Fatalf("接触轴 cd2b 床边摔应 P^F≥0.55（report 档），得 %.4f", dec.PFallen)
	}
	if dec.Band != "report" || !fired {
		t.Errorf("拍法 A 守门：N_r=2 + ≥55%% 真摔必报（band 应 report、Fire），得 band=%s fire=%v —— 多人折扣绝不折掉证据自足档", dec.Band, fired)
	}
	t.Logf("EG3 ✓ pillar D：N_r=%d + P^F=%.4f → band=%s fire=%v（≥55%% 证据自足，C_FN 不参与折扣）", dec.PeopleCount, dec.PFallen, dec.Band, fired)
}
