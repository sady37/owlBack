package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"

	"go.uber.org/zap"
)

// BedEventCoordinator 协调 Sleepad 床的 InBed/LeftBed 与 monitor 缓冲：不一致时短暂 pending，对齐或超时后落库。
type BedEventCoordinator struct {
	mu      sync.Mutex
	pending map[string]*pendingBedEvent // cardID -> 至多一条待决
}

type pendingBedEvent struct {
	kind       string
	tenantID   string
	cardID     string
	deviceID   string
	deviceUID  string
	deviceType string
	eventTs    int64
	deadlineMs int64
}

func NewBedEventCoordinator() *BedEventCoordinator {
	return &BedEventCoordinator{pending: make(map[string]*pendingBedEvent)}
}

func bedBoundHasSleepad(meta *CardMeta) bool {
	if meta == nil || meta.BedPref == "" {
		return false
	}
	for _, addr := range meta.BedBoundDeviceAddrs() {
		dm := meta.Devices[addr]
		if dm == nil {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad") {
			return true
		}
	}
	return false
}

func sleepadInBedTrackCount(buf *MonitorBuffer, meta *CardMeta) int {
	if buf == nil || meta == nil || meta.CardID == "" || meta.BedPref == "" {
		return 0
	}
	snap := buf.SnapshotCard(meta.CardID)
	if snap == nil {
		return 0
	}
	return SleepadTrackCountFromSnapshot(snap, meta, meta.BedBoundDeviceAddrs())
}

// ClearAll clears all pending bed events (used on reset).
func (c *BedEventCoordinator) ClearAll() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.pending = make(map[string]*pendingBedEvent)
	c.mu.Unlock()
}

// ClearCard 删卡或重置时清 pending。
func (c *BedEventCoordinator) ClearCard(cardID string) {
	if c == nil || cardID == "" {
		return
	}
	c.mu.Lock()
	delete(c.pending, cardID)
	c.mu.Unlock()
}

// InBed 返回 skipCaller：为 true 时调用方不要再 PublishBedStateFromEvent(InBed)。written 供 RemovePendingAlarm。
func (c *BedEventCoordinator) InBed(ctx context.Context, st *StateService, al *AlarmService, mc *DeviceMetaCache, buf *MonitorBuffer, cardID, tenantID, deviceID, deviceUID, deviceType string, ts int64, removeLeftBedPending func(written bool)) (skipCaller bool, written bool) {
	if c == nil || st == nil || mc == nil || buf == nil {
		return false, false
	}
	meta := mc.GetOrLoad(ctx, cardID)
	if meta == nil || !bedBoundHasSleepad(meta) {
		return false, false
	}
	count := sleepadInBedTrackCount(buf, meta)
	now := time.Now().UnixMilli()
	if count >= 1 {
		wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.InBed, deviceType, ts, 0, &count)
		if removeLeftBedPending != nil {
			removeLeftBedPending(wr)
		}
		if wr {
			_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
		}
		c.ClearCard(cardID)
		return true, wr
	}
	c.mu.Lock()
	c.pending[cardID] = &pendingBedEvent{
		kind: alarm.InBed, tenantID: tenantID, cardID: cardID, deviceID: deviceID, deviceUID: deviceUID,
		deviceType: deviceType, eventTs: ts, deadlineMs: now + BedEventPendingDeadlineMs,
	}
	c.mu.Unlock()
	return true, false
}

// LeftBed 返回 skipCaller / written；defer 时不写床态。written 时调用方勿重复 Reconcile（本处已调）。
func (c *BedEventCoordinator) LeftBed(ctx context.Context, st *StateService, mc *DeviceMetaCache, buf *MonitorBuffer, cardID, tenantID, deviceID, deviceUID, deviceType string, ts int64, afterLeftBed func()) (skipCaller bool, written bool) {
	if c == nil || st == nil || mc == nil || buf == nil {
		return false, false
	}
	meta := mc.GetOrLoad(ctx, cardID)
	if meta == nil || !bedBoundHasSleepad(meta) {
		return false, false
	}
	count := sleepadInBedTrackCount(buf, meta)
	now := time.Now().UnixMilli()
	if count == 0 {
		z := 0
		wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.LeftBed, deviceType, ts, 0, &z)
		if wr {
			_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
			if afterLeftBed != nil {
				afterLeftBed()
			}
		}
		c.ClearCard(cardID)
		return true, wr
	}
	c.mu.Lock()
	c.pending[cardID] = &pendingBedEvent{
		kind: alarm.LeftBed, tenantID: tenantID, cardID: cardID, deviceID: deviceID, deviceUID: deviceUID,
		deviceType: deviceType, eventTs: ts, deadlineMs: now + BedEventPendingDeadlineMs,
	}
	c.mu.Unlock()
	return true, false
}

