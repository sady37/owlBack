package consumer

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"owl-common/card"
	"owl-common/observation"
	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type MonitorHandler struct {
	buffer    *service.MonitorBuffer
	writer    *card.Writer
	metaCache *service.DeviceMetaCache
	resolver  *service.DeviceCardResolver
	logger    *zap.Logger
}

const (
	MonitorFieldTTL = 6000 // ms, 6s prune / drop message older than 4s
)

// 实时流不做报警处理：雷达/sleepace 已有阀值，立即报警型由网关直接发 iot:alarm:stream。
func NewMonitorHandler(buffer *service.MonitorBuffer, writer *card.Writer, metaCache *service.DeviceMetaCache, resolver *service.DeviceCardResolver, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{buffer: buffer, writer: writer, metaCache: metaCache, resolver: resolver, logger: logger}
}

// RunLoop 与 monitor 相关的定时：1 秒发 snap，6 秒 PruneFields。不包含 derive。
// 写优先：PruneFields 用写锁、Flush 用读锁；两次加锁之间释放锁，Handle.Write 可插入。
func (h *MonitorHandler) RunLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick++
			nowMs := time.Now().UnixMilli()
			if tick%4 == 0 {
				h.buffer.PruneFields(nowMs, MonitorFieldTTL)
				activeDevs := h.buffer.ActiveDevicesByCard()
				for cid, devSet := range activeDevs {
					h.buffer.AdvancePruneTick(cid, devSet)
				}
			}
			// Flush 用 RLock，与 Write 不互斥；PruneFields 与 Write 之间已释放锁，写可优先
			snapshots := h.buffer.Flush(nowMs)
			for _, snap := range snapshots {
				if err := publishRealtimeSnap(ctx, h.writer, snap, h.logger); err != nil {
					h.logger.Warn("publish realtime", zap.String("cid", snap.CardID), zap.Error(err))
				}
			}
		}
	}
}

func publishRealtimeSnap(ctx context.Context, w *card.Writer, snap service.CardSnapshot, logger *zap.Logger) error {
	data := make(map[string]any, len(snap.Devices))
	for _, dev := range snap.Devices {
		data[dev.DeviceID] = dev.Tracks
	}
	if err := w.PublishMonitor(ctx, snap.CardID, "", data); err != nil {
		return err
	}
	return nil
}

func (h *MonitorHandler) Handle(ctx context.Context, msg interface{}) error {
	raw, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}

	m, err := ParseMessage(raw)
	if err != nil {
		h.logger.Warn("parse monitor", zap.Error(err))
		return nil
	}

	if m.DeviceUID == "" {
		return nil
	}
	deviceKey := h.metaCache.ResolveDeviceID(ctx, m.CardID, m.DeviceID, m.DeviceUID)
	if deviceKey == "" {
		return nil
	}

	nowMs := time.Now().UnixMilli()
	age := nowMs - m.Timestamp
	if age > MonitorFieldTTL {
		return nil
	}

	fields := redis.FirstDataValue(m.DataValue)
	if fields == nil {
		return nil
	}
	trackID := resolveTrackID(fields)
	h.buffer.Write(m.CardID, deviceKey, strconv.Itoa(trackID), fields, m.Timestamp)
	return nil
}

// resolveTrackID 从 payload 取 track_id。缺省、类型无法解析或与 88 一样走无效桶，避免 string/缺字段误落 key "0" 与真人轨迹合并成双目标。
func resolveTrackID(fields map[string]interface{}) int {
	if fields == nil {
		return observation.TrackInvalid
	}
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
