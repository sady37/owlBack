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
	ListUnitsWithFullHierarchy(ctx context.Context, req ListUnitsWithFullHierarchyRequest) (*ListUnitsWithFullHierarchyResponse, error)
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
	unitsRepo     repository.UnitsRepository
	branchesRepo  repository.BranchesRepository
	residentsRepo repository.ResidentsRepository
	devicesRepo   repository.DevicesRepository
	db            *sql.DB
	cardSync      *CardSyncService
	logger        *zap.Logger
}

// NewUnitService 创建 UnitService 实例
func NewUnitService(unitsRepo repository.UnitsRepository, branchesRepo repository.BranchesRepository, residentsRepo repository.ResidentsRepository, devicesRepo repository.DevicesRepository, db *sql.DB, cardSync *CardSyncService, logger *zap.Logger) UnitService {
	return &unitService{
		unitsRepo:     unitsRepo,
		branchesRepo:  branchesRepo,
		residentsRepo: residentsRepo,
		devicesRepo:   devicesRepo,
		db:            db,
		cardSync:      cardSync,
		logger:        logger,
	}
}

// ============================================
// 统一的权限验证辅助函数
// 权限验证主要验证两个维度：
//   1. tenant: 确保资源属于请求的 tenant（通过 Repository 层的 WHERE tenant_id = $1 自动验证）
//   2. branch: 确保资源属于请求的 branch（通过检查资源的 branch_id 是否匹配）
// ============================================

// getResourcePermission 查询资源权限配置（Service 层内部方法）
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func (s *unitService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*resourcePermissionCheck, error) {
	const SystemTenantID = "00000000-0000-0000-0000-000000000001"
	var permissionScope string
	err := s.db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID, roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &resourcePermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	// 将 permission_scope 转换为 assigned_only 和 branch_only 标志
	var assignedOnly, branchOnly bool
	switch permissionScope {
	case "A":
		// All (no restriction)
		assignedOnly = false
		branchOnly = false
	case "S":
		// assigned_only
		assignedOnly = true
		branchOnly = false
	case "B":
		// branch_only
		assignedOnly = false
		branchOnly = true
	default:
		// 未知值，返回最严格的权限（安全默认值）
		assignedOnly = true
		branchOnly = true
	}

	return &resourcePermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// getUserBranchID 查询用户所属的 branch_id（Service 层内部方法）
// 从 user_branches 表查询用户关联的第一个院区 ID
// 返回：
//   - branchID: 用户的第一个院区 ID（如果存在）
//   - hasBranches: 用户是否有关联的院区（false 表示可以访问所有院区或 NULL 院区）
func (s *unitService) getUserBranchID(ctx context.Context, tenantID, userID string) (branchID string, hasBranches bool, err error) {
	if tenantID == "" || userID == "" {
		return "", false, nil
	}

	// 查询用户的第一个院区
	var branchIDStr sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT ub.branch_id::text FROM user_branches ub
		 LEFT JOIN branches b ON b.branch_id = ub.branch_id
		 WHERE ub.tenant_id = $1 AND ub.user_id::text = $2
		 ORDER BY COALESCE(b.branch_name, '') ASC
		 LIMIT 1`,
		tenantID, userID,
	).Scan(&branchIDStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户没有关联的院区
			return "", false, nil
		}
		return "", false, fmt.Errorf("failed to query user branch: %w", err)
	}

	if branchIDStr.Valid && branchIDStr.String != "" {
		// 检查用户是否有关联的院区
		var count int
		err2 := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM user_branches 
			 WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, userID,
		).Scan(&count)
		if err2 == nil && count > 0 {
			// 用户有关联的院区
			return branchIDStr.String, true, nil
		}
		if err2 != nil {
			return "", false, fmt.Errorf("failed to query user branch count: %w", err2)
		}
	}

	return "", false, nil // 用户没有关联任何院区
}

// getUserBranchIDs 查询用户所属的所有 branch_id 列表（Service 层内部方法）
// 从 user_branches 表查询用户关联的所有院区 ID
// 如果 user_branches 表返回空且 role=Admin，返回所有 Branch_id
// 返回：
//   - branchIDs: 用户所属的 branch_id 列表（可能为空）
//   - hasBranches: 用户是否有关联的院区（false 表示可以访问所有院区或 NULL 院区）
func (s *unitService) getUserBranchIDs(ctx context.Context, tenantID, userID string) (branchIDs []string, hasBranches bool, err error) {
	if tenantID == "" || userID == "" {
		return nil, false, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT branch_id::text FROM user_branches 
		 WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, userID,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to query user branches: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var branchID string
		if err := rows.Scan(&branchID); err != nil {
			return nil, false, fmt.Errorf("failed to scan branch_id: %w", err)
		}
		ids = append(ids, branchID)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("failed to iterate user branches: %w", err)
	}

	// 如果 user_branches 表返回空，检查是否是 Admin 的 *(ALL) 情况
	if len(ids) == 0 {
		return s.handleAdminAllBranches(ctx, tenantID, userID)
	}

	return ids, true, nil
}

// handleAdminAllBranches 处理 Admin 的 *(ALL) 情况（user_branches 表为空时返回所有 branch）
// 返回所有 branch_id 列表
func (s *unitService) handleAdminAllBranches(ctx context.Context, tenantID, userID string) (branchIDs []string, hasBranches bool, err error) {
	// 获取用户信息，检查是否是 Admin
	var userRole sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, userID,
	).Scan(&userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在，返回空列表
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get user role: %w", err)
	}

	// 如果是 Admin 角色，返回所有 branch_id
	if userRole.Valid && strings.EqualFold(userRole.String, "Admin") {
		allBranches, _, err := s.branchesRepo.ListBranches(ctx, tenantID, "", 1, 1000)
		if err != nil {
			s.logger.Warn("Failed to list all branches for Admin user", zap.String("user_id", userID), zap.Error(err))
			// 如果查询失败，返回空列表
			return nil, false, nil
		}

		// 构建所有 branch_id 列表（直接使用命名返回值 branchIDs）
		branchIDs = make([]string, 0, len(allBranches))
		for _, branch := range allBranches {
			branchIDs = append(branchIDs, branch.BranchID)
		}

		return branchIDs, true, nil
	}

	// 非 Admin 角色，返回空列表（未分配状态）
	return nil, false, nil
}

// getBranchIDForPermission 统一获取 branch_id 用于权限验证
// 优先级：
//  1. 如果请求中提供了 BranchID，则验证该 BranchID 是否在用户绑定的 branch_id 列表中
//  2. 如果未提供，则从 user_branches 表查询用户的主院区
//
// 返回：
//   - branchID: 用于权限验证的 branch_id（可能为空，表示可以访问所有院区或 NULL 院区）
//   - hasBranches: 用户是否有关联的院区
func (s *unitService) getBranchIDForPermission(ctx context.Context, tenantID, userID, providedBranchID string) (branchID string, hasBranches bool, err error) {
	// 如果提供了 BranchID，需要验证用户是否有权限访问该 branch（即该 branch_id 是否在用户的 branch_id 列表中）
	if providedBranchID != "" {
		// 查询用户的所有 branch_id 列表
		userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, tenantID, userID)
		if err != nil {
			return "", false, fmt.Errorf("failed to verify user branch permission: %w", err)
		}
		// 如果用户有关联的院区，验证提供的 branch_id 是否在列表中
		if hasBranches {
			found := false
			for _, bid := range userBranchIDs {
				if bid == providedBranchID {
					found = true
					break
				}
			}
			if !found {
				return "", false, fmt.Errorf("user does not have permission to access branch %s (user's branches: %v)", providedBranchID, userBranchIDs)
			}
		}
		// 用户有权限，使用提供的 branch_id
		return providedBranchID, hasBranches, nil
	}

	// 未提供 BranchID，从 user_branches 表查询用户的主院区
	return s.getUserBranchID(ctx, tenantID, userID)
}

// verifyUnitPermission 验证 unit 的权限（tenant + branch）
// 参数：
//   - tenantID: 必填，用于验证 unit 是否属于该 tenant
//   - unitID: 必填，要验证的 unit ID
//   - branchID: 可选，如果提供则验证 unit 是否属于该 branch；如果为空字符串，跳过 branch 验证（用于权限 scope 为 'A' 的情况）
//
// 返回：
//   - *domain.Unit: 验证通过后返回 unit 对象
//   - error: 如果验证失败（tenant 不匹配、unit 不存在、branch 不匹配等）
func (s *unitService) verifyUnitPermission(ctx context.Context, tenantID, unitID, branchID string) (*domain.Unit, error) {
	// 1. 获取 unit 信息（Repository 层会自动验证 tenant_id）
	unit, err := s.unitsRepo.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("unit not found: unit_id=%s (tenant validation failed or unit does not exist)", unitID)
		}
		s.logger.Error("verifyUnitPermission: failed to get unit",
			zap.String("tenant_id", tenantID),
			zap.String("unit_id", unitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 2. 验证 branch 权限
	// 如果提供了 branchID，验证 unit 是否属于该 branch
	// 如果未提供 branchID（空字符串），表示权限 scope 为 'A'（无限制），跳过 branch 验证
	if branchID != "" {
		if !unit.BranchID.Valid || unit.BranchID.String != branchID {
			// 查询 branch_name（优先使用 branch_name）
			var branchName sql.NullString
			if unit.BranchID.Valid {
				err := s.db.QueryRowContext(ctx,
					`SELECT branch_name FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
					tenantID, unit.BranchID.String,
				).Scan(&branchName)
				if err != nil && err != sql.ErrNoRows {
					s.logger.Warn("verifyUnitPermission: failed to get branch_name",
						zap.String("branch_id", unit.BranchID.String),
						zap.Error(err),
					)
				}
			}

			// 查询请求的 branch_name
			var requestedBranchName sql.NullString
			err := s.db.QueryRowContext(ctx,
				`SELECT branch_name FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
				tenantID, branchID,
			).Scan(&requestedBranchName)
			if err != nil && err != sql.ErrNoRows {
				s.logger.Warn("verifyUnitPermission: failed to get requested branch_name",
					zap.String("branch_id", branchID),
					zap.Error(err),
				)
			}

			s.logger.Warn("verifyUnitPermission: branch_id mismatch",
				zap.String("tenant_id", tenantID),
				zap.String("unit_id", unitID),
				zap.String("requested_branch_id", branchID),
				zap.String("requested_branch_name", func() string {
					if requestedBranchName.Valid {
						return requestedBranchName.String
					}
					return ""
				}()),
				zap.String("unit_branch_id", func() string {
					if unit.BranchID.Valid {
						return unit.BranchID.String
					}
					return "NULL"
				}()),
				zap.String("unit_branch_name", func() string {
					if branchName.Valid {
						return branchName.String
					}
					return ""
				}()),
			)

			// 优先使用 branch_name，如果不存在则使用 branch_id
			var errorMsg string
			if requestedBranchName.Valid && requestedBranchName.String != "" {
				// 优先使用 branch_name
				errorMsg = fmt.Sprintf("permission denied: unit does not belong to branch_name=%s (branch_id=%s)", requestedBranchName.String, branchID)
			} else {
				// 如果没有 branch_name，使用 branch_id，并尽可能显示 unit 的 branch_name
				if branchName.Valid && branchName.String != "" {
					errorMsg = fmt.Sprintf("permission denied: unit does not belong to branch_id=%s (unit belongs to branch_name=%s)", branchID, branchName.String)
				} else {
					errorMsg = fmt.Sprintf("permission denied: unit does not belong to branch_id=%s", branchID)
				}
			}
			return nil, fmt.Errorf(errorMsg)
		}
	}

	return unit, nil
}

// verifyRoomPermission 验证 room 的权限（tenant + branch）
// 通过 room 所属的 unit 来验证权限
// 参数：
//   - tenantID: 必填，用于验证 room 是否属于该 tenant
//   - roomID: 必填，要验证的 room ID
//   - branchID: 可选，如果提供则验证 room 所属的 unit 是否属于该 branch；如果为空字符串，跳过 branch 验证
//
// 返回：
//   - *domain.Room: 验证通过后返回 room 对象
//   - *domain.Unit: 验证通过后返回 room 所属的 unit 对象
//   - error: 如果验证失败（tenant 不匹配、room 不存在、branch 不匹配等）
func (s *unitService) verifyRoomPermission(ctx context.Context, tenantID, roomID, branchID string) (*domain.Room, *domain.Unit, error) {
	// 1. 获取 room 信息（Repository 层会自动验证 tenant_id）
	room, err := s.unitsRepo.GetRoom(ctx, tenantID, roomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, nil, fmt.Errorf("room not found: room_id=%s (tenant validation failed or room does not exist)", roomID)
		}
		s.logger.Error("verifyRoomPermission: failed to get room",
			zap.String("tenant_id", tenantID),
			zap.String("room_id", roomID),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to get room: %w", err)
	}

	// 2. 通过 unit 验证权限（包括 tenant 和 branch）
	unit, err := s.verifyUnitPermission(ctx, tenantID, room.UnitID, branchID)
	if err != nil {
		return nil, nil, err
	}

	return room, unit, nil
}

// verifyBedPermission 验证 bed 的权限（tenant + branch）
// 通过 bed -> room -> unit 的层级关系来验证权限
// 参数：
//   - tenantID: 必填，用于验证 bed 是否属于该 tenant
//   - bedID: 必填，要验证的 bed ID
//   - branchID: 可选，如果提供则验证 bed 所属的 room 所属的 unit 是否属于该 branch；如果为空字符串，跳过 branch 验证
//
// 返回：
//   - *domain.Bed: 验证通过后返回 bed 对象
//   - *domain.Room: 验证通过后返回 bed 所属的 room 对象
//   - *domain.Unit: 验证通过后返回 room 所属的 unit 对象
//   - error: 如果验证失败（tenant 不匹配、bed 不存在、branch 不匹配等）
func (s *unitService) verifyBedPermission(ctx context.Context, tenantID, bedID, branchID string) (*domain.Bed, *domain.Room, *domain.Unit, error) {
	// 1. 获取 bed 信息（Repository 层会自动验证 tenant_id）
	bed, err := s.unitsRepo.GetBed(ctx, tenantID, bedID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, nil, nil, fmt.Errorf("bed not found: bed_id=%s (tenant validation failed or bed does not exist)", bedID)
		}
		s.logger.Error("verifyBedPermission: failed to get bed",
			zap.String("tenant_id", tenantID),
			zap.String("bed_id", bedID),
			zap.Error(err),
		)
		return nil, nil, nil, fmt.Errorf("failed to get bed: %w", err)
	}

	// 2. 通过 room 验证权限（包括 tenant 和 branch）
	room, unit, err := s.verifyRoomPermission(ctx, tenantID, bed.RoomID, branchID)
	if err != nil {
		return nil, nil, nil, err
	}

	return bed, room, unit, nil
}

// ============================================
// Building 相关请求/响应结构
// ============================================

type ListBuildingsRequest struct {
	TenantID   string // 必填
	BranchID   string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName string // 可选（向后兼容，如果 BranchID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
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
	Units    []BuildingUnit   `json:"units"` // 该 building 下的 units 列表
}

// BuildingUnit 用于 GetBuilding 返回的 unit 信息（简化版，只包含必要字段）
type BuildingUnit struct {
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name"`
	Floor    string `json:"floor,omitempty"` // 可选，可能为 NULL
}

