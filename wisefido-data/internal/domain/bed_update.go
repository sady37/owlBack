package domain

// BedUpdate 床位更新模型（对应 beds 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（bed_id, tenant_id, room_id）和必填字段（bed_name）不在更新模型中
type BedUpdate struct {
	// 床位名称
	BedName *UpdateString // VARCHAR(50), NOT NULL（但更新时可以为空，表示不更新）

	// 床垫材质
	MattressMaterial *UpdateString // VARCHAR(50), nullable

	// 床垫厚度
	MattressThickness *UpdateString // VARCHAR(20), nullable
}

