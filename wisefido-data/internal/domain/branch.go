package domain

import (
	"database/sql"
)

// Branch 院区领域模型（对应 branches 表）
// 业务规则：
//  1. Branch 是独立的实体，可以独立创建、编辑、删除
//  2. Branch 不依赖 buildings/units/users 的存在
//  3. 用于统一管理院区信息，避免 branch_name 在多个表中重复
//  4. branch_name 可以为 "-"（表示没有特定院区，使用默认值）
type Branch struct {
	// 主键
	BranchID string `db:"branch_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL

	// 院区名称：例如 "A 院区主楼"、"Spring 区域组SP"、"Denver-East  DVE"
	// 注意：
	//   - 必填字段，不能为 NULL
	//   - 可以为 "-"（表示没有特定院区，使用默认值）
	//   - 同一租户内院区名称唯一
	//   - 用于分组、路由、权限过滤等
	BranchName string `db:"branch_name"` // VARCHAR(255), NOT NULL

	// 院区描述（可选）
	Description sql.NullString `db:"description"` // TEXT, nullable

	// 创建和更新时间
	CreatedAt sql.NullTime `db:"created_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
	UpdatedAt sql.NullTime `db:"updated_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
}

