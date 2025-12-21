package repository

import (
	"context"
)

// AuthRepository 认证相关的Repository接口
// 专门用于登录查询，支持优先级排序和完整信息返回
type AuthRepository interface {
	// ========== Staff 登录查询 ==========
	
	// GetUserForLogin 根据 tenant_id, account_hash, password_hash 查询用户（用于登录）
	// 支持优先级：email_hash > phone_hash > user_account_hash
	// 返回完整信息：包括 tenant_name, domain, branch_tag
	// 状态检查：status = 'active'
	GetUserForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*UserLoginInfo, error)
	
	// SearchTenantsForUserLogin 根据 account_hash, password_hash 搜索匹配的机构（用于 tenant_id 自动解析）
	// 返回匹配的 tenant_id 列表（按优先级排序）
	SearchTenantsForUserLogin(ctx context.Context, accountHash, passwordHash []byte) ([]TenantLoginMatch, error)
	
	// UpdateUserLastLogin 更新用户的 last_login_at
	UpdateUserLastLogin(ctx context.Context, userID string) error
	
	// ========== Resident 登录查询 ==========
	
	// GetResidentForLogin 根据 tenant_id, account_hash, password_hash 查询住户（用于登录）
	// 支持优先级：email_hash > phone_hash > resident_account_hash
	// 返回完整信息：包括 tenant_name, domain, branch_tag
	// 状态检查：status = 'active' AND can_view_status = true
	GetResidentForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*ResidentLoginInfo, error)
	
	// GetResidentContactForLogin 根据 tenant_id, account_hash, password_hash 查询联系人（用于登录）
	// 支持优先级：email_hash > phone_hash
	// 返回完整信息：包括 tenant_name, domain, branch_tag, resident_id, slot
	// 状态检查：is_enabled = true AND can_view_status = true
	GetResidentContactForLogin(ctx context.Context, tenantID string, accountHash, passwordHash []byte) (*ResidentContactLoginInfo, error)
	
	// SearchTenantsForResidentLogin 根据 account_hash, password_hash 搜索匹配的机构（用于 tenant_id 自动解析）
	// 包含两步查询：先查 resident_contacts，再查 residents
	// 返回匹配的 tenant_id 列表（按优先级排序）
	SearchTenantsForResidentLogin(ctx context.Context, accountHash, passwordHash []byte) ([]TenantLoginMatch, error)
}

// UserLoginInfo 用户登录信息（包含完整信息）
type UserLoginInfo struct {
	UserID      string
	UserAccount string
	Nickname    string
	Role        string
	Status      string
	TenantID    string
	TenantName  string
	Domain      string
	BranchTag   string
	AccountType string // "email" | "phone" | "account"
}

// ResidentLoginInfo 住户登录信息（包含完整信息）
type ResidentLoginInfo struct {
	ResidentID      string
	ResidentAccount string
	Nickname        string
	Role            string
	Status          string
	TenantID        string
	TenantName      string
	Domain          string
	BranchTag       string
	AccountType     string // "email" | "phone" | "account"
}

// ResidentContactLoginInfo 住户联系人登录信息（包含完整信息）
type ResidentContactLoginInfo struct {
	ContactID       string
	ResidentID      string
	Slot            string
	ContactFirstName string
	ContactLastName  string
	Role            string
	IsEnabled       bool
	TenantID        string
	TenantName      string
	Domain          string
	BranchTag       string
	AccountType     string // "email" | "phone"
}

// TenantLoginMatch 机构登录匹配信息
type TenantLoginMatch struct {
	TenantID    string
	AccountType string // "email" | "phone" | "account"
}

