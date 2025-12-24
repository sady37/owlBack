// +build integration

package service

import (
	"context"
	"database/sql"
	"testing"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// setupTestDBForUnit 设置测试数据库（复用 getTestDBForService）
func setupTestDBForUnit(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// createTestTenantForUnit 创建测试租户
func createTestTenantForUnit(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000996"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Unit Tenant", "test-unit.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// cleanupTestDataForUnit 清理测试数据
func cleanupTestDataForUnit(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：beds -> rooms -> units -> buildings -> tags_catalog -> tenants
	_, _ = db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestLoggerForUnit 获取测试日志记录器（复用 getTestLogger）
func getTestLoggerForUnit() *zap.Logger {
	return getTestLogger()
}

// ============================================
// Building 测试
// ============================================

// TestUnitService_ListBuildings_Success 测试查询楼栋列表成功
func TestUnitService_ListBuildings_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 创建测试数据
	building1 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID1, err := unitsRepo.CreateBuilding(context.Background(), tenantID, building1)
	if err != nil {
		t.Fatalf("Failed to create building1: %v", err)
	}

	building2 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building B",
	}
	buildingID2, err := unitsRepo.CreateBuilding(context.Background(), tenantID, building2)
	if err != nil {
		t.Fatalf("Failed to create building2: %v", err)
	}

	// 测试查询所有楼栋
	req := ListBuildingsRequest{
		TenantID:  tenantID,
		BranchName: "",
	}
	resp, err := unitService.ListBuildings(context.Background(), req)
	if err != nil {
		t.Fatalf("ListBuildings failed: %v", err)
	}

	if len(resp.Items) < 2 {
		t.Fatalf("Expected at least 2 buildings, got %d", len(resp.Items))
	}

	// 验证返回的楼栋
	found1, found2 := false, false
	for _, b := range resp.Items {
		if b.BuildingID == buildingID1 {
			found1 = true
			if b.BuildingName != "Building A" {
				t.Errorf("Expected Building A, got %s", b.BuildingName)
			}
		}
		if b.BuildingID == buildingID2 {
			found2 = true
			if b.BuildingName != "Building B" {
				t.Errorf("Expected Building B, got %s", b.BuildingName)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("Expected to find both buildings, found1=%v, found2=%v", found1, found2)
	}
}

// TestUnitService_CreateBuilding_Success 测试创建楼栋成功
func TestUnitService_CreateBuilding_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 测试创建楼栋
	req := CreateBuildingRequest{
		TenantID:     tenantID,
		BranchName:    "BRANCH-1",
		BuildingName: "Test Building",
	}

	resp, err := unitService.CreateBuilding(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBuilding failed: %v", err)
	}

	if resp.BuildingID == "" {
		t.Fatal("Expected building_id, got empty string")
	}

	// 验证楼栋已创建
	building, err := unitsRepo.GetBuilding(context.Background(), tenantID, resp.BuildingID)
	if err != nil {
		t.Fatalf("Failed to get building: %v", err)
	}

	if building.BuildingName != "Test Building" {
		t.Errorf("Expected Building Name 'Test Building', got %s", building.BuildingName)
	}
}

// ============================================
// Unit 测试
// ============================================

// TestUnitService_ListUnits_Success 测试查询单元列表成功
func TestUnitService_ListUnits_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 创建测试数据
	unit1 := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Unit A",
		Building:   sql.NullString{String: "Building A", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID1, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit1)
	if err != nil {
		t.Fatalf("Failed to create unit1: %v", err)
	}

	unit2 := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Unit B",
		Building:   sql.NullString{String: "Building A", Valid: true},
		Floor:      sql.NullString{String: "2F", Valid: true},
		UnitNumber: "201",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID2, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit2)
	if err != nil {
		t.Fatalf("Failed to create unit2: %v", err)
	}

	// 测试查询所有单元
	req := ListUnitsRequest{
		TenantID: tenantID,
		Page:     1,
		Size:     100,
	}
	resp, err := unitService.ListUnits(context.Background(), req)
	if err != nil {
		t.Fatalf("ListUnits failed: %v", err)
	}

	if resp.Total < 2 {
		t.Fatalf("Expected at least 2 units, got %d", resp.Total)
	}

	// 验证返回的单元
	found1, found2 := false, false
	for _, u := range resp.Items {
		if u.UnitID == unitID1 {
			found1 = true
			if u.UnitName != "Unit A" {
				t.Errorf("Expected Unit Name 'Unit A', got %s", u.UnitName)
			}
		}
		if u.UnitID == unitID2 {
			found2 = true
			if u.UnitName != "Unit B" {
				t.Errorf("Expected Unit Name 'Unit B', got %s", u.UnitName)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("Expected to find both units, found1=%v, found2=%v", found1, found2)
	}
}

// TestUnitService_CreateUnit_Success 测试创建单元成功
func TestUnitService_CreateUnit_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 测试创建单元
	req := CreateUnitRequest{
		TenantID:   tenantID,
		BranchName:  "BRANCH-1",
		UnitName:   "Test Unit",
		Building:   "Test Building",
		Floor:      "1F",
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}

	resp, err := unitService.CreateUnit(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateUnit failed: %v", err)
	}

	if resp.UnitID == "" {
		t.Fatal("Expected unit_id, got empty string")
	}

	// 验证单元已创建
	unit, err := unitsRepo.GetUnit(context.Background(), tenantID, resp.UnitID)
	if err != nil {
		t.Fatalf("Failed to get unit: %v", err)
	}

	if unit.UnitName != "Test Unit" {
		t.Errorf("Expected Unit Name 'Test Unit', got %s", unit.UnitName)
	}
}

// TestUnitService_GetUnit_Success 测试获取单个单元成功
func TestUnitService_GetUnit_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Test Unit",
		Building:   sql.NullString{String: "Test Building", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试获取单元
	req := GetUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}

	resp, err := unitService.GetUnit(context.Background(), req)
	if err != nil {
		t.Fatalf("GetUnit failed: %v", err)
	}

	if resp.Unit.UnitID != unitID {
		t.Errorf("Expected Unit ID %s, got %s", unitID, resp.Unit.UnitID)
	}

	if resp.Unit.UnitName != "Test Unit" {
		t.Errorf("Expected Unit Name 'Test Unit', got %s", resp.Unit.UnitName)
	}
}

// TestUnitService_UpdateUnit_Success 测试更新单元成功
func TestUnitService_UpdateUnit_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Original Unit",
		Building:   sql.NullString{String: "Test Building", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试更新单元
	req := UpdateUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
		UnitName: "Updated Unit",
	}

	resp, err := unitService.UpdateUnit(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateUnit failed: %v", err)
	}

	if !resp.Success {
		t.Fatal("Expected success=true, got false")
	}

	// 验证单元已更新
	updatedUnit, err := unitsRepo.GetUnit(context.Background(), tenantID, unitID)
	if err != nil {
		t.Fatalf("Failed to get unit: %v", err)
	}

	if updatedUnit.UnitName != "Updated Unit" {
		t.Errorf("Expected Unit Name 'Updated Unit', got %s", updatedUnit.UnitName)
	}
}

