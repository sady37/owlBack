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
			u.user_tags::text,
			ub.branch_id::text as branch_id,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.preferences::text
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ub2.branch_id FROM user_branches ub2
			LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id
			WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id 
			ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1
		) ub ON true
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
		WHERE u.tenant_id = $1 AND u.user_id = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchID, branchName, preferences sql.NullString
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
		&branchID,
		&branchName,
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
	user.BranchID = branchID
	user.BranchName = branchName
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
			u.user_tags::text,
			ub.branch_id::text as branch_id,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.preferences::text
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ub2.branch_id FROM user_branches ub2
			LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id
			WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id 
			ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1
		) ub ON true
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
		WHERE u.tenant_id = $1 AND u.user_account = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchID, branchName, preferences sql.NullString
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
		&branchID,
		&branchName,
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
	user.BranchID = branchID
	user.BranchName = branchName
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
			u.user_tags::text,
			ub.branch_id::text as branch_id,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.preferences::text
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ub2.branch_id FROM user_branches ub2
			LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id
			WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id 
			ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1
		) ub ON true
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
		WHERE u.tenant_id = $1 AND u.email_hash = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHashDB, phoneHash sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchID, branchName, preferences sql.NullString
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
		&branchID,
		&branchName,
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
	user.BranchID = branchID
	user.BranchName = branchName
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
			u.user_tags::text,
			ub.branch_id::text as branch_id,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.preferences::text
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ub2.branch_id FROM user_branches ub2
			LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id
			WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id 
			ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1
		) ub ON true
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
		WHERE u.tenant_id = $1 AND u.phone_hash = $2
	`

	var user domain.User
	var passwordHash, pinHash, emailHash, phoneHashDB sql.Null[[]byte]
	var nickname, email, phone, alarmScope, tags, branchID, branchName, preferences sql.NullString
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
		&branchID,
		&branchName,
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
	user.BranchID = branchID
	user.BranchName = branchName
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
	if filters.BranchNameNull {
		// 匹配没有主院区或主院区名称为 NULL、""、"-"（都视为空院区）
		where = append(where, "(ub.user_id IS NULL OR b.branch_name IS NULL OR b.branch_name = '' OR b.branch_name = 'default')")
	} else if len(filters.BranchIDs) > 0 {
		// 支持多个 branch_id 的 IN 查询（1 对多关系）
		// 使用 EXISTS 子查询，检查用户是否关联了任何一个指定的 branch_id
		placeholders := make([]string, len(filters.BranchIDs))
		for i, branchID := range filters.BranchIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, branchID)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_branches ub2 WHERE ub2.tenant_id = u.tenant_id AND ub2.user_id = u.user_id AND ub2.branch_id::text IN (%s))",
			strings.Join(placeholders, ", "),
		))
	} else if filters.BranchName != "" {
		// 单个 branch_name 精确匹配（主院区）
		where = append(where, fmt.Sprintf("b.branch_name = $%d", argIdx))
		args = append(args, filters.BranchName)
		argIdx++
	}
	if filters.Tag != "" {
		where = append(where, fmt.Sprintf("u.user_tags ? $%d", argIdx))
		args = append(args, filters.Tag)
		argIdx++
	}
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("(u.user_account ILIKE $%d OR COALESCE(u.nickname,'') ILIKE $%d OR COALESCE(u.email,'') ILIKE $%d OR COALESCE(u.phone,'') ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	// 计算总数
	countQuery := "SELECT COUNT(*) FROM users u LEFT JOIN LATERAL (SELECT ub2.branch_id FROM user_branches ub2 LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1) ub ON true LEFT JOIN branches b ON b.branch_id = ub.branch_id WHERE " + strings.Join(where, " AND ")
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
			u.user_tags::text,
			ub.branch_id::text as branch_id,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.preferences::text
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ub2.branch_id FROM user_branches ub2
			LEFT JOIN branches b2 ON b2.branch_id = ub2.branch_id
			WHERE ub2.user_id = u.user_id AND ub2.tenant_id = u.tenant_id 
			ORDER BY COALESCE(b2.branch_name, '') ASC LIMIT 1
		) ub ON true
		LEFT JOIN branches b ON b.branch_id = ub.branch_id
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
	var nickname, email, phone, alarmScope, tags, branchID, branchName, preferences sql.NullString
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
		&branchID,
		&branchName,
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
	user.BranchID = branchID
	user.BranchName = branchName
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
			last_login_at, user_tags, preferences
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17::jsonb, $18::jsonb
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
		preferencesArg,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	// 如果提供了 BranchID 或 BranchName，在 user_branches 表中创建主院区关联
	var branchIDToUse sql.NullString
	if user.BranchID.Valid && user.BranchID.String != "" {
		branchIDToUse = user.BranchID
	} else if user.BranchName.Valid && user.BranchName.String != "" {
		// 通过 branch_name 查找 branch_id
		err = tx.QueryRowContext(ctx,
			`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`,
			tenantID, user.BranchName.String,
		).Scan(&branchIDToUse)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("branch not found: branch_name '%s'", user.BranchName.String)
			}
			return "", fmt.Errorf("failed to find branch: %w", err)
		}
	}

	if branchIDToUse.Valid && branchIDToUse.String != "" {
		// 在 user_branches 表中创建院区关联
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_branches (tenant_id, user_id, branch_id)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (tenant_id, user_id, branch_id) DO NOTHING`,
			tenantID, userID, branchIDToUse.String,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create user branch association: %w", err)
		}
	}

	// 注意：tags_catalog 表已删除，不再需要同步 tags 到目录
	// 用户 tags 直接存储在 users.user_tags (JSONB) 字段中
	// 注意：branch_tag 不应该在这里创建
	// branch_tag 应该由前端在 TagList 页面创建（tag_name = "Branch"）
	// user 的 branch_tag 只是数据，不需要同步到 tags_catalog

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

	// 获取旧的tags（用于验证 JSON 格式，不再获取 branch_name，因为 user_branches 由 Service 层管理）
	var oldTags sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT u.user_tags::text FROM users u WHERE u.tenant_id = $1 AND u.user_id = $2`,
		tenantID, userID,
	).Scan(&oldTags)
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
	// Email 字段：独立处理
	// Service 层已经明确表达了意图：
	// - user.Email.Valid = true && user.Email.String != "" -> 更新 email 字段
	// - user.Email.Valid = true && user.Email.String == "" -> 删除 email 字段（设置为 NULL）
	// - user.Email 未设置（零值）-> 不更新 email 字段
	// Repository 层只执行指令，不做业务判断
	if user.Email.Valid {
		if user.Email.String != "" {
			// Valid == true && String != ""，更新字段
			updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
			args = append(args, user.Email.String)
			argIdx++
		} else {
			// Valid == true && String == ""，删除字段（设置为 NULL）
			updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		}
	}
	// 如果 Email 是零值（未设置），不处理（不更新）
	// EmailHash 字段：独立处理
	// Service 层已经明确表达了意图：
	// - user.EmailHash != nil 且 len > 0 -> 更新 email_hash 字段
	// - user.EmailHash != nil 且 len == 0 -> 删除 email_hash 字段（设置为 NULL）
	// - user.EmailHash == nil -> 不更新 email_hash 字段
	if user.EmailHash != nil {
		if len(user.EmailHash) > 0 {
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			args = append(args, user.EmailHash)
			argIdx++
		} else {
			// EmailHash 为空 slice，设置为 NULL
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		}
	}

	// Phone 字段：独立处理
	// Service 层已经明确表达了意图：
	// - user.Phone.Valid = true && user.Phone.String != "" -> 更新 phone 字段
	// - user.Phone.Valid = true && user.Phone.String == "" -> 删除 phone 字段（设置为 NULL）
	// - user.Phone 未设置（零值）-> 不更新 phone 字段
	// Repository 层只执行指令，不做业务判断
	if user.Phone.Valid {
		if user.Phone.String != "" {
			// Valid == true && String != ""，更新字段
			updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
			args = append(args, user.Phone.String)
			argIdx++
		} else {
			// Valid == true && String == ""，删除字段（设置为 NULL）
			updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		}
	}
	// 如果 Phone 是零值（未设置），不处理（不更新）
	// PhoneHash 字段：独立处理
	// Service 层已经明确表达了意图：
	// - user.PhoneHash != nil 且 len > 0 -> 更新 phone_hash 字段
	// - user.PhoneHash != nil 且 len == 0 -> 删除 phone_hash 字段（设置为 NULL）
	// - user.PhoneHash == nil -> 不更新 phone_hash 字段
	if user.PhoneHash != nil {
		if len(user.PhoneHash) > 0 {
			updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
			args = append(args, user.PhoneHash)
			argIdx++
		} else {
			// PhoneHash 为空 slice，设置为 NULL
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
	// AlarmLevels 字段：空数组 → NULL
	if user.AlarmLevels != nil {
		if len(user.AlarmLevels) == 0 {
			// 空数组 → NULL
			updates = append(updates, "alarm_levels = NULL")
		} else {
			updates = append(updates, fmt.Sprintf("alarm_levels = $%d", argIdx))
			args = append(args, pq.Array(user.AlarmLevels))
			argIdx++
		}
	}
	// AlarmChannels 字段：空数组 → NULL
	if user.AlarmChannels != nil {
		if len(user.AlarmChannels) == 0 {
			// 空数组 → NULL
			updates = append(updates, "alarm_channels = NULL")
		} else {
			updates = append(updates, fmt.Sprintf("alarm_channels = $%d", argIdx))
			args = append(args, pq.Array(user.AlarmChannels))
			argIdx++
		}
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
		updates = append(updates, fmt.Sprintf("user_tags = $%d::jsonb", argIdx))
		args = append(args, user.Tags.String)
		argIdx++
	}
	// 处理 branch 更新：更新 user_branches 表（在事务提交前处理）
	// 注意：这里不更新 users 表，而是在事务提交前更新 user_branches 表
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

	// 注意：tags_catalog 表已删除，不再需要同步 tags 到目录
	// 用户 tags 直接存储在 users.user_tags (JSONB) 字段中
	// tags 的更新已经通过上面的 UPDATE 语句完成

	// 注意：user_branches 表的更新现在由 Service 层的 updateUserBranches 函数统一管理
	// Repository 层的 UpdateUser 不再处理 user_branches 表的更新，避免与 Service 层逻辑冲突
	// users.branch_id 和 users.branch_name 字段仅用于向后兼容，通过 LEFT JOIN LATERAL 从 user_branches 表获取
	// 这些字段不应该在 Repository 层的 UpdateUser 中更新，因为它们是从 user_branches 表派生出来的

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

// GetResourcePermission 查询资源权限配置
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func (r *PostgresUsersRepository) GetResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	// SystemTenantID 常量
	const SystemTenantID = "00000000-0000-0000-0000-000000000001"

	var permissionScope string
	err := r.db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID, roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	// 将 permission_scope 转换为 assigned_only 和 branch_only 标志
	var assignedOnly, branchOnly bool
	switch permissionScope {
	case "A":
		// All (no restriction)
		assignedOnly = false
		branchOnly = false
	case "S":
		// assigned_only
		assignedOnly = true
		branchOnly = false
	case "B":
		// branch_only
		assignedOnly = false
		branchOnly = true
	default:
		// 未知值，返回最严格的权限（安全默认值）
		assignedOnly = true
		branchOnly = true
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

// CheckAdminCredentialDuplicate 跨租户检查 Admin/Manager 的 account_hash + password_hash 是否重复
func (r *PostgresUsersRepository) CheckAdminCredentialDuplicate(ctx context.Context, accountHash, passwordHash []byte, excludeUserID string) error {
	var query string
	var args []interface{}
	if excludeUserID != "" {
		query = `SELECT COUNT(*) FROM users WHERE user_account_hash = $1 AND password_hash = $2 AND role IN ('Admin', 'Manager') AND user_id::text != $3`
		args = []interface{}{accountHash, passwordHash, excludeUserID}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE user_account_hash = $1 AND password_hash = $2 AND role IN ('Admin', 'Manager')`
		args = []interface{}{accountHash, passwordHash}
	}
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("Security Validation Failed: Please use a more complex password to continue.")
	}
	return nil
}
