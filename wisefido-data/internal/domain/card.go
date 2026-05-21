package domain

import (
	"database/sql"
	"time"
)

// Card 卡片领域模型（v2.5: card_id INET PK，无 card_type/is_active/enabled_at/disabled_at 列）
//
// v2 列：card_id / unit_id / card_name / card_dns / resident_id / has_bed / has_bathroom / has_kitchen
// 派生字段（无 DB 列，由 SELECT 别名注入）：SpatialPrefix=CardID / CardType=masklen+name CASE / IsActive=TRUE
type Card struct {
	CardID        string         `db:"card_id"`
	SpatialPrefix string         `db:"-"` // 派生：等于 CardID（v2 无独立列）
	CardType      string         `db:"-"` // 派生：masklen + card_name CASE
	CardName      sql.NullString `db:"card_name"`
	DNSShortName  sql.NullString `db:"card_dns"`
	ResidentID    sql.NullString `db:"resident_id"`
	IsActive      bool           `db:"-"` // 派生：v2.5 无 is_active 列，always TRUE
	EnabledAt     sql.NullTime   `db:"-"` // v2.5 无列
	DisabledAt    sql.NullTime   `db:"-"` // v2.5 无列
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

