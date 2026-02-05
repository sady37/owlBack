package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
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
	TenantID       *string
	RoleCode       string
	ResourceType   string
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
	PermissionID    string  `json:"permission_id"`
	TenantID        *string `json:"tenant_id"`
	RoleCode        string  `json:"role_code"`
	ResourceType    string  `json:"resource_type"`
	PermissionType  string  `json:"permission_type"`  // "read", "create", "update", "delete"
	PermissionScope string  `json:"permission_scope"` // "A" (all), "S" (assigned_only), "B" (branch_only)
	IsActive        bool    `json:"is_active"`
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

	// 系统角色（SystemAdmin、SystemOperator）的权限在系统租户下，前端用用户 tenant_id 请求时需查系统租户
	queryTenantID := req.TenantID
	if req.TenantID != nil && *req.TenantID != "" && (filter.RoleCode == "SystemAdmin" || filter.RoleCode == "SystemOperator") {
		sysT := SystemTenantID
		queryTenantID = &sysT
	}

	// 查询权限列表
	permissions, total, err := s.permRepo.ListPermissions(ctx, queryTenantID, filter, req.Page, req.Size)
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
	TenantID        string
	UserRole        string // 用于权限检查
	RoleCode        string
	ResourceType    string
	PermissionType  string // "read", "create", "update", "delete"
	PermissionScope string // "A" (all), "S" (assigned_only), "B" (branch_only)
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
	// 验证并设置 permission_scope
	permissionScope := strings.TrimSpace(req.PermissionScope)
	if permissionScope == "" {
		permissionScope = "A" // 默认：all
	}
	if permissionScope != "A" && permissionScope != "S" && permissionScope != "B" {
		return nil, fmt.Errorf("invalid permission_scope: %s (must be A, S, or B)", permissionScope)
	}
	permission := &domain.RolePermission{
		RoleCode:        req.RoleCode,
		ResourceType:    req.ResourceType,
		PermissionType:  permTypeDB,
		PermissionScope: permissionScope,
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
	TenantID    string
	UserRole    string // 用于权限检查
	RoleCode    string
	Permissions []BatchPermissionItem
}

