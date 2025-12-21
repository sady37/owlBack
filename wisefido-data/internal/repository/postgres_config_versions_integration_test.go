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

// 创建测试租户（config_versions只需要tenant_id）
func createTestTenantForConfigVersions(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000973"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant ConfigVersions", "test-configversions.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForConfigVersions(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM config_versions WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// ConfigVersionsRepository 测试
// ============================================

func TestPostgresConfigVersionsRepository_GetConfigVersion(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	// 先创建一个config_version
	entityID := "00000000-0000-0000-0000-000000000972"
	configData := json.RawMessage(`{"layout": {"width": 100, "height": 200}}`)
	validFrom := time.Now()
	configVersion := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: "room_layout",
		EntityID:   entityID,
		ConfigData: configData,
		ValidFrom:  validFrom,
	}

	versionID, err := repo.CreateConfigVersion(ctx, tenantID, configVersion)
	if err != nil {
		t.Fatalf("CreateConfigVersion failed: %v", err)
	}

	// 测试：获取config_version
	got, err := repo.GetConfigVersion(ctx, tenantID, versionID)
	if err != nil {
		t.Fatalf("GetConfigVersion failed: %v", err)
	}

	if got.VersionID != versionID {
		t.Errorf("Expected version_id '%s', got '%s'", versionID, got.VersionID)
	}
	if got.ConfigType != "room_layout" {
		t.Errorf("Expected config_type 'room_layout', got '%s'", got.ConfigType)
	}
	if got.EntityID != entityID {
		t.Errorf("Expected entity_id '%s', got '%s'", entityID, got.EntityID)
	}
	if string(got.ConfigData) != string(configData) {
		t.Errorf("Expected config_data '%s', got '%s'", string(configData), string(got.ConfigData))
	}

	t.Logf("✅ GetConfigVersion test passed")
}

func TestPostgresConfigVersionsRepository_GetConfigVersionAtTime(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	entityID := "00000000-0000-0000-0000-000000000971"
	configType := "device_config"

	// 创建第一个版本（valid_from = now - 2 hours, valid_to = now - 1 hour）
	configData1 := json.RawMessage(`{"config": {"param1": "value1"}}`)
	validFrom1 := time.Now().Add(-2 * time.Hour)
	validTo1 := time.Now().Add(-1 * time.Hour)
	configVersion1 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData1,
		ValidFrom:  validFrom1,
		ValidTo:    &validTo1,
	}

	versionID1, err := repo.CreateConfigVersion(ctx, tenantID, configVersion1)
	if err != nil {
		t.Fatalf("CreateConfigVersion 1 failed: %v", err)
	}

	// 创建第二个版本（valid_from = now - 1 hour, valid_to = nil）
	configData2 := json.RawMessage(`{"config": {"param1": "value2"}}`)
	validFrom2 := time.Now().Add(-1 * time.Hour)
	configVersion2 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData2,
		ValidFrom:  validFrom2,
		ValidTo:    nil, // 当前生效
	}

	versionID2, err := repo.CreateConfigVersion(ctx, tenantID, configVersion2)
	if err != nil {
		t.Fatalf("CreateConfigVersion 2 failed: %v", err)
	}

	// 测试：查询某个时间点的配置（应该返回第一个版本）
	atTime := time.Now().Add(-90 * time.Minute) // 在第一个版本的有效期内
	got, err := repo.GetConfigVersionAtTime(ctx, tenantID, configType, entityID, atTime)
	if err != nil {
		t.Fatalf("GetConfigVersionAtTime failed: %v", err)
	}

	if got.VersionID != versionID1 {
		t.Errorf("Expected version_id '%s' for time %v, got '%s'", versionID1, atTime, got.VersionID)
	}
	if string(got.ConfigData) != string(configData1) {
		t.Errorf("Expected config_data '%s', got '%s'", string(configData1), string(got.ConfigData))
	}

	// 测试：查询当前时间点的配置（应该返回第二个版本）
	atTimeNow := time.Now()
	got, err = repo.GetConfigVersionAtTime(ctx, tenantID, configType, entityID, atTimeNow)
	if err != nil {
		t.Fatalf("GetConfigVersionAtTime for now failed: %v", err)
	}

	if got.VersionID != versionID2 {
		t.Errorf("Expected version_id '%s' for current time, got '%s'", versionID2, got.VersionID)
	}
	if string(got.ConfigData) != string(configData2) {
		t.Errorf("Expected config_data '%s', got '%s'", string(configData2), string(got.ConfigData))
	}

	t.Logf("✅ GetConfigVersionAtTime test passed")
}

