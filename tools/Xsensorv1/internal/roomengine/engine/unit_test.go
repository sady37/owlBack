package engine

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// unit_test.go — 步4 多房 unit 编排器（纯时间窗 neighbor 接线）验收：
//   UV1 hand-off：A 丢真人 ↔ 兄弟房 B 窗内新现真人 → ρ_xroom(A)>0 → A 的 Blind→Fallen 整流入 Left（人挪去B）
//   UV2 no-handoff：A 丢真人、无兄弟房新现 → ρ(A)=0 → 保 lost-fall（A 的 Fallen 不被抑制，安全默认）
//   UV3 单房零回归：单房 unit 无兄弟 → ρ≡0（与单房 Room.Tick(fi,0) 等价；cd2b 走 replay 另验）
//   UV4 ★噪声防线（§64）：A 丢 ↔ B 仅噪声 churn 假 track（永不达确认 K）→ 不算合法 GainedReal → ρ(A)=0 → 保 lost-fall

// openFloorTrack 远离床的开阔地行人（pose 站立，x 微动=真人走动锁 Real）。
func openFloorTrack(x int) []adapter.TrackObs { return []adapter.TrackObs{tk(0, x, 400, 0)} }

func unitFrame(now int64, beds []adapter.Rect, cov []float64, tracks []adapter.TrackObs) adapter.FrameInput {
	return adapter.FrameInput{
		NowMs: now, Tracks: tracks,
		Sleepads: []adapter.SleepadFrame{{Present: true, Reading: belief.BedLeftBed}}, // 床空（人在开阔地）
		Beds:     beds, Covers: cov, Onbed: cov, Overlap: cov,
		Census: adapter.Census{AloneContinuousMin: 30},
	}
}

// runUnitAB 驱动两房 A/B 一段时间线；handoff=true 则 B 在 A 丢轨后 +2s 新现真人（hand-off），否则 B 全程空。
// 返回 A 末帧 S 边缘 + A 曾收到的最大 ρ。
func runUnitAB(handoff bool) (belief.Vector, float64) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	mk := func() *Room {
		return NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)
	}
	u := NewUnit(map[string]*Room{"A": mk(), "B": mk()}, 1)

	maxRhoA := 0.0
	for step := 1; step <= 60; step++ {
		now := int64(step) * 1000
		// A：1-15 行人在开阔地走；16+ 消失（blind，lost-fall 二义）。
		var aTracks []adapter.TrackObs
		if step <= 15 {
			aTracks = openFloorTrack(400 + step*10)
		}
		// B：默认空；handoff 时第 18s 起有人（A 丢@16s → Δ=2s 峰区）。
		var bTracks []adapter.TrackObs
		if handoff && step >= 18 {
			bTracks = openFloorTrack(700 + step*10)
		}
		u.Tick("A", unitFrame(now, beds, cov, aTracks))
		u.Tick("B", unitFrame(now, beds, cov, bTracks))
		if rho := u.LastRho("A"); rho > maxRhoA {
			maxRhoA = rho
		}
	}
	return u.rooms["A"].MarginalS(), maxRhoA
}

func TestUnitHandoffRoutesToLeft(t *testing.T) {
	margH, rhoH := runUnitAB(true)
	margN, rhoN := runUnitAB(false)

	// 编排器接线：hand-off 检出 ρ>0；无 hand-off ρ=0。
	if rhoH <= 0 {
		t.Errorf("UV1 hand-off 应检出 ρ(A)>0（A 丢 ↔ B 窗内新现，守恒+时间窗），得 %.4f", rhoH)
	}
	if rhoN != 0 {
		t.Errorf("UV2 无 hand-off ρ(A) 应=0（无兄弟房新现），得 %.4f", rhoN)
	}
	// 效果方向：hand-off 把 Blind→Fallen 整流入 Left → A 的 Left 较高、Fallen 较低（人挪去 B 非本房摔）。
	if margH[belief.SLeft] < margN[belief.SLeft] {
		t.Errorf("UV1 hand-off 应抬 A 的 P(Left)：handoff %.4f vs none %.4f（人挪去兄弟房）", margH[belief.SLeft], margN[belief.SLeft])
	}
	if margH[belief.SFallen] > margN[belief.SFallen] {
		t.Errorf("UV1 hand-off 应压 A 的 P(Fallen)：handoff %.4f vs none %.4f", margH[belief.SFallen], margN[belief.SFallen])
	}
	t.Logf("UV1/2 ✓ ρ(A) hand-off=%.3f / none=%.3f；A P(Left) %.4f→%.4f、P(Fallen) %.4f→%.4f（整流入 Left）",
		rhoH, rhoN, margN[belief.SLeft], margH[belief.SLeft], margN[belief.SFallen], margH[belief.SFallen])
}

