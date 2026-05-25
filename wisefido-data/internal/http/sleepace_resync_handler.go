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
// 端点（POST，无 body，幂等；path 占位 {device_addr} = devices.device_addr INET host 或 /128 CIDR）：
//
//	POST /internal/sleepace/device/{device_addr}/resync-timezone
//	POST /internal/sleepace/device/{device_addr}/resync-report-time
//	POST /internal/sleepace/device/{device_addr}/upgrade   (body: { version | filename })
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
	// path: /internal/sleepace/device/{device_addr}/{action}
	//   或 /sleepace/api/v1/sleepace/device/{device_addr}/{action}（FE 走 nginx /sleepace/api/）
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/internal/sleepace/device/")
	path = strings.TrimPrefix(path, "/sleepace/api/v1/sleepace/device/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeResyncJSON(w, http.StatusBadRequest, map[string]any{
			"status": "error",
			"error":  "path must be {prefix}/sleepace/device/{device_addr}/{resync-timezone|resync-report-time|upgrade}",
		})
		return
	}
	deviceAddr, action := parts[0], parts[1]

	tenantID, err := h.lookupTenantID(r, deviceAddr)
	if err != nil {
		writeResyncJSON(w, http.StatusNotFound, map[string]any{"status": "error", "error": err.Error()})
		return
	}

	switch action {
	case "resync-timezone":
		iana, secs, err := h.svc.ResyncDeviceTimezone(r.Context(), tenantID, deviceAddr)
		resp := map[string]any{
			"device_addr": deviceAddr,
			"tenant_id":   tenantID,
			"iana":        iana,
			"tz_seconds":  secs,
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
		hour, err := h.svc.ResyncDeviceReportTime(r.Context(), tenantID, deviceAddr)
		resp := map[string]any{
			"device_addr":        deviceAddr,
			"tenant_id":          tenantID,
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
	case "upgrade":
		// body 二选一：{"version":"6.89"} 直传；{"filename":"mcu_sleepace_..."} 由后端读 update.ini 查 deviceVerison。
		var body struct {
			Version  string `json:"version"`
			Filename string `json:"filename"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		version := body.Version
		if version == "" && body.Filename != "" {
			v, err := h.svc.ResolveSleepaceUpgradeVersion(body.Filename)
			if err != nil {
				writeResyncJSON(w, http.StatusBadRequest, map[string]any{
					"status": "error", "device_addr": deviceAddr, "tenant_id": tenantID,
					"filename": body.Filename, "error": "resolve version: " + err.Error(),
				})
				return
			}
			version = v
		}
		resp := map[string]any{
			"device_addr": deviceAddr,
			"tenant_id":   tenantID,
			"version":     version,
			"filename":    body.Filename,
		}
		if err := h.svc.TriggerSleepaceUpgrade(r.Context(), tenantID, deviceAddr, version); err != nil {
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
			"error":  "unknown action; expected resync-timezone | resync-report-time | upgrade",
		})
	}
}

// lookupTenantID 内部端点没有 session token，tenantID 从 device 行查出来。
func (h *SleepaceResyncHandler) lookupTenantID(r *http.Request, deviceAddr string) (string, error) {
	if h.db == nil {
		return "", errResyncDBUnavailable
	}
	var tenantID string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT tenant_id::text FROM devices WHERE device_addr = $1::INET`, deviceAddr).Scan(&tenantID)
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
