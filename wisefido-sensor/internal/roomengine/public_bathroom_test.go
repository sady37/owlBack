// public_bathroom_test.go — sensor_v2 PR-7 Public Bathroom Standalone Mode 验收
//
// PR-7 仅范围内（7a + 7b）：
//   - RegisterRoom 检测 cfg.IsPublicBathroom 调 census.MarkPublicBathroom
//   - SuiteID = bathroom /128 自身（与 suite bathroom 不冲突 census key）
//   - BathroomGate 在 public mode 下成员仍可累计（自家方法），但 census 操作 noop
//   - BathroomGhostAdjudicator 规则 1/5 在 public mode 下仍生效，规则 2/3 因 GetBathroomCount=-1 自然跳过
//
// 7c/7d（fall 触发主体改造）留 Phase C PR-10。

package roomengine

import (
	"testing"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"
)

const (
	pubBathroomRoom  = "fd00:0:3:444:80::/128"
	pubBathroomSuite = "fd00:0:3:444:80::/128" // public mode: SuiteID = room 自身 /128
)

// RegisterRoom 触发 MarkPublicBathroom 的端到端验证。
// 由于 RegisterRoom 内部要走完整 grid stamp 流程，本测试用最小 cfg（只设必要字段）+ 预挂 census。
func TestRegisterRoom_PublicBathroom_MarksCensus(t *testing.T) {
	e := newTestEngineForPublicBathroom(t)
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), zap.NewNop())
	e.SetSuiteCensus(census)

	e.RegisterRoom(RoomConfig{
		RoomID:           pubBathroomRoom,
		RoomName:         "public bathroom #1F",
		RoomKind:         "bathroom",
		IsPublicBathroom: true,
		SuiteID:          pubBathroomSuite,
		RoomW:            200,
		RoomH:            200,
	})

	c := census.Get(pubBathroomSuite)
	if c == nil {
		t.Fatalf("RegisterRoom should have created census bucket")
	}
	if !c.IsPublicBathroom {
		t.Errorf("census.IsPublicBathroom should be true after RegisterRoom with IsPublicBathroom=true")
	}
}

// IsPublicBathroom=true 但 RoomKind 不是 bathroom → 配置无效，忽略
func TestRegisterRoom_NonBathroomIgnoresPublicFlag(t *testing.T) {
	e := newTestEngineForPublicBathroom(t)
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), zap.NewNop())
	e.SetSuiteCensus(census)

	const bedroomID = "fd00:0:3:444:81::/128"
	e.RegisterRoom(RoomConfig{
		RoomID:           bedroomID,
		RoomName:         "bedroom A",
		RoomKind:         "", // default bedroom
		IsPublicBathroom: true, // 配置错误：bedroom 不能 public
		SuiteID:          "fd00:0:3:444::/80",
		RoomW:            200,
		RoomH:            200,
	})

	// bedroom 不应触发 MarkPublicBathroom — census 不应有此 SuiteID bucket（或有也非 public）
	if c := census.Get("fd00:0:3:444::/80"); c != nil && c.IsPublicBathroom {
		t.Errorf("non-bathroom room should not mark census as public, got IsPublicBathroom=true")
	}
}

// IsPublicBathroom=true 但 census 未注入 → 不 panic（运维 misconfiguration 兜底）
func TestRegisterRoom_PublicBathroom_NoCensusGraceful(t *testing.T) {
	e := newTestEngineForPublicBathroom(t)
	// 故意不调 SetSuiteCensus

	// 不应 panic
	e.RegisterRoom(RoomConfig{
		RoomID:           pubBathroomRoom,
		RoomName:         "public bathroom",
		RoomKind:         "bathroom",
		IsPublicBathroom: true,
		SuiteID:          pubBathroomSuite,
		RoomW:            200,
		RoomH:            200,
	})
}

// End-to-end: public bathroom mode 下 BathroomGate.Process 不影响 census BathroomCount，
// 但 gate 自己的 members 仍累计（gate 不知道是 public，靠 census 边界 sentinel 自然降级）。
func TestPublicBathroom_GateProcessNoOpOnCensus(t *testing.T) {
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), zap.NewNop())
	census.MarkPublicBathroom(pubBathroomSuite)

	gate := NewBathroomGate(pubBathroomRoom, pubBathroomSuite, census, zap.NewNop())
	gate.Process([]TrackStatusBase{
		{TrackID: 7, RoomID: pubBathroomRoom},
		{TrackID: 8, RoomID: pubBathroomRoom},
	}, 1_000_000)

	// gate 本地 members 累计（gate 不分 mode）
	if gate.MembersCount() != 2 {
		t.Errorf("gate members should accumulate even in public mode, got %d", gate.MembersCount())
	}

	// census.BathroomCount 仍是 0（AdjustBathroomCount 在 public mode 返 -1 sentinel）
	c := census.Get(pubBathroomSuite)
	if c.BathroomCount != 0 {
		t.Errorf("public mode census.BathroomCount should remain 0, got %d", c.BathroomCount)
	}
}

