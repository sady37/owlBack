package domain

import (
	"database/sql"
)

// DefaultBranchName 租户默认院区名称；未指定院区时使用此 branch。创建 tenant 时自动创建。
const DefaultBranchName = "default"

// DefaultBranchDescription 默认院区的 description：pending/resident 临时归属，表示待分配。
const DefaultBranchDescription = "default branch. DefaultBranch: pending, resident, provisional designation – indicates to be allocated."

// Branch 院区领域模型（对应 branches 表）
// 业务规则：
//  1. Branch 是独立的实体，可以独立创建、编辑、删除
//  2. Branch 不依赖 buildings/units/users 的存在
//  3. 用于统一管理院区信息，避免 branch_name 在多个表中重复
//  4. branch_name 可以为 DefaultBranchName（表示默认院区，未指定时归入此处）
type Branch struct {
	// 主键
	BranchID string `db:"branch_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL

	// 院区名称：例如 "A 院区主楼"、"Spring 区域组SP"、或 DefaultBranchName 表示默认院区
	// 注意：
	//   - 必填字段，不能为 NULL
	//   - 可以为 DefaultBranchName（默认院区，未指定时归入此处）
	//   - 同一租户内院区名称唯一
	//   - 用于分组、路由、权限过滤等
	BranchName string `db:"branch_name"` // VARCHAR(255), NOT NULL

	// 院区描述（可选）
	Description sql.NullString `db:"description"` // TEXT, nullable

	// 创建和更新时间
	CreatedAt sql.NullTime `db:"created_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
	UpdatedAt sql.NullTime `db:"updated_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
}

