package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"owl-common/card"
)

// CardInfo 卡片信息（用于数据聚合）
type CardInfo struct {
	CardID        string
	TenantID      string
	CardType      string // "ActiveBed" 或 "Location"
	BedID         *string
	UnitID        string
	CardName      string
	CardAddress   string
	ResidentID    *string // ActiveBed 卡片的主住户
	UnhandledAlarm0 *int
	UnhandledAlarm1 *int
	UnhandledAlarm2 *int
	UnhandledAlarm3 *int
	UnhandledAlarm4 *int
	IconAlarmLevel  *int
	PopAlarmEmerge   *int
}

// GetCardByID 根据卡片ID获取卡片信息（用于数据聚合）
func (r *CardRepository) GetCardByID(tenantID, cardID string) (*CardInfo, error) {
	query := `
		SELECT 
			card_id,
			tenant_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			resident_id,
			unhandled_alarm_0,
			unhandled_alarm_1,
			unhandled_alarm_2,
			unhandled_alarm_3,
			unhandled_alarm_4,
			icon_alarm_level,
			pop_alarm_emerge
		FROM cards
		WHERE card_id = $1 AND tenant_id = $2
	`

	var card CardInfo
	var bedID, residentID sql.NullString
	var unhandled0, unhandled1, unhandled2, unhandled3, unhandled4 sql.NullInt64
	var iconLevel, popEmerge sql.NullInt64

	err := r.db.QueryRow(query, cardID, tenantID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&bedID,
		&card.UnitID,
		&card.CardName,
		&card.CardAddress,
		&residentID,
		&unhandled0,
		&unhandled1,
		&unhandled2,
		&unhandled3,
		&unhandled4,
		&iconLevel,
		&popEmerge,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}

	if bedID.Valid {
		card.BedID = &bedID.String
	}
	if residentID.Valid {
		card.ResidentID = &residentID.String
	}
	if unhandled0.Valid {
		val := int(unhandled0.Int64)
		card.UnhandledAlarm0 = &val
	}
	if unhandled1.Valid {
		val := int(unhandled1.Int64)
		card.UnhandledAlarm1 = &val
	}
	if unhandled2.Valid {
		val := int(unhandled2.Int64)
		card.UnhandledAlarm2 = &val
	}
	if unhandled3.Valid {
		val := int(unhandled3.Int64)
		card.UnhandledAlarm3 = &val
	}
	if unhandled4.Valid {
		val := int(unhandled4.Int64)
		card.UnhandledAlarm4 = &val
	}
	if iconLevel.Valid {
		val := int(iconLevel.Int64)
		card.IconAlarmLevel = &val
	}
	if popEmerge.Valid {
		val := int(popEmerge.Int64)
		card.PopAlarmEmerge = &val
	}

	return &card, nil
}

// GetCardDevices 获取卡片绑定的设备列表（从 cards.devices JSONB 字段）
func (r *CardRepository) GetCardDevices(cardID string) ([]card.DeviceInfo, error) {
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

	// 先解析为 DeviceJSON（匹配 JSONB 字段名）
	var deviceJSONs []card.DeviceJSON
	if err := json.Unmarshal(devicesJSON, &deviceJSONs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}

	// 转换为 DeviceInfo
	devices := make([]card.DeviceInfo, 0, len(deviceJSONs))
	for _, dj := range deviceJSONs {
		devices = append(devices, card.DeviceInfo{
			DeviceID:          dj.DeviceID,
			DeviceName:        dj.DeviceName,
			DeviceType:        dj.DeviceType,
			DeviceModel:        dj.DeviceModel,
			BoundBedID:         dj.BedID,
			BedName:            dj.BedName,
			BoundRoomID:        dj.RoomID,
			RoomName:           dj.RoomName,
			UnitID:             dj.UnitID,
			MonitoringEnabled:  true, // 卡片中的设备默认启用监控
		})
	}

	return devices, nil
}

