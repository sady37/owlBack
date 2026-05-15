package wiring

import (
	"net/netip"
	"strings"

	"owl-common/observation"
	"wisefido-sensor/internal/service"
)

// MonitorVitalSource 实现 zoneengine.VitalSource —— 包装 service.MonitorBuffer。
//
// 扫描逻辑：
//   1. 遍历所有 active card
//   2. 对每个 card snapshot，按 device 看任一 track 是否同时有 HR>0 且 RR>0
//   3. device 的 LastTs 在 freshness 窗内 → emit (cardID, devAddr/96 bedZoneID, ts)
//
// **Bed 派生约束**：device addr /96 = bed prefix，假设 device 物理上绑床。
// Non-bed-bound devices（如 living room 雷达）派生的 /96 是无意义的；它们的 sustain evidence
// 进 zone engine 后会创建一个永远不翻转的 Bed zone 实例（不会污染状态，仅占内存）。
// 后期如有 N 个非床雷达污染严重，可加 BedZoneLookup 过滤；目前不需要。
//
// device_type 不影响扫描逻辑（只看 HR/RR 是否 fresh）；source 字段统一填 "vital"。
type MonitorVitalSource struct {
	buf *service.MonitorBuffer
}

func NewMonitorVitalSource(buf *service.MonitorBuffer) *MonitorVitalSource {
	return &MonitorVitalSource{buf: buf}
}

// ScanActiveBedVitals satisfy zoneengine.VitalSource。
func (s *MonitorVitalSource) ScanActiveBedVitals(nowMs, freshnessMs int64, emit func(cardID, bedZoneID string, ts int64)) {
	if s.buf == nil || emit == nil {
		return
	}
	cards := s.buf.ActiveCardIDs()
	for _, cardID := range cards {
		snap := s.buf.SnapshotCard(cardID)
		if snap == nil {
			continue
		}
		for _, dev := range snap.Devices {
			// device-level freshness：dev 在 snapshot 里的所有 track 取 max ts；用 track 里嵌的 "ts" 字段
			var devLastTs int64
			var hasHRRR bool
			for _, fields := range dev.Tracks {
				if ts := tsFromTrackMap(fields); ts > devLastTs {
					devLastTs = ts
				}
				if hrRRPresent(fields) {
					hasHRRR = true
				}
			}
			if !hasHRRR {
				continue
			}
			if nowMs-devLastTs > freshnessMs {
				continue
			}
			bedZoneID := bedPrefixFromDeviceID(dev.DeviceID)
			if bedZoneID == "" {
				continue
			}
			emit(cardID, bedZoneID, devLastTs)
		}
	}
}

// hrRRPresent 单条 track fields 是否同时有 heart_rate>0 && respiratory_rate>0。
func hrRRPresent(fields map[string]any) bool {
	hr := positiveInt(fields[observation.FieldHeartRate])
	rr := positiveInt(fields[observation.FieldRespiratoryRate])
	return hr > 0 && rr > 0
}

func positiveInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return 0
	}
}

func tsFromTrackMap(fields map[string]any) int64 {
	if ts, ok := fields["ts"].(int64); ok {
		return ts
	}
	if ts, ok := fields["ts"].(float64); ok {
		return int64(ts)
	}
	return 0
}

// bedPrefixFromDeviceID device_id（/128 host text 或 CIDR）→ /96 CIDR text。
// 不合法 / 不能 parse → 返回 ""。
func bedPrefixFromDeviceID(deviceID string) string {
	if deviceID == "" {
		return ""
	}
	// device_id 通常是 /128 host text，无 CIDR mask；MonitorConsumer 写入用 addr.String()
	s := deviceID
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.IsValid() {
		return ""
	}
	return netip.PrefixFrom(addr, 96).Masked().String()
}
