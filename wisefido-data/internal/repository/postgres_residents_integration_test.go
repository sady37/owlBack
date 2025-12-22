// +build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"wisefido-data/internal/domain"
)

// 创建测试租户和unit（resident需要unit_id）
func createTestTenantAndUnitForResidents(t *testing.T, db *sql.DB) (string, string) {
	tenantID := "00000000-0000-0000-0000-000000000998"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Residents", "test-residents.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000997"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// 创建测试unit（需要branch_tag、unit_number和timezone）
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

	return tenantID, unitID
}

// 清理测试数据
func cleanupTestDataForResidents(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：resident_caregivers -> resident_contacts -> resident_phi -> residents -> beds -> rooms -> units -> buildings -> tags_catalog -> tenants
	db.Exec(`DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// hashString 函数已在 postgres_users_integration_test.go 中定义，这里不再重复定义

// ============================================
// Residents 表操作测试
// ============================================

func TestPostgresResidentsRepository_CreateResident(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident001",
		ResidentAccountHash: hashString("testresident001"),
		Nickname:            "Test Resident 001",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		Role:                "Resident",
		FamilyTag:           "F0001",
		CanViewStatus:       true,
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	if residentID == "" {
		t.Fatal("Expected non-empty resident_id")
	}

	// 验证创建成功
	createdResident, err := repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if createdResident.Nickname != resident.Nickname {
		t.Errorf("Expected nickname '%s', got '%s'", resident.Nickname, createdResident.Nickname)
	}
	if createdResident.FamilyTag != resident.FamilyTag {
		t.Errorf("Expected family_tag '%s', got '%s'", resident.FamilyTag, createdResident.FamilyTag)
	}

	// 验证family_tag已同步到tags_catalog
	var tagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "F0001", "family_tag").Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to verify tag in catalog: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("Expected 1 tag in catalog, got %d", tagCount)
	}

	t.Logf("✅ CreateResident test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_GetResident(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident002",
		ResidentAccountHash: hashString("testresident002"),
		Nickname:            "Test Resident 002",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 获取住户
	gotResident, err := repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if gotResident.ResidentID != residentID {
		t.Errorf("Expected resident_id '%s', got '%s'", residentID, gotResident.ResidentID)
	}
	if gotResident.Nickname != resident.Nickname {
		t.Errorf("Expected nickname '%s', got '%s'", resident.Nickname, gotResident.Nickname)
	}

	t.Logf("✅ GetResident test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_GetResidentByAccount(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	account := "testresident003"
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     account,
		ResidentAccountHash: hashString(account),
		Nickname:            "Test Resident 003",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 根据account_hash获取住户
	gotResident, err := repo.GetResidentByAccount(ctx, tenantID, hashString(account))
	if err != nil {
		t.Fatalf("GetResidentByAccount failed: %v", err)
	}

	if gotResident.ResidentID != residentID {
		t.Errorf("Expected resident_id '%s', got '%s'", residentID, gotResident.ResidentID)
	}

	t.Logf("✅ GetResidentByAccount test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_GetResidentByEmail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户（带email_hash）
	email := "resident@example.com"
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident004",
		ResidentAccountHash: hashString("testresident004"),
		Nickname:             "Test Resident 004",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		EmailHash:           hashString(email),
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 根据email_hash获取住户
	gotResident, err := repo.GetResidentByEmail(ctx, tenantID, hashString(email))
	if err != nil {
		t.Fatalf("GetResidentByEmail failed: %v", err)
	}

	if gotResident.ResidentID != residentID {
		t.Errorf("Expected resident_id '%s', got '%s'", residentID, gotResident.ResidentID)
	}

	t.Logf("✅ GetResidentByEmail test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_GetResidentByPhone(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户（带phone_hash）
	phone := "1234567890"
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident005",
		ResidentAccountHash: hashString("testresident005"),
		Nickname:            "Test Resident 005",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		PhoneHash:           hashString(phone),
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 根据phone_hash获取住户
	gotResident, err := repo.GetResidentByPhone(ctx, tenantID, hashString(phone))
	if err != nil {
		t.Fatalf("GetResidentByPhone failed: %v", err)
	}

	if gotResident.ResidentID != residentID {
		t.Errorf("Expected resident_id '%s', got '%s'", residentID, gotResident.ResidentID)
	}

	t.Logf("✅ GetResidentByPhone test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_ListResidents(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident1 := &domain.Resident{
		ResidentAccount:     "testresident006",
		ResidentAccountHash: hashString("testresident006"),
		Nickname:            "Test Resident 006",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}
	residentID1, err := repo.CreateResident(ctx, tenantID, resident1)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	resident2 := &domain.Resident{
		ResidentAccount:     "testresident007",
		ResidentAccountHash: hashString("testresident007"),
		Nickname:            "Test Resident 007",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		ServiceLevel:        "Independent",
		UnitID:              unitID,
	}
	residentID2, err := repo.CreateResident(ctx, tenantID, resident2)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 测试：查询所有住户
	filter := ResidentFilters{}
	residents, total, err := repo.ListResidents(ctx, tenantID, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListResidents failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 residents, got %d", total)
	}
	if len(residents) < 2 {
		t.Errorf("Expected at least 2 residents in result, got %d", len(residents))
	}

	// 测试：按status过滤
	filter = ResidentFilters{Status: "active"}
	residents, total, err = repo.ListResidents(ctx, tenantID, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListResidents (with status filter) failed: %v", err)
	}

	for _, r := range residents {
		if r.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", r.Status)
		}
	}

	// 测试：按service_level过滤
	filter = ResidentFilters{ServiceLevel: "Independent"}
	residents, total, err = repo.ListResidents(ctx, tenantID, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListResidents (with service_level filter) failed: %v", err)
	}

	found := false
	for _, r := range residents {
		if r.ResidentID == residentID2 {
			found = true
			if r.ServiceLevel != "Independent" {
				t.Errorf("Expected service_level 'Independent', got '%s'", r.ServiceLevel)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find resident with service_level 'Independent'")
	}

	// 测试：按nickname搜索
	filter = ResidentFilters{Search: "006"}
	residents, total, err = repo.ListResidents(ctx, tenantID, filter, 1, 10)
	if err != nil {
		t.Fatalf("ListResidents (with search) failed: %v", err)
	}

	found = false
	for _, r := range residents {
		if r.ResidentID == residentID1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find resident with nickname containing '006'")
	}

	t.Logf("✅ ListResidents test passed: total=%d", total)
}

func TestPostgresResidentsRepository_UpdateResident(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident008",
		ResidentAccountHash: hashString("testresident008"),
		Nickname:            "Test Resident 008",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 更新住户
	updatedResident := &domain.Resident{
		Nickname:     "Updated Resident 008",
		ServiceLevel: "Assisted",
		FamilyTag:    "F0002",
		Note:         "Updated note",
	}

	err = repo.UpdateResident(ctx, tenantID, residentID, updatedResident)
	if err != nil {
		t.Fatalf("UpdateResident failed: %v", err)
	}

	// 验证更新成功
	gotResident, err := repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if gotResident.Nickname != updatedResident.Nickname {
		t.Errorf("Expected nickname '%s', got '%s'", updatedResident.Nickname, gotResident.Nickname)
	}
	if gotResident.ServiceLevel != updatedResident.ServiceLevel {
		t.Errorf("Expected service_level '%s', got '%s'", updatedResident.ServiceLevel, gotResident.ServiceLevel)
	}
	if gotResident.FamilyTag != updatedResident.FamilyTag {
		t.Errorf("Expected family_tag '%s', got '%s'", updatedResident.FamilyTag, gotResident.FamilyTag)
	}

	// 验证family_tag已同步到tags_catalog
	var tagCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "F0002", "family_tag").Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to verify tag in catalog: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("Expected 1 tag in catalog, got %d", tagCount)
	}

	t.Logf("✅ UpdateResident test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_DeleteResident(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident009",
		ResidentAccountHash: hashString("testresident009"),
		Nickname:            "Test Resident 009",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		FamilyTag:           "F0003",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 记录family_tag在tags_catalog中的存在
	var tagCountBefore int
	db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2`,
		tenantID, "F0003").Scan(&tagCountBefore)

	// 删除住户
	err = repo.DeleteResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("DeleteResident failed: %v", err)
	}

	// 验证住户已删除
	_, err = repo.GetResident(ctx, tenantID, residentID)
	if err == nil {
		t.Error("Expected error when getting deleted resident")
	}

	// 验证family_tag仍然保留在tags_catalog中（不删除）
	var tagCountAfter int
	db.QueryRow(`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2`,
		tenantID, "F0003").Scan(&tagCountAfter)
	if tagCountAfter != tagCountBefore {
		t.Errorf("Expected family_tag to remain in catalog (before=%d, after=%d)", tagCountBefore, tagCountAfter)
	}

	t.Logf("✅ DeleteResident test passed: residentID=%s (family_tag preserved)", residentID)
}

