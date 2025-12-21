// +build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"wisefido-data/internal/domain"
)

// 创建测试租户（alarm_cloud只需要tenant_id）
func createTestTenantForAlarmCloud(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000986"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant AlarmCloud", "test-alarmcloud.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// 清理测试数据
func cleanupTestDataForAlarmCloud(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM alarm_cloud WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// AlarmCloudRepository 测试
// ============================================

func TestPostgresAlarmCloudRepository_GetAlarmCloud(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmCloud(t, db)
	defer cleanupTestDataForAlarmCloud(t, db, tenantID)

	repo := NewPostgresAlarmCloudRepository(db)
	ctx := context.Background()

	// 先创建一个alarm_cloud
	deviceAlarms := json.RawMessage(`{"Radar": {"Fall": "EMERGENCY", "Radar_LeftBed": "WARNING"}}`)
	alarmCloud := &domain.AlarmCloud{
		TenantID:     tenantID,
		OfflineAlarm: "WARNING",
		LowBattery:   "WARNING",
		DeviceFailure: "EMERGENCY",
		DeviceAlarms: deviceAlarms,
	}

	err := repo.UpsertAlarmCloud(ctx, tenantID, alarmCloud)
	if err != nil {
		t.Fatalf("UpsertAlarmCloud failed: %v", err)
	}

	// 测试：获取alarm_cloud
	got, err := repo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetAlarmCloud failed: %v", err)
	}

	if got.TenantID != tenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", tenantID, got.TenantID)
	}
	if got.OfflineAlarm != "WARNING" {
		t.Errorf("Expected OfflineAlarm 'WARNING', got '%s'", got.OfflineAlarm)
	}
	if string(got.DeviceAlarms) != string(deviceAlarms) {
		t.Errorf("Expected device_alarms '%s', got '%s'", string(deviceAlarms), string(got.DeviceAlarms))
	}

	t.Logf("✅ GetAlarmCloud test passed")
}

func TestPostgresAlarmCloudRepository_UpsertAlarmCloud(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmCloud(t, db)
	defer cleanupTestDataForAlarmCloud(t, db, tenantID)

	repo := NewPostgresAlarmCloudRepository(db)
	ctx := context.Background()

	// 测试：创建alarm_cloud
	deviceAlarms := json.RawMessage(`{"Radar": {"Fall": "EMERGENCY", "Radar_LeftBed": "WARNING"}}`)
	conditions := json.RawMessage(`{"heart_rate": {"EMERGENCY": {"ranges": [{"min": 0, "max": 44}]}}}`)
	alarmCloud := &domain.AlarmCloud{
		TenantID:     tenantID,
		OfflineAlarm: "WARNING",
		LowBattery:   "WARNING",
		DeviceFailure: "EMERGENCY",
		DeviceAlarms: deviceAlarms,
		Conditions:   conditions,
	}

	err := repo.UpsertAlarmCloud(ctx, tenantID, alarmCloud)
	if err != nil {
		t.Fatalf("UpsertAlarmCloud failed: %v", err)
	}

	// 验证创建成功
	got, err := repo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetAlarmCloud failed: %v", err)
	}

	if got.TenantID != tenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", tenantID, got.TenantID)
	}

	// 测试：更新alarm_cloud
	updatedDeviceAlarms := json.RawMessage(`{"Radar": {"Fall": "WARNING", "Radar_LeftBed": "WARNING"}, "SleepPad": {"SleepPad_LeftBed": "WARNING"}}`)
	alarmCloud.DeviceAlarms = updatedDeviceAlarms
	alarmCloud.OfflineAlarm = "EMERGENCY"

	err = repo.UpsertAlarmCloud(ctx, tenantID, alarmCloud)
	if err != nil {
		t.Fatalf("UpsertAlarmCloud update failed: %v", err)
	}

	// 验证更新成功
	got, err = repo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetAlarmCloud after update failed: %v", err)
	}

	if got.OfflineAlarm != "EMERGENCY" {
		t.Errorf("Expected updated OfflineAlarm 'EMERGENCY', got '%s'", got.OfflineAlarm)
	}
	if string(got.DeviceAlarms) != string(updatedDeviceAlarms) {
		t.Errorf("Expected updated device_alarms '%s', got '%s'", string(updatedDeviceAlarms), string(got.DeviceAlarms))
	}

	t.Logf("✅ UpsertAlarmCloud test passed")
}

func TestPostgresAlarmCloudRepository_GetSystemAlarmCloud(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresAlarmCloudRepository(db)
	ctx := context.Background()

	// 先创建系统默认模板（如果不存在）
	systemTenantID := SystemTenantID
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		systemTenantID, "System Tenant", "system.local",
	)
	if err != nil {
		t.Fatalf("Failed to create system tenant: %v", err)
	}

	deviceAlarms := json.RawMessage(`{"Radar": {"Fall": "EMERGENCY"}}`)
	systemAlarmCloud := &domain.AlarmCloud{
		TenantID:     systemTenantID,
		OfflineAlarm: "ERROR",
		LowBattery:   "WARNING",
		DeviceFailure: "ERROR",
		DeviceAlarms: deviceAlarms,
	}

	err = repo.UpsertAlarmCloud(ctx, systemTenantID, systemAlarmCloud)
	if err != nil {
		t.Fatalf("UpsertAlarmCloud for system failed: %v", err)
	}

	// 测试：获取系统默认模板
	got, err := repo.GetSystemAlarmCloud(ctx)
	if err != nil {
		t.Fatalf("GetSystemAlarmCloud failed: %v", err)
	}

	if got.TenantID != systemTenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", systemTenantID, got.TenantID)
	}
	if got.OfflineAlarm != "ERROR" {
		t.Errorf("Expected OfflineAlarm 'ERROR', got '%s'", got.OfflineAlarm)
	}

	t.Logf("✅ GetSystemAlarmCloud test passed")
}

func TestPostgresAlarmCloudRepository_DeleteAlarmCloud(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmCloud(t, db)
	defer cleanupTestDataForAlarmCloud(t, db, tenantID)

	repo := NewPostgresAlarmCloudRepository(db)
	ctx := context.Background()

	// 先创建一个alarm_cloud
	deviceAlarms := json.RawMessage(`{"Radar": {"Fall": "EMERGENCY"}}`)
	alarmCloud := &domain.AlarmCloud{
		TenantID:     tenantID,
		OfflineAlarm: "WARNING",
		DeviceAlarms: deviceAlarms,
	}

	err := repo.UpsertAlarmCloud(ctx, tenantID, alarmCloud)
	if err != nil {
		t.Fatalf("UpsertAlarmCloud failed: %v", err)
	}

	// 测试：删除alarm_cloud
	err = repo.DeleteAlarmCloud(ctx, tenantID)
	if err != nil {
		t.Fatalf("DeleteAlarmCloud failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetAlarmCloud(ctx, tenantID)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	// 测试：不能删除系统模板
	err = repo.DeleteAlarmCloud(ctx, SystemTenantID)
	if err == nil {
		t.Fatal("Expected error when deleting system alarm cloud, got nil")
	}

	t.Logf("✅ DeleteAlarmCloud test passed")
}

