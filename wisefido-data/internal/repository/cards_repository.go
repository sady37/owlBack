package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// CardsRepository 卡片Repository接口
type CardsRepository interface {
	// ListCards 查询卡片列表（返回所有可见的卡片，不分页）
	ListCards(ctx context.Context, req ListCardsRequest) ([]*domain.CardWithUnitInfo, error)
}

// ListCardsRequest 查询卡片列表请求
type ListCardsRequest struct {
	TenantID        string
	CardID          string // 可选：查询单个卡片
	Search          string // 搜索关键词
	CardType        string // "ActiveBed" | "Unit"
	UnitType        string // "Home" | "Facility"
	IsPublicSpace   *bool
	IsMultiPersonRoom *bool
	Sort            string // "card_name" | "card_address"
	Direction       string // "asc" | "desc"

	// 权限过滤参数（可选）
	PermissionFilter *PermissionFilter
}

// PermissionFilter 权限过滤参数
type PermissionFilter struct {
	// Resident/Family 权限
	UserID   string // resident_id（Family 用户需要先转换）
	UserType string // "resident" | "family"

	// BranchOnly 权限
	UserBranchTag *string // 如果指定，只返回同分支的卡片

	// AssignedOnly 权限
	AssignedOnly        bool   // 如果为 true，在 SQL 中使用 CTE 过滤
	UserIDForAssignment string // 用于 AssignedOnly 过滤的用户 ID
}

// PostgresCardsRepository 卡片Repository实现
type PostgresCardsRepository struct {
	db *sql.DB
}

// NewPostgresCardsRepository 创建卡片Repository
func NewPostgresCardsRepository(db *sql.DB) *PostgresCardsRepository {
	return &PostgresCardsRepository{db: db}
}

