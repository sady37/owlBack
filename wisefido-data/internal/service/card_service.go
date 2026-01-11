package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/models"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// CardService 卡片服务接口
type CardService interface {
	// GetCardOverview 获取卡片概览列表（返回所有可见的卡片）
	GetCardOverview(ctx context.Context, req GetCardOverviewRequest) (*GetCardOverviewResponse, error)

	// ListVitalFocusCards 获取 Vital Focus 卡片列表（包含实时数据和报警数据）
	ListVitalFocusCards(ctx context.Context, req ListVitalFocusCardsRequest) (*ListVitalFocusCardsResponse, error)

	// GetVitalFocusCard 获取单个 Vital Focus 卡片（包含实时数据和报警数据）
	// 复用权限过滤逻辑，先从 cards 表查询（应用权限过滤），然后从 Redis full cache 读取完整数据
	GetVitalFocusCard(ctx context.Context, req GetVitalFocusCardRequest) (*models.VitalFocusCard, error)

	// GetVitalFocusPreferences 获取用户的 Vital Focus preferences（selectedCardIds）
	GetVitalFocusPreferences(ctx context.Context, req GetVitalFocusPreferencesRequest) (*GetVitalFocusPreferencesResponse, error)

	// SaveVitalFocusPreferences 保存用户的 Vital Focus preferences（selectedCardIds）
	SaveVitalFocusPreferences(ctx context.Context, req SaveVitalFocusPreferencesRequest) error
}

// cardService 卡片服务实现
type cardService struct {
	cardsRepo     repository.CardsRepository
	residentsRepo repository.ResidentsRepository
	devicesRepo   repository.DevicesRepository
	usersRepo     repository.UsersRepository
	kv            store.KV // 新增：读取 Redis full cache
	db            *sql.DB  // 用于复杂查询（批量查询直接使用 db）
	logger        *zap.Logger
}

// NewCardService 创建卡片服务
func NewCardService(
	cardsRepo repository.CardsRepository,
	residentsRepo repository.ResidentsRepository,
	devicesRepo repository.DevicesRepository,
	usersRepo repository.UsersRepository,
	kv store.KV, // 新增：Redis KV 存储
	db *sql.DB,
	logger *zap.Logger,
) CardService {
	return &cardService{
		cardsRepo:     cardsRepo,
		residentsRepo: residentsRepo,
		devicesRepo:   devicesRepo,
		usersRepo:     usersRepo,
		kv:            kv,
		db:            db,
		logger:        logger,
	}
}

