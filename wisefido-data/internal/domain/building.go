package domain

import (
	"database/sql"
)

// Building 楼栋领域模型（对应 buildings 表）
// 基于实际DB表结构：5个字段（floors已删除）
type Building struct {
	BuildingID   string         `db:"building_id"`
	TenantID     string         `db:"tenant_id"`
	BranchTag    sql.NullString `db:"branch_tag"`    // nullable
	BuildingName string         `db:"building_name"` // NOT NULL, default '-'
	CreatedAt    sql.NullTime   `db:"created_at"`     // nullable
	UpdatedAt    sql.NullTime   `db:"updated_at"`     // nullable
}

