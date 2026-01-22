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
// 查询逻辑（简化）：
// 1. 获取设备的 unit_id：
//    - 如果设备绑定到 bed，通过 bed.room_id -> room.unit_id 获取
//    - 如果设备绑定到 room，通过 room.unit_id 获取
// 2. 查找 cards 表中 unit_id 匹配的卡片（card.unit_id = device.unit_id）
// 3. 检查该 card 的 devices JSONB 字段中是否包含该 device_id
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfoForFusion, error) {
	query := `
		WITH device_unit AS (
			-- 获取设备对应的 unit_id
			SELECT 
				d.device_id,
				d.tenant_id,
				COALESCE(
					-- 如果绑定到 bed，通过 bed -> room -> unit 获取 unit_id
					(SELECT r.unit_id FROM beds b 
					 INNER JOIN rooms r ON r.room_id = b.room_id 
					 WHERE b.bed_id = d.bound_bed_id AND b.tenant_id = d.tenant_id),
					-- 如果绑定到 room，直接获取 room.unit_id
					(SELECT r.unit_id FROM rooms r 
					 WHERE r.room_id = d.bound_room_id AND r.tenant_id = d.tenant_id)
				) AS unit_id
			FROM devices d
			WHERE d.device_id = $1 AND d.tenant_id = $2
		)
		SELECT 
			c.card_id,
			c.tenant_id,
			c.card_type,
			c.bed_id,
			c.unit_id
		FROM cards c
		INNER JOIN device_unit du ON c.unit_id = du.unit_id AND c.tenant_id = du.tenant_id
		WHERE EXISTS (
			-- 检查 card 的 devices JSONB 字段中是否包含该 device_id
			SELECT 1
			FROM jsonb_array_elements(c.devices) AS device
			WHERE (device->>'device_id')::uuid = $1::uuid
		)
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