// End-to-end: public mode 下 BathroomGhostAdjudicator 规则 1/5 仍生效，规则 2/3 跳过。
func TestPublicBathroom_GhostAdjudicator_Rules1And5Active(t *testing.T) {
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), zap.NewNop())
	census.MarkPublicBathroom(pubBathroomSuite)

	g := makeBathroomGrid(t, true)
	a := NewBathroomGhostAdjudicator(
		census,
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return pubBathroomSuite },
		zap.NewNop(),
	)

	// 帧 1：建立 track 7 = real（紧贴 entry，不触发 Rule 1）
	a.Adjudicate([]TrackStatusBase{
		{TrackID: 7, RoomID: pubBathroomRoom, X: -90, Y: 100, Verdict: VerdictReal},
	}, 1_000_000)

	// 帧 2：track 8 出生远离 entry → Rule 1 ghost；track 9 紧贴 7 → Rule 5 ghost
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: pubBathroomRoom, X: -90, Y: 100, Verdict: VerdictReal},
		{TrackID: 8, RoomID: pubBathroomRoom, X: 50, Y: 150, Verdict: VerdictPending},
		{TrackID: 9, RoomID: pubBathroomRoom, X: -80, Y: 110, Verdict: VerdictPending},
	}
	deltas := a.Adjudicate(bases, 1_001_000)

	found8 := false
	found9 := false
	for _, d := range deltas {
		if d.TrackID == 8 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			found8 = true
		}
		if d.TrackID == 9 && d.NewVerdict != nil && *d.NewVerdict == VerdictGhost {
			found9 = true
		}
	}
	if !found8 {
		t.Errorf("public mode: Rule 1 should ghost track 8 (far from entry)")
	}
	if !found9 {
		t.Errorf("public mode: Rule 5 should ghost track 9 (split-adjacent to real)")
	}

	// Rule 3 不应执行：BathroomCount 仍 0（不修正）
	if c := census.Get(pubBathroomSuite); c.BathroomCount != 0 {
		t.Errorf("public mode Rule 3 should skip (GetBathroomCount returns -1 sentinel), but BathroomCount=%d", c.BathroomCount)
	}
}

// suite vs public 共存：同 /80 下挂 1 个 public bathroom + 1 个 suite (bedroom + bathroom)，
// SuiteID 取值不同 → census buckets 不冲突。
func TestPublicAndSuiteBathroom_CensusKeysDisjoint(t *testing.T) {
	census := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), zap.NewNop())

	// suite census (unit /80)
	suiteUnit := "fd00:0:3:555::/80"
	if _, ok := census.UpdatePersonFromTrack(suiteUnit, "resident_uuid", 1, true, 0, true, 1_000_000); !ok {
		t.Fatal("setup: suite resident upgrade failed")
	}

	// public bathroom census (room /128)
	publicRoom := "fd00:0:3:555:88::/128"
	census.MarkPublicBathroom(publicRoom)

	// 两个 bucket 独立存在
	suite := census.Get(suiteUnit)
	pub := census.Get(publicRoom)
	if suite == nil || pub == nil {
		t.Fatalf("both buckets should exist independently: suite=%v pub=%v", suite, pub)
	}
	if suite.IsPublicBathroom {
		t.Errorf("suite bucket must not be marked public")
	}
	if !pub.IsPublicBathroom {
		t.Errorf("public bucket must be marked public")
	}
	if len(suite.Persons) == 0 {
		t.Errorf("suite bucket should still have resident person")
	}
	if len(pub.Persons) != 0 {
		t.Errorf("public bucket should not have persons, got %d", len(pub.Persons))
	}
}

// --- helpers ---

// newTestEngineForPublicBathroom 构造最小 Engine（不连 redis），用于 RegisterRoom e2e。
func newTestEngineForPublicBathroom(t *testing.T) *Engine {
	t.Helper()
	// redis client = nil 容忍，PR-3 publishTrackStatuses 在 nil 时早 return
	var rc *redis.Client
	return NewEngine(rc, zap.NewNop())
}
