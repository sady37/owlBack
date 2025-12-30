//go:build integration
// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
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
		`INSERT INTO buildings (building_id, tenant_id, building_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name`,
		buildingID, tenantID, "Test Building",
	)
	require.NoError(t, err)

	// 创建测试unit
	unitID := "00000000-0000-0000-0000-000000000997"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building_name, floor, unit_type, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "America/Denver",
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

// 注意：以下测试暂时注释，因为 ListResidents 和 GetResident 方法尚未实现
// 待实现后再启用这些测试

/*
// TestListResidents_Basic 测试基本的 ListResidents 功能
func TestListResidents_Basic(t *testing.T) {
	// ... 已注释
}

// TestGetResident_Basic 测试基本的 GetResident 功能
func TestGetResident_Basic(t *testing.T) {
	// ... 已注释
}

// TestListResidents_ResidentLogin 测试 Resident 登录时的权限过滤
func TestListResidents_ResidentLogin(t *testing.T) {
	// ... 已注释
}
*/

// ============================================
// CreateResident 测试
// ============================================

// createTestUserForCreateResident 创建测试用户（用于 CreateResident 测试）
func createTestUserForCreateResident(t *testing.T, db *sql.DB, tenantID, userID, userAccount, password, role, branchTag string) {
	// 计算 hash
	accountHash := sha256.Sum256([]byte(strings.ToLower(userAccount)))
	passwordHash := sha256.Sum256([]byte(password))

	// users 表没有 branch_id 字段，branch 信息通过 user_branches 表关联
	// 这里只创建用户，不设置 branch 关联
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   user_account_hash = EXCLUDED.user_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = 'active'`,
		userID, tenantID, strings.ToLower(userAccount), accountHash[:], passwordHash[:], userAccount, role,
	)
	require.NoError(t, err)

	// 如果提供了 branchTag，需要在 user_branches 表中创建关联
	if branchTag != "" {
		var branchID string
		err := db.QueryRow(`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`, tenantID, branchTag).Scan(&branchID)
		if err == nil {
			_, err = db.Exec(
				`INSERT INTO user_branches (tenant_id, user_id, branch_id, is_primary)
				 VALUES ($1, $2, $3, true)
				 ON CONFLICT (tenant_id, user_id, branch_id) DO UPDATE SET is_primary = EXCLUDED.is_primary`,
				tenantID, userID, branchID,
			)
			require.NoError(t, err)
		}
	}
}

// createTestBranchForCreateResident 创建测试 branch
func createTestBranchForCreateResident(t *testing.T, db *sql.DB, tenantID, branchID, branchName string) {
	_, err := db.Exec(
		`INSERT INTO branches (branch_id, tenant_id, branch_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id) DO UPDATE SET branch_name = EXCLUDED.branch_name`,
		branchID, tenantID, branchName,
	)
	require.NoError(t, err)
}

// createTestPermissionForCreateResident 创建测试权限配置
func createTestPermissionForCreateResident(t *testing.T, db *sql.DB, roleCode string) {
	_, err := db.Exec(
		`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
		 VALUES ($1, $2, 'residents', 'C', false, false)
		 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
		 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only`,
		SystemTenantID, roleCode,
	)
	require.NoError(t, err)
}

// TestCreateResident_Basic 测试基本的 CreateResident 功能
func TestCreateResident_Basic(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户（Admin）
	userID := "00000000-0000-0000-0000-000000000990"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	// 准备请求（使用三层结构）
	admissionDate := time.Now().Unix()
	req := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident001",
			Nickname:        "Test Resident 001",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	// 执行创建
	ctx := context.Background()
	resp, err := service.CreateResident(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ResidentID)

	t.Logf("CreateResident succeeded, resident_id: %s", resp.ResidentID)

	// 验证数据已创建
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM residents WHERE resident_id = $1`, resp.ResidentID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestCreateResident_IndividualUser_Forbidden 测试 Individual 用户禁止创建
func TestCreateResident_IndividualUser_Forbidden(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _ := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Individual 用户
	userID := "00000000-0000-0000-0000-000000000991"
	createTestUserForCreateResident(t, db, tenantID, userID, "testindividual", "password123", "Individual", "")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	// 准备请求（使用三层结构）
	admissionDate := time.Now().Unix()
	req := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Individual",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident002",
			Nickname:        "Test Resident 002",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
	}

	// 执行创建（应该失败）
	ctx := context.Background()
	resp, err := service.CreateResident(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "individual users cannot create residents")

	t.Logf("CreateResident correctly rejected Individual user: %v", err)
}

