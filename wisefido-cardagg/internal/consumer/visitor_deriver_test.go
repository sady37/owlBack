// visitor_deriver_test.go — Tier1 episode state machine（不动 DB / metaCache / reader）。
//
// 直接调 tickCard 喂 peopleCount 序列；store 用 fakeVisitorHistoryStore 收 INSERT/UPDATE 调用；
// 不构造 metaCache / reader / merger / bedPeople（只测 episode 状态机本身）。

package consumer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// fakeMerger 记录 ApplyVisitor 调用，按需返回 merged。
type fakeMerger struct {
	mu             sync.Mutex
	calls          int
	lastVisitor    service.VisitorFields
	mergedToReturn *card.TargetState
}

func (m *fakeMerger) ApplyVisitor(_ context.Context, _ string, v service.VisitorFields) *card.TargetState {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.lastVisitor = v
	return m.mergedToReturn
}

// fakeWriter 记录 WriteCardStatus 调用。
type fakeWriter struct {
	mu    sync.Mutex
	calls []*card.CardStatus
}

func (w *fakeWriter) WriteCardStatus(_ context.Context, status *card.CardStatus) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.calls = append(w.calls, status)
	return nil
}

// fakeRefresher 记录 RebuildDisplay 调用。
type fakeRefresher struct {
	mu    sync.Mutex
	calls []string
}

func (r *fakeRefresher) Rebuild(_ context.Context, cardID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, cardID)
}

// fakeMetaSource 给 tick() 喂 cardID 列表（Path A / Path B 范围）。
type fakeMetaSource struct {
	bedCards     []string
	privateRooms []string
}

func (m *fakeMetaSource) ListBedCardsWithBedBoundRadar(_ context.Context) []string {
	return m.bedCards
}

func (m *fakeMetaSource) ListPrivateRoomCardIDs(_ context.Context) []string {
	return m.privateRooms
}

// fakeBedPeople 按 cardID 查表返 peopleCount。
type fakeBedPeople struct {
	counts map[string]int
}

func (b *fakeBedPeople) CardPeopleCount(_ context.Context, cardID string) int {
	return b.counts[cardID]
}

// fakeReader Path B 用：按 cardID 返 CardStatus（带 RoomState.TotalPeople）。
type fakeReader struct {
	statuses map[string]*card.CardStatus
}

func (r *fakeReader) ReadCardStatus(_ context.Context, cardID string) (*card.CardStatus, error) {
	return r.statuses[cardID], nil
}

type fakeStoreCall struct {
	op            string // "insert" / "close"
	cardID        string
	episodeID     string
	startedAtMs   int64
	endedAtMs     int64
	durationSec   int
}

type fakeVisitorHistoryStore struct {
	mu         sync.Mutex
	calls      []fakeStoreCall
	nextID     int
	insertErr  error
}

func newFakeStore() *fakeVisitorHistoryStore {
	return &fakeVisitorHistoryStore{}
}

func (f *fakeVisitorHistoryStore) insertEpisode(_ context.Context, cardID string, startedAtMs int64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertErr != nil {
		return "", f.insertErr
	}
	f.nextID++
	id := fmt.Sprintf("eid-%d", f.nextID)
	f.calls = append(f.calls, fakeStoreCall{op: "insert", cardID: cardID, episodeID: id, startedAtMs: startedAtMs})
	return id, nil
}

func (f *fakeVisitorHistoryStore) closeEpisode(_ context.Context, episodeID string, endedAtMs int64, durationSec int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeStoreCall{op: "close", episodeID: episodeID, endedAtMs: endedAtMs, durationSec: durationSec})
	return nil
}

func newTestVisitorDeriver(store visitorHistoryStore) *VisitorDeriver {
	return &VisitorDeriver{
		store:    store,
		logger:   zap.NewNop(),
		interval: time.Minute,
		segments: make(map[string]*visitorSegment),
	}
}