// TestUnitService_DeleteUnit_Success 测试删除单元成功
func TestUnitService_DeleteUnit_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Test Unit",
		Building:   sql.NullString{String: "Test Building", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试删除单元
	req := DeleteUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}

	resp, err := unitService.DeleteUnit(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteUnit failed: %v", err)
	}

	if !resp.Success {
		t.Fatal("Expected success=true, got false")
	}

	// 验证单元已删除
	_, err = unitsRepo.GetUnit(context.Background(), tenantID, unitID)
	if err == nil {
		t.Fatal("Expected unit to be deleted, but it still exists")
	}
}

// ============================================
// Room 测试
// ============================================

// TestUnitService_CreateRoom_Success 测试创建房间成功
func TestUnitService_CreateRoom_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 先创建单元
	unit := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Test Unit",
		Building:   sql.NullString{String: "Test Building", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试创建房间
	req := CreateRoomRequest{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Test Room",
	}

	resp, err := unitService.CreateRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}

	if resp.RoomID == "" {
		t.Fatal("Expected room_id, got empty string")
	}

	// 验证房间已创建
	room, err := unitsRepo.GetRoom(context.Background(), tenantID, resp.RoomID)
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	if room.RoomName != "Test Room" {
		t.Errorf("Expected Room Name 'Test Room', got %s", room.RoomName)
	}
}

// ============================================
// Bed 测试
// ============================================

// TestUnitService_CreateBed_Success 测试创建床位成功
func TestUnitService_CreateBed_Success(t *testing.T) {
	db := setupTestDBForUnit(t)
	defer db.Close()

	tenantID := createTestTenantForUnit(t, db)
	defer cleanupTestDataForUnit(t, db, tenantID)

	// 创建 Service
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	unitService := NewUnitService(unitsRepo, getTestLoggerForUnit())

	// 先创建单元和房间
	unit := &domain.Unit{
		TenantID:   tenantID,
		BranchName:  sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:   "Test Unit",
		Building:   sql.NullString{String: "Test Building", Valid: true},
		Floor:      sql.NullString{String: "1F", Valid: true},
		UnitNumber: "101",
		UnitType:   "Facility",
		Timezone:   "America/Denver",
	}
	unitID, err := unitsRepo.CreateUnit(context.Background(), tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Test Room",
	}
	roomID, err := unitsRepo.CreateRoom(context.Background(), tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 测试创建床位
	req := CreateBedRequest{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Test Bed",
	}

	resp, err := unitService.CreateBed(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBed failed: %v", err)
	}

	if resp.BedID == "" {
		t.Fatal("Expected bed_id, got empty string")
	}

	// 验证床位已创建
	bed, err := unitsRepo.GetBed(context.Background(), tenantID, resp.BedID)
	if err != nil {
		t.Fatalf("Failed to get bed: %v", err)
	}

	if bed.BedName != "Test Bed" {
		t.Errorf("Expected Bed Name 'Test Bed', got %s", bed.BedName)
	}
}


