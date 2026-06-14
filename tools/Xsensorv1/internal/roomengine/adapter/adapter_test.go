package adapter

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// adapter_test.go — 翻译层验收（集成步骤1）：
//   AD1 floorStrip 派生（床内→false / 床缘 margin 内→true / 远→false）
//   AD2 Gxy 派生（内→peak / 近→Near / 远→0 / 近两床→均匀）
//   AD3 BuildObservation 映射（pose=6/sleepad TTL/VitalSourceOnline/nearBed/HRRR）
//   AD4 端到端 mini cd2b 离线 replay（adapter→belief，XY 派生 floor-strip 复现 E5：P(Fallen) 浮出）

func bed(x1, y1, x2, y2 int) Rect { return Rect{x1, y1, x2, y2} }

// AD1：floorStrip 派生（down 姿门控 + 床外近缘）。床 (0,0,100,200)，margin=60。
func TestFloorStripDerivation(t *testing.T) {
	p := DefaultParams()
	mk := func(x, y, pose int) FrameInput {
		return FrameInput{Track: RadarTrack{Online: true, X: x, Y: y, Pose: pose}, Beds: []Rect{bed(0, 0, 100, 200)}}
	}
	if floorStrip(mk(50, 100, 6), p) {
		t.Error("床内(50,100) lying 不应是 floor-strip（on-pad）")
	}
	if !floorStrip(mk(130, 100, 6), p) {
		t.Error("床缘外 30cm(130,100,≤60 margin) lying 应是 floor-strip")
	}
	if floorStrip(mk(300, 100, 6), p) {
		t.Error("远(300,100,>60 margin) 不应是 floor-strip")
	}
	// down-pose 门控（AC-2 修，HR-5 误火根因）：走动(pose=1)在床缘 不是 floor-strip。
	if floorStrip(mk(130, 100, 1), p) {
		t.Error("床缘 walking(pose=1) 不应是 floor-strip（δ 仅对 down 姿）")
	}
	if floorStrip(FrameInput{Track: RadarTrack{Online: false, Pose: 6}, Beds: []Rect{bed(0, 0, 100, 200)}}, p) {
		t.Error("雷达离线不应是 floor-strip")
	}
}

// AD2：Gxy 派生。
func TestGxyDerivation(t *testing.T) {
	p := DefaultParams()
	beds := []Rect{bed(0, 0, 100, 200), bed(300, 0, 400, 200)} // 两床，间距 200cm
	// XY 在 bed0 内 → bed0 peak, bed1 远=0
	g := Gxy(FrameInput{Track: RadarTrack{Online: true, X: 50, Y: 100}, Beds: beds}, p)
	if g[0] != p.PeakInsideGxy || g[1] != 0 {
		t.Errorf("床内：g=%v 应 [peak,0]", g)
	}
	// XY 在两床中间各 margin 内（bed0 右缘 100、bed1 左缘 300；x=130 距 bed0=30、距 bed1=170>60）
	// 取 x=130 → 仅 bed0 near。要"近两床均匀"需更靠中且双 margin 重叠——此布局 margin=60 不重叠，
	// 改测：x=130 仅 bed0 Near。
	g2 := Gxy(FrameInput{Track: RadarTrack{Online: true, X: 130, Y: 100}, Beds: beds}, p)
	if g2[0] != p.NearGxy || g2[1] != 0 {
		t.Errorf("床0缘外30：g=%v 应 [Near,0]", g2)
	}
	// 雷达离线 → 全 0
	g3 := Gxy(FrameInput{Track: RadarTrack{Online: false}, Beds: beds}, p)
	if g3[0] != 0 || g3[1] != 0 {
		t.Errorf("离线 g=%v 应全 0", g3)
	}
}

// AD2b：近两床→均匀（紧邻布局，margin 重叠区）。
func TestGxyUniformBetweenBeds(t *testing.T) {
	p := DefaultParams()
	beds := []Rect{bed(0, 0, 100, 200), bed(150, 0, 250, 200)} // 间距 50cm < 2*margin
	g := Gxy(FrameInput{Track: RadarTrack{Online: true, X: 125, Y: 100}, Beds: beds}, p) // 缝中点，距两床各 25
	if g[0] != p.NearGxy || g[1] != p.NearGxy {
		t.Errorf("两床缝中(125,100)：g=%v 应均匀 [Near,Near]（§4 床间均匀）", g)
	}
}

