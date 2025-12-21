// +build integration

package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// setupSleepaceTestData 设置 Sleepace Report 测试数据
func setupSleepaceTestData(t *testing.T, db *sql.DB, tenantID string) (deviceID, residentID, unitID, roomID, bedID, caregiverID, managerID string) {
	ctx := context.Background()

	// 1. 创建单元（unit）
	unitID = "00000000-0000-0000-0000-000000000101"
	_, err := db.ExecContext(ctx,
		`INSERT INTO units (unit_id, tenant_id, unit_name, branch_tag, area_tag, unit_number, unit_type, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, branch_tag = EXCLUDED.branch_tag, unit_type = EXCLUDED.unit_type, timezone = EXCLUDED.timezone`,
		unitID, tenantID, "Test Unit", "BranchA", "Area1", "101", "Home", "Asia/Shanghai",
	)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	// 2. 创建房间（room）
	roomID = "00000000-0000-0000-0000-000000000201"
	_, err = db.ExecContext(ctx,
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID, tenantID, unitID, "Test Room",
	)
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// 3. 创建床位（bed）
	bedID = "00000000-0000-0000-0000-000000000301"
	_, err = db.ExecContext(ctx,
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name, bed_type)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name, bed_type = EXCLUDED.bed_type`,
		bedID, tenantID, roomID, "Test Bed", "ActiveBed",
	)
	if err != nil {
		t.Fatalf("Failed to create test bed: %v", err)
	}

	// 4. 创建住户（resident）
	residentID = "00000000-0000-0000-0000-000000000401"
	accountHash := []byte("test_resident_account_hash")
	_, err = db.ExecContext(ctx,
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, status, admission_date, unit_id, room_id, bed_id)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_DATE, $7, $8, $9)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID, tenantID, "test_resident", accountHash, "Test Resident", "active", unitID, roomID, bedID,
	)
	if err != nil {
		t.Fatalf("Failed to create test resident: %v", err)
	}

	// 5. 创建设备（device）
	deviceID = "00000000-0000-0000-0000-000000000501"
	_, err = db.ExecContext(ctx,
		`INSERT INTO devices (device_id, tenant_id, device_name, serial_number, status, business_access, bound_bed_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name`,
		deviceID, tenantID, "Test Sleepace Device", "SN123456", "online", "approved", bedID,
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	// 6. 创建 Caregiver 用户
	caregiverID = "00000000-0000-0000-0000-000000000601"
	accountHash2 := []byte("test_caregiver_account_hash")
	_, err = db.ExecContext(ctx,
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET role = EXCLUDED.role`,
		caregiverID, tenantID, "test_caregiver", accountHash2, []byte("password_hash"), "Test Caregiver", "Caregiver",
	)
	if err != nil {
		t.Fatalf("Failed to create test caregiver: %v", err)
	}

	// 7. 创建 Manager 用户
	managerID = "00000000-0000-0000-0000-000000000701"
	accountHash3 := []byte("test_manager_account_hash")
	_, err = db.ExecContext(ctx,
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status, branch_tag)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', $8)
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET role = EXCLUDED.role, branch_tag = EXCLUDED.branch_tag`,
		managerID, tenantID, "test_manager", accountHash3, []byte("password_hash"), "Test Manager", "Manager", "BranchA",
	)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// 8. 创建权限配置（Caregiver assigned_only）
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
		 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only`,
		SystemTenantID(), "Caregiver", "residents", "R", true, false,
	)
	if err != nil {
		t.Fatalf("Failed to create role permission for Caregiver: %v", err)
	}

	// 9. 创建权限配置（Manager branch_only）
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
		 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only`,
		SystemTenantID(), "Manager", "residents", "R", false, true,
	)
	if err != nil {
		t.Fatalf("Failed to create role permission for Manager: %v", err)
	}

	// 10. 创建住户分配关系（resident_caregivers）
	userListJSON, _ := json.Marshal([]string{caregiverID})
	_, err = db.ExecContext(ctx,
		`INSERT INTO resident_caregivers (tenant_id, resident_id, userList)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, resident_id) DO UPDATE SET userList = EXCLUDED.userList`,
		tenantID, residentID, userListJSON,
	)
	if err != nil {
		t.Fatalf("Failed to create resident_caregivers: %v", err)
	}

	return deviceID, residentID, unitID, roomID, bedID, caregiverID, managerID
}

// cleanupSleepaceTestData 清理 Sleepace Report 测试数据
func cleanupSleepaceTestData(t *testing.T, db *sql.DB, tenantID string) {
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM sleepace_report WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM role_permissions WHERE tenant_id = $1`, SystemTenantID())
}

