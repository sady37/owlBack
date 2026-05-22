package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
)

// PostgresDeviceStoreRepository 设备库存Repository实现（v2 schema）。
//
// v2.5 把 v1 的 device_store 表拆为 3 张表（drs 已退役，online/last_seen 走 Redis device:status）：
//   - device_factory_meta — 出厂元数据 + firmware_version（device_uid PK VARCHAR(50)，mac_wifi/mac_ble, ...）
//   - devices — 业务空间绑定（device_addr INET /128 PK, device_uid UNIQUE FK→dfm, card_id INET, monitoring_enabled）
//   - device_ota — OTA 升级计划（device_addr PK FK→devices, target_firmware_version, approve_way enum, schedule, status, ...）
//
// Phase 2 一刀切：device_id UUID 退役；identity 收口到 device_uid，业务寻址走 device_addr。
//
// FE 期望的 v1 字段（OTAPermit/OTAWay/OTASchedule/AllowAccess 等）由本 repo 在读写两端做映射。
type PostgresDeviceStoreRepository struct {
	db *sql.DB
}

// NewPostgresDeviceStoreRepository 创建设备库存Repository
func NewPostgresDeviceStoreRepository(db *sql.DB) *PostgresDeviceStoreRepository {
	return &PostgresDeviceStoreRepository{db: db}
}

// v2 默认未分配 tenant prefix（trash/2 号 slot；FE/import 不指定 TenantID 时落入此池）。
const defaultUnboundTenantPrefix = "fd00:0:2::/48"

// v2 pivot tenants（system + trash）— device 调拨业务规则（v1 R-001 保留）：
// 跨 tenant 转移必须经 pivot 中转（System 是合法跳板，trash 是丢弃入口）。
//   System  = fd00:0:1::/48  (slot=1)
//   Trash   = fd00:0:2::/48  (slot=2)
func isV2PivotTenantPrefix(prefix string) bool {
	p := strings.TrimSpace(prefix)
	return p == "fd00:0:1::/48" || p == "fd00:0:2::/48"
}

// tenantNamesOrIDsByPrefix 拉两个 tenant 的 display name；查不到时回 prefix 字符串
func tenantNamesOrIDsByPrefix(ctx context.Context, tx *sql.Tx, p1, p2 string) (string, string) {
	get := func(p string) string {
		var n sql.NullString
		_ = tx.QueryRowContext(ctx,
			`SELECT tenant_name FROM tenants WHERE tenant_id = $1::INET`, p).Scan(&n)
		if n.Valid && n.String != "" {
			return n.String
		}
		return p
	}
	return get(p1), get(p2)
}

// ----------------------------------------------------------------------------
// 排序白名单 — v2.5 表名映射（dfm/d/t/o；drs 已退役）
// ----------------------------------------------------------------------------

func orderByClauseDeviceStoreV2(sortKey, direction string) string {
	dir := "ASC"
	if strings.TrimSpace(strings.ToUpper(direction)) == "DESC" {
		dir = "DESC"
	}
	col := strings.TrimSpace(strings.ToLower(sortKey))
	switch col {
	case "device_uid", "device_code", "device_type", "device_model", "imei", "comm_mode", "mcu_model", "firmware_version":
		return "dfm." + col + " " + dir
	case "mac":
		return "dfm.mac_wifi " + dir
	case "import_date":
		return "dfm.import_date " + dir
	case "allocate_time":
		return "d.created_at " + dir
	case "tenant_name":
		return "t.tenant_name " + dir
	case "device_addr":
		return "d.device_addr " + dir
	case "ota_target_firmware_version":
		return "o.target_firmware_version " + dir
	case "ota_target_mcu_model":
		return "o.target_mcu_model " + dir
	case "ota_status":
		return "o.status " + dir
	case "ota_progress":
		return "o.progress " + dir
	case "ota_updated_at":
		return "o.updated_at " + dir
	default:
		return "dfm.import_date DESC, dfm.device_type, dfm.device_uid"
	}
}

// ----------------------------------------------------------------------------
// v1 ⇄ v2 OTA 字段映射 helpers
// ----------------------------------------------------------------------------

// v1Schedule 形如 "now+3d" / "MMDDHH+xH" / "" 解析为 timestamptz；空串→NULL。
//
// 解析规则：
//
//	"now+Nd"      → NOW() + N*1day
//	"now+Nh"      → NOW() + N*1hour
//	"MMDDHH+xH"   → 当年下一次 MM-DD HH:00 之后再 +x 小时（若已过则 +1 年）
//	其它格式      → 返回 NULL（视为无 schedule）
var (
	reNowDelta = regexp.MustCompile(`^now\+(\d+)([dh])$`)
	reMMDDHH   = regexp.MustCompile(`^(\d{2})(\d{2})(\d{2})\+(\d+)H$`)
)

