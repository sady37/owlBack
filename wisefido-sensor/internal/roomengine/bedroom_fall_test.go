// bedroom_fall_test.go — sensor_v2 PR-11 BedroomFallRules 验收
//
// 覆盖 §6.B 11b + 11c：
//   - 11b BedsideFall: BedSession.LeftBedAtMs latch + 床边静止 15min + 夜间风险时段
//   - 11c LostFall: SuitePerson AnchorRoomType=bedroom + LastActiveMs > 10min
//   - per-session / per-person dedup
//   - 健康检查兜底 + IsPublicBathroom 不参与 bedroom 路径

package roomengine

import (
	"testing"
	"time"

	"owl-common/card"

	"go.uber.org/zap"
)

const (
	tbRoom  = "fd00:0:3:444:81::/128"
	tbSuite = "fd00:0:3:444::/80"
	tbRes   = "fd00:0:3:444:ff00:01:1:1/128"
)

// makeBedroomGrid 200×200 grid + 中心 AreaBed cell（让 IsNearPriorType(0,100, AreaBed, 100) 命中）。
func makeBedroomGrid(t *testing.T) *RoomGrid {
	t.Helper()
	g := NewRoomGrid(200, 200, 10)
	// 标 (0, 100) 附近为 AreaBed（cell index col=10, row=10 中心 ≈ (5, 105)）
	// 实际放 col 8..12, row 8..12 一个 5x5 块作 bed
	for row := 8; row <= 12; row++ {
		for col := 8; col <= 12; col++ {
			if row >= g.Height || col >= g.Width {
				continue
			}
			c := &g.Cells[row*g.Width+col]
			c.Belief[0] = BeliefState{Type: AreaBed, Confidence: 99, Source: SourceHuman}
			c.AreaType = AreaBed
		}
	}
	for i := range g.Cells {
		g.Cells[i].InRoom = true
		g.Cells[i].InFOV = true
	}
	return g
}

func makeBedroomFallRules(t *testing.T) (*BedroomFallRules, *SuiteCensusManager, *captureFallPublisher) {
	t.Helper()
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	g := makeBedroomGrid(t)
	pub := &captureFallPublisher{}
	r := NewBedroomFallRules(
		m,
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return tbSuite },
		pub,
		zap.NewNop(),
	)
	// 强制夜间风险时段：注入 UTC tz + 用半夜时刻测试
	r.SetTimezone(time.UTC)
	return r, m, pub
}

// nightMs 返回 UTC 02:00 的 unix ms（落在夜间风险时段 23:30-07:30 内）
func nightMs(t *testing.T, offsetSec int64) int64 {
	t.Helper()
	t0 := time.Date(2026, 5, 18, 2, 0, 0, 0, time.UTC)
	return t0.UnixMilli() + offsetSec*1000
}

func upgradeResidentInBedroom(t *testing.T, m *SuiteCensusManager, suiteID, residentID string, nowMs int64) *SuitePerson {
	t.Helper()
	p, ok := m.UpdatePersonFromTrack(suiteID, residentID, 1, true, 0, true, nowMs)
	if !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	if p.AnchorRoomType != card.RoomTypeDefault {
		t.Fatalf("setup: anchor want bedroom, got %d", p.AnchorRoomType)
	}
	return p
}

// ---- 11b BedsideFall ----

