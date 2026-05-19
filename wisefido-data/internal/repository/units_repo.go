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
	// residentID: 住户绑定时传当前 resident_id，beds 仅返回可用；空表示仅未被占用的 bed
	ListRoomsWithBeds(ctx context.Context, tenantID, unitID, search, residentID string) ([]*RoomWithBeds, error)
	GetRoom(ctx context.Context, tenantID, roomID string) (*domain.Room, error)
	CreateRoom(ctx context.Context, tenantID, unitID string, room *domain.Room) (string, error)
	UpdateRoom(ctx context.Context, tenantID, roomID string, room *domain.Room) error
	DeleteRoom(ctx context.Context, tenantID, roomID string) error

	// Bed 操作
	ListBeds(ctx context.Context, tenantID, roomID, search string) ([]*domain.Bed, error)
	// ListAvailableBeds: 仅返回可用床位（未被占用或仅被 residentID 占用），住户绑定时使用
	ListAvailableBeds(ctx context.Context, tenantID, roomID, search, residentID string) ([]*domain.Bed, error)
	// ListBedsWithResident: 返回 room 下全部 bed 及 resident_id（供 resident 弹窗 bed 状态色）
	ListBedsWithResident(ctx context.Context, tenantID, roomID, search string) ([]*BedWithResident, error)
	GetBed(ctx context.Context, tenantID, bedID string) (*domain.Bed, error)
	CreateBed(ctx context.Context, tenantID, roomID string, bed *domain.Bed) (string, error)
	UpdateBed(ctx context.Context, tenantID, bedID string, bed *domain.Bed) error
	DeleteBed(ctx context.Context, tenantID, bedID string) error

	// 批量查询（用于 ListUnitsWithFullHierarchy）
	ListRoomsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]*domain.Room, error)
	ListBedsByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) ([]*domain.Bed, error)

	// ListRoomsByBranch: 按 branch 列出所有 room，带 unit 信息与 full/bound 标识（供前端灰行、(full) 红字、橙/绿）
	ListRoomsByBranch(ctx context.Context, tenantID, branchID string) ([]*RoomWithAvailability, error)

	// GetUnitAvailability: 批量查询 unit 的 has_available_bed、is_bound（unitIDs 为空则返回空 map）
	GetUnitAvailability(ctx context.Context, tenantID string, unitIDs []string) (hasAvailableBed, isBound map[string]bool, err error)
}

// RoomWithAvailability 按 branch 列 room 时的项：含 building、floor、unit、room、unit_property、unit_type、is_full、is_bound
type RoomWithAvailability struct {
	RoomID       string
	TenantID     string
	UnitID       string
	UnitName     string
	BuildingName string // 楼栋名，便于前端展示
	Floor        string // 楼层
	RoomName     string
	UnitProperty int  // 0=Home, 1=Facility（与 owl-common/card.UnitProperty* 对齐）
	UnitType     int  // 1=Private, 2=Share, 3=Public（与 owl-common/card.UnitType* 对齐）
	IsFull       bool // room 下所有 bed 均已被绑定
	IsBound      bool // 有住户绑定了该 room（room_id 或该 room 下 bed）
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
	UnitType   *int // UnitType*: 1=Private/2=Share/3=Public；nil=不过滤
	Search     string // 模糊搜索 unit_name, unit_number
	// ResidentID: 非 nil 表示住户绑定时，Private 单元仅返回未被其他住户占用的（nil=不过滤）
	ResidentID *string
}

// RoomWithBeds 房间及其床位（用于 ListRoomsWithBeds）
type RoomWithBeds struct {
	Room *domain.Room
	Beds []*domain.Bed
}

// BedWithResident bed + 绑定住户 id（供 resident 弹窗 bed 状态 Full/Binded/Unbind）
type BedWithResident struct {
	*domain.Bed
	ResidentID *string
}

// BuildingUnitInfo 用于 GetBuildingUnits 返回的 unit 信息（简化版，只包含必要字段）
type BuildingUnitInfo struct {
	UnitID   string
	UnitName string
	Floor    sql.NullString
}
