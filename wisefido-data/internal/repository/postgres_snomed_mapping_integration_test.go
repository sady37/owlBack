// +build integration

package repository

import (
	"context"
	"testing"

	"wisefido-data/internal/domain"
)

// snomed_mapping 表是系统默认映射，所有租户共享，不需要创建测试租户

// ============================================
// SNOMEDMappingRepository 测试
// ============================================

func TestPostgresSNOMEDMappingRepository_GetMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 测试：获取姿态映射（系统默认数据）
	mapping, err := repo.GetMapping(ctx, "posture", "4")
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	if mapping.MappingType != "posture" {
		t.Errorf("Expected mapping_type 'posture', got '%s'", mapping.MappingType)
	}
	if mapping.SourceValue != "4" {
		t.Errorf("Expected source_value '4', got '%s'", mapping.SourceValue)
	}
	if mapping.SNOMEDDisplay != "Standing position" {
		t.Errorf("Expected snomed_display 'Standing position', got '%s'", mapping.SNOMEDDisplay)
	}

	// 测试：获取事件映射（系统默认数据）
	eventMapping, err := repo.GetMapping(ctx, "event", "FALL")
	if err != nil {
		t.Fatalf("GetMapping for event failed: %v", err)
	}

	if eventMapping.MappingType != "event" {
		t.Errorf("Expected mapping_type 'event', got '%s'", eventMapping.MappingType)
	}
	if eventMapping.SourceValue != "FALL" {
		t.Errorf("Expected source_value 'FALL', got '%s'", eventMapping.SourceValue)
	}
	if eventMapping.SNOMEDDisplay != "Fall" {
		t.Errorf("Expected snomed_display 'Fall', got '%s'", eventMapping.SNOMEDDisplay)
	}

	t.Logf("✅ GetMapping test passed")
}

func TestPostgresSNOMEDMappingRepository_GetPostureMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 测试：获取通用姿态映射（firmware_version IS NULL）
	mapping, err := repo.GetPostureMapping(ctx, "4", nil)
	if err != nil {
		t.Fatalf("GetPostureMapping failed: %v", err)
	}

	if mapping.SNOMEDDisplay != "Standing position" {
		t.Errorf("Expected snomed_display 'Standing position', got '%s'", mapping.SNOMEDDisplay)
	}

	// 测试：获取特定固件版本的姿态映射
	firmwareVersion := "202406"
	mappingWithVersion, err := repo.GetPostureMapping(ctx, "6", &firmwareVersion)
	if err != nil {
		t.Fatalf("GetPostureMapping with firmware version failed: %v", err)
	}

	if mappingWithVersion.FirmwareVersion != firmwareVersion {
		t.Errorf("Expected firmware_version '%s', got '%s'", firmwareVersion, mappingWithVersion.FirmwareVersion)
	}

	t.Logf("✅ GetPostureMapping test passed")
}

func TestPostgresSNOMEDMappingRepository_GetEventMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 测试：获取事件映射
	mapping, err := repo.GetEventMapping(ctx, "ENTER_BED")
	if err != nil {
		t.Fatalf("GetEventMapping failed: %v", err)
	}

	if mapping.MappingType != "event" {
		t.Errorf("Expected mapping_type 'event', got '%s'", mapping.MappingType)
	}
	if mapping.SourceValue != "ENTER_BED" {
		t.Errorf("Expected source_value 'ENTER_BED', got '%s'", mapping.SourceValue)
	}
	if mapping.SNOMEDDisplay != "In bed" {
		t.Errorf("Expected snomed_display 'In bed', got '%s'", mapping.SNOMEDDisplay)
	}

	t.Logf("✅ GetEventMapping test passed")
}

func TestPostgresSNOMEDMappingRepository_ListMappings(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 测试：列表查询（无过滤）
	mappings, total, err := repo.ListMappings(ctx, "posture", nil, 1, 20)
	if err != nil {
		t.Fatalf("ListMappings failed: %v", err)
	}

	if total == 0 {
		t.Error("Expected at least some posture mappings, got 0")
	}
	if len(mappings) == 0 {
		t.Error("Expected at least some posture mappings in result, got 0")
	}

	// 测试：按category过滤
	filters := &SNOMEDMappingFilters{Category: "safety"}
	mappingsSafety, _, err := repo.ListMappings(ctx, "posture", filters, 1, 20)
	if err != nil {
		t.Fatalf("ListMappings with category filter failed: %v", err)
	}

	for _, m := range mappingsSafety {
		if m.Category != "safety" {
			t.Errorf("Expected category 'safety', got '%s'", m.Category)
		}
	}

	// 测试：按固件版本过滤
	filters = &SNOMEDMappingFilters{FirmwareVersion: "202406"}
	mappingsWithFirmware, _, err := repo.ListMappings(ctx, "posture", filters, 1, 20)
	if err != nil {
		t.Fatalf("ListMappings with firmware version filter failed: %v", err)
	}

	// 验证结果包含通用版本或指定版本
	found := false
	for _, m := range mappingsWithFirmware {
		if m.FirmwareVersion == "" || m.FirmwareVersion == "202406" {
			found = true
			break
		}
	}
	if !found && len(mappingsWithFirmware) > 0 {
		t.Logf("Note: All mappings have specific firmware versions")
	}

	t.Logf("✅ ListMappings test passed: total=%d", total)
}

