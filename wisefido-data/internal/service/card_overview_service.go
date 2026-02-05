package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	commoncard "owl-common/card"
	"wisefido-data/internal/domain"
	"go.uber.org/zap"
)

// GetCardOverviewRequest 卡片概览请求
type GetCardOverviewRequest struct {
	TenantID        string
	CardID          string
	Search          string
	CardType        string
	UnitType        string
	IsPublicSpace   *bool
	IsSharedUnit    *bool
	Sort            string
	Direction       string
	Page            int
	PageSize        int
	CurrentUserID   string
	CurrentUserType string
	CurrentUserRole string
}

// GetCardOverviewResponse 卡片概览响应
type GetCardOverviewResponse struct {
	Items []domain.CardOverviewItem
	Total int
	Page  int
	Size  int
}

// GetCardOverview 获取卡片概览（分页，每页最多 20 条）
func (s *CardStaticService) GetCardOverview(ctx context.Context, req GetCardOverviewRequest) (*GetCardOverviewResponse, error) {
	if req.TenantID == "" || req.CurrentUserID == "" || req.CurrentUserType == "" {
		return nil, fmt.Errorf("tenant_id, current_user_id, current_user_type are required")
	}
	if req.CurrentUserType != "resident" && req.CurrentUserType != "staff" {
		return nil, fmt.Errorf("invalid current_user_type: must be resident or staff")
	}
	if s.perm == nil {
		return nil, fmt.Errorf("permission provider not configured")
	}

	page, pageSize := normPage(req.Page, req.PageSize)
	ids, err := s.getAllowedCardIDsWithCache(ctx, req.TenantID, req.CurrentUserID, req.CurrentUserType, req.CurrentUserRole)
	if err != nil {
		return nil, fmt.Errorf("get allowed cards: %w", err)
	}
	all, err := s.listCardsWithFilter(ctx, ListVitalFocusCardInfoRequest{
		TenantID: req.TenantID,
	}, idsToSet(ids))
	if err != nil {
		return nil, err
	}

	// 应用概览过滤器
	filtered := s.applyOverviewFilters(ctx, req, all)
	// 排序
	s.applyOverviewSort(filtered, req.Sort, req.Direction)
	// 分页
	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return &GetCardOverviewResponse{Items: []domain.CardOverviewItem{}, Total: total, Page: page, Size: pageSize}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pageItems := filtered[start:end]

	// 转换为 CardOverviewItem（需补全 Unit 信息）
	items, err := s.toCardOverviewItems(ctx, req.TenantID, pageItems)
	if err != nil {
		return nil, err
	}
	return &GetCardOverviewResponse{Items: items, Total: total, Page: page, Size: pageSize}, nil
}

