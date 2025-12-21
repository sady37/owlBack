package service

import (
	"context"
	"database/sql"
	"fmt"
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
	BranchTag string // 可选
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
	BranchTag   string // 可选
	BuildingName string // 必填（branch_tag 或 building_name 至少一个）
}

type CreateBuildingResponse struct {
	BuildingID string `json:"building_id"`
}

type UpdateBuildingRequest struct {
	TenantID     string // 必填
	BuildingID   string // 必填
	BranchTag    string // 可选
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
	TenantID   string // 必填
	BranchTag  string // 可选（空字符串表示匹配 NULL）
	Building   string // 可选
	Floor      string // 可选
	AreaTag    string // 可选
	UnitNumber string // 可选
	UnitName   string // 可选
	UnitType   string // 可选
	Search     string // 可选（模糊搜索 unit_name, unit_number）
	Page       int    // 可选，默认 1
	Size       int    // 可选，默认 100
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
	BranchTag         string // 可选
	UnitName          string // 必填
	Building          string // 可选（默认 "-"）
	Floor             string // 可选（默认 "1F"）
	AreaTag           string // 可选
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
	BranchTag         string // 可选
	UnitName          string // 可选
	Building          string // 可选
	Floor             string // 可选
	AreaTag           string // 可选
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
	BedType          string // 必填
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
	BedType          string // 可选
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

	items, err := s.unitsRepo.ListBuildings(ctx, req.TenantID, req.BranchTag)
	if err != nil {
		s.logger.Error("ListBuildings failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_tag", req.BranchTag),
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

	// 验证：branch_tag 或 building_name 必须有一个不为空
	branchTagValue := strings.TrimSpace(req.BranchTag)
	if (branchTagValue == "" || branchTagValue == "-") && (req.BuildingName == "" || req.BuildingName == "-") {
		return nil, fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
	}

	building := &domain.Building{
		TenantID:    req.TenantID,
		BranchTag:   normalizeBranchTag(req.BranchTag),
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
	if req.BranchTag != "" || req.BranchTag == "" { // 允许设置为空
		building.BranchTag = normalizeBranchTag(req.BranchTag)
	}
	if req.BuildingName != "" {
		building.BuildingName = strings.TrimSpace(req.BuildingName)
	}

	// 设置默认值
	if building.BuildingName == "" {
		building.BuildingName = "-"
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

	// 2. 构建过滤器（与旧 Handler 逻辑对齐）
	filters := repository.UnitFilters{
		BranchTag:  strings.TrimSpace(req.BranchTag),
		Building:   strings.TrimSpace(req.Building),
		Floor:      strings.TrimSpace(req.Floor),
		AreaTag:    strings.TrimSpace(req.AreaTag),
		UnitNumber: strings.TrimSpace(req.UnitNumber),
		UnitName:   strings.TrimSpace(req.UnitName),
		UnitType:   strings.TrimSpace(req.UnitType),
		Search:     strings.TrimSpace(req.Search),
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
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitName == "" {
		return nil, fmt.Errorf("unit_name is required")
	}
	if req.UnitNumber == "" {
		return nil, fmt.Errorf("unit_number is required")
	}
	if req.UnitType == "" {
		return nil, fmt.Errorf("unit_type is required")
	}
	if req.Timezone == "" {
		return nil, fmt.Errorf("timezone is required")
	}

	// 2. 构建 domain.Unit（与 Repository 逻辑对齐）
	unit := &domain.Unit{
		TenantID:          req.TenantID,
		BranchTag:         normalizeBranchTag(req.BranchTag),
		UnitName:          strings.TrimSpace(req.UnitName),
		Building:          strings.TrimSpace(req.Building),
		Floor:             strings.TrimSpace(req.Floor),
		AreaTag:           normalizeAreaTag(req.AreaTag),
		UnitNumber:        strings.TrimSpace(req.UnitNumber),
		LayoutConfig:      normalizeLayoutConfig(req.LayoutConfig),
		UnitType:          strings.TrimSpace(req.UnitType),
		IsPublicSpace:     req.IsPublicSpace,
		IsMultiPersonRoom: req.IsMultiPersonRoom,
		Timezone:          strings.TrimSpace(req.Timezone),
	}

	// 3. 设置默认值（与 Repository 逻辑对齐）
	if unit.Building == "" {
		unit.Building = "-"
	}
	if unit.Floor == "" {
		unit.Floor = "1F"
	}

	// 4. 业务规则验证（Repository 会再次验证，但 Service 层提前验证更友好）
	branchTagValue := ""
	if unit.BranchTag.Valid {
		branchTagValue = unit.BranchTag.String
	}
	if (branchTagValue == "" || branchTagValue == "-") && (unit.Building == "" || unit.Building == "-") {
		return nil, fmt.Errorf("branch_tag and building cannot both be empty (at least one must be provided)")
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
		BranchTag:         currentUnit.BranchTag,
		UnitName:          currentUnit.UnitName,
		Building:          currentUnit.Building,
		Floor:             currentUnit.Floor,
		AreaTag:           currentUnit.AreaTag,
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
	// - 如果 unit.BranchTag.Valid == true，会更新（即使 String 为空，也会设置为 NULL）
	// - 如果 unit.BranchTag.Valid == false，不会更新（保持原值）
	// 
	// Service 层策略：
	// - 如果请求中提供了非空值，设置 Valid=true 和 String 值
	// - 如果请求中提供了空字符串，设置 Valid=true 和 String=""（Repository 会转换为 NULL）
	// - 如果请求中未提供（空字符串且当前值不存在），保持 Valid=false（不更新）
	// 
	// 注意：由于无法区分"未提供"和"空字符串"，我们采用：
	// - 非空值：更新
	// - 空字符串：如果当前值存在，清除它（设置为 NULL）；否则不更新
	
	// branch_tag: 如果提供了非空值，更新；如果为空字符串，清除（Repository 会处理为 NULL）
	if req.BranchTag != "" {
		unit.BranchTag = normalizeBranchTag(req.BranchTag)
	} else if req.BranchTag == "" && currentUnit.BranchTag.Valid {
		// 请求值为空且当前值存在，清除它（设置为 Valid=true, String=""，Repository 会转换为 NULL）
		unit.BranchTag = sql.NullString{String: "", Valid: true}
	}
	// 如果 req.BranchTag == "" 且当前值不存在，保持 unit.BranchTag.Valid = false（不更新）
	
	if req.UnitName != "" {
		unit.UnitName = strings.TrimSpace(req.UnitName)
	}
	if req.Building != "" {
		unit.Building = strings.TrimSpace(req.Building)
	}
	if req.Floor != "" {
		unit.Floor = strings.TrimSpace(req.Floor)
	}
	
	// area_tag: 类似 branch_tag
	if req.AreaTag != "" {
		unit.AreaTag = normalizeAreaTag(req.AreaTag)
	} else if req.AreaTag == "" && currentUnit.AreaTag.Valid {
		unit.AreaTag = sql.NullString{String: "", Valid: true}
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
		unit.UnitType = strings.TrimSpace(req.UnitType)
	}
	if req.IsPublicSpace != nil {
		unit.IsPublicSpace = *req.IsPublicSpace
	}
	if req.IsMultiPersonRoom != nil {
		unit.IsMultiPersonRoom = *req.IsMultiPersonRoom
	}
	if req.Timezone != "" {
		unit.Timezone = strings.TrimSpace(req.Timezone)
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
	if req.BedType == "" {
		return nil, fmt.Errorf("bed_type is required")
	}

	bed := &domain.Bed{
		TenantID:         req.TenantID,
		RoomID:            req.RoomID,
		BedName:          strings.TrimSpace(req.BedName),
		BedType:          strings.TrimSpace(req.BedType),
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
		BedType:          currentBed.BedType,
		MattressMaterial: currentBed.MattressMaterial,
		MattressThickness: currentBed.MattressThickness,
	}

	// 更新提供的字段
	if req.BedName != "" {
		bed.BedName = strings.TrimSpace(req.BedName)
	}
	if req.BedType != "" {
		bed.BedType = strings.TrimSpace(req.BedType)
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

