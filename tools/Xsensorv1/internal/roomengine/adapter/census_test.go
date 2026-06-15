package adapter

import "testing"

// census_test.go — 房内多 track N_r 验收（§G 主职 + §G六 ghost 仅 track==2）：
//   CN1 单 track：N_r=1（永发，不进 ghost）。
//   CN2 真人+影子（共动 + 反射几何）：影子判 Mirror 排除 → N_r=1（防当 2 人降级独处真人风险）。
//   CN3 两独立真人（异步运动、无反射几何）：都 Real → N_r=2。
//   CN4 3 track（含一条反射标注）：ghost 不处理 → 不排 mirror → N_r=3（作用域边界，§G六）。

func runTrackCensus(frames int, step func(f int) []TrackObs) *TrackCensus {
	c := NewTrackCensus(DefaultTrackCensusParams())
	for f := 0; f < frames; f++ {
		c.Update(step(f))
	}
	return c
}

func TestCN1SingleTrack(t *testing.T) {
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{{X: 100 + f*30, Y: 100}} // 一个人走动
	})
	if c.Nr() != 1 {
		t.Errorf("单 track 应 N_r=1（永发，不进 ghost），got %d", c.Nr())
	}
}

func TestCN2RealPlusMirror(t *testing.T) {
	// 真人在 y=100 走、影子在 y=400 同步同速走（co-existence ρ 高）；影子带反射几何标注。
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{
			{X: 100 + f*30, Y: 100, IsReflection: false}, // 真人
			{X: 100 + f*30, Y: 400, IsReflection: true},  // 影子：同步 + 反射几何
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
			{X: 100 + f*30, Y: 100},    // 甲匀速走
			{X: 600, Y: 400 + (f%2)*5}, // 乙基本不动/微动（异步）
		}
	})
	if c.Nr() != 2 {
		t.Errorf("两独立真人应 N_r=2（无反射几何不判 mirror），got %d", c.Nr())
	}
}

func TestCN5ArtifactGhostExcluded(t *testing.T) {
	// 单 track 持续超速（200cm/帧 > SpeedCeil=100，< AssocCm=250 仍关联）→ 伪迹 ghost 累积 → 排除出 N_r。
	c := runTrackCensus(12, func(f int) []TrackObs {
		return []TrackObs{{X: 100 + f*200, Y: 100}}
	})
	if c.Nr() != 0 {
		t.Errorf("持续超速伪迹应判 ghost 排除 → N_r=0（§G七桶一），got %d", c.Nr())
	}
}

func TestCN4ThreeTracksGhostNotProcessed(t *testing.T) {
	// 3 track（含一条反射标注）：ghost 仅 track==2 处理、3+ 不处理 → mirror 不排 → N_r=3。
	c := runTrackCensus(10, func(f int) []TrackObs {
		return []TrackObs{
			{X: 100 + f*30, Y: 100},
			{X: 100 + f*30, Y: 400, IsReflection: true},
			{X: 800, Y: 800},
		}
	})
	if c.Nr() != 3 {
		t.Errorf("3 track ghost 不处理应 N_r=3（作用域边界 §G六），got %d", c.Nr())
	}
}