func TestPostgresResidentsRepository_BindResidentToLocation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试room和bed
	roomID := "00000000-0000-0000-0000-000000000995"
	_, err := db.Exec(
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID, tenantID, unitID, "Test Room 001",
	)
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bedID := "00000000-0000-0000-0000-000000000994"
	_, err = db.Exec(
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID, tenantID, roomID, "Test Bed 001",
	)
	if err != nil {
		t.Fatalf("Failed to create test bed: %v", err)
	}

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident010",
		ResidentAccountHash: hashString("testresident010"),
		Nickname:            "Test Resident 010",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 测试：绑定到unit+room+bed
	err = repo.BindResidentToLocation(ctx, tenantID, residentID, &unitID, &roomID, &bedID)
	if err != nil {
		t.Fatalf("BindResidentToLocation failed: %v", err)
	}

	gotResident, err := repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if gotResident.UnitID != unitID {
		t.Errorf("Expected unit_id '%s', got '%s'", unitID, gotResident.UnitID)
	}
	if gotResident.RoomID != roomID {
		t.Errorf("Expected room_id '%s', got '%s'", roomID, gotResident.RoomID)
	}
	if gotResident.BedID != bedID {
		t.Errorf("Expected bed_id '%s', got '%s'", bedID, gotResident.BedID)
	}

	// 测试：解绑bed（传入nil）
	err = repo.BindResidentToLocation(ctx, tenantID, residentID, &unitID, &roomID, nil)
	if err != nil {
		t.Fatalf("BindResidentToLocation (unbind bed) failed: %v", err)
	}

	gotResident, err = repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if gotResident.BedID != "" {
		t.Errorf("Expected bed_id to be empty after unbind, got '%s'", gotResident.BedID)
	}

	// 测试：解绑room（传入nil）
	err = repo.BindResidentToLocation(ctx, tenantID, residentID, &unitID, nil, nil)
	if err != nil {
		t.Fatalf("BindResidentToLocation (unbind room) failed: %v", err)
	}

	gotResident, err = repo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResident failed: %v", err)
	}

	if gotResident.RoomID != "" {
		t.Errorf("Expected room_id to be empty after unbind, got '%s'", gotResident.RoomID)
	}

	t.Logf("✅ BindResidentToLocation test passed: residentID=%s", residentID)
}

