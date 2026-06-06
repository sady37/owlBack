package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresTenantsRepository 租户 Repository — owl_v2 schema 版本。
//
// v2 schema 与 v1 关键差异（不向后兼容）：
//   - 主键从 tenant_id UUID 改为 tenant_id INET (/48 CIDR)
//   - 列删除：tenant_type / domain / status / metadata
//   - 列改名：email→contact_email / phone→contact_phone
//   - 列新增：tenant_slot SMALLINT (uint16, 全局自增) / timezone / contact_name / created_at / updated_at
//   - 删除走 hard DELETE（FK CASCADE 清 branches/units/...）
//
// 兼容策略：domain.Tenant 字段不变；TenantID 在 v2 装 tenant_id CIDR 字符串；
// TenantType/Status/Domain/Metadata 字段对外伪装值（'organization' / 'active' / '' / '{}'）。
type PostgresTenantsRepository struct {
	db *sql.DB
}

func NewPostgresTenantsRepository(db *sql.DB) *PostgresTenantsRepository {
	return &PostgresTenantsRepository{db: db}
}

var _ TenantsRepository = (*PostgresTenantsRepository)(nil)

// 公共 SELECT 子句：v2 schema 列适配回 v1 domain.Tenant 字段
const tenantsSelectCols = `
		host(tenant_id) || '/48' AS tenant_id,
		'organization' AS tenant_type,
		tenant_name,
		COALESCE(kind, 'B2B') AS kind,
		COALESCE(timezone, 'America/Los_Angeles') AS timezone,
		'' AS domain,
		COALESCE(contact_email, '') AS email,
		COALESCE(contact_phone, '') AS phone,
		COALESCE(status, 'active') AS status,
		'{}'::jsonb AS metadata
`

// GetTenant 按 tenant_id (实为 tenant_id CIDR) 查询租户。
func (r *PostgresTenantsRepository) GetTenant(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if !looksLikeINETPrefix(tenantID) {
		return nil, fmt.Errorf("tenant not found: %q is not a v2 INET prefix", tenantID)
	}

	q := `SELECT ` + tenantsSelectCols + ` FROM tenants WHERE tenant_id = $1::INET`
	return r.scanOneTenant(ctx, q, tenantID)
}

// GetTenantByDomain v2 schema 删除了 domain 列；保留接口返回 not found 兜底。
func (r *PostgresTenantsRepository) GetTenantByDomain(ctx context.Context, domainName string) (*domain.Tenant, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain is required")
	}
	return nil, fmt.Errorf("tenant not found: domain lookup deprecated in owl_v2")
}

// GetTenantIDByName 按 tenant_name 精确匹配（trim + 不区分大小写）；唯一则返回 tenant_id CIDR。
func (r *PostgresTenantsRepository) GetTenantIDByName(ctx context.Context, tenantName string) (string, error) {
	name := strings.TrimSpace(tenantName)
	if name == "" {
		return "", fmt.Errorf("tenant_name is empty")
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT host(tenant_id) || '/48' FROM tenants
		 WHERE LOWER(TRIM(tenant_name)) = LOWER($1)`, name)
	if err != nil {
		return "", fmt.Errorf("lookup tenant by name: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	switch len(ids) {
	case 0:
		return "", fmt.Errorf("tenant not found for name %q", name)
	case 1:
		return ids[0], nil
	default:
		return "", fmt.Errorf("ambiguous tenant_name %q (%d rows)", name, len(ids))
	}
}

// ListTenants v2: 支持 search + status 过滤 + 分页。
//
// 行为：默认列出 active+suspended（不显示 deleted 软删除的）；filter.Status 给定具体值则按此过滤。
func (r *PostgresTenantsRepository) ListTenants(ctx context.Context, filter TenantFilters, page, size int) ([]*domain.Tenant, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	where := []string{}
	args := []any{}
	argIdx := 1

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("tenant_name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("COALESCE(status, 'active') = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	} else {
		// 默认隐藏 deleted（软删除的）
		where = append(where, "COALESCE(status, 'active') <> 'deleted'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	listQ := `SELECT ` + tenantsSelectCols + ` FROM tenants ` + whereClause +
		fmt.Sprintf(` ORDER BY tenant_slot ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tenants: %w", err)
	}
	return tenants, total, nil
}

