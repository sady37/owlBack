package domain

// ResidentUpdate 住户更新模型（对应 residents 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（resident_id, tenant_id）和必填字段（resident_account, resident_account_hash, password_hash, nickname, admission_date, status, role）不在更新模型中
type ResidentUpdate struct {
	// 住户账号（机构内部唯一标识，不包含姓名）
	ResidentAccount     *UpdateString // VARCHAR(100), NOT NULL（但更新时可以为空，表示不更新）
	ResidentAccountHash *UpdateBytes  // BYTEA, NOT NULL（但更新时可以为空，表示不更新）

	// 昵称（用于匿名化展示和查询）
	Nickname *UpdateString // VARCHAR(100), NOT NULL（但更新时可以为空，表示不更新）

	// 日期
	AdmissionDate *UpdateTime // DATE, NOT NULL（但更新时可以为空，表示不更新）
	DischargeDate *UpdateTime // DATE, nullable

	// 护理级别
	ServiceLevel *UpdateString // VARCHAR(20), nullable

	// 状态
	Status *UpdateString // VARCHAR(50), NOT NULL（但更新时可以为空，表示不更新）

	// 角色
	Role *UpdateString // VARCHAR(50), NOT NULL（但更新时可以为空，表示不更新）

	// 扩展信息
	Metadata *UpdateJSON   // JSONB, nullable
	Note     *UpdateString // TEXT, nullable

	// 联系方式（可选 PHI）
	Phone        *UpdateString // VARCHAR(50), nullable
	Email        *UpdateString // VARCHAR(255), nullable
	PhoneHash    *UpdateBytes  // BYTEA, nullable
	EmailHash    *UpdateBytes  // BYTEA, nullable
	PasswordHash *UpdateBytes  // BYTEA, NOT NULL（但更新时可以为空，表示不更新）

	// 权限控制
	IsAccessEnabled *UpdateBool // BOOLEAN, NOT NULL（但更新时可以为空，表示不更新）

	// 院区关联
	BranchID *UpdateString // UUID, nullable

	// 位置绑定关系
	UnitID *UpdateString // UUID, nullable
	RoomID *UpdateString // UUID, nullable
	BedID  *UpdateString // UUID, nullable
}
