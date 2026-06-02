// bathroom_fall_test.go — sensor_v2 PR-10 BathroomFallRules 验收
//
// 覆盖 §6.A 4 类 fall 路径（10a/10b/10c/10d）+ 4 个关键边界：
//   - 10a 触发条件：cell + pose + still
//   - 10b 90s grace
//   - 10b 决定 18 真多人抑制（含 ghost 配对例外）
//   - 10c bathroom 空 30s timeout 起点重置
//   - 10d SuitePerson 单 dedup（同 person 不重复 fire）
//   - public mode 触发主体改"非 ghost track"
//   - 健康检查兜底（无 census / 从未观测 → noop）

package roomengine

import (
	"context"
	"testing"

	"owl-common/card"
	"owl-common/observation"

	"go.uber.org/zap"
)

const (
	tfRoom  = "fd00:0:3:333:80::/128"
	tfSuite = "fd00:0:3:333::/80"
	tfRes   = "fd00:0:3:333:ff00:01:1:1/128"
)

// captureFallPublisher 单测用 stub —— 仅记录 fire 调用，不实发 redis
type captureFallPublisher struct {
	fired []AIPayload
}

func (p *captureFallPublisher) PublishAIEvent(_ context.Context, _ AIPayload, _ string, _ int64) {
}

func (p *captureFallPublisher) PublishAIAlarm(_ context.Context, payload AIPayload, _ string, _ int64) {
	p.fired = append(p.fired, payload)
}

func (p *captureFallPublisher) DeviceUIDHex(_ string) string { return "" }

// countByReason 数 fired 列表中特定 reason 出现次数
func (p *captureFallPublisher) countByReason(reason string) int {
	n := 0
	for _, f := range p.fired {
		if f.Reason == reason {
			n++
		}
	}
	return n
}

// makeFallRules 标准 setup：census + grid + publisher + rules
func makeFallRules(t *testing.T) (*BathroomFallRules, *SuiteCensusManager, *captureFallPublisher) {
	t.Helper()
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	g := makeBathroomGrid(t, true) // 复用 PR-6 测试的 grid helper
	pub := &captureFallPublisher{}
	r := NewBathroomFallRules(
		m,
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return tfSuite },
		pub,
		zap.NewNop(),
	)
	return r, m, pub
}

// upgradeResidentInBathroom 升 resident 并标到 bathroom（PR-10 触发主语前提）
func upgradeResidentInBathroom(t *testing.T, m *SuiteCensusManager, suiteID, residentID string, nowMs int64) {
	t.Helper()
	if _, ok := m.UpdatePersonFromTrack(suiteID, residentID, 1, true, 0, true, nowMs); !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	if !m.TryFlipSoleResidentRoomType(suiteID, card.RoomTypeBathroom, nowMs) {
		t.Fatal("setup: anchor flip to bathroom failed")
	}
	m.AdjustBathroomCount(suiteID, 1, nowMs) // BathroomCount=1
}

// ---- 10a BathroomStillFall ----

func TestBathroomFall_StillFall_TriggersOnToiletStandStill(t *testing.T) {
	const nowMs = int64(1_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// track 在 toilet cell + Stand + 11min 静止（夜间风险时段阈值 10min）
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: -90, Y: 100, Pose: observation.PoseStanding,
			CellAreaType: AreaToilet, StillSec: 11 * 60, Verdict: VerdictReal},
	}
	// 模拟 BathroomGate 已记 occupied；EverObserved 让健康检查通过
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 5*60_000
	state.LastBases = bases
	state.EverObserved = true

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomStill) != 1 {
		t.Errorf("expected 1 still-fall fired, got %d (all fires: %d)", pub.countByReason(ReasonBathroomStill), len(pub.fired))
	}
}

func TestBathroomFall_StillFall_NotFireOnSitPose(t *testing.T) {
	const nowMs = int64(2_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 用 5min 静止：不到 10a (10min) / 不到 10b (8min) / 不到 10d (7min)。隔离 still-fall pose 判定。
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: -90, Y: 100, Pose: observation.PoseSitting,
			CellAreaType: AreaToilet, StillSec: 5 * 60, Verdict: VerdictReal},
	}
	state := r.getOrCreateState(tfRoom)
	state.LastBases = bases
	state.EverObserved = true

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomStill) != 0 {
		t.Errorf("Sit pose on toilet should NOT fire still fall, got %d still-fall fires", pub.countByReason(ReasonBathroomStill))
	}
}

