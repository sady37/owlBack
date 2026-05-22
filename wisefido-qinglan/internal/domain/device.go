package domain

import (
	"database/sql"
	"encoding/json"
	"sync"
)

// Device 设备领域模型（Phase 2 一刀切后）：
//   - devices PK = device_addr INET /128
//   - 硬件 identity = device_uid VARCHAR(50)（dfm PK；FK devices.device_uid）
type Device struct {
	DeviceUID string `db:"device_uid"` // VARCHAR(50)，dfm.device_uid (logMAC)，硬件 identity 不变量

	DeviceAddr string `db:"device_addr"` // INET /128, devices.device_addr (text repr)
	TenantID   string `db:"-"`           // 派生：host(network(set_masklen(device_addr, 48)))

	DeviceName string `db:"-"` // 派生：COALESCE(dfm.device_code, dfm.device_uid)

	BoundRoomID sql.NullString `db:"-"` // /88 prefix 派生 host(...)::text；mask <88 时 NULL
	BoundBedID  sql.NullString `db:"-"` // /96 prefix 派生 host(...)::text；mask <96 时 NULL

	Status            string `db:"-"` // online/offline；Redis device:status:{addr} 读
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

// DeviceStatusItem 设备在线状态项（用于批量查询 API）。
// Phase 2 一刀切：DeviceAddr (INET /128) 是业务侧寻址；DeviceUID 是硬件 identity。
type DeviceStatusItem struct {
	DeviceUID  string `json:"device_uid"`
	DeviceAddr string `json:"device_addr,omitempty"`
	TenantID   string `json:"tenant_id"`
	Status     string `json:"status"` // "online", "offline", "unsubscribed"
}

var (
	// DeviceCache 设备缓存（key: uid, value: *DeviceWithLocation）
	DeviceCache = &sync.Map{}

	// AccessCache 设备审批位缓存（key: uid, value: bool = devices.access）
	AccessCache = &sync.Map{}
)

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_addr":        d.DeviceAddr,
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
