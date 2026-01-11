package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wisefido-data/internal/domain"
)

// PostgresUserBranchesRepository 用户-院区关联Repository实现
// 实现UserBranchesRepository接口，使用domain.UserBranch领域模型
type PostgresUserBranchesRepository struct {
	db *sql.DB
}

// NewPostgresUserBranchesRepository 创建用户-院区关联Repository
func NewPostgresUserBranchesRepository(db *sql.DB) *PostgresUserBranchesRepository {
	return &PostgresUserBranchesRepository{db: db}
}

// 确保实现了接口
var _ UserBranchesRepository = (*PostgresUserBranchesRepository)(nil)

// GetUserBranches 获取用户的所有院区关联
func (r *PostgresUserBranchesRepository) GetUserBranches(ctx context.Context, tenantID, userID string) ([]*domain.UserBranch, error) {
	if tenantID == "" || userID == "" {
		return []*domain.UserBranch{}, nil
	}

	query := `
		SELECT 
			ub.user_branch_id::text,
			ub.tenant_id::text,
			ub.user_id::text,
			ub.branch_id::text
		FROM user_branches ub
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
		WHERE ub.tenant_id = $1 AND ub.user_id = $2
		ORDER BY COALESCE(b.branch_name, '') ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user branches: %w", err)
	}
	defer rows.Close()

	userBranches := []*domain.UserBranch{}
	for rows.Next() {
		var ub domain.UserBranch
		err := rows.Scan(
			&ub.UserBranchID,
			&ub.TenantID,
			&ub.UserID,
			&ub.BranchID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user branch: %w", err)
		}
		userBranches = append(userBranches, &ub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user branches: %w", err)
	}

	return userBranches, nil
}

// GetBranchUsers 获取院区的所有用户关联
func (r *PostgresUserBranchesRepository) GetBranchUsers(ctx context.Context, tenantID, branchID string) ([]*domain.UserBranch, error) {
	if tenantID == "" || branchID == "" {
		return []*domain.UserBranch{}, nil
	}

	query := `
		SELECT 
			user_branch_id::text,
			tenant_id::text,
			user_id::text,
			branch_id::text
		FROM user_branches
		WHERE tenant_id = $1 AND branch_id = $2
		ORDER BY user_id::text ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch users: %w", err)
	}
	defer rows.Close()

	userBranches := []*domain.UserBranch{}
	for rows.Next() {
		var ub domain.UserBranch
		err := rows.Scan(
			&ub.UserBranchID,
			&ub.TenantID,
			&ub.UserID,
			&ub.BranchID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user branch: %w", err)
		}
		userBranches = append(userBranches, &ub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate branch users: %w", err)
	}

	return userBranches, nil
}

// CreateUserBranch 创建用户-院区关联
func (r *PostgresUserBranchesRepository) CreateUserBranch(ctx context.Context, tenantID string, userBranch *domain.UserBranch) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if userBranch == nil {
		return "", fmt.Errorf("user_branch is required")
	}
	if userBranch.UserID == "" {
		return "", fmt.Errorf("user_id is required")
	}
	if userBranch.BranchID == "" {
		return "", fmt.Errorf("branch_id is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 检查是否已存在该关联
	var existingID string
	err = tx.QueryRowContext(ctx,
		`SELECT user_branch_id::text FROM user_branches 
		 WHERE tenant_id = $1 AND user_id = $2 AND branch_id = $3`,
		tenantID, userBranch.UserID, userBranch.BranchID,
	).Scan(&existingID)
	if err == nil {
		// 已存在，返回现有ID
		return existingID, nil
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check existing user branch: %w", err)
	}

	// 插入新关联
	var userBranchID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO user_branches (tenant_id, user_id, branch_id)
		 VALUES ($1, $2, $3)
		 RETURNING user_branch_id::text`,
		tenantID, userBranch.UserID, userBranch.BranchID,
	).Scan(&userBranchID)
	if err != nil {
		return "", fmt.Errorf("failed to create user branch: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userBranchID, nil
}

// UpdateUserBranch 更新用户-院区关联
func (r *PostgresUserBranchesRepository) UpdateUserBranch(ctx context.Context, tenantID, userBranchID string, userBranch *domain.UserBranch) error {
	if tenantID == "" || userBranchID == "" {
		return fmt.Errorf("tenant_id and user_branch_id are required")
	}
	if userBranch == nil {
		return fmt.Errorf("user_branch is required")
	}

	// 更新关联（目前只更新 branch_id，如果需要可以扩展）
	// 注意：由于移除了 is_primary 字段，UpdateUserBranch 主要用于更新 branch_id
	// 如果需要更新其他字段，可以在这里扩展
	result, err := r.db.ExecContext(ctx,
		`UPDATE user_branches 
		 SET branch_id = $1
		 WHERE tenant_id = $2 AND user_branch_id = $3`,
		userBranch.BranchID, tenantID, userBranchID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user branch: %w", err)
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user branch not found: tenant_id '%s', user_branch_id '%s'", tenantID, userBranchID)
	}

	return nil
}

// DeleteUserBranch 删除用户-院区关联（通过 user_branch_id）
func (r *PostgresUserBranchesRepository) DeleteUserBranch(ctx context.Context, tenantID, userBranchID string) error {
	if tenantID == "" || userBranchID == "" {
		return fmt.Errorf("tenant_id and user_branch_id are required")
	}

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_branches WHERE tenant_id = $1 AND user_branch_id = $2`,
		tenantID, userBranchID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete user branch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user branch not found: tenant_id '%s', user_branch_id '%s'", tenantID, userBranchID)
	}

	return nil
}

// DeleteUserBranchByUserAndBranch 删除用户-院区关联（通过 user_id 和 branch_id）
func (r *PostgresUserBranchesRepository) DeleteUserBranchByUserAndBranch(ctx context.Context, tenantID, userID, branchID string) error {
	if tenantID == "" || userID == "" || branchID == "" {
		return fmt.Errorf("tenant_id, user_id, and branch_id are required")
	}

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_branches WHERE tenant_id = $1 AND user_id = $2 AND branch_id = $3`,
		tenantID, userID, branchID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete user branch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user branch not found: tenant_id '%s', user_id '%s', branch_id '%s'", tenantID, userID, branchID)
	}

	return nil
}

// DeleteAllUserBranches 删除用户的所有院区关联（通过 user_id）
func (r *PostgresUserBranchesRepository) DeleteAllUserBranches(ctx context.Context, tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return fmt.Errorf("tenant_id and user_id are required")
	}

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_branches WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete all user branches: %w", err)
	}

	return nil
}
