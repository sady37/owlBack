package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"wisefido-data/internal/repository"
)

type TenantsHandler struct {
	Repo repository.TenantsRepo
	Auth *AuthStore // optional (dev only)
	DB   *sql.DB    // optional (dev only): seed bootstrap admin into DB users table
}

func NewTenantsHandler(repo repository.TenantsRepo, auth *AuthStore, db *sql.DB) *TenantsHandler {
	return &TenantsHandler{Repo: repo, Auth: auth, DB: db}
}

func genTempPassword() string {
	// 12-char URL-safe base64 temp password (dev only)
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (h *TenantsHandler) upsertBootstrapAdminUserInDB(tenantID, password string) {
	if h == nil || h.DB == nil {
		return
	}
	// password_hash should only depend on password itself (independent of account/phone/email)
	ah, _ := hex.DecodeString(HashAccount("admin"))
	aph, _ := hex.DecodeString(HashPassword(password))
	// Ensure at least non-empty hashes; if decode fails, skip DB write (dev helper only).
	if len(ah) == 0 || len(aph) == 0 {
		return
	}
	// Best-effort insert/update. Login currently uses AuthStore, but DB should reflect the bootstrap admin.
	_, _ = h.DB.Exec(
		`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active')
		 ON CONFLICT (tenant_id, user_account)
		 DO UPDATE SET user_account_hash = EXCLUDED.user_account_hash,
		               password_hash = EXCLUDED.password_hash,
		               nickname = EXCLUDED.nickname,
		               role = EXCLUDED.role,
		               status = 'active'`,
		tenantID,
		"admin",
		ah,
		aph,
		"Admin",
		"Admin",
	)
}

func (h *TenantsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.Repo == nil {
		writeJSON(w, http.StatusOK, Fail("tenants repo is not configured"))
		return
	}

	switch {
	case r.URL.Path == "/admin/api/v1/tenants":
		switch r.Method {
		case http.MethodGet:
			status := r.URL.Query().Get("status")
			page := parseInt(r.URL.Query().Get("page"), 1)
			size := parseInt(r.URL.Query().Get("size"), 50)
			items, total, err := h.Repo.ListTenants(r.Context(), status, page, size)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to list tenants"))
				return
			}
			out := make([]any, 0, len(items))
			for _, t := range items {
				out = append(out, map[string]any{
					"tenant_id":   t.TenantID,
					"tenant_name": t.TenantName,
					"domain":      t.Domain,
					"email":       t.Email,
					"phone":       t.Phone,
					"status":      t.Status,
					"metadata":    t.Metadata,
				})
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": out, "total": total}))
			return
		case http.MethodPost:
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			// default status
			if _, ok := payload["status"]; !ok {
				payload["status"] = "active"
			}
			t, err := h.Repo.CreateTenant(r.Context(), payload)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to create tenant"))
				return
			}
			out := map[string]any{
				"tenant_id":   t.TenantID,
				"tenant_name": t.TenantName,
				"domain":      t.Domain,
				"email":       t.Email,
				"phone":       t.Phone,
				"status":      t.Status,
				"metadata":    t.Metadata,
			}
			// Dev bootstrap account: create ONLY admin with TEMP password.
			// Manager is a role; real "manager" users should be created by admin.
			if h.Auth != nil {
				adminPwd := genTempPassword()
				_ = h.Auth.UpsertUser(t.TenantID, "admin", "Admin", adminPwd)
				h.upsertBootstrapAdminUserInDB(t.TenantID, adminPwd)
				out["bootstrap_accounts"] = []any{
					map[string]any{"user_account": "admin", "role": "Admin", "temp_password": adminPwd},
				}
			}
			writeJSON(w, http.StatusOK, Ok(out))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/tenants/"):
		rest := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tenants/")
		if rest == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		parts := strings.Split(rest, "/")
		id := parts[0]
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPost:
			// Dev-only helper: reset bootstrap passwords for this tenant (AuthStore is in-memory).
			// POST /admin/api/v1/tenants/:id/bootstrap-accounts/reset
			if len(parts) == 3 && parts[1] == "bootstrap-accounts" && parts[2] == "reset" {
				if h.DB == nil {
					writeJSON(w, http.StatusOK, Fail("database not available"))
					return
				}
				// Reset ONLY admin (bootstrap account).
				// Optional query: user_account=admin (accepted). Any other value is invalid.
				ua := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("user_account")))
				if ua != "" && ua != "admin" {
					writeJSON(w, http.StatusOK, Fail("invalid user_account"))
					return
				}
				// Read password from request body if provided, otherwise generate temp password
				// IMPORTANT: password_hash should only depend on password itself (no trim, no modification)
				var adminPwd string
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				// Read password from payload (use as-is, no trim)
				if pwd, ok := payload["new_password"].(string); ok && pwd != "" {
					// Use password as-is, no trim (password_hash = SHA256(password))
					adminPwd = pwd
				}
				// If no password provided in body, generate temp password (backward compatibility)
				if adminPwd == "" {
					adminPwd = genTempPassword()
				}
				// Update password in DB
				// password_hash should only depend on password itself (independent of account/phone/email)
				ah, _ := hex.DecodeString(HashAccount("admin"))
				aph, _ := hex.DecodeString(HashPassword(adminPwd))
				if len(ah) == 0 || len(aph) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash credentials"))
					return
				}
				_, err := h.DB.ExecContext(
					r.Context(),
					`UPDATE users SET password_hash = $2
					 WHERE tenant_id = $1 AND user_account = 'admin'`,
					id, aph,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset admin password: %v", err)))
					return
				}
				// Optional: keep AuthStore in sync for dev/stub flows
				if h.Auth != nil {
					_ = h.Auth.UpsertUser(id, "admin", "Admin", adminPwd)
				}
				outAccounts := []any{
					map[string]any{"user_account": "admin", "role": "Admin", "temp_password": adminPwd},
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{
					"tenant_id":          id,
					"bootstrap_accounts": outAccounts,
				}))
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPut:
			if len(parts) != 1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			// convenience: if only status is provided, call SetTenantStatus
			if st, ok := payload["status"].(string); ok && len(payload) == 1 {
				if err := h.Repo.SetTenantStatus(r.Context(), id, st); err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to update tenant status"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			t, err := h.Repo.UpdateTenant(r.Context(), id, payload)
			if err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusOK, Fail("tenant not found"))
					return
				}
				writeJSON(w, http.StatusOK, Fail("failed to update tenant"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":   t.TenantID,
				"tenant_name": t.TenantName,
				"domain":      t.Domain,
				"email":       t.Email,
				"phone":       t.Phone,
				"status":      t.Status,
				"metadata":    t.Metadata,
			}))
			return
		case http.MethodDelete:
			if len(parts) != 1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// soft delete
			_ = h.Repo.SetTenantStatus(r.Context(), id, "deleted")
			writeJSON(w, http.StatusOK, Ok[any](nil))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
