package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"wisefido-data/internal/service"
)

// SleepaceResyncHandler 挂在 /internal/sleepace/device/* 下，供外部 DST 脚本使用。
//
// 两个端点（无 body，幂等）：
//
//	POST /internal/sleepace/device/{device_id}/resync-timezone
//	POST /internal/sleepace/device/{device_id}/resync-report-time
//
// 外部不需要传 timezone 或 hour——service 内部按 effective 规则（device > tenant > default）
// 查当前值后下发给厂家。典型用法：DST 切换当天 cron 脚本遍历所有 Sleepad 设备各调一次。
//
// 路径前缀 `/internal/*` 走 auth skip（DefaultSkippedPaths），不需要 token。
type SleepaceResyncHandler struct {
	svc    service.DeviceMonitorSettingsService
	db     *sql.DB
	logger *zap.Logger
}

func NewSleepaceResyncHandler(svc service.DeviceMonitorSettingsService, db *sql.DB, logger *zap.Logger) *SleepaceResyncHandler {
	return &SleepaceResyncHandler{svc: svc, db: db, logger: logger}
}

// Dispatch 基于 path 后缀分发。
func (h *SleepaceResyncHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// path: /internal/sleepace/device/{device_id}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/internal/sleepace/device/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeResyncJSON(w, http.StatusBadRequest, map[string]any{
			"status": "error",
			"error":  "path must be /internal/sleepace/device/{device_id}/{resync-timezone|resync-report-time}",
		})
		return
	}
	deviceID, action := parts[0], parts[1]

	tenantID, err := h.lookupTenantID(r, deviceID)
	if err != nil {
		writeResyncJSON(w, http.StatusNotFound, map[string]any{"status": "error", "error": err.Error()})
		return
	}

	switch action {
	case "resync-timezone":
		iana, secs, err := h.svc.ResyncDeviceTimezone(r.Context(), tenantID, deviceID)
		resp := map[string]any{
			"device_id":  deviceID,
			"tenant_id":  tenantID,
			"iana":       iana,
			"tz_seconds": secs,
		}
		if err != nil {
			resp["status"] = "error"
			resp["error"] = err.Error()
			writeResyncJSON(w, http.StatusInternalServerError, resp)
			return
		}
		resp["status"] = "ok"
		writeResyncJSON(w, http.StatusOK, resp)
	case "resync-report-time":
		hour, err := h.svc.ResyncDeviceReportTime(r.Context(), tenantID, deviceID)
		resp := map[string]any{
			"device_id":         deviceID,
			"tenant_id":         tenantID,
			"report_upload_time": hour,
		}
		if err != nil {
			resp["status"] = "error"
			resp["error"] = err.Error()
			writeResyncJSON(w, http.StatusInternalServerError, resp)
			return
		}
		resp["status"] = "ok"
		writeResyncJSON(w, http.StatusOK, resp)
	default:
		writeResyncJSON(w, http.StatusBadRequest, map[string]any{
			"status": "error",
			"error":  "unknown action; expected resync-timezone or resync-report-time",
		})
	}
}

// lookupTenantID 内部端点没有 session token，tenantID 从 device 行查出来。
func (h *SleepaceResyncHandler) lookupTenantID(r *http.Request, deviceID string) (string, error) {
	if h.db == nil {
		return "", errResyncDBUnavailable
	}
	var tenantID string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT tenant_id::text FROM devices WHERE device_id = $1::uuid`, deviceID).Scan(&tenantID)
	if err != nil {
		return "", err
	}
	return tenantID, nil
}

func writeResyncJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// 简单错误（不引入新 sentinel 到其他文件）
type resyncErr string

func (e resyncErr) Error() string { return string(e) }

const errResyncDBUnavailable = resyncErr("database not available")
