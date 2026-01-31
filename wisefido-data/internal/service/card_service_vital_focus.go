package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/models"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// ListVitalFocusCardsRequest 获取 Vital Focus 卡片列表请求
type ListVitalFocusCardsRequest struct {
	TenantID string // 租户 ID（从 Query 或 Header 获取，不信任，需要验证）
	Page     int    // 页码（从 Query 获取，默认 1）
	PageSize int    // 每页大小（从 Query 获取，默认 10）

	// 权限相关参数（从 Header 获取，不信任，需要验证）
	CurrentUserID   string // X-User-Id
	CurrentUserType string // X-User-Type: "resident" | "staff"
	CurrentUserRole string // X-User-Role (仅 staff 需要)
}

// ListVitalFocusCardsResponse 获取 Vital Focus 卡片列表响应
type ListVitalFocusCardsResponse struct {
	Items      []models.VitalFocusCard // 卡片列表（包含实时数据和报警数据）
	Pagination Pagination              // 分页信息
}

// Pagination 分页信息
type Pagination struct {
	Page     int `json:"page"`      // 当前页码
	PageSize int `json:"page_size"` // 每页大小
	Total    int `json:"total"`     // 总记录数
}

// ListVitalFocusCards 获取 Vital Focus 卡片列表（包含实时数据和报警数据）
func (s *cardService) ListVitalFocusCards(ctx context.Context, req ListVitalFocusCardsRequest) (*ListVitalFocusCardsResponse, error) {
	// 步骤 1：安全验证 - 根据 X-User-Id 查询数据库，验证 tenant_id, role 是否一致
	validatedTenantID, validatedUserID, validatedUserRole, validatedBranchIDs, validatedAlarmScope, err := s.validateUserSecurity(ctx, req)
	if err != nil {
		// 验证失败，直接中断会话（返回错误）
		s.logger.Error("Security validation failed, terminating session",
			zap.String("user_id", req.CurrentUserID),
			zap.String("user_type", req.CurrentUserType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// 步骤 2：根据用户类型和 alarm_scope，获取需要查询的 resident_id 列表
	// 注意：权限基于 users 表中的 alarm_scope 字段：
	//   - 'ALL': 返回所有卡片（tenant_id 过滤）
	//   - 'BRANCH': 通过 branch_id 过滤（使用 user_branches 表）
	//   - 'ASSIGNED_ONLY': 通过 resident_caregivers 表查询分配的 resident_id
	//   - Resident: 硬性规定只能读自己的卡片
	var assignedResidentIDs []string

	if req.CurrentUserType == "staff" {
		// 根据 alarm_scope 判断权限逻辑
		switch validatedAlarmScope {
		case "ALL":
			// ALL: 返回所有卡片（tenant_id 过滤），无需 assignedResidentIDs
			// assignedResidentIDs 保持为空

		case "BRANCH":
			// BRANCH: 通过 branch_id 过滤，无需 assignedResidentIDs
			// assignedResidentIDs 保持为空

		case "ASSIGNED_ONLY":
			// ASSIGNED_ONLY: 通过 resident_caregivers 表查询分配的 resident_id
			assignedResidentIDs, err = s.getAssignedResidentIDs(ctx, validatedTenantID, validatedUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get assigned resident IDs: %w", err)
			}

		default:
			// alarm_scope 为空或未知值：返回空列表
			s.logger.Warn("User with unknown or empty alarm_scope, returning empty list",
				zap.String("user_id", validatedUserID),
				zap.String("role", validatedUserRole),
				zap.String("alarm_scope", validatedAlarmScope),
			)
			return &ListVitalFocusCardsResponse{
				Items: []models.VitalFocusCard{},
				Pagination: Pagination{
					Page:     req.Page,
					PageSize: req.PageSize,
					Total:    0,
				},
			}, nil
		}
	} else if req.CurrentUserType == "resident" {
		// Resident 用户：只能看到自己的卡片
		// 步骤 4：根据卡片生成逻辑，查询符合条件的卡片
		// 1. card_type='ActiveBed' && resident_id相同的 加入
		// 2. unit_type != 'SharedUnit' && unit_id相同的 加入
		cardIDs, err := s.getCardIDsForResident(ctx, validatedTenantID, validatedUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get card IDs for resident: %w", err)
		}

		// 步骤 5：从 Redis full cache 读取完整数据，如果 miss 则从 DB 构建基础卡片
		allCards := make([]models.VitalFocusCard, 0, len(cardIDs))
		missedCardIDs := make([]string, 0)
		
		for _, cardID := range cardIDs {
			key := "vital-focus:card:" + cardID + ":full"
			raw, err := s.kv.Get(ctx, key)
			if err != nil {
				if errors.Is(err, store.ErrMiss) {
					s.logger.Debug("Redis cache miss for card, will fallback to DB",
						zap.String("card_id", cardID),
					)
					missedCardIDs = append(missedCardIDs, cardID)
				} else {
					s.logger.Warn("Failed to get card from Redis, will fallback to DB",
						zap.String("card_id", cardID),
						zap.Error(err),
					)
					missedCardIDs = append(missedCardIDs, cardID)
				}
				continue
			}

			card, ok := decodeAndNormalizeFullCard(raw)
			if !ok {
				s.logger.Warn("Failed to decode and normalize card from Redis, will fallback to DB",
					zap.String("card_id", cardID),
				)
				missedCardIDs = append(missedCardIDs, cardID)
				continue
			}

			allCards = append(allCards, card)
		}

		// 步骤 5.5：对 cache miss 的卡片，从 DB 构建基础 VitalFocusCard
		if len(missedCardIDs) > 0 {
			dbCards, err := s.buildVitalFocusCardsFromDB(ctx, validatedTenantID, missedCardIDs)
			if err != nil {
				s.logger.Warn("Failed to build cards from DB for cache miss, some cards may be missing",
					zap.Error(err),
					zap.Int("missed_count", len(missedCardIDs)),
				)
			} else {
				allCards = append(allCards, dbCards...)
			}
		}

		// 步骤 6：排序（按 card_name）
		sortCardsByName(allCards)

		// 步骤 7：分页
		total := len(allCards)
		page := req.Page
		if page <= 0 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 10
		}

		start := (page - 1) * pageSize
		if start > total {
			start = total
		}
		end := start + pageSize
		if end > total {
			end = total
		}

		var items []models.VitalFocusCard
		if start < total {
			items = allCards[start:end]
		}

		return &ListVitalFocusCardsResponse{
			Items: items,
			Pagination: Pagination{
				Page:     page,
				PageSize: pageSize,
				Total:    total,
			},
		}, nil
	}

	// 步骤 3：使用辅助函数：resident_id --> resident 表 --> unit_id 列表
	unitIDs, err := s.getUnitsByResidentIDs(ctx, validatedTenantID, assignedResidentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get units by resident IDs: %w", err)
	}

	// 步骤 4：根据卡片生成逻辑，查询符合条件的卡片
	cardIDs, err := s.getCardIDsForStaff(ctx, validatedTenantID, validatedUserID, validatedAlarmScope, unitIDs, assignedResidentIDs, validatedBranchIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get card IDs for staff: %w", err)
	}

	// 步骤 5：从 Redis full cache 读取完整数据，如果 miss 则从 DB 构建基础卡片
	allCards := make([]models.VitalFocusCard, 0, len(cardIDs))
	missedCardIDs := make([]string, 0)
	
	for _, cardID := range cardIDs {
		key := "vital-focus:card:" + cardID + ":full"
		raw, err := s.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				s.logger.Debug("Redis cache miss for card, will fallback to DB",
					zap.String("card_id", cardID),
				)
				missedCardIDs = append(missedCardIDs, cardID)
			} else {
				s.logger.Warn("Failed to get card from Redis, will fallback to DB",
					zap.String("card_id", cardID),
					zap.Error(err),
				)
				missedCardIDs = append(missedCardIDs, cardID)
			}
			continue
		}

		card, ok := decodeAndNormalizeFullCard(raw)
		if !ok {
			s.logger.Warn("Failed to decode and normalize card from Redis, will fallback to DB",
				zap.String("card_id", cardID),
			)
			missedCardIDs = append(missedCardIDs, cardID)
			continue
		}

		// 处理时间字段（bed_status_timestamp 和 status_duration）
		if err := s.enrichCardTimeFields(ctx, &card); err != nil {
			s.logger.Warn("Failed to enrich card time fields, continuing with empty time fields",
				zap.String("card_id", cardID),
				zap.Error(err),
			)
			// 继续处理，时间字段保持为空
		}

		allCards = append(allCards, card)
	}

	// 步骤 5.5：对 cache miss 的卡片，从 DB 构建基础 VitalFocusCard
	if len(missedCardIDs) > 0 {
		dbCards, err := s.buildVitalFocusCardsFromDB(ctx, validatedTenantID, missedCardIDs)
		if err != nil {
			s.logger.Warn("Failed to build cards from DB for cache miss, some cards may be missing",
				zap.Error(err),
				zap.Int("missed_count", len(missedCardIDs)),
			)
		} else {
			allCards = append(allCards, dbCards...)
		}
	}

	// 步骤 6：排序（按 card_name）
	sortCardsByName(allCards)

	// 步骤 7：分页
	total := len(allCards)
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	var items []models.VitalFocusCard
	if start < total {
		items = allCards[start:end]
	}

	return &ListVitalFocusCardsResponse{
		Items: items,
		Pagination: Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

// GetVitalFocusCardRequest 获取单个 Vital Focus 卡片请求
type GetVitalFocusCardRequest struct {
	TenantID string // 租户 ID（前端传入，不信任）
	CardID   string // 卡片 ID（必填）

	// 权限相关参数（前端传入，不信任）
	CurrentUserID   string // 当前用户 ID
	CurrentUserType string // 当前用户类型："resident" | "staff"
	CurrentUserRole string // 当前用户角色
}

// GetVitalFocusCard 获取单个 Vital Focus 卡片（包含实时数据和报警数据）
// 复用权限过滤逻辑，先从 cards 表查询（应用权限过滤），然后从 Redis full cache 读取完整数据
// 注意：假设 Header 中的信息（X-Tenant-Id, X-User-Id, X-User-Role）已经从 JWT token 验证过，可以信任
func (s *cardService) GetVitalFocusCard(ctx context.Context, req GetVitalFocusCardRequest) (*models.VitalFocusCard, error) {
	// 直接使用 Header 中的信息（假设已从 JWT 验证）
	// 注意：这里不再进行数据库验证，因为每2秒调用一次，性能考虑
	validatedTenantID := req.TenantID
	validatedUserID := req.CurrentUserID
	validatedUserRole := req.CurrentUserRole

	// 构建权限过滤请求（复用 GetCardOverview 的逻辑）
	repoReq := repository.ListCardsRequest{
		TenantID: validatedTenantID, // 使用验证后的 tenant_id（信任）
		CardID:   req.CardID,        // 查询指定卡片
	}

	if req.CurrentUserType == "resident" {
		if validatedUserID != "" {
			repoReq.PermissionFilter = &repository.PermissionFilter{
				UserID:   validatedUserID,
				UserType: "resident",
			}
			s.logger.Info("Card permission filter: Resident user for single card",
				zap.String("resident_id", validatedUserID),
				zap.String("tenant_id", validatedTenantID),
				zap.String("card_id", req.CardID),
			)
		}
	} else if req.CurrentUserType == "staff" {
		// 根据角色判断权限逻辑（不依赖 role_permissions 表）
		repoReq.PermissionFilter = &repository.PermissionFilter{}

		switch validatedUserRole {
		case "Admin":
			// Admin: 可以访问任何卡片（tenant_id 过滤），无需额外过滤
			s.logger.Info("Card permission filter: Admin user for single card",
				zap.String("user_id", validatedUserID),
				zap.String("role", validatedUserRole),
				zap.String("tenant_id", validatedTenantID),
				zap.String("card_id", req.CardID),
			)
			// PermissionFilter 保持为空，将返回所有 tenant_id 匹配的卡片

		case "Manager":
			// Manager: 通过 branch_id 过滤
			// 注意：这里需要从 user_branches 表获取 branch_id，而不是从 users.branch_name
			// 但为了简化，这里先使用 ListCards 的权限过滤逻辑
			// TODO: 如果 ListCards 不支持 branch_id 过滤，需要单独实现
			s.logger.Info("Card permission filter: Manager user for single card",
				zap.String("user_id", validatedUserID),
				zap.String("role", validatedUserRole),
				zap.String("tenant_id", validatedTenantID),
				zap.String("card_id", req.CardID),
			)
			// TODO: 实现 Manager 的 branch_id 过滤逻辑

		case "Nurse", "Caregiver":
			// Nurse/Caregiver: 通过 resident_caregivers 表查询分配的 resident_id
			if validatedUserID != "" {
				assignedResidentIDs, err := s.getAssignedResidentIDs(ctx, validatedTenantID, validatedUserID)
				if err != nil {
					return nil, fmt.Errorf("failed to get assigned resident IDs: %w", err)
				}

				if len(assignedResidentIDs) == 0 {
					// 没有分配的住户，返回空
					s.logger.Info("Card permission filter: Nurse/Caregiver user has no assigned residents for single card",
						zap.String("user_id", validatedUserID),
						zap.String("tenant_id", validatedTenantID),
						zap.String("card_id", req.CardID),
					)
					return nil, fmt.Errorf("card not found or access denied")
				}

				repoReq.PermissionFilter.AssignedResidentIDs = assignedResidentIDs
				s.logger.Info("Card permission filter: Nurse/Caregiver user for single card",
					zap.String("user_id", validatedUserID),
					zap.String("role", validatedUserRole),
					zap.String("tenant_id", validatedTenantID),
					zap.String("card_id", req.CardID),
					zap.Int("assigned_resident_count", len(assignedResidentIDs)),
				)
			}

		default:
			// 未知角色：拒绝访问
			s.logger.Warn("Card permission filter: Unknown staff role, denying access",
				zap.String("user_id", validatedUserID),
				zap.String("role", validatedUserRole),
				zap.String("tenant_id", validatedTenantID),
				zap.String("card_id", req.CardID),
			)
			return nil, fmt.Errorf("card not found or access denied")
		}
	} else {
		return nil, fmt.Errorf("invalid user type: %s", req.CurrentUserType)
	}

	// 步骤 3：从 cards 表查询卡片（应用权限过滤）
	// 如果查询结果为空，说明用户没有权限访问该卡片
	cards, err := s.cardsRepo.ListCards(ctx, repoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards for permission check: %w", err)
	}
	if len(cards) == 0 {
		s.logger.Warn("User does not have permission to access card or card not found",
			zap.String("user_id", validatedUserID),
			zap.String("tenant_id", validatedTenantID),
			zap.String("card_id", req.CardID),
		)
		return nil, fmt.Errorf("card not found or access denied")
	}
	// 确保返回的是请求的 card_id
	if cards[0].Card.CardID != req.CardID {
		s.logger.Error("Permission check returned wrong card",
			zap.String("requested_card_id", req.CardID),
			zap.String("returned_card_id", cards[0].Card.CardID),
		)
		return nil, fmt.Errorf("card not found or access denied")
	}

	// 步骤 4：从 Redis full cache 读取完整数据
	key := "vital-focus:card:" + req.CardID + ":full"
	raw, err := s.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, store.ErrMiss) {
			s.logger.Debug("Redis cache miss for card",
				zap.String("card_id", req.CardID),
			)
		} else {
			s.logger.Warn("Failed to get card from Redis",
				zap.String("card_id", req.CardID),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to retrieve real-time data for card: %w", err)
	}

	// 步骤 5：数据规范化（decodeAndNormalizeFullCard）
	card, ok := decodeAndNormalizeFullCard(raw)
	if !ok {
		s.logger.Error("Failed to decode and normalize card from Redis",
			zap.String("card_id", req.CardID),
		)
		return nil, fmt.Errorf("failed to process card data")
	}

	// 处理时间字段（bed_status_timestamp 和 status_duration）
	if err := s.enrichCardTimeFields(ctx, &card); err != nil {
		s.logger.Warn("Failed to enrich card time fields, continuing with empty time fields",
			zap.String("card_id", req.CardID),
			zap.Error(err),
		)
		// 继续处理，时间字段保持为空
	}

	return &card, nil
}

// getAssignedResidentIDs 查询分配给用户的 resident_id 列表
//
// 逻辑说明：
// 1. 从 users 表获取用户的 user_tags（JSONB 数组）
// 2. 从 resident_caregivers 表查询，匹配条件：
//   - user_list JSONB 数组包含该 userID
//   - 或者 group_list JSONB 数组与用户的 user_tags 有交集（使用 ?| 操作符）
//
// 返回：分配给该用户的所有 resident_id 列表（去重）
func (s *cardService) getAssignedResidentIDs(ctx context.Context, tenantID, userID string) ([]string, error) {
	// 步骤 1：获取用户的 user_tags（JSONB 数组）
	var userTagsJSON json.RawMessage
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(user_tags, '[]'::jsonb)
		 FROM users
		 WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userID,
	).Scan(&userTagsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在，返回空列表
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get user tags: %w", err)
	}

	// 步骤 2：解析 user_tags JSONB 为字符串数组
	var userTagsArray []string
	if len(userTagsJSON) > 0 {
		if err := json.Unmarshal(userTagsJSON, &userTagsArray); err != nil {
			s.logger.Warn("Failed to parse user_tags JSONB, treating as empty",
				zap.String("user_id", userID),
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			userTagsArray = []string{}
		}
	}

	// 步骤 3：查询 resident_caregivers 表
	// 匹配条件：
	//   - user_list JSONB 数组包含 userID（使用 @> 操作符检查 JSONB 数组是否包含值）
	//   - 或者 group_list JSONB 数组与 user_tags 有交集（使用 ?| 操作符）
	query := `
		SELECT DISTINCT resident_id::text
		FROM resident_caregivers
		WHERE tenant_id = $1
		  AND (
			-- 条件 A: user_list JSONB 数组包含 userID
			user_list @> $2::jsonb
			OR
			-- 条件 B: group_list JSONB 数组与 user_tags 有交集（如果 user_tags 不为空）
			(
				$3::text[] IS NOT NULL
				AND array_length($3::text[], 1) > 0
				AND group_list IS NOT NULL
				AND group_list ?| $3::text[]
			)
		  )
	`

	// 构建 userID 的 JSONB 值（用于 @> 操作符）
	userIDJSON, _ := json.Marshal([]string{userID})

	rows, err := s.db.QueryContext(ctx, query, tenantID, userIDJSON, pq.Array(userTagsArray))
	if err != nil {
		return nil, fmt.Errorf("failed to query assigned residents: %w", err)
	}
	defer rows.Close()

	// 步骤 4：收集所有 resident_id
	var residentIDs []string
	for rows.Next() {
		var residentID string
		if err := rows.Scan(&residentID); err != nil {
			return nil, fmt.Errorf("failed to scan resident_id: %w", err)
		}
		residentIDs = append(residentIDs, residentID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate assigned residents: %w", err)
	}

	return residentIDs, nil
}

// defaultAlarmScopeForRole 按角色返回默认 alarm_scope（与 07_users.sql 注释一致，为空表示使用角色默认）
func defaultAlarmScopeForRole(role string) string {
	switch role {
	case "SystemAdmin", "Admin", "IT":
		return "ALL"
	case "Manager":
		return "BRANCH"
	case "Nurse", "Caregiver":
		return "ASSIGNED_ONLY"
	default:
		return "ASSIGNED_ONLY"
	}
}

// validateUserSecurity 安全验证 - 根据 X-User-Id 查询数据库，验证 tenant_id, role 是否一致
// 如果不一致，返回错误（中断会话）
// 返回：validatedTenantID, validatedUserID, validatedUserRole, validatedBranchIDs, validatedAlarmScope, error
// 注意：从 user_branches 表查询用户的所有绑定院区 branch_id 列表
// 注意：从 users 表查询 alarm_scope（用于卡片权限过滤）
func (s *cardService) validateUserSecurity(
	ctx context.Context,
	req ListVitalFocusCardsRequest,
) (validatedTenantID, validatedUserID, validatedUserRole string, validatedBranchIDs []string, validatedAlarmScope string, err error) {
	// 1. 根据 CurrentUserType 查询不同的表
	var actualTenantID, actualUserRole, userStatus string

	if req.CurrentUserType == "resident" {
		// 从 residents 表查询
		var residentBranchID sql.NullString
		err = s.db.QueryRowContext(ctx,
			`SELECT tenant_id::text, role, COALESCE(status, 'active'), branch_id::text
			 FROM residents
			 WHERE resident_id = $1`,
			req.CurrentUserID,
		).Scan(&actualTenantID, &actualUserRole, &userStatus, &residentBranchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", "", "", nil, "", fmt.Errorf("resident not found: resident_id=%s", req.CurrentUserID)
			}
			return "", "", "", nil, "", fmt.Errorf("failed to get resident: %w", err)
		}
		if residentBranchID.Valid {
			validatedBranchIDs = []string{residentBranchID.String}
		}
	} else if req.CurrentUserType == "staff" {
		// 从 users 表查询基本信息（包括 alarm_scope）
		var alarmScope sql.NullString
		err = s.db.QueryRowContext(ctx,
			`SELECT tenant_id::text, role, COALESCE(status, 'active'), alarm_scope
			 FROM users
			 WHERE user_id = $1`,
			req.CurrentUserID,
		).Scan(&actualTenantID, &actualUserRole, &userStatus, &alarmScope)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", "", "", nil, "", fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
			}
			return "", "", "", nil, "", fmt.Errorf("failed to get user: %w", err)
		}
		// alarm_scope 未设置时，按角色默认（与 07_users.sql 注释一致）
		if !alarmScope.Valid || alarmScope.String == "" {
			validatedAlarmScope = defaultAlarmScopeForRole(actualUserRole)
			s.logger.Debug("alarm_scope not set for staff, using role default",
				zap.String("user_id", req.CurrentUserID),
				zap.String("role", actualUserRole),
				zap.String("alarm_scope", validatedAlarmScope),
			)
		} else {
			validatedAlarmScope = alarmScope.String
		}

		// 从 user_branches 表查询用户的所有绑定院区 branch_id 列表
		rows, err := s.db.QueryContext(ctx,
			`SELECT branch_id::text
			 FROM user_branches
			 WHERE tenant_id = $1 AND user_id = $2`,
			actualTenantID, req.CurrentUserID,
		)
		if err != nil {
			return "", "", "", nil, "", fmt.Errorf("failed to get user branches: %w", err)
		}
		defer rows.Close()

		var branchIDs []string
		for rows.Next() {
			var branchID string
			if err := rows.Scan(&branchID); err != nil {
				return "", "", "", nil, "", fmt.Errorf("failed to scan branch_id: %w", err)
			}
			branchIDs = append(branchIDs, branchID)
		}
		if err := rows.Err(); err != nil {
			return "", "", "", nil, "", fmt.Errorf("failed to iterate user branches: %w", err)
		}
		validatedBranchIDs = branchIDs
	} else {
		return "", "", "", nil, "", fmt.Errorf("invalid user type: %s", req.CurrentUserType)
	}

	// 2. 验证用户状态是否有效
	if userStatus != "active" {
		s.logger.Warn("User is not active",
			zap.String("user_id", req.CurrentUserID),
			zap.String("user_type", req.CurrentUserType),
			zap.String("status", userStatus),
		)
		return "", "", "", nil, "", fmt.Errorf("user is not active: status=%s", userStatus)
	}

	// 3. 验证前端传入的 TenantID 是否与后端查询的一致
	if req.TenantID != "" && req.TenantID != actualTenantID {
		s.logger.Error("Security violation: tenant_id mismatch",
			zap.String("user_id", req.CurrentUserID),
			zap.String("user_type", req.CurrentUserType),
			zap.String("frontend_tenant_id", req.TenantID),
			zap.String("backend_tenant_id", actualTenantID),
		)
		return "", "", "", nil, "", fmt.Errorf("security violation: tenant_id mismatch")
	}

	// 4. 验证前端传入的 CurrentUserRole 是否与后端查询的一致（仅 staff 用户）
	if req.CurrentUserType == "staff" && req.CurrentUserRole != "" && req.CurrentUserRole != actualUserRole {
		s.logger.Error("Security violation: user_role mismatch",
			zap.String("user_id", req.CurrentUserID),
			zap.String("user_type", req.CurrentUserType),
			zap.String("frontend_role", req.CurrentUserRole),
			zap.String("backend_role", actualUserRole),
		)
		return "", "", "", nil, "", fmt.Errorf("security violation: user_role mismatch")
	}

	// 5. 返回验证后的值
	// 注意：对于 resident 用户，alarm_scope 为空字符串
	validatedAlarmScopeValue := ""
	if req.CurrentUserType == "staff" {
		// 已经在上面查询并赋值了
		validatedAlarmScopeValue = validatedAlarmScope
	}
	return actualTenantID, req.CurrentUserID, actualUserRole, validatedBranchIDs, validatedAlarmScopeValue, nil
}

