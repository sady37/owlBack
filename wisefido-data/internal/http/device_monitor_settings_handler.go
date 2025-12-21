package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// DeviceMonitorSettingsHandler 设备监控配置 Handler
type DeviceMonitorSettingsHandler struct {
	deviceMonitorSettingsService service.DeviceMonitorSettingsService
	logger                       *zap.Logger
	base                         *StubHandler // 用于 tenantIDFromReq
}

// NewDeviceMonitorSettingsHandler 创建设备监控配置 Handler
func NewDeviceMonitorSettingsHandler(deviceMonitorSettingsService service.DeviceMonitorSettingsService, logger *zap.Logger) *DeviceMonitorSettingsHandler {
	return &DeviceMonitorSettingsHandler{
		deviceMonitorSettingsService: deviceMonitorSettingsService,
		logger:                       logger,
		base:                         &StubHandler{}, // 用于 tenantIDFromReq
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceMonitorSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 解析设备类型和设备ID
	var deviceType string
	var deviceID string

	if strings.HasPrefix(path, "/settings/api/v1/monitor/sleepace/") {
		deviceType = "sleepace"
		deviceID = strings.TrimPrefix(path, "/settings/api/v1/monitor/sleepace/")
	} else if strings.HasPrefix(path, "/settings/api/v1/monitor/radar/") {
		deviceType = "radar"
		deviceID = strings.TrimPrefix(path, "/settings/api/v1/monitor/radar/")
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 验证 deviceID 不为空且不包含 "/"
	if deviceID == "" || strings.Contains(deviceID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 根据 HTTP 方法分发
	switch r.Method {
	case http.MethodGet:
		h.GetDeviceMonitorSettings(w, r, deviceType, deviceID)
	case http.MethodPut:
		h.UpdateDeviceMonitorSettings(w, r, deviceType, deviceID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GetDeviceMonitorSettings 获取设备监控配置
func (h *DeviceMonitorSettingsHandler) GetDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType, deviceID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
	}

	resp, err := h.deviceMonitorSettingsService.GetDeviceMonitorSettings(ctx, req)
	if err != nil {
		h.logger.Error("GetDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(resp.Settings))
}

// UpdateDeviceMonitorSettings 更新设备监控配置
func (h *DeviceMonitorSettingsHandler) UpdateDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType, deviceID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]interface{}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		Settings:   payload,
	}

	resp, err := h.deviceMonitorSettingsService.UpdateDeviceMonitorSettings(ctx, req)
	if err != nil {
		h.logger.Error("UpdateDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"success": resp.Success,
	}))
}

