package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// UserService 用户管理服务接口
type UserService interface {
	// 查询
	ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
	GetUser(ctx context.Context, req GetUserRequest) (*GetUserResponse, error)

	// 创建
	CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)

	// 更新
	UpdateUser(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error)

	// 删除
	DeleteUser(ctx context.Context, req DeleteUserRequest) (*DeleteUserResponse, error)

	// 密码和 PIN 管理
	ResetPassword(ctx context.Context, req UserResetPasswordRequest) (*UserResetPasswordResponse, error)
	ResetPIN(ctx context.Context, req UserResetPINRequest) (*UserResetPINResponse, error)

	// 账户设置管理（统一 API）
	GetAccountSettings(ctx context.Context, req GetAccountSettingsRequest) (*GetAccountSettingsResponse, error)
	UpdateAccountSettings(ctx context.Context, req UpdateAccountSettingsRequest) (*UpdateAccountSettingsResponse, error)
}

// userService 实现
type userService struct {
	usersRepo repository.UsersRepository
	logger    *zap.Logger
}

// NewUserService 创建 UserService 实例
func NewUserService(usersRepo repository.UsersRepository, logger *zap.Logger) UserService {
	return &userService{
		usersRepo: usersRepo,
		logger:    logger,
	}
}

// ============================================
// Request/Response DTOs
// ============================================

// ListUsersRequest 查询用户列表请求
type ListUsersRequest struct {
	TenantID      string // 必填
	CurrentUserID string // 当前用户 ID（用于权限过滤）
	Search        string // 可选：搜索关键词（user_account, nickname, email, phone）
	Page          int    // 可选，默认 1
	Size          int    // 可选，默认 20
}

// ListUsersResponse 查询用户列表响应
type ListUsersResponse struct {
	Items []*UserDTO // 用户列表
	Total int        // 总数量
}

// GetUserRequest 查询用户详情请求
type GetUserRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
}

// GetUserResponse 查询用户详情响应
type GetUserResponse struct {
	User *UserDTO // 用户信息
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	TenantID      string   // 必填
	CurrentUserID string   // 当前用户 ID（用于权限检查）
	UserAccount   string   // 必填
	Password      string   // 必填
	Role          string   // 必填
	Nickname      string   // 可选
	Email         string   // 可选
	Phone         string   // 可选
	Status        string   // 可选，默认 "active"
	AlarmLevels   []string // 可选
	AlarmChannels []string // 可选
	AlarmScope    string   // 可选，根据角色设置默认值
	Tags          []string // 可选
	BranchTag     string   // 可选
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	UserID string // 新创建的用户 ID
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
	// 可选字段（nil 表示不更新，空字符串表示清空）
	Nickname      *string  // 可选
	Email         *string  // 可选（null 表示删除）
	EmailHash     *string  // 可选（前端计算的 hash）
	Phone         *string  // 可选（null 表示删除）
	PhoneHash     *string  // 可选（前端计算的 hash）
	Role          *string  // 可选
	Status        *string  // 可选
	AlarmLevels   []string // 可选（nil 表示不更新，空数组表示清空）
	AlarmChannels []string // 可选（nil 表示不更新，空数组表示清空）
	AlarmScope    *string  // 可选
	Tags          []string // 可选（nil 表示不更新，空数组表示清空）
	BranchTag     *string  // 可选（空字符串表示 NULL）
}

// UpdateUserResponse 更新用户响应
type UpdateUserResponse struct {
	Success bool // 是否成功
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
}

// DeleteUserResponse 删除用户响应
type DeleteUserResponse struct {
	Success bool // 是否成功
}

// UserResetPasswordRequest 重置用户密码请求
type UserResetPasswordRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
	NewPassword   string // 必填
}

// UserResetPasswordResponse 重置用户密码响应
type UserResetPasswordResponse struct {
	Success bool   // 是否成功
	Message string // 消息（可选）
}

// UserResetPINRequest 重置用户 PIN 请求
type UserResetPINRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
	NewPIN        string // 必填（必须是 4 位数字）
}

