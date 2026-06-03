// monitor_handler.go — iot:monitor:stream 消费者。
//
// 四件薄事（CLAUDE.md 规则 #2.4 maintainer 模式）：
//   1. tid=88 (TrackInvalid) → AIOverrideCache.ClearDevice（firmware no-target heartbeat
//      = 该 device 当前无 track；旧 verdict 全部作废避免新 track_id 复用旧 ghost 判定）
//   2. AIOverrideCache.Apply 把 sensor verdict (confidence/source) 合到 track fields
//      （release 模式生效；sandbox 模式仅 log）— **必须在 buffer.Write 前**，让合并值进 snapshot
//   3. buffer.Write 把 sample 累入 MonitorBuffer（12s TTL，per-card per-device per-track）
//   4. deviceTracker.TouchLastSeen 触刷设备活跃（180s 看门狗输入）
//
// RunLoop 1s tick → buffer.Flush snapshot → card.Writer.PublishMonitor → card:realtime:stream
//
// 不做：派生 / 融合 / Stay FSM / bed event 协调（全 sensor 或 data SSE 层管）

package consumer

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"owl-common/card"
	"owl-common/observation"
	owlredis "owl-common/redis"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

const (
	monitorFieldTTLMs       = 12_000
	monitorMaxInboundAgeMs  = 6_000
	monitorSnapshotInterval = 1 * time.Second
	monitorPruneEvery       = 4
)

// DeviceLivenessTouch monitor handler 仅需的设备活跃刷新接口；DeviceStatusTracker 实现。
type DeviceLivenessTouch interface {
	TouchLastSeen(ctx context.Context, deviceAddr, deviceType string)
}

type MonitorHandler struct {
	buffer      *service.MonitorBuffer
	writer      *card.Writer
	devTouch    DeviceLivenessTouch
	cache       *service.SpatialCache    // device_addr → cardID 1-shot 反查
	aiOverrides *service.AIOverrideCache // 可 nil（未 wire 时 Apply 退化为 no-op）
	logger      *zap.Logger
}

// NewMonitorHandler 注入 cache —— envelope.SubjectEntity 承载 device_uid 不是 cardID；
// buffer 聚合按 cardID 索引，需要 device_addr → cardID 反查。
func NewMonitorHandler(buffer *service.MonitorBuffer, writer *card.Writer, devTouch DeviceLivenessTouch, cache *service.SpatialCache, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{
		buffer:   buffer,
		writer:   writer,
		devTouch: devTouch,
		cache:    cache,
		logger:   logger,
	}
}

// SetAIOverrides main wiring 注入 AI 裁决合并缓存；nil 时本 handler 不做 verdict 合并。
func (h *MonitorHandler) SetAIOverrides(c *service.AIOverrideCache) {
	if h == nil {
		return
	}
	h.aiOverrides = c
}

// RunLoop 1s snapshot + 4s prune。
func (h *MonitorHandler) RunLoop(ctx context.Context) {
	ticker := time.NewTicker(monitorSnapshotInterval)
	defer ticker.Stop()
	tick := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick++
			nowMs := time.Now().UnixMilli()
			if tick%monitorPruneEvery == 0 {
				h.buffer.PruneFields(nowMs, monitorFieldTTLMs)
			}
			snapshots := h.buffer.Flush(nowMs)
			for _, snap := range snapshots {
				if err := h.publishSnap(ctx, snap); err != nil {
					h.logger.Warn("publish realtime", zap.String("cid", snap.CardID), zap.Error(err))
				}
			}
		}
	}
}

func (h *MonitorHandler) publishSnap(ctx context.Context, snap service.CardSnapshot) error {
	data := make(map[string]any, len(snap.Devices))
	for _, dev := range snap.Devices {
		data[dev.DeviceAddr] = dev.Tracks
	}
	return h.writer.PublishMonitor(ctx, snap.CardID, "", data)
}

func (h *MonitorHandler) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if !msg.DeviceAddr.IsValid() {
		return nil
	}
	if time.Now().UnixMilli()-msg.Timestamp > monitorMaxInboundAgeMs {
		return nil
	}
	fields := owlredis.FirstDataValue(msg.DataValue)
	if fields == nil {
		return nil
	}
	deviceAddr := msg.DeviceAddr.String()
	trackID := resolveTrackID(fields)
	isNoTarget := trackID == observation.TrackInvalid // 88 = firmware 无人心跳
	if h.aiOverrides != nil {
		if isNoTarget {
			// firmware no-target heartbeat：当前 device 无 track，旧 verdict 全清
			h.aiOverrides.ClearDevice(deviceAddr)
		} else {
			h.aiOverrides.Apply(deviceAddr, trackID, fields)
		}
	}
	// buffer 按 cardID 聚合：envelope.SubjectEntity = device_uid 不是 cardID；
	// 通过 SpatialCache entry 表反查 device_addr → card_id。
	cardID := ""
	if h.cache != nil {
		if owner := h.cache.LookupCardByDevice(msg.DeviceAddr); owner.IsValid() {
			cardID = owner.String()
		}
	}
	// 88 无人心跳：人已不在 → 立即清掉该 device 旧 track（否则 card:realtime 用新 ts_ms 把旧位置一直
	// 重发，前端残影 >60s），且 88 本身不入 buffer。device online 心跳仍维护（88 是 online 信号）。
	// 前端 6s presence 窗会平滑帧间偶发 88，不闪；持续 88（人真走）→ 6s 后前端自然清。
	if isNoTarget {
		if cardID != "" {
			h.buffer.ClearDeviceTracks(cardID, deviceAddr)
		}
		h.devTouch.TouchLastSeen(ctx, deviceAddr, msg.DeviceType)
		return nil
	}
	if cardID == "" {
		// 未绑卡（pool 设备、unbound）— monitor 数据无 card 可归属，跳 buffer.Write。
		// 仍触发 liveness 让 device 状态正常更新。
		h.devTouch.TouchLastSeen(ctx, deviceAddr, msg.DeviceType)
		return nil
	}
	h.buffer.Write(cardID, deviceAddr, strconv.Itoa(trackID), fields, msg.Timestamp)
	h.devTouch.TouchLastSeen(ctx, deviceAddr, msg.DeviceType)
	return nil
}

func resolveTrackID(fields map[string]interface{}) int {
	v, ok := fields["track_id"]
	if !ok || v == nil {
		return observation.TrackInvalid
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return observation.TrackInvalid
		}
		return int(n)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return observation.TrackInvalid
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return observation.TrackInvalid
		}
		return n
	default:
		return observation.TrackInvalid
	}
}
