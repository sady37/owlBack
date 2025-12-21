package domain

import "encoding/json"

// Tenant 租户领域模型（对应 tenants 表）
// 基于实际DB表结构：7个字段
type Tenant struct {
	// 主键
	TenantID string `db:"tenant_id"` // UUID, PRIMARY KEY

	// 基本信息
	TenantName string `db:"tenant_name"` // VARCHAR(255), NOT NULL
	Domain     string `db:"domain"`      // VARCHAR(255), UNIQUE, nullable
	Email      string `db:"email"`      // VARCHAR(255), nullable
	Phone      string `db:"phone"`      // VARCHAR(50), nullable

	// 状态
	Status string `db:"status"` // VARCHAR(50), DEFAULT 'active' (active/suspended/deleted)

	// 扩展配置
	Metadata json.RawMessage `db:"metadata"` // JSONB, nullable
}

