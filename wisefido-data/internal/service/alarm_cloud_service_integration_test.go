// +build integration

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"wisefido-data/internal/repository"
)

func TestAlarmCloudService_GetAlarmCloudConfig(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
	alarmCloudService := NewAlarmCloudService(alarmCloudRepo, nil, getTestLogger())

	// 测试查询告警配置
	req := GetAlarmCloudConfigRequest{
		TenantID: SystemTenantID,
		UserID:   "test-user",
		UserRole: "SystemAdmin",
	}

	resp, err := alarmCloudService.GetAlarmCloudConfig(ctx, req)
	if err != nil {
		t.Fatalf("GetAlarmCloudConfig failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetAlarmCloudConfig returned nil response")
	}

	t.Logf("GetAlarmCloudConfig success: tenant_id=%s, device_alarms=%s", resp.TenantID, string(resp.DeviceAlarms))
}

func TestAlarmCloudService_GetAlarmCloudConfig_WithFallback(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
	alarmCloudService := NewAlarmCloudService(alarmCloudRepo, nil, getTestLogger())

	// 测试查询不存在的租户配置（应该回退到系统默认配置）
	// 使用一个不存在的 tenant_id
	testTenantID := "00000000-0000-0000-0000-000000000999"
	req := GetAlarmCloudConfigRequest{
		TenantID: testTenantID,
		UserID:   "test-user",
		UserRole: "SystemAdmin",
	}

	resp, err := alarmCloudService.GetAlarmCloudConfig(ctx, req)
	if err != nil {
		t.Fatalf("GetAlarmCloudConfig failed: %v", err)
	}

	if resp == nil {
		t.Fatal("GetAlarmCloudConfig returned nil response")
	}

	// 如果系统默认配置存在，应该返回系统默认配置的 tenant_id
	// 如果系统默认配置不存在，应该返回请求的 tenant_id
	t.Logf("GetAlarmCloudConfig with fallback success: tenant_id=%s", resp.TenantID)
}

func TestAlarmCloudService_UpdateAlarmCloudConfig(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
	alarmCloudService := NewAlarmCloudService(alarmCloudRepo, nil, getTestLogger())

	// 使用一个测试租户ID（不是系统租户）
	testTenantID := "00000000-0000-0000-0000-000000000999"

	// 清理测试数据
	defer func() {
		_, _ = db.Exec(`DELETE FROM alarm_cloud WHERE tenant_id = $1`, testTenantID)
	}()

	// 测试更新告警配置
	deviceAlarmsJSON := json.RawMessage(`{"Radar": {"Fall": "EMERGENCY"}}`)
	req := UpdateAlarmCloudConfigRequest{
		TenantID:     testTenantID,
		UserID:       "test-user",
		UserRole:     "SystemAdmin",
		OfflineAlarm: func() *string { s := "WARNING"; return &s }(),
		DeviceAlarms: deviceAlarmsJSON,
	}

	resp, err := alarmCloudService.UpdateAlarmCloudConfig(ctx, req)
	if err != nil {
		t.Fatalf("UpdateAlarmCloudConfig failed: %v", err)
	}

	if resp == nil {
		t.Fatal("UpdateAlarmCloudConfig returned nil response")
	}

	if resp.TenantID != testTenantID {
		t.Fatalf("UpdateAlarmCloudConfig returned wrong tenant_id: expected %s, got %s", testTenantID, resp.TenantID)
	}

	t.Logf("UpdateAlarmCloudConfig success: tenant_id=%s", resp.TenantID)
}

func TestAlarmCloudService_UpdateAlarmCloudConfig_SystemTenant_ShouldFail(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
	alarmCloudService := NewAlarmCloudService(alarmCloudRepo, nil, getTestLogger())

	// 测试更新系统默认配置（应该失败）
	req := UpdateAlarmCloudConfigRequest{
		TenantID: SystemTenantID,
		UserID:   "test-user",
		UserRole: "SystemAdmin",
		OfflineAlarm: func() *string { s := "WARNING"; return &s }(),
	}

	_, err := alarmCloudService.UpdateAlarmCloudConfig(ctx, req)
	if err == nil {
		t.Fatal("UpdateAlarmCloudConfig should fail for system tenant")
	}

	if !strings.Contains(err.Error(), "cannot update system alarm cloud config") {
		t.Fatalf("UpdateAlarmCloudConfig returned wrong error: %v", err)
	}

	t.Logf("UpdateAlarmCloudConfig correctly rejected system tenant: %v", err)
}

