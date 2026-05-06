package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// 与 owlFront WaveMonitor / QueryPanel 约定一致，在后端强制防止大时间窗扫库。
// 最早可查起点按角色分层（与 X-User-Role 对齐）：管理层事故调查/合规可 30 天；一线护理 48 小时。
const (
	PlaybackLookbackAdminTier  = 30 * 24 * time.Hour // 产品对外口径：Admin、Manager；不含 IT
	PlaybackLookbackClinical   = 48 * time.Hour      // Nurse, Caregiver
	PlaybackLookbackDefault    = 48 * time.Hour      // 未知角色、空角色、Resident 等：与临床档一致，严于 30 天
	PlaybackLookbackWindow     = PlaybackLookbackDefault
	PlaybackTrackMaxDur        = 15 * time.Minute
	PlaybackVitalMaxDur        = 30 * time.Minute
	playbackClockSkew          = 2 * time.Minute // end 略晚于服务端时钟时允许

	playbackRawMaxRows = 3000
	playbackPageSize   = 500
)

// PlaybackLookbackForRole 返回该角色允许的回溯时长（仅用于 ValidatePlaybackWindow 的 oldest 边界）。
func PlaybackLookbackForRole(role string) time.Duration {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case "admin", "manager", "systemadmin", "systemoperator":
		return PlaybackLookbackAdminTier
	case "nurse", "caregiver":
		return PlaybackLookbackClinical
	default:
		return PlaybackLookbackDefault
	}
}

func formatPlaybackLookback(d time.Duration) string {
	if d == PlaybackLookbackAdminTier {
		return "30 days"
	}
	if d == PlaybackLookbackClinical { // same numeric value as PlaybackLookbackDefault
		return "48 hours"
	}
	return d.String()
}

type PlaybackKind string

const (
	PlaybackKindTrack PlaybackKind = "track"
	PlaybackKindVital PlaybackKind = "vital"
)

