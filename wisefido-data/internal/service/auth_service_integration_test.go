// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"

	"owl-common/database"
	"owl-common/config"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// createTestTenantForAuth 创建测试租户
func createTestTenantForAuth(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000999"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Auth Tenant", "test-auth.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestUserForAuth 创建测试用户（staff）
func createTestUserForAuth(t *testing.T, db *sql.DB, tenantID, userAccount, password, email, phone, role string) {
	// 计算 hash
	accountHash := sha256.Sum256([]byte(userAccount))
	passwordHash := sha256.Sum256([]byte(password))
	var emailHash, phoneHash []byte
	if email != "" {
		eh := sha256.Sum256([]byte(email))
		emailHash = eh[:]
	}
	if phone != "" {
		ph := sha256.Sum256([]byte(phone))
		phoneHash = ph[:]
	}

	_, err := db.Exec(
		`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, email_hash, phone_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   user_account_hash = EXCLUDED.user_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   email_hash = EXCLUDED.email_hash,
		   phone_hash = EXCLUDED.phone_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = 'active'`,
		tenantID, userAccount, accountHash[:], passwordHash[:], emailHash, phoneHash, userAccount, role,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

// cleanupTestDataForAuth 清理测试数据
func cleanupTestDataForAuth(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// hashAccount 计算账号 hash
func hashAccount(account string) string {
	h := sha256.Sum256([]byte(account))
	return hex.EncodeToString(h[:])
}

// hashPassword 计算密码 hash
func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

func TestAuthService_Login_Staff_Success(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	userAccount := "test-staff-001"
	password := "test-password-123"
	email := "test-staff-001@test.com"
	phone := "720101001"
	role := "Manager"

	createTestUserForAuth(t, db, tenantID, userAccount, password, email, phone, role)

	// 测试登录（使用 user_account）
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "staff",
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Login returned nil response")
	}

	if resp.UserID == "" {
		t.Fatal("Login returned empty user_id")
	}

	if resp.UserType != "staff" {
		t.Fatalf("Login returned wrong user_type: expected 'staff', got '%s'", resp.UserType)
	}

	if resp.UserAccount != userAccount {
		t.Fatalf("Login returned wrong user_account: expected '%s', got '%s'", userAccount, resp.UserAccount)
	}

	if resp.Role != role {
		t.Fatalf("Login returned wrong role: expected '%s', got '%s'", role, resp.Role)
	}

	if resp.TenantID != tenantID {
		t.Fatalf("Login returned wrong tenant_id: expected '%s', got '%s'", tenantID, resp.TenantID)
	}

	t.Logf("✅ Login Staff (user_account) test passed: userID=%s, userAccount=%s, role=%s", resp.UserID, resp.UserAccount, resp.Role)

	// 测试登录（使用 email）
	req2 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "staff",
		AccountHash:  hashAccount(email),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp2, err := authService.Login(ctx, req2)
	if err != nil {
		t.Fatalf("Login with email failed: %v", err)
	}

	if resp2.UserID != resp.UserID {
		t.Fatalf("Login with email returned different user_id: expected '%s', got '%s'", resp.UserID, resp2.UserID)
	}

	t.Logf("✅ Login Staff (email) test passed: userID=%s", resp2.UserID)

	// 测试登录（使用 phone）
	req3 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "staff",
		AccountHash:  hashAccount(phone),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp3, err := authService.Login(ctx, req3)
	if err != nil {
		t.Fatalf("Login with phone failed: %v", err)
	}

	if resp3.UserID != resp.UserID {
		t.Fatalf("Login with phone returned different user_id: expected '%s', got '%s'", resp.UserID, resp3.UserID)
	}

	t.Logf("✅ Login Staff (phone) test passed: userID=%s", resp3.UserID)
}

func TestAuthService_Login_MissingCredentials(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 测试缺少 accountHash
	req1 := LoginRequest{
		TenantID:     "00000000-0000-0000-0000-000000000001",
		UserType:     "staff",
		AccountHash:  "",
		PasswordHash: hashPassword("test-password"),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err := authService.Login(ctx, req1)
	if err == nil {
		t.Fatal("Login should fail with missing accountHash")
	}

	if err.Error() != "missing credentials" {
		t.Fatalf("Login returned wrong error: expected 'missing credentials', got '%s'", err.Error())
	}

	// 测试缺少 passwordHash
	req2 := LoginRequest{
		TenantID:     "00000000-0000-0000-0000-000000000001",
		UserType:     "staff",
		AccountHash:  hashAccount("test-account"),
		PasswordHash: "",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err = authService.Login(ctx, req2)
	if err == nil {
		t.Fatal("Login should fail with missing passwordHash")
	}

	if err.Error() != "missing credentials" {
		t.Fatalf("Login returned wrong error: expected 'missing credentials', got '%s'", err.Error())
	}

	t.Logf("✅ Login missing credentials test passed")
}

func TestAuthService_Login_InvalidHash(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 测试无效的 accountHash
	req1 := LoginRequest{
		TenantID:     "00000000-0000-0000-0000-000000000001",
		UserType:     "staff",
		AccountHash:  "invalid-hex-string",
		PasswordHash: hashPassword("test-password"),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err := authService.Login(ctx, req1)
	if err == nil {
		t.Fatal("Login should fail with invalid accountHash")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("Login returned wrong error: expected 'invalid credentials', got '%s'", err.Error())
	}

	// 测试无效的 passwordHash
	req2 := LoginRequest{
		TenantID:     "00000000-0000-0000-0000-000000000001",
		UserType:     "staff",
		AccountHash:  hashAccount("test-account"),
		PasswordHash: "invalid-hex-string",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err = authService.Login(ctx, req2)
	if err == nil {
		t.Fatal("Login should fail with invalid passwordHash")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("Login returned wrong error: expected 'invalid credentials', got '%s'", err.Error())
	}

	t.Logf("✅ Login invalid hash test passed")
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	userAccount := "test-staff-002"
	password := "test-password-123"
	createTestUserForAuth(t, db, tenantID, userAccount, password, "", "", "Manager")

	// 测试错误的密码
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "staff",
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword("wrong-password"),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err := authService.Login(ctx, req)
	if err == nil {
		t.Fatal("Login should fail with wrong password")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("Login returned wrong error: expected 'invalid credentials', got '%s'", err.Error())
	}

	t.Logf("✅ Login invalid credentials test passed")
}

func TestAuthService_Login_UserNotActive(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	userAccount := "test-staff-inactive"
	password := "test-password-123"

	// 创建非激活用户
	accountHash := sha256.Sum256([]byte(userAccount))
	passwordHash := sha256.Sum256([]byte(password))
	_, err := db.Exec(
		`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'inactive')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET status = 'inactive'`,
		tenantID, userAccount, accountHash[:], passwordHash[:], userAccount, "Manager",
	)
	if err != nil {
		t.Fatalf("Failed to create inactive user: %v", err)
	}

	// 测试登录非激活用户
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "staff",
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err = authService.Login(ctx, req)
	if err == nil {
		t.Fatal("Login should fail for inactive user")
	}

	if err.Error() != "user is not active" {
		t.Fatalf("Login returned wrong error: expected 'user is not active', got '%s'", err.Error())
	}

	t.Logf("✅ Login user not active test passed")
}

func TestAuthService_Login_AutoResolveTenantID(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	userAccount := "test-staff-auto-tenant"
	password := "test-password-123"
	createTestUserForAuth(t, db, tenantID, userAccount, password, "", "", "Manager")

	// 测试自动解析 tenant_id（不提供 tenant_id）
	req := LoginRequest{
		TenantID:     "", // 不提供 tenant_id，应该自动解析
		UserType:     "staff",
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login with auto-resolve tenant_id failed: %v", err)
	}

	if resp.TenantID != tenantID {
		t.Fatalf("Login returned wrong tenant_id: expected '%s', got '%s'", tenantID, resp.TenantID)
	}

	t.Logf("✅ Login auto-resolve tenant_id test passed: tenantID=%s", resp.TenantID)
}

func TestAuthService_SearchInstitutions_Staff_Success(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	userAccount := "test-staff-search"
	password := "test-password-123"
	createTestUserForAuth(t, db, tenantID, userAccount, password, "", "", "Manager")

	// 测试搜索机构
	req := SearchInstitutionsRequest{
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		UserType:     "staff",
	}

	resp, err := authService.SearchInstitutions(ctx, req)
	if err != nil {
		t.Fatalf("SearchInstitutions failed: %v", err)
	}

	if resp == nil {
		t.Fatal("SearchInstitutions returned nil response")
	}

	if len(resp.Institutions) == 0 {
		t.Fatal("SearchInstitutions should return at least one institution")
	}

	found := false
	for _, inst := range resp.Institutions {
		if inst.ID == tenantID {
			found = true
			if inst.Name != "Test Auth Tenant" {
				t.Fatalf("SearchInstitutions returned wrong tenant_name: expected 'Test Auth Tenant', got '%s'", inst.Name)
			}
			break
		}
	}

	if !found {
		t.Fatalf("SearchInstitutions did not return expected tenant_id: %s", tenantID)
	}

	t.Logf("✅ SearchInstitutions Staff test passed: found %d institutions", len(resp.Institutions))
}

func TestAuthService_SearchInstitutions_NoMatch(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 测试搜索不存在的账号
	req := SearchInstitutionsRequest{
		AccountHash:  hashAccount("non-existent-account"),
		PasswordHash: hashPassword("non-existent-password"),
		UserType:     "staff",
	}

	resp, err := authService.SearchInstitutions(ctx, req)
	if err != nil {
		t.Fatalf("SearchInstitutions should return empty array for no match, got error: %v", err)
	}

	if resp == nil {
		t.Fatal("SearchInstitutions returned nil response")
	}

	if len(resp.Institutions) != 0 {
		t.Fatalf("SearchInstitutions should return empty array for no match, got %d institutions", len(resp.Institutions))
	}

	t.Logf("✅ SearchInstitutions no match test passed")
}

func TestAuthService_SearchInstitutions_InvalidHash(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 测试无效的 hash
	req := SearchInstitutionsRequest{
		AccountHash:  "invalid-hex",
		PasswordHash: "invalid-hex",
		UserType:     "staff",
	}

	resp, err := authService.SearchInstitutions(ctx, req)
	if err != nil {
		t.Fatalf("SearchInstitutions should return empty array for invalid hash, got error: %v", err)
	}

	if resp == nil {
		t.Fatal("SearchInstitutions returned nil response")
	}

	if len(resp.Institutions) != 0 {
		t.Fatalf("SearchInstitutions should return empty array for invalid hash, got %d institutions", len(resp.Institutions))
	}

	t.Logf("✅ SearchInstitutions invalid hash test passed")
}

func TestAuthService_SearchInstitutions_MultipleTenants(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建两个测试租户
	tenantID1 := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID1)

	tenantID2 := "00000000-0000-0000-0000-000000000998"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID2, "Test Auth Tenant 2", "test-auth-2.local",
	)
	if err != nil {
		t.Fatalf("Failed to create second test tenant: %v", err)
	}
	defer cleanupTestDataForAuth(t, db, tenantID2)

	// 创建相同账号的用户（跨机构）
	userAccount := "test-staff-multi-tenant"
	password := "test-password-123"
	createTestUserForAuth(t, db, tenantID1, userAccount, password, "", "", "Manager")
	createTestUserForAuth(t, db, tenantID2, userAccount, password, "", "", "Manager")

	// 测试搜索机构（应该返回多个机构）
	req := SearchInstitutionsRequest{
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		UserType:     "staff",
	}

	resp, err := authService.SearchInstitutions(ctx, req)
	if err != nil {
		t.Fatalf("SearchInstitutions failed: %v", err)
	}

	if resp == nil {
		t.Fatal("SearchInstitutions returned nil response")
	}

	if len(resp.Institutions) < 2 {
		t.Fatalf("SearchInstitutions should return at least 2 institutions, got %d", len(resp.Institutions))
	}

	t.Logf("✅ SearchInstitutions multiple tenants test passed: found %d institutions", len(resp.Institutions))
}

func TestAuthService_Login_MultipleTenants_ShouldFail(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建两个测试租户
	tenantID1 := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID1)

	tenantID2 := "00000000-0000-0000-0000-000000000998"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID2, "Test Auth Tenant 2", "test-auth-2.local",
	)
	if err != nil {
		t.Fatalf("Failed to create second test tenant: %v", err)
	}
	defer cleanupTestDataForAuth(t, db, tenantID2)

	// 创建相同账号的用户（跨机构）
	userAccount := "test-staff-multi-tenant-login"
	password := "test-password-123"
	createTestUserForAuth(t, db, tenantID1, userAccount, password, "", "", "Manager")
	createTestUserForAuth(t, db, tenantID2, userAccount, password, "", "", "Manager")

	// 测试登录（不提供 tenant_id，应该失败因为匹配到多个机构）
	req := LoginRequest{
		TenantID:     "", // 不提供 tenant_id
		UserType:     "staff",
		AccountHash:  hashAccount(userAccount),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err = authService.Login(ctx, req)
	if err == nil {
		t.Fatal("Login should fail when multiple tenants match")
	}

	if err.Error() != "Multiple institutions found, please select one" {
		t.Fatalf("Login returned wrong error: expected 'Multiple institutions found, please select one', got '%s'", err.Error())
	}

	t.Logf("✅ Login multiple tenants test passed")
}

// createTestResidentForAuth 创建测试住户
func createTestResidentForAuth(t *testing.T, db *sql.DB, tenantID, unitID, residentAccount, password, email, phone string) string {
	// 计算 hash
	accountHash := sha256.Sum256([]byte(residentAccount))
	passwordHash := sha256.Sum256([]byte(password))
	var emailHash, phoneHash []byte
	if email != "" {
		eh := sha256.Sum256([]byte(email))
		emailHash = eh[:]
	}
	if phone != "" {
		ph := sha256.Sum256([]byte(phone))
		phoneHash = ph[:]
	}

	var residentID string
	err := db.QueryRow(
		`INSERT INTO residents (tenant_id, resident_account, resident_account_hash, password_hash, email_hash, phone_hash, nickname, role, status, unit_id, can_view_status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'Resident', 'active', $8, true)
		 ON CONFLICT (tenant_id, resident_account) DO UPDATE SET
		   resident_account_hash = EXCLUDED.resident_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   email_hash = EXCLUDED.email_hash,
		   phone_hash = EXCLUDED.phone_hash,
		   nickname = EXCLUDED.nickname,
		   status = 'active',
		   can_view_status = true
		 RETURNING resident_id::text`,
		tenantID, residentAccount, accountHash[:], passwordHash[:], emailHash, phoneHash, residentAccount, unitID,
	).Scan(&residentID)
	if err != nil {
		t.Fatalf("Failed to create test resident: %v", err)
	}
	return residentID
}

// createTestUnitForAuth 创建测试 unit（resident 需要 unit_id）
func createTestUnitForAuth(t *testing.T, db *sql.DB, tenantID string) string {
	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// 创建测试unit
	unitID := "00000000-0000-0000-0000-000000000996"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver",
	)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	return unitID
}

// cleanupTestDataForAuth 清理测试数据（更新）
func cleanupTestDataForAuth(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

func TestAuthService_Login_Resident_Success(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	unitID := createTestUnitForAuth(t, db, tenantID)

	residentAccount := "test-resident-001"
	password := "test-password-123"
	email := "test-resident-001@test.com"
	phone := "820101001"

	residentID := createTestResidentForAuth(t, db, tenantID, unitID, residentAccount, password, email, phone)

	// 测试登录（使用 resident_account）
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(residentAccount),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Login returned nil response")
	}

	if resp.UserID != residentID {
		t.Fatalf("Login returned wrong user_id: expected '%s', got '%s'", residentID, resp.UserID)
	}

	if resp.UserType != "resident" {
		t.Fatalf("Login returned wrong user_type: expected 'resident', got '%s'", resp.UserType)
	}

	if resp.UserAccount != residentAccount {
		t.Fatalf("Login returned wrong user_account: expected '%s', got '%s'", residentAccount, resp.UserAccount)
	}

	if resp.Role != "Resident" {
		t.Fatalf("Login returned wrong role: expected 'Resident', got '%s'", resp.Role)
	}

	t.Logf("✅ Login Resident (resident_account) test passed: userID=%s, userAccount=%s", resp.UserID, resp.UserAccount)

	// 测试登录（使用 email）
	req2 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(email),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp2, err := authService.Login(ctx, req2)
	if err != nil {
		t.Fatalf("Login with email failed: %v", err)
	}

	if resp2.UserID != residentID {
		t.Fatalf("Login with email returned different user_id: expected '%s', got '%s'", residentID, resp2.UserID)
	}

	t.Logf("✅ Login Resident (email) test passed: userID=%s", resp2.UserID)

	// 测试登录（使用 phone）
	req3 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(phone),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp3, err := authService.Login(ctx, req3)
	if err != nil {
		t.Fatalf("Login with phone failed: %v", err)
	}

	if resp3.UserID != residentID {
		t.Fatalf("Login with phone returned different user_id: expected '%s', got '%s'", residentID, resp3.UserID)
	}

	t.Logf("✅ Login Resident (phone) test passed: userID=%s", resp3.UserID)
}

func TestAuthService_Login_ResidentContact_Success(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	unitID := createTestUnitForAuth(t, db, tenantID)

	residentAccount := "test-resident-contact"
	password := "test-password-123"
	residentID := createTestResidentForAuth(t, db, tenantID, unitID, residentAccount, password, "", "")

	// 创建 resident_contact
	email := "test-contact@test.com"
	phone := "820101002"
	emailHash := sha256.Sum256([]byte(email))
	phoneHash := sha256.Sum256([]byte(phone))
	passwordHash := sha256.Sum256([]byte(password))

	var contactID string
	err := db.QueryRow(
		`INSERT INTO resident_contacts (tenant_id, resident_id, slot, is_enabled, role, email_hash, phone_hash, password_hash, contact_first_name, contact_last_name)
		 VALUES ($1, $2, 'A', true, 'Family', $3, $4, $5, 'Test', 'Contact')
		 ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
		   email_hash = EXCLUDED.email_hash,
		   phone_hash = EXCLUDED.phone_hash,
		   password_hash = EXCLUDED.password_hash,
		   is_enabled = true
		 RETURNING contact_id::text`,
		tenantID, residentID, emailHash[:], phoneHash[:], passwordHash[:],
	).Scan(&contactID)
	if err != nil {
		t.Fatalf("Failed to create test resident contact: %v", err)
	}

	// 测试登录（使用 email）
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(email),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Login returned nil response")
	}

	if resp.UserID != contactID {
		t.Fatalf("Login returned wrong user_id: expected '%s', got '%s'", contactID, resp.UserID)
	}

	if resp.UserType != "resident" {
		t.Fatalf("Login returned wrong user_type: expected 'resident', got '%s'", resp.UserType)
	}

	if resp.UserAccount != contactID {
		t.Fatalf("Login returned wrong user_account: expected '%s', got '%s'", contactID, resp.UserAccount)
	}

	if resp.Role != "Family" {
		t.Fatalf("Login returned wrong role: expected 'Family', got '%s'", resp.Role)
	}

	t.Logf("✅ Login ResidentContact (email) test passed: userID=%s, userAccount=%s", resp.UserID, resp.UserAccount)

	// 测试登录（使用 phone）
	req2 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(phone),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	resp2, err := authService.Login(ctx, req2)
	if err != nil {
		t.Fatalf("Login with phone failed: %v", err)
	}

	if resp2.UserID != contactID {
		t.Fatalf("Login with phone returned different user_id: expected '%s', got '%s'", contactID, resp2.UserID)
	}

	t.Logf("✅ Login ResidentContact (phone) test passed: userID=%s", resp2.UserID)
}

func TestAuthService_Login_ResidentContact_NotEnabled(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	unitID := createTestUnitForAuth(t, db, tenantID)

	residentAccount := "test-resident-contact-disabled"
	password := "test-password-123"
	residentID := createTestResidentForAuth(t, db, tenantID, unitID, residentAccount, password, "", "")

	// 创建未启用的 resident_contact
	email := "test-contact-disabled@test.com"
	emailHash := sha256.Sum256([]byte(email))
	passwordHash := sha256.Sum256([]byte(password))

	var contactID string
	err := db.QueryRow(
		`INSERT INTO resident_contacts (tenant_id, resident_id, slot, is_enabled, role, email_hash, password_hash, contact_first_name, contact_last_name)
		 VALUES ($1, $2, 'A', false, 'Family', $3, $4, 'Test', 'Contact')
		 ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
		   email_hash = EXCLUDED.email_hash,
		   password_hash = EXCLUDED.password_hash,
		   is_enabled = false
		 RETURNING contact_id::text`,
		tenantID, residentID, emailHash[:], passwordHash[:],
	).Scan(&contactID)
	if err != nil {
		t.Fatalf("Failed to create test resident contact: %v", err)
	}

	// 测试登录未启用的联系人（应该失败）
	req := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  hashAccount(email),
		PasswordHash: hashPassword(password),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}

	_, err = authService.Login(ctx, req)
	if err == nil {
		t.Fatal("Login should fail for disabled resident contact")
	}

	if err.Error() != "user is not active" {
		t.Fatalf("Login returned wrong error: expected 'user is not active', got '%s'", err.Error())
	}

	t.Logf("✅ Login ResidentContact not enabled test passed")
}

func TestAuthService_SearchInstitutions_Resident(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, nil, getTestLogger())

	// 创建测试数据
	tenantID := createTestTenantForAuth(t, db)
	defer cleanupTestDataForAuth(t, db, tenantID)

	unitID := createTestUnitForAuth(t, db, tenantID)

	residentAccount := "test-resident-search"
	password := "test-password-123"
	createTestResidentForAuth(t, db, tenantID, unitID, residentAccount, password, "", "")

	// 测试搜索机构
	req := SearchInstitutionsRequest{
		AccountHash:  hashAccount(residentAccount),
		PasswordHash: hashPassword(password),
		UserType:     "resident",
	}

	resp, err := authService.SearchInstitutions(ctx, req)
	if err != nil {
		t.Fatalf("SearchInstitutions failed: %v", err)
	}

	if resp == nil {
		t.Fatal("SearchInstitutions returned nil response")
	}

	if len(resp.Institutions) == 0 {
		t.Fatal("SearchInstitutions should return at least one institution")
	}

	found := false
	for _, inst := range resp.Institutions {
		if inst.ID == tenantID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("SearchInstitutions did not return expected tenant_id: %s", tenantID)
	}

	t.Logf("✅ SearchInstitutions Resident test passed: found %d institutions", len(resp.Institutions))
}

