package domain

import (
	"database/sql"
)

// Bed 床位领域模型（对应 beds 表）
// 基于实际DB表结构：6个字段（bound_device_count已删除）
type Bed struct {
	BedID            string         `db:"bed_id"`
	TenantID         string         `db:"tenant_id"`
	RoomID           string         `db:"room_id"`
	BedName          string         `db:"bed_name"`
	BedType          string         `db:"bed_type"`           // NOT NULL
	MattressMaterial sql.NullString `db:"mattress_material"`  // nullable
	MattressThickness sql.NullString `db:"mattress_thickness"` // nullable
}

