package domain

// UserBranch 用户-院区关联领域模型（对应 user_branches 表）
// 用途：
//  1. 支持一个用户属于多个院区的多对多关系
//  2. 用于权限过滤（branch_only）：通过 user_branches 表查询用户所属的院区
//  3. 用于告警接收范围过滤（relegation = 'BRANCH'）
type UserBranch struct {
	// 主键
	UserBranchID string `db:"user_branch_id"` // UUID, PRIMARY KEY

	// 租户和关联
	TenantID string `db:"tenant_id"` // UUID, NOT NULL
	UserID   string `db:"user_id"`   // UUID, NOT NULL
	BranchID string `db:"branch_id"` // UUID, NOT NULL
}