// UserResetPINResponse 重置用户 PIN 响应
type UserResetPINResponse struct {
	Success bool // 是否成功
}

// GetAccountSettingsRequest 获取账户设置请求
type GetAccountSettingsRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
}

// GetAccountSettingsResponse 获取账户设置响应
type GetAccountSettingsResponse struct {
	ID        string  // UUID: user_id（前端需要）
	Account   string  // user_account
	Nickname  string  // 昵称
	Email     *string // 邮箱（可选，nil 表示不存在）
	Phone     *string // 电话（可选，nil 表示不存在）
	Role      string  // 角色代码（前端需要，用于判断使用哪种表）
	SaveEmail bool    // 是否保存 email 明文（Staff 总是 true）
	SavePhone bool    // 是否保存 phone 明文（Staff 总是 true）
}

// UpdateAccountSettingsRequest 更新账户设置请求（统一 API，在同一个事务中处理所有更新）
type UpdateAccountSettingsRequest struct {
	TenantID      string  // 必填
	UserID        string  // 必填
	CurrentUserID string  // 当前用户 ID（用于权限检查）
	PasswordHash  *string // 可选：密码 hash（nil 表示不更新）
	Email         *string // 可选：邮箱（nil 表示不更新，空字符串表示删除）
	EmailHash     *string // 可选：邮箱 hash（前端计算的 hash）
	Phone         *string // 可选：电话（nil 表示不更新，空字符串表示删除）
	PhoneHash     *string // 可选：电话 hash（前端计算的 hash）
}

// UpdateAccountSettingsResponse 更新账户设置响应
type UpdateAccountSettingsResponse struct {
	Success bool   // 是否成功
	Message string // 消息（可选，用于错误详情）
}

