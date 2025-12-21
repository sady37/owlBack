package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresRolePermissionsRepository 角色权限Repository实现（强类型版本）
// 实现RolePermissionsRepository接口，使用domain.RolePermission领域模型
// 遵循"bottom-up"设计原则，Repository层负责数据访问和数据完整性验证
// Repository层不限制业务规则，只负责数据访问
type PostgresRolePermissionsRepository struct {
	db *sql.DB
}

// NewPostgresRolePermissionsRepository 创建角色权限Repository
func NewPostgresRolePermissionsRepository(db *sql.DB) *PostgresRolePermissionsRepository {
	return &PostgresRolePermissionsRepository{db: db}
}

// 确保实现了接口
var _ RolePermissionsRepository = (*PostgresRolePermissionsRepository)(nil)

// GetPermission 查询单个权限
// 功能：根据permissionID查询单个权限
func (r *PostgresRolePermissionsRepository) GetPermission(ctx context.Context, permissionID string) (*domain.RolePermission, error) {
	if permissionID == "" {
		return nil, fmt.Errorf("permission_id is required")
	}

	query := `
		SELECT 
			permission_id::text,
			tenant_id,
			role_code,
			resource_type,
			permission_type,
			assigned_only,
			branch_only
		FROM role_permissions
		WHERE permission_id = $1
	`

	var perm domain.RolePermission
	err := r.db.QueryRowContext(ctx, query, permissionID).Scan(
		&perm.PermissionID,
		&perm.TenantID,
		&perm.RoleCode,
		&perm.ResourceType,
		&perm.PermissionType,
		&perm.AssignedOnly,
		&perm.BranchOnly,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("permission not found: permission_id=%s", permissionID)
		}
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

	return &perm, nil
}

// GetPermissionByKey 通过(role_code, resource_type, permission_type)查询权限
// 功能：根据role_code, resource_type, permission_type查询权限（用于检查特定权限是否存在）
func (r *PostgresRolePermissionsRepository) GetPermissionByKey(ctx context.Context, tenantID *string, roleCode, resourceType, permissionType string) (*domain.RolePermission, error) {
	if roleCode == "" || resourceType == "" || permissionType == "" {
		return nil, fmt.Errorf("role_code, resource_type, permission_type are required")
	}

	// 验证permission_type格式
	if permissionType != "R" && permissionType != "C" && permissionType != "U" && permissionType != "D" {
		return nil, fmt.Errorf("invalid permission_type: %s (must be R, C, U, or D)", permissionType)
	}

	var query string
	var args []any

	if tenantID != nil && *tenantID != "" {
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND role_code = $2 AND resource_type = $3 AND permission_type = $4
		`
		args = []any{*tenantID, roleCode, resourceType, permissionType}
	} else {
		// 查询系统权限（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND role_code = $2 AND resource_type = $3 AND permission_type = $4
		`
		args = []any{systemTenantID, roleCode, resourceType, permissionType}
	}

	var perm domain.RolePermission
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&perm.PermissionID,
		&perm.TenantID,
		&perm.RoleCode,
		&perm.ResourceType,
		&perm.PermissionType,
		&perm.AssignedOnly,
		&perm.BranchOnly,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("permission not found: role_code=%s, resource_type=%s, permission_type=%s", roleCode, resourceType, permissionType)
		}
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

	return &perm, nil
}

