package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// UnitService 单元管理服务接口
type UnitService interface {
	// Building 管理
	ListBuildings(ctx context.Context, req ListBuildingsRequest) (*ListBuildingsResponse, error)
	GetBuilding(ctx context.Context, req GetBuildingRequest) (*GetBuildingResponse, error)
	CreateBuilding(ctx context.Context, req CreateBuildingRequest) (*CreateBuildingResponse, error)
	UpdateBuilding(ctx context.Context, req UpdateBuildingRequest) (*UpdateBuildingResponse, error)
	DeleteBuilding(ctx context.Context, req DeleteBuildingRequest) (*DeleteBuildingResponse, error)

	// Unit 管理
	ListUnits(ctx context.Context, req ListUnitsRequest) (*ListUnitsResponse, error)
	GetUnit(ctx context.Context, req GetUnitRequest) (*GetUnitResponse, error)
	CreateUnit(ctx context.Context, req CreateUnitRequest) (*CreateUnitResponse, error)
	UpdateUnit(ctx context.Context, req UpdateUnitRequest) (*UpdateUnitResponse, error)
	DeleteUnit(ctx context.Context, req DeleteUnitRequest) (*DeleteUnitResponse, error)

	// Room 管理
	ListRooms(ctx context.Context, req ListRoomsRequest) (*ListRoomsResponse, error)
	ListRoomsWithBeds(ctx context.Context, req ListRoomsWithBedsRequest) (*ListRoomsWithBedsResponse, error)
	GetRoom(ctx context.Context, req GetRoomRequest) (*GetRoomResponse, error)
	CreateRoom(ctx context.Context, req CreateRoomRequest) (*CreateRoomResponse, error)
	UpdateRoom(ctx context.Context, req UpdateRoomRequest) (*UpdateRoomResponse, error)
	DeleteRoom(ctx context.Context, req DeleteRoomRequest) (*DeleteRoomResponse, error)

	// Bed 管理
	ListBeds(ctx context.Context, req ListBedsRequest) (*ListBedsResponse, error)
	GetBed(ctx context.Context, req GetBedRequest) (*GetBedResponse, error)
	CreateBed(ctx context.Context, req CreateBedRequest) (*CreateBedResponse, error)
	UpdateBed(ctx context.Context, req UpdateBedRequest) (*UpdateBedResponse, error)
	DeleteBed(ctx context.Context, req DeleteBedRequest) (*DeleteBedResponse, error)
}

// unitService 实现
type unitService struct {
	unitsRepo repository.UnitsRepository
	logger    *zap.Logger
}

// NewUnitService 创建 UnitService 实例
func NewUnitService(unitsRepo repository.UnitsRepository, logger *zap.Logger) UnitService {
	return &unitService{
		unitsRepo: unitsRepo,
		logger:    logger,
	}
}

// ============================================
// Building 相关请求/响应结构
// ============================================

type ListBuildingsRequest struct {
	TenantID  string // 必填
	BranchName string // 可选
}

type ListBuildingsResponse struct {
	Items []*domain.Building `json:"items"`
}

type GetBuildingRequest struct {
	TenantID   string // 必填
	BuildingID string // 必填
}

type GetBuildingResponse struct {
	Building *domain.Building `json:"building"`
}

type CreateBuildingRequest struct {
	TenantID    string // 必填
	BranchName   string // 可选
	BuildingName string // 必填（branch_tag 或 building_name 至少一个）
}

type CreateBuildingResponse struct {
	BuildingID string `json:"building_id"`
}

type UpdateBuildingRequest struct {
	TenantID     string // 必填
	BuildingID   string // 必填
	BranchName    string // 可选
	BuildingName string // 可选
}

type UpdateBuildingResponse struct {
	Success bool `json:"success"`
}

type DeleteBuildingRequest struct {
	TenantID   string // 必填
	BuildingID string // 必填
}

type DeleteBuildingResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Unit 相关请求/响应结构
// ============================================

