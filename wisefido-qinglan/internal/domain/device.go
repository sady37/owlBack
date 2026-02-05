package domain

import (
	"database/sql"
	"encoding/json"
	"sync"
)

// Device 设备领域模型（对应 devices 表）
// 基于 owlRD/db/17_devices.sql 的实际表结构
type Device struct {
	// 主键 / SaaS
	DeviceID  string `db:"device_id"`  // UUID, PRIMARY KEY REFERENCES device_store(device_id)
	DeviceUID string `db:"device_uid"` // VARCHAR(50), 用于首次连接匹配和查询
	TenantID  string `db:"tenant_id"`  // UUID, NOT NULL REFERENCES tenants(tenant_id)

	// 标识 / 资产
	DeviceName string `db:"device_name"` // VARCHAR(100), NOT NULL

	// 位置绑定（互斥）
	BoundRoomID sql.NullString `db:"bound_room_id"` // UUID REFERENCES rooms(room_id)
	BoundBedID  sql.NullString `db:"bound_bed_id"`  // UUID REFERENCES beds(bed_id)

	// 状态 / 维护
	Status            string `db:"status"`             // VARCHAR(20), NOT NULL DEFAULT 'offline'
	BusinessAccess    string `db:"business_access"`    // VARCHAR(20), NOT NULL DEFAULT 'pending'
	MonitoringEnabled bool   `db:"monitoring_enabled"` // BOOLEAN, NOT NULL DEFAULT FALSE

	// 其他扩展配置 / 标签
	Metadata sql.NullString `db:"metadata"` // JSONB

	// 物理属性（通过 JOIN device_store 获取）
	DeviceType      sql.NullString `db:"device_type"`      // from device_store
	DeviceModel     sql.NullString `db:"device_model"`     // from device_store
	IMEI            sql.NullString `db:"imei"`             // from device_store
	CommMode        sql.NullString `db:"comm_mode"`        // from device_store
	MCUModel        sql.NullString `db:"mcu_model"`        // from device_store
	FirmwareVersion sql.NullString `db:"firmware_version"` // from device_store
}

// DeviceLocationInfo 设备位置信息领域模型
// 包含完整的地址层级信息和设备规格信息
type DeviceLocationInfo struct {
	// 设备基本信息
	DeviceID   string `db:"device_id"`
	DeviceUID  string `db:"device_uid"`
	TenantID   string `db:"tenant_id"`
	DeviceName string `db:"device_name"`

	// 设备状态
	Status            string `db:"status"`
	BusinessAccess    string `db:"business_access"`
	MonitoringEnabled bool   `db:"monitoring_enabled"`

	// 绑定信息
	BoundRoomID sql.NullString `db:"bound_room_id"`
	BoundBedID  sql.NullString `db:"bound_bed_id"`

	// 位置信息（通过 JOIN 查询得到）
	BranchID     sql.NullString `db:"branch_id"`
	BranchName   sql.NullString `db:"branch_name"`
	BuildingID   sql.NullString `db:"building_id"`
	BuildingName sql.NullString `db:"building_name"`
	UnitID       sql.NullString `db:"unit_id"`
	UnitName     sql.NullString `db:"unit_name"`
	RoomID       sql.NullString `db:"room_id"`
	RoomName     sql.NullString `db:"room_name"`
	BedID        sql.NullString `db:"bed_id"`
	BedName      sql.NullString `db:"bed_name"`

	// 设备规格信息（来自 device_store）
	DeviceType      sql.NullString `db:"device_type"`
	DeviceModel     sql.NullString `db:"device_model"`
	IMEI            sql.NullString `db:"imei"`
	MAC             sql.NullString `db:"mac"`
	CommMode        sql.NullString `db:"comm_mode"`
	MCUModel        sql.NullString `db:"mcu_model"`
	FirmwareVersion sql.NullString `db:"firmware_version"`
}

// DeviceStatusItem 设备在线状态项（用于批量查询 API）
// device_id 与 device_uid 均保留，前端可能持其一查询，结构保持一致
type DeviceStatusItem struct {
	DeviceUID string `json:"device_uid"`
	DeviceID  string `json:"device_id"`
	TenantID  string `json:"tenant_id"`
	Status    string `json:"status"` // "online", "offline", "unsubscribed"
}

// DeviceWithLocation 带位置信息的设备结构
// 用于 Auth 缓存和消息处理
type DeviceWithLocation struct {
	Device       *Device
	LocationInfo *DeviceLocationInfo
}

var (
	// DeviceCache 设备缓存（key: uid, value: *DeviceWithLocation）
	// 在 Auth 时填充，供 MQTT Consumer 使用，避免重复查询数据库
	DeviceCache = &sync.Map{}
)

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_id":          d.DeviceID,
		"device_uid":         d.DeviceUID,
		"tenant_id":          d.TenantID,
		"device_name":        d.DeviceName,
		"status":             d.Status,
		"business_access":    d.BusinessAccess,
		"monitoring_enabled": d.MonitoringEnabled,
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
		var jsonData any
		if err := json.Unmarshal([]byte(d.Metadata.String), &jsonData); err == nil {
			m["metadata"] = jsonData
		} else {
			m["metadata"] = d.Metadata.String
		}
	}

	// 物理属性
	if d.DeviceType.Valid {
		m["device_type"] = d.DeviceType.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
	}
	if d.IMEI.Valid {
		m["imei"] = d.IMEI.String
	}
	if d.CommMode.Valid {
		m["comm_mode"] = d.CommMode.String
	}
	if d.MCUModel.Valid {
		m["mcu_model"] = d.MCUModel.String
	}
	if d.FirmwareVersion.Valid {
		m["firmware_version"] = d.FirmwareVersion.String
	}

	return m
}
