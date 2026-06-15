package belief

import "testing"

// neighbor_test.go — 跨房 neighbor 隐轴涌现验收（§A ρ_xroom 有向门控）：
//   NV1 fresh 有向 hand-off + 单住户 → ρ 高 → Blind→Fallen 整流入 Left（人挪去邻房）
//   NV2 无 hand-off → ρ=0 → Blind 行不变（保 lost-fall 安全默认，铁律）
//   NV3 stale（Δ 超窗）→ w^dir=0 → ρ=0 → 不抑制
//   NV4 有向性：反向（宿房先于源房，|Δ|>jitter）→ w^dir=0 → ρ=0（区别 ghost 对称）
//   NV5 sole-resident 连续衰减：多住户 ρ < 单住户 ρ 但 >0（弱不归零）

const nveps = 1e-9

func freshSib() SiblingHandoff {
	return SiblingHandoff{ArrivalDeltaMs: 10_000, CAttr: 0.9, PRealPresent: 0.9} // 先走后到 10s，sleepad，去ghost占用高
}

// NV1：fresh 有向 hand-off → ρ 高 → Blind 行 F 整流入 L。
func TestNeighborHandoffRoutesToLeft(t *testing.T) {
	p := DefaultNeighborParams()
	rho := RhoXroom(p, 1, []SiblingHandoff{freshSib()})
	if rho < 0.5 {
		t.Fatalf("fresh 有向 hand-off + 单住户 ρ=%.4f 应高", rho)
	}
	// Blind 行（假设 →Fallen 倾向 0.5、→Left 0.1），整流后 F 降、L 升。
	var row [numStates]float64
	row[SFallen] = 0.5
	row[SLeft] = 0.1
	g := GateBlindRow(row, rho)
	if g[SFallen] >= row[SFallen] || g[SLeft] <= row[SLeft] {
		t.Errorf("应把 F 整流入 L：F %.3f→%.3f，L %.3f→%.3f", row[SFallen], g[SFallen], row[SLeft], g[SLeft])
	}
	if d := (g[SFallen] + g[SLeft]) - (row[SFallen] + row[SLeft]); d > nveps || d < -nveps {
		t.Errorf("F+L 行和应守恒，差 %.6f", d)
	}
	t.Logf("NV1 ✓ fresh hand-off ρ=%.3f：F %.3f→%.3f、L %.3f→%.3f（人挪去邻房）", rho, row[SFallen], g[SFallen], row[SLeft], g[SLeft])
}

// NV2：无 hand-off → ρ=0 → Blind 行不变（保 lost-fall）。
func TestNeighborNoHandoffPreservesFall(t *testing.T) {
	p := DefaultNeighborParams()
	rho := RhoXroom(p, 1, nil)
	if rho != 0 {
		t.Fatalf("无 hand-off ρ=%.4f 应=0", rho)
	}
	var row [numStates]float64
	row[SFallen] = 0.5
	g := GateBlindRow(row, rho)
	if g[SFallen] != row[SFallen] {
		t.Errorf("无 hand-off Blind 行应不变（保 lost-fall 安全默认），F %.3f→%.3f", row[SFallen], g[SFallen])
	}
	t.Logf("NV2 ✓ 无 hand-off ρ=0：Blind→Fallen 保留（lost-fall 不抑制）")
}

// NV3：stale（Δ 超窗）→ ρ=0 → 不抑制。
func TestNeighborStaleNoSuppress(t *testing.T) {
	p := DefaultNeighborParams()
	stale := SiblingHandoff{ArrivalDeltaMs: 120_000, CAttr: 0.9, PRealPresent: 0.9} // 120s > W=60s
	if rho := RhoXroom(p, 1, []SiblingHandoff{stale}); rho != 0 {
		t.Errorf("stale（超窗）ρ=%.4f 应=0（stale 证不了此刻在哪，不抑制 fall）", rho)
	}
	t.Logf("NV3 ✓ stale 超窗 ρ=0：不抑制 lost-fall（铁律 partial_monitoring）")
}

// NV4【有向性，区别 ghost 对称】：反向（宿房先于源房，|Δ|>jitter）→ ρ=0。
func TestNeighborDirectedAsymmetry(t *testing.T) {
	p := DefaultNeighborParams()
	forward := RhoXroom(p, 1, []SiblingHandoff{{ArrivalDeltaMs: 10_000, CAttr: 0.9, PRealPresent: 0.9}})  // 先走后到
	reverse := RhoXroom(p, 1, []SiblingHandoff{{ArrivalDeltaMs: -10_000, CAttr: 0.9, PRealPresent: 0.9}}) // 宿房先(|Δ|>J=5s)
	if forward < 0.5 {
		t.Errorf("正向 hand-off ρ=%.4f 应高", forward)
	}
	if reverse != 0 {
		t.Errorf("反向（宿房先于源房，超 jitter）ρ=%.4f 应=0（有向，非对称——区别 ghost 共存）", reverse)
	}
	t.Logf("NV4 ✓ 有向：正向 ρ=%.3f vs 反向 ρ=%.3f（不照搬 ghost 对称核）", forward, reverse)
}

// NV5：sole-resident 连续衰减——多住户 ρ < 单住户但 >0。
func TestNeighborSoleResidentDecay(t *testing.T) {
	p := DefaultNeighborParams()
	rho1 := RhoXroom(p, 1, []SiblingHandoff{freshSib()})
	rho3 := RhoXroom(p, 3, []SiblingHandoff{freshSib()})
	if rho3 >= rho1 {
		t.Errorf("多住户 ρ(rc=3)=%.4f 应 < 单住户 ρ(rc=1)=%.4f", rho3, rho1)
	}
	if rho3 <= 0 {
		t.Errorf("多住户 ρ(rc=3)=%.4f 应 >0（弱不归零；进 §8 C_FN 折扣漏报代价不归零）", rho3)
	}
	t.Logf("NV5 ✓ sole-resident 连续衰减：rc=1 ρ=%.3f → rc=3 ρ=%.3f（弱不归零）", rho1, rho3)
}