// ============================================
// ResidentPHI 表操作测试
// ============================================

func TestPostgresResidentsRepository_GetResidentPHI(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident011",
		ResidentAccountHash: hashString("testresident011"),
		Nickname:            "Test Resident 011",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建PHI
	dob := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC)
	phi := &domain.ResidentPHI{
		FirstName:    "John",
		LastName:     "Doe",
		Gender:       "Male",
		DateOfBirth:  &dob,
		ResidentPhone: "1234567890",
		ResidentEmail: "john@example.com",
	}

	err = repo.UpsertResidentPHI(ctx, tenantID, residentID, phi)
	if err != nil {
		t.Fatalf("UpsertResidentPHI failed: %v", err)
	}

	// 获取PHI
	gotPHI, err := repo.GetResidentPHI(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentPHI failed: %v", err)
	}

	if gotPHI.FirstName != phi.FirstName {
		t.Errorf("Expected first_name '%s', got '%s'", phi.FirstName, gotPHI.FirstName)
	}
	if gotPHI.LastName != phi.LastName {
		t.Errorf("Expected last_name '%s', got '%s'", phi.LastName, gotPHI.LastName)
	}

	t.Logf("✅ GetResidentPHI test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_UpsertResidentPHI(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident012",
		ResidentAccountHash: hashString("testresident012"),
		Nickname:            "Test Resident 012",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建PHI
	dob := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC)
	weight := 150.0
	heightFt := 5.0
	heightIn := 10.0
	phi := &domain.ResidentPHI{
		FirstName:     "Jane",
		LastName:      "Smith",
		Gender:         "Female",
		DateOfBirth:    &dob,
		WeightLb:       &weight,
		HeightFt:       &heightFt,
		HeightIn:       &heightIn,
		HasHypertension: true,
		HasAlzheimer:    true,
	}

	err = repo.UpsertResidentPHI(ctx, tenantID, residentID, phi)
	if err != nil {
		t.Fatalf("UpsertResidentPHI failed: %v", err)
	}

	// 更新PHI
	phi.FirstName = "Jane Updated"
	phi.HasHypertension = false

	err = repo.UpsertResidentPHI(ctx, tenantID, residentID, phi)
	if err != nil {
		t.Fatalf("UpsertResidentPHI (update) failed: %v", err)
	}

	// 验证更新成功
	gotPHI, err := repo.GetResidentPHI(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentPHI failed: %v", err)
	}

	if gotPHI.FirstName != phi.FirstName {
		t.Errorf("Expected first_name '%s', got '%s'", phi.FirstName, gotPHI.FirstName)
	}
	if gotPHI.HasHypertension != phi.HasHypertension {
		t.Errorf("Expected has_hypertension %v, got %v", phi.HasHypertension, gotPHI.HasHypertension)
	}

	t.Logf("✅ UpsertResidentPHI test passed: residentID=%s", residentID)
}

