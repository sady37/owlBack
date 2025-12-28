package domain

// ConfigVersionUpdate 配置版本更新模型（对应 config_versions 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（version_id, tenant_id, config_type, entity_id）和必填字段（config_data, valid_from）不在更新模型中
type ConfigVersionUpdate struct {
	// 关联到当前实体表的ID（可选，用于已删除的实体）
	CurrentEntityID *UpdateString // UUID, nullable

	// 配置数据快照（JSONB）
	ConfigData *UpdateJSON // JSONB, NOT NULL（但更新时可以为空，表示不更新）

	// 版本生效时间区间
	ValidFrom *UpdateTime // TIMESTAMPTZ, NOT NULL（但更新时可以为空，表示不更新）
	ValidTo   *UpdateTime // TIMESTAMPTZ, nullable
}

