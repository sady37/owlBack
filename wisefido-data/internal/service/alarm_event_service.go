package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	commoncard "owl-common/card"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

// ConfigPublisher 配置发布器接口
type ConfigPublisher interface {
	PublishAlarmProcessMessage(ctx context.Context, tenantID, cardID, deviceID, alarmLevel, alarmType, processType string, alarmTimestamp int64) error
	PublishAlarmDeviceMessage(ctx context.Context, tenantID, deviceID, deviceUID, settingType string, settingData map[string]interface{}) error
}

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
	redisClient     *redis.Client
	configPublisher ConfigPublisher
	logger          *zap.Logger
}

// NewAlarmEventService 创建 AlarmEventService 实例
func NewAlarmEventService(
	alarmEventsRepo repository.AlarmEventsRepository,
	devicesRepo repository.DevicesRepository,
	unitsRepo repository.UnitsRepository,
	usersRepo repository.UsersRepository,
	db *sql.DB,
	redisClient *redis.Client,
	configPublisher ConfigPublisher,
	logger *zap.Logger,
) AlarmEventService {
	return &alarmEventService{
		alarmEventsRepo: alarmEventsRepo,
		devicesRepo:     devicesRepo,
		unitsRepo:       unitsRepo,
		usersRepo:       usersRepo,
		db:              db,
		redisClient:     redisClient,
		configPublisher: configPublisher,
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
	BranchName string // 分支名称搜索
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
	DeviceName *string `json:"device_name,omitempty"` // 设备名称（通过 device_id JOIN devices 获取）

	// 住户信息（通过 device → bed → resident 获取）
	ResidentID      *string `json:"resident_id,omitempty"`      // 住户ID
	ResidentName    *string `json:"resident_name,omitempty"`    // 住户名称（nickname 或 first_name + last_name）
	ResidentGender  *string `json:"resident_gender,omitempty"`  // 住户性别（从 residents.phi 获取）
	ResidentAge     *int    `json:"resident_age,omitempty"`     // 住户年龄（从 residents.phi.birth_date 计算）
	ResidentNetwork *string `json:"resident_network,omitempty"` // 住户网络（从 residents 表获取）

	// 地址信息（通过 device → unit/room/bed → locations 获取）
	BranchID       *string `json:"branch_id,omitempty"`       // 分支ID（前端需要 ID 来选择对象）
	BranchName     *string `json:"branch_name,omitempty"`     // 分支名称（用于显示）
	BuildingID     *string `json:"building_id,omitempty"`     // 建筑ID（前端需要 ID 来选择对象）
	BuildingName   *string `json:"building_name,omitempty"`   // 建筑名称（用于显示）
	Floor          *int    `json:"floor,omitempty"`           // 楼层（数字，如 1, 2, 3，前端显示为 "1F", "2F"）
	UnitID         *string `json:"unit_id,omitempty"`         // 单元ID（前端需要 ID 来选择对象）
	UnitName       *string `json:"unit_name,omitempty"`       // 单元名称（用于显示）
	RoomID         *string `json:"room_id,omitempty"`         // 房间ID（前端需要 ID 来选择对象）
	RoomName       *string `json:"room_name,omitempty"`       // 房间名称（用于显示）
	BedID          *string `json:"bed_id,omitempty"`          // 床位ID（前端需要 ID 来选择对象）
	BedName        *string `json:"bed_name,omitempty"`        // 床位名称（用于显示）
	AddressDisplay *string `json:"address_display,omitempty"` // 格式化地址显示（"branch_name-building_name-floor-unit_name-room_name-bed_name"）

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
	CurrentUserType string // 当前用户类型：'resident' | 'staff'（注意：'family' 已被禁止，resident_contacts 不能登录系统）
	CurrentUserRole string // 当前用户角色（用于权限检查）

	// 处理参数
	AlarmStatus string // 'acknowledged' | 'resolved' - 目标状态
	HandleType  string // 'verified' | 'false_alarm' | 'test' - 处理类型（仅 resolved 时需要）
	Remarks     string // 备注（可选）
}

// HandleAlarmEventResponse 处理报警事件响应
// 仅包含前端显示 + cardagg 更新显示需要的字段
type HandleAlarmEventResponse struct {
	Success        bool   `json:"success"`                   // 处理是否成功
	CardID         string `json:"card_id,omitempty"`         // 卡片ID（cardagg用来找RealtimeData）
	DeviceID       string `json:"device_id,omitempty"`       // 设备ID
	AlarmLevel     string `json:"alarm_level,omitempty"`     // 报警级别（更新计数）
	AlarmType      string `json:"alarm_type,omitempty"`      // 报警类型（清理NowAlarm）
	AlarmTimestamp int64  `json:"alarm_timestamp,omitempty"` // 报警触发时间戳（防止旧数据覆盖）
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
	if req.BranchName != "" {
		filters.BranchTag = &req.BranchName // Repository 层仍使用 BranchTag 字段名，但实际存储的是 branch_name
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
	// 注意：权限过滤逻辑在 Service 层实现，参考 card 创建逻辑
	// 对于 Resident 用户：
	//   - shareUnit: 只能看到绑定到自己 bed 的设备（不能看到公共区域的设备）
	//   - 非 shareUnit: 该 unit 的住户能看到该 unit 下所有的 device（包括绑定到 bed 和绑定到 room 的设备）
	// 对于 Staff 用户：
	//   - Caregiver/Nurse: 通过指定的 resident → unit，可以看到该 unit 下的所有 device
	//   - Manager: 指定 Branchs 的所有报警
	//   - Admin: tenant 下的所有报警（不需要过滤）

	// 判断用户类型（通过查询 users 或 residents 表）
	var residentBedID, residentUnitID sql.NullString
	var isSharedUnit bool
	var userRole string

	// 先尝试从 users 表查询（staff 用户）
	var userAccount sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT user_account, role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
		req.TenantID, req.CurrentUserID,
	).Scan(&userAccount, &userRole)

	if err == nil && userAccount.Valid {
		// Staff 用户：根据角色进行权限过滤
		userRole = strings.TrimSpace(userRole)

		if strings.EqualFold(userRole, "Admin") {
			// Admin: tenant 下的所有报警，不需要过滤
			// 不设置 filters.DeviceIDs，允许查看所有设备
		} else if strings.EqualFold(userRole, "Manager") {
			// Manager: 指定 Branchs 的所有报警
			// 查询用户关联的 branch_id 列表
			userBranchIDs, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				s.logger.Warn("Failed to get user branch IDs for Manager, returning empty list",
					zap.String("user_id", req.CurrentUserID),
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(userBranchIDs) == 0 {
				// 用户没有关联任何 branch，返回空列表
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 查询这些 branch 下的所有 unit_ids
			allUnitIDs, err := s.getUnitIDsByBranchIDs(ctx, req.TenantID, userBranchIDs)
			if err != nil {
				s.logger.Warn("Failed to get unit IDs for Manager branches, returning empty list",
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(allUnitIDs) == 0 {
				// 没有关联的 unit，返回空列表
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 查询这些 unit 下的所有 device_ids
			allowedDeviceIDs, err := s.getDeviceIDsByUnitIDs(ctx, req.TenantID, allUnitIDs)
			if err != nil {
				s.logger.Warn("Failed to get device IDs for Manager units, returning empty list",
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(allowedDeviceIDs) == 0 {
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 将允许的 device_ids 添加到过滤器中
			if len(filters.DeviceIDs) > 0 {
				// 取交集
				allowedSet := make(map[string]bool)
				for _, id := range allowedDeviceIDs {
					allowedSet[id] = true
				}
				filteredDeviceIDs := []string{}
				for _, id := range filters.DeviceIDs {
					if allowedSet[id] {
						filteredDeviceIDs = append(filteredDeviceIDs, id)
					}
				}
				filters.DeviceIDs = filteredDeviceIDs
			} else {
				filters.DeviceIDs = allowedDeviceIDs
			}
		} else if strings.EqualFold(userRole, "Caregiver") || strings.EqualFold(userRole, "Nurse") {
			// Caregiver/Nurse: 通过指定的 resident → unit，可以看到该 unit 下的所有 device
			// 查询分配给该用户的 resident 列表
			assignedResidentIDs, err := s.getAssignedResidentIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				s.logger.Warn("Failed to get assigned resident IDs for Caregiver/Nurse, returning empty list",
					zap.String("user_id", req.CurrentUserID),
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(assignedResidentIDs) == 0 {
				// 没有分配的 resident，返回空列表
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 查询这些 resident 的 unit_id 列表
			unitIDs, err := s.getUnitIDsByResidentIDs(ctx, req.TenantID, assignedResidentIDs)
			if err != nil {
				s.logger.Warn("Failed to get unit IDs for assigned residents, returning empty list",
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(unitIDs) == 0 {
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 查询这些 unit 下的所有 device_ids（该 unit 下所有的 device，包括绑定到 bed 和绑定到 room 的）
			allowedDeviceIDs, err := s.getDeviceIDsByUnitIDs(ctx, req.TenantID, unitIDs)
			if err != nil {
				s.logger.Warn("Failed to get device IDs for Caregiver/Nurse units, returning empty list",
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(allowedDeviceIDs) == 0 {
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 将允许的 device_ids 添加到过滤器中
			if len(filters.DeviceIDs) > 0 {
				// 取交集
				allowedSet := make(map[string]bool)
				for _, id := range allowedDeviceIDs {
					allowedSet[id] = true
				}
				filteredDeviceIDs := []string{}
				for _, id := range filters.DeviceIDs {
					if allowedSet[id] {
						filteredDeviceIDs = append(filteredDeviceIDs, id)
					}
				}
				filters.DeviceIDs = filteredDeviceIDs
			} else {
				filters.DeviceIDs = allowedDeviceIDs
			}
		} else {
			// 其他 Staff 角色（如 IT）：暂时不进行权限过滤，允许查看所有报警
			// 或者可以根据需要添加其他角色的权限逻辑
		}
	} else {
		// 尝试从 residents 表查询（resident 用户）
		err = s.db.QueryRowContext(ctx,
			`SELECT bed_id, unit_id 
			 FROM residents 
			 WHERE tenant_id = $1 AND resident_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		).Scan(&residentBedID, &residentUnitID)

		if err == nil {
			// Resident 用户：查询 unit 的 is_shared_unit
			if residentUnitID.Valid {
				err = s.db.QueryRowContext(ctx,
					`SELECT is_shared_unit 
					 FROM units 
					 WHERE tenant_id = $1 AND unit_id = $2`,
					req.TenantID, residentUnitID.String,
				).Scan(&isSharedUnit)
				if err != nil {
					s.logger.Warn("Failed to get unit is_shared_unit, defaulting to false",
						zap.String("unit_id", residentUnitID.String),
						zap.Error(err),
					)
					isSharedUnit = false
				}
			}

			// 查询该 resident 有权限查看的设备
			allowedDeviceIDs, err := s.getDeviceIDsForResident(ctx, req.TenantID, req.CurrentUserID, residentBedID, residentUnitID, isSharedUnit)
			if err != nil {
				s.logger.Warn("Failed to get device IDs for resident, returning empty list",
					zap.String("resident_id", req.CurrentUserID),
					zap.Error(err),
				)
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			if len(allowedDeviceIDs) == 0 {
				// 如果没有关联的设备，返回空列表
				return &ListAlarmEventsResponse{
					Items: []*AlarmEventDTO{},
					Pagination: PaginationDTO{
						Size:  req.PageSize,
						Page:  req.Page,
						Count: 0,
						Total: 0,
					},
				}, nil
			}

			// 将允许的 device_ids 添加到过滤器中
			if len(filters.DeviceIDs) > 0 {
				// 取交集：只保留既在 filters.DeviceIDs 中，又在 allowedDeviceIDs 中的设备
				allowedSet := make(map[string]bool)
				for _, id := range allowedDeviceIDs {
					allowedSet[id] = true
				}
				filteredDeviceIDs := []string{}
				for _, id := range filters.DeviceIDs {
					if allowedSet[id] {
						filteredDeviceIDs = append(filteredDeviceIDs, id)
					}
				}
				filters.DeviceIDs = filteredDeviceIDs
			} else {
				filters.DeviceIDs = allowedDeviceIDs
			}
		} else {
			// 既不是 staff 也不是 resident，返回空列表
			s.logger.Warn("User not found in users or residents table",
				zap.String("user_id", req.CurrentUserID),
			)
			return &ListAlarmEventsResponse{
				Items: []*AlarmEventDTO{},
				Pagination: PaginationDTO{
					Size:  req.PageSize,
					Page:  req.Page,
					Count: 0,
					Total: 0,
				},
			}, nil
		}
	}

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

	// 更新相关卡片的报警计数（因为报警状态已改变）
	cardID, err := s.getCardIDByDeviceID(ctx, req.TenantID, event.DeviceID)
	if err != nil {
		s.logger.Warn("Failed to get card ID for alarm event, will publish with empty card_id",
			zap.String("device_id", event.DeviceID),
			zap.String("event_id", req.EventID),
			zap.Error(err),
		)
		// 继续处理，cardID 为空
		cardID = ""
	} else if cardID != "" {
		if err := s.updateCardAlarmCounts(ctx, req.TenantID, cardID); err != nil {
			s.logger.Warn("Failed to update card alarm counts after handling alarm",
				zap.String("card_id", cardID),
				zap.String("event_id", req.EventID),
				zap.Error(err),
			)
			// 不返回错误，报警处理已成功
		} else {
			s.logger.Debug("Updated card alarm counts after handling alarm",
				zap.String("card_id", cardID),
				zap.String("event_id", req.EventID),
			)
		}
	}

	// 构建响应：包含前端和cardagg都需要的信息
	response := &HandleAlarmEventResponse{
		Success:        true,
		CardID:         cardID,
		DeviceID:       event.DeviceID,
		AlarmLevel:     event.AlarmLevel,
		AlarmType:      event.EventType, // 使用 EventType 作为 AlarmType
		AlarmTimestamp: event.TriggeredAt.Unix(),
	}

	// 异步发布 alarmProcess 消息到 config:alarmProcess:stream（供cardagg消费）
	go func() {
		pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.configPublisher.PublishAlarmProcessMessage(
			pubCtx,
			req.TenantID,
			cardID,
			event.DeviceID,
			event.AlarmLevel,
			event.EventType,
			rediscommon.AlarmProcessActionAck, // 使用常量：ack（用户已确认报警）
			event.TriggeredAt.Unix(),
		); err != nil {
			s.logger.Warn("Failed to publish alarm process message",
				zap.String("event_id", req.EventID),
				zap.String("card_id", cardID),
				zap.Error(err),
			)
		}
	}()

	return response, nil
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
		TriggeredAt: event.TriggeredAt.Unix(),
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
			dto.UnitID = &unitID
			unit, err := s.unitsRepo.GetUnit(ctx, tenantID, unitID)
			if err == nil {
				if unit.UnitName != "" {
					dto.UnitName = &unit.UnitName
				}
				// 设置 branch_id 和 branch_name
				if unit.BranchID.Valid && unit.BranchID.String != "" {
					dto.BranchID = &unit.BranchID.String
					// 查询 branch_name
					if unit.BranchName.Valid && unit.BranchName.String != "" {
						dto.BranchName = &unit.BranchName.String
					}
				}
				// 设置 building_name（从 unit.BuildingName 获取）
				if unit.BuildingName.Valid && unit.BuildingName.String != "" {
					dto.BuildingName = &unit.BuildingName.String
					// 注意：building_id 需要从 buildings 表查询，暂时不设置
				}
				// 设置 floor（从 string "1F" 转换为 number 1）
				if unit.Floor.Valid && unit.Floor.String != "" {
					// 提取数字部分（如 "1F" -> 1, "2F" -> 2）
					floorStr := unit.Floor.String
					// 移除 "F" 后缀并提取数字
					floorNum := 0
					if len(floorStr) > 0 {
						// 尝试解析数字（支持 "1F", "1", "2F" 等格式）
						var n int
						if _, err := fmt.Sscanf(floorStr, "%d", &n); err == nil {
							floorNum = n
						} else if len(floorStr) > 1 && floorStr[len(floorStr)-1] == 'F' {
							// 处理 "1F" 格式
							if _, err := fmt.Sscanf(floorStr[:len(floorStr)-1], "%d", &n); err == nil {
								floorNum = n
							}
						}
					}
					if floorNum > 0 {
						dto.Floor = &floorNum
					}
				}
			}
		}

		// 设置 room_id 和 bed_id
		if device.BoundRoomID.Valid {
			roomID := device.BoundRoomID.String
			dto.RoomID = &roomID
		}
		if device.BoundBedID.Valid {
			bedID := device.BoundBedID.String
			dto.BedID = &bedID
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
					// 获取用户的 branch_id 列表（从 user_branches 表）
					rows, err := s.db.QueryContext(ctx,
						`SELECT branch_id::text FROM user_branches WHERE tenant_id = $1 AND user_id::text = $2`,
						tenantID, userID,
					)
					if err == nil {
						defer rows.Close()
						userBranchIDs := make(map[string]bool)
						for rows.Next() {
							var branchID sql.NullString
							if err := rows.Scan(&branchID); err == nil && branchID.Valid {
								userBranchIDs[branchID.String] = true
							}
						}

						// 获取住户的 branch_id（通过 unit_id → units.branch_id）
						if residentInfo.UnitID.Valid && residentInfo.UnitID.String != "" {
							var residentBranchID sql.NullString
							err := s.db.QueryRowContext(ctx,
								`SELECT branch_id::text FROM units WHERE tenant_id = $1 AND unit_id = $2`,
								tenantID, residentInfo.UnitID.String,
							).Scan(&residentBranchID)
							if err == nil {
								if len(userBranchIDs) == 0 {
									// 用户没有关联任何 branch：只能访问 branch_id 为 NULL 的住户
									if residentBranchID.Valid && residentBranchID.String != "" {
										return fmt.Errorf("access denied: can only access residents with null branch_id")
									}
								} else {
									// 用户有关联 branch：只能访问匹配的 branch
									if !residentBranchID.Valid || !userBranchIDs[residentBranchID.String] {
										return fmt.Errorf("access denied: resident belongs to different branch")
									}
								}
							}
						} else {
							// 住户没有 unit_id：如果用户有 branch 限制，不允许访问
							if len(userBranchIDs) > 0 {
								return fmt.Errorf("access denied: resident has no unit_id")
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
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func (s *alarmEventService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	// 使用 SystemTenantID（全局权限配置）
	systemTenantID := "00000000-0000-0000-0000-000000000001"

	var permissionScope string
	err := s.db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		systemTenantID, roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
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

	return &PermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// residentInfo 住户信息（用于权限检查）
type residentInfo struct {
	ResidentID string
	BranchName sql.NullString // 从 units.branch_name 获取（通过 JOIN branches 表）
	UnitID     sql.NullString
}

// getResidentByDeviceID 通过 device_id 获取关联的住户信息
// 查询路径：devices → beds → residents 或 devices → rooms → units → residents
func (s *alarmEventService) getResidentByDeviceID(ctx context.Context, tenantID, deviceID string) (*residentInfo, error) {
	// 查询设备关联的住户（优先通过 bed，其次通过 room）
	query := `
		SELECT DISTINCT
			r.resident_id::text,
			u.branch_name,
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
		&info.BranchName,
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
	// 查询 resident_caregivers 表的 user_list 字段（JSONB 数组）
	query := `
		SELECT user_list
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

// getDeviceIDsForResident 查询住户有权限查看的设备 ID 列表
// 优先通过 card 表查找，失败时回退到直接查询 devices 表
// 参考 card 创建逻辑：
//   - shareUnit: 只能看到绑定到自己 bed 的设备（不能看到公共区域的设备）
//   - 非 shareUnit: 可以看到绑定到自己 bed 的设备，以及同一 unit 下未绑定到 bed 的设备（绑定到 room 的设备）
func (s *alarmEventService) getDeviceIDsForResident(ctx context.Context, tenantID, residentID string, bedID, unitID sql.NullString, isSharedUnit bool) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 优先尝试通过 card 表查找
	deviceIDs, err := s.getDeviceIDsFromCardsForResident(ctx, tenantID, bedID, unitID, isSharedUnit)
	if err == nil && len(deviceIDs) > 0 {
		// 成功从 card 表获取，直接返回
		return deviceIDs, nil
	}

	// card 表查找失败，回退到直接查询 devices 表
	if err != nil {
		s.logger.Warn("Failed to get device IDs from cards, falling back to direct query",
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
	}

	var query string
	var args []interface{}

	if isSharedUnit {
		// shareUnit: 只能看到绑定到自己 bed 的设备
		if !bedID.Valid {
			// 如果没有 bed_id，返回空列表
			return []string{}, nil
		}
		query = `
			SELECT DISTINCT d.device_id::text
			FROM devices d
			WHERE d.tenant_id = $1
			  AND d.bound_bed_id = $2
			  AND d.status <> 'disabled'
		`
		args = []interface{}{tenantID, bedID.String}
	} else {
		// 非 shareUnit: 该 unit 的住户能看到该 unit 下所有的 device（包括绑定到 bed 和绑定到 room 的设备）
		if !unitID.Valid {
			// 如果没有 unit_id，返回空列表
			return []string{}, nil
		}

		// 查询该 unit 下所有的 device（无论绑定到 bed 还是 room）
		query = `
				SELECT DISTINCT d.device_id::text
				FROM devices d
				LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND b.tenant_id = $1
				LEFT JOIN rooms r1 ON b.room_id = r1.room_id AND r1.tenant_id = $1
				LEFT JOIN rooms r2 ON d.bound_room_id = r2.room_id AND r2.tenant_id = $1
				WHERE d.tenant_id = $1
				  AND (
					-- 路径1: 设备绑定到 bed，bed 属于该 unit
					(d.bound_bed_id IS NOT NULL AND r1.unit_id = $2)
					-- 路径2: 设备绑定到 room，room 属于该 unit
					OR (d.bound_room_id IS NOT NULL AND r2.unit_id = $2)
				  )
				  AND d.status <> 'disabled'
			`
		args = []interface{}{tenantID, unitID.String}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query device IDs for resident: %w", err)
	}
	defer rows.Close()

	deviceIDs = []string{}
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device IDs: %w", err)
	}

	return deviceIDs, nil
}

// getDeviceIDsFromCardsForResident 通过 card 表查询住户有权限查看的设备 ID 列表
func (s *alarmEventService) getDeviceIDsFromCardsForResident(ctx context.Context, tenantID string, bedID, unitID sql.NullString, isSharedUnit bool) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var query string
	var args []interface{}

	if isSharedUnit {
		// shareUnit: 通过 bed_id 查找 ActiveBed card
		if !bedID.Valid {
			return []string{}, nil
		}
		query = `
			SELECT devices
			FROM cards
			WHERE tenant_id = $1
			  AND card_type = 'ActiveBedCard'
			  AND bed_id = $2
			LIMIT 1
		`
		args = []interface{}{tenantID, bedID.String}
	} else {
		// 非 shareUnit: 查找该 unit 下的所有 card（ActiveBedCard 和 UnitCard）
		if !unitID.Valid {
			return []string{}, nil
		}
		query = `
			SELECT devices
			FROM cards
			WHERE tenant_id = $1
			  AND (
				(card_type = 'ActiveBedCard' AND bed_id IN (
					SELECT bed_id FROM beds WHERE room_id IN (
						SELECT room_id FROM rooms WHERE unit_id = $2
					)
				))
				OR (card_type = 'UnitCard' AND unit_id = $2)
			  )
		`
		args = []interface{}{tenantID, unitID.String}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	deviceIDSet := make(map[string]bool)
	for rows.Next() {
		var devicesJSON []byte
		if err := rows.Scan(&devicesJSON); err != nil {
			continue
		}

		// 解析 JSONB 数组
		var devices []map[string]interface{}
		if err := json.Unmarshal(devicesJSON, &devices); err != nil {
			continue
		}

		// 提取 device_id
		for _, device := range devices {
			if deviceID, ok := device["device_id"].(string); ok {
				deviceIDSet[deviceID] = true
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	// 转换为数组
	deviceIDs := make([]string, 0, len(deviceIDSet))
	for deviceID := range deviceIDSet {
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, nil
}

// getUserBranchIDs 查询用户所属的 branch_id 列表
func (s *alarmEventService) getUserBranchIDs(ctx context.Context, tenantID, userID string) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT branch_id::text 
		 FROM user_branches 
		 WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user branch IDs: %w", err)
	}
	defer rows.Close()

	branchIDs := []string{}
	for rows.Next() {
		var branchID sql.NullString
		if err := rows.Scan(&branchID); err == nil && branchID.Valid {
			branchIDs = append(branchIDs, branchID.String)
		}
	}

	return branchIDs, rows.Err()
}

// getUnitIDsByBranchIDs 通过多个 branch_id 查询所有关联的 unit_ids
func (s *alarmEventService) getUnitIDsByBranchIDs(ctx context.Context, tenantID string, branchIDs []string) ([]string, error) {
	if len(branchIDs) == 0 {
		return []string{}, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 构建 IN 查询的占位符
	placeholders := make([]string, len(branchIDs))
	args := make([]interface{}, len(branchIDs)+1)
	args[0] = tenantID
	for i, branchID := range branchIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = branchID
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT unit_id::text
		FROM units
		WHERE tenant_id = $1 AND branch_id::text IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query unit IDs by branch IDs: %w", err)
	}
	defer rows.Close()

	unitIDs := []string{}
	for rows.Next() {
		var unitID string
		if err := rows.Scan(&unitID); err != nil {
			return nil, fmt.Errorf("failed to scan unit ID: %w", err)
		}
		unitIDs = append(unitIDs, unitID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unit IDs: %w", err)
	}

	return unitIDs, nil
}

// getDeviceIDsByUnitIDs 通过多个 unit_id 查询关联的所有 device_ids
// 优先通过 card 表查找，失败时回退到直接查询 devices 表
// 查询路径：units → cards → devices（优先）或 units → rooms → devices（回退）
// 返回该 unit 下所有的 device（包括绑定到 bed 和绑定到 room 的设备）
func (s *alarmEventService) getDeviceIDsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]string, error) {
	if len(unitIDs) == 0 {
		return []string{}, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 优先尝试通过 card 表查找
	deviceIDs, err := s.getDeviceIDsFromCardsByUnitIDs(ctx, tenantID, unitIDs)
	if err == nil && len(deviceIDs) > 0 {
		// 成功从 card 表获取，直接返回
		return deviceIDs, nil
	}

	// card 表查找失败，回退到直接查询 devices 表
	if err != nil {
		s.logger.Warn("Failed to get device IDs from cards, falling back to direct query",
			zap.Strings("unit_ids", unitIDs),
			zap.Error(err),
		)
	}

	// 构建 IN 查询的占位符
	placeholders := make([]string, len(unitIDs))
	args := make([]interface{}, len(unitIDs)+1)
	args[0] = tenantID
	for i, unitID := range unitIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = unitID
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT d.device_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND b.tenant_id = $1
		LEFT JOIN rooms r1 ON b.room_id = r1.room_id AND r1.tenant_id = $1
		LEFT JOIN rooms r2 ON d.bound_room_id = r2.room_id AND r2.tenant_id = $1
		WHERE d.tenant_id = $1
		  AND (
			-- 路径1: 设备绑定到 bed，bed 属于这些 unit
			(d.bound_bed_id IS NOT NULL AND r1.unit_id::text IN (%s))
			-- 路径2: 设备绑定到 room，room 属于这些 unit
			OR (d.bound_room_id IS NOT NULL AND r2.unit_id::text IN (%s))
		  )
		  AND d.status <> 'disabled'
	`, strings.Join(placeholders, ", "), strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query device IDs by unit IDs: %w", err)
	}
	defer rows.Close()

	deviceIDs = []string{}
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device IDs: %w", err)
	}

	return deviceIDs, nil
}

// getDeviceIDsFromCardsByUnitIDs 通过 card 表查询多个 unit_id 关联的所有 device_ids
func (s *alarmEventService) getDeviceIDsFromCardsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]string, error) {
	if len(unitIDs) == 0 {
		return []string{}, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 构建 IN 查询的占位符
	placeholders := make([]string, len(unitIDs))
	args := make([]interface{}, len(unitIDs)+1)
	args[0] = tenantID
	for i, unitID := range unitIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = unitID
	}

	// 查询这些 unit 下的所有 card（ActiveBedCard 和 UnitCard）
	query := fmt.Sprintf(`
		SELECT devices
		FROM cards
		WHERE tenant_id = $1
		  AND (
			-- UnitCard : 直接通过 unit_id 匹配
			(card_type = 'UnitCard' AND unit_id::text IN (%s))
			-- ActiveBed card: 通过 bed → room → unit 匹配
			OR (card_type = 'ActiveBed' AND bed_id IN (
				SELECT bed_id FROM beds WHERE room_id IN (
					SELECT room_id FROM rooms WHERE unit_id::text IN (%s)
				)
			))
		  )
	`, strings.Join(placeholders, ", "), strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards by unit IDs: %w", err)
	}
	defer rows.Close()

	deviceIDSet := make(map[string]bool)
	for rows.Next() {
		var devicesJSON []byte
		if err := rows.Scan(&devicesJSON); err != nil {
			continue
		}

		// 解析 JSONB 数组
		var devices []map[string]interface{}
		if err := json.Unmarshal(devicesJSON, &devices); err != nil {
			continue
		}

		// 提取 device_id
		for _, device := range devices {
			if deviceID, ok := device["device_id"].(string); ok {
				deviceIDSet[deviceID] = true
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	// 转换为数组
	deviceIDs := make([]string, 0, len(deviceIDSet))
	for deviceID := range deviceIDSet {
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, nil
}

// getAssignedResidentIDs 查询分配给指定用户的住户 ID 列表
// 通过 resident_caregivers 表的 user_list 字段查询
func (s *alarmEventService) getAssignedResidentIDs(ctx context.Context, tenantID, userID string) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := `
		SELECT DISTINCT resident_id::text
		FROM resident_caregivers
		WHERE tenant_id = $1
		  AND (user_list::text LIKE $2 OR user_list::text LIKE $3)
	`
	pattern1 := fmt.Sprintf("%%\"%s\"%%", userID)
	pattern2 := fmt.Sprintf("%%\"%s\"%%", userID)

	rows, err := s.db.QueryContext(ctx, query, tenantID, pattern1, pattern2)
	if err != nil {
		return nil, fmt.Errorf("failed to query assigned resident IDs: %w", err)
	}
	defer rows.Close()

	residentIDs := []string{}
	for rows.Next() {
		var residentID string
		if err := rows.Scan(&residentID); err != nil {
			return nil, fmt.Errorf("failed to scan resident ID: %w", err)
		}
		residentIDs = append(residentIDs, residentID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating resident IDs: %w", err)
	}

	return residentIDs, nil
}

// getUnitIDsByResidentIDs 通过多个 resident_id 查询关联的 unit_id 列表
func (s *alarmEventService) getUnitIDsByResidentIDs(ctx context.Context, tenantID string, residentIDs []string) ([]string, error) {
	if len(residentIDs) == 0 {
		return []string{}, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 构建 IN 查询的占位符
	placeholders := make([]string, len(residentIDs))
	args := make([]interface{}, len(residentIDs)+1)
	args[0] = tenantID
	for i, residentID := range residentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = residentID
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT unit_id::text
		FROM residents
		WHERE tenant_id = $1 
		  AND resident_id::text IN (%s)
		  AND unit_id IS NOT NULL
	`, strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query unit IDs by resident IDs: %w", err)
	}
	defer rows.Close()

	unitIDs := []string{}
	for rows.Next() {
		var unitID sql.NullString
		if err := rows.Scan(&unitID); err != nil {
			return nil, fmt.Errorf("failed to scan unit ID: %w", err)
		}
		if unitID.Valid {
			unitIDs = append(unitIDs, unitID.String)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unit IDs: %w", err)
	}

	return unitIDs, nil
}

// getCardIDByDeviceID 通过 device_id 查询 card_id（从Redis缓存获取）
func (s *alarmEventService) getCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error) {
	// 扫描所有静态卡片缓存 vital-focus:card:*:static
	const staticKeyPrefix = "vital-focus:card:"
	const staticKeySuffix = ":static"

	// 使用redis scan扫描key
	var keys []string
	iter := s.redisClient.Scan(ctx, 0, staticKeyPrefix+"*"+staticKeySuffix, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		s.logger.Warn("scan redis keys failed", zap.Error(err))
		// fallback to DB查询
		return s.getCardIDByDeviceIDFromDB(ctx, tenantID, deviceID)
	}

	// 遍历所有卡片缓存查找device_id
	for _, key := range keys {
		raw, err := s.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			continue
		}

		var card commoncard.VitalFocusCardInfo
		if err := json.Unmarshal([]byte(raw), &card); err != nil {
			continue
		}

		// 检查tenant
		if card.TenantID != tenantID {
			continue
		}

		// 遍历devices查找匹配的deviceID
		for _, dev := range card.Devices {
			if dev.DeviceID == deviceID {
				return card.CardID, nil
			}
		}
	}

	// 缓存miss，fallback to DB
	return s.getCardIDByDeviceIDFromDB(ctx, tenantID, deviceID)
}

// getCardIDByDeviceIDFromDB 从数据库查询 card_id（fallback）
func (s *alarmEventService) getCardIDByDeviceIDFromDB(ctx context.Context, tenantID, deviceID string) (string, error) {
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
				WHEN c.card_type = 'ActiveBedCard' AND c.bed_id IS NOT NULL THEN
					COALESCE(u.unit_type, 'Home')
				WHEN c.card_type = 'UnitCard' AND c.unit_id IS NOT NULL THEN
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

// updateCardAlarmCounts 更新卡片的未处理报警计数
func (s *alarmEventService) updateCardAlarmCounts(ctx context.Context, tenantID, cardID string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// 1. 查询卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
	query := `
		SELECT devices
		FROM cards
		WHERE tenant_id = $1 AND card_id = $2
	`

	var devicesJSON []byte
	err := s.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&devicesJSON)
	if err != nil {
		return fmt.Errorf("failed to get card devices: %w", err)
	}

	// 2. 解析 devices JSONB，提取 device_id 列表
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}

	if len(devices) == 0 {
		// 没有设备，将所有计数设为 0
		return s.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
	}

	// 提取 device_id 列表
	deviceIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		if deviceID, ok := device["device_id"].(string); ok && deviceID != "" {
			deviceIDs = append(deviceIDs, deviceID)
		}
	}

	if len(deviceIDs) == 0 {
		// 没有有效的 device_id，将所有计数设为 0
		return s.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
	}

	// 3. 统计这些设备的未处理报警（alarm_status = 'active'）
	// 按 alarm_level 分组统计（映射到 0-4）
	// 构建 IN 查询的占位符
	placeholders := make([]string, len(deviceIDs))
	args := make([]interface{}, len(deviceIDs)+1)
	args[0] = tenantID
	for i, deviceID := range deviceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = deviceID
	}

	countQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) FILTER (WHERE alarm_level IN ('0', 'EMERG')) as count_0,
			COUNT(*) FILTER (WHERE alarm_level IN ('1', 'ALERT')) as count_1,
			COUNT(*) FILTER (WHERE alarm_level IN ('2', 'CRIT')) as count_2,
			COUNT(*) FILTER (WHERE alarm_level IN ('3', 'ERR')) as count_3,
			COUNT(*) FILTER (WHERE alarm_level IN ('4', 'WARNING')) as count_4
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id::text IN (%s)
		  AND alarm_status = 'active'
		  AND alarm_level IN ('0', '1', '2', '3', '4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')
		  AND (metadata->>'deleted_at' IS NULL)
	`, strings.Join(placeholders, ", "))

	var count0, count1, count2, count3, count4 int
	err = s.db.QueryRowContext(ctx, countQuery, args...).Scan(
		&count0, &count1, &count2, &count3, &count4,
	)
	if err != nil {
		return fmt.Errorf("failed to count alarm events: %w", err)
	}

	// 4. 更新 cards 表的 unhandled_alarm_0 到 unhandled_alarm_4 字段
	updateQuery := `
		UPDATE cards
		SET 
			unhandled_alarm_0 = $3,
			unhandled_alarm_1 = $4,
			unhandled_alarm_2 = $5,
			unhandled_alarm_3 = $6,
			unhandled_alarm_4 = $7
		WHERE tenant_id = $1 AND card_id = $2
	`

	_, err = s.db.ExecContext(ctx, updateQuery, tenantID, cardID, count0, count1, count2, count3, count4)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts: %w", err)
	}

	return nil
}

// updateCardAlarmCountsToZero 将卡片的报警计数全部设为 0
func (s *alarmEventService) updateCardAlarmCountsToZero(ctx context.Context, tenantID, cardID string) error {
	updateQuery := `
		UPDATE cards
		SET 
			unhandled_alarm_0 = 0,
			unhandled_alarm_1 = 0,
			unhandled_alarm_2 = 0,
			unhandled_alarm_3 = 0,
			unhandled_alarm_4 = 0
		WHERE tenant_id = $1 AND card_id = $2
	`

	_, err := s.db.ExecContext(ctx, updateQuery, tenantID, cardID)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts to zero: %w", err)
	}

	return nil
}
