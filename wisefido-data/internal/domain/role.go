package domain

import (
	"database/sql"
)

// Role 角色领域模型（对应 roles 表）
// 基于实际DB表结构：6个字段
type Role struct {
	// 主键和租户
	RoleID   string `db:"role_id"`
	TenantID sql.NullString `db:"tenant_id"` // nullable: System tenant = 系统预定义角色，其他 = 租户自定义角色

	// 角色信息
	RoleCode    string `db:"role_code"`    // NOT NULL: 角色代码，用于程序引用
	Description string `db:"description"`  // NOT NULL: 两行格式，第一行角色名称，第二行详细描述

	// 系统角色标识
	IsSystem bool `db:"is_system"` // NOT NULL DEFAULT FALSE: 是否为系统预定义角色
	IsActive sql.NullBool `db:"is_active"` // DEFAULT TRUE: 是否启用
}

