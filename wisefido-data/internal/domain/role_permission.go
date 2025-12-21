package domain

import (
	"database/sql"
)

// RolePermission 角色权限领域模型（对应 role_permissions 表）
// 基于实际DB表结构：7个字段
type RolePermission struct {
	// 主键和租户
	PermissionID string         `db:"permission_id"`
	TenantID     sql.NullString `db:"tenant_id"` // nullable: System tenant = 系统预定义权限，其他 = 租户自定义权限

	// 权限标识
	RoleCode       string `db:"role_code"`        // NOT NULL: 角色代码（引用roles.role_code，非外键，通过值匹配）
	ResourceType   string `db:"resource_type"`    // NOT NULL: 资源类型（表名或资源标识符，如'users', 'residents'等）
	PermissionType string `db:"permission_type"`  // NOT NULL: 权限操作类型（'R', 'C', 'U', 'D'）

	// 权限范围
	AssignedOnly bool `db:"assigned_only"` // DEFAULT FALSE: 权限范围（FALSE=所有资源，TRUE=仅分配的资源）
	BranchOnly   bool `db:"branch_only"`   // DEFAULT FALSE: 权限范围（FALSE=所有资源，TRUE=仅同分支的资源）
}