// getUnitsByResidentIDs 辅助函数：根据 resident_id 列表查询关联的 unit_id 列表
// 逻辑：resident_id --> residents 表 --> unit_id
func (s *cardService) getUnitsByResidentIDs(ctx context.Context, tenantID string, residentIDs []string) ([]string, error) {
	if len(residentIDs) == 0 {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT unit_id::text
		FROM residents
		WHERE tenant_id = $1
		  AND resident_id = ANY($2::uuid[])
		  AND unit_id IS NOT NULL
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(residentIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query units by resident IDs: %w", err)
	}
	defer rows.Close()

	var unitIDs []string
	for rows.Next() {
		var unitID string
		if err := rows.Scan(&unitID); err != nil {
			return nil, fmt.Errorf("failed to scan unit_id: %w", err)
		}
		unitIDs = append(unitIDs, unitID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate units: %w", err)
	}

	return unitIDs, nil
}

// getCardIDsForStaff 根据卡片生成逻辑，查询 staff 用户可见的卡片 ID 列表
//
// 逻辑说明（基于 users.alarm_scope 字段，不依赖 role_permissions 表）：
//   - 'ALL': 返回该租户的所有卡片（tenant_id 过滤）
//   - 'BRANCH': 通过 branch_id 过滤，返回该分支下的所有卡片
//   - 'ASSIGNED_ONLY': 通过 resident_caregivers 表查询分配的 resident_id，然后过滤：
//     Step 1: 如果 unit_id 是非 shared 型的（is_shared_unit = FALSE），直接加入 card list
//     Step 2: 如果 resident_id 与 resident_id 相同，得到 resident_id 的 ActiveBed 卡，直接加入 card list
//
// 注意：需要去重，因为同一个卡片可能同时满足两个条件
func (s *cardService) getCardIDsForStaff(
	ctx context.Context,
	tenantID string,
	userID string, // 用户 ID（用于明确判断和日志）
	alarmScope string, // 用户的 alarm_scope（'ALL' | 'BRANCH' | 'ASSIGNED_ONLY'）
	unitIDs []string,
	assignedResidentIDs []string,
	userBranchIDs []string, // 传入用户的所有绑定分支 ID 列表
) ([]string, error) {
	var query strings.Builder
	var args []any
	argIdx := 1

	// SELECT 子句：只查询 card_id，用于去重
	query.WriteString(`
		SELECT DISTINCT c.card_id::text
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		LEFT JOIN branches br ON u.branch_id = br.branch_id
		WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
	`)
	args = append(args, tenantID)
	argIdx++

	var conditions []string

	switch alarmScope {
	case "ALL":
		// ALL: 返回所有卡片（tenant_id 过滤），无需额外条件
		// conditions 保持为空，将返回所有属于该 tenant 的卡片
		s.logger.Debug("User with ALL alarm_scope: returning all cards for tenant",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
		)

	case "BRANCH":
		// BRANCH: 过滤同分支的卡片（匹配用户的所有绑定院区）
		if len(userBranchIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf(`u.branch_id = ANY($%d::uuid[])`, argIdx))
			args = append(args, pq.Array(userBranchIDs))
			argIdx++
		} else {
			// 如果用户没有绑定任何院区，根据业务规则视为可以访问所有院区（或 NULL 院区）
			// 但为了安全，这里返回空列表
			s.logger.Warn("User with BRANCH alarm_scope has no associated branches, returning empty list",
				zap.String("tenant_id", tenantID),
				zap.String("user_id", userID),
			)
			return []string{}, nil
		}
		s.logger.Debug("User with BRANCH alarm_scope: filtering by branch_ids",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.Strings("user_branch_ids", userBranchIDs),
		)

	case "ASSIGNED_ONLY":
		// ASSIGNED_ONLY: 通过 resident_caregivers 表查询分配的 resident_id，然后过滤
		// 过滤分配的 resident_id 和非共享 unit
		if len(unitIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf(
				`(c.unit_id = ANY($%d::uuid[]) AND u.is_shared_unit = FALSE)`,
				argIdx,
			))
			args = append(args, pq.Array(unitIDs))
			argIdx++
		}

		if len(assignedResidentIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf(
				`(c.card_type = 'ActiveBed' AND c.resident_id = ANY($%d::uuid[]))`,
				argIdx,
			))
			args = append(args, pq.Array(assignedResidentIDs))
			argIdx++
		}

		// 如果没有分配的 unit 和 resident，返回空列表
		if len(conditions) == 0 {
			s.logger.Warn("User with ASSIGNED_ONLY alarm_scope has no assigned units or residents, returning empty list",
				zap.String("tenant_id", tenantID),
				zap.String("user_id", userID),
				zap.String("alarm_scope", alarmScope),
			)
			return []string{}, nil
		}

		s.logger.Debug("User with ASSIGNED_ONLY alarm_scope: filtering by assigned residents",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("alarm_scope", alarmScope),
			zap.Int("unit_ids_count", len(unitIDs)),
			zap.Int("assigned_resident_ids_count", len(assignedResidentIDs)),
		)

	default:
		// 未知的 alarm_scope，返回空列表
		s.logger.Warn("Unknown or empty alarm_scope, returning empty list",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("alarm_scope", alarmScope),
		)
		return []string{}, nil
	}

	// 组合条件：OR 连接
	if len(conditions) > 0 {
		query.WriteString(` AND (`)
		query.WriteString(strings.Join(conditions, " OR "))
		query.WriteString(`)`)
	}
	// Admin 角色：conditions 为空，不需要额外的 WHERE 子句，因为 tenant_id 已经过滤

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query card IDs: %w", err)
	}
	defer rows.Close()

	var cardIDs []string
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			return nil, fmt.Errorf("failed to scan card_id: %w", err)
		}
		cardIDs = append(cardIDs, cardID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate card IDs: %w", err)
	}

	return cardIDs, nil
}

