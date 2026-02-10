//go:build integration
// +build integration

package service

import (
	"context"
	"testing"

	"wisefido-data/internal/repository"
)

func TestRolePermissionService_ListPermissions(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 测试查询权限列表
	req := ListPermissionsRequest{
		TenantID: func() *string { s := SystemTenantID; return &s }(),
		Page:     1,
		Size:     100,
	}

	resp, err := permService.ListPermissions(ctx, req)
	if err != nil {
		t.Fatalf("ListPermissions failed: %v", err)
	}

	if resp == nil {
		t.Fatal("ListPermissions returned nil response")
	}

	t.Logf("ListPermissions success: total=%d, items=%d", resp.Total, len(resp.Items))
}

func TestRolePermissionService_CreatePermission(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 测试创建权限
	req := CreatePermissionRequest{
		TenantID:       SystemTenantID,
		UserRole:       "SystemAdmin",
		RoleCode:       "TestRole",
		ResourceType:   "test_resource",
		PermissionType: "read",
		Scope:          "all",
		BranchOnly:     false,
	}

	resp, err := permService.CreatePermission(ctx, req)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}

	if resp == nil || resp.PermissionID == "" {
		t.Fatal("CreatePermission returned invalid response")
	}

	t.Logf("CreatePermission success: permission_id=%s", resp.PermissionID)

	// 清理：删除测试权限
	_ = permRepo.DeletePermission(ctx, resp.PermissionID)
}

func TestRolePermissionService_BatchCreatePermissions(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 测试批量创建权限
	req := BatchCreatePermissionsRequest{
		TenantID: SystemTenantID,
		UserRole: "SystemAdmin",
		RoleCode: "TestRole",
		Permissions: []BatchPermissionItem{
			{
				ResourceType:   "test_resource",
				PermissionType: "manage",
				Scope:          "all",
				BranchOnly:     false,
				IsActive:       true,
			},
			{
				ResourceType:   "test_resource2",
				PermissionType: "read",
				Scope:          "assigned_only",
				BranchOnly:     true,
				IsActive:       true,
			},
		},
	}

	resp, err := permService.BatchCreatePermissions(ctx, req)
	if err != nil {
		t.Fatalf("BatchCreatePermissions failed: %v", err)
	}

	if resp == nil {
		t.Fatal("BatchCreatePermissions returned nil response")
	}

	t.Logf("BatchCreatePermissions success: success_count=%d, failed_count=%d", resp.SuccessCount, resp.FailedCount)

	// 清理：删除测试权限
	_ = permRepo.DeletePermissionsByRole(ctx, SystemTenantID, req.RoleCode)
}

func TestRolePermissionService_UpdatePermission(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 先创建一个测试权限
	createReq := CreatePermissionRequest{
		TenantID:       SystemTenantID,
		UserRole:       "SystemAdmin",
		RoleCode:       "TestRole",
		ResourceType:   "test_resource",
		PermissionType: "read",
		Scope:          "all",
		BranchOnly:     false,
	}
	createResp, err := permService.CreatePermission(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	defer permRepo.DeletePermission(ctx, createResp.PermissionID)

	// 测试更新权限
	updateReq := UpdatePermissionRequest{
		PermissionID: createResp.PermissionID,
		TenantID:     SystemTenantID,
		UserRole:     "SystemAdmin",
		Scope:        func() *string { s := "assigned_only"; return &s }(),
		BranchOnly:   func() *bool { b := true; return &b }(),
	}

	err = permService.UpdatePermission(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdatePermission failed: %v", err)
	}

	t.Logf("UpdatePermission success: permission_id=%s", createResp.PermissionID)
}

func TestRolePermissionService_DeletePermission(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 先创建一个测试权限
	createReq := CreatePermissionRequest{
		TenantID:       SystemTenantID,
		UserRole:       "SystemAdmin",
		RoleCode:       "TestRole",
		ResourceType:   "test_resource",
		PermissionType: "read",
		Scope:          "all",
		BranchOnly:     false,
	}
	createResp, err := permService.CreatePermission(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}

	// 测试删除权限
	deleteReq := DeletePermissionRequest{
		PermissionID: createResp.PermissionID,
		TenantID:     SystemTenantID,
		UserRole:     "SystemAdmin",
	}

	err = permService.DeletePermission(ctx, deleteReq)
	if err != nil {
		t.Fatalf("DeletePermission failed: %v", err)
	}

	t.Logf("DeletePermission success: permission_id=%s", createResp.PermissionID)
}

func TestRolePermissionService_GetResourceTypes(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 测试获取资源类型列表
	resp, err := permService.GetResourceTypes(ctx)
	if err != nil {
		t.Fatalf("GetResourceTypes failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetResourceTypes returned nil response")
	}

	t.Logf("GetResourceTypes success: resource_types=%v", resp.ResourceTypes)
}

func TestRolePermissionService_PermissionCheck(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	permRepo := repository.NewPostgresRolePermissionsRepository(db)
	permService := NewRolePermissionService(permRepo, getTestLogger())

	// 测试权限检查（非 SystemAdmin 应该失败）
	req := CreatePermissionRequest{
		TenantID:       SystemTenantID,
		UserRole:       "Admin", // 不是 SystemAdmin
		RoleCode:       "TestRole",
		ResourceType:   "test_resource",
		PermissionType: "read",
		Scope:          "all",
		BranchOnly:     false,
	}

	_, err := permService.CreatePermission(ctx, req)
	if err == nil {
		t.Fatal("CreatePermission should fail for non-SystemAdmin")
	}

	t.Logf("PermissionCheck test success: error=%v", err)
}
