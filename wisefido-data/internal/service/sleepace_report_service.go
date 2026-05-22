package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// 厂家 Sleepace 与 wisefido 库表对应（HTTP/MQTT 报文）：
//   sleepace.deviceId    ↔ device_store.device_code（空则 COALESCE 用 devices.device_uid）
//   sleepace.userId      ↔ devices.device_id（UUID）
//   sleepace.deviceName  ↔ devices.device_uid

// sleepaceReportGateway 经 wisefido-sleepace 透明代理调厂家并本进程落库；测试可注入 mock。
type sleepaceReportGateway interface {
	// deviceID=devices.device_id；deviceCode=device_store.device_code（厂家 MQTT/API deviceId）；deviceUID=devices.device_uid。
	DownloadReportPersist(ctx context.Context, tenantID, deviceID, deviceCode, deviceUID string, startTime, endTime int64) error
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

	// DownloadReportByDeviceCode 按厂家 data.deviceId（= device_store.device_code，可回落匹配 device_uid）拉取并入库
	DownloadReportByDeviceCode(ctx context.Context, sleepaceDeviceID string, startTime, endTime int64) error
	// DownloadReportByWisefidoDeviceID 按厂家 data.userId（= devices.device_id UUID）拉取并入库
	DownloadReportByWisefidoDeviceID(ctx context.Context, deviceID string, startTime, endTime int64) error

	// BackfillFromVendor 区间内按日对比 DB，缺失的连续段各调一次厂家拉取落库。
	// 由 SleepaceReportBackfillScheduler 周期调用；不应在用户即时路径同步等待
	// （新 device / 长时间未访问的 device 30 天 missing 串行多段 vendor download 会阻塞 5-30s）。
	BackfillFromVendor(ctx context.Context, tenantID, deviceID string, startDate, endDate int) error
}

// sleepaceReportService 实现
type sleepaceReportService struct {
	reportsRepo   repository.SleepaceReportsRepository
	db            *sql.DB               // 用于设备验证等复杂查询
	reportGateway sleepaceReportGateway // *SleepaceGatewayClient
	logger        *zap.Logger
}

// NewSleepaceReportService 创建 SleepaceReportService 实例
func NewSleepaceReportService(reportsRepo repository.SleepaceReportsRepository, db *sql.DB, logger *zap.Logger) SleepaceReportService {
	return &sleepaceReportService{
		reportsRepo: reportsRepo,
		db:          db,
		logger:      logger,
		// reportGateway 通过 SetSleepaceGatewayClient 设置
	}
}

// SetSleepaceGatewayClient 设置经 wisefido-sleepace 的网关客户端（延迟初始化，避免循环依赖）
func (s *sleepaceReportService) SetSleepaceGatewayClient(gw *SleepaceGatewayClient) {
	s.reportGateway = gw
}

// SetSleepaceReportGatewayForTest 测试注入 mock
func (s *sleepaceReportService) SetSleepaceReportGatewayForTest(g sleepaceReportGateway) {
	s.reportGateway = g
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
	ID          string `json:"id"`         // report_id
	DeviceAddr  string `json:"deviceAddr"` // INET /128 canonical text, 业务侧寻址
	DeviceCode  string `json:"deviceCode"` // device_store.device_code（厂家 deviceId）
	DeviceUID   string `json:"deviceUid"`  // devices.device_uid logMAC
	RecordCount int    `json:"recordCount"`
	StartTime   int64  `json:"startTime"` // Unix 时间戳（秒）
	EndTime     int64  `json:"endTime"`   // Unix 时间戳（秒）
	Date        int    `json:"date"`      // YYYYMMDD 格式
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
	ID                      string                       `json:"id"`         // report_id
	DeviceAddr              string                       `json:"deviceAddr"` // INET /128 canonical text, 业务侧寻址
	DeviceCode              string                       `json:"deviceCode"` // device_store.device_code（厂家 deviceId）
	DeviceUID               string                       `json:"deviceUid"`  // devices.device_uid logMAC
	RecordCount             int                          `json:"recordCount"`
	StartTime               int64                        `json:"startTime"` // Unix 时间戳（秒）
	EndTime                 int64                        `json:"endTime"`   // Unix 时间戳（秒）
	Date                    int                          `json:"date"`      // YYYYMMDD 格式
	StopMode                int                          `json:"stopMode"`
	TimeStep                int                          `json:"timeStep"`
	Timezone                int                          `json:"timezone"`
	Report                string                   `json:"report"` // 完整报告数据（JSON 字符串）
	WeeklySleepEfficiency WeeklySleepEfficiencyDTO `json:"weeklySleepEfficiency"`
}

