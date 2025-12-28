package domain

// BuildingUpdate 楼栋更新模型（对应 buildings 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（building_id, tenant_id）和必填字段（building_name）不在更新模型中
type BuildingUpdate struct {
	// 院区关联
	BranchID *UpdateString // UUID, nullable, FK → branches.branch_id

	// 楼栋名称
	BuildingName *UpdateString // VARCHAR(50), NOT NULL（但更新时可以为空，表示不更新）
}

