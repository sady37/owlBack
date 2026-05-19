// bathroom_gate_test.go — sensor_v2 PR-5 BathroomGate 单测
//
// 覆盖 §4.A.2 + PR-5 验收"track 跨入口时 BathroomCount 正确累积"：
//   - entry 触发 +1 + 跨 0 边界 sole-resident anchor flip
//   - 60s timeout exit 触发 -1 + anchor flip back
//   - Public bathroom census noop（决定 24）
//   - 多 track 不重复 anchor flip（仅跨 0 边界一次）
//   - 多 resident 场景 anchor flip 跳过（决定 19）

package roomengine

import (
	"testing"

	"owl-common/card"

	"go.uber.org/zap"
)

const (
	testBathroomSuite = "fd00:0:3:222::/80"
	testBathroomRoom  = "fd00:0:3:222:80::/128"
	testResidentB     = "fd00:0:3:222:ff00:01:1:1/128"
	testResidentC     = "fd00:0:3:222:ff00:01:1:2/128"
)

// 辅助：建一个 census + 升 1 resident，模拟典型部署
func makeSuiteWithResident(t *testing.T, suiteID, residentID string, nowMs int64) *SuiteCensusManager {
	t.Helper()
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	// sleepad 强升格 → resident.AnchorRoomType = bedroom
	if _, ok := m.UpdatePersonFromTrack(suiteID, residentID, 1 /*trackID*/, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	return m
}

func TestBathroomGate_EntryIncrementsCount(t *testing.T) {
	nowMs := int64(10_000_000)
	m := makeSuiteWithResident(t, testBathroomSuite, testResidentB, nowMs)
	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())

	// 首次 track 出现 → entry → +1 + sole-resident flip to bathroom
	bases := []TrackStatusBase{
		{TrackID: 7, CellAreaType: AreaToilet, RoomID: testBathroomRoom},
	}
	gate.Process(bases, nowMs+1000)

	c := m.Get(testBathroomSuite)
	if c.BathroomCount != 1 {
		t.Errorf("BathroomCount want 1, got %d", c.BathroomCount)
	}
	if c.Persons[testResidentB].AnchorRoomType != card.RoomTypeBathroom {
		t.Errorf("AnchorRoomType want bathroom after entry, got %d",
			c.Persons[testResidentB].AnchorRoomType)
	}
	if gate.MembersCount() != 1 {
		t.Errorf("gate members want 1, got %d", gate.MembersCount())
	}
}

func TestBathroomGate_TimeoutExitDecrements(t *testing.T) {
	nowMs := int64(20_000_000)
	m := makeSuiteWithResident(t, testBathroomSuite, testResidentB, nowMs)
	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())

	gate.Process([]TrackStatusBase{{TrackID: 7, RoomID: testBathroomRoom}}, nowMs+1000)
	if c := m.Get(testBathroomSuite); c.BathroomCount != 1 {
		t.Fatalf("setup: count after entry want 1, got %d", c.BathroomCount)
	}

	// 30s 后空帧 — 还未超 60s timeout → 不触发 exit
	gate.Process(nil, nowMs+1000+30_000)
	if c := m.Get(testBathroomSuite); c.BathroomCount != 1 {
		t.Errorf("count must not drop within 60s timeout, got %d", c.BathroomCount)
	}

	// 70s 后空帧 — 超时 → exit -1 + sole-resident flip back to bedroom
	gate.Process(nil, nowMs+1000+70_000)
	c := m.Get(testBathroomSuite)
	if c.BathroomCount != 0 {
		t.Errorf("BathroomCount want 0 after timeout exit, got %d", c.BathroomCount)
	}
	if c.Persons[testResidentB].AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("AnchorRoomType want bedroom after exit, got %d",
			c.Persons[testResidentB].AnchorRoomType)
	}
	if gate.MembersCount() != 0 {
		t.Errorf("gate members want 0 after timeout, got %d", gate.MembersCount())
	}
}

// 多 track 进入：anchor 只翻一次（跨 0 边界），count 累加全部
func TestBathroomGate_MultipleEntries_SingleAnchorFlip(t *testing.T) {
	nowMs := int64(30_000_000)
	m := makeSuiteWithResident(t, testBathroomSuite, testResidentB, nowMs)
	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())

	// 同帧 2 个 track（resident + visitor 同行进入）
	gate.Process([]TrackStatusBase{
		{TrackID: 7},
		{TrackID: 8},
	}, nowMs+1000)

	c := m.Get(testBathroomSuite)
	if c.BathroomCount != 2 {
		t.Errorf("BathroomCount want 2 for two entries, got %d", c.BathroomCount)
	}
	// AnchorRoomType 在 0→1 第一次跨 0 时已翻转；第二个 track 跨 1→2 不再触发 flip
	if c.Persons[testResidentB].AnchorRoomType != card.RoomTypeBathroom {
		t.Errorf("AnchorRoomType want bathroom, got %d",
			c.Persons[testResidentB].AnchorRoomType)
	}
}