// checkSleepaceReportTableExists 检查 sleepace_report 表是否存在
func checkSleepaceReportTableExists(t *testing.T, db *sql.DB) bool {
	var exists bool
	err := db.QueryRowContext(context.Background(),
		`SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sleepace_report'
		)`,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if sleepace_report table exists: %v", err)
	}
	return exists
}

// TestSleepaceReportHandler_Resident_CanViewOwnReports 测试住户可以查看自己的报告
func TestSleepaceReportHandler_Resident_CanViewOwnReports(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 检查 sleepace_report 表是否存在
	if !checkSleepaceReportTableExists(t, db) {
		t.Skip("sleepace_report table does not exist, skipping test")
	}

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)
	defer cleanupSleepaceTestData(t, db, tenantID)

	deviceID, residentID, _, _, _, _, _ := setupSleepaceTestData(t, db, tenantID)

	// 创建 Handler
	logger := zap.NewNop()
	sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
	handler := NewSleepaceReportHandler(sleepaceReportService, db, logger)

	// 准备请求（住户查看自己的报告）
	req := httptest.NewRequest(http.MethodGet, "/sleepace/api/v1/sleepace/reports/"+deviceID, nil)
	req.Header.Set("X-Tenant-Id", tenantID)
	req.Header.Set("X-User-Id", residentID)
	req.Header.Set("X-User-Type", "resident")
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应（应该成功，即使没有报告数据）
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 权限检查应该通过（住户可以查看自己的报告）
	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d, message: %s", result.Code, result.Message)
	}
}

// TestSleepaceReportHandler_Resident_CannotViewOtherReports 测试住户不能查看其他住户的报告
func TestSleepaceReportHandler_Resident_CannotViewOtherReports(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 检查 sleepace_report 表是否存在
	if !checkSleepaceReportTableExists(t, db) {
		t.Skip("sleepace_report table does not exist, skipping test")
	}

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)
	defer cleanupSleepaceTestData(t, db, tenantID)

	deviceID, _, _, _, _, _, _ := setupSleepaceTestData(t, db, tenantID)

	// 创建另一个住户
	otherResidentID := "00000000-0000-0000-0000-000000000402"
	accountHash := []byte("other_resident_account_hash")
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, status, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_DATE)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		otherResidentID, tenantID, "other_resident", accountHash, "Other Resident", "active",
	)
	if err != nil {
		t.Fatalf("Failed to create other resident: %v", err)
	}

	// 创建 Handler
	logger := zap.NewNop()
	sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
	handler := NewSleepaceReportHandler(sleepaceReportService, db, logger)

	// 准备请求（住户查看其他住户的报告）
	req := httptest.NewRequest(http.MethodGet, "/sleepace/api/v1/sleepace/reports/"+deviceID, nil)
	req.Header.Set("X-Tenant-Id", tenantID)
	req.Header.Set("X-User-Id", otherResidentID) // 不同的住户ID
	req.Header.Set("X-User-Type", "resident")
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应（应该被拒绝）
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 权限检查应该失败
	if result.Code == 2000 {
		t.Errorf("Expected permission denied, but got success. Response: %s", w.Body.String())
	}
	if result.Type != "error" {
		t.Errorf("Expected type 'error', got '%s'", result.Type)
	}
	if !strings.Contains(result.Message, "access denied") {
		t.Errorf("Expected error message to contain 'access denied', got '%s'", result.Message)
	}
}

