package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
	"github.com/lib/pq"
)

// CardService 卡片服务接口
type CardService interface {
	// GetCardOverview 获取卡片概览列表（返回所有可见的卡片）
	GetCardOverview(ctx context.Context, req GetCardOverviewRequest) (*GetCardOverviewResponse, error)
}

// cardService 卡片服务实现
type cardService struct {
	cardsRepo     repository.CardsRepository
	residentsRepo repository.ResidentsRepository
	devicesRepo   repository.DevicesRepository
	usersRepo     repository.UsersRepository
	db            *sql.DB // 用于复杂查询（批量查询直接使用 db）
	logger        *zap.Logger
}

// NewCardService 创建卡片服务
func NewCardService(
	cardsRepo repository.CardsRepository,
	residentsRepo repository.ResidentsRepository,
	devicesRepo repository.DevicesRepository,
	usersRepo repository.UsersRepository,
	db *sql.DB,
	logger *zap.Logger,
) CardService {
	return &cardService{
		cardsRepo:     cardsRepo,
		residentsRepo: residentsRepo,
		devicesRepo:   devicesRepo,
		usersRepo:     usersRepo,
		db:            db,
		logger:        logger,
	}
}

// GetCardOverviewRequest 获取卡片概览请求
type GetCardOverviewRequest struct {
	TenantID        string
	CardID          string // 可选：查询单个卡片
	Search          string // 搜索关键词
	CardType        string // "ActiveBed" | "Unit"
	UnitType        string // "Home" | "Facility"
	IsPublicSpace   *bool
	IsMultiPersonRoom *bool
	Sort            string // "card_name" | "card_address"
	Direction       string // "asc" | "desc"

	// 权限相关
	CurrentUserID   string
	CurrentUserType string // "resident" | "family" | "staff"
	CurrentUserRole string // "Nurse" | "Caregiver" | "Manager" | "SystemAdmin"
}

// GetCardOverviewResponse 获取卡片概览响应
type GetCardOverviewResponse struct {
	Items []*domain.CardOverviewItem // 所有可见的卡片（前端负责分页）
	Total int                        // 总数
}

// GetCardOverview 获取卡片概览列表
func (s *cardService) GetCardOverview(ctx context.Context, req GetCardOverviewRequest) (*GetCardOverviewResponse, error) {
	// 1. 构建 Repository 请求
	repoReq := repository.ListCardsRequest{
		TenantID:        req.TenantID,
		CardID:          req.CardID,
		Search:          req.Search,
		CardType:        req.CardType,
		UnitType:        req.UnitType,
		IsPublicSpace:   req.IsPublicSpace,
		IsMultiPersonRoom: req.IsMultiPersonRoom,
		Sort:            req.Sort,
		Direction:       req.Direction,
	}

	// 2. 处理 Family 用户类型
	if req.CurrentUserType == "family" {
		residentID, err := s.getResidentIDByContactID(ctx, req.TenantID, req.CurrentUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resident_id for contact: %w", err)
		}
		repoReq.PermissionFilter = &repository.PermissionFilter{
			UserID:   residentID,
			UserType: "resident",
		}
	} else if req.CurrentUserType == "resident" {
		repoReq.PermissionFilter = &repository.PermissionFilter{
			UserID:   req.CurrentUserID,
			UserType: "resident",
		}
	} else if req.CurrentUserType == "staff" {
		// Staff：检查权限配置
		perm, err := s.getResourcePermission(ctx, req.CurrentUserRole, "cards", "R")
		if err != nil {
			return nil, fmt.Errorf("failed to get resource permission: %w", err)
		}

		repoReq.PermissionFilter = &repository.PermissionFilter{}

		if perm.BranchOnly {
			// BranchOnly：在 SQL 中过滤
			user, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user: %w", err)
			}
			if user.BranchTag.Valid {
				branchTag := user.BranchTag.String
				repoReq.PermissionFilter.UserBranchTag = &branchTag
			} else {
				emptyTag := ""
				repoReq.PermissionFilter.UserBranchTag = &emptyTag
			}
		}

		if perm.AssignedOnly {
			// AssignedOnly：在 SQL 中过滤（使用 CTE）
			repoReq.PermissionFilter.AssignedOnly = true
			repoReq.PermissionFilter.UserIDForAssignment = req.CurrentUserID
		}
	}

	// 3. Repository 查询（返回所有可见的卡片，不分页）
	cards, err := s.cardsRepo.ListCards(ctx, repoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}

	// 4. 数据聚合（devices, residents）
	allCards, err := s.aggregateCardData(ctx, cards)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate card data: %w", err)
	}

	// 5. 计算 family_view 并规范化 card_type（数据库使用 'Location'，API 返回 'Unit'）
	for _, card := range allCards {
		// 规范化 card_type：数据库中的 'Location' 转换为 API 的 'Unit'
		if card.CardType == "Location" {
			card.CardType = "Unit"
		}

		familyView, err := s.getFamilyView(ctx, card)
		if err != nil {
			s.logger.Warn("Failed to get family_view, using default",
				zap.Error(err),
				zap.String("card_id", card.CardID),
			)
			card.FamilyView = false // 默认值
		} else {
			card.FamilyView = familyView
		}
	}

	// 6. 返回所有卡片（前端负责分页）
	return &GetCardOverviewResponse{
		Items: allCards,
		Total: len(allCards),
	}, nil
}

