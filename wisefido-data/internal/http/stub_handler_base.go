package httpapi

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"wisefido-data/internal/repository"
)

// StubHandler：用于 DB/真实逻辑未就绪时，先保证 owlFront 不 404、页面可渲染（code=2000 + 空数据）
type StubHandler struct {
	Tenants   repository.TenantsRepo
	AuthStore *AuthStore
	DB        *sql.DB // optional: when set, some admin endpoints read/write real DB
}

func NewStubHandler(tenants repository.TenantsRepo, auth *AuthStore, db *sql.DB) *StubHandler {
	return &StubHandler{Tenants: tenants, AuthStore: auth, DB: db}
}

func allowAuthStoreFallback() bool {
	// Security hardening:
	// - AuthStore is in-memory and should NOT be used in real deployments.
	// - Only allow it when explicitly enabled for local dev.
	return os.Getenv("ALLOW_AUTHSTORE_FALLBACK") == "true"
}

func (s *StubHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
		return tid, true
	}
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	// Try to resolve tenant from user ID via DB query (if DB is available)
	if s != nil && s.DB != nil {
		userID := r.Header.Get("X-User-Id")
		if userID != "" {
			var tenantID string
			err := s.DB.QueryRowContext(r.Context(), "SELECT tenant_id::text FROM users WHERE user_id = $1", userID).Scan(&tenantID)
			if err == nil && tenantID != "" {
				return tenantID, true
			}
		}
	}
	// Convenience: SystemAdmin without tenant header falls back to System tenant.
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// SystemTenantID is the fixed platform tenant id used for SystemAdmin (dev bootstrap).
func SystemTenantID() string {
	// IMPORTANT:
	// - Do NOT use 00000000-0000-0000-0000-000000000000 because owlRD uses it as a sentinel
	//   meaning "unassigned" (e.g. device_store.tenant_id).
	return "00000000-0000-0000-0000-000000000001"
}
