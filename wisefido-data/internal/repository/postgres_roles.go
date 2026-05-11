package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresRolesRepository 角色 Repository 实现 — owl_v2 schema 版本。
//
// v2 schema 与 v1 关键差异：
//   - 主键不变 role_id UUID；但租户字段从 tenant_id UUID 改为 tenant_id INET
//   - 无 is_active 字段（v2 设计：禁用走 user_roles 解绑而非 role 自身禁用）；本 repo 对外永远返回 IsActive=true
//   - 多了 role_name VARCHAR(100) NOT NULL（v1 只有 description）；这里把 description 同时写入 role_name 以满足 NOT NULL
//   - system roles tenant_id 为 NULL（与 v1 'System tenant UUID' 语义对应）
//
// 兼容策略：domain.Role 字段不变；TenantID 在 v2 装 prefix CIDR 字符串（'fd00:0:T::/48'），系统角色装空串。
type PostgresRolesRepository struct {
	db *sql.DB
}

func NewPostgresRolesRepository(db *sql.DB) *PostgresRolesRepository {
	return &PostgresRolesRepository{db: db}
}

var _ RolesRepository = (*PostgresRolesRepository)(nil)

// 公共 SELECT 子句：v2 schema 列直接返；role_name / description 独立字段，service 层不再做拼接。
const rolesSelectCols = `
		role_id::text,
		COALESCE(host(tenant_id) || '/48', '') AS tenant_id,
		role_code,
		COALESCE(role_name, '') AS role_name,
		COALESCE(description, '') AS description,
		is_system,
		TRUE AS is_active
`

// resolveTenantFilter 把 v1 调用者传入的 tenantID（可能是 UUID 老格式 / prefix CIDR / 空）
// 翻译成 v2 SQL 谓词（绑参形式）。
//
// v2 永远把 system roles（tenant_id IS NULL）一并返回，保证 admin 能看到全套预定义角色。
//
// 返回 (whereSnippet, args)。whereSnippet 形如 "(tenant_id IS NULL OR tenant_id = $1::INET)" 或 "tenant_id IS NULL"。
func resolveTenantFilter(tenantID *string, startArgN int) (string, []any) {
	if tenantID == nil || *tenantID == "" || isV1SystemTenantUUID(*tenantID) {
		return "tenant_id IS NULL", nil
	}
	if !looksLikeINETPrefix(*tenantID) {
		// 兼容：调用方传了非 prefix 的 UUID（自定义 tenant），v2 schema 没法直接用，回退到 system roles
		return "tenant_id IS NULL", nil
	}
	return fmt.Sprintf("(tenant_id IS NULL OR tenant_id = $%d::INET)", startArgN), []any{*tenantID}
}

func isV1SystemTenantUUID(s string) bool {
	return s == "00000000-0000-0000-0000-000000000001"
}
func looksLikeINETPrefix(s string) bool {
	return strings.Contains(s, ":") || strings.Contains(s, "/")
}

// GetRole 按 role_id 查单个角色。
func (r *PostgresRolesRepository) GetRole(ctx context.Context, roleID string) (*domain.Role, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role_id is required")
	}
	role, err := r.scanOneByQuery(ctx, `SELECT `+rolesSelectCols+` FROM roles WHERE role_id = $1`, roleID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found: role_id=%s", roleID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}
	return role, nil
}

// GetRoleByCode 按 (tenant, role_code) 查角色。tenant 为空 / system tenant UUID / 非 prefix 时只查系统角色（tenant_id IS NULL）。
func (r *PostgresRolesRepository) GetRoleByCode(ctx context.Context, tenantID *string, roleCode string) (*domain.Role, error) {
	if roleCode == "" {
		return nil, fmt.Errorf("role_code is required")
	}
	tFilter, args := resolveTenantFilter(tenantID, 1)
	args = append(args, roleCode)
	codeArgN := len(args) // 1-based
	q := `SELECT ` + rolesSelectCols + ` FROM roles WHERE ` + tFilter +
		fmt.Sprintf(` AND role_code = $%d LIMIT 1`, codeArgN)
	role, err := r.scanOneByQuery(ctx, q, args...)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found: role_code=%s", roleCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}
	return role, nil
}

// ListRoles 列出角色。filter.IsActive/IsSystem 仍兼容；v2 没有 is_active 列，IsActive 过滤被忽略。
func (r *PostgresRolesRepository) ListRoles(ctx context.Context, tenantID *string, filter RolesFilter, page, size int) ([]*domain.Role, int, error) {
	tFilter, args := resolveTenantFilter(tenantID, 1)
	where := []string{tFilter}
	argN := len(args) + 1

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(role_code ILIKE $%d OR role_name ILIKE $%d OR description ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+filter.Search+"%")
		argN++
	}
	if filter.IsSystem != nil {
		where = append(where, fmt.Sprintf("is_system = $%d", argN))
		args = append(args, *filter.IsSystem)
		argN++
	}
	// filter.IsActive: v2 schema 无此列，忽略

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM roles `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	listQ := `SELECT ` + rolesSelectCols + ` FROM roles ` + whereClause +
		fmt.Sprintf(` ORDER BY is_system DESC, role_code ASC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := []*domain.Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return roles, total, nil
}

