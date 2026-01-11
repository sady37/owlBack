package httpapi

import (
	"database/sql"
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
	db          *sql.DB      // 用于权限检查和获取 branch 信息
}

// NewUserHandler 创建用户管理 Handler
func NewUserHandler(userService service.UserService, db *sql.DB, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		base:        &StubHandler{}, // 用于 tenantIDFromReq
		db:          db,
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
	// GetAvailableBranches
	case path == "/admin/api/v1/users/branches" && r.Method == http.MethodGet:
		h.GetAvailableBranches(w, r)
	// GetAvailableCaregivers
	case path == "/admin/api/v1/users/caregivers" && r.Method == http.MethodGet:
		h.GetAvailableCaregivers(w, r)
	// GetAvailableCaregiverGroups
	case path == "/admin/api/v1/users/caregiver-groups" && r.Method == http.MethodGet:
		h.GetAvailableCaregiverGroups(w, r)
	// GetAccountSettings (必须在 GetUser 之前，因为路径更具体)
	case strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodGet:
		userID := strings.TrimSuffix(path, "/account-settings")
		userID = strings.TrimPrefix(userID, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.GetAccountSettings(w, r, userID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateAccountSettings (必须在 UpdateUser 之前，因为路径更具体)
	case strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodPut:
		userID := strings.TrimSuffix(path, "/account-settings")
		userID = strings.TrimPrefix(userID, "/admin/api/v1/users/")
		if userID != "" && !strings.Contains(userID, "/") {
			h.UpdateAccountSettings(w, r, userID)
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
		// 始终返回 branch_ids，即使是空数组或 ["*"]
		// 空数组表示未分配，["*"] 表示所有分支（仅 Admin）
		if u.BranchIDs != nil {
			item["branch_ids"] = u.BranchIDs
		}
		// 向后兼容
		if u.BranchID != "" {
			item["branch_id"] = u.BranchID
		}
		if u.BranchName != "" {
			item["branch_name"] = u.BranchName
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

	// 解析 BranchIDs 查询参数（Create 模式时使用）
	var branchIDs []string
	if branchIDsParam := r.URL.Query().Get("branch_ids"); branchIDsParam != "" {
		// 支持逗号分隔的 branch_ids，例如：?branch_ids=id1,id2,id3
		branchIDsList := strings.Split(branchIDsParam, ",")
		for _, id := range branchIDsList {
			id = strings.TrimSpace(id)
			if id != "" {
				branchIDs = append(branchIDs, id)
			}
		}
	}

	req := service.GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
		BranchIDs:     branchIDs,
	}

	resp, err := h.userService.GetUser(ctx, req)
	if err != nil {
		h.logger.Error("GetUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式
	// Create 模式下，resp.User 为 nil，返回 available_tags 和 available_branches
	if resp.User == nil {
		// Create 模式：返回最小化的响应，包含 available_tags 和 available_branches
		// 前端已有 emptyUser 初始化表单，不需要用户对象的默认值
		// 但为了符合 User 接口类型要求，返回最小化的用户对象
		item := map[string]any{
			"user_id":            userID, // 对于 'new'，返回 'new'
			"tenant_id":          tenantID,
			"user_account":       "",
			"role":               "",
			"status":             "active",
			"available_tags":     []string{},
			"available_branches": []map[string]any{},
		}
		if len(resp.AvailableTags) > 0 {
			item["available_tags"] = resp.AvailableTags
		}
		// 转换 AvailableBranches 为前端格式
		if len(resp.AvailableBranches) > 0 {
			branches := make([]map[string]any, 0, len(resp.AvailableBranches))
			for _, branch := range resp.AvailableBranches {
				branches = append(branches, map[string]any{
					"branch_id":   branch.BranchID,
					"branch_name": branch.BranchName,
				})
			}
			item["available_branches"] = branches
		}
		writeJSON(w, http.StatusOK, Ok(item))
		return
	}

	// Edit 模式：返回完整的用户信息
	item := map[string]any{
		"user_id":      resp.User.UserID,
		"tenant_id":    resp.User.TenantID,
		"user_account": resp.User.UserAccount,
		"role":         resp.User.Role,
		"status":       resp.User.Status,
	}

	if resp.User != nil {
		// 普通字段：字段不存在/空/null → 不返回字段
		if resp.User.Nickname != "" {
			item["nickname"] = resp.User.Nickname
		}

		// 有 Hash 的字段（email/email_hash, phone/phone_hash）：
		// Service 层返回：
		// - 有值 → 实际值
		// - 无值但 hash 有值 → 占位符（"xxx@xxx.xxx" / "xxx-xxx-xxxx"）
		// - 无值且 hash 无值 → 空字符串（表示 null）
		// Handler 层处理：
		// - 空字符串 → 返回 null
		// - 占位符 → 返回占位符
		// - 有值 → 返回值
		if resp.User.Email != "" {
			if resp.User.Email == "xxx@xxx.xxx" {
				// 占位符，返回占位符
				item["email"] = resp.User.Email
			} else {
				// 有值，返回值
				item["email"] = resp.User.Email
			}
		} else {
			// 空字符串，返回 null
			item["email"] = nil
		}

		if resp.User.Phone != "" {
			if resp.User.Phone == "xxx-xxx-xxxx" {
				// 占位符，返回占位符
				item["phone"] = resp.User.Phone
			} else {
				// 有值，返回值
				item["phone"] = resp.User.Phone
			}
		} else {
			// 空字符串，返回 null
			item["phone"] = nil
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
		if len(resp.User.BranchIDs) > 0 {
			item["branch_ids"] = resp.User.BranchIDs
		}
		// 向后兼容
		if resp.User.BranchID != "" {
			item["branch_id"] = resp.User.BranchID
		}
		if resp.User.BranchName != "" {
			item["branch_name"] = resp.User.BranchName
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
	}
	// 返回可用 tags（当前用户所在 Branch 中存在的 tags）
	if len(resp.AvailableTags) > 0 {
		item["available_tags"] = resp.AvailableTags
	} else {
		item["available_tags"] = []string{}
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

	// 解析 branch_ids（支持多个 branch，向后兼容单个 branch_id）
	var branchIDs []string

	// 优先使用 branch_ids（数组）
	if branchIDsVal, ok := payload["branch_ids"].([]any); ok && len(branchIDsVal) > 0 {
		for _, v := range branchIDsVal {
			if s, ok := v.(string); ok && s != "" {
				branchIDs = append(branchIDs, s)
			}
		}
	}
	// 向后兼容：如果 branch_ids 为空，尝试使用单个 branch_id
	if len(branchIDs) == 0 {
		if branchIDVal, ok := payload["branch_id"].(string); ok && branchIDVal != "" {
			branchIDs = append(branchIDs, branchIDVal)
		}
	}

	if len(branchIDs) == 0 {
		writeJSON(w, http.StatusOK, Fail("at least one branch_id is required"))
		return
	}

	// 注意：AvailableBranches 不应传递给 Service 层
	// Service 层会自己从数据库查询用户的 branch 信息（这是用户本身的属性，不能信任前端传递的值）
	// 如果前端需要获取可用 branch 列表，应该调用专门的 API（如 GetAvailableBranches）
	// 注意：primary_branch_id 已移除，Service 层会使用第一个 branch 用于向后兼容

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
		BranchIDs:     branchIDs,
	}

	resp, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		h.logger.Error("CreateUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	result := map[string]any{
		"user_id": resp.UserID,
	}
	// 如果是 Individual 用户，返回 Resident 创建状态
	if resp.ResidentCreated || resp.ResidentError != "" {
		result["resident_created"] = resp.ResidentCreated
		if resp.ResidentError != "" {
			result["resident_error"] = resp.ResidentError
		}
	}

	writeJSON(w, http.StatusOK, Ok(result))
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

	// 注意：AvailableBranches 不应传递给 Service 层
	// Service 层会自己从数据库查询用户的 branch 信息（这是用户本身的属性，不能信任前端传递的值）
	// 如果前端需要获取可用 branch 列表，应该调用专门的 API（如 GetAvailableBranches）

	// 解析可选字段（按照统一设计规则）
	req := service.UpdateUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	// 统一的字段解析辅助函数
	// 规则 1：普通字段
	// 字段不存在 → 返回 nil（不更新）
	// 字段为 null → 返回空字符串指针（清空）
	// 字段有值 → 返回值指针（更新）
	parseStringField := func(fieldName string) *string {
		val, ok := payload[fieldName]
		if !ok {
			return nil // 字段不存在，不更新
		}
		if val == nil {
			empty := ""
			return &empty // null，清空
		}
		if s, ok := val.(string); ok {
			return &s // 有值，更新
		}
		return nil // 类型不匹配，不更新
	}

	// 规则 2：数组字段
	// 字段不存在 → 返回 nil（不更新）
	// 字段为 null → 返回空数组（清空）
	// 字段有值 → 返回数组（更新）
	parseStringArrayField := func(fieldName string) []string {
		val, ok := payload[fieldName]
		if !ok {
			return nil // 字段不存在，不更新
		}
		if val == nil {
			return []string{} // null，清空（空数组）
		}
		if arr, ok := val.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					// 允许空字符串，因为 '*' 是有效值
					result = append(result, s)
				}
			}
			return result // 有值，更新（可能是 ['*'] 或其他值）
		}
		return nil // 类型不匹配，不更新
	}

	// 使用统一的解析函数处理所有字段
	// 普通字段
	req.Nickname = parseStringField("nickname")
	req.Role = parseStringField("role")
	req.Status = parseStringField("status")
	req.AlarmScope = parseStringField("alarm_scope")

	// 有 Hash 的字段
	req.Email = parseStringField("email")
	req.EmailHash = parseStringField("email_hash")
	req.Phone = parseStringField("phone")
	req.PhoneHash = parseStringField("phone_hash")

	// 数组字段
	req.AlarmLevels = parseStringArrayField("alarm_levels")
	req.AlarmChannels = parseStringArrayField("alarm_channels")
	req.Tags = parseStringArrayField("tags")

	// Branch: 支持 branch_ids（数组），向后兼容单个 branch_id
	// 注意：primary_branch_id 已移除，Service 层会使用第一个 branch 用于向后兼容
	// 使用统一的 parseStringArrayField 函数处理，以区分"字段不存在"和"空数组"
	// 规则：字段不存在 → nil（不更新），字段为 null → []（清空），字段有值 → ['*'] 或 ['id1', ...]（更新）
	var branchIDs []string
	branchIDsParsed := parseStringArrayField("branch_ids")

	if branchIDsParsed != nil {
		// 字段存在（可能是空数组或有效值），使用解析后的值
		branchIDs = branchIDsParsed
	} else {
		// 字段不存在，尝试向后兼容：使用单个 branch_id
		if branchIDVal, ok := payload["branch_id"].(string); ok {
			if branchIDVal != "" {
				branchIDs = []string{branchIDVal}
			}
		}
	}

	// 如果解析到了 branch_ids（包括空数组），设置到请求中
	// nil 表示字段不存在（不更新），[] 表示清空，['*'] 或 ['id1', ...] 表示更新
	if branchIDs != nil {
		req.BranchIDs = branchIDs
	}

	resp, err := h.userService.UpdateUser(ctx, req)
	if err != nil {
		h.logger.Error("UpdateUser failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	if resp.Success {
		result := map[string]any{"success": true}
		// 如果是 Individual 用户被禁用，返回 Resident 更新状态（转院）
		if resp.ResidentTransferred || resp.ResidentError != "" {
			result["resident_transferred"] = resp.ResidentTransferred
			if resp.ResidentError != "" {
				result["resident_error"] = resp.ResidentError
			}
		}
		writeJSON(w, http.StatusOK, Ok(result))
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
		result := map[string]any{"success": true}
		// 如果是 Individual 用户，返回 Resident 更新状态
		if resp.ResidentDischarged || resp.ResidentError != "" {
			result["resident_discharged"] = resp.ResidentDischarged
			if resp.ResidentError != "" {
				result["resident_error"] = resp.ResidentError
			}
		}
		writeJSON(w, http.StatusOK, Ok(result))
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

	passwordHash, _ := payload["password_hash"].(string)
	if passwordHash == "" {
		writeJSON(w, http.StatusOK, Fail("password_hash is required"))
		return
	}

	req := service.UserResetPasswordRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
		NewPassword:   passwordHash,
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

// ============================================
// GetAccountSettings 获取账户设置
// ============================================

// GetAccountSettings 获取账户设置
func (h *UserHandler) GetAccountSettings(w http.ResponseWriter, r *http.Request, userID string) {
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

	req := service.GetAccountSettingsRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	resp, err := h.userService.GetAccountSettings(ctx, req)
	if err != nil {
		h.logger.Error("GetAccountSettings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	item := map[string]any{
		"id":         resp.ID,
		"account":    resp.Account,
		"nickname":   resp.Nickname,
		"role":       resp.Role,
		"save_email": resp.SaveEmail,
		"save_phone": resp.SavePhone,
	}
	if resp.Email != nil {
		item["email"] = *resp.Email
	}
	if resp.Phone != nil {
		item["phone"] = *resp.Phone
	}

	// Branch 相关字段
	if resp.BranchIDs != nil {
		item["branch_ids"] = resp.BranchIDs
	}
	if resp.AvailableBranches != nil && len(resp.AvailableBranches) > 0 {
		availableBranches := make([]map[string]string, 0, len(resp.AvailableBranches))
		for _, branch := range resp.AvailableBranches {
			availableBranches = append(availableBranches, map[string]string{
				"branch_id":   branch.BranchID,
				"branch_name": branch.BranchName,
			})
		}
		item["available_branches"] = availableBranches
	}

	writeJSON(w, http.StatusOK, Ok(item))
}

// ============================================
// UpdateAccountSettings 更新账户设置
// ============================================

// UpdateAccountSettings 更新账户设置（统一 API）
func (h *UserHandler) UpdateAccountSettings(w http.ResponseWriter, r *http.Request, userID string) {
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

	req := service.UpdateAccountSettingsRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: currentUserID,
	}

	// 解析 password_hash（按照统一设计规则）
	// 字段不存在 → req.PasswordHash = nil（不更新）
	// 字段为 null → req.PasswordHash = nil（不更新，密码字段不允许清空）
	// 字段有值 → req.PasswordHash = &hash（更新）
	if passwordHashVal, ok := payload["password_hash"]; ok {
		if passwordHashVal == nil {
			// null 表示不更新（密码字段不允许清空）
			// 不设置 req.PasswordHash
		} else if passwordHash, ok := passwordHashVal.(string); ok && passwordHash != "" {
			req.PasswordHash = &passwordHash
		}
	}

	// 解析 pin_hash（按照统一设计规则）
	// 字段不存在 → req.PinHash = nil（不更新）
	// 字段为 null → req.PinHash = nil（不更新）
	// 字段有值 → req.PinHash = &hash（更新）
	if pinHashVal, ok := payload["pin_hash"]; ok {
		if pinHashVal == nil {
			// null 表示不更新
			// 不设置 req.PinHash
		} else if pinHash, ok := pinHashVal.(string); ok {
			// 允许空字符串（表示删除 PIN）
			req.PinHash = &pinHash
		}
	}

	// 解析 email 和 email_hash
	// 处理 email 可能是 null 或空字符串的情况
	if emailVal, ok := payload["email"]; ok {
		if emailVal == nil {
			// null 转换为空字符串（表示删除明文但保留 hash）
			emptyStr := ""
			req.Email = &emptyStr
		} else if email, ok := emailVal.(string); ok {
			req.Email = &email
		}
	}
	// email_hash 处理：
	// 1. 如果字段存在且值为 null -> 删除 hash（设置为空字符串指针）
	// 2. 如果字段存在且值为字符串 -> 更新 hash
	// 3. 如果字段不存在（undefined）-> 不更新 hash（保持 nil）
	if emailHashVal, ok := payload["email_hash"]; ok {
		if emailHashVal == nil {
			// null 表示删除 hash（前端明确传递 null）
			emptyStr := ""
			req.EmailHash = &emptyStr
		} else if emailHash, ok := emailHashVal.(string); ok {
			req.EmailHash = &emailHash
		}
	}
	// 如果字段不存在（ok == false），req.EmailHash 保持为 nil，表示不更新 hash

	// 解析 phone 和 phone_hash
	// 处理 phone 可能是 null 或空字符串的情况
	if phoneVal, ok := payload["phone"]; ok {
		if phoneVal == nil {
			// null 转换为空字符串（表示删除明文但保留 hash）
			emptyStr := ""
			req.Phone = &emptyStr
		} else if phone, ok := phoneVal.(string); ok {
			req.Phone = &phone
		}
	}
	// phone_hash 处理：
	// 1. 如果字段存在且值为 null -> 删除 hash（设置为空字符串指针）
	// 2. 如果字段存在且值为字符串 -> 更新 hash
	// 3. 如果字段不存在（undefined）-> 不更新 hash（保持 nil）
	if phoneHashVal, ok := payload["phone_hash"]; ok {
		if phoneHashVal == nil {
			// null 表示删除 hash（前端明确传递 null）
			emptyStr := ""
			req.PhoneHash = &emptyStr
		} else if phoneHash, ok := phoneHashVal.(string); ok {
			req.PhoneHash = &phoneHash
		}
	}
	// 如果字段不存在（ok == false），req.PhoneHash 保持为 nil，表示不更新 hash

	// 检查是否有任何更新
	if req.PasswordHash == nil && req.PinHash == nil && req.Email == nil && req.Phone == nil {
		writeJSON(w, http.StatusOK, Fail("no fields to update"))
		return
	}

	resp, err := h.userService.UpdateAccountSettings(ctx, req)
	if err != nil {
		h.logger.Error("UpdateAccountSettings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
		"message": resp.Message,
	}))
}

// ============================================
// GetAvailableBranches 获取可用 branch 列表
// ============================================

// GetAvailableBranches 获取可用 branch 列表（用于前端创建用户时选择 branch）
// GET /admin/api/v1/users/branches
func (h *UserHandler) GetAvailableBranches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 获取 tenant_id
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 获取 current_user_id（可选，用于权限过滤）
	currentUserID := r.Header.Get("X-User-Id")

	// 3. 调用 Service
	req := service.GetAvailableBranchesRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,
	}

	resp, err := h.userService.GetAvailableBranches(ctx, req)
	if err != nil {
		h.logger.Error("GetAvailableBranches failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 转换为前端格式
	branches := make([]map[string]any, 0, len(resp.Branches))
	for _, branch := range resp.Branches {
		branches = append(branches, map[string]any{
			"branch_id":   branch.BranchID,
			"branch_name": branch.BranchName,
		})
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"branches": branches,
	}))
}

// GetAvailableCaregivers 获取可用 caregivers 列表
func (h *UserHandler) GetAvailableCaregivers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 获取 tenant_id
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 获取 current_user_id（必填，用于权限过滤）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("current_user_id is required"))
		return
	}

	// 3. 获取 branch_id（必填，从 query 参数获取）
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusOK, Fail("branch_id is required"))
		return
	}

	// 4. 获取 unit_id（可选，从 query 参数获取）
	unitID := r.URL.Query().Get("unit_id")

	// 5. 调用 Service
	req := service.GetAvailableCaregiversRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,
		BranchID:      branchID,
		UnitID:        unitID,
	}

	resp, err := h.userService.GetAvailableCaregivers(ctx, req)
	if err != nil {
		h.logger.Error("GetAvailableCaregivers failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 转换为前端格式
	items := make([]map[string]any, 0, len(resp.Items))
	for _, user := range resp.Items {
		item := map[string]any{
			"user_id":       user.UserID,
			"tenant_id":     user.TenantID,
			"user_account":  user.UserAccount,
			"user_nickname": user.Nickname, // 前端使用 user_nickname
			"nickname":      user.Nickname, // 向后兼容
			"email":         user.Email,
			"phone":         user.Phone,
			"role":          user.Role,
			"status":        user.Status,
			"tags":          user.Tags, // 返回 tags 数组
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
	}))
}

