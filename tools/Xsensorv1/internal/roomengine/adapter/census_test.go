package adapter

import "testing"

// census_test.go — 房内多 track N_r 验收（§G 主职 + §G六 ghost 仅 track==2）：
//   CN1 单 track：N_r=1（永发，孤轨永 Real）。
//   CN2 真人+影子（同步 + 墙外反射几何）：影子判 Mirror 排除 → N_r=1（防当 2 人降级独处真人风险）。
//   CN3 两独立真人（异步运动、无反射几何）：都 Real → N_r=2。
//   CN4 3 track（含一条反射几何）：ghost 不处理(Coexist=0) → 不排 mirror → N_r=3（作用域边界，§G六）。
//   桶二 BK-A/B/D/E：reflectSep 镜像几何（§69 五柱，返裕度 cm）。

func tob(x, y int) TrackObs {
	return TrackObs{RadarTrack: RadarTrack{Online: true, X: x, Y: y}}
}

// 桶二测试几何：雷达在房中 (100,250)；镜面墙 = 横矩形 y∈[340,360] 跨 x∈[0,700]。
//
//	影子在 y=400（墙外，radar→影子穿墙、最近交点→影子≥30）→ reflectSep>0。
//	真人在 y=100（与 radar 同侧，radar→真人不穿墙）→ 0。
var testRadar = Point{X: 100, Y: 250}
var testWalls = []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}

func runTrackCensus(frames int, radar Point, walls []Rect, step func(f int) []TrackObs) *TrackCensus {
	c := NewTrackCensus(DefaultTrackCensusParams())
	for f := 0; f < frames; f++ {
		c.Update(int64(f+1)*1000, step(f), radar, walls, nil) // 1Hz；无 enter 区(entrances=nil → D=-1)
	}
	return c
}

func TestCN1SingleTrack(t *testing.T) {
	c := runTrackCensus(10, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{tob(100+f*30, 100)} // 一个人走动
	})
	if c.Nr() != 1 {
		t.Errorf("单 track 应 N_r=1（孤轨永 Real），got %d", c.Nr())
	}
}

func TestCN2RealPlusMirror(t *testing.T) {
	// 真人在 y=100 走、影子在 y=400 同步同速走（co-existence ρ 高）；墙外反射几何由 census 算（桶二）。
	c := runTrackCensus(12, testRadar, testWalls, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100), // 真人（与 radar 同侧 → 不穿墙 → 非反射）
			tob(100+f*30, 400), // 影子：同步 + 墙外反射（radar→影子穿墙）→ Mirror
		}
	})
	if c.Nr() != 1 {
		t.Errorf("真人+影子应 N_r=1（影子墙外反射判 Mirror 排除），got %d", c.Nr())
	}
}

func TestCN3TwoIndependent(t *testing.T) {
	// 两个独立真人：异步运动（速度不同步），无墙 → 无 mirror 证据 → 都 Real。
	c := runTrackCensus(10, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100),    // 甲匀速走
			tob(600, 400+(f%2)*5), // 乙基本不动/微动（异步）
		}
	})
	if c.Nr() != 2 {
		t.Errorf("两独立真人应 N_r=2（无反射几何/异步不判 mirror），got %d", c.Nr())
	}
}

func TestCN4ThreeTracksGhostNotProcessed(t *testing.T) {
	// 3 track（含一条几何反射）：ghost 仅 track==2 处理、3+ Coexist=0 不处理 → mirror 不排 → N_r=3。
	c := runTrackCensus(10, testRadar, testWalls, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100),
			tob(100+f*30, 400),
			tob(800, 800),
		}
	})
	if c.Nr() != 3 {
		t.Errorf("3 track ghost 不处理应 N_r=3（作用域边界 §G六），got %d", c.Nr())
	}
}

// ===== 桶二镜像几何验收（§69 五柱；reflectSep 返裕度 cm，>0=反射）=====

// BK-A 柱A 几何正确（§69 柱A）：墙外+连线穿墙+交点≥30cm → sep>0；墙内/不穿墙/<30cm → 0。
func TestBucket2GeomCorrectness(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}
	cases := []struct {
		name string
		gx   int
		gy   int
		want bool
	}{
		{"墙外+穿墙+交点远(影子)", 100, 400, true},   // radar→(100,400) 穿墙，最近交点(100,360)→ghost=40≥30
		{"墙内侧不穿墙(真人)", 100, 100, false},     // radar 同侧，连线不穿墙
		{"墙内(在墙矩形里)", 100, 350, false},      // ghost 在墙内 → 非反射
		{"墙外但<30cm(贴墙噪声)", 100, 375, false}, // (100,375) 墙外，最近交点(100,360)→ghost=15<30 → 不判（FN-safe）
	}
	for _, tc := range cases {
		if got := reflectSep(tc.gx, tc.gy, radar, walls, 30) > 0; got != tc.want {
			t.Errorf("%s: reflectSep>0=%v want %v", tc.name, got, tc.want)
		}
	}
}

// BK-B 🔴 柱B FN-safe（§69 柱B）：墙内真人/不穿墙/<30cm 边缘 → 绝不误判镜像（0）。
func TestBucket2FNSafe(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}
	for _, p := range []Point{{X: 100, Y: 100}, {X: 300, Y: 50}, {X: 50, Y: 300}, {X: 650, Y: 330}} {
		if reflectSep(p.X, p.Y, radar, walls, 30) > 0 {
			t.Errorf("FN-safe 破: 真人 %v 被误判镜像（→排出 N_r 漏报）", p)
		}
	}
	if reflectSep(100, 375, radar, walls, 30) > 0 {
		t.Error("FN-safe: <30cm 边缘应偏 0（宁漏镜像不误真人）")
	}
}

// BK-D 柱D 多墙全局最近（§69 柱D）：连线穿多墙取距 ghost 最近交点判阈。
func TestBucket2MultiWallNearest(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	// 两道横墙：y=340-360 与 y=380-400；ghost 在 y=500（两墙都穿）。最近 ghost 交点在第二墙 y=400，
	//   →ghost=100≥30 → sep>0。
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}, {X1: 0, Y1: 380, X2: 700, Y2: 400}}
	if reflectSep(100, 500, radar, walls, 30) <= 0 {
		t.Error("多墙：应取全局最近交点判反射")
	}
	// ghost 紧贴第二墙外 (y=410)：最近交点 y=400→ghost=10<30 → 0（取的是最近墙非最远墙）。
	if reflectSep(100, 410, radar, walls, 30) > 0 {
		t.Error("多墙：最近交点<30cm 应 0（取最近非最远）")
	}
}

// BK-E 柱E 零回归（§69 柱E）：无墙配置 → 恒 0（不动既有 N_r 行为）。
func TestBucket2NoWallsAlwaysZero(t *testing.T) {
	if reflectSep(100, 400, Point{X: 100, Y: 250}, nil, 30) > 0 {
		t.Error("无墙配置应恒 0（零回归）")
	}
}