// ============================================
// ResidentContacts 表操作测试
// ============================================

func TestPostgresResidentsRepository_GetResidentContacts(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident013",
		ResidentAccountHash: hashString("testresident013"),
		Nickname:            "Test Resident 013",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建联系人
	contact1 := &domain.ResidentContact{
		Slot:              "A",
		IsEnabled:          true,
		Relationship:       "Child",
		ContactFirstName:   "Contact",
		ContactLastName:    "One",
		ContactPhone:       "1111111111",
		ContactEmail:       "contact1@example.com",
		IsEmergencyContact: true,
	}

	contactID1, err := repo.CreateResidentContact(ctx, tenantID, residentID, contact1)
	if err != nil {
		t.Fatalf("CreateResidentContact failed: %v", err)
	}

	contact2 := &domain.ResidentContact{
		Slot:            "B",
		IsEnabled:       true,
		Relationship:    "Spouse",
		ContactFirstName: "Contact",
		ContactLastName:  "Two",
	}

	contactID2, err := repo.CreateResidentContact(ctx, tenantID, residentID, contact2)
	if err != nil {
		t.Fatalf("CreateResidentContact failed: %v", err)
	}

	// 获取所有联系人
	contacts, err := repo.GetResidentContacts(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentContacts failed: %v", err)
	}

	if len(contacts) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(contacts))
	}

	found1, found2 := false, false
	for _, c := range contacts {
		if c.ContactID == contactID1 {
			found1 = true
			if c.Slot != "A" {
				t.Errorf("Expected slot 'A', got '%s'", c.Slot)
			}
		}
		if c.ContactID == contactID2 {
			found2 = true
			if c.Slot != "B" {
				t.Errorf("Expected slot 'B', got '%s'", c.Slot)
			}
		}
	}
	if !found1 || !found2 {
		t.Error("Expected to find both contacts")
	}

	t.Logf("✅ GetResidentContacts test passed: residentID=%s", residentID)
}