// ListPermissions 查询权限列表
// 功能：查询权限列表，支持过滤和分页
// 输入：tenantID（可选），filters（role_code, resource_type, permission_type, assigned_only, branch_only），page, size
func (r *PostgresRolePermissionsRepository) ListPermissions(ctx context.Context, tenantID *string, filter RolePermissionsFilter, page, size int) ([]*domain.RolePermission, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// tenantID过滤
	if tenantID != nil && *tenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", argN))
		args = append(args, *tenantID)
		argN++
	} else {
		// 默认查询系统权限（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		where = append(where, fmt.Sprintf("tenant_id = $%d", argN))
		args = append(args, systemTenantID)
		argN++
	}

	// role_code过滤
	if filter.RoleCode != "" {
		where = append(where, fmt.Sprintf("role_code = $%d", argN))
		args = append(args, filter.RoleCode)
		argN++
	}

	// resource_type过滤
	if filter.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", argN))
		args = append(args, filter.ResourceType)
		argN++
	}

	// permission_type过滤
	if filter.PermissionType != "" {
		where = append(where, fmt.Sprintf("permission_type = $%d", argN))
		args = append(args, filter.PermissionType)
		argN++
	}

	// assigned_only过滤
	if filter.AssignedOnly != nil {
		where = append(where, fmt.Sprintf("assigned_only = $%d", argN))
		args = append(args, *filter.AssignedOnly)
		argN++
	}

	// branch_only过滤
	if filter.BranchOnly != nil {
		where = append(where, fmt.Sprintf("branch_only = $%d", argN))
		args = append(args, *filter.BranchOnly)
		argN++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM role_permissions ` + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
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
			permission_id::text,
			tenant_id,
			role_code,
			resource_type,
			permission_type,
			assigned_only,
			branch_only
		FROM role_permissions
		` + whereClause + `
		ORDER BY role_code, resource_type, permission_type
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	args = append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := []*domain.RolePermission{}
	for rows.Next() {
		var perm domain.RolePermission
		if err := rows.Scan(
			&perm.PermissionID,
			&perm.TenantID,
			&perm.RoleCode,
			&perm.ResourceType,
			&perm.PermissionType,
			&perm.AssignedOnly,
			&perm.BranchOnly,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return permissions, total, nil
}

// GetPermissionsByRole 查询某个角色的所有权限
// 功能：查询某个角色的所有权限（用于权限管理界面）
func (r *PostgresRolePermissionsRepository) GetPermissionsByRole(ctx context.Context, tenantID *string, roleCode string) ([]*domain.RolePermission, error) {
	if roleCode == "" {
		return nil, fmt.Errorf("role_code is required")
	}

	var query string
	var args []any

	if tenantID != nil && *tenantID != "" {
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND role_code = $2
			ORDER BY resource_type, permission_type
		`
		args = []any{*tenantID, roleCode}
	} else {
		// 查询系统权限（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND role_code = $2
			ORDER BY resource_type, permission_type
		`
		args = []any{systemTenantID, roleCode}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := []*domain.RolePermission{}
	for rows.Next() {
		var perm domain.RolePermission
		if err := rows.Scan(
			&perm.PermissionID,
			&perm.TenantID,
			&perm.RoleCode,
			&perm.ResourceType,
			&perm.PermissionType,
			&perm.AssignedOnly,
			&perm.BranchOnly,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return permissions, nil
}

// GetPermissionsByResource 查询某个资源的所有权限
// 功能：查询某个资源的所有权限（用于权限管理界面）
func (r *PostgresRolePermissionsRepository) GetPermissionsByResource(ctx context.Context, tenantID *string, resourceType string) ([]*domain.RolePermission, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource_type is required")
	}

	var query string
	var args []any

	if tenantID != nil && *tenantID != "" {
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND resource_type = $2
			ORDER BY role_code, permission_type
		`
		args = []any{*tenantID, resourceType}
	} else {
		// 查询系统权限（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		query = `
			SELECT 
				permission_id::text,
				tenant_id,
				role_code,
				resource_type,
				permission_type,
				assigned_only,
				branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND resource_type = $2
			ORDER BY role_code, permission_type
		`
		args = []any{systemTenantID, resourceType}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := []*domain.RolePermission{}
	for rows.Next() {
		var perm domain.RolePermission
		if err := rows.Scan(
			&perm.PermissionID,
			&perm.TenantID,
			&perm.RoleCode,
			&perm.ResourceType,
			&perm.PermissionType,
			&perm.AssignedOnly,
			&perm.BranchOnly,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return permissions, nil
}

// CreatePermission 创建权限
// 功能：创建权限
// 验证（Repository层）：
//   - role_code, resource_type, permission_type必填
//   - permission_type必须是'R', 'C', 'U', 'D'之一
//   - (tenant_id, role_code, resource_type, permission_type)唯一性检查
//   - role_code必须存在于roles表（外键验证，通过值匹配）
func (r *PostgresRolePermissionsRepository) CreatePermission(ctx context.Context, tenantID string, permission *domain.RolePermission) (string, error) {
	if permission == nil {
		return "", fmt.Errorf("permission is required")
	}

	// 验证必填字段
	if permission.RoleCode == "" {
		return "", fmt.Errorf("role_code is required")
	}
	if permission.ResourceType == "" {
		return "", fmt.Errorf("resource_type is required")
	}
	if permission.PermissionType == "" {
		return "", fmt.Errorf("permission_type is required")
	}

	// 验证permission_type格式
	if permission.PermissionType != "R" && permission.PermissionType != "C" && permission.PermissionType != "U" && permission.PermissionType != "D" {
		return "", fmt.Errorf("invalid permission_type: %s (must be R, C, U, or D)", permission.PermissionType)
	}

	// 验证role_code存在（通过值匹配，非外键）
	var roleExists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM roles WHERE role_code = $1)
	`, permission.RoleCode).Scan(&roleExists)
	if err != nil {
		return "", fmt.Errorf("failed to validate role_code: %w", err)
	}
	if !roleExists {
		return "", fmt.Errorf("role_code not found: role_code=%s (role must exist in roles table)", permission.RoleCode)
	}

	// 检查(tenant_id, role_code, resource_type, permission_type)唯一性
	var existingPermissionID string
	checkQuery := `
		SELECT permission_id::text
		FROM role_permissions
		WHERE (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($1::uuid, '00000000-0000-0000-0000-000000000000'::uuid))
		  AND role_code = $2
		  AND resource_type = $3
		  AND permission_type = $4
		LIMIT 1
	`
	var tenantIDVal interface{}
	if tenantID != "" {
		tenantIDVal = tenantID
	} else {
		tenantIDVal = nil
	}

	err = r.db.QueryRowContext(ctx, checkQuery, tenantIDVal, permission.RoleCode, permission.ResourceType, permission.PermissionType).Scan(&existingPermissionID)
	if err == nil {
		return "", fmt.Errorf("permission already exists: role_code=%s, resource_type=%s, permission_type=%s (permission_id=%s)", permission.RoleCode, permission.ResourceType, permission.PermissionType, existingPermissionID)
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check permission uniqueness: %w", err)
	}

	// 插入新权限
	insertQuery := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING permission_id::text
	`

	var permissionID string
	err = r.db.QueryRowContext(ctx, insertQuery,
		tenantIDVal,
		permission.RoleCode,
		permission.ResourceType,
		permission.PermissionType,
		permission.AssignedOnly,
		permission.BranchOnly,
	).Scan(&permissionID)
	if err != nil {
		return "", fmt.Errorf("failed to create permission: %w", err)
	}

	return permissionID, nil
}

// BatchCreatePermissions 批量创建权限
// 功能：批量插入权限，使用ON CONFLICT处理重复（用于初始化系统权限）
func (r *PostgresRolePermissionsRepository) BatchCreatePermissions(ctx context.Context, tenantID string, permissions []*domain.RolePermission) (int, []error, error) {
	if len(permissions) == 0 {
		return 0, nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var successCount int
	var errors []error

	var tenantIDVal interface{}
	if tenantID != "" {
		tenantIDVal = tenantID
	} else {
		tenantIDVal = nil
	}

	for _, perm := range permissions {
		if perm == nil {
			errors = append(errors, fmt.Errorf("permission is nil"))
			continue
		}

		// 验证必填字段
		if perm.RoleCode == "" {
			errors = append(errors, fmt.Errorf("role_code is required"))
			continue
		}
		if perm.ResourceType == "" {
			errors = append(errors, fmt.Errorf("resource_type is required"))
			continue
		}
		if perm.PermissionType == "" {
			errors = append(errors, fmt.Errorf("permission_type is required"))
			continue
		}

		// 验证permission_type格式
		if perm.PermissionType != "R" && perm.PermissionType != "C" && perm.PermissionType != "U" && perm.PermissionType != "D" {
			errors = append(errors, fmt.Errorf("invalid permission_type: %s (must be R, C, U, or D)", perm.PermissionType))
			continue
		}

		// 验证role_code存在
		var roleExists bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE role_code = $1)`, perm.RoleCode).Scan(&roleExists)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to validate role_code %s: %w", perm.RoleCode, err))
			continue
		}
		if !roleExists {
			errors = append(errors, fmt.Errorf("role_code not found: role_code=%s", perm.RoleCode))
			continue
		}

		// 插入权限（使用ON CONFLICT处理重复）
		insertQuery := `
			INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
			DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only
		`

		_, err = tx.ExecContext(ctx, insertQuery,
			tenantIDVal,
			perm.RoleCode,
			perm.ResourceType,
			perm.PermissionType,
			perm.AssignedOnly,
			perm.BranchOnly,
		)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create permission (role_code=%s, resource_type=%s, permission_type=%s): %w", perm.RoleCode, perm.ResourceType, perm.PermissionType, err))
			continue
		}

		successCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, errors, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return successCount, errors, nil
}

