// +build integration

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestDBForSleepace 设置测试数据库
func setupTestDBForSleepace(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// getTestLoggerForSleepace 获取测试日志记录器
func getTestLoggerForSleepace() *zap.Logger {
	return getTestLogger()
}

// createTestTenantAndDeviceForSleepace 创建测试租户和设备
func createTestTenantAndDeviceForSleepace(t *testing.T, db *sql.DB) (string, string) {
	tenantID := "00000000-0000-0000-0000-000000000999"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Sleepace Tenant", "test-sleepace.local",
	)
	require.NoError(t, err)

	// 创建测试设备
	deviceID := "00000000-0000-0000-0000-000000000001"
	_, err = db.Exec(
		`INSERT INTO devices (device_id, tenant_id, device_name, device_type, serial_number, status)
		 VALUES ($1, $2, $3, $4, $5, 'active')
		 ON CONFLICT (device_id) DO UPDATE SET device_name = EXCLUDED.device_name, status = EXCLUDED.status`,
		deviceID, tenantID, "Test Sleepace Device", "Sleepace", "SP001", "active",
	)
	require.NoError(t, err)

	return tenantID, deviceID
}

// cleanupTestDataForSleepace 清理测试数据
func cleanupTestDataForSleepace(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM sleepace_report WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM devices WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// createTestReport 创建测试报告数据
func createTestReport(t *testing.T, db *sql.DB, tenantID, deviceID, deviceCode string, date int) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()
	endTime := startTime + 86400 // 24 hours

	_, err := db.Exec(
		`INSERT INTO sleepace_report (
			tenant_id, device_id, device_code, record_count,
			start_time, end_time, date, stop_mode, time_step, timezone,
			sleep_state, report
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (tenant_id, device_id, date) DO UPDATE SET
		   record_count = EXCLUDED.record_count,
		   start_time = EXCLUDED.start_time,
		   end_time = EXCLUDED.end_time,
		   sleep_state = EXCLUDED.sleep_state,
		   report = EXCLUDED.report`,
		tenantID, deviceID, deviceCode, 1440, // record_count: 24 hours * 60 minutes
		startTime, endTime, date, 0, 60, 28800, // stop_mode: 0, time_step: 60s, timezone: UTC+8
		"[1,1,1,2,2,2,3,3,3,2,2,1,1,1]", // sleep_state
		`[{"summary":{"recordCount":1440,"startTime":` + strconv.FormatInt(startTime, 10) + `,"stopMode":0,"timeStep":60,"timezone":28800},"analysis":{"sleepStateStr":[1,1,1,2,2,2,3,3,3,2,2,1,1,1]}}]`, // report
	)
	require.NoError(t, err)
}

// TestGetSleepaceReports_Basic 测试基本的 GetSleepaceReports 功能
func TestGetSleepaceReports_Basic(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建测试报告
	date1 := 20240820
	date2 := 20240821
	createTestReport(t, db, tenantID, deviceID, "SP001", date1)
	createTestReport(t, db, tenantID, deviceID, "SP001", date2)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试获取报告列表
	req := GetSleepaceReportsRequest{
		TenantID:  tenantID,
		DeviceID:  deviceID,
		StartDate: date1,
		EndDate:   date2,
		Page:      1,
		PageSize:  10,
	}

	resp, err := service.GetSleepaceReports(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(resp.Items), 2)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 10, resp.Size)
	require.GreaterOrEqual(t, resp.Total, 2)

	// 验证报告数据
	foundDate1 := false
	foundDate2 := false
	for _, item := range resp.Items {
		if item.Date == date1 {
			foundDate1 = true
			require.Equal(t, deviceID, item.DeviceID)
			require.Equal(t, "SP001", item.DeviceCode)
		}
		if item.Date == date2 {
			foundDate2 = true
			require.Equal(t, deviceID, item.DeviceID)
			require.Equal(t, "SP001", item.DeviceCode)
		}
	}
	require.True(t, foundDate1, "Should find report for date1")
	require.True(t, foundDate2, "Should find report for date2")
}

// TestGetSleepaceReports_Pagination 测试分页功能
func TestGetSleepaceReports_Pagination(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建多个测试报告
	for i := 0; i < 5; i++ {
		date := 20240820 + i
		createTestReport(t, db, tenantID, deviceID, "SP001", date)
	}

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试第一页
	req1 := GetSleepaceReportsRequest{
		TenantID:  tenantID,
		DeviceID:  deviceID,
		StartDate: 20240820,
		EndDate:   20240824,
		Page:      1,
		PageSize:  2,
	}

	resp1, err := service.GetSleepaceReports(context.Background(), req1)
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Equal(t, 2, len(resp1.Items))
	require.Equal(t, 1, resp1.Page)
	require.Equal(t, 2, resp1.Size)
	require.Equal(t, 5, resp1.Total)

	// 测试第二页
	req2 := GetSleepaceReportsRequest{
		TenantID:  tenantID,
		DeviceID:  deviceID,
		StartDate: 20240820,
		EndDate:   20240824,
		Page:      2,
		PageSize:  2,
	}

	resp2, err := service.GetSleepaceReports(context.Background(), req2)
	require.NoError(t, err)
	require.NotNil(t, resp2)
	require.Equal(t, 2, len(resp2.Items))
	require.Equal(t, 2, resp2.Page)
	require.Equal(t, 2, resp2.Size)
	require.Equal(t, 5, resp2.Total)
}

