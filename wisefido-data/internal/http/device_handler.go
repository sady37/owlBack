package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// SystemTenantID 从 stub_handler_base.go 导入（同一包内可直接使用）
// 如果编译错误，需要确保 stub_handler_base.go 在同一包中

// DeviceHandler 设备管理 Handler
type DeviceHandler struct {
	deviceService service.DeviceService
	db            *sql.DB
	logger        *zap.Logger
}

// NewDeviceHandler 创建设备管理 Handler
func NewDeviceHandler(deviceService service.DeviceService, logger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger,
	}
}

// SetDB 注入数据库连接（用于 ApproveOTA 等直接操作 DB 的场景）
func (h *DeviceHandler) SetDB(db *sql.DB) {
	h.db = db
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/devices" && r.Method == http.MethodGet:
		h.ListDevices(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") && r.Method == http.MethodGet:
		h.GetDevice(w, r)
	case strings.HasSuffix(r.URL.Path, "/ota-approve") && r.Method == http.MethodPost:
		h.ApproveOTA(w, r)
	case strings.HasSuffix(r.URL.Path, "/ota-schedule") && r.Method == http.MethodPost:
		h.SetOTASchedule(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") && r.Method == http.MethodPut:
		h.UpdateDevice(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") && r.Method == http.MethodDelete:
		h.DeleteDevice(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// GetDeviceRelations 查询设备关联关系
func (h *DeviceHandler) GetDeviceRelations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	// 路径格式：/device/api/v1/device/:id/relations
	if !strings.HasPrefix(r.URL.Path, "/device/api/v1/device/") || !strings.HasSuffix(r.URL.Path, "/relations") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	deviceID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/device/api/v1/device/"), "/relations")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 调用 Service
	req := service.GetDeviceRelationsRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := h.deviceService.GetDeviceRelations(ctx, req)
	if err != nil {
		h.logger.Error("GetDeviceRelations failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（匹配前端期望的格式）
	residents := make([]map[string]any, len(resp.Residents))
	for i, r := range resp.Residents {
		residents[i] = map[string]any{
			"id":       r.ID,
			"name":     r.Name,
			"gender":   r.Gender,
			"birthday": r.Birthday,
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"deviceId":           resp.DeviceID,
		"deviceName":         resp.DeviceName,
		"deviceInternalCode": resp.DeviceInternalCode,
		"deviceType":         resp.DeviceType,
		"addressId":          resp.AddressID,
		"addressName":        resp.AddressName,
		"addressType":        resp.AddressType,
		"residents":          residents,
	}))
}

// ListDevices 查询设备列表
// 严格限制：仅能查询本人 tenant 的设备。tenant_id 来自 X-Tenant-Id header（或 SystemAdmin 回退 System），禁止 query 覆盖。
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// status can be repeated ?status=online&status=offline or status[]=...
	statuses := r.URL.Query()["status"]
	// Some frontend uses status as array directly; if it's comma-separated, split
	if len(statuses) == 1 && strings.Contains(statuses[0], ",") {
		statuses = strings.Split(statuses[0], ",")
	}

	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	direction := strings.TrimSpace(r.URL.Query().Get("direction"))
	if direction != "desc" {
		direction = "asc"
	}

	// 2. 调用 Service
	// 严格限制：所有用户（包括 SystemAdmin）只能查看/编辑本 tenant 的设备
	req := service.ListDevicesRequest{
		TenantID:       tenantID,
		IsSystemAdmin:  false, // 始终按 tenant 过滤，SystemAdmin 也只能查看 System tenant 的设备
		Status:         statuses,
		BusinessAccess: r.URL.Query().Get("business_access"),
		DeviceType:     r.URL.Query().Get("device_type"),
		SearchType:     r.URL.Query().Get("search_type"),
		SearchKeyword:  r.URL.Query().Get("search_keyword"),
		Page:           page,
		Size:           size,
		Sort:           sort,
		Direction:      direction,
	}

	resp, err := h.deviceService.ListDevices(ctx, req)
	if err != nil {
		h.logger.Error("ListDevices failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（与旧 Handler 格式一致）
	out := make([]any, 0, len(resp.Items))
	for _, d := range resp.Items {
		out = append(out, d.ToJSON())
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": resp.Total,
	}))
}

// GetDevice 查询设备详情
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	deviceID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 调用 Service
	req := service.GetDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := h.deviceService.GetDevice(ctx, req)
	if err != nil {
		h.logger.Error("GetDevice failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（与旧 Handler 格式一致）
	writeJSON(w, http.StatusOK, Ok(resp.Device.ToJSON()))
}

// UpdateDevice 更新设备
func (h *DeviceHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	deviceID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 业务规则验证（unit_id 验证，与旧 Handler 一致）
	// 关键对齐：前端不会"只传 unit_id"，它会先 ensureUnitRoom 再传 bound_room_id
	// 因此这里收紧：如果请求里携带了 unit_id，但 bound_room_id/bound_bed_id 都为空/缺失，直接报错，避免后端兜底掩盖问题
	unitID, _ := payload["unit_id"].(string)
	if unitID != "" {
		roomVal, hasRoom := payload["bound_room_id"]
		bedVal, hasBed := payload["bound_bed_id"]
		roomEmpty := !hasRoom || roomVal == nil || roomVal == ""
		bedEmpty := !hasBed || bedVal == nil || bedVal == ""
		if roomEmpty && bedEmpty {
			writeJSON(w, http.StatusOK, Fail("invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"))
			return
		}
	}

	// 3. 数据转换（map → domain.Device）
	device := payloadToDevice(payload)

	// 4. 检查 payload 中是否包含需要更新的字段
	// 根据数据库设计文档 (17_devices.sql) 的设备绑定规则：
	//   - 绑定到 Bed：bound_bed_id IS NOT NULL, bound_room_id IS NULL
	//   - 绑定到 Room：bound_room_id IS NOT NULL, bound_bed_id IS NULL
	//   - 未绑定：bound_room_id IS NULL, bound_bed_id IS NULL
	// 前端在绑定/解绑时总是会传递这两个字段（绑定时一个为 null，解绑时两个都为 null）
	// 如果字段在 payload 中，就需要更新（即使为 null）
	_, hasBoundRoomID := payload["bound_room_id"]
	_, hasBoundBedID := payload["bound_bed_id"]
	_, hasBusinessAccess := payload["business_access"]
	_, hasMonitoringEnabled := payload["monitoring_enabled"]

	// 5. 调用 Service
	req := service.UpdateDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Device:   device,
		// 传递 payload 信息，让 Repository 层知道哪些字段需要更新
		// 如果字段在 payload 中（即使为 null），就更新它
		UpdateBoundRoomID:       hasBoundRoomID,
		UpdateBoundBedID:        hasBoundBedID,
		UpdateBusinessAccess:    hasBusinessAccess,
		UpdateMonitoringEnabled: hasMonitoringEnabled,
	}

	resp, err := h.deviceService.UpdateDevice(ctx, req)
	if err != nil {
		h.logger.Error("UpdateDevice failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 5. 构建响应（与旧 Handler 格式一致）
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": resp.Success}))
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	deviceID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 调用 Service
	req := service.DeleteDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := h.deviceService.DeleteDevice(ctx, req)
	if err != nil {
		h.logger.Error("DeleteDevice failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（与旧 Handler 格式一致）
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": resp.Success}))
}

// tenantIDFromReq 从请求中获取 tenant_id（用于 Device API）
// 重要：Device 页面查询必须限于用户自己的 tenant_id，禁止通过 query 覆盖。
// 仅使用 X-Tenant-Id header；SystemAdmin 无 header 时回退到 System 租户。
// 不使用 URL query 的 tenant_id，防止越权查看其他租户设备。
func (h *DeviceHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// ApproveOTA 租户管理员审批 OTA 升级
// POST /admin/api/v1/devices/{id}/ota-approve
func (h *DeviceHandler) ApproveOTA(w http.ResponseWriter, r *http.Request) {
	// 解析 device_id：路径格式 /admin/api/v1/devices/{id}/ota-approve
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	deviceID := strings.TrimSuffix(path, "/ota-approve")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		writeJSON(w, http.StatusOK, Fail("device_id is required"))
		return
	}

	if h.db == nil {
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		`UPDATE device_store SET ota_tenant_approved = TRUE, ota_updated_at = CURRENT_TIMESTAMP WHERE device_id = $1`,
		deviceID)
	if err != nil {
		h.logger.Error("ApproveOTA failed", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("failed to approve OTA"))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// SetOTASchedule allows tenant to update OTA schedule for their device
func (h *DeviceHandler) SetOTASchedule(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	deviceID := strings.TrimSuffix(path, "/ota-schedule")
	if deviceID == "" || strings.Contains(deviceID, "/") {
		writeJSON(w, http.StatusOK, Fail("device_id is required"))
		return
	}

	if h.db == nil {
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}

	var body struct {
		Schedule string `json:"schedule"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid request"))
		return
	}

	// Only allow schedule update when way=tenant
	var way sql.NullString
	h.db.QueryRowContext(r.Context(), `SELECT ota_way FROM device_store WHERE device_id = $1`, deviceID).Scan(&way)
	if !way.Valid || way.String != "tenant" {
		writeJSON(w, http.StatusOK, Fail("schedule can only be modified when OTA way is tenant"))
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		`UPDATE device_store SET ota_schedule = $1, ota_updated_at = CURRENT_TIMESTAMP WHERE device_id = $2`,
		sql.NullString{String: body.Schedule, Valid: body.Schedule != ""}, deviceID)
	if err != nil {
		h.logger.Error("SetOTASchedule failed", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("failed to update schedule"))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// payloadToDevice 将map[string]any转换为domain.Device
func payloadToDevice(payload map[string]any) *domain.Device {
	device := &domain.Device{}

	if v, ok := payload["device_name"].(string); ok {
		device.DeviceName = v
	}
	if v, ok := payload["device_uid"].(string); ok && v != "" {
		device.DeviceUID = v
	}
	// Handle bound_room_id: support both string and null values
	if val, exists := payload["bound_room_id"]; exists {
		if val == nil {
			// Explicit null value - set to NULL in database
			device.BoundRoomID = sql.NullString{Valid: false}
		} else if v, ok := val.(string); ok {
			if v != "" {
				device.BoundRoomID = sql.NullString{String: v, Valid: true}
			} else {
				device.BoundRoomID = sql.NullString{Valid: false}
			}
		}
	}
	// Handle bound_bed_id: support both string and null values
	if val, exists := payload["bound_bed_id"]; exists {
		if val == nil {
			// Explicit null value - set to NULL in database
			device.BoundBedID = sql.NullString{Valid: false}
		} else if v, ok := val.(string); ok {
			if v != "" {
				device.BoundBedID = sql.NullString{String: v, Valid: true}
			} else {
				device.BoundBedID = sql.NullString{Valid: false}
			}
		}
	}
	if v, ok := payload["status"].(string); ok {
		device.Status = v
	}
	// Handle business_access: must check if field exists in payload to distinguish between "not provided" and "empty string"
	if val, exists := payload["business_access"]; exists {
		if v, ok := val.(string); ok {
			device.BusinessAccess = v // Allow empty string to be set (for validation purposes)
		}
	}
	// Handle monitoring_enabled: must check if field exists in payload
	if val, exists := payload["monitoring_enabled"]; exists {
		if v, ok := val.(bool); ok {
			device.MonitoringEnabled = v
		}
	}
	if v, ok := payload["metadata"].(string); ok && v != "" {
		device.Metadata = sql.NullString{String: v, Valid: true}
	}

	return device
}
