package domain

import (
	"database/sql"
)

// Room 房间领域模型（对应 rooms 表）
// 基于实际DB表结构：4个字段
type Room struct {
	RoomID       string         `db:"room_id"`
	TenantID     string         `db:"tenant_id"`
	UnitID       string         `db:"unit_id"`
	RoomName     string         `db:"room_name"`
	LayoutConfig sql.NullString `db:"layout_config"` // nullable, JSONB
}

