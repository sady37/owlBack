package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// TenantsRepository 租户Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type TenantsRepository interface {
	// ========== 查询（单个）==========
	// GetTenant 根据tenant_id获取租户信息
	GetTenant(ctx context.Context, tenantID string) (*domain.Tenant, error)

	// GetTenantByDomain 根据domain获取租户信息（用于域名路由）
	// 注意：domain有唯一索引，支持此查询
	GetTenantByDomain(ctx context.Context, domain string) (*domain.Tenant, error)

	// ========== 查询（列表）==========
	// ListTenants 查询租户列表（支持分页、过滤、搜索）
	// 过滤条件：status（active/suspended/deleted）
	// 搜索条件：tenant_name（模糊匹配）
	ListTenants(ctx context.Context, filter TenantFilters, page, size int) ([]*domain.Tenant, int, error)

	// ========== 创建 ==========
	// CreateTenant 创建新租户
	// 注意：domain唯一性约束由数据库保证
	CreateTenant(ctx context.Context, tenant *domain.Tenant) (string, error)

	// ========== 更新 ==========
	// UpdateTenant 更新租户信息
	// 注意：domain唯一性约束由数据库保证
	UpdateTenant(ctx context.Context, tenantID string, tenant *domain.Tenant) error

	// SetTenantStatus 更新租户状态（active/suspended/deleted）
	// 单独方法，便于状态管理
	SetTenantStatus(ctx context.Context, tenantID string, status string) error

	// ========== 删除 ==========
	// DeleteTenant 删除租户（软删除：设置status='deleted'）
	DeleteTenant(ctx context.Context, tenantID string) error
}

// TenantFilters 租户查询过滤器
type TenantFilters struct {
	Status string // 可选，按status过滤（active/suspended/deleted）
	Search string // 可选，按tenant_name搜索（模糊匹配）
}

