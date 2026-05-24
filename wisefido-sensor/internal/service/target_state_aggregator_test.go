// target_state_aggregator_test.go — TargetStateAggregator T1 契约测试骨架。
//
// P2 scaffold：只测 lifecycle / push channel / pull snapshot 基础路径；
// P3/P4 业务字段填入后，按累加器分组补完整 table-driven test：
//   - TestAggregator_LastActive_*    (P3)
//   - TestAggregator_StandingMin_*   (P3)
//   - TestAggregator_WeakBio_*       (P4，含 80+ escalation rising edge)
//
// 注：Visitor 累加器已挪到 cardagg VisitorDeriver，不在此 sensor 模块测试范围。

package service

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestAggregator(t *testing.T) *TargetStateAggregator {
	t.Helper()
	return NewTargetStateAggregator(nil, zap.NewNop())
}

// TestAggregator_Lifecycle Run 跑起来 + ctx cancel 干净退出。
func TestAggregator_Lifecycle(t *testing.T) {
	a := newTestAggregator(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return within 1s after ctx cancel")
	}
}

// TestAggregator_ZoneEvent_TotalPeopleCache OnZoneEvent 缓存 totalPeople。
// P3/P4 字段补完后 lastActive / standing / visitor gate 都依赖这个缓存。
func TestAggregator_ZoneEvent_TotalPeopleCache(t *testing.T) {
	a := newTestAggregator(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.Run(ctx)

	spatialPrefix := "fd00:0:3:111:3:101::/96"
	a.OnZoneEvent(spatialPrefix, 1, 1_700_000_000_000)
	time.Sleep(20 * time.Millisecond)

	a.mu.RLock()
	acc, ok := a.accums[spatialPrefix]
	a.mu.RUnlock()
	if !ok {
		t.Fatal("accumulator not created for spatialPrefix after ZoneEvent push")
	}
	if acc.totalPeople != 1 {
		t.Errorf("totalPeople cache = %d, want 1", acc.totalPeople)
	}
}

// TestAggregator_GetSnapshot_Empty 无 entry 时返回 ok=false。
func TestAggregator_GetSnapshot_Empty(t *testing.T) {
	a := newTestAggregator(t)
	_, _, _, ok := a.GetSnapshot("fd00:0:3:nonexistent::/96")
	if ok {
		t.Error("GetSnapshot should return ok=false for nonexistent spatialPrefix")
	}
}

// TestAggregator_GetSnapshot_AfterZoneEvent 有 entry 时返回 target + ok=true。
// P2 scaffold 所有累加器字段为 0；P3/P4 后该测试需扩展。
func TestAggregator_GetSnapshot_AfterZoneEvent(t *testing.T) {
	a := newTestAggregator(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.Run(ctx)

	spatialPrefix := "fd00:0:3:111:3:101::/96"
	a.OnZoneEvent(spatialPrefix, 2, 1_700_000_000_000)
	time.Sleep(20 * time.Millisecond)

	target, standingMin, _, ok := a.GetSnapshot(spatialPrefix)
	if !ok {
		t.Fatal("GetSnapshot ok=false after ZoneEvent")
	}
	if target == nil {
		t.Fatal("target nil")
	}
	if standingMin != 0 {
		t.Errorf("standingMin = %d, want 0 (P2 stub)", standingMin)
	}
}

// TestAggregator_ActiveSpatialPrefixes 返回所有 entry 的物理地址。
func TestAggregator_ActiveSpatialPrefixes(t *testing.T) {
	a := newTestAggregator(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.Run(ctx)

	a.OnZoneEvent("fd00:0:3:111:3:101::/96", 1, 1_700_000_000_000)
	a.OnZoneEvent("fd00:0:3:111:3:102::/96", 1, 1_700_000_000_000)
	time.Sleep(20 * time.Millisecond)

	ids := a.ActiveSpatialPrefixes()
	if len(ids) != 2 {
		t.Errorf("ActiveSpatialPrefixes len = %d, want 2", len(ids))
	}
}

// TestAggregator_MarkPublished_ClearsDirty publish 完后 dirty=false。
func TestAggregator_MarkPublished_ClearsDirty(t *testing.T) {
	a := newTestAggregator(t)
	spatialPrefix := "fd00:0:3:111:3:101::/96"
	a.mu.Lock()
	acc := a.getOrCreateLocked(spatialPrefix)
	acc.dirty = true
	a.mu.Unlock()

	if _, _, dirty, _ := a.GetSnapshot(spatialPrefix); !dirty {
		t.Error("dirty should be true before MarkPublished")
	}
	a.MarkPublished(spatialPrefix, time.Now().UnixMilli())
	if _, _, dirty, _ := a.GetSnapshot(spatialPrefix); dirty {
		t.Error("dirty should be false after MarkPublished")
	}
}

// === WeakBio 累加器测试（权重修订 2026-05-19：HR=5 / RR=5 / ApneaH=15；raw 不动 max 0-60）===

// TestWeakBio_ScoreCompute 单测计算公式：max(raw) + 5×HR + 5×RR + 15×Apnea，封顶 100。
func TestWeakBio_ScoreCompute(t *testing.T) {
	tests := []struct {
		name      string
		events    []weakBioEvent
		wantScore int
	}{
		{"empty", nil, 0},
		{"single weak raw=20", []weakBioEvent{{alarmType: "WeakBiometricSignal", rawValue: 20}}, 20},
		{"raw 取 max(20,60,40)=60", []weakBioEvent{{alarmType: "WeakBiometricSignal", rawValue: 20}, {alarmType: "WeakBiometricSignal", rawValue: 60}, {alarmType: "WeakBiometricSignal", rawValue: 40}}, 60},
		{"weak60 + 1 ApneaH = 75 (黄区)", []weakBioEvent{{alarmType: "WeakBiometricSignal", rawValue: 60}, {alarmType: "ApneaHypopnea"}}, 75},
		{"weak60 + 2 ApneaH = 90 (红区)", []weakBioEvent{{alarmType: "WeakBiometricSignal", rawValue: 60}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}}, 90},
		{"6 ApneaH = 90 (AHI≈12 轻度，红区)", []weakBioEvent{{alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}}, 90},
		{"12 HR = 60 (持续真异常，黄区)", []weakBioEvent{{alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}}, 60},
		{"单 1 ApneaH = 15 (低于 30 不显示，避免单条噪声)", []weakBioEvent{{alarmType: "ApneaHypopnea"}}, 15},
		{"5 HR + 3 RR = 40 (灰区 attention)", []weakBioEvent{{alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "HeartRateAlert"}, {alarmType: "RespRateAlert"}, {alarmType: "RespRateAlert"}, {alarmType: "RespRateAlert"}}, 40},
		{"封顶 100", []weakBioEvent{{alarmType: "WeakBiometricSignal", rawValue: 60}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}, {alarmType: "ApneaHypopnea"}}, 100},
		{"HR+RR (1+1) = 10", []weakBioEvent{{alarmType: "HeartRateAlert"}, {alarmType: "RespRateAlert"}}, 10},
		{"无关 alarm 不计 + 1 HR = 5", []weakBioEvent{{alarmType: "Fall"}, {alarmType: "Stay"}, {alarmType: "HeartRateAlert"}}, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := computeWeakBioScore(tc.events); got != tc.wantScore {
				t.Errorf("computeWeakBioScore = %d, want %d", got, tc.wantScore)
			}
		})
	}
}