// getResidentIDByContactID 根据 contact_id 获取 resident_id
func (s *cardService) getResidentIDByContactID(ctx context.Context, tenantID, contactID string) (string, error) {
	var residentID string
	err := s.db.QueryRowContext(ctx,
		`SELECT resident_id::text 
		 FROM resident_contacts 
		 WHERE tenant_id = $1 AND contact_id::text = $2`,
		tenantID, contactID,
	).Scan(&residentID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("contact not found: %s", contactID)
		}
		return "", fmt.Errorf("failed to query resident_id: %w", err)
	}

	return residentID, nil
}

// getResourcePermission 查询资源权限配置
func (s *cardService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	var perm PermissionCheck
	err := s.db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(assigned_only, FALSE) as assigned_only,
			COALESCE(branch_only, FALSE) as branch_only
		 FROM role_permissions
		 WHERE tenant_id = '00000000-0000-0000-0000-000000000000'
		   AND role_code = $1
		   AND resource_type = $2
		   AND permission_type = $3`,
		roleCode, resourceType, permissionType,
	).Scan(&perm.AssignedOnly, &perm.BranchOnly)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有权限配置，返回默认值（无限制）
			return &PermissionCheck{
				AssignedOnly: false,
				BranchOnly:   false,
			}, nil
		}
		return nil, fmt.Errorf("failed to query resource permission: %w", err)
	}

	return &perm, nil
}

// PermissionCheck 权限检查结果（已在 alarm_event_service.go 中定义，这里不再重复定义）

// aggregateCardData 聚合卡片数据（devices, residents）
func (s *cardService) aggregateCardData(ctx context.Context, cards []*domain.CardWithUnitInfo) ([]*domain.CardOverviewItem, error) {
	if len(cards) == 0 {
		return []*domain.CardOverviewItem{}, nil
	}

	// 1. 收集所有需要查询的 ID
	deviceIDs := make(map[string]bool)
	residentIDs := make(map[string]bool)

	for _, card := range cards {
		// 收集设备 ID
		var deviceIDsFromCard []string
		if err := json.Unmarshal(card.Card.Devices, &deviceIDsFromCard); err == nil {
			for _, id := range deviceIDsFromCard {
				deviceIDs[id] = true
			}
		}

		// 收集住户 ID
		if card.Card.CardType == "ActiveBed" && card.Card.ResidentID.Valid {
			residentIDs[card.Card.ResidentID.String] = true
		} else if card.Card.CardType == "Unit" {
			var residentIDsFromCard []string
			if err := json.Unmarshal(card.Card.Residents, &residentIDsFromCard); err == nil {
				for _, id := range residentIDsFromCard {
					residentIDs[id] = true
				}
			}
		}
	}

	// 2. 批量查询（一次性查询，最多 500 条）
	devices, err := s.batchGetDevices(ctx, cards[0].Card.TenantID, mapKeys(deviceIDs))
	if err != nil {
		s.logger.Warn("Failed to batch get devices, continuing with empty devices",
			zap.Error(err),
			zap.String("tenant_id", cards[0].Card.TenantID),
		)
		devices = make(map[string]*domain.Device) // 继续处理，使用空 map
	}

	residents, err := s.batchGetResidents(ctx, cards[0].Card.TenantID, mapKeys(residentIDs))
	if err != nil {
		s.logger.Warn("Failed to batch get residents, continuing with empty residents",
			zap.Error(err),
			zap.String("tenant_id", cards[0].Card.TenantID),
		)
		residents = make(map[string]*domain.Resident) // 继续处理，使用空 map
	}

	// 3. 聚合数据
	var result []*domain.CardOverviewItem
	for _, card := range cards {
		item, err := s.aggregateSingleCard(card, devices, residents)
		if err != nil {
			// 单个卡片聚合失败，记录日志，跳过该卡片
			s.logger.Warn("Failed to aggregate card, skipping",
				zap.Error(err),
				zap.String("card_id", card.Card.CardID),
			)
			continue // 跳过该卡片，不中断整个请求
		}

		result = append(result, item)
	}

	return result, nil
}

// aggregateSingleCard 聚合单个卡片的数据
func (s *cardService) aggregateSingleCard(
	card *domain.CardWithUnitInfo,
	devices map[string]*domain.Device,
	residents map[string]*domain.Resident,
) (*domain.CardOverviewItem, error) {
	item := &domain.CardOverviewItem{
		CardID:      card.Card.CardID,
		TenantID:    card.Card.TenantID,
		CardType:    card.Card.CardType,
		CardName:    card.Card.CardName,
		CardAddress: card.Card.CardAddress,
		UnhandledAlarm0: card.Card.UnhandledAlarm0,
		UnhandledAlarm1: card.Card.UnhandledAlarm1,
		UnhandledAlarm2: card.Card.UnhandledAlarm2,
		UnhandledAlarm3: card.Card.UnhandledAlarm3,
		UnhandledAlarm4: card.Card.UnhandledAlarm4,
		IconAlarmLevel:  card.Card.IconAlarmLevel,
		PopAlarmEmerge:  card.Card.PopAlarmEmerge,
	}

	// 设置 nullable 字段
	if card.Card.BedID.Valid {
		item.BedID = &card.Card.BedID.String
	}
	if card.Card.UnitID.Valid {
		item.UnitID = &card.Card.UnitID.String
	}
	if card.Card.ResidentID.Valid {
		item.ResidentID = &card.Card.ResidentID.String
	}

	// 设置 Unit 信息
	if card.Unit != nil {
		item.UnitType = card.Unit.UnitType
		item.IsPublicSpace = card.Unit.IsPublicSpace
		item.IsMultiPersonRoom = card.Unit.IsMultiPersonRoom
	}

	// 聚合设备
	var deviceIDsFromCard []string
	if err := json.Unmarshal(card.Card.Devices, &deviceIDsFromCard); err == nil {
		for _, id := range deviceIDsFromCard {
			if device, ok := devices[id]; ok {
				item.Devices = append(item.Devices, domain.CardDevice{
					DeviceID:   device.DeviceID,
					DeviceName: device.DeviceName,
					DeviceType: "", // 可以从 device_store 获取，暂时留空
				})
			} else {
				// 设备不存在，记录警告
				s.logger.Warn("Device not found, skipping",
					zap.String("device_id", id),
					zap.String("card_id", card.Card.CardID),
				)
			}
		}
	}

	// 聚合住户
	if card.Card.CardType == "ActiveBed" && card.Card.ResidentID.Valid {
		if resident, ok := residents[card.Card.ResidentID.String]; ok {
			item.Residents = append(item.Residents, domain.CardResident{
				ResidentID:   resident.ResidentID,
				Nickname:     resident.Nickname,
				ServiceLevel: resident.ServiceLevel,
			})
		}
	} else if card.Card.CardType == "Location" || card.Card.CardType == "Unit" {
		var residentIDsFromCard []string
		if err := json.Unmarshal(card.Card.Residents, &residentIDsFromCard); err == nil {
			for _, id := range residentIDsFromCard {
				if resident, ok := residents[id]; ok {
					item.Residents = append(item.Residents, domain.CardResident{
						ResidentID:   resident.ResidentID,
						Nickname:     resident.Nickname,
						ServiceLevel: resident.ServiceLevel,
					})
				} else {
					// 住户不存在，记录警告
					s.logger.Warn("Resident not found, skipping",
						zap.String("resident_id", id),
						zap.String("card_id", card.Card.CardID),
					)
				}
			}
		}
	}

	// 设置计数字段
	item.DeviceCount = len(item.Devices)
	item.ResidentCount = len(item.Residents)
	item.CaregiverCount = len(item.Caregivers)

	// 初始化可选字段
	if item.CaregiverGroups == nil {
		item.CaregiverGroups = []string{}
	}
	if item.Caregivers == nil {
		item.Caregivers = []domain.CardCaregiver{}
	}

	return item, nil
}

// batchGetDevices 批量查询设备
func (s *cardService) batchGetDevices(ctx context.Context, tenantID string, deviceIDs []string) (map[string]*domain.Device, error) {
	if len(deviceIDs) == 0 {
		return make(map[string]*domain.Device), nil
	}

	query := `
		SELECT 
			device_id::text,
			tenant_id::text,
			device_store_id::text,
			device_name,
			serial_number,
			uid,
			bound_room_id::text,
			bound_bed_id::text,
			status,
			business_access,
			monitoring_enabled,
			metadata
		FROM devices
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(deviceIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*domain.Device)
	for rows.Next() {
		var device domain.Device
		var deviceStoreID, serialNumber, uid, boundRoomID, boundBedID sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&device.DeviceID,
			&device.TenantID,
			&deviceStoreID,
			&device.DeviceName,
			&serialNumber,
			&uid,
			&boundRoomID,
			&boundBedID,
			&device.Status,
			&device.BusinessAccess,
			&device.MonitoringEnabled,
			&metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if deviceStoreID.Valid {
			device.DeviceStoreID = sql.NullString{String: deviceStoreID.String, Valid: true}
		}
		if serialNumber.Valid {
			device.SerialNumber = sql.NullString{String: serialNumber.String, Valid: true}
		}
		if uid.Valid {
			device.UID = sql.NullString{String: uid.String, Valid: true}
		}
		if boundRoomID.Valid {
			device.BoundRoomID = sql.NullString{String: boundRoomID.String, Valid: true}
		}
		if boundBedID.Valid {
			device.BoundBedID = sql.NullString{String: boundBedID.String, Valid: true}
		}
		if metadata.Valid {
			device.Metadata = sql.NullString{String: metadata.String, Valid: true}
		}

		result[device.DeviceID] = &device
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate devices: %w", err)
	}

	return result, nil
}

