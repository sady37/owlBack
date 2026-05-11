package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresAuthRepository 认证Repository实现
type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresAuthRepository 创建认证Repository
func NewPostgresAuthRepository(db *sql.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

// 确保实现了接口
var _ AuthRepository = (*PostgresAuthRepository)(nil)

// GetUserForLogin 根据 tenant_id, account_hash, password_hash 查询用户（用于登录）
func (r *PostgresAuthRepository) GetUserForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*UserLoginInfo, error) {
	if tenantID == "" || len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("tenant_id, account_hash, and password_hash are required")
	}

	query := `
		SELECT u.user_id::text,
		       u.user_account,
		       COALESCE(u.nickname,''),
		       u.role,
		       COALESCE(u.status,'active'),
		       COALESCE(t.tenant_name,''),
		       COALESCE(t.domain,''),
		       CASE
		         WHEN u.email_hash = $2 THEN 'email'
		         WHEN u.phone_hash = $2 THEN 'phone'
		         WHEN u.user_account_hash = $2 THEN 'account'
		         ELSE 'account'
		       END as account_type
		  FROM users u
		  JOIN tenants t ON t.tenant_id = u.tenant_id
		 WHERE u.tenant_id = $1
		   AND u.password_hash = $3
		   AND (u.email_hash = $2 OR u.phone_hash = $2 OR u.user_account_hash = $2)
		 ORDER BY 
		   CASE
		     WHEN u.phone_hash = $2 THEN 1
		     WHEN u.email_hash = $2 THEN 2
		     WHEN u.user_account_hash = $2 THEN 3
		     ELSE 4
		   END ASC
		 LIMIT 1
	`

	var info UserLoginInfo
	err := r.db.QueryRowContext(ctx, query, tenantID, accountHash, passwordHash).Scan(
		&info.UserID,
		&info.UserAccount,
		&info.Nickname,
		&info.Role,
		&info.Status,
		&info.TenantName,
		&info.Domain,
		&info.AccountType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user for login: %w", err)
	}

	info.TenantID = tenantID
	return &info, nil
}

// SearchTenantsForUserLogin 根据 account_hash, password_hash 搜索匹配的机构
func (r *PostgresAuthRepository) SearchTenantsForUserLogin(ctx context.Context, accountHash, passwordHash []byte) ([]TenantLoginMatch, error) {
	if len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("account_hash and password_hash are required")
	}

	query := `
		SELECT DISTINCT u.tenant_id::text,
		       CASE
		         WHEN u.phone_hash = $1 THEN 'phone'
		         WHEN u.email_hash = $1 THEN 'email'
		         WHEN u.user_account_hash = $1 THEN 'account'
		         ELSE 'account'
		       END as account_type,
		       CASE
		         WHEN u.phone_hash = $1 THEN 1
		         WHEN u.email_hash = $1 THEN 2
		         WHEN u.user_account_hash = $1 THEN 3
		         ELSE 4
		       END as priority
		  FROM users u
		 WHERE u.password_hash = $2
		   AND COALESCE(u.status,'active') = 'active'
		   AND (u.email_hash = $1 OR u.phone_hash = $1 OR u.user_account_hash = $1)
		 ORDER BY priority ASC, u.tenant_id::text ASC
	`

	rows, err := r.db.QueryContext(ctx, query, accountHash, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to search tenants for user login: %w", err)
	}
	defer rows.Close()

	var matches []TenantLoginMatch
	for rows.Next() {
		var match TenantLoginMatch
		var priority int
		if err := rows.Scan(&match.TenantID, &match.AccountType, &priority); err != nil {
			continue
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// UpdateUserLastLogin 更新用户的 last_login_at
func (r *PostgresAuthRepository) UpdateUserLastLogin(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET last_login_at = NOW() WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update last_login_at: %w", err)
	}

	return nil
}

// GetResidentForLogin 根据 tenant_id, account_hash, password_hash 查询住户（用于登录）
func (r *PostgresAuthRepository) GetResidentForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*ResidentLoginInfo, error) {
	if tenantID == "" || len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("tenant_id, account_hash, and password_hash are required")
	}

	query := `
		SELECT r.resident_id::text,
		       r.resident_account,
		       COALESCE(r.nickname,''),
		       r.role,
		       COALESCE(r.status,'active'),
		       COALESCE(t.tenant_name,''),
		       COALESCE(t.domain,''),
		       CASE
		         WHEN r.email_hash = $2 THEN 'email'
		         WHEN r.phone_hash = $2 THEN 'phone'
		         WHEN r.resident_account_hash = $2 THEN 'account'
		         ELSE 'account'
		       END as account_type
		  FROM residents r
		  JOIN tenants t ON t.tenant_id = r.tenant_id
		  LEFT JOIN units u ON u.unit_id = r.unit_id
		 WHERE r.tenant_id = $1
		   AND r.password_hash = $3
		   AND (r.email_hash = $2 OR r.phone_hash = $2 OR r.resident_account_hash = $2)
		   AND COALESCE(r.status,'active') = 'active'
		   AND COALESCE(r.family_access,true) = true
		 ORDER BY 
		   CASE
		     WHEN r.phone_hash = $2 THEN 1
		     WHEN r.email_hash = $2 THEN 2
		     WHEN r.resident_account_hash = $2 THEN 3
		     ELSE 4
		   END ASC
		 LIMIT 1
	`

	var info ResidentLoginInfo
	err := r.db.QueryRowContext(ctx, query, tenantID, accountHash, passwordHash).Scan(
		&info.ResidentID,
		&info.ResidentAccount,
		&info.Nickname,
		&info.Role,
		&info.Status,
		&info.TenantName,
		&info.Domain,
		&info.AccountType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found")
		}
		return nil, fmt.Errorf("failed to get resident for login: %w", err)
	}

	info.TenantID = tenantID
	return &info, nil
}

// SearchTenantsForResidentLogin 根据 account_hash, password_hash 搜索匹配的机构（只查询 residents 表）
// Note: Emergency contacts (resident_contacts) cannot login, they only receive notifications
func (r *PostgresAuthRepository) SearchTenantsForResidentLogin(ctx context.Context, accountHash, passwordHash []byte) ([]TenantLoginMatch, error) {
	if len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("account_hash and password_hash are required")
	}

	// Query residents table only (resident_contacts cannot login)
	query := `
		SELECT DISTINCT r.tenant_id::text,
		       CASE
		         WHEN r.phone_hash = $1 THEN 'phone'
		         WHEN r.email_hash = $1 THEN 'email'
		         WHEN r.resident_account_hash = $1 THEN 'account'
		         ELSE 'account'
		       END as account_type,
		       CASE
		         WHEN r.phone_hash = $1 THEN 1
		         WHEN r.email_hash = $1 THEN 2
		         WHEN r.resident_account_hash = $1 THEN 3
		         ELSE 4
		       END as priority
		  FROM residents r
		 WHERE r.password_hash = $2
		   AND COALESCE(r.status,'active') = 'active'
		   AND (r.email_hash = $1 OR r.phone_hash = $1 OR r.resident_account_hash = $1)
		   AND COALESCE(r.family_access,true) = true
		 ORDER BY priority ASC, r.tenant_id::text ASC
	`

	rows, err := r.db.QueryContext(ctx, query, accountHash, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to search tenants from residents: %w", err)
	}
	defer rows.Close()

	var matches []TenantLoginMatch
	for rows.Next() {
		var match TenantLoginMatch
		var priority int
		if err := rows.Scan(&match.TenantID, &match.AccountType, &priority); err != nil {
			continue
		}
		matches = append(matches, match)
	}

	return matches, nil
}
