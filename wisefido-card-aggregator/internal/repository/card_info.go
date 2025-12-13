package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// GetCardResidents 获取卡片绑定的住户列表（从 cards.residents JSONB 字段）
func (r *CardRepository) GetCardResidents(cardID string) ([]ResidentInfo, error) {
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
	var residents []ResidentInfo
	if err := json.Unmarshal(residentsJSON, &residents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal residents JSON: %w", err)
	}

	return residents, nil
}

// GetAllCards 获取所有卡片（用于数据聚合）
func (r *CardRepository) GetAllCards(tenantID string) ([]CardInfo, error) {
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
		WHERE tenant_id = $1
		ORDER BY card_id
	`

	rows, err := r.db.Query(query, tenantID)
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