// CreateTenant 新建租户：自动分配 tenant_slot = MAX+1，派生 tenant_id。
//
// 并发安全：用 advisory lock + 事务（MAX+1 模式无 lock 时会冲突）。
func (r *PostgresTenantsRepository) CreateTenant(ctx context.Context, tenant *domain.Tenant) (string, error) {
	if tenant == nil {
		return "", fmt.Errorf("tenant is required")
	}
	trimmed := strings.TrimSpace(tenant.TenantName)
	if trimmed == "" {
		return "", fmt.Errorf("tenant_name is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// advisory lock：'owl_v2.tenants' 哈希 key（任意常量，让所有 tenant 创建串行）
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('owl_v2.tenants.alloc'))`); err != nil {
		return "", fmt.Errorf("advisory lock: %w", err)
	}

	// 唯一性 (tenant_name)
	var dupID string
	err = tx.QueryRowContext(ctx,
		`SELECT host(tenant_id) || '/48' FROM tenants WHERE LOWER(TRIM(tenant_name)) = LOWER($1) LIMIT 1`,
		trimmed,
	).Scan(&dupID)
	if err == nil {
		return "", fmt.Errorf("tenant_name already exists: %q", trimmed)
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("check tenant_name: %w", err)
	}

	// tenant_slot 1..65534（0=unbound、0xFFFF=wildcard 保留）。MAX+1 优先，撞顶回收空号。
	nextSlot, err := AllocSlotReclaim(ctx, tx, 65534,
		`SELECT COALESCE(MAX(tenant_slot), 0) + 1 FROM tenants`,
		`SELECT tenant_slot FROM tenants`)
	if err != nil {
		return "", fmt.Errorf("alloc tenant_slot: %w", err)
	}
	prefixStr := fmt.Sprintf("fd00:0:%x::/48", nextSlot)

	// timezone 默认 UTC（v2 schema NOT NULL DEFAULT 'UTC'）
	timezone := "UTC"

	status := tenant.Status
	if status == "" {
		status = "active"
	}

	var newPrefix string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tenants (tenant_id, tenant_slot, tenant_name, timezone, contact_email, contact_phone, status)
		VALUES ($1::INET, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7)
		RETURNING host(tenant_id) || '/48'
	`, prefixStr, nextSlot, trimmed, timezone, tenant.Email, tenant.Phone, status).Scan(&newPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return newPrefix, nil
}

// UpdateTenant 更新 tenant_name / contact_email / contact_phone。其它 v1 字段 (tenant_type / domain / status / metadata) 在 v2 不存在，忽略。
func (r *PostgresTenantsRepository) UpdateTenant(ctx context.Context, tenantID string, tenant *domain.Tenant) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !looksLikeINETPrefix(tenantID) {
		return fmt.Errorf("tenant not found: %q is not a v2 INET prefix", tenantID)
	}
	if tenant == nil {
		return fmt.Errorf("tenant is required")
	}
	trimmed := strings.TrimSpace(tenant.TenantName)
	if trimmed == "" {
		return fmt.Errorf("tenant_name is required")
	}

	// 唯一性 (tenant_name) 排除自己
	var conflict string
	err := r.db.QueryRowContext(ctx,
		`SELECT host(tenant_id) || '/48' FROM tenants WHERE LOWER(TRIM(tenant_name)) = LOWER($1) AND tenant_id <> $2::INET`,
		trimmed, tenantID,
	).Scan(&conflict)
	if err == nil {
		return fmt.Errorf("tenant_name already exists: %q", trimmed)
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("check tenant_name: %w", err)
	}

	// status 仅在非空时更新（前端 UpdateTenant 经常不传 status；专门走 SetTenantStatus）
	q := `
		UPDATE tenants
		   SET tenant_name = $2,
		       contact_email = NULLIF($3, ''),
		       contact_phone = NULLIF($4, ''),
		       updated_at = NOW()
	`
	args := []any{tenantID, trimmed, tenant.Email, tenant.Phone}
	if tenant.Status != "" {
		q += `, status = $5`
		args = append(args, tenant.Status)
	}
	q += ` WHERE tenant_id = $1::INET`
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("tenant not found: tenant_id '%s' does not exist", tenantID)
	}
	return nil
}

