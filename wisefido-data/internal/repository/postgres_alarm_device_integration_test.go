// +build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"wisefido-data/internal/domain"
)

// 创建测试租户和设备（alarm_device需要device_id）
func createTestTenantAndDeviceForAlarmDevice(t *testing.T, db *sql.DB) (string, string) {
	tenantID := "00000000-0000-0000-0000-000000000991"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name`,
		tenantID, "Test Tenant AlarmDevice", "test-alarmdevice.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// 创建测试device_store
	deviceStoreID := "00000000-0000-0000-0000-000000000990"
	_, err = db.Exec(
		`INSERT INTO device_store (device_store_id, tenant_id, device_type, device_model, serial_number, allow_access)
		 VALUES ($1, $2, $3, $4, $5, true)
		 ON CONFLICT (device_store_id) DO UPDATE SET device_type = EXCLUDED.device_type`,
		deviceStoreID, tenantID, "Radar", "Radar-001", "TEST-SN-001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device_store: %v", err)
	}

	// 创建测试device
	deviceID := "00000000-0000-0000-0000-000000000989"
	_, err = db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, status)
		 VALUES ($1, $2, $3, $4, 'active')
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name`,
		deviceID, tenantID, deviceStoreID, "Test Device 001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	return tenantID, deviceID
}

// 清理测试数据
func cleanupTestDataForAlarmDevice(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM alarm_device WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM device_store WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// ============================================
// AlarmDeviceRepository 测试
// ============================================

func TestPostgresAlarmDeviceRepository_GetAlarmDevice(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForAlarmDevice(t, db)
	defer cleanupTestDataForAlarmDevice(t, db, tenantID)

	repo := NewPostgresAlarmDeviceRepository(db)
	ctx := context.Background()

	// 先创建一个alarm_device
	monitorConfig := json.RawMessage(`{"alarms": {"Fall": {"level": "EMERGENCY", "enabled": true}}}`)
	alarmDevice := &domain.AlarmDevice{
		DeviceID:     deviceID,
		TenantID:     tenantID,
		MonitorConfig: monitorConfig,
	}

	err := repo.UpsertAlarmDevice(ctx, tenantID, deviceID, alarmDevice)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice failed: %v", err)
	}

	// 测试：获取alarm_device
	got, err := repo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		t.Fatalf("GetAlarmDevice failed: %v", err)
	}

	if got.DeviceID != deviceID {
		t.Errorf("Expected device_id '%s', got '%s'", deviceID, got.DeviceID)
	}
	if string(got.MonitorConfig) != string(monitorConfig) {
		t.Errorf("Expected monitor_config '%s', got '%s'", string(monitorConfig), string(got.MonitorConfig))
	}

	t.Logf("✅ GetAlarmDevice test passed")
}

func TestPostgresAlarmDeviceRepository_UpsertAlarmDevice(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForAlarmDevice(t, db)
	defer cleanupTestDataForAlarmDevice(t, db, tenantID)

	repo := NewPostgresAlarmDeviceRepository(db)
	ctx := context.Background()

	// 测试：创建alarm_device
	monitorConfig := json.RawMessage(`{"alarms": {"Fall": {"level": "EMERGENCY", "enabled": true}}}`)
	vendorConfig := json.RawMessage(`{"heart_rate": {"typical_range": [55, 95]}}`)
	alarmDevice := &domain.AlarmDevice{
		DeviceID:     deviceID,
		TenantID:     tenantID,
		MonitorConfig: monitorConfig,
		VendorConfig:  vendorConfig,
	}

	err := repo.UpsertAlarmDevice(ctx, tenantID, deviceID, alarmDevice)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice failed: %v", err)
	}

	// 验证创建成功
	got, err := repo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		t.Fatalf("GetAlarmDevice failed: %v", err)
	}

	if got.DeviceID != deviceID {
		t.Errorf("Expected device_id '%s', got '%s'", deviceID, got.DeviceID)
	}

	// 测试：更新alarm_device
	updatedMonitorConfig := json.RawMessage(`{"alarms": {"Fall": {"level": "WARNING", "enabled": true}}}`)
	alarmDevice.MonitorConfig = updatedMonitorConfig

	err = repo.UpsertAlarmDevice(ctx, tenantID, deviceID, alarmDevice)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice update failed: %v", err)
	}

	// 验证更新成功
	got, err = repo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		t.Fatalf("GetAlarmDevice after update failed: %v", err)
	}

	if string(got.MonitorConfig) != string(updatedMonitorConfig) {
		t.Errorf("Expected updated monitor_config '%s', got '%s'", string(updatedMonitorConfig), string(got.MonitorConfig))
	}

	t.Logf("✅ UpsertAlarmDevice test passed")
}

