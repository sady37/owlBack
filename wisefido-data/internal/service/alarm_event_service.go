package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// AlarmEventService 报警事件服务接口
type AlarmEventService interface {
	// 查询报警事件列表（支持复杂过滤、跨表 JOIN、权限过滤）
	ListAlarmEvents(ctx context.Context, req ListAlarmEventsRequest) (*ListAlarmEventsResponse, error)

	// 处理报警事件（确认或解决）
	HandleAlarmEvent(ctx context.Context, req HandleAlarmEventRequest) (*HandleAlarmEventResponse, error)
}

// alarmEventService 实现
type alarmEventService struct {
	alarmEventsRepo repository.AlarmEventsRepository
	devicesRepo     repository.DevicesRepository
	unitsRepo       repository.UnitsRepository
	usersRepo       repository.UsersRepository
	db              *sql.DB // 用于查询卡片信息（临时方案）
	logger          *zap.Logger
}

// NewAlarmEventService 创建 AlarmEventService 实例
func NewAlarmEventService(
	alarmEventsRepo repository.AlarmEventsRepository,
	devicesRepo repository.DevicesRepository,
	unitsRepo repository.UnitsRepository,
	usersRepo repository.UsersRepository,
	db *sql.DB,
	logger *zap.Logger,
) AlarmEventService {
	return &alarmEventService{
		alarmEventsRepo: alarmEventsRepo,
		devicesRepo:     devicesRepo,
		unitsRepo:       unitsRepo,
		usersRepo:       usersRepo,
		db:              db,
		logger:          logger,
	}
}

// ============================================
// Request/Response DTOs
// ============================================

// ListAlarmEventsRequest 查询报警事件列表请求
type ListAlarmEventsRequest struct {
	// 必填字段
	TenantID        string // 租户ID
	CurrentUserID   string // 当前用户ID（用于权限过滤）
	CurrentUserRole string // 当前用户角色（用于权限过滤）

	// 状态过滤
	Status string // 'active' | 'resolved' - 报警状态过滤

	// 时间范围过滤
	AlarmTimeStart *int64 // timestamp (开始时间)
	AlarmTimeEnd   *int64 // timestamp (结束时间)

	// 搜索参数（模糊匹配）
	Resident   string // 住户搜索（姓名或账号）
	BranchTag  string // 位置标签搜索
	UnitName   string // 单元名称搜索
	DeviceName string // 设备名称搜索

	// 过滤参数（多选）
	EventTypes  []string // 事件类型过滤
	Categories  []string // 类别过滤
	AlarmLevels []string // 报警级别过滤

	// 关联过滤
	CardID    string   // 按卡片ID过滤（后端会转换为 device_ids）
	DeviceIDs []string // 按设备ID过滤

	// 分页
	Page     int // 页码，默认 1
	PageSize int // 每页数量，默认 20，最大 100
}

// ListAlarmEventsResponse 查询报警事件列表响应
type ListAlarmEventsResponse struct {
	Items      []*AlarmEventDTO // 报警事件列表（包含关联数据）
	Pagination PaginationDTO    // 分页信息
}

// PaginationDTO 分页信息
type PaginationDTO struct {
	Size  int // 每页数量
	Page  int // 当前页码
	Count int // 当前页数量
	Total int // 总数量
}

