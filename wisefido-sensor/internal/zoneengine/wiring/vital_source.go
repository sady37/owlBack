package wiring

import (
	"net/netip"
	"strings"

	"owl-common/observation"
	"wisefido-sensor/internal/service"
)

// MonitorVitalSource 实现 zoneengine.VitalSource —— 包装 service.MonitorBuffer。
//
// 扫描逻辑（2026-05-24 修订）：
//   1. 遍历所有 active card
//   2. 对每个 card snapshot，按 device 看：
//      a) device_type ∈ {Sleepad, SleepPad}（**仅 sleepad；radar 排除**）
//      b) **任一 track 有任一压感衍生信号**：HR>0 OR RR>0 OR body_move>0 OR turn_over>0 OR bed_status==0
//   3. device 的 LastTs 在 freshness 窗内 → emit (bedZoneID, ts)
//
// **设计约束**：
//   - sleepad 是接触式压感设备；HR/RR/body_move/turn_over 任一为正 + firmware bed_status==0 = 在床
//   - 任一字段缺失 ≠ 反证不在床（2026-05-20 拍板）；故合取改成析取
//   - **radar 排除**（radar 可在床外 detect 到坐沙发的人 HR/RR；不发 sustain）
//   - radar in-bed 由 firmware 的 InBed/LeftBed alarm 走 adapter_radar
//
// **历史 bug**（2026-05-24 修复）：
//   - 原 `HR>0 AND RR>0` 太严格 → 实际 sleepad 帧常单 HR 无 RR / 仅 body_move → 漏判
//   - 加 SustainStaleMs 60s 配合 4G 上报周期 30s（详 scorer.go const 注释）
//
// **Bed 派生约束**：device addr /96 = bed prefix，假设 device 物理上绑床。
// Non-bed-bound sleepad 应该不存在（sleepad 物理上就贴床）；如有 sustain 进 zone 不污染状态。
type MonitorVitalSource struct {
	buf *service.MonitorBuffer
}

func NewMonitorVitalSource(buf *service.MonitorBuffer) *MonitorVitalSource {
	return &MonitorVitalSource{buf: buf}
}

// ScanActiveBedVitals satisfy zoneengine.VitalSource。
// 物理寻址：emit (bed /96 CIDR, ts)，不依赖 card 概念。MonitorBuffer 的内部按 card 聚合
// 仍是 v1 残留 —— v2 zone engine 只需要 device→/96 派生。
func (s *MonitorVitalSource) ScanActiveBedVitals(nowMs, freshnessMs int64, emit func(bedZoneID string, ts int64)) {
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
			// 设计约束（2026-05-20）：sustain 仅信 sleepad；radar HR/RR 不证明 in-bed。
			if !isSleepad(dev.DeviceType) {
				continue
			}
			// device-level freshness：dev 在 snapshot 里的所有 track 取 max ts；用 track 里嵌的 "ts" 字段
			var devLastTs int64
			var hasPresence bool
			for _, fields := range dev.Tracks {
				if ts := tsFromTrackMap(fields); ts > devLastTs {
					devLastTs = ts
				}
				if anyPresenceEvidence(fields) {
					hasPresence = true
				}
			}
			if !hasPresence {
				continue
			}
			if nowMs-devLastTs > freshnessMs {
				continue
			}
			bedZoneID := bedPrefixFromDeviceAddr(dev.DeviceAddr)
			if bedZoneID == "" {
				continue
			}
			emit(bedZoneID, devLastTs)
		}
	}
}

// isSleepad device_type 是否属 sleepad 家族（不区分大小写 + 兼容 "SleepPad" 拼写）。
func isSleepad(deviceType string) bool {
	dt := strings.ToLower(deviceType)
	return dt == "sleepad" || dt == "sleeppad" || dt == "sleepace"
}

// anyPresenceEvidence 单条 track fields 是否有任一"床上有人持续压力"证据。
// sleepad 是接触式压感设备：心率/呼吸/体动/翻身任一为正 + firmware bed_status==0 都是在床信号。
// 单字段缺失常见（呼吸幅度小测不到 / 静卧只压感无动作），故合取改成析取。
func anyPresenceEvidence(fields map[string]any) bool {
	if positiveInt(fields[observation.FieldHeartRate]) > 0 {
		return true
	}
	if positiveInt(fields[observation.FieldRespiratoryRate]) > 0 {
		return true
	}
	if positiveInt(fields[observation.FieldBodyMove]) > 0 {
		return true
	}
	if positiveInt(fields[observation.FieldTurnOver]) > 0 {
		return true
	}
	// bed_status==0 (firmware 显式 "在床"，最直接信号)；==1 是离床，==8 是待机。
	if v, ok := fields[observation.FieldBedStatus]; ok {
		switch x := v.(type) {
		case int:
			if x == 0 {
				return true
			}
		case int64:
			if x == 0 {
				return true
			}
		case float64:
			if int(x) == 0 {
				return true
			}
		}
	}
	return false
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

// bedPrefixFromDeviceAddr device_addr（/128 host text 或 CIDR）→ /96 CIDR text。
// 不合法 / 不能 parse → 返回 ""。
func bedPrefixFromDeviceAddr(deviceAddr string) string {
	if deviceAddr == "" {
		return ""
	}
	// device_addr 通常是 /128 host text，无 CIDR mask；MonitorConsumer 写入用 addr.String()
	s := deviceAddr
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.IsValid() {
		return ""
	}
	return netip.PrefixFrom(addr, 96).Masked().String()
}
