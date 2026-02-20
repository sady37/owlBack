package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
)

// CardDeviceItem 卡片上的设备项，供 vue-radar 画布 Bind 使用。device_code：Sleepace 有、Radar 无，空则回退 device_uid。
type CardDeviceItem struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	DeviceUID  string `json:"device_uid"`
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
}

// CardsRepository 卡片Repository接口
type CardsRepository interface {
	// ListCards 查询卡片列表（返回所有可见的卡片，不分页）
	ListCards(ctx context.Context, req ListCardsRequest) ([]*domain.CardWithUnitInfo, error)
	// GetCardDevices 获取卡片上所有设备（cards.devices JSONB），供 radar-trajectory 画布 Bind 使用
	GetCardDevices(ctx context.Context, tenantID, cardID string) ([]CardDeviceItem, error)
	// GetCardIDByDeviceID 根据 device_id 查找其所属卡片的 card_id（devices JSONB 中某元素 device_id 匹配）
	GetCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error)
}

// ListCardsRequest 查询卡片列表请求
type ListCardsRequest struct {
	TenantID      string
	CardID        string // 可选：查询单个卡片
	Search        string // 搜索关键词
	CardType      string // "ActiveBedCard" | "UnitCard"
	UnitType      string // "Home" | "Facility"
	IsPublicSpace *bool
	IsSharedUnit  *bool
	Sort          string // "card_name" | "card_address"
	Direction     string // "asc" | "desc"

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
	AssignedResidentIDs []string // 分配给用户的 resident_id 列表（由 service 层查询后传入）
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
			c.pop_alarm,
			u.unit_id::text,
			u.tenant_id::text,
			COALESCE(br.branch_name, NULL) as branch_name,
			u.unit_name,
			COALESCE(bld.building_name, NULL) as building_name,
			u.floor,
			u.layout_config,
			u.unit_type,
			u.is_public,
			u.is_shared_unit,
			u.timezone
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		LEFT JOIN branches br ON u.branch_id = br.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
	`)

	// AssignedOnly 权限过滤：使用简单的 WHERE 子句，基于 service 层传入的 resident_id 列表
	if req.PermissionFilter != nil && len(req.PermissionFilter.AssignedResidentIDs) > 0 {
		// 使用 ANY 数组匹配，逻辑更清晰
		query.WriteString(` AND (
			-- ActiveBedCard 卡片：直接匹配 resident_id
			(c.card_type = 'ActiveBedCard' AND c.resident_id = ANY($` + fmt.Sprintf("%d", argIdx) + `::uuid[]))
			OR
			-- UnitCard 卡片：检查 residents JSONB 数组是否包含任何分配的 resident_id
			(c.card_type = 'UnitCard' 
				AND (u.is_public = FALSE AND u.is_shared_unit = FALSE)
				AND (
					-- 第一个住户匹配
					(c.residents->0->>'resident_id')::uuid = ANY($` + fmt.Sprintf("%d", argIdx) + `::uuid[])
					OR
					-- 第二个住户匹配（如果存在且允许）
					(
						jsonb_array_length(c.residents) >= 2
						AND (c.residents->1->>'resident_id')::uuid = ANY($` + fmt.Sprintf("%d", argIdx) + `::uuid[])
						AND EXISTS (
							SELECT 1 FROM residents r
							WHERE r.tenant_id = c.tenant_id
								AND r.resident_id::text = (c.residents->1->>'resident_id')::text
								AND r.is_access_enabled = TRUE
						)
					)
				)
			)
		) `)
		args = append(args, pq.Array(req.PermissionFilter.AssignedResidentIDs))
		argIdx++
	}

	query.WriteString(` WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + ` `)
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
		query.WriteString(` AND u.is_public = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, *req.IsPublicSpace)
		argIdx++
	}

	if req.IsSharedUnit != nil {
		query.WriteString(` AND u.is_shared_unit = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, *req.IsSharedUnit)
		argIdx++
	}

	// Resident/Family 权限过滤
	if req.PermissionFilter != nil && req.PermissionFilter.UserID != "" {
		query.WriteString(` AND (
			-- ActiveBedCard 卡片：直接匹配 resident_id
			(c.card_type = 'ActiveBedCard' AND c.resident_id::text = $` + fmt.Sprintf("%d", argIdx) + `)
			OR
			-- Unit 卡片（数据库中使用 'UnitCard'）：检查权限
			(c.card_type = 'UnitCard' 
				-- 不是 share unit
				AND (u.is_public = FALSE AND u.is_shared_unit = FALSE)
				-- 是第一个住户或第二个住户（且第二个住户允许）
				AND (
					-- 第一个住户（从 JSONB 对象中提取 resident_id）
					(c.residents->0->>'resident_id')::text = $` + fmt.Sprintf("%d", argIdx) + `
					OR
					-- 第二个住户（且允许）
					(
						jsonb_array_length(c.residents) >= 2
						AND (c.residents->1->>'resident_id')::text = $` + fmt.Sprintf("%d", argIdx) + `
						AND EXISTS (
							SELECT 1 FROM residents r
							WHERE r.tenant_id = c.tenant_id
								AND r.resident_id::text = $` + fmt.Sprintf("%d", argIdx) + `
								AND r.is_access_enabled = TRUE
						)
					)
				)
			)
		) `)
		args = append(args, req.PermissionFilter.UserID)
		argIdx++
	}

	// BranchOnly 权限过滤（通过 JOIN branches 表获取 branch_name）
	if req.PermissionFilter != nil && req.PermissionFilter.UserBranchTag != nil {
		userBranchTag := req.PermissionFilter.UserBranchTag
		if *userBranchTag == "" {
			// 用户 branch_tag 为 NULL：只能查看 unit.branch_name 为 NULL 的卡片
			query.WriteString(` AND (br.branch_name IS NULL OR br.branch_name = '-') `)
		} else {
			query.WriteString(` AND br.branch_name = $` + fmt.Sprintf("%d", argIdx) + ` `)
			args = append(args, *userBranchTag)
			argIdx++
		}
	}

	// AssignedOnly 权限过滤已在上面处理（使用 AssignedResidentIDs）

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
		var branchTag, buildingTag, layoutConfig sql.NullString
		var isPublic, isSharedUnit bool

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
			&card.PopAlarm,
			&unit.UnitID,
			&unit.TenantID,
			&branchTag,
			&unit.UnitName,
			&buildingTag,
			&unit.Floor,
			&layoutConfig,
			&unit.UnitType,
			&isPublic,
			&isSharedUnit,
			&unit.Timezone,
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
		if buildingTag.Valid {
			unit.BuildingName = sql.NullString{String: buildingTag.String, Valid: true}
		}
		if layoutConfig.Valid {
			unit.LayoutConfig = sql.NullString{String: layoutConfig.String, Valid: true}
		}
		// 设置 Unit 的 boolean 字段（注意：domain.Unit 中的字段名与数据库不一致，但这里先按数据库字段名设置）
		unit.IsPublic = isPublic
		unit.IsSharedUnit = isSharedUnit

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

