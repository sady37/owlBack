package domain

// AlarmCloudUpdate 云端告警策略更新模型（对应 alarm_cloud 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
type AlarmCloudUpdate struct {
	// 通用报警（全局的，与具体设备型号无关）
	OfflineAlarm  *UpdateString // VARCHAR(20), nullable, DEFAULT 'ERROR' - DangerLevel
	LowBattery    *UpdateString // VARCHAR(20), nullable, DEFAULT 'WARNING' - DangerLevel
	DeviceFailure *UpdateString // VARCHAR(20), nullable, DEFAULT 'ERROR' - DangerLevel

	// 设备特定报警配置（JSONB）
	DeviceAlarms *UpdateJSON // JSONB, NOT NULL, DEFAULT '{}'::jsonb

	// 报警阈值配置（JSONB）
	Conditions *UpdateJSON // JSONB, nullable

	// 通知规则（JSONB）
	NotificationRules *UpdateJSON // JSONB, nullable

	// 元数据
	Metadata *UpdateJSON // JSONB, nullable
}