// SetTenantStatus 真正更新 status 列：active / suspended / deleted。
//
// HIPAA + 业务约束：tenant 不允许硬删（下游 audit_log/event_log/resident_phi 等都引用 tenant_id）。
// 'deleted' 也是软删（设 status='deleted' 隐藏 from list；历史数据保留）。
func (r *PostgresTenantsRepository) SetTenantStatus(ctx context.Context, tenantID string, status string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !looksLikeINETPrefix(tenantID) {
		return fmt.Errorf("tenant not found: %q is not a v2 INET prefix", tenantID)
	}
	switch status {
	case "active", "suspended", "deleted":
	default:
		return fmt.Errorf("invalid status %q (allowed: active/suspended/deleted)", status)
	}
	// 系统内置 tenant (system / trash) 仅允许保持 active；不允许 disable / delete / suspend
	if status != "active" {
		protected, err := r.isProtectedSystemTenant(ctx, tenantID)
		if err != nil {
			return fmt.Errorf("check protected tenant: %w", err)
		}
		if protected {
			return fmt.Errorf("system built-in tenant cannot be disabled or deleted")
		}
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE tenants SET status = $2, updated_at = NOW() WHERE tenant_id = $1::INET`,
		tenantID, status)
	if err != nil {
		return fmt.Errorf("failed to set tenant status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("tenant not found: tenant_id '%s' does not exist", tenantID)
	}
	return nil
}

// DeleteTenant 智能删除（与 branch 同约定）：
//   - 空 tenant（无业务数据）→ 真物理 DELETE，连带清 default branch + bootstrap admin user
//   - 非空 tenant → 拒绝（要求先清 branch/resident/device 等子数据）；
//     HIPAA 数据保留靠子层数据本身，不靠 tenant 软删
//
// "空"判定（business data only，配置不算）：
//   - branches 排除 'default' 占位
//   - users 不算（bootstrap admin 自动创建，删 tenant 时一并清）
//   - residents / devices 算
func (r *PostgresTenantsRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !looksLikeINETPrefix(tenantID) {
		return fmt.Errorf("tenant not found: %q is not a v2 INET prefix", tenantID)
	}
	// 系统内置 tenant (system / trash) 不允许 delete
	protected, err := r.isProtectedSystemTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("check protected tenant: %w", err)
	}
	if protected {
		return fmt.Errorf("system built-in tenant cannot be deleted")
	}
	empty, err := r.isTenantEmpty(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("check tenant empty: %w", err)
	}
	if !empty {
		return fmt.Errorf("tenant has children: delete branches/residents/devices first")
	}
	return r.hardDeleteTenant(ctx, tenantID)
}

// hardDeleteTenant transaction 内清 default branch + bootstrap users + tenant 行。
// v2 schema 没设 ON DELETE CASCADE FK（用 prefix-match 而非 FK），需要应用层级联。
func (r *PostgresTenantsRepository) hardDeleteTenant(ctx context.Context, tenantPrefix string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 清 user_roles → users（bootstrap admin 自动创建的，删 tenant 时一并清）
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_roles
		 WHERE user_id IN (SELECT user_id FROM users WHERE tenant_id = $1::INET)`,
		tenantPrefix); err != nil {
		return fmt.Errorf("hard delete user_roles: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM users WHERE tenant_id = $1::INET`, tenantPrefix); err != nil {
		return fmt.Errorf("hard delete users: %w", err)
	}

	// 清掉 v1 自动创建的 default branch（如果有）
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM branches WHERE branch_id <<= $1::INET AND branch_name = 'default'`,
		tenantPrefix); err != nil {
		return fmt.Errorf("hard delete default branch: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		`DELETE FROM tenants WHERE tenant_id = $1::INET`, tenantPrefix)
	if err != nil {
		return fmt.Errorf("hard delete tenant: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("tenant not found: tenant_id '%s' does not exist", tenantPrefix)
	}
	return tx.Commit()
}

// isProtectedSystemTenant 系统内置 tenant（slot 1=system / 2=trash）禁止 delete/disable。
// 通过 tenant_slot 判定（直接查 DB）；NULL/不存在返回 false（保守）。
func (r *PostgresTenantsRepository) isProtectedSystemTenant(ctx context.Context, tenantPrefix string) (bool, error) {
	var slot int
	err := r.db.QueryRowContext(ctx,
		`SELECT tenant_slot FROM tenants WHERE tenant_id = $1::INET`, tenantPrefix).Scan(&slot)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return slot == 1 || slot == 2, nil
}

// isTenantEmpty 判定 tenant 是否无业务数据（business data only；配置/占位不算）：
//   - branches: 仅算非 'default' 的（'default' 是 CreateTenant 自动建的占位）
//   - residents / devices: 算（真实业务）
//   - users: 不算（bootstrap admin 是 CreateTenant 自动建的；hardDelete 一并清）
func (r *PostgresTenantsRepository) isTenantEmpty(ctx context.Context, tenantPrefix string) (bool, error) {
	var hasData bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM branches  WHERE branch_id <<= $1::INET
		    AND COALESCE(branch_name, '') <> 'default'
		  UNION ALL
		  SELECT 1 FROM residents WHERE resident_id IS NOT NULL AND resident_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM devices   WHERE device_addr IS NOT NULL AND device_addr <<= $1::INET
		)
	`, tenantPrefix).Scan(&hasData)
	if err != nil {
		return false, err
	}
	return !hasData, nil
}

// =============================================================================
// helpers
// =============================================================================

func (r *PostgresTenantsRepository) scanOneTenant(ctx context.Context, q string, args ...any) (*domain.Tenant, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tenant not found")
	}
	return scanTenant(rows)
}

func scanTenant(rs rowScanner) (*domain.Tenant, error) {
	var tenant domain.Tenant
	var metadataRaw json.RawMessage
	if err := rs.Scan(
		&tenant.TenantID,
		&tenant.TenantType,
		&tenant.TenantName,
		&tenant.Kind,
		&tenant.Timezone,
		&tenant.Domain,
		&tenant.Email,
		&tenant.Phone,
		&tenant.Status,
		&metadataRaw,
	); err != nil {
		return nil, fmt.Errorf("failed to scan tenant: %w", err)
	}
	tenant.Metadata = metadataRaw
	return &tenant, nil
}
