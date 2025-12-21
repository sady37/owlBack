package domain

import (
	"database/sql"
	"github.com/lib/pq"
)

// User 用户领域模型（对应 users 表）
// 基于实际DB表结构：20个字段
type User struct {
	// 主键和租户
	UserID   string `db:"user_id"`
	TenantID string `db:"tenant_id"`

	// 账号信息
	UserAccount      string `db:"user_account"`      // NOT NULL
	UserAccountHash  []byte `db:"user_account_hash"` // NOT NULL
	PasswordHash     []byte `db:"password_hash"`     // nullable
	PinHash          []byte `db:"pin_hash"`          // nullable

	// 基本信息
	Nickname sql.NullString `db:"nickname"` // nullable
	Email    sql.NullString `db:"email"`    // nullable
	Phone    sql.NullString `db:"phone"`    // nullable
	Role     string         `db:"role"`     // NOT NULL
	Status   string         `db:"status"`   // nullable, default 'active'

	// 联系方式哈希
	EmailHash []byte `db:"email_hash"` // nullable
	PhoneHash []byte `db:"phone_hash"` // nullable

	// 告警设置
	AlarmLevels  pq.StringArray `db:"alarm_levels"`  // nullable, VARCHAR[]
	AlarmChannels pq.StringArray `db:"alarm_channels"` // nullable, VARCHAR[]
	AlarmScope   sql.NullString `db:"alarm_scope"`   // nullable

	// 登录和标签
	LastLoginAt sql.NullTime   `db:"last_login_at"` // nullable
	Tags        sql.NullString `db:"tags"`        // nullable, JSONB数组
	BranchTag   sql.NullString `db:"branch_tag"`  // nullable

	// 偏好设置
	Preferences sql.NullString `db:"preferences"` // nullable, JSONB, default '{}'::jsonb
}

