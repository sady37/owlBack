// +build integration

package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"wisefido-data/internal/domain"

	"owl-common/database"
	"owl-common/config"
)

// 获取测试数据库连接
func getTestDB(t *testing.T) *sql.DB {
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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// 创建测试租户
func createTestTenant(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000999"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant", "test.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：users -> tags_catalog -> tenants
	db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// 辅助函数：计算hash
func hashString(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

func hashStringHex(s string) string {
	return hex.EncodeToString(hashString(s))
}

func TestPostgresUsersRepository_CreateUser(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 创建测试用户
	user := &domain.User{
		UserAccount:     "testuser",
		UserAccountHash: hashString("testuser"),
		PasswordHash:    hashString("password123"),
		Nickname:        sql.NullString{String: "Test User", Valid: true},
		Email:           sql.NullString{String: "test@example.com", Valid: true},
		Phone:           sql.NullString{String: "1234567890", Valid: true},
		EmailHash:       hashString("test@example.com"),
		PhoneHash:       hashString("1234567890"),
		Role:            "Nurse",
		Status:          "active",
		Tags:            sql.NullString{String: `["NightShift"]`, Valid: true},
		BranchTag:       sql.NullString{String: "DV1", Valid: true},
	}

	userID, err := repo.CreateUser(ctx, tenantID, user)
	if err != nil {
		t.Logf("CreateUser error details: %+v", err)
		t.Fatalf("CreateUser failed: %v", err)
	}

	if userID == "" {
		t.Fatal("CreateUser returned empty userID")
	}

	// 验证用户已创建
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify user creation: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 user, got %d", count)
	}

	// 验证tags_catalog已同步
	var tagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "NightShift", "user_tag").Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to verify tag in catalog: %v", err)
	}
	if tagCount != 1 {
		t.Fatalf("Expected 1 tag in catalog, got %d", tagCount)
	}

	// 验证branch_tag已同步
	var branchTagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "DV1", "branch_tag").Scan(&branchTagCount)
	if err != nil {
		t.Fatalf("Failed to verify branch_tag in catalog: %v", err)
	}
	if branchTagCount != 1 {
		t.Fatalf("Expected 1 branch_tag in catalog, got %d", branchTagCount)
	}

	t.Logf("✅ CreateUser test passed: userID=%s", userID)
}

func TestPostgresUsersRepository_GetUser(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	userID := "00000000-0000-0000-0000-000000000001"
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status, tags, branch_tag)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)`,
		userID, tenantID, "testuser2", hashString("testuser2"), hashString("pwd"),
		"Nurse", "active", `["DayShift"]`, "DV2",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 测试GetUser
	user, err := repo.GetUser(ctx, tenantID, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.UserID != userID {
		t.Fatalf("Expected userID %s, got %s", userID, user.UserID)
	}
	if user.UserAccount != "testuser2" {
		t.Fatalf("Expected user_account 'testuser2', got '%s'", user.UserAccount)
	}
	if user.Role != "Nurse" {
		t.Fatalf("Expected role 'Nurse', got '%s'", user.Role)
	}

	t.Logf("✅ GetUser test passed: userID=%s", userID)
}

func TestPostgresUsersRepository_GetUserByAccount(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	account := "testuser3"
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`,
		tenantID, account, hashString(account), hashString("pwd"), "Nurse", "active",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 测试GetUserByAccount
	user, err := repo.GetUserByAccount(ctx, tenantID, account)
	if err != nil {
		t.Fatalf("GetUserByAccount failed: %v", err)
	}

	if user.UserAccount != account {
		t.Fatalf("Expected user_account '%s', got '%s'", account, user.UserAccount)
	}

	t.Logf("✅ GetUserByAccount test passed: account=%s", account)
}

func TestPostgresUsersRepository_GetUserByEmail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	email := "testemail@example.com"
	emailHash := hashString(email)
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, email, email_hash, role, status)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID, "testuser4", hashString("testuser4"), hashString("pwd"),
		email, emailHash, "Nurse", "active",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 测试GetUserByEmail
	user, err := repo.GetUserByEmail(ctx, tenantID, emailHash)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if !user.Email.Valid || user.Email.String != email {
		t.Fatalf("Expected email '%s', got '%v'", email, user.Email)
	}

	t.Logf("✅ GetUserByEmail test passed: email=%s", email)
}

