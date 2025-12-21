package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo repository.RolesRepository
	logger   *zap.Logger
}

// NewRoleService 创建角色服务
func NewRoleService(roleRepo repository.RolesRepository, logger *zap.Logger) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// SystemTenantID 系统租户ID
const SystemTenantID = "00000000-0000-0000-0000-000000000001"

// ProtectedRoles 受保护的关键系统角色（不能禁用）
var ProtectedRoles = []string{"SystemAdmin", "SystemOperator", "Admin", "Manager", "Caregiver", "Resident", "Family"}

// ListRolesRequest 查询角色列表请求
type ListRolesRequest struct {
	TenantID *string
	Search   string
	Page     int
	Size     int
}

// ListRolesResponse 查询角色列表响应
type ListRolesResponse struct {
	Items []RoleItem `json:"items"`
	Total int        `json:"total"`
}

// RoleItem 角色项（前端格式）
type RoleItem struct {
	RoleID      string  `json:"role_id"`
	TenantID    *string `json:"tenant_id"`
	RoleCode    string  `json:"role_code"`
	DisplayName string  `json:"display_name"`
	Description string  `json:"description"`
	IsSystem    bool    `json:"is_system"`
	IsActive    bool    `json:"is_active"`
}

// ListRoles 查询角色列表
func (s *RoleService) ListRoles(ctx context.Context, req ListRolesRequest) (*ListRolesResponse, error) {
	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	// 构建过滤器
	filter := repository.RolesFilter{
		Search: strings.TrimSpace(req.Search),
	}

	// 查询角色列表
	roles, total, err := s.roleRepo.ListRoles(ctx, req.TenantID, filter, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// 转换为前端格式
	items := make([]RoleItem, 0, len(roles))
	for _, role := range roles {
		item := s.roleToItem(role)
		items = append(items, item)
	}

	return &ListRolesResponse{
		Items: items,
		Total: total,
	}, nil
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	TenantID    string
	RoleCode    string
	DisplayName string
	Description string
}

// CreateRoleResponse 创建角色响应
type CreateRoleResponse struct {
	RoleID string `json:"role_id"`
}

// CreateRole 创建角色（非系统角色）
func (s *RoleService) CreateRole(ctx context.Context, req CreateRoleRequest) (*CreateRoleResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	if req.RoleCode == "" {
		return nil, fmt.Errorf("role_code is required")
	}

	// 构建描述（两行格式：第一行显示名称，第二行详细描述）
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = req.RoleCode
	}
	fullDesc := displayName
	if strings.TrimSpace(req.Description) != "" {
		fullDesc = fullDesc + "\n" + strings.TrimSpace(req.Description)
	}

	// 创建角色领域模型
	role := &domain.Role{
		RoleCode:    req.RoleCode,
		Description: fullDesc,
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	if req.TenantID != SystemTenantID {
		role.TenantID = sql.NullString{String: req.TenantID, Valid: true}
	}

	// 调用 Repository
	roleID, err := s.roleRepo.CreateRole(ctx, req.TenantID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &CreateRoleResponse{
		RoleID: roleID,
	}, nil
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleID      string
	UserRole    string // 用于权限检查
	DisplayName *string
	Description *string
	IsActive    *bool
	Delete      *bool
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(ctx context.Context, req UpdateRoleRequest) error {
	// 参数验证
	if req.RoleID == "" {
		return fmt.Errorf("role_id is required")
	}

	// 获取当前角色
	role, err := s.roleRepo.GetRole(ctx, req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 处理删除
	if req.Delete != nil && *req.Delete {
		if role.IsSystem {
			return fmt.Errorf("system roles cannot be deleted")
		}
		return s.roleRepo.DeleteRole(ctx, req.RoleID)
	}

	// 处理状态更新
	if req.IsActive != nil {
		// 检查是否为受保护角色
		if !*req.IsActive {
			for _, protected := range ProtectedRoles {
				if role.RoleCode == protected {
					return fmt.Errorf("%s is a critical system role and cannot be disabled", role.RoleCode)
				}
			}
		}
		role.IsActive = sql.NullBool{Bool: *req.IsActive, Valid: true}
		return s.roleRepo.UpdateRole(ctx, req.RoleID, role)
	}

	// 处理字段更新
	if role.IsSystem {
		// 系统角色只能由 SystemAdmin 修改
		if !strings.EqualFold(req.UserRole, "SystemAdmin") {
			return fmt.Errorf("system roles can only be modified by SystemAdmin")
		}
	}

	// 更新显示名称和描述
	if req.DisplayName != nil || req.Description != nil {
		displayName := role.RoleCode
		if req.DisplayName != nil {
			displayName = strings.TrimSpace(*req.DisplayName)
			if displayName == "" {
				displayName = role.RoleCode
			}
		} else {
			// 从现有描述中提取显示名称
			if p := strings.SplitN(role.Description, "\n", 2); len(p) > 0 && strings.TrimSpace(p[0]) != "" {
				displayName = strings.TrimSpace(p[0])
			}
		}

		desc := ""
		if req.Description != nil {
			desc = strings.TrimSpace(*req.Description)
		} else if len(strings.SplitN(role.Description, "\n", 2)) > 1 {
			desc = strings.TrimSpace(strings.SplitN(role.Description, "\n", 2)[1])
		}

		fullDesc := displayName
		if desc != "" {
			fullDesc = fullDesc + "\n" + desc
		}
		role.Description = fullDesc
	}

	return s.roleRepo.UpdateRole(ctx, req.RoleID, role)
}

// roleToItem 将领域模型转换为前端格式
func (s *RoleService) roleToItem(role *domain.Role) RoleItem {
	item := RoleItem{
		RoleID:   role.RoleID,
		RoleCode: role.RoleCode,
		IsSystem: role.IsSystem,
	}

	if role.TenantID.Valid {
		item.TenantID = &role.TenantID.String
	}

	// 提取显示名称和描述
	if p := strings.SplitN(role.Description, "\n", 2); len(p) > 0 && strings.TrimSpace(p[0]) != "" {
		item.DisplayName = strings.TrimSpace(p[0])
	} else {
		item.DisplayName = role.RoleCode
	}
	item.Description = role.Description

	if role.IsActive.Valid {
		item.IsActive = role.IsActive.Bool
	} else {
		item.IsActive = true // 默认启用
	}

	return item
}