func parseV1OTAScheduleToTime(v1 string) sql.NullTime {
	s := strings.TrimSpace(v1)
	if s == "" {
		return sql.NullTime{}
	}
	if m := reNowDelta.FindStringSubmatch(strings.ToLower(s)); m != nil {
		var n int
		fmt.Sscanf(m[1], "%d", &n)
		now := time.Now().UTC()
		switch m[2] {
		case "d":
			return sql.NullTime{Time: now.Add(time.Duration(n) * 24 * time.Hour), Valid: true}
		case "h":
			return sql.NullTime{Time: now.Add(time.Duration(n) * time.Hour), Valid: true}
		}
	}
	if m := reMMDDHH.FindStringSubmatch(s); m != nil {
		var mm, dd, hh, plusH int
		fmt.Sscanf(m[1], "%d", &mm)
		fmt.Sscanf(m[2], "%d", &dd)
		fmt.Sscanf(m[3], "%d", &hh)
		fmt.Sscanf(m[4], "%d", &plusH)
		now := time.Now().UTC()
		t := time.Date(now.Year(), time.Month(mm), dd, hh, 0, 0, 0, time.UTC).Add(time.Duration(plusH) * time.Hour)
		if t.Before(now) {
			t = t.AddDate(1, 0, 0)
		}
		return sql.NullTime{Time: t, Valid: true}
	}
	// 兜底：尝试 RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	return sql.NullTime{}
}

// mapV1OTAWayToV2ApproveWay 把 v1 的 (ota_permit, ota_way) 映射为 v2 approve_way enum。
//   - permit='false' / 空           → NULL（无 OTA 计划）
//   - permit='true' + way='manual'   → "system_manual"   (无 tenant context 时默认 system_*)
//   - permit='true' + way='schedule' → "system_schedule"
//   - permit='true' + way='tenant'   → "tenant_manual"   (历史兼容：v1 'tenant' 视作 tenant_manual)
//
// tenantApproved=true 时强制走 tenant_*；false/未设视为 system_*。
func mapV1OTAWayToV2ApproveWay(otaPermit, otaWay sql.NullString, tenantApproved bool) sql.NullString {
	permit := strings.ToLower(strings.TrimSpace(otaPermit.String))
	way := strings.ToLower(strings.TrimSpace(otaWay.String))
	if !otaPermit.Valid || permit == "" || permit == "false" {
		return sql.NullString{}
	}
	prefix := "system"
	if tenantApproved || way == "tenant" {
		prefix = "tenant"
	}
	suffix := "manual"
	switch way {
	case "schedule":
		suffix = "schedule"
	case "manual", "":
		suffix = "manual"
	case "tenant":
		// v1 'tenant' 历史用法 = tenant_manual
		suffix = "manual"
	}
	return sql.NullString{String: prefix + "_" + suffix, Valid: true}
}

// ----------------------------------------------------------------------------
// 共用 SELECT 列表（List/Get/GetByID 复用），返回 v1 形 ToJSON 兼容字段
// ----------------------------------------------------------------------------

const deviceStoreSelectColumnsV2 = `
  COALESCE(host(d.device_addr), '') AS device_addr,
  dfm.device_uid,
  dfm.device_code,
  dfm.device_type::text,
  dfm.device_model,
  dfm.mac_wifi      AS mac,
  dfm.imei,
  dfm.comm_mode,
  dfm.mcu_model,
  dfm.firmware_version,
  -- OTA 字段：v2 表 device_ota → v1 形
  o.target_firmware_version AS ota_target_firmware_version,
  o.target_mcu_model        AS ota_target_mcu_model,
  CASE WHEN o.approve_way IS NULL THEN 'false' ELSE 'true' END AS ota_permit,
  CASE
    WHEN o.approve_way IS NULL THEN NULL
    WHEN o.approve_way LIKE '%_schedule' THEN 'schedule'
    WHEN o.approve_way LIKE 'tenant_%'   THEN 'tenant'
    WHEN o.approve_way LIKE '%_manual'   THEN 'manual'
    ELSE NULL
  END AS ota_way,
  to_char(o.schedule, 'YYYY-MM-DD HH24:MI:SS') AS ota_schedule,
  o.status   AS ota_status,
  o.progress AS ota_progress,
  o.error    AS ota_error,
  o.updated_at AS ota_updated_at,
  ''::text   AS ota_firmware_url,
  ''::text   AS ota_firmware_sha256,
  0::bigint  AS ota_firmware_size,
  CASE WHEN o.approve_way LIKE 'tenant_%' THEN TRUE ELSE FALSE END AS ota_tenant_approved,
  -- tenant 反推：device_addr /48 (CIDR 字符串)
  COALESCE(host(network(set_masklen(d.device_addr, 48))) || '/48', $TENANT_DEFAULT$) AS tenant_id,
  COALESCE(t.tenant_name, '') AS tenant_name,
  d.created_at AS allocate_time,
  dfm.import_date,
  'offline'::text AS online_status,
  COALESCE(d.access, FALSE) AS allow_access
`

