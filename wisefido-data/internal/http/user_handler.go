package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// UserHandler 用户管理 Handler
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
	base        *StubHandler // 用于 tenantIDFromReq
}

// NewUserHandler 创建用户管理 Handler
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		base:        &StubHandler{}, // 用于 tenantIDFromReq
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	path := r.URL.Path
	switch {
	// ListUsers
	case path == "/admin/api/v1/users" && r.Method == http.MethodGet:
		h.ListUsers(w, r)
	// CreateUser
	case path == "/admin/api/v1/users" && r.Method == http.MethodPost:
		h.CreateUser(w, r)
	// GetUser
	case strings.HasPrefix(path, "/admin/api/v1/users/") && r.Method == http.MethodGet:
		userID := strings.TrimPrefix(path, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.GetUser(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateUser
	case strings.HasPrefix(path, "/admin/api/v1/users/") && r.Method == http.MethodPut:
		userID := strings.TrimPrefix(path, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.UpdateUser(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// DeleteUser
	case strings.HasPrefix(path, "/admin/api/v1/users/") && r.Method == http.MethodDelete:
		userID := strings.TrimPrefix(path, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.DeleteUser(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// ResetPassword
	case strings.HasSuffix(path, "/reset-password") && r.Method == http.MethodPost:
		userID := strings.TrimSuffix(path, "/reset-password")
		userID = strings.TrimPrefix(userID, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.ResetPassword(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// ResetPIN
	case strings.HasSuffix(path, "/reset-pin") && r.Method == http.MethodPost:
		userID := strings.TrimSuffix(path, "/reset-pin")
		userID = strings.TrimPrefix(userID, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.ResetPIN(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ============================================
// ListUsers 查询用户列表
// ============================================

// ListUsers 查询用户列表
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)

	req := service.ListUsersRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,
		Search:        search,
		Page:          page,
		Size:          size,
	}

	resp, err := h.userService.ListUsers(ctx, req)
	if err != nil {
		h.logger.Error("ListUsers failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式
	items := make([]any, 0, len(resp.Items))
	for _, u := range resp.Items {
		item := map[string]any{
			"user_id":      u.UserID,
			"tenant_id":    u.TenantID,
			"user_account": u.UserAccount,
			"role":         u.Role,
			"status":       u.Status,
		}
		if u.Nickname != "" {
			item["nickname"] = u.Nickname
		}
		if u.Email != "" {
			item["email"] = u.Email
		}
		if u.Phone != "" {
			item["phone"] = u.Phone
		}
		if len(u.AlarmLevels) > 0 {
			item["alarm_levels"] = u.AlarmLevels
		}
		if len(u.AlarmChannels) > 0 {
			item["alarm_channels"] = u.AlarmChannels
		}
		if u.AlarmScope != "" {
			item["alarm_scope"] = u.AlarmScope
		}
		if u.BranchTag != "" {
			item["branch_tag"] = u.BranchTag
		}
		if u.LastLoginAt != "" {
			item["last_login_at"] = u.LastLoginAt
		}
		if len(u.Tags) > 0 {
			item["tags"] = u.Tags
		}
		if u.Preferences != nil {
			item["preferences"] = u.Preferences
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": resp.Total,
	}))
}

// ============================================
// GetUser 查询用户详情
// ============================================

// GetUser 查询用户详情
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	req := service.GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	resp, err := h.userService.GetUser(ctx, req)
	if err != nil {
		h.logger.Error("GetUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式
	item := map[string]any{
		"user_id":      resp.User.UserID,
		"tenant_id":    resp.User.TenantID,
		"user_account": resp.User.UserAccount,
		"role":         resp.User.Role,
		"status":       resp.User.Status,
	}
	if resp.User.Nickname != "" {
		item["nickname"] = resp.User.Nickname
	}
	if resp.User.Email != "" {
		item["email"] = resp.User.Email
	}
	if resp.User.Phone != "" {
		item["phone"] = resp.User.Phone
	}
	if len(resp.User.AlarmLevels) > 0 {
		item["alarm_levels"] = resp.User.AlarmLevels
	}
	if len(resp.User.AlarmChannels) > 0 {
		item["alarm_channels"] = resp.User.AlarmChannels
	}
	if resp.User.AlarmScope != "" {
		item["alarm_scope"] = resp.User.AlarmScope
	}
	if resp.User.BranchTag != "" {
		item["branch_tag"] = resp.User.BranchTag
	}
	if resp.User.LastLoginAt != "" {
		item["last_login_at"] = resp.User.LastLoginAt
	}
	if len(resp.User.Tags) > 0 {
		item["tags"] = resp.User.Tags
	}
	if resp.User.Preferences != nil {
		item["preferences"] = resp.User.Preferences
	}

	writeJSON(w, http.StatusOK, Ok(item))
}

// ============================================
// CreateUser 创建用户
// ============================================

// CreateUser 创建用户
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 解析必填字段
	userAccount, _ := payload["user_account"].(string)
	role, _ := payload["role"].(string)
	password, _ := payload["password"].(string)

	if strings.TrimSpace(userAccount) == "" || strings.TrimSpace(role) == "" || password == "" {
		writeJSON(w, http.StatusOK, Fail("user_account, role, password are required"))
		return
	}

	// 解析可选字段
	nickname, _ := payload["nickname"].(string)
	email, _ := payload["email"].(string)
	phone, _ := payload["phone"].(string)
	status, _ := payload["status"].(string)

	// 解析 alarm_levels
	var alarmLevels []string
	if levels, ok := payload["alarm_levels"].([]any); ok {
		for _, l := range levels {
			if s, ok := l.(string); ok && s != "" {
				alarmLevels = append(alarmLevels, s)
			}
		}
	}

	// 解析 alarm_channels
	var alarmChannels []string
	if channels, ok := payload["alarm_channels"].([]any); ok {
		for _, c := range channels {
			if s, ok := c.(string); ok && s != "" {
				alarmChannels = append(alarmChannels, s)
			}
		}
	}

	// 解析 alarm_scope
	alarmScope, _ := payload["alarm_scope"].(string)

	// 解析 tags
	var tags []string
	if tagsRaw, ok := payload["tags"].([]any); ok {
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok && s != "" {
				tags = append(tags, s)
			}
		}
	}

	// 解析 branch_tag
	branchTag, _ := payload["branch_tag"].(string)

	req := service.CreateUserRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,
		UserAccount:   userAccount,
		Password:      password,
		Role:          role,
		Nickname:      nickname,
		Email:         email,
		Phone:         phone,
		Status:        status,
		AlarmLevels:   alarmLevels,
		AlarmChannels: alarmChannels,
		AlarmScope:    alarmScope,
		Tags:          tags,
		BranchTag:     branchTag,
	}

	resp, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		h.logger.Error("CreateUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"user_id": resp.UserID,
	}))
}

// ============================================
// UpdateUser 更新用户
// ============================================

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 检查是否为软删除
	if del, ok := payload["_delete"].(bool); ok && del {
		req := service.DeleteUserRequest{
			TenantID:      tenantID,
			UserID:        userID,
			CurrentUserID: currentUserID,
		}

		resp, err := h.userService.DeleteUser(ctx, req)
		if err != nil {
			h.logger.Error("DeleteUser failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}

		if resp.Success {
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		} else {
			writeJSON(w, http.StatusOK, Fail("failed to delete user"))
		}
		return
	}

	// 解析可选字段（nil 表示不更新，空字符串表示清空）
	req := service.UpdateUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	// Nickname
	if nickname, ok := payload["nickname"].(string); ok {
		req.Nickname = &nickname
	}

	// Email 和 EmailHash（复杂逻辑）
	if emailVal, ok := payload["email"]; ok {
		if emailVal == nil {
			// null 表示删除 email
			emptyEmail := ""
			req.Email = &emptyEmail
		} else if email, ok := emailVal.(string); ok {
			req.Email = &email
		}
	}
	if emailHashHex, ok := payload["email_hash"].(string); ok {
		req.EmailHash = &emailHashHex
	}

	// Phone 和 PhoneHash（同 Email）
	if phoneVal, ok := payload["phone"]; ok {
		if phoneVal == nil {
			emptyPhone := ""
			req.Phone = &emptyPhone
		} else if phone, ok := phoneVal.(string); ok {
			req.Phone = &phone
		}
	}
	if phoneHashHex, ok := payload["phone_hash"].(string); ok {
		req.PhoneHash = &phoneHashHex
	}

	// Role
	if role, ok := payload["role"].(string); ok {
		req.Role = &role
	}

	// Status
	if status, ok := payload["status"].(string); ok {
		req.Status = &status
	}

	// AlarmLevels
	if levels, ok := payload["alarm_levels"].([]any); ok {
		alarmLevels := make([]string, 0, len(levels))
		for _, l := range levels {
			if s, ok := l.(string); ok && s != "" {
				alarmLevels = append(alarmLevels, s)
			}
		}
		req.AlarmLevels = alarmLevels
	}

	// AlarmChannels
	if channels, ok := payload["alarm_channels"].([]any); ok {
		alarmChannels := make([]string, 0, len(channels))
		for _, c := range channels {
			if s, ok := c.(string); ok && s != "" {
				alarmChannels = append(alarmChannels, s)
			}
		}
		req.AlarmChannels = alarmChannels
	}

	// AlarmScope
	if scope, ok := payload["alarm_scope"].(string); ok {
		req.AlarmScope = &scope
	}

	// Tags
	if tagsRaw, ok := payload["tags"].([]any); ok {
		tags := make([]string, 0, len(tagsRaw))
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok && s != "" {
				tags = append(tags, s)
			}
		}
		req.Tags = tags
	}

	// BranchTag
	if branchTag, ok := payload["branch_tag"].(string); ok {
		req.BranchTag = &branchTag
	}

	resp, err := h.userService.UpdateUser(ctx, req)
	if err != nil {
		h.logger.Error("UpdateUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	if resp.Success {
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
	} else {
		writeJSON(w, http.StatusOK, Fail("failed to update user"))
	}
}

// ============================================
// DeleteUser 删除用户
// ============================================

// DeleteUser 删除用户（软删除）
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	req := service.DeleteUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	resp, err := h.userService.DeleteUser(ctx, req)
	if err != nil {
		h.logger.Error("DeleteUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	if resp.Success {
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
	} else {
		writeJSON(w, http.StatusOK, Fail("failed to delete user"))
	}
}

// ============================================
// ResetPassword 重置密码
// ============================================

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	newPassword, _ := payload["new_password"].(string)
	if newPassword == "" {
		writeJSON(w, http.StatusOK, Fail("new_password is required"))
		return
	}

	req := service.UserResetPasswordRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
		NewPassword:   newPassword,
	}

	resp, err := h.userService.ResetPassword(ctx, req)
	if err != nil {
		h.logger.Error("ResetPassword failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
		"message": resp.Message,
	}))
}

// ============================================
// ResetPIN 重置 PIN
// ============================================

// ResetPIN 重置 PIN
func (h *UserHandler) ResetPIN(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	newPIN, _ := payload["new_pin"].(string)
	if newPIN == "" {
		writeJSON(w, http.StatusOK, Fail("new_pin is required"))
		return
	}

	req := service.UserResetPINRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
		NewPIN:        newPIN,
	}

	resp, err := h.userService.ResetPIN(ctx, req)
	if err != nil {
		h.logger.Error("ResetPIN failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}

