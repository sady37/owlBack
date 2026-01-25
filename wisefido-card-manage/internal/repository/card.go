package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"owl-common/card"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// CardRepository implements card.RepositoryInterface
type CardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCardRepository creates a new card repository
func NewCardRepository(db *sql.DB, logger *zap.Logger) *CardRepository {
	return &CardRepository{
		db:     db,
		logger: logger,
	}
}

// GetActiveBedsByUnit gets all ActiveBeds under the specified unit
// ActiveBed condition: 床上有 monitoring_enabled = TRUE 的设备即可
// 注意：bed_type 字段已删除，改为动态查询设备绑定状态
func (r *CardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedInfo, error) {
	query := `
		SELECT DISTINCT
			b.bed_id,
			r.unit_id,
			COUNT(DISTINCT d.device_id)::int AS bound_device_count,
			r2.resident_id,
			b.room_id,
			b.bed_name
		FROM beds b
		INNER JOIN rooms r ON b.room_id = r.room_id
		INNER JOIN devices d ON d.bound_bed_id = b.bed_id
		LEFT JOIN residents r2 ON r2.bed_id = b.bed_id AND r2.tenant_id = $1
		WHERE b.tenant_id = $1
		  AND r.unit_id = $2
		  AND d.monitoring_enabled = TRUE
		  AND d.status <> 'disabled'
		GROUP BY b.bed_id, r.unit_id, r2.resident_id, b.room_id, b.bed_name
		HAVING COUNT(DISTINCT d.device_id) > 0
		ORDER BY b.bed_name
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active beds: %w", err)
	}
	defer rows.Close()

	var beds []card.ActiveBedInfo
	for rows.Next() {
		var bed card.ActiveBedInfo
		var residentID sql.NullString
		var bedName sql.NullString // bed_name is selected but not used in ActiveBedInfo

		if err := rows.Scan(
			&bed.BedID,
			&bed.UnitID,
			&bed.BoundDeviceCount,
			&residentID,
			&bed.RoomID,
			&bedName, // Scan bed_name even though it's not used
		); err != nil {
			return nil, fmt.Errorf("failed to scan bed: %w", err)
		}

		if residentID.Valid {
			bed.ResidentID = &residentID.String
		}

		beds = append(beds, bed)
	}

	return beds, nil
}

// GetUnitInfo gets Unit information
func (r *CardRepository) GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error) {
	query := `
		SELECT 
			u.unit_id,
			u.unit_name,
			COALESCE(b.branch_name, '') as branch_name,
			COALESCE(bld.building_name, '') as building,
			u.is_public,
			u.is_shared_unit,
			u.unit_type
		FROM units u
		LEFT JOIN branches b ON u.branch_id = b.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`

	var unitInfo card.UnitInfo

	err := r.db.QueryRow(query, tenantID, unitID).Scan(
		&unitInfo.UnitID,
		&unitInfo.UnitName,
		&unitInfo.BranchName,
		&unitInfo.Building,
		&unitInfo.IsPublic,
		&unitInfo.IsSharedUnit,
		&unitInfo.UnitType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found: %s", unitID)
		}
		return nil, fmt.Errorf("failed to query unit: %w", err)
	}

	return &unitInfo, nil
}

// GetDevicesByBed gets all devices with monitoring_enabled = TRUE bound to the specified bed
func (r *CardRepository) GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			COALESCE(d.device_uid, ''),
			COALESCE(ds.device_code, ''),
			d.device_name,
			ds.device_type,
			ds.device_model,
			d.bound_bed_id,
			b.bed_name,
			d.bound_room_id,
			r.room_name,
			u.unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_id = ds.device_id
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON b.room_id = r.room_id AND b.tenant_id = r.tenant_id
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE d.tenant_id = $1
		  AND d.bound_bed_id = $2
		  AND d.monitoring_enabled = TRUE
		ORDER BY d.device_name
	`

	rows, err := r.db.Query(query, tenantID, bedID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []card.DeviceInfo
	for rows.Next() {
		var device card.DeviceInfo
		var boundBedID, bedName, boundRoomID, roomName sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.DeviceCode,
			&device.DeviceName,
			&device.DeviceType,
			&device.DeviceModel,
			&boundBedID,
			&bedName,
			&boundRoomID,
			&roomName,
			&device.UnitID,
			&device.MonitoringEnabled,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if boundBedID.Valid {
			device.BoundBedID = &boundBedID.String
		}
		if bedName.Valid {
			device.BedName = &bedName.String
		}
		if boundRoomID.Valid {
			device.BoundRoomID = &boundRoomID.String
		}
		if roomName.Valid {
			device.RoomName = &roomName.String
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetUnboundDevicesByUnit gets all devices with monitoring_enabled = TRUE that are bound to a room in the specified unit but not bound to any bed
func (r *CardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			COALESCE(d.device_uid, ''),
			COALESCE(ds.device_code, ''),
			d.device_name,
			ds.device_type,
			ds.device_model,
			d.bound_bed_id,
			b.bed_name,
			d.bound_room_id,
			r.room_name,
			u.unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_id = ds.device_id
		LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		WHERE d.tenant_id = $1
		  AND u.unit_id = $2
		  AND d.bound_room_id IS NOT NULL
		  AND d.bound_bed_id IS NULL
		  AND d.monitoring_enabled = TRUE
		ORDER BY d.device_name
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unbound devices: %w", err)
	}
	defer rows.Close()

	var devices []card.DeviceInfo
	for rows.Next() {
		var device card.DeviceInfo
		var boundBedID, bedName, boundRoomID, roomName sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.DeviceCode,
			&device.DeviceName,
			&device.DeviceType,
			&device.DeviceModel,
			&boundBedID,
			&bedName,
			&boundRoomID,
			&roomName,
			&device.UnitID,
			&device.MonitoringEnabled,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if boundBedID.Valid {
			device.BoundBedID = &boundBedID.String
		}
		if bedName.Valid {
			device.BedName = &bedName.String
		}
		if boundRoomID.Valid {
			device.BoundRoomID = &boundRoomID.String
		}
		if roomName.Valid {
			device.RoomName = &roomName.String
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetResidentByBed gets the resident bound to the specified bed
func (r *CardRepository) GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error) {
	query := `
		SELECT 
			r.resident_id,
			r.nickname,
			r.unit_id,
			r.bed_id
		FROM residents r
		WHERE r.tenant_id = $1
		  AND r.bed_id = $2
		LIMIT 1
	`

	var resident card.ResidentInfo
	var unitID, residentBedID sql.NullString

	err := r.db.QueryRow(query, tenantID, bedID).Scan(
		&resident.ResidentID,
		&resident.Nickname,
		&unitID,
		&residentBedID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Bed is not bound to any resident
		}
		return nil, fmt.Errorf("failed to query resident: %w", err)
	}

	if unitID.Valid {
		resident.UnitID = &unitID.String
	}
	if residentBedID.Valid {
		resident.BedID = &residentBedID.String
	}

	return &resident, nil
}

// GetResidentsByUnit gets all residents under the specified unit
func (r *CardRepository) GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error) {
	query := `
		SELECT 
			r.resident_id,
			r.nickname,
			r.unit_id,
			r.bed_id
		FROM residents r
		WHERE r.tenant_id = $1
		  AND r.unit_id = $2
		ORDER BY r.nickname
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query residents: %w", err)
	}
	defer rows.Close()

	var residents []card.ResidentInfo
	for rows.Next() {
		var resident card.ResidentInfo
		var unitID, bedID sql.NullString

		if err := rows.Scan(
			&resident.ResidentID,
			&resident.Nickname,
			&unitID,
			&bedID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}

		if unitID.Valid {
			resident.UnitID = &unitID.String
		}
		if bedID.Valid {
			resident.BedID = &bedID.String
		}

		residents = append(residents, resident)
	}

	return residents, nil
}

// GetAllUnits gets all units (for full card creation)
func (r *CardRepository) GetAllUnits(tenantID string) ([]string, error) {
	query := `
		SELECT unit_id
		FROM units
		WHERE tenant_id = $1
		ORDER BY unit_name
	`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query units: %w", err)
	}
	defer rows.Close()

	var unitIDs []string
	for rows.Next() {
		var unitID string
		if err := rows.Scan(&unitID); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		unitIDs = append(unitIDs, unitID)
	}

	return unitIDs, nil
}

// GetUnitIDByBedID gets unit_id by bed_id
func (r *CardRepository) GetUnitIDByBedID(tenantID, bedID string) (string, error) {
	query := `
		SELECT unit_id
		FROM beds
		WHERE tenant_id = $1 AND bed_id = $2
		LIMIT 1
	`

	var unitID string
	err := r.db.QueryRow(query, tenantID, bedID).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("bed not found: %s", bedID)
		}
		return "", fmt.Errorf("failed to query unit_id: %w", err)
	}

	return unitID, nil
}

// GetCardsByUnit gets all cards under the specified unit (including devices and residents JSONB)
func (r *CardRepository) GetCardsByUnit(tenantID, unitID string) ([]card.CardWithContent, error) {
	query := `
		SELECT 
			card_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			resident_id,
			devices,
			residents
		FROM cards
		WHERE tenant_id = $1
		  AND unit_id = $2
		ORDER BY card_type, bed_id
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []card.CardWithContent
	for rows.Next() {
		var cardItem card.CardWithContent
		var bedID, residentID sql.NullString
		var devicesJSON, residentsJSON json.RawMessage

		err := rows.Scan(
			&cardItem.CardID,
			&cardItem.CardType,
			&bedID,
			&cardItem.UnitID,
			&cardItem.CardName,
			&cardItem.CardAddress,
			&residentID,
			&devicesJSON,
			&residentsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		if bedID.Valid {
			cardItem.BedID = &bedID.String
		}
		if residentID.Valid {
			cardItem.ResidentID = &residentID.String
		}

		// Store raw JSON for comparison
		cardItem.DevicesJSON = devicesJSON
		cardItem.ResidentsJSON = residentsJSON

		cards = append(cards, cardItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}

	return cards, nil
}

// DeleteCard deletes a single card by card_id
func (r *CardRepository) DeleteCard(tenantID, cardID string) error {
	query := `
		DELETE FROM cards
		WHERE tenant_id = $1
		  AND card_id = $2
	`

	result, err := r.db.Exec(query, tenantID, cardID)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found: %s", cardID)
	}

	return nil
}

// DeleteCardsByUnit deletes all cards under the specified unit (for recreation)
func (r *CardRepository) DeleteCardsByUnit(tenantID, unitID string) error {
	query := `
		DELETE FROM cards
		WHERE tenant_id = $1
		  AND unit_id = $2
	`

	_, err := r.db.Exec(query, tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to delete cards: %w", err)
	}

	return nil
}

// CountCardsByTenant counts total number of cards for a tenant
func (r *CardRepository) CountCardsByTenant(tenantID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM cards
		WHERE tenant_id = $1
	`

	var count int
	err := r.db.QueryRow(query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cards: %w", err)
	}

	return count, nil
}

// UpdateCard updates an existing card by card_id
// Updates all fields except card_id, tenant_id, and alarm statistics
// This preserves card_id to avoid frontend cache invalidation
func (r *CardRepository) UpdateCard(
	tenantID, cardID string,
	cardType string,
	bedID *string, unitID, cardName, cardAddress string,
	residentID *string,
	devicesJSON, residentsJSON []byte,
) error {
	query := `
		UPDATE cards
		SET
			card_type = $3,
			bed_id = $4,
			unit_id = $5,
			card_name = $6,
			card_address = $7,
			resident_id = $8,
			devices = $9,
			residents = $10
		WHERE tenant_id = $1
		  AND card_id = $2
	`

	result, err := r.db.Exec(
		query,
		tenantID,
		cardID,
		cardType,
		bedID,
		unitID,
		cardName,
		cardAddress,
		residentID,
		devicesJSON,
		residentsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found: %s", cardID)
	}

	return nil
}

// CreateCard creates a card
//
// Fields to insert:
// - Required fields: tenant_id, card_type, bed_id/unit_id, card_name, card_address, devices, residents
// - Optional fields: resident_id (primary resident for ActiveBed cards)
//
// Fields using default values (not inserted):
// - unhandled_alarm_0 ~ unhandled_alarm_4 (unhandled alarm statistics, default 0)
// - icon_alarm_level (icon alarm level threshold, default 3)
// - pop_alarm_emerge (popup alarm level threshold, default 0)
//
// Constraint checks:
// - ActiveBed: bed_id IS NOT NULL, unit_id can be NULL (redundant)
// - Location: unit_id IS NOT NULL, bed_id must be NULL
//
// Note: Alarm routing configuration (routing_alarm_user_ids, routing_alarm_tags) has been removed.
// Cards only handle alarm level display, not alarm routing.
func (r *CardRepository) CreateCard(
	tenantID string,
	cardType string, // "ActiveBed" or "Location"
	bedID *string,
	unitID string,
	cardName string,
	cardAddress string,
	residentID *string,
	devicesJSON []byte,
	residentsJSON []byte,
) (string, error) {
	query := `
		INSERT INTO cards (
			tenant_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			resident_id,
			devices,
			residents
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING card_id
	`

	var cardID string
	err := r.db.QueryRow(
		query,
		tenantID,
		cardType,
		bedID,
		unitID,
		cardName,
		cardAddress,
		residentID,
		devicesJSON,
		residentsJSON,
	).Scan(&cardID)

	if err != nil {
		return "", fmt.Errorf("failed to create card: %w", err)
	}

	return cardID, nil
}

// DeviceJSON device JSON format (for cards.devices JSONB field).
// 写 cards.devices 时需含 device_uid：DeviceJSON 加该字段，或组装 DeviceInfo 的查询（GetDevicesByBed、GetUnboundDevicesByUnit）从 devices 表选 d.device_uid 填入。
type DeviceJSON struct {
	DeviceID    string  `json:"device_id"`
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`    // Bed ID where device is bound (if bound to bed)
	BedName     *string `json:"bed_name,omitempty"`  // Bed name (if bound to bed)
	RoomID      *string `json:"room_id,omitempty"`   // Room ID where device is bound (if bound to room)
	RoomName    *string `json:"room_name,omitempty"` // Room name (if bound to room)
	UnitID      string  `json:"unit_id"`             // Unit ID where device is bound
}

// ResidentJSON resident JSON format (for cards.residents JSONB field)
type ResidentJSON struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
}

// UpdateAllCardsAlarmCounts 批量更新所有卡片的报警计数（用于服务启动和定时任务）
func (r *CardRepository) UpdateAllCardsAlarmCounts(ctx context.Context, tenantID string) error {
	// 获取所有卡片 ID
	query := `SELECT card_id FROM cards WHERE tenant_id = $1`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all card IDs: %w", err)
	}
	defer rows.Close()

	var cardIDs []string
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			return fmt.Errorf("failed to scan card ID: %w", err)
		}
		cardIDs = append(cardIDs, cardID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating card IDs: %w", err)
	}

	// 批量更新每个卡片的报警计数
	successCount := 0
	errorCount := 0
	for _, cardID := range cardIDs {
		if err := r.UpdateCardAlarmCounts(ctx, tenantID, cardID); err != nil {
			r.logger.Warn("Failed to update alarm counts for card",
				zap.String("card_id", cardID),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	r.logger.Info("Completed updating alarm counts for all cards",
		zap.Int("total", len(cardIDs)),
		zap.Int("success", successCount),
		zap.Int("error", errorCount),
	)

	return nil
}

// UpdateCardAlarmCounts 更新单个卡片的未处理报警计数
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

// ConvertDevicesToJSON and ConvertResidentsToJSON are in owl-common/card/utils.go
