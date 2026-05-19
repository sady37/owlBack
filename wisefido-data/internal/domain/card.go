package domain

import (
	"database/sql"
	"time"
)

// Card 卡片领域模型（v2: 对应新 cards 表 schema）
//
// 三层身份：
//   - spatial_prefix INET — 业务身份（北极星）
//   - card_id UUID        — DB PK 稳定性 / FK
//   - dns_short_name      — 永久人类可读名（bed-stable，无 PHI）
//
// 删除的 v1 字段：tenant_id / branch_id / bed_id / unit_id / card_address / timezone /
// devices JSONB / residents JSONB / unhandled_alarm_0..4 / pop_alarm_level / pop_alarm_type /
// pop_alarm_event_id（counter / pop 由 alarm_events 实时聚合）。
type Card struct {
	CardID        string         `db:"card_id"`        // UUID PK
	SpatialPrefix string         `db:"spatial_prefix"` // INET CIDR 字符串（业务身份；mask 决定 card_type）
	CardType      string         `db:"card_type"`      // 'tenant'|'branch'|'site'|'unit'|'public'|'room'|'bed'|'device'
	CardName      sql.NullString `db:"card_name"`
	DNSShortName  sql.NullString `db:"dns_short_name"` // 永久 DNS 名，如 "u42-r03-b01.tenant1.owl"
	ResidentID    sql.NullString `db:"resident_id"`    // INET HoA；NULL = 空床
	IsActive      bool           `db:"is_active"`
	EnabledAt     sql.NullTime   `db:"enabled_at"`
	DisabledAt    sql.NullTime   `db:"disabled_at"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// CardWithUnitInfo 卡片及其关联的 Unit 信息（用于 Repository 层返回）
// v2: Unit 信息由 spatial_prefix 通过 unit_id (INET /80) join units 表反查；
//     不再随 cards 表共存。
type CardWithUnitInfo struct {
	Card *Card
	Unit *Unit
}