// 上升沿：5min 累加跨阈值 INSERT 一行；阈值后多 tick 不重复 INSERT。
func TestEpisode_RisingEdgeInsertsOnce(t *testing.T) {
	store := newFakeStore()
	v := newTestVisitorDeriver(store)
	cardID := "fd00:0:3:111:3:101::/96"
	baseMs := int64(1_700_000_000_000)
	date := "2026-05-19"

	// 5 个 tick 累加到 ≥5min 触发上升沿
	for i := 0; i < 7; i++ {
		v.tickCard(context.Background(), cardID, 2, baseMs+int64(i)*60_000, date)
	}

	inserts := 0
	for _, c := range store.calls {
		if c.op == "insert" {
			inserts++
			if c.cardID != cardID {
				t.Errorf("insert cardID got %s want %s", c.cardID, cardID)
			}
			if c.startedAtMs != baseMs {
				t.Errorf("insert started_at got %d want %d (seg start)", c.startedAtMs, baseMs)
			}
		}
	}
	if inserts != 1 {
		t.Errorf("expected exactly 1 insert across 7 ≥2 ticks, got %d (calls=%v)", inserts, store.calls)
	}

	// segment 状态：visitorStartTs 锚定，currentEpisode 非空
	v.mu.Lock()
	seg := v.segments[cardID]
	v.mu.Unlock()
	if seg.visitorStartTs != baseMs {
		t.Errorf("visitorStartTs got %d want %d", seg.visitorStartTs, baseMs)
	}
	if seg.currentEpisode == "" {
		t.Errorf("currentEpisode should be set after rising edge")
	}
	if !seg.hasToday {
		t.Errorf("hasToday should be true")
	}
}

// 快闪段：未跨 5min 阈值就降回 → 不 INSERT。
func TestEpisode_FlashSegmentNoInsert(t *testing.T) {
	store := newFakeStore()
	v := newTestVisitorDeriver(store)
	cardID := "fd00:0:3:111:3:101::/96"
	baseMs := int64(1_700_000_000_000)
	date := "2026-05-19"

	// 3 tick ≥2 然后降回 — 未跨 5min 阈值
	for i := 0; i < 3; i++ {
		v.tickCard(context.Background(), cardID, 2, baseMs+int64(i)*60_000, date)
	}
	v.tickCard(context.Background(), cardID, 1, baseMs+3*60_000, date)

	for _, c := range store.calls {
		if c.op == "insert" {
			t.Errorf("flash segment should not insert, got %+v", c)
		}
		if c.op == "close" {
			t.Errorf("flash segment should not close (no episode opened), got %+v", c)
		}
	}

	v.mu.Lock()
	seg := v.segments[cardID]
	v.mu.Unlock()
	if seg.hasToday {
		t.Errorf("hasToday should remain false (segment never crossed threshold)")
	}
	if seg.segmentStartTs != 0 || seg.segDurationMin != 0 {
		t.Errorf("segment should be cleared after falling edge: %+v", seg)
	}
}