// UserDTO 用户数据传输对象（用于响应）
type UserDTO struct {
	UserID        string                 `json:"user_id"`
	TenantID      string                 `json:"tenant_id"`
	UserAccount   string                 `json:"user_account"`
	Nickname      string                 `json:"nickname,omitempty"`
	Email         string                 `json:"email,omitempty"`
	Phone         string                 `json:"phone,omitempty"`
	Role          string                 `json:"role"`
	Status        string                 `json:"status"`
	AlarmLevels   []string               `json:"alarm_levels,omitempty"`
	AlarmChannels []string               `json:"alarm_channels,omitempty"`
	AlarmScope    string                 `json:"alarm_scope,omitempty"`
	BranchTag     string                 `json:"branch_tag,omitempty"`
	LastLoginAt   string                 `json:"last_login_at,omitempty"` // RFC3339 格式
	Tags          []string               `json:"tags,omitempty"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
}

// ============================================
// Helper Functions
// ============================================

// getRoleLevel 返回角色的层级（数字越小，权限越高）
func getRoleLevel(role string) int {
	switch strings.ToLower(role) {
	case "systemadmin", "systemoperator":
		return 1
	case "admin":
		return 2
	case "manager", "it":
		return 3
	case "nurse", "caregiver":
		return 4
	case "resident", "family":
		return 5
	default:
		return 999 // 未知角色，最严格
	}
}

// canCreateRole 检查当前用户是否可以创建指定角色
// 规则：可以创建同级或下级角色
func canCreateRole(currentRole, targetRole string) bool {
	// SystemAdmin 和 SystemOperator 只能由 SystemAdmin 创建（已有单独检查）
	if targetRole == "SystemAdmin" || targetRole == "SystemOperator" {
		return false // 这个检查在调用前已经单独处理
	}

	currentLevel := getRoleLevel(currentRole)
	targetLevel := getRoleLevel(targetRole)

	// 方案A：允许创建同级或下级角色
	return targetLevel >= currentLevel
}

// HashAccount 哈希账号（SHA256(lower(account))）
func HashAccount(account string) string {
	// 这个函数应该在 httpapi 包中，但为了 Service 层独立，我们在这里实现
	// 实际实现应该调用 httpapi.HashAccount
	// 暂时先实现一个简单版本
	normalized := strings.ToLower(strings.TrimSpace(account))
	hash := sha256Hex(normalized)
	return hash
}

// HashPassword 哈希密码（SHA256(password)，独立于 account）
func HashPassword(password string) string {
	// 这个函数应该在 httpapi 包中，但为了 Service 层独立，我们在这里实现
	// 实际实现应该调用 httpapi.HashPassword
	hash := sha256Hex(password)
	return hash
}

// sha256Hex 计算 SHA256 并返回 hex 字符串
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// domainUserToDTO 将 domain.User 转换为 UserDTO
func domainUserToDTO(user *domain.User) *UserDTO {
	dto := &UserDTO{
		UserID:      user.UserID,
		TenantID:    user.TenantID,
		UserAccount: user.UserAccount,
		Role:        user.Role,
		Status:      user.Status,
	}

	if user.Nickname.Valid {
		dto.Nickname = user.Nickname.String
	}
	if user.Email.Valid {
		dto.Email = user.Email.String
	}
	if user.Phone.Valid {
		dto.Phone = user.Phone.String
	}
	if len(user.AlarmLevels) > 0 {
		dto.AlarmLevels = []string(user.AlarmLevels)
	}
	if len(user.AlarmChannels) > 0 {
		dto.AlarmChannels = []string(user.AlarmChannels)
	}
	if user.AlarmScope.Valid {
		dto.AlarmScope = user.AlarmScope.String
	}
	if user.BranchTag.Valid {
		dto.BranchTag = user.BranchTag.String
	}
	if user.LastLoginAt.Valid {
		dto.LastLoginAt = user.LastLoginAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if user.Tags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(user.Tags.String), &tags); err == nil {
			dto.Tags = tags
		}
	}
	if user.Preferences.Valid {
		var prefs map[string]interface{}
		if err := json.Unmarshal([]byte(user.Preferences.String), &prefs); err == nil {
			dto.Preferences = prefs
		}
	}

	return dto
}

// ============================================
// Service 方法实现
// ============================================

// ListUsers 查询用户列表
func (s *userService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// 2. 获取当前用户信息（用于权限过滤）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 权限检查
	var permCheck *repository.PermissionCheck
	if currentUser.Role != "" {
		permCheck, err = s.usersRepo.GetResourcePermission(ctx, currentUser.Role, "users", "R")
		if err != nil {
			s.logger.Warn("Failed to check resource permission, using default", zap.Error(err))
			// 默认最严格权限
			permCheck = &repository.PermissionCheck{AssignedOnly: true, BranchOnly: true}
		}
	} else {
		// 如果没有角色，使用最严格权限
		permCheck = &repository.PermissionCheck{AssignedOnly: true, BranchOnly: true}
	}

	// 4. 构建过滤器
	filters := repository.UserFilters{
		Search: strings.TrimSpace(req.Search),
	}

	// 应用权限过滤
	if permCheck.AssignedOnly {
		// Caregiver/Nurse: 只能查看自己
		// 直接返回当前用户
		return &ListUsersResponse{
			Items: []*UserDTO{domainUserToDTO(currentUser)},
			Total: 1,
		}, nil
	} else if permCheck.BranchOnly {
		// Manager: 只能查看同 branch 的用户
		if currentUser.BranchTag.Valid && currentUser.BranchTag.String != "" {
			filters.BranchTag = currentUser.BranchTag.String
		} else {
			// 如果当前用户的 branch_tag 为 NULL，只能查看 branch_tag 为 NULL 或 '-' 的用户
			filters.BranchTagNull = true
		}
	}
	// Admin/IT: 无额外过滤（可以查看所有用户）

	// 5. 分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}

	// 6. 调用 Repository
	users, total, err := s.usersRepo.ListUsers(ctx, req.TenantID, filters, page, size)
	if err != nil {
		s.logger.Error("ListUsers failed", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 7. 转换为 DTO
	items := make([]*UserDTO, 0, len(users))
	for _, user := range users {
		items = append(items, domainUserToDTO(user))
	}

	return &ListUsersResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetUser 查询用户详情
func (s *userService) GetUser(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// 2. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 权限检查
	isViewingSelf := req.CurrentUserID == req.UserID
	if !isViewingSelf {
		// 获取目标用户信息
		targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// 检查是否可以查看目标用户（角色层级检查）
		if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
			return nil, fmt.Errorf("not allowed to view %s role user (current role: %s)", targetUser.Role, currentUser.Role)
		}
	}

	// 4. 查询用户详情
	user, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		s.logger.Error("GetUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &GetUserResponse{
		User: domainUserToDTO(user),
	}, nil
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if strings.TrimSpace(req.UserAccount) == "" {
		return nil, fmt.Errorf("user_account is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if strings.TrimSpace(req.Role) == "" {
		return nil, fmt.Errorf("role is required")
	}

	// 2. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	role := strings.TrimSpace(req.Role)

	// 3. 权限检查
	// 系统角色检查
	if role == "SystemAdmin" || role == "SystemOperator" {
		if req.TenantID != SystemTenantID || !strings.EqualFold(currentUser.Role, "SystemAdmin") {
			return nil, fmt.Errorf("not allowed to assign system role")
		}
	} else {
		// 角色层级检查
		if currentUser.Role != "" && !canCreateRole(currentUser.Role, role) {
			return nil, fmt.Errorf("not allowed to create %s role (current role: %s)", role, currentUser.Role)
		}
	}

	// 4. 数据准备
	userAccount := strings.ToLower(strings.TrimSpace(req.UserAccount))
	accountHash, err := hex.DecodeString(HashAccount(userAccount))
	if err != nil || len(accountHash) == 0 {
		return nil, fmt.Errorf("failed to hash account")
	}

	// 密码哈希（前端已 hash，这里直接解码 hex 字符串）
	// 前端发送的是 SHA256(password) 的 hex 字符串，直接解码为 byte slice
	passwordHash, err := hex.DecodeString(req.Password)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to decode password hash: %w", err)
	}

	// Email 和 Phone 哈希
	var emailHash, phoneHash []byte
	if req.Email != "" {
		emailHash, _ = hex.DecodeString(HashAccount(req.Email))
	}
	if req.Phone != "" {
		phoneHash, _ = hex.DecodeString(HashAccount(req.Phone))
	}

	// Status 默认值
	status := req.Status
	if status == "" {
		status = "active"
	}

	// AlarmScope 默认值（根据角色）
	var alarmScope string
	if req.AlarmScope != "" {
		alarmScope = req.AlarmScope
	} else {
		roleLower := strings.ToLower(role)
		if roleLower == "caregiver" || roleLower == "nurse" {
			alarmScope = "ASSIGNED_ONLY"
		} else if roleLower == "manager" {
			alarmScope = "BRANCH"
		}
		// 其他角色：留空（NULL）
	}

	// Tags 转换为 JSONB
	var tagsJSON []byte
	if len(req.Tags) > 0 {
		tagsJSON, _ = json.Marshal(req.Tags)
	}

	// 5. 唯一性检查
	if req.Email != "" {
		if err := s.usersRepo.CheckEmailUniqueness(ctx, req.TenantID, req.Email, ""); err != nil {
			return nil, err
		}
	}
	if req.Phone != "" {
		if err := s.usersRepo.CheckPhoneUniqueness(ctx, req.TenantID, req.Phone, ""); err != nil {
			return nil, err
		}
	}

	// 6. 构建 domain.User
	user := &domain.User{
		TenantID:        req.TenantID,
		UserAccount:     userAccount,
		UserAccountHash: accountHash,
		PasswordHash:    passwordHash,
		Role:            role,
		Status:          status,
	}

	if req.Nickname != "" {
		user.Nickname = sql.NullString{String: req.Nickname, Valid: true}
	}
	if req.Email != "" {
		user.Email = sql.NullString{String: req.Email, Valid: true}
		user.EmailHash = emailHash
	}
	if req.Phone != "" {
		user.Phone = sql.NullString{String: req.Phone, Valid: true}
		user.PhoneHash = phoneHash
	}
	if alarmScope != "" {
		user.AlarmScope = sql.NullString{String: alarmScope, Valid: true}
	}
	if len(req.AlarmLevels) > 0 {
		user.AlarmLevels = req.AlarmLevels
	}
	if len(req.AlarmChannels) > 0 {
		user.AlarmChannels = req.AlarmChannels
	}
	if req.BranchTag != "" {
		user.BranchTag = sql.NullString{String: req.BranchTag, Valid: true}
	}
	if len(tagsJSON) > 0 {
		user.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
	}

	// 7. 创建用户
	userID, err := s.usersRepo.CreateUser(ctx, req.TenantID, user)
	if err != nil {
		s.logger.Error("CreateUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 8. 同步标签到目录
	if len(req.Tags) > 0 {
		if err := s.usersRepo.SyncUserTagsToCatalog(ctx, req.TenantID, userID, req.Tags); err != nil {
			s.logger.Warn("Failed to sync tags to catalog", zap.Error(err))
			// 不返回错误，标签同步失败不影响用户创建
		}
	}

	return &CreateUserResponse{
		UserID: userID,
	}, nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// 2. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 获取目标用户信息
	targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 4. 权限检查
	isUpdatingSelf := req.CurrentUserID == req.UserID
	updatingRole := req.Role != nil && *req.Role != ""
	updatingStatus := req.Status != nil && *req.Status != ""
	updatingOtherFields := req.Nickname != nil || req.Email != nil || req.Phone != nil ||
		req.AlarmLevels != nil || req.AlarmChannels != nil ||
		req.AlarmScope != nil || req.Tags != nil || req.BranchTag != nil

	// 权限规则：如果更新自己且只更新 password/email/phone，无限制
	// 如果更新其他用户或更新 role/status/otherFields，需要权限检查
	if !isUpdatingSelf || updatingRole || updatingStatus || updatingOtherFields {
		// 角色更新检查
		if updatingRole {
			role := strings.TrimSpace(*req.Role)
			// 系统角色检查
			if role == "SystemAdmin" || role == "SystemOperator" {
				if req.TenantID != SystemTenantID || !strings.EqualFold(currentUser.Role, "SystemAdmin") {
					return nil, fmt.Errorf("not allowed to assign system role")
				}
			} else {
				// 角色层级检查
				if currentUser.Role != "" && !canCreateRole(currentUser.Role, role) {
					return nil, fmt.Errorf("not allowed to assign %s role (current role: %s)", role, currentUser.Role)
				}
			}
		}

		// 管理权限检查
		if !isUpdatingSelf || updatingStatus || updatingOtherFields {
			if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
				return nil, fmt.Errorf("not allowed to update %s role user (current role: %s)", targetUser.Role, currentUser.Role)
			}
		}
	}

	// 5. Status 验证
	if updatingStatus {
		status := strings.TrimSpace(*req.Status)
		if status != "" && status != "active" && status != "disabled" && status != "left" {
			return nil, fmt.Errorf("invalid status")
		}
	}

	// 6. 构建更新数据（只更新提供的字段）
	// 注意：UpdateUser 需要完整的 domain.User，但只更新非零字段
	// 这里我们需要先获取现有用户，然后只更新提供的字段
	updateUser := *targetUser // 复制现有用户

	// 更新字段
	if req.Nickname != nil {
		if *req.Nickname == "" {
			updateUser.Nickname = sql.NullString{Valid: false}
		} else {
			updateUser.Nickname = sql.NullString{String: *req.Nickname, Valid: true}
		}
	}

	// Email 和 EmailHash 的业务逻辑处理（Service 层负责）
	// 规则（与旧 Handler 一致）：
	// 1. 如果 Email 提供：
	//    - Email 为 null/空：删除 email（设置为 NULL），同时删除 hash
	//    - Email 有值：保存 email，计算 hash（如果 EmailHash 提供，使用提供的；否则计算）
	// 2. 如果 EmailHash 单独提供（Email 未提供）：
	//    - 只更新 hash，不更新 email
	// 3. 如果两者都不提供：不更新（保持原值）
	if req.Email != nil {
		if *req.Email == "" {
			// 删除 email（设置为 NULL）
			updateUser.Email = sql.NullString{Valid: false}
			// 删除 hash（设置为空 slice，Repository 不会更新）
			updateUser.EmailHash = nil
		} else {
			// 保存 email
			updateUser.Email = sql.NullString{String: *req.Email, Valid: true}
			// 计算或使用提供的 hash
			if req.EmailHash != nil && *req.EmailHash != "" {
				emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
				if len(emailHashBytes) > 0 {
					updateUser.EmailHash = emailHashBytes
				}
			} else {
				// 计算 hash
				emailHash, _ := hex.DecodeString(HashAccount(*req.Email))
				updateUser.EmailHash = emailHash
			}
		}
	} else if req.EmailHash != nil {
		// 只更新 hash，不更新 email
		if *req.EmailHash != "" {
			emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
			if len(emailHashBytes) > 0 {
				updateUser.EmailHash = emailHashBytes
			}
		}
		// 如果 EmailHash 为空字符串，表示删除 hash，但 Repository 层不支持单独删除 hash
		// 这种情况应该通过更新 email 来触发，所以这里不做处理
	}
	// 如果两者都不提供：保持原值不变（不设置 updateUser.Email 和 updateUser.EmailHash）

	// Phone 和 PhoneHash 的业务逻辑处理（同 Email）
	if req.Phone != nil {
		if *req.Phone == "" {
			updateUser.Phone = sql.NullString{Valid: false}
			updateUser.PhoneHash = nil
		} else {
			updateUser.Phone = sql.NullString{String: *req.Phone, Valid: true}
			if req.PhoneHash != nil && *req.PhoneHash != "" {
				phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
				if len(phoneHashBytes) > 0 {
					updateUser.PhoneHash = phoneHashBytes
				}
			} else {
				phoneHash, _ := hex.DecodeString(HashAccount(*req.Phone))
				updateUser.PhoneHash = phoneHash
			}
		}
	} else if req.PhoneHash != nil {
		if *req.PhoneHash != "" {
			phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
			if len(phoneHashBytes) > 0 {
				updateUser.PhoneHash = phoneHashBytes
			}
		}
	}

	if updatingRole {
		updateUser.Role = strings.TrimSpace(*req.Role)
	}
	if updatingStatus {
		updateUser.Status = strings.TrimSpace(*req.Status)
	}
	if req.AlarmLevels != nil {
		updateUser.AlarmLevels = req.AlarmLevels
	}
	if req.AlarmChannels != nil {
		updateUser.AlarmChannels = req.AlarmChannels
	}
	if req.AlarmScope != nil {
		if *req.AlarmScope == "" {
			updateUser.AlarmScope = sql.NullString{Valid: false}
		} else {
			updateUser.AlarmScope = sql.NullString{String: *req.AlarmScope, Valid: true}
		}
	}
	if req.BranchTag != nil {
		if *req.BranchTag == "" {
			updateUser.BranchTag = sql.NullString{Valid: false}
		} else {
			updateUser.BranchTag = sql.NullString{String: *req.BranchTag, Valid: true}
		}
	}
	if req.Tags != nil {
		if len(req.Tags) == 0 {
			updateUser.Tags = sql.NullString{String: "[]", Valid: true}
		} else {
			tagsJSON, _ := json.Marshal(req.Tags)
			updateUser.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
		}
	}

	// 7. 唯一性检查
	if req.Email != nil && *req.Email != "" {
		if err := s.usersRepo.CheckEmailUniqueness(ctx, req.TenantID, *req.Email, req.UserID); err != nil {
			return nil, err
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if err := s.usersRepo.CheckPhoneUniqueness(ctx, req.TenantID, *req.Phone, req.UserID); err != nil {
			return nil, err
		}
	}

	// 8. 更新用户
	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("UpdateUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 9. 同步标签到目录（如果 tags 更新）
	if req.Tags != nil {
		if err := s.usersRepo.SyncUserTagsToCatalog(ctx, req.TenantID, req.UserID, req.Tags); err != nil {
			s.logger.Warn("Failed to sync tags to catalog", zap.Error(err))
		}
	}

	return &UpdateUserResponse{
		Success: true,
	}, nil
}

// DeleteUser 删除用户（软删除）
func (s *userService) DeleteUser(ctx context.Context, req DeleteUserRequest) (*DeleteUserResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// 2. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 获取目标用户信息
	targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 4. 权限检查
	if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
		return nil, fmt.Errorf("not allowed to delete %s role (current role: %s)", targetUser.Role, currentUser.Role)
	}

	// 5. 软删除（设置 status = 'left'）
	updateUser := *targetUser
	updateUser.Status = "left"

	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("DeleteUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &DeleteUserResponse{
		Success: true,
	}, nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, req UserResetPasswordRequest) (*UserResetPasswordResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.NewPassword == "" {
		return nil, fmt.Errorf("new_password is required")
	}

	// 2. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 权限检查
	isResettingSelf := req.CurrentUserID == req.UserID
	if !isResettingSelf {
		// 获取目标用户信息
		targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// 检查是否可以重置目标用户的密码
		if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
			return nil, fmt.Errorf("not allowed to reset password for %s role user (current role: %s)", targetUser.Role, currentUser.Role)
		}
	}

	// 4. 获取目标用户信息
	targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 5. 密码哈希（前端已 hash，这里直接解码 hex 字符串）
	// 前端发送的是 SHA256(password) 的 hex 字符串，直接解码为 byte slice
	passwordHash, err := hex.DecodeString(req.NewPassword)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to decode password hash: %w", err)
	}

	// 6. 更新密码
	updateUser := *targetUser
	updateUser.PasswordHash = passwordHash

	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("ResetPassword failed", zap.Error(err))
		return nil, fmt.Errorf("failed to reset password: %w", err)
	}

	return &UserResetPasswordResponse{
		Success: true,
		Message: "ok",
	}, nil
}

// ResetPIN 重置 PIN
func (s *userService) ResetPIN(ctx context.Context, req UserResetPINRequest) (*UserResetPINResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.NewPIN == "" {
		return nil, fmt.Errorf("new_pin is required")
	}

	// 2. PIN 格式验证（必须是 4 位数字）
	if len(req.NewPIN) != 4 {
		return nil, fmt.Errorf("PIN must be exactly 4 digits")
	}
	for _, c := range req.NewPIN {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("PIN must contain only digits")
		}
	}

	// 3. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 4. 权限检查
	isResettingSelf := req.CurrentUserID == req.UserID
	if !isResettingSelf {
		// 获取目标用户信息
		targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// 检查是否可以重置目标用户的 PIN
		if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
			return nil, fmt.Errorf("not allowed to reset PIN for %s role user (current role: %s)", targetUser.Role, currentUser.Role)
		}
	}

	// 5. 获取目标用户信息
	targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 6. PIN 哈希
	pinHash, err := hex.DecodeString(HashPassword(req.NewPIN))
	if err != nil || len(pinHash) == 0 {
		return nil, fmt.Errorf("failed to hash PIN")
	}

	// 7. 更新 PIN
	updateUser := *targetUser
	updateUser.PinHash = pinHash

	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("ResetPIN failed", zap.Error(err))
		return nil, fmt.Errorf("failed to reset PIN: %w", err)
	}

	return &UserResetPINResponse{
		Success: true,
	}, nil
}

// ============================================
// GetAccountSettings 获取账户设置
// ============================================

// GetAccountSettings 获取账户设置（只返回账户设置相关字段）
// 注意：这个 API 只能查看自己的账户设置，不允许查看其他用户的
func (s *userService) GetAccountSettings(ctx context.Context, req GetAccountSettingsRequest) (*GetAccountSettingsResponse, error) {
	// 1. 权限检查：只能查看自己的账户设置
	if req.CurrentUserID != req.UserID {
		return nil, fmt.Errorf("permission denied: can only view own account settings")
	}

	// 2. 获取用户信息
	user, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 3. 构建响应（只返回账户设置相关字段）
	resp := &GetAccountSettingsResponse{
		ID:        user.UserID,
		Account:   user.UserAccount,
		Nickname:  "",
		Role:      user.Role,
		SaveEmail: true, // Staff 总是保存
		SavePhone: true, // Staff 总是保存
	}
	if user.Nickname.Valid {
		resp.Nickname = user.Nickname.String
	}

	// Email 处理：如果存在返回值，否则返回 nil，但 save_email 总是 true
	if user.Email.Valid && user.Email.String != "" {
		email := user.Email.String
		resp.Email = &email
		resp.SaveEmail = true // 存在明文，已保存
	} else {
		resp.Email = nil
		resp.SaveEmail = true // Staff 总是保存，即使当前不存在，将来添加时也会保存
	}

	// Phone 处理：如果存在返回值，否则返回 nil，但 save_phone 总是 true
	if user.Phone.Valid && user.Phone.String != "" {
		phone := user.Phone.String
		resp.Phone = &phone
		resp.SavePhone = true // 存在明文，已保存
	} else {
		resp.Phone = nil
		resp.SavePhone = true // Staff 总是保存，即使当前不存在，将来添加时也会保存
	}

	return resp, nil
}

// ============================================
// UpdateAccountSettings 更新账户设置（统一 API）
// ============================================

// UpdateAccountSettings 更新账户设置（在同一个事务中处理所有更新）
// 注意：这个 API 只能更新自己的账户设置，不允许更新其他用户的
func (s *userService) UpdateAccountSettings(ctx context.Context, req UpdateAccountSettingsRequest) (*UpdateAccountSettingsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.UserID == "" || req.CurrentUserID == "" {
		return nil, fmt.Errorf("tenant_id, user_id, and current_user_id are required")
	}

	// 2. 权限检查：只能更新自己的账户设置
	if req.CurrentUserID != req.UserID {
		return nil, fmt.Errorf("permission denied: can only update own account settings")
	}

	// 3. 获取目标用户信息
	targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 4. 构建更新对象（只更新提供的字段）
	updateUser := domain.User{
		UserID:   targetUser.UserID,
		TenantID: targetUser.TenantID,
	}

	// 4.1 更新密码（如果提供，!= nil 就更新，不进行任何判断）
	if req.PasswordHash != nil {
		passwordHashBytes, _ := hex.DecodeString(*req.PasswordHash)
		updateUser.PasswordHash = passwordHashBytes
	}

	// 4.2 更新 email 字段（如果提供，!= nil 就更新，不进行任何判断）
	if req.Email != nil {
		updateUser.Email = sql.NullString{String: *req.Email, Valid: true}
	}

	// 4.2.1 更新 email_hash 字段（如果提供，!= nil 就更新，不进行任何判断）
	if req.EmailHash != nil {
		if *req.EmailHash == "" {
			// 空字符串，删除 email_hash 字段（设置为 NULL）
			updateUser.EmailHash = []byte{}
		} else {
			// 解码 hex 字符串
			emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
			updateUser.EmailHash = emailHashBytes
		}
	}

	// 4.3 更新 phone 字段（如果提供，!= nil 就更新，不进行任何判断）
	if req.Phone != nil {
		updateUser.Phone = sql.NullString{String: *req.Phone, Valid: true}
	}

	// 4.3.1 更新 phone_hash 字段（如果提供，!= nil 就更新，不进行任何判断）
	if req.PhoneHash != nil {
		if *req.PhoneHash == "" {
			// 空字符串，删除 phone_hash 字段（设置为 NULL）
			updateUser.PhoneHash = []byte{}
		} else {
			// 解码 hex 字符串
			phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
			updateUser.PhoneHash = phoneHashBytes
		}
	}

	// 6. 执行更新（Repository 层会在事务中处理）
	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("UpdateAccountSettings failed", zap.Error(err))
		return nil, fmt.Errorf("failed to update account settings: %w", err)
	}

	return &UpdateAccountSettingsResponse{
		Success: true,
		Message: "Account settings updated successfully",
	}, nil
}