func TestBathroomFall_StillFall_DedupSameTrack(t *testing.T) {
	const nowMs = int64(3_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: -90, Y: 100, Pose: observation.PoseStanding,
			CellAreaType: AreaShower, StillSec: 15 * 60, Verdict: VerdictReal},
	}
	state := r.getOrCreateState(tfRoom)
	state.LastBases = bases
	state.EverObserved = true

	r.Evaluate(tfRoom, bases, nowMs+1000)
	r.Evaluate(tfRoom, bases, nowMs+2000)
	r.Evaluate(tfRoom, bases, nowMs+3000)
	if pub.countByReason(ReasonBathroomStill) != 1 {
		t.Errorf("same trackID should fire still-fall only once, got %d", pub.countByReason(ReasonBathroomStill))
	}
}

// ---- 10b BathroomBedsideFall ----

func TestBathroomFall_BedsideFall_TriggersAfter90sGraceAnd8min(t *testing.T) {
	const nowMs = int64(4_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 10*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: 50, Y: 150, Pose: observation.PoseLying,
			CellAreaType: AreaActive, StillSec: 9 * 60, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomLongStatic) != 1 {
		t.Errorf("expected 1 bedside fall, got %d", pub.countByReason(ReasonBathroomLongStatic))
	}
}

func TestBathroomFall_BedsideFall_BlockedBy90sGrace(t *testing.T) {
	const nowMs = int64(5_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 60_000
	state.EverObserved = true

	// 用 5min static：不到 10b (8min) 也不到 10d (7min)，隔离 10b grace 测试
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: 50, Y: 150, StillSec: 5 * 60,
			CellAreaType: AreaActive, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomLongStatic) != 0 {
		t.Errorf("within 90s grace should not fire bedside, got %d", pub.countByReason(ReasonBathroomLongStatic))
	}
}

func TestBathroomFall_BedsideFall_SuppressedByMultiResident(t *testing.T) {
	const nowMs = int64(6_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 决定 18：count==2 且 0 ghost → 真多人陪同，不 fire bedside
	m.AdjustBathroomCount(tfSuite, +1, nowMs)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 10*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, StillSec: 15 * 60, CellAreaType: AreaActive, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tfRoom, StillSec: 15 * 60, CellAreaType: AreaActive, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	// 10b 应被 decision 18 抑制；10d 仍可能 fire（10d 不 check multi-resident）
	if pub.countByReason(ReasonBathroomLongStatic) != 0 {
		t.Errorf("decision 18 multi-resident should suppress bedside fall, got %d", pub.countByReason(ReasonBathroomLongStatic))
	}
}

func TestBathroomFall_BedsideFall_NotSuppressedWithGhostCompanion(t *testing.T) {
	const nowMs = int64(7_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// count==2，但 1 个 ghost → "1 真人 + 1 ghost"配对，bedside 仍 fire（共生律）
	m.AdjustBathroomCount(tfSuite, +1, nowMs)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 10*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, StillSec: 15 * 60, CellAreaType: AreaActive, Verdict: VerdictReal},
		{TrackID: 8, RoomID: tfRoom, StillSec: 15 * 60, CellAreaType: AreaActive, Verdict: VerdictGhost},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomLongStatic) != 1 {
		t.Errorf("ghost companion should not suppress bedside (共生律), got %d bedside fires", pub.countByReason(ReasonBathroomLongStatic))
	}
}

// ---- 10c BathroomLostFall strong ----