// TryResolveAfterMonitorWrite 每条 iot:monitor 写入 buffer 后调用：该卡若有 pending 立即重判（不必等 1s Tick）。
func (c *BedEventCoordinator) TryResolveAfterMonitorWrite(ctx context.Context, st *StateService, al *AlarmService, mc *DeviceMetaCache, buf *MonitorBuffer, cardID string, logger *zap.Logger) {
	if c == nil || cardID == "" {
		return
	}
	c.processPendingForCard(ctx, st, al, mc, buf, cardID, time.Now().UnixMilli(), logger)
}

// Tick 周期性扫全卡 pending（与 TryResolveAfterMonitorWrite 共用逻辑）。
func (c *BedEventCoordinator) Tick(ctx context.Context, st *StateService, al *AlarmService, mc *DeviceMetaCache, buf *MonitorBuffer, logger *zap.Logger) {
	if c == nil || st == nil || mc == nil || buf == nil {
		return
	}
	c.mu.Lock()
	ids := make([]string, 0, len(c.pending))
	for id := range c.pending {
		ids = append(ids, id)
	}
	c.mu.Unlock()

	now := time.Now().UnixMilli()
	for _, cardID := range ids {
		c.processPendingForCard(ctx, st, al, mc, buf, cardID, now, logger)
	}
}

func (c *BedEventCoordinator) processPendingForCard(ctx context.Context, st *StateService, al *AlarmService, mc *DeviceMetaCache, buf *MonitorBuffer, cardID string, now int64, logger *zap.Logger) {
	if c == nil || st == nil || mc == nil || buf == nil {
		return
	}
	c.mu.Lock()
	p, ok := c.pending[cardID]
	c.mu.Unlock()
	if !ok || p == nil {
		return
	}
	meta := mc.GetOrLoad(ctx, cardID)
	if meta == nil || !bedBoundHasSleepad(meta) {
		c.ClearCard(cardID)
		return
	}
	count := sleepadInBedTrackCount(buf, meta)

	switch p.kind {
	case alarm.InBed:
		if count >= 1 {
			wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.InBed, p.deviceType, p.eventTs, 0, &count)
			if wr && al != nil {
				_ = al.RemovePendingAlarm(ctx, p.tenantID, cardID, p.deviceID, alarm.LeftBed)
			}
			if wr {
				_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
			}
			c.ClearCard(cardID)
			if logger != nil && wr {
				logger.Debug("bed pending resolved InBed (aligned)", zap.String("cid", cardID), zap.Int("count", count))
			}
			return
		}
		if now < p.deadlineMs {
			return
		}
		wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.InBed, p.deviceType, p.eventTs, 0, nil)
		if wr && al != nil {
			_ = al.RemovePendingAlarm(ctx, p.tenantID, cardID, p.deviceID, alarm.LeftBed)
		}
		if wr {
			_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
		}
		c.ClearCard(cardID)
		if logger != nil {
			logger.Debug("bed pending resolved InBed (timeout trust event)", zap.String("cid", cardID), zap.Bool("written", wr))
		}

	case alarm.LeftBed:
		if count == 0 {
			z := 0
			wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.LeftBed, p.deviceType, p.eventTs, 0, &z)
			if wr {
				_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
				if al != nil && meta.TenantPref != "" {
					PersistSuspectedFallPoseLyingIfEnabled(ctx, al, cardID, meta.TenantPref, buf, meta, "ImmediateLeftBedFall_lying_pending")
				}
				if radarID := RadarDeviceIDBoundToBed(meta); radarID != "" {
					StartLeftBedFall(cardID, radarID)
				}
			}
			c.ClearCard(cardID)
			if logger != nil && wr {
				logger.Debug("bed pending resolved LeftBed (aligned)", zap.String("cid", cardID))
			}
			return
		}
		if now < p.deadlineMs {
			return
		}
		wr, _ := st.PublishBedStateFromEvent(ctx, cardID, alarm.LeftBed, p.deviceType, p.eventTs, 0, nil)
		if wr {
			_ = st.ReconcileRoomStateFromBedState(ctx, cardID)
			if al != nil && meta.TenantPref != "" {
				PersistSuspectedFallPoseLyingIfEnabled(ctx, al, cardID, meta.TenantPref, buf, meta, "ImmediateLeftBedFall_lying_pending_timeout")
			}
			if radarID := RadarDeviceIDBoundToBed(meta); radarID != "" {
				StartLeftBedFall(cardID, radarID)
			}
		}
		c.ClearCard(cardID)
		if logger != nil {
			logger.Debug("bed pending resolved LeftBed (timeout trust event)", zap.String("cid", cardID), zap.Bool("written", wr))
		}
	}
}
