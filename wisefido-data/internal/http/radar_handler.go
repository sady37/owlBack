package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"wisefido-data/internal/service"
	
	"go.uber.org/zap"
)

// RadarHandler Radar 设备 API Handler
// 实现 v1.0 兼容的 API 端点
type RadarHandler struct {
	radarService *service.RadarService
	stubHandler  *StubHandler // 用于复用 tenantIDFromReq 逻辑
	logger       *zap.Logger
}

// NewRadarHandler 创建 Radar Handler
func NewRadarHandler(radarService *service.RadarService, stubHandler *StubHandler, logger *zap.Logger) *RadarHandler {
	return &RadarHandler{
		radarService: radarService,
		stubHandler:  stubHandler,
		logger:       logger,
	}
}

// GetRealtimeData 获取实时数据
// GET /radar-device/api/v1/radar-device/device/:id/realtime
// 参考 v1.0 API 格式
func (h *RadarHandler) GetRealtimeData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 device_id
	// 路径格式：/radar-device/api/v1/radar-device/device/{id}/realtime
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/realtime")
	if deviceID == "" {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	
	// 获取 tenant_id（用于后续实现）
	_, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	
	// TODO: 实现实时数据获取
	// 当前实现：返回空数据（后续从 Redis Streams 或 PostgreSQL 获取）
	// 参考 v1.0 格式：
	// {
	//   "positions": [...],
	//   "vital": {...}
	// }
	
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"positions": []interface{}{},
		"vital":     nil,
	}))
}

// GetOriginalProperties 获取设备原始属性
// GET /radar-device/api/v1/radar-device/device/:id/original-properties
// 参考 v1.0 API 格式
// 返回：JSON 字符串，包含雷达所有配置参数
func (h *RadarHandler) GetOriginalProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 device_id
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/original-properties")
	if deviceID == "" {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	
	// 获取 tenant_id
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	
	// 调用服务获取设备属性
	propertiesJSON, err := h.radarService.GetOriginalProperties(r.Context(), tenantID, deviceID)
	if err != nil {
		h.logger.Error("Failed to get device original properties",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	
	// v1.0 格式：直接返回 JSON 字符串
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(propertiesJSON))
}

// UpdateConfig 更新设备配置
// PUT /radar-device/api/v1/radar-device/device/:id/config
// 参考 v1.0 API 格式
func (h *RadarHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 device_id
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/config")
	if deviceID == "" {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	
	// 获取 tenant_id
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	
	// 解析请求体
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 调用服务更新配置
	if err := h.radarService.UpdateConfig(r.Context(), tenantID, deviceID, config); err != nil {
		h.logger.Error("Failed to update device config",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	
	// v1.0 格式：返回字符串数组（可能表示更新的字段列表）
	writeJSON(w, http.StatusOK, Ok([]string{}))
}

// extractRadarDeviceIDFromPath 从 URL 路径中提取 device_id（Radar 专用）
func extractRadarDeviceIDFromPath(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	deviceID := path[len(prefix) : len(path)-len(suffix)]
	if deviceID == "" || strings.Contains(deviceID, "/") {
		return ""
	}
	return deviceID
}

// tenantIDFromReq 从请求中获取 tenant_id（复用 stub_handler_base.go 的逻辑）
func (h *RadarHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.stubHandler != nil {
		return h.stubHandler.tenantIDFromReq(w, r)
	}
	// Fallback: 直接从 header 获取
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