// getCardIDsForResident 根据卡片生成逻辑，查询 resident 可见的卡片 ID 列表
//
// 逻辑说明：
// 1. card_type='ActiveBed' && resident_id相同的 加入
// 2. is_shared_unit = FALSE && unit_id相同的 加入
//
// 注意：需要去重，因为同一个卡片可能同时满足两个条件
func (s *cardService) getCardIDsForResident(
	ctx context.Context,
	tenantID string,
	residentID string,
) ([]string, error) {
	// 步骤 1：获取 resident 的 unit_id
	var unitID sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT unit_id::text
		 FROM residents
		 WHERE tenant_id = $1 AND resident_id = $2`,
		tenantID, residentID,
	).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Resident 不存在，返回空列表
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get resident unit_id: %w", err)
	}

	// 如果 unit_id 为 NULL，说明没有分配，返回空列表
	if !unitID.Valid || unitID.String == "" {
		s.logger.Debug("Resident has no unit_id assigned, returning empty list",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
		)
		return []string{}, nil
	}

	var query strings.Builder
	var args []any
	argIdx := 1

	// SELECT 子句：只查询 card_id，用于去重
	query.WriteString(`
		SELECT DISTINCT c.card_id::text
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
	`)
	args = append(args, tenantID)
	argIdx++

	var conditions []string

	// 条件 1: card_type='ActiveBed' && resident_id相同的
	conditions = append(conditions, fmt.Sprintf(
		`(c.card_type = 'ActiveBed' AND c.resident_id = $%d::uuid)`,
		argIdx,
	))
	args = append(args, residentID)
	argIdx++

	// 条件 2: is_shared_unit = FALSE && unit_id相同的
	conditions = append(conditions, fmt.Sprintf(
		`(c.unit_id = $%d::uuid AND u.is_shared_unit = FALSE)`,
		argIdx,
	))
	args = append(args, unitID.String)
	argIdx++

	// 组合条件：OR 连接
	query.WriteString(` AND (`)
	query.WriteString(strings.Join(conditions, " OR "))
	query.WriteString(`)`)

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query card IDs for resident: %w", err)
	}
	defer rows.Close()

	var cardIDs []string
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			return nil, fmt.Errorf("failed to scan card_id: %w", err)
		}
		cardIDs = append(cardIDs, cardID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate card IDs: %w", err)
	}

	s.logger.Debug("Resident card IDs query result",
		zap.String("tenant_id", tenantID),
		zap.String("resident_id", residentID),
		zap.String("unit_id", unitID.String),
		zap.Int("card_count", len(cardIDs)),
	)

	return cardIDs, nil
}

// sortCardsByName 按 card_name 排序卡片列表
func sortCardsByName(cards []models.VitalFocusCard) {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].CardName < cards[j].CardName
	})
}

// GetVitalFocusPreferencesRequest 获取 Vital Focus preferences 请求
type GetVitalFocusPreferencesRequest struct {
	TenantID      string // 租户 ID
	CurrentUserID string // 当前用户 ID（从 Header 获取，需要验证）
}

// GetVitalFocusPreferencesResponse 获取 Vital Focus preferences 响应
type GetVitalFocusPreferencesResponse struct {
	SelectedCardIDs []string `json:"selected_card_ids"` // 选中的卡片 ID 列表
}

// GetVitalFocusPreferences 获取用户的 Vital Focus preferences（selectedCardIds）
func (s *cardService) GetVitalFocusPreferences(ctx context.Context, req GetVitalFocusPreferencesRequest) (*GetVitalFocusPreferencesResponse, error) {
	// STD: 输入参数日志
	s.logger.Info("[GetVitalFocusPreferences] INPUT",
		zap.String("tenant_id", req.TenantID),
		zap.String("current_user_id", req.CurrentUserID),
	)

	// 1. 安全验证：验证用户是否存在且属于该租户
	var actualTenantID string
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM users WHERE user_id = $1`,
		req.CurrentUserID,
	).Scan(&actualTenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Error("[GetVitalFocusPreferences] User not found",
				zap.String("user_id", req.CurrentUserID),
			)
			return nil, fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. 验证前端传入的 TenantID 是否与后端查询的一致
	if req.TenantID != "" && req.TenantID != actualTenantID {
		s.logger.Error("Security violation: tenant_id mismatch",
			zap.String("user_id", req.CurrentUserID),
			zap.String("frontend_tenant_id", req.TenantID),
			zap.String("backend_tenant_id", actualTenantID),
		)
		return nil, fmt.Errorf("security violation: tenant_id mismatch")
	}

	// 3. 查询用户的 preferences
	var preferencesJSON sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT preferences FROM users WHERE tenant_id = $1 AND user_id = $2`,
		actualTenantID, req.CurrentUserID,
	).Scan(&preferencesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("[GetVitalFocusPreferences] User not found in preferences query",
				zap.String("user_id", req.CurrentUserID),
				zap.String("tenant_id", actualTenantID),
			)
			return nil, fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// STD: 从数据库获取的原始 preferences JSON
	prefsJSONStr := "<NULL>"
	if preferencesJSON.Valid {
		prefsJSONStr = preferencesJSON.String
	}
	s.logger.Info("[GetVitalFocusPreferences] Raw preferences from DB",
		zap.String("user_id", req.CurrentUserID),
		zap.Bool("preferences_valid", preferencesJSON.Valid),
		zap.String("preferences_json", prefsJSONStr),
	)

	// 4. 解析 preferences JSON
	var selectedCardIDs []string
	if preferencesJSON.Valid && preferencesJSON.String != "" {
		var prefs map[string]interface{}
		if err := json.Unmarshal([]byte(preferencesJSON.String), &prefs); err == nil {
			if vitalFocus, ok := prefs["vitalFocus"].(map[string]interface{}); ok {
				if cardIds, ok := vitalFocus["selectedCardIds"].([]interface{}); ok {
					for _, id := range cardIds {
						if idStr, ok := id.(string); ok {
							selectedCardIDs = append(selectedCardIDs, idStr)
						}
					}
				} else {
					s.logger.Warn("[GetVitalFocusPreferences] selectedCardIds not found or not array in vitalFocus",
						zap.String("user_id", req.CurrentUserID),
					)
				}
			} else {
				s.logger.Warn("[GetVitalFocusPreferences] vitalFocus not found in preferences",
					zap.String("user_id", req.CurrentUserID),
				)
			}
		} else {
			s.logger.Warn("[GetVitalFocusPreferences] Failed to unmarshal preferences JSON",
				zap.String("user_id", req.CurrentUserID),
				zap.Error(err),
			)
		}
	} else {
		s.logger.Info("[GetVitalFocusPreferences] Preferences is NULL or empty",
			zap.String("user_id", req.CurrentUserID),
		)
	}

	// STD: 输出结果日志
	s.logger.Info("[GetVitalFocusPreferences] OUTPUT",
		zap.String("user_id", req.CurrentUserID),
		zap.Int("selected_card_ids_count", len(selectedCardIDs)),
		zap.Strings("selected_card_ids", selectedCardIDs),
	)

	return &GetVitalFocusPreferencesResponse{
		SelectedCardIDs: selectedCardIDs,
	}, nil
}

// SaveVitalFocusPreferencesRequest 保存 Vital Focus preferences 请求
type SaveVitalFocusPreferencesRequest struct {
	TenantID        string   // 租户 ID（前端传入，不信任）
	CurrentUserID   string   // 当前用户 ID（从 Header 获取，需要验证）
	SelectedCardIDs []string // 选中的卡片 ID 列表
}

// SaveVitalFocusPreferences 保存用户的 Vital Focus preferences（selectedCardIds）
func (s *cardService) SaveVitalFocusPreferences(ctx context.Context, req SaveVitalFocusPreferencesRequest) error {
	// STD: 输入参数日志
	s.logger.Info("[SaveVitalFocusPreferences] INPUT",
		zap.String("tenant_id", req.TenantID),
		zap.String("current_user_id", req.CurrentUserID),
		zap.Int("selected_card_ids_count", len(req.SelectedCardIDs)),
		zap.Strings("selected_card_ids", req.SelectedCardIDs),
	)

	// 1. 安全验证：验证用户是否存在且属于该租户
	var actualTenantID string
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM users WHERE user_id = $1`,
		req.CurrentUserID,
	).Scan(&actualTenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Error("[SaveVitalFocusPreferences] User not found",
				zap.String("user_id", req.CurrentUserID),
			)
			return fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 2. 验证前端传入的 TenantID 是否与后端查询的一致
	if req.TenantID != "" && req.TenantID != actualTenantID {
		s.logger.Error("Security violation: tenant_id mismatch",
			zap.String("user_id", req.CurrentUserID),
			zap.String("frontend_tenant_id", req.TenantID),
			zap.String("backend_tenant_id", actualTenantID),
		)
		return fmt.Errorf("security violation: tenant_id mismatch")
	}

	// 3. 获取当前用户的 preferences
	var currentPrefsJSON sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT preferences FROM users WHERE tenant_id = $1 AND user_id = $2`,
		actualTenantID, req.CurrentUserID,
	).Scan(&currentPrefsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Error("[SaveVitalFocusPreferences] User not found in preferences query",
				zap.String("user_id", req.CurrentUserID),
				zap.String("tenant_id", actualTenantID),
			)
			return fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
		}
		return fmt.Errorf("failed to get user preferences: %w", err)
	}

	// STD: 从数据库获取的原始 preferences（保存前）
	currentPrefsStr := "<NULL>"
	if currentPrefsJSON.Valid {
		currentPrefsStr = currentPrefsJSON.String
	}
	s.logger.Info("[SaveVitalFocusPreferences] Current preferences from DB (before update)",
		zap.String("user_id", req.CurrentUserID),
		zap.Bool("preferences_valid", currentPrefsJSON.Valid),
		zap.String("preferences_json", currentPrefsStr),
	)

	// 4. 解析现有 preferences
	var prefs map[string]interface{}
	if currentPrefsJSON.Valid && currentPrefsJSON.String != "" {
		if err := json.Unmarshal([]byte(currentPrefsJSON.String), &prefs); err != nil {
			s.logger.Warn("[SaveVitalFocusPreferences] Failed to unmarshal current preferences, creating new",
				zap.String("user_id", req.CurrentUserID),
				zap.Error(err),
			)
			prefs = make(map[string]interface{})
		}
	} else {
		prefs = make(map[string]interface{})
	}

	// 5. 更新 vitalFocus.selectedCardIds
	if prefs["vitalFocus"] == nil {
		prefs["vitalFocus"] = make(map[string]interface{})
	}
	vitalFocus, ok := prefs["vitalFocus"].(map[string]interface{})
	if !ok {
		vitalFocus = make(map[string]interface{})
		prefs["vitalFocus"] = vitalFocus
	}
	vitalFocus["selectedCardIds"] = req.SelectedCardIDs

	// 6. 保存回数据库
	updatedPrefsJSON, err := json.Marshal(prefs)
	if err != nil {
		s.logger.Error("[SaveVitalFocusPreferences] Failed to marshal updated preferences",
			zap.String("user_id", req.CurrentUserID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	// STD: 待保存的 preferences JSON
	s.logger.Info("[SaveVitalFocusPreferences] Updated preferences JSON (to be saved)",
		zap.String("user_id", req.CurrentUserID),
		zap.String("updated_preferences_json", string(updatedPrefsJSON)),
	)

	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET preferences = $1::jsonb WHERE tenant_id = $2 AND user_id = $3`,
		string(updatedPrefsJSON), actualTenantID, req.CurrentUserID,
	)
	if err != nil {
		s.logger.Error("[SaveVitalFocusPreferences] Failed to update user preferences in DB",
			zap.String("user_id", req.CurrentUserID),
			zap.String("tenant_id", actualTenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	// STD: 输出结果日志
	s.logger.Info("[SaveVitalFocusPreferences] OUTPUT - Saved successfully",
		zap.String("user_id", req.CurrentUserID),
		zap.String("tenant_id", actualTenantID),
		zap.Int("selected_card_count", len(req.SelectedCardIDs)),
		zap.Strings("selected_card_ids", req.SelectedCardIDs),
		zap.String("saved_preferences_json", string(updatedPrefsJSON)),
	)

	return nil
}

// buildVitalFocusCardsFromDB 从数据库构建基础的 VitalFocusCard（当 Redis cache miss 时使用）
// 返回基础的卡片信息，实时数据字段（如报警统计、设备状态等）为空或使用默认值
func (s *cardService) buildVitalFocusCardsFromDB(ctx context.Context, tenantID string, cardIDs []string) ([]models.VitalFocusCard, error) {
	if len(cardIDs) == 0 {
		return []models.VitalFocusCard{}, nil
	}

	// 批量从 DB 查询卡片基础信息（使用 SQL IN 查询）
	query := `
		SELECT 
			c.card_id::text,
			c.tenant_id::text,
			c.card_type,
			c.bed_id::text,
			c.unit_id::text,
			c.card_name,
			c.card_address,
			c.resident_id::text,
			c.devices,
			c.residents,
			c.unhandled_alarm_0,
			c.unhandled_alarm_1,
			c.unhandled_alarm_2,
			c.unhandled_alarm_3,
			c.unhandled_alarm_4,
			c.icon_alarm_level,
			c.pop_alarm_emerge
		FROM cards c
		WHERE c.tenant_id = $1 AND c.card_id = ANY($2::uuid[])
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(cardIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query cards from DB: %w", err)
	}
	defer rows.Close()

	var cards []*domain.CardWithUnitInfo
	for rows.Next() {
		var card domain.CardWithUnitInfo
		var bedID, unitID, residentID sql.NullString
		var devicesJSON, residentsJSON json.RawMessage

		err := rows.Scan(
			&card.Card.CardID,
			&card.Card.TenantID,
			&card.Card.CardType,
			&bedID,
			&unitID,
			&card.Card.CardName,
			&card.Card.CardAddress,
			&residentID,
			&devicesJSON,
			&residentsJSON,
			&card.Card.UnhandledAlarm0,
			&card.Card.UnhandledAlarm1,
			&card.Card.UnhandledAlarm2,
			&card.Card.UnhandledAlarm3,
			&card.Card.UnhandledAlarm4,
			&card.Card.IconAlarmLevel,
			&card.Card.PopAlarmEmerge,
		)
		if err != nil {
			s.logger.Warn("Failed to scan card row, skipping",
				zap.Error(err),
			)
			continue
		}

		if bedID.Valid {
			card.Card.BedID = bedID
		}
		if unitID.Valid {
			card.Card.UnitID = unitID
		}
		if residentID.Valid {
			card.Card.ResidentID = residentID
		}
		card.Card.Devices = devicesJSON
		card.Card.Residents = residentsJSON

		cards = append(cards, &card)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate card rows: %w", err)
	}

	// 构建基础的 VitalFocusCard
	result := make([]models.VitalFocusCard, 0, len(cards))
	for _, card := range cards {
		vitalCard := models.VitalFocusCard{
			CardID:     card.Card.CardID,
			TenantID:   card.Card.TenantID,
			CardType:   card.Card.CardType,
			CardName:   card.Card.CardName,
			CardAddress: card.Card.CardAddress,
			Residents:  []models.CardResident{},
			Devices:    []models.CardDevice{},
		}

		// 规范化 card_type（数据库使用 'Location'，API 返回 'Unit'）
		if vitalCard.CardType == "Location" {
			vitalCard.CardType = "Unit"
		}

		// 设置 BedID/LocationID
		if card.Card.BedID.Valid {
			vitalCard.BedID = card.Card.BedID.String
		}
		if card.Card.UnitID.Valid {
			vitalCard.LocationID = card.Card.UnitID.String
		}
		if card.Card.ResidentID.Valid {
			vitalCard.PrimaryResidentID = card.Card.ResidentID.String
		}

		// 从 card 表获取报警统计（作为基础值）
		if card.Card.UnhandledAlarm0 > 0 {
			vitalCard.UnhandledAlarm0 = &card.Card.UnhandledAlarm0
		}
		if card.Card.UnhandledAlarm1 > 0 {
			vitalCard.UnhandledAlarm1 = &card.Card.UnhandledAlarm1
		}
		if card.Card.UnhandledAlarm2 > 0 {
			vitalCard.UnhandledAlarm2 = &card.Card.UnhandledAlarm2
		}
		if card.Card.UnhandledAlarm3 > 0 {
			vitalCard.UnhandledAlarm3 = &card.Card.UnhandledAlarm3
		}
		if card.Card.UnhandledAlarm4 > 0 {
			vitalCard.UnhandledAlarm4 = &card.Card.UnhandledAlarm4
		}
		if card.Card.IconAlarmLevel > 0 {
			vitalCard.IconAlarmLevel = &card.Card.IconAlarmLevel
		}
		if card.Card.PopAlarmEmerge > 0 {
			vitalCard.PopAlarmEmerge = &card.Card.PopAlarmEmerge
		}

		// 解析 devices JSONB
		var devicesFromCard []map[string]interface{}
		if err := json.Unmarshal(card.Card.Devices, &devicesFromCard); err == nil {
			for _, deviceObj := range devicesFromCard {
				deviceUID, _ := deviceObj["device_uid"].(string)
				if deviceUID == "" {
					deviceUID, _ = deviceObj["uid"].(string)
				}
				if deviceUID == "" {
					deviceUID, _ = deviceObj["device_id"].(string)
				}
				deviceName, _ := deviceObj["device_name"].(string)
				deviceTypeStr, _ := deviceObj["device_type"].(string)
				deviceModel, _ := deviceObj["device_model"].(string)

				if deviceUID != "" {
					var deviceTypeNum interface{} = nil
					if deviceTypeStr != "" {
						deviceTypeLower := strings.ToLower(deviceTypeStr)
						if deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad" {
							deviceTypeNum = 1
						} else if deviceTypeLower == "radar" {
							deviceTypeNum = 2
						}
					}

					deviceID, _ := deviceObj["device_id"].(string)
					if deviceID == "" {
						deviceID = deviceUID
					}

					vitalCard.Devices = append(vitalCard.Devices, models.CardDevice{
						DeviceID:    deviceID,
						DeviceName:  deviceName,
						DeviceType:  deviceTypeNum,
						DeviceModel: deviceModel,
					})
				}
			}
		}
		vitalCard.DeviceCount = len(vitalCard.Devices)

		// 解析 residents JSONB
		var residentsFromCard []map[string]interface{}
		if card.Card.ResidentID.Valid {
			// ActiveBed: 使用 resident_id 字段
			vitalCard.Residents = append(vitalCard.Residents, models.CardResident{
				ResidentID: card.Card.ResidentID.String,
			})
		} else if err := json.Unmarshal(card.Card.Residents, &residentsFromCard); err == nil {
			for _, residentObj := range residentsFromCard {
				residentID, _ := residentObj["resident_id"].(string)
				nickname, _ := residentObj["nickname"].(string)
				if residentID != "" {
					vitalCard.Residents = append(vitalCard.Residents, models.CardResident{
						ResidentID: residentID,
						Nickname:   nickname,
					})
				}
			}
		}
		vitalCard.ResidentCount = len(vitalCard.Residents)

		result = append(result, vitalCard)
	}

	return result, nil
}
