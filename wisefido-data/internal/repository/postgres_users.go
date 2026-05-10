package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// PostgresUsersRepository — owl_v2 schema 版本。
//
// v2 users 表关键差异（不向后兼容）：
//   - 主键 user_id UUID 不变；租户字段 tenant_id UUID → tenant_prefix INET
//   - 列改名：user_account → username（无 user_account_hash bytea）
//   - 列删除：pin_hash bytea / email_hash / phone_hash / alarm_levels / alarm_channels / alarm_scope
//             / preferences / user_tags
//   - 列新增：mobile_pin_hash varchar (bcrypt) / hoa inet / subject_slot / employee_code
//             / hire_date / leave_date / notify_mode / work_days / work_time_*
//   - 不再有 user_branches 关联表；branch 关系用 user_roles.role_id + role_permissions.resource_scope INET 派生
//   - password 双层 hash：DB 存 bcrypt(sha256(plain))；输入端 user.PasswordHash 是 sha256 bytes，repo 在 INSERT 前 bcrypt
//
// 兼容策略：UsersRepository interface + domain.User struct 不变；
// v2 没有的字段（hash bytea / alarm_* / preferences / branch_id / tags）SELECT 时填空值，写入时忽略。
type PostgresUsersRepository struct {
	db     *sql.DB
	logger sqlLogger // 可选；保留旧接口
}

// NewPostgresUsersRepository 创建 v2 users repo。
func NewPostgresUsersRepository(db *sql.DB) *PostgresUsersRepository {
	return &PostgresUsersRepository{db: db}
}

// SetLogger 兼容旧接口（不强制使用）
func (r *PostgresUsersRepository) SetLogger(l sqlLogger) {
	r.logger = l
}

var _ UsersRepository = (*PostgresUsersRepository)(nil)

// 公共 SELECT 子句：把 v2 列适配回 v1 domain.User 字段。
//
// 重要：role 字段在 v2 schema 优先取 user_roles.role 关联的 roles.role_code（v2 RBAC 主路径）；
// users.role 字段（v1 留下的）作 fallback。Go 侧 scanUser 会 normalize 成 PascalCase 给上层。
const usersFromClause = `
	FROM users u
	LEFT JOIN LATERAL (
		SELECT r.role_code
		  FROM user_roles ur
		  JOIN roles r ON r.role_id = ur.role_id
		 WHERE ur.user_id = u.user_id
		 ORDER BY r.is_system DESC, r.role_code
		 LIMIT 1
	) ur ON TRUE
`

const usersSelectCols = `
		u.user_id::text,
		COALESCE(host(u.tenant_prefix) || '/48', '') AS tenant_id,
		u.username AS user_account,
		ARRAY[]::bytea[] AS user_account_hash_placeholder,
		COALESCE(u.password_hash, '')::bytea AS password_hash,
		COALESCE(u.mobile_pin_hash, '')::bytea AS pin_hash,
		u.nickname,
		u.email,
		u.phone,
		COALESCE(NULLIF(u.role, ''), ur.role_code, '') AS role,
		COALESCE(u.status, 'active') AS status,
		ARRAY[]::bytea[] AS email_hash_placeholder,
		ARRAY[]::bytea[] AS phone_hash_placeholder,
		ARRAY[]::text[] AS alarm_levels,
		ARRAY[]::text[] AS alarm_channels,
		NULL::text AS alarm_scope,
		u.last_login_at,
		NULL::text AS user_tags,
		NULL::text AS branch_id,
		NULL::text AS branch_name,
		NULL::text AS preferences
`

// scanUser 配合 usersSelectCols。注意 user_account_hash/email_hash/phone_hash 是 placeholder bytea[]，
// 这里 scan 到一个临时变量然后丢弃；domain.User 的 hash bytea 字段总是空。
type sqlLogger interface {
	Debug(msg string, fields ...any)
}