// 下降沿：跨阈值后 peopleCount → <2 关 episode；duration_sec 是 (now - segmentStartTs)/1000。
func TestEpisode_FallingEdgeClosesWithDuration(t *testing.T) {
	store := newFakeStore()
	v := newTestVisitorDeriver(store)
	cardID := "fd00:0:3:111:3:101::/96"
	baseMs := int64(1_700_000_000_000)
	date := "2026-05-19"

	// 5 tick 跨阈值 + 第 6 tick 降回
	for i := 0; i < 5; i++ {
		v.tickCard(context.Background(), cardID, 2, baseMs+int64(i)*60_000, date)
	}
	closeMs := baseMs + 5*60_000 // 5min later
	v.tickCard(context.Background(), cardID, 1, closeMs, date)

	var closes []fakeStoreCall
	for _, c := range store.calls {
		if c.op == "close" {
			closes = append(closes, c)
		}
	}
	if len(closes) != 1 {
		t.Fatalf("expected 1 close, got %d (calls=%v)", len(closes), store.calls)
	}
	cl := closes[0]
	if cl.endedAtMs != closeMs {
		t.Errorf("close endedAt got %d want %d", cl.endedAtMs, closeMs)
	}
	wantDur := int((closeMs - baseMs) / 1000)
	if cl.durationSec != wantDur {
		t.Errorf("close duration_sec got %d want %d", cl.durationSec, wantDur)
	}

	v.mu.Lock()
	seg := v.segments[cardID]
	v.mu.Unlock()
	if seg.currentEpisode != "" {
		t.Errorf("currentEpisode should be cleared after close, got %s", seg.currentEpisode)
	}
	// 但 visitorStartTs 和 hasToday 保留（"今天最近一次"显示需要）
	if seg.visitorStartTs != baseMs {
		t.Errorf("visitorStartTs should remain after close, got %d", seg.visitorStartTs)
	}
	if !seg.hasToday {
		t.Errorf("hasToday should remain after close (today still has visitor)")
	}
}

// 午夜 reset：跨日时关进行中 episode（ended_at=midnight），清今日字段。
func TestEpisode_MidnightClosesOngoing(t *testing.T) {
	store := newFakeStore()
	v := newTestVisitorDeriver(store)
	cardID := "fd00:0:3:111:3:101::/96"

	// 第 1 天 23:55 起 + 累加 5 分钟到 23:59 跨阈值 + 持续到 00:01 第 2 天
	day1 := "2026-05-19"
	day2 := "2026-05-20"
	startMs := mustParseTime(t, "2026-05-19T23:55:00Z")

	// 5 ticks day1
	for i := 0; i < 5; i++ {
		v.tickCard(context.Background(), cardID, 2, startMs+int64(i)*60_000, day1)
	}
	// 跨日：第 6 tick 在 day2 00:00:00
	v.tickCard(context.Background(), cardID, 2, startMs+5*60_000, day2)

	var closes []fakeStoreCall
	for _, c := range store.calls {
		if c.op == "close" {
			closes = append(closes, c)
		}
	}
	if len(closes) != 1 {
		t.Fatalf("expected 1 midnight close, got %d (calls=%v)", len(closes), store.calls)
	}
	midnightMs := mustParseTime(t, "2026-05-20T00:00:00Z")
	if closes[0].endedAtMs != midnightMs {
		t.Errorf("midnight close endedAt got %d want %d", closes[0].endedAtMs, midnightMs)
	}
	// duration = 23:55→24:00 = 5min = 300s
	if closes[0].durationSec != 300 {
		t.Errorf("midnight close duration_sec got %d want 300", closes[0].durationSec)
	}

	// 跨日后 segment 重置（今日字段清，新一天重新观察）
	v.mu.Lock()
	seg := v.segments[cardID]
	v.mu.Unlock()
	if seg.hasToday {
		t.Errorf("hasToday should be false after midnight reset")
	}
	if seg.visitorStartTs != 0 {
		t.Errorf("visitorStartTs should be 0 after midnight reset, got %d", seg.visitorStartTs)
	}
	if seg.currentEpisode != "" {
		t.Errorf("currentEpisode should be cleared after midnight reset, got %s", seg.currentEpisode)
	}
}

