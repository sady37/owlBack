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
	if tenant.TenantName == "" {
		return "", fmt.Errorf("tenant_name is required")
	}

	// 处理默认值
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
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO tenants (tenant_name, domain, email, phone, status, metadata)
		 VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), $5, $6::jsonb)
		 RETURNING tenant_id::text`,
		tenant.TenantName,
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
func (r *PostgresTenantsRepository) UpdateTenant(ctx context.Context, tenantID string, tenant *domain.Tenant) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if tenant == nil {
		return fmt.Errorf("tenant is required")
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID}
	argIdx := 2

	if tenant.TenantName != "" {
		updates = append(updates, fmt.Sprintf("tenant_name = $%d", argIdx))
		args = append(args, tenant.TenantName)
		argIdx++
	}

	// domain, email, phone 使用 NULLIF 处理空字符串
	if tenant.Domain != "" {
		updates = append(updates, fmt.Sprintf("domain = NULLIF($%d, '')", argIdx))
		args = append(args, tenant.Domain)
		argIdx++
	}

	if tenant.Email != "" {
		updates = append(updates, fmt.Sprintf("email = NULLIF($%d, '')", argIdx))
		args = append(args, tenant.Email)
		argIdx++
	}

	if tenant.Phone != "" {
		updates = append(updates, fmt.Sprintf("phone = NULLIF($%d, '')", argIdx))
		args = append(args, tenant.Phone)
		argIdx++
	}

	if tenant.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, tenant.Status)
		argIdx++
	}

	if len(tenant.Metadata) > 0 {
		updates = append(updates, fmt.Sprintf("metadata = $%d::jsonb", argIdx))
		args = append(args, string(tenant.Metadata))
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE tenants
		SET %s
		WHERE tenant_id = $1::uuid
	`, strings.Join(updates, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)
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