func TestPostgresConfigVersionsRepository_ListConfigVersions(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	entityID := "00000000-0000-0000-0000-000000000970"
	configType := "alarm_cloud"

	// 创建多个版本
	configData1 := json.RawMessage(`{"OfflineAlarm": "WARNING"}`)
	validFrom1 := time.Now().Add(-2 * time.Hour)
	validTo1 := time.Now().Add(-1 * time.Hour)
	configVersion1 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData1,
		ValidFrom:  validFrom1,
		ValidTo:    &validTo1,
	}
	_, err := repo.CreateConfigVersion(ctx, tenantID, configVersion1)
	if err != nil {
		t.Fatalf("CreateConfigVersion 1 failed: %v", err)
	}

	configData2 := json.RawMessage(`{"OfflineAlarm": "EMERGENCY"}`)
	validFrom2 := time.Now().Add(-1 * time.Hour)
	configVersion2 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData2,
		ValidFrom:  validFrom2,
		ValidTo:    nil,
	}
	_, err = repo.CreateConfigVersion(ctx, tenantID, configVersion2)
	if err != nil {
		t.Fatalf("CreateConfigVersion 2 failed: %v", err)
	}

	// 测试：列表查询（无过滤）
	versions, total, err := repo.ListConfigVersions(ctx, tenantID, configType, entityID, nil, 1, 20)
	if err != nil {
		t.Fatalf("ListConfigVersions failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 config versions, got total=%d", total)
	}
	if len(versions) < 2 {
		t.Errorf("Expected at least 2 config versions in result, got %d", len(versions))
	}

	// 验证按valid_from DESC排序
	if len(versions) >= 2 {
		if versions[0].ValidFrom.Before(versions[1].ValidFrom) {
			t.Error("Expected versions sorted by valid_from DESC")
		}
	}

	// 测试：按时间范围过滤
	startTime := time.Now().Add(-3 * time.Hour)
	endTime := time.Now()
	filters := &ConfigVersionFilters{StartTime: &startTime, EndTime: &endTime}
	versionsFiltered, _, err := repo.ListConfigVersions(ctx, tenantID, configType, entityID, filters, 1, 20)
	if err != nil {
		t.Fatalf("ListConfigVersions with time filter failed: %v", err)
	}
	if len(versionsFiltered) < 2 {
		t.Errorf("Expected at least 2 config versions in time range, got %d", len(versionsFiltered))
	}

	t.Logf("✅ ListConfigVersions test passed: total=%d", total)
}

func TestPostgresConfigVersionsRepository_CreateConfigVersion(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	entityID := "00000000-0000-0000-0000-000000000969"
	configType := "alarm_device"

	// 创建第一个版本（当前生效）
	configData1 := json.RawMessage(`{"monitor_config": {"alarms": {"Fall": "EMERGENCY"}}}`)
	validFrom1 := time.Now().Add(-1 * time.Hour)
	configVersion1 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData1,
		ValidFrom:  validFrom1,
		ValidTo:    nil, // 当前生效
	}

	versionID1, err := repo.CreateConfigVersion(ctx, tenantID, configVersion1)
	if err != nil {
		t.Fatalf("CreateConfigVersion 1 failed: %v", err)
	}

	if versionID1 == "" {
		t.Fatal("Expected non-empty version_id")
	}

	// 验证创建成功
	got1, err := repo.GetConfigVersion(ctx, tenantID, versionID1)
	if err != nil {
		t.Fatalf("GetConfigVersion 1 failed: %v", err)
	}

	if got1.ValidTo != nil {
		t.Error("Expected valid_to to be nil for first version")
	}

	// 创建第二个版本（应该自动将第一个版本的valid_to设置为当前时间）
	configData2 := json.RawMessage(`{"monitor_config": {"alarms": {"Fall": "WARNING"}}}`)
	validFrom2 := time.Now()
	configVersion2 := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: configType,
		EntityID:   entityID,
		ConfigData: configData2,
		ValidFrom:  validFrom2,
		ValidTo:    nil,
	}

	versionID2, err := repo.CreateConfigVersion(ctx, tenantID, configVersion2)
	if err != nil {
		t.Fatalf("CreateConfigVersion 2 failed: %v", err)
	}

	// 验证第一个版本的valid_to已被设置
	got1Updated, err := repo.GetConfigVersion(ctx, tenantID, versionID1)
	if err != nil {
		t.Fatalf("GetConfigVersion 1 after update failed: %v", err)
	}

	if got1Updated.ValidTo == nil {
		t.Error("Expected valid_to to be set for first version after creating second version")
	}
	if got1Updated.ValidTo != nil && got1Updated.ValidTo.After(validFrom2) {
		t.Error("Expected valid_to to be before or equal to valid_from of second version")
	}

	// 验证第二个版本创建成功
	got2, err := repo.GetConfigVersion(ctx, tenantID, versionID2)
	if err != nil {
		t.Fatalf("GetConfigVersion 2 failed: %v", err)
	}

	if got2.ValidTo != nil {
		t.Error("Expected valid_to to be nil for second version (current)")
	}

	t.Logf("✅ CreateConfigVersion test passed: versionID1=%s, versionID2=%s", versionID1, versionID2)
}

