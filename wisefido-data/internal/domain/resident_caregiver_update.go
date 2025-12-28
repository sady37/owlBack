package domain

// ResidentCaregiverUpdate 住户护理人员关联更新模型（对应 resident_caregivers 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（caregiver_id, tenant_id, resident_id）不在更新模型中
// 注意：此表使用 UPSERT 语义（UNIQUE(tenant_id, resident_id)）
type ResidentCaregiverUpdate struct {
	// 警报通报组（resident级别，可选）
	GroupList *UpdateJSON // JSONB, nullable（用户组，JSON格式，用于告警路由）
	UserList  *UpdateJSON // JSONB, nullable（用户数组，JSON格式，直接指定用户列表，用于告警路由）
}