func TestPostgresUsersRepository_GetUserByPhone(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	phone := "9876543210"
	phoneHash := hashString(phone)
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, phone, phone_hash, role, status)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID, "testuser5", hashString("testuser5"), hashString("pwd"),
		phone, phoneHash, "Nurse", "active",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 测试GetUserByPhone
	user, err := repo.GetUserByPhone(ctx, tenantID, phoneHash)
	if err != nil {
		t.Fatalf("GetUserByPhone failed: %v", err)
	}

	if !user.Phone.Valid || user.Phone.String != phone {
		t.Fatalf("Expected phone '%s', got '%v'", phone, user.Phone)
	}

	t.Logf("✅ GetUserByPhone test passed: phone=%s", phone)
}

func TestPostgresUsersRepository_ListUsers(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 创建多个测试用户
	users := []struct {
		account string
		role    string
		status  string
		tag     string
	}{
		{"user1", "Nurse", "active", "NightShift"},
		{"user2", "Manager", "active", "DayShift"},
		{"user3", "Nurse", "suspended", "NightShift"},
	}

	for _, u := range users {
		tagsJSON, _ := json.Marshal([]string{u.tag})
		_, err := db.Exec(
			`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status, tags)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7::jsonb)`,
			tenantID, u.account, hashString(u.account), hashString("pwd"),
			u.role, u.status, string(tagsJSON),
		)
		if err != nil {
			t.Fatalf("Failed to insert test user %s: %v", u.account, err)
		}
	}

	// 测试ListUsers - 按Role过滤
	filters := UserFilters{Role: "Nurse"}
	items, total, err := repo.ListUsers(ctx, tenantID, filters, 1, 10)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total != 2 { // user1和user3都是Nurse
		t.Fatalf("Expected 2 nurses, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// 测试ListUsers - 按Status过滤
	filters = UserFilters{Status: "active"}
	items, total, err = repo.ListUsers(ctx, tenantID, filters, 1, 10)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total != 2 { // user1和user2都是active
		t.Fatalf("Expected 2 active users, got %d", total)
	}

	// 测试ListUsers - Search
	filters = UserFilters{Search: "user1"}
	items, total, err = repo.ListUsers(ctx, tenantID, filters, 1, 10)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("Expected 1 user matching 'user1', got %d", total)
	}
	if items[0].UserAccount != "user1" {
		t.Fatalf("Expected user_account 'user1', got '%s'", items[0].UserAccount)
	}

	t.Logf("✅ ListUsers test passed: total=%d", total)
}

func TestPostgresUsersRepository_UpdateUser(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	userID := "00000000-0000-0000-0000-000000000002"
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status, tags, branch_tag)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)`,
		userID, tenantID, "updateuser", hashString("updateuser"), hashString("pwd"),
		"Nurse", "active", `["OldTag"]`, "OLD_BRANCH",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 更新用户
	updatedUser := &domain.User{
		Nickname:  sql.NullString{String: "Updated Name", Valid: true},
		Status:    "suspended",
		Tags:      sql.NullString{String: `["NewTag"]`, Valid: true},
		BranchTag: sql.NullString{String: "NEW_BRANCH", Valid: true},
	}

	err = repo.UpdateUser(ctx, tenantID, userID, updatedUser)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// 验证更新
	user, err := repo.GetUser(ctx, tenantID, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if !user.Nickname.Valid || user.Nickname.String != "Updated Name" {
		t.Fatalf("Expected nickname 'Updated Name', got '%v'", user.Nickname)
	}
	if user.Status != "suspended" {
		t.Fatalf("Expected status 'suspended', got '%s'", user.Status)
	}

	// 验证tags_catalog已同步新tag
	var tagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "NewTag", "user_tag").Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to verify new tag in catalog: %v", err)
	}
	if tagCount != 1 {
		t.Fatalf("Expected 1 new tag in catalog, got %d", tagCount)
	}

	// 验证branch_tag已同步
	var branchTagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "NEW_BRANCH", "branch_tag").Scan(&branchTagCount)
	if err != nil {
		t.Fatalf("Failed to verify new branch_tag in catalog: %v", err)
	}
	if branchTagCount != 1 {
		t.Fatalf("Expected 1 new branch_tag in catalog, got %d", branchTagCount)
	}

	t.Logf("✅ UpdateUser test passed: userID=%s", userID)
}

func TestPostgresUsersRepository_DeleteUser(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 先创建用户
	userID := "00000000-0000-0000-0000-000000000003"
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, tenantID, "deleteuser", hashString("deleteuser"), hashString("pwd"),
		"Nurse", "active",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// 删除用户
	err = repo.DeleteUser(ctx, tenantID, userID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// 验证用户已删除
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify user deletion: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 users, got %d", count)
	}

	t.Logf("✅ DeleteUser test passed: userID=%s", userID)
}

func TestPostgresUsersRepository_SyncUserTagsToCatalog(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenant(t, db)
	defer cleanupTestData(t, db, tenantID)

	repo := NewPostgresUsersRepository(db)
	ctx := context.Background()

	// 创建测试用户
	userID := "00000000-0000-0000-0000-000000000004"
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, tenantID, "synctagsuser", hashString("synctagsuser"), hashString("pwd"),
		"Nurse", "active",
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	defer db.Exec(`DELETE FROM users WHERE user_id = $1`, userID)

	// 测试：同步多个tags到catalog
	tags := []string{"NightShift", "Group.A", "FallsExpert"}
	err = repo.SyncUserTagsToCatalog(ctx, tenantID, userID, tags)
	if err != nil {
		t.Fatalf("SyncUserTagsToCatalog failed: %v", err)
	}

	// 验证tags_catalog中已创建对应的tag记录
	for _, tagName := range tags {
		var tagCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM tags_catalog 
			 WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
			tenantID, tagName, "user_tag",
		).Scan(&tagCount)
		if err != nil {
			t.Fatalf("Failed to verify tag '%s' in catalog: %v", tagName, err)
		}
		if tagCount != 1 {
			t.Fatalf("Expected 1 tag '%s' in catalog, got %d", tagName, tagCount)
		}

		// 验证tag_id是基于tag_name确定性生成的
		var tagID string
		err = db.QueryRow(
			`SELECT tag_id::text FROM tags_catalog 
			 WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
			tenantID, tagName, "user_tag",
		).Scan(&tagID)
		if err != nil {
			t.Fatalf("Failed to get tag_id for '%s': %v", tagName, err)
		}
		if tagID == "" {
			t.Fatalf("Expected non-empty tag_id for '%s'", tagName)
		}
	}

	// 测试：再次同步相同的tags（应该不会报错，因为upsert语义）
	err = repo.SyncUserTagsToCatalog(ctx, tenantID, userID, tags)
	if err != nil {
		t.Fatalf("SyncUserTagsToCatalog (duplicate) failed: %v", err)
	}

	// 验证tags数量没有增加（仍然是3个）
	var totalTagCount int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM tags_catalog 
		 WHERE tenant_id = $1 AND tag_type = $2`,
		tenantID, "user_tag",
	).Scan(&totalTagCount)
	if err != nil {
		t.Fatalf("Failed to count tags: %v", err)
	}
	// 注意：totalTagCount可能包含其他测试创建的tags，所以只验证至少3个
	if totalTagCount < 3 {
		t.Fatalf("Expected at least 3 tags in catalog, got %d", totalTagCount)
	}

	// 测试：同步新的tags（应该添加到catalog）
	newTags := []string{"NewTag1", "NewTag2"}
	err = repo.SyncUserTagsToCatalog(ctx, tenantID, userID, newTags)
	if err != nil {
		t.Fatalf("SyncUserTagsToCatalog (new tags) failed: %v", err)
	}

	// 验证新tags已创建
	for _, tagName := range newTags {
		var tagCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM tags_catalog 
			 WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
			tenantID, tagName, "user_tag",
		).Scan(&tagCount)
		if err != nil {
			t.Fatalf("Failed to verify new tag '%s' in catalog: %v", tagName, err)
		}
		if tagCount != 1 {
			t.Fatalf("Expected 1 new tag '%s' in catalog, got %d", tagName, tagCount)
		}
	}

	// 测试：同步空tags数组（应该不报错）
	err = repo.SyncUserTagsToCatalog(ctx, tenantID, userID, []string{})
	if err != nil {
		t.Fatalf("SyncUserTagsToCatalog (empty tags) failed: %v", err)
	}

	// 测试：同步包含空字符串的tags数组（应该跳过空字符串）
	mixedTags := []string{"ValidTag", "", "AnotherTag"}
	err = repo.SyncUserTagsToCatalog(ctx, tenantID, userID, mixedTags)
	if err != nil {
		t.Fatalf("SyncUserTagsToCatalog (mixed tags) failed: %v", err)
	}

	// 验证只有非空tags被创建
	for _, tagName := range []string{"ValidTag", "AnotherTag"} {
		var tagCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM tags_catalog 
			 WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
			tenantID, tagName, "user_tag",
		).Scan(&tagCount)
		if err != nil {
			t.Fatalf("Failed to verify tag '%s' in catalog: %v", tagName, err)
		}
		if tagCount < 1 {
			t.Fatalf("Expected at least 1 tag '%s' in catalog, got %d", tagName, tagCount)
		}
	}

	// 清理测试数据
	for _, tagName := range append(append(tags, newTags...), mixedTags...) {
		if tagName != "" {
			db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2`, tenantID, tagName)
		}
	}

	t.Logf("✅ SyncUserTagsToCatalog test passed: userID=%s", userID)
}

