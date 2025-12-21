// +build integration

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// setupTestDBForAlarmEvent 设置测试数据库
func setupTestDBForAlarmEvent(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// createTestTenantForAlarmEvent 创建测试租户
func createTestTenantForAlarmEvent(t *testing.T, db *sql.DB) string {
	tenantID := "00000000-0000-0000-0000-000000000997"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test AlarmEvent Tenant", "test-alarm.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return tenantID
}

// createTestDeviceForAlarmEvent 创建测试设备
func createTestDeviceForAlarmEvent(t *testing.T, db *sql.DB, tenantID, deviceID, deviceName string) string {
	_, err := db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_name, device_type, status)
		 VALUES ($1, $2, $3, 'sensor', 'active')
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name, status = EXCLUDED.status`,
		deviceID, tenantID, deviceName,
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}
	return deviceID
}

// createTestAlarmEvent 创建测试报警事件
func createTestAlarmEvent(t *testing.T, db *sql.DB, tenantID, deviceID, eventType, category, alarmLevel string) *domain.AlarmEvent {
	eventID := uuid.New().String()
	now := time.Now()

	triggerData := json.RawMessage(`{"heart_rate": 120, "event_type": "` + eventType + `"}`)
	notifiedUsers := json.RawMessage(`[]`)
	metadata := json.RawMessage(`{"source": "test"}`)

	_, err := db.Exec(
		`INSERT INTO alarm_events (
			event_id, tenant_id, device_id, event_type, category, alarm_level, alarm_status,
			triggered_at, trigger_data, notified_users, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		eventID, tenantID, deviceID, eventType, category, alarmLevel, "active",
		now, triggerData, notifiedUsers, metadata, now, now,
	)
	if err != nil {
		t.Fatalf("Failed to create test alarm event: %v", err)
	}

	return &domain.AlarmEvent{
		EventID:      eventID,
		TenantID:     tenantID,
		DeviceID:     deviceID,
		EventType:    eventType,
		Category:     category,
		AlarmLevel:   alarmLevel,
		AlarmStatus:  "active",
		TriggeredAt:  now,
		TriggerData:  triggerData,
		NotifiedUsers: notifiedUsers,
		Metadata:     metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// createTestUserForAlarmEvent 创建测试用户
func createTestUserForAlarmEvent(t *testing.T, db *sql.DB, tenantID, userID, userAccount, role string) string {
	_, err := db.Exec(
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET role = EXCLUDED.role, status = 'active'`,
		userID, tenantID, userAccount, []byte("hash"), []byte("hash"), userAccount, role,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// cleanupTestDataForAlarmEvent 清理测试数据
func cleanupTestDataForAlarmEvent(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`UPDATE alarm_events SET metadata = jsonb_set(metadata, '{deleted_at}', to_jsonb(now()::text)) WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM users WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestLoggerForAlarmEvent 获取测试日志记录器
func getTestLoggerForAlarmEvent() *zap.Logger {
	return getTestLogger()
}

// ============================================
// ListAlarmEvents 测试
// ============================================

// TestAlarmEventService_ListAlarmEvents_Success 测试查询报警事件列表成功
func TestAlarmEventService_ListAlarmEvents_Success(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID1 := uuid.New().String()
	deviceID2 := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID1, "Device 1")
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID2, "Device 2")

	event1 := createTestAlarmEvent(t, db, tenantID, deviceID1, "Fall", "safety", "ALERT")
	event2 := createTestAlarmEvent(t, db, tenantID, deviceID2, "HeartRate", "clinical", "WARNING")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试查询所有报警事件
	req := ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		Page:            1,
		PageSize:        20,
	}

	resp, err := alarmEventService.ListAlarmEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListAlarmEvents failed: %v", err)
	}

	if resp.Pagination.Total < 2 {
		t.Fatalf("Expected at least 2 alarm events, got %d", resp.Pagination.Total)
	}

	// 验证返回的报警事件
	found1, found2 := false, false
	for _, item := range resp.Items {
		if item.EventID == event1.EventID {
			found1 = true
			if item.EventType != "Fall" {
				t.Errorf("Expected event_type 'Fall', got %s", item.EventType)
			}
			if item.Category != "safety" {
				t.Errorf("Expected category 'safety', got %s", item.Category)
			}
			if item.AlarmLevel != "ALERT" {
				t.Errorf("Expected alarm_level 'ALERT', got %s", item.AlarmLevel)
			}
		}
		if item.EventID == event2.EventID {
			found2 = true
			if item.EventType != "HeartRate" {
				t.Errorf("Expected event_type 'HeartRate', got %s", item.EventType)
			}
			if item.Category != "clinical" {
				t.Errorf("Expected category 'clinical', got %s", item.Category)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("Expected to find both alarm events, found1=%v, found2=%v", found1, found2)
	}
}

// TestAlarmEventService_ListAlarmEvents_WithStatusFilter 测试状态过滤
func TestAlarmEventService_ListAlarmEvents_WithStatusFilter(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event1 := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")
	
	// 创建一个已确认的报警事件
	eventID2 := uuid.New().String()
	now := time.Now()
	triggerData := json.RawMessage(`{"heart_rate": 120}`)
	notifiedUsers := json.RawMessage(`[]`)
	metadata := json.RawMessage(`{"source": "test"}`)
	_, err := db.Exec(
		`INSERT INTO alarm_events (
			event_id, tenant_id, device_id, event_type, category, alarm_level, alarm_status,
			triggered_at, trigger_data, notified_users, metadata, created_at, updated_at, hand_time, handler
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		eventID2, tenantID, deviceID, "HeartRate", "clinical", "WARNING", "acknowledged",
		now, triggerData, notifiedUsers, metadata, now, now, now, uuid.New().String(),
	)
	if err != nil {
		t.Fatalf("Failed to create acknowledged alarm event: %v", err)
	}

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试查询 active 状态的报警事件
	req := ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		Status:          "active",
		Page:            1,
		PageSize:        20,
	}

	resp, err := alarmEventService.ListAlarmEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListAlarmEvents failed: %v", err)
	}

	// 验证只返回 active 状态的报警事件
	for _, item := range resp.Items {
		if item.EventID == event1.EventID {
			if item.AlarmStatus != "active" {
				t.Errorf("Expected alarm_status 'active', got %s", item.AlarmStatus)
			}
		}
		if item.EventID == eventID2 {
			t.Errorf("Expected not to find acknowledged event, but found it")
		}
	}
}

// TestAlarmEventService_ListAlarmEvents_WithTimeRange 测试时间范围过滤
func TestAlarmEventService_ListAlarmEvents_WithTimeRange(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	// 创建一个在时间范围内的报警事件
	event1 := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")

	// 创建一个在时间范围外的报警事件（过去 2 小时）
	eventID2 := uuid.New().String()
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	triggerData := json.RawMessage(`{"heart_rate": 120}`)
	notifiedUsers := json.RawMessage(`[]`)
	metadata := json.RawMessage(`{"source": "test"}`)
	_, err := db.Exec(
		`INSERT INTO alarm_events (
			event_id, tenant_id, device_id, event_type, category, alarm_level, alarm_status,
			triggered_at, trigger_data, notified_users, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		eventID2, tenantID, deviceID, "HeartRate", "clinical", "WARNING", "active",
		twoHoursAgo, triggerData, notifiedUsers, metadata, twoHoursAgo, twoHoursAgo,
	)
	if err != nil {
		t.Fatalf("Failed to create old alarm event: %v", err)
	}

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试查询最近 1 小时内的报警事件
	oneHourAgo := time.Now().Add(-1 * time.Hour).Unix()
	now := time.Now().Unix()

	req := ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		AlarmTimeStart:  &oneHourAgo,
		AlarmTimeEnd:    &now,
		Page:            1,
		PageSize:        20,
	}

	resp, err := alarmEventService.ListAlarmEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListAlarmEvents failed: %v", err)
	}

	// 验证只返回时间范围内的报警事件
	found1, found2 := false, false
	for _, item := range resp.Items {
		if item.EventID == event1.EventID {
			found1 = true
		}
		if item.EventID == eventID2 {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("Expected to find event1 in time range")
	}
	if found2 {
		t.Errorf("Expected not to find event2 (outside time range)")
	}
}

// TestAlarmEventService_ListAlarmEvents_WithEventTypeFilter 测试事件类型过滤
func TestAlarmEventService_ListAlarmEvents_WithEventTypeFilter(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event1 := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")
	event2 := createTestAlarmEvent(t, db, tenantID, deviceID, "HeartRate", "clinical", "WARNING")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试查询特定事件类型
	req := ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		EventTypes:      []string{"Fall"},
		Page:            1,
		PageSize:        20,
	}

	resp, err := alarmEventService.ListAlarmEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListAlarmEvents failed: %v", err)
	}

	// 验证只返回指定事件类型的报警事件
	found1, found2 := false, false
	for _, item := range resp.Items {
		if item.EventID == event1.EventID {
			found1 = true
			if item.EventType != "Fall" {
				t.Errorf("Expected event_type 'Fall', got %s", item.EventType)
			}
		}
		if item.EventID == event2.EventID {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("Expected to find event1 (Fall)")
	}
	if found2 {
		t.Errorf("Expected not to find event2 (HeartRate)")
	}
}

// ============================================
// HandleAlarmEvent 测试
// ============================================

// TestAlarmEventService_HandleAlarmEvent_Acknowledge 测试确认报警事件
func TestAlarmEventService_HandleAlarmEvent_Acknowledge(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试确认报警事件
	req := HandleAlarmEventRequest{
		TenantID:        tenantID,
		EventID:         event.EventID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		AlarmStatus:     "acknowledged",
	}

	resp, err := alarmEventService.HandleAlarmEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleAlarmEvent failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证报警事件状态已更新
	updatedEvent, err := alarmEventsRepo.GetAlarmEvent(context.Background(), tenantID, event.EventID)
	if err != nil {
		t.Fatalf("Failed to get updated alarm event: %v", err)
	}

	if updatedEvent.AlarmStatus != "acknowledged" {
		t.Errorf("Expected alarm_status 'acknowledged', got %s", updatedEvent.AlarmStatus)
	}
	if updatedEvent.Handler == nil || *updatedEvent.Handler != userID {
		t.Errorf("Expected handler to be %s, got %v", userID, updatedEvent.Handler)
	}
	if updatedEvent.HandTime == nil {
		t.Errorf("Expected hand_time to be set")
	}
}

// TestAlarmEventService_HandleAlarmEvent_Resolve 测试解决报警事件
func TestAlarmEventService_HandleAlarmEvent_Resolve(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试解决报警事件
	req := HandleAlarmEventRequest{
		TenantID:        tenantID,
		EventID:         event.EventID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		AlarmStatus:     "resolved",
		HandleType:      "verified",
		Remarks:         "Test remarks",
	}

	resp, err := alarmEventService.HandleAlarmEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleAlarmEvent failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	// 验证报警事件操作已更新
	updatedEvent, err := alarmEventsRepo.GetAlarmEvent(context.Background(), tenantID, event.EventID)
	if err != nil {
		t.Fatalf("Failed to get updated alarm event: %v", err)
	}

	if updatedEvent.Operation == nil || *updatedEvent.Operation != "verified_and_processed" {
		t.Errorf("Expected operation 'verified_and_processed', got %v", updatedEvent.Operation)
	}
	if updatedEvent.Notes == nil || *updatedEvent.Notes != "Test remarks" {
		t.Errorf("Expected notes 'Test remarks', got %v", updatedEvent.Notes)
	}
	if updatedEvent.Handler == nil || *updatedEvent.Handler != userID {
		t.Errorf("Expected handler to be %s, got %v", userID, updatedEvent.Handler)
	}
}

// TestAlarmEventService_HandleAlarmEvent_InvalidStatus 测试无效状态转换
func TestAlarmEventService_HandleAlarmEvent_InvalidStatus(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试无效状态
	req := HandleAlarmEventRequest{
		TenantID:        tenantID,
		EventID:         event.EventID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		AlarmStatus:     "invalid_status",
	}

	_, err := alarmEventService.HandleAlarmEvent(context.Background(), req)
	if err == nil {
		t.Errorf("Expected error for invalid alarm_status, got nil")
	}
}

// TestAlarmEventService_HandleAlarmEvent_ResolveWithoutHandleType 测试解决报警事件但缺少 handle_type
func TestAlarmEventService_HandleAlarmEvent_ResolveWithoutHandleType(t *testing.T) {
	db := setupTestDBForAlarmEvent(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID := createTestTenantForAlarmEvent(t, db)
	defer cleanupTestDataForAlarmEvent(t, db, tenantID)

	// 创建测试数据
	deviceID := uuid.New().String()
	createTestDeviceForAlarmEvent(t, db, tenantID, deviceID, "Device 1")

	event := createTestAlarmEvent(t, db, tenantID, deviceID, "Fall", "safety", "ALERT")

	userID := uuid.New().String()
	createTestUserForAlarmEvent(t, db, tenantID, userID, "testuser", "Admin")

	// 创建 Service
	alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	unitsRepo := repository.NewPostgresUnitsRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	alarmEventService := NewAlarmEventService(alarmEventsRepo, devicesRepo, unitsRepo, usersRepo, db, getTestLoggerForAlarmEvent())

	// 测试解决报警事件但缺少 handle_type
	req := HandleAlarmEventRequest{
		TenantID:        tenantID,
		EventID:         event.EventID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		AlarmStatus:     "resolved",
		// HandleType 缺失
	}

	_, err := alarmEventService.HandleAlarmEvent(context.Background(), req)
	if err == nil {
		t.Errorf("Expected error for missing handle_type, got nil")
	}
}

