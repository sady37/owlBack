package httpapi

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"owl-common/alarm"

	"go.uber.org/zap"
)

// Body 可带 replay_kind、request_source 供审计；user_id / tenant_id 仅信任请求头（与 axios 拦截器一致，勿在 body 传用户身份）。
// X-Tenant-Id：可为租户 UUID，或 tenant_name（如 demo），由 ResolveTenantIDFromHeader 解析为 tenant_id。

// PlaybackHandler POST /api/radar/playback、/api/vital/playback（鉴权与租户与现有 API 一致）
type PlaybackHandler struct {
	track   *service.TrackPlaybackService
	tenants repository.TenantsRepository
	db      *sql.DB // 用于 Phase 3 device scope 校验
	log     *zap.Logger
}

func NewPlaybackHandler(track *service.TrackPlaybackService, tenants repository.TenantsRepository, db *sql.DB, log *zap.Logger) *PlaybackHandler {
	if log == nil {
		log = zap.NewNop()
	}
	return &PlaybackHandler{track: track, tenants: tenants, db: db, log: log}
}

type radarPlaybackBody struct {
	DeviceAddr      string `json:"deviceId"`
	DataType      string `json:"dataType"`
	StartTime     int64  `json:"startTime"`
	EndTime       int64  `json:"endTime"`
	ReplayKind    string `json:"replay_kind,omitempty"`
	RequestSource string `json:"request_source,omitempty"`
}

func replayKindFromAlarmEventType(et string) string {
	switch strings.TrimSpace(et) {
	case alarm.Fall, alarm.SittingOnGround:
		return "track"
	case alarm.HeartRateAlert, alarm.RespRateAlert, alarm.ApneaHypopnea, alarm.WeakBiometricSignal:
		return "vital"
	default:
		return "unknown"
	}
}

func normRequestSource(s string) string {
	s = strings.TrimSpace(s)
	if s != "" {
		return s
	}
	return "unknown"
}

func normReplayKindBodyOrFallback(body, fallback string) string {
	k := strings.TrimSpace(strings.ToLower(body))
	if k != "" {
		return k
	}
	return fallback
}

func (h *PlaybackHandler) logReplayAudit(
	endpoint string,
	r *http.Request,
	replayKind, requestSource string,
	reqTime time.Time,
	success bool,
	err error,
	extra ...zap.Field,
) {
	fields := []zap.Field{
		zap.String("endpoint", endpoint),
		zap.String("route", r.URL.Path),
		zap.String("replay_kind", replayKind),
		zap.String("request_source", requestSource),
		zap.String("user_id", strings.TrimSpace(r.Header.Get("X-User-Id"))),
		zap.String("tenant_id", strings.TrimSpace(r.Header.Get("X-Tenant-Id"))),
		zap.Time("request_time", reqTime),
		zap.Bool("success", success),
	}
	fields = append(fields, extra...)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	h.log.Info("replay_audit", fields...)
}

