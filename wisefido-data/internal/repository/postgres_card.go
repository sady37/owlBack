package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"owl-common/card"
	"wisefido-data/internal/domain"

	"go.uber.org/zap"
)

// RepositoryInterface 卡片仓储接口（建卡/同步用），由 PostgresCardRepository 实现
type RepositoryInterface interface {
	GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error)
	GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedRow, error)
	GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error)
	GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error)
	GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error)
	GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error)
	GetCardsByUnit(tenantID, unitID string) ([]card.CardWithContent, error)
	DeleteCard(tenantID, cardID string) error
	DeleteCardsByUnit(tenantID, unitID string) error
	CreateCard(
		tenantID, cardType string,
		bedID *string, unitID, cardName, cardAddress, timezone string,
		residentID *string,
		devices []card.DeviceInfo, residents []card.ResidentInfo,
	) (string, error)
	GetAllUnits(tenantID string) ([]string, error)
	GetUnitIDByBedID(tenantID, bedID string) (string, error)
	CountCardsByTenant(tenantID string) (int, error)
	UpdateCard(
		tenantID, cardID string,
		cardType string,
		bedID *string, unitID, cardName, cardAddress, timezone string,
		residentID *string,
		devices []card.DeviceInfo, residents []card.ResidentInfo,
	) error
}

// PostgresCardRepository implements RepositoryInterface；写 DB，并记录 affected 供 config 通知
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

func (r *PostgresCardRepository) appendRecorded(op, tenantID, cardID, unitID string, affectedUIDs []string) {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	var u []string
	if len(affectedUIDs) > 0 {
		u = dedupeNonEmptyStrings(affectedUIDs)
	}
	r.recorded = append(r.recorded, domain.CardSyncAffected{
		TenantID:           tenantID,
		CardID:             cardID,
		UnitID:             unitID,
		Op:                 op,
		AffectedDeviceUIDs: u,
	})
}

// CardRowForCache 卡片+单元信息（用于生成 CardStatic 静态缓存）
type CardRowForCache struct {
	CardID           string
	TenantID         string
	CardType         string
	BedID            *string
	BedName          *string // ✅ 新增
	RoomID           *string // ✅ 新增
	RoomName         *string // ✅ 新增
	UnitID           string
	UnitName         string // ✅ 新增
	Building         string // ✅ 新增
	CardName         string
	CardAddress      string
	Timezone         string
	ResidentID       *string
	DevicesJSON      []byte
	ResidentsJSON    []byte
	BranchID         string
	BranchName       string
	IconAlarmLevel   int
	PopAlarm int
}

// convertDevicesToJSON 将设备列表转换为 JSONB 格式
func convertDevicesToJSON(devices []card.DeviceInfo) ([]byte, error) {
	if len(devices) == 0 {
		return []byte("[]"), nil
	}

	type DeviceJSON struct {
		DeviceID            string  `json:"device_id"`
		DeviceUID           string  `json:"device_uid,omitempty"`
		DeviceCode          string  `json:"device_code,omitempty"`
		DeviceName          string  `json:"device_name"`
		DeviceType          string  `json:"device_type"`
		DeviceModel         string  `json:"device_model"`
		BedID               *string `json:"bed_id,omitempty"`
		BedName             *string `json:"bed_name,omitempty"`
		RoomID              *string `json:"room_id,omitempty"`
		RoomName            *string `json:"room_name,omitempty"`
		UnitID              string  `json:"unit_id"`
		MonitoringEnabled bool    `json:"monitoring_enabled,omitempty"`
		Status              string  `json:"status,omitempty"`
	}

	var deviceJSONs []DeviceJSON
	for _, device := range devices {
		deviceJSONs = append(deviceJSONs, DeviceJSON{
			DeviceID:            device.DeviceID,
			DeviceUID:           device.DeviceUID,
			DeviceCode:          device.DeviceCode,
			DeviceName:          device.DeviceName,
			DeviceType:          fmt.Sprint(device.DeviceType), // 兼容数字和字符串类型
			DeviceModel:         device.DeviceModel,
			BedID:               device.BoundBedID,
			RoomID:              device.BoundRoomID,
			UnitID:              device.UnitID,
			MonitoringEnabled:   device.MonitoringEnabled,
			Status:              device.Status,
		})
	}

	return json.Marshal(deviceJSONs)
}