// ListCards 查询卡片列表
func (r *PostgresCardsRepository) ListCards(ctx context.Context, req ListCardsRequest) ([]*domain.CardWithUnitInfo, error) {
	if req.TenantID == "" {
		return []*domain.CardWithUnitInfo{}, nil
	}

	// 构建 SQL 查询
	var query strings.Builder
	var args []any
	argIdx := 1

	// 如果 AssignedOnly，使用 CTE
	if req.PermissionFilter != nil && req.PermissionFilter.AssignedOnly {
		query.WriteString(`
			WITH assigned_residents AS (
				SELECT DISTINCT rc.resident_id
				FROM resident_caregivers rc
				WHERE rc.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
					AND (
						-- 检查 userList JSONB 是否包含 userID
						rc.userList::text LIKE '%"' || $` + fmt.Sprintf("%d", argIdx+1) + ` || '"%'
						OR
						-- 检查 groupList JSONB 是否匹配用户的 tags
						EXISTS (
							SELECT 1 FROM users u
							WHERE u.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
								AND u.user_id::text = $` + fmt.Sprintf("%d", argIdx+1) + `
								AND u.tags ?| (
									SELECT ARRAY(SELECT jsonb_array_elements_text(rc.groupList))
								)
						)
					)
			)
		`)
		args = append(args, req.TenantID, req.PermissionFilter.UserIDForAssignment)
		argIdx += 2
	}

	// SELECT 子句
	query.WriteString(`
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
			c.pop_alarm_emerge,
			u.unit_id::text,
			u.tenant_id::text,
			u.branch_name,
			u.unit_name,
			u.building,
			u.floor,
			u.area_name,
			u.unit_number,
			u.layout_config,
			u.unit_type,
			u.is_public_space,
			u.is_multi_person_room,
			u.timezone,
			u.groupList,
			u.userList
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
	`)
	args = append(args, req.TenantID)
	argIdx++

	// 单个卡片查询
	if req.CardID != "" {
		query.WriteString(` AND c.card_id::text = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, req.CardID)
		argIdx++
	}

	// 搜索过滤
	if req.Search != "" {
		query.WriteString(` AND (c.card_name LIKE $` + fmt.Sprintf("%d", argIdx) + ` OR c.card_address LIKE $` + fmt.Sprintf("%d", argIdx) + `) `)
		args = append(args, "%"+req.Search+"%")
		argIdx++
	}

	// 类型过滤
	if req.CardType != "" {
		query.WriteString(` AND c.card_type = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, req.CardType)
		argIdx++
	}

	if req.UnitType != "" {
		query.WriteString(` AND u.unit_type = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, req.UnitType)
		argIdx++
	}

	if req.IsPublicSpace != nil {
		query.WriteString(` AND u.is_public_space = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, *req.IsPublicSpace)
		argIdx++
	}

	if req.IsMultiPersonRoom != nil {
		query.WriteString(` AND u.is_multi_person_room = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, *req.IsMultiPersonRoom)
		argIdx++
	}

	// Resident/Family 权限过滤
	if req.PermissionFilter != nil && req.PermissionFilter.UserID != "" {
		query.WriteString(` AND (
			-- ActiveBed 卡片：直接匹配 resident_id
			(c.card_type = 'ActiveBed' AND c.resident_id::text = $` + fmt.Sprintf("%d", argIdx) + `)
			OR
			-- Unit 卡片（数据库中使用 'Location'）：检查权限
			(c.card_type = 'Location' 
				-- 不是 share unit
				AND (u.is_public_space = FALSE AND u.is_multi_person_room = FALSE)
				-- 是第一个住户或第二个住户（且第二个住户允许）
				AND (
					-- 第一个住户
					c.residents->0 = to_jsonb($` + fmt.Sprintf("%d", argIdx) + `::text)
					OR
					-- 第二个住户（且允许）
					(
						jsonb_array_length(c.residents) >= 2
						AND c.residents->1 = to_jsonb($` + fmt.Sprintf("%d", argIdx) + `::text)
						AND EXISTS (
							SELECT 1 FROM residents r
							WHERE r.tenant_id = c.tenant_id
								AND r.resident_id::text = $` + fmt.Sprintf("%d", argIdx) + `
								AND r.can_view_status = TRUE
						)
					)
				)
			)
		) `)
		args = append(args, req.PermissionFilter.UserID)
		argIdx++
	}

	// BranchOnly 权限过滤
	if req.PermissionFilter != nil && req.PermissionFilter.UserBranchTag != nil {
		userBranchTag := req.PermissionFilter.UserBranchTag
		if *userBranchTag == "" {
			// 用户 branch_tag 为 NULL：只能查看 unit.branch_name 为 NULL 的卡片
			query.WriteString(` AND (u.branch_name IS NULL OR u.branch_name = '-') `)
		} else {
			query.WriteString(` AND u.branch_name = $` + fmt.Sprintf("%d", argIdx) + ` `)
			args = append(args, *userBranchTag)
			argIdx++
		}
	}

	// AssignedOnly 权限过滤
	if req.PermissionFilter != nil && req.PermissionFilter.AssignedOnly {
		query.WriteString(` AND (
			-- ActiveBed 卡片：检查 resident_id 是否在分配列表中
			(c.card_type = 'ActiveBed' AND c.resident_id IN (SELECT resident_id FROM assigned_residents))
			OR
			-- Unit 卡片（数据库中使用 'Location'）：检查权限
			(c.card_type = 'Location' 
				-- 不是 share unit
				AND (u.is_public_space = FALSE AND u.is_multi_person_room = FALSE)
				-- 是第一个住户或第二个住户（且第二个住户允许）
				AND EXISTS (
					SELECT 1 FROM assigned_residents ar
					WHERE (
						c.residents->0 = to_jsonb(ar.resident_id::text)
						OR
						(
							jsonb_array_length(c.residents) >= 2
							AND c.residents->1 = to_jsonb(ar.resident_id::text)
							AND EXISTS (
								SELECT 1 FROM residents r
								WHERE r.tenant_id = c.tenant_id
									AND r.resident_id = ar.resident_id
									AND r.can_view_status = TRUE
							)
						)
					)
				)
			)
		) `)
	}

	// 排序
	sortField := "card_name"
	if req.Sort != "" {
		sortField = req.Sort
	}
	direction := "ASC"
	if req.Direction == "desc" {
		direction = "DESC"
	}
	query.WriteString(` ORDER BY c.` + sortField + ` ` + direction)

	// 执行查询
	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var results []*domain.CardWithUnitInfo
	for rows.Next() {
		card := &domain.Card{}
		unit := &domain.Unit{}

		var bedID, unitID, residentID sql.NullString
		var devicesRaw, residentsRaw sql.NullString
		var branchTag, areaTag, layoutConfig, groupList, userList sql.NullString

		err := rows.Scan(
			&card.CardID,
			&card.TenantID,
			&card.CardType,
			&bedID,
			&unitID,
			&card.CardName,
			&card.CardAddress,
			&residentID,
			&devicesRaw,
			&residentsRaw,
			&card.UnhandledAlarm0,
			&card.UnhandledAlarm1,
			&card.UnhandledAlarm2,
			&card.UnhandledAlarm3,
			&card.UnhandledAlarm4,
			&card.IconAlarmLevel,
			&card.PopAlarmEmerge,
			&unit.UnitID,
			&unit.TenantID,
			&branchTag,
			&unit.UnitName,
			&unit.Building,
			&unit.Floor,
			&areaTag,
			&unit.UnitNumber,
			&layoutConfig,
			&unit.UnitType,
			&unit.IsPublicSpace,
			&unit.IsMultiPersonRoom,
			&unit.Timezone,
			&groupList,
			&userList,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// 设置 nullable 字段
		if bedID.Valid {
			card.BedID = sql.NullString{String: bedID.String, Valid: true}
		}
		if unitID.Valid {
			card.UnitID = sql.NullString{String: unitID.String, Valid: true}
		}
		if residentID.Valid {
			card.ResidentID = sql.NullString{String: residentID.String, Valid: true}
		}
		if devicesRaw.Valid {
			card.Devices = json.RawMessage(devicesRaw.String)
		} else {
			card.Devices = json.RawMessage("[]")
		}
		if residentsRaw.Valid {
			card.Residents = json.RawMessage(residentsRaw.String)
		} else {
			card.Residents = json.RawMessage("[]")
		}

		// 设置 Unit 的 nullable 字段
		if branchTag.Valid {
			unit.BranchName = sql.NullString{String: branchTag.String, Valid: true}
		}
		if areaTag.Valid {
			unit.AreaName = sql.NullString{String: areaTag.String, Valid: true}
		}
		if layoutConfig.Valid {
			unit.LayoutConfig = sql.NullString{String: layoutConfig.String, Valid: true}
		}
		if groupList.Valid {
			unit.GroupList = sql.NullString{String: groupList.String, Valid: true}
		}
		if userList.Valid {
			unit.UserList = sql.NullString{String: userList.String, Valid: true}
		}

		results = append(results, &domain.CardWithUnitInfo{
			Card: card,
			Unit: unit,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}

	return results, nil
}

// 确保实现了接口
var _ CardsRepository = (*PostgresCardsRepository)(nil)