func (r *PostgresUsersRepository) scanUser(rs interface {
	Scan(dest ...any) error
}) (*domain.User, error) {
	var user domain.User
	var passwordHash, pinHash []byte
	var nickname, email, phone, role, alarmScope, tags, branchID, branchName, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray
	var dummyUserHash, dummyEmailHash, dummyPhoneHash pq.ByteaArray

	if err := rs.Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&dummyUserHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&role,
		&user.Status,
		&dummyEmailHash,
		&dummyPhoneHash,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchID,
		&branchName,
		&preferences,
	); err != nil {
		return nil, err
	}
	_ = dummyUserHash
	_ = dummyEmailHash
	_ = dummyPhoneHash

	user.PasswordHash = passwordHash
	user.PinHash = pinHash
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	if role.Valid {
		// v2: 优先 user_roles JOIN role_code (snake_case)；fallback users.role 字段。
		// 这里 normalize 成前端 hasPagePermission 期待的 PascalCase（与 auth_v2_handler::normalizeRole 同语义）。
		user.Role = normalizeUserRoleToV1(role.String)
	}
	user.AlarmLevels = alarmLevels
	user.AlarmChannels = alarmChannels
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.Tags = tags
	user.BranchID = branchID
	user.BranchName = branchName
	user.Preferences = preferences
	return &user, nil
}

// =============================================================================
// 单条查询
// =============================================================================

// GetUser 按 user_id 查；可选 tenantID（INET prefix CIDR 字符串）做租户限定。
func (r *PostgresUsersRepository) GetUser(ctx context.Context, tenantID, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, sql.ErrNoRows
	}
	q := `SELECT ` + usersSelectCols + ` ` + usersFromClause + ` WHERE u.user_id = $1::uuid`
	args := []any{userID}
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		q += ` AND u.tenant_prefix = $2::INET`
		args = append(args, tenantID)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user not found: user_id=%s", userID)
	}
	return r.scanUser(rows)
}

// GetUserByAccount 按 (tenant, username) 查。
func (r *PostgresUsersRepository) GetUserByAccount(ctx context.Context, tenantID, account string) (*domain.User, error) {
	if account == "" {
		return nil, sql.ErrNoRows
	}
	q := `SELECT ` + usersSelectCols + ` ` + usersFromClause + ` WHERE u.username = $1`
	args := []any{account}
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		q += ` AND u.tenant_prefix = $2::INET`
		args = append(args, tenantID)
	}
	q += ` LIMIT 1`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by account: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user not found: account=%s", account)
	}
	return r.scanUser(rows)
}

// GetUserByEmail v2 schema 删除了 email_hash 列；hash 查询不再支持。
// 兼容期：返回 sql.ErrNoRows，让 caller fallback 到 username 查询路径。
func (r *PostgresUsersRepository) GetUserByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.User, error) {
	return nil, sql.ErrNoRows
}

// GetUserByPhone v2 schema 同样删除了 phone_hash；不支持。
func (r *PostgresUsersRepository) GetUserByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.User, error) {
	return nil, sql.ErrNoRows
}

// =============================================================================
// 列表 + 过滤
// =============================================================================