// convertResidentsToJSON 将住户列表转换为 JSONB 格式
func convertResidentsToJSON(residents []card.ResidentInfo) ([]byte, error) {
	if len(residents) == 0 {
		return []byte("[]"), nil
	}

	type ResidentJSON struct {
		ResidentID string `json:"resident_id"`
		Nickname   string `json:"nickname"`
	}

	var residentJSONs []ResidentJSON
	for _, resident := range residents {
		residentJSONs = append(residentJSONs, ResidentJSON{
			ResidentID: resident.ResidentID,
			Nickname:   resident.Nickname,
		})
	}

	return json.Marshal(residentJSONs)
}

// deserializeDevices 从 JSON 反序列化设备列表（与 owl-common/card.ParseDevicesFromCardsJSONB 一致）
func deserializeDevices(data []byte) ([]card.DeviceInfo, error) {
	return card.ParseDevicesFromCardsJSONB(data)
}

// deserializeResidents 从 JSON 反序列化住户列表
func deserializeResidents(data []byte) ([]card.ResidentInfo, error) {
	if len(data) == 0 || string(data) == "[]" {
		return []card.ResidentInfo{}, nil
	}

	type ResidentJSON struct {
		ResidentID string `json:"resident_id"`
		Nickname   string `json:"nickname"`
	}

	var residentJSONs []ResidentJSON
	if err := json.Unmarshal(data, &residentJSONs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal residents: %w", err)
	}

	residents := make([]card.ResidentInfo, 0, len(residentJSONs))
	for _, rj := range residentJSONs {
		residents = append(residents, card.ResidentInfo{
			ResidentID: rj.ResidentID,
			Nickname:   rj.Nickname,
		})
	}
	return residents, nil
}

