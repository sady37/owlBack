package domain

import "encoding/json"

// ResidentContact 住户联系人领域模型（对应 resident_contacts 表）
// 住户紧急联系人 / 家属账号
type ResidentContact struct {
	// 主键
	ContactID string `db:"contact_id"` // UUID, PRIMARY KEY

	// 租户和住户
	TenantID   string `db:"tenant_id"`   // UUID, NOT NULL
	ResidentID string `db:"resident_id"` // UUID, NOT NULL

	// 槽位
	Slot string `db:"slot"` // VARCHAR(1), NOT NULL（'A','B','C','D','E'），UNIQUE(tenant_id, resident_id, slot)

	// 启用状态
	IsEnabled bool `db:"is_enabled"` // BOOLEAN, NOT NULL, DEFAULT TRUE

	// 关系
	Relationship string `db:"relationship"` // VARCHAR(50), nullable（Child/Spouse/Friend/Caregiver）

	// 角色
	Role string `db:"role"` // VARCHAR(20), NOT NULL, DEFAULT 'Family'

	// 告警接收控制
	IsEmergencyContact bool          `db:"is_emergency_contact"` // BOOLEAN, NOT NULL, DEFAULT FALSE
	AlertTimeWindow    json.RawMessage `db:"alert_time_window"`  // JSONB, nullable

	// 可选的PHI（姓名/联系方式）
	ContactFirstName string `db:"contact_first_name"` // VARCHAR(100), nullable
	ContactLastName  string `db:"contact_last_name"` // VARCHAR(100), nullable
	ContactPhone     string `db:"contact_phone"`     // VARCHAR(25), nullable
	ContactEmail     string `db:"contact_email"`     // VARCHAR(255), nullable
	ReceiveSMS        bool   `db:"receive_sms"`      // BOOLEAN, DEFAULT FALSE
	ReceiveEmail      bool   `db:"receive_email"`    // BOOLEAN, DEFAULT FALSE

	// 登录/重置用的联系方式哈希
	PhoneHash    []byte `db:"phone_hash"`     // BYTEA, nullable
	EmailHash    []byte `db:"email_hash"`     // BYTEA, nullable
	PasswordHash []byte `db:"password_hash"` // BYTEA, nullable
}

