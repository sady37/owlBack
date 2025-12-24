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
		       COALESCE(u.branch_tag, '') as branch_tag,
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
		     WHEN u.email_hash = $2 THEN 1
		     WHEN u.phone_hash = $2 THEN 2
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
		&info.BranchTag,
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
		         WHEN u.email_hash = $1 THEN 'email'
		         WHEN u.phone_hash = $1 THEN 'phone'
		         WHEN u.user_account_hash = $1 THEN 'account'
		         ELSE 'account'
		       END as account_type,
		       CASE
		         WHEN u.email_hash = $1 THEN 1
		         WHEN u.phone_hash = $1 THEN 2
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
		       COALESCE(u.branch_name, '') as branch_tag,
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
		   AND COALESCE(r.can_view_status,true) = true
		 ORDER BY 
		   CASE
		     WHEN r.email_hash = $2 THEN 1
		     WHEN r.phone_hash = $2 THEN 2
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
		&info.BranchTag,
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

// GetResidentContactForLogin 根据 tenant_id, account_hash, password_hash 查询联系人（用于登录）
func (r *PostgresAuthRepository) GetResidentContactForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*ResidentContactLoginInfo, error) {
	if tenantID == "" || len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("tenant_id, account_hash, and password_hash are required")
	}

	query := `
		SELECT rc.contact_id::text,
		       rc.resident_id::text,
		       rc.slot,
		       COALESCE(rc.contact_first_name,''),
		       COALESCE(rc.contact_last_name,''),
		       rc.role,
		       COALESCE(rc.is_enabled,true),
		       COALESCE(t.tenant_name,''),
		       COALESCE(t.domain,''),
		       COALESCE(u.branch_name, '') as branch_tag,
		       CASE
		         WHEN rc.email_hash = $2 THEN 'email'
		         WHEN rc.phone_hash = $2 THEN 'phone'
		         ELSE 'phone'
		       END as account_type
		  FROM resident_contacts rc
		  JOIN tenants t ON t.tenant_id = rc.tenant_id
		  JOIN residents r ON r.resident_id = rc.resident_id AND r.tenant_id = rc.tenant_id
		  LEFT JOIN units u ON u.unit_id = r.unit_id
		 WHERE rc.tenant_id = $1
		   AND rc.password_hash = $3
		   AND (rc.email_hash = $2 OR rc.phone_hash = $2)
		   AND COALESCE(rc.is_enabled,true) = true
		   -- 注意：一个 contact 的 email_hash/phone_hash 可能被多个 resident_contacts 记录共享（同一联系人关联多个住户）
		   -- 优先选择关联的 resident 的 can_view_status=true 的 contact
		   AND EXISTS (
		     SELECT 1 FROM residents r2
		     WHERE r2.resident_id = rc.resident_id
		       AND r2.tenant_id = rc.tenant_id
		       AND COALESCE(r2.can_view_status,true) = true
		   )
		 ORDER BY 
		   -- 优先选择 can_view_status=true 的记录
		   CASE WHEN COALESCE(r.can_view_status,true) = true THEN 0 ELSE 1 END ASC,
		   CASE
		     WHEN rc.email_hash = $2 THEN 1
		     WHEN rc.phone_hash = $2 THEN 2
		     ELSE 3
		   END ASC
		 LIMIT 1
	`

	var info ResidentContactLoginInfo
	err := r.db.QueryRowContext(ctx, query, tenantID, accountHash, passwordHash).Scan(
		&info.ContactID,
		&info.ResidentID,
		&info.Slot,
		&info.ContactFirstName,
		&info.ContactLastName,
		&info.Role,
		&info.IsEnabled,
		&info.TenantName,
		&info.Domain,
		&info.BranchTag,
		&info.AccountType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident contact not found")
		}
		return nil, fmt.Errorf("failed to get resident contact for login: %w", err)
	}

	info.TenantID = tenantID
	return &info, nil
}

// SearchTenantsForResidentLogin 根据 account_hash, password_hash 搜索匹配的机构（包含 resident_contacts 和 residents 两步查询）
func (r *PostgresAuthRepository) SearchTenantsForResidentLogin(ctx context.Context, accountHash, passwordHash []byte) ([]TenantLoginMatch, error) {
	if len(accountHash) == 0 || len(passwordHash) == 0 {
		return nil, fmt.Errorf("account_hash and password_hash are required")
	}

	// Step 1: 查询 resident_contacts 表
	// 注意：一个 contacemail_hash/phone_hash 可能被多个 resident_contacts 记录共享（同一联系人关联多个住户）
	// 因此，需要检查该 email_hash/phone_hash 对应的所有关联住户，只要至少有一个住户的 can_view_status=true，就允许登录
	query1 := `
		SELECT DISTINCT rc.tenant_id::text,
		       CASE
		         WHEN rc.email_hash = $1 THEN 'email'
		         WHEN rc.phone_hash = $1 THEN 'phone'
		         ELSE 'phone'
		       END as account_type,
		       CASE
		         WHEN rc.email_hash = $1 THEN 1
		         WHEN rc.phone_hash = $1 THEN 2
		         ELSE 3
		       END as priority
		  FROM resident_contacts rc
		 WHERE rc.password_hash = $2
		   AND COALESCE(rc.is_enabled,true) = true
		   AND (rc.email_hash = $1 OR rc.phone_hash = $1)
		   AND EXISTS (
		     SELECT 1 FROM residents r
		     WHERE r.resident_id = rc.resident_id
		       AND r.tenant_id = rc.tenant_id
		       AND COALESCE(r.can_view_status,true) = true
		   )
		 ORDER BY priority ASC, rc.tenant_id::text ASC
	`

	rows1, err := r.db.QueryContext(ctx, query1, accountHash, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to search tenants from resident_contacts: %w", err)
	}
	defer rows1.Close()

	var matches []TenantLoginMatch
	var count int
	for rows1.Next() {
		var match TenantLoginMatch
		var priority int
		if err := rows1.Scan(&match.TenantID, &match.AccountType, &priority); err != nil {
			continue
		}
		matches = append(matches, match)
		count++
	}

	// Step 2: 如果 Step 1 无匹配，查询 residents 表
	if count == 0 {
		query2 := `
			SELECT DISTINCT r.tenant_id::text,
			       CASE
			         WHEN r.email_hash = $1 THEN 'email'
			         WHEN r.phone_hash = $1 THEN 'phone'
			         WHEN r.resident_account_hash = $1 THEN 'account'
			         ELSE 'account'
			       END as account_type,
			       CASE
			         WHEN r.email_hash = $1 THEN 1
			         WHEN r.phone_hash = $1 THEN 2
			         WHEN r.resident_account_hash = $1 THEN 3
			         ELSE 4
			       END as priority
			  FROM residents r
			 WHERE r.password_hash = $2
			   AND COALESCE(r.status,'active') = 'active'
			   AND (r.email_hash = $1 OR r.phone_hash = $1 OR r.resident_account_hash = $1)
			   AND COALESCE(r.can_view_status,true) = true
			 ORDER BY priority ASC, r.tenant_id::text ASC
		`

		rows2, err := r.db.QueryContext(ctx, query2, accountHash, passwordHash)
		if err != nil {
			return nil, fmt.Errorf("failed to search tenants from residents: %w", err)
		}
		defer rows2.Close()

		for rows2.Next() {
			var match TenantLoginMatch
			var priority int
			if err := rows2.Scan(&match.TenantID, &match.AccountType, &priority); err != nil {
				continue
			}
			matches = append(matches, match)
		}
	}

	return matches, nil
}
