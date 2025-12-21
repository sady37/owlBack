package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"github.com/lib/pq"
)

// PostgresUsersRepository 用户Repository实现（强类型版本）
// 实现UsersRepository接口，使用domain.User领域模型
type PostgresUsersRepository struct {
	db *sql.DB
}

// NewPostgresUsersRepository 创建用户Repository
func NewPostgresUsersRepository(db *sql.DB) *PostgresUsersRepository {
	return &PostgresUsersRepository{db: db}
}

// 确保实现了接口
var _ UsersRepository = (*PostgresUsersRepository)(nil)

// GetUser 获取用户基本信息
func (r *PostgresUsersRepository) GetUser(ctx context.Context, tenantID, userID string) (*domain.User, error) {
	if tenantID == "" || userID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			user_id::text,
			tenant_id::text,
			user_account,
			user_account_hash,
			password_hash,
			pin_hash,
			nickname,
			email,
			phone,
			email_hash,
			phone_hash,
			role,
			status,
			alarm_levels,
			alarm_channels,
			alarm_scope,
			last_login_at,
			tags::text,
			branch_tag,
			preferences::text
		FROM users
		WHERE tenant_id = $1 AND user_id = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchTag, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&user.UserAccountHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&emailHash,
		&phoneHash,
		&user.Role,
		&user.Status,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchTag,
		&preferences,
	)
	if err != nil {
		return nil, err
	}

	// 处理可空字段
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.V
	}
	if pinHash.Valid {
		user.PinHash = pinHash.V
	}
	if emailHash.Valid {
		user.EmailHash = emailHash.V
	}
	if phoneHash.Valid {
		user.PhoneHash = phoneHash.V
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.BranchTag = branchTag
	user.Preferences = preferences

	// 处理数组字段
	if alarmLevels != nil {
		user.AlarmLevels = alarmLevels
	}
	if alarmChannels != nil {
		user.AlarmChannels = alarmChannels
	}

	// 处理tags JSONB数组
	if tags.Valid && tags.String != "" {
		var tagsArray []string
		if err := json.Unmarshal([]byte(tags.String), &tagsArray); err == nil {
			user.Tags = sql.NullString{String: tags.String, Valid: true}
		}
	}

	return &user, nil
}

// GetUserByAccount 根据账号获取用户
func (r *PostgresUsersRepository) GetUserByAccount(ctx context.Context, tenantID, account string) (*domain.User, error) {
	if tenantID == "" || account == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			user_id::text,
			tenant_id::text,
			user_account,
			user_account_hash,
			password_hash,
			pin_hash,
			nickname,
			email,
			phone,
			email_hash,
			phone_hash,
			role,
			status,
			alarm_levels,
			alarm_channels,
			alarm_scope,
			last_login_at,
			tags::text,
			branch_tag,
			preferences::text
		FROM users
		WHERE tenant_id = $1 AND user_account = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchTag, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID, account).Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&user.UserAccountHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&emailHash,
		&phoneHash,
		&user.Role,
		&user.Status,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchTag,
		&preferences,
	)
	if err != nil {
		return nil, err
	}

	// 处理可空字段（与GetUser相同）
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.V
	}
	if pinHash.Valid {
		user.PinHash = pinHash.V
	}
	if emailHash.Valid {
		user.EmailHash = emailHash.V
	}
	if phoneHash.Valid {
		user.PhoneHash = phoneHash.V
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.BranchTag = branchTag
	user.Preferences = preferences

	if alarmLevels != nil {
		user.AlarmLevels = alarmLevels
	}
	if alarmChannels != nil {
		user.AlarmChannels = alarmChannels
	}

	if tags.Valid && tags.String != "" {
		user.Tags = sql.NullString{String: tags.String, Valid: true}
	}

	return &user, nil
}