// noiseTrack cd2b 式噪声假 track：每帧跳变 >AssocCm=250 → census 每帧拆新 logicID（churn），永不确认。
func noiseTrack(step int) []adapter.TrackObs {
	x := 700
	if step%2 == 0 {
		x = 1300 // 600cm 跳变 > AssocCm → 每帧新 logicID，streak 永=1 < K
	}
	return []adapter.TrackObs{tk(0, x, 400, 0)}
}

// UV4（★ §64 噪声防线，C 要求补）：A 丢真人 ↔ B 仅噪声 churn 假 track（永不达确认 K）→ 不算合法 GainedReal
//
//	→ ρ(A)=0 → A 真摔(lost-fall)照保（不被假 hand-off 整流成 Left）。防 §61 同源噪声在 neighbor 层造漏报。
func TestUnitNoiseFakeTrackNotHandoff(t *testing.T) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	mk := func() *Room {
		return NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)
	}
	u := NewUnit(map[string]*Room{"A": mk(), "B": mk()}, 1)

	maxRhoA := 0.0
	for step := 1; step <= 60; step++ {
		now := int64(step) * 1000
		var aTracks []adapter.TrackObs
		if step <= 15 {
			aTracks = openFloorTrack(400 + step*10)
		}
		var bTracks []adapter.TrackObs
		if step >= 18 {
			bTracks = noiseTrack(step) // B 全程仅噪声 churn 假 track（无真人）
		}
		u.Tick("A", unitFrame(now, beds, cov, aTracks))
		u.Tick("B", unitFrame(now, beds, cov, bTracks))
		if rho := u.LastRho("A"); rho > maxRhoA {
			maxRhoA = rho
		}
	}
	margNoise := u.rooms["A"].MarginalS()
	margNone, _ := runUnitAB(false) // 对照：A 同样丢、B 全程空

	if maxRhoA != 0 {
		t.Errorf("UV4 噪声防线破：B 噪声 churn 假 track 不应算合法 GainedReal → ρ(A) 应=0，得 %.4f（否则假 hand-off 漏真摔）", maxRhoA)
	}
	if margNoise[belief.SFallen] < margNone[belief.SFallen]-1e-9 {
		t.Errorf("UV4 噪声不得压 A 的 lost-fall：noise P(Fallen)=%.4f < none %.4f（漏报）", margNoise[belief.SFallen], margNone[belief.SFallen])
	}
	t.Logf("UV4 ✓ B 噪声 churn → ρ(A)=%.3f（不算合法落点）→ A lost-fall 保留 P(Fallen)=%.4f（=no-handoff，不漏真摔）", maxRhoA, margNoise[belief.SFallen])
}

func TestUnitSingleRoomZeroNeighbor(t *testing.T) {
	beds := []adapter.Rect{bedRect(0, 0, 100, 200)}
	cov := []float64{1}
	r := NewRoom(adapter.BedGeoms(adapter.FrameInput{Beds: beds, Covers: cov, Onbed: cov, Overlap: cov}), 1)
	u := NewUnit(map[string]*Room{"A": r}, 1)
	for step := 1; step <= 30; step++ {
		u.Tick("A", unitFrame(int64(step)*1000, beds, cov, openFloorTrack(400+step*10)))
		if rho := u.LastRho("A"); rho != 0 {
			t.Fatalf("UV3 单房 unit 无兄弟 → ρ≡0，第 %d 帧得 %.4f", step, rho)
		}
	}
	// 人消失后也不应有 neighbor 抑制（无兄弟房）。
	for step := 31; step <= 50; step++ {
		u.Tick("A", unitFrame(int64(step)*1000, beds, cov, nil))
		if rho := u.LastRho("A"); rho != 0 {
			t.Fatalf("UV3 单房消失后 ρ 仍应=0（无兄弟房 hand-off），第 %d 帧得 %.4f", step, rho)
		}
	}
	t.Logf("UV3 ✓ 单房 unit ρ≡0（无兄弟→无 hand-off，与单房 Room.Tick(fi,0) 等价，cd2b 零回归）")
}
