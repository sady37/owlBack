// +build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"wisefido-data/internal/domain"
)

// 创建测试租户、unit和user（rounds需要executor_id和unit_id）
func createTestTenantUnitAndUserForRounds(t *testing.T, db *sql.DB) (string, string, string) {
	tenantID := "00000000-0000-0000-0000-000000000985"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Rounds", "test-rounds.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000984"
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
	unitID := "00000000-0000-0000-0000-000000000983"
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
	userID := "00000000-0000-0000-0000-000000000982"
	_, err = db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active')
		 ON CONFLICT (user_id) DO UPDATE SET user_account = EXCLUDED.user_account`,
		userID, tenantID, "test_executor", hashString("test_executor"), hashString("pwd"), "Nurse",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return tenantID, unitID, userID
}

// 清理测试数据
func cleanupTestDataForRounds(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM round_details WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rounds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// RoundsRepository 测试
// ============================================

func TestPostgresRoundsRepository_GetRound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 先创建一个round
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "completed",
		Notes:      "Test round",
	}

	roundID, err := repo.CreateRound(ctx, tenantID, round)
	if err != nil {
		t.Fatalf("CreateRound failed: %v", err)
	}

	// 测试：获取round
	got, err := repo.GetRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRound failed: %v", err)
	}

	if got.RoundID != roundID {
		t.Errorf("Expected round_id '%s', got '%s'", roundID, got.RoundID)
	}
	if got.RoundType != "location" {
		t.Errorf("Expected round_type 'location', got '%s'", got.RoundType)
	}
	if got.ExecutorID != userID {
		t.Errorf("Expected executor_id '%s', got '%s'", userID, got.ExecutorID)
	}

	t.Logf("✅ GetRound test passed")
}

func TestPostgresRoundsRepository_ListRounds(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 创建多个rounds
	round1 := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "completed",
	}
	roundID1, err := repo.CreateRound(ctx, tenantID, round1)
	if err != nil {
		t.Fatalf("CreateRound 1 failed: %v", err)
	}

	round2 := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "manual",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now().Add(time.Hour),
		Status:     "draft",
	}
	roundID2, err := repo.CreateRound(ctx, tenantID, round2)
	if err != nil {
		t.Fatalf("CreateRound 2 failed: %v", err)
	}

	// 测试：列表查询（无过滤）
	_, total, err := repo.ListRounds(ctx, tenantID, nil, 1, 20)
	if err != nil {
		t.Fatalf("ListRounds failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 rounds, got total=%d", total)
	}

	// 测试：按status过滤
	filters := &RoundFilters{Status: "completed"}
	roundsCompleted, _, err := repo.ListRounds(ctx, tenantID, filters, 1, 20)
	if err != nil {
		t.Fatalf("ListRounds with status filter failed: %v", err)
	}

	foundRound1 := false
	for _, r := range roundsCompleted {
		if r.RoundID == roundID1 {
			foundRound1 = true
		}
		if r.RoundID == roundID2 {
			t.Error("Found draft round in completed filter result")
		}
	}
	if !foundRound1 {
		t.Error("Did not find completed round in filtered result")
	}

	// 测试：按round_type过滤
	filters = &RoundFilters{RoundType: "manual"}
	roundsManual, _, err := repo.ListRounds(ctx, tenantID, filters, 1, 20)
	if err != nil {
		t.Fatalf("ListRounds with round_type filter failed: %v", err)
	}

	for _, r := range roundsManual {
		if r.RoundID == roundID2 {
			// Found expected round
		}
	}

	// 测试：按时间范围过滤
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now().Add(time.Hour)
	filters = &RoundFilters{StartTime: &startTime, EndTime: &endTime}
	roundsTimeFiltered, _, err := repo.ListRounds(ctx, tenantID, filters, 1, 20)
	if err != nil {
		t.Fatalf("ListRounds with time filter failed: %v", err)
	}
	if len(roundsTimeFiltered) < 2 {
		t.Errorf("Expected at least 2 rounds in time range, got %d", len(roundsTimeFiltered))
	}

	t.Logf("✅ ListRounds test passed: total=%d", total)
}

func TestPostgresRoundsRepository_CreateRound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 测试：创建round
	roundTime := time.Now()
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  roundTime,
		Status:     "completed",
		Notes:      "Test round creation",
	}

	roundID, err := repo.CreateRound(ctx, tenantID, round)
	if err != nil {
		t.Fatalf("CreateRound failed: %v", err)
	}

	if roundID == "" {
		t.Fatal("Expected non-empty round_id")
	}

	// 验证创建成功
	got, err := repo.GetRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRound failed: %v", err)
	}

	if got.RoundType != "location" {
		t.Errorf("Expected round_type 'location', got '%s'", got.RoundType)
	}
	if got.Notes != "Test round creation" {
		t.Errorf("Expected notes 'Test round creation', got '%s'", got.Notes)
	}

	t.Logf("✅ CreateRound test passed: roundID=%s", roundID)
}

func TestPostgresRoundsRepository_UpdateRound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 先创建一个round
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "completed",
		Notes:      "Original notes",
	}

	roundID, err := repo.CreateRound(ctx, tenantID, round)
	if err != nil {
		t.Fatalf("CreateRound failed: %v", err)
	}

	// 测试：更新round
	updatedRound := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "manual",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "draft",
		Notes:      "Updated notes",
	}

	err = repo.UpdateRound(ctx, tenantID, roundID, updatedRound)
	if err != nil {
		t.Fatalf("UpdateRound failed: %v", err)
	}

	// 验证更新成功
	got, err := repo.GetRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRound after update failed: %v", err)
	}

	if got.RoundType != "manual" {
		t.Errorf("Expected updated round_type 'manual', got '%s'", got.RoundType)
	}
	if got.Status != "draft" {
		t.Errorf("Expected updated status 'draft', got '%s'", got.Status)
	}
	if got.Notes != "Updated notes" {
		t.Errorf("Expected updated notes 'Updated notes', got '%s'", got.Notes)
	}

	t.Logf("✅ UpdateRound test passed")
}

func TestPostgresRoundsRepository_DeleteRound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 先创建一个round
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "completed",
	}

	roundID, err := repo.CreateRound(ctx, tenantID, round)
	if err != nil {
		t.Fatalf("CreateRound failed: %v", err)
	}

	// 测试：删除round
	err = repo.DeleteRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("DeleteRound failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetRound(ctx, tenantID, roundID)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	t.Logf("✅ DeleteRound test passed")
}

func TestPostgresRoundsRepository_SetRoundStatus(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, userID := createTestTenantUnitAndUserForRounds(t, db)
	defer cleanupTestDataForRounds(t, db, tenantID)

	repo := NewPostgresRoundsRepository(db)
	ctx := context.Background()

	// 先创建一个round
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "location",
		UnitID:     unitID,
		ExecutorID: userID,
		RoundTime:  time.Now(),
		Status:     "draft",
	}

	roundID, err := repo.CreateRound(ctx, tenantID, round)
	if err != nil {
		t.Fatalf("CreateRound failed: %v", err)
	}

	// 测试：更新status
	err = repo.SetRoundStatus(ctx, tenantID, roundID, "completed")
	if err != nil {
		t.Fatalf("SetRoundStatus failed: %v", err)
	}

	// 验证更新成功
	got, err := repo.GetRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRound after status update failed: %v", err)
	}

	if got.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", got.Status)
	}

	// 测试：更新为cancelled
	err = repo.SetRoundStatus(ctx, tenantID, roundID, "cancelled")
	if err != nil {
		t.Fatalf("SetRoundStatus to cancelled failed: %v", err)
	}

	got, err = repo.GetRound(ctx, tenantID, roundID)
	if err != nil {
		t.Fatalf("GetRound after cancelled update failed: %v", err)
	}

	if got.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", got.Status)
	}

	t.Logf("✅ SetRoundStatus test passed")
}

