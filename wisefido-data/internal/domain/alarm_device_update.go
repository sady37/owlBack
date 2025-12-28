package domain

// AlarmDeviceUpdate 设备告警配置更新模型（对应 alarm_device 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
type AlarmDeviceUpdate struct {
	// 设备的完整监控配置（JSONB）
	MonitorConfig *UpdateJSON // JSONB, NOT NULL, DEFAULT '{"alarms": {}}'::jsonb

	// 厂家参考配置（JSONB，只读）
	VendorConfig *UpdateJSON // JSONB, nullable

	// 元数据
	Metadata *UpdateJSON // JSONB, nullable
}

