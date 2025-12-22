package domain

import (
	"database/sql"
)

// Bed 床位领域模型（对应 beds 表）
// 基于实际DB表结构：5个字段（bound_device_count 和 bed_type 已删除）
// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
type Bed struct {
	BedID            string         `db:"bed_id"`
	TenantID         string         `db:"tenant_id"`
	RoomID           string         `db:"room_id"`
	BedName          string         `db:"bed_name"`
	MattressMaterial sql.NullString `db:"mattress_material"`  // nullable
	MattressThickness sql.NullString `db:"mattress_thickness"` // nullable
}