// GetCardOverviewRequest 获取卡片概览请求
type GetCardOverviewRequest struct {
	TenantID      string
	CardID        string // 可选：查询单个卡片
	Search        string // 搜索关键词
	CardType      string // "ActiveBed" | "Unit"
	UnitType      string // "Home" | "Facility"
	IsPublicSpace *bool
	IsSharedUnit  *bool
	Sort          string // "card_name" | "card_address"
	Direction     string // "asc" | "desc"

	// 权限相关
	CurrentUserID   string
	CurrentUserType string // "resident" | "staff"（注意：'family' 已被禁止，resident_contacts 不能登录系统）
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
		TenantID:      req.TenantID,
		CardID:        req.CardID,
		Search:        req.Search,
		CardType:      req.CardType,
		UnitType:      req.UnitType,
		IsPublicSpace: req.IsPublicSpace,
		IsSharedUnit:  req.IsSharedUnit,
		Sort:          req.Sort,
		Direction:     req.Direction,
	}

	// 2. 处理用户类型权限过滤
	// 注意：contact_id 已不存在，contact 仅是 resident 的属性，不再有独立的 contact 用户类型
	if req.CurrentUserType == "resident" {
		// Resident 用户：只能看到自己的卡片
		// CurrentUserID 应该是 resident_id（从登录 API 返回的 userId）
		if req.CurrentUserID != "" {
			repoReq.PermissionFilter = &repository.PermissionFilter{
				UserID:   req.CurrentUserID,
				UserType: "resident",
			}
			s.logger.Info("Card permission filter: Resident user",
				zap.String("resident_id", req.CurrentUserID),
				zap.String("tenant_id", req.TenantID),
			)
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
			if user.BranchName.Valid {
				branchName := user.BranchName.String
				repoReq.PermissionFilter.UserBranchTag = &branchName
			} else {
				emptyTag := ""
				repoReq.PermissionFilter.UserBranchTag = &emptyTag
			}
		}

		if perm.AssignedOnly {
			// AssignedOnly：先查询 resident_caregivers 获取分配的 resident_id 列表
			if req.CurrentUserID != "" {
				assignedResidentIDs, err := s.getAssignedResidentIDs(ctx, req.TenantID, req.CurrentUserID)
				if err != nil {
					return nil, fmt.Errorf("failed to get assigned resident IDs: %w", err)
				}

				if len(assignedResidentIDs) == 0 {
					// 没有分配的住户，返回空结果
					s.logger.Info("Card permission filter: Staff user with AssignedOnly but no assigned residents",
						zap.String("user_id", req.CurrentUserID),
						zap.String("role", req.CurrentUserRole),
						zap.String("tenant_id", req.TenantID),
					)
					return &GetCardOverviewResponse{
						Items: []*domain.CardOverviewItem{},
						Total: 0,
					}, nil
				}

				// 将 assignedResidentIDs 传递给 repository
				repoReq.PermissionFilter.AssignedResidentIDs = assignedResidentIDs
				s.logger.Info("Card permission filter: Staff user with AssignedOnly",
					zap.String("user_id", req.CurrentUserID),
					zap.String("role", req.CurrentUserRole),
					zap.String("tenant_id", req.TenantID),
					zap.Int("assigned_resident_count", len(assignedResidentIDs)),
				)
			}
		} else {
			// Staff 用户没有 AssignedOnly 限制，但也没有设置 PermissionFilter
			// 这意味着可以看到所有卡片（根据其他权限配置）
			s.logger.Info("Card permission filter: Staff user without AssignedOnly",
				zap.String("user_id", req.CurrentUserID),
				zap.String("role", req.CurrentUserRole),
				zap.Bool("branch_only", perm.BranchOnly),
				zap.String("tenant_id", req.TenantID),
			)
		}
	} else {
		// 未知用户类型或未提供用户类型
		s.logger.Warn("Card permission filter: Unknown or missing user type",
			zap.String("current_user_type", req.CurrentUserType),
			zap.String("current_user_id", req.CurrentUserID),
			zap.String("current_user_role", req.CurrentUserRole),
			zap.String("tenant_id", req.TenantID),
		)
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

	// 5. 规范化 card_type（数据库使用 'Location'，API 返回 'Unit'）
	for _, card := range allCards {
		// 规范化 card_type：数据库中的 'Location' 转换为 API 的 'Unit'
		if card.CardType == "Location" {
			card.CardType = "Unit"
		}
	}

	// 6. 返回所有卡片（前端负责分页）
	return &GetCardOverviewResponse{
		Items: allCards,
		Total: len(allCards),
	}, nil
}