func TestPostgresConfigVersionsRepository_UpdateConfigVersion(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	// 先创建一个config_version
	entityID := "00000000-0000-0000-0000-000000000968"
	configData := json.RawMessage(`{"layout": {"width": 100}}`)
	validFrom := time.Now()
	configVersion := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: "room_layout",
		EntityID:   entityID,
		ConfigData: configData,
		ValidFrom:  validFrom,
	}

	versionID, err := repo.CreateConfigVersion(ctx, tenantID, configVersion)
	if err != nil {
		t.Fatalf("CreateConfigVersion failed: %v", err)
	}

	// 测试：更新config_version
	updatedConfigData := json.RawMessage(`{"layout": {"width": 200, "height": 300}}`)
	validTo := time.Now().Add(time.Hour)
	updatedConfigVersion := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: "room_layout",
		EntityID:   entityID,
		ConfigData: updatedConfigData,
		ValidFrom:  validFrom,
		ValidTo:    &validTo,
	}

	err = repo.UpdateConfigVersion(ctx, tenantID, versionID, updatedConfigVersion)
	if err != nil {
		t.Fatalf("UpdateConfigVersion failed: %v", err)
	}

	// 验证更新成功
	got, err := repo.GetConfigVersion(ctx, tenantID, versionID)
	if err != nil {
		t.Fatalf("GetConfigVersion after update failed: %v", err)
	}

	if string(got.ConfigData) != string(updatedConfigData) {
		t.Errorf("Expected updated config_data '%s', got '%s'", string(updatedConfigData), string(got.ConfigData))
	}
	if got.ValidTo == nil || !got.ValidTo.Equal(validTo) {
		t.Errorf("Expected valid_to %v, got %v", validTo, got.ValidTo)
	}

	t.Logf("✅ UpdateConfigVersion test passed")
}

func TestPostgresConfigVersionsRepository_DeleteConfigVersion(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForConfigVersions(t, db)
	defer cleanupTestDataForConfigVersions(t, db, tenantID)

	repo := NewPostgresConfigVersionsRepository(db)
	ctx := context.Background()

	// 先创建一个config_version
	entityID := "00000000-0000-0000-0000-000000000967"
	configData := json.RawMessage(`{"config": {"param": "value"}}`)
	validFrom := time.Now()
	configVersion := &domain.ConfigVersion{
		TenantID:   tenantID,
		ConfigType: "device_config",
		EntityID:   entityID,
		ConfigData: configData,
		ValidFrom:  validFrom,
	}

	versionID, err := repo.CreateConfigVersion(ctx, tenantID, configVersion)
	if err != nil {
		t.Fatalf("CreateConfigVersion failed: %v", err)
	}

	// 测试：删除config_version
	err = repo.DeleteConfigVersion(ctx, tenantID, versionID)
	if err != nil {
		t.Fatalf("DeleteConfigVersion failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetConfigVersion(ctx, tenantID, versionID)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	t.Logf("✅ DeleteConfigVersion test passed")
}

