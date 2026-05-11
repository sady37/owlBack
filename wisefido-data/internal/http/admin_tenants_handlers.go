package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type TenantsHandler struct {
	Repo              repository.TenantsRepository  // 使用新的 TenantsRepository 接口
	BranchesRepo      repository.BranchesRepository // optional: 用于创建默认 branch
	Auth              *AuthStore                    // optional (dev only)
	DB                *sql.DB                       // optional (dev only): seed bootstrap admin into DB users table
	logger            *zap.Logger                   // optional: 用于记录创建默认 branch 的错误
	tenantStatusCache *TenantStatusCache            // D-001b：改 tenant.status 后调 Invalidate
}

// SetTenantStatusCache 注入 D-001b cache；改 tenant.status / 删 tenant 后会 Invalidate(prefix)。
func (h *TenantsHandler) SetTenantStatusCache(c *TenantStatusCache) {
	if h != nil {
		h.tenantStatusCache = c
	}
}

// invalidateTenantCache 内部 helper：cache 非 nil 时清掉指定 prefix 的 entry。
func (h *TenantsHandler) invalidateTenantCache(tenantPrefix string) {
	if h != nil && h.tenantStatusCache != nil {
		h.tenantStatusCache.Invalidate(tenantPrefix)
	}
}

func NewTenantsHandler(repo repository.TenantsRepository, branchesRepo repository.BranchesRepository, auth *AuthStore, db *sql.DB) *TenantsHandler {
	// 如果 logger 不可用，创建一个 no-op logger
	logger := zap.NewNop()
	return &TenantsHandler{Repo: repo, BranchesRepo: branchesRepo, Auth: auth, DB: db, logger: logger}
}

func NewTenantsHandlerWithLogger(repo repository.TenantsRepository, branchesRepo repository.BranchesRepository, auth *AuthStore, db *sql.DB, logger *zap.Logger) *TenantsHandler {
	return &TenantsHandler{Repo: repo, BranchesRepo: branchesRepo, Auth: auth, DB: db, logger: logger}
}