// 午夜 hash 写：midnight 跨日时主动落 hash 清残留（修「凌晨无 activity event sensor 不推 → hash 卡昨日字段」）。
func TestEpisode_MidnightWritesHashAndRefreshesParent(t *testing.T) {
	store := newFakeStore()
	mergedTS := &card.TargetState{LastActiveTs: 1_700_000_000_000, UpdatedAt: 1_700_000_000_000}
	merger := &fakeMerger{mergedToReturn: mergedTS}
	writer := &fakeWriter{}
	refresher := &fakeRefresher{}

	v := newTestVisitorDeriver(store)
	v.merger = merger
	v.writer = writer
	v.picker = refresher

	cardID := "fd00:0:3:111:3:101::/96"
	day1 := "2026-05-19"
	day2 := "2026-05-20"
	startMs := mustParseTime(t, "2026-05-19T22:00:00Z")

	// 第 1 天累加超阈值 + 持续 → segment open
	for i := 0; i < 6; i++ {
		v.tickCard(context.Background(), cardID, 2, startMs+int64(i)*60_000, day1)
	}
	// 写 hash 不应在这之前被触发（非午夜 tick）
	if len(writer.calls) != 0 {
		t.Errorf("non-midnight tick should not write hash, got %d writes", len(writer.calls))
	}

	// 第 7 tick 跨日（day2） — midnight reset 触发
	v.tickCard(context.Background(), cardID, 2, mustParseTime(t, "2026-05-20T00:00:00Z"), day2)

	// 验证 midnight 写发生
	if len(writer.calls) != 1 {
		t.Fatalf("midnight should trigger exactly 1 write, got %d", len(writer.calls))
	}
	w := writer.calls[0]
	if w.CardID != cardID {
		t.Errorf("write CardID got %s want %s", w.CardID, cardID)
	}
	if w.Target != mergedTS {
		t.Errorf("write Target should be merger's returned merged TS")
	}

	// 验证 ApplyVisitor 收到的是 reset 后的字段（visitorStartTs=0, hasToday=false）
	if merger.lastVisitor != (service.MakeVisitorFields(0, false)) {
		t.Errorf("midnight ApplyVisitor should receive cleared fields, got %+v", merger.lastVisitor)
	}

	// 验证 RebuildDisplay 跟着调
	if len(refresher.calls) != 1 || refresher.calls[0] != cardID {
		t.Errorf("RebuildDisplay should be called once with cardID, got %v", refresher.calls)
	}
}

// 非午夜 tick 不写 hash（沿用 sensor target.state 自然带的路径）。
func TestEpisode_NonMidnightDoesNotWriteHash(t *testing.T) {
	merger := &fakeMerger{mergedToReturn: &card.TargetState{}}
	writer := &fakeWriter{}

	v := newTestVisitorDeriver(newFakeStore())
	v.merger = merger
	v.writer = writer

	cardID := "fd00:0:3:111:3:101::/96"
	date := "2026-05-19"
	baseMs := int64(1_700_000_000_000)

	// 同日内大量 tick（上升沿 + 持续 + 下降沿都不应触发 hash 写）
	for i := 0; i < 10; i++ {
		v.tickCard(context.Background(), cardID, 2, baseMs+int64(i)*60_000, date)
	}
	v.tickCard(context.Background(), cardID, 1, baseMs+10*60_000, date)

	if len(writer.calls) != 0 {
		t.Errorf("non-midnight ticks should not write hash, got %d (calls=%v)", len(writer.calls), writer.calls)
	}
	// 但 merger.ApplyVisitor 每 tick 都调
	if merger.calls != 11 {
		t.Errorf("ApplyVisitor should be called every tick: got %d want 11", merger.calls)
	}
}

// midnightMsForDate: 给 dateUTC 返回该日 23:59:59 之后的 0:00:00 UTC（即次日开始 ms）。
func TestMidnightMsForDate(t *testing.T) {
	got := midnightMsForDate("2026-05-19")
	want := mustParseTime(t, "2026-05-20T00:00:00Z")
	if got != want {
		t.Errorf("midnight 2026-05-19 got %d want %d", got, want)
	}
	if midnightMsForDate("bad-date") != 0 {
		t.Errorf("bad date should return 0")
	}
}

