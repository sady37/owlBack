package domain

import "encoding/json"

// AlarmCloud 云端告警策略领域模型（对应 alarm_cloud 表）
// 定义设备类型有什么报警选项及默认设置参数
type AlarmCloud struct {
	// 主键
	TenantID string `db:"tenant_id"` // UUID, PRIMARY KEY, FK to tenants

	// 通用报警（全局的，与具体设备型号无关）
	// 这些是所有设备类型都支持的通用报警项
	// 如果为 NULL，表示使用全局默认值
	OfflineAlarm  string `db:"offline_alarm"`  // VARCHAR(20), nullable - DangerLevel
	LowBattery    string `db:"low_battery"`    // VARCHAR(20), nullable - DangerLevel
	DeviceFailure string `db:"device_failure"` // VARCHAR(20), nullable - DangerLevel

	// 设备特定报警配置（JSONB）
	// 包含所有设备类型的报警配置，第一层 key 为设备类型，第二层 key 为报警类型
	DeviceAlarms json.RawMessage `db:"device_alarms"` // JSONB, NOT NULL, DEFAULT '{}'::jsonb

	// 报警阈值配置（JSONB）
	// 用于生理指标类报警，定义什么数值范围触发什么级别
	Conditions json.RawMessage `db:"conditions"` // JSONB, nullable

	// 通知规则（JSONB）
	// 包含通知通道、发送方式、升级规则、抑制规则、静默规则等完整配置
	NotificationRules json.RawMessage `db:"notification_rules"` // JSONB, nullable

	// 元数据
	Metadata json.RawMessage `db:"metadata"` // JSONB, nullable
}

