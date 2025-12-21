package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// UnitsRepository 单元Repository接口
// 使用强类型领域模型，不使用map[string]any
type UnitsRepository interface {
	// Building 操作
	ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]*domain.Building, error)
	GetBuilding(ctx context.Context, tenantID, buildingID string) (*domain.Building, error)
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
	ListRooms(ctx context.Context, tenantID, unitID string) ([]*domain.Room, error)
	ListRoomsWithBeds(ctx context.Context, tenantID, unitID string) ([]*RoomWithBeds, error)
	GetRoom(ctx context.Context, tenantID, roomID string) (*domain.Room, error)
	CreateRoom(ctx context.Context, tenantID, unitID string, room *domain.Room) (string, error)
	UpdateRoom(ctx context.Context, tenantID, roomID string, room *domain.Room) error
	DeleteRoom(ctx context.Context, tenantID, roomID string) error

	// Bed 操作
	ListBeds(ctx context.Context, tenantID, roomID string) ([]*domain.Bed, error)
	GetBed(ctx context.Context, tenantID, bedID string) (*domain.Bed, error)
	CreateBed(ctx context.Context, tenantID, roomID string, bed *domain.Bed) (string, error)
	UpdateBed(ctx context.Context, tenantID, bedID string, bed *domain.Bed) error
	DeleteBed(ctx context.Context, tenantID, bedID string) error
}

// UnitFilters 单元查询过滤器
type UnitFilters struct {
	BranchTag   string
	Building    string
	Floor       string
	AreaTag     string
	UnitNumber  string
	UnitName    string
	UnitType    string
	Search      string // 模糊搜索 unit_name, unit_number
}

// RoomWithBeds 房间及其床位（用于 ListRoomsWithBeds）
type RoomWithBeds struct {
	Room *domain.Room
	Beds []*domain.Bed
}

