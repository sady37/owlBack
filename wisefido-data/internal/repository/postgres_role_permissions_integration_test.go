// +build integration

package repository

import (
	"context"
	"database/sql"
	"testing"

	"owl-common/database"
	"owl-common/config"
	"wisefido-data/internal/domain"
)

// 获取测试数据库连接（复用postgres_users_integration_test.go中的函数）
func getTestDBForRolePermissions(t *testing.T) *sql.DB {
	cfg := &config.DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnvInt("TEST_DB_PORT", 5432),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		Database: getEnv("TEST_DB_NAME", "owlrd"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test: cannot ping database: %v", err)
		return nil
	}

	return db
}

// 创建测试租户
func createTestTenantForRolePermissions(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000996"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Role Permissions", "test-role-permissions.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForRolePermissions(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：role_permissions -> roles -> tenants
	db.Exec(`DELETE FROM role_permissions WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM roles WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// RolePermissionsRepository 测试
// ============================================

func TestPostgresRolePermissionsRepository_GetPermission(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()

	// 查询系统权限（SystemAdmin应该存在）
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	var permissionID string
	err := db.QueryRowContext(ctx, `
		SELECT permission_id::text
		FROM role_permissions
		WHERE tenant_id = $1 AND role_code = 'SystemAdmin' AND resource_type = 'tenants' AND permission_type = 'R'
		LIMIT 1
	`, systemTenantID).Scan(&permissionID)
	if err != nil {
		t.Skipf("SystemAdmin permission not found, skipping test")
		return
	}

	// 测试：查询权限
	permission, err := repo.GetPermission(ctx, permissionID)
	if err != nil {
		t.Fatalf("GetPermission failed: %v", err)
	}
	if permission.PermissionID != permissionID {
		t.Errorf("Expected permission_id=%s, got %s", permissionID, permission.PermissionID)
	}
	if permission.RoleCode != "SystemAdmin" {
		t.Errorf("Expected role_code=SystemAdmin, got %s", permission.RoleCode)
	}
	if permission.ResourceType != "tenants" {
		t.Errorf("Expected resource_type=tenants, got %s", permission.ResourceType)
	}
	if permission.PermissionType != "R" {
		t.Errorf("Expected permission_type=R, got %s", permission.PermissionType)
	}
}

func TestPostgresRolePermissionsRepository_GetPermissionByKey(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()

	// 测试：通过key查询系统权限（不指定tenantID）
	permission, err := repo.GetPermissionByKey(ctx, nil, "Admin", "users", "R")
	if err != nil {
		t.Fatalf("GetPermissionByKey failed: %v", err)
	}
	if permission.RoleCode != "Admin" {
		t.Errorf("Expected role_code=Admin, got %s", permission.RoleCode)
	}
	if permission.ResourceType != "users" {
		t.Errorf("Expected resource_type=users, got %s", permission.ResourceType)
	}
	if permission.PermissionType != "R" {
		t.Errorf("Expected permission_type=R, got %s", permission.PermissionType)
	}
}

func TestPostgresRolePermissionsRepository_ListPermissions(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()

	// 测试：查询系统权限列表（默认）
	permissions, total, err := repo.ListPermissions(ctx, nil, RolePermissionsFilter{}, 1, 100)
	if err != nil {
		t.Fatalf("ListPermissions failed: %v", err)
	}
	if total == 0 {
		t.Fatal("Expected at least one system permission, got 0")
	}
	if len(permissions) == 0 {
		t.Fatal("Expected at least one system permission in list, got 0")
	}
	t.Logf("Found %d system permissions (total: %d)", len(permissions), total)

	// 测试：按role_code过滤
	permissions2, total2, err := repo.ListPermissions(ctx, nil, RolePermissionsFilter{RoleCode: "Admin"}, 1, 100)
	if err != nil {
		t.Fatalf("ListPermissions with role_code filter failed: %v", err)
	}
	t.Logf("Found %d Admin permissions (total: %d)", len(permissions2), total2)

	// 测试：按assigned_only过滤
	assignedOnly := true
	permissions3, total3, err := repo.ListPermissions(ctx, nil, RolePermissionsFilter{AssignedOnly: &assignedOnly}, 1, 100)
	if err != nil {
		t.Fatalf("ListPermissions with assigned_only filter failed: %v", err)
	}
	t.Logf("Found %d permissions with assigned_only=true (total: %d)", len(permissions3), total3)
}

func TestPostgresRolePermissionsRepository_GetPermissionsByRole(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()

	// 测试：查询Admin角色的所有权限
	permissions, err := repo.GetPermissionsByRole(ctx, nil, "Admin")
	if err != nil {
		t.Fatalf("GetPermissionsByRole failed: %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("Expected at least one Admin permission, got 0")
	}
	t.Logf("Found %d permissions for Admin role", len(permissions))

	// 验证：所有权限都是Admin角色的
	for _, perm := range permissions {
		if perm.RoleCode != "Admin" {
			t.Errorf("Expected all permissions to be for Admin role, but found role_code=%s", perm.RoleCode)
		}
	}
}

func TestPostgresRolePermissionsRepository_GetPermissionsByResource(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()

	// 测试：查询users资源的所有权限
	permissions, err := repo.GetPermissionsByResource(ctx, nil, "users")
	if err != nil {
		t.Fatalf("GetPermissionsByResource failed: %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("Expected at least one permission for users resource, got 0")
	}
	t.Logf("Found %d permissions for users resource", len(permissions))

	// 验证：所有权限都是users资源的
	for _, perm := range permissions {
		if perm.ResourceType != "users" {
			t.Errorf("Expected all permissions to be for users resource, but found resource_type=%s", perm.ResourceType)
		}
	}
}

func TestPostgresRolePermissionsRepository_CreatePermission(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForPerm",
		Description: "Test Role\nFor permission testing",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 测试：创建权限
	permission := &domain.RolePermission{
		RoleCode:       "TestRoleForPerm",
		ResourceType:   "residents",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}

	permissionID, err := repo.CreatePermission(ctx, tenantID, permission)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	if permissionID == "" {
		t.Fatal("CreatePermission returned empty permission_id")
	}
	t.Logf("Created permission_id: %s", permissionID)
	defer db.Exec(`DELETE FROM role_permissions WHERE permission_id = $1`, permissionID)

	// 验证：查询创建的权限
	createdPerm, err := repo.GetPermission(ctx, permissionID)
	if err != nil {
		t.Fatalf("GetPermission failed: %v", err)
	}
	if createdPerm.RoleCode != "TestRoleForPerm" {
		t.Errorf("Expected role_code=TestRoleForPerm, got %s", createdPerm.RoleCode)
	}
	if createdPerm.ResourceType != "residents" {
		t.Errorf("Expected resource_type=residents, got %s", createdPerm.ResourceType)
	}
	if createdPerm.PermissionType != "R" {
		t.Errorf("Expected permission_type=R, got %s", createdPerm.PermissionType)
	}
}

func TestPostgresRolePermissionsRepository_CreatePermission_DuplicateKey(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForPerm2",
		Description: "Test Role 2",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 创建第一个权限
	permission1 := &domain.RolePermission{
		RoleCode:       "TestRoleForPerm2",
		ResourceType:   "devices",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}
	permissionID1, err := repo.CreatePermission(ctx, tenantID, permission1)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	defer db.Exec(`DELETE FROM role_permissions WHERE permission_id = $1`, permissionID1)

	// 测试：尝试创建重复的权限（应该失败）
	permission2 := &domain.RolePermission{
		RoleCode:       "TestRoleForPerm2",
		ResourceType:   "devices",
		PermissionType: "R",
		AssignedOnly:   true, // 不同的assigned_only，但key相同
		BranchOnly:     false,
	}
	_, err = repo.CreatePermission(ctx, tenantID, permission2)
	if err == nil {
		t.Fatal("CreatePermission should fail for duplicate key, but succeeded")
	}
	t.Logf("Expected error for duplicate key: %v", err)
}

func TestPostgresRolePermissionsRepository_BatchCreatePermissions(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForBatch",
		Description: "Test Role For Batch",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 测试：批量创建权限
	permissions := []*domain.RolePermission{
		{
			RoleCode:       "TestRoleForBatch",
			ResourceType:   "residents",
			PermissionType: "R",
			AssignedOnly:   false,
			BranchOnly:     false,
		},
		{
			RoleCode:       "TestRoleForBatch",
			ResourceType:   "residents",
			PermissionType: "C",
			AssignedOnly:   false,
			BranchOnly:     false,
		},
		{
			RoleCode:       "TestRoleForBatch",
			ResourceType:   "devices",
			PermissionType: "R",
			AssignedOnly:   true,
			BranchOnly:     false,
		},
	}

	successCount, errors, err := repo.BatchCreatePermissions(ctx, tenantID, permissions)
	if err != nil {
		t.Fatalf("BatchCreatePermissions failed: %v", err)
	}
	if successCount != 3 {
		t.Errorf("Expected 3 successful creations, got %d", successCount)
	}
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d errors: %v", len(errors), errors)
	}
	t.Logf("Batch created %d permissions", successCount)
	defer db.Exec(`DELETE FROM role_permissions WHERE role_code = 'TestRoleForBatch' AND tenant_id = $1`, tenantID)
}

func TestPostgresRolePermissionsRepository_UpdatePermission(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色和权限
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForUpdate",
		Description: "Test Role For Update",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	permission := &domain.RolePermission{
		RoleCode:       "TestRoleForUpdate",
		ResourceType:   "residents",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}
	permissionID, err := repo.CreatePermission(ctx, tenantID, permission)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	defer db.Exec(`DELETE FROM role_permissions WHERE permission_id = $1`, permissionID)

	// 测试：更新权限（部分更新）
	updatePermission := &domain.RolePermission{
		AssignedOnly: true,
		BranchOnly:   true,
	}
	err = repo.UpdatePermission(ctx, permissionID, updatePermission)
	if err != nil {
		t.Fatalf("UpdatePermission failed: %v", err)
	}

	// 验证：查询更新的权限
	updatedPerm, err := repo.GetPermission(ctx, permissionID)
	if err != nil {
		t.Fatalf("GetPermission failed: %v", err)
	}
	if !updatedPerm.AssignedOnly {
		t.Errorf("Expected assigned_only=true, got %v", updatedPerm.AssignedOnly)
	}
	if !updatedPerm.BranchOnly {
		t.Errorf("Expected branch_only=true, got %v", updatedPerm.BranchOnly)
	}
}

func TestPostgresRolePermissionsRepository_BatchUpdatePermissions(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色和权限
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForBatchUpdate",
		Description: "Test Role For Batch Update",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 创建两个权限
	perm1 := &domain.RolePermission{
		RoleCode:       "TestRoleForBatchUpdate",
		ResourceType:   "residents",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}
	permissionID1, err := repo.CreatePermission(ctx, tenantID, perm1)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	defer db.Exec(`DELETE FROM role_permissions WHERE permission_id = $1`, permissionID1)

	perm2 := &domain.RolePermission{
		RoleCode:       "TestRoleForBatchUpdate",
		ResourceType:   "devices",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}
	permissionID2, err := repo.CreatePermission(ctx, tenantID, perm2)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	defer db.Exec(`DELETE FROM role_permissions WHERE permission_id = $1`, permissionID2)

	// 测试：批量更新权限
	updates := []PermissionUpdate{
		{
			PermissionID: permissionID1,
			Permission: &domain.RolePermission{
				AssignedOnly: true,
			},
		},
		{
			PermissionID: permissionID2,
			Permission: &domain.RolePermission{
				BranchOnly: true,
			},
		},
	}

	successCount, errors, err := repo.BatchUpdatePermissions(ctx, updates)
	if err != nil {
		t.Fatalf("BatchUpdatePermissions failed: %v", err)
	}
	if successCount != 2 {
		t.Errorf("Expected 2 successful updates, got %d", successCount)
	}
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d errors: %v", len(errors), errors)
	}

	// 验证：查询更新的权限
	updatedPerm1, _ := repo.GetPermission(ctx, permissionID1)
	if !updatedPerm1.AssignedOnly {
		t.Errorf("Expected assigned_only=true for permission1, got %v", updatedPerm1.AssignedOnly)
	}

	updatedPerm2, _ := repo.GetPermission(ctx, permissionID2)
	if !updatedPerm2.BranchOnly {
		t.Errorf("Expected branch_only=true for permission2, got %v", updatedPerm2.BranchOnly)
	}
}

func TestPostgresRolePermissionsRepository_DeletePermission(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色和权限
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForDelete",
		Description: "Test Role For Delete",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	permission := &domain.RolePermission{
		RoleCode:       "TestRoleForDelete",
		ResourceType:   "residents",
		PermissionType: "R",
		AssignedOnly:   false,
		BranchOnly:     false,
	}
	permissionID, err := repo.CreatePermission(ctx, tenantID, permission)
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}

	// 测试：删除权限
	err = repo.DeletePermission(ctx, permissionID)
	if err != nil {
		t.Fatalf("DeletePermission failed: %v", err)
	}

	// 验证：权限已删除
	_, err = repo.GetPermission(ctx, permissionID)
	if err == nil {
		t.Fatal("Permission should be deleted, but GetPermission succeeded")
	}
}

func TestPostgresRolePermissionsRepository_DeletePermissionsByRole(t *testing.T) {
	db := getTestDBForRolePermissions(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolePermissionsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRolePermissions(t, db)
	defer cleanupTestDataForRolePermissions(t, db, tenantID)

	// 先创建一个测试角色
	rolesRepo := NewPostgresRolesRepository(db)
	testRole := &domain.Role{
		RoleCode:    "TestRoleForDeleteAll",
		Description: "Test Role For Delete All",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := rolesRepo.CreateRole(ctx, tenantID, testRole)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 创建多个权限
	permissions := []*domain.RolePermission{
		{RoleCode: "TestRoleForDeleteAll", ResourceType: "residents", PermissionType: "R", AssignedOnly: false, BranchOnly: false},
		{RoleCode: "TestRoleForDeleteAll", ResourceType: "residents", PermissionType: "C", AssignedOnly: false, BranchOnly: false},
		{RoleCode: "TestRoleForDeleteAll", ResourceType: "devices", PermissionType: "R", AssignedOnly: false, BranchOnly: false},
	}

	for _, perm := range permissions {
		_, err := repo.CreatePermission(ctx, tenantID, perm)
		if err != nil {
			t.Fatalf("CreatePermission failed: %v", err)
		}
	}

	// 测试：删除角色的所有权限
	err = repo.DeletePermissionsByRole(ctx, tenantID, "TestRoleForDeleteAll")
	if err != nil {
		t.Fatalf("DeletePermissionsByRole failed: %v", err)
	}

	// 验证：所有权限已删除
	remainingPerms, err := repo.GetPermissionsByRole(ctx, &tenantID, "TestRoleForDeleteAll")
	if err != nil {
		t.Fatalf("GetPermissionsByRole failed: %v", err)
	}
	if len(remainingPerms) > 0 {
		t.Errorf("Expected all permissions to be deleted, but found %d remaining", len(remainingPerms))
	}
}

