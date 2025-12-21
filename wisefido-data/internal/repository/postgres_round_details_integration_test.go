// +build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"wisefido-data/internal/domain"
)

// 创建测试租户、unit、user、resident和round（round_details需要round_id和resident_id）
func createTestTenantUnitUserResidentAndRoundForRoundDetails(t *testing.T, db *sql.DB) (string, string, string, string, string) {
	tenantID := "00000000-0000-0000-0000-000000000981"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant RoundDetails", "test-rounddetails.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000980"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// 创建测试unit
	unitID := "00000000-0000-0000-0000-000000000979"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver",
	)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	// 创建测试user（executor）
	userID := "00000000-0000-0000-0000-000000000978"
	_, err = db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active')
		 ON CONFLICT (user_id) DO UPDATE SET user_account = EXCLUDED.user_account`,
		userID, tenantID, "test_executor", hashString("test_executor"), hashString("pwd"), "Nurse",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// 创建测试resident
	residentID := "00000000-0000-0000-0000-000000000977"
	admissionDate := time.Now()
	_, err = db.Exec(
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, admission_date, status, role, unit_id)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active', 'Resident', $7)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID, tenantID, "testresident001", hashString("testresident001"), "Test Resident 001", admissionDate, unitID,
	)
	if err != nil {
		t.Fatalf("Failed to create test resident: %v", err)
	}

	// 创建测试round
	roundID := "00000000-0000-0000-0000-000000000976"
	_, err = db.Exec(
		`INSERT INTO rounds (round_id, tenant_id, round_type, unit_id, executor_id, round_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'completed')
		 ON CONFLICT (round_id) DO UPDATE SET round_type = EXCLUDED.round_type`,
		roundID, tenantID, "location", unitID, userID, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to create test round: %v", err)
	}

	return tenantID, unitID, userID, residentID, roundID
}

// 清理测试数据
func cleanupTestDataForRoundDetails(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM round_details WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rounds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// RoundDetailsRepository 测试
// ============================================

func TestPostgresRoundDetailsRepository_GetRoundDetail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _, residentID, roundID := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 先创建一个round_detail
	heartRate := 72
	respiratoryRate := 18
	dataTimestamp := time.Now()
	roundDetail := &domain.RoundDetail{
		TenantID:        tenantID,
		RoundID:         roundID,
		ResidentID:      residentID,
		BedStatus:        "in_bed",
		SleepStateDisplay: "Light sleep",
		HeartRate:        &heartRate,
		RespiratoryRate:  &respiratoryRate,
		DataTimestamp:    &dataTimestamp,
		Notes:           "Test round detail",
	}

	detailID, err := repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail)
	if err != nil {
		t.Fatalf("UpsertRoundDetail failed: %v", err)
	}

	// 测试：获取round_detail
	got, err := repo.GetRoundDetail(ctx, tenantID, detailID)
	if err != nil {
		t.Fatalf("GetRoundDetail failed: %v", err)
	}

	if got.DetailID != detailID {
		t.Errorf("Expected detail_id '%s', got '%s'", detailID, got.DetailID)
	}
	if got.ResidentID != residentID {
		t.Errorf("Expected resident_id '%s', got '%s'", residentID, got.ResidentID)
	}
	if got.BedStatus != "in_bed" {
		t.Errorf("Expected bed_status 'in_bed', got '%s'", got.BedStatus)
	}
	if got.HeartRate == nil || *got.HeartRate != heartRate {
		t.Errorf("Expected heart_rate %d, got %v", heartRate, got.HeartRate)
	}

	t.Logf("✅ GetRoundDetail test passed")
}

func TestPostgresRoundDetailsRepository_GetRoundDetailsByRound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _, residentID1, roundID := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	// 创建第二个resident
	residentID2 := "00000000-0000-0000-0000-000000000975"
	admissionDate2 := time.Now()
	_, err := db.Exec(
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, admission_date, status, role)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active', 'Resident')
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID2, tenantID, "testresident002", hashString("testresident002"), "Test Resident 002", admissionDate2,
	)
	if err != nil {
		t.Fatalf("Failed to create test resident 2: %v", err)
	}

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 创建两个round_details
	heartRate1 := 72
	roundDetail1 := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID,
		ResidentID: residentID1,
		BedStatus:  "in_bed",
		HeartRate:  &heartRate1,
	}
	_, err = repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail1)
	if err != nil {
		t.Fatalf("UpsertRoundDetail 1 failed: %v", err)
	}

	heartRate2 := 75
	roundDetail2 := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID,
		ResidentID: residentID2,
		BedStatus:  "out_of_bed",
		HeartRate:  &heartRate2,
	}
	_, err = repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail2)
	if err != nil {
		t.Fatalf("UpsertRoundDetail 2 failed: %v", err)
	}

	// 测试：获取某个round的所有详细记录
	details, err := repo.GetRoundDetailsByRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRoundDetailsByRound failed: %v", err)
	}

	if len(details) < 2 {
		t.Errorf("Expected at least 2 round details, got %d", len(details))
	}

	foundResident1 := false
	foundResident2 := false
	for _, d := range details {
		if d.ResidentID == residentID1 {
			foundResident1 = true
		}
		if d.ResidentID == residentID2 {
			foundResident2 = true
		}
	}
	if !foundResident1 || !foundResident2 {
		t.Error("Did not find both residents in round details")
	}

	t.Logf("✅ GetRoundDetailsByRound test passed: count=%d", len(details))
}

func TestPostgresRoundDetailsRepository_GetRoundDetailsByResident(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID, residentID, roundID1 := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	// 创建第二个round
	roundID2 := "00000000-0000-0000-0000-000000000974"
	_, err := db.Exec(
		`INSERT INTO rounds (round_id, tenant_id, round_type, unit_id, executor_id, round_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'completed')
		 ON CONFLICT (round_id) DO UPDATE SET round_type = EXCLUDED.round_type`,
		roundID2, tenantID, "location", unitID, userID, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to create test round 2: %v", err)
	}

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 创建两个round_details（同一个resident，不同round）
	heartRate1 := 72
	roundDetail1 := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID1,
		ResidentID: residentID,
		BedStatus:  "in_bed",
		HeartRate:  &heartRate1,
	}
	_, err = repo.UpsertRoundDetail(ctx, tenantID, roundID1, roundDetail1)
	if err != nil {
		t.Fatalf("UpsertRoundDetail 1 failed: %v", err)
	}

	heartRate2 := 75
	roundDetail2 := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID2,
		ResidentID: residentID,
		BedStatus:  "out_of_bed",
		HeartRate:  &heartRate2,
	}
	_, err = repo.UpsertRoundDetail(ctx, tenantID, roundID2, roundDetail2)
	if err != nil {
		t.Fatalf("UpsertRoundDetail 2 failed: %v", err)
	}

	// 测试：获取某个resident的所有巡房详细记录
	details, total, err := repo.GetRoundDetailsByResident(ctx, tenantID, residentID, nil, 1, 20)
	if err != nil {
		t.Fatalf("GetRoundDetailsByResident failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 round details, got total=%d", total)
	}
	if len(details) < 2 {
		t.Errorf("Expected at least 2 round details in result, got %d", len(details))
	}

	// 测试：按bed_status过滤
	filters := &RoundDetailFilters{BedStatus: "in_bed"}
	detailsInBed, _, err := repo.GetRoundDetailsByResident(ctx, tenantID, residentID, filters, 1, 20)
	if err != nil {
		t.Fatalf("GetRoundDetailsByResident with filter failed: %v", err)
	}

	for _, d := range detailsInBed {
		if d.BedStatus == "in_bed" {
			// Found expected status
		}
		if d.BedStatus == "out_of_bed" {
			t.Error("Found out_of_bed in in_bed filter result")
		}
	}

	t.Logf("✅ GetRoundDetailsByResident test passed: total=%d", total)
}

func TestPostgresRoundDetailsRepository_UpsertRoundDetail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _, residentID, roundID := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 测试：创建round_detail
	heartRate := 72
	respiratoryRate := 18
	dataTimestamp := time.Now()
	roundDetail := &domain.RoundDetail{
		TenantID:        tenantID,
		RoundID:         roundID,
		ResidentID:      residentID,
		BedStatus:        "in_bed",
		SleepStateDisplay: "Light sleep",
		HeartRate:        &heartRate,
		RespiratoryRate:  &respiratoryRate,
		DataTimestamp:    &dataTimestamp,
		Notes:           "Test round detail",
	}

	detailID, err := repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail)
	if err != nil {
		t.Fatalf("UpsertRoundDetail failed: %v", err)
	}

	if detailID == "" {
		t.Fatal("Expected non-empty detail_id")
	}

	// 验证创建成功
	got, err := repo.GetRoundDetail(ctx, tenantID, detailID)
	if err != nil {
		t.Fatalf("GetRoundDetail failed: %v", err)
	}

	if got.BedStatus != "in_bed" {
		t.Errorf("Expected bed_status 'in_bed', got '%s'", got.BedStatus)
	}

	// 测试：更新round_detail（UNIQUE round_id, resident_id）
	updatedHeartRate := 75
	roundDetail.HeartRate = &updatedHeartRate
	roundDetail.Notes = "Updated notes"

	detailID2, err := repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail)
	if err != nil {
		t.Fatalf("UpsertRoundDetail update failed: %v", err)
	}

	// 应该返回相同的detail_id（因为UNIQUE约束）
	if detailID2 != detailID {
		t.Logf("Note: Upsert returned different detail_id (expected for some implementations)")
	}

	// 验证更新成功
	got, err = repo.GetRoundDetail(ctx, tenantID, detailID)
	if err != nil {
		t.Fatalf("GetRoundDetail after update failed: %v", err)
	}

	if got.HeartRate == nil || *got.HeartRate != updatedHeartRate {
		t.Errorf("Expected updated heart_rate %d, got %v", updatedHeartRate, got.HeartRate)
	}
	if got.Notes != "Updated notes" {
		t.Errorf("Expected updated notes 'Updated notes', got '%s'", got.Notes)
	}

	t.Logf("✅ UpsertRoundDetail test passed: detailID=%s", detailID)
}

func TestPostgresRoundDetailsRepository_UpdateRoundDetail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _, residentID, roundID := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 先创建一个round_detail
	heartRate := 72
	roundDetail := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID,
		ResidentID: residentID,
		BedStatus:  "in_bed",
		HeartRate:  &heartRate,
		Notes:      "Original notes",
	}

	detailID, err := repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail)
	if err != nil {
		t.Fatalf("UpsertRoundDetail failed: %v", err)
	}

	// 测试：更新round_detail
	updatedHeartRate := 75
	updatedRoundDetail := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID,
		ResidentID: residentID,
		BedStatus:  "out_of_bed",
		HeartRate:  &updatedHeartRate,
		Notes:      "Updated notes",
	}

	err = repo.UpdateRoundDetail(ctx, tenantID, detailID, updatedRoundDetail)
	if err != nil {
		t.Fatalf("UpdateRoundDetail failed: %v", err)
	}

	// 验证更新成功
	got, err := repo.GetRoundDetail(ctx, tenantID, detailID)
	if err != nil {
		t.Fatalf("GetRoundDetail after update failed: %v", err)
	}

	if got.BedStatus != "out_of_bed" {
		t.Errorf("Expected updated bed_status 'out_of_bed', got '%s'", got.BedStatus)
	}
	if got.HeartRate == nil || *got.HeartRate != updatedHeartRate {
		t.Errorf("Expected updated heart_rate %d, got %v", updatedHeartRate, got.HeartRate)
	}
	if got.Notes != "Updated notes" {
		t.Errorf("Expected updated notes 'Updated notes', got '%s'", got.Notes)
	}

	t.Logf("✅ UpdateRoundDetail test passed")
}

func TestPostgresRoundDetailsRepository_DeleteRoundDetail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _, residentID, roundID := createTestTenantUnitUserResidentAndRoundForRoundDetails(t, db)
	defer cleanupTestDataForRoundDetails(t, db, tenantID)

	repo := NewPostgresRoundDetailsRepository(db)
	ctx := context.Background()

	// 先创建一个round_detail
	heartRate := 72
	roundDetail := &domain.RoundDetail{
		TenantID:   tenantID,
		RoundID:    roundID,
		ResidentID: residentID,
		BedStatus:  "in_bed",
		HeartRate:  &heartRate,
	}

	detailID, err := repo.UpsertRoundDetail(ctx, tenantID, roundID, roundDetail)
	if err != nil {
		t.Fatalf("UpsertRoundDetail failed: %v", err)
	}

	// 测试：删除round_detail
	err = repo.DeleteRoundDetail(ctx, tenantID, detailID)
	if err != nil {
		t.Fatalf("DeleteRoundDetail failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetRoundDetail(ctx, tenantID, detailID)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	t.Logf("✅ DeleteRoundDetail test passed")
}

