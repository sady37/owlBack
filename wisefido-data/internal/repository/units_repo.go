package repository

import (
	"context"
	"database/sql"
	"wisefido-data/internal/domain"
)

// UnitsRepository 单元Repository接口
// 使用强类型领域模型，不使用map[string]any
type UnitsRepository interface {
	// Building 操作
	// ListBuildings: 查询楼栋列表
	// - branchID: 如果 Valid=true，使用 branch_id 过滤；如果 Valid=false，忽略此参数
	// - branchName: 如果 branchID 无效且 branchName 不为空，使用 branch_name 过滤（向后兼容）；如果为空，查询所有
	ListBuildings(ctx context.Context, tenantID string, branchID sql.NullString, branchName string) ([]*domain.Building, error)
	GetBuilding(ctx context.Context, tenantID, buildingID string) (*domain.Building, error)
	// GetBuildingUnits: 获取指定 building 下的所有 units（只返回 unit_id, unit_name, floor）
	GetBuildingUnits(ctx context.Context, tenantID, buildingID string) ([]BuildingUnitInfo, error)
	// FindBuildingByBranchAndName: 通过 branch_id + building_name（小写比较）查找 building，返回 building_id
	FindBuildingByBranchAndName(ctx context.Context, tenantID, branchID, buildingName string) (string, error)
	CreateBuilding(ctx context.Context, tenantID string, building *domain.Building) (string, error)
	UpdateBuilding(ctx context.Context, tenantID, buildingID string, building *domain.Building) error
	DeleteBuilding(ctx context.Context, tenantID, buildingID string) error

	// Unit 操作
	ListUnits(ctx context.Context, tenantID string, filters UnitFilters, page, size int) ([]*domain.Unit, int, error)
	GetUnit(ctx context.Context, tenantID, unitID string) (*domain.Unit, error)
	CreateUnit(ctx context.Context, tenantID string, unit *domain.Unit) (string, error)
	UpdateUnit(ctx context.Context, tenantID, unitID string, unit *domain.Unit) error
	DeleteUnit(ctx context.Context, tenantID, unitID string) error

	// Room 操作
	ListRooms(ctx context.Context, tenantID, unitID string, search string) ([]*domain.Room, error)
	ListRoomsWithBeds(ctx context.Context, tenantID, unitID string, search string) ([]*RoomWithBeds, error)
	GetRoom(ctx context.Context, tenantID, roomID string) (*domain.Room, error)
	CreateRoom(ctx context.Context, tenantID, unitID string, room *domain.Room) (string, error)
	UpdateRoom(ctx context.Context, tenantID, roomID string, room *domain.Room) error
	DeleteRoom(ctx context.Context, tenantID, roomID string) error

	// Bed 操作
	ListBeds(ctx context.Context, tenantID, roomID string, search string) ([]*domain.Bed, error)
	GetBed(ctx context.Context, tenantID, bedID string) (*domain.Bed, error)
	CreateBed(ctx context.Context, tenantID, roomID string, bed *domain.Bed) (string, error)
	UpdateBed(ctx context.Context, tenantID, bedID string, bed *domain.Bed) error
	DeleteBed(ctx context.Context, tenantID, bedID string) error

	// 批量查询（用于 ListUnitsWithFullHierarchy）
	ListRoomsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]*domain.Room, error)
	ListBedsByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) ([]*domain.Bed, error)
}

// UnitFilters 单元查询过滤器
type UnitFilters struct {
	BranchID   string   // 优先使用（通过 branch_id 过滤）
	BranchIDs  []string // 支持多个 branch_id（IN 查询，如果提供则优先使用，忽略 BranchID）
	BranchName string   // 可选（向后兼容，如果 BranchID 和 BranchIDs 都未提供则使用此字段）
	BuildingID string   // 优先使用（通过 building_id 过滤，UUID 类型）
	Building   string   // 可选（向后兼容，如果 BuildingID 未提供则使用此字段，通过 building_name 过滤）
	Floor      string
	AreaName   string
	UnitNumber string
	UnitName   string
	UnitType   string
	Search     string // 模糊搜索 unit_name, unit_number
}

// RoomWithBeds 房间及其床位（用于 ListRoomsWithBeds）
type RoomWithBeds struct {
	Room *domain.Room
	Beds []*domain.Bed
}

// BuildingUnitInfo 用于 GetBuildingUnits 返回的 unit 信息（简化版，只包含必要字段）
type BuildingUnitInfo struct {
	UnitID   string
	UnitName string
	Floor    sql.NullString
}
