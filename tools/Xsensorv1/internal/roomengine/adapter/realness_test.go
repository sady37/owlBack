package adapter

import (
	"testing"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// realness_test.go — lean-extract TrackArchive 译 RealnessObs 验收（C §44 边界：原语单源、每字段有源）。
//   RA1 出生 + 位移闸：首帧 birth、未位移 Displaced=false；偏离出生位 >MoveCm → Displaced=true（RV4 命脉）。
//   RA2 AgeLongStatic：age≥LongStaticMs 且困出生点 → true；走开 → false。
//   RA3 CrossedStillPeriod：静止 run 满 StillPeriodMs → 单调 true。
//   RA4 断流不重生（安全红线）：稀疏/大 gap 帧绝不重生——track 消失（摔倒静止降功率）正是最危险态，
//        出生位/Real 信心必须跨消失存活，不被擦写成抑制。

func TestRA1BirthAndDisplaceGate(t *testing.T) {
	a := NewTrackArchive(DefaultRealnessParams())
	o, birth := a.Observe(RadarTrack{Online: true, X: 100, Y: 100}, 1000)
	if !birth {
		t.Fatalf("首帧应为出生帧")
	}
	if o.Displaced {
		t.Errorf("出生帧未位移，Displaced 应 false")
	}
	// 出生位 +40cm（<MoveCm=50）→ 仍未位移。
	if o, _ := a.Observe(RadarTrack{Online: true, X: 140, Y: 100}, 2000); o.Displaced {
		t.Errorf("40cm<MoveCm，Displaced 应 false")
	}
	// 出生位 +80cm（>MoveCm）→ 位移（cd2b 床心→床沿）。
	if o, birth := a.Observe(RadarTrack{Online: true, X: 180, Y: 100}, 3000); o.Displaced != true || birth {
		t.Errorf("80cm>MoveCm 应 Displaced=true、非重生，got Displaced=%v birth=%v", o.Displaced, birth)
	}
}

func TestRA2AgeLongStatic(t *testing.T) {
	p := DefaultRealnessParams()
	a := NewTrackArchive(p)
	// 连续帧（每 1s，<RebirthGapMs，不重生）困出生点附近，直到达 LongStaticMs。
	var o belief.RealnessObs
	for ms := int64(0); ms <= p.LongStaticMs; ms += 1000 {
		o, _ = a.Observe(RadarTrack{Online: true, X: 105, Y: 100}, ms)
		if ms < p.LongStaticMs && o.AgeLongStatic {
			t.Fatalf("age=%d 未达 LongStaticMs，AgeLongStatic 应 false", ms)
		}
	}
	if !o.AgeLongStatic {
		t.Errorf("age≥LongStaticMs 且困出生点，AgeLongStatic 应 true")
	}
	// 达龄后走开出生点（>MoveCm）：false（不再困住 → 非 static）。
	if o, _ := a.Observe(RadarTrack{Online: true, X: 300, Y: 100}, p.LongStaticMs+1000); o.AgeLongStatic {
		t.Errorf("走开出生点，AgeLongStatic 应 false")
	}
}

func TestRA3CrossedStillPeriod(t *testing.T) {
	p := DefaultRealnessParams()
	a := NewTrackArchive(p)
	var last bool
	for ms := int64(0); ms <= p.StillPeriodMs+5000; ms += 1000 {
		o, _ := a.Observe(RadarTrack{Online: true, X: 100, Y: 100}, ms) // 钉死一点（box=0）
		last = o.CrossedStillPeriod
	}
	if !last {
		t.Errorf("静止 run 满 StillPeriodMs，CrossedStillPeriod 应 true")
	}
}

func TestRA4NoRebirthOnGap(t *testing.T) {
	a := NewTrackArchive(DefaultRealnessParams())
	a.Observe(RadarTrack{Online: true, X: 100, Y: 100}, 1000) // 出生位=(100,100)
	// 大 gap（cd2b 摔倒末段稀疏帧，gap 数十秒）后远点：绝不重生，仍以原出生位算位移 → Displaced 守住。
	o, birth := a.Observe(RadarTrack{Online: true, X: 400, Y: 100}, 1000+60_000)
	if birth {
		t.Fatalf("大 gap 不应重生（track 消失须保 Real 信心）")
	}
	if !o.Displaced {
		t.Errorf("仍以原出生位(100,100)算，300cm 位移 Displaced 应 true（信心跨消失存活）")
	}
}