func TestBedroomFall_BedsideFall_FiresWithLeftBedAndStatic(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	// track 距 bed cell ≤ 100cm + 静止 16min
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", InBedSinceMs: nowMs - 30*60_000, LeftBedAtMs: nowMs - 16*60_000},
	}

	r.Evaluate(tbRoom, bases, beds, nowMs+1000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 1 {
		t.Errorf("expected 1 bedside fire, got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}
}

func TestBedroomFall_BedsideFall_NotFireWithoutLeftBed(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	// 无 LeftBed event
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", InBedSinceMs: nowMs - 30*60_000, LeftBedAtMs: 0},
	}

	r.Evaluate(tbRoom, bases, beds, nowMs+1000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 0 {
		t.Errorf("no LeftBed = no bedside fire, got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}
}

func TestBedroomFall_BedsideFall_NotFireDuringDay(t *testing.T) {
	// 白天 14:00 UTC（非风险时段）
	t0 := time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC)
	nowMs := t0.UnixMilli()

	r, m, pub := makeBedroomFallRules(t)
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", LeftBedAtMs: nowMs - 16*60_000},
	}

	r.Evaluate(tbRoom, bases, beds, nowMs+1000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 0 {
		t.Errorf("daytime should not fire (night-only rule), got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}
}

func TestBedroomFall_BedsideFall_DedupPerLeftBedSession(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", LeftBedAtMs: nowMs - 16*60_000},
	}

	r.Evaluate(tbRoom, bases, beds, nowMs+1000)
	r.Evaluate(tbRoom, bases, beds, nowMs+2000)
	r.Evaluate(tbRoom, bases, beds, nowMs+3000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 1 {
		t.Errorf("same LeftBed session should only fire once, got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}

	// 新 LeftBed session（ts 不同）→ 允许再 fire
	beds2 := []BedSessionLatch{
		{DeviceUID: "sleepad-1", LeftBedAtMs: nowMs - 5*60_000}, // 新 ts
	}
	r.Evaluate(tbRoom, bases, beds2, nowMs+4000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 2 {
		t.Errorf("new LeftBed session should allow re-fire, got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}
}

func TestBedroomFall_BedsideFall_GhostNotTrigger(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	// 仅 1 个 ghost track 在床边
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictGhost},
	}
	beds := []BedSessionLatch{{DeviceUID: "sleepad-1", LeftBedAtMs: nowMs - 16*60_000}}

	r.Evaluate(tbRoom, bases, beds, nowMs+1000)
	if pub.countByReason(ReasonBedroomBedsideStatic) != 0 {
		t.Errorf("ghost should not trigger bedside, got %d", pub.countByReason(ReasonBedroomBedsideStatic))
	}
}

// ---- 11c LostFall ----

func TestBedroomFall_LostFall_FiresAfterSilent10min(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true
	state.LastBases = []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, StillSec: 11 * 60, Verdict: VerdictReal},
	}
	// 无 active bed session（人未在床上） → 不抑制
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("expected 1 lost fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// PR-11.1 关键守卫：resident 睡觉时 LastActiveMs 自然冻结，sleepad active bed session 必须抑制
// lost_fall —— 不然每晚 22:10 都误报。
func TestBedroomFall_LostFall_SuppressedBySleepadActiveBed(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 30*60_000 // 静止 30min（≫ 10min lost_fall 阈值）

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 30 * 60, Verdict: VerdictReal},
	}
	// sleepad 在床 30min，未离床 — active session
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", InBedSinceMs: nowMs - 30*60_000, LeftBedAtMs: 0, HasHRRR: true},
	}
	r.Evaluate(tbRoom, bases, beds, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("sleeping resident with active bed session must NOT trigger lost_fall, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// 离床后仍静默 → sleepad gating 解除，正常触发 lost_fall
func TestBedroomFall_LostFall_FiresAfterLeftBedAndSilent(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true
	state.LastBases = []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, StillSec: 11 * 60, Verdict: VerdictReal},
	}
	// sleepad 报已离床（LeftBedAtMs > 0）— gating 解除
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-1", InBedSinceMs: nowMs - 60*60_000, LeftBedAtMs: nowMs - 11*60_000},
	}
	r.Evaluate(tbRoom, bases, beds, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("post-LeftBed silent person should fire lost_fall, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

func TestBedroomFall_LostFall_NotFireWhenActive(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 5*60_000 // 5min 前活动，未超 10min

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, StillSec: 30, Verdict: VerdictReal},
	}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("LastActiveMs within 10min should not fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

func TestBedroomFall_LostFall_DedupSamePerson(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal},
	}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)
	r.Evaluate(tbRoom, bases, nil, nowMs+2000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("same person should only fire once, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

func TestBedroomFall_LostFall_DedupResetOnPersonLeaveBedroom(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Fatalf("expected initial fire")
	}

	// person 走到 bathroom（AnchorRoomType 变）
	m.TryFlipSoleResidentRoomType(tbSuite, card.RoomTypeBathroom, nowMs+10_000)
	r.Evaluate(tbRoom, bases, nil, nowMs+20_000)
	// person 不在 bedroom → 不 fire 也清 dedup
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("person not in bedroom should not re-fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}

	// person 回 bedroom + 还是静默
	m.TryFlipSoleResidentRoomType(tbSuite, card.RoomTypeDefault, nowMs+30_000)
	p2 := m.Get(tbSuite).Persons[tbRes]
	p2.LastActiveMs = nowMs + 30_000 - 11*60_000 // 让回到 bedroom 后仍计算为 11min 静默
	r.Evaluate(tbRoom, bases, nil, nowMs+40_000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 2 {
		t.Errorf("after re-entering bedroom + still silent, should re-fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// PR-11.1 lock-in：stale BedSession 抑制行为（已知 limitation，PR-X 会加 staleness guard）。
// 当 sleepad 失联时 BedSession.InBedSinceMs 永远不刷新，本测试锁定"当前会持续抑制"的行为，
// 防止未来 refactor 时无意改变；PR-X 实施 staleness guard 时此测试需更新。
func TestBedroomFall_LostFall_StaleBedSession_SuppressionLockedIn(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 60*60_000 // 1h 静默（远超 lost 阈值）

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true
	state.LastBases = []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, Verdict: VerdictReal}}

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, StillSec: 60 * 60, Verdict: VerdictReal},
	}
	// stale BedSession：InBedSinceMs 是 24h 前 sleepad 失联前的最后一次 InBed event
	// 真实场景：sleepad 设备 offline 后 BedSession 永远停在 in-bed 状态
	beds := []BedSessionLatch{
		{DeviceUID: "sleepad-offline-1", InBedSinceMs: nowMs - 24*60*60_000, LeftBedAtMs: 0},
	}

	r.Evaluate(tbRoom, bases, beds, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("known limitation: stale BedSession currently suppresses lost_fall (PR-X staleness guard 后会改变)，got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// PR-11.1 dedup-reset path：当之前 fire 过的 person 不再是 sole resident in bedroom（任何原因
// — 离开 bedroom / decay / 升降级），LostFallFired 该 personID 项应释放。
// 验证 dedup map 不会无限累积 stale entries。
func TestBedroomFall_LostFall_DedupReleaseOnPersonChange(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true
	state.LastBases = []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal},
	}
	// 帧 1: 触发 fire + dedup map 记 personID
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Fatal("setup: expected initial fire")
	}
	if !state.LostFallFired[tbRes] {
		t.Fatal("setup: dedup map should record fired person")
	}

	// 帧 2: resident decay 出 census（模拟长时间无活动被 DecayInactive 清掉）
	delete(m.Get(tbSuite).Persons, tbRes)
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)

	// LostFallFired 应被释放（防 dedup map 累积 stale entries）
	if state.LostFallFired[tbRes] {
		t.Errorf("dedup entry should be released when person no longer sole resident in bedroom")
	}
}

// ---- Health / safety ----

func TestBedroomFall_NoCensus_Noop(t *testing.T) {
	nowMs := nightMs(t, 0)
	pub := &captureFallPublisher{}
	g := makeBedroomGrid(t)
	r := NewBedroomFallRules(
		NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil),
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return tbSuite },
		pub,
		zap.NewNop(),
	)
	r.SetTimezone(time.UTC)
	// 不 upgrade resident → census 中 tbSuite 不存在
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	beds := []BedSessionLatch{{LeftBedAtMs: nowMs - 16*60_000}}
	r.Evaluate(tbRoom, bases, beds, nowMs)
	if len(pub.fired) != 0 {
		t.Errorf("no census → noop, got %+v", pub.fired)
	}
}

func TestBedroomFall_PublicMode_Noop(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub := makeBedroomFallRules(t)
	// 先 upgrade resident（census 默认非 public），再 mark public flip — 模拟"census 错误标记"边界 case
	upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	m.MarkPublicBathroom(tbSuite)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 0, Y: 100, StillSec: 16 * 60, Verdict: VerdictReal},
	}
	beds := []BedSessionLatch{{LeftBedAtMs: nowMs - 16*60_000}}
	r.Evaluate(tbRoom, bases, beds, nowMs)
	if len(pub.fired) != 0 {
		t.Errorf("public mode bedroom should noop (rule shouldn't even reach here, but defensive), got %+v", pub.fired)
	}
}

// ---- PR-X: cell-typed lost_fall 阈值 + Source-aware exempt ----

// stampCellArea 把 (x, y) 所在 cell 的 Belief[0] 改写为指定 (type, source)。
// 用 grid.CellAt 反查指针写入；越界返回 nil 时测试自身 fail。
func stampCellArea(t *testing.T, g *RoomGrid, x, y int, ty AreaType, src Source) {
	t.Helper()
	cell := g.CellAt(x, y)
	if cell == nil {
		t.Fatalf("stampCellArea: cell (%d,%d) out of grid", x, y)
	}
	cell.Belief[0] = BeliefState{Type: ty, Confidence: 99, Source: src}
	cell.AreaType = ty
}

// makeBedroomFallRulesWithGrid 暴露 grid 引用以便测试在 anchor 位置打标。
func makeBedroomFallRulesWithGrid(t *testing.T) (*BedroomFallRules, *SuiteCensusManager, *captureFallPublisher, *RoomGrid) {
	t.Helper()
	m := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	g := makeBedroomGrid(t)
	pub := &captureFallPublisher{}
	r := NewBedroomFallRules(
		m,
		func(_ string) *RoomGrid { return g },
		func(_ string) string { return tbSuite },
		pub,
		zap.NewNop(),
	)
	r.SetTimezone(time.UTC)
	return r, m, pub, g
}

// 人工标定 AreaBed (SourceHuman) → 不报 lost_fall（老人在床上久躺是 normal use case）
func TestBedroomFall_LostFall_Exempt_HumanAreaBed(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 4*60*60_000 // 4h 静默（远超任何阈值）

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaBed, SourceHuman)
	bases := []TrackStatusBase{
		{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal},
	}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("human-set AreaBed should be exempt, got %d fires", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// SourceHuman + AreaSit/Toilet/Shower 不享 exempt — 坐姿/卫浴突然静止是真跌倒高发场景
func TestBedroomFall_LostFall_HumanAreaSit_NotExempt_FiresAt30min(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 31*60_000 // 31min — 超 AreaSit 30min 阈值

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaSit, SourceHuman)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("SourceHuman AreaSit must NOT be exempt — sitting still ≥30min should fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

func TestBedroomFall_LostFall_HumanAreaToilet_NotExempt(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000 // 11min — 超默认 10min 阈值

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaToilet, SourceHuman)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("SourceHuman AreaToilet must NOT be exempt — sitting on toilet motionless ≥10min should fire (晕厥/突发疾病高风险场景), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

func TestBedroomFall_LostFall_HumanAreaShower_NotExempt(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaShower, SourceHuman)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("SourceHuman AreaShower must NOT be exempt — 淋浴间滑倒是真跌倒高发场景, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// 学习来源 AreaBed → 2h 阈值（学习值最高 80 conf，不享 exempt）
func TestBedroomFall_LostFall_LearnedAreaBed_FiresAfter2h(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaBed, SourceLearned)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	// 1h59min 静默 — 不到 2h 阈值
	p.LastActiveMs = nowMs - (2*60-1)*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("learned AreaBed @119min: should not fire (<2h), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}

	// 2h1min 静默 — 越阈值
	p.LastActiveMs = nowMs - (2*60+1)*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("learned AreaBed @121min: should fire (>2h), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// 学习来源 AreaSit → 30min 阈值
func TestBedroomFall_LostFall_LearnedAreaSit_FiresAfter30min(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaSit, SourceLearned)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	// 29min 静默 — 不到 30min
	p.LastActiveMs = nowMs - 29*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("learned AreaSit @29min: should not fire (<30min), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}

	// 31min 静默 — 越阈值
	p.LastActiveMs = nowMs - 31*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("learned AreaSit @31min: should fire (>30min), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// AreaActive cell（走道）→ 5min 阈值（短：走廊伫立异常）
func TestBedroomFall_LostFall_AreaActive_FiresAfter5min(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaActive, SourceLearned)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	// 4min 静默 — 不到
	p.LastActiveMs = nowMs - 4*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("AreaActive @4min: should not fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}

	// 6min 静默 — 越阈值
	p.LastActiveMs = nowMs - 6*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("AreaActive @6min: should fire (>5min), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// AreaUnknown / 未打标 → 10min 默认阈值（向后兼容原 PR-11 baseline）
func TestBedroomFall_LostFall_AreaUnknown_DefaultThreshold(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	_ = g // 默认不打标 → AreaUnknown
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	// 9min 静默 — 不到 10min
	p.LastActiveMs = nowMs - 9*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("AreaUnknown @9min: should not fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}

	// 11min 静默 — 越阈值（默认 10min）
	p.LastActiveMs = nowMs - 11*60_000
	r.Evaluate(tbRoom, bases, nil, nowMs+1000)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("AreaUnknown @11min: should fire (>10min default), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// 学习来源 AreaBed (NOT SourceHuman) 不享 exempt — 走 2h 阈值（防"学到 AreaBed 就永远不报"）
func TestBedroomFall_LostFall_LearnedAreaBed_NotExempt(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 3*60*60_000 // 3h 静默 — 超 2h 阈值

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	stampCellArea(t, g, 50, 50, AreaBed, SourceLearned)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}

	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("SourceLearned AreaBed @3h should fire (2h threshold, not exempt), got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// 边界：grid 越界 → cell == nil → 退化为 default 10min（防 nil panic）
func TestBedroomFall_LostFall_CellOutOfGrid_DefaultThreshold(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, _ := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 11*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	// (X=9999) 远超 grid 范围 → CellAt 返回 nil
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 9999, Y: 9999, Verdict: VerdictReal}}
	r.Evaluate(tbRoom, bases, nil, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 1 {
		t.Errorf("out-of-grid cell @11min should fall back to default 10min and fire, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}

// Sleepad gating 仍然 active 时优先于 cell-typed 阈值（即使 cell 是 AreaActive 5min 阈值，
// sleepad 在床 → 抑制）。验证 PR-11.1 守卫不会因 cell-typed 改造而被绕过。
func TestBedroomFall_LostFall_SleepadGate_TrumpsCellThreshold(t *testing.T) {
	nowMs := nightMs(t, 0)
	r, m, pub, g := makeBedroomFallRulesWithGrid(t)
	p := upgradeResidentInBedroom(t, m, tbSuite, tbRes, nowMs)
	p.LastActiveMs = nowMs - 30*60_000

	state := r.getOrCreateState(tbRoom)
	state.EverObserved = true

	// cell 是 AreaActive（5min 阈值 — 短）+ 30min 静默 → 过阈值
	stampCellArea(t, g, 50, 50, AreaActive, SourceLearned)
	bases := []TrackStatusBase{{TrackID: 7, RoomID: tbRoom, X: 50, Y: 50, Verdict: VerdictReal}}
	// 但 sleepad active in-bed → 抑制
	beds := []BedSessionLatch{{DeviceUID: "sleepad-1", InBedSinceMs: nowMs - 30*60_000, LeftBedAtMs: 0}}
	r.Evaluate(tbRoom, bases, beds, nowMs)
	if pub.countByReason(ReasonBedroomPersonSilent) != 0 {
		t.Errorf("active sleepad must suppress even when cell threshold exceeded, got %d", pub.countByReason(ReasonBedroomPersonSilent))
	}
}
