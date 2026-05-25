package domain

import (
	"database/sql"
	"time"
)

// DeviceStore 设备库存领域模型。v2 没有 device_store 表 — 该 struct 是
// device_factory_meta + devices + device_ota 三表 JOIN 的扁平视图（v1 兼容 wire format）。
// 字段是 SELECT alias 派生，不对应单一物理列。
//
// device_id UUID 退役 → identity = device_uid VARCHAR(50)；
//                业务寻址 = device_addr INET /128（devices PK）。
type DeviceStore struct {
	DeviceAddr string `db:"device_addr"` // devices.device_addr INET /128；同时是 redis device:status key

	DeviceUID  string         `db:"device_uid"`  // dfm.device_uid (logMAC)，如 BM87224601903；硬件 identity 不变量
	DeviceCode sql.NullString `db:"device_code"` // dfm.device_code（厂家会话 ID）

	DeviceType  string         `db:"device_type"`  // dfm.device_type
	DeviceModel sql.NullString `db:"device_model"` // dfm.device_model

	DeviceName sql.NullString `db:"-"` // v2 无 device_name 列；派生 COALESCE(device_code, device_uid)

	MAC  sql.NullString `db:"mac"`  // SELECT alias 自 dfm.mac_wifi
	IMEI sql.NullString `db:"imei"` // dfm.imei

	CommMode sql.NullString `db:"comm_mode"` // dfm.comm_mode
	MCUModel sql.NullString `db:"mcu_model"` // dfm.mcu_model

	FirmwareVersion          sql.NullString `db:"firmware_version"`            // dfm.firmware_version
	OTATargetFirmwareVersion sql.NullString `db:"ota_target_firmware_version"` // device_ota.target_firmware_version
	OTATargetMCUModel        sql.NullString `db:"ota_target_mcu_model"`        // device_ota.target_mcu_model

	OTAPermit      sql.NullString `db:"ota_permit"`
	OTAWay         sql.NullString `db:"ota_way"`
	OTASchedule    sql.NullString `db:"ota_schedule"`
	OTAStatus      sql.NullString `db:"ota_status"`
	OTAProgress    sql.NullInt32  `db:"ota_progress"`
	OTAError       sql.NullString `db:"-"` // v2 device_ota 无此列
	OTAUpdatedAt   sql.NullTime   `db:"-"` // v2 device_ota 无此列
	OTAFirmwareURL sql.NullString `db:"-"` // v2 update.ini 管，不入库
	OTAFirmwareSHA sql.NullString `db:"-"` // v2 update.ini 管，不入库
	OTAFirmwareSize sql.NullInt64 `db:"-"` // v2 update.ini 管，不入库
	OTATenantApproved bool       `db:"ota_tenant_approved"`

	TenantID string `db:"-"` // v2 派生：host(network(set_masklen(devices.device_addr, 48)))

	ImportDate   sql.NullTime `db:"import_date"`
	AllocateTime sql.NullTime `db:"-"` // v2 无此列

	Access bool `db:"-"` // platform_admin 审批位，来自 devices.access

	// Batch PATCH：仅当为 true 时更新对应列（避免只改 device_code 时误把 access 写成 false）
	DeviceCodeSet bool `db:"-" json:"-"`
	AccessSet     bool `db:"-" json:"-"`
	OTAPermitSet   bool `db:"-" json:"-"`
	OTAWaySet      bool `db:"-" json:"-"`
	OTAScheduleSet bool `db:"-" json:"-"`
	OTAStatusSet    bool `db:"-" json:"-"`
	OTATargetFWSet  bool `db:"-" json:"-"`
	OTATargetMCUSet bool `db:"-" json:"-"`

	TenantName sql.NullString `db:"-"` // v2 无：派生 + JOIN 取

	OnlineStatus string `db:"-"` // Redis device:status:{addr} 读

	// 实时健康标志位（从 device:status:{addr} hash 读取，不存数据库）
	// 不论设备是否绑卡都会写入 hash，admin 视角无差别可见
	Offline        int   `db:"-"` // 0/1 网络/心跳维度（与 OnlineStatus 互补，前端展示用）
	SignalPoor     int   `db:"-"` // 0/1 WiFi 弱（设备仍能上行）
	AngleAbnormal  int   `db:"-"` // 0/1 雷达倾角异常（设备仍能上行）
	SensorDetached int   `db:"-"` // 0/1 Sleepad 传感器脱落
	LastSeenMs     int64 `db:"-"` // 最近上行时间（毫秒）
}

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *DeviceStore) ToJSON() map[string]any {
	m := map[string]any{
		"device_addr":  d.DeviceAddr,
		"device_uid":   d.DeviceUID,
		"device_type":  d.DeviceType,
		"tenant_id":    d.TenantID,
		"access":      d.Access,
	}
	if d.DeviceCode.Valid {
		m["device_code"] = d.DeviceCode.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
	}
	if d.DeviceName.Valid && d.DeviceName.String != "" {
		m["device_name"] = d.DeviceName.String
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
	if d.OTAError.Valid {
		m["ota_error"] = d.OTAError.String
	}
	if d.OTAUpdatedAt.Valid {
		m["ota_updated_at"] = d.OTAUpdatedAt.Time.Format(time.RFC3339)
	}
	if d.OTAFirmwareURL.Valid {
		m["ota_firmware_url"] = d.OTAFirmwareURL.String
	}
	if d.OTAFirmwareSHA.Valid {
		m["ota_firmware_sha256"] = d.OTAFirmwareSHA.String
	}
	if d.OTAFirmwareSize.Valid {
		m["ota_firmware_size"] = d.OTAFirmwareSize.Int64
	}
	m["ota_tenant_approved"] = d.OTATenantApproved
	if d.TenantName.Valid {
		m["tenant_name"] = d.TenantName.String
	}
	if d.ImportDate.Valid {
		m["import_date"] = d.ImportDate.Time.Format("2006-01-02 15:04:05")
	}
	if d.AllocateTime.Valid {
		m["allocate_time"] = d.AllocateTime.Time.Format("2006-01-02 15:04:05")
	}
	if d.OnlineStatus != "" {
		m["online_status"] = d.OnlineStatus
	} else {
		m["online_status"] = "offline"
	}
	// 健康标志位（仅在非零时输出，前端 undefined 兜底为 0）。omitempty 风格保持响应紧凑。
	m["offline"] = d.Offline
	if d.SignalPoor != 0 {
		m["signal_poor"] = d.SignalPoor
	}
	if d.AngleAbnormal != 0 {
		m["angle_abnormal"] = d.AngleAbnormal
	}
	if d.SensorDetached != 0 {
		m["sensor_detached"] = d.SensorDetached
	}
	if d.LastSeenMs > 0 {
		m["last_seen_ms"] = d.LastSeenMs
	}
	return m
}
