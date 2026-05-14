package repository

import (
	"context"
	"database/sql"
	"net/netip"

	"owl-common/alarm"
	"wisefido-qinglan/internal/domain"
)

// DeviceStoreInfo 设备库存信息（用于认证 + 路由）
//
// device_ipv6 单程票 (doc/device_ipv6_migration_checklist.md) 后：
//   - DeviceAddr 是路由层主键
//   - DeviceID/DeviceUID/TenantID 保留外部 API + auth 边界使用 (R-002/R-003)
type DeviceStoreInfo struct {
	DeviceID        string         // UUID, PRIMARY KEY
	DeviceUID       string         // VARCHAR(50), UNIQUE
	DeviceCode      sql.NullString // device_store.device_code
	DeviceType      string         // VARCHAR(50), NOT NULL
	DeviceModel     sql.NullString
	MAC             sql.NullString
	IMEI            sql.NullString
	CommMode        sql.NullString
	MCUModel        sql.NullString
	FirmwareVersion sql.NullString
	TenantID        string     // /48 INET CIDR text，e.g. "fd00:0:3::/48"（v2 已替换 UUID）
	Access          bool       // devices.access — platform_admin 审批位（v1 allow_access/business_access 已合并）
	DeviceAddr      netip.Addr // /128 INET，路由层主键
}

// DeviceRepository 设备仓库接口
// 定义设备相关的数据访问契约
type DeviceRepository interface {
	// GetDeviceByUID 根据设备UID获取设备信息
	GetDeviceByUID(ctx context.Context, uid string) (*domain.Device, error)

	// GetDevicesByTenant 根据租户获取设备列表
	GetDevicesByTenant(ctx context.Context, tenantID string) ([]*domain.Device, error)

	// UpdateDeviceMonitoring 更新设备监控状态
	UpdateDeviceMonitoring(ctx context.Context, uid string, enabled bool) error

	// GetDeviceProperties 获取设备属性
	GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error)

	// SetDeviceProperties 设置设备属性
	SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error

	// CreateDevice 创建设备
	//CreateDevice(ctx context.Context, device *domain.Device) error

	// UpdateDevice 更新设备信息
	UpdateDevice(ctx context.Context, device *domain.Device) error

	// SearchDevices 搜索设备
	SearchDevices(ctx context.Context, criteria map[string]interface{}) ([]*domain.Device, error)

	// CountDevicesByStatus 按状态统计设备数量
	CountDevicesByStatus(ctx context.Context, tenantID string) (map[string]int, error)

	// GetDeviceStoreInfo 根据设备UID获取 device_store 信息（含 device_code）
	GetDeviceStoreInfo(ctx context.Context, deviceUID string) (*DeviceStoreInfo, error)

	// GetDeviceStoreByDeviceID 根据设备ID获取 device_store 信息（用于 device_store 变化信号处理）
	// 返回 DeviceStoreInfo 包含 device_uid 和 access
	GetDeviceStoreByDeviceID(ctx context.Context, deviceID string) (*DeviceStoreInfo, error)

	// GetAlarmEnablement 获取设备的报警使能配置
	// 从 alarm_device.monitor_config.alarms 中解析报警项，返回 []AlarmEnablementItem
	// 先查缓存，缓存未命中再查数据库并存入缓存
	GetAlarmEnablement(ctx context.Context, tenantID, deviceUID string) ([]alarm.AlarmEnablementItem, error)

	// ClearAlarmEnablementCache 清除指定设备的报警使能配置缓存
	ClearAlarmEnablementCache(tenantID, deviceUID string)

	// PreloadAlarmEnablement 预加载指定设备的报警使能配置到缓存
	PreloadAlarmEnablement(ctx context.Context, tenantID, deviceUID string) error

	// GetAllAccessibleDevices 获取所有可访问的设备（用于启动时主动订阅）
	// 条件：dfm.device_type='Radar' AND devices.access=TRUE
	GetAllAccessibleDevices(ctx context.Context) ([]string, error)

	// GetAllDeviceStoreInfo 获取所有 device_factory_meta 记录（用于启动时初始化设备缓存）
	// 返回所有设备的 device_uid 和 access 状态
	GetAllDeviceStoreInfo(ctx context.Context) ([]*DeviceStoreInfo, error)
}
