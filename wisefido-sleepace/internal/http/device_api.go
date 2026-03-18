package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"wisefido-sleepace/internal/service"

	"go.uber.org/zap"
)

type DeviceAPI struct {
	sleepaceAPI *service.SleepaceAPI
	cardMapping *service.CardMappingService
	logger      *zap.Logger
}

func NewDeviceAPI(api *service.SleepaceAPI, mapping *service.CardMappingService, logger *zap.Logger) *DeviceAPI {
	return &DeviceAPI{sleepaceAPI: api, cardMapping: mapping, logger: logger}
}

// HandleInitialize POST /api/v1/sleepace/device/initialize
// Body: device_code (wisefido.device_code= Sleepace.device_id), user_id (wisefido.device_uid = Sleepace.userid). timezone 可选。
func (d *DeviceAPI) HandleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		DeviceCode string `json:"device_code"`
		UserID     string `json:"user_id"`
		Timezone   int    `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid json"})
		return
	}
	if req.DeviceCode == "" || req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "device_code and user_id required"})
		return
	}

	d.logger.Info("initialize request", zap.String("device_code", req.DeviceCode), zap.String("user_id", req.UserID), zap.Int("timezone", req.Timezone))
	platformID, err := d.sleepaceAPI.InitializeDeviceByCode(req.DeviceCode, req.UserID, req.Timezone)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "" {
			errMsg = "unknown"
		}
		d.logger.Error("initialize device failed", zap.String("device_code", req.DeviceCode), zap.String("user_id", req.UserID), zap.String("error", errMsg))
		status := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(errMsg), "device not found") {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{"success": false, "error": errMsg})
		return
	}

	if platformID != "" && req.DeviceCode != "" && d.cardMapping != nil {
		d.cardMapping.PutMapping(platformID, req.DeviceCode)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"device_code": platformID,
	})
}

// HandleUnbind DELETE /api/v1/sleepace/device/{code}
func (d *DeviceAPI) HandleUnbind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	const prefix = "/api/v1/sleepace/device/"
	code := strings.TrimPrefix(r.URL.Path, prefix)
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "device_code required"})
		return
	}

	if err := d.sleepaceAPI.UnbindDevice(code); err != nil {
		d.logger.Error("unbind device failed", zap.String("device_code", code), zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "device_code": code})
}
