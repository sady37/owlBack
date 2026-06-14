package wiring

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/zonealarm"

	"go.uber.org/zap"
)

// BackChannelAlarmFirer 实现 zonealarm.AlarmFirer (state-anchored 重构后只剩 Fire)。
// 把 zone-derived alarm 经 consumer.AlarmBackChannel publish 到 iot:alarm:stream，
// cardagg alarm_handler 接收。
//
// 详见 owlBack/doc/platform_agent_addressing.md（envelope 字段契约 §4.1）。
//
// 关键字段填充：
//   - producer       sensor 自己的 /128 IPv6（AlarmBackChannel 内部填）
//   - device_addr    SpatialCache.FindPrimaryDevice 按 alarm 类型查 zone 主 device；fallback zone CIDR host=0
//   - timestamp      e.AnchorTs（state-change-anchored 锚点；alone/leftbed/vacant 起始时刻）
//   - sequence_number AlarmBackChannel.nextSeq Redis INCR
//   - payload         derived_by/trigger_zone/anchor_ts_ms/fire_at_ms（trace evidence chain 已废）
type BackChannelAlarmFirer struct {
	channel *consumer.AlarmBackChannel
	spatial *SpatialCache // optional，nil 时退化 CIDR prefix fallback
	logger  *zap.Logger
}

func NewBackChannelAlarmFirer(channel *consumer.AlarmBackChannel, spatial *SpatialCache, logger *zap.Logger) *BackChannelAlarmFirer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BackChannelAlarmFirer{channel: channel, spatial: spatial, logger: logger}
}

// Fire — start event_status，cardagg 走 PersistAlarmAndPublish。
//
// triggered_at = e.AnchorTs（state-change-anchored 起点；如 LeftBed = BedStatusTs，
// Stay = AloneSinceTs，NightAbsence = LastExitTs）；
// fire_at_ms = time.Now() 进 evidence，记录 evaluator 阈值到那一刻。
//
// subject_entity 留空：sensor 物理层不持 card 概念；cardagg alarm_router 拿 device_addr
// 走 LPM 反查 cards 表自家归属。详 owlBack/doc/platform_agent_addressing.md §4.1。
func (f *BackChannelAlarmFirer) Fire(ctx context.Context, e zonealarm.FireEvent) error {
	devAddr := f.pickDeviceAddr(e.Key.ZoneID, e.Key.AlarmType)
	triggerTsMs := e.AnchorTs
	if triggerTsMs == 0 {
		triggerTsMs = e.NowMs
	}
	_, err := f.channel.PublishAlarmFire(ctx, devAddr, "", e.Key.AlarmType, triggerTsMs,
		buildTriggerData(e))
	return err
}

// pickDeviceAddr 选定本次 alarm 归因的 trigger device /128。
//
// state-anchored 重构后无 ZoneEvent.Trace 入口；仅剩两档：
//  1. SpatialCache.FindPrimaryDevice 按 alarm 类型查 zone 内主 device（LeftBed→Radar 等）
//  2. fallback ZoneID prefix host=0（最差，保 alarm 链不断；cardagg snapshot JOIN 拿不到 device 字段）
func (f *BackChannelAlarmFirer) pickDeviceAddr(zoneID, alarmType string) netip.Addr {
	if f.spatial != nil {
		if addr := f.spatial.FindPrimaryDevice(zoneID, alarmType); addr.IsValid() {
			return addr
		}
	}
	if zoneID != "" {
		if idx := strings.Index(zoneID, "/"); idx >= 0 {
			if addr, err := netip.ParseAddr(zoneID[:idx]); err == nil {
				return addr
			}
		}
	}
	return netip.Addr{}
}

// buildTriggerData zone-derived alarm 的 trigger_data 落到 alarm_events.payload (JSONB)。
//
// 字段：
//   - derived_by:    'zonealarm'
//   - trigger_zone:  CIDR（/96 bed / /88 room）
//   - alarm_type:    冗余便于查询
//   - anchor_ts_ms:  state-change-anchored 起点（计时起点）
//   - fire_at_ms:    evaluator 阈值到的那一刻
func buildTriggerData(e zonealarm.FireEvent) map[string]interface{} {
	return map[string]interface{}{
		"derived_by":   "zonealarm",
		"trigger_zone": e.Key.ZoneID,
		"alarm_type":   e.Key.AlarmType,
		"anchor_ts_ms": e.AnchorTs,
		"fire_at_ms":   time.Now().UnixMilli(),
	}
}
