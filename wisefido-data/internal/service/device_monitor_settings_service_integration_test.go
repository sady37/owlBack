// +build integration

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestDBForDeviceMonitorSettings 设置测试数据库
func setupTestDBForDeviceMonitorSettings(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// createTestTenantForDeviceMonitorSettings 创建测试租户
func createTestTenantForDeviceMonitorSettings(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000996"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test DeviceMonitorSettings Tenant", "test-monitor.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestDeviceStoreForDeviceMonitorSettings 创建测试设备库存
func createTestDeviceStoreForDeviceMonitorSettings(t *testing.T, db *sql.DB, tenantID, deviceType string) string {
	deviceStoreID := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (device_store_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id, device_type = EXCLUDED.device_type`,
		deviceStoreID, tenantID, deviceType, "SN-"+deviceStoreID[:8], "UID-"+deviceStoreID[:8],
	)
	require.NoError(t, err)
	return deviceStoreID
}

// createTestDeviceForDeviceMonitorSettings 创建测试设备
func createTestDeviceForDeviceMonitorSettings(t *testing.T, db *sql.DB, tenantID, deviceStoreID, deviceID, deviceName string) string {
	_, err := db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, status, business_access, monitoring_enabled)
		 VALUES ($1, $2, $3, $4, 'online', 'approved', true)
		 ON CONFLICT (device_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id, device_store_id = EXCLUDED.device_store_id, device_name = EXCLUDED.device_name`,
		deviceID, tenantID, deviceStoreID, deviceName,
	)
	require.NoError(t, err)
	return deviceID
}

// createTestAlarmDeviceForDeviceMonitorSettings 创建测试 alarm_device 记录
func createTestAlarmDeviceForDeviceMonitorSettings(t *testing.T, db *sql.DB, tenantID, deviceID string, monitorConfig json.RawMessage) {
	_, err := db.Exec(
		`INSERT INTO alarm_device (device_id, tenant_id, monitor_config)
		 VALUES ($1, $2, $3::jsonb)
		 ON CONFLICT (device_id) DO UPDATE SET monitor_config = EXCLUDED.monitor_config`,
		deviceID, tenantID, string(monitorConfig),
	)
	require.NoError(t, err)
}

// cleanupTestDataForDeviceMonitorSettings 清理测试数据
func cleanupTestDataForDeviceMonitorSettings(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM alarm_device WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM device_store WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestLoggerForDeviceMonitorSettings 获取测试日志记录器
func getTestLoggerForDeviceMonitorSettings() *zap.Logger {
	return getTestLogger()
}

// ============================================
// GetDeviceMonitorSettings 测试
// ============================================

// TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_Sleepace_Success 测试获取 Sleepace 设备监控配置成功
func TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_Sleepace_Success(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Sleepace")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Sleepace Device")

	// 创建 alarm_device 记录（带配置）
	monitorConfig := json.RawMessage(`{
		"alarms": {
			"SleepPad_LeftBed": {
				"level": "WARNING",
				"enabled": true,
				"threshold": {
					"start_hour": 22,
					"start_minute": 0,
					"end_hour": 6,
					"end_minute": 30,
					"duration": 300
				}
			},
			"HeartRate": {
				"level": "EMERGENCY",
				"enabled": true,
				"threshold": {
					"min": 60,
					"max": 100,
					"duration": 60
				}
			},
			"BreathRate": {
				"level": "WARNING",
				"enabled": true,
				"threshold": {
					"min": 12,
					"max": 20,
					"duration": 60
				}
			}
		}
	}`)
	createTestAlarmDeviceForDeviceMonitorSettings(t, db, tenantID, deviceID, monitorConfig)

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	req := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
	}

	resp, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Settings)

	// 验证返回的配置项
	assert.Equal(t, 22, resp.Settings["left_bed_start_hour"])
	assert.Equal(t, 0, resp.Settings["left_bed_start_minute"])
	assert.Equal(t, 6, resp.Settings["left_bed_end_hour"])
	assert.Equal(t, 30, resp.Settings["left_bed_end_minute"])
	assert.Equal(t, 300, resp.Settings["left_bed_duration"])
	assert.Equal(t, "WARNING", resp.Settings["left_bed_alarm_level"])

	assert.Equal(t, 60, resp.Settings["min_heart_rate"])
	assert.Equal(t, 100, resp.Settings["max_heart_rate"])
	assert.Equal(t, 60, resp.Settings["heart_rate_slow_duration"])
	assert.Equal(t, 60, resp.Settings["heart_rate_fast_duration"])
	assert.Equal(t, "EMERGENCY", resp.Settings["heart_rate_slow_alarm_level"])
	assert.Equal(t, "EMERGENCY", resp.Settings["heart_rate_fast_alarm_level"])

	assert.Equal(t, 12, resp.Settings["min_breath_rate"])
	assert.Equal(t, 20, resp.Settings["max_breath_rate"])
	assert.Equal(t, 60, resp.Settings["breath_rate_slow_duration"])
	assert.Equal(t, 60, resp.Settings["breath_rate_fast_duration"])
	assert.Equal(t, "WARNING", resp.Settings["breath_rate_slow_alarm_level"])
	assert.Equal(t, "WARNING", resp.Settings["breath_rate_fast_alarm_level"])
}

// TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_Radar_Success 测试获取 Radar 设备监控配置成功
func TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_Radar_Success(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Radar")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Radar Device")

	// 创建 alarm_device 记录（带配置）
	monitorConfig := json.RawMessage(`{
		"alarms": {
			"Fall": {
				"level": "EMERGENCY",
				"enabled": true,
				"threshold": {
					"duration": 5
				}
			}
		}
	}`)
	createTestAlarmDeviceForDeviceMonitorSettings(t, db, tenantID, deviceID, monitorConfig)

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	req := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "radar",
	}

	resp, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Settings)

	// 验证返回的配置项
	assert.Equal(t, 5, resp.Settings["suspected_fall_duration"])
	assert.Equal(t, "EMERGENCY", resp.Settings["fall_alarm_level"])
}

// TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DefaultSettings 测试获取默认配置（当设备没有配置时）
func TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DefaultSettings(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备（但不创建 alarm_device 记录）
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Sleepace")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Sleepace Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	req := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
	}

	resp, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Settings)

	// 验证返回的是默认配置
	assert.Equal(t, 0, resp.Settings["left_bed_start_hour"])
	assert.Equal(t, "disabled", resp.Settings["left_bed_alarm_level"])
	assert.Equal(t, "disabled", resp.Settings["heart_rate_slow_alarm_level"])
}

// TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DeviceNotFound 测试设备不存在
func TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DeviceNotFound(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	req := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   uuid.New().String(), // 不存在的设备ID
		DeviceType: "sleepace",
	}

	_, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

// TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DeviceTypeMismatch 测试设备类型不匹配
func TestDeviceMonitorSettingsService_GetDeviceMonitorSettings_DeviceTypeMismatch(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建 Radar 设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Radar")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Radar Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	// 尝试获取 Sleepace 配置（但设备是 Radar）
	req := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
	}

	_, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device type mismatch")
}

// ============================================
// UpdateDeviceMonitorSettings 测试
// ============================================

// TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_Sleepace_Success 测试更新 Sleepace 设备监控配置成功
func TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_Sleepace_Success(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Sleepace")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Sleepace Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	// 更新配置
	req := UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
		Settings: map[string]interface{}{
			"left_bed_start_hour":        22,
			"left_bed_start_minute":      0,
			"left_bed_end_hour":          6,
			"left_bed_end_minute":        30,
			"left_bed_duration":          300,
			"left_bed_alarm_level":       "WARNING",
			"min_heart_rate":             60,
			"max_heart_rate":             100,
			"heart_rate_slow_duration":   60,
			"heart_rate_fast_duration":   60,
			"heart_rate_slow_alarm_level": "EMERGENCY",
			"heart_rate_fast_alarm_level": "EMERGENCY",
		},
	}

	resp, err := deviceMonitorSettingsService.UpdateDeviceMonitorSettings(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)

	// 验证配置已保存
	getReq := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
	}
	getResp, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), getReq)
	require.NoError(t, err)
	assert.Equal(t, 22, getResp.Settings["left_bed_start_hour"])
	assert.Equal(t, "WARNING", getResp.Settings["left_bed_alarm_level"])
	assert.Equal(t, 60, getResp.Settings["min_heart_rate"])
	assert.Equal(t, "EMERGENCY", getResp.Settings["heart_rate_slow_alarm_level"])
}

// TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_Radar_Success 测试更新 Radar 设备监控配置成功
func TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_Radar_Success(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Radar")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Radar Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	// 更新配置
	req := UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "radar",
		Settings: map[string]interface{}{
			"suspected_fall_duration": 5,
			"fall_alarm_level":       "EMERGENCY",
		},
	}

	resp, err := deviceMonitorSettingsService.UpdateDeviceMonitorSettings(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)

	// 验证配置已保存
	getReq := GetDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "radar",
	}
	getResp, err := deviceMonitorSettingsService.GetDeviceMonitorSettings(context.Background(), getReq)
	require.NoError(t, err)
	assert.Equal(t, 5, getResp.Settings["suspected_fall_duration"])
	assert.Equal(t, "EMERGENCY", getResp.Settings["fall_alarm_level"])
}

// TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_InvalidAlarmLevel 测试无效的报警级别
func TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_InvalidAlarmLevel(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建设备库存和设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Sleepace")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Sleepace Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	// 更新配置（使用无效的报警级别）
	req := UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
		Settings: map[string]interface{}{
			"left_bed_alarm_level": "INVALID_LEVEL",
		},
	}

	_, err := deviceMonitorSettingsService.UpdateDeviceMonitorSettings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid alarm level")
}

// TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_DeviceNotFound 测试设备不存在
func TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_DeviceNotFound(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	req := UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   uuid.New().String(), // 不存在的设备ID
		DeviceType: "sleepace",
		Settings: map[string]interface{}{
			"left_bed_alarm_level": "WARNING",
		},
	}

	_, err := deviceMonitorSettingsService.UpdateDeviceMonitorSettings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

// TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_DeviceTypeMismatch 测试设备类型不匹配
func TestDeviceMonitorSettingsService_UpdateDeviceMonitorSettings_DeviceTypeMismatch(t *testing.T) {
	db := setupTestDBForDeviceMonitorSettings(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForDeviceMonitorSettings(t, db)
	defer cleanupTestDataForDeviceMonitorSettings(t, db, tenantID)

	// 创建 Radar 设备
	deviceStoreID := createTestDeviceStoreForDeviceMonitorSettings(t, db, tenantID, "Radar")
	deviceID := uuid.New().String()
	createTestDeviceForDeviceMonitorSettings(t, db, tenantID, deviceStoreID, deviceID, "Test Radar Device")

	// 创建 Service
	alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
	deviceMonitorSettingsService := NewDeviceMonitorSettingsService(
		alarmDeviceRepo,
		devicesRepo,
		deviceStoreRepo,
		getTestLoggerForDeviceMonitorSettings(),
	)

	// 尝试更新 Sleepace 配置（但设备是 Radar）
	req := UpdateDeviceMonitorSettingsRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceType: "sleepace",
		Settings: map[string]interface{}{
			"left_bed_alarm_level": "WARNING",
		},
	}

	_, err := deviceMonitorSettingsService.UpdateDeviceMonitorSettings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device type mismatch")
}