func (s *CardStaticService) applyOverviewFilters(ctx context.Context, req GetCardOverviewRequest, cards []commoncard.VitalFocusCardInfo) []commoncard.VitalFocusCardInfo {
	var out []commoncard.VitalFocusCardInfo
	searchLower := strings.ToLower(strings.TrimSpace(req.Search))
	for _, c := range cards {
		if req.CardID != "" && c.CardID != req.CardID {
			continue
		}
		if req.CardType != "" && c.CardType != req.CardType {
			continue
		}
		if searchLower != "" {
			nameMatch := strings.Contains(strings.ToLower(c.CardName), searchLower)
			addrMatch := strings.Contains(strings.ToLower(c.CardAddress), searchLower)
			if !nameMatch && !addrMatch {
				continue
			}
		}
		if req.UnitType != "" || req.IsPublicSpace != nil || req.IsSharedUnit != nil {
			if s.unitsRepo == nil || c.UnitID == "" {
				continue
			}
			unit, err := s.unitsRepo.GetUnit(ctx, req.TenantID, c.UnitID)
			if err != nil || unit == nil {
				continue
			}
			if req.UnitType != "" && unit.UnitType != req.UnitType {
				continue
			}
			if req.IsPublicSpace != nil && unit.IsPublic != *req.IsPublicSpace {
				continue
			}
			if req.IsSharedUnit != nil && unit.IsSharedUnit != *req.IsSharedUnit {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func (s *CardStaticService) applyOverviewSort(cards []commoncard.VitalFocusCardInfo, sortField, direction string) {
	if sortField == "" {
		sortField = "card_name"
	}
	asc := direction != "desc"
	switch sortField {
	case "card_name":
		sort.Slice(cards, func(i, j int) bool {
			cmp := strings.Compare(cards[i].CardName, cards[j].CardName)
			return (asc && cmp < 0) || (!asc && cmp > 0)
		})
	case "card_id":
		sort.Slice(cards, func(i, j int) bool {
			cmp := strings.Compare(cards[i].CardID, cards[j].CardID)
			return (asc && cmp < 0) || (!asc && cmp > 0)
		})
	default:
		sort.Slice(cards, func(i, j int) bool {
			cmp := strings.Compare(cards[i].CardName, cards[j].CardName)
			return (asc && cmp < 0) || (!asc && cmp > 0)
		})
	}
}

func (s *CardStaticService) toCardOverviewItems(ctx context.Context, tenantID string, cards []commoncard.VitalFocusCardInfo) ([]domain.CardOverviewItem, error) {
	out := make([]domain.CardOverviewItem, 0, len(cards))
	for _, c := range cards {
		item := domain.CardOverviewItem{
			CardID:       c.CardID,
			TenantID:     c.TenantID,
			CardType:     c.CardType,
			BedID:        c.BedID,
			CardName:     c.CardName,
			CardAddress:  c.CardAddress,
			ResidentID:   c.PrimaryResidentID,
			ResidentAccess: true,
		}
		if c.UnitID != "" {
			item.UnitID = &c.UnitID
		}
		if s.unitsRepo != nil && c.UnitID != "" {
			unit, err := s.unitsRepo.GetUnit(ctx, tenantID, c.UnitID)
			if err == nil && unit != nil {
				item.UnitType = unit.UnitType
				item.IsPublicSpace = unit.IsPublic
				item.IsSharedUnit = unit.IsSharedUnit
			}
		}
		// 转换设备
		for _, d := range c.Devices {
			item.Devices = append(item.Devices, domain.CardDevice{
				DeviceID:    d.DeviceID,
				UID:         d.DeviceUID,
				DeviceUID:   d.DeviceUID,
				DeviceCode:  d.DeviceCode,
				DeviceName:  d.DeviceName,
				DeviceType:  d.DeviceType,
				DeviceModel: d.DeviceModel,
				Status:      d.Status,
			})
		}
		// 转换住户
		for _, r := range c.Residents {
			item.Residents = append(item.Residents, domain.CardResident{
				ResidentID:   r.ResidentID,
				Nickname:     r.Nickname,
				ServiceLevel: r.ServiceLevel,
			})
		}
		
		if c.IconAlarmLevel != nil {
			item.IconAlarmLevel = *c.IconAlarmLevel
		}
		if c.PopAlarmEmerge != nil {
			item.PopAlarmEmerge = *c.PopAlarmEmerge
		}
		
		// 为每个卡片查询相关的护理人员信息
		if s.residentsRepo != nil {
			// 收集所有住户ID
			var residentIDs []string
			for _, resident := range c.Residents {
				residentIDs = append(residentIDs, resident.ResidentID)
			}
			
			// 查询所有住户相关的护理人员
			for _, residentID := range residentIDs {
				caregivers, err := s.residentsRepo.GetResidentCaregivers(ctx, tenantID, residentID)
				if err != nil {
					s.logger.Error("Failed to get resident caregivers",
						zap.String("tenant_id", tenantID),
						zap.String("resident_id", residentID),
						zap.Error(err),
					)
					continue
				}
				
				// 将护理人员信息添加到卡片中
				for _, caregiver := range caregivers {
					if caregiver != nil && len(caregiver.UserList) > 0 {
						var userIDs []string
						if err := json.Unmarshal(caregiver.UserList, &userIDs); err != nil {
							s.logger.Error("Failed to unmarshal caregiver user list",
								zap.String("tenant_id", tenantID),
								zap.String("resident_id", residentID),
								zap.Error(err),
							)
						} else {
							for _, userID := range userIDs {
								item.Caregivers = append(item.Caregivers, domain.CardCaregiver{
									UserID: userID,
									Nickname: "", // 可以通过用户服务查询更详细的信息
									UserAccount: "",
									UserRole: "",
								})
							}
						}
						
						// 添加用户组信息
						if len(caregiver.GroupList) > 0 {
							var groupNames []string
							if err := json.Unmarshal(caregiver.GroupList, &groupNames); err != nil {
								s.logger.Error("Failed to unmarshal caregiver group list",
									zap.String("tenant_id", tenantID),
									zap.String("resident_id", residentID),
									zap.Error(err),
								)
							} else {
								item.CaregiverGroups = append(item.CaregiverGroups, groupNames...)
							}
						}
					}
				}
			}
		}
		
		// 如果需要获取告警计数，可以通过另外的查询获得
		// 这里暂时保留原始结构，后续可根据需要添加告警计数查询
		
		item.CaregiverCount = len(item.Caregivers)
		item.DeviceCount = len(item.Devices)
		item.ResidentCount = len(item.Residents)
		out = append(out, item)
	}
	return out, nil
}
