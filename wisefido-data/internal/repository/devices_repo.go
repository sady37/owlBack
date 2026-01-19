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
	GetDeviceRelations(ctx context.Context, tenantID, deviceID string) (*DeviceRelations, error)

	// 创建（手动创建设备绑定）
	CreateDevice(ctx context.Context, tenantID string, device *domain.Device) (string, error)

	// 更新
	// Deprecated: 使用 UpdateDeviceFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	UpdateDevice(ctx context.Context, tenantID, deviceID string, device *domain.Device) error

	// UpdateDeviceWithFlags 更新设备（带更新标志，用于区分字段是否在 payload 中）
	UpdateDeviceWithFlags(ctx context.Context, tenantID, deviceID string, device *domain.Device, updateBoundRoomID, updateBoundBedID, updateBusinessAccess, updateMonitoringEnabled bool) error


	// 删除（物理删除，仅当设备未使用时）
	DeleteDevice(ctx context.Context, tenantID, deviceID string) error

	// 软删除（禁用设备）
	DisableDevice(ctx context.Context, tenantID, deviceID string) error

	// 自动创建（设备首次连接时自动创建）
	GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*domain.Device, error)

	// 批量查询（用于 ListUnitsWithFullHierarchy）
	// DeviceInfo 用于返回设备的 ID 和 Name
	GetDevicesByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) (map[string][]DeviceInfo, error)
	GetDevicesByBedIDs(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceInfo, error)
}

// DeviceInfo 用于返回设备的 ID 和 Name（用于 ListUnitsWithFullHierarchy）
type DeviceInfo struct {
	ID   string
	Name string
}

// DeviceRelations 设备关联关系
type DeviceRelations struct {
	DeviceID           string
	DeviceName         string
	DeviceInternalCode string // device_uid
	DeviceType         int    // 从 device_store.device_type 转换
	AddressID          string // unit_id
	AddressName        string // unit_name
	AddressType        int    // 0=Facility, 1=Home (从 unit_type 转换)
	Residents          []DeviceRelationResident
}

// DeviceRelationResident 设备关联的住户信息
type DeviceRelationResident struct {
	ID       string
	Name     string
	Gender   string
	Birthday string
}

// DeviceFilters 设备查询过滤器
type DeviceFilters struct {
	IsSystemAdmin  bool     // SystemAdmin 查看所有租户的设备（不限制 tenant_id）
	Status         []string // 设备状态过滤（online, offline, error）
	BusinessAccess string   // 业务访问权限（pending, approved, rejected）
	DeviceType     string   // 设备类型
	SearchType     string   // 搜索类型（device_name, device_uid）
	SearchKeyword  string   // 搜索关键词
}

// DeviceLocationInfo 设备位置信息
type DeviceLocationInfo struct {
	DeviceID     string  // 设备ID
	TenantID     string  // 租户ID
	BranchID     *string // 分支ID（可选）
	BranchName   *string // 分支名称（可选）
	BuildingID   *string // 建筑ID（可选）
	BuildingName *string // 建筑名称（可选）
	UnitID       *string // 单元ID（可选）
	UnitName     *string // 单元名称（可选）
	RoomID       *string // 房间ID（可选）
	RoomName     *string // 房间名称（可选）
	BedID        *string // 床位ID（可选）
	BedName      *string // 床位名称（可选）
}