// TestWeakBio_AccumulatesScore 多事件依次累加 score 缓存正确（无 escalation 路径）。
func TestWeakBio_AccumulatesScore(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	now := int64(1_700_000_000_000)

	// 1. raw=60 → score=60
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "WeakBiometricSignal", TsMs: now, RawValue: 60,
	})
	if got := snapshotScore(a, spatial); got != 60 {
		t.Errorf("after weak raw=60: score=%d want 60", got)
	}

	// 2. + 1 ApneaH → 75
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "ApneaHypopnea", TsMs: now + 1000,
	})
	if got := snapshotScore(a, spatial); got != 75 {
		t.Errorf("after +1 ApneaH: score=%d want 75", got)
	}

	// 3. + 2 ApneaH → 90 (红)
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "ApneaHypopnea", TsMs: now + 2000,
	})
	if got := snapshotScore(a, spatial); got != 90 {
		t.Errorf("after +2 ApneaH: score=%d want 90", got)
	}

	// 4. 累计 → 封顶 100
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "ApneaHypopnea", TsMs: now + 3000,
	})
	if got := snapshotScore(a, spatial); got != 100 {
		t.Errorf("cap at 100: score=%d want 100", got)
	}
}

func snapshotScore(a *TargetStateAggregator, spatial string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	acc := a.accums[spatial]
	if acc == nil {
		return -1
	}
	return acc.weakBio.score
}

