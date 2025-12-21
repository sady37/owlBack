package domain

import "encoding/json"

// ResidentCaregiver 住户护理人员关联领域模型（对应 resident_caregivers 表）
// 表示哪些护理人员/staff主要负责某位住户（护理分配关系）
type ResidentCaregiver struct {
	// 主键
	CaregiverID string `db:"caregiver_id"` // UUID, PRIMARY KEY

	// 租户和住户
	TenantID   string `db:"tenant_id"`   // UUID, NOT NULL
	ResidentID string `db:"resident_id"` // UUID, NOT NULL, UNIQUE(tenant_id, resident_id)

	// 警报通报组（resident级别，可选）
	// 路由优先级：
	//   1) 优先使用 resident_caregivers 表的配置（如果有）
	//   2) 如果没有或需要补充，使用 unit 级别的配置
	//   3) 两者取并集
	GroupList json.RawMessage `db:"group_list"` // JSONB, nullable（用户组，JSON格式，用于告警路由）
	UserList  json.RawMessage `db:"user_list"`  // JSONB, nullable（用户数组，JSON格式，直接指定用户列表，用于告警路由）

	// 配置来源标识（用于GetResidentCaregivers返回时区分unit级别和resident级别）
	// 注意：此字段不在数据库表中，仅用于返回结果
	Source string `db:"-"` // "unit" 或 "resident"
}

