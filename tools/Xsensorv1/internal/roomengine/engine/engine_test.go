package engine

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// engine_test.go — 单房 roomengine 端到端验收（DBN-wire-roadmap §6 + §57 步2 多 track）：
//   EG1 cd2b 单房零回归（单 track → 1 filter，Ψ 相容涌现 SFallen）
//   EG2 pillar C：真人+墙外反射 ghost → decide.PeopleCount=1（census 排 ghost，§56）
//   EG3 pillar D：N_r=2 + ≥55% 真摔 → report fire（拍法 A 守门）
//   EG4 步2 blind 续存：faller track 消失 → Filter 自持 → 告警连续；N_r 掉 0（§60）
//   EG5 步2 OR 聚合：一房两真人、仅一人摔 → 房间 fire（§60 架构师拍）

func bedRect(x1, y1, x2, y2 int) adapter.Rect { return adapter.Rect{X1: x1, Y1: y1, X2: x2, Y2: y2} }

// track 全量 obs（§57 步2）。
func tk(pose, x, y int, still float64) adapter.TrackObs {
	return adapter.TrackObs{RadarTrack: adapter.RadarTrack{Online: true, Pose: pose, X: x, Y: y, StillSec: still}}
}

func newRoom1Bed() (*Room, []adapter.Rect, []float64) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	r := NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)
	return r, beds, cov
}

func TestEngineCd2bSingleRoomReproduces(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	mk := func(now int64, reading belief.BedReading, x int) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs:    now,
			Tracks:   []adapter.TrackObs{tk(6, x, 100, 120)}, // 单住户 → N_r=1
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
	}
	// 阶段A：在床 InBed 充 occ（雷达 X=50 床心=出生位，未位移）。
	for step := 1; step <= 30; step++ {
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50), 0)
	}
	// 阶段B：在线 LeftBed + 雷达仍躺静止近床（cd2b 床边真摔，§32 二态不离线）。X=130 偏离出生位 → Displaced 锁 Real。
	var fired bool
	for step := 31; step <= 150; step++ {
		if r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130), 0).Decision.Fire {
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
	t.Logf("EG1 ✓ cd2b 单房零回归：Room.Tick 端到端 P(SFallen)=%.4f fire=%v", pF, fired)
}

// EG2（pillar C）：真人 + 墙外反射 mirror ghost 经 Room.Tick → census 排 mirror → decide.PeopleCount=1（独处）。
func TestEnginePillarCGhostExcludedToDecide(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	radar := adapter.Point{X: 50, Y: 250}
	walls := []adapter.Rect{{X1: -200, Y1: 340, X2: 300, Y2: 360}} // 横墙：影子在墙外、radar→影子穿墙
	var dec belief.Decision
	for step := 1; step <= 15; step++ {
		fi := adapter.FrameInput{
			NowMs: int64(step) * 1000,
			Tracks: []adapter.TrackObs{
				tk(6, 50+step*5, 100, 120), // 真人（radar 同侧墙内，不穿墙 → Real）
				tk(6, 50+step*5, 400, 120), // 影子（墙外、radar→影子穿墙，同步同速）→ Mirror 几何自判别
			},
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: belief.BedLeftBed}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Walls: walls, RadarPos: radar,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
		dec = r.Tick(fi, 0).Decision
	}
	if dec.PeopleCount != 1 {
		t.Errorf("真人+墙外反射 mirror 应 N_r=1 到 decide（census 排 ghost 防虚增），得 PeopleCount=%d", dec.PeopleCount)
	}
	t.Logf("EG2 ✓ pillar C：census 排墙外反射 ghost → decide.PeopleCount=%d", dec.PeopleCount)
}

// EG3（pillar D）：拍法 A 守门——N_r=2 + 一人 ≥55% 真摔 → 必报，C_FN 不参与（≥55% 证据自足）。
func TestEnginePillarDHoldsAtReportBand(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	mk := func(now int64, reading belief.BedReading, x int) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs: now,
			Tracks: []adapter.TrackObs{ // 两独立真人（异步：甲随 x 动+躺、乙固定站立远处）→ N_r=2
				tk(6, x, 100, 120),
				tk(0, 600, 400, 0),
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
		t.Errorf("拍法 A 守门：N_r=2 + ≥55%% 真摔必报（band 应 report、Fire），得 band=%s fire=%v", dec.Band, fired)
	}
	t.Logf("EG3 ✓ pillar D：N_r=%d + P^F=%.4f → band=%s fire=%v", dec.PeopleCount, dec.PFallen, dec.Band, fired)
}

