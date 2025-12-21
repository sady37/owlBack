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
	
	// ========== 写入接口 ==========
	
	// SaveReport 保存或更新报告（如果已存在则更新，否则插入）
	// 唯一性约束：tenant_id + device_id + date
	// 注意：device_code 字段存储厂家的设备标识符（等价于 devices.serial_number 或 devices.uid）
	//       如果传入的 report.DeviceID 为空，可以通过 device_code 匹配 devices 表来获取 device_id
	SaveReport(ctx context.Context, tenantID string, report *domain.SleepaceReport) error
	
	// GetDeviceIDByDeviceCode 根据 device_code 获取 device_id
	// device_code 等价于 devices.serial_number 或 devices.uid
	// 用于在保存报告时，如果只有 device_code，需要通过此方法找到对应的 device_id
	GetDeviceIDByDeviceCode(ctx context.Context, tenantID, deviceCode string) (string, error)
}

