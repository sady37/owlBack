// +build integration

package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"owl-common/database"
	"wisefido-data/internal/config"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *sql.DB {
	cfg := config.Load()
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	return db
}

// createTestTenantForHandler 创建测试租户
func createTestTenantForHandler(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000998"
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Handler Tenant", "test-handler.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestUserForHandler 创建测试用户
func createTestUserForHandler(t *testing.T, db *sql.DB, tenantID, userAccount, password, email, phone, role string) {
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

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, email_hash, phone_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET
		   user_account_hash = EXCLUDED.user_account_hash,
		   password_hash = EXCLUDED.password_hash,
		   email_hash = EXCLUDED.email_hash,
		   phone_hash = EXCLUDED.phone_hash,
		   nickname = EXCLUDED.nickname,
		   role = EXCLUDED.role,
		   status = EXCLUDED.status`,
		tenantID, userAccount, accountHash[:], passwordHash[:], emailHash, phoneHash, "Test User", role,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

// cleanupTestDataForHandler 清理测试数据
func cleanupTestDataForHandler(t *testing.T, db *sql.DB, tenantID string) {
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
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

// TestAuthHandler_Login_Success 测试登录成功
func TestAuthHandler_Login_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)

	userAccount := "testuser"
	password := "testpass123"
	createTestUserForHandler(t, db, tenantID, userAccount, password, "test@example.com", "", "Manager")

	// 创建 Handler
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	logger := zap.NewNop()
	authService := service.NewAuthService(authRepo, tenantsRepo, nil, logger)
	authHandler := NewAuthHandler(authService, logger)

	// 准备请求
	accountHash := hashAccount(userAccount)
	passwordHash := hashPassword(password)
	reqBody := map[string]any{
		"tenant_id":    tenantID,
		"userType":     "staff",
		"accountHash":  accountHash,
		"passwordHash": passwordHash,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	authHandler.Login(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
		Result  struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			UserID       string `json:"userId"`
			UserAccount  string `json:"user_account"`
			UserType     string `json:"userType"`
			Role         string `json:"role"`
			NickName     string `json:"nickName"`
			TenantID     string `json:"tenant_id"`
			TenantName   string `json:"tenant_name"`
			Domain       string `json:"domain"`
			HomePath     string `json:"homePath"`
		} `json:"result"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 验证响应格式
	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d", result.Code)
	}
	if result.Type != "success" {
		t.Errorf("Expected type 'success', got '%s'", result.Type)
	}
	if result.Result.UserAccount != userAccount {
		t.Errorf("Expected user_account '%s', got '%s'", userAccount, result.Result.UserAccount)
	}
	if result.Result.UserType != "staff" {
		t.Errorf("Expected userType 'staff', got '%s'", result.Result.UserType)
	}
	if result.Result.TenantID != tenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", tenantID, result.Result.TenantID)
	}
}