// scanDeviceStoreRowV2 共用 row 扫描（List/Get 都用）。
func scanDeviceStoreRowV2(scan func(...any) error) (*domain.DeviceStore, error) {
	var d domain.DeviceStore
	var deviceCode, deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var otaTargetFW, otaTargetMCU, otaPermit, otaWay, otaSchedule, otaStatus, otaError sql.NullString
	var otaURL, otaSHA sql.NullString
	var otaProgress sql.NullInt32
	var otaSize sql.NullInt64
	var otaUpdatedAt sql.NullTime
	var otaTenantApproved sql.NullBool
	var tenantName sql.NullString
	var allocateTime sql.NullTime
	var importDate sql.NullTime
	var tenantID, onlineStatus string
	var allowAccess bool

	err := scan(
		&d.DeviceAddr,
		&d.DeviceUID,
		&deviceCode,
		&d.DeviceType,
		&deviceModel,
		&mac,
		&imei,
		&commMode,
		&mcuModel,
		&firmwareVersion,
		&otaTargetFW,
		&otaTargetMCU,
		&otaPermit,
		&otaWay,
		&otaSchedule,
		&otaStatus,
		&otaProgress,
		&otaError,
		&otaUpdatedAt,
		&otaURL,
		&otaSHA,
		&otaSize,
		&otaTenantApproved,
		&tenantID,
		&tenantName,
		&allocateTime,
		&importDate,
		&onlineStatus,
		&allowAccess,
	)
	if err != nil {
		return nil, err
	}
	d.DeviceCode = deviceCode
	d.DeviceModel = deviceModel
	d.MAC = mac
	d.IMEI = imei
	d.CommMode = commMode
	d.MCUModel = mcuModel
	d.FirmwareVersion = firmwareVersion
	d.OTATargetFirmwareVersion = otaTargetFW
	d.OTATargetMCUModel = otaTargetMCU
	d.OTAPermit = otaPermit
	d.OTAWay = otaWay
	d.OTASchedule = otaSchedule
	d.OTAStatus = otaStatus
	d.OTAProgress = otaProgress
	d.OTAError = otaError
	d.OTAUpdatedAt = otaUpdatedAt
	d.OTAFirmwareURL = otaURL
	d.OTAFirmwareSHA = otaSHA
	d.OTAFirmwareSize = otaSize
	if otaTenantApproved.Valid {
		d.OTATenantApproved = otaTenantApproved.Bool
	}
	d.TenantID = tenantID
	d.TenantName = tenantName
	d.AllocateTime = allocateTime
	d.ImportDate = importDate
	d.OnlineStatus = onlineStatus
	// v2: devicestore Access toggle = devices.access (platform 审批位，platform_admin 管)
	d.AllowAccess = allowAccess
	return &d, nil
}

// expandSelectColumnsV2 替换占位符 $TENANT_DEFAULT$ 为引号包裹的字符串字面量。
func expandSelectColumnsV2() string {
	return strings.ReplaceAll(deviceStoreSelectColumnsV2, "$TENANT_DEFAULT$", "'"+defaultUnboundTenantPrefix+"'")
}

// ----------------------------------------------------------------------------
// ListDeviceStores
// ----------------------------------------------------------------------------

