package domain

import (
	"encoding/json"
	"time"
)

// ConfigVersion 配置版本领域模型（对应 config_versions 表）
// 统一配置历史表，按时间保存所有配置类型的快照
type ConfigVersion struct {
	// 主键
	VersionID string `db:"version_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL, FK to tenants

	// 配置类型
	ConfigType string `db:"config_type"` // VARCHAR(50), NOT NULL - 'room_layout'/'device_config'/'alarm_cloud'/'alarm_device'/'device_installation'

	// 实体关联
	EntityID string `db:"entity_id"` // UUID, NOT NULL

	// 关联到当前实体表的ID（可选，用于已删除的实体）
	CurrentEntityID string `db:"current_entity_id"` // UUID, nullable

	// 配置数据快照（JSONB）
	ConfigData json.RawMessage `db:"config_data"` // JSONB, NOT NULL - 存储完整的配置内容

	// 版本生效时间区间：[valid_from, valid_to)
	ValidFrom time.Time  `db:"valid_from"` // TIMESTAMPTZ, NOT NULL - 配置开始生效时间
	ValidTo   *time.Time `db:"valid_to"`   // TIMESTAMPTZ, nullable - 配置失效时间（NULL表示当前仍生效）
}

