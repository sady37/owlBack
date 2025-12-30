package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresBranchesRepository 院区Repository实现
// 实现BranchesRepository接口，使用domain.Branch领域模型
type PostgresBranchesRepository struct {
	db *sql.DB
}

// NewPostgresBranchesRepository 创建院区Repository
func NewPostgresBranchesRepository(db *sql.DB) *PostgresBranchesRepository {
	return &PostgresBranchesRepository{db: db}
}

// 确保实现了接口
var _ BranchesRepository = (*PostgresBranchesRepository)(nil)

// GetBranch 获取院区信息（通过 branch_id）
func (r *PostgresBranchesRepository) GetBranch(ctx context.Context, tenantID, branchID string) (*domain.Branch, error) {
	if tenantID == "" || branchID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			branch_id::text,
			tenant_id::text,
			branch_name,
			description,
			created_at,
			updated_at
		FROM branches
		WHERE tenant_id = $1 AND branch_id = $2
	`

	var branch domain.Branch
	var description sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, branchID).Scan(
		&branch.BranchID,
		&branch.TenantID,
		&branch.BranchName,
		&description,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("branch not found: tenant_id '%s', branch_id '%s'", tenantID, branchID)
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	branch.Description = description
	branch.CreatedAt = createdAt
	branch.UpdatedAt = updatedAt

	return &branch, nil
}

// GetBranchByName 获取院区信息（通过 branch_name）
func (r *PostgresBranchesRepository) GetBranchByName(ctx context.Context, tenantID, branchName string) (*domain.Branch, error) {
	if tenantID == "" || branchName == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			branch_id::text,
			tenant_id::text,
			branch_name,
			description,
			created_at,
			updated_at
		FROM branches
		WHERE tenant_id = $1 AND branch_name = $2
	`

	var branch domain.Branch
	var description sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, branchName).Scan(
		&branch.BranchID,
		&branch.TenantID,
		&branch.BranchName,
		&description,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("branch not found: tenant_id '%s', branch_name '%s'", tenantID, branchName)
		}
		return nil, fmt.Errorf("failed to get branch by name: %w", err)
	}

	branch.Description = description
	branch.CreatedAt = createdAt
	branch.UpdatedAt = updatedAt

	return &branch, nil
}

// ListBranches 列出所有院区（支持搜索）
func (r *PostgresBranchesRepository) ListBranches(ctx context.Context, tenantID string, search string, page, size int) ([]*domain.Branch, int, error) {
	if tenantID == "" {
		return []*domain.Branch{}, 0, nil
	}

	// 构建 WHERE 条件
	where := []string{"b.tenant_id = $1"}
	args := []any{tenantID}
	argIdx := 2

	// 搜索条件：模糊匹配 branch_name, description, user_nickname, unit_name, resident_nickname
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		// 使用 EXISTS 子查询检查关联的 users, units, residents
		// 每个 $%d 占位符都需要一个参数，虽然值相同，但位置不同
		searchCondition := fmt.Sprintf(`(
			LOWER(b.branch_name) LIKE $%d OR
			LOWER(COALESCE(b.description, '')) LIKE $%d OR
			EXISTS (
				SELECT 1 FROM user_branches ub
				INNER JOIN users u ON u.user_id = ub.user_id AND u.tenant_id = ub.tenant_id
				WHERE ub.tenant_id = b.tenant_id AND ub.branch_id::text = b.branch_id::text
				  AND (LOWER(u.user_account) LIKE $%d OR LOWER(COALESCE(u.nickname, '')) LIKE $%d)
			) OR
			EXISTS (
				SELECT 1 FROM units u
				WHERE u.tenant_id = b.tenant_id AND u.branch_id::text = b.branch_id::text
				  AND LOWER(u.unit_name) LIKE $%d
			) OR
			EXISTS (
				SELECT 1 FROM residents r
				INNER JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
				WHERE r.tenant_id = b.tenant_id AND u.branch_id::text = b.branch_id::text
				  AND LOWER(r.nickname) LIKE $%d
			)
		)`, argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5)
		where = append(where, searchCondition)
		// 每个占位符都需要一个参数，虽然值相同
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
		argIdx += 6
	}

	whereClause := strings.Join(where, " AND ")

	// 计算总数
	countQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT b.branch_id) FROM branches b WHERE %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count branches: %w", err)
	}

	// 查询列表
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`
		SELECT DISTINCT
			b.branch_id::text,
			b.tenant_id::text,
			b.branch_name,
			b.description,
			b.created_at,
			b.updated_at
		FROM branches b
		WHERE %s
		ORDER BY b.branch_name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list branches: %w", err)
	}
	defer rows.Close()

	branches := []*domain.Branch{}
	for rows.Next() {
		var branch domain.Branch
		var description sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&branch.BranchID,
			&branch.TenantID,
			&branch.BranchName,
			&description,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan branch: %w", err)
		}

		branch.Description = description
		branch.CreatedAt = createdAt
		branch.UpdatedAt = updatedAt

		branches = append(branches, &branch)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branches, total, nil
}

// CreateBranch 创建院区
func (r *PostgresBranchesRepository) CreateBranch(ctx context.Context, tenantID string, branch *domain.Branch) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if branch == nil {
		return "", fmt.Errorf("branch is required")
	}

	// 处理可空字段
	// 注意：branch_name 的空值处理应该在 Service 层完成，Repository 层不做业务逻辑处理
	branchName := branch.BranchName
	var descriptionArg any = nil
	if branch.Description.Valid && branch.Description.String != "" {
		descriptionArg = branch.Description.String
	}

	// 插入新院区
	var branchID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO branches (tenant_id, branch_name, description)
		 VALUES ($1, $2, $3)
		 RETURNING branch_id::text`,
		tenantID, branchName, descriptionArg,
	).Scan(&branchID)
	if err != nil {
		// 检查是否是唯一性约束冲突
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return "", fmt.Errorf("branch_name already exists: tenant_id '%s', branch_name '%s'", tenantID, branchName)
		}
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	return branchID, nil
}

