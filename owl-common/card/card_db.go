package card

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// BusinessAccessDefault 无 devices 行、JOIN 未命中或查库失败时的默认业务访问；pending 与新建 devices 默认一致，非 approved 仍不放行业务数据。
const BusinessAccessDefault = "pending"

// CardDB wraps *sql.DB for card-related queries.
// All gateways share this; callers never import "database/sql" directly.
type CardDB struct {
	db *sql.DB
}

func NewCardDB(db *sql.DB) *CardDB {
	return &CardDB{db: db}
}

// pgUIDNormExpr PostgreSQL：与 NormalizeDeviceLookupKey 对齐的 UID/MAC 比较（大小写与分隔符不敏感）。
func pgUIDNormExpr(col string) string {
	return fmt.Sprintf(`regexp_replace(upper(btrim(COALESCE(%s, ''))), E'[:.\\s-]', '', 'g')`, col)
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

// ListSleepadBaselinesForHealth 列出 device_store 中已配置 device_code 的 Sleepad 类设备，联合 devices 取策略字段，供定时在线探测（与 MQTT 无关）。
func (c *CardDB) ListSleepadBaselinesForHealth(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT ds.tenant_id::text,
		       COALESCE(dev.device_id::text, ''),
		       ds.device_uid,
		       COALESCE(ds.allow_access, false),
		       COALESCE(dev.business_access, 'pending'),
		       COALESCE(dev.monitoring_enabled, false),
		       COALESCE(ds.device_code, ''),
		       COALESCE(ds.device_type, ''),
		       COALESCE(cmap.card_id, '')
		FROM device_store ds
		LEFT JOIN devices dev ON dev.device_uid = ds.device_uid
		LEFT JOIN (
		    SELECT c2.card_id::text AS card_id, `+pgUIDNormExpr("d2.device_uid")+` AS norm_uid
		    FROM cards c2, jsonb_to_recordset(c2.devices) AS d2(device_uid text)
		) cmap ON `+pgUIDNormExpr("ds.device_uid")+` = cmap.norm_uid
		WHERE ds.device_type = ANY($1::text[])
		  AND ds.device_code IS NOT NULL AND BTRIM(ds.device_code) <> ''
	`, pq.Array(deviceTypes))
	if err != nil {
		return nil, fmt.Errorf("list sleepad baselines: %w", err)
	}
	defer rows.Close()
	var out []DeviceBaseline
	for rows.Next() {
		var b DeviceBaseline
		if err := rows.Scan(&b.TenantID, &b.DeviceID, &b.DeviceUID,
			&b.AllowAccess, &b.BusinessAccess, &b.MonitoringEnabled, &b.DeviceCode, &b.DeviceType, &b.CardID); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ResolveDevice resolves a single device_uid or device_code → (device_uid, device_code).
func (c *CardDB) ResolveDevice(ctx context.Context, deviceKey string) (deviceUID, deviceCode string, err error) {
	err = c.db.QueryRowContext(ctx, `
		SELECT device_uid, device_code
		FROM device_store
		WHERE `+pgUIDNormExpr("device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g')
		   OR device_code = $1
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

// LookupCard resolves a device key → DeviceBaseline (device_store + cards + devices).
// deviceKey can be device_uid ("BM87...") or device_code ("1ua3erivl9pv1"); 入参用 code 时 WITH 子句先解析为 uid。
func (c *CardDB) LookupCard(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
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
		       COALESCE(ds.allow_access, false),
		       COALESCE(dev.business_access, 'pending'),
		       COALESCE(dev.monitoring_enabled, false),
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, ''),
		       COALESCE(c.bed_id::text, ''),
		       COALESCE(bd.room_id::text, dev.bound_room_id::text, '')
		FROM resolved r
		LEFT JOIN device_store ds ON `+pgUIDNormExpr("ds.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		, cards c
		LEFT JOIN beds bd ON bd.bed_id = c.bed_id
		, jsonb_to_recordset(c.devices) AS d(device_uid text, device_id text)
		LEFT JOIN devices dev ON dev.device_id = NULLIF(TRIM(d.device_id), '')::uuid
		WHERE `+pgUIDNormExpr("d.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.BranchID, &m.UnitID, &m.CardID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType,
		&m.BedID, &m.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in cards: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup card: %w", err)
	}
	return &m, nil
}

// DeviceKeysInTenantUnit 返回 tenant+unit 下 cards.devices 里出现的 device_id / device_uid（去重），供按 unit 失效缓存。
func (c *CardDB) DeviceKeysInTenantUnit(ctx context.Context, tenantID, unitID string) (deviceIDs, deviceUIDs []string) {
	if c == nil || c.db == nil || tenantID == "" || unitID == "" {
		return nil, nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT j->>'device_id', j->>'device_uid'
		FROM cards c, jsonb_array_elements(COALESCE(c.devices, '[]'::jsonb)) AS j
		WHERE c.tenant_id = $1::uuid AND c.unit_id = $2::uuid
		  AND COALESCE(j->>'device_id', '') <> ''`,
		tenantID, unitID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	seenD := make(map[string]struct{})
	seenU := make(map[string]struct{})
	for rows.Next() {
		var did, uid sql.NullString
		if err := rows.Scan(&did, &uid); err != nil {
			continue
		}
		if did.Valid && did.String != "" {
			if _, ok := seenD[did.String]; !ok {
				seenD[did.String] = struct{}{}
				deviceIDs = append(deviceIDs, did.String)
			}
		}
		if uid.Valid && uid.String != "" {
			if _, ok := seenU[uid.String]; !ok {
				seenU[uid.String] = struct{}{}
				deviceUIDs = append(deviceUIDs, uid.String)
			}
		}
	}
	return deviceIDs, deviceUIDs
}

// DeviceKeysInCard 返回指定 card_id 下 cards.devices 中出现的 device_id / device_uid（去重）。
func (c *CardDB) DeviceKeysInCard(ctx context.Context, cardID string) (deviceIDs, deviceUIDs []string) {
	if c == nil || c.db == nil || cardID == "" {
		return nil, nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT j->>'device_id', j->>'device_uid'
		FROM cards c, jsonb_array_elements(COALESCE(c.devices, '[]'::jsonb)) AS j
		WHERE c.card_id = $1::uuid
		  AND (COALESCE(j->>'device_id', '') <> '' OR COALESCE(j->>'device_uid', '') <> '')`,
		cardID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	seenD := make(map[string]struct{})
	seenU := make(map[string]struct{})
	for rows.Next() {
		var did, uid sql.NullString
		if err := rows.Scan(&did, &uid); err != nil {
			continue
		}
		if did.Valid && did.String != "" {
			if _, ok := seenD[did.String]; !ok {
				seenD[did.String] = struct{}{}
				deviceIDs = append(deviceIDs, did.String)
			}
		}
		if uid.Valid && uid.String != "" {
			if _, ok := seenU[uid.String]; !ok {
				seenU[uid.String] = struct{}{}
				deviceUIDs = append(deviceUIDs, uid.String)
			}
		}
	}
	return deviceIDs, deviceUIDs
}

// lookupMappingByDeviceInCardDevices 按 devices JSON 中 device_uid 或 device_id 命中卡（LookupCard 仅按解析后的 uid 匹配，此处补 device_id UUID 入参）。
func (c *CardDB) lookupMappingByDeviceInCardDevices(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	if c == nil || c.db == nil || deviceKey == "" {
		return nil, sql.ErrNoRows
	}
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	// RoomID 取值优先级（兜底链）：
	//   1. beds.room_id (via c.bed_id)  — ActiveBedCard 走床→房路径
	//   2. devices.bound_room_id        — UnitCard / DeviceCard 走 device 自身绑定（公共设备等无床）
	// 任何 device_type 都能可靠取到 RoomID，让 envelope.SemanticLocation 100% 填充。
	err := c.db.QueryRowContext(ctx, `
		SELECT d.device_uid, c.tenant_id, c.branch_id,
		       COALESCE(c.unit_id::text, ''), c.card_id::text,
		       COALESCE(NULLIF(TRIM(d.device_id), ''), ''),
		       COALESCE(ds.allow_access, false),
		       COALESCE(dev.business_access, 'pending'),
		       COALESCE(dev.monitoring_enabled, false),
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, ''),
		       COALESCE(c.bed_id::text, ''),
		       COALESCE(bd.room_id::text, dev.bound_room_id::text, '')
		FROM cards c
		LEFT JOIN beds bd ON bd.bed_id = c.bed_id,
		     jsonb_to_recordset(c.devices) AS d(device_uid text, device_id text)
		LEFT JOIN device_store ds ON `+pgUIDNormExpr("ds.device_uid")+` = `+pgUIDNormExpr("d.device_uid")+`
		   OR (NULLIF(TRIM(d.device_id), '') IS NOT NULL AND ds.device_id::text = NULLIF(TRIM(d.device_id), ''))
		LEFT JOIN devices dev ON dev.device_id = NULLIF(TRIM(d.device_id), '')::uuid
		WHERE (d.device_uid IS NOT NULL AND `+pgUIDNormExpr("d.device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g'))
		   OR (NULLIF(TRIM(d.device_id), '') IS NOT NULL AND NULLIF(TRIM(d.device_id), '') = $1)
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.BranchID, &m.UnitID, &m.CardID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType,
		&m.BedID, &m.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("lookup card by devices element: %w", err)
	}
	return &m, nil
}

// ResolveDeviceBaseline 统一构建路径：LookupCard → devices JSON 按 id/uid → LookupDeviceOnly → LookupDeviceStoreOnly；与网关/聚合侧缓存失效后的「再查库」一致。
func (c *CardDB) ResolveDeviceBaseline(ctx context.Context, deviceKey string) (DeviceBaseline, bool) {
	if c == nil || c.db == nil || deviceKey == "" {
		return DeviceBaseline{}, false
	}
	if m, err := c.LookupCard(ctx, deviceKey); err == nil && m != nil && m.CardID != "" {
		return *m, true
	}
	if m, err := c.lookupMappingByDeviceInCardDevices(ctx, deviceKey); err == nil && m != nil && m.CardID != "" {
		return *m, true
	}
	if m, err := c.LookupDeviceOnly(ctx, deviceKey); err == nil && m != nil && m.DeviceID != "" {
		return *m, true
	}
	if m, err := c.LookupDeviceStoreOnly(ctx, deviceKey); err == nil && m != nil && m.DeviceID != "" {
		return *m, true
	}
	return DeviceBaseline{}, false
}

// LookupDeviceOnly resolves device key when no card exists (devices + device_store).
// Returns DeviceBaseline with CardID = DeviceID (virtual card).
func (c *CardDB) LookupDeviceOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_store WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT d.device_uid, d.tenant_id, d.device_id,
		       COALESCE(ds.allow_access, false),
		       d.business_access,
		       d.monitoring_enabled,
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, ''),
		       COALESCE(d.bound_room_id::text, '')
		FROM resolved r
		JOIN devices d ON `+pgUIDNormExpr("d.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LEFT JOIN device_store ds ON `+pgUIDNormExpr("ds.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType, &m.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in devices table: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device: %w", err)
	}
	m.CardID = m.DeviceID
	return &m, nil
}

// LookupDeviceStoreOnly 仅 device_store 有记录时使用：tenant_id / device_id 等均来自 device_store（未绑卡、无 devices 行时 LookupCard/LookupDeviceOnly 会失败）。
func (c *CardDB) LookupDeviceStoreOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_store WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT ds.device_uid, ds.tenant_id::text,
		       COALESCE(ds.device_id::text, ''),
		       COALESCE(ds.allow_access, false),
		       'pending'::text,
		       false,
		       COALESCE(ds.device_code, ''), COALESCE(ds.device_type, '')
		FROM resolved r
		JOIN device_store ds ON `+pgUIDNormExpr("ds.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in device_store: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device_store only: %w", err)
	}
	m.CardID = m.DeviceID
	return &m, nil
}

// ListAllBaselines returns DeviceBaseline for devices of given types (or all if deviceTypes is empty).
func (c *CardDB) ListAllBaselines(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	whereClause := ""
	var args []interface{}
	if len(deviceTypes) > 0 {
		whereClause = "WHERE ds.device_type = ANY($1::text[])"
		args = append(args, pq.Array(deviceTypes))
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT ds.device_uid,
		       ds.tenant_id::text,
		       COALESCE(ds.device_id::text, ''),
		       COALESCE(ds.device_code, ''),
		       COALESCE(ds.device_type, ''),
		       COALESCE(ds.allow_access, false),
		       COALESCE(dev.business_access, 'pending'),
		       COALESCE(dev.monitoring_enabled, false),
		       COALESCE(cmap.card_id, ''),
		       COALESCE(cmap.branch_id, ''),
		       COALESCE(cmap.unit_id, ''),
		       COALESCE(cmap.room_id, ''),
		       COALESCE(cmap.bed_id, '')
		FROM device_store ds
		LEFT JOIN devices dev ON dev.device_uid = ds.device_uid
		LEFT JOIN (
		    SELECT c2.card_id::text AS card_id,
		           c2.branch_id::text AS branch_id,
		           COALESCE(c2.unit_id::text, '') AS unit_id,
		           COALESCE(d2.bed_id, '') AS bed_id,
		           COALESCE(bd.room_id::text, '') AS room_id,
		           `+pgUIDNormExpr("d2.device_uid")+` AS norm_uid
		    FROM cards c2,
		         jsonb_to_recordset(c2.devices) AS d2(device_uid text, bed_id text)
		    LEFT JOIN beds bd ON bd.bed_id::text = NULLIF(TRIM(d2.bed_id), '')::uuid::text
		) cmap ON `+pgUIDNormExpr("ds.device_uid")+` = cmap.norm_uid
		`+whereClause+`
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list all baselines: %w", err)
	}
	defer rows.Close()
	var out []DeviceBaseline
	for rows.Next() {
		var b DeviceBaseline
		if err := rows.Scan(&b.DeviceUID, &b.TenantID, &b.DeviceID,
			&b.DeviceCode, &b.DeviceType,
			&b.AllowAccess, &b.BusinessAccess, &b.MonitoringEnabled,
			&b.CardID, &b.BranchID, &b.UnitID, &b.RoomID, &b.BedID); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
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

// RoomIdentifiersForCard 返回与卡片相关的全部房间：卡主床所在房（bed_id→beds→rooms）与 cards.devices JSONB 中各设备 BoundRoomID 的并集，名称来自 rooms 表。
func (c *CardDB) RoomIdentifiersForCard(ctx context.Context, tenantID, cardID string) ([]RoomIdentifier, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	if tenantID == "" || cardID == "" {
		return nil, nil
	}
	var bedRoomID sql.NullString
	var devicesJSON []byte
	err := c.db.QueryRowContext(ctx, `
		SELECT COALESCE(room.room_id::text, ''), c.devices
		FROM cards c
		LEFT JOIN beds bed ON c.bed_id = bed.bed_id AND c.tenant_id = bed.tenant_id
		LEFT JOIN rooms room ON bed.room_id = room.room_id AND c.tenant_id = room.tenant_id
		WHERE c.tenant_id = $1 AND c.card_id = $2
	`, tenantID, cardID).Scan(&bedRoomID, &devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("room identifiers for card: %w", err)
	}

	seen := make(map[string]struct{})
	var ids []string
	if bedRoomID.Valid && bedRoomID.String != "" {
		seen[bedRoomID.String] = struct{}{}
		ids = append(ids, bedRoomID.String)
	}
	devices, err := ParseDevicesFromCardsJSONB(devicesJSON)
	if err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}
	for _, d := range devices {
		if d.BoundRoomID == nil || *d.BoundRoomID == "" {
			continue
		}
		rid := *d.BoundRoomID
		if _, ok := seen[rid]; ok {
			continue
		}
		seen[rid] = struct{}{}
		ids = append(ids, rid)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT room_id::text, room_name FROM rooms
		WHERE tenant_id = $1 AND room_id = ANY($2::uuid[])`,
		tenantID, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("rooms by ids: %w", err)
	}
	defer rows.Close()
	var out []RoomIdentifier
	for rows.Next() {
		var rid, rname string
		if err := rows.Scan(&rid, &rname); err != nil {
			continue
		}
		out = append(out, RoomIdentifier{RoomID: rid, RoomName: rname})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDeviceStoreReportedVersion 首次上连或设备上报版本时：用 device_uid 或 device_code 定位设备，若上报版本与 firmware_version 不一致则写入 ota_target_firmware_version，并更新 firmware_version 为上报值。
func (c *CardDB) UpdateDeviceStoreReportedVersion(ctx context.Context, deviceKey, reportedVersion string) error {
	if deviceKey == "" || reportedVersion == "" {
		return nil
	}
	var currentFirmware sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT firmware_version FROM device_store WHERE `+pgUIDNormExpr("device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g') OR device_code = $1 LIMIT 1
	`, deviceKey).Scan(&currentFirmware)
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
		WHERE `+pgUIDNormExpr("device_uid")+` = regexp_replace(upper(btrim($2::text)), E'[:.\\s-]', '', 'g') OR device_code = $2
	`, reportedVersion, deviceKey)
	if err != nil {
		return fmt.Errorf("update device_store version: %w", err)
	}
	return nil
}
