package httpapi

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"wisefido-data/internal/repository"
)

// ResolveTenantIDFromHeader：X-Tenant-Id 为 INET prefix (如 'fd00:0:3::/48') 直接用；
// 否则按 tenants.tenant_name 反查（如 'demo' → tenant_id）。
// v2 全栈 tenant_id = INET CIDR，UUID legacy 已退役。
func ResolveTenantIDFromHeader(ctx context.Context, tr repository.TenantsRepository, rawHeader string) (string, error) {
	raw := strings.TrimSpace(rawHeader)
	if raw == "" || raw == "null" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if _, err := netip.ParsePrefix(raw); err == nil {
		return raw, nil
	}
	if tr == nil {
		return "", fmt.Errorf("tenant resolve not available")
	}
	return tr.GetTenantIDByName(ctx, raw)
}
