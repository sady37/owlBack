package roomengine

import (
	"testing"

	"owl-common/radarutils"
)

// TestEncodeSnapshot_OmitsTrivialCells 验证稀疏存储：未触碰过的 cell 不写入 payload
func TestEncodeSnapshot_OmitsTrivialCells(t *testing.T) {
	g := NewRoomGrid(400, 500, 10) // 40 × 50 = 2000 cells
	// 只设一个 cell 有 counter
	g.Cells[100].RealDecay = 5
	g.Cells[100].ActiveType[ActiveIdxSit] = 200
	// 设一个 cell 有 SourceLearned Belief（无 counter）
	g.Cells[200].Belief[0] = BeliefState{Type: AreaSit, Confidence: 80, Source: SourceLearned}
	g.Cells[200].AreaType = AreaSit

	snap := EncodeSnapshot(g)

	if len(snap.Cells) != 2 {
		t.Fatalf("expected 2 sparse cells, got %d", len(snap.Cells))
	}
	if snap.Grid.W != g.Width || snap.Grid.H != g.Height {
		t.Errorf("grid dims wrong: w=%d h=%d (grid=%dx%d)", snap.Grid.W, snap.Grid.H, g.Width, g.Height)
	}
	if snap.SchemaVer != SnapshotSchemaVersion {
		t.Errorf("schema_v=%d expected %d", snap.SchemaVer, SnapshotSchemaVersion)
	}
}

// TestEncodeDecode_Roundtrip 验证 encode→marshal→unmarshal→decode 不丢字段
func TestEncodeDecode_Roundtrip(t *testing.T) {
	g1 := NewRoomGrid(400, 500, 10)
	// 准备多种 cell 状态
	g1.Cells[10] = Cell{
		InRoom:        true,
		InFOV:         true,
		RealDecay:     12,
		GhostDecay:    3,
		ActiveType:    [4]uint16{120, 5, 80, 0},
		TraverseCount: 45,
		LieRetract:    2,
		FlowX:         3,
		FlowY:         -1,
		DwellEMA:      150,
		Belief: [3]BeliefState{
			{Type: AreaSit, Confidence: 92, Source: SourceLearned},
			{Type: AreaSit, Confidence: 88, Source: SourceLearned},
			{Type: AreaActive, Confidence: 60, Source: SourceLearned},
		},
		AreaType: AreaSit,
	}
	g1.Cells[20] = Cell{
		InRoom: true,
		Belief: [3]BeliefState{
			{Type: AreaBed, Confidence: 99, Source: SourceHuman},
			{Type: AreaBed, Confidence: 99, Source: SourceHuman},
			{Type: AreaBed, Confidence: 99, Source: SourceHuman},
		},
		AreaType:          AreaBed,
		SleepadInBedCount: 7,
	}

	snap1 := EncodeSnapshot(g1)
	payload, _, err := MarshalSnapshot(snap1)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	snap2, err := UnmarshalSnapshot(payload)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 灌回到一个 fresh grid（geometry 已由 NewRoomGrid 设好）
	g2 := NewRoomGrid(400, 500, 10)
	if err := DecodeSnapshot(snap2, g2); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// 对比关键字段（注意：geometry 不持久化，所以 InRoom/InFOV 不验）
	c10 := &g2.Cells[10]
	if c10.RealDecay != 12 || c10.GhostDecay != 3 {
		t.Errorf("cell 10 RealDecay=%d GhostDecay=%d", c10.RealDecay, c10.GhostDecay)
	}
	if c10.ActiveType != [4]uint16{120, 5, 80, 0} {
		t.Errorf("cell 10 ActiveType=%v", c10.ActiveType)
	}
	if c10.TraverseCount != 45 {
		t.Errorf("cell 10 TraverseCount=%d", c10.TraverseCount)
	}
	if c10.Belief[0].Type != AreaSit || c10.Belief[0].Confidence != 92 {
		t.Errorf("cell 10 Belief[0]=%+v", c10.Belief[0])
	}
	if c10.AreaType != AreaSit {
		t.Errorf("cell 10 AreaType=%d", c10.AreaType)
	}

	c20 := &g2.Cells[20]
	if c20.SleepadInBedCount != 7 {
		t.Errorf("cell 20 SleepadInBedCount=%d", c20.SleepadInBedCount)
	}
	if c20.Belief[0].Source != SourceHuman {
		t.Errorf("cell 20 Belief[0].Source=%d expected SourceHuman", c20.Belief[0].Source)
	}
}

