package card

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"

	"github.com/lib/pq"
)

// BusinessAccessDefault 无 device runtime / 查库失败时的默认业务访问；非 approved 仍不放行业务数据。
const BusinessAccessDefault = "pending"

// CardDB wraps *sql.DB for card-related queries.
// v2: 反查不再依赖 cards.devices JSONB；改走 devices.device_ipv6 + cards.spatial_prefix INET LPM。
//
//	device_factory_meta (device_id PK, device_uid UNIQUE, device_code, device_type, device_model, ...)
//	devices             (device_ipv6 PK INET /128, device_id, monitoring_enabled, access)
//	device_runtime_state(device_id PK, online, firmware_version, last_heartbeat_at, ...)
//	cards               (card_id UUID PK, spatial_prefix INET, card_type, resident_id INET, dns_short_name, ...)
//
// 反查链：device_uid|device_code → device_factory_meta → device_id
//
//	→ devices.device_ipv6 (LPM 路由起点)
//	→ cards.spatial_prefix >>= device_ipv6  (LPM)
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

// LookupCardByDeviceAddr — 内部消费链 v2 唯一公共反查 API：
// device_addr (/128 INET) → cards.spatial_prefix LPM → card_id (UUID PK)。
//
// 命中 → 返回 card_id 字符串；未命中（unbound device，R-009）→ ("", sql.ErrNoRows)。
//
// 单 caller 用例：cardagg IotPreparedHandler / sensor consumer / wisefido-data alarm path。
// 不再走 device_id UUID 路径（device_ipv6 单程票 R-001）。
func LookupCardByDeviceAddr(ctx context.Context, db *sql.DB, addr netip.Addr) (string, error) {
	if !addr.IsValid() {
		return "", fmt.Errorf("device addr invalid")
	}
	var cardID sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT card_id::text
		  FROM cards
		 WHERE spatial_prefix >>= $1::INET
		 ORDER BY masklen(spatial_prefix) DESC
		 LIMIT 1
	`, addr.String()).Scan(&cardID)
	if err != nil {
		return "", err
	}
	if !cardID.Valid || cardID.String == "" {
		return "", sql.ErrNoRows
	}
	return cardID.String, nil
}

// LoadCodeToUIDMap batch loads device_code → device_uid from device_factory_meta.
func (c *CardDB) LoadCodeToUIDMap(ctx context.Context, deviceTypes []string) (map[string]string, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT device_code, device_uid
		FROM device_factory_meta
		WHERE device_code IS NOT NULL AND device_code != ''
		  AND device_uid  IS NOT NULL AND device_uid  != ''
		  AND device_type::text = ANY($1::text[])
	`, pq.Array(deviceTypes))
	if err != nil {
		return nil, fmt.Errorf("query device_factory_meta: %w", err)
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

// ListSleepadBaselinesForHealth 列出已注册的 Sleepad-类设备，联合 devices 取策略字段，供定时在线探测。
// v2: tenant_id 派生自 devices.device_ipv6 的 /48 prefix；card_id 走 LPM 反查。
func (c *CardDB) ListSleepadBaselinesForHealth(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT
		    host(set_masklen(d.device_ipv6, 48))::text AS tenant_id,
		    dfm.device_id::text,
		    dfm.device_uid,
		    COALESCE(d.access, false),
		    CASE WHEN COALESCE(d.access, false) THEN 'approved' ELSE 'pending' END AS business_access,
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    COALESCE(find_card_by_device_addr(d.device_ipv6)::text, '')
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_type::text = ANY($1::text[])
		  AND dfm.device_code IS NOT NULL AND BTRIM(dfm.device_code) <> ''
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
		SELECT device_uid, COALESCE(device_code, '')
		FROM device_factory_meta
		WHERE `+pgUIDNormExpr("device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g')
		   OR device_code = $1
		LIMIT 1
	`, deviceKey).Scan(&deviceUID, &deviceCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("device not in device_factory_meta: %s", deviceKey)
		}
		return "", "", fmt.Errorf("resolve device: %w", err)
	}
	return
}

