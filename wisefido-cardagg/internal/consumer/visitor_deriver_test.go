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

	"go.uber.org/zap"
)

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

func mustParseTime(t *testing.T, s string) int64 {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return v.UnixMilli()
}
