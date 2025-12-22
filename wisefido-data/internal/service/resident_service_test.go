// +build integration

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestDBForResident 设置测试数据库
func setupTestDBForResident(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// getTestLoggerForResident 获取测试日志记录器
func getTestLoggerForResident() *zap.Logger {
	return getTestLogger()
}

// createTestTenantAndUnitForResident 创建测试租户和单元
func createTestTenantAndUnitForResident(t *testing.T, db *sql.DB) (string, string) {
	tenantID := "00000000-0000-0000-0000-000000000999"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Resident Tenant", "test-resident.local",
	)
	require.NoError(t, err)

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000998"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	require.NoError(t, err)

	// 创建测试unit
	unitID := "00000000-0000-0000-0000-000000000997"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver",
	)
	require.NoError(t, err)

	return tenantID, unitID
}

// cleanupTestDataForResident 清理测试数据
func cleanupTestDataForResident(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// TestListResidents_Basic 测试基本的 ListResidents 功能
func TestListResidents_Basic(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试住户
	now := time.Now()
	var residentID string
	err := db.QueryRow(`
		INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, nickname,
			status, admission_date, can_view_status, unit_id
		) VALUES (
			$1, 'test_account', '\x' || encode(sha256('test_account'::bytea), 'hex')::bytea, 'Test Resident',
			'active', $2, true, $3
		)
		RETURNING resident_id::text
	`, tenantID, now, unitID).Scan(&residentID)
	require.NoError(t, err)

	// 创建测试数据
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, logger)

	// 测试基本查询
	req := ListResidentsRequest{
		TenantID:        tenantID,
		CurrentUserType: "staff",
		Page:            1,
		PageSize:        20,
	}

	resp, err := service.ListResidents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListResidents failed: %v", err)
	}

	if resp == nil {
		t.Fatal("ListResidents returned nil response")
	}

	if resp.Items == nil {
		t.Fatal("ListResidents returned nil items")
	}

	// 验证响应结构
	t.Logf("ListResidents returned %d items, total: %d", len(resp.Items), resp.Total)
}

// TestGetResident_Basic 测试基本的 GetResident 功能
func TestGetResident_Basic(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试住户
	var residentID string
	now := time.Now()
	err := db.QueryRow(`
		INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, nickname,
			status, admission_date, can_view_status, unit_id
		) VALUES (
			$1, 'test_account', '\x' || encode(sha256('test_account'::bytea), 'hex')::bytea, 'Test Resident',
			'active', $2, true, $3
		)
		RETURNING resident_id::text
	`, tenantID, now, unitID).Scan(&residentID)
	require.NoError(t, err)

	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, logger)

	// 测试基本查询
	req := GetResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserType: "staff",
		IncludePHI:      false,
		IncludeContacts: false,
	}

	resp, err := service.GetResident(context.Background(), req)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetResident returned nil response")
	}

	if resp.Resident == nil {
		t.Fatal("GetResident returned nil resident")
	}

	if resp.Resident.ResidentID != residentID {
		t.Errorf("Expected resident_id %s, got %s", residentID, resp.Resident.ResidentID)
	}

	t.Logf("GetResident returned resident: %+v", resp.Resident)
}

// TestListResidents_ResidentLogin 测试 Resident 登录时的权限过滤
func TestListResidents_ResidentLogin(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试住户
	var residentID string
	now := time.Now()
	err := db.QueryRow(`
		INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, nickname,
			status, admission_date, can_view_status, unit_id
		) VALUES (
			$1, 'test_resident', '\x' || encode(sha256('test_resident'::bytea), 'hex')::bytea, 'Test Resident',
			'active', $2, true, $3
		)
		RETURNING resident_id::text
	`, tenantID, now, unitID).Scan(&residentID)
	require.NoError(t, err)

	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, logger)

	// 测试 Resident 登录（只能查看自己）
	req := ListResidentsRequest{
		TenantID:        tenantID,
		CurrentUserID:   residentID,
		CurrentUserType: "resident",
		Page:            1,
		PageSize:        20,
	}

	resp, err := service.ListResidents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListResidents failed: %v", err)
	}

	if resp == nil || len(resp.Items) == 0 {
		t.Fatal("ListResidents should return at least one item for resident login")
	}

	// 验证只能看到自己
	if len(resp.Items) != 1 {
		t.Errorf("Expected 1 item for resident login, got %d", len(resp.Items))
	}

	if resp.Items[0].ResidentID != residentID {
		t.Errorf("Expected resident_id %s, got %s", residentID, resp.Items[0].ResidentID)
	}

	t.Logf("ListResidents (resident login) returned %d items", len(resp.Items))
}