// BatchPermissionItem 批量权限项
type BatchPermissionItem struct {
	ResourceType    string `json:"resource_type"`
	PermissionType  string `json:"permission_type"`  // "read", "create", "update", "delete", "manage"
	PermissionScope string `json:"permission_scope"` // "A" (all), "S" (assigned_only), "B" (branch_only)
	IsActive        bool   `json:"is_active"`
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

	// 构建权限列表（局部更新：只更新前端发送的权限）
	// 注意：不再删除所有权限，而是使用 upsert（ON CONFLICT）来更新/创建权限
	// 这样其他权限会保持不变，只更新前端发送的变化部分
	//
	// 权限删除处理：
	// - 如果前端发送 is_active=false，需要删除该权限
	// - 如果前端不发送某个权限（未在列表中），该权限保持不变
	permissionsToUpdate := make([]*domain.RolePermission, 0)
	permissionsToDelete := make([]*domain.RolePermission, 0)

	for _, item := range req.Permissions {
		// 处理 "manage" 类型（展开为 R, C, U, D）
		permTypes := s.expandPermissionType(item.PermissionType)
		if len(permTypes) == 0 {
			continue
		}

		permissionScope := strings.TrimSpace(item.PermissionScope)
		if permissionScope == "" {
			permissionScope = "A" // 默认：all
		}
		if permissionScope != "A" && permissionScope != "S" && permissionScope != "B" {
			continue // 跳过无效的 permission_scope
		}

		for _, permType := range permTypes {
			permission := &domain.RolePermission{
				RoleCode:        req.RoleCode,
				ResourceType:    strings.TrimSpace(item.ResourceType),
				PermissionType:  permType,
				PermissionScope: permissionScope,
			}

			if !item.IsActive {
				// is_active=false 表示要删除该权限
				permissionsToDelete = append(permissionsToDelete, permission)
			} else {
				// is_active=true 表示要更新/创建该权限
				permissionsToUpdate = append(permissionsToUpdate, permission)
			}
		}
	}

	// 删除需要删除的权限
	systemTenantID := SystemTenantID
	tenantIDPtr := &systemTenantID
	for _, perm := range permissionsToDelete {
		existingPerm, err := s.permRepo.GetPermissionByKey(ctx, tenantIDPtr, perm.RoleCode, perm.ResourceType, perm.PermissionType)
		if err == nil && existingPerm != nil {
			// 权限存在，删除它
			if err := s.permRepo.DeletePermission(ctx, existingPerm.PermissionID); err != nil {
				// 记录错误但继续处理其他权限
				// 可以考虑收集错误并返回
			}
		}
		// 如果权限不存在，忽略（已经是想要的状态）
	}

	permissions := permissionsToUpdate

	// 批量更新/创建权限（使用 upsert：ON CONFLICT DO UPDATE）
	// 只更新前端发送的权限，其他权限保持不变（局部更新）
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
	PermissionID    string
	TenantID        string
	UserRole        string  // 用于权限检查
	PermissionScope *string // "A" (all), "S" (assigned_only), "B" (branch_only)
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
	if req.PermissionScope != nil {
		permission.PermissionScope = strings.TrimSpace(*req.PermissionScope)
		// 验证 permission_scope 值
		if permission.PermissionScope != "A" && permission.PermissionScope != "S" && permission.PermissionScope != "B" {
			return fmt.Errorf("invalid permission_scope: %s (must be A, S, or B)", permission.PermissionScope)
		}
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

	// 提取唯一的资源类型，过滤掉 service_levels
	resourceTypeMap := make(map[string]bool)
	for _, perm := range permissions {
		if perm.ResourceType != "" && perm.ResourceType != "service_levels" {
			resourceTypeMap[perm.ResourceType] = true
		}
	}

	// 确保 branches 存在（如果数据库中有相关权限记录）
	// 注意：即使数据库中没有 branches 权限记录，也强制添加 branches 到列表中
	if !resourceTypeMap["branches"] {
		resourceTypeMap["branches"] = true
	}

	// 定义资源类型分类和排序顺序
	// 分类：系统管理 -> 空间管理 -> 住户管理 -> 设备管理 -> 告警管理 -> 其他
	resourceOrder := []string{
		// 系统管理
		"tenants",
		"roles",
		"role_permissions",
		"users",
		"branches",
		// 空间管理
		"units",
		"rooms",
		"beds",
		// 住户管理
		"residents",
		"resident_phi",
		"resident_contacts",
		"resident_caregivers",
		// 设备管理
		"devices",
		"device_store",
		"iot_timeseries",
		// 告警管理
		"alarm_events",
		"alarm_device",
		"alarm_cloud",
		// 其他
		"config_versions",
		"cards",
		"rounds",
		"round_details",
	}

	// 按顺序构建输出列表
	resourceTypes := []string{}
	added := make(map[string]bool)
	// 先添加有序的资源类型
	for _, rt := range resourceOrder {
		if resourceTypeMap[rt] && !added[rt] {
			resourceTypes = append(resourceTypes, rt)
			added[rt] = true
		}
	}
	// 添加其他未在排序列表中的资源类型（按字母顺序）
	others := []string{}
	for rt := range resourceTypeMap {
		if !added[rt] {
			others = append(others, rt)
		}
	}
	// 对 others 按字母顺序排序
	sort.Strings(others)
	resourceTypes = append(resourceTypes, others...)

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
		PermissionID:    perm.PermissionID,
		RoleCode:        perm.RoleCode,
		ResourceType:    perm.ResourceType,
		PermissionType:  s.permissionTypeFromDB(perm.PermissionType),
		PermissionScope: perm.PermissionScope,
		IsActive:        true, // 存在即表示激活
	}

	if perm.TenantID.Valid {
		item.TenantID = &perm.TenantID.String
	}

	return item
}

// convertToPermissionScope 将前端的 scope 和 branch_only 转换为 permission_scope
// scope: "all" 或 "assigned_only"
// branchOnly: true 或 false
// 返回: "A" (all), "S" (assigned_only), "B" (branch_only)
func (s *RolePermissionService) convertToPermissionScope(scope string, branchOnly bool) string {
	if branchOnly {
		return "B" // branch_only 优先级更高
	}
	if strings.TrimSpace(scope) == "assigned_only" {
		return "S"
	}
	return "A" // 默认：all
}
