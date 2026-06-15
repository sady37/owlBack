package adapter

// realness_test.go — lean-extract TrackArchive 译 Displaced 验收（DBN-Zone-Room §G：出生位原语单源）：
//   RA1 出生 + 位移闸：首帧 birth、未位移 Displaced=false；偏离出生位 >MoveCm → Displaced=true（锁 Real 命脉）。
//   RA4 断流不重生（安全红线）：稀疏/远帧绝不重生——track 消失（摔倒静止降功率）最危险态，出生位须跨消失存活。

import "testing"

func TestRA1BirthAndDisplaceGate(t *testing.T) {
	a := NewTrackArchive(DefaultRealnessParams())
	o, birth := a.Observe(RadarTrack{Online: true, X: 100, Y: 100})
	if !birth {
		t.Fatalf("首帧应为出生帧")
	}
	if o.Displaced {
		t.Errorf("出生帧未位移，Displaced 应 false")
	}
	// 出生位 +40cm（<MoveCm=50）→ 仍未位移。
	if o, _ := a.Observe(RadarTrack{Online: true, X: 140, Y: 100}); o.Displaced {
		t.Errorf("40cm<MoveCm，Displaced 应 false")
	}
	// 出生位 +80cm（>MoveCm）→ 位移（cd2b 床心→床沿）。
	if o, birth := a.Observe(RadarTrack{Online: true, X: 180, Y: 100}); !o.Displaced || birth {
		t.Errorf("80cm>MoveCm 应 Displaced=true、非重生，got Displaced=%v birth=%v", o.Displaced, birth)
	}
}

func TestRA4NoRebirth(t *testing.T) {
	a := NewTrackArchive(DefaultRealnessParams())
	a.Observe(RadarTrack{Online: true, X: 100, Y: 100}) // 出生位=(100,100)
	// 稀疏后远点（cd2b 摔倒末段哨兵/丢失帧）：绝不重生，仍以原出生位算位移 → Displaced 守住。
	o, birth := a.Observe(RadarTrack{Online: true, X: 400, Y: 100})
	if birth {
		t.Fatalf("不应重生（track 消失须保出生位/Real 信心）")
	}
	if !o.Displaced {
		t.Errorf("仍以原出生位(100,100)算，300cm 位移 Displaced 应 true（信心跨消失存活）")
	}
}
