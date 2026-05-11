package domain

import (
	"database/sql"
)

// Unit 单元领域模型（对应 units 表）
// 基于实际DB表结构：
//   - branch_id: UUID, NOT NULL, FK → branches.branch_id（业务规则：一家机构必然有一个分支或总部）
//   - building_id: UUID, nullable, FK → buildings.building_id（支持 home care 等无 building 的场景）
//   - 通过 JOIN branches 和 buildings 表获取 branch_name 和 building_name
type Unit struct {
	UnitID       string         `db:"unit_id"`
	TenantID     string         `db:"tenant_id"`
	BranchID     sql.NullString `db:"branch_id"`      // UUID, NOT NULL, FK → branches.branch_id（业务规则要求）
	BranchName   sql.NullString `db:"branch_name"`    // nullable, 通过 JOIN branches 表获取（不存储在 units 表）
	UnitName     string         `db:"unit_name"`      // NOT NULL
	BuildingID   sql.NullString `db:"building_id"`    // UUID, nullable, FK → buildings.building_id（支持 home care 场景）
	BuildingName sql.NullString `db:"building_name"`  // nullable, 通过 JOIN buildings 表获取（不存储在 units 表）
	Floor        sql.NullString `db:"floor"`          // nullable, default '1F' (由 Service 层控制)
	LayoutConfig sql.NullString `db:"layout_config"`  // nullable, JSONB
	// v2 双维度（2026-05-09 重设计）：
	//   UnitProperty 0=Home, 1=Facility (default)
	//   UnitType     0=unknown, 1=single (Private), 2=share (default), 3=public
	//   约束：Home → UnitType=0；Facility → UnitType ∈ {1,2,3}
	UnitProperty int8   `db:"unit_property"` // 0=Home, 1=Facility
	UnitType     int8   `db:"unit_type"`     // 0/1/2/3 enum
	Timezone     string `db:"timezone"`      // NOT NULL
}