func TestPostgresAlarmDeviceRepository_DeleteAlarmDevice(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForAlarmDevice(t, db)
	defer cleanupTestDataForAlarmDevice(t, db, tenantID)

	repo := NewPostgresAlarmDeviceRepository(db)
	ctx := context.Background()

	// 先创建一个alarm_device
	monitorConfig := json.RawMessage(`{"alarms": {"Fall": {"level": "EMERGENCY", "enabled": true}}}`)
	alarmDevice := &domain.AlarmDevice{
		DeviceID:     deviceID,
		TenantID:     tenantID,
		MonitorConfig: monitorConfig,
	}

	err := repo.UpsertAlarmDevice(ctx, tenantID, deviceID, alarmDevice)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice failed: %v", err)
	}

	// 测试：删除alarm_device
	err = repo.DeleteAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		t.Fatalf("DeleteAlarmDevice failed: %v", err)
	}

	// 验证删除成功
	_, err = repo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}

	t.Logf("✅ DeleteAlarmDevice test passed")
}

func TestPostgresAlarmDeviceRepository_ListAlarmDevices(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID1 := createTestTenantAndDeviceForAlarmDevice(t, db)
	defer cleanupTestDataForAlarmDevice(t, db, tenantID)

	// 创建第二个device
	deviceStoreID := "00000000-0000-0000-0000-000000000988"
	_, err := db.Exec(
		`INSERT INTO device_store (device_store_id, tenant_id, device_type, device_model, serial_number, allow_access)
		 VALUES ($1, $2, $3, $4, $5, true)
		 ON CONFLICT (device_store_id) DO UPDATE SET device_type = EXCLUDED.device_type`,
		deviceStoreID, tenantID, "SleepPad", "SleepPad-001", "TEST-SN-002",
	)
	if err != nil {
		t.Fatalf("Failed to create test device_store 2: %v", err)
	}

	deviceID2 := "00000000-0000-0000-0000-000000000987"
	_, err = db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, status)
		 VALUES ($1, $2, $3, $4, 'active')
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name`,
		deviceID2, tenantID, deviceStoreID, "Test Device 002",
	)
	if err != nil {
		t.Fatalf("Failed to create test device 2: %v", err)
	}

	repo := NewPostgresAlarmDeviceRepository(db)
	ctx := context.Background()

	// 创建两个alarm_device
	monitorConfig1 := json.RawMessage(`{"alarms": {"Fall": {"level": "EMERGENCY", "enabled": true}}}`)
	alarmDevice1 := &domain.AlarmDevice{
		DeviceID:     deviceID1,
		TenantID:     tenantID,
		MonitorConfig: monitorConfig1,
	}
	err = repo.UpsertAlarmDevice(ctx, tenantID, deviceID1, alarmDevice1)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice 1 failed: %v", err)
	}

	monitorConfig2 := json.RawMessage(`{"alarms": {"SleepPad_LeftBed": {"level": "WARNING", "enabled": true}}}`)
	alarmDevice2 := &domain.AlarmDevice{
		DeviceID:     deviceID2,
		TenantID:     tenantID,
		MonitorConfig: monitorConfig2,
	}
	err = repo.UpsertAlarmDevice(ctx, tenantID, deviceID2, alarmDevice2)
	if err != nil {
		t.Fatalf("UpsertAlarmDevice 2 failed: %v", err)
	}

	// 测试：列表查询
	alarmDevices, total, err := repo.ListAlarmDevices(ctx, tenantID, 1, 20)
	if err != nil {
		t.Fatalf("ListAlarmDevices failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 alarm devices, got total=%d", total)
	}
	if len(alarmDevices) < 2 {
		t.Errorf("Expected at least 2 alarm devices in result, got %d", len(alarmDevices))
	}

	// 验证分页
	alarmDevicesPage1, totalPage1, err := repo.ListAlarmDevices(ctx, tenantID, 1, 1)
	if err != nil {
		t.Fatalf("ListAlarmDevices page 1 failed: %v", err)
	}
	if len(alarmDevicesPage1) != 1 {
		t.Errorf("Expected 1 alarm device on page 1, got %d", len(alarmDevicesPage1))
	}
	if totalPage1 != total {
		t.Errorf("Expected total=%d, got %d", total, totalPage1)
	}

	t.Logf("✅ ListAlarmDevices test passed: total=%d", total)
}

