package domain

import (
	"database/sql"
	"encoding/json"
)

// Device 设备领域模型（Phase 2 一刀切后）：
//   - devices PK = device_addr INET /128
//   - 硬件 identity = device_uid VARCHAR(50)（dfm PK；FK devices.device_uid）
//   - 其余字段由 JOIN device_factory_meta / device_ota 取，或由 device_addr prefix 派生。
type Device struct {
	DeviceAddr string `db:"device_addr"` // INET /128，devices.device_addr；同时是 redis device:status key
	DeviceUID  string `db:"device_uid"`  // logMAC，硬件 identity 不变量（dfm PK）
	TenantID   string `db:"-"`           // 派生：host(network(set_masklen(device_addr, 48)))

	// 物理属性（JOIN device_factory_meta 取）
	DeviceType  sql.NullString `db:"-"`
	DeviceModel sql.NullString `db:"-"`
	DeviceCode  sql.NullString `db:"-"`
	MAC         sql.NullString `db:"-"` // 来自 dfm.mac_wifi
	IMEI        sql.NullString `db:"-"`
	CommMode    sql.NullString `db:"-"`
	MCUModel    sql.NullString `db:"-"`
	FirmwareVersion sql.NullString `db:"-"`

	// OTA 字段（JOIN device_ota 取）
	OTATargetFW  sql.NullString `db:"-"`
	OTATargetMCU sql.NullString `db:"-"`
	OTAPermit    sql.NullString `db:"-"`
	OTAWay       sql.NullString `db:"-"`
	OTASchedule  sql.NullString `db:"-"`
	OTAStatus    sql.NullString `db:"-"`
	OTAProgress  sql.NullInt32  `db:"-"`
	OTATenantApproved bool      `db:"-"`

	DeviceName string `db:"-"` // 派生：COALESCE(dfm.device_code, dfm.device_uid)

	// 位置绑定 — 派生自 device_addr 字节段
	BoundRoomID sql.NullString `db:"-"` // /88 prefix when byte 10 != 0
	BoundBedID  sql.NullString `db:"-"` // /96 prefix when byte 11 != 0
	RoomID      sql.NullString `db:"-"`
	UnitID      sql.NullString `db:"-"`
	CardID      sql.NullString `db:"-"`
	UnitName    sql.NullString `db:"-"`
	DNSShortName sql.NullString `db:"-"`
	BranchID    sql.NullString `db:"-"`
	BranchName  sql.NullString `db:"-"`

	// 状态
	Status            string `db:"-"` // SELECT alias 派生：'Enabled' / 'Disabled' 软删用；非物理列
	OnlineStatus      string `db:"-"` // Redis device:status:{addr} 读
	Access            bool   `db:"access"`
	MonitoringEnabled bool   `db:"monitoring_enabled"`

	Metadata sql.NullString `db:"-"` // v1 devices.metadata JSONB 已 drop；保留字段供 ToJSON 兼容
}

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_addr":        d.DeviceAddr,
		"device_uid":         d.DeviceUID,
		"tenant_id":          d.TenantID,
		"device_name":        d.DeviceName,
		"status":             d.Status,        // 数据库状态（Enabled/Disabled/Error）
		"online_status":      d.OnlineStatus,  // 实时在线状态（online/offline/unsubscribed）
		"access":             d.Access,
		"monitoring_enabled": d.MonitoringEnabled,
	}
	// 物理属性（从 device_store 获取）
	if d.DeviceType.Valid {
		m["device_type"] = d.DeviceType.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
	}
	if d.DeviceCode.Valid {
		m["device_code"] = d.DeviceCode.String
	}
	if d.MAC.Valid {
		m["mac"] = d.MAC.String
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
	if d.OTATargetFW.Valid {
		m["ota_target_firmware_version"] = d.OTATargetFW.String
	}
	if d.OTATargetMCU.Valid {
		m["ota_target_mcu_model"] = d.OTATargetMCU.String
	}
	if d.OTAPermit.Valid {
		m["ota_permit"] = d.OTAPermit.String
	}
	if d.OTAWay.Valid {
		m["ota_way"] = d.OTAWay.String
	}
	if d.OTASchedule.Valid {
		m["ota_schedule"] = d.OTASchedule.String
	}
	if d.OTAStatus.Valid {
		m["ota_status"] = d.OTAStatus.String
	} else {
		m["ota_status"] = "idle"
	}
	if d.OTAProgress.Valid {
		m["ota_progress"] = d.OTAProgress.Int32
	}
	m["ota_tenant_approved"] = d.OTATenantApproved
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
	// 位置信息（用于前端显示，通过 JOIN 查询得到）
	if d.RoomID.Valid {
		m["room_id"] = d.RoomID.String
	} else {
		m["room_id"] = nil
	}
	if d.UnitID.Valid {
		m["unit_id"] = d.UnitID.String
	} else {
		m["unit_id"] = nil
	}
	if d.CardID.Valid {
		m["card_id"] = d.CardID.String
	} else {
		m["card_id"] = nil
	}
	if d.UnitName.Valid {
		m["unit_name"] = d.UnitName.String
	} else {
		m["unit_name"] = nil
	}
	if d.DNSShortName.Valid && d.DNSShortName.String != "" {
		m["dns_short_name"] = d.DNSShortName.String
	}
	if d.BranchID.Valid {
		m["branch_id"] = d.BranchID.String
	} else {
		m["branch_id"] = nil
	}
	if d.BranchName.Valid {
		m["branch_name"] = d.BranchName.String
	} else {
		m["branch_name"] = nil
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