func TestBathroomFall_LostStrong_TriggersAfter30sEmpty(t *testing.T) {
	const nowMs = int64(8_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 帧 1: 1 个 track（让 EverObserved + LastBases sticky）
	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: -50, Y: 100, Verdict: VerdictReal},
	}, nowMs)

	// 帧 2：空 — EmptySinceMs 开始计时
	r.Evaluate(tfRoom, nil, nowMs+1000)

	// 帧 3：35s 后空 — 应 fire
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)
	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 1 {
		t.Fatalf("expected lost-strong fire after 30s empty, got %d", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_ResetOnNewTrack(t *testing.T) {
	const nowMs = int64(9_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 帧 1：1 track
	r.Evaluate(tfRoom, []TrackStatusBase{{TrackID: 7, RoomID: tfRoom, Verdict: VerdictReal}}, nowMs)
	// 帧 2：空 20s（未到 30s timeout）
	r.Evaluate(tfRoom, nil, nowMs+20_000)
	// 帧 3：track 又出现 — EmptySinceMs 应 reset
	r.Evaluate(tfRoom, []TrackStatusBase{{TrackID: 7, RoomID: tfRoom, Verdict: VerdictReal}}, nowMs+25_000)
	// 帧 4：再次空（EmptySinceMs 从此刻重新计时）
	const emptyStartOffset = 30_000 // after track re-appeared 30s later, empty begins
	r.Evaluate(tfRoom, nil, nowMs+25_000+emptyStartOffset)
	// 帧 5：emptyStart + 35s（empty 持续 35s，> 30s timeout 应 fire）
	r.Evaluate(tfRoom, nil, nowMs+25_000+emptyStartOffset+35_000)
	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 1 {
		t.Errorf("EmptySinceMs reset must require fresh 30s, got %d lost-strong fires", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_SuppressedWhenLastTrackGhost(t *testing.T) {
	const nowMs = int64(8_500_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 帧 1：失锁前最后观测是 ghost（镜面反射越界帧，RawV 超 FOV 深度）
	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 0, RoomID: tfRoom, X: 130, Y: 440, RawV: 200, Verdict: VerdictGhost, GhostPenalty: 80},
	}, nowMs)
	// 帧 2：空 — EmptySinceMs 开始计时
	r.Evaluate(tfRoom, nil, nowMs+1000)
	// 帧 3：35s 后仍空 — 因最后是 ghost，不应 fire
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)

	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 0 {
		t.Fatalf("ghost-last loss must NOT fire lost-strong, got %d fires", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_FiresWhenLastTrackRealMixedGhost(t *testing.T) {
	const nowMs = int64(8_700_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 失锁前最后观测含一个 real track（1 真人 + 1 ghost）→ 不抑制，真人消失仍 fire
	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 0, RoomID: tfRoom, X: -50, Y: 100, Verdict: VerdictReal},
		{TrackID: 1, RoomID: tfRoom, X: 130, Y: 440, Verdict: VerdictGhost, GhostPenalty: 80},
	}, nowMs)
	r.Evaluate(tfRoom, nil, nowMs+1000)
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)

	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 1 {
		t.Fatalf("real-track loss must still fire lost-strong, got %d fires", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_SuppressedDoorExitWithNpZero(t *testing.T) {
	const nowMs = int64(8_900_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)
	// np=0 在 empty 起点(nowMs+1000)附近到达 → 门区 + np=0 合取成立
	r.SetNumberPeopleZeroLookup(func(string) int64 { return nowMs + 1000 })

	// 帧 1：失锁前最后帧在门区(Enter cell) + real
	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 0, RoomID: tfRoom, X: 81, Y: 312, CellAreaType: AreaEnter, Verdict: VerdictReal},
	}, nowMs)
	r.Evaluate(tfRoom, nil, nowMs+1000)
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)

	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 0 {
		t.Fatalf("door-exit + np=0 must NOT fire lost-strong, got %d", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_FiresWhenLostAtToiletDespiteNpZero(t *testing.T) {
	const nowMs = int64(9_100_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)
	r.SetNumberPeopleZeroLookup(func(string) int64 { return nowMs + 1000 })

	// 末帧在马桶(非门区)：水气/盲区丢信号 ≠ 离开 → 即便 np=0 也照报
	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 0, RoomID: tfRoom, X: -90, Y: 100, CellAreaType: AreaToilet, Verdict: VerdictReal},
	}, nowMs)
	r.Evaluate(tfRoom, nil, nowMs+1000)
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)

	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 1 {
		t.Fatalf("loss at toilet must still fire despite np=0, got %d", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

func TestBathroomFall_LostStrong_FiresAtDoorWithoutNpZero(t *testing.T) {
	const nowMs = int64(9_300_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)
	// 无 np=0（lookup 返回 0）：门区位置单独不可信（门口倒地）→ 照报
	r.SetNumberPeopleZeroLookup(func(string) int64 { return 0 })

	r.Evaluate(tfRoom, []TrackStatusBase{
		{TrackID: 0, RoomID: tfRoom, X: 81, Y: 312, CellAreaType: AreaEnter, Verdict: VerdictReal},
	}, nowMs)
	r.Evaluate(tfRoom, nil, nowMs+1000)
	r.Evaluate(tfRoom, nil, nowMs+1000+35_000)

	if pub.countByReason(ReasonSuitePersonCompletelyLost) != 1 {
		t.Fatalf("door position without np=0 must still fire, got %d", pub.countByReason(ReasonSuitePersonCompletelyLost))
	}
}

// ---- 10d BathroomLostFall weak ----

func TestBathroomFall_LostWeak_TriggersOnTrackStatic7min(t *testing.T) {
	const nowMs = int64(10_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 10*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: 50, Y: 150, StillSec: 8 * 60,
			CellAreaType: AreaActive, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonSuitePersonSilentWithGhost) != 1 {
		t.Errorf("expected 1 weak-lost fire, got %d", pub.countByReason(ReasonSuitePersonSilentWithGhost))
	}
}

func TestBathroomFall_LostWeak_DedupSamePerson(t *testing.T) {
	const nowMs = int64(11_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 10*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, StillSec: 8 * 60, CellAreaType: AreaActive, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	r.Evaluate(tfRoom, bases, nowMs+2000)
	r.Evaluate(tfRoom, bases, nowMs+3000)
	if pub.countByReason(ReasonSuitePersonSilentWithGhost) != 1 {
		t.Errorf("same person should only weak-fire once, got %d", pub.countByReason(ReasonSuitePersonSilentWithGhost))
	}
}

// ---- System health / safety ----

func TestBathroomFall_NoCensus_Noop(t *testing.T) {
	const nowMs = int64(12_000_000)
	pub := &captureFallPublisher{}
	g := makeBathroomGrid(t, true)
	r := NewBathroomFallRules(
		NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil),
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return tfSuite },
		pub,
		zap.NewNop(),
	)
	// 不调 GetOrCreate → census 中 tfSuite 不存在
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, StillSec: 15 * 60, CellAreaType: AreaToilet,
			Pose: observation.PoseStanding, Verdict: VerdictReal},
	}
	r.Evaluate(tfRoom, bases, nowMs)
	if len(pub.fired) != 0 {
		t.Errorf("no census → fall must be noop, got %+v", pub.fired)
	}
}

func TestBathroomFall_EmptyBasesAndNoHistory_Noop(t *testing.T) {
	const nowMs = int64(13_000_000)
	r, m, pub := makeFallRules(t)
	upgradeResidentInBathroom(t, m, tfSuite, tfRes, nowMs)

	// 从未观测到 bases — 健康检查应阻止 fire
	r.Evaluate(tfRoom, nil, nowMs+100*1000)
	if len(pub.fired) != 0 {
		t.Errorf("never-observed bathroom should noop (radar uninitialized), got %+v", pub.fired)
	}
}

// ---- Public mode ----

func TestBathroomFall_PublicMode_StillFires_NoSuitePerson(t *testing.T) {
	const nowMs = int64(14_000_000)
	r, m, pub := makeFallRules(t)
	m.MarkPublicBathroom(tfSuite)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 5*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, X: -90, Y: 100, Pose: observation.PoseStanding,
			CellAreaType: AreaToilet, StillSec: 15 * 60, Verdict: VerdictReal},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if pub.countByReason(ReasonBathroomStill) != 1 {
		t.Errorf("public mode still fall should fire by 'non-ghost track', got %d still fires", pub.countByReason(ReasonBathroomStill))
	}
}

func TestBathroomFall_PublicMode_GhostTrackAlone_NoFire(t *testing.T) {
	const nowMs = int64(15_000_000)
	r, m, pub := makeFallRules(t)
	m.MarkPublicBathroom(tfSuite)
	state := r.getOrCreateState(tfRoom)
	state.BathroomOccupiedSinceMs = nowMs - 5*60_000
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tfRoom, Pose: observation.PoseStanding,
			CellAreaType: AreaToilet, StillSec: 15 * 60, Verdict: VerdictGhost},
	}
	state.LastBases = bases

	r.Evaluate(tfRoom, bases, nowMs+1000)
	if len(pub.fired) != 0 {
		t.Errorf("public mode with only ghost should not fire (no real person proxy), got %+v", pub.fired)
	}
}
