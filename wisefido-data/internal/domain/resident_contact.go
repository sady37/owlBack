package domain

import (
	"database/sql"
	"encoding/json"
)

// ResidentContact 住户联系人领域模型（对应 resident_contacts 表）
// 注意：联系人不能登录系统，仅作为住户的属性用于接收告警通知，不是角色
// 主键：PRIMARY KEY (resident_id, slot)
// 注意：contact_id字段已从数据库删除，但前端DTO需要返回标识，这里使用空字符串表示
type ResidentContact struct {
	// 租户和住户
	TenantID   string `db:"tenant_id"`   // UUID, NOT NULL
	ResidentID string `db:"resident_id"` // UUID, NOT NULL

	// 槽位（主键的一部分）
	Slot string `db:"slot"` // VARCHAR(1), NOT NULL（'A','B','C','D','E'），PRIMARY KEY (resident_id, slot), UNIQUE(tenant_id, resident_id, slot)
	
	// ContactID 用于前端标识（数据库表已删除contact_id字段，这里为空字符串或使用resident_id+slot组合）
	ContactID string `db:"-"` // 不映射到数据库，前端需要时返回空字符串或组合值

	// 关系
	Relationship sql.NullString `db:"relationship"` // VARCHAR(50), nullable（Child/Spouse/Friend/Caregiver）

	// 启用控制
	IsEnabled        bool            `db:"is_enabled"`        // BOOLEAN, NOT NULL, DEFAULT FALSE（是否启用该联系人，对应前端的 "slot Enable" 复选框）
	AlertTimeWindow  json.RawMessage `db:"alert_time_window"` // JSONB, nullable（告警接收时间窗口）

	// 可选的PHI（姓名/联系方式）
	ContactFirstName sql.NullString `db:"contact_first_name"` // VARCHAR(100), nullable
	ContactLastName  sql.NullString `db:"contact_last_name"`  // VARCHAR(100), nullable
	ContactPhone     sql.NullString `db:"contact_phone"`      // VARCHAR(25), nullable
	ContactEmail     sql.NullString `db:"contact_email"`      // VARCHAR(255), nullable
	ContactPhoneHash []byte         `db:"contact_phone_hash"` // VARCHAR(64), nullable (hash for search)
	ContactEmailHash []byte         `db:"contact_email_hash"` // VARCHAR(64), nullable (hash for search)
	ReceiveSMS       bool           `db:"receive_sms"`        // BOOLEAN, DEFAULT FALSE
	ReceiveEmail     bool           `db:"receive_email"`      // BOOLEAN, DEFAULT FALSE
}
