package adapter

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// adapter_test.go — 翻译层验收（集成步骤1；§32 拆 floor-strip 后）：
//   AD2 Gxy 派生（§4 attachment，保留：内→peak / 近→Near / 远→0 / 近两床→均匀）
//   AD3 BuildObservation 映射（pose=6/sleepad/VitalSourceOnline/nearBed/HRRR）
//   AD4 端到端框架涌现（adapter→belief，sleepad 在线 InBed→LeftBed，零 floor-strip → fire via 接触轴）

func bed(x1, y1, x2, y2 int) Rect { return Rect{x1, y1, x2, y2} }

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
		Sleepads: []SleepadFrame{{Present: true, Reading: belief.BedLeftBed}}, // 在线但离床
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
	if !o.NearBed {
		t.Error("床缘 30cm 应 NearBed")
	}
	// 铁律 [[radar_hr_rr_bed_enter_gated]]：HR=RR=0（雷达未返 vital）= 零信息，HRRRObserved/Present 均 false，
	// 不触发 §D absent 否决（近床但 enter-gate 未测 ≠ 观测到 absent）。
	if o.HRRRObserved || o.HRRRPresent {
		t.Error("HR=RR=0 应 HRRRObserved=false + HRRRPresent=false（雷达未测 vital=零信息）")
	}
	// §32：FloorStripXY 已删，cd2b 靠 LeftBed→B vac 经 Ψ 涌现（见 AD4），非雷达 XY 几何。
	// sleepad 离线 → NoReport + VitalSourceOnline=false
	fi.Sleepads = []SleepadFrame{{Present: true, Reading: belief.BedNoReport}}
	o2 := BuildObservation(fi, p)
	if o2.Sleepad[0] != belief.BedNoReport || o2.VitalSourceOnline {
		t.Errorf("¬Fresh 应 NoReport + VitalSourceOnline=false，得 %v / %v", o2.Sleepad[0], o2.VitalSourceOnline)
	}
}

// AD4（§32 框架涌现，替原离线+floor-strip 补丁测试）：端到端 adapter→belief cd2b。
// sleepad **全程在线** InBed→LeftBed，雷达躺+静止，**零 floor-strip**。LeftBed→B vac→Ψ→SFallen
// 涌现，独处持续过 T_hold → fire。证翻译层 + 框架接触轴解 cd2b，不靠几何 δ。
func TestAdapterCd2bFrameworkEmergence(t *testing.T) {
	p := DefaultParams()
	geomFI := FrameInput{
		Beds:   []Rect{bed(0, 0, 100, 200)},
		Covers: []float64{1}, Onbed: []float64{1}, Overlap: []float64{1},
	}
	geom := BedGeoms(geomFI)
	f := belief.NewFilter(belief.DefaultModel(), 1)
	js := f.Space()
	cp := belief.NewCoupling(geom)
	em := belief.NewEmission(geom)

	// 阶段A：人在床、sleepad 在线 InBed。
	inBed := FrameInput{
		Track:    RadarTrack{Online: true, Pose: 6, X: 50, Y: 100, StillSec: 60},
		Sleepads: []SleepadFrame{{Present: true, Reading: belief.BedInBed}},
		Beds:     geomFI.Beds, Covers: geomFI.Covers, Onbed: geomFI.Onbed, Overlap: geomFI.Overlap,
	}
	for step := 1; step <= 30; step++ {
		inBed.NowMs = int64(step) * 1000
		f.Step(inBed.NowMs, Online(inBed), cp.LogPsi(js, Gxy(inBed, p)), em.LogPhi(js, BuildObservation(inBed, p)), 0, 1)
	}
	if pb := js.MarginalB(f.Alpha(), 0); pb < 0.9 {
		t.Fatalf("阶段A 后 P(B occ)=%.3f 应高（在床充 occ）", pb)
	}

	// 阶段B：sleepad **在线**报 LeftBed（人离床）+ 雷达仍躺+静止 + 近床。**零 floor-strip**。
	fall := FrameInput{
		Track:    RadarTrack{Online: true, Pose: 6, X: 130, Y: 100, HR: 0, RR: 0, StillSec: 120},
		Sleepads: []SleepadFrame{{Present: true, Reading: belief.BedLeftBed}}, // 在线 LeftBed（§32 二态，不离线）
		Beds:     geomFI.Beds, Covers: geomFI.Covers, Onbed: geomFI.Onbed, Overlap: geomFI.Overlap,
	}
	for step := 31; step <= 120; step++ {
		fall.NowMs = int64(step) * 1000
		f.Step(fall.NowMs, Online(fall), cp.LogPsi(js, Gxy(fall, p)), em.LogPhi(js, BuildObservation(fall, p)), 0, 1)
	}
	pF := js.PFallen(f.Alpha())
	if pF < 0.55 {
		t.Fatalf("框架涌现：cd2b 零补丁 P(SFallen)=%.4f 应≥0.55（LeftBed→B vac→Ψ）", pF)
	}

	// 裁决：独处 → fire（持续过 T_hold）。
	d := belief.NewDecider()
	var dec belief.Decision
	rc := belief.RiskContext{AloneContinuousMin: 30, PeopleCount: 1}
	lam := belief.ComputeLambda(js, cp.LogPsi(js, Gxy(fall, p)), em.LogPhi(js, BuildObservation(fall, p)))
	for step := 121; step <= 230; step++ {
		dec = d.Step(int64(step)*1000, pF, lam, rc)
	}
	if !dec.Fire {
		t.Errorf("独处 cd2b 接触轴摔，持续过 T_hold 应 fire（P^F=%.3f Λ=%.2f band=%s）", pF, lam, dec.Band)
	}
	t.Logf("AD4 ✓ adapter→belief 框架接触轴：P(SFallen)=%.4f，独处 fire=%v（零 floor-strip，sleepad 在线 LeftBed）",
		pF, dec.Fire)
}
