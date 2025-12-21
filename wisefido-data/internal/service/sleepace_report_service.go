package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// sleepaceClientInterface Sleepace 客户端接口（用于测试和扩展）
type sleepaceClientInterface interface {
	Get24HourDailyWithMaxReport(deviceID, deviceCode string, startTime, endTime int64) ([]json.RawMessage, error)
}

// SleepaceReportService Sleepace 睡眠报告服务接口
type SleepaceReportService interface {
	// GetSleepaceReports 获取睡眠报告列表
	GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error)

	// GetSleepaceReportDetail 获取睡眠报告详情
	GetSleepaceReportDetail(ctx context.Context, req GetSleepaceReportDetailRequest) (*GetSleepaceReportDetailResponse, error)

	// GetSleepaceReportDates 获取有数据的日期列表
	GetSleepaceReportDates(ctx context.Context, req GetSleepaceReportDatesRequest) (*GetSleepaceReportDatesResponse, error)

	// DownloadReport 从厂家服务下载报告并保存到数据库
	DownloadReport(ctx context.Context, req DownloadReportRequest) error
}

// sleepaceReportService 实现
type sleepaceReportService struct {
	reportsRepo    repository.SleepaceReportsRepository
	db             *sql.DB // 用于设备验证等复杂查询
	sleepaceClient sleepaceClientInterface // Sleepace 厂家 API 客户端（使用接口，支持测试）
	logger         *zap.Logger
}

// NewSleepaceReportService 创建 SleepaceReportService 实例
func NewSleepaceReportService(reportsRepo repository.SleepaceReportsRepository, db *sql.DB, logger *zap.Logger) SleepaceReportService {
	return &sleepaceReportService{
		reportsRepo: reportsRepo,
		db:          db,
		logger:      logger,
		// sleepaceClient 需要通过 SetSleepaceClient 设置（延迟初始化）
	}
}

// SetSleepaceClient 设置 Sleepace 客户端（延迟初始化，避免循环依赖）
func (s *sleepaceReportService) SetSleepaceClient(client *SleepaceClient) {
	s.sleepaceClient = client
}

// SetSleepaceClientForTest 设置 Sleepace 客户端接口（用于测试）
func (s *sleepaceReportService) SetSleepaceClientForTest(client sleepaceClientInterface) {
	s.sleepaceClient = client
}


// ============================================
// Request/Response DTOs
// ============================================

// GetSleepaceReportsRequest 获取报告列表请求
type GetSleepaceReportsRequest struct {
	TenantID  string // 必填
	DeviceID  string // 必填（设备 ID）
	StartDate int    // 开始日期（YYYYMMDD 格式，如 20240820）
	EndDate   int    // 结束日期（YYYYMMDD 格式，如 20240820）
	Page      int    // 页码，默认 1
	PageSize  int    // 每页数量，默认 10
}

