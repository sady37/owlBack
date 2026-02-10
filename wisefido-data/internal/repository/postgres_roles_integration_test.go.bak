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
func getTestDBForRoles(t *testing.T) *sql.DB {
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
func createTestTenantForRoles(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Roles", "test-roles.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForRoles(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：roles -> tenants
	db.Exec(`DELETE FROM roles WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// RolesRepository 测试
// ============================================

func TestPostgresRolesRepository_GetRole(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()

	// 查询系统角色（SystemAdmin应该存在）
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	var roleID string
	err := db.QueryRowContext(ctx, `
		SELECT role_id::text
		FROM roles
		WHERE tenant_id = $1 AND role_code = 'SystemAdmin'
		LIMIT 1
	`, systemTenantID).Scan(&roleID)
	if err != nil {
		t.Skipf("SystemAdmin role not found, skipping test")
		return
	}

	// 测试：查询角色
	role, err := repo.GetRole(ctx, roleID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if role.RoleID != roleID {
		t.Errorf("Expected role_id=%s, got %s", roleID, role.RoleID)
	}
	if role.RoleCode != "SystemAdmin" {
		t.Errorf("Expected role_code=SystemAdmin, got %s", role.RoleCode)
	}
	if !role.IsSystem {
		t.Errorf("Expected is_system=true, got %v", role.IsSystem)
	}
}

func TestPostgresRolesRepository_GetRoleByCode(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()

	// 测试：通过role_code查询系统角色（不指定tenantID）
	role, err := repo.GetRoleByCode(ctx, nil, "Admin")
	if err != nil {
		t.Fatalf("GetRoleByCode failed: %v", err)
	}
	if role.RoleCode != "Admin" {
		t.Errorf("Expected role_code=Admin, got %s", role.RoleCode)
	}
	if !role.IsSystem {
		t.Errorf("Expected is_system=true, got %v", role.IsSystem)
	}

	// 测试：通过role_code查询系统角色（指定System tenantID）
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	role2, err := repo.GetRoleByCode(ctx, &systemTenantID, "Nurse")
	if err != nil {
		t.Fatalf("GetRoleByCode failed: %v", err)
	}
	if role2.RoleCode != "Nurse" {
		t.Errorf("Expected role_code=Nurse, got %s", role2.RoleCode)
	}
}

func TestPostgresRolesRepository_ListRoles(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()

	// 测试：查询系统角色列表（默认）
	roles, total, err := repo.ListRoles(ctx, nil, RolesFilter{}, 1, 100)
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}
	if total == 0 {
		t.Fatal("Expected at least one system role, got 0")
	}
	if len(roles) == 0 {
		t.Fatal("Expected at least one system role in list, got 0")
	}
	t.Logf("Found %d system roles (total: %d)", len(roles), total)

	// 验证：所有角色都是系统角色
	for _, role := range roles {
		if !role.IsSystem {
			t.Errorf("Expected all roles to be system roles, but found is_system=false: role_code=%s", role.RoleCode)
		}
	}

	// 测试：按is_active过滤
	isActive := true
	roles2, total2, err := repo.ListRoles(ctx, nil, RolesFilter{IsActive: &isActive}, 1, 100)
	if err != nil {
		t.Fatalf("ListRoles with is_active filter failed: %v", err)
	}
	t.Logf("Found %d active roles (total: %d)", len(roles2), total2)

	// 测试：搜索过滤
	roles3, total3, err := repo.ListRoles(ctx, nil, RolesFilter{Search: "Admin"}, 1, 100)
	if err != nil {
		t.Fatalf("ListRoles with search filter failed: %v", err)
	}
	t.Logf("Found %d roles matching 'Admin' (total: %d)", len(roles3), total3)
}

func TestPostgresRolesRepository_CreateRole(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRoles(t, db)
	defer cleanupTestDataForRoles(t, db, tenantID)

	// 测试：创建租户自定义角色
	role := &domain.Role{
		RoleCode:    "TestRole",
		Description: "Test Role\nThis is a test role for integration testing",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}

	roleID, err := repo.CreateRole(ctx, tenantID, role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	if roleID == "" {
		t.Fatal("CreateRole returned empty role_id")
	}
	t.Logf("Created role_id: %s", roleID)
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 验证：查询创建的角色
	createdRole, err := repo.GetRole(ctx, roleID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if createdRole.RoleCode != "TestRole" {
		t.Errorf("Expected role_code=TestRole, got %s", createdRole.RoleCode)
	}
	if createdRole.IsSystem {
		t.Errorf("Expected is_system=false, got %v", createdRole.IsSystem)
	}
	if !createdRole.TenantID.Valid || createdRole.TenantID.String != tenantID {
		t.Errorf("Expected tenant_id=%s, got %v", tenantID, createdRole.TenantID)
	}
}

func TestPostgresRolesRepository_CreateRole_DuplicateRoleCode(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRoles(t, db)
	defer cleanupTestDataForRoles(t, db, tenantID)

	// 创建第一个角色
	role1 := &domain.Role{
		RoleCode:    "DuplicateRole",
		Description: "First Role",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID1, err := repo.CreateRole(ctx, tenantID, role1)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID1)

	// 测试：尝试创建重复的role_code（应该失败）
	role2 := &domain.Role{
		RoleCode:    "DuplicateRole",
		Description: "Second Role",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	_, err = repo.CreateRole(ctx, tenantID, role2)
	if err == nil {
		t.Fatal("CreateRole should fail for duplicate role_code, but succeeded")
	}
	t.Logf("Expected error for duplicate role_code: %v", err)
}

func TestPostgresRolesRepository_UpdateRole(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRoles(t, db)
	defer cleanupTestDataForRoles(t, db, tenantID)

	// 先创建一个角色
	role := &domain.Role{
		RoleCode:    "UpdateTestRole",
		Description: "Original Description\nOriginal detailed description",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := repo.CreateRole(ctx, tenantID, role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 测试：更新角色（部分更新）
	updateRole := &domain.Role{
		Description: "Updated Description\nUpdated detailed description",
		IsActive:    sql.NullBool{Bool: false, Valid: true},
	}
	err = repo.UpdateRole(ctx, roleID, updateRole)
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	// 验证：查询更新的角色
	updatedRole, err := repo.GetRole(ctx, roleID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if updatedRole.Description != "Updated Description\nUpdated detailed description" {
		t.Errorf("Expected updated description, got %s", updatedRole.Description)
	}
	if !updatedRole.IsActive.Valid || updatedRole.IsActive.Bool {
		t.Errorf("Expected is_active=false, got %v", updatedRole.IsActive)
	}
}

func TestPostgresRolesRepository_UpdateRole_CannotChangeIsSystem(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRoles(t, db)
	defer cleanupTestDataForRoles(t, db, tenantID)

	// 先创建一个系统角色（模拟）
	role := &domain.Role{
		RoleCode:    "SystemTestRole",
		Description: "System Role",
		IsSystem:    true,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := repo.CreateRole(ctx, tenantID, role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer db.Exec(`DELETE FROM roles WHERE role_id = $1`, roleID)

	// 测试：尝试改变is_system（应该失败）
	updateRole := &domain.Role{
		IsSystem: false,
	}
	err = repo.UpdateRole(ctx, roleID, updateRole)
	if err == nil {
		t.Fatal("UpdateRole should fail when changing is_system, but succeeded")
	}
	t.Logf("Expected error for changing is_system: %v", err)
}

func TestPostgresRolesRepository_DeleteRole(t *testing.T) {
	db := getTestDBForRoles(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresRolesRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForRoles(t, db)
	defer cleanupTestDataForRoles(t, db, tenantID)

	// 先创建一个角色
	role := &domain.Role{
		RoleCode:    "DeleteTestRole",
		Description: "Role to be deleted",
		IsSystem:    false,
		IsActive:    sql.NullBool{Bool: true, Valid: true},
	}
	roleID, err := repo.CreateRole(ctx, tenantID, role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// 测试：删除角色
	err = repo.DeleteRole(ctx, roleID)
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	// 验证：角色已删除
	_, err = repo.GetRole(ctx, roleID)
	if err == nil {
		t.Fatal("Role should be deleted, but GetRole succeeded")
	}
}

