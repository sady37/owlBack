package repository

import (
	"context"
	"database/sql"

	"wisefido-qinglan/internal/domain"
	"owl-common/alarm"
)

// DeviceStoreInfo 设备库存信息（用于认证）
// 对应 device_store 表的完整结构
type DeviceStoreInfo struct {
	DeviceID        string         // UUID, PRIMARY KEY
	DeviceUID       string         // VARCHAR(50), UNIQUE
	DeviceType      string         // VARCHAR(50), NOT NULL
	DeviceModel     sql.NullString
	MAC             sql.NullString
	IMEI            sql.NullString
	CommMode        sql.NullString
	MCUModel        sql.NullString
	FirmwareVersion sql.NullString
	TenantID        string         // UUID, NOT NULL
	AllowAccess     bool           // BOOLEAN, NOT NULL DEFAULT FALSE
}

// DeviceRepository 设备仓库接口
// 定义设备相关的数据访问契约
type DeviceRepository interface {
	// GetDeviceByUID 根据设备UID获取设备信息
	GetDeviceByUID(ctx context.Context, uid string) (*domain.Device, error)

	// GetDeviceLocationInfo 获取设备位置信息
	GetDeviceLocationInfo(ctx context.Context, uid string) (*domain.DeviceLocationInfo, error)

	// GetDevicesByTenant 根据租户获取设备列表
	GetDevicesByTenant(ctx context.Context, tenantID string) ([]*domain.Device, error)

	// UpdateDeviceStatus 更新设备状态
	UpdateDeviceStatus(ctx context.Context, uid, status string) error

	// UpdateDeviceMonitoring 更新设备监控状态
	UpdateDeviceMonitoring(ctx context.Context, uid string, enabled bool) error

	// GetDeviceProperties 获取设备属性
	GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error)

	// SetDeviceProperties 设置设备属性
	SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error

	// CreateDevice 创建设备
	CreateDevice(ctx context.Context, device *domain.Device) error

	// UpdateDevice 更新设备信息
	UpdateDevice(ctx context.Context, device *domain.Device) error

	// DeleteDevice 删除设备（软删除）
	DeleteDevice(ctx context.Context, uid string) error

	// SearchDevices 搜索设备
	SearchDevices(ctx context.Context, criteria map[string]interface{}) ([]*domain.Device, error)

	// CountDevicesByStatus 按状态统计设备数量
	CountDevicesByStatus(ctx context.Context, tenantID string) (map[string]int, error)

	// GetDeviceStoreInfoAndLocation 根据设备UID获取 device_store 信息和位置信息（用于认证）
	// 一次性查询 device_store 和 devices 表（LEFT JOIN），获取所有信息包括位置信息
	GetDeviceStoreInfoAndLocation(ctx context.Context, deviceUID string) (*DeviceStoreInfo, *domain.DeviceLocationInfo, error)

	// GetAlarmEnablement 获取设备的报警使能配置
	// 从 alarm_device.monitor_config.alarms 中解析报警项，返回 []AlarmEnablementItem
	// 先查缓存，缓存未命中再查数据库并存入缓存
	GetAlarmEnablement(ctx context.Context, tenantID, deviceUID string) ([]alarm.AlarmEnablementItem, error)

	// ClearAlarmEnablementCache 清除指定设备的报警使能配置缓存
	ClearAlarmEnablementCache(tenantID, deviceUID string)

	// PreloadAlarmEnablement 预加载指定设备的报警使能配置到缓存
	PreloadAlarmEnablement(ctx context.Context, tenantID, deviceUID string) error

	// GetAllAccessibleDevices 获取所有可访问的设备（用于启动时主动订阅）
	// 条件：device_store.allow_access = TRUE 且 devices.business_access = 'approved'
	GetAllAccessibleDevices(ctx context.Context) ([]string, error)
}