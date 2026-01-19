package httpapi

import (
	"encoding/json"
	"net/http"

	"owl-common/alarm"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// AlarmCloudHandler 告警配置管理 Handler
type AlarmCloudHandler struct {
	alarmCloudService service.AlarmCloudService
	logger            *zap.Logger
}

// NewAlarmCloudHandler 创建告警配置管理 Handler
func NewAlarmCloudHandler(alarmCloudService service.AlarmCloudService, logger *zap.Logger) *AlarmCloudHandler {
	return &AlarmCloudHandler{
		alarmCloudService: alarmCloudService,
		logger:            logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *AlarmCloudHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/alarm-cloud" && r.Method == http.MethodGet:
		h.GetAlarmCloudConfig(w, r)
	case r.URL.Path == "/admin/api/v1/alarm-cloud" && r.Method == http.MethodPut:
		h.UpdateAlarmCloudConfig(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// GetAlarmCloudConfig 查询告警配置
func (h *AlarmCloudHandler) GetAlarmCloudConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证（tenant_id 规范化，与旧 Handler 逻辑一致）
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" || tenantID == "null" {
		tenantID = r.Header.Get("X-Tenant-Id")
	}
	// Normalize: empty string or "null" means use SystemTenantID
	if tenantID == "" || tenantID == "null" {
		tenantID = SystemTenantID()
	}

	userID := r.Header.Get("X-User-Id")
	userRole := r.Header.Get("X-User-Role")

	// 2. 调用 Service
	req := service.GetAlarmCloudConfigRequest{
		TenantID: tenantID,
		UserID:   userID,
		UserRole: userRole,
	}

	resp, err := h.alarmCloudService.GetAlarmCloudConfig(ctx, req)
	if err != nil {
		h.logger.Error("GetAlarmCloudConfig failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 直接返回完整的 AlarmCloudConfig 对象（JSON 序列化）
	// 前端会直接使用这个完整的配置对象
	writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdateAlarmCloudConfig 更新告警配置
func (h *AlarmCloudHandler) UpdateAlarmCloudConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证（tenant_id 规范化，与旧 Handler 逻辑一致）
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// Get tenant_id from payload or header
	tenantID, _ := payload["tenant_id"].(string)
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-Id")
	}
	if tenantID == "" {
		tenantID = SystemTenantID()
	}

	userID := r.Header.Get("X-User-Id")
	userRole := r.Header.Get("X-User-Role")

	// 2. 解析完整的 AlarmCloudConfig 对象
	// 将 payload map 转换为 JSON，然后解析为 AlarmCloudConfig
	configJSON, err := json.Marshal(payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid config format: "+err.Error()))
		return
	}

	var config alarm.AlarmCloudConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid config format: "+err.Error()))
		return
	}

	// 3. 构建请求
	req := service.UpdateAlarmCloudConfigRequest{
		TenantID: tenantID,
		UserID:   userID,
		UserRole: userRole,
		Config:   &config,
	}

	// 4. 调用 Service
	resp, err := h.alarmCloudService.UpdateAlarmCloudConfig(ctx, req)
	if err != nil {
		h.logger.Error("UpdateAlarmCloudConfig failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 5. 直接返回完整的 AlarmCloudConfig 对象（JSON 序列化）
	writeJSON(w, http.StatusOK, Ok(resp))
}