func (r *PostgresDeviceStoreRepository) ListDeviceStores(ctx context.Context, filters DeviceStoreFilters, page, size int, sort, direction string) ([]*domain.DeviceStore, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	if s := strings.TrimSpace(filters.Search); s != "" {
		where = append(where, fmt.Sprintf("(dfm.device_uid ILIKE $%d OR dfm.device_code ILIKE $%d OR dfm.mac_wifi ILIKE $%d OR dfm.mac_ble ILIKE $%d OR dfm.imei ILIKE $%d)",
			argN, argN, argN, argN, argN))
		args = append(args, "%"+s+"%")
		argN++
	}
	if t := strings.TrimSpace(filters.TenantID); t != "" {
		// v2 tenant 是 INET /48；devices.device_addr 在该范围内即视为绑定到该 tenant
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM devices d2 WHERE d2.device_uid = dfm.device_uid AND d2.device_addr <<= $%d::INET)", argN))
		args = append(args, t)
		argN++
	}
	if dt := strings.TrimSpace(filters.DeviceType); dt != "" {
		where = append(where, fmt.Sprintf("dfm.device_type::text = $%d", argN))
		args = append(args, dt)
		argN++
	}
	if u := strings.TrimSpace(filters.DeviceUID); u != "" {
		where = append(where, fmt.Sprintf("dfm.device_uid ILIKE $%d", argN))
		args = append(args, "%"+u+"%")
		argN++
	}
	if c := strings.TrimSpace(filters.DeviceCode); c != "" {
		where = append(where, fmt.Sprintf("dfm.device_code ILIKE $%d", argN))
		args = append(args, "%"+c+"%")
		argN++
	}
	if fv := strings.TrimSpace(filters.FirmwareVersion); fv != "" {
		where = append(where, fmt.Sprintf("dfm.firmware_version ILIKE $%d", argN))
		args = append(args, "%"+fv+"%")
		argN++
	}
	// v1 OTA 过滤 → v2 WHERE 翻译
	if v := strings.TrimSpace(filters.OTAStatus); v != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := strings.TrimSpace(filters.OTAPermit); v != "" {
		switch strings.ToLower(v) {
		case "true":
			where = append(where, "o.approve_way IS NOT NULL")
		case "false":
			where = append(where, "o.approve_way IS NULL")
		}
	}
	if v := strings.TrimSpace(filters.OTAWay); v != "" {
		switch strings.ToLower(v) {
		case "manual":
			where = append(where, "o.approve_way LIKE '%_manual'")
		case "schedule":
			where = append(where, "o.approve_way LIKE '%_schedule'")
		case "tenant":
			where = append(where, "o.approve_way LIKE 'tenant_%'")
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := `
		SELECT COUNT(*)
		  FROM device_factory_meta dfm
		  LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		  LEFT JOIN device_ota o ON o.device_uid = d.device_uid
		  ` + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count device_factory_meta: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, size, offset)
	limitN := argN
	offsetN := argN + 1

	q := `
		SELECT ` + expandSelectColumnsV2() + `
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		LEFT JOIN device_ota o ON o.device_uid = d.device_uid
		LEFT JOIN tenants t ON t.tenant_id = network(set_masklen(d.device_addr, 48))
		` + whereClause + `
		ORDER BY ` + orderByClauseDeviceStoreV2(sort, direction) + `
		LIMIT $` + fmt.Sprintf("%d", limitN) + ` OFFSET $` + fmt.Sprintf("%d", offsetN)

	rows, err := r.db.QueryContext(ctx, q, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list device_factory_meta: %w", err)
	}
	defer rows.Close()

	out := []*domain.DeviceStore{}
	for rows.Next() {
		ds, err := scanDeviceStoreRowV2(rows.Scan)
		if err != nil {
			return nil, 0, fmt.Errorf("scan device_factory_meta: %w", err)
		}
		out = append(out, ds)
	}
	return out, total, rows.Err()
}

// ----------------------------------------------------------------------------
// GetDeviceStore / GetDeviceStoreByDeviceID
// ----------------------------------------------------------------------------

