package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
)

// CardDeviceItem 卡片上的设备项，供 vue-radar 画布 Bind 使用。
// v2: device_code、device_uid 来自 device_factory_meta；name 来自 device_factory_meta.device_uid（待 Phase F 引入 device_label 列）。
type CardDeviceItem struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	DeviceUID  string `json:"device_uid"`
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
}

// CardsRepository 卡片Repository接口
type CardsRepository interface {
	ListCards(ctx context.Context, req ListCardsRequest) ([]*domain.CardWithUnitInfo, error)
	GetCardDevices(ctx context.Context, tenantID, cardID string) ([]CardDeviceItem, error)
	GetCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error)
}

// ListCardsRequest 查询卡片列表请求
// v2:
//   - TenantID 是 tenant /48 prefix string (INET CIDR)，不是 UUID
//   - CardType 用 v2 枚举: 'tenant'|'branch'|'site'|'unit'|'public'|'room'|'active_bed'|'device'
type ListCardsRequest struct {
	TenantID      string // INET CIDR /48
	CardID        string
	Search        string
	CardType      string // v2 枚举
	UnitType      string
	IsPublicSpace *bool
	IsSharedUnit  *bool
	Sort          string
	Direction     string

	PermissionFilter *PermissionFilter
}

// PermissionFilter 权限过滤参数
// v2: UserID = resident_id INET HoA；UserBranchTag = branch /56 prefix string；AssignedResidentIDs = INET 列表
type PermissionFilter struct {
	UserID   string // INET HoA（resident_id）
	UserType string // "resident" | "family"

	UserBranchTag *string // branch INET CIDR /56

	AssignedResidentIDs []string // INET HoA list
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
//
// v2 查询：cards 表只剩 spatial_prefix INET 一根独立空间柱。Unit 信息通过
//
//	set_masklen(c.spatial_prefix, 80) → units.unit_id 反查；branch 同理 (/56)。
//
// 权限过滤：active_bed 卡 = 直接比 c.resident_id；unit/public 卡 = 当前 (tenant level) 看
// 数据流不冗余 resident_id；business intent 让该 unit 范围内有 active resident 的 family 可见。
func (r *PostgresCardsRepository) ListCards(ctx context.Context, req ListCardsRequest) ([]*domain.CardWithUnitInfo, error) {
	if req.TenantID == "" {
		return []*domain.CardWithUnitInfo{}, nil
	}

	var query strings.Builder
	var args []any
	argIdx := 1

	query.WriteString(`
		SELECT
			c.card_id::text,
			text(c.spatial_prefix) AS spatial_prefix,
			c.card_type,
			COALESCE(c.card_name, ''),
			COALESCE(c.dns_short_name, ''),
			COALESCE(text(c.resident_id), ''),
			c.is_active,
			COALESCE(u.unit_id::text, ''),
			COALESCE(u.unit_name, ''),
			COALESCE(u.timezone, ''),
			COALESCE(u.unit_type, 0),
			COALESCE(br.branch_name, ''),
			''::text AS building_name,
			''::text AS floor,
			''::text AS layout_config
		FROM cards c
		LEFT JOIN units u ON u.unit_id = set_masklen(c.spatial_prefix, 80)
		LEFT JOIN branches br ON br.branch_id = set_masklen(c.spatial_prefix, 56)
		WHERE c.spatial_prefix <<= $` + fmt.Sprintf("%d", argIdx) + `::inet
	`)
	args = append(args, req.TenantID)
	argIdx++

	if req.CardID != "" {
		query.WriteString(` AND c.card_id::text = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, req.CardID)
		argIdx++
	}

	if req.Search != "" {
		query.WriteString(` AND (c.card_name ILIKE $` + fmt.Sprintf("%d", argIdx) + ` OR c.dns_short_name ILIKE $` + fmt.Sprintf("%d", argIdx) + `) `)
		args = append(args, "%"+req.Search+"%")
		argIdx++
	}

	if req.CardType != "" {
		query.WriteString(` AND c.card_type = $` + fmt.Sprintf("%d", argIdx) + ` `)
		args = append(args, req.CardType)
		argIdx++
	}

	// PermissionFilter — UserID (resident) self-view
	if req.PermissionFilter != nil && req.PermissionFilter.UserID != "" {
		query.WriteString(` AND (
			c.card_type = 'active_bed' AND c.resident_id::text = $` + fmt.Sprintf("%d", argIdx) + `
		) `)
		args = append(args, req.PermissionFilter.UserID)
		argIdx++
	}

	// AssignedOnly — 仅返回分配给当前用户的 resident 所属 card
	if req.PermissionFilter != nil && len(req.PermissionFilter.AssignedResidentIDs) > 0 {
		query.WriteString(` AND c.resident_id = ANY($` + fmt.Sprintf("%d", argIdx) + `::inet[]) `)
		args = append(args, pq.Array(req.PermissionFilter.AssignedResidentIDs))
		argIdx++
	}

	// BranchOnly — UserBranchTag 是 branch /56 prefix
	if req.PermissionFilter != nil && req.PermissionFilter.UserBranchTag != nil {
		bt := *req.PermissionFilter.UserBranchTag
		if bt != "" {
			query.WriteString(` AND c.spatial_prefix <<= $` + fmt.Sprintf("%d", argIdx) + `::inet `)
			args = append(args, bt)
			argIdx++
		}
	}

	// 排序
	sortField := "card_name"
	switch req.Sort {
	case "card_name", "dns_short_name", "created_at":
		sortField = req.Sort
	}
	direction := "ASC"
	if strings.EqualFold(req.Direction, "desc") {
		direction = "DESC"
	}
	query.WriteString(` ORDER BY c.` + sortField + ` ` + direction)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var results []*domain.CardWithUnitInfo
	for rows.Next() {
		card := &domain.Card{}
		unit := &domain.Unit{}

		var residentIDStr, unitIDStr, unitName, timezone, branchName, buildingName, layoutConfig string
		var floorStr string
		var unitType int8

		if err := rows.Scan(
			&card.CardID,
			&card.SpatialPrefix,
			&card.CardType,
			&card.CardName.String,
			&card.DNSShortName.String,
			&residentIDStr,
			&card.IsActive,
			&unitIDStr,
			&unitName,
			&timezone,
			&unitType,
			&branchName,
			&buildingName,
			&floorStr,
			&layoutConfig,
		); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		if card.CardName.String != "" {
			card.CardName.Valid = true
		}
		if card.DNSShortName.String != "" {
			card.DNSShortName.Valid = true
		}
		if residentIDStr != "" {
			card.ResidentID = sql.NullString{String: residentIDStr, Valid: true}
		}

		// Unit
		unit.UnitID = unitIDStr
		unit.UnitName = unitName
		unit.Timezone = timezone
		unit.UnitType = unitType
		if branchName != "" {
			unit.BranchName = sql.NullString{String: branchName, Valid: true}
		}
		if buildingName != "" {
			unit.BuildingName = sql.NullString{String: buildingName, Valid: true}
		}
		if floorStr != "" {
			unit.Floor = sql.NullString{String: floorStr, Valid: true}
		}
		if layoutConfig != "" {
			unit.LayoutConfig = sql.NullString{String: layoutConfig, Valid: true}
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

// GetCardDevices 获取卡片上所有设备（v2: cards.spatial_prefix LPM device.device_ipv6）。
func (r *PostgresCardsRepository) GetCardDevices(ctx context.Context, tenantID, cardID string) ([]CardDeviceItem, error) {
	if cardID == "" {
		return nil, nil
	}
	q := `
		SELECT
			dfm.device_id::text,
			COALESCE(dfm.device_type::text, ''),
			dfm.device_uid,
			COALESCE(dfm.device_code, ''),
			dfm.device_uid AS device_name
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE c.card_id = $1::uuid
		ORDER BY d.device_ipv6
	`
	rows, err := r.db.QueryContext(ctx, q, cardID)
	if err != nil {
		return nil, fmt.Errorf("get card devices: %w", err)
	}
	defer rows.Close()

	var out []CardDeviceItem
	for rows.Next() {
		var it CardDeviceItem
		if err := rows.Scan(&it.DeviceID, &it.DeviceType, &it.DeviceUID, &it.DeviceCode, &it.DeviceName); err != nil {
			return nil, fmt.Errorf("scan card device: %w", err)
		}
		if it.DeviceCode == "" {
			it.DeviceCode = it.DeviceUID
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// GetCardIDByDeviceID 根据 device_id 查找其所属卡片的 card_id（v2: LPM）
func (r *PostgresCardsRepository) GetCardIDByDeviceID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if deviceID == "" {
		return "", fmt.Errorf("device_id is required")
	}
	var cardID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT find_card_by_device_addr(d.device_ipv6)::text
		FROM devices d
		WHERE d.device_id = $1::uuid
		LIMIT 1
	`, deviceID).Scan(&cardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device not found: %s", deviceID)
		}
		return "", fmt.Errorf("get card by device_id: %w", err)
	}
	if !cardID.Valid || cardID.String == "" {
		return "", fmt.Errorf("no card for device_id: %s", deviceID)
	}
	return cardID.String, nil
}

// 确保实现了接口
var _ CardsRepository = (*PostgresCardsRepository)(nil)
