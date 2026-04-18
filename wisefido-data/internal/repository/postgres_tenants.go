package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresTenantsRepository 租户Repository实现（强类型版本）
// 实现TenantsRepository接口，使用domain.Tenant领域模型
type PostgresTenantsRepository struct {
	db *sql.DB
}

// NewPostgresTenantsRepository 创建租户Repository
func NewPostgresTenantsRepository(db *sql.DB) *PostgresTenantsRepository {
	return &PostgresTenantsRepository{db: db}
}

// 确保实现了接口
var _ TenantsRepository = (*PostgresTenantsRepository)(nil)

// GetTenant 根据tenant_id获取租户信息
func (r *PostgresTenantsRepository) GetTenant(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	query := `
		SELECT 
			tenant_id::text,
			COALESCE(tenant_type, 'organization') as tenant_type,
			tenant_name,
			COALESCE(domain, '') as domain,
			COALESCE(email, '') as email,
			COALESCE(phone, '') as phone,
			COALESCE(status, 'active') as status,
			COALESCE(metadata, '{}'::jsonb) as metadata
		FROM tenants
		WHERE tenant_id = $1::uuid
	`

	var tenant domain.Tenant
	var metadataRaw json.RawMessage
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.TenantID,
		&tenant.TenantType,
		&tenant.TenantName,
		&tenant.Domain,
		&tenant.Email,
		&tenant.Phone,
		&tenant.Status,
		&metadataRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Metadata = metadataRaw
	return &tenant, nil
}

// GetTenantByDomain 根据domain获取租户信息（用于域名路由）
func (r *PostgresTenantsRepository) GetTenantByDomain(ctx context.Context, domainName string) (*domain.Tenant, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain is required")
	}

	query := `
		SELECT 
			tenant_id::text,
			COALESCE(tenant_type, 'organization') as tenant_type,
			tenant_name,
			COALESCE(domain, '') as domain,
			COALESCE(email, '') as email,
			COALESCE(phone, '') as phone,
			COALESCE(status, 'active') as status,
			COALESCE(metadata, '{}'::jsonb) as metadata
		FROM tenants
		WHERE domain = $1
	`

	var tenant domain.Tenant
	var metadataRaw json.RawMessage
	err := r.db.QueryRowContext(ctx, query, domainName).Scan(
		&tenant.TenantID,
		&tenant.TenantType,
		&tenant.TenantName,
		&tenant.Domain,
		&tenant.Email,
		&tenant.Phone,
		&tenant.Status,
		&metadataRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	tenant.Metadata = metadataRaw
	return &tenant, nil
}

