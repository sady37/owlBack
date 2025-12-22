package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// CardRepository card repository
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

// ActiveBedInfo ActiveBed information
type ActiveBedInfo struct {
	BedID            string
	UnitID           string
	BoundDeviceCount int
	ResidentID       *string
	RoomID           string
}

// UnitInfo Unit information
type UnitInfo struct {
	UnitID            string
	UnitName          string
	BranchName        string
	Building          string
	IsPublicSpace     bool
	IsMultiPersonRoom bool
	UnitType          string
	GroupList         []byte // JSONB format, user group list (for alarm routing)
	UserList          []byte // JSONB format, user ID list (for alarm routing)
}

// DeviceInfo device information
type DeviceInfo struct {
	DeviceID          string
	DeviceName        string
	DeviceType        string
	DeviceModel       string
	BoundBedID        *string
	BedName           *string // Bed name (if bound to bed)
	BoundRoomID       *string // Room ID where device is bound (if bound to room)
	RoomName          *string // Room name (if bound to room)
	UnitID            string
	MonitoringEnabled bool
}

// ResidentInfo resident information
type ResidentInfo struct {
	ResidentID string
	Nickname   string
	UnitID     *string
	BedID      *string
}

// GetActiveBedsByUnit gets all ActiveBeds under the specified unit
// ActiveBed condition: 床上有 monitoring_enabled = TRUE 的设备即可
// 注意：bed_type 字段已删除，改为动态查询设备绑定状态
func (r *CardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]ActiveBedInfo, error) {
	query := `
		SELECT DISTINCT
			b.bed_id,
			r.unit_id,
			COUNT(DISTINCT d.device_id)::int AS bound_device_count,
			r2.resident_id,
			b.room_id
		FROM beds b
		INNER JOIN rooms r ON b.room_id = r.room_id
		INNER JOIN devices d ON d.bound_bed_id = b.bed_id
		LEFT JOIN residents r2 ON r2.bed_id = b.bed_id AND r2.tenant_id = $1
		WHERE b.tenant_id = $1
		  AND r.unit_id = $2
		  AND d.monitoring_enabled = TRUE
		  AND d.status <> 'disabled'
		GROUP BY b.bed_id, r.unit_id, r2.resident_id, b.room_id
		HAVING COUNT(DISTINCT d.device_id) > 0
		ORDER BY b.bed_name
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active beds: %w", err)
	}
	defer rows.Close()

	var beds []ActiveBedInfo
	for rows.Next() {
		var bed ActiveBedInfo
		var residentID sql.NullString

		if err := rows.Scan(
			&bed.BedID,
			&bed.UnitID,
			&bed.BoundDeviceCount,
			&residentID,
			&bed.RoomID,
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
func (r *CardRepository) GetUnitInfo(tenantID, unitID string) (*UnitInfo, error) {
	query := `
		SELECT 
			unit_id,
			unit_name,
			branch_name,
			building,
			is_public_space,
			is_multi_person_room,
			unit_type,
			groupList,
			userList
		FROM units
		WHERE tenant_id = $1 AND unit_id = $2
	`

	var unit UnitInfo
	var groupList, userList sql.NullString

	err := r.db.QueryRow(query, tenantID, unitID).Scan(
		&unit.UnitID,
		&unit.UnitName,
		&unit.BranchName,
		&unit.Building,
		&unit.IsPublicSpace,
		&unit.IsMultiPersonRoom,
		&unit.UnitType,
		&groupList,
		&userList,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found: %s", unitID)
		}
		return nil, fmt.Errorf("failed to query unit: %w", err)
	}

	// Handle JSONB fields
	if groupList.Valid {
		unit.GroupList = []byte(groupList.String)
	} else {
		unit.GroupList = []byte("[]")
	}

	if userList.Valid {
		unit.UserList = []byte(userList.String)
	} else {
		unit.UserList = []byte("[]")
	}

	return &unit, nil
}

// GetDevicesByBed gets all devices with monitoring_enabled = TRUE bound to the specified bed
func (r *CardRepository) GetDevicesByBed(tenantID, bedID string) ([]DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			d.device_name,
			ds.device_type,
			ds.device_model,
			d.bound_bed_id,
			b.bed_name,
			d.bound_room_id,
			r.room_name,
			d.unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
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

	var devices []DeviceInfo
	for rows.Next() {
		var device DeviceInfo
		var boundBedID, bedName, boundRoomID, roomName sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
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

// GetUnboundDevicesByUnit gets all devices with monitoring_enabled = TRUE that are not bound to any bed in the specified unit
func (r *CardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]DeviceInfo, error) {
	query := `
		SELECT 
			d.device_id,
			d.device_name,
			ds.device_type,
			ds.device_model,
			d.bound_bed_id,
			b.bed_name,
			d.bound_room_id,
			r.room_name,
			d.unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
		WHERE d.tenant_id = $1
		  AND d.unit_id = $2
		  AND d.bound_bed_id IS NULL
		  AND d.monitoring_enabled = TRUE
		ORDER BY d.device_name
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unbound devices: %w", err)
	}
	defer rows.Close()

	var devices []DeviceInfo
	for rows.Next() {
		var device DeviceInfo
		var boundBedID, bedName, boundRoomID, roomName sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
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
func (r *CardRepository) GetResidentByBed(tenantID, bedID string) (*ResidentInfo, error) {
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

	var resident ResidentInfo
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
func (r *CardRepository) GetResidentsByUnit(tenantID, unitID string) ([]ResidentInfo, error) {
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

	var residents []ResidentInfo
	for rows.Next() {
		var resident ResidentInfo
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

// DeviceJSON device JSON format (for cards.devices JSONB field)
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

// ConvertDevicesToJSON converts device list to JSON
func ConvertDevicesToJSON(devices []DeviceInfo) ([]byte, error) {
	var deviceJSONs []DeviceJSON
	for _, device := range devices {
		deviceJSONs = append(deviceJSONs, DeviceJSON{
			DeviceID:    device.DeviceID,
			DeviceName:  device.DeviceName,
			DeviceType:  device.DeviceType,
			DeviceModel: device.DeviceModel,
			BedID:       device.BoundBedID,
			BedName:     device.BedName,
			RoomID:      device.BoundRoomID,
			RoomName:    device.RoomName,
			UnitID:      device.UnitID,
		})
	}
	return json.Marshal(deviceJSONs)
}

// ConvertResidentsToJSON converts resident list to JSON
func ConvertResidentsToJSON(residents []ResidentInfo) ([]byte, error) {
	var residentJSONs []ResidentJSON
	for _, resident := range residents {
		residentJSONs = append(residentJSONs, ResidentJSON{
			ResidentID: resident.ResidentID,
			Nickname:   resident.Nickname,
		})
	}
	return json.Marshal(residentJSONs)
}
