package adapter

import "testing"

// census_test.go — 房内多 track N_r 验收（§G 主职 + §G六 ghost 仅 track==2）：
//   CN1 单 track：N_r=1（永发，不进 ghost）。
//   CN2 真人+影子（共动 + 反射几何）：影子判 Mirror 排除 → N_r=1（防当 2 人降级独处真人风险）。
//   CN3 两独立真人（异步运动、无反射几何）：都 Real → N_r=2。
//   CN4 3 track（含一条反射标注）：ghost 不处理 → 不排 mirror → N_r=3（作用域边界，§G六）。
//   CN5 单 track 持续超速：伪迹 ghost 累积 → 排除 → N_r=0。

func tob(x, y int, refl bool) TrackObs {
	return TrackObs{RadarTrack: RadarTrack{Online: true, X: x, Y: y}, IsReflection: refl}
}

func runTrackCensus(frames int, step func(f int) []TrackObs) *TrackCensus {
	c := NewTrackCensus(DefaultTrackCensusParams())
	for f := 0; f < frames; f++ {
		c.Update(int64(f+1)*1000, step(f)) // 1Hz：cm/帧 ≈ cm/s（CN5 200cm/s 仍超 SpeedCeil=100）
	}
	return c
}

func TestCN1SingleTrack(t *testing.T) {
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{tob(100+f*30, 100, false)} // 一个人走动
	})
	if c.Nr() != 1 {
		t.Errorf("单 track 应 N_r=1（永发，不进 ghost），got %d", c.Nr())
	}
}

func TestCN2RealPlusMirror(t *testing.T) {
	// 真人在 y=100 走、影子在 y=400 同步同速走（co-existence ρ 高）；影子带反射几何标注。
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100, false), // 真人
			tob(100+f*30, 400, true),  // 影子：同步 + 反射几何
		}
	})
	if c.Nr() != 1 {
		t.Errorf("真人+影子应 N_r=1（影子判 Mirror 排除），got %d", c.Nr())
	}
}

func TestCN3TwoIndependent(t *testing.T) {
	// 两个独立真人：异步运动（速度不同步），都无反射几何 → 不判 mirror。
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100, false),    // 甲匀速走
			tob(600, 400+(f%2)*5, false), // 乙基本不动/微动（异步）
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
	c := runTrackCensus(12, func(f int) []TrackObs {
		return []TrackObs{tob(100+f*200, 100, false)}
	})
	if c.Nr() != 0 {
		t.Errorf("单 track 持续超速 artifact → 排除出 N_r=0（人数口径），got %d", c.Nr())
	}
}

func TestCN6TwoTracksArtifactExcluded(t *testing.T) {
	// track==2：真人正常走 + 一条持续超速 phantom（异步、非反射）→ 超速判运动伪迹 ghost → 排除 → N_r=1。
	c := runTrackCensus(12, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100, false),  // 真人 30cm/s
			tob(100+f*200, 500, false), // 超速 phantom 200cm/s > SpeedCeil=100（<AssocCm=250 不churn）
		}
	})
	if c.Nr() != 1 {
		t.Errorf("track==2 含一条超速伪迹应排除 → N_r=1（§G七桶一 at track==2），got %d", c.Nr())
	}
}

func TestCN4ThreeTracksGhostNotProcessed(t *testing.T) {
	// 3 track（含一条反射标注）：ghost 仅 track==2 处理、3+ 不处理 → mirror 不排 → N_r=3。
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{
			tob(100+f*30, 100, false),
			tob(100+f*30, 400, true),
			tob(800, 800, false),
		}
	})
	if c.Nr() != 3 {
		t.Errorf("3 track ghost 不处理应 N_r=3（作用域边界 §G六），got %d", c.Nr())
	}
}
