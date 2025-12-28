package domain

import (
	"database/sql"
)

// Building 楼栋领域模型（对应 buildings 表）
// 业务规则：
//  1. Building 是独立的实体，可以独立创建、编辑、删除
//  2. Building 不依赖 units 的存在，即使没有 units 也可以创建 building
//  3. building_name 可以为 "-"（表示没有特定楼栋，使用默认值）
type Building struct {
	// 主键
	BuildingID string `db:"building_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL

	// 院区关联：引用 branches.branch_id
	// 注意：可以为 NULL，表示没有特定院区（默认值）
	BranchID sql.NullString `db:"branch_id"` // UUID, nullable, FK → branches.branch_id

	// 院区名称：从 branches 表 JOIN 获取（用于显示，不存储在 buildings 表）
	BranchName sql.NullString `db:"-"` // 不映射到数据库字段，通过 JOIN 获取

	// 楼栋名称：例如 "Building A"、"主楼"、"A"
	// 注意：
	//   - 必填字段，不能为 NULL
	//   - 可以为 "-"（表示没有特定楼栋，使用默认值）
	BuildingName string `db:"building_name"` // VARCHAR(50), NOT NULL

	// 创建和更新时间
	CreatedAt sql.NullTime `db:"created_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
	UpdatedAt sql.NullTime `db:"updated_at"` // TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP
}