// TestSleepaceReportHandler_Caregiver_CanViewAssignedResidentReports 测试 Caregiver 可以查看分配的住户报告
func TestSleepaceReportHandler_Caregiver_CanViewAssignedResidentReports(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 检查 sleepace_report 表是否存在
	if !checkSleepaceReportTableExists(t, db) {
		t.Skip("sleepace_report table does not exist, skipping test")
	}

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)
	defer cleanupSleepaceTestData(t, db, tenantID)

	deviceID, _, _, _, _, caregiverID, _ := setupSleepaceTestData(t, db, tenantID)

	// 创建 Handler
	logger := zap.NewNop()
	sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
	handler := NewSleepaceReportHandler(sleepaceReportService, db, logger)

	// 准备请求（Caregiver 查看分配的住户报告）
	req := httptest.NewRequest(http.MethodGet, "/sleepace/api/v1/sleepace/reports/"+deviceID, nil)
	req.Header.Set("X-Tenant-Id", tenantID)
	req.Header.Set("X-User-Id", caregiverID)
	req.Header.Set("X-User-Type", "staff")
	req.Header.Set("X-User-Role", "Caregiver")
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应（应该成功）
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 权限检查应该通过（Caregiver 可以查看分配的住户报告）
	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d, message: %s", result.Code, result.Message)
	}
}

// TestSleepaceReportHandler_Manager_CanViewBranchResidentReports 测试 Manager 可以查看同分支的住户报告
func TestSleepaceReportHandler_Manager_CanViewBranchResidentReports(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 检查 sleepace_report 表是否存在
	if !checkSleepaceReportTableExists(t, db) {
		t.Skip("sleepace_report table does not exist, skipping test")
	}

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)
	defer cleanupSleepaceTestData(t, db, tenantID)

	deviceID, _, _, _, _, _, managerID := setupSleepaceTestData(t, db, tenantID)

	// 创建 Handler
	logger := zap.NewNop()
	sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
	handler := NewSleepaceReportHandler(sleepaceReportService, db, logger)

	// 准备请求（Manager 查看同分支的住户报告）
	req := httptest.NewRequest(http.MethodGet, "/sleepace/api/v1/sleepace/reports/"+deviceID, nil)
	req.Header.Set("X-Tenant-Id", tenantID)
	req.Header.Set("X-User-Id", managerID)
	req.Header.Set("X-User-Type", "staff")
	req.Header.Set("X-User-Role", "Manager")
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应（应该成功）
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 权限检查应该通过（Manager 可以查看同分支的住户报告）
	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d, message: %s", result.Code, result.Message)
	}
}

// TestSleepaceReportHandler_DeviceWithoutResident_Allowed 测试设备没有关联住户时允许访问（fallback）
func TestSleepaceReportHandler_DeviceWithoutResident_Allowed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 检查 sleepace_report 表是否存在
	if !checkSleepaceReportTableExists(t, db) {
		t.Skip("sleepace_report table does not exist, skipping test")
	}

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)

	// 创建设备（不关联住户）
	deviceID := "00000000-0000-0000-0000-000000000501"
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO devices (device_id, tenant_id, device_name, serial_number, status, business_access)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name`,
		deviceID, tenantID, "Test Device Without Resident", "SN999999", "online", "approved",
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}
	defer db.ExecContext(context.Background(), `DELETE FROM devices WHERE tenant_id = $1`, tenantID)

	// 创建 Handler
	logger := zap.NewNop()
	sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
	handler := NewSleepaceReportHandler(sleepaceReportService, db, logger)

	// 准备请求（任何用户都可以访问没有关联住户的设备）
	req := httptest.NewRequest(http.MethodGet, "/sleepace/api/v1/sleepace/reports/"+deviceID, nil)
	req.Header.Set("X-Tenant-Id", tenantID)
	req.Header.Set("X-User-Id", "00000000-0000-0000-0000-000000000999")
	req.Header.Set("X-User-Type", "staff")
	req.Header.Set("X-User-Role", "Manager")
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应（应该成功，因为设备没有关联住户，fallback 允许访问）
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 应该允许访问（fallback）
	if result.Code != 2000 {
		t.Errorf("Expected code 2000 (fallback for device without resident), got %d, message: %s", result.Code, result.Message)
	}
}

