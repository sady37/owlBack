package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"wisefido-data/internal/domain"
)

// MemoryTenantsRepo supports admin tenants management when DB is disabled.
// NOTE: This is "platform-level" data (not per-tenant).
// 实现 TenantsRepository 接口（新接口，使用强类型 domain.Tenant）
type MemoryTenantsRepo struct {
	mu      sync.RWMutex
	tenants map[string]*domain.Tenant // tenantID -> domain.Tenant
}

func NewMemoryTenantsRepo() *MemoryTenantsRepo {
	return &MemoryTenantsRepo{
		tenants: map[string]*domain.Tenant{},
	}
}

// ========== 实现 TenantsRepository 接口（新接口）==========

// GetTenant 根据tenant_id获取租户信息
func (r *MemoryTenantsRepo) GetTenant(_ context.Context, tenantID string) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tenants[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found")
	}
	// 返回副本
	result := *t
	return &result, nil
}

// GetTenantByDomain 根据domain获取租户信息
func (r *MemoryTenantsRepo) GetTenantByDomain(_ context.Context, domain string) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.tenants {
		if t.Domain == domain {
			// 返回副本
			result := *t
			return &result, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}

// ListTenants 查询租户列表（支持分页、过滤、搜索）
func (r *MemoryTenantsRepo) ListTenants(_ context.Context, filter TenantFilters, page, size int) ([]*domain.Tenant, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*domain.Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		// 状态过滤
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		// 搜索过滤（tenant_name 模糊匹配）
		if filter.Search != "" {
			if !strings.Contains(strings.ToLower(t.TenantName), strings.ToLower(filter.Search)) {
				continue
			}
		}
		// 返回副本
		result := *t
		all = append(all, &result)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].TenantName < all[j].TenantName
	})

	total := len(all)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// CreateTenant 创建新租户
func (r *MemoryTenantsRepo) CreateTenant(_ context.Context, tenant *domain.Tenant) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tenant.TenantID == "" {
		tenant.TenantID = uuid.NewString()
	}
	if tenant.Status == "" {
		tenant.Status = "active"
	}

	// 检查 domain 唯一性
	if tenant.Domain != "" {
		for _, t := range r.tenants {
			if t.Domain == tenant.Domain && t.TenantID != tenant.TenantID {
				return "", fmt.Errorf("domain already exists")
			}
		}
	}

	// 创建副本
	result := *tenant
	r.tenants[tenant.TenantID] = &result
	return tenant.TenantID, nil
}

// UpdateTenant 更新租户信息
func (r *MemoryTenantsRepo) UpdateTenant(_ context.Context, tenantID string, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found")
	}

	// 检查 domain 唯一性（如果 domain 被修改）
	if tenant.Domain != "" && tenant.Domain != t.Domain {
		for id, existing := range r.tenants {
			if existing.Domain == tenant.Domain && id != tenantID {
				return fmt.Errorf("domain already exists")
			}
		}
	}

	// 更新字段
	if tenant.TenantName != "" {
		t.TenantName = tenant.TenantName
	}
	if tenant.Domain != "" {
		t.Domain = tenant.Domain
	}
	if tenant.Email != "" {
		t.Email = tenant.Email
	}
	if tenant.Phone != "" {
		t.Phone = tenant.Phone
	}
	if tenant.Status != "" {
		t.Status = tenant.Status
	}
	if tenant.Metadata != nil {
		t.Metadata = tenant.Metadata
	}

	return nil
}

// SetTenantStatus 更新租户状态
func (r *MemoryTenantsRepo) SetTenantStatus(_ context.Context, tenantID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found")
	}
	t.Status = status
	return nil
}

// DeleteTenant 删除租户（软删除）
func (r *MemoryTenantsRepo) DeleteTenant(_ context.Context, tenantID string) error {
	return r.SetTenantStatus(context.Background(), tenantID, "deleted")
}

// ========== 兼容旧接口 TenantsRepo（向后兼容）==========
// 注意：这些方法保留用于向后兼容 StubHandler 和 TenantsHandler（旧版本）

// ListTenantsOld 旧接口方法（兼容 TenantsRepo）
func (r *MemoryTenantsRepo) ListTenantsOld(_ context.Context, status string, page, size int) ([]Tenant, int, error) {
	filter := TenantFilters{Status: status}
	tenants, total, err := r.ListTenants(context.Background(), filter, page, size)
	if err != nil {
		return nil, 0, err
	}
	// 转换为旧类型
	result := make([]Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = Tenant{
			TenantID:   t.TenantID,
			TenantName: t.TenantName,
			Domain:     t.Domain,
			Email:      t.Email,
			Phone:      t.Phone,
			Status:     t.Status,
			Metadata:   t.Metadata,
		}
	}
	return result, total, nil
}

// CreateTenantOld 旧接口方法（兼容 TenantsRepo）
func (r *MemoryTenantsRepo) CreateTenantOld(_ context.Context, payload map[string]any) (*Tenant, error) {
	tenant := &domain.Tenant{
		TenantID:   uuid.NewString(),
		TenantName: getString(payload, "tenant_name"),
		Domain:     getString(payload, "domain"),
		Email:      getString(payload, "email"),
		Phone:      getString(payload, "phone"),
		Status:     getStringOrDefault(payload, "status", "active"),
	}
	if v, ok := payload["metadata"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			tenant.Metadata = b
		}
	}
	tenantID, err := r.CreateTenant(context.Background(), tenant)
	if err != nil {
		return nil, err
	}
	// 转换为旧类型
	t, _ := r.GetTenant(context.Background(), tenantID)
	return &Tenant{
		TenantID:   t.TenantID,
		TenantName: t.TenantName,
		Domain:     t.Domain,
		Email:      t.Email,
		Phone:      t.Phone,
		Status:     t.Status,
		Metadata:   t.Metadata,
	}, nil
}

// UpdateTenantOld 旧接口方法（兼容 TenantsRepo）
func (r *MemoryTenantsRepo) UpdateTenantOld(_ context.Context, tenantID string, payload map[string]any) (*Tenant, error) {
	existing, err := r.GetTenant(context.Background(), tenantID)
	if err != nil {
		// create-on-update for dev convenience
		existing = &domain.Tenant{
			TenantID: tenantID,
			Status:   "active",
		}
	}
	// 更新字段
	if v := getString(payload, "tenant_name"); v != "" {
		existing.TenantName = v
	}
	if _, ok := payload["domain"]; ok {
		existing.Domain = getString(payload, "domain")
	}
	if _, ok := payload["email"]; ok {
		existing.Email = getString(payload, "email")
	}
	if _, ok := payload["phone"]; ok {
		existing.Phone = getString(payload, "phone")
	}
	if v := getString(payload, "status"); v != "" {
		existing.Status = v
	}
	if v, ok := payload["metadata"]; ok {
		if v == nil {
			existing.Metadata = nil
		} else if b, err := json.Marshal(v); err == nil {
			existing.Metadata = b
		}
	}
	err = r.UpdateTenant(context.Background(), tenantID, existing)
	if err != nil {
		return nil, err
	}
	// 转换为旧类型
	t, _ := r.GetTenant(context.Background(), tenantID)
	return &Tenant{
		TenantID:   t.TenantID,
		TenantName: t.TenantName,
		Domain:     t.Domain,
		Email:      t.Email,
		Phone:      t.Phone,
		Status:     t.Status,
		Metadata:   t.Metadata,
	}, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringOrDefault(m map[string]any, key, defaultValue string) string {
	if v := getString(m, key); v != "" {
		return v
	}
	return defaultValue
}
