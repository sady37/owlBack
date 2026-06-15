package belief

import "testing"

// ghost_test.go — realness 隐轴涌现验收（§34 track=2；C 验收点：一真一镜→P(ghost)对镜涌现、对真低）：
//   GH1 co-existence + 镜面指向 B → P(GhostB) 涌现高、P(GhostA) 低（零硬外挂，纯 ρ 涌现）
//   GH2 独立两 track（ρ=0）→ 皆 Real（P(ghost)≈0，§10 孤立安全）
//   GH3 病根规避：无 co-existence（ρ=0）→ 永不 ghost（不靠单 track 静止）
//   GH4 对称（镜面未知，ρ 高）→ 恰一个 ghost（PExactlyOne 高、PGG≈0），但不确定哪个

// GH1：一真(A) 一镜(B)，共动 + 镜面指向 B。
func TestGhostEmergenceMirror(t *testing.T) {
	g := NewGhostPair()
	for s := 0; s < 30; s++ {
		g.Update(0.9, 0.1) // ρ 高，biasA=0.1（镜面指向 B 是反射）
	}
	if g.PGhostB() < 0.8 {
		t.Errorf("镜 track B 的 P(ghost)=%.4f 应高（co-existence 涌现）", g.PGhostB())
	}
	if g.PGhostA() > 0.2 {
		t.Errorf("真 track A 的 P(ghost)=%.4f 应低", g.PGhostA())
	}
	t.Logf("GH1 ✓ 一真一镜：P(GhostB)=%.4f（镜，涌现）> P(GhostA)=%.4f（真）零硬外挂", g.PGhostB(), g.PGhostA())
}

// GH2：独立两 track（无共现，ρ=0）→ 皆 Real。
func TestGhostIndependentBothReal(t *testing.T) {
	g := NewGhostPair()
	for s := 0; s < 30; s++ {
		g.Update(0.0, 0.5) // ρ=0：两个真人各自活动，无共现
	}
	if g.PGhostA() > 0.05 || g.PGhostB() > 0.05 {
		t.Errorf("独立两 track 应皆 Real：P(GhostA)=%.4f P(GhostB)=%.4f 应≈0（§10 孤立安全）", g.PGhostA(), g.PGhostB())
	}
	t.Logf("GH2 ✓ 独立两 track 皆 Real：P(GhostA)=%.4f P(GhostB)=%.4f", g.PGhostA(), g.PGhostB())
}

// GH3（病根规避）：ρ=0 时无论多久都不产 ghost——ghost 只由 co-existence，不由单 track 静止/realness。
func TestGhostNoStillnessRoot(t *testing.T) {
	g := NewGhostPair()
	// 模拟「一个 track 久静」场景：但 realness 轴**没有**静止输入，只有 co-existence ρ。ρ=0 → 永不 ghost。
	for s := 0; s < 200; s++ {
		g.Update(0.0, 0.5)
	}
	if g.PExactlyOneGhost() > 0.05 {
		t.Errorf("病根：无 co-existence 久持 P(恰一 ghost)=%.4f 应≈0（绝不靠单 track 静止判 ghost）", g.PExactlyOneGhost())
	}
	t.Logf("GH3 ✓ 病根规避：ρ=0 久持 P(恰一 ghost)=%.4f（ghost 只从 co-existence 涌现）", g.PExactlyOneGhost())
}

// GH4：对称（镜面未知 biasA=0.5，ρ 高）→ 恰一个 ghost，但不确定哪个。
func TestGhostSymmetricExactlyOne(t *testing.T) {
	g := NewGhostPair()
	for s := 0; s < 30; s++ {
		g.Update(0.9, 0.5) // ρ 高、镜面未知
	}
	if g.PExactlyOneGhost() < 0.8 {
		t.Errorf("co-existence 高 P(恰一 ghost)=%.4f 应高", g.PExactlyOneGhost())
	}
	if g.p[pGG] > 0.05 {
		t.Errorf("P(GG 两镜)=%.4f 应≈0（不能都是镜）", g.p[pGG])
	}
	// 对称：哪个是 ghost 不确定，两者接近
	if d := g.PGhostA() - g.PGhostB(); d > 0.2 || d < -0.2 {
		t.Errorf("镜面未知应对称：P(GhostA)=%.4f≈P(GhostB)=%.4f", g.PGhostA(), g.PGhostB())
	}
	t.Logf("GH4 ✓ 对称恰一 ghost：PExactlyOne=%.4f PGG=%.4f PGhostA=%.4f≈PGhostB=%.4f",
		g.PExactlyOneGhost(), g.p[pGG], g.PGhostA(), g.PGhostB())
}
