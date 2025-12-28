package domain

// ResidentContactUpdate 住户联系人更新模型（对应 resident_contacts 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（tenant_id, resident_id, slot）不在更新模型中
// 注意：主键是 PRIMARY KEY (resident_id, slot)
type ResidentContactUpdate struct {
	// 关系
	Relationship *UpdateString // VARCHAR(50), nullable

	// 启用控制
	IsEnabled       *UpdateBool // BOOLEAN, NOT NULL, DEFAULT FALSE（是否启用该联系人）
	AlertTimeWindow *UpdateJSON // JSONB, nullable（告警接收时间窗口）

	// 可选的PHI（姓名/联系方式）
	ContactFirstName *UpdateString // VARCHAR(100), nullable
	ContactLastName  *UpdateString // VARCHAR(100), nullable
	ContactPhone     *UpdateString // VARCHAR(25), nullable
	ContactEmail     *UpdateString // VARCHAR(255), nullable
	ReceiveSMS       *UpdateBool   // BOOLEAN, DEFAULT FALSE
	ReceiveEmail     *UpdateBool   // BOOLEAN, DEFAULT FALSE
}

