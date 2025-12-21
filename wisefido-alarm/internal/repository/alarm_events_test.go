package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"wisefido-alarm/internal/models"
)

func setupMockAlarmEventsDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *AlarmEventsRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	repo := NewAlarmEventsRepository(db, logger)

	return db, mock, repo
}

// ============================================
// 基础 CRUD 操作测试
// ============================================

func TestGetAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()
	deviceID := uuid.New().String()
	triggeredAt := time.Now()
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).AddRow(
		eventID, tenantID, deviceID, "Fall", "safety",
		"ALERT", "active", triggeredAt, nil,
		nil, `{"heart_rate": 120}`, nil, nil,
		nil, `[]`, `{}`, createdAt, updatedAt,
	)

	mock.ExpectQuery(`SELECT`).
		WithArgs(eventID, tenantID).
		WillReturnRows(rows)

	event, err := repo.GetAlarmEvent(ctx, tenantID, eventID)

	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, eventID, event.EventID)
	assert.Equal(t, tenantID, event.TenantID)
	assert.Equal(t, deviceID, event.DeviceID)
	assert.Equal(t, "Fall", event.EventType)
	assert.Equal(t, "safety", event.Category)
	assert.Equal(t, "ALERT", event.AlarmLevel)
	assert.Equal(t, "active", event.AlarmStatus)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAlarmEvent_NotFound(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()

	mock.ExpectQuery(`SELECT`).
		WithArgs(eventID, tenantID).
		WillReturnError(sql.ErrNoRows)

	event, err := repo.GetAlarmEvent(ctx, tenantID, eventID)

	assert.Error(t, err)
	assert.Nil(t, event)
	assert.Contains(t, err.Error(), "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAlarmEvent_InvalidTenantID(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	eventID := uuid.New().String()

	event, err := repo.GetAlarmEvent(ctx, "", eventID)

	assert.Error(t, err)
	assert.Nil(t, event)
	assert.Contains(t, err.Error(), "tenant_id is required")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()
	deviceID := uuid.New().String()
	now := time.Now()

	event := &models.AlarmEvent{
		EventID:      eventID,
		TenantID:     tenantID,
		DeviceID:     deviceID,
		EventType:    "Fall",
		Category:     "safety",
		AlarmLevel:   "ALERT",
		AlarmStatus:  "active",
		TriggeredAt:  now,
		TriggerData:  `{"heart_rate": 120}`,
		NotifiedUsers: `[]`,
		Metadata:     `{}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock.ExpectExec(`INSERT INTO alarm_events`).
		WithArgs(
			eventID, tenantID, deviceID, "Fall", "safety",
			"ALERT", "active", now, nil, nil,
			`{"heart_rate": 120}`, nil, nil, nil,
			`[]`, `{}`, now, now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateAlarmEvent(ctx, tenantID, event)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAlarmEvent_InvalidTenantID(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	event := &models.AlarmEvent{
		EventID: uuid.New().String(),
	}

	err := repo.CreateAlarmEvent(ctx, "", event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()
	handlerID := uuid.New().String()
	now := time.Now()

	updates := map[string]interface{}{
		"alarm_status": "acknowledged",
		"hand_time":    now,
		"handler":      handlerID,
	}

	mock.ExpectExec(`UPDATE alarm_events`).
		WithArgs("acknowledged", now, handlerID, eventID, tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateAlarmEvent(ctx, tenantID, eventID, updates)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAlarmEvent_NotFound(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()

	updates := map[string]interface{}{
		"alarm_status": "acknowledged",
	}

	mock.ExpectExec(`UPDATE alarm_events`).
		WithArgs("acknowledged", eventID, tenantID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateAlarmEvent(ctx, tenantID, eventID, updates)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()

	// 第一次查询 metadata
	metadataRows := sqlmock.NewRows([]string{"metadata"}).
		AddRow([]byte(`{}`))

	mock.ExpectQuery(`SELECT metadata`).
		WithArgs(eventID, tenantID).
		WillReturnRows(metadataRows)

	// 更新 metadata 设置 deleted_at
	mock.ExpectExec(`UPDATE alarm_events`).
		WithArgs(sqlmock.AnyArg(), eventID, tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteAlarmEvent(ctx, tenantID, eventID)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAlarmEvent_NotFound(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()

	mock.ExpectQuery(`SELECT metadata`).
		WithArgs(eventID, tenantID).
		WillReturnError(sql.ErrNoRows)

	err := repo.DeleteAlarmEvent(ctx, tenantID, eventID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// 查询操作测试
// ============================================

func TestListAlarmEvents_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID1 := uuid.New().String()
	eventID2 := uuid.New().String()
	now := time.Now()

	// Count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(tenantID).
		WillReturnRows(countRows)

	// List query
	listRows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).
		AddRow(eventID1, tenantID, uuid.New().String(), "Fall", "safety",
			"ALERT", "active", now, nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, now, now).
		AddRow(eventID2, tenantID, uuid.New().String(), "LeftBed", "behavioral",
			"WARNING", "active", now, nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, now, now)

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(tenantID, 20, 0).
		WillReturnRows(listRows)

	filters := AlarmEventFilters{}
	events, total, err := repo.ListAlarmEvents(ctx, tenantID, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, events, 2)
	assert.Equal(t, eventID1, events[0].EventID)
	assert.Equal(t, eventID2, events[1].EventID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAlarmEvents_WithFilters(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	deviceID := uuid.New().String()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	// Count query with filters
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(tenantID, startTime, endTime, deviceID).
		WillReturnRows(countRows)

	// List query with filters
	listRows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).
		AddRow(uuid.New().String(), tenantID, deviceID, "Fall", "safety",
			"ALERT", "active", time.Now(), nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(tenantID, startTime, endTime, deviceID, 20, 0).
		WillReturnRows(listRows)

	filters := AlarmEventFilters{
		StartTime: &startTime,
		EndTime:   &endTime,
		DeviceID:  &deviceID,
	}
	events, total, err := repo.ListAlarmEvents(ctx, tenantID, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, events, 1)
	assert.Equal(t, deviceID, events[0].DeviceID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecentAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	deviceID := uuid.New().String()
	eventID := uuid.New().String()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).AddRow(
		eventID, tenantID, deviceID, "Fall", "safety",
		"ALERT", "active", now, nil, nil,
		`{}`, nil, nil, nil, `[]`, `{}`, now, now,
	)

	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, deviceID, "Fall", sqlmock.AnyArg()).
		WillReturnRows(rows)

	event, err := repo.GetRecentAlarmEvent(ctx, tenantID, deviceID, "Fall", 5)

	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, eventID, event.EventID)
	assert.Equal(t, "Fall", event.EventType)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecentAlarmEvent_NotFound(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	deviceID := uuid.New().String()

	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, deviceID, "Fall", sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	event, err := repo.GetRecentAlarmEvent(ctx, tenantID, deviceID, "Fall", 5)

	require.NoError(t, err)
	assert.Nil(t, event)

	require.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// 状态管理测试
// ============================================

func TestAcknowledgeAlarmEvent_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()
	handlerID := uuid.New().String()

	mock.ExpectExec(`UPDATE alarm_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), eventID, tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AcknowledgeAlarmEvent(ctx, tenantID, eventID, handlerID)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAlarmEventOperation_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()
	handlerID := uuid.New().String()
	notes := "处理完成"

	mock.ExpectExec(`UPDATE alarm_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), eventID, tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateAlarmEventOperation(ctx, tenantID, eventID, "verified_and_processed", handlerID, &notes)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAlarmEventOperation_InvalidOperation(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	eventID := uuid.New().String()

	err := repo.UpdateAlarmEventOperation(ctx, tenantID, eventID, "invalid_operation", "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operation")

	require.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// 统计查询测试
// ============================================

func TestCountAlarmEvents_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(tenantID).
		WillReturnRows(countRows)

	filters := AlarmEventFilters{}
	count, err := repo.CountAlarmEvents(ctx, tenantID, filters)

	require.NoError(t, err)
	assert.Equal(t, 10, count)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAlarmEventsByDevice_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()
	deviceID := uuid.New().String()

	// Count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(tenantID, deviceID).
		WillReturnRows(countRows)

	// List query
	listRows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).
		AddRow(uuid.New().String(), tenantID, deviceID, "Fall", "safety",
			"ALERT", "active", time.Now(), nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(tenantID, deviceID, 20, 0).
		WillReturnRows(listRows)

	filters := AlarmEventFilters{}
	events, total, err := repo.GetAlarmEventsByDevice(ctx, tenantID, deviceID, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, events, 1)
	assert.Equal(t, deviceID, events[0].DeviceID)

	require.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// 高级功能测试
// ============================================

func TestGetActiveAlarmEvents_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()

	// Count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(countRows)

	// List query
	listRows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).
		AddRow(uuid.New().String(), tenantID, uuid.New().String(), "Fall", "safety",
			"ALERT", "active", time.Now(), nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(listRows)

	filters := AlarmEventFilters{}
	events, total, err := repo.GetActiveAlarmEvents(ctx, tenantID, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, events, 1)
	assert.Equal(t, "active", events[0].AlarmStatus)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInformationalEvents_Success(t *testing.T) {
	db, mock, repo := setupMockAlarmEventsDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New().String()

	// Count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(countRows)

	// List query
	listRows := sqlmock.NewRows([]string{
		"event_id", "tenant_id", "device_id", "event_type", "category",
		"alarm_level", "alarm_status", "triggered_at", "hand_time",
		"iot_timeseries_id", "trigger_data", "handler", "operation",
		"notes", "notified_users", "metadata", "created_at", "updated_at",
	}).
		AddRow(uuid.New().String(), tenantID, uuid.New().String(), "DeviceOnline", "device",
			"INFO", "active", time.Now(), nil, nil,
			`{}`, nil, nil, nil, `[]`, `{}`, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(listRows)

	filters := AlarmEventFilters{}
	events, total, err := repo.GetInformationalEvents(ctx, tenantID, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, events, 1)
	assert.Equal(t, "INFO", events[0].AlarmLevel)

	require.NoError(t, mock.ExpectationsWereMet())
}

