// target_state_aggregator_test.go — TargetStateAggregator T1 契约测试骨架。
//
// P2 scaffold：只测 lifecycle / push channel / pull snapshot 基础路径；
// P3/P4 业务字段填入后，按累加器分组补完整 table-driven test：
//   - TestAggregator_LastActive_*    (P3)
//   - TestAggregator_StandingMin_*   (P3)
//   - TestAggregator_Visitor_*       (P4)
//   - TestAggregator_WeakBio_*       (P4，含 80+ escalation rising edge)
//   - TestAggregator_DailyReset      (P5)

package service

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestAggregator(t *testing.T) *TargetStateAggregator {
	t.Helper()
	return NewTargetStateAggregator(nil, nil, zap.NewNop())
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
	// /80 unit prefix 应该自动派生（纯 IPv6 运算）
	// /96 fd00:0:3:111:3:101 截到 /80 = 前 5 段 = fd00:0:3:111:3
	wantUnit := "fd00:0:3:111:3::/80"
	if acc.unitPrefix != wantUnit {
		t.Errorf("unitPrefix = %q, want %q (派生自 spatialPrefix /96 截 /80)", acc.unitPrefix, wantUnit)
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

// TestAggregator_PushAlarmEvent_DropOwnEscalation 自家产 alarm 不再入队（防 loop）。
func TestAggregator_PushAlarmEvent_DropOwnEscalation(t *testing.T) {
	a := newTestAggregator(t)
	spatialPrefix := "fd00:0:3:111:3:101::/96"
	a.PushAlarmEvent(AlarmEventSnapshot{
		SpatialPrefix: spatialPrefix,
		AlarmType:     "WeakBiometricSignal",
		Producer:      EscalationProducerTag,
		RawValue:      60,
	})
	if _, ok := a.accums[spatialPrefix]; ok {
		t.Error("own escalation alarm should not create accumulator entry")
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

// TestAggregator_UnitMetaLookup_FillsUnitType callback 注入后 unit_type 被填入。
// 验证 sensor 通过 units 表（物理属性）拿 unit_type，不查 cards 表。
func TestAggregator_UnitMetaLookup_FillsUnitType(t *testing.T) {
	a := newTestAggregator(t)
	// mock unitMetaLookup（实际 wiring 时由 wiring 注入查 units 表的实现）
	a.SetUnitMetaLookup(func(unitPrefix string) int {
		if unitPrefix == "fd00:0:3:111:3::/80" {
			return 1 // UnitTypePrivate
		}
		return 0
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.Run(ctx)

	spatialPrefix := "fd00:0:3:111:3:101::/96"
	a.OnZoneEvent(spatialPrefix, spatialPrefix, 1, 1_700_000_000_000)
	time.Sleep(20 * time.Millisecond)

	a.mu.RLock()
	acc := a.accums[spatialPrefix]
	a.mu.RUnlock()
	if acc.unitType != 1 {
		t.Errorf("unitType = %d, want 1 (Private; via unitMetaLookup callback)", acc.unitType)
	}
}
