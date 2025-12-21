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

// RolePermissionService 角色权限服务
type RolePermissionService struct {
	permRepo repository.RolePermissionsRepository
	logger   *zap.Logger
}

// NewRolePermissionService 创建角色权限服务
func NewRolePermissionService(permRepo repository.RolePermissionsRepository, logger *zap.Logger) *RolePermissionService {
	return &RolePermissionService{
		permRepo: permRepo,
		logger:   logger,
	}
}

// ListPermissionsRequest 查询权限列表请求
type ListPermissionsRequest struct {
	TenantID      *string
	RoleCode      string
	ResourceType  string
	PermissionType string // "read", "create", "update", "delete", "manage"
	Page           int
	Size           int
}

// ListPermissionsResponse 查询权限列表响应
type ListPermissionsResponse struct {
	Items []PermissionItem `json:"items"`
	Total int              `json:"total"`
}

// PermissionItem 权限项（前端格式）
type PermissionItem struct {
	PermissionID   string  `json:"permission_id"`
	TenantID       *string `json:"tenant_id"`
	RoleCode       string  `json:"role_code"`
	ResourceType   string  `json:"resource_type"`
	PermissionType string  `json:"permission_type"` // "read", "create", "update", "delete"
	Scope          string  `json:"scope"`           // "all", "assigned_only"
	BranchOnly     bool    `json:"branch_only"`
	IsActive       bool    `json:"is_active"`
}

// ListPermissions 查询权限列表
func (s *RolePermissionService) ListPermissions(ctx context.Context, req ListPermissionsRequest) (*ListPermissionsResponse, error) {
	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 100
	}

	// 构建过滤器
	filter := repository.RolePermissionsFilter{
		RoleCode:       strings.TrimSpace(req.RoleCode),
		ResourceType:   strings.TrimSpace(req.ResourceType),
		PermissionType: s.permissionTypeToDB(strings.TrimSpace(req.PermissionType)),
	}

	// 查询权限列表
	permissions, total, err := s.permRepo.ListPermissions(ctx, req.TenantID, filter, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	// 转换为前端格式
	items := make([]PermissionItem, 0, len(permissions))
	for _, perm := range permissions {
		item := s.permissionToItem(perm)
		items = append(items, item)
	}

	return &ListPermissionsResponse{
		Items: items,
		Total: total,
	}, nil
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	TenantID      string
	UserRole      string // 用于权限检查
	RoleCode      string
	ResourceType  string
	PermissionType string // "read", "create", "update", "delete"
	Scope         string  // "all", "assigned_only"
	BranchOnly    bool
}

// CreatePermissionResponse 创建权限响应
type CreatePermissionResponse struct {
	PermissionID string `json:"permission_id"`
}

