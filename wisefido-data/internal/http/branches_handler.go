package httpapi

import (
	"database/sql"
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// BranchesHandler 院区管理 Handler
type BranchesHandler struct {
	branchService *service.BranchService
	logger        *zap.Logger
	base          *StubHandler // 用于 tenantIDFromReq
	db            *sql.DB      // 用于权限检查
}

// NewBranchesHandler 创建院区管理 Handler
func NewBranchesHandler(branchService *service.BranchService, db *sql.DB, logger *zap.Logger) *BranchesHandler {
	return &BranchesHandler{
		branchService: branchService,
		logger:        logger,
		base:          &StubHandler{}, // 用于 tenantIDFromReq
		db:            db,             // 用于权限检查
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *BranchesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/branches" && r.Method == http.MethodGet:
		h.ListBranches(w, r)
	case r.URL.Path == "/admin/api/v1/branches" && r.Method == http.MethodPost:
		h.CreateBranch(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/branches/") && r.Method == http.MethodPut:
		h.UpdateBranch(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/branches/") && r.Method == http.MethodDelete:
		h.DeleteBranch(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListBranches 查询院区列表
func (h *BranchesHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 权限检查：检查是否有 R 权限
	currentUserRole := r.Header.Get("X-User-Role")
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserRole == "" {
		writeJSON(w, http.StatusOK, Fail("user role is required"))
		return
	}

	// 检查是否有读取权限
	perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "branches", "R")
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("permission check failed"))
		return
	}
	// 如果没有权限（返回默认严格权限），检查是否是允许的角色
	if perm.AssignedOnly && perm.BranchOnly && currentUserRole != "SystemAdmin" && currentUserRole != "SystemOperator" && currentUserRole != "Admin" && currentUserRole != "Manager" {
		writeJSON(w, http.StatusOK, Fail("permission denied: only SystemAdmin, SystemOperator, Admin, and Manager can view branches"))
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)

	// 3. 调用 Service（传递权限信息和用户信息）
	req := service.ListBranchesRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		BranchOnly:      perm.BranchOnly, // Manager 只能查看本 Branch
		Search:          search,
		Page:            page,
		Size:            size,
	}

	resp, err := h.branchService.ListBranches(ctx, req)
	if err != nil {
		h.logger.Error("ListBranches failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	// is_primary 标记：FE 自己拿 store.current_branch_id 与每行 branch_id 比对即可，BE 无需再注入
	writeJSON(w, http.StatusOK, Ok(resp))
}

// CreateBranch 创建院区
func (h *BranchesHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 权限检查：只允许 SystemAdmin, SystemOperator, Admin 创建
	currentUserRole := r.Header.Get("X-User-Role")
	if currentUserRole != "SystemAdmin" && currentUserRole != "SystemOperator" && currentUserRole != "Admin" {
		writeJSON(w, http.StatusOK, Fail("permission denied: only SystemAdmin, SystemOperator, and Admin can create branches"))
		return
	}

	// 检查是否有创建权限
	perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "branches", "C")
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("permission check failed"))
		return
	}
	// 如果没有权限（返回默认严格权限），拒绝
	if perm.AssignedOnly && perm.BranchOnly && currentUserRole != "SystemAdmin" && currentUserRole != "SystemOperator" && currentUserRole != "Admin" {
		writeJSON(w, http.StatusOK, Fail("permission denied: no create permission for branches"))
		return
	}

	var payload struct {
		BranchName  string `json:"branch_name"`
		Description string `json:"description"`
		Timezone    string `json:"timezone"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 验证：禁止空字符串作为 BranchName
	branchNameTrimmed := strings.TrimSpace(payload.BranchName)
	if branchNameTrimmed == "" {
		writeJSON(w, http.StatusOK, Fail("branch_name cannot be empty"))
		return
	}

	// 3. 调用 Service
	req := service.CreateBranchRequest{
		TenantID:    tenantID,
		BranchName:  branchNameTrimmed,
		Description: payload.Description,
		Timezone:    strings.TrimSpace(payload.Timezone),
	}

	resp, err := h.branchService.CreateBranch(ctx, req)
	if err != nil {
		h.logger.Error("CreateBranch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdateBranch 更新院区
func (h *BranchesHandler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	branchID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/branches/")
	// v2: branch_id 是 IPv6 prefix CIDR（'fd00:0:3:100::/56'），URL 自带 '/' 是合法的；
	// 仅拒绝多段路径（如 ".../foo/bar/baz" 这种 sub-action 格式）
	if branchID == "" || isMultiSegmentPath(branchID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 2. 权限检查：v2 RBAC — 通过 role_permissions 表（'branches.update' / 'branches.*' / 'branches.config' / 'tenant.*' / '*'）。
	// Manager 通过 seed 'manager: branches.config' 拿到 update 权（自管 branch；scope 校验留待 user_roles.scope 数据齐全后做）。
	currentUserRole := r.Header.Get("X-User-Role")
	perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "branches", "U")
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("permission check failed"))
		return
	}
	if perm.AssignedOnly && perm.BranchOnly {
		writeJSON(w, http.StatusOK, Fail("permission denied: role cannot update branches"))
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 3. 构建请求
	req := service.UpdateBranchRequest{
		BranchID: branchID,
	}

	// 处理 _delete
	if v, ok := payload["_delete"].(bool); ok && v {
		deleteFlag := true
		req.Delete = &deleteFlag
	}

	// 处理 branch_name：禁止空字符串
	if v, ok := payload["branch_name"].(string); ok {
		branchNameTrimmed := strings.TrimSpace(v)
		if branchNameTrimmed == "" {
			writeJSON(w, http.StatusOK, Fail("branch_name cannot be empty"))
			return
		}
		req.BranchName = &branchNameTrimmed
	}

	// 处理 description
	if v, ok := payload["description"].(string); ok {
		req.Description = &v
	}

	// 处理 timezone
	if v, ok := payload["timezone"].(string); ok {
		req.Timezone = &v
	}

	// 4. 调用 Service
	err = h.branchService.UpdateBranch(ctx, req)
	if err != nil {
		h.logger.Error("UpdateBranch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 5. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteBranch 删除院区
func (h *BranchesHandler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	branchID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/branches/")
	// v2: branch_id 是 IPv6 prefix CIDR（'fd00:0:3:100::/56'），URL 自带 '/' 是合法的；
	// 仅拒绝多段路径（如 ".../foo/bar/baz" 这种 sub-action 格式）
	if branchID == "" || isMultiSegmentPath(branchID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 2. 权限检查：只允许 SystemAdmin, SystemOperator, Admin 删除
	currentUserRole := r.Header.Get("X-User-Role")
	if currentUserRole != "SystemAdmin" && currentUserRole != "SystemOperator" && currentUserRole != "Admin" {
		writeJSON(w, http.StatusOK, Fail("permission denied: only SystemAdmin, SystemOperator, and Admin can delete branches"))
		return
	}

	// 检查是否有删除权限
	perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "branches", "D")
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("permission check failed"))
		return
	}
	// 如果没有权限（返回默认严格权限），拒绝
	if perm.AssignedOnly && perm.BranchOnly && currentUserRole != "SystemAdmin" && currentUserRole != "SystemOperator" && currentUserRole != "Admin" {
		writeJSON(w, http.StatusOK, Fail("permission denied: no delete permission for branches"))
		return
	}

	// 3. 调用 Service
	req := service.UpdateBranchRequest{
		BranchID: branchID,
		Delete:   func() *bool { b := true; return &b }(),
	}

	err = h.branchService.UpdateBranch(ctx, req)
	if err != nil {
		h.logger.Error("DeleteBranch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}
