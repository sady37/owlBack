// +build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"wisefido-data/internal/domain"

	"owl-common/database"
	"owl-common/config"
)

// 获取测试数据库连接
func getTestDBForTenants(t *testing.T) *sql.DB {
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

	// 测试连接
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test: cannot ping database: %v", err)
		return nil
	}

	return db
}

// getEnv 和 getEnvInt 已在 postgres_users_integration_test.go 中定义，这里不再重复定义

// 清理测试数据
func cleanupTestDataForTenants(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

func TestPostgresTenantsRepository_CreateTenant(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 创建测试租户
	tenant := &domain.Tenant{
		TenantName: "Test Tenant Create",
		Domain:     "test-create.local",
		Email:      "test@example.com",
		Phone:      "1234567890",
		Status:     "active",
		Metadata:   json.RawMessage(`{"key": "value"}`),
	}

	tenantID, err := repo.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}
	defer cleanupTestDataForTenants(t, db, tenantID)

	if tenantID == "" {
		t.Fatal("Expected non-empty tenant_id")
	}

	// 验证创建成功
	createdTenant, err := repo.GetTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}

	if createdTenant.TenantName != tenant.TenantName {
		t.Errorf("Expected tenant_name '%s', got '%s'", tenant.TenantName, createdTenant.TenantName)
	}
	if createdTenant.Domain != tenant.Domain {
		t.Errorf("Expected domain '%s', got '%s'", tenant.Domain, createdTenant.Domain)
	}
	if createdTenant.Email != tenant.Email {
		t.Errorf("Expected email '%s', got '%s'", tenant.Email, createdTenant.Email)
	}
	if createdTenant.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", createdTenant.Status)
	}

	t.Logf("✅ CreateTenant test passed: tenantID=%s", tenantID)
}

func TestPostgresTenantsRepository_GetTenant(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 使用System tenant进行测试
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	tenant, err := repo.GetTenant(ctx, systemTenantID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}

	if tenant.TenantID != systemTenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", systemTenantID, tenant.TenantID)
	}
	if tenant.TenantName != "System" {
		t.Errorf("Expected tenant_name 'System', got '%s'", tenant.TenantName)
	}

	t.Logf("✅ GetTenant test passed: tenantID=%s", systemTenantID)
}

func TestPostgresTenantsRepository_GetTenantByDomain(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 使用System tenant的domain进行测试
	domain := "system.local"
	tenant, err := repo.GetTenantByDomain(ctx, domain)
	if err != nil {
		t.Fatalf("GetTenantByDomain failed: %v", err)
	}

	if tenant.Domain != domain {
		t.Errorf("Expected domain '%s', got '%s'", domain, tenant.Domain)
	}
	if tenant.TenantName != "System" {
		t.Errorf("Expected tenant_name 'System', got '%s'", tenant.TenantName)
	}

	t.Logf("✅ GetTenantByDomain test passed: domain=%s", domain)
}

func TestPostgresTenantsRepository_ListTenants(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 测试：查询所有租户
	filter := TenantFilters{}
	tenants, total, err := repo.ListTenants(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}

	if total < 1 {
		t.Errorf("Expected at least 1 tenant, got %d", total)
	}
	if len(tenants) < 1 {
		t.Errorf("Expected at least 1 tenant in result, got %d", len(tenants))
	}

	// 测试：按status过滤
	filter = TenantFilters{Status: "active"}
	tenants, total, err = repo.ListTenants(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListTenants (with status filter) failed: %v", err)
	}

	for _, tenant := range tenants {
		if tenant.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", tenant.Status)
		}
	}

	// 测试：按tenant_name搜索
	filter = TenantFilters{Search: "System"}
	tenants, total, err = repo.ListTenants(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListTenants (with search) failed: %v", err)
	}

	found := false
	for _, tenant := range tenants {
		if tenant.TenantName == "System" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'System' tenant in search results")
	}

	t.Logf("✅ ListTenants test passed: total=%d", total)
}

func TestPostgresTenantsRepository_UpdateTenant(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 创建测试租户
	tenant := &domain.Tenant{
		TenantName: "Test Tenant Update",
		Domain:     "test-update.local",
		Status:     "active",
	}

	tenantID, err := repo.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}
	defer cleanupTestDataForTenants(t, db, tenantID)

	// 更新租户
	updatedTenant := &domain.Tenant{
		TenantName: "Updated Tenant Name",
		Email:      "updated@example.com",
		Phone:      "9876543210",
		Status:     "suspended",
	}

	err = repo.UpdateTenant(ctx, tenantID, updatedTenant)
	if err != nil {
		t.Fatalf("UpdateTenant failed: %v", err)
	}

	// 验证更新成功
	tenant, err = repo.GetTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}

	if tenant.TenantName != updatedTenant.TenantName {
		t.Errorf("Expected tenant_name '%s', got '%s'", updatedTenant.TenantName, tenant.TenantName)
	}
	if tenant.Email != updatedTenant.Email {
		t.Errorf("Expected email '%s', got '%s'", updatedTenant.Email, tenant.Email)
	}
	if tenant.Status != updatedTenant.Status {
		t.Errorf("Expected status '%s', got '%s'", updatedTenant.Status, tenant.Status)
	}

	t.Logf("✅ UpdateTenant test passed: tenantID=%s", tenantID)
}

func TestPostgresTenantsRepository_SetTenantStatus(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 创建测试租户
	tenant := &domain.Tenant{
		TenantName: "Test Tenant Status",
		Domain:     "test-status.local",
		Status:     "active",
	}

	tenantID, err := repo.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}
	defer cleanupTestDataForTenants(t, db, tenantID)

	// 更新状态
	err = repo.SetTenantStatus(ctx, tenantID, "suspended")
	if err != nil {
		t.Fatalf("SetTenantStatus failed: %v", err)
	}

	// 验证状态更新
	tenant, err = repo.GetTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}

	if tenant.Status != "suspended" {
		t.Errorf("Expected status 'suspended', got '%s'", tenant.Status)
	}

	t.Logf("✅ SetTenantStatus test passed: tenantID=%s", tenantID)
}

func TestPostgresTenantsRepository_DeleteTenant(t *testing.T) {
	db := getTestDBForTenants(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTenantsRepository(db)
	ctx := context.Background()

	// 创建测试租户
	tenant := &domain.Tenant{
		TenantName: "Test Tenant Delete",
		Domain:     "test-delete.local",
		Status:     "active",
	}

	tenantID, err := repo.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}
	defer cleanupTestDataForTenants(t, db, tenantID)

	// 软删除租户
	err = repo.DeleteTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("DeleteTenant failed: %v", err)
	}

	// 验证软删除（status='deleted'）
	tenant, err = repo.GetTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}

	if tenant.Status != "deleted" {
		t.Errorf("Expected status 'deleted' (soft delete), got '%s'", tenant.Status)
	}

	// 验证租户仍然存在（软删除）
	if tenant.TenantID != tenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", tenantID, tenant.TenantID)
	}

	t.Logf("✅ DeleteTenant test passed: tenantID=%s (soft delete)", tenantID)
}

