package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"owl-common/card"
	"wisefido-data/internal/domain"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresCardRepository implements card.RepositoryInterface；写 DB，并记录 affected 供 config 通知
type PostgresCardRepository struct {
	db         *sql.DB
	logger     *zap.Logger
	recorded   []domain.CardSyncAffected
	recordedMu sync.Mutex
}

// NewPostgresCardRepository creates a new card repository
func NewPostgresCardRepository(db *sql.DB, logger *zap.Logger) *PostgresCardRepository {
	return &PostgresCardRepository{
		db:     db,
		logger: logger,
	}
}

// ClearRecorded 清空本次同步记录（CreateCardsForUnit 前调用）
func (r *PostgresCardRepository) ClearRecorded() {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	r.recorded = nil
}

// GetRecordedAndClear 返回本次同步的 created/updated/deleted 并清空
func (r *PostgresCardRepository) GetRecordedAndClear() []domain.CardSyncAffected {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	out := r.recorded
	r.recorded = nil
	return out
}

func (r *PostgresCardRepository) appendRecorded(op, tenantID, cardID, unitID string) {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	r.recorded = append(r.recorded, domain.CardSyncAffected{TenantID: tenantID, CardID: cardID, UnitID: unitID, Op: op})
}

// CardForNotify 供 config.card.* 通知用（从 DB 读）
type CardForNotify struct {
	TenantID    string
	CardID      string
	UnitID      string
	DevicesJSON []byte
}

// CardRowForCache 单卡行+unit 信息，用于生成 VitalFocusCardInfo 静态缓存
type CardRowForCache struct {
	CardID         string
	TenantID       string
	CardType       string
	BedID          *string
	UnitID         string
	CardName       string
	CardAddress    string
	Timezone       string
	ResidentID     *string
	DevicesJSON    []byte
	ResidentsJSON  []byte
	BranchID       string
	BranchName     string
	IconAlarmLevel int
	PopAlarmEmerge int
}

