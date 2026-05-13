package domain

import (
	"database/sql"
	"encoding/json"
	"sync"
)

// Device 设备领域模型（v2: 派生自 device_factory_meta + device_runtime_state + devices）
//
// v2 schema 映射（详见 owlRD/dbv2/23_devices.sql / 21_device_factory_meta.sql / 22_device_runtime_state.sql）：
//   - DeviceID/DeviceUID/DeviceType/DeviceModel/IMEI/CommMode/MCUModel: device_factory_meta
//   - DeviceIPv6/MonitoringEnabled/AllowAccess: devices (PK = device_ipv6)
//   - FirmwareVersion/Online/SensorDetached: device_runtime_state
//   - TenantID/BoundRoomID/BoundBedID/DeviceName: 由 IPv6 prefix-mask / device_code 派生（无独立列）
//   - BusinessAccess: legacy 字段映射 (access=TRUE → "approved" / FALSE → "pending")
type Device struct {
	// 主键 / 身份（dfm）
	DeviceID  string `db:"device_id"`  // UUID, dfm.device_id
	DeviceUID string `db:"device_uid"` // VARCHAR(50), dfm.device_uid (logMAC)

	// 空间地址（v2 first-class）
	DeviceIPv6 string `db:"device_ipv6"` // INET /128, devices.device_ipv6 (text repr)
	TenantID   string `db:"tenant_id"`   // 派生：host(network(set_masklen(device_ipv6, 48)))，"fd00:0:1::" 等

	// 显示用（dfm.device_code 优先，fallback device_uid）
	DeviceName string `db:"device_name"` // COALESCE(dfm.device_code, dfm.device_uid)

	// 位置绑定（v2 派生自 device_ipv6 字节段；保留 v1 字段名供 caller）
	BoundRoomID sql.NullString `db:"bound_room_id"` // /88 prefix host(...)::text；NULL if mask <88
	BoundBedID  sql.NullString `db:"bound_bed_id"`  // /96 prefix host(...)::text；NULL if mask <96

	// 运行时状态（drs + devices.access）
	Status            string `db:"status"`             // "online"|"offline"，drs.online 派生
	BusinessAccess    string `db:"business_access"`    // "approved"|"pending"，devices.access 派生
	AllowAccess       bool   `db:"allow_access"`       // alias of devices.access
	MonitoringEnabled bool   `db:"monitoring_enabled"` // devices.monitoring_enabled

	// 物理属性（dfm + drs）
	DeviceType      sql.NullString `db:"device_type"`      // dfm.device_type
	DeviceModel     sql.NullString `db:"device_model"`     // dfm.device_model
	IMEI            sql.NullString `db:"imei"`             // dfm.imei
	CommMode        sql.NullString `db:"comm_mode"`        // dfm.comm_mode
	MCUModel        sql.NullString `db:"mcu_model"`        // dfm.mcu_model
	FirmwareVersion sql.NullString `db:"firmware_version"` // drs.firmware_version

	// v2: 不持久化的扩展槽（v1 devices.metadata JSONB 已删；caller 临时挂任意 JSON 用）
	Metadata sql.NullString `db:"-"` // 不写 DB；保留字段让 ToJSON 兼容旧 API
}

/*
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

// DeviceWithLocation 带位置信息的设备结构
// 用于 Auth 缓存和消息处理
type DeviceWithLocation struct {
	Device       *Device
	LocationInfo *DeviceLocationInfo
}


*/

// DeviceStatusItem 设备在线状态项（用于批量查询 API）
// device_id 与 device_uid 均保留，前端可能持其一查询，结构保持一致
type DeviceStatusItem struct {
	DeviceUID string `json:"device_uid"`
	DeviceID  string `json:"device_id"`
	TenantID  string `json:"tenant_id"`
	Status    string `json:"status"` // "online", "offline", "unsubscribed"
}



var (
	// DeviceCache 设备缓存（key: uid, value: *DeviceWithLocation）
	// 在 Auth 时填充，供 MQTT Consumer 使用，避免重复查询数据库
	DeviceCache = &sync.Map{}

	// AllowAccessCache 设备认证状态缓存（key: uid, value: bool）
	// true = 设备已认证允许访问，false = 设备未认证或认证失败
	// 由 auth_service 维护，mqtt_consumer 仅查询此缓存以决定是否处理消息
	AllowAccessCache = &sync.Map{}
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
