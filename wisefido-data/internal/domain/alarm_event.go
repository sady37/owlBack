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
	TenantID string  `db:"tenant_id"` // UUID, NOT NULL
	DeviceAddr string  `db:"device_addr"` // INET /128, NOT NULL
	CardID   *string `db:"card_id"`   // UUID, nullable (FK → cards)

	// 事件类型和级别
	EventType  string `db:"event_type"`  // VARCHAR(50), NOT NULL
	Category   string `db:"category"`   // VARCHAR(50), CHECK IN ('safety', 'clinical', 'behavioral', 'device')
	AlarmLevel string `db:"alarm_level"` // VARCHAR(20), NOT NULL

	AlarmStatus string `db:"alarm_status"` // VARCHAR(20), DEFAULT 'active', CHECK IN ('active','acked','resolved','auto_resolved','expired')

	// 时间信息
	TriggeredAt time.Time  `db:"triggered_at"` // TIMESTAMPTZ, NOT NULL — 实际发生时刻（incident）
	AlertedAt   *time.Time `db:"alerted_at"`   // TIMESTAMPTZ, nullable — 系统决策上抛时刻；推断类 fall ＞ triggered_at；NULL = cutover 前历史行
	HandTime    *time.Time `db:"hand_time"`     // TIMESTAMPTZ, nullable

	// 触发数据
	TriggerData json.RawMessage `db:"trigger_data"` // JSONB

	// 处理信息
	Handler   *string `db:"handler"`   // UUID, nullable, REFERENCES users(user_id)
	Operation *string `db:"operation"` // VARCHAR(30), nullable
	Notes     *string `db:"notes"`     // TEXT, nullable

	// 元数据
	Metadata json.RawMessage `db:"metadata"` // JSONB, DEFAULT '{}'::JSONB

	// 时间戳
	CreatedAt time.Time `db:"created_at"` // TIMESTAMPTZ, NOT NULL
	UpdatedAt time.Time `db:"updated_at"` // TIMESTAMPTZ, NOT NULL
}