// TestWeakBio_WindowLazyDrop 30min 外事件不计入 score（新权重 HR=5）。
func TestWeakBio_WindowLazyDrop(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)

	// 老事件 35min 前
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "HeartRateAlert", TsMs: t0,
	})
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "HeartRateAlert", TsMs: t0,
	})

	// 新事件，35min 后
	tLater := t0 + 35*60*1000
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "HeartRateAlert", TsMs: tLater,
	})

	a.mu.RLock()
	acc := a.accums[spatial]
	score := acc.weakBio.score
	eventCount := len(acc.weakBio.events)
	a.mu.RUnlock()

	if eventCount != 1 {
		t.Errorf("events after lazy drop = %d, want 1 (only the new one)", eventCount)
	}
	if score != 5 {
		t.Errorf("score after lazy drop = %d, want 5 (1 HR @ weight 5)", score)
	}
}

// TestWeakBio_ExpireOnRead_WeakBioScore: WeakBioScore 调用时 lazy drop 窗外 events + 重算 score。
// 验证「sensor 主动 expire」修：score 不再卡老值随心跳偷渡到 cardagg。
func TestWeakBio_ExpireOnRead_WeakBioScore(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"

	// alarm 31min ago — 已在 30min 窗外
	pastMs := time.Now().UnixMilli() - 31*60*1000
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "HeartRateAlert", TsMs: pastMs,
	})

	// handleAlarmEvent 用 e.TsMs 自身作 cutoff 参考，单条 event 入队 score=5
	a.mu.RLock()
	initialScore := a.accums[spatial].weakBio.score
	a.mu.RUnlock()
	if initialScore != 5 {
		t.Fatalf("initial score after handleAlarmEvent want 5, got %d", initialScore)
	}

	// 读：expire-on-read 用 wall-clock now 作 cutoff，过期 event 被丢
	score := a.WeakBioScore(spatial)
	if score != 0 {
		t.Errorf("after expire-on-read 31min stale event, score want 0, got %d", score)
	}

	a.mu.RLock()
	eventCount := len(a.accums[spatial].weakBio.events)
	dirty := a.accums[spatial].dirty
	a.mu.RUnlock()
	if eventCount != 0 {
		t.Errorf("events after expire want 0, got %d", eventCount)
	}
	if !dirty {
		t.Error("dirty should be true after score decay (publisher 下次 tick 需推新 score=0)")
	}
}

// TestWeakBio_ExpireOnRead_GetSnapshot: GetSnapshot 也走 expire-on-read，并标 dirty。
// 这是关键路径——publisher tick 时通过此 path 拿到 target；防 weakBio 老值随 LastActive 心跳"偷渡"。
func TestWeakBio_ExpireOnRead_GetSnapshot(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"

	pastMs := time.Now().UnixMilli() - 31*60*1000
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "ApneaHypopnea", TsMs: pastMs,
	})
	// 入队 score = 15

	// 模拟 publisher 已 publish 过：清 dirty
	a.mu.Lock()
	a.accums[spatial].dirty = false
	a.mu.Unlock()

	target, _, dirty, ok := a.GetSnapshot(spatial)
	if !ok {
		t.Fatal("snapshot should exist")
	}
	if target.WeakBiometricSignal != 0 {
		t.Errorf("snapshot weakBio want 0 after expire, got %d", target.WeakBiometricSignal)
	}
	if !dirty {
		t.Error("dirty should be set after score decayed via GetSnapshot expire")
	}
}

// TestWeakBio_FreshEventsNoSpuriousDirty: 窗内 events 读时不应触发 spurious dirty=true。
// 否则 publisher 会被无意义唤醒推同样的 score。
func TestWeakBio_FreshEventsNoSpuriousDirty(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"

	nowMs := time.Now().UnixMilli()
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "HeartRateAlert", TsMs: nowMs,
	})

	a.mu.Lock()
	a.accums[spatial].dirty = false
	a.mu.Unlock()

	_ = a.WeakBioScore(spatial)

	a.mu.RLock()
	dirty := a.accums[spatial].dirty
	a.mu.RUnlock()
	if dirty {
		t.Error("fresh events should not cause spurious dirty=true on read")
	}
}

