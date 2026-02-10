// +build integration

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"owl-common/database"
	"owl-common/config"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// getTestDB 获取测试数据库连接
func getTestDBForService(t *testing.T) *sql.DB {
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

// getTestLogger 获取测试日志
func getTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

func TestRoleService_ListRoles(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 测试查询角色列表
	req := ListRolesRequest{
		TenantID: func() *string { s := SystemTenantID; return &s }(),
		Page:     1,
		Size:     20,
	}

	resp, err := roleService.ListRoles(ctx, req)
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}

	if resp == nil {
		t.Fatal("ListRoles returned nil response")
	}

	t.Logf("ListRoles success: total=%d, items=%d", resp.Total, len(resp.Items))
}

func TestRoleService_CreateRole(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 测试创建角色
	req := CreateRoleRequest{
		TenantID:    SystemTenantID,
		RoleCode:    "TestRole",
		DisplayName: "Test Role",
		Description: "Test role description",
	}

	resp, err := roleService.CreateRole(ctx, req)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if resp == nil || resp.RoleID == "" {
		t.Fatal("CreateRole returned invalid response")
	}

	t.Logf("CreateRole success: role_id=%s", resp.RoleID)

	// 清理：删除测试角色
	_ = roleRepo.DeleteRole(ctx, resp.RoleID)
}

func TestRoleService_UpdateRole(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 先创建一个测试角色
	createReq := CreateRoleRequest{
		TenantID:    SystemTenantID,
		RoleCode:    "TestRoleUpdate",
		DisplayName: "Test Role Update",
		Description: "Test role description",
	}
	createResp, err := roleService.CreateRole(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer roleRepo.DeleteRole(ctx, createResp.RoleID)

	// 测试更新角色
	updateReq := UpdateRoleRequest{
		RoleID:      createResp.RoleID,
		UserRole:    "SystemAdmin",
		DisplayName: func() *string { s := "Updated Test Role"; return &s }(),
		Description: func() *string { s := "Updated description"; return &s }(),
	}

	err = roleService.UpdateRole(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	t.Logf("UpdateRole success: role_id=%s", createResp.RoleID)
}

func TestRoleService_UpdateRoleStatus(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 先创建一个测试角色
	createReq := CreateRoleRequest{
		TenantID:    SystemTenantID,
		RoleCode:    "TestRoleStatus",
		DisplayName: "Test Role Status",
		Description: "Test role description",
	}
	createResp, err := roleService.CreateRole(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	defer roleRepo.DeleteRole(ctx, createResp.RoleID)

	// 测试更新角色状态
	updateReq := UpdateRoleRequest{
		RoleID:   createResp.RoleID,
		IsActive: func() *bool { b := false; return &b }(),
	}

	err = roleService.UpdateRole(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateRoleStatus failed: %v", err)
	}

	t.Logf("UpdateRoleStatus success: role_id=%s", createResp.RoleID)
}

func TestRoleService_DeleteRole(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 先创建一个测试角色
	createReq := CreateRoleRequest{
		TenantID:    SystemTenantID,
		RoleCode:    "TestRoleDelete",
		DisplayName: "Test Role Delete",
		Description: "Test role description",
	}
	createResp, err := roleService.CreateRole(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// 测试删除角色
	deleteReq := UpdateRoleRequest{
		RoleID:  createResp.RoleID,
		Delete:  func() *bool { b := true; return &b }(),
	}

	err = roleService.UpdateRole(ctx, deleteReq)
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	t.Logf("DeleteRole success: role_id=%s", createResp.RoleID)
}

func TestRoleService_ProtectedRoles(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	roleRepo := repository.NewPostgresRolesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	roleService := NewRoleService(roleRepo, usersRepo, getTestLogger())

	// 获取 SystemAdmin 角色
	sysAdminRole, err := roleRepo.GetRoleByCode(ctx, func() *string { s := SystemTenantID; return &s }(), "SystemAdmin")
	if err != nil {
		t.Skipf("SystemAdmin role not found, skipping test: %v", err)
		return
	}

	// 测试禁用受保护角色（应该失败）
	updateReq := UpdateRoleRequest{
		RoleID:   sysAdminRole.RoleID,
		IsActive: func() *bool { b := false; return &b }(),
	}

	err = roleService.UpdateRole(ctx, updateReq)
	if err == nil {
		t.Fatal("UpdateRole should fail for protected role")
	}

	t.Logf("ProtectedRoles test success: error=%v", err)
}