// GetCardRowForCache 从 DB 取单卡+unit（branch_id、branch_name、icon_alarm_level、pop_alarm_emerge），用于写 VitalFocusCardInfo 静态缓存
func (r *PostgresCardRepository) GetCardRowForCache(ctx context.Context, tenantID, cardID string) (*CardRowForCache, error) {
	query := `
		SELECT c.card_id, c.tenant_id, c.card_type, c.bed_id, c.unit_id, c.card_name, c.card_address,
		       COALESCE(c.timezone, 'UTC'), c.resident_id, c.devices, c.residents,
		       COALESCE(u.branch_id::text, '') as branch_id,
		       COALESCE(b.branch_name, '') as branch_name,
		       COALESCE(c.icon_alarm_level, 3), COALESCE(c.pop_alarm_emerge, 0)
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id AND c.tenant_id = u.tenant_id
		LEFT JOIN branches b ON u.branch_id = b.branch_id
		WHERE c.tenant_id = $1 AND c.card_id = $2
	`
	var row CardRowForCache
	var bedID, residentID sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(
		&row.CardID, &row.TenantID, &row.CardType, &bedID, &row.UnitID, &row.CardName, &row.CardAddress,
		&row.Timezone, &residentID, &row.DevicesJSON, &row.ResidentsJSON, &row.BranchID, &row.BranchName,
		&row.IconAlarmLevel, &row.PopAlarmEmerge,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if bedID.Valid {
		row.BedID = &bedID.String
	}
	if residentID.Valid {
		row.ResidentID = &residentID.String
	}
	return &row, nil
}

// GetBranchIDByUnit 查询 unit 的 branch_id（供 card 变更时失效 user:cards 缓存）
func (r *PostgresCardRepository) GetBranchIDByUnit(ctx context.Context, tenantID, unitID string) (string, error) {
	if tenantID == "" || unitID == "" {
		return "", nil
	}
	var branchID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT branch_id::text FROM units WHERE tenant_id = $1 AND unit_id = $2`,
		tenantID, unitID,
	).Scan(&branchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if branchID.Valid && branchID.String != "" {
		return branchID.String, nil
	}
	return "", nil
}

// GetCardByID 从 DB 按 card_id 取卡片（供 emitCardChange 取 deviceIDs）
func (r *PostgresCardRepository) GetCardByID(ctx context.Context, tenantID, cardID string) (*CardForNotify, error) {
	query := `
		SELECT tenant_id, card_id, unit_id, devices
		FROM cards
		WHERE tenant_id = $1 AND card_id = $2
	`
	var tenantIDOut, cardIDOut, unitID string
	var devicesJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&tenantIDOut, &cardIDOut, &unitID, &devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &CardForNotify{
		TenantID:    tenantIDOut,
		CardID:      cardIDOut,
		UnitID:      unitID,
		DevicesJSON: devicesJSON,
	}, nil
}

// GetActiveBedsByUnit gets all ActiveBeds under the specified unit
// ActiveBed condition: 床上有 monitoring_enabled = TRUE 的设备即可
// 注意：bed_type 字段已删除，改为动态查询设备绑定状态
func (r *PostgresCardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedRow, error) {
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

	var beds []card.ActiveBedRow
	for rows.Next() {
		var bed card.ActiveBedRow
		var residentID sql.NullString
		var bedName sql.NullString

		if err := rows.Scan(
			&bed.BedID,
			&bed.UnitID,
			&bed.BoundDeviceCount,
			&residentID,
			&bed.RoomID,
			&bedName,
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
func (r *PostgresCardRepository) GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error) {
	query := `
		SELECT 
			u.unit_id,
			u.unit_name,
			COALESCE(u.branch_id::text, '') as branch_id,
			COALESCE(bld.building_name, '') as building,
			u.is_public,
			u.is_shared_unit,
			u.unit_type,
			COALESCE(u.timezone, 'UTC') as timezone
		FROM units u
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`

	var unit card.UnitInfo

	err := r.db.QueryRow(query, tenantID, unitID).Scan(
		&unit.UnitID,
		&unit.UnitName,
		&unit.BranchID,
		&unit.Building,
		&unit.IsPublic,
		&unit.IsSharedUnit,
		&unit.UnitType,
		&unit.Timezone,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found: %s", unitID)
		}
		return nil, fmt.Errorf("failed to query unit: %w", err)
	}

	return &unit, nil
}

// GetDevicesByBed gets all devices with monitoring_enabled = TRUE bound to the specified bed.
// 写 cards.devices 时需含 device_uid：从 devices 表选 d.device_uid 填入 DeviceInfo，供 ConvertDevicesToJSON 写入 JSON。
func (r *PostgresCardRepository) GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			COALESCE(d.device_uid, '') AS device_uid,
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
		JOIN device_store ds ON d.device_uid = ds.device_uid
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

// GetUnboundDevicesByUnit gets all devices with monitoring_enabled = TRUE that are bound to a room in the specified unit but not bound to any bed.
// 写 cards.devices 时需含 device_uid：从 devices 表选 d.device_uid 填入 DeviceInfo，供 ConvertDevicesToJSON 写入 JSON。
func (r *PostgresCardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			COALESCE(d.device_uid, '') AS device_uid,
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
		JOIN device_store ds ON d.device_uid = ds.device_uid
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
func (r *PostgresCardRepository) GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error) {
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
func (r *PostgresCardRepository) GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error) {
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
func (r *PostgresCardRepository) GetAllUnits(tenantID string) ([]string, error) {
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
func (r *PostgresCardRepository) GetUnitIDByBedID(tenantID, bedID string) (string, error) {
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
func (r *PostgresCardRepository) GetCardsByUnit(tenantID, unitID string) ([]card.CardWithContent, error) {
	query := `
		SELECT 
			card_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			COALESCE(timezone, 'UTC'),
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
			&cardItem.Timezone,
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
func (r *PostgresCardRepository) DeleteCard(tenantID, cardID string) error {
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
func (r *PostgresCardRepository) DeleteCardsByUnit(tenantID, unitID string) error {
	rows, err := r.db.Query(`SELECT card_id FROM cards WHERE tenant_id = $1 AND unit_id = $2`, tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to list cards: %w", err)
	}
	var cardIDs []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			rows.Close()
			return err
		}
		cardIDs = append(cardIDs, cid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, cid := range cardIDs {
		r.appendRecorded("deleted", tenantID, cid, unitID)
	}
	_, err = r.db.Exec(`DELETE FROM cards WHERE tenant_id = $1 AND unit_id = $2`, tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to delete cards: %w", err)
	}
	return nil
}

// CountCardsByTenant counts total number of cards for a tenant
func (r *PostgresCardRepository) CountCardsByTenant(tenantID string) (int, error) {
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
func (r *PostgresCardRepository) UpdateCard(
	tenantID, cardID string,
	cardType string,
	bedID *string, unitID, cardName, cardAddress, timezone string,
	residentID *string,
	devicesJSON, residentsJSON []byte,
) error {
	if timezone == "" {
		timezone = "UTC"
	}
	query := `
		UPDATE cards
		SET
			card_type = $3,
			bed_id = $4,
			unit_id = $5,
			card_name = $6,
			card_address = $7,
			timezone = $8,
			resident_id = $9,
			devices = $10,
			residents = $11
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
		timezone,
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
	r.appendRecorded("updated", tenantID, cardID, unitID)
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
func (r *PostgresCardRepository) CreateCard(
	tenantID string,
	cardType string, // "ActiveBed" or "Location"
	bedID *string,
	unitID string,
	cardName string,
	cardAddress string,
	timezone string,
	residentID *string,
	devicesJSON []byte,
	residentsJSON []byte,
) (string, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	query := `
		INSERT INTO cards (
			tenant_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			timezone,
			resident_id,
			devices,
			residents
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		timezone,
		residentID,
		devicesJSON,
		residentsJSON,
	).Scan(&cardID)

	if err != nil {
		return "", fmt.Errorf("failed to create card: %w", err)
	}
	r.appendRecorded("created", tenantID, cardID, unitID)
	return cardID, nil
}

// UpdateCardAlarmCounts 更新卡片的未处理报警计数
// 通过 card_id 找到该卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
// 统计这些设备的未处理报警（alarm_status = 'active' 且 alarm_level IN ('0'-'4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')）
// 按 alarm_level 分组统计（映射到 0-4），更新 cards 表的 unhandled_alarm_0 到 unhandled_alarm_4 字段
func (r *PostgresCardRepository) UpdateCardAlarmCounts(ctx context.Context, tenantID, cardID string) error {
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
	// 注意：alarm_level 可能是数字格式（'0', '1', '2', '3', '4'）或文本格式（'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING'）
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
func (r *PostgresCardRepository) updateCardAlarmCountsToZero(ctx context.Context, tenantID, cardID string) error {
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

// UpdateAllCardsAlarmCounts 批量更新所有卡片的报警计数（用于服务启动和定时任务）
func (r *PostgresCardRepository) UpdateAllCardsAlarmCounts(ctx context.Context, tenantID string) error {
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

// ConvertDevicesToJSON and ConvertResidentsToJSON are in owl-common/card/utils.go
