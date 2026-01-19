package domain

import (
	"database/sql"
)

// DeviceStore 设备库存领域模型（对应 device_store 表）
// 基于实际DB表结构：device_store表的所有字段
type DeviceStore struct {
	// 主键
	DeviceID string `db:"device_id"` // PRIMARY KEY (UUID)

	// 设备唯一标识（用于首次连接匹配和查询）
	DeviceUID  string         `db:"device_uid"`  // UNIQUE, 用于设备识别和查找
	DeviceCode sql.NullString `db:"device_code"` // nullable, sleepace 平台设备编码

	// 设备类型（必填）
	DeviceType  string         `db:"device_type"`  // NOT NULL
	DeviceModel sql.NullString `db:"device_model"` // nullable

	// MAC/IMEI
	MAC  sql.NullString `db:"mac"`  // nullable, mac address for wifi devices
	IMEI sql.NullString `db:"imei"` // nullable, 4G device IMEI

	// 物理属性
	CommMode sql.NullString `db:"comm_mode"` // nullable
	MCUModel sql.NullString `db:"mcu_model"` // nullable

	// 固件版本
	FirmwareVersion          sql.NullString `db:"firmware_version"`            // nullable
	OTATargetFirmwareVersion sql.NullString `db:"ota_target_firmware_version"` // nullable
	OTATargetMCUModel        sql.NullString `db:"ota_target_mcu_model"`        // nullable

	// 租户分配
	TenantID string `db:"tenant_id"` // NOT NULL, default '00000000-0000-0000-0000-000000000000'

	// 时间戳
	ImportDate   sql.NullTime `db:"import_date"`   // NOT NULL, default CURRENT_TIMESTAMP
	AllocateTime sql.NullTime `db:"allocate_time"` // nullable

	// 系统级访问权限
	AllowAccess bool `db:"allow_access"` // NOT NULL, default false

	// 关联租户名称（查询时JOIN获取，不存储在device_store表）
	TenantName sql.NullString `db:"tenant_name"` // 仅用于查询结果
}

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *DeviceStore) ToJSON() map[string]any {
	m := map[string]any{
		"device_id":    d.DeviceID,
		"device_uid":   d.DeviceUID,
		"device_type":  d.DeviceType,
		"tenant_id":    d.TenantID,
		"allow_access": d.AllowAccess,
	}
	if d.DeviceCode.Valid {
		m["device_code"] = d.DeviceCode.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
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
	if d.OTATargetFirmwareVersion.Valid {
		m["ota_target_firmware_version"] = d.OTATargetFirmwareVersion.String
	}
	if d.OTATargetMCUModel.Valid {
		m["ota_target_mcu_model"] = d.OTATargetMCUModel.String
	}
	if d.TenantName.Valid {
		m["tenant_name"] = d.TenantName.String
	}
	if d.ImportDate.Valid {
		m["import_date"] = d.ImportDate.Time.Format("2006-01-02 15:04:05")
	}
	if d.AllocateTime.Valid {
		m["allocate_time"] = d.AllocateTime.Time.Format("2006-01-02 15:04:05")
	}
	return m
}
