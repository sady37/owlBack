package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// SleepaceReportsRepository Sleepace 报告 Repository 接口
// 使用强类型领域模型，不使用 map[string]any
// 设计原则：从底层（数据库）向上设计，Repository 层只负责数据访问
type SleepaceReportsRepository interface {
	// ========== 查询接口 ==========
	
	// GetReport 根据 device_id 和 date 获取报告详情
	GetReport(ctx context.Context, tenantID, deviceID string, date int) (*domain.SleepaceReport, error)
	
	// ListReports 查询报告列表（支持分页）
	ListReports(ctx context.Context, tenantID, deviceID string, startDate, endDate int, page, size int) ([]*domain.SleepaceReport, int, error)
	
	// GetValidDates 获取设备的所有有效日期列表
	GetValidDates(ctx context.Context, tenantID, deviceID string) ([]int, error)

	// ListReportsAllInRange 区间内全部报告（无分页），按 date 升序
	ListReportsAllInRange(ctx context.Context, tenantID, deviceID string, startDate, endDate int) ([]*domain.SleepaceReport, error)

	// GetValidDatesInRange 获取区间内已有报告的日期（YYYYMMDD）
	GetValidDatesInRange(ctx context.Context, tenantID, deviceID string, startDate, endDate int) ([]int, error)
	
	// ========== 写入接口 ==========
	
	// SaveReport 保存或更新报告（如果已存在则更新，否则插入）
	// 唯一性约束：tenant_id + device_id + date
	// device_uid 列 = devices.device_uid；若 report.DeviceID 为空，可用 DeviceUID 查 devices 得 device_id
	SaveReport(ctx context.Context, tenantID string, report *domain.SleepaceReport) error

	GetDeviceIDByDeviceUID(ctx context.Context, tenantID, deviceUID string) (string, error)
}