// GetTenantIDByName 按 tenant_name 精确匹配（trim + 不区分大小写），唯一则返回 tenant_id
func (r *PostgresTenantsRepository) GetTenantIDByName(ctx context.Context, tenantName string) (string, error) {
	name := strings.TrimSpace(tenantName)
	if name == "" {
		return "", fmt.Errorf("tenant_name is empty")
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id::text FROM tenants
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

// ListTenants 查询租户列表（支持分页、过滤、搜索）
func (r *PostgresTenantsRepository) ListTenants(ctx context.Context, filter TenantFilters, page, size int) ([]*domain.Tenant, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	// 构建WHERE条件
	where := []string{}
	args := []any{}
	argIdx := 1

	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("tenant_name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tenants %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// 查询列表（带分页）
	query := fmt.Sprintf(`
		SELECT 
			tenant_id::text,
			COALESCE(tenant_type, 'organization') as tenant_type,
			tenant_name,
			COALESCE(domain, '') as domain,
			COALESCE(email, '') as email,
			COALESCE(phone, '') as phone,
			COALESCE(status, 'active') as status,
			COALESCE(metadata, '{}'::jsonb) as metadata
		FROM tenants
		%s
		ORDER BY tenant_name
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		var tenant domain.Tenant
		var metadataRaw json.RawMessage
		err := rows.Scan(
			&tenant.TenantID,
			&tenant.TenantType,
			&tenant.TenantName,
			&tenant.Domain,
			&tenant.Email,
			&tenant.Phone,
			&tenant.Status,
			&metadataRaw,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.Metadata = metadataRaw
		tenants = append(tenants, &tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenants, total, nil
}

// CreateTenant 创建新租户
func (r *PostgresTenantsRepository) CreateTenant(ctx context.Context, tenant *domain.Tenant) (string, error) {
	if tenant == nil {
		return "", fmt.Errorf("tenant is required")
	}
	trimmed := strings.TrimSpace(tenant.TenantName)
	if trimmed == "" {
		return "", fmt.Errorf("tenant_name is required")
	}
	var existingID string
	err := r.db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM tenants WHERE LOWER(TRIM(tenant_name)) = LOWER($1) LIMIT 1`,
		trimmed,
	).Scan(&existingID)
	if err == nil {
		return "", fmt.Errorf("tenant_name already exists: %q", trimmed)
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("check tenant_name: %w", err)
	}

	// 处理默认值
	tenantType := tenant.TenantType
	if tenantType == "" {
		tenantType = "organization"
	}
	status := tenant.Status
	if status == "" {
		status = "active"
	}

	// 处理metadata
	metadataArg := "{}"
	if len(tenant.Metadata) > 0 {
		metadataArg = string(tenant.Metadata)
	}

	// 处理可空字段（使用NULLIF将空字符串转为NULL）
	var tenantID string
	err = r.db.QueryRowContext(ctx,
		`INSERT INTO tenants (tenant_type, tenant_name, domain, email, phone, status, metadata)
		 VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, $7::jsonb)
		 RETURNING tenant_id::text`,
		tenantType,
		trimmed,
		tenant.Domain,
		tenant.Email,
		tenant.Phone,
		status,
		metadataArg,
	).Scan(&tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenantID, nil
}

// UpdateTenant 更新租户信息
// 注意：调用方（handler）已加载 existing tenant 并合并 payload 字段，
// 因此这里无条件更新所有核心字段（domain/email/phone 空字符串 → NULL）。
func (r *PostgresTenantsRepository) UpdateTenant(ctx context.Context, tenantID string, tenant *domain.Tenant) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if tenant == nil {
		return fmt.Errorf("tenant is required")
	}

	// tenant_name 非空校验 + 唯一性检查
	trimmed := strings.TrimSpace(tenant.TenantName)
	if trimmed == "" {
		return fmt.Errorf("tenant_name is required")
	}
	var conflictID string
	err := r.db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM tenants WHERE LOWER(TRIM(tenant_name)) = LOWER($1) AND tenant_id <> $2::uuid`,
		trimmed, tenantID,
	).Scan(&conflictID)
	if err == nil {
		return fmt.Errorf("tenant_name already exists: %q", trimmed)
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("check tenant_name: %w", err)
	}

	// 处理 metadata
	metadataArg := "{}"
	if len(tenant.Metadata) > 0 {
		metadataArg = string(tenant.Metadata)
	}

	// 处理 status 默认值
	status := tenant.Status
	if status == "" {
		status = "active"
	}

	// 处理 tenant_type 默认值
	tenantType := tenant.TenantType
	if tenantType == "" {
		tenantType = "organization"
	}

	// 无条件更新所有核心字段，空字符串通过 NULLIF 转为 NULL
	query := `
		UPDATE tenants
		SET tenant_type = $2,
		    tenant_name = $3,
		    domain = NULLIF($4, ''),
		    email = NULLIF($5, ''),
		    phone = NULLIF($6, ''),
		    status = $7,
		    metadata = $8::jsonb
		WHERE tenant_id = $1::uuid
	`

	result, err := r.db.ExecContext(ctx, query,
		tenantID,
		tenantType,
		trimmed,
		tenant.Domain,
		tenant.Email,
		tenant.Phone,
		status,
		metadataArg,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found: tenant_id '%s' does not exist", tenantID)
	}

	return nil
}

// SetTenantStatus 更新租户状态
func (r *PostgresTenantsRepository) SetTenantStatus(ctx context.Context, tenantID string, status string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if status == "" {
		return fmt.Errorf("status is required")
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE tenants SET status = $2 WHERE tenant_id = $1::uuid`,
		tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("failed to set tenant status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found: tenant_id '%s' does not exist", tenantID)
	}

	return nil
}

// DeleteTenant 删除租户（软删除：设置status='deleted'）
func (r *PostgresTenantsRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	// 软删除：设置status='deleted'
	return r.SetTenantStatus(ctx, tenantID, "deleted")
}
