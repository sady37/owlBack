package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// RolePermissionsRepository 角色权限Repository接口
// 使用强类型领域模型，不使用map[string]any
// Repository层不限制业务规则，只负责数据访问和数据完整性验证
type RolePermissionsRepository interface {
	// 查询
	GetPermission(ctx context.Context, permissionID string) (*domain.RolePermission, error)
	GetPermissionByKey(ctx context.Context, tenantID *string, roleCode, resourceType, permissionType string) (*domain.RolePermission, error)
	ListPermissions(ctx context.Context, tenantID *string, filter RolePermissionsFilter, page, size int) ([]*domain.RolePermission, int, error)
	GetPermissionsByRole(ctx context.Context, tenantID *string, roleCode string) ([]*domain.RolePermission, error)
	GetPermissionsByResource(ctx context.Context, tenantID *string, resourceType string) ([]*domain.RolePermission, error)

	// 创建
	CreatePermission(ctx context.Context, tenantID string, permission *domain.RolePermission) (string, error)
	BatchCreatePermissions(ctx context.Context, tenantID string, permissions []*domain.RolePermission) (int, []error, error)

	// 更新
	UpdatePermission(ctx context.Context, permissionID string, permission *domain.RolePermission) error
	BatchUpdatePermissions(ctx context.Context, updates []PermissionUpdate) (int, []error, error)

	// 删除
	DeletePermission(ctx context.Context, permissionID string) error
	DeletePermissionsByRole(ctx context.Context, tenantID, roleCode string) error
}

// RolePermissionsFilter 权限查询过滤器
type RolePermissionsFilter struct {
	RoleCode       string // 可选，按role_code过滤
	ResourceType   string // 可选，按resource_type过滤
	PermissionType string // 可选，按permission_type过滤（'R', 'C', 'U', 'D'）
	AssignedOnly   *bool  // 可选，按assigned_only过滤
	BranchOnly     *bool  // 可选，按branch_only过滤
}

// PermissionUpdate 权限更新请求
type PermissionUpdate struct {
	PermissionID string
	Permission   *domain.RolePermission
}