// EG4（步2 blind 续存，§60）：faller track 消失 → 其 Filter 仅 Predict 自持 → SFallen 持住 → 告警连续；
// 同时 census 不数消失者 → N_r 掉 0（FN-safe，当独处更易报）。证「不蒸发的是告警、非人数」。
func TestEngineBlindPersistsAlarmNotCount(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	mk := func(now int64, reading belief.BedReading, x int, tracks []adapter.TrackObs) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs: now, Tracks: tracks,
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
	}
	one := func(x int) []adapter.TrackObs { return []adapter.TrackObs{tk(6, x, 100, 120)} }
	for step := 1; step <= 30; step++ { // 在床充 occ
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50, one(50)), 0)
	}
	var dec belief.Decision
	for step := 31; step <= 150; step++ { // 床边真摔 → fire
		dec = r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130, one(130)), 0).Decision
	}
	if !dec.Fire {
		t.Fatalf("前置：faller 在场应已 fire，得 fire=%v band=%s", dec.Fire, dec.Band)
	}
	// track 消失（Tracks 空）：blind 续存 → 告警应继续，N_r 掉 0。
	var afterFire bool
	for step := 151; step <= 200; step++ {
		dec = r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 0, nil), 0).Decision
		afterFire = dec.Fire
	}
	if dec.PeopleCount != 0 {
		t.Errorf("faller 消失 → census 不数 → N_r=0（FN-safe），得 %d", dec.PeopleCount)
	}
	if !afterFire {
		t.Errorf("blind 续存：faller track 消失后告警应连续（S^(i) 自持 Fallen），得 fire=%v", afterFire)
	}
	t.Logf("EG4 ✓ blind 续存：消失后 fire=%v（告警连续）/ N_r=%d（人数掉出=FN-safe）—— 不蒸发的是告警非人数", afterFire, dec.PeopleCount)
}

// EG5（步2 OR 聚合，§60 架构师拍）：一房两真人、仅一人床边摔，另一人站立正常 → 房间 fire（任一真人摔即报）。
func TestEngineORAggregation(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	mk := func(now int64, reading belief.BedReading, x int) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs: now,
			Tracks: []adapter.TrackObs{
				tk(6, x, 100, 120), // 甲：床边躺/摔
				tk(0, 700, 500, 0), // 乙：远处站立正常（不摔）
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
		t.Fatalf("两真人应 N_r=2，得 %d", dec.PeopleCount)
	}
	if !fired {
		t.Errorf("OR 聚合：两真人一人摔 → 房间应 fire（任一真人 Fallen 过阈），得 fire=%v", fired)
	}
	t.Logf("EG5 ✓ OR 聚合：N_r=%d 一人摔 → 房间 fire=%v（报到房间）", dec.PeopleCount, fired)
}

// EG6（§61 消费门控）：孤轨（无共存源）即便 XY 噪声跳变，engine 强制 pFallReal=1（永发）→ 真摔照报。
// realness 重写后孤轨 Coexist=0 → PReal 恒 1（mEv=0，不再被压），§61 门控是叠加防线：证「噪声 XY 不得漏孤身真摔」。
func TestEngineConsumptionGateLoneArtifactStillFires(t *testing.T) {
	r, beds, cov := newRoom1Bed()
	mk := func(now int64, reading belief.BedReading, x int, still float64) adapter.FrameInput {
		return adapter.FrameInput{
			NowMs:    now,
			Tracks:   []adapter.TrackObs{tk(6, x, 100, still)},
			Sleepads: []adapter.SleepadFrame{{Present: true, Reading: reading}},
			Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
			Census: adapter.Census{AloneContinuousMin: 30},
		}
	}
	// 阶段0：XY 来回跳（<AssocCm=250 不churn）。realness 重写后孤轨 PReal 不再被压，门控仍须保证永发。
	for step := 1; step <= 20; step++ {
		x := 50
		if step%2 == 0 {
			x = 200
		}
		r.Tick(mk(int64(step)*1000, belief.BedInBed, x, 0), 0)
	}
	// 阶段A：在床静止充 occ。
	for step := 21; step <= 50; step++ {
		r.Tick(mk(int64(step)*1000, belief.BedInBed, 50, 120), 0)
	}
	// 阶段B：床边真摔。
	var dec belief.Decision
	var fired bool
	for step := 51; step <= 175; step++ {
		dec = r.Tick(mk(int64(step)*1000, belief.BedLeftBed, 130, 120), 0).Decision
		if dec.Fire {
			fired = true
		}
	}
	if !fired {
		t.Errorf("孤轨无共存源 → 消费门控 pFallReal=1 永发 → 应 fire（噪声 XY 不得漏孤身真摔）")
	}
	t.Logf("EG6 ✓ 消费门控：孤轨无共存源→永发 fire=%v / N_r=%d", fired, dec.PeopleCount)
}
