// monitor_handler.go — iot:monitor:stream 消费者。
//
// 三件薄事（CLAUDE.md 规则 #2.4 maintainer 模式）：
//   1. buffer.Write 把 sample 累入 MonitorBuffer（12s TTL，per-card per-device per-track）
//   2. deviceTracker.TouchLastSeen 触刷设备活跃（180s 看门狗输入）
//   3. RunLoop 1s tick → buffer.Flush snapshot → card.Writer.PublishMonitor → card:realtime:stream
//
// 不做：派生 / 融合 / Stay FSM / bed event 协调 / AI override（全 sensor 或 data SSE 层管）

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
	buffer    *service.MonitorBuffer
	writer    *card.Writer
	devTouch  DeviceLivenessTouch
	logger    *zap.Logger
}

func NewMonitorHandler(buffer *service.MonitorBuffer, writer *card.Writer, devTouch DeviceLivenessTouch, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{
		buffer:   buffer,
		writer:   writer,
		devTouch: devTouch,
		logger:   logger,
	}
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
		data[dev.DeviceID] = dev.Tracks
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
	h.buffer.Write(msg.SubjectEntity, deviceAddr, strconv.Itoa(trackID), fields, msg.Timestamp)
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