// CreateRole 新建角色。tenantID 必须是 INET prefix CIDR；空/UUID 视作系统角色（tenant_id=NULL）。
//
// v2 强制 role_name NOT NULL；这里以 description 第一行作 role_name fallback（v1 description 业务约定就是头一行 = 角色名）。
func (r *PostgresRolesRepository) CreateRole(ctx context.Context, tenantID string, role *domain.Role) (string, error) {
	if role == nil {
		return "", fmt.Errorf("role is required")
	}
	if role.RoleCode == "" {
		return "", fmt.Errorf("role_code is required")
	}
	if role.Description == "" {
		return "", fmt.Errorf("description is required")
	}

	var tenantPrefix any
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		tenantPrefix = tenantID
	} else {
		tenantPrefix = nil // system role
	}

	// 唯一性：(tenant_id, role_code)
	dup, err := r.db.QueryContext(ctx, `
		SELECT role_id::text FROM roles
		 WHERE COALESCE(tenant_id::text, '') = COALESCE($1::INET::text, '')
		   AND role_code = $2 LIMIT 1
	`, tenantPrefix, role.RoleCode)
	if err != nil {
		return "", fmt.Errorf("failed to check role uniqueness: %w", err)
	}
	if dup.Next() {
		var existingRoleID string
		_ = dup.Scan(&existingRoleID)
		dup.Close()
		return "", fmt.Errorf("role already exists: role_code=%s (role_id=%s)", role.RoleCode, existingRoleID)
	}
	dup.Close()

	// role_name 来自 domain.RoleName（v2 独立列）；若空则回退到 RoleCode
	roleName := role.RoleName
	if roleName == "" {
		roleName = role.RoleCode
	}

	var roleID string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO roles (tenant_id, role_code, role_name, description, is_system)
		VALUES ($1::INET, $2, $3, $4, $5)
		RETURNING role_id::text
	`, tenantPrefix, role.RoleCode, roleName, role.Description, role.IsSystem).Scan(&roleID)
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}
	return roleID, nil
}

// UpdateRole 更新角色（部分更新）。v2 schema 无 is_active 列；IsActive 输入被忽略。
func (r *PostgresRolesRepository) UpdateRole(ctx context.Context, roleID string, role *domain.Role) error {
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}
	if role == nil {
		return fmt.Errorf("role is required")
	}

	// 取现有
	existing, err := r.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	// is_system 不可改
	if role.IsSystem != existing.IsSystem {
		return fmt.Errorf("cannot change is_system: role_id=%s (current=%v, new=%v)", roleID, existing.IsSystem, role.IsSystem)
	}

	set := []string{}
	args := []any{roleID}
	argN := 2

	if role.RoleCode != "" && role.RoleCode != existing.RoleCode {
		set = append(set, fmt.Sprintf("role_code = $%d", argN))
		args = append(args, role.RoleCode)
		argN++
	}
	if role.RoleName != "" && role.RoleName != existing.RoleName {
		set = append(set, fmt.Sprintf("role_name = $%d", argN))
		args = append(args, role.RoleName)
		argN++
	}
	if role.Description != existing.Description {
		set = append(set, fmt.Sprintf("description = $%d", argN))
		args = append(args, role.Description)
		argN++
	}
	// IsActive: v2 schema 不再支持；忽略

	if len(set) == 0 {
		return nil // nothing to update
	}
	updateQuery := "UPDATE roles SET " + strings.Join(set, ", ") + " WHERE role_id = $1"
	if _, err := r.db.ExecContext(ctx, updateQuery, args...); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// DeleteRole 删角色（系统角色由 service 层挡）。
func (r *PostgresRolesRepository) DeleteRole(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}
	res, err := r.db.ExecContext(ctx, "DELETE FROM roles WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("role not found: role_id=%s", roleID)
	}
	return nil
}

// =============================================================================
// helpers
// =============================================================================

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRole(rs rowScanner) (*domain.Role, error) {
	var role domain.Role
	if err := rs.Scan(
		&role.RoleID,
		&role.TenantID, // sql.NullString — '' 时表示 system role
		&role.RoleCode,
		&role.RoleName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive, // 总是 TRUE（COALESCE 兜底）
	); err != nil {
		return nil, fmt.Errorf("failed to scan role: %w", err)
	}
	return &role, nil
}

func (r *PostgresRolesRepository) scanOneByQuery(ctx context.Context, q string, args ...any) (*domain.Role, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	return scanRole(rows)
}

// firstLine 返回字符串第一行（不含换行符），用于把 v1 description 头一行映射到 v2 role_name。
func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}
