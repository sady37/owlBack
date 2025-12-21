package domain

import (
	"database/sql"
	"encoding/json"
)

// Device 设备领域模型（对应 devices 表）
// 基于实际DB表结构：devices表的所有字段
type Device struct {
	// 主键和租户
	DeviceID   string `db:"device_id"`
	TenantID   string `db:"tenant_id"` // NOT NULL

	// 关联 device_store
	DeviceStoreID sql.NullString `db:"device_store_id"` // nullable

	// 标识/资产
	DeviceName   string         `db:"device_name"`   // NOT NULL
	SerialNumber sql.NullString `db:"serial_number"` // nullable
	UID          sql.NullString `db:"uid"`           // nullable

	// 位置绑定（互斥）
	BoundRoomID sql.NullString `db:"bound_room_id"` // nullable
	BoundBedID  sql.NullString `db:"bound_bed_id"`  // nullable

	// 状态/维护
	Status            string `db:"status"`              // NOT NULL, default 'offline'
	BusinessAccess    string `db:"business_access"`      // NOT NULL, default 'pending'
	MonitoringEnabled bool   `db:"monitoring_enabled"`  // NOT NULL, default false

	// 元数据
	Metadata sql.NullString `db:"metadata"` // nullable, JSONB
}

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_id":          d.DeviceID,
		"tenant_id":          d.TenantID,
		"device_name":        d.DeviceName,
		"status":             d.Status,
		"business_access":    d.BusinessAccess,
		"monitoring_enabled": d.MonitoringEnabled,
	}
	if d.DeviceStoreID.Valid {
		m["device_store_id"] = d.DeviceStoreID.String
	}
	if d.SerialNumber.Valid {
		m["serial_number"] = d.SerialNumber.String
	}
	if d.UID.Valid {
		m["uid"] = d.UID.String
	}
	if d.BoundRoomID.Valid {
		m["bound_room_id"] = d.BoundRoomID.String
	} else {
		m["bound_room_id"] = nil
	}
	if d.BoundBedID.Valid {
		m["bound_bed_id"] = d.BoundBedID.String
	} else {
		m["bound_bed_id"] = nil
	}
	if d.Metadata.Valid {
		// 尝试解析JSON，如果失败则返回字符串
		var jsonData any
		if err := json.Unmarshal([]byte(d.Metadata.String), &jsonData); err == nil {
			m["metadata"] = jsonData
		} else {
			m["metadata"] = d.Metadata.String
		}
	}
	return m
}