// AD3：BuildObservation 映射。
func TestBuildObservation(t *testing.T) {
	p := DefaultParams()
	fi := FrameInput{
		Track:    RadarTrack{Online: true, Pose: 6, X: 130, Y: 100, HR: 0, RR: 0, StillSec: 80},
		Sleepads: []SleepadFrame{{InBed: false, Fresh: true}}, // 在线但离床
		Beds:     []Rect{bed(0, 0, 100, 200)},
	}
	o := BuildObservation(fi, p)
	if !o.PoseLying {
		t.Error("pose=6 应 PoseLying")
	}
	if o.Sleepad[0] != belief.BedLeftBed {
		t.Errorf("fresh ∧ ¬InBed 应 BedLeftBed，得 %v", o.Sleepad[0])
	}
	if !o.VitalSourceOnline {
		t.Error("sleepad fresh → VitalSourceOnline=true")
	}
	if !o.NearBed || !o.HRRRObserved {
		t.Error("床缘 30cm 应 NearBed + HRRRObserved（近床返 vital 通道）")
	}
	if o.HRRRPresent {
		t.Error("HR=RR=0 应 HRRRPresent=false")
	}
	if !o.FloorStripXY {
		t.Error("床缘外 30cm 应 FloorStripXY")
	}
	// sleepad 离线 → NoReport + VitalSourceOnline=false
	fi.Sleepads = []SleepadFrame{{InBed: true, Fresh: false}}
	o2 := BuildObservation(fi, p)
	if o2.Sleepad[0] != belief.BedNoReport || o2.VitalSourceOnline {
		t.Errorf("¬Fresh 应 NoReport + VitalSourceOnline=false，得 %v / %v", o2.Sleepad[0], o2.VitalSourceOnline)
	}
}

// AD4：端到端 mini cd2b 离线 replay。raw 帧(sleepad 离线 + pose lying + XY 床沿地条) 经 adapter→belief，
// P(Fallen) 浮出（复现 E5，但 floor-strip 由 XY 派生而非手设，证翻译层驱动正确）。
func TestAdapterCd2bOfflineReplay(t *testing.T) {
	p := DefaultParams()
	geomFI := FrameInput{
		Beds:    []Rect{bed(0, 0, 100, 200)},
		Covers:  []float64{1}, Onbed: []float64{1}, Overlap: []float64{1},
	}
	geom := BedGeoms(geomFI)
	f := belief.NewFilter(belief.DefaultModel(), 1)
	js := f.Space()
	cp := belief.NewCoupling(geom)
	em := belief.NewEmission(geom)

	// 默认 prior 起步 → 阶段A 充 occ + S→Bed → 阶段B 离线摔床沿（公开 API 全程，无需手改 alpha）。
	// 阶段A：人在床、sleepad 在线 InBed。
	inBed := FrameInput{
		Track:    RadarTrack{Online: true, Pose: 6, X: 50, Y: 100, StillSec: 60}, // 床内
		Sleepads: []SleepadFrame{{InBed: true, Fresh: true}},
		Beds:     geomFI.Beds, Covers: geomFI.Covers, Onbed: geomFI.Onbed, Overlap: geomFI.Overlap,
	}
	for step := 1; step <= 30; step++ {
		inBed.NowMs = int64(step) * 1000
		f.Step(inBed.NowMs, Online(inBed), cp.LogPsi(js, Gxy(inBed, p)), em.LogPhi(js, BuildObservation(inBed, p)))
	}
	if pb := js.MarginalB(f.Alpha(), 0); pb < 0.9 {
		t.Fatalf("阶段A 后 P(B occ)=%.3f 应高（在床充 occ）", pb)
	}

	// 阶段B：sleepad 离线 + 人摔到床沿地条（XY 出床、pose lying、HR/RR 灭）。
	fall := FrameInput{
		Track:    RadarTrack{Online: true, Pose: 6, X: 130, Y: 100, HR: 0, RR: 0, StillSec: 120}, // 床缘外 30
		Sleepads: []SleepadFrame{{InBed: true, Fresh: false}},                                    // 离线（陈旧 InBed）
		Beds:     geomFI.Beds, Covers: geomFI.Covers, Onbed: geomFI.Onbed, Overlap: geomFI.Overlap,
	}
	for step := 31; step <= 120; step++ {
		fall.NowMs = int64(step) * 1000
		f.Step(fall.NowMs, Online(fall), cp.LogPsi(js, Gxy(fall, p)), em.LogPhi(js, BuildObservation(fall, p)))
	}
	ms := js.MarginalS(f.Alpha())
	if ms[belief.SFallen] <= ms[belief.SBed] {
		t.Fatalf("cd2b 离线 replay：P(Fallen)=%.4f 应 > P(AtBed)=%.4f（adapter XY 派生 floor-strip 驱动）", ms[belief.SFallen], ms[belief.SBed])
	}

	// 裁决：独处 → fire（持续过 T_hold）。
	d := belief.NewDecider()
	var dec belief.Decision
	rc := belief.RiskContext{AloneContinuousMin: 30, PeopleCount: 1}
	pF := js.PFallen(f.Alpha())
	lam := belief.ComputeLambda(js, cp.LogPsi(js, Gxy(fall, p)), em.LogPhi(js, BuildObservation(fall, p)))
	for step := 121; step <= 230; step++ {
		dec = d.Step(int64(step)*1000, pF, lam, rc)
	}
	if !dec.Fire {
		t.Errorf("独处 cd2b 离线摔，持续过 T_hold 应 fire（P^F=%.3f Λ=%.2f margin=%.3f）", pF, lam, dec.Margin)
	}
	t.Logf("AD4 ✓ adapter→belief cd2b 离线 replay：P(Fallen)=%.4f > P(AtBed)=%.4f，独处 fire=%v（XY 派生 floor-strip）",
		ms[belief.SFallen], ms[belief.SBed], dec.Fire)
}