func (r *PostgresDeviceStoreRepository) GetDeviceStore(ctx context.Context, deviceUID string) (*domain.DeviceStore, error) {
	q := `
		SELECT ` + expandSelectColumnsV2() + `
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		LEFT JOIN device_ota o ON o.device_uid = d.device_uid
		LEFT JOIN tenants t ON t.tenant_id = network(set_masklen(d.device_addr, 48))
		WHERE dfm.device_uid = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, q, deviceUID)
	ds, err := scanDeviceStoreRowV2(row.Scan)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
	}
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// GetDeviceStoreByDeviceAddr Phase 2 一刀切：device_id UUID 退役，业务键改 device_addr。
func (r *PostgresDeviceStoreRepository) GetDeviceStoreByDeviceAddr(ctx context.Context, deviceAddr string) (*domain.DeviceStore, error) {
	q := `
		SELECT ` + expandSelectColumnsV2() + `
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		LEFT JOIN device_ota o ON o.device_uid = d.device_uid
		LEFT JOIN tenants t ON t.tenant_id = network(set_masklen(d.device_addr, 48))
		WHERE d.device_addr = $1::INET
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, q, deviceAddr)
	ds, err := scanDeviceStoreRowV2(row.Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// ----------------------------------------------------------------------------
// CreateDeviceStore — 写入 factory_meta + runtime_state + devices(stateless addr)
// ----------------------------------------------------------------------------

func (r *PostgresDeviceStoreRepository) CreateDeviceStore(ctx context.Context, deviceStore *domain.DeviceStore) (string, error) {
	if deviceStore == nil {
		return "", fmt.Errorf("device_store is required")
	}
	if deviceStore.DeviceType == "" {
		return "", fmt.Errorf("device_type is required")
	}
	if deviceStore.DeviceUID == "" {
		return "", fmt.Errorf("device_uid is required")
	}

	// 1. 重复检查
	var existingUID string
	err := r.db.QueryRowContext(ctx,
		`SELECT device_uid FROM device_factory_meta WHERE device_uid = $1 LIMIT 1`,
		deviceStore.DeviceUID,
	).Scan(&existingUID)
	if err == nil {
		return "", fmt.Errorf("device already exists: device_uid=%s", existingUID)
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check existing device: %w", err)
	}

	// 2. 决定 tenant prefix（未提供 → trash 池）
	tenantPrefix := strings.TrimSpace(deviceStore.TenantID)
	if tenantPrefix == "" {
		tenantPrefix = defaultUnboundTenantPrefix
	}

	// 3. 派生 device_addr：tenant_id + 0:0:0:0:MAC32（最后 32bit 取自 MAC 或 device_uid）
	mac32 := deriveMAC32Suffix(deviceStore.MAC, deviceStore.DeviceUID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 4. INSERT device_factory_meta（PK = device_uid VARCHAR(50)，硬件 identity 不变量）
	insertMeta := `
		INSERT INTO device_factory_meta (
			device_uid, device_code, device_type, device_model,
			mcu_model, mac_wifi, mac_ble, imei, comm_mode, firmware_version
		) VALUES ($1, $2, $3::device_type_enum, $4, $5, $6, $7, $8, $9, $10)
	`
	if _, err := tx.ExecContext(ctx, insertMeta,
		deviceStore.DeviceUID,
		nullStringToAny(deviceStore.DeviceCode),
		deviceStore.DeviceType,
		nullStringToAny(deviceStore.DeviceModel),
		nullStringToAny(deviceStore.MCUModel),
		nullStringToAny(deviceStore.MAC), // mac_wifi
		nil,                              // mac_ble: 不接 v1 字段
		nullStringToAny(deviceStore.IMEI),
		nullStringToAny(deviceStore.CommMode),
		nullStringToAny(deviceStore.FirmwareVersion),
	); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", fmt.Errorf("device already exists (duplicate device_uid)")
		}
		return "", fmt.Errorf("failed to insert device_factory_meta: %w", err)
	}

	// 5. INSERT devices —— device_addr = tenant_id host bytes 全 0 ++ MAC32
	insertDevice := `
		INSERT INTO devices (device_addr, device_uid, monitoring_enabled)
		VALUES (
		  set_masklen(network(set_masklen($1::INET, 48)), 128) | ('::' || $2)::INET,
		  $3, TRUE
		)
		RETURNING host(device_addr)
	`
	var deviceAddr string
	if err := tx.QueryRowContext(ctx, insertDevice, tenantPrefix, mac32, deviceStore.DeviceUID).Scan(&deviceAddr); err != nil {
		return "", fmt.Errorf("failed to insert devices: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return deviceAddr, nil
}

// deriveMAC32Suffix 从 MAC 取低 32bit（"::ffff:ffff" 形式 hex 字符串），不可用则 hash device_uid。
//
// 返回 IPv6 host suffix 字面量片段（不含 "::"），形如 "abcd:1234"（=最后 4 字节）。
func deriveMAC32Suffix(mac sql.NullString, deviceUID string) string {
	hexBytes := func(s string) []byte {
		// 接受 "AA:BB:CC:DD:EE:FF" / "aabbccddeeff" / 任意分隔符
		clean := strings.Map(func(r rune) rune {
			switch {
			case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
				return r
			default:
				return -1
			}
		}, s)
		b, err := hex.DecodeString(clean)
		if err != nil {
			return nil
		}
		return b
	}
	var b []byte
	if mac.Valid {
		b = hexBytes(mac.String)
	}
	if len(b) < 4 {
		// fallback: 用 device_uid 生成可重复的 4 字节
		b = hexBytes(deviceUID)
	}
	if len(b) < 4 {
		// 极端情况：填 0
		b = []byte{0, 0, 0, 0}
	}
	last4 := b[len(b)-4:]
	return fmt.Sprintf("%x%x:%x%x", last4[0], last4[1], last4[2], last4[3])
}

// ----------------------------------------------------------------------------
// BatchUpdateDeviceStores — v2 schema
// ----------------------------------------------------------------------------

func (r *PostgresDeviceStoreRepository) BatchUpdateDeviceStores(ctx context.Context, updates []*domain.DeviceStore) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, update := range updates {
		if update == nil || update.DeviceUID == "" {
			continue
		}
		deviceUID := update.DeviceUID

		// 解出当前 device_addr
		var currentAddr sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT host(d.device_addr) || '/128'
			  FROM device_factory_meta dfm
			  LEFT JOIN devices d ON d.device_uid = dfm.device_uid
			 WHERE dfm.device_uid = $1
		`, deviceUID).Scan(&currentAddr)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
			}
			return fmt.Errorf("failed to query device: %w", err)
		}

		// 1) device_factory_meta 字段（device_type / device_model / device_code / mac / imei / comm_mode / mcu_model）
		factorySet := []string{}
		factoryArgs := []any{}
		fN := 1
		if update.DeviceType != "" {
			factorySet = append(factorySet, fmt.Sprintf("device_type = $%d::device_type_enum", fN))
			factoryArgs = append(factoryArgs, update.DeviceType)
			fN++
		}
		if update.DeviceModel.Valid {
			factorySet = append(factorySet, fmt.Sprintf("device_model = $%d", fN))
			factoryArgs = append(factoryArgs, nullStringToAny(update.DeviceModel))
			fN++
		}
		if update.DeviceCodeSet {
			factorySet = append(factorySet, fmt.Sprintf("device_code = $%d", fN))
			factoryArgs = append(factoryArgs, nullStringToAny(update.DeviceCode))
			fN++
		}
		if len(factorySet) > 0 {
			factoryArgs = append(factoryArgs, deviceUID)
			q := fmt.Sprintf(`UPDATE device_factory_meta SET %s WHERE device_uid = $%d`,
				strings.Join(factorySet, ", "), fN)
			if _, err := tx.ExecContext(ctx, q, factoryArgs...); err != nil {
				return fmt.Errorf("update device_factory_meta: %w", err)
			}
		}

		// 2) tenant 迁移 — 用 reset_device_prefix 把 branch+下层全清零，再用 string-substitute 替换 tenant /48 头部
		if update.TenantID != "" && currentAddr.Valid {
			// 当前 tenant /48
			var currentTenantPrefix string
			if err := tx.QueryRowContext(ctx, `
				SELECT host(network(set_masklen($1::INET, 48))) || '/48'
			`, currentAddr.String).Scan(&currentTenantPrefix); err != nil {
				return fmt.Errorf("derive current tenant prefix: %w", err)
			}
			if currentTenantPrefix != update.TenantID {
				// 业务规则（v1 R-001 保留）：device 不能直接从一个非 pivot tenant 调到另一个非 pivot tenant，
				// 必须先回 System (fd00:0:1::/48) 中转。pivot = system / trash (fd00:0:1::/48 / fd00:0:2::/48)。
				currentIsPivot := isV2PivotTenantPrefix(currentTenantPrefix)
				newIsPivot := isV2PivotTenantPrefix(update.TenantID)
				if !currentIsPivot && !newIsPivot {
					name1, name2 := tenantNamesOrIDsByPrefix(ctx, tx, currentTenantPrefix, update.TenantID)
					return fmt.Errorf("cannot migrate device directly from %s to %s: move to System first", name1, name2)
				}
				// reset 到 branch 层 — 把当前 addr 的 byte 6..11 清零（保留 MAC bytes 12-15）
				// 用 platform_admin 的 fd00::/32 scope 调用
				var resetAddr string
				if err := tx.QueryRowContext(ctx, `
					SELECT host(reset_device_prefix(
					  $1::INET, 'fd00::/32'::INET, 32::SMALLINT, 'branch'
					)) || '/128'
				`, currentAddr.String).Scan(&resetAddr); err != nil {
					return fmt.Errorf("reset_device_prefix: %w", err)
				}
				// 替换 tenant /48 头部：取 resetAddr 的 host bits + target tenant 的 network bits
				// SQL: target_tenant_prefix（/48 网络部分）OR reset_addr 的低 80bit
				var newAddr string
				if err := tx.QueryRowContext(ctx, `
					SELECT host(
					  set_masklen(network(set_masklen($1::INET, 48)), 128)
					  | ($2::INET & '::ffff:ffff:ffff:ffff:ffff'::INET)
					) || '/128'
				`, update.TenantID, resetAddr).Scan(&newAddr); err != nil {
					return fmt.Errorf("compose new tenant addr: %w", err)
				}
				// UPDATE devices.device_addr — device_ota PK=device_uid 不受 device_addr 变更影响
				if _, err := tx.ExecContext(ctx, `
					UPDATE devices SET device_addr = $2::INET, updated_at = NOW()
					 WHERE device_addr = $1::INET
				`, currentAddr.String, newAddr); err != nil {
					return fmt.Errorf("update devices.device_addr: %w", err)
				}
				// 后续 OTA UPSERT 用新地址定位
				currentAddr = sql.NullString{String: newAddr, Valid: true}
			}
		}

		// 3) OTA 字段 — UPSERT device_ota（v1→v2 字段映射）
		otaPlanChanged := update.OTATargetFWSet || update.OTATargetMCUSet ||
			update.OTAPermitSet || update.OTAWaySet || update.OTAScheduleSet
		if (otaPlanChanged || update.OTAStatusSet) && currentAddr.Valid {
			// 计算 approve_way（仅当 OTAPermitSet 或 OTAWaySet 时改）
			otaSet := []string{}
			otaArgs := []any{}
			oN := 1

			if update.OTATargetFWSet {
				otaSet = append(otaSet, fmt.Sprintf("target_firmware_version = $%d", oN))
				otaArgs = append(otaArgs, nullStringToAny(update.OTATargetFirmwareVersion))
				oN++
			}
			if update.OTATargetMCUSet {
				otaSet = append(otaSet, fmt.Sprintf("target_mcu_model = $%d", oN))
				otaArgs = append(otaArgs, nullStringToAny(update.OTATargetMCUModel))
				oN++
			}
			if update.OTAPermitSet || update.OTAWaySet {
				approveWay := mapV1OTAWayToV2ApproveWay(update.OTAPermit, update.OTAWay, update.OTATenantApproved)
				otaSet = append(otaSet, fmt.Sprintf("approve_way = $%d", oN))
				otaArgs = append(otaArgs, nullStringToAny(approveWay))
				oN++
			}
			if update.OTAScheduleSet {
				sched := parseV1OTAScheduleToTime(update.OTASchedule.String)
				otaSet = append(otaSet, fmt.Sprintf("schedule = $%d", oN))
				if sched.Valid {
					otaArgs = append(otaArgs, sched.Time)
				} else {
					otaArgs = append(otaArgs, nil)
				}
				oN++
			}
			if update.OTAStatusSet {
				otaSet = append(otaSet, fmt.Sprintf("status = $%d", oN))
				otaArgs = append(otaArgs, nullStringToAny(update.OTAStatus))
				oN++
			}
			// auto-reset：plan 变了但 status 没显式设 → status='idle'
			if otaPlanChanged && !update.OTAStatusSet {
				otaSet = append(otaSet, "status = 'idle'", "progress = NULL", "error = NULL")
			}
			otaSet = append(otaSet, "updated_at = NOW()")

			// UPSERT — Phase 2: device_ota PK = device_uid (logMAC)；通过 devices.device_addr 反查 device_uid
			otaArgs = append(otaArgs, currentAddr.String)
			q := fmt.Sprintf(`
				INSERT INTO device_ota (device_uid, status, updated_at)
				SELECT device_uid, 'idle', NOW() FROM devices WHERE device_addr = $%d::INET
				ON CONFLICT (device_uid) DO UPDATE SET %s
			`, oN, strings.Join(otaSet, ", "))
			if _, err := tx.ExecContext(ctx, q, otaArgs...); err != nil {
				return fmt.Errorf("upsert device_ota: %w", err)
			}
		}

		// 4) allow_access → devices.access (platform 审批位)
		if update.AllowAccessSet && currentAddr.Valid {
			if _, err := tx.ExecContext(ctx, `
				UPDATE devices SET access = $2, updated_at = NOW()
				 WHERE device_addr = $1::INET
			`, currentAddr.String, update.AllowAccess); err != nil {
				return fmt.Errorf("update devices.access: %w", err)
			}
		}
	}

	return tx.Commit()
}

// ----------------------------------------------------------------------------
// DeleteDeviceStore — 移除 devices(CASCADE device_ota) 与 device_factory_meta(CASCADE runtime_state)
// ----------------------------------------------------------------------------

func (r *PostgresDeviceStoreRepository) DeleteDeviceStore(ctx context.Context, deviceUID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 校验 device_uid 存在
	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM device_factory_meta WHERE device_uid = $1)`,
		deviceUID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("failed to query device: %w", err)
	}
	if !exists {
		return fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
	}

	// devices CASCADE 删 device_ota（FK ON DELETE CASCADE）
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM devices WHERE device_uid = $1`, deviceUID,
	); err != nil {
		return fmt.Errorf("failed to delete devices: %w", err)
	}
	// 删 device_factory_meta
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM device_factory_meta WHERE device_uid = $1`, deviceUID,
	); err != nil {
		return fmt.Errorf("failed to delete device_factory_meta: %w", err)
	}
	return tx.Commit()
}