// AlarmEventDTO 报警事件 DTO（包含关联数据）
type AlarmEventDTO struct {
	// 基础字段（来自 alarm_events 表）
	EventID     string `json:"event_id"`     // UUID
	TenantID    string `json:"tenant_id"`    // 租户ID
	DeviceID    string `json:"device_id"`    // 设备ID
	EventType   string `json:"event_type"`   // 事件类型
	Category    string `json:"category"`     // 类别（safety, clinical, behavioral, device）
	AlarmLevel  string `json:"alarm_level"`  // 报警级别
	AlarmStatus string `json:"alarm_status"` // 报警状态（'active', 'acknowledged', 'resolved'）
	TriggeredAt int64  `json:"triggered_at"` // timestamp（触发时间）

	// 处理信息
	HandlingState   *string `json:"handling_state,omitempty"`   // 'verified' | 'false_alarm' | 'test'（从 operation 映射）
	HandlingDetails *string `json:"handling_details,omitempty"` // 备注（从 notes 获取）
	HandlerID       *string `json:"handler_id,omitempty"`       // 处理人ID（从 handler 获取）
	HandlerName     *string `json:"handler_name,omitempty"`     // 处理人名称（通过 JOIN users 获取）
	HandledAt       *int64  `json:"handled_at,omitempty"`       // timestamp（处理时间，从 hand_time 获取）

	// 关联数据（通过 JOIN 查询）
	CardID     *string `json:"card_id,omitempty"`     // 卡片ID（通过 device_id JOIN cards 获取）
	DeviceName *string  `json:"device_name,omitempty"` // 设备名称（通过 device_id JOIN devices 获取）

	// 住户信息（通过 device → bed → resident 获取）
	ResidentID      *string `json:"resident_id,omitempty"`      // 住户ID
	ResidentName    *string `json:"resident_name,omitempty"`    // 住户名称（nickname 或 first_name + last_name）
	ResidentGender  *string `json:"resident_gender,omitempty"`  // 住户性别（从 residents.phi 获取）
	ResidentAge     *int    `json:"resident_age,omitempty"`     // 住户年龄（从 residents.phi.birth_date 计算）
	ResidentNetwork *string `json:"resident_network,omitempty"` // 住户网络（从 residents 表获取）

	// 地址信息（通过 device → unit/room/bed → locations 获取）
	BranchTag      *string `json:"branch_tag,omitempty"`      // 分支标签
	Building       *string `json:"building,omitempty"`        // 建筑名称
	Floor          *string `json:"floor,omitempty"`           // 楼层（如 "2F"）
	AreaTag        *string `json:"area_tag,omitempty"`       // 区域标签（从 units 表获取）
	UnitName       *string `json:"unit_name,omitempty"`      // 单元名称
	RoomName       *string `json:"room_name,omitempty"`      // 房间名称（从 rooms 表获取）
	BedName        *string `json:"bed_name,omitempty"`       // 床位名称（从 beds 表获取）
	AddressDisplay *string `json:"address_display,omitempty"` // 格式化地址显示（"branch_tag-Building-UnitName"）

	// JSONB 字段（解析后返回）
	TriggerData   map[string]interface{} `json:"trigger_data,omitempty"`   // 触发数据快照
	NotifiedUsers []interface{}          `json:"notified_users,omitempty"` // 通知接收者列表
	Metadata      map[string]interface{} `json:"metadata,omitempty"`       // 元数据
}

// HandleAlarmEventRequest 处理报警事件请求
type HandleAlarmEventRequest struct {
	// 必填字段
	TenantID        string // 租户ID
	EventID         string // 报警事件ID
	CurrentUserID   string // 当前用户ID（处理人）
	CurrentUserType string // 当前用户类型：'resident' | 'family' | 'staff'（用于权限检查）
	CurrentUserRole string // 当前用户角色（用于权限检查）

	// 处理参数
	AlarmStatus string // 'acknowledged' | 'resolved' - 目标状态
	HandleType  string // 'verified' | 'false_alarm' | 'test' - 处理类型（仅 resolved 时需要）
	Remarks     string // 备注（可选）
}

// HandleAlarmEventResponse 处理报警事件响应
type HandleAlarmEventResponse struct {
	Success bool `json:"success"` // 是否成功
}

// ============================================
// Service 方法实现
// ============================================

