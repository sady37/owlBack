//go:build integration
// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
)

// ============================================
// GetResidentAccountSettings 测试
// ============================================

// TestResidentService_GetResidentAccountSettings_Success 测试获取住户账户设置成功
func TestResidentService_GetResidentAccountSettings_Success(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建测试住户
	residentID := createTestResidentWithAccount(t, db, tenantID, unitID, "testresident", "password123", "test@example.com", "1234567890")

	// 测试获取自己的账户设置
	req := GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID, // 自己查看自己的账户设置
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
	}

	resp, err := residentService.GetResidentAccountSettings(context.Background(), req)
	require.NoError(t, err, "GetResidentAccountSettings should succeed")

	require.NotNil(t, resp, "Response should not be nil")
	require.Equal(t, "testresident", *resp.ResidentAccount, "Account should match")
	require.Equal(t, "test@example.com", *resp.Email, "Email should match")
	require.Equal(t, "1234567890", *resp.Phone, "Phone should match")
	require.True(t, resp.SaveEmail, "SaveEmail should be true")
	require.True(t, resp.SavePhone, "SavePhone should be true")
	require.False(t, resp.IsContact, "IsContact should be false")
}

// TestResidentService_GetResidentAccountSettings_PermissionDenied 测试权限拒绝
func TestResidentService_GetResidentAccountSettings_PermissionDenied(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建两个测试住户
	residentID1 := createTestResidentWithAccount(t, db, tenantID, unitID, "resident1", "password1", "resident1@test.com", "1111111111")
	residentID2 := createTestResidentWithAccount(t, db, tenantID, unitID, "resident2", "password2", "resident2@test.com", "2222222222")

	// 测试：Resident 不能查看其他 Resident 的账户设置
	req := GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID2,
		CurrentUserID:   residentID1, // Resident1 尝试查看 Resident2 的账户设置
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
	}

	_, err := residentService.GetResidentAccountSettings(context.Background(), req)
	require.Error(t, err, "Should return error")
	require.Contains(t, err.Error(), "permission denied", "Error should contain 'permission denied'")
}

// TestResidentService_GetResidentAccountSettings_FamilyRole_NotSupported 测试 Family 角色不支持
func TestResidentService_GetResidentAccountSettings_FamilyRole_NotSupported(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建测试住户
	residentID := createTestResidentWithAccount(t, db, tenantID, unitID, "testresident", "password123", "test@example.com", "1234567890")

	// 测试：Family 角色不支持账户设置
	req := GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID,
		CurrentUserType: "Contact",
		CurrentUserRole: "Family", // Family 角色
	}

	_, err := residentService.GetResidentAccountSettings(context.Background(), req)
	require.Error(t, err, "Should return error")
	require.Contains(t, err.Error(), "contacts do not log in", "Error should indicate contacts don't log in")
}

// ============================================
// UpdateResidentAccountSettings 测试
// ============================================

// TestResidentService_UpdateResidentAccountSettings_Success 测试更新账户设置成功
func TestResidentService_UpdateResidentAccountSettings_Success(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建测试住户
	residentID := createTestResidentWithAccount(t, db, tenantID, unitID, "testresident", "oldpassword", "old@example.com", "1111111111")

	// 生成新密码 hash
	newPassword := "newpassword123"
	passwordHash := sha256.Sum256([]byte(newPassword))
	passwordHashHex := hex.EncodeToString(passwordHash[:])

	// 生成 email hash
	emailHash := sha256.Sum256([]byte("new@example.com"))
	emailHashHex := hex.EncodeToString(emailHash[:])

	// 生成 phone hash
	phoneHash := sha256.Sum256([]byte("9999999999"))
	phoneHashHex := hex.EncodeToString(phoneHash[:])

	// 测试更新账户设置
	req := UpdateResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID, // 自己更新自己的账户设置
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
		PasswordHash:    &passwordHashHex,
		Email:           stringPtr("new@example.com"),
		EmailHash:       &emailHashHex,
		Phone:           stringPtr("9999999999"),
		PhoneHash:       &phoneHashHex,
		SaveEmail:       boolPtr(true),
		SavePhone:       boolPtr(true),
	}

	resp, err := residentService.UpdateResidentAccountSettings(context.Background(), req)
	require.NoError(t, err, "UpdateResidentAccountSettings should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	require.True(t, resp.Success, "Success should be true")

	// 验证更新后的数据
	getReq := GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID,
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
	}

	getResp, err := residentService.GetResidentAccountSettings(context.Background(), getReq)
	require.NoError(t, err, "GetResidentAccountSettings should succeed")
	require.Equal(t, "new@example.com", *getResp.Email, "Email should be updated")
	require.Equal(t, "9999999999", *getResp.Phone, "Phone should be updated")
	require.True(t, getResp.SaveEmail, "SaveEmail should be true")
	require.True(t, getResp.SavePhone, "SavePhone should be true")
}

