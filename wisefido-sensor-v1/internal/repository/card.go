package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"

	"go.uber.org/zap"
)

// CardRepository 卡片仓库（用于报警评估）
//
// v2 schema: cards 表只有 spatial_prefix INET PK + card_type + card_name；
// tenant_id/unit_id/bed_id/room_id/devices JSONB 全部退役，由 spatial_prefix 派生 +
// JOIN devices LPM (devices.device_ipv6 <<= cards.spatial_prefix) 反查设备列表。
//
// CardType 兼容映射：v2 lowercase ('active_bed'/'unit') ↔ v1 PascalCase ('ActiveBedCard'/'UnitCard')
// 内部存 v1 风格供 evaluator/event_consumer 既有 string 比较直接复用。
type CardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCardRepository 创建卡片仓库
func NewCardRepository(db *sql.DB, logger *zap.Logger) *CardRepository {
	return &CardRepository{
		db:     db,
		logger: logger,
	}
}

// CardInfo 卡片信息（v2 派生形态）。
//
// CardID = cards.spatial_prefix INET CIDR text，e.g. "fd00:0:3:111:3:101::/96"
// 其余 BedID/UnitID/RoomID/TenantID 也都是 INET CIDR text，按 mask 长度派生。
type CardInfo struct {
	CardID   string  // spatial_prefix INET CIDR
	TenantID string  // /48 prefix CIDR
	CardType string  // "ActiveBedCard" | "UnitCard" (v1 style，repo 内部映射)
	BedID    *string // /96 CIDR (active_bed 才有)
	UnitID   string  // /80 CIDR
	CardName string
	RoomID   *string // /88 CIDR
}

// v2CardTypeToV1 v2 lowercase → v1 PascalCase（caller 既有 == "ActiveBedCard" 比较保留）
func v2CardTypeToV1(v2 string) string {
	switch v2 {
	case "bed":
		return "ActiveBedCard"
	case "unit":
		return "UnitCard"
	case "room":
		return "RoomCard"
	case "public":
		return "PublicCard"
	default:
		return v2
	}
}

// cardSelectExpr v2 cards 行 → CardInfo 字段的派生表达式（INET CIDR 文本）。
// caller 直接 SELECT cardSelectExpr 后 Scan 8 个字段。
const cardSelectExpr = `
		c.spatial_prefix::text                                                                                AS card_id,
		host(network(set_masklen(c.spatial_prefix, 48)))::text || '/48'                                       AS tenant_id,
		c.card_type                                                                                            AS card_type,
		CASE WHEN masklen(c.spatial_prefix) >= 96
		     THEN host(network(set_masklen(c.spatial_prefix, 96)))::text || '/96'
		END                                                                                                    AS bed_id,
		CASE WHEN masklen(c.spatial_prefix) >= 80
		     THEN host(network(set_masklen(c.spatial_prefix, 80)))::text || '/80'
		END                                                                                                    AS unit_id,
		c.card_name                                                                                            AS card_name,
		CASE WHEN masklen(c.spatial_prefix) >= 88
		     THEN host(network(set_masklen(c.spatial_prefix, 88)))::text || '/88'
		END                                                                                                    AS room_id`

// scanCard 把 8 列扫到 CardInfo（v1 兼容字段名 + CardType v2→v1 映射）。
func scanCard(scanner interface {
	Scan(dest ...interface{}) error
}) (*CardInfo, error) {
	var c CardInfo
	var bedID, unitID, roomID, cardName sql.NullString
	if err := scanner.Scan(&c.CardID, &c.TenantID, &c.CardType, &bedID, &unitID, &cardName, &roomID); err != nil {
		return nil, err
	}
	if bedID.Valid {
		c.BedID = &bedID.String
	}
	if unitID.Valid {
		c.UnitID = unitID.String
	}
	if cardName.Valid {
		c.CardName = cardName.String
	}
	if roomID.Valid {
		c.RoomID = &roomID.String
	}
	c.CardType = v2CardTypeToV1(c.CardType)
	return &c, nil
}

// GetCardByID 根据卡片ID（v2 = spatial_prefix INET CIDR）获取卡片信息。
func (r *CardRepository) GetCardByID(tenantID, cardID string) (*CardInfo, error) {
	query := `SELECT ` + cardSelectExpr + `
		FROM cards c
		WHERE c.spatial_prefix = $1::INET
		  AND network(set_masklen(c.spatial_prefix, 48)) = $2::INET
		LIMIT 1`
	row := r.db.QueryRow(query, cardID, tenantID)
	c, err := scanCard(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}
	return c, nil
}

