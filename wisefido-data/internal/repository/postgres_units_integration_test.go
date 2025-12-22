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

// 获取测试数据库连接
func getTestDBForUnits(t *testing.T) *sql.DB {
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
func createTestTenantForUnits(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant Units", "test-units.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForUnits(t *testing.T, db *sql.DB, tenantID string) {
	// 删除顺序：beds -> rooms -> units -> buildings -> tags_catalog -> tenants
	db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// Building 操作测试
// ============================================

func TestPostgresUnitsRepository_ListBuildings(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试数据
	building1 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID1, err := repo.CreateBuilding(ctx, tenantID, building1)
	if err != nil {
		t.Fatalf("Failed to create building1: %v", err)
	}

	building2 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building B",
	}
	buildingID2, err := repo.CreateBuilding(ctx, tenantID, building2)
	if err != nil {
		t.Fatalf("Failed to create building2: %v", err)
	}

	// 测试：查询所有楼栋
	buildings, err := repo.ListBuildings(ctx, tenantID, "")
	if err != nil {
		t.Fatalf("ListBuildings failed: %v", err)
	}
	if len(buildings) < 2 {
		t.Fatalf("Expected at least 2 buildings, got %d", len(buildings))
	}

	// 测试：按 branchTag 过滤
	buildingsFiltered, err := repo.ListBuildings(ctx, tenantID, "BRANCH-1")
	if err != nil {
		t.Fatalf("ListBuildings with filter failed: %v", err)
	}
	if len(buildingsFiltered) < 2 {
		t.Fatalf("Expected at least 2 buildings with BRANCH-1, got %d", len(buildingsFiltered))
	}

	// 验证返回的数据
	found1, found2 := false, false
	for _, b := range buildingsFiltered {
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
		t.Errorf("Not all buildings found: found1=%v, found2=%v", found1, found2)
	}
}

func TestPostgresUnitsRepository_GetBuilding(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试数据
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 测试：获取单个楼栋
	got, err := repo.GetBuilding(ctx, tenantID, buildingID)
	if err != nil {
		t.Fatalf("GetBuilding failed: %v", err)
	}
	if got.BuildingID != buildingID {
		t.Errorf("Expected buildingID %s, got %s", buildingID, got.BuildingID)
	}
	if got.BuildingName != "Building A" {
		t.Errorf("Expected Building A, got %s", got.BuildingName)
	}
	if !got.BranchTag.Valid || got.BranchTag.String != "BRANCH-1" {
		t.Errorf("Expected BRANCH-1, got %v", got.BranchTag)
	}

	// 测试：不存在的楼栋
	_, err = repo.GetBuilding(ctx, tenantID, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("Expected error for non-existent building")
	}
}

func TestPostgresUnitsRepository_CreateBuilding(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 测试：创建楼栋（带 branch_tag）
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("CreateBuilding failed: %v", err)
	}
	if buildingID == "" {
		t.Error("Expected non-empty buildingID")
	}

	// 验证：查询创建的楼栋
	got, err := repo.GetBuilding(ctx, tenantID, buildingID)
	if err != nil {
		t.Fatalf("GetBuilding failed: %v", err)
	}
	if got.BuildingName != "Building A" {
		t.Errorf("Expected Building A, got %s", got.BuildingName)
	}

	// 测试：创建楼栋（不带 branch_tag，使用 NULL）
	building2 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{Valid: false},
		BuildingName: "Building B",
	}
	buildingID2, err := repo.CreateBuilding(ctx, tenantID, building2)
	if err != nil {
		t.Fatalf("CreateBuilding without branch_tag failed: %v", err)
	}
	if buildingID2 == "" {
		t.Error("Expected non-empty buildingID2")
	}

	// 测试：验证标签同步到 tags_catalog
	var tagCount int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "BRANCH-1", "branch_tag",
	).Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to check tags_catalog: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("Expected 1 tag in catalog, got %d", tagCount)
	}

	// 测试：验证错误情况 - branch_tag 和 building_name 都为空
	building3 := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{Valid: false},
		BuildingName: "",
	}
	_, err = repo.CreateBuilding(ctx, tenantID, building3)
	if err == nil {
		t.Error("Expected error when both branch_tag and building_name are empty")
	}
}