type ListUnitsRequest struct {
	TenantID   string  // 必填
	BranchName  *string // 可选（nil 表示匹配 NULL）
	Building   *string // 可选（nil 表示未提供）
	Floor      *string // 可选（nil 表示未提供）
	AreaName    *string // 可选（nil 表示未提供）
	UnitNumber *string // 可选（nil 表示未提供）
	UnitName   *string // 可选（nil 表示未提供）
	UnitType   *string // 可选（nil 表示未提供）
	Search     *string // 可选（nil 表示未提供，模糊搜索 unit_name, unit_number）
	Page       int     // 可选，默认 1
	Size       int     // 可选，默认 100
}

type ListUnitsResponse struct {
	Items []*domain.Unit `json:"items"`
	Total int            `json:"total"`
}

type GetUnitRequest struct {
	TenantID string // 必填
	UnitID   string // 必填
}

type GetUnitResponse struct {
	Unit *domain.Unit `json:"unit"`
}

type CreateUnitRequest struct {
	TenantID          string // 必填
	BranchName         string // 可选
	UnitName          string // 必填
	Building          string // 可选（保持空字符串，不使用 "-"）
	Floor             string // 可选（默认 "1F"）
	AreaName           string // 可选
	UnitNumber        string // 必填
	LayoutConfig      string // 可选（JSON 字符串）
	UnitType          string // 必填
	IsPublicSpace     bool   // 可选（默认 false）
	IsMultiPersonRoom bool   // 可选（默认 false）
	Timezone          string // 必填
}

type CreateUnitResponse struct {
	UnitID string `json:"unit_id"`
}

type UpdateUnitRequest struct {
	TenantID          string // 必填
	UnitID            string // 必填
	BranchName         string // 可选
	UnitName          string // 可选
	Building          string // 可选
	Floor             string // 可选
	AreaName           string // 可选
	UnitNumber        string // 可选
	LayoutConfig      string // 可选（JSON 字符串）
	UnitType          string // 可选
	IsPublicSpace     *bool  // 可选（指针类型，nil 表示不更新）
	IsMultiPersonRoom *bool  // 可选（指针类型，nil 表示不更新）
	Timezone          string // 可选
}

type UpdateUnitResponse struct {
	Success bool `json:"success"`
}

type DeleteUnitRequest struct {
	TenantID string // 必填
	UnitID   string // 必填
}

type DeleteUnitResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Room 相关请求/响应结构
// ============================================

type ListRoomsRequest struct {
	TenantID string // 必填
	UnitID   string // 必填
}

type ListRoomsResponse struct {
	Items []*domain.Room `json:"items"`
}

type ListRoomsWithBedsRequest struct {
	TenantID string // 必填
	UnitID   string // 必填
}

type ListRoomsWithBedsResponse struct {
	Items []*repository.RoomWithBeds `json:"items"`
}

type GetRoomRequest struct {
	TenantID string // 必填
	RoomID   string // 必填
}

type GetRoomResponse struct {
	Room *domain.Room `json:"room"`
}

