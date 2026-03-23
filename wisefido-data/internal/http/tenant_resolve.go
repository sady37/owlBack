package httpapi

import (
	"context"
	"fmt"
	"strings"

	"wisefido-data/internal/repository"

	"github.com/google/uuid"
)

// ResolveTenantIDFromHeader：X-Tenant-Id 为 UUID 则直接用；否则按 tenants.tenant_name 反查（如 demo → tenant_id）
func ResolveTenantIDFromHeader(ctx context.Context, tr repository.TenantsRepository, rawHeader string) (string, error) {
	raw := strings.TrimSpace(rawHeader)
	if raw == "" || raw == "null" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if _, err := uuid.Parse(raw); err == nil {
		return raw, nil
	}
	if tr == nil {
		return "", fmt.Errorf("tenant resolve not available")
	}
	return tr.GetTenantIDByName(ctx, raw)
}
