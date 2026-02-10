package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
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
	CardType string // "ActiveBedCard" 或 "UnitCard"
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

// GetCardByDeviceID 根据设备ID获取卡片信息（从 cards.devices JSONB 字段中查找）
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfo, error) {
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
		  AND c.devices @> jsonb_build_array(jsonb_build_object('device_id', $2))
		LIMIT 1
	`

	var card CardInfo
	var roomID sql.NullString

	err := r.db.QueryRow(query, tenantID, deviceID).Scan(
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
			return nil, fmt.Errorf("card not found for device: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query card by device_id: %w", err)
	}

	if roomID.Valid {
		card.RoomID = &roomID.String
	}

	return &card, nil
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

// UpdateCardAlarmCounts 更新卡片的未处理报警计数
// 通过 card_id 找到该卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
// 统计这些设备的未处理报警（alarm_status = 'active' 且 alarm_level IN ('0'-'4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')）
// 按 alarm_level 分组统计（映射到 0-4），更新 cards 表的 unhandled_alarm_0 到 unhandled_alarm_4 字段
func (r *CardRepository) UpdateCardAlarmCounts(ctx context.Context, tenantID, cardID string) error {
	// 1. 查询卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
	query := `
		SELECT devices
		FROM cards
		WHERE tenant_id = $1 AND card_id = $2
	`
	
	var devicesJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&devicesJSON)
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
		return r.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
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
		return r.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
	}
	
	// 3. 统计这些设备的未处理报警（alarm_status = 'active'）
	// 按 alarm_level 分组统计（映射到 0-4）
	countQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE alarm_level IN ('0', 'EMERG')) as count_0,
			COUNT(*) FILTER (WHERE alarm_level IN ('1', 'ALERT')) as count_1,
			COUNT(*) FILTER (WHERE alarm_level IN ('2', 'CRIT')) as count_2,
			COUNT(*) FILTER (WHERE alarm_level IN ('3', 'ERR')) as count_3,
			COUNT(*) FILTER (WHERE alarm_level IN ('4', 'WARNING')) as count_4
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status = 'active'
		  AND alarm_level IN ('0', '1', '2', '3', '4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')
		  AND (metadata->>'deleted_at' IS NULL)
	`
	
	var count0, count1, count2, count3, count4 int
	err = r.db.QueryRowContext(ctx, countQuery, tenantID, pq.Array(deviceIDs)).Scan(
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
	
	_, err = r.db.ExecContext(ctx, updateQuery, tenantID, cardID, count0, count1, count2, count3, count4)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts: %w", err)
	}
	
	return nil
}

// updateCardAlarmCountsToZero 将卡片的报警计数全部设为 0
func (r *CardRepository) updateCardAlarmCountsToZero(ctx context.Context, tenantID, cardID string) error {
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
	
	_, err := r.db.ExecContext(ctx, updateQuery, tenantID, cardID)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts to zero: %w", err)
	}
	
	return nil
}