// TestWeakBio_UnrelatedAlarmIgnored 非 4 种关联 alarm 不进累加。
func TestWeakBio_UnrelatedAlarmIgnored(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	a.handleAlarmEvent(context.Background(), AlarmEventSnapshot{
		SpatialPrefix: spatial, AlarmType: "Fall", TsMs: 1_700_000_000_000,
	})
	a.mu.RLock()
	_, exists := a.accums[spatial]
	a.mu.RUnlock()
	if exists {
		t.Error("unrelated alarm should not create accumulator entry")
	}
}

// ============================================================================
// S2/FU4 handleEventFrame — LastActive + StandingContinuousMin
// ============================================================================

// 测试辅助：直接 ack totalPeople（不走 channel 避免 goroutine race）
func setTotalPeople(a *TargetStateAggregator, spatial string, count int, tsMs int64) {
	a.handleZoneEvent(ZoneEventSnapshot{SpatialPrefix: spatial, TotalPeople: count, UpdatedAtMs: tsMs})
}

func snapshotLastActive(a *TargetStateAggregator, spatial string) int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	acc := a.accums[spatial]
	if acc == nil {
		return -1
	}
	return acc.lastActive.lastActiveTs
}

func snapshotStanding(a *TargetStateAggregator, spatial string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	acc := a.accums[spatial]
	if acc == nil {
		return -1
	}
	return acc.standing.continuousMin
}

func TestEventFrame_LastActive_WalkDistanceTriggers(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	now := int64(1_700_000_000_000)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: now, WalkDistanceMeters: 2})
	if got := snapshotLastActive(a, spatial); got != now {
		t.Errorf("walk_distance≥2m should trigger lastActive=now=%d, got %d", now, got)
	}
}

func TestEventFrame_LastActive_WalkDurationTriggers(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	now := int64(1_700_000_000_000)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: now, WalkDurationSec: 6})
	if got := snapshotLastActive(a, spatial); got != now {
		t.Errorf("walk_duration≥6s should trigger lastActive=now, got %d", got)
	}
}

func TestEventFrame_LastActive_BelowThresholdNoTrigger(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	now := int64(1_700_000_000_000)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: now, WalkDistanceMeters: 1, WalkDurationSec: 5})
	if got := snapshotLastActive(a, spatial); got != 0 {
		t.Errorf("walk<2m & duration<6s should NOT trigger lastActive, got %d", got)
	}
}

func TestEventFrame_LastActive_60sThrottle(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t1 := int64(1_700_000_000_000)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t1, WalkDistanceMeters: 5})
	// 30s 后再 push：被节流
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t1 + 30_000, WalkDistanceMeters: 5})
	if got := snapshotLastActive(a, spatial); got != t1 {
		t.Errorf("within 60s throttle: lastActive should stay at t1=%d, got %d", t1, got)
	}
	// 60s+ 后 push：通过
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t1 + 60_000, WalkDistanceMeters: 5})
	if got := snapshotLastActive(a, spatial); got != t1+60_000 {
		t.Errorf("after 60s: lastActive should update, got %d", got)
	}
}

func TestEventFrame_Standing_AccumulateWithGate(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, spatial, 1, t0)

	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 60_000, StandDurationSec: 58})
	if got := snapshotStanding(a, spatial); got != 1 {
		t.Errorf("first stand≥55 with tp=1: standing want 1, got %d", got)
	}
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 120_000, StandDurationSec: 58})
	if got := snapshotStanding(a, spatial); got != 2 {
		t.Errorf("second stand≥55: standing want 2, got %d", got)
	}
}

func TestEventFrame_Standing_Cap8(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, spatial, 1, t0)
	for i := 1; i <= 10; i++ {
		a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + int64(i)*60_000, StandDurationSec: 58})
	}
	if got := snapshotStanding(a, spatial); got != 8 {
		t.Errorf("standing should cap at 8 after 10 increments, got %d", got)
	}
}

func TestEventFrame_Standing_ResetOnShortStand(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, spatial, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 60_000, StandDurationSec: 58})
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 120_000, StandDurationSec: 58})
	// stand<55 → reset
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 180_000, StandDurationSec: 10})
	if got := snapshotStanding(a, spatial); got != 0 {
		t.Errorf("short stand should reset standing to 0, got %d", got)
	}
}

func TestEventFrame_Standing_ResetOnMultiPerson(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, spatial, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 60_000, StandDurationSec: 58})
	// MultiPersonDuration > 0 → reset 即便 stand≥55
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 120_000, StandDurationSec: 58, MultiPersonDurationSec: 30})
	if got := snapshotStanding(a, spatial); got != 0 {
		t.Errorf("multi_person_duration>0 should reset standing, got %d", got)
	}
}

