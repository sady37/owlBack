package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// DevicesRepository 设备Repository接口
// 使用强类型领域模型，不使用map[string]any
type DevicesRepository interface {
	// 查询
	ListDevices(ctx context.Context, tenantID string, filters DeviceFilters, page, size int, sort string, direction string) ([]*domain.Device, int, error)
	GetDevice(ctx context.Context, tenantID, deviceID string) (*domain.Device, error)
	GetDeviceByUID(ctx context.Context, tenantID, deviceUID string) (*domain.Device, error)
	GetDeviceRelations(ctx context.Context, tenantID, deviceID string) (*DeviceRelations, error)

	// 创建（手动创建设备绑定）
	CreateDevice(ctx context.Context, tenantID string, device *domain.Device) (string, error)

	// 更新
	// Deprecated: 使用 UpdateDeviceFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	UpdateDevice(ctx context.Context, tenantID, deviceID string, device *domain.Device) error

	// UpdateDeviceWithFlags 更新设备（带更新标志，用于区分字段是否在 payload 中）
	// updateBranchID: tenant_admin rebind device 到指定 branch（仅 bed/room flag 均 false 时生效）
	UpdateDeviceWithFlags(ctx context.Context, tenantID, deviceID string, device *domain.Device, updateBoundRoomID, updateBoundBedID, updateAccess, updateMonitoringEnabled, updateBranchID bool) error


	// 删除（物理删除，仅当设备未使用时）
	DeleteDevice(ctx context.Context, tenantID, deviceID string) error

	// 自动创建（设备首次连接时自动创建）
	GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*domain.Device, error)

	// 批量查询（用于 ListUnitsWithFullHierarchy）
	GetDevicesByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) (map[string][]DeviceInfo, error)
	GetDevicesByBedIDs(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceInfo, error)

	// 按绑定关系查询（用于 DeleteUnit/DeleteRoom/DeleteBed 前检查）
	GetDevicesBoundToRoom(ctx context.Context, tenantID, roomID string) ([]*domain.Device, error)
	GetDevicesBoundToBed(ctx context.Context, tenantID, bedID string) ([]*domain.Device, error)
	GetDevicesBoundToRoomsOrBeds(ctx context.Context, tenantID string, roomIDs, bedIDs []string) ([]*domain.Device, error)

	// GetRoomBoundDeviceTypeLetters 返回绑定到 room 的设备类型字母（R=Radar, S=Sleepad），供前端 RoomName(R) 展示
	GetRoomBoundDeviceTypeLetters(ctx context.Context, tenantID, roomID string) ([]string, error)

	// GetDevicesBoundToBedsWithDetails 返回多个 bed 上绑定的设备类型字母及 monitor 状态，供 resident 弹窗 Bed(R,S) 展示
	GetDevicesBoundToBedsWithDetails(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceTypeDetail, error)
}

// DeviceTypeDetail 设备类型字母与 monitor 状态（R=Radar, S=Sleepad；Green=on, Red=off）
type DeviceTypeDetail struct {
	Letter           string
	MonitoringEnabled bool
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
	IsSystemAdmin bool     // SystemAdmin 查看所有租户的设备（不限制 tenant_id）
	Status        []string // 设备状态过滤（online, offline, error）
	DeviceType    string   // 设备类型
	SearchType    string   // 搜索类型（device_name, device_code, device_uid）
	SearchKeyword string   // 搜索关键词
	BranchPrefix  string   // Current Branch scope /56 CIDR (Phase 3)；空=tenant 全部
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