func TestPostgresUnitsRepository_UpdateBuilding(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试数据
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 测试：更新楼栋名称
	building.BuildingName = "Building A Updated"
	err = repo.UpdateBuilding(ctx, tenantID, buildingID, building)
	if err != nil {
		t.Fatalf("UpdateBuilding failed: %v", err)
	}

	// 验证更新
	got, err := repo.GetBuilding(ctx, tenantID, buildingID)
	if err != nil {
		t.Fatalf("GetBuilding failed: %v", err)
	}
	if got.BuildingName != "Building A Updated" {
		t.Errorf("Expected Building A Updated, got %s", got.BuildingName)
	}

	// 测试：更新 branch_tag
	building.BranchTag = sql.NullString{String: "BRANCH-2", Valid: true}
	err = repo.UpdateBuilding(ctx, tenantID, buildingID, building)
	if err != nil {
		t.Fatalf("UpdateBuilding branch_tag failed: %v", err)
	}

	// 验证标签同步
	var tagCount int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "BRANCH-2", "branch_tag",
	).Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to check tags_catalog: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("Expected 1 tag in catalog, got %d", tagCount)
	}
}

func TestPostgresUnitsRepository_DeleteBuilding(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试数据
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	buildingID, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 测试：删除楼栋
	err = repo.DeleteBuilding(ctx, tenantID, buildingID)
	if err != nil {
		t.Fatalf("DeleteBuilding failed: %v", err)
	}

	// 验证：楼栋已删除
	_, err = repo.GetBuilding(ctx, tenantID, buildingID)
	if err == nil {
		t.Error("Expected error for deleted building")
	}
}

// ============================================
// Unit 操作测试
// ============================================

func TestPostgresUnitsRepository_ListUnits(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 创建测试数据
	unit1 := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID1, err := repo.CreateUnit(ctx, tenantID, unit1)
	if err != nil {
		t.Fatalf("Failed to create unit1: %v", err)
	}

	unit2 := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 102",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "102",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID2, err := repo.CreateUnit(ctx, tenantID, unit2)
	if err != nil {
		t.Fatalf("Failed to create unit2: %v", err)
	}

	// 测试：查询所有单元
	units, total, err := repo.ListUnits(ctx, tenantID, UnitFilters{}, 1, 100)
	if err != nil {
		t.Fatalf("ListUnits failed: %v", err)
	}
	if total < 2 {
		t.Fatalf("Expected at least 2 units, got total=%d", total)
	}
	if len(units) < 2 {
		t.Fatalf("Expected at least 2 units, got %d", len(units))
	}

	// 测试：按 branchTag 过滤
	filters := UnitFilters{BranchTag: "BRANCH-1"}
	unitsFiltered, totalFiltered, err := repo.ListUnits(ctx, tenantID, filters, 1, 100)
	if err != nil {
		t.Fatalf("ListUnits with filter failed: %v", err)
	}
	if totalFiltered < 2 {
		t.Fatalf("Expected at least 2 units with BRANCH-1, got total=%d", totalFiltered)
	}

	// 测试：搜索功能
	filtersSearch := UnitFilters{Search: "101"}
	unitsSearched, _, err := repo.ListUnits(ctx, tenantID, filtersSearch, 1, 100)
	if err != nil {
		t.Fatalf("ListUnits with search failed: %v", err)
	}
	found := false
	for _, u := range unitsSearched {
		if u.UnitID == unitID1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Search did not find unit 101")
	}

	// 验证返回的数据
	found1, found2 := false, false
	for _, u := range unitsFiltered {
		if u.UnitID == unitID1 {
			found1 = true
		}
		if u.UnitID == unitID2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("Not all units found: found1=%v, found2=%v", found1, found2)
	}
}

