// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// setupTestDBForUser 设置测试数据库
func setupTestDBForUser(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// createTestTenantForUser 创建测试租户
func createTestTenantForUser(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000998"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test User Tenant", "test-user.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestUserForUser 创建测试用户（用于测试）
func createTestUserForUser(t *testing.T, db *sql.DB, tenantID, userID, userAccount, password, email, phone, role, branchTag string) string {
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
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, email, email_hash, phone, phone_hash, nickname, role, status, branch_tag)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active', $12)
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   user_account_hash = EXCLUDED.user_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   email = EXCLUDED.email,
		   email_hash = EXCLUDED.email_hash,
		   phone = EXCLUDED.phone,
		   phone_hash = EXCLUDED.phone_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = 'active',
		   branch_tag = EXCLUDED.branch_tag`,
		userID, tenantID, userAccount, accountHash[:], passwordHash[:], email, emailHash, phone, phoneHash, userAccount, role, branchTag,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// cleanupTestDataForUser 清理测试数据
func cleanupTestDataForUser(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestLoggerForUser 获取测试日志记录器
func getTestLoggerForUser() *zap.Logger {
	return getTestLogger()
}

// ============================================
// ListUsers 测试
// ============================================

// TestUserService_ListUsers_Success 测试查询用户列表成功
func TestUserService_ListUsers_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	// 创建 Service
	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	// 创建测试用户
	userID1 := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "user1", "password1", "user1@test.com", "1234567890", "Admin", "BRANCH-1")
	userID2 := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "user2", "password2", "user2@test.com", "0987654321", "Manager", "BRANCH-1")

	// 测试查询所有用户（Admin 角色，无权限限制）
	req := ListUsersRequest{
		TenantID:      tenantID,
		CurrentUserID: userID1, // Admin 用户
		Page:          1,
		Size:          20,
	}

	resp, err := userService.ListUsers(context.Background(), req)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if resp.Total < 2 {
		t.Fatalf("Expected at least 2 users, got %d", resp.Total)
	}

	// 验证返回的用户
	found1, found2 := false, false
	for _, u := range resp.Items {
		if u.UserID == userID1 {
			found1 = true
			if u.UserAccount != "user1" {
				t.Errorf("Expected user_account 'user1', got %s", u.UserAccount)
			}
		}
		if u.UserID == userID2 {
			found2 = true
			if u.UserAccount != "user2" {
				t.Errorf("Expected user_account 'user2', got %s", u.UserAccount)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("Expected to find both users, found1=%v, found2=%v", found1, found2)
	}
}

// TestUserService_ListUsers_WithSearch 测试搜索功能
func TestUserService_ListUsers_WithSearch(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	userID1 := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password1", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "manager", "password2", "manager@test.com", "0987654321", "Manager", "BRANCH-1")

	// 测试搜索
	req := ListUsersRequest{
		TenantID:      tenantID,
		CurrentUserID: userID1,
		Search:        "admin",
		Page:          1,
		Size:          20,
	}

	resp, err := userService.ListUsers(context.Background(), req)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	// 应该只找到 admin 用户
	found := false
	for _, u := range resp.Items {
		if u.UserAccount == "admin" {
			found = true
		}
		if u.UserAccount == "manager" {
			t.Errorf("Should not find 'manager' when searching for 'admin'")
		}
	}

	if !found {
		t.Errorf("Expected to find 'admin' user")
	}
}

// ============================================
// GetUser 测试
// ============================================

// TestUserService_GetUser_Success 测试查询用户详情成功
func TestUserService_GetUser_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "testuser", "password", "test@test.com", "1234567890", "Admin", "BRANCH-1")

	req := GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: userID, // 查询自己
	}

	resp, err := userService.GetUser(context.Background(), req)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if resp.User == nil {
		t.Fatalf("Expected user, got nil")
	}

	if resp.User.UserID != userID {
		t.Errorf("Expected user_id %s, got %s", userID, resp.User.UserID)
	}

	if resp.User.UserAccount != "testuser" {
		t.Errorf("Expected user_account 'testuser', got %s", resp.User.UserAccount)
	}

	if resp.User.Email != "test@test.com" {
		t.Errorf("Expected email 'test@test.com', got %s", resp.User.Email)
	}
}

// ============================================
// CreateUser 测试
// ============================================

// TestUserService_CreateUser_Success 测试创建用户成功
func TestUserService_CreateUser_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	// 创建 Admin 用户作为当前用户
	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")

	req := CreateUserRequest{
		TenantID:      tenantID,
		CurrentUserID: adminID,
		UserAccount:   "newuser",
		Password:      "newpassword",
		Role:          "Manager",
		Nickname:      "New User",
		Email:         "newuser@test.com",
		Phone:         "9876543210",
		Status:        "active",
		BranchTag:     "BRANCH-1",
	}

	resp, err := userService.CreateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if resp.UserID == "" {
		t.Fatalf("Expected user_id, got empty string")
	}

	// 验证用户已创建
	getReq := GetUserRequest{
		TenantID:      tenantID,
		UserID:        resp.UserID,
		CurrentUserID: adminID,
	}

	getResp, err := userService.GetUser(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if getResp.User.UserAccount != "newuser" {
		t.Errorf("Expected user_account 'newuser', got %s", getResp.User.UserAccount)
	}

	if getResp.User.Role != "Manager" {
		t.Errorf("Expected role 'Manager', got %s", getResp.User.Role)
	}
}

// TestUserService_CreateUser_DuplicateEmail 测试重复邮箱
func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")

	// 创建第一个用户
	req1 := CreateUserRequest{
		TenantID:      tenantID,
		CurrentUserID: adminID,
		UserAccount:   "user1",
		Password:      "password1",
		Role:          "Manager",
		Email:         "duplicate@test.com",
	}

	_, err := userService.CreateUser(context.Background(), req1)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 尝试创建第二个用户，使用相同的邮箱
	req2 := CreateUserRequest{
		TenantID:      tenantID,
		CurrentUserID: adminID,
		UserAccount:   "user2",
		Password:      "password2",
		Role:          "Manager",
		Email:         "duplicate@test.com",
	}

	_, err = userService.CreateUser(context.Background(), req2)
	if err == nil {
		t.Fatalf("Expected error for duplicate email, got nil")
	}
}

// ============================================
// UpdateUser 测试
// ============================================

// TestUserService_UpdateUser_Success 测试更新用户成功
func TestUserService_UpdateUser_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	newNickname := "Updated Nickname"
	req := UpdateUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
		Nickname:      &newNickname,
	}

	resp, err := userService.UpdateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证更新
	getReq := GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
	}

	getResp, err := userService.GetUser(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if getResp.User.Nickname != "Updated Nickname" {
		t.Errorf("Expected nickname 'Updated Nickname', got %s", getResp.User.Nickname)
	}
}

// TestUserService_UpdateUser_DeleteEmail 测试删除邮箱
func TestUserService_UpdateUser_DeleteEmail(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	// 删除邮箱（设置为空字符串）
	emptyEmail := ""
	req := UpdateUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
		Email:         &emptyEmail,
	}

	resp, err := userService.UpdateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证邮箱已删除
	getReq := GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
	}

	getResp, err := userService.GetUser(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if getResp.User.Email != "" {
		t.Errorf("Expected email to be empty, got %s", getResp.User.Email)
	}
}

// ============================================
// DeleteUser 测试
// ============================================

// TestUserService_DeleteUser_Success 测试删除用户成功（软删除）
func TestUserService_DeleteUser_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	req := DeleteUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
	}

	resp, err := userService.DeleteUser(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证用户状态已更新为 'left'
	getReq := GetUserRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
	}

	getResp, err := userService.GetUser(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if getResp.User.Status != "left" {
		t.Errorf("Expected status 'left', got %s", getResp.User.Status)
	}
}

// ============================================
// ResetPassword 测试
// ============================================

// TestUserService_ResetPassword_Success 测试重置密码成功
func TestUserService_ResetPassword_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	req := UserResetPasswordRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
		NewPassword:   "newpassword123",
	}

	resp, err := userService.ResetPassword(context.Background(), req)
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证密码已更新（通过 Repository 直接查询）
	user, err := usersRepo.GetUser(context.Background(), tenantID, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	// 验证新密码的 hash
	newPasswordHash := sha256.Sum256([]byte("newpassword123"))
	if len(user.PasswordHash) == 0 {
		t.Errorf("Expected password_hash to be set")
	}

	// 验证 hash 匹配
	if len(user.PasswordHash) != len(newPasswordHash[:]) {
		t.Errorf("Password hash length mismatch")
	}
}

// ============================================
// ResetPIN 测试
// ============================================

// TestUserService_ResetPIN_Success 测试重置 PIN 成功
func TestUserService_ResetPIN_Success(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	req := UserResetPINRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
		NewPIN:        "1234",
	}

	resp, err := userService.ResetPIN(context.Background(), req)
	if err != nil {
		t.Fatalf("ResetPIN failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证 PIN 已更新
	user, err := usersRepo.GetUser(context.Background(), tenantID, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if len(user.PinHash) == 0 {
		t.Errorf("Expected pin_hash to be set")
	}
}

// TestUserService_ResetPIN_InvalidPIN 测试无效 PIN
func TestUserService_ResetPIN_InvalidPIN(t *testing.T) {
	db := setupTestDBForUser(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForUser(t, db)
	defer cleanupTestDataForUser(t, db, tenantID)

	usersRepo := repository.NewPostgresUsersRepository(db)
	userService := NewUserService(usersRepo, getTestLoggerForUser())

	adminID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000001", tenantID, "admin", "password", "admin@test.com", "1234567890", "Admin", "BRANCH-1")
	userID := createTestUserForUser(t, db, "00000000-0000-0000-0000-000000000002", tenantID, "testuser", "password", "test@test.com", "1234567890", "Manager", "BRANCH-1")

	// 测试 PIN 长度不正确
	req := UserResetPINRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CurrentUserID: adminID,
		NewPIN:        "123", // 只有 3 位
	}

	_, err := userService.ResetPIN(context.Background(), req)
	if err == nil {
		t.Fatalf("Expected error for invalid PIN length, got nil")
	}

	// 测试 PIN 包含非数字字符
	req.NewPIN = "12ab"
	_, err = userService.ResetPIN(context.Background(), req)
	if err == nil {
		t.Fatalf("Expected error for invalid PIN format, got nil")
	}
}

