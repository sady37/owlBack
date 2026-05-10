package httpapi

// AuthV2Handler implements owl_v2 IPv6-native authentication.
//
// 凭证传输形态（双层 hash 设计）：
//   - 前端 username 明文 + sha256(password) → hex 上送
//   - DB 存 bcrypt(sha256(password)) — bcrypt 抗暴力，sha256 防止网络/日志层明文泄漏
//   - 服务端 bcrypt.CompareHashAndPassword(stored_bcrypt_bytes, sha256_hex_bytes) 验证
//
// 与 v1 区别（不向后兼容）：
//   - v1 客户端 sha256(account)/sha256(account:password) → bytea 与 DB bytea 直接比；v2 走 bcrypt 抗暴力
//   - v1 用 user_branches 表算 branch_ids；v2 用户关系完全靠 IPv6 prefix 自带层级派生（owl-common/spatial）
//   - v1 返回 tenant_id (UUID)；v2 返回 tenant_prefix (INET /48 字符串)
//
// 端点：
//   POST /auth/api/v2/search-tenants  body:{username,password_hash}                 → matches[]
//   POST /auth/api/v2/login            body:{username,password_hash,tenant_prefix?} → session

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthV2Handler 提供 v2 (owl_v2) 鉴权 endpoint。
type AuthV2Handler struct {
	db           *sql.DB
	sessionStore SessionStore
	logger       *zap.Logger
	rateLimiter  *loginRateLimiter
}

// NewAuthV2Handler 构造 v2 auth handler。
func NewAuthV2Handler(db *sql.DB, sessionStore SessionStore, logger *zap.Logger) *AuthV2Handler {
	return &AuthV2Handler{
		db:           db,
		sessionStore: sessionStore,
		logger:       logger,
		rateLimiter:  newLoginRateLimiter(10*time.Minute, 10),
	}
}

// RegisterAuthV2Routes 注册 v2 auth 路由（中间件 skip 列表里要把 v2 路径加进去，否则被 401）。
func (r *Router) RegisterAuthV2Routes(h *AuthV2Handler) {
	r.Handle("/auth/api/v2/search-tenants", h.HandleSearchTenants)
	r.Handle("/auth/api/v2/login", h.HandleLogin)
}

// =============================================================================
// DTOs
// =============================================================================

