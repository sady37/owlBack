package domain

// BranchUpdate 院区更新模型（对应 branches 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（branch_id, tenant_id）和必填字段（branch_name）不在更新模型中
type BranchUpdate struct {
	// 院区名称
	BranchName *UpdateString // VARCHAR(255), NOT NULL（但更新时可以为空，表示不更新）

	// 院区描述
	Description *UpdateString // TEXT, nullable

	// 时区（IANA tz name；空 = 继承 tenant.timezone）
	Timezone *UpdateString
}