// WeeklySleepEfficiencyDTO 本周与上周各 7 格（周一至周日）睡眠效率 seIndex（%）；无报告或解析失败为 null，由前端自行比对
type WeeklySleepEfficiencyDTO struct {
	WeekMonday int        `json:"weekMonday"` // 当日所在自然周的周一 YYYYMMDD（本地日历）
	ThisWeek   []*float64 `json:"thisWeek"`   // 索引 0=周一 … 6=周日
	LastWeek   []*float64 `json:"lastWeek"`   // 上一完整周，同上
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

// DownloadReportRequest 下载报告请求（字段与厂家 get24HourDailyWithMaxReport 对齐，见文件头注释）
type DownloadReportRequest struct {
	TenantID         string // 必填
	DeviceID         string // 必填 → 厂家 userId（devices.device_id UUID）
	SleepaceDeviceID string // 必填 → 厂家 deviceId（device_store.device_code，空已 COALESCE device_uid）
	DeviceUID        string // 必填 → 厂家 deviceName（devices.device_uid）
	StartTime        int64  // 开始时间（Unix 时间戳，秒）
	EndTime          int64  // 结束时间（Unix 时间戳，秒）
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

	// (2026-05-15) vendor fill 已从这里 **移到独立 SleepaceReportBackfillScheduler**：
	// 即时路径 (用户进入页面) 只读 cached DB，不再阻塞等厂家 N 段串行 download (新 device 30 天 missing 实测 5-30s)。
	// 后台 scheduler 周期跑 fillMissingSleepaceReportsFromVendor (即下面 line 631 那个函数)，
	// 对所有 bound sleepad device 自动 backfill。新 device 第一次进入可能看到不完整列表，
	// 下次进入或手动刷新 (next scheduler tick 之后) 即可补齐。
	// 见 internal/service/sleepace_report_backfill_scheduler.go + memory: sleepace-interval-scheduler 同模式。

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
			DeviceAddr:  report.DeviceAddr,
			DeviceCode:  report.DeviceCode,
			DeviceUID:   report.DeviceUID,
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

	s.logger.Info("GetSleepaceReports ok",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.Int("start_date", startDate),
		zap.Int("end_date", endDate),
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int("returned", len(items)),
		zap.Int("total", total),
	)

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

	reportData := normalizeSleepaceReportArrayJSON(report.Report)

	thisWeekMon := mondayOfWeekContainingYYYYMMDD(report.Date)
	lastWeekMon := addDaysYYYYMMDD(thisWeekMon, -7)
	thisWeekSun := addDaysYYYYMMDD(thisWeekMon, 6)

	// (2026-05-15) detail 页"周效率"原同步等 vendor 拉 lastWeekMon~thisWeekSun 缺失天，已挪到
	// SleepaceReportBackfillScheduler 周期补；正常情况 backfill scheduler 30 天回溯已覆盖周窗口。
	// 极端情况（前一日刚生成 + 当前小时 scheduler 还没跑）weekly efficiency 可能少 1 天数据 —
	// trade-off 接受，详详情页打开速度（删除前 detail 也可能 1-3s vendor 等待）。

	rangeReports, err := s.reportsRepo.ListReportsAllInRange(ctx, req.TenantID, req.DeviceID, lastWeekMon, thisWeekSun)
	if err != nil {
		s.logger.Error("ListReportsAllInRange for weekly efficiency failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to load reports for weekly efficiency: %w", err)
	}
	weekly := buildWeeklySleepEfficiencyFromReports(rangeReports, thisWeekMon, lastWeekMon)

	s.logger.Info("GetSleepaceReportDetail ok",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.Int("date", req.Date),
		zap.String("report_id", report.ReportID),
		zap.Int("report_json_len", len(reportData)),
		zap.Int("record_count", report.RecordCount),
	)

	return &GetSleepaceReportDetailResponse{
		ID:                    report.ReportID,
		DeviceAddr:            report.DeviceAddr,
		DeviceCode:            report.DeviceCode,
		DeviceUID:             report.DeviceUID,
		RecordCount:           report.RecordCount,
		StartTime:             report.StartTime,
		EndTime:               report.EndTime,
		Date:                  report.Date,
		StopMode:              report.StopMode,
		TimeStep:              report.TimeStep,
		Timezone:              report.Timezone,
		Report:                reportData,
		WeeklySleepEfficiency: weekly,
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

func normalizeSleepaceReportArrayJSON(s string) string {
	if len(s) > 0 && s[0] != '[' {
		return "[" + s + "]"
	}
	return s
}

// mondayOfWeekContainingYYYYMMDD 返回包含 d 的那一周周一（本地时区，12:00 避免 DST 边界）
func mondayOfWeekContainingYYYYMMDD(d int) int {
	y := d / 10000
	m := (d % 10000) / 100
	day := d % 100
	t := time.Date(y, time.Month(m), day, 12, 0, 0, 0, time.Local)
	wd := t.Weekday() // Sunday=0, Monday=1, ...
	offset := int((wd + 6) % 7)
	return dateToInt(t.AddDate(0, 0, -offset))
}

func parseSleepEfficiencyFromSleepaceReportJSON(reportData string) (float64, bool) {
	var arr []struct {
		Analysis struct {
			MaxReport struct {
				SeIndex float64 `json:"seIndex"`
			} `json:"maxReport"`
		} `json:"analysis"`
	}
	if err := json.Unmarshal([]byte(reportData), &arr); err != nil || len(arr) == 0 {
		return 0, false
	}
	return arr[0].Analysis.MaxReport.SeIndex, true
}

func addDaysYYYYMMDD(d int, deltaDays int) int {
	y := d / 10000
	m := (d % 10000) / 100
	day := d % 100
	t := time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.Local)
	return dateToInt(t.AddDate(0, 0, deltaDays))
}

func buildWeeklySleepEfficiencyFromReports(reports []*domain.SleepaceReport, thisWeekMon, lastWeekMon int) WeeklySleepEfficiencyDTO {
	byDate := make(map[int]string, len(reports))
	for _, r := range reports {
		if r == nil {
			continue
		}
		byDate[r.Date] = normalizeSleepaceReportArrayJSON(r.Report)
	}
	thisW := make([]*float64, 7)
	lastW := make([]*float64, 7)
	for i := 0; i < 7; i++ {
		dThis := addDaysYYYYMMDD(thisWeekMon, i)
		if raw, ok := byDate[dThis]; ok {
			if v, ok2 := parseSleepEfficiencyFromSleepaceReportJSON(raw); ok2 {
				thisW[i] = &v
			}
		}
		dLast := addDaysYYYYMMDD(lastWeekMon, i)
		if raw, ok := byDate[dLast]; ok {
			if v, ok2 := parseSleepEfficiencyFromSleepaceReportJSON(raw); ok2 {
				lastW[i] = &v
			}
		}
	}
	return WeeklySleepEfficiencyDTO{
		WeekMonday: thisWeekMon,
		ThisWeek:   thisW,
		LastWeek:   lastW,
	}
}

// DownloadReport 校验后经透明代理调厂家 get24HourDailyWithMaxReport，由本服务写入 sleepace_report。
func (s *sleepaceReportService) DownloadReport(ctx context.Context, req DownloadReportRequest) error {
	if req.TenantID == "" || req.DeviceID == "" || req.SleepaceDeviceID == "" || req.DeviceUID == "" {
		return fmt.Errorf("tenant_id, device_id, sleepace_device_id and device_uid are required")
	}
	if req.StartTime == 0 || req.EndTime == 0 {
		return fmt.Errorf("start_time and end_time are required")
	}
	if s.reportGateway == nil {
		return fmt.Errorf("sleepace gateway client not initialized")
	}
	if err := s.validateDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		return err
	}
	// req.DeviceID = devices.device_id（UUID），作厂家 data.userId；SleepaceDeviceID 仅校验用。
	if err := s.reportGateway.DownloadReportPersist(ctx, req.TenantID, req.DeviceID, req.SleepaceDeviceID, req.DeviceUID, req.StartTime, req.EndTime); err != nil {
		s.logger.Error("report download via gateway failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.String("device_uid", req.DeviceUID),
			zap.String("sleepace_device_id", req.SleepaceDeviceID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// DownloadReportByDeviceCode 按厂家 deviceId（优先 device_store.device_code，其次 uid）解析后拉取
func (s *sleepaceReportService) DownloadReportByDeviceCode(ctx context.Context, sleepaceDeviceID string, startTime, endTime int64) error {
	if sleepaceDeviceID == "" {
		return fmt.Errorf("sleepace deviceId is required")
	}
	if startTime == 0 || endTime == 0 {
		return fmt.Errorf("start_time and end_time are required")
	}
	tenantID, deviceID, sid, uid, err := s.lookupSleepaceReportIdentityBySleepaceDeviceID(ctx, sleepaceDeviceID)
	if err != nil {
		return err
	}
	return s.DownloadReport(ctx, DownloadReportRequest{
		TenantID:         tenantID,
		DeviceID:         deviceID,
		SleepaceDeviceID: sid,
		DeviceUID:        uid,
		StartTime:        startTime,
		EndTime:          endTime,
	})
}

// DownloadReportByWisefidoDeviceID 按 devices.device_id（厂家 userId）拉取
func (s *sleepaceReportService) DownloadReportByWisefidoDeviceID(ctx context.Context, deviceID string, startTime, endTime int64) error {
	if deviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	if startTime == 0 || endTime == 0 {
		return fmt.Errorf("start_time and end_time are required")
	}
	tenantID, devAddr, sid, uid, err := s.lookupSleepaceReportIdentityByDeviceAddr(ctx, deviceID)
	if err != nil {
		return err
	}
	return s.DownloadReport(ctx, DownloadReportRequest{
		TenantID:         tenantID,
		DeviceID:         devAddr,
		SleepaceDeviceID: sid,
		DeviceUID:        uid,
		StartTime:        startTime,
		EndTime:          endTime,
	})
}

// lookupSleepaceReportIdentityByDeviceAddr — Phase 2: 入参 deviceAddr (INET /128 canonical text)；
// tenant_id 派生自 device_addr 前 /48；device_uid / device_code 全在 device_factory_meta。
func (s *sleepaceReportService) lookupSleepaceReportIdentityByDeviceAddr(ctx context.Context, deviceAddr string) (tenantID, devAddrOut, sleepaceDeviceID, deviceUID string, err error) {
	q := `
		SELECT host(network(set_masklen(d.device_addr, 48))) || '/48',
			host(d.device_addr),
			COALESCE(NULLIF(TRIM(dfm.device_code), ''), dfm.device_uid),
			dfm.device_uid
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE d.device_addr = $1::INET
		LIMIT 1
	`
	err = s.db.QueryRowContext(ctx, q, deviceAddr).Scan(&tenantID, &devAddrOut, &sleepaceDeviceID, &deviceUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", "", fmt.Errorf("device not found: device_addr=%s", deviceAddr)
		}
		return "", "", "", "", fmt.Errorf("lookup device: %w", err)
	}
	if sleepaceDeviceID == "" || deviceUID == "" {
		return "", "", "", "", fmt.Errorf("device identifiers incomplete for device_addr=%s", deviceAddr)
	}
	return tenantID, devAddrOut, sleepaceDeviceID, deviceUID, nil
}

// lookupSleepaceReportIdentityBySleepaceDeviceID — Sleepace 厂家 deviceId（= device_code 或 device_uid）反查 wisefido device。
// v2: 全部从 device_factory_meta 取标识；tenant 派生 /48；devices 表无 status 列。
// Phase 2: device_id UUID 列已退役；返回 devID = host(device_addr) 文本。
func (s *sleepaceReportService) lookupSleepaceReportIdentityBySleepaceDeviceID(ctx context.Context, sleepaceDeviceID string) (tenantID, devID, sid, uid string, err error) {
	q := `
		SELECT host(network(set_masklen(d.device_addr, 48))) || '/48',
			host(d.device_addr),
			COALESCE(NULLIF(TRIM(dfm.device_code), ''), dfm.device_uid),
			dfm.device_uid
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE (NULLIF(TRIM(dfm.device_code), '') IS NOT NULL AND TRIM(dfm.device_code) = $1)
		   OR dfm.device_uid = $1
		LIMIT 1
	`
	err = s.db.QueryRowContext(ctx, q, sleepaceDeviceID).Scan(&tenantID, &devID, &sid, &uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", "", fmt.Errorf("device not found for sleepace deviceId=%s", sleepaceDeviceID)
		}
		return "", "", "", "", fmt.Errorf("lookup device: %w", err)
	}
	if sid == "" || uid == "" {
		return "", "", "", "", fmt.Errorf("device identifiers incomplete for sleepace deviceId=%s", sleepaceDeviceID)
	}
	return tenantID, devID, sid, uid, nil
}

// ============================================
// 辅助方法
// ============================================

func addOneDayYYYYMMDD(d int) int {
	y := d / 10000
	m := (d % 10000) / 100
	day := d % 100
	t := time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.Local)
	return dateToInt(t.AddDate(0, 0, 1))
}

func enumerateYYYYMMDDRange(startDate, endDate int) []int {
	if startDate == 0 || endDate == 0 || startDate > endDate {
		return nil
	}
	out := make([]int, 0)
	for d := startDate; d <= endDate; d = addOneDayYYYYMMDD(d) {
		out = append(out, d)
	}
	return out
}

func groupConsecutiveYYYYMMDD(dates []int) [][]int {
	if len(dates) == 0 {
		return nil
	}
	sort.Ints(dates)
	uniq := make([]int, 0, len(dates))
	for _, d := range dates {
		if len(uniq) == 0 || d != uniq[len(uniq)-1] {
			uniq = append(uniq, d)
		}
	}
	runs := make([][]int, 0)
	run := []int{uniq[0]}
	for i := 1; i < len(uniq); i++ {
		if addOneDayYYYYMMDD(uniq[i-1]) == uniq[i] {
			run = append(run, uniq[i])
		} else {
			runs = append(runs, run)
			run = []int{uniq[i]}
		}
	}
	return append(runs, run)
}

func yyyymmddLocalStartUnix(d int) int64 {
	y := d / 10000
	m := (d % 10000) / 100
	day := d % 100
	t := time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.Local)
	return t.Unix()
}

func yyyymmddLocalEndUnix(d int) int64 {
	y := d / 10000
	m := (d % 10000) / 100
	day := d % 100
	t := time.Date(y, time.Month(m), day, 23, 59, 59, 0, time.Local)
	return t.Unix()
}

// BackfillFromVendor 区间内按日对比库表，缺失的连续段各调一次厂家拉取落库。
// 公开给 SleepaceReportBackfillScheduler 调；旧名 fillMissingSleepaceReportsFromVendor 已 rename。
func (s *sleepaceReportService) BackfillFromVendor(ctx context.Context, tenantID, deviceID string, startDate, endDate int) error {
	existing, err := s.reportsRepo.GetValidDatesInRange(ctx, tenantID, deviceID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("get valid dates in range: %w", err)
	}
	have := make(map[int]struct{}, len(existing))
	for _, d := range existing {
		have[d] = struct{}{}
	}
	expected := enumerateYYYYMMDDRange(startDate, endDate)
	missing := make([]int, 0)
	for _, d := range expected {
		if _, ok := have[d]; !ok {
			missing = append(missing, d)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	_, _, sid, uid, err := s.lookupSleepaceReportIdentityByDeviceAddr(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("lookup sleepace identity: %w", err)
	}
	for _, run := range groupConsecutiveYYYYMMDD(missing) {
		if len(run) == 0 {
			continue
		}
		st := yyyymmddLocalStartUnix(run[0])
		et := yyyymmddLocalEndUnix(run[len(run)-1])
		if err := s.reportGateway.DownloadReportPersist(ctx, tenantID, deviceID, sid, uid, st, et); err != nil {
			s.logger.Warn("sleepace vendor download for date gap failed",
				zap.String("device_id", deviceID),
				zap.Ints("date_run", run),
				zap.Error(err),
			)
		}
	}
	return nil
}

// validateDevice 验证设备是否存在且属于该租户
// v2: devices 表已删 tenant_id 列（tenant 派生自 device_addr 前 /48）；
//     也删 status 列（row 存在 = enabled）；
//     tenantID 入参现为 /48 INET CIDR 字符串。
// Phase 2: deviceID 入参 = device_addr (INET /128 canonical text)。
func (s *sleepaceReportService) validateDevice(ctx context.Context, tenantID, deviceID string) error {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM devices
			WHERE device_addr = $1::INET
			  AND device_addr <<= $2::INET
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