// GetCardDevices 获取卡片下的设备列表（v2: cards.devices JSONB 退役，改 JOIN devices LPM 反查）。
//
// cardID = cards.spatial_prefix INET CIDR；devices.device_ipv6 <<= spatial_prefix 命中即归属该卡。
// device_factory_meta (dfm) 提供 device_uid/device_type/device_model；rooms LEFT JOIN 取 room_name 给
// bathroom 判断。
func (r *CardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
	query := `
		SELECT COALESCE(json_agg(json_build_object(
		    'device_id',    dfm.device_id::text,
		    'device_uid',   dfm.device_uid,
		    'device_type',  COALESCE(dfm.device_type::text, ''),
		    'device_model', COALESCE(dfm.device_model, ''),
		    'bed_id',       host(network(set_masklen(d.device_ipv6, 96)))::text || '/96',
		    'room_id',      host(network(set_masklen(d.device_ipv6, 88)))::text || '/88',
		    'room_name',    r.room_name,
		    'unit_id',      host(network(set_masklen(d.device_ipv6, 80)))::text || '/80'
		)), '[]'::json)
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		LEFT JOIN rooms r            ON r.room_id = network(set_masklen(d.device_ipv6, 88))
		WHERE d.device_ipv6 <<= $1::INET`
	var devicesJSON json.RawMessage
	if err := r.db.QueryRow(query, cardID).Scan(&devicesJSON); err != nil {
		return nil, fmt.Errorf("failed to query card devices: %w", err)
	}
	var devices []DeviceInfo
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}
	return devices, nil
}

// GetAllCards 获取所有卡片（用于报警评估）。tenantID 是 INET CIDR text e.g. "fd00:0:3::/48"。
func (r *CardRepository) GetAllCards(tenantID string) ([]CardInfo, error) {
	query := `SELECT ` + cardSelectExpr + `
		FROM cards c
		WHERE network(set_masklen(c.spatial_prefix, 48)) = $1::INET
		ORDER BY c.spatial_prefix`
	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var out []CardInfo
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}
	return out, nil
}

// GetCardByDeviceID 根据 device 标识查所属卡（v2 = device_addr canonical IPv6 string）。
//
// device_ipv6 单程票后 deviceID 入参是 canonical IPv6 (e.g. "fd00:0:3:111:3:101:a2ac:d523")；
// LPM cards.spatial_prefix >>= device_ipv6 命中。
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfo, error) {
	addr, perr := netip.ParseAddr(deviceID)
	if perr != nil {
		return nil, fmt.Errorf("device id is not a valid IPv6 addr (%q): %w", deviceID, perr)
	}
	query := `SELECT ` + cardSelectExpr + `
		FROM cards c
		WHERE c.spatial_prefix >>= $1::INET
		  AND network(set_masklen(c.spatial_prefix, 48)) = $2::INET
		ORDER BY masklen(c.spatial_prefix) DESC
		LIMIT 1`
	row := r.db.QueryRow(query, addr.String(), tenantID)
	c, err := scanCard(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found for device: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query card by device_id: %w", err)
	}
	return c, nil
}

// DeviceInfo 设备信息（cards 下的 device 列表）。
//
// device_ipv6 单程票后所有 *ID 字段为 INET CIDR text 派生（device_addr / bed_id /96 / room_id /88 / unit_id /80）。
type DeviceInfo struct {
	DeviceID    string  `json:"device_id"`            // dfm.device_id UUID（FK 稳定）
	DeviceName  string  `json:"device_name"`          // 当前 v2 schema 暂无 device_name 列；保留字段
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	DeviceUID   string  `json:"device_uid,omitempty"` // dfm.device_uid (MAC)
	BedID       *string `json:"bed_id,omitempty"`     // /96 CIDR
	BedName     *string `json:"bed_name,omitempty"`
	RoomID      *string `json:"room_id,omitempty"`    // /88 CIDR
	RoomName    *string `json:"room_name,omitempty"`  // 用于 bathroom 判定
	UnitID      string  `json:"unit_id"`              // /80 CIDR
}
