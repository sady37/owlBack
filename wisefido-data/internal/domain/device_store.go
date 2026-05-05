package domain

import (
	"database/sql"
	"time"
)

// DeviceStore 设备库存领域模型（对应 device_store 表）
// 基于实际DB表结构：device_store表的所有字段
type DeviceStore struct {
	// 主键
	DeviceID string `db:"device_id"` // PRIMARY KEY (UUID)

	// 设备唯一标识（用于首次连接匹配和查询）
	DeviceUID  string         `db:"device_uid"`  // UNIQUE; 对应 Sleepace deviceName，如 BM87224601903
	DeviceCode sql.NullString `db:"device_code"` // nullable; 对应 Sleepace device_id（平台侧 ID），如 1ua3erivl9pv1；绑定时传此值给 Sleepace

	// 设备类型（必填）
	DeviceType  string         `db:"device_type"`  // NOT NULL
	DeviceModel sql.NullString `db:"device_model"` // nullable

	// 设备显示名称（可选，导入/分配时若为空则默认 DeviceModel 或 DeviceType + - + UID 后4位，如 Radar-0523）
	DeviceName sql.NullString `db:"device_name"` // nullable

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

	// OTA 管理字段
	OTAPermit      sql.NullString `db:"ota_permit"`
	OTAWay         sql.NullString `db:"ota_way"`
	OTASchedule    sql.NullString `db:"ota_schedule"`
	OTAStatus      sql.NullString `db:"ota_status"`
	OTAProgress    sql.NullInt32  `db:"ota_progress"`
	OTAError       sql.NullString `db:"ota_error"`
	OTAUpdatedAt   sql.NullTime   `db:"ota_updated_at"`
	OTAFirmwareURL sql.NullString `db:"ota_firmware_url"`
	OTAFirmwareSHA sql.NullString `db:"ota_firmware_sha256"`
	OTAFirmwareSize sql.NullInt64 `db:"ota_firmware_size"`
	OTATenantApproved bool       `db:"ota_tenant_approved"`

	// 租户分配
	TenantID string `db:"tenant_id"` // NOT NULL default Unallocated 002; 000 trash 001 system

	// 时间戳
	ImportDate   sql.NullTime `db:"import_date"`   // NOT NULL, default CURRENT_TIMESTAMP
	AllocateTime sql.NullTime `db:"allocate_time"` // nullable

	// 系统级访问权限
	AllowAccess bool `db:"allow_access"` // NOT NULL, default false

	// Batch PATCH：仅当为 true 时更新对应列（避免只改 device_code 时误把 allow_access 写成 false）
	DeviceCodeSet  bool `db:"-" json:"-"`
	AllowAccessSet bool `db:"-" json:"-"`
	OTAPermitSet   bool `db:"-" json:"-"`
	OTAWaySet      bool `db:"-" json:"-"`
	OTAScheduleSet bool `db:"-" json:"-"`
	OTAStatusSet    bool `db:"-" json:"-"`
	OTATargetFWSet  bool `db:"-" json:"-"`
	OTATargetMCUSet bool `db:"-" json:"-"`

	// 关联租户名称（查询时JOIN获取，不存储在device_store表）
	TenantName sql.NullString `db:"tenant_name"` // 仅用于查询结果

	// 实时在线状态（从 Redis 读取，不存储到数据库）
	OnlineStatus string `db:"-"` // 实时在线状态（online/offline/unsubscribed），从 Redis 读取

	// 实时健康标志位（从 device:status:{deviceID} hash 读取，不存数据库）
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
