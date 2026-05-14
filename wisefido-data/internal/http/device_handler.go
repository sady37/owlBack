package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	case strings.HasSuffix(r.URL.Path, "/reset") && r.Method == http.MethodPost:
		h.ResetDevicePrefix(w, r)
	case strings.HasSuffix(r.URL.Path, "/ota-approve") && r.Method == http.MethodPost:
		h.ApproveOTA(w, r)
	case strings.HasSuffix(r.URL.Path, "/ota-schedule") && r.Method == http.MethodPost:
		h.SetOTASchedule(w, r)
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

// ResetDevicePrefix 解绑 device 到指定层级（清零 device_ipv6 字节段，保留 MAC）。
// 路径：POST /admin/api/v1/devices/{device_ipv6}/reset?layer=branch|site|unit|room|bed
// 设计说明详见 memory/device_unbind_via_prefix_reset.md：
//   - device_ipv6 (PK /128) 不删行；UPDATE 清零 [byte_start..11]，bytes 12-15 (MAC) 保留
//   - 权限：调用者 role.scope_prefix_len 必须 ≤ byte_start*8（不能清零 scope 之外的字节）
//   - SQL 集中校验：reset_device_prefix(addr, session_prefix, session_masklen, layer)
//   - 触发器已放宽：unbound device 落在 branch /56 池内合法
func (h *DeviceHandler) ResetDevicePrefix(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}

	// 1. 解析 device_ipv6 from path: /admin/api/v1/devices/{addr}/reset
	// addr 是 IPv6 CIDR /128，含 '/'；不能简单按 '/' split，要走 stitchCIDRSegments
	rest := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
	rest = strings.TrimSuffix(rest, "/reset")
	if rest == "" || isMultiSegmentPath(rest) {
		writeJSON(w, http.StatusOK, Fail("device_ipv6 is required (IPv6 CIDR /128)"))
		return
	}
	spatialAddr := rest

	// 2. layer 参数
	layer := r.URL.Query().Get("layer")
	switch layer {
	case "branch", "site", "unit", "room", "bed":
		// ok
	default:
		writeJSON(w, http.StatusOK, Fail("layer must be one of: branch, site, unit, room, bed"))
		return
	}

	// 3. 调用者 scope：role + tenant_id
	role := r.Header.Get("X-User-Role")
	if role == "" {
		writeJSON(w, http.StatusOK, Fail("X-User-Role header required"))
		return
	}
	tenantID := r.Header.Get("X-Tenant-Id")

	// 查 roles.scope_prefix_len
	var scopeLen int
	err := h.db.QueryRowContext(r.Context(),
		`SELECT scope_prefix_len FROM roles WHERE LOWER(role_code) = LOWER($1) AND is_system = TRUE LIMIT 1`,
		role).Scan(&scopeLen)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, Fail("unknown role: "+role))
		return
	}
	if err != nil {
		h.logger.Error("ResetDevicePrefix: lookup role failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("role lookup failed"))
		return
	}

	// 决定 session_prefix：platform=fd00::/32, tenant=X-Tenant-Id, 其它=暂不支持
	var sessionPrefix string
	switch scopeLen {
	case 32:
		sessionPrefix = "fd00::/32"
	case 48:
		if tenantID == "" {
			writeJSON(w, http.StatusOK, Fail("X-Tenant-Id header required for tenant-scoped role"))
			return
		}
		sessionPrefix = tenantID
	default:
		// branch_manager (56) / unit_manager (80) 需 X-Branch-Id / X-Unit-Id；先 not-implemented
		writeJSON(w, http.StatusOK, Fail("role scope not yet wired (only platform_admin and tenant_admin supported in this PR)"))
		return
	}

	// 4. UPDATE devices 用 reset_device_prefix() 派生新 addr
	var newAddr string
	err = h.db.QueryRowContext(r.Context(), `
		UPDATE devices
		   SET device_ipv6 = reset_device_prefix(device_ipv6, $2::INET, $3::SMALLINT, $4),
		       updated_at = NOW()
		 WHERE device_ipv6 = $1::INET
		RETURNING host(device_ipv6) || '/128'`,
		spatialAddr, sessionPrefix, scopeLen, layer).Scan(&newAddr)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, Fail("device not found: "+spatialAddr))
		return
	}
	if err != nil {
		h.logger.Error("ResetDevicePrefix: UPDATE failed",
			zap.String("addr", spatialAddr), zap.String("layer", layer), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"old_device_ipv6": spatialAddr,
		"new_device_ipv6": newAddr,
		"layer":            layer,
		"caller_role":      role,
		"caller_scope":     sessionPrefix,
	}))
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
		IsSystemAdmin:  false, // 始终按 tenant 过滤
		CurrentUserID:  r.Header.Get("X-User-Id"),
		Status:         statuses,
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
	_, hasAccess := payload["access"]
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
		UpdateAccess: hasAccess,
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

	// Support toggle: {approved: true/false}, default true for backwards compat
	approved := true
	var body struct {
		Approved *bool `json:"approved"`
	}
	if json.NewDecoder(r.Body).Decode(&body) == nil && body.Approved != nil {
		approved = *body.Approved
	}

	// v2: tenant approve = device_ota.approve_way 前缀切换 (system_* ↔ tenant_*)。
	// suffix (_manual/_schedule) 保持现状；approve_way 为 NULL 时无 OTA 计划，跳过。
	var newPrefix string
	if approved {
		newPrefix = "tenant"
	} else {
		newPrefix = "system"
	}
	_, err := h.db.ExecContext(r.Context(), `
		UPDATE device_ota o
		   SET approve_way = $1 || '_' || SPLIT_PART(o.approve_way, '_', 2),
		       updated_at = NOW()
		  FROM devices d
		 WHERE o.device_ipv6 = d.device_ipv6
		   AND d.device_id = $2::uuid
		   AND o.approve_way IS NOT NULL
	`, newPrefix, deviceID)
	if err != nil {
		h.logger.Error("ApproveOTA failed", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("failed to update OTA approval"))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "approved": approved}))
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

	// v2: 仅当 approve_way LIKE 'tenant_%' 时允许租户改 schedule。
	var approveWay sql.NullString
	h.db.QueryRowContext(r.Context(), `
		SELECT o.approve_way
		  FROM device_ota o
		  JOIN devices d ON d.device_ipv6 = o.device_ipv6
		 WHERE d.device_id = $1::uuid
	`, deviceID).Scan(&approveWay)
	if !approveWay.Valid || !strings.HasPrefix(approveWay.String, "tenant_") {
		writeJSON(w, http.StatusOK, Fail("schedule can only be modified when OTA way is tenant"))
		return
	}

	// v1 schedule 形如 "now+3d"/"MMDDHH+xH" — 直接落 text 字段以保持向后兼容会破 v2 CHECK；
	// 让 wisefido-data Repository 层 mapV1OTAWayToV2 类似函数解析。这里仅做最小切换：
	// 空串 → NULL（取消）；非空时解析为 timestamptz。
	schedTime := parseV1OTAScheduleToTimestamp(body.Schedule)
	_, err := h.db.ExecContext(r.Context(), `
		UPDATE device_ota o
		   SET schedule = $1, updated_at = NOW()
		  FROM devices d
		 WHERE o.device_ipv6 = d.device_ipv6
		   AND d.device_id = $2::uuid
	`, schedTime, deviceID)
	if err != nil {
		h.logger.Error("SetOTASchedule failed", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("failed to update schedule"))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// parseV1OTAScheduleToTimestamp 接 v1 形 schedule 字串 → sql.NullTime (timestamptz)。
// 空串 → NULL；不可解析 → NULL（与 repository.parseV1OTAScheduleToTime 同语义，简化版）。
func parseV1OTAScheduleToTimestamp(v1 string) sql.NullTime {
	s := strings.TrimSpace(v1)
	if s == "" {
		return sql.NullTime{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	// v1 形态由 wisefido-data Repository 解析；这里 caller 一般传 RFC3339。
	return sql.NullTime{}
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
	// Handle access: platform_admin 审批位（bool）
	if val, exists := payload["access"]; exists {
		if v, ok := val.(bool); ok {
			device.Access = v
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
