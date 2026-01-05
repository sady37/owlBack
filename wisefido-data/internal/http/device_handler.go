package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// SystemTenantID 从 stub_handler_base.go 导入（同一包内可直接使用）
// 如果编译错误，需要确保 stub_handler_base.go 在同一包中

// DeviceHandler 设备管理 Handler
type DeviceHandler struct {
	deviceService service.DeviceService
	logger        *zap.Logger
}

// NewDeviceHandler 创建设备管理 Handler
func NewDeviceHandler(deviceService service.DeviceService, logger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/devices" && r.Method == http.MethodGet:
		h.ListDevices(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") && r.Method == http.MethodGet:
		h.GetDevice(w, r)
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
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	// SystemAdmin 可以查看所有租户的设备，如果 tenant_id 未指定或为 System 租户，允许查询所有设备
	userRole := r.Header.Get("X-User-Role")
	isSystemAdmin := strings.EqualFold(userRole, "SystemAdmin")

	var tenantID string
	var ok bool

	// 如果 URL 参数中明确指定了 tenant_id，使用它
	if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
		tenantID = tid
		ok = true
	} else if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		tenantID = tid
		ok = true
	} else if isSystemAdmin {
		// SystemAdmin 如果没有指定 tenant_id，使用 System 租户 ID（但查询时会显示所有设备）
		tenantID = SystemTenantID()
		ok = true
	}

	if !ok {
		writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
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

	// 2. 调用 Service
	// SystemAdmin 且 tenant_id 为 System 租户时，传递特殊标记以查询所有设备
	req := service.ListDevicesRequest{
		TenantID:       tenantID,
		IsSystemAdmin:  isSystemAdmin && tenantID == SystemTenantID(), // SystemAdmin 查看所有设备
		Status:         statuses,
		BusinessAccess: r.URL.Query().Get("business_access"),
		DeviceType:     r.URL.Query().Get("device_type"),
		SearchType:     r.URL.Query().Get("search_type"),
		SearchKeyword:  r.URL.Query().Get("search_keyword"),
		Page:           page,
		Size:           size,
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

	// 4. 检查 payload 中是否包含 bound_room_id 和 bound_bed_id
	// 根据数据库设计文档 (17_devices.sql) 的设备绑定规则：
	//   - 绑定到 Bed：bound_bed_id IS NOT NULL, bound_room_id IS NULL
	//   - 绑定到 Room：bound_room_id IS NOT NULL, bound_bed_id IS NULL
	//   - 未绑定：bound_room_id IS NULL, bound_bed_id IS NULL
	// 前端在绑定/解绑时总是会传递这两个字段（绑定时一个为 null，解绑时两个都为 null）
	// 如果字段在 payload 中，就需要更新（即使为 null）
	_, hasBoundRoomID := payload["bound_room_id"]
	_, hasBoundBedID := payload["bound_bed_id"]

	// 5. 调用 Service
	req := service.UpdateDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Device:   device,
		// 传递 payload 信息，让 Repository 层知道哪些字段需要更新
		// 如果字段在 payload 中（即使为 null），就更新它
		UpdateBoundRoomID: hasBoundRoomID,
		UpdateBoundBedID:  hasBoundBedID,
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

// tenantIDFromReq 从请求中获取 tenant_id（复用 AdminAPI 的逻辑）
func (h *DeviceHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.URL.Query().Get("tenant_id"); tid != "" {
		return tid, true
	}
	// Prefer tenant header (owlFront axios injects it for all requests after login)
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	// Convenience: SystemAdmin without tenant header falls back to System tenant
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// payloadToDevice 函数已在 admin_units_devices_impl.go 中定义，直接使用
