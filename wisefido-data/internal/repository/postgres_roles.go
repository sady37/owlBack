package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresRolesRepository 角色Repository实现（强类型版本）
// 实现RolesRepository接口，使用domain.Role领域模型
// 遵循"bottom-up"设计原则，Repository层负责数据访问和数据完整性验证
type PostgresRolesRepository struct {
	db *sql.DB
}

// NewPostgresRolesRepository 创建角色Repository
func NewPostgresRolesRepository(db *sql.DB) *PostgresRolesRepository {
	return &PostgresRolesRepository{db: db}
}

// 确保实现了接口
var _ RolesRepository = (*PostgresRolesRepository)(nil)

// GetRole 查询单个角色
// 功能：根据roleID查询单个角色
func (r *PostgresRolesRepository) GetRole(ctx context.Context, roleID string) (*domain.Role, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role_id is required")
	}

	query := `
		SELECT 
			role_id::text,
			tenant_id,
			role_code,
			description,
			is_system,
			is_active
		FROM roles
		WHERE role_id = $1
	`

	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.RoleID,
		&role.TenantID,
		&role.RoleCode,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found: role_id=%s", roleID)
		}
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return &role, nil
}

// GetRoleByCode 通过role_code查询角色
// 功能：根据role_code查询角色（用于程序引用）
// 输入：tenantID（可选，用于区分系统角色和租户角色），roleCode
func (r *PostgresRolesRepository) GetRoleByCode(ctx context.Context, tenantID *string, roleCode string) (*domain.Role, error) {
	if roleCode == "" {
		return nil, fmt.Errorf("role_code is required")
	}

	var query string
	var args []any

	if tenantID != nil && *tenantID != "" {
		// 查询指定租户的角色
		query = `
			SELECT 
				role_id::text,
				tenant_id,
				role_code,
				description,
				is_system,
				is_active
			FROM roles
			WHERE tenant_id = $1 AND role_code = $2
		`
		args = []any{*tenantID, roleCode}
	} else {
		// 查询系统角色（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		query = `
			SELECT 
				role_id::text,
				tenant_id,
				role_code,
				description,
				is_system,
				is_active
			FROM roles
			WHERE tenant_id = $1 AND role_code = $2
		`
		args = []any{systemTenantID, roleCode}
	}

	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&role.RoleID,
		&role.TenantID,
		&role.RoleCode,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found: role_code=%s", roleCode)
		}
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return &role, nil
}

// ListRoles 查询角色列表
// 功能：查询角色列表，支持过滤和分页
// 输入：tenantID（可选），filters（search, is_system, is_active），page, size
func (r *PostgresRolesRepository) ListRoles(ctx context.Context, tenantID *string, filter RolesFilter, page, size int) ([]*domain.Role, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// tenantID过滤
	if tenantID != nil && *tenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", argN))
		args = append(args, *tenantID)
		argN++
	} else {
		// 默认查询系统角色（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		where = append(where, fmt.Sprintf("tenant_id = $%d", argN))
		args = append(args, systemTenantID)
		argN++
	}

	// search过滤（模糊搜索role_code, description）
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(role_code ILIKE $%d OR description ILIKE $%d)", argN, argN))
		args = append(args, "%"+filter.Search+"%")
		argN++
	}

	// is_system过滤
	if filter.IsSystem != nil {
		where = append(where, fmt.Sprintf("is_system = $%d", argN))
		args = append(args, *filter.IsSystem)
		argN++
	}

	// is_active过滤
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active = $%d", argN))
		args = append(args, *filter.IsActive)
		argN++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM roles ` + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// 分页
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	// 查询数据
	query := `
		SELECT 
			role_id::text,
			tenant_id,
			role_code,
			description,
			is_system,
			is_active
		FROM roles
		` + whereClause + `
		ORDER BY is_system DESC, role_code ASC
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	args = append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := []*domain.Role{}
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(
			&role.RoleID,
			&role.TenantID,
			&role.RoleCode,
			&role.Description,
			&role.IsSystem,
			&role.IsActive,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return roles, total, nil
}

// CreateRole 创建角色
// 功能：创建角色（Repository层不限制业务规则，业务规则在Service层验证）
// 验证（Repository层）：
//   - role_code必填
//   - (tenant_id, role_code)唯一性检查
//   - description必填
func (r *PostgresRolesRepository) CreateRole(ctx context.Context, tenantID string, role *domain.Role) (string, error) {
	if role == nil {
		return "", fmt.Errorf("role is required")
	}

	// 验证必填字段
	if role.RoleCode == "" {
		return "", fmt.Errorf("role_code is required")
	}
	if role.Description == "" {
		return "", fmt.Errorf("description is required")
	}

	// 验证tenant_id存在（外键约束）
	if tenantID != "" {
		var exists bool
		err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE tenant_id = $1)", tenantID).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("failed to validate tenant_id: %w", err)
		}
		if !exists {
			return "", fmt.Errorf("tenant not found: tenant_id=%s", tenantID)
		}
	}

	// 检查(tenant_id, role_code)唯一性
	var existingRoleID string
	checkQuery := `
		SELECT role_id::text
		FROM roles
		WHERE (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($1::uuid, '00000000-0000-0000-0000-000000000000'::uuid))
		  AND role_code = $2
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, checkQuery, tenantID, role.RoleCode).Scan(&existingRoleID)
	if err == nil {
		return "", fmt.Errorf("role already exists: role_code=%s (role_id=%s)", role.RoleCode, existingRoleID)
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check role uniqueness: %w", err)
	}

	// 插入新角色
	insertQuery := `
		INSERT INTO roles (tenant_id, role_code, description, is_system, is_active)
		VALUES ($1, $2, $3, $4, COALESCE($5, TRUE))
		RETURNING role_id::text
	`

	var roleID string
	var tenantIDVal interface{}
	if tenantID != "" {
		tenantIDVal = tenantID
	} else {
		tenantIDVal = nil
	}

	var isActiveVal interface{}
	if role.IsActive.Valid {
		isActiveVal = role.IsActive.Bool
	} else {
		isActiveVal = nil // 使用COALESCE默认值TRUE
	}

	err = r.db.QueryRowContext(ctx, insertQuery,
		tenantIDVal,
		role.RoleCode,
		role.Description,
		role.IsSystem,
		isActiveVal,
	).Scan(&roleID)
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}

	return roleID, nil
}

