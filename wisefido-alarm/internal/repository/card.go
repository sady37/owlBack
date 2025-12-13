package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// CardRepository 卡片仓库（用于报警评估）
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

// CardInfo 卡片信息
type CardInfo struct {
	CardID   string
	TenantID string
	CardType string // "ActiveBed" 或 "Location"
	BedID    *string
	UnitID   string
	CardName string
	RoomID   *string // 通过 bed_id 或 unit_id 查询得到
}

// GetCardByID 根据卡片ID获取卡片信息
func (r *CardRepository) GetCardByID(tenantID, cardID string) (*CardInfo, error) {
	query := `
		SELECT 
			c.card_id,
			c.tenant_id,
			c.card_type,
			c.bed_id,
			c.unit_id,
			c.card_name,
			COALESCE(
				(SELECT r.room_id FROM rooms r WHERE r.bed_id = c.bed_id AND r.tenant_id = c.tenant_id LIMIT 1),
				(SELECT r.room_id FROM rooms r WHERE r.unit_id = c.unit_id AND r.tenant_id = c.tenant_id LIMIT 1),
				NULL
			) as room_id
		FROM cards c
		WHERE c.card_id = $1 AND c.tenant_id = $2
	`

	var card CardInfo
	var roomID sql.NullString

	err := r.db.QueryRow(query, cardID, tenantID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&card.BedID,
		&card.UnitID,
		&card.CardName,
		&roomID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}

	if roomID.Valid {
		card.RoomID = &roomID.String
	}

	return &card, nil
}

// GetCardDevices 获取卡片绑定的设备列表（从 cards.devices JSONB 字段）
func (r *CardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
	query := `
		SELECT devices
		FROM cards
		WHERE card_id = $1
	`

	var devicesJSON json.RawMessage
	err := r.db.QueryRow(query, cardID).Scan(&devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card devices: %w", err)
	}

	// 解析 JSONB
	var devices []DeviceInfo
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}

	return devices, nil
}

// GetAllCards 获取所有卡片（用于报警评估）
func (r *CardRepository) GetAllCards(tenantID string) ([]CardInfo, error) {
	query := `
		SELECT 
			c.card_id,
			c.tenant_id,
			c.card_type,
			c.bed_id,
			c.unit_id,
			c.card_name,
			COALESCE(
				(SELECT r.room_id FROM rooms r JOIN beds b ON r.room_id = b.room_id WHERE b.bed_id = c.bed_id AND r.tenant_id = c.tenant_id LIMIT 1),
				(SELECT r.room_id FROM rooms r WHERE r.unit_id = c.unit_id AND r.tenant_id = c.tenant_id LIMIT 1),
				NULL
			) as room_id
		FROM cards c
		WHERE c.tenant_id = $1
		ORDER BY c.card_id
	`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []CardInfo
	for rows.Next() {
		var card CardInfo
		var roomID sql.NullString

		err := rows.Scan(
			&card.CardID,
			&card.TenantID,
			&card.CardType,
			&card.BedID,
			&card.UnitID,
			&card.CardName,
			&roomID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		if roomID.Valid {
			card.RoomID = &roomID.String
		}

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}

	return cards, nil
}

// DeviceInfo 设备信息（从 cards.devices JSONB 解析）
type DeviceInfo struct {
	DeviceID    string  `json:"device_id"`
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`    // 设备绑定的床ID（如果绑定到床）
	BedName     *string `json:"bed_name,omitempty"`  // 床名称（如果绑定到床）
	RoomID      *string `json:"room_id,omitempty"`   // 设备绑定的房间ID（如果绑定到房间）
	RoomName    *string `json:"room_name,omitempty"` // 房间名称（如果绑定到房间，主要用于 alarm 判断是否是 bathroom）
	UnitID      string  `json:"unit_id"`             // 设备绑定的单元ID
}