type v2LoginReq struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`           // sha256(password) hex；DB 存 bcrypt(sha256(password))
	TenantPrefix string `json:"tenant_prefix,omitempty"` // 可选；多 tenant 重名 username 时必填
}

type v2TenantMatch struct {
	TenantPrefix string `json:"tenant_prefix"`
	TenantName   string `json:"tenant_name"`
	Role         string `json:"role,omitempty"`
}

type v2SearchResp struct {
	Matches []v2TenantMatch `json:"matches"`
}

type v2LoginResp struct {
	AccessToken  string `json:"accessToken"`
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	NickName     string `json:"nickName"`
	Role         string `json:"role"`
	TenantPrefix string `json:"tenant_prefix"`
	TenantName   string `json:"tenant_name"`
	HoA          string `json:"hoa,omitempty"` // resident/caregiver 才有；admin/staff 为空
	HomePath     string `json:"homePath,omitempty"`
}

// =============================================================================
// Handlers
// =============================================================================

// HandleSearchTenants 当 username 在多 tenant 重名时，前端先调这个拿 tenant 列表让用户选；
// 命中 1 条时前端可直接 auto-fill 调 login。
//
// 业务规则：必须 username + password 都对，才返回机构列表（与 v1 SearchInstitutions 同语义，避免泄露 username 是否存在）。
func (h *AuthV2Handler) HandleSearchTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body v2LoginReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid JSON body: "+err.Error()))
		return
	}
	username := strings.TrimSpace(body.Username)
	password := body.PasswordHash
	if username == "" || password == "" {
		writeJSON(w, http.StatusOK, Fail("username and password_hash required"))
		return
	}

	rateKey := r.RemoteAddr + "|" + strings.ToLower(username)
	if wait := h.rateLimiter.isBlocked(rateKey); wait > 0 {
		w.Header().Set("Retry-After", itoa(wait))
		writeJSON(w, http.StatusTooManyRequests, Fail("too many attempts; retry later"))
		return
	}

	matches, err := h.searchTenantsByCredentials(r.Context(), username, password)
	if err != nil {
		h.logger.Warn("v2 auth: search-tenants db error", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("search failed"))
		return
	}
	if len(matches) == 0 {
		h.rateLimiter.recordFailure(rateKey)
	}
	writeJSON(w, http.StatusOK, Ok(v2SearchResp{Matches: matches}))
}

// HandleLogin 标准 v2 登录。如果 tenant_prefix 给了，仅在该 tenant 内查；否则要求全局唯一。
func (h *AuthV2Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body v2LoginReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid JSON body: "+err.Error()))
		return
	}
	username := strings.TrimSpace(body.Username)
	password := body.PasswordHash
	tenantPrefix := strings.TrimSpace(body.TenantPrefix)
	if username == "" || password == "" {
		writeJSON(w, http.StatusOK, Fail("username and password_hash required"))
		return
	}

	rateKey := r.RemoteAddr + "|" + strings.ToLower(username)
	if wait := h.rateLimiter.isBlocked(rateKey); wait > 0 {
		w.Header().Set("Retry-After", itoa(wait))
		writeJSON(w, http.StatusTooManyRequests, Fail("too many attempts; retry later"))
		return
	}

	row, err := h.findActiveUser(r.Context(), username, password, tenantPrefix)
	if err != nil {
		if errors.Is(err, errMultipleTenants) {
			writeJSON(w, http.StatusOK, Fail("multiple institutions found, please select one"))
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			h.rateLimiter.recordFailure(rateKey)
			h.recordLoginFailure(r.Context(), username)
			writeJSON(w, http.StatusOK, Fail("invalid credentials"))
			return
		}
		h.logger.Warn("v2 auth: login db error", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("login failed"))
		return
	}
	// 兜底：password_check_hash 已确认明文一致，bcrypt 再做"抗暴力"层验证（DB 内 stored bcrypt(sha256_hex) 必须匹配）
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		h.rateLimiter.recordFailure(rateKey)
		h.recordLoginFailure(r.Context(), username)
		writeJSON(w, http.StatusOK, Fail("invalid credentials"))
		return
	}

	// 异步更新 last_login_at + 重置失败计数 / 自动解锁（locked → active）
	go func(uid string) {
		_, _ = h.db.Exec(`
			UPDATE users SET
			  last_login_at = NOW(),
			  last_active_at = NOW(),
			  failed_login_count = 0,
			  locked_until = NULL,
			  status = CASE WHEN status = 'locked' THEN 'active' ELSE status END
			WHERE user_id = $1`, uid)
	}(row.UserID)

	token := GenerateSessionToken()
	if h.sessionStore != nil {
		_ = h.sessionStore.Set(r.Context(), token, SessionData{
			UserID:       row.UserID,
			TenantID:     row.TenantPrefix, // v2: 用 prefix 字符串占 TenantID 字段，middleware 注入 X-Tenant-Prefix
			TenantPrefix: row.TenantPrefix,
			UserType:     "staff",
			Role:         row.Role,
			HoA:          row.HoA,
		}, 24*time.Hour)
	}

	resp := v2LoginResp{
		AccessToken:  token,
		UserID:       row.UserID,
		Username:     row.Username,
		NickName:     fallback(row.Nickname, row.Username),
		Role:         row.Role,
		TenantPrefix: row.TenantPrefix,
		TenantName:   row.TenantName,
		HoA:          row.HoA,
		HomePath:     "/monitoring/overview",
	}
	writeJSON(w, http.StatusOK, Ok(resp))
}

// =============================================================================
// Internal: SQL helpers (v2 schema only)
// =============================================================================

var errMultipleTenants = errors.New("multiple institutions match this username; tenant_prefix required")

type userRow struct {
	UserID       string
	Username     string
	Nickname     string
	Role         string
	PasswordHash string
	TenantPrefix string
	TenantName   string
	HoA          string
}

// findActiveUser 在 v2 users + tenants + user_roles + roles 表里按 (username, password_check_hash)
// 反向定位活跃用户。这是 v2 cutover 的"不输 organize 也能登录"实现：
//
//   - sha256(plain) 无 salt 存为 password_check_hash bytea；前端发的 password_hash hex decode 后比对
//   - admin 类（platform_admin/tenant_admin）有 partial UNIQUE → 命中必然 0 或 1 条
//   - 普通 user 多点执业可能同 (username, password) → 命中 N 条 → errMultipleTenants
//   - tenant_prefix 给了就额外限定（用户在 LoginForm 手工选 organize 的情况）
//
// SQL 设计：
//   - JOIN tenants 拿 tenant_name
//   - LEFT JOIN user_roles + roles 拿 RBAC role_code（normalize 成 PascalCase 给前端）
//   - status='active' 才放行
//
// password 参数：前端发的 sha256(plain) hex 字符串，SQL 内 decode($2, 'hex') 转 bytea 比对。
func (h *AuthV2Handler) findActiveUser(ctx context.Context, username, passwordHashHex, tenantPrefix string) (*userRow, error) {
	// 状态门槛（用 OR 组合）：
	//   D-001a tenant 级联停权：tenant.status='active' 才放行；platform_admin 例外
	//   账号锁定 (B-009)：status='active' 直接放行；status='locked' 仅 locked_until <= NOW() 才放行（自动解锁）
	const baseSelect = `
		SELECT u.user_id::text,
		       u.username,
		       COALESCE(u.nickname, ''),
		       COALESCE(u.role, ''),
		       COALESCE(r.role_name, '') AS rbac_role,
		       COALESCE(u.password_hash, ''),
		       host(u.tenant_prefix) || '/48' AS tenant_prefix_cidr,
		       COALESCE(t.tenant_name, ''),
		       COALESCE(host(u.hoa), '')
		  FROM users u
		  JOIN tenants t ON t.spatial_prefix = u.tenant_prefix
		  LEFT JOIN user_roles ur ON ur.user_id = u.user_id
		  LEFT JOIN roles r ON r.role_id = ur.role_id
		 WHERE u.username = $1
		   AND u.password_check_hash = decode($2, 'hex')
		   AND (
		         COALESCE(u.status,'active') = 'active'
		         OR (u.status = 'locked' AND u.locked_until IS NOT NULL AND u.locked_until <= NOW())
		       )
		   AND (
		         COALESCE(t.status,'active') = 'active'
		         OR u.role = 'platform_admin'
		         OR r.role_code = 'platform_admin'
		       )
	`
	scan := func(rs interface {
		Scan(...any) error
	}, out *userRow) error {
		var rbacRole string
		if err := rs.Scan(&out.UserID, &out.Username, &out.Nickname, &out.Role, &rbacRole,
			&out.PasswordHash, &out.TenantPrefix, &out.TenantName, &out.HoA); err != nil {
			return err
		}
		out.Role = normalizeRole(out.Role, rbacRole)
		return nil
	}

	if tenantPrefix != "" {
		row := h.db.QueryRowContext(ctx, baseSelect+` AND u.tenant_prefix = $3::INET LIMIT 1`, username, passwordHashHex, tenantPrefix)
		out := &userRow{}
		if err := scan(row, out); err != nil {
			return nil, err
		}
		return out, nil
	}

	// 无 tenant_prefix：(username, password_check_hash) 命中多条 → 多点执业冲突，要求选 organize
	rows, err := h.db.QueryContext(ctx, baseSelect, username, passwordHashHex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var first *userRow
	count := 0
	for rows.Next() {
		count++
		if count > 1 {
			return nil, errMultipleTenants
		}
		first = &userRow{}
		if err := scan(rows, first); err != nil {
			return nil, err
		}
	}
	if first == nil {
		return nil, sql.ErrNoRows
	}
	return first, nil
}

// normalizeRole 把 (users.role, roles.role_name) 归一到前端 pagePermissions 期待的 PascalCase
// role code。优先 users.role，空时回落到 roles.role_name 友好名做映射。
//
// 映射来自当前 owl_v2 seed 数据（90_seed_roles.sql 等）。未来若引入新角色名，扩这里。
func normalizeRole(usersRole, rolesName string) string {
	s := strings.TrimSpace(usersRole)
	if s == "" {
		s = strings.TrimSpace(rolesName)
	}
	switch strings.ToLower(s) {
	case "platform_admin", "platform administrator", "system administrator", "systemadmin":
		return "SystemAdmin"
	case "system_operator", "system operator", "systemoperator":
		return "SystemOperator"
	case "tenant_admin", "tenant administrator", "admin":
		return "Admin"
	case "branch_manager", "manager":
		return "Manager"
	case "nurse":
		return "Nurse"
	case "caregiver":
		return "Caregiver"
	case "family":
		return "Family"
	case "viewer":
		return "Viewer"
	case "it":
		return "IT"
	case "resident":
		return "Resident"
	}
	return s // 未知名直接透传，前端再处理（避免吞掉）
}

// searchTenantsByCredentials 按 (username, password_check_hash) 找所有匹配的 tenant，用于 LoginForm
// 选机构（多点执业场景）。SQL 直接走 password_check_hash 索引，O(log n)；不再 N+1 bcrypt 比较。
//
// admin 类有 partial UNIQUE(password_check_hash) → 永远 0/1 条；普通 user 多点执业可能 N 条。
//
// D-001a tenant 级联停权：suspended/deleted tenant 不进 search 结果（platform_admin 例外）。
func (h *AuthV2Handler) searchTenantsByCredentials(ctx context.Context, username, passwordHashHex string) ([]v2TenantMatch, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT host(u.tenant_prefix) || '/48' AS tenant_prefix_cidr,
		       COALESCE(t.tenant_name, ''),
		       COALESCE(u.role, '')
		  FROM users u
		  JOIN tenants t ON t.spatial_prefix = u.tenant_prefix
		 WHERE u.username = $1
		   AND u.password_check_hash = decode($2, 'hex')
		   AND COALESCE(u.status,'active') = 'active'
		   AND (
		         COALESCE(t.status,'active') = 'active'
		         OR u.role = 'platform_admin'
		       )
	`, username, passwordHashHex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]v2TenantMatch, 0, 2)
	for rows.Next() {
		var prefix, name, role string
		if err := rows.Scan(&prefix, &name, &role); err != nil {
			continue
		}
		out = append(out, v2TenantMatch{TenantPrefix: prefix, TenantName: name, Role: role})
	}
	return out, nil
}