// GetCardRowForCache 从 DB 取单卡+unit（branch_id、branch_name、icon_alarm_level、pop_alarm），用于写 CardStatic 静态缓存
func (r *PostgresCardRepository) GetCardRowForCache(ctx context.Context, tenantID, cardID string) (*CardRowForCache, error) {
	query := `
		SELECT c.card_id, c.tenant_id, c.card_type, c.bed_id, c.unit_id, c.card_name, c.card_address,
		       COALESCE(c.timezone, 'UTC'), c.resident_id, c.devices, c.residents,
		       COALESCE(u.branch_id::text, '') as branch_id,
		       COALESCE(b.branch_name, '') as branch_name,
		       COALESCE(u.unit_name, '') as unit_name,
			COALESCE(bld.building_name, '') as building,
		       COALESCE(c.icon_alarm_level, 2), COALESCE(c.pop_alarm, 2),
		       COALESCE(bed.bed_name, ''), COALESCE(room.room_id::text, ''), COALESCE(room.room_name, '')
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id AND c.tenant_id = u.tenant_id
		LEFT JOIN branches b ON u.branch_id = b.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		LEFT JOIN beds bed ON c.bed_id = bed.bed_id AND c.tenant_id = bed.tenant_id
		LEFT JOIN rooms room ON bed.room_id = room.room_id AND c.tenant_id = room.tenant_id
		WHERE c.tenant_id = $1 AND c.card_id = $2
	`
	var row CardRowForCache
	var bedID, residentID, bedName, roomID, roomName sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(
		&row.CardID, &row.TenantID, &row.CardType, &bedID, &row.UnitID, &row.CardName, &row.CardAddress,
		&row.Timezone, &residentID, &row.DevicesJSON, &row.ResidentsJSON, &row.BranchID, &row.BranchName,
		&row.UnitName, &row.Building,
		&row.IconAlarmLevel, &row.PopAlarm,
		&bedName, &roomID, &roomName,
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
	if bedName.Valid {
		row.BedName = &bedName.String
	}
	if roomID.Valid {
		row.RoomID = &roomID.String
	}
	if roomName.Valid {
		row.RoomName = &roomName.String
	}
	if residentID.Valid {
		row.ResidentID = &residentID.String
	}
	return &row, nil
}

// GetBranchIDByCard 查询 card 的 branch_id（直接从 cards 表，性能最佳）
// 优先使用此方法而非 GetBranchIDByUnit
func (r *PostgresCardRepository) GetBranchIDByCard(ctx context.Context, tenantID, cardID string) (string, error) {
	if tenantID == "" || cardID == "" {
		return "", nil
	}
	var branchID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(branch_id::text, '') FROM cards WHERE tenant_id = $1 AND card_id = $2`,
		tenantID, cardID,
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

// GetBranchIDByUnit 查询 unit 的 branch_id（供卡片创建前用，已创建的卡片应用 GetBranchIDByCard）
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

// GetActiveBedsByUnit 获取单元下所有活跃床位（床上有启用监控的设备）
func (r *PostgresCardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedRow, error) {
	query := `
		SELECT
			b.bed_id,
			COALESCE(b.bed_name, '') as bed_name,
			COALESCE(rm.room_name, '') as room_name,
			MAX(r2.resident_id::text) as resident_id
		FROM beds b
		INNER JOIN rooms rm ON b.room_id = rm.room_id AND b.tenant_id = rm.tenant_id
		INNER JOIN devices d ON d.bound_bed_id = b.bed_id AND b.tenant_id = d.tenant_id
		LEFT JOIN residents r2 ON r2.bed_id = b.bed_id AND r2.tenant_id = $1
		WHERE b.tenant_id = $1
		  AND rm.unit_id = $2
		  AND d.monitoring_enabled = TRUE
		  AND d.status <> 'disabled'
		GROUP BY b.bed_id, b.bed_name, rm.room_name
		HAVING COUNT(DISTINCT d.device_id) > 0
		ORDER BY bed_name ASC, b.bed_id ASC
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("query active beds: %w", err)
	}
	defer rows.Close()

	var beds []card.ActiveBedRow
	for rows.Next() {
		var bedID, bedName, roomName string
		var residentID sql.NullString

		if err := rows.Scan(&bedID, &bedName, &roomName, &residentID); err != nil {
			return nil, fmt.Errorf("scan bed: %w", err)
		}

		bed := card.ActiveBedRow{
			BedID:   bedID,
			BedName: bedName,
		}

		// RoomName 存储到结构体中（供卡片生成使用）
		if roomName != "" {
			bed.RoomName = roomName
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
			COALESCE(b.branch_name, '') as branch_name,
			COALESCE(bld.building_name, '') as building,
			u.is_public,
			u.is_shared_unit,
			u.unit_type,
			COALESCE(u.timezone, 'UTC') as timezone
		FROM units u
		LEFT JOIN branches b ON u.branch_id = b.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`

	var unit card.UnitInfo

	err := r.db.QueryRow(query, tenantID, unitID).Scan(
		&unit.UnitID,
		&unit.UnitName,
		&unit.BranchName,
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
			d.bound_room_id,
			COALESCE(r.unit_id, (SELECT unit_id FROM beds WHERE bed_id = d.bound_bed_id AND tenant_id = $1 LIMIT 1)) as unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_uid = ds.device_uid
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

	var devices []card.DeviceInfo
	for rows.Next() {
		var device card.DeviceInfo
		var boundBedID, boundRoomID, unitID sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.DeviceName,
			&device.DeviceType,
			&device.DeviceModel,
			&boundBedID,
			&boundRoomID,
			&unitID,
			&device.MonitoringEnabled,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if boundBedID.Valid {
			device.BoundBedID = &boundBedID.String
		}
		if boundRoomID.Valid {
			device.BoundRoomID = &boundRoomID.String
		}
		if unitID.Valid {
			device.UnitID = unitID.String
		}

		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate devices: %w", err)
	}

	return devices, nil
}

// GetUnboundDevicesByUnit gets all devices with monitoring_enabled = TRUE bound to a room in the specified unit
// (not bound to any bed)

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
			d.bound_room_id,
			u.unit_id,
			d.monitoring_enabled
		FROM devices d
		JOIN device_store ds ON d.device_uid = ds.device_uid
		LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
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
		var boundBedID, boundRoomID, unitIDFromQuery sql.NullString

		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.DeviceName,
			&device.DeviceType,
			&device.DeviceModel,
			&boundBedID,
			&boundRoomID,
			&unitIDFromQuery,
			&device.MonitoringEnabled,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if boundBedID.Valid {
			device.BoundBedID = &boundBedID.String
		}
		if boundRoomID.Valid {
			device.BoundRoomID = &boundRoomID.String
		}
		if unitIDFromQuery.Valid {
			device.UnitID = unitIDFromQuery.String
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
			COALESCE(rp.last_name, '') as last_name,
			r.nickname,
			r.bed_id
		FROM residents r
		LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
		WHERE r.tenant_id = $1
		  AND r.bed_id = $2
		LIMIT 1
	`

	var resident card.ResidentInfo
	var residentBedID sql.NullString
	var lastName string

	err := r.db.QueryRow(query, tenantID, bedID).Scan(
		&resident.ResidentID,
		&lastName,
		&resident.Nickname,
		&residentBedID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Bed is not bound to any resident
		}
		return nil, fmt.Errorf("failed to query resident: %w", err)
	}

	resident.LastName = lastName

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
			COALESCE(rp.last_name, '') as last_name,
			r.nickname,
			r.bed_id
		FROM residents r
		LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
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
		var bedID sql.NullString
		var lastName string

		if err := rows.Scan(
			&resident.ResidentID,
			&lastName,
			&resident.Nickname,
			&bedID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}

		resident.LastName = lastName

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
		var devicesJSON, residentsJSON []byte

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

		// Deserialize JSON to structured data
		devices, err := deserializeDevices(devicesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize devices for card %s: %w", cardItem.CardID, err)
		}
		residents, err := deserializeResidents(residentsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize residents for card %s: %w", cardItem.CardID, err)
		}

		cardItem.Devices = devices
		cardItem.Residents = residents

		cards = append(cards, cardItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate cards: %w", err)
	}

	return cards, nil
}

// DeleteCard 删除卡片
// 删除前将卡片关联设备的 active/acked 报警标记为 expired（metadata.expired_reason = "card_delete"）
func (r *PostgresCardRepository) DeleteCard(tenantID, cardID string) error {
	// 1. 获取卡片设备列表，expire 其报警
	deviceIDs := r.getCardDeviceIDList(cardID)
	if len(deviceIDs) > 0 {
		if err := card.ExpireAlarmsByDeviceIDs(context.Background(), r.db, tenantID, deviceIDs, "card_delete"); err != nil {
			r.logger.Warn("DeleteCard: failed to expire alarms",
				zap.String("card_id", cardID), zap.Error(err))
		}
	}

	var unitIDStr string
	_ = r.db.QueryRow(`SELECT unit_id::text FROM cards WHERE tenant_id = $1 AND card_id = $2`, tenantID, cardID).Scan(&unitIDStr)
	uids := r.getCardDeviceUIDList(tenantID, cardID)

	// 2. 删除卡片
	_, err := r.db.Exec("DELETE FROM cards WHERE tenant_id = $1 AND card_id = $2", tenantID, cardID)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	r.appendRecorded("delete", tenantID, cardID, unitIDStr, uids)

	return nil
}

// ListAllCardsForClear 删除前列出所有卡片（tenant_id, card_id, unit_id），用于推送 config:card:stream deleted
func (r *PostgresCardRepository) ListAllCardsForClear(ctx context.Context) ([]domain.CardSyncAffected, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tenant_id::text, card_id::text, COALESCE(unit_id::text, '') FROM cards`)
	if err != nil {
		return nil, fmt.Errorf("list cards for clear: %w", err)
	}
	defer rows.Close()
	var out []domain.CardSyncAffected
	for rows.Next() {
		var a domain.CardSyncAffected
		if err := rows.Scan(&a.TenantID, &a.CardID, &a.UnitID); err != nil {
			return nil, err
		}
		a.Op = "deleted"
		out = append(out, a)
	}
	return out, rows.Err()
}

// ClearAllCards 清理全局所有卡片记录（删除所有租户的卡片）
func (r *PostgresCardRepository) ClearAllCards() error {
	result, err := r.db.Exec("DELETE FROM cards")
	if err != nil {
		return fmt.Errorf("clear all cards: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Cleared all cards globally",
		zap.Int64("rows_affected", rowsAffected),
	)

	// 清空记录器，避免影响后续操作
	r.ClearRecorded()
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
		r.appendRecorded("deleted", tenantID, cid, unitID, r.getCardDeviceUIDList(tenantID, cid))
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
//
// 移除的设备的 active/acked 报警会被标记为 expired（metadata.expired_reason = "card_update"）
func (r *PostgresCardRepository) UpdateCard(
	tenantID, cardID string,
	cardType string,
	bedID *string, unitID, cardName, cardAddress, timezone string,
	residentID *string,
	devices []card.DeviceInfo, residents []card.ResidentInfo,
) error {
	if timezone == "" {
		timezone = "UTC"
	}

	// 1. 查询旧设备列表（用于比较移除的设备）
	oldDeviceIDs := r.getCardDeviceIDList(cardID)
	oldDeviceUIDs := r.getCardDeviceUIDList(tenantID, cardID)

	// 2. 将结构化数据转换为 JSON
	devicesJSON, err := convertDevicesToJSON(devices)
	if err != nil {
		return fmt.Errorf("failed to convert devices to JSON: %w", err)
	}
	residentsJSON, err := convertResidentsToJSON(residents)
	if err != nil {
		return fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	// Get branch_id from units table (for denormalization)
	var branchID *string
	if unitID != "" {
		var bid sql.NullString
		err := r.db.QueryRow(
			`SELECT branch_id::text FROM units WHERE tenant_id = $1 AND unit_id = $2`,
			tenantID, unitID,
		).Scan(&bid)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get branch_id from unit: %w", err)
		}
		if bid.Valid && bid.String != "" {
			branchID = &bid.String
		}
	}

	query := `
		UPDATE cards
		SET
			branch_id = $3,
			card_type = $4,
			bed_id = $5,
			unit_id = $6,
			card_name = $7,
			card_address = $8,
			timezone = $9,
			resident_id = $10,
			devices = $11,
			residents = $12
		WHERE tenant_id = $1
		  AND card_id = $2
	`

	result, err := r.db.Exec(
		query,
		tenantID,
		cardID,
		branchID,
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

	// 3. 找出被移除的设备，expire 其报警
	newDeviceSet := make(map[string]bool, len(devices))
	for _, d := range devices {
		newDeviceSet[d.DeviceID] = true
	}
	var removedIDs []string
	for _, id := range oldDeviceIDs {
		if !newDeviceSet[id] {
			removedIDs = append(removedIDs, id)
		}
	}
	if len(removedIDs) > 0 {
		if err := card.ExpireAlarmsByDeviceIDs(context.Background(), r.db, tenantID, removedIDs, "card_update"); err != nil {
			r.logger.Warn("UpdateCard: failed to expire alarms for removed devices",
				zap.String("card_id", cardID), zap.Error(err))
		}
	}

	// 4. owl-common 重算报警计数（devices 可能变更）
	if _, err := card.RecalcCardAlarmState(context.Background(), r.db, cardID, tenantID); err != nil {
		r.logger.Warn("UpdateCard: failed to recalc alarm counts",
			zap.String("card_id", cardID), zap.Error(err))
	}
	affectedUIDs := unionDeviceUIDSlices(oldDeviceUIDs, deviceUIDsFromDeviceInfos(devices))
	r.appendRecorded("updated", tenantID, cardID, unitID, affectedUIDs)
	return nil
}

// CreateCard creates a card and initializes alarm counts via owl-common
func (r *PostgresCardRepository) CreateCard(
	tenantID string,
	cardType string, // "ActiveBedCard" or "UnitCard"
	bedID *string,
	unitID string,
	cardName string,
	cardAddress string,
	timezone string,
	residentID *string,
	devices []card.DeviceInfo,
	residents []card.ResidentInfo,
) (string, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	// 将结构化数据转换为 JSON
	devicesJSON, err := convertDevicesToJSON(devices)
	if err != nil {
		return "", fmt.Errorf("failed to convert devices to JSON: %w", err)
	}
	residentsJSON, err := convertResidentsToJSON(residents)
	if err != nil {
		return "", fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	// Get branch_id from units table (for denormalization)
	var branchID *string
	if unitID != "" {
		var bid sql.NullString
		err := r.db.QueryRow(
			`SELECT branch_id::text FROM units WHERE tenant_id = $1 AND unit_id = $2`,
			tenantID, unitID,
		).Scan(&bid)
		if err != nil && err != sql.ErrNoRows {
			return "", fmt.Errorf("failed to get branch_id from unit: %w", err)
		}
		if bid.Valid && bid.String != "" {
			branchID = &bid.String
		}
	}

	query := `
		INSERT INTO cards (
			tenant_id,
			branch_id,
			card_type,
			bed_id,
			unit_id,
			card_name,
			card_address,
			timezone,
			resident_id,
			devices,
			residents
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING card_id
	`

	var cardID string
	err = r.db.QueryRow(
		query,
		tenantID,
		branchID,
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
	// owl-common 初始化报警计数（统计 active alarms + pop_alarm）
	if _, err := card.InitializeCardAlarmCounts(context.Background(), r.db, cardID, tenantID); err != nil {
		r.logger.Warn("CreateCard: failed to initialize alarm counts",
			zap.String("card_id", cardID), zap.Error(err))
	}
	r.appendRecorded("created", tenantID, cardID, unitID, deviceUIDsFromDeviceInfos(devices))
	return cardID, nil
}

// getCardDeviceIDList 从 cards.devices JSONB 提取 device_id 列表（内部用）
func (r *PostgresCardRepository) getCardDeviceIDList(cardID string) []string {
	var devicesJSON []byte
	err := r.db.QueryRow(`SELECT devices FROM cards WHERE card_id = $1`, cardID).Scan(&devicesJSON)
	if err != nil || len(devicesJSON) == 0 {
		return nil
	}
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil
	}
	ids := make([]string, 0, len(devices))
	for _, d := range devices {
		if did, ok := d["device_id"].(string); ok && did != "" {
			ids = append(ids, did)
		}
	}
	return ids
}

// GetDeviceUIDsForCard 从 cards.devices JSONB 提取 device_uid（供 config 消息补全 affected_device_uids）；须 tenant+card 双条件以符合多租户隔离。
func (r *PostgresCardRepository) GetDeviceUIDsForCard(tenantID, cardID string) []string {
	return dedupeNonEmptyStrings(r.getCardDeviceUIDList(tenantID, cardID))
}

func getCardDeviceUIDListFromJSON(devicesJSON []byte) []string {
	if len(devicesJSON) == 0 {
		return nil
	}
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil
	}
	uids := make([]string, 0, len(devices))
	for _, d := range devices {
		if uid, ok := d["device_uid"].(string); ok {
			if s := strings.TrimSpace(uid); s != "" {
				uids = append(uids, s)
			}
		}
	}
	return uids
}

func (r *PostgresCardRepository) getCardDeviceUIDList(tenantID, cardID string) []string {
	if tenantID == "" || cardID == "" {
		return nil
	}
	var devicesJSON []byte
	err := r.db.QueryRow(
		`SELECT devices FROM cards WHERE tenant_id = $1 AND card_id = $2`,
		tenantID, cardID,
	).Scan(&devicesJSON)
	if err != nil || len(devicesJSON) == 0 {
		return nil
	}
	return getCardDeviceUIDListFromJSON(devicesJSON)
}

func deviceUIDsFromDeviceInfos(devices []card.DeviceInfo) []string {
	if len(devices) == 0 {
		return nil
	}
	out := make([]string, 0, len(devices))
	for _, d := range devices {
		if s := strings.TrimSpace(d.DeviceUID); s != "" {
			out = append(out, s)
		}
	}
	return dedupeNonEmptyStrings(out)
}

func unionDeviceUIDSlices(a, b []string) []string {
	if len(a) == 0 {
		return dedupeNonEmptyStrings(b)
	}
	if len(b) == 0 {
		return dedupeNonEmptyStrings(a)
	}
	return dedupeNonEmptyStrings(append(append([]string{}, a...), b...))
}

func dedupeNonEmptyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ConvertDevicesToJSON and ConvertResidentsToJSON are in owl-common/card/utils.go

// CreateDeviceCard 为未绑定 card 的设备创建 DeviceCard（card_id = device_id）
// DeviceCard 的 unit_id 和 bed_id 均为 NULL，devices JSONB 包含该设备信息
func (r *PostgresCardRepository) CreateDeviceCard(tenantID string, device card.DeviceInfo) (string, error) {
	if tenantID == "" || device.DeviceID == "" {
		return "", fmt.Errorf("tenant_id and device_id are required")
	}

	devicesJSON, err := convertDevicesToJSON([]card.DeviceInfo{device})
	if err != nil {
		return "", fmt.Errorf("failed to convert device to JSON: %w", err)
	}

	// card_name = device_name, card_address = device_uid
	cardName := device.DeviceName
	if cardName == "" {
		cardName = device.DeviceUID
	}
	cardAddress := device.DeviceUID

	query := `
		INSERT INTO cards (
			card_id, tenant_id, card_type,
			bed_id, unit_id,
			card_name, card_address, timezone,
			resident_id, devices, residents
		) VALUES ($1, $2, 'DeviceCard', NULL, NULL, $3, $4, 'UTC', NULL, $5, '[]'::jsonb)
		ON CONFLICT (card_id) DO UPDATE SET
			card_name = EXCLUDED.card_name,
			card_address = EXCLUDED.card_address,
			devices = EXCLUDED.devices
		RETURNING card_id
	`

	var cardID string
	err = r.db.QueryRow(query,
		device.DeviceID, // card_id = device_id
		tenantID,
		cardName,
		cardAddress,
		devicesJSON,
	).Scan(&cardID)
	if err != nil {
		return "", fmt.Errorf("failed to create device card: %w", err)
	}

	r.appendRecorded("created", tenantID, cardID, "", []string{device.DeviceUID})
	return cardID, nil
}

// DeleteDeviceCard 删除 DeviceCard（设备绑定到 unit 后清理）
func (r *PostgresCardRepository) DeleteDeviceCard(tenantID, deviceID string) error {
	if tenantID == "" || deviceID == "" {
		return nil
	}
	result, err := r.db.Exec(
		`DELETE FROM cards WHERE tenant_id = $1 AND card_id = $2 AND card_type = 'DeviceCard'`,
		tenantID, deviceID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete device card: %w", err)
	}
	if n, _ := result.RowsAffected(); n > 0 {
		r.appendRecorded("deleted", tenantID, deviceID, "", nil)
	}
	return nil
}

// CleanupDeviceCardsByUnit 清理指定 unit 中所有设备的 DeviceCard
// 当设备被纳入 unit card（ActiveBedCard/UnitCard）后，独立的 DeviceCard 不再需要
func (r *PostgresCardRepository) CleanupDeviceCardsByUnit(tenantID, unitID string) (int, error) {
	if tenantID == "" || unitID == "" {
		return 0, nil
	}

	// 找出该 unit 下所有设备的 device_id，删除对应的 DeviceCard
	query := `
		DELETE FROM cards
		WHERE tenant_id = $1
		  AND card_type = 'DeviceCard'
		  AND card_id IN (
			SELECT d.device_id FROM devices d
			JOIN rooms r ON d.bound_room_id = r.room_id
			JOIN units u ON r.unit_id = u.unit_id
			WHERE u.tenant_id = $1 AND u.unit_id = $2
			UNION
			SELECT d.device_id FROM devices d
			JOIN beds b ON d.bound_bed_id = b.bed_id
			JOIN rooms r ON b.room_id = r.room_id
			JOIN units u ON r.unit_id = u.unit_id
			WHERE u.tenant_id = $1 AND u.unit_id = $2
		  )
		RETURNING card_id
	`

	rows, err := r.db.Query(query, tenantID, unitID)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup device cards by unit: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err == nil {
			r.appendRecorded("deleted", tenantID, cardID, "", nil)
			count++
		}
	}
	return count, rows.Err()
}

// GetDevicesWithoutCard 查询所有没有出现在任何 card 的 devices JSONB 中的设备
// 返回可直接用于创建 DeviceCard 的 DeviceInfo 列表
func (r *PostgresCardRepository) GetDevicesWithoutCard(tenantID string) ([]card.DeviceInfo, error) {
	if tenantID == "" {
		return nil, nil
	}

	query := `
		SELECT
			d.device_id::text,
			d.device_uid,
			COALESCE(d.device_name, ''),
			COALESCE(ds.device_type, ''),
			COALESCE(ds.device_model, ''),
			d.status
		FROM devices d
		LEFT JOIN device_store ds ON ds.device_uid = d.device_uid
		WHERE d.tenant_id = $1
		  AND NOT EXISTS (
			SELECT 1 FROM cards c
			WHERE c.tenant_id = $1
			  AND c.devices @> jsonb_build_array(jsonb_build_object('device_id', d.device_id::text))
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM cards c2
			WHERE c2.tenant_id = $1
			  AND c2.card_id = d.device_id
			  AND c2.card_type = 'DeviceCard'
		  )
	`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices without card: %w", err)
	}
	defer rows.Close()

	var devices []card.DeviceInfo
	for rows.Next() {
		var di card.DeviceInfo
		if err := rows.Scan(&di.DeviceID, &di.DeviceUID, &di.DeviceName, &di.DeviceType, &di.DeviceModel, &di.Status); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, di)
	}
	return devices, rows.Err()
}

// CleanOrphanDeviceCards 清理孤儿 DeviceCard：card_type='DeviceCard' 但对应 devices 行已不在或租户不一致。
// 典型成因：设备跨租户迁移时未同步清理旧租户的 DeviceCard。返回被删除的行数。
func (r *PostgresCardRepository) CleanOrphanDeviceCards(ctx context.Context) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM cards
		WHERE card_type = 'DeviceCard'
		  AND card_id IN (
			SELECT c.card_id FROM cards c
			LEFT JOIN devices d ON d.device_id = c.card_id
			WHERE c.card_type = 'DeviceCard'
			  AND (d.device_id IS NULL OR d.tenant_id <> c.tenant_id)
		  )
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to clean orphan device cards: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
