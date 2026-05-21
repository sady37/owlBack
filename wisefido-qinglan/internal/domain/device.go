package domain

import (
	"database/sql"
	"encoding/json"
	"sync"
)

// Device 设备领域模型。v2 schema：devices (device_ipv6 PK, device_id UUID FK, access, monitoring_enabled)。
// 其余字段由 JOIN device_factory_meta 取，或由 device_ipv6 prefix 派生。
type Device struct {
	DeviceID  string `db:"device_id"`  // UUID, dfm.device_id
	DeviceUID string `db:"device_uid"` // VARCHAR(50), dfm.device_uid (logMAC)

	DeviceIPv6 string `db:"device_ipv6"` // INET /128, devices.device_ipv6 (text repr)
	TenantID   string `db:"-"`           // 派生：host(network(set_masklen(device_ipv6, 48)))

	DeviceName string `db:"-"` // 派生：COALESCE(dfm.device_code, dfm.device_uid)

	BoundRoomID sql.NullString `db:"-"` // /88 prefix 派生 host(...)::text；mask <88 时 NULL
	BoundBedID  sql.NullString `db:"-"` // /96 prefix 派生 host(...)::text；mask <96 时 NULL

	Status            string `db:"-"` // online/offline；Redis device:status:{ipv6} 读
	Access            bool   `db:"access"`
	MonitoringEnabled bool   `db:"monitoring_enabled"`

	DeviceType      sql.NullString `db:"device_type"`      // dfm.device_type
	DeviceModel     sql.NullString `db:"device_model"`     // dfm.device_model
	IMEI            sql.NullString `db:"imei"`             // dfm.imei
	CommMode        sql.NullString `db:"comm_mode"`        // dfm.comm_mode
	MCUModel        sql.NullString `db:"mcu_model"`        // dfm.mcu_model
	FirmwareVersion sql.NullString `db:"firmware_version"` // dfm.firmware_version

	Metadata sql.NullString `db:"-"` // 非持久化扩展槽
}

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
		"access":             d.Access,
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
