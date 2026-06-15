package belief

import "testing"

// neighbor_test.go — 跨房 neighbor 隐轴涌现验收（§A ρ_xroom；§38 五领域规则补全后）：
//   NV1 fresh 有向 hand-off（峰区 Δ）+ 单住户 → ρ 高 → Blind→Fallen 整流入 Left（人挪去邻房）
//   NV2 无 hand-off → ρ=0 → Blind 行不变（保 lost-fall 安全默认，铁律）
//   NV3 stale（Δ 超窗）→ w^dir=0 → ρ=0 → 不抑制
//   NV4 有向性：反向（宿房先于源房，|Δ|>jitter）→ w^dir=0 → ρ=0（区别 ghost 对称）
//   NV5 sole-resident 连续衰减：多住户 ρ < 单住户 ρ 但 >0（弱不归零）
//   NV6【规则③ band-pass】同 tick(Δ≈0) 压低但非零 < 峰(1-5s) > 衰减(近窗尾)；同 tick ρ < 峰 ρ
//   NV7【规则④】W 随公共度收小（公共 unit hand-off 窗 < 私有 suite）
//   NV8【规则⑤】D 随覆盖差放长，锚静止门限+余量（覆盖差 = 静止8min+余量2min = 10min 定值）

const nveps = 1e-9

func freshSib() SiblingHandoff {
	return SiblingHandoff{ArrivalDeltaMs: 2_500, CAttr: 0.9, GainedReal: 0.9} // 峰区(~1-5s)，sleepad，守恒重现高
}

// NV1：fresh 有向 hand-off → ρ 高 → Blind 行 F 整流入 L。
func TestNeighborHandoffRoutesToLeft(t *testing.T) {
	p := DefaultNeighborParams()
	rho := RhoXroom(p, 1, []SiblingHandoff{freshSib()})
	if rho < 0.5 {
		t.Fatalf("fresh 有向 hand-off（峰区）+ 单住户 ρ=%.4f 应高", rho)
	}
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
	stale := SiblingHandoff{ArrivalDeltaMs: 120_000, CAttr: 0.9, GainedReal: 0.9} // 120s > W=60s
	if rho := RhoXroom(p, 1, []SiblingHandoff{stale}); rho != 0 {
		t.Errorf("stale（超窗）ρ=%.4f 应=0（stale 证不了此刻在哪，不抑制 fall）", rho)
	}
	t.Logf("NV3 ✓ stale 超窗 ρ=0：不抑制 lost-fall（铁律 partial_monitoring）")
}

// NV4【有向性，区别 ghost 对称】：反向（宿房先于源房，|Δ|>jitter）→ ρ=0。
func TestNeighborDirectedAsymmetry(t *testing.T) {
	p := DefaultNeighborParams()
	forward := RhoXroom(p, 1, []SiblingHandoff{{ArrivalDeltaMs: 2_500, CAttr: 0.9, GainedReal: 0.9}})  // 先走后到(峰区)
	reverse := RhoXroom(p, 1, []SiblingHandoff{{ArrivalDeltaMs: -10_000, CAttr: 0.9, GainedReal: 0.9}}) // 宿房先(|Δ|>J=5s)
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

// NV6【规则③ band-pass】：同 tick(Δ≈0) 压低但非零 < 峰(1-5s) > 衰减(近窗尾)。
func TestNeighborBandpassKernel(t *testing.T) {
	p := DefaultNeighborParams()
	w0 := wDir(p, 0)       // 同 tick：人没法瞬移 = 可疑（压低但非零）
	wPeak := wDir(p, 2_500) // 1-5s：正常走过去 = 峰
	wLate := wDir(p, 55_000) // 近窗尾：越久越可能别的事 = 衰减
	if w0 <= 0 {
		t.Errorf("同 tick w(0)=%.4f 应 >0（压低但非零，不完全排除时钟抖动）", w0)
	}
	if w0 >= wPeak {
		t.Errorf("同 tick w(0)=%.4f 应 < 峰 w(2.5s)=%.4f（band-pass 先升，旧 Δ=0 峰是错的）", w0, wPeak)
	}
	if wPeak <= wLate {
		t.Errorf("峰 w(2.5s)=%.4f 应 > 衰减 w(55s)=%.4f（band-pass 后降）", wPeak, wLate)
	}
	// ρ 层同样：同 tick ρ 应被压低于峰区 ρ。
	rho0 := RhoXroom(p, 1, []SiblingHandoff{{ArrivalDeltaMs: 0, CAttr: 0.9, GainedReal: 0.9}})
	rhoPeak := RhoXroom(p, 1, []SiblingHandoff{freshSib()})
	if rho0 >= rhoPeak {
		t.Errorf("同 tick ρ=%.4f 应 < 峰 ρ=%.4f（同 tick 两房不相关→少抑制→倾向当摔，安全）", rho0, rhoPeak)
	}
	t.Logf("NV6 ✓ band-pass：w(0)=%.3f < w(2.5s)=%.3f > w(55s)=%.4f；ρ(0)=%.3f < ρ(2.5s)=%.3f", w0, wPeak, wLate, rho0, rhoPeak)
}

// NV7【规则④】：W 随公共度收小（公共/陌生人流大 → 窗小，防陌生人偶合误判 hand-off）。
func TestNeighborWindowShrinksPublic(t *testing.T) {
	base := int64(60_000)
	priv := HandoffWindowFor(base, 0)  // 私有 suite
	pub := HandoffWindowFor(base, 1)   // 楼层公共
	if priv != base {
		t.Errorf("私有 suite W=%d 应=base %d", priv, base)
	}
	if pub >= priv {
		t.Errorf("公共 unit W=%d 应 < 私有 W=%d（防陌生人偶合）", pub, priv)
	}
	t.Logf("NV7 ✓ W 随公共度收小：私有 %dms → 公共 %dms", priv, pub)
}

// NV8【规则⑤】：D 随覆盖差放长，锚静止门限+余量（覆盖差 → 10min 定值）。
func TestNeighborDelayWindowAnchored(t *testing.T) {
	stillVanish := int64(480_000) // 静止消失门限 8min
	margin := int64(120_000)      // 余量 2min
	good := DelayWindowFor(stillVanish, margin, 1) // 全覆盖 → 接近门限
	poor := DelayWindowFor(stillVanish, margin, 0) // 大量盲区 → 门限+满余量
	if good != stillVanish {
		t.Errorf("全覆盖 D=%d 应=静止门限 %d（无需久等）", good, stillVanish)
	}
	if poor != stillVanish+margin {
		t.Errorf("覆盖差 D=%d 应=静止门限8min+余量2min=10min(%d)（neighbor 唯一裁决，等过门限+缓冲）", poor, stillVanish+margin)
	}
	if poor <= good {
		t.Errorf("覆盖差 D=%d 应 > 覆盖好 D=%d（规则⑤放长）", poor, good)
	}
	t.Logf("NV8 ✓ D 锚静止门限+余量：覆盖好 %dms(8min) → 覆盖差 %dms(10min 定值)", good, poor)
}
