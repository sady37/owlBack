package adapter

import "testing"

// census_test.go — 房内多 track N_r 验收（§G 主职 + §G六 ghost 仅 track==2）：
//   CN1 单 track：N_r=1（永发，不进 ghost）。
//   CN2 真人+影子（共动 + 反射几何）：影子判 Mirror 排除 → N_r=1（防当 2 人降级独处真人风险）。
//   CN3 两独立真人（异步运动、无反射几何）：都 Real → N_r=2。
//   CN4 3 track（含一条反射几何）：ghost 不处理 → 不排 mirror → N_r=3（作用域边界，§G六）。
//   CN5 单 track 持续超速：伪迹 ghost 累积 → 排除 → N_r=0。
//   桶二 BK-A~F：镜像几何（§69 五柱 + §70 时序柱F）。

func tob(x, y int) TrackObs {
	return TrackObs{RadarTrack: RadarTrack{Online: true, X: x, Y: y}}
}

// 桶二测试几何：雷达在房中 (100,250)；镜面墙 = 横矩形 y∈[340,360] 跨 x∈[0,700]。
//   影子在 y=400（墙外，radar→影子穿墙、最近交点(100,360)→影子=40≥30）→ IsReflection=true。
//   真人在 y=100（与 radar 同侧，radar→真人不穿墙）→ false。
var testRadar = Point{X: 100, Y: 250}
var testWalls = []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}

func runTrackCensus(frames int, radar Point, walls []Rect, step func(f int) []TrackObs) *TrackCensus {
	c := NewTrackCensus(DefaultTrackCensusParams())
	for f := 0; f < frames; f++ {
		c.Update(int64(f+1)*1000, step(f), radar, walls) // 1Hz：cm/帧 ≈ cm/s（CN5 200cm/s 仍超 SpeedCeil=100）
	}
	return c
}

func TestCN1SingleTrack(t *testing.T) {
	c := runTrackCensus(10, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{tob(100+f*30, 100)} // 一个人走动
	})
	if c.Nr() != 1 {
		t.Errorf("单 track 应 N_r=1（永发，不进 ghost），got %d", c.Nr())
	}
}

func TestCN2RealPlusMirror(t *testing.T) {
	// 真人在 y=100 走、影子在 y=400 同步同速走（co-existence ρ 高）；IsReflection 由 census 几何算（桶二）。
	c := runTrackCensus(10, testRadar, testWalls, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100), // 真人（与 radar 同侧 → 不穿墙 → 非反射）
			tob(100+f*30, 400), // 影子：同步 + 几何反射（墙外、radar→影子穿墙）
		}
	})
	if c.Nr() != 1 {
		t.Errorf("真人+影子应 N_r=1（影子几何判 Mirror 排除），got %d", c.Nr())
	}
}

func TestCN3TwoIndependent(t *testing.T) {
	// 两个独立真人：异步运动（速度不同步），无墙 → 不判 mirror。
	c := runTrackCensus(10, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100),    // 甲匀速走
			tob(600, 400+(f%2)*5), // 乙基本不动/微动（异步）
		}
	})
	if c.Nr() != 2 {
		t.Errorf("两独立真人应 N_r=2（无反射几何不判 mirror），got %d", c.Nr())
	}
}

func TestCN5SingleTrackArtifactExcludedFromNr(t *testing.T) {
	// 单 track 持续超速 → artifact 累积 → PReal<0.5 → 排除出 N_r（§61：artifact 照算照喂 N_r 排假人头，
	//   不加 track==2 计算门控）。N_r=0 只是「人数」口径；这条孤轨的**摔不被压**由 engine 消费门控保
	//   （无共存源→pFallReal=1 永发，见 engine EG6）——两件事分开：N_r 数人头 / pFallReal 压不压摔。
	c := runTrackCensus(12, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{tob(100+f*200, 100)}
	})
	if c.Nr() != 0 {
		t.Errorf("单 track 持续超速 artifact → 排除出 N_r=0（人数口径），got %d", c.Nr())
	}
}

func TestCN6TwoTracksArtifactExcluded(t *testing.T) {
	// track==2：真人正常走 + 一条持续超速 phantom（异步、非反射）→ 超速判运动伪迹 ghost → 排除 → N_r=1。
	c := runTrackCensus(12, Point{}, nil, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100),  // 真人 30cm/s
			tob(100+f*200, 500), // 超速 phantom 200cm/s > SpeedCeil=100（<AssocCm=250 不churn）
		}
	})
	if c.Nr() != 1 {
		t.Errorf("track==2 含一条超速伪迹应排除 → N_r=1（§G七桶一 at track==2），got %d", c.Nr())
	}
}