func TestEventFrame_Standing_GateTotalPeopleZero(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	// 不调 setTotalPeople（默认 0）→ standing 不累加
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0, StandDurationSec: 58})
	if got := snapshotStanding(a, spatial); got != 0 {
		t.Errorf("tp=0 (no one in room): standing should stay 0, got %d", got)
	}
}

func TestEventFrame_Standing_GateTotalPeopleTwo(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	spatial := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, spatial, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 60_000, StandDurationSec: 58})
	// 切到 2 人 → standing reset
	setTotalPeople(a, spatial, 2, t0+90_000)
	a.handleEventFrame(EventFrame{SpatialPrefix: spatial, TsMs: t0 + 120_000, StandDurationSec: 58})
	if got := snapshotStanding(a, spatial); got != 0 {
		t.Errorf("tp=2: standing should reset, got %d", got)
	}
}

func TestEventFrame_EmptySpatialPrefixIgnored(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	a.handleEventFrame(EventFrame{SpatialPrefix: "", TsMs: 1_700_000_000_000, WalkDistanceMeters: 5})
	a.mu.RLock()
	n := len(a.accums)
	a.mu.RUnlock()
	if n != 0 {
		t.Errorf("empty SpatialPrefix should not create accumulator, got %d entries", n)
	}
}

// ============================================================================
// ForgetDevice — "device offline = 内存重启"
// ============================================================================

func TestForgetDevice_Clears96And88(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	deviceAddr := "fd00:0:3:111:3:101::1"
	// 用 netip canonical form 算 sp96 / sp88（与 ForgetDevice 内部算法一致）
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		t.Fatalf("setup: parse %q: %v", deviceAddr, err)
	}
	sp96 := netip.PrefixFrom(addr, 96).Masked().String()
	sp88 := netip.PrefixFrom(addr, 88).Masked().String()
	t0 := int64(1_700_000_000_000)

	// 在 /96 和 /88 都填 accumulator entry
	setTotalPeople(a, sp96, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: sp96, TsMs: t0, WalkDistanceMeters: 5})
	setTotalPeople(a, sp88, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: sp88, TsMs: t0, WalkDistanceMeters: 5})

	if snapshotLastActive(a, sp96) == 0 || snapshotLastActive(a, sp88) == 0 {
		t.Fatalf("setup: expected entries in both /96 (%s) and /88 (%s)", sp96, sp88)
	}

	a.ForgetDevice(deviceAddr)

	if _, ok := a.accums[sp96]; ok {
		t.Errorf("ForgetDevice: sp96 (%s) entry should be cleared", sp96)
	}
	if _, ok := a.accums[sp88]; ok {
		t.Errorf("ForgetDevice: sp88 (%s) entry should be cleared", sp88)
	}
}

func TestForgetDevice_InvalidAddrNoOp(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	sp := "fd00:0:3:111:3:101::/96"
	t0 := int64(1_700_000_000_000)
	setTotalPeople(a, sp, 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: sp, TsMs: t0, WalkDistanceMeters: 5})

	a.ForgetDevice("garbage")

	if snapshotLastActive(a, sp) == 0 {
		t.Error("invalid addr: should not affect existing entries")
	}
}

func TestForgetDevice_OtherSpatialUntouched(t *testing.T) {
	a := NewTargetStateAggregator(nil, zap.NewNop())
	t0 := int64(1_700_000_000_000)

	// dev-a 在 /96 sp-a
	setTotalPeople(a, "fd00:0:3:111:3:101::/96", 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: "fd00:0:3:111:3:101::/96", TsMs: t0, WalkDistanceMeters: 5})
	// dev-b 在 /96 sp-b（不同房间）
	setTotalPeople(a, "fd00:0:3:222:3:101::/96", 1, t0)
	a.handleEventFrame(EventFrame{SpatialPrefix: "fd00:0:3:222:3:101::/96", TsMs: t0, WalkDistanceMeters: 5})

	a.ForgetDevice("fd00:0:3:111:3:101::1") // 只清 dev-a

	if _, ok := a.accums["fd00:0:3:111:3:101::/96"]; ok {
		t.Error("dev-a sp should be cleared")
	}
	if _, ok := a.accums["fd00:0:3:222:3:101::/96"]; !ok {
		t.Error("dev-b sp should be untouched")
	}
}