// TestAuthHandler_Login_MissingCredentials 测试缺少凭证
func TestAuthHandler_Login_MissingCredentials(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	logger := zap.NewNop()
	authService := service.NewAuthService(authRepo, tenantsRepo, nil, logger)
	authHandler := NewAuthHandler(authService, logger)

	// 测试缺少 accountHash
	reqBody := map[string]any{
		"tenant_id":    "00000000-0000-0000-0000-000000000001",
		"userType":     "staff",
		"passwordHash": hashPassword("testpass123"),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.Code != -1 {
		t.Errorf("Expected code -1, got %d", result.Code)
	}
	if result.Type != "error" {
		t.Errorf("Expected type 'error', got '%s'", result.Type)
	}
	if result.Message == "" {
		t.Error("Expected error message, got empty")
	}
}

// TestAuthHandler_SearchInstitutions_Success 测试搜索机构成功
func TestAuthHandler_SearchInstitutions_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := createTestTenantForHandler(t, db)
	defer cleanupTestDataForHandler(t, db, tenantID)

	userAccount := "testuser"
	password := "testpass123"
	createTestUserForHandler(t, db, tenantID, userAccount, password, "test@example.com", "", "Manager")

	// 创建 Handler
	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	logger := zap.NewNop()
	authService := service.NewAuthService(authRepo, tenantsRepo, nil, logger)
	authHandler := NewAuthHandler(authService, logger)

	// 准备请求
	accountHash := hashAccount(userAccount)
	passwordHash := hashPassword(password)
	req := httptest.NewRequest(http.MethodGet, "/auth/api/v1/institutions/search?accountHash="+accountHash+"&passwordHash="+passwordHash+"&userType=staff", nil)
	w := httptest.NewRecorder()

	// 执行请求
	authHandler.SearchInstitutions(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int      `json:"code"`
		Type    string   `json:"type"`
		Message string   `json:"message"`
		Result  []any    `json:"result"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 验证响应格式
	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d", result.Code)
	}
	if result.Type != "success" {
		t.Errorf("Expected type 'success', got '%s'", result.Type)
	}
	if len(result.Result) == 0 {
		t.Error("Expected at least one institution, got empty")
	}

	// 验证机构信息
	if len(result.Result) > 0 {
		inst := result.Result[0].(map[string]any)
		if inst["id"] == nil {
			t.Error("Expected institution id, got nil")
		}
		if inst["name"] == nil {
			t.Error("Expected institution name, got nil")
		}
		if inst["accountType"] == nil {
			t.Error("Expected accountType, got nil")
		}
	}
}

// TestAuthHandler_SearchInstitutions_NoMatch 测试无匹配
func TestAuthHandler_SearchInstitutions_NoMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	logger := zap.NewNop()
	authService := service.NewAuthService(authRepo, tenantsRepo, nil, logger)
	authHandler := NewAuthHandler(authService, logger)

	// 使用不存在的账号
	accountHash := hashAccount("nonexistent")
	passwordHash := hashPassword("wrongpass")
	req := httptest.NewRequest(http.MethodGet, "/auth/api/v1/institutions/search?accountHash="+accountHash+"&passwordHash="+passwordHash+"&userType=staff", nil)
	w := httptest.NewRecorder()

	authHandler.SearchInstitutions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result struct {
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
		Result  []any  `json:"result"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.Code != 2000 {
		t.Errorf("Expected code 2000, got %d", result.Code)
	}
	if len(result.Result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result.Result))
	}
}

// TestAuthHandler_ServeHTTP_Routing 测试路由分发
func TestAuthHandler_ServeHTTP_Routing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authRepo := repository.NewPostgresAuthRepository(db)
	tenantsRepo := repository.NewPostgresTenantsRepository(db)
	logger := zap.NewNop()
	authService := service.NewAuthService(authRepo, tenantsRepo, nil, logger)
	authHandler := NewAuthHandler(authService, logger)

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{"POST /auth/api/v1/login", http.MethodPost, "/auth/api/v1/login", http.StatusOK},
		{"GET /auth/api/v1/institutions/search", http.MethodGet, "/auth/api/v1/institutions/search", http.StatusOK},
		{"POST /auth/api/v1/forgot-password/send-code", http.MethodPost, "/auth/api/v1/forgot-password/send-code", http.StatusOK},
		{"POST /auth/api/v1/forgot-password/verify-code", http.MethodPost, "/auth/api/v1/forgot-password/verify-code", http.StatusOK},
		{"POST /auth/api/v1/forgot-password/reset", http.MethodPost, "/auth/api/v1/forgot-password/reset", http.StatusOK},
		{"GET /auth/api/v1/login (wrong method)", http.MethodGet, "/auth/api/v1/login", http.StatusMethodNotAllowed},
		{"POST /auth/api/v1/institutions/search (wrong method)", http.MethodPost, "/auth/api/v1/institutions/search", http.StatusMethodNotAllowed},
		{"Unknown path", http.MethodGet, "/auth/api/v1/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			authHandler.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

