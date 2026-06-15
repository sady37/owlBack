package belief

import "testing"

// realness_test.go — realness 隐轴涌现验收（§35 三缺口补全；C 验收点）：
//   RV1 真人：入口出生 + 走动 → P(real) 涌现
//   RV2 静止金属反射体：困 BirthPos + 久 + 近墙（从未走开）→ P(static) 涌现
//   RV3 动态镜像：共动 + 镜面指向 → P(mirror) 涌现（一真一镜）
//   RV4【安全攸关】闸门：已确认 real（走动过）后静止消失 → P(real) **保持高**，不被静止重判 ghost（cd2b 病另一面）
//   RV5 共生律：丢真人 track 但镜像存活 → PRoomHasReal 仍高（fall 不被「无 track」抑制）

const rveps = 0.05

// RV1：真人 = 入口出生 + 走动。
func TestRealnessHumanWalksIn(t *testing.T) {
	r := NewRealnessTrack(true /*入口出生*/, false)
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{Displaced: true}) // 持续走动
	}
	if r.PReal() < 0.9 {
		t.Errorf("入口出生+走动 P(real)=%.4f 应高", r.PReal())
	}
	t.Logf("RV1 ✓ 真人：P(real)=%.4f P(mirror)=%.4f P(static)=%.4f", r.PReal(), r.PMirror(), r.PStatic())
}

// RV2：静止金属 = 困 BirthPos + 久 + 近墙，从未走开。
func TestRealnessStaticMetal(t *testing.T) {
	r := NewRealnessTrack(false, true /*金属区出生*/)
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{ConfinedNearWall: true, AgeLongStatic: true}) // 困住、久、近墙；从不 Displaced
	}
	if r.PStatic() < 0.8 {
		t.Errorf("困+久+近墙+从未走开 P(static)=%.4f 应高", r.PStatic())
	}
	t.Logf("RV2 ✓ 静止金属：P(static)=%.4f P(real)=%.4f", r.PStatic(), r.PReal())
}

// RV3：动态镜像 = 共动 + 镜面指向本 track。
func TestRealnessMirror(t *testing.T) {
	r := NewRealnessTrack(false, false)
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{Displaced: true, CoexistRho: 0.9, IsReflection: true}) // 共动 + 是反射
	}
	if r.PMirror() < 0.8 {
		t.Errorf("共动+镜面指向 P(mirror)=%.4f 应高", r.PMirror())
	}
	t.Logf("RV3 ✓ 动态镜像：P(mirror)=%.4f P(real)=%.4f", r.PMirror(), r.PReal())
}

// RV4【安全攸关】：已确认 real（走动过）→ 后续静止消失，P(real) 保持高，不被静止重判 ghost。
// 这是 cd2b 病另一面的防护：真人摔倒静止 → track 静止/消失 → 绝不当 ghost 否定。
func TestRealnessGateConfirmedRealStaysReal(t *testing.T) {
	r := NewRealnessTrack(true, false)
	// 阶段A：走入并走动 → 确认 real（置 movedFromBirth 闩）。
	for s := 0; s < 15; s++ {
		r.Update(RealnessObs{Displaced: true})
	}
	pRealConfirmed := r.PReal()
	if pRealConfirmed < 0.9 {
		t.Fatalf("阶段A 应确认 real，P(real)=%.4f", pRealConfirmed)
	}
	// 阶段B：摔倒后静止（不再 Displaced；甚至呈现 static 签名——但已走开过，闸门挡住 Static 重判）。
	for s := 0; s < 60; s++ {
		r.Update(RealnessObs{ConfinedNearWall: true, AgeLongStatic: true}) // 静止，static 签名俱全
	}
	if r.PReal() < 0.7 {
		t.Errorf("闸门失效：已确认 real 静止后 P(real)=%.4f 跌了 → 真人摔倒被当 ghost 漏报(cd2b 病)", r.PReal())
	}
	if r.PStatic() > 0.2 {
		t.Errorf("闸门失效：走动过的真人被重判 Static=%.4f（movedFromBirth 应挡住）", r.PStatic())
	}
	t.Logf("RV4 ✓ 闸门：确认 real→静止60帧后 P(real)=%.4f 仍高、P(static)=%.4f 被挡（真人摔倒不当 ghost）",
		r.PReal(), r.PStatic())
}

// RV5：共生律——丢真人但镜像存活 → PRoomHasReal 仍高。
func TestRealnessSymbiosis(t *testing.T) {
	// 真人 track（已确认）+ 镜像 track（高 P(mirror)）。
	human := NewRealnessTrack(true, false)
	mirror := NewRealnessTrack(false, false)
	for s := 0; s < 20; s++ {
		human.Update(RealnessObs{Displaced: true})
		mirror.Update(RealnessObs{Displaced: true, CoexistRho: 0.9, IsReflection: true})
	}
	full := PRoomHasReal([]*RealnessTrack{human, mirror})
	// 真人 track 丢失（摔倒被滤），只剩镜像 track → 共生律：镜像必有真人源 → 房内仍有真人。
	onlyMirror := PRoomHasReal([]*RealnessTrack{mirror})
	if onlyMirror < 0.8 {
		t.Errorf("共生律失效：只剩镜像 PRoomHasReal=%.4f 应仍高（镜像必有真人源 → 真人仍在 → fall 不抑制）", onlyMirror)
	}
	t.Logf("RV5 ✓ 共生律：双 track PRoomHasReal=%.4f；丢真人只剩镜像 PRoomHasReal=%.4f（真人仍在，fall 不抑制）",
		full, onlyMirror)
	_ = rveps
}