func TestCN4ThreeTracksGhostNotProcessed(t *testing.T) {
	// 3 track（含一条几何反射）：ghost 仅 track==2 处理、3+ 不处理 → mirror 不排 → N_r=3。
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

// ===== 桶二镜像几何验收（§69 五柱 + §70 柱F；纯几何函数级 + census 时序级）=====

// BK-A 柱A 几何正确（§69 柱A）：墙外+连线穿墙+交点≥30cm → true；墙内/不穿墙/<30cm → false。
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
		if got := reflectsAcrossWall(tc.gx, tc.gy, radar, walls, 30); got != tc.want {
			t.Errorf("%s: reflectsAcrossWall=%v want %v", tc.name, got, tc.want)
		}
	}
}

// BK-B 🔴 柱B FN-safe（§69 柱B）：墙内真人/不穿墙/<30cm 边缘 → 绝不误判镜像（false）。
func TestBucket2FNSafe(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}
	// 真人在房内各处（与 radar 同侧、不穿墙）→ 必须全 false（否则排出 N_r 漏报）。
	for _, p := range []Point{{X: 100, Y: 100}, {X: 300, Y: 50}, {X: 50, Y: 300}, {X: 650, Y: 330}} {
		if reflectsAcrossWall(p.X, p.Y, radar, walls, 30) {
			t.Errorf("FN-safe 破: 真人 %v 被误判镜像（→排出 N_r 漏报）", p)
		}
	}
	// <30cm 边缘宁可漏判（false），不可误判真人。
	if reflectsAcrossWall(100, 375, radar, walls, 30) {
		t.Error("FN-safe: <30cm 边缘应偏 false（宁漏镜像不误真人）")
	}
}

// BK-D 柱D 多墙全局最近（§69 柱D）：连线穿多墙取距 ghost 最近交点判阈。
func TestBucket2MultiWallNearest(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	// 两道横墙：y=340-360 与 y=380-400；ghost 在 y=500（两墙都穿）。最近 ghost 交点在第二墙 y=400，
	//   →ghost=100≥30 → true。
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}, {X1: 0, Y1: 380, X2: 700, Y2: 400}}
	if !reflectsAcrossWall(100, 500, radar, walls, 30) {
		t.Error("多墙：应取全局最近交点判 true")
	}
	// ghost 紧贴第二墙外 (y=410)：最近交点 y=400→ghost=10<30 → false（取的是最近墙非最远墙）。
	if reflectsAcrossWall(100, 410, radar, walls, 30) {
		t.Error("多墙：最近交点<30cm 应 false（取最近非最远）")
	}
}

// BK-E 柱E 零回归（§69 柱E）：无墙配置 → 恒 false（不动既有 N_r 行为）。
func TestBucket2NoWallsAlwaysFalse(t *testing.T) {
	if reflectsAcrossWall(100, 400, Point{X: 100, Y: 250}, nil, 30) {
		t.Error("无墙配置应恒 false（零回归）")
	}
}

// BK-F 柱F 时序（§70）：出生窗内每帧算 provisional（保护 movedFromBirth），过 ReflSettleMs 冻结、之后不重算。
func TestBucket2BirthWindowLockTiming(t *testing.T) {
	radar := Point{X: 100, Y: 250}
	walls := []Rect{{X1: 0, Y1: 340, X2: 700, Y2: 360}}
	c := NewTrackCensus(DefaultTrackCensusParams()) // ReflSettleMs=3000
	// 一条 track：出生在墙外反射位 (100,400)，settle 后(age≥3000)漂回墙内真人位 (100,100)。
	//   冻结值取 settle 帧几何（反射位 true），之后漂移不翻（锁定，抗噪声抖动）。
	c.Update(1000, []TrackObs{tob(100, 400)}, radar, walls) // 出生 ms=1000，provisional=true
	tr := c.tracks[1]
	if !tr.isRefl {
		t.Fatal("出生窗 provisional 应 true（墙外反射位）")
	}
	if tr.reflLocked {
		t.Error("未到 settle（age=0）不应冻结")
	}
	c.Update(4000, []TrackObs{tob(100, 400)}, radar, walls) // ms=4000，age=3000≥settle → 冻结 true
	if !tr.reflLocked || !tr.isRefl {
		t.Errorf("到 settle 应冻结 isRefl=true，got locked=%v isRefl=%v", tr.reflLocked, tr.isRefl)
	}
	c.Update(5000, []TrackObs{tob(100, 100)}, radar, walls) // 之后漂到真人位（几何 false）→ 锁定不翻
	if !tr.isRefl {
		t.Error("冻结后漂移不应翻 IsReflection（抗噪声抖动，§70）")
	}
}
