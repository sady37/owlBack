package domain

import (
	"database/sql"
)

// Unit 单元领域模型（对应 units 表）
// 基于实际DB表结构：15个字段
type Unit struct {
	UnitID            string         `db:"unit_id"`
	TenantID          string         `db:"tenant_id"`
	BranchName        sql.NullString `db:"branch_name"`        // nullable
	UnitName          string         `db:"unit_name"`           // NOT NULL
	Building          sql.NullString `db:"building"`            // nullable (如果为 NULL，保存为 NULL)
	Floor             sql.NullString `db:"floor"`               // nullable, default '1F' (由 Service 层控制)
	AreaName          sql.NullString `db:"area_name"`           // nullable
	UnitNumber        string         `db:"unit_number"`         // NOT NULL
	LayoutConfig      sql.NullString `db:"layout_config"`       // nullable, JSONB
	UnitType          string         `db:"unit_type"`           // NOT NULL
	IsPublicSpace     bool           `db:"is_public_space"`     // NOT NULL, default false
	IsMultiPersonRoom bool           `db:"is_multi_person_room"` // NOT NULL, default false
	Timezone          string         `db:"timezone"`            // NOT NULL
	GroupList         sql.NullString `db:"groupList"`          // nullable, JSONB
	UserList          sql.NullString `db:"userList"`           // nullable, JSONB
}

