package domain

// AlarmEventUpdate 报警事件更新模型（对应 alarm_events 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（event_id, tenant_id, device_id）和必填字段（event_type, alarm_level, alarm_status, triggered_at）不在更新模型中
type AlarmEventUpdate struct {
	// 事件类型和级别
	Category   *UpdateString // VARCHAR(50), nullable, CHECK IN ('safety', 'clinical', 'behavioral', 'device')
	AlarmLevel *UpdateString // VARCHAR(20), NOT NULL（但更新时可以为空，表示不更新）

	// 报警状态
	AlarmStatus *UpdateString // VARCHAR(20), NOT NULL（但更新时可以为空，表示不更新）

	// 时间信息
	HandTime *UpdateTime // TIMESTAMPTZ, nullable

	// 触发数据
	IoTTimeSeriesID *UpdateInt64         // BIGINT, nullable
	TriggerData     *UpdateJSON          // JSONB, nullable

	// 处理信息
	Handler   *UpdateString // UUID, nullable, REFERENCES users(user_id)
	Operation *UpdateString // VARCHAR(30), nullable
	Notes     *UpdateString // TEXT, nullable

	// 通知信息
	NotifiedUsers *UpdateJSON // JSONB, DEFAULT '[]'::JSONB

	// 元数据
	Metadata *UpdateJSON // JSONB, DEFAULT '{}'::JSONB
}