// GetUserByEmail 根据email_hash获取用户
func (r *PostgresUsersRepository) GetUserByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.User, error) {
	if tenantID == "" || len(emailHash) == 0 {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			user_id::text,
			tenant_id::text,
			user_account,
			user_account_hash,
			password_hash,
			pin_hash,
			nickname,
			email,
			phone,
			email_hash,
			phone_hash,
			role,
			status,
			alarm_levels,
			alarm_channels,
			alarm_scope,
			last_login_at,
			tags::text,
			branch_tag,
			preferences::text
		FROM users
		WHERE tenant_id = $1 AND email_hash = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHashDB, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchTag, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID, emailHash).Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&user.UserAccountHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&emailHashDB,
		&phoneHash,
		&user.Role,
		&user.Status,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchTag,
		&preferences,
	)
	if err != nil {
		return nil, err
	}

	// 处理可空字段（与GetUser相同）
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.V
	}
	if pinHash.Valid {
		user.PinHash = pinHash.V
	}
	if emailHashDB.Valid {
		user.EmailHash = emailHashDB.V
	}
	if phoneHash.Valid {
		user.PhoneHash = phoneHash.V
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.BranchTag = branchTag
	user.Preferences = preferences

	if alarmLevels != nil {
		user.AlarmLevels = alarmLevels
	}
	if alarmChannels != nil {
		user.AlarmChannels = alarmChannels
	}

	if tags.Valid && tags.String != "" {
		user.Tags = sql.NullString{String: tags.String, Valid: true}
	}

	return &user, nil
}

// GetUserByPhone 根据phone_hash获取用户
func (r *PostgresUsersRepository) GetUserByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.User, error) {
	if tenantID == "" || len(phoneHash) == 0 {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			user_id::text,
			tenant_id::text,
			user_account,
			user_account_hash,
			password_hash,
			pin_hash,
			nickname,
			email,
			phone,
			email_hash,
			phone_hash,
			role,
			status,
			alarm_levels,
			alarm_channels,
			alarm_scope,
			last_login_at,
			tags::text,
			branch_tag,
			preferences::text
		FROM users
		WHERE tenant_id = $1 AND phone_hash = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHashDB sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchTag, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID, phoneHash).Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&user.UserAccountHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&emailHash,
		&phoneHashDB,
		&user.Role,
		&user.Status,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchTag,
		&preferences,
	)
	if err != nil {
		return nil, err
	}

	// 处理可空字段（与GetUser相同）
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.V
	}
	if pinHash.Valid {
		user.PinHash = pinHash.V
	}
	if emailHash.Valid {
		user.EmailHash = emailHash.V
	}
	if phoneHashDB.Valid {
		user.PhoneHash = phoneHashDB.V
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.BranchTag = branchTag
	user.Preferences = preferences

	if alarmLevels != nil {
		user.AlarmLevels = alarmLevels
	}
	if alarmChannels != nil {
		user.AlarmChannels = alarmChannels
	}

	if tags.Valid && tags.String != "" {
		user.Tags = sql.NullString{String: tags.String, Valid: true}
	}

	return &user, nil
}