// 2 resident 场景：anchor flip 跳过（决定 19 留 PR-X），count 仍正确
func TestBathroomGate_MultiResident_AnchorFlipSkipped(t *testing.T) {
	nowMs := int64(40_000_000)
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	// 升 2 个 resident
	if _, ok := m.UpdatePersonFromTrack(testBathroomSuite, testResidentB, 11, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident B upgrade failed")
	}
	if _, ok := m.UpdatePersonFromTrack(testBathroomSuite, testResidentC, 12, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident C upgrade failed")
	}

	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())
	gate.Process([]TrackStatusBase{{TrackID: 7}}, nowMs+1000)

	c := m.Get(testBathroomSuite)
	if c.BathroomCount != 1 {
		t.Errorf("BathroomCount want 1 in multi-resident, got %d", c.BathroomCount)
	}
	// 两 resident 都应保留 bedroom anchor（multi-resident 路径不 flip）
	if c.Persons[testResidentB].AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("multi-resident anchor should not auto-flip, B got %d",
			c.Persons[testResidentB].AnchorRoomType)
	}
	if c.Persons[testResidentC].AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("multi-resident anchor should not auto-flip, C got %d",
			c.Persons[testResidentC].AnchorRoomType)
	}
}

// Public bathroom 模式：gate 调用 census 操作全 noop
func TestBathroomGate_PublicMode_Noop(t *testing.T) {
	nowMs := int64(50_000_000)
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	m.MarkPublicBathroom(testBathroomSuite)
	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())

	gate.Process([]TrackStatusBase{{TrackID: 7}}, nowMs+1000)

	c := m.Get(testBathroomSuite)
	// public mode 下 AdjustBathroomCount 返回 (-1,-1)，count 保持 0
	if c.BathroomCount != 0 {
		t.Errorf("public mode BathroomCount should not change via gate, got %d", c.BathroomCount)
	}
}

// track ID 复用：firmware 把同 trackID 给了不同 person —— gate 视作连续 membership
// 不重复触发 entry 事件（直到 60s 失锁后再生为止）。
func TestBathroomGate_SameTrackIDPersistsAcrossFrames(t *testing.T) {
	nowMs := int64(60_000_000)
	m := makeSuiteWithResident(t, testBathroomSuite, testResidentB, nowMs)
	gate := NewBathroomGate(testBathroomRoom, testBathroomSuite, m, zap.NewNop())

	gate.Process([]TrackStatusBase{{TrackID: 7}}, nowMs+1000)
	gate.Process([]TrackStatusBase{{TrackID: 7}}, nowMs+2000)
	gate.Process([]TrackStatusBase{{TrackID: 7}}, nowMs+3000)

	c := m.Get(testBathroomSuite)
	if c.BathroomCount != 1 {
		t.Errorf("repeated frames of same track should only fire 1 entry, got count %d", c.BathroomCount)
	}
}

// AdjustBathroomCount 边界：负值不会让 count 跌穿 0
func TestAdjustBathroomCount_FloorAtZero(t *testing.T) {
	nowMs := int64(70_000_000)
	m := makeSuiteWithResident(t, testBathroomSuite, testResidentB, nowMs)

	// 没有 entry 直接 -1 → count 应 floor 在 0
	old, newCount := m.AdjustBathroomCount(testBathroomSuite, -1, nowMs)
	if old != 0 || newCount != 0 {
		t.Errorf("AdjustBathroomCount(-1) on zero count: want (0,0), got (%d,%d)", old, newCount)
	}
}

// AdjustBathroomCount 在 suite 不存在或 public 模式下返回 (-1, -1)
func TestAdjustBathroomCount_NoopReturnSentinel(t *testing.T) {
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)

	// 不存在的 suite
	if old, newCount := m.AdjustBathroomCount("nonexistent_suite", +1, 1000); old != -1 || newCount != -1 {
		t.Errorf("nonexistent suite should return (-1,-1), got (%d,%d)", old, newCount)
	}

	// public 模式
	m.MarkPublicBathroom(testBathroomSuite)
	if old, newCount := m.AdjustBathroomCount(testBathroomSuite, +1, 1000); old != -1 || newCount != -1 {
		t.Errorf("public mode should return (-1,-1), got (%d,%d)", old, newCount)
	}
}

// TryFlipSoleResidentRoomType 各种边界
func TestTryFlipSoleResidentRoomType_EdgeCases(t *testing.T) {
	nowMs := int64(80_000_000)
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)

	// 0 resident — noop
	if m.TryFlipSoleResidentRoomType(testBathroomSuite, card.RoomTypeBathroom, nowMs) {
		t.Errorf("empty suite should not flip")
	}

	// 单 resident，已在 target → noop（避免冗余写）
	if _, ok := m.UpdatePersonFromTrack(testBathroomSuite, testResidentB, 1, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	if m.TryFlipSoleResidentRoomType(testBathroomSuite, card.RoomTypeDefault, nowMs) {
		t.Errorf("flip to same kind should noop")
	}
	if c := m.Get(testBathroomSuite); c.Persons[testResidentB].AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("anchor must not change on noop flip")
	}

	// 单 resident → 翻成功
	if !m.TryFlipSoleResidentRoomType(testBathroomSuite, card.RoomTypeBathroom, nowMs+1000) {
		t.Errorf("sole resident should flip")
	}
	if c := m.Get(testBathroomSuite); c.Persons[testResidentB].AnchorRoomType != card.RoomTypeBathroom {
		t.Errorf("anchor must change on successful flip")
	}
}
