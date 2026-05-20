// target_merger_test.go — S3 (FU5) TargetMerger T1 单元测试。
//
// 覆盖 sensor v2 entity-keyed (CIDR subject) 与 v3 device-keyed (/128) 双路径，
// + mergeForCard 字段级合并 + Standing 2min staleness + Visitor 注入。
//
// metaCache 用 nil DB + 直接 seed cards map（同 package 私字段可访问）。

package service

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"owl-common/card"

	"go.uber.org/zap"
)

func newTestMetaCache(t *testing.T, cardID string, deviceAddrs []string) *DeviceMetaCache {
	t.Helper()
	mc := NewDeviceMetaCache(nil, zap.NewNop())
	devs := make(map[string]*DeviceMeta, len(deviceAddrs))
	for _, addrStr := range deviceAddrs {
		addr, _ := netip.ParseAddr(addrStr)
		dm := &DeviceMeta{DeviceAddr: addr, RuntimeStatus: make(map[string]string)}
		dm.fillFromAddr()
		devs[addrStr] = dm
	}
	mc.cards[cardID] = &CardMeta{
		CardID:   cardID,
		Devices:  devs,
		dbLoaded: true,
	}
	return mc
}

// -----------------------------------------------------------------------------
// OnDeviceTarget — v2 entity-keyed (CIDR subject) path
// -----------------------------------------------------------------------------

func TestOnDeviceTarget_CIDRSubject_TreatsAsCardID(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{LastActiveTs: now, WeakBiometricSignal: 40, UpdatedAt: now}
	gotID, merged := m.OnDeviceTarget(context.Background(), cardID, ts)

	if gotID != cardID {
		t.Errorf("ownerCardID = %q, want %q (entity = card)", gotID, cardID)
	}
	if merged == nil {
		t.Fatal("merged should not be nil")
	}
	if merged.LastActiveTs != now || merged.WeakBiometricSignal != 40 {
		t.Errorf("merged fields not applied: %+v", merged)
	}
}

func TestOnDeviceTarget_EmptySubject_ReturnsEmpty(t *testing.T) {
	m := NewTargetMerger(NewDeviceMetaCache(nil, zap.NewNop()))
	gotID, merged := m.OnDeviceTarget(context.Background(), "", &card.TargetState{})
	if gotID != "" || merged != nil {
		t.Errorf("empty subject should return (\"\", nil), got (%q, %+v)", gotID, merged)
	}
}

func TestOnDeviceTarget_InvalidSubject_ReturnsEmpty(t *testing.T) {
	m := NewTargetMerger(NewDeviceMetaCache(nil, zap.NewNop()))
	gotID, merged := m.OnDeviceTarget(context.Background(), "garbage", &card.TargetState{})
	if gotID != "" || merged != nil {
		t.Errorf("invalid subject should return (\"\", nil), got (%q, %+v)", gotID, merged)
	}
}

func TestOnDeviceTarget_NilTs_ReturnsEmpty(t *testing.T) {
	m := NewTargetMerger(NewDeviceMetaCache(nil, zap.NewNop()))
	gotID, merged := m.OnDeviceTarget(context.Background(), "fd00:0:3:111:3:101::/96", nil)
	if gotID != "" || merged != nil {
		t.Errorf("nil ts should return (\"\", nil), got (%q, %+v)", gotID, merged)
	}
}

// -----------------------------------------------------------------------------
// mergeForCard — entity-keyed snapshot 字段透传
// -----------------------------------------------------------------------------

func TestMergeForCard_EntityKeyed_LastActiveAndWeakBio(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{LastActiveTs: now - 10_000, WeakBiometricSignal: 75, UpdatedAt: now}
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.mergeForCard(context.Background(), cardID, visitorFields{})
	if merged.LastActiveTs != now-10_000 {
		t.Errorf("LastActiveTs = %d, want %d", merged.LastActiveTs, now-10_000)
	}
	if merged.WeakBiometricSignal != 75 {
		t.Errorf("WeakBio = %d, want 75", merged.WeakBiometricSignal)
	}
}

func TestMergeForCard_EntityKeyed_StandingFreshApplies(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{StandingContinuousMin: 5, UpdatedAt: now - 30_000} // 30s 内 fresh
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.mergeForCard(context.Background(), cardID, visitorFields{})
	if merged.StandingContinuousMin != 5 {
		t.Errorf("fresh standing should apply, got %d want 5", merged.StandingContinuousMin)
	}
}

func TestMergeForCard_EntityKeyed_StandingStaleFiltered(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{StandingContinuousMin: 5, UpdatedAt: now - 3*60*1000} // 3min stale > 2min 阈值
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.mergeForCard(context.Background(), cardID, visitorFields{})
	if merged.StandingContinuousMin != 0 {
		t.Errorf("stale standing should be filtered to 0, got %d", merged.StandingContinuousMin)
	}
}

func TestMergeForCard_VisitorInjected(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{LastActiveTs: now - 60_000, UpdatedAt: now}
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.ApplyVisitor(context.Background(), cardID, MakeVisitorFields(now-300_000, true))
	if merged.LastActiveTs != now-60_000 {
		t.Errorf("LastActive lost after visitor apply: %d", merged.LastActiveTs)
	}
	if !merged.HasVisitorToday || merged.VisitorStartTs != now-300_000 {
		t.Errorf("visitor not applied: %+v", merged)
	}
}

func TestMergeForCard_NoSnapshot_ReturnsEmptyVisitorOnly(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	// 仅 visitor，无 sensor snapshot
	merged := m.ApplyVisitor(context.Background(), cardID, MakeVisitorFields(1_700_000_000_000, true))
	if merged.LastActiveTs != 0 || merged.StandingContinuousMin != 0 || merged.WeakBiometricSignal != 0 {
		t.Errorf("no sensor snapshot but sensor fields populated: %+v", merged)
	}
	if !merged.HasVisitorToday {
		t.Error("visitor should still apply")
	}
}

// W3: WeakBio 30min staleness — 老 snapshot (>30min) 不参与 merge
func TestMergeForCard_EntityKeyed_WeakBioStaleFiltered(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{WeakBiometricSignal: 85, UpdatedAt: now - 31*60*1000} // 31min stale > 30min 阈值
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.mergeForCard(context.Background(), cardID, visitorFields{})
	if merged.WeakBiometricSignal != 0 {
		t.Errorf("stale WeakBio (31min) should be filtered to 0, got %d", merged.WeakBiometricSignal)
	}
}

func TestMergeForCard_EntityKeyed_WeakBioFreshApplies(t *testing.T) {
	cardID := "fd00:0:3:111:3:101::/96"
	mc := newTestMetaCache(t, cardID, nil)
	m := NewTargetMerger(mc)

	now := time.Now().UnixMilli()
	ts := &card.TargetState{WeakBiometricSignal: 85, UpdatedAt: now - 25*60*1000} // 25min fresh < 30min
	m.OnDeviceTarget(context.Background(), cardID, ts)

	merged := m.mergeForCard(context.Background(), cardID, visitorFields{})
	if merged.WeakBiometricSignal != 85 {
		t.Errorf("fresh WeakBio (25min) should apply 85, got %d", merged.WeakBiometricSignal)
	}
}
