package consumer

import (
	"context"
	"strconv"
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
	// 前后台统一用 device_id (UUID) 为 key；IotPreparedHandler 未绑卡时会回填 device_id，此处仅无 device_id 时用 device_uid 兜底 buffer 写入
	deviceKey := m.DeviceID
	if deviceKey == "" {
		deviceKey = m.DeviceUID
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
	// track_id=88 视为无效轨迹，用 key "88" 写入以便设备能标为在线，但 Flush 发布后前端只渲染 track 0-8，不会多出一个在雷达位置的人
	if trackID == observation.TrackInvalid {
		trackID = observation.TrackInvalid
	}
	h.buffer.Write(m.CardID, deviceKey, strconv.Itoa(trackID), fields, m.Timestamp)
	return nil
}

// resolveTrackID 从 payload 取 track_id，无或非法则返回 DefaultTrackID 对应数值 0。
func resolveTrackID(fields map[string]interface{}) int {
	if fields == nil {
		return 0
	}
	switch v := fields["track_id"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}