// ListAlarmEvents 查询报警事件列表
func (s *alarmEventService) ListAlarmEvents(ctx context.Context, req ListAlarmEventsRequest) (*ListAlarmEventsResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.CurrentUserRole == "" {
		return nil, fmt.Errorf("current_user_role is required")
	}

	// 分页参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建 Repository 层的过滤器
	filters := repository.AlarmEventFilters{}

	// 状态过滤
	if req.Status == "active" {
		status := "active"
		filters.AlarmStatus = &status
	} else if req.Status == "resolved" {
		// resolved 状态包括 acknowledged 和已设置 operation 的
		statuses := []string{"acknowledged"}
		filters.AlarmStatuses = statuses
		// 注意：resolved 实际上是通过 operation 不为 NULL 来判断的
		// 这里先使用 acknowledged，后续可以根据实际需求调整
	}

	// 时间范围过滤
	if req.AlarmTimeStart != nil {
		startTime := time.Unix(*req.AlarmTimeStart, 0)
		filters.StartTime = &startTime
	}
	if req.AlarmTimeEnd != nil {
		endTime := time.Unix(*req.AlarmTimeEnd, 0)
		filters.EndTime = &endTime
	}

	// 搜索参数
	if req.DeviceName != "" {
		filters.DeviceName = &req.DeviceName
	}
	if req.BranchTag != "" {
		filters.BranchTag = &req.BranchTag
	}
	// 注意：Resident 和 UnitName 搜索需要更复杂的 JOIN，暂时先不实现
	// 后续可以在 Repository 层扩展

	// 过滤参数（多选）
	if len(req.EventTypes) > 0 {
		// 注意：Repository 层目前只支持单个 EventType，需要扩展支持 EventTypes 数组
		// 暂时使用第一个
		if len(req.EventTypes) == 1 {
			filters.EventType = &req.EventTypes[0]
		}
	}
	if len(req.Categories) > 0 {
		if len(req.Categories) == 1 {
			filters.Category = &req.Categories[0]
		}
	}
	if len(req.AlarmLevels) > 0 {
		filters.AlarmLevels = req.AlarmLevels
	}

	// 关联过滤
	if req.CardID != "" {
		// 查询卡片关联的设备列表
		deviceIDs, err := s.getCardDeviceIDs(ctx, req.TenantID, req.CardID)
		if err != nil {
			s.logger.Warn("Failed to get card devices, ignoring card_id filter",
				zap.String("card_id", req.CardID),
				zap.Error(err),
			)
		} else {
			filters.DeviceIDs = deviceIDs
		}
	}
	if len(req.DeviceIDs) > 0 {
		filters.DeviceIDs = req.DeviceIDs
	}

	// 权限过滤（根据用户角色）
	// 注意：权限过滤逻辑需要在 Repository 层或 Service 层实现
	// 暂时先不实现，后续根据需求添加

	// 调用 Repository
	events, total, err := s.alarmEventsRepo.ListAlarmEvents(ctx, req.TenantID, filters, req.Page, req.PageSize)
	if err != nil {
		s.logger.Error("Failed to list alarm events",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list alarm events: %w", err)
	}

	// 数据转换：domain.AlarmEvent → AlarmEventDTO
	items := make([]*AlarmEventDTO, 0, len(events))
	for _, event := range events {
		dto, err := s.convertAlarmEventToDTO(ctx, req.TenantID, event)
		if err != nil {
			s.logger.Warn("Failed to convert alarm event to DTO, skipping",
				zap.String("event_id", event.EventID),
				zap.Error(err),
			)
			continue
		}
		items = append(items, dto)
	}

	return &ListAlarmEventsResponse{
		Items: items,
		Pagination: PaginationDTO{
			Size:  req.PageSize,
			Page:  req.Page,
			Count: len(items),
			Total: total,
		},
	}, nil
}

// HandleAlarmEvent 处理报警事件
func (s *alarmEventService) HandleAlarmEvent(ctx context.Context, req HandleAlarmEventRequest) (*HandleAlarmEventResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.EventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.CurrentUserRole == "" {
		return nil, fmt.Errorf("current_user_role is required")
	}
	if req.AlarmStatus != "acknowledged" && req.AlarmStatus != "resolved" {
		return nil, fmt.Errorf("invalid alarm_status: %s (must be 'acknowledged' or 'resolved')", req.AlarmStatus)
	}
	if req.AlarmStatus == "resolved" && req.HandleType == "" {
		return nil, fmt.Errorf("handle_type is required when alarm_status is 'resolved'")
	}

	// 查询报警事件
	event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, req.TenantID, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alarm event: %w", err)
	}

	// 权限检查（重要）
	// 如果 CurrentUserType 为空，默认为 "staff"
	userType := req.CurrentUserType
	if userType == "" {
		userType = "staff"
	}
	err = s.checkHandlePermission(ctx, req.TenantID, event.DeviceID, req.CurrentUserID, userType, req.CurrentUserRole)
	if err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// 状态转换验证
	if req.AlarmStatus == "acknowledged" {
		// 确认报警：只能从 active 状态转换
		if event.AlarmStatus != "active" {
			return nil, fmt.Errorf("can only acknowledge active alarms, current status: %s", event.AlarmStatus)
		}
		// 调用 Repository
		err = s.alarmEventsRepo.AcknowledgeAlarmEvent(ctx, req.TenantID, req.EventID, req.CurrentUserID)
		if err != nil {
			s.logger.Error("Failed to acknowledge alarm event",
				zap.String("tenant_id", req.TenantID),
				zap.String("event_id", req.EventID),
				zap.String("handler_id", req.CurrentUserID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to acknowledge alarm event: %w", err)
		}
	} else if req.AlarmStatus == "resolved" {
		// 解决报警：可以从 active 或 acknowledged 状态转换
		if event.AlarmStatus != "active" && event.AlarmStatus != "acknowledged" {
			return nil, fmt.Errorf("can only resolve active or acknowledged alarms, current status: %s", event.AlarmStatus)
		}

		// 映射 handle_type 到 operation
		operation := mapHandleTypeToOperation(req.HandleType)
		if operation == "" {
			return nil, fmt.Errorf("invalid handle_type: %s", req.HandleType)
		}

		// 更新状态为 resolved（通过设置 operation）
		var notes *string
		if req.Remarks != "" {
			notes = &req.Remarks
		}

		// 先更新 operation
		err = s.alarmEventsRepo.UpdateAlarmEventOperation(ctx, req.TenantID, req.EventID, operation, req.CurrentUserID, notes)
		if err != nil {
			s.logger.Error("Failed to update alarm event operation",
				zap.String("tenant_id", req.TenantID),
				zap.String("event_id", req.EventID),
				zap.String("operation", operation),
				zap.String("handler_id", req.CurrentUserID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to update alarm event operation: %w", err)
		}

		// 更新状态为 resolved（如果 operation 已设置，状态可以保持 acknowledged 或设置为 resolved）
		// 注意：根据业务规则，resolved 状态可能不需要单独设置，只需要设置 operation
		// 这里先不更新状态，保持 acknowledged
	}

	s.logger.Info("Alarm event handled",
		zap.String("tenant_id", req.TenantID),
		zap.String("event_id", req.EventID),
		zap.String("alarm_status", req.AlarmStatus),
		zap.String("handler_id", req.CurrentUserID),
	)

	return &HandleAlarmEventResponse{Success: true}, nil
}

// ============================================
// 辅助方法
// ============================================

// convertAlarmEventToDTO 将 domain.AlarmEvent 转换为 AlarmEventDTO
func (s *alarmEventService) convertAlarmEventToDTO(ctx context.Context, tenantID string, event *domain.AlarmEvent) (*AlarmEventDTO, error) {
	dto := &AlarmEventDTO{
		EventID:     event.EventID,
		TenantID:    event.TenantID,
		DeviceID:    event.DeviceID,
		EventType:   event.EventType,
		Category:    event.Category,
		AlarmLevel:  event.AlarmLevel,
		AlarmStatus: event.AlarmStatus,
		TriggeredAt:  event.TriggeredAt.Unix(),
	}

	// 处理信息
	if event.HandTime != nil {
		handledAt := event.HandTime.Unix()
		dto.HandledAt = &handledAt
	}
	if event.Handler != nil {
		dto.HandlerID = event.Handler
		// 查询处理人名称
		if user, err := s.usersRepo.GetUser(ctx, tenantID, *event.Handler); err == nil {
			nickname := user.Nickname.String
			if nickname != "" {
				dto.HandlerName = &nickname
			}
		}
	}
	if event.Operation != nil {
		handlingState := mapOperationToHandleType(*event.Operation)
		dto.HandlingState = &handlingState
	}
	if event.Notes != nil {
		dto.HandlingDetails = event.Notes
	}

	// JSONB 字段解析
	if len(event.TriggerData) > 0 {
		var triggerData map[string]interface{}
		if err := json.Unmarshal(event.TriggerData, &triggerData); err == nil {
			dto.TriggerData = triggerData
		}
	}
	if len(event.NotifiedUsers) > 0 {
		var notifiedUsers []interface{}
		if err := json.Unmarshal(event.NotifiedUsers, &notifiedUsers); err == nil {
			dto.NotifiedUsers = notifiedUsers
		}
	}
	if len(event.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(event.Metadata, &metadata); err == nil {
			dto.Metadata = metadata
		}
	}

	// 查询关联数据（设备、卡片、住户、地址信息）
	// 注意：这些查询可能会产生 N+1 问题，后续可以优化为批量查询
	err := s.enrichAlarmEventDTO(ctx, tenantID, event, dto)
	if err != nil {
		s.logger.Warn("Failed to enrich alarm event DTO",
			zap.String("event_id", event.EventID),
			zap.Error(err),
		)
		// 不返回错误，继续处理
	}

	return dto, nil
}

// enrichAlarmEventDTO 丰富 AlarmEventDTO 的关联数据
func (s *alarmEventService) enrichAlarmEventDTO(ctx context.Context, tenantID string, event *domain.AlarmEvent, dto *AlarmEventDTO) error {
	// 查询设备信息
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, event.DeviceID)
	if err == nil {
		if device.DeviceName != "" {
			dto.DeviceName = &device.DeviceName
		}
	} else {
		// 如果查询设备失败，继续处理其他关联数据
		device = nil
	}

	// 查询卡片信息（通过 device_id）
	cardID, err := s.getCardIDByDeviceID(ctx, tenantID, event.DeviceID)
	if err == nil && cardID != "" {
		dto.CardID = &cardID
	}

	// 查询地址信息（通过 device → bed/room → unit）
	if device != nil {
		var unitID string
		if device.BoundBedID.Valid {
			// 通过 bed 查询 unit
			bedID := device.BoundBedID.String
			unitID, err = s.getUnitIDByBedID(ctx, tenantID, bedID)
		} else if device.BoundRoomID.Valid {
			// 通过 room 查询 unit
			roomID := device.BoundRoomID.String
			unitID, err = s.getUnitIDByRoomID(ctx, tenantID, roomID)
		}

		if err == nil && unitID != "" {
			unit, err := s.unitsRepo.GetUnit(ctx, tenantID, unitID)
			if err == nil {
				if unit.UnitName != "" {
					dto.UnitName = &unit.UnitName
				}
				if unit.AreaName.Valid && unit.AreaName.String != "" {
					dto.AreaTag = &unit.AreaName.String
				}
				// 查询 location 信息（branch_tag, building, floor）
				// 注意：需要扩展 UnitsRepository 或直接查询 locations 表
			}
		}
	}

	// 查询住户信息（通过 device → bed → resident）
	// 注意：需要扩展 ResidentsRepository 或直接查询

	return nil
}

// checkHandlePermission 检查处理报警的权限
// 权限规则：
// 1. Facility 类型卡片：只有 Nurse 或 Caregiver 可以处理
// 2. Home 类型卡片：所有角色都可以处理
// 3. Caregiver/Nurse：可处理 assign-only 住户的报警
// 4. Manager：可处理 branch 住户的报警，如果 branch=null，处理 branch=null 的 unit 的住户
func (s *alarmEventService) checkHandlePermission(ctx context.Context, tenantID, deviceID, userID, userType, userRole string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// 1. 查询卡片信息（通过 device_id）
	cardID, err := s.getCardIDByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		// 如果找不到卡片，允许处理（fallback）
		return nil
	}

	// 2. 查询卡片的 unit_type
	unitType, err := s.getCardUnitType(ctx, tenantID, cardID)
	if err != nil {
		// 如果找不到 unit_type，允许处理（fallback）
		return nil
	}

	// 3. Facility 类型卡片：只有 Nurse 或 Caregiver 可以处理
	if unitType == "Facility" {
		if userRole != "Nurse" && userRole != "Caregiver" {
			return fmt.Errorf("only Nurse or Caregiver can handle alarms for Facility cards")
		}
	}

	// 4. Home 类型卡片：检查 assigned_only 和 branch_only 权限
	if unitType == "Home" {
		// 通过 device_id 获取关联的住户信息
		residentInfo, err := s.getResidentByDeviceID(ctx, tenantID, deviceID)
		if err != nil {
			// 如果设备没有关联住户，允许处理（fallback）
			return nil
		}

		// 5. Staff 角色权限检查
		if userType == "staff" && userRole != "" {
			// 5.1 Caregiver/Nurse：检查 assign-only
			if userRole == "Caregiver" || userRole == "Nurse" {
				// 检查权限配置
				perm, err := s.getResourcePermission(ctx, userRole, "residents", "R")
				if err == nil && perm.AssignedOnly {
					// 检查是否分配给该用户
					if !s.isResidentAssignedToUser(ctx, tenantID, residentInfo.ResidentID, userID) {
						return fmt.Errorf("access denied: resident not assigned to you")
					}
				}
				return nil
			}

			// 5.2 Manager：检查 branch-only
			if userRole == "Manager" {
				// 检查权限配置
				perm, err := s.getResourcePermission(ctx, userRole, "residents", "R")
				if err == nil && perm.BranchOnly {
					// 获取用户的 branch_tag
					var userBranchTag sql.NullString
					err := s.db.QueryRowContext(ctx,
						`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
						tenantID, userID,
					).Scan(&userBranchTag)
					if err == nil {
						// 检查住户的 branch_tag
						if !userBranchTag.Valid || userBranchTag.String == "" {
							// 用户 branch_tag 为 NULL：只能访问 branch_tag 为 NULL 或 '-' 的住户
							if residentInfo.BranchTag.Valid && residentInfo.BranchTag.String != "" && residentInfo.BranchTag.String != "-" {
								return fmt.Errorf("access denied: can only access residents with branch_tag NULL or '-'")
							}
						} else {
							// 用户 branch_tag 有值：只能访问匹配的 branch
							if !residentInfo.BranchTag.Valid || residentInfo.BranchTag.String != userBranchTag.String {
								return fmt.Errorf("access denied: resident belongs to different branch")
							}
						}
					}
				}
				return nil
			}
		}

		// 6. 其他角色：默认允许（SystemAdmin 等）
		return nil
	}

	return nil
}

// PermissionCheck 权限检查结果
type PermissionCheck struct {
	AssignedOnly bool // 是否仅限分配的资源
	BranchOnly   bool // 是否仅限同一 Branch 的资源
}

// getResourcePermission 查询资源权限配置
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
func (s *alarmEventService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	// 使用 SystemTenantID（全局权限配置）
	systemTenantID := "00000000-0000-0000-0000-000000000000"

	var assignedOnly, branchOnly bool
	err := s.db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(assigned_only, FALSE) as assigned_only,
			COALESCE(branch_only, FALSE) as branch_only
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		systemTenantID, roleCode, resourceType, permissionType,
	).Scan(&assignedOnly, &branchOnly)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	return &PermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// residentInfo 住户信息（用于权限检查）
type residentInfo struct {
	ResidentID string
	BranchTag  sql.NullString
	UnitID     sql.NullString
}

// getResidentByDeviceID 通过 device_id 获取关联的住户信息
// 查询路径：devices → beds → residents 或 devices → rooms → units → residents
func (s *alarmEventService) getResidentByDeviceID(ctx context.Context, tenantID, deviceID string) (*residentInfo, error) {
	// 查询设备关联的住户（优先通过 bed，其次通过 room）
	query := `
		SELECT DISTINCT
			r.resident_id::text,
			u.branch_tag,
			u.unit_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
		LEFT JOIN rooms rm ON (d.bound_room_id = rm.room_id OR b.room_id = rm.room_id)
		LEFT JOIN units u ON rm.unit_id = u.unit_id
		LEFT JOIN residents r ON (r.bed_id = b.bed_id OR r.room_id = rm.room_id OR r.unit_id = u.unit_id)
		WHERE d.tenant_id = $1::uuid
		  AND d.device_id = $2::uuid
		  AND r.resident_id IS NOT NULL
		LIMIT 1
	`

	var info residentInfo
	err := s.db.QueryRowContext(ctx, query, tenantID, deviceID).Scan(
		&info.ResidentID,
		&info.BranchTag,
		&info.UnitID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no resident found for device")
		}
		return nil, fmt.Errorf("failed to get resident by device_id: %w", err)
	}

	return &info, nil
}

// isResidentAssignedToUser 检查住户是否分配给该用户
// resident_caregivers 表通过 userList (JSONB) 存储用户ID列表
func (s *alarmEventService) isResidentAssignedToUser(ctx context.Context, tenantID, residentID, userID string) bool {
	// 查询 resident_caregivers 表的 userList 字段（JSONB 数组）
	query := `
		SELECT userList
		FROM resident_caregivers
		WHERE tenant_id = $1::uuid
		  AND resident_id = $2::uuid
		LIMIT 1
	`
	var userListJSON []byte
	err := s.db.QueryRowContext(ctx, query, tenantID, residentID).Scan(&userListJSON)
	if err != nil {
		// 如果查询失败或记录不存在，返回 false
		return false
	}

	// 解析 JSONB 数组
	var userList []string
	if err := json.Unmarshal(userListJSON, &userList); err != nil {
		// 如果解析失败，返回 false
		return false
	}

	// 检查 userID 是否在列表中
	for _, id := range userList {
		if id == userID {
			return true
		}
	}

	return false
}

// getCardDeviceIDs 查询卡片关联的设备ID列表
func (s *alarmEventService) getCardDeviceIDs(ctx context.Context, tenantID, cardID string) ([]string, error) {
	query := `
		SELECT devices
		FROM cards
		WHERE tenant_id = $1 AND card_id = $2
	`

	var devicesJSON []byte
	err := s.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&devicesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get card devices: %w", err)
	}

	// 解析 JSONB
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}

	deviceIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		if deviceID, ok := device["device_id"].(string); ok {
			deviceIDs = append(deviceIDs, deviceID)
		}
	}

	return deviceIDs, nil
}

// getCardIDByDeviceID 通过 device_id 查询 card_id
func (s *alarmEventService) getCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error) {
	query := `
		SELECT card_id::text
		FROM cards
		WHERE tenant_id = $1
		  AND devices @> $2::jsonb
		LIMIT 1
	`

	deviceJSON := fmt.Sprintf(`[{"device_id":"%s"}]`, deviceID)
	var cardID string
	err := s.db.QueryRowContext(ctx, query, tenantID, deviceJSON).Scan(&cardID)
	if err != nil {
		return "", fmt.Errorf("failed to get card by device_id: %w", err)
	}

	return cardID, nil
}

// getCardUnitType 查询卡片的 unit_type
func (s *alarmEventService) getCardUnitType(ctx context.Context, tenantID, cardID string) (string, error) {
	query := `
		SELECT 
			CASE 
				WHEN c.card_type = 'ActiveBed' AND c.bed_id IS NOT NULL THEN
					COALESCE(u.unit_type, 'Home')
				WHEN c.card_type = 'Location' AND c.unit_id IS NOT NULL THEN
					COALESCE(u.unit_type, 'Home')
				ELSE 'Home'
			END as unit_type
		FROM cards c
		LEFT JOIN beds b ON c.bed_id = b.bed_id
		LEFT JOIN rooms r ON b.room_id = r.room_id OR c.unit_id = r.unit_id
		LEFT JOIN units u ON r.unit_id = u.unit_id OR c.unit_id = u.unit_id
		WHERE c.tenant_id = $1 AND c.card_id = $2
		LIMIT 1
	`

	var unitType string
	err := s.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&unitType)
	if err != nil {
		return "", fmt.Errorf("failed to get card unit_type: %w", err)
	}

	return unitType, nil
}

// getUnitIDByBedID 通过 bed_id 查询 unit_id
func (s *alarmEventService) getUnitIDByBedID(ctx context.Context, tenantID, bedID string) (string, error) {
	query := `
		SELECT u.unit_id::text
		FROM beds b
		JOIN rooms r ON b.room_id = r.room_id
		JOIN units u ON r.unit_id = u.unit_id
		WHERE b.tenant_id = $1 AND b.bed_id = $2
		LIMIT 1
	`

	var unitID string
	err := s.db.QueryRowContext(ctx, query, tenantID, bedID).Scan(&unitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit_id by bed_id: %w", err)
	}

	return unitID, nil
}

// getUnitIDByRoomID 通过 room_id 查询 unit_id
func (s *alarmEventService) getUnitIDByRoomID(ctx context.Context, tenantID, roomID string) (string, error) {
	query := `
		SELECT unit_id::text
		FROM rooms
		WHERE tenant_id = $1 AND room_id = $2
		LIMIT 1
	`

	var unitID string
	err := s.db.QueryRowContext(ctx, query, tenantID, roomID).Scan(&unitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit_id by room_id: %w", err)
	}

	return unitID, nil
}

// mapHandleTypeToOperation 映射 handle_type 到 operation
func mapHandleTypeToOperation(handleType string) string {
	switch handleType {
	case "verified":
		return "verified_and_processed"
	case "false_alarm":
		return "false_alarm"
	case "test":
		return "test"
	default:
		return ""
	}
}

// mapOperationToHandleType 映射 operation 到 handle_type
func mapOperationToHandleType(operation string) string {
	switch operation {
	case "verified_and_processed":
		return "verified"
	case "false_alarm":
		return "false_alarm"
	case "test":
		return "test"
	default:
		return ""
	}
}

