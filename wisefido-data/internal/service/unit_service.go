package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/scope"

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
	ListUnitsWithAvailability(ctx context.Context, req ListUnitsRequest) (*ListUnitsWithAvailabilityResponse, error)
	ListUnitsWithFullHierarchy(ctx context.Context, req ListUnitsWithFullHierarchyRequest) (*ListUnitsWithFullHierarchyResponse, error)
	GetUnit(ctx context.Context, req GetUnitRequest) (*GetUnitResponse, error)
	CreateUnit(ctx context.Context, req CreateUnitRequest) (*CreateUnitResponse, error)
	UpdateUnit(ctx context.Context, req UpdateUnitRequest) (*UpdateUnitResponse, error)
	DeleteUnit(ctx context.Context, req DeleteUnitRequest) (*DeleteUnitResponse, error)

	// Room 管理
	ListRooms(ctx context.Context, req ListRoomsRequest) (*ListRoomsResponse, error)
	ListRoomsWithBeds(ctx context.Context, req ListRoomsWithBedsRequest) (*ListRoomsWithBedsResponse, error)
	ListRoomsByBranch(ctx context.Context, req ListRoomsByBranchRequest) (*ListRoomsByBranchResponse, error)
	GetRoom(ctx context.Context, req GetRoomRequest) (*GetRoomResponse, error)
	CreateRoom(ctx context.Context, req CreateRoomRequest) (*CreateRoomResponse, error)
	UpdateRoom(ctx context.Context, req UpdateRoomRequest) (*UpdateRoomResponse, error)
	DeleteRoom(ctx context.Context, req DeleteRoomRequest) (*DeleteRoomResponse, error)

	// Bed 管理
	ListBeds(ctx context.Context, req ListBedsRequest) (*ListBedsResponse, error)
	ListBedsWithDetails(ctx context.Context, req ListBedsWithDetailsRequest) (*ListBedsWithDetailsResponse, error)
	GetBed(ctx context.Context, req GetBedRequest) (*GetBedResponse, error)
	CreateBed(ctx context.Context, req CreateBedRequest) (*CreateBedResponse, error)
	UpdateBed(ctx context.Context, req UpdateBedRequest) (*UpdateBedResponse, error)
	DeleteBed(ctx context.Context, req DeleteBedRequest) (*DeleteBedResponse, error)
}

// unitService 实现
type unitService struct {
	unitsRepo    repository.UnitsRepository
	branchesRepo repository.BranchesRepository
	devicesRepo  repository.DevicesRepository
	db           *sql.DB
	logger       *zap.Logger
}

// NewUnitService 创建 UnitService 实例
//
// 注意：v2 不再注入 ResidentsRepository — "查 prefix 下 resident" 改走本 service 内
// findResidentNamesInPrefix raw SQL（v2 schema 用 resident_unit.spatial_prefix）。
func NewUnitService(unitsRepo repository.UnitsRepository, branchesRepo repository.BranchesRepository, devicesRepo repository.DevicesRepository, db *sql.DB, logger *zap.Logger) UnitService {
	return &unitService{
		unitsRepo:    unitsRepo,
		branchesRepo: branchesRepo,
		devicesRepo:  devicesRepo,
		db:           db,
		logger:       logger,
	}
}