func TestPostgresResidentsRepository_CreateResidentContact(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident014",
		ResidentAccountHash: hashString("testresident014"),
		Nickname:            "Test Resident 014",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建联系人
	contact := &domain.ResidentContact{
		Slot:              "A",
		IsEnabled:          true,
		Relationship:       "Child",
		ContactFirstName:   "Contact",
		ContactLastName:    "Test",
		ContactPhone:       "2222222222",
		ContactEmail:       "contact@example.com",
		IsEmergencyContact: true,
		ReceiveSMS:         true,
		ReceiveEmail:       true,
	}

	contactID, err := repo.CreateResidentContact(ctx, tenantID, residentID, contact)
	if err != nil {
		t.Fatalf("CreateResidentContact failed: %v", err)
	}

	if contactID == "" {
		t.Fatal("Expected non-empty contact_id")
	}

	// 验证创建成功
	contacts, err := repo.GetResidentContacts(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentContacts failed: %v", err)
	}

	found := false
	for _, c := range contacts {
		if c.ContactID == contactID {
			found = true
			if c.Slot != contact.Slot {
				t.Errorf("Expected slot '%s', got '%s'", contact.Slot, c.Slot)
			}
			if c.ContactFirstName != contact.ContactFirstName {
				t.Errorf("Expected contact_first_name '%s', got '%s'", contact.ContactFirstName, c.ContactFirstName)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find created contact")
	}

	t.Logf("✅ CreateResidentContact test passed: contactID=%s", contactID)
}

func TestPostgresResidentsRepository_UpdateResidentContact(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident015",
		ResidentAccountHash: hashString("testresident015"),
		Nickname:            "Test Resident 015",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建联系人
	contact := &domain.ResidentContact{
		Slot:            "A",
		IsEnabled:       true,
		Relationship:    "Child",
		ContactFirstName: "Contact",
		ContactLastName:  "Original",
	}

	contactID, err := repo.CreateResidentContact(ctx, tenantID, residentID, contact)
	if err != nil {
		t.Fatalf("CreateResidentContact failed: %v", err)
	}

	// 更新联系人
	updatedContact := &domain.ResidentContact{
		Slot:              "A",
		IsEnabled:         false,
		Relationship:      "Spouse",
		ContactFirstName:  "Contact",
		ContactLastName:   "Updated",
		IsEmergencyContact: true,
	}

	err = repo.UpdateResidentContact(ctx, tenantID, contactID, updatedContact)
	if err != nil {
		t.Fatalf("UpdateResidentContact failed: %v", err)
	}

	// 验证更新成功
	contacts, err := repo.GetResidentContacts(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentContacts failed: %v", err)
	}

	found := false
	for _, c := range contacts {
		if c.ContactID == contactID {
			found = true
			if c.IsEnabled != updatedContact.IsEnabled {
				t.Errorf("Expected is_enabled %v, got %v", updatedContact.IsEnabled, c.IsEnabled)
			}
			if c.Relationship != updatedContact.Relationship {
				t.Errorf("Expected relationship '%s', got '%s'", updatedContact.Relationship, c.Relationship)
			}
			if c.ContactLastName != updatedContact.ContactLastName {
				t.Errorf("Expected contact_last_name '%s', got '%s'", updatedContact.ContactLastName, c.ContactLastName)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find updated contact")
	}

	t.Logf("✅ UpdateResidentContact test passed: contactID=%s", contactID)
}

func TestPostgresResidentsRepository_DeleteResidentContact(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident016",
		ResidentAccountHash: hashString("testresident016"),
		Nickname:            "Test Resident 016",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建联系人
	contact := &domain.ResidentContact{
		Slot:         "A",
		IsEnabled:     true,
		Relationship:  "Child",
	}

	contactID, err := repo.CreateResidentContact(ctx, tenantID, residentID, contact)
	if err != nil {
		t.Fatalf("CreateResidentContact failed: %v", err)
	}

	// 删除联系人
	err = repo.DeleteResidentContact(ctx, tenantID, contactID)
	if err != nil {
		t.Fatalf("DeleteResidentContact failed: %v", err)
	}

	// 验证删除成功
	contacts, err := repo.GetResidentContacts(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentContacts failed: %v", err)
	}

	for _, c := range contacts {
		if c.ContactID == contactID {
			t.Error("Expected contact to be deleted")
		}
	}

	t.Logf("✅ DeleteResidentContact test passed: contactID=%s", contactID)
}

// ============================================
// ResidentCaregivers 表操作测试
// ============================================

func TestPostgresResidentsRepository_GetResidentCaregivers(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 设置unit级别的caregiver配置
	unitGroupList := json.RawMessage(`["Group1", "Group2"]`)
	unitUserList := json.RawMessage(`["user1", "user2"]`)
	_, err := db.Exec(
		`UPDATE units SET groupList = $1::jsonb, userList = $2::jsonb WHERE tenant_id = $3 AND unit_id = $4`,
		string(unitGroupList), string(unitUserList), tenantID, unitID,
	)
	if err != nil {
		t.Fatalf("Failed to update unit caregiver config: %v", err)
	}

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident017",
		ResidentAccountHash: hashString("testresident017"),
		Nickname:            "Test Resident 017",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 设置resident级别的caregiver配置
	residentGroupList := json.RawMessage(`["Group3"]`)
	residentUserList := json.RawMessage(`["user3"]`)
	caregiver := &domain.ResidentCaregiver{
		GroupList: residentGroupList,
		UserList:  residentUserList,
	}

	err = repo.UpsertResidentCaregiver(ctx, tenantID, residentID, caregiver)
	if err != nil {
		t.Fatalf("UpsertResidentCaregiver failed: %v", err)
	}

	// 获取caregivers（应该包含unit级别和resident级别）
	caregivers, err := repo.GetResidentCaregivers(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentCaregivers failed: %v", err)
	}

	if len(caregivers) < 1 {
		t.Errorf("Expected at least 1 caregiver config, got %d", len(caregivers))
	}

	// 验证unit级别的配置
	foundUnit := false
	foundResident := false
	for _, c := range caregivers {
		if c.Source == "unit" {
			foundUnit = true
			if len(c.GroupList) == 0 && len(c.UserList) == 0 {
				t.Error("Expected unit-level caregiver config to have groupList or userList")
			}
		}
		if c.Source == "resident" {
			foundResident = true
			if len(c.GroupList) == 0 && len(c.UserList) == 0 {
				t.Error("Expected resident-level caregiver config to have groupList or userList")
			}
		}
	}

	if !foundUnit {
		t.Error("Expected to find unit-level caregiver config")
	}
	if !foundResident {
		t.Error("Expected to find resident-level caregiver config")
	}

	t.Logf("✅ GetResidentCaregivers test passed: residentID=%s, found %d configs", residentID, len(caregivers))
}

func TestPostgresResidentsRepository_UpsertResidentCaregiver(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResidents(t, db)
	defer cleanupTestDataForResidents(t, db, tenantID)

	repo := NewPostgresResidentsRepository(db)
	ctx := context.Background()

	// 创建测试住户
	admissionDate := time.Now()
	resident := &domain.Resident{
		ResidentAccount:     "testresident018",
		ResidentAccountHash: hashString("testresident018"),
		Nickname:            "Test Resident 018",
		AdmissionDate:       &admissionDate,
		Status:              "active",
		UnitID:              unitID,
	}

	residentID, err := repo.CreateResident(ctx, tenantID, resident)
	if err != nil {
		t.Fatalf("CreateResident failed: %v", err)
	}

	// 创建caregiver配置
	groupList := json.RawMessage(`["Group1", "Group2"]`)
	userList := json.RawMessage(`["user1", "user2"]`)
	caregiver := &domain.ResidentCaregiver{
		GroupList: groupList,
		UserList:  userList,
	}

	err = repo.UpsertResidentCaregiver(ctx, tenantID, residentID, caregiver)
	if err != nil {
		t.Fatalf("UpsertResidentCaregiver failed: %v", err)
	}

	// 更新caregiver配置
	updatedGroupList := json.RawMessage(`["Group3"]`)
	updatedUserList := json.RawMessage(`["user3"]`)
	updatedCaregiver := &domain.ResidentCaregiver{
		GroupList: updatedGroupList,
		UserList:  updatedUserList,
	}

	err = repo.UpsertResidentCaregiver(ctx, tenantID, residentID, updatedCaregiver)
	if err != nil {
		t.Fatalf("UpsertResidentCaregiver (update) failed: %v", err)
	}

	// 验证更新成功
	caregivers, err := repo.GetResidentCaregivers(ctx, tenantID, residentID)
	if err != nil {
		t.Fatalf("GetResidentCaregivers failed: %v", err)
	}

	found := false
	for _, c := range caregivers {
		if c.Source == "resident" {
			found = true
			if string(c.GroupList) != string(updatedGroupList) {
				t.Errorf("Expected groupList '%s', got '%s'", string(updatedGroupList), string(c.GroupList))
			}
			if string(c.UserList) != string(updatedUserList) {
				t.Errorf("Expected userList '%s', got '%s'", string(updatedUserList), string(c.UserList))
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find resident-level caregiver config")
	}

	t.Logf("✅ UpsertResidentCaregiver test passed: residentID=%s", residentID)
}