// UpdatePermission 更新权限
// 功能：更新权限信息（部分更新）
// 验证（Repository层）：
//   - 数据完整性：唯一性检查（如果更新role_code/resource_type/permission_type）
func (r *PostgresRolePermissionsRepository) UpdatePermission(ctx context.Context, permissionID string, permission *domain.RolePermission) error {
	if permissionID == "" {
		return fmt.Errorf("permission_id is required")
	}
	if permission == nil {
		return fmt.Errorf("permission is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. 查询现有权限
	var existingPerm domain.RolePermission
	err = tx.QueryRowContext(ctx, `
		SELECT permission_id::text, tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only
		FROM role_permissions
		WHERE permission_id = $1
	`, permissionID).Scan(
		&existingPerm.PermissionID,
		&existingPerm.TenantID,
		&existingPerm.RoleCode,
		&existingPerm.ResourceType,
		&existingPerm.PermissionType,
		&existingPerm.AssignedOnly,
		&existingPerm.BranchOnly,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("permission not found: permission_id=%s", permissionID)
		}
		return fmt.Errorf("failed to query permission: %w", err)
	}

	// 2. 如果更新role_code/resource_type/permission_type，检查唯一性
	roleCodeChanged := permission.RoleCode != "" && permission.RoleCode != existingPerm.RoleCode
	resourceTypeChanged := permission.ResourceType != "" && permission.ResourceType != existingPerm.ResourceType
	permissionTypeChanged := permission.PermissionType != "" && permission.PermissionType != existingPerm.PermissionType

	if roleCodeChanged || resourceTypeChanged || permissionTypeChanged {
		newRoleCode := existingPerm.RoleCode
		if roleCodeChanged {
			newRoleCode = permission.RoleCode
		}
		newResourceType := existingPerm.ResourceType
		if resourceTypeChanged {
			newResourceType = permission.ResourceType
		}
		newPermissionType := existingPerm.PermissionType
		if permissionTypeChanged {
			newPermissionType = permission.PermissionType
			// 验证permission_type格式
			if newPermissionType != "R" && newPermissionType != "C" && newPermissionType != "U" && newPermissionType != "D" {
				return fmt.Errorf("invalid permission_type: %s (must be R, C, U, or D)", newPermissionType)
			}
		}

		// 如果更新role_code，验证role_code存在
		if roleCodeChanged {
			var roleExists bool
			err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE role_code = $1)`, newRoleCode).Scan(&roleExists)
			if err != nil {
				return fmt.Errorf("failed to validate role_code: %w", err)
			}
			if !roleExists {
				return fmt.Errorf("role_code not found: role_code=%s", newRoleCode)
			}
		}

		// 检查唯一性
		var existingID string
		checkQuery := `
			SELECT permission_id::text
			FROM role_permissions
			WHERE (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($1::uuid, '00000000-0000-0000-0000-000000000000'::uuid))
			  AND role_code = $2
			  AND resource_type = $3
			  AND permission_type = $4
			  AND permission_id != $5
			LIMIT 1
		`
		var tenantIDVal interface{}
		if existingPerm.TenantID.Valid {
			tenantIDVal = existingPerm.TenantID.String
		} else {
			tenantIDVal = nil
		}
		err = tx.QueryRowContext(ctx, checkQuery, tenantIDVal, newRoleCode, newResourceType, newPermissionType, permissionID).Scan(&existingID)
		if err == nil {
			return fmt.Errorf("permission already exists: role_code=%s, resource_type=%s, permission_type=%s (permission_id=%s)", newRoleCode, newResourceType, newPermissionType, existingID)
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check permission uniqueness: %w", err)
		}
	}

	// 3. 构建UPDATE语句（部分更新）
	set := []string{}
	args := []any{permissionID}
	argN := 2

	if roleCodeChanged {
		set = append(set, fmt.Sprintf("role_code = $%d", argN))
		args = append(args, permission.RoleCode)
		argN++
	}

	if resourceTypeChanged {
		set = append(set, fmt.Sprintf("resource_type = $%d", argN))
		args = append(args, permission.ResourceType)
		argN++
	}

	if permissionTypeChanged {
		set = append(set, fmt.Sprintf("permission_type = $%d", argN))
		args = append(args, permission.PermissionType)
		argN++
	}

	// assigned_only和branch_only总是可以更新（如果提供）
	if permission.AssignedOnly != existingPerm.AssignedOnly {
		set = append(set, fmt.Sprintf("assigned_only = $%d", argN))
		args = append(args, permission.AssignedOnly)
		argN++
	}

	if permission.BranchOnly != existingPerm.BranchOnly {
		set = append(set, fmt.Sprintf("branch_only = $%d", argN))
		args = append(args, permission.BranchOnly)
		argN++
	}

	if len(set) == 0 {
		// 没有需要更新的字段
		return tx.Commit()
	}

	// 4. 执行UPDATE
	updateQuery := "UPDATE role_permissions SET " + strings.Join(set, ", ") + " WHERE permission_id = $1"
	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return tx.Commit()
}

// BatchUpdatePermissions 批量更新权限
// 功能：批量更新权限，如批量修改assigned_only或branch_only
func (r *PostgresRolePermissionsRepository) BatchUpdatePermissions(ctx context.Context, updates []PermissionUpdate) (int, []error, error) {
	if len(updates) == 0 {
		return 0, nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var successCount int
	var errors []error

	for _, update := range updates {
		if update.PermissionID == "" {
			errors = append(errors, fmt.Errorf("permission_id is required"))
			continue
		}
		if update.Permission == nil {
			errors = append(errors, fmt.Errorf("permission is required"))
			continue
		}

		// 使用UpdatePermission的逻辑，但在事务内执行
		err := r.updatePermissionTx(ctx, tx, update.PermissionID, update.Permission)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to update permission %s: %w", update.PermissionID, err))
			continue
		}

		successCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, errors, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return successCount, errors, nil
}

// updatePermissionTx 在事务内更新权限（内部辅助方法）
func (r *PostgresRolePermissionsRepository) updatePermissionTx(ctx context.Context, tx *sql.Tx, permissionID string, permission *domain.RolePermission) error {
	// 查询现有权限
	var existingPerm domain.RolePermission
	err := tx.QueryRowContext(ctx, `
		SELECT permission_id::text, tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only
		FROM role_permissions
		WHERE permission_id = $1
	`, permissionID).Scan(
		&existingPerm.PermissionID,
		&existingPerm.TenantID,
		&existingPerm.RoleCode,
		&existingPerm.ResourceType,
		&existingPerm.PermissionType,
		&existingPerm.AssignedOnly,
		&existingPerm.BranchOnly,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("permission not found: permission_id=%s", permissionID)
		}
		return fmt.Errorf("failed to query permission: %w", err)
	}

	// 构建UPDATE语句（部分更新）
	set := []string{}
	args := []any{permissionID}
	argN := 2

	if permission.RoleCode != "" && permission.RoleCode != existingPerm.RoleCode {
		set = append(set, fmt.Sprintf("role_code = $%d", argN))
		args = append(args, permission.RoleCode)
		argN++
	}

	if permission.ResourceType != "" && permission.ResourceType != existingPerm.ResourceType {
		set = append(set, fmt.Sprintf("resource_type = $%d", argN))
		args = append(args, permission.ResourceType)
		argN++
	}

	if permission.PermissionType != "" && permission.PermissionType != existingPerm.PermissionType {
		// 验证permission_type格式
		if permission.PermissionType != "R" && permission.PermissionType != "C" && permission.PermissionType != "U" && permission.PermissionType != "D" {
			return fmt.Errorf("invalid permission_type: %s (must be R, C, U, or D)", permission.PermissionType)
		}
		set = append(set, fmt.Sprintf("permission_type = $%d", argN))
		args = append(args, permission.PermissionType)
		argN++
	}

	// assigned_only和branch_only总是可以更新
	if permission.AssignedOnly != existingPerm.AssignedOnly {
		set = append(set, fmt.Sprintf("assigned_only = $%d", argN))
		args = append(args, permission.AssignedOnly)
		argN++
	}

	if permission.BranchOnly != existingPerm.BranchOnly {
		set = append(set, fmt.Sprintf("branch_only = $%d", argN))
		args = append(args, permission.BranchOnly)
		argN++
	}

	if len(set) == 0 {
		return nil // 没有需要更新的字段
	}

	// 执行UPDATE
	updateQuery := "UPDATE role_permissions SET " + strings.Join(set, ", ") + " WHERE permission_id = $1"
	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// DeletePermission 删除权限
// 功能：删除权限
// 验证（Repository层）：
//   - 数据完整性：检查是否存在
func (r *PostgresRolePermissionsRepository) DeletePermission(ctx context.Context, permissionID string) error {
	if permissionID == "" {
		return fmt.Errorf("permission_id is required")
	}

	// 检查权限是否存在
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM role_permissions WHERE permission_id = $1)", permissionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("permission not found: permission_id=%s", permissionID)
	}

	// 删除权限
	_, err = r.db.ExecContext(ctx, "DELETE FROM role_permissions WHERE permission_id = $1", permissionID)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// DeletePermissionsByRole 删除某个角色的所有权限
// 功能：批量删除某个角色的所有权限（用于删除角色时清理权限）
func (r *PostgresRolePermissionsRepository) DeletePermissionsByRole(ctx context.Context, tenantID, roleCode string) error {
	if roleCode == "" {
		return fmt.Errorf("role_code is required")
	}

	var query string
	var args []any

	if tenantID != "" {
		query = `DELETE FROM role_permissions WHERE tenant_id = $1 AND role_code = $2`
		args = []any{tenantID, roleCode}
	} else {
		// 删除系统权限（System tenant）
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		query = `DELETE FROM role_permissions WHERE tenant_id = $1 AND role_code = $2`
		args = []any{systemTenantID, roleCode}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete permissions by role: %w", err)
	}

	return nil
}