// ListUsers 列出用户
func (r *PostgresUsersRepository) ListUsers(ctx context.Context, tenantID string, filters UserFilters, page, size int) ([]*domain.User, int, error) {
	if tenantID == "" {
		return []*domain.User{}, 0, nil
	}

	// 构建WHERE子句
	where := []string{"u.tenant_id = $1"}
	args := []any{tenantID}
	argIdx := 2

	if filters.Role != "" {
		where = append(where, fmt.Sprintf("u.role = $%d", argIdx))
		args = append(args, filters.Role)
		argIdx++
	}
	if filters.Status != "" {
		where = append(where, fmt.Sprintf("u.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.BranchTagNull {
		// 匹配 branch_tag IS NULL OR branch_tag = '-'
		where = append(where, "(u.branch_tag IS NULL OR u.branch_tag = '-')")
	} else if filters.BranchTag != "" {
		where = append(where, fmt.Sprintf("u.branch_tag = $%d", argIdx))
		args = append(args, filters.BranchTag)
		argIdx++
	}
	if filters.Tag != "" {
		where = append(where, fmt.Sprintf("u.tags ? $%d", argIdx))
		args = append(args, filters.Tag)
		argIdx++
	}
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("(u.user_account ILIKE $%d OR COALESCE(u.nickname,'') ILIKE $%d OR COALESCE(u.email,'') ILIKE $%d OR COALESCE(u.phone,'') ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	// 计算总数
	countQuery := "SELECT COUNT(*) FROM users u WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	query := `
		SELECT 
			u.user_id::text,
			u.tenant_id::text,
			u.user_account,
			u.user_account_hash,
			u.password_hash,
			u.pin_hash,
			u.nickname,
			u.email,
			u.phone,
			u.email_hash,
			u.phone_hash,
			u.role,
			u.status,
			u.alarm_levels,
			u.alarm_channels,
			u.alarm_scope,
			u.last_login_at,
			u.tags::text,
			u.branch_tag,
			u.preferences::text
		FROM users u
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY u.user_account ASC
		LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)

	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// scanUser 从rows扫描User
func (r *PostgresUsersRepository) scanUser(rows *sql.Rows) (*domain.User, error) {
	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchTag, preferences sql.NullString
	var lastLoginAt sql.NullTime
	var alarmLevels, alarmChannels pq.StringArray

	err := rows.Scan(
		&user.UserID,
		&user.TenantID,
		&user.UserAccount,
		&user.UserAccountHash,
		&passwordHash,
		&pinHash,
		&nickname,
		&email,
		&phone,
		&emailHash,
		&phoneHash,
		&user.Role,
		&user.Status,
		&alarmLevels,
		&alarmChannels,
		&alarmScope,
		&lastLoginAt,
		&tags,
		&branchTag,
		&preferences,
	)
	if err != nil {
		return nil, err
	}

	// 处理可空字段
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.V
	}
	if pinHash.Valid {
		user.PinHash = pinHash.V
	}
	if emailHash.Valid {
		user.EmailHash = emailHash.V
	}
	if phoneHash.Valid {
		user.PhoneHash = phoneHash.V
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.AlarmScope = alarmScope
	user.LastLoginAt = lastLoginAt
	user.BranchTag = branchTag
	user.Preferences = preferences

	if alarmLevels != nil {
		user.AlarmLevels = alarmLevels
	}
	if alarmChannels != nil {
		user.AlarmChannels = alarmChannels
	}

	if tags.Valid && tags.String != "" {
		user.Tags = sql.NullString{String: tags.String, Valid: true}
	}

	return &user, nil
}

// CreateUser 创建用户
func (r *PostgresUsersRepository) CreateUser(ctx context.Context, tenantID string, user *domain.User) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if user == nil {
		return "", fmt.Errorf("user is required")
	}

	// 验证必填字段
	if user.UserAccount == "" {
		return "", fmt.Errorf("user_account is required")
	}
	if len(user.UserAccountHash) == 0 {
		return "", fmt.Errorf("user_account_hash is required")
	}
	if user.Role == "" {
		return "", fmt.Errorf("role is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 处理默认值
	status := user.Status
	if status == "" {
		status = "active" // DB默认值
	}

	// 处理tags JSONB数组（验证JSON有效性）
	var tagsArg any = nil
	if user.Tags.Valid && user.Tags.String != "" {
		// 验证是否为有效JSON数组
		var tagsArray []string
		if err := json.Unmarshal([]byte(user.Tags.String), &tagsArray); err != nil {
			return "", fmt.Errorf("invalid tags JSON: %w", err)
		}
		tagsArg = user.Tags.String
	}

	// 处理preferences JSONB
	var preferencesArg any = nil
	if user.Preferences.Valid && user.Preferences.String != "" {
		preferencesArg = user.Preferences.String
	} else {
		preferencesArg = "{}" // DB默认值
	}

	// 插入users表（让DB自动生成user_id）
	query := `
		INSERT INTO users (
			tenant_id, user_account, user_account_hash,
			password_hash, pin_hash, nickname, email, phone,
			email_hash, phone_hash, role, status,
			alarm_levels, alarm_channels, alarm_scope,
			last_login_at, tags, branch_tag, preferences
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17::jsonb, $18, $19::jsonb
		)
		RETURNING user_id::text
	`

	var passwordHash, pinHash, emailHash, phoneHash any
	if len(user.PasswordHash) > 0 {
		passwordHash = user.PasswordHash
	}
	if len(user.PinHash) > 0 {
		pinHash = user.PinHash
	}
	if len(user.EmailHash) > 0 {
		emailHash = user.EmailHash
	}
	if len(user.PhoneHash) > 0 {
		phoneHash = user.PhoneHash
	}

	// 辅助函数：将sql.NullString转为any
	toAnyString := func(ns sql.NullString) any {
		if ns.Valid {
			return ns.String
		}
		return nil
	}
	// 辅助函数：将sql.NullTime转为any
	toAnyTime := func(nt sql.NullTime) any {
		if nt.Valid {
			return nt.Time
		}
		return nil
	}

	var userID string
	err = tx.QueryRowContext(ctx, query,
		tenantID,
		user.UserAccount,
		user.UserAccountHash,
		passwordHash,
		pinHash,
		toAnyString(user.Nickname),
		toAnyString(user.Email),
		toAnyString(user.Phone),
		emailHash,
		phoneHash,
		user.Role,
		status,
		pq.Array(user.AlarmLevels),
		pq.Array(user.AlarmChannels),
		toAnyString(user.AlarmScope),
		toAnyTime(user.LastLoginAt),
		tagsArg,
		toAnyString(user.BranchTag),
		preferencesArg,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	// 同步tags到tags_catalog目录（如果tags存在）
	if user.Tags.Valid && user.Tags.String != "" {
		var tagsArray []string
		if err := json.Unmarshal([]byte(user.Tags.String), &tagsArray); err == nil {
			for _, tagName := range tagsArray {
				if tagName != "" {
					_, err = tx.ExecContext(ctx,
						`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
						tenantID, tagName, "user_tag",
					)
					if err != nil {
						return "", fmt.Errorf("failed to sync tag %s to catalog: %w", tagName, err)
					}
				}
			}
		}
	}

	// 同步branch_tag到tags_catalog目录（如果branch_tag存在）
	if user.BranchTag.Valid && user.BranchTag.String != "" {
		_, err = tx.ExecContext(ctx,
			`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
			tenantID, user.BranchTag.String, "branch_tag",
		)
		if err != nil {
			return "", fmt.Errorf("failed to sync branch_tag to catalog: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userID, nil
}

// UpdateUser 更新用户
func (r *PostgresUsersRepository) UpdateUser(ctx context.Context, tenantID, userID string, user *domain.User) error {
	if tenantID == "" || userID == "" {
		return fmt.Errorf("tenant_id and user_id are required")
	}
	if user == nil {
		return fmt.Errorf("user is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取旧的tags和branch_tag
	var oldTags, oldBranchTag sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT tags::text, branch_tag FROM users WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userID,
	).Scan(&oldTags, &oldBranchTag)
	if err != nil {
		return fmt.Errorf("failed to get old user data: %w", err)
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID, userID}
	argIdx := 3

	if user.UserAccount != "" {
		updates = append(updates, fmt.Sprintf("user_account = $%d", argIdx))
		args = append(args, user.UserAccount)
		argIdx++
	}
	if len(user.UserAccountHash) > 0 {
		updates = append(updates, fmt.Sprintf("user_account_hash = $%d", argIdx))
		args = append(args, user.UserAccountHash)
		argIdx++
	}
	if len(user.PasswordHash) > 0 {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, user.PasswordHash)
		argIdx++
	}
	if len(user.PinHash) > 0 {
		updates = append(updates, fmt.Sprintf("pin_hash = $%d", argIdx))
		args = append(args, user.PinHash)
		argIdx++
	}
	if user.Nickname.Valid {
		updates = append(updates, fmt.Sprintf("nickname = $%d", argIdx))
		args = append(args, user.Nickname)
		argIdx++
	}
	// Email 和 EmailHash：常规 CRUD 逻辑
	// Service 层已经处理了所有业务逻辑，这里只需要根据字段值决定是否更新
	// 如果 Email.Valid = true，更新 email；如果 Email.Valid = false，设置为 NULL（通过 nil 参数）
	// 如果 EmailHash 有值，更新 hash；如果 EmailHash 为 nil，设置为 NULL
	if user.Email.Valid {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, user.Email)
		argIdx++
	} else if user.EmailHash != nil {
		// Email.Valid = false 且 EmailHash 被设置，说明要删除 email 但保留 hash
		// 这种情况在 Service 层已经处理，这里只需要设置 email 为 NULL
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	}
	// 更新 email_hash（如果被设置）
	if user.EmailHash != nil {
		if len(user.EmailHash) > 0 {
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			args = append(args, user.EmailHash)
			argIdx++
		} else {
			// EmailHash 为 nil slice，设置为 NULL
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		}
	}

	// Phone 和 PhoneHash：同 Email 逻辑
	if user.Phone.Valid {
		updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, user.Phone)
		argIdx++
	} else if user.PhoneHash != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	}
	if user.PhoneHash != nil {
		if len(user.PhoneHash) > 0 {
			updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
			args = append(args, user.PhoneHash)
			argIdx++
		} else {
			updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		}
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
	if user.AlarmLevels != nil {
		updates = append(updates, fmt.Sprintf("alarm_levels = $%d", argIdx))
		args = append(args, pq.Array(user.AlarmLevels))
		argIdx++
	}
	if user.AlarmChannels != nil {
		updates = append(updates, fmt.Sprintf("alarm_channels = $%d", argIdx))
		args = append(args, pq.Array(user.AlarmChannels))
		argIdx++
	}
	if user.AlarmScope.Valid {
		updates = append(updates, fmt.Sprintf("alarm_scope = $%d", argIdx))
		args = append(args, user.AlarmScope)
		argIdx++
	}
	if user.LastLoginAt.Valid {
		updates = append(updates, fmt.Sprintf("last_login_at = $%d", argIdx))
		args = append(args, user.LastLoginAt)
		argIdx++
	}
	if user.Tags.Valid {
		// 验证tags JSON有效性
		var tagsArray []string
		if err := json.Unmarshal([]byte(user.Tags.String), &tagsArray); err != nil {
			return fmt.Errorf("invalid tags JSON: %w", err)
		}
		updates = append(updates, fmt.Sprintf("tags = $%d::jsonb", argIdx))
		args = append(args, user.Tags.String)
		argIdx++
	}
	if user.BranchTag.Valid {
		updates = append(updates, fmt.Sprintf("branch_tag = $%d", argIdx))
		args = append(args, user.BranchTag)
		argIdx++
	}
	if user.Preferences.Valid {
		// 验证preferences JSON有效性
		var prefsMap map[string]any
		if err := json.Unmarshal([]byte(user.Preferences.String), &prefsMap); err != nil {
			return fmt.Errorf("invalid preferences JSON: %w", err)
		}
		updates = append(updates, fmt.Sprintf("preferences = $%d::jsonb", argIdx))
		args = append(args, user.Preferences.String)
		argIdx++
	}

	if len(updates) == 0 {
		return nil // 没有需要更新的字段
	}

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE tenant_id = $1 AND user_id = $2`,
		strings.Join(updates, ", "),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// 处理tags变化：同步到tags_catalog目录
	if user.Tags.Valid {
		var newTagsArray []string
		if err := json.Unmarshal([]byte(user.Tags.String), &newTagsArray); err == nil {
			// 获取旧tags数组
			var oldTagsArray []string
			if oldTags.Valid && oldTags.String != "" {
				json.Unmarshal([]byte(oldTags.String), &oldTagsArray)
			}

			// 找出新增的tags
			newTagMap := make(map[string]bool)
			for _, tag := range newTagsArray {
				newTagMap[tag] = true
			}
			for _, tag := range oldTagsArray {
				if !newTagMap[tag] {
					// 旧tag不在新tags中，但不需要从目录删除（目录只是记录）
				}
			}

			// 添加新tags到目录
			for _, tagName := range newTagsArray {
				if tagName != "" {
					_, err = tx.ExecContext(ctx,
						`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
						tenantID, tagName, "user_tag",
					)
					if err != nil {
						return fmt.Errorf("failed to sync tag %s to catalog: %w", tagName, err)
					}
				}
			}
		}
	}

	// 处理branch_tag变化：同步到tags_catalog目录
	if user.BranchTag.Valid {
		oldBranchTagValue := ""
		if oldBranchTag.Valid {
			oldBranchTagValue = oldBranchTag.String
		}
		if oldBranchTagValue != user.BranchTag.String {
			// 如果新branch_tag不为空，添加到目录
			if user.BranchTag.String != "" {
				_, err = tx.ExecContext(ctx,
					`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
					tenantID, user.BranchTag.String, "branch_tag",
				)
				if err != nil {
					return fmt.Errorf("failed to sync branch_tag to catalog: %w", err)
				}
			}
			// 注意：不需要从旧branch_tag移除，因为tag_objects已删除
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (r *PostgresUsersRepository) DeleteUser(ctx context.Context, tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return fmt.Errorf("tenant_id and user_id are required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 注意：不需要清理tags_catalog，因为没有反向索引（tag_objects已删除）

	// 删除users记录
	_, err = tx.ExecContext(ctx,
		`DELETE FROM users WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SyncUserTagsToCatalog 同步用户tags到tags_catalog目录
func (r *PostgresUsersRepository) SyncUserTagsToCatalog(ctx context.Context, tenantID, userID string, tags []string) error {
	if tenantID == "" || userID == "" {
		return fmt.Errorf("tenant_id and user_id are required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 遍历tags数组，逐个调用upsert_tag_to_catalog
	for _, tagName := range tags {
		if tagName != "" {
			_, err = tx.ExecContext(ctx,
				`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
				tenantID, tagName, "user_tag",
			)
			if err != nil {
				return fmt.Errorf("failed to sync tag %s to catalog: %w", tagName, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetResourcePermission 查询资源权限配置
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
func (r *PostgresUsersRepository) GetResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	// SystemTenantID 常量
	const SystemTenantID = "00000000-0000-0000-0000-000000000001"

	var assignedOnly, branchOnly bool
	err := r.db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(assigned_only, FALSE) as assigned_only,
			COALESCE(branch_only, FALSE) as branch_only
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID, roleCode, resourceType, permissionType,
	).Scan(&assignedOnly, &branchOnly)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	return &PermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// CheckEmailUniqueness 检查 email 唯一性
func (r *PostgresUsersRepository) CheckEmailUniqueness(ctx context.Context, tenantID, email, excludeUserID string) error {
	if email == "" {
		return nil
	}
	var query string
	var args []interface{}
	if excludeUserID != "" {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND email = $2 AND user_id::text != $3`
		args = []interface{}{tenantID, email, excludeUserID}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND email = $2`
		args = []interface{}{tenantID, email}
	}
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("email already exists in this organization")
	}
	return nil
}

// CheckPhoneUniqueness 检查 phone 唯一性
func (r *PostgresUsersRepository) CheckPhoneUniqueness(ctx context.Context, tenantID, phone, excludeUserID string) error {
	if phone == "" {
		return nil
	}
	var query string
	var args []interface{}
	if excludeUserID != "" {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND phone = $2 AND user_id::text != $3`
		args = []interface{}{tenantID, phone, excludeUserID}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND phone = $2`
		args = []interface{}{tenantID, phone}
	}
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("phone already exists in this organization")
	}
	return nil
}