// recordLoginFailure 持久化登录失败计数。
//
// 当 (username, password_check_hash) 不匹配（密码错或账号不存在），按 username 找所有同名 user
// 累加 failed_login_count + last_failed_at；累计 ≥ failedLoginThreshold 时设 status='locked' +
// locked_until = NOW() + lockoutDuration。
//
// 注意：v2 username per-tenant UNIQUE，admin 攻击 'admin' 时会一起锁所有 tenant 的 admin 用户
// （同 username 在多 tenant 全部 +1）。这是保守的"广撒网"防护：攻击者无法定向某 tenant 的 admin。
//
// 同步执行，确保在响应前 DB 已写入；写失败仅 log，不阻塞响应。
const (
	failedLoginThreshold = 10
	lockoutDuration      = "2 hours"
)

func (h *AuthV2Handler) recordLoginFailure(ctx context.Context, username string) {
	if username == "" {
		return
	}
	_, err := h.db.ExecContext(ctx, `
		UPDATE users SET
		  failed_login_count = failed_login_count + 1,
		  last_failed_at = NOW(),
		  status = CASE
		             WHEN failed_login_count + 1 >= $2 THEN 'locked'
		             ELSE status
		           END,
		  locked_until = CASE
		                   WHEN failed_login_count + 1 >= $2 THEN NOW() + ($3 || '')::INTERVAL
		                   ELSE locked_until
		                 END
		WHERE username = $1
	`, username, failedLoginThreshold, lockoutDuration)
	if err != nil {
		h.logger.Warn("v2 auth: record login failure failed", zap.Error(err))
	}
}

// =============================================================================
// utils
// =============================================================================

func fallback(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
