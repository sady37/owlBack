package spatial

import (
	"net/netip"
	"testing"
)

var (
	unitP     = netip.MustParsePrefix("fd00:0:0:0:1::/80")
	roomP     = netip.MustParsePrefix("fd00:0:0:0:1:300::/88")
	bedP      = netip.MustParsePrefix("fd00:0:0:0:1:364::/96")
	radarP    = netip.MustParsePrefix("fd00:0:0:0:1:364:0:a1/128") // 注：放进 bed1 前缀仅为构造，实际房绑雷达走几何
	sleepadP  = netip.MustParsePrefix("fd00:0:0:0:1:364:0:b2/128")
	bed2P     = netip.MustParsePrefix("fd00:0:0:0:1:399::/96")
	sleepad2P = netip.MustParsePrefix("fd00:0:0:0:1:399:0:c3/128")
)

func obj(p netip.Prefix, k ObjectKind) Object { return Object{Prefix: p, Kind: k} }

func TestRelOf_FloatConfidence(t *testing.T) {
	// room ⊃ bed：contains 确定 = 1；反向 = 0（有向）
	if c := RelOf(obj(roomP, ObjRoom), obj(bedP, ObjBed), nil); c != ConfCertain {
		t.Errorf("room→bed 应 1，got %v", c)
	}
	if c := RelOf(obj(bedP, ObjBed), obj(roomP, ObjRoom), nil); c != ConfNone {
		t.Errorf("bed→room 应 0，got %v", c)
	}
	// onbed 双向都确定：sleepad∈bed=onbed=1，bed⊃sleepad=contains=1
	if c := RelOf(obj(sleepadP, ObjSleepad), obj(bedP, ObjBed), nil); c != ConfCertain {
		t.Errorf("sleepad→bed(onbed) 应 1，got %v", c)
	}
	if c := RelOf(obj(bedP, ObjBed), obj(sleepadP, ObjSleepad), nil); c != ConfCertain {
		t.Errorf("bed→sleepad(contains) 应 1，got %v", c)
	}
	// covers 候选：几何谓词返回 0.5
	covers := func(r, b netip.Prefix) float32 { return 0.5 }
	if c := RelOf(obj(radarP, ObjRadar), obj(bed2P, ObjBed), covers); c != 0.5 {
		t.Errorf("radar→bed 候选应 0.5，got %v", c)
	}
}

func TestMatrix_SameBedConfPropagates(t *testing.T) {
	// 你的场景：room 内 bed1、bed2；sleepad 绑 bed1；radar 房绑（前缀不进任何 bed），
	// 几何说 radar 候选盖 bed1=0.5、bed2=0.5。
	roomRadar := netip.MustParsePrefix("fd00:0:0:0:1:300:0:a1/128") // 在 room /88 内，但不在任何 bed /96 内
	objs := []Object{
		obj(roomP, ObjRoom),
		obj(bedP, ObjBed),
		obj(bed2P, ObjBed),
		obj(roomRadar, ObjRadar),
		obj(sleepadP, ObjSleepad),
	}
	covers := func(r, b netip.Prefix) float32 {
		if r == roomRadar && (b == bedP || b == bed2P) {
			return 0.5 // 候选：分不清 A/B
		}
		return 0
	}
	mx := Build(unitP, objs, covers)

	// 对角自指 = 1
	if mx.Conf(roomRadar, roomRadar) != ConfCertain {
		t.Error("对角应 1")
	}
	// 候选 covers 都是 0.5
	if mx.Conf(roomRadar, bedP) != 0.5 || mx.Conf(roomRadar, bed2P) != 0.5 {
		t.Errorf("radar 候选盖两床应各 0.5，got %v %v", mx.Conf(roomRadar, bedP), mx.Conf(roomRadar, bed2P))
	}
	// onbed 确定
	if mx.Conf(sleepadP, bedP) != ConfCertain {
		t.Error("sleepad→bed1 应 1")
	}
	// 派生 SameBed：min(covers=0.5, onbed=1)=0.5，候选不确定性传播过来；对称
	if c := mx.SameBedConf(roomRadar, sleepadP); c != 0.5 {
		t.Errorf("radar↔sleepad 学习前同床置信应 0.5（候选传播），got %v", c)
	}
	if mx.SameBedConf(sleepadP, roomRadar) != 0.5 {
		t.Error("SameBed 应对称")
	}
	// FuseGroup(bed1) = {radar@0.5, sleepad@1}
	fg := mx.FuseGroup(bedP)
	if len(fg) != 2 {
		t.Fatalf("FuseGroup(bed1) 应含 radar+sleepad，got %v", fg)
	}
	var sawRadar, sawSleepad bool
	for _, s := range fg {
		if s.Prefix == roomRadar && s.Conf == 0.5 {
			sawRadar = true
		}
		if s.Prefix == sleepadP && s.Conf == ConfCertain {
			sawSleepad = true
		}
	}
	if !sawRadar || !sawSleepad {
		t.Errorf("FuseGroup 权重不对：radar 应 0.5 / sleepad 应 1，got %v", fg)
	}
}
