package replay

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

const (
	cd2bCase = "../../../../../doc/cases/case-cd2b-0604-16141631"
	cd2bUID  = "9D8A32A1CD2B"
)

// HR-1：解析忠实。window.json→FrameInput 无丢字段；NowMs 单调；各切片长度=numBeds。
func TestHR1ParseFidelity(t *testing.T) {
	tl, err := BuildTimeline(cd2bCase, cd2bUID)
	if err != nil {
		t.Fatalf("BuildTimeline err: %v", err)
	}
	if len(tl) == 0 {
		t.Fatal("时间线空（解析失败或 radar UID 不匹配）")
	}
	nb := len(tl[0].Beds)
	if nb == 0 {
		t.Fatal("床矩形为 0（layout 解析失败）")
	}
	var last int64
	for i, fi := range tl {
		if fi.NowMs < last {
			t.Errorf("帧 %d NowMs=%d < 前 %d（非单调）", i, fi.NowMs, last)
		}
		last = fi.NowMs
		if len(fi.Sleepads) != nb || len(fi.Covers) != nb || len(fi.Onbed) != nb || len(fi.Overlap) != nb {
			t.Fatalf("帧 %d 切片长度 ≠ numBeds=%d", i, nb)
		}
	}
	t.Logf("HR-1 ✓ 解析 %d 帧，numBeds=%d，床矩形=%+v，NowMs 单调", len(tl), nb, tl[0].Beds)
}

// HR-2/AC-拆3（§32 框架接触轴）：cd2b 经 harness 端到端（sleepad 全程在线、零 floor-strip、零 TTL）
// → 经 LeftBed→B vac→Ψ 涌现 SFallen → 独处 fire。**且 fire 在 LeftBed 之后**（人离床才摔），
// 在床段（sleepad InBed）不误火。证框架接触轴解 cd2b，无几何 δ、无离线补丁。
func TestHR2Cd2bContactAxis(t *testing.T) {
	tl, err := BuildTimeline(cd2bCase, cd2bUID)
	if err != nil {
		t.Fatalf("BuildTimeline err: %v", err)
	}
	frames, finalS := Run(tl)
	start := tl[0].NowMs

	// LeftBed 起点（sleepad 在线报离床的首帧）。
	var leftBedMs int64
	for _, fi := range tl {
		if len(fi.Sleepads) > 0 && fi.Sleepads[0].Reading == belief.BedLeftBed {
			leftBedMs = fi.NowMs
			break
		}
	}
	// 首个 fire 帧 + 在床段(LeftBed 前)是否误火。
	var fired bool
	var fireAtMs int64
	var inBedFalseFire bool
	for _, fr := range frames {
		if fr.Decision.Fire && !fired {
			fired = true
			fireAtMs = fr.Probe.Ts
		}
		if fr.Decision.Fire && leftBedMs > 0 && fr.Probe.Ts < leftBedMs {
			inBedFalseFire = true // LeftBed 之前 fire = 在床误火
		}
	}

	t.Logf("HR-2 实测：帧数=%d finalP(Fallen)=%.4f LeftBed@+%.0fs fire=%v@+%.0fs",
		len(frames), finalS[belief.SFallen], float64(leftBedMs-start)/1000, fired, float64(fireAtMs-start)/1000)

	if !fired {
		t.Errorf("AC-拆3：cd2b 应经接触轴 fire（LeftBed→B vac→Ψ→SFallen），未 fire")
	}
	if inBedFalseFire {
		t.Errorf("在床段（LeftBed 前）误火——sleepad InBed 时 B occ 应经 Ψ ε_art 压住 SFallen")
	}
	if fired && leftBedMs > 0 && fireAtMs < leftBedMs {
		t.Errorf("fire@+%.0fs 早于 LeftBed@+%.0fs（应人离床后才摔）", float64(fireAtMs-start)/1000, float64(leftBedMs-start)/1000)
	}
}
