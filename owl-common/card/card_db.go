package card

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// DeviceCardMapping is the identity tuple resolved from device key (DB 联合查询直接 Scan).
// Used by all gateways to populate redis.IoTStreamMessage header and DB operations.
type DeviceCardMapping struct {
	DeviceUID  string // resolved hardware identifier (always populated)
	TenantID   string
	BranchID   string
	UnitID     string
	CardID     string
	DeviceID   string // database device_id (UUID)
	RoomID     string // device bound room_id (optional, gateway may fill)
	BedID      string // device bound bed_id (optional, gateway may fill)
	DeviceCode string // device_store.device_code (e.g. Sleepace MQTT deviceId)
	DeviceType string // device_store.device_type (e.g. SleepPad, Radar)
}

// CardDB wraps *sql.DB for card-related queries.
// All gateways share this; callers never import "database/sql" directly.
type CardDB struct {
	db *sql.DB
}

func NewCardDB(db *sql.DB) *CardDB {
	return &CardDB{db: db}
}

// LoadCodeToUIDMap batch loads device_code → device_uid from device_store.
func (c *CardDB) LoadCodeToUIDMap(ctx context.Context, deviceTypes []string) (map[string]string, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT device_code, device_uid
		FROM device_store
		WHERE device_code IS NOT NULL AND device_code != ''
		  AND device_uid  IS NOT NULL AND device_uid  != ''
		  AND device_type = ANY($1::text[])
	`, pq.Array(deviceTypes))
	if err != nil {
		return nil, fmt.Errorf("query device_store: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var code, uid string
		if err := rows.Scan(&code, &uid); err != nil {
			continue
		}
		m[code] = uid
	}
	return m, rows.Err()
}

// ResolveDevice resolves a single device_uid or device_code → (device_uid, device_code).
func (c *CardDB) ResolveDevice(ctx context.Context, deviceKey string) (deviceUID, deviceCode string, err error) {
	err = c.db.QueryRowContext(ctx, `
		SELECT device_uid, device_code
		FROM device_store
		WHERE device_uid = $1 OR device_code = $1
		LIMIT 1
	`, deviceKey).Scan(&deviceUID, &deviceCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("device not in device_store: %s", deviceKey)
		}
		return "", "", fmt.Errorf("resolve device: %w", err)
	}
	return
}

// LookupCard resolves a device key → full identity tuple (device_store + cards + devices).
// deviceKey can be device_uid ("BM87...") or device_code ("1ua3erivl9pv1"); 入参用 code 时 WITH 子句先解析为 uid。
func (c *CardDB) LookupCard(ctx context.Context, deviceKey string) (*DeviceCardMapping, error) {
	var m DeviceCardMapping
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_store WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT d.device_uid, c.tenant_id, c.branch_id,
		       COALESCE(c.unit_id::text, ''), c.card_id::text,
		       COALESCE(d.device_id, ''),
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, '')
		FROM resolved r
		LEFT JOIN device_store ds ON ds.device_uid = r.uid
		, cards c, jsonb_to_recordset(c.devices) AS d(device_uid text, device_id text)
		WHERE d.device_uid = r.uid
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.BranchID, &m.UnitID, &m.CardID, &m.DeviceID, &m.DeviceCode, &m.DeviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in cards: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup card: %w", err)
	}
	return &m, nil
}

// LookupDeviceOnly resolves device key when no card exists (devices + device_store).
// Returns DeviceCardMapping with CardID = DeviceID (virtual card).
func (c *CardDB) LookupDeviceOnly(ctx context.Context, deviceKey string) (*DeviceCardMapping, error) {
	var m DeviceCardMapping
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_store WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT d.device_uid, d.tenant_id, d.device_id,
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, '')
		FROM resolved r
		JOIN devices d ON d.device_uid = r.uid
		LEFT JOIN device_store ds ON ds.device_uid = r.uid
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.DeviceID, &m.DeviceCode, &m.DeviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in devices table: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device: %w", err)
	}
	m.CardID = m.DeviceID
	return &m, nil
}

// InitializeCardAlarmCounts 在 card INSERT 之后调用，初始化报警计数。
func InitializeCardAlarmCounts(ctx context.Context, db *sql.DB, cardID, tenantID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := recalcAndUpdateCards(ctx, tx, cardID, tenantID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return state, nil
}

// UpdateDeviceStoreReportedVersion 首次上连或设备上报版本时：用 device_uid 或 device_code 定位设备，若上报版本与 firmware_version 不一致则写入 ota_target_firmware_version，并更新 firmware_version 为上报值。
func (c *CardDB) UpdateDeviceStoreReportedVersion(ctx context.Context, deviceKey, reportedVersion string) error {
	if deviceKey == "" || reportedVersion == "" {
		return nil
	}
	var currentFirmware sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT firmware_version FROM device_store WHERE device_uid = $1 OR device_code = $1 LIMIT 1
	`, deviceKey, deviceKey).Scan(&currentFirmware)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("query device_store: %w", err)
	}
	current := ""
	if currentFirmware.Valid {
		current = currentFirmware.String
	}
	if current == reportedVersion {
		return nil
	}
	_, err = c.db.ExecContext(ctx, `
		UPDATE device_store SET firmware_version = $1, ota_target_firmware_version = $1
		WHERE device_uid = $2 OR device_code = $2
	`, reportedVersion, reportedVersion, deviceKey)
	if err != nil {
		return fmt.Errorf("update device_store version: %w", err)
	}
	return nil
}
