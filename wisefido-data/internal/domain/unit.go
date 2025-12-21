package domain

import (
	"database/sql"
)

// Unit 单元领域模型（对应 units 表）
// 基于实际DB表结构：15个字段
type Unit struct {
	UnitID            string         `db:"unit_id"`
	TenantID          string         `db:"tenant_id"`
	BranchTag         sql.NullString `db:"branch_tag"`         // nullable
	UnitName          string         `db:"unit_name"`           // NOT NULL
	Building          string         `db:"building"`            // NOT NULL, default '-'
	Floor             string         `db:"floor"`               // NOT NULL, default '1F'
	AreaTag           sql.NullString `db:"area_tag"`            // nullable
	UnitNumber        string         `db:"unit_number"`         // NOT NULL
	LayoutConfig      sql.NullString `db:"layout_config"`       // nullable, JSONB
	UnitType          string         `db:"unit_type"`           // NOT NULL
	IsPublicSpace     bool           `db:"is_public_space"`     // NOT NULL, default false
	IsMultiPersonRoom bool           `db:"is_multi_person_room"` // NOT NULL, default false
	Timezone          string         `db:"timezone"`            // NOT NULL
	GroupList         sql.NullString `db:"groupList"`          // nullable, JSONB
	UserList          sql.NullString `db:"userList"`           // nullable, JSONB
}

