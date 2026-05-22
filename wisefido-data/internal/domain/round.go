package domain

import "time"

// RoundItemChecked 一轮巡房中某一项的勾选快照（用于审计）
type RoundItemChecked struct {
	Key       string `json:"key"`
	Checked   bool   `json:"checked"`
	Count     int    `json:"count"`
	CheckedAt string `json:"checked_at"`
}

// Round 巡房记录（对应 rounds 表 + rounds_report_snapshot.snapshot）
type Round struct {
	RoundID   string `db:"round_id"`
	TenantID  string `db:"tenant_id"` // INET /48 snapshot at create
	UserID    string `db:"user_id"`   // UUID
	RoundType string `db:"round_type"`
	UnitID    string `db:"unit_id"` // INET CIDR (unit_prefix)
	// 时间窗：started_at 必填；ended_at NULL = in_progress
	StartedAt *time.Time `db:"started_at"`
	EndedAt   *time.Time `db:"ended_at"`
	Notes     string     `db:"notes"`
	Status    string     `db:"status"`

	// rounds_report_snapshot.snapshot JSONB 完整 blob
	// 形态: {"items_checked": [...], "schema_version": N, "facility_license": "...", "rows": [...]}
	Snapshot []byte `db:"snapshot"`

	// 显示派生（LEFT JOIN users/tenants/units 后填充）
	ExecutorDisplay string `db:"executor_display"`
	FacilityName    string `db:"facility_name"`
	UnitName        string `db:"unit_name"`
}
