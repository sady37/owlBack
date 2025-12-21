package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// DevicesRepository 设备Repository接口
// 使用强类型领域模型，不使用map[string]any
type DevicesRepository interface {
	// 查询
	ListDevices(ctx context.Context, tenantID string, filters DeviceFilters, page, size int) ([]*domain.Device, int, error)
	GetDevice(ctx context.Context, tenantID, deviceID string) (*domain.Device, error)

	// 创建（手动创建设备绑定）
	CreateDevice(ctx context.Context, tenantID string, device *domain.Device) (string, error)

	// 更新
	UpdateDevice(ctx context.Context, tenantID, deviceID string, device *domain.Device) error

	// 删除（物理删除，仅当设备未使用时）
	DeleteDevice(ctx context.Context, tenantID, deviceID string) error

	// 软删除（禁用设备）
	DisableDevice(ctx context.Context, tenantID, deviceID string) error

	// 自动创建（设备首次连接时自动创建）
	GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*domain.Device, error)
}

// DeviceFilters 设备查询过滤器
type DeviceFilters struct {
	Status         []string // 设备状态过滤（online, offline, error）
	BusinessAccess string   // 业务访问权限（pending, approved, rejected）
	DeviceType     string   // 设备类型
	SearchType     string   // 搜索类型（device_name, serial_number, uid）
	SearchKeyword  string   // 搜索关键词
}


