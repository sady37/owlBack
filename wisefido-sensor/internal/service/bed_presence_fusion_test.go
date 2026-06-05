package service

import "testing"

// 2026-06-05 room total_people bug 回归：占用期 radar 漏静卧人 → bed_np 仍计；离床期不双计。
func TestOccupiedBedsInRoom(t *testing.T) {
	const now = int64(1_000_000_000)
	f := NewBedPresenceFusion()
	f.now = func() int64 { return now }
	room := "fd00:0:3:111:3:100::/88"
	bedA := "fd00:0:3:111:3:101::/96"
	bedB := "fd00:0:3:111:3:102::/96"

	if n := f.OccupiedBedsInRoom(room); n != 0 {
		t.Fatalf("empty: want 0 got %d", n)
	}
	// 占用期靶:sleepad InBed(radar room count 掉 0) → bed_np=1
	f.SetSleepad(bedA, true, now)
	if n := f.OccupiedBedsInRoom(room); n != 1 {
		t.Errorf("sleepad-only: want 1 got %d", n)
	}
	// 同床 radar 也 InBed → 仍 1(不双计)
	f.SetRadar(bedA, true, now)
	if n := f.OccupiedBedsInRoom(room); n != 1 {
		t.Errorf("both-same-bed: want 1 got %d", n)
	}
	// 第二张床 radar InBed → 2
	f.SetRadar(bedB, true, now)
	if n := f.OccupiedBedsInRoom(room); n != 2 {
		t.Errorf("two beds: want 2 got %d", n)
	}
	// 离床靶:bedA 两源都 LeftBed → 只剩 bedB → 1
	f.SetSleepad(bedA, false, now)
	f.SetRadar(bedA, false, now)
	if n := f.OccupiedBedsInRoom(room); n != 1 {
		t.Errorf("after A left: want 1 got %d", n)
	}
	// stale(超 10min 未更新) → 不算占用
	f.SetRadar(bedB, true, now-bedPresenceFreshMs-1)
	if n := f.OccupiedBedsInRoom(room); n != 0 {
		t.Errorf("stale: want 0 got %d", n)
	}
}
