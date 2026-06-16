package belief

import "testing"

// realness_test.go — realness 2 态前向滤波验收（device-room-zone.md §9）。
//   RV1 真人自主走动(非同步) → Real。
//   RV2【安全攸关】确认 real 后静止无证据 → PReal 保持高（Real 吸收 latch + 无 mirror 证据；cd2b 真人摔静止不当 ghost）。
//   RV3 同步移动镜像(track==2, 后到 B) → Mirror 涌现。
//   RV3b 墙外反射镜像(几何自判别，无需后到) → Mirror 涌现。
//   RV4 孤轨(Coexist=0) → 即便高 ρ/墙外裕度也停 real 先验（track==2 涌现，非硬 gate）。
//   RV5 近门出生(D 小) → Real（D 软发射）。
//   RV6 单 tick 瞬态(一帧高 ρ + 小 dt) → 不翻 Mirror（dt 积分，§9.4）。

const dtFrame = 1000 // 1s/帧

// RV1：真人自主独立走动(ρ 低，无门信息) → Real。
func TestRealnessHumanWalks(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 20; s++ {
		r.Update(RealnessObs{BirthDoorD: -1, Displaced: true, Coexist: 1, CoexistRho: 0.1, DtMs: dtFrame})
	}
	if r.PReal() < 0.9 {
		t.Errorf("自主走动 PReal=%.4f 应高(Real)", r.PReal())
	}
	t.Logf("RV1 ✓ 真人走动 PReal=%.4f", r.PReal())
}

// RV2【安全攸关】：确认 real 后静止无证据 → PReal 保持高，不被无证据/无 track 拽低(cd2b 病另一面)。
func TestRealnessStaysRealWhenStill(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 15; s++ { // 走动确认 real
		r.Update(RealnessObs{BirthDoorD: -1, Displaced: true, CoexistRho: 0.1, DtMs: dtFrame})
	}
	if pre := r.PReal(); pre < 0.9 {
		t.Fatalf("应先确认 real，PReal=%.4f", pre)
	}
	for s := 0; s < 120; s++ { // 摔倒后静止：无 Displaced/无 mirror 证据
		r.Update(RealnessObs{BirthDoorD: -1, DtMs: dtFrame})
	}
	if r.PReal() < 0.99 {
		t.Errorf("确认 real 静止后 PReal=%.4f 被拽低 → 真人摔倒当 ghost 漏报(cd2b)", r.PReal())
	}
	t.Logf("RV2 ✓ 确认 real→静止 120 帧 PReal=%.4f 仍高", r.PReal())
}

// RV3：同步移动镜像(track==2, 后到 B；墙内无几何，靠时序+同步) → Mirror。
func TestRealnessMirrorSync(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 30; s++ {
		r.Update(RealnessObs{BirthDoorD: 400, CoexistRho: 0.95, LaterBorn: true, Coexist: 1, DtMs: dtFrame})
	}
	if r.PMirror() < 0.7 {
		t.Errorf("后到+同步移动 PMirror=%.4f 应高(Mirror)", r.PMirror())
	}
	t.Logf("RV3 ✓ 同步镜像 PMirror=%.4f", r.PMirror())
}

// RV3b：墙外反射镜像(几何自判别，WallMargin>0，不依赖 LaterBorn) → Mirror。
func TestRealnessMirrorWall(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 30; s++ {
		r.Update(RealnessObs{BirthDoorD: 400, WallMargin: 0.5, LaterBorn: false, Coexist: 1, DtMs: dtFrame})
	}
	if r.PMirror() < 0.7 {
		t.Errorf("墙外反射 PMirror=%.4f 应高(Mirror，几何自判别)", r.PMirror())
	}
	t.Logf("RV3b ✓ 墙外镜像 PMirror=%.4f", r.PMirror())
}

// RV4：孤轨(Coexist=0) → mirror 发射中性 → 停 real 先验，即便高 ρ/墙外裕度(track==2 涌现，孤轨永 Real)。
func TestRealnessLoneAlwaysReal(t *testing.T) {
	r := NewRealnessTrack()
	for s := 0; s < 30; s++ {
		r.Update(RealnessObs{BirthDoorD: 400, CoexistRho: 0.95, LaterBorn: true, WallMargin: 1, Coexist: 0, DtMs: dtFrame})
	}
	if r.PReal() < 0.7 {
		t.Errorf("孤轨 PReal=%.4f 应停 real 先验(孤轨永 Real)", r.PReal())
	}
	t.Logf("RV4 ✓ 孤轨 PReal=%.4f(永 Real)", r.PReal())
}

// RV5：近门出生(D 小) = real 软证据——同样同步-ambiguous 下，近门 PReal > 远门，且近门仍判 real(>0.5)。
func TestRealnessBornNearDoor(t *testing.T) {
	near, far := NewRealnessTrack(), NewRealnessTrack()
	for s := 0; s < 20; s++ {
		near.Update(RealnessObs{BirthDoorD: 20, CoexistRho: 0.5, LaterBorn: true, Coexist: 1, DtMs: dtFrame})
		far.Update(RealnessObs{BirthDoorD: 400, CoexistRho: 0.5, LaterBorn: true, Coexist: 1, DtMs: dtFrame})
	}
	if near.PReal() <= far.PReal() {
		t.Errorf("近门 PReal=%.3f 应 > 远门 %.3f（D 软发射=real 证据）", near.PReal(), far.PReal())
	}
	if near.PReal() < 0.5 {
		t.Errorf("近门同步-ambiguous 应仍判 real(>0.5)，got %.3f", near.PReal())
	}
	t.Logf("RV5 ✓ 近门(D=20) PReal=%.3f > 远门(D=400) %.3f（D 把 ambiguous 拉回 real）", near.PReal(), far.PReal())
}

// RV6：单 tick 瞬态(仅一帧高 ρ + 小 dt) → 不翻 Mirror（§9.4 dt 积分）。
func TestRealnessSingleTickIgnored(t *testing.T) {
	r := NewRealnessTrack()
	r.Update(RealnessObs{BirthDoorD: 400, CoexistRho: 0.95, LaterBorn: true, Coexist: 1, DtMs: 20}) // 小 dt 瞬态
	r.Update(RealnessObs{BirthDoorD: 400, CoexistRho: 0.1, LaterBorn: true, Coexist: 1, DtMs: dtFrame})
	if r.PMirror() > 0.3 {
		t.Errorf("单 tick 瞬态 PMirror=%.4f 不该翻 Mirror(§9.4)", r.PMirror())
	}
	t.Logf("RV6 ✓ 单 tick 瞬态 PMirror=%.4f(忽略)", r.PMirror())
}