type CreateRoomRequest struct {
	TenantID     string // 必填
	UnitID       string // 必填
	RoomName     string // 必填
	LayoutConfig string // 可选（JSON 字符串）
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type UpdateRoomRequest struct {
	TenantID     string // 必填
	RoomID       string // 必填
	RoomName     string // 可选
	LayoutConfig string // 可选（JSON 字符串）
}

type UpdateRoomResponse struct {
	Success bool `json:"success"`
}

type DeleteRoomRequest struct {
	TenantID string // 必填
	RoomID   string // 必填
}

type DeleteRoomResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Bed 相关请求/响应结构
// ============================================

type ListBedsRequest struct {
	TenantID string // 必填
	RoomID   string // 必填
}

type ListBedsResponse struct {
	Items []*domain.Bed `json:"items"`
}

type GetBedRequest struct {
	TenantID string // 必填
	BedID    string // 必填
}

type GetBedResponse struct {
	Bed *domain.Bed `json:"bed"`
}

type CreateBedRequest struct {
	TenantID         string // 必填
	RoomID           string // 必填
	BedName          string // 必填
	// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	MattressMaterial string // 可选
	MattressThickness string // 可选
}

type CreateBedResponse struct {
	BedID string `json:"bed_id"`
}

type UpdateBedRequest struct {
	TenantID         string // 必填
	BedID            string // 必填
	BedName          string // 可选
	// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	MattressMaterial string // 可选
	MattressThickness string // 可选
}

type UpdateBedResponse struct {
	Success bool `json:"success"`
}

type DeleteBedRequest struct {
	TenantID string // 必填
	BedID    string // 必填
}

type DeleteBedResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Building 方法实现
// ============================================

// ListBuildings 查询楼栋列表
func (s *unitService) ListBuildings(ctx context.Context, req ListBuildingsRequest) (*ListBuildingsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	items, err := s.unitsRepo.ListBuildings(ctx, req.TenantID, req.BranchName)
	if err != nil {
		s.logger.Error("ListBuildings failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_name", req.BranchName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list buildings: %w", err)
	}

	return &ListBuildingsResponse{
		Items: items,
	}, nil
}

// GetBuilding 获取单个楼栋详情
func (s *unitService) GetBuilding(ctx context.Context, req GetBuildingRequest) (*GetBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BuildingID == "" {
		return nil, fmt.Errorf("building_id is required")
	}

	building, err := s.unitsRepo.GetBuilding(ctx, req.TenantID, req.BuildingID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("building not found")
		}
		s.logger.Error("GetBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_id", req.BuildingID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	return &GetBuildingResponse{
		Building: building,
	}, nil
}

// CreateBuilding 创建楼栋
func (s *unitService) CreateBuilding(ctx context.Context, req CreateBuildingRequest) (*CreateBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 验证：branch_name 或 building_name 必须有一个不为空
	branchNameValue := strings.TrimSpace(req.BranchName)
	if (branchNameValue == "" || branchNameValue == "-") && (req.BuildingName == "" || req.BuildingName == "-") {
		return nil, fmt.Errorf("branch_name or building_name must be provided (at least one must not be empty)")
	}

	building := &domain.Building{
		TenantID:    req.TenantID,
		BranchTag:   normalizeBranchTag(req.BranchName),
		BuildingName: strings.TrimSpace(req.BuildingName),
	}

	// 设置默认值
	if building.BuildingName == "" {
		building.BuildingName = "-"
	}

	buildingID, err := s.unitsRepo.CreateBuilding(ctx, req.TenantID, building)
	if err != nil {
		s.logger.Error("CreateBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_name", req.BuildingName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create building: %w", err)
	}

	return &CreateBuildingResponse{
		BuildingID: buildingID,
	}, nil
}

// UpdateBuilding 更新楼栋
func (s *unitService) UpdateBuilding(ctx context.Context, req UpdateBuildingRequest) (*UpdateBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BuildingID == "" {
		return nil, fmt.Errorf("building_id is required")
	}

	// 获取当前 building
	currentBuilding, err := s.unitsRepo.GetBuilding(ctx, req.TenantID, req.BuildingID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	// 构建更新后的 building
	building := &domain.Building{
		BuildingID:   req.BuildingID,
		TenantID:     req.TenantID,
		BranchTag:    currentBuilding.BranchTag,
		BuildingName: currentBuilding.BuildingName,
	}

	// 更新提供的字段
	if req.BranchName != "" || req.BranchName == "" { // 允许设置为空
		building.BranchTag = normalizeBranchTag(req.BranchName)
	}
	if req.BuildingName != "" {
		building.BuildingName = strings.TrimSpace(req.BuildingName)
	}

	// 设置默认值
	if building.BuildingName == "" {
		building.BuildingName = "-"
	}

	// 验证：branch_tag 或 building_name 必须有一个不为空（更新后）
	branchTagValue := ""
	if building.BranchTag.Valid {
		branchTagValue = building.BranchTag.String
	}
	buildingNameValue := strings.TrimSpace(building.BuildingName)
	if (branchTagValue == "" || branchTagValue == "-") && (buildingNameValue == "" || buildingNameValue == "-") {
		return nil, fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
	}

	err = s.unitsRepo.UpdateBuilding(ctx, req.TenantID, req.BuildingID, building)
	if err != nil {
		s.logger.Error("UpdateBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_id", req.BuildingID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update building: %w", err)
	}

	return &UpdateBuildingResponse{
		Success: true,
	}, nil
}

// DeleteBuilding 删除楼栋
func (s *unitService) DeleteBuilding(ctx context.Context, req DeleteBuildingRequest) (*DeleteBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BuildingID == "" {
		return nil, fmt.Errorf("building_id is required")
	}

	err := s.unitsRepo.DeleteBuilding(ctx, req.TenantID, req.BuildingID)
	if err != nil {
		s.logger.Error("DeleteBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_id", req.BuildingID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete building: %w", err)
	}

	return &DeleteBuildingResponse{
		Success: true,
	}, nil
}

// ============================================
// Unit 方法实现
// ============================================

// ListUnits 查询单元列表
func (s *unitService) ListUnits(ctx context.Context, req ListUnitsRequest) (*ListUnitsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 2. 构建过滤器
	// 关键逻辑：当提供了 building 时，必须同时匹配 branch_tag 和 building
	// - branch_tag 为 nil：查询 branch_tag IS NULL
	// - branch_tag 不为 nil：查询 branch_tag = X
	// - building 为 nil：不添加 building 过滤条件
	// - building 不为 nil：添加 building 过滤条件
	// 空字符串视为 null（nil → ""，用于 Repository 层转换为 IS NULL）
	filters := repository.UnitFilters{
		BranchName:  stringValueOrEmpty(req.BranchName),
		Building:   stringValueOrEmpty(req.Building),
		Floor:      stringValueOrEmpty(req.Floor),
		AreaName:    stringValueOrEmpty(req.AreaName),
		UnitNumber: stringValueOrEmpty(req.UnitNumber),
		UnitName:   stringValueOrEmpty(req.UnitName),
		UnitType:   stringValueOrEmpty(req.UnitType),
		Search:     stringValueOrEmpty(req.Search),
	}

	// 3. 分页参数（与旧 Handler 逻辑对齐：默认 page=1, size=100）
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 100
	}

	// 4. 调用 Repository
	items, total, err := s.unitsRepo.ListUnits(ctx, req.TenantID, filters, page, size)
	if err != nil {
		s.logger.Error("ListUnits failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	// 5. 构建响应
	return &ListUnitsResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetUnit 获取单个单元详情
func (s *unitService) GetUnit(ctx context.Context, req GetUnitRequest) (*GetUnitResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	// 2. 调用 Repository
	unit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("GetUnit: unit not found",
				zap.String("tenant_id", req.TenantID),
				zap.String("unit_id", req.UnitID),
			)
			return nil, fmt.Errorf("unit not found")
		}
		s.logger.Error("GetUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 3. 构建响应
	return &GetUnitResponse{
		Unit: unit,
	}, nil
}

// CreateUnit 创建单元
func (s *unitService) CreateUnit(ctx context.Context, req CreateUnitRequest) (*CreateUnitResponse, error) {
	// 1. 参数验证（必填字段）
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitName == "" {
		return nil, fmt.Errorf("unit_name is required")
	}
	if req.UnitNumber == "" {
		return nil, fmt.Errorf("unit_number is required")
	}

	// 2. 应用默认值和格式转换（可选字段）
	unitType := normalizeUnitType(req.UnitType)      // "" → "Facility"
	building := normalizeBuilding(req.Building)      // 保持空字符串 ''（不再使用 "-"）
	floor := normalizeFloor(req.Floor)                // ""/"1"/1 → sql.NullString{String: "1F", Valid: true}
	timezone := normalizeTimezone(req.Timezone)       // "" → "America/Denver" (IANA 标识符)

	// 3. 构建 domain.Unit
	unit := &domain.Unit{
		TenantID:          req.TenantID,
		BranchName:         normalizeBranchTag(req.BranchName),
		UnitName:          strings.TrimSpace(req.UnitName),
		Building:          building,
		Floor:             floor,
		AreaName:           normalizeAreaTag(req.AreaName),
		UnitNumber:        strings.TrimSpace(req.UnitNumber),
		LayoutConfig:      normalizeLayoutConfig(req.LayoutConfig),
		UnitType:          unitType,
		IsPublicSpace:     req.IsPublicSpace,
		IsMultiPersonRoom: req.IsMultiPersonRoom,
		Timezone:          timezone,
	}

	// 4. 业务规则验证
	// 如果 Unit 没有 building，则必须提供 branch_name
	// 如果 Unit 有 building，则不需要验证（Building 的 Service 层已经保证了 branch_name 或 building_name 至少有一个不为空）
	if !unit.Building.Valid {
		branchNameValue := ""
		if unit.BranchName.Valid {
			branchNameValue = unit.BranchName.String
		}
		if branchNameValue == "" || branchNameValue == "-" {
			return nil, fmt.Errorf("branch_name is required when building is not provided")
		}
	}

	// 5. 调用 Repository
	unitID, err := s.unitsRepo.CreateUnit(ctx, req.TenantID, unit)
	if err != nil {
		// 检查唯一约束错误（Repository 会返回数据库错误）
		s.logger.Error("CreateUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_name", req.UnitName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}

	// 6. 构建响应
	return &CreateUnitResponse{
		UnitID: unitID,
	}, nil
}

// UpdateUnit 更新单元
func (s *unitService) UpdateUnit(ctx context.Context, req UpdateUnitRequest) (*UpdateUnitResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	// 2. 先获取当前 unit（用于部分更新）
	currentUnit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("UpdateUnit: unit not found",
				zap.String("tenant_id", req.TenantID),
				zap.String("unit_id", req.UnitID),
			)
			return nil, fmt.Errorf("unit not found")
		}
		s.logger.Error("UpdateUnit: failed to get current unit",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 3. 构建更新后的 unit（只更新提供的字段）
	unit := &domain.Unit{
		UnitID:            req.UnitID,
		TenantID:          req.TenantID,
		BranchName:         currentUnit.BranchName,
		UnitName:          currentUnit.UnitName,
		Building:          currentUnit.Building,
		Floor:             currentUnit.Floor,
		AreaName:           currentUnit.AreaName,
		UnitNumber:        currentUnit.UnitNumber,
		LayoutConfig:      currentUnit.LayoutConfig,
		UnitType:          currentUnit.UnitType,
		IsPublicSpace:     currentUnit.IsPublicSpace,
		IsMultiPersonRoom: currentUnit.IsMultiPersonRoom,
		Timezone:          currentUnit.Timezone,
		GroupList:         currentUnit.GroupList,
		UserList:          currentUnit.UserList,
	}

	// 更新提供的字段
	// Repository 的 UpdateUnit 逻辑：
	// - 如果 unit.BranchName.Valid == true，会更新（即使 String 为空，也会设置为 NULL）
	// - 如果 unit.BranchName.Valid == false，不会更新（保持原值）
	// 
	// Service 层策略：
	// - 如果请求中提供了非空值，设置 Valid=true 和 String 值
	// - 如果请求中提供了空字符串，设置 Valid=true 和 String=""（Repository 会转换为 NULL）
	// - 如果请求中未提供（空字符串且当前值不存在），保持 Valid=false（不更新）
	// 
	// 注意：由于无法区分"未提供"和"空字符串"，我们采用：
	// - 非空值：更新
	// - 空字符串：如果当前值存在，清除它（设置为 NULL）；否则不更新
	
	// branch_name: 如果提供了非空值，更新；如果为空字符串，清除（Repository 会处理为 NULL）
	if req.BranchName != "" {
		unit.BranchName = normalizeBranchTag(req.BranchName)
	} else if req.BranchName == "" && currentUnit.BranchName.Valid {
		// 请求值为空且当前值存在，清除它（设置为 Valid=true, String=""，Repository 会转换为 NULL）
		unit.BranchName = sql.NullString{String: "", Valid: true}
	}
	// 如果 req.BranchName == "" 且当前值不存在，保持 unit.BranchName.Valid = false（不更新）
	
	if req.UnitName != "" {
		unit.UnitName = strings.TrimSpace(req.UnitName)
	}
	if req.Building != "" {
		unit.Building = normalizeBuilding(req.Building)
	} else {
		// 如果请求中未提供 building，保持原值（不更新）
		unit.Building = currentUnit.Building
	}
	if req.Floor != "" {
		unit.Floor = normalizeFloor(req.Floor)
	}
	
	// area_name: 类似 branch_name
	if req.AreaName != "" {
		unit.AreaName = normalizeAreaTag(req.AreaName)
	} else if req.AreaName == "" && currentUnit.AreaName.Valid {
		unit.AreaName = sql.NullString{String: "", Valid: true}
	}
	
	if req.UnitNumber != "" {
		unit.UnitNumber = strings.TrimSpace(req.UnitNumber)
	}
	
	// layout_config: 类似处理
	if req.LayoutConfig != "" {
		unit.LayoutConfig = normalizeLayoutConfig(req.LayoutConfig)
	} else if req.LayoutConfig == "" && currentUnit.LayoutConfig.Valid {
		unit.LayoutConfig = sql.NullString{String: "", Valid: true}
	}
	
	if req.UnitType != "" {
		unit.UnitType = normalizeUnitType(req.UnitType)
	}
	if req.IsPublicSpace != nil {
		unit.IsPublicSpace = *req.IsPublicSpace
	}
	if req.IsMultiPersonRoom != nil {
		unit.IsMultiPersonRoom = *req.IsMultiPersonRoom
	}
	if req.Timezone != "" {
		unit.Timezone = normalizeTimezone(req.Timezone)
	}

	// 4. 调用 Repository
	err = s.unitsRepo.UpdateUnit(ctx, req.TenantID, req.UnitID, unit)
	if err != nil {
		s.logger.Error("UpdateUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update unit: %w", err)
	}

	// 5. 构建响应
	return &UpdateUnitResponse{
		Success: true,
	}, nil
}

// DeleteUnit 删除单元
func (s *unitService) DeleteUnit(ctx context.Context, req DeleteUnitRequest) (*DeleteUnitResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	// 2. 调用 Repository
	err := s.unitsRepo.DeleteUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		s.logger.Error("DeleteUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete unit: %w", err)
	}

	// 3. 构建响应
	return &DeleteUnitResponse{
		Success: true,
	}, nil
}

// ============================================
// Room 方法实现
// ============================================

// ListRooms 查询房间列表
func (s *unitService) ListRooms(ctx context.Context, req ListRoomsRequest) (*ListRoomsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	items, err := s.unitsRepo.ListRooms(ctx, req.TenantID, req.UnitID)
	if err != nil {
		s.logger.Error("ListRooms failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	return &ListRoomsResponse{
		Items: items,
	}, nil
}

// ListRoomsWithBeds 查询房间及其床位列表
func (s *unitService) ListRoomsWithBeds(ctx context.Context, req ListRoomsWithBedsRequest) (*ListRoomsWithBedsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	items, err := s.unitsRepo.ListRoomsWithBeds(ctx, req.TenantID, req.UnitID)
	if err != nil {
		s.logger.Error("ListRoomsWithBeds failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms with beds: %w", err)
	}

	return &ListRoomsWithBedsResponse{
		Items: items,
	}, nil
}

// GetRoom 获取单个房间详情
func (s *unitService) GetRoom(ctx context.Context, req GetRoomRequest) (*GetRoomResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	room, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		s.logger.Error("GetRoom failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return &GetRoomResponse{
		Room: room,
	}, nil
}

// CreateRoom 创建房间
func (s *unitService) CreateRoom(ctx context.Context, req CreateRoomRequest) (*CreateRoomResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}
	if req.RoomName == "" {
		return nil, fmt.Errorf("room_name is required")
	}

	room := &domain.Room{
		TenantID:     req.TenantID,
		UnitID:       req.UnitID,
		RoomName:     strings.TrimSpace(req.RoomName),
		LayoutConfig: normalizeLayoutConfig(req.LayoutConfig),
	}

	roomID, err := s.unitsRepo.CreateRoom(ctx, req.TenantID, req.UnitID, room)
	if err != nil {
		s.logger.Error("CreateRoom failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.String("room_name", req.RoomName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return &CreateRoomResponse{
		RoomID: roomID,
	}, nil
}

// UpdateRoom 更新房间
func (s *unitService) UpdateRoom(ctx context.Context, req UpdateRoomRequest) (*UpdateRoomResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	// 获取当前 room
	currentRoom, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// 构建更新后的 room
	room := &domain.Room{
		RoomID:       req.RoomID,
		TenantID:     req.TenantID,
		UnitID:       currentRoom.UnitID,
		RoomName:     currentRoom.RoomName,
		LayoutConfig: currentRoom.LayoutConfig,
	}

	// 更新提供的字段
	if req.RoomName != "" {
		room.RoomName = strings.TrimSpace(req.RoomName)
	}
	if req.LayoutConfig != "" || req.LayoutConfig == "" { // 允许设置为空
		room.LayoutConfig = normalizeLayoutConfig(req.LayoutConfig)
	}

	err = s.unitsRepo.UpdateRoom(ctx, req.TenantID, req.RoomID, room)
	if err != nil {
		s.logger.Error("UpdateRoom failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return &UpdateRoomResponse{
		Success: true,
	}, nil
}

// DeleteRoom 删除房间
func (s *unitService) DeleteRoom(ctx context.Context, req DeleteRoomRequest) (*DeleteRoomResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	err := s.unitsRepo.DeleteRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		s.logger.Error("DeleteRoom failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete room: %w", err)
	}

	return &DeleteRoomResponse{
		Success: true,
	}, nil
}

// ============================================
// Bed 方法实现
// ============================================

// ListBeds 查询床位列表
func (s *unitService) ListBeds(ctx context.Context, req ListBedsRequest) (*ListBedsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	items, err := s.unitsRepo.ListBeds(ctx, req.TenantID, req.RoomID)
	if err != nil {
		s.logger.Error("ListBeds failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list beds: %w", err)
	}

	return &ListBedsResponse{
		Items: items,
	}, nil
}

// GetBed 获取单个床位详情
func (s *unitService) GetBed(ctx context.Context, req GetBedRequest) (*GetBedResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}

	bed, err := s.unitsRepo.GetBed(ctx, req.TenantID, req.BedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bed not found")
		}
		s.logger.Error("GetBed failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get bed: %w", err)
	}

	return &GetBedResponse{
		Bed: bed,
	}, nil
}

// CreateBed 创建床位
func (s *unitService) CreateBed(ctx context.Context, req CreateBedRequest) (*CreateBedResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}
	if req.BedName == "" {
		return nil, fmt.Errorf("bed_name is required")
	}

	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bed := &domain.Bed{
		TenantID:         req.TenantID,
		RoomID:            req.RoomID,
		BedName:          strings.TrimSpace(req.BedName),
		MattressMaterial: normalizeMattressMaterial(req.MattressMaterial),
		MattressThickness: normalizeMattressThickness(req.MattressThickness),
	}

	bedID, err := s.unitsRepo.CreateBed(ctx, req.TenantID, req.RoomID, bed)
	if err != nil {
		s.logger.Error("CreateBed failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.String("bed_name", req.BedName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create bed: %w", err)
	}

	return &CreateBedResponse{
		BedID: bedID,
	}, nil
}

// UpdateBed 更新床位
func (s *unitService) UpdateBed(ctx context.Context, req UpdateBedRequest) (*UpdateBedResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}

	// 获取当前 bed
	currentBed, err := s.unitsRepo.GetBed(ctx, req.TenantID, req.BedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bed not found")
		}
		return nil, fmt.Errorf("failed to get bed: %w", err)
	}

	// 构建更新后的 bed
	bed := &domain.Bed{
		BedID:            req.BedID,
		TenantID:         req.TenantID,
		RoomID:           currentBed.RoomID,
		BedName:          currentBed.BedName,
		MattressMaterial: currentBed.MattressMaterial,
		MattressThickness: currentBed.MattressThickness,
	}

	// 更新提供的字段
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	if req.BedName != "" {
		bed.BedName = strings.TrimSpace(req.BedName)
	}
	if req.MattressMaterial != "" || req.MattressMaterial == "" { // 允许设置为空
		bed.MattressMaterial = normalizeMattressMaterial(req.MattressMaterial)
	}
	if req.MattressThickness != "" || req.MattressThickness == "" { // 允许设置为空
		bed.MattressThickness = normalizeMattressThickness(req.MattressThickness)
	}

	err = s.unitsRepo.UpdateBed(ctx, req.TenantID, req.BedID, bed)
	if err != nil {
		s.logger.Error("UpdateBed failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update bed: %w", err)
	}

	return &UpdateBedResponse{
		Success: true,
	}, nil
}

// DeleteBed 删除床位
func (s *unitService) DeleteBed(ctx context.Context, req DeleteBedRequest) (*DeleteBedResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}

	err := s.unitsRepo.DeleteBed(ctx, req.TenantID, req.BedID)
	if err != nil {
		s.logger.Error("DeleteBed failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete bed: %w", err)
	}

	return &DeleteBedResponse{
		Success: true,
	}, nil
}

// ============================================
// 辅助函数
// ============================================

// stringValueOrEmpty 将 *string 转换为 string（nil → ""，非 nil → 去除首尾空格）
// 用于将 Service 层的 *string（nil 表示 null）转换为 Repository 层的 string（"" 表示 null）
func stringValueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// normalizeBranchTag 规范化 branch_tag：空字符串或 "-" 视为 NULL
func normalizeBranchTag(branchTag string) sql.NullString {
	if branchTag == "" || branchTag == "-" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: branchTag, Valid: true}
}

// normalizeAreaTag 规范化 area_tag：空字符串视为 NULL
func normalizeAreaTag(areaTag string) sql.NullString {
	if areaTag == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: areaTag, Valid: true}
}

// normalizeLayoutConfig 规范化 layout_config：空字符串视为 NULL
func normalizeLayoutConfig(layoutConfig string) sql.NullString {
	if layoutConfig == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: layoutConfig, Valid: true}
}

// normalizeMattressMaterial 规范化 mattress_material：空字符串视为 NULL
func normalizeMattressMaterial(material string) sql.NullString {
	if material == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: material, Valid: true}
}