// GetCardDevices 获取卡片上所有设备（cards.devices JSONB），供 radar-trajectory 画布 Bind 使用。
// 从每个元素 m 读 device_uid、uid；若 JSON 无则 DeviceUID 为空，不做降级。写 cards.devices 时从 devices 表选 d.device_uid 填入。
func (r *PostgresCardsRepository) GetCardDevices(ctx context.Context, tenantID, cardID string) ([]CardDeviceItem, error) {
	if tenantID == "" || cardID == "" {
		return nil, nil
	}
	q := `SELECT devices FROM cards WHERE card_id = $1 AND tenant_id = $2`
	var raw json.RawMessage
	if err := r.db.QueryRowContext(ctx, q, cardID, tenantID).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("get card devices: %w", err)
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("parse card devices JSON: %w", err)
	}
	out := make([]CardDeviceItem, 0, len(arr))
	for _, m := range arr {
		uid, _ := m["device_uid"].(string)
		if uid == "" {
			uid, _ = m["uid"].(string)
		}
		name, _ := m["device_name"].(string)
		code, _ := m["device_code"].(string)
		if code == "" {
			code = uid // device_code 仅 Sleepace 有，Radar 无；空则用 device_uid
		}
		did, _ := m["device_id"].(string)
		if did == "" {
			did = uid
		}
		dtype := ""
		switch v := m["device_type"].(type) {
		case float64:
			if v == 1 {
				dtype = "Sleepad"
			} else if v == 2 {
				dtype = "Radar"
			}
		case string:
			dtype = v
		}
		out = append(out, CardDeviceItem{DeviceID: did, DeviceType: dtype, DeviceUID: uid, DeviceCode: code, DeviceName: name})
	}
	return out, nil
}

// GetCardIDByDeviceID 根据 device_id 查找其所属卡片的 card_id
func (r *PostgresCardsRepository) GetCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if tenantID == "" || deviceID == "" {
		return "", fmt.Errorf("tenant_id and device_id are required")
	}
	q := `
		SELECT card_id::text FROM cards
		WHERE tenant_id = $1
		  AND EXISTS (
		    SELECT 1 FROM jsonb_array_elements(devices) AS elem
		    WHERE elem->>'device_id' = $2 OR elem->>'device_uid' = $2 OR elem->>'uid' = $2
		  )
		LIMIT 1
	`
	var cardID string
	if err := r.db.QueryRowContext(ctx, q, tenantID, deviceID).Scan(&cardID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("card not found for device_id: %s", deviceID)
		}
		return "", fmt.Errorf("get card by device_id: %w", err)
	}
	return cardID, nil
}

// 确保实现了接口
var _ CardsRepository = (*PostgresCardsRepository)(nil)