func TestPostgresUnitsRepository_GetUnit(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试：获取单个单元
	got, err := repo.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("GetUnit failed: %v", err)
	}
	if got.UnitID != unitID {
		t.Errorf("Expected unitID %s, got %s", unitID, got.UnitID)
	}
	if got.UnitName != "Unit 101" {
		t.Errorf("Expected Unit 101, got %s", got.UnitName)
	}
	if got.Building != "Building A" {
		t.Errorf("Expected Building A, got %s", got.Building)
	}

	// 测试：不存在的单元
	_, err = repo.GetUnit(ctx, tenantID, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("Expected error for non-existent unit")
	}
}

func TestPostgresUnitsRepository_CreateUnit(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 测试：创建单元
	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		AreaTag:      sql.NullString{String: "Area A", Valid: true},
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("CreateUnit failed: %v", err)
	}
	if unitID == "" {
		t.Error("Expected non-empty unitID")
	}

	// 验证：查询创建的单元
	got, err := repo.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("GetUnit failed: %v", err)
	}
	if got.UnitName != "Unit 101" {
		t.Errorf("Expected Unit 101, got %s", got.UnitName)
	}

	// 测试：验证标签同步到 tags_catalog
	var branchTagCount, areaTagCount int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "BRANCH-1", "branch_tag",
	).Scan(&branchTagCount)
	if err != nil {
		t.Fatalf("Failed to check branch_tag in catalog: %v", err)
	}
	if branchTagCount < 1 {
		t.Errorf("Expected at least 1 branch_tag in catalog, got %d", branchTagCount)
	}

	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "Area A", "area_tag",
	).Scan(&areaTagCount)
	if err != nil {
		t.Fatalf("Failed to check area_tag in catalog: %v", err)
	}
	if areaTagCount != 1 {
		t.Errorf("Expected 1 area_tag in catalog, got %d", areaTagCount)
	}

	// 测试：验证错误情况 - 必填字段缺失
	unitInvalid := &domain.Unit{
		TenantID: tenantID,
		// 缺少 unit_name
	}
	_, err = repo.CreateUnit(ctx, tenantID, unitInvalid)
	if err == nil {
		t.Error("Expected error when unit_name is missing")
	}

	// 测试：验证错误情况 - building 不存在
	unitInvalidBuilding := &domain.Unit{
		TenantID:   tenantID,
		UnitName:   "Unit 999",
		Building:   "Non-existent Building",
		UnitNumber: "999",
		UnitType:   "Facility",
		Timezone:   "America/Los_Angeles",
	}
	_, err = repo.CreateUnit(ctx, tenantID, unitInvalidBuilding)
	if err == nil {
		t.Error("Expected error when building does not exist")
	}
}