// nil store: VisitorDeriver 应 tolerant — 不写 DB 不 panic，hash 字段仍更新。
func TestEpisode_NilStoreTolerant(t *testing.T) {
	v := newTestVisitorDeriver(nil)
	cardID := "fd00:0:3:111:3:101::/96"
	baseMs := int64(1_700_000_000_000)
	date := "2026-05-19"

	for i := 0; i < 7; i++ {
		v.tickCard(context.Background(), cardID, 2, baseMs+int64(i)*60_000, date)
	}
	v.tickCard(context.Background(), cardID, 1, baseMs+8*60_000, date)

	v.mu.Lock()
	seg := v.segments[cardID]
	v.mu.Unlock()
	if !seg.hasToday {
		t.Errorf("hasToday should be true even without store")
	}
	if seg.visitorStartTs != baseMs {
		t.Errorf("visitorStartTs should be set even without store, got %d", seg.visitorStartTs)
	}
}

// =============================================================================
// FU11 — tick() 主循环：Path A / Path B scope + skipParents 父子去重
// =============================================================================

// Path A：bed-bound radar bed cards 范围内每张卡按 BedPeopleTracker 喂 peopleCount。
func TestTick_PathA_BedBoundRadar(t *testing.T) {
	bedA := "fd00:0:99:101:65:101::/96"
	bedB := "fd00:0:99:101:65:102::/96"
	bedC := "fd00:0:99:101:65:103::/96"

	v := newTestVisitorDeriver(newFakeStore())
	v.metaCache = &fakeMetaSource{bedCards: []string{bedA, bedB, bedC}}
	v.bedPeople = &fakeBedPeople{counts: map[string]int{
		bedA: 2, // ≥2 → segment 起步
		bedB: 1, // <2 → segment 不起
		bedC: 3, // ≥2 → segment 起步
	}}
	merger := &fakeMerger{}
	v.merger = merger

	now := time.Unix(1_700_000_000, 0).UTC() // 2023-11-14T22:13:20Z 任意非午夜
	v.tick(context.Background(), now)

	// 三张 bed card 都被 tick（segment entry 创建）
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.segments[bedA] == nil || v.segments[bedA].segmentStartTs == 0 {
		t.Errorf("bedA should have segment started, got %+v", v.segments[bedA])
	}
	if v.segments[bedB] == nil || v.segments[bedB].segmentStartTs != 0 {
		t.Errorf("bedB peopleCount<2 → segment should NOT start, got %+v", v.segments[bedB])
	}
	if v.segments[bedC] == nil || v.segments[bedC].segmentStartTs == 0 {
		t.Errorf("bedC should have segment started, got %+v", v.segments[bedC])
	}
	if merger.calls != 3 {
		t.Errorf("merger.ApplyVisitor calls want 3 (one per bed card), got %d", merger.calls)
	}
}

// Path B：Private /88 room cards 用 reader.ReadCardStatus 拿 room_state.total_people。
func TestTick_PathB_PrivateRoom(t *testing.T) {
	room1 := "fd00:0:99:101:65:200::/88"
	room2 := "fd00:0:99:101:65:201::/88"

	v := newTestVisitorDeriver(newFakeStore())
	v.metaCache = &fakeMetaSource{privateRooms: []string{room1, room2}}
	v.reader = &fakeReader{statuses: map[string]*card.CardStatus{
		room1: {RoomState: &card.RoomState{TotalPeople: 3}}, // ≥2
		room2: {RoomState: &card.RoomState{TotalPeople: 0}}, // <2
	}}
	v.merger = &fakeMerger{}

	now := time.Unix(1_700_000_000, 0).UTC()
	v.tick(context.Background(), now)

	v.mu.Lock()
	defer v.mu.Unlock()
	if v.segments[room1] == nil || v.segments[room1].segmentStartTs == 0 {
		t.Errorf("room1 totalPeople=3 → segment should start, got %+v", v.segments[room1])
	}
	if v.segments[room2] == nil || v.segments[room2].segmentStartTs != 0 {
		t.Errorf("room2 totalPeople=0 → segment should NOT start, got %+v", v.segments[room2])
	}
}

