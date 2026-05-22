package card

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"

	"github.com/lib/pq"
)

// CardDB wraps *sql.DB for card-related queries.
// Phase 2 一刀切：device_id UUID 退役。dfm PK = device_uid VARCHAR(50)；devices PK = device_addr INET /128。
//
//	device_factory_meta (device_uid PK VARCHAR(50), device_code, device_type, device_model, firmware_version, ...)
//	devices             (device_addr PK INET /128, device_uid UNIQUE FK→dfm, card_id, monitoring_enabled, access)
//	cards               (spatial_prefix PK INET, card_name, card_dns, resident_id INET, has_bed/has_bathroom/has_kitchen)
//
// 反查链：device_uid|device_code → device_factory_meta → device_uid
//
//	→ devices.device_addr (LPM 路由起点)
//	→ cards.card_id >>= device_addr  (LPM)
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
// device_addr (/128 INET) → cards.card_id LPM → card_id (spatial_prefix text)。
//
// 命中 → 返回 card_id 字符串；未命中（unbound device，R-009）→ ("", sql.ErrNoRows)。
func LookupCardByDeviceAddr(ctx context.Context, db *sql.DB, addr netip.Addr) (string, error) {
	if !addr.IsValid() {
		return "", fmt.Errorf("device addr invalid")
	}
	var cardID sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT card_id::text
		  FROM cards
		 WHERE card_id >>= $1::INET
		 ORDER BY masklen(card_id) DESC
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
// v2: tenant_id 派生自 devices.device_addr 的 /48 prefix；card_id 走 LPM 反查。
func (c *CardDB) ListSleepadBaselinesForHealth(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("card db nil")
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT
		    host(set_masklen(d.device_addr, 48))::text AS tenant_id,
		    dfm.device_uid,
		    COALESCE(d.access, false),
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    COALESCE(find_card_by_device_addr(d.device_addr)::text, ''),
		    host(d.device_addr)::text AS device_addr
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_uid = dfm.device_uid
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
		var addrStr sql.NullString
		if err := rows.Scan(&b.TenantID, &b.DeviceUID,
			&b.Access, &b.MonitoringEnabled, &b.DeviceCode, &b.DeviceType, &b.CardID, &addrStr); err != nil {
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
// v2: 反查走 device_addr LPM，从 spatial_prefix 派生空间字段（INET CIDR 形式）。
// deviceKey 可以是 device_uid 或 device_code；未绑 device_addr 时返回 ErrNoRows。
func (c *CardDB) LookupCard(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		Access:            false,
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
		    host(set_masklen(d.device_addr, 48))::text AS tenant_id,
		    host(set_masklen(d.device_addr, 56))::text AS branch_id,
		    host(set_masklen(d.device_addr, 80))::text AS unit_id,
		    COALESCE(find_card_by_device_addr(d.device_addr)::text, '') AS card_id,
		    COALESCE(d.access, false),
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    host(set_masklen(d.device_addr, 96))::text AS bed_id,
		    host(set_masklen(d.device_addr, 88))::text AS room_id,
		    host(d.device_addr)::text AS device_addr
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		JOIN devices d ON d.device_uid = dfm.device_uid
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID, &m.BranchID, &m.UnitID, &m.CardID,
		&m.Access, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType,
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

// DeviceUIDsInTenantUnit 返回 unitPrefix（/80 INET CIDR）下 devices 中出现的 device_uid（去重）。
// v2: 用 devices.device_addr <<= $unitPrefix 反查；tenantID 参数保留为兼容签名，但实际寻址用 unitPrefix。
func (c *CardDB) DeviceUIDsInTenantUnit(ctx context.Context, tenantID, unitPrefix string) (deviceUIDs []string) {
	if c == nil || c.db == nil || unitPrefix == "" {
		return nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT d.device_uid
		FROM devices d
		WHERE d.device_addr <<= $1::inet
	`, unitPrefix)
	if err != nil {
		return nil
	}
	defer rows.Close()
	seen := make(map[string]struct{})
	for rows.Next() {
		var uid sql.NullString
		if err := rows.Scan(&uid); err != nil {
			continue
		}
		if uid.Valid && uid.String != "" {
			if _, ok := seen[uid.String]; !ok {
				seen[uid.String] = struct{}{}
				deviceUIDs = append(deviceUIDs, uid.String)
			}
		}
	}
	return deviceUIDs
}

// DeviceUIDsInCard 返回 cardID 对应的 spatial_prefix 下的 device_uid（去重）。
// v2: cards.card_id >>= devices.device_addr 反查。
func (c *CardDB) DeviceUIDsInCard(ctx context.Context, cardID string) (deviceUIDs []string) {
	if c == nil || c.db == nil || cardID == "" {
		return nil
	}
	// v2 unified: card_id ≡ spatial_prefix
	rows, err := c.db.QueryContext(ctx, `
		SELECT d.device_uid
		FROM cards c
		JOIN devices d ON d.device_addr <<= c.card_id
		WHERE c.card_id = $1::INET
	`, cardID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	seen := make(map[string]struct{})
	for rows.Next() {
		var uid sql.NullString
		if err := rows.Scan(&uid); err != nil {
			continue
		}
		if uid.Valid && uid.String != "" {
			if _, ok := seen[uid.String]; !ok {
				seen[uid.String] = struct{}{}
				deviceUIDs = append(deviceUIDs, uid.String)
			}
		}
	}
	return deviceUIDs
}

// ResolveDeviceBaseline 统一构建路径：LookupCard → LookupDeviceOnly → LookupDeviceStoreOnly。
// v2: LookupCard 走 LPM 内含 card_id 解析；后两层 fallback 用 DeviceUID 判存在性。
func (c *CardDB) ResolveDeviceBaseline(ctx context.Context, deviceKey string) (DeviceBaseline, bool) {
	if c == nil || c.db == nil || deviceKey == "" {
		return DeviceBaseline{}, false
	}
	if m, err := c.LookupCard(ctx, deviceKey); err == nil && m != nil && m.CardID != "" {
		return *m, true
	}
	if m, err := c.LookupDeviceOnly(ctx, deviceKey); err == nil && m != nil && m.DeviceUID != "" {
		return *m, true
	}
	if m, err := c.LookupDeviceStoreOnly(ctx, deviceKey); err == nil && m != nil && m.DeviceUID != "" {
		return *m, true
	}
	return DeviceBaseline{}, false
}

// LookupDeviceOnly resolves device key when no card LPM hit; CardID 留空（R-009 不再 UUID 占位）。
func (c *CardDB) LookupDeviceOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		Access:            false,
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
		    host(set_masklen(d.device_addr, 48))::text AS tenant_id,
		    COALESCE(d.access, false),
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    host(set_masklen(d.device_addr, 88))::text AS room_id,
		    host(d.device_addr)::text AS device_addr
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		JOIN devices d ON d.device_uid = dfm.device_uid
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID,
		&m.Access, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType, &m.RoomID, &addrStr)
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
	return &m, nil
}

// LookupDeviceStoreOnly 仅 device_factory_meta 有记录但无 devices 行（未分配 device_addr）时使用。
func (c *CardDB) LookupDeviceStoreOnly(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	m := DeviceBaseline{
		Access:            false,
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
		    false AS access,
		    false AS monitoring_enabled,
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, '')
		FROM resolved r
		JOIN device_factory_meta dfm
		  ON `+pgUIDNormExpr("dfm.device_uid")+` = `+pgUIDNormExpr("r.uid")+`
		LIMIT 1
	`, deviceKey).Scan(&m.DeviceUID, &m.TenantID,
		&m.Access, &m.MonitoringEnabled, &m.DeviceCode, &m.DeviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not in device_factory_meta: %s", deviceKey)
		}
		return nil, fmt.Errorf("lookup device_factory_meta only: %w", err)
	}
	return &m, nil
}

// ListAllBaselines returns DeviceBaseline for devices of given types (or all if deviceTypes is empty).
// v2: 所有空间字段从 device_addr 派生（INET CIDR 形式）；card_id 走 find_card_by_device_addr。
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
		    host(set_masklen(d.device_addr, 48))::text AS tenant_id,
		    COALESCE(dfm.device_code, ''),
		    COALESCE(dfm.device_type::text, ''),
		    COALESCE(d.access, false),
		    COALESCE(d.monitoring_enabled, false),
		    COALESCE(find_card_by_device_addr(d.device_addr)::text, ''),
		    host(set_masklen(d.device_addr, 56))::text AS branch_id,
		    host(set_masklen(d.device_addr, 80))::text AS unit_id,
		    host(set_masklen(d.device_addr, 88))::text AS room_id,
		    host(set_masklen(d.device_addr, 96))::text AS bed_id,
		    host(d.device_addr)::text AS device_addr
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_uid = dfm.device_uid
		`+whereClause, args...)
	if err != nil {
		return nil, fmt.Errorf("list all baselines: %w", err)
	}
	defer rows.Close()
	var out []DeviceBaseline
	for rows.Next() {
		var b DeviceBaseline
		var addrStr sql.NullString
		if err := rows.Scan(&b.DeviceUID, &b.TenantID,
			&b.DeviceCode, &b.DeviceType,
			&b.Access, &b.MonitoringEnabled,
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
// v2: 从 cards.card_id 出发，列出 prefix 内所有 device_addr 的 /88 派生 ID 去重；
// 实际 room name 由 caller 在自己的 spatial 视图层 join 解析。本函数仅返回 RoomID 占位。
func (c *CardDB) RoomIdentifiersForCard(ctx context.Context, tenantID, cardID string) ([]RoomIdentifier, error) {
	if c == nil || c.db == nil || cardID == "" {
		return nil, nil
	}
	// v2 unified: card_id ≡ spatial_prefix
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT host(set_masklen(d.device_addr, 88))::text AS room_id
		FROM cards c
		JOIN devices d ON d.device_addr <<= c.card_id
		WHERE c.card_id = $1::INET
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

// UpdateDeviceStoreReportedVersion 设备上报固件版本时：更新 device_factory_meta.firmware_version。
// 返回 changed=true 表示版本真的写入了；false 表示与当前一致 no-op（caller 据此决定是否 Info-log）。
func (c *CardDB) UpdateDeviceStoreReportedVersion(ctx context.Context, deviceKey, reportedVersion string) (changed bool, err error) {
	if deviceKey == "" || reportedVersion == "" {
		return false, nil
	}
	var currentFirmware sql.NullString
	err = c.db.QueryRowContext(ctx, `
		SELECT dfm.firmware_version
		FROM device_factory_meta dfm
		WHERE `+pgUIDNormExpr("dfm.device_uid")+` = regexp_replace(upper(btrim($1::text)), E'[:.\\s-]', '', 'g')
		   OR dfm.device_code = $1
		LIMIT 1
	`, deviceKey).Scan(&currentFirmware)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("query device_factory_meta: %w", err)
	}
	current := ""
	if currentFirmware.Valid {
		current = currentFirmware.String
	}
	if current == reportedVersion {
		return false, nil
	}
	_, err = c.db.ExecContext(ctx, `
		UPDATE device_factory_meta
		   SET firmware_version = $1
		 WHERE `+pgUIDNormExpr("device_uid")+` = regexp_replace(upper(btrim($2::text)), E'[:.\\s-]', '', 'g')
		    OR device_code = $2
	`, reportedVersion, deviceKey)
	if err != nil {
		return false, fmt.Errorf("update dfm firmware_version: %w", err)
	}
	return true, nil
}