// GetAvailableCaregiverGroups 获取可用 caregiver groups 列表
func (h *UserHandler) GetAvailableCaregiverGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 获取 tenant_id
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 获取 current_user_id（必填，用于权限过滤）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("current_user_id is required"))
		return
	}

	// 3. 获取 branch_id（必填，从 query 参数获取）
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusOK, Fail("branch_id is required"))
		return
	}

	// 4. 调用 Service
	req := service.GetAvailableCaregiverGroupsRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,
		BranchID:      branchID,
	}

	resp, err := h.userService.GetAvailableCaregiverGroups(ctx, req)
	if err != nil {
		h.logger.Error("GetAvailableCaregiverGroups failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 4. 转换为前端格式
	items := make([]map[string]any, 0, len(resp.Items))
	for _, group := range resp.Items {
		// 转换 members 为前端格式
		members := make([]map[string]any, 0, len(group.Members))
		for _, member := range group.Members {
			memberMap := map[string]any{
				"user_id":       member.UserID,
				"user_account":  member.UserAccount,
				"user_nickname": member.Nickname,
				"role":          member.Role,
				"tags":          member.Tags,
			}
			members = append(members, memberMap)
		}
		item := map[string]any{
			"tag_name":     group.TagName,
			"member_count": group.MemberCount,
			"member_names": group.MemberNames, // 向后兼容：成员昵称列表
			"members":      members,           // 成员详细信息列表
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
	}))
}
