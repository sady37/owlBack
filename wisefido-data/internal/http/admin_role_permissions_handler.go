package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// RolePermissionsHandler 角色权限管理 Handler
type RolePermissionsHandler struct {
	permService *service.RolePermissionService
	logger      *zap.Logger
}

// NewRolePermissionsHandler 创建角色权限管理 Handler
func NewRolePermissionsHandler(permService *service.RolePermissionService, logger *zap.Logger) *RolePermissionsHandler {
	return &RolePermissionsHandler{
		permService: permService,
		logger:      logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *RolePermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/role-permissions" && r.Method == http.MethodGet:
		h.ListPermissions(w, r)
	case r.URL.Path == "/admin/api/v1/role-permissions" && r.Method == http.MethodPost:
		h.CreatePermission(w, r)
	case r.URL.Path == "/admin/api/v1/role-permissions/batch" && r.Method == http.MethodPost:
		h.BatchCreatePermissions(w, r)
	case r.URL.Path == "/admin/api/v1/role-permissions/resource-types" && r.Method == http.MethodGet:
		h.GetResourceTypes(w, r)
	case strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodPut:
		h.UpdatePermissionStatus(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/") && r.Method == http.MethodPut:
		h.UpdatePermission(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/") && r.Method == http.MethodDelete:
		h.DeletePermission(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListPermissions 查询权限列表
func (h *RolePermissionsHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	roleCode := strings.TrimSpace(r.URL.Query().Get("role_code"))
	resourceType := strings.TrimSpace(r.URL.Query().Get("resource_type"))
	permType := strings.TrimSpace(r.URL.Query().Get("permission_type"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 100)

	// 2. 调用 Service
	req := service.ListPermissionsRequest{
		TenantID:       &tenantID,
		RoleCode:       roleCode,
		ResourceType:   resourceType,
		PermissionType: permType,
		Page:           page,
		Size:           size,
	}

	resp, err := h.permService.ListPermissions(ctx, req)
	if err != nil {
		h.logger.Error("ListPermissions failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// CreatePermission 创建权限
func (h *RolePermissionsHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		RoleCode       string `json:"role_code"`
		ResourceType   string `json:"resource_type"`
		PermissionType string `json:"permission_type"`
		Scope          string `json:"scope"`
		BranchOnly     bool   `json:"branch_only"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.CreatePermissionRequest{
		TenantID:       tenantID,
		UserRole:       userRole,
		RoleCode:       payload.RoleCode,
		ResourceType:   payload.ResourceType,
		PermissionType: payload.PermissionType,
		Scope:          payload.Scope,
		BranchOnly:     payload.BranchOnly,
	}

	resp, err := h.permService.CreatePermission(ctx, req)
	if err != nil {
		h.logger.Error("CreatePermission failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// BatchCreatePermissions 批量创建权限
func (h *RolePermissionsHandler) BatchCreatePermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		RoleCode    string                              `json:"role_code"`
		Permissions []service.BatchPermissionItem `json:"permissions"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.BatchCreatePermissionsRequest{
		TenantID:    tenantID,
		UserRole:    userRole,
		RoleCode:    payload.RoleCode,
		Permissions: payload.Permissions,
	}

	resp, err := h.permService.BatchCreatePermissions(ctx, req)
	if err != nil {
		h.logger.Error("BatchCreatePermissions failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdatePermission 更新权限
func (h *RolePermissionsHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	permissionID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/role-permissions/")
	if permissionID == "" || strings.Contains(permissionID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		Scope      string `json:"scope"`
		BranchOnly bool   `json:"branch_only"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.UpdatePermissionRequest{
		PermissionID: permissionID,
		TenantID:     tenantID,
		UserRole:     userRole,
		Scope:        &payload.Scope,
		BranchOnly:   &payload.BranchOnly,
	}

	err := h.permService.UpdatePermission(ctx, req)
	if err != nil {
		h.logger.Error("UpdatePermission failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// UpdatePermissionStatus 更新权限状态（删除权限）
func (h *RolePermissionsHandler) UpdatePermissionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	permissionID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/api/v1/role-permissions/"), "/status")
	if permissionID == "" || strings.Contains(permissionID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		IsActive bool `json:"is_active"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 如果 is_active 为 false，删除权限
	if !payload.IsActive {
		req := service.DeletePermissionRequest{
			PermissionID: permissionID,
			TenantID:     tenantID,
			UserRole:     userRole,
		}

		err := h.permService.DeletePermission(ctx, req)
		if err != nil {
			h.logger.Error("UpdatePermissionStatus failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeletePermission 删除权限
func (h *RolePermissionsHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	permissionID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/role-permissions/")
	if permissionID == "" || strings.Contains(permissionID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	// 2. 调用 Service
	req := service.DeletePermissionRequest{
		PermissionID: permissionID,
		TenantID:     tenantID,
		UserRole:     userRole,
	}

	err := h.permService.DeletePermission(ctx, req)
	if err != nil {
		h.logger.Error("DeletePermission failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// GetResourceTypes 获取资源类型列表
func (h *RolePermissionsHandler) GetResourceTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 调用 Service
	resp, err := h.permService.GetResourceTypes(ctx)
	if err != nil {
		h.logger.Error("GetResourceTypes failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// tenantIDFromReq 从请求中获取 tenant_id
func (h *RolePermissionsHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	// 复用 StubHandler 的逻辑
	if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
		return tid, true
	}
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