// ListUsers v2: filter.BranchName/BranchNameNull/BranchIDs/Tag 在 v2 schema 不再支持；忽略不报错。
// Role/Status/Search 仍工作。
func (r *PostgresUsersRepository) ListUsers(ctx context.Context, tenantID string, filters UserFilters, page, size int) ([]*domain.User, int, error) {
	where := []string{}
	args := []any{}
	argIdx := 1

	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		where = append(where, fmt.Sprintf("u.tenant_prefix = $%d::INET", argIdx))
		args = append(args, tenantID)
		argIdx++
	}
	if filters.Role != "" {
		where = append(where, fmt.Sprintf("u.role = $%d", argIdx))
		args = append(args, filters.Role)
		argIdx++
	}
	if filters.Status != "" {
		where = append(where, fmt.Sprintf("COALESCE(u.status,'active') = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Search != "" {
		pat := "%" + strings.ToLower(filters.Search) + "%"
		where = append(where, fmt.Sprintf(`(
			LOWER(u.username) LIKE $%d OR
			LOWER(COALESCE(u.nickname,'')) LIKE $%d OR
			LOWER(COALESCE(u.email,'')) LIKE $%d OR
			LOWER(COALESCE(u.phone,'')) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, pat)
		argIdx++
	}
	// filter.BranchName / BranchNameNull / BranchIDs / Tag — v2 schema 不再支持，忽略

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) `+usersFromClause+` `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	listQ := `SELECT ` + usersSelectCols + ` ` + usersFromClause + ` ` + whereClause +
		fmt.Sprintf(` ORDER BY u.username ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return users, total, nil
}

// =============================================================================
// 写入
// =============================================================================

// CreateUser 在 v2 users 表插入新员工。
//
// 字段映射：
//   - tenantID → tenant_prefix INET（必填；非 INET 字符串拒绝）
//   - user.UserAccount → username
//   - user.PasswordHash bytea (sha256 hex bytes from frontend) → bcrypt → password_hash varchar
//   - user.Nickname/Email/Phone/Role/Status → 同名列
//   - 其它 v1-only 字段（PinHash bytea / AlarmLevels / Tags / Preferences）忽略
func (r *PostgresUsersRepository) CreateUser(ctx context.Context, tenantID string, user *domain.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is required")
	}
	if user.UserAccount == "" {
		return "", fmt.Errorf("username is required")
	}
	if tenantID != "" && !looksLikeINETPrefix(tenantID) {
		return "", fmt.Errorf("tenant_id %q is not a v2 INET prefix", tenantID)
	}

	// 双层 hash + 检索：
	//   user.PasswordHash 由调用方传入 = sha256(plaintext) bytes（前端约定形态）
	//   - password_hash      = bcrypt(sha256_hex)  -> 真正登录验证（抗暴力）
	//   - password_check_hash = sha256_bytes        -> (username, password_check_hash) 反向定位 + admin 类全局唯一
	var passwordHash sql.NullString
	var passwordCheckHash []byte
	if len(user.PasswordHash) > 0 {
		hexed := hex.EncodeToString(user.PasswordHash)
		bc, err := bcrypt.GenerateFromPassword([]byte(hexed), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("bcrypt password: %w", err)
		}
		passwordHash = sql.NullString{String: string(bc), Valid: true}
		passwordCheckHash = user.PasswordHash // 直接是 sha256(plain) bytes
	}

	var pinHash sql.NullString
	if len(user.PinHash) > 0 {
		hexed := hex.EncodeToString(user.PinHash)
		bc, err := bcrypt.GenerateFromPassword([]byte(hexed), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("bcrypt pin: %w", err)
		}
		pinHash = sql.NullString{String: string(bc), Valid: true}
	}

	var tenantPrefixArg any
	if tenantID != "" {
		tenantPrefixArg = tenantID
	}

	role := strings.TrimSpace(user.Role)
	status := user.Status
	if status == "" {
		status = "active"
	}

	nickname := nullToString(user.Nickname)
	if nickname == "" {
		// v2 schema NOT NULL；用 username 兜底
		nickname = user.UserAccount
	}

	var userID string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (tenant_prefix, username, password_hash, password_check_hash, mobile_pin_hash,
		                   nickname, email, phone, role, status)
		VALUES ($1::INET, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), $10)
		RETURNING user_id::text
	`,
		tenantPrefixArg, user.UserAccount,
		passwordHash, passwordCheckHash, pinHash,
		nickname,
		nullToString(user.Email),
		nullToString(user.Phone),
		role,
		status,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			// (tenant_prefix, username) 撞 → 重名；password_check_hash 撞（admin 类）→ 全局密码冲突
			if strings.Contains(err.Error(), "uniq_admin_password_check") {
				return "", fmt.Errorf("password collision: another tenant admin uses the same plaintext password")
			}
			return "", fmt.Errorf("username already exists in this tenant: %q", user.UserAccount)
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

// UpdateUser 更新 v2 users 表（部分字段）。
func (r *PostgresUsersRepository) UpdateUser(ctx context.Context, tenantID, userID string, user *domain.User) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if user == nil {
		return fmt.Errorf("user is required")
	}

	updates := []string{}
	args := []any{userID}
	argIdx := 2

	if user.UserAccount != "" {
		updates = append(updates, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, user.UserAccount)
		argIdx++
	}
	if len(user.PasswordHash) > 0 {
		bc, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(user.PasswordHash)), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("bcrypt password: %w", err)
		}
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, string(bc))
		argIdx++
		updates = append(updates, fmt.Sprintf("password_check_hash = $%d", argIdx))
		args = append(args, user.PasswordHash) // sha256(plain) bytes — 反向定位 + admin 类全局唯一
		argIdx++
		updates = append(updates, "password_set_at = NOW()", "must_change_password = false")
	}
	if len(user.PinHash) > 0 {
		bc, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(user.PinHash)), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("bcrypt pin: %w", err)
		}
		updates = append(updates, fmt.Sprintf("mobile_pin_hash = $%d", argIdx))
		args = append(args, string(bc))
		argIdx++
	}
	if user.Nickname.Valid && user.Nickname.String != "" {
		updates = append(updates, fmt.Sprintf("nickname = $%d", argIdx))
		args = append(args, user.Nickname.String)
		argIdx++
	}
	if user.Email.Valid {
		// Valid + 空字符串 → 清空
		updates = append(updates, fmt.Sprintf("email = NULLIF($%d, '')", argIdx))
		args = append(args, user.Email.String)
		argIdx++
	}
	if user.Phone.Valid {
		updates = append(updates, fmt.Sprintf("phone = NULLIF($%d, '')", argIdx))
		args = append(args, user.Phone.String)
		argIdx++
	}
	if user.Role != "" {
		updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, user.Role)
		argIdx++
	}
	if user.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, user.Status)
		argIdx++
	}
	updates = append(updates, "updated_at = NOW()")

	if len(updates) == 1 {
		return nil // only updated_at, nothing meaningful
	}

	q := fmt.Sprintf(`UPDATE users SET %s WHERE user_id = $1::uuid`, strings.Join(updates, ", "))
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		q += fmt.Sprintf(` AND tenant_prefix = $%d::INET`, argIdx)
		args = append(args, tenantID)
	}
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		if isUniqueViolation(err) {
			if strings.Contains(err.Error(), "uniq_admin_password_check") {
				return fmt.Errorf("password collision: another tenant admin uses the same plaintext password")
			}
			return fmt.Errorf("username already exists in this tenant: %q", user.UserAccount)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found: user_id=%s", userID)
	}
	return nil
}