// TestGetSleepaceReports_DefaultPagination 测试默认分页参数
func TestGetSleepaceReports_DefaultPagination(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试默认分页参数（page=0, size=0）
	req := GetSleepaceReportsRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Page:     0, // 应该默认为 1
		PageSize: 0, // 应该默认为 10
	}

	resp, err := service.GetSleepaceReports(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 10, resp.Size)
}

// TestGetSleepaceReports_InvalidDevice 测试无效设备
func TestGetSleepaceReports_InvalidDevice(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _ := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试无效设备 ID
	req := GetSleepaceReportsRequest{
		TenantID: tenantID,
		DeviceID: "00000000-0000-0000-0000-000000000999", // 不存在的设备
		Page:     1,
		PageSize: 10,
	}

	_, err := service.GetSleepaceReports(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "device not found")
}

// TestGetSleepaceReports_MissingParams 测试缺少参数
func TestGetSleepaceReports_MissingParams(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试缺少 tenant_id
	req1 := GetSleepaceReportsRequest{
		DeviceID: "00000000-0000-0000-0000-000000000001",
		Page:     1,
		PageSize: 10,
	}

	_, err := service.GetSleepaceReports(context.Background(), req1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")

	// 测试缺少 device_id
	req2 := GetSleepaceReportsRequest{
		TenantID: "00000000-0000-0000-0000-000000000999",
		Page:     1,
		PageSize: 10,
	}

	_, err = service.GetSleepaceReports(context.Background(), req2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")
}

// TestGetSleepaceReportDetail_Basic 测试基本的 GetSleepaceReportDetail 功能
func TestGetSleepaceReportDetail_Basic(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建测试报告
	date := 20240820
	createTestReport(t, db, tenantID, deviceID, "SP001", date)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试获取报告详情
	req := GetSleepaceReportDetailRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Date:     date,
	}

	resp, err := service.GetSleepaceReportDetail(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, deviceID, resp.DeviceID)
	require.Equal(t, "SP001", resp.DeviceCode)
	require.Equal(t, date, resp.Date)
	require.Equal(t, 1440, resp.RecordCount)
	require.NotEmpty(t, resp.Report)
	require.True(t, len(resp.Report) > 0)
	require.Equal(t, '[', resp.Report[0], "Report should start with '['")
}

// TestGetSleepaceReportDetail_NotFound 测试报告不存在
func TestGetSleepaceReportDetail_NotFound(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试不存在的报告
	req := GetSleepaceReportDetailRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Date:     99999999, // 不存在的日期
	}

	_, err := service.GetSleepaceReportDetail(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestGetSleepaceReportDetail_MissingParams 测试缺少参数
func TestGetSleepaceReportDetail_MissingParams(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试缺少 date
	req1 := GetSleepaceReportDetailRequest{
		TenantID: "00000000-0000-0000-0000-000000000999",
		DeviceID: "00000000-0000-0000-0000-000000000001",
		Date:     0,
	}

	_, err := service.GetSleepaceReportDetail(context.Background(), req1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")
}

// TestGetSleepaceReportDates_Basic 测试基本的 GetSleepaceReportDates 功能
func TestGetSleepaceReportDates_Basic(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建多个测试报告
	dates := []int{20240820, 20240821, 20240822}
	for _, date := range dates {
		createTestReport(t, db, tenantID, deviceID, "SP001", date)
	}

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试获取有效日期列表
	req := GetSleepaceReportDatesRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := service.GetSleepaceReportDates(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(resp.Dates), 3)

	// 验证所有日期都在结果中
	dateMap := make(map[int]bool)
	for _, date := range resp.Dates {
		dateMap[date] = true
	}
	for _, expectedDate := range dates {
		require.True(t, dateMap[expectedDate], "Date %d should be in the result", expectedDate)
	}
}

// TestGetSleepaceReportDates_Empty 测试没有报告的情况
func TestGetSleepaceReportDates_Empty(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试没有报告的情况
	req := GetSleepaceReportDatesRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := service.GetSleepaceReportDates(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Dates)
}

// TestGetSleepaceReportDates_MissingParams 测试缺少参数
func TestGetSleepaceReportDates_MissingParams(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试缺少 tenant_id
	req1 := GetSleepaceReportDatesRequest{
		DeviceID: "00000000-0000-0000-0000-000000000001",
	}

	_, err := service.GetSleepaceReportDates(context.Background(), req1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")

	// 测试缺少 device_id
	req2 := GetSleepaceReportDatesRequest{
		TenantID: "00000000-0000-0000-0000-000000000999",
	}

	_, err = service.GetSleepaceReportDates(context.Background(), req2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")
}

// TestValidateDevice_Basic 测试设备验证功能
func TestValidateDevice_Basic(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试有效设备
	req := GetSleepaceReportsRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Page:     1,
		PageSize: 10,
	}

	_, err := service.GetSleepaceReports(context.Background(), req)
	require.NoError(t, err, "Valid device should pass validation")
}

// TestValidateDevice_Disabled 测试禁用设备
func TestValidateDevice_Disabled(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 禁用设备
	_, err := db.Exec(
		`UPDATE devices SET status = 'disabled' WHERE device_id = $1`,
		deviceID,
	)
	require.NoError(t, err)

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试禁用设备
	req := GetSleepaceReportsRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Page:     1,
		PageSize: 10,
	}

	_, err = service.GetSleepaceReports(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "device not found")
}

// mockSleepaceClient 模拟 Sleepace 客户端
type mockSleepaceClient struct {
	reports []json.RawMessage
	err     error
}

func (m *mockSleepaceClient) Get24HourDailyWithMaxReport(deviceID, deviceCode string, startTime, endTime int64) ([]json.RawMessage, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reports, nil
}

// TestDownloadReport_Basic 测试基本的 DownloadReport 功能
func TestDownloadReport_Basic(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建模拟报告数据
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()
	endTime := startTime + 86400

	reportData := json.RawMessage(`{
		"summary": {
			"recordCount": 1440,
			"startTime": ` + strconv.FormatInt(startTime, 10) + `,
			"stopMode": 0,
			"timeStep": 60,
			"timezone": 28800
		},
		"analysis": {
			"sleepStateStr": [1,1,1,2,2,2,3,3,3,2,2,1,1,1]
		}
	}`)

	mockClient := &mockSleepaceClient{
		reports: []json.RawMessage{reportData},
		err:     nil,
	}

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	serviceImpl := NewSleepaceReportService(reportsRepo, db, logger).(*sleepaceReportService)
	serviceImpl.SetSleepaceClientForTest(mockClient)
	service := SleepaceReportService(serviceImpl)

	// 测试下载报告
	req := DownloadReportRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceCode: "SP001",
		StartTime:  startTime,
		EndTime:    endTime,
	}

	err := service.DownloadReport(context.Background(), req)
	require.NoError(t, err)

	// 验证报告已保存
	// timeToDate 将 Unix 时间戳转换为 YYYYMMDD 格式的整数
	tm := time.Unix(startTime, 0).UTC()
	date := tm.Year()*10000 + int(tm.Month())*100 + tm.Day()
	report, err := reportsRepo.GetReport(context.Background(), tenantID, deviceID, date)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, deviceID, report.DeviceID)
	require.Equal(t, "SP001", report.DeviceCode)
	require.Equal(t, 1440, report.RecordCount)
}

// TestDownloadReport_MissingParams 测试缺少参数
func TestDownloadReport_MissingParams(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试缺少 tenant_id
	req1 := DownloadReportRequest{
		DeviceID:   "00000000-0000-0000-0000-000000000001",
		DeviceCode: "SP001",
		StartTime:  1234567890,
		EndTime:    1234567890,
	}

	err := service.DownloadReport(context.Background(), req1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")

	// 测试缺少 device_code
	req2 := DownloadReportRequest{
		TenantID:  "00000000-0000-0000-0000-000000000999",
		DeviceID:  "00000000-0000-0000-0000-000000000001",
		StartTime: 1234567890,
		EndTime:   1234567890,
	}

	err = service.DownloadReport(context.Background(), req2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required")
}

// TestDownloadReport_ClientNotInitialized 测试客户端未初始化
func TestDownloadReport_ClientNotInitialized(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建 Service（不设置客户端）
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	service := NewSleepaceReportService(reportsRepo, db, logger)

	// 测试下载报告（客户端未初始化）
	req := DownloadReportRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceCode: "SP001",
		StartTime:  1234567890,
		EndTime:    1234567890,
	}

	err := service.DownloadReport(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "client not initialized")
}

// TestDownloadReport_APIFailure 测试 API 调用失败
func TestDownloadReport_APIFailure(t *testing.T) {
	db := setupTestDBForSleepace(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, deviceID := createTestTenantAndDeviceForSleepace(t, db)
	defer cleanupTestDataForSleepace(t, db, tenantID)

	// 创建模拟客户端（返回错误）
	mockClient := &mockSleepaceClient{
		reports: nil,
		err:     fmt.Errorf("API call failed"),
	}

	// 创建 Service
	reportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
	logger := getTestLoggerForSleepace()
	serviceImpl := NewSleepaceReportService(reportsRepo, db, logger).(*sleepaceReportService)
	serviceImpl.SetSleepaceClientForTest(mockClient)
	service := SleepaceReportService(serviceImpl)

	// 测试下载报告（API 失败）
	req := DownloadReportRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceCode: "SP001",
		StartTime:  1234567890,
		EndTime:    1234567890,
	}

	err := service.DownloadReport(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get reports from Sleepace API")
}

