package repository

import (
	"context"
	"encoding/json"
)

// Tenant aligns with owlRD/db/01_tenants.sql
type Tenant struct {
	TenantID   string          `json:"tenant_id"`
	TenantName string          `json:"tenant_name"`
	Domain     string          `json:"domain,omitempty"`
	Email      string          `json:"email,omitempty"`
	Phone      string          `json:"phone,omitempty"`
	Status     string          `json:"status"` // active | suspended | deleted
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

type TenantsRepo interface {
	ListTenants(ctx context.Context, status string, page, size int) (items []Tenant, total int, err error)
	CreateTenant(ctx context.Context, payload map[string]any) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID string, payload map[string]any) (*Tenant, error)
	SetTenantStatus(ctx context.Context, tenantID string, status string) error
}