// DeleteUser 软删除（status='deleted'）。v2 schema 仍有 status 列。
func (r *PostgresUsersRepository) DeleteUser(ctx context.Context, tenantID, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	q := `UPDATE users SET status = 'deleted', updated_at = NOW() WHERE user_id = $1::uuid`
	args := []any{userID}
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		q += ` AND tenant_prefix = $2::INET`
		args = append(args, tenantID)
	}
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found: user_id=%s", userID)
	}
	return nil
}

// =============================================================================
// 权限 + 唯一性检查
// =============================================================================

// GetResourcePermission 转发到 permission_utils.GetResourcePermission（v2 RBAC）。
func (r *PostgresUsersRepository) GetResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	pc, err := callV2GetResourcePermission(r.db, ctx, roleCode, resourceType, permissionType)
	if err != nil {
		return nil, err
	}
	return &PermissionCheck{AssignedOnly: pc.AssignedOnly, BranchOnly: pc.BranchOnly}, nil
}

// CheckEmailUniqueness 在同一租户内 email 不重复。v2 schema email 列存明文。
func (r *PostgresUsersRepository) CheckEmailUniqueness(ctx context.Context, tenantID, email, excludeUserID string) error {
	if email == "" {
		return nil
	}
	args := []any{email}
	q := `SELECT EXISTS (SELECT 1 FROM users WHERE LOWER(email) = LOWER($1)`
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		args = append(args, tenantID)
		q += fmt.Sprintf(` AND tenant_prefix = $%d::INET`, len(args))
	}
	if excludeUserID != "" {
		args = append(args, excludeUserID)
		q += fmt.Sprintf(` AND user_id <> $%d::uuid`, len(args))
	}
	q += `)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&exists); err != nil {
		return fmt.Errorf("check email uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("email already in use: %s", email)
	}
	return nil
}

// CheckPhoneUniqueness 同 email。
func (r *PostgresUsersRepository) CheckPhoneUniqueness(ctx context.Context, tenantID, phone, excludeUserID string) error {
	if phone == "" {
		return nil
	}
	args := []any{phone}
	q := `SELECT EXISTS (SELECT 1 FROM users WHERE phone = $1`
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		args = append(args, tenantID)
		q += fmt.Sprintf(` AND tenant_prefix = $%d::INET`, len(args))
	}
	if excludeUserID != "" {
		args = append(args, excludeUserID)
		q += fmt.Sprintf(` AND user_id <> $%d::uuid`, len(args))
	}
	q += `)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&exists); err != nil {
		return fmt.Errorf("check phone uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("phone already in use: %s", phone)
	}
	return nil
}