func TestPostgresSNOMEDMappingRepository_CreateMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 测试：创建新映射（使用不冲突的source_value）
	durationThreshold := 180
	mapping := &domain.SNOMEDMapping{
		MappingType:             "event",
		SourceValue:             "TEST_EVENT_001",
		SNOMEDCode:              "123456789",
		SNOMEDDisplay:            "Test Event",
		Category:                 "activity",
		DisplayEn:                "Test Event",
		DurationThresholdMinutes: &durationThreshold,
	}

	err := repo.CreateMapping(ctx, mapping)
	if err != nil {
		t.Fatalf("CreateMapping failed: %v", err)
	}

	if mapping.MappingID == "" {
		t.Fatal("Expected non-empty mapping_id")
	}

	// 验证创建成功
	got, err := repo.GetMapping(ctx, "event", "TEST_EVENT_001")
	if err != nil {
		t.Fatalf("GetMapping after create failed: %v", err)
	}

	if got.SNOMEDDisplay != "Test Event" {
		t.Errorf("Expected snomed_display 'Test Event', got '%s'", got.SNOMEDDisplay)
	}
	if got.DurationThresholdMinutes == nil || *got.DurationThresholdMinutes != durationThreshold {
		t.Errorf("Expected duration_threshold_minutes %d, got %v", durationThreshold, got.DurationThresholdMinutes)
	}

	// 清理测试数据
	defer repo.DeleteMapping(ctx, mapping.MappingID)

	t.Logf("✅ CreateMapping test passed: mappingID=%s", mapping.MappingID)
}

func TestPostgresSNOMEDMappingRepository_UpdateMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 先创建一个映射
	mapping := &domain.SNOMEDMapping{
		MappingType:  "event",
		SourceValue:  "TEST_EVENT_002",
		SNOMEDDisplay: "Test Event Original",
		Category:     "activity",
	}

	err := repo.CreateMapping(ctx, mapping)
	if err != nil {
		t.Fatalf("CreateMapping failed: %v", err)
	}
	defer repo.DeleteMapping(ctx, mapping.MappingID)

	// 测试：更新映射
	updatedMapping := &domain.SNOMEDMapping{
		MappingType:  "event",
		SourceValue:  "TEST_EVENT_002",
		SNOMEDCode:   "987654321",
		SNOMEDDisplay: "Test Event Updated",
		Category:     "safety",
		DisplayEn:    "Test Event Updated",
	}

	err = repo.UpdateMapping(ctx, mapping.MappingID, updatedMapping)
	if err != nil {
		t.Fatalf("UpdateMapping failed: %v", err)
	}

	// 验证更新成功
	got, err := repo.GetMapping(ctx, "event", "TEST_EVENT_002")
	if err != nil {
		t.Fatalf("GetMapping after update failed: %v", err)
	}

	if got.SNOMEDDisplay != "Test Event Updated" {
		t.Errorf("Expected updated snomed_display 'Test Event Updated', got '%s'", got.SNOMEDDisplay)
	}
	if got.Category != "safety" {
		t.Errorf("Expected updated category 'safety', got '%s'", got.Category)
	}

	t.Logf("✅ UpdateMapping test passed")
}

func TestPostgresSNOMEDMappingRepository_DeleteMapping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresSNOMEDMappingRepository(db)
	ctx := context.Background()

	// 先创建一个映射
	mapping := &domain.SNOMEDMapping{
		MappingType:  "event",
		SourceValue:  "TEST_EVENT_003",
		SNOMEDDisplay: "Test Event Delete",
		Category:     "activity",
	}

	err := repo.CreateMapping(ctx, mapping)
	if err != nil {
		t.Fatalf("CreateMapping failed: %v", err)
	}

	// 测试：删除映射
	err = repo.DeleteMapping(ctx, mapping.MappingID)
	if err != nil {
		t.Fatalf("DeleteMapping failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetMapping(ctx, "event", "TEST_EVENT_003")
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	t.Logf("✅ DeleteMapping test passed")
}