// normalizeMattressThickness 规范化 mattress_thickness：空字符串视为 NULL
func normalizeMattressThickness(thickness string) sql.NullString {
	if thickness == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: thickness, Valid: true}
}

// normalizeBuilding 规范化 building：空字符串或 "-" → NULL，否则返回 trimmed 字符串
func normalizeBuilding(building string) sql.NullString {
	b := strings.TrimSpace(building)
	if b == "" || b == "-" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: b, Valid: true}
}

// normalizeUnitType 规范化 unit_type：空字符串 → "Facility"
func normalizeUnitType(unitType string) string {
	t := strings.TrimSpace(unitType)
	if t == "" {
		return "Facility"
	}
	return t
}

// normalizeFloor 规范化 floor：
// - number (1) → sql.NullString{String: "1F", Valid: true}
// - string without "F" ("1") → sql.NullString{String: "1F", Valid: true}
// - string with "F" ("1F") → sql.NullString{String: "1F", Valid: true}
// - empty string → sql.NullString{String: "1F", Valid: true} (default)
func normalizeFloor(floor interface{}) sql.NullString {
	if floor == nil {
		return sql.NullString{String: "1F", Valid: true}
	}

	var floorStr string
	switch v := floor.(type) {
	case int:
		floorStr = fmt.Sprintf("%dF", v)
	case float64:
		floorStr = fmt.Sprintf("%.0fF", v)
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			floorStr = "1F"
		} else {
			// 如果是纯数字，添加 "F" 后缀
			if matched, _ := regexp.MatchString(`^\d+$`, s); matched {
				floorStr = s + "F"
			} else if strings.HasSuffix(s, "F") || strings.HasSuffix(s, "f") {
				// 如果已经有 "F" 或 "f" 后缀，保持不变
				floorStr = s
			} else {
				// 其他情况，添加 "F" 后缀
				floorStr = s + "F"
			}
		}
	default:
		floorStr = "1F"
	}
	return sql.NullString{String: floorStr, Valid: true}
}

// normalizeTimezone 规范化 timezone：空字符串 → "America/Denver" (Mountain Time, 有夏令时)
func normalizeTimezone(timezone string) string {
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		return "America/Denver"
	}
	return tz
}

