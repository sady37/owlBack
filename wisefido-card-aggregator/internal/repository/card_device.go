package repository

import (
	"database/sql"
	"fmt"
)

// CardInfoForFusion 用于数据融合的卡片信息
type CardInfoForFusion struct {
	CardID   string
	TenantID string
	CardType string // "ActiveBed" 或 "Location"
	BedID    *string
	UnitID   string
}

// GetCardByDeviceID 根据设备ID获取关联的卡片（用于数据融合）
// 
// 查询逻辑：
// 1. 如果设备绑定到 bed（bound_bed_id IS NOT NULL）：
//    - 查询 ActiveBed 类型的卡片（cards.bed_id = bound_bed_id）
// 2. 如果设备绑定到 room（bound_room_id IS NOT NULL）且未绑床：
//    - 通过 room.unit_id 查询 Location 类型的卡片（cards.unit_id = room.unit_id）
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfoForFusion, error) {
	query := `
		WITH device_info AS (
			SELECT 
				d.device_id,
				d.tenant_id,
				d.bound_bed_id,
				d.bound_room_id
			FROM devices d
			WHERE d.device_id = $1 AND d.tenant_id = $2
		),
		bed_card AS (
			-- 场景 1：设备绑定到床，查询 ActiveBed 卡片
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.bed_id = di.bound_bed_id AND c.tenant_id = di.tenant_id
			WHERE di.bound_bed_id IS NOT NULL
			  AND c.card_type = 'ActiveBed'
			LIMIT 1
		),
		room_card AS (
			-- 场景 2：设备绑定到房间，通过 room.unit_id 查询 Location 卡片
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.unit_id = (
				SELECT r.unit_id FROM rooms r WHERE r.room_id = di.bound_room_id AND r.tenant_id = di.tenant_id
			) AND c.tenant_id = di.tenant_id
			WHERE di.bound_room_id IS NOT NULL
			  AND di.bound_bed_id IS NULL
			  AND c.card_type = 'Location'
			LIMIT 1
		)
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM bed_card
		UNION ALL
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM room_card
		LIMIT 1
	`
	
	card := &CardInfoForFusion{}
	var bedID, unitID sql.NullString
	
	err := r.db.QueryRow(query, deviceID, tenantID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&bedID,
		&unitID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found for device: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}
	
	if bedID.Valid {
		card.BedID = &bedID.String
	}
	if unitID.Valid {
		card.UnitID = unitID.String
	}
	
	return card, nil
}