// CreatePermission 创建权限（只有 System tenant 的 SystemAdmin 可以）
func (s *RolePermissionService) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*CreatePermissionResponse, error) {
	// 权限检查
	if err := s.checkSystemAdminPermission(req.TenantID, req.UserRole); err != nil {
		return nil, err
	}

	// 参数验证
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	req.ResourceType = strings.TrimSpace(req.ResourceType)
	req.PermissionType = strings.TrimSpace(req.PermissionType)
	if req.RoleCode == "" || req.ResourceType == "" || req.PermissionType == "" {
		return nil, fmt.Errorf("role_code, resource_type, permission_type are required")
	}

	// 转换权限类型
	permTypeDB := s.permissionTypeToDB(req.PermissionType)
	if permTypeDB == "" {
		return nil, fmt.Errorf("invalid permission_type: %s", req.PermissionType)
	}

	// 构建权限领域模型
	assignedOnly := strings.TrimSpace(req.Scope) == "assigned_only"
	permission := &domain.RolePermission{
		RoleCode:       req.RoleCode,
		ResourceType:   req.ResourceType,
		PermissionType: permTypeDB,
		AssignedOnly:   assignedOnly,
		BranchOnly:     req.BranchOnly,
	}
	if req.TenantID != SystemTenantID {
		permission.TenantID = sql.NullString{String: req.TenantID, Valid: true}
	}

	// 调用 Repository（使用 UPSERT 语义）
	permissionID, err := s.permRepo.CreatePermission(ctx, SystemTenantID, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return &CreatePermissionResponse{
		PermissionID: permissionID,
	}, nil
}

// BatchCreatePermissionsRequest 批量创建权限请求
type BatchCreatePermissionsRequest struct {
	TenantID   string
	UserRole   string // 用于权限检查
	RoleCode   string
	Permissions []BatchPermissionItem
}

// BatchPermissionItem 批量权限项
type BatchPermissionItem struct {
	ResourceType   string `json:"resource_type"`
	PermissionType string `json:"permission_type"` // "read", "create", "update", "delete", "manage"
	Scope          string `json:"scope"`            // "all", "assigned_only"
	BranchOnly     bool   `json:"branch_only"`
	IsActive       bool   `json:"is_active"`
}

// BatchCreatePermissionsResponse 批量创建权限响应
type BatchCreatePermissionsResponse struct {
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}

// BatchCreatePermissions 批量创建权限（替换角色的所有权限）
func (s *RolePermissionService) BatchCreatePermissions(ctx context.Context, req BatchCreatePermissionsRequest) (*BatchCreatePermissionsResponse, error) {
	// 权限检查
	if err := s.checkSystemAdminPermission(req.TenantID, req.UserRole); err != nil {
		return nil, err
	}

	// 参数验证
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	if req.RoleCode == "" {
		return nil, fmt.Errorf("role_code is required")
	}

	// 删除该角色的所有现有权限
	if err := s.permRepo.DeletePermissionsByRole(ctx, SystemTenantID, req.RoleCode); err != nil {
		return nil, fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// 构建权限列表
	permissions := make([]*domain.RolePermission, 0)
	for _, item := range req.Permissions {
		// 跳过非激活的权限
		if !item.IsActive {
			continue
		}

		// 处理 "manage" 类型（展开为 R, C, U, D）
		permTypes := s.expandPermissionType(item.PermissionType)
		if len(permTypes) == 0 {
			continue
		}

		assignedOnly := strings.TrimSpace(item.Scope) == "assigned_only"
		for _, permType := range permTypes {
			permission := &domain.RolePermission{
				RoleCode:       req.RoleCode,
				ResourceType:   strings.TrimSpace(item.ResourceType),
				PermissionType: permType,
				AssignedOnly:   assignedOnly,
				BranchOnly:     item.BranchOnly,
			}
			permissions = append(permissions, permission)
		}
	}

	// 批量创建权限
	successCount, errors, err := s.permRepo.BatchCreatePermissions(ctx, SystemTenantID, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to batch create permissions: %w", err)
	}

	failedCount := len(errors)
	return &BatchCreatePermissionsResponse{
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	PermissionID string
	TenantID     string
	UserRole     string // 用于权限检查
	Scope        *string // "all", "assigned_only"
	BranchOnly   *bool
}

// UpdatePermission 更新权限
func (s *RolePermissionService) UpdatePermission(ctx context.Context, req UpdatePermissionRequest) error {
	// 权限检查
	if err := s.checkSystemAdminPermission(req.TenantID, req.UserRole); err != nil {
		return err
	}

	// 参数验证
	if req.PermissionID == "" {
		return fmt.Errorf("permission_id is required")
	}

	// 获取当前权限
	permission, err := s.permRepo.GetPermission(ctx, req.PermissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// 更新字段
	if req.Scope != nil {
		permission.AssignedOnly = strings.TrimSpace(*req.Scope) == "assigned_only"
	}
	if req.BranchOnly != nil {
		permission.BranchOnly = *req.BranchOnly
	}

	return s.permRepo.UpdatePermission(ctx, req.PermissionID, permission)
}

// DeletePermissionRequest 删除权限请求
type DeletePermissionRequest struct {
	PermissionID string
	TenantID     string
	UserRole     string // 用于权限检查
}

// DeletePermission 删除权限
func (s *RolePermissionService) DeletePermission(ctx context.Context, req DeletePermissionRequest) error {
	// 权限检查
	if err := s.checkSystemAdminPermission(req.TenantID, req.UserRole); err != nil {
		return err
	}

	// 参数验证
	if req.PermissionID == "" {
		return fmt.Errorf("permission_id is required")
	}

	return s.permRepo.DeletePermission(ctx, req.PermissionID)
}

// GetResourceTypesResponse 获取资源类型列表响应
type GetResourceTypesResponse struct {
	ResourceTypes []string `json:"resource_types"`
}

// GetResourceTypes 获取资源类型列表
func (s *RolePermissionService) GetResourceTypes(ctx context.Context) (*GetResourceTypesResponse, error) {
	// 查询所有权限，提取唯一的资源类型
	permissions, _, err := s.permRepo.ListPermissions(ctx, nil, repository.RolePermissionsFilter{}, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	// 提取唯一的资源类型
	resourceTypeMap := make(map[string]bool)
	for _, perm := range permissions {
		if perm.ResourceType != "" {
			resourceTypeMap[perm.ResourceType] = true
		}
	}

	resourceTypes := make([]string, 0, len(resourceTypeMap))
	for rt := range resourceTypeMap {
		resourceTypes = append(resourceTypes, rt)
	}

	return &GetResourceTypesResponse{
		ResourceTypes: resourceTypes,
	}, nil
}

// checkSystemAdminPermission 检查是否为 System tenant 的 SystemAdmin
func (s *RolePermissionService) checkSystemAdminPermission(tenantID, userRole string) error {
	if tenantID != SystemTenantID {
		return fmt.Errorf("only System tenant's SystemAdmin can modify role permissions")
	}
	if !strings.EqualFold(userRole, "SystemAdmin") {
		return fmt.Errorf("only System tenant's SystemAdmin can modify role permissions")
	}
	return nil
}

// permissionTypeToDB 将前端权限类型转换为数据库格式
func (s *RolePermissionService) permissionTypeToDB(permType string) string {
	m := map[string]string{
		"read":   "R",
		"create": "C",
		"update": "U",
		"delete": "D",
	}
	return m[strings.ToLower(permType)]
}

// permissionTypeFromDB 将数据库权限类型转换为前端格式
func (s *RolePermissionService) permissionTypeFromDB(permType string) string {
	m := map[string]string{
		"R": "read",
		"C": "create",
		"U": "update",
		"D": "delete",
	}
	return m[permType]
}

// expandPermissionType 展开权限类型（"manage" -> ["R", "C", "U", "D"]）
func (s *RolePermissionService) expandPermissionType(permType string) []string {
	switch strings.ToLower(permType) {
	case "manage":
		return []string{"R", "C", "U", "D"}
	case "read":
		return []string{"R"}
	case "create":
		return []string{"C"}
	case "update":
		return []string{"U"}
	case "delete":
		return []string{"D"}
	default:
		return []string{}
	}
}

// permissionToItem 将领域模型转换为前端格式
func (s *RolePermissionService) permissionToItem(perm *domain.RolePermission) PermissionItem {
	item := PermissionItem{
		PermissionID:   perm.PermissionID,
		RoleCode:       perm.RoleCode,
		ResourceType:   perm.ResourceType,
		PermissionType: s.permissionTypeFromDB(perm.PermissionType),
		BranchOnly:     perm.BranchOnly,
		IsActive:       true, // 存在即表示激活
	}

	if perm.TenantID.Valid {
		item.TenantID = &perm.TenantID.String
	}

	if perm.AssignedOnly {
		item.Scope = "assigned_only"
	} else {
		item.Scope = "all"
	}

	return item
}

