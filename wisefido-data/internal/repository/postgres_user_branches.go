package repository

import (
	"context"
	"database/sql"

	"wisefido-data/internal/domain"
)

// PostgresUserBranchesRepository — owl_v2 stub 实现。
//
// v2 schema 已删除 user_branches 表（用户与院区关系改为通过 user_roles.role_id +
// roles.tenant_prefix 派生 + role_permissions.resource_scope INET 表达，由 utils/spatial
// 计算）。这个 repo 保留接口签名以兼容上层 service / handler 调用，所有方法返回空数据 / no-op，
// 不再触发 SQL。
//
// 调用方语义影响：
//   - 查询返回空数组 → service 层会把"无 branch 限制"视为"全部 branch 可访问"（管理员行为）
//   - 创建/更新/删除 no-op 静默成功，前端"分配 branch"按钮不会报错但也不持久化
//
// 后续真正业务化时改用 v2 RBAC（user_roles 表），那时回来重写本 repo 或删除。
type PostgresUserBranchesRepository struct {
	db *sql.DB
}

func NewPostgresUserBranchesRepository(db *sql.DB) *PostgresUserBranchesRepository {
	return &PostgresUserBranchesRepository{db: db}
}

var _ UserBranchesRepository = (*PostgresUserBranchesRepository)(nil)

func (r *PostgresUserBranchesRepository) GetUserBranches(ctx context.Context, tenantID, userID string) ([]*domain.UserBranch, error) {
	return []*domain.UserBranch{}, nil
}

func (r *PostgresUserBranchesRepository) GetBranchUsers(ctx context.Context, tenantID, branchID string) ([]*domain.UserBranch, error) {
	return []*domain.UserBranch{}, nil
}

func (r *PostgresUserBranchesRepository) CreateUserBranch(ctx context.Context, tenantID string, userBranch *domain.UserBranch) (string, error) {
	// no-op 兼容；返回零 UUID 占位
	return "00000000-0000-0000-0000-000000000000", nil
}

func (r *PostgresUserBranchesRepository) UpdateUserBranch(ctx context.Context, tenantID, userBranchID string, userBranch *domain.UserBranch) error {
	return nil
}

func (r *PostgresUserBranchesRepository) DeleteUserBranch(ctx context.Context, tenantID, userBranchID string) error {
	return nil
}

func (r *PostgresUserBranchesRepository) DeleteUserBranchByUserAndBranch(ctx context.Context, tenantID, userID, branchID string) error {
	return nil
}

func (r *PostgresUserBranchesRepository) DeleteAllUserBranches(ctx context.Context, tenantID, userID string) error {
	return nil
}