// GetCardResidents 获取卡片绑定的住户列表（从 cards.residents JSONB 字段）
func (r *CardRepository) GetCardResidents(cardID string) ([]card.ResidentInfo, error) {
	query := `
		SELECT residents
		FROM cards
		WHERE card_id = $1
	`

	var residentsJSON json.RawMessage
	err := r.db.QueryRow(query, cardID).Scan(&residentsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card residents: %w", err)
	}

	// 解析 JSONB
	var residents []card.ResidentInfo
	if err := json.Unmarshal(residentsJSON, &residents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal residents JSON: %w", err)
	}

	return residents, nil
}

// GetAllCards 获取所有卡片（用于数据聚合）
// 如果 tenantID 为空，则获取所有租户的卡片；否则只获取指定租户的卡片
func (r *CardRepository) GetAllCards(tenantID string) ([]CardInfo, error) {
	var query string
	var args []interface{}
	
	if tenantID == "" {
		// 查询所有租户的卡片
		query = `
			SELECT 
				card_id,
				tenant_id,
				card_type,
				bed_id,
				unit_id,
				card_name,
				card_address,
				resident_id,
				unhandled_alarm_0,
				unhandled_alarm_1,
				unhandled_alarm_2,
				unhandled_alarm_3,
				unhandled_alarm_4,
				icon_alarm_level,
				pop_alarm_emerge
			FROM cards
			ORDER BY tenant_id, card_id
		`
		args = []interface{}{}
	} else {
		// 查询指定租户的卡片
		query = `
			SELECT 
				card_id,
				tenant_id,
				card_type,
				bed_id,
				unit_id,
				card_name,
				card_address,
				resident_id,
				unhandled_alarm_0,
				unhandled_alarm_1,
				unhandled_alarm_2,
				unhandled_alarm_3,
				unhandled_alarm_4,
				icon_alarm_level,
				pop_alarm_emerge
			FROM cards
			WHERE tenant_id = $1
			ORDER BY card_id
		`
		args = []interface{}{tenantID}
	}

	var rows *sql.Rows
	var err error
	if len(args) == 0 {
		rows, err = r.db.Query(query)
	} else {
		rows, err = r.db.Query(query, args...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []CardInfo
	for rows.Next() {
		var card CardInfo
		var bedID, residentID sql.NullString
		var unhandled0, unhandled1, unhandled2, unhandled3, unhandled4 sql.NullInt64
		var iconLevel, popEmerge sql.NullInt64

		err := rows.Scan(
			&card.CardID,
			&card.TenantID,
			&card.CardType,
			&bedID,
			&card.UnitID,
			&card.CardName,
			&card.CardAddress,
			&residentID,
			&unhandled0,
			&unhandled1,
			&unhandled2,
			&unhandled3,
			&unhandled4,
			&iconLevel,
			&popEmerge,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		if bedID.Valid {
			card.BedID = &bedID.String
		}
		if residentID.Valid {
			card.ResidentID = &residentID.String
		}
		if unhandled0.Valid {
			val := int(unhandled0.Int64)
			card.UnhandledAlarm0 = &val
		}
		if unhandled1.Valid {
			val := int(unhandled1.Int64)
			card.UnhandledAlarm1 = &val
		}
		if unhandled2.Valid {
			val := int(unhandled2.Int64)
			card.UnhandledAlarm2 = &val
		}
		if unhandled3.Valid {
			val := int(unhandled3.Int64)
			card.UnhandledAlarm3 = &val
		}
		if unhandled4.Valid {
			val := int(unhandled4.Int64)
			card.UnhandledAlarm4 = &val
		}
		if iconLevel.Valid {
			val := int(iconLevel.Int64)
			card.IconAlarmLevel = &val
		}
		if popEmerge.Valid {
			val := int(popEmerge.Int64)
			card.PopAlarmEmerge = &val
		}

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}

	return cards, nil
}

// GetAllTenants 获取所有租户 ID（用于多租户聚合）
func (r *CardRepository) GetAllTenants(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT tenant_id::text
		FROM cards
		ORDER BY tenant_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenantIDs []string
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			return nil, fmt.Errorf("failed to scan tenant_id: %w", err)
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenantIDs, nil
}

