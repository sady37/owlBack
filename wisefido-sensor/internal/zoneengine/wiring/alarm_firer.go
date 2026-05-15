package wiring

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/zonealarm"
	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// BackChannelAlarmFirer 实现 zonealarm.AlarmFirer，把 zone-derived alarm 经
// consumer.AlarmBackChannel publish 到 iot:alarm:stream，cardagg alarm_handler 接收。
//
// 关键考虑：
//   - cardagg alarm_handler 期望每条 alarm 带有效 device_addr（用于反查 tenant + device_id）。
//     zone-derived alarm 没有"单一触发设备"——本实现从 ZoneEvent.Trace 取最近一条 SignalEvidence
//     的 source 设备（如果可解析）。无法解析时使用 ZoneID 派生的网络地址作 fallback；
//     fallback 在 cardagg 侧 metaCache.ResolveDeviceID 会失败 drop。
//
//   - Arm/Cancel 走 PublishPendingArm/PublishPendingCancel —— cardagg 对应的 A9 分支
//     调 AddPendingAlarm/RemovePendingAlarm 维护 pending Hash（zonealarm 自己也持
//     pending，双写并不冲突；zonealarm 是权威，cardagg 仅做 UI/通知层）。
type BackChannelAlarmFirer struct {
	channel *consumer.AlarmBackChannel
	logger  *zap.Logger
}

func NewBackChannelAlarmFirer(channel *consumer.AlarmBackChannel, logger *zap.Logger) *BackChannelAlarmFirer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BackChannelAlarmFirer{channel: channel, logger: logger}
}

// Arm pending_arm event_status，cardagg 走 AddPendingAlarm。
func (f *BackChannelAlarmFirer) Arm(ctx context.Context, p zonealarm.Pending) error {
	devAddr := pickDeviceAddr(p.Trigger)
	durationSec := int((p.DueAt - p.ArmedAt) / 1000)
	tsMs := time.Now().UnixMilli()
	_, err := f.channel.PublishPendingArm(ctx, devAddr, p.Key.CardID, p.Key.AlarmType, p.Level,
		durationSec, p.Key.AlarmType /* upgradeTo == self（无升级）*/, p.ArmedAt, tsMs,
		buildTriggerData(p))
	return err
}

// Cancel pending_cancel event_status。
func (f *BackChannelAlarmFirer) Cancel(ctx context.Context, key zonealarm.PendingKey, reason zoneengine.ZoneEvent) error {
	devAddr := pickDeviceAddr(reason)
	_, err := f.channel.PublishPendingCancel(ctx, devAddr, key.CardID, key.AlarmType, time.Now().UnixMilli())
	return err
}

// Fire start event_status，cardagg 走 PersistAlarmAndPublish。
func (f *BackChannelAlarmFirer) Fire(ctx context.Context, p zonealarm.Pending) error {
	devAddr := pickDeviceAddr(p.Trigger)
	tsMs := time.Now().UnixMilli()
	_, err := f.channel.PublishAlarmFire(ctx, devAddr, p.Key.CardID, p.Key.AlarmType, p.Level, tsMs,
		buildTriggerData(p))
	return err
}

// pickDeviceAddr 从 ZoneEvent.Trace 取一条 SignalEvidence 的设备地址。
//
// 解析顺序：
//   1. trace 反向遍历，找 TriggerData["device_addr"] string → ParseAddr
//   2. 都失败 → fallback：zone_id 派生（截 CIDR 的 host 段；与 cardagg 路径不一致但不 panic）
//   3. fallback 也失败 → invalid Addr，由 cardagg drop
//
// 当 ZoneEvent 来自 vital sustain（无 trigger device 字段）或 subset_invariant 自抬升时，
// 没有真正的"触发设备"——alarm 大概率被 cardagg 因 device 不在 metaCache drop。
// 这是已知 phase-1 限制；后续在 wiring 注入 BedDeviceLookup 改进。
func pickDeviceAddr(e zoneengine.ZoneEvent) netip.Addr {
	for i := len(e.Trace) - 1; i >= 0; i-- {
		ev := e.Trace[i]
		if ev.TriggerData == nil {
			continue
		}
		if v, ok := ev.TriggerData["device_addr"].(string); ok {
			if addr, err := netip.ParseAddr(v); err == nil && addr.IsValid() {
				return addr
			}
		}
	}
	// fallback：zone_id CIDR text → addr 部分（host=0）
	if e.ZoneID != "" {
		if idx := strings.Index(e.ZoneID, "/"); idx >= 0 {
			if addr, err := netip.ParseAddr(e.ZoneID[:idx]); err == nil {
				return addr
			}
		}
	}
	return netip.Addr{}
}

// buildTriggerData zone-derived alarm 的 trigger_data，包含 trigger ZoneEvent 摘要。
func buildTriggerData(p zonealarm.Pending) map[string]interface{} {
	return map[string]interface{}{
		"derived_by":      "zonealarm",
		"trigger_zone":    p.Trigger.ZoneType.String(),
		"trigger_zone_id": p.Trigger.ZoneID,
		"trigger_status":  p.Trigger.NewState.Status.String(),
		"armed_at_ms":     p.ArmedAt,
		"due_at_ms":       p.DueAt,
	}
}
