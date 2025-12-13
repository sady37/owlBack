package repository

import (
	"context"
	"encoding/json"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// MemoryTenantsRepo supports admin tenants management when DB is disabled.
// NOTE: This is "platform-level" data (not per-tenant).
type MemoryTenantsRepo struct {
	mu      sync.RWMutex
	tenants map[string]Tenant // tenantID -> Tenant
}

func NewMemoryTenantsRepo() *MemoryTenantsRepo {
	return &MemoryTenantsRepo{
		tenants: map[string]Tenant{},
	}
}

func (r *MemoryTenantsRepo) ListTenants(_ context.Context, status string, page, size int) ([]Tenant, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		if status != "" && t.Status != status {
			continue
		}
		all = append(all, t)
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

func (r *MemoryTenantsRepo) CreateTenant(_ context.Context, payload map[string]any) (*Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name, _ := payload["tenant_name"].(string)
	domain, _ := payload["domain"].(string)
	email, _ := payload["email"].(string)
	phone, _ := payload["phone"].(string)
	status, _ := payload["status"].(string)
	if status == "" {
		status = "active"
	}

	var metadata json.RawMessage
	if v, ok := payload["metadata"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			metadata = b
		}
	}

	id := uuid.NewString()
	t := Tenant{
		TenantID:   id,
		TenantName: name,
		Domain:     domain,
		Email:      email,
		Phone:      phone,
		Status:     status,
		Metadata:   metadata,
	}
	r.tenants[id] = t
	return &t, nil
}

func (r *MemoryTenantsRepo) UpdateTenant(_ context.Context, tenantID string, payload map[string]any) (*Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tenants[tenantID]
	if !ok {
		// create-on-update for dev convenience
		t = Tenant{TenantID: tenantID, Status: "active"}
	}
	if v, ok := payload["tenant_name"].(string); ok && v != "" {
		t.TenantName = v
	}
	if v, ok := payload["domain"].(string); ok {
		t.Domain = v
	}
	if v, ok := payload["email"].(string); ok {
		t.Email = v
	}
	if v, ok := payload["phone"].(string); ok {
		t.Phone = v
	}
	if v, ok := payload["status"].(string); ok && v != "" {
		t.Status = v
	}
	if v, ok := payload["metadata"]; ok {
		if v == nil {
			t.Metadata = nil
		} else if b, err := json.Marshal(v); err == nil {
			t.Metadata = b
		}
	}
	r.tenants[tenantID] = t
	return &t, nil
}

func (r *MemoryTenantsRepo) SetTenantStatus(_ context.Context, tenantID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[tenantID]
	if !ok {
		// create placeholder for dev convenience
		t = Tenant{TenantID: tenantID}
	}
	t.Status = status
	r.tenants[tenantID] = t
	return nil
}
