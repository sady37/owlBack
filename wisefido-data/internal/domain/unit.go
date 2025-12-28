package domain

import (
	"database/sql"
)

// Unit 单元领域模型（对应 units 表）
// 基于实际DB表结构：units 表有 branch_id 字段（UUID, nullable），通过 JOIN branches 表获取 branch_name
type Unit struct {
	UnitID       string         `db:"unit_id"`
	TenantID     string         `db:"tenant_id"`
	BranchID     sql.NullString `db:"branch_id"`      // UUID, nullable, FK → branches.branch_id
	BranchName   sql.NullString `db:"branch_name"`    // nullable, 通过 JOIN branches 表获取
	UnitName     string         `db:"unit_name"`      // NOT NULL
	BuildingName sql.NullString `db:"building_name"`  // nullable (如果为 NULL，保存为 NULL)
	Floor        sql.NullString `db:"floor"`          // nullable, default '1F' (由 Service 层控制)
	LayoutConfig sql.NullString `db:"layout_config"`  // nullable, JSONB
	UnitType     string         `db:"unit_type"`      // NOT NULL
	IsPublic     bool           `db:"is_public"`      // NOT NULL, default false
	IsSharedUnit bool           `db:"is_shared_unit"` // NOT NULL, default false
	Timezone     string         `db:"timezone"`       // NOT NULL
}
