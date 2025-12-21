// +build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// ============================================
// IoTTimeSeriesRepository 测试
// ============================================

func TestPostgresIoTTimeSeriesRepository_GetTimeSeriesData(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresIoTTimeSeriesRepository(db)
	ctx := context.Background()

	// 创建测试数据
	tenantID := createTestTenant(t, db)
	deviceID := createTestDevice(t, db, tenantID)

	// 插入测试数据
	tsID := insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "activity", nil, nil)

	// 测试：获取时序数据
	data, err := repo.GetTimeSeriesData(ctx, tsID)
	if err != nil {
		t.Fatalf("GetTimeSeriesData failed: %v", err)
	}

	if data.ID != tsID {
		t.Errorf("Expected ID %d, got %d", tsID, data.ID)
	}
	if data.TenantID != tenantID {
		t.Errorf("Expected tenant_id %s, got %s", tenantID, data.TenantID)
	}
	if data.DeviceID != deviceID {
		t.Errorf("Expected device_id %s, got %s", deviceID, data.DeviceID)
	}
	if data.DataType != "observation" {
		t.Errorf("Expected data_type 'observation', got '%s'", data.DataType)
	}

	t.Logf("✅ GetTimeSeriesData test passed: ID=%d", tsID)
}

func TestPostgresIoTTimeSeriesRepository_GetLatestData(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresIoTTimeSeriesRepository(db)
	ctx := context.Background()

	// 创建测试数据
	tenantID := createTestTenant(t, db)
	deviceID := createTestDevice(t, db, tenantID)

	// 插入多条测试数据
	insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "activity", nil, nil)
	time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	tsID2 := insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "vital-signs", nil, nil)

	// 测试：获取最新数据
	results, err := repo.GetLatestData(ctx, tenantID, deviceID, 10)
	if err != nil {
		t.Fatalf("GetLatestData failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	// 验证结果按时间倒序
	if results[0].ID != tsID2 {
		t.Errorf("Expected latest ID %d, got %d", tsID2, results[0].ID)
	}

	t.Logf("✅ GetLatestData test passed: count=%d", len(results))
}

func TestPostgresIoTTimeSeriesRepository_GetDataByDevice(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresIoTTimeSeriesRepository(db)
	ctx := context.Background()

	// 创建测试数据
	tenantID := createTestTenant(t, db)
	deviceID := createTestDevice(t, db, tenantID)

	// 插入测试数据
	insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "activity", nil, nil)
	insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "vital-signs", nil, nil)

	// 测试：按设备查询（无过滤）
	results, total, err := repo.GetDataByDevice(ctx, tenantID, deviceID, nil, 1, 20)
	if err != nil {
		t.Fatalf("GetDataByDevice failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 total, got %d", total)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	// 测试：按设备查询（带过滤）
	filters := &IoTTimeSeriesFilters{
		DataType: "observation",
		Category: "activity",
	}
	resultsFiltered, totalFiltered, err := repo.GetDataByDevice(ctx, tenantID, deviceID, filters, 1, 20)
	if err != nil {
		t.Fatalf("GetDataByDevice with filters failed: %v", err)
	}

	for _, r := range resultsFiltered {
		if r.DataType != "observation" {
			t.Errorf("Expected data_type 'observation', got '%s'", r.DataType)
		}
		if r.Category != "activity" {
			t.Errorf("Expected category 'activity', got '%s'", r.Category)
		}
	}

	t.Logf("✅ GetDataByDevice test passed: total=%d, filtered=%d", total, totalFiltered)
}

func TestPostgresIoTTimeSeriesRepository_GetDataByTimeRange(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresIoTTimeSeriesRepository(db)
	ctx := context.Background()

	// 创建测试数据
	tenantID := createTestTenant(t, db)
	deviceID := createTestDevice(t, db, tenantID)

	// 插入测试数据
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)

	insertTestIoTTimeSeries(t, db, tenantID, deviceID, "observation", "activity", nil, nil)

	// 测试：时间范围查询
	_, total, err := repo.GetDataByTimeRange(ctx, tenantID, startTime, endTime, nil, 1, 20)
	if err != nil {
		t.Fatalf("GetDataByTimeRange failed: %v", err)
	}

	if total == 0 {
		t.Error("Expected at least 1 result in time range, got 0")
	}

	t.Logf("✅ GetDataByTimeRange test passed: total=%d", total)
}

