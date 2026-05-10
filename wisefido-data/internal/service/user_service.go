package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

	// Branch 管理（用于前端创建用户时选择 branch）
	GetAvailableBranches(ctx context.Context, req GetAvailableBranchesRequest) (*GetAvailableBranchesResponse, error)

	// Caregivers 管理（用于前端选择 caregivers 和 caregiver groups）
	GetAvailableCaregivers(ctx context.Context, req GetAvailableCaregiversRequest) (*GetAvailableCaregiversResponse, error)
	GetAvailableCaregiverGroups(ctx context.Context, req GetAvailableCaregiverGroupsRequest) (*GetAvailableCaregiverGroupsResponse, error)
}

// userService 实现
type userService struct {
	usersRepo        repository.UsersRepository
	branchesRepo     repository.BranchesRepository     // 用于获取可用 branch 列表
	userBranchesRepo repository.UserBranchesRepository // 用于管理 user_branches 表
	db               *sql.DB                           // 用于复杂查询（JOIN、权限过滤）
	logger           *zap.Logger
}

// NewUserService 创建 UserService 实例
func NewUserService(usersRepo repository.UsersRepository, branchesRepo repository.BranchesRepository, userBranchesRepo repository.UserBranchesRepository, db *sql.DB, logger *zap.Logger) UserService {
	return &userService{
		usersRepo:        usersRepo,
		branchesRepo:     branchesRepo,
		userBranchesRepo: userBranchesRepo,
		db:               db,
		logger:           logger,
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
	TenantID      string   // 必填
	UserID        string   // 可选：Edit 模式必填，Create 模式为空或 "new"
	CurrentUserID string   // 当前用户 ID（用于权限检查）
	BranchIDs     []string // 可选：Create 模式时，如果已指定 branch，传入 branch_ids
}

// GetUserResponse 查询用户详情响应
type GetUserResponse struct {
	User              *UserDTO    // 用户信息（Create 模式下为 nil）
	AvailableTags     []string    // 当前用户所在 Branch 中存在的 tags（用于前端显示和选择）
	AvailableBranches []BranchDTO // 可用的 branch 列表（Create 模式下返回，Edit 模式下为空）
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
	BranchIDs     []string // 可选：通过 branch_id 列表在 user_branches 表中创建多个院区关联

	// 注意：AvailableBranches 不应由 Handler 传递，Service 层会自己从数据库查询用户的 branch 信息
	// 这是用户本身的属性，不能信任前端传递的值
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	UserID          string // 新创建的用户 ID
	ResidentCreated bool   // Individual 用户关联的 Resident 是否创建成功
	ResidentError   string // 如果 Resident 创建失败，错误信息（可选）
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
	BranchIDs     []string // 可选：通过 branch_id 列表在 user_branches 表中更新多个院区关联（nil 表示不更新，空数组表示删除所有关联）

	// 注意：AvailableBranches 不应由 Handler 传递，Service 层会自己从数据库查询用户的 branch 信息
	// 这是用户本身的属性，不能信任前端传递的值
}

// UpdateUserResponse 更新用户响应
type UpdateUserResponse struct {
	Success             bool   // 是否成功
	ResidentTransferred bool   // Individual 用户关联的 Resident 是否已设置为 transferred（转院）
	ResidentError       string // 如果 Resident 更新失败，错误信息（可选）
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	TenantID      string // 必填
	UserID        string // 必填
	CurrentUserID string // 当前用户 ID（用于权限检查）
}

// DeleteUserResponse 删除用户响应
type DeleteUserResponse struct {
	Success            bool   // 是否成功
	ResidentDischarged bool   // Individual 用户关联的 Resident 是否已设置为 discharged
	ResidentError      string // 如果 Resident 更新失败，错误信息（可选）
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
	ID                string      // UUID: user_id（前端需要）
	Account           string      // user_account
	Nickname          string      // 昵称
	Email             *string     // 邮箱（可选，nil 表示不存在）
	Phone             *string     // 电话（可选，nil 表示不存在）
	Role              string      // 角色代码（前端需要，用于判断使用哪种表）
	SaveEmail         bool        // 是否保存 email 明文（Staff 总是 true）
	SavePhone         bool        // 是否保存 phone 明文（Staff 总是 true）
	BranchIDs         []string    // 用户有权限的所有 branch_id 列表（Admin 可能是 ["*", "branch-id-1", ...]）
	AvailableBranches []BranchDTO // 用户有权限的所有 branch 列表（包含 branch_id 和 branch_name，用于前端下拉选择）
}

// UpdateAccountSettingsRequest 更新账户设置请求（统一 API，在同一个事务中处理所有更新）
type UpdateAccountSettingsRequest struct {
	TenantID      string  // 必填
	UserID        string  // 必填
	CurrentUserID string  // 当前用户 ID（用于权限检查）
	PasswordHash  *string // 可选：密码 hash（nil 表示不更新）
	PinHash       *string // 可选：PIN hash（nil 表示不更新）
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

// GetAvailableBranchesRequest 获取可用 branch 列表请求
type GetAvailableBranchesRequest struct {
	TenantID      string // 必填
	CurrentUserID string // 当前用户 ID（用于权限过滤，可选）
}

// GetAvailableBranchesResponse 获取可用 branch 列表响应
type GetAvailableBranchesResponse struct {
	Branches []BranchDTO // 可用 branch 列表（包含 branch_id 和 branch_name）
}

// BranchDTO branch 数据传输对象（用于响应）
type BranchDTO struct {
	BranchID   string `json:"branch_id"`   // branch_id（前端需要 ID 来选择对象）
	BranchName string `json:"branch_name"` // branch_name（用于显示）
}

// GetAvailableCaregiversRequest 获取可用 caregivers 请求
type GetAvailableCaregiversRequest struct {
	TenantID      string // 必填
	CurrentUserID string // 当前用户 ID（用于权限过滤，必填）
	BranchID      string // 必填：指定 branch_id，只返回该 branch 内的 caregivers/nurse
	UnitID        string // 可选：指定 unit_id，如果 unit_type 是 'Home'，允许选择 Individual 类型
}

// GetAvailableCaregiversResponse 获取可用 caregivers 响应
type GetAvailableCaregiversResponse struct {
	Items []UserDTO // 可用 caregivers 列表（role='Nurse' or 'Caregiver' and status='active'，且在指定的 branch 内）
}

// GetAvailableCaregiverGroupsRequest 获取可用 caregiver groups 请求
type GetAvailableCaregiverGroupsRequest struct {
	TenantID      string // 必填
	CurrentUserID string // 当前用户 ID（用于权限过滤，必填）
	BranchID      string // 必填：指定 branch_id，只返回该 branch 内的 caregiver groups
}

// GetAvailableCaregiverGroupsResponse 获取可用 caregiver groups 响应
type GetAvailableCaregiverGroupsResponse struct {
	Items []CaregiverGroupDTO // 可用 caregiver groups 列表（tag 名称和成员数量）
}

// CaregiverGroupDTO caregiver group 数据传输对象
type CaregiverGroupDTO struct {
	TagName     string    `json:"tag_name"`     // 标签名称
	MemberCount int       `json:"member_count"` // 成员数量（该 tag 下有多少个 active 的 caregiver/nurse）
	MemberNames []string  `json:"member_names"` // 成员昵称列表（用于前端显示，向后兼容）
	Members     []UserDTO `json:"members"`      // 成员详细信息列表（user_id, user_account, user_nickname, role, tags）
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
	BranchIDs     []string               `json:"branch_ids,omitempty"`    // 返回所有 branch_id 列表
	BranchID      string                 `json:"branch_id,omitempty"`     // 向后兼容：第一个 branch ID（按 branch_name 排序）
	BranchName    string                 `json:"branch_name,omitempty"`   // 向后兼容：第一个 branch 名称（用于显示）
	LastLoginAt   string                 `json:"last_login_at,omitempty"` // RFC3339 格式
	Tags          []string               `json:"tags,omitempty"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
}

// ============================================
// Helper Functions
// ============================================

// normalizeBranchName 规范化 branch_name：将 ""、"-" 视为空院区（返回 sql.NullString{Valid: false}）
// null、""、"-" 都表示空院区，等价处理
func normalizeBranchName(branchName string) sql.NullString {
	if branchName == "" || branchName == "-" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: branchName, Valid: true}
}

// UserBranchInfo 用户院区信息（包含 branch_id 和 branch_name）
type UserBranchInfo struct {
	BranchID   string // 院区 ID
	BranchName string // 院区名称（用于显示）
}

// createUserBranches 创建用户与多个院区的关联（Service 层内部方法）
// 使用 Repository 层方法，而不是直接操作数据库
// 注意：此方法不应接收包含 '*' 的 branchIDs 数组，因为 '*' 表示所有分支（user_branches 表为空）
// 但是为了防御性编程，这里会过滤掉 '*' 字符
func (s *userService) createUserBranches(ctx context.Context, tenantID, userID string, branchIDs []string) error {
	if len(branchIDs) == 0 {
		return nil // 如果没有 branch，不创建关联
	}

	// 1. 先删除用户的所有现有关联（如果存在，使用 Repository 方法）
	// 注意：CreateUser 时用户可能还没有关联，所以这里删除可能没有记录，但不影响
	if err := s.userBranchesRepo.DeleteAllUserBranches(ctx, tenantID, userID); err != nil {
		// 如果删除失败（可能是用户还没有关联），继续执行创建
		s.logger.Warn("Failed to delete existing user branches (may not exist)", zap.Error(err))
	}

	// 2. 过滤掉 '*' 字符（防御性编程）
	// 注意：理论上不应收到 '*'，因为调用方已经在 CreateUser 中过滤了
	// 但是如果收到了，这里会跳过，避免尝试将 '*' 作为 UUID 插入数据库
	validBranchIDs := make([]string, 0, len(branchIDs))
	for _, branchID := range branchIDs {
		if branchID != "" && branchID != "*" {
			validBranchIDs = append(validBranchIDs, branchID)
		}
	}

	if len(validBranchIDs) == 0 {
		return nil // 过滤后没有有效的 branch_id，不创建关联
	}

	// 3. 创建新的关联（使用 Repository 方法）
	for _, branchID := range validBranchIDs {
		userBranch := &domain.UserBranch{
			TenantID: tenantID,
			UserID:   userID,
			BranchID: branchID,
		}
		_, err := s.userBranchesRepo.CreateUserBranch(ctx, tenantID, userBranch)
		if err != nil {
			// CreateUserBranch 已经处理了唯一性约束（如果已存在，返回现有ID，不报错）
			// 所以这里只需要检查其他错误
			return fmt.Errorf("failed to create user branch association for branch_id '%s': %w", branchID, err)
		}
	}

	return nil
}

// updateUserBranches 更新用户与多个院区的关联（Service 层内部方法）
// 使用 Repository 层方法，而不是直接操作数据库
// 注意：此方法不应接收包含 '*' 的 branchIDs 数组，因为 '*' 表示所有分支（user_branches 表为空）
// 但是为了防御性编程，这里会过滤掉 '*' 字符
func (s *userService) updateUserBranches(ctx context.Context, tenantID, userID string, branchIDs []string) error {
	// 1. 先删除用户的所有现有关联（使用 Repository 方法）
	if err := s.userBranchesRepo.DeleteAllUserBranches(ctx, tenantID, userID); err != nil {
		return fmt.Errorf("failed to delete existing user branches: %w", err)
	}

	// 2. 如果有新的 branchIDs，过滤掉 '*' 字符后创建新的关联（使用 Repository 方法）
	// 注意：如果 branchIDs 为空或只包含 '*'，则删除所有关联后不创建新关联（表示 *(ALL) 或未分配）
	if len(branchIDs) > 0 {
		// 过滤掉 '*' 字符（防御性编程）
		// 注意：理论上不应收到 '*'，因为调用方已经在 UpdateUser 中过滤了
		// 但是如果收到了，这里会跳过，避免尝试将 '*' 作为 UUID 插入数据库
		validBranchIDs := make([]string, 0, len(branchIDs))
		for _, branchID := range branchIDs {
			if branchID != "" && branchID != "*" {
				validBranchIDs = append(validBranchIDs, branchID)
			}
		}

		// 3. 创建新的关联（使用 Repository 方法）
		for _, branchID := range validBranchIDs {
			userBranch := &domain.UserBranch{
				TenantID: tenantID,
				UserID:   userID,
				BranchID: branchID,
			}
			_, err := s.userBranchesRepo.CreateUserBranch(ctx, tenantID, userBranch)
			if err != nil {
				// CreateUserBranch 已经处理了唯一性约束（如果已存在，返回现有ID，不报错）
				// 所以这里只需要检查其他错误
				return fmt.Errorf("failed to create user branch association for branch_id '%s': %w", branchID, err)
			}
		}
	}

	return nil
}

// getUserBranchIDs 查询用户所属的院区信息（Service 层内部方法）
//
// v2 改造说明：owl_v2 schema 删除了 user_branches 表；用户的 branch 关系由 user_roles +
// roles.tenant_prefix + role_permissions.resource_scope INET 派生（用 utils/spatial 在调用侧推），
// 不再有"用户属于哪几个 branch"的关系列表。本方法直接返回 (nil, false, nil) — service 层会按
// "无 branch 限制 = 可访问全部"处理（即 admin/sysadmin 行为），与 v2 RBAC 默认权限一致。
func (s *userService) getUserBranchIDs(ctx context.Context, tenantID, userID string) (branches []UserBranchInfo, hasBranches bool, err error) {
	return nil, false, nil
}

// 旧 v1 实现（依赖 user_branches 表，v2 已删）保留以下死代码以备 grep；活路径走上面 stub。
// nolint:unused
func (s *userService) getUserBranchIDs_v1Legacy(ctx context.Context, tenantID, userID string) (branches []UserBranchInfo, hasBranches bool, err error) {
	if tenantID == "" || userID == "" {
		return nil, false, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT
			ub.branch_id::text,
			COALESCE(b.branch_name, '') as branch_name
		 FROM user_branches ub
		 LEFT JOIN branches b ON b.branch_id = ub.branch_id
		 WHERE ub.tenant_id = $1 AND ub.user_id::text = $2
		 ORDER BY b.branch_name ASC`,
		tenantID, userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户没有关联任何院区，检查是否是 Admin 的 *(ALL) 情况
			return s.handleAdminAllBranches(ctx, tenantID, userID)
		}
		return nil, false, fmt.Errorf("failed to query user branches: %w", err)
	}
	defer rows.Close()

	var branchList []UserBranchInfo
	for rows.Next() {
		var branchID, branchName sql.NullString
		if err := rows.Scan(&branchID, &branchName); err != nil {
			return nil, false, fmt.Errorf("failed to scan branch info: %w", err)
		}
		if branchID.Valid && branchID.String != "" {
			branchList = append(branchList, UserBranchInfo{
				BranchID:   branchID.String,
				BranchName: branchName.String, // 如果 branch_name 为 NULL，返回空字符串
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("failed to iterate user branches: %w", err)
	}

	if len(branchList) == 0 {
		// 用户没有关联任何院区，检查是否是 Admin 的 *(ALL) 情况
		return s.handleAdminAllBranches(ctx, tenantID, userID)
	}

	return branchList, true, nil
}

// handleAdminAllBranches 处理 Admin 的 *(ALL) 情况（user_branches 表为空时返回所有 branch）
// 返回格式：["*", "branch-id-1", "branch-id-2", ...]，前面加 "*" 表示所有分支
func (s *userService) handleAdminAllBranches(ctx context.Context, tenantID, userID string) (branches []UserBranchInfo, hasBranches bool, err error) {
	// 获取用户信息，检查是否是 Admin
	user, err := s.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		// 如果用户不存在，返回空列表
		return nil, false, nil
	}

	// 如果是 Admin 角色，返回所有 branch_id/branch_name，并在前面加上 "*"
	if strings.EqualFold(user.Role, "Admin") {
		allBranches, _, err := s.branchesRepo.ListBranches(ctx, tenantID, "", 1, 1000)
		if err != nil {
			s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", userID), zap.Error(err))
			// 如果查询失败，至少返回 ["*"] 作为标记
			return []UserBranchInfo{
				{BranchID: "*", BranchName: "*(ALL)"},
			}, true, nil
		}

		// 构建所有 branch 的列表，前面加上 "*"
		// 格式：["*", "branch-id-1", "branch-id-2", ...]
		branchList := make([]UserBranchInfo, 0, len(allBranches)+1)
		// 第一个元素是 "*"，表示所有分支
		branchList = append(branchList, UserBranchInfo{
			BranchID:   "*",
			BranchName: "*(ALL)",
		})
		// 然后添加所有实际的 branch
		for _, branch := range allBranches {
			branchList = append(branchList, UserBranchInfo{
				BranchID:   branch.BranchID,
				BranchName: branch.BranchName,
			})
		}

		return branchList, true, nil
	}

	// 非 Admin 角色，返回空列表（未分配状态）
	return nil, false, nil
}

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

// canViewRole 检查当前用户是否可以查看指定角色的用户
// 与 canCreateRole 不同：SystemAdmin 可以查看 SystemAdmin/SystemOperator
func canViewRole(currentRole, targetRole string) bool {
	currentLevel := getRoleLevel(currentRole)
	targetLevel := getRoleLevel(targetRole)
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

	// 普通字段：字段不存在/空/null → 不返回字段（omitempty）
	if user.Nickname.Valid && user.Nickname.String != "" {
		dto.Nickname = user.Nickname.String
	}

	// 有 Hash 的字段（email/email_hash, phone/phone_hash）：
	// 规则 3：有 Hash 的字段
	// 情况 1：字段有值且不为 "" → 直接返回值
	// 情况 2：字段无值或 ""
	//   - 当对应的 hash 有值且不为空 → 返回占位符
	//   - 当对应的 hash null 或 "" → 返回 null（通过指针类型或特殊值）
	if user.Email.Valid && user.Email.String != "" {
		// 情况 1：有值，直接返回
		dto.Email = user.Email.String
	} else {
		// 情况 2：无值或空字符串，检查 hash
		if len(user.EmailHash) > 0 {
			// hash 有值，返回占位符
			dto.Email = "xxx@xxx.xxx"
		} else {
			// hash 为空，返回 null
			// 注意：由于 UserDTO.Email 是 string 类型，我们需要用特殊值表示 null
			// 或者修改 UserDTO 结构，使用 *string
			// 暂时先不设置，让 Handler 层处理（Handler 层会检查空字符串并返回 null）
			dto.Email = "" // Handler 层会处理为空字符串的情况
		}
	}

	if user.Phone.Valid && user.Phone.String != "" {
		// 情况 1：有值，直接返回
		dto.Phone = user.Phone.String
	} else {
		// 情况 2：无值或空字符串，检查 hash
		if len(user.PhoneHash) > 0 {
			// hash 有值，返回占位符
			dto.Phone = "xxx-xxx-xxxx"
		} else {
			// hash 为空，返回 null
			dto.Phone = "" // Handler 层会处理为空字符串的情况
		}
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
	// BranchID 和 BranchName 从 Repository 层的 LATERAL 子查询获取（第一个 branch，按 branch_name 排序）
	// 注意：不再使用 users.branch_id 字段（已废弃），也不使用 is_primary 字段（已删除）
	if user.BranchID.Valid {
		dto.BranchID = user.BranchID.String
	}
	if user.BranchName.Valid && user.BranchName.String != "" && user.BranchName.String != "-" {
		dto.BranchName = user.BranchName.String
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
		// 查询用户关联的所有院区信息（支持 1 对多关系，包含 branch_id 和 branch_name）
		userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
		if err != nil {
			s.logger.Error("Failed to get user branch IDs", zap.Error(err))
			return nil, fmt.Errorf("failed to get user branch IDs: %w", err)
		}

		if hasBranches && len(userBranches) > 0 {
			// 用户有关联的院区：提取 branch_id 列表进行 IN 查询
			// 注意：如果第一个元素是 "*"，说明是 Admin 的 *(ALL) 情况，不设置 filters.BranchIDs（查询所有用户）
			if userBranches[0].BranchID == "*" {
				// Admin 的 *(ALL) 情况：不设置 filters.BranchIDs，查询所有用户
				// filters.BranchIDs 保持为 nil，表示不过滤
			} else {
				// 提取实际的 branch_id 列表（过滤掉 "*"）
				branchIDs := make([]string, 0, len(userBranches))
				for _, branch := range userBranches {
					if branch.BranchID != "" && branch.BranchID != "*" {
						branchIDs = append(branchIDs, branch.BranchID)
					}
				}
				if len(branchIDs) > 0 {
					filters.BranchIDs = branchIDs
				}
			}
		} else {
			// 用户没有关联任何院区：只能查看 branch_name 为 NULL、"" 或 'default' 的用户（都视为空院区/默认院区）
			filters.BranchNameNull = true
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
	users, _, err := s.usersRepo.ListUsers(ctx, req.TenantID, filters, page, size)
	if err != nil {
		s.logger.Error("ListUsers failed", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 7. 角色层级过滤：过滤掉当前用户不能查看的用户
	// 注意：即使 branch 匹配，Manager 也不能查看 Admin 角色用户
	filteredUsers := make([]*domain.User, 0, len(users))
	for _, user := range users {
		// 自己总是可以查看
		if user.UserID == req.CurrentUserID {
			filteredUsers = append(filteredUsers, user)
			continue
		}
		// 检查是否可以查看该用户（角色层级检查）
		if currentUser.Role != "" && !canViewRole(currentUser.Role, user.Role) {
			// 不能查看，跳过该用户
			continue
		}
		filteredUsers = append(filteredUsers, user)
	}

	// 8. 转换为 DTO，并为每个用户填充 branch_ids
	items := make([]*UserDTO, 0, len(filteredUsers))
	for _, user := range filteredUsers {
		dto := domainUserToDTO(user)

		// 查询用户的所有 branch_ids（支持多 branch 和 *(ALL) 逻辑）
		userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, user.UserID)
		if err != nil {
			s.logger.Warn("Failed to get user branch IDs in ListUsers", zap.String("user_id", user.UserID), zap.Error(err))
			// 如果查询失败，根据角色设置默认值
			// Admin 角色：尝试查询所有 branch_id，失败则返回 ['*']
			// 其他角色：返回空数组
			if strings.EqualFold(user.Role, "Admin") {
				allBranches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
				if err != nil {
					s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", user.UserID), zap.Error(err))
					dto.BranchIDs = []string{"*"}
				} else {
					dto.BranchIDs = make([]string, 0, len(allBranches)+1)
					dto.BranchIDs = append(dto.BranchIDs, "*")
					for _, branch := range allBranches {
						dto.BranchIDs = append(dto.BranchIDs, branch.BranchID)
					}
					if len(allBranches) > 0 {
						dto.BranchID = allBranches[0].BranchID
						dto.BranchName = allBranches[0].BranchName
					}
				}
			} else {
				dto.BranchIDs = []string{}
			}
		} else {
			// 填充 branch_ids
			if hasBranches && len(userBranches) > 0 {
				// 用户有指定的 branch（从 user_branches 表查询到记录）
				dto.BranchIDs = make([]string, 0, len(userBranches))
				for _, branch := range userBranches {
					dto.BranchIDs = append(dto.BranchIDs, branch.BranchID)
				}
				// 第一个 branch 用于向后兼容（branch_id 和 branch_name）
				// 逻辑：如果 hasBranches == true，说明 user_branches 表有记录，Repository 层的 LATERAL 子查询应该已经设置了 dto.BranchID
				// 如果 dto.BranchID == ""，说明 Repository 层没有找到 branch，不应该随便填充（数据不一致）
				// 由于 Repository 层和 Service 层都按 branch_name 排序，如果 Repository 层有值，应该和 userBranches[0] 一致
				if dto.BranchID != "" {
					// Repository 层已经设置了 BranchID，确保 BranchName 也正确（可能 Repository 层没有 JOIN 到 branch_name）
					if dto.BranchName == "" && len(userBranches) > 0 {
						// 找到匹配的 branch 并设置 BranchName
						for _, branch := range userBranches {
							if branch.BranchID == dto.BranchID {
								dto.BranchName = branch.BranchName
								break
							}
						}
					}
				}
				// 如果 dto.BranchID == ""，说明 Repository 层没有找到 branch，不应该随便填充
			} else {
				// 用户没有关联任何 branch（user_branches 表没有记录）
				// 检查角色：如果是 Admin，返回所有 branch_id 列表，并在前面加上 '*'；否则返回 []（未分配）
				if strings.EqualFold(user.Role, "Admin") {
					// Admin 角色：user_branches 表无记录表示可以访问所有分支
					// 查询 branches 表，获取该 tenant 下的所有 branch_id
					allBranches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
					if err != nil {
						s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", user.UserID), zap.Error(err))
						// 如果查询失败，至少返回 ['*'] 作为标记
						dto.BranchIDs = []string{"*"}
					} else {
						// 构建 branch_ids 列表：['*', 'branch-id-1', 'branch-id-2', ...]
						dto.BranchIDs = make([]string, 0, len(allBranches)+1)
						dto.BranchIDs = append(dto.BranchIDs, "*")
						for _, branch := range allBranches {
							dto.BranchIDs = append(dto.BranchIDs, branch.BranchID)
						}
						// 设置第一个 branch 用于向后兼容
						if len(allBranches) > 0 {
							dto.BranchID = allBranches[0].BranchID
							dto.BranchName = allBranches[0].BranchName
						}
					}
				} else {
					// 非 Admin 角色：user_branches 表无记录表示未分配，返回空数组
					dto.BranchIDs = []string{}
					// 向后兼容：不设置 branch_id 和 branch_name
					dto.BranchID = ""
					dto.BranchName = ""
				}
			}
		}

		items = append(items, dto)
	}

	// 9. 重新计算 total（因为过滤后数量可能不同）
	// 注意：这里需要重新查询总数，但只计算符合角色层级过滤的用户
	// 为了简化，这里先返回当前页的过滤后数量
	// 如果需要准确的分页，需要在 Repository 层添加角色过滤
	return &ListUsersResponse{
		Items: items,
		Total: len(filteredUsers), // 注意：这里返回过滤后的数量，分页可能不准确
	}, nil
}

// GetUser 查询用户详情
func (s *userService) GetUser(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	// UserID 可以为空（Create 模式）

	// 2. 判断是 Create 模式还是 Edit 模式
	isCreateMode := req.UserID == "" || req.UserID == "new"

	if isCreateMode {
		// ========== CREATE 模式 ==========
		// 2.1 获取 available_branches（用于前端选择 branch）
		var availableBranches []BranchDTO
		branchReq := GetAvailableBranchesRequest{
			TenantID:      req.TenantID,
			CurrentUserID: req.CurrentUserID,
		}
		branchResp, err := s.GetAvailableBranches(ctx, branchReq)
		if err != nil {
			s.logger.Warn("Failed to get available branches", zap.Error(err))
			// 不返回错误，只是记录警告，availableBranches 为空数组
			availableBranches = []BranchDTO{}
		} else {
			availableBranches = branchResp.Branches
		}

		// 2.3 计算 available_tags
		var availableTags []string

		if len(req.BranchIDs) > 0 {
			// 分支 2.3.1: Branch 已指定，使用指定的 BranchIDs
			availableTags, err = s.getAvailableTagsFromBranchIDs(ctx, req.TenantID, req.BranchIDs)
		} else {
			// 分支 2.3.2: Branch 未指定，使用当前登录用户的 branch
			availableTags, err = s.getAvailableTagsFromBranches(ctx, req.TenantID, req.CurrentUserID)
		}

		if err != nil {
			s.logger.Warn("Failed to get available tags", zap.Error(err))
			// 不返回错误，只是记录警告，availableTags 为空数组
			availableTags = []string{}
		}

		// 2.4 返回结果
		return &GetUserResponse{
			User:              nil, // Create 模式下没有用户数据
			AvailableTags:     availableTags,
			AvailableBranches: availableBranches,
		}, nil
	}

	// ========== EDIT 模式 ==========
	// 3. 获取当前用户信息（用于权限检查）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		s.logger.Error("Failed to get current user", zap.Error(err))
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 4. 权限检查
	isViewingSelf := req.CurrentUserID == req.UserID
	if !isViewingSelf {
		// 获取目标用户信息
		targetUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// 检查是否可以查看目标用户（角色层级检查）
		if currentUser.Role != "" && !canViewRole(currentUser.Role, targetUser.Role) {
			return nil, fmt.Errorf("not allowed to view %s role user (current role: %s)", targetUser.Role, currentUser.Role)
		}
	}

	// 5. 查询用户详情
	user, err := s.usersRepo.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		s.logger.Error("GetUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 6. 查询用户的所有 branch 关联
	userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.UserID)
	if err != nil {
		s.logger.Error("Failed to get user branches", zap.Error(err))
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}

	// 7. 转换为 DTO
	dto := domainUserToDTO(user)

	// 8. 填充所有 branch_ids
	// 规则：
	// - 如果 user_branches 表有记录，返回具体的 branch_id 列表
	// - 如果 user_branches 表无记录：
	//   - 如果是 Admin 角色 → 返回 ['*']（表示所有分支）
	//   - 如果不是 Admin 角色 → 返回 []（未分配状态）
	// 注意：user.BranchID 是从 user_branches 表通过 LEFT JOIN LATERAL 获取的，
	// 所以如果 user_branches 表没有记录，user.BranchID 也会是 NULL
	if hasBranches && len(userBranches) > 0 {
		// 用户有指定的 branch（从 user_branches 表查询到记录）
		dto.BranchIDs = make([]string, 0, len(userBranches))
		for _, branch := range userBranches {
			dto.BranchIDs = append(dto.BranchIDs, branch.BranchID)
		}
		// 第一个 branch 用于向后兼容（branch_id 和 branch_name）
		dto.BranchID = userBranches[0].BranchID
		dto.BranchName = userBranches[0].BranchName
	} else {
		// 用户没有关联任何 branch（user_branches 表没有记录）
		// 检查角色：如果是 Admin，返回所有 branch_id 列表，并在前面加上 '*'；否则返回 []（未分配）
		if strings.EqualFold(user.Role, "Admin") {
			// Admin 角色：user_branches 表无记录表示可以访问所有分支
			// 查询 branches 表，获取该 tenant 下的所有 branch_id
			allBranches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
			if err != nil {
				s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", req.UserID), zap.Error(err))
				// 如果查询失败，至少返回 ['*'] 作为标记
				dto.BranchIDs = []string{"*"}
				dto.BranchID = ""
				dto.BranchName = ""
			} else {
				// 构建 branch_ids 列表：['*', 'branch-id-1', 'branch-id-2', ...]
				dto.BranchIDs = make([]string, 0, len(allBranches)+1)
				dto.BranchIDs = append(dto.BranchIDs, "*")
				for _, branch := range allBranches {
					dto.BranchIDs = append(dto.BranchIDs, branch.BranchID)
				}
				// 设置第一个 branch 用于向后兼容
				if len(allBranches) > 0 {
					dto.BranchID = allBranches[0].BranchID
					dto.BranchName = allBranches[0].BranchName
				}
			}
		} else {
			// 非 Admin 角色：user_branches 表无记录表示未分配，返回空数组
			dto.BranchIDs = []string{}
			// 向后兼容：不设置 branch_id 和 branch_name
			dto.BranchID = ""
			dto.BranchName = ""
		}
	}

	// 9. 获取被编辑用户所在 Branch 中存在的 tags（使用 req.UserID 而不是 req.CurrentUserID）
	availableTags, err := s.getAvailableTagsFromBranches(ctx, req.TenantID, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get available tags", zap.Error(err))
		// 不返回错误，只是记录警告，availableTags 为空数组
		availableTags = []string{}
	}

	return &GetUserResponse{
		User:              dto,
		AvailableTags:     availableTags,
		AvailableBranches: []BranchDTO{}, // Edit 模式下不需要返回 available branches（前端已有）
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

	// 4. Manager 特殊限制：如果当前用户是 Manager，验证 branch 范围
	if strings.EqualFold(currentUser.Role, "Manager") {
		// 4.1 验证目标角色是否是 Manager 可以创建的角色
		allowedRolesForManager := []string{"Manager", "IT", "Caregiver", "Nurse"}
		roleAllowed := false
		for _, allowedRole := range allowedRolesForManager {
			if strings.EqualFold(role, allowedRole) {
				roleAllowed = true
				break
			}
		}
		if !roleAllowed {
			return nil, fmt.Errorf("Manager can only create users with roles: Manager, IT, Caregiver, Nurse")
		}

		// 4.2 验证 branch_ids 必须在 Manager 的 branch 范围内
		// Service 层自己查询用户的 branch 信息，不信任 Handler 传递的值
		// 注意：*(ALL) 选项只有 Admin 可以使用，Manager 不能使用
		// 空数组 [] 表示未分配，Manager 可以创建未分配的用户
		if len(req.BranchIDs) > 0 {
			// 检查是否包含 *(ALL) 选项
			if req.BranchIDs[0] == "*" {
				return nil, fmt.Errorf("Manager cannot assign *(ALL) branch (only Admin can)")
			}

			// 验证具体的 branch_id 列表
			managerBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				s.logger.Error("Failed to get manager branches", zap.Error(err))
				return nil, fmt.Errorf("failed to get manager branches: %w", err)
			}

			if !hasBranches || len(managerBranches) == 0 {
				return nil, fmt.Errorf("Manager must have at least one branch assigned")
			}

			// 构建 Manager 的 branch_id 集合
			managerBranchIDSet := make(map[string]bool)
			for _, mb := range managerBranches {
				managerBranchIDSet[mb.BranchID] = true
			}

			// 验证所有请求的 branch_id 都在 Manager 的 branch 范围内
			for _, requestedBranchID := range req.BranchIDs {
				if !managerBranchIDSet[requestedBranchID] {
					return nil, fmt.Errorf("branch_id '%s' is not in Manager's branch scope", requestedBranchID)
				}
			}
		} else {
			// Manager 创建用户时，branch_ids 不能为空（必须指定具体的 branch）
			return nil, fmt.Errorf("Manager must assign specific branch (cannot leave unassigned)")
		}
		// 如果 branchIDs 为空数组，允许创建（表示未分配）
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

	// 5.1 Admin/Manager 凭据去重：防止跨租户出现相同 account+password 的高权限账号
	roleLowerForCheck := strings.ToLower(role)
	if roleLowerForCheck == "admin" || roleLowerForCheck == "manager" {
		if err := s.usersRepo.CheckAdminCredentialDuplicate(ctx, accountHash, passwordHash, ""); err != nil {
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
	// 处理 BranchIDs：
	// - 空数组 []：表示未分配，不创建 user_branches 记录，users.branch_id 为 NULL
	// - ['*']：表示所有分支，不创建 user_branches 记录，users.branch_id 为 NULL
	// - ['branch-id-1', ...]：创建 user_branches 记录，users.branch_id 设置为第一个
	if len(req.BranchIDs) > 0 {
		if req.BranchIDs[0] == "*" {
			// *(ALL) 选项：不设置 branch_id，表示用户可以访问所有分支
			user.BranchID = sql.NullString{Valid: false}
		} else {
			// 具体的 branch_id 列表：设置第一个 branch 用于向后兼容（users.branch_id 字段）
			user.BranchID = sql.NullString{String: req.BranchIDs[0], Valid: true}
		}
	} else {
		// 空数组：表示未分配，不设置 branch_id
		user.BranchID = sql.NullString{Valid: false}
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

	// 8. 创建多个 user_branches 关联
	// - 空数组 []：不创建关联（未分配状态）
	// - ['*']：不创建关联（所有分支）
	// - ['branch-id-1', ...]：创建关联（具体的 branch_id 列表）
	if len(req.BranchIDs) > 0 && req.BranchIDs[0] != "*" {
		if err := s.createUserBranches(ctx, req.TenantID, userID, req.BranchIDs); err != nil {
			s.logger.Error("Failed to create user branches", zap.Error(err))
			return nil, fmt.Errorf("failed to create user branches: %w", err)
		}
	}
	// 如果 branchIDs 为空或包含 '*'，不创建 user_branches 记录

	// 9. 如果角色是 Individual，自动创建对应的 Resident 记录
	// 处理 Individual 用户自动创建 Resident
	residentCreated := false
	residentError := ""
	if strings.EqualFold(role, "Individual") {
		if err := s.createResidentForIndividual(ctx, req.TenantID, userID, req.Nickname, req.BranchIDs); err != nil {
			s.logger.Error("Failed to create resident for Individual user", 
				zap.String("user_id", userID),
				zap.String("nickname", req.Nickname),
				zap.Error(err))
			residentError = err.Error()
		} else {
			residentCreated = true
		}
	}

	return &CreateUserResponse{
		UserID:          userID,
		ResidentCreated: residentCreated,
		ResidentError:   residentError,
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
		req.AlarmScope != nil || req.Tags != nil || req.BranchIDs != nil

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

	// 统一的字段更新辅助函数
	// 规则 1：普通字段（带值比较）
	// req.Field == nil → 不更新
	// req.Field != nil && *req.Field == currentValue → 不更新（值未变）
	// req.Field != nil && *req.Field != currentValue → 更新
	updateStringField := func(reqVal *string, currentVal sql.NullString, updateFunc func(string)) {
		if reqVal == nil {
			return // 不更新
		}
		newVal := strings.TrimSpace(*reqVal)
		var oldVal string
		if currentVal.Valid {
			oldVal = currentVal.String
		}
		if oldVal == newVal {
			return // 值未变，不更新
		}
		updateFunc(newVal) // 值改变，更新
	}

	// 规则 2：数组字段（带值比较）
	// req.Field == nil → 不更新
	// req.Field != nil && arraysEqual(req.Field, currentValue) → 不更新（值未变）
	// req.Field != nil && !arraysEqual(req.Field, currentValue) → 更新
	updateStringArrayField := func(reqVal []string, currentVal pq.StringArray, updateFunc func([]string)) {
		if reqVal == nil {
			return // 不更新
		}
		if stringSlicesEqual([]string(currentVal), reqVal) {
			return // 值未变，不更新
		}
		updateFunc(reqVal) // 值改变，更新
	}

	// 更新字段（带值比较，避免不必要的更新）
	// Nickname
	updateStringField(req.Nickname, targetUser.Nickname, func(val string) {
		if val == "" {
			updateUser.Nickname = sql.NullString{Valid: false} // 清空为 NULL
		} else {
			updateUser.Nickname = sql.NullString{String: val, Valid: true}
		}
	})

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

	// Role
	updateStringField(req.Role, sql.NullString{String: targetUser.Role, Valid: true}, func(val string) {
		updateUser.Role = val
	})

	// Status
	updateStringField(req.Status, sql.NullString{String: targetUser.Status, Valid: true}, func(val string) {
		updateUser.Status = val
	})

	// AlarmLevels
	updateStringArrayField(req.AlarmLevels, targetUser.AlarmLevels, func(val []string) {
		if len(val) == 0 {
			updateUser.AlarmLevels = []string{} // Repository 层会转换为 NULL
		} else {
			updateUser.AlarmLevels = val
		}
	})

	// AlarmChannels
	updateStringArrayField(req.AlarmChannels, targetUser.AlarmChannels, func(val []string) {
		if len(val) == 0 {
			updateUser.AlarmChannels = []string{} // Repository 层会转换为 NULL
		} else {
			updateUser.AlarmChannels = val
		}
	})

	// AlarmScope
	updateStringField(req.AlarmScope, targetUser.AlarmScope, func(val string) {
		if val == "" {
			updateUser.AlarmScope = sql.NullString{Valid: false} // 清空为 NULL
		} else {
			updateUser.AlarmScope = sql.NullString{String: val, Valid: true}
		}
	})
	// 处理 BranchIDs：如果提供了，更新 user_branches 关联
	if req.BranchIDs != nil {
		// Manager 特殊限制：如果当前用户是 Manager，验证 branch 范围
		// Service 层自己查询用户的 branch 信息，不信任 Handler 传递的值
		// 注意：*(ALL) 选项只有 Admin 可以使用，Manager 不能使用
		if strings.EqualFold(currentUser.Role, "Manager") {
			// Manager 不能选择 *(ALL)，必须指定具体的 branch
			if len(req.BranchIDs) > 0 {
				// 检查是否包含 *(ALL) 选项
				if req.BranchIDs[0] == "*" {
					return nil, fmt.Errorf("Manager cannot assign *(ALL) branch (only Admin can)")
				}

				managerBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
				if err != nil {
					s.logger.Error("Failed to get manager branches", zap.Error(err))
					return nil, fmt.Errorf("failed to get manager branches: %w", err)
				}

				if !hasBranches || len(managerBranches) == 0 {
					return nil, fmt.Errorf("Manager must have at least one branch assigned")
				}

				// 构建 Manager 的 branch_id 集合
				managerBranchIDSet := make(map[string]bool)
				for _, mb := range managerBranches {
					managerBranchIDSet[mb.BranchID] = true
				}

				// 验证所有请求的 branch_id 都在 Manager 的 branch 范围内
				for _, requestedBranchID := range req.BranchIDs {
					if !managerBranchIDSet[requestedBranchID] {
						return nil, fmt.Errorf("branch_id '%s' is not in Manager's branch scope", requestedBranchID)
					}
				}
			} else {
				// Manager 更新用户时，branch_ids 不能为空（必须指定具体的 branch）
				return nil, fmt.Errorf("Manager must assign specific branch (cannot leave unassigned)")
			}
		}

		// 更新 user_branches 关联
		// - 空数组 []：删除所有关联（未分配状态）
		// - ['*']：删除所有关联（所有分支，仅 Admin 可用，需要验证 role）
		// - ['branch-id-1', ...]：删除所有关联，然后创建指定的 user_branches 关联
		if len(req.BranchIDs) > 0 && req.BranchIDs[0] == "*" {
			// *(ALL) 选项：验证 role 必须是 Admin
			// 注意：如果 req.Role 不为 nil，说明正在更新 role，需要检查新 role；否则检查当前 role
			roleToCheck := targetUser.Role
			if req.Role != nil && *req.Role != "" {
				roleToCheck = *req.Role
			}
			if !strings.EqualFold(roleToCheck, "Admin") {
				return nil, fmt.Errorf("only Admin role can assign *(ALL) branch")
			}
			// 删除所有 user_branches 记录，表示用户可以访问所有分支
			if err := s.updateUserBranches(ctx, req.TenantID, req.UserID, []string{}); err != nil {
				s.logger.Error("Failed to update user branches", zap.Error(err))
				return nil, fmt.Errorf("failed to update user branches: %w", err)
			}
			updateUser.BranchID = sql.NullString{Valid: false}
		} else {
			// 正常更新：创建指定的 user_branches 关联
			if err := s.updateUserBranches(ctx, req.TenantID, req.UserID, req.BranchIDs); err != nil {
				s.logger.Error("Failed to update user branches", zap.Error(err))
				return nil, fmt.Errorf("failed to update user branches: %w", err)
			}
			// 同时更新 domain.User 中的 BranchID（用于向后兼容，使用第一个 branch）
			if len(req.BranchIDs) > 0 {
				updateUser.BranchID = sql.NullString{String: req.BranchIDs[0], Valid: true}
			} else {
				updateUser.BranchID = sql.NullString{Valid: false}
			}
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

	// 9. 如果 Individual 用户被禁用（status 改为 'disabled'），将其关联的 Resident 状态改为 'transferred'（转院）
	residentTransferred := false
	residentError := ""
	if strings.EqualFold(targetUser.Role, "Individual") && updatingStatus && req.Status != nil {
		newStatus := strings.TrimSpace(*req.Status)
		// 检查是否从其他状态改为 'disabled'（大小写不敏感比较）
		if strings.EqualFold(newStatus, "disabled") && !strings.EqualFold(targetUser.Status, "disabled") {
			if err := s.updateResidentsStatusForIndividual(ctx, req.TenantID, req.UserID, "transferred"); err != nil {
				s.logger.Error("Failed to transfer residents for Individual user",
					zap.String("user_id", req.UserID),
					zap.Error(err))
				residentError = err.Error()
			} else {
				residentTransferred = true
			}
		}
	}

	return &UpdateUserResponse{
		Success:             true,
		ResidentTransferred: residentTransferred,
		ResidentError:      residentError,
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
	// 4.1 禁止删除自己
	if req.CurrentUserID == req.UserID {
		return nil, fmt.Errorf("not allowed to delete yourself")
	}

	// 4.2 系统角色保护：SystemAdmin 和 SystemOperator 只能由 SystemAdmin 删除
	if targetUser.Role == "SystemAdmin" || targetUser.Role == "SystemOperator" {
		if req.TenantID != SystemTenantID || !strings.EqualFold(currentUser.Role, "SystemAdmin") {
			return nil, fmt.Errorf("not allowed to delete system role user (only SystemAdmin can delete system roles)")
		}
	} else {
		// 4.3 角色层级检查：可以删除同级或下级角色
		if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
			return nil, fmt.Errorf("not allowed to delete %s role (current role: %s)", targetUser.Role, currentUser.Role)
		}
	}

	// 5. 如果 Individual 用户被删除，先将其关联的 Resident 状态改为 'discharged'（出院）
	residentDischarged := false
	residentError := ""
	if strings.EqualFold(targetUser.Role, "Individual") {
		if err := s.updateResidentsStatusForIndividual(ctx, req.TenantID, req.UserID, "discharged"); err != nil {
			s.logger.Error("Failed to discharge residents for Individual user",
				zap.String("user_id", req.UserID),
				zap.Error(err))
			residentError = err.Error()
		} else {
			residentDischarged = true
		}
	}

	// 6. 软删除（设置 status = 'left'）
	updateUser := *targetUser
	updateUser.Status = "left"

	if err := s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser); err != nil {
		s.logger.Error("DeleteUser failed", zap.Error(err))
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &DeleteUserResponse{
		Success:            true,
		ResidentDischarged: residentDischarged,
		ResidentError:      residentError,
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

	// 6. Admin/Manager 凭据去重：重置密码时也要检查
	roleLower := strings.ToLower(targetUser.Role)
	if roleLower == "admin" || roleLower == "manager" {
		if err := s.usersRepo.CheckAdminCredentialDuplicate(ctx, targetUser.UserAccountHash, passwordHash, req.UserID); err != nil {
			return nil, err
		}
	}

	// 7. 更新密码
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

		// 4.1 系统角色保护：SystemAdmin 和 SystemOperator 只能由 SystemAdmin 重置 PIN
		if targetUser.Role == "SystemAdmin" || targetUser.Role == "SystemOperator" {
			if req.TenantID != SystemTenantID || !strings.EqualFold(currentUser.Role, "SystemAdmin") {
				return nil, fmt.Errorf("not allowed to reset PIN for system role user (only SystemAdmin can reset system roles)")
			}
		} else {
			// 4.2 角色层级检查：可以重置同级或下级角色的 PIN
			if currentUser.Role != "" && !canCreateRole(currentUser.Role, targetUser.Role) {
				return nil, fmt.Errorf("not allowed to reset PIN for %s role user (current role: %s)", targetUser.Role, currentUser.Role)
			}
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

	// 4. 获取用户的 branch 信息
	userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.UserID)
	if err != nil {
		s.logger.Error("Failed to get user branches", zap.Error(err))
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}

	// 5. 填充 BranchIDs 和 AvailableBranches
	if hasBranches && len(userBranches) > 0 {
		// 用户有指定的 branch（从 user_branches 表查询到记录）
		resp.BranchIDs = make([]string, 0, len(userBranches))
		for _, branch := range userBranches {
			resp.BranchIDs = append(resp.BranchIDs, branch.BranchID)
		}
		// 构建 AvailableBranches 列表
		resp.AvailableBranches = make([]BranchDTO, 0, len(userBranches))
		for _, branch := range userBranches {
			resp.AvailableBranches = append(resp.AvailableBranches, BranchDTO{
				BranchID:   branch.BranchID,
				BranchName: branch.BranchName,
			})
		}
	} else {
		// 用户没有关联任何 branch（user_branches 表没有记录）
		// 检查角色：如果是 Admin，返回所有 branch_id 列表，并在前面加上 '*'；否则返回 []（未分配）
		if strings.EqualFold(user.Role, "Admin") {
			// Admin 角色：user_branches 表无记录表示可以访问所有分支
			// 查询 branches 表，获取该 tenant 下的所有 branch_id
			allBranches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
			if err != nil {
				s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", req.UserID), zap.Error(err))
				// 如果查询失败，至少返回 ['*'] 作为标记
				resp.BranchIDs = []string{"*"}
				resp.AvailableBranches = []BranchDTO{}
			} else {
				// 构建 branch_ids 列表：['*', 'branch-id-1', 'branch-id-2', ...]
				resp.BranchIDs = make([]string, 0, len(allBranches)+1)
				resp.BranchIDs = append(resp.BranchIDs, "*")
				for _, branch := range allBranches {
					resp.BranchIDs = append(resp.BranchIDs, branch.BranchID)
				}
				// 构建 AvailableBranches 列表
				resp.AvailableBranches = make([]BranchDTO, 0, len(allBranches))
				for _, branch := range allBranches {
					resp.AvailableBranches = append(resp.AvailableBranches, BranchDTO{
						BranchID:   branch.BranchID,
						BranchName: branch.BranchName,
					})
				}
			}
		} else {
			// 非 Admin 角色：未分配状态
			resp.BranchIDs = []string{}
			resp.AvailableBranches = []BranchDTO{}
		}
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

	// 3.1 验证 tenant_id 一致性（安全措施）
	if targetUser.TenantID != req.TenantID {
		s.logger.Error("Security violation: tenant_id mismatch in UpdateAccountSettings",
			zap.String("user_id", req.UserID),
			zap.String("request_tenant_id", req.TenantID),
			zap.String("db_tenant_id", targetUser.TenantID))
		return nil, fmt.Errorf("security violation: tenant_id mismatch")
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

	// 4.1.1 更新 PIN（如果提供，!= nil 就更新，不进行任何判断）
	if req.PinHash != nil {
		if *req.PinHash == "" {
			// 空字符串，删除 PIN（设置为 NULL）
			updateUser.PinHash = []byte{}
		} else {
			// 解码 hex 字符串
			pinHashBytes, _ := hex.DecodeString(*req.PinHash)
			updateUser.PinHash = pinHashBytes
		}
	}

	// 4.2 Email 和 EmailHash 的业务逻辑处理
	// 规则：
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

	// 4.3 Phone 和 PhoneHash 的业务逻辑处理（同 Email）
	if req.Phone != nil {
		if *req.Phone == "" {
			// 删除 phone（设置为 NULL）
			updateUser.Phone = sql.NullString{Valid: false}
			// 删除 hash
			updateUser.PhoneHash = nil
		} else {
			// 保存 phone
			updateUser.Phone = sql.NullString{String: *req.Phone, Valid: true}
			// 计算或使用提供的 hash
			if req.PhoneHash != nil && *req.PhoneHash != "" {
				phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
				if len(phoneHashBytes) > 0 {
					updateUser.PhoneHash = phoneHashBytes
				}
			} else {
				// 计算 hash
				phoneHash, _ := hex.DecodeString(HashAccount(*req.Phone))
				updateUser.PhoneHash = phoneHash
			}
		}
	} else if req.PhoneHash != nil {
		// 只更新 hash，不更新 phone
		if *req.PhoneHash != "" {
			phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
			if len(phoneHashBytes) > 0 {
				updateUser.PhoneHash = phoneHashBytes
			}
		}
		// 如果 PhoneHash 为空字符串，表示删除 hash，但 Repository 层不支持单独删除 hash
		// 这种情况应该通过更新 phone 来触发，所以这里不做处理
	}

	// 5. 唯一性检查（Email 和 Phone 必须在租户内唯一）
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

// ============================================
// GetAvailableBranches 获取可用 branch 列表
// ============================================

// GetAvailableBranches 获取可用 branch 列表（用于前端创建用户时选择 branch）
// 返回所有可用的 branch（包含 branch_id 和 branch_name）
// 如果当前用户是 Manager，只返回 Manager 的 branch
func (s *userService) GetAvailableBranches(ctx context.Context, req GetAvailableBranchesRequest) (*GetAvailableBranchesResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 2. 如果提供了 CurrentUserID，检查用户角色并过滤 branch
	var branchDTOs []BranchDTO
	if req.CurrentUserID != "" {
		currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
		if err != nil {
			s.logger.Error("Failed to get current user", zap.Error(err))
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}

		// 如果是 Manager，只返回 Manager 的 branch
		if strings.EqualFold(currentUser.Role, "Manager") {
			managerBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				s.logger.Error("Failed to get manager branches", zap.Error(err))
				return nil, fmt.Errorf("failed to get manager branches: %w", err)
			}
			if hasBranches && len(managerBranches) > 0 {
				branchDTOs = make([]BranchDTO, 0, len(managerBranches))
				for _, branch := range managerBranches {
					branchDTOs = append(branchDTOs, BranchDTO{
						BranchID:   branch.BranchID,
						BranchName: branch.BranchName,
					})
				}
			} else {
				// Manager 没有关联 branch，返回空列表
				branchDTOs = []BranchDTO{}
			}
		} else {
			// Admin/IT/其他角色：返回所有 branch
			branches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
			if err != nil {
				s.logger.Error("Failed to list branches", zap.Error(err))
				return nil, fmt.Errorf("failed to list branches: %w", err)
			}
			branchDTOs = make([]BranchDTO, 0, len(branches))
			for _, branch := range branches {
				branchDTOs = append(branchDTOs, BranchDTO{
					BranchID:   branch.BranchID,
					BranchName: branch.BranchName,
				})
			}
		}
	} else {
		// 没有提供 CurrentUserID，返回所有 branch（向后兼容）
		branches, _, err := s.branchesRepo.ListBranches(ctx, req.TenantID, "", 1, 1000)
		if err != nil {
			s.logger.Error("Failed to list branches", zap.Error(err))
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}
		branchDTOs = make([]BranchDTO, 0, len(branches))
		for _, branch := range branches {
			branchDTOs = append(branchDTOs, BranchDTO{
				BranchID:   branch.BranchID,
				BranchName: branch.BranchName,
			})
		}
	}

	return &GetAvailableBranchesResponse{
		Branches: branchDTOs,
	}, nil
}

// GetAvailableCaregivers 获取可用 caregivers 列表
// 返回指定 branch 内的所有 active 状态的 caregiver/nurse
// 注意：必须先指定 branch_id，只能选择 resident 所在 branch 的 caregivers/nurse
func (s *userService) GetAvailableCaregivers(ctx context.Context, req GetAvailableCaregiversRequest) (*GetAvailableCaregiversResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.BranchID == "" {
		return nil, fmt.Errorf("branch_id is required")
	}

	// 2. 权限验证：检查当前用户是否有权限访问指定的 branch
	userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}

	// 检查是否是 Admin 的 *(ALL) 情况（user_branches 表为空，但 handleAdminAllBranches 返回了 ["*", ...]）
	isAdminAllBranches := false
	if hasBranches && len(userBranches) > 0 && userBranches[0].BranchID == "*" {
		// Admin 用户，可以访问所有 branch
		isAdminAllBranches = true
	}

	if !hasBranches || len(userBranches) == 0 {
		// 用户没有关联任何 branch，且不是 Admin，无权限
		return nil, fmt.Errorf("permission denied: user has no accessible branches")
	}

	// 检查指定的 branch_id 是否在用户可管理的 branch 列表中
	hasPermission := false
	if isAdminAllBranches {
		// Admin 用户可以访问任何 branch
		hasPermission = true
	} else {
		// 非 Admin 用户，检查 branch_id 是否在列表中
		for _, branch := range userBranches {
			if branch.BranchID == req.BranchID {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		return nil, fmt.Errorf("permission denied: user cannot access branch %s", req.BranchID)
	}

	// 3. 判断是否需要包含 Individual 角色（如果 unit_type 是 'Home'）
	includeIndividual := false
	if req.UnitID != "" {
		var unitType string
		err := s.db.QueryRowContext(ctx,
			`SELECT unit_type FROM units WHERE tenant_id = $1 AND unit_id = $2`,
			req.TenantID, req.UnitID,
		).Scan(&unitType)
		if err == nil && unitType == "Home" {
			// 如果 unit 是 Home 型，允许选择 Individual 类型
			includeIndividual = true
		}
	}

	// 4. 查询指定 branch_id 中所有 active 状态的 caregiver/nurse（如果 unit_type 是 'Home'，也包含 Individual）
	// 使用 INNER JOIN user_branches 来过滤只属于该 branch 的用户
	// 包含 user_tags (JSONB) 字段
	// 注意：不进行排序，由前端 Vue 自己排序和过滤
	roleCondition := "u.role IN ('Nurse', 'Caregiver')"
	if includeIndividual {
		// 如果 unit 是 Home 型，也包含 Individual 角色
		roleCondition = "u.role IN ('Nurse', 'Caregiver', 'Individual')"
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT
			u.user_id::text,
			u.tenant_id::text,
			u.user_account,
			COALESCE(u.nickname, '') as nickname,
			COALESCE(u.email, '') as email,
			COALESCE(u.phone, '') as phone,
			u.role,
			u.status,
			COALESCE(u.user_tags, '[]'::jsonb)::text as user_tags
		FROM users u
		INNER JOIN user_branches ub ON u.user_id = ub.user_id
		WHERE u.tenant_id = $1
		  AND ub.branch_id = $2::uuid
		  AND %s
		  AND u.status = 'active'
	`, roleCondition)

	rows, err := s.db.QueryContext(ctx, query, req.TenantID, req.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available caregivers: %w", err)
	}
	defer rows.Close()

	var caregivers []UserDTO
	for rows.Next() {
		var user UserDTO
		var userTagsRaw sql.NullString
		err := rows.Scan(
			&user.UserID,
			&user.TenantID,
			&user.UserAccount,
			&user.Nickname,
			&user.Email,
			&user.Phone,
			&user.Role,
			&user.Status,
			&userTagsRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan caregiver: %w", err)
		}
		// 解析 user_tags (JSONB array)
		if userTagsRaw.Valid && userTagsRaw.String != "" && userTagsRaw.String != "[]" {
			var tags []string
			if err := json.Unmarshal([]byte(userTagsRaw.String), &tags); err == nil {
				user.Tags = tags
			} else {
				user.Tags = []string{}
			}
		} else {
			user.Tags = []string{}
		}
		caregivers = append(caregivers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate caregivers: %w", err)
	}

	return &GetAvailableCaregiversResponse{
		Items: caregivers,
	}, nil
}

// GetAvailableCaregiverGroups 获取可用 caregiver groups 列表
// 返回当前用户可管理的 branch 内的所有 active 状态的 caregiver/nurse 的 tag 合集
// 注意：所有角色（包括 Admin）都基于绑定的 branch 进行过滤
// 如果 Admin 的 user_branches 表为空（*(ALL)），允许访问任何 branch
func (s *userService) GetAvailableCaregiverGroups(ctx context.Context, req GetAvailableCaregiverGroupsRequest) (*GetAvailableCaregiverGroupsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.BranchID == "" {
		return nil, fmt.Errorf("branch_id is required")
	}

	// 2. 获取当前用户信息（用于检查角色）
	currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// 3. 权限验证：检查当前用户是否有权限访问指定的 branch
	userBranches, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}

	// 检查权限：如果是 Admin 且 user_branches 表为空（*(ALL)），允许访问任何 branch
	hasPermission := false
	if !hasBranches || len(userBranches) == 0 {
		// 用户没有关联任何 branch
		if strings.EqualFold(currentUser.Role, "Admin") {
			// Admin 的 *(ALL) 情况：允许访问任何 branch
			hasPermission = true
		} else {
			// 非 Admin 用户没有关联任何 branch，返回空列表
			return &GetAvailableCaregiverGroupsResponse{
				Items: []CaregiverGroupDTO{},
			}, nil
		}
	} else {
		// 检查指定的 branch_id 是否在用户可管理的 branch 列表中
		for _, branch := range userBranches {
			if branch.BranchID == req.BranchID {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			// 用户没有权限访问指定的 branch，返回空列表
			return &GetAvailableCaregiverGroupsResponse{
				Items: []CaregiverGroupDTO{},
			}, nil
		}
	}

	// 4. 分两步计算：
	//    步骤1：查询指定 branch_id 中所有 active 状态的 caregiver/nurse 及其 tags
	//    步骤2：在 Go 代码中按 tag 分组，将相同 tag 的 user 归到同一 tag 组
	query := `
		SELECT DISTINCT
			u.user_id::text,
			u.tenant_id::text,
			u.user_account,
			COALESCE(u.nickname, '') as nickname,
			u.role,
			COALESCE(u.user_tags, '[]'::jsonb)::text as user_tags
		FROM users u
		INNER JOIN user_branches ub ON u.user_id = ub.user_id
		WHERE u.tenant_id = $1
		  AND ub.branch_id = $2::uuid
		  AND u.role IN ('Nurse', 'Caregiver')
		  AND u.status = 'active'
		ORDER BY u.user_account
	`

	rows, err := s.db.QueryContext(ctx, query, req.TenantID, req.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available caregivers: %w", err)
	}
	defer rows.Close()

	// 步骤2：在内存中按 tag 分组
	// tagGroups: map[tag_name] -> []UserDTO
	tagGroups := make(map[string][]UserDTO)

	for rows.Next() {
		var user UserDTO
		var userTagsRaw sql.NullString
		err := rows.Scan(
			&user.UserID,
			&user.TenantID,
			&user.UserAccount,
			&user.Nickname,
			&user.Role,
			&userTagsRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan caregiver: %w", err)
		}

		// 解析 user_tags (JSONB array)
		var tags []string
		if userTagsRaw.Valid && userTagsRaw.String != "" && userTagsRaw.String != "[]" {
			if err := json.Unmarshal([]byte(userTagsRaw.String), &tags); err != nil {
				// 如果解析失败，跳过该用户的 tags
				tags = []string{}
			}
		}
		user.Tags = tags

		// 将 user 添加到每个 tag 对应的组中
		for _, tag := range tags {
			if tag != "" {
				// 如果该 tag 组不存在，创建它
				if _, exists := tagGroups[tag]; !exists {
					tagGroups[tag] = make([]UserDTO, 0)
				}
				// 检查该 user 是否已经在该 tag 组中（避免重复）
				found := false
				for _, existingUser := range tagGroups[tag] {
					if existingUser.UserID == user.UserID {
						found = true
						break
					}
				}
				if !found {
					tagGroups[tag] = append(tagGroups[tag], user)
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate caregivers: %w", err)
	}

	// 步骤3：转换为 CaregiverGroupDTO 列表，按 tag 名称排序
	var groups []CaregiverGroupDTO
	tagNames := make([]string, 0, len(tagGroups))
	for tagName := range tagGroups {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)

	for _, tagName := range tagNames {
		members := tagGroups[tagName]
		// 生成 member_names（用于向后兼容）
		memberNames := make([]string, 0, len(members))
		for _, member := range members {
			name := member.Nickname
			if name == "" {
				name = member.UserAccount
			}
			memberNames = append(memberNames, name)
		}
		sort.Strings(memberNames) // 排序 member_names

		group := CaregiverGroupDTO{
			TagName:     tagName,
			MemberCount: len(members),
			MemberNames: memberNames,
			Members:     members,
		}
		groups = append(groups, group)
	}

	return &GetAvailableCaregiverGroupsResponse{
		Items: groups,
	}, nil
}

// createResidentForIndividual 为 Individual 用户自动创建对应的 Resident 记录
// 规则：
//   - Resident 的 FirstName = Individual 的 Nickname
//   - Resident 的 lastName = Branch 的 branch_name
//   - Resident 的 branch_id = Individual 的 branch_id（从 user_branches 表获取第一个）
//   - Resident 的 caregiver = Individual（在 resident_caregivers 表中，user_list 包含 Individual 的 user_id）
func (s *userService) createResidentForIndividual(ctx context.Context, tenantID, individualUserID, individualNickname string, branchIDs []string) error {
	// 1. 获取 Individual 用户的 branch_id（从 user_branches 表获取第一个）
	var branchID string
	var branchName string

	if len(branchIDs) > 0 && branchIDs[0] != "*" {
		// 使用传入的 branchIDs（取第一个）
		branchID = branchIDs[0]
	} else {
		// 从 user_branches 表查询
		err := s.db.QueryRowContext(ctx,
			`SELECT ub.branch_id::text 
			 FROM user_branches ub 
			 WHERE ub.tenant_id = $1 AND ub.user_id::text = $2 
			 ORDER BY ub.branch_id 
			 LIMIT 1`,
			tenantID, individualUserID,
		).Scan(&branchID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Individual 用户没有关联 branch，无法创建 Resident
				return fmt.Errorf("individual user has no branch assigned, cannot create resident")
			}
			return fmt.Errorf("failed to get individual user branch: %w", err)
		}
	}

	// 2. 获取 branch_name（从 branches 表）
	err := s.db.QueryRowContext(ctx,
		`SELECT branch_name FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
		tenantID, branchID,
	).Scan(&branchName)
	if err != nil {
		if err == sql.ErrNoRows {
			branchName = "" // 如果 branch 不存在，使用空字符串
		} else {
			return fmt.Errorf("failed to get branch name: %w", err)
		}
	}

	// 3. 生成 resident_id (UUID)
	residentID := uuid.New().String()

	// 4. 生成 resident_account（使用 Individual 的 user_account，或者生成一个新的）
	// 先获取 Individual 的 user_account
	var userAccount string
	err = s.db.QueryRowContext(ctx,
		`SELECT user_account FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, individualUserID,
	).Scan(&userAccount)
	if err != nil {
		return fmt.Errorf("failed to get individual user account: %w", err)
	}

	// 使用 Individual 的 user_account 作为 resident_account
	residentAccount := strings.ToLower(strings.TrimSpace(userAccount))
	accountHashHex := HashAccount(residentAccount)
	accountHash, err := hex.DecodeString(accountHashHex)
	if err != nil || len(accountHash) == 0 {
		return fmt.Errorf("failed to hash resident account")
	}

	// 5. 生成默认密码 hash（使用 Individual 的默认密码 "individual!"）
	passwordHashHex := HashPassword("individual!")
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil || len(passwordHash) == 0 {
		return fmt.Errorf("failed to hash password")
	}

	// 6. 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 7. 插入 residents 表
	admissionDate := time.Now()
	nickname := individualNickname
	if nickname == "" {
		nickname = userAccount // 如果没有 nickname，使用 user_account
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO residents (
			resident_id, tenant_id, resident_account, resident_account_hash, 
			password_hash, nickname, status, role, admission_date, branch_id
		) VALUES ($1, $2, LOWER($3), $4, $5, $6, 'active', 'Resident', $7, $8)`,
		residentID, tenantID, residentAccount, accountHash, passwordHash, nickname, admissionDate, branchID,
	)
	if err != nil {
		return fmt.Errorf("failed to create resident: %w", err)
	}

	// 8. 插入 resident_phi 表
	// first_name = Individual 的 nickname
	// last_name = Branch 的 branch_name
	firstName := nickname
	lastName := branchName

	_, err = tx.ExecContext(ctx,
		`INSERT INTO resident_phi (
			tenant_id, resident_id, first_name, last_name
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, resident_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name`,
		tenantID, residentID, firstName, lastName,
	)
	if err != nil {
		return fmt.Errorf("failed to create resident_phi: %w", err)
	}

	// 9. 插入 resident_caregivers 表
	// user_list 包含 Individual 的 user_id
	// caregiver_id 由数据库自动生成（如果表有 DEFAULT gen_random_uuid()）
	userListJSON, err := json.Marshal([]string{individualUserID})
	if err != nil {
		return fmt.Errorf("failed to marshal user_list: %w", err)
	}

	// 生成 caregiver_id (UUID)
	caregiverID := uuid.New().String()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO resident_caregivers (
			caregiver_id, tenant_id, resident_id, user_list
		) VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (tenant_id, resident_id) DO UPDATE SET
			user_list = EXCLUDED.user_list`,
		caregiverID, tenantID, residentID, userListJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create resident_caregivers: %w", err)
	}

	// 10. 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Created resident for Individual user",
		zap.String("tenant_id", tenantID),
		zap.String("individual_user_id", individualUserID),
		zap.String("resident_id", residentID),
		zap.String("branch_id", branchID),
		zap.String("branch_name", branchName),
	)

	return nil
}

// updateResidentsStatusForIndividual 将 Individual 用户关联的 Resident 状态更新为指定状态
// 通过查询 resident_caregivers 表，查找 user_list JSONB 数组包含 Individual user_id 的记录
// newStatus: 'discharged'（出院）或 'transferred'（转院）
func (s *userService) updateResidentsStatusForIndividual(ctx context.Context, tenantID, individualUserID, newStatus string) error {
	if newStatus != "discharged" && newStatus != "transferred" {
		return fmt.Errorf("invalid status: %s (must be 'discharged' or 'transferred')", newStatus)
	}
	// 1. 查询 Individual 用户关联的所有 Resident IDs
	// 使用 @> 操作符检查 user_list JSONB 数组是否包含 user_id
	userIDJSON, err := json.Marshal([]string{individualUserID})
	if err != nil {
		return fmt.Errorf("failed to marshal user_id: %w", err)
	}

	query := `
		SELECT DISTINCT resident_id::text
		FROM resident_caregivers
		WHERE tenant_id = $1
		  AND user_list @> $2::jsonb
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, userIDJSON)
	if err != nil {
		return fmt.Errorf("failed to query resident_caregivers: %w", err)
	}
	defer rows.Close()

	var residentIDs []string
	for rows.Next() {
		var residentID string
		if err := rows.Scan(&residentID); err != nil {
			return fmt.Errorf("failed to scan resident_id: %w", err)
		}
		residentIDs = append(residentIDs, residentID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate resident_ids: %w", err)
	}

	// 2. 如果没有关联的 Resident，直接返回（不报错）
	if len(residentIDs) == 0 {
		s.logger.Info("No residents found for Individual user",
			zap.String("tenant_id", tenantID),
			zap.String("individual_user_id", individualUserID))
		return nil
	}

	// 3. 更新所有关联的 Resident 状态为指定状态
	// 使用事务确保原子性
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 使用 IN 查询批量更新
	placeholders := make([]string, len(residentIDs))
	var updateQuery string
	var args []any
	if newStatus == "transferred" {
		var defaultBranchID string
		if err := tx.QueryRowContext(ctx,
			`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`,
			tenantID, domain.DefaultBranchName,
		).Scan(&defaultBranchID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("default branch not found for tenant")
			}
			return fmt.Errorf("resolve default branch: %w", err)
		}
		args = make([]any, len(residentIDs)+3)
		args[0] = tenantID
		args[1] = newStatus
		args[2] = defaultBranchID
		for i, residentID := range residentIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+4)
			args[i+3] = residentID
		}
		updateQuery = fmt.Sprintf(`
		UPDATE residents
		SET status = $2,
		    branch_id = $3::uuid,
		    unit_id = NULL,
		    room_id = NULL,
		    bed_id = NULL
		WHERE tenant_id = $1
		  AND resident_id::text IN (%s)
		  AND status != $2
	`, strings.Join(placeholders, ", "))
	} else {
		args = make([]any, len(residentIDs)+2)
		args[0] = tenantID
		args[1] = newStatus
		for i, residentID := range residentIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args[i+2] = residentID
		}
		updateQuery = fmt.Sprintf(`
		UPDATE residents
		SET status = $2
		WHERE tenant_id = $1
		  AND resident_id::text IN (%s)
		  AND status != $2
	`, strings.Join(placeholders, ", "))
	}

	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update residents status: %w", err)
	}

	if newStatus == "transferred" && len(residentIDs) > 0 {
		ph := make([]string, len(residentIDs))
		delArgs := make([]any, len(residentIDs)+1)
		delArgs[0] = tenantID
		for i, rid := range residentIDs {
			ph[i] = fmt.Sprintf("$%d", i+2)
			delArgs[i+1] = rid
		}
		delQ := fmt.Sprintf(
			`DELETE FROM resident_caregivers WHERE tenant_id = $1 AND resident_id::text IN (%s)`,
			strings.Join(ph, ", "),
		)
		if _, err := tx.ExecContext(ctx, delQ, delArgs...); err != nil {
			return fmt.Errorf("failed to clear resident_caregivers for transferred: %w", err)
		}
	}

	// 4. 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Updated residents status for Individual user",
		zap.String("tenant_id", tenantID),
		zap.String("individual_user_id", individualUserID),
		zap.String("new_status", newStatus),
		zap.Int("resident_count", len(residentIDs)),
		zap.Strings("resident_ids", residentIDs))

	return nil
}

// getStringFromMap 从 map 中获取字符串值（辅助函数）
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getAvailableTagsFromBranches 获取当前用户所在 Branch 中存在的 tags
// 查询逻辑：
// 1. 获取当前用户的所有 branch_ids（如果是 Admin 且 user_branches 表为空，返回所有 branch）
// 2. 查询这些 branch_ids 中所有用户的 tags（排除当前用户自己）
// 3. 去重并返回
func (s *userService) getAvailableTagsFromBranches(ctx context.Context, tenantID, currentUserID string) ([]string, error) {
	if tenantID == "" || currentUserID == "" {
		return []string{}, nil
	}

	// 1. 获取当前用户的所有 branch_ids
	// 注意：getUserBranchIDs 已经处理了 Admin 的 *(ALL) 情况（user_branches 表为空时返回所有 branch）
	userBranches, hasBranches, err := s.getUserBranchIDs(ctx, tenantID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}

	if !hasBranches || len(userBranches) == 0 {
		// 用户没有关联任何 branch（且不是 Admin 的 *(ALL) 情况），返回空数组
		return []string{}, nil
	}

	// 2. 提取 branch_id 列表（过滤掉 "*"，只使用实际的 branch_id）
	branchIDs := make([]string, 0, len(userBranches))
	for _, branch := range userBranches {
		if branch.BranchID != "" && branch.BranchID != "*" {
			branchIDs = append(branchIDs, branch.BranchID)
		}
	}

	if len(branchIDs) == 0 {
		return []string{}, nil
	}

	// 3. 先获取当前用户的 tags（用于后续排除）
	currentUser, err := s.usersRepo.GetUser(ctx, tenantID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	currentUserTagsSet := make(map[string]bool)
	// 解析当前用户的 tags（JSONB 格式）
	if currentUser.Tags.Valid && currentUser.Tags.String != "" {
		var tags []string
		if err := json.Unmarshal([]byte(currentUser.Tags.String), &tags); err == nil {
			for _, tag := range tags {
				if tag != "" {
					currentUserTagsSet[tag] = true
				}
			}
		}
	}

	// 4. 查询这些 branch_ids 中所有用户的 tags（排除当前用户自己）
	// 使用 PostgreSQL 的 ANY 操作符和 JSONB 数组操作
	query := `
		SELECT DISTINCT tag
		FROM users u
		INNER JOIN user_branches ub ON u.user_id = ub.user_id
		CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(u.user_tags, '[]'::jsonb)) AS tag
		WHERE u.tenant_id = $1
		  AND ub.branch_id = ANY($2::uuid[])
		  AND u.user_id::text != $3
		  AND tag IS NOT NULL
		  AND tag != ''
		ORDER BY tag
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(branchIDs), currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		if tag != "" {
			// 排除当前用户已有的 tags（即使其他用户也有相同的 tag）
			if !currentUserTagsSet[tag] {
				tags = append(tags, tag)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// getAvailableTagsFromBranchIDs 根据 branch_ids 直接查询 tags（不依赖 user_id）
// 用于 Create 模式：当用户已指定 branch 时，直接查询这些 branch 内的所有 tags
// 注意：如果 branchIDs 包含 "*"，会过滤掉 "*" 字符；如果只包含 "*"，则查询所有 branch 的 tags
func (s *userService) getAvailableTagsFromBranchIDs(ctx context.Context, tenantID string, branchIDs []string) ([]string, error) {
	if tenantID == "" {
		return []string{}, nil
	}
	if len(branchIDs) == 0 {
		return []string{}, nil
	}

	// 过滤掉 "*" 字符（防御性编程）
	validBranchIDs := make([]string, 0, len(branchIDs))
	isAllBranches := false
	for _, branchID := range branchIDs {
		if branchID == "*" {
			isAllBranches = true
			// 如果包含 "*"，表示查询所有 branch 的 tags，不设置 branch_id 过滤
		} else if branchID != "" {
			validBranchIDs = append(validBranchIDs, branchID)
		}
	}

	// 构建 SQL 查询
	var query string
	var rows *sql.Rows
	var err error

	if isAllBranches || len(validBranchIDs) == 0 {
		// 如果包含 "*" 或过滤后为空（表示 *(ALL) 情况），查询所有 branch 的 tags
		query = `
			SELECT DISTINCT tag
			FROM users u
			INNER JOIN user_branches ub ON u.user_id = ub.user_id
			CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(u.user_tags, '[]'::jsonb)) AS tag
			WHERE u.tenant_id = $1
			  AND tag IS NOT NULL
			  AND tag != ''
			ORDER BY tag
		`
		rows, err = s.db.QueryContext(ctx, query, tenantID)
	} else {
		// 查询指定 branch_ids 中的 tags
		query = `
			SELECT DISTINCT tag
			FROM users u
			INNER JOIN user_branches ub ON u.user_id = ub.user_id
			CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(u.user_tags, '[]'::jsonb)) AS tag
			WHERE u.tenant_id = $1
			  AND ub.branch_id = ANY($2::uuid[])
			  AND tag IS NOT NULL
			  AND tag != ''
			ORDER BY tag
		`
		rows, err = s.db.QueryContext(ctx, query, tenantID, pq.Array(validBranchIDs))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query available tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// tagsEqual 比较两个字符串切片是否相等（忽略顺序）
func tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// 创建 map 用于快速查找
	aMap := make(map[string]int)
	for _, s := range a {
		aMap[s]++
	}

	bMap := make(map[string]int)
	for _, s := range b {
		bMap[s]++
	}

	// 比较两个 map
	if len(aMap) != len(bMap) {
		return false
	}

	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}

// stringSlicesEqual 比较两个字符串切片是否相等（忽略顺序）
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// 创建 map 用于快速查找
	aMap := make(map[string]int)
	for _, s := range a {
		aMap[s]++
	}

	bMap := make(map[string]int)
	for _, s := range b {
		bMap[s]++
	}

	// 比较两个 map
	if len(aMap) != len(bMap) {
		return false
	}

	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}