// CheckAdminCredentialDuplicate v2 没有 hash bytea 凭据；bcrypt 抗碰撞极强，跨租户不冲突。直接放行。
func (r *PostgresUsersRepository) CheckAdminCredentialDuplicate(ctx context.Context, accountHash, passwordHash []byte, excludeUserID string) error {
	return nil
}

// =============================================================================
// helpers
// =============================================================================

// normalizeUserRoleToV1 把 v2 role_code (snake_case) 或 v1 字段值 normalize 成前端期待的 PascalCase。
// 与 http/auth_v2_handler.go::normalizeRole 同语义；保留独立副本避免跨包依赖。
func normalizeUserRoleToV1(s string) string {
	s = strings.TrimSpace(s)
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
	case "resident":
		return "Resident"
	case "it":
		return "IT"
	}
	return s
}

func nullToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate") || strings.Contains(s, "violates unique constraint")
}

// callV2GetResourcePermission 调用 http 层的 GetResourcePermission（v2 RBAC）。
// 由于 repo 层不能 import http 层，这里用 db 直接复刻逻辑（保持与 permission_utils.go 一致）。
type v2PermissionCheck struct {
	AssignedOnly bool
	BranchOnly   bool
}

func callV2GetResourcePermission(db *sql.DB, ctx context.Context, roleCode, resourceType, permissionType string) (*v2PermissionCheck, error) {
	v2Role := mapV1RoleToV2(roleCode)
	action := permissionVerb(permissionType)

	target := resourceType + "." + action
	resourceAll := resourceType + ".*"
	const tenantAll = "tenant.*"
	const platformAll = "*"

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM role_permissions rp
			  JOIN roles r ON r.role_id = rp.role_id
			 WHERE r.role_code = $1
			   AND rp.permission IN ($2, $3, $4, $5)
		)
	`, v2Role, target, resourceAll, tenantAll, platformAll).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &v2PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	return &v2PermissionCheck{AssignedOnly: false, BranchOnly: false}, nil
}

// mapV1RoleToV2 与 http/permission_utils.go::mapRoleToV2 同语义；这里保留独立副本避免跨包依赖。
func mapV1RoleToV2(role string) string {
	switch role {
	case "SystemAdmin", "SystemOperator":
		return "platform_admin"
	case "Admin":
		return "tenant_admin"
	case "Manager":
		return "manager"
	case "Nurse":
		return "nurse"
	case "Caregiver":
		return "caregiver"
	case "Family":
		return "family"
	case "Viewer":
		return "viewer"
	}
	return strings.ToLower(role)
}

func permissionVerb(p string) string {
	switch strings.ToUpper(p) {
	case "R":
		return "read"
	case "C":
		return "create"
	case "U":
		return "update"
	case "D":
		return "delete"
	}
	return strings.ToLower(p)
}
