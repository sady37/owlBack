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
	a.OnZoneEvent(spatialPrefix, spatialPrefix, 1, 1_700_000_000_000)
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
	a.OnZoneEvent(spatialPrefix, spatialPrefix, 2, 1_700_000_000_000)
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

	a.OnZoneEvent("fd00:0:3:111:3:101::/96", "fd00:0:3:111:3:101::/96", 1, 1_700_000_000_000)
	a.OnZoneEvent("fd00:0:3:111:3:102::/96", "fd00:0:3:111:3:102::/96", 1, 1_700_000_000_000)
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