func TestPostgresUnitsRepository_UpdateUnit(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
		GroupList:    sql.NullString{String: `["group1"]`, Valid: true},
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试：更新单元名称
	unit.UnitName = "Unit 101 Updated"
	err = repo.UpdateUnit(ctx, tenantID, unitID, unit)
	if err != nil {
		t.Fatalf("UpdateUnit failed: %v", err)
	}

	// 验证更新
	got, err := repo.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("GetUnit failed: %v", err)
	}
	if got.UnitName != "Unit 101 Updated" {
		t.Errorf("Expected Unit 101 Updated, got %s", got.UnitName)
	}

	// 测试：更新 branch_tag
	unit.BranchTag = sql.NullString{String: "BRANCH-2", Valid: true}
	err = repo.UpdateUnit(ctx, tenantID, unitID, unit)
	if err != nil {
		t.Fatalf("UpdateUnit branch_tag failed: %v", err)
	}

	// 验证标签同步
	var tagCount int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2 AND tag_type = $3`,
		tenantID, "BRANCH-2", "branch_tag",
	).Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to check tags_catalog: %v", err)
	}
	if tagCount < 1 {
		t.Errorf("Expected at least 1 tag in catalog, got %d", tagCount)
	}

	// 测试：更新 groupList（验证同步到 cards）
	unit.GroupList = sql.NullString{String: `["group1", "group2"]`, Valid: true}
	err = repo.UpdateUnit(ctx, tenantID, unitID, unit)
	if err != nil {
		t.Fatalf("UpdateUnit groupList failed: %v", err)
	}

	// 注意：cards 表可能不存在测试数据，这里只验证没有错误
}

func TestPostgresUnitsRepository_DeleteUnit(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// 创建测试数据
	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试：删除单元
	err = repo.DeleteUnit(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("DeleteUnit failed: %v", err)
	}

	// 验证：单元已删除
	_, err = repo.GetUnit(ctx, tenantID, unitID)
	if err == nil {
		t.Error("Expected error for deleted unit")
	}
}

// ============================================
// Room 操作测试
// ============================================

func TestPostgresUnitsRepository_ListRooms(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 创建测试数据
	room1 := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID1, err := repo.CreateRoom(ctx, tenantID, unitID, room1)
	if err != nil {
		t.Fatalf("Failed to create room1: %v", err)
	}

	room2 := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 2",
	}
	roomID2, err := repo.CreateRoom(ctx, tenantID, unitID, room2)
	if err != nil {
		t.Fatalf("Failed to create room2: %v", err)
	}

	// 测试：查询房间列表
	rooms, err := repo.ListRooms(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("ListRooms failed: %v", err)
	}
	if len(rooms) < 2 {
		t.Fatalf("Expected at least 2 rooms, got %d", len(rooms))
	}

	// 验证返回的数据
	found1, found2 := false, false
	for _, r := range rooms {
		if r.RoomID == roomID1 {
			found1 = true
			if r.RoomName != "Room 1" {
				t.Errorf("Expected Room 1, got %s", r.RoomName)
			}
		}
		if r.RoomID == roomID2 {
			found2 = true
			if r.RoomName != "Room 2" {
				t.Errorf("Expected Room 2, got %s", r.RoomName)
			}
		}
	}
	if !found1 || !found2 {
		t.Errorf("Not all rooms found: found1=%v, found2=%v", found1, found2)
	}
}

func TestPostgresUnitsRepository_ListRoomsWithBeds(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 创建测试 room
	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 创建测试 bed
	bed := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed A",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID, err := repo.CreateBed(ctx, tenantID, roomID, bed)
	if err != nil {
		t.Fatalf("Failed to create bed: %v", err)
	}

	// 测试：查询房间及其床位
	roomsWithBeds, err := repo.ListRoomsWithBeds(ctx, tenantID, unitID)
	if err != nil {
		t.Fatalf("ListRoomsWithBeds failed: %v", err)
	}
	if len(roomsWithBeds) < 1 {
		t.Fatalf("Expected at least 1 room, got %d", len(roomsWithBeds))
	}

	// 验证返回的数据
	found := false
	for _, rwb := range roomsWithBeds {
		if rwb.Room.RoomID == roomID {
			found = true
			if len(rwb.Beds) < 1 {
				t.Errorf("Expected at least 1 bed, got %d", len(rwb.Beds))
			} else {
				if rwb.Beds[0].BedID != bedID {
					t.Errorf("Expected bedID %s, got %s", bedID, rwb.Beds[0].BedID)
				}
			}
			break
		}
	}
	if !found {
		t.Error("Room not found in ListRoomsWithBeds")
	}
}

func TestPostgresUnitsRepository_GetRoom(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 创建测试数据
	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 测试：获取单个房间
	got, err := repo.GetRoom(ctx, tenantID, roomID)
	if err != nil {
		t.Fatalf("GetRoom failed: %v", err)
	}
	if got.RoomID != roomID {
		t.Errorf("Expected roomID %s, got %s", roomID, got.RoomID)
	}
	if got.RoomName != "Room 1" {
		t.Errorf("Expected Room 1, got %s", got.RoomName)
	}

	// 测试：不存在的房间
	_, err = repo.GetRoom(ctx, tenantID, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("Expected error for non-existent room")
	}
}

func TestPostgresUnitsRepository_CreateRoom(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 测试：创建房间
	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}
	if roomID == "" {
		t.Error("Expected non-empty roomID")
	}

	// 验证：查询创建的房间
	got, err := repo.GetRoom(ctx, tenantID, roomID)
	if err != nil {
		t.Fatalf("GetRoom failed: %v", err)
	}
	if got.RoomName != "Room 1" {
		t.Errorf("Expected Room 1, got %s", got.RoomName)
	}

	// 测试：验证错误情况 - unit 不存在
	roomInvalid := &domain.Room{
		TenantID: tenantID,
		UnitID:   "00000000-0000-0000-0000-000000000000",
		RoomName: "Room Invalid",
	}
	_, err = repo.CreateRoom(ctx, tenantID, "00000000-0000-0000-0000-000000000000", roomInvalid)
	if err == nil {
		t.Error("Expected error when unit does not exist")
	}
}

func TestPostgresUnitsRepository_UpdateRoom(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 创建测试数据
	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 测试：更新房间名称
	room.RoomName = "Room 1 Updated"
	err = repo.UpdateRoom(ctx, tenantID, roomID, room)
	if err != nil {
		t.Fatalf("UpdateRoom failed: %v", err)
	}

	// 验证更新
	got, err := repo.GetRoom(ctx, tenantID, roomID)
	if err != nil {
		t.Fatalf("GetRoom failed: %v", err)
	}
	if got.RoomName != "Room 1 Updated" {
		t.Errorf("Expected Room 1 Updated, got %s", got.RoomName)
	}
}

func TestPostgresUnitsRepository_DeleteRoom(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building 和 unit
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// 创建测试数据
	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 测试：删除房间
	err = repo.DeleteRoom(ctx, tenantID, roomID)
	if err != nil {
		t.Fatalf("DeleteRoom failed: %v", err)
	}

	// 验证：房间已删除
	_, err = repo.GetRoom(ctx, tenantID, roomID)
	if err == nil {
		t.Error("Expected error for deleted room")
	}
}

// ============================================
// Bed 操作测试
// ============================================

func TestPostgresUnitsRepository_ListBeds(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building, unit, room
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 创建测试数据
	bed1 := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed A",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID1, err := repo.CreateBed(ctx, tenantID, roomID, bed1)
	if err != nil {
		t.Fatalf("Failed to create bed1: %v", err)
	}

	bed2 := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed B",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID2, err := repo.CreateBed(ctx, tenantID, roomID, bed2)
	if err != nil {
		t.Fatalf("Failed to create bed2: %v", err)
	}

	// 测试：查询床位列表
	beds, err := repo.ListBeds(ctx, tenantID, roomID)
	if err != nil {
		t.Fatalf("ListBeds failed: %v", err)
	}
	if len(beds) < 2 {
		t.Fatalf("Expected at least 2 beds, got %d", len(beds))
	}

	// 验证返回的数据
	found1, found2 := false, false
	for _, b := range beds {
		if b.BedID == bedID1 {
			found1 = true
			if b.BedName != "Bed A" {
				t.Errorf("Expected Bed A, got %s", b.BedName)
			}
		}
		if b.BedID == bedID2 {
			found2 = true
			if b.BedName != "Bed B" {
				t.Errorf("Expected Bed B, got %s", b.BedName)
			}
		}
	}
	if !found1 || !found2 {
		t.Errorf("Not all beds found: found1=%v, found2=%v", found1, found2)
	}
}

func TestPostgresUnitsRepository_GetBed(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building, unit, room
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 创建测试数据
	bed := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed A",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID, err := repo.CreateBed(ctx, tenantID, roomID, bed)
	if err != nil {
		t.Fatalf("Failed to create bed: %v", err)
	}

	// 测试：获取单个床位
	got, err := repo.GetBed(ctx, tenantID, bedID)
	if err != nil {
		t.Fatalf("GetBed failed: %v", err)
	}
	if got.BedID != bedID {
		t.Errorf("Expected bedID %s, got %s", bedID, got.BedID)
	}
	if got.BedName != "Bed A" {
		t.Errorf("Expected Bed A, got %s", got.BedName)
	}

	// 测试：不存在的床位
	_, err = repo.GetBed(ctx, tenantID, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("Expected error for non-existent bed")
	}
}

func TestPostgresUnitsRepository_CreateBed(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building, unit, room
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 测试：创建床位
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bed := &domain.Bed{
		TenantID:         tenantID,
		RoomID:           roomID,
		BedName:          "Bed A",
		MattressMaterial: sql.NullString{String: "Memory Foam", Valid: true},
	}
	bedID, err := repo.CreateBed(ctx, tenantID, roomID, bed)
	if err != nil {
		t.Fatalf("CreateBed failed: %v", err)
	}
	if bedID == "" {
		t.Error("Expected non-empty bedID")
	}

	// 验证：查询创建的床位
	got, err := repo.GetBed(ctx, tenantID, bedID)
	if err != nil {
		t.Fatalf("GetBed failed: %v", err)
	}
	if got.BedName != "Bed A" {
		t.Errorf("Expected Bed A, got %s", got.BedName)
	}
	if !got.MattressMaterial.Valid || got.MattressMaterial.String != "Memory Foam" {
		t.Errorf("Expected Memory Foam, got %v", got.MattressMaterial)
	}

	// 测试：验证错误情况 - room 不存在
	bedInvalid := &domain.Bed{
		TenantID: tenantID,
		RoomID:   "00000000-0000-0000-0000-000000000000",
		BedName:  "Bed Invalid",
		BedType:  "NonActiveBed",
	}
	_, err = repo.CreateBed(ctx, tenantID, "00000000-0000-0000-0000-000000000000", bedInvalid)
	if err == nil {
		t.Error("Expected error when room does not exist")
	}
}

func TestPostgresUnitsRepository_UpdateBed(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building, unit, room
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 创建测试数据
	bed := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed A",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID, err := repo.CreateBed(ctx, tenantID, roomID, bed)
	if err != nil {
		t.Fatalf("Failed to create bed: %v", err)
	}

	// 测试：更新床位名称
	bed.BedName = "Bed A Updated"
	err = repo.UpdateBed(ctx, tenantID, bedID, bed)
	if err != nil {
		t.Fatalf("UpdateBed failed: %v", err)
	}

	// 验证更新
	got, err := repo.GetBed(ctx, tenantID, bedID)
	if err != nil {
		t.Fatalf("GetBed failed: %v", err)
	}
	if got.BedName != "Bed A Updated" {
		t.Errorf("Expected Bed A Updated, got %s", got.BedName)
	}
}

func TestPostgresUnitsRepository_DeleteBed(t *testing.T) {
	db := getTestDBForUnits(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresUnitsRepository(db)
	ctx := context.Background()
	tenantID := createTestTenantForUnits(t, db)
	defer cleanupTestDataForUnits(t, db, tenantID)

	// 创建测试 building, unit, room
	building := &domain.Building{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		BuildingName: "Building A",
	}
	_, err := repo.CreateBuilding(ctx, tenantID, building)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	unit := &domain.Unit{
		TenantID:     tenantID,
		BranchTag:    sql.NullString{String: "BRANCH-1", Valid: true},
		UnitName:     "Unit 101",
		Building:     "Building A",
		Floor:        "1F",
		UnitNumber:   "101",
		UnitType:     "Facility",
		Timezone:     "America/Los_Angeles",
	}
	unitID, err := repo.CreateUnit(ctx, tenantID, unit)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	room := &domain.Room{
		TenantID: tenantID,
		UnitID:   unitID,
		RoomName: "Room 1",
	}
	roomID, err := repo.CreateRoom(ctx, tenantID, unitID, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 创建测试数据
	bed := &domain.Bed{
		TenantID: tenantID,
		RoomID:   roomID,
		BedName:  "Bed A",
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	}
	bedID, err := repo.CreateBed(ctx, tenantID, roomID, bed)
	if err != nil {
		t.Fatalf("Failed to create bed: %v", err)
	}

	// 测试：删除床位
	err = repo.DeleteBed(ctx, tenantID, bedID)
	if err != nil {
		t.Fatalf("DeleteBed failed: %v", err)
	}

	// 验证：床位已删除
	_, err = repo.GetBed(ctx, tenantID, bedID)
	if err == nil {
		t.Error("Expected error for deleted bed")
	}
}