// TestResidentService_UpdateResidentAccountSettings_PermissionDenied 测试权限拒绝
func TestResidentService_UpdateResidentAccountSettings_PermissionDenied(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建两个测试住户
	residentID1 := createTestResidentWithAccount(t, db, tenantID, unitID, "resident1", "password1", "resident1@test.com", "1111111111")
	residentID2 := createTestResidentWithAccount(t, db, tenantID, unitID, "resident2", "password2", "resident2@test.com", "2222222222")

	// 测试：Resident 不能更新其他 Resident 的账户设置
	req := UpdateResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID2,
		CurrentUserID:   residentID1, // Resident1 尝试更新 Resident2 的账户设置
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
		Email:           stringPtr("hacked@example.com"),
	}

	_, err := residentService.UpdateResidentAccountSettings(context.Background(), req)
	require.Error(t, err, "Should return error")
	require.Contains(t, err.Error(), "permission denied", "Error should contain 'permission denied'")
}

// TestResidentService_UpdateResidentAccountSettings_FamilyRole_NotSupported 测试 Family 角色不支持
func TestResidentService_UpdateResidentAccountSettings_FamilyRole_NotSupported(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建测试住户
	residentID := createTestResidentWithAccount(t, db, tenantID, unitID, "testresident", "password123", "test@example.com", "1234567890")

	// 测试：Family 角色不支持账户设置更新
	req := UpdateResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID,
		CurrentUserType: "Contact",
		CurrentUserRole: "Family", // Family 角色
		Email:           stringPtr("new@example.com"),
	}

	_, err := residentService.UpdateResidentAccountSettings(context.Background(), req)
	require.Error(t, err, "Should return error")
	require.Contains(t, err.Error(), "contacts do not log in", "Error should indicate contacts don't log in")
}

// TestResidentService_UpdateResidentAccountSettings_PlaceholderHandling 测试占位符处理
func TestResidentService_UpdateResidentAccountSettings_PlaceholderHandling(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建 Service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, nil, getTestLoggerForResident())

	// 创建测试住户（带占位符）
	residentID := createTestResidentWithPlaceholder(t, db, tenantID, unitID, "testresident", "password123")

	// 获取账户设置（应该显示占位符，save_email/save_phone 为 false）
	getReq := GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   residentID,
		CurrentUserType: "Resident",
		CurrentUserRole: "Resident",
	}

	getResp, err := residentService.GetResidentAccountSettings(context.Background(), getReq)
	require.NoError(t, err, "GetResidentAccountSettings should succeed")

	// 占位符应该不返回在 email/phone 中，或者 save_email/save_phone 为 false
	if getResp.Email != nil {
		require.NotEqual(t, "***@***", *getResp.Email, "Email should not be placeholder")
	}
	if getResp.Phone != nil {
		require.NotEqual(t, "xxx-xxx-xxxx", *getResp.Phone, "Phone should not be placeholder")
	}
}

// ============================================
// 辅助函数
// ============================================

// createTestResidentWithAccount 创建带账户信息的测试住户（使用基于 account 的稳定 UUID）
func createTestResidentWithAccount(t *testing.T, db *sql.DB, tenantID, unitID, account, password, email, phone string) string {
	// 使用 account 的 hash 生成一个稳定的 UUID（基于 account 名称）
	// 这样同一个 account 总是生成相同的 ID，避免重复
	hash := sha256.Sum256([]byte(account))
	uuidBytes := hash[:16]
	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x40 // Version 4
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80 // Variant 10
	
	residentID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:16])
	
	return createTestResidentWithAccountAndID(t, db, tenantID, unitID, residentID, account, password, email, phone)
}

