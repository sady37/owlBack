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

// setupTestDBForDevice 设置测试数据库（复用 getTestDBForService）
func setupTestDBForDevice(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// createTestTenantForDevice 创建测试租户
func createTestTenantForDevice(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Device Tenant", "test-device.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestDeviceStoreForDevice 创建测试设备库存
func createTestDeviceStoreForDevice(t *testing.T, db *sql.DB, tenantID string) string {
	deviceStoreID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (device_store_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id, device_type = EXCLUDED.device_type, serial_number = EXCLUDED.serial_number, uid = EXCLUDED.uid`,
		deviceStoreID, tenantID, "Radar", "TEST-SERIAL-001", "TEST-UID-001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device store: %v", err)
	}
	return deviceStoreID
}

// createTestDeviceForDevice 创建测试设备
func createTestDeviceForDevice(t *testing.T, db *sql.DB, tenantID, deviceStoreID string) string {
	deviceID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, serial_number, uid, status, business_access, monitoring_enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, 'online', 'approved', true)
		 ON CONFLICT (device_id) DO UPDATE SET
		   tenant_id = EXCLUDED.tenant_id,
		   device_store_id = EXCLUDED.device_store_id,
		   device_name = EXCLUDED.device_name,
		   serial_number = EXCLUDED.serial_number,
		   uid = EXCLUDED.uid,
		   status = EXCLUDED.status,
		   business_access = EXCLUDED.business_access,
		   monitoring_enabled = EXCLUDED.monitoring_enabled`,
		deviceID, tenantID, deviceStoreID, "Test Device", "TEST-SERIAL-001", "TEST-UID-001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}
	return deviceID
}

// cleanupTestDataForDevice 清理测试数据
func cleanupTestDataForDevice(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM device_store WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestLoggerForDevice 获取测试日志记录器（复用 getTestLogger）
func getTestLoggerForDevice() *zap.Logger {
	return getTestLogger()
}

// TestDeviceService_ListDevices_Success 测试查询设备列表成功
func TestDeviceService_ListDevices_Success(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	deviceID := createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	// 创建 Service
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试查询设备列表
	req := ListDevicesRequest{
		TenantID: tenantID,
		Page:     1,
		Size:     20,
	}

	resp, err := deviceService.ListDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	if resp == nil {
		t.Fatal("ListDevices returned nil response")
	}

	if len(resp.Items) == 0 {
		t.Fatal("ListDevices returned empty items")
	}

	found := false
	for _, item := range resp.Items {
		if item.DeviceID == deviceID {
			found = true
			if item.DeviceName != "Test Device" {
				t.Errorf("Expected device_name 'Test Device', got '%s'", item.DeviceName)
			}
			if item.Status != "online" {
				t.Errorf("Expected status 'online', got '%s'", item.Status)
			}
			if item.BusinessAccess != "approved" {
				t.Errorf("Expected business_access 'approved', got '%s'", item.BusinessAccess)
			}
			break
		}
	}

	if !found {
		t.Fatal("Test device not found in list")
	}

	if resp.Total < 1 {
		t.Errorf("Expected total >= 1, got %d", resp.Total)
	}
}

// TestDeviceService_ListDevices_WithFilters 测试带过滤条件的查询
func TestDeviceService_ListDevices_WithFilters(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	deviceID := createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	// 创建 Service
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试按状态过滤
	req := ListDevicesRequest{
		TenantID: tenantID,
		Status:   []string{"online"},
		Page:     1,
		Size:     20,
	}

	resp, err := deviceService.ListDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	found := false
	for _, item := range resp.Items {
		if item.DeviceID == deviceID {
			found = true
			if item.Status != "online" {
				t.Errorf("Expected status 'online', got '%s'", item.Status)
			}
			break
		}
	}

	if !found {
		t.Fatal("Test device not found in filtered list")
	}

	// 测试按业务访问权限过滤
	req = ListDevicesRequest{
		TenantID:       tenantID,
		BusinessAccess: "approved",
		Page:           1,
		Size:           20,
	}

	resp, err = deviceService.ListDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	found = false
	for _, item := range resp.Items {
		if item.DeviceID == deviceID {
			found = true
			if item.BusinessAccess != "approved" {
				t.Errorf("Expected business_access 'approved', got '%s'", item.BusinessAccess)
			}
			break
		}
	}

	if !found {
		t.Fatal("Test device not found in filtered list")
	}
}

// TestDeviceService_ListDevices_MissingTenantID 测试缺少 tenant_id
func TestDeviceService_ListDevices_MissingTenantID(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	req := ListDevicesRequest{
		Page: 1,
		Size: 20,
	}

	_, err := deviceService.ListDevices(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing tenant_id")
	}

	if err.Error() != "tenant_id is required" {
		t.Errorf("Expected error 'tenant_id is required', got '%s'", err.Error())
	}
}

// TestDeviceService_GetDevice_Success 测试查询设备详情成功
func TestDeviceService_GetDevice_Success(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	deviceID := createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	// 创建 Service
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试查询设备详情
	req := GetDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := deviceService.GetDevice(context.Background(), req)
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetDevice returned nil response")
	}

	if resp.Device == nil {
		t.Fatal("GetDevice returned nil device")
	}

	if resp.Device.DeviceID != deviceID {
		t.Errorf("Expected device_id '%s', got '%s'", deviceID, resp.Device.DeviceID)
	}

	if resp.Device.DeviceName != "Test Device" {
		t.Errorf("Expected device_name 'Test Device', got '%s'", resp.Device.DeviceName)
	}

	if resp.Device.TenantID != tenantID {
		t.Errorf("Expected tenant_id '%s', got '%s'", tenantID, resp.Device.TenantID)
	}
}

// TestDeviceService_GetDevice_NotFound 测试设备不存在
func TestDeviceService_GetDevice_NotFound(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	req := GetDeviceRequest{
		TenantID: tenantID,
		DeviceID: "00000000-0000-0000-0000-000000000000",
	}

	_, err := deviceService.GetDevice(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for device not found")
	}

	if err.Error() != "device not found" {
		t.Errorf("Expected error 'device not found', got '%s'", err.Error())
	}
}

// TestDeviceService_GetDevice_MissingTenantID 测试缺少 tenant_id
func TestDeviceService_GetDevice_MissingTenantID(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	req := GetDeviceRequest{
		DeviceID: "00000000-0000-0000-0000-000000000000",
	}

	_, err := deviceService.GetDevice(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing tenant_id")
	}

	if err.Error() != "tenant_id is required" {
		t.Errorf("Expected error 'tenant_id is required', got '%s'", err.Error())
	}
}

// TestDeviceService_UpdateDevice_Success 测试更新设备成功
func TestDeviceService_UpdateDevice_Success(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	deviceID := createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	// 创建 Service
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试更新设备
	device := &domain.Device{
		DeviceName:     "Updated Device Name",
		Status:         "offline",
		BusinessAccess: "pending",
	}

	req := UpdateDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Device:   device,
	}

	resp, err := deviceService.UpdateDevice(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateDevice failed: %v", err)
	}

	if resp == nil {
		t.Fatal("UpdateDevice returned nil response")
	}

	if !resp.Success {
		t.Error("UpdateDevice returned success=false")
	}

	// 验证更新结果
	getReq := GetDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	getResp, err := deviceService.GetDevice(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}

	if getResp.Device.DeviceName != "Updated Device Name" {
		t.Errorf("Expected device_name 'Updated Device Name', got '%s'", getResp.Device.DeviceName)
	}

	if getResp.Device.Status != "offline" {
		t.Errorf("Expected status 'offline', got '%s'", getResp.Device.Status)
	}

	if getResp.Device.BusinessAccess != "pending" {
		t.Errorf("Expected business_access 'pending', got '%s'", getResp.Device.BusinessAccess)
	}
}

// TestDeviceService_UpdateDevice_MissingTenantID 测试缺少 tenant_id
func TestDeviceService_UpdateDevice_MissingTenantID(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	req := UpdateDeviceRequest{
		DeviceID: "00000000-0000-0000-0000-000000000000",
		Device:   &domain.Device{},
	}

	_, err := deviceService.UpdateDevice(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing tenant_id")
	}

	if err.Error() != "tenant_id is required" {
		t.Errorf("Expected error 'tenant_id is required', got '%s'", err.Error())
	}
}

// TestDeviceService_DeleteDevice_Success 测试删除设备成功
func TestDeviceService_DeleteDevice_Success(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	deviceID := createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	// 创建 Service
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试删除设备
	req := DeleteDeviceRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := deviceService.DeleteDevice(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}

	if resp == nil {
		t.Fatal("DeleteDevice returned nil response")
	}

	if !resp.Success {
		t.Error("DeleteDevice returned success=false")
	}

	// 验证设备已被禁用（软删除）
	// 注意：ListDevices 会自动过滤 status='disabled' 的设备
	listReq := ListDevicesRequest{
		TenantID: tenantID,
		Page:     1,
		Size:     20,
	}

	listResp, err := deviceService.ListDevices(context.Background(), listReq)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	found := false
	for _, item := range listResp.Items {
		if item.DeviceID == deviceID {
			found = true
			break
		}
	}

	if found {
		t.Error("Device should be disabled and not appear in list")
	}
}

// TestDeviceService_DeleteDevice_MissingTenantID 测试缺少 tenant_id
func TestDeviceService_DeleteDevice_MissingTenantID(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	req := DeleteDeviceRequest{
		DeviceID: "00000000-0000-0000-0000-000000000000",
	}

	_, err := deviceService.DeleteDevice(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing tenant_id")
	}

	if err.Error() != "tenant_id is required" {
		t.Errorf("Expected error 'tenant_id is required', got '%s'", err.Error())
	}
}

// TestDeviceService_ListDevices_StatusCommaSeparated 测试 status 参数逗号分隔
func TestDeviceService_ListDevices_StatusCommaSeparated(t *testing.T) {
	db := setupTestDBForDevice(t)
	defer db.Close()

	tenantID := createTestTenantForDevice(t, db)
	defer cleanupTestDataForDevice(t, db, tenantID)

	deviceStoreID := createTestDeviceStoreForDevice(t, db, tenantID)
	createTestDeviceForDevice(t, db, tenantID, deviceStoreID)

	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceService := NewDeviceService(devicesRepo, getTestLoggerForDevice())

	// 测试逗号分隔的 status
	req := ListDevicesRequest{
		TenantID: tenantID,
		Status:   []string{"online,offline"},
		Page:     1,
		Size:     20,
	}

	resp, err := deviceService.ListDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	// 应该能正常处理，不会报错
	if resp == nil {
		t.Fatal("ListDevices returned nil response")
	}
}

