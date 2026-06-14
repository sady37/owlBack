package zoneengine

import (
	"testing"

	"go.uber.org/zap"
)

type mockZ struct {
	z  int
	ok bool
}

func (m mockZ) GetZ(string) (int, bool) { return m.z, m.ok }

type mockBed struct{ n int }

func (m mockBed) OccupiedBedsInRoom(string) int { return m.n }

type mockReal struct{ n int }

func (m mockReal) RealPeopleInRoom(string) int { return m.n }

// realLookup(非 ghost track 数)优先于 firmware Z；ghost 撑高 firmware np 时取 realCount 去掉 ghost。
func TestDedupedTotalPeople_RealLookupExcludesGhost(t *testing.T) {
	const room = "fd00:0:3:111:3:100::/88"
	// cycle B 靶:firmware Z=2(real+ghost),realCount=1(去 ghost),bed=0 → 1
	p := NewStreamPublisher(nil, zap.NewNop())
	p.SetRoomDedup(mockZ{2, true}, mockBed{0})
	p.SetRealPeopleLookup(mockReal{1})
	if got := p.dedupedTotalPeople(room, 0); got != 1 {
		t.Errorf("ghost excluded: got %d want 1 (firmware Z=2 含 ghost, realCount=1)", got)
	}
	// realCount<0(房未注册)→ 退回 firmware Z
	p2 := NewStreamPublisher(nil, zap.NewNop())
	p2.SetRoomDedup(mockZ{2, true}, mockBed{0})
	p2.SetRealPeopleLookup(mockReal{-1})
	if got := p2.dedupedTotalPeople(room, 0); got != 2 {
		t.Errorf("unregistered room → firmware Z: got %d want 2", got)
	}
}

// 2026-06-05 room total_people bug 回归:max(radar_np, bed_np),非 Z+extras 相加。
// cycle A 实测两个翻车相位 + 正常 + 多人 + 无 cache fallback。
func TestDedupedTotalPeople_Max(t *testing.T) {
	const room = "fd00:0:3:111:3:100::/88"
	cases := []struct {
		name     string
		z        int
		zok      bool
		bed      int
		fallback int
		want     int
	}{
		{"occupancy: radar lost static, bed sees", 0, true, 1, 0, 1},       // max(0,1)=1（旧版早退→0）
		{"departure: radar tracks walkout, bed cleared", 1, true, 0, 1, 1}, // max(1,0)=1（旧版 1+1=2）
		{"normal: both see bed occupant", 1, true, 1, 1, 1},                // max(1,1)=1（不加 2）
		{"two people", 2, true, 1, 2, 2},
		{"no Z cache → fallback as radar_np", 0, false, 0, 3, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := NewStreamPublisher(nil, zap.NewNop())
			p.SetRoomDedup(mockZ{c.z, c.zok}, mockBed{c.bed})
			if got := p.dedupedTotalPeople(room, c.fallback); got != c.want {
				t.Errorf("got %d want %d", got, c.want)
			}
		})
	}
}
