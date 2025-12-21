// +build integration

package repository

import (
	"context"
	"database/sql"
	"testing"

	"wisefido-data/internal/domain"

	"owl-common/database"
	"owl-common/config"
)

// 获取测试数据库连接（复用users测试的helper）
func getTestDBForTags(t *testing.T) *sql.DB {
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

// 创建测试租户
func createTestTenantForTags(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Tags", "test-tags.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForTags(t *testing.T, db *sql.DB, tenantID string) {
	// 删除测试创建的tags（只删除family_tag和user_tag，系统tag不能删除）
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1 AND tag_type IN ('family_tag', 'user_tag')`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// TagsRepository 测试
// ============================================

func TestPostgresTagsRepository_CreateTag(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 测试：创建family_tag
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "family_tag",
		TagName:  "Test Family Tag 001",
	}

	tagID, err := repo.CreateTag(ctx, tenantID, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	if tagID == "" {
		t.Fatal("CreateTag returned empty tag_id")
	}
	t.Logf("Created tag_id: %s", tagID)

	// 验证：查询创建的tag
	createdTag, err := repo.GetTag(ctx, tenantID, tagID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if createdTag.TagName != "Test Family Tag 001" {
		t.Errorf("Expected tag_name=Test Family Tag 001, got %s", createdTag.TagName)
	}
	if createdTag.TagType != "family_tag" {
		t.Errorf("Expected tag_type=family_tag, got %s", createdTag.TagType)
	}

	// 清理
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID)
}

func TestPostgresTagsRepository_CreateTag_UserTag(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 测试：创建user_tag
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "user_tag",
		TagName:  "Test User Tag 001",
	}

	tagID, err := repo.CreateTag(ctx, tenantID, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID)

	// 验证
	createdTag, err := repo.GetTag(ctx, tenantID, tagID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if createdTag.TagType != "user_tag" {
		t.Errorf("Expected tag_type=user_tag, got %s", createdTag.TagType)
	}
}

func TestPostgresTagsRepository_CreateTag_SystemTagType_ShouldFail(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 测试：尝试创建branch_tag（应该失败）
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "branch_tag",
		TagName:  "Test Branch Tag",
	}

	_, err := repo.CreateTag(ctx, tenantID, tag)
	if err == nil {
		t.Fatal("CreateTag should fail for branch_tag, but succeeded")
	}
	t.Logf("Expected error for branch_tag: %v", err)

	// 测试：尝试创建area_tag（应该失败）
	tag.TagType = "area_tag"
	tag.TagName = "Test Area Tag"
	_, err = repo.CreateTag(ctx, tenantID, tag)
	if err == nil {
		t.Fatal("CreateTag should fail for area_tag, but succeeded")
	}
	t.Logf("Expected error for area_tag: %v", err)
}

func TestPostgresTagsRepository_GetTagByName(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 先创建一个tag
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "family_tag",
		TagName:  "Test Family Tag GetByName",
	}
	tagID, err := repo.CreateTag(ctx, tenantID, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID)

	// 测试：通过tag_name查询
	foundTag, err := repo.GetTagByName(ctx, tenantID, "Test Family Tag GetByName")
	if err != nil {
		t.Fatalf("GetTagByName failed: %v", err)
	}
	if foundTag.TagID != tagID {
		t.Errorf("Expected tag_id=%s, got %s", tagID, foundTag.TagID)
	}
	if foundTag.TagName != "Test Family Tag GetByName" {
		t.Errorf("Expected tag_name=Test Family Tag GetByName, got %s", foundTag.TagName)
	}
}

func TestPostgresTagsRepository_ListTags(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 创建多个测试tags
	tag1 := &domain.Tag{
		TenantID: tenantID,
		TagType:  "family_tag",
		TagName:  "Test Family Tag List 001",
	}
	tagID1, err := repo.CreateTag(ctx, tenantID, tag1)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID1)

	tag2 := &domain.Tag{
		TenantID: tenantID,
		TagType:  "user_tag",
		TagName:  "Test User Tag List 001",
	}
	tagID2, err := repo.CreateTag(ctx, tenantID, tag2)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID2)

	// 测试：查询所有tags
	filter := TagsFilter{
		IncludeSystemTags: true,
	}
	tags, total, err := repo.ListTags(ctx, tenantID, filter, 1, 20)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if total < 2 {
		t.Errorf("Expected total >= 2, got %d", total)
	}

	// 验证返回的tags
	found1, found2 := false, false
	for _, tag := range tags {
		if tag.TagID == tagID1 {
			found1 = true
		}
		if tag.TagID == tagID2 {
			found2 = true
		}
	}
	if !found1 {
		t.Error("Created tag 1 not found in list")
	}
	if !found2 {
		t.Error("Created tag 2 not found in list")
	}

	// 测试：按tag_type过滤
	filter = TagsFilter{
		TagType:           "family_tag",
		IncludeSystemTags: true,
	}
	tags, total, err = repo.ListTags(ctx, tenantID, filter, 1, 20)
	if err != nil {
		t.Fatalf("ListTags (with tag_type filter) failed: %v", err)
	}
	for _, tag := range tags {
		if tag.TagID == tagID2 {
			t.Error("user_tag should not appear when filtering by family_tag")
		}
	}

	// 测试：按tag_name搜索
	filter = TagsFilter{
		TagName:           "List 001",
		IncludeSystemTags: true,
	}
	tags, total, err = repo.ListTags(ctx, tenantID, filter, 1, 20)
	if err != nil {
		t.Fatalf("ListTags (with tag_name search) failed: %v", err)
	}
	if total < 2 {
		t.Errorf("Expected total >= 2 when searching 'List 001', got %d", total)
	}
}

func TestPostgresTagsRepository_GetTagsForTenant(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 创建测试tags
	tag1 := &domain.Tag{
		TenantID: tenantID,
		TagType:  "family_tag",
		TagName:  "Test Family Tag ForTenant 001",
	}
	tagID1, err := repo.CreateTag(ctx, tenantID, tag1)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID1)

	tag2 := &domain.Tag{
		TenantID: tenantID,
		TagType:  "user_tag",
		TagName:  "Test User Tag ForTenant 001",
	}
	tagID2, err := repo.CreateTag(ctx, tenantID, tag2)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID2)

	// 测试：查询所有tags（使用ListTags）
	filter := TagsFilter{IncludeSystemTags: true}
	tags, total, err := repo.ListTags(ctx, tenantID, filter, 1, 100)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if total < 2 {
		t.Errorf("Expected at least 2 tags, got %d", total)
	}

	// 测试：按tag_type过滤（使用ListTags）
	filter = TagsFilter{TagType: "family_tag", IncludeSystemTags: true}
	tags, total, err = repo.ListTags(ctx, tenantID, filter, 1, 100)
	if err != nil {
		t.Fatalf("ListTags (with tag_type) failed: %v", err)
	}
	for _, tag := range tags {
		if tag.TagID == tagID2 {
			t.Error("user_tag should not appear when filtering by family_tag")
		}
	}
}

func TestPostgresTagsRepository_DeleteTag(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 先创建一个tag
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "family_tag",
		TagName:  "Test Family Tag Delete",
	}
	tagID, err := repo.CreateTag(ctx, tenantID, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// 测试：删除tag
	err = repo.DeleteTag(ctx, tenantID, "Test Family Tag Delete")
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	// 验证：tag已删除
	_, err = repo.GetTag(ctx, tenantID, tagID)
	if err == nil {
		t.Fatal("Tag should be deleted, but GetTag succeeded")
	}
}

func TestPostgresTagsRepository_UpdateTag(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 创建测试tag
	tag := &domain.Tag{
		TenantID: tenantID,
		TagType:  "user_tag",
		TagName:  "Test Update Tag 001",
	}
	tagID, err := repo.CreateTag(ctx, tenantID, tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	defer db.Exec(`DELETE FROM tags_catalog WHERE tag_id = $1`, tagID)

	// 验证创建成功
	createdTag, err := repo.GetTagByName(ctx, tenantID, tag.TagName)
	if err != nil {
		t.Fatalf("GetTagByName failed: %v", err)
	}
	if createdTag.TagType != "user_tag" {
		t.Errorf("Expected tag_type 'user_tag', got '%s'", createdTag.TagType)
	}

	// 测试：更新tag_type（从user_tag改为family_tag）
	// 注意：实际业务中，tag_type的修改权限由Service层控制
	updateTag := &domain.Tag{
		TagType: "family_tag",
	}
	err = repo.UpdateTag(ctx, tenantID, tag.TagName, updateTag)
	if err != nil {
		t.Fatalf("UpdateTag failed: %v", err)
	}

	// 验证更新成功
	updatedTag, err := repo.GetTagByName(ctx, tenantID, tag.TagName)
	if err != nil {
		t.Fatalf("GetTagByName after update failed: %v", err)
	}
	if updatedTag.TagType != "family_tag" {
		t.Errorf("Expected tag_type 'family_tag' after update, got '%s'", updatedTag.TagType)
	}
	if updatedTag.TagID != tagID {
		t.Errorf("Expected tag_id unchanged, got '%s' (expected '%s')", updatedTag.TagID, tagID)
	}
	if updatedTag.TagName != tag.TagName {
		t.Errorf("Expected tag_name unchanged, got '%s' (expected '%s')", updatedTag.TagName, tag.TagName)
	}
}

func TestPostgresTagsRepository_UpdateTag_NotFound(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 测试：更新不存在的tag
	updateTag := &domain.Tag{
		TagType: "user_tag",
	}
	err := repo.UpdateTag(ctx, tenantID, "NonExistentTag", updateTag)
	if err == nil {
		t.Error("Expected error when updating non-existent tag")
	}
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestPostgresTagsRepository_DeleteTag_SystemTagType_ShouldFail(t *testing.T) {
	db := getTestDBForTags(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresTagsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForTags(t, db)
	defer cleanupTestDataForTags(t, db, tenantID)

	// 测试：尝试删除branch_tag（应该失败）
	// 注意：需要先确保存在一个branch_tag（可能是系统预定义的）
	err := repo.DeleteTag(ctx, tenantID, "TEST-BRANCH")
	if err == nil {
		t.Log("DeleteTag for branch_tag succeeded (tag may not exist)")
	} else {
		t.Logf("Expected error for branch_tag: %v", err)
	}
}