func TestPostgresIoTTimeSeriesRepository_GetDataByLocation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPostgresIoTTimeSeriesRepository(db)
	ctx := context.Background()

	// 创建测试数据
	tenantID := createTestTenant(t, db)
	unitID := createTestUnit(t, db, tenantID)
	roomID := createTestRoom(t, db, tenantID, unitID)
	deviceID := createTestDevice(t, db, tenantID)

	// 更新设备位置
	_, err := db.Exec(`
		UPDATE devices
		SET bound_room_id = $1
		WHERE device_id = $2
	`, roomID, deviceID)
	if err != nil {
		t.Fatalf("Failed to update device location: %v", err)
	}

	// 插入测试数据（带位置信息）
	insertTestIoTTimeSeriesWithLocation(t, db, tenantID, deviceID, unitID, roomID, "observation", "activity", nil, nil)

	// 测试：按位置查询
	unitIDPtr := &unitID
	results, total, err := repo.GetDataByLocation(ctx, tenantID, unitIDPtr, nil, nil, 1, 20)
	if err != nil {
		t.Fatalf("GetDataByLocation failed: %v", err)
	}

	if total == 0 {
		t.Error("Expected at least 1 result for location, got 0")
	}

	for _, r := range results {
		if r.UnitID != unitID {
			t.Errorf("Expected unit_id %s, got %s", unitID, r.UnitID)
		}
	}

	t.Logf("✅ GetDataByLocation test passed: total=%d", total)
}

// ============================================
// 测试辅助函数
// ============================================

func createTestDevice(t *testing.T, db *sql.DB, tenantID string) string {
	// 先创建 device_store
	deviceStoreID := "00000000-0000-0000-0000-000000000099"
	_, err := db.Exec(
		`INSERT INTO device_store (device_store_id, tenant_id, device_type, device_model, serial_number, allow_access)
		 VALUES ($1, $2, $3, $4, $5, true)
		 ON CONFLICT (device_store_id) DO UPDATE SET device_type = EXCLUDED.device_type`,
		deviceStoreID, tenantID, "Radar", "Radar-001", "TEST-SN-001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device_store: %v", err)
	}

	// 创建 device
	deviceID := "00000000-0000-0000-0000-000000000100"
	_, err = db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, serial_number, uid, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'online')
		 ON CONFLICT (device_id) DO UPDATE SET serial_number = EXCLUDED.serial_number`,
		deviceID, tenantID, deviceStoreID, "Test Device", "TEST-SN-001", "TEST-UID-001",
	)
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	return deviceID
}

func insertTestIoTTimeSeries(t *testing.T, db *sql.DB, tenantID, deviceID, dataType, category string, heartRate, respRate *int) int64 {
	return insertTestIoTTimeSeriesWithLocation(t, db, tenantID, deviceID, "", "", dataType, category, heartRate, respRate)
}

func insertTestIoTTimeSeriesWithLocation(t *testing.T, db *sql.DB, tenantID, deviceID, unitID, roomID, dataType, category string, heartRate, respRate *int) int64 {
	var unitIDVal, roomIDVal interface{}
	if unitID != "" {
		unitIDVal = unitID
	}
	if roomID != "" {
		roomIDVal = roomID
	}

	var heartRateVal, respRateVal interface{}
	if heartRate != nil {
		heartRateVal = *heartRate
	}
	if respRate != nil {
		respRateVal = *respRate
	}

	query := `
		INSERT INTO iot_timeseries (
			tenant_id,
			device_id,
			timestamp,
			data_type,
			category,
			raw_original,
			raw_format,
			unit_id,
			room_id,
			heart_rate,
			respiratory_rate
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var id int64
	err := db.QueryRow(
		query,
		tenantID,
		deviceID,
		time.Now(),
		dataType,
		category,
		[]byte(`{"test": "data"}`),
		"json",
		unitIDVal,
		roomIDVal,
		heartRateVal,
		respRateVal,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Failed to insert test iot_timeseries: %v", err)
	}

	return id
}

