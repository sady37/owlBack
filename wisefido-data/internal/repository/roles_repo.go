package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// RolesRepository 角色Repository接口
// 使用强类型领域模型，不使用map[string]any
type RolesRepository interface {
	// 查询
	GetRole(ctx context.Context, roleID string) (*domain.Role, error)
	GetRoleByCode(ctx context.Context, tenantID *string, roleCode string) (*domain.Role, error)
	ListRoles(ctx context.Context, tenantID *string, filter RolesFilter, page, size int) ([]*domain.Role, int, error)

	// 创建（Repository层不限制业务规则，业务规则在Service层验证）
	CreateRole(ctx context.Context, tenantID string, role *domain.Role) (string, error)

	// 更新（Repository层只做数据完整性验证，业务规则在Service层验证）
	UpdateRole(ctx context.Context, roleID string, role *domain.Role) error

	// 删除（Repository层只做数据完整性验证，业务规则在Service层验证）
	DeleteRole(ctx context.Context, roleID string) error
}

// RolesFilter 角色查询过滤器
type RolesFilter struct {
	Search    string // 模糊搜索 role_code, description
	IsSystem  *bool  // 可选，按is_system过滤
	IsActive  *bool  // 可选，按is_active过滤
}