// PostRadarPlayback POST /api/radar/playback
func (h *PlaybackHandler) PostRadarPlayback(w http.ResponseWriter, r *http.Request) {
	reqTime := time.Now().UTC()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tenantRaw := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
	if tenantRaw == "" || tenantRaw == "null" {
		writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
		return
	}
	tenantID, err := ResolveTenantIDFromHeader(r.Context(), h.tenants, tenantRaw)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.logReplayAudit("PostRadarPlayback", r, "track", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	var req radarPlaybackBody
	if err := json.Unmarshal(body, &req); err != nil {
		h.logReplayAudit("PostRadarPlayback", r, "track", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	rk := normReplayKindBodyOrFallback(req.ReplayKind, "track")
	src := normRequestSource(req.RequestSource)
	if strings.TrimSpace(req.DeviceAddr) == "" {
		h.logReplayAudit("PostRadarPlayback", r, rk, src, reqTime, false, nil)
		writeJSON(w, http.StatusOK, Fail("deviceId is required"))
		return
	}
	if req.DataType != "" && req.DataType != "track" {
		h.logReplayAudit("PostRadarPlayback", r, rk, src, reqTime, false, nil)
		writeJSON(w, http.StatusOK, Fail("unsupported dataType"))
		return
	}

	did := strings.TrimSpace(req.DeviceAddr)
	// Phase 3: device 必须在 user scope 内
	if h.db != nil {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.db, r.Context(), uid, role, did); err != nil {
			h.logReplayAudit("PostRadarPlayback", r, rk, src, reqTime, false, err)
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	startRFC := time.UnixMilli(req.StartTime).UTC().Format(time.RFC3339Nano)
	endRFC := time.UnixMilli(req.EndTime).UTC().Format(time.RFC3339Nano)
	h.log.Info("PostRadarPlayback received",
		zap.String("tenant_header", tenantRaw),
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", did),
		zap.Int64("start_time_ms", req.StartTime),
		zap.Int64("end_time_ms", req.EndTime),
		zap.String("start_utc", startRFC),
		zap.String("end_utc", endRFC),
		zap.String("replay_kind", rk),
		zap.String("request_source", src),
		zap.String("user_id", strings.TrimSpace(r.Header.Get("X-User-Id"))),
	)

	userRole := strings.TrimSpace(r.Header.Get("X-User-Role"))
	result, err := h.track.RadarTrackPlayback(r.Context(), tenantID, did, req.StartTime, req.EndTime, userRole)
	if err != nil {
		h.logReplayAudit("PostRadarPlayback", r, rk, src, reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	h.logReplayAudit("PostRadarPlayback", r, rk, src, reqTime, true, nil)
	writeJSON(w, http.StatusOK, Ok(result))
}

// PostVitalPlayback POST /api/vital/playback — 与 track 相同时间窗校验；业务查询走 monitor_stream
func (h *PlaybackHandler) PostVitalPlayback(w http.ResponseWriter, r *http.Request) {
	reqTime := time.Now().UTC()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tenantRaw := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
	if tenantRaw == "" || tenantRaw == "null" {
		writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
		return
	}
	tenantID, err := ResolveTenantIDFromHeader(r.Context(), h.tenants, tenantRaw)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.logReplayAudit("PostVitalPlayback", r, "vital", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	var req radarPlaybackBody
	if err := json.Unmarshal(body, &req); err != nil {
		h.logReplayAudit("PostVitalPlayback", r, "vital", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	rk := normReplayKindBodyOrFallback(req.ReplayKind, "vital")
	src := normRequestSource(req.RequestSource)
	if req.EndTime < req.StartTime {
		h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, false, nil)
		writeJSON(w, http.StatusOK, Fail("invalid range"))
		return
	}
	startT := time.UnixMilli(req.StartTime).UTC()
	endT := time.UnixMilli(req.EndTime).UTC()
	h.log.Info("PostVitalPlayback received",
		zap.String("tenant_header", tenantRaw),
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", strings.TrimSpace(req.DeviceAddr)),
		zap.Int64("start_time_ms", req.StartTime),
		zap.Int64("end_time_ms", req.EndTime),
		zap.Time("start_utc", startT),
		zap.Time("end_utc", endT),
		zap.String("replay_kind", rk),
		zap.String("request_source", src),
		zap.String("user_id", strings.TrimSpace(r.Header.Get("X-User-Id"))),
	)
	vitalRole := strings.TrimSpace(r.Header.Get("X-User-Role"))
	if err := service.ValidatePlaybackWindow(startT, endT, service.PlaybackKindVital, service.PlaybackLookbackForRole(vitalRole)); err != nil {
		h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if strings.TrimSpace(req.DeviceAddr) == "" {
		h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, false, nil)
		writeJSON(w, http.StatusOK, Fail("deviceId is required"))
		return
	}
	// Phase 3: device 必须在 user scope 内
	if h.db != nil {
		uid := r.Header.Get("X-User-Id")
		if err := VerifyDeviceInScope(h.db, r.Context(), uid, vitalRole, strings.TrimSpace(req.DeviceAddr)); err != nil {
			h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, false, err)
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	result, err := h.track.RadarVitalPlayback(r.Context(), tenantID, strings.TrimSpace(req.DeviceAddr), req.StartTime, req.EndTime, vitalRole)
	if err != nil {
		h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	h.logReplayAudit("PostVitalPlayback", r, rk, src, reqTime, true, nil)
	writeJSON(w, http.StatusOK, Ok(result))
}

// alarmReplayContextBody WaveMonitor 从 alarm_records 点 RePlay 时上报，供后端查询/审计（与回放数据面解耦）
type alarmReplayContextBody struct {
	DeviceAddr      string `json:"device_addr"`
	DeviceName    string `json:"device_name"`
	AlarmTimeSec  int64  `json:"alarm_time_sec"`
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	ReplayKind    string `json:"replay_kind,omitempty"`
	RequestSource string `json:"request_source,omitempty"`
}

// PostAlarmReplayContext POST /api/radar/alarm-replay-context
func (h *PlaybackHandler) PostAlarmReplayContext(w http.ResponseWriter, r *http.Request) {
	reqTime := time.Now().UTC()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.logReplayAudit("PostAlarmReplayContext", r, "unknown", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	var req alarmReplayContextBody
	if err := json.Unmarshal(body, &req); err != nil {
		h.logReplayAudit("PostAlarmReplayContext", r, "unknown", "unknown", reqTime, false, err)
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	rk := strings.TrimSpace(strings.ToLower(req.ReplayKind))
	if rk == "" {
		rk = replayKindFromAlarmEventType(req.EventType)
	}
	src := normRequestSource(req.RequestSource)
	// Phase 3: device 必须在 user scope 内
	if h.db != nil && strings.TrimSpace(req.DeviceAddr) != "" {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.db, r.Context(), uid, role, strings.TrimSpace(req.DeviceAddr)); err != nil {
			h.logReplayAudit("PostAlarmReplayContext", r, rk, src, reqTime, false, err)
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	h.logReplayAudit("PostAlarmReplayContext", r, rk, src, reqTime, true, nil,
		zap.String("device_addr", req.DeviceAddr),
		zap.String("event_id", req.EventID),
		zap.String("event_type", req.EventType),
		zap.Int64("alarm_time_sec", req.AlarmTimeSec))
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"ok": true}))
}
