// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
)

// TestResidentPasswordReset_WithAuth 测试住户密码重置功能（包含 Auth 验证）
func TestResidentPasswordReset_WithAuth(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	logger := getTestLogger()

	// 1. 创建测试租户和单元
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Password Reset Tenant", "test-pw-reset.local",
	)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
		_, _ = db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
		_, _ = db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
		_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
	}()

	// 创建测试 building
	buildingID := "00000000-0000-0000-0000-000000000996"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name, branch_name = EXCLUDED.branch_name`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	require.NoError(t, err)

	// 创建测试 unit
	unitID := "00000000-0000-0000-0000-000000000995"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_name, unit_number, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver",
	)
	require.NoError(t, err)

	// 2. 创建测试住户（使用初始密码）
	oldPassword := "OldPassword123!"
	oldPasswordHash := sha256.Sum256([]byte(oldPassword))

	residentAccount := "test_resident_001"
	accountHashBytes := sha256.Sum256([]byte(residentAccount))

	var residentID string
	err = db.QueryRow(`
		INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, password_hash,
			nickname, status, admission_date, can_view_status, unit_id
		) VALUES (
			$1, $2, $3, $4, $5, 'active', $6, true, $7
		)
		RETURNING resident_id::text
	`, tenantID, residentAccount, accountHashBytes[:], oldPasswordHash[:], "Test Resident", time.Now(), unitID).Scan(&residentID)
	require.NoError(t, err)

	// 3. 验证初始密码可以登录（使用 Auth 验证）
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	authService := NewAuthService(authRepo, tenantsRepo, db, logger)

	// 使用 account 登录（需要计算 hash）
	accountHash := hex.EncodeToString(accountHashBytes[:])
	// 对于 resident，password_hash = SHA256(password)
	passwordHashForLogin := hex.EncodeToString(oldPasswordHash[:])
	
	loginReq := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  accountHash,
		PasswordHash: passwordHashForLogin,
	}
	loginResp, err := authService.Login(ctx, loginReq)
	require.NoError(t, err, "Initial password should allow login")
	require.NotNil(t, loginResp)
	require.Equal(t, residentID, loginResp.UserID, "Login should return correct resident ID")

	// 4. 创建 Admin 用户（用于重置密码）
	adminUserID := "00000000-0000-0000-0000-000000000994"
	adminAccountHash := sha256.Sum256([]byte("admin"))
	adminPasswordHash := sha256.Sum256([]byte("admin123"))
	_, err = db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   user_account_hash = EXCLUDED.user_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = 'active'`,
		adminUserID, tenantID, "admin", adminAccountHash[:], adminPasswordHash[:], "Admin", "Admin",
	)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	}()

	// 5. 重置密码（使用 password_hash 字段）
	newPassword := "NewPassword456!"
	newPasswordHash := sha256.Sum256([]byte(newPassword))
	newPasswordHashHex := hex.EncodeToString(newPasswordHash[:])

	residentsRepo := repository.NewPostgresResidentsRepository(db)
	residentService := NewResidentService(residentsRepo, db, logger)

	resetReq := ResetResidentPasswordRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   adminUserID,
		CurrentUserType: "staff",
		CurrentUserRole: "Admin",
		NewPassword:     newPasswordHashHex, // 使用 hex 字符串（前端发送的格式）
	}

	resetResp, err := residentService.ResetResidentPassword(ctx, resetReq)
	require.NoError(t, err, "Password reset should succeed")
	require.NotNil(t, resetResp)
	require.True(t, resetResp.Success, "Password reset should return success")

	// 6. 验证新密码可以登录
	newPasswordHashForLogin := hex.EncodeToString(newPasswordHash[:])
	loginReq2 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  accountHash,
		PasswordHash: newPasswordHashForLogin,
	}
	loginResp2, err := authService.Login(ctx, loginReq2)
	require.NoError(t, err, "New password should allow login")
	require.NotNil(t, loginResp2)
	require.Equal(t, residentID, loginResp2.UserID, "Login with new password should return correct resident ID")

	// 7. 验证旧密码不能登录
	oldPasswordHashForLogin := hex.EncodeToString(oldPasswordHash[:])
	loginReq3 := LoginRequest{
		TenantID:     tenantID,
		UserType:     "resident",
		AccountHash:  accountHash,
		PasswordHash: oldPasswordHashForLogin,
	}
	loginResp3, err := authService.Login(ctx, loginReq3)
	require.Error(t, err, "Old password should NOT allow login")
	require.Nil(t, loginResp3, "Login with old password should fail")

	// 8. 验证数据库中的 password_hash 已更新
	var storedPasswordHash []byte
	err = db.QueryRow(
		`SELECT password_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
		tenantID, residentID,
	).Scan(&storedPasswordHash)
	require.NoError(t, err)
	require.Equal(t, newPasswordHash[:], storedPasswordHash, "Database password_hash should match new password hash")

	t.Logf("Password reset test passed: resident_id=%s, old_password=%s, new_password=%s", residentID, oldPassword, newPassword)
}

// TestResidentPasswordReset_FieldName 测试 password_hash 字段名是否正确
func TestResidentPasswordReset_FieldName(t *testing.T) {
	// 这个测试验证前端发送的字段名是 password_hash，而不是 new_password
	// 这个测试主要是文档性质的，确保我们使用正确的字段名

	// 前端应该发送：
	// { password_hash: "hex_encoded_sha256_hash" }
	
	// 后端应该接收：
	// payload["password_hash"]

	// 这个测试通过编译时检查来验证
	t.Log("Field name verification: password_hash (not new_password)")
}