// UpdateBranch 更新院区信息
func (r *PostgresBranchesRepository) UpdateBranch(ctx context.Context, tenantID, branchID string, branch *domain.Branch) error {
	if tenantID == "" || branchID == "" {
		return fmt.Errorf("tenant_id and branch_id are required")
	}
	if branch == nil {
		return fmt.Errorf("branch is required")
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID, branchID}
	argIdx := 3

	if branch.BranchName != "" {
		updates = append(updates, fmt.Sprintf("branch_name = $%d", argIdx))
		args = append(args, branch.BranchName)
		argIdx++
	}

	// 处理 description
	if branch.Description.Valid {
		if branch.Description.String != "" {
			updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
			args = append(args, branch.Description.String)
			argIdx++
		} else {
			updates = append(updates, "description = NULL")
		}
	}

	// 自动更新 updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	if len(updates) == 1 {
		// 只有 updated_at，没有其他字段需要更新
		return nil
	}

	query := fmt.Sprintf(
		`UPDATE branches SET %s WHERE tenant_id = $1 AND branch_id = $2`,
		strings.Join(updates, ", "),
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		// 检查是否是唯一性约束冲突
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("branch_name already exists: tenant_id '%s', branch_name '%s'", tenantID, branch.BranchName)
		}
		return fmt.Errorf("failed to update branch: %w", err)
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("branch not found: tenant_id '%s', branch_id '%s'", tenantID, branchID)
	}

	return nil
}

// UpdateBranchFields 更新院区信息（使用更新模型）
func (r *PostgresBranchesRepository) UpdateBranchFields(ctx context.Context, tenantID, branchID string, update *domain.BranchUpdate) error {
	if tenantID == "" || branchID == "" {
		return fmt.Errorf("tenant_id and branch_id are required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updates := []string{}
	args := []any{tenantID, branchID}
	argIdx := 3

	// 处理 UpdateString
	if update.BranchName != nil {
		switch update.BranchName.Action {
		case domain.UpdateActionUpdate:
			// 注意：branch_name 的空值处理应该在 Service 层完成，Repository 层不做业务逻辑处理
			// 如果 Service 层未处理，这里会返回错误（NOT NULL 约束）
			if update.BranchName.Value == "" {
				return fmt.Errorf("branch_name cannot be empty (NOT NULL constraint)")
			}
			// 检查唯一性约束（如果更新 branch_name）
			var exists bool
			err = tx.QueryRowContext(ctx,
				`SELECT EXISTS(
					SELECT 1 FROM branches 
					WHERE tenant_id = $1 AND branch_name = $2 AND branch_id != $3
				)`,
				tenantID, update.BranchName.Value, branchID,
			).Scan(&exists)
			if err != nil {
				return fmt.Errorf("failed to check branch_name uniqueness: %w", err)
			}
			if exists {
				return fmt.Errorf("branch_name already exists: tenant_id '%s', branch_name '%s'", tenantID, update.BranchName.Value)
			}
			updates = append(updates, fmt.Sprintf("branch_name = $%d", argIdx))
			args = append(args, update.BranchName.Value)
			argIdx++
		case domain.UpdateActionDelete:
			// branch_name 是 NOT NULL，不能删除，只能更新
			return fmt.Errorf("branch_name cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// 不更新，跳过
		}
	}

	if update.Description != nil {
		switch update.Description.Action {
		case domain.UpdateActionUpdate:
			updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
			args = append(args, update.Description.Value)
			argIdx++
		case domain.UpdateActionDelete:
			updates = append(updates, "description = NULL")
		case domain.UpdateActionKeep:
			// 不更新，跳过
		}
	}

	if len(updates) == 0 {
		// 没有字段需要更新，但可以只更新 updated_at
		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	} else {
		// 自动更新 updated_at
		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	}

	query := fmt.Sprintf(`
		UPDATE branches
		SET %s
		WHERE tenant_id = $1 AND branch_id = $2
	`, strings.Join(updates, ", "))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		// 检查是否是唯一性约束冲突
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("branch_name already exists: tenant_id '%s'", tenantID)
		}
		return fmt.Errorf("failed to update branch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("branch not found: tenant_id '%s', branch_id '%s'", tenantID, branchID)
	}

	return tx.Commit()
}

// DeleteBranch 删除院区
func (r *PostgresBranchesRepository) DeleteBranch(ctx context.Context, tenantID, branchID string) error {
	if tenantID == "" || branchID == "" {
		return fmt.Errorf("tenant_id and branch_id are required")
	}

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
		tenantID, branchID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("branch not found: tenant_id '%s', branch_id '%s'", tenantID, branchID)
	}

	return nil
}