// skipParents：bed card 命中 Path A 时，其父 /88 room 在 Path B 必须跳过（防父子双显）。
func TestTick_SkipParents_BedCoveredRoomSkipped(t *testing.T) {
	bedA := "fd00:0:99:101:65:101::/96"
	parentRoom := "fd00:0:99:101:65:100::/88" // bedA 的父 /88（保留 6 组高 8 位）
	otherRoom := "fd00:0:99:101:65:200::/88"  // 不同 /88，无 bed 覆盖

	v := newTestVisitorDeriver(newFakeStore())
	v.metaCache = &fakeMetaSource{
		bedCards:     []string{bedA},
		privateRooms: []string{parentRoom, otherRoom},
	}
	v.bedPeople = &fakeBedPeople{counts: map[string]int{bedA: 2}}
	v.reader = &fakeReader{statuses: map[string]*card.CardStatus{
		parentRoom: {RoomState: &card.RoomState{TotalPeople: 5}}, // 即便 ≥2 也要被跳过
		otherRoom:  {RoomState: &card.RoomState{TotalPeople: 4}},
	}}
	v.merger = &fakeMerger{}

	now := time.Unix(1_700_000_000, 0).UTC()
	v.tick(context.Background(), now)

	v.mu.Lock()
	defer v.mu.Unlock()
	if v.segments[bedA] == nil {
		t.Errorf("bedA Path A should be processed")
	}
	if _, exists := v.segments[parentRoom]; exists {
		t.Errorf("parentRoom should be skipped (its child bedA already covered in Path A), got %+v", v.segments[parentRoom])
	}
	if v.segments[otherRoom] == nil {
		t.Errorf("otherRoom should be processed (no child bed in Path A)")
	}
}

// parentRoomCardID util：/96 → /88（保留 6 组高 8 位）；非 /96 输入返空。
func TestParentRoomCardID(t *testing.T) {
	// fd00:0:99:101:65:101::/96 → 第 6 组 0101 高 8 位 = 01 → /88 = fd00:0:99:101:65:100::/88
	if got := parentRoomCardID("fd00:0:99:101:65:101::/96"); got != "fd00:0:99:101:65:100::/88" {
		t.Errorf("parentRoomCardID(/96 with 0x0101) want fd00:0:99:101:65:100::/88, got %s", got)
	}
	// 第 6 组 0200 高 8 位 = 02 → /88 = fd00:0:99:101:65:200::/88
	if got := parentRoomCardID("fd00:0:99:101:65:201::/96"); got != "fd00:0:99:101:65:200::/88" {
		t.Errorf("parentRoomCardID(/96 with 0x0201) want fd00:0:99:101:65:200::/88, got %s", got)
	}
	if got := parentRoomCardID("fd00:0:99:101:65::/88"); got != "" {
		t.Errorf("parentRoomCardID(/88 input) want empty (not /96), got %s", got)
	}
	if got := parentRoomCardID("not-an-ipv6"); got != "" {
		t.Errorf("parentRoomCardID(invalid) want empty, got %s", got)
	}
}

// readTotalPeople：reader=nil 安全 / status=nil 安全 / 正常路径。
func TestReadTotalPeople(t *testing.T) {
	room := "fd00:0:99:101:65:200::/88"

	// nil reader
	v := newTestVisitorDeriver(nil)
	v.reader = nil
	if got := v.readTotalPeople(context.Background(), room); got != 0 {
		t.Errorf("nil reader want 0, got %d", got)
	}

	// reader 返 nil status
	v.reader = &fakeReader{statuses: map[string]*card.CardStatus{}}
	if got := v.readTotalPeople(context.Background(), room); got != 0 {
		t.Errorf("nil status want 0, got %d", got)
	}

	// 正常
	v.reader = &fakeReader{statuses: map[string]*card.CardStatus{
		room: {RoomState: &card.RoomState{TotalPeople: 7}},
	}}
	if got := v.readTotalPeople(context.Background(), room); got != 7 {
		t.Errorf("want 7, got %d", got)
	}
}

func mustParseTime(t *testing.T, s string) int64 {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return v.UnixMilli()
}
