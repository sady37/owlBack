package domain

import "time"

// RoundItemChecked 一轮巡房中某一项的勾选快照（用于审计）
type RoundItemChecked struct {
	Key       string `json:"key"`        // 如 unhandled, leftbed, inbed
	Checked   bool   `json:"checked"`    // 是否已勾选
	Count     int    `json:"count"`      // 勾选时该类别数量
	CheckedAt string `json:"checked_at"` // ISO8601 可选
}

// Round 巡房记录领域模型（对应 rounds 表）
// 记录自动巡房的执行记录；round_type='manual' 时对应 Overview Automatic Rounds 完成的一轮
type Round struct {
	// 主键
	RoundID string `db:"round_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL, FK to tenants

	// 巡房类型
	RoundType string `db:"round_type"` // VARCHAR(20), NOT NULL, DEFAULT 'location' ('location'/'manual'/'scheduled')

	// 关联位置（可选，用于按位置巡房）
	UnitID string `db:"unit_id"` // UUID, nullable, FK to units

	// 执行人
	ExecutorID string `db:"executor_id"` // UUID, NOT NULL, FK to users

	// 巡房时间（完成时间）
	RoundTime time.Time `db:"round_time"` // TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP

	// 本轮开始时间（可选，manual 轮可填）
	StartedAt *time.Time `db:"started_at"` // TIMESTAMPTZ, nullable

	// 本轮勾选项快照（manual 轮：Unhandled/In Bed/Left Bed 等勾选与数量）
	ItemsChecked []byte `db:"items_checked"` // JSONB, nullable

	// 巡房说明/备注
	Notes string `db:"notes"` // TEXT, nullable

	// 巡房状态
	Status string `db:"status"` // VARCHAR(20), NOT NULL, DEFAULT 'completed' ('draft'/'completed'/'cancelled')
}

