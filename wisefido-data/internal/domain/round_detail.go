package domain

import "time"

// RoundDetail 巡房详细记录领域模型（对应 round_details 表）
// 记录巡房中每个住户的详细状态信息
type RoundDetail struct {
	// 主键
	DetailID string `db:"detail_id"` // UUID, PRIMARY KEY

	// 租户和巡房记录
	TenantID string `db:"tenant_id"` // UUID, NOT NULL, FK to tenants
	RoundID  string `db:"round_id"`  // UUID, NOT NULL, FK to rounds

	// 关联住户和位置
	ResidentID string `db:"resident_id"` // UUID, NOT NULL, FK to residents
	BedID      string `db:"bed_id"`      // UUID, nullable, FK to beds
	UnitID     string `db:"unit_id"`     // UUID, nullable, FK to units

	// 自动获取的状态数据（从 iot_timeseries 获取）
	BedStatus string `db:"bed_status"` // VARCHAR(20) - 'in_bed'/'out_of_bed'/'unknown'

	// 睡眠状态（仅当在床时有效）
	SleepStateSNOMEDCode string `db:"sleep_state_snomed_code"` // VARCHAR(50), nullable
	SleepStateDisplay    string `db:"sleep_state_display"`     // VARCHAR(100), nullable

	// 生命体征（仅当在床时有效）
	HeartRate       *int `db:"heart_rate"`        // INTEGER, nullable - 心率（bpm）
	RespiratoryRate *int `db:"respiratory_rate"`  // INTEGER, nullable - 呼吸率（次/分钟）

	// 非在床时的姿态（仅当离床时有效）
	PostureSNOMEDCode string `db:"posture_snomed_code"` // VARCHAR(50), nullable
	PostureDisplay    string `db:"posture_display"`     // VARCHAR(100), nullable

	// 数据获取时间戳
	DataTimestamp *time.Time `db:"data_timestamp"` // TIMESTAMPTZ, nullable - 用于标识数据的新鲜度

	// 手动填写的信息
	Notes string `db:"notes"` // TEXT, nullable - 备注（手动填写）
}