// UpdateRole 更新角色
// 功能：更新角色信息（部分更新）
// 验证（Repository层）：
//   - 数据完整性：role_code唯一性（如果更新role_code）
//   - 数据一致性：不能将is_system从TRUE改为FALSE（或反之）
// 业务规则（Service层）：
//   - 系统角色（is_system=TRUE）：只能更新is_active，不能修改description和role_code
//   - Resident和Family：不能禁用（is_active不能改为FALSE）
func (r *PostgresRolesRepository) UpdateRole(ctx context.Context, roleID string, role *domain.Role) error {
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}
	if role == nil {
		return fmt.Errorf("role is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. 查询现有角色
	var existingRole domain.Role
	err = tx.QueryRowContext(ctx, `
		SELECT role_id::text, tenant_id, role_code, description, is_system, is_active
		FROM roles
		WHERE role_id = $1
	`, roleID).Scan(
		&existingRole.RoleID,
		&existingRole.TenantID,
		&existingRole.RoleCode,
		&existingRole.Description,
		&existingRole.IsSystem,
		&existingRole.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("role not found: role_id=%s", roleID)
		}
		return fmt.Errorf("failed to query role: %w", err)
	}

	// 2. 验证数据一致性：is_system不能改变
	if role.IsSystem != existingRole.IsSystem {
		return fmt.Errorf("cannot change is_system: role_id=%s (current=%v, new=%v)", roleID, existingRole.IsSystem, role.IsSystem)
	}

	// 3. 如果更新role_code，检查唯一性
	if role.RoleCode != "" && role.RoleCode != existingRole.RoleCode {
		var existingID string
		checkQuery := `
			SELECT role_id::text
			FROM roles
			WHERE (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($1::uuid, '00000000-0000-0000-0000-000000000000'::uuid))
			  AND role_code = $2
			  AND role_id != $3
			LIMIT 1
		`
		var tenantIDVal interface{}
		if existingRole.TenantID.Valid {
			tenantIDVal = existingRole.TenantID.String
		} else {
			tenantIDVal = nil
		}
		err = tx.QueryRowContext(ctx, checkQuery, tenantIDVal, role.RoleCode, roleID).Scan(&existingID)
		if err == nil {
			return fmt.Errorf("role_code already exists: role_code=%s (role_id=%s)", role.RoleCode, existingID)
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check role_code uniqueness: %w", err)
		}
	}

	// 4. 构建UPDATE语句（部分更新）
	set := []string{}
	args := []any{roleID}
	argN := 2

	if role.RoleCode != "" && role.RoleCode != existingRole.RoleCode {
		set = append(set, fmt.Sprintf("role_code = $%d", argN))
		args = append(args, role.RoleCode)
		argN++
	}

	if role.Description != "" && role.Description != existingRole.Description {
		set = append(set, fmt.Sprintf("description = $%d", argN))
		args = append(args, role.Description)
		argN++
	}

	if role.IsActive.Valid {
		set = append(set, fmt.Sprintf("is_active = $%d", argN))
		args = append(args, role.IsActive.Bool)
		argN++
	}

	if len(set) == 0 {
		// 没有需要更新的字段
		return tx.Commit()
	}

	// 5. 执行UPDATE
	updateQuery := "UPDATE roles SET " + strings.Join(set, ", ") + " WHERE role_id = $1"
	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return tx.Commit()
}

// DeleteRole 删除角色
// 功能：删除角色
// 验证（Repository层）：
//   - 数据完整性：检查是否存在
// 业务规则（Service层）：
//   - 系统角色（is_system=TRUE）：不能删除
func (r *PostgresRolesRepository) DeleteRole(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}

	// 检查角色是否存在
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM roles WHERE role_id = $1)", roleID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("role not found: role_id=%s", roleID)
	}

	// 删除角色
	_, err = r.db.ExecContext(ctx, "DELETE FROM roles WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