// TestDecodeSnapshot_PreservesHumanBelief 关键安全测试：
// 即使 snapshot 里某 cell 的 Belief 不是 Human，
// 如果 grid 在加载前已被 SetPrior 标为 Human（layout 后来扩了床），不能被 snapshot 覆盖。
func TestDecodeSnapshot_PreservesHumanBelief(t *testing.T) {
	g := NewRoomGrid(400, 500, 10)
	// 模拟 RegisterRoom→SetPrior 的结果：cell 5 是人标 Bed
	g.Cells[5].Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	g.Cells[5].Belief[1] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	g.Cells[5].Belief[2] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
	g.Cells[5].AreaType = AreaBed

	// snapshot 里这个 cell 是 SourceLearned Sit（旧 layout 时学到的）
	snap := GridSnapshot{
		SchemaVer: SnapshotSchemaVersion,
	}
	snap.Grid.W = g.Width
	snap.Grid.H = g.Height
	snap.Grid.OX = g.OriginX
	snap.Grid.OY = g.OriginY
	snap.Cells = []CellSnapshot{
		{
			I: 5,
			B: [3][3]int{
				{int(AreaSit), 80, int(SourceLearned)},
				{int(AreaSit), 75, int(SourceLearned)},
				{int(AreaSit), 70, int(SourceLearned)},
			},
		},
	}

	if err := DecodeSnapshot(snap, g); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if g.Cells[5].Belief[0].Source != SourceHuman ||
		g.Cells[5].Belief[0].Type != AreaBed {
		t.Errorf("Human Belief was overwritten by snapshot: %+v", g.Cells[5].Belief[0])
	}
}

// TestLayoutHash_Stability 同输入同输出
func TestLayoutHash_Stability(t *testing.T) {
	cfg := makeTestRoomConfig()
	h1 := LayoutHash(cfg)
	h2 := LayoutHash(cfg)
	if h1 != h2 {
		t.Errorf("LayoutHash unstable: %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("hash len=%d expected 64 (sha256 hex)", len(h1))
	}
}

// TestLayoutHash_DetectsChanges 任何几何字段变化 → hash 变
func TestLayoutHash_DetectsChanges(t *testing.T) {
	base := LayoutHash(makeTestRoomConfig())

	// 移动床
	cfg := makeTestRoomConfig()
	cfg.Beds[0].X1 += 10
	if LayoutHash(cfg) == base {
		t.Error("moving bed didn't change hash")
	}

	// 加门
	cfg = makeTestRoomConfig()
	cfg.Enters = append(cfg.Enters, radarutils.Rect{X1: 10, Y1: 10, X2: 30, Y2: 30})
	if LayoutHash(cfg) == base {
		t.Error("adding enter didn't change hash")
	}

	// 改雷达 rotation
	cfg = makeTestRoomConfig()
	cfg.Radar.Rotation += 30
	if LayoutHash(cfg) == base {
		t.Error("changing radar mount didn't change hash")
	}
}

// TestDecodeSnapshot_SchemaMismatch 拒绝不支持的 schema_v
func TestDecodeSnapshot_SchemaMismatch(t *testing.T) {
	g := NewRoomGrid(400, 500, 10)
	snap := GridSnapshot{SchemaVer: 99}
	snap.Grid.W = g.Width
	snap.Grid.H = g.Height
	if err := DecodeSnapshot(snap, g); err == nil {
		t.Error("expected schema mismatch error")
	}
}

// TestDecodeSnapshot_DimensionMismatch 拒绝尺寸不匹配
func TestDecodeSnapshot_DimensionMismatch(t *testing.T) {
	g := NewRoomGrid(400, 500, 10)
	snap := GridSnapshot{SchemaVer: SnapshotSchemaVersion}
	snap.Grid.W = 99
	snap.Grid.H = g.Height
	if err := DecodeSnapshot(snap, g); err == nil {
		t.Error("expected dimension mismatch error")
	}
}

func makeTestRoomConfig() RoomConfig {
	return RoomConfig{
		RoomID:  "test-room",
		RoomW:   400,
		RoomH:   500,
		OriginX: 0,
		OriginY: 0,
		WallPolygon: []radarutils.Point{
			{X: 0, Y: 0}, {X: 400, Y: 0}, {X: 400, Y: 500}, {X: 0, Y: 500},
		},
		Enters: []radarutils.Rect{{X1: 0, Y1: 240, X2: 30, Y2: 260}},
		Beds:   []radarutils.Rect{{X1: 100, Y1: 100, X2: 200, Y2: 200}},
		Radar: radarutils.RadarMount{
			Center:   radarutils.Point{X: 200, Y: 250},
			Rotation: 90,
			HFOV:     120, VFOV: 90, AzimuthFOV: 120, ElevationFOV: 90,
		},
	}
}
