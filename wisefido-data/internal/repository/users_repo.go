package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// PermissionCheck 权限检查结果
type PermissionCheck struct {
	AssignedOnly bool // 是否仅限分配的资源
	BranchOnly   bool // 是否仅限同一 Branch 的资源
}

// UsersRepository 用户Repository接口
// 使用强类型领域模型，不使用map[string]any
type UsersRepository interface {
	// 用户基本信息
	GetUser(ctx context.Context, tenantID, userID string) (*domain.User, error)
	GetUserByAccount(ctx context.Context, tenantID, account string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.User, error)
	GetUserByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.User, error)
	ListUsers(ctx context.Context, tenantID string, filters UserFilters, page, size int) ([]*domain.User, int, error)
	CreateUser(ctx context.Context, tenantID string, user *domain.User) (string, error)
	UpdateUser(ctx context.Context, tenantID, userID string, user *domain.User) error
	DeleteUser(ctx context.Context, tenantID, userID string) error

	// 权限检查
	GetResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error)

	// 唯一性检查
	CheckEmailUniqueness(ctx context.Context, tenantID, email, excludeUserID string) error
	CheckPhoneUniqueness(ctx context.Context, tenantID, phone, excludeUserID string) error
}

// UserFilters 用户查询过滤器
type UserFilters struct {
	Role           string
	Status         string
	BranchName     string   // 精确匹配 branch_name（单院区场景）
	BranchNameNull bool     // 如果为 true，匹配 branch_name IS NULL OR branch_name = '' OR branch_name = '-'（null、""、"-" 都视为空院区）
	BranchIDs      []string // 多个 branch_id 的 IN 查询（支持 1 对多关系）
	Tag            string   // 查询包含指定tag的用户
	Search         string   // 模糊搜索：支持user_account, nickname, email, phone
}