type CreateBuildingRequest struct {
	TenantID     string // 必填
	BranchID     string // 必填（不再支持 BranchName）
	BuildingName string // 必填（不能为空或空格）
}

type CreateBuildingResponse struct {
	BuildingID string `json:"building_id"`
}

type UpdateBuildingRequest struct {
	TenantID     string // 必填
	BuildingID   string // 必填
	BuildingName string // 必填（只能修改 building_name，不能修改 branch_id）
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
	BranchID   *string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName *string // 可选（向后兼容，如果 BranchID 未提供则使用此字段）
	BuildingID *string // 可选（优先使用，UUID 类型，如果提供则忽略 Building）
	Building   *string // 可选（向后兼容，如果 BuildingID 未提供则使用此字段，通过 building_name 过滤）
	Floor      *string // 可选（nil 表示未提供）
	AreaName   *string // 可选（nil 表示未提供）
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

// ListUnitsWithFullHierarchyRequest 查询 Units 及其完整层级结构的请求
type ListUnitsWithFullHierarchyRequest struct {
	TenantID      string  // 必填
	CurrentUserID string  // 当前用户ID（用于权限过滤，从 user_branches 表查询用户的 branch_id）
	BranchID      *string // 可选（优先使用，如果提供则忽略从 user_branches 查询的结果）
	BranchName    *string // 可选（向后兼容，如果 BranchID 未提供则使用此字段）
	BuildingID    *string // 可选（优先使用，UUID 类型，如果提供则忽略 Building）
	Building      *string // 可选（向后兼容，如果 BuildingID 未提供则使用此字段）
	Floor         *string // 可选
	UnitType      *string // 可选
	Search        *string // 可选（模糊搜索 unit_name）
	// 注意：不分页，返回所有匹配的 units（因为前端需要完整层级结构）
}

// ListUnitsWithFullHierarchyResponse 查询 Units 及其完整层级结构的响应
type ListUnitsWithFullHierarchyResponse struct {
	Items []*UnitWithFullHierarchy `json:"items"`
	Total int                      `json:"total"`
}

// UnitWithFullHierarchy 包含完整的层级结构
type UnitWithFullHierarchy struct {
	*domain.Unit                           // Unit 基本信息
	Rooms        []*RoomWithBedsAndDevices `json:"rooms"`
}

// RoomWithBedsAndDevices 房间及其床位和设备
type RoomWithBedsAndDevices struct {
	*domain.Room                   // Room 基本信息
	Beds         []*BedWithDevices `json:"beds"`
	DeviceIDs    []string          `json:"device_ids"`   // 绑定到 room 的 device IDs（用于前端选中并向后端传递）
	DeviceNames  []string          `json:"device_names"` // 绑定到 room 的 device names（用于前端显示）
}

// BedWithDevices 床位及其设备
type BedWithDevices struct {
	*domain.Bed          // Bed 基本信息
	DeviceIDs   []string `json:"device_ids"`   // 绑定到 bed 的 device IDs（用于前端选中并向后端传递）
	DeviceNames []string `json:"device_names"` // 绑定到 bed 的 device names（用于前端显示）
}

type GetUnitRequest struct {
	TenantID string // 必填
	UnitID   string // 必填
}

type GetUnitResponse struct {
	Unit *domain.Unit `json:"unit"`
}

type CreateUnitRequest struct {
	TenantID      string // 必填
	BranchID      string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName    string // 可选（向后兼容，如果 BranchID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	UnitName      string // 必填
	BuildingID    string // 可选（优先使用，如果提供则忽略 BuildingName）
	BuildingName  string // 可选（向后兼容，如果 BuildingID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	Floor         string // 可选（默认 "1F"）
	AreaName      string // 可选
	UnitNumber    string // 必填
	LayoutConfig  string // 可选（JSON 字符串）
	UnitType      string // 必填
	IsPublicSpace bool   // 可选（默认 false）
	IsSharedUnit  bool   // 可选（默认 false）- 统一使用 IsSharedUnit，不再使用 IsMultiPersonRoom
	Timezone      string // 必填
}

type CreateUnitResponse struct {
	UnitID string `json:"unit_id"`
}

type UpdateUnitRequest struct {
	TenantID      string // 必填
	UnitID        string // 必填
	BranchID      string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName    string // 可选（向后兼容，如果 BranchID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	UnitName      string // 可选
	BuildingID    string // 可选（优先使用，如果提供则忽略 BuildingName）
	BuildingName  string // 可选（向后兼容，如果 BuildingID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	Floor         string // 可选
	AreaName      string // 可选
	UnitNumber    string // 可选
	LayoutConfig  string // 可选（JSON 字符串）
	UnitType      string // 可选
	IsPublicSpace *bool  // 可选（指针类型，nil 表示不更新）
	IsSharedUnit  *bool  // 可选（指针类型，nil 表示不更新）- 统一使用 IsSharedUnit，不再使用 IsMultiPersonRoom
	Timezone      string // 可选
}

type UpdateUnitResponse struct {
	Success bool `json:"success"`
}

type DeleteUnitRequest struct {
	TenantID      string // 必填
	UnitID        string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
}

type DeleteUnitResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Room 相关请求/响应结构
// ============================================

type ListRoomsRequest struct {
	TenantID      string // 必填
	UnitID        string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
	Search        string // 可选（按 room_name 模糊搜索）
}

type ListRoomsResponse struct {
	Items []*domain.Room `json:"items"`
}

type ListRoomsWithBedsRequest struct {
	TenantID      string // 必填
	UnitID        string // 必填
	CurrentUserID string // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string // 可选（保留字段，不再使用）
	Search        string // 可选（按 room_name 模糊搜索）
}

type ListRoomsWithBedsResponse struct {
	Items []*repository.RoomWithBeds `json:"items"`
}

type GetRoomRequest struct {
	TenantID      string // 必填
	RoomID        string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
}

type GetRoomResponse struct {
	Room *domain.Room `json:"room"`
}

type CreateRoomRequest struct {
	TenantID      string // 必填
	UnitID        string // 必填
	CurrentUserID string // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string // 可选（保留字段，不再使用）
	RoomName      string // 必填
	LayoutConfig  string // 可选（JSON 字符串）
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type UpdateRoomRequest struct {
	TenantID      string // 必填
	RoomID        string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
	RoomName      string // 可选
	LayoutConfig  string // 可选（JSON 字符串）
}

type UpdateRoomResponse struct {
	Success bool `json:"success"`
}

type DeleteRoomRequest struct {
	TenantID      string // 必填
	RoomID        string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
}

type DeleteRoomResponse struct {
	Success bool `json:"success"`
}

// ============================================
// Bed 相关请求/响应结构
// ============================================

type ListBedsRequest struct {
	TenantID      string // 必填
	RoomID        string // 必填
	CurrentUserID string // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string // 可选（保留字段，不再使用）
	Search        string // 可选（按 bed_name 模糊搜索）
}

type ListBedsResponse struct {
	Items []*domain.Bed `json:"items"`
}

type GetBedRequest struct {
	TenantID        string // 必填
	BedID           string // 必填
	CurrentUserID   string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	CurrentUserRole string // 可选（用于权限验证，检查权限 scope，如果是 'A' 则跳过 branch 验证）
	BranchID        string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
}

type GetBedResponse struct {
	Bed *domain.Bed `json:"bed"`
}

type CreateBedRequest struct {
	TenantID      string // 必填
	RoomID        string // 必填
	CurrentUserID string // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string // 可选（保留字段，不再使用）
	BedName       string // 必填
	// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	MattressMaterial  string // 可选
	MattressThickness string // 可选
}

type CreateBedResponse struct {
	BedID string `json:"bed_id"`
}

type UpdateBedRequest struct {
	TenantID      string // 必填
	BedID         string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
	BedName       string // 可选
	// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
	MattressMaterial  string // 可选
	MattressThickness string // 可选
}

type UpdateBedResponse struct {
	Success bool `json:"success"`
}

type DeleteBedRequest struct {
	TenantID      string // 必填
	BedID         string // 必填
	CurrentUserID string // 必填（用于权限验证，从 user_branches 表查询用户的 branch_id）
	BranchID      string // 可选（如果提供则验证用户是否有权限访问该 branch，否则从 user_branches 查询）
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

	// 处理 branch_id：优先使用 BranchID，如果没有则通过 BranchName 查找
	var branchID sql.NullString
	var branchNameForQuery string

	if req.BranchID != "" {
		// 优先使用 BranchID
		branchID = sql.NullString{String: req.BranchID, Valid: true}
		// 如果提供了 BranchID，不需要 branch_name（Repository 层会直接使用 branch_id 过滤）
		branchNameForQuery = ""
	} else if req.BranchName != "" {
		// 向后兼容：使用 BranchName
		// Vue 层已经 trim 空格，这里处理空字符串自动转换为 "-"
		branchNameTrimmed := strings.TrimSpace(req.BranchName)
		if branchNameTrimmed == "" {
			branchNameTrimmed = "-"
		}
		branchNameForQuery = branchNameTrimmed
	} else {
		// 都没有提供：查询整个 tenant 的所有 buildings
		branchNameForQuery = ""
	}

	// 如果提供了 BranchID，Repository 层需要支持通过 branch_id 过滤
	// 否则使用 branch_name 过滤（向后兼容）
	items, err := s.unitsRepo.ListBuildings(ctx, req.TenantID, branchID, branchNameForQuery)
	if err != nil {
		s.logger.Error("ListBuildings failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_id", req.BranchID),
			zap.String("branch_name", req.BranchName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list buildings: %w", err)
	}

	return &ListBuildingsResponse{
		Items: items,
	}, nil
}

// GetBuilding 获取单个楼栋详情（包含关联的 units 列表）
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

	// 获取该 building 下的所有 units（只包含 unit_id, unit_name, floor）
	unitsData, err := s.unitsRepo.GetBuildingUnits(ctx, req.TenantID, req.BuildingID)
	if err != nil {
		s.logger.Warn("GetBuildingUnits failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_id", req.BuildingID),
			zap.Error(err),
		)
		// 如果查询 units 失败，返回空列表而不是错误（building 信息仍然有效）
		unitsData = []repository.BuildingUnitInfo{}
	}

	// 转换为 BuildingUnit 列表
	units := make([]BuildingUnit, 0, len(unitsData))
	for _, u := range unitsData {
		unit := BuildingUnit{
			UnitID:   u.UnitID,
			UnitName: u.UnitName,
		}
		if u.Floor.Valid {
			unit.Floor = u.Floor.String
		}
		units = append(units, unit)
	}

	return &GetBuildingResponse{
		Building: building,
		Units:    units,
	}, nil
}

// CreateBuilding 创建楼栋
func (s *unitService) CreateBuilding(ctx context.Context, req CreateBuildingRequest) (*CreateBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// branch_id 必须提供，不能为空（业务规则：一家机构必然有一个分支或总部）
	if req.BranchID == "" {
		return nil, fmt.Errorf("branch_id is required and cannot be empty")
	}

	// building_name 必须提供，不能为空或空格
	buildingNameTrimmed := strings.TrimSpace(req.BuildingName)
	if buildingNameTrimmed == "" {
		return nil, fmt.Errorf("building_name is required and cannot be empty or whitespace")
	}

	// 检查是否已存在相同的 building（通过 branch_id + building_name 小写比较）
	// 如果已存在，返回现有的 building_id，不创建新记录
	existingBuildingID, err := s.unitsRepo.FindBuildingByBranchAndName(ctx, req.TenantID, req.BranchID, buildingNameTrimmed)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		s.logger.Warn("FindBuildingByBranchAndName failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_id", req.BranchID),
			zap.String("building_name", buildingNameTrimmed),
			zap.Error(err),
		)
		// 如果查询失败，继续创建流程（可能是数据库错误，不影响创建）
	} else if err == nil && existingBuildingID != "" {
		// 已存在相同的 building，返回现有的 building_id
		return &CreateBuildingResponse{
			BuildingID: existingBuildingID,
		}, nil
	}

	building := &domain.Building{
		TenantID:     req.TenantID,
		BranchID:     sql.NullString{String: req.BranchID, Valid: true},
		BuildingName: buildingNameTrimmed,
	}

	buildingID, err := s.unitsRepo.CreateBuilding(ctx, req.TenantID, building)
	if err != nil {
		s.logger.Error("CreateBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_id", req.BranchID),
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
// 注意：只能修改 building_name，不能修改 branch_id
func (s *unitService) UpdateBuilding(ctx context.Context, req UpdateBuildingRequest) (*UpdateBuildingResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BuildingID == "" {
		return nil, fmt.Errorf("building_id is required")
	}
	if req.BuildingName == "" {
		return nil, fmt.Errorf("building_name is required and cannot be empty")
	}

	// 获取当前 building
	currentBuilding, err := s.unitsRepo.GetBuilding(ctx, req.TenantID, req.BuildingID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	// 验证：branch_id 必须存在（不能为 NULL）
	if !currentBuilding.BranchID.Valid || currentBuilding.BranchID.String == "" {
		return nil, fmt.Errorf("building has no branch_id, cannot update")
	}

	// Trim 首尾空格
	buildingNameTrimmed := strings.TrimSpace(req.BuildingName)
	if buildingNameTrimmed == "" {
		return nil, fmt.Errorf("building_name is required and cannot be empty or whitespace")
	}

	// 验证：新的 building_name 不能与同一 branch_id 下的 branch_name 重名（忽略大小写）
	// 获取 branch 信息
	branch, err := s.branchesRepo.GetBranch(ctx, req.TenantID, currentBuilding.BranchID.String)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("branch not found: branch_id=%s", currentBuilding.BranchID.String)
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	// 检查 building_name（小写）是否与 branch_name（小写）相同
	if strings.EqualFold(buildingNameTrimmed, branch.BranchName) {
		return nil, fmt.Errorf("building_name cannot be the same as branch_name (case-insensitive): building_name=%s, branch_name=%s", buildingNameTrimmed, branch.BranchName)
	}

	// 构建更新后的 building（只更新 building_name，保持 branch_id 不变）
	building := &domain.Building{
		BuildingID:   req.BuildingID,
		TenantID:     req.TenantID,
		BranchID:     currentBuilding.BranchID, // 保持原有 branch_id，不允许修改
		BuildingName: buildingNameTrimmed,
	}

	err = s.unitsRepo.UpdateBuilding(ctx, req.TenantID, req.BuildingID, building)
	if err != nil {
		s.logger.Error("UpdateBuilding failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("building_id", req.BuildingID),
			zap.String("building_name", req.BuildingName),
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
	// 优先使用 BranchID，如果提供则忽略 BranchName
	// 空字符串视为 null（nil → ""，用于 Repository 层转换为 IS NULL）
	branchID := stringValueOrEmpty(req.BranchID)
	branchName := ""
	if branchID == "" {
		// 如果 BranchID 未提供，则使用 BranchName（向后兼容）
		branchName = stringValueOrEmpty(req.BranchName)
	}

	// 优先使用 BuildingID，如果提供则忽略 Building
	buildingID := stringValueOrEmpty(req.BuildingID)
	building := ""
	if buildingID == "" {
		// 如果 BuildingID 未提供，则使用 Building（向后兼容，通过 building_name 过滤）
		building = stringValueOrEmpty(req.Building)
	}

	filters := repository.UnitFilters{
		BranchID:   branchID,
		BranchName: branchName,
		BuildingID: buildingID,
		Building:   building,
		Floor:      stringValueOrEmpty(req.Floor),
		AreaName:   stringValueOrEmpty(req.AreaName),
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

// ListUnitsWithFullHierarchy 查询 Units 及其完整的层级结构（Rooms, Beds, Devices）
func (s *unitService) ListUnitsWithFullHierarchy(ctx context.Context, req ListUnitsWithFullHierarchyRequest) (*ListUnitsWithFullHierarchyResponse, error) {
	// 输出输入参数到标准输出
	// fmt.Printf("=== ListUnitsWithFullHierarchy INPUT ===\n")
	// fmt.Printf("TenantID: %s\n", req.TenantID)
	// fmt.Printf("CurrentUserID: %s\n", req.CurrentUserID)
	// if req.BranchID != nil {
	// 	fmt.Printf("BranchID: %s\n", *req.BranchID)
	// }
	// if req.BranchName != nil {
	// 	fmt.Printf("BranchName: %s\n", *req.BranchName)
	// }
	// if req.BuildingID != nil {
	// 	fmt.Printf("BuildingID: %s\n", *req.BuildingID)
	// }
	// if req.Building != nil {
	// 	fmt.Printf("Building: %s\n", *req.Building)
	// }
	// if req.Floor != nil {
	// 	fmt.Printf("Floor: %s\n", *req.Floor)
	// }
	// if req.UnitType != nil {
	// 	fmt.Printf("UnitType: %s\n", *req.UnitType)
	// }
	// if req.Search != nil {
	// 	fmt.Printf("Search: %s\n", *req.Search)
	// }
	// fmt.Printf("========================================\n\n")

	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 2. 获取 branch_id：优先使用请求中的 BranchID，如果没有则从 user_branches 表查询用户的所有 branch_id
	branchID := stringValueOrEmpty(req.BranchID)
	branchIDs := []string{}
	branchName := ""

	if branchID == "" && req.CurrentUserID != "" {
		// 从 user_branches 表查询用户的所有 branch_id
		userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
		if err != nil {
			s.logger.Error("ListUnitsWithFullHierarchy: getUserBranchIDs failed",
				zap.String("tenant_id", req.TenantID),
				zap.String("user_id", req.CurrentUserID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to get user branches: %w", err)
		}
		if hasBranches && len(userBranchIDs) > 0 {
			// 如果用户有多个 branch_id，使用 BranchIDs 数组进行 IN 查询
			// 如果用户只有一个 branch_id，使用 BranchID 过滤（向后兼容）
			if len(userBranchIDs) == 1 {
				branchID = userBranchIDs[0]
			} else {
				branchIDs = userBranchIDs
			}
		} else {
			// 用户没有关联任何院区，匹配 branch_id IS NULL
			// branchID 和 branchName 都为空，Repository 层会匹配 branch_id IS NULL
		}
	} else if branchID == "" {
		// 如果请求中也没有提供 branch_id，使用 branch_name（向后兼容）
		branchName = stringValueOrEmpty(req.BranchName)
	}

	buildingID := stringValueOrEmpty(req.BuildingID)
	building := ""
	if buildingID == "" {
		building = stringValueOrEmpty(req.Building)
	}

	filters := repository.UnitFilters{
		BranchID:   branchID,
		BranchIDs:  branchIDs, // 支持多个 branch_id（IN 查询）
		BranchName: branchName,
		BuildingID: buildingID,
		Building:   building,
		Floor:      stringValueOrEmpty(req.Floor),
		UnitType:   stringValueOrEmpty(req.UnitType),
		Search:     stringValueOrEmpty(req.Search),
	}

	// Debug: 输出过滤器信息
	// fmt.Printf("=== ListUnitsWithFullHierarchy: Filters ===\n")
	// fmt.Printf("BranchID: '%s'\n", filters.BranchID)
	// fmt.Printf("BranchName: '%s'\n", filters.BranchName)
	// fmt.Printf("BuildingID: '%s'\n", filters.BuildingID)
	// fmt.Printf("Building: '%s'\n", filters.Building)
	// fmt.Printf("Floor: '%s'\n", filters.Floor)
	// fmt.Printf("UnitType: '%s'\n", filters.UnitType)
	// fmt.Printf("Search: '%s'\n", filters.Search)
	// fmt.Printf("==========================================\n\n")

	// 3. 查询 units（不分页，返回所有匹配的 units）
	units, total, err := s.unitsRepo.ListUnits(ctx, req.TenantID, filters, 1, 10000)
	if err != nil {
		s.logger.Error("ListUnitsWithFullHierarchy: ListUnits failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	// 输出查询到的 units
	// fmt.Printf("=== ListUnitsWithFullHierarchy: Units Found ===\n")
	// fmt.Printf("Total: %d\n", total)
	// fmt.Printf("Units Count: %d\n", len(units))
	// for i, unit := range units {
	// 	fmt.Printf("Unit[%d]: unit_id=%s, unit_name=%s, building_name=%v, floor=%v, unit_type=%s\n",
	// 		i, unit.UnitID, unit.UnitName, unit.BuildingName, unit.Floor, unit.UnitType)
	// }
	// fmt.Printf("==============================================\n\n")

	if len(units) == 0 {
		return &ListUnitsWithFullHierarchyResponse{
			Items: []*UnitWithFullHierarchy{},
			Total: total,
		}, nil
	}

	// 4. 批量查询所有 rooms
	unitIDs := make([]string, len(units))
	for i, unit := range units {
		unitIDs[i] = unit.UnitID
	}
	allRooms, err := s.unitsRepo.ListRoomsByUnitIDs(ctx, req.TenantID, unitIDs)
	if err != nil {
		s.logger.Error("ListUnitsWithFullHierarchy: ListRoomsByUnitIDs failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	// 输出查询到的 rooms
	// fmt.Printf("=== ListUnitsWithFullHierarchy: Rooms Found ===\n")
	// fmt.Printf("Rooms Count: %d\n", len(allRooms))
	// for i, room := range allRooms {
	// 	fmt.Printf("Room[%d]: room_id=%s, room_name=%s, unit_id=%s\n",
	// 		i, room.RoomID, room.RoomName, room.UnitID)
	// }
	// fmt.Printf("==============================================\n\n")

	// 5. 批量查询所有 beds
	roomIDs := make([]string, len(allRooms))
	for i, room := range allRooms {
		roomIDs[i] = room.RoomID
	}
	allBeds, err := s.unitsRepo.ListBedsByRoomIDs(ctx, req.TenantID, roomIDs)
	if err != nil {
		s.logger.Error("ListUnitsWithFullHierarchy: ListBedsByRoomIDs failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list beds: %w", err)
	}

	// 输出查询到的 beds
	// fmt.Printf("=== ListUnitsWithFullHierarchy: Beds Found ===\n")
	// fmt.Printf("Beds Count: %d\n", len(allBeds))
	// for i, bed := range allBeds {
	// 	fmt.Printf("Bed[%d]: bed_id=%s, bed_name=%s, room_id=%s\n",
	// 		i, bed.BedID, bed.BedName, bed.RoomID)
	// }
	// fmt.Printf("=============================================\n\n")

	// 6. 批量查询 device IDs 和 names
	bedIDs := make([]string, len(allBeds))
	for i, bed := range allBeds {
		bedIDs[i] = bed.BedID
	}
	roomDevices, err := s.devicesRepo.GetDevicesByRoomIDs(ctx, req.TenantID, roomIDs)
	if err != nil {
		s.logger.Error("ListUnitsWithFullHierarchy: GetDevicesByRoomIDs failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get devices by room IDs: %w", err)
	}
	bedDevices, err := s.devicesRepo.GetDevicesByBedIDs(ctx, req.TenantID, bedIDs)
	if err != nil {
		s.logger.Error("ListUnitsWithFullHierarchy: GetDevicesByBedIDs failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get devices by bed IDs: %w", err)
	}

	// 输出查询到的 devices
	// fmt.Printf("=== ListUnitsWithFullHierarchy: Devices Found ===\n")
	// fmt.Printf("Room Devices Count: %d\n", len(roomDevices))
	// for roomID, devices := range roomDevices {
	// 	fmt.Printf("Room[%s] Devices: %d\n", roomID, len(devices))
	// 	for j, device := range devices {
	// 		fmt.Printf("  Device[%d]: id=%s, name=%s\n", j, device.ID, device.Name)
	// 	}
	// }
	// fmt.Printf("Bed Devices Count: %d\n", len(bedDevices))
	// for bedID, devices := range bedDevices {
	// 	fmt.Printf("Bed[%s] Devices: %d\n", bedID, len(devices))
	// 	for j, device := range devices {
	// 		fmt.Printf("  Device[%d]: id=%s, name=%s\n", j, device.ID, device.Name)
	// 	}
	// }
	// fmt.Printf("================================================\n\n")

	// 7. 组装数据
	result := make([]*UnitWithFullHierarchy, 0, len(units))
	for _, unit := range units {
		// 过滤该 unit 的 rooms
		rooms := make([]*domain.Room, 0)
		for _, room := range allRooms {
			if room.UnitID == unit.UnitID {
				rooms = append(rooms, room)
			}
		}

		roomsWithBeds := make([]*RoomWithBedsAndDevices, 0, len(rooms))
		for _, room := range rooms {
			// 过滤该 room 的 beds
			beds := make([]*domain.Bed, 0)
			for _, bed := range allBeds {
				if bed.RoomID == room.RoomID {
					beds = append(beds, bed)
				}
			}

			// 构建 beds with devices
			bedsWithDevices := make([]*BedWithDevices, 0, len(beds))
			for _, bed := range beds {
				bedDeviceList := bedDevices[bed.BedID]
				deviceIDs := make([]string, 0, len(bedDeviceList))
				deviceNames := make([]string, 0, len(bedDeviceList))
				for _, device := range bedDeviceList {
					deviceIDs = append(deviceIDs, device.ID)
					deviceNames = append(deviceNames, device.Name)
				}
				bedsWithDevices = append(bedsWithDevices, &BedWithDevices{
					Bed:         bed,
					DeviceIDs:   deviceIDs,
					DeviceNames: deviceNames,
				})
			}

			// 获取该 room 的 devices
			roomDeviceList := roomDevices[room.RoomID]
			roomDeviceIDs := make([]string, 0, len(roomDeviceList))
			roomDeviceNames := make([]string, 0, len(roomDeviceList))
			for _, device := range roomDeviceList {
				roomDeviceIDs = append(roomDeviceIDs, device.ID)
				roomDeviceNames = append(roomDeviceNames, device.Name)
			}

			roomsWithBeds = append(roomsWithBeds, &RoomWithBedsAndDevices{
				Room:        room,
				Beds:        bedsWithDevices,
				DeviceIDs:   roomDeviceIDs,
				DeviceNames: roomDeviceNames,
			})
		}

		result = append(result, &UnitWithFullHierarchy{
			Unit:  unit,
			Rooms: roomsWithBeds,
		})
	}

	// 输出最终结果到标准输出
	// fmt.Printf("=== ListUnitsWithFullHierarchy OUTPUT ===\n")
	// fmt.Printf("Total: %d\n", total)
	// fmt.Printf("Result Items Count: %d\n", len(result))
	// for i, item := range result {
	// 	fmt.Printf("Result[%d]: unit_id=%s, unit_name=%s, building_name=%v, floor=%v, rooms_count=%d\n",
	// 		i, item.Unit.UnitID, item.Unit.UnitName, item.Unit.BuildingName, item.Unit.Floor, len(item.Rooms))
	// 	for j, room := range item.Rooms {
	// 		fmt.Printf("  Room[%d]: room_id=%s, room_name=%s, beds_count=%d, device_ids_count=%d\n",
	// 			j, room.Room.RoomID, room.Room.RoomName, len(room.Beds), len(room.DeviceIDs))
	// 	}
	// }
	// fmt.Printf("==========================================\n\n")

	return &ListUnitsWithFullHierarchyResponse{
		Items: result,
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

	// 2. 处理 branch_id：优先使用 BranchID，如果没有则通过 BranchName 查找
	var branchID sql.NullString
	if req.BranchID != "" {
		// 优先使用 BranchID
		branchID = sql.NullString{String: req.BranchID, Valid: true}
	} else if req.BranchName != "" {
		// 向后兼容：通过 BranchName 查找 BranchID
		// Vue 层已经 trim 空格，这里处理空字符串自动转换为 "-"
		branchNameTrimmed := strings.TrimSpace(req.BranchName)
		if branchNameTrimmed == "" {
			branchNameTrimmed = "-"
		}
		if branchNameTrimmed != "-" {
			branch, err := s.branchesRepo.GetBranchByName(ctx, req.TenantID, branchNameTrimmed)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, fmt.Errorf("branch not found: branch_name=%s", branchNameTrimmed)
				}
				return nil, fmt.Errorf("failed to find branch: %w", err)
			}
			branchID = sql.NullString{String: branch.BranchID, Valid: true}
		}
		// 如果 branchNameTrimmed == "-"，不设置 branchID（保持 NULL），表示默认 branch
	}

	// 3. 处理 building_id：可选，可以为空（支持 home care 等无 building 的场景）
	var buildingID sql.NullString
	if req.BuildingID != "" {
		// 优先使用 BuildingID
		buildingID = sql.NullString{String: req.BuildingID, Valid: true}
		// 验证 building_id 是否存在
		_, err := s.unitsRepo.GetBuilding(ctx, req.TenantID, req.BuildingID)
		if err != nil {
			if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("building not found: building_id=%s", req.BuildingID)
			}
			return nil, fmt.Errorf("failed to validate building: %w", err)
		}
	} else if req.BuildingName != "" {
		// 向后兼容：通过 BuildingName 查找 BuildingID
		buildingNameTrimmed := strings.TrimSpace(req.BuildingName)
		if buildingNameTrimmed == "" {
			// building_name 为空时，不查找，building_id 保持为 NULL（允许为空）
			buildingID = sql.NullString{Valid: false}
		} else {
			// 通过 ListBuildings 查找 building（需要匹配 branch_id 和 building_name）
			buildings, err := s.unitsRepo.ListBuildings(ctx, req.TenantID, branchID, buildingNameTrimmed)
			if err != nil {
				return nil, fmt.Errorf("failed to find building: %w", err)
			}
			if len(buildings) == 0 {
				return nil, fmt.Errorf("building not found: building_name=%s", buildingNameTrimmed)
			}
			if len(buildings) > 1 {
				// 如果找到多个，优先选择匹配 branch_id 的
				var matchedBuilding *domain.Building
				for _, b := range buildings {
					if branchID.Valid && b.BranchID.Valid && b.BranchID.String == branchID.String {
						matchedBuilding = b
						break
					}
				}
				if matchedBuilding == nil {
					matchedBuilding = buildings[0] // 如果都不匹配，使用第一个
				}
				buildingID = sql.NullString{String: matchedBuilding.BuildingID, Valid: true}
			} else {
				buildingID = sql.NullString{String: buildings[0].BuildingID, Valid: true}
			}
		}
	}
	// 如果都没有提供，building_id 保持为 NULL（允许为空，支持 home care 场景）

	// 4. 应用默认值和格式转换（可选字段）
	unitType := normalizeUnitType(req.UnitType) // "" → "Facility"
	floor := normalizeFloor(req.Floor)          // ""/"1"/1 → sql.NullString{String: "1F", Valid: true}
	timezone := normalizeTimezone(req.Timezone) // "" → "America/Denver" (IANA 标识符)

	// 5. 构建 domain.Unit（使用新的字段名）
	unit := &domain.Unit{
		TenantID:     req.TenantID,
		BranchID:     branchID,
		UnitName:     strings.TrimSpace(req.UnitName),
		BuildingID:   buildingID,
		Floor:        floor,
		LayoutConfig: normalizeLayoutConfig(req.LayoutConfig),
		UnitType:     unitType,
		IsPublic:     req.IsPublicSpace,
		IsSharedUnit: req.IsSharedUnit,
		Timezone:     timezone,
	}

	// 6. 业务规则验证
	// 注意：branch_id 现在必须提供（已在步骤 2 中验证），building_id 可以为 NULL
	// 不需要额外的验证

	// 6. 调用 Repository
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

	// 7. 构建响应
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

	// 3. 构建更新后的 unit（只更新提供的字段，使用新的字段名）
	unit := &domain.Unit{
		UnitID:       req.UnitID,
		TenantID:     req.TenantID,
		BranchID:     currentUnit.BranchID,
		UnitName:     currentUnit.UnitName,
		BuildingID:   currentUnit.BuildingID,
		Floor:        currentUnit.Floor,
		LayoutConfig: currentUnit.LayoutConfig,
		UnitType:     currentUnit.UnitType,
		IsPublic:     currentUnit.IsPublic,
		IsSharedUnit: currentUnit.IsSharedUnit,
		Timezone:     currentUnit.Timezone,
	}

	// 4. 处理 branch_id：必须提供，不能为空（业务规则：一家机构必然有一个分支或总部）
	if req.BranchID != "" {
		// 优先使用 BranchID
		unit.BranchID = sql.NullString{String: req.BranchID, Valid: true}
	} else if req.BranchName != "" {
		// 向后兼容：通过 BranchName 查找 BranchID
		branchNameTrimmed := strings.TrimSpace(req.BranchName)
		if branchNameTrimmed == "" {
			return nil, fmt.Errorf("branch_id or branch_name is required and cannot be empty")
		}
		branch, err := s.branchesRepo.GetBranchByName(ctx, req.TenantID, branchNameTrimmed)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("branch not found: branch_name=%s", branchNameTrimmed)
			}
			return nil, fmt.Errorf("failed to find branch: %w", err)
		}
		unit.BranchID = sql.NullString{String: branch.BranchID, Valid: true}
	}
	// 如果都没有提供，保持当前值（不更新）
	// 注意：不允许将 branch_id 设置为 NULL（业务规则要求）

	if req.UnitName != "" {
		unit.UnitName = strings.TrimSpace(req.UnitName)
	}

	// 5. 处理 building_id：可选，可以为空（支持 home care 等无 building 的场景）
	if req.BuildingID != "" {
		// 优先使用 BuildingID
		unit.BuildingID = sql.NullString{String: req.BuildingID, Valid: true}
		// 验证 building_id 是否存在
		_, err := s.unitsRepo.GetBuilding(ctx, req.TenantID, req.BuildingID)
		if err != nil {
			if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("building not found: building_id=%s", req.BuildingID)
			}
			return nil, fmt.Errorf("failed to validate building: %w", err)
		}
	} else if req.BuildingName != "" {
		// 向后兼容：通过 BuildingName 查找 BuildingID
		buildingNameTrimmed := strings.TrimSpace(req.BuildingName)
		if buildingNameTrimmed == "" {
			// building_name 为空时，设置为 NULL（允许为空）
			unit.BuildingID = sql.NullString{Valid: false}
		} else {
			// 通过 ListBuildings 查找 building（需要匹配 branch_id 和 building_name）
			currentBranchID := unit.BranchID
			if !currentBranchID.Valid {
				currentBranchID = currentUnit.BranchID
			}
			buildings, err := s.unitsRepo.ListBuildings(ctx, req.TenantID, currentBranchID, buildingNameTrimmed)
			if err != nil {
				return nil, fmt.Errorf("failed to find building: %w", err)
			}
			if len(buildings) == 0 {
				return nil, fmt.Errorf("building not found: building_name=%s", buildingNameTrimmed)
			}
			if len(buildings) > 1 {
				// 如果找到多个，优先选择匹配 branch_id 的
				var matchedBuilding *domain.Building
				for _, b := range buildings {
					if currentBranchID.Valid && b.BranchID.Valid && b.BranchID.String == currentBranchID.String {
						matchedBuilding = b
						break
					}
				}
				if matchedBuilding == nil {
					matchedBuilding = buildings[0] // 如果都不匹配，使用第一个
				}
				unit.BuildingID = sql.NullString{String: matchedBuilding.BuildingID, Valid: true}
			} else {
				unit.BuildingID = sql.NullString{String: buildings[0].BuildingID, Valid: true}
			}
		}
	} else if req.BuildingName == "" && currentUnit.BuildingID.Valid {
		// 请求值为空字符串且当前值存在，清除它（设置为 NULL，允许为空）
		unit.BuildingID = sql.NullString{Valid: false}
	}
	// 如果请求中未提供 building_id 或 building_name，保持原值（不更新）

	if req.Floor != "" {
		unit.Floor = normalizeFloor(req.Floor)
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
		unit.IsPublic = *req.IsPublicSpace
	}
	if req.IsSharedUnit != nil {
		unit.IsSharedUnit = *req.IsSharedUnit
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

	if s.cardSync != nil {
		if _, err := s.cardSync.CreateCardsForUnit(ctx, req.TenantID, req.UnitID); err != nil {
			s.logger.Warn("Failed to sync cards after unit change", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("unit_id", req.UnitID))
		} else {
			s.logger.Info("Synced cards after unit change", zap.String("tenant_id", req.TenantID), zap.String("unit_id", req.UnitID))
		}
	}

	// 6. 构建响应
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
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 2. 获取 branch_id 用于权限验证（从 user_branches 表查询）
	branchID, _, err := s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 3. 验证权限（tenant + branch）
	unit, err := s.verifyUnitPermission(ctx, req.TenantID, req.UnitID, branchID)
	if err != nil {
		return nil, err
	}

	// 3. 检查关联数据：residents、devices、rooms
	var errorDetails []string

	// 3.1 检查 residents
	residentFilters := repository.ResidentFilters{
		UnitID: req.UnitID,
	}
	residents, _, err := s.residentsRepo.ListResidents(ctx, req.TenantID, residentFilters, 1, 1000)
	if err != nil {
		s.logger.Error("DeleteUnit: failed to check residents",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check residents: %w", err)
	}
	if len(residents) > 0 {
		residentNames := make([]string, 0, len(residents))
		for _, r := range residents {
			if r.Nickname != "" {
				residentNames = append(residentNames, r.Nickname)
			} else {
				residentNames = append(residentNames, r.ResidentID)
			}
		}
		errorDetails = append(errorDetails, fmt.Sprintf("residents: %s", strings.Join(residentNames, ", ")))
	}

	// 3.2 检查 rooms（以及通过 rooms 关联的 beds）
	rooms, err := s.unitsRepo.ListRooms(ctx, req.TenantID, req.UnitID, "")
	if err != nil {
		s.logger.Error("DeleteUnit: failed to check rooms",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check rooms: %w", err)
	}
	if len(rooms) > 0 {
		roomNames := make([]string, 0, len(rooms))
		for _, room := range rooms {
			roomNames = append(roomNames, room.RoomName)
		}
		errorDetails = append(errorDetails, fmt.Sprintf("rooms: %s", strings.Join(roomNames, ", ")))

		// 检查每个 room 下的 beds
		allBedNames := make([]string, 0)
		for _, room := range rooms {
			beds, err := s.unitsRepo.ListBeds(ctx, req.TenantID, room.RoomID, "")
			if err != nil {
				s.logger.Warn("DeleteUnit: failed to check beds for room",
					zap.String("tenant_id", req.TenantID),
					zap.String("room_id", room.RoomID),
					zap.Error(err),
				)
				continue
			}
			for _, bed := range beds {
				allBedNames = append(allBedNames, fmt.Sprintf("%s/%s", room.RoomName, bed.BedName))
			}
		}
		if len(allBedNames) > 0 {
			errorDetails = append(errorDetails, fmt.Sprintf("beds: %s", strings.Join(allBedNames, ", ")))
		}
	}

	// 3.3 检查 devices（通过 bound_room_id 或 bound_bed_id 间接关联到 unit）
	// 查询所有 rooms 的 room_id
	roomIDs := make([]string, 0, len(rooms))
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.RoomID)
	}
	// 查询所有 beds 的 bed_id
	bedIDs := make([]string, 0)
	for _, room := range rooms {
		beds, err := s.unitsRepo.ListBeds(ctx, req.TenantID, room.RoomID, "")
		if err != nil {
			continue
		}
		for _, bed := range beds {
			bedIDs = append(bedIDs, bed.BedID)
		}
	}

	// 查询绑定到这些 rooms 或 beds 的 devices
	deviceFilters := repository.DeviceFilters{}
	allDevices, _, err := s.devicesRepo.ListDevices(ctx, req.TenantID, deviceFilters, 1, 10000)
	if err != nil {
		s.logger.Error("DeleteUnit: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
	}

	// 过滤出绑定到该 unit 的 devices
	unitDevices := make([]*domain.Device, 0)
	for _, device := range allDevices {
		// 检查是否绑定到该 unit 下的 room
		if device.BoundRoomID.Valid {
			for _, roomID := range roomIDs {
				if device.BoundRoomID.String == roomID {
					unitDevices = append(unitDevices, device)
					break
				}
			}
		}
		// 检查是否绑定到该 unit 下的 bed
		if device.BoundBedID.Valid {
			for _, bedID := range bedIDs {
				if device.BoundBedID.String == bedID {
					unitDevices = append(unitDevices, device)
					break
				}
			}
		}
	}

	if len(unitDevices) > 0 {
		deviceNames := make([]string, 0, len(unitDevices))
		for _, d := range unitDevices {
			deviceNames = append(deviceNames, d.DeviceName)
		}
		errorDetails = append(errorDetails, fmt.Sprintf("devices: %s", strings.Join(deviceNames, ", ")))
	}

	// 4. 如果有关联数据，禁止删除并返回详细信息
	if len(errorDetails) > 0 {
		errorMsg := fmt.Sprintf("cannot delete unit: unit has associated data (%s)", strings.Join(errorDetails, "; "))
		s.logger.Warn("DeleteUnit: unit has associated data",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.String("unit_name", unit.UnitName),
			zap.Strings("error_details", errorDetails),
		)
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// 5. 调用 Repository 删除
	err = s.unitsRepo.DeleteUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		s.logger.Error("DeleteUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete unit: %w", err)
	}

	// 6. 构建响应
	return &DeleteUnitResponse{
		Success: true,
	}, nil
}

// ============================================
// Room 方法实现
// ============================================

// ListRooms 查询房间列表
func (s *unitService) ListRooms(ctx context.Context, req ListRoomsRequest) (*ListRoomsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 2. 获取 branch_id 用于权限验证（从 user_branches 表查询）
	branchID, _, err := s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 3. 验证权限（tenant + branch）
	_, err = s.verifyUnitPermission(ctx, req.TenantID, req.UnitID, branchID)
	if err != nil {
		return nil, err
	}

	// 3. 调用 Repository（传递搜索参数）
	search := strings.TrimSpace(req.Search)
	items, err := s.unitsRepo.ListRooms(ctx, req.TenantID, req.UnitID, search)
	if err != nil {
		s.logger.Error("ListRooms failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.String("search", search),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	return &ListRoomsResponse{
		Items: items,
	}, nil
}

// ListRoomsWithBeds 查询房间及其床位列表
// ListRoomsWithBeds 查询房间及其床位列表
// 注意：如果用户已经在编辑 unit，说明已经通过了权限验证（unit 的 branch_id 在用户的 branch_id 列表中）
// 因此在获取同一个 unit 的 rooms 时，不需要再次验证权限，只需要验证 unit 是否存在即可
func (s *unitService) ListRoomsWithBeds(ctx context.Context, req ListRoomsWithBedsRequest) (*ListRoomsWithBedsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UnitID == "" {
		return nil, fmt.Errorf("unit_id is required")
	}

	// 2. 验证 unit 是否存在（Repository 层会自动验证 tenant_id）
	// 如果 unit 不存在或不属于该 tenant，会返回错误
	_, err := s.unitsRepo.GetUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("unit not found: unit_id=%s (tenant validation failed or unit does not exist)", req.UnitID)
		}
		s.logger.Error("ListRoomsWithBeds: failed to get unit",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 3. 调用 Repository（传递搜索参数）
	search := strings.TrimSpace(req.Search)
	items, err := s.unitsRepo.ListRoomsWithBeds(ctx, req.TenantID, req.UnitID, search)
	if err != nil {
		s.logger.Error("ListRoomsWithBeds failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.String("search", search),
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
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 获取 branch_id 用于权限验证（从 user_branches 表查询）
	branchID, _, err := s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 验证权限（tenant + branch）
	room, _, err := s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchID)
	if err != nil {
		return nil, err
	}

	return &GetRoomResponse{
		Room: room,
	}, nil
}

// CreateRoom 创建房间
// 注意：如果用户已经在编辑 unit，说明已经通过了权限验证（unit 的 branch_id 在用户的 branch_id 列表中）
// 因此在同一个 unit 下添加 room 时，不需要再次验证权限，只需要验证 unit 是否存在即可
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

	// 验证 unit 是否存在（Repository 层会自动验证 tenant_id）
	// 如果 unit 不存在或不属于该 tenant，会返回错误
	_, err := s.unitsRepo.GetUnit(ctx, req.TenantID, req.UnitID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("unit not found: unit_id=%s (tenant validation failed or unit does not exist)", req.UnitID)
		}
		s.logger.Error("CreateRoom: failed to get unit",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
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
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 获取 branch_id 用于权限验证（从 user_branches 表查询）
	branchID, _, err := s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 验证权限（tenant + branch）
	currentRoom, _, err := s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchID)
	if err != nil {
		return nil, err
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
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 2. 先获取 room 和 unit 信息（用于确定 unit 的 branch_id）
	room, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("room not found: room_id=%s", req.RoomID)
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	unit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, room.UnitID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("unit not found: unit_id=%s", room.UnitID)
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 3. 获取 branch_id 用于权限验证
	// 如果请求中提供了 branchID，验证用户是否有权限访问该 branch
	// 如果未提供，检查用户是否属于该 unit 的 branch
	var branchID string
	if req.BranchID != "" {
		// 提供了 branchID，验证用户是否有权限访问该 branch
		branchID, _, err = s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
		if err != nil {
			return nil, err
		}
	} else {
		// 未提供 branchID，检查用户是否属于该 unit 的 branch
		if unit.BranchID.Valid && unit.BranchID.String != "" {
			// 检查用户是否属于该 unit 的 branch
			userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to verify user branch permission: %w", err)
			}
			if hasBranches {
				// 检查用户的 branch 列表中是否包含 unit 的 branch
				found := false
				for _, bid := range userBranchIDs {
					if bid == unit.BranchID.String {
						found = true
						branchID = bid
						break
					}
				}
				if !found {
					// 用户不属于该 unit 的 branch
					var branchName sql.NullString
					s.db.QueryRowContext(ctx,
						`SELECT branch_name FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
						req.TenantID, unit.BranchID.String,
					).Scan(&branchName)
					var unitBranchName string
					if branchName.Valid {
						unitBranchName = branchName.String
					}
					return nil, fmt.Errorf("permission denied: unit does not belong to any of your branches (unit belongs to branch_name=%s, branch_id=%s)", unitBranchName, unit.BranchID.String)
				}
			} else {
				// 用户没有关联任何院区，可以访问所有院区
				branchID = "" // 空字符串表示跳过 branch 验证
			}
		} else {
			// unit 没有关联 branch，跳过 branch 验证
			branchID = ""
		}
	}

	// 4. 验证权限（tenant + branch）
	_, _, err = s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchID)
	if err != nil {
		return nil, err
	}

	// 3. 检查关联数据：beds、devices
	var errorDetails []string

	// 3.1 检查 beds
	beds, err := s.unitsRepo.ListBeds(ctx, req.TenantID, req.RoomID, "")
	if err != nil {
		s.logger.Error("DeleteRoom: failed to check beds",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check beds: %w", err)
	}
	if len(beds) > 0 {
		bedNames := make([]string, 0, len(beds))
		for _, bed := range beds {
			bedNames = append(bedNames, bed.BedName)
		}
		errorDetails = append(errorDetails, fmt.Sprintf("beds: %s", strings.Join(bedNames, ", ")))
	}

	// 3.2 检查 devices（通过 bound_room_id 关联到 room）
	// 注意：DeviceFilters 没有 BoundRoomID 字段，需要直接查询数据库
	// 这里我们需要通过 devicesRepo 的内部方法或直接查询，但为了保持架构清晰，
	// 我们可以通过 ListDevices 获取所有设备，然后过滤 bound_room_id
	// 或者，我们可以添加一个专门的查询方法，但为了简化，先使用 ListDevices 然后过滤
	allDevices, _, err := s.devicesRepo.ListDevices(ctx, req.TenantID, repository.DeviceFilters{}, 1, 10000)
	if err != nil {
		s.logger.Error("DeleteRoom: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
	}
	// 过滤出绑定到该 room 的 devices
	boundDevices := make([]*domain.Device, 0)
	for _, device := range allDevices {
		if device.BoundRoomID.Valid && device.BoundRoomID.String == req.RoomID {
			boundDevices = append(boundDevices, device)
		}
	}
	if len(boundDevices) > 0 {
		deviceNames := make([]string, 0, len(boundDevices))
		for _, device := range boundDevices {
			if device.DeviceName != "" {
				deviceNames = append(deviceNames, device.DeviceName)
			} else if device.DeviceUID != "" {
				deviceNames = append(deviceNames, device.DeviceUID)
			} else {
				deviceNames = append(deviceNames, device.DeviceID)
			}
		}
		errorDetails = append(errorDetails, fmt.Sprintf("devices: %s", strings.Join(deviceNames, ", ")))
	}

	// 4. 如果有关联数据，禁止删除并返回详细错误信息
	if len(errorDetails) > 0 {
		errorMsg := fmt.Sprintf("cannot delete room '%s' (unit: %s) because it has associated data: %s",
			room.RoomName,
			func() string {
				if unit.UnitName != "" {
					return unit.UnitName
				}
				return unit.UnitID
			}(),
			strings.Join(errorDetails, "; "),
		)
		s.logger.Warn("DeleteRoom: room has associated data",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.String("room_name", room.RoomName),
			zap.Strings("error_details", errorDetails),
		)
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// 5. 调用 Repository 删除
	err = s.unitsRepo.DeleteRoom(ctx, req.TenantID, req.RoomID)
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
// ListBeds 查询床位列表
// 注意：如果用户已经在编辑 room，说明已经通过了权限验证（room 所属的 unit 的 branch_id 在用户的 branch_id 列表中）
// 因此在获取同一个 room 的 beds 时，不需要再次验证权限，只需要验证 room 是否存在即可
func (s *unitService) ListBeds(ctx context.Context, req ListBedsRequest) (*ListBedsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	// 2. 验证 room 是否存在（Repository 层会自动验证 tenant_id）
	// 如果 room 不存在或不属于该 tenant，会返回错误
	_, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("room not found: room_id=%s (tenant validation failed or room does not exist)", req.RoomID)
		}
		s.logger.Error("ListBeds: failed to get room",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// 3. 调用 Repository（传递搜索参数）
	search := strings.TrimSpace(req.Search)
	items, err := s.unitsRepo.ListBeds(ctx, req.TenantID, req.RoomID, search)
	if err != nil {
		s.logger.Error("ListBeds failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.String("search", search),
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
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 检查权限 scope：如果是 'A'（无限制），跳过 branch 验证
	var branchID string
	if req.CurrentUserRole != "" {
		perm, err := s.getResourcePermission(ctx, req.CurrentUserRole, "beds", "R")
		if err == nil && !perm.BranchOnly {
			// 权限 scope 为 'A'，跳过 branch 验证
			branchID = ""
		} else {
			// 权限 scope 为 'B' 或 'S'，需要 branch 验证
			// 获取 branch_id 用于权限验证（从 user_branches 表查询）
			branchID, _, err = s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// 未提供 CurrentUserRole，使用默认逻辑（需要 branch 验证）
		var err error
		branchID, _, err = s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
		if err != nil {
			return nil, err
		}
	}

	// 验证权限（tenant + branch）
	bed, _, _, err := s.verifyBedPermission(ctx, req.TenantID, req.BedID, branchID)
	if err != nil {
		return nil, err
	}

	return &GetBedResponse{
		Bed: bed,
	}, nil
}

// CreateBed 创建床位
// 注意：如果用户已经在编辑 unit/room，说明已经通过了权限验证（unit 的 branch_id 在用户的 branch_id 列表中）
// 因此在同一个 room 下添加 bed 时，不需要再次验证权限，只需要验证 room 是否存在即可
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

	// 验证 room 是否存在（Repository 层会自动验证 tenant_id）
	// 如果 room 不存在或不属于该 tenant，会返回错误
	_, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("room not found: room_id=%s (tenant validation failed or room does not exist)", req.RoomID)
		}
		s.logger.Error("CreateBed: failed to get room",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// 验证 bed_name 唯一性：同一 room 下不能有重复的 bed_name
	// 数据库约束：beds_tenant_id_room_id_bed_name_key UNIQUE (tenant_id, room_id, bed_name)
	bedNameTrimmed := strings.TrimSpace(req.BedName)
	if bedNameTrimmed == "" {
		return nil, fmt.Errorf("bed_name is required and cannot be empty or whitespace")
	}

	// 检查是否已存在相同 bed_name 的 bed（在同一 room 下）
	// ListBeds 需要 tenantID, roomID, 和可选的 currentUserID（用于权限验证，这里传空字符串）
	beds, err := s.unitsRepo.ListBeds(ctx, req.TenantID, req.RoomID, "")
	if err != nil {
		s.logger.Error("CreateBed: failed to list beds for uniqueness check",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check bed name uniqueness: %w", err)
	}

	for _, existingBed := range beds {
		if strings.EqualFold(existingBed.BedName, bedNameTrimmed) {
			return nil, fmt.Errorf("bed_name already exists in this room: bed_name='%s' (case-insensitive, constraint: beds_tenant_id_room_id_bed_name_key)", bedNameTrimmed)
		}
	}

	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bed := &domain.Bed{
		TenantID:          req.TenantID,
		RoomID:            req.RoomID,
		BedName:           bedNameTrimmed,
		MattressMaterial:  normalizeMattressMaterial(req.MattressMaterial),
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
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required for permission validation")
	}

	// 获取 branch_id 用于权限验证（从 user_branches 表查询）
	branchID, _, err := s.getBranchIDForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 验证权限（tenant + branch）
	currentBed, _, _, err := s.verifyBedPermission(ctx, req.TenantID, req.BedID, branchID)
	if err != nil {
		return nil, err
	}

	// 构建更新后的 bed
	bed := &domain.Bed{
		BedID:             req.BedID,
		TenantID:          req.TenantID,
		RoomID:            currentBed.RoomID,
		BedName:           currentBed.BedName,
		MattressMaterial:  currentBed.MattressMaterial,
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
// 注意：如果用户已经在编辑 unit，说明已经通过了权限验证（unit 的 branch_id 在用户的 branch_id 列表中）
// 因此在同一个 unit 下删除 bed 时，不需要再次验证权限，只需要验证 bed 是否存在即可
func (s *unitService) DeleteBed(ctx context.Context, req DeleteBedRequest) (*DeleteBedResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}
	// CurrentUserID is optional for logging, not for permission validation here
	// BranchID is no longer used for permission validation here

	// 2. 获取 bed 信息（Repository 层会自动验证 tenant_id）
	bed, err := s.unitsRepo.GetBed(ctx, req.TenantID, req.BedID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("bed not found: bed_id=%s (tenant validation failed or bed does not exist)", req.BedID)
		}
		s.logger.Error("DeleteBed: failed to get bed",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get bed: %w", err)
	}

	// 3. 获取 room 和 unit 信息（用于后续检查）
	room, err := s.unitsRepo.GetRoom(ctx, req.TenantID, bed.RoomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("room not found: room_id=%s (bed's room does not exist)", bed.RoomID)
		}
		s.logger.Error("DeleteBed: failed to get room",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", bed.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	unit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, room.UnitID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("unit not found: unit_id=%s (room's unit does not exist)", room.UnitID)
		}
		s.logger.Error("DeleteBed: failed to get unit",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", room.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	// 4. 如果 unit 有 branch_id，验证用户是否有权限访问该 branch
	// 简化逻辑：用户已在编辑 unit，说明已通过权限验证，无需再次验证
	// 此处仅保留 bed/room/unit 存在性检查
	_ = unit // Mark unit as used to avoid linter warning

	// 3. 检查关联数据：devices、residents
	var errorDetails []string

	// 3.1 检查 devices（通过 bound_bed_id 关联到 bed）
	allDevices, _, err := s.devicesRepo.ListDevices(ctx, req.TenantID, repository.DeviceFilters{}, 1, 10000)
	if err != nil {
		s.logger.Error("DeleteBed: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
	}
	// 过滤出绑定到该 bed 的 devices
	boundDevices := make([]*domain.Device, 0)
	for _, device := range allDevices {
		if device.BoundBedID.Valid && device.BoundBedID.String == req.BedID {
			boundDevices = append(boundDevices, device)
		}
	}
	if len(boundDevices) > 0 {
		deviceNames := make([]string, 0, len(boundDevices))
		for _, device := range boundDevices {
			if device.DeviceName != "" {
				deviceNames = append(deviceNames, device.DeviceName)
			} else if device.DeviceUID != "" {
				deviceNames = append(deviceNames, device.DeviceUID)
			} else {
				deviceNames = append(deviceNames, device.DeviceID)
			}
		}
		errorDetails = append(errorDetails, fmt.Sprintf("devices: %s", strings.Join(deviceNames, ", ")))
	}

	// 3.2 检查 residents（通过 bed_id 关联到 bed）
	residentFilters := repository.ResidentFilters{
		BedID: req.BedID,
	}
	residents, _, err := s.residentsRepo.ListResidents(ctx, req.TenantID, residentFilters, 1, 1000)
	if err != nil {
		s.logger.Error("DeleteBed: failed to check residents",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check residents: %w", err)
	}
	if len(residents) > 0 {
		residentNames := make([]string, 0, len(residents))
		for _, r := range residents {
			if r.Nickname != "" {
				residentNames = append(residentNames, r.Nickname)
			} else {
				residentNames = append(residentNames, r.ResidentID)
			}
		}
		errorDetails = append(errorDetails, fmt.Sprintf("residents: %s", strings.Join(residentNames, ", ")))
	}

	// 4. 如果有关联数据，禁止删除并返回详细错误信息
	if len(errorDetails) > 0 {
		errorMsg := fmt.Sprintf("cannot delete bed '%s' (room: %s, unit: %s) because it has associated data: %s",
			bed.BedName,
			func() string {
				if room.RoomName != "" {
					return room.RoomName
				}
				return room.RoomID
			}(),
			func() string {
				if unit.UnitName != "" {
					return unit.UnitName
				}
				return unit.UnitID
			}(),
			strings.Join(errorDetails, "; "),
		)
		s.logger.Warn("DeleteBed: bed has associated data",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.String("bed_name", bed.BedName),
			zap.Strings("error_details", errorDetails),
		)
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// 5. 调用 Repository 删除
	err = s.unitsRepo.DeleteBed(ctx, req.TenantID, req.BedID)
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
