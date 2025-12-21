package domain

import "encoding/json"

// AlarmDevice 设备告警配置领域模型（对应 alarm_device 表）
// IoT设备实时报警配置表，为每个设备存储完整的监控配置
type AlarmDevice struct {
	// 主键
	DeviceID string `db:"device_id"` // UUID, PRIMARY KEY, FK to devices

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL, FK to tenants

	// 设备的完整监控配置（JSONB）
	// 包含睡眠时间、各报警项及其级别、阈值等
	// 由前端完全控制，DB只是保存
	MonitorConfig json.RawMessage `db:"monitor_config"` // JSONB, NOT NULL, DEFAULT '{"alarms": {}}'::jsonb

	// 厂家参考配置（JSONB，只读）
	// 厂家给的参考值，方便前端参考
	// 这是只读参考值，不会被修改
	VendorConfig json.RawMessage `db:"vendor_config"` // JSONB, nullable

	// 元数据
	Metadata json.RawMessage `db:"metadata"` // JSONB, nullable
}