// batchGetResidents 批量查询住户
func (s *cardService) batchGetResidents(ctx context.Context, tenantID string, residentIDs []string) (map[string]*domain.Resident, error) {
	if len(residentIDs) == 0 {
		return make(map[string]*domain.Resident), nil
	}

	query := `
		SELECT 
			resident_id::text,
			tenant_id::text,
			resident_account,
			nickname,
			status,
			can_view_status
		FROM residents
		WHERE tenant_id = $1
		  AND resident_id = ANY($2::uuid[])
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(residentIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query residents: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*domain.Resident)
	for rows.Next() {
		var resident domain.Resident
		err := rows.Scan(
			&resident.ResidentID,
			&resident.TenantID,
			&resident.ResidentAccount,
			&resident.Nickname,
			&resident.Status,
			&resident.CanViewStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}

		result[resident.ResidentID] = &resident
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate residents: %w", err)
	}

	return result, nil
}

// getFamilyView 计算 family_view
func (s *cardService) getFamilyView(ctx context.Context, item *domain.CardOverviewItem) (bool, error) {
	// 1. Share Unit：不对个人开放
	if item.IsPublicSpace || item.IsMultiPersonRoom {
		return false, nil
	}

	// 2. ActiveBed 卡片：从 resident_id 获取
	if item.CardType == "ActiveBed" && item.ResidentID != nil {
		var canViewStatus sql.NullBool
		err := s.db.QueryRowContext(ctx,
			`SELECT can_view_status FROM residents 
			 WHERE tenant_id = $1 AND resident_id = $2`,
			item.TenantID, *item.ResidentID,
		).Scan(&canViewStatus)
		if err != nil {
			return false, err
		}
		return canViewStatus.Valid && canViewStatus.Bool, nil
	}

	// 3. Unit 卡片（数据库中使用 'Location'）：从第一个住户获取
	if (item.CardType == "Unit" || item.CardType == "Location") && len(item.Residents) > 0 {
		// 第一个住户的 can_view_status
		var canViewStatus1 sql.NullBool
		err := s.db.QueryRowContext(ctx,
			`SELECT can_view_status FROM residents 
			 WHERE tenant_id = $1 AND resident_id = $2`,
			item.TenantID, item.Residents[0].ResidentID,
		).Scan(&canViewStatus1)
		if err != nil {
			return false, err
		}

		// 如果只有一个人，返回第一个住户的 can_view_status
		if len(item.Residents) == 1 {
			return canViewStatus1.Valid && canViewStatus1.Bool, nil
		}

		// 如果有两个人（夫妻套房），检查第二个住户
		if len(item.Residents) >= 2 {
			var canViewStatus2 sql.NullBool
			err := s.db.QueryRowContext(ctx,
				`SELECT can_view_status FROM residents 
				 WHERE tenant_id = $1 AND resident_id = $2`,
				item.TenantID, item.Residents[1].ResidentID,
			).Scan(&canViewStatus2)
			if err != nil {
				return false, err
			}

			// OR 逻辑：只要有一个允许，就允许
			return (canViewStatus1.Valid && canViewStatus1.Bool) ||
				(canViewStatus2.Valid && canViewStatus2.Bool), nil
		}
	}

	return false, nil
}

// mapKeys 从 map 中提取 keys
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