// createTestResidentWithAccountAndID 创建带账户信息的测试住户（指定 ID）
func createTestResidentWithAccountAndID(t *testing.T, db *sql.DB, tenantID, unitID, residentID, account, password, email, phone string) string {

	// 生成密码 hash
	passwordHash := sha256.Sum256([]byte(password))
	passwordHashBytes := passwordHash[:]

	// 生成 account hash
	accountHash := sha256.Sum256([]byte(strings.ToLower(account)))
	accountHashBytes := accountHash[:]

	// 生成 email hash
	emailHash := sha256.Sum256([]byte(strings.ToLower(email)))
	emailHashBytes := emailHash[:]

	// 生成 phone hash
	phoneHash := sha256.Sum256([]byte(phone))
	phoneHashBytes := phoneHash[:]

	// 插入 residents 表（admission_date 是必填字段）
	admissionDate := "2024-01-01"
	_, err := db.Exec(
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, password_hash, email_hash, phone_hash, nickname, status, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', $9)
		 ON CONFLICT (tenant_id, resident_account) DO UPDATE SET
		 resident_account_hash = EXCLUDED.resident_account_hash,
		 password_hash = EXCLUDED.password_hash,
		 email_hash = EXCLUDED.email_hash,
		 phone_hash = EXCLUDED.phone_hash,
		 nickname = EXCLUDED.nickname,
		 admission_date = EXCLUDED.admission_date`,
		residentID, tenantID, account, accountHashBytes, passwordHashBytes, emailHashBytes, phoneHashBytes, "Test Resident "+account, admissionDate,
	)
	require.NoError(t, err)

	// 插入 resident_phi 表（保存明文）
	_, err = db.Exec(
		`INSERT INTO resident_phi (tenant_id, resident_id, resident_email, resident_phone)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, resident_id) DO UPDATE SET
		 resident_email = EXCLUDED.resident_email,
		 resident_phone = EXCLUDED.resident_phone`,
		tenantID, residentID, email, phone,
	)
	require.NoError(t, err)

	return residentID
}

// createTestResidentWithPlaceholder 创建带占位符的测试住户
func createTestResidentWithPlaceholder(t *testing.T, db *sql.DB, tenantID, unitID, account, password string) string {
	residentID := "00000000-0000-0000-0000-000000000101"

	// 生成密码 hash
	passwordHash := sha256.Sum256([]byte(password))
	passwordHashBytes := passwordHash[:]

	// 生成 account hash
	accountHash := sha256.Sum256([]byte(strings.ToLower(account)))
	accountHashBytes := accountHash[:]

	// 生成 email hash（有 hash 但未保存明文）
	emailHash := sha256.Sum256([]byte("test@example.com"))
	emailHashBytes := emailHash[:]

	// 生成 phone hash（有 hash 但未保存明文）
	phoneHash := sha256.Sum256([]byte("1234567890"))
	phoneHashBytes := phoneHash[:]

	// 插入 residents 表（有 hash，admission_date 是必填字段）
	admissionDate := "2024-01-01"
	_, err := db.Exec(
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, password_hash, email_hash, phone_hash, nickname, status, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', $9)
		 ON CONFLICT (tenant_id, resident_account) DO UPDATE SET
		 resident_account_hash = EXCLUDED.resident_account_hash,
		 password_hash = EXCLUDED.password_hash,
		 email_hash = EXCLUDED.email_hash,
		 phone_hash = EXCLUDED.phone_hash,
		 nickname = EXCLUDED.nickname,
		 admission_date = EXCLUDED.admission_date`,
		residentID, tenantID, account, accountHashBytes, passwordHashBytes, emailHashBytes, phoneHashBytes, "Test Resident "+account, admissionDate,
	)
	require.NoError(t, err)

	// 插入 resident_phi 表（带占位符）
	_, err = db.Exec(
		`INSERT INTO resident_phi (tenant_id, resident_id, resident_email, resident_phone)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, resident_id) DO UPDATE SET
		 resident_email = EXCLUDED.resident_email,
		 resident_phone = EXCLUDED.resident_phone`,
		tenantID, residentID, "***@***", "xxx-xxx-xxxx",
	)
	require.NoError(t, err)

	return residentID
}

// stringPtr 创建字符串指针
func stringPtr(s string) *string {
	return &s
}

// boolPtr 创建布尔指针
func boolPtr(b bool) *bool {
	return &b
}