func genTempPassword() string {
	// 12-char URL-safe base64 temp password (dev only)
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// upsertBootstrapAdminUserInDB v2 schema 适配：
//   - tenantPrefix 是 CIDR ('fd00:0:T::/48')，写入 users.tenant_id INET
//   - user_account = 'admin'（v2 改 per-tenant UNIQUE，与 v1 一致；不再需要 admin_<slot> 派生）
//   - password 三层：
//       password_hash       = bcrypt(sha256(plain))   真正登录验证（抗暴力）
//       password_check_hash = sha256(plain) bytes     反向定位 user + DB partial UNIQUE 拦 admin 全局密码冲突
//   - role = 'tenant_admin'（partial UNIQUE 索引基于此过滤）
//   - must_change_password=true 强制首次登录改密
//
// 返回 (user_account, error)；error 包含 "uniq_admin_password_check" 时调用方应换密码重试。
func (h *TenantsHandler) upsertBootstrapAdminUserInDB(tenantPrefix, password string) (string, error) {
	if h == nil || h.DB == nil {
		return "", fmt.Errorf("db unavailable")
	}
	if tenantPrefix == "" || password == "" {
		return "", fmt.Errorf("tenant_id and password required")
	}

	// 双层 hash + 检索 hash
	shaHex := HashPassword(password) // sha256(plain) hex
	checkHash, err := hex.DecodeString(shaHex)
	if err != nil {
		return "", fmt.Errorf("decode sha256 hex: %w", err)
	}
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(shaHex), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}

	const user_account = "admin"
	// INSERT or UPDATE：(tenant_id, user_account) UNIQUE 处理同 tenant 重入；
	// admin 类的 password_check_hash 全局 UNIQUE 拦明文密码碰撞（DB 层）。
	_, err = h.DB.Exec(
		`INSERT INTO users (tenant_id, user_account, password_hash, password_check_hash,
		                    nickname, role, status, must_change_password)
		 VALUES ($1::INET, $2, $3, $4, 'Admin', 'tenant_admin', 'active', true)
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   password_hash = EXCLUDED.password_hash,
		   password_check_hash = EXCLUDED.password_check_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = 'active',
		   must_change_password = true,
		   updated_at = NOW()`,
		tenantPrefix, user_account, string(bcryptHash), checkHash,
	)
	if err != nil {
		return "", fmt.Errorf("upsert user: %w", err)
	}
	return user_account, nil
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
			search := r.URL.Query().Get("search")
			page := parseInt(r.URL.Query().Get("page"), 1)
			size := parseInt(r.URL.Query().Get("size"), 50)
			filter := repository.TenantFilters{
				Status: status,
				Search: search,
			}
			items, total, err := h.Repo.ListTenants(r.Context(), filter, page, size)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to list tenants"))
				return
			}
			out := make([]any, 0, len(items))
			for _, t := range items {
				out = append(out, map[string]any{
					"tenant_id":   t.TenantID,
					"tenant_type": t.TenantType,
					"tenant_name": t.TenantName,
					"kind":        t.Kind,
					"timezone":    t.Timezone,
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
			// 转换为 domain.Tenant
			tenant := &domain.Tenant{
				TenantType: getStringOrDefaultFromMap(payload, "tenant_type", "organization"),
				TenantName: getStringFromMap(payload, "tenant_name"),
				Domain:     getStringFromMap(payload, "domain"),
				Email:      getStringFromMap(payload, "email"),
				Phone:      getStringFromMap(payload, "phone"),
				Status:     getStringOrDefaultFromMap(payload, "status", "active"),
			}
			if v, ok := payload["metadata"]; ok && v != nil {
				if b, err := json.Marshal(v); err == nil {
					tenant.Metadata = b
				}
			}
			if strings.TrimSpace(tenant.TenantName) == "" {
				writeJSON(w, http.StatusOK, Fail("tenant_name is required"))
				return
			}
			tenant.TenantName = strings.TrimSpace(tenant.TenantName)
			tenantID, err := h.Repo.CreateTenant(r.Context(), tenant)
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "already exists") {
					writeJSON(w, http.StatusOK, Fail("Tenant name already exists, please choose a different name"))
					return
				}
				writeJSON(w, http.StatusOK, Fail("failed to create tenant: "+errMsg))
				return
			}
			
			// 自动创建默认 branch（branch_name = "default"）
			if h.BranchesRepo != nil {
				defaultBranch := &domain.Branch{
					BranchName:    domain.DefaultBranchName,
					Description:   sql.NullString{String: domain.DefaultBranchDescription, Valid: true},
				}
				_, err = h.BranchesRepo.CreateBranch(r.Context(), tenantID, defaultBranch)
				if err != nil {
					// 记录错误但继续执行（不影响 tenant 创建）
					// 如果创建失败，用户可以手动创建 branch
					if h.logger != nil {
						h.logger.Warn("Failed to create default branch for tenant",
							zap.String("tenant_id", tenantID),
							zap.Error(err),
						)
					}
				}
			}
			
			// 获取创建的租户
			t, err := h.Repo.GetTenant(r.Context(), tenantID)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to get created tenant"))
				return
			}
			out := map[string]any{
				"tenant_id":   t.TenantID,
				"tenant_type": t.TenantType,
				"tenant_name": t.TenantName,
				"kind":        t.Kind,
				"timezone":    t.Timezone,
				"domain":      t.Domain,
				"email":       t.Email,
				"phone":       t.Phone,
				"status":      t.Status,
				"metadata":    t.Metadata,
			}
			// Dev bootstrap account: create ONLY admin with TEMP password.
			// v2: user_account 全局 UNIQUE，由 upsertBootstrapAdminUserInDB 派生为 'admin_<slot>'；
			// 明文密码全局唯一约束 → 撞了重试（temp 是随机 12 字节 base64，碰撞概率近 0）。
			if h.Auth != nil {
				var adminPwd, user_account string
				for retry := 0; retry < 3; retry++ {
					adminPwd = genTempPassword()
					u, err := h.upsertBootstrapAdminUserInDB(t.TenantID, adminPwd)
					if err == nil {
						user_account = u
						break
					}
					if !strings.Contains(err.Error(), "collision") {
						break // 非密码碰撞错误就放弃
					}
				}
				if user_account != "" {
					_ = h.Auth.UpsertUser(t.TenantID, user_account, "Admin", adminPwd)
					out["bootstrap_accounts"] = []any{
						map[string]any{"user_account": user_account, "role": "Admin", "temp_password": adminPwd},
					}
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
		// v2: tenant_id 现在装 IPv6 prefix CIDR（'fd00:0:6::/48'），URL 拆段时
		// '/48' 会被当成下一个 path segment。识别 IPv6 串 + 数字尾巴重新拼回 mask。
		parts := stitchCIDRSegments(strings.Split(rest, "/"))
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
				// v2: 走双层 hash bcrypt(sha256(plain))；user_account 全局 UNIQUE 派生 'admin_<slot>'；
				// 业务约束 admin 明文密码全局唯一 → 命中 collision 时 admin 必须换密码重试。
				user_account, upErr := h.upsertBootstrapAdminUserInDB(id, adminPwd)
				if upErr != nil {
					if strings.Contains(upErr.Error(), "collision") {
						writeJSON(w, http.StatusOK, Fail("password already used by another tenant admin; please choose a different password"))
						return
					}
					writeJSON(w, http.StatusOK, Fail("failed to reset admin password: "+upErr.Error()))
					return
				}
				if h.Auth != nil {
					_ = h.Auth.UpsertUser(id, user_account, "Admin", adminPwd)
				}
				outAccounts := []any{
					map[string]any{"user_account": user_account, "role": "Admin", "temp_password": adminPwd},
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
				h.invalidateTenantCache(id) // D-001b
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			// 获取现有租户
			existing, err := h.Repo.GetTenant(r.Context(), id)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("tenant not found"))
				return
			}
			// 更新字段
			if v := getStringFromMap(payload, "tenant_type"); v != "" {
				existing.TenantType = v
			}
			if _, ok := payload["tenant_name"]; ok {
				v := strings.TrimSpace(getStringFromMap(payload, "tenant_name"))
				if v == "" {
					writeJSON(w, http.StatusOK, Fail("tenant_name cannot be empty"))
					return
				}
				existing.TenantName = v
			}
			if _, ok := payload["domain"]; ok {
				existing.Domain = getStringFromMap(payload, "domain")
			}
			if _, ok := payload["email"]; ok {
				existing.Email = getStringFromMap(payload, "email")
			}
			if _, ok := payload["phone"]; ok {
				existing.Phone = getStringFromMap(payload, "phone")
			}
			if _, ok := payload["status"]; ok {
				// v2: status 走 active/suspended/deleted（HIPAA 软删）
				existing.Status = getStringFromMap(payload, "status")
			}
			if v, ok := payload["metadata"]; ok {
				if v == nil {
					existing.Metadata = nil
				} else if b, err := json.Marshal(v); err == nil {
					existing.Metadata = b
				}
			}
			err = h.Repo.UpdateTenant(r.Context(), id, existing)
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "already exists") {
					writeJSON(w, http.StatusOK, Fail("Tenant name already exists, please choose a different name"))
					return
				}
				writeJSON(w, http.StatusOK, Fail("failed to update tenant: "+errMsg))
				return
			}
			h.invalidateTenantCache(id) // D-001b: status 字段可能在 update 中变了
			// 获取更新后的租户
			t, err := h.Repo.GetTenant(r.Context(), id)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to get updated tenant"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":   t.TenantID,
				"tenant_type": t.TenantType,
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
			// 智能删除：repo.DeleteTenant 内部判定空 tenant 硬删 / 非空软删
			if err := h.Repo.DeleteTenant(r.Context(), id); err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to delete tenant: "+err.Error()))
				return
			}
			h.invalidateTenantCache(id) // D-001b
			writeJSON(w, http.StatusOK, Ok[any](nil))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

// 辅助函数（避免与 unit_handler.go 中的 getString 冲突）
func getStringFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringOrDefaultFromMap(m map[string]any, key, defaultValue string) string {
	if v := getStringFromMap(m, key); v != "" {
		return v
	}
	return defaultValue
}