// getResourcePermission 查询资源权限配置
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func (s *cardService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error) {
	var permissionScope string
	err := s.db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
		   AND role_code = $1
		   AND resource_type = $2
		   AND permission_type = $3`,
		roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

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
		// 未知值，返回默认值（无限制）
		assignedOnly = false
		branchOnly = false
	}

	return &PermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// PermissionCheck 权限检查结果（已在 alarm_event_service.go 中定义，这里使用相同的结构）
// 注意：如果 alarm_event_service.go 中的定义改变，这里也需要同步更新

// aggregateCardData 聚合卡片数据（devices, residents）
func (s *cardService) aggregateCardData(ctx context.Context, cards []*domain.CardWithUnitInfo) ([]*domain.CardOverviewItem, error) {
	if len(cards) == 0 {
		return []*domain.CardOverviewItem{}, nil
	}

	// 1. 收集所有需要查询的 ID
	deviceIDs := make(map[string]bool)
	residentIDs := make(map[string]bool)

	for _, card := range cards {
		// 收集设备 ID（从 JSONB 对象数组中提取）
		var devicesFromCard []map[string]interface{}
		if err := json.Unmarshal(card.Card.Devices, &devicesFromCard); err != nil {
			s.logger.Warn("Failed to parse devices JSONB, skipping",
				zap.Error(err),
				zap.String("card_id", card.Card.CardID),
			)
			continue
		}
		// JSONB 存储的是对象数组
		for _, deviceObj := range devicesFromCard {
			if deviceID, ok := deviceObj["device_id"].(string); ok && deviceID != "" {
				deviceIDs[deviceID] = true
			}
		}

		// 收集住户 ID
		if card.Card.CardType == "ActiveBed" {
			// ActiveBed 卡片：优先使用 resident_id，如果为 NULL 则从 residents JSONB 读取
			if card.Card.ResidentID.Valid {
				residentIDs[card.Card.ResidentID.String] = true
			} else {
				// resident_id 为 NULL，从 residents JSONB 字段读取
				var residentsFromCard []map[string]interface{}
				if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err == nil {
					for _, residentObj := range residentsFromCard {
						if residentID, ok := residentObj["resident_id"].(string); ok && residentID != "" {
							residentIDs[residentID] = true
						}
					}
				}
			}
		} else if card.Card.CardType == "Location" || card.Card.CardType == "Unit" {
			// 注意：数据库中使用 'Location'，但逻辑上与 'Unit' 相同
			// 从 JSONB 对象数组中提取 resident_id
			var residentsFromCard []map[string]interface{}
			if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err != nil {
				s.logger.Warn("Failed to parse residents JSONB, skipping",
					zap.Error(err),
					zap.String("card_id", card.Card.CardID),
				)
				continue
			}
			// JSONB 存储的是对象数组
			for _, residentObj := range residentsFromCard {
				if residentID, ok := residentObj["resident_id"].(string); ok && residentID != "" {
					residentIDs[residentID] = true
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

	// 3. 聚合数据（从 card 表获取基础信息，从 Redis full cache 获取动态数据）
	var result []*domain.CardOverviewItem
	for _, card := range cards {
		item, err := s.aggregateSingleCard(ctx, card, devices, residents)
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
// 基础信息从 card 表获取，动态数据（报警统计等）从 Redis full cache（card aggregator 写入）获取
func (s *cardService) aggregateSingleCard(
	ctx context.Context,
	card *domain.CardWithUnitInfo,
	devices map[string]*domain.Device,
	residents map[string]*domain.Resident,
) (*domain.CardOverviewItem, error) {
	// 1. 初始化基础信息（从 card 表）
	item := &domain.CardOverviewItem{
		CardID:       card.Card.CardID,
		TenantID:     card.Card.TenantID,
		CardType:     card.Card.CardType,
		CardName:     card.Card.CardName,
		CardAddress:  card.Card.CardAddress,
		IconAlarmLevel: card.Card.IconAlarmLevel,
		PopAlarmEmerge:  card.Card.PopAlarmEmerge,
		// 报警统计初始值从 card 表获取（作为回退）
		UnhandledAlarm0: card.Card.UnhandledAlarm0,
		UnhandledAlarm1: card.Card.UnhandledAlarm1,
		UnhandledAlarm2: card.Card.UnhandledAlarm2,
		UnhandledAlarm3: card.Card.UnhandledAlarm3,
		UnhandledAlarm4: card.Card.UnhandledAlarm4,
	}

	// 2. 尝试从 Redis full cache 获取动态数据（card aggregator 写入的）
	// 如果 Redis 中有数据，使用 Redis 中的报警统计数据（更实时）
	key := "vital-focus:card:" + card.Card.CardID + ":full"
	raw, err := s.kv.Get(ctx, key)
	if err == nil && raw != "" {
		// 成功从 Redis 获取数据，解析并合并动态数据
		vitalCard, ok := decodeAndNormalizeFullCard(raw)
		if ok {
			// 使用 Redis 中的报警统计数据（更实时）
			if vitalCard.UnhandledAlarm0 != nil {
				item.UnhandledAlarm0 = *vitalCard.UnhandledAlarm0
			}
			if vitalCard.UnhandledAlarm1 != nil {
				item.UnhandledAlarm1 = *vitalCard.UnhandledAlarm1
			}
			if vitalCard.UnhandledAlarm2 != nil {
				item.UnhandledAlarm2 = *vitalCard.UnhandledAlarm2
			}
			if vitalCard.UnhandledAlarm3 != nil {
				item.UnhandledAlarm3 = *vitalCard.UnhandledAlarm3
			}
			if vitalCard.UnhandledAlarm4 != nil {
				item.UnhandledAlarm4 = *vitalCard.UnhandledAlarm4
			}
			if vitalCard.IconAlarmLevel != nil {
				item.IconAlarmLevel = *vitalCard.IconAlarmLevel
			}
			if vitalCard.PopAlarmEmerge != nil {
				item.PopAlarmEmerge = *vitalCard.PopAlarmEmerge
			}
			s.logger.Debug("Merged dynamic data from Redis cache",
				zap.String("card_id", card.Card.CardID),
			)
		} else {
			// Redis 数据解析失败，使用 card 表的数据（已初始化）
			s.logger.Debug("Failed to decode Redis cache, using card table data",
				zap.String("card_id", card.Card.CardID),
			)
		}
	} else {
		// Redis 中没有数据，使用 card 表的数据（已初始化）
		s.logger.Debug("Redis cache not available, using card table data",
			zap.String("card_id", card.Card.CardID),
			zap.Error(err),
		)
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
		item.IsPublicSpace = card.Unit.IsPublic
		item.IsSharedUnit = card.Unit.IsSharedUnit
	}

	// 聚合设备
	// 注意：cards.devices JSONB 字段存储的是对象数组
	var devicesFromCard []map[string]interface{}
	if err := json.Unmarshal(card.Card.Devices, &devicesFromCard); err != nil {
		s.logger.Warn("Failed to parse devices JSONB",
			zap.Error(err),
			zap.String("card_id", card.Card.CardID),
		)
		// 解析失败，跳过设备处理
	} else {
		for _, deviceObj := range devicesFromCard {
			deviceID, _ := deviceObj["device_id"].(string)
			deviceName, _ := deviceObj["device_name"].(string)
			deviceTypeStr, _ := deviceObj["device_type"].(string)
			deviceModel, _ := deviceObj["device_model"].(string)

			if deviceID != "" {
				// 将 device_type 从字符串转换为数字（前端期望：1=sleepace, 2=radar）
				var deviceTypeNum interface{} = nil
				if deviceTypeStr != "" {
					// 统一转换为小写进行比较（支持各种大小写组合）
					deviceTypeLower := strings.ToLower(deviceTypeStr)
					if deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad" {
						deviceTypeNum = 1 // 数字类型
					} else if deviceTypeLower == "radar" {
						deviceTypeNum = 2 // 数字类型
					}
				}

				// 尝试从数据库获取完整信息（用于获取 status, serial_number 等）
				if device, ok := devices[deviceID]; ok {
					// 如果 JSONB 中没有 device_type，尝试从数据库获取
					if deviceTypeNum == nil && device.DeviceType.Valid {
						deviceTypeStrFromDB := device.DeviceType.String
						deviceTypeLowerFromDB := strings.ToLower(deviceTypeStrFromDB)
						if deviceTypeLowerFromDB == "sleepace" || deviceTypeLowerFromDB == "sleepad" || deviceTypeLowerFromDB == "sleeppad" {
							deviceTypeNum = 1
						} else if deviceTypeLowerFromDB == "radar" {
							deviceTypeNum = 2
						}
					}

					// 使用数据库中的 device_model（如果 JSONB 中没有）
					if deviceModel == "" && device.DeviceModel.Valid {
						deviceModel = device.DeviceModel.String
					}

					// 使用数据库中的完整信息
					serialNumber := ""
					if device.SerialNumber.Valid {
						serialNumber = device.SerialNumber.String
					}
					uid := ""
					if device.UID.Valid {
						uid = device.UID.String
					}
					item.Devices = append(item.Devices, domain.CardDevice{
						DeviceID:     device.DeviceID,
						DeviceName:   device.DeviceName,
						DeviceType:   deviceTypeNum, // 数字类型（1 或 2）
						DeviceModel:  deviceModel,
						SerialNumber: serialNumber,
						UID:          uid,
						Status:       device.Status,
					})
				} else {
					// 如果数据库中没有，使用 JSONB 中的数据
					serialNumber, _ := deviceObj["serial_number"].(string)
					uid, _ := deviceObj["uid"].(string)
					status, _ := deviceObj["status"].(string)
					item.Devices = append(item.Devices, domain.CardDevice{
						DeviceID:     deviceID,
						DeviceName:   deviceName,
						DeviceType:   deviceTypeNum, // 数字类型（1 或 2）
						DeviceModel:  deviceModel,
						SerialNumber: serialNumber,
						UID:          uid,
						Status:       status,
					})
				}
			}
		}
	}

	// 聚合住户并计算 ResidentAccess
	// ResidentAccess: 是否允许住户访问（从 residents.is_access_enabled 获取）
	item.ResidentAccess = false
	if card.Card.CardType == "ActiveBed" {
		// ActiveBed 卡片：优先使用 resident_id，如果为 NULL 则从 residents JSONB 读取
		if card.Card.ResidentID.Valid {
			// 使用 resident_id 字段
			if resident, ok := residents[card.Card.ResidentID.String]; ok {
				item.Residents = append(item.Residents, domain.CardResident{
					ResidentID:   resident.ResidentID,
					Nickname:     resident.Nickname,
					ServiceLevel: resident.ServiceLevel,
				})
				item.ResidentAccess = resident.IsAccessEnabled
			}
		} else {
			// resident_id 为 NULL，从 residents JSONB 字段读取
			var residentsFromCard []map[string]interface{}
			if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err != nil {
				s.logger.Warn("Failed to parse residents JSONB for ActiveBed card",
					zap.Error(err),
					zap.String("card_id", card.Card.CardID),
				)
			} else {
				// 解析住户信息
				for _, residentObj := range residentsFromCard {
					residentID, _ := residentObj["resident_id"].(string)
					nickname, _ := residentObj["nickname"].(string)

					if residentID != "" {
						// 尝试从数据库获取完整信息（用于获取 service_level, is_access_enabled 等）
						if resident, ok := residents[residentID]; ok {
							// 使用数据库中的完整信息
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   resident.ResidentID,
								Nickname:     resident.Nickname,
								ServiceLevel: resident.ServiceLevel,
							})
							// 如果至少有一个 resident 的 is_access_enabled = TRUE，则 ResidentAccess = TRUE
							if resident.IsAccessEnabled {
								item.ResidentAccess = true
							}
						} else {
							// 如果数据库中没有，使用 JSONB 中的数据
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   residentID,
								Nickname:     nickname,
								ServiceLevel: "",
							})
						}
					}
				}
			}
		}
	} else if card.Card.CardType == "Location" || card.Card.CardType == "Unit" {
		// Unit 卡片：根据 Unit 类型决定
		// 1. 如果 is_shared_unit = TRUE 或 is_public = TRUE，说明是公共区域，不能访问，ResidentAccess = FALSE
		// 2. 如果 is_shared_unit = FALSE 且 is_public = FALSE，检查该 Unit 里的 resident 是否 is_access_enabled
		if card.Unit != nil {
			// 注意：cards.residents JSONB 字段存储的是对象数组
			var residentsFromCard []map[string]interface{}
			if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err != nil {
				s.logger.Warn("Failed to parse residents JSONB",
					zap.Error(err),
					zap.String("card_id", card.Card.CardID),
				)
				// 解析失败，跳过住户处理
			} else {
				// 解析住户信息（无论是否为共享单元或公共空间，都应该显示住户信息）
				for _, residentObj := range residentsFromCard {
					residentID, _ := residentObj["resident_id"].(string)
					nickname, _ := residentObj["nickname"].(string)

					if residentID != "" {
						// 尝试从数据库获取完整信息（用于获取 service_level, is_access_enabled 等）
						if resident, ok := residents[residentID]; ok {
							// 使用数据库中的完整信息
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   resident.ResidentID,
								Nickname:     resident.Nickname,
								ServiceLevel: resident.ServiceLevel,
							})
							// 如果至少有一个 resident 的 is_access_enabled = TRUE，则 ResidentAccess = TRUE
							if resident.IsAccessEnabled {
								item.ResidentAccess = true
							}
						} else {
							// 如果数据库中没有，使用 JSONB 中的数据
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   residentID,
								Nickname:     nickname,
								ServiceLevel: "",
							})
						}
					}
				}

				// 权限控制：公共区域或共享单元，不允许住户访问
				if card.Unit.IsSharedUnit || card.Unit.IsPublic {
					item.ResidentAccess = false
				}
			}
		} else {
			// Unit 信息不存在，尝试解析住户信息（使用 JSONB 数据）
			var residentsFromCard []map[string]interface{}
			if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err == nil {
				for _, residentObj := range residentsFromCard {
					residentID, _ := residentObj["resident_id"].(string)
					nickname, _ := residentObj["nickname"].(string)

					if residentID != "" {
						// 尝试从数据库获取完整信息
						if resident, ok := residents[residentID]; ok {
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   resident.ResidentID,
								Nickname:     resident.Nickname,
								ServiceLevel: resident.ServiceLevel,
							})
							if resident.IsAccessEnabled {
								item.ResidentAccess = true
							}
						} else {
							// 使用 JSONB 中的数据
							item.Residents = append(item.Residents, domain.CardResident{
								ResidentID:   residentID,
								Nickname:     nickname,
								ServiceLevel: "",
							})
						}
					}
				}
			}
			// Unit 信息不存在，默认不允许访问
			item.ResidentAccess = false
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
			d.device_id::text,
			d.tenant_id::text,
			d.device_store_id::text,
			d.device_name,
			d.serial_number,
			d.uid,
			d.bound_room_id::text,
			d.bound_bed_id::text,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			d.metadata,
			ds.device_type,
			ds.device_model
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.tenant_id = $1
		  AND d.device_id = ANY($2::uuid[])
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
		var metadata, deviceType, deviceModel sql.NullString

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
			&deviceType,
			&deviceModel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if deviceStoreID.Valid {
			device.DeviceStoreID = sql.NullString{String: deviceStoreID.String, Valid: true}
		}
		if deviceType.Valid {
			device.DeviceType = sql.NullString{String: deviceType.String, Valid: true}
		}
		if deviceModel.Valid {
			device.DeviceModel = sql.NullString{String: deviceModel.String, Valid: true}
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
			service_level,
			is_access_enabled
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
			&resident.ServiceLevel,
			&resident.IsAccessEnabled,
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

// mapKeys 从 map 中提取 keys
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