// ----------------------------------------------------------------------------
// ImportDeviceStores — 批量导入；后者覆盖前者；同 Create 写 4 张表
// ----------------------------------------------------------------------------

func deviceStoreNormUID(uid string) string {
	return strings.TrimSpace(uid)
}

func deviceStoreNormCode(c sql.NullString) string {
	if !c.Valid {
		return ""
	}
	return strings.TrimSpace(c.String)
}

func (r *PostgresDeviceStoreRepository) ImportDeviceStores(ctx context.Context, items []*domain.DeviceStore) (successCount int, inserted []*domain.DeviceStore, skipped []*domain.DeviceStore, errors []*domain.DeviceStore, err error) {
	if len(items) == 0 {
		return 0, nil, nil, nil, nil
	}

	var errs []*domain.DeviceStore
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.DeviceType == "" || deviceStoreNormUID(item.DeviceUID) == "" {
			errs = append(errs, item)
		}
	}

	// 后者覆盖前者：按 device_uid 只保留最后一行
	lastByUID := make(map[string]*domain.DeviceStore)
	for _, item := range items {
		if item == nil || item.DeviceType == "" {
			continue
		}
		normUID := deviceStoreNormUID(item.DeviceUID)
		if normUID == "" {
			continue
		}
		lastByUID[normUID] = item
	}
	keys := make([]string, 0, len(lastByUID))
	for k := range lastByUID {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sk []*domain.DeviceStore
	var ins []*domain.DeviceStore

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	defer tx.Rollback()

	for _, k := range keys {
		item := lastByUID[k]
		normUID := k
		normCode := deviceStoreNormCode(item.DeviceCode)
		tenantPrefix := strings.TrimSpace(item.TenantID)
		if tenantPrefix == "" {
			tenantPrefix = defaultUnboundTenantPrefix
		}
		codeForInsert := sql.NullString{String: normCode, Valid: normCode != ""}

		// 已存在 → UPDATE device_factory_meta（不动 device_addr）
		var existingUID string
		err := tx.QueryRowContext(ctx,
			`SELECT device_uid FROM device_factory_meta WHERE device_uid = $1`,
			normUID,
		).Scan(&existingUID)
		if err == nil {
			updateMeta := `
				UPDATE device_factory_meta SET
					device_code = $2,
					device_type = $3::device_type_enum,
					device_model = $4,
					mac_wifi = $5,
					imei = $6,
					comm_mode = $7,
					mcu_model = $8
				WHERE device_uid = $1
			`
			if _, e := tx.ExecContext(ctx, updateMeta,
				normUID,
				nullStringToAny(codeForInsert),
				item.DeviceType,
				nullStringToAny(item.DeviceModel),
				nullStringToAny(item.MAC),
				nullStringToAny(item.IMEI),
				nullStringToAny(item.CommMode),
				nullStringToAny(item.MCUModel),
			); e != nil {
				errs = append(errs, item)
				continue
			}
			// firmware_version 若提供则同步到 device_factory_meta
			if item.FirmwareVersion.Valid && item.FirmwareVersion.String != "" {
				if _, e := tx.ExecContext(ctx, `
					UPDATE device_factory_meta SET firmware_version = $1
					 WHERE device_uid = $2
				`, item.FirmwareVersion.String, normUID); e != nil {
					errs = append(errs, item)
					continue
				}
			}
			successCount++
			ins = append(ins, &domain.DeviceStore{
				DeviceUID:  normUID,
				DeviceType: item.DeviceType,
				DeviceCode: codeForInsert,
				TenantID:   tenantPrefix,
			})
			continue
		}
		if err != sql.ErrNoRows {
			errs = append(errs, item)
			continue
		}

		// 新增 — 两张表写（dfm + devices）
		if _, e := tx.ExecContext(ctx, `
			INSERT INTO device_factory_meta (
				device_uid, device_code, device_type, device_model,
				mcu_model, mac_wifi, imei, comm_mode, firmware_version
			) VALUES ($1, $2, $3::device_type_enum, $4, $5, $6, $7, $8, $9)
		`,
			normUID,
			nullStringToAny(codeForInsert),
			item.DeviceType,
			nullStringToAny(item.DeviceModel),
			nullStringToAny(item.MCUModel),
			nullStringToAny(item.MAC),
			nullStringToAny(item.IMEI),
			nullStringToAny(item.CommMode),
			nullStringToAny(item.FirmwareVersion),
		); e != nil {
			errs = append(errs, item)
			continue
		}
		mac32 := deriveMAC32Suffix(item.MAC, normUID)
		if _, e := tx.ExecContext(ctx, `
			INSERT INTO devices (device_addr, device_uid, monitoring_enabled)
			VALUES (
			  set_masklen(network(set_masklen($1::INET, 48)), 128) | ('::' || $2)::INET,
			  $3, TRUE
			)
		`, tenantPrefix, mac32, normUID); e != nil {
			errs = append(errs, item)
			continue
		}

		successCount++
		ins = append(ins, &domain.DeviceStore{
			DeviceUID:  normUID,
			DeviceType: item.DeviceType,
			DeviceCode: codeForInsert,
			TenantID:   tenantPrefix,
		})
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, nil, nil, err
	}

	return successCount, ins, sk, errs, nil
}

// ----------------------------------------------------------------------------
// UpdateDeviceCode / UpdateFirmwareVersion
// ----------------------------------------------------------------------------

// UpdateDeviceCode 更新 device_code（厂家会话 ID） — 写 device_factory_meta
func (r *PostgresDeviceStoreRepository) UpdateDeviceCode(ctx context.Context, deviceUID, deviceCode string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE device_factory_meta SET device_code = $1 WHERE device_uid = $2`,
		deviceCode, deviceUID)
	return err
}

// UpdateFirmwareVersion 仅更新 firmware_version — 写 device_factory_meta
func (r *PostgresDeviceStoreRepository) UpdateFirmwareVersion(ctx context.Context, deviceUID, firmwareVersion string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE device_factory_meta SET firmware_version = $1 WHERE device_uid = $2
	`, firmwareVersion, deviceUID)
	return err
}
