package service

import (
	"context"
	"fmt"
	"time"

	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// 与 owlFront WaveMonitor / QueryPanel 约定一致，在后端强制防止大时间窗扫库。
const (
	PlaybackLookbackWindow = 24 * time.Hour
	PlaybackTrackMaxDur    = 15 * time.Minute
	PlaybackVitalMaxDur    = 30 * time.Minute
	playbackClockSkew      = 2 * time.Minute // end 略晚于服务端时钟时允许

	playbackRawMaxRows = 3000
	playbackPageSize   = 500
)

type PlaybackKind string

const (
	PlaybackKindTrack PlaybackKind = "track"
	PlaybackKindVital PlaybackKind = "vital"
)

// ValidatePlaybackWindow 校验历史查询时间窗（track / vital 共用入口）。
// - 仅允许查询「当前时刻」之前的数据（可容忍小幅时钟偏差）。
// - 整段必须落在最近 PlaybackLookbackWindow 内，防止任意历史大扫库。
// - 时长：track ≤15min，vital ≤30min（与 WaveMonitor 输入上限一致）。
func ValidatePlaybackWindow(start, end time.Time, kind PlaybackKind) error {
	if end.Before(start) {
		return fmt.Errorf("invalid range: end before start")
	}
	now := time.Now()
	if start.After(now.Add(playbackClockSkew)) {
		return fmt.Errorf("start time must not be in the future")
	}
	if end.After(now.Add(playbackClockSkew)) {
		return fmt.Errorf("end time must not be in the future")
	}
	oldest := now.Add(-PlaybackLookbackWindow)
	if start.Before(oldest) {
		return fmt.Errorf("query range too old: only last %s is allowed", PlaybackLookbackWindow)
	}
	if end.Sub(start) < time.Second {
		return fmt.Errorf("range too short")
	}
	maxDur := PlaybackTrackMaxDur
	if kind == PlaybackKindVital {
		maxDur = PlaybackVitalMaxDur
	}
	if end.Sub(start) > maxDur {
		return fmt.Errorf("range too long: max %s for %s playback", maxDur, kind)
	}
	return nil
}

// TrackPlaybackService 轨迹历史回放：iot_timeseries 原始 monitor 行 → 前端自行解析
type TrackPlaybackService struct {
	devices repository.DevicesRepository
	iot     repository.IoTTimeSeriesRepository
	log     *zap.Logger
}

func NewTrackPlaybackService(devices repository.DevicesRepository, iot repository.IoTTimeSeriesRepository, log *zap.Logger) *TrackPlaybackService {
	if log == nil {
		log = zap.NewNop()
	}
	return &TrackPlaybackService{devices: devices, iot: iot, log: log}
}

// RadarTrackPlayback 返回 result：{ layout, data }，其中 data.rows 为 IoT 原始行（data_value 与入库一致），data.pages 为每页 500 条。
func (s *TrackPlaybackService) RadarTrackPlayback(ctx context.Context, tenantID, deviceID string, startMs, endMs int64) (map[string]interface{}, error) {
	start := time.UnixMilli(startMs).UTC()
	end := time.UnixMilli(endMs).UTC()
	s.log.Info("RadarTrackPlayback begin",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.Int64("start_ms", startMs),
		zap.Int64("end_ms", endMs),
		zap.Time("start_utc", start),
		zap.Time("end_utc", end),
		zap.String("kind", string(PlaybackKindTrack)),
	)
	if err := ValidatePlaybackWindow(start, end, PlaybackKindTrack); err != nil {
		s.log.Warn("RadarTrackPlayback window rejected by ValidatePlaybackWindow",
			zap.Error(err),
			zap.Time("start_utc", start),
			zap.Time("end_utc", end),
			zap.String("note", "Alarm/Event replay failed: lookback window exceeds 24 hours."),
		)
		return nil, err
	}

	dev, err := s.devices.GetDevice(ctx, tenantID, deviceID)
	if err != nil || dev == nil {
		s.log.Warn("RadarTrackPlayback device lookup failed", zap.String("tenant_id", tenantID), zap.String("device_id", deviceID), zap.Error(err))
		return nil, fmt.Errorf("device not found or access denied")
	}
	if dev.DeviceUID == "" {
		s.log.Warn("RadarTrackPlayback device_uid empty", zap.String("device_id", deviceID))
		return nil, fmt.Errorf("device_uid missing")
	}

	s.log.Info("RadarTrackPlayback querying iot_timeseries raw",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", dev.DeviceUID),
		zap.Time("filter_start", start),
		zap.Time("filter_end", end),
	)

	rows, err := s.iot.GetPlaybackRawMonitorRowsForDevice(ctx, tenantID, dev.DeviceUID, start, end, playbackRawMaxRows)
	if err != nil {
		s.log.Warn("RadarTrackPlayback iot_timeseries raw error",
			zap.Error(err),
			zap.String("tenant_id", tenantID),
			zap.String("device_uid", dev.DeviceUID),
			zap.Time("start", start),
			zap.Time("end", end),
		)
		return nil, fmt.Errorf("query iot_timeseries failed")
	}
	s.log.Info("RadarTrackPlayback iot_timeseries raw ok",
		zap.Int("row_count", len(rows)),
		zap.String("device_uid", dev.DeviceUID),
	)

	pages := chunkPlaybackPages(rows, playbackPageSize)
	out := map[string]interface{}{
		"layout": nil,
		"data": map[string]interface{}{
			"rows":     rows,
			"pages":    pages,
			"pageSize": playbackPageSize,
			"total":    len(rows),
		},
	}
	return out, nil
}

func chunkPlaybackPages(rows []map[string]interface{}, pageSize int) [][]map[string]interface{} {
	if pageSize <= 0 {
		pageSize = playbackPageSize
	}
	if len(rows) == 0 {
		return nil
	}
	var pages [][]map[string]interface{}
	for i := 0; i < len(rows); i += pageSize {
		j := i + pageSize
		if j > len(rows) {
			j = len(rows)
		}
		pages = append(pages, rows[i:j])
	}
	return pages
}
