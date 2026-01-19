package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"owl-common/alarm"
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

	// 检查是否是获取默认设置的请求
	if strings.HasPrefix(path, "/settings/api/v1/monitor/default/sleepace") {
		if r.Method == http.MethodGet {
			h.GetDefaultDeviceMonitorSettings(w, r, "sleepace")
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if strings.HasPrefix(path, "/settings/api/v1/monitor/default/radar") {
		if r.Method == http.MethodGet {
			h.GetDefaultDeviceMonitorSettings(w, r, "radar")
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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

	// 验证 deviceID 不为空且不包含 "/"，且不能是 "undefined" 或 "null"
	if deviceID == "" || strings.Contains(deviceID, "/") || deviceID == "undefined" || deviceID == "null" {
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

	writeJSON(w, http.StatusOK, Ok(resp.AlarmItems))
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

	// 从 payload 解析 alarm_items 数组
	alarmItemsRaw, ok := payload["alarm_items"]
	if !ok {
		writeJSON(w, http.StatusOK, Fail("alarm_items is required"))
		return
	}

	// 转换为 JSON 再解析为 []alarm.AlarmItem
	alarmItemsJSON, err := json.Marshal(alarmItemsRaw)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to marshal alarm_items: "+err.Error()))
		return
	}

	var alarmItems []alarm.AlarmItem
	if err := json.Unmarshal(alarmItemsJSON, &alarmItems); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to unmarshal alarm_items: "+err.Error()))
		return
	}

	// 获取 UserID
	userID := r.Header.Get("X-User-Id")

	req := service.UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		UserID:     userID,
		AlarmItems: alarmItems,
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

// GetDefaultDeviceMonitorSettings 获取默认设备监控配置
// 阈值：硬编码（与 System 租户模板设备的值相同）
// Alarm Level：优先从当前租户的 alarm_cloud 读取，如果没有则使用硬编码值
func (h *DeviceMonitorSettingsHandler) GetDefaultDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType string) {
	ctx := r.Context()

	// 从请求中获取 tenantID
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	resp, err := h.deviceMonitorSettingsService.GetDefaultDeviceMonitorSettings(ctx, tenantID, deviceType)
	if err != nil {
		h.logger.Error("GetDefaultDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(resp.AlarmItems))
}