// TestCreateResident_Manager_BranchRestriction 测试 Manager 只能创建自己 branch 的住户
func TestCreateResident_Manager_BranchRestriction(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _ := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试 branch
	branchID1 := "00000000-0000-0000-0000-000000000980"
	branchID2 := "00000000-0000-0000-0000-000000000981"
	createTestBranchForCreateResident(t, db, tenantID, branchID1, "Branch-1")
	createTestBranchForCreateResident(t, db, tenantID, branchID2, "Branch-2")

	// 创建 Manager 用户（属于 Branch-1）
	userID := "00000000-0000-0000-0000-000000000992"
	createTestUserForCreateResident(t, db, tenantID, userID, "testmanager", "password123", "Manager", "Branch-1")
	createTestPermissionForCreateResident(t, db, "Manager")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	// 测试1：Manager 尝试创建不同 branch 的住户（应该失败）
	admissionDate := time.Now().Unix()
	req1 := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Manager",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident003",
			Nickname:        "Test Resident 003",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		// UnitRelation 中的 unit 属于不同的 branch (TODO: 需要添加 branch_id 检查)
	}

	ctx := context.Background()
	resp1, err1 := service.CreateResident(ctx, req1)
	require.Error(t, err1)
	require.Nil(t, resp1)
	require.Contains(t, err1.Error(), "manager can only create residents in their own branch")

	// 测试2：Manager 创建自己 branch 的住户（应该成功）
	req2 := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Manager",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident004",
			Nickname:        "Test Resident 004",
			Status:          "active",
			AdmissionDate:   &admissionDate,
			// BranchID:        branchID1, // 相同的 branch (TODO: 需要添加 BranchID 字段)
		},
	}

	resp2, err2 := service.CreateResident(ctx, req2)
	require.NoError(t, err2)
	require.NotNil(t, resp2)
	require.NotEmpty(t, resp2.ResidentID)

	t.Logf("CreateResident (Manager own branch) succeeded, resident_id: %s", resp2.ResidentID)

	// 测试3：Manager 创建住户时 branch_id 为空（应该使用 Manager 的 branch_id）
	req3 := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Manager",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident005",
			Nickname:        "Test Resident 005",
			Status:          "active",
			AdmissionDate:   &admissionDate,
			// BranchID 为空
		},
	}

	resp3, err3 := service.CreateResident(ctx, req3)
	require.NoError(t, err3)
	require.NotNil(t, resp3)
	require.NotEmpty(t, resp3.ResidentID)

	// 验证创建的住户使用了 Manager 的 branch_id
	var createdBranchID string
	err := db.QueryRow(`SELECT branch_id::text FROM residents WHERE resident_id = $1`, resp3.ResidentID).Scan(&createdBranchID)
	require.NoError(t, err)
	require.Equal(t, branchID1, createdBranchID)

	t.Logf("CreateResident (Manager auto branch) succeeded, resident_id: %s, branch_id: %s", resp3.ResidentID, createdBranchID)
}

// TestCreateResident_DuplicateAccount 测试重复账号
func TestCreateResident_DuplicateAccount(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户
	userID := "00000000-0000-0000-0000-000000000993"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin2", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	// 先创建一个住户
	admissionDate := time.Now().Unix()
	req1 := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident006",
			Nickname:        "Test Resident 006",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	ctx := context.Background()
	resp1, err1 := service.CreateResident(ctx, req1)
	require.NoError(t, err1)
	require.NotNil(t, resp1)

	// 尝试使用相同的账号创建（应该失败）
	req2 := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident006", // 重复的账号
			Nickname:        "Test Resident 007",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	resp2, err2 := service.CreateResident(ctx, req2)
	require.Error(t, err2)
	require.Nil(t, resp2)
	require.Contains(t, err2.Error(), "resident_account already exists")

	t.Logf("CreateResident correctly rejected duplicate account: %v", err2)
}

// TestCreateResident_WithPHIAndContacts 测试创建包含 PHI 和 Contacts 的住户
func TestCreateResident_WithPHIAndContacts(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户
	userID := "00000000-0000-0000-0000-000000000994"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin3", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	// 准备请求（包含 PHI 和 Contacts）
	admissionDate := time.Now().Unix()
	dob := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	emailHash := hex.EncodeToString(sha256Hash("test@example.com"))
	phoneHash := hex.EncodeToString(sha256Hash("1234567890"))

	req := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident008",
			Nickname:        "Test Resident 008",
			Status:          "active",
			AdmissionDate:   &admissionDate,
			EmailHash:       emailHash,
			PhoneHash:       phoneHash,
			PHI: &CreateResidentPHIRequest{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "Male",
				DateOfBirth:   &dob,
				ResidentEmail: "test@example.com",
				ResidentPhone: "1234567890",
				SaveEmail:     true,
				SavePhone:     true,
			},
			Contacts: []*CreateResidentContactRequest{
			{
				Slot:             "A",
				IsEnabled:        true,
				Relationship:     "Child",
				ContactFirstName: "Jane",
				ContactLastName:  "Doe",
				ContactEmail:     "jane@example.com",
				ContactPhone:     "0987654321",
				ReceiveSMS:       true,
				ReceiveEmail:     true,
			},
		},
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	// 执行创建
	ctx := context.Background()
	resp, err := service.CreateResident(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ResidentID)

	t.Logf("CreateResident (with PHI and Contacts) succeeded, resident_id: %s", resp.ResidentID)

	// 验证 PHI 已创建
	var phiCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM resident_phi WHERE resident_id = $1`, resp.ResidentID).Scan(&phiCount)
	require.NoError(t, err)
	require.Equal(t, 1, phiCount)

	// 验证 Contact 已创建
	var contactCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM resident_contacts WHERE resident_id = $1`, resp.ResidentID).Scan(&contactCount)
	require.NoError(t, err)
	require.Equal(t, 1, contactCount)
}
