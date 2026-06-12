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
//   2. 对每个 card snapshot，按 device 看：
//      a) sleepad（接触式）或 radar
//      b) **任一 track 有任一压感/生命衍生信号**：HR>0 OR RR>0 OR body_move>0 OR turn_over>0 OR bed_status==0
//   3. device 的 LastTs 在 freshness 窗内 → emit (bedZoneID, source, ts)
//
// **归床**：
//   - sleepad：device addr /96 = bed prefix（sleepad 物理贴床，device 即床）。
//   - radar：HR/RR 由 firmware bed-enter 门控（只在 track 进床才开 sleep_monitor），存在即在床；
//     radar 自己 /96 ≠ 床 /96 → 经 BedResolver(radarAddr, roomPref) 归床（复用 adapter_radar 同款）。
//     多床候选未定 → BedResolver 返回 "" → 不发，让位每床自己的 sleepad（接触式权威）。
//
// **置信层级**：radar sustain 的 LR = lrRadarVital(ln1.75) 是所有正向腿里最弱，叠 γ 冲突 tempering +
// CoversWeight → 永远压不过有把握的 sleepad（接触式 > 雷达）。
//
// **设计约束**：HR/RR/body_move/turn_over 任一为正 + firmware bed_status==0 = 在床；
// 任一字段缺失 ≠ 反证不在床（2026-05-20 拍板）；故合取改成析取。
type MonitorVitalSource struct {
	buf      *service.MonitorBuffer
	resolver bedResolver
}

// bedResolver radar HR/RR 归床（radar 设备 + 所在 /88 → 它覆盖的床 /96）。MatrixCache 满足。
type bedResolver interface {
	ResolveBed(radarAddr, roomPref string) (string, float64)
}

func NewMonitorVitalSource(buf *service.MonitorBuffer) *MonitorVitalSource {
	return &MonitorVitalSource{buf: buf}
}

// SetBedResolver main wiring 注入（MatrixCache）；不调时 radar device 无法归床，radar vital 静默跳过。
func (s *MonitorVitalSource) SetBedResolver(r bedResolver) {
	s.resolver = r
}

// ScanActiveBedVitals satisfy zoneengine.VitalSource。
// 物理寻址：emit (bed /96 CIDR, ts)，不依赖 card 概念。MonitorBuffer 的内部按 card 聚合
// 仍是 v1 残留 —— v2 zone engine 只需要 device→/96 派生。
func (s *MonitorVitalSource) ScanActiveBedVitals(nowMs, freshnessMs int64, emit func(bedZoneID, source string, ts int64)) {
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
			isSp := isSleepad(dev.DeviceType)
			isRadar := strings.EqualFold(dev.DeviceType, "radar")
			if !isSp && !isRadar {
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
			var bedZoneID, source string
			if isSp {
				bedZoneID = bedPrefixFromDeviceAddr(dev.DeviceAddr)
				source = "vital"
			} else {
				// radar：HR/RR 由 firmware bed-enter 门控，存在即在床；经 BedResolver 归床。
				// 多床候选未定 → "" → 不发，让位 sleepad。
				if s.resolver == nil {
					continue
				}
				bedZoneID, _ = s.resolver.ResolveBed(dev.DeviceAddr, roomPrefixFromDeviceAddr(dev.DeviceAddr))
				source = "radar"
			}
			if bedZoneID == "" {
				continue
			}
			emit(bedZoneID, source, devLastTs)
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

// roomPrefixFromDeviceAddr device_addr → /88 room CIDR text（喂 BedResolver 的 roomPref）。
func roomPrefixFromDeviceAddr(deviceAddr string) string {
	if deviceAddr == "" {
		return ""
	}
	s := deviceAddr
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.IsValid() {
		return ""
	}
	return netip.PrefixFrom(addr, 88).Masked().String()
}
