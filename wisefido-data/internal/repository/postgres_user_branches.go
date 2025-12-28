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
			user_branch_id::text,
			tenant_id::text,
			user_id::text,
			branch_id::text,
			is_primary
		FROM user_branches
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY is_primary DESC, branch_id::text ASC
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
			&ub.IsPrimary,
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

// GetUserPrimaryBranch 获取用户的主院区
func (r *PostgresUserBranchesRepository) GetUserPrimaryBranch(ctx context.Context, tenantID, userID string) (*domain.UserBranch, error) {
	if tenantID == "" || userID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			user_branch_id::text,
			tenant_id::text,
			user_id::text,
			branch_id::text,
			is_primary
		FROM user_branches
		WHERE tenant_id = $1 AND user_id = $2 AND is_primary = TRUE
		LIMIT 1
	`

	var ub domain.UserBranch
	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(
		&ub.UserBranchID,
		&ub.TenantID,
		&ub.UserID,
		&ub.BranchID,
		&ub.IsPrimary,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有主院区，返回 nil
		}
		return nil, fmt.Errorf("failed to get user primary branch: %w", err)
	}

	return &ub, nil
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
			branch_id::text,
			is_primary
		FROM user_branches
		WHERE tenant_id = $1 AND branch_id = $2
		ORDER BY is_primary DESC, user_id::text ASC
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
			&ub.IsPrimary,
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

	// 检查用户是否已有其他院区关联
	var count int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_branches WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userBranch.UserID,
	).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("failed to count user branches: %w", err)
	}

	// 如果这是第一个院区关联，自动设置为主院区
	isPrimary := userBranch.IsPrimary
	if count == 0 {
		isPrimary = true
	} else if isPrimary {
		// 如果设置为主院区，需要先将其他主院区设置为 FALSE
		_, err = tx.ExecContext(ctx,
			`UPDATE user_branches SET is_primary = FALSE 
			 WHERE tenant_id = $1 AND user_id = $2 AND is_primary = TRUE`,
			tenantID, userBranch.UserID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to clear other primary branches: %w", err)
		}
	}

	// 插入新关联
	var userBranchID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO user_branches (tenant_id, user_id, branch_id, is_primary)
		 VALUES ($1, $2, $3, $4)
		 RETURNING user_branch_id::text`,
		tenantID, userBranch.UserID, userBranch.BranchID, isPrimary,
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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 如果设置为主院区，需要先将该用户的其他主院区设置为 FALSE
	if userBranch.IsPrimary {
		_, err = tx.ExecContext(ctx,
			`UPDATE user_branches SET is_primary = FALSE 
			 WHERE tenant_id = $1 AND user_id = $2 AND is_primary = TRUE AND user_branch_id != $3`,
			tenantID, userBranch.UserID, userBranchID,
		)
		if err != nil {
			return fmt.Errorf("failed to clear other primary branches: %w", err)
		}
	}

	// 更新关联
	_, err = tx.ExecContext(ctx,
		`UPDATE user_branches 
		 SET is_primary = $1
		 WHERE tenant_id = $2 AND user_branch_id = $3`,
		userBranch.IsPrimary, tenantID, userBranchID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user branch: %w", err)
	}

	// 检查是否有行被更新
	result, err := tx.ExecContext(ctx,
		`SELECT 1 FROM user_branches WHERE tenant_id = $1 AND user_branch_id = $2`,
		tenantID, userBranchID,
	)
	if err != nil {
		return fmt.Errorf("failed to check user branch: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user branch not found: tenant_id '%s', user_branch_id '%s'", tenantID, userBranchID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetPrimaryBranch 设置用户的主院区
func (r *PostgresUserBranchesRepository) SetPrimaryBranch(ctx context.Context, tenantID, userID, branchID string) error {
	if tenantID == "" || userID == "" || branchID == "" {
		return fmt.Errorf("tenant_id, user_id, and branch_id are required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 先将该用户的所有主院区设置为 FALSE
	_, err = tx.ExecContext(ctx,
		`UPDATE user_branches SET is_primary = FALSE 
		 WHERE tenant_id = $1 AND user_id = $2 AND is_primary = TRUE`,
		tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear other primary branches: %w", err)
	}

	// 将指定的院区关联设置为主院区
	result, err := tx.ExecContext(ctx,
		`UPDATE user_branches SET is_primary = TRUE 
		 WHERE tenant_id = $1 AND user_id = $2 AND branch_id = $3`,
		tenantID, userID, branchID,
	)
	if err != nil {
		return fmt.Errorf("failed to set primary branch: %w", err)
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user branch not found: tenant_id '%s', user_id '%s', branch_id '%s'", tenantID, userID, branchID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
