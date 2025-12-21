package httpapi

import (
	"encoding/json"
	"net/http"

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

	// 3. 转换为旧 Handler 的响应格式（确保完全一致）
	// 旧 Handler 返回 map[string]any，需要转换为相同格式
	result := make(map[string]any)
	result["tenant_id"] = resp.TenantID

	// 处理 device_alarms（JSONB 字段，与旧 Handler 逻辑一致）
	var deviceAlarms map[string]any
	if len(resp.DeviceAlarms) > 0 {
		if err := json.Unmarshal(resp.DeviceAlarms, &deviceAlarms); err == nil {
			result["device_alarms"] = deviceAlarms
		} else {
			result["device_alarms"] = map[string]any{}
		}
	} else {
		result["device_alarms"] = map[string]any{}
	}

	// 处理可选字段（与旧 Handler 逻辑一致）
	if resp.OfflineAlarm != nil {
		result["OfflineAlarm"] = *resp.OfflineAlarm
	}
	if resp.LowBattery != nil {
		result["LowBattery"] = *resp.LowBattery
	}
	if resp.DeviceFailure != nil {
		result["DeviceFailure"] = *resp.DeviceFailure
	}
	if len(resp.Conditions) > 0 {
		var conditions any
		if err := json.Unmarshal(resp.Conditions, &conditions); err == nil {
			result["conditions"] = conditions
		}
	}
	if len(resp.NotificationRules) > 0 {
		var notificationRules any
		if err := json.Unmarshal(resp.NotificationRules, &notificationRules); err == nil {
			result["notification_rules"] = notificationRules
		}
	}

	// 4. 返回响应
	writeJSON(w, http.StatusOK, Ok(result))
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

	// 2. 解析字段（与旧 Handler 逻辑一致）
	req := service.UpdateAlarmCloudConfigRequest{
		TenantID: tenantID,
		UserID:   userID,
		UserRole: userRole,
	}

	// 处理 OfflineAlarm, LowBattery, DeviceFailure（与旧 Handler 逻辑一致）
	// 旧 Handler: 如果为空字符串，不更新（使用 sql.NullString）
	// 新 Handler: 如果为空字符串或不存在，不更新（使用指针 nil）
	if val, ok := payload["OfflineAlarm"].(string); ok && val != "" {
		req.OfflineAlarm = &val
	}
	if val, ok := payload["LowBattery"].(string); ok && val != "" {
		req.LowBattery = &val
	}
	if val, ok := payload["DeviceFailure"].(string); ok && val != "" {
		req.DeviceFailure = &val
	}

	// 处理 device_alarms（JSONB 字段）
	if val, ok := payload["device_alarms"].(map[string]any); ok && val != nil {
		deviceAlarmsJSON, err := json.Marshal(val)
		if err == nil {
			req.DeviceAlarms = deviceAlarmsJSON
		}
	}

	// 处理 conditions, notification_rules（JSONB 字段）
	if val, ok := payload["conditions"]; ok && val != nil {
		conditionsJSON, err := json.Marshal(val)
		if err == nil {
			req.Conditions = conditionsJSON
		}
	}
	if val, ok := payload["notification_rules"]; ok && val != nil {
		notificationRulesJSON, err := json.Marshal(val)
		if err == nil {
			req.NotificationRules = notificationRulesJSON
		}
	}

	// 处理 metadata（JSONB 字段）
	if val, ok := payload["metadata"]; ok && val != nil {
		metadataJSON, err := json.Marshal(val)
		if err == nil {
			req.Metadata = metadataJSON
		}
	}

	// 3. 调用 Service
	resp, err := h.alarmCloudService.UpdateAlarmCloudConfig(ctx, req)
	if err != nil {
		h.logger.Error("UpdateAlarmCloudConfig failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 转换为旧 Handler 的响应格式（确保完全一致）
	result := make(map[string]any)
	result["tenant_id"] = resp.TenantID

	// 处理 device_alarms（JSONB 字段，与旧 Handler 逻辑一致）
	var deviceAlarms map[string]any
	if len(resp.DeviceAlarms) > 0 {
		if err := json.Unmarshal(resp.DeviceAlarms, &deviceAlarms); err == nil {
			result["device_alarms"] = deviceAlarms
		} else {
			result["device_alarms"] = map[string]any{}
		}
	} else {
		result["device_alarms"] = map[string]any{}
	}

	// 处理可选字段（与旧 Handler 逻辑一致）
	// 旧 Handler: 使用 sql.NullString，Valid 为 true 时才包含
	// 新 Handler: 使用指针，不为 nil 且不为空字符串时才包含
	if resp.OfflineAlarm != nil && *resp.OfflineAlarm != "" {
		result["OfflineAlarm"] = *resp.OfflineAlarm
	}
	if resp.LowBattery != nil && *resp.LowBattery != "" {
		result["LowBattery"] = *resp.LowBattery
	}
	if resp.DeviceFailure != nil && *resp.DeviceFailure != "" {
		result["DeviceFailure"] = *resp.DeviceFailure
	}
	if len(resp.Conditions) > 0 {
		var conditions any
		if err := json.Unmarshal(resp.Conditions, &conditions); err == nil {
			result["conditions"] = conditions
		}
	}
	if len(resp.NotificationRules) > 0 {
		var notificationRules any
		if err := json.Unmarshal(resp.NotificationRules, &notificationRules); err == nil {
			result["notification_rules"] = notificationRules
		}
	}

	// 5. 返回响应
	writeJSON(w, http.StatusOK, Ok(result))
}