// LookupCard resolves a device key → DeviceBaseline (device_factory_meta + devices + cards LPM).
// v2: 反查走 device_ipv6 LPM，从 spatial_prefix 派生空间字段（INET CIDR 形式）。
// deviceKey 可以是 device_uid 或 device_code；未绑 device_ipv6 时返回 ErrNoRows。
func (c *CardDB) LookupCard(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	var addrStr sql.NullString
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_factory_meta WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT
		    dfm.device_uid,
		    host(set_masklen(d.device_ipv6, 48))::text AS tenant_id,
		    host(set_masklen(d.device_ipv6, 56))::text AS branch_id,
		    host(set_masklen(d.device_ipv6, 80))::text AS unit_id,
		    COALESCE(find_card_by_device_addr(d.device_ipv6)::text, '') AS card_id,
		    dfm.device_id::text,
		    COALESCE(d.access, false),
		    CASE WHEN COALESCE(d.access, false) THEN 'approved' ELSE 'pending' END AS business_access,
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    host(set_masklen(d.device_ipv6, 96))::text AS bed_id,
		    host(set_masklen(d.device_ipv6, 88))::text AS room_id,
		    host(d.device_ipv6)::text AS device_addr
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		JOIN devices d ON d.device_id = dfm.device_id
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.BranchID, &m.UnitID, &m.CardID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType,
		&m.BedID, &m.RoomID, &addrStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in cards: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup card: %w", err)
	}
	if addrStr.Valid && addrStr.String != "" {
		if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
			m.DeviceAddr = a
		}
	}
	return &m, nil
}

// DeviceKeysInTenantUnit 返回 unitPrefix（/80 INET CIDR）下 devices 中出现的 device_id / device_uid（去重）。
// v2: 用 devices.device_ipv6 <<= $unitPrefix 反查；tenantID 参数保留为兼容签名，但实际寻址用 unitPrefix。
func (c *CardDB) DeviceKeysInTenantUnit(ctx context.Context, tenantID, unitPrefix string) (deviceIDs, deviceUIDs []string) {
	if c == nil || c.db == nil || unitPrefix == "" {
		return nil, nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT dfm.device_id::text, dfm.device_uid
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::inet
	`, unitPrefix)
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

// DeviceKeysInCard 返回 cardID 对应的 spatial_prefix 下的 device_id / device_uid（去重）。
// v2: cards.spatial_prefix >>= devices.device_ipv6 反查。
func (c *CardDB) DeviceKeysInCard(ctx context.Context, cardID string) (deviceIDs, deviceUIDs []string) {
	if c == nil || c.db == nil || cardID == "" {
		return nil, nil
	}
	// v2 unified: card_id ≡ spatial_prefix
	rows, err := c.db.QueryContext(ctx, `
		SELECT dfm.device_id::text, dfm.device_uid
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE c.spatial_prefix = $1::INET
	`, cardID)
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

// ResolveDeviceBaseline 统一构建路径：LookupCard → LookupDeviceOnly → LookupDeviceStoreOnly。
// v2: 已删除 lookupMappingByDeviceInCardDevices（无 JSONB 反查路径）；LookupCard 走 LPM 内含 card_id 解析。
func (c *CardDB) ResolveDeviceBaseline(ctx context.Context, deviceKey string) (DeviceBaseline, bool) {
	if c == nil || c.db == nil || deviceKey == "" {
		return DeviceBaseline{}, false
	}
	if m, err := c.LookupCard(ctx, deviceKey); err == nil && m != nil && m.CardID != "" {
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

// LookupDeviceOnly resolves device key when no card LPM hit; CardID 留空（R-009 不再 UUID 占位）。
func (c *CardDB) LookupDeviceOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	var addrStr sql.NullString
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_factory_meta WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT
		    dfm.device_uid,
		    host(set_masklen(d.device_ipv6, 48))::text AS tenant_id,
		    dfm.device_id::text,
		    COALESCE(d.access, false),
		    CASE WHEN COALESCE(d.access, false) THEN 'approved' ELSE 'pending' END AS business_access,
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    host(set_masklen(d.device_ipv6, 88))::text AS room_id,
		    host(d.device_ipv6)::text AS device_addr
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		JOIN devices d ON d.device_id = dfm.device_id
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType, &m.RoomID, &addrStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in devices table: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device: %w", err)
	}
	if addrStr.Valid && addrStr.String != "" {
		if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
			m.DeviceAddr = a
		}
	}
	// R-009: 不再用 DeviceID UUID 充 CardID；unbound device CardID 留空
	return &m, nil
}

// LookupDeviceStoreOnly 仅 device_factory_meta 有记录但无 devices 行（未分配 IPv6）时使用。
func (c *CardDB) LookupDeviceStoreOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		AllowAccess:       false,
		BusinessAccess:    BusinessAccessDefault,
		MonitoringEnabled: false,
	}
	err := c.db.QueryRowContext(ctx, `
		WITH resolved AS (
			SELECT COALESCE(
				(SELECT device_uid FROM device_factory_meta WHERE device_code = $1 LIMIT 1),
				$1
			) AS uid
		)
		SELECT
		    dfm.device_uid,
		    ''::text AS tenant_id,
		    dfm.device_id::text,
		    false AS access,
		    'pending'::text AS business_access,
		    false AS monitoring_enabled,
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, '')
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.DeviceID,
		&m.AllowAccess, &m.BusinessAccess, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in device_factory_meta: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device_factory_meta only: %w", err)
	}
	// R-009: 不再用 DeviceID UUID 充 CardID；unbound device CardID 留空
	return &m, nil
}

