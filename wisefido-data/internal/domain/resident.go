package domain

import (
	"encoding/json"
	"time"
)

// Resident 住户领域模型（对应 residents 表）
// 完全匿名化，无PII存储
type Resident struct {
	// 主键
	ResidentID string `db:"resident_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL

	// 住户账号（机构内部唯一标识，不包含姓名）
	ResidentAccount     string `db:"resident_account"`      // VARCHAR(100), NOT NULL, UNIQUE(tenant_id, resident_account)
	ResidentAccountHash []byte `db:"resident_account_hash"` // BYTEA, NOT NULL

	// 昵称（用于匿名化展示和查询）
	Nickname string `db:"nickname"` // VARCHAR(100), NOT NULL, UNIQUE(tenant_id, nickname)

	// 日期
	AdmissionDate *time.Time `db:"admission_date"` // DATE, NOT NULL
	DischargeDate *time.Time `db:"discharge_date"` // DATE, nullable（仅在discharged/transferred状态时有值）

	// 护理级别
	ServiceLevel string `db:"service_level"` // VARCHAR(20), nullable（引用service_levels.level_code）

	// 状态
	Status string `db:"status"` // VARCHAR(50), NOT NULL, DEFAULT 'active' (active/discharged/transferred)

	// 角色
	Role string `db:"role"` // VARCHAR(50), NOT NULL, DEFAULT 'Resident'

	// 扩展信息
	Metadata json.RawMessage `db:"metadata"` // JSONB, nullable（仅包含非PII信息）
	Note     string          `db:"note"`    // TEXT, nullable（客户备注，非PII信息）

	// 登录/重置用的联系方式哈希
	PhoneHash    []byte `db:"phone_hash"`     // BYTEA, nullable
	EmailHash    []byte `db:"email_hash"`      // BYTEA, nullable
	PasswordHash []byte `db:"password_hash"`  // BYTEA, nullable

	// 家庭标签
	FamilyTag string `db:"family_tag"` // VARCHAR(100), nullable（家庭标识符）

	// 权限控制
	CanViewStatus bool `db:"can_view_status"` // BOOLEAN, NOT NULL, DEFAULT TRUE

	// 位置绑定关系
	UnitID string `db:"unit_id"` // UUID, NOT NULL（必须指定unit）
	RoomID string `db:"room_id"` // UUID, nullable
	BedID  string `db:"bed_id"`  // UUID, nullable（如果指定bed_id，必须同时指定room_id）
}