// findResidentNamesInPrefix — 查 spatial_prefix 下 active resident 的 nickname 列表
// v2 schema: resident_unit.spatial_prefix INET (CIDR /80/88/96)，prefix <<= 查询
// 用于删除 unit/room/bed 前检查占用，避免误删
func (s *unitService) findResidentNamesInPrefix(ctx context.Context, prefix string) ([]string, error) {
	if strings.TrimSpace(prefix) == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(r.nickname, host(r.resident_id))
		  FROM residents r
		  JOIN resident_unit ru ON ru.resident_id = r.resident_id
		 WHERE ru.valid_to IS NULL
		   AND ru.spatial_prefix <<= $1::INET
		   AND COALESCE(r.status, 'active') <> 'deleted'`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		out = append(out, name)
	}
	return out, rows.Err()
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

// getUserBranchIDs 查询用户的 Current Branch（user_branches.is_primary=TRUE）
//
// Phase 3：业务 scope 统一为单一 Current Branch（不再用 branch_ids 集合）。
// Step B: 优先从 ctx 取 scope.ScopeContext；fallback 直查 user_branches。
// Admin（user_branches 表无行）→ 走 handleAdminAllBranches 兜底，返所有 branch。
func (s *unitService) getUserBranchIDs(ctx context.Context, tenantID, userID string) (branchIDs []string, hasBranches bool, err error) {
	if userID == "" {
		return nil, false, nil
	}
	// 优先：从 middleware 注入的 ScopeContext 拿
	if sc := scope.MustFromContext(ctx); sc != nil && sc.UserID == userID {
		if sc.IsTenantWide() {
			return s.handleAdminAllBranches(ctx, tenantID, userID)
		}
		if sc.HasCurrentBranch() {
			return []string{sc.CurrentBranchID}, true, nil
		}
		return s.handleAdminAllBranches(ctx, tenantID, userID)
	}
	// fallback：直查 DB
	var bid sql.NullString
	err = s.db.QueryRowContext(ctx, `
		SELECT host(branch_prefix) || '/56'
		  FROM user_branches
		 WHERE user_id = $1::UUID
		   AND is_primary = TRUE
		   AND valid_to IS NULL
		 LIMIT 1`, userID).Scan(&bid)
	if err == sql.ErrNoRows {
		return s.handleAdminAllBranches(ctx, tenantID, userID)
	}
	if err != nil {
		return nil, false, fmt.Errorf("query current branch: %w", err)
	}
	if !bid.Valid || bid.String == "" {
		return s.handleAdminAllBranches(ctx, tenantID, userID)
	}
	return []string{bid.String}, true, nil
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

// getBranchIDsForPermission 统一获取用户可访问的 branch_id 列表用于权限验证
// 优先级：
//  1. 如果请求中提供了 BranchID，则验证该 BranchID 是否在用户绑定的 branch_id 列表中，返回 [providedBranchID]
//  2. 如果未提供，则返回用户的所有 branch_id 列表（支持多 branch 用户）
//
// 返回：
//   - branchIDs: 用于权限验证的 branch_id 列表（空表示可以访问所有院区）
func (s *unitService) getBranchIDsForPermission(ctx context.Context, tenantID, userID, providedBranchID string) (branchIDs []string, err error) {
	// 查询用户的所有 branch_id 列表
	userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user branch permission: %w", err)
	}

	if !hasBranches {
		return nil, nil
	}

	if providedBranchID != "" {
		found := false
		for _, bid := range userBranchIDs {
			if bid == providedBranchID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("user does not have permission to access branch %s (user's branches: %v)", providedBranchID, userBranchIDs)
		}
		return []string{providedBranchID}, nil
	}

	return userBranchIDs, nil
}

// verifyUnitPermission 验证 unit 的权限（tenant + branch）
// verifyUnitPermission 验证 unit 的权限（tenant + branch）
// 参数：
//   - tenantID: 必填，用于验证 unit 是否属于该 tenant
//   - unitID: 必填，要验证的 unit ID
//   - branchIDs: 用户可访问的 branch_id 列表；空列表表示跳过 branch 验证（scope 'A'）
//
// 返回：
//   - *domain.Unit: 验证通过后返回 unit 对象
//   - error: 如果验证失败（tenant 不匹配、unit 不存在、branch 不匹配等）
func (s *unitService) verifyUnitPermission(ctx context.Context, tenantID, unitID string, branchIDs []string) (*domain.Unit, error) {
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

	if len(branchIDs) == 0 {
		return unit, nil
	}

	if unit.BranchID.Valid && unit.BranchID.String != "" {
		for _, bid := range branchIDs {
			if bid == unit.BranchID.String {
				return unit, nil
			}
		}
	}

	// branch mismatch
	var branchName sql.NullString
	if unit.BranchID.Valid {
		_ = s.db.QueryRowContext(ctx,
			`SELECT branch_name FROM branches WHERE tenant_id = $1 AND branch_id = $2`,
			tenantID, unit.BranchID.String,
		).Scan(&branchName)
	}
	unitBranch := ""
	if branchName.Valid {
		unitBranch = branchName.String
	}
	s.logger.Warn("verifyUnitPermission: branch mismatch",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Strings("user_branch_ids", branchIDs),
		zap.String("unit_branch_id", func() string {
			if unit.BranchID.Valid {
				return unit.BranchID.String
			}
			return "NULL"
		}()),
		zap.String("unit_branch_name", unitBranch),
	)
	if unitBranch != "" {
		return nil, fmt.Errorf("permission denied: unit belongs to branch %s, not in user's branches", unitBranch)
	}
	return nil, fmt.Errorf("permission denied: unit does not belong to any of user's branches")
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
func (s *unitService) verifyRoomPermission(ctx context.Context, tenantID, roomID string, branchIDs []string) (*domain.Room, *domain.Unit, error) {
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

	unit, err := s.verifyUnitPermission(ctx, tenantID, room.UnitID, branchIDs)
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
func (s *unitService) verifyBedPermission(ctx context.Context, tenantID, bedID string, branchIDs []string) (*domain.Bed, *domain.Room, *domain.Unit, error) {
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

	room, unit, err := s.verifyRoomPermission(ctx, tenantID, bed.RoomID, branchIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return bed, room, unit, nil
}

// syncCardsForUnit 单元层级（unit/room/bed）或展示字段变化后刷新该 unit 下卡片
func (s *unitService) syncCardsForUnit(ctx context.Context, tenantID, unitID, _ string) {
	SyncUnitCards(ctx, tenantID, unitID)
}

// collectUnitTypeChangeBlockers 检查 unit 下是否已有 resident，
// 返回阻挡 unit_type 变更的实体类型列表（空表示可以变更）。
// 只锁 residents：切换 unit_type 会让原本被 Facility/Public/Shared 隔开的住户跨权限看到他人的卡片和 PHI。
// 设备绑定不锁：cards 重建只是体系切换，无 PHI 越权问题。
func (s *unitService) collectUnitTypeChangeBlockers(ctx context.Context, tenantID, unitID string) ([]string, error) {
	var blockers []string

	names, err := s.findResidentNamesInPrefix(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("check residents: %w", err)
	}
	if len(names) > 0 {
		blockers = append(blockers, "residents")
	}
	_ = tenantID

	return blockers, nil
}

// ============================================
// Building 相关请求/响应结构
// ============================================

type ListBuildingsRequest struct {
	TenantID      string // 必填
	BranchID      string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName    string // 可选（向后兼容，如果 BranchID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	CurrentUserID string // Phase 3：BranchID/Name 都没传时，按 user.is_primary (Current Branch) 兜底过滤
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
	// ResidentID: 住户绑定时传入（可为空表示新住户），Private 单元仅返回未被其他住户占用的
	ResidentID *string
}

type ListUnitsResponse struct {
	Items []*domain.Unit `json:"items"`
	Total int            `json:"total"`
}

// UnitWithAvailability 带 has_available_bed、is_bound 的 unit（供前端 (full) 灰行红字、橙/绿）
type UnitWithAvailability struct {
	*domain.Unit
	HasAvailableBed bool `json:"has_available_bed"`
	IsBound         bool `json:"is_bound"`
}

type ListUnitsWithAvailabilityResponse struct {
	Items []*UnitWithAvailability `json:"items"`
	Total int                     `json:"total"`
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
	TenantID     string // 必填
	BranchID     string // 可选（优先使用，如果提供则忽略 BranchName）
	BranchName   string // 可选（向后兼容，如果 BranchID 未提供则使用此字段查找 ID，空字符串自动转换为 "-"）
	UnitName     string // 必填
	BuildingID   string // 可选（优先使用，如果提供则忽略 BuildingName）
	BuildingName string // 可选（向后兼容）
	Floor        string // 可选（默认 "1F"）
	AreaName     string // 可选
	UnitNumber   string // 可选（v2 已删除该字段，但 FE 仍可发送，忽略）
	LayoutConfig string // 可选（JSON 字符串）
	// v2 双维度 (2026-05-09 重设计)
	UnitProperty int8 // 0=Home, 1=Facility (default)
	UnitType     int8 // 0=unknown, 1=single (Private), 2=share (default), 3=public
	Timezone     string // 必填
}

type CreateUnitResponse struct {
	UnitID string `json:"unit_id"`
}

type UpdateUnitRequest struct {
	TenantID     string // 必填
	UnitID       string // 必填
	BranchID     string // 可选
	BranchName   string // 可选
	UnitName     string // 可选
	BuildingID   string // 可选
	BuildingName string // 可选
	Floor        string // 可选
	AreaName     string // 可选
	UnitNumber   string // 可选
	LayoutConfig string // 可选（JSON 字符串）
	// v2 双维度 — 指针类型表示"未提供则不更新"
	UnitProperty *int8
	UnitType     *int8
	Timezone     string // 可选
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
	ResidentID    string // 可选，住户绑定时传当前 resident_id，只返回可用床位
	CurrentUserID string // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string // 可选（保留字段，不再使用）
	Search        string // 可选（按 room_name 模糊搜索）
}

// BedWithDeviceDetails 床位及其绑定的设备（字母 + monitor 状态），供前端 bed 行展示
type BedWithDeviceDetails struct {
	Bed     *domain.Bed
	Devices []BedDeviceDetail // R=Radar, S=Sleepad；monitoring_enabled 供颜色展示
}

// RoomWithBedsItem 房间及其床位，含 room 级设备类型字母、每 bed 的绑定设备及 monitor 状态
type RoomWithBedsItem struct {
	Room        *domain.Room
	Beds        []*BedWithDeviceDetails
	DeviceTypes []string // room 级：R=Radar, S=Sleepad，供 RoomName(R) 展示
}

type ListRoomsWithBedsResponse struct {
	Items []*RoomWithBedsItem `json:"items"`
}

type ListRoomsByBranchRequest struct {
	TenantID string // 必填
	BranchID string // 必填
}

// RoomWithAvailabilityItem 供 resident 选 unit 弹窗：含 device_types（R=Radar, S=Sleepad）供 RoomName(R) 展示
type RoomWithAvailabilityItem struct {
	*repository.RoomWithAvailability
	DeviceTypes []string
}

type ListRoomsByBranchResponse struct {
	Items []*RoomWithAvailabilityItem `json:"items"`
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
	TenantID      string  // 必填
	RoomID        string  // 必填
	ResidentID    *string // 可选，住户绑定时传当前 resident_id，只返回可用床位；nil=返回全部
	CurrentUserID string  // 可选（保留字段，用于日志记录，不再用于权限验证）
	BranchID      string  // 可选（保留字段，不再使用）
	Search        string  // 可选（按 bed_name 模糊搜索）
}

type ListBedsResponse struct {
	Items []*domain.Bed `json:"items"`
}

type ListBedsWithDetailsRequest struct {
	TenantID   string  // 必填
	RoomID     string  // 必填
	ResidentID *string // 可选，传则只返回可用床位
	Search     string
}

type BedDeviceDetail struct {
	Letter            string `json:"letter"`
	MonitoringEnabled bool   `json:"monitoring_enabled"`
}

type BedWithDetailsItem struct {
	BedID     string             `json:"bed_id"`
	BedName   string             `json:"bed_name"`
	ResidentID *string           `json:"resident_id,omitempty"`
	Devices   []BedDeviceDetail  `json:"devices"`
}

type ListBedsWithDetailsResponse struct {
	Items []*BedWithDetailsItem `json:"items"`
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

	// 处理 branch_id：优先使用 BranchID，否则通过 BranchName 查找
	// Phase 3：未显式传 BranchID/Name 时，默认按 user 的 Current Branch (is_primary) 过滤
	var branchID sql.NullString
	var branchNameForQuery string

	if req.BranchID != "" {
		branchID = sql.NullString{String: req.BranchID, Valid: true}
		branchNameForQuery = ""
	} else if req.BranchName != "" {
		branchNameTrimmed := strings.TrimSpace(req.BranchName)
		if branchNameTrimmed == "" {
			branchNameTrimmed = domain.DefaultBranchName
		}
		branchNameForQuery = branchNameTrimmed
	} else if req.CurrentUserID != "" {
		// Phase 3：默认按 Current Branch 过滤
		userBranches, _, _ := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
		if len(userBranches) == 1 {
			branchID = sql.NullString{String: userBranches[0], Valid: true}
		}
		// Admin 兜底返多 branch 的情况：保持 NULL（不过滤）
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
	// 优先使用 BranchID，如果提供则忽略 BranchName；先选 branch，再列该 branch 下可用的 units
	branchID := stringValueOrEmpty(req.BranchID)
	branchName := ""
	if branchID == "" {
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
		ResidentID: req.ResidentID,
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

// ListUnitsWithAvailability 查询 Units 并附带 has_available_bed、is_bound（供前端 (full) 灰行红字、橙/绿）
func (s *unitService) ListUnitsWithAvailability(ctx context.Context, req ListUnitsRequest) (*ListUnitsWithAvailabilityResponse, error) {
	resp, err := s.ListUnits(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return &ListUnitsWithAvailabilityResponse{Items: nil, Total: resp.Total}, nil
	}
	unitIDs := make([]string, 0, len(resp.Items))
	for _, u := range resp.Items {
		unitIDs = append(unitIDs, u.UnitID)
	}
	hasAvail, isBound, err := s.unitsRepo.GetUnitAvailability(ctx, req.TenantID, unitIDs)
	if err != nil {
		s.logger.Error("GetUnitAvailability failed", zap.String("tenant_id", req.TenantID), zap.Error(err))
		return nil, fmt.Errorf("failed to get unit availability: %w", err)
	}
	out := make([]*UnitWithAvailability, 0, len(resp.Items))
	for _, u := range resp.Items {
		out = append(out, &UnitWithAvailability{
			Unit:            u,
			HasAvailableBed: hasAvail[u.UnitID],
			IsBound:         isBound[u.UnitID],
		})
	}
	return &ListUnitsWithAvailabilityResponse{Items: out, Total: resp.Total}, nil
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
	units, total, err := s.unitsRepo.ListUnits(ctx, req.TenantID, filters, 1, 1000)
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
		// Vue 层已经 trim 空格，这里处理空字符串自动转换为默认院区名
		branchNameTrimmed := strings.TrimSpace(req.BranchName)
		if branchNameTrimmed == "" {
			branchNameTrimmed = domain.DefaultBranchName
		}
		if branchNameTrimmed != domain.DefaultBranchName {
			branch, err := s.branchesRepo.GetBranchByName(ctx, req.TenantID, branchNameTrimmed)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, fmt.Errorf("branch not found: branch_name=%s", branchNameTrimmed)
				}
				return nil, fmt.Errorf("failed to find branch: %w", err)
			}
			branchID = sql.NullString{String: branch.BranchID, Valid: true}
		}
		// 若为默认院区名则不设置 branchID（保持 NULL），与“未指定院区”一致
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
	floor := normalizeFloor(req.Floor)          // ""/"1"/1 → sql.NullString{String: "1F", Valid: true}
	timezone := normalizeTimezone(req.Timezone) // "" → "America/Denver" (IANA 标识符)

	// v2 双维度默认值 + 配对约束兜底（repo 层也会再 normalize 一次）
	unitProperty := req.UnitProperty // 0=Home, 1=Facility
	unitType := req.UnitType
	if unitProperty == 0 {
		unitType = 0 // Home 强制 unknown
	} else if unitType < 1 || unitType > 3 {
		unitType = 2 // Facility 默认 share
	}

	// 5. 构建 domain.Unit（使用新的字段名）
	unit := &domain.Unit{
		TenantID:     req.TenantID,
		BranchID:     branchID,
		UnitName:     strings.TrimSpace(req.UnitName),
		BuildingID:   buildingID,
		Floor:        floor,
		LayoutConfig: normalizeLayoutConfig(req.LayoutConfig),
		UnitProperty: unitProperty,
		UnitType:     unitType,
		Timezone:     timezone,
	}

	// 调用 Repository
	unitID, err := s.unitsRepo.CreateUnit(ctx, req.TenantID, unit)
	if err != nil {
		s.logger.Error("CreateUnit failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_name", req.UnitName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}

	// public unit (unit_type=3): 创建同名 synthetic resident "public-<unit_name>"
	// 让 card detail / device-monitor-setting 等下游 UI 有 resident_id 可显示，避免空 "—"
	if unitType == 3 {
		if err := s.createPublicResident(ctx, req.TenantID, unitID, unit.UnitName); err != nil {
			s.logger.Warn("createPublicResident failed (unit created OK)",
				zap.String("tenant_id", req.TenantID),
				zap.String("unit_id", unitID),
				zap.Error(err))
		}
	}

	// 8. 构建响应
	s.syncCardsForUnit(ctx, req.TenantID, unitID, "unit_create")

	return &CreateUnitResponse{
		UnitID: unitID,
	}, nil
}

// createPublicResident — 为 public unit (unit_type=3) 创建同名 synthetic resident
//   - nickname = "public-<unit_name>"
//   - spatial_prefix = unit_id (/80)
//   - resident_account = NULL (避免占用人类账号)
// DeleteUnit 时通过同名 + 同 prefix 匹配并删除。
func (s *unitService) createPublicResident(ctx context.Context, tenantID, unitID, unitName string) error {
	nickname := "public-" + strings.TrimSpace(unitName)
	tenantPrefix := strings.TrimSpace(tenantID)
	if !strings.Contains(tenantPrefix, "/") {
		tenantPrefix = tenantPrefix + "/48"
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 分 slot：per-tenant MAX+1（与 PostgresResidentsRepository.CreateResident 一致逻辑）
	var slot int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(resident_slot), 0) + 1
		  FROM residents
		 WHERE network(set_masklen(resident_id, 48)) = $1::INET`, tenantPrefix,
	).Scan(&slot); err != nil {
		return fmt.Errorf("alloc slot: %w", err)
	}
	if slot < 1 || slot >= 65535 {
		return fmt.Errorf("slot out of range: %d", slot)
	}

	// 构造 hoa：tenant /48 host 部分 + ":ff01:<slot hex>::"
	tenantHost := strings.SplitN(strings.TrimSuffix(tenantPrefix, "/48"), "::", 2)[0]
	hoa := fmt.Sprintf("%s:ff01:%x::", tenantHost, slot)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO residents (resident_id, resident_slot, nickname, status)
		VALUES ($1::INET, $2, $3, 'active')`,
		hoa, slot, nickname,
	); err != nil {
		return fmt.Errorf("insert resident: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO resident_unit (resident_id, spatial_prefix, move_reason)
		VALUES ($1::INET, $2::INET, 'initial')`,
		hoa, unitID,
	); err != nil {
		return fmt.Errorf("insert resident_unit: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	// 触发 CardSync — 让 /80 unit card 立即 LPM 匹配上新建的 public-xxx resident
	// （绕过了 residentsRepo hook，要手动调）
	if globalCardSync != nil {
		if err := globalCardSync.ReconcileCards(ctx, unitID); err != nil {
			s.logger.Warn("createPublicResident: ReconcileCards failed",
				zap.String("unit_id", unitID), zap.Error(err))
		}
	}
	s.logger.Info("public-unit synthetic resident created",
		zap.String("unit_id", unitID), zap.String("nickname", nickname), zap.String("resident_id", hoa))
	return nil
}

// renamePublicResidentForUnit — unit_type 保持 3 但 unit_name 变更时同步改 synthetic resident 的 nickname。
// 通过旧 nickname + 同 prefix 锁定，避免误改其他真实 resident。
func (s *unitService) renamePublicResidentForUnit(ctx context.Context, unitID, oldUnitName, newUnitName string) error {
	oldNick := "public-" + strings.TrimSpace(oldUnitName)
	newNick := "public-" + strings.TrimSpace(newUnitName)
	if oldNick == newNick {
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE residents r
		   SET nickname = $1, updated_at = NOW()
		 WHERE r.nickname = $2
		   AND EXISTS (
		     SELECT 1 FROM resident_unit ru
		      WHERE ru.resident_id = r.resident_id
		        AND ru.valid_to IS NULL
		        AND ru.spatial_prefix = $3::INET
		   )`, newNick, oldNick, unitID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 && globalCardSync != nil {
		_ = globalCardSync.ReconcileCards(ctx, unitID)
		s.logger.Info("public-unit synthetic resident renamed",
			zap.String("unit_id", unitID), zap.String("old", oldNick), zap.String("new", newNick))
	}
	return nil
}

// deletePublicResidentForUnit — DeleteUnit 时清除同名 public-xxx synthetic resident
// 匹配条件：nickname = 'public-<unit_name>' AND 当前 active 在该 unit /80 prefix 内
// resident_unit ON DELETE CASCADE 会自动清理空间绑定行。
func (s *unitService) deletePublicResidentForUnit(ctx context.Context, unitID, unitName string) error {
	nickname := "public-" + strings.TrimSpace(unitName)
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM residents r
		 WHERE r.nickname = $1
		   AND EXISTS (
		     SELECT 1 FROM resident_unit ru
		      WHERE ru.resident_id = r.resident_id
		        AND ru.valid_to IS NULL
		        AND ru.spatial_prefix = $2::INET
		   )`, nickname, unitID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		s.logger.Info("public-unit synthetic resident deleted",
			zap.String("unit_id", unitID), zap.String("nickname", nickname), zap.Int64("count", n))
		// 同步刷一次 unit card 的 resident_id（discharge → NoOne/Public 回落）
		if globalCardSync != nil {
			_ = globalCardSync.ReconcileCards(ctx, unitID)
		}
	}
	return nil
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
		UnitProperty: currentUnit.UnitProperty,
		UnitType:     currentUnit.UnitType,
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

	// v2 双维度更新：当 UnitProperty 或 UnitType 任一变化时校验
	// (Home → type=0；Facility → type ∈ {1,2,3})
	newProperty := currentUnit.UnitProperty
	newType := currentUnit.UnitType
	changed := false
	if req.UnitProperty != nil {
		newProperty = *req.UnitProperty
		changed = changed || (newProperty != currentUnit.UnitProperty)
	}
	if req.UnitType != nil {
		newType = *req.UnitType
		changed = changed || (newType != currentUnit.UnitType)
	}
	// 强制配对约束
	if newProperty == 0 {
		newType = 0
	} else if newType < 1 || newType > 3 {
		newType = 2
	}
	if changed {
		// 3→非3 切换前先清掉自己创建的 public-<old_name> synthetic resident，
		// 不让它把 collectUnitTypeChangeBlockers 的 residents 检查卡住
		if currentUnit.UnitType == 3 && newType != 3 {
			if err := s.deletePublicResidentForUnit(ctx, req.UnitID, currentUnit.UnitName); err != nil {
				s.logger.Warn("UpdateUnit: deletePublicResidentForUnit failed (continuing)",
					zap.String("unit_id", req.UnitID), zap.Error(err))
			}
		}
		// unit_type/property 是建模级字段：切换会触发卡片体系按新模型重建，
		// 已绑 resident/device 的 unit 切换会越权放开 PHI 可见性、错位 cards.residents 绑定。
		// 因此一旦 unit 已被使用，禁止切换。
		if blockers, err := s.collectUnitTypeChangeBlockers(ctx, req.TenantID, req.UnitID); err != nil {
			return nil, err
		} else if len(blockers) > 0 {
			return nil, fmt.Errorf("cannot change unit type/property: unit has bound %s, please unbind first", strings.Join(blockers, " and "))
		}
	}
	unit.UnitProperty = newProperty
	unit.UnitType = newType
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

	// public unit (unit_type=3) synthetic resident 同步：
	//   非3 → 3 : 新建 public-<name>
	//   3 → 3 且名字变 : 重命名 nickname + cards.card_name
	if newType == 3 && currentUnit.UnitType != 3 {
		if err := s.createPublicResident(ctx, req.TenantID, req.UnitID, unit.UnitName); err != nil {
			s.logger.Warn("UpdateUnit: createPublicResident failed",
				zap.String("unit_id", req.UnitID), zap.Error(err))
		}
	} else if newType == 3 && currentUnit.UnitType == 3 && unit.UnitName != currentUnit.UnitName {
		if err := s.renamePublicResidentForUnit(ctx, req.UnitID, currentUnit.UnitName, unit.UnitName); err != nil {
			s.logger.Warn("UpdateUnit: renamePublicResidentForUnit failed",
				zap.String("unit_id", req.UnitID), zap.Error(err))
		}
	}

	s.syncCardsForUnit(ctx, req.TenantID, req.UnitID, "unit_update")

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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	unit, err := s.verifyUnitPermission(ctx, req.TenantID, req.UnitID, branchIDs)
	if err != nil {
		return nil, err
	}

	// public unit (unit_type=3): 先清除 synthetic public-<unit_name> resident，
	// 避免它把 DeleteUnit 的 "residents not empty" 检查卡住
	if unit != nil && unit.UnitType == 3 {
		if err := s.deletePublicResidentForUnit(ctx, req.UnitID, unit.UnitName); err != nil {
			s.logger.Warn("deletePublicResidentForUnit failed (continuing)",
				zap.String("unit_id", req.UnitID), zap.Error(err))
		}
	}

	// 3. 检查关联数据：residents、devices、rooms
	var errorDetails []string

	// 3.1 检查 residents（v2 走 resident_unit.spatial_prefix）
	residentNames, err := s.findResidentNamesInPrefix(ctx, req.UnitID)
	if err != nil {
		s.logger.Error("DeleteUnit: failed to check residents",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check residents: %w", err)
	}
	if len(residentNames) > 0 {
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

	// 3.3 检查 devices：仅允许删除「无设备绑定」的 unit。过滤语义：绑定到该 unit 下任意 room 或 bed 的设备（原逻辑为拉全量再按 roomIDs/bedIDs 过滤，现改为 repo 按 bound_room_id/bound_bed_id 查询，业务一致）
	roomIDs := make([]string, 0, len(rooms))
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.RoomID)
	}
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

	unitDevices, err := s.devicesRepo.GetDevicesBoundToRoomsOrBeds(ctx, req.TenantID, roomIDs, bedIDs)
	if err != nil {
		s.logger.Error("DeleteUnit: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUnitPermission(ctx, req.TenantID, req.UnitID, branchIDs)
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
	items, err := s.unitsRepo.ListRoomsWithBeds(ctx, req.TenantID, req.UnitID, search, strings.TrimSpace(req.ResidentID))
	if err != nil {
		s.logger.Error("ListRoomsWithBeds failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.String("search", search),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms with beds: %w", err)
	}

	// 4. 汇总所有 bed_id，查询每个 bed 绑定的设备及 monitor 状态
	allBedIDs := make([]string, 0)
	for _, rwb := range items {
		for _, b := range rwb.Beds {
			allBedIDs = append(allBedIDs, b.BedID)
		}
	}
	bedDevDetails, _ := s.devicesRepo.GetDevicesBoundToBedsWithDetails(ctx, req.TenantID, allBedIDs)
	if bedDevDetails == nil {
		bedDevDetails = make(map[string][]repository.DeviceTypeDetail)
	}

	// 5. 为每个 room 查 room 级设备字母，并组装每 bed 的 devices（letter + monitoring_enabled）
	out := make([]*RoomWithBedsItem, 0, len(items))
	for _, rwb := range items {
		letters, _ := s.devicesRepo.GetRoomBoundDeviceTypeLetters(ctx, req.TenantID, rwb.Room.RoomID)
		if letters == nil {
			letters = []string{}
		}
		bedsWithDev := make([]*BedWithDeviceDetails, 0, len(rwb.Beds))
		for _, b := range rwb.Beds {
			devs := bedDevDetails[b.BedID]
			dd := make([]BedDeviceDetail, 0, len(devs))
			for _, d := range devs {
				dd = append(dd, BedDeviceDetail{Letter: d.Letter, MonitoringEnabled: d.MonitoringEnabled})
			}
			bedsWithDev = append(bedsWithDev, &BedWithDeviceDetails{Bed: b, Devices: dd})
		}
		out = append(out, &RoomWithBedsItem{
			Room:        rwb.Room,
			Beds:        bedsWithDev,
			DeviceTypes: letters,
		})
	}

	return &ListRoomsWithBedsResponse{
		Items: out,
	}, nil
}

// ListRoomsByBranch 按 branch 列出所有 room（带 is_full、is_bound、facility_type，供前端 (full) 灰行红字、橙/绿）
func (s *unitService) ListRoomsByBranch(ctx context.Context, req ListRoomsByBranchRequest) (*ListRoomsByBranchResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.BranchID == "" {
		return nil, fmt.Errorf("branch_id is required")
	}
	items, err := s.unitsRepo.ListRoomsByBranch(ctx, req.TenantID, req.BranchID)
	if err != nil {
		s.logger.Error("ListRoomsByBranch failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("branch_id", req.BranchID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list rooms by branch: %w", err)
	}
	out := make([]*RoomWithAvailabilityItem, 0, len(items))
	for _, it := range items {
		letters, _ := s.devicesRepo.GetRoomBoundDeviceTypeLetters(ctx, req.TenantID, it.RoomID)
		if letters == nil {
			letters = []string{}
		}
		out = append(out, &RoomWithAvailabilityItem{RoomWithAvailability: it, DeviceTypes: letters})
	}
	return &ListRoomsByBranchResponse{Items: out}, nil
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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	room, _, err := s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchIDs)
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
	unit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, req.UnitID)
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

	roomName := strings.TrimSpace(req.RoomName)
	if roomName == "" {
		roomName = unit.UnitName
	}

	existingRooms, err := s.unitsRepo.ListRooms(ctx, req.TenantID, req.UnitID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}
	for _, r := range existingRooms {
		if strings.EqualFold(r.RoomName, roomName) {
			return nil, fmt.Errorf("room name '%s' already exists in this unit", roomName)
		}
	}

	room := &domain.Room{
		TenantID:     req.TenantID,
		UnitID:       req.UnitID,
		RoomName:     roomName,
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

	s.syncCardsForUnit(ctx, req.TenantID, req.UnitID, "room_create")

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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	currentRoom, _, err := s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchIDs)
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

	s.syncCardsForUnit(ctx, req.TenantID, currentRoom.UnitID, "room_update")

	return &UpdateRoomResponse{
		Success: true,
	}, nil
}

// DeleteRoom 删除房间
// 规则：删除物理关联（room）前，Service 层必须确保没有业务关联（设备绑定 bound_room_id、住户绑定 room_id）；否则禁止删除。
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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	_, unit, err := s.verifyRoomPermission(ctx, req.TenantID, req.RoomID, branchIDs)
	if err != nil {
		return nil, err
	}

	// 3. 业务关联检查：无设备绑定、无住户绑定方可删除物理关联（room）
	var errorDetails []string

	// 3.1 物理子节点：存在 beds 须先删 bed，再删 room
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

	// 3.2 业务关联：设备绑定到该 room（bound_room_id）
	boundDevices, err := s.devicesRepo.GetDevicesBoundToRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		s.logger.Error("DeleteRoom: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
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

	// 3.3 业务关联：住户绑定到该 room（v2: resident_unit.spatial_prefix <<= room /88）
	residentNames, err := s.findResidentNamesInPrefix(ctx, req.RoomID)
	if err != nil {
		s.logger.Error("DeleteRoom: failed to check residents",
			zap.String("tenant_id", req.TenantID),
			zap.String("room_id", req.RoomID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check residents: %w", err)
	}
	if len(residentNames) > 0 {
		errorDetails = append(errorDetails, fmt.Sprintf("residents: %s", strings.Join(residentNames, ", ")))
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

	s.syncCardsForUnit(ctx, req.TenantID, room.UnitID, "room_delete")

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

	// 3. 住户绑定时传 ResidentID 则只返回可用床位，否则返回全部
	search := strings.TrimSpace(req.Search)
	var items []*domain.Bed
	if req.ResidentID != nil {
		items, err = s.unitsRepo.ListAvailableBeds(ctx, req.TenantID, req.RoomID, search, strings.TrimSpace(*req.ResidentID))
	} else {
		items, err = s.unitsRepo.ListBeds(ctx, req.TenantID, req.RoomID, search)
	}
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

// ListBedsWithDetails 返回床位列表（含 resident_id、设备类型及 monitor 状态），供 resident 弹窗 bed 状态色与 R/S 展示
func (s *unitService) ListBedsWithDetails(ctx context.Context, req ListBedsWithDetailsRequest) (*ListBedsWithDetailsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}
	_, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("room not found: room_id=%s", req.RoomID)
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	search := strings.TrimSpace(req.Search)
	withResident, err := s.unitsRepo.ListBedsWithResident(ctx, req.TenantID, req.RoomID, search)
	if err != nil {
		return nil, fmt.Errorf("failed to list beds: %w", err)
	}
	s.logger.Info("ListBedsWithDetails",
		zap.String("tenant_id", req.TenantID),
		zap.String("room_id", req.RoomID),
		zap.Int("beds_count", len(withResident)))
	// 不做可用性过滤，仅返回 bed 客观信息：是否绑人(resident_id)、是否绑设备及设备 monitor on/off
	bedIDs := make([]string, 0, len(withResident))
	for _, b := range withResident {
		bedIDs = append(bedIDs, b.Bed.BedID)
	}
	devDetails, err := s.devicesRepo.GetDevicesBoundToBedsWithDetails(ctx, req.TenantID, bedIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get device details for beds: %w", err)
	}
	items := make([]*BedWithDetailsItem, 0, len(withResident))
	for _, b := range withResident {
		devs := devDetails[b.Bed.BedID]
		dd := make([]BedDeviceDetail, 0, len(devs))
		for _, d := range devs {
			dd = append(dd, BedDeviceDetail{Letter: d.Letter, MonitoringEnabled: d.MonitoringEnabled})
		}
		items = append(items, &BedWithDetailsItem{
			BedID:      b.Bed.BedID,
			BedName:    b.Bed.BedName,
			ResidentID: b.ResidentID,
			Devices:    dd,
		})
	}
	return &ListBedsWithDetailsResponse{Items: items}, nil
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

	var branchIDs []string
	if req.CurrentUserRole != "" {
		perm, err := s.getResourcePermission(ctx, req.CurrentUserRole, "beds", "R")
		if err == nil && !perm.BranchOnly {
			branchIDs = nil
		} else {
			branchIDs, err = s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		var err error
		branchIDs, err = s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
		if err != nil {
			return nil, err
		}
	}

	bed, _, _, err := s.verifyBedPermission(ctx, req.TenantID, req.BedID, branchIDs)
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
	room, err := s.unitsRepo.GetRoom(ctx, req.TenantID, req.RoomID)
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

	s.syncCardsForUnit(ctx, req.TenantID, room.UnitID, "bed_create")

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

	branchIDs, err := s.getBranchIDsForPermission(ctx, req.TenantID, req.CurrentUserID, req.BranchID)
	if err != nil {
		return nil, err
	}

	currentBed, _, unit, err := s.verifyBedPermission(ctx, req.TenantID, req.BedID, branchIDs)
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

	if unit != nil {
		s.syncCardsForUnit(ctx, req.TenantID, unit.UnitID, "bed_update")
	}

	return &UpdateBedResponse{
		Success: true,
	}, nil
}

// DeleteBed 删除床位
// 规则：删除物理关联（bed）前，Service 层必须确保没有业务关联（设备绑定 bound_bed_id、住户绑定 bed_id）；否则禁止删除。
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

	// 3. 业务关联检查：无设备绑定、无住户绑定方可删除物理关联（bed）
	var errorDetails []string

	// 3.1 业务关联：设备绑定到该 bed（bound_bed_id）
	boundDevices, err := s.devicesRepo.GetDevicesBoundToBed(ctx, req.TenantID, req.BedID)
	if err != nil {
		s.logger.Error("DeleteBed: failed to check devices",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check devices: %w", err)
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

	// 3.2 业务关联：住户绑定到该 bed（v2: resident_unit.spatial_prefix <<= bed /96）
	residentNames, err := s.findResidentNamesInPrefix(ctx, req.BedID)
	if err != nil {
		s.logger.Error("DeleteBed: failed to check residents",
			zap.String("tenant_id", req.TenantID),
			zap.String("bed_id", req.BedID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check residents: %w", err)
	}
	if len(residentNames) > 0 {
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

	s.syncCardsForUnit(ctx, req.TenantID, room.UnitID, "bed_delete")

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
