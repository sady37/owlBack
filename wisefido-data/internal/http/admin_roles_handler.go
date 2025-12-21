package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// RolesHandler 角色管理 Handler
type RolesHandler struct {
	roleService *service.RoleService
	logger      *zap.Logger
}

// NewRolesHandler 创建角色管理 Handler
func NewRolesHandler(roleService *service.RoleService, logger *zap.Logger) *RolesHandler {
	return &RolesHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *RolesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/roles" && r.Method == http.MethodGet:
		h.ListRoles(w, r)
	case r.URL.Path == "/admin/api/v1/roles" && r.Method == http.MethodPost:
		h.CreateRole(w, r)
	case strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodPut:
		h.UpdateRoleStatus(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/roles/") && r.Method == http.MethodPut:
		h.UpdateRole(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/roles/") && r.Method == http.MethodDelete:
		h.DeleteRole(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListRoles 查询角色列表
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)

	// 2. 调用 Service
	req := service.ListRolesRequest{
		TenantID: &tenantID,
		Search:   search,
		Page:     page,
		Size:     size,
	}

	resp, err := h.roleService.ListRoles(ctx, req)
	if err != nil {
		h.logger.Error("ListRoles failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// CreateRole 创建角色
func (h *RolesHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload struct {
		RoleCode    string `json:"role_code"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.CreateRoleRequest{
		TenantID:    tenantID,
		RoleCode:    payload.RoleCode,
		DisplayName: payload.DisplayName,
		Description: payload.Description,
	}

	resp, err := h.roleService.CreateRole(ctx, req)
	if err != nil {
		h.logger.Error("CreateRole failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdateRole 更新角色
func (h *RolesHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	roleID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/")
	if roleID == "" || strings.Contains(roleID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 构建请求
	req := service.UpdateRoleRequest{
		RoleID:   roleID,
		UserRole: userRole,
	}

	// 处理 is_active
	if v, ok := payload["is_active"].(bool); ok {
		req.IsActive = &v
	}

	// 处理 _delete
	if v, ok := payload["_delete"].(bool); ok && v {
		deleteFlag := true
		req.Delete = &deleteFlag
	}

	// 处理 display_name 和 description
	if v, ok := payload["display_name"].(string); ok {
		req.DisplayName = &v
	}
	if v, ok := payload["description"].(string); ok {
		req.Description = &v
	}

	// 3. 调用 Service
	err := h.roleService.UpdateRole(ctx, req)
	if err != nil {
		h.logger.Error("UpdateRole failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// UpdateRoleStatus 更新角色状态
func (h *RolesHandler) UpdateRoleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	roleID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/"), "/status")
	if roleID == "" || strings.Contains(roleID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var payload struct {
		IsActive bool `json:"is_active"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.UpdateRoleRequest{
		RoleID:   roleID,
		IsActive: &payload.IsActive,
	}

	err := h.roleService.UpdateRole(ctx, req)
	if err != nil {
		h.logger.Error("UpdateRoleStatus failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteRole 删除角色
func (h *RolesHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	roleID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/")
	if roleID == "" || strings.Contains(roleID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 2. 调用 Service
	req := service.UpdateRoleRequest{
		RoleID: roleID,
		Delete: func() *bool { b := true; return &b }(),
	}

	err := h.roleService.UpdateRole(ctx, req)
	if err != nil {
		h.logger.Error("DeleteRole failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// tenantIDFromReq 从请求中获取 tenant_id
func (h *RolesHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
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
