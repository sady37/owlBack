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
// 详见 owlBack/doc/platform_agent_addressing.md（envelope 字段契约 §4.1）。
//
// 关键字段填充（与权威 doc 同步）：
//   - producer       sensor 自己的 /128 IPv6（在 AlarmBackChannel 内部填）
//   - device_addr    trigger device 的 /128（优先 trace SignalEvidence；fallback BedDeviceLookup；最后 ZoneID prefix）
//   - timestamp      Fire/Cancel: p.Trigger.Ts (实际触发瞬时；不是 Fire 决策时刻)
//                    Arm: p.ArmedAt（兼容旧字段，战役 3 删 Arm）
//   - sequence_number AlarmBackChannel.nextSeq 从 Redis INCR 取
//   - payload (DataValue) 含 trigger evidence trace + zone meta + engine_fire_ms
type BackChannelAlarmFirer struct {
	channel    *consumer.AlarmBackChannel
	bedLookup  *BedDeviceLookup // optional，nil 时退化为 ZoneID prefix fallback
	logger     *zap.Logger
}

func NewBackChannelAlarmFirer(channel *consumer.AlarmBackChannel, bedLookup *BedDeviceLookup, logger *zap.Logger) *BackChannelAlarmFirer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BackChannelAlarmFirer{channel: channel, bedLookup: bedLookup, logger: logger}
}

// Arm / Cancel：no-op（2026-05-15 platform-agent cutover）。
//
// sensor.Supervisor 自己维护 in-memory pending map + timer，到期直接 Fire。
// 不再 publish pending_arm/pending_cancel envelope（cardagg 也不再消费）。
// Supervisor 仍调本接口，仅作 hook；future cognitive layer 想监听可重新接入。
func (f *BackChannelAlarmFirer) Arm(ctx context.Context, p zonealarm.Pending) error {
	return nil
}

func (f *BackChannelAlarmFirer) Cancel(ctx context.Context, key zonealarm.PendingKey, reason zoneengine.ZoneEvent) error {
	return nil
}

// Fire start event_status，cardagg 走 PersistAlarmAndPublish。
//
// triggered_at = p.Trigger.Ts (bed→vacant / room→vacant 等实际瞬时)；
// engine_fire_ms = time.Now() 进 evidence，记录 30min 定时到期那一刻。
func (f *BackChannelAlarmFirer) Fire(ctx context.Context, p zonealarm.Pending) error {
	devAddr := f.pickDeviceAddr(p.Trigger, p.Key.AlarmType)
	triggerTsMs := p.Trigger.Ts
	if triggerTsMs == 0 {
		// 边界场景兜底（理论上 ZoneEvent 总带 Ts）
		triggerTsMs = p.ArmedAt
	}
	_, err := f.channel.PublishAlarmFire(ctx, devAddr, p.Key.CardID, p.Key.AlarmType, p.Level, triggerTsMs,
		buildTriggerData(p))
	return err
}

// pickDeviceAddr 选定本次 alarm 归因的 trigger device /128。
//
// 顺序：
//  1. ZoneEvent.Trace 反向遍历，找 SignalEvidence.TriggerData["device_addr"] → 真设备 /128 ✓
//  2. BedDeviceLookup 按 alarm 类型查 zone 内主 device（LeftBed→Radar 等，per [[track_fusion_and_gate_cardid]] / 用户拍板）
//  3. fallback ZoneID prefix host=0（最差，保 alarm 链不断；cardagg snapshot JOIN 拿不到 device 字段）
//
// 改自旧 fallback-only 版（zone_id host=0 导致 device_uid/resident_nickname 等 snapshot 全 NULL）。
func (f *BackChannelAlarmFirer) pickDeviceAddr(e zoneengine.ZoneEvent, alarmType string) netip.Addr {
	// 1) trace 里真设备 device_addr 优先
	for i := len(e.Trace) - 1; i >= 0; i-- {
		ev := e.Trace[i]
		if ev.TriggerData == nil {
			continue
		}
		if v, ok := ev.TriggerData["device_addr"].(string); ok {
			if addr, err := netip.ParseAddr(strings.TrimSpace(v)); err == nil && addr.IsValid() {
				return addr
			}
		}
	}
	// 2) BedDeviceLookup：按 alarm 类型 + zone 前缀查主 device
	if f.bedLookup != nil {
		if addr := f.bedLookup.FindPrimaryDevice(e.ZoneID, alarmType); addr.IsValid() {
			return addr
		}
	}
	// 3) fallback：zone_id CIDR text → addr 部分（host=0）
	if e.ZoneID != "" {
		if idx := strings.Index(e.ZoneID, "/"); idx >= 0 {
			if addr, err := netip.ParseAddr(e.ZoneID[:idx]); err == nil {
				return addr
			}
		}
	}
	return netip.Addr{}
}

// buildTriggerData zone-derived alarm 的 trigger_data。
//
// 落到 alarm_events.payload (JSONB)，前端 detail 页 + 后续 trace replay 用。
//
// 字段：
//   - derived_by:       'zonealarm'（写入方标记，cardagg 拿到能识别这条不是 device 直发）
//   - trigger_zone:     'bed' / 'room' / 'bathroom'
//   - trigger_zone_id:  CIDR
//   - trigger_status:   'occupied' / 'vacant'
//   - armed_at_ms:      sensor.Supervisor 起 pending 那一刻
//   - due_at_ms:        预期到期点
//   - engine_fire_ms:   sensor 实际 fire 那一刻（distinct from triggered_at = p.Trigger.Ts）
//   - trace:            [...] 触发本次翻转的 SignalEvidence 完整链（雷达 track id / sleepad 序号 / etc）
func buildTriggerData(p zonealarm.Pending) map[string]interface{} {
	out := map[string]interface{}{
		"derived_by":      "zonealarm",
		"trigger_zone":    p.Trigger.ZoneType.String(),
		"trigger_zone_id": p.Trigger.ZoneID,
		"trigger_status":  p.Trigger.NewState.Status.String(),
		"armed_at_ms":     p.ArmedAt,
		"due_at_ms":       p.DueAt,
		"engine_fire_ms":  time.Now().UnixMilli(),
	}
	if len(p.Trigger.Trace) > 0 {
		// 完整保留 trace evidence chain；fe detail 页 + audit / replay 复用
		out["trace"] = p.Trigger.Trace
	}
	return out
}