// ListAllBaselines returns DeviceBaseline for devices of given types (or all if deviceTypes is empty).
// v2: 所有空间字段从 device_ipv6 派生（INET CIDR 形式）；card_id 走 find_card_by_device_addr。
func (c *CardDB) ListAllBaselines(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	whereClause := ""
	var args []interface{}
	if len(deviceTypes) > 0 {
		whereClause = "WHERE dfm.device_type::text = ANY($1::text[])"
		args = append(args, pq.Array(deviceTypes))
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT
		    dfm.device_uid,
		    host(set_masklen(d.device_ipv6, 48))::text AS tenant_id,
		    dfm.device_id::text,
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    COALESCE(d.access, false),
		    CASE WHEN COALESCE(d.access, false) THEN 'approved' ELSE 'pending' END AS business_access,
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(find_card_by_device_addr(d.device_ipv6)::text, ''),
		    host(set_masklen(d.device_ipv6, 56))::text AS branch_id,
		    host(set_masklen(d.device_ipv6, 80))::text AS unit_id,
		    host(set_masklen(d.device_ipv6, 88))::text AS room_id,
		    host(set_masklen(d.device_ipv6, 96))::text AS bed_id,
		    host(d.device_ipv6)::text AS device_addr
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_id = dfm.device_id
		`+whereClause, args...)
	if err != nil {
		return nil, fmt.Errorf("list all baselines: %w", err)
	}
	defer rows.Close()
	var out []DeviceBaseline
	for rows.Next() {
		var b DeviceBaseline
		var addrStr sql.NullString
		if err := rows.Scan(&b.DeviceUID, &b.TenantID, &b.DeviceID,
			&b.DeviceCode, &b.DeviceType,
			&b.AllowAccess, &b.BusinessAccess, &b.MonitoringEnabled,
			&b.CardID, &b.BranchID, &b.UnitID, &b.RoomID, &b.BedID, &addrStr); err != nil {
			return nil, err
		}
		if addrStr.Valid && addrStr.String != "" {
			if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
				b.DeviceAddr = a
			}
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// RoomIdentifiersForCard 返回 card 关联的全部 room ID（INET /88 CIDR 字符串列表）。
// v2: 从 cards.spatial_prefix 出发，列出 prefix 内所有 device_ipv6 的 /88 派生 ID 去重；
// 实际 room name 由 caller 在自己的 spatial 视图层 join 解析。本函数仅返回 RoomID 占位。
func (c *CardDB) RoomIdentifiersForCard(ctx context.Context, tenantID, cardID string) ([]RoomIdentifier, error) {
	if c == nil || c.db == nil || cardID == "" {
		return nil, nil
	}
	// v2 unified: card_id ≡ spatial_prefix
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT host(set_masklen(d.device_ipv6, 88))::text AS room_id
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		WHERE c.spatial_prefix = $1::INET
	`, cardID)
	if err != nil {
		return nil, fmt.Errorf("room identifiers for card: %w", err)
	}
	defer rows.Close()
	var out []RoomIdentifier
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			continue
		}
		out = append(out, RoomIdentifier{RoomID: rid})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDeviceStoreReportedVersion 设备上报固件版本时：更新 device_runtime_state.firmware_version。
// v2: firmware_version 已从 device_store 迁出到 device_runtime_state；OTA target 字段独立到 device_ota 表。
func (c *CardDB) UpdateDeviceStoreReportedVersion(ctx context.Context, deviceKey, reportedVersion string) error {
	if deviceKey == "" || reportedVersion == "" {
		return nil
	}
	var currentFirmware sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT drs.firmware_version
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		WHERE `+pgUIDNormExpr("dfm.device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g')
		   OR dfm.device_code = $1
		LIMIT 1
	`, deviceKey).Scan(&currentFirmware)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("query device_factory_meta: %w", err)
	}
	current := ""
	if currentFirmware.Valid {
		current = currentFirmware.String
	}
	if current == reportedVersion {
		return nil
	}
	_, err = c.db.ExecContext(ctx, `
		INSERT INTO device_runtime_state (device_id, firmware_version, updated_at)
		SELECT dfm.device_id, $1, NOW()
		FROM device_factory_meta dfm
		WHERE `+pgUIDNormExpr("dfm.device_uid")+` = regexp_replace(upper(btrim($2::text)), E'[:.\\s-]', '', 'g')
		   OR dfm.device_code = $2
		ON CONFLICT (device_id) DO UPDATE SET firmware_version = EXCLUDED.firmware_version, updated_at = NOW()
	`, reportedVersion, deviceKey)
	if err != nil {
		return fmt.Errorf("update device_runtime_state version: %w", err)
	}
	return nil
}
