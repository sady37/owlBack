package domain

import (
	"encoding/json"
	"time"
)

// AlarmEvent 报警事件领域模型（对应 alarm_events 表）
type AlarmEvent struct {
	// 主键
	EventID string `db:"event_id"` // UUID, PRIMARY KEY

	// 租户和设备关联
	TenantID string `db:"tenant_id"` // UUID, NOT NULL
	DeviceID string `db:"device_id"` // UUID, NOT NULL

	// 事件类型和级别
	EventType  string `db:"event_type"`  // VARCHAR(50), NOT NULL
	Category   string `db:"category"`   // VARCHAR(50), CHECK IN ('safety', 'clinical', 'behavioral', 'device')
	AlarmLevel string `db:"alarm_level"` // VARCHAR(20), NOT NULL

	// 报警状态
	AlarmStatus string `db:"alarm_status"` // VARCHAR(20), DEFAULT 'active', CHECK IN ('active', 'acknowledged')

	// 时间信息
	TriggeredAt time.Time  `db:"triggered_at"` // TIMESTAMPTZ, NOT NULL
	HandTime    *time.Time `db:"hand_time"`     // TIMESTAMPTZ, nullable

	// 触发数据
	IoTTimeSeriesID *int64         `db:"iot_timeseries_id"` // BIGINT, nullable
	TriggerData     json.RawMessage `db:"trigger_data"`      // JSONB

	// 处理信息
	Handler   *string `db:"handler"`   // UUID, nullable, REFERENCES users(user_id)
	Operation *string `db:"operation"` // VARCHAR(30), nullable
	Notes     *string `db:"notes"`     // TEXT, nullable

	// 通知信息
	NotifiedUsers json.RawMessage `db:"notified_users"` // JSONB, DEFAULT '[]'::JSONB

	// 元数据
	Metadata json.RawMessage `db:"metadata"` // JSONB, DEFAULT '{}'::JSONB

	// 时间戳
	CreatedAt time.Time `db:"created_at"` // TIMESTAMPTZ, NOT NULL
	UpdatedAt time.Time `db:"updated_at"` // TIMESTAMPTZ, NOT NULL
}

