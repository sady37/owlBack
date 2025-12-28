package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// UserBranchesRepository 用户-院区关联Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type UserBranchesRepository interface {
	// ========== 查询接口 ==========

	// GetUserBranches 获取用户的所有院区关联
	// 返回用户所属的所有院区（包括主院区和非主院区）
	GetUserBranches(ctx context.Context, tenantID, userID string) ([]*domain.UserBranch, error)

	// GetUserPrimaryBranch 获取用户的主院区
	// 返回用户的主院区关联（is_primary = TRUE）
	// 如果用户没有主院区，返回 nil, nil
	GetUserPrimaryBranch(ctx context.Context, tenantID, userID string) (*domain.UserBranch, error)

	// GetBranchUsers 获取院区的所有用户关联
	// 返回属于该院区的所有用户关联
	GetBranchUsers(ctx context.Context, tenantID, branchID string) ([]*domain.UserBranch, error)

	// ========== 创建接口 ==========

	// CreateUserBranch 创建用户-院区关联
	// 注意：
	//   - 唯一性约束：同一租户下，一个用户不能重复关联同一个院区
	//   - 如果这是用户的第一个院区关联，自动设置为主院区（is_primary = TRUE）
	//   - 如果用户已有其他院区关联，默认 is_primary = FALSE
	CreateUserBranch(ctx context.Context, tenantID string, userBranch *domain.UserBranch) (string, error)

	// ========== 更新接口 ==========

	// UpdateUserBranch 更新用户-院区关联
	// 可以更新 is_primary 字段
	// 注意：如果设置 is_primary = TRUE，需要先将该用户的其他主院区设置为 FALSE
	UpdateUserBranch(ctx context.Context, tenantID, userBranchID string, userBranch *domain.UserBranch) error

	// SetPrimaryBranch 设置用户的主院区
	// 业务逻辑：
	//   1) 将指定 branch_id 的关联设置为 is_primary = TRUE
	//   2) 将该用户的其他所有关联设置为 is_primary = FALSE
	//   3) 确保一个用户只有一个主院区
	SetPrimaryBranch(ctx context.Context, tenantID, userID, branchID string) error

	// ========== 删除接口 ==========

	// DeleteUserBranch 删除用户-院区关联（通过 user_branch_id）
	DeleteUserBranch(ctx context.Context, tenantID, userBranchID string) error

	// DeleteUserBranchByUserAndBranch 删除用户-院区关联（通过 user_id 和 branch_id）
	// 用于根据用户和院区删除关联
	DeleteUserBranchByUserAndBranch(ctx context.Context, tenantID, userID, branchID string) error
}