// ValidatePlaybackWindow 校验历史查询时间窗（track / vital 共用入口）。
// - 仅允许查询「当前时刻」之前的数据（可容忍小幅时钟偏差）。
// - 整段 start 不得早于 now-lookback（lookback 由角色决定），防止任意历史大扫库。
// - 时长：track ≤15min，vital ≤30min（与 WaveMonitor 输入上限一致）。
func ValidatePlaybackWindow(start, end time.Time, kind PlaybackKind, lookback time.Duration) error {
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
	oldest := now.Add(-lookback)
	if start.Before(oldest) {
		return fmt.Errorf("query range too old: only last %s is allowed for your role", formatPlaybackLookback(lookback))
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
// userRole 来自请求头 X-User-Role，用于 PlaybackLookbackForRole。
func (s *TrackPlaybackService) RadarTrackPlayback(ctx context.Context, tenantID, deviceID string, startMs, endMs int64, userRole string) (map[string]interface{}, error) {
	start := time.UnixMilli(startMs).UTC()
	end := time.UnixMilli(endMs).UTC()
	lb := PlaybackLookbackForRole(userRole)
	s.log.Info("RadarTrackPlayback begin",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.Int64("start_ms", startMs),
		zap.Int64("end_ms", endMs),
		zap.Time("start_utc", start),
		zap.Time("end_utc", end),
		zap.String("kind", string(PlaybackKindTrack)),
		zap.String("user_role", strings.TrimSpace(userRole)),
		zap.Duration("lookback", lb),
	)
	if err := ValidatePlaybackWindow(start, end, PlaybackKindTrack, lb); err != nil {
		s.log.Warn("RadarTrackPlayback window rejected by ValidatePlaybackWindow",
			zap.Error(err),
			zap.Time("start_utc", start),
			zap.Time("end_utc", end),
			zap.Duration("lookback", lb),
			zap.String("note", "Replay start outside role-based lookback window."),
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

// VitalPlaybackPoint 一条 vital 数据点：来自 iot_timeseries data_value 中的 heart_rate / respiratory_rate / sleep_stage。
// 时间戳毫秒级；空值字段以 0 表示（前端会过滤 0 不画图）。
type VitalPlaybackPoint struct {
	Timestamp  int64 `json:"timestamp"`
	HR         int   `json:"hr"`
	RR         int   `json:"rr"`
	SleepStage int   `json:"sleepStage"`
}

// RadarVitalPlayback 从 iot_timeseries 抽出 HR/RR/sleep_stage 数据点（Sleepad 主源；Radar 暂未发 vital MQTT）。
// 与 RadarTrackPlayback 同样走 ValidatePlaybackWindow + 设备解析；返回 { points: [...], total: N }。
func (s *TrackPlaybackService) RadarVitalPlayback(ctx context.Context, tenantID, deviceID string, startMs, endMs int64, userRole string) (map[string]interface{}, error) {
	start := time.UnixMilli(startMs).UTC()
	end := time.UnixMilli(endMs).UTC()
	lb := PlaybackLookbackForRole(userRole)
	s.log.Info("RadarVitalPlayback begin",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.Int64("start_ms", startMs),
		zap.Int64("end_ms", endMs),
		zap.String("user_role", strings.TrimSpace(userRole)),
		zap.Duration("lookback", lb),
	)
	if err := ValidatePlaybackWindow(start, end, PlaybackKindVital, lb); err != nil {
		s.log.Warn("RadarVitalPlayback window rejected", zap.Error(err))
		return nil, err
	}

	dev, err := s.devices.GetDevice(ctx, tenantID, deviceID)
	if err != nil || dev == nil {
		s.log.Warn("RadarVitalPlayback device lookup failed", zap.String("tenant_id", tenantID), zap.String("device_id", deviceID), zap.Error(err))
		return nil, fmt.Errorf("device not found or access denied")
	}
	if dev.DeviceUID == "" {
		return nil, fmt.Errorf("device_uid missing")
	}

	rows, err := s.iot.GetPlaybackRawMonitorRowsForDevice(ctx, tenantID, dev.DeviceUID, start, end, playbackRawMaxRows)
	if err != nil {
		return nil, fmt.Errorf("query iot_timeseries failed")
	}

	points := extractVitalPoints(rows)
	s.log.Info("RadarVitalPlayback ok",
		zap.String("device_uid", dev.DeviceUID),
		zap.Int("rows_scanned", len(rows)),
		zap.Int("points_extracted", len(points)),
	)

	return map[string]interface{}{
		"data": map[string]interface{}{
			"points": points,
			"total":  len(points),
		},
	}, nil
}

// extractVitalPoints 解析 iot_timeseries 行的 data_value（数组或单 obj），保留含 HR/RR/sleep_stage 的项。
func extractVitalPoints(rows []map[string]interface{}) []VitalPlaybackPoint {
	out := make([]VitalPlaybackPoint, 0, len(rows))
	for _, row := range rows {
		ts, _ := row["timestamp"].(int64)
		dv := row["data_value"]
		items := normalizeToItemArray(dv)
		for _, item := range items {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			hr := intFromAny(m["heart_rate"])
			rr := intFromAny(m["respiratory_rate"])
			if rr == 0 {
				rr = intFromAny(m["breath_rate"]) // 兼容雷达 vital 子项命名
			}
			ss := intFromAny(m["sleep_stage"])
			if hr == 0 && rr == 0 && ss == 0 {
				continue // 没有 vital 字段，跳过
			}
			out = append(out, VitalPlaybackPoint{
				Timestamp:  ts,
				HR:         hr,
				RR:         rr,
				SleepStage: ss,
			})
		}
	}
	return out
}

// normalizeToItemArray 把 data_value（可能是 []interface{} 或 map[string]interface{}）拍平成数组。
func normalizeToItemArray(v interface{}) []interface{} {
	switch t := v.(type) {
	case []interface{}:
		return t
	case map[string]interface{}:
		return []interface{}{t}
	}
	return nil
}

func intFromAny(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case float32:
		return int(t)
	}
	return 0
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
