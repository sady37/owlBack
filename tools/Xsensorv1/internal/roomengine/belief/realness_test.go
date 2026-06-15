package belief

import "testing"

// realness_test.go — realness 隐轴验收（DBN-Zone-Room §G：两类 Real/Mirror，Static 溶解）：
//   RV1 真人：走动 → 锁 Real（P(real)=1）。
//   RV3 动态镜像：共动 + 镜面指向 → P(mirror) 涌现。
//   RV4【安全攸关】无 leak + 闸门：已确认 real（走动过）后**静止/空闲不再有证据** → P(real) 保持 1，
//        绝不被「无证据」拽低（§G 删 leak 治本：cd2b 病另一面，真人摔倒静止不当 ghost 漏报）。
//   RV5 共生律：丢真人 track 但镜像存活 → PRoomHasReal 仍高（fall 不被「无 track」抑制）。

// RV1：真人 = 走动 → 自主走动锁 Real。
func TestRealnessHumanWalks(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{Displaced: true})
	}
	if r.PReal() < 0.9 {
		t.Errorf("走动 P(real)=%.4f 应高（锁 Real）", r.PReal())
	}
	t.Logf("RV1 ✓ 真人：P(real)=%.4f P(mirror)=%.4f", r.PReal(), r.PMirror())
}

// RV3：动态镜像 = 共动 + 镜面指向本 track（同步反射，非自主走动）。
func TestRealnessMirror(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{Displaced: true, CoexistRho: 0.9, IsReflection: true}) // 共动 + 是反射
	}
	if r.PMirror() < 0.8 {
		t.Errorf("共动+镜面指向 P(mirror)=%.4f 应高", r.PMirror())
	}
	t.Logf("RV3 ✓ 动态镜像：P(mirror)=%.4f P(real)=%.4f", r.PMirror(), r.PReal())
}

// RV4【安全攸关】：已确认 real（走动过）→ 后续无证据（静止/空闲），P(real) 保持 1，不被「无证据」拽低。
// §G 删 leak 治本：无 ghost 正证据绝不凭空造 ghost 质量；真人摔倒静止 → track 静止/消失 → 绝不当 ghost 否定。
func TestRealnessNoLeakStaysReal(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 15; s++ { // 阶段A：走动确认 real（置 movedFromBirth 闩）。
		r.Update(RealnessObs{Displaced: true})
	}
	if r.PReal() < 0.9 {
		t.Fatalf("阶段A 应确认 real，P(real)=%.4f", r.PReal())
	}
	for s := 0; s < 60; s++ { // 阶段B：摔倒后静止/空闲，无任何证据（无 Displaced、无 co-existence）。
		r.Update(RealnessObs{})
	}
	if r.PReal() < 0.99 {
		t.Errorf("leak 复活：确认 real 静止后 P(real)=%.4f 被无证据拽低 → 真人摔倒被当 ghost 漏报(cd2b 病)", r.PReal())
	}
	t.Logf("RV4 ✓ 无 leak + 闸门：确认 real→静止 60 帧后 P(real)=%.4f 仍 1（真人摔倒不当 ghost）", r.PReal())
}

// RV5：共生律——丢真人但镜像存活 → PRoomHasReal 仍高（任一 track 皆蕴含真人源）。
func TestRealnessSymbiosis(t *testing.T) {
	human := NewRealnessTrack()
	mirror := NewRealnessTrack()
	for s := 0; s < 20; s++ {
		human.Update(RealnessObs{Displaced: true})
		mirror.Update(RealnessObs{Displaced: true, CoexistRho: 0.9, IsReflection: true})
	}
	full := PRoomHasReal([]*RealnessTrack{human, mirror})
	// 真人 track 丢失（摔倒被滤），只剩镜像 track → 共生律：镜像必有真人源 → 房内仍有真人。
	onlyMirror := PRoomHasReal([]*RealnessTrack{mirror})
	if onlyMirror < 0.8 {
		t.Errorf("共生律失效：只剩镜像 PRoomHasReal=%.4f 应仍高", onlyMirror)
	}
	t.Logf("RV5 ✓ 共生律：双 track=%.4f；丢真人只剩镜像=%.4f（真人仍在，fall 不抑制）", full, onlyMirror)
}
