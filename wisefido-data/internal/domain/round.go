package domain

import "time"

// Round 巡房记录领域模型（对应 rounds 表）
// 记录自动巡房的执行记录
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

	// 巡房时间
	RoundTime time.Time `db:"round_time"` // TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP

	// 巡房说明/备注
	Notes string `db:"notes"` // TEXT, nullable

	// 巡房状态
	Status string `db:"status"` // VARCHAR(20), NOT NULL, DEFAULT 'completed' ('draft'/'completed'/'cancelled')
}