// GetSleepaceReportsResponse 获取报告列表响应
type GetSleepaceReportsResponse struct {
	Items []*SleepaceReportOutlineDTO `json:"items"`
	Total int                         `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

// SleepaceReportOutlineDTO 报告概要 DTO（列表项，不包含完整 report 字段）
type SleepaceReportOutlineDTO struct {
	ID          string `json:"id"`          // report_id
	DeviceID    string `json:"deviceId"`   // device_id
	DeviceCode  string `json:"deviceCode"` // device_code
	RecordCount int    `json:"recordCount"`
	StartTime   int64  `json:"startTime"`  // Unix 时间戳（秒）
	EndTime     int64  `json:"endTime"`    // Unix 时间戳（秒）
	Date        int    `json:"date"`       // YYYYMMDD 格式
	StopMode    int    `json:"stopMode"`
	TimeStep    int    `json:"timeStep"`
	Timezone    int    `json:"timezone"`
	SleepState  string `json:"sleepState"` // JSON 字符串数组
}

// GetSleepaceReportDetailRequest 获取报告详情请求
type GetSleepaceReportDetailRequest struct {
	TenantID string // 必填
	DeviceID string // 必填
	Date     int    // 日期（YYYYMMDD 格式，如 20240820）
}

// GetSleepaceReportDetailResponse 获取报告详情响应
type GetSleepaceReportDetailResponse struct {
	ID          string `json:"id"`          // report_id
	DeviceID    string `json:"deviceId"`   // device_id
	DeviceCode  string `json:"deviceCode"` // device_code
	RecordCount int    `json:"recordCount"`
	StartTime   int64  `json:"startTime"`  // Unix 时间戳（秒）
	EndTime     int64  `json:"endTime"`    // Unix 时间戳（秒）
	Date        int    `json:"date"`       // YYYYMMDD 格式
	StopMode    int    `json:"stopMode"`
	TimeStep    int    `json:"timeStep"`
	Timezone    int    `json:"timezone"`
	Report      string `json:"report"` // 完整报告数据（JSON 字符串）
}

// GetSleepaceReportDatesRequest 获取有效日期列表请求
type GetSleepaceReportDatesRequest struct {
	TenantID string // 必填
	DeviceID string // 必填
}

// GetSleepaceReportDatesResponse 获取有效日期列表响应
type GetSleepaceReportDatesResponse struct {
	Dates []int `json:"dates"` // 日期列表（YYYYMMDD 格式）
}

// DownloadReportRequest 下载报告请求
type DownloadReportRequest struct {
	TenantID   string // 必填
	DeviceID   string // 必填（设备 ID）
	DeviceCode string // 必填（设备编码，对应 devices.serial_number 或 devices.uid）
	StartTime  int64  // 开始时间（Unix 时间戳，秒）
	EndTime    int64  // 结束时间（Unix 时间戳，秒）
}

// ============================================
// Service 方法实现
// ============================================

// GetSleepaceReports 获取睡眠报告列表
func (s *sleepaceReportService) GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error) {
	if req.TenantID == "" || req.DeviceID == "" {
		return nil, fmt.Errorf("tenant_id and device_id are required")
	}

	// 验证设备是否存在且属于该租户
	if err := s.validateDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		return nil, err
	}

	// 默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.PageSize
	if size <= 0 {
		size = 10
	}

	// 默认日期范围（如果未指定，使用最近 30 天）
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate == 0 || endDate == 0 {
		now := time.Now()
		endDate = dateToInt(now)
		startDate = dateToInt(now.AddDate(0, 0, -30))
	}

	// 查询报告列表
	reports, total, err := s.reportsRepo.ListReports(ctx, req.TenantID, req.DeviceID, startDate, endDate, page, size)
	if err != nil {
		s.logger.Error("failed to list sleepace reports",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list sleepace reports: %w", err)
	}

	// 转换为 DTO
	items := make([]*SleepaceReportOutlineDTO, 0, len(reports))
	for _, report := range reports {
		items = append(items, &SleepaceReportOutlineDTO{
			ID:          report.ReportID,
			DeviceID:    report.DeviceID,
			DeviceCode:  report.DeviceCode,
			RecordCount: report.RecordCount,
			StartTime:   report.StartTime,
			EndTime:     report.EndTime,
			Date:        report.Date,
			StopMode:    report.StopMode,
			TimeStep:    report.TimeStep,
			Timezone:    report.Timezone,
			SleepState:  report.SleepState,
		})
	}

	return &GetSleepaceReportsResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// GetSleepaceReportDetail 获取睡眠报告详情
func (s *sleepaceReportService) GetSleepaceReportDetail(ctx context.Context, req GetSleepaceReportDetailRequest) (*GetSleepaceReportDetailResponse, error) {
	if req.TenantID == "" || req.DeviceID == "" || req.Date == 0 {
		return nil, fmt.Errorf("tenant_id, device_id and date are required")
	}

	// 验证设备是否存在且属于该租户
	if err := s.validateDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		return nil, err
	}

	// 查询报告详情
	report, err := s.reportsRepo.GetReport(ctx, req.TenantID, req.DeviceID, req.Date)
	if err != nil {
		s.logger.Error("failed to get sleepace report detail",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Int("date", req.Date),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get sleepace report detail: %w", err)
	}

	if report == nil {
		return nil, fmt.Errorf("report not found for device %s on date %d", req.DeviceID, req.Date)
	}

	// 确保 report 字段以 '[' 开头（v1.0 兼容性）
	reportData := report.Report
	if len(reportData) > 0 && reportData[0] != '[' {
		reportData = "[" + reportData + "]"
	}

	return &GetSleepaceReportDetailResponse{
		ID:          report.ReportID,
		DeviceID:    report.DeviceID,
		DeviceCode:  report.DeviceCode,
		RecordCount: report.RecordCount,
		StartTime:   report.StartTime,
		EndTime:     report.EndTime,
		Date:        report.Date,
		StopMode:    report.StopMode,
		TimeStep:    report.TimeStep,
		Timezone:    report.Timezone,
		Report:      reportData,
	}, nil
}

// GetSleepaceReportDates 获取有数据的日期列表
func (s *sleepaceReportService) GetSleepaceReportDates(ctx context.Context, req GetSleepaceReportDatesRequest) (*GetSleepaceReportDatesResponse, error) {
	if req.TenantID == "" || req.DeviceID == "" {
		return nil, fmt.Errorf("tenant_id and device_id are required")
	}

	// 验证设备是否存在且属于该租户
	if err := s.validateDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		return nil, err
	}

	// 查询有效日期列表
	dates, err := s.reportsRepo.GetValidDates(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		s.logger.Error("failed to get sleepace report dates",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get sleepace report dates: %w", err)
	}

	return &GetSleepaceReportDatesResponse{
		Dates: dates,
	}, nil
}

// DownloadReport 从厂家服务下载报告并保存到数据库
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport
func (s *sleepaceReportService) DownloadReport(ctx context.Context, req DownloadReportRequest) error {
	if req.TenantID == "" || req.DeviceID == "" || req.DeviceCode == "" {
		return fmt.Errorf("tenant_id, device_id and device_code are required")
	}

	if req.StartTime == 0 || req.EndTime == 0 {
		return fmt.Errorf("start_time and end_time are required")
	}

	// 获取客户端（支持接口，用于测试）
	var client sleepaceClientInterface
	if s.sleepaceClient != nil {
		client = s.sleepaceClient
	} else {
		return fmt.Errorf("sleepace client not initialized")
	}

	// 验证设备是否存在且属于该租户
	if err := s.validateDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		return err
	}

	// 调用 Sleepace 厂家 API 获取报告
	reports, err := client.Get24HourDailyWithMaxReport(req.DeviceID, req.DeviceCode, req.StartTime, req.EndTime)
	if err != nil {
		s.logger.Error("Failed to get reports from Sleepace API",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.String("device_code", req.DeviceCode),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get reports from Sleepace API: %w", err)
	}

	// 解析并保存每个报告
	for i := len(reports) - 1; i >= 0; i-- {
		reportData := reports[i]

		// 解析报告数据
		var report struct {
			Summary struct {
				RecordCount int   `json:"recordCount"`
				StartTime   int64 `json:"startTime"`
				StopMode    int   `json:"stopMode"`
				TimeStep    int   `json:"timeStep"`
				Timezone    int   `json:"timezone"`
			} `json:"summary"`
			Analysis struct {
				SleepStateStr json.RawMessage `json:"sleepStateStr"`
			} `json:"analysis"`
		}

		if err := json.Unmarshal(reportData, &report); err != nil {
			s.logger.Error("Failed to unmarshal report data",
				zap.Error(err),
				zap.Int("index", i),
			)
			continue // 跳过无效的报告
		}

		// 转换为领域模型
		domainReport := &domain.SleepaceReport{
			DeviceID:    req.DeviceID,
			DeviceCode:  req.DeviceCode,
			RecordCount: report.Summary.RecordCount,
			StartTime:   report.Summary.StartTime,
			EndTime:     report.Summary.StartTime + int64(report.Summary.TimeStep)*int64(report.Summary.RecordCount),
			Date:        timeToDate(report.Summary.StartTime),
			StopMode:    report.Summary.StopMode,
			TimeStep:    report.Summary.TimeStep,
			Timezone:    report.Summary.Timezone,
			SleepState:  string(report.Analysis.SleepStateStr),
			Report:      "[" + string(reportData) + "]", // 确保 report 字段是 JSON 数组格式
		}

		// 保存到数据库
		if err := s.reportsRepo.SaveReport(ctx, req.TenantID, domainReport); err != nil {
			s.logger.Error("Failed to save report",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.Int("date", domainReport.Date),
				zap.Error(err),
			)
			// 继续处理其他报告，不中断整个流程
			continue
		}

		s.logger.Info("Successfully saved report",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Int("date", domainReport.Date),
		)
	}

	return nil
}

// ============================================
// 辅助方法
// ============================================

// validateDevice 验证设备是否存在且属于该租户
func (s *sleepaceReportService) validateDevice(ctx context.Context, tenantID, deviceID string) error {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM devices
			WHERE device_id = $1::uuid
			  AND tenant_id = $2::uuid
			  AND status <> 'disabled'
		)
	`
	var exists bool
	if err := s.db.QueryRowContext(ctx, query, deviceID, tenantID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to validate device: %w", err)
	}
	if !exists {
		return fmt.Errorf("device not found or not accessible")
	}
	return nil
}

// dateToInt 将 time.Time 转换为 YYYYMMDD 格式的整数
func dateToInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// timeToDate 将 Unix 时间戳转换为 YYYYMMDD 格式的整数
func timeToDate(timestamp int64) int {
	tm := time.Unix(timestamp, 0).UTC()
	return tm.Year()*10000 + int(tm.Month())*100 + tm.Day()
}

