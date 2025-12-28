package domain

// UserBranch 用户-院区关联领域模型（对应 user_branches 表）
// 用途：
//  1. 支持一个用户属于多个院区的多对多关系
//  2. 用于权限过滤（branch_only）：通过 user_branches 表查询用户所属的院区
//  3. 用于告警接收范围过滤（alarm_scope = 'BRANCH'）
//  4. 支持主院区标识（is_primary），用于默认显示和业务逻辑
type UserBranch struct {
	// 主键
	UserBranchID string `db:"user_branch_id"` // UUID, PRIMARY KEY

	// 租户和关联
	TenantID string `db:"tenant_id"` // UUID, NOT NULL
	UserID   string `db:"user_id"`   // UUID, NOT NULL
	BranchID string `db:"branch_id"` // UUID, NOT NULL

	// 是否为主院区（用于默认显示和业务逻辑）
	// 注意：
	//   - 一个用户只能有一个主院区（is_primary = TRUE）
	//   - 如果用户只有一个院区关联，该院区自动为主院区
	//   - 如果用户有多个院区关联，必须明确指定一个为主院区
	//   - 用于权限过滤时的默认院区判断
	IsPrimary bool `db:"is_primary"` // BOOLEAN, NOT NULL, DEFAULT FALSE
}
