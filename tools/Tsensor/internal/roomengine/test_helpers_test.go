package roomengine

// 共享测试 helper（原在已删的 fall_verify_test.go;迁出供 track_manager_clear_device_test 等复用）。

func newTestGrid() *RoomGrid {
	g := NewRoomGrid(500, 500, 10)
	for i := range g.Cells {
		g.Cells[i].InRoom = true
		g.Cells[i].InFOV = true
		g.Cells[i].EdgeDist = 100
	}
	return g
}

func newTestTM() (*TrackManager, *RoomGrid) {
	g := newTestGrid()
	return NewTrackManager("test-room", g), g
}
